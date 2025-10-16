# Guia de Integração: Bureau de Crédito com MCP-IAM Observability

## 1. Introdução

Este documento detalha o processo de integração do módulo Bureau de Crédito (Central de Risco) com o MCP-IAM Observability e demais módulos core da plataforma INNOVABIZ. O guia segue a arquitetura multi-mercado, multi-tenant, multi-camada e multi-contexto, abordando todos os aspectos técnicos, de compliance e governança necessários para uma implementação bem-sucedida.

### 1.1 Escopo do Documento

- Processo completo de integração do Bureau de Crédito
- Configurações específicas por mercado (Angola, Brasil, União Europeia, EUA, China, CPLP, SADC, PALOP, BRICS)
- Integrações com módulos core (IAM, Payment Gateway, Risk Management, Mobile Money, Marketplace)
- Implementação de observabilidade avançada
- Configurações de compliance e governança
- Gestão de dados e segurança
- Implantação e monitoramento

### 1.2 Público-Alvo

- Equipes de Desenvolvimento e DevOps
- Arquitetos de Sistemas e Segurança
- Engenheiros de Observabilidade
- Especialistas em Compliance e Governança
- Administradores de Sistemas
- Gestores de Projeto

### 1.3 Requisitos Prévios

- Acesso ao ambiente Kubernetes da plataforma INNOVABIZ
- Credenciais para registros de contêineres
- Módulos IAM e OpenTelemetry já implantados
- Configuração do API Gateway KrakenD
- Acesso às ferramentas de observabilidade (Prometheus, Grafana, Jaeger)
- Documentação das regulamentações específicas por mercado

## 2. Visão Geral da Arquitetura de Integração

O Bureau de Crédito integra-se com o MCP-IAM Observability e demais módulos core através de uma arquitetura de microserviços, utilizando comunicação gRPC para operações síncronas de alto desempenho e mensageria assíncrona para operações que podem ser processadas em background.

### 2.1 Diagrama de Integração

```
┌─────────────────────────────────────────────────────────────────────┐
│                         API Gateway (KrakenD)                        │
└─────────────────────────────────────────────────────────────────────┘
                ▲                    ▲                    ▲
                │                    │                    │
                ▼                    ▼                    ▼
┌───────────────────┐    ┌─────────────────────┐    ┌────────────────┐
│       IAM         │◄─►│   Bureau Crédito    │◄─►│  Payment Gateway │
└───────────────────┘    └─────────────────────┘    └────────────────┘
                ▲                    ▲                    ▲
                │                    │                    │
                ▼                    ▼                    ▼
┌───────────────────┐    ┌─────────────────────┐    ┌────────────────┐
│  Risk Management  │◄─►│    Mobile Money     │◄─►│   Marketplace   │
└───────────────────┘    └─────────────────────┘    └────────────────┘
                            ▲           ▲
                            │           │
                ┌───────────┘           └───────────┐
                ▼                                   ▼
┌───────────────────────┐               ┌─────────────────────────┐
│ OpenTelemetry Collector│               │ MCP-IAM Observability  │
└───────────────────────┘               └─────────────────────────┘
                ▲                                   ▲
                │                                   │
                ▼                                   ▼
┌───────────────────────┐               ┌─────────────────────────┐
│      Prometheus       │               │        Grafana          │
└───────────────────────┘               └─────────────────────────┘
```

### 2.2 Fluxos Principais de Integração

1. **Bureau de Crédito → IAM**
   - Autenticação de usuários
   - Autorização baseada em escopos
   - Validação de MFA
   - Gestão de consentimento

2. **Bureau de Crédito → Payment Gateway**
   - Verificação de score para autorização de pagamentos
   - Validação de restrições para transações

3. **Bureau de Crédito → Risk Management**
   - Fornecimento de histórico de crédito para análise de risco
   - Informações para scoring de clientes

4. **Bureau de Crédito → Mobile Money**
   - Validação de clientes
   - Suporte a decisões de concessão de limite

5. **Bureau de Crédito → MCP-IAM Observability**
   - Exportação de spans de rastreabilidade
   - Registro de eventos de auditoria e segurança
   - Envio de métricas de operação

## 3. Integração com MCP-IAM Observability

### 3.1 Configuração do Adaptador de Observabilidade

O primeiro passo da integração é configurar o adaptador de observabilidade no Bureau de Crédito:

