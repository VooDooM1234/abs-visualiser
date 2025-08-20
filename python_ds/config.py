import os
from dotenv import load_dotenv

def load_config():
    load_dotenv()

    config = {
        "DB_NAME": os.getenv("DATABASE_NAME"),
        "DB_USER": os.getenv("DATABASE_USER"),
        "DB_PASSWORD": os.getenv("DATABASE_PASSWORD"),
        "DB_HOST": os.getenv("DATABASE_HOST"),
    }

    if not all(config.values()):
        raise RuntimeError("Missing one or more required environment variables")

    return config
