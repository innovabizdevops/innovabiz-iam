/**
 * @file assessment_models.go
 * @description Define modelos para orquestração de avaliações do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package models

import (
	"time"

	"innovabiz/iam/src/bureau-credito/adapters"
	"innovabiz/iam/src/bureau-credito/fraud-detection"
	"innovabiz/iam/src/bureau-credito/risk-engine"
)

// AssessmentType define os tipos de avaliação suportados
type AssessmentType string

const (
	TypeIdentity    AssessmentType = "IDENTITY"
	TypeCredit      AssessmentType = "CREDIT"
	TypeFraud       AssessmentType = "FRAUD"
	TypeCompliance  AssessmentType = "COMPLIANCE"
	TypeRisk        AssessmentType = "RISK"
	TypeComprehensive AssessmentType = "COMPREHENSIVE"
)

// AssessmentStatus representa o status de uma avaliação
type AssessmentStatus string

const (
	StatusPending   AssessmentStatus = "PENDING"
	StatusProcessing AssessmentStatus = "PROCESSING"
	StatusCompleted AssessmentStatus = "COMPLETED"
	StatusFailed    AssessmentStatus = "FAILED"
	StatusCancelled AssessmentStatus = "CANCELLED"
	StatusTimeout   AssessmentStatus = "TIMEOUT"
)

// AssessmentRequest representa uma solicitação de avaliação orquestrada
type AssessmentRequest struct {
	// Identificadores
	RequestID       string            `json:"requestId"`
	CorrelationID   string            `json:"correlationId,omitempty"`
	UserID          string            `json:"userId"`
	TenantID        string            `json:"tenantId"`
	
	// Metadados
	RequestTimestamp time.Time         `json:"requestTimestamp"`
	Priority         int               `json:"priority"` // 1-5, onde 5 é maior prioridade
	Source           string            `json:"source,omitempty"`
	
	// Configuração de avaliação
	AssessmentTypes  []AssessmentType  `json:"assessmentTypes"`
	CreditProviders  []string          `json:"creditProviders,omitempty"`
	IdentityProviders []string         `json:"identityProviders,omitempty"`
	ComplianceRules  []string          `json:"complianceRules,omitempty"`
	
	// Dados para avaliação
	IdentityData     *IdentityData     `json:"identityData,omitempty"`
	CreditData       *CreditData       `json:"creditData,omitempty"`
	DeviceData       *DeviceData       `json:"deviceData,omitempty"`
	NetworkData      *NetworkData      `json:"networkData,omitempty"`
	TransactionData  *TransactionData  `json:"transactionData,omitempty"`
	BehavioralData   *BehavioralData   `json:"behavioralData,omitempty"`
	
	// Configurações de processamento
	Timeout          time.Duration     `json:"timeout,omitempty"`
	ForceRefresh     bool              `json:"forceRefresh"`
	RequireAllResults bool             `json:"requireAllResults"`
	FailFast         bool              `json:"failFast"`
	
	// Dados adicionais específicos do contexto
	CustomAttributes map[string]interface{} `json:"customAttributes,omitempty"`
}

// IdentityData contém informações de identidade para avaliação
type IdentityData struct {
	DocumentNumber   string            `json:"documentNumber,omitempty"`
	DocumentType     string            `json:"documentType,omitempty"`
	Name             string            `json:"name,omitempty"`
	DateOfBirth      string            `json:"dateOfBirth,omitempty"`
	Email            string            `json:"email,omitempty"`
	PhoneNumber      string            `json:"phoneNumber,omitempty"`
	Address          string            `json:"address,omitempty"`
	Nationality      string            `json:"nationality,omitempty"`
	BiometricData    map[string]string `json:"biometricData,omitempty"`
	VerificationLevel int               `json:"verificationLevel"`
}

// CreditData contém informações financeiras para avaliação de crédito
type CreditData struct {
	AccountAge       int               `json:"accountAge,omitempty"` // Idade da conta em dias
	PaymentHistory   string            `json:"paymentHistory,omitempty"`
	CreditHistory    string            `json:"creditHistory,omitempty"`
	AnnualIncome     float64           `json:"annualIncome,omitempty"`
	Occupation       string            `json:"occupation,omitempty"`
	EmploymentStatus string            `json:"employmentStatus,omitempty"`
	Assets           float64           `json:"assets,omitempty"`
	Liabilities      float64           `json:"liabilities,omitempty"`
	HasPendingLoans  bool              `json:"hasPendingLoans"`
	PendingLoansAmount float64         `json:"pendingLoansAmount,omitempty"`
}

// DeviceData contém informações sobre o dispositivo
type DeviceData struct {
	DeviceID         string            `json:"deviceId"`
	DeviceType       string            `json:"deviceType"`
	OS               string            `json:"os,omitempty"`
	OSVersion        string            `json:"osVersion,omitempty"`
	Browser          string            `json:"browser,omitempty"`
	BrowserVersion   string            `json:"browserVersion,omitempty"`
	ScreenResolution string            `json:"screenResolution,omitempty"`
	DeviceModel      string            `json:"deviceModel,omitempty"`
	DeviceBrand      string            `json:"deviceBrand,omitempty"`
	Jailbroken       bool              `json:"jailbroken"`
	Emulator         bool              `json:"emulator"`
	DeviceLanguage   string            `json:"deviceLanguage,omitempty"`
	TimeZone         string            `json:"timeZone,omitempty"`
	DeviceFingerprint string           `json:"deviceFingerprint,omitempty"`
}

// NetworkData contém informações de rede
type NetworkData struct {
	IPAddress        string            `json:"ipAddress"`
	ISP              string            `json:"isp,omitempty"`
	ConnectionType   string            `json:"connectionType,omitempty"`
	HostName         string            `json:"hostName,omitempty"`
	ASNumber         string            `json:"asNumber,omitempty"`
	ProxyDetected    bool              `json:"proxyDetected"`
	VPNDetected      bool              `json:"vpnDetected"`
	TorDetected      bool              `json:"torDetected"`
	Latitude         float64           `json:"latitude,omitempty"`
	Longitude        float64           `json:"longitude,omitempty"`
	Country          string            `json:"country,omitempty"`
	Region           string            `json:"region,omitempty"`
	City             string            `json:"city,omitempty"`
}

// TransactionData contém informações sobre transações financeiras
type TransactionData struct {
	TransactionID    string            `json:"transactionId"`
	TransactionType  string            `json:"transactionType"`
	Amount           float64           `json:"amount"`
	Currency         string            `json:"currency"`
	Timestamp        time.Time         `json:"timestamp"`
	MerchantID       string            `json:"merchantId,omitempty"`
	MerchantName     string            `json:"merchantName,omitempty"`
	MerchantCategory string            `json:"merchantCategory,omitempty"`
	Description      string            `json:"description,omitempty"`
	PaymentMethod    string            `json:"paymentMethod,omitempty"`
	RecipientID      string            `json:"recipientId,omitempty"`
	SourceAccount    string            `json:"sourceAccount,omitempty"`
	DestinationAccount string          `json:"destinationAccount,omitempty"`
}

// BehavioralData contém informações comportamentais do usuário
type BehavioralData struct {
	SessionID        string            `json:"sessionId,omitempty"`
	SessionDuration  int               `json:"sessionDuration,omitempty"` // Duração em segundos
	ClickPattern     string            `json:"clickPattern,omitempty"`
	TypingSpeed      int               `json:"typingSpeed,omitempty"`
	NavigationFlow   []string          `json:"navigationFlow,omitempty"`
	TimeOnPage       int               `json:"timeOnPage,omitempty"`
	InteractionCount int               `json:"interactionCount,omitempty"`
	UnusualActivity  bool              `json:"unusualActivity"`
	ActivityDetails  map[string]interface{} `json:"activityDetails,omitempty"`
}

// AssessmentResponse representa a resposta completa da avaliação orquestrada
type AssessmentResponse struct {
	// Identificadores
	ResponseID      string            `json:"responseId"`
	RequestID       string            `json:"requestId"`
	CorrelationID   string            `json:"correlationId,omitempty"`
	UserID          string            `json:"userId"`
	TenantID        string            `json:"tenantId"`
	
	// Status
	Status          AssessmentStatus  `json:"status"`
	CompletedAt     time.Time         `json:"completedAt,omitempty"`
	ProcessingTimeMs int64            `json:"processingTimeMs"`
	
	// Resultados consolidados
	TrustScore      int               `json:"trustScore"`        // 0-100
	RiskLevel       string            `json:"riskLevel"`         // LOW, MEDIUM, HIGH, VERY_HIGH
	Decision        string            `json:"decision"`          // APPROVE, REJECT, REVIEW
	Confidence      float64           `json:"confidence"`        // 0-100
	
	// Resultados detalhados
	IdentityResults *IdentityResults  `json:"identityResults,omitempty"`
	CreditResults   *CreditResults    `json:"creditResults,omitempty"`
	FraudResults    *FraudResults     `json:"fraudResults,omitempty"`
	ComplianceResults *ComplianceResults `json:"complianceResults,omitempty"`
	RiskResults     *RiskResults      `json:"riskResults,omitempty"`
	
	// Ações recomendadas
	RequiredActions []string          `json:"requiredActions,omitempty"`
	SuggestedActions []string         `json:"suggestedActions,omitempty"`
	Warnings        []string          `json:"warnings,omitempty"`
	
	// Detalhes da falha (se aplicável)
	ErrorDetails    *ErrorDetails     `json:"errorDetails,omitempty"`
	
	// Metadados
	DataSources     []string          `json:"dataSources,omitempty"`
	AdditionalData  map[string]interface{} `json:"additionalData,omitempty"`
}

// IdentityResults contém resultados da avaliação de identidade
type IdentityResults struct {
	IdentityVerified bool              `json:"identityVerified"`
	VerificationLevel int               `json:"verificationLevel"`
	VerificationScore float64           `json:"verificationScore"`
	MatchingScore    float64           `json:"matchingScore,omitempty"`
	DataQuality      int               `json:"dataQuality"` // 0-100
	VerifiedAttributes []string        `json:"verifiedAttributes,omitempty"`
	UnverifiedAttributes []string      `json:"unverifiedAttributes,omitempty"`
	VerificationMethod string          `json:"verificationMethod,omitempty"`
	ProviderResults  map[string]interface{} `json:"providerResults,omitempty"`
}

// CreditResults contém resultados da avaliação de crédito
type CreditResults struct {
	CreditScore     int               `json:"creditScore"` // 0-999
	CreditRating    string            `json:"creditRating,omitempty"`
	HasPendingDebts bool              `json:"hasPendingDebts"`
	HasLegalIssues  bool              `json:"hasLegalIssues"`
	IsBlacklisted   bool              `json:"isBlacklisted"`
	CreditCapacity  float64           `json:"creditCapacity,omitempty"`
	DebtToIncomeRatio float64         `json:"debtToIncomeRatio,omitempty"`
	PaymentHistory  string            `json:"paymentHistory,omitempty"`
	ReportDate      time.Time         `json:"reportDate,omitempty"`
	ProviderResponses map[string]adapters.CreditReportResponse `json:"providerResponses,omitempty"`
}

// FraudResults contém resultados da avaliação de fraude
type FraudResults struct {
	FraudDetected   bool              `json:"fraudDetected"`
	FraudProbability float64          `json:"fraudProbability"` // 0-100
	FraudScore      float64           `json:"fraudScore"` // 0-100
	RiskFactors     []frauddetection.RiskFactor `json:"riskFactors,omitempty"`
	AnomalyDetails  []frauddetection.Anomaly    `json:"anomalyDetails,omitempty"`
	FraudVerdict    string            `json:"fraudVerdict"` // APPROVED, REVIEW, REJECTED
	DeviceReputation string           `json:"deviceReputation,omitempty"`
	IPReputation    string            `json:"ipReputation,omitempty"`
	DetectionDetails map[string]interface{} `json:"detectionDetails,omitempty"`
}

// ComplianceResults contém resultados da avaliação de conformidade
type ComplianceResults struct {
	Compliant       bool              `json:"compliant"`
	RegulatoryLists []string          `json:"regulatoryLists,omitempty"`
	Sanctions       []string          `json:"sanctions,omitempty"`
	PEPStatus       bool              `json:"pepStatus"` // Politically Exposed Person
	ComplianceScore int               `json:"complianceScore"` // 0-100
	RiskCategory    string            `json:"riskCategory,omitempty"`
	KYCStatus       string            `json:"kycStatus,omitempty"`
	AMLStatus       string            `json:"amlStatus,omitempty"`
	ComplianceNotes []string          `json:"complianceNotes,omitempty"`
	RuleViolations  []string          `json:"ruleViolations,omitempty"`
}

// RiskResults contém resultados da avaliação de risco
type RiskResults struct {
	RiskScore       float64           `json:"riskScore"` // 0-100
	RiskLevel       string            `json:"riskLevel"`
	RiskFactors     []riskengine.RiskFactor `json:"riskFactors,omitempty"`
	TrustLevel      string            `json:"trustLevel"`
	RecommendedActions []string       `json:"recommendedActions,omitempty"`
	AllowOperation  bool              `json:"allowOperation"`
	RequireAdditionalAuth bool        `json:"requireAdditionalAuth"`
	ContextualData  map[string]interface{} `json:"contextualData,omitempty"`
}

// ErrorDetails contém detalhes de erros ocorridos durante a avaliação
type ErrorDetails struct {
	ErrorCode       string            `json:"errorCode,omitempty"`
	ErrorMessage    string            `json:"errorMessage,omitempty"`
	FailedServices  []string          `json:"failedServices,omitempty"`
	ErrorTimestamp  time.Time         `json:"errorTimestamp,omitempty"`
	Retryable       bool              `json:"retryable"`
	PartialResults  bool              `json:"partialResults"`
	ErrorSource     string            `json:"errorSource,omitempty"`
	ErrorDetails    string            `json:"errorDetails,omitempty"`
}