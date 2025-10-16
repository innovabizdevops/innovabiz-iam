# INNOVABIZ - Plano de Testes Automatizados do Módulo Bureau de Crédito

## 1. Introdução

Este documento detalha a estratégia e os casos de teste automatizados para o módulo Bureau de Crédito (Central de Risco) da plataforma INNOVABIZ. O plano de testes foi desenvolvido considerando a arquitetura multi-mercado, multi-tenant, multi-camada e multi-contexto, garantindo a conformidade com regulamentações internacionais e a integração perfeita com outros módulos core da plataforma.

### 1.1 Objetivos

- Validar todas as funcionalidades do módulo Bureau de Crédito
- Garantir a conformidade com regulamentações específicas por mercado
- Verificar a integração adequada com outros módulos core
- Validar a implementação de observabilidade via MCP-IAM Observability
- Garantir a segurança e proteção de dados em todas as operações
- Verificar o desempenho e a escalabilidade do módulo

### 1.2 Escopo

- Testes unitários para componentes internos
- Testes de integração com outros módulos core
- Testes de compliance por mercado
- Testes de observabilidade e telemetria
- Testes de segurança e autorização
- Testes de desempenho e carga
- Testes de resiliência e recuperação

### 1.3 Ambiente de Testes

Os testes serão executados nos seguintes ambientes:

| Ambiente | Descrição | Finalidade |
|----------|-----------|------------|
| Desenvolvimento | Ambiente isolado para desenvolvedores | Testes unitários e desenvolvimento |
| Qualidade | Ambiente integrado com instâncias de teste de outros módulos | Testes de integração e compliance |
| Homologação | Réplica de produção com dados sanitizados | Testes de carga e desempenho |
| Sandbox | Ambiente para simulação de cenários específicos | Testes de casos extremos e recuperação |

## 2. Testes Unitários

### 2.1 Framework e Configuração

```go
// Configuração para testes unitários
package bureau_test

import (
    "testing"
    "context"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/innovabiz/iam/modules/bureau"
    "github.com/innovabiz/iam/observability/adapter"
)

// Mock do adaptador de observabilidade
type MockObservabilityAdapter struct {
    mock.Mock
}

// Implementar todos os métodos necessários do adaptador
func (m *MockObservabilityAdapter) Tracer() trace.Tracer {
    args := m.Called()
    return args.Get(0).(trace.Tracer)
}

// ... outros métodos conforme interface adapter.IAMObservability
```

### 2.2 Casos de Teste Unitários

#### 2.2.1 Validação de Autenticação e Autorização

```go
func TestVerificarAutenticacaoAutorizacao(t *testing.T) {
    // Configurar mock e bureau
    mockObs := new(MockObservabilityAdapter)
    bc := bureau.NewBureauCredito(mockObs)
    
    // Caso: MFA insuficiente para consulta completa
    t.Run("MFA_Insuficiente", func(t *testing.T) {
        // Configurar mock para retornar falha de MFA
        mockObs.On("ValidateMFA", mock.Anything, mock.Anything, mock.Anything, "low").
            Return(false, fmt.Errorf("nível MFA insuficiente"))
            
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-1",
            TipoConsulta:     bureau.ConsultaCompleta,
            MFALevel:         "low",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        err := bc.verificarAutenticacao(context.Background(), &consulta)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "MFA insuficiente")
        mockObs.AssertExpectations(t)
    })
    
    // Caso: Escopo inadequado para o tipo de consulta
    t.Run("Escopo_Inadequado", func(t *testing.T) {
        // Configurar mock para retornar sucesso em MFA mas falha em escopo
        mockObs.On("ValidateMFA", mock.Anything, mock.Anything, mock.Anything, "high").
            Return(true, nil)
        mockObs.On("ValidateScope", mock.Anything, mock.Anything, mock.Anything, 
            "bureau_credito:consulta:completa").
            Return(false, fmt.Errorf("escopo insuficiente"))
            
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-2",
            TipoConsulta:     bureau.ConsultaCompleta,
            MFALevel:         "high",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        err := bc.verificarAutenticacao(context.Background(), &consulta)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "escopo insuficiente")
        mockObs.AssertExpectations(t)
    })
    
    // Caso: Autenticação e autorização com sucesso
    t.Run("Autenticacao_Sucesso", func(t *testing.T) {
        // Configurar mock para retornar sucesso em ambos
        mockObs.On("ValidateMFA", mock.Anything, mock.Anything, mock.Anything, "high").
            Return(true, nil)
        mockObs.On("ValidateScope", mock.Anything, mock.Anything, mock.Anything, 
            "bureau_credito:consulta:completa").
            Return(true, nil)
            
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-3",
            TipoConsulta:     bureau.ConsultaCompleta,
            MFALevel:         "high",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        err := bc.verificarAutenticacao(context.Background(), &consulta)
        
        assert.NoError(t, err)
        mockObs.AssertExpectations(t)
    })
}
```

#### 2.2.2 Validação de Consentimento

