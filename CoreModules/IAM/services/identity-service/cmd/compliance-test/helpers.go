package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// executarTestesRegionais executa os testes de compliance para uma região específica
func executarTestesRegionais(logger *zap.Logger, config Config, region string) {
	// Caminho da matriz de conformidade regional
	matrixPath := filepath.Join(config.TestsDir, "regions", region, "compliance_matrix.json")
	
	// Verifica se a matriz existe
	if _, err := os.Stat(matrixPath); os.IsNotExist(err) {
		logger.Error("Matriz de conformidade não encontrada", 
			zap.String("region", region),
			zap.String("path", matrixPath))
		return
	}

	// Carrega a matriz de conformidade
	matrix, err := carregarMatrizConformidade(matrixPath)
	if err != nil {
		logger.Error("Erro ao carregar matriz de conformidade", 
			zap.String("region", region),
			zap.Error(err))
		return
	}

	logger.Info("Matriz de conformidade carregada com sucesso",
		zap.String("region", region),
		zap.String("regionName", matrix.RegionName),
		zap.Int("frameworks", len(matrix.Frameworks)),
		zap.Int("requirements", len(matrix.Requirements)))
	
	// Verifica se remediação está habilitada
	if config.Remediate {
		// Verifica se as regras de remediação para a região existem
		regionRulesPath := filepath.Join(config.RulesPath, strings.ToLower(region)+"_remediation_rules.json")
		if _, err := os.Stat(regionRulesPath); os.IsNotExist(err) {
			logger.Warn("Remediação habilitada, mas arquivo de regras não encontrado",
				zap.String("region", region),
				zap.String("expected_path", regionRulesPath))
			fmt.Printf("%s Remediação habilitada, mas arquivo de regras para %s não encontrado em: %s\n", 
				color.YellowString("⚠️"), 
				region,
				regionRulesPath)
		} else {
			logger.Info("Arquivo de regras de remediação encontrado",
				zap.String("region", region),
				zap.String("rules_path", regionRulesPath))
		}
	}

	// Prepara o sumário de testes
	summary := &TestSummary{
		Region:          region,
		RegionName:      matrix.RegionName,
		FrameworkScores: make(map[string]FrameworkScore),
		ExecutedAt:      time.Now(),
	}

	// Cria mapeamento de requisitos para frameworks
	reqToFramework := make(map[string]string)
	reqToCriticality := make(map[string]string)
	frameworkMap := make(map[string]Framework)
	
	for _, framework := range matrix.Frameworks {
		frameworkMap[framework.ID] = framework
	}
	
	for _, req := range matrix.Requirements {
		reqToFramework[req.ID] = req.FrameworkID
		reqToCriticality[req.ID] = req.Criticality
		
		// Inicializa contadores para frameworks
		if _, exists := summary.FrameworkScores[req.FrameworkID]; !exists {
			framework := frameworkMap[req.FrameworkID]
			summary.FrameworkScores[req.FrameworkID] = FrameworkScore{
				ID:   framework.ID,
				Name: framework.Name,
			}
		}
	}

	// Carrega todos os casos de teste para a região
	testCases, err := carregarCasosTeste(config.TestsDir, region, config.Tags, config.Frameworks)
	if err != nil {
		logger.Error("Erro ao carregar casos de teste", 
			zap.String("region", region),
			zap.Error(err))
		return
	}

	logger.Info("Casos de teste carregados",
		zap.String("region", region),
		zap.Int("testCount", len(testCases)))

	// Inicializa o contador para o sumário
	summary.TotalTests = len(testCases)
	
	// Rastreia o tempo de execução
	startTime := time.Now()
	
	// Executa os testes
	for _, testCase := range testCases {
		// Executa o caso de teste
		result, err := executarTeste(logger, config.OPAPath, testCase, reqToFramework, reqToCriticality)
		if err != nil {
			logger.Error("Erro ao executar teste",
				zap.String("testId", testCase.ID),
				zap.Error(err))
			continue
		}
		
		// Adiciona resultado ao sumário
		summary.TestResults = append(summary.TestResults, result)
		
		if result.Passed {
			summary.PassedTests++
			// Atualiza requisitos atendidos
			for _, reqID := range result.Requirements {
				if !contains(summary.RequirementsMet, reqID) {
					summary.RequirementsMet = append(summary.RequirementsMet, reqID)
				}
			}
		} else {
			summary.FailedTests++
			// Atualiza requisitos não atendidos
			for _, reqID := range result.Requirements {
				if !contains(summary.RequirementsFailed, reqID) {
					summary.RequirementsFailed = append(summary.RequirementsFailed, reqID)
				}
			}
		}
		
		// Atualiza pontuações por framework
		for _, frameworkID := range result.Frameworks {
			score := summary.FrameworkScores[frameworkID]
			score.TotalTests++
			
			if result.Passed {
				score.PassedTests++
			} else {
				score.FailedTests++
			}
			
			score.ComplianceScore = float64(score.PassedTests) / float64(score.TotalTests) * 100
			summary.FrameworkScores[frameworkID] = score
		}
	}
	
	// Calcula pontuação geral de conformidade
	if summary.TotalTests > 0 {
		summary.ComplianceScore = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}
	
	// Calcula duração total
	summary.Duration = time.Since(startTime).Milliseconds()
	
	// Gera relatórios
	gerarRelatorios(logger, config, summary)
	
	// Aplicar remediação se habilitada e se houver falhas nos testes
	if config.Remediate && summary.FailedTests > 0 {
		remediationConfig := RemediationConfig{
			Enabled:                 true,
			DryRun:                  config.DryRun,
			MaxSeverity:             config.MaxSeverity,
			MinSeverity:             config.MinSeverity,
			Frameworks:              config.Frameworks,
			RulesPath:               config.RulesPath,
			BackupDir:               config.BackupDir,
			IgnoreTypes:             config.IgnoreTypes,
			MaxRemediationsPerPolicy: config.MaxRemediationsPerPolicy,
			RequireApproval:         config.RequireApproval,
		}

		logger.Info("Iniciando processo de remediação automática",
			zap.String("region", region),
			zap.Bool("dry_run", config.DryRun),
			zap.Int("violations", summary.FailedTests),
			zap.String("max_severity", config.MaxSeverity))

		// Preparar a conversão do sumário para o formato esperado pelo remediador
		violationReport := ConversionTestSummaryToComplianceReport(summary)

		// Chamar o remediador
		remediationSummary, err := ApplyRemediations(context.Background(), violationReport, remediationConfig)
		if err != nil {
			logger.Error("Erro ao aplicar remediações", zap.Error(err))
		} else {
			// Adiciona informações de remediação ao relatório
			summary.RemediationApplied = true
			
			// Converte o sumário de remediação para o formato do relatório final
			remediationResult := &RemediationResult{
				Enabled:               remediationSummary.Enabled,
				Success:               remediationSummary.Successful,
				Timestamp:             remediationSummary.Timestamp,
				Message:               remediationSummary.Message,
				TotalViolations:       remediationSummary.TotalViolations,
				AttemptedRemediations: remediationSummary.AttemptedRemediations,
				SuccessfulRemediations: remediationSummary.SuccessfulRemediations,
				FailedRemediations:    remediationSummary.FailedRemediations,
				PolicyFilesModified:   remediationSummary.PolicyFilesModified,
				Errors:                remediationSummary.Errors,
				DryRun:               remediationSummary.DryRun,
			}
			
			// Atualiza o sumário com os resultados da remediação
			summary.RemediationResult = remediationResult
			
			// Exibir resumo da remediação
			logger.Info("Processo de remediação concluído",
				zap.Bool("success", remediationSummary.Successful),
				zap.Int("total_violations", remediationSummary.TotalViolations),
				zap.Int("attempted", remediationSummary.AttemptedRemediations),
				zap.Int("successful", remediationSummary.SuccessfulRemediations),
				zap.Int("failed", remediationSummary.FailedRemediations))
		}
	}
	
	// Imprime sumário no console se solicitado
	if config.ShowSummary {
		exibirSumarioConsole(summary)
	}

	// Gerar relatórios
	if config.Json {
		gerarRelatorioJSON(summary, config.OutputDir)
	}

	if config.HTML {
		gerarRelatorioHTML(summary, config.OutputDir)
	}

	// Registra estatísticas de teste
	logger.Info("Testes de compliance concluídos",
		zap.String("region", region),
		zap.Int("total", summary.TotalTests),
		zap.Int("passed", summary.PassedTests),
		zap.Int("failed", summary.FailedTests),
		zap.Float64("compliance_score", summary.ComplianceScore),
		zap.Duration("duration", time.Duration(summary.Duration)*time.Millisecond))
}

