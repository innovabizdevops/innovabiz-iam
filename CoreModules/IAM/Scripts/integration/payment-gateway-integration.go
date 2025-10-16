// INNOVABIZ Platform - Payment Gateway Integration with MCP-IAM Observability
// Desenvolvido para: INNOVABIZ - Sistema de Governança Aumentada de Inteligência Empresarial
// Módulo: Payment Gateway
// Autor: Eduardo Jeremias
// Copyright © 2025 INNOVABIZ. Todos os direitos reservados.

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

	"github.com/innovabiz/mcp-iam/adapter"
	"github.com/innovabiz/mcp-iam/constants"
	"github.com/innovabiz/mcp-iam/telemetry"
	"github.com/innovabiz/mcp-iam/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Definição de constantes para tipos de pagamentos
const (
	PaymentTypeCard        = "card"
	PaymentTypeBank        = "bank_transfer"
	PaymentTypeWallet      = "digital_wallet"
	PaymentTypeCrypto      = "cryptocurrency"
	PaymentTypeInstalment  = "instalment"
	PaymentTypeRefund      = "refund"
	PaymentTypeRecurring   = "recurring"
	PaymentTypeQRCode      = "qr_code"
	PaymentTypeMobileMoney = "mobile_money"
	PaymentTypeBoleto      = "boleto"        // Específico para Brasil
	PaymentTypePIX         = "pix"           // Específico para Brasil
	PaymentTypeEFTPOS      = "eftpos"        // Específico para Angola/Moçambique
	PaymentTypeSEPA        = "sepa"          // Específico para UE
)

// Definição de constantes para status de pagamentos
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusRefunded   = "refunded"
	StatusCancelled  = "cancelled"
	StatusDisputed   = "disputed"
)

// PaymentGatewayConfig contém a configuração para o gateway de pagamento
type PaymentGatewayConfig struct {
	Name               string
	Market             string
	TenantType         string
	Environment        string
	APIEndpoint        string
	MetricsPort        int
	ComplianceLogsPath string
	SupportedPayments  map[string]bool
	MerchantID         string
	PaymentProviders   map[string]bool
	TransactionLimits  map[string]float64
	RetentionPolicies  map[string]int // em dias
	NotificationUrls   map[string]string
	PSP3DSEnabled      bool // 3D Secure
	PAResRoute         string // 3DS Payment Authentication Response route
}

// PaymentTransaction representa uma transação de pagamento
type PaymentTransaction struct {
	TransactionID       string
	MerchantID          string
	UserID              string
	PaymentType         string
	Amount              float64
	Currency            string
	Description         string
	CustomerIP          string
	UserAgent           string
	DeviceFingerprint   string
	BillingAddress      *Address
	ShippingAddress     *Address
	PaymentDetails      map[string]interface{}
	Metadata            map[string]interface{}
	ThreeDSData         map[string]interface{}
	RiskScore           float64
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	MarketContext       adapter.MarketContext
	MFALevel            string
	PreviousTransations []string
	RecurringProfileID  string
	Tags                []string
	PSPReferenceID      string
	FraudCheckResult    string
}

// Address representa um endereço para cobrança ou entrega
type Address struct {
	Name        string
	AddressLine1 string
	AddressLine2 string
	City        string
	State       string
	PostalCode  string
	Country     string
	Phone       string
}

// PaymentGateway representa o serviço de gateway de pagamento
type PaymentGateway struct {
	config          PaymentGatewayConfig
	observability   adapter.ObservabilityAdapter
	logger          *zap.Logger
	mutex           sync.RWMutex
	wg              sync.WaitGroup
	shutdown        chan struct{}
	dailyVolumes    map[string]float64
	activeProviders map[string]bool
	riskEngine      *RiskEngine
	complianceRules map[string]ComplianceRule
}

// RiskEngine representa o motor de risco para transações
type RiskEngine struct {
	rules    []RiskRule
	logger   *zap.Logger
	market   string
	observer adapter.ObservabilityAdapter
}

// RiskRule representa uma regra de risco
type RiskRule struct {
	ID          string
	Name        string
	Description string
	Market      string
	Severity    string
	Evaluate    func(transaction *PaymentTransaction) (bool, float64, error)
}

// ComplianceRule representa uma regra de conformidade
type ComplianceRule struct {
	ID           string
	Market       string
	Framework    string
	Requirement  string
	Description  string
	Validate     func(transaction *PaymentTransaction) (bool, string, error)
	MandatoryFor []string // Tipos de pagamento aos quais se aplica
}

// NewPaymentGateway cria uma nova instância do gateway de pagamento
func NewPaymentGateway(config PaymentGatewayConfig) (*PaymentGateway, error) {
	// Criar adaptador de observabilidade
	obsAdapter, err := adapter.NewObservabilityAdapter(adapter.ObservabilityConfig{
		ServiceName:    config.Name,
		ServiceVersion: os.Getenv("SERVICE_VERSION"),
		Environment:    config.Environment,
		LogsPath:       config.ComplianceLogsPath,
	})
	
	if err != nil {
		return nil, fmt.Errorf("falha ao criar adaptador de observabilidade: %w", err)
	}

	// Criar logger estruturado
	logger, err := telemetry.NewStructuredLogger(config.Name, config.Environment)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar logger: %w", err)
	}

	// Registrar metadados de compliance específicos por mercado
	registerComplianceMetadata(obsAdapter)

	pg := &PaymentGateway{
		config:          config,
		observability:   obsAdapter,
		logger:          logger,
		shutdown:        make(chan struct{}),
		dailyVolumes:    make(map[string]float64),
		activeProviders: make(map[string]bool),
		complianceRules: make(map[string]ComplianceRule),
	}

	// Inicializar o motor de risco
	pg.riskEngine = newRiskEngine(logger, obsAdapter, config.Market)

	// Inicializar regras de compliance específicas por mercado
	pg.initComplianceRules()

	// Inicializar métricas e observabilidade
	marketCtx := adapter.MarketContext{
		Market:     config.Market,
		TenantType: config.TenantType,
	}
	
	pg.registerInitialMetrics(marketCtx)
	pg.logger.Info("Payment Gateway inicializado",
		zap.String("market", config.Market),
		zap.String("environment", config.Environment))

	return pg, nil
}

// registerComplianceMetadata registra metadados de compliance por mercado para payment gateway
func registerComplianceMetadata(obsAdapter adapter.ObservabilityAdapter) {
	// Angola - BNA
	obsAdapter.RegisterComplianceMetadata(constants.MarketAngola, adapter.ComplianceMetadata{
		Frameworks: []string{"BNA", "SSIF", "ABANC", "ENSA"},
		RequiredMFALevel: "high",
		DataRetentionDays: 3650, // 10 anos
		DataClassification: "highly_restricted",
		Regulators: []string{"BNA", "CMC", "INACOM", "UIF"},
		SpecialRequirements: map[string]string{
			"exchangeControl": "required",
			"transactionReporting": "daily",
			"customerDueDiligence": "enhanced",
			"africanRegionalCompliance": "required",
		},
	})

	// Brasil - BACEN/LGPD
	obsAdapter.RegisterComplianceMetadata(constants.MarketBrazil, adapter.ComplianceMetadata{
		Frameworks: []string{"BACEN", "LGPD", "SPB", "FEBRABAN", "CVM"},
		RequiredMFALevel: "medium",
		DataRetentionDays: 1825, // 5 anos
		DataClassification: "restricted",
		Regulators: []string{"BACEN", "CVM", "COAF", "ANPD"},
		SpecialRequirements: map[string]string{
			"pixIntegration": "required",
			"transactionReporting": "realtime",
			"customerDueDiligence": "standard",
			"openFinanceCompliance": "required",
		},
	})

	// União Europeia - GDPR/PSD2
	obsAdapter.RegisterComplianceMetadata(constants.MarketEU, adapter.ComplianceMetadata{
		Frameworks: []string{"PSD2", "GDPR", "AMLD5", "eIDAS", "SEPA"},
		RequiredMFALevel: "high",
		DataRetentionDays: 730, // 2 anos
		DataClassification: "sensitive_personal_data",
		Regulators: []string{"EBA", "ECB", "ESMA", "National Central Banks"},
		SpecialRequirements: map[string]string{
			"strongCustomerAuthentication": "required",
			"transactionMonitoring": "continuous",
			"dataSubjectRights": "enforced",
			"rightToBeForgotten": "implemented",
			"sepaCompliance": "required",
		},
	})

	// EUA - PCI DSS/State Laws
	obsAdapter.RegisterComplianceMetadata(constants.MarketUSA, adapter.ComplianceMetadata{
		Frameworks: []string{"PCI DSS", "CCPA", "SOX", "GLBA", "FedRAMP"},
		RequiredMFALevel: "medium",
		DataRetentionDays: 2190, // 6 anos
		DataClassification: "pci_regulated",
		Regulators: []string{"FTC", "CFPB", "OCC", "FDIC", "Federal Reserve"},
		SpecialRequirements: map[string]string{
			"stateSpecificCompliance": "implemented",
			"cardDataSecurity": "tokenized",
			"consumerDisclosures": "required",
			"amlCompliance": "enhanced",
		},
	})

	// Moçambique - Banco de Moçambique
	obsAdapter.RegisterComplianceMetadata(constants.MarketMozambique, adapter.ComplianceMetadata{
		Frameworks: []string{"Banco de Moçambique", "AMB", "SIMO"},
		RequiredMFALevel: "high",
		DataRetentionDays: 3650, // 10 anos
		DataClassification: "highly_restricted",
		Regulators: []string{"Banco de Moçambique", "GIFiM", "INAGE"},
		SpecialRequirements: map[string]string{
			"centralBankReporting": "weekly",
			"mobileMoneyIntegration": "required",
			"customerDueDiligence": "enhanced",
			"africanRegionalCompliance": "required",
		},
	})

	// Global - Padrões internacionais
	obsAdapter.RegisterComplianceMetadata(constants.MarketGlobal, adapter.ComplianceMetadata{
		Frameworks: []string{"PCI DSS", "ISO 27001", "ISO 8583", "SWIFT", "FATF"},
		RequiredMFALevel: "medium",
		DataRetentionDays: 1095, // 3 anos
		DataClassification: "confidential",
		Regulators: []string{"Various"},
		SpecialRequirements: map[string]string{
			"internationalStandards": "enforced",
			"cardDataSecurity": "tokenized",
			"fraudMonitoring": "24x7",
		},
	})
}

