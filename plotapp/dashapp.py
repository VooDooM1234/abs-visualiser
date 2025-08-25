import dash
from dash import html, dcc, dash_table
from dash.dependencies import Input, Output
from flask import Flask, request, jsonify
from urllib.parse import parse_qs

import flask
import pandas as pd
import os
import plotly as px
import logging

from plotapp import fetch_ABS_SDMX as fsdmx

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("uvicorn.error")

# CHNGE TO NOT BE GLOBAL THINGO BUT SQLITE
SDMX_DATA = {}

server = flask.Flask(__name__)
server.secret_key = os.environ.get("secret_key", "secret")

app = dash.Dash(__name__, server=server, requests_pathname_prefix="/dashboard/")


# app.layout = html.Div([
#     # html.H3("ABS Dashboard"),
#     dash_table.DataTable(id="data-table", data=[], page_size=10),
#     # Poll every 2s - fix this to not poll and by dynamic through API call
#     dcc.Interval(id="interval", interval=2000, n_intervals=0)  
# ])

app.layout = html.Div([
    dcc.Location(id="url", refresh=False),
    dcc.Store(id="data-store"),
    dash_table.DataTable(id="data-table", data=[], page_size=10),
    dcc.Graph(id="graph")
])

@app.callback(
    Output("data-store", "data"),
    Input("url", "search")
)
def load_data_from_url(search):
    params = parse_qs(search[1:] if search else "")
    datatype = params.get("datatype", ["default"])[0]
    records = SDMX_DATA.get(datatype, [])
    return {"datatype": datatype, "records": records}

@app.callback(
    Output("data-table", "data"),
    Input("data-store", "data")
)
def update_table(store_data):
    return store_data.get("records", [])

# Callback to automatically update the table from global dashboard_data
# @app.callback(
#     Output("data-table", "data"),
#     Input("interval", "n_intervals")
# )
# def update_table(n):
#     df = dashboard_data["df"]
#     print(f"[Interval {n}] Dashboard df shape: {df.shape}, empty: {df.empty}")
#     if df.empty:
#         return []
#     return df.reset_index().to_dict("records")

@server.route("/refresh-dashboard/", methods=['POST'])
def update_data():
    data = request.json
    dataflowid = data["dataflowid"]
    logger.info(f"Dashboard data updated for: {dataflowid}")
    df = fsdmx.get_data(dataflowid)
    records = df.reset_index().to_dict("records")
    SDMX_DATA.update(records) 
    return jsonify({"status": "ok"})

@app.callback(
    Output("graph", "figure"),
    Input("data-store", "data")
)
def update_graph(store_data):
    records = store_data.get("records", [])
    df = pd.DataFrame(records)
    if df.empty:
        df = pd.DataFrame({"x": [], "y": []})
    fig = px.line(df, x="x", y="y", title=f"Dataset: {store_data.get('datatype', '')}")
    return fig
    
    
    

if __name__ == "__main__":
    app.run(debug=True)
