package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"scamshield/internal/analysis"
	"scamshield/internal/domain"
	"scamshield/internal/privacy"
	"scamshield/internal/store"
)

type Config struct {
	VerifyToken        string
	RateLimitPerMinute int
	MLServiceURL       string
	GenAIServiceURL    string
}

type Server struct {
	config       Config
	orchestrator *analysis.Orchestrator
	repo         *store.MemoryStore
	mlClient     *analysis.MLClient
	genAIClient  *analysis.GenAIClient
	events       chan domain.WhatsAppInbound
	i18nMu       sync.RWMutex
	i18nCache    map[string]domain.UIBundleResponse
}

func NewServer(config Config, orchestrator *analysis.Orchestrator, repo *store.MemoryStore) *Server {
	genAIClient := analysis.NewHTTPGenAIClient(config.GenAIServiceURL)
	if orchestrator != nil {
		orchestrator.SetGenAIClient(genAIClient)
	}
	return &Server{
		config:       config,
		orchestrator: orchestrator,
		repo:         repo,
		mlClient:     analysis.NewHTTPMLClient(config.MLServiceURL),
		genAIClient:  genAIClient,
		events:       make(chan domain.WhatsAppInbound, 128),
		i18nCache:    map[string]domain.UIBundleResponse{},
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleAdmin)
	mux.HandleFunc("GET /app", s.handleAdmin)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)
	mux.HandleFunc("POST /v1/check", s.handleCheck)
	mux.HandleFunc("POST /v1/feedback", s.handleFeedback)
	mux.HandleFunc("POST /v1/recovery/start", s.handleRecoveryStart)
	mux.HandleFunc("POST /v1/evidence", s.handleCreateEvidence)
	mux.HandleFunc("GET /v1/evidence", s.handleListEvidence)
	mux.HandleFunc("DELETE /v1/evidence/{evidenceID}", s.handleDeleteEvidence)
	mux.HandleFunc("GET /v1/risk/payee/{payeeHash}", s.handlePayeeRisk)
	mux.HandleFunc("GET /v1/reports/{reportID}", s.handleReport)
	mux.HandleFunc("GET /v1/decisions/{decisionID}", s.handleDecisionDetail)
	mux.HandleFunc("GET /v1/decisions/{decisionID}/share", s.handleDecisionShare)
	mux.HandleFunc("GET /v1/users/{userID}/history", s.handleUserHistory)
	mux.HandleFunc("GET /v1/insights/trends", s.handleInsightsTrends)
	mux.HandleFunc("GET /v1/outbox", s.handleOutbox)
	mux.HandleFunc("GET /v1/events", s.handleEvents)
	mux.HandleFunc("GET /v1/i18n/bundle", s.handleI18nBundle)
	mux.HandleFunc("POST /v1/i18n/bundle", s.handleGenAIUIBundle)
	mux.HandleFunc("GET /v1/admin/summary", s.handleAdminSummary)
	mux.HandleFunc("GET /v1/admin/decisions", s.handleAdminDecisions)
	mux.HandleFunc("GET /v1/admin/merchants", s.handleAdminMerchants)
	mux.HandleFunc("GET /v1/admin/feedback", s.handleAdminFeedback)
	mux.HandleFunc("GET /v1/admin/reports", s.handleAdminReports)
	mux.HandleFunc("POST /v1/simulate/stream", s.handleSimulation)
	mux.HandleFunc("POST /internal/model/score-text", s.handleModelScoreText)
	mux.HandleFunc("POST /internal/model/score-url", s.handleModelScoreURL)
	mux.HandleFunc("GET /internal/model/metadata", s.handleModelMetadata)
	mux.HandleFunc("POST /internal/genai/normalize-input", s.handleGenAINormalize)
	mux.HandleFunc("POST /internal/genai/render", s.handleGenAIRender)
	mux.HandleFunc("POST /internal/genai/chat", s.handleGenAIChat)
	mux.HandleFunc("POST /internal/genai/ui-bundle", s.handleGenAIUIBundle)
	mux.HandleFunc("GET /internal/genai/languages", s.handleGenAILanguages)
	mux.HandleFunc("GET /internal/genai/metadata", s.handleGenAIMetadata)
	mux.HandleFunc("POST /internal/payee/observe", s.handlePayeeObserve)
	mux.HandleFunc("POST /internal/payee/report", s.handlePayeeReport)
	mux.HandleFunc("GET /webhooks/whatsapp", s.handleWhatsAppVerify)
	mux.HandleFunc("POST /webhooks/whatsapp", s.handleWhatsAppWebhook)
	mux.HandleFunc("GET /admin", s.handleAdmin)
	mux.HandleFunc("GET /admin/", s.handleAdmin)
	return withJSONErrors(mux)
}

