package store

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"scamshield/internal/domain"
)

type MemoryStore struct {
	mu          sync.RWMutex
	hashSalt    string
	decisions   map[string]domain.RiskDecision
	merchant    map[string]domain.MerchantRisk
	reports     map[string]domain.RecoveryReport
	evidence    map[string]domain.EvidenceObject
	events      []domain.EventEnvelope
	outbox      []domain.WhatsAppReply
	feedbackLog []domain.FeedbackRequest
	seenMessage map[string]time.Time
	rateWindows map[string][]time.Time
	languages   map[string]string
}

func NewMemoryStore(hashSalt string) *MemoryStore {
	return &MemoryStore{
		hashSalt:    hashSalt,
		decisions:   make(map[string]domain.RiskDecision),
		merchant:    make(map[string]domain.MerchantRisk),
		reports:     make(map[string]domain.RecoveryReport),
		evidence:    make(map[string]domain.EvidenceObject),
		seenMessage: make(map[string]time.Time),
		rateWindows: make(map[string][]time.Time),
		languages:   make(map[string]string),
	}
}

func (s *MemoryStore) SeedDemoMerchants() {
	for _, seed := range []struct {
		upi        string
		aliases    []string
		complaints int
		score      float64
	}{
		{"kyc-helpdesk@upi", []string{"RBI KYC Helpdesk", "Bank Verification Desk"}, 7, 0.86},
		{"refund-care@okaxis", []string{"Marketplace Refund Agent"}, 5, 0.78},
		{"taskbonus@paytm", []string{"Daily Task Commission"}, 9, 0.91},
	} {
		hash := s.HashPayee(seed.upi)
		now := time.Now().UTC()
		s.merchant[hash] = domain.MerchantRisk{
			PayeeHash:        hash,
			RiskScore:        seed.score,
			ComplaintCount:   seed.complaints,
			Aliases:          seed.aliases,
			FirstSeen:        now.AddDate(0, -2, 0),
			LastSeen:         now,
			NeedsHumanReview: true,
		}
	}
}

func (s *MemoryStore) HashPayee(raw string) string {
	canonical := strings.ToLower(strings.TrimSpace(raw))
	sum := sha256.Sum256([]byte(s.hashSalt + ":" + canonical))
	return hex.EncodeToString(sum[:])
}

func (s *MemoryStore) SaveDecision(decision domain.RiskDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decisions[decision.DecisionID] = decision
}

func (s *MemoryStore) DecisionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.decisions)
}

func (s *MemoryStore) GetDecision(id string) (domain.RiskDecision, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	decision, ok := s.decisions[id]
	return decision, ok
}

func (s *MemoryStore) ListDecisions(limit int) []domain.RiskDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.RiskDecision, 0, len(s.decisions))
	for _, decision := range s.decisions {
		items = append(items, decision)
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) ListUserDecisions(userID string, limit int) []domain.RiskDecision {
	userID = strings.TrimSpace(userID)
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.RiskDecision, 0)
	for _, decision := range s.decisions {
		if userID == "" || decision.UserID == userID {
			items = append(items, decision)
		}
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) GetMerchantRisk(payeeHash string) (domain.MerchantRisk, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	risk, ok := s.merchant[payeeHash]
	return risk, ok
}

func (s *MemoryStore) ListMerchantRisks(limit int) []domain.MerchantRisk {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.MerchantRisk, 0, len(s.merchant))
	for _, risk := range s.merchant {
		items = append(items, risk)
	}
	sort.Slice(items, func(i int, j int) bool {
		if items[i].RiskScore == items[j].RiskScore {
			return items[i].ComplaintCount > items[j].ComplaintCount
		}
		return items[i].RiskScore > items[j].RiskScore
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) MerchantCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.merchant)
}

func (s *MemoryStore) ObservePayee(rawUPI string, alias string) domain.MerchantRisk {
	hash := s.HashPayee(rawUPI)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	risk, ok := s.merchant[hash]
	if !ok {
		risk = domain.MerchantRisk{
			PayeeHash: hash,
			RiskScore: 0.18,
			FirstSeen: now,
			Aliases:   []string{},
			LastSeen:  now,
		}
	}
	risk.LastSeen = now
	if alias != "" && !containsFold(risk.Aliases, alias) {
		risk.Aliases = append(risk.Aliases, alias)
		sort.Strings(risk.Aliases)
	}
	if risk.ComplaintCount >= 3 && risk.RiskScore < 0.65 {
		risk.RiskScore = 0.65
	}
	s.merchant[hash] = risk
	return risk
}

