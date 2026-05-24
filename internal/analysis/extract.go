package analysis

import (
	"net/url"
	"regexp"
	"strings"

	"scamshield/internal/domain"
)

var (
	urlPattern    = regexp.MustCompile(`(?i)\b((?:https?://|www\.)[^\s<>"']+)`)
	upiPattern    = regexp.MustCompile(`(?i)\b[a-z0-9._-]{2,}@[a-z0-9._-]{2,}\b`)
	amountPattern = regexp.MustCompile(`(?i)(?:rs\.?|inr|₹)\s*([0-9][0-9,]*(?:\.[0-9]{1,2})?)`)
	shorteners    = map[string]bool{
		"bit.ly": true, "tinyurl.com": true, "t.co": true, "goo.gl": true, "is.gd": true,
		"cutt.ly": true, "shorturl.at": true, "rebrand.ly": true, "wa.link": true,
	}
)

func Extract(req domain.CheckRequest) domain.ExtractedEntities {
	textParts := []string{req.Text, req.URL, req.UPIID, req.QRPayload}
	normalized := normalizeText(strings.Join(textParts, " "))

	entities := domain.ExtractedEntities{
		NormalizedText: normalized,
		URLs:           extractURLs(strings.Join(textParts, " ")),
		UPIIDs:         uniqueStrings(upiPattern.FindAllString(strings.Join(textParts, " "), -1)),
		Amounts:        uniqueStrings(amountPattern.FindAllString(strings.Join(textParts, " "), -1)),
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(req.QRPayload)), "upi://") {
		entities.QR = parseUPIQR(req.QRPayload)
		if entities.QR != nil && entities.QR.PayeeUPI != "" {
			entities.UPIIDs = appendUnique(entities.UPIIDs, entities.QR.PayeeUPI)
		}
	}
	return entities
}

func normalizeText(raw string) string {
	value := strings.ToLower(raw)
	replacements := map[string]string{
		"₹":       " inr ",
		"otp":     " otp ",
		"upi pin": " upi pin ",
	}
	for old, newValue := range replacements {
		value = strings.ReplaceAll(value, old, newValue)
	}
	value = regexp.MustCompile(`\s+`).ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func extractURLs(raw string) []domain.URLFinding {
	matches := urlPattern.FindAllString(raw, -1)
	var findings []domain.URLFinding
	seen := map[string]bool{}
	for _, match := range matches {
		clean := strings.TrimRight(match, ".,);]")
		parseTarget := clean
		if strings.HasPrefix(strings.ToLower(parseTarget), "www.") {
			parseTarget = "https://" + parseTarget
		}
		parsed, err := url.Parse(parseTarget)
		if err != nil || parsed.Host == "" {
			continue
		}
		host := strings.ToLower(parsed.Hostname())
		if seen[clean] {
			continue
		}
		seen[clean] = true
		findings = append(findings, domain.URLFinding{
			Raw:         clean,
			Host:        host,
			Scheme:      parsed.Scheme,
			IsShortener: shorteners[host],
		})
	}
	return findings
}

func parseUPIQR(raw string) *domain.QRFinding {
	parsed, err := url.Parse(raw)
	if err != nil {
		return &domain.QRFinding{RawPayload: raw}
	}
	query := parsed.Query()
	return &domain.QRFinding{
		RawPayload: raw,
		PayeeUPI:   strings.TrimSpace(query.Get("pa")),
		PayeeName:  strings.TrimSpace(query.Get("pn")),
		Amount:     strings.TrimSpace(query.Get("am")),
		Note:       strings.TrimSpace(query.Get("tn")),
	}
}

func uniqueStrings(values []string) []string {
	var result []string
	for _, value := range values {
		result = appendUnique(result, strings.TrimSpace(value))
	}
	return result
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return values
		}
	}
	return append(values, value)
}