// carregarMatrizConformidade carrega a matriz de conformidade de um arquivo JSON
func carregarMatrizConformidade(filePath string) (*ComplianceMatrix, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo matriz: %w", err)
	}
	
	var matrix ComplianceMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		return nil, fmt.Errorf("erro ao decodificar matriz: %w", err)
	}
	
	return &matrix, nil
}

// carregarCasosTeste carrega os casos de teste para uma região específica
func carregarCasosTeste(baseDir, region string, tags, frameworks []string) ([]TestCase, error) {
	// Diretório de casos de teste para a região
	testCasesDir := filepath.Join(baseDir, "regions", region, "test_cases")
	
	// Lista todos os subdiretórios (categorias de teste)
	dirs, err := ioutil.ReadDir(testCasesDir)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias de teste: %w", err)
	}
	
	var testCases []TestCase
	
	// Para cada categoria, carrega os testes
	for _, dir := range dirs {
		if dir.IsDir() {
			categoryDir := filepath.Join(testCasesDir, dir.Name())
			
			// Lista arquivos JSON na categoria
			files, err := filepath.Glob(filepath.Join(categoryDir, "*.json"))
			if err != nil {
				return nil, fmt.Errorf("erro ao listar arquivos de teste: %w", err)
			}
			
			// Processa cada arquivo JSON como um caso de teste
			for _, file := range files {
				data, err := ioutil.ReadFile(file)
				if err != nil {
					return nil, fmt.Errorf("erro ao ler arquivo de teste %s: %w", file, err)
				}
				
				var testCase TestCase
				if err := json.Unmarshal(data, &testCase); err != nil {
					return nil, fmt.Errorf("erro ao decodificar caso de teste %s: %w", file, err)
				}
				
				// Filtra por tags se especificado
				if len(tags) > 0 && !anyTagMatch(testCase.Tags, tags) {
					continue
				}
				
				// Filtra por framework se especificado
				// Nota: isso requer um processamento adicional após mapear os requisitos
				// aos frameworks, que será feito em executarTestesRegionais
				
				testCases = append(testCases, testCase)
			}
		}
	}
	
	return testCases, nil
}

// anyTagMatch verifica se qualquer tag no slice tags corresponde a qualquer tag no slice filter
func anyTagMatch(tags, filter []string) bool {
	for _, tag := range tags {
		for _, f := range filter {
			if strings.EqualFold(tag, f) {
				return true
			}
		}
	}
	return false
}

// contains verifica se um slice contém um determinado item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}