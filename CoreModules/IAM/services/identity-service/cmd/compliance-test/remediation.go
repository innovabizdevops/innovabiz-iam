package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/innovabizdevops/innovabiz-iam/remediator"
	"github.com/sirupsen/logrus"
)

// RemediationConfig define as configurações para o processo de remediação
type RemediationConfig struct {
	// Determina se a remediação deve ser executada
	Enabled bool
	// Modo simulação (não aplica mudanças reais)
	DryRun bool
	// Aplica apenas remediações abaixo desta severidade
	MaxSeverity string
	// Frameworks para aplicar remediação
	Frameworks []string
	// Caminho para as regras de remediação
	RulesPath string
	// Diretório para armazenar backups
	BackupDir string
	// Ignorar determinados tipos de violação
	IgnoreTypes []string
	// Severidade mínima para aplicação de remediação
	MinSeverity string
	// Número máximo de remediações por política
	MaxRemediationsPerPolicy int
	// Requer aprovação manual antes de aplicar
	RequireApproval bool
}

// InitRemediationConfig inicializa a configuração de remediação com valores padrão
func InitRemediationConfig() RemediationConfig {
	return RemediationConfig{
		Enabled:                 false,
		DryRun:                  true, // Por padrão, sempre em modo simulação
		MaxSeverity:             "alta",
		RulesPath:               filepath.Join("remediator", "rules"),
		BackupDir:               filepath.Join("remediator", "backups"),
		MinSeverity:             "baixa",
		MaxRemediationsPerPolicy: 5,
		RequireApproval:         true,
	}
}

// ApplyRemediations executa o processo de remediação com base nos resultados dos testes
func ApplyRemediations(ctx context.Context, testResults *ComplianceReport, config RemediationConfig) (*RemediationSummary, error) {
	if !config.Enabled {
		logger.Info("A remediação automática está desabilitada")
		return &RemediationSummary{Enabled: false}, nil
	}

	logger.WithFields(logrus.Fields{
		"dry_run":      config.DryRun,
		"max_severity": config.MaxSeverity,
		"min_severity": config.MinSeverity,
	}).Info("Iniciando processo de remediação automática")

	// Carregar regras de remediação específicas para a região
	rulesFiles, err := findRegionalRemediationRules(config.RulesPath, testResults.Region)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar regras de remediação: %v", err)
	}

	if len(rulesFiles) == 0 {
		logger.Warn("Nenhuma regra de remediação encontrada para a região: ", testResults.Region)
		return &RemediationSummary{
			Enabled:    true,
			Successful: false,
			Message:    fmt.Sprintf("Nenhuma regra de remediação encontrada para a região: %s", testResults.Region),
		}, nil
	}

	// Carregar e consolidar todas as regras encontradas
	allRules, err := loadRemediationRules(rulesFiles)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar regras de remediação: %v", err)
	}

	// Filtrar regras por frameworks, se especificado
	var filteredRules []remediator.RemediationRule
	if len(config.Frameworks) > 0 {
		for _, rule := range allRules {
			for _, framework := range config.Frameworks {
				if rule.Framework == framework {
					filteredRules = append(filteredRules, rule)
					break
				}
			}
		}
	} else {
		filteredRules = allRules
	}

	// Inicializar o remediador
	remedConfig := remediator.Config{
		DryRun:                  config.DryRun,
		BackupDir:               config.BackupDir,
		IgnoreViolationTypes:    config.IgnoreTypes,
		MinSeverity:             config.MinSeverity,
		MaxSeverity:             config.MaxSeverity,
		MaxRemediationsPerPolicy: config.MaxRemediationsPerPolicy,
	}

	remedEngine, err := remediator.New(remedConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar o remediador: %v", err)
	}

	// Preparar violações com base nos testes falhos
	violations := prepareViolationsFromTestResults(testResults)

	// Se não houver violações, não há nada a remediar
	if len(violations) == 0 {
		logger.Info("Nenhuma violação encontrada que necessite remediação")
		return &RemediationSummary{
			Enabled:    true,
			Successful: true,
			Message:    "Nenhuma violação encontrada que necessite remediação",
		}, nil
	}

	// Solicitar confirmação do usuário se necessário
	if !config.DryRun && config.RequireApproval {
		fmt.Printf("\n%s Foram encontradas %d violações para remediação. Deseja prosseguir? [S/n]: ",
			color.YellowString("⚠️"),
			len(violations))
		
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "s" && response != "" {
			return &RemediationSummary{
				Enabled:    true,
				Successful: false,
				Message:    "Remediação cancelada pelo usuário",
			}, nil
		}
	}

	// Executar remediação
	results, err := remedEngine.RemediateViolations(ctx, filteredRules, violations)
	if err != nil {
		return nil, fmt.Errorf("erro durante o processo de remediação: %v", err)
	}

	// Preparar relatório de remediação
	summary := &RemediationSummary{
		Enabled:              true,
		Successful:           true,
		Timestamp:            time.Now().Format(time.RFC3339),
		TotalViolations:      len(violations),
		AttemptedRemediations: len(results),
		SuccessfulRemediations: 0,
		FailedRemediations:    0,
		PolicyFilesModified:   make(map[string][]string),
		DryRun:               config.DryRun,
	}

	// Contabilizar sucessos e falhas
	for _, result := range results {
		if result.Success {
			summary.SuccessfulRemediations++
			
			// Registrar as políticas modificadas e quais regras foram aplicadas
			if _, exists := summary.PolicyFilesModified[result.PolicyFile]; !exists {
				summary.PolicyFilesModified[result.PolicyFile] = []string{}
			}
			summary.PolicyFilesModified[result.PolicyFile] = append(
				summary.PolicyFilesModified[result.PolicyFile], 
				result.RuleID)
		} else {
			summary.FailedRemediations++
			summary.Errors = append(summary.Errors, RemediationError{
				PolicyFile:   result.PolicyFile,
				RuleID:       result.RuleID,
				ErrorMessage: result.Error,
			})
		}
	}

	// Exibir resumo da remediação
	if config.DryRun {
		fmt.Printf("\n%s %s\n", 
			color.BlueString("ℹ️"), 
			color.BlueString("Remediação executada em modo simulação (dry-run)"))
	} else {
		if summary.SuccessfulRemediations > 0 {
			fmt.Printf("\n%s %s\n", 
				color.GreenString("✅"), 
				color.GreenString("Remediação concluída com sucesso"))
		}
	}

	fmt.Printf("   %s Tentativas de remediação: %d\n", 
		color.WhiteString("•"),
		summary.AttemptedRemediations)
	
	fmt.Printf("   %s Remediações bem-sucedidas: %s\n", 
		color.WhiteString("•"),
		color.GreenString("%d", summary.SuccessfulRemediations))
	
	fmt.Printf("   %s Remediações falhas: %s\n", 
		color.WhiteString("•"),
		summary.FailedRemediations > 0 ? color.RedString("%d", summary.FailedRemediations) : color.GreenString("%d", summary.FailedRemediations))

	// Detalhes das políticas modificadas
	if len(summary.PolicyFilesModified) > 0 && !config.DryRun {
		fmt.Printf("\n%s %s\n", 
			color.WhiteString("📄"),
			color.WhiteString("Arquivos de política modificados:"))
		
		for file, rules := range summary.PolicyFilesModified {
			fmt.Printf("   %s %s (%d regras aplicadas)\n", 
				color.WhiteString("•"),
				file, len(rules))
		}
	}

	return summary, nil
}

