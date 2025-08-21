import os
from dotenv import load_dotenv
from urllib.parse import urlparse

def load_config():
    load_dotenv()

    db_url = os.getenv("DATABASE_URL")

    parsed = urlparse(db_url)

    db_user = parsed.username
    db_password = parsed.password
    db_host = parsed.hostname
    db_port = parsed.port
    db_name = parsed.path.lstrip("/")  # removes leading "/"

    config = {
        "DB_NAME": db_name,
        "DB_USER": db_user,
        "DB_PASSWORD": db_password,
        "DB_HOST": db_host,
        "DB_PORT": db_port,
    }

    if not all(config.values()):
        raise RuntimeError("Missing one or more required environment variables")

    return config