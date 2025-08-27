import logging
import pandasdmx as sdmx
import pandas as pd
from io import BytesIO
from urllib.parse import urlencode
from yaspin import yaspin
from uvicorn.logging import DefaultFormatter

logger = logging.getLogger("sdmx")

# https://data.api.abs.gov.au/rest/codelist/ABS/CL_CPI_INDEX_17
def get_dataflow():
    try:
        logger.info("Getting dataflow for ABS")
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

# add the ability to pass hashmap of params
# Check ABS swagger for datakeys - dependent on the dataflowid need to add logic to get correct key or will error - some have differnt num of args D:D:D:D:D:
def get_data(id: str, datakey: str = None, timeout: int = 300):
    abs = sdmx.Request('ABS_XML', timeout=timeout)
    try:
        logging.info(f"Fetching data for ID: {id} — Might take a while... good luck :)")

        # data_msg = abs_xml.data(resource_id=id, force=True)
    
        params={
            'format':'jsondata',
            'detail': 'dataonly'
        }
        
        # startPeriod = 2000
        # endPeriod = 2025
        
        # if startPeriod:
        #     params['startPeriod'] = endPeriod
        # if endPeriod:
        #     params['endPeriod'] = endPeriod
        
        if datakey is None:
            # default to CPI one for now - fix when adding dynamic logic
            datakey = "1.10001...Q"
            
        id = f"{id}/{datakey}"
        

        # I have no idea why params arent working and I have spent to long on this they dont matter to much for us anyway
        msg = abs.get(
            resource_type='data',
            resource_id=id,
            # params=params,
            force=True
        )
         
        df = sdmx.to_pandas(msg.data)
        print(df.head())
        if df.empty:
            logging.error(f"No data found for ID: {id}")
            return pd.DataFrame()
        logging.info(f"Fetched data for ID {id} successfully.")
        return df
    except Exception as e:
        logging.error(f"Error fetching data for ID {id}: {e}")
        return pd.DataFrame()
    
def get_dsd(id: str):
    abs_xml = sdmx.Request('ABS_XML')
    try:
        logging.info(f"Fetching DSD for ID: {id}")
        dsd_msg = abs_xml.data_structure(resource_id=id, force=True)
        dsd = sdmx.to_pandas(dsd_msg.data_structure)
        if dsd.empty:
            logging.error(f"No DSD found for ID: {id}")
            return pd.DataFrame()
        logging.info(f"Fetched DSD for ID {id} successfully.")
        return dsd
    except Exception as e:
        logging.error(f"Error fetching DSD for ID {id}: {e}")
        return pd.DataFrame()
    
def get_codelists(dataflow_id):
    abs = sdmx.Request('ABS_XML')

    msg = abs.get(
        resource_type='datastructure',
        resource_id=f'{dataflow_id}',
        force=True
    )

    return msg
    
    
# dataflows = get_dataflow()
# contents = dataflows.reset_index().to_dict(orient="records")

# print("Available dataflows (first 10):")
# print(dataflows.head(10))

# df = get_data('CPI')
# print("Sample data (first 5 rows):")
# print(df.head(5))

# df.to_csv('../.testdata/ABORIGINAL_ID_POP_PROJ.csv')

# df_dsd = get_dsd('CPI')
# print("Data Structure Definition (DSD):")
# print(df_dsd.head(5))

# codelists = get_codelists('CPI')


    