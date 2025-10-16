package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/innovabiz/iam/common/logging"
	"github.com/innovabiz/iam/common/telemetry"
	"github.com/innovabiz/iam/models"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
)

// RegionComplianceFramework define a estrutura principal para testes de conformidade regional
// das políticas OPA, suportando múltiplos contextos de conformidade e regulações
type RegionComplianceFramework struct {
	Logger            logging.Logger
	Telemetry         telemetry.Provider
	PolicyDir         string
	TestCasesDir      string
	ComplianceMatrices map[string]*ComplianceMatrix
	Regions           []string
	ConfigProvider    ConfigProvider
}

// ComplianceMatrix representa a matriz de conformidade para uma região específica
type ComplianceMatrix struct {
	RegionCode     string                 `json:"regionCode"`     // Código da região (ex: "AO", "EU", "BR")
	RegionName     string                 `json:"regionName"`     // Nome da região (ex: "Angola", "Europa", "Brasil")
	Frameworks     []ComplianceFramework  `json:"frameworks"`     // Frameworks de conformidade aplicáveis
	Requirements   []ComplianceRequirement `json:"requirements"`   // Requisitos específicos
	TestCases      map[string][]TestCase   `json:"testCases"`      // Casos de teste por categoria
	CrossRegional  bool                   `json:"crossRegional"`  // Indica se é uma matriz cross-regional
}

// ComplianceFramework representa um framework de conformidade (ex: GDPR, LGPD, Lei n.º 22/11)
type ComplianceFramework struct {
	ID          string   `json:"id"`          // Identificador único do framework
	Name        string   `json:"name"`        // Nome do framework
	Version     string   `json:"version"`     // Versão do framework
	Description string   `json:"description"` // Descrição do framework
	References  []string `json:"references"`  // Referências para documentação
}

// ComplianceRequirement representa um requisito específico de conformidade
type ComplianceRequirement struct {
	ID          string   `json:"id"`          // Identificador único do requisito
	Name        string   `json:"name"`        // Nome do requisito
	Description string   `json:"description"` // Descrição do requisito
	FrameworkID string   `json:"frameworkId"` // Framework ao qual pertence
	ArticleRefs []string `json:"articleRefs"` // Referências para artigos específicos
	Criticality string   `json:"criticality"` // Criticidade (alta, média, baixa)
	TestCaseIDs []string `json:"testCaseIds"` // IDs dos casos de teste associados
}

// TestCase representa um caso de teste específico para validação de conformidade
type TestCase struct {
	ID               string                 `json:"id"`               // Identificador único do caso de teste
	Name             string                 `json:"name"`             // Nome do caso de teste
	Description      string                 `json:"description"`      // Descrição do caso de teste
	RequirementIDs   []string               `json:"requirementIds"`   // Requisitos associados
	PolicyPath       string                 `json:"policyPath"`       // Caminho da política a ser testada
	Input            map[string]interface{} `json:"input"`            // Input para o teste
	ExpectedDecision interface{}            `json:"expectedDecision"` // Decisão esperada
	ExpectedErrors   []string               `json:"expectedErrors"`   // Erros esperados
	Tags             []string               `json:"tags"`             // Tags para categorização
	Context          map[string]interface{} `json:"context"`          // Contexto adicional para o teste
}

// ConfigProvider é a interface para obtenção de configurações específicas por região
type ConfigProvider interface {
	GetRegionConfig(region string) (map[string]interface{}, error)
	GetTenantConfig(region string, tenantID string) (map[string]interface{}, error)
	GetComplianceConfig(region string, framework string) (map[string]interface{}, error)
}

