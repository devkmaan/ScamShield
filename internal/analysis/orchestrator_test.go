package analysis

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scamshield/internal/domain"
	"scamshield/internal/store"
)

func newTestOrchestrator() *Orchestrator {
	repo := store.NewMemoryStore("test-salt")
	repo.SeedDemoMerchants()
	return NewOrchestrator(repo)
}

func TestAnalyzeUPIReceivePINScam(t *testing.T) {
	o := newTestOrchestrator()
	decision := o.Analyze(domain.CheckRequest{
		InputType: domain.InputText,
		Text:      "Refund lene ke liye QR scan karo and UPI PIN enter karo to receive money urgently.",
	})

	if decision.RiskLevel != domain.RiskHigh && decision.RiskLevel != domain.RiskCritical {
		t.Fatalf("expected high or critical risk, got %s score %.2f", decision.RiskLevel, decision.Score)
	}
	if decision.ScamType != domain.ScamUPICollect {
		t.Fatalf("expected UPI collect scam, got %s", decision.ScamType)
	}
	assertSignal(t, decision.TopSignals, "upi_pin_to_receive_money")
}

func TestAnalyzePhishingURL(t *testing.T) {
	o := newTestOrchestrator()
	decision := o.Analyze(domain.CheckRequest{
		InputType: domain.InputText,
		Text:      "Your SBI KYC is blocked. Click https://sbi-verify-support.com/login and share OTP immediately.",
	})

	if decision.RiskLevel != domain.RiskHigh && decision.RiskLevel != domain.RiskCritical {
		t.Fatalf("expected high or critical risk, got %s score %.2f", decision.RiskLevel, decision.Score)
	}
	assertSignal(t, decision.TopSignals, "brand_spoofing")
	assertSignal(t, decision.TopSignals, "otp_request")
}

func TestAnalyzeLegitOfficialDomainDoesNotSpoof(t *testing.T) {
	o := newTestOrchestrator()
	decision := o.Analyze(domain.CheckRequest{
		InputType: domain.InputURL,
		URL:       "https://www.hdfcbank.com/personal/pay/cards",
		Text:      "Please verify card details from official HDFC Bank website.",
	})

	if decision.RiskLevel == domain.RiskHigh || decision.RiskLevel == domain.RiskCritical {
		t.Fatalf("expected non-high risk for official domain, got %s score %.2f", decision.RiskLevel, decision.Score)
	}
	for _, signal := range decision.TopSignals {
		if signal == "brand_spoofing" {
			t.Fatal("official domain should not trigger brand_spoofing")
		}
	}
}

func TestAlreadyPaidCreatesRecoveryReport(t *testing.T) {
	o := newTestOrchestrator()
	decision := o.Analyze(domain.CheckRequest{
		UserID:      "user-1",
		InputType:   domain.InputText,
		Text:        "I already paid 5000 to taskbonus@paytm after telegram task scam.",
		AlreadyPaid: true,
	})

	if decision.ReportID == "" {
		t.Fatal("expected report id")
	}
}

func TestAnalyzeUsesGenAINormalizationAndRender(t *testing.T) {
	genAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/genai/normalize-input":
			_ = json.NewEncoder(w).Encode(domain.GenAINormalizeResponse{
				DetectedLanguage: "hi",
				NormalizedText:   "Your SBI KYC is blocked. Share OTP immediately.",
				InputSummary:     "KYC OTP phishing",
				ModelVersion:     "genai-normalizer-test-v1",
			})
		case "/internal/genai/render":
			_ = json.NewEncoder(w).Encode(domain.GenAIRenderResponse{
				Language:           "hi",
				UserMessage:        "High risk KYC scam signal mila hai. Payment ya OTP share mat karo.",
				RecommendedActions: []string{"OTP, UPI PIN, card details ya remote access share mat karo."},
				ModelVersion:       "genai-render-test-v1",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer genAIServer.Close()

	repo := store.NewMemoryStore("test-salt")
	orchestrator := NewOrchestrator(repo)
	orchestrator.SetGenAIClient(NewHTTPGenAIClient(genAIServer.URL))

	decision := orchestrator.Analyze(domain.CheckRequest{
		InputType: domain.InputText,
		Language:  "hi",
		Text:      "आपका SBI KYC बंद है। OTP तुरंत शेयर करें।",
	})

	if decision.Language != "hi" || decision.InputLanguage != "hi" {
		t.Fatalf("expected Hindi language metadata, got language=%q input=%q", decision.Language, decision.InputLanguage)
	}
	if decision.ModelVersions["normalizer"] != "genai-normalizer-test-v1" {
		t.Fatalf("expected normalizer version, got %#v", decision.ModelVersions)
	}
	if decision.ModelVersions["genai"] != "genai-render-test-v1" {
		t.Fatalf("expected genai version, got %#v", decision.ModelVersions)
	}
	if decision.UserMessage == "" || decision.RiskLevel == domain.RiskLow {
		t.Fatalf("expected rendered risky decision, got %+v", decision)
	}
}

func assertSignal(t *testing.T, signals []string, expected string) {
	t.Helper()
	for _, signal := range signals {
		if signal == expected {
			return
		}
	}
	t.Fatalf("expected signal %q in %v", expected, signals)
}
