# Plano de Testes Automatizados - Autorização Avançada MCP-IAM

## Sumário Executivo

Este documento define o plano de testes automatizados para o sistema de Autorização Avançada MCP-IAM da plataforma INNOVABIZ, assegurando que todas as funcionalidades críticas e integrações com os servidores MCP estejam adequadamente testadas, seguras e conformes com normas internacionais como ISO 27001, ISO 42001, NIST CSF, SOC 2, PCI-DSS, LGPD, GDPR e regulamentações específicas de Open Banking e Open Finance.

### Visão Geral

O sistema de Autorização Avançada MCP-IAM é responsável por controlar o acesso às funcionalidades dos diversos servidores MCP (Model Context Protocol) dentro da plataforma INNOVABIZ, implementando modelos avançados de autorização como RBAC, ABAC, ReBAC e CBAC, bem como mecanismos sofisticados de elevação temporária de privilégios, avaliação de risco, delegação de autoridade e segregação de deveres.

### Objetivos do Plano de Testes

1. **Garantia de Qualidade**: Validar que todas as funcionalidades de autorização operam conforme as especificações
2. **Verificação de Segurança**: Confirmar que os controles de segurança são efetivos e resilientes a ataques
3. **Conformidade Regulatória**: Assegurar que o sistema atende a todos os requisitos regulatórios aplicáveis
4. **Performance e Escalabilidade**: Validar que o sistema mantém desempenho adequado sob carga
5. **Integração Confiável**: Garantir a integração correta com todos os servidores MCP e módulos da plataforma

### Escopo

O plano de testes abrange:

1. **Núcleo de Autorização**: Engine OPA, resolução de papéis, avaliação de risco, políticas e regras
2. **Mecanismos Avançados**: Elevação temporária, delegação, quorum, SoD
3. **Integrações MCP**: Docker, Desktop-Commander, GitHub, Memory, Figma
4. **Auditoria e Rastreabilidade**: Registro de eventos, análise de anomalias
5. **Conformidade**: Validação contra frameworks regulatórios

## 1. Estratégia de Testes

### 1.1. Abordagem em Camadas

O plano implementa uma estratégia de testes em múltiplas camadas para garantir cobertura abrangente:

| Camada | Foco | Abordagem | Ferramentas |
|--------|------|-----------|------------|
| Unitário | Componentes individuais | White-box, mocks | Go testing, testify, gomock |
| Integração | Interações entre componentes | Black/Gray-box | Go testing, testcontainers |
| API | Endpoints REST/GraphQL | Black-box | Postman, Newman, K6 |
| E2E | Fluxos completos | Black-box | Cypress, Selenium |
| Segurança | Vulnerabilidades | Especializado | OWASP ZAP, SonarQube |
| Performance | Comportamento sob carga | Especializado | K6, Gatling, Prometheus |
| Conformidade | Requisitos regulatórios | Especializado | OPA Conftest, Checkov |

### 1.2. Abordagem BDD/TDD

Os testes serão desenvolvidos seguindo as metodologias BDD (Behavior-Driven Development) e TDD (Test-Driven Development):

- **Especificação BDD**: Cenários em formato Gherkin para capturar requisitos de negócio
- **TDD para Implementação**: Desenvolvimento guiado por testes para componentes core
- **Documentação Viva**: Testes como documentação executável do comportamento do sistema

### 1.3. CI/CD e DevSecOps

A automação de testes será integrada no pipeline CI/CD:

- **Testes Unitários**: Executados em cada commit
- **Testes de Integração**: Executados em cada PR/MR
- **Testes E2E**: Executados antes de deploy em ambientes
- **Testes de Segurança**: Executados diariamente e em PRs
- **Testes de Performance**: Executados semanalmente e antes de releases
- **Testes de Conformidade**: Executados antes de deploy em produção

### 1.4. Ambientes de Teste

| Ambiente | Propósito | Configuração | Dados |
|----------|-----------|--------------|-------|
| Desenvolvimento | Testes unitários e integração básica | Containers locais | Dados sintéticos |
| QA | Testes de integração e E2E | Cluster K8s dedicado | Dados anonimizados |
| Staging | Validação pré-produção | Espelho de produção | Dados anonimizados |
| Sandbox | Testes de conformidade | Configuração específica | Datasets certificados |
| Performance | Testes de carga | Infraestrutura escalável | Dados volumétricos |

## 2. Tipos de Testes

### 2.1. Testes Unitários

#### 2.1.1. Componentes Core

| Componente | Aspectos Testados | Técnicas | Cobertura Mínima |
|------------|------------------|-----------|-----------------|
| OPA Engine | Avaliação de políticas, integridade | Mocking, table-driven | 95% |
| Risk Evaluator | Cálculo de risco, thresholds | Table-driven, property-based | 90% |
| Role Resolver | Resolução hierárquica, herança | Graph-based testing | 95% |
| Decision Cache | Armazenamento, invalidação | Concurrency testing | 90% |
| Elevation Manager | Aprovações, expiração | Time-based testing | 95% |

#### 2.1.2. Políticas OPA

| Conjunto de Políticas | Aspectos Testados | Técnicas | Cobertura Mínima |
|----------------------|------------------|-----------|-----------------|
| Base Policies | Lógica de decisão básica | Table-driven | 100% |
| RBAC Policies | Hierarquia de papéis | Graph-based testing | 95% |
| ABAC Policies | Avaliação de atributos | Property-based | 90% |
| ReBAC Policies | Resolução de relacionamentos | Graph-based testing | 90% |
| CBAC Policies | Avaliação contextual | Context simulation | 90% |

#### 2.1.3. Exemplo de Teste Unitário (Go)

