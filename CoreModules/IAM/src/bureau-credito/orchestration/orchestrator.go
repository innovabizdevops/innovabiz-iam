/**
 * @file orchestrator.go
 * @description Orquestrador principal para serviços do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package orchestration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	
	"github.com/google/uuid"
	
	"innovabiz/iam/src/bureau-credito/adapters"
	"innovabiz/iam/src/bureau-credito/fraud-detection"
	"innovabiz/iam/src/bureau-credito/orchestration/models"
	"innovabiz/iam/src/bureau-credito/risk-engine"
)

// BureauOrchestrator coordena a execução de múltiplos serviços de avaliação
type BureauOrchestrator struct {
	// Provedores de serviços
	creditProviderFactory adapters.CreditProviderFactory
	creditProviders       map[string]adapters.CreditProvider
	riskEngine            riskengine.RiskEngine
	fraudEngine           frauddetection.FraudDetectionEngine
	complianceService     ComplianceService
	identityService       IdentityService
	
	// Outros componentes
	eventBus             EventBus
	cache                CacheService
	
	// Configurações
	config               OrchestratorConfig
	
	// Estado do orquestrador
	providersMutex       sync.RWMutex
	activeAssessments    map[string]*models.AssessmentRequest
	assessmentsMutex     sync.RWMutex
}

// OrchestratorConfig define as configurações para o orquestrador
type OrchestratorConfig struct {
	DefaultTimeout            time.Duration
	MaxConcurrentAssessments  int
	EnableCache               bool
	CacheTTLSeconds           int
	DefaultCreditProviders    []string
	DefaultComplianceRules    []string
	TelemetryEnabled          bool
	MaxRetries                int
	RetryDelayMs              int
}

// EventBus define a interface para publicação de eventos
type EventBus interface {
	// PublishEvent publica um evento para outros serviços
	PublishEvent(ctx context.Context, event interface{}) error
}

// CacheService define a interface para cache
type CacheService interface {
	// Get obtém um valor do cache
	Get(key string) (interface{}, bool)
	
	// Set armazena um valor no cache
	Set(key string, value interface{}, ttl time.Duration)
	
	// Delete remove um valor do cache
	Delete(key string)
}

// ComplianceService define a interface para serviços de conformidade
type ComplianceService interface {
	// CheckCompliance verifica conformidade com regulamentações
	CheckCompliance(ctx context.Context, request *models.AssessmentRequest) (*models.ComplianceResults, error)
	
	// GetComplianceRules retorna as regras de conformidade disponíveis
	GetComplianceRules(ctx context.Context, tenantID string) ([]string, error)
}

// IdentityService define a interface para serviços de identidade
type IdentityService interface {
	// VerifyIdentity verifica a identidade do usuário
	VerifyIdentity(ctx context.Context, request *models.AssessmentRequest) (*models.IdentityResults, error)
	
	// GetIdentityProviders retorna os provedores de identidade disponíveis
	GetIdentityProviders(ctx context.Context, tenantID string) ([]string, error)
}

// NewBureauOrchestrator cria uma nova instância do orquestrador
func NewBureauOrchestrator(
	creditProviderFactory adapters.CreditProviderFactory,
	riskEngine riskengine.RiskEngine,
	fraudEngine frauddetection.FraudDetectionEngine,
	complianceService ComplianceService,
	identityService IdentityService,
	eventBus EventBus,
	cache CacheService,
	config OrchestratorConfig,
) *BureauOrchestrator {
	return &BureauOrchestrator{
		creditProviderFactory: creditProviderFactory,
		creditProviders:       make(map[string]adapters.CreditProvider),
		riskEngine:            riskEngine,
		fraudEngine:           fraudEngine,
		complianceService:     complianceService,
		identityService:       identityService,
		eventBus:              eventBus,
		cache:                 cache,
		config:                config,
		activeAssessments:     make(map[string]*models.AssessmentRequest),
	}
}

// RequestAssessment inicia uma nova avaliação orquestrada
func (o *BureauOrchestrator) RequestAssessment(
	ctx context.Context,
	request models.AssessmentRequest,
) (*models.AssessmentResponse, error) {
	startTime := time.Now()
	
	// Gerar um ID de solicitação se não for fornecido
	if request.RequestID == "" {
		request.RequestID = uuid.New().String()
	}
	
	// Definir o timestamp de solicitação
	if request.RequestTimestamp.IsZero() {
		request.RequestTimestamp = time.Now()
	}
	
	// Verificar parâmetros obrigatórios
	if request.UserID == "" || request.TenantID == "" {
		return nil, errors.New("userID e tenantID são obrigatórios")
	}
	
	// Verificar se os tipos de avaliação foram especificados
	if len(request.AssessmentTypes) == 0 {
		return nil, errors.New("pelo menos um tipo de avaliação deve ser especificado")
	}
	
	// Definir timeout padrão se não especificado
	if request.Timeout == 0 {
		request.Timeout = o.config.DefaultTimeout
	}
	
	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(ctx, request.Timeout)
	defer cancel()
	
	// Registrar solicitação ativa
	o.registerActiveAssessment(request.RequestID, &request)
	defer o.unregisterActiveAssessment(request.RequestID)
	
	// Verificar cache se não for forçada a atualização
	if o.config.EnableCache && !request.ForceRefresh {
		if cachedResponse, found := o.checkCache(request); found {
			return cachedResponse, nil
		}
	}
	
	// Criar resposta inicial
	response := &models.AssessmentResponse{
		ResponseID:      uuid.New().String(),
		RequestID:       request.RequestID,
		CorrelationID:   request.CorrelationID,
		UserID:          request.UserID,
		TenantID:        request.TenantID,
		Status:          models.StatusProcessing,
		DataSources:     []string{},
	}
	
	// Executar avaliações de acordo com os tipos solicitados
	var wg sync.WaitGroup
	var resultMutex sync.Mutex
	errCh := make(chan error, len(request.AssessmentTypes))
	
	// Para cada tipo de avaliação, executar em paralelo
	for _, assessmentType := range request.AssessmentTypes {
		wg.Add(1)
		
		go func(aType models.AssessmentType) {
			defer wg.Done()
			
			var err error
			
			switch aType {
			case models.TypeIdentity:
				err = o.performIdentityAssessment(ctx, &request, response, &resultMutex)
			case models.TypeCredit:
				err = o.performCreditAssessment(ctx, &request, response, &resultMutex)
			case models.TypeFraud:
				err = o.performFraudAssessment(ctx, &request, response, &resultMutex)
			case models.TypeCompliance:
				err = o.performComplianceAssessment(ctx, &request, response, &resultMutex)
			case models.TypeRisk:
				err = o.performRiskAssessment(ctx, &request, response, &resultMutex)
			case models.TypeComprehensive:
				err = o.performComprehensiveAssessment(ctx, &request, response, &resultMutex)
			default:
				err = fmt.Errorf("tipo de avaliação não suportado: %s", aType)
			}
			
			if err != nil {
				errCh <- err
			}
		}(assessmentType)
	}
	
	// Aguardar conclusão de todas as avaliações
	wg.Wait()
	close(errCh)
	
	// Verificar erros
	var errors []string
	for err := range errCh {
		errors = append(errors, err.Error())
	}
	
	// Definir status final
	if len(errors) > 0 {
		if len(errors) == len(request.AssessmentTypes) {
			// Todas as avaliações falharam
			response.Status = models.StatusFailed
			response.ErrorDetails = &models.ErrorDetails{
				ErrorCode:      "ASSESSMENT_FAILED",
				ErrorMessage:   "Todas as avaliações falharam",
				FailedServices: errors,
				ErrorTimestamp: time.Now(),
				Retryable:      true,
				PartialResults: false,
			}
		} else {
			// Algumas avaliações falharam
			response.Status = models.StatusCompleted
			response.ErrorDetails = &models.ErrorDetails{
				ErrorCode:      "PARTIAL_FAILURE",
				ErrorMessage:   "Algumas avaliações falharam",
				FailedServices: errors,
				ErrorTimestamp: time.Now(),
				Retryable:      true,
				PartialResults: true,
			}
		}
	} else {
		response.Status = models.StatusCompleted
	}
	
	// Consolidar resultados
	o.consolidateResults(response)
	
	// Registrar tempo de processamento
	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	response.CompletedAt = time.Now()
	
	// Armazenar em cache se habilitado
	if o.config.EnableCache && response.Status == models.StatusCompleted {
		o.storeInCache(request, response)
	}
	
	// Publicar evento de avaliação concluída
	if o.eventBus != nil {
		o.publishAssessmentEvent(request, response)
	}
	
	return response, nil
}

// GetAssessmentStatus obtém o status atual de uma avaliação
func (o *BureauOrchestrator) GetAssessmentStatus(ctx context.Context, requestID string) (*models.AssessmentResponse, error) {
	// Verificar se a avaliação está ativa
	o.assessmentsMutex.RLock()
	_, isActive := o.activeAssessments[requestID]
	o.assessmentsMutex.RUnlock()
	
	if isActive {
		// Avaliação ainda em processamento
		return &models.AssessmentResponse{
			RequestID: requestID,
			Status:    models.StatusProcessing,
		}, nil
	}
	
	// Verificar no cache
	if o.cache != nil {
		if cachedData, found := o.cache.Get("assessment:" + requestID); found {
			if response, ok := cachedData.(*models.AssessmentResponse); ok {
				return response, nil
			}
		}
	}
	
	// Não encontrado
	return nil, fmt.Errorf("avaliação não encontrada: %s", requestID)
}

// GetCreditProviders retorna os provedores de crédito disponíveis
func (o *BureauOrchestrator) GetCreditProviders() []string {
	o.providersMutex.RLock()
	defer o.providersMutex.RUnlock()
	
	providers := []string{}
	if o.creditProviderFactory != nil {
		providers = o.creditProviderFactory.ListAvailableProviders()
	}
	
	return providers
}

// CancelAssessment cancela uma avaliação em andamento
func (o *BureauOrchestrator) CancelAssessment(ctx context.Context, requestID string) error {
	o.assessmentsMutex.Lock()
	defer o.assessmentsMutex.Unlock()
	
	if _, exists := o.activeAssessments[requestID]; !exists {
		return fmt.Errorf("avaliação não encontrada ou já concluída: %s", requestID)
	}
	
	// Remover da lista de avaliações ativas
	delete(o.activeAssessments, requestID)
	
	// Criar resposta de avaliação cancelada
	if o.cache != nil {
		response := &models.AssessmentResponse{
			RequestID:     requestID,
			Status:        models.StatusCancelled,
			CompletedAt:   time.Now(),
			ErrorDetails: &models.ErrorDetails{
				ErrorCode:     "ASSESSMENT_CANCELLED",
				ErrorMessage:  "Avaliação cancelada pelo usuário",
				ErrorTimestamp: time.Now(),
				Retryable:     true,
				PartialResults: false,
			},
		}
		
		// Armazenar em cache
		o.cache.Set("assessment:"+requestID, response, time.Duration(o.config.CacheTTLSeconds)*time.Second)
	}
	
	return nil
}

// BatchRequestAssessment inicia múltiplas avaliações em lote
func (o *BureauOrchestrator) BatchRequestAssessment(
	ctx context.Context,
	requests []models.AssessmentRequest,
) ([]*models.AssessmentResponse, error) {
	responses := make([]*models.AssessmentResponse, len(requests))
	
	// Processar cada solicitação individualmente
	var wg sync.WaitGroup
	var errorsMutex sync.Mutex
	errors := make([]error, 0)
	
	for i, req := range requests {
		wg.Add(1)
		
		go func(idx int, request models.AssessmentRequest) {
			defer wg.Done()
			
			resp, err := o.RequestAssessment(ctx, request)
			responses[idx] = resp
			
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("erro na solicitação %d: %w", idx, err))
				errorsMutex.Unlock()
			}
		}(i, req)
	}
	
	wg.Wait()
	
	// Verificar erros
	if len(errors) > 0 {
		return responses, fmt.Errorf("ocorreram erros em %d solicitações", len(errors))
	}
	
	return responses, nil
}

// Métodos privados

// registerActiveAssessment registra uma avaliação ativa
func (o *BureauOrchestrator) registerActiveAssessment(requestID string, request *models.AssessmentRequest) {
	o.assessmentsMutex.Lock()
	defer o.assessmentsMutex.Unlock()
	
	o.activeAssessments[requestID] = request
}

// unregisterActiveAssessment remove uma avaliação ativa
func (o *BureauOrchestrator) unregisterActiveAssessment(requestID string) {
	o.assessmentsMutex.Lock()
	defer o.assessmentsMutex.Unlock()
	
	delete(o.activeAssessments, requestID)
}

// checkCache verifica se há uma resposta em cache para a solicitação
func (o *BureauOrchestrator) checkCache(request models.AssessmentRequest) (*models.AssessmentResponse, bool) {
	if o.cache == nil {
		return nil, false
	}
	
	cacheKey := fmt.Sprintf("assessment:%s:%s:%s", request.UserID, request.TenantID, request.RequestID)
	cachedData, found := o.cache.Get(cacheKey)
	if !found {
		return nil, false
	}
	
	response, ok := cachedData.(*models.AssessmentResponse)
	if !ok {
		return nil, false
	}
	
	return response, true
}

// storeInCache armazena uma resposta em cache
func (o *BureauOrchestrator) storeInCache(request models.AssessmentRequest, response *models.AssessmentResponse) {
	if o.cache == nil {
		return
	}
	
	cacheKey := fmt.Sprintf("assessment:%s:%s:%s", request.UserID, request.TenantID, request.RequestID)
	o.cache.Set(cacheKey, response, time.Duration(o.config.CacheTTLSeconds)*time.Second)
	
	// Armazenar também pelo RequestID para consultas de status
	o.cache.Set("assessment:"+request.RequestID, response, time.Duration(o.config.CacheTTLSeconds)*time.Second)
}

// publishAssessmentEvent publica um evento de avaliação
func (o *BureauOrchestrator) publishAssessmentEvent(request models.AssessmentRequest, response *models.AssessmentResponse) {
	if o.eventBus == nil {
		return
	}
	
	// Criar evento simplificado (sem dados sensíveis completos)
	event := map[string]interface{}{
		"eventType":     "ASSESSMENT_COMPLETED",
		"requestId":     request.RequestID,
		"correlationId": request.CorrelationID,
		"userId":        request.UserID,
		"tenantId":      request.TenantID,
		"timestamp":     time.Now(),
		"status":        response.Status,
		"trustScore":    response.TrustScore,
		"riskLevel":     response.RiskLevel,
		"decision":      response.Decision,
	}
	
	// Publicar evento de forma assíncrona
	go func() {
		_ = o.eventBus.PublishEvent(context.Background(), event)
	}()
}