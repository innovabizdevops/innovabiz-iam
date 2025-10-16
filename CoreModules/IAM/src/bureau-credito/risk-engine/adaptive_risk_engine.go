/**
 * @file adaptive_risk_engine.go
 * @description Implementação do motor adaptativo de avaliação de risco
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package riskengine

import (
	"errors"
	"fmt"
	"sync"
	"time"
	
	"github.com/google/uuid"
	
	"innovabiz/iam/src/bureau-credito/adapters"
)

// AdaptiveRiskEngine implementa o motor de avaliação de risco com regras adaptativas
type AdaptiveRiskEngine struct {
	rules         map[string]RiskRule
	profiles      map[string]RiskEvaluationProfile
	rulesMutex    sync.RWMutex
	profilesMutex sync.RWMutex
	
	// Dependências
	creditProviderFactory adapters.CreditProviderFactory
	ruleRepository        RuleRepository
	profileRepository     ProfileRepository
	cachingService        CachingService
	
	// Configurações
	config RiskEngineConfig
}

// RiskEngineConfig define configurações para o motor de risco
type RiskEngineConfig struct {
	DefaultMinimumConfidence float64
	EnableCache              bool
	CacheTTLSeconds          int
	DefaultRiskThresholds    map[RiskLevel]float64
	EnableMachineLearning    bool
	TelemetryEnabled         bool
}

// NewAdaptiveRiskEngine cria uma nova instância do motor de avaliação de risco
func NewAdaptiveRiskEngine(
	ruleRepo RuleRepository,
	profileRepo ProfileRepository,
	cachingService CachingService,
	providerFactory adapters.CreditProviderFactory,
	config RiskEngineConfig,
) *AdaptiveRiskEngine {
	
	engine := &AdaptiveRiskEngine{
		rules:                 make(map[string]RiskRule),
		profiles:              make(map[string]RiskEvaluationProfile),
		ruleRepository:        ruleRepo,
		profileRepository:     profileRepo,
		cachingService:        cachingService,
		creditProviderFactory: providerFactory,
		config:                config,
	}
	
	// Carregar regras e perfis do repositório
	engine.loadRules()
	engine.loadProfiles()
	
	return engine
}

// EvaluateRisk avalia o risco com base nos dados fornecidos
func (e *AdaptiveRiskEngine) EvaluateRisk(request RiskAssessmentRequest) (*RiskAssessmentResponse, error) {
	startTime := time.Now()
	
	// Verificar dados obrigatórios
	if request.UserID == "" || request.TenantID == "" {
		return nil, errors.New("userID e tenantID são obrigatórios")
	}
	
	// Preparar resposta
	response := &RiskAssessmentResponse{
		AssessmentID:    uuid.New().String(),
		RequestID:       request.RequestID,
		UserID:          request.UserID,
		TenantID:        request.TenantID,
		AssessmentTime:  time.Now(),
		AllowOperation:  true, // Default permitir, regras podem mudar isto
		RiskFactors:     []RiskFactor{},
		DataSources:     []string{"internal-rules"},
	}
	
	// Verificar cache se habilitado
	if e.config.EnableCache && request.UseCache {
		cachedResponse := e.checkCache(request)
		if cachedResponse != nil {
			cachedResponse.CacheUsed = true
			return cachedResponse, nil
		}
	}
	
	// Obter perfis de avaliação aplicáveis
	profiles := e.getApplicableProfiles(request)
	if len(profiles) == 0 {
		// Se nenhum perfil específico for encontrado, usar perfil padrão
		defaultProfile, err := e.getDefaultProfile()
		if err != nil {
			return nil, fmt.Errorf("erro ao obter perfil padrão: %w", err)
		}
		profiles = append(profiles, defaultProfile)
	}
	
	// Coletar dados de crédito se necessário para avaliação
	creditData, err := e.collectCreditData(request)
	if err != nil {
		// Log do erro, mas continuamos com dados disponíveis
		// Em produção, registrar telemetria do erro
	}
	if creditData != nil {
		response.DataSources = append(response.DataSources, creditData.ProviderName)
	}
	
	// Aplicar regras de todos os perfis aplicáveis
	var totalRiskScore float64
	var totalWeight float64
	
	for _, profile := range profiles {
		profileScore, profileFactors, err := e.evaluateProfile(profile, request, creditData)
		if err != nil {
			return nil, fmt.Errorf("erro ao avaliar perfil %s: %w", profile.ID, err)
		}
		
		// Acumular fatores de risco identificados
		response.RiskFactors = append(response.RiskFactors, profileFactors...)
		
		// Acumular score ponderado
		weight := float64(profile.Priority)
		totalRiskScore += profileScore * weight
		totalWeight += weight
	}
	
	// Calcular score de risco final
	if totalWeight > 0 {
		response.RiskScore = totalRiskScore / totalWeight
	}
	
	// Determinar nível de risco com base nos limiares do perfil com maior prioridade
	highestPriorityProfile := getHighestPriorityProfile(profiles)
	response.RiskLevel = e.determineRiskLevel(response.RiskScore, highestPriorityProfile)
	
	// Determinar ações recomendadas com base no nível de risco
	response.RecommendedActions = e.determineRecommendedActions(response.RiskLevel)
	
	// Definir se requer autenticação adicional
	response.RequireAdditionalAuth = e.requiresAdditionalAuth(response.RiskLevel)
	
	// Determinar se deve bloquear a operação
	response.AllowOperation = e.allowOperation(response.RiskLevel)
	
	// Calcular confiança na avaliação
	response.ConfidenceScore = e.calculateConfidence(request, response)
	
	// Registrar tempo de processamento
	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	
	// Armazenar em cache se habilitado
	if e.config.EnableCache {
		e.storeInCache(request, response)
	}
	
	return response, nil
}

// GetRules retorna as regras configuradas
func (e *AdaptiveRiskEngine) GetRules(tenantID string, category string) ([]RiskRule, error) {
	e.rulesMutex.RLock()
	defer e.rulesMutex.RUnlock()
	
	result := []RiskRule{}
	
	// Filtrar regras por tenant e categoria
	for _, rule := range e.rules {
		// Incluir regras globais ou específicas para este tenant
		if !rule.TenantSpecific || containsString(rule.ApplicableTenants, tenantID) {
			// Filtrar por categoria se especificada
			if category == "" || rule.Category == category {
				result = append(result, rule)
			}
		}
	}
	
	return result, nil
}

// Funções auxiliares

// checkCache verifica se há uma avaliação em cache para a requisição
func (e *AdaptiveRiskEngine) checkCache(request RiskAssessmentRequest) *RiskAssessmentResponse {
	if e.cachingService == nil {
		return nil
	}
	
	cacheKey := fmt.Sprintf("risk:%s:%s:%s", request.UserID, request.TenantID, request.OperationType)
	cachedData, found := e.cachingService.Get(cacheKey)
	if !found {
		return nil
	}
	
	response, ok := cachedData.(*RiskAssessmentResponse)
	if !ok {
		return nil
	}
	
	return response
}

// storeInCache armazena uma avaliação em cache
func (e *AdaptiveRiskEngine) storeInCache(request RiskAssessmentRequest, response *RiskAssessmentResponse) {
	if e.cachingService == nil {
		return
	}
	
	cacheKey := fmt.Sprintf("risk:%s:%s:%s", request.UserID, request.TenantID, request.OperationType)
	e.cachingService.Set(cacheKey, response, time.Duration(e.config.CacheTTLSeconds)*time.Second)
}

// getApplicableProfiles retorna os perfis aplicáveis para a requisição
func (e *AdaptiveRiskEngine) getApplicableProfiles(request RiskAssessmentRequest) []RiskEvaluationProfile {
	e.profilesMutex.RLock()
	defer e.profilesMutex.RUnlock()
	
	applicable := []RiskEvaluationProfile{}
	
	// Se perfis específicos foram solicitados
	if len(request.EvaluationProfiles) > 0 {
		for _, profileID := range request.EvaluationProfiles {
			if profile, exists := e.profiles[profileID]; exists {
				applicable = append(applicable, profile)
			}
		}
		
		if len(applicable) > 0 {
			return applicable
		}
	}
	
	// Caso contrário, encontrar perfis pelo tenant e tipo de operação
	for _, profile := range e.profiles {
		// Verificar se o perfil é aplicável para o tenant
		tenantMatch := profile.TenantID == "" || profile.TenantID == request.TenantID
		
		// Verificar se o tipo de operação está coberto
		opTypeMatch := false
		for _, opType := range profile.OperationTypes {
			if opType == request.OperationType || opType == "*" {
				opTypeMatch = true
				break
			}
		}
		
		if tenantMatch && opTypeMatch {
			applicable = append(applicable, profile)
		}
	}
	
	return applicable
}

// Função utilitária
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}