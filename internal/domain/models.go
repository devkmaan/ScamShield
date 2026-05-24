package domain

import "time"

type InputType string

const (
	InputAuto       InputType = "AUTO"
	InputText       InputType = "TEXT"
	InputURL        InputType = "URL"
	InputQR         InputType = "QR"
	InputUPIID      InputType = "UPI_ID"
	InputScreenshot InputType = "SCREENSHOT"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskCaution  RiskLevel = "CAUTION"
	RiskHigh     RiskLevel = "HIGH_RISK"
	RiskCritical RiskLevel = "CRITICAL"
)

type ScamType string

const (
	ScamUnknown       ScamType = "UNKNOWN"
	ScamUPICollect    ScamType = "UPI_COLLECT"
	ScamPhishing      ScamType = "PHISHING"
	ScamImpersonation ScamType = "IMPERSONATION"
	ScamJob           ScamType = "JOB_SCAM"
	ScamInvestment    ScamType = "INVESTMENT"
	ScamLoan          ScamType = "LOAN_APP"
	ScamFakeReceipt   ScamType = "FAKE_RECEIPT"
)

type CheckRequest struct {
	UserID      string    `json:"userId,omitempty"`
	InputType   InputType `json:"inputType,omitempty"`
	Text        string    `json:"text,omitempty"`
	URL         string    `json:"url,omitempty"`
	UPIID       string    `json:"upiId,omitempty"`
	QRPayload   string    `json:"qrPayload,omitempty"`
	MediaRef    string    `json:"mediaRef,omitempty"`
	AlreadyPaid bool      `json:"alreadyPaid,omitempty"`
	Language    string    `json:"language,omitempty"`
}