// RemediationSummary contém o resumo das ações de remediação
type RemediationSummary struct {
	Enabled               bool                `json:"enabled"`
	Successful            bool                `json:"successful"`
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

// findRegionalRemediationRules busca arquivos de regras de remediação para uma região específica
func findRegionalRemediationRules(rulesPath, region string) ([]string, error) {
	var rulesFiles []string

	// Verificar se o diretório existe
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("diretório de regras não encontrado: %s", rulesPath)
	}

	// Ler todos os arquivos no diretório
	files, err := os.ReadDir(rulesPath)
	if err != nil {
		return nil, err
	}

	// Filtrar arquivos que correspondem ao padrão regional
	regionLower := strings.ToLower(region)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			fileName := strings.ToLower(file.Name())
			// Verificar se o nome do arquivo contém o código da região
			if strings.Contains(fileName, regionLower) {
				rulesFiles = append(rulesFiles, filepath.Join(rulesPath, file.Name()))
			}
		}
	}

	return rulesFiles, nil
}

// loadRemediationRules carrega todas as regras de remediação dos arquivos especificados
func loadRemediationRules(rulesFiles []string) ([]remediator.RemediationRule, error) {
	var allRules []remediator.RemediationRule

	for _, filePath := range rulesFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler arquivo de regras %s: %v", filePath, err)
		}

		var rules []remediator.RemediationRule
		if err := json.Unmarshal(data, &rules); err != nil {
			return nil, fmt.Errorf("erro ao processar arquivo de regras %s: %v", filePath, err)
		}

		allRules = append(allRules, rules...)
		logger.WithField("file", filePath).Infof("Carregadas %d regras de remediação", len(rules))
	}

	return allRules, nil
}

