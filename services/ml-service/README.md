# scamshield-ml-service

Python FastAPI service for ScamShield text and URL risk scoring.

The default backend now tries to train a MiniLM embedding classifier from
`data/ml_training_samples.jsonl`:

- embeddings: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`
- classifier: scikit-learn logistic regression
- calibration: deterministic holdout bucket calibration

If SentenceTransformers, Torch, scikit-learn, or the model weights are not
available, the service falls back to the original lightweight Naive Bayes
classifier and reports the fallback reason in `/internal/model/metadata`.
Go still remains the final fraud decision-maker; this service only contributes
model signals.

## Run

```powershell
py -3 -m pip install -r .\services\ml-service\requirements.txt
py -3 .\services\ml-service\train.py
py -3 -m uvicorn app:app --app-dir .\services\ml-service --host 127.0.0.1 --port 8090
```

From the service directory:

```powershell
py -3 -m uvicorn app:app --host 127.0.0.1 --port 8090
```

## Endpoints

- `POST /internal/model/score-text`
- `POST /internal/model/score-url`
- `GET /internal/model/metadata`
- `GET /internal/model/evaluate`

## Backend Controls

```powershell
# Default: try MiniLM, then fall back to Naive Bayes if unavailable.
$env:SCAMSHIELD_ML_BACKEND = "auto"

# Force the lightweight fallback for fast tests/offline development.
$env:SCAMSHIELD_ML_BACKEND = "naive_bayes"

# Require MiniLM and fail startup if dependencies or weights are missing.
$env:SCAMSHIELD_ML_BACKEND = "minilm"
$env:SCAMSHIELD_REQUIRE_MINILM = "true"

# Swap the embedding model later without changing the Go contract.
$env:SCAMSHIELD_EMBEDDING_MODEL = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
```

If `pip install` fails on Python 3.13 because Torch wheels are unavailable,
create a Python 3.11 or 3.12 virtual environment for this service and install
the same requirements there.
