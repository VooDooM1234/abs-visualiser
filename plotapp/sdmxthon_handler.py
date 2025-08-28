
# alternative lib for SDMX extracts
import sdmxthon
import pandas as pd
import plotly.express as px

abs_base_url = "https://data.api.abs.gov.au/rest/"
structure_id = "CPI"
dsd_url = f"{abs_base_url}datastructure/ABS/{structure_id}/?references=children"

message_metadata = sdmxthon.read_sdmx('https://api.data.abs.gov.au/dataflow/ABS/CPI?references=all')
print(message_metadata.payload)
message_data = sdmxthon.read_sdmx('https://data.api.abs.gov.au/rest/data/CPI/1.10001...Q?detail=full')
print(message_data.payload)
dataset = message_data.content['ABS:CPI(1.1.0)']
print(f"Datset data:\n{dataset.data}")

dsd = message_metadata.content['DataStructures']['ABS:CPI(1.1.0)']
print(f"DSD\n: {dsd.content}")


# Data Transformation
df = dataset.data.copy()
df = df.drop_duplicates(subset='TIME_PERIOD')
df['TIME_PERIOD'] = pd.PeriodIndex(df['TIME_PERIOD'], freq='Q').to_timestamp()
df['OBS_VALUE'] = df['OBS_VALUE'].astype('float64').round(2)

# Reset DataFrame for Plotly
df = df.set_index('TIME_PERIOD')
df = df[['OBS_VALUE']]

# Debug
# print("Column data types:\n", df.dtypes)
# print("\nIndex type:", type(df.index))
# print("Index name:", df.index.name)

fig = px.line(
    df,
    # if x-axis isnt index then add back in if needed later
    # x="TIME_PERIOD",
    y="OBS_VALUE", 
    title='CPI over Time'
).update_layout(title_x=0.5)

fig.show()