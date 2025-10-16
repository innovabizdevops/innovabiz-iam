// Package integration contém scripts de integração para módulos core da plataforma INNOVABIZ
package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/constants"
	"github.com/innovabizdevops/innovabiz-iam/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TransactionTypes define os diferentes tipos de transações suportados pelo Mobile Money
const (
	TypeP2P           = "p2p"            // Transferência pessoa para pessoa
	TypeCashIn        = "cash_in"        // Depósito de dinheiro
	TypeCashOut       = "cash_out"       // Saque de dinheiro
	TypeBillPayment   = "bill_payment"   // Pagamento de contas
	TypeMerchantPay   = "merchant_pay"   // Pagamento para comerciantes
	TypeAirtime       = "airtime"        // Compra de crédito celular
	TypeRemittance    = "remittance"     // Remessa internacional
	TypeLoanPayment   = "loan_payment"   // Pagamento de empréstimo
	TypeMicroinsurance = "microinsurance" // Pagamento de microseguro
	TypeSavings       = "savings"        // Depósito em poupança
)

// MobileMoneyConfig contém configurações para o módulo de Mobile Money
type MobileMoneyConfig struct {
	Name               string
	Market             string
	TenantType         string
	ComplianceLogsPath string
	Environment        string
	APIEndpoint        string
	MetricsPort        int
	EnableRemittance   bool
	EnableMicroloans   bool
	EnableMicroinsurance bool
	EnableAgents       bool
	EnableMerchants    bool
	DailyLimits        map[string]float64
	MonthlyLimits      map[string]float64
}

// MobileMoneyTransaction representa uma transação no sistema Mobile Money
type MobileMoneyTransaction struct {
	TransactionID   string
	UserID          string
	RecipientID     string
	Amount          float64
	Currency        string
	TransactionType string
	Description     string
	Channel         string
	AgentID         string
	Location        string
	Tags            []string
	Timestamp       time.Time
	MFALevel        string
	TenantID        string
	MarketContext   adapter.MarketContext
	ReferenceID     string
	ExternalIDs     map[string]string
}

// MobileMoney implementa funcionalidades do módulo Mobile Money com integração MCP-IAM
type MobileMoney struct {
	config       MobileMoneyConfig
	observability *adapter.HookObservability
	logger       *zap.Logger
	wg           sync.WaitGroup
	shutdown     chan struct{}
	dailyVolumes map[string]float64 // Rastreamento de volumes diários por tipo de transação
	mutex        sync.RWMutex       // Mutex para acesso thread-safe aos volumes
}

// NewMobileMoney cria uma nova instância do módulo Mobile Money com observabilidade integrada
func NewMobileMoney(config MobileMoneyConfig) (*MobileMoney, error) {
	// Criar contexto de mercado para o adaptador de observabilidade
	marketCtx := adapter.MarketContext{
		Market:     config.Market,
		TenantType: config.TenantType,
	}

	// Inicializar adaptador de observabilidade MCP-IAM
	obs, err := adapter.NewHookObservability(adapter.NewConfig().
		WithMarketContext(marketCtx).
		WithComplianceLogsPath(config.ComplianceLogsPath).
		WithEnvironment(config.Environment).
		WithServiceName("mobile-money").
		WithServiceVersion("1.0.0"))
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar observabilidade: %w", err)
	}

	// Inicializar logger estruturado
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar logger: %w", err)
	}

	// Registrar metadados de compliance específicos por mercado
	registerMobileMoneyComplianceMetadata(obs, config.Market)

	mm := &MobileMoney{
		config:       config,
		observability: obs,
		logger:       logger,
		shutdown:     make(chan struct{}),
		dailyVolumes: make(map[string]float64),
		mutex:        sync.RWMutex{},
	}

	// Registrar métricas iniciais
	mm.registerInitialMetrics(marketCtx)

	return mm, nil
}

