import dash
from dash import html, dcc, dash_table
from dash.dependencies import Input, Output
from flask import Flask, request, jsonify
from urllib.parse import parse_qs
from plotapp.config import load_config
from urllib.parse import urlparse
from urllib.parse import parse_qs

import flask
import pandas as pd
import os
import plotly.express as px
import logging

from plotapp import fetch_ABS_SDMX as fsdmx
from uvicorn.logging import DefaultFormatter

logger = logging.getLogger("dash")
logger.propagate = False
logger.info("Dash App Loading...")
config = load_config()



# CHANGES NEEDED!!!!
# 1: change the SDMX from just a dict of the data (list in store_data) to a dict witch top level field with metadata
#  like x/y axis, title etc
# 2: Change SDMX_DATA to not be a glbal dict - use sqlite instead


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
    dcc.Loading(
        id="loading-indicator",
        type="circle",
        children=[
            dcc.Store(id="data-store"),
            dash_table.DataTable(id="data-table", data=[], page_size=10),
            dcc.Graph(id="graph")
        ]
    )
])

@app.callback(
    Output("data-store", "data"),
    Input("url", "href")
)
def load_data_from_url(href):
    logger.info("DASH - Loading data from URL callback")
    parsed_url = urlparse(href)
    dataflowid = parse_qs(parsed_url.query).get("dataflowid", [None])[0]
    
    logger.debug(f"Parsed_url: {parsed_url}")
    logger.debug(f"dataflowid: {dataflowid}")
    
    if not dataflowid:
        logger.warning("No datatypeid found in URL")
        return {"records": []}
    
    return SDMX_DATA

@server.route("/refresh-dashboard/", methods=['POST'])
def refresh_dashboard():
    data = request.get_json()
    dataflowid = data.get("dataflowid") if data else None

    if not dataflowid:  # catches None or empty string
        logger.error(f"Missing or empty dataflowid: {data}")
        return jsonify({"status": "error", "message": "Missing or empty dataflowid"}), 400

    logger.info(f"Refreshing dashboard data for: {dataflowid}")

    try:
        df = fsdmx.get_data(dataflowid)

        if df.empty:
            logger.warning(f"No data returned for dataflowid: {dataflowid}")
            return jsonify({"status": "error", "message": "No data returned"}), 404

        # Return records directly
        records = df.reset_index().to_dict("records")
        logger.debug(f"Returning {len(records)} records for {dataflowid}")
        return jsonify({"status": "ok", "records": records})

    except Exception as e:
        logger.error(f"Failed to fetch data for {dataflowid}: {e}")
        return jsonify({"status": "error", "message": str(e)}), 500


# @app.callback(
#     Output("data-table", "data"),
#     Input("data-store", "data")
# )
# def update_table(store_data):
#     logger.info("Dash - Updating table...")
#     return store_data.get("records", [])

@app.callback(
    Output("graph", "figure"),
    Input("data-store", "data")
)
def update_graph(store_data):
    logger.info("Dash - Updating graphs")
    logger.debug(f"store_data type: {type(store_data)}")
    # logger.debug(f"store_data content: {store_data}")
    # records = store_data.get("records", [])
    df = pd.DataFrame(store_data)
    logger.debug(f"DASHAPP - Data frame in update graph:\n{df.head()}")

    if df.empty:
        logger.warning("Dash - No Data in records for update graph")
        df = pd.DataFrame({"x": [], "y": []})
        
    fig = px.line(df, x="TIME_PERIOD", y="value") #, title=f"Dataset: {store_data.get('datatype', '')}")
    return fig

@server.route("/debug/sdmx-cache", methods=["GET"])
def debug_sdmx_cache():
    return jsonify(SDMX_DATA)