```go
func TestRiskEvaluator_CalculateScore(t *testing.T) {
    tests := []struct {
        name     string
        ctx      *SecurityContext
        request  *ToolRequest
        expected float64
        epsilon  float64
    }{
        {
            name: "baixo risco - usuário confiável, operação comum",
            ctx: &SecurityContext{
                User: User{
                    RiskHistory:               0.1,
                    SecurityTrainingCompleted: true,
                    LoginAnomalies:            0,
                },
                IPAddress:        "192.168.1.1",
                IPReputationScore: 0.9,
                Time:             time.Date(2025, 8, 5, 14, 30, 0, 0, time.UTC),
                UserLocation:     "Angola",
                Device: Device{
                    Recognized:   true,
                    RiskScore:    0.1,
                },
            },
            request: &ToolRequest{
                Tool: Tool{
                    Name:   "mcp1_read_file",
                    Server: "desktop-commander",
                },
                Parameters: map[string]interface{}{
                    "path": "/public/docs/readme.md",
                },
            },
            expected: 0.15,
            epsilon:  0.05,
        },
        {
            name: "alto risco - horário incomum, operação sensível",
            ctx: &SecurityContext{
                User: User{
                    RiskHistory:               0.4,
                    SecurityTrainingCompleted: false,
                    LoginAnomalies:            2,
                },
                IPAddress:        "203.0.113.1",
                IPReputationScore: 0.2,
                Time:             time.Date(2025, 8, 5, 3, 15, 0, 0, time.UTC),
                UserLocation:     "Unknown",
                Device: Device{
                    Recognized:   false,
                    RiskScore:    0.7,
                },
            },
            request: &ToolRequest{
                Tool: Tool{
                    Name:   "mcp1_set_config_value",
                    Server: "desktop-commander",
                },
                Parameters: map[string]interface{}{
                    "key":   "allowedDirectories",
                    "value": []string{},
                },
            },
            expected: 0.85,
            epsilon:  0.05,
        },
    }

    evaluator := NewRiskEvaluator()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            score := evaluator.CalculateScore(tt.ctx, tt.request)
            
            if math.Abs(score-tt.expected) > tt.epsilon {
                t.Errorf("CalculateScore() = %v, want %v (±%v)", score, tt.expected, tt.epsilon)
            }
        })
    }
}
```

### 2.2. Testes de Integração#### 2.2.1. Integrações Entre Componentes

| Integração | Aspectos Testados | Técnicas | Ferramentas |
|------------|------------------|-----------|------------|
| IAM Core + OPA Engine | Avaliação de políticas em tempo real | API testing | Testcontainers, OPA |
| Risk Engine + Decision Engine | Adaptação baseada em risco | Simulação de contexto | Docker Compose |
| Audit + Storage | Persistência de eventos | Mock externo + verificação | Testcontainers |
| IAM + MCP Servers | Comunicação entre módulos | Simulação de servidores | Wiremock, gRPC mock |
| Cache + Invalidation | Gestão de cache distribuído | Testes de concorrência | Redis, etcd |

#### 2.2.2. Integrações com Servidores MCP

| Servidor MCP | Aspectos Testados | Abordagem | Simulação |
|--------------|------------------|-----------|-----------|
| MCP_DOCKER | Autorização para operações K8s/Docker | API mocking | Simulação de respostas K8s |
| Desktop-Commander | Autorização para sistema de arquivos | Virtualização | Diretório virtualizado |
| GitHub MCP | Verificação de acesso a repositórios | API mocking | GitHub API simulada |
| Memory MCP | Controle de acesso a dados de memória | API mocking | Simulação de backend |
| Figma MCP | Permissões para recursos de design | API mocking | Endpoints simulados |

#### 2.2.3. Exemplo de Teste de Integração (Go)

```go
func TestDockerAuthzHooks_IntegrationWithOPA(t *testing.T) {
    // Iniciar OPA em container de teste
    ctx := context.Background()
    opaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "openpolicyagent/opa:latest",
            ExposedPorts: []string{"8181/tcp"},
            Cmd:          []string{"run", "--server", "--addr", ":8181"},
            WaitingFor:   wait.ForHTTP("/health").WithPort("8181/tcp"),
        },
        Started: true,
    })
    if err != nil {
        t.Fatalf("Falha ao iniciar container OPA: %v", err)
    }
    defer opaContainer.Terminate(ctx)
    
    // Obter porta mapeada
    opaPort, err := opaContainer.MappedPort(ctx, "8181/tcp")
    if err != nil {
        t.Fatalf("Falha ao obter porta do OPA: %v", err)
    }
    opaURL := fmt.Sprintf("http://localhost:%s", opaPort.Port())
    
    // Carregar políticas de teste para OPA
    policyData := `
package innovabiz.mcp_docker

default allow = false