// TestResult representa o resultado de um teste de conformidade
type TestResult struct {
	TestCaseID       string                 `json:"testCaseId"`       // ID do caso de teste
	RegionCode       string                 `json:"regionCode"`       // Código da região
	FrameworkIDs     []string               `json:"frameworkIds"`     // Frameworks testados
	RequirementIDs   []string               `json:"requirementIds"`   // Requisitos testados
	Input            map[string]interface{} `json:"input"`            // Input utilizado
	ActualDecision   interface{}            `json:"actualDecision"`   // Decisão real obtida
	ExpectedDecision interface{}            `json:"expectedDecision"` // Decisão esperada
	Success          bool                   `json:"success"`          // Sucesso do teste
	ErrorMsg         string                 `json:"errorMsg"`         // Mensagem de erro se houver
	Duration         time.Duration          `json:"duration"`         // Duração do teste
	Trace            []string               `json:"trace,omitempty"`  // Trace de execução quando aplicável
	PolicyPath       string                 `json:"policyPath"`       // Caminho da política testada
}

// ComplianceSummary representa um resumo dos resultados de conformidade
type ComplianceSummary struct {
	RegionCode      string                          `json:"regionCode"`      // Código da região
	TotalTests      int                             `json:"totalTests"`      // Total de testes executados
	PassedTests     int                             `json:"passedTests"`     // Testes aprovados
	FailedTests     int                             `json:"failedTests"`     // Testes falhos
	Coverage        float64                         `json:"coverage"`        // Cobertura geral de conformidade
	FrameworkScores map[string]FrameworkScore       `json:"frameworkScores"` // Pontuação por framework
	Requirements    map[string]RequirementCompliance `json:"requirements"`    // Conformidade por requisito
	Timestamp       time.Time                       `json:"timestamp"`       // Timestamp da execução
	Duration        time.Duration                   `json:"duration"`        // Duração total dos testes
}

// FrameworkScore representa a pontuação de conformidade para um framework
type FrameworkScore struct {
	FrameworkID       string  `json:"frameworkId"`       // ID do framework
	Name              string  `json:"name"`              // Nome do framework
	TotalRequirements int     `json:"totalRequirements"` // Total de requisitos
	CoveredRequirements int   `json:"coveredRequirements"` // Requisitos cobertos
	ComplianceScore   float64 `json:"complianceScore"`   // Pontuação de conformidade
	CriticalIssues    int     `json:"criticalIssues"`    // Problemas críticos encontrados
}

// RequirementCompliance representa o estado de conformidade para um requisito específico
type RequirementCompliance struct {
	RequirementID string `json:"requirementId"` // ID do requisito
	Name          string `json:"name"`          // Nome do requisito
	FrameworkID   string `json:"frameworkId"`   // ID do framework associado
	Compliant     bool   `json:"compliant"`     // Indica se está conforme
	TestsPassed   int    `json:"testsPassed"`   // Quantidade de testes aprovados
	TestsFailed   int    `json:"testsFailed"`   // Quantidade de testes falhos
	Criticality   string `json:"criticality"`   // Criticidade do requisito
}

// NewRegionComplianceFramework cria uma nova instância do framework de testes de conformidade
func NewRegionComplianceFramework(
	logger logging.Logger,
	telemetry telemetry.Provider,
	policyDir, testCasesDir string,
	regions []string,
	configProvider ConfigProvider,
) (*RegionComplianceFramework, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger é obrigatório")
	}

	if policyDir == "" {
		return nil, fmt.Errorf("diretório de políticas é obrigatório")
	}

	if testCasesDir == "" {
		return nil, fmt.Errorf("diretório de casos de teste é obrigatório")
	}

	if len(regions) == 0 {
		return nil, fmt.Errorf("pelo menos uma região deve ser especificada")
	}

	if configProvider == nil {
		return nil, fmt.Errorf("provedor de configuração é obrigatório")
	}

	framework := &RegionComplianceFramework{
		Logger:            logger,
		Telemetry:         telemetry,
		PolicyDir:         policyDir,
		TestCasesDir:      testCasesDir,
		Regions:           regions,
		ConfigProvider:    configProvider,
		ComplianceMatrices: make(map[string]*ComplianceMatrix),
	}

	// Carrega as matrizes de conformidade para todas as regiões configuradas
	for _, region := range regions {
		matrix, err := framework.loadComplianceMatrix(region)
		if err != nil {
			return nil, fmt.Errorf("falha ao carregar matriz de conformidade para região %s: %w", region, err)
		}
		framework.ComplianceMatrices[region] = matrix
	}

	return framework, nil
}

