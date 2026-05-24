# scamshield-admin-web

Owns the analyst dashboard, manual review queue, feedback triage, payee risk exploration, and model quality panels.

Current MVP implementation is the simple HTML page at:

- `internal/api/admin.go`

Future extraction steps:

1. Build a Next.js/React admin app.
2. Add auth and role-based access.
3. Consume Admin API endpoints from `risk-core` and `merchant-risk`.

