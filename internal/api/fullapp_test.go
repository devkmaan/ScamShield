package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scamshield/internal/domain"
)

func TestAdminSummaryAndSimulation(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	req := httptest.NewRequest(http.MethodPost, "/v1/simulate/stream", bytes.NewBufferString(`{"count":4}`))
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	summaryReq := httptest.NewRequest(http.MethodGet, "/v1/admin/summary", nil)
	summaryRes := httptest.NewRecorder()
	server.Router().ServeHTTP(summaryRes, summaryReq)
	if summaryRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", summaryRes.Code, summaryRes.Body.String())
	}
	var summary domain.AdminSummary
	if err := json.Unmarshal(summaryRes.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.DecisionCount < 4 {
		t.Fatalf("expected at least 4 decisions, got %d", summary.DecisionCount)
	}
	if summary.HighRiskCount == 0 {
		t.Fatal("expected high-risk simulated decisions")
	}
}

func TestEvidenceCreateListDelete(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	req := httptest.NewRequest(http.MethodPost, "/v1/evidence", bytes.NewBufferString(`{
		"userId":"u1",
		"mediaType":"TEXT",
		"source":"TEST",
		"content":"OTP is 123456 and card 4111 1111 1111 1111"
	}`))
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}
	var evidence domain.EvidenceObject
	if err := json.Unmarshal(res.Body.Bytes(), &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.EvidenceID == "" || evidence.SHA256 == "" {
		t.Fatalf("expected stored evidence, got %+v", evidence)
	}
	if bytes.Contains(res.Body.Bytes(), []byte("123456")) || bytes.Contains(res.Body.Bytes(), []byte("4111")) {
		t.Fatal("sensitive evidence content leaked in response")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/evidence", nil)
	listRes := httptest.NewRecorder()
	server.Router().ServeHTTP(listRes, listReq)
	if !bytes.Contains(listRes.Body.Bytes(), []byte(evidence.EvidenceID)) {
		t.Fatalf("expected evidence in list, got %s", listRes.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/v1/evidence/"+evidence.EvidenceID, nil)
	delRes := httptest.NewRecorder()
	server.Router().ServeHTTP(delRes, delReq)
	if delRes.Code != http.StatusOK {
		t.Fatalf("expected delete 200, got %d", delRes.Code)
	}
}

func TestModelAndPayeeInternalEndpoints(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	modelReq := httptest.NewRequest(http.MethodPost, "/internal/model/score-text", bytes.NewBufferString(`{"text":"Guaranteed return crypto profit double money"}`))
	modelRes := httptest.NewRecorder()
	server.Router().ServeHTTP(modelRes, modelReq)
	if modelRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", modelRes.Code, modelRes.Body.String())
	}
	if !bytes.Contains(modelRes.Body.Bytes(), []byte("modelVersion")) {
		t.Fatalf("expected model response, got %s", modelRes.Body.String())
	}

	payeeReq := httptest.NewRequest(http.MethodPost, "/internal/payee/report", bytes.NewBufferString(`{"rawPayee":"risk-payee@upi","alias":"Risk Payee"}`))
	payeeRes := httptest.NewRecorder()
	server.Router().ServeHTTP(payeeRes, payeeReq)
	if payeeRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", payeeRes.Code, payeeRes.Body.String())
	}
	if bytes.Contains(payeeRes.Body.Bytes(), []byte("risk-payee@upi")) {
		t.Fatal("raw payee leaked in response")
	}
}

func TestI18nBundleAndGenAIMetadataFallback(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/v1/i18n/bundle?language=hi", nil)
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte("nav.dashboard")) {
		t.Fatalf("expected UI bundle keys, got %s", res.Body.String())
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/internal/genai/metadata", nil)
	metaRes := httptest.NewRecorder()
	server.Router().ServeHTTP(metaRes, metaReq)
	if metaRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", metaRes.Code, metaRes.Body.String())
	}
	if !bytes.Contains(metaRes.Body.Bytes(), []byte("go-fallback")) {
		t.Fatalf("expected fallback metadata, got %s", metaRes.Body.String())
	}
}

