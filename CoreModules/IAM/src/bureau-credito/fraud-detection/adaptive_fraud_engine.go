/**
 * @file adaptive_fraud_engine.go
 * @description Implementação do motor adaptativo de detecção de fraudes
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package frauddetection

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/bureau-credito/risk-engine"
)

// AdaptiveFraudEngine implementa a interface FraudDetectionEngine
type AdaptiveFraudEngine struct {
	// Dependências
	userProfileRepo UserProfileRepository
	ruleEngine      riskengine.RiskEngine
	ruleRepo        FraudRuleRepository
	mlProcessor     MLProcessor
	geoService      GeoIPService
	behaviorAnalyzer BehaviorAnalyzer
	eventBus        EventBus
	
	// Cache e estado
	profileCache    *sync.Map
	
	// Configuração
	config          FraudEngineConfig
}

// FraudEngineConfig contém configurações do motor de fraudes
type FraudEngineConfig struct {
	UseML                  bool    // Usar modelos de ML
	EnableBehaviorAnalysis bool    // Analisar padrões comportamentais
	MinConfidenceThreshold float64 // Confiança mínima para decisão
	EnableRealTimeUpdates  bool    // Atualizar perfis em tempo real
	VelocityCheckPeriodMin int     // Período para verificação de velocidade (minutos)
	HighRiskThreshold      float64 // Limiar para classificar risco alto (0-100)
	MediumRiskThreshold    float64 // Limiar para classificar risco médio (0-100)
	LowRiskThreshold       float64 // Limiar para classificar risco baixo (0-100)
	MaxCacheEntries        int     // Número máximo de entradas em cache
	CacheTTLMinutes        int     // Tempo de vida do cache em minutos
}

// NewAdaptiveFraudEngine cria uma nova instância do motor de fraudes
func NewAdaptiveFraudEngine(
	userProfileRepo UserProfileRepository,
	ruleEngine riskengine.RiskEngine,
	ruleRepo FraudRuleRepository,
	mlProcessor MLProcessor,
	geoService GeoIPService,
	behaviorAnalyzer BehaviorAnalyzer,
	eventBus EventBus,
	config FraudEngineConfig,
) *AdaptiveFraudEngine {
	return &AdaptiveFraudEngine{
		userProfileRepo:  userProfileRepo,
		ruleEngine:       ruleEngine,
		ruleRepo:         ruleRepo,
		mlProcessor:      mlProcessor,
		geoService:       geoService,
		behaviorAnalyzer: behaviorAnalyzer,
		eventBus:         eventBus,
		profileCache:     &sync.Map{},
		config:           config,
	}
}

// DetectFraud analisa os dados e retorna uma avaliação de fraude
func (e *AdaptiveFraudEngine) DetectFraud(ctx context.Context, request FraudDetectionRequest) (*FraudDetectionResponse, error) {
	startTime := time.Now()
	
	// Criar ID de resposta único
	responseID := uuid.New().String()
	
	// Preparar resposta padrão
	response := &FraudDetectionResponse{
		ResponseID:   responseID,
		RequestID:    request.RequestID,
		Timestamp:    time.Now(),
		RiskFactors:  []RiskFactor{},
		AnomalyDetails: []Anomaly{},
	}
	
	// Verificar parâmetros obrigatórios
	if request.UserID == "" || request.TenantID == "" {
		return nil, fmt.Errorf("userID e tenantID são obrigatórios")
	}
	
	// Enriquecer dados com informações de geolocalização
	if e.geoService != nil {
		geoInfo, err := e.geoService.EnrichIPData(ctx, request.NetworkInfo.IPAddress)
		if err == nil && geoInfo != nil {
			// Atualizar informações de rede com dados de geolocalização
			request.NetworkInfo.ISP = geoInfo.ISP
			request.NetworkInfo.ASNumber = geoInfo.ASNumber
			
			// Atualizar localização se não estiver definida na requisição
			if request.NetworkInfo.GeoLocation.Country == "" {
				request.NetworkInfo.GeoLocation = *geoInfo.Location
			}
		}
	}
	
	// Buscar perfil do usuário
	userProfile, err := e.getUserProfile(ctx, request.UserID, request.TenantID)
	if err != nil {
		// Se não conseguir buscar o perfil, criar um novo vazio
		userProfile = &UserRiskProfile{
			UserID:      request.UserID,
			TenantID:    request.TenantID,
			RiskScore:   50, // Score neutro para início
			RiskLevel:   "MEDIUM",
			LastUpdated: time.Now(),
		}
	}
	
	// 1. Executar verificações baseadas em regras
	ruleResults, err := e.evaluateRules(ctx, request)
	if err != nil {
		// Log do erro, mas continuar com outras verificações
		// Em produção, enviar telemetria
	}
	
	// 2. Analisar padrões de comportamento
	behaviorResults, err := e.analyzeBehavior(ctx, request, userProfile)
	if err != nil {
		// Log do erro, mas continuar com outras verificações
	}
	
	// 3. Aplicar modelos de ML para detecção de anomalias
	mlResults, err := e.applyMLModels(ctx, request, userProfile)
	if err != nil {
		// Log do erro, mas continuar com outras verificações
	}
	
	// 4. Verificar velocidade e padrão de transações
	velocityResults, err := e.checkVelocity(ctx, request)
	if err != nil {
		// Log do erro, mas continuar com outras verificações
	}
	
	// 5. Consolidar resultados e calcular score final
	e.consolidateResults(response, ruleResults, behaviorResults, mlResults, velocityResults)
	
	// 6. Determinar o veredicto com base no score
	response.FraudVerdict = e.determineVerdict(response.FraudScore)
	
	// 7. Definir ações recomendadas
	response.RecommendedAction = e.recommendAction(response.FraudScore, response.FraudVerdict, response.RiskFactors)
	
	// 8. Registrar tempo de processamento
	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	
	// 9. Registrar versões de modelo e regras
	response.ModelVersion = "1.0.0" // Em produção, buscar versão real
	response.RulesVersion = "1.0.0" // Em produção, buscar versão real
	
	// 10. Atualizar perfil do usuário de forma assíncrona
	if e.config.EnableRealTimeUpdates {
		go e.updateUserProfileAsync(request, response)
	}
	
	// 11. Publicar evento de detecção de fraude para outros sistemas
	if e.eventBus != nil {
		e.publishFraudDetectionEvent(request, response)
	}
	
	return response, nil
}

// BatchDetectFraud processa múltiplas solicitações de detecção em lote
func (e *AdaptiveFraudEngine) BatchDetectFraud(ctx context.Context, requests []FraudDetectionRequest) ([]*FraudDetectionResponse, error) {
	responses := make([]*FraudDetectionResponse, len(requests))
	errors := make([]error, len(requests))
	
	var wg sync.WaitGroup
	wg.Add(len(requests))
	
	// Processar solicitações em paralelo
	for i, req := range requests {
		go func(idx int, request FraudDetectionRequest) {
			defer wg.Done()
			
			resp, err := e.DetectFraud(ctx, request)
			responses[idx] = resp
			errors[idx] = err
		}(i, req)
	}
	
	wg.Wait()
	
	// Verificar se houve erros
	for _, err := range errors {
		if err != nil {
			return responses, fmt.Errorf("erro em pelo menos uma solicitação do lote: %w", err)
		}
	}
	
	return responses, nil
}

// UpdateUserProfile atualiza o perfil de comportamento do usuário com novos dados
func (e *AdaptiveFraudEngine) UpdateUserProfile(ctx context.Context, userID, tenantID string, data FraudDetectionRequest) error {
	// Buscar perfil atual do usuário
	profile, err := e.getUserProfile(ctx, userID, tenantID)
	if err != nil {
		// Se não existe, criar um novo
		profile = &UserRiskProfile{
			UserID:      userID,
			TenantID:    tenantID,
			LastUpdated: time.Now(),
		}
	}
	
	// Atualizar dados do perfil com base nas novas informações
	e.updateProfileData(profile, data)
	
	// Salvar perfil atualizado
	if err := e.userProfileRepo.SaveUserProfile(ctx, *profile); err != nil {
		return fmt.Errorf("erro ao salvar perfil do usuário: %w", err)
	}
	
	// Atualizar cache
	cacheKey := fmt.Sprintf("%s:%s", userID, tenantID)
	e.profileCache.Store(cacheKey, profile)
	
	return nil
}

// GetUserRiskProfile obtém o perfil de risco atual de um usuário
func (e *AdaptiveFraudEngine) GetUserRiskProfile(ctx context.Context, userID, tenantID string) (*UserRiskProfile, error) {
	return e.getUserProfile(ctx, userID, tenantID)
}

// Métodos auxiliares privados

// getUserProfile busca o perfil do usuário no cache ou no repositório
func (e *AdaptiveFraudEngine) getUserProfile(ctx context.Context, userID, tenantID string) (*UserRiskProfile, error) {
	cacheKey := fmt.Sprintf("%s:%s", userID, tenantID)
	
	// Tentar buscar do cache primeiro
	if cachedProfile, found := e.profileCache.Load(cacheKey); found {
		if profile, ok := cachedProfile.(*UserRiskProfile); ok {
			return profile, nil
		}
	}
	
	// Se não estiver em cache, buscar do repositório
	profile, err := e.userProfileRepo.GetUserProfile(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	
	// Armazenar em cache para uso futuro
	e.profileCache.Store(cacheKey, profile)
	
	return profile, nil
}

// evaluateRules aplica regras de negócio para detecção de fraude
func (e *AdaptiveFraudEngine) evaluateRules(ctx context.Context, request FraudDetectionRequest) ([]RiskFactor, error) {
	// Obter regras aplicáveis
	rules, err := e.ruleRepo.GetRules(ctx, request.TenantID, request.OperationType)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar regras: %w", err)
	}
	
	riskFactors := []RiskFactor{}
	
	// Avaliar cada regra
	for _, rule := range rules {
		// Aplicar condição da regra aos dados da requisição
		matched, confidence, err := e.evaluateRuleCondition(rule, request)
		if err != nil {
			// Log de erro e continuar com próxima regra
			continue
		}
		
		// Se regra for acionada, adicionar fator de risco
		if matched {
			riskFactor := RiskFactor{
				Code:        rule.Code,
				Name:        rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				Score:       rule.Score,
				Confidence:  confidence,
				Category:    rule.Category,
			}
			
			riskFactors = append(riskFactors, riskFactor)
		}
	}
	
	return riskFactors, nil
}

// evaluateRuleCondition avalia se uma regra é acionada para os dados fornecidos
func (e *AdaptiveFraudEngine) evaluateRuleCondition(rule FraudRule, request FraudDetectionRequest) (bool, float64, error) {
	// Implementação simplificada - em produção, usar um motor de regras real
	// que avalia expressões condicionais dinâmicas
	
	// Exemplo: regra de detecção de proxy/VPN/Tor
	if rule.Code == "ANONYMIZER_DETECTED" && 
	   (request.NetworkInfo.ProxyDetected || 
	    request.NetworkInfo.VPNDetected || 
		request.NetworkInfo.TorDetected) {
		return true, 90.0, nil
	}
	
	// Exemplo: regra de dispositivo emulado
	if rule.Code == "EMULATOR_DETECTED" && request.DeviceInfo.Emulator {
		return true, 80.0, nil
	}
	
	// Exemplo: regra de dispositivo root/jailbreak
	if rule.Code == "JAILBROKEN_DEVICE" && request.DeviceInfo.Jailbroken {
		return true, 85.0, nil
	}
	
	// Regra não acionada
	return false, 0.0, nil
}

// analyzeBehavior analisa padrões comportamentais do usuário
func (e *AdaptiveFraudEngine) analyzeBehavior(
	ctx context.Context,
	request FraudDetectionRequest,
	profile *UserRiskProfile,
) ([]Anomaly, error) {
	if !e.config.EnableBehaviorAnalysis || e.behaviorAnalyzer == nil {
		return nil, nil
	}
	
	// Delegar análise comportamental para o analisador especializado
	return e.behaviorAnalyzer.AnalyzeBehavior(ctx, request, profile)
}

// applyMLModels aplica modelos de machine learning para detecção de anomalias
func (e *AdaptiveFraudEngine) applyMLModels(
	ctx context.Context, 
	request FraudDetectionRequest,
	profile *UserRiskProfile,
) ([]Anomaly, error) {
	if !e.config.UseML || e.mlProcessor == nil {
		return nil, nil
	}
	
	// Delegar processamento de ML para processador especializado
	return e.mlProcessor.ProcessRequest(ctx, request, profile)
}

// checkVelocity verifica padrões de velocidade e frequência de transações
func (e *AdaptiveFraudEngine) checkVelocity(ctx context.Context, request FraudDetectionRequest) ([]Anomaly, error) {
	// Implementação simplificada - em produção, implementar análise
	// de velocidade de transações, frequência de logins, etc.
	return nil, nil
}

// consolidateResults consolida resultados de diferentes análises
func (e *AdaptiveFraudEngine) consolidateResults(
	response *FraudDetectionResponse,
	ruleResults []RiskFactor,
	behaviorResults []Anomaly,
	mlResults []Anomaly,
	velocityResults []Anomaly,
) {
	// 1. Adicionar fatores de risco das regras
	response.RiskFactors = append(response.RiskFactors, ruleResults...)
	
	// 2. Adicionar anomalias comportamentais
	response.AnomalyDetails = append(response.AnomalyDetails, behaviorResults...)
	
	// 3. Adicionar anomalias detectadas por ML
	response.AnomalyDetails = append(response.AnomalyDetails, mlResults...)
	
	// 4. Adicionar anomalias de velocidade
	response.AnomalyDetails = append(response.AnomalyDetails, velocityResults...)
	
	// 5. Calcular score de fraude global
	var totalScore float64
	var totalWeight float64
	
	// Considerar fatores de risco
	for _, factor := range response.RiskFactors {
		weight := e.getSeverityWeight(factor.Severity)
		totalScore += factor.Score * weight * (factor.Confidence / 100)
		totalWeight += weight
	}
	
	// Considerar anomalias
	for _, anomaly := range response.AnomalyDetails {
		weight := e.getSeverityWeight(anomaly.Severity)
		// Usar valor padrão de 70 para anomalias sem score específico
		anomalyScore := 70.0
		if score, ok := anomaly.Details["score"]; ok {
			if scoreVal, ok := score.(float64); ok {
				anomalyScore = scoreVal
			}
		}
		
		totalScore += anomalyScore * weight * (anomaly.Confidence / 100)
		totalWeight += weight
	}
	
	// Calcular score final
	if totalWeight > 0 {
		response.FraudScore = math.Min(100, totalScore/totalWeight)
	} else {
		// Se nenhum fator de risco ou anomalia, score baixo
		response.FraudScore = 10
	}
	
	// Calcular confiança na avaliação
	response.FraudConfidence = e.calculateConfidence(response)
}

// getSeverityWeight retorna o peso para um nível de severidade
func (e *AdaptiveFraudEngine) getSeverityWeight(severity string) float64 {
	switch severity {
	case "LOW":
		return 1.0
	case "MEDIUM":
		return 2.0
	case "HIGH":
		return 3.0
	case "CRITICAL":
		return 5.0
	default:
		return 1.0
	}
}

// calculateConfidence calcula o nível de confiança na avaliação
func (e *AdaptiveFraudEngine) calculateConfidence(response *FraudDetectionResponse) float64 {
	// Implementação simplificada - em produção, usar algoritmo mais sofisticado
	
	// Se não houver fatores de risco ou anomalias, confiança baixa
	if len(response.RiskFactors) == 0 && len(response.AnomalyDetails) == 0 {
		return 30.0
	}
	
	// Calcular média ponderada das confianças individuais
	var totalConfidence float64
	var totalWeight float64
	
	for _, factor := range response.RiskFactors {
		weight := e.getSeverityWeight(factor.Severity)
		totalConfidence += factor.Confidence * weight
		totalWeight += weight
	}
	
	for _, anomaly := range response.AnomalyDetails {
		weight := e.getSeverityWeight(anomaly.Severity)
		totalConfidence += anomaly.Confidence * weight
		totalWeight += weight
	}
	
	if totalWeight > 0 {
		return math.Min(100, totalConfidence/totalWeight)
	}
	
	return 50.0 // Valor neutro
}

// determineVerdict determina o veredicto final com base no score de fraude
func (e *AdaptiveFraudEngine) determineVerdict(fraudScore float64) string {
	switch {
	case fraudScore >= e.config.HighRiskThreshold:
		return "REJECTED"
	case fraudScore >= e.config.MediumRiskThreshold:
		return "REVIEW"
	default:
		return "APPROVED"
	}
}

// recommendAction sugere ações com base no score e veredicto
func (e *AdaptiveFraudEngine) recommendAction(
	score float64,
	verdict string,
	factors []RiskFactor,
) string {
	switch verdict {
	case "REJECTED":
		return "BLOCK"
	case "REVIEW":
		// Se houver fatores críticos, sugerir autenticação adicional
		for _, factor := range factors {
			if factor.Severity == "CRITICAL" || factor.Severity == "HIGH" {
				return "ADDITIONAL_AUTHENTICATION"
			}
		}
		return "RISK_ASSESSMENT"
	default:
		return "ALLOW"
	}
}

// updateUserProfileAsync atualiza o perfil do usuário de forma assíncrona
func (e *AdaptiveFraudEngine) updateUserProfileAsync(request FraudDetectionRequest, response *FraudDetectionResponse) {
	// Usar um novo contexto para operação assíncrona
	ctx := context.Background()
	
	// Atualizar perfil com novos dados
	_ = e.UpdateUserProfile(ctx, request.UserID, request.TenantID, request)
}

// publishFraudDetectionEvent publica um evento de detecção de fraude
func (e *AdaptiveFraudEngine) publishFraudDetectionEvent(request FraudDetectionRequest, response *FraudDetectionResponse) {
	if e.eventBus == nil {
		return
	}
	
	event := FraudDetectionEvent{
		EventType:   "FRAUD_DETECTION",
		Timestamp:   time.Now(),
		UserID:      request.UserID,
		TenantID:    request.TenantID,
		RequestID:   request.RequestID,
		ResponseID:  response.ResponseID,
		FraudScore:  response.FraudScore,
		FraudVerdict: response.FraudVerdict,
		Operation:   request.OperationType,
	}
	
	// Publicar evento de forma assíncrona
	go func() {
		_ = e.eventBus.PublishEvent(context.Background(), event)
	}()
}