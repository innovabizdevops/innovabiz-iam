// Package testutil fornece utilitários compartilhados para testes do módulo IAM
package testutil

import (
	"fmt"
	"math/rand"
	"time"
)

// AuthorizationRequest representa uma requisição de autorização para testes
// Estrutura completa para avaliar decisões de autorização em vários contextos
type AuthorizationRequest struct {
	TestID      string                 // Identificador único para rastreabilidade
	Subject     string                 // Sujeito solicitante (usuário, serviço)
	Resource    string                 // Recurso sendo acessado
	Action      string                 // Ação solicitada (read, write, delete)
	Environment *Environment           // Contexto ambiental da requisição
	Tenant      string                 // Identificador do tenant (multi-tenancy)
	Context     string                 // Contexto operacional (normal, emergência)
	Attributes  map[string]interface{} // Atributos adicionais para ABAC
	Metadata    map[string]interface{} // Metadados de teste
}

// Environment representa o ambiente de execução da requisição de autorização
// Captura fatores contextuais que influenciam decisões de autorização adaptativa
type Environment struct {
	Type               string    // Tipo de ambiente (dev, qa, prod, dr)
	Context            string    // Contexto operacional
	IPAddress          string    // Endereço IP de origem
	AccessTime         time.Time // Momento do acesso
	RiskScore          float64   // Pontuação de risco calculada
	IsRecognizedDevice bool      // Dispositivo reconhecido previamente
	Location           string    // Localização geográfica
	Market             string    // Mercado específico (Angola, Brasil, etc.)
}

// AuthorizationTestDataGenerator gera dados de teste para cenários de autorização
// Implementa abordagem paramétrica para cobrir diversas condições de teste
type AuthorizationTestDataGenerator struct {
	roles            []string            // Papéis de acesso disponíveis
	resources        []string            // Recursos protegidos
	actions          []string            // Ações possíveis
	environmentTypes []string            // Tipos de ambiente
	contexts         []string            // Contextos operacionais
	tenants          []string            // Tenants disponíveis
	markets          []string            // Mercados geográficos
	riskProfiles     map[string]float64  // Perfis de risco pré-configurados
	deviceProfiles   map[string]bool     // Perfis de dispositivos
}

// NewAuthorizationTestDataGenerator cria uma nova instância do gerador de dados
// Inicializa com conjuntos realistas de valores para geração paramétrica
func NewAuthorizationTestDataGenerator() *AuthorizationTestDataGenerator {
	return &AuthorizationTestDataGenerator{
		// Papéis representando hierarquia organizacional e funcional
		roles: []string{
			"admin", "manager", "operator", "auditor", "developer",
			"security_admin", "compliance_officer", "business_analyst",
			"support_level_1", "support_level_2", "support_level_3",
			"finance_officer", "finance_manager", "finance_auditor",
			"risk_analyst", "ciso", "cio", "payment_approver",
			"marketplace_vendor", "marketplace_admin", "customer_support",
		},
		
		// Recursos dos diversos servidores MCP
		resources: []string{
			"mcp_docker:kubectl", "mcp_docker:docker", "mcp_docker:helm",
			"desktop_commander:filesystem", "desktop_commander:config", "desktop_commander:process",
			"github:repository", "github:pullrequest", "github:issue", "github:secrets",
			"memory:graph", "memory:entity", "memory:relation", "memory:observation",
			"figma:design", "figma:comment", "figma:project", "figma:team",
			"payment_gateway:transaction", "payment_gateway:refund", "payment_gateway:config",
			"mobile_money:wallet", "mobile_money:transfer", "mobile_money:settings",
			"risk_engine:rules", "risk_engine:alerts", "risk_engine:reports",
			"crm:customer", "crm:interaction", "crm:campaign", "crm:report",
			"marketplace:product", "marketplace:order", "marketplace:vendor", "marketplace:review",
		},
		
		// Ações possíveis sobre recursos
		actions: []string{
			"read", "write", "delete", "execute", "create", "update", 
			"approve", "reject", "list", "search", "analyze", "transfer",
			"process", "configure", "enable", "disable", "escalate",
			"cancel", "suspend", "resume", "export", "import",
		},
		
		// Tipos de ambiente para testing e produção
		environmentTypes: []string{
			"development", "testing", "staging", "production", "disaster_recovery",
			"sandbox", "demo", "training", "audit",
		},
		
		// Contextos operacionais com variações de risco
		contexts: []string{
			"normal", "emergency", "maintenance", "audit", "recovery",
			"investigation", "security_incident", "compliance_review",
			"customer_support", "batch_processing",
		},
		
		// Implementação multi-tenant para testes de isolamento
		tenants: []string{
			"tenant_angola_1", "tenant_angola_2", "tenant_brasil_1",
			"tenant_portugal_1", "tenant_eua_1", "tenant_china_1",
			"tenant_sadc_1", "tenant_cplp_1", "tenant_palop_1",
			"tenant_global", "tenant_eu_1", "tenant_africa_1",
		},
		
		// Mercados geográficos conforme requisitos do projeto
		markets: []string{
			"Angola", "Brasil", "Portugal", "EUA", "China",
			"Moçambique", "Cabo Verde", "Guiné-Bissau", "São Tomé e Príncipe",
			"África do Sul", "Namíbia", "União Europeia", "Global",
		},
		
		// Perfis de risco pré-definidos para simulação
		riskProfiles: map[string]float64{
			"very_low":    0.05,
			"low":         0.2,
			"medium_low":  0.35,
			"medium":      0.5,
			"medium_high": 0.65,
			"high":        0.8,
			"very_high":   0.95,
		},
		
		// Perfis de dispositivos para simulação de contexto
		deviceProfiles: map[string]bool{
			"registered_corporate": true,
			"registered_personal":  true,
			"new_device":           false,
			"unknown_device":       false,
		},
	}
}