```go
observability, err := adapter.NewIAMObservability(adapter.ObservabilityConfig{
    ServiceName:       "bureau-credito",
    ServiceVersion:    os.Getenv("SERVICE_VERSION"),
    Environment:       environment,
    DefaultMarket:     market,
    DefaultTenantID:   "global",
    DefaultTenantType: tenantType,
})
if err != nil {
    logger.Fatal("Falha ao inicializar observabilidade", zap.Error(err))
}
defer observability.Shutdown(context.Background())
```

### 3.2 Implementação de Rastreabilidade

Para cada operação principal do Bureau de Crédito, implemente spans de rastreabilidade:

```go
// Iniciar span para consulta ao bureau
ctx, span := bc.observability.Tracer().Start(ctx, "bureau_credito_consulta",
    trace.WithAttributes(
        attribute.String("consulta_id", consulta.ConsultaID),
        attribute.String("tipo_consulta", string(consulta.TipoConsulta)),
        attribute.String("finalidade", string(consulta.Finalidade)),
        attribute.String("entidade_id", consulta.EntidadeID),
        attribute.String("documento_cliente", consulta.DocumentoCliente),
        attribute.String("market", consulta.MarketContext.Market),
    ),
)
defer span.End()
```

### 3.3 Registro de Eventos de Auditoria

Para garantir auditabilidade completa, registre eventos de auditoria para todas as operações críticas:

```go
// Registrar evento de auditoria para consulta iniciada
bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
    "bureau_credito_consulta_iniciada",
    fmt.Sprintf("Consulta %s iniciada para documento %s (tipo: %s, finalidade: %s)",
        consulta.ConsultaID, consulta.DocumentoCliente, 
        consulta.TipoConsulta, consulta.Finalidade))
```

### 3.4 Registro de Eventos de Segurança

Para eventos relacionados à segurança, utilize:

```go
// Registrar evento de segurança para falha de autenticação
bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
    constants.SecurityEventSeverityHigh, "bureau_credito_auth_failure",
    fmt.Sprintf("Falha de autenticação na consulta %s: %v", 
        consulta.ConsultaID, err))
```

### 3.5 Registro de Métricas

Para monitoramento e análise, registre métricas de operação:

```go
// Registrar métrica de consulta
bc.observability.RecordMetric(consulta.MarketContext, 
    "bureau_credito_consultas_total", string(consulta.TipoConsulta), 1)

// Registrar histograma de tempo de processamento
bc.observability.RecordHistogram(consulta.MarketContext, 
    "bureau_credito_tempo_processamento", float64(processTime), 
    string(consulta.TipoConsulta))
```## 4. Configurações Específicas por Mercado

A integração do Bureau de Crédito deve considerar configurações específicas para cada mercado-alvo, incluindo regulamentações, frameworks de compliance e requisitos específicos.

### 4.1 Angola (BNA)

#### 4.1.1 Regulamentações Aplicáveis
- Lei do BNA sobre Centrais de Informação de Crédito
- Lei de Proteção de Dados Pessoais de Angola
- Regulamentos do Sistema de Pagamentos de Angola (SPA)

#### 4.1.2 Configurações Específicas

```go
// Configurações BNA para Bureau de Crédito
bnaConfig := BureauCreditoConfig{
    Market:                 constants.MarketAngola,
    TempoRetencaoHistorico: map[string]int{
        constants.MarketAngola: 60, // 5 anos conforme BNA
    },
    ConsentimentoObrigatorio: map[string]bool{
        constants.MarketAngola: true, // BNA exige consentimento explícito
    },
    NotificacaoObrigatoria: map[string]bool{
        constants.MarketAngola: true, // Notificação obrigatória para consultas
    },
    CamposObrigatorios: map[string][]string{
        string(ConsultaCompleta): {"documentoCliente", "nomeCliente", "finalidade", "solicitanteID", "consentimentoID"},
    },
}

// Regra específica BNA para consultas completas
bc.RegistrarRegraCompliance(RegrasCompliance{
    ID:           "COMPLIANCE-AO-001",
    Market:       constants.MarketAngola,
    Description:  "Validação de regras BNA para consultas completas",
    Framework:    []string{"BNA", "ISO 27001"},
    MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico)},
    Validate: func(consulta *ConsultaCredito) (bool, string, error) {
        // Implementação da regra específica BNA
        // ...
    },
})
```

#### 4.1.3 Requisitos de Notificação

Para Angola, é obrigatório enviar notificação ao cliente para qualquer consulta completa:

