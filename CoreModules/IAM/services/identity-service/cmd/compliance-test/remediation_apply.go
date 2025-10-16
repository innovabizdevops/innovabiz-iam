package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/innovabizdevops/innovabiz-iam/remediator"
	"go.uber.org/zap"
)

// RemediationConfig contém as configurações para o processo de remediação
type RemediationConfig struct {
	Enabled                 bool     // Se a remediação está habilitada
	DryRun                  bool     // Se é apenas simulação (sem aplicar mudanças)
	MaxSeverity             string   // Severidade máxima a remediar (baixa, media, alta)
	MinSeverity             string   // Severidade mínima a remediar (baixa, media, alta)
	Frameworks              []string // Frameworks específicos para remediar (vazio = todos)
	RulesPath               string   // Caminho base para regras de remediação
	BackupDir               string   // Diretório para backup dos arquivos
	IgnoreTypes             []string // Tipos de violação a ignorar
	RequireApproval         bool     // Se requer aprovação do usuário antes de aplicar
	MaxRemediationsPerPolicy int     // Número máximo de remediações por arquivo de política
}

// RemediationSummary contém o resumo das remediações aplicadas
type RemediationSummary struct {
	Enabled               bool                `json:"enabled"`
	Successful            bool                `json:"successful"`
	Timestamp             string              `json:"timestamp"`
	Message               string              `json:"message"`
	TotalViolations       int                 `json:"total_violations"`
	AttemptedRemediations int                 `json:"attempted_remediations"`
	SuccessfulRemediations int                `json:"successful_remediations"`
	FailedRemediations    int                 `json:"failed_remediations"`
	PolicyFilesModified   map[string][]string `json:"policy_files_modified"`
	Errors                []RemediationError  `json:"errors"`
	DryRun               bool                `json:"dry_run"`
}

// RuleFile representa o formato do arquivo de regras de remediação
type RuleFile struct {
	Rules []remediator.RemediationRule `json:"rules"`
}