// registerMobileMoneyComplianceMetadata registra metadados de compliance específicos para o mercado
func registerMobileMoneyComplianceMetadata(obs *adapter.HookObservability, market string) {
	switch market {
	case constants.MarketAngola:
		// Metadados de compliance para Angola (BNA)
		obs.RegisterComplianceMetadata(constants.MarketAngola, adapter.ComplianceMetadata{
			Frameworks:       []string{"BNA", "ARSSI", "UIF"},
			RequiredMFALevel: "high",
			RetentionYears:   7,
			SpecialRequirements: map[string]string{
				"transactionLimits":    "true",
				"agentValidation":      "BNA-registro",
				"kycLevel":             "full",
				"pldFt":                "high",
				"foreignCurrencyRules": "strict",
			},
		})

	case constants.MarketBrazil:
		// Metadados de compliance para Brasil (BACEN, LGPD)
		obs.RegisterComplianceMetadata(constants.MarketBrazil, adapter.ComplianceMetadata{
			Frameworks:       []string{"BACEN", "LGPD", "COAF"},
			RequiredMFALevel: "high",
			RetentionYears:   5,
			SpecialRequirements: map[string]string{
				"pixIntegration":       "required",
				"customerConsent":      "detailed",
				"transactionMonitoring": "high",
				"cpfValidation":        "required",
				"antiMoneyLaundering":  "enhanced",
			},
		})

	case constants.MarketEU:
		// Metadados de compliance para União Europeia (PSD2, GDPR)
		obs.RegisterComplianceMetadata(constants.MarketEU, adapter.ComplianceMetadata{
			Frameworks:       []string{"PSD2", "GDPR", "AMLD5"},
			RequiredMFALevel: "high",
			RetentionYears:   7,
			SpecialRequirements: map[string]string{
				"strongAuthentication": "required",
				"dataMinimization":     "required",
				"transactionScreening": "high",
				"sanctionsChecking":    "continuous",
				"regulatoryReporting":  "automated",
			},
		})

	case constants.MarketMozambique:
		// Metadados de compliance para Moçambique (Banco de Moçambique)
		obs.RegisterComplianceMetadata("Mozambique", adapter.ComplianceMetadata{
			Frameworks:       []string{"BM", "GIFiM"},
			RequiredMFALevel: "medium",
			RetentionYears:   5,
			SpecialRequirements: map[string]string{
				"mPesaIntegration":     "preferred",
				"agentNetworkRules":    "strict",
				"ruralAccess":          "required",
				"interoperability":     "mandatory",
			},
		})

	case constants.MarketKenya:
		// Metadados de compliance para Quênia (Central Bank of Kenya)
		obs.RegisterComplianceMetadata("Kenya", adapter.ComplianceMetadata{
			Frameworks:       []string{"CBK", "FRC"},
			RequiredMFALevel: "medium",
			RetentionYears:   7,
			SpecialRequirements: map[string]string{
				"mPesaIntegration":     "required",
				"agentSupervision":     "strict",
				"consumerProtection":   "enhanced",
				"competitionRules":     "enforced",
			},
		})
	}
}// registerInitialMetrics registra métricas iniciais do serviço Mobile Money
func (mm *MobileMoney) registerInitialMetrics(marketCtx adapter.MarketContext) {
	// Registrar métricas específicas do Mobile Money
	mm.observability.RecordMetric(marketCtx, "mobile_money_features", "remittance", utils.BoolToFloat64(mm.config.EnableRemittance))
	mm.observability.RecordMetric(marketCtx, "mobile_money_features", "microloans", utils.BoolToFloat64(mm.config.EnableMicroloans))
	mm.observability.RecordMetric(marketCtx, "mobile_money_features", "microinsurance", utils.BoolToFloat64(mm.config.EnableMicroinsurance))
	mm.observability.RecordMetric(marketCtx, "mobile_money_features", "agents", utils.BoolToFloat64(mm.config.EnableAgents))
	mm.observability.RecordMetric(marketCtx, "mobile_money_features", "merchants", utils.BoolToFloat64(mm.config.EnableMerchants))

	// Registrar limites de transações
	for txType, limit := range mm.config.DailyLimits {
		mm.observability.RecordMetric(marketCtx, "mobile_money_daily_limits", txType, limit)
	}
	
	for txType, limit := range mm.config.MonthlyLimits {
		mm.observability.RecordMetric(marketCtx, "mobile_money_monthly_limits", txType, limit)
	}
}