```go
func TestVerificarConsentimento(t *testing.T) {
    // Configurar mock e bureau
    mockObs := new(MockObservabilityAdapter)
    bc := bureau.NewBureauCredito(mockObs)
    
    // Configurar mercados que exigem consentimento
    bc.ConsentimentoObrigatorio = map[string]bool{
        "angola": true,
        "brazil": true,
        "eu": true,
        "usa": false,
    }
    
    // Caso: Mercado exige consentimento, mas ID não fornecido
    t.Run("Consentimento_Ausente", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-1",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
            ConsentimentoID:  "", // Ausente
        }
        
        err := bc.verificarConsentimento(context.Background(), &consulta)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "consentimento obrigatório")
    })
    
    // Caso: Mercado exige consentimento e valida com sucesso
    t.Run("Consentimento_Valido", func(t *testing.T) {
        mockObs.On("ValidateConsent", mock.Anything, mock.Anything, "12345678901", "consent-123").
            Return(true, nil)
            
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-2",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
            ConsentimentoID:  "consent-123",
        }
        
        err := bc.verificarConsentimento(context.Background(), &consulta)
        
        assert.NoError(t, err)
        mockObs.AssertExpectations(t)
    })
    
    // Caso: Mercado não exige consentimento
    t.Run("Consentimento_NaoExigido", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-3",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "usa"},
            ConsentimentoID:  "", // Ausente, mas não exigido
        }
        
        err := bc.verificarConsentimento(context.Background(), &consulta)
        
        assert.NoError(t, err)
    })
}
```

#### 2.2.3 Validação de Regras de Compliance

```go
func TestValidarRegrasCompliance(t *testing.T) {
    // Configurar bureau
    mockObs := new(MockObservabilityAdapter)
    bc := bureau.NewBureauCredito(mockObs)
    
    // Registrar regras de compliance para teste
    bc.RegistrarRegraCompliance(bureau.RegrasCompliance{
        ID:           "TEST-RULE-1",
        Market:       "global",
        Description:  "Regra de teste global",
        Framework:    []string{"ISO 27001"},
        MandatoryFor: []string{string(bureau.ConsultaCompleta)},
        Validate: func(consulta *bureau.ConsultaCredito) (bool, string, error) {
            // Regra simples: documento deve ter pelo menos 5 caracteres
            if len(consulta.DocumentoCliente) < 5 {
                return false, "documento inválido", nil
            }
            return true, "", nil
        },
    })
    
    bc.RegistrarRegraCompliance(bureau.RegrasCompliance{
        ID:           "TEST-RULE-2",
        Market:       "brazil",
        Description:  "Regra de teste Brasil",
        Framework:    []string{"LGPD"},
        MandatoryFor: []string{string(bureau.ConsultaCompleta), string(bureau.ConsultaHistorico)},
        Validate: func(consulta *bureau.ConsultaCredito) (bool, string, error) {
            // Regra simples: finalidade deve ser especificada
            if consulta.Finalidade == "" {
                return false, "finalidade não especificada", nil
            }
            return true, "", nil
        },
    })
    
    // Caso: Regra global falha
    t.Run("Falha_Regra_Global", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-1",
            TipoConsulta:     bureau.ConsultaCompleta,
            DocumentoCliente: "1234", // Menos de 5 caracteres
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        valido, mensagens := bc.validarRegrasCompliance(context.Background(), &consulta)
        
        assert.False(t, valido)
        assert.Contains(t, mensagens, "documento inválido")
    })
    
    // Caso: Regra específica de mercado falha
    t.Run("Falha_Regra_Mercado", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-2",
            TipoConsulta:     bureau.ConsultaHistorico,
            DocumentoCliente: "12345", // Válido para regra global
            Finalidade:       "", // Inválido para regra do Brasil
            MarketContext:    adapter.MarketContext{Market: "brazil"},
        }
        
        valido, mensagens := bc.validarRegrasCompliance(context.Background(), &consulta)
        
        assert.False(t, valido)
        assert.Contains(t, mensagens, "finalidade não especificada")
    })
    
    // Caso: Todas as regras passam
    t.Run("Regras_Validas", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-3",
            TipoConsulta:     bureau.ConsultaCompleta,
            DocumentoCliente: "12345", // Válido para regra global
            Finalidade:       "FinalidadeConcessaoCredito", // Válido para regra do Brasil
            MarketContext:    adapter.MarketContext{Market: "brazil"},
        }
        
        valido, mensagens := bc.validarRegrasCompliance(context.Background(), &consulta)
        
        assert.True(t, valido)
        assert.Empty(t, mensagens)
    })
}
```#### 2.2.4 Verificação de Limites Diários

```go
func TestVerificarLimiteConsultas(t *testing.T) {
    // Configurar bureau
    mockObs := new(MockObservabilityAdapter)
    bc := bureau.NewBureauCredito(mockObs)
    
    // Configurar limites para teste
    bc.LimitesDiarios = map[string]int{
        "default": 100,
        "angola":  50,
        "brazil":  200,
    }
    
    // Inicializar contadores para simulação
    bc.Contadores = map[string]int{
        "entidade-1": 48,
        "entidade-2": 100,
    }
    bc.mu = &sync.Mutex{} // Inicializar mutex para testes
    
    // Caso: Dentro do limite
    t.Run("Dentro_Limite", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-1",
            EntidadeID:       "entidade-1",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        err := bc.verificarLimiteConsultas(context.Background(), &consulta)
        
        assert.NoError(t, err)
        assert.Equal(t, 49, bc.Contadores["entidade-1"], "contador deve ser incrementado")
    })
    
    // Caso: Limite excedido
    t.Run("Limite_Excedido", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-2",
            EntidadeID:       "entidade-1",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        // Primeiro incremento (já em 49)
        err := bc.verificarLimiteConsultas(context.Background(), &consulta)
        assert.NoError(t, err)
        assert.Equal(t, 50, bc.Contadores["entidade-1"])
        
        // Segundo incremento (deve exceder)
        err = bc.verificarLimiteConsultas(context.Background(), &consulta)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "limite diário de consultas excedido")
        assert.Equal(t, 50, bc.Contadores["entidade-1"], "contador não deve incrementar ao exceder")
    })
    
    // Caso: Limite para mercado sem configuração específica
    t.Run("Limite_Default", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-3",
            EntidadeID:       "entidade-3", // Nova entidade
            MarketContext:    adapter.MarketContext{Market: "global"}, // Sem config específica
        }
        
        err := bc.verificarLimiteConsultas(context.Background(), &consulta)
        
        assert.NoError(t, err)
        assert.Equal(t, 1, bc.Contadores["entidade-3"], "contador deve inicializar e incrementar")
    })
}
```

