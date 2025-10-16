# Arquitetura do IAM Audit Service - INNOVABIZ

## Visão Geral

Este documento descreve a arquitetura detalhada do serviço IAM Audit, um componente crítico da plataforma INNOVABIZ, responsável por registrar, processar, analisar e governar eventos de auditoria relacionados à gestão de identidade e acesso. A arquitetura segue os princípios e diretrizes estabelecidos nas regras globais da plataforma INNOVABIZ, bem como as melhores práticas internacionais de segurança, observabilidade e governança de dados.

**Autor:** INNOVABIZ DevOps & Architecture Team  
**Versão:** 2.1.0  
**Data:** 2025-07-31  
**Classificação:** Restrito

## 1. Contexto Estratégico

O IAM Audit Service se posiciona estrategicamente na plataforma INNOVABIZ como um componente transversal que implementa os requisitos regulatórios e de compliance para operações de identidade e acesso em todos os módulos da plataforma. Conforme análises do Gartner e da Forrester, a implementação de auditoria robusta em sistemas IAM é um componente crítico para a maturidade de segurança digital.

### 1.1 Alinhamento com Tendências de Mercado

**Gartner Magic Quadrant para PAM (2025):**
- Auditoria imutável como requisito para soluções líderes
- Capacidades de detecção de anomalias em tempo real
- Integração com observabilidade empresarial

**Forrester Wave para CIAM (2025):**
- Conformidade com múltiplas jurisdições regulatórias
- Auditoria baseada em ML para comportamentos anômalos
- Transparência em processamento de dados pessoais

## 2. Arquitetura de Referência

A arquitetura do IAM Audit Service segue o modelo de arquitetura hexagonal (ports & adapters), permitindo isolamento claro entre o domínio de negócios e os aspectos tecnológicos, facilitando adaptabilidade a diferentes ambientes e requisitos regulatórios.

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│                      IAM Audit Service                          │
│                                                                 │
│  ┌─────────────────┐       ┌─────────────────┐       ┌────────┐ │
│  │     Adapters    │       │     Domain      │       │  Ports │ │
│  │                 │       │                 │       │        │ │
│  │  ┌───────────┐  │       │ ┌─────────────┐ │       │        │ │
│  │  │HTTP/REST  │──┼───────┼─┤Audit Events │ │       │        │ │
│  │  │Controllers│  │       │ │  Service    │ │       │        │ │
│  │  └───────────┘  │       │ └─────────────┘ │       │        │ │
│  │                 │       │                 │       │        │ │
│  │  ┌───────────┐  │       │ ┌─────────────┐ │       │┌──────┐│ │
│  │  │Event      │──┼───────┼─┤Retention    │ │       ││DB    ││ │
│  │  │Publishers │  │       │ │Policies     │ │       ││Repo  ││ │
│  │  └───────────┘  │       │ └─────────────┘ │       │└──────┘│ │
│  │                 │       │                 │       │        │ │
│  │  ┌───────────┐  │       │ ┌─────────────┐ │       │┌──────┐│ │
│  │  │Metrics &  │──┼───────┼─┤Compliance   │ │       ││Event ││ │
│  │  │Observab.  │  │       │ │Verification │ │       ││Store ││ │
│  │  └───────────┘  │       │ └─────────────┘ │       │└──────┘│ │
│  │                 │       │                 │       │        │ │
│  └─────────────────┘       └─────────────────┘       └────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.1 Componentes Principais

1. **Event Processor**
   - Captura eventos de auditoria de todos os módulos IAM
   - Enriquece eventos com contexto multi-tenant e multi-regional
   - Aplica validação e normalização de dados

2. **Retention Manager**
   - Implementa políticas de retenção configuráveis por tenant/região
   - Executa procedimentos de retenção e expurgo conforme regulamentações
   - Garante imutabilidade durante período de retenção obrigatório

3. **Compliance Engine**
   - Verifica conformidade dos eventos com múltiplos frameworks regulatórios
   - Gera relatórios de compliance para auditoria interna e externa
   - Integra-se com o módulo de governança corporativa da plataforma

4. **Observability Integration**
   - Coleta e expõe métricas relevantes para monitoramento
   - Fornece health checks e endpoints diagnósticos
   - Integra-se com o sistema central de observabilidade da plataforma

## 3. Modelo de Dados

### 3.1 Evento de Auditoria

```json
{
  "event_id": "uuid-v4",
  "event_type": "user.permission.granted",
  "entity_type": "user",
  "entity_id": "user-123",
  "actor_id": "admin-456",
  "timestamp": "2025-07-31T08:45:30.123Z",
  "tenant_id": "tenant-abc",
  "region": "br-east",
  "environment": "production",
  "correlation_id": "request-789",
  "resource": {
    "type": "permission",
    "id": "perm-xyz",
    "name": "approve_transactions"
  },
  "context": {
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "session_id": "sess-456",
    "request_path": "/api/v1/permissions",
    "request_method": "POST"
  },
  "metadata": {
    "regulatory_frameworks": ["PCI-DSS", "LGPD"],
    "data_classifications": ["PII", "RESTRICTED"],
    "risk_level": "HIGH"
  }
}
```

