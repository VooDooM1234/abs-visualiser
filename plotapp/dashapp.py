import dash
from dash import html, dcc, dash_table
from dash.dependencies import Input, Output
from flask import Flask, request, jsonify
from urllib.parse import parse_qs
from plotapp.config import load_config

import flask
import pandas as pd
import os
import plotly.express as px
import logging

from plotapp import fetch_ABS_SDMX as fsdmx

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("uvicorn.error")
config = load_config()

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
    dataflowid = request.args.get('dataflowid')
    if dataflowid is None:
        logger.error("No query param passed")
    records = SDMX_DATA.get(dataflowid, [])
    return {"dataflowid": dataflowid, "records": records}

@app.callback(
    Output("data-table", "data"),
    Input("data-store", "data")
)
def update_table(store_data):
    return store_data.get("records", [])

@server.route("/refresh-dashboard/", methods=['POST'])
def update_data():
    # dataflowid = request.form.get('dataflowid')
    data = request.get_json()
    if not data or "dataflowid" not in data:
        return jsonify({"status": "error", "message": "Missing dataflowid"}), 400
    
    dataflowid = data["dataflowid"]
    logger.info(f"Dashboard data updated for: {dataflowid}")
    
    df = fsdmx.get_data(dataflowid)
    records = df.reset_index().to_dict("records")
    SDMX_DATA[dataflowid] = records
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
    
# if __name__ == "dashapp":
#     app.run(debug=True, port=8083)
