import dash
from dash import html, dcc, dash_table
from dash.dependencies import Input, Output
import flask
import pandas as pd
import os

# Global data store is shared with FastAPI
dashboard_data = {
    "df": pd.DataFrame()
}

server = flask.Flask(__name__)
server.secret_key = os.environ.get("secret_key", "secret")

app = dash.Dash(__name__, server=server, requests_pathname_prefix="/dashboard/")

app.layout = html.Div([
    html.H3("ABS Dashboard"),
    dash_table.DataTable(id="data-table", data=[], page_size=10),
    dcc.Interval(id="interval", interval=2000, n_intervals=0)  # Poll every 2s
])

# Callback to automatically update the table from global dashboard_data
@app.callback(
    Output("data-table", "data"),
    Input("interval", "n_intervals")
)
def update_table(n):
    df = dashboard_data["df"]
    print(f"[Interval {n}] Dashboard df shape: {df.shape}, empty: {df.empty}")
    if df.empty:
        return []
    return df.reset_index().to_dict("records")