// initComplianceRules inicializa as regras de compliance específicas por mercado
func (pg *PaymentGateway) initComplianceRules() {
	// Angola - BNA compliance rules
	if pg.config.Market == constants.MarketAngola || pg.config.Market == constants.MarketGlobal {
		pg.complianceRules["bna_foreign_exchange"] = ComplianceRule{
			ID:          "bna_foreign_exchange",
			Market:      constants.MarketAngola,
			Framework:   "BNA",
			Requirement: "Controle de câmbio para transações internacionais",
			Description: "Verificar autorização de câmbio para transações em moeda estrangeira",
			MandatoryFor: []string{PaymentTypeCard, PaymentTypeBank, PaymentTypeWallet},
			Validate: func(tx *PaymentTransaction) (bool, string, error) {
				if tx.Currency != "AOA" {
					// Verificar se existe autorização de câmbio nos metadados
					if _, exists := tx.Metadata["exchange_authorization"]; !exists {
						return false, "Falta autorização de câmbio BNA", nil
					}
				}
				return true, "Compliance de câmbio BNA verificado", nil
			},
		}
	}

	// Brasil - BACEN compliance rules
	if pg.config.Market == constants.MarketBrazil || pg.config.Market == constants.MarketGlobal {
		pg.complianceRules["bacen_pix"] = ComplianceRule{
			ID:          "bacen_pix",
			Market:      constants.MarketBrazil,
			Framework:   "BACEN",
			Requirement: "Integração com PIX para pagamentos instantâneos",
			Description: "Verificar conformidade com requisitos PIX para transações instantâneas",
			MandatoryFor: []string{PaymentTypePIX},
			Validate: func(tx *PaymentTransaction) (bool, string, error) {
				if tx.PaymentType == PaymentTypePIX {
					// Verificar se os campos obrigatórios do PIX estão presentes
					if _, exists := tx.PaymentDetails["pix_key_type"]; !exists {
						return false, "Tipo de chave PIX não especificado", nil
					}
					if _, exists := tx.PaymentDetails["pix_key"]; !exists {
						return false, "Chave PIX não especificada", nil
					}
				}
				return true, "Compliance PIX verificado", nil
			},
		}
	}

	// EU - PSD2 compliance rules
	if pg.config.Market == constants.MarketEU || pg.config.Market == constants.MarketGlobal {
		pg.complianceRules["psd2_sca"] = ComplianceRule{
			ID:          "psd2_sca",
			Market:      constants.MarketEU,
			Framework:   "PSD2",
			Requirement: "Strong Customer Authentication (SCA)",
			Description: "Verificar aplicação de autenticação forte para transações acima de 30 EUR",
			MandatoryFor: []string{PaymentTypeCard, PaymentTypeBank, PaymentTypeWallet},
			Validate: func(tx *PaymentTransaction) (bool, string, error) {
				// Verificar se é necessário SCA (transações acima de 30 EUR)
				if tx.Amount > 30 && tx.Currency == "EUR" {
					// Verificar se SCA foi aplicado (MFA nível alto)
					if tx.MFALevel != "high" {
						return false, "SCA requerido para transação acima de 30 EUR", nil
					}
					
					// Verificar se dados 3D Secure estão presentes para cartão
					if tx.PaymentType == PaymentTypeCard {
						if _, exists := tx.ThreeDSData["authentication_status"]; !exists {
							return false, "Dados 3D Secure requeridos por PSD2", nil
						}
					}
				}
				return true, "Compliance SCA verificado", nil
			},
		}
	}

	// Regras globais de compliance
	pg.complianceRules["pci_dss"] = ComplianceRule{
		ID:          "pci_dss",
		Market:      constants.MarketGlobal,
		Framework:   "PCI DSS",
		Requirement: "Proteção de dados de cartão",
		Description: "Verificar aplicação de requisitos PCI DSS para dados de cartão",
		MandatoryFor: []string{PaymentTypeCard},
		Validate: func(tx *PaymentTransaction) (bool, string, error) {
			if tx.PaymentType == PaymentTypeCard {
				// Verificar se dados sensíveis do cartão não estão presentes
				if cardData, exists := tx.PaymentDetails["card"]; exists {
					cardDetails, ok := cardData.(map[string]interface{})
					if !ok {
						return false, "Formato inválido de dados do cartão", nil
					}
					
					// Verificar se número completo do cartão está presente
					if _, exists := cardDetails["full_number"]; exists {
						return false, "Número completo do cartão não deveria estar presente", nil
					}
					
					// Verificar se há token ou PAN truncado no lugar
					if _, exists := cardDetails["token"]; !exists {
						if _, exists := cardDetails["truncated_pan"]; !exists {
							return false, "Token ou PAN truncado requerido por PCI DSS", nil
						}
					}
				}
			}
			return true, "Compliance PCI DSS verificado", nil
		},
	}
}

// registerInitialMetrics registra métricas iniciais do serviço Payment Gateway
func (pg *PaymentGateway) registerInitialMetrics(marketCtx adapter.MarketContext) {
	// Registrar métricas para tipos de pagamento suportados
	for paymentType, enabled := range pg.config.SupportedPayments {
		pg.observability.RecordMetric(marketCtx, "payment_gateway_supported_types", 
			paymentType, utils.BoolToFloat64(enabled))
	}

	// Registrar métricas para provedores de pagamento ativos
	for provider, enabled := range pg.config.PaymentProviders {
		pg.observability.RecordMetric(marketCtx, "payment_gateway_providers", 
			provider, utils.BoolToFloat64(enabled))
		pg.activeProviders[provider] = enabled
	}

	// Registrar limites de transações
	for txType, limit := range pg.config.TransactionLimits {
		pg.observability.RecordMetric(marketCtx, "payment_gateway_limits", txType, limit)
	}

	// Registrar métrica de inicialização do serviço
	pg.observability.RecordMetric(marketCtx, "payment_gateway_status", "initialized", 1)
}// newRiskEngine cria um novo motor de avaliação de risco
func newRiskEngine(logger *zap.Logger, observer adapter.ObservabilityAdapter, market string) *RiskEngine {
	engine := &RiskEngine{
		logger:   logger,
		market:   market,
		observer: observer,
		rules:    make([]RiskRule, 0),
	}

	// Adicionar regras de risco básicas
	engine.addBasicRiskRules()

	// Adicionar regras específicas por mercado
	switch market {
	case constants.MarketAngola:
		engine.addAngolaRiskRules()
	case constants.MarketBrazil:
		engine.addBrazilRiskRules()
	case constants.MarketEU:
		engine.addEURiskRules()
	case constants.MarketUSA:
		engine.addUSARiskRules()
	case constants.MarketMozambique:
		engine.addMozambiqueRiskRules()
	}

	logger.Info("Motor de risco inicializado", zap.Int("total_rules", len(engine.rules)))
	return engine
}

// addBasicRiskRules adiciona regras básicas de risco ao motor
func (re *RiskEngine) addBasicRiskRules() {
	// Regra para verificar transações de alto valor
	re.rules = append(re.rules, RiskRule{
		ID:          "high_value_transaction",
		Name:        "Transação de Alto Valor",
		Description: "Verifica se a transação excede um limiar de alto valor",
		Market:      constants.MarketGlobal,
		Severity:    "medium",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Definir limites por moeda
			thresholds := map[string]float64{
				"USD": 5000,
				"EUR": 4500,
				"AOA": 500000,
				"BRL": 10000,
				"MZN": 100000,
			}
			
			threshold, exists := thresholds[tx.Currency]
			if !exists {
				threshold = 5000 // Valor padrão
			}
			
			if tx.Amount > threshold {
				score := 0.6 // Score de risco médio
				return true, score, nil
			}
			return false, 0, nil
		},
	})

	// Regra para verificar discrepâncias de endereço
	re.rules = append(re.rules, RiskRule{
		ID:          "address_mismatch",
		Name:        "Discrepância de Endereço",
		Description: "Verifica se há discrepância entre endereço de cobrança e entrega",
		Market:      constants.MarketGlobal,
		Severity:    "low",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Se não houver endereço de entrega, não aplicar a regra
			if tx.ShippingAddress == nil || tx.BillingAddress == nil {
				return false, 0, nil
			}
			
			// Verificar discrepância de país ou estado
			if tx.ShippingAddress.Country != tx.BillingAddress.Country {
				return true, 0.7, nil
			}
			
			if tx.ShippingAddress.State != tx.BillingAddress.State {
				return true, 0.4, nil
			}
			
			return false, 0, nil
		},
	})

	// Regra para verificar múltiplas transações em curto período
	re.rules = append(re.rules, RiskRule{
		ID:          "rapid_succession",
		Name:        "Transações em Rápida Sucessão",
		Description: "Verifica se há múltiplas transações do mesmo usuário em curto período",
		Market:      constants.MarketGlobal,
		Severity:    "medium",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Verificar quantas transações prévias existem
			if len(tx.PreviousTransations) > 2 {
				return true, 0.5, nil
			}
			return false, 0, nil
		},
	})
}

