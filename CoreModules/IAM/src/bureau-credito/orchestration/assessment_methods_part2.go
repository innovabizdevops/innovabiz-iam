/**
 * @file assessment_methods_part2.go
 * @description Métodos adicionais de avaliação para o orquestrador
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package orchestration

import (
	"context"
	"fmt"
	"sync"

	"innovabiz/iam/src/bureau-credito/orchestration/models"
	"innovabiz/iam/src/bureau-credito/risk-engine"
)

// performComplianceAssessment executa a avaliação de conformidade regulatória
func (o *BureauOrchestrator) performComplianceAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	// Verificar se há serviço de conformidade configurado
	if o.complianceService == nil {
		return fmt.Errorf("serviço de conformidade não configurado")
	}

	// Obter regras de conformidade a serem verificadas
	complianceRules := request.ComplianceRules
	if len(complianceRules) == 0 {
		// Usar regras padrão se não especificadas
		var err error
		complianceRules, err = o.complianceService.GetComplianceRules(ctx, request.TenantID)
		if err != nil {
			return fmt.Errorf("falha ao obter regras de conformidade: %w", err)
		}
	}

	// Se ainda não houver regras, usar as configuradas globalmente
	if len(complianceRules) == 0 {
		complianceRules = o.config.DefaultComplianceRules
	}

	// Executar verificação de conformidade
	complianceResults, err := o.complianceService.CheckCompliance(ctx, request)
	if err != nil {
		return fmt.Errorf("falha na verificação de conformidade: %w", err)
	}

	// Atualizar resposta com resultados de conformidade
	resultMutex.Lock()
	response.ComplianceResults = complianceResults
	response.DataSources = append(response.DataSources, "COMPLIANCE_CHECK")
	resultMutex.Unlock()

	return nil
}

// performRiskAssessment executa a avaliação de risco
func (o *BureauOrchestrator) performRiskAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	// Verificar se há motor de risco configurado
	if o.riskEngine == nil {
		return fmt.Errorf("motor de avaliação de risco não configurado")
	}

	// Criar solicitação para avaliação de risco
	riskRequest := risk_engine.AssessmentRequest{
		UserID:    request.UserID,
		TenantID:  request.TenantID,
		RequestID: request.RequestID,
	}

	// Adicionar dados contextuais baseados na solicitação
	contextData := make(map[string]interface{})

	// Adicionar dados de identidade se disponíveis
	if request.IdentityData != nil {
		contextData["identity"] = map[string]interface{}{
			"documentNumber":    request.IdentityData.DocumentNumber,
			"documentType":      request.IdentityData.DocumentType,
			"name":              request.IdentityData.Name,
			"dateOfBirth":       request.IdentityData.DateOfBirth,
			"email":             request.IdentityData.Email,
			"phoneNumber":       request.IdentityData.PhoneNumber,
			"address":           request.IdentityData.Address,
			"nationality":       request.IdentityData.Nationality,
			"verificationLevel": request.IdentityData.VerificationLevel,
		}
	}

	// Adicionar dados de crédito se disponíveis
	if request.CreditData != nil {
		contextData["credit"] = map[string]interface{}{
			"accountAge":        request.CreditData.AccountAge,
			"paymentHistory":    request.CreditData.PaymentHistory,
			"creditHistory":     request.CreditData.CreditHistory,
			"annualIncome":      request.CreditData.AnnualIncome,
			"occupation":        request.CreditData.Occupation,
			"employmentStatus":  request.CreditData.EmploymentStatus,
			"assets":            request.CreditData.Assets,
			"liabilities":       request.CreditData.Liabilities,
			"hasPendingLoans":   request.CreditData.HasPendingLoans,
			"pendingLoansAmount": request.CreditData.PendingLoansAmount,
		}
	}

	// Adicionar dados de dispositivo se disponíveis
	if request.DeviceData != nil {
		contextData["device"] = map[string]interface{}{
			"deviceId":          request.DeviceData.DeviceID,
			"deviceType":        request.DeviceData.DeviceType,
			"os":                request.DeviceData.OS,
			"osVersion":         request.DeviceData.OSVersion,
			"jailbroken":        request.DeviceData.Jailbroken,
			"emulator":          request.DeviceData.Emulator,
			"deviceFingerprint": request.DeviceData.DeviceFingerprint,
		}
	}

	// Adicionar dados de rede se disponíveis
	if request.NetworkData != nil {
		contextData["network"] = map[string]interface{}{
			"ipAddress":      request.NetworkData.IPAddress,
			"proxyDetected":  request.NetworkData.ProxyDetected,
			"vpnDetected":    request.NetworkData.VPNDetected,
			"torDetected":    request.NetworkData.TorDetected,
			"country":        request.NetworkData.Country,
			"region":         request.NetworkData.Region,
			"city":           request.NetworkData.City,
		}
	}

	// Adicionar dados de transação se disponíveis
	if request.TransactionData != nil {
		contextData["transaction"] = map[string]interface{}{
			"transactionId":   request.TransactionData.TransactionID,
			"transactionType": request.TransactionData.TransactionType,
			"amount":          request.TransactionData.Amount,
			"currency":        request.TransactionData.Currency,
			"timestamp":       request.TransactionData.Timestamp,
			"merchantId":      request.TransactionData.MerchantID,
			"merchantCategory": request.TransactionData.MerchantCategory,
		}
	}

	// Adicionar dados comportamentais se disponíveis
	if request.BehavioralData != nil {
		contextData["behavioral"] = map[string]interface{}{
			"sessionId":        request.BehavioralData.SessionID,
			"sessionDuration":  request.BehavioralData.SessionDuration,
			"unusualActivity":  request.BehavioralData.UnusualActivity,
			"interactionCount": request.BehavioralData.InteractionCount,
		}
	}

	// Adicionar atributos personalizados se disponíveis
	if len(request.CustomAttributes) > 0 {
		for k, v := range request.CustomAttributes {
			contextData[k] = v
		}
	}

	// Definir dados de contexto para avaliação de risco
	riskRequest.ContextData = contextData

	// Executar avaliação de risco
	riskResponse, err := o.riskEngine.EvaluateRisk(ctx, riskRequest)
	if err != nil {
		return fmt.Errorf("falha na avaliação de risco: %w", err)
	}

	// Criar resultados de risco
	riskResults := &models.RiskResults{
		RiskScore:          riskResponse.RiskScore,
		RiskLevel:          riskResponse.RiskLevel,
		RiskFactors:        riskResponse.RiskFactors,
		TrustLevel:         riskResponse.TrustLevel,
		RecommendedActions: riskResponse.RecommendedActions,
		AllowOperation:     riskResponse.AllowOperation,
		RequireAdditionalAuth: riskResponse.RequireAdditionalAuth,
		ContextualData:     riskResponse.ContextualData,
	}

	// Atualizar resposta com resultados de risco
	resultMutex.Lock()
	response.RiskResults = riskResults
	response.DataSources = append(response.DataSources, "RISK_ASSESSMENT")
	resultMutex.Unlock()

	return nil
}

// performComprehensiveAssessment executa uma avaliação abrangente combinando todas as outras
func (o *BureauOrchestrator) performComprehensiveAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4) // para os 4 tipos de avaliação
	
	// Executar todas as avaliações em paralelo
	wg.Add(4)
	
	// Identidade
	go func() {
		defer wg.Done()
		if err := o.performIdentityAssessment(ctx, request, response, resultMutex); err != nil {
			errChan <- fmt.Errorf("falha na avaliação de identidade: %w", err)
		}
	}()
	
	// Crédito
	go func() {
		defer wg.Done()
		if err := o.performCreditAssessment(ctx, request, response, resultMutex); err != nil {
			errChan <- fmt.Errorf("falha na avaliação de crédito: %w", err)
		}
	}()
	
	// Fraude
	go func() {
		defer wg.Done()
		if err := o.performFraudAssessment(ctx, request, response, resultMutex); err != nil {
			errChan <- fmt.Errorf("falha na avaliação de fraude: %w", err)
		}
	}()
	
	// Conformidade
	go func() {
		defer wg.Done()
		if err := o.performComplianceAssessment(ctx, request, response, resultMutex); err != nil {
			errChan <- fmt.Errorf("falha na avaliação de conformidade: %w", err)
		}
	}()
	
	// Aguardar conclusão de todas as avaliações
	wg.Wait()
	close(errChan)
	
	// Verificar erros
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}
	
	// Se todas as avaliações falharam
	if len(errors) == 4 {
		return fmt.Errorf("todas as avaliações abrangentes falharam: %v", errors)
	}
	
	// Se algumas avaliações falharam, registrar os erros mas continuar
	if len(errors) > 0 {
		// Armazenar erros na resposta
		resultMutex.Lock()
		if response.ErrorDetails == nil {
			response.ErrorDetails = &models.ErrorDetails{
				ErrorCode:      "PARTIAL_COMPREHENSIVE_FAILURE",
				ErrorMessage:   "Algumas avaliações abrangentes falharam",
				FailedServices: errors,
				PartialResults: true,
				Retryable:      true,
			}
		} else {
			response.ErrorDetails.FailedServices = append(response.ErrorDetails.FailedServices, errors...)
			response.ErrorDetails.PartialResults = true
		}
		resultMutex.Unlock()
	}
	
	// Finalmente, executar avaliação de risco que depende dos resultados anteriores
	if err := o.performRiskAssessment(ctx, request, response, resultMutex); err != nil {
		return fmt.Errorf("falha na avaliação de risco após avaliações abrangentes: %w", err)
	}
	
	return nil
}

// getCreditProvider obtém ou cria um provedor de crédito pelo ID
func (o *BureauOrchestrator) getCreditProvider(providerID string) (adapters.CreditProvider, error) {
	o.providersMutex.RLock()
	provider, exists := o.creditProviders[providerID]
	o.providersMutex.RUnlock()
	
	if exists {
		return provider, nil
	}
	
	// Criar novo provedor se não existir
	if o.creditProviderFactory == nil {
		return nil, fmt.Errorf("fábrica de provedores de crédito não configurada")
	}
	
	newProvider, err := o.creditProviderFactory.CreateProvider(providerID)
	if err != nil {
		return nil, err
	}
	
	// Armazenar para uso futuro
	o.providersMutex.Lock()
	o.creditProviders[providerID] = newProvider
	o.providersMutex.Unlock()
	
	return newProvider, nil
}