```go
// Processamento de notificação BNA
if consulta.MarketContext.Market == constants.MarketAngola && consulta.TipoConsulta == ConsultaCompleta {
    // Enviar notificação (implementação)
    bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
        "bna_notificacao_enviada",
        fmt.Sprintf("Notificação BNA enviada para documento %s referente à consulta %s", 
            consulta.DocumentoCliente, consulta.ConsultaID))
}
```

### 4.2 Brasil (BACEN/LGPD)

#### 4.2.1 Regulamentações Aplicáveis
- LGPD (Lei Geral de Proteção de Dados)
- Resolução CMN sobre SCR (Sistema de Informações de Crédito)
- Normas BACEN sobre compartilhamento de informações
- Lei do Cadastro Positivo

#### 4.2.2 Configurações Específicas

```go
// Configurações BACEN/LGPD para Bureau de Crédito
brazilConfig := BureauCreditoConfig{
    Market:                 constants.MarketBrazil,
    TempoRetencaoHistorico: map[string]int{
        constants.MarketBrazil: 60, // 5 anos conforme BACEN
    },
    ConsentimentoObrigatorio: map[string]bool{
        constants.MarketBrazil: true, // LGPD exige consentimento explícito
    },
    NotificacaoObrigatoria: map[string]bool{
        constants.MarketBrazil: true, // Notificação para consultas com restrições
    },
    CamposObrigatorios: map[string][]string{
        string(ConsultaCompleta): {"documentoCliente", "nomeCliente", "finalidade", "solicitanteID", "consentimentoID"},
        string(ConsultaHistorico): {"documentoCliente", "finalidade", "consentimentoID"},
    },
}

// Regra específica LGPD para consultas
bc.RegistrarRegraCompliance(RegrasCompliance{
    ID:           "COMPLIANCE-BR-001",
    Market:       constants.MarketBrazil,
    Description:  "Validação de regras BACEN e LGPD",
    Framework:    []string{"BACEN", "LGPD", "SCR"},
    MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico)},
    Validate: func(consulta *ConsultaCredito) (bool, string, error) {
        // Implementação da regra específica BACEN/LGPD
        // ...
    },
})
```

#### 4.2.3 Requisitos de Notificação

Para Brasil, é obrigatório notificar o cliente quando a consulta retornar restrições:

```go
// Processamento de notificação LGPD/BACEN
if consulta.MarketContext.Market == constants.MarketBrazil && len(resultado.RestricoesList) > 0 {
    // Enviar notificação (implementação)
    bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
        "lgpd_bacen_notificacao_enviada",
        fmt.Sprintf("Notificação LGPD/BACEN enviada para documento %s referente a %d restrições", 
            consulta.DocumentoCliente, len(resultado.RestricoesList)))
}
```

### 4.3 União Europeia (GDPR/PSD2)

#### 4.3.1 Regulamentações Aplicáveis
- GDPR (General Data Protection Regulation)
- PSD2 (Payment Services Directive 2)
- EBA Guidelines on Loan Origination and Monitoring
- DORA (Digital Operational Resilience Act)

#### 4.3.2 Configurações Específicas

```go
// Configurações GDPR/PSD2 para Bureau de Crédito
euConfig := BureauCreditoConfig{
    Market:                 constants.MarketEU,
    TempoRetencaoHistorico: map[string]int{
        constants.MarketEU: 24, // 2 anos conforme GDPR
    },
    ConsentimentoObrigatorio: map[string]bool{
        constants.MarketEU: true, // GDPR exige consentimento explícito
    },
    NotificacaoObrigatoria: map[string]bool{
        constants.MarketEU: true, // Notificação obrigatória para todas consultas
    },
    CamposObrigatorios: map[string][]string{
        string(ConsultaCompleta): {"documentoCliente", "nomeCliente", "finalidade", "solicitanteID", "consentimentoID"},
    },
}

// Regra específica GDPR para minimização de dados
bc.RegistrarRegraCompliance(RegrasCompliance{
    ID:           "COMPLIANCE-EU-001",
    Market:       constants.MarketEU,
    Description:  "Validação de regras GDPR para consultas de crédito",
    Framework:    []string{"GDPR", "PSD2"},
    MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico), string(ConsultaScore)},
    Validate: func(consulta *ConsultaCredito) (bool, string, error) {
        // Implementação da regra específica GDPR
        // ...
    },
})
```

### 4.4 EUA (FCRA/GLBA)