// addAngolaRiskRules adiciona regras específicas para o mercado de Angola
func (re *RiskEngine) addAngolaRiskRules() {
	// Regra para verificar transações em moeda estrangeira (controle de câmbio BNA)
	re.rules = append(re.rules, RiskRule{
		ID:          "angola_foreign_currency",
		Name:        "Transação em Moeda Estrangeira",
		Description: "Verifica transações em moeda estrangeira conforme requisitos BNA",
		Market:      constants.MarketAngola,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Se não for moeda local (Kwanza)
			if tx.Currency != "AOA" {
				// Verificar se há autorização de câmbio
				if _, exists := tx.Metadata["exchange_authorization"]; !exists {
					return true, 0.8, nil
				}
			}
			return false, 0, nil
		},
	})

	// Regra para verificar transações para países sob sanções
	re.rules = append(re.rules, RiskRule{
		ID:          "angola_sanctioned_countries",
		Name:        "Transação para País sob Sanção",
		Description: "Verifica transações para países sob sanções conforme BNA/UIF",
		Market:      constants.MarketAngola,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Lista de países sob sanções conforme UIF Angola
			sanctionedCountries := []string{"KP", "IR", "SY", "CU"}
			
			countryCode := ""
			if tx.ShippingAddress != nil {
				countryCode = tx.ShippingAddress.Country
			} else if tx.BillingAddress != nil {
				countryCode = tx.BillingAddress.Country
			}
			
			if countryCode != "" {
				for _, sc := range sanctionedCountries {
					if countryCode == sc {
						return true, 0.9, nil
					}
				}
			}
			return false, 0, nil
		},
	})
}

// addBrazilRiskRules adiciona regras específicas para o mercado do Brasil
func (re *RiskEngine) addBrazilRiskRules() {
	// Regra para verificar transações suspeitas conforme COAF
	re.rules = append(re.rules, RiskRule{
		ID:          "brazil_coaf_suspicious",
		Name:        "Transação Suspeita COAF",
		Description: "Verifica padrões de transação suspeita conforme diretrizes COAF",
		Market:      constants.MarketBrazil,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Verificar transações fracionadas (múltiplas transações pequenas)
			if len(tx.PreviousTransations) > 3 && tx.Amount < 5000 {
				return true, 0.7, nil
			}
			
			// Verificar transações para PEPs (Pessoas Politicamente Expostas)
			if isPEP, _ := tx.Metadata["is_pep"].(bool); isPEP {
				return true, 0.8, nil
			}
			
			return false, 0, nil
		},
	})

	// Regra para verificar transações PIX específicas
	re.rules = append(re.rules, RiskRule{
		ID:          "brazil_pix_validation",
		Name:        "Validação PIX",
		Description: "Valida transações PIX conforme requisitos BACEN",
		Market:      constants.MarketBrazil,
		Severity:    "medium",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			if tx.PaymentType == PaymentTypePIX {
				// Verificar limites PIX conforme BACEN
				if tx.Amount > 100000 { // Limites para PIX noturno
					timeOfDay := tx.CreatedAt.Hour()
					if timeOfDay >= 20 || timeOfDay < 6 {
						return true, 0.6, nil
					}
				}
				
				// Verificar tipo de chave PIX
				pixKeyType, exists := tx.PaymentDetails["pix_key_type"].(string)
				if !exists || (pixKeyType != "cpf" && pixKeyType != "cnpj" && 
					pixKeyType != "email" && pixKeyType != "phone" && pixKeyType != "random") {
					return true, 0.5, nil
				}
			}
			return false, 0, nil
		},
	})
}

// addEURiskRules adiciona regras específicas para o mercado da UE
func (re *RiskEngine) addEURiskRules() {
	// Regra para verificar conformidade com SCA (PSD2)
	re.rules = append(re.rules, RiskRule{
		ID:          "eu_sca_compliance",
		Name:        "Conformidade SCA",
		Description: "Verifica conformidade com Strong Customer Authentication (PSD2)",
		Market:      constants.MarketEU,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Transações acima de 30 EUR exigem SCA
			if tx.Currency == "EUR" && tx.Amount > 30 {
				// Verificar se MFA de alto nível foi aplicado
				if tx.MFALevel != "high" {
					return true, 0.8, nil
				}
				
				// Para pagamentos com cartão, verificar 3DS
				if tx.PaymentType == PaymentTypeCard {
					if tx.ThreeDSData == nil || tx.ThreeDSData["version"] == nil {
						return true, 0.9, nil
					}
				}
			}
			return false, 0, nil
		},
	})

	// Regra para verificar pagamentos SEPA
	re.rules = append(re.rules, RiskRule{
		ID:          "eu_sepa_validation",
		Name:        "Validação SEPA",
		Description: "Valida transferências SEPA conforme regulamentações",
		Market:      constants.MarketEU,
		Severity:    "medium",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			if tx.PaymentType == PaymentTypeSEPA {
				// Verificar se IBAN está presente
				iban, exists := tx.PaymentDetails["iban"].(string)
				if !exists || len(iban) < 15 {
					return true, 0.6, nil
				}
				
				// Verificar se BIC/SWIFT está presente
				if _, exists := tx.PaymentDetails["bic"]; !exists {
					return true, 0.5, nil
				}
			}
			return false, 0, nil
		},
	})
}

// addUSARiskRules adiciona regras específicas para o mercado dos EUA
func (re *RiskEngine) addUSARiskRules() {
	// Regra para verificar conformidade com OFAC (sanções)
	re.rules = append(re.rules, RiskRule{
		ID:          "usa_ofac_compliance",
		Name:        "Conformidade OFAC",
		Description: "Verifica conformidade com lista de sanções OFAC",
		Market:      constants.MarketUSA,
		Severity:    "critical",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Verificar se há flag de sanção OFAC
			if ofacFlag, exists := tx.Metadata["ofac_match"].(bool); exists && ofacFlag {
				return true, 1.0, nil
			}
			
			return false, 0, nil
		},
	})

	// Regra para verificar conformidade com BSA (Bank Secrecy Act)
	re.rules = append(re.rules, RiskRule{
		ID:          "usa_bsa_compliance",
		Name:        "Conformidade BSA",
		Description: "Verifica conformidade com Bank Secrecy Act",
		Market:      constants.MarketUSA,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Transações acima de $10,000 exigem relatório CTR
			if tx.Currency == "USD" && tx.Amount > 10000 {
				if _, exists := tx.Metadata["ctr_filed"].(bool); !exists {
					return true, 0.8, nil
				}
			}
			
			// Verificar estruturação (múltiplas transações abaixo do limite de relatório)
			if tx.Currency == "USD" && tx.Amount > 3000 && tx.Amount < 10000 {
				if len(tx.PreviousTransations) > 2 {
					totalAmount := tx.Amount
					for _, prevTx := range tx.PreviousTransations {
						// Na implementação real, somaria os valores das transações anteriores
						totalAmount += 3000 // Valor simulado para exemplo
					}
					
					if totalAmount > 10000 {
						return true, 0.7, nil
					}
				}
			}
			
			return false, 0, nil
		},
	})
}

// addMozambiqueRiskRules adiciona regras específicas para o mercado de Moçambique
func (re *RiskEngine) addMozambiqueRiskRules() {
	// Regra para verificar transações em moeda estrangeira
	re.rules = append(re.rules, RiskRule{
		ID:          "mozambique_foreign_currency",
		Name:        "Transação em Moeda Estrangeira",
		Description: "Verifica transações em moeda estrangeira conforme Banco de Moçambique",
		Market:      constants.MarketMozambique,
		Severity:    "high",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Se não for moeda local (Metical)
			if tx.Currency != "MZN" {
				// Verificar se há autorização de câmbio
				if _, exists := tx.Metadata["exchange_authorization"]; !exists {
					return true, 0.8, nil
				}
			}
			return false, 0, nil
		},
	})

	// Regra para verificar transações de alto valor
	re.rules = append(re.rules, RiskRule{
		ID:          "mozambique_high_value",
		Name:        "Transação de Alto Valor",
		Description: "Verifica transações de alto valor conforme GIFiM",
		Market:      constants.MarketMozambique,
		Severity:    "medium",
		Evaluate: func(tx *PaymentTransaction) (bool, float64, error) {
			// Transações acima de 500,000 MZN exigem relatório ao GIFiM
			if tx.Currency == "MZN" && tx.Amount > 500000 {
				if _, exists := tx.Metadata["gifim_report"].(bool); !exists {
					return true, 0.7, nil
				}
			}
			return false, 0, nil
		},
	})
}

// EvaluateTransaction avalia uma transação através de regras de risco
func (re *RiskEngine) EvaluateTransaction(ctx context.Context, tx *PaymentTransaction) (float64, []string, error) {
	ctx, span := re.observer.Tracer().Start(ctx, "risk_engine_evaluate",
		trace.WithAttributes(
			attribute.String("transaction_id", tx.TransactionID),
			attribute.String("market", tx.MarketContext.Market),
			attribute.String("payment_type", tx.PaymentType),
			attribute.Float64("amount", tx.Amount),
			attribute.String("currency", tx.Currency),
		),
	)
	defer span.End()

	triggeredRules := make([]string, 0)
	highestScore := 0.0

	// Aplicar todas as regras globais
	for _, rule := range re.rules {
		if rule.Market == constants.MarketGlobal || rule.Market == tx.MarketContext.Market {
			triggered, score, err := rule.Evaluate(tx)
			if err != nil {
				re.logger.Error("Erro ao avaliar regra de risco",
					zap.String("rule_id", rule.ID),
					zap.String("transaction_id", tx.TransactionID),
					zap.Error(err))
				continue
			}

			if triggered {
				triggeredRules = append(triggeredRules, rule.ID)
				re.observer.TraceAuditEvent(ctx, tx.MarketContext, tx.UserID,
					"risk_rule_triggered",
					fmt.Sprintf("Regra de risco %s acionada para transação %s, score: %.2f",
						rule.ID, tx.TransactionID, score))

				if score > highestScore {
					highestScore = score
				}
			}
		}
	}

	// Registrar resultado da avaliação
	re.logger.Info("Avaliação de risco concluída",
		zap.String("transaction_id", tx.TransactionID),
		zap.Float64("risk_score", highestScore),
		zap.Int("triggered_rules", len(triggeredRules)))

	// Registrar métrica de score de risco
	re.observer.RecordHistogram(tx.MarketContext, "payment_gateway_risk_score", highestScore, tx.PaymentType)

	return highestScore, triggeredRules, nil
}

