package analysis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"scamshield/internal/domain"
	"scamshield/internal/store"
)

type Orchestrator struct {
	repo        *store.MemoryStore
	mlClient    *MLClient
	genAIClient *GenAIClient
}

func NewOrchestrator(repo *store.MemoryStore, clients ...*MLClient) *Orchestrator {
	var client *MLClient
	if len(clients) > 0 {
		client = clients[0]
	}
	return &Orchestrator{repo: repo, mlClient: client}
}

func (o *Orchestrator) SetGenAIClient(client *GenAIClient) {
	o.genAIClient = client
}

func (o *Orchestrator) Analyze(req domain.CheckRequest) domain.RiskDecision {
	now := time.Now().UTC()
	if req.InputType == "" {
		req.InputType = domain.InputAuto
	}
	language := domain.NormalizeLanguage(req.Language)
	req.Language = language
	rawEntities := Extract(req)
	normalizedReq, detectedLanguage, normalizerVersion := o.normalizeForAnalysis(req)
	entities := mergeEntities(rawEntities, Extract(normalizedReq))
	inputType := resolveInputType(req, entities)

	signals := EvaluateRules(entities)
	modelVersions := map[string]string{
		"rules":       "rules-v1",
		"merchant":    "merchant-graph-rules-v1",
		"explanation": "template-explainer-v1",
	}
	if normalizerVersion != "" {
		modelVersions["normalizer"] = normalizerVersion
	}

	ruleScore := aggregateScore(signals)
	if shouldUseTextModel(ruleScore, entities) {
		modelSignals, version := o.textModelSignals(entities, language)
		if version != "" {
			modelVersions["text"] = version
		}
		signals = append(signals, modelSignals...)
	}
	if len(entities.URLs) > 0 {
		modelSignals, version := o.urlModelSignals(entities)
		if version != "" {
			modelVersions["url"] = version
		}
		signals = append(signals, modelSignals...)
	}

	payeeHash := ""
	for _, upiID := range entities.UPIIDs {
		alias := ""
		if entities.QR != nil {
			alias = entities.QR.PayeeName
		}
		risk := o.repo.ObservePayee(upiID, alias)
		payeeHash = risk.PayeeHash
		if risk.RiskScore >= 0.65 {
			signals = append(signals, domain.Signal{
				ID:          "merchant_graph_risk",
				Description: "Payee has repeated suspicious reports or risky graph signals.",
				Weight:      risk.RiskScore * 0.45,
				ScamType:    domain.ScamUPICollect,
				Reason:      "This payee has prior complaint or graph-risk signals.",
			})
		} else if risk.ComplaintCount > 0 {
			signals = append(signals, domain.Signal{
				ID:          "merchant_graph_watchlist",
				Description: "Payee has early complaint signals.",
				Weight:      0.18,
				ScamType:    domain.ScamUnknown,
				Reason:      "This payee has limited complaint history and needs verification.",
			})
		}
	}

	sortSignals(signals)
	score := aggregateScore(signals)
	level := riskLevel(score)
	scamType := dominantScamType(signals)
	confidence := confidence(score, signals)

	decision := domain.RiskDecision{
		DecisionID:         newID("dec"),
		UserID:             req.UserID,
		InputType:          inputType,
		Language:           language,
		InputLanguage:      detectedLanguage,
		RiskLevel:          level,
		Score:              round(score),
		Confidence:         round(confidence),
		ScamType:           scamType,
		TopSignals:         signalIDs(signals, 5),
		RecommendedActions: nil,
		NeedsHumanReview:   shouldReview(score, signals),
		PayeeHash:          payeeHash,
		ModelVersions:      modelVersions,
		CreatedAt:          now,
	}
	decision.UserMessage = buildUserMessage(decision, signals)
	decision.RecommendedActions = recommendedActions(decision)
	decision = o.renderDecision(req, decision, signals)

	if req.AlreadyPaid || alreadyPaidText(entities.NormalizedText) {
		report := buildRecoveryReport(req, decision)
		report = o.renderRecoveryReport(req, decision, report)
		o.repo.SaveReport(report)
		decision.ReportID = report.ReportID
	}

	o.repo.SaveDecision(decision)
	return decision
}

