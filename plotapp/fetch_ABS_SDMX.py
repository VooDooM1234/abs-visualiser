import logging
import pandasdmx as sdmx
import pandas as pd
from io import BytesIO
from urllib.parse import urlencode
from yaspin import yaspin
from uvicorn.logging import DefaultFormatter
import pprint
import collections

logger = logging.getLogger("main")
logger.propagate = False

def get_dataflow():
    try:
        logger.info("Getting dataflow for ABS")
        abs=sdmx.Request('ABS_XML')
        flow_msg = abs.dataflow(force=True)
        dataflows = sdmx.to_pandas(flow_msg.dataflow)
        if dataflows.empty:
            logger.warning("No dataflows found in the ABS XML response.")
            return pd.DataFrame()
        
        logger.info(f"Fetched {len(dataflows)} dataflows successfully.")
        return dataflows

    except sdmx.exceptions.RequestError as e:
        logger.error(f"RequestError while fetching dataflows: {e}")
    except sdmx.exceptions.ParseError as e:
        logger.error(f"ParseError while decoding the ABS response: {e}")
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
    return pd.DataFrame()

def get_data(id: str, timeout: int = 120):
    abs = sdmx.Request('ABS_XML', timeout=timeout)
    try:
        logger.info(f"Fetching data for ID: {id} — Might take a while... good luck :)")
                    
        dims, dsd = get_metadata(id)
        codelists = get_codelists(id)
        logger.debug(f"Number of Codelists: {len(codelists)}")
        params = dict(startPeriod='2024')
        key = key_generator(dims=dims, codelists=codelists)
        
        data = abs.data(id, key=key).data[0]

        # Step 5 - Transform data response
        df = sdmx.to_pandas(data, datetime='TIME_PERIOD')
        df.columns = ['OBS_VALUE']
         
        if df.empty:
            logger.error(f"No data found for ID: {id}")
            return pd.DataFrame()
        
        logger.info(f"Fetched data for ID {id} successfully.")
        
        return df
    except Exception as e:
        logger.error(f"Error fetching data for ID {id}: {e}")
        return pd.DataFrame()
    
def get_metadata(id: str):
    try:
        abs = sdmx.Request('ABS_XML')
        msg = abs.datastructure(id)
        logger.debug("Message: %s", msg)

        concept_scheme = msg.concept_scheme
        logger.debug("Concept Scheme: %s", concept_scheme)

        dsd = msg.structure[id]
        logger.debug("DSD: %s", dsd)
                
        components = dsd.dimensions.components
        logger.debug("Components: %s", components)

        dims = tuple(item.id for item in components)
        logger.debug("Dimensions: %s", dims)
        return dims, dsd
    except Exception as e:
        logger.error(f"Error fetching metadata for ID {id}: {e}")

# Returns nested dict with component name(to join on all other metadata){Codelistname: dict of codelists}
def get_codelists(id: str) -> dict:
    abs = sdmx.Request('ABS_XML')
    dsd = abs.datastructure(id).structure[id]
            
    codelists = collections.defaultdict(dict)
    for dim in dsd.dimensions.components:
        if dim.local_representation and dim.local_representation.enumerated:
            cl = dim.local_representation.enumerated 
            df = sdmx.to_pandas(cl)
            codelists[dim.id][cl.id] = df
      
    for _, series in codelists.items():
        logger.debug("Codelists\n%s", series)
    return codelists     
    
# generates the datakey needed for a dataflowid.
# datakeys are vary widely in scope for each dataflowid, thus a method to get the metadata for the dataflowid, then determine a key from there.
# Basic understanding:
#   INDEX: 10001 is for top level data, will default to that for now and maybe allow a drop down for the different indexs later or something
#   FREQ: single letter represiatnions, will default to Q for quater - might need to do a prio list or something - or let user decide later, will aftect the TIME_PERIOD column 
def key_generator(dims: tuple, codelists: dict, key_map: dict = {}) -> str:
    logger.info('Generating Key String')
    
    # override values to protect bad inputs
    key_map_override = {
    'FREQ': 'A',
    'FREQUENCY':'A',
    }
    
    # Set default map to first item in codelist   
    if not key_map:
        logger.info('No key map supplied, creating default map')
        for dimname, cl in codelists.items():
            for codeName, codes in cl.items():
                if dimname in key_map_override:
                    key_map[dimname] = key_map_override[dimname]
                else:
                    key_map[dimname] = codes.index[0]  
        logger.debug(f'Default key map: {key_map}')        
    
    # Sort by the dims order - sanity check 
    keys = {k: key_map[k] for k in dims if k in key_map}

     # Construct ABS datakey string 
    key_string: str = "" 
    for _,v in keys.items(): 
        key_string += f'{v}.'
    logger.info("Key String Generated: %s", key_string)
    
    return key_string


from config import load_config
config = load_config()  
df = get_data('ALC')
print(df)