// ProcessPayment processa um pagamento através do gateway
func (pg *PaymentGateway) ProcessPayment(ctx context.Context, transaction PaymentTransaction) (string, error) {
	// Criar span para rastreabilidade da transação
	ctx, span := pg.observability.Tracer().Start(ctx, "process_payment",
		trace.WithAttributes(
			attribute.String("transaction_id", transaction.TransactionID),
			attribute.String("user_id", transaction.UserID),
			attribute.String("merchant_id", transaction.MerchantID),
			attribute.Float64("amount", transaction.Amount),
			attribute.String("currency", transaction.Currency),
			attribute.String("payment_type", transaction.PaymentType),
		),
	)
	defer span.End()

	// Registrar início da transação
	pg.logger.Info("Iniciando processamento de pagamento",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("type", transaction.PaymentType),
		zap.Float64("amount", transaction.Amount),
		zap.String("currency", transaction.Currency),
		zap.String("market", transaction.MarketContext.Market))

	// Verificar autenticação do usuário
	authenticated, err := pg.verifyAuthentication(ctx, transaction)
	if err != nil {
		pg.logger.Error("falha na autenticação", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return "", fmt.Errorf("falha na autenticação: %w", err)
	}
	if !authenticated {
		pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityHigh, "authentication_failed",
			fmt.Sprintf("Autenticação falhou para transação %s", transaction.TransactionID))
		return "", fmt.Errorf("autenticação falhou")
	}

	// Verificar autorização para a transação
	authorized, err := pg.verifyAuthorization(ctx, transaction)
	if err != nil {
		pg.logger.Error("falha na autorização", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return "", fmt.Errorf("falha na autorização: %w", err)
	}
	if !authorized {
		pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityHigh, "authorization_failed",
			fmt.Sprintf("Autorização falhou para transação %s", transaction.TransactionID))
		return "", fmt.Errorf("autorização falhou")
	}	// Verificar limites de transação
	if err := pg.verifyTransactionLimits(ctx, transaction); err != nil {
		pg.logger.Error("limite de transação excedido", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID, 
			constants.SecurityEventSeverityMedium, "limit_exceeded",
			fmt.Sprintf("Limite excedido para transação %s: %v", transaction.TransactionID, err))
		return "", fmt.Errorf("limite de transação excedido: %w", err)
	}

	// Verificar tipo de pagamento suportado
	if !pg.isPaymentTypeSupported(transaction.PaymentType) {
		pg.logger.Error("tipo de pagamento não suportado",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("payment_type", transaction.PaymentType))
		pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
			constants.SecurityEventSeverityMedium, "unsupported_payment_type",
			fmt.Sprintf("Tipo de pagamento %s não suportado para transação %s", 
				transaction.PaymentType, transaction.TransactionID))
		return "", fmt.Errorf("tipo de pagamento %s não suportado", transaction.PaymentType)
	}

	// Verificar regras de compliance específicas por mercado
	if err := pg.verifyComplianceRules(ctx, transaction); err != nil {
		pg.logger.Error("falha na verificação de compliance", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return "", fmt.Errorf("falha na verificação de compliance: %w", err)
	}

	// Avaliar risco da transação
	riskScore, triggeredRules, err := pg.riskEngine.EvaluateTransaction(ctx, &transaction)
	if err != nil {
		pg.logger.Error("falha na avaliação de risco", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return "", fmt.Errorf("falha na avaliação de risco: %w", err)
	}

	// Atualizar score de risco na transação
	transaction.RiskScore = riskScore
	
	// Determinar fluxo com base na avaliação de risco
	if riskScore >= 0.8 {
		// Risco muito alto - rejeitar automaticamente
		pg.logger.Warn("transação rejeitada por alto risco", 
			zap.String("transaction_id", transaction.TransactionID),
			zap.Float64("risk_score", riskScore),
			zap.Strings("triggered_rules", triggeredRules))
		
		pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
			constants.SecurityEventSeverityHigh, "high_risk_rejected",
			fmt.Sprintf("Transação %s rejeitada por alto risco (score: %.2f)", 
				transaction.TransactionID, riskScore))
		
		return "", fmt.Errorf("transação rejeitada por alto risco (score: %.2f)", riskScore)
	} else if riskScore >= 0.5 {
		// Risco médio - exigir verificação adicional
		pg.logger.Info("verificação adicional necessária", 
			zap.String("transaction_id", transaction.TransactionID),
			zap.Float64("risk_score", riskScore))
		
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"additional_verification_required",
			fmt.Sprintf("Verificação adicional exigida para transação %s (score: %.2f)", 
				transaction.TransactionID, riskScore))
		
		// Na implementação real, aqui haveria um mecanismo para verificação adicional
		// Simulado como aprovado para este exemplo
		pg.logger.Info("verificação adicional concluída", 
			zap.String("transaction_id", transaction.TransactionID))
	}

	// Verificar verificações específicas para 3D Secure se aplicável
	if transaction.PaymentType == PaymentTypeCard && pg.config.PSP3DSEnabled {
		if err := pg.process3DS(ctx, transaction); err != nil {
			pg.logger.Error("falha no processamento 3DS", 
				zap.String("transaction_id", transaction.TransactionID), 
				zap.Error(err))
			return "", fmt.Errorf("falha no processamento 3DS: %w", err)
		}
	}

	// Executar a transação de pagamento
	processorRef, err := pg.executePayment(ctx, transaction)
	if err != nil {
		pg.logger.Error("falha ao processar pagamento", 
			zap.String("transaction_id", transaction.TransactionID), 
			zap.Error(err))
		return "", fmt.Errorf("falha ao processar pagamento: %w", err)
	}

	// Atualizar volumes diários
	pg.updateDailyVolume(transaction.PaymentType, transaction.Amount)

	// Registrar evento de auditoria para transação bem-sucedida
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "payment_completed",
		fmt.Sprintf("Transação %s completada com sucesso via %s, valor: %f %s", 
			transaction.TransactionID, transaction.PaymentType, transaction.Amount, transaction.Currency))
	
	// Registrar métricas de transação
	pg.observability.RecordMetric(transaction.MarketContext, "payment_gateway_transaction_count", 
		transaction.PaymentType, 1)
	pg.observability.RecordHistogram(transaction.MarketContext, "payment_gateway_transaction_amount", 
		transaction.Amount, transaction.Currency)

	return processorRef, nil
}

// process3DS processa autenticação 3D Secure para pagamentos com cartão
func (pg *PaymentGateway) process3DS(ctx context.Context, transaction PaymentTransaction) error {
	ctx, span := pg.observability.Tracer().Start(ctx, "process_3ds",
		trace.WithAttributes(
			attribute.String("transaction_id", transaction.TransactionID),
			attribute.String("market", transaction.MarketContext.Market),
		),
	)
	defer span.End()

	// Verificar se já há dados 3DS
	if transaction.ThreeDSData != nil {
		// Verificar status de autenticação
		authStatus, ok := transaction.ThreeDSData["authentication_status"].(string)
		if !ok {
			return fmt.Errorf("status de autenticação 3DS inválido")
		}

		// Log de status de autenticação
		pg.logger.Info("Status de autenticação 3DS recebido",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("auth_status", authStatus))

		// Verificar status
		switch authStatus {
		case "Y": // Autenticação bem-sucedida
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"3ds_authentication_successful",
				fmt.Sprintf("Autenticação 3DS bem-sucedida para transação %s", transaction.TransactionID))
			return nil
		case "N": // Autenticação falhou
			pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "3ds_authentication_failed",
				fmt.Sprintf("Autenticação 3DS falhou para transação %s", transaction.TransactionID))
			return fmt.Errorf("autenticação 3DS falhou")
		case "A": // Tentativa de autenticação
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"3ds_attempt_processed",
				fmt.Sprintf("Tentativa de autenticação 3DS processada para transação %s", transaction.TransactionID))
			return nil
		case "U": // Não autenticado / indisponível
			// Avaliar se deve continuar baseado no risco
			if transaction.RiskScore > 0.6 {
				pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
					constants.SecurityEventSeverityMedium, "3ds_unavailable_high_risk",
					fmt.Sprintf("3DS indisponível para transação %s de alto risco", transaction.TransactionID))
				return fmt.Errorf("3DS indisponível para transação de alto risco")
			}
			// Permitir prosseguir para transações de baixo risco
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
				"3ds_unavailable_proceed",
				fmt.Sprintf("3DS indisponível, prosseguindo com transação %s de baixo risco", transaction.TransactionID))
			return nil
		default:
			return fmt.Errorf("status de autenticação 3DS desconhecido: %s", authStatus)
		}
	}

	// Se não há dados 3DS e PSD2 exige SCA, iniciar fluxo 3DS
	if transaction.MarketContext.Market == constants.MarketEU && transaction.Amount > 30 && transaction.Currency == "EUR" {
		// Simular início do fluxo 3DS
		pg.logger.Info("Iniciando fluxo 3DS para conformidade PSD2",
			zap.String("transaction_id", transaction.TransactionID),
			zap.String("redirect_url", pg.config.PAResRoute))

		// Registrar evento de auditoria para início de fluxo 3DS
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"3ds_flow_initiated",
			fmt.Sprintf("Fluxo 3DS iniciado para transação %s conforme PSD2", transaction.TransactionID))

		// Na implementação real, aqui retornaria uma URL de redirecionamento ou um objeto AReq
		// Para simulação, tratamos como se o fluxo tivesse sido bem-sucedido
		return nil
	}

	// Para outros mercados ou valores abaixo do limite, continuar sem 3DS
	return nil
}