#### 2.2.5 Processamento de Notificações Regulatórias

```go
func TestProcessarNotificacoesRegulatorias(t *testing.T) {
    // Configurar mock e bureau
    mockObs := new(MockObservabilityAdapter)
    mockObs.On("TraceAuditEvent", mock.Anything, mock.Anything, mock.Anything, 
        mock.MatchedBy(func(s string) bool { return strings.Contains(s, "notificacao") }), 
        mock.Anything).Return()
    
    bc := bureau.NewBureauCredito(mockObs)
    
    // Configurar notificações obrigatórias por mercado
    bc.NotificacaoObrigatoria = map[string]bool{
        "angola": true,
        "brazil": false, // Apenas para restrições
        "eu":     true,
    }
    
    // Caso: Notificação obrigatória para Angola
    t.Run("Notificacao_Angola", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-1",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "angola"},
        }
        
        resultado := bureau.ResultadoConsulta{
            ConsultaID:     "test-1",
            RestricoesList: []bureau.RegistroCredito{}, // Sem restrições
        }
        
        bc.processarNotificacoesRegulatorias(context.Background(), &consulta, &resultado)
        
        // Verificar se o método de auditoria foi chamado pelo menos uma vez
        mockObs.AssertNumberOfCalls(t, "TraceAuditEvent", 1)
    })
    
    // Caso: Notificação para Brasil apenas com restrições
    t.Run("Notificacao_Brasil_Com_Restricoes", func(t *testing.T) {
        mockObs.ExpectedCalls = nil // Limpar chamadas esperadas
        mockObs.On("TraceAuditEvent", mock.Anything, mock.Anything, mock.Anything, 
            mock.MatchedBy(func(s string) bool { return strings.Contains(s, "notificacao") }), 
            mock.Anything).Return()
        
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-2",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
        }
        
        resultado := bureau.ResultadoConsulta{
            ConsultaID: "test-2",
            RestricoesList: []bureau.RegistroCredito{
                {Descricao: "Restrição de teste"},
            },
        }
        
        bc.processarNotificacoesRegulatorias(context.Background(), &consulta, &resultado)
        
        // Verificar se o método de auditoria foi chamado pelo menos uma vez
        mockObs.AssertNumberOfCalls(t, "TraceAuditEvent", 1)
    })
    
    // Caso: Sem notificação para Brasil sem restrições
    t.Run("Sem_Notificacao_Brasil_Sem_Restricoes", func(t *testing.T) {
        mockObs.ExpectedCalls = nil // Limpar chamadas esperadas
        mockObs.On("TraceAuditEvent", mock.Anything, mock.Anything, mock.Anything, 
            mock.MatchedBy(func(s string) bool { return strings.Contains(s, "notificacao") }), 
            mock.Anything).Return()
        
        consulta := bureau.ConsultaCredito{
            ConsultaID:       "test-3",
            DocumentoCliente: "12345678901",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
        }
        
        resultado := bureau.ResultadoConsulta{
            ConsultaID:     "test-3",
            RestricoesList: []bureau.RegistroCredito{}, // Sem restrições
        }
        
        bc.processarNotificacoesRegulatorias(context.Background(), &consulta, &resultado)
        
        // Verificar que o método de auditoria não foi chamado
        mockObs.AssertNumberOfCalls(t, "TraceAuditEvent", 0)
    })
}
```

### 2.3 Execução dos Testes Unitários

Os testes unitários devem ser executados como parte do processo de build e integração contínua:

```bash
cd CoreModules/IAM
go test -v -cover ./modules/bureau/...
```

Requisitos de cobertura:
- Cobertura mínima: 85% para funções críticas
- Cobertura ideal: 95% para todo o módulo

## 3. Testes de Integração

### 3.1 Framework e Configuração

Os testes de integração utilizam o framework Testcontainers com Docker para simular o ambiente completo:

```go
package integration_test

import (
    "context"
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/innovabiz/iam/modules/bureau"
    "github.com/innovabiz/iam/observability/adapter"
)

// Configuração do ambiente de teste
type TestEnvironment struct {
    IAMContainer          testcontainers.Container
    BureauCreditoContainer testcontainers.Container
    OTelCollectorContainer testcontainers.Container
    IAMClient             *iam.Client
    BureauCreditoClient   *bureau.Client
    CleanupFunc           func()
}

// Inicializar ambiente de teste
func setupTestEnvironment(t *testing.T) *TestEnvironment {
    ctx := context.Background()
    
    // Inicializar contêiner OpenTelemetry Collector
    otelContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "otel/opentelemetry-collector:0.92.0",
            ExposedPorts: []string{"4317:4317", "4318:4318"},
            WaitingFor:   wait.ForHTTP("/").WithPort("4318"),
        },
        Started: true,
    })
    require.NoError(t, err)
    
    // Inicializar contêiner IAM
    iamContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "innovabiz/iam-service:latest",
            ExposedPorts: []string{"8080:8080"},
            Env: map[string]string{
                "ENVIRONMENT":              "test",
                "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
            },
            WaitingFor: wait.ForHTTP("/health").WithPort("8080"),
        },
        Started: true,
    })
    require.NoError(t, err)
    
    // Inicializar contêiner Bureau de Crédito
    bureauContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "innovabiz/bureau-credito:latest",
            ExposedPorts: []string{"8080:8080"},
            Env: map[string]string{
                "ENVIRONMENT":              "test",
                "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
            },
            WaitingFor: wait.ForHTTP("/health").WithPort("8080"),
        },
        Started: true,
    })
    require.NoError(t, err)
    
    // Configurar clientes para os serviços
    iamHost, _ := iamContainer.Host(ctx)
    iamPort, _ := iamContainer.MappedPort(ctx, "8080")
    iamClient, err := iam.NewClient(fmt.Sprintf("http://%s:%s", iamHost, iamPort))
    require.NoError(t, err)
    
    bureauHost, _ := bureauContainer.Host(ctx)
    bureauPort, _ := bureauContainer.MappedPort(ctx, "8080")
    bureauClient, err := bureau.NewClient(fmt.Sprintf("http://%s:%s", bureauHost, bureauPort))
    require.NoError(t, err)
    
    // Retornar ambiente configurado
    return &TestEnvironment{
        IAMContainer:          iamContainer,
        BureauCreditoContainer: bureauContainer,
        OTelCollectorContainer: otelContainer,
        IAMClient:             iamClient,
        BureauCreditoClient:   bureauClient,
        CleanupFunc: func() {
            bureauContainer.Terminate(ctx)
            iamContainer.Terminate(ctx)
            otelContainer.Terminate(ctx)
        },
    }
}
```

### 3.2 Casos de Teste de Integração

#### 3.2.1 Fluxo Completo de Consulta

```go
func TestFluxoCompletoConsulta(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso válido do IAM
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    // Criar contexto com token
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Caso: Consulta Completa com Sucesso
    t.Run("Consulta_Completa_Sucesso", func(t *testing.T) {
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            Finalidade:       bureau.FinalidadeConcessaoCredito,
            EntidadeID:       "entidade-teste",
            TipoEntidade:     "PF",
            DocumentoCliente: "12345678901",
            NomeCliente:      "Cliente Teste",
            UsuarioID:        "admin",
            ConsentimentoID:  "consent-123",
            SolicitanteID:    "test-integration",
            MarketContext:    adapter.MarketContext{
                Market:     "brazil",
                TenantID:   "tenant-teste",
                TenantType: "default",
            },
            MFALevel: "high",
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        // Validações
        assert.NoError(t, err)
        assert.NotNil(t, resultado)
        assert.Equal(t, consulta.ConsultaID, resultado.ConsultaID)
        assert.NotNil(t, resultado.ScoreCredito)
    })
    
    // Caso: Consulta com Erro de Autorização
    t.Run("Consulta_Erro_Autorizacao", func(t *testing.T) {
        // Criar token com escopos insuficientes
        tokenLimitado, err := env.IAMClient.LoginWithScope(context.Background(), 
            "user", "password", []string{"bureau_credito:consulta:basica"})
        require.NoError(t, err)
        
        ctxLimitado := context.WithValue(context.Background(), "Authorization", "Bearer "+tokenLimitado)
        
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta, // Tenta consulta completa com escopo básico
            Finalidade:       bureau.FinalidadeConcessaoCredito,
            EntidadeID:       "entidade-teste",
            DocumentoCliente: "12345678901",
            UsuarioID:        "user",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
            MFALevel:         "standard",
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctxLimitado, consulta)
        
        // Validações
        assert.Error(t, err)
        assert.Nil(t, resultado)
        assert.Contains(t, err.Error(), "autorização")
    })
}
```

#### 3.2.2 Integração com Observabilidade

```go
func TestIntegracaoObservabilidade(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso válido do IAM
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    // Criar contexto com token
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Gerar ID único para rastreamento
    traceID := fmt.Sprintf("trace-%s", uuid.New().String())
    ctx = context.WithValue(ctx, "trace-id", traceID)
    
    // Realizar uma consulta para gerar telemetria
    consulta := bureau.ConsultaCredito{
        ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
        TipoConsulta:     bureau.ConsultaScore,
        Finalidade:       bureau.FinalidadeVerificacaoCliente,
        EntidadeID:       "entidade-teste",
        DocumentoCliente: "12345678901",
        UsuarioID:        "admin",
        MarketContext:    adapter.MarketContext{Market: "angola"},
        MFALevel:         "standard",
    }
    
    // Realizar consulta
    _, err = env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
    require.NoError(t, err)
    
    // Aguardar propagação da telemetria
    time.Sleep(2 * time.Second)
    
    // Verificar spans no coletor OpenTelemetry
    otelHost, _ := env.OTelCollectorContainer.Host(context.Background())
    otelPort, _ := env.OTelCollectorContainer.MappedPort(context.Background(), "4318")
    
    resp, err := http.Get(fmt.Sprintf("http://%s:%s/v1/traces?service.name=bureau-credito&trace-id=%s", 
        otelHost, otelPort, traceID))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Validar resposta
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    body, err := ioutil.ReadAll(resp.Body)
    require.NoError(t, err)
    
    // Verificar presença de spans
    var traces map[string]interface{}
    err = json.Unmarshal(body, &traces)
    require.NoError(t, err)
    
    // Validar presença de spans específicos
    assert.NotEmpty(t, traces["data"])
    
    // Verificar métricas no Prometheus
    // ... código similar para verificar métricas ...
}
```

### 3.3 Testes de Integração por Mercado

#### 3.3.1 Testes de Compliance Angola (BNA)

