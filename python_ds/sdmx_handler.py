import logging
import pandasdmx
import pandas as pd

def get_dataflow():
    try:
        abs_xml = pandasdmx.Request('ABS_XML', log_level=logging.INFO)
        flow_msg = abs_xml.dataflow(force=True)
        dataflows = pandasdmx.to_pandas(flow_msg.dataflow)
        if dataflows.empty:
            logging.warning("No dataflows found in the ABS XML response.")
            return pd.DataFrame()
        
        logging.info(f"Fetched {len(dataflows)} dataflows successfully.")
        return dataflows

    except pandasdmx.exceptions.RequestError as e:
        logging.error(f"RequestError while fetching dataflows: {e}")
    except pandasdmx.exceptions.ParseError as e:
        logging.error(f"ParseError while decoding the ABS response: {e}")
    except Exception as e:
        logging.error(f"Unexpected error: {e}")
    return pd.DataFrame()

def get_data(id: str, timeout: int = 300):
    abs_xml = pandasdmx.Request('ABS_XML', timeout=timeout)
    try:
        data_msg = abs_xml.data(resource_id=id, force=True)
        df = pandasdmx.to_pandas(data_msg.data)
        if df.empty:
            logging.error(f"No data found for ID: {id}")
            return pd.DataFrame()
        logging.info(f"Fetched data for ID {id} successfully.")
        return df
    except Exception as e:
        logging.error(f"Error fetching data for ID {id}: {e}")
        return pd.DataFrame()
    

# dataflows = get_dataflow()
# print("Available dataflows (first 10):")
# print(dataflows.head(10))

df = get_data('CPI')
print("Sample data (first 5 rows):")
print(df.head(5))

df.to_csv('../.testdata/cpi_test_data.csv', index=False)


    