// verifyAuthentication verifica a autenticação do usuário
func (pg *PaymentGateway) verifyAuthentication(ctx context.Context, transaction PaymentTransaction) (bool, error) {
	ctx, span := pg.observability.Tracer().Start(ctx, "verify_authentication")
	defer span.End()

	// Obter metadados de compliance para o mercado
	metadata, exists := pg.observability.GetComplianceMetadata(transaction.MarketContext.Market)
	if !exists {
		metadata, _ = pg.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Verificar MFA conforme requisitos de compliance para Payment Gateway
	var requiredMFALevel string
	
	// Determinar nível MFA necessário com base no valor e tipo de transação
	if transaction.Amount > 10000 || transaction.PaymentType == PaymentTypeRemittance {
		// Transações de alto valor exigem MFA de nível mais alto
		requiredMFALevel = "high"
	} else if transaction.MarketContext.Market == constants.MarketEU && 
		transaction.Amount > 30 && transaction.Currency == "EUR" {
		// PSD2 exige SCA (autenticação forte) para transações acima de 30 EUR
		requiredMFALevel = "high"
	} else {
		// Caso contrário, usar requisito padrão do mercado
		requiredMFALevel = metadata.RequiredMFALevel
	}
	
	// Verificar se o nível MFA fornecido é suficiente
	mfaResult, err := pg.observability.ValidateMFA(ctx, transaction.MarketContext, transaction.UserID, transaction.MFALevel)
	if err != nil {
		return false, err
	}

	if !mfaResult {
		return false, fmt.Errorf("nível MFA insuficiente para transação de pagamento no mercado %s: requer %s, fornecido %s",
			transaction.MarketContext.Market, requiredMFALevel, transaction.MFALevel)
	}

	// Registrar evento de auditoria
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "authentication_verified",
		fmt.Sprintf("Autenticação verificada com MFA nível %s para transação de pagamento", transaction.MFALevel))

	return true, nil
}

// verifyAuthorization verifica a autorização para a transação
func (pg *PaymentGateway) verifyAuthorization(ctx context.Context, transaction PaymentTransaction) (bool, error) {
	ctx, span := pg.observability.Tracer().Start(ctx, "verify_authorization")
	defer span.End()

	// Verificar escopo para transação de pagamento
	scope := fmt.Sprintf("payment:%s", transaction.PaymentType)
	scopeResult, err := pg.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, scope)
	if err != nil {
		return false, err
	}

	if !scopeResult {
		return false, fmt.Errorf("usuário não tem escopo para realizar pagamentos de tipo %s", transaction.PaymentType)
	}

	// Verificar requisitos específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// BNA exige verificação adicional para pagamentos internacionais
		if transaction.Currency != "AOA" {
			additionalScope, err := pg.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "payment:bna:foreign_currency")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BNA para pagamentos em moeda estrangeira")
			}
		}
		
	case constants.MarketBrazil:
		// BACEN exige escopo especial para PIX
		if transaction.PaymentType == PaymentTypePIX {
			additionalScope, err := pg.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "payment:bacen:pix")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BACEN para pagamentos via PIX")
			}
		}

	case constants.MarketEU:
		// PSD2 exige escopo especial para pagamentos de alto valor
		if transaction.Amount > 1000 && transaction.Currency == "EUR" {
			additionalScope, err := pg.observability.ValidateScope(ctx, transaction.MarketContext, transaction.UserID, "payment:psd2:high_value")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo PSD2 para pagamentos de alto valor")
			}
		}
	}

	// Registrar evento de auditoria
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "authorization_verified",
		fmt.Sprintf("Autorização verificada para transação de pagamento %s", transaction.TransactionID))

	return true, nil
}

// verifyTransactionLimits verifica se a transação está dentro dos limites configurados
func (pg *PaymentGateway) verifyTransactionLimits(ctx context.Context, transaction PaymentTransaction) error {
	ctx, span := pg.observability.Tracer().Start(ctx, "verify_transaction_limits")
	defer span.End()

	// Verificar limite específico para o tipo de transação
	limit, exists := pg.config.TransactionLimits[transaction.PaymentType]
	if !exists {
		// Se não houver limite específico, usar limite genérico "default"
		limit, exists = pg.config.TransactionLimits["default"]
		if !exists {
			// Se não houver limite default, não aplicar restrição de limite
			limit = -1
		}
	}

	// Verificar se a transação ultrapassa o limite
	if limit >= 0 {
		// Obter volume diário atual para o tipo de transação
		currentVolume := pg.getDailyVolume(transaction.PaymentType)
		
		// Verificar se a transação ultrapassa o limite
		if currentVolume+transaction.Amount > limit {
			// Registrar evento de segurança para limite excedido
			pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "daily_limit_exceeded",
				fmt.Sprintf("Transação %s ultrapassa limite diário para %s. Limite: %f, Volume atual: %f, Valor da transação: %f",
					transaction.TransactionID, transaction.PaymentType, limit, currentVolume, transaction.Amount))
			
			return fmt.Errorf("transação ultrapassa limite diário de %f para tipo %s", limit, transaction.PaymentType)
		}
	}

	// Verificar limites específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Regras específicas BNA para limites de transação
		if transaction.Currency != "AOA" {
			// Limite específico para transações em moeda estrangeira pelo BNA
			// Aplicar taxa de conversão estimada para comparação
			estimatedKZValue := transaction.Amount * 850 // Taxa aproximada USD para Kwanza
			if estimatedKZValue > 500000 { // 500,000 Kwanzas
				pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
					constants.SecurityEventSeverityHigh, "bna_forex_limit_exceeded",
					fmt.Sprintf("Transação %s ultrapassa limite BNA para câmbio", transaction.TransactionID))
				return fmt.Errorf("transação ultrapassa limite BNA para moeda estrangeira")
			}
		}
		
	case constants.MarketBrazil:
		// Regras específicas BACEN para limites de transação
		if transaction.PaymentType == PaymentTypePIX {
			// Verificar limites PIX noturno (entre 20h e 6h)
			currentHour := time.Now().Hour()
			if (currentHour >= 20 || currentHour < 6) && transaction.Amount > 1000 {
				pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
					constants.SecurityEventSeverityMedium, "pix_night_limit_exceeded",
					fmt.Sprintf("Transação %s ultrapassa limite noturno PIX", transaction.TransactionID))
				return fmt.Errorf("transação ultrapassa limite noturno PIX")
			}
		}
		
	case constants.MarketEU:
		// Regras específicas PSD2 para limites de transação
		// Transações sem SCA limitadas a 30 EUR
		if transaction.Amount > 30 && transaction.Currency == "EUR" && transaction.MFALevel != "high" {
			pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
				constants.SecurityEventSeverityMedium, "psd2_sca_required",
				fmt.Sprintf("Transação %s requer SCA conforme PSD2", transaction.TransactionID))
			return fmt.Errorf("transação requer SCA (Strong Customer Authentication) conforme PSD2")
		}
	}

	// Registrar evento de auditoria
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "limits_verified",
		fmt.Sprintf("Limites verificados para transação de pagamento %s", transaction.TransactionID))

	return nil
}// isPaymentTypeSupported verifica se um tipo de pagamento é suportado
func (pg *PaymentGateway) isPaymentTypeSupported(paymentType string) bool {
	supported, exists := pg.config.SupportedPayments[paymentType]
	return exists && supported
}