```go
func TestComplianceAngola(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Caso: Consentimento Obrigatório Angola
    t.Run("Consentimento_Obrigatorio_Angola", func(t *testing.T) {
        // Consulta sem consentimento
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            Finalidade:       bureau.FinalidadeConcessaoCredito,
            EntidadeID:       "entidade-angola",
            DocumentoCliente: "AO12345678901",
            UsuarioID:        "admin",
            MarketContext:    adapter.MarketContext{Market: "angola"},
            MFALevel:         "high",
            // ConsentimentoID intencionalmente omitido
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        // Validações
        assert.Error(t, err)
        assert.Nil(t, resultado)
        assert.Contains(t, err.Error(), "consentimento obrigatório")
    })
    
    // Caso: MFA Alto Obrigatório Angola
    t.Run("MFA_Alto_Obrigatorio_Angola", func(t *testing.T) {
        // Consulta com MFA baixo
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            Finalidade:       bureau.FinalidadeConcessaoCredito,
            EntidadeID:       "entidade-angola",
            DocumentoCliente: "AO12345678901",
            UsuarioID:        "admin",
            ConsentimentoID:  "consent-123",
            MarketContext:    adapter.MarketContext{Market: "angola"},
            MFALevel:         "low", // MFA baixo
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        // Validações
        assert.Error(t, err)
        assert.Nil(t, resultado)
        assert.Contains(t, err.Error(), "MFA")
    })
    
    // Caso: Notificação Obrigatória Angola
    t.Run("Notificacao_Obrigatoria_Angola", func(t *testing.T) {
        // TODO: Implementar teste para verificar notificações BNA
        // Este teste requer um mock do sistema de notificações ou
        // acesso a logs/eventos para verificar que a notificação foi enviada
    })
}
```

#### 3.3.2 Testes de Compliance Brasil (BACEN/LGPD)

```go
func TestComplianceBrasil(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Caso: Finalidade Obrigatória LGPD
    t.Run("Finalidade_Obrigatoria_LGPD", func(t *testing.T) {
        // Consulta sem finalidade
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            EntidadeID:       "entidade-brasil",
            DocumentoCliente: "12345678901",
            UsuarioID:        "admin",
            ConsentimentoID:  "consent-123",
            MarketContext:    adapter.MarketContext{Market: "brazil"},
            MFALevel:         "high",
            // Finalidade intencionalmente omitida
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        // Validações
        assert.Error(t, err)
        assert.Nil(t, resultado)
        assert.Contains(t, err.Error(), "finalidade")
    })
    
    // Caso: Notificação Restrições BACEN
    t.Run("Notificacao_Restricoes_BACEN", func(t *testing.T) {
        // TODO: Implementar teste para verificar notificações BACEN
        // Esse teste requer um mock que força o resultado a conter restrições
        // e verifica se a notificação foi registrada adequadamente
    })
}
```#### 3.3.3 Testes de Compliance União Europeia (GDPR/PSD2)

```go
func TestComplianceEU(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Caso: Minimização de Dados GDPR
    t.Run("Minimizacao_Dados_GDPR", func(t *testing.T) {
        // Consulta com finalidade que requer minimização
        consultaMinimizada := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaBasica,
            Finalidade:       bureau.FinalidadeVerificacaoCliente,
            EntidadeID:       "entidade-eu",
            DocumentoCliente: "EU12345678901",
            UsuarioID:        "admin",
            ConsentimentoID:  "consent-123",
            MarketContext:    adapter.MarketContext{Market: "eu"},
            MFALevel:         "standard",
        }
        
        // Realizar consulta minimizada
        resultadoMinimizado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consultaMinimizada)
        require.NoError(t, err)
        
        // Consulta completa (que não deveria ter todos dados para esta finalidade)
        consultaCompleta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            Finalidade:       bureau.FinalidadeVerificacaoCliente, // Mesma finalidade
            EntidadeID:       "entidade-eu",
            DocumentoCliente: "EU12345678901",
            UsuarioID:        "admin",
            ConsentimentoID:  "consent-456",
            MarketContext:    adapter.MarketContext{Market: "eu"},
            MFALevel:         "high",
        }
        
        // Realizar consulta completa
        resultadoCompleto, err := env.BureauCreditoClient.RealizarConsulta(ctx, consultaCompleta)
        require.NoError(t, err)
        
        // Validações de minimização de dados
        assert.NotNil(t, resultadoMinimizado.ScoreCredito)
        assert.Empty(t, resultadoMinimizado.RegistrosCredito, "Não deve retornar registros para consulta básica")
        
        // Verificar campos sensíveis no resultado completo
        for _, registro := range resultadoCompleto.RegistrosCredito {
            if valor, ok := registro.Metadata["cpf_completo"]; ok {
                assert.Fail(t, "Campo sensível 'cpf_completo' não deveria estar presente", 
                    "valor encontrado: %v", valor)
            }
        }
    })
    
    // Caso: Direito ao Esquecimento GDPR
    t.Run("Direito_Esquecimento_GDPR", func(t *testing.T) {
        // TODO: Implementar teste para o direito ao esquecimento
        // Requer implementação da função de exclusão de dados
    })
}
```

#### 3.3.4 Testes de Compliance EUA (FCRA/GLBA)

