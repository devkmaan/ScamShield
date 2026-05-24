package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"scamshield/internal/domain"
)

func (s *Server) handleGenAILanguages(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": domain.SupportedLanguages()})
}

func (s *Server) handleGenAIMetadata(w http.ResponseWriter, r *http.Request) {
	if s.genAIClient != nil {
		if response, err := s.genAIClient.Metadata(r.Context()); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"activeModels": map[string]string{
			"renderer":   "local-safe-fallback-v1",
			"normalizer": "local-safe-fallback-v1",
		},
		"mode":      "go-fallback",
		"policy":    "GenAI is disabled or unavailable. Go still owns risk decisions.",
		"updatedAt": time.Now().UTC(),
	})
}

func (s *Server) handleGenAINormalize(w http.ResponseWriter, r *http.Request) {
	var req domain.GenAINormalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if s.genAIClient != nil {
		if response, err := s.genAIClient.NormalizeInput(r.Context(), req); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	writeJSON(w, http.StatusOK, domain.GenAINormalizeResponse{
		DetectedLanguage: domain.NormalizeLanguage(req.UserSelectedLanguage),
		NormalizedText:   req.Text,
		InputSummary:     "GenAI normalizer unavailable; original text was used.",
		ModelVersion:     "go-normalizer-fallback-v1",
		FallbackUsed:     true,
	})
}

func (s *Server) handleGenAIRender(w http.ResponseWriter, r *http.Request) {
	var req domain.GenAIRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if s.genAIClient != nil {
		if response, err := s.genAIClient.Render(r.Context(), req); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	writeJSON(w, http.StatusOK, fallbackRenderResponse(req))
}

func (s *Server) handleGenAIChat(w http.ResponseWriter, r *http.Request) {
	var req domain.GenAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if s.genAIClient != nil {
		if response, err := s.genAIClient.Chat(r.Context(), req); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	writeJSON(w, http.StatusOK, domain.GenAIChatResponse{
		Language:         domain.NormalizeLanguage(req.Language),
		Reply:            "Send the suspicious message, link, UPI ID, QR, or screenshot context. If money is already lost, call 1930 and file at cybercrime.gov.in.",
		SuggestedActions: []string{"Check message", "Check link", "Need Help"},
		ModelVersion:     "go-chat-fallback-v1",
		FallbackUsed:     true,
	})
}

func (s *Server) handleGenAIUIBundle(w http.ResponseWriter, r *http.Request) {
	var req domain.UIBundleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response := s.uiBundle(r, req.Language, req.Keys)
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleI18nBundle(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "en"
	}
	writeJSON(w, http.StatusOK, s.uiBundle(r, language, defaultUIBundle()))
}

func (s *Server) uiBundle(r *http.Request, language string, keys map[string]string) domain.UIBundleResponse {
	language = domain.NormalizeLanguage(language)
	if len(keys) == 0 {
		keys = defaultUIBundle()
	}
	cacheKey := language
	if len(keys) > 0 {
		cacheKey = language + ":" + strconv.Itoa(len(keys))
	}
	s.i18nMu.RLock()
	if cached, ok := s.i18nCache[cacheKey]; ok {
		s.i18nMu.RUnlock()
		return cached
	}
	s.i18nMu.RUnlock()

	response := domain.UIBundleResponse{
		Language:     language,
		Bundle:       keys,
		ModelVersion: "go-ui-fallback-v1",
		FallbackUsed: true,
	}
	if language == "en" {
		response.FallbackUsed = false
	} else if s.genAIClient != nil {
		if generated, err := s.genAIClient.UIBundle(r.Context(), domain.UIBundleRequest{Language: language, Keys: keys}); err == nil && len(generated.Bundle) > 0 {
			response = generated
		}
	}
	s.i18nMu.Lock()
	s.i18nCache[cacheKey] = response
	s.i18nMu.Unlock()
	return response
}

func fallbackRenderResponse(req domain.GenAIRenderRequest) domain.GenAIRenderResponse {
	level := req.Decision.RiskLevel
	if level == "" {
		level = domain.RiskCaution
	}
	return domain.GenAIRenderResponse{
		Language:    domain.NormalizeLanguage(req.Language),
		UserMessage: "Risk level is " + string(level) + ". Pause and verify through official channels before acting.",
		RecommendedActions: []string{
			"Do not share OTP, UPI PIN, card details, screen, or remote access.",
			"If money is already lost, contact your bank, call 1930, and file at cybercrime.gov.in.",
		},
		Summary: "Possible cyber financial fraud. This draft helps collect evidence and act quickly; it is not an official complaint.",
		Checklist: []string{
			"Contact your bank or payment app support immediately.",
			"Call 1930 quickly if money moved recently.",
			"File a complaint at cybercrime.gov.in and keep the acknowledgement number.",
			"Save chats, phone numbers, UPI IDs, transaction IDs, URLs, and receipts.",
		},
		OfficialHelp: []string{
			"National Cybercrime Helpline: 1930",
			"National Cyber Crime Reporting Portal: https://cybercrime.gov.in",
			"Your bank/payment app's official fraud support channel",
		},
		ModelVersion: "go-render-fallback-v1",
		FallbackUsed: true,
	}
}

func defaultUIBundle() map[string]string {
	return map[string]string{
		"app.eyebrow":              "WhatsApp-first consumer protection",
		"app.title":                "Real-time fraud and scam detection",
		"app.ready":                "ready",
		"app.checking":             "checking",
		"app.backendUnavailable":   "backend unavailable",
		"nav.dashboard":            "Dashboard",
		"nav.check":                "Risk Check",
		"nav.whatsapp":             "WhatsApp",
		"nav.recovery":             "Recovery",
		"nav.merchant":             "Merchant Risk",
		"nav.model":                "Model Lab",
		"nav.events":               "Events",
		"dashboard.eyebrow":        "Operations",
		"dashboard.title":          "Risk overview",
		"dashboard.generate":       "Generate Demo Data",
		"dashboard.generating":     "Generating",
		"metric.decisions":         "Decisions",
		"metric.highRisk":          "High Risk",
		"metric.reviewQueue":       "Review Queue",
		"metric.evidence":          "Evidence",
		"panel.recentDecisions":    "Recent Decisions",
		"panel.topRiskPayees":      "Top Risk Payees",
		"panel.output":             "Output",
		"panel.eventLog":           "Event Log",
		"check.title":              "Manual Risk Check",
		"field.inputType":          "Input Type",
		"field.content":            "Content",
		"field.language":           "Language",
		"button.analyze":           "Analyze",
		"button.analyzing":         "Analyzing",
		"button.refresh":           "Refresh",
		"button.send":              "Send",
		"button.setLanguage":       "Set Language",
		"button.outbox":            "Outbox",
		"button.createDraft":       "Create Draft",
		"button.saveEvidence":      "Save Evidence",
		"button.observe":           "Observe",
		"button.report":            "Report",
		"button.score":             "Score",
		"button.metadata":          "Metadata",
		"whatsapp.title":           "WhatsApp Simulator",
		"whatsapp.phone":           "User Phone",
		"whatsapp.message":         "Forwarded Message",
		"whatsapp.setLanguageHint": "Send /language command first if you want this WhatsApp user session to change language.",
		"recovery.title":           "Recovery Case",
		"recovery.userId":          "User ID",
		"recovery.lossContext":     "Loss Context",
		"merchant.title":           "Merchant / Payee Risk",
		"merchant.upiId":           "UPI ID",
		"merchant.alias":           "Alias",
		"model.title":              "Model Lab",
		"model.text":               "Text",
		"empty.noDecision":         "No decision yet.",
		"empty.noDecisions":        "No decisions yet.",
		"empty.noPayees":           "No payees yet.",
		"status.waiting":           "waiting",
		"decision.score":           "Score",
		"decision.actions":         "Actions",
		"table.risk":               "Risk",
		"table.score":              "Score",
		"table.type":               "Type",
		"table.model":              "Model",
		"table.signals":            "Signals",
		"table.complaints":         "Complaints",
		"table.payeeHash":          "Payee Hash",
	}
}
