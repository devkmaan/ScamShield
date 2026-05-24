package domain

import "time"

const (
	EventWhatsAppInbound        = "whatsapp.inbound"
	EventRiskDecisionCreated    = "risk.decision.created"
	EventWhatsAppReplyRequested = "whatsapp.reply.requested"
	EventFeedbackReceived       = "feedback.received"
	EventMerchantRiskUpdated    = "merchant.risk.updated"
	EventRecoveryReportCreated  = "recovery.report.created"
)

type EventEnvelope struct {
	EventID       string    `json:"eventId"`
	EventType     string    `json:"eventType"`
	SchemaVersion string    `json:"schemaVersion"`
	CorrelationID string    `json:"correlationId"`
	CausationID   string    `json:"causationId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Producer      string    `json:"producer"`
	Payload       any       `json:"payload"`
}

type EventSink interface {
	AppendEvent(event EventEnvelope)
}
