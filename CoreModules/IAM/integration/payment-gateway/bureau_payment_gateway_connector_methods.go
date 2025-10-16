package paymentgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	cv "github.com/innovabizdevops/innovabiz-iam/integration/cross-verification"
)

// Continuação da implementação do BureauPaymentGatewayConnector

// Verifica os limites de transação
func (c *BureauPaymentGatewayConnector) checkTransactionLimits(ctx context.Context, req *PaymentRequest) (*LimitCheckResult, error) {
	ctx, span := c.tracer.StartSpan(ctx, "checkTransactionLimits")
	defer span.End()
	
	// Resultado padrão
	result := &LimitCheckResult{
		Allowed: true,
	}
	
	// Recuperar limites aplicáveis com base em tipo de pagamento e região
	var limitKey string
	if req.HighRiskCategory {
		limitKey = "high_risk"
	} else if req.InternationalPayment {
		limitKey = "international"
	} else if req.RecurringPayment {
		limitKey = "recurring"
	} else {
		limitKey = "standard"
	}
	
	// Adicionar região ao limitKey se existirem configurações específicas
	regionLimitKey := limitKey + "_" + req.RegionCode
	
	// Verificar limites aplicáveis
	var limit TransactionLimit
	if l, ok := c.config.TransactionLimits[regionLimitKey]; ok {
		limit = l
	} else if l, ok := c.config.TransactionLimits[limitKey]; ok {
		limit = l
	} else {
		c.logger.WarnWithContext(ctx, "Nenhum limite de transação configurado, permitindo transação",
			"limit_key", limitKey, "region", req.RegionCode)
		return result, nil
	}
	
	// Verificar limite de transação única
	if limit.SingleTransactionMax > 0 && req.Amount > limit.SingleTransactionMax {
		result.Allowed = false
		result.Reason = "single_transaction_limit_exceeded"
		result.Details = fmt.Sprintf("Valor da transação (%.2f %s) excede o limite permitido (%.2f %s)",
			req.Amount, req.Currency, limit.SingleTransactionMax, req.Currency)
		
		c.logger.InfoWithContext(ctx, "Limite de transação única excedido",
			"request_id", req.RequestID,
			"amount", req.Amount,
			"limit", limit.SingleTransactionMax)
		
		return result, nil
	}
	
	// Aqui seriam implementadas verificações adicionais para limites diários e mensais
	// Isso exigiria a busca de histórico de transações do usuário, que não está implementado neste exemplo
	
	// Verificar se é necessária verificação avançada
	if limit.RequiresEnhancedVerification {
		c.logger.InfoWithContext(ctx, "Transação requer verificação avançada",
			"request_id", req.RequestID,
			"limit_key", limitKey)
	}
	
	return result, nil
}

