from typing import Optional

from ai_tagging.app.models.clip_model import CLIPService


class TagResult:
    def __init__(self, label: str, confidence: float):
        self.label = label
        self.confidence = confidence

    def to_dict(self):
        return {"label": self.label, "confidence": float(self.confidence)}

class TaggingService:
    def __init__(self, clip_service: CLIPService, labels: list[str]):
        self.clip_service = clip_service
        self.labels = labels

    @classmethod
    async def create(cls, labels_file: Optional[str]= None) -> "TaggingService":
        pass

