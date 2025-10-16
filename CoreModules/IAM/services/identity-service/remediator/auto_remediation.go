package remediator

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
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
)

// AutoRemediator é responsável por detectar e corrigir automaticamente
// problemas comuns nas políticas OPA, focando em conformidade regional
// e requisitos regulatórios específicos de mercado
type AutoRemediator struct {
	Logger        logging.Logger
	Telemetry     telemetry.Provider
	PolicyDir     string
	RemediationDir string
	ConfigProvider ConfigProvider
	RemediationRules map[string][]RemediationRule
	ComplianceFrameworks map[string]ComplianceFramework
	BackupDir     string
	DryRun        bool
}

// ConfigProvider define a interface para obtenção de configurações
type ConfigProvider interface {
	GetRemediationConfig(region string) (map[string]interface{}, error)
	GetComplianceRules(region string, framework string) ([]ComplianceRule, error)
}

// RemediationRule define uma regra de remediação automática
type RemediationRule struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	RegionCodes []string `json:"regionCodes" yaml:"regionCodes"`
	Frameworks  []string `json:"frameworks" yaml:"frameworks"`
	Severity    string   `json:"severity" yaml:"severity"`
	Detector    Detector `json:"detector" yaml:"detector"`
	Remediation Remediation `json:"remediation" yaml:"remediation"`
	Tags        []string `json:"tags" yaml:"tags"`
}

// Detector define critérios para detectar um problema específico
type Detector struct {
	Type        string                 `json:"type" yaml:"type"`               // Tipo de detector: pattern, ast, static_analysis, test_result
	Pattern     string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"` // Padrão de texto a procurar (para type=pattern)
	Query       string                 `json:"query,omitempty" yaml:"query,omitempty"`     // Query Rego para avaliar (para type=ast)
	FailedTests []string               `json:"failedTests,omitempty" yaml:"failedTests,omitempty"` // IDs de testes que falham (para type=test_result)
	Parameters  map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`   // Parâmetros adicionais
}

// Remediation define como corrigir um problema detectado
type Remediation struct {
	Type        string                 `json:"type" yaml:"type"`                // Tipo de remediação: replace, insert, delete, template
	Pattern     string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`     // Padrão a ser substituído (para type=replace)
	Replacement string                 `json:"replacement,omitempty" yaml:"replacement,omitempty"` // Texto de substituição (para type=replace)
	Position    string                 `json:"position,omitempty" yaml:"position,omitempty"`    // Posição para inserção (para type=insert)
	Template    string                 `json:"template,omitempty" yaml:"template,omitempty"`    // Template para aplicar (para type=template)
	Parameters  map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`   // Parâmetros adicionais
	HumanReview bool                   `json:"humanReview" yaml:"humanReview"`       // Indica se requer revisão humana
	References  []string               `json:"references,omitempty" yaml:"references,omitempty"`  // Referências para documentação
}

// ComplianceFramework representa um framework de conformidade regulatória
type ComplianceFramework struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	RegionCodes []string `json:"regionCodes" yaml:"regionCodes"`
	Description string   `json:"description" yaml:"description"`
	Rules       []ComplianceRule `json:"rules" yaml:"rules"`
}

// ComplianceRule define uma regra específica de conformidade
type ComplianceRule struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Severity    string   `json:"severity" yaml:"severity"`
	ArticleRefs []string `json:"articleRefs" yaml:"articleRefs"`
	TestCases   []string `json:"testCases" yaml:"testCases"`
}

// RemediationResult armazena o resultado de uma operação de remediação
type RemediationResult struct {
	RuleID        string    `json:"ruleId"`
	PolicyFile    string    `json:"policyFile"`
	Detected      bool      `json:"detected"`
	Remediated    bool      `json:"remediated"`
	HumanReview   bool      `json:"humanReview"`
	ErrorMsg      string    `json:"errorMsg,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	BackupFile    string    `json:"backupFile,omitempty"`
	Changes       []Change  `json:"changes,omitempty"`
}

// Change representa uma alteração feita em um arquivo
type Change struct {
	Type     string `json:"type"`     // add, remove, replace
	Line     int    `json:"line"`     // Número da linha
	Original string `json:"original"` // Conteúdo original
	New      string `json:"new"`      // Novo conteúdo
}

// NewAutoRemediator cria uma nova instância do AutoRemediator
func NewAutoRemediator(
	logger logging.Logger,
	telemetry telemetry.Provider,
	policyDir string,
	remediationDir string,
	backupDir string,
	configProvider ConfigProvider,
	dryRun bool,
) (*AutoRemediator, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger é obrigatório")
	}

	if policyDir == "" {
		return nil, fmt.Errorf("diretório de políticas é obrigatório")
	}

	if configProvider == nil {
		return nil, fmt.Errorf("provedor de configuração é obrigatório")
	}

	// Cria o diretório de backup se não existir e não estiver em modo dry-run
	if !dryRun && backupDir != "" {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return nil, fmt.Errorf("falha ao criar diretório de backup: %w", err)
		}
	}

	remediation := &AutoRemediator{
		Logger:        logger,
		Telemetry:     telemetry,
		PolicyDir:     policyDir,
		RemediationDir: remediationDir,
		ConfigProvider: configProvider,
		BackupDir:     backupDir,
		DryRun:        dryRun,
		RemediationRules: make(map[string][]RemediationRule),
		ComplianceFrameworks: make(map[string]ComplianceFramework),
	}

	return remediation, nil
}

// LoadRemediationRules carrega as regras de remediação automática
func (ar *AutoRemediator) LoadRemediationRules(regions []string) error {
	ar.Logger.Info("Carregando regras de remediação", 
		logging.StringSlice("regions", regions))

	// Carrega regras para cada região
	for _, region := range regions {
		rulesDir := filepath.Join(ar.RemediationDir, "rules", strings.ToLower(region))

		// Carrega regras específicas da região
		err := ar.loadRegionRules(region, rulesDir)
		if err != nil {
			ar.Logger.Error("Falha ao carregar regras para região", 
				logging.String("region", region),
				logging.Error(err))
			return fmt.Errorf("falha ao carregar regras para região %s: %w", region, err)
		}
		
		// Carrega regras globais que se aplicam a todas as regiões
		globalRulesDir := filepath.Join(ar.RemediationDir, "rules", "global")
		err = ar.loadRegionRules(region, globalRulesDir)
		if err != nil {
			ar.Logger.Warn("Falha ao carregar regras globais para região", 
				logging.String("region", region),
				logging.Error(err))
		}
	}

	ar.Logger.Info("Regras de remediação carregadas com sucesso", 
		logging.Int("total_regions", len(ar.RemediationRules)))
	
	return nil
}