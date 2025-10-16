// INNOVABIZ - Risk Management Integration with MCP-IAM Observability
// Implementação Multi-Mercado, Multi-Tenant e Multi-Contexto
// Compliance: ISO 27001, GDPR, LGPD, BNA, PSD2, SOX, CSL

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/constants"
	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RiskManager implementa o gerenciamento de riscos com observabilidade integrada
type RiskManager struct {
	observability *adapter.HookObservability
	mutex         sync.Mutex
	logger        *log.Logger
	rules         map[string]map[string]RiskRule // market -> rule_id -> rule
}

// RiskRule define uma regra de detecção de risco
type RiskRule struct {
	ID          string
	Name        string
	Description string
	Market      string
	RiskLevel   string
	Framework   string
	Threshold   float64
	Tags        []string
	Enabled     bool
}

// RiskAssessment representa os dados de uma avaliação de risco
type RiskAssessment struct {
	AssessmentID string
	UserID       string
	EntityID     string
	EntityType   string
	Market       string
	TenantType   string
	Timestamp    time.Time
	Indicators   map[string]float64
	Tags         []string
	RuleMatches  []string
	RiskScore    float64
	RiskLevel    string
	Signals      []RiskSignal
	MFALevel     string
}

// RiskSignal representa um sinal de risco detectado
type RiskSignal struct {
	RuleID      string
	Confidence  float64
	Description string
	Severity    string
	Timestamp   time.Time
}

// NewRiskManager cria um novo gerenciador de riscos com observabilidade
func NewRiskManager(market, tenantType string) (*RiskManager, error) {
	// Configurar o adaptador de observabilidade
	config := adapter.NewConfig().
		WithMarketContext(adapter.MarketContext{
			Market:     market,
			TenantType: tenantType,
		}).
		WithComplianceLogsPath("/var/log/innovabiz/risk-management").
		WithEnvironment("production")

	// Criar o adaptador de observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar adaptador de observabilidade: %w", err)
	}

	// Registrar metadados de compliance específicos por mercado
	registerComplianceMetadata(obs)

	// Iniciar serviço de métricas Prometheus
	err = obs.StartPrometheusServer(":9092")
	if err != nil {
		return nil, fmt.Errorf("falha ao iniciar servidor Prometheus: %w", err)
	}

	// Configurar logger
	logger := log.New(os.Stdout, fmt.Sprintf("[RiskManagement-%s-%s] ", market, tenantType), log.Ldate|log.Ltime)

	// Criar gerenciador de risco
	rm := &RiskManager{
		observability: obs,
		logger:        logger,
		rules:         make(map[string]map[string]RiskRule),
	}

	// Inicializar regras por mercado
	rm.initializeRules()

	return rm, nil
}