### 3.2 Política de Retenção

```json
{
  "policy_id": "pol-123",
  "name": "LGPD-Compliance",
  "description": "Política de retenção conforme LGPD",
  "tenant_id": "tenant-abc",
  "region": "br-east",
  "retention_period_days": 365,
  "applies_to_event_types": ["user.*", "data.access.*"],
  "applies_to_entity_types": ["user", "customer"],
  "regulatory_requirements": [
    {
      "framework": "LGPD",
      "article": "Art. 46",
      "description": "Manutenção de registros por período determinado"
    }
  ],
  "archive_strategy": "cold_storage",
  "encryption_required": true,
  "immutable_period_days": 90
}
```

## 4. Integrações

### 4.1 Integrações Internas

| Módulo | Propósito | Método de Integração |
|--------|-----------|----------------------|
| IAM Core | Captura de eventos de gerenciamento de identidade | Streaming via Kafka + REST API |
| Gateway de Pagamentos | Auditoria de autorizações de transações | Streaming via Kafka |
| Mobile Money | Registro de autenticações e autorizações | REST API + Webhooks |
| Microcrédito | Auditoria de aprovações e verificações | Streaming via Kafka |
| CRM | Registro de acesso a dados sensíveis | REST API |

### 4.2 Integrações Externas

| Sistema | Propósito | Método de Integração |
|---------|-----------|----------------------|
| SIEM | Correlação de eventos de segurança | REST API + Log Shipping |
| GRC Platform | Relatórios de compliance | REST API |
| Observability Stack | Métricas e alertas | Prometheus + Grafana |
| Data Lake | Análise histórica de dados | ETL + Apache Beam |

## 5. Requisitos Não-Funcionais

### 5.1 Requisitos de Segurança

- **Confidencialidade**: 
  - Criptografia em repouso (AES-256)
  - Criptografia em trânsito (TLS 1.3)
  - Tokenização de dados sensíveis (PII, PCI)

- **Integridade**:
  - Assinatura digital de eventos (EdDSA)
  - Verificação de integridade com blockchain
  - Imutabilidade dos registros armazenados

- **Disponibilidade**:
  - SLA de 99.99% (52 minutos de downtime/ano)
  - Arquitetura resiliente multi-regional
  - Recuperação de desastres automatizada

### 5.2 Requisitos de Performance

- **Capacidade de Processamento**:
  - 10,000 eventos por segundo por tenant
  - Latência de processamento < 200ms (p95)
  - Escalabilidade horizontal automatizada

- **Armazenamento**:
  - Suporte a petabytes de dados de auditoria
  - Estratégia de particionamento por tenant/data
  - Políticas de tiering automático de dados

### 5.3 Requisitos de Compliance

| Framework | Requisitos Atendidos | Região Aplicável |
|-----------|---------------------|------------------|
| PCI DSS 4.0 | 10.2, 10.4, 10.5, 10.7 | Global |
| GDPR | Art. 30, 32, 35 | EU |
| LGPD | Art. 46, 47, 48, 50 | Brasil |
| SOX | Seção 404 | Global |
| BNA-REGULATION | Cap. IV, Art 15-18 | Angola |

## 6. Arquitetura de Implantação

### 6.1 Infraestrutura Cloud-Native

```
┌─────────────────────────────────────────────────────────────────┐
│                      Kubernetes Cluster                         │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Ingress/API    │  │  IAM Audit      │  │  Prometheus     │  │
│  │  Gateway        │  │  Service Pods   │  │  & Grafana      │  │
│  │  (Krakend)      │  │  (Multi-tenant) │  │                 │  │
│  └────────┬────────┘  └────────┬────────┘  └─────────────────┘  │
│           │                    │                    ▲           │
│           ▼                    ▼                    │           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌────────┴────────┐  │
│  │                 │  │                 │  │                 │  │
│  │   Kafka         │  │   PostgreSQL    │  │   Redis         │  │
│  │   Event Stream  │  │   (TimescaleDB) │  │   Cache         │  │
│  │                 │  │                 │  │                 │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 6.2 Estratégia Multi-Regional

Para garantir resiliência e conformidade com requisitos de soberania de dados:

- **Região Primária**: Brasil (br-east)
- **Regiões Secundárias**: Angola (ao-central), Europa (eu-central), EUA (us-east)
- **Replicação**: Assíncrona com garantia de entrega (at-least-once)
- **Failover**: Automatizado com RTO < 30 segundos

### 6.3 Estratégia de CI/CD

- **Pipeline de Implantação**: GitLab CI/CD com GitOps
- **Estratégia de Release**: Canary Releases com 5% de tráfego
- **Testes Automatizados**:
  - Testes unitários: > 90% de cobertura
  - Testes de integração: Completos para todas as interfaces
  - Testes de performance: Simulação de 2x carga máxima esperada
  - Testes de segurança: SAST, DAST, SCA, IAST

## 7. Governança de Dados

### 7.1 Classificação de Dados

| Categoria | Exemplos | Controles Aplicados |
|-----------|----------|---------------------|
| Altamente Sensível | Credenciais, Tokens de Acesso | Criptografia de campo, Mascaramento |
| Sensível | PII, Dados de Transação | Criptografia, Acesso Restrito |
| Interno | Metadados de Evento | Controle de Acesso |
| Público | Métricas Agregadas | Sem restrições especiais |

### 7.2 Controle de Acesso

Implementação de controle de acesso baseado em:

- **RBAC**: Papéis definidos por função organizacional
- **ABAC**: Atributos baseados em contexto (tenant, região, classificação de dados)
- **Just-in-Time Access**: Aprovação temporária para operações privilegiadas

### 7.3 Ciclo de Vida dos Dados

1. **Ingestão**: Validação, normalização e enriquecimento
2. **Processamento**: Aplicação de regras de negócio e compliance
3. **Armazenamento**: Distribuição em camadas por temperatura de acesso
4. **Arquivamento**: Migração para storage frio após período configurável
5. **Expurgo**: Exclusão definitiva conforme políticas de retenção

## 8. Design de API

### 8.1 Princípios de Design

- **RESTful**: Seguindo práticas REST para recursos de auditoria
- **GraphQL**: Para consultas complexas e relatórios personalizados
- **Event-Driven**: Streaming de eventos em tempo real via Kafka

### 8.2 Versionamento

- **Estratégia**: URL-based versioning (ex: /api/v1/audit/events)
- **Compatibilidade**: Suporte a duas versões ativas simultaneamente
- **Ciclo de Vida**: Depreciação com aviso de 6 meses, EOL após 1 ano

### 8.3 Documentação

- **Especificação**: OpenAPI 3.1
- **Portal de Desenvolvedores**: Documentação interativa com Swagger UI
- **Exemplos**: Bibliotecas de cliente em múltiplas linguagens (Java, Python, Go)

## 9. Observabilidade e Monitoramento

### 9.1 Métricas Chave

| Categoria | Métricas | Alerta |
|-----------|----------|--------|
| Disponibilidade | Uptime, Health checks | < 99.99% |
| Performance | Latência (p95, p99), Taxa de throughput | p95 > 200ms |
| Qualidade | Taxa de erros, Dropped events | > 0.1% |
| Compliance | Políticas violadas, Checks falhados | Qualquer falha |

### 9.2 Dashboards

- **Operacional**: Métricas de saúde do serviço em tempo real
- **Negócios**: KPIs de auditoria e compliance por tenant
- **Executivo**: Resumo consolidado de postura de segurança

### 9.3 Alertas

- **Severity Levels**: Critical, High, Medium, Low
- **Canais**: Slack, Email, SMS, PagerDuty
- **Automação**: Remediação automática para cenários conhecidos

## 10. Roadmap Tecnológico

### 10.1 Curto Prazo (Q3-Q4 2025)
- Implementação completa de observabilidade avançada
- Integração com OpenTelemetry para tracing distribuído
- Dashboard de compliance automatizado

### 10.2 Médio Prazo (Q1-Q2 2026)
- ML para detecção de anomalias em padrões de acesso
- Expansão de suporte regulatório para mercados emergentes
- Integração com blockchain para imutabilidade verificável

### 10.3 Longo Prazo (2026+)
- Análise preditiva de riscos baseada em comportamento histórico
- Remediação autônoma de violações de políticas
- Suporte a regulações emergentes de IA e ML

## 11. Conformidade com Padrões

### 11.1 Padrões de Arquitetura

| Padrão | Aplicação |
|--------|-----------|
| ISO/IEC 42010:2011 | Estruturação da documentação de arquitetura |
| TOGAF 9.2 | Framework de arquitetura empresarial |
| BIAN | Modelagem de serviços financeiros |

### 11.2 Padrões de Desenvolvimento

| Padrão | Aplicação |
|--------|-----------|
| SOLID | Princípios de design orientado a objetos |
| Clean Architecture | Separação de camadas e responsabilidades |
| 12-Factor App | Design para aplicações cloud-native |
| DDD | Domain-Driven Design para modelagem de domínio |

### 11.3 Padrões de Operação

| Padrão | Aplicação |
|--------|-----------|
| SRE | Princípios de Site Reliability Engineering |
| DevSecOps | Integração de segurança no pipeline DevOps |
| GitOps | Gerenciamento de configuração como código |
| ITIL 4 | Gerenciamento de serviços de TI |

## 12. Referências

1. Gartner Magic Quadrant for Privileged Access Management, 2025
2. Forrester Wave for Customer Identity and Access Management, 2025
3. NIST SP 800-53 Rev. 5 - Controles de Segurança para Sistemas Federais
4. ISO/IEC 27001:2022 - Sistema de Gestão de Segurança da Informação
5. PCI DSS v4.0 - Payment Card Industry Data Security Standard
6. OWASP API Security Project, Top 10 2025
7. INNOVABIZ Platform Global Standards Documentation