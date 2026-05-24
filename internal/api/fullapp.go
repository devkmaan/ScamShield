package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"scamshield/internal/analysis"
	"scamshield/internal/domain"
	"scamshield/internal/privacy"
)

func (s *Server) handleAdminSummary(w http.ResponseWriter, r *http.Request) {
	decisions := s.repo.ListDecisions(200)
	merchants := s.repo.ListMerchantRisks(10)
	summary := domain.AdminSummary{
		DecisionCount:      len(decisions),
		MerchantCount:      s.repo.MerchantCount(),
		FeedbackCount:      s.repo.FeedbackCount(),
		ReportCount:        s.repo.ReportCount(),
		EvidenceCount:      s.repo.EvidenceCount(),
		EventCount:         s.repo.EventCount(),
		RiskLevelBreakdown: map[domain.RiskLevel]int{},
		ScamTypeBreakdown:  map[domain.ScamType]int{},
		RecentDecisions:    s.repo.ListDecisions(8),
		TopRiskMerchants:   merchants,
	}
	for _, decision := range decisions {
		summary.RiskLevelBreakdown[decision.RiskLevel]++
		summary.ScamTypeBreakdown[decision.ScamType]++
		if decision.RiskLevel == domain.RiskHigh || decision.RiskLevel == domain.RiskCritical {
			summary.HighRiskCount++
		}
		if decision.NeedsHumanReview {
			summary.HumanReviewCount++
		}
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleAdminDecisions(w http.ResponseWriter, r *http.Request) {
	decisions := filterDecisions(s.repo.ListDecisions(0), r)
	page := parsePositiveInt(r, "page", 1)
	pageSize := parsePositiveInt(r, "pageSize", parseLimit(r, 10))
	if pageSize > 100 {
		pageSize = 100
	}
	total := len(decisions)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, domain.DecisionListResponse{
		Items:       decisions[start:end],
		Page:        page,
		PageSize:    pageSize,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     totalPages > 0 && page < totalPages,
		HasPrevious: totalPages > 0 && page > 1,
	})
}

func (s *Server) handleAdminMerchants(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.repo.ListMerchantRisks(parseLimit(r, 50))})
}

func (s *Server) handleAdminFeedback(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.repo.ListFeedback(parseLimit(r, 50))})
}

func (s *Server) handleAdminReports(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.repo.ListReports(parseLimit(r, 50))})
}

func (s *Server) handleDecisionDetail(w http.ResponseWriter, r *http.Request) {
	decisionID := r.PathValue("decisionID")
	decision, ok := s.repo.GetDecision(decisionID)
	if !ok {
		writeError(w, http.StatusNotFound, "decision not found")
		return
	}
	writeJSON(w, http.StatusOK, decision)
}

func (s *Server) handleUserHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.repo.ListUserDecisions(userID, parseLimit(r, 20)),
	})
}

func (s *Server) handleDecisionShare(w http.ResponseWriter, r *http.Request) {
	decisionID := r.PathValue("decisionID")
	decision, ok := s.repo.GetDecision(decisionID)
	if !ok {
		writeError(w, http.StatusNotFound, "decision not found")
		return
	}
	summary := buildDecisionShareSummary(decision)
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleInsightsTrends(w http.ResponseWriter, r *http.Request) {
	decisions := s.repo.ListDecisions(500)
	writeJSON(w, http.StatusOK, buildInsights(decisions))
}

func (s *Server) handleRecoveryStart(w http.ResponseWriter, r *http.Request) {
	var req domain.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Text == "" {
		req.Text = "User requested recovery guidance after a suspected payment scam."
	}
	req.AlreadyPaid = true
	req.Text = privacy.RedactSensitive(req.Text)
	decision := s.orchestrator.Analyze(req)
	s.repo.AppendEvent(newEvent(domain.EventRiskDecisionCreated, newEventID("corr"), "", decision))
	if decision.ReportID != "" {
		if report, ok := s.repo.GetReport(decision.ReportID); ok {
			s.repo.AppendEvent(newEvent(domain.EventRecoveryReportCreated, newEventID("corr"), decision.DecisionID, report))
		}
	}
	writeJSON(w, http.StatusOK, decision)
}

func (s *Server) handleCreateEvidence(w http.ResponseWriter, r *http.Request) {
	var req domain.EvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	now := time.Now().UTC()
	redacted := privacy.RedactSensitive(req.Content)
	hash := sha256.Sum256([]byte(redacted))
	evidence := domain.EvidenceObject{
		EvidenceID:     newEventID("evd"),
		ReportID:       req.ReportID,
		DecisionID:     req.DecisionID,
		UserID:         req.UserID,
		MediaType:      defaultString(req.MediaType, "TEXT"),
		Source:         defaultString(req.Source, "USER_UPLOAD"),
		SHA256:         hex.EncodeToString(hash[:]),
		Preview:        preview(redacted, 180),
		RetentionUntil: now.AddDate(0, 3, 0),
		CreatedAt:      now,
	}
	s.repo.SaveEvidence(evidence)
	s.repo.AppendEvent(newEvent("evidence.created", newEventID("corr"), req.DecisionID, evidence))
	writeJSON(w, http.StatusCreated, evidence)
}

func (s *Server) handleListEvidence(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.repo.ListEvidence(parseLimit(r, 50))})
}