// registerComplianceMetadata registra metadados de compliance específicos por mercado
func registerComplianceMetadata(obs *adapter.HookObservability) {
	// Angola - Banco Nacional de Angola (BNA)
	obs.RegisterComplianceMetadata(
		constants.MarketAngola,
		"BNA",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		7,     // 7 anos de retenção
	)

	// Brasil - LGPD e BACEN
	obs.RegisterComplianceMetadata(
		constants.MarketBrazil,
		"LGPD",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		5,     // 5 anos de retenção
	)
	
	// União Europeia - GDPR
	obs.RegisterComplianceMetadata(
		constants.MarketEU,
		"GDPR",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		7,     // 7 anos de retenção
	)
	
	// China - Cybersecurity Law
	obs.RegisterComplianceMetadata(
		constants.MarketChina,
		"CSL",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		5,     // 5 anos de retenção
	)
	
	// Estados Unidos - SOX
	obs.RegisterComplianceMetadata(
		constants.MarketUSA,
		"SOX",
		true,  // Requer aprovação dual
		constants.MFALevelMedium,
		7,     // 7 anos de retenção
	)
	
	// Configuração global padrão
	obs.RegisterComplianceMetadata(
		constants.MarketGlobal,
		"ISO27001",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelMedium,
		3,     // 3 anos de retenção
	)
}// initializeRules configura as regras de risco específicas por mercado
func (rm *RiskManager) initializeRules() {
	// Configurar regras de risco para Angola (BNA)
	angolaRules := map[string]RiskRule{
		"bna-001": {
			ID:          "bna-001",
			Name:        "Detecção de Transação Suspeita BNA",
			Description: "Detecta transações que atendem aos critérios de suspeita conforme BNA Aviso 02/2018",
			Market:      constants.MarketAngola,
			RiskLevel:   "high",
			Framework:   "BNA",
			Threshold:   0.7,
			Tags:        []string{"aml", "bna", "suspicious"},
			Enabled:     true,
		},
		"bna-002": {
			ID:          "bna-002",
			Name:        "Verificação de Limites BNA",
			Description: "Verifica se as transações excedem os limites definidos pelo BNA Aviso 07/2021",
			Market:      constants.MarketAngola,
			RiskLevel:   "medium",
			Framework:   "BNA",
			Threshold:   0.6,
			Tags:        []string{"limit", "bna", "compliance"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketAngola] = angolaRules

	// Configurar regras de risco para Brasil (BACEN/LGPD)
	brazilRules := map[string]RiskRule{
		"bacen-001": {
			ID:          "bacen-001",
			Name:        "PLD/FT Circular 3.978",
			Description: "Detecta operações suspeitas conforme critérios da Circular BACEN 3.978/2020",
			Market:      constants.MarketBrazil,
			RiskLevel:   "high",
			Framework:   "BACEN",
			Threshold:   0.7,
			Tags:        []string{"aml", "bacen", "suspicious"},
			Enabled:     true,
		},
		"lgpd-001": {
			ID:          "lgpd-001",
			Name:        "Proteção de Dados Sensíveis",
			Description: "Monitora acesso a dados sensíveis conforme definido pela LGPD",
			Market:      constants.MarketBrazil,
			RiskLevel:   "high",
			Framework:   "LGPD",
			Threshold:   0.8,
			Tags:        []string{"data-protection", "lgpd", "privacy"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketBrazil] = brazilRules

	// Configurar regras de risco para União Europeia (GDPR/PSD2)
	euRules := map[string]RiskRule{
		"gdpr-001": {
			ID:          "gdpr-001",
			Name:        "Detecção de Vazamento de Dados",
			Description: "Detecta possíveis vazamentos de dados pessoais conforme GDPR",
			Market:      constants.MarketEU,
			RiskLevel:   "critical",
			Framework:   "GDPR",
			Threshold:   0.8,
			Tags:        []string{"data-breach", "gdpr", "privacy"},
			Enabled:     true,
		},
		"psd2-001": {
			ID:          "psd2-001",
			Name:        "Validação de SCA",
			Description: "Verifica se a autenticação forte do cliente (SCA) foi aplicada corretamente conforme PSD2",
			Market:      constants.MarketEU,
			RiskLevel:   "high",
			Framework:   "PSD2",
			Threshold:   0.9,
			Tags:        []string{"authentication", "psd2", "sca"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketEU] = euRules

	// Configurar regras de risco para EUA (SOX)
	usaRules := map[string]RiskRule{
		"sox-001": {
			ID:          "sox-001",
			Name:        "Controle Financeiro SOX",
			Description: "Monitora controles financeiros conforme requisitos SOX",
			Market:      constants.MarketUSA,
			RiskLevel:   "high",
			Framework:   "SOX",
			Threshold:   0.7,
			Tags:        []string{"financial", "sox", "control"},
			Enabled:     true,
		},
		"ccpa-001": {
			ID:          "ccpa-001",
			Name:        "Proteção de Dados CCPA",
			Description: "Monitora conformidade com requisitos de privacidade do CCPA",
			Market:      constants.MarketUSA,
			RiskLevel:   "medium",
			Framework:   "CCPA",
			Threshold:   0.6,
			Tags:        []string{"privacy", "ccpa", "data-protection"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketUSA] = usaRules

	// Configurar regras de risco para China (CSL)
	chinaRules := map[string]RiskRule{
		"csl-001": {
			ID:          "csl-001",
			Name:        "Localização de Dados CSL",
			Description: "Monitora conformidade com requisitos de localização de dados da CSL",
			Market:      constants.MarketChina,
			RiskLevel:   "high",
			Framework:   "CSL",
			Threshold:   0.8,
			Tags:        []string{"data-localization", "csl", "compliance"},
			Enabled:     true,
		},
		"csl-002": {
			ID:          "csl-002",
			Name:        "Avaliação de Segurança de Dados Transfronteiriços",
			Description: "Verifica se transferências de dados transfronteiriças foram aprovadas conforme CSL",
			Market:      constants.MarketChina,
			RiskLevel:   "critical",
			Framework:   "CSL",
			Threshold:   0.9,
			Tags:        []string{"cross-border", "csl", "data-security"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketChina] = chinaRules

	// Configurar regras de risco globais
	globalRules := map[string]RiskRule{
		"iso27001-001": {
			ID:          "iso27001-001",
			Name:        "Controle de Acesso ISO 27001",
			Description: "Monitora controles de acesso conforme ISO 27001 A.9",
			Market:      constants.MarketGlobal,
			RiskLevel:   "medium",
			Framework:   "ISO27001",
			Threshold:   0.6,
			Tags:        []string{"access-control", "iso27001", "security"},
			Enabled:     true,
		},
	}
	rm.rules[constants.MarketGlobal] = globalRules
}

// AssessRisk realiza uma avaliação de risco com observabilidade completa
func (rm *RiskManager) AssessRisk(ctx context.Context, assessment RiskAssessment) (*RiskAssessment, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Criar contexto de mercado
	marketCtx := adapter.MarketContext{
		Market:     assessment.Market,
		TenantType: assessment.TenantType,
	}

	// Iniciar span raiz para a avaliação de risco
	ctx, rootSpan := rm.observability.Tracer().Start(ctx, "risk_assessment",
		trace.WithAttributes(
			attribute.String("assessment_id", assessment.AssessmentID),
			attribute.String("entity_id", assessment.EntityID),
			attribute.String("entity_type", assessment.EntityType),
			attribute.String("market", assessment.Market),
			attribute.String("tenant_type", assessment.TenantType),
		),
	)
	defer rootSpan.End()

	// Registrar evento de auditoria para o início da avaliação
	rm.observability.TraceAuditEvent(ctx, marketCtx, assessment.UserID, "risk_assessment_initiated", 
		fmt.Sprintf("Iniciando avaliação de risco %s para %s tipo %s", 
			assessment.AssessmentID, assessment.EntityID, assessment.EntityType))

	// Verificar autenticação do usuário usando hook MCP-IAM
	authResult, err := rm.verifyAuthentication(ctx, marketCtx, assessment)
	if err != nil {
		// Registrar métrica de falha
		rm.observability.RecordMetric(marketCtx, "risk_assessment_failures", "auth", 1)
		
		// Registrar evento de segurança
		rm.observability.TraceSecurityEvent(ctx, marketCtx, assessment.UserID, 
			constants.SecurityEventSeverityCritical, "auth_failure",
			fmt.Sprintf("Falha de autenticação na avaliação de risco: %v", err))
		
		return nil, fmt.Errorf("falha de autenticação: %w", err)
	}

	// Verificar autorização para avaliação de risco
	authzResult, err := rm.verifyAuthorization(ctx, marketCtx, assessment)
	if err != nil {
		// Registrar métrica de falha
		rm.observability.RecordMetric(marketCtx, "risk_assessment_failures", "authz", 1)
		
		// Registrar evento de segurança
		rm.observability.TraceSecurityEvent(ctx, marketCtx, assessment.UserID, 
			constants.SecurityEventSeverityHigh, "authz_failure",
			fmt.Sprintf("Falha de autorização na avaliação de risco: %v", err))
		
		return nil, fmt.Errorf("falha de autorização: %w", err)
	}

	// Aplicar regras de avaliação de risco
	if err := rm.applyRiskRules(ctx, marketCtx, &assessment); err != nil {
		// Registrar métrica de falha
		rm.observability.RecordMetric(marketCtx, "risk_assessment_failures", "rules", 1)
		return nil, fmt.Errorf("falha na aplicação de regras: %w", err)
	}

	// Registrar métrica de sucesso
	rm.observability.RecordMetric(marketCtx, "risk_assessment_successes", assessment.EntityType, 1)
	rm.observability.RecordHistogram(marketCtx, "risk_score", assessment.RiskScore, assessment.RiskLevel)

	// Registrar evento de auditoria para conclusão da avaliação
	rm.observability.TraceAuditEvent(ctx, marketCtx, assessment.UserID, "risk_assessment_completed", 
		fmt.Sprintf("Avaliação de risco %s concluída: score %f, nível %s", 
			assessment.AssessmentID, assessment.RiskScore, assessment.RiskLevel))

	return &assessment, nil
}// verifyAuthentication verifica a autenticação do usuário usando hook MCP-IAM
func (rm *RiskManager) verifyAuthentication(ctx context.Context, marketCtx adapter.MarketContext, assessment RiskAssessment) (bool, error) {
	ctx, span := rm.observability.Tracer().Start(ctx, "auth_verification")
	defer span.End()
	
	// Obter metadados de compliance para o mercado
	metadata, exists := rm.observability.GetComplianceMetadata(marketCtx.Market)
	if !exists {
		metadata, _ = rm.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Verificar MFA conforme requisitos de compliance
	mfaResult, err := rm.observability.ValidateMFA(ctx, marketCtx, assessment.UserID, assessment.MFALevel)
	if err != nil {
		return false, err
	}

	// Verificar se o nível MFA atende aos requisitos do mercado
	if !mfaResult {
		return false, fmt.Errorf("nível MFA insuficiente para o mercado %s: requer %s, fornecido %s",
			marketCtx.Market, metadata.RequiredMFALevel, assessment.MFALevel)
	}

	// Registrar evento de auditoria
	rm.observability.TraceAuditEvent(ctx, marketCtx, assessment.UserID, "authentication_verified",
		fmt.Sprintf("Autenticação verificada com MFA nível %s para avaliação de risco", assessment.MFALevel))

	return true, nil
}

// verifyAuthorization verifica autorização para realizar a avaliação de risco
func (rm *RiskManager) verifyAuthorization(ctx context.Context, marketCtx adapter.MarketContext, assessment RiskAssessment) (bool, error) {
	ctx, span := rm.observability.Tracer().Start(ctx, "authorization_verification")
	defer span.End()

	// Verificar escopo para avaliação de risco
	scopeResult, err := rm.observability.ValidateScope(ctx, marketCtx, assessment.UserID, 
		fmt.Sprintf("risk:assess:%s", assessment.EntityType))
	if err != nil {
		return false, err
	}

	if !scopeResult {
		return false, fmt.Errorf("usuário não tem escopo para realizar avaliação de risco para %s", assessment.EntityType)
	}

	// Verificar requisitos específicos por mercado
	switch marketCtx.Market {
	case constants.MarketAngola:
		// BNA exige verificação adicional para entidades específicas
		if assessment.EntityType == "financial_institution" {
			additionalScope, err := rm.observability.ValidateScope(ctx, marketCtx, assessment.UserID, "risk:bna:financial")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BNA para avaliação de instituições financeiras")
			}
		}
	case constants.MarketBrazil:
		// BACEN exige escopo especial para certas avaliações
		for _, tag := range assessment.Tags {
			if tag == "pld_ft" {
				additionalScope, err := rm.observability.ValidateScope(ctx, marketCtx, assessment.UserID, "risk:bacen:pld_ft")
				if err != nil || !additionalScope {
					return false, fmt.Errorf("usuário não tem escopo BACEN para avaliações PLD/FT")
				}
			}
		}
	}

	// Registrar evento de auditoria
	rm.observability.TraceAuditEvent(ctx, marketCtx, assessment.UserID, "authorization_verified",
		fmt.Sprintf("Autorização verificada para avaliação de risco %s", assessment.AssessmentID))

	return true, nil
}

// applyRiskRules aplica as regras de risco específicas para o mercado
func (rm *RiskManager) applyRiskRules(ctx context.Context, marketCtx adapter.MarketContext, assessment *RiskAssessment) error {
	ctx, span := rm.observability.Tracer().Start(ctx, "apply_risk_rules")
	defer span.End()

	// Obter regras específicas para o mercado
	marketRules, exists := rm.rules[marketCtx.Market]
	if !exists {
		marketRules = rm.rules[constants.MarketGlobal]
	}

	// Aplicar todas as regras ativas
	var totalRiskScore float64
	var ruleCount int
	var highestRiskLevel string

	for _, rule := range marketRules {
		if !rule.Enabled {
			continue
		}

		// Executar avaliação da regra
		ruleMatched, riskSignal, err := rm.evaluateRule(ctx, marketCtx, rule, assessment)
		if err != nil {
			// Registrar falha na aplicação da regra
			rm.observability.TraceSecurityEvent(ctx, marketCtx, assessment.UserID,
				constants.SecurityEventSeverityMedium, "rule_evaluation_failed",
				fmt.Sprintf("Falha ao avaliar regra %s: %v", rule.ID, err))
			continue
		}

		// Se a regra foi correspondida, adicionar à lista
		if ruleMatched {
			assessment.RuleMatches = append(assessment.RuleMatches, rule.ID)
			assessment.Signals = append(assessment.Signals, *riskSignal)
			
			// Contribuir para o score total
			totalRiskScore += riskSignal.Confidence
			ruleCount++
			
			// Atualizar o nível de risco mais alto
			if highestRiskLevel == "" || 
			   (riskSignal.Severity == "critical") || 
			   (riskSignal.Severity == "high" && highestRiskLevel != "critical") || 
			   (riskSignal.Severity == "medium" && highestRiskLevel != "critical" && highestRiskLevel != "high") {
				highestRiskLevel = riskSignal.Severity
			}
			
			// Registrar evento de auditoria para regra correspondida
			rm.observability.TraceAuditEvent(ctx, marketCtx, assessment.UserID, "risk_rule_matched",
				fmt.Sprintf("Regra %s (%s) correspondida com confiança %f", 
					rule.ID, rule.Name, riskSignal.Confidence))
		}
	}

	// Calcular score final
	if ruleCount > 0 {
		assessment.RiskScore = totalRiskScore / float64(ruleCount)
		assessment.RiskLevel = highestRiskLevel
	} else {
		assessment.RiskScore = 0.0
		assessment.RiskLevel = "low"
	}

	// Registrar métricas de regras
	rm.observability.RecordMetric(marketCtx, "risk_rules_matched", marketCtx.Market, float64(len(assessment.RuleMatches)))
	rm.observability.RecordHistogram(marketCtx, "risk_signals_count", float64(len(assessment.Signals)), "count")

	return nil
}

// evaluateRule avalia uma única regra de risco contra os dados de avaliação
func (rm *RiskManager) evaluateRule(ctx context.Context, marketCtx adapter.MarketContext, rule RiskRule, assessment *RiskAssessment) (bool, *RiskSignal, error) {
	ctx, span := rm.observability.Tracer().Start(ctx, "evaluate_rule_"+rule.ID,
		trace.WithAttributes(
			attribute.String("rule_id", rule.ID),
			attribute.String("rule_name", rule.Name),
			attribute.String("framework", rule.Framework),
		),
	)
	defer span.End()

	// Simular avaliação de regra
	// Em produção, aqui seria implementada a lógica real de avaliação
	var confidence float64 = 0.0
	
	// Verificar se há indicadores relevantes para esta regra
	for _, tag := range rule.Tags {
		for indicator, value := range assessment.Indicators {
			if strings.Contains(indicator, tag) {
				confidence = math.Max(confidence, value)
			}
		}
	}

	// Verificar se supera o limiar da regra
	if confidence >= rule.Threshold {
		signal := &RiskSignal{
			RuleID:      rule.ID,
			Confidence:  confidence,
			Description: fmt.Sprintf("Regra %s (%s) acionada com confiança %f", rule.ID, rule.Name, confidence),
			Severity:    rule.RiskLevel,
			Timestamp:   time.Now(),
		}
		return true, signal, nil
	}

	return false, nil, nil
}

// Close encerra o gerenciador de riscos e libera recursos
func (rm *RiskManager) Close() {
	rm.observability.Shutdown()
}

func main() {
	// Criar gerenciadores de risco para diferentes mercados
	markets := []string{
		constants.MarketAngola,
		constants.MarketBrazil, 
		constants.MarketEU,
		constants.MarketUSA,
		constants.MarketChina,
	}
	
	tenantTypes := []string{
		constants.TenantTypeBusiness,
		constants.TenantTypeIndividual,
		constants.TenantTypeGovernment,
	}

	var managers []*RiskManager
	
	// Inicializar gerenciadores para cada combinação de mercado/tenant
	for _, market := range markets {
		for _, tenantType := range tenantTypes {
			manager, err := NewRiskManager(market, tenantType)
			if err != nil {
				log.Fatalf("Falha ao criar gerenciador para %s-%s: %v", market, tenantType, err)
			}
			managers = append(managers, manager)
			log.Printf("Gerenciador inicializado para %s-%s", market, tenantType)
		}
	}

	// Configurar signal handler para shutdown gracioso
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Executar até receber sinal de interrupção
	<-c
	log.Println("Recebido sinal de interrupção, encerrando gerenciadores...")
	
	// Fechar todos os gerenciadores
	for _, manager := range managers {
		manager.Close()
	}
	
	log.Println("Shutdown completo")
}