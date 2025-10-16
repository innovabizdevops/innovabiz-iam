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
	"github.com/olekukonko/tablewriter"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// executarTeste executa um caso de teste específico contra a política OPA
func executarTeste(logger *zap.Logger, opaPath string, testCase TestCase, reqToFramework map[string]string, reqToCriticality map[string]string) (*TestResult, error) {
	// Prepara o resultado do teste
	result := &TestResult{
		TestCase:     testCase,
		ExecutedAt:   time.Now(),
		PolicyPath:   testCase.PolicyPath,
		Requirements: testCase.RequirementIDs,
		Tags:         testCase.Tags,
	}
	
	// Extrai a região de compliance do contexto
	if region, ok := testCase.Context["compliance_region"]; ok {
		result.ComplianceRegion = region
	}
	
	// Mapeia requisitos para frameworks
	for _, reqID := range testCase.RequirementIDs {
		if frameworkID, ok := reqToFramework[reqID]; ok {
			if !contains(result.Frameworks, frameworkID) {
				result.Frameworks = append(result.Frameworks, frameworkID)
			}
		}
		
		// Define criticality baseado no requisito de maior criticidade
		if criticality, ok := reqToCriticality[reqID]; ok {
			if result.Criticality == "" || criticality == "alta" || 
				(criticality == "média" && result.Criticality == "baixa") {
				result.Criticality = criticality
			}
		}
	}
	
	// Mede o tempo de execução
	startTime := time.Now()
	
	// Prepara a consulta Rego
	ctx := context.Background()
	
	// Forma o caminho da política completo
	fullPolicyPath := filepath.Join(opaPath, testCase.PolicyPath)
	
	// Executa a consulta usando o OPA
	r := rego.New(
		rego.Query("data."+strings.ReplaceAll(testCase.PolicyPath, "/", ".")),
		rego.Load([]string{fullPolicyPath}, nil),
	)
	
	// Prepara o input do teste
	inputBytes, err := json.Marshal(testCase.Input)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar input: %w", err)
	}
	
	var input interface{}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return nil, fmt.Errorf("erro ao preparar input: %w", err)
	}
	
	// Executa a consulta com o input
	rs, err := r.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("Erro ao avaliar política: %v", err)
		return result, nil
	}
	
	// Extrai o resultado
	var decision interface{}
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		decision = rs[0].Expressions[0].Value
	}
	
	// Registra o resultado e o tempo
	result.ActualDecision = decision
	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	
	// Compara com o resultado esperado
	expectedBytes, err := json.Marshal(testCase.ExpectedDecision)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar decisão esperada: %w", err)
	}
	
	actualBytes, err := json.Marshal(decision)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar decisão atual: %w", err)
	}
	
	// Compara resultados
	result.Passed = string(expectedBytes) == string(actualBytes)
	if !result.Passed {
		result.Message = fmt.Sprintf("Resultado esperado não corresponde ao resultado atual")
		
		// Extrai violações se existirem
		if actual, ok := decision.(map[string]interface{}); ok {
			if violations, ok := actual["violations"].([]interface{}); ok {
				for _, v := range violations {
					if violation, ok := v.(map[string]interface{}); ok {
						if code, ok := violation["code"].(string); ok {
							result.Violations = append(result.Violations, code)
						}
					}
				}
			}
		}
	}
	
	// Log do resultado do teste
	logLevel := zap.InfoLevel
	if !result.Passed {
		logLevel = zap.WarnLevel
	}
	
	logger.Log(logLevel, "Resultado do teste",
		zap.String("testId", testCase.ID),
		zap.String("name", testCase.Name),
		zap.Bool("passed", result.Passed),
		zap.String("criticality", result.Criticality),
		zap.Int64("executionTimeMs", result.ExecutionTimeMs))
	
	return result, nil
}