```go
func TestComplianceUSA(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Caso: Finalidade Permissível FCRA
    t.Run("Finalidade_Permissivel_FCRA", func(t *testing.T) {
        // Consulta com finalidade não permissível
        consulta := bureau.ConsultaCredito{
            ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
            TipoConsulta:     bureau.ConsultaCompleta,
            Finalidade:       "FinalidadeNaoPermissivel", // Finalidade não reconhecida
            EntidadeID:       "entidade-usa",
            DocumentoCliente: "US12345678901",
            UsuarioID:        "admin",
            MarketContext:    adapter.MarketContext{Market: "usa"},
            MFALevel:         "standard",
        }
        
        // Realizar consulta
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        // Validações
        assert.Error(t, err)
        assert.Nil(t, resultado)
        assert.Contains(t, err.Error(), "finalidade não permissível")
    })
    
    // Caso: Notificação de Decisão Negativa FCRA
    t.Run("Notificacao_Decisao_Negativa_FCRA", func(t *testing.T) {
        // TODO: Implementar teste para verificar notificações FCRA
        // Requer mock para forçar decisão negativa e verificar notificação
    })
}
```

### 3.4 Testes de Integração com Módulos Core

#### 3.4.1 Integração com Payment Gateway

```go
func TestIntegracaoPaymentGateway(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Configurar cliente Payment Gateway
    pgClient := setupPaymentGatewayClient(t)
    
    // Caso: Verificação de Score para Autorização
    t.Run("Verificacao_Score_Autorizacao", func(t *testing.T) {
        // Criar transação de teste
        transacao := payment.Transaction{
            TransactionID:     fmt.Sprintf("pg-%s", uuid.New().String()),
            Amount:            1000.0,
            Currency:          "BRL",
            CustomerID:        "customer-test",
            CustomerDocument:  "12345678901",
            CustomerType:      "PF",
            MerchantID:        "merchant-test",
            MarketContext:     adapter.MarketContext{Market: "brazil"},
        }
        
        // Processar transação (internamente consulta Bureau de Crédito)
        resultado, err := pgClient.ProcessTransaction(context.Background(), transacao)
        require.NoError(t, err)
        
        // Validações
        assert.NotNil(t, resultado)
        assert.Contains(t, resultado.Metadata, "bureau_score")
    })
}
```

#### 3.4.2 Integração com Risk Management

```go
func TestIntegracaoRiskManagement(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Configurar cliente Risk Management
    rmClient := setupRiskManagementClient(t)
    
    // Caso: Avaliação de Risco Completa
    t.Run("Avaliacao_Risco_Completa", func(t *testing.T) {
        // Solicitar avaliação de risco
        avaliacao, err := rmClient.AvaliarRiscoCliente(context.Background(), 
            "cliente-test", "12345678901")
        require.NoError(t, err)
        
        // Validações
        assert.NotNil(t, avaliacao)
        assert.Contains(t, avaliacao.FontesDados, "bureau_credito")
        assert.NotNil(t, avaliacao.ScoreCredito)
    })
}
```

#### 3.4.3 Integração com Mobile Money

```go
func TestIntegracaoMobileMoney(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Configurar cliente Mobile Money
    mmClient := setupMobileMoneyClient(t)
    
    // Caso: Microcrédito com Análise de Bureau
    t.Run("Microcredito_Analise_Bureau", func(t *testing.T) {
        // Solicitar análise para microcrédito
        decisao, err := mmClient.AvaliarMicrocredito(context.Background(), 
            "cliente-test", "12345678901", 500.0)
        require.NoError(t, err)
        
        // Validações
        assert.NotNil(t, decisao)
        assert.Contains(t, []string{"APROVADO", "APROVADO_PARCIAL", "NEGADO"}, 
            decisao.StatusAprovacao)
    })
}
```

## 4. Testes de Desempenho e Carga

### 4.1 Configuração do Ambiente de Testes

Para os testes de desempenho e carga, utilizamos o framework k6 com a seguinte configuração:

```javascript
// k6-bureau-credito-performance.js
import http from 'k6/http';
import { check, sleep } from 'k6';

// Parâmetros de teste configuráveis
export let options = {
    stages: [
        { duration: '30s', target: 10 },  // Subida gradual para 10 usuários
        { duration: '1m', target: 50 },   // Subida para 50 usuários
        { duration: '3m', target: 50 },   // Manutenção em 50 usuários
        { duration: '30s', target: 0 },   // Descida gradual para 0
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],  // 95% das requisições devem ser abaixo de 500ms
        http_req_failed: ['rate<0.01'],    // Menos de 1% de falhas
    },
};

// Token JWT para autenticação (renovado periodicamente)
let token;

// Função para obter token
function getToken() {
    const loginRes = http.post('http://iam-service:8080/auth/login', 
        JSON.stringify({ username: 'perftest', password: 'perftest123' }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    
    check(loginRes, {
        'login successful': (r) => r.status === 200,
    });
    
    return loginRes.json('token');
}

// Configuração inicial
export function setup() {
    token = getToken();
    return { token };
}

// Teste principal
export default function(data) {
    // Renovar token se necessário
    if (Math.random() < 0.1) {  // 10% de chance de renovar o token
        token = getToken();
    }
    
    // Gerar dados aleatórios para consulta
    const documentoCliente = `${Math.floor(10000000000 + Math.random() * 90000000000)}`;
    const entidadeID = `entity-${Math.floor(Math.random() * 1000)}`;
    const tiposConsulta = ['ConsultaBasica', 'ConsultaScore', 'ConsultaRestricoes'];
    const tipoConsulta = tiposConsulta[Math.floor(Math.random() * tiposConsulta.length)];
    const mercados = ['angola', 'brazil', 'eu', 'usa', 'global'];
    const mercado = mercados[Math.floor(Math.random() * mercados.length)];
    
    // Configurar payload da consulta
    const payload = JSON.stringify({
        tipoConsulta: tipoConsulta,
        finalidade: 'FinalidadeVerificacaoCliente',
        entidadeID: entidadeID,
        documentoCliente: documentoCliente,
        usuarioID: 'perftest',
        marketContext: {
            market: mercado,
            tenantID: 'tenant-test',
            tenantType: 'default'
        },
        mfaLevel: 'standard',
        consentimentoID: `consent-${documentoCliente}`
    });
    
    // Realizar requisição de consulta
    const res = http.post('http://bureau-credito-service:8080/consultas', payload, {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
    });
    
    // Verificar resultado
    check(res, {
        'status is 200': (r) => r.status === 200,
        'consulta processada': (r) => r.json('consultaID') !== undefined,
    });
    
    // Pausa entre requisições
    sleep(Math.random() * 3 + 1);  // 1-4 segundos
}

// Limpeza após o teste
export function teardown(data) {
    // Nada a limpar
}
```