func (o *Orchestrator) textModelSignals(entities domain.ExtractedEntities, language string) ([]domain.Signal, string) {
	if o.mlClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel()
		response, err := o.mlClient.ScoreText(ctx, domain.ModelScoreRequest{
			Text:         entities.NormalizedText,
			LanguageHint: language,
			Context:      map[string]string{"source": "risk-core"},
		})
		if err == nil {
			return signalsFromModel("ml_text_classifier", response), response.ModelVersion
		}
	}
	return ScoreTextModel(entities), "local-keyword-fallback-v1"
}

func (o *Orchestrator) normalizeForAnalysis(req domain.CheckRequest) (domain.CheckRequest, string, string) {
	detectedLanguage := req.Language
	if strings.TrimSpace(req.Text) == "" || o.genAIClient == nil {
		return req, detectedLanguage, ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), genAIHTTPTimeout())
	defer cancel()
	response, err := o.genAIClient.NormalizeInput(ctx, domain.GenAINormalizeRequest{
		Text:                 req.Text,
		TargetLanguage:       "en",
		UserSelectedLanguage: req.Language,
		Context: map[string]string{
			"source":    "risk-core",
			"inputType": string(req.InputType),
		},
	})
	if err != nil || strings.TrimSpace(response.NormalizedText) == "" {
		return req, detectedLanguage, ""
	}
	normalized := req
	normalized.Text = response.NormalizedText
	if response.DetectedLanguage != "" {
		detectedLanguage = domain.NormalizeLanguage(response.DetectedLanguage)
	}
	return normalized, detectedLanguage, response.ModelVersion
}

func (o *Orchestrator) renderDecision(req domain.CheckRequest, decision domain.RiskDecision, signals []domain.Signal) domain.RiskDecision {
	if o.genAIClient == nil {
		return decision
	}
	ctx, cancel := context.WithTimeout(context.Background(), genAIHTTPTimeout())
	defer cancel()
	response, err := o.genAIClient.Render(ctx, domain.GenAIRenderRequest{
		Surface:  "risk_decision",
		Language: decision.Language,
		Decision: decision,
		Reasons:  topReasonText(signals, 5),
		Context: map[string]string{
			"source":          "risk-core",
			"alreadyPaid":     fmt.Sprintf("%t", req.AlreadyPaid),
			"immutablePolicy": "GenAI can only render text; Go owns risk facts.",
		},
	})
	if err != nil {
		return decision
	}
	if strings.TrimSpace(response.UserMessage) != "" {
		decision.UserMessage = response.UserMessage
	}
	if len(response.RecommendedActions) > 0 {
		decision.RecommendedActions = enforceRequiredActions(decision.RiskLevel, response.RecommendedActions)
	}
	if response.ModelVersion != "" {
		decision.ModelVersions["genai"] = response.ModelVersion
		decision.ModelVersions["explanation"] = response.ModelVersion
	}
	return decision
}

func (o *Orchestrator) renderRecoveryReport(req domain.CheckRequest, decision domain.RiskDecision, report domain.RecoveryReport) domain.RecoveryReport {
	if o.genAIClient == nil {
		return report
	}
	ctx, cancel := context.WithTimeout(context.Background(), genAIHTTPTimeout())
	defer cancel()
	response, err := o.genAIClient.Render(ctx, domain.GenAIRenderRequest{
		Surface:  "recovery_report",
		Language: decision.Language,
		Decision: decision,
		Report:   &report,
		Context: map[string]string{
			"source":      "risk-core",
			"userId":      req.UserID,
			"officialUrl": "https://cybercrime.gov.in",
		},
	})
	if err != nil {
		return report
	}
	if strings.TrimSpace(response.Summary) != "" {
		report.Summary = response.Summary
	}
	if len(response.Checklist) > 0 {
		report.Checklist = response.Checklist
	}
	if len(response.OfficialHelp) > 0 {
		report.OfficialHelp = enforceOfficialHelp(response.OfficialHelp)
	}
	if response.ModelVersion != "" {
		decision.ModelVersions["genai_report"] = response.ModelVersion
	}
	return report
}