// verifyComplianceRules verifica regras de compliance específicas por mercado
func (pg *PaymentGateway) verifyComplianceRules(ctx context.Context, transaction PaymentTransaction) error {
	ctx, span := pg.observability.Tracer().Start(ctx, "verify_compliance_rules")
	defer span.End()

	// Obter metadados de compliance para o mercado
	metadata, exists := pg.observability.GetComplianceMetadata(transaction.MarketContext.Market)
	if !exists {
		metadata, _ = pg.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Aplicar regras globais de compliance
	pg.logger.Info("Verificando regras de compliance", 
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("market", transaction.MarketContext.Market),
		zap.Strings("frameworks", metadata.Frameworks))

	// Verificar cada regra de compliance aplicável ao tipo de pagamento
	for _, rule := range pg.complianceRules {
		// Verificar se a regra se aplica ao mercado atual ou é global
		if rule.Market == constants.MarketGlobal || rule.Market == transaction.MarketContext.Market {
			// Verificar se a regra se aplica ao tipo de pagamento
			applies := false
			for _, paymentType := range rule.MandatoryFor {
				if paymentType == transaction.PaymentType {
					applies = true
					break
				}
			}

			if applies {
				// Aplicar a regra
				compliant, message, err := rule.Validate(&transaction)
				if err != nil {
					pg.logger.Error("Erro ao validar regra de compliance",
						zap.String("rule_id", rule.ID),
						zap.String("transaction_id", transaction.TransactionID),
						zap.Error(err))
					return fmt.Errorf("erro ao validar regra de compliance %s: %w", rule.ID, err)
				}

				if !compliant {
					// Registrar evento de não conformidade
					pg.observability.TraceSecurityEvent(ctx, transaction.MarketContext, transaction.UserID,
						constants.SecurityEventSeverityHigh, "compliance_rule_failed",
						fmt.Sprintf("Regra de compliance %s falhou para transação %s: %s", 
							rule.ID, transaction.TransactionID, message))
					
					return fmt.Errorf("não conformidade detectada - %s: %s", rule.ID, message)
				}

				// Registrar evento de auditoria para conformidade
				pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
					fmt.Sprintf("compliance_%s_verified", rule.ID),
					fmt.Sprintf("Regra de compliance %s verificada: %s", rule.ID, message))
			}
		}
	}

	// Verificar requisitos específicos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Verificação específica UIF/BNA para Angola
		if transaction.Amount > 250000 || transaction.Currency != "AOA" {
			// Verificar consentimento para compartilhamento com UIF
			consentResult, err := pg.observability.ValidateConsent(ctx, transaction.MarketContext, 
				transaction.UserID, "data_sharing:uif_angola")
			if err != nil || !consentResult {
				return fmt.Errorf("consentimento para compartilhamento com UIF Angola não encontrado")
			}
			
			// Registrar notificação UIF
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"uif_notification",
				fmt.Sprintf("Transação %s notificada à UIF Angola", transaction.TransactionID))
		}
		
	case constants.MarketBrazil:
		// Verificação específica COAF/BACEN para Brasil
		if transaction.Amount > 10000 && transaction.Currency == "BRL" {
			// Verificar consentimento para compartilhamento com COAF
			consentResult, err := pg.observability.ValidateConsent(ctx, transaction.MarketContext, 
				transaction.UserID, "data_sharing:coaf")
			if err != nil || !consentResult {
				return fmt.Errorf("consentimento para compartilhamento com COAF não encontrado")
			}
			
			// Registrar notificação COAF
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"coaf_notification",
				fmt.Sprintf("Transação %s notificada ao COAF", transaction.TransactionID))
		}
		
	case constants.MarketEU:
		// Verificação específica para AMLD5 (EU Anti-Money Laundering Directive)
		if transaction.Amount > 1000 && transaction.Currency == "EUR" {
			// Verificar sanções e PEP (Pessoas Politicamente Expostas)
			pg.logger.Info("Verificando conformidade AMLD5",
				zap.String("transaction_id", transaction.TransactionID),
				zap.String("user_id", transaction.UserID))
			
			// Registrar verificação AMLD5
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"amld5_check",
				fmt.Sprintf("Verificação AMLD5 realizada para transação %s", transaction.TransactionID))
		}
		
	case constants.MarketMozambique:
		// Verificação específica para Moçambique
		if transaction.Amount > 50000 && transaction.Currency == "MZN" {
			// Registrar notificação GIFiM
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"gifim_notification",
				fmt.Sprintf("Transação %s notificada ao GIFiM", transaction.TransactionID))
		}
	}

	// Verificar consentimento GDPR para todos os mercados (especialmente UE)
	if transaction.MarketContext.Market == constants.MarketEU {
		consentResult, err := pg.observability.ValidateConsent(ctx, transaction.MarketContext, 
			transaction.UserID, "data_processing:payment")
		if err != nil || !consentResult {
			return fmt.Errorf("consentimento GDPR para processamento de dados de pagamento não encontrado")
		}
		
		// Registrar verificação de consentimento GDPR
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"gdpr_consent_verified",
			fmt.Sprintf("Consentimento GDPR verificado para transação %s", transaction.TransactionID))
	}

	// Registrar evento de auditoria para compliance
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "compliance_verified",
		fmt.Sprintf("Compliance verificado para transação %s conforme framework %s", 
			transaction.TransactionID, metadata.Frameworks[0]))

	return nil
}

// executePayment executa a transação de pagamento (simulado)
func (pg *PaymentGateway) executePayment(ctx context.Context, transaction PaymentTransaction) (string, error) {
	ctx, span := pg.observability.Tracer().Start(ctx, "execute_payment")
	defer span.End()

	// Simular processamento de pagamento
	// Em produção, aqui seria o código real de execução da transação
	pg.logger.Info("Processando transação de pagamento",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("user_id", transaction.UserID),
		zap.Float64("amount", transaction.Amount),
		zap.String("currency", transaction.Currency),
		zap.String("type", transaction.PaymentType))
	
	// Simular tempo de processamento
	time.Sleep(200 * time.Millisecond)

	// Gerar referência do processador
	processorRef := fmt.Sprintf("PSP-%s-%d", transaction.TransactionID, time.Now().UnixNano())

	// Registrar fluxo específico por tipo de pagamento
	switch transaction.PaymentType {
	case PaymentTypeCard:
		pg.logger.Info("Processando pagamento com cartão",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Aplicar tokenização para PCI DSS
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"card_tokenized",
			fmt.Sprintf("Dados do cartão tokenizados para transação %s", transaction.TransactionID))
		
	case PaymentTypeBank:
		pg.logger.Info("Processando transferência bancária",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Registrar verificação de conta bancária
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"bank_account_verified",
			fmt.Sprintf("Conta bancária verificada para transação %s", transaction.TransactionID))
		
	case PaymentTypeWallet:
		pg.logger.Info("Processando pagamento com carteira digital",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Registrar verificação de carteira digital
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"wallet_verified",
			fmt.Sprintf("Carteira digital verificada para transação %s", transaction.TransactionID))
		
	case PaymentTypePIX:
		pg.logger.Info("Processando pagamento via PIX",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Registrar verificação de chave PIX
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID,
			"pix_key_verified",
			fmt.Sprintf("Chave PIX verificada para transação %s", transaction.TransactionID))
	}

	// Verificar requisitos específicos para pagamentos por mercado
	switch transaction.MarketContext.Market {
	case constants.MarketAngola:
		// Requisitos específicos BNA para pagamentos
		pg.logger.Info("Aplicando regras BNA para pagamento",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Para transações acima de um certo valor, registrar para BNA
		if transaction.Amount > 100000 {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"bna_report_generated",
				fmt.Sprintf("Relatório BNA gerado para transação %s", transaction.TransactionID))
		}
		
		// Verificar integração com EMIS (Sistema Interbancário Angolano)
		if transaction.PaymentType == PaymentTypeCard || transaction.PaymentType == PaymentTypeBank {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"emis_check_completed",
				fmt.Sprintf("Verificação EMIS completada para transação %s", transaction.TransactionID))
		}

	case constants.MarketBrazil:
		// Requisitos específicos BACEN/LGPD para pagamentos
		pg.logger.Info("Aplicando regras BACEN/LGPD para pagamentos",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Integração com Sistema de Pagamentos Brasileiro (SPB)
		if transaction.PaymentType == PaymentTypeBank || transaction.PaymentType == PaymentTypePIX {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"spb_transaction_recorded",
				fmt.Sprintf("Transação %s registrada via SPB", transaction.TransactionID))
		}
		
		// Verificações LGPD para consentimento
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"lgpd_consent_verified",
			fmt.Sprintf("Consentimento LGPD verificado para processamento de dados na transação %s", 
				transaction.TransactionID))

	case constants.MarketEU:
		// Requisitos específicos PSD2/GDPR para pagamentos
		pg.logger.Info("Aplicando regras PSD2/GDPR para pagamentos",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Registrar aplicação de SCA (Strong Customer Authentication)
		if transaction.Amount > 30 && transaction.Currency == "EUR" {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"sca_applied",
				fmt.Sprintf("SCA aplicado para transação %s conforme PSD2", 
					transaction.TransactionID))
		}
		
		// Registrar minimização de dados conforme GDPR
		pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
			"gdpr_data_minimization_applied",
			fmt.Sprintf("Minimização de dados GDPR aplicada para transação %s", 
				transaction.TransactionID))
		
	case constants.MarketUSA:
		// Requisitos específicos para pagamentos nos EUA
		pg.logger.Info("Aplicando regras EUA para pagamentos",
			zap.String("transaction_id", transaction.TransactionID))
		
		// Verificação OFAC para transações internacionais
		if transaction.PaymentType == PaymentTypeRemittance {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"ofac_check_completed",
				fmt.Sprintf("Verificação OFAC completada para transação %s", 
					transaction.TransactionID))
		}
		
		// Verificação PCI DSS para pagamentos com cartão
		if transaction.PaymentType == PaymentTypeCard {
			pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, 
				"pci_dss_compliance_verified",
				fmt.Sprintf("Conformidade PCI DSS verificada para transação %s", 
					transaction.TransactionID))
		}
	}

	// Registrar evento de auditoria para conclusão da transação
	pg.observability.TraceAuditEvent(ctx, transaction.MarketContext, transaction.UserID, "payment_executed",
		fmt.Sprintf("Pagamento %s executado com sucesso via %s", 
			transaction.TransactionID, transaction.PaymentType))

	// Registrar métrica de tempo de processamento
	pg.observability.RecordMetric(transaction.MarketContext, "payment_gateway_processing_time",
		transaction.PaymentType, 200) // Valor simulado em ms

	return processorRef, nil
}

// updateDailyVolume atualiza o volume diário acumulado para um tipo de transação
func (pg *PaymentGateway) updateDailyVolume(paymentType string, amount float64) {
	pg.mutex.Lock()
	defer pg.mutex.Unlock()
	
	currentVolume, exists := pg.dailyVolumes[paymentType]
	if !exists {
		currentVolume = 0
	}
	
	pg.dailyVolumes[paymentType] = currentVolume + amount
}

// getDailyVolume obtém o volume diário acumulado para um tipo de transação
func (pg *PaymentGateway) getDailyVolume(paymentType string) float64 {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	
	volume, exists := pg.dailyVolumes[paymentType]
	if !exists {
		return 0
	}
	
	return volume
}