allow {
    input.user.roles[_] == "k8s_admin"
    input.tool.name == "mcp0_kubectl_get"
}
`
    client := resty.New()
    _, err = client.R().
        SetHeader("Content-Type", "text/plain").
        SetBody(policyData).
        Put(opaURL + "/v1/policies/docker_authz")
    if err != nil {
        t.Fatalf("Falha ao carregar política: %v", err)
    }
    
    // Criar hooks com dependência em OPA
    hooks := &DockerAuthzHooks{
        baseAuthorizer: &OPAAuthorizer{
            client:    client,
            opaURL:    opaURL,
            opaPath:   "/v1/data/innovabiz/mcp_docker/allow",
            cacheTTL:  time.Second * 30,
        },
        k8sRBACResolver:  newMockK8sRBACResolver(),
        imageRegistry:    newMockImageRegistry(),
        namespaceManager: newMockNamespaceManager(),
        complianceChecker: newMockComplianceChecker(),
        auditLogger:      newMockAuditLogger(),
    }
    
    // Definir casos de teste
    tests := []struct {
        name     string
        ctx      *SecurityContext
        request  *ToolRequest
        expected bool
    }{
        {
            name: "permitir acesso para administrador",
            ctx: &SecurityContext{
                User: User{
                    ID:    "user123",
                    Roles: []string{"k8s_admin"},
                },
            },
            request: &ToolRequest{
                Tool: Tool{
                    Name:   "mcp0_kubectl_get",
                    Server: "mcp_docker",
                },
                Parameters: map[string]interface{}{
                    "namespace":    "default",
                    "resourceType": "pods",
                },
            },
            expected: true,
        },
        {
            name: "negar acesso para usuário comum",
            ctx: &SecurityContext{
                User: User{
                    ID:    "user456",
                    Roles: []string{"developer"},
                },
            },
            request: &ToolRequest{
                Tool: Tool{
                    Name:   "mcp0_kubectl_get",
                    Server: "mcp_docker",
                },
                Parameters: map[string]interface{}{
                    "namespace":    "default",
                    "resourceType": "pods",
                },
            },
            expected: false,
        },
    }
    
    // Executar testes
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            allowed, _, _ := hooks.PreExecuteHook(tt.ctx, tt.request)
            
            if allowed != tt.expected {
                t.Errorf("PreExecuteHook() allowed = %v, expected %v", allowed, tt.expected)
            }
        })
    }
}
```

### 2.3. Testes de API

#### 2.3.1. Endpoints de Autorização

| Endpoint | Método | Aspectos Testados | Casos de Teste |
|----------|--------|------------------|---------------|
| `/v1/auth/decision` | POST | Decisão de autorização | Permissões válidas/inválidas, diferentes contextos |
| `/v1/auth/check` | POST | Verificação rápida | Verificações baseadas em tokens |
| `/v1/auth/batch` | POST | Decisões em lote | Múltiplas decisões, resultados agregados |
| `/v1/auth/tools/{tool}` | GET | Ferramentas permitidas | Listagem por usuário/contexto |
| `/v1/auth/elevation/request` | POST | Elevação temporária | Solicitações válidas/inválidas |

#### 2.3.2. Endpoints GraphQL

| Operação | Aspectos Testados | Variáveis | Assertivas |
|----------|------------------|-----------|-----------|
| `authorizeToolUse` | Decisão de autorização | Ferramenta, parâmetros, contexto | Status, motivo |
| `getUserPermissions` | Listagem de permissões | ID do usuário | Lista completa, papéis |
| `checkElevationStatus` | Status de elevação | ID de elevação | Estado atual, expiração |
| `getApprovalRequests` | Solicitações pendentes | Filtros, paginação | Lista correta, contadores |

#### 2.3.3. Exemplo de Teste de API (Postman/Newman)

```json
{
  "info": {
    "name": "Testes de Autorização MCP-IAM",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Autorizar uso de ferramenta - Caso positivo",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          },
          {
            "key": "Authorization",
            "value": "Bearer {{token_admin}}"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"tool\": {\n    \"name\": \"mcp1_read_file\",\n    \"server\": \"desktop-commander\"\n  },\n  \"parameters\": {\n    \"path\": \"/public/docs/readme.md\"\n  },\n  \"context\": {\n    \"ipAddress\": \"192.168.1.1\",\n    \"userLocation\": \"Angola\",\n    \"deviceRecognized\": true\n  }\n}"
        },
        "url": {
          "raw": "{{base_url}}/v1/auth/decision",
          "host": ["{{base_url}}"],
          "path": ["v1", "auth", "decision"]
        }
      },
      "test": [
        "pm.test('Status code é 200', function() {",
        "  pm.response.to.have.status(200);",
        "});",
        "",
        "pm.test('Decisão é permitida', function() {",
        "  var jsonData = pm.response.json();",
        "  pm.expect(jsonData.allowed).to.be.true;",
        "});",
        "",
        "pm.test('Decisão inclui contexto expandido', function() {",
        "  var jsonData = pm.response.json();",
        "  pm.expect(jsonData.context).to.be.an('object');",
        "  pm.expect(jsonData.context.riskScore).to.be.a('number');",
        "});",
        "",
        "pm.test('Decisão inclui metadados de auditoria', function() {",
        "  var jsonData = pm.response.json();",
        "  pm.expect(jsonData.audit).to.be.an('object');",
        "  pm.expect(jsonData.audit.decisionId).to.be.a('string');",
        "  pm.expect(jsonData.audit.timestamp).to.be.a('string');",
        "});"
      ]
    },
    {
      "name": "Autorizar uso de ferramenta - Caso negativo",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          },
          {
            "key": "Authorization",
            "value": "Bearer {{token_basic}}"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"tool\": {\n    \"name\": \"mcp1_set_config_value\",\n    \"server\": \"desktop-commander\"\n  },\n  \"parameters\": {\n    \"key\": \"allowedDirectories\",\n    \"value\": []\n  },\n  \"context\": {\n    \"ipAddress\": \"192.168.1.1\",\n    \"userLocation\": \"Angola\",\n    \"deviceRecognized\": true\n  }\n}"
        },
        "url": {
          "raw": "{{base_url}}/v1/auth/decision",
          "host": ["{{base_url}}"],
          "path": ["v1", "auth", "decision"]
        }
      },
      "test": [
        "pm.test('Status code é 403', function() {",
        "  pm.response.to.have.status(403);",
        "});",
        "",
        "pm.test('Decisão é negada com motivo', function() {",
        "  var jsonData = pm.response.json();",
        "  pm.expect(jsonData.allowed).to.be.false;",
        "  pm.expect(jsonData.reason).to.be.a('string').and.not.empty;",
        "});"
      ]
    }
  ]
}
```