// ApplyRemediations aplica regras de remediação às violações de compliance
func ApplyRemediations(ctx context.Context, report *ComplianceReport, config RemediationConfig) (*RemediationSummary, error) {
	// Inicializa o logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Se não estiver habilitada, retorna imediatamente
	if !config.Enabled {
		return &RemediationSummary{
			Enabled:   false,
			Successful: true,
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Remediação não habilitada",
			DryRun:    config.DryRun,
		}, nil
	}

	// Inicializa o resumo da remediação
	summary := &RemediationSummary{
		Enabled:             true,
		Timestamp:          time.Now().Format(time.RFC3339),
		DryRun:             config.DryRun,
		PolicyFilesModified: make(map[string][]string),
	}

	// Carrega regras de remediação para a região
	rules, err := loadRemediationRules(ctx, report.Region, config)
	if err != nil {
		summary.Message = fmt.Sprintf("Erro ao carregar regras de remediação: %v", err)
		summary.Successful = false
		return summary, err
	}

	// Se não houver regras, retorna com aviso
	if len(rules) == 0 {
		summary.Message = "Nenhuma regra de remediação encontrada para a região"
		summary.Successful = true
		return summary, nil
	}

	// Prepara as violações para remediação
	violations, policyViolationsMap := prepareViolations(report, config)
	summary.TotalViolations = len(violations)

	// Se não houver violações aplicáveis, retorna
	if len(violations) == 0 {
		summary.Message = "Nenhuma violação encontrada para remediar"
		summary.Successful = true
		return summary, nil
	}

	// Log de violações encontradas
	logger.Info("Violações encontradas para remediação",
		zap.Int("total", len(violations)),
		zap.Int("policies_affected", len(policyViolationsMap)))

	// Se requer aprovação do usuário, solicitá-la
	if config.RequireApproval && !config.DryRun {
		approved := promptUserApproval(len(violations), len(policyViolationsMap))
		if !approved {
			summary.Message = "Remediação cancelada pelo usuário"
			summary.Successful = false
			return summary, fmt.Errorf("remediação cancelada pelo usuário")
		}
	}

	// Cria o diretório de backup se necessário e se não for dry-run
	if !config.DryRun {
		if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
			logger.Error("Não foi possível criar diretório de backup",
				zap.String("path", config.BackupDir),
				zap.Error(err))
			summary.Message = fmt.Sprintf("Erro ao criar diretório de backup: %v", err)
			summary.Successful = false
			return summary, err
		}
	}

	// Cria o remediador
	remedier := &remediator.Remediator{
		DryRun:   config.DryRun,
		BackupDir: config.BackupDir,
		Logger:    logger,
	}

	// Aplica remediações por arquivo de política
	for policyFile, fileViolations := range policyViolationsMap {
		// Limita o número de remediações por arquivo
		applicableViolations := fileViolations
		if len(fileViolations) > config.MaxRemediationsPerPolicy {
			applicableViolations = fileViolations[:config.MaxRemediationsPerPolicy]
			logger.Warn("Número de violações excede o limite por arquivo",
				zap.String("policy", policyFile),
				zap.Int("total", len(fileViolations)),
				zap.Int("applied", config.MaxRemediationsPerPolicy))
		}

		// Filtra regras aplicáveis a esta política
		applicableRules := []remediator.RemediationRule{}
		for _, rule := range rules {
			for _, violation := range applicableViolations {
				if isRuleApplicable(rule, violation) {
					applicableRules = append(applicableRules, rule)
					break
				}
			}
		}

		// Se não houver regras aplicáveis, continua para o próximo arquivo
		if len(applicableRules) == 0 {
			logger.Info("Nenhuma regra aplicável para remediação",
				zap.String("policy", policyFile))
			continue
		}

		// Aplica remediações para este arquivo
		results, err := remedier.RemediatePolicy(ctx, policyFile, applicableRules, fileViolations)
		if err != nil {
			logger.Error("Erro ao remediar política",
				zap.String("policy", policyFile),
				zap.Error(err))
			summary.Errors = append(summary.Errors, RemediationError{
				PolicyFile:   policyFile,
				RuleID:       "",
				ErrorMessage: err.Error(),
			})
			continue
		}

		// Atualiza estatísticas
		summary.AttemptedRemediations += len(results)
		for _, result := range results {
			if result.Success {
				summary.SuccessfulRemediations++
				summary.PolicyFilesModified[policyFile] = append(
					summary.PolicyFilesModified[policyFile],
					result.RuleID)
			} else {
				summary.FailedRemediations++
				summary.Errors = append(summary.Errors, RemediationError{
					PolicyFile:   policyFile,
					RuleID:       result.RuleID,
					ErrorMessage: result.Error,
				})
			}
		}
	}

	// Define status de sucesso e mensagem final
	if summary.FailedRemediations == 0 && summary.SuccessfulRemediations > 0 {
		summary.Successful = true
		summary.Message = "Remediação aplicada com sucesso"
	} else if summary.SuccessfulRemediations > 0 {
		summary.Successful = true
		summary.Message = "Remediação parcialmente bem-sucedida"
	} else {
		summary.Successful = false
		summary.Message = "Falha ao aplicar remediações"
	}

	return summary, nil
}

// loadRemediationRules carrega regras de remediação para uma região específica
func loadRemediationRules(ctx context.Context, region string, config RemediationConfig) ([]remediator.RemediationRule, error) {
	// Caminho para o arquivo de regras
	rulesPath := filepath.Join(config.RulesPath, strings.ToLower(region)+"_remediation_rules.json")

	// Verifica se arquivo existe
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo de regras não encontrado: %s", rulesPath)
	}

	// Lê o arquivo
	data, err := ioutil.ReadFile(rulesPath)
	if err != nil {
		return nil, err
	}

	// Decodifica as regras
	var ruleFile RuleFile
	if err := json.Unmarshal(data, &ruleFile); err != nil {
		return nil, err
	}

	// Filtra regras por framework se especificado
	if len(config.Frameworks) > 0 {
		filteredRules := []remediator.RemediationRule{}
		for _, rule := range ruleFile.Rules {
			// Se a regra não especificar frameworks, é aplicável a todos
			if len(rule.Frameworks) == 0 {
				filteredRules = append(filteredRules, rule)
				continue
			}

			// Verifica interseção entre frameworks da regra e os solicitados
			for _, framework := range config.Frameworks {
				for _, ruleFramework := range rule.Frameworks {
					if strings.EqualFold(framework, ruleFramework) {
						filteredRules = append(filteredRules, rule)
						break
					}
				}
			}
		}
		return filteredRules, nil
	}

	return ruleFile.Rules, nil
}