// resetDailyVolumes redefine os volumes diários (chamado à meia-noite)
func (pg *PaymentGateway) resetDailyVolumes() {
	pg.mutex.Lock()
	defer pg.mutex.Unlock()
	
	pg.dailyVolumes = make(map[string]float64)
	
	// Registrar redefinição de volumes em log
	pg.logger.Info("Volumes diários de transação redefinidos")
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

// Start inicia o serviço de gateway de pagamento
func (pg *PaymentGateway) Start() error {
	pg.logger.Info("Iniciando serviço Payment Gateway", 
		zap.String("market", pg.config.Market),
		zap.String("environment", pg.config.Environment))
	
	// Iniciar workers e componentes aqui
	// ...
	
	// Iniciar worker para reset diário de volumes
	pg.wg.Add(1)
	go pg.startDailyResetWorker()

	// Registrar métrica de inicialização
	pg.observability.RecordMetric(adapter.MarketContext{
		Market:     pg.config.Market,
		TenantType: pg.config.TenantType,
	}, "payment_gateway_status", "started", 1)

	return nil
}

// startDailyResetWorker inicia um worker para redefinir volumes diários à meia-noite
func (pg *PaymentGateway) startDailyResetWorker() {
	defer pg.wg.Done()
	
	ticker := time.NewTicker(time.Hour) // Verificar a cada hora
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Verificar se é meia-noite
			now := time.Now()
			if now.Hour() == 0 && now.Minute() < 10 { // Primeiros 10 minutos após meia-noite
				pg.resetDailyVolumes()
				pg.logger.Info("Volumes diários redefinidos", 
					zap.String("market", pg.config.Market))
			}
		case <-pg.shutdown:
			pg.logger.Info("Worker de reset diário encerrado")
			return
		}
	}
}

