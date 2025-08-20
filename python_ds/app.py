import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import plotly.express as px
import psycopg
import os

import csv

from dotenv import load_dotenv
import plotly.io as pio

load_dotenv('../.env')

# api_key = os.getenv("API_KEY")

DATABASE_user = os.getenv("DATABASE_USER")
DATABASE_password = os.getenv("DATABASE_PASSWORD")
DATABASE_name = os.getenv("DATABASE_NAME")
DATABASE_host = os.getenv("DATABASE_HOST", "localhost") 


# with open('../.testdata/CPI.csv', newline='') as csvfile:
#     reader = csv.DictReader(csvfile)
#     for row in reader:
#         print(row['TIME_PERIOD'], row['OBS_VALUE'])
# print(row)


# Connect to an existing database
with psycopg.connect(dbname=DATABASE_name,user=DATABASE_user,password=DATABASE_password,host=DATABASE_host) as conn:

    # Open a cursor to perform database operations
    with conn.cursor() as cur:

        cur.execute("SELECT * FROM ABS_CPI")

    query = "SELECT * FROM ABS_CPI"
    df = pd.read_sql(query, conn)

    df['value'] = pd.to_numeric(df['value'], errors='coerce')
    df['value'] = df['value'].round(2)
    df['value'] = df['value'].astype(float)  # Ensure float type

fig = px.bar(
    df,
    x="time_period",
    y="value",
    title="Consumer Price Index - Quarterly",
    labels={'time_period': 'Time Period', 'value': 'CPI'}
)

fig.update_yaxes(tickformat=".2f")
fig.update_traces(hovertemplate='Time Period=%{x}<br>CPI=%{y:.2f}<extra></extra>')

fig.show()

for trace in fig.data:
    trace.hovertemplate = 'Time Period=%{x}<br>CPI=%{y:.2f}<extra></extra>'
    trace.texttemplate = '%{y:.2f}'  # If you want bar labels to show decimals

fig.update_yaxes(tickformat=".2f")

pio.write_html(
    fig,
    file="../templates/div_frags/bar_plot_abs_cpi.html",
    full_html=False,
    include_plotlyjs=False
)

fig_json = fig.to_json()
print(fig_json)



