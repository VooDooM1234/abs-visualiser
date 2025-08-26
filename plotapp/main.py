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

from plotapp.config import load_config
from plotapp import fetch_ABS_SDMX as fsdmx

import uvicorn
from plotapp.dashapp import app as dash_app

app = FastAPI()

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

@app.get("/request-dataflow/ABS/", response_class=JSONResponse)
async def get_dataflow_all():
    try:
        with yaspin(text="Fetching dataflows for ABS", color="cyan") as spinner:
            logger.info("GET request for ABS dataflow")
            
            df = fsdmx.get_dataflow()
            df_flat = df.reset_index()
            df_flat.rename(columns={
                df_flat.columns[0]: "dataflowid",
                df_flat.columns[1]: "dataflowname"
            }, inplace=True)
            
            df_dict = df_flat.to_dict("records")
            logger.debug(json.dumps(df_dict, indent=4))
            spinner.ok("✅")

        return JSONResponse(content=df_dict)
    
    except Exception as e:
        logger.error(f"Failed to fetch dataflow: {e}")
        return JSONResponse(content={"status": "failed"})
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
        records = codelists.reset_index().to_dict(orient="records")

        return JSONResponse(records)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=f"requestDataCodelist: {str(e)}")
    
    
# dash_app = create_dashboard(requests_pathname_prefix="/dashboard/")
# dash_app.enable_dev_tools()
app.mount("/dashboard", WSGIMiddleware(dash_app.server))


# if __name__ == "__main__":
#     uvicorn.run(app, port=8000)