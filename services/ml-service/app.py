from __future__ import annotations

from pathlib import Path
from typing import Any

from fastapi import FastAPI
from pydantic import BaseModel, Field

from model_core import URL_MODEL_VERSION, backend_requested, build_model, score_url

ROOT = Path(__file__).resolve().parents[2]
DATASET_PATH = ROOT / "data" / "ml_training_samples.jsonl"
MODEL = build_model(DATASET_PATH)

app = FastAPI(title="ScamShield ML Service", version="1.0.0")


class ScoreTextRequest(BaseModel):
    text: str = ""
    languageHint: str | None = None
    context: dict[str, str] = Field(default_factory=dict)


class ScoreURLRequest(BaseModel):
    url: str = ""
    text: str = ""
    context: dict[str, str] = Field(default_factory=dict)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "scamshield-ml-service"}


@app.post("/internal/model/score-text")
def score_text(req: ScoreTextRequest) -> dict[str, Any]:
    return MODEL.predict_text(req.text)


@app.post("/internal/model/score-url")
def score_url_endpoint(req: ScoreURLRequest) -> dict[str, Any]:
    return score_url(req.url, req.text)


@app.get("/internal/model/metadata")
def metadata() -> dict[str, Any]:
    model_meta = MODEL.metadata()
    return {
        "mode": "python-fastapi-ml",
        "requestedBackend": backend_requested(),
        "activeModels": {
            "text": model_meta["modelVersion"],
            "url": URL_MODEL_VERSION,
            "calibration": model_meta["calibrationVersion"],
        },
        "backend": model_meta["backend"],
        "embeddingModel": model_meta["embeddingModel"],
        "fallbackReason": model_meta["fallbackReason"],
        "dataset": {
            "path": str(DATASET_PATH),
            "focus": "English + Hinglish scam/safe examples",
        },
        "evaluation": MODEL.evaluation.__dict__,
        "policy": "Go risk-core remains final decision-maker; LLM/explainer cannot lower high-risk decisions.",
    }


@app.get("/internal/model/evaluate")
def evaluate() -> dict[str, Any]:
    return MODEL.evaluation.__dict__
