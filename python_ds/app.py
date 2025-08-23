# python_ds/app.py
# Plotting microservice for ABS data visualization; plots graphs using Plotly and serves them via FastAPI.
# Api flow /plot/{graphname}/{dataflow}

from fastapi import FastAPI
from fastapi.responses import HTMLResponse, JSONResponse
import plotly.express as px
import plotly.io as pio
import pandas as pd
from python_ds.db import init_db
import logging
from .config import load_config
import json
from pydantic import BaseModel
from . import sdmx_handler
import requests


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

# Expects dataflow passed to already be check in GO backend
@app.get("/plot/{graphName}/{dataflow}", response_class=HTMLResponse)
async def get_abs_cpi_plot(graphName: str, dataflow: str):
    try:
        # Get data to graph
        df = dataflow_to_query(dataflow)
         
        plot_func = graphRegistry.get(graphName)
        if plot_func is None:
            logger.error(f"Graph function {graphName} not found for dataflow {dataflow}")
            return HTMLResponse(content="<h3>Graph not found</h3>", status_code=404)
        #TODO: Add more params for labels, titles, etc.
        fig = plot_func(df)
        if fig is None:
            logger.error(f"Graph function {graphName} returned None for dataflow {dataflow}")
            return HTMLResponse(content="<h3>Graph not found</h3>", status_code=404)
        return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

    except Exception as e:
        logger.error(f"Error generating plot for graphName {graphName} and dataflow {dataflow}: {e}")
    
        return HTMLResponse(content=f"<h3>Micro Service: Error generating plot: {e}</h3>", status_code=500)

@app.get("/dashboard/database/{dataflowid}/", response_class=HTMLResponse)
async def get_dashboard(dataflowid: str):
    try:
        df = dataflow_to_query(dataflowid)
        if df.empty:
            logger.error(f"No data available for dataflowid: {dataflowid}")
            return HTMLResponse(content="<h3>No data available for the selected dataflowid</h3>", status_code=404)
        
        for graphName, plot_func in graphRegistry.items():
            fig = plot_func(df)

            html_plot = pio.to_html(fig, full_html=False, include_plotlyjs='cdn')
           
            combined_html += f"<div id='{graphName}' class='plot-container'>{html_plot}</div>"

        return HTMLResponse(content=combined_html)

    except Exception as e:
        logger.error(f"Error generating dashboard for dataflowid {dataflowid}: {e}")    
        return HTMLResponse(content=f"Error:{e}</h3>", status_code=500)
    
#Docs: https://fastapi.tiangolo.com/tutorial/body/#import-pydantics-basemodel   

class DashboardRequest(BaseModel):
    dataflowid: str
    timePeriod  : str
    value: float  
    
@app.post("/dashboard/api/{dataflowid}/", response_class=HTMLResponse)
async def get_dashboard(dataflowid: str, dashboardRequest: DashboardRequest):
    try:
        data_dict = dashboardRequest.model_dump()
        df = pd.DataFrame([data_dict])
        if df.empty:
            logger.error(f"No data available for dataflowid: {dataflowid}")
            return HTMLResponse(content="", status_code=404)
        
        for graphName, plot_func in graphRegistry.items():
            fig = plot_func(df)

            html_plot = pio.to_html(fig, full_html=False, include_plotlyjs='cdn')
           
            combined_html += f"<div id='{graphName}' class='plot-container'>{html_plot}</div>"

        return HTMLResponse(content=combined_html)

    except Exception as e:
        logger.error(f"Error generating dashboard for dataflowid {dataflowid}: {e}")    
        return HTMLResponse(content=f"Error</h3>", status_code=500)

class requestDataABS(BaseModel):
    dataflowid: str

# Fix the merging of the codelist to give real names
# fix payload so onle period and value is repeated, keep others as metadata
@app.post("/request-data/ABS/", response_class=JSONResponse)
async def get_data_abs(payload: requestDataABS):
    try:
        dataflowid = payload.dataflowid
        logger.info(f"POST request for ABS data received: {dataflowid}")
       
        df = sdmx_handler.get_data(dataflowid)
        # codelist = sdmx_handler.get_codelists(dataflowid)
        
        # df = pd.merge(df, codelist, on='INDEX').query("INDEX != ''")    
        
        df_flat = df.reset_index()
        content = df_flat.to_dict(orient="records")  
        
        return JSONResponse(content=content)
    except ValueError as e:
        error_payload = {
            "error": "requestDataFailed",
            "message": str(e),
        }
        return JSONResponse(content=error_payload)
    
class requestCodelistABS(BaseModel):
    dataflowid: str    
    
@app.post("/request-codelist/ABS/", response_class=JSONResponse)
async def get_data_abs(payload: requestCodelistABS):
    try:
        dataflowid = payload.dataflowid
        logger.info(f"POST request for ABS data received: {dataflowid}")
        codelists = sdmx_handler.get_codelists(dataflowid)
        
        
        return JSONResponse(codelists)
    except ValueError as e:
        error_payload = {
            "error": "requestDataCodelist",
            "message": str(e),
        }
        return JSONResponse(error_payload)
    
    


# for isolation testing, run with:
# if __name__ == "__main__":
#     import uvicorn
#     uvicorn.run("python_ds.app:app", host="0.0.0.0", port=8082, reload=True)


