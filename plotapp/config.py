import os
import sys
import json
import logging
import logging.config
from dotenv import load_dotenv
from urllib.parse import urlparse

class StreamToLogger:
    def __init__(self, logger, level):
        self.logger = logger
        self.level = level
    def write(self, message):
        if message.strip():
            self.logger.log(self.level, message.strip())
    def flush(self):
        pass
    def isatty(self):
        return True

def load_config():
    load_dotenv()

    LOGGING_CONFIG = {
        "version": 1,
        "disable_existing_loggers": False,
        "formatters": {
            "default": {
                "()": "uvicorn.logging.DefaultFormatter",
                "format": "%(levelprefix)s %(message)s",
                "use_colors": True,
            },
        },
        "handlers": {
            "default": {
                "formatter": "default",
                "class": "logging.StreamHandler",
                "stream": "ext://sys.stdout",
            },
        },
        "loggers": {
            "uvicorn.error": {"handlers": ["default"], "level": "DEBUG"},
            "uvicorn.access": {"handlers": ["default"], "level": "INFO"},
            "dash": {"handlers": ["default"], "level": "DEBUG"},
            "flask.app": {"handlers": ["default"], "level": "DEBUG"},
            "sdmx": {"handlers": ["default"], "level": "DEBUG"},
            "main": {"handlers": ["default"], "level": "DEBUG"},
        },
    }

    logging.config.dictConfig(LOGGING_CONFIG)

    # Load environment variables
    db_url = os.getenv("DATABASE_URL")
    if not db_url:
        raise RuntimeError("DATABASE_URL not set")

    parsed = urlparse(db_url)
    db_user = parsed.username
    db_password = parsed.password
    db_host = parsed.hostname
    db_port = parsed.port
    db_name = parsed.path.lstrip("/")

    # Load config.json
    with open("config.json", "r") as f:
        config_json = json.load(f)
        dash_port = config_json.get("dash_port")

    config = {
        "DB_NAME": db_name,
        "DB_USER": db_user,
        "DB_PASSWORD": db_password,
        "DB_HOST": db_host,
        "DB_PORT": db_port,
        "DASH_PORT": dash_port,
    }

    if not all(config.values()):
        raise RuntimeError("Missing one or more required environment variables or config values")

    return config