### 4.2 Casos de Teste de Desempenho

#### 4.2.1 Teste de Carga Padrão

Execução do teste de carga com carga moderada para validar desempenho em condições normais:

```bash
k6 run --env ENVIRONMENT=quality k6-bureau-credito-performance.js
```

Critérios de aceitação:
- Latência média < 200ms
- Latência P95 < 500ms
- Taxa de erro < 1%

#### 4.2.2 Teste de Pico

Execução do teste com picos de carga para validar comportamento em momentos de alta demanda:

```javascript
// k6-bureau-credito-peak.js (modificação do script base)
export let options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '30s', target: 100 },  // Pico para 100 usuários
        { duration: '1m', target: 100 },   // Manutenção do pico
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<1000'],  // 95% abaixo de 1s durante pico
        http_req_failed: ['rate<0.05'],     // Menos de 5% de falhas
    },
};
```

```bash
k6 run --env ENVIRONMENT=homologation k6-bureau-credito-peak.js
```

Critérios de aceitação:
- Sistema deve suportar o pico sem falhas de serviço
- Recuperação completa após o pico

#### 4.2.3 Teste de Resistência

Execução do teste com carga constante por período prolongado:

```javascript
// k6-bureau-credito-endurance.js
export let options = {
    stages: [
        { duration: '2m', target: 30 },   // Subida para 30 usuários
        { duration: '30m', target: 30 },  // Manutenção por 30 minutos
        { duration: '2m', target: 0 },    // Descida gradual
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};
```

```bash
k6 run --env ENVIRONMENT=homologation k6-bureau-credito-endurance.js
```

Critérios de aceitação:
- Sem degradação de desempenho ao longo do tempo
- Uso de memória estável (sem vazamentos)

## 5. Testes de Resiliência e Recuperação

### 5.1 Injeção de Falhas com Chaos Engineering

Utilizamos Chaos Mesh para simular condições de falha e validar a resiliência do sistema:

```yaml
# chaos-network-delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: bureau-credito-network-delay
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - innovabiz
    labelSelectors:
      app: bureau-credito
  delay:
    latency: "200ms"
    correlation: "25"
    jitter: "50ms"
  duration: "5m"
  scheduler:
    cron: "@every 10m"
```

### 5.2 Casos de Teste de Resiliência

#### 5.2.1 Teste de Recuperação após Falha do IAM

```go
func TestRecuperacaoFalhaIAM(t *testing.T) {
    // Configurar ambiente de teste
    env := setupTestEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token inicial
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Consulta inicial para verificar funcionamento
    consulta := bureau.ConsultaCredito{
        ConsultaID:       fmt.Sprintf("test-%s", uuid.New().String()),
        TipoConsulta:     bureau.ConsultaBasica,
        Finalidade:       bureau.FinalidadeVerificacaoCliente,
        EntidadeID:       "entidade-teste",
        DocumentoCliente: "12345678901",
        UsuarioID:        "admin",
        MarketContext:    adapter.MarketContext{Market: "global"},
        MFALevel:         "standard",
    }
    
    // Verificar funcionamento normal
    _, err = env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
    require.NoError(t, err)
    
    // Parar contêiner IAM
    err = env.IAMContainer.Stop(context.Background())
    require.NoError(t, err)
    
    // Tentar consulta durante falha (deve usar cache de token/autorização)
    consulta.ConsultaID = fmt.Sprintf("test-%s", uuid.New().String())
    resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
    
    // Em caso de cache de autorização implementado
    if err == nil {
        assert.NotNil(t, resultado)
    } else {
        // Falha esperada se não houver cache
        assert.Contains(t, err.Error(), "autenticação")
    }
    
    // Reiniciar contêiner IAM
    err = env.IAMContainer.Start(context.Background())
    require.NoError(t, err)
    
    // Aguardar recuperação
    time.Sleep(5 * time.Second)
    
    // Obter novo token após recuperação
    newToken, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    newCtx := context.WithValue(context.Background(), "Authorization", "Bearer "+newToken)
    
    // Verificar recuperação completa
    consulta.ConsultaID = fmt.Sprintf("test-%s", uuid.New().String())
    resultado, err = env.BureauCreditoClient.RealizarConsulta(newCtx, consulta)
    
    assert.NoError(t, err)
    assert.NotNil(t, resultado)
}
```

#### 5.2.2 Teste de Circuit Breaker

