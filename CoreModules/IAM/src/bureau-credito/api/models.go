/**
 * @file models.go
 * @description Modelos para API REST/GraphQL do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package api

import (
	"time"

	"innovabiz/iam/src/bureau-credito/orchestration/models"
)

// AssessmentRequest representa a solicitação de API para uma nova avaliação
type AssessmentRequest struct {
	// Identificadores
	UserID        string            `json:"userId" validate:"required" example:"user-123456"`
	TenantID      string            `json:"tenantId" validate:"required" example:"tenant-789012"`
	CorrelationID string            `json:"correlationId,omitempty" example:"corr-abcdef123456"`
	
	// Configuração de avaliação
	AssessmentTypes  []string        `json:"assessmentTypes" validate:"required,dive,oneof=IDENTITY CREDIT FRAUD COMPLIANCE RISK COMPREHENSIVE" example:"FRAUD,CREDIT"`
	CreditProviders  []string        `json:"creditProviders,omitempty" example:"SERASA,SPC"`
	IdentityProviders []string       `json:"identityProviders,omitempty" example:"SERASA,SERPRO"`
	ComplianceRules  []string        `json:"complianceRules,omitempty" example:"AML,KYC,FATCA"`
	
	// Dados para avaliação
	IdentityData     *IdentityData   `json:"identityData,omitempty"`
	CreditData       *CreditData     `json:"creditData,omitempty"`
	DeviceData       *DeviceData     `json:"deviceData,omitempty"`
	NetworkData      *NetworkData    `json:"networkData,omitempty"`
	TransactionData  *TransactionData `json:"transactionData,omitempty"`
	BehavioralData   *BehavioralData `json:"behavioralData,omitempty"`
	
	// Configurações de processamento
	TimeoutMs        int64           `json:"timeoutMs,omitempty" example:"30000"`
	ForceRefresh     bool            `json:"forceRefresh" example:"false"`
	RequireAllResults bool           `json:"requireAllResults" example:"true"`
	FailFast         bool            `json:"failFast" example:"false"`
	
	// Dados adicionais específicos do contexto
	CustomAttributes map[string]interface{} `json:"customAttributes,omitempty"`
}

// IdentityData representa informações de identidade para API
type IdentityData struct {
	DocumentNumber   string            `json:"documentNumber,omitempty" validate:"omitempty" example:"12345678901"`
	DocumentType     string            `json:"documentType,omitempty" example:"CPF"`
	Name             string            `json:"name,omitempty" example:"João Silva"`
	DateOfBirth      string            `json:"dateOfBirth,omitempty" example:"1990-01-01"`
	Email            string            `json:"email,omitempty" validate:"omitempty,email" example:"joao.silva@email.com"`
	PhoneNumber      string            `json:"phoneNumber,omitempty" example:"+55 11 98765-4321"`
	Address          string            `json:"address,omitempty" example:"Rua Exemplo, 123, São Paulo - SP"`
	Nationality      string            `json:"nationality,omitempty" example:"Brasileira"`
	BiometricData    map[string]string `json:"biometricData,omitempty"`
	VerificationLevel int               `json:"verificationLevel" example:"2"`
}

// CreditData representa informações financeiras para API
type CreditData struct {
	AccountAge       int               `json:"accountAge,omitempty" example:"730"`
	PaymentHistory   string            `json:"paymentHistory,omitempty" example:"GOOD"`
	CreditHistory    string            `json:"creditHistory,omitempty" example:"GOOD"`
	AnnualIncome     float64           `json:"annualIncome,omitempty" example:"120000"`
	Occupation       string            `json:"occupation,omitempty" example:"Engineer"`
	EmploymentStatus string            `json:"employmentStatus,omitempty" example:"EMPLOYED"`
	Assets           float64           `json:"assets,omitempty" example:"500000"`
	Liabilities      float64           `json:"liabilities,omitempty" example:"100000"`
	HasPendingLoans  bool              `json:"hasPendingLoans" example:"false"`
	PendingLoansAmount float64         `json:"pendingLoansAmount,omitempty" example:"0"`
}

// DeviceData representa informações sobre o dispositivo para API
type DeviceData struct {
	DeviceID         string            `json:"deviceId" validate:"required" example:"device-xyz789"`
	DeviceType       string            `json:"deviceType" validate:"required" example:"MOBILE"`
	OS               string            `json:"os,omitempty" example:"Android"`
	OSVersion        string            `json:"osVersion,omitempty" example:"12"`
	Browser          string            `json:"browser,omitempty" example:"Chrome"`
	BrowserVersion   string            `json:"browserVersion,omitempty" example:"96.0.4664.110"`
	ScreenResolution string            `json:"screenResolution,omitempty" example:"1080x2400"`
	DeviceModel      string            `json:"deviceModel,omitempty" example:"Samsung Galaxy S21"`
	DeviceBrand      string            `json:"deviceBrand,omitempty" example:"Samsung"`
	Jailbroken       bool              `json:"jailbroken" example:"false"`
	Emulator         bool              `json:"emulator" example:"false"`
	DeviceLanguage   string            `json:"deviceLanguage,omitempty" example:"pt-BR"`
	TimeZone         string            `json:"timeZone,omitempty" example:"America/Sao_Paulo"`
	DeviceFingerprint string           `json:"deviceFingerprint,omitempty" example:"a1b2c3d4e5f6g7h8i9j0"`
}

// NetworkData representa informações de rede para API
type NetworkData struct {
	IPAddress        string            `json:"ipAddress" validate:"required,ip" example:"187.32.45.67"`
	ISP              string            `json:"isp,omitempty" example:"Vivo"`
	ConnectionType   string            `json:"connectionType,omitempty" example:"4G"`
	HostName         string            `json:"hostName,omitempty" example:"host-example"`
	ASNumber         string            `json:"asNumber,omitempty" example:"AS13878"`
	ProxyDetected    bool              `json:"proxyDetected" example:"false"`
	VPNDetected      bool              `json:"vpnDetected" example:"false"`
	TorDetected      bool              `json:"torDetected" example:"false"`
	Latitude         float64           `json:"latitude,omitempty" example:"-23.5505"`
	Longitude        float64           `json:"longitude,omitempty" example:"-46.6333"`
	Country          string            `json:"country,omitempty" example:"Brasil"`
	Region           string            `json:"region,omitempty" example:"São Paulo"`
	City             string            `json:"city,omitempty" example:"São Paulo"`
}

// TransactionData representa informações sobre transações financeiras para API
type TransactionData struct {
	TransactionID    string            `json:"transactionId" validate:"required" example:"tx-123456789"`
	TransactionType  string            `json:"transactionType" validate:"required" example:"PURCHASE"`
	Amount           float64           `json:"amount" validate:"required" example:"250.75"`
	Currency         string            `json:"currency" validate:"required" example:"BRL"`
	Timestamp        time.Time         `json:"timestamp" validate:"required" example:"2025-03-15T14:30:45Z"`
	MerchantID       string            `json:"merchantId,omitempty" example:"merchant-123"`
	MerchantName     string            `json:"merchantName,omitempty" example:"Loja Online ABC"`
	MerchantCategory string            `json:"merchantCategory,omitempty" example:"RETAIL"`
	Description      string            `json:"description,omitempty" example:"Compra online"`
	PaymentMethod    string            `json:"paymentMethod,omitempty" example:"CREDIT_CARD"`
	RecipientID      string            `json:"recipientId,omitempty" example:"recipient-456"`
	SourceAccount    string            `json:"sourceAccount,omitempty" example:"account-src-789"`
	DestinationAccount string          `json:"destinationAccount,omitempty" example:"account-dst-012"`
}

// BehavioralData representa informações comportamentais do usuário para API
type BehavioralData struct {
	SessionID        string            `json:"sessionId,omitempty" example:"session-abcdef123456"`
	SessionDuration  int               `json:"sessionDuration,omitempty" example:"1800"`
	ClickPattern     string            `json:"clickPattern,omitempty" example:"NORMAL"`
	TypingSpeed      int               `json:"typingSpeed,omitempty" example:"180"`
	NavigationFlow   []string          `json:"navigationFlow,omitempty" example:"HOME,PRODUCT,CART,CHECKOUT"`
	TimeOnPage       int               `json:"timeOnPage,omitempty" example:"45"`
	InteractionCount int               `json:"interactionCount,omitempty" example:"37"`
	UnusualActivity  bool              `json:"unusualActivity" example:"false"`
	ActivityDetails  map[string]interface{} `json:"activityDetails,omitempty"`
}

// AssessmentResponse representa a resposta da API para uma avaliação
type AssessmentResponse struct {
	// Identificadores
	ResponseID      string            `json:"responseId" example:"resp-789012345"`
	RequestID       string            `json:"requestId" example:"req-123456789"`
	CorrelationID   string            `json:"correlationId,omitempty" example:"corr-abcdef123456"`
	UserID          string            `json:"userId" example:"user-123456"`
	TenantID        string            `json:"tenantId" example:"tenant-789012"`
	
	// Status
	Status          string            `json:"status" example:"COMPLETED"`
	CompletedAt     time.Time         `json:"completedAt,omitempty" example:"2025-03-15T14:31:05Z"`
	ProcessingTimeMs int64            `json:"processingTimeMs" example:"20043"`
	
	// Resultados consolidados
	TrustScore      int               `json:"trustScore" example:"85"`
	RiskLevel       string            `json:"riskLevel" example:"LOW"`
	Decision        string            `json:"decision" example:"APPROVE"`
	Confidence      float64           `json:"confidence" example:"92.5"`
	
	// Resultados detalhados (mantidos opacos para a API)
	IdentityResults map[string]interface{} `json:"identityResults,omitempty"`
	CreditResults   map[string]interface{} `json:"creditResults,omitempty"`
	FraudResults    map[string]interface{} `json:"fraudResults,omitempty"`
	ComplianceResults map[string]interface{} `json:"complianceResults,omitempty"`
	RiskResults     map[string]interface{} `json:"riskResults,omitempty"`
	
	// Ações recomendadas
	RequiredActions []string          `json:"requiredActions,omitempty" example:"VERIFY_IDENTITY,ADDITIONAL_FRAUD_VERIFICATION"`
	SuggestedActions []string         `json:"suggestedActions,omitempty" example:"UPGRADE_IDENTITY_VERIFICATION"`
	Warnings        []string          `json:"warnings,omitempty" example:"UNUSUAL_DEVICE_LOCATION"`
	
	// Detalhes da falha (se aplicável)
	ErrorDetails    *ErrorDetails     `json:"errorDetails,omitempty"`
	
	// Metadados
	DataSources     []string          `json:"dataSources,omitempty" example:"IDENTITY_VERIFICATION,FRAUD_DETECTION"`
}

// ErrorDetails representa detalhes de erros para API
type ErrorDetails struct {
	ErrorCode       string            `json:"errorCode,omitempty" example:"ASSESSMENT_FAILED"`
	ErrorMessage    string            `json:"errorMessage,omitempty" example:"Falha ao processar avaliação"`
	FailedServices  []string          `json:"failedServices,omitempty" example:"CREDIT_SERVICE,IDENTITY_SERVICE"`
	PartialResults  bool              `json:"partialResults" example:"true"`
	ErrorSource     string            `json:"errorSource,omitempty" example:"CREDIT_PROVIDER"`
}

// BatchAssessmentRequest representa uma solicitação em lote para avaliações
type BatchAssessmentRequest struct {
	Requests []AssessmentRequest `json:"requests" validate:"required,min=1,max=100"`
}

// BatchAssessmentResponse representa uma resposta em lote para avaliações
type BatchAssessmentResponse struct {
	Responses []AssessmentResponse `json:"responses"`
	Success   int                 `json:"success" example:"9"`
	Failed    int                 `json:"failed" example:"1"`
	Total     int                 `json:"total" example:"10"`
}

// AssessmentStatusResponse representa uma resposta de status de avaliação
type AssessmentStatusResponse struct {
	RequestID  string   `json:"requestId" example:"req-123456789"`
	Status     string   `json:"status" example:"PROCESSING"`
}

// Método para converter o modelo de API para o modelo interno
func (req *AssessmentRequest) ToInternalModel() *models.AssessmentRequest {
	// Converter tipos de avaliação
	assessmentTypes := make([]models.AssessmentType, 0, len(req.AssessmentTypes))
	for _, t := range req.AssessmentTypes {
		switch t {
		case "IDENTITY":
			assessmentTypes = append(assessmentTypes, models.TypeIdentity)
		case "CREDIT":
			assessmentTypes = append(assessmentTypes, models.TypeCredit)
		case "FRAUD":
			assessmentTypes = append(assessmentTypes, models.TypeFraud)
		case "COMPLIANCE":
			assessmentTypes = append(assessmentTypes, models.TypeCompliance)
		case "RISK":
			assessmentTypes = append(assessmentTypes, models.TypeRisk)
		case "COMPREHENSIVE":
			assessmentTypes = append(assessmentTypes, models.TypeComprehensive)
		}
	}

	// Criar modelo interno
	internalReq := &models.AssessmentRequest{
		UserID:          req.UserID,
		TenantID:        req.TenantID,
		CorrelationID:   req.CorrelationID,
		RequestTimestamp: time.Now(),
		AssessmentTypes: assessmentTypes,
		CreditProviders: req.CreditProviders,
		IdentityProviders: req.IdentityProviders,
		ComplianceRules: req.ComplianceRules,
		ForceRefresh:    req.ForceRefresh,
		RequireAllResults: req.RequireAllResults,
		FailFast:        req.FailFast,
		CustomAttributes: req.CustomAttributes,
	}

	// Definir timeout se fornecido
	if req.TimeoutMs > 0 {
		internalReq.Timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}

	// Converter dados de identidade se fornecidos
	if req.IdentityData != nil {
		internalReq.IdentityData = &models.IdentityData{
			DocumentNumber:   req.IdentityData.DocumentNumber,
			DocumentType:     req.IdentityData.DocumentType,
			Name:             req.IdentityData.Name,
			DateOfBirth:      req.IdentityData.DateOfBirth,
			Email:            req.IdentityData.Email,
			PhoneNumber:      req.IdentityData.PhoneNumber,
			Address:          req.IdentityData.Address,
			Nationality:      req.IdentityData.Nationality,
			BiometricData:    req.IdentityData.BiometricData,
			VerificationLevel: req.IdentityData.VerificationLevel,
		}
	}

	// Converter dados de crédito se fornecidos
	if req.CreditData != nil {
		internalReq.CreditData = &models.CreditData{
			AccountAge:      req.CreditData.AccountAge,
			PaymentHistory:  req.CreditData.PaymentHistory,
			CreditHistory:   req.CreditData.CreditHistory,
			AnnualIncome:    req.CreditData.AnnualIncome,
			Occupation:      req.CreditData.Occupation,
			EmploymentStatus: req.CreditData.EmploymentStatus,
			Assets:          req.CreditData.Assets,
			Liabilities:     req.CreditData.Liabilities,
			HasPendingLoans: req.CreditData.HasPendingLoans,
			PendingLoansAmount: req.CreditData.PendingLoansAmount,
		}
	}

	// Converter dados do dispositivo se fornecidos
	if req.DeviceData != nil {
		internalReq.DeviceData = &models.DeviceData{
			DeviceID:         req.DeviceData.DeviceID,
			DeviceType:       req.DeviceData.DeviceType,
			OS:               req.DeviceData.OS,
			OSVersion:        req.DeviceData.OSVersion,
			Browser:          req.DeviceData.Browser,
			BrowserVersion:   req.DeviceData.BrowserVersion,
			ScreenResolution: req.DeviceData.ScreenResolution,
			DeviceModel:      req.DeviceData.DeviceModel,
			DeviceBrand:      req.DeviceData.DeviceBrand,
			Jailbroken:       req.DeviceData.Jailbroken,
			Emulator:         req.DeviceData.Emulator,
			DeviceLanguage:   req.DeviceData.DeviceLanguage,
			TimeZone:         req.DeviceData.TimeZone,
			DeviceFingerprint: req.DeviceData.DeviceFingerprint,
		}
	}

	// Converter dados de rede se fornecidos
	if req.NetworkData != nil {
		internalReq.NetworkData = &models.NetworkData{
			IPAddress:      req.NetworkData.IPAddress,
			ISP:            req.NetworkData.ISP,
			ConnectionType: req.NetworkData.ConnectionType,
			HostName:       req.NetworkData.HostName,
			ASNumber:       req.NetworkData.ASNumber,
			ProxyDetected:  req.NetworkData.ProxyDetected,
			VPNDetected:    req.NetworkData.VPNDetected,
			TorDetected:    req.NetworkData.TorDetected,
			Latitude:       req.NetworkData.Latitude,
			Longitude:      req.NetworkData.Longitude,
			Country:        req.NetworkData.Country,
			Region:         req.NetworkData.Region,
			City:           req.NetworkData.City,
		}
	}

	// Converter dados de transação se fornecidos
	if req.TransactionData != nil {
		internalReq.TransactionData = &models.TransactionData{
			TransactionID:   req.TransactionData.TransactionID,
			TransactionType: req.TransactionData.TransactionType,
			Amount:          req.TransactionData.Amount,
			Currency:        req.TransactionData.Currency,
			Timestamp:       req.TransactionData.Timestamp,
			MerchantID:      req.TransactionData.MerchantID,
			MerchantName:    req.TransactionData.MerchantName,
			MerchantCategory: req.TransactionData.MerchantCategory,
			Description:     req.TransactionData.Description,
			PaymentMethod:   req.TransactionData.PaymentMethod,
			RecipientID:     req.TransactionData.RecipientID,
			SourceAccount:   req.TransactionData.SourceAccount,
			DestinationAccount: req.TransactionData.DestinationAccount,
		}
	}

	// Converter dados comportamentais se fornecidos
	if req.BehavioralData != nil {
		internalReq.BehavioralData = &models.BehavioralData{
			SessionID:       req.BehavioralData.SessionID,
			SessionDuration: req.BehavioralData.SessionDuration,
			ClickPattern:    req.BehavioralData.ClickPattern,
			TypingSpeed:     req.BehavioralData.TypingSpeed,
			NavigationFlow:  req.BehavioralData.NavigationFlow,
			TimeOnPage:      req.BehavioralData.TimeOnPage,
			InteractionCount: req.BehavioralData.InteractionCount,
			UnusualActivity: req.BehavioralData.UnusualActivity,
			ActivityDetails: req.BehavioralData.ActivityDetails,
		}
	}

	return internalReq
}

// FromInternalModel converte um modelo interno para o modelo de API
func FromInternalModel(resp *models.AssessmentResponse) *AssessmentResponse {
	if resp == nil {
		return nil
	}
	
	// Converter status
	status := string(resp.Status)
	
	// Criar modelo de API
	apiResp := &AssessmentResponse{
		ResponseID:      resp.ResponseID,
		RequestID:       resp.RequestID,
		CorrelationID:   resp.CorrelationID,
		UserID:          resp.UserID,
		TenantID:        resp.TenantID,
		Status:          status,
		CompletedAt:     resp.CompletedAt,
		ProcessingTimeMs: resp.ProcessingTimeMs,
		TrustScore:      resp.TrustScore,
		RiskLevel:       resp.RiskLevel,
		Decision:        resp.Decision,
		Confidence:      resp.Confidence,
		RequiredActions: resp.RequiredActions,
		SuggestedActions: resp.SuggestedActions,
		Warnings:        resp.Warnings,
		DataSources:     resp.DataSources,
	}
	
	// Converter detalhes de erro se presentes
	if resp.ErrorDetails != nil {
		apiResp.ErrorDetails = &ErrorDetails{
			ErrorCode:      resp.ErrorDetails.ErrorCode,
			ErrorMessage:   resp.ErrorDetails.ErrorMessage,
			FailedServices: resp.ErrorDetails.FailedServices,
			PartialResults: resp.ErrorDetails.PartialResults,
			ErrorSource:    resp.ErrorDetails.ErrorSource,
		}
	}

	// Converter resultados detalhados para formato de API
	// (simplificados para versão de API pública)
	
	// Resultados de identidade
	if resp.IdentityResults != nil {
		apiResp.IdentityResults = map[string]interface{}{
			"identityVerified":   resp.IdentityResults.IdentityVerified,
			"verificationLevel":  resp.IdentityResults.VerificationLevel,
			"verificationScore":  resp.IdentityResults.VerificationScore,
			"verifiedAttributes": resp.IdentityResults.VerifiedAttributes,
			"dataQuality":        resp.IdentityResults.DataQuality,
		}
	}
	
	// Resultados de crédito
	if resp.CreditResults != nil {
		apiResp.CreditResults = map[string]interface{}{
			"creditScore":       resp.CreditResults.CreditScore,
			"creditRating":      resp.CreditResults.CreditRating,
			"hasPendingDebts":   resp.CreditResults.HasPendingDebts,
			"hasLegalIssues":    resp.CreditResults.HasLegalIssues,
			"isBlacklisted":     resp.CreditResults.IsBlacklisted,
			"debtToIncomeRatio": resp.CreditResults.DebtToIncomeRatio,
		}
	}
	
	// Resultados de fraude
	if resp.FraudResults != nil {
		apiResp.FraudResults = map[string]interface{}{
			"fraudDetected":    resp.FraudResults.FraudDetected,
			"fraudProbability": resp.FraudResults.FraudProbability,
			"fraudScore":       resp.FraudResults.FraudScore,
			"fraudVerdict":     resp.FraudResults.FraudVerdict,
			"deviceReputation": resp.FraudResults.DeviceReputation,
			"ipReputation":     resp.FraudResults.IPReputation,
		}
	}
	
	// Resultados de conformidade
	if resp.ComplianceResults != nil {
		apiResp.ComplianceResults = map[string]interface{}{
			"compliant":       resp.ComplianceResults.Compliant,
			"complianceScore": resp.ComplianceResults.ComplianceScore,
			"pepStatus":       resp.ComplianceResults.PEPStatus,
			"kycStatus":       resp.ComplianceResults.KYCStatus,
			"amlStatus":       resp.ComplianceResults.AMLStatus,
			"riskCategory":    resp.ComplianceResults.RiskCategory,
		}
	}
	
	// Resultados de risco
	if resp.RiskResults != nil {
		apiResp.RiskResults = map[string]interface{}{
			"riskScore":          resp.RiskResults.RiskScore,
			"riskLevel":          resp.RiskResults.RiskLevel,
			"trustLevel":         resp.RiskResults.TrustLevel,
			"allowOperation":     resp.RiskResults.AllowOperation,
			"requireAdditionalAuth": resp.RiskResults.RequireAdditionalAuth,
		}
	}
	
	return apiResp
}