// loadComplianceMatrix carrega a matriz de conformidade para uma região específica
func (rcf *RegionComplianceFramework) loadComplianceMatrix(region string) (*ComplianceMatrix, error) {
	filePath := filepath.Join(rcf.TestCasesDir, region, "compliance_matrix.json")
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo de matriz de conformidade: %w", err)
	}

	var matrix ComplianceMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		return nil, fmt.Errorf("falha ao decodificar matriz de conformidade: %w", err)
	}

	// Carrega os casos de teste para cada categoria
	matrix.TestCases = make(map[string][]TestCase)
	testCasesDir := filepath.Join(rcf.TestCasesDir, region, "test_cases")
	
	categories, err := ioutil.ReadDir(testCasesDir)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler diretório de casos de teste: %w", err)
	}

	for _, categoryDir := range categories {
		if categoryDir.IsDir() {
			category := categoryDir.Name()
			casesPath := filepath.Join(testCasesDir, category)
			
			files, err := ioutil.ReadDir(casesPath)
			if err != nil {
				return nil, fmt.Errorf("falha ao ler casos de teste para categoria %s: %w", category, err)
			}
			
			var testCases []TestCase
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					testCasePath := filepath.Join(casesPath, file.Name())
					testCaseData, err := ioutil.ReadFile(testCasePath)
					if err != nil {
						return nil, fmt.Errorf("falha ao ler caso de teste %s: %w", testCasePath, err)
					}
					
					var testCase TestCase
					if err := json.Unmarshal(testCaseData, &testCase); err != nil {
						return nil, fmt.Errorf("falha ao decodificar caso de teste %s: %w", testCasePath, err)
					}
					
					testCases = append(testCases, testCase)
				}
			}
			
			matrix.TestCases[category] = testCases
		}
	}

	return &matrix, nil
}

