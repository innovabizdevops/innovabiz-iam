/**
 * @file results_consolidation.go
 * @description Métodos para consolidação de resultados de avaliações
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package orchestration

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"innovabiz/iam/src/bureau-credito/adapters"
	"innovabiz/iam/src/bureau-credito/orchestration/models"
)

// consolidateResults consolida os resultados de todas as avaliações
func (o *BureauOrchestrator) consolidateResults(response *models.AssessmentResponse) {
	// Calcular pontuação de confiança geral
	confidence := o.calculateOverallConfidence(response)
	response.Confidence = confidence

	// Determinar nível de risco geral
	riskLevel := o.determineOverallRiskLevel(response)
	response.RiskLevel = riskLevel

	// Calcular pontuação de confiança
	trustScore := o.calculateTrustScore(response)
	response.TrustScore = trustScore

	// Determinar decisão final
	decision := o.makeFinalDecision(response)
	response.Decision = decision

	// Determinar ações necessárias e sugeridas
	requiredActions, suggestedActions := o.determineActions(response)
	response.RequiredActions = requiredActions
	response.SuggestedActions = suggestedActions
}

// consolidateCreditResults consolida os resultados de múltiplos provedores de crédito
func (o *BureauOrchestrator) consolidateCreditResults(creditResults *models.CreditResults) {
	var totalScore int
	var count int
	var hasPendingDebts bool
	var hasLegalIssues bool
	var isBlacklisted bool

	// Processar respostas de cada provedor
	for _, response := range creditResults.ProviderResponses {
		// Somar pontuações de crédito
		totalScore += response.CreditScore
		count++

		// Verificar sinalizações negativas
		if response.HasPendingDebts {
			hasPendingDebts = true
		}
		if response.HasLegalIssues {
			hasLegalIssues = true
		}
		if response.IsBlacklisted {
			isBlacklisted = true
		}
	}

	// Calcular média de pontuação de crédito
	if count > 0 {
		creditResults.CreditScore = totalScore / count
	} else {
		creditResults.CreditScore = 0
	}

	// Atribuir sinalizações consolidadas
	creditResults.HasPendingDebts = hasPendingDebts
	creditResults.HasLegalIssues = hasLegalIssues
	creditResults.IsBlacklisted = isBlacklisted

	// Determinar classificação de crédito
	creditResults.CreditRating = o.calculateCreditRating(creditResults.CreditScore)
}

// calculateCreditRating calcula a classificação de crédito com base na pontuação
func (o *BureauOrchestrator) calculateCreditRating(score int) string {
	switch {
	case score >= 900:
		return "EXCELENTE"
	case score >= 800:
		return "MUITO_BOM"
	case score >= 700:
		return "BOM"
	case score >= 600:
		return "REGULAR"
	case score >= 500:
		return "BAIXO"
	default:
		return "MUITO_BAIXO"
	}
}

// calculateOverallConfidence calcula a confiança geral na avaliação
func (o *BureauOrchestrator) calculateOverallConfidence(response *models.AssessmentResponse) float64 {
	var confidenceFactors []float64
	var weights []float64

	// Confiança na identidade
	if response.IdentityResults != nil {
		identityConfidence := response.IdentityResults.VerificationScore
		confidenceFactors = append(confidenceFactors, identityConfidence)
		weights = append(weights, 1.0) // Peso alto para verificação de identidade
	}

	// Confiança no crédito
	if response.CreditResults != nil {
		// Calcular confiança com base na pontuação de crédito e fornecedores disponíveis
		creditConfidence := float64(len(response.CreditResults.ProviderResponses)) / 3.0 // Normalizado para 1.0
		if creditConfidence > 1.0 {
			creditConfidence = 1.0
		}
		confidenceFactors = append(confidenceFactors, creditConfidence)
		weights = append(weights, 0.8)
	}

	// Confiança na detecção de fraude
	if response.FraudResults != nil {
		// Alta confiança quando a probabilidade de fraude está nos extremos (muito baixa ou muito alta)
		fraudConfidence := math.Abs(response.FraudResults.FraudProbability-50.0) / 50.0
		confidenceFactors = append(confidenceFactors, fraudConfidence)
		weights = append(weights, 0.9)
	}

	// Confiança na conformidade
	if response.ComplianceResults != nil {
		complianceConfidence := float64(response.ComplianceResults.ComplianceScore) / 100.0
		confidenceFactors = append(confidenceFactors, complianceConfidence)
		weights = append(weights, 0.7)
	}

	// Confiança na avaliação de risco
	if response.RiskResults != nil {
		riskConfidence := 1.0 - (math.Abs(response.RiskResults.RiskScore-50.0) / 50.0)
		confidenceFactors = append(confidenceFactors, riskConfidence)
		weights = append(weights, 1.0)
	}

	// Se não houver fatores, retornar confiança zero
	if len(confidenceFactors) == 0 {
		return 0.0
	}

	// Calcular média ponderada
	var weightedSum float64
	var weightSum float64

	for i, factor := range confidenceFactors {
		weightedSum += factor * weights[i]
		weightSum += weights[i]
	}

	// Calcular confiança geral (0-100%)
	if weightSum > 0 {
		return (weightedSum / weightSum) * 100.0
	}

	return 0.0
}

// determineOverallRiskLevel determina o nível de risco geral
func (o *BureauOrchestrator) determineOverallRiskLevel(response *models.AssessmentResponse) string {
	var riskFactors []int
	var weights []int

	// Riscos baseados na identidade
	if response.IdentityResults != nil {
		identityRisk := 100 - int(response.IdentityResults.VerificationScore)
		if !response.IdentityResults.IdentityVerified {
			identityRisk += 50 // Penalidade para identidade não verificada
		}
		riskFactors = append(riskFactors, identityRisk)
		weights = append(weights, 25) // Peso para identidade
	}

	// Riscos baseados no crédito
	if response.CreditResults != nil {
		creditRisk := 100 - response.CreditResults.CreditScore/10 // Normalizado para 0-100
		if response.CreditResults.IsBlacklisted {
			creditRisk += 100 // Penalidade alta para blacklisting
		}
		if response.CreditResults.HasLegalIssues {
			creditRisk += 75 // Penalidade para problemas legais
		}
		if response.CreditResults.HasPendingDebts {
			creditRisk += 50 // Penalidade para dívidas pendentes
		}
		riskFactors = append(riskFactors, creditRisk)
		weights = append(weights, 20) // Peso para crédito
	}

	// Riscos baseados em fraude
	if response.FraudResults != nil {
		fraudRisk := int(response.FraudResults.FraudProbability)
		if response.FraudResults.FraudDetected {
			fraudRisk += 50 // Penalidade para fraude detectada
		}
		riskFactors = append(riskFactors, fraudRisk)
		weights = append(weights, 30) // Peso alto para fraude
	}

	// Riscos baseados em conformidade
	if response.ComplianceResults != nil {
		complianceRisk := 100 - response.ComplianceResults.ComplianceScore
		if !response.ComplianceResults.Compliant {
			complianceRisk += 75 // Penalidade alta para não conformidade
		}
		if response.ComplianceResults.PEPStatus {
			complianceRisk += 25 // Risco adicional para PEPs
		}
		riskFactors = append(riskFactors, complianceRisk)
		weights = append(weights, 25) // Peso para conformidade
	}

	// Se já temos resultado de risco
	if response.RiskResults != nil {
		riskFactors = append(riskFactors, int(response.RiskResults.RiskScore))
		weights = append(weights, 35) // Peso maior para avaliação de risco específica
	}

	// Se não houver fatores de risco, definir como desconhecido
	if len(riskFactors) == 0 {
		return "UNKNOWN"
	}

	// Calcular pontuação de risco ponderada
	var weightedSum int
	var weightSum int

	for i, factor := range riskFactors {
		weightedSum += factor * weights[i]
		weightSum += weights[i]
	}

	var riskScore int
	if weightSum > 0 {
		riskScore = weightedSum / weightSum
	} else {
		riskScore = 50 // Valor padrão médio
	}

	// Determinar nível de risco com base na pontuação
	switch {
	case riskScore >= 80:
		return "VERY_HIGH"
	case riskScore >= 60:
		return "HIGH"
	case riskScore >= 40:
		return "MEDIUM"
	case riskScore >= 20:
		return "LOW"
	default:
		return "VERY_LOW"
	}
}

// calculateTrustScore calcula a pontuação de confiança geral (0-100)
func (o *BureauOrchestrator) calculateTrustScore(response *models.AssessmentResponse) int {
	// Se tivermos resultados de risco, usar a pontuação de risco como base
	if response.RiskResults != nil {
		// Inverter a pontuação de risco para obter a pontuação de confiança
		trustScore := 100 - int(response.RiskResults.RiskScore)
		
		// Ajustar com base em outros fatores
		if response.FraudResults != nil && response.FraudResults.FraudDetected {
			trustScore -= 40 // Redução drástica para fraude detectada
		}
		
		if response.ComplianceResults != nil && !response.ComplianceResults.Compliant {
			trustScore -= 30 // Redução para não conformidade
		}
		
		if response.IdentityResults != nil && !response.IdentityResults.IdentityVerified {
			trustScore -= 50 // Redução drástica para identidade não verificada
		}
		
		if response.CreditResults != nil && response.CreditResults.IsBlacklisted {
			trustScore -= 40 // Redução para blacklisting
		}
		
		// Garantir que a pontuação esteja no intervalo 0-100
		if trustScore < 0 {
			trustScore = 0
		}
		if trustScore > 100 {
			trustScore = 100
		}
		
		return trustScore
	}
	
	// Cálculo alternativo se não tivermos resultados específicos de risco
	// Inverter a pontuação de risco do nível geral
	var baseScore int
	switch response.RiskLevel {
	case "VERY_HIGH":
		baseScore = 10
	case "HIGH":
		baseScore = 30
	case "MEDIUM":
		baseScore = 50
	case "LOW":
		baseScore = 70
	case "VERY_LOW":
		baseScore = 90
	default:
		baseScore = 50
	}
	
	// Ajustar com base na confiança
	confidenceAdjustment := int(response.Confidence / 10.0)
	trustScore := baseScore + confidenceAdjustment - 5 // -5 para compensar o ajuste positivo
	
	// Garantir intervalo 0-100
	if trustScore < 0 {
		trustScore = 0
	}
	if trustScore > 100 {
		trustScore = 100
	}
	
	return trustScore
}

// makeFinalDecision determina a decisão final com base nos resultados
func (o *BureauOrchestrator) makeFinalDecision(response *models.AssessmentResponse) string {
	// Fatores de rejeição automática
	if response.FraudResults != nil && response.FraudResults.FraudDetected {
		return "REJECT"
	}
	
	if response.ComplianceResults != nil && !response.ComplianceResults.Compliant {
		return "REJECT"
	}
	
	if response.CreditResults != nil && response.CreditResults.IsBlacklisted {
		return "REJECT"
	}
	
	// Decisão baseada no nível de risco e pontuação de confiança
	if response.TrustScore >= 70 && (response.RiskLevel == "VERY_LOW" || response.RiskLevel == "LOW") {
		return "APPROVE"
	}
	
	if response.TrustScore < 40 || response.RiskLevel == "VERY_HIGH" {
		return "REJECT"
	}
	
	// Casos intermediários para revisão manual
	return "REVIEW"
}

// determineActions determina as ações necessárias e sugeridas
func (o *BureauOrchestrator) determineActions(response *models.AssessmentResponse) ([]string, []string) {
	var requiredActions []string
	var suggestedActions []string
	
	// Ações baseadas na identidade
	if response.IdentityResults != nil {
		if !response.IdentityResults.IdentityVerified {
			requiredActions = append(requiredActions, "VERIFY_IDENTITY")
		}
		
		if response.IdentityResults.VerificationLevel < 2 {
			suggestedActions = append(suggestedActions, "UPGRADE_IDENTITY_VERIFICATION")
		}
		
		if len(response.IdentityResults.UnverifiedAttributes) > 0 {
			suggestedActions = append(suggestedActions, fmt.Sprintf("VERIFY_ATTRIBUTES:%s", strings.Join(response.IdentityResults.UnverifiedAttributes, ",")))
		}
	}
	
	// Ações baseadas no crédito
	if response.CreditResults != nil {
		if response.CreditResults.HasPendingDebts {
			suggestedActions = append(suggestedActions, "CHECK_PENDING_DEBTS")
		}
		
		if response.CreditResults.HasLegalIssues {
			requiredActions = append(requiredActions, "REVIEW_LEGAL_ISSUES")
		}
		
		if response.CreditResults.IsBlacklisted {
			requiredActions = append(requiredActions, "VERIFY_BLACKLIST_STATUS")
		}
	}
	
	// Ações baseadas em fraude
	if response.FraudResults != nil {
		if response.FraudResults.FraudDetected {
			requiredActions = append(requiredActions, "INVESTIGATE_FRAUD")
		}
		
		if response.FraudResults.FraudProbability >= 50 {
			requiredActions = append(requiredActions, "ADDITIONAL_FRAUD_VERIFICATION")
		} else if response.FraudResults.FraudProbability >= 30 {
			suggestedActions = append(suggestedActions, "MONITOR_FOR_FRAUD")
		}
	}
	
	// Ações baseadas em conformidade
	if response.ComplianceResults != nil {
		if !response.ComplianceResults.Compliant {
			requiredActions = append(requiredActions, "RESOLVE_COMPLIANCE_ISSUES")
		}
		
		if len(response.ComplianceResults.RuleViolations) > 0 {
			requiredActions = append(requiredActions, "REVIEW_COMPLIANCE_VIOLATIONS")
		}
		
		if response.ComplianceResults.PEPStatus {
			requiredActions = append(requiredActions, "ENHANCED_DUE_DILIGENCE")
		}
	}
	
	// Ações baseadas em risco
	if response.RiskResults != nil {
		if len(response.RiskResults.RecommendedActions) > 0 {
			// Adicionar recomendações do motor de risco
			requiredActions = append(requiredActions, response.RiskResults.RecommendedActions...)
		}
		
		if response.RiskResults.RequireAdditionalAuth {
			requiredActions = append(requiredActions, "REQUIRE_ADDITIONAL_AUTHENTICATION")
		}
		
		if !response.RiskResults.AllowOperation {
			requiredActions = append(requiredActions, "BLOCK_OPERATION")
		}
	}
	
	// Adicionar ações baseadas na decisão final
	switch response.Decision {
	case "REVIEW":
		requiredActions = append(requiredActions, "MANUAL_REVIEW")
	case "REJECT":
		suggestedActions = append(suggestedActions, "PROVIDE_REJECTION_REASON")
	}
	
	// Remover duplicatas
	requiredActions = removeDuplicates(requiredActions)
	suggestedActions = removeDuplicates(suggestedActions)
	
	// Ordenar para consistência
	sort.Strings(requiredActions)
	sort.Strings(suggestedActions)
	
	return requiredActions, suggestedActions
}

// removeDuplicates remove itens duplicados de uma lista de strings
func removeDuplicates(items []string) []string {
	uniqueMap := make(map[string]bool)
	result := []string{}
	
	for _, item := range items {
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = true
			result = append(result, item)
		}
	}
	
	return result
}