func (s *Server) StartWorkers(ctx context.Context, count int) {
	for i := 0; i < count; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case inbound := <-s.events:
					s.processInbound(inbound)
				}
			}
		}()
	}
	<-ctx.Done()
}

func (s *Server) processInbound(inbound domain.WhatsAppInbound) {
	correlationID := inbound.CorrelationID
	if correlationID == "" {
		correlationID = newEventID("corr")
	}
	body := inbound.Body
	if inbound.Caption != "" {
		body += " " + inbound.Caption
	}
	if code, ok := parseLanguageCommand(body); ok {
		s.repo.SetUserLanguage(inbound.From, code)
		reply := domain.WhatsAppReply{
			To:        inbound.From,
			MessageID: inbound.MessageID,
			Text:      languageUpdatedMessage(code),
			Buttons:   []string{"Scam", "Not Scam", "Need Help"},
			CreatedAt: time.Now().UTC(),
		}
		s.repo.SaveReply(reply)
		s.repo.AppendEvent(newEvent(domain.EventWhatsAppReplyRequested, correlationID, inbound.MessageID, reply))
		return
	}
	if s.repo.GetUserLanguage(inbound.From) == "" && isLanguageGreeting(body) {
		reply := domain.WhatsAppReply{
			To:        inbound.From,
			MessageID: inbound.MessageID,
			Text:      languageSelectionMessage(),
			Buttons:   []string{"English", "Hindi", "Hinglish"},
			CreatedAt: time.Now().UTC(),
		}
		s.repo.SaveReply(reply)
		s.repo.AppendEvent(newEvent(domain.EventWhatsAppReplyRequested, correlationID, inbound.MessageID, reply))
		return
	}
	language := s.repo.GetUserLanguage(inbound.From)
	if language == "" {
		language = "hinglish"
		s.repo.SetUserLanguage(inbound.From, language)
	}
	req := domain.CheckRequest{
		UserID:    inbound.From,
		InputType: domain.InputText,
		Text:      body,
		MediaRef:  inbound.MediaID,
		Language:  language,
	}
	if inbound.Type == "image" || inbound.Type == "document" {
		req.InputType = domain.InputScreenshot
	}
	decision := s.orchestrator.Analyze(req)
	s.repo.AppendEvent(newEvent(domain.EventRiskDecisionCreated, correlationID, inbound.MessageID, decision))
	reply := domain.WhatsAppReply{
		To:        inbound.From,
		MessageID: inbound.MessageID,
		Text:      formatWhatsAppReply(decision),
		Buttons:   []string{"Scam", "Not Scam", "Need Help"},
		CreatedAt: time.Now().UTC(),
	}
	s.repo.SaveReply(reply)
	s.repo.AppendEvent(newEvent(domain.EventWhatsAppReplyRequested, correlationID, decision.DecisionID, reply))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "scamshield",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ready",
		"service":       "scamshield",
		"queuedEvents":  len(s.events),
		"eventCount":    s.repo.EventCount(),
		"decisionCount": s.repo.DecisionCount(),
		"timestamp":     time.Now().UTC(),
	})
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	var req domain.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !s.allow(req.UserID) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}
	req.Text = privacy.RedactSensitive(req.Text)
	decision := s.orchestrator.Analyze(req)
	s.repo.AppendEvent(newEvent(domain.EventRiskDecisionCreated, newEventID("corr"), "", decision))
	writeJSON(w, http.StatusOK, decision)
}

func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	var req domain.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Verdict == "" {
		writeError(w, http.StatusBadRequest, "verdict is required")
		return
	}
	req.Comment = privacy.RedactSensitive(req.Comment)
	s.repo.SaveFeedback(req)
	s.repo.AppendEvent(newEvent(domain.EventFeedbackReceived, newEventID("corr"), req.DecisionID, req))

	payeeHash := ""
	if strings.EqualFold(req.Verdict, "SCAM") && req.PayeeUPI != "" {
		risk := s.repo.AddMerchantComplaint(req.PayeeUPI, "")
		payeeHash = risk.PayeeHash
		s.repo.AppendEvent(newEvent(domain.EventMerchantRiskUpdated, newEventID("corr"), req.DecisionID, risk))
	}
	writeJSON(w, http.StatusOK, domain.FeedbackResponse{
		Status:    "accepted",
		PayeeHash: payeeHash,
		Message:   "Feedback recorded. Merchant reports are queued for review and do not auto-block by themselves.",
	})
}

func (s *Server) handlePayeeRisk(w http.ResponseWriter, r *http.Request) {
	payeeHash := r.PathValue("payeeHash")
	risk, ok := s.repo.GetMerchantRisk(payeeHash)
	if !ok {
		writeJSON(w, http.StatusOK, domain.MerchantRisk{
			PayeeHash: payeeHash,
			RiskScore: 0.12,
			Aliases:   []string{},
		})
		return
	}
	writeJSON(w, http.StatusOK, risk)
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	reportID := r.PathValue("reportID")
	report, ok := s.repo.GetReport(reportID)
	if !ok {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleOutbox(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.repo.OutboxFor(userID),
	})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.repo.RecentEvents(50),
	})
}