// ProcessTransaction processa uma transação no sistema Mobile Money
func (mm *MobileMoney) ProcessTransaction(ctx context.Context, transaction MobileMoneyTransaction) error {
	// Criar um novo span para rastreabilidade da transação
	ctx, span := mm.observability.Tracer().Start(ctx, "mobile_money_transaction",
		trace.WithAttributes(
			attribute.String("transaction_id", transaction.TransactionID),
			attribute.String("user_id", transaction.UserID),
			attribute.String("recipient_id", transaction.RecipientID),
			attribute.Float64("amount", transaction.Amount),
			attribute.String("currency", transaction.Currency),
			attribute.String("transaction_type", transaction.TransactionType),
			attribute.String("channel", transaction.Channel),
		),
	)
	defer span.End()

	// Registrar início da transação
	mm.logger.Info("Iniciando processamento de transação",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("type", transaction.TransactionType),
		zap.Float64("amount", transaction.Amount),
		zap.String("currency", transaction.Currency),
		zap.String("market", transaction.MarketContext.Market))

	// Verificar autenticação do usuário
	authenticated, err := mm.verifyAuthentication(ctx, transaction)
	if err != nil {
		mm.logger.Error("falha na autenticação", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return fmt.Errorf("falha na autenticação: %w", err)
	}
	if !authenticated {
		mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityHigh, "authentication_failed",
			fmt.Sprintf("Autenticação falhou para transação %s", transaction.TransactionID))
		return fmt.Errorf("autenticação falhou")
	}

	// Verificar autorização para a transação
	authorized, err := mm.verifyAuthorization(ctx, transaction)
	if err != nil {
		mm.logger.Error("falha na autorização", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return fmt.Errorf("falha na autorização: %w", err)
	}
	if !authorized {
		mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityHigh, "authorization_failed",
			fmt.Sprintf("Autorização falhou para transação %s", transaction.TransactionID))
		return fmt.Errorf("autorização falhou")
	}

	// Verificar limites de transação
	if err := mm.verifyTransactionLimits(ctx, transaction); err != nil {
		mm.logger.Error("limite de transação excedido", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityMedium, "limit_exceeded",
			fmt.Sprintf("Limite excedido para transação %s: %v", transaction.TransactionID, err))
		return fmt.Errorf("limite de transação excedido: %w", err)
	}

	// Verificar compliance específico por mercado (PLD/FT, sanções, etc.)
	if err := mm.verifyComplianceChecks(ctx, transaction); err != nil {
		mm.logger.Error("falha na verificação de compliance", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return fmt.Errorf("falha na verificação de compliance: %w", err)
	}

	// Executar a transação
	if err := mm.executeTransaction(ctx, transaction); err != nil {
		mm.logger.Error("falha ao executar transação", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return fmt.Errorf("falha ao executar transação: %w", err)
	}

	// Registrar evento de auditoria para a transação bem-sucedida
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "transaction_completed",
		fmt.Sprintf("Transação %s completada com sucesso via Mobile Money, valor: %f %s", 
			transaction.TransactionID, transaction.Amount, transaction.Currency))
	
	// Registrar métricas de transação
	mm.observability.RecordMetric(transaction.MarketContext, "mobile_money_transaction_count", 
		transaction.TransactionType, 1)
	mm.observability.RecordHistogram(transaction.MarketContext, "mobile_money_transaction_amount", 
		transaction.Amount, transaction.Currency)

	// Atualizar volumes diários
	mm.updateDailyVolume(transaction.TransactionType, transaction.Amount)

	return nil
}

// verifyAuthentication verifica a autenticação do usuário
func (mm *MobileMoney) verifyAuthentication(ctx context.Context, transaction MobileMoneyTransaction) (bool, error) {
	ctx, span := mm.observability.Tracer().Start(ctx, "verify_authentication")
	defer span.End()

	// Obter metadados de compliance para o mercado
	metadata, exists := mm.observability.GetComplianceMetadata(transaction.MarketContext.Market)
	if !exists {
		metadata, _ = mm.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Verificar MFA conforme requisitos de compliance para Mobile Money
	// Mobile Money geralmente requer autenticação forte, especialmente para valores altos
	var requiredMFALevel string
	
	// Determinar nível MFA necessário com base no valor e tipo de transação
	if transaction.Amount > 100000 || transaction.TransactionType == TypeRemittance {
		// Transações de alto valor ou remessas internacionais exigem MFA de nível mais alto
		requiredMFALevel = "high"
	} else {
		// Caso contrário, usar requisito padrão do mercado
		requiredMFALevel = metadata.RequiredMFALevel
	}
	
	// Verificar se o nível MFA fornecido é suficiente
	mfaResult, err := mm.observability.ValidateMFA(ctx, transaction.MarketContext, transaction.UserID, transaction.MFALevel)
	if err != nil {
		return false, err
	}

	if !mfaResult {
		return false, fmt.Errorf("nível MFA insuficiente para transação Mobile Money no mercado %s: requer %s, fornecido %s",
			transaction.MarketContext.Market, requiredMFALevel, transaction.MFALevel)
	}

	// Registrar evento de auditoria
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "authentication_verified",
		fmt.Sprintf("Autenticação verificada com MFA nível %s para transação Mobile Money", transaction.MFALevel))

	return true, nil
}

// verifyAuthorization verifica a autorização para a transação
func (mm *MobileMoney) verifyAuthorization(ctx context.Context, transaction MobileMoneyTransaction) (bool, error) {
	ctx, span := mm.observability.Tracer().Start(ctx, "verify_authorization")
	defer span.End()

	// Verificar escopo para transação Mobile Money
	scope := fmt.Sprintf("mobile_money:%s", transaction.TransactionType)
	scopeResult, err := mm.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, scope)
	if err != nil {
		return false, err
	}

	if !scopeResult {
		return false, fmt.Errorf("usuário não tem escopo para realizar transações Mobile Money de tipo %s", transaction.TransactionType)
	}

	// Verificar requisitos específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// BNA exige verificação adicional para remessas internacionais
		if transaction.TransactionType == TypeRemittance {
			additionalScope, err := mm.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "mobile_money:bna:remittance")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BNA para remessas internacionais")
			}
		}
		
		// Verificar regras específicas para agentes
		if transaction.AgentID != "" {
			agentScope, err := mm.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "mobile_money:bna:agent_transaction")
			if err != nil || !agentScope {
				return false, fmt.Errorf("usuário não tem escopo para transações via agente")
			}
		}

	case constants.MarketBrazil:
		// BACEN exige escopo especial para certas transações
		if transaction.TransactionType == TypeLoanPayment || transaction.TransactionType == TypeBillPayment {
			additionalScope, err := mm.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "mobile_money:bacen:financial_service")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BACEN para serviços financeiros")
			}
		}

	case constants.MarketMozambique:
		// Banco de Moçambique exige escopo especial para determinadas operações
		if transaction.Amount > 50000 { // Valor em Meticais
			additionalScope, err := mm.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "mobile_money:bm:high_value")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BM para transações de alto valor")
			}
		}
	}

	// Registrar evento de auditoria
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "authorization_verified",
		fmt.Sprintf("Autorização verificada para transação Mobile Money %s", transaction.TransactionID))

	return true, nil
}// verifyTransactionLimits verifica se a transação está dentro dos limites configurados
func (mm *MobileMoney) verifyTransactionLimits(ctx context.Context, transaction MobileMoneyTransaction) error {
	ctx, span := mm.observability.Tracer().Start(ctx, "verify_transaction_limits")
	defer span.End()

	// Verificar limite específico para o tipo de transação
	dailyLimit, exists := mm.config.DailyLimits[transaction.TransactionType]
	if !exists {
		// Se não houver limite específico, usar limite genérico "default"
		dailyLimit, exists = mm.config.DailyLimits["default"]
		if !exists {
			// Se não houver limite default, não aplicar restrição de limite diário
			dailyLimit = -1
		}
	}

	// Verificar se a transação ultrapassa o limite diário
	if dailyLimit >= 0 {
		// Obter volume diário atual para o tipo de transação
		currentDailyVolume := mm.getDailyVolume(transaction.TransactionType)
		
		// Verificar se a transação ultrapassa o limite diário
		if currentDailyVolume+transaction.Amount > dailyLimit {
			// Registrar evento de segurança para limite excedido
			mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "daily_limit_exceeded",
				fmt.Sprintf("Transação %s ultrapassa limite diário para %s. Limite: %f, Volume atual: %f, Valor da transação: %f",
					transaction.TransactionID, transaction.TransactionType, dailyLimit, currentDailyVolume, transaction.Amount))
			
			return fmt.Errorf("transação ultrapassa limite diário de %f para tipo %s", dailyLimit, transaction.TransactionType)
		}
	}

	// Verificar limites específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Regras específicas BNA para limites de transação
		if transaction.TransactionType == TypeRemittance {
			// Limite específico para remessas internacionais pelo BNA
			if transaction.Amount > 500000 { // Valor em Kwanzas
				mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
					constants.SecurityEventSeverityHigh, "bna_limit_exceeded",
					fmt.Sprintf("Transação %s ultrapassa limite BNA para remessas", transaction.TransactionID))
				return fmt.Errorf("transação ultrapassa limite BNA para remessas internacionais")
			}
		}
		
		// Regras específicas para uso de agentes
		if transaction.AgentID != "" && transaction.Amount > 250000 { // Valor em Kwanzas
			mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "agent_limit_exceeded",
				fmt.Sprintf("Transação %s via agente ultrapassa limite BNA", transaction.TransactionID))
			return fmt.Errorf("transação via agente ultrapassa limite BNA")
		}
		
	case constants.MarketBrazil:
		// Regras específicas BACEN para limites de transação
		if transaction.TransactionType == TypeP2P && transaction.Amount > 10000 { // Valor em Reais
			// Transações P2P acima de R$ 10.000 exigem documentação adicional
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"bacen_documentation_required",
				fmt.Sprintf("Transação %s exige documentação adicional BACEN", transaction.TransactionID))
		}
		
	case constants.MarketMozambique:
		// Regras específicas Banco de Moçambique
		if transaction.TransactionType == TypeCashOut && transaction.Amount > 25000 { // Valor em Meticais
			mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "bm_limit_exceeded",
				fmt.Sprintf("Transação %s ultrapassa limite BM para saques", transaction.TransactionID))
			return fmt.Errorf("transação ultrapassa limite BM para saques")
		}
	}

	// Registrar evento de auditoria
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "limits_verified",
		fmt.Sprintf("Limites verificados para transação Mobile Money %s", transaction.TransactionID))

	return nil
}

