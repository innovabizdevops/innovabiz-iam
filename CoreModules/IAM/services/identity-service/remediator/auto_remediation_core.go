package remediator

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/innovabiz/iam/common/logging"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// loadRegionRules carrega regras de remediação para uma região específica
func (ar *AutoRemediator) loadRegionRules(region, rulesDir string) error {
	// Verifica se o diretório existe
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		ar.Logger.Debug("Diretório de regras não encontrado, ignorando",
			logging.String("region", region),
			logging.String("dir", rulesDir))
		return nil
	}

	// Lista arquivos de regras (YAML e JSON)
	files, err := filepath.Glob(filepath.Join(rulesDir, "*.{yaml,yml,json}"))
	if err != nil {
		return fmt.Errorf("falha ao listar arquivos de regras: %w", err)
	}

	var rules []RemediationRule
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			ar.Logger.Warn("Falha ao ler arquivo de regras",
				logging.String("file", file),
				logging.Error(err))
			continue
		}

		var fileRules []RemediationRule
		ext := strings.ToLower(filepath.Ext(file))
		if ext == ".json" {
			if err := json.Unmarshal(data, &fileRules); err != nil {
				ar.Logger.Warn("Falha ao decodificar regras JSON",
					logging.String("file", file),
					logging.Error(err))
				continue
			}
		} else {
			if err := yaml.Unmarshal(data, &fileRules); err != nil {
				ar.Logger.Warn("Falha ao decodificar regras YAML",
					logging.String("file", file),
					logging.Error(err))
				continue
			}
		}

		// Adiciona regras somente se aplicáveis à região
		for _, rule := range fileRules {
			if len(rule.RegionCodes) == 0 || containsString(rule.RegionCodes, region) || containsString(rule.RegionCodes, "ALL") {
				rules = append(rules, rule)
			}
		}
	}

	ar.RemediationRules[region] = append(ar.RemediationRules[region], rules...)
	ar.Logger.Info("Regras de remediação carregadas para região",
		logging.String("region", region),
		logging.Int("rule_count", len(rules)))

	return nil
}

// RunRemediation executa a remediação automática para políticas específicas
func (ar *AutoRemediator) RunRemediation(
	ctx context.Context,
	region string,
	policyFiles []string,
	failedTests map[string]string,
) ([]*RemediationResult, error) {
	ar.Logger.Info("Iniciando remediação automática",
		logging.String("region", region),
		logging.Int("policy_count", len(policyFiles)))

	// Obtém as regras aplicáveis à região
	rules, ok := ar.RemediationRules[region]
	if !ok || len(rules) == 0 {
		ar.Logger.Warn("Nenhuma regra de remediação encontrada para região",
			logging.String("region", region))
		return nil, nil
	}

	var results []*RemediationResult

	// Para cada arquivo de política, aplica as regras de remediação
	for _, policyFile := range policyFiles {
		fileResults, err := ar.remediateFile(ctx, policyFile, region, rules, failedTests)
		if err != nil {
			ar.Logger.Error("Erro ao remediar arquivo",
				logging.String("file", policyFile),
				logging.Error(err))
			continue
		}
		results = append(results, fileResults...)
	}

	// Sumariza os resultados
	detected := 0
	remediated := 0
	humanReview := 0

	for _, result := range results {
		if result.Detected {
			detected++
			if result.Remediated {
				remediated++
			}
			if result.HumanReview {
				humanReview++
			}
		}
	}

	ar.Logger.Info("Remediação automática concluída",
		logging.String("region", region),
		logging.Int("policies", len(policyFiles)),
		logging.Int("issues_detected", detected),
		logging.Int("auto_remediated", remediated),
		logging.Int("human_review", humanReview))

	return results, nil
}