func (s *Server) handleDeleteEvidence(w http.ResponseWriter, r *http.Request) {
	evidenceID := r.PathValue("evidenceID")
	if !s.repo.DeleteEvidence(evidenceID) {
		writeError(w, http.StatusNotFound, "evidence not found")
		return
	}
	s.repo.AppendEvent(newEvent("evidence.deleted", newEventID("corr"), evidenceID, map[string]string{"evidenceId": evidenceID}))
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleModelScoreText(w http.ResponseWriter, r *http.Request) {
	var req domain.ModelScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if s.mlClient != nil {
		if response, err := s.mlClient.ScoreText(r.Context(), req); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	entities := analysis.Extract(domain.CheckRequest{Text: privacy.RedactSensitive(req.Text)})
	signals := append(analysis.EvaluateRules(entities), analysis.ScoreTextModel(entities)...)
	writeJSON(w, http.StatusOK, modelResponse("local-keyword-fallback-v1", signals))
}

func (s *Server) handleModelScoreURL(w http.ResponseWriter, r *http.Request) {
	var req domain.ModelScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if s.mlClient != nil {
		if response, err := s.mlClient.ScoreURL(r.Context(), req); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	entities := analysis.Extract(domain.CheckRequest{URL: req.URL, Text: req.Text})
	signals := analysis.EvaluateRules(entities)
	writeJSON(w, http.StatusOK, modelResponse("url-lexical-rules-v1", signals))
}

func (s *Server) handleModelMetadata(w http.ResponseWriter, r *http.Request) {
	if s.mlClient != nil {
		if response, err := s.mlClient.Metadata(r.Context()); err == nil {
			writeJSON(w, http.StatusOK, response)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"activeModels": map[string]string{
			"text":     "local-keyword-fallback-v1",
			"url":      "url-lexical-rules-v1",
			"merchant": "merchant-graph-rules-v1",
		},
		"mode":      "go-fallback",
		"policy":    "High-risk user decisions must be backed by rules, model, or merchant graph signals; explanation cannot lower risk.",
		"updatedAt": time.Now().UTC(),
	})
}

func (s *Server) handlePayeeObserve(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RawPayee string `json:"rawPayee"`
		Alias    string `json:"alias,omitempty"`
		Source   string `json:"source,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.RawPayee == "" {
		writeError(w, http.StatusBadRequest, "rawPayee is required")
		return
	}
	risk := s.repo.ObservePayee(req.RawPayee, req.Alias)
	writeJSON(w, http.StatusOK, risk)
}

func (s *Server) handlePayeeReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RawPayee string `json:"rawPayee"`
		Alias    string `json:"alias,omitempty"`
		Comment  string `json:"comment,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.RawPayee == "" {
		writeError(w, http.StatusBadRequest, "rawPayee is required")
		return
	}
	risk := s.repo.AddMerchantComplaint(req.RawPayee, req.Alias)
	s.repo.AppendEvent(newEvent(domain.EventMerchantRiskUpdated, newEventID("corr"), "", risk))
	writeJSON(w, http.StatusOK, risk)
}

func (s *Server) handleSimulation(w http.ResponseWriter, r *http.Request) {
	var req domain.SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Count <= 0 {
		req.Count = 6
	}
	if req.Count > 50 {
		req.Count = 50
	}
	samples := []domain.CheckRequest{
		{UserID: "sim-1", InputType: domain.InputText, Text: "Refund receive karne ke liye QR scan karo and UPI PIN enter karo.", Language: "hinglish"},
		{UserID: "sim-2", InputType: domain.InputText, Text: "Your SBI KYC is blocked. Click https://sbi-verify-support.com and share OTP immediately."},
		{UserID: "sim-3", InputType: domain.InputText, Text: "Part time job daily task. Pay registration fee and earn commission."},
		{UserID: "sim-4", InputType: domain.InputText, Text: "Guaranteed return crypto profit. Double money in 7 days."},
		{UserID: "sim-5", InputType: domain.InputText, Text: "Fraud alert from bank. Move your money to safe account immediately."},
		{UserID: "sim-6", InputType: domain.InputQR, QRPayload: "upi://pay?pa=refund-care@okaxis&pn=Marketplace%20Refund%20Agent&am=4999&tn=Refund"},
	}
	simOrchestrator := analysis.NewOrchestrator(s.repo, s.mlClient)
	decisions := make([]domain.RiskDecision, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		check := samples[i%len(samples)]
		decision := simOrchestrator.Analyze(check)
		decisions = append(decisions, decision)
		s.repo.AppendEvent(newEvent(domain.EventRiskDecisionCreated, newEventID("corr"), "", decision))
	}
	writeJSON(w, http.StatusOK, domain.SimulationResponse{Status: "created", Created: len(decisions), Decisions: decisions})
}

func modelResponse(version string, signals []domain.Signal) domain.ModelScoreResponse {
	sort.SliceStable(signals, func(i int, j int) bool {
		return signals[i].Weight > signals[j].Weight
	})
	score := 0.08
	typeScores := map[domain.ScamType]float64{}
	ids := make([]string, 0, len(signals))
	for _, signal := range signals {
		if signal.Weight > 0 {
			score += signal.Weight * (1 - score)
		}
		if signal.ScamType != domain.ScamUnknown {
			typeScores[signal.ScamType] += signal.Weight
		}
		if signal.ID != "" {
			ids = append(ids, signal.ID)
		}
	}
	if score > 0.98 {
		score = 0.98
	}
	confidence := 0.48 + float64(len(signals))*0.06
	if confidence > 0.94 {
		confidence = 0.94
	}
	return domain.ModelScoreResponse{
		ModelVersion:   version,
		Score:          round2(score),
		Confidence:     round2(confidence),
		ScamTypeScores: roundTypeScores(typeScores),
		Signals:        ids,
	}
}

func buildDecisionShareSummary(decision domain.RiskDecision) domain.DecisionShareSummary {
	officialHelp := []string{
		"National Cybercrime Helpline: 1930",
		"National Cyber Crime Reporting Portal: https://cybercrime.gov.in",
	}
	actions := append([]string{}, decision.RecommendedActions...)
	summary := domain.DecisionShareSummary{
		DecisionID:         decision.DecisionID,
		InputType:          decision.InputType,
		Language:           decision.Language,
		RiskLevel:          decision.RiskLevel,
		Score:              decision.Score,
		Confidence:         decision.Confidence,
		ScamType:           decision.ScamType,
		TopSignals:         append([]string{}, decision.TopSignals...),
		UserMessage:        decision.UserMessage,
		RecommendedActions: actions,
		OfficialHelp:       officialHelp,
		ReportID:           decision.ReportID,
		CreatedAt:          decision.CreatedAt,
	}
	summary.ShareText = buildShareText(summary)
	return summary
}

func buildShareText(summary domain.DecisionShareSummary) string {
	var builder strings.Builder
	builder.WriteString("ScamShield verdict: ")
	builder.WriteString(string(summary.RiskLevel))
	builder.WriteString(" (")
	builder.WriteString(string(summary.ScamType))
	builder.WriteString("), score ")
	builder.WriteString(strconv.FormatFloat(summary.Score, 'f', 2, 64))
	builder.WriteString(". ")
	if summary.UserMessage != "" {
		builder.WriteString(summary.UserMessage)
		builder.WriteString(" ")
	}
	if len(summary.RecommendedActions) > 0 {
		builder.WriteString("Action: ")
		builder.WriteString(summary.RecommendedActions[0])
		builder.WriteString(" ")
	}
	builder.WriteString("If money is lost, call 1930 and report at cybercrime.gov.in.")
	return builder.String()
}

func buildInsights(decisions []domain.RiskDecision) domain.InsightsResponse {
	riskCounts := map[string]int{}
	scamCounts := map[string]int{}
	languageCounts := map[string]int{}
	hourly := map[string]*domain.TrendPoint{}
	total := len(decisions)
	for _, decision := range decisions {
		riskCounts[string(decision.RiskLevel)]++
		scamCounts[string(decision.ScamType)]++
		language := decision.Language
		if language == "" {
			language = "en"
		}
		languageCounts[language]++
		label := decision.CreatedAt.Local().Format("15:00")
		point := hourly[label]
		if point == nil {
			point = &domain.TrendPoint{Label: label}
			hourly[label] = point
		}
		point.Count++
		if decision.RiskLevel == domain.RiskHigh || decision.RiskLevel == domain.RiskCritical {
			point.HighRiskCount++
		}
		if decision.NeedsHumanReview {
			point.ReviewCount++
		}
	}
	return domain.InsightsResponse{
		RiskLevels:     bucketsFromCounts(riskCounts, total),
		ScamTypes:      bucketsFromCounts(scamCounts, total),
		Languages:      bucketsFromCounts(languageCounts, total),
		RecentActivity: sortedTrendPoints(hourly),
		GeneratedAt:    time.Now().UTC(),
	}
}

func bucketsFromCounts(counts map[string]int, total int) []domain.InsightBucket {
	buckets := make([]domain.InsightBucket, 0, len(counts))
	for label, count := range counts {
		percentage := 0.0
		if total > 0 {
			percentage = round2(float64(count) / float64(total) * 100)
		}
		buckets = append(buckets, domain.InsightBucket{Label: label, Count: count, Percentage: percentage})
	}
	sort.SliceStable(buckets, func(i int, j int) bool {
		if buckets[i].Count == buckets[j].Count {
			return buckets[i].Label < buckets[j].Label
		}
		return buckets[i].Count > buckets[j].Count
	})
	return buckets
}

func sortedTrendPoints(points map[string]*domain.TrendPoint) []domain.TrendPoint {
	result := make([]domain.TrendPoint, 0, len(points))
	for _, point := range points {
		result = append(result, *point)
	}
	sort.SliceStable(result, func(i int, j int) bool {
		return result[i].Label < result[j].Label
	})
	if len(result) > 12 {
		result = result[len(result)-12:]
	}
	return result
}

func roundTypeScores(values map[domain.ScamType]float64) map[domain.ScamType]float64 {
	for key, value := range values {
		if value > 1 {
			value = 1
		}
		values[key] = round2(value)
	}
	return values
}

func filterDecisions(decisions []domain.RiskDecision, r *http.Request) []domain.RiskDecision {
	query := r.URL.Query()
	scamType := strings.TrimSpace(query.Get("scamType"))
	riskLevel := strings.TrimSpace(query.Get("riskLevel"))
	inputType := strings.TrimSpace(query.Get("inputType"))
	language := strings.TrimSpace(query.Get("language"))
	userID := strings.TrimSpace(query.Get("userId"))
	reviewOnly := strings.EqualFold(strings.TrimSpace(query.Get("needsHumanReview")), "true")
	filtered := make([]domain.RiskDecision, 0, len(decisions))
	for _, decision := range decisions {
		if scamType != "" && !strings.EqualFold(string(decision.ScamType), scamType) {
			continue
		}
		if riskLevel != "" && !strings.EqualFold(string(decision.RiskLevel), riskLevel) {
			continue
		}
		if inputType != "" && !strings.EqualFold(string(decision.InputType), inputType) {
			continue
		}
		if language != "" && !strings.EqualFold(decision.Language, language) {
			continue
		}
		if userID != "" && decision.UserID != userID {
			continue
		}
		if reviewOnly && !decision.NeedsHumanReview {
			continue
		}
		filtered = append(filtered, decision)
	}
	return filtered
}

func parseLimit(r *http.Request, fallback int) int {
	limit := fallback
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return limit
}

func parsePositiveInt(r *http.Request, key string, fallback int) int {
	if fallback <= 0 {
		fallback = 1
	}
	if raw := r.URL.Query().Get(key); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func preview(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