#### 4.4.1 Regulamentações Aplicáveis
- FCRA (Fair Credit Reporting Act)
- GLBA (Gramm-Leach-Bliley Act)
- ECOA (Equal Credit Opportunity Act)
- CFPB (Consumer Financial Protection Bureau) regulations

#### 4.4.2 Configurações Específicas

```go
// Configurações FCRA/GLBA para Bureau de Crédito
usaConfig := BureauCreditoConfig{
    Market:                 constants.MarketUSA,
    TempoRetencaoHistorico: map[string]int{
        constants.MarketUSA: 84, // 7 anos conforme FCRA
    },
    ConsentimentoObrigatorio: map[string]bool{
        constants.MarketUSA: false, // Consentimento não necessário para finalidades permissíveis
    },
    NotificacaoObrigatoria: map[string]bool{
        constants.MarketUSA: true, // Notificação para decisões negativas
    },
    CamposObrigatorios: map[string][]string{
        string(ConsultaCompleta): {"documentoCliente", "finalidade", "solicitanteID"},
    },
}

// Regra específica FCRA para finalidades permissíveis
bc.RegistrarRegraCompliance(RegrasCompliance{
    ID:           "COMPLIANCE-US-001",
    Market:       constants.MarketUSA,
    Description:  "Validação de regras FCRA para consultas de crédito",
    Framework:    []string{"FCRA", "GLBA"},
    MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico), string(ConsultaScore)},
    Validate: func(consulta *ConsultaCredito) (bool, string, error) {
        // Implementação da regra específica FCRA
        // ...
    },
})
```

## 5. Integração com Módulos Core

### 5.1 Integração com IAM

#### 5.1.1 Configuração de Escopos

O Bureau de Crédito requer os seguintes escopos no IAM:

```json
{
  "scopes": [
    "bureau_credito:consulta:completa",
    "bureau_credito:consulta:score",
    "bureau_credito:consulta:basica",
    "bureau_credito:consulta:restricoes",
    "bureau_credito:consulta:historico",
    "bureau_credito:consulta:relacionamento",
    "bureau_credito:finalidade:concessao_credito",
    "bureau_credito:finalidade:revisao_limites",
    "bureau_credito:finalidade:gerencial_regulador",
    "bureau_credito:finalidade:verificacao_cliente",
    "bureau_credito:finalidade:abertura_conta",
    "bureau_credito:finalidade:prevencao_fraude"
  ]
}
```

#### 5.1.2 Validação de Autenticação e Autorização

```go
// Verificação de MFA
mfaResult, err := bc.observability.ValidateMFA(
    ctx, 
    consulta.MarketContext, 
    consulta.UsuarioID, 
    consulta.MFALevel
)

// Verificação de escopo
scopeResult, err := bc.observability.ValidateScope(
    ctx, 
    consulta.MarketContext, 
    consulta.UsuarioID, 
    fmt.Sprintf("bureau_credito:%s", consulta.TipoConsulta)
)
```

#### 5.1.3 Gestão de Consentimento

```go
// Verificar consentimento
consentResult, err := bc.observability.ValidateConsent(
    ctx,
    consulta.MarketContext,
    consulta.DocumentoCliente,
    consulta.ConsentimentoID
)
```

### 5.2 Integração com Payment Gateway

#### 5.2.1 Consumo de Serviço pelo Payment Gateway

O Payment Gateway pode solicitar consultas ao Bureau de Crédito durante o processamento de transações:

```go
// No Payment Gateway
func (pg *PaymentGateway) verificarRiscoTransacao(ctx context.Context, transaction Transaction) (bool, error) {
    // Construir consulta ao Bureau de Crédito
    consulta := ConsultaCredito{
        ConsultaID:       fmt.Sprintf("PG-%s-%s", transaction.TransactionID, time.Now().Format("20060102150405")),
        TipoConsulta:     ConsultaScore,
        Finalidade:       FinalidadePrevencaoFraude,
        EntidadeID:       transaction.CustomerID,
        TipoEntidade:     transaction.CustomerType,
        DocumentoCliente: transaction.CustomerDocument,
        UsuarioID:        transaction.MerchantID,
        DataConsulta:     time.Now(),
        SolicitanteID:    "payment-gateway",
        MarketContext:    transaction.MarketContext,
        MFALevel:         "standard",
    }
    
    // Chamar o Bureau de Crédito
    resultado, err := bureauCreditoClient.RealizarConsulta(ctx, consulta)
    if err != nil {
        return false, err
    }
    
    // Analisar resultado para decisão de risco
    if resultado.ScoreCredito != nil && *resultado.ScoreCredito < 500 {
        return false, nil // Alto risco, rejeitar transação
    }
    
    return true, nil // Baixo risco, aprovar transação
}
```