type RiskDecision struct {
	DecisionID         string            `json:"decisionId"`
	UserID             string            `json:"userId,omitempty"`
	InputType          InputType         `json:"inputType"`
	Language           string            `json:"language,omitempty"`
	InputLanguage      string            `json:"inputLanguage,omitempty"`
	RiskLevel          RiskLevel         `json:"riskLevel"`
	Score              float64           `json:"score"`
	Confidence         float64           `json:"confidence"`
	ScamType           ScamType          `json:"scamType"`
	TopSignals         []string          `json:"topSignals"`
	UserMessage        string            `json:"userMessage"`
	RecommendedActions []string          `json:"recommendedActions"`
	NeedsHumanReview   bool              `json:"needsHumanReview"`
	PayeeHash          string            `json:"payeeHash,omitempty"`
	ReportID           string            `json:"reportId,omitempty"`
	ModelVersions      map[string]string `json:"modelVersions,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
}

type Signal struct {
	ID          string
	Description string
	Weight      float64
	ScamType    ScamType
	Reason      string
}

type ExtractedEntities struct {
	NormalizedText string
	URLs           []URLFinding
	UPIIDs         []string
	QR             *QRFinding
	Amounts        []string
}

type URLFinding struct {
	Raw         string
	Host        string
	Scheme      string
	IsShortener bool
}

type QRFinding struct {
	RawPayload string `json:"rawPayload,omitempty"`
	PayeeUPI   string `json:"payeeUpi,omitempty"`
	PayeeName  string `json:"payeeName,omitempty"`
	Amount     string `json:"amount,omitempty"`
	Note       string `json:"note,omitempty"`
}

type MerchantRisk struct {
	PayeeHash          string    `json:"payeeHash"`
	RiskScore          float64   `json:"riskScore"`
	ComplaintCount     int       `json:"complaintCount"`
	Aliases            []string  `json:"aliases"`
	FirstSeen          time.Time `json:"firstSeen"`
	LastSeen           time.Time `json:"lastSeen"`
	NeedsHumanReview   bool      `json:"needsHumanReview"`
	ConnectedRiskCount int       `json:"connectedRiskCount"`
}

type FeedbackRequest struct {
	DecisionID string `json:"decisionId"`
	UserID     string `json:"userId,omitempty"`
	Verdict    string `json:"verdict"`
	PayeeUPI   string `json:"payeeUpi,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

type FeedbackResponse struct {
	Status    string `json:"status"`
	PayeeHash string `json:"payeeHash,omitempty"`
	Message   string `json:"message"`
}

type EvidenceObject struct {
	EvidenceID     string    `json:"evidenceId"`
	ReportID       string    `json:"reportId,omitempty"`
	DecisionID     string    `json:"decisionId,omitempty"`
	UserID         string    `json:"userId,omitempty"`
	MediaType      string    `json:"mediaType"`
	Source         string    `json:"source"`
	SHA256         string    `json:"sha256"`
	Preview        string    `json:"preview,omitempty"`
	RetentionUntil time.Time `json:"retentionUntil"`
	CreatedAt      time.Time `json:"createdAt"`
}

type EvidenceRequest struct {
	ReportID   string `json:"reportId,omitempty"`
	DecisionID string `json:"decisionId,omitempty"`
	UserID     string `json:"userId,omitempty"`
	MediaType  string `json:"mediaType"`
	Source     string `json:"source"`
	Content    string `json:"content"`
}

type RecoveryReport struct {
	ReportID     string    `json:"reportId"`
	UserID       string    `json:"userId,omitempty"`
	DecisionID   string    `json:"decisionId,omitempty"`
	Status       string    `json:"status"`
	Summary      string    `json:"summary"`
	Checklist    []string  `json:"checklist"`
	OfficialHelp []string  `json:"officialHelp"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminSummary struct {
	DecisionCount      int               `json:"decisionCount"`
	HighRiskCount      int               `json:"highRiskCount"`
	HumanReviewCount   int               `json:"humanReviewCount"`
	MerchantCount      int               `json:"merchantCount"`
	FeedbackCount      int               `json:"feedbackCount"`
	ReportCount        int               `json:"reportCount"`
	EvidenceCount      int               `json:"evidenceCount"`
	EventCount         int               `json:"eventCount"`
	RiskLevelBreakdown map[RiskLevel]int `json:"riskLevelBreakdown"`
	ScamTypeBreakdown  map[ScamType]int  `json:"scamTypeBreakdown"`
	RecentDecisions    []RiskDecision    `json:"recentDecisions"`
	TopRiskMerchants   []MerchantRisk    `json:"topRiskMerchants"`
}

type DecisionListResponse struct {
	Items       []RiskDecision `json:"items"`
	Page        int            `json:"page"`
	PageSize    int            `json:"pageSize"`
	Total       int            `json:"total"`
	TotalPages  int            `json:"totalPages"`
	HasNext     bool           `json:"hasNext"`
	HasPrevious bool           `json:"hasPrevious"`
}

type DecisionShareSummary struct {
	DecisionID         string    `json:"decisionId"`
	InputType          InputType `json:"inputType"`
	Language           string    `json:"language,omitempty"`
	RiskLevel          RiskLevel `json:"riskLevel"`
	Score              float64   `json:"score"`
	Confidence         float64   `json:"confidence"`
	ScamType           ScamType  `json:"scamType"`
	TopSignals         []string  `json:"topSignals"`
	UserMessage        string    `json:"userMessage"`
	RecommendedActions []string  `json:"recommendedActions"`
	OfficialHelp       []string  `json:"officialHelp"`
	ReportID           string    `json:"reportId,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	ShareText          string    `json:"shareText"`
}

type InsightBucket struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type TrendPoint struct {
	Label         string `json:"label"`
	Count         int    `json:"count"`
	HighRiskCount int    `json:"highRiskCount"`
	ReviewCount   int    `json:"reviewCount"`
}

type InsightsResponse struct {
	RiskLevels     []InsightBucket `json:"riskLevels"`
	ScamTypes      []InsightBucket `json:"scamTypes"`
	Languages      []InsightBucket `json:"languages"`
	RecentActivity []TrendPoint    `json:"recentActivity"`
	GeneratedAt    time.Time       `json:"generatedAt"`
}

type ModelScoreRequest struct {
	Text         string            `json:"text,omitempty"`
	URL          string            `json:"url,omitempty"`
	LanguageHint string            `json:"languageHint,omitempty"`
	Context      map[string]string `json:"context,omitempty"`
}

type ModelScoreResponse struct {
	ModelVersion   string               `json:"modelVersion"`
	Score          float64              `json:"score"`
	Confidence     float64              `json:"confidence"`
	ScamTypeScores map[ScamType]float64 `json:"scamTypeScores"`
	Signals        []string             `json:"signals"`
}

type GenAINormalizeRequest struct {
	Text                 string            `json:"text,omitempty"`
	TargetLanguage       string            `json:"targetLanguage,omitempty"`
	UserSelectedLanguage string            `json:"userSelectedLanguage,omitempty"`
	Context              map[string]string `json:"context,omitempty"`
}

type GenAINormalizeResponse struct {
	DetectedLanguage string `json:"detectedLanguage"`
	NormalizedText   string `json:"normalizedText"`
	InputSummary     string `json:"inputSummary,omitempty"`
	ModelVersion     string `json:"modelVersion"`
	FallbackUsed     bool   `json:"fallbackUsed"`
}

type GenAIRenderRequest struct {
	Surface  string            `json:"surface"`
	Language string            `json:"language"`
	Decision RiskDecision      `json:"decision"`
	Reasons  []string          `json:"reasons,omitempty"`
	Report   *RecoveryReport   `json:"report,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

type GenAIRenderResponse struct {
	Language           string   `json:"language"`
	UserMessage        string   `json:"userMessage,omitempty"`
	RecommendedActions []string `json:"recommendedActions,omitempty"`
	Summary            string   `json:"summary,omitempty"`
	Checklist          []string `json:"checklist,omitempty"`
	OfficialHelp       []string `json:"officialHelp,omitempty"`
	Reply              string   `json:"reply,omitempty"`
	SuggestedActions   []string `json:"suggestedActions,omitempty"`
	ModelVersion       string   `json:"modelVersion"`
	FallbackUsed       bool     `json:"fallbackUsed"`
}

type GenAIChatRequest struct {
	Language string         `json:"language"`
	Message  string         `json:"message"`
	Context  map[string]any `json:"context,omitempty"`
}

type GenAIChatResponse struct {
	Language         string   `json:"language"`
	Reply            string   `json:"reply"`
	SuggestedActions []string `json:"suggestedActions,omitempty"`
	ModelVersion     string   `json:"modelVersion"`
	FallbackUsed     bool     `json:"fallbackUsed"`
}

type UIBundleRequest struct {
	Language string            `json:"language"`
	Keys     map[string]string `json:"keys"`
}

type UIBundleResponse struct {
	Language     string            `json:"language"`
	Bundle       map[string]string `json:"bundle"`
	ModelVersion string            `json:"modelVersion"`
	FallbackUsed bool              `json:"fallbackUsed"`
}

type SimulationRequest struct {
	Count int    `json:"count"`
	Mode  string `json:"mode,omitempty"`
}

type SimulationResponse struct {
	Status    string         `json:"status"`
	Created   int            `json:"created"`
	Decisions []RiskDecision `json:"decisions"`
}

type WhatsAppInbound struct {
	MessageID     string    `json:"messageId"`
	From          string    `json:"from"`
	Type          string    `json:"type"`
	Body          string    `json:"body"`
	MediaID       string    `json:"mediaId,omitempty"`
	Caption       string    `json:"caption,omitempty"`
	CorrelationID string    `json:"correlationId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type WhatsAppReply struct {
	To        string    `json:"to"`
	MessageID string    `json:"messageId,omitempty"`
	Text      string    `json:"text"`
	Buttons   []string  `json:"buttons"`
	CreatedAt time.Time `json:"createdAt"`
}