### 2.4. Testes End-to-End (E2E)

#### 2.4.1. Fluxos Principais

| Fluxo | Descrição | Componentes Envolvidos | Validações |
|-------|-----------|------------------------|------------|
| Autorização Básica | Controle de acesso padrão | IAM Core, OPA, MCP Servers | Decisões corretas por papel |
| Autorização Adaptativa | Ajustes baseados em risco | Risk Engine, Decision Engine | Adaptação a diferentes níveis de risco |
| Elevação de Privilégios | Just-in-Time Access | Elevation Manager, Approval System | Workflow completo de aprovação |
| Segregação de Deveres | Prevenção de conflito | SoD Checker, Audit System | Detecção correta de violações |
| Delegação de Autoridade | Transferência controlada | Delegation Manager, Audit | Limites respeitados, auditoria |

#### 2.4.2. Cenários BDD

Os cenários abaixo seguem o formato Gherkin para definir comportamentos esperados em situações específicas:

```gherkin
Feature: Autorização Adaptativa Baseada em Risco

  Background:
    Given um usuário "administrador" com papel "system_admin"
    And uma ferramenta "mcp1_set_config_value" no servidor "desktop-commander"
    And o nível de risco base do usuário é "baixo"

  Scenario: Autorização com risco baixo
    Given o usuário acessa de um dispositivo reconhecido
    And o acesso ocorre durante o horário comercial
    And o usuário acessa de uma localização conhecida
    When o usuário tenta usar a ferramenta "mcp1_set_config_value"
    Then a autorização é concedida
    And não há requisitos adicionais de verificação

  Scenario: Autorização com risco médio
    Given o usuário acessa de um dispositivo novo
    And o acesso ocorre durante o horário comercial
    And o usuário acessa de uma localização conhecida
    When o usuário tenta usar a ferramenta "mcp1_set_config_value"
    Then a autorização é concedida
    But requer verificação MFA adicional

  Scenario: Autorização com risco alto
    Given o usuário acessa de um dispositivo desconhecido
    And o acesso ocorre fora do horário comercial
    And o usuário acessa de uma localização incomum
    When o usuário tenta usar a ferramenta "mcp1_set_config_value"
    Then a autorização é negada
    And é registrado um alerta de segurança
    And o usuário é notificado para contatar a segurança

Feature: Elevação Temporária de Privilégios

  Background:
    Given um usuário "desenvolvedor" com papel "developer"
    And uma ferramenta "mcp0_kubectl_delete" no servidor "mcp_docker"
    And o usuário não tem permissão padrão para essa ferramenta

  Scenario: Solicitação de elevação aprovada
    Given o usuário solicita elevação para papel "k8s_operator"
    And fornece justificativa "Limpeza de pods em falha"
    And um aprovador "teamlead" está disponível
    When o aprovador revisa e aprova a solicitação
    Then o usuário recebe acesso temporário ao papel "k8s_operator"
    And o acesso expira após 2 horas
    And todas as ações são registradas em auditoria

  Scenario: Solicitação de elevação negada
    Given o usuário solicita elevação para papel "k8s_admin"
    And fornece justificativa "Teste de permissões"
    And um aprovador "seguranca" revisa a solicitação
    When o aprovador rejeita a solicitação
    Then o usuário não recebe acesso ao papel solicitado
    And o usuário é notificado da rejeição
    And o evento é registrado em auditoria de segurança
```

#### 2.4.3. Exemplo de Teste E2E (Cypress)

```javascript
describe('Autorização MCP com Elevação de Privilégios', () => {
  beforeEach(() => {
    // Setup: Login com usuário desenvolvedor
    cy.login('developer@innovabiz.com', 'Password123!');
    cy.visit('/mcp-dashboard');
  });

  it('deve solicitar e obter elevação temporária de privilégios', () => {
    // Verificar que botão de kubernetes está presente mas desabilitado
    cy.contains('button', 'Kubernetes Operations').should('be.disabled');
    
    // Solicitar elevação
    cy.contains('Request Elevation').click();
    cy.get('#elevation-role-select').select('k8s_operator');
    cy.get('#elevation-reason').type('Debugging pod failures in QA environment');
    cy.get('#elevation-duration').select('2 hours');
    cy.contains('button', 'Submit Request').click();
    
    // Verificar que solicitação foi enviada
    cy.contains('Elevation request submitted').should('be.visible');
    
    // Login como aprovador
    cy.login('teamlead@innovabiz.com', 'Password123!');
    cy.visit('/approvals');
    
    // Aprovar solicitação
    cy.contains('tr', 'k8s_operator').within(() => {
      cy.contains('button', 'Review').click();
    });
    cy.get('#approval-decision').select('Approve');
    cy.get('#approval-comment').type('Approved for QA debugging');
    cy.contains('button', 'Submit Decision').click();
    
    // Verificar que aprovação foi registrada
    cy.contains('Approval submitted').should('be.visible');
    
    // Voltar para usuário desenvolvedor
    cy.login('developer@innovabiz.com', 'Password123!');
    cy.visit('/mcp-dashboard');
    
    // Verificar que agora tem acesso
    cy.contains('button', 'Kubernetes Operations').should('not.be.disabled').click();
    cy.contains('h2', 'Kubernetes Operations').should('be.visible');
    
    // Verificar expiração visível
    cy.contains('Elevated access expires in').should('be.visible');
    cy.contains('k8s_operator').should('be.visible');
    
    // Verificar badge de elevação na UI
    cy.get('.elevation-badge').should('be.visible')
      .and('contain.text', 'ELEVATED');
  });
  
  it('deve registrar todas as ações realizadas com privilégios elevados', () => {
    // Setup: Garantir que já tem elevação (continuação do teste anterior)
    cy.contains('button', 'Kubernetes Operations').should('not.be.disabled').click();
    
    // Executar operação com privilégio elevado
    cy.contains('List Pods').click();
    cy.get('#namespace-select').select('qa');
    cy.contains('button', 'Apply').click();
    
    // Verificar logs de auditoria
    cy.visit('/audit-logs');
    cy.contains('Advanced Search').click();
    cy.get('#audit-user-filter').type('developer@innovabiz.com');
    cy.get('#audit-action-filter').select('Kubernetes Operations');
    cy.contains('button', 'Search').click();
    
    // Verificar que log contém informação de elevação
    cy.contains('tr', 'List Pods').within(() => {
      cy.contains('ELEVATED').should('be.visible');
      cy.contains('k8s_operator').should('be.visible');
    });
    
    // Verificar detalhes do log
    cy.contains('tr', 'List Pods').click();
    cy.get('.audit-detail-panel').within(() => {
      cy.contains('Elevation ID').should('be.visible');
      cy.contains('Approved by').should('contain.text', 'teamlead');
    });
  });
});
```

