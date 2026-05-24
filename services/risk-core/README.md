# scamshield-risk-core

Owns scam orchestration, deterministic rules, risk aggregation, final verdict policy, and recovery report coordination.

Current MVP implementation lives in:

- `internal/analysis`
- `internal/domain`

Future extraction steps:

1. Consume `whatsapp.inbound.v1`.
2. Call ML, Merchant Risk, Evidence, and Explanation adapters.
3. Persist risk decisions in PostgreSQL.
4. Publish `risk.decision.created.v1` and `whatsapp.reply.requested.v1`.

