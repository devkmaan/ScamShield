# scamshield-merchant-risk

Owns salted payee hashing, payee risk profiles, complaint features, alias clustering, and manual review state.

Current MVP implementation lives in:

- `internal/store/memory.go`

Future extraction steps:

1. Move payee hashing into this service boundary.
2. Persist `payee_risk_profiles`.
3. Consume `feedback.received.v1`.
4. Publish `merchant.risk.updated.v1`.