// prepareViolationsFromTestResults converte resultados de teste falhos em violações para remediação
func prepareViolationsFromTestResults(testResults *ComplianceReport) []remediator.PolicyViolation {
	var violations []remediator.PolicyViolation

	for _, frameworkResult := range testResults.FrameworkResults {
		for _, testResult := range frameworkResult.TestResults {
			if !testResult.Passed {
				// Determinar o tipo de violação com base no teste falho
				violationType := determineViolationType(testResult)
				
				// Se não foi possível determinar o tipo de violação, pular
				if violationType == "" {
					continue
				}

				// Obter caminho do arquivo de política
				policyFile := filepath.Join("policies", testResult.PolicyFile)
				
				// Verificar se o arquivo existe
				if _, err := os.Stat(policyFile); os.IsNotExist(err) {
					logger.WithField("policy_file", policyFile).Warn("Arquivo de política não encontrado, pulando violação")
					continue
				}

				// Adicionar violação à lista
				violation := remediator.PolicyViolation{
					PolicyFile:    policyFile,
					ViolationType: violationType,
					Framework:     frameworkResult.FrameworkID,
					TestCase:      testResult.TestID,
					Severity:      testResult.Criticality,
				}
				
				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// determineViolationType infere o tipo de violação a partir do resultado do teste falho
func determineViolationType(testResult TestResult) string {
	// Se o teste já tiver informado o tipo de violação explicitamente
	if testResult.ViolationType != "" {
		return testResult.ViolationType
	}
	
	// Tentar inferir a partir dos detalhes do teste
	// Analisar as decisões esperadas
	expected := testResult.ExpectedDecision
	actual := testResult.ActualDecision
	
	// Converter para mapas para facilitar a comparação
	expectedMap := make(map[string]interface{})
	actualMap := make(map[string]interface{})
	
	// Converter expected para map (se for possível)
	if expectedBytes, err := json.Marshal(expected); err == nil {
		json.Unmarshal(expectedBytes, &expectedMap)
	}
	
	// Converter actual para map (se for possível)
	if actualBytes, err := json.Marshal(actual); err == nil {
		json.Unmarshal(actualBytes, &actualMap)
	}
	
	// Padrões comuns de violação baseados nas saídas
	// Verifica se tem campo reasons no actual ou expected
	var reasons []string
	if r, ok := actualMap["reasons"].([]interface{}); ok {
		for _, reason := range r {
			reasons = append(reasons, fmt.Sprintf("%v", reason))
		}
	} else if r, ok := expectedMap["reasons"].([]interface{}); ok {
		for _, reason := range r {
			reasons = append(reasons, fmt.Sprintf("%v", reason))
		}
	}

	// Determinar tipo de violação baseado nos motivos
	for _, reason := range reasons {
		switch {
		case strings.Contains(reason, "consent"):
			return "missing_explicit_consent"
		case strings.Contains(reason, "strong_auth") || strings.Contains(reason, "authentication"):
			return "missing_strong_auth_check"
		case strings.Contains(reason, "pep"):
			return "missing_pep_check"
		case strings.Contains(reason, "retention"):
			return "missing_retention_limit_check"
		case strings.Contains(reason, "geo") || strings.Contains(reason, "country"):
			return "missing_geo_restriction"
		case strings.Contains(reason, "audit"):
			return "missing_audit_trail"
		case strings.Contains(reason, "aml") || strings.Contains(reason, "transaction_limit"):
			return "missing_aml_transaction_limit_check"
		case strings.Contains(reason, "personal_data"):
			return "missing_personal_data_authorization"
		case strings.Contains(reason, "multicaixa") || strings.Contains(reason, "payment"):
			return "missing_multicaixa_validation"
		case strings.Contains(reason, "age"):
			return "missing_age_verification"
		}
	}
	
	// Tentar inferir pelo nome do teste
	testName := strings.ToLower(testResult.Name)
	
	switch {
	case strings.Contains(testName, "consent"):
		return "missing_explicit_consent"
	case strings.Contains(testName, "autenticação") || strings.Contains(testName, "strong auth"):
		return "missing_strong_auth_check"
	case strings.Contains(testName, "pep"):
		return "missing_pep_check"
	case strings.Contains(testName, "retenção") || strings.Contains(testName, "retention"):
		return "missing_retention_limit_check"
	case strings.Contains(testName, "geográfic") || strings.Contains(testName, "internacional"):
		return "missing_geo_restriction"
	case strings.Contains(testName, "audit") || strings.Contains(testName, "trilha"):
		return "missing_audit_trail"
	case strings.Contains(testName, "aml") || strings.Contains(testName, "limit"):
		return "missing_aml_transaction_limit_check"
	case strings.Contains(testName, "dados pessoais") || strings.Contains(testName, "personal data"):
		return "missing_personal_data_authorization"
	case strings.Contains(testName, "multicaixa") || strings.Contains(testName, "payment"):
		return "missing_multicaixa_validation"
	case strings.Contains(testName, "idade") || strings.Contains(testName, "age"):
		return "missing_age_verification"
	}
	
	// Não foi possível determinar o tipo específico
	return ""
}