### 2.5. Testes de Segurança#### 2.5.1. Análise Estática

| Aspecto | Ferramentas | Conformidade | Critérios |
|---------|------------|--------------|-----------|
| Vulnerabilidades | SonarQube, GoSec | OWASP Top 10, CWE/SANS Top 25 | Zero vulnerabilidades críticas/altas |
| Qualidade de Código | SonarQube, golangci-lint | ISO/IEC 25010 | >85% nota de qualidade |
| Vazamentos de Segredos | Gitleaks, TruffleHog | NIST SP 800-53, PCI-DSS | Zero segredos expostos |
| Dependências | Nancy, Snyk | OWASP Top 10, CVE | Zero CVEs críticas/altas |
| Conformidade | Terrascan, Checkov | ISO 27001, SOC2, PCI-DSS | 100% de compliance |

#### 2.5.2. Análise Dinâmica

| Técnica | Ferramentas | Conformidade | Escopo |
|---------|------------|--------------|--------|
| DAST | OWASP ZAP, Burp Suite | OWASP ASVS 4.0 | APIs REST/GraphQL |
| API Fuzzing | API Fuzzer, RESTler | OWASP API Security Top 10 | Endpoints de autorização |
| Injeção de Políticas | OPA Fuzzer (custom) | NIST SP 800-53 AC-3 | Engine OPA |
| Runtime Protection | Falco, Tracee | CIS Benchmarks | Comportamento em execução |

#### 2.5.3. Testes de Penetração

| Cenário | Alvo | Técnicas | Conformidade |
|---------|------|----------|--------------|
| Bypass de Autorização | API de autorização | Manipulação de tokens/sessão | OWASP ASVS 4.0 |
| Elevação de Privilégios | Workflow de aprovação | Logic flaws, race conditions | ISO 27001 A.9 |
| Roubo de Tokens | Fluxos de autorização | MitM, XSS, token hijacking | NIST SP 800-63 |
| Falsificação de Contexto | Avaliação de risco | Spoofing de IP/dispositivo | ISO 42001 |

#### 2.5.4. Exemplo de Teste de Segurança (Script ZAP)

```python
#!/usr/bin/env python3
import time
import sys
from zapv2 import ZAPv2

# Configuração do ZAP
zap = ZAPv2(apikey='API_KEY', proxies={'http': 'http://localhost:8080', 'https': 'http://localhost:8080'})

target_url = 'https://iam-api.innovabiz.internal'
api_key = sys.argv[1] if len(sys.argv) > 1 else None

if not api_key:
    print('Necessário fornecer API key como argumento')
    sys.exit(1)

# Definir contexto e usuários para teste
print('Configurando contexto...')
context_id = zap.context.new_context('IAM_Authorization')
zap.context.include_in_context('IAM_Authorization', '^https://iam-api\\.innovabiz\\.internal.*$')

# Configurar autenticação
auth_method = 'scriptBasedAuthentication'
auth_params = {
    'scriptName': 'iam-auth.js',
    'scriptEngine': 'Oracle Nashorn',
    'scriptType': 'authentication'
}
zap.authentication.set_authentication_method(context_id, auth_method, auth_params)

# Configurar usuários
admin_id = zap.users.new_user(context_id, 'admin')
zap.users.set_authentication_credentials(context_id, admin_id, {
    'username': 'admin@innovabiz.com',
    'apikey': api_key
})
zap.users.set_user_enabled(context_id, admin_id, True)

basic_id = zap.users.new_user(context_id, 'basic')
zap.users.set_authentication_credentials(context_id, basic_id, {
    'username': 'basic@innovabiz.com',
    'apikey': api_key
})
zap.users.set_user_enabled(context_id, basic_id, True)

# Definir pontos de entrada
print('Adicionando pontos de entrada para autorização...')
zap.spider.scan_as_user(context_id, admin_id, target_url)
time.sleep(10)

# Configurar regras específicas de autorização
print('Configurando regras de varredura...')
zap.ascan.enable_scanners('40012')  # CSRF
zap.ascan.enable_scanners('40014')  # IDOR
zap.ascan.enable_scanners('40018')  # Path Traversal
zap.ascan.enable_scanners('90019')  # Server Side Include
zap.ascan.enable_scanners('90020')  # Remote File Include
zap.ascan.enable_scanners('90021')  # XPath Injection
zap.ascan.enable_scanners('40024')  # Logic flaws

# Executar scan com usuário Admin
print('Executando varredura com usuário Admin...')
scan_id = zap.ascan.scan_as_user(target_url + '/v1/auth/', context_id, admin_id, True, apikey=api_key)
time.sleep(5)

# Verificar progresso
while int(zap.ascan.status(scan_id)) < 100:
    print('Progresso da varredura: {}%'.format(zap.ascan.status(scan_id)))
    time.sleep(5)

# Executar testes específicos de autorização
print('Executando testes específicos de autorização...')

# Teste de elevação de privilégios
zap.replacer.add_rule('privilège_elevation', 'URL', target_url + '/v1/auth/elevation/request', 
                       '{"role":"developer"}', '{"role":"system_admin"}', 'REQ', False)

# Teste de bypass de aprovação
zap.replacer.add_rule('bypass_approval', 'URL', target_url + '/v1/auth/elevation/approve', 
                       '"approved":false', '"approved":true', 'REQ', False)

# Executar segundo scan com usuário básico e regras modificadas
print('Executando varredura com usuário Básico e regras modificadas...')
scan_id2 = zap.ascan.scan_as_user(target_url + '/v1/auth/', context_id, basic_id, True, apikey=api_key)
time.sleep(5)

# Verificar progresso
while int(zap.ascan.status(scan_id2)) < 100:
    print('Progresso da varredura: {}%'.format(zap.ascan.status(scan_id2)))
    time.sleep(5)

# Obter resultados
high_alerts = zap.core.alerts(target_url, 'High')
medium_alerts = zap.core.alerts(target_url, 'Medium')

# Gerar relatório
print('Gerando relatório...')
report_title = 'Relatório de Segurança - Autorização MCP-IAM'
report_template = 'traditional-xml'
report_file = 'authorization_security_report.xml'
zap.reports.generate(title=report_title, template=report_template, reportfilename=report_file)

print('Relatório gerado: {}'.format(report_file))
print('Alertas de alta severidade: {}'.format(len(high_alerts)))
print('Alertas de média severidade: {}'.format(len(medium_alerts)))

# Validar resultados contra critérios
if len(high_alerts) > 0:
    print('FALHA: Alertas de alta severidade detectados')
    sys.exit(1)
else:
    print('SUCESSO: Nenhum alerta de alta severidade detectado')
    sys.exit(0)
```

