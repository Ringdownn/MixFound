import os
import torch
from typing import Optional
from pydantic_settings import BaseSettings

def get_device() -> str:
    if torch.cuda.is_available():
        return "cuda"
    elif torch.backends.mps.is_available():
        return "mps"
    else:
        return "cpu"

class Settings(BaseSettings):
    MODEL_NAME: str = ""
    LABEL_FILE: str = ""
    DEFAULT_TOP_K: int = 3
    MAX_IMAGE_SIZE: int = 512
    DEVICE: str = get_device()
    API_HOST: str = ""
    API_PORT: int = 8080
    REQUEST_TIMEOUT: int = 30

    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()