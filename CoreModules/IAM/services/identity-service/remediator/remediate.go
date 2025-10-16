package remediator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/innovabiz/iam/logger"
)

// RemediatorConfig representa a configuração do sistema de auto-remediação
type RemediatorConfig struct {
	// Se true, apenas simula as alterações sem aplicá-las
	DryRun bool `json:"dryRun"`
	// Diretório para salvar backups dos arquivos originais
	BackupDir string `json:"backupDir"`
	// Caminho para os arquivos de regras de remediação
	RulesDir string `json:"rulesDir"`
	// Severidade mínima para aplicar remediação automática
	MinSeverity string `json:"minSeverity"`
	// Tipos de violações a serem ignoradas
	IgnoreViolations []string `json:"ignoreViolations"`
}

// RemediationRule define uma regra de remediação para um tipo específico de violação
type RemediationRule struct {
	// ID único da regra
	ID string `json:"id"`
	// Nome descritivo da regra
	Name string `json:"name"`
	// Descrição da regra e seu propósito
	Description string `json:"description"`
	// Framework regulatório associado à regra
	Framework string `json:"framework"`
	// Região à qual a regra se aplica (AO para Angola)
	Region string `json:"region"`
	// Tipo de violação que esta regra trata
	ViolationType string `json:"violationType"`
	// Severidade da violação (alta, media, baixa)
	Severity string `json:"severity"`
	// Padrão regex para detectar o código com problema
	DetectionPattern string `json:"detectionPattern"`
	// Tipo de ação de remediação (replace, insert, delete, template)
	RemediationType string `json:"remediationType"`
	// Modelo de substituição para o código com problema
	ReplacementTemplate string `json:"replacementTemplate"`
	// Variáveis de contexto específicas da regra
	Context map[string]string `json:"context"`
	// Se true, requer revisão manual após aplicação
	RequiresReview bool `json:"requiresReview"`
	// Documentação de referência sobre a regra
	References []string `json:"references"`
}

// ViolationInfo contém informações sobre uma violação de conformidade
type ViolationInfo struct {
	// Tipo da violação (usado para mapear para regras)
	Type string `json:"type"`
	// Mensagem descritiva sobre a violação
	Message string `json:"message"`
	// Caminho do arquivo de política com violação
	PolicyPath string `json:"policyPath"`
	// Localização aproximada da violação no arquivo (opcional)
	Location *FileLocation `json:"location,omitempty"`
	// Frameworks regulatórios afetados pela violação
	Frameworks []string `json:"frameworks"`
	// Severidade da violação
	Severity string `json:"severity"`
	// Contexto adicional sobre a violação
	Context map[string]interface{} `json:"context"`
}

// FileLocation representa uma localização em um arquivo
type FileLocation struct {
	// Linha de início (1-indexed)
	StartLine int `json:"startLine"`
	// Linha de fim (1-indexed)
	EndLine int `json:"endLine"`
	// Coluna de início (0-indexed)
	StartColumn int `json:"startColumn"`
	// Coluna de fim (0-indexed)
	EndColumn int `json:"endColumn"`
}

// RemediationResult contém o resultado de uma tentativa de remediação
type RemediationResult struct {
	// Se a remediação foi aplicada com sucesso
	Success bool `json:"success"`
	// Violação que foi tratada
	Violation ViolationInfo `json:"violation"`
	// Regra que foi aplicada para remediação
	Rule *RemediationRule `json:"rule,omitempty"`
	// Mensagem de erro, se houver
	Error string `json:"error,omitempty"`
	// Caminho do arquivo de backup, se criado
	BackupPath string `json:"backupPath,omitempty"`
	// Se a remediação foi apenas simulada (dry run)
	DryRun bool `json:"dryRun"`
	// Alterações aplicadas
	Changes *FileChanges `json:"changes,omitempty"`
}

// FileChanges contém informações sobre alterações feitas em um arquivo
type FileChanges struct {
	// Conteúdo antes da remediação
	Before string `json:"before"`
	// Conteúdo após a remediação
	After string `json:"after"`
	// Linhas afetadas
	AffectedLines []int `json:"affectedLines"`
}

// Remediator é o principal motor de remediação para políticas OPA
type Remediator struct {
	// Configuração do sistema de remediação
	Config RemediatorConfig
	// Regras de remediação carregadas
	Rules []*RemediationRule
	// Logger para registrar ações e erros
	Logger *logger.Logger
}