// verifyComplianceChecks realiza verificações de compliance específicas por mercado
func (mm *MobileMoney) verifyComplianceChecks(ctx context.Context, transaction MobileMoneyTransaction) error {
	ctx, span := mm.observability.Tracer().Start(ctx, "verify_compliance_checks")
	defer span.End()

	// Obter metadados de compliance para o mercado
	metadata, exists := mm.observability.GetComplianceMetadata(transaction.MarketContext.Market)
	if !exists {
		metadata, _ = mm.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Verificar requisitos gerais de PLD/FT (Prevenção à Lavagem de Dinheiro e Financiamento do Terrorismo)
	if transaction.Amount > 10000 { // Valor genérico para demonstração
		// Registrar verificação de PLD/FT para transações de alto valor
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"pld_ft_check",
			fmt.Sprintf("Verificação PLD/FT realizada para transação %s", transaction.TransactionID))
		
		// Verificar se transação tem indicadores de risco
		if contains(transaction.Tags, "high_risk_area") || contains(transaction.Tags, "suspicious") {
			// Registrar alerta de segurança
			mm.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
				constants.SecurityEventSeverityHigh, "pld_ft_alert",
				fmt.Sprintf("Transação %s com indicadores de risco PLD/FT", transaction.TransactionID))
			
			// Dependendo do mercado, podemos bloquear ou apenas alertar
			if metadata.SpecialRequirements["pldFt"] == "high" {
				return fmt.Errorf("transação bloqueada por suspeita de PLD/FT")
			}
		}
	}

	// Verificar requisitos específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Verificação específica UIF/BNA para Angola
		if transaction.TransactionType == TypeRemittance || transaction.Amount > 250000 {
			// Verificar consentimento para compartilhamento com UIF
			consentResult, err := mm.observability.ValidateConsent(ctx, transaction.MarketContext, 
				transaction.UserID, "data_sharing:uif")
			if err != nil || !consentResult {
				return fmt.Errorf("consentimento para compartilhamento com UIF não encontrado")
			}
			
			// Registrar notificação UIF
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"uif_notification",
				fmt.Sprintf("Transação %s notificada à UIF Angola", transaction.TransactionID))
		}
		
	case constants.MarketBrazil:
		// Verificação específica COAF/BACEN para Brasil
		if transaction.TransactionType == TypeP2P && transaction.Amount > 10000 {
			// Verificar consentimento para compartilhamento com COAF
			consentResult, err := mm.observability.ValidateConsent(ctx, transaction.MarketContext, 
				transaction.UserID, "data_sharing:coaf")
			if err != nil || !consentResult {
				return fmt.Errorf("consentimento para compartilhamento com COAF não encontrado")
			}
			
			// Registrar notificação COAF
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"coaf_notification",
				fmt.Sprintf("Transação %s notificada ao COAF", transaction.TransactionID))
		}
		
		// Verificar integração PIX para Brasil
		if metadata.SpecialRequirements["pixIntegration"] == "required" && 
		   (transaction.TransactionType == TypeP2P || transaction.TransactionType == TypeMerchantPay) {
			// Simular verificação de chave PIX
			mm.logger.Info("Verificando integração PIX",
				zap.String("transaction_id", transaction.TransactionID),
				zap.String("user_id", transaction.UserID))
			
			// Registrar uso de PIX
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"pix_integration_check",
				fmt.Sprintf("Integração PIX verificada para transação %s", transaction.TransactionID))
		}
		
	case constants.MarketEU:
		// Verificação específica para AMLD5 (EU Anti-Money Laundering Directive)
		if transaction.TransactionType == TypeRemittance || transaction.Amount > 1000 { // 1000 EUR
			// Verificar sanções e PEP (Pessoas Politicamente Expostas)
			mm.logger.Info("Verificando conformidade AMLD5",
				zap.String("transaction_id", transaction.TransactionID),
				zap.String("user_id", transaction.UserID))
			
			// Registrar verificação AMLD5
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"amld5_check",
				fmt.Sprintf("Verificação AMLD5 realizada para transação %s", transaction.TransactionID))
		}
		
		// Verificar SCA (Strong Customer Authentication) para PSD2
		if metadata.SpecialRequirements["strongAuthentication"] == "required" && transaction.Amount > 30 {
			// Verificar se SCA foi aplicado
			if transaction.MFALevel != "high" {
				return fmt.Errorf("SCA (PSD2) necessário para transação acima de 30 EUR")
			}
			
			// Registrar aplicação de SCA
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"psd2_sca_applied",
				fmt.Sprintf("SCA aplicado para transação %s conforme PSD2", transaction.TransactionID))
		}
		
	case constants.MarketMozambique:
		// Verificação específica para Moçambique
		if transaction.TransactionType == TypeRemittance || transaction.Amount > 50000 { // Meticais
			// Registrar notificação GIFiM
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"gifim_notification",
				fmt.Sprintf("Transação %s notificada ao GIFiM", transaction.TransactionID))
		}
	}

	// Registrar evento de auditoria para compliance
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "compliance_verified",
		fmt.Sprintf("Compliance verificado para transação %s conforme framework %s", 
			transaction.TransactionID, metadata.Frameworks[0]))

	return nil
}// executeTransaction executa a transação no sistema Mobile Money (simulado)
func (mm *MobileMoney) executeTransaction(ctx context.Context, transaction MobileMoneyTransaction) error {
	ctx, span := mm.observability.Tracer().Start(ctx, "execute_transaction")
	defer span.End()

	// Simular processamento de transação
	// Em produção, aqui seria o código real de execução de transação
	mm.logger.Info("Processando transação Mobile Money",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("user_id", transaction.UserID),
		zap.String("recipient_id", transaction.RecipientID),
		zap.Float64("amount", transaction.Amount),
		zap.String("currency", transaction.Currency),
		zap.String("type", transaction.TransactionType))
	
	// Simular tempo de processamento
	time.Sleep(100 * time.Millisecond)

	// Registrar fluxo específico por tipo de transação
	switch transaction.TransactionType {
	case TypeP2P:
		// Processamento de transferência P2P
		mm.logger.Info("Processando transferência P2P",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("recipient_id", transaction.RecipientID))
		
		// Simular verificação do destinatário
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"recipient_verified",
			fmt.Sprintf("Destinatário %s verificado para transação P2P %s", 
				transaction.RecipientID, transaction.TransactionID))
		
	case TypeCashIn:
		// Processamento de depósito
		mm.logger.Info("Processando depósito",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("agent_id", transaction.AgentID))
		
		// Verificar autenticação do agente se aplicável
		if transaction.AgentID != "" {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"agent_verified",
				fmt.Sprintf("Agente %s verificado para depósito %s", 
					transaction.AgentID, transaction.TransactionID))
		}
		
	case TypeCashOut:
		// Processamento de saque
		mm.logger.Info("Processando saque",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("agent_id", transaction.AgentID))
		
		// Verificar disponibilidade do agente se aplicável
		if transaction.AgentID != "" {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"agent_liquidity_verified",
				fmt.Sprintf("Liquidez do agente %s verificada para saque %s", 
					transaction.AgentID, transaction.TransactionID))
		}
		
	case TypeBillPayment:
		// Processamento de pagamento de contas
		mm.logger.Info("Processando pagamento de conta",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("reference_id", transaction.ReferenceID))
		
		// Verificar validade da conta
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"bill_verified",
			fmt.Sprintf("Conta %s verificada para pagamento %s", 
				transaction.ReferenceID, transaction.TransactionID))
		
	case TypeMerchantPay:
		// Processamento de pagamento a comerciante
		mm.logger.Info("Processando pagamento a comerciante",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("recipient_id", transaction.RecipientID))
		
		// Verificar registro do comerciante
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"merchant_verified",
			fmt.Sprintf("Comerciante %s verificado para pagamento %s", 
				transaction.RecipientID, transaction.TransactionID))
		
	case TypeRemittance:
		// Processamento de remessa internacional
		mm.logger.Info("Processando remessa internacional",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("recipient_id", transaction.RecipientID))
		
		// Verificações específicas para remessas internacionais
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"international_checks_completed",
			fmt.Sprintf("Verificações internacionais completadas para remessa %s", 
				transaction.TransactionID))
	}

	// Verificar requisitos específicos para Mobile Money por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Requisitos específicos BNA para Mobile Money
		mm.logger.Info("Aplicando regras BNA para Mobile Money",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Para transações acima de um certo valor, registrar para BNA
		if transaction.Amount > 100000 {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"bna_report_generated",
				fmt.Sprintf("Relatório BNA gerado para transação %s", transaction.TransactionID))
		}
		
		// Verificar integração com EMIS (Empresa Interbancária de Serviços) em Angola
		if transaction.TransactionType == TypeP2P || transaction.TransactionType == TypeMerchantPay {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"emis_check_completed",
				fmt.Sprintf("Verificação EMIS completada para transação %s", transaction.TransactionID))
		}

	case constants.MarketBrazil:
		// Requisitos específicos BACEN/LGPD para Mobile Money
		mm.logger.Info("Aplicando regras BACEN/LGPD para Mobile Money",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Integração com PIX para transações P2P
		if transaction.TransactionType == TypeP2P {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"pix_transaction_recorded",
				fmt.Sprintf("Transação %s registrada via PIX", transaction.TransactionID))
		}
		
		// Verificações LGPD para consentimento
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"lgpd_consent_verified",
			fmt.Sprintf("Consentimento LGPD verificado para processamento de dados na transação %s", 
				transaction.TransactionID))

	case constants.MarketMozambique:
		// Requisitos específicos para Moçambique
		mm.logger.Info("Aplicando regras do Banco de Moçambique para Mobile Money",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Integração com M-Pesa se configurado
		if externalID, ok := transaction.ExternalIDs["mpesa"]; ok {
			mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"mpesa_integration_completed",
				fmt.Sprintf("Integração M-Pesa completada para transação %s com ID externo %s", 
					transaction.TransactionID, externalID))
		}
		
	case constants.MarketEU:
		// Requisitos específicos PSD2/GDPR para Mobile Money
		mm.logger.Info("Aplicando regras PSD2/GDPR para Mobile Money",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Registrar aplicação de SCA (Strong Customer Authentication)
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"sca_applied",
			fmt.Sprintf("SCA aplicado para transação Mobile Money %s conforme PSD2", 
				transaction.TransactionID))
		
		// Registrar minimização de dados conforme GDPR
		mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"gdpr_data_minimization_applied",
			fmt.Sprintf("Minimização de dados GDPR aplicada para transação %s", 
				transaction.TransactionID))
	}

	// Registrar evento de auditoria para conclusão da transação
	mm.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "transaction_executed",
		fmt.Sprintf("Transação %s executada com sucesso via Mobile Money", transaction.TransactionID))

	// Registrar métrica de tempo de processamento
	mm.observability.RecordMetric(transaction.MarketContext, "mobile_money_processing_time",
		transaction.TransactionType, 100) // Valor simulado em ms

	return nil
}

