package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/innovabiz/iam/logging"
	"github.com/innovabiz/iam/telemetry"
	"github.com/innovabizdevops/innovabiz-iam/remediator"
	"github.com/open-policy-agent/opa/rego"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
)

// Estruturas para os testes de compliance e matrizes
type ComplianceMatrix struct {
	RegionCode     string        `json:"regionCode"`
	RegionName     string        `json:"regionName"`
	CrossRegional  bool          `json:"crossRegional"`
	Frameworks     []Framework   `json:"frameworks"`
	Requirements   []Requirement `json:"requirements"`
}

type Framework struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	References  []string `json:"references"`
}

type Requirement struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	FrameworkID  string   `json:"frameworkId"`
	ArticleRefs  []string `json:"articleRefs"`
	Criticality  string   `json:"criticality"`
	TestCaseIDs  []string `json:"testCaseIds"`
}

type TestCase struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	RequirementIDs   []string          `json:"requirementIds"`
	PolicyPath       string            `json:"policyPath"`
	Input            interface{}       `json:"input"`
	ExpectedDecision interface{}       `json:"expectedDecision"`
	Tags             []string          `json:"tags"`
	Context          map[string]string `json:"context"`
}

type TestResult struct {
	TestCase         TestCase         `json:"testCase"`
	ActualDecision   interface{}      `json:"actualDecision"`
	Passed           bool             `json:"passed"`
	Message          string           `json:"message,omitempty"`
	ExecutionTimeMs  int64            `json:"executionTimeMs"`
	PolicyPath       string           `json:"policyPath"`
	Requirements     []string         `json:"requirements"`
	Frameworks       []string         `json:"frameworks"`
	Criticality      string           `json:"criticality"`
	ComplianceRegion string           `json:"complianceRegion"`
	ExecutedAt       time.Time        `json:"executedAt"`
	Violations       []string         `json:"violations,omitempty"`
	Tags             []string         `json:"tags"`
}

type TestSummary struct {
	Region                string                    `json:"region"`
	RegionName            string                    `json:"regionName"`
	TotalTests            int                       `json:"totalTests"`
	PassedTests           int                       `json:"passedTests"`
	FailedTests           int                       `json:"failedTests"`
	ComplianceScore       float64                   `json:"complianceScore"`
	FrameworkScores       map[string]FrameworkScore `json:"frameworkScores"`
	RequirementsMet       []string                  `json:"requirementsMet"`
	RequirementsFailed    []string                  `json:"requirementsFailed"`
	TestResults           []*TestResult             `json:"testResults"`
	ExecutedAt            time.Time                 `json:"executedAt"`
	Duration              int64                     `json:"durationMs"`
	RemediationApplied    bool                      `json:"remediationApplied,omitempty"`
	RemediationResult     *RemediationResult        `json:"remediationResult,omitempty"`
}

type FrameworkScore struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	TotalTests       int     `json:"totalTests"`
	PassedTests      int     `json:"passedTests"`
	FailedTests      int     `json:"failedTests"`
	ComplianceScore  float64 `json:"complianceScore"`
}

// Configurações da CLI
type Config struct {
	OPAPath                  string
	TestsDir                 string
	OutputDir                string
	Regions                  []string
	Frameworks               []string
	Tags                     []string
	ReportFormat             string
	Verbose                  bool
	Remediate                bool
	DryRun                   bool
	ShowSummary              bool
	Json                     bool
	HTML                     bool
	RulesPath                string
	BackupDir                string
	MaxSeverity              string
	MinSeverity              string
	IgnoreTypes              []string
	RequireApproval          bool
	MaxRemediationsPerPolicy int
}

func main() {
	// Configuração da CLI
	config := parseFlags()
	
	// Inicializa o logger
	logger, err := setupLogger(config.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao configurar logger: %v\n", err)
		os.Exit(1)
	}

	// Exibe informações iniciais
	logger.Info("Iniciando testes de compliance",
		zap.Strings("regions", config.Regions),
		zap.Strings("frameworks", config.Frameworks),
		zap.Strings("tags", config.Tags),
		zap.String("opa_path", config.OPAPath))
	
	// Status da remediação
	if config.Remediate {
		if config.DryRun {
			fmt.Printf("%s Remediação automática habilitada (modo simulação)\n", 
				color.YellowString("⚠️"))
		} else {
			fmt.Printf("%s Remediação automática habilitada (modo real - modificará arquivos)\n", 
				color.RedString("⚠️"))
		}
	}

	// Executa os testes para cada região selecionada
	for _, region := range config.Regions {
		executarTestesRegionais(logger, config, region)
	}
}