// GenerateAuthRequest gera uma requisição de autorização com parâmetros aleatórios
// Ideal para testes de robustez e cobertura ampla de cenários
func (g *AuthorizationTestDataGenerator) GenerateAuthRequest() *AuthorizationRequest {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Gerar ID único para este teste com timestamp para rastreabilidade
	testID := fmt.Sprintf("test-%d", time.Now().UnixNano())
	
	// Selecionar elementos aleatórios dos conjuntos disponíveis
	role := g.roles[r.Intn(len(g.roles))]
	resource := g.resources[r.Intn(len(g.resources))]
	action := g.actions[r.Intn(len(g.actions))]
	env := g.environmentTypes[r.Intn(len(g.environmentTypes))]
	ctx := g.contexts[r.Intn(len(g.contexts))]
	tenant := g.tenants[r.Intn(len(g.tenants))]
	market := g.markets[r.Intn(len(g.markets))]
	
	// Selecionar perfil de risco e de dispositivo
	riskKeys := make([]string, 0, len(g.riskProfiles))
	for k := range g.riskProfiles {
		riskKeys = append(riskKeys, k)
	}
	riskProfile := riskKeys[r.Intn(len(riskKeys))]
	
	deviceKeys := make([]string, 0, len(g.deviceProfiles))
	for k := range g.deviceProfiles {
		deviceKeys = append(deviceKeys, k)
	}
	deviceProfile := deviceKeys[r.Intn(len(deviceKeys))]
	
	// Gerar IP aleatório para simulação de origem
	ip := fmt.Sprintf("%d.%d.%d.%d", 
		r.Intn(256), r.Intn(256), r.Intn(256), r.Intn(256))
	
	// Gerar tempo de acesso com variação realista
	var accessTime time.Time
	if r.Intn(100) < 75 { // 75% dentro do horário comercial
		hour := 8 + r.Intn(9) // Entre 8:00 e 17:00
		minute := r.Intn(60)
		accessTime = time.Date(2025, 8, 5, hour, minute, 0, 0, time.UTC)
	} else {
		hour := 18 + r.Intn(14) % 24 // Entre 18:00 e 7:00
		minute := r.Intn(60)
		accessTime = time.Date(2025, 8, 5, hour, minute, 0, 0, time.UTC)
	}
	
	// Construir a requisição completa de autorização
	return &AuthorizationRequest{
		TestID:   testID,
		Subject:  fmt.Sprintf("user:%s:%s", role, testID),
		Resource: fmt.Sprintf("%s:%s", resource, testID),
		Action:   action,
		Tenant:   tenant,
		Context:  ctx,
		Environment: &Environment{
			Type:               env,
			Context:            ctx,
			IPAddress:          ip,
			AccessTime:         accessTime,
			RiskScore:          g.riskProfiles[riskProfile],
			IsRecognizedDevice: g.deviceProfiles[deviceProfile],
			Location:           market,
			Market:             market,
		},
		Attributes: map[string]interface{}{
			"role":            role,
			"risk_profile":    riskProfile,
			"device_profile":  deviceProfile,
			"authentication_method": r.Intn(100) < 30 ? "mfa" : "password", // 30% MFA
			"session_duration": r.Intn(100) < 50 ? "extended" : "standard", // 50% sessão estendida
		},
		Metadata: map[string]interface{}{
			"test_generated": true,
			"generator":      "AuthorizationTestDataGenerator",
			"version":        "1.0",
			"timestamp":      time.Now().Format(time.RFC3339),
		},
	}
}