// Determina o nível de verificação necessário para a transação
func (c *BureauPaymentGatewayConnector) determineVerificationLevel(ctx context.Context, req *PaymentRequest) (string, []string) {
	ctx, span := c.tracer.StartSpan(ctx, "determineVerificationLevel")
	defer span.End()
	
	// Verificar configurações regionais
	var settings RegionalConnectorSettings
	var found bool
	
	if s, ok := c.config.RegionalSettings[req.RegionCode]; ok {
		settings = s
		found = true
	}
	
	// Definir nível de verificação padrão
	verificationLevel := VerificationLevelStandard
	if found && settings.RequiredVerificationLevel != "" {
		verificationLevel = settings.RequiredVerificationLevel
	}
	
	extraChecks := []string{}
	
	// Aplicar regras específicas para diferentes cenários de transação
	rules := c.config.VerificationRules
	
	// Transações de alto valor
	if isHighValueTransaction(req) {
		if rules.HighValueTransactions.RequiredVerificationLevel != "" {
			verificationLevel = getHighestVerificationLevel(verificationLevel, rules.HighValueTransactions.RequiredVerificationLevel)
		}
		extraChecks = append(extraChecks, rules.HighValueTransactions.ExtraVerifications...)
		
		c.logger.InfoWithContext(ctx, "Aplicada regra para transação de alto valor",
			"request_id", req.RequestID,
			"verification_level", verificationLevel)
	}
	
	// Contas novas (menos de 30 dias)
	if isNewAccount(req) {
		if rules.NewAccounts.RequiredVerificationLevel != "" {
			verificationLevel = getHighestVerificationLevel(verificationLevel, rules.NewAccounts.RequiredVerificationLevel)
		}
		extraChecks = append(extraChecks, rules.NewAccounts.ExtraVerifications...)
		
		c.logger.InfoWithContext(ctx, "Aplicada regra para conta nova",
			"request_id", req.RequestID,
			"verification_level", verificationLevel)
	}
	
	// Transações internacionais
	if req.InternationalPayment {
		if rules.InternationalTransactions.RequiredVerificationLevel != "" {
			verificationLevel = getHighestVerificationLevel(verificationLevel, rules.InternationalTransactions.RequiredVerificationLevel)
		}
		extraChecks = append(extraChecks, rules.InternationalTransactions.ExtraVerifications...)
		
		c.logger.InfoWithContext(ctx, "Aplicada regra para transação internacional",
			"request_id", req.RequestID,
			"verification_level", verificationLevel)
	}
	
	// Transações recorrentes
	if req.RecurringPayment {
		if rules.RecurringTransactions.RequiredVerificationLevel != "" {
			verificationLevel = getHighestVerificationLevel(verificationLevel, rules.RecurringTransactions.RequiredVerificationLevel)
		}
		extraChecks = append(extraChecks, rules.RecurringTransactions.ExtraVerifications...)
		
		c.logger.InfoWithContext(ctx, "Aplicada regra para transação recorrente",
			"request_id", req.RequestID,
			"verification_level", verificationLevel)
	}
	
	// Categorias de alto risco
	if req.HighRiskCategory || isHighRiskMerchantCategory(req.MerchantCategory) {
		if rules.HighRiskCategories.RequiredVerificationLevel != "" {
			verificationLevel = getHighestVerificationLevel(verificationLevel, rules.HighRiskCategories.RequiredVerificationLevel)
		}
		extraChecks = append(extraChecks, rules.HighRiskCategories.ExtraVerifications...)
		
		c.logger.InfoWithContext(ctx, "Aplicada regra para categoria de alto risco",
			"request_id", req.RequestID,
			"verification_level", verificationLevel,
			"merchant_category", req.MerchantCategory)
	}
	
	// Remover verificações duplicadas
	extraChecks = removeDuplicates(extraChecks)
	
	return verificationLevel, extraChecks
}

// Executa a verificação cruzada usando o orquestrador
func (c *BureauPaymentGatewayConnector) performCrossVerification(ctx context.Context, req *PaymentRequest, verificationReq *cv.CredentialFinancialVerificationRequest) (*cv.CredentialFinancialVerificationResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx, "performCrossVerification")
	defer span.End()
	
	if c.verifier == nil {
		return nil, fmt.Errorf("orquestrador de verificação cruzada não configurado")
	}
	
	c.logger.InfoWithContext(ctx, "Executando verificação cruzada",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"verification_level", verificationReq.VerificationLevel)
	
	// Executar verificação cruzada
	verificationResp, err := c.verifier.Verify(ctx, verificationReq)
	if err != nil {
		c.logger.ErrorWithContext(ctx, "Erro na verificação cruzada",
			"request_id", req.RequestID,
			"transaction_id", req.TransactionID,
			"error", err.Error())
		return nil, err
	}
	
	c.metricsRecorder.HistogramObserve("payment_verification_trust_score", float64(verificationResp.TrustScore), map[string]string{
		"verification_level": verificationReq.VerificationLevel,
		"region": req.RegionCode,
	})
	
	c.logger.InfoWithContext(ctx, "Verificação cruzada concluída com sucesso",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"trust_score", verificationResp.TrustScore,
		"status", verificationResp.Status,
		"processing_time_ms", verificationResp.ProcessingTimeMs)
	
	return verificationResp, nil
}