// RunComplianceTests executa os testes de conformidade para uma região específica
func (rcf *RegionComplianceFramework) RunComplianceTests(
	ctx context.Context, 
	region string,
	categories []string,
	tenantID string,
) (*ComplianceSummary, []*TestResult, error) {
	startTime := time.Now()
	
	matrix, exists := rcf.ComplianceMatrices[region]
	if !exists {
		return nil, nil, fmt.Errorf("matriz de conformidade não encontrada para região: %s", region)
	}

	rcf.Logger.Info("Iniciando testes de conformidade", 
		logging.String("region", region),
		logging.String("tenant", tenantID),
		logging.StringSlice("categories", categories))

	// Prepara os casos de teste a serem executados
	var allTestCases []TestCase
	if len(categories) == 0 {
		// Se nenhuma categoria for especificada, executa todos os casos de teste
		for _, testCases := range matrix.TestCases {
			allTestCases = append(allTestCases, testCases...)
		}
	} else {
		// Caso contrário, executa apenas as categorias especificadas
		for _, category := range categories {
			testCases, exists := matrix.TestCases[category]
			if exists {
				allTestCases = append(allTestCases, testCases...)
			}
		}
	}

	if len(allTestCases) == 0 {
		rcf.Logger.Warn("Nenhum caso de teste encontrado para execução", 
			logging.String("region", region),
			logging.StringSlice("categories", categories))
		return nil, nil, nil
	}

	// Carrega as configurações regionais e do tenant
	regionConfig, err := rcf.ConfigProvider.GetRegionConfig(region)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao obter configuração da região %s: %w", region, err)
	}

	tenantConfig, err := rcf.ConfigProvider.GetTenantConfig(region, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao obter configuração do tenant %s na região %s: %w", tenantID, region, err)
	}

	// Executa os casos de teste
	var results []*TestResult
	var requirementResults = make(map[string]RequirementCompliance)
	var frameworkResults = make(map[string]FrameworkScore)

	// Inicializa os contadores de frameworks
	for _, framework := range matrix.Frameworks {
		frameworkResults[framework.ID] = FrameworkScore{
			FrameworkID: framework.ID,
			Name:        framework.Name,
		}
	}

	// Inicializa os contadores de requisitos
	for _, req := range matrix.Requirements {
		requirementResults[req.ID] = RequirementCompliance{
			RequirementID: req.ID,
			Name:          req.Name,
			FrameworkID:   req.FrameworkID,
			Criticality:   req.Criticality,
		}
	}

	// Monta o contexto de execução dos testes
	testContext := map[string]interface{}{
		"region":        region,
		"tenant_id":     tenantID,
		"region_config": regionConfig,
		"tenant_config": tenantConfig,
	}

	// Executa cada caso de teste
	for _, testCase := range allTestCases {
		result, err := rcf.runTestCase(ctx, testCase, region, tenantID, testContext)
		if err != nil {
			rcf.Logger.Error("Erro ao executar caso de teste", 
				logging.String("test_case_id", testCase.ID), 
				logging.String("region", region),
				logging.Error(err))
			continue
		}

		results = append(results, result)

		// Atualiza os resultados dos requisitos
		for _, reqID := range testCase.RequirementIDs {
			req := requirementResults[reqID]
			
			if result.Success {
				req.TestsPassed++
			} else {
				req.TestsFailed++
			}
			
			requirementResults[reqID] = req
		}
	}

	// Analisa os resultados dos requisitos
	for reqID, req := range requirementResults {
		req.Compliant = req.TestsFailed == 0 && req.TestsPassed > 0
		requirementResults[reqID] = req
		
		// Atualiza os resultados do framework associado
		frameworkScore := frameworkResults[req.FrameworkID]
		frameworkScore.TotalRequirements++
		
		if req.Compliant {
			frameworkScore.CoveredRequirements++
		} else if req.Criticality == "alta" {
			frameworkScore.CriticalIssues++
		}
		
		frameworkResults[req.FrameworkID] = frameworkScore
	}

	// Calcula as pontuações de conformidade
	for fwID, fw := range frameworkResults {
		if fw.TotalRequirements > 0 {
			fw.ComplianceScore = float64(fw.CoveredRequirements) / float64(fw.TotalRequirements) * 100
		}
		frameworkResults[fwID] = fw
	}

	// Prepara o resumo de conformidade
	summary := &ComplianceSummary{
		RegionCode:      region,
		TotalTests:      len(results),
		Timestamp:       time.Now(),
		Duration:        time.Since(startTime),
		FrameworkScores: frameworkResults,
		Requirements:    requirementResults,
	}

	// Calcula estatísticas gerais
	for _, result := range results {
		if result.Success {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
	}

	if summary.TotalTests > 0 {
		summary.Coverage = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}

	rcf.Logger.Info("Testes de conformidade concluídos", 
		logging.String("region", region),
		logging.String("tenant", tenantID),
		logging.Int("total_tests", summary.TotalTests),
		logging.Int("passed", summary.PassedTests),
		logging.Int("failed", summary.FailedTests),
		logging.Float64("coverage", summary.Coverage))

	return summary, results, nil
}