func TestDecisionDetailHistoryShareAndInsights(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	body := bytes.NewBufferString(`{"inputType":"TEXT","userId":"history-user","text":"Share OTP for KYC update and pay kyc-helpdesk@upi"}`)
	checkReq := httptest.NewRequest(http.MethodPost, "/v1/check", body)
	checkRes := httptest.NewRecorder()
	server.Router().ServeHTTP(checkRes, checkReq)
	if checkRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", checkRes.Code, checkRes.Body.String())
	}
	var decision domain.RiskDecision
	if err := json.Unmarshal(checkRes.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.UserID != "history-user" {
		t.Fatalf("expected user id on decision, got %#v", decision)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/v1/decisions/"+decision.DecisionID, nil)
	detailRes := httptest.NewRecorder()
	server.Router().ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected detail 200, got %d: %s", detailRes.Code, detailRes.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/v1/users/history-user/history?limit=10", nil)
	historyRes := httptest.NewRecorder()
	server.Router().ServeHTTP(historyRes, historyReq)
	if historyRes.Code != http.StatusOK {
		t.Fatalf("expected history 200, got %d: %s", historyRes.Code, historyRes.Body.String())
	}
	if !bytes.Contains(historyRes.Body.Bytes(), []byte(decision.DecisionID)) {
		t.Fatalf("expected decision in history, got %s", historyRes.Body.String())
	}

	shareReq := httptest.NewRequest(http.MethodGet, "/v1/decisions/"+decision.DecisionID+"/share", nil)
	shareRes := httptest.NewRecorder()
	server.Router().ServeHTTP(shareRes, shareReq)
	if shareRes.Code != http.StatusOK {
		t.Fatalf("expected share 200, got %d: %s", shareRes.Code, shareRes.Body.String())
	}
	if bytes.Contains(shareRes.Body.Bytes(), []byte("kyc-helpdesk@upi")) {
		t.Fatalf("raw UPI leaked in share response: %s", shareRes.Body.String())
	}
	if !bytes.Contains(shareRes.Body.Bytes(), []byte("1930")) || !bytes.Contains(shareRes.Body.Bytes(), []byte("cybercrime.gov.in")) {
		t.Fatalf("expected official reporting guidance, got %s", shareRes.Body.String())
	}

	insightsReq := httptest.NewRequest(http.MethodGet, "/v1/insights/trends", nil)
	insightsRes := httptest.NewRecorder()
	server.Router().ServeHTTP(insightsRes, insightsReq)
	if insightsRes.Code != http.StatusOK {
		t.Fatalf("expected insights 200, got %d: %s", insightsRes.Code, insightsRes.Body.String())
	}
	if !bytes.Contains(insightsRes.Body.Bytes(), []byte("riskLevels")) {
		t.Fatalf("expected insight buckets, got %s", insightsRes.Body.String())
	}
}

func TestAdminDecisionsFilteringAndPagination(t *testing.T) {
	server, cancel := newTestServer()
	defer cancel()

	req := httptest.NewRequest(http.MethodPost, "/v1/simulate/stream", bytes.NewBufferString(`{"count":12}`))
	res := httptest.NewRecorder()
	server.Router().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected simulation 200, got %d: %s", res.Code, res.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/admin/decisions?page=1&pageSize=5", nil)
	listRes := httptest.NewRecorder()
	server.Router().ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d: %s", listRes.Code, listRes.Body.String())
	}
	var page domain.DecisionListResponse
	if err := json.Unmarshal(listRes.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Page != 1 || page.PageSize != 5 || page.Total < 12 || len(page.Items) != 5 || !page.HasNext {
		t.Fatalf("unexpected pagination response: %+v", page)
	}

	filterReq := httptest.NewRequest(http.MethodGet, "/v1/admin/decisions?scamType=PHISHING&pageSize=50", nil)
	filterRes := httptest.NewRecorder()
	server.Router().ServeHTTP(filterRes, filterReq)
	if filterRes.Code != http.StatusOK {
		t.Fatalf("expected filter 200, got %d: %s", filterRes.Code, filterRes.Body.String())
	}
	var filtered domain.DecisionListResponse
	if err := json.Unmarshal(filterRes.Body.Bytes(), &filtered); err != nil {
		t.Fatal(err)
	}
	if filtered.Total == 0 {
		t.Fatal("expected phishing decisions from simulation")
	}
	for _, decision := range filtered.Items {
		if decision.ScamType != domain.ScamPhishing {
			t.Fatalf("expected only phishing decisions, got %#v", decision.ScamType)
		}
	}

	combinedReq := httptest.NewRequest(http.MethodGet, "/v1/admin/decisions?riskLevel=HIGH_RISK&inputType=TEXT&pageSize=50", nil)
	combinedRes := httptest.NewRecorder()
	server.Router().ServeHTTP(combinedRes, combinedReq)
	if combinedRes.Code != http.StatusOK {
		t.Fatalf("expected combined filter 200, got %d: %s", combinedRes.Code, combinedRes.Body.String())
	}
	if bytes.Contains(combinedRes.Body.Bytes(), []byte("refund-care@okaxis")) {
		t.Fatalf("raw UPI leaked in analytics response: %s", combinedRes.Body.String())
	}
}