func (s *MemoryStore) AddMerchantComplaint(rawUPI string, alias string) domain.MerchantRisk {
	risk := s.ObservePayee(rawUPI, alias)
	s.mu.Lock()
	defer s.mu.Unlock()
	risk.ComplaintCount++
	risk.NeedsHumanReview = true
	if risk.ComplaintCount >= 6 {
		risk.RiskScore = maxFloat(risk.RiskScore, 0.84)
	} else if risk.ComplaintCount >= 3 {
		risk.RiskScore = maxFloat(risk.RiskScore, 0.68)
	} else {
		risk.RiskScore = maxFloat(risk.RiskScore, 0.42)
	}
	risk.LastSeen = time.Now().UTC()
	s.merchant[risk.PayeeHash] = risk
	return risk
}

func (s *MemoryStore) SaveFeedback(feedback domain.FeedbackRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbackLog = append(s.feedbackLog, feedback)
}

func (s *MemoryStore) ListFeedback(limit int) []domain.FeedbackRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.FeedbackRequest, len(s.feedbackLog))
	copy(items, s.feedbackLog)
	if limit > 0 && len(items) > limit {
		items = items[len(items)-limit:]
	}
	return items
}

func (s *MemoryStore) FeedbackCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.feedbackLog)
}

func (s *MemoryStore) SaveReport(report domain.RecoveryReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[report.ReportID] = report
}

func (s *MemoryStore) GetReport(id string) (domain.RecoveryReport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, ok := s.reports[id]
	return report, ok
}

func (s *MemoryStore) ListReports(limit int) []domain.RecoveryReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.RecoveryReport, 0, len(s.reports))
	for _, report := range s.reports {
		items = append(items, report)
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) ReportCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.reports)
}

func (s *MemoryStore) SaveEvidence(evidence domain.EvidenceObject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evidence[evidence.EvidenceID] = evidence
}

func (s *MemoryStore) DeleteEvidence(evidenceID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.evidence[evidenceID]; !ok {
		return false
	}
	delete(s.evidence, evidenceID)
	return true
}

func (s *MemoryStore) ListEvidence(limit int) []domain.EvidenceObject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.EvidenceObject, 0, len(s.evidence))
	for _, evidence := range s.evidence {
		items = append(items, evidence)
	}
	sort.Slice(items, func(i int, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) EvidenceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.evidence)
}

func (s *MemoryStore) SaveReply(reply domain.WhatsAppReply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outbox = append(s.outbox, reply)
}

func (s *MemoryStore) AppendEvent(event domain.EventEnvelope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *MemoryStore) RecentEvents(limit int) []domain.EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}
	start := len(s.events) - limit
	result := make([]domain.EventEnvelope, limit)
	copy(result, s.events[start:])
	return result
}

func (s *MemoryStore) EventCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

func (s *MemoryStore) OutboxFor(userID string) []domain.WhatsAppReply {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var replies []domain.WhatsAppReply
	for _, reply := range s.outbox {
		if userID == "" || reply.To == userID {
			replies = append(replies, reply)
		}
	}
	return replies
}

func (s *MemoryStore) SetUserLanguage(userID string, language string) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.languages[userID] = domain.NormalizeLanguage(language)
}

func (s *MemoryStore) GetUserLanguage(userID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.languages[userID]
}

func (s *MemoryStore) MarkMessageSeen(messageID string, ttl time.Duration) bool {
	if messageID == "" {
		return false
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, seenAt := range s.seenMessage {
		if now.Sub(seenAt) > ttl {
			delete(s.seenMessage, id)
		}
	}
	if _, exists := s.seenMessage[messageID]; exists {
		return true
	}
	s.seenMessage[messageID] = now
	return false
}

func (s *MemoryStore) AllowUserEvent(userID string, limit int, window time.Duration) bool {
	if userID == "" || limit <= 0 {
		return true
	}
	now := time.Now().UTC()
	cutoff := now.Add(-window)
	s.mu.Lock()
	defer s.mu.Unlock()
	history := s.rateWindows[userID]
	filtered := history[:0]
	for _, at := range history {
		if at.After(cutoff) {
			filtered = append(filtered, at)
		}
	}
	if len(filtered) >= limit {
		s.rateWindows[userID] = filtered
		return false
	}
	filtered = append(filtered, now)
	s.rateWindows[userID] = filtered
	return true
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func maxFloat(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