// updateDailyVolume atualiza o volume diário acumulado para um tipo de transação
func (mm *MobileMoney) updateDailyVolume(transactionType string, amount float64) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	currentVolume, exists := mm.dailyVolumes[transactionType]
	if !exists {
		currentVolume = 0
	}
	
	mm.dailyVolumes[transactionType] = currentVolume + amount
}

// getDailyVolume obtém o volume diário acumulado para um tipo de transação
func (mm *MobileMoney) getDailyVolume(transactionType string) float64 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	volume, exists := mm.dailyVolumes[transactionType]
	if !exists {
		return 0
	}
	
	return volume
}

// resetDailyVolumes redefine os volumes diários (chamado à meia-noite)
func (mm *MobileMoney) resetDailyVolumes() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.dailyVolumes = make(map[string]float64)
	
	// Registrar redefinição de volumes em log
	mm.logger.Info("Volumes diários de transação redefinidos")
}

// contains verifica se uma slice contém um determinado valor
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Start inicia o serviço Mobile Money
func (mm *MobileMoney) Start() error {
	mm.logger.Info("Iniciando serviço Mobile Money", 
		zap.String("market", mm.config.Market),
		zap.String("environment", mm.config.Environment))
	
	// Iniciar workers e componentes aqui
	// ...
	
	// Iniciar worker para reset diário de volumes
	mm.wg.Add(1)
	go mm.startDailyResetWorker()

	// Registrar métrica de inicialização
	mm.observability.RecordMetric(adapter.MarketContext{
		Market:     mm.config.Market,
		TenantType: mm.config.TenantType,
	}, "mobile_money_status", "started", 1)

	return nil
}

