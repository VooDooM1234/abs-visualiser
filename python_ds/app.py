# python_ds/app.py
# Plotting microservice for ABS data visualization; plots graphs using Plotly and serves them via FastAPI.
# Api flow /plot/{graphname}/{dataflow}

from fastapi import FastAPI
from fastapi.responses import HTMLResponse, JSONResponse
import plotly.express as px
import plotly.io as pio
import pandas as pd
from .db import init_db
from .config import load_config
import json

app = FastAPI()
db_conn = init_db(load_config())


def bar_plot(df):
    return px.bar(df, x="time_period", y="value", title="Bar Chart")
def line_plot(df):
    return px.line(df, x="time_period", y="value", title="Line Chart")
def scatter_plot(df):
    return px.scatter(df, x="time_period", y="value", title="Scatter Plot")
def pie_plot(df):
    return px.pie(df, names="time_period", values="value", title="Pie Chart")
def histogram_plot(df):
    return px.histogram(df, x="time_period", y="value", title="Histogram")

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

@app.get("/plot/bar-abs-cpi", response_class=HTMLResponse)
async def get_abs_cpi_plot():
    try:
        query = "SELECT * FROM ABS_CPI"
        df = pd.read_sql(query, db_conn)

        df['value'] = pd.to_numeric(df['value'], errors='coerce').round(2).astype(float)

        fig = px.bar(
            df,
            x="time_period",
            y="value",
            title="Consumer Price Index - Quarterly",
            labels={'time_period': 'Time Period', 'value': 'CPI'}
        )
        return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

    except Exception as e:
        return HTMLResponse(content=f"<h3>Error generating plot: {e}</h3>", status_code=500)
    
@app.get("/metadata/valid-graphs", response_class=JSONResponse)
async def put_plot_metadata_valid_graphs():
    valid_graphs = list(graphRegistry.keys())
    return JSONResponse(content={"valid_graphs": valid_graphs}, status_code=200)

# Expects dataflow passed to already be check in GO backend
@app.get("/plot/{graphName}/{dataflow}", response_class=HTMLResponse)
async def get_abs_cpi_plot(graphName: str, dataflow: str):
    try:
        # Get data to graph
        dataflow_to_query(dataflow)
         
        graphRegistry.get(graphName)
        fig = graphRegistry[graphName]
        if fig is None:
            return HTMLResponse(content="<h3>Graph not found</h3>", status_code=404)
        return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

    except Exception as e:
        return HTMLResponse(content=f"<h3>Micro Service: Error generating plot: {e}</h3>", status_code=500)


def dataflow_to_query(dataflow):
    dataflow = dataflow.replace("-", "_").upper()
    query = f"SELECT * FROM {dataflow}"
    df = pd.read_sql(query, db_conn)
    df['value'] = pd.to_numeric(df['value'], errors='coerce').round(2).astype(float)
    return df




