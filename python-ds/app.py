import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import plotly.express as px
import psycopg
import os

import csv

from dotenv import load_dotenv

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

        for record in cur:
            print(record)

    query = "SELECT * FROM ABS_CPI"
    df = pd.read_sql(query, conn)
    print(df.head())

# fig = px.bar(x=["a", "b", "c"], y=[1, 3, 2])
# fig.show()

# plt.plot([1,2,3], [4,5,6])
# plt.show()

# print("Hello world")
