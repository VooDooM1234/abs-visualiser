import pandasdmx as sdmx
import pandas as pd

# Create a pandasdmx Request object for ABS
abs_api = sdmx.Request('ABS', timeout=120)

# Manually fetch the dataflow from the ABS REST endpoint
flow_msg = abs_api.get("https://data.api.abs.gov.au/rest/dataflow")

# Convert the dataflow to pandas Series
dataflows = sdmx.to_pandas(flow_msg.dataflow)

# Print first 10 available dataflows
print("Available dataflows (first 10):")
print(dataflows.head(10))