// gerarRelatorios gera os relatórios de conformidade em diferentes formatos
func gerarRelatorios(logger *zap.Logger, config Config, summary *TestSummary) {
	// Cria diretório de saída se não existir
	outputDir := filepath.Join(config.OutputDir, summary.Region)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("Erro ao criar diretório de saída",
			zap.String("dir", outputDir),
			zap.Error(err))
		return
	}
	
	// Gera relatório JSON
	if config.Json {
		jsonPath := filepath.Join(outputDir, fmt.Sprintf("compliance_report_%s_%s.json", 
			summary.Region, time.Now().Format("20060102_150405")))
		
		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			logger.Error("Erro ao gerar relatório JSON", zap.Error(err))
		} else {
			if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
				logger.Error("Erro ao salvar relatório JSON", zap.Error(err))
			} else {
				logger.Info("Relatório JSON gerado com sucesso", zap.String("path", jsonPath))
			}
		}
	}
	
	// Gera relatório HTML
	if config.HTML {
		htmlPath := filepath.Join(outputDir, fmt.Sprintf("compliance_report_%s_%s.html", 
			summary.Region, time.Now().Format("20060102_150405")))
		
		// Modelo HTML para relatório
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-PT">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Relatório de Conformidade - %s</title>
    <style>
        body {
            font-family: 'Segoe UI', Arial, sans-serif;
            margin: 20px;
            color: #333;
            line-height: 1.6;
        }
        header {
            background-color: #0056b3;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        h1 {
            margin: 0;
            font-weight: 600;
        }
        h2 {
            color: #0056b3;
            border-bottom: 1px solid #ddd;
            padding-bottom: 10px;
            margin-top: 30px;
        }
        .summary {
            background-color: #f9f9f9;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 30px;
            border-left: 5px solid #0056b3;
        }
        .summary-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            border-bottom: 1px dotted #ddd;
            padding-bottom: 5px;
        }
        .summary-label {
            font-weight: 600;
            flex: 1;
        }
        .summary-value {
            flex: 2;
            text-align: right;
        }
        .passed {
            color: #28a745;
            font-weight: bold;
        }
        .failed {
            color: #dc3545;
            font-weight: bold;
        }
        .warning {
            color: #ffc107;
            font-weight: bold;
        }
        table {
            width: 100%%;
            border-collapse: collapse;
            margin-bottom: 30px;
            box-shadow: 0 0 20px rgba(0,0,0,0.05);
        }
        th {
            background-color: #0056b3;
            color: white;
            padding: 12px;
            text-align: left;
        }
        td {
            padding: 12px;
            border-bottom: 1px solid #ddd;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        tr:hover {
            background-color: #f1f1f1;
        }
        .compliance-gauge {
            width: 100%%;
            background-color: #f0f0f0;
            border-radius: 5px;
            margin: 10px 0;
            height: 30px;
            position: relative;
        }
        .compliance-gauge-fill {
            height: 100%%;
            border-radius: 5px;
            background-color: #28a745;
            width: %f%%;
            position: absolute;
            left: 0;
            top: 0;
        }
        .compliance-gauge-text {
            position: absolute;
            width: 100%%;
            text-align: center;
            top: 50%%;
            transform: translateY(-50%%);
            font-weight: bold;
            color: white;
            text-shadow: 1px 1px 3px rgba(0,0,0,0.5);
            z-index: 1;
        }
        .framework {
            margin-bottom: 30px;
            padding: 15px;
            border-radius: 5px;
            background-color: #f9f9f9;
            border-left: 5px solid #0056b3;
        }
        .footer {
            text-align: center;
            margin-top: 50px;
            color: #888;
            font-size: 0.9em;
            border-top: 1px solid #ddd;
            padding-top: 20px;
        }
        .test-details {
            margin-top: 30px;
        }
        .test-item {
            border: 1px solid #ddd;
            margin-bottom: 15px;
            border-radius: 5px;
            overflow: hidden;
        }
        .test-header {
            padding: 10px;
            font-weight: bold;
            display: flex;
            justify-content: space-between;
            background-color: #f0f0f0;
        }
        .test-content {
            padding: 15px;
        }
        .test-passed .test-header {
            background-color: #d4edda;
            color: #155724;
        }
        .test-failed .test-header {
            background-color: #f8d7da;
            color: #721c24;
        }
        .criticality-alta {
            color: #dc3545;
        }
        .criticality-media {
            color: #fd7e14;
        }
        .criticality-baixa {
            color: #6c757d;
        }
    </style>
