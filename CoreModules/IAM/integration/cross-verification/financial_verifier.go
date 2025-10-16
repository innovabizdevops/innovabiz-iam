package crossverification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

const (
	// Categoria do verificador
	CategoryFinancial = "financial"
	
	// Peso padrão da categoria na pontuação geral
	DefaultFinancialWeight = 25
)

// FinancialVerifierConfig contém configurações para o verificador financeiro
type FinancialVerifierConfig struct {
	Weight                 int                `json:"weight"`
	RequiredFields         []string           `json:"required_fields"`
	MinCreditScore         int                `json:"min_credit_score"`
	MaxDebtToIncomeRatio   float64            `json:"max_debt_to_income_ratio"`
	MinAccountAgeDays      int                `json:"min_account_age_days"`
	HighRiskFlags          []string           `json:"high_risk_flags"`
	RegionalSettings       map[string]FinancialRegionalSettings `json:"regional_settings"`
}

// FinancialRegionalSettings contém configurações regionais para verificação financeira
type FinancialRegionalSettings struct {
	MinAcceptableScore     int                `json:"min_acceptable_score"`
	MinCreditScore         int                `json:"min_credit_score"`
	MaxDebtToIncomeRatio   float64            `json:"max_debt_to_income_ratio"`
	RequiredProducts       []string           `json:"required_products"`
	HighRiskCategories     map[string]float64 `json:"high_risk_categories"`
}

// FinancialVerifier implementa a verificação de dados financeiros
type FinancialVerifier struct {
	config FinancialVerifierConfig
	logger logging.Logger
	tracer tracing.Tracer
}

// NewFinancialVerifier cria uma nova instância de verificador financeiro
func NewFinancialVerifier(config FinancialVerifierConfig, logger logging.Logger, tracer tracing.Tracer) *FinancialVerifier {
	// Configurar peso padrão se não especificado
	if config.Weight <= 0 {
		config.Weight = DefaultFinancialWeight
	}
	
	return &FinancialVerifier{
		config: config,
		logger: logger,
		tracer: tracer,
	}
}

// Verify implementa a interface CategoryVerifier
func (v *FinancialVerifier) Verify(ctx context.Context, req *CredentialFinancialVerificationRequest) (*VerificationResult, error) {
	ctx, span := v.tracer.StartSpan(ctx, "FinancialVerifier.Verify")
	defer span.End()
	
	v.logger.InfoWithContext(ctx, "Iniciando verificação financeira",
		"request_id", req.RequestID,
		"user_id", req.UserID)
		
	// Resultado inicial da verificação
	result := &VerificationResult{
		Category:       CategoryFinancial,
		Status:         VerificationStatusPending,
		Score:          0,
		Description:    "Verificação de dados financeiros e credenciais",
		VerifiedFields: []string{},
		FailedFields:   []string{},
		Details:        make(map[string]interface{}),
	}
	
	// Verificar campos obrigatórios
	missingFields := v.checkRequiredFields(req)
	if len(missingFields) > 0 {
		result.Status = VerificationStatusFailed
		result.Description = fmt.Sprintf("Campos obrigatórios financeiros ausentes: %s", strings.Join(missingFields, ", "))
		result.FailedFields = missingFields
		return result, nil
	}
	
	// Recuperar configurações regionais específicas
	regionalSettings := v.getRegionalSettings(req.RegionCode)
	
	// Verificações financeiras a serem realizadas
	checks := []struct {
		name     string
		checkFn  func(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string)
		weight   int
	}{
		{
			name:    "credit_score",
			checkFn: v.verifyCreditScore,
			weight:  20,
		},
		{
			name:    "debt_to_income",
			checkFn: v.verifyDebtToIncome,
			weight:  15,
		},
		{
			name:    "account_history",
			checkFn: v.verifyAccountHistory,
			weight:  15,
		},
		{
			name:    "payment_history",
			checkFn: v.verifyPaymentHistory,
			weight:  20,
		},
		{
			name:    "income_verification",
			checkFn: v.verifyIncome,
			weight:  15,
		},
		{
			name:    "high_risk_indicators",
			checkFn: v.checkHighRiskIndicators,
			weight:  15,
		},
	}
	
	// Executar verificações
	totalWeight := 0
	weightedScore := 0
	
	for _, check := range checks {
		passed, detail := check.checkFn(req, regionalSettings)
		
		if passed {
			result.VerifiedFields = append(result.VerifiedFields, check.name)
			weightedScore += check.weight
		} else {
			result.FailedFields = append(result.FailedFields, check.name)
			result.Details[check.name+"_details"] = detail
		}
		
		totalWeight += check.weight
	}
	
	// Calcular pontuação final
	if totalWeight > 0 {
		result.Score = (weightedScore * 100) / totalWeight
	}
	
	// Determinar status com base na pontuação e configurações regionais
	if result.Score >= regionalSettings.MinAcceptableScore {
		result.Status = VerificationStatusPassed
		result.Description = fmt.Sprintf("Verificação financeira concluída com pontuação %d/100", result.Score)
	} else if result.Score >= regionalSettings.MinAcceptableScore/2 {
		result.Status = VerificationStatusPartial
		result.Description = fmt.Sprintf("Verificação financeira parcial com pontuação %d/100", result.Score)
	} else {
		result.Status = VerificationStatusFailed
		result.Description = fmt.Sprintf("Verificação financeira falhou com pontuação %d/100", result.Score)
	}
	
	v.logger.InfoWithContext(ctx, "Verificação financeira concluída",
		"request_id", req.RequestID,
		"status", result.Status,
		"score", result.Score,
		"verified_fields", len(result.VerifiedFields),
		"failed_fields", len(result.FailedFields))
		
	return result, nil
}