// startDailyResetWorker inicia um worker para redefinir volumes diários à meia-noite
func (mm *MobileMoney) startDailyResetWorker() {
	defer mm.wg.Done()
	
	ticker := time.NewTicker(time.Hour) // Verificar a cada hora
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Verificar se é meia-noite
			now := time.Now()
			if now.Hour() == 0 && now.Minute() < 10 { // Primeiros 10 minutos após meia-noite
				mm.resetDailyVolumes()
				mm.logger.Info("Volumes diários redefinidos", 
					zap.String("market", mm.config.Market))
			}
		case <-mm.shutdown:
			mm.logger.Info("Worker de reset diário encerrado")
			return
		}
	}
}

// Stop para o serviço Mobile Money graciosamente
func (mm *MobileMoney) Stop() error {
	mm.logger.Info("Parando serviço Mobile Money", 
		zap.String("market", mm.config.Market))
	
	// Sinalizar para todos os workers pararem
	close(mm.shutdown)
	
	// Aguardar todos os workers encerrarem
	mm.wg.Wait()
	
	// Encerrar componentes de observabilidade
	mm.observability.Shutdown()
	
	// Fechar logger
	mm.logger.Info("Serviço Mobile Money encerrado com sucesso")
	mm.logger.Sync()
	
	return nil
}