</head>
<body>
    <header>
        <h1>Relatório de Conformidade Regulatória</h1>
        <p>%s (%s)</p>
    </header>

    <div class="summary">
        <h2>Resumo de Conformidade</h2>
        <div class="summary-row">
            <span class="summary-label">Data de Execução:</span>
            <span class="summary-value">%s</span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Pontuação de Conformidade:</span>
            <span class="summary-value">%.2f%%</span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Total de Testes:</span>
            <span class="summary-value">%d</span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Testes com Sucesso:</span>
            <span class="summary-value"><span class="passed">%d</span></span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Testes Falhados:</span>
            <span class="summary-value"><span class="failed">%d</span></span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Tempo de Execução:</span>
            <span class="summary-value">%d ms</span>
        </div>
        <div class="compliance-gauge">
            <div class="compliance-gauge-fill" style="width: %.2f%%;"></div>
            <div class="compliance-gauge-text">%.2f%%</div>
        </div>
    </div>
`, 
    summary.RegionName, 
    summary.ComplianceScore,
    summary.RegionName, summary.Region, 
    summary.ExecutedAt.Format("02/01/2006 15:04:05"), 
    summary.ComplianceScore, 
    summary.TotalTests, 
    summary.PassedTests, 
    summary.FailedTests, 
    summary.Duration,
    summary.ComplianceScore,
    summary.ComplianceScore)
		
		// Adiciona seção de frameworks
		html += `
    <h2>Conformidade por Framework Regulatório</h2>
`

		for _, fs := range summary.FrameworkScores {
			scoreClass := "passed"
			if fs.ComplianceScore < 70 {
				scoreClass = "failed"
			} else if fs.ComplianceScore < 90 {
				scoreClass = "warning"
			}
			
			html += fmt.Sprintf(`
    <div class="framework">
        <h3>%s</h3>
        <div class="summary-row">
            <span class="summary-label">Pontuação de Conformidade:</span>
            <span class="summary-value"><span class="%s">%.2f%%</span></span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Testes Passados:</span>
            <span class="summary-value">%d de %d</span>
        </div>
        <div class="compliance-gauge">
            <div class="compliance-gauge-fill" style="width: %.2f%%;"></div>
            <div class="compliance-gauge-text">%.2f%%</div>
        </div>
    </div>
`, 
            fs.Name, 
            scoreClass,
            fs.ComplianceScore, 
            fs.PassedTests, fs.TotalTests,
            fs.ComplianceScore,
            fs.ComplianceScore)
		}
		
		// Adiciona detalhes dos testes
		html += `
    <h2>Detalhes dos Testes</h2>
    <div class="test-details">
`
		
		// Adiciona cada resultado de teste
		for _, tr := range summary.TestResults {
			testClass := "test-passed"
			testStatus := "Passou"
			if !tr.Passed {
				testClass = "test-failed"
				testStatus = "Falhou"
			}
			
			criticality := "criticality-baixa"
			if tr.Criticality == "alta" {
				criticality = "criticality-alta"
			} else if tr.Criticality == "média" || tr.Criticality == "media" {
				criticality = "criticality-media"
			}
			
			html += fmt.Sprintf(`
        <div class="test-item %s">
            <div class="test-header">
                <span>%s - %s</span>
                <span>%s</span>
            </div>
            <div class="test-content">
                <p>%s</p>
                <p><strong>ID:</strong> %s</p>
                <p><strong>Criticidade:</strong> <span class="%s">%s</span></p>
                <p><strong>Tempo de Execução:</strong> %d ms</p>