func (o *Orchestrator) urlModelSignals(entities domain.ExtractedEntities) ([]domain.Signal, string) {
	if o.mlClient == nil || len(entities.URLs) == 0 {
		return nil, ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	response, err := o.mlClient.ScoreURL(ctx, domain.ModelScoreRequest{
		URL:     entities.URLs[0].Raw,
		Text:    entities.NormalizedText,
		Context: map[string]string{"source": "risk-core"},
	})
	if err != nil {
		return nil, "url-lexical-rules-v1"
	}
	return signalsFromModel("ml_url_classifier", response), response.ModelVersion
}

func shouldUseTextModel(ruleScore float64, entities domain.ExtractedEntities) bool {
	return entities.NormalizedText != "" && ruleScore < 0.85
}

func signalsFromModel(id string, response domain.ModelScoreResponse) []domain.Signal {
	if response.Score < 0.35 {
		return nil
	}
	scamType := highestScamType(response.ScamTypeScores)
	weight := 0.18 + response.Score*0.55
	if weight > 0.70 {
		weight = 0.70
	}
	if weight < 0.35 {
		weight = 0.35
	}
	reason := "ML model score " + formatScore(response.Score) + " from " + response.ModelVersion
	if len(response.Signals) > 0 {
		reason += " using signals: " + strings.Join(response.Signals, ", ")
	}
	return []domain.Signal{{
		ID:          id,
		Description: "External ML service risk score.",
		Weight:      weight,
		ScamType:    scamType,
		Reason:      reason,
	}}
}

func highestScamType(scores map[domain.ScamType]float64) domain.ScamType {
	selected := domain.ScamUnknown
	best := 0.0
	for scamType, score := range scores {
		if score > best {
			selected = scamType
			best = score
		}
	}
	return selected
}

func mergeEntities(primary domain.ExtractedEntities, secondary domain.ExtractedEntities) domain.ExtractedEntities {
	merged := secondary
	if merged.NormalizedText == "" {
		merged.NormalizedText = primary.NormalizedText
	}
	merged.URLs = mergeURLFindings(primary.URLs, merged.URLs)
	merged.UPIIDs = mergeStrings(primary.UPIIDs, merged.UPIIDs)
	merged.Amounts = mergeStrings(primary.Amounts, merged.Amounts)
	if merged.QR == nil {
		merged.QR = primary.QR
	}
	return merged
}

func mergeURLFindings(primary []domain.URLFinding, secondary []domain.URLFinding) []domain.URLFinding {
	seen := map[string]bool{}
	result := make([]domain.URLFinding, 0, len(primary)+len(secondary))
	for _, item := range append(primary, secondary...) {
		key := strings.ToLower(strings.TrimSpace(item.Raw))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func mergeStrings(primary []string, secondary []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(primary)+len(secondary))
	for _, item := range append(primary, secondary...) {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func enforceRequiredActions(level domain.RiskLevel, actions []string) []string {
	clean := make([]string, 0, len(actions)+1)
	seen := map[string]bool{}
	for _, action := range actions {
		action = strings.TrimSpace(action)
		if action == "" {
			continue
		}
		key := strings.ToLower(action)
		if seen[key] {
			continue
		}
		seen[key] = true
		clean = append(clean, action)
	}
	if level == domain.RiskHigh || level == domain.RiskCritical {
		required := "If money is already lost, contact your bank, call 1930, and file at cybercrime.gov.in."
		has1930 := false
		hasCybercrime := false
		for _, action := range clean {
			has1930 = has1930 || strings.Contains(action, "1930")
			hasCybercrime = hasCybercrime || strings.Contains(strings.ToLower(action), "cybercrime.gov.in")
		}
		if !has1930 || !hasCybercrime {
			clean = append(clean, required)
		}
	}
	return clean
}

func enforceOfficialHelp(items []string) []string {
	required := []string{
		"National Cybercrime Helpline: 1930",
		"National Cyber Crime Reporting Portal: https://cybercrime.gov.in",
		"Your bank/payment app's official fraud support channel",
	}
	result := make([]string, 0, len(items)+len(required))
	seen := map[string]bool{}
	for _, item := range append(items, required...) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func formatScore(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func buildRecoveryReport(req domain.CheckRequest, decision domain.RiskDecision) domain.RecoveryReport {
	return domain.RecoveryReport{
		ReportID:   newID("rep"),
		UserID:     req.UserID,
		DecisionID: decision.DecisionID,
		Status:     "DRAFT_GUIDANCE_ONLY",
		Summary:    "Possible cyber financial fraud. This draft helps collect evidence and act quickly; it is not an official complaint.",
		Checklist: []string{
			"Immediately contact your bank/payment app support and ask for fraud dispute/freeze support.",
			"Call 1930 as soon as possible if money has moved recently.",
			"File a complaint at cybercrime.gov.in and keep the acknowledgement number.",
			"Save screenshots of chats, phone numbers, UPI IDs, transaction IDs, URLs, and payment receipts.",
			"Do not delete chats/call logs and do not engage further with the suspected scammer.",
		},
		OfficialHelp: []string{
			"National Cybercrime Helpline: 1930",
			"National Cyber Crime Reporting Portal: https://cybercrime.gov.in",
			"Your bank/payment app's official fraud support channel",
		},
		CreatedAt: time.Now().UTC(),
	}
}

func resolveInputType(req domain.CheckRequest, entities domain.ExtractedEntities) domain.InputType {
	if req.InputType != "" && req.InputType != domain.InputAuto {
		return req.InputType
	}
	if entities.QR != nil {
		return domain.InputQR
	}
	if req.UPIID != "" || len(entities.UPIIDs) > 0 {
		return domain.InputUPIID
	}
	if req.URL != "" || len(entities.URLs) > 0 {
		return domain.InputURL
	}
	if req.MediaRef != "" {
		return domain.InputScreenshot
	}
	return domain.InputText
}

func aggregateScore(signals []domain.Signal) float64 {
	if len(signals) == 0 {
		return 0.08
	}
	var clean float64
	for _, signal := range signals {
		if signal.Weight <= 0 {
			continue
		}
		clean += signal.Weight * (1 - clean)
	}
	if clean > 0.98 {
		return 0.98
	}
	return clean
}

func riskLevel(score float64) domain.RiskLevel {
	switch {
	case score >= 0.85:
		return domain.RiskCritical
	case score >= 0.65:
		return domain.RiskHigh
	case score >= 0.35:
		return domain.RiskCaution
	default:
		return domain.RiskLow
	}
}

func dominantScamType(signals []domain.Signal) domain.ScamType {
	if len(signals) == 0 {
		return domain.ScamUnknown
	}
	scores := map[domain.ScamType]float64{}
	for _, signal := range signals {
		if signal.ScamType == domain.ScamUnknown {
			continue
		}
		scores[signal.ScamType] += signal.Weight
	}
	var selected domain.ScamType = domain.ScamUnknown
	var best float64
	for scamType, score := range scores {
		if score > best {
			best = score
			selected = scamType
		}
	}
	return selected
}

func confidence(score float64, signals []domain.Signal) float64 {
	if len(signals) == 0 {
		return 0.45
	}
	value := 0.52 + math.Min(0.35, float64(len(signals))*0.06)
	if score >= 0.85 {
		value += 0.07
	}
	if value > 0.96 {
		return 0.96
	}
	return value
}

func shouldReview(score float64, signals []domain.Signal) bool {
	if score >= 0.35 && score < 0.65 {
		return true
	}
	for _, signal := range signals {
		if strings.HasPrefix(signal.ID, "merchant_graph") {
			return true
		}
	}
	return false
}

func signalIDs(signals []domain.Signal, limit int) []string {
	ids := []string{}
	seen := map[string]bool{}
	for _, signal := range signals {
		if signal.ID == "" || seen[signal.ID] {
			continue
		}
		seen[signal.ID] = true
		ids = append(ids, signal.ID)
		if len(ids) == limit {
			break
		}
	}
	return ids
}

func sortSignals(signals []domain.Signal) {
	sort.SliceStable(signals, func(i int, j int) bool {
		return signals[i].Weight > signals[j].Weight
	})
}

func alreadyPaidText(text string) bool {
	return strings.Contains(text, "already paid") ||
		strings.Contains(text, "i paid") ||
		strings.Contains(text, "maine pay") ||
		strings.Contains(text, "paise chale gaye") ||
		strings.Contains(text, "money lost")
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func newID(prefix string) string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405")
	}
	return prefix + "-" + hex.EncodeToString(bytes[:])
}
