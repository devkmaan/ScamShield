# scamshield-bot-gateway

Owns WhatsApp webhook verification, inbound normalization, idempotency, user rate limiting, and outbound WhatsApp sending.

Current MVP implementation lives in:

- `internal/api/server.go`
- `internal/api/whatsapp.go`

Future extraction steps:

1. Move WhatsApp webhook handlers here.
2. Publish `whatsapp.inbound.v1`.
3. Consume `whatsapp.reply.requested.v1`.
4. Replace local outbox with WhatsApp Cloud API send calls.