// NewRemediator cria uma nova instância do motor de remediação
func NewRemediator(config RemediatorConfig, log *logger.Logger) (*Remediator, error) {
	// Configurar logger
	if log == nil {
		log = logger.NewLogger("remediator", logger.InfoLevel)
	}

	// Garantir que o diretório de backup existe
	if config.BackupDir != "" {
		if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
			return nil, fmt.Errorf("falha ao criar diretório de backup: %w", err)
		}
	}

	r := &Remediator{
		Config: config,
		Rules:  []*RemediationRule{},
		Logger: log,
	}

	// Carregar regras de remediação
	if config.RulesDir != "" {
		if err := r.LoadRules(config.RulesDir); err != nil {
			return nil, fmt.Errorf("falha ao carregar regras: %w", err)
		}
	}

	return r, nil
}

// LoadRules carrega regras de remediação de um diretório
func (r *Remediator) LoadRules(rulesDir string) error {
	// Verificar se o diretório existe
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return fmt.Errorf("diretório de regras não existe: %s", rulesDir)
	}

	r.Logger.Info("Carregando regras de remediação de %s", rulesDir)

	// Encontrar todos os arquivos JSON no diretório
	var ruleFiles []string
	err := filepath.Walk(rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			ruleFiles = append(ruleFiles, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("erro ao buscar arquivos de regras: %w", err)
	}

	// Carregar cada arquivo de regra
	var rules []*RemediationRule
	for _, file := range ruleFiles {
		fileRules, err := r.loadRulesFromFile(file)
		if err != nil {
			r.Logger.Error("Falha ao carregar regras de %s: %v", file, err)
			continue
		}
		rules = append(rules, fileRules...)
	}

	r.Rules = rules
	r.Logger.Info("Carregadas %d regras de remediação", len(rules))
	return nil
}

// loadRulesFromFile carrega regras de um único arquivo
func (r *Remediator) loadRulesFromFile(filePath string) ([]*RemediationRule, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo de regras: %w", err)
	}

	var rules []*RemediationRule
	if err := json.Unmarshal(data, &rules); err != nil {
		// Tentar como uma única regra
		var singleRule RemediationRule
		if err2 := json.Unmarshal(data, &singleRule); err2 != nil {
			return nil, fmt.Errorf("falha ao decodificar regras: %w", err)
		}
		rules = append(rules, &singleRule)
	}

	return rules, nil
}