// runTestCase executa um caso de teste individual
func (rcf *RegionComplianceFramework) runTestCase(
	ctx context.Context,
	testCase TestCase,
	region string,
	tenantID string,
	testContext map[string]interface{},
) (*TestResult, error) {
	startTime := time.Now()

	// Prepara o resultado do teste
	result := &TestResult{
		TestCaseID:       testCase.ID,
		RegionCode:       region,
		RequirementIDs:   testCase.RequirementIDs,
		Input:            testCase.Input,
		ExpectedDecision: testCase.ExpectedDecision,
		PolicyPath:       testCase.PolicyPath,
	}

	// Adiciona o contexto regional e do tenant ao input do teste
	enrichedInput := make(map[string]interface{})
	for k, v := range testCase.Input {
		enrichedInput[k] = v
	}
	
	// Adiciona o contexto específico do teste
	for k, v := range testCase.Context {
		enrichedInput[k] = v
	}
	
	// Adiciona o contexto global de teste
	for k, v := range testContext {
		if _, exists := enrichedInput[k]; !exists {
			enrichedInput[k] = v
		}
	}

	// Constrói o caminho completo para a política
	policyPath := testCase.PolicyPath
	if !filepath.IsAbs(policyPath) {
		policyPath = filepath.Join(rcf.PolicyDir, policyPath)
	}

	// Prepara o módulo Rego para execução do teste
	r := rego.New(
		rego.Query(testCase.PolicyPath),
		rego.Load([]string{policyPath}, nil),
	)

	// Adiciona tracing se necessário
	var tracer *topdown.BufferTracer
	if len(testCase.Tags) > 0 && containsString(testCase.Tags, "trace") {
		tracer = topdown.NewBufferTracer()
		r = r.WithTracer(tracer)
	}

	// Prepara e executa a query
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("falha ao preparar query: %v", err)
		result.Duration = time.Since(startTime)
		return result, nil
	}

	rs, err := pq.Eval(ctx, rego.EvalInput(enrichedInput))
	if err != nil {
		// Verifica se o erro era esperado
		if len(testCase.ExpectedErrors) > 0 && containsErrorMessage(testCase.ExpectedErrors, err.Error()) {
			result.Success = true
			result.ErrorMsg = err.Error()
		} else {
			result.Success = false
			result.ErrorMsg = fmt.Sprintf("falha na avaliação: %v", err)
		}
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Processa o resultado da avaliação
	if len(rs) == 0 {
		result.ActualDecision = nil
		
		// Verifica se o resultado vazio era esperado
		if testCase.ExpectedDecision == nil {
			result.Success = true
		} else {
			result.Success = false
			result.ErrorMsg = "nenhum resultado obtido, mas um era esperado"
		}
	} else {
		result.ActualDecision = rs[0].Expressions[0].Value
		result.Success = compareDecisions(result.ActualDecision, result.ExpectedDecision)
		
		if !result.Success {
			result.ErrorMsg = fmt.Sprintf(
				"decisão real (%v) não corresponde à esperada (%v)",
				result.ActualDecision,
				result.ExpectedDecision,
			)
		}
	}

	// Adiciona o trace se disponível
	if tracer != nil {
		var tracedEvents []string
		topdown.PrettyTrace(tracer, &tracedEvents)
		result.Trace = tracedEvents
	}

	result.Duration = time.Since(startTime)
	
	// Registra o resultado do teste no telemetry
	if rcf.Telemetry != nil {
		rcf.Telemetry.RecordComplianceTestResult(testCase.ID, region, result.Success, result.Duration)
	}

	return result, nil
}

// GenerateComplianceReport gera um relatório detalhado de conformidade em formato JSON
func (rcf *RegionComplianceFramework) GenerateComplianceReport(
	summary *ComplianceSummary,
	results []*TestResult,
	outputPath string,
) error {
	// Prepara os dados do relatório
	report := map[string]interface{}{
		"summary":     summary,
		"test_results": results,
		"timestamp":   time.Now().Format(time.RFC3339),
		"region":      summary.RegionCode,
	}

	// Serializa o relatório para JSON
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("falha ao serializar relatório: %w", err)
	}

	// Cria o diretório de saída se não existir
	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("falha ao criar diretório de saída: %w", err)
	}

	// Salva o relatório no arquivo
	err = ioutil.WriteFile(outputPath, reportData, 0644)
	if err != nil {
		return fmt.Errorf("falha ao gravar relatório: %w", err)
	}

	rcf.Logger.Info("Relatório de conformidade gerado com sucesso",
		logging.String("output_path", outputPath),
		logging.String("region", summary.RegionCode),
		logging.Float64("coverage", summary.Coverage))

	return nil
}

// Função auxiliar para verificar se uma string está contida em um slice
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Função auxiliar para verificar se uma mensagem de erro corresponde a alguma das esperadas
func containsErrorMessage(expectedErrors []string, errorMsg string) bool {
	for _, expected := range expectedErrors {
		if strings.Contains(errorMsg, expected) {
			return true
		}
	}
	return false
}

// Função auxiliar para comparar decisões
func compareDecisions(actual, expected interface{}) bool {
	// Serializa ambos os valores para JSON para comparação estrutural
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return false
	}
	
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	
	return string(actualJSON) == string(expectedJSON)
}