#### 5.2.2 Modelo de Integração via API

```yaml
# API Gateway (KrakenD) - Configuração de Endpoint
{
  "endpoint": "/v1/bureau-credito/consultas",
  "method": "POST",
  "backend": [
    {
      "url_pattern": "/consultas",
      "host": ["http://bureau-credito-service:8080"],
      "method": "POST"
    }
  ],
  "extra_config": {
    "auth/validator": {
      "alg": "RS256",
      "jwk_url": "http://iam-service:8080/.well-known/jwks.json",
      "cache": true,
      "scopes": [
        "bureau_credito:consulta:score"
      ]
    }
  }
}
```

### 5.3 Integração com Risk Management

#### 5.3.1 Consumo de Serviço pelo Risk Management

O módulo de Risk Management utiliza consultas detalhadas para análises de risco avançadas:

```go
// No Risk Management
func (rm *RiskManagement) avaliarRiscoCliente(ctx context.Context, clienteID string, documentoCliente string) (*AvaliacaoRisco, error) {
    // Consulta completa ao Bureau de Crédito
    consulta := ConsultaCredito{
        ConsultaID:       fmt.Sprintf("RM-%s-%s", clienteID, time.Now().Format("20060102150405")),
        TipoConsulta:     ConsultaCompleta,
        Finalidade:       FinalidadeGerencialRegulador,
        EntidadeID:       clienteID,
        DocumentoCliente: documentoCliente,
        UsuarioID:        "risk-management-system",
        DataConsulta:     time.Now(),
        ConsentimentoID:  rm.getConsentimentoID(documentoCliente), // Obter ID de consentimento
        SolicitanteID:    "risk-management",
        MarketContext:    rm.getMarketContext(),
        MFALevel:         "high",
    }
    
    // Chamar o Bureau de Crédito
    resultado, err := bureauCreditoClient.RealizarConsulta(ctx, consulta)
    if err != nil {
        return nil, err
    }
    
    // Processar resultado para análise de risco
    avaliacaoRisco := &AvaliacaoRisco{
        ClienteID:        clienteID,
        DocumentoCliente: documentoCliente,
        DataAnalise:      time.Now(),
    }
    
    // Processar score de crédito
    if resultado.ScoreCredito != nil {
        avaliacaoRisco.ScoreCredito = *resultado.ScoreCredito
    }
    
    // Processar restrições
    avaliacaoRisco.QuantidadeRestricoes = len(resultado.RestricoesList)
    
    // Calcular nível de risco
    avaliacaoRisco.NivelRisco = rm.calcularNivelRisco(resultado)
    
    return avaliacaoRisco, nil
}
```### 5.4 Integração com Mobile Money

#### 5.4.1 Consultas para Verificação de Clientes

O Mobile Money utiliza o Bureau de Crédito para verificação de clientes e concessão de microcrédito:

```go
// No Mobile Money
func (mm *MobileMoney) verificarCliente(ctx context.Context, clienteID string, documentoCliente string) error {
    // Consulta básica ao Bureau de Crédito
    consulta := ConsultaCredito{
        ConsultaID:       fmt.Sprintf("MM-%s-%s", clienteID, time.Now().Format("20060102150405")),
        TipoConsulta:     ConsultaBasica,
        Finalidade:       FinalidadeVerificacaoCliente,
        EntidadeID:       clienteID,
        DocumentoCliente: documentoCliente,
        UsuarioID:        "mobile-money-system",
        DataConsulta:     time.Now(),
        SolicitanteID:    "mobile-money",
        MarketContext:    mm.getMarketContext(),
        MFALevel:         "standard",
    }
    
    // Chamar o Bureau de Crédito
    resultado, err := bureauCreditoClient.RealizarConsulta(ctx, consulta)
    if err != nil {
        return err
    }
    
    // Verificar se cliente tem restrições impeditivas
    if len(resultado.RestricoesList) > 0 {
        for _, restricao := range resultado.RestricoesList {
            if restricao.Metadata["tipo"] == "impeditiva" {
                return fmt.Errorf("cliente possui restrição impeditiva: %s", restricao.Descricao)
            }
        }
    }
    
    return nil
}
```

