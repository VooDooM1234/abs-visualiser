import psycopg
from typing import Any

def init_db(config: dict) -> Any:
    conn = psycopg.connect(
        dbname=config["DB_NAME"],
        user=config["DB_USER"],
        password=config["DB_PASSWORD"],
        host=config["DB_HOST"]
    )
    return conn
