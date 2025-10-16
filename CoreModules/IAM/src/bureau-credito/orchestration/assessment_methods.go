/**
 * @file assessment_methods.go
 * @description Implementações de métodos de avaliação para o orquestrador
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"innovabiz/iam/src/bureau-credito/adapters"
	"innovabiz/iam/src/bureau-credito/fraud-detection"
	"innovabiz/iam/src/bureau-credito/orchestration/models"
	"innovabiz/iam/src/bureau-credito/risk-engine"
)

// performIdentityAssessment executa a avaliação de identidade
func (o *BureauOrchestrator) performIdentityAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	if request.IdentityData == nil {
		return fmt.Errorf("dados de identidade são obrigatórios para avaliação de identidade")
	}

	// Obter resultado da avaliação de identidade
	identityResults, err := o.identityService.VerifyIdentity(ctx, request)
	if err != nil {
		return fmt.Errorf("falha na avaliação de identidade: %w", err)
	}

	// Atualizar resposta com resultados de identidade
	resultMutex.Lock()
	response.IdentityResults = identityResults
	response.DataSources = append(response.DataSources, "IDENTITY_VERIFICATION")
	resultMutex.Unlock()

	return nil
}

// performCreditAssessment executa a avaliação de crédito
func (o *BureauOrchestrator) performCreditAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	// Verificar dados de crédito
	if request.CreditData == nil {
		return fmt.Errorf("dados de crédito são obrigatórios para avaliação de crédito")
	}

	// Determinar quais provedores de crédito usar
	creditProviders := request.CreditProviders
	if len(creditProviders) == 0 {
		creditProviders = o.config.DefaultCreditProviders
	}

	// Criar resposta para resultados de crédito
	creditResults := &models.CreditResults{
		ProviderResponses: make(map[string]adapters.CreditReportResponse),
		ReportDate:        time.Now(),
	}

	// Para cada provedor, obter relatório de crédito
	var wg sync.WaitGroup
	var providerMutex sync.Mutex
	providerErrors := make([]string, 0)

	for _, providerID := range creditProviders {
		wg.Add(1)

		go func(provider string) {
			defer wg.Done()

			// Obter ou criar provedor
			creditProvider, err := o.getCreditProvider(provider)
			if err != nil {
				providerMutex.Lock()
				providerErrors = append(providerErrors, fmt.Sprintf("%s: %s", provider, err.Error()))
				providerMutex.Unlock()
				return
			}

			// Criar solicitação de relatório de crédito
			reportRequest := adapters.CreditReportRequest{
				UserID:           request.UserID,
				TenantID:         request.TenantID,
				DocumentNumber:   request.IdentityData.DocumentNumber,
				DocumentType:     request.IdentityData.DocumentType,
				Name:             request.IdentityData.Name,
				DateOfBirth:      request.IdentityData.DateOfBirth,
				Nationality:      request.IdentityData.Nationality,
				Email:            request.IdentityData.Email,
				PhoneNumber:      request.IdentityData.PhoneNumber,
				Address:          request.IdentityData.Address,
				ContextData:      request.CustomAttributes,
				RequestTimestamp: time.Now(),
			}

			// Adicionar dados de crédito se disponíveis
			if request.CreditData != nil {
				reportRequest.AnnualIncome = request.CreditData.AnnualIncome
				reportRequest.Occupation = request.CreditData.Occupation
				reportRequest.EmploymentStatus = request.CreditData.EmploymentStatus
				reportRequest.HasPendingLoans = request.CreditData.HasPendingLoans
			}

			// Solicitar relatório de crédito
			reportResponse, err := creditProvider.GetCreditReport(ctx, reportRequest)
			if err != nil {
				providerMutex.Lock()
				providerErrors = append(providerErrors, fmt.Sprintf("%s: %s", provider, err.Error()))
				providerMutex.Unlock()
				return
			}

			// Adicionar resposta ao resultado
			providerMutex.Lock()
			creditResults.ProviderResponses[provider] = reportResponse
			providerMutex.Unlock()
		}(providerID)
	}

	// Aguardar conclusão de todas as solicitações
	wg.Wait()

	// Verificar erros
	if len(providerErrors) == len(creditProviders) {
		return fmt.Errorf("falha em todos os provedores de crédito: %v", providerErrors)
	}

	// Consolidar resultados de todos os provedores
	o.consolidateCreditResults(creditResults)

	// Atualizar resposta com resultados de crédito
	resultMutex.Lock()
	response.CreditResults = creditResults
	response.DataSources = append(response.DataSources, "CREDIT_ASSESSMENT")
	for provider := range creditResults.ProviderResponses {
		response.DataSources = append(response.DataSources, "CREDIT_"+provider)
	}
	resultMutex.Unlock()

	return nil
}

// performFraudAssessment executa a avaliação de fraude
func (o *BureauOrchestrator) performFraudAssessment(
	ctx context.Context,
	request *models.AssessmentRequest,
	response *models.AssessmentResponse,
	resultMutex *sync.Mutex,
) error {
	// Criar solicitação de detecção de fraude
	fraudRequest := frauddetection.DetectionRequest{
		UserID:    request.UserID,
		TenantID:  request.TenantID,
		RequestID: request.RequestID,
	}

	// Adicionar dados de dispositivo, se disponíveis
	if request.DeviceData != nil {
		fraudRequest.DeviceID = request.DeviceData.DeviceID
		fraudRequest.DeviceInfo = &frauddetection.DeviceInfo{
			DeviceType:       request.DeviceData.DeviceType,
			OS:               request.DeviceData.OS,
			OSVersion:        request.DeviceData.OSVersion,
			Browser:          request.DeviceData.Browser,
			BrowserVersion:   request.DeviceData.BrowserVersion,
			ScreenResolution: request.DeviceData.ScreenResolution,
			DeviceModel:      request.DeviceData.DeviceModel,
			DeviceBrand:      request.DeviceData.DeviceBrand,
			Jailbroken:       request.DeviceData.Jailbroken,
			Emulator:         request.DeviceData.Emulator,
			DeviceLanguage:   request.DeviceData.DeviceLanguage,
			TimeZone:         request.DeviceData.TimeZone,
			DeviceFingerprint: request.DeviceData.DeviceFingerprint,
		}
	}

	// Adicionar dados de rede, se disponíveis
	if request.NetworkData != nil {
		fraudRequest.NetworkInfo = &frauddetection.NetworkInfo{
			IPAddress:      request.NetworkData.IPAddress,
			ISP:            request.NetworkData.ISP,
			ConnectionType: request.NetworkData.ConnectionType,
			HostName:       request.NetworkData.HostName,
			ASNumber:       request.NetworkData.ASNumber,
			ProxyDetected:  request.NetworkData.ProxyDetected,
			VPNDetected:    request.NetworkData.VPNDetected,
			TorDetected:    request.NetworkData.TorDetected,
			Latitude:       request.NetworkData.Latitude,
			Longitude:      request.NetworkData.Longitude,
			Country:        request.NetworkData.Country,
			Region:         request.NetworkData.Region,
			City:           request.NetworkData.City,
		}
	}

	// Adicionar dados comportamentais, se disponíveis
	if request.BehavioralData != nil {
		fraudRequest.BehavioralInfo = &frauddetection.BehavioralInfo{
			SessionID:        request.BehavioralData.SessionID,
			SessionDuration:  request.BehavioralData.SessionDuration,
			ClickPattern:     request.BehavioralData.ClickPattern,
			TypingSpeed:      request.BehavioralData.TypingSpeed,
			NavigationFlow:   request.BehavioralData.NavigationFlow,
			TimeOnPage:       request.BehavioralData.TimeOnPage,
			InteractionCount: request.BehavioralData.InteractionCount,
			UnusualActivity:  request.BehavioralData.UnusualActivity,
		}
	}

	// Adicionar dados de transação, se disponíveis
	if request.TransactionData != nil {
		fraudRequest.TransactionInfo = &frauddetection.TransactionInfo{
			TransactionID:       request.TransactionData.TransactionID,
			TransactionType:     request.TransactionData.TransactionType,
			Amount:              request.TransactionData.Amount,
			Currency:            request.TransactionData.Currency,
			Timestamp:           request.TransactionData.Timestamp,
			MerchantID:          request.TransactionData.MerchantID,
			MerchantName:        request.TransactionData.MerchantName,
			MerchantCategory:    request.TransactionData.MerchantCategory,
			Description:         request.TransactionData.Description,
			PaymentMethod:       request.TransactionData.PaymentMethod,
			RecipientID:         request.TransactionData.RecipientID,
			SourceAccount:       request.TransactionData.SourceAccount,
			DestinationAccount:  request.TransactionData.DestinationAccount,
		}
	}

	// Executar detecção de fraude
	fraudResponse, err := o.fraudEngine.DetectFraud(ctx, fraudRequest)
	if err != nil {
		return fmt.Errorf("falha na detecção de fraude: %w", err)
	}

	// Criar resultados de fraude
	fraudResults := &models.FraudResults{
		FraudDetected:    fraudResponse.FraudDetected,
		FraudProbability: fraudResponse.FraudProbability,
		FraudScore:       fraudResponse.FraudScore,
		RiskFactors:      fraudResponse.RiskFactors,
		AnomalyDetails:   fraudResponse.Anomalies,
		FraudVerdict:     fraudResponse.Verdict,
		DeviceReputation: fraudResponse.DeviceReputation,
		IPReputation:     fraudResponse.IPReputation,
		DetectionDetails: fraudResponse.Details,
	}

	// Atualizar resposta com resultados de fraude
	resultMutex.Lock()
	response.FraudResults = fraudResults
	response.DataSources = append(response.DataSources, "FRAUD_DETECTION")
	resultMutex.Unlock()

	return nil
}