### 2.6. Testes de Performance

#### 2.6.1. Métricas e Objetivos

| Métrica | Descrição | Meta | Criticidade |
|---------|-----------|------|-------------|
| Latência Média | Tempo médio de decisão | < 50ms | Alta |
| Latência p95 | Tempo no percentil 95 | < 100ms | Alta |
| Latência p99 | Tempo no percentil 99 | < 200ms | Média |
| Throughput | Decisões por segundo | > 500 | Alta |
| Escalabilidade | Linearidade até 10x carga | < 20% degradação | Média |
| Cache Hit Rate | Taxa de acerto em cache | > 85% | Média |
| CPU Utilization | Uso de CPU sob carga | < 70% | Média |
| Memory Growth | Crescimento de memória sob carga | < 10% | Alta |

#### 2.6.2. Cenários de Carga

| Cenário | Descrição | Volume | Duração | Métricas Principais |
|---------|-----------|--------|---------|---------------------|
| Carga Base | Operações normais | 100 RPS | 30 min | Latência média, throughput |
| Pico | Pico de utilização | 500 RPS | 10 min | p95, CPU, erros |
| Sustentado | Carga constante alta | 300 RPS | 2 horas | Estabilidade, memória |
| Failover | Falha de nó | 200 RPS | 15 min | Tempo de recuperação |
| Cache Warming | População inicial de cache | 100 RPS | 5 min | Cache hit rate |

#### 2.6.3. Exemplo de Teste de Performance (K6)

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// Definir métricas customizadas
const authFailRate = new Rate('auth_failures');
const cacheHitRate = new Rate('cache_hit');

// Carregar dados de teste
const authRequests = new SharedArray('auth_requests', function() {
  return JSON.parse(open('./auth_requests.json'));
});

// Configuração de teste
export const options = {
  scenarios: {
    // Cenário de carga base
    base_load: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 100,
      stages: [
        { target: 100, duration: '5m' },  // Rampa até 100 RPS
        { target: 100, duration: '15m' }, // Manter 100 RPS
        { target: 20, duration: '5m' }    // Rampa para baixo
      ],
    },
    // Cenário de pico
    peak_load: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { target: 500, duration: '2m' },  // Rampa até 500 RPS
        { target: 500, duration: '5m' },  // Manter 500 RPS
        { target: 100, duration: '2m' }   // Rampa para baixo
      ],
      startTime: '30m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<100', 'p(99)<200', 'avg<50'],
    'http_req_duration{scenario:peak_load}': ['p(95)<150', 'p(99)<300', 'avg<70'],
    auth_failures: ['rate<0.01'],  // Menos de 1% de falhas
    cache_hit: ['rate>0.85'],      // Taxa de cache hit acima de 85%
  },
};

// Funções de ajuda
function getRandomAuthRequest() {
  return authRequests[Math.floor(Math.random() * authRequests.length)];
}

function getToken() {
  const response = http.post('https://iam-api.innovabiz.internal/v1/token', {
    username: 'perftest@innovabiz.com',
    password: 'Password123!'
  });
  return JSON.parse(response.body).token;
}

// Configuração inicial
let token = null;

export function setup() {
  token = getToken();
  return { token };
}