`,
                testClass,
                tr.TestCase.ID,
                tr.TestCase.Name,
                testStatus,
                tr.TestCase.Description,
                tr.TestCase.ID,
                criticality,
                tr.Criticality,
                tr.ExecutionTimeMs)
                
            if !tr.Passed && tr.Message != "" {
                html += fmt.Sprintf(`
                <p><strong>Mensagem:</strong> %s</p>
`, tr.Message)
            }
            
            if len(tr.Violations) > 0 {
                html += `
                <p><strong>Violações:</strong></p>
                <ul>
`
                for _, v := range tr.Violations {
                    html += fmt.Sprintf(`
                    <li>%s</li>
`, v)
                }
                html += `
                </ul>
`
            }
            
            html += `
            </div>
        </div>
`
		}
		
		html += `
    </div>

    <div class="footer">
        <p>Gerado por INNOVABIZ IAM Compliance Testing Framework</p>
        <p>© 2025 INNOVABIZ - Todos os direitos reservados</p>
    </div>
</body>
</html>
`
		
		// Salva o arquivo HTML
		if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
			logger.Error("Erro ao salvar relatório HTML", zap.Error(err))
		} else {
			logger.Info("Relatório HTML gerado com sucesso", zap.String("path", htmlPath))
		}
	}
}

// exibirSumarioConsole exibe um sumário dos testes no console
func exibirSumarioConsole(summary *TestSummary) {
	fmt.Printf("\n%s Relatório de Compliance: %s (%s)\n",
		color.CyanString("📊"),
		color.CyanString(summary.RegionName),
		summary.Region)

	fmt.Printf("   %s Testes executados: %d\n", color.WhiteString("•"), summary.TotalTests)
	fmt.Printf("   %s Testes aprovados: %s\n", 
		color.WhiteString("•"), 
		color.GreenString("%d", summary.PassedTests))
	fmt.Printf("   %s Testes reprovados: %s\n", 
		color.WhiteString("•"), 
		summary.FailedTests > 0 ? color.RedString("%d", summary.FailedTests) : color.GreenString("%d", summary.FailedTests))
	fmt.Printf("   %s Pontuação de compliance: %s\n", 
		color.WhiteString("•"), 
		formatComplianceScore(summary.ComplianceScore))

	fmt.Println("\nResultados por Framework:")
	for _, frameworkScore := range summary.FrameworkScores {
		fmt.Printf("   %s %s: %s (%d/%d testes aprovados)\n",
			color.WhiteString("•"),
			frameworkScore.Name,
			formatComplianceScore(frameworkScore.ComplianceScore),
			frameworkScore.PassedTests,
			frameworkScore.TotalTests)
	}

	// Exibir requisitos não atendidos se houver
	if len(summary.RequirementsFailed) > 0 {
		fmt.Printf("\n%s %s\n", 
			color.RedString("⚠️"),
			color.RedString("Requisitos não atendidos:"))
			
		for _, req := range summary.RequirementsFailed {
			fmt.Printf("   %s %s\n", color.WhiteString("•"), req)
		}
	}
	
	// Exibir sumário da remediação se aplicada
	if summary.RemediationApplied && summary.RemediationResult != nil {
		exibirSumarioRemediacao(summary.RemediationResult)
	}

	// Exibir tempo de execução
	fmt.Printf("\n%s Tempo de execução: %s\n", 
		color.WhiteString("⏰"),
		formatDuration(time.Duration(summary.Duration)*time.Millisecond))
}

// exibirSumarioRemediacao exibe o sumário dos resultados da remediação
func exibirSumarioRemediacao(result *RemediationResult) {
	// Cabeçalho da seção
	fmt.Printf("\n%s %s\n", 
		color.YellowString("🔧"),
		color.YellowString("Remediação Automática"))
	
	// Status da remediação
	if result.DryRun {
		fmt.Printf("   %s Modo: %s\n", 
			color.WhiteString("•"),
			color.BlueString("Simulação (dry-run)"))
	} else {
		fmt.Printf("   %s Modo: %s\n", 
			color.WhiteString("•"),
			color.MagentaString("Aplicação real"))
	}
	
	// Estatísticas
	fmt.Printf("   %s Total de violações detectadas: %d\n", 
		color.WhiteString("•"),
		result.TotalViolations)
		
	fmt.Printf("   %s Remediações tentadas: %d\n", 
		color.WhiteString("•"),
		result.AttemptedRemediations)
		
	fmt.Printf("   %s Remediações bem-sucedidas: %s\n", 
		color.WhiteString("•"),
		color.GreenString("%d", result.SuccessfulRemediations))
	
	// Mostrar falhas se houver
	if result.FailedRemediations > 0 {
		fmt.Printf("   %s Remediações falhas: %s\n", 
			color.WhiteString("•"),
			color.RedString("%d", result.FailedRemediations))
	} else {
		fmt.Printf("   %s Remediações falhas: %s\n", 
			color.WhiteString("•"),
			color.GreenString("%d", result.FailedRemediations))
	}
	
	// Mostrar arquivos modificados se estiver em modo real e houver modificações
	if !result.DryRun && len(result.PolicyFilesModified) > 0 {
		fmt.Printf("\n%s %s\n", 
			color.WhiteString("📄"),
			color.WhiteString("Arquivos de política modificados:"))
		
		for file, rules := range result.PolicyFilesModified {
			fmt.Printf("   %s %s (%d regras aplicadas)\n", 
				color.WhiteString("•"),
				file, len(rules))
		}
	}
	
	// Mostrar erros se houver
	if len(result.Errors) > 0 && result.FailedRemediations > 0 {
		fmt.Printf("\n%s %s\n", 
			color.RedString("❌"),
			color.RedString("Erros durante remediação:"))
		
		// Mostrar os 3 primeiros erros para não sobrecarregar o console
		showErrors := result.Errors
		if len(showErrors) > 3 {
			showErrors = showErrors[:3]
		}
		
		for _, err := range showErrors {
			fmt.Printf("   %s Arquivo: %s, Regra: %s\n     %s\n", 
				color.WhiteString("•"),
				err.PolicyFile,
				err.RuleID,
				color.RedString(err.ErrorMessage))
		}
		
		if len(result.Errors) > 3 {
			fmt.Printf("   %s %s\n", 
				color.WhiteString("•"),
				color.YellowString("... e mais %d erros (veja o relatório completo para detalhes)", len(result.Errors) - 3))
		}
	}
}

// exibirSumario exibe um resumo dos resultados no terminal
func exibirSumario(summary *TestSummary) {
	// Exibe cabeçalho
	fmt.Println()
	color.New(color.FgHiWhite, color.Bold).Printf("=== RELATÓRIO DE CONFORMIDADE - %s (%s) ===\n\n", 
		summary.RegionName, summary.Region)
	
	// Exibe pontuação geral
	fmt.Print("Pontuação de Conformidade: ")
	if summary.ComplianceScore >= 90 {
		color.New(color.FgHiGreen, color.Bold).Printf("%.2f%%\n", summary.ComplianceScore)
	} else if summary.ComplianceScore >= 70 {
		color.New(color.FgHiYellow, color.Bold).Printf("%.2f%%\n", summary.ComplianceScore)
	} else {
		color.New(color.FgHiRed, color.Bold).Printf("%.2f%%\n", summary.ComplianceScore)
	}
	
	// Resumo dos testes
	fmt.Printf("Total de Testes: %d (", summary.TotalTests)
	color.New(color.FgGreen).Printf("Passou: %d", summary.PassedTests)
	fmt.Print(", ")
	color.New(color.FgRed).Printf("Falhou: %d", summary.FailedTests)
	fmt.Println(")")
	
	fmt.Printf("Tempo de Execução: %d ms\n\n", summary.Duration)
	
	// Tabela de frameworks
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Framework", "Pontuação", "Passou", "Falhou", "Total"})
	table.SetBorder(false)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
	)
	
	for _, fs := range summary.FrameworkScores {
		row := []string{
			fs.Name,
			fmt.Sprintf("%.2f%%", fs.ComplianceScore),
			fmt.Sprintf("%d", fs.PassedTests),
			fmt.Sprintf("%d", fs.FailedTests),
			fmt.Sprintf("%d", fs.TotalTests),
		}
		table.Append(row)
	}
	
	fmt.Println("Pontuação por Framework Regulatório:")
	table.Render()
	
	// Requisitos não atendidos
	if len(summary.RequirementsFailed) > 0 {
		color.New(color.FgHiRed, color.Bold).Println("\nRequisitos Regulatórios Não Atendidos:")
		for _, req := range summary.RequirementsFailed {
			fmt.Printf("  • %s\n", req)
		}
	}
}

// setupLogger inicializa o logger para a CLI
func setupLogger(verbose bool) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	if !verbose {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return config.Build()
}

// parseFlags processa os argumentos da linha de comando
func parseFlags() Config {
	// Configuração básica
	opaPath := flag.String("opa", "./policies", "Caminho raiz das políticas OPA")
	testsDir := flag.String("tests", "./tests/opa-compliance", "Diretório dos testes de compliance")
	outputDir := flag.String("output", "./reports", "Diretório para salvar relatórios")
	regionStr := flag.String("regions", "AO", "Regiões a testar (separadas por vírgula)")
	frameworkStr := flag.String("frameworks", "", "Frameworks a testar (separados por vírgula)")
	tagStr := flag.String("tags", "", "Tags a filtrar (separadas por vírgula)")
	format := flag.String("format", "table", "Formato do relatório (table, json, html)")
	verbose := flag.Bool("verbose", false, "Modo verboso")
	summary := flag.Bool("summary", true, "Exibir sumário no console")
	json := flag.Bool("json", true, "Gerar relatório JSON")
	html := flag.Bool("html", true, "Gerar relatório HTML")
	
	// Configuração de remediação
	remediate := flag.Bool("remediate", false, "Ativar remediação automática para falhas de compliance")
	dryRun := flag.Bool("dry-run", true, "Executar remediação em modo simulação (sem aplicar mudanças)")
	maxSeverity := flag.String("max-severity", "alta", "Severidade máxima para remediação (baixa, media, alta)")
	minSeverity := flag.String("min-severity", "baixa", "Severidade mínima para remediação (baixa, media, alta)")
	ignoreTypesStr := flag.String("ignore-types", "", "Tipos de violação a ignorar (separados por vírgula)")
	rulesPath := flag.String("rules-path", filepath.Join("remediator", "rules"), "Caminho para regras de remediação")
	backupDir := flag.String("backup-dir", filepath.Join("remediator", "backups"), "Diretório para backups de arquivos")
	noApproval := flag.Bool("no-approval", false, "Não solicitar aprovação antes de aplicar remediações")
	maxRemedPerPolicy := flag.Int("max-remed-per-policy", 5, "Número máximo de remediações por arquivo de política")
	
	flag.Parse()
	
	// Configuração base
	config := Config{
		OPAPath:      *opaPath,
		TestsDir:     *testsDir,
		OutputDir:    *outputDir,
		Regions:      strings.Split(*regionStr, ","),
		ReportFormat: *format,
		Verbose:      *verbose,
		ShowSummary:  *summary,
		Json:         *json,
		HTML:         *html,
		
		// Configuração de remediação
		Remediate:                *remediate,
		DryRun:                   *dryRun,
		MaxSeverity:              *maxSeverity,
		MinSeverity:              *minSeverity,
		RulesPath:                *rulesPath,
		BackupDir:                *backupDir,
		RequireApproval:          !*noApproval,
		MaxRemediationsPerPolicy: *maxRemedPerPolicy,
	}
	
	// Processar strings separadas por vírgulas
	if *frameworkStr != "" {
		config.Frameworks = strings.Split(*frameworkStr, ",")
	}
	
	if *tagStr != "" {
		config.Tags = strings.Split(*tagStr, ",")
	}
	
	if *ignoreTypesStr != "" {
		config.IgnoreTypes = strings.Split(*ignoreTypesStr, ",")
	}
	
	return config
}