// RemediateViolation tenta remediar uma única violação de conformidade
func (r *Remediator) RemediateViolation(violation ViolationInfo) (*RemediationResult, error) {
	result := &RemediationResult{
		Violation: violation,
		DryRun:    r.Config.DryRun,
		Success:   false,
	}

	// Verificar se o arquivo da política existe
	if _, err := os.Stat(violation.PolicyPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("arquivo de política não encontrado: %s", violation.PolicyPath)
		return result, fmt.Errorf(result.Error)
	}

	// Encontrar uma regra de remediação aplicável
	matchingRule := r.findMatchingRule(violation)
	if matchingRule == nil {
		result.Error = fmt.Sprintf("nenhuma regra de remediação encontrada para violação: %s", violation.Type)
		return result, fmt.Errorf(result.Error)
	}

	result.Rule = matchingRule

	// Ler o conteúdo do arquivo de política
	policyContent, err := ioutil.ReadFile(violation.PolicyPath)
	if err != nil {
		result.Error = fmt.Sprintf("falha ao ler arquivo de política: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	policyText := string(policyContent)

	// Criar um backup antes de modificar
	backupPath, err := r.createBackup(violation.PolicyPath)
	if err != nil {
		r.Logger.Warn("Falha ao criar backup: %v", err)
		// Continuar mesmo sem backup em caso de erro
	} else {
		result.BackupPath = backupPath
	}

	// Aplicar a regra de remediação
	changes := &FileChanges{
		Before: policyText,
	}

	// Compilar o padrão regex
	pattern, err := regexp.Compile(matchingRule.DetectionPattern)
	if err != nil {
		result.Error = fmt.Sprintf("padrão de detecção inválido: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	// Aplicar a remediação com base no tipo
	switch matchingRule.RemediationType {
	case "replace":
		after, affected := r.applyReplacement(policyText, pattern, matchingRule.ReplacementTemplate)
		changes.After = after
		changes.AffectedLines = affected
	case "insert":
		after, affected := r.applyInsertion(policyText, pattern, matchingRule.ReplacementTemplate, violation.Location)
		changes.After = after
		changes.AffectedLines = affected
	case "delete":
		after, affected := r.applyDeletion(policyText, pattern)
		changes.After = after
		changes.AffectedLines = affected
	case "template":
		after, affected := r.applyTemplateReplacement(policyText, pattern, matchingRule.ReplacementTemplate, violation.Context)
		changes.After = after
		changes.AffectedLines = affected
	default:
		result.Error = fmt.Sprintf("tipo de remediação não suportado: %s", matchingRule.RemediationType)
		return result, fmt.Errorf(result.Error)
	}

	// Verificar se houve alterações
	if changes.After == policyText {
		result.Error = "nenhuma alteração aplicada pela regra de remediação"
		return result, fmt.Errorf(result.Error)
	}

	// Gravar as alterações no arquivo (a menos que seja dry run)
	if !r.Config.DryRun {
		if err := ioutil.WriteFile(violation.PolicyPath, []byte(changes.After), 0644); err != nil {
			result.Error = fmt.Sprintf("falha ao gravar alterações: %v", err)
			return result, fmt.Errorf(result.Error)
		}
	}

	// Registrar resultado bem-sucedido
	result.Success = true
	result.Changes = changes

	r.Logger.Info("Remediação %s aplicada com sucesso para %s [%s]",
		matchingRule.ID, filepath.Base(violation.PolicyPath), violation.Type)

	return result, nil
}

// RemediateViolations tenta remediar múltiplas violações de conformidade
func (r *Remediator) RemediateViolations(violations []ViolationInfo) []*RemediationResult {
	results := make([]*RemediationResult, 0, len(violations))

	for _, violation := range violations {
		// Verificar se o tipo de violação está na lista de ignorados
		if r.shouldIgnoreViolation(violation.Type) {
			r.Logger.Debug("Ignorando violação do tipo %s (configurado para ignorar)", violation.Type)
			continue
		}

		// Verificar a severidade mínima configurada
		if !r.meetsMinimumSeverity(violation.Severity) {
			r.Logger.Debug("Ignorando violação com severidade %s (abaixo do mínimo %s)",
				violation.Severity, r.Config.MinSeverity)
			continue
		}

		result, err := r.RemediateViolation(violation)
		if err != nil {
			r.Logger.Error("Falha na remediação: %v", err)
		}

		results = append(results, result)
	}

	// Gerar sumário
	succeeded := 0
	for _, res := range results {
		if res.Success {
			succeeded++
		}
	}

	r.Logger.Info("Remediação concluída: %d de %d violações remediadas com sucesso",
		succeeded, len(results))

	return results
}

// findMatchingRule encontra uma regra de remediação aplicável para a violação
func (r *Remediator) findMatchingRule(violation ViolationInfo) *RemediationRule {
	for _, rule := range r.Rules {
		// Verificar se a regra se aplica à região correta
		if rule.Region != "" && rule.Region != violation.Violation.Region {
			continue
		}

		// Verificar se o tipo de violação corresponde
		if rule.ViolationType == violation.Type {
			return rule
		}
	}
	return nil
}

// createBackup cria um backup do arquivo antes da modificação
func (r *Remediator) createBackup(filePath string) (string, error) {
	if r.Config.BackupDir == "" {
		return "", nil
	}

	// Gerar nome de arquivo de backup único
	timestamp := time.Now().Format("20060102_150405")
	fileName := filepath.Base(filePath)
	backupName := fmt.Sprintf("%s.%s.bak", fileName, timestamp)
	backupPath := filepath.Join(r.Config.BackupDir, backupName)

	// Copiar o arquivo
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("falha ao ler arquivo para backup: %w", err)
	}

	if err := ioutil.WriteFile(backupPath, input, 0644); err != nil {
		return "", fmt.Errorf("falha ao escrever arquivo de backup: %w", err)
	}

	return backupPath, nil
}

// applyReplacement aplica uma substituição simples de texto
func (r *Remediator) applyReplacement(content string, pattern *regexp.Regexp, replacement string) (string, []int) {
	// Encontrar todas as ocorrências e linhas afetadas
	matches := pattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	// Calcular linhas afetadas
	lines := strings.Split(content, "\n")
	affectedLines := []int{}

	for _, match := range matches {
		// Encontrar em qual linha o match está
		start := match[0]
		prefixContent := content[:start]
		lineNum := strings.Count(prefixContent, "\n") + 1
		affectedLines = append(affectedLines, lineNum)
	}

	// Aplicar a substituição
	result := pattern.ReplaceAllString(content, replacement)
	return result, affectedLines
}

// applyInsertion insere texto em uma posição específica
func (r *Remediator) applyInsertion(content string, pattern *regexp.Regexp, insertion string, location *FileLocation) (string, []int) {
	var insertPosition int
	affectedLines := []int{}

	if location != nil && location.StartLine > 0 {
		// Usar a localização fornecida
		lines := strings.Split(content, "\n")
		lineIndex := location.StartLine - 1 // Ajustar para 0-indexed
		
		if lineIndex >= len(lines) {
			// Posição inválida
			return content, affectedLines
		}

		// Calcular posição de inserção
		insertPosition = 0
		for i := 0; i < lineIndex; i++ {
			insertPosition += len(lines[i]) + 1 // +1 para o \n
		}
		
		// Adicionar offset de coluna se especificado
		if location.StartColumn > 0 {
			insertPosition += location.StartColumn
		}
		
		affectedLines = append(affectedLines, location.StartLine)
	} else {
		// Usar o padrão regex para encontrar posição de inserção
		match := pattern.FindStringIndex(content)
		if match == nil {
			// Padrão não encontrado
			return content, affectedLines
		}
		
		insertPosition = match[1] // Inserir após o match
		
		// Calcular linha afetada
		prefixContent := content[:match[1]]
		lineNum := strings.Count(prefixContent, "\n") + 1
		affectedLines = append(affectedLines, lineNum)
	}

	// Inserir o texto na posição
	result := content[:insertPosition] + insertion + content[insertPosition:]
	return result, affectedLines
}

// applyDeletion remove texto que corresponda ao padrão
func (r *Remediator) applyDeletion(content string, pattern *regexp.Regexp) (string, []int) {
	// Encontrar todas as ocorrências e linhas afetadas
	matches := pattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	// Calcular linhas afetadas
	affectedLines := []int{}
	for _, match := range matches {
		prefixContent := content[:match[0]]
		lineNum := strings.Count(prefixContent, "\n") + 1
		affectedLines = append(affectedLines, lineNum)
	}

	// Aplicar a deleção
	result := pattern.ReplaceAllString(content, "")
	return result, affectedLines
}

// applyTemplateReplacement aplica substituição com valores dinâmicos do contexto
func (r *Remediator) applyTemplateReplacement(content string, pattern *regexp.Regexp, template string, context map[string]interface{}) (string, []int) {
	// Encontrar todas as ocorrências e linhas afetadas
	matches := pattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	// Preparar template com valores do contexto
	replacement := template
	for key, value := range context {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		replacement = strings.ReplaceAll(replacement, placeholder, valueStr)
	}

	// Calcular linhas afetadas
	affectedLines := []int{}
	for _, match := range matches {
		prefixContent := content[:match[0]]
		lineNum := strings.Count(prefixContent, "\n") + 1
		affectedLines = append(affectedLines, lineNum)
	}

	// Aplicar a substituição
	result := pattern.ReplaceAllString(content, replacement)
	return result, affectedLines
}

// shouldIgnoreViolation verifica se um tipo de violação deve ser ignorado
func (r *Remediator) shouldIgnoreViolation(violationType string) bool {
	for _, ignoredType := range r.Config.IgnoreViolations {
		if ignoredType == violationType {
			return true
		}
	}
	return false
}

// meetsMinimumSeverity verifica se a severidade atende ao mínimo configurado
func (r *Remediator) meetsMinimumSeverity(severity string) bool {
	// Se não há configuração de severidade mínima, aceitar qualquer uma
	if r.Config.MinSeverity == "" {
		return true
	}

	// Mapear severidade para valores numéricos
	severityMap := map[string]int{
		"baixa": 1,
		"media": 2,
		"alta":  3,
	}

	violationSev := severityMap[strings.ToLower(severity)]
	minSev := severityMap[strings.ToLower(r.Config.MinSeverity)]

	return violationSev >= minSev
}