func (s *Server) handleWhatsAppVerify(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")
	if mode == "subscribe" && token == s.config.VerifyToken {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(challenge))
		return
	}
	writeError(w, http.StatusForbidden, "invalid verify token")
}

func (s *Server) handleWhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WhatsAppWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid WhatsApp webhook payload")
		return
	}
	inbound := NormalizeWhatsAppPayload(payload)
	accepted := 0
	duplicates := 0
	rateLimited := 0
	for _, message := range inbound {
		if s.repo.MarkMessageSeen(message.MessageID, 24*time.Hour) {
			duplicates++
			continue
		}
		if !s.allow(message.From) {
			rateLimited++
			continue
		}
		message.Body = privacy.RedactSensitive(message.Body)
		message.Caption = privacy.RedactSensitive(message.Caption)
		message.CorrelationID = newEventID("corr")
		s.repo.AppendEvent(newEvent(domain.EventWhatsAppInbound, message.CorrelationID, "", message))
		select {
		case s.events <- message:
			accepted++
		default:
			writeError(w, http.StatusServiceUnavailable, "event queue is full")
			return
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":      "accepted",
		"processed":   accepted,
		"duplicates":  duplicates,
		"rateLimited": rateLimited,
	})
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(adminHTML))
}

func formatWhatsAppReply(decision domain.RiskDecision) string {
	actions := strings.Join(decision.RecommendedActions, "\n- ")
	if actions != "" {
		actions = "\n\nActions:\n- " + actions
	}
	report := ""
	if decision.ReportID != "" {
		report = "\n\nRecovery checklist ready. Report ID: " + decision.ReportID
	}
	return "Verdict: " + string(decision.RiskLevel) +
		"\nScore: " + formatFloat(decision.Score) +
		"\nType: " + string(decision.ScamType) +
		"\n\n" + decision.UserMessage + actions + report
}

func parseLanguageCommand(body string) (string, bool) {
	cleaned := strings.ToLower(strings.TrimSpace(body))
	cleaned = strings.TrimPrefix(cleaned, "/")
	parts := strings.Fields(cleaned)
	if len(parts) == 1 {
		if code := languageFromLabel(parts[0]); code != "" {
			return code, true
		}
		return "", false
	}
	if len(parts) >= 2 && (parts[0] == "language" || parts[0] == "lang") {
		if code := languageFromLabel(parts[1]); code != "" {
			return code, true
		}
	}
	return "", false
}

func languageFromLabel(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if domain.IsSupportedLanguage(value) {
		return value
	}
	switch value {
	case "english", "eng":
		return "en"
	case "hindi", "hin":
		return "hi"
	case "hinglish":
		return "hinglish"
	case "bangla", "bengali":
		return "bn"
	case "tamil":
		return "ta"
	case "telugu":
		return "te"
	case "marathi":
		return "mr"
	case "gujarati":
		return "gu"
	case "kannada":
		return "kn"
	case "malayalam":
		return "ml"
	case "punjabi":
		return "pa"
	case "urdu":
		return "ur"
	}
	return ""
}

func isLanguageGreeting(body string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(body))
	switch cleaned {
	case "hi", "hello", "start", "/start", "namaste", "language", "/language":
		return true
	default:
		return false
	}
}

func languageSelectionMessage() string {
	return "Choose reply language with /language en, /language hinglish, /language hi, /language bn, /language ta, /language te, /language mr, /language gu, /language kn, /language ml, /language pa, or /language ur."
}

func languageUpdatedMessage(code string) string {
	return "Reply language updated to " + domain.LanguageName(code) + ". Send a suspicious message, link, QR, screenshot, or UPI ID to check."
}

func (s *Server) allow(userID string) bool {
	limit := s.config.RateLimitPerMinute
	if limit == 0 {
		limit = 30
	}
	return s.repo.AllowUserEvent(userID, limit, time.Minute)
}

func newEvent(eventType string, correlationID string, causationID string, payload any) domain.EventEnvelope {
	if correlationID == "" {
		correlationID = newEventID("corr")
	}
	return domain.EventEnvelope{
		EventID:       newEventID("evt"),
		EventType:     eventType,
		SchemaVersion: "1.0.0",
		CorrelationID: correlationID,
		CausationID:   causationID,
		CreatedAt:     time.Now().UTC(),
		Producer:      "scamshield-mvp",
		Payload:       payload,
	}
}

func newEventID(prefix string) string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405")
	}
	return prefix + "-" + hex.EncodeToString(bytes[:])
}