// Função principal de teste
export default function(data) {
  const authRequest = getRandomAuthRequest();
  
  // Adicionar token e cabeçalhos
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
    'X-Request-ID': `perf-test-${__VU}-${__ITER}`,
  };
  
  // Enviar solicitação de autorização
  const response = http.post(
    'https://iam-api.innovabiz.internal/v1/auth/decision',
    JSON.stringify(authRequest),
    { headers }
  );
  
  // Verificar resultado
  check(response, {
    'status é 200': (r) => r.status === 200,
    'resposta em formato correto': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.hasOwnProperty('allowed') && body.hasOwnProperty('context');
      } catch (e) {
        return false;
      }
    },
    'tempo de resposta < 200ms': (r) => r.timings.duration < 200,
  });
  
  // Métricas adicionais
  if (response.status !== 200) {
    authFailRate.add(1);
  } else {
    authFailRate.add(0);
  }
  
  // Verificar se resposta veio do cache
  if (response.headers['X-Cache-Hit']) {
    cacheHitRate.add(1);
  } else {
    cacheHitRate.add(0);
  }
  
  sleep(Math.random() * 0.5);
}
```

### 2.7. Testes de Conformidade

#### 2.7.1. Regulamentos e Normas Aplicáveis

| Regulamento/Norma | Aspectos Relevantes | Ferramentas/Técnicas | Periodicidade |
|-------------------|---------------------|----------------------|---------------|
| ISO 27001 A.9 | Controle de acesso | OPA Conftest, Auditor Interno | Trimestral |
| GDPR Art. 25, 32 | Privacy by Design, segurança | GDPR Checklist, análise de dados | Trimestral |
| PCI-DSS Req. 7, 8 | Acesso baseado em necessidade | PCI DSS Validator | Semestral |
| SOC 2 CC5.0 | Controle de acesso lógico | Auditor SOC 2 | Anual |
| NIST SP 800-53 AC | Controle de acesso e auditoria | OSCAL Validator | Trimestral |
| Open Banking | Gestão de consentimento | FAPI Conformance Suite | Trimestral |
| ISO 42001 | Gestão de IA | AI Risk Assessment | Trimestral |
| LGPD Art. 46 | Segurança de dados | Data Security Auditor | Trimestral |

#### 2.7.2. Validação por Princípios de Design

| Princípio | Requisito | Validação | Ferramenta |
|-----------|-----------|-----------|------------|
| Least Privilege | Mínimo acesso necessário | Verificação de permissões | OPA Policy Validator |
| Defense in Depth | Múltiplas camadas de controle | Análise de componentes | Architecture Validator |
| Separation of Duties | Funções críticas divididas | Análise de papéis | SoD Analyzer |
| Complete Mediation | Toda solicitação verificada | Testes de bypass | API Security Scanner |
| Zero Trust | Verificação contínua | Simulação de contexto | Context Simulator |
| Privacy by Design | Dados mínimos para decisão | Análise de dados | Data Flow Analyzer |
| Auditability | Registro completo de ações | Verificação de logs | Audit Completeness Check |
| Secure Defaults | Configuração segura inicial | Análise de padrões | Default Config Checker |

#### 2.7.3. Exemplo de Teste de Conformidade (OPA Conftest)

```hcl
# policy/authorization.rego
package authorization

# Política RBAC - ISO 27001 A.9.2.3 (Gestão de Direitos de Acesso Privilegiado)
deny[msg] {
  input.kind == "Role"
  not input.metadata.labels["audit-trail"]
  msg = "Roles devem ter label audit-trail para rastreabilidade (ISO 27001 A.9.2.3)"
}

deny[msg] {
  input.kind == "Role"
  contains(input.rules[_].resources, "secrets")
  not input.metadata.labels["data-classification"]
  msg = "Roles com acesso a secrets devem ter classificação de dados (GDPR Art. 32)"
}

# Política para Segregação de Deveres - ISO 27001 A.6.1.2
deny[msg] {
  input.kind == "RoleBinding"
  role_name := input.roleRef.name
  user_name := input.subjects[_].name
  is_admin_role(role_name)
  is_developer_role(user_binding_roles[user_name][_])
  msg = sprintf("Usuário %s possui papéis conflitantes (admin e developer) - violação de SoD (ISO 27001 A.6.1.2)", [user_name])
}

# Política para Aprovação Multi-nível - PCI-DSS Req 7.1.4
deny[msg] {
  input.kind == "Tool"
  input.metadata.annotations["security-impact"] == "high"
  not input.spec.requireApproval
  msg = "Ferramentas de alto impacto requerem aprovação multi-nível (PCI-DSS Req 7.1.4)"
}

# Política para Auditoria de Acesso - SOC 2 CC5.0
deny[msg] {
  input.kind == "AccessPolicy"
  not input.spec.auditLevel
  msg = "Políticas de acesso devem especificar nível de auditoria (SOC 2 CC5.0)"
}

# Verificação de Tempo Limitado - NIST SP 800-53 AC-2(3)
deny[msg] {
  input.kind == "ElevationRequest"
  not input.spec.validUntil
  msg = "Solicitações de elevação devem ter tempo limitado (NIST SP 800-53 AC-2(3))"
}

# Funções auxiliares
is_admin_role(role) {
  contains(role, "admin")
}

is_developer_role(role) {
  contains(role, "developer")
}

