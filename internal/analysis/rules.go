package analysis

import (
	"strings"

	"scamshield/internal/domain"
)

type textRule struct {
	id          string
	description string
	weight      float64
	scamType    domain.ScamType
	all         []string
	any         []string
}

var rules = []textRule{
	{
		id:          "upi_pin_to_receive_money",
		description: "Message asks for UPI PIN or QR scan to receive money.",
		weight:      0.68,
		scamType:    domain.ScamUPICollect,
		any:         []string{"receive money", "paise receive", "refund lene", "refund receive", "receive karne", "paise lene", "money receive", "collect request"},
		all:         []string{"upi pin"},
	},
	{
		id:          "qr_receive_trap",
		description: "QR code is positioned as a way to receive money.",
		weight:      0.48,
		scamType:    domain.ScamUPICollect,
		any:         []string{"scan qr to receive", "qr scan karo receive", "refund qr", "receive refund by scanning"},
	},
	{
		id:          "otp_request",
		description: "Message asks for OTP, verification code, or login code.",
		weight:      0.48,
		scamType:    domain.ScamImpersonation,
		any:         []string{"share otp", "send otp", "verification code", "login code", "6 digit code", "six digit code"},
	},
	{
		id:          "kyc_update_pressure",
		description: "KYC/PAN/account update pressure with threat of blocking.",
		weight:      0.42,
		scamType:    domain.ScamPhishing,
		any:         []string{"kyc", "pan update", "aadhaar update", "account blocked", "bank account suspend", "sim blocked"},
	},
	{
		id:          "fake_fraud_alert",
		description: "Fake fraud alert or safe-account transfer language.",
		weight:      0.68,
		scamType:    domain.ScamImpersonation,
		any:         []string{"fraud alert", "suspicious transaction", "safe account", "move your money", "secure wallet", "rbi officer"},
	},
	{
		id:          "remote_access_app",
		description: "Asks user to install a remote access or screen-sharing app.",
		weight:      0.44,
		scamType:    domain.ScamImpersonation,
		any:         []string{"anydesk", "teamviewer", "quick support", "screenshare", "screen share", "remote access"},
	},
	{
		id:          "job_task_deposit",
		description: "Job/task scam signals with deposit, commission, or recharge request.",
		weight:      0.46,
		scamType:    domain.ScamJob,
		any:         []string{"part time job", "daily task", "telegram task", "commission", "prepaid task", "security deposit", "registration fee"},
	},
	{
		id:          "investment_guarantee",
		description: "Investment scam signals with guaranteed or unusually high returns.",
		weight:      0.52,
		scamType:    domain.ScamInvestment,
		any:         []string{"guaranteed return", "double money", "crypto profit", "stock tip", "vip signal", "fixed profit", "risk free investment"},
	},
	{
		id:          "loan_app_pressure",
		description: "Loan repayment intimidation or pressure language.",
		weight:      0.38,
		scamType:    domain.ScamLoan,
		any:         []string{"loan overdue", "contact list", "legal notice", "harassment", "pay immediately", "morph photo"},
	},
	{
		id:          "urgency_pressure",
		description: "Strong urgency or fear pressure.",
		weight:      0.20,
		scamType:    domain.ScamUnknown,
		any:         []string{"urgent", "within 10 minutes", "immediately", "last chance", "account will be blocked", "aaj hi"},
	},
	{
		id:          "fake_receipt_claim",
		description: "Possible fake payment receipt or screenshot claim.",
		weight:      0.28,
		scamType:    domain.ScamFakeReceipt,
		any:         []string{"payment screenshot", "amount credited", "check screenshot", "fake payment", "transaction successful"},
	},
}

func EvaluateRules(entities domain.ExtractedEntities) []domain.Signal {
	var signals []domain.Signal
	for _, rule := range rules {
		if rule.matches(entities.NormalizedText) {
			signals = append(signals, domain.Signal{
				ID:          rule.id,
				Description: rule.description,
				Weight:      rule.weight,
				ScamType:    rule.scamType,
				Reason:      rule.description,
			})
		}
	}
	for _, finding := range entities.URLs {
		signals = append(signals, evaluateURL(finding)...)
	}
	if entities.QR != nil && entities.QR.PayeeUPI != "" {
		if strings.Contains(entities.NormalizedText, "receive") || strings.Contains(entities.NormalizedText, "refund") {
			signals = append(signals, domain.Signal{
				ID:          "qr_has_upi_payee_with_receive_context",
				Description: "QR contains a payable UPI ID while the message talks about receiving/refund.",
				Weight:      0.48,
				ScamType:    domain.ScamUPICollect,
				Reason:      "UPI QR/PIN flow is for paying money, not receiving money.",
			})
		}
	}
	return signals
}

func (r textRule) matches(text string) bool {
	for _, token := range r.all {
		if !strings.Contains(text, token) {
			return false
		}
	}
	if len(r.any) == 0 {
		return true
	}
	for _, token := range r.any {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func evaluateURL(finding domain.URLFinding) []domain.Signal {
	var signals []domain.Signal
	if finding.IsShortener {
		signals = append(signals, domain.Signal{
			ID:          "shortened_url",
			Description: "Shortened URL hides the final destination.",
			Weight:      0.20,
			ScamType:    domain.ScamPhishing,
			Reason:      "Short links are commonly used to hide phishing pages.",
		})
	}
	host := finding.Host
	brandTerms := []string{"sbi", "hdfc", "icici", "axisbank", "paytm", "phonepe", "gpay", "googlepay", "rbi", "amazon"}
	officialDomains := []string{
		"sbi.co.in", "hdfcbank.com", "icicibank.com", "axisbank.com", "paytm.com",
		"phonepe.com", "google.com", "rbi.org.in", "amazon.in", "amazon.com",
	}
	for _, brand := range brandTerms {
		if strings.Contains(host, brand) && !isOfficialDomain(host, officialDomains) {
			signals = append(signals, domain.Signal{
				ID:          "brand_spoofing",
				Description: "URL appears to imitate a trusted brand.",
				Weight:      0.42,
				ScamType:    domain.ScamPhishing,
				Reason:      "The domain contains a trusted brand name but is not an official domain.",
			})
			break
		}
	}
	if strings.Contains(host, "verify") || strings.Contains(host, "kyc") || strings.Contains(host, "support") {
		signals = append(signals, domain.Signal{
			ID:          "suspicious_url_keyword",
			Description: "URL uses verification/support keywords often seen in phishing.",
			Weight:      0.18,
			ScamType:    domain.ScamPhishing,
			Reason:      "Scam links often use urgent verification wording.",
		})
	}
	return signals
}

func isOfficialDomain(host string, officialDomains []string) bool {
	for _, domainName := range officialDomains {
		if host == domainName || strings.HasSuffix(host, "."+domainName) {
			return true
		}
	}
	return false
}
