/**
 * @file fraud_detection_types.go
 * @description Define tipos e interfaces para o sistema de detecção de fraudes
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package frauddetection

import (
	"context"
	"time"
)

// FraudDetectionRequest representa uma solicitação de detecção de fraude
type FraudDetectionRequest struct {
	// Identificadores
	RequestID        string                 `json:"requestId"`
	SessionID        string                 `json:"sessionId,omitempty"`
	UserID           string                 `json:"userId"`
	TenantID         string                 `json:"tenantId"`
	
	// Contexto da operação
	OperationType    string                 `json:"operationType"` // LOGIN, PAYMENT, PROFILE_UPDATE
	Timestamp        time.Time              `json:"timestamp"`
	
	// Dados do dispositivo e rede
	DeviceInfo       DeviceInfo             `json:"deviceInfo"`
	NetworkInfo      NetworkInfo            `json:"networkInfo"`
	
	// Dados comportamentais
	BehavioralData   BehavioralData         `json:"behavioralData,omitempty"`
	
	// Dados da transação (quando aplicável)
	TransactionData  *TransactionData       `json:"transactionData,omitempty"`
	
	// Dados biométricos (quando aplicável)
	BiometricData    *BiometricData         `json:"biometricData,omitempty"`
	
	// Contexto adicional
	CustomAttributes map[string]interface{} `json:"customAttributes,omitempty"`
}

// DeviceInfo contém informações sobre o dispositivo do usuário
type DeviceInfo struct {
	DeviceID         string `json:"deviceId"`
	DeviceType       string `json:"deviceType"`         // MOBILE, DESKTOP, TABLET, IOT, AR, VR
	OS               string `json:"os,omitempty"`
	OSVersion        string `json:"osVersion,omitempty"`
	Browser          string `json:"browser,omitempty"`
	BrowserVersion   string `json:"browserVersion,omitempty"`
	ScreenResolution string `json:"screenResolution,omitempty"`
	DeviceModel      string `json:"deviceModel,omitempty"`
	DeviceBrand      string `json:"deviceBrand,omitempty"`
	Jailbroken       bool   `json:"jailbroken"`
	Emulator         bool   `json:"emulator"`
	DeviceLanguage   string `json:"deviceLanguage,omitempty"`
	TimeZone         string `json:"timeZone,omitempty"`
}

// NetworkInfo contém informações sobre a rede do usuário
type NetworkInfo struct {
	IPAddress          string    `json:"ipAddress"`
	ISP                string    `json:"isp,omitempty"`
	ConnectionType     string    `json:"connectionType,omitempty"` // WIFI, CELLULAR, ETHERNET
	HostName           string    `json:"hostName,omitempty"`
	ASNumber           string    `json:"asNumber,omitempty"`       // Autonomous System Number
	ProxyDetected      bool      `json:"proxyDetected"`
	VPNDetected        bool      `json:"vpnDetected"`
	TorDetected        bool      `json:"torDetected"`
	GeoLocation        GeoLocation `json:"geoLocation"`
	AnonymizationScore int       `json:"anonymizationScore"`      // 0-100, maior = mais anônimo
}

// GeoLocation contém informações de localização geográfica
type GeoLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Country     string  `json:"country,omitempty"`
	Region      string  `json:"region,omitempty"`
	City        string  `json:"city,omitempty"`
	PostalCode  string  `json:"postalCode,omitempty"`
	Accuracy    float64 `json:"accuracy,omitempty"` // precisão em metros
}

// BehavioralData contém informações sobre o comportamento do usuário
type BehavioralData struct {
	TypingPattern      string        `json:"typingPattern,omitempty"`
	MouseMovements     string        `json:"mouseMovements,omitempty"`
	NavigationFlow     []string      `json:"navigationFlow,omitempty"`
	TimeSpent          time.Duration `json:"timeSpent,omitempty"`
	DwellTime          time.Duration `json:"dwellTime,omitempty"`
	InteractionCount   int           `json:"interactionCount,omitempty"`
	UnusualTimeOfDay   bool          `json:"unusualTimeOfDay"`
	CopyPasteDetected  bool          `json:"copyPasteDetected"`
	AutoFillDetected   bool          `json:"autoFillDetected"`
	TypingSpeed        int           `json:"typingSpeed,omitempty"` // caracteres por minuto
	TypingErrorRate    float64       `json:"typingErrorRate,omitempty"`
	HesitationCount    int           `json:"hesitationCount,omitempty"`
}

// TransactionData contém detalhes de uma transação financeira
type TransactionData struct {
	TransactionID      string    `json:"transactionId"`
	Amount             float64   `json:"amount"`
	Currency           string    `json:"currency"`
	Source             string    `json:"source,omitempty"`
	Destination        string    `json:"destination,omitempty"`
	PaymentMethod      string    `json:"paymentMethod,omitempty"`
	PaymentType        string    `json:"paymentType,omitempty"`
	MerchantID         string    `json:"merchantId,omitempty"`
	MerchantName       string    `json:"merchantName,omitempty"`
	MerchantCategory   string    `json:"merchantCategory,omitempty"`
	TransactionTime    time.Time `json:"transactionTime"`
	ItemCount          int       `json:"itemCount,omitempty"`
	IsRecurring        bool      `json:"isRecurring"`
	FrequencyInDays    int       `json:"frequencyInDays,omitempty"`
	BillingAddress     string    `json:"billingAddress,omitempty"`
	ShippingAddress    string    `json:"shippingAddress,omitempty"`
	FirstTimeWithMerchant bool   `json:"firstTimeWithMerchant"`
}

// BiometricData contém informações biométricas
type BiometricData struct {
	FaceID        string  `json:"faceId,omitempty"`
	FaceScore     float64 `json:"faceScore,omitempty"`    // 0-1, maior = mais confiança
	FingerprintID string  `json:"fingerprintId,omitempty"`
	VoiceprintID  string  `json:"voiceprintId,omitempty"`
	VoiceScore    float64 `json:"voiceScore,omitempty"`   // 0-1, maior = mais confiança
	LivenessScore float64 `json:"livenessScore,omitempty"` // 0-1, maior = mais confiança
}

// FraudDetectionResponse representa o resultado da detecção de fraude
type FraudDetectionResponse struct {
	// Identificação
	ResponseID      string    `json:"responseId"`
	RequestID       string    `json:"requestId"`
	Timestamp       time.Time `json:"timestamp"`
	
	// Resultado da análise
	FraudScore      float64   `json:"fraudScore"`      // 0-100, maior = mais suspeito
	FraudConfidence float64   `json:"fraudConfidence"` // 0-100, confiança na detecção
	FraudVerdict    string    `json:"fraudVerdict"`    // APPROVED, REVIEW, REJECTED
	
	// Detalhes de análise
	RiskFactors     []RiskFactor `json:"riskFactors,omitempty"`
	AnomalyDetails  []Anomaly    `json:"anomalyDetails,omitempty"`
	
	// Ações recomendadas
	RecommendedAction string    `json:"recommendedAction"`
	RequiredChecks   []string  `json:"requiredChecks,omitempty"`
	
	// Metadados
	ProcessingTimeMs int64     `json:"processingTimeMs"`
	ModelVersion     string    `json:"modelVersion,omitempty"`
	RulesVersion     string    `json:"rulesVersion,omitempty"`
}

// RiskFactor representa um fator de risco identificado
type RiskFactor struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Score       float64 `json:"score"`    // Impacto no score de fraude (0-100)
	Confidence  float64 `json:"confidence"` // 0-100
	Category    string  `json:"category"`
}

// Anomaly representa uma anomalia detectada
type Anomaly struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Confidence  float64                `json:"confidence"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// FraudDetectionEngine define a interface do motor de detecção de fraudes
type FraudDetectionEngine interface {
	// DetectFraud analisa os dados e retorna uma avaliação de fraude
	DetectFraud(ctx context.Context, request FraudDetectionRequest) (*FraudDetectionResponse, error)
	
	// BatchDetectFraud processa múltiplas solicitações de detecção em lote
	BatchDetectFraud(ctx context.Context, requests []FraudDetectionRequest) ([]*FraudDetectionResponse, error)
	
	// UpdateUserProfile atualiza o perfil de comportamento do usuário com novos dados
	UpdateUserProfile(ctx context.Context, userID, tenantID string, data FraudDetectionRequest) error
	
	// GetUserRiskProfile obtém o perfil de risco atual de um usuário
	GetUserRiskProfile(ctx context.Context, userID, tenantID string) (*UserRiskProfile, error)
}

// UserRiskProfile representa o perfil de risco de um usuário
type UserRiskProfile struct {
	UserID           string    `json:"userId"`
	TenantID         string    `json:"tenantId"`
	RiskScore        float64   `json:"riskScore"`        // 0-100
	RiskLevel        string    `json:"riskLevel"`        // LOW, MEDIUM, HIGH
	LastUpdated      time.Time `json:"lastUpdated"`
	CommonLocations  []GeoLocation `json:"commonLocations,omitempty"`
	CommonDevices    []string  `json:"commonDevices,omitempty"`
	UsualHours       []int     `json:"usualHours,omitempty"`      // Horas do dia mais comuns (0-23)
	TypicalBehavior  map[string]interface{} `json:"typicalBehavior,omitempty"`
	AnomalyHistory   []Anomaly `json:"anomalyHistory,omitempty"`
	LastIncidents    []string  `json:"lastIncidents,omitempty"`
	TrustFactors     map[string]float64 `json:"trustFactors,omitempty"`
}