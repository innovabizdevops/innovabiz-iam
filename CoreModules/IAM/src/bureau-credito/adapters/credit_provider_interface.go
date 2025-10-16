/**
 * @file credit_provider_interface.go
 * @description Define interfaces para adaptadores de provedores de crédito externos
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package adapters

import (
	"context"
	"time"
)

// CreditReportRequest representa uma solicitação de relatório de crédito
type CreditReportRequest struct {
	// Identificadores pessoais
	UserID                string            `json:"userId"`
	DocumentNumber        string            `json:"documentNumber,omitempty"` // CPF, RG, etc.
	DocumentType          string            `json:"documentType,omitempty"`   // CPF, CNPJ, PASSPORT, etc.
	Name                  string            `json:"name,omitempty"`
	
	// Contexto da requisição
	TenantID              string            `json:"tenantId"`
	RequestReason         string            `json:"requestReason"` // AUTH, TRANSACTION, ONBOARDING, etc.
	RequestPriority       RequestPriority   `json:"requestPriority"`
	
	// Metadados para avaliação de risco
	TransactionAmount     float64           `json:"transactionAmount,omitempty"`
	Currency              string            `json:"currency,omitempty"`
	DeviceInfo            DeviceInfo        `json:"deviceInfo,omitempty"`
	GeoLocation           GeoLocation       `json:"geoLocation,omitempty"`
	IPAddress             string            `json:"ipAddress,omitempty"`
	
	// Dados adicionais específicos por provedor
	AdditionalAttributes  map[string]string `json:"additionalAttributes,omitempty"`
	
	// Controle de cache
	ForceRefresh          bool              `json:"forceRefresh"`
	MaxCacheAge           time.Duration     `json:"maxCacheAge"`
}

// DeviceInfo contém informações sobre o dispositivo do usuário
type DeviceInfo struct {
	ID            string `json:"id,omitempty"`
	Type          string `json:"type,omitempty"` // DESKTOP, MOBILE, TABLET, AR, VR
	OS            string `json:"os,omitempty"`
	Browser       string `json:"browser,omitempty"`
	IsTrusted     bool   `json:"isTrusted"`
	IsMobile      bool   `json:"isMobile"`
	Fingerprint   string `json:"fingerprint,omitempty"`
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

// RequestPriority define a prioridade da solicitação
type RequestPriority string

const (
	PriorityLow    RequestPriority = "LOW"
	PriorityMedium RequestPriority = "MEDIUM"
	PriorityHigh   RequestPriority = "HIGH"
	PriorityCritical RequestPriority = "CRITICAL"
)

// CreditReportResponse representa a resposta de um relatório de crédito
type CreditReportResponse struct {
	// Identificadores
	RequestID        string            `json:"requestId"`
	ProviderName     string            `json:"providerName"`
	ReportDate       time.Time         `json:"reportDate"`
	
	// Pontuações e avaliação
	CreditScore      int               `json:"creditScore"`                 // 0-999
	RiskScore        int               `json:"riskScore,omitempty"`         // 0-100
	TrustLevel       TrustLevel        `json:"trustLevel"`                  // Nível de confiança
	RiskAssessment   RiskAssessment    `json:"riskAssessment"`              // Avaliação de risco
	
	// Detalhes do relatório
	HasPendingDebts  bool              `json:"hasPendingDebts"`             // Se possui dívidas pendentes
	HasLegalIssues   bool              `json:"hasLegalIssues"`              // Se possui problemas legais
	IsBlacklisted    bool              `json:"isBlacklisted"`               // Se está na lista negra
	
	// Dados detalhados (específicos por provedor)
	DetailedReport   map[string]interface{} `json:"detailedReport,omitempty"`
	
	// Metadados
	IsFromCache      bool              `json:"isFromCache"`
	CacheTimestamp   *time.Time        `json:"cacheTimestamp,omitempty"`
	ProcessingTimeMs int64             `json:"processingTimeMs"`
	
	// Informações de erros
	Error            *CreditError      `json:"error,omitempty"`
}

// TrustLevel define o nível de confiança
type TrustLevel string

const (
	TrustLevelUnknown TrustLevel = "UNKNOWN"
	TrustLevelVeryLow TrustLevel = "VERY_LOW"
	TrustLevelLow     TrustLevel = "LOW"
	TrustLevelMedium  TrustLevel = "MEDIUM"
	TrustLevelHigh    TrustLevel = "HIGH"
	TrustLevelVeryHigh TrustLevel = "VERY_HIGH"
)

// RiskAssessment define a avaliação de risco
type RiskAssessment string

const (
	RiskUnknown   RiskAssessment = "UNKNOWN"
	RiskVeryHigh  RiskAssessment = "VERY_HIGH"
	RiskHigh      RiskAssessment = "HIGH"
	RiskMedium    RiskAssessment = "MEDIUM"
	RiskLow       RiskAssessment = "LOW"
	RiskVeryLow   RiskAssessment = "VERY_LOW"
)

// CreditError representa um erro na obtenção do relatório de crédito
type CreditError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	Retryable   bool   `json:"retryable"`
}

// CreditProviderConfig define configurações específicas para um provedor
type CreditProviderConfig struct {
	ProviderID        string            `json:"providerId"`
	APIKey            string            `json:"apiKey,omitempty"`
	APISecret         string            `json:"apiSecret,omitempty"`
	BaseURL           string            `json:"baseURL,omitempty"`
	TimeoutSeconds    int               `json:"timeoutSeconds"`
	CacheTTLSeconds   int               `json:"cacheTTLSeconds"`
	RetryCount        int               `json:"retryCount"`
	RetryInterval     time.Duration     `json:"retryInterval"`
	AdditionalConfig  map[string]string `json:"additionalConfig,omitempty"`
}

// CreditProvider define a interface para adaptadores de provedores de crédito
type CreditProvider interface {
	// GetCreditReport obtém um relatório de crédito para o usuário especificado
	GetCreditReport(ctx context.Context, request CreditReportRequest) (*CreditReportResponse, error)
	
	// BatchGetCreditReports obtém relatórios de crédito em lote para múltiplos usuários
	BatchGetCreditReports(ctx context.Context, requests []CreditReportRequest) ([]*CreditReportResponse, error)
	
	// GetProviderInfo retorna informações sobre o provedor
	GetProviderInfo() ProviderInfo
	
	// Initialize inicializa o provedor com configurações específicas
	Initialize(config CreditProviderConfig) error
	
	// IsHealthy verifica se o provedor está operacional
	IsHealthy(ctx context.Context) bool
}

// ProviderInfo contém metadados sobre o provedor de crédito
type ProviderInfo struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	SupportedCountries []string `json:"supportedCountries"`
	SupportedFeatures  []string `json:"supportedFeatures"`
	MaxQPS             int      `json:"maxQPS"` // Consultas máximas por segundo
	Version            string   `json:"version"`
}

// CreditProviderFactory é uma interface para criação dinâmica de provedores
type CreditProviderFactory interface {
	// CreateProvider cria uma nova instância de um provedor de crédito
	CreateProvider(providerType string, config CreditProviderConfig) (CreditProvider, error)
	
	// ListAvailableProviders lista todos os tipos de provedores disponíveis
	ListAvailableProviders() []string
}