// Processa o resultado da verificação e determina a ação a ser tomada
func (c *BureauPaymentGatewayConnector) processVerificationResult(ctx context.Context, req *PaymentRequest, verification *cv.CredentialFinancialVerificationResponse) (*PaymentResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx, "processVerificationResult")
	defer span.End()
	
	start := time.Now()
	
	// Verificar limites mínimos de confiança para a região
	var minTrustScore int = 60 // valor padrão
	
	if rs, ok := c.config.RegionalSettings[req.RegionCode]; ok {
		minTrustScore = rs.MinTrustScore
	}
	
	// Determinar status com base na pontuação e anomalias
	status := TransactionStatusApproved
	challengeRequired := false
	statusDescription := "Transação aprovada"
	statusCode := "approved"
	
	// Processar com base no status de verificação e pontuação
	switch verification.Status {
	case cv.VerificationStatusPassed:
		// Aprovado, nenhuma ação adicional necessária
	
	case cv.VerificationStatusPartial:
		// Verificar se pontuação está acima do mínimo
		if verification.TrustScore < minTrustScore {
			// Pontuação abaixo do mínimo, exigir desafio
			status = TransactionStatusChallenged
			challengeRequired = true
			statusDescription = "Verificação adicional necessária"
			statusCode = "challenge_required"
		}
	
	case cv.VerificationStatusFailed:
		// Verificação falhou, transação recusada
		status = TransactionStatusDenied
		statusDescription = "Transação recusada devido a falha na verificação"
		statusCode = "verification_failed"
	
	default:
		// Status desconhecido ou erro
		status = TransactionStatusError
		statusDescription = "Erro no processo de verificação"
		statusCode = "verification_error"
	}
	
	// Verificar anomalias críticas
	if hasCriticalAnomaly(verification.DetectedAnomalies) {
		status = TransactionStatusDenied
		statusDescription = "Transação recusada devido a anomalias críticas detectadas"
		statusCode = "critical_anomaly"
		
		c.logger.WarnWithContext(ctx, "Anomalias críticas detectadas, transação recusada",
			"request_id", req.RequestID,
			"transaction_id", req.TransactionID,
			"anomalies", len(verification.DetectedAnomalies))
	}
	
	// Criar detalhes do desafio se necessário
	var challengeDetails *ChallengeDetails
	if challengeRequired {
		challengeDetails = c.createChallenge(ctx, req, verification)
	}
	
	// Criar resposta de pagamento
	response := &PaymentResponse{
		RequestID:          req.RequestID,
		TransactionID:      req.TransactionID,
		Status:             status,
		ProcessingTimeMs:   time.Since(start).Milliseconds(),
		TrustScore:         verification.TrustScore,
		TrustLevel:         verification.TrustLevel,
		StatusDescription:  statusDescription,
		StatusCode:         statusCode,
		ChallengeRequired:  challengeRequired,
		ChallengeDetails:   challengeDetails,
		DetectedAnomalies:  verification.DetectedAnomalies,
		RiskLevel:          getRiskLevelFromScore(verification.TrustScore),
		VerificationDetails: VerificationDetails{
			VerificationID:    verification.VerificationID,
			VerificationLevel: req.UserData.VerificationLevel,
			VerifiedFields:    getVerifiedFields(verification),
			FailedFields:      getFailedFields(verification),
			VerificationTime:  time.Now(),
			EnhancedVerification: isEnhancedVerificationLevel(req.UserData.VerificationLevel),
		},
		Timestamp: time.Now(),
	}
	
	// Adicionar código de aprovação para transações aprovadas
	if status == TransactionStatusApproved {
		response.ApprovalCode = generateApprovalCode()
		response.AuthorizationID = fmt.Sprintf("auth-%s", uuid.New().String()[0:8])
	}
	
	// Registrar métricas do resultado
	c.metricsRecorder.CounterInc("payment_transactions_total", map[string]string{
		"status": status,
		"region": req.RegionCode,
		"payment_method": req.PaymentMethod,
	})
	
	// Adicionar ao cache se habilitado
	if c.config.EnableCaching {
		c.transactionCache.Store(req.RequestID, response)
	}
	
	c.logger.InfoWithContext(ctx, "Processamento de pagamento concluído",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"status", status,
		"trust_score", verification.TrustScore,
		"processing_time_ms", response.ProcessingTimeMs)
	
	return response, nil
}