#### 5.4.2 Análise para Concessão de Microcrédito

```go
// No Mobile Money
func (mm *MobileMoney) avaliarMicrocredito(ctx context.Context, clienteID string, documentoCliente string, valorSolicitado float64) (*DecisaoMicrocredito, error) {
    // Consulta de score ao Bureau de Crédito
    consulta := ConsultaCredito{
        ConsultaID:       fmt.Sprintf("MM-MC-%s-%s", clienteID, time.Now().Format("20060102150405")),
        TipoConsulta:     ConsultaScore,
        Finalidade:       FinalidadeConcessaoCredito,
        EntidadeID:       clienteID,
        DocumentoCliente: documentoCliente,
        UsuarioID:        "mobile-money-system",
        DataConsulta:     time.Now(),
        ConsentimentoID:  mm.getConsentimentoID(documentoCliente), // Obter ID de consentimento
        SolicitanteID:    "mobile-money",
        MarketContext:    mm.getMarketContext(),
        MFALevel:         "high",
    }
    
    // Chamar o Bureau de Crédito
    resultado, err := bureauCreditoClient.RealizarConsulta(ctx, consulta)
    if err != nil {
        return nil, err
    }
    
    // Análise de decisão de microcrédito
    decisao := &DecisaoMicrocredito{
        ClienteID:        clienteID,
        DocumentoCliente: documentoCliente,
        ValorSolicitado:  valorSolicitado,
        DataAnalise:      time.Now(),
    }
    
    // Utilizar score para decisão
    if resultado.ScoreCredito != nil {
        score := *resultado.ScoreCredito
        
        if score >= 700 {
            // Score alto - aprovar valor total
            decisao.ValorAprovado = valorSolicitado
            decisao.StatusAprovacao = "APROVADO"
        } else if score >= 500 {
            // Score médio - aprovar parcialmente
            decisao.ValorAprovado = valorSolicitado * 0.7
            decisao.StatusAprovacao = "APROVADO_PARCIAL"
        } else {
            // Score baixo - negar
            decisao.ValorAprovado = 0
            decisao.StatusAprovacao = "NEGADO"
        }
    } else {
        decisao.ValorAprovado = 0
        decisao.StatusAprovacao = "NEGADO"
    }
    
    return decisao, nil
}
```

### 5.5 Integração com Marketplace

#### 5.5.1 Verificação de Vendedores e Compradores

```go
// No Marketplace
func (mp *Marketplace) verificarVendedor(ctx context.Context, vendedorID string, documentoVendedor string) (*VerificacaoVendedor, error) {
    // Consulta de restrições ao Bureau de Crédito
    consulta := ConsultaCredito{
        ConsultaID:       fmt.Sprintf("MP-V-%s-%s", vendedorID, time.Now().Format("20060102150405")),
        TipoConsulta:     ConsultaRestricoes,
        Finalidade:       FinalidadeVerificacaoCliente,
        EntidadeID:       vendedorID,
        DocumentoCliente: documentoVendedor,
        UsuarioID:        "marketplace-system",
        DataConsulta:     time.Now(),
        SolicitanteID:    "marketplace",
        MarketContext:    mp.getMarketContext(),
        MFALevel:         "standard",
    }
    
    // Chamar o Bureau de Crédito
    resultado, err := bureauCreditoClient.RealizarConsulta(ctx, consulta)
    if err != nil {
        return nil, err
    }
    
    // Análise de verificação do vendedor
    verificacao := &VerificacaoVendedor{
        VendedorID:        vendedorID,
        DocumentoVendedor: documentoVendedor,
        DataVerificacao:   time.Now(),
    }
    
    // Verificar restrições
    verificacao.PossuiRestricoes = len(resultado.RestricoesList) > 0
    verificacao.QuantidadeRestricoes = len(resultado.RestricoesList)
    
    // Determinar nível de verificação
    if !verificacao.PossuiRestricoes {
        verificacao.NivelVerificacao = "COMPLETO"
    } else if verificacao.QuantidadeRestricoes < 2 {
        verificacao.NivelVerificacao = "PARCIAL"
    } else {
        verificacao.NivelVerificacao = "REJEITADO"
    }
    
    return verificacao, nil
}
```

## 6. Implantação e Configuração

### 6.1 Configuração de Kubernetes

#### 6.1.1 Manifesto de Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bureau-credito
  namespace: innovabiz
  labels:
    app: bureau-credito
    module: core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bureau-credito
  template:
    metadata:
      labels:
        app: bureau-credito
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: bureau-credito
        image: innovabiz/bureau-credito:1.0.0
        imagePullPolicy: Always
        env:
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: innovabiz-config
              key: environment
        - name: MARKET
          valueFrom:
            configMapKeyRef:
              name: innovabiz-config
              key: default_market
        - name: TENANT_TYPE
          valueFrom:
            configMapKeyRef:
              name: innovabiz-config
              key: default_tenant_type
        - name: SERVICE_VERSION
          value: "1.0.0"
        - name: LOG_LEVEL
          value: "info"
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector.observability.svc.cluster.local:4317"
        ports:
        - containerPort: 8080
        - containerPort: 9090
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
      imagePullSecrets:
      - name: innovabiz-registry
```

#### 6.1.2 Manifesto de Serviço

```yaml
apiVersion: v1
kind: Service
metadata:
  name: bureau-credito
  namespace: innovabiz
  labels:
    app: bureau-credito
    module: core
spec:
  selector:
    app: bureau-credito
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

### 6.2 Configuração de API Gateway (KrakenD)

```json
{
  "version": 3,
  "endpoints": [
    {
      "endpoint": "/api/v1/bureau-credito/consultas",
      "method": "POST",
      "backend": [
        {
          "url_pattern": "/consultas",
          "host": [
            "http://bureau-credito.innovabiz.svc.cluster.local:8080"
          ],
          "method": "POST",
          "encoding": "json",
          "timeout": "3s"
        }
      ],
      "extra_config": {
        "auth/validator": {
          "alg": "RS256",
          "jwk_url": "http://iam.innovabiz.svc.cluster.local:8080/.well-known/jwks.json",
          "cache": true,
          "disable_jwk_security": false,
          "operation_debug": false,
          "propagate_claims": [
            ["sub", "user_id"],
            ["iss", "issuer"]
          ]
        },
        "qos/circuit-breaker": {
          "interval": 60,
          "timeout": 10,
          "max_errors": 5,
          "log_status_change": true
        },
        "qos/ratelimit/router": {
          "max_rate": 100,
          "client_max_rate": 10,
          "strategy": "ip"
        },
        "security/cors": {
          "allow_origins": ["*"],
          "allow_methods": ["POST", "OPTIONS"],
          "allow_headers": ["Origin", "Authorization", "Content-Type"],
          "expose_headers": ["Content-Length"],
          "max_age": "12h"
        }
      }
    }
  ]
}
```

### 6.3 Configuração de Dashboards (Grafana)

Implemente os seguintes dashboards para monitoramento do Bureau de Crédito:

#### 6.3.1 Dashboard Operacional

```json
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 28,
  "links": [],
  "panels": [
    {
      "title": "Volume de Consultas por Tipo",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(bureau_credito_consultas_total{job=\"bureau-credito\"}[5m])) by (tipo_consulta)",
          "legendFormat": "{{tipo_consulta}}"
        }
      ]
    },
    {
      "title": "Tempo Médio de Resposta",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "rate(bureau_credito_tempo_processamento_sum{job=\"bureau-credito\"}[5m]) / rate(bureau_credito_tempo_processamento_count{job=\"bureau-credito\"}[5m])",
          "legendFormat": "Tempo Médio (ms)"
        }
      ]
    },
    {
      "title": "Taxa de Erros",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(bureau_credito_consultas_erro_total{job=\"bureau-credito\"}[5m])) / sum(rate(bureau_credito_consultas_total{job=\"bureau-credito\"}[5m]))",
          "legendFormat": "Taxa de Erro (%)"
        }
      ]
    }
  ],
  "schemaVersion": 27,
  "style": "dark",
  "tags": ["bureau-credito", "operacional"],
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "browser",
  "title": "Bureau de Crédito - Dashboard Operacional",
  "uid": "bureau-credito-ops"
}
```

#### 6.3.2 Dashboard de Compliance

```json
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 29,
  "links": [],
  "panels": [
    {
      "title": "Violações de Compliance por Mercado",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(bureau_credito_compliance_violation_total{job=\"bureau-credito\"}[5m])) by (market)",
          "legendFormat": "{{market}}"
        }
      ]
    },
    {
      "title": "Notificações por Regulador",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(bureau_credito_notificacoes{job=\"bureau-credito\"}[5m])) by (regulador)",
          "legendFormat": "{{regulador}}"
        }
      ]
    }
  ],
  "schemaVersion": 27,
  "style": "dark",
  "tags": ["bureau-credito", "compliance"],
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "browser",
  "title": "Bureau de Crédito - Dashboard de Compliance",
  "uid": "bureau-credito-compliance"
}
```

## 7. Verificação de Integração

### 7.1 Lista de Verificação da Integração

Utilize a seguinte lista de verificação para validar a integração do Bureau de Crédito:

| Item | Descrição | Status |
|------|-----------|--------|
| 1 | Verificar comunicação com IAM para autenticação/autorização | ⬜ |
| 2 | Confirmar telemetria OpenTelemetry no coletor | ⬜ |
| 3 | Validar registro de métricas no Prometheus | ⬜ |
| 4 | Confirmar eventos de auditoria no storage | ⬜ |
| 5 | Validar eventos de segurança no SIEM | ⬜ |
| 6 | Verificar configurações específicas por mercado | ⬜ |
| 7 | Confirmar integração com Payment Gateway | ⬜ |
| 8 | Verificar integração com Risk Management | ⬜ |
| 9 | Validar integração com Mobile Money | ⬜ |
| 10 | Testar API via Krakend | ⬜ |
| 11 | Verificar dashboards no Grafana | ⬜ |
| 12 | Validar alarmes e notificações | ⬜ |

### 7.2 Testes de Integração

Execute os seguintes testes para validar a integração completa:

```bash
# 1. Teste de autenticação/autorização
curl -X POST \
  https://api.innovabiz.com/api/v1/bureau-credito/consultas \
  -H "Authorization: Bearer $IAM_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tipoConsulta": "ConsultaScore",
    "finalidade": "FinalidadeVerificacaoCliente",
    "documentoCliente": "12345678901",
    "entidadeID": "test-entity"
  }'

# 2. Verificar registro de spans no Jaeger
curl -X GET \
  http://jaeger-query.observability.svc.cluster.local:16686/api/traces?service=bureau-credito \
  -H "Content-Type: application/json"

# 3. Verificar métricas no Prometheus
curl -X GET \
  http://prometheus.observability.svc.cluster.local:9090/api/v1/query?query=bureau_credito_consultas_total \
  -H "Content-Type: application/json"
```

## 8. Melhores Práticas e Considerações

### 8.1 Segurança

- Utilize sempre autenticação JWT com validação adequada de tokens
- Implemente MFA adaptativo conforme criticidade da operação
- Aplique mascaramento de dados sensíveis em logs e telemetria
- Mantenha registros de auditoria imutáveis e completos
- Implemente proteção contra ataques de injeção e DoS

### 8.2 Performance

- Utilize cache distribuído para resultados frequentes
- Implemente circuit breaker para proteger serviços dependentes
- Configure timeouts adequados para chamadas entre serviços
- Utilize instrumentação OpenTelemetry para identificar gargalos
- Monitore ativamente tempos de resposta e taxas de erro

### 8.3 Compliance

- Mantenha documentação atualizada por mercado
- Revise periodicamente regras de compliance
- Implemente processos para atualizações regulatórias
- Realize auditorias periódicas de acessos e consultas
- Garanta rastreabilidade completa de todas as operações

### 8.4 Observabilidade

- Configure alertas para anomalias em padrões de uso
- Implemente logging contextual com correlação de spans
- Utilize dashboards específicos por função (operacional, compliance, segurança)
- Configure retenção adequada de telemetria conforme regulamentações
- Implemente health checks completos com verificação de dependências

## 9. Referências

1. [Documentação MCP-IAM Observability](https://internal.innovabiz.com/docs/mcp-iam/observability)
2. [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
3. [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)
4. [KrakenD API Gateway Documentation](https://www.krakend.io/docs/)
5. [GDPR Official Documentation](https://gdpr-info.eu/)
6. [FCRA Compliance Guide](https://www.ftc.gov/business-guidance/resources/fair-credit-reporting-act)
7. [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)
8. [Banco Nacional de Angola - Regulações](https://www.bna.ao/)
9. [BACEN - Sistema de Informações de Crédito](https://www.bcb.gov.br/estabilidadefinanceira/scr)

---

**Autor**: Equipe de Desenvolvimento INNOVABIZ  
**Versão**: 1.0.0  
**Data**: 2025-02-18  
**Status**: Aprovado  
**Classificação**: Confidencial