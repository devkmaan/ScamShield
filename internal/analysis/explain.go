package analysis

import (
	"fmt"
	"strings"

	"scamshield/internal/domain"
)

func buildUserMessage(decision domain.RiskDecision, signals []domain.Signal) string {
	level := decision.RiskLevel
	var tone string
	switch level {
	case domain.RiskCritical:
		tone = "Critical risk lag raha hai. Payment mat karo."
	case domain.RiskHigh:
		tone = "High risk scam signals mile hain. Please pause before paying."
	case domain.RiskCaution:
		tone = "Kuch warning signs mile hain. Verify independently before acting."
	default:
		tone = "Major scam signals nahi mile, but payment se pehle details verify kar lo."
	}

	reasons := topReasonText(signals, 3)
	if len(reasons) == 0 {
		return tone
	}
	return fmt.Sprintf("%s Red flags: %s.", tone, strings.Join(reasons, "; "))
}

func recommendedActions(decision domain.RiskDecision) []string {
	switch decision.RiskLevel {
	case domain.RiskCritical, domain.RiskHigh:
		return []string{
			"Do not share OTP, UPI PIN, card details, screen, or remote access.",
			"Do not pay through QR/UPI collect if someone says it is needed to receive money.",
			"Verify using the official bank/app number or app, not the number/link in the message.",
			"If money is already lost, contact your bank, call 1930, and file at cybercrime.gov.in.",
		}
	case domain.RiskCaution:
		return []string{
			"Verify the sender through an official app, website, or known phone number.",
			"Open links only from official domains typed manually.",
			"Send more context, screenshot, QR, or UPI ID if you want a stronger check.",
		}
	default:
		return []string{
			"Still verify payee name, amount, and reason before payment.",
			"Never enter UPI PIN to receive money.",
		}
	}
}

func topReasonText(signals []domain.Signal, limit int) []string {
	var reasons []string
	seen := map[string]bool{}
	for _, signal := range signals {
		reason := signal.Reason
		if reason == "" {
			reason = signal.Description
		}
		if reason == "" || seen[reason] {
			continue
		}
		seen[reason] = true
		reasons = append(reasons, reason)
		if len(reasons) == limit {
			break
		}
	}
	return reasons
}