// prepareViolations prepara as violações para remediação, filtrando por severidade e tipo
func prepareViolations(report *ComplianceReport, config RemediationConfig) ([]remediator.ComplianceViolation, map[string][]remediator.ComplianceViolation) {
	violations := []remediator.ComplianceViolation{}
	policyViolationsMap := make(map[string][]remediator.ComplianceViolation)

	// Mapeia severidades para valores numéricos para comparação
	severityValues := map[string]int{
		"baixa": 1,
		"media": 2,
		"alta":  3,
		"low":   1,
		"medium": 2,
		"high":  3,
	}

	minSeverityValue := severityValues[strings.ToLower(config.MinSeverity)]
	maxSeverityValue := severityValues[strings.ToLower(config.MaxSeverity)]

	// Prepara mapa de tipos de violação a ignorar para busca rápida
	ignoreTypes := make(map[string]bool)
	for _, t := range config.IgnoreTypes {
		ignoreTypes[strings.ToLower(t)] = true
	}

	// Processa resultados de cada framework
	for _, frameworkResult := range report.FrameworkResults {
		// Pula frameworks não selecionados se houver filtro
		if len(config.Frameworks) > 0 {
			frameworkSelected := false
			for _, selectedFramework := range config.Frameworks {
				if strings.EqualFold(selectedFramework, frameworkResult.FrameworkID) {
					frameworkSelected = true
					break
				}
			}
			if !frameworkSelected {
				continue
			}
		}

		// Processa cada teste falho
		for _, testResult := range frameworkResult.TestResults {
			if testResult.Passed {
				continue // Pula testes que passaram
			}

			// Verifica se o tipo de violação deve ser ignorado
			if _, shouldIgnore := ignoreTypes[strings.ToLower(testResult.ViolationType)]; shouldIgnore {
				continue
			}

			// Verifica se a severidade está dentro dos limites
			severityValue := severityValues[strings.ToLower(testResult.Criticality)]
			if severityValue < minSeverityValue || severityValue > maxSeverityValue {
				continue
			}

			// Cria a violação
			violation := remediator.ComplianceViolation{
				TestID:          testResult.TestID,
				PolicyFile:      testResult.PolicyFile,
				ViolationType:   testResult.ViolationType,
				Severity:        testResult.Criticality,
				ActualValue:     testResult.ActualDecision,
				ExpectedValue:   testResult.ExpectedDecision,
				Framework:       frameworkResult.FrameworkID,
				Description:     testResult.Description,
				RequirementIDs:  testResult.RequirementIDs,
			}

			// Adiciona à lista de violações
			violations = append(violations, violation)

			// Agrupa por arquivo de política para processamento eficiente
			policyViolationsMap[testResult.PolicyFile] = append(
				policyViolationsMap[testResult.PolicyFile],
				violation)
		}
	}

	return violations, policyViolationsMap
}

// isRuleApplicable verifica se uma regra é aplicável a uma violação específica
func isRuleApplicable(rule remediator.RemediationRule, violation remediator.ComplianceViolation) bool {
	// Verifica o tipo de violação
	if rule.ViolationType != "" && !strings.EqualFold(rule.ViolationType, violation.ViolationType) {
		return false
	}

	// Verifica se o ID do teste corresponde se especificado na regra
	if rule.TestID != "" && !strings.EqualFold(rule.TestID, violation.TestID) {
		return false
	}

	// Verifica se algum padrão de expressão regular corresponde ao arquivo
	if len(rule.PolicyPatterns) > 0 {
		matched := false
		for _, pattern := range rule.PolicyPatterns {
			if strings.Contains(violation.PolicyFile, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// promptUserApproval solicita aprovação do usuário antes de aplicar remediações
func promptUserApproval(violationCount, policyCount int) bool {
	fmt.Printf("\n%s %s\n",
		color.YellowString("⚠️"),
		color.YellowString("Solicitação de aprovação para remediação automática"))

	fmt.Printf("Serão remediadas %s violações em %s arquivos de política.\n",
		color.CyanString("%d", violationCount),
		color.CyanString("%d", policyCount))

	fmt.Printf("Os arquivos serão modificados. Backups serão criados automaticamente.\n\n")

	fmt.Printf("Deseja prosseguir com a remediação? [s/N]: ")
	
	var response string
	fmt.Scanln(&response)
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "s" || response == "sim" || response == "y" || response == "yes"
}