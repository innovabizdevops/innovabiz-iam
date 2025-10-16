/**
 * @file risk_engine_types.go
 * @description Define tipos e interfaces para o motor de avaliação de risco
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package riskengine

import (
	"time"
	
	"innovabiz/iam/src/bureau-credito/adapters"
)

// RiskLevel define o nível de risco calculado
type RiskLevel string

const (
	RiskLevelUnknown   RiskLevel = "UNKNOWN"
	RiskLevelNegligible RiskLevel = "NEGLIGIBLE"
	RiskLevelLow       RiskLevel = "LOW"
	RiskLevelMedium    RiskLevel = "MEDIUM"
	RiskLevelHigh      RiskLevel = "HIGH"
	RiskLevelCritical  RiskLevel = "CRITICAL"
)

// RiskFactor representa um fator de risco identificado
type RiskFactor struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`     // Peso relativo deste fator (0-1)
	Score       float64 `json:"score"`      // Score atribuído (0-100)
	Category    string  `json:"category"`   // Ex: IDENTITY, FINANCIAL, DEVICE, LOCATION
	Evidence    string  `json:"evidence,omitempty"`
}

// RiskAssessmentRequest contém os dados necessários para avaliar o risco
type RiskAssessmentRequest struct {
	// Identificadores
	RequestID          string                  `json:"requestId"`
	UserID             string                  `json:"userId"`
	TenantID           string                  `json:"tenantId"`
	
	// Contexto
	OperationType      string                  `json:"operationType"` // LOGIN, TRANSACTION, CHANGE_PROFILE, etc.
	OperationDetails   map[string]interface{}  `json:"operationDetails,omitempty"`
	
	// Informações de contexto para análise
	DeviceInfo         adapters.DeviceInfo     `json:"deviceInfo,omitempty"`
	GeoLocation        adapters.GeoLocation    `json:"geoLocation,omitempty"`
	IPAddress          string                  `json:"ipAddress,omitempty"`
	UserAgent          string                  `json:"userAgent,omitempty"`
	SessionData        map[string]interface{}  `json:"sessionData,omitempty"`
	
	// Informações financeiras
	TransactionAmount  float64                 `json:"transactionAmount,omitempty"`
	Currency           string                  `json:"currency,omitempty"`
	SourceAccount      string                  `json:"sourceAccount,omitempty"`
	DestinationAccount string                  `json:"destinationAccount,omitempty"`
	
	// Configurações para avaliação
	EvaluationProfiles []string                `json:"evaluationProfiles,omitempty"` // Perfis de regras a aplicar
	UseCache           bool                    `json:"useCache"`
	Priority           adapters.RequestPriority `json:"priority"`
}

// RiskAssessmentResponse contém o resultado da avaliação de risco
type RiskAssessmentResponse struct {
	// Identificadores
	AssessmentID       string                 `json:"assessmentId"`
	RequestID          string                 `json:"requestId"`
	UserID             string                 `json:"userId"`
	TenantID           string                 `json:"tenantId"`
	
	// Resultados
	RiskLevel          RiskLevel              `json:"riskLevel"`
	RiskScore          float64                `json:"riskScore"` // 0-100, onde 100 é risco máximo
	ConfidenceScore    float64                `json:"confidenceScore"` // 0-100, confiança na avaliação
	
	// Fatores de risco identificados
	RiskFactors        []RiskFactor           `json:"riskFactors,omitempty"`
	
	// Recomendações
	RecommendedActions []string               `json:"recommendedActions,omitempty"`
	AllowOperation     bool                   `json:"allowOperation"`
	RequireAdditionalAuth bool                `json:"requireAdditionalAuth"`
	
	// Metadados
	AssessmentTime     time.Time              `json:"assessmentTime"`
	ProcessingTimeMs   int64                  `json:"processingTimeMs"`
	DataSources        []string               `json:"dataSources,omitempty"`
	CacheUsed          bool                   `json:"cacheUsed"`
	
	// Informações adicionais específicas por contexto
	ContextualData     map[string]interface{} `json:"contextualData,omitempty"`
}

// RiskRule define uma regra para avaliação de risco
type RiskRule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`        // DEVICE, GEO, BEHAVIOR, IDENTITY, etc.
	Severity        RiskLevel `json:"severity"`        // Severidade se a regra for acionada
	Weight          float64   `json:"weight"`          // Peso da regra no score final (0-1)
	Condition       string    `json:"condition"`       // Expressão de condição (formato específico)
	Action          string    `json:"action"`          // Ação a tomar se condição for verdadeira
	Enabled         bool      `json:"enabled"`
	TenantSpecific  bool      `json:"tenantSpecific"`  // Se a regra é específica para tenant
	ApplicableTenants []string `json:"applicableTenants,omitempty"`
}

// RiskEvaluationProfile define um conjunto de regras aplicáveis
type RiskEvaluationProfile struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	RuleIDs         []string `json:"ruleIds"`         // IDs das regras neste perfil
	TenantID        string   `json:"tenantId"`        // Se específico para tenant
	OperationTypes  []string `json:"operationTypes"`  // LOGIN, PAYMENT, etc.
	Priority        int      `json:"priority"`        // Prioridade em caso de múltiplos perfis (maior = mais prioritário)
	ThresholdScores map[RiskLevel]float64 `json:"thresholdScores"` // Limiares para cada nível
}

// RiskEngine define a interface para o motor de avaliação de risco
type RiskEngine interface {
	// EvaluateRisk avalia o risco com base nos dados fornecidos
	EvaluateRisk(request RiskAssessmentRequest) (*RiskAssessmentResponse, error)
	
	// GetRules retorna as regras configuradas
	GetRules(tenantID string, category string) ([]RiskRule, error)
	
	// GetProfiles retorna os perfis de avaliação configurados
	GetProfiles(tenantID string) ([]RiskEvaluationProfile, error)
	
	// AddRule adiciona uma nova regra
	AddRule(rule RiskRule) error
	
	// UpdateRule atualiza uma regra existente
	UpdateRule(ruleID string, rule RiskRule) error
	
	// DeleteRule remove uma regra
	DeleteRule(ruleID string) error
	
	// CreateProfile cria um novo perfil de avaliação
	CreateProfile(profile RiskEvaluationProfile) error
	
	// UpdateProfile atualiza um perfil existente
	UpdateProfile(profileID string, profile RiskEvaluationProfile) error
	
	// DeleteProfile remove um perfil
	DeleteProfile(profileID string) error
}