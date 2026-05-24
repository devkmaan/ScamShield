# ScamShield Service Split

This directory records the planned production service boundaries while the runnable MVP still lives in the root Go application.

Each subdirectory should become its own repository when the contracts stabilize:

- `bot-gateway`
- `risk-core`
- `ml-service`
- `genai-service`
- `merchant-risk`
- `evidence-service`
- `admin-web`

The source of truth for APIs and events is `../contracts`.
