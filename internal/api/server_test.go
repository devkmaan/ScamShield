package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scamshield/internal/analysis"
	"scamshield/internal/domain"
	"scamshield/internal/store"
)

func newTestServer() (*Server, context.CancelFunc) {
	repo := store.NewMemoryStore("test-salt")
	repo.SeedDemoMerchants()
	server := NewServer(Config{VerifyToken: "verify"}, analysis.NewOrchestrator(repo), repo)
	ctx, cancel := context.WithCancel(context.Background())
	go server.StartWorkers(ctx, 1)
	return server, cancel
}

func TestCheckEndpoint(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	body := bytes.NewBufferString(`{"inputType":"TEXT","text":"RBI officer bol raha hu, suspicious transaction hai, move your money to safe account"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/check", body)
	res := httptest.NewRecorder()

	server.Router().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var decision domain.RiskDecision
	if err := json.Unmarshal(res.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.RiskLevel != domain.RiskHigh && decision.RiskLevel != domain.RiskCritical {
		t.Fatalf("expected high/critical risk, got %s", decision.RiskLevel)
	}
}

func TestWhatsAppWebhookQueuesReply(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	payload := `{
	  "entry": [{
	    "changes": [{
	      "value": {
	        "messages": [{
	          "id": "wamid.demo",
	          "from": "919999999999",
	          "type": "text",
	          "text": { "body": "KYC update urgent. Click https://paytm-verify-help.com and share OTP" }
	        }]
	      }
	    }]
	  }]
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", bytes.NewBufferString(payload))
	res := httptest.NewRecorder()

	server.Router().ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", res.Code, res.Body.String())
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		outboxReq := httptest.NewRequest(http.MethodGet, "/v1/outbox?userId=919999999999", nil)
		outboxRes := httptest.NewRecorder()
		server.Router().ServeHTTP(outboxRes, outboxReq)
		if bytes.Contains(outboxRes.Body.Bytes(), []byte("HIGH_RISK")) || bytes.Contains(outboxRes.Body.Bytes(), []byte("CRITICAL")) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected high-risk reply in outbox")
}

func TestWhatsAppWebhookDedupesMessageID(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	payload := `{
	  "entry": [{
	    "changes": [{
	      "value": {
	        "messages": [{
	          "id": "wamid.duplicate",
	          "from": "919999999999",
	          "type": "text",
	          "text": { "body": "Share OTP for KYC update" }
	        }]
	      }
	    }]
	  }]
	}`
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", bytes.NewBufferString(payload))
		res := httptest.NewRecorder()
		server.Router().ServeHTTP(res, req)
		if res.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d: %s", res.Code, res.Body.String())
		}
	}

	time.Sleep(100 * time.Millisecond)
	outboxReq := httptest.NewRequest(http.MethodGet, "/v1/outbox?userId=919999999999", nil)
	outboxRes := httptest.NewRecorder()
	server.Router().ServeHTTP(outboxRes, outboxReq)
	var payloadOut struct {
		Items []domain.WhatsAppReply `json:"items"`
	}
	if err := json.Unmarshal(outboxRes.Body.Bytes(), &payloadOut); err != nil {
		t.Fatal(err)
	}
	if len(payloadOut.Items) != 1 {
		t.Fatalf("expected one reply after duplicate webhook, got %d", len(payloadOut.Items))
	}
}

func TestWhatsAppLanguageCommandStoresPreference(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	payload := `{
	  "entry": [{
	    "changes": [{
	      "value": {
	        "messages": [{
	          "id": "wamid.language",
	          "from": "919999999998",
	          "type": "text",
	          "text": { "body": "/language hi" }
	        }]
	      }
	    }]
	  }]
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", bytes.NewBufferString(payload))
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", res.Code, res.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if server.repo.GetUserLanguage("919999999998") == "hi" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected Hindi preference to be stored")
}

func TestEventsEndpointIncludesRiskDecision(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	body := bytes.NewBufferString(`{"inputType":"TEXT","userId":"u1","text":"Share OTP for KYC update"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/check", body)
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	eventsRes := httptest.NewRecorder()
	server.Router().ServeHTTP(eventsRes, eventsReq)
	if !bytes.Contains(eventsRes.Body.Bytes(), []byte("risk.decision.created")) {
		t.Fatalf("expected risk decision event, got %s", eventsRes.Body.String())
	}
}

func TestFeedbackUpdatesPayeeRisk(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	body := bytes.NewBufferString(`{"verdict":"SCAM","payeeUpi":"newfraud@upi","comment":"asked for advance"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", body)
	res := httptest.NewRecorder()

	server.Router().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var feedback domain.FeedbackResponse
	if err := json.Unmarshal(res.Body.Bytes(), &feedback); err != nil {
		t.Fatal(err)
	}
	if feedback.PayeeHash == "" {
		t.Fatal("expected payee hash")
	}

	riskReq := httptest.NewRequest(http.MethodGet, "/v1/risk/payee/"+feedback.PayeeHash, nil)
	riskRes := httptest.NewRecorder()
	server.Router().ServeHTTP(riskRes, riskReq)
	if riskRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", riskRes.Code)
	}
	if bytes.Contains(riskRes.Body.Bytes(), []byte("newfraud@upi")) {
		t.Fatal("raw UPI id leaked in risk response")
	}
}