// GenerateTestDataset gera conjunto de dados de teste com tamanho especificado
// Útil para testes de volume, performance e análise estatística
func (g *AuthorizationTestDataGenerator) GenerateTestDataset(size int) []*AuthorizationRequest {
	dataset := make([]*AuthorizationRequest, size)
	for i := 0; i < size; i++ {
		dataset[i] = g.GenerateAuthRequest()
	}
	return dataset
}

// GenerateMultiTenantDataset gera dados específicos para testes multi-tenant
// Garante isolamento e separação adequada entre tenants
func (g *AuthorizationTestDataGenerator) GenerateMultiTenantDataset() map[string][]*AuthorizationRequest {
	result := make(map[string][]*AuthorizationRequest)
	
	// Para cada tenant, gerar 5 requisições
	for _, tenant := range g.tenants {
		requests := make([]*AuthorizationRequest, 5)
		for i := 0; i < 5; i++ {
			req := g.GenerateAuthRequest()
			req.Tenant = tenant
			requests[i] = req
		}
		result[tenant] = requests
	}
	
	return result
}

// GenerateSoDTestData gera dados específicos para teste de segregação de deveres
// Cria conjuntos de papéis conflitantes e requisições que acionariam verificação SoD
func (g *AuthorizationTestDataGenerator) GenerateSoDTestData() ([]RoleAssignment, []AuthorizationRequest) {
	// Definir papéis conflitantes por política de SoD
	conflictingRoles := [][]string{
		{"finance_manager", "finance_auditor"},
		{"system_admin", "security_auditor"},
		{"developer", "quality_assurance"},
		{"procurement_officer", "payment_approver"},
		{"marketplace_vendor", "marketplace_admin"},
	}
	
	// Gerar usuários de teste
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	
	// Atribuir papéis (alguns com conflito)
	roleAssignments := make([]RoleAssignment, 0)
	for i, user := range users {
		roleSet := i % len(conflictingRoles)
		
		// Primeiro papel sempre atribuído
		roleAssignments = append(roleAssignments, RoleAssignment{
			UserID: user,
			RoleID: conflictingRoles[roleSet][0],
		})
		
		// 50% de chance de ter papel conflitante
		if rand.Intn(100) < 50 {
			roleAssignments = append(roleAssignments, RoleAssignment{
				UserID: user,
				RoleID: conflictingRoles[roleSet][1],
			})
		}
	}
	
	// Gerar requisições que acionariam verificação de SoD
	requests := make([]AuthorizationRequest, 0)
	for _, user := range users {
		requests = append(requests, AuthorizationRequest{
			Subject:  user,
			Resource: "financial_transaction:approve",
			Action:   "approve",
			Context:  "regular_operation",
		})
		
		requests = append(requests, AuthorizationRequest{
			Subject:  user,
			Resource: "system_config:modify",
			Action:   "write",
			Context:  "maintenance",
		})
		
		requests = append(requests, AuthorizationRequest{
			Subject:  user,
			Resource: "marketplace:vendor_approval",
			Action:   "approve",
			Context:  "regular_operation",
		})
	}
	
	return roleAssignments, requests
}

// RoleAssignment representa a atribuição de um papel a um usuário
type RoleAssignment struct {
	UserID string
	RoleID string
}