# Mapeamento usuário -> papéis
user_binding_roles[user_name] = roles {
  bindings := [b | input.items[i].kind == "RoleBinding"; b := input.items[i]]
  user_bindings := [b | b := bindings[_]; b.subjects[_].name == user_name]
  roles := [b.roleRef.name | b := user_bindings[_]]
}
```

## 3. Infraestrutura de Testes

### 3.1. Ambientes e Configurações

| Ambiente | Propósito | Infraestrutura | Configuração |
|----------|-----------|----------------|--------------|
| Local | Desenvolvimento, testes unitários | Docker Desktop | Mocks, containers locais |
| CI/CD | Pipeline automatizado | GitHub Actions, Jenkins | K8s em containers |
| Dev | Testes integrados | AKS/EKS/GKE cluster | Escala reduzida |
| QA | Testes E2E, segurança | AKS/EKS/GKE cluster | Similar à produção |
| Performance | Testes de carga | AKS/EKS/GKE dedicado | Escala completa |

### 3.2. Dados de Teste

| Tipo de Dados | Uso | Fonte | Sensibilidade |
|---------------|-----|-------|---------------|
| Usuários Sintéticos | Testes funcionais | Gerados automaticamente | Baixa |
| Permissões Simuladas | Testes de autorização | Baseados em produção, anonimizados | Média |
| Cenários de Risco | Testes adaptativos | Modelados de incidentes reais | Média |
| Padrões de Carga | Testes de performance | Baseados em telemetria | Baixa |
| Mocks de Serviços | Testes de integração | Contratos de API | Baixa |

### 3.3. Automação e Orquestração

| Aspecto | Ferramenta | Integração | Frequência |
|---------|------------|------------|------------|
| Pipeline CI | GitHub Actions | Gatilho por PR/commit | A cada commit |
| Build & Test | Jenkins | Integrado ao Git | Diária |
| Relatórios | Allure Report | Dashboard central | Após cada execução |
| Monitoramento | Grafana, Prometheus | Observabilidade | Contínuo |
| Gestão de Defeitos | Jira | Integrado a testes | Automático |

## 4. Gestão e Governança de Testes

### 4.1. Papéis e Responsabilidades

| Papel | Responsabilidades | Competências | Equipe |
|-------|------------------|--------------|--------|
| Gerente de Testes | Estratégia, planejamento, acompanhamento | Gestão de qualidade, processos | QA |
| Engenheiro de Testes | Implementação de testes automatizados | Automação, código, APIs | QA/Dev |
| Especialista em Segurança | Testes de segurança, análise de vulnerabilidades | Segurança ofensiva, ethical hacking | Segurança |
| Engenheiro de Performance | Testes de carga, análise de performance | Ferramentas de carga, análise de dados | Infraestrutura |
| Especialista em Compliance | Testes regulatórios | Normas, regulamentos | Compliance |

### 4.2. Métricas e KPIs

| Métrica | Descrição | Meta | Periodicidade |
|---------|-----------|------|---------------|
| Cobertura de Código | % de código coberto por testes | >90% | Diária |
| Cobertura Funcional | % de requisitos cobertos | 100% | Semanal |
| Taxa de Aprovação | % de testes bem-sucedidos | >98% | Diária |
| Tempo Médio de Execução | Duração de suítes de teste | <30 min | Diária |
| Tempo Médio de Correção | Tempo para corrigir falhas | <2 dias | Semanal |
| Vulnerabilidades Detectadas | Número por tipo e severidade | 0 críticas/altas | Semanal |
| Conformidade Regulatória | % de controles validados | 100% | Mensal |

### 4.3. Gestão de Riscos

| Risco | Impacto | Probabilidade | Mitigação | Responsável |
|-------|---------|---------------|-----------|------------|
| Falsos Positivos | Médio | Média | Refinar testes, revisões | Eng. de Testes |
| Degradação de Performance | Alto | Baixa | Testes de regressão, monitoramento | Eng. Performance |
| Falha de Cobertura | Alto | Baixa | Revisão de cobertura, análise de código | Gerente de Testes |
| Atraso em Resolução | Médio | Média | Priorização, escalação | Gerente de Testes |
| Mudança Regulatória | Alto | Média | Monitoramento, atualizações regulares | Esp. Compliance |

## 5. Cronograma e Priorização

### 5.1. Fases de Implementação

| Fase | Escopo | Duração | Dependências | Prioridade |
|------|--------|---------|--------------|------------|
| 1 - Fundação | Testes unitários, políticas OPA base | 3 semanas | Desenvolvimento inicial | Alta |
| 2 - Integração | Testes de integração, servidores MCP | 4 semanas | Fase 1 | Alta |
| 3 - E2E | Fluxos completos, testes de API | 3 semanas | Fase 2 | Média |
| 4 - Segurança | Testes de segurança, análise de vulnerabilidades | 2 semanas | Fase 3 | Alta |
| 5 - Performance | Testes de carga, otimização | 2 semanas | Fase 3 | Média |
| 6 - Compliance | Validação regulatória, auditoria | 3 semanas | Fase 4, 5 | Alta |

### 5.2. Priorização de Casos de Teste

Utilizamos a matriz RCRA (Risco, Criticidade, Regulatório, Adoção) para priorizar os casos de teste:

| Prioridade | Critérios | Exemplos |
|------------|-----------|----------|
| P0 (Bloqueador) | Segurança crítica, compliance obrigatório | Autorização para operações críticas, SoD |
| P1 (Alta) | Funcionalidades core, alto uso | Decisões de autorização base, elevação |
| P2 (Média) | Funcionalidades importantes, uso moderado | Delegação, avaliação de risco |
| P3 (Baixa) | Funcionalidades secundárias | Relatórios, configurações avançadas |

## 6. Referências

1. ISO/IEC 27001:2022 - Information Security Management
2. NIST SP 800-53 Rev. 5 - Security and Privacy Controls
3. PCI-DSS v4.0 - Payment Card Industry Data Security Standard
4. SOC 2 Type 2 - Service Organization Control Report
5. GDPR - General Data Protection Regulation
6. LGPD - Lei Geral de Proteção de Dados
7. ISO/IEC 42001 - Artificial Intelligence Management Systems
8. OWASP ASVS 4.0 - Application Security Verification Standard
9. OWASP API Security Top 10
10. CIS Benchmarks for Kubernetes
11. Financial-grade API Security Profile (FAPI)

---

**Autor(es):** Eduardo Jeremias  
**Data de Criação:** 2025-08-07  
**Última Atualização:** 2025-08-07  
**Status:** Proposta  
**Classificação:** Confidencial  
**Versão:** 1.0