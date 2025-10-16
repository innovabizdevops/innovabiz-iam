package main

import (
	"strings"
	"time"
)

// RemediationResult contém o resultado da aplicação de remediações
type RemediationResult struct {
	Enabled               bool                `json:"enabled"`
	Success               bool                `json:"success"`
	Timestamp             string              `json:"timestamp,omitempty"`
	Message               string              `json:"message,omitempty"`
	TotalViolations       int                 `json:"total_violations,omitempty"`
	AttemptedRemediations int                 `json:"attempted_remediations,omitempty"`
	SuccessfulRemediations int                `json:"successful_remediations,omitempty"`
	FailedRemediations    int                 `json:"failed_remediations,omitempty"`
	PolicyFilesModified   map[string][]string `json:"policy_files_modified,omitempty"`
	Errors                []RemediationError  `json:"errors,omitempty"`
	DryRun               bool                `json:"dry_run,omitempty"`
}

// RemediationError representa um erro durante a remediação
type RemediationError struct {
	PolicyFile   string `json:"policy_file"`
	RuleID       string `json:"rule_id"`
	ErrorMessage string `json:"error_message"`
}

// ComplianceReport representa o relatório completo de compliance para uma região
type ComplianceReport struct {
	Region           string             `json:"region"`
	Timestamp        time.Time          `json:"timestamp"`
	Passed           bool               `json:"passed"`
	Score            float64            `json:"score"`
	FrameworkResults []FrameworkResult  `json:"framework_results"`
	RemediationApplied bool             `json:"remediation_applied,omitempty"`
	RemediationSummary RemediationSummary `json:"remediation_summary,omitempty"`
}

// FrameworkResult representa os resultados dos testes para um framework específico
type FrameworkResult struct {
	FrameworkID    string       `json:"framework_id"`
	FrameworkName  string       `json:"framework_name"`
	Passed         bool         `json:"passed"`
	Score          float64      `json:"score"`
	TestResults    []TestResult `json:"test_results"`
}

