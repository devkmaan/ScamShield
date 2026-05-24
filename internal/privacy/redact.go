package privacy

import "regexp"

var (
	otpPattern  = regexp.MustCompile(`(?i)\b(?:otp|one[- ]?time password|verification code)\s*(?:is|:|-)?\s*[0-9]{4,8}\b`)
	pinPattern  = regexp.MustCompile(`(?i)\b(?:upi\s*)?pin\s*(?:is|:|-)?\s*[0-9]{4,8}\b`)
	cardPattern = regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)
)

func RedactSensitive(raw string) string {
	value := otpPattern.ReplaceAllString(raw, "[REDACTED_OTP]")
	value = pinPattern.ReplaceAllString(value, "[REDACTED_PIN]")
	value = cardPattern.ReplaceAllStringFunc(value, func(match string) string {
		digits := 0
		for _, r := range match {
			if r >= '0' && r <= '9' {
				digits++
			}
		}
		if digits >= 13 {
			return "[REDACTED_CARD]"
		}
		return match
	})
	return value
}
