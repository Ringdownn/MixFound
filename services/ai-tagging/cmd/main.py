import logging
import sys
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI
from sympy import false

from ai_tagging.app.api.routes import router, init_tagging_service
from ai_tagging.app.config import settings
from ai_tagging.app.services.tagging import TaggingService

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Starting AI Tagging Service...")
    tagging_service = await TaggingService.create()
    init_tagging_service(tagging_service)
    yield
    logger.info("Stopping AI Tagging Service...")

app = FastAPI(
    title="AI Tagging Service",
    version="1.0",
    lifespan=lifespan
)

app.include_router(router)

if __name__ == "__main__":
    uvicorn.run(
        "ai_tagging.app.main:app",
        host=settings.API_HOST,
        port=settings.API_PORT,
        reload=False,
    )
