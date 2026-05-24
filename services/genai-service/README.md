# scamshield-genai-service

Local-first multilingual GenAI service for ScamShield.

The service uses an Ollama/OpenAI-compatible endpoint when available and falls back to safe deterministic output when a local model is not running. Go remains the final risk decision-maker; this service only normalizes text and renders user-facing copy.

## Run

Install Ollama separately, then pull the default model:

```powershell
ollama pull qwen2.5:3b
```

Start the service:

```powershell
py -3 -m pip install -r .\services\genai-service\requirements.txt
py -3 -m uvicorn app:app --app-dir .\services\genai-service --host 127.0.0.1 --port 8091
```

Useful env vars:

- `OLLAMA_BASE_URL=http://localhost:11434/v1`
- `GENAI_MODEL=qwen2.5:3b`
- `GENAI_TIMEOUT_MS=3500`

## Endpoints

- `POST /internal/genai/normalize-input`
- `POST /internal/genai/render`
- `POST /internal/genai/chat`
- `POST /internal/genai/ui-bundle`
- `GET /internal/genai/languages`
- `GET /internal/genai/metadata`
