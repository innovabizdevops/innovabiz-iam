/**
 * @file serasa_adapter.go
 * @description Adaptador para integração com Serasa Experian
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SerasaAdapter implementa CreditProvider para integração com Serasa
type SerasaAdapter struct {
	config      CreditProviderConfig
	httpClient  *http.Client
	initialized bool
}

// SerasaCreditReport representa a estrutura de resposta específica do Serasa
type SerasaCreditReport struct {
	Status      string `json:"status"`
	Score       int    `json:"score"`
	ScoreRating string `json:"scoreRating"`
	
	// Informações de crédito
	CreditProfile struct {
		HasRestriction   bool    `json:"hasRestriction"`
		HasProtest       bool    `json:"hasProtest"`
		RestrictionsQty  int     `json:"restrictionsQty"`
		TotalDebtAmount  float64 `json:"totalDebtAmount"`
		YearsOfHistory   int     `json:"yearsOfHistory"`
		CreditLimit      float64 `json:"creditLimit,omitempty"`
		LastConsultation string  `json:"lastConsultation,omitempty"`
	} `json:"creditProfile"`
	
	// Análise comportamental
	BehavioralAnalysis struct {
		RiskLevel           string   `json:"riskLevel"`
		FraudProbability    float64  `json:"fraudProbability"`
		AnomalyScore        int      `json:"anomalyScore"`
		RiskFactors         []string `json:"riskFactors"`
		AuthRecommendation  string   `json:"authRecommendation"`
		RequiredFactors     []string `json:"requiredFactors,omitempty"`
	} `json:"behavioralAnalysis"`
}

// NewSerasaAdapter cria um novo adaptador para Serasa
func NewSerasaAdapter() *SerasaAdapter {
	return &SerasaAdapter{
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		initialized: false,
	}
}

// Initialize inicializa o adaptador com configurações específicas
func (s *SerasaAdapter) Initialize(config CreditProviderConfig) error {
	if config.APIKey == "" || config.APISecret == "" {
		return errors.New("APIKey e APISecret são obrigatórios para o adaptador Serasa")
	}
	
	if config.BaseURL == "" {
		config.BaseURL = "https://api.serasa.com.br/v1"
	}
	
	s.config = config
	s.httpClient.Timeout = time.Duration(config.TimeoutSeconds) * time.Second
	s.initialized = true
	
	return nil
}

// GetCreditReport obtém um relatório de crédito do Serasa
func (s *SerasaAdapter) GetCreditReport(ctx context.Context, request CreditReportRequest) (*CreditReportResponse, error) {
	if !s.initialized {
		return nil, errors.New("adaptador Serasa não inicializado")
	}
	
	startTime := time.Now()
	
	// Validar parâmetros obrigatórios
	if request.DocumentNumber == "" || request.DocumentType == "" {
		return nil, errors.New("documentNumber e documentType são obrigatórios para consulta Serasa")
	}
	
	// Preparar endpoint com base no tipo de documento
	endpoint := fmt.Sprintf("%s/credit-report", s.config.BaseURL)
	if request.DocumentType == "CNPJ" {
		endpoint = fmt.Sprintf("%s/business/credit-report", s.config.BaseURL)
	}
	
	// Fazer a requisição para o Serasa
	serasaReport, err := s.fetchSerasaReport(ctx, endpoint, request)
	if err != nil {
		return &CreditReportResponse{
			RequestID:    fmt.Sprintf("serasa-%d", time.Now().Unix()),
			ProviderName: "SERASA",
			ReportDate:   time.Now(),
			TrustLevel:   TrustLevelUnknown,
			Error: &CreditError{
				Code:      "SERASA_API_ERROR",
				Message:   "Erro ao consultar Serasa",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}
	
	// Mapear resposta do Serasa para o formato padrão
	response := mapSerasaResponse(serasaReport, request)
	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	response.ProviderName = "SERASA"
	response.RequestID = fmt.Sprintf("serasa-%d", time.Now().Unix())
	response.ReportDate = time.Now()
	
	return response, nil
}

// BatchGetCreditReports implementa consulta em lote
func (s *SerasaAdapter) BatchGetCreditReports(ctx context.Context, requests []CreditReportRequest) ([]*CreditReportResponse, error) {
	if !s.initialized {
		return nil, errors.New("adaptador Serasa não inicializado")
	}
	
	responses := make([]*CreditReportResponse, len(requests))
	
	// Para cada solicitação, obter relatório individual
	// Em uma implementação real, poderíamos otimizar com consultas em paralelo
	for i, req := range requests {
		resp, err := s.GetCreditReport(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("erro na requisição %d: %w", i, err)
		}
		responses[i] = resp
	}
	
	return responses, nil
}

// GetProviderInfo retorna informações sobre o provedor Serasa
func (s *SerasaAdapter) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		ID:                 "serasa-experian",
		Name:               "Serasa Experian",
		Description:        "Provedor brasileiro de dados de crédito e verificação de identidade",
		SupportedCountries: []string{"BR"},
		SupportedFeatures:  []string{"CREDIT_SCORE", "FRAUD_DETECTION", "IDENTITY_VERIFICATION"},
		MaxQPS:             10,
		Version:            "1.0.0",
	}
}

// IsHealthy verifica se o serviço Serasa está operacional
func (s *SerasaAdapter) IsHealthy(ctx context.Context) bool {
	if !s.initialized {
		return false
	}
	
	url := fmt.Sprintf("%s/health", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}
	
	s.addAuthHeaders(req)
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// Métodos auxiliares privados

// fetchSerasaReport faz a requisição HTTP para o Serasa
func (s *SerasaAdapter) fetchSerasaReport(ctx context.Context, endpoint string, request CreditReportRequest) (*SerasaCreditReport, error) {
	// Construir payload para a requisição
	payload := map[string]interface{}{
		"documentNumber": request.DocumentNumber,
		"documentType":   request.DocumentType,
		"requestReason":  request.RequestReason,
	}
	
	// Adicionar campos opcionais se presentes
	if request.Name != "" {
		payload["name"] = request.Name
	}
	
	// Adicionar contexto de transação se relevante
	if request.TransactionAmount > 0 {
		payload["transactionDetails"] = map[string]interface{}{
			"amount":   request.TransactionAmount,
			"currency": request.Currency,
		}
	}
	
	// Converter payload para JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %w", err)
	}
	
	// Preparar requisição HTTP
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		strings.NewReader(string(jsonPayload)),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	
	// Adicionar headers de autenticação e conteúdo
	req.Header.Set("Content-Type", "application/json")
	s.addAuthHeaders(req)
	
	// Executar requisição
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()
	
	// Verificar código de status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API do Serasa: status %d", resp.StatusCode)
	}
	
	// Decodificar resposta
	var serasaReport SerasaCreditReport
	err = json.NewDecoder(resp.Body).Decode(&serasaReport)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}
	
	return &serasaReport, nil
}

// addAuthHeaders adiciona os headers de autenticação à requisição
func (s *SerasaAdapter) addAuthHeaders(req *http.Request) {
	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("X-API-Secret", s.config.APISecret)
}

// mapSerasaResponse converte a resposta do Serasa para o formato padrão
func mapSerasaResponse(serasaReport *SerasaCreditReport, request CreditReportRequest) *CreditReportResponse {
	response := &CreditReportResponse{
		CreditScore:    serasaReport.Score,
		HasPendingDebts: serasaReport.CreditProfile.HasRestriction,
		HasLegalIssues:  serasaReport.CreditProfile.HasProtest,
		DetailedReport:  make(map[string]interface{}),
	}
	
	// Mapear nível de risco
	switch serasaReport.BehavioralAnalysis.RiskLevel {
	case "VERY_LOW":
		response.RiskAssessment = RiskVeryLow
		response.TrustLevel = TrustLevelVeryHigh
	case "LOW":
		response.RiskAssessment = RiskLow
		response.TrustLevel = TrustLevelHigh
	case "MEDIUM":
		response.RiskAssessment = RiskMedium
		response.TrustLevel = TrustLevelMedium
	case "HIGH":
		response.RiskAssessment = RiskHigh
		response.TrustLevel = TrustLevelLow
	case "VERY_HIGH":
		response.RiskAssessment = RiskVeryHigh
		response.TrustLevel = TrustLevelVeryLow
	default:
		response.RiskAssessment = RiskUnknown
		response.TrustLevel = TrustLevelUnknown
	}
	
	// Calcular score de risco baseado na probabilidade de fraude
	response.RiskScore = int(serasaReport.BehavioralAnalysis.FraudProbability * 100)
	
	// Verificar se está na lista negra com base em fatores de risco
	response.IsBlacklisted = false
	for _, factor := range serasaReport.BehavioralAnalysis.RiskFactors {
		if factor == "BLACKLISTED" || factor == "FRAUDULENT_HISTORY" {
			response.IsBlacklisted = true
			break
		}
	}
	
	// Adicionar detalhes completos do relatório
	response.DetailedReport["creditProfile"] = serasaReport.CreditProfile
	response.DetailedReport["behavioralAnalysis"] = serasaReport.BehavioralAnalysis
	response.DetailedReport["scoreRating"] = serasaReport.ScoreRating
	
	return response
}