import json
from typing import Optional
import asyncio
import logging

import numpy as np

from ai_tagging.app.config import settings
from ai_tagging.app.models.clip_model import CLIPService
from ai_tagging.app.services.image_loader import ImageLoader

logger = logging.getLogger(__name__)

class TagResult:
    def __init__(self, label: str, confidence: float):
        self.label = label
        self.confidence = confidence

    def to_dict(self):
        return {"label": self.label, "confidence": float(self.confidence)}

class TaggingService:
    def __init__(self, clip_service: CLIPService, labels: list[str]):
        self.clip = clip_service
        self.labels = labels

    @classmethod
    async def create(cls, labels_file: Optional[str]= None) -> "TaggingService":
        labels_file = labels_file or settings.LABEL_FILE
        
        # 在线程池中加载模型，避免阻塞事件循环
        loop = asyncio.get_event_loop()
        clip_service = await loop.run_in_executor(
            None,
            lambda: CLIPService(settings.MODEL_NAME)
        )

        # 读取标签文件
        with open(labels_file, "r", encoding="utf-8") as f:
            labels = json.load(f)

        return cls(clip_service, labels)

    async def tag_image_from_url(self, image_url: str, top_k: int = 3) -> list[TagResult]:
        image = await ImageLoader.load_image_from_url(image_url, settings.REQUEST_TIMEOUT)

        if image is None:
            logger.error(f"Failed to load image from {image_url}")
            return []
        
        # 编码图片
        image_features = self.clip.encode_image(image)
        # 编码所有标签文本
        text_features = self.clip.encode_text(self.labels)
        
        # 计算相似度
        similarity = self.clip.get_similarity(image_features, text_features)
        top_indices = similarity.argsort()[::-1][:top_k]

        result = []

        for index in top_indices:
            result.append(TagResult(
                label=self.labels[index],
                confidence=similarity[index])
            )

        return result