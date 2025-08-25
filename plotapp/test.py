import logging
import pandasdmx as sdmx
import pandas as pd
from io import BytesIO
from urllib.parse import urlencode
from yaspin import yaspin
import json


# Example: Fetch data from ABS using pandaSDMX
# Replace with your actual request
# client = sdmx.Client('ABS')
# resp = client.data('CPI', key={'FREQ': 'Q', 'MEASURE': 'IXOB'}, params={'startPeriod': '2020'})
# df = resp.to_pandas()

# For demo, let's assume df looks like:
arrays = [
    ['AU', 'AU', 'NZ'],
    ['CPI', 'GDP', 'CPI']
]
index = pd.MultiIndex.from_arrays(arrays, names=('Country', 'Indicator'))
data = {'2020-Q1': [100, 200, 110], '2020-Q2': [101, 205, 111]}
df = pd.DataFrame(data, index=index)

print("Original MultiIndex DF:\n", df)

# 1. Reset index to flatten MultiIndex into columns
flat_df = df.reset_index()
print("\nFlattened DF:\n", flat_df)

# 2. Convert to list of dictionaries (records)
json_ready = flat_df.to_dict(orient='records')
print("\nJSON Ready:\n", json_ready)

def get_dataflow():
    try:
        abs= sdmx.Request('ABS_XML', log_level=logging.INFO)
        flow_msg = abs.dataflow(force=True)
        dataflows = sdmx.to_pandas(flow_msg.dataflow)
        if dataflows.empty:
            logging.warning("No dataflows found in the ABS XML response.")
            return pd.DataFrame()
        
        logging.info(f"Fetched {len(dataflows)} dataflows successfully.")
        return dataflows

    except sdmx.exceptions.RequestError as e:
        logging.error(f"RequestError while fetching dataflows: {e}")
    except sdmx.exceptions.ParseError as e:
        logging.error(f"ParseError while decoding the ABS response: {e}")
    except Exception as e:
        logging.error(f"Unexpected error: {e}")
    return pd.DataFrame()

df = get_dataflow()
print("Original MultiIndex DF:\n", df)
df_flat = df.reset_index()
df_flat.rename(columns={df_flat.columns[0]: "datatypeid", df_flat.columns[1]: "datatypename"}, inplace=True)
print("\nFlattened DF:\n", df_flat)
df_dict= df_flat.to_dict("records")
print(json.dumps(df_dict, indent=4))