// GenerateComplexAccessScenarios gera cenários complexos para testes avançados
// Foca em casos limítrofes e cenários de decisão complexa
func (g *AuthorizationTestDataGenerator) GenerateComplexAccessScenarios() []*AuthorizationRequest {
	scenarios := []*AuthorizationRequest{
		// Cenário 1: Acesso crítico durante horário não comercial
		{
			Subject:  "user:admin:critical",
			Resource: "payment_gateway:config:production",
			Action:   "write",
			Tenant:   "tenant_angola_1",
			Context:  "emergency",
			Environment: &Environment{
				Type:               "production",
				Context:            "emergency",
				IPAddress:          "198.51.100.1", // IP não reconhecido
				AccessTime:         time.Date(2025, 8, 5, 2, 30, 0, 0, time.UTC), // 2:30 AM
				RiskScore:          0.85,
				IsRecognizedDevice: false,
				Location:           "Angola",
				Market:             "Angola",
			},
			Attributes: map[string]interface{}{
				"role":                 "admin",
				"emergency_authorized": true,
				"incident_id":          "INC-2025-08-05-001",
			},
		},
		
		// Cenário 2: Operação de alto risco de localização não usual
		{
			Subject:  "user:finance_manager:high_risk",
			Resource: "mobile_money:batch_transfer",
			Action:   "execute",
			Tenant:   "tenant_angola_2",
			Context:  "normal",
			Environment: &Environment{
				Type:               "production",
				Context:            "normal",
				IPAddress:          "203.0.113.42", // IP de região incomum
				AccessTime:         time.Date(2025, 8, 5, 14, 20, 0, 0, time.UTC), // Horário comercial
				RiskScore:          0.75,
				IsRecognizedDevice: true,
				Location:           "Desconhecida", // Localização não usual
				Market:             "Angola",
			},
			Attributes: map[string]interface{}{
				"role":              "finance_manager",
				"transfer_amount":   "5000000", // Valor elevado
				"beneficiaries":     50,        // Muitos beneficiários
				"previous_approval": "missing",
			},
		},
		
		// Cenário 3: Tentativa de elevação de privilégio via API
		{
			Subject:  "user:developer:escalation",
			Resource: "desktop_commander:config:allowedDirectories",
			Action:   "write",
			Tenant:   "tenant_global",
			Context:  "normal",
			Environment: &Environment{
				Type:               "production",
				Context:            "normal",
				IPAddress:          "192.168.1.100",
				AccessTime:         time.Date(2025, 8, 5, 10, 15, 0, 0, time.UTC),
				RiskScore:          0.60,
				IsRecognizedDevice: true,
				Location:           "Brasil",
				Market:             "Brasil",
			},
			Attributes: map[string]interface{}{
				"role":                    "developer",
				"elevation_request":       true,
				"justification":           "Configuração de ambiente de desenvolvimento",
				"requested_access_period": 60, // 60 minutos
			},
		},
		
		// Cenário 4: Acesso legítimo mas incomum via marketplace
		{
			Subject:  "user:marketplace_vendor:unusual",
			Resource: "marketplace:product:bulk_update",
			Action:   "write",
			Tenant:   "tenant_brasil_1",
			Context:  "normal",
			Environment: &Environment{
				Type:               "production",
				Context:            "normal",
				IPAddress:          "198.51.100.42",
				AccessTime:         time.Date(2025, 8, 5, 22, 30, 0, 0, time.UTC), // 22:30 PM
				RiskScore:          0.45,
				IsRecognizedDevice: true,
				Location:           "Brasil",
				Market:             "Brasil",
			},
			Attributes: map[string]interface{}{
				"role":                "marketplace_vendor",
				"products_affected":   120, // Grande volume
				"account_age_days":    5,   // Conta recente
				"previous_operations": 3,   // Poucas operações anteriores
			},
		},
		
		// Cenário 5: Operação crítica durante manutenção programada
		{
			Subject:  "user:operator:maintenance",
			Resource: "mcp_docker:kubectl:deployment",
			Action:   "delete",
			Tenant:   "tenant_eua_1",
			Context:  "maintenance",
			Environment: &Environment{
				Type:               "production",
				Context:            "maintenance",
				IPAddress:          "10.0.0.15", // IP interno
				AccessTime:         time.Date(2025, 8, 5, 3, 45, 0, 0, time.UTC), // 3:45 AM
				RiskScore:          0.35,
				IsRecognizedDevice: true,
				Location:           "EUA",
				Market:             "EUA",
			},
			Attributes: map[string]interface{}{
				"role":                  "operator",
				"maintenance_window_id": "MW-2025-08-05-01",
				"change_request_id":     "CR-2025-07-30-42",
				"approved_by":           "system_admin",
			},
		},
	}
	
	// Adicionar metadados a todos os cenários
	for _, scenario := range scenarios {
		if scenario.Metadata == nil {
			scenario.Metadata = make(map[string]interface{})
		}
		scenario.Metadata["test_type"] = "complex_scenario"
		scenario.Metadata["generator"] = "AuthorizationTestDataGenerator"
		scenario.Metadata["version"] = "1.0"
		scenario.Metadata["timestamp"] = time.Now().Format(time.RFC3339)
	}
	
	return scenarios
}