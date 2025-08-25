#main.py
import json
import logging
import requests
import pandas as pd
import plotly.express as px
import plotly.io as pio

from fastapi import FastAPI, HTTPException, Request, Form
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.wsgi import WSGIMiddleware

from pydantic import BaseModel
from yaspin import yaspin

from plotapp.db import init_db
from plotapp.config import load_config
from plotapp import fetch_ABS_SDMX as fsdmx

import uvicorn
from plotapp.dashapp import app as dash_app, dashboard_data
# from plotapp.dashapp import create_dashboard

app = FastAPI()
db_conn = init_db(load_config())
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("uvicorn.error")

def bar_plot(df):
    return px.bar(df, x="timePeriod", y="value", title="Bar Chart")
def line_plot(df):
    return px.line(df, x="timePeriod", y="value", title="Line Chart")
def scatter_plot(df):
    return px.scatter(df, x="timePeriod", y="value", title="Scatter Plot")
def pie_plot(df):
    return px.pie(df, names="timePeriod", values="value", title="Pie Chart")
def histogram_plot(df):
    return px.histogram(df, x="timePeriod", y="value", title="Histogram")

graphRegistry = {
    "bar": bar_plot,
    "line": line_plot,
    "scatter": scatter_plot,
    "pie": pie_plot,
    "histogram": histogram_plot,
}   

def dataflow_to_query(dataflow):
    dataflow = dataflow.replace("-", "_").upper()
    query = f"SELECT * FROM {dataflow}"
    df = pd.read_sql(query, db_conn)
    df['value'] = pd.to_numeric(df['value'], errors='coerce').round(2).astype(float)
    df = df.rename(columns={"time_period": "timePeriod", "value": "value"})
    return df


@app.get("/plot/test", response_class=HTMLResponse)
async def get_plot():
    fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
    return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

@app.get("/plot/test/json", response_class=JSONResponse)
async def get_plot_json():
    fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
    return fig.to_dict()
    
@app.get("/metadata/valid-graphs", response_class=JSONResponse)
async def put_plot_metadata_valid_graphs():
    valid_graphs = list(graphRegistry.keys())
    return JSONResponse(content={"valid_graphs": valid_graphs}, status_code=200)

# This is being mounted now so i dont think I need??
class updateDashboardRequest(BaseModel):
    dataflowid: str

@app.post("/refresh-dashboard/", response_class=JSONResponse)
async def get_plot_dashboard(payload: updateDashboardRequest):
    try:
        with yaspin(text=f"Fetching data for ID: {id} — Might take a while... good luck :)", color="cyan") as spinner:
            logger.info(f"POST request for ABS dashboard received: {payload.dataflowid}")
            df = fsdmx.get_data(payload.dataflowid) 
            logger.info(f"Dashboard data updated for: {payload.dataflowid}")
            dashboard_data["df"] = df 
            print(f"[POST] Updated dashboard_data df shape: {df.shape}, empty: {df.empty}")
            spinner.ok("✅")
        return JSONResponse(content={"status": "success"})
        
    except Exception as e:
        # raise HTTPException(status_code=500, detail=f"Dashboard generation failed: {str(e)}")
        return JSONResponse(content={"status": "fail"})

@app.get("/plot/{graphName}/{dataflow}", response_class=HTMLResponse)
async def get_abs_cpi_plot(graphName: str, dataflow: str):
    try:
        df = dataflow_to_query(dataflow)
         
        plot_func = graphRegistry.get(graphName)
        if plot_func is None:
            logger.error(f"Graph function {graphName} not found for dataflow {dataflow}")
            raise HTTPException(status_code=404, detail="Graph not found")
        fig = plot_func(df)
        if fig is None:
            logger.error(f"Graph function {graphName} returned None for dataflow {dataflow}")
            raise HTTPException(status_code=404, detail="Graph not found")
        return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

    except Exception as e:
        logger.error(f"Error generating plot for graphName {graphName} and dataflow {dataflow}: {e}")
        raise HTTPException(status_code=500, detail=f"Micro Service: Error generating plot: {str(e)}")

@app.get("/dashboard-sample/", response_class=HTMLResponse)
async def get_dashboard(dataflowid: str):
    try:
        df = dataflow_to_query(dataflowid)
        if df.empty:
            logger.error(f"No data available for dataflowid: {dataflowid}")
            raise HTTPException(status_code=404, detail="No data available for the selected dataflowid")
        
        for graphName, plot_func in graphRegistry.items():
            fig = plot_func(df)
            html_plot = pio.to_html(fig, full_html=False, include_plotlyjs='cdn')
            combined_html += f"<div id='{graphName}' class='plot-container'>{html_plot}</div>"

        return HTMLResponse(content=combined_html)

    except Exception as e:
        logger.error(f"Error generating dashboard for dataflowid {dataflowid}: {e}")    
        raise HTTPException(status_code=500, detail=f"Error generating dashboard: {str(e)}")

class DashboardRequest(BaseModel):
    dataflowid: str
    timePeriod  : str
    value: float  
    
# @app.post("/dashboard/api/{dataflowid}/", response_class=HTMLResponse)
# async def get_dashboard(dataflowid: str, dashboardRequest: DashboardRequest):
#     try:
#         data_dict = dashboardRequest.model_dump()
#         df = pd.DataFrame([data_dict])
#         if df.empty:
#             logger.error(f"No data available for dataflowid: {dataflowid}")
#             raise HTTPException(status_code=404, detail="No data available for the selected dataflowid")
        
#         for graphName, plot_func in graphRegistry.items():
#             fig = plot_func(df)
#             html_plot = pio.to_html(fig, full_html=False, include_plotlyjs='cdn')
#             combined_html += f"<div id='{graphName}' class='plot-container'>{html_plot}</div>"

#         return HTMLResponse(content=combined_html)

#     except Exception as e:
#         logger.error(f"Error generating dashboard for dataflowid {dataflowid}: {e}")    
#         raise HTTPException(status_code=500, detail=f"Error generating dashboard: {str(e)}")

class requestDataABS(BaseModel):
    dataflowid: str

@app.post("/request-data/ABS/", response_class=JSONResponse)
async def get_data_abs(payload: requestDataABS):
    try:
        dataflowid = payload.dataflowid
        logger.info(f"POST request for ABS data received: {dataflowid}")
        df = fsdmx.get_data(dataflowid)
        df_flat = df.reset_index()
        content = df_flat.to_dict(orient="records")  
        return JSONResponse(content=content)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=f"requestDataFailed: {str(e)}")

class requestCodelistABS(BaseModel):
    dataflowid: str    
    
@app.post("/request-codelist/ABS/", response_class=JSONResponse)
async def get_data_abs(payload: requestCodelistABS):
    try:
        dataflowid = payload.dataflowid
        logger.info(f"POST request for ABS data received: {dataflowid}")
        codelists = fsdmx.get_codelists(dataflowid)
        return JSONResponse(codelists)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=f"requestDataCodelist: {str(e)}")
    
    
# dash_app = create_dashboard(requests_pathname_prefix="/dashboard/")
# dash_app.enable_dev_tools()
app.mount("/dashboard", WSGIMiddleware(dash_app.server))


# if __name__ == "__main__":
#     uvicorn.run(app, port=8000)