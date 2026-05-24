# ScamShield Contracts and Events

## 1. Contract Principles

- Contracts are versioned before implementation.
- Every event has a schema and at least one valid fixture.
- Consumers must ignore unknown fields.
- Producers must keep deprecated fields until all consumers migrate.
- Final risk decisions are immutable; later review outcomes are separate events.

## 2. Canonical Event Envelope

```json
{
  "eventId": "evt-01H...",
  "eventType": "risk.decision.created",
  "schemaVersion": "1.0.0",
  "correlationId": "corr-01H...",
  "causationId": "evt-01G...",
  "createdAt": "2026-04-25T08:00:00Z",
  "producer": "scamshield-risk-core",
  "payload": {}
}
```

Required envelope fields:

- `eventId`: globally unique event ID.
- `eventType`: stable event name.
- `schemaVersion`: semantic schema version.
- `correlationId`: one ID across a user flow.
- `causationId`: upstream event ID that caused this event.
- `createdAt`: UTC timestamp.
- `producer`: service name.
- `payload`: event-specific body.

## 3. Core Topics

### `whatsapp.inbound.v1`

Produced by `scamshield-bot-gateway`.

Payload:

```json
{
  "messageId": "wamid.demo",
  "userHash": "usr_hash",
  "messageType": "TEXT",
  "text": "KYC update urgent. Share OTP.",
  "mediaRef": null,
  "receivedAt": "2026-04-25T08:00:00Z"
}
```

### `risk.decision.created.v1`

Produced by `scamshield-risk-core`.

Payload:

```json
{
  "decisionId": "dec-123",
  "userHash": "usr_hash",
  "inputType": "TEXT",
  "riskLevel": "CRITICAL",
  "score": 0.92,
  "confidence": 0.94,
  "scamType": "PHISHING",
  "topSignals": ["otp_request", "brand_spoofing"],
  "recommendedActions": [
    "Do not share OTP, UPI PIN, card details, screen, or remote access."
  ],
  "needsHumanReview": false,
  "modelVersions": {
    "text": "rules-plus-v1",
    "url": "lexical-v1"
  }
}
```

### `whatsapp.reply.requested.v1`

Produced by `scamshield-risk-core`.

Consumed by `scamshield-bot-gateway`.

Payload:

```json
{
  "userHash": "usr_hash",
  "replyType": "RISK_VERDICT",
  "text": "Critical risk lag raha hai. Payment mat karo.",
  "buttons": ["Scam", "Not Scam", "Need Help"],
  "decisionId": "dec-123"
}
```

### `feedback.received.v1`

Produced by `scamshield-bot-gateway` or Admin API.

Payload:

```json
{
  "feedbackId": "fb-123",
  "decisionId": "dec-123",
  "userHash": "usr_hash",
  "verdict": "SCAM",
  "payeeHash": "payee_hash",
  "comment": "Asked me to enter UPI PIN to receive refund."
}
```

### `merchant.risk.updated.v1`

Produced by `scamshield-merchant-risk`.

Payload:

```json
{
  "payeeHash": "payee_hash",
  "riskScore": 0.84,
  "complaintCount": 6,
  "reviewStatus": "NEEDS_REVIEW",
  "topSignals": ["complaint_velocity", "alias_similarity"]
}
```

### `recovery.report.created.v1`

Produced by `scamshield-risk-core` or Recovery Report Service.

Payload:

```json
{
  "reportId": "rep-123",
  "userHash": "usr_hash",
  "decisionId": "dec-123",
  "status": "DRAFT_GUIDANCE_ONLY",
  "officialHelp": [
    "Call 1930",
    "File at https://cybercrime.gov.in",
    "Contact your bank or payment app support"
  ]
}
```

## 4. Internal HTTP Contracts

### Score Text

`POST /internal/model/score-text`

Request:

```json
{
  "text": "Your KYC is blocked. Share OTP.",
  "languageHint": "hinglish",
  "context": {
    "inputType": "TEXT"
  }
}
```

Response:

```json
{
  "modelVersion": "text-scam-v1",
  "score": 0.88,
  "confidence": 0.81,
  "scamTypeScores": {
    "PHISHING": 0.88,
    "IMPERSONATION": 0.72
  },
  "signals": ["kyc_pressure", "otp_request"]
}
```

### Observe Payee

`POST /internal/payee/observe`

Request:

```json
{
  "rawPayee": "merchant@upi",
  "alias": "Merchant Name",
  "source": "QR",
  "correlationId": "corr-123"
}
```

Response:

```json
{
  "payeeHash": "payee_hash",
  "riskScore": 0.18,
  "complaintCount": 0,
  "needsHumanReview": false
}
```

The raw payee must not be logged in plaintext.

## 5. Compatibility Rules

- New optional fields are safe.
- New enum values require all consumers to have fallback handling.
- Removing fields requires a major schema version.
- Changing risk threshold behavior requires release notes even if the schema is unchanged.
- Event replay must be supported for all core events.

