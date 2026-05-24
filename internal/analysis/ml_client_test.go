package analysis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scamshield/internal/domain"
	"scamshield/internal/store"
)

func TestOrchestratorUsesExternalMLModelVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/model/score-text" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(domain.ModelScoreResponse{
			ModelVersion: "text-scam-test-v1",
			Score:        0.82,
			Confidence:   0.88,
			ScamTypeScores: map[domain.ScamType]float64{
				domain.ScamJob: 0.82,
			},
			Signals: []string{"ml_task_deposit"},
		})
	}))
	defer server.Close()

	repo := store.NewMemoryStore("test-salt")
	orchestrator := NewOrchestrator(repo, NewHTTPMLClient(server.URL))
	decision := orchestrator.Analyze(domain.CheckRequest{
		InputType: domain.InputText,
		Text:      "Daily task commission deposit required",
	})

	if decision.ModelVersions["text"] != "text-scam-test-v1" {
		t.Fatalf("expected external model version, got %#v", decision.ModelVersions)
	}
	if decision.RiskLevel != domain.RiskHigh {
		t.Fatalf("expected high risk from external ML, got %s score %.2f", decision.RiskLevel, decision.Score)
	}
}

func TestMLClientMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode": "test",
			"activeModels": map[string]string{
				"text": "text-scam-test-v1",
			},
		})
	}))
	defer server.Close()

	client := NewHTTPMLClient(server.URL)
	payload, err := client.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if payload["mode"] != "test" {
		t.Fatalf("unexpected metadata: %#v", payload)
	}
}