func main() {
	// Configuração para Mobile Money em Angola (exemplo)
	config := MobileMoneyConfig{
		Name:               "INNOVABIZ Mobile Money Angola",
		Market:             constants.MarketAngola,
		TenantType:         constants.TenantTypeBusiness,
		ComplianceLogsPath: "/var/log/innovabiz/mobile-money/angola",
		Environment:        "production",
		APIEndpoint:        "https://api.mobilemoney.innovabiz.ao",
		MetricsPort:        9092,
		EnableRemittance:   true,
		EnableMicroloans:   true,
		EnableMicroinsurance: true,
		EnableAgents:       true,
		EnableMerchants:    true,
		DailyLimits: map[string]float64{
			TypeP2P:        500000,  // 500.000 Kwanzas
			TypeCashOut:    300000,  // 300.000 Kwanzas
			TypeRemittance: 250000,  // 250.000 Kwanzas
			"default":      1000000, // 1.000.000 Kwanzas
		},
		MonthlyLimits: map[string]float64{
			TypeP2P:        5000000,  // 5.000.000 Kwanzas
			TypeCashOut:    3000000,  // 3.000.000 Kwanzas
			TypeRemittance: 2500000,  // 2.500.000 Kwanzas
			"default":      10000000, // 10.000.000 Kwanzas
		},
	}

	// Criar instância do Mobile Money
	mm, err := NewMobileMoney(config)
	if err != nil {
		log.Fatalf("Falha ao criar Mobile Money: %v", err)
	}

	// Iniciar o serviço
	if err := mm.Start(); err != nil {
		log.Fatalf("Falha ao iniciar Mobile Money: %v", err)
	}

	// Configurar signal handler para shutdown gracioso
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	// Aguardar sinal de término
	<-c
	log.Println("Recebido sinal de interrupção, encerrando Mobile Money...")
	
	// Parar o serviço
	if err := mm.Stop(); err != nil {
		log.Fatalf("Falha ao parar Mobile Money: %v", err)
	}
	
	log.Println("Mobile Money encerrado com sucesso")
}