// Cria um desafio para o usuário
func (c *BureauPaymentGatewayConnector) createChallenge(ctx context.Context, req *PaymentRequest, verification *cv.CredentialFinancialVerificationResponse) *ChallengeDetails {
	// Determinar o tipo de desafio com base na pontuação de confiança
	challengeType := "otp_sms" // padrão
	
	if verification.TrustScore < 40 {
		challengeType = "multi_factor"
	} else if verification.TrustScore < 50 {
		challengeType = "biometric"
	}
	
	// Criar detalhes do desafio
	challengeID := fmt.Sprintf("chal-%s", uuid.New().String()[0:8])
	expiration := time.Now().Add(15 * time.Minute)
	
	challengeDetails := &ChallengeDetails{
		ChallengeID:     challengeID,
		ChallengeType:   challengeType,
		ChallengeMethod: mapChallengeTypeToMethod(challengeType),
		Instructions:    generateChallengeInstructions(challengeType),
		ExpirationTime:  expiration,
		RetryCount:      0,
		MaxRetries:      3,
		VerificationURL: fmt.Sprintf("/api/v1/challenges/%s/verify", challengeID),
	}
	
	c.logger.InfoWithContext(ctx, "Desafio criado para transação",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"challenge_id", challengeID,
		"challenge_type", challengeType,
		"expiration", expiration.Format(time.RFC3339))
	
	return challengeDetails
}

// Cria uma resposta de erro
func (c *BureauPaymentGatewayConnector) createErrorResponse(req *PaymentRequest, errorCode string, errorDescription string) *PaymentResponse {
	return &PaymentResponse{
		RequestID:         req.RequestID,
		TransactionID:     req.TransactionID,
		Status:            TransactionStatusError,
		StatusDescription: errorDescription,
		StatusCode:        errorCode,
		ProcessingTimeMs:  0,
		Timestamp:         time.Now(),
	}
}

// Inicia limpeza periódica do cache
func (c *BureauPaymentGatewayConnector) startCacheCleanup() {
	ticker := time.NewTicker(c.config.CacheTTL / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		expiredCount := 0
		
		c.transactionCache.Range(func(key, value interface{}) bool {
			resp, ok := value.(*PaymentResponse)
			if !ok {
				c.transactionCache.Delete(key)
				expiredCount++
				return true
			}
			
			// Verifica se o resultado está no cache há mais tempo que o TTL
			if now.Sub(resp.Timestamp) > c.config.CacheTTL {
				c.transactionCache.Delete(key)
				expiredCount++
			}
			
			return true
		})
		
		if expiredCount > 0 {
			c.logger.Debug("Limpeza de cache concluída", "expired_items", expiredCount)
		}
	}
}

// Funções auxiliares

// Verifica se uma transação é considerada de alto valor
func isHighValueTransaction(req *PaymentRequest) bool {
	// Esta seria uma implementação mais completa que consultaria
	// thresholds configurados por moeda e região
	// Simplificando para fins deste exemplo
	
	if req.Currency == "USD" {
		return req.Amount > 1000.0
	} else if req.Currency == "EUR" {
		return req.Amount > 900.0
	} else if req.Currency == "BRL" {
		return req.Amount > 5000.0
	} else if req.Currency == "AOA" {
		return req.Amount > 500000.0
	}
	
	// Valor padrão para outras moedas
	return req.Amount > 500.0
}

// Verifica se a conta do usuário é nova (menos de 30 dias)
func isNewAccount(req *PaymentRequest) bool {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return req.UserData.AccountCreated.After(thirtyDaysAgo)
}