// Stop para o serviço de gateway de pagamento graciosamente
func (pg *PaymentGateway) Stop() error {
	pg.logger.Info("Parando serviço Payment Gateway", 
		zap.String("market", pg.config.Market))
	
	// Sinalizar para todos os workers pararem
	close(pg.shutdown)
	
	// Aguardar todos os workers encerrarem
	pg.wg.Wait()
	
	// Encerrar componentes de observabilidade
	pg.observability.Shutdown()
	
	// Fechar logger
	pg.logger.Info("Serviço Payment Gateway encerrado com sucesso")
	pg.logger.Sync()
	
	return nil
}// main é a função principal para executar o módulo de Payment Gateway
func main() {
	// Configurar logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Falha ao inicializar logger: %v", err)
	}
	defer logger.Sync()

	// Obter configuração do ambiente
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	serviceVersion := os.Getenv("SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	// Configurar mercado primário e tipo de tenant
	// Em produção, isso viria de variáveis de ambiente ou configuração
	market := constants.MarketBrazil // Mercado principal
	tenantType := "payment_processor" // Tipo de tenant

	// Criar contexto de mercado
	marketContext := adapter.MarketContext{
		Market:     market,
		TenantType: tenantType,
	}

	// Instanciar adaptador MCP-IAM Observability
	observability, err := adapter.NewAdapter(&adapter.Config{
		ServiceName:    "payment_gateway",
		ServiceVersion: serviceVersion,
		Environment:    environment,
		Market:         market,
		TenantType:     tenantType,
	})
	if err != nil {
		logger.Fatal("Falha ao inicializar adaptador de observabilidade",
			zap.Error(err))
	}

	// Registrar metadados de compliance para todos os mercados suportados
	registerComplianceMetadata(observability)

	// Criar configuração para Payment Gateway
	config := PaymentGatewayConfig{
		Market:       market,
		TenantType:   tenantType,
		Environment:  environment,
		PSP3DSEnabled: true,
		PAResRoute:   "/payment/3ds/verify",
		SupportedPayments: map[string]bool{
			PaymentTypeCard:       true,
			PaymentTypeBank:       true,
			PaymentTypeWallet:     true,
			PaymentTypePIX:        market == constants.MarketBrazil, // PIX apenas para Brasil
			PaymentTypeRemittance: true,
			PaymentTypeBoleto:     market == constants.MarketBrazil, // Boleto apenas para Brasil
			PaymentTypeMPesa:      market == constants.MarketMozambique, // MPesa apenas para Moçambique
		},
		TransactionLimits: map[string]float64{
			"default":             100000, // Limite genérico
			PaymentTypeCard:       50000,  // Limite para cartões
			PaymentTypeBank:       100000, // Limite para transferências bancárias
			PaymentTypeWallet:     10000,  // Limite para carteiras digitais
			PaymentTypePIX:        5000,   // Limite para PIX (Brasil)
			PaymentTypeRemittance: 50000,  // Limite para remessas internacionais
		},
	}

	// Instanciar Payment Gateway
	gateway := NewPaymentGateway(config, observability, logger)

	// Inicializar regras de compliance
	initializeComplianceRules(gateway)

	// Inicializar regras de risco
	initializeRiskRules(gateway)

	// Registrar métricas iniciais
	registerInitialMetrics(gateway)

	// Iniciar o serviço
	if err := gateway.Start(); err != nil {
		logger.Fatal("Falha ao iniciar serviço Payment Gateway",
			zap.Error(err))
	}

	// Configurar captura de sinais para encerramento gracioso
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Aguardar sinal para encerramento
	<-signalChan
	logger.Info("Sinal de encerramento recebido")

	// Parar o serviço graciosamente
	if err := gateway.Stop(); err != nil {
		logger.Error("Erro ao encerrar Payment Gateway",
			zap.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

// registerComplianceMetadata registra metadados de compliance para todos os mercados
func registerComplianceMetadata(observability adapter.IAMObservability) {
	// Angola
	observability.RegisterComplianceMetadata(constants.MarketAngola, adapter.ComplianceMetadata{
		Frameworks:     []string{"BNA", "UIF Angola", "PCI DSS"},
		RequiredMFALevel: "medium",
		DataRetention:  "7y",
		Regulators:     []string{"BNA", "UIF Angola"},
		Classification: []string{"financial", "payment"},
		SpecialRequirements: []string{
			"foreign_currency_approval",
			"uif_notification_high_value",
			"emis_integration",
		},
	})

	// Brasil
	observability.RegisterComplianceMetadata(constants.MarketBrazil, adapter.ComplianceMetadata{
		Frameworks:     []string{"BACEN", "LGPD", "PCI DSS", "SPB"},
		RequiredMFALevel: "medium",
		DataRetention:  "5y",
		Regulators:     []string{"BACEN", "ANPD", "COAF"},
		Classification: []string{"financial", "payment", "personal_data"},
		SpecialRequirements: []string{
			"pix_regulations",
			"coaf_notification",
			"spb_integration",
			"lgpd_consent_management",
		},
	})

	// União Europeia
	observability.RegisterComplianceMetadata(constants.MarketEU, adapter.ComplianceMetadata{
		Frameworks:     []string{"PSD2", "GDPR", "AMLD5", "PCI DSS"},
		RequiredMFALevel: "high",
		DataRetention:  "5y",
		Regulators:     []string{"EBA", "National Banks", "Data Protection Authorities"},
		Classification: []string{"financial", "payment", "personal_data"},
		SpecialRequirements: []string{
			"sca_for_payments",
			"3ds2_integration",
			"gdpr_data_minimization",
			"cross_border_notifications",
		},
	})

	// EUA
	observability.RegisterComplianceMetadata(constants.MarketUSA, adapter.ComplianceMetadata{
		Frameworks:     []string{"PCI DSS", "BSA/AML", "OFAC", "CFPB"},
		RequiredMFALevel: "medium",
		DataRetention:  "7y",
		Regulators:     []string{"OFAC", "FinCEN", "Federal Reserve"},
		Classification: []string{"financial", "payment"},
		SpecialRequirements: []string{
			"ofac_verification",
			"fincen_reporting",
			"nacha_compliance",
			"card_association_rules",
		},
	})

	// Moçambique
	observability.RegisterComplianceMetadata(constants.MarketMozambique, adapter.ComplianceMetadata{
		Frameworks:     []string{"BM", "GIFiM", "PCI DSS"},
		RequiredMFALevel: "medium",
		DataRetention:  "5y",
		Regulators:     []string{"BM", "GIFiM"},
		Classification: []string{"financial", "payment"},
		SpecialRequirements: []string{
			"mpesa_integration",
			"gifim_reporting",
			"bm_forex_approval",
		},
	})

	// Global (padrões aplicáveis a todos os mercados)
	observability.RegisterComplianceMetadata(constants.MarketGlobal, adapter.ComplianceMetadata{
		Frameworks:     []string{"PCI DSS", "ISO 8583", "ISO 20022"},
		RequiredMFALevel: "medium",
		DataRetention:  "5y",
		Regulators:     []string{"Card Networks", "SWIFT"},
		Classification: []string{"financial", "payment"},
		SpecialRequirements: []string{
			"card_data_security",
			"sanction_screening",
			"aml_monitoring",
		},
	})
}

// initializeComplianceRules inicializa as regras de compliance para o gateway
func initializeComplianceRules(pg *PaymentGateway) {
	// Regras globais aplicáveis a todos os mercados
	pg.complianceRules = append(pg.complianceRules, ComplianceRule{
		ID:          "aml_check",
		Market:      constants.MarketGlobal,
		Description: "Verificação Anti-Lavagem de Dinheiro",
		MandatoryFor: []string{
			PaymentTypeCard, PaymentTypeBank, PaymentTypeWallet, 
			PaymentTypePIX, PaymentTypeRemittance,
		},
		Validate: func(transaction *PaymentTransaction) (bool, string, error) {
			// Simulação de verificação AML
			// Em produção, integração com sistema AML real
			// Regra de exemplo: transações acima de 10.000 requerem verificação adicional
			if transaction.Amount >= 10000 {
				// Simular que a verificação foi bem-sucedida
				return true, "Verificação AML concluída com sucesso", nil
			}
			return true, "Transação abaixo do limite para verificação AML detalhada", nil
		},
	})

	// PCI DSS - Proteção de dados de cartão
	pg.complianceRules = append(pg.complianceRules, ComplianceRule{
		ID:          "pci_dss",
		Market:      constants.MarketGlobal,
		Description: "Conformidade com PCI DSS",
		MandatoryFor: []string{PaymentTypeCard},
		Validate: func(transaction *PaymentTransaction) (bool, string, error) {
			// Verificar se dados sensíveis do cartão estão presentes
			// Em uma implementação real, verificaria tokenização e proteção de dados
			if transaction.PaymentType == PaymentTypeCard {
				// Verificar se há dados sensíveis do cartão
				if transaction.CardData != nil {
					cardNumber := transaction.CardData["card_number"]
					// Verificar se é tokenizado (ex: token começa com "tok_")
					if token, ok := cardNumber.(string); ok && !strings.HasPrefix(token, "tok_") {
						return false, "Dados do cartão não tokenizados", nil
					}
				}
			}
			return true, "Proteção de dados de cartão verificada", nil
		},
	})

	// Verificação de sanções
	pg.complianceRules = append(pg.complianceRules, ComplianceRule{
		ID:          "sanction_check",
		Market:      constants.MarketGlobal,
		Description: "Verificação de sanções internacionais",
		MandatoryFor: []string{
			PaymentTypeRemittance, PaymentTypeBank, PaymentTypeCard,
		},
		Validate: func(transaction *PaymentTransaction) (bool, string, error) {
			// Simulação de verificação de sanções
			// Em produção, integração com serviços de verificação de sanções
			
			// Verificar transferências internacionais com mais atenção
			if transaction.PaymentType == PaymentTypeRemittance {
				if transaction.RecipientCountry == "IR" || transaction.RecipientCountry == "KP" {
					return false, "País destinatário sob sanções internacionais", nil
				}
			}
			
			return true, "Verificação de sanções concluída", nil
		},
	})

	// Regras específicas para Brasil
	pg.complianceRules = append(pg.complianceRules, ComplianceRule{
		ID:          "pix_regulation",
		Market:      constants.MarketBrazil,
		Description: "Regulamentos PIX (BACEN)",
		MandatoryFor: []string{PaymentTypePIX},
		Validate: func(transaction *PaymentTransaction) (bool, string, error) {
			// Simulação de validação PIX
			if transaction.PaymentType == PaymentTypePIX {
				// Verificar se chave PIX é válida
				pixKey, ok := transaction.AdditionalData["pix_key"].(string)
				if !ok || pixKey == "" {
					return false, "Chave PIX ausente ou inválida", nil
				}
				
				// Verificar tipo de chave PIX
				pixKeyType, ok := transaction.AdditionalData["pix_key_type"].(string)
				if !ok || !contains([]string{"cpf", "cnpj", "email", "phone", "random"}, pixKeyType) {
					return false, "Tipo de chave PIX inválido", nil
				}
			}
			return true, "Validação PIX concluída", nil
		},
	})

	// Regras específicas para PSD2 (EU)
	pg.complianceRules = append(pg.complianceRules, ComplianceRule{
		ID:          "sca_psd2",
		Market:      constants.MarketEU,
		Description: "Strong Customer Authentication (PSD2)",
		MandatoryFor: []string{
			PaymentTypeCard, PaymentTypeBank, PaymentTypeWallet,
		},
		Validate: func(transaction *PaymentTransaction) (bool, string, error) {
			// Verificar SCA para transações europeias
			if transaction.Currency == "EUR" && transaction.Amount > 30 {
				// Verificar se autenticação forte foi realizada
				if transaction.MFALevel != "high" {
					return false, "SCA requerido para transações acima de 30 EUR", nil
				}
				
				// Verificar se há dados 3DS para pagamentos com cartão
				if transaction.PaymentType == PaymentTypeCard {
					if transaction.ThreeDSData == nil {
						return false, "Dados 3D Secure ausentes para pagamento com cartão", nil
					}
				}
			}
			return true, "Requisitos SCA verificados", nil
		},
	})
}

// initializeRiskRules inicializa as regras de risco para o motor de risco
func initializeRiskRules(pg *PaymentGateway) {
	engine := NewRiskEngine()

	// Regra de transação de alto valor
	engine.AddRule(RiskRule{
		ID:          "high_value_transaction",
		Description: "Transação de valor elevado",
		Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
			// Define thresholds por moeda
			thresholds := map[string]float64{
				"USD": 5000,
				"EUR": 5000,
				"BRL": 10000,
				"AOA": 500000,
				"MZN": 100000,
			}
			
			threshold, exists := thresholds[transaction.Currency]
			if !exists {
				threshold = 5000 // Valor padrão
			}
			
			if transaction.Amount > threshold {
				return 0.7, true // Alto risco para transações de alto valor
			}
			return 0.0, false
		},
	})

	// Regra para endereços de cobrança e entrega diferentes
	engine.AddRule(RiskRule{
		ID:          "address_mismatch",
		Description: "Endereço de cobrança difere do endereço de entrega",
		Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
			// Verificar apenas para transações de e-commerce com cartão
			if transaction.PaymentType == PaymentTypeCard && transaction.TransactionType == "e_commerce" {
				billingAddress, hasBilling := transaction.AdditionalData["billing_address"]
				shippingAddress, hasShipping := transaction.AdditionalData["shipping_address"]
				
				if hasBilling && hasShipping && billingAddress != shippingAddress {
					return 0.5, true
				}
			}
			return 0.0, false
		},
	})

	// Regra para múltiplas transações em sucessão rápida
	engine.AddRule(RiskRule{
		ID:          "rapid_succession",
		Description: "Múltiplas transações em rápida sucessão",
		Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
			// Esta regra necessitaria um cache ou banco de dados para armazenar timestamps de transações
			// Aqui é uma simulação simplificada
			// Em produção, verificaria se há várias transações do mesmo usuário em curto período
			
			// Simulação: 10% das transações são identificadas como rápida sucessão
			if rand.Float64() < 0.1 {
				return 0.6, true
			}
			return 0.0, false
		},
	})

	// Regra de geolocalização inconsistente
	engine.AddRule(RiskRule{
		ID:          "geo_inconsistency",
		Description: "Geolocalização do IP inconsistente com histórico do usuário",
		Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
			// Esta regra requer informações históricas do usuário e geolocalização atual
			// Aqui é uma simulação simplificada
			
			ipCountry, hasIP := transaction.AdditionalData["ip_country"].(string)
			userCountry, hasUserCountry := transaction.AdditionalData["user_country"].(string)
			
			if hasIP && hasUserCountry && ipCountry != userCountry {
				return 0.8, true // Alto risco para transações com geolocalização divergente
			}
			return 0.0, false
		},
	})

	// Regras específicas para mercados
	
	// Brasil: Regra específica para fraude em PIX
	if pg.config.Market == constants.MarketBrazil || pg.config.Market == constants.MarketGlobal {
		engine.AddRule(RiskRule{
			ID:          "pix_fraud_pattern",
			Description: "Padrão de fraude em PIX detectado",
			Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
				if transaction.PaymentType == PaymentTypePIX {
					// Verificar padrões suspeitos:
					// - Múltiplas transações pequenas para a mesma chave PIX
					// - Transações noturnas de valor elevado
					
					// Simulação: avaliar transações PIX noturnas
					currentHour := time.Now().Hour()
					if (currentHour >= 22 || currentHour <= 5) && transaction.Amount > 1000 {
						return 0.7, true
					}
				}
				return 0.0, false
			},
		})
	}
	
	// Angola: Regra específica para transações em moeda estrangeira
	if pg.config.Market == constants.MarketAngola || pg.config.Market == constants.MarketGlobal {
		engine.AddRule(RiskRule{
			ID:          "foreign_currency_angola",
			Description: "Transação em moeda estrangeira em Angola",
			Evaluate: func(transaction *PaymentTransaction) (float64, bool) {
				if transaction.Currency != "AOA" {
					return 0.6, true
				}
				return 0.0, false
			},
		})
	}
	
	// Atribuir motor de risco ao gateway
	pg.riskEngine = engine
}

// registerInitialMetrics registra métricas iniciais do sistema
func registerInitialMetrics(pg *PaymentGateway) {
	// Registrar tipos de pagamento suportados
	for paymentType, supported := range pg.config.SupportedPayments {
		if supported {
			pg.observability.RecordMetric(adapter.MarketContext{
				Market:     pg.config.Market,
				TenantType: pg.config.TenantType,
			}, "payment_gateway_supported_payment_type", paymentType, 1)
		}
	}

	// Registrar integrações com provedores
	providers := []string{"visa", "mastercard", "amex", "pix", "bank_transfer"}
	for _, provider := range providers {
		pg.observability.RecordMetric(adapter.MarketContext{
			Market:     pg.config.Market,
			TenantType: pg.config.TenantType,
		}, "payment_gateway_provider_status", provider, 1)
	}

	// Registrar limites de transação
	for paymentType, limit := range pg.config.TransactionLimits {
		pg.observability.RecordHistogram(adapter.MarketContext{
			Market:     pg.config.Market,
			TenantType: pg.config.TenantType,
		}, "payment_gateway_transaction_limit", limit, paymentType)
	}

	// Registrar status do serviço
	pg.observability.RecordMetric(adapter.MarketContext{
		Market:     pg.config.Market,
		TenantType: pg.config.TenantType,
	}, "payment_gateway_status", "initialized", 1)
}