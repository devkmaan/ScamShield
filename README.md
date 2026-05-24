# ScamShield

ScamShield is a WhatsApp-first India MVP for real-time fraud and scam detection. It accepts suspicious messages, URLs, UPI IDs, UPI QR payloads, payment screenshots/media references, and "already paid" recovery requests, then returns a risk verdict with multilingual guidance.

This implementation is a runnable Go backend that mirrors the planned architecture with in-memory infrastructure. The service boundaries are explicit so Kafka/Redpanda, PostgreSQL, Redis, MinIO, and a Python model service can replace local adapters later.

This `ScamShield_FullApp` folder is the expanded application build. It includes the Go API, separate React frontend, contracts, migration baseline, service split notes, simulator, evidence flow, model endpoints, and admin dashboard APIs.

## What Is Implemented

- `POST /webhooks/whatsapp` for mocked WhatsApp Cloud API inbound messages.
- `GET /webhooks/whatsapp` verification challenge support.
- `POST /v1/check` for synchronous scam checks.
- `POST /v1/feedback` for `Scam`, `Not Scam`, and `Need Help` feedback.
- `GET /v1/risk/payee/{payeeHash}` for privacy-safe payee risk lookup.
- `GET /v1/reports/{reportId}` for recovery checklist drafts.
- `GET /v1/outbox` for local WhatsApp reply inspection.
- `GET /admin` for a tiny local analyst console.
- Hybrid risk pipeline: rules, Python ML scoring, URL analysis, UPI/QR extraction, merchant graph signals, and bounded multilingual GenAI rendering.
- Local-first GenAI service for language normalization, localized explanations, recovery wording, chat replies, and generated UI bundles.

## Run Locally

```powershell
go test ./...
go run ./cmd/scamshield
```

Open the Go-served fallback console only when running the backend directly:

- API health: `http://localhost:8081/health`
- Go fallback console: `http://localhost:8081/admin`

## React Frontend

Run the ML service and GenAI service, then the Go backend on port `8081`, then start the separate frontend:

```powershell
py -3 -m pip install -r .\services\ml-service\requirements.txt
.\scripts\run-ml-service.ps1

py -3 -m pip install -r .\services\genai-service\requirements.txt
.\scripts\run-genai-service.ps1

$env:ML_SERVICE_URL = "http://localhost:8090"
$env:GENAI_SERVICE_URL = "http://localhost:8091"
.\scripts\run.ps1

cd .\web
npm install
npm run dev
```

Open:

- React app: `http://localhost:5173`
- Go backend: `http://localhost:8081/ready`
- ML service: `http://localhost:8090/internal/model/metadata`
- GenAI service: `http://localhost:8091/internal/genai/metadata`

The React app uses Vite proxy rules in `web/vite.config.js`, so browser API calls to `/v1`, `/internal`, and `/webhooks` are forwarded to the Go backend.

The ML service defaults to MiniLM embeddings plus a calibrated logistic
classifier when SentenceTransformers/Torch/scikit-learn and the MiniLM weights
are available. If that stack is not installed or the model cannot be loaded, it
falls back to the lightweight Naive Bayes classifier and exposes the reason at
`http://localhost:8090/internal/model/metadata`. For fast offline runs, set:

```powershell
$env:SCAMSHIELD_ML_BACKEND = "naive_bayes"
```

For real local multilingual generation, install Ollama and pull the default model:

```powershell
ollama pull qwen2.5:3b
```

If Ollama is not running, the GenAI service returns safe fallback text and the Go risk pipeline still works.
For local `qwen2.5:3b` testing on modest hardware, keep `GENAI_TIMEOUT_MS` and
`GENAI_CLIENT_TIMEOUT_MS` around `35000`; lower values will intentionally fall
back to safe templates when the model is slow.

## Try a Manual Check

```powershell
$body = @{
  inputType = "TEXT"
  userId = "demo-user"
  language = "hi"
  text = "Your SBI KYC is blocked. Click https://sbi-verify-support.com and share OTP immediately."
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri http://localhost:8081/v1/check -ContentType "application/json" -Body $body
```

## Try the WhatsApp Webhook Mock

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8081/webhooks/whatsapp `
  -ContentType "application/json" `
  -InFile .\data\mock_whatsapp_webhook.json

Invoke-RestMethod http://localhost:8081/v1/outbox?userId=919999999999
```

## Recovery Flow

```powershell
$body = @{
  inputType = "TEXT"
  userId = "demo-user"
  alreadyPaid = $true
  text = "I already paid 5000 after a Telegram task commission message to taskbonus@paytm"
} | ConvertTo-Json

$decision = Invoke-RestMethod -Method Post -Uri http://localhost:8081/v1/check -ContentType "application/json" -Body $body
Invoke-RestMethod "http://localhost:8081/v1/reports/$($decision.reportId)"
```

## Production Upgrade Path

- Replace the in-memory event channel with Kafka/Redpanda topics:
  - `whatsapp.inbound`
  - `risk.decisions`
  - `feedback.received`
  - `merchant.risk.updated`
- Replace `MemoryStore` with PostgreSQL repositories and Redis hot cache.
- Move `ScoreTextModel` behind a Python FastAPI model service.
- Add real OCR/QR media extraction for WhatsApp media downloads.
- Extend the GenAI adapter with stronger local Indic translation models, while keeping Go as the final risk decision-maker.
- Add WhatsApp Cloud API outbound sending instead of local `/v1/outbox`.

## Production Scaffolding Now Present

- API contracts: `contracts/openapi`
- Event contracts: `contracts/asyncapi`
- JSON schemas and fixtures: `contracts/schemas`, `contracts/fixtures`
- PostgreSQL migration baseline: `infra/postgres/migrations/001_init.sql`
- Split-ready service notes: `services`
- Local scripts: `scripts/test.ps1`, `scripts/run.ps1`, `scripts/local-up.ps1`
- React frontend: `web`
- Environment template: `.env.example`

## Architecture Planning Docs

- [Full production architecture plan](docs/full-architecture-plan.md)
- [Multi-repository decomposition plan](docs/multi-repository-plan.md)
- [Contracts and event plan](docs/contracts-and-events.md)
- [Full app implementation notes](docs/full-app-implementation.md)
- [Current MVP architecture notes](docs/architecture.md)