```go
func TestCircuitBreaker(t *testing.T) {
    // Configurar ambiente de teste com múltiplas instâncias
    env := setupMultiInstanceEnvironment(t)
    defer env.CleanupFunc()
    
    // Obter token de acesso
    token, err := env.IAMClient.Login(context.Background(), "admin", "admin")
    require.NoError(t, err)
    
    ctx := context.WithValue(context.Background(), "Authorization", "Bearer "+token)
    
    // Consulta padrão
    consulta := bureau.ConsultaCredito{
        TipoConsulta:     bureau.ConsultaBasica,
        Finalidade:       bureau.FinalidadeVerificacaoCliente,
        EntidadeID:       "entidade-teste",
        DocumentoCliente: "12345678901",
        UsuarioID:        "admin",
        MarketContext:    adapter.MarketContext{Market: "global"},
        MFALevel:         "standard",
    }
    
    // Parar uma instância e verificar failover
    err = env.BureauCreditoContainers[0].Stop(context.Background())
    require.NoError(t, err)
    
    // Realizar múltiplas consultas para verificar circuit breaker
    successCount := 0
    for i := 0; i < 20; i++ {
        consulta.ConsultaID = fmt.Sprintf("test-%s", uuid.New().String())
        resultado, err := env.BureauCreditoClient.RealizarConsulta(ctx, consulta)
        
        if err == nil {
            successCount++
            assert.NotNil(t, resultado)
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Verificar que algumas consultas foram bem-sucedidas (através de outras instâncias)
    assert.True(t, successCount > 0, "Nenhuma consulta bem-sucedida após failover")
}
```

## 6. Configuração do Pipeline de CI/CD

### 6.1 Pipeline Jenkins

```groovy
// Jenkinsfile
pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: golang
    image: golang:1.22
    command:
    - cat
    tty: true
  - name: k6
    image: loadimpact/k6:latest
    command:
    - cat
    tty: true
  - name: docker
    image: docker:latest
    command:
    - cat
    tty: true
    volumeMounts:
    - mountPath: /var/run/docker.sock
      name: docker-sock
  volumes:
  - name: docker-sock
    hostPath:
      path: /var/run/docker.sock
"""
        }
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Unit Tests') {
            steps {
                container('golang') {
                    sh 'cd CoreModules/IAM && go test -v -cover ./modules/bureau/...'
                }
            }
        }
        
        stage('Integration Tests') {
            steps {
                container('golang') {
                    sh 'cd CoreModules/IAM && go test -v -tags=integration ./tests/integration/bureau/...'
                }
            }
        }
        
        stage('Build Image') {
            steps {
                container('docker') {
                    sh 'docker build -t innovabiz/bureau-credito:${BUILD_NUMBER} CoreModules/IAM/bureau'
                }
            }
        }
        
        stage('Performance Tests') {
            steps {
                container('k6') {
                    sh 'k6 run --env ENVIRONMENT=quality CoreModules/IAM/tests/performance/k6-bureau-credito-performance.js'
                }
            }
        }
        
        stage('Push Image') {
            when {
                branch 'main'
            }
            steps {
                container('docker') {
                    withCredentials([string(credentialsId: 'docker-registry-credentials', variable: 'DOCKER_AUTH')]) {
                        sh 'echo $DOCKER_AUTH | docker login -u innovabiz --password-stdin'
                        sh 'docker push innovabiz/bureau-credito:${BUILD_NUMBER}'
                        sh 'docker tag innovabiz/bureau-credito:${BUILD_NUMBER} innovabiz/bureau-credito:latest'
                        sh 'docker push innovabiz/bureau-credito:latest'
                    }
                }
            }
        }
    }
    
    post {
        always {
            junit 'CoreModules/IAM/**/junit-*.xml'
            archiveArtifacts artifacts: 'CoreModules/IAM/tests/performance/results/*.json', fingerprint: true
            
            // Publicar relatório de cobertura
            publishHTML([
                allowMissing: false,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'CoreModules/IAM/coverage',
                reportFiles: 'index.html',
                reportName: 'Coverage Report'
            ])
        }
    }
}
```

### 6.2 Configuração SonarQube

```json
// sonar-project.properties
sonar.projectKey=innovabiz-bureau-credito
sonar.projectName=INNOVABIZ Bureau de Crédito
sonar.sources=CoreModules/IAM/modules/bureau
sonar.tests=CoreModules/IAM/modules/bureau/test,CoreModules/IAM/tests/integration/bureau
sonar.go.coverage.reportPaths=CoreModules/IAM/coverage/coverage.out
sonar.coverage.exclusions=**/*_test.go
sonar.sourceEncoding=UTF-8
```

## 7. Conclusão

Este plano de testes automatizados para o módulo Bureau de Crédito abrange testes unitários, de integração, conformidade por mercado, desempenho e resiliência. A implementação destes testes garante que o módulo atenda aos requisitos técnicos, regulatórios e de integração dentro da plataforma INNOVABIZ.

### 7.1 Métricas de Qualidade

| Métrica | Meta | Método de Verificação |
|---------|------|------------------------|
| Cobertura de Código | >= 85% | Relatórios de cobertura Go |
| Conformidade Regulatória | 100% | Testes de compliance por mercado |
| Tempo de Resposta P95 | < 500ms | Testes de desempenho k6 |
| Taxa de Erro | < 1% | Testes de carga e monitoramento |
| MTTR (Tempo Médio de Recuperação) | < 60s | Testes de resiliência |

### 7.2 Próximos Passos

1. Implementação de testes end-to-end com simulações de clientes reais
2. Expansão dos testes de compliance para mercados adicionais (China, BRICS)
3. Testes de segurança automatizados (DAST/SAST)
4. Integração com sistema de geração de dados de teste sintéticos
5. Automação de testes de observabilidade avançados

---

**Autor**: Equipe de Qualidade INNOVABIZ  
**Versão**: 1.0.0  
**Data**: 2025-02-18  
**Status**: Aprovado  
**Classificação**: Confidencial