// GetCategory implementa a interface CategoryVerifier
func (v *FinancialVerifier) GetCategory() string {
	return CategoryFinancial
}

// GetWeight implementa a interface CategoryVerifier
func (v *FinancialVerifier) GetWeight() int {
	return v.config.Weight
}

// Verifica se todos os campos obrigatórios estão presentes
func (v *FinancialVerifier) checkRequiredFields(req *CredentialFinancialVerificationRequest) []string {
	missing := []string{}
	
	// Verifica campos gerais obrigatórios
	for _, field := range v.config.RequiredFields {
		switch field {
		case "account_details":
			if len(req.FinancialData.AccountDetails) == 0 {
				missing = append(missing, field)
			}
		case "credit_score":
			if req.FinancialData.CreditScore == 0 {
				missing = append(missing, field)
			}
		case "financial_profile":
			if req.FinancialData.FinancialProfile.MonthlyIncome == 0 && 
			   req.FinancialData.FinancialProfile.DebtToIncomeRatio == 0 {
				missing = append(missing, field)
			}
		}
	}
	
	// Verificar produtos financeiros requeridos pela região
	regionalSettings := v.getRegionalSettings(req.RegionCode)
	for _, product := range regionalSettings.RequiredProducts {
		found := false
		for _, userProduct := range req.FinancialProducts {
			if product == userProduct {
				found = true
				break
			}
		}
		
		if !found {
			missing = append(missing, "product_"+product)
		}
	}
	
	return missing
}

// Recupera as configurações regionais ou usa padrões
func (v *FinancialVerifier) getRegionalSettings(regionCode string) FinancialRegionalSettings {
	if rs, ok := v.config.RegionalSettings[regionCode]; ok {
		return rs
	}
	
	// Configurações padrão se específicas não encontradas
	return FinancialRegionalSettings{
		MinAcceptableScore:   60,
		MinCreditScore:       v.config.MinCreditScore,
		MaxDebtToIncomeRatio: v.config.MaxDebtToIncomeRatio,
		RequiredProducts:     []string{},
		HighRiskCategories:   map[string]float64{},
	}
}

// Verifica o score de crédito
func (v *FinancialVerifier) verifyCreditScore(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	creditScore := req.FinancialData.CreditScore
	
	// Usa configuração regional ou padrão
	minScore := settings.MinCreditScore
	if minScore == 0 {
		minScore = v.config.MinCreditScore
	}
	
	if creditScore >= minScore {
		return true, fmt.Sprintf("Score de crédito adequado: %d (mínimo: %d)", creditScore, minScore)
	}
	
	return false, fmt.Sprintf("Score de crédito abaixo do mínimo: %d (mínimo: %d)", creditScore, minScore)
}

// Verifica a relação dívida/renda
func (v *FinancialVerifier) verifyDebtToIncome(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	ratio := req.FinancialData.FinancialProfile.DebtToIncomeRatio
	
	// Se a informação não estiver disponível
	if ratio == 0 {
		return true, "Relação dívida/renda não disponível, ignorando verificação"
	}
	
	// Usa configuração regional ou padrão
	maxRatio := settings.MaxDebtToIncomeRatio
	if maxRatio == 0 {
		maxRatio = v.config.MaxDebtToIncomeRatio
	}
	
	if ratio <= maxRatio {
		return true, fmt.Sprintf("Relação dívida/renda adequada: %.2f (máximo: %.2f)", ratio, maxRatio)
	}
	
	return false, fmt.Sprintf("Relação dívida/renda acima do máximo: %.2f (máximo: %.2f)", ratio, maxRatio)
}

// Verifica o histórico de contas
func (v *FinancialVerifier) verifyAccountHistory(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	// Verificar idade mínima da conta
	youngestAccount := 9999999
	oldestAccount := 0
	accountCount := len(req.FinancialData.AccountDetails)
	
	if accountCount == 0 {
		return false, "Nenhuma conta financeira encontrada"
	}
	
	activeCounts := 0
	for _, account := range req.FinancialData.AccountDetails {
		if account.AccountAgeDays < youngestAccount {
			youngestAccount = account.AccountAgeDays
		}
		
		if account.AccountAgeDays > oldestAccount {
			oldestAccount = account.AccountAgeDays
		}
		
		if account.AccountStatus == "active" || account.AccountStatus == "open" {
			activeCounts++
		}
	}
	
	// Verifica se pelo menos uma conta está ativa
	if activeCounts == 0 {
		return false, "Nenhuma conta ativa encontrada"
	}
	
	// Verifica idade mínima de conta
	if youngestAccount < v.config.MinAccountAgeDays {
		return false, fmt.Sprintf("Conta mais recente tem apenas %d dias (mínimo: %d dias)", 
			youngestAccount, v.config.MinAccountAgeDays)
	}
	
	return true, fmt.Sprintf("Histórico de conta adequado: %d contas, mais antiga com %d dias", 
		accountCount, oldestAccount)
}

// Verifica o histórico de pagamentos
func (v *FinancialVerifier) verifyPaymentHistory(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	paymentHistory := req.FinancialData.PaymentHistory
	
	if len(paymentHistory) == 0 {
		return true, "Histórico de pagamentos não disponível, ignorando verificação"
	}
	
	// Analisa histórico de pagamentos
	totalPayments := len(paymentHistory)
	failedPayments := 0
	
	for _, payment := range paymentHistory {
		if payment.Status == "failed" || payment.Status == "rejected" || payment.Status == "reversed" {
			failedPayments++
		}
	}
	
	// Calcula taxa de falha
	failRate := float64(failedPayments) / float64(totalPayments)
	
	// Aceita até 10% de falhas
	if failRate <= 0.1 {
		return true, fmt.Sprintf("Histórico de pagamentos adequado: %.1f%% de falhas em %d pagamentos", 
			failRate*100, totalPayments)
	}
	
	return false, fmt.Sprintf("Taxa de falha de pagamentos elevada: %.1f%% (%d de %d pagamentos)", 
		failRate*100, failedPayments, totalPayments)
}

// Verifica a renda
func (v *FinancialVerifier) verifyIncome(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	incomeVerification := req.FinancialData.IncomeVerification
	
	// Se a informação não estiver disponível
	if incomeVerification.DeclaredIncome == 0 && incomeVerification.VerifiedIncome == 0 {
		return true, "Verificação de renda não disponível, ignorando verificação"
	}
	
	// Se a verificação foi realizada
	if incomeVerification.Verified {
		// Verifica discrepância entre renda declarada e verificada
		if incomeVerification.DeclaredIncome > 0 && incomeVerification.VerifiedIncome > 0 {
			ratio := incomeVerification.VerifiedIncome / incomeVerification.DeclaredIncome
			
			// Aceita até 20% de diferença
			if ratio >= 0.8 && ratio <= 1.2 {
				return true, fmt.Sprintf("Renda verificada coerente com declarada (%.1f%% de correspondência)", ratio*100)
			}
			
			return false, fmt.Sprintf("Discrepância significativa entre renda verificada e declarada (%.1f%% de correspondência)",
				ratio*100)
		}
		
		return true, fmt.Sprintf("Renda verificada positivamente via %s", incomeVerification.VerificationMethod)
	}
	
	return false, "Renda não verificada"
}

// Verifica indicadores de alto risco
func (v *FinancialVerifier) checkHighRiskIndicators(req *CredentialFinancialVerificationRequest, settings FinancialRegionalSettings) (bool, string) {
	highRiskCount := 0
	highRiskDetails := []string{}
	
	// Verificar padrões de alto risco nos eventos de crédito
	for _, event := range req.FinancialData.CreditHistory {
		for _, flag := range v.config.HighRiskFlags {
			if strings.Contains(strings.ToLower(event.EventType), strings.ToLower(flag)) ||
			   strings.Contains(strings.ToLower(event.Description), strings.ToLower(flag)) {
				highRiskCount++
				highRiskDetails = append(highRiskDetails, 
					fmt.Sprintf("%s: %s (%s)", event.EventType, event.Description, event.EventDate))
				break
			}
		}
	}
	
	// Verificar categorias de alto risco nas transações
	totalHighRiskValue := 0.0
	for category, threshold := range settings.HighRiskCategories {
		if value, ok := req.FinancialData.FinancialProfile.SpendingCategories[category]; ok {
			if value > threshold {
				highRiskCount++
				totalHighRiskValue += value
				highRiskDetails = append(highRiskDetails, 
					fmt.Sprintf("Categoria de alto risco: %s (%.2f, limite: %.2f)", category, value, threshold))
			}
		}
	}
	
	// Resultado baseado na presença de indicadores de alto risco
	if highRiskCount == 0 {
		return true, "Nenhum indicador de alto risco detectado"
	} else if highRiskCount <= 2 {
		result := &VerificationResult{
			Status:     VerificationStatusPartial,
			Description: fmt.Sprintf("%d indicadores de risco de baixa severidade detectados", highRiskCount),
		}
		return false, fmt.Sprintf("Indicadores de risco identificados: %s", strings.Join(highRiskDetails, "; "))
	}
	
	return false, fmt.Sprintf("%d indicadores de alto risco identificados: %s", 
		highRiskCount, strings.Join(highRiskDetails, "; "))
}