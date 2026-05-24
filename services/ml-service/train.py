from __future__ import annotations

import json
from pathlib import Path

from model_core import build_model

ROOT = Path(__file__).resolve().parents[2]
DATASET_PATH = ROOT / "data" / "ml_training_samples.jsonl"


def main() -> None:
    model = build_model(DATASET_PATH)
    print(json.dumps({
        "modelVersion": model.model_version,
        "backend": model.backend_name,
        "embeddingModel": model.embedding_model,
        "fallbackReason": model.fallback_reason,
        "dataset": str(DATASET_PATH),
        "evaluation": model.evaluation.__dict__,
    }, indent=2))


if __name__ == "__main__":
    main()