// Verifica se a categoria do comerciante é de alto risco
func isHighRiskMerchantCategory(category string) bool {
	// Lista simplificada de categorias de alto risco
	highRiskCategories := []string{
		"gambling", "crypto", "adult", "gaming", "digital_goods",
		"travel", "money_transfer", "online_pharmacy", "betting",
	}
	
	for _, c := range highRiskCategories {
		if strings.Contains(strings.ToLower(category), c) {
			return true
		}
	}
	
	return false
}

// Retorna o nível de verificação mais alto entre dois
func getHighestVerificationLevel(level1, level2 string) string {
	levels := map[string]int{
		VerificationLevelBasic:    1,
		VerificationLevelStandard: 2,
		VerificationLevelAdvanced: 3,
		VerificationLevelPremium:  4,
	}
	
	if levels[level2] > levels[level1] {
		return level2
	}
	return level1
}

// Remove duplicatas de uma lista de strings
func removeDuplicates(items []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	
	return result
}

// Verifica se há anomalias críticas
func hasCriticalAnomaly(anomalies []cv.Anomaly) bool {
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" || anomaly.ConfidenceScore > 0.9 {
			return true
		}
	}
	return false
}

// Gera código de aprovação aleatório
func generateApprovalCode() string {
	return fmt.Sprintf("%02d%02d%02d", 
		time.Now().Nanosecond() % 100,
		time.Now().Second() % 100,
		time.Now().Minute() % 100)
}

// Mapeia tipo de desafio para método
func mapChallengeTypeToMethod(challengeType string) string {
	switch challengeType {
	case "otp_sms":
		return "sms"
	case "otp_email":
		return "email"
	case "biometric":
		return "facial_recognition"
	case "multi_factor":
		return "app_push"
	default:
		return "sms"
	}
}

// Gera instruções para desafio
func generateChallengeInstructions(challengeType string) string {
	switch challengeType {
	case "otp_sms":
		return "Um código de verificação foi enviado para seu telefone celular. Por favor, digite-o para confirmar a transação."
	case "otp_email":
		return "Um código de verificação foi enviado para seu email. Por favor, digite-o para confirmar a transação."
	case "biometric":
		return "Por favor, complete a verificação biométrica facial para confirmar sua identidade."
	case "multi_factor":
		return "Uma solicitação de aprovação foi enviada para seu aplicativo. Por favor, aprove para confirmar a transação."
	default:
		return "Por favor, complete a verificação adicional para continuar com a transação."
	}
}

// Obtém nível de risco com base na pontuação
func getRiskLevelFromScore(score int) string {
	if score >= 80 {
		return "low"
	} else if score >= 60 {
		return "medium"
	} else if score >= 40 {
		return "high"
	} else {
		return "critical"
	}
}

// Verifica se o nível de verificação é considerado avançado
func isEnhancedVerificationLevel(level string) bool {
	return level == VerificationLevelAdvanced || level == VerificationLevelPremium
}

// Extrai campos verificados da resposta de verificação
func getVerifiedFields(verification *cv.CredentialFinancialVerificationResponse) []string {
	fields := []string{}
	
	for category, result := range verification.VerificationResults {
		if result.Status == cv.VerificationStatusPassed || result.Status == cv.VerificationStatusPartial {
			for _, field := range result.VerifiedFields {
				fields = append(fields, fmt.Sprintf("%s.%s", category, field))
			}
		}
	}
	
	return fields
}

// Extrai campos que falharam da resposta de verificação
func getFailedFields(verification *cv.CredentialFinancialVerificationResponse) []string {
	fields := []string{}
	
	for category, result := range verification.VerificationResults {
		if result.Status == cv.VerificationStatusFailed || result.Status == cv.VerificationStatusPartial {
			for _, field := range result.FailedFields {
				fields = append(fields, fmt.Sprintf("%s.%s", category, field))
			}
		}
	}
	
	return fields
}