// remediateFile aplica remediação a um único arquivo de política
func (ar *AutoRemediator) remediateFile(
	ctx context.Context,
	policyFile string,
	region string,
	rules []RemediationRule,
	failedTests map[string]string,
) ([]*RemediationResult, error) {
	// Lê o conteúdo do arquivo
	content, err := ioutil.ReadFile(policyFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo de política: %w", err)
	}

	originalContent := string(content)
	modifiedContent := originalContent
	var results []*RemediationResult

	// Aplica cada regra de remediação ao arquivo
	for _, rule := range rules {
		result := &RemediationResult{
			RuleID:     rule.ID,
			PolicyFile: policyFile,
			Timestamp:  time.Now(),
		}

		// Verifica se a regra se aplica ao arquivo
		applies, err := ar.ruleApplies(ctx, rule, policyFile, modifiedContent, failedTests)
		if err != nil {
			result.ErrorMsg = fmt.Sprintf("falha ao verificar aplicabilidade: %v", err)
			results = append(results, result)
			continue
		}

		if !applies {
			// Regra não se aplica, continua para a próxima
			continue
		}

		result.Detected = true
		result.HumanReview = rule.Remediation.HumanReview

		// Se a remediação requer revisão humana e não estamos em modo de teste,
		// apenas registra e continua
		if rule.Remediation.HumanReview && !ar.DryRun {
			result.ErrorMsg = "remediação requer revisão humana"
			results = append(results, result)
			continue
		}

		// Aplica a remediação
		newContent, changes, err := ar.applyRemediation(rule.Remediation, modifiedContent, policyFile, region)
		if err != nil {
			result.ErrorMsg = fmt.Sprintf("falha ao aplicar remediação: %v", err)
			results = append(results, result)
			continue
		}

		result.Changes = changes

		// Se o conteúdo foi modificado e não estamos em modo dry-run, salva as alterações
		if newContent != modifiedContent && !ar.DryRun {
			// Cria backup do arquivo original (apenas para a primeira modificação)
			if modifiedContent == originalContent && ar.BackupDir != "" {
				backupFile := filepath.Join(
					ar.BackupDir,
					fmt.Sprintf("%s.%s.bak", filepath.Base(policyFile), time.Now().Format("20060102_150405")),
				)
				
				if err := ioutil.WriteFile(backupFile, []byte(originalContent), 0644); err != nil {
					ar.Logger.Warn("Falha ao criar backup do arquivo",
						logging.String("policy", policyFile),
						logging.String("backup", backupFile),
						logging.Error(err))
				} else {
					result.BackupFile = backupFile
				}
			}

			// Escreve as alterações no arquivo
			if err := ioutil.WriteFile(policyFile, []byte(newContent), 0644); err != nil {
				result.ErrorMsg = fmt.Sprintf("falha ao salvar alterações: %v", err)
				results = append(results, result)
				continue
			}

			// Atualiza o conteúdo para as próximas regras
			modifiedContent = newContent
			result.Remediated = true
		} else if newContent != modifiedContent {
			// Em modo dry-run apenas indica que seria remediado
			result.Remediated = true
		}

		results = append(results, result)
	}

	return results, nil
}

// ruleApplies verifica se uma regra de remediação se aplica ao arquivo
func (ar *AutoRemediator) ruleApplies(
	ctx context.Context,
	rule RemediationRule,
	policyFile string,
	content string,
	failedTests map[string]string,
) (bool, error) {
	detector := rule.Detector

	switch detector.Type {
	case "pattern":
		// Verifica se o padrão existe no conteúdo
		if detector.Pattern == "" {
			return false, fmt.Errorf("padrão não especificado para detector tipo pattern")
		}
		
		re, err := regexp.Compile(detector.Pattern)
		if err != nil {
			return false, fmt.Errorf("padrão regexp inválido: %w", err)
		}
		
		return re.MatchString(content), nil

	case "ast":
		// Analisa a política usando o AST do OPA
		if detector.Query == "" {
			return false, fmt.Errorf("query não especificada para detector tipo ast")
		}
		
		// Compila a política para AST
		module, err := ast.ParseModule(policyFile, content)
		if err != nil {
			return false, fmt.Errorf("falha ao analisar política: %w", err)
		}
		
		// Executa a query Rego contra o AST
		r := rego.New(
			rego.Query(detector.Query),
			rego.ParsedModule(module),
		)
		
		rs, err := r.Eval(ctx)
		if err != nil {
			return false, fmt.Errorf("falha ao avaliar query: %w", err)
		}
		
		// Se a query retornou algum resultado, a regra se aplica
		return len(rs) > 0 && len(rs[0].Expressions) > 0, nil

	case "test_result":
		// Verifica se algum dos testes falhos está na lista de testes da regra
		if len(detector.FailedTests) == 0 {
			return false, fmt.Errorf("lista de testes não especificada para detector tipo test_result")
		}
		
		// Extrai o ID base da política para comparar com os testes falhos
		base := filepath.Base(policyFile)
		ext := filepath.Ext(base)
		policyID := strings.TrimSuffix(base, ext)
		
		for testID, policyPath := range failedTests {
			if strings.Contains(policyPath, policyID) && containsString(detector.FailedTests, testID) {
				return true, nil
			}
		}
		
		return false, nil

	default:
		return false, fmt.Errorf("tipo de detector não suportado: %s", detector.Type)
	}
}

// applyRemediation aplica a estratégia de remediação ao conteúdo
func (ar *AutoRemediator) applyRemediation(
	remediation Remediation,
	content string,
	policyFile string,
	region string,
) (string, []Change, error) {
	var changes []Change

	switch remediation.Type {
	case "replace":
		if remediation.Pattern == "" {
			return content, nil, fmt.Errorf("padrão não especificado para remediação tipo replace")
		}
		
		re, err := regexp.Compile(remediation.Pattern)
		if err != nil {
			return content, nil, fmt.Errorf("padrão regexp inválido: %w", err)
		}
		
		// Prepara a substituição, processando variáveis se existirem
		replacement := remediation.Replacement
		
		// Aplica substituições de variáveis (ex: ${region}, ${current_date})
		replacement = strings.ReplaceAll(replacement, "${region}", region)
		replacement = strings.ReplaceAll(replacement, "${current_date}", time.Now().Format("2006-01-02"))
		
		// Divide em linhas para rastrear mudanças
		lines := strings.Split(content, "\n")
		modifiedContent := content
		matches := re.FindAllStringIndex(content, -1)
		
		for _, match := range matches {
			original := content[match[0]:match[1]]
			
			// Rastreia a linha da mudança
			lineNum := 1
			for pos := 0; pos < match[0]; pos++ {
				if content[pos] == '\n' {
					lineNum++
				}
			}
			
			changes = append(changes, Change{
				Type:     "replace",
				Line:     lineNum,
				Original: original,
				New:      replacement,
			})
		}
		
		// Aplica a substituição
		modifiedContent = re.ReplaceAllString(content, replacement)
		return modifiedContent, changes, nil

	case "insert":
		// Implementação de inserção de conteúdo
		position := remediation.Position // before_line, after_line, start, end
		lines := strings.Split(content, "\n")
		
		lineNum := 1
		if params, ok := remediation.Parameters["line_number"]; ok {
			if ln, ok := params.(float64); ok {
				lineNum = int(ln)
				if lineNum < 1 {
					lineNum = 1
				} else if lineNum > len(lines) {
					lineNum = len(lines)
				}
			}
		}
		
		// Prepara o conteúdo a inserir
		insertContent := remediation.Replacement
		insertContent = strings.ReplaceAll(insertContent, "${region}", region)
		insertContent = strings.ReplaceAll(insertContent, "${current_date}", time.Now().Format("2006-01-02"))
		
		switch position {
		case "before_line":
			changes = append(changes, Change{
				Type:     "add",
				Line:     lineNum,
				Original: "",
				New:      insertContent,
			})
			
			newLines := make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:lineNum-1]...)
			newLines = append(newLines, insertContent)
			newLines = append(newLines, lines[lineNum-1:]...)
			return strings.Join(newLines, "\n"), changes, nil
			
		case "after_line":
			changes = append(changes, Change{
				Type:     "add",
				Line:     lineNum + 1,
				Original: "",
				New:      insertContent,
			})
			
			newLines := make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:lineNum]...)
			newLines = append(newLines, insertContent)
			newLines = append(newLines, lines[lineNum:]...)
			return strings.Join(newLines, "\n"), changes, nil
			
		case "start":
			changes = append(changes, Change{
				Type:     "add",
				Line:     1,
				Original: "",
				New:      insertContent,
			})
			
			return insertContent + "\n" + content, changes, nil
			
		case "end":
			changes = append(changes, Change{
				Type:     "add",
				Line:     len(lines) + 1,
				Original: "",
				New:      insertContent,
			})
			
			return content + "\n" + insertContent, changes, nil
			
		default:
			return content, nil, fmt.Errorf("posição de inserção não suportada: %s", position)
		}
		
	case "delete":
		if remediation.Pattern == "" {
			return content, nil, fmt.Errorf("padrão não especificado para remediação tipo delete")
		}
		
		re, err := regexp.Compile(remediation.Pattern)
		if err != nil {
			return content, nil, fmt.Errorf("padrão regexp inválido: %w", err)
		}
		
		// Divide em linhas para rastrear mudanças
		lines := strings.Split(content, "\n")
		modifiedContent := content
		matches := re.FindAllStringIndex(content, -1)
		
		for _, match := range matches {
			original := content[match[0]:match[1]]
			
			// Rastreia a linha da mudança
			lineNum := 1
			for pos := 0; pos < match[0]; pos++ {
				if content[pos] == '\n' {
					lineNum++
				}
			}
			
			changes = append(changes, Change{
				Type:     "remove",
				Line:     lineNum,
				Original: original,
				New:      "",
			})
		}
		
		// Aplica a remoção
		modifiedContent = re.ReplaceAllString(content, "")
		return modifiedContent, changes, nil
		
	case "template":
		// Implementação para substituição baseada em template
		if remediation.Template == "" {
			return content, nil, fmt.Errorf("template não especificado para remediação tipo template")
		}
		
		// Carrega o template
		templatePath := remediation.Template
		if !filepath.IsAbs(templatePath) {
			templatePath = filepath.Join(ar.RemediationDir, "templates", region, templatePath)
		}
		
		templateData, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return content, nil, fmt.Errorf("falha ao ler template: %w", err)
		}
		
		// Aplica variáveis ao template
		template := string(templateData)
		template = strings.ReplaceAll(template, "${region}", region)
		template = strings.ReplaceAll(template, "${current_date}", time.Now().Format("2006-01-02"))
		template = strings.ReplaceAll(template, "${policy_file}", filepath.Base(policyFile))
		
		// Parâmetros adicionais do template
		for key, value := range remediation.Parameters {
			if strValue, ok := value.(string); ok {
				template = strings.ReplaceAll(template, "${"+key+"}", strValue)
			}
		}
		
		// Adiciona mudança
		changes = append(changes, Change{
			Type:     "replace",
			Line:     1,
			Original: content,
			New:      template,
		})
		
		return template, changes, nil
		
	default:
		return content, nil, fmt.Errorf("tipo de remediação não suportado: %s", remediation.Type)
	}
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