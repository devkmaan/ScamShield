package analysis

import (
	"strings"

	"scamshield/internal/domain"
)

// ScoreTextModel is a lightweight, explainable model-shaped scorer for the MVP.
// It is intentionally deterministic so tests can assert behavior; a trained model
// service can replace this behind the same function boundary later.
func ScoreTextModel(entities domain.ExtractedEntities) []domain.Signal {
	text := entities.NormalizedText
	if text == "" {
		return nil
	}

	var signals []domain.Signal
	categoryScores := []struct {
		id       string
		terms    []string
		scamType domain.ScamType
	}{
		{"ml_impersonation_probability", []string{"bank", "rbi", "police", "customer care", "support", "account", "verification"}, domain.ScamImpersonation},
		{"ml_phishing_probability", []string{"click", "link", "verify", "login", "password", "kyc", "blocked"}, domain.ScamPhishing},
		{"ml_job_scam_probability", []string{"task", "salary", "work from home", "telegram", "deposit", "commission"}, domain.ScamJob},
		{"ml_investment_probability", []string{"profit", "trading", "crypto", "return", "double", "signal", "portfolio"}, domain.ScamInvestment},
	}

	for _, category := range categoryScores {
		hits := 0
		for _, term := range category.terms {
			if strings.Contains(text, term) {
				hits++
			}
		}
		if hits >= 3 {
			weight := 0.22 + float64(hits-3)*0.04
			if weight > 0.34 {
				weight = 0.34
			}
			signals = append(signals, domain.Signal{
				ID:          category.id,
				Description: "Text classifier found a dense cluster of scam-like terms.",
				Weight:      weight,
				ScamType:    category.scamType,
				Reason:      "The message combines multiple terms commonly seen in this scam category.",
			})
		}
	}
	if containsObfuscatedUrgency(text) {
		signals = append(signals, domain.Signal{
			ID:          "obfuscated_pressure_text",
			Description: "Message uses obfuscated urgency or spacing.",
			Weight:      0.16,
			ScamType:    domain.ScamUnknown,
			Reason:      "Scammers often obfuscate urgent words to evade filters.",
		})
	}
	return signals
}

func containsObfuscatedUrgency(text string) bool {
	compacted := strings.ReplaceAll(text, " ", "")
	return strings.Contains(compacted, "u r g e n t") ||
		strings.Contains(compacted, "verifyimmediately") ||
		strings.Contains(compacted, "blockedin24")
}