// TestResult representa o resultado de um teste de conformidade individual
// Esta é uma extensão da estrutura TestResult existente com campos adicionais para remediação
type TestResult struct {
	TestID          string      `json:"test_id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Passed          bool        `json:"passed"`
	PolicyFile      string      `json:"policy_file"`
	ExpectedDecision interface{} `json:"expected_decision"`
	ActualDecision   interface{} `json:"actual_decision"`
	Criticality     string      `json:"criticality"`
	ViolationType   string      `json:"violation_type,omitempty"`
	Tags            []string    `json:"tags,omitempty"`
	RequirementIDs  []string    `json:"requirement_ids,omitempty"`
}

// ConversionTestSummaryToComplianceReport converte um TestSummary para o formato ComplianceReport
func ConversionTestSummaryToComplianceReport(summary *TestSummary) *ComplianceReport {
	report := &ComplianceReport{
		Region:    summary.Region,
		Timestamp: summary.ExecutedAt,
		Passed:    summary.FailedTests == 0,
		Score:     summary.ComplianceScore,
	}

	// Converter resultados por framework
	frameworkResults := make(map[string]*FrameworkResult)
	
	// Primeiro, criar estrutura básica para cada framework
	for frameworkID, frameworkScore := range summary.FrameworkScores {
		frameworkResults[frameworkID] = &FrameworkResult{
			FrameworkID:   frameworkID,
			FrameworkName: frameworkScore.Name,
			Passed:        frameworkScore.FailedTests == 0,
			Score:         frameworkScore.ComplianceScore,
		}
	}
	
	// Adicionar resultados de testes aos frameworks correspondentes
	for _, testResult := range summary.TestResults {
		// Ignorar testes bem-sucedidos, pois não precisam de remediação
		if testResult.Passed {
			continue
		}
		
		// Para cada framework associado ao teste
		for _, frameworkID := range testResult.Frameworks {
			if frameworkResult, exists := frameworkResults[frameworkID]; exists {
				// Inferir tipo de violação se não estiver explicitamente definido
				violationType := testResult.ViolationType
				if violationType == "" {
					violationType = determineViolationType(*testResult)
				}
				
				// Criar o resultado do teste com informações necessárias para remediação
				result := TestResult{
					TestID:          testResult.TestCase.ID,
					Name:            testResult.TestCase.Name,
					Description:     testResult.TestCase.Description,
					Passed:          false,
					PolicyFile:      testResult.PolicyPath,
					ExpectedDecision: testResult.TestCase.ExpectedDecision,
					ActualDecision:   testResult.ActualDecision,
					Criticality:     testResult.Criticality,
					ViolationType:   violationType,
					Tags:            testResult.Tags,
					RequirementIDs:  testResult.TestCase.RequirementIDs,
				}
				
				frameworkResult.TestResults = append(frameworkResult.TestResults, result)
			}
		}
	}
	
	// Converter o mapa em slice
	for _, frameworkResult := range frameworkResults {
		report.FrameworkResults = append(report.FrameworkResults, *frameworkResult)
	}
	
	return report
}

// determineViolationType infere o tipo de violação com base nas informações do teste
func determineViolationType(testResult TestResult) string {
	// Analisar tags para identificar o tipo
	for _, tag := range testResult.Tags {
		tag = strings.ToLower(tag)
		
		// Mapear tags comuns para tipos de violação
		switch {
		case strings.Contains(tag, "missing") || strings.Contains(tag, "ausente"):
			return "missing_policy"
		case strings.Contains(tag, "incorrect") || strings.Contains(tag, "wrong") || strings.Contains(tag, "incorreto"):
			return "incorrect_implementation"
		case strings.Contains(tag, "incomplete") || strings.Contains(tag, "incompleto"):
			return "incomplete_implementation"
		case strings.Contains(tag, "permission") || strings.Contains(tag, "permissão"):
			return "permission_issue"
		case strings.Contains(tag, "data") || strings.Contains(tag, "dados"):
			return "data_handling"
		case strings.Contains(tag, "outdated") || strings.Contains(tag, "desatualizado"):
			return "outdated_policy"
		case strings.Contains(tag, "config") || strings.Contains(tag, "configuração"):
			return "configuration_error"
		}
	}
	
	// Analisar nome e descrição para inferir o tipo
	testName := strings.ToLower(testResult.Name)
	testDesc := ""
	if testResult.Description != "" {
		testDesc = strings.ToLower(testResult.Description)
	}
	
	// Padrões comuns em nomes de testes
	switch {
	case strings.Contains(testName, "permiss") || strings.Contains(testDesc, "permiss"):
		return "permission_issue"
	case strings.Contains(testName, "validat") || strings.Contains(testDesc, "validat") || strings.Contains(testName, "verificar"):
		return "validation_error"
	case strings.Contains(testName, "log") || strings.Contains(testDesc, "log"):
		return "logging_issue"
	case strings.Contains(testName, "report") || strings.Contains(testDesc, "report") || strings.Contains(testName, "relatorio"):
		return "reporting_issue"
	case strings.Contains(testName, "auth") || strings.Contains(testDesc, "auth") || strings.Contains(testName, "autent"):
		return "authentication_issue"
	case strings.Contains(testName, "data") || strings.Contains(testDesc, "data") || strings.Contains(testName, "dado"):
		return "data_handling"
	}
	
	// Severidade pode dar pistas sobre o tipo de violação
	if testResult.Criticality != "" {
		switch strings.ToLower(testResult.Criticality) {
		case "alta", "high", "critical":
			return "critical_policy_violation"
		case "media", "medium":
			return "standard_policy_violation"
		case "baixa", "low":
			return "minor_policy_violation"
		}
	}
	
	// Tipo padrão se não for possível inferir
	return "general_policy_violation"
}