import dash
from dash import html, dcc, dash_table
from dash.dependencies import Input, Output
from flask import request, jsonify
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

logger = logging.getLogger("dash")
logger.propagate = False
logger.info("Dash App Loading...")
config = load_config()



# CHANGES NEEDED!!!!
# 1: change the SDMX from just a dict of the data (list in store_data) to a dict witch top level field with metadata
#  like x/y axis, title etc
# 2: Change SDMX_DATA to not be a glbal dict - use sqlite instead


SDMX_DATA = {}

LabelMap = {
    'TIME_PERIOD': 'Time Period',
    'OBS_VALUE': 'Value'
}

valid_columns_list = ["TIME_PERIOD", "OBS_VALUE"]

server = flask.Flask(__name__)
server.secret_key = os.environ.get("secret_key", "secret")

app = dash.Dash(__name__, server=server, requests_pathname_prefix="/dashboard/")

app.layout = html.Div([
    dcc.Location(id="url", refresh=False),
    dcc.Loading(
        id="loading-indicator",
        type="circle",
        children=[
            dcc.Store(id="data-store"),
            dash_table.DataTable(id="data-table", data=[], page_size=10),
            dcc.Graph(id="line"),
            dcc.Graph(id="bar")
        ]
    )
])

@server.route("/refresh-dashboard/", methods=['POST'])
def refresh_dashboard():
    global SDMX_DATA
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
        
        df = df.reset_index()
        logger.debug(f"DF RESET INDEX {df}")
        df = store_data_preprocess(df)
        
        records = df.reset_index().to_dict("records")
        records = [{k.upper(): v for k, v in rec.items()} for rec in records]

        logger.debug(f"Returning {len(records)} records for {dataflowid}")
        records = [{k.upper(): v for k, v in rec.items()} for rec in records]
        logger.debug(f"SDMX_DATA going to dcc.Store (head): {records[:5]}")
              
        SDMX_DATA = records

        return jsonify({"status": "ok", "records": records})

    except Exception as e:
        logger.error(f"Failed to fetch data for {dataflowid}: {e}")
        return jsonify({"status": "error", "message": str(e)}), 500
    
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

# might need to update store_data to pass metadata for labeletc.
@app.callback(
    Output("line", "figure"),
    Output("bar", "figure"),
    Input("data-store", "data")
)
def update_graphs(store_data):
    logger.info("Dash - Updating graphs")
    
    logger.debug(f"store_data type: {type(store_data)}")
    logger.debug(f"store_data from dcc.Store: {SDMX_DATA[:5]}")

    df = pd.DataFrame(store_data)
    df = store_data_preprocess(df)
    
    logger.debug(f"DASH - Data frame in update graph:\n{df.head()}")

    if df.empty:
        logger.warning("Dash - No Data in records for update graph")
        df = pd.DataFrame({"x": [], "y": []})
        
        
    
    fig_line = px.line(df, y="OBS_VALUE", labels=LabelMap, title=f"")
    fig_bar = px.bar(df, y="OBS_VALUE", labels=LabelMap, title=f"")
    return fig_line, fig_bar



def store_data_preprocess(df: pd.DataFrame)-> pd.DataFrame:
    try:
        # Data Transformation
        # df = dataset.data.copy()
        df = df.drop_duplicates(subset='TIME_PERIOD')
        df['TIME_PERIOD'] = pd.PeriodIndex(df['TIME_PERIOD'], freq='Q').to_timestamp()
        df['OBS_VALUE'] = df['OBS_VALUE'].astype('float64').round(2)

        # Reset DataFrame for Plotly
        df = df.set_index('TIME_PERIOD')
        df = df[['OBS_VALUE']]
    
        logger.debug(f"DASH - processed data frame:\n{df.head()}")
        return df       
    except Exception as e:
        logger.error(f"Dash - Failed to preprocess dataframe: {e}")
        return df


