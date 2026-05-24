# Full App Implementation

This folder is the fuller application implementation built from the original MVP and architecture docs.

## Implemented Product Surface

- WhatsApp-style webhook ingestion and local outbox replies.
- Manual risk checking for text, URL, UPI ID, QR payload, and screenshots/media references.
- Recovery flow for already-paid scam cases.
- Evidence metadata storage with OTP, PIN, and card redaction.
- Merchant/payee risk observe and report APIs.
- Internal model scoring APIs for text and URL risk, backed by the Python ML service when `ML_SERVICE_URL` is set.
- Admin dashboard APIs for decisions, merchants, feedback, reports, evidence, and event logs.
- Synthetic stream generator for demos and dashboard population.
- Single-page operational web console at `/app`, `/admin`, and `/`.

## Implemented Architecture Surface

- Event envelope logging with correlation IDs.
- Webhook idempotency by WhatsApp message ID.
- Per-user rate limiting.
- Privacy-safe salted payee hashing.
- Contract scaffolding through OpenAPI, AsyncAPI, JSON schemas, and fixtures.
- PostgreSQL migration baseline for the eventual persistence layer.
- Split-ready service folders for bot gateway, risk core, ML service, merchant risk, evidence service, and admin web.

## What Is Still Simulated

- Real WhatsApp outbound sending is represented by `/v1/outbox`.
- OCR and QR image decoding are represented by text/media payload parsing and QR intent payload parsing.
- Text and URL model endpoints are implemented by `services/ml-service` when running; Go falls back to deterministic local scorers if the ML service is unavailable.
- Kafka, Postgres, Redis, and MinIO are scaffolded through contracts/migrations/compose, while the runnable app uses in-memory adapters.

## Run

```powershell
py -3 -m pip install -r .\services\ml-service\requirements.txt
py -3 .\services\ml-service\train.py
.\scripts\run-ml-service.ps1
$env:ML_SERVICE_URL = "http://localhost:8090"
go test ./...
go run ./cmd/scamshield
```

Open:

- React frontend: `http://localhost:5173`
- Go backend: `http://localhost:8081/ready`
- ML service: `http://localhost:8090/internal/model/metadata`
