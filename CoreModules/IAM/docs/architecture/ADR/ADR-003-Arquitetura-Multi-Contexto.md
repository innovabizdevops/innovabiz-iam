# ADR-003: Arquitetura Multi-Contexto (Multi-Tenant e Multi-Regional) para IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service precisa operar em um ambiente complexo e globalizado, atendendo a múltiplas dimensões de contexto:

- **Multi-tenant**: Suporte a múltiplas organizações isoladas logicamente
- **Multi-regional**: Operação em diferentes regiões geográficas globais
- **Multi-idioma**: Suporte a mais de 100 idiomas conforme requisitos INNOVABIZ
- **Multi-moeda**: Compatibilidade com +170 moedas internacionais
- **Multi-regulatório**: Conformidade com regulamentações específicas de cada jurisdição

Esta arquitetura deve garantir isolamento completo, performance consistente e conformidade regulatória em todos os contextos, além de integrar-se com o framework de observabilidade e monitoramento.

## Decisão

Implementar uma arquitetura Multi-Contexto baseada em **Tenant-Context-Aware Services** com as seguintes características:

1. **Modelo de Contexto Hierárquico**:
   - Tenant como nível primário de isolamento
   - Região como segundo nível de contexto
   - Ambiente como terceiro nível (prod, staging, dev)
   - Dimensões adicionais como idioma e moeda

2. **Propagação de Contexto**:
   - HTTP Headers padronizados (X-Tenant-ID, X-Region, X-Environment)
   - Variáveis de ambiente para configuração default (DEFAULT_TENANT, DEFAULT_REGION)
   - Context propagation automatizado entre serviços

3. **Isolamento de Dados**:
   - Implementação de Row-Level Security (RLS) no PostgreSQL
   - Partition sharding por tenant e região
   - Query middleware automático para aplicação de filtros de tenant

4. **Observabilidade Contextualizada**:
   - Labels Prometheus padronizados (tenant, region, environment)
   - Variáveis de template Grafana para filtros de contexto
   - Alertas configuráveis por tenant/região/ambiente
   - Métricas isoladas por dimensão de contexto

### Implementação Técnica

- **Middleware FastAPI** para extração e propagação automática de contexto
- **Middleware de observabilidade** para enriquecimento de métricas com dados de contexto
- **Decoradores Python** para funções que requerem awareness de contexto
- **Classes de abstração** para operações multi-tenant (TenantAwareRepository)
- **Sistema de configuração** com herança hierárquica baseada em contexto
- **PostgreSQL RLS** para isolamento a nível de banco de dados

## Alternativas Consideradas

### 1. Isolamento Físico por Tenant/Região

**Prós:**
- Isolamento máximo
- Performance dedicada
- Simplicidade de implementação

**Contras:**
- Custo proibitivo para escala
- Complexidade operacional
- Subutilização de recursos
- Dificuldade para manutenção e atualizações

### 2. Databases Separados com Aplicação Compartilhada

**Prós:**
- Bom isolamento de dados
- Menor complexidade de código
- Escala independente por tenant

**Contras:**
- Overhead de conexões
- Complexidade de manutenção de schemas
- Limitações para queries cross-tenant
- Maior footprint de infraestrutura

### 3. Sharding Manual por Tenant

**Prós:**
- Bom balanceamento de carga
- Isolamento para tenants grandes

**Contras:**
- Complexidade de implementação
- Dificuldade para rebalanceamento
- Overhead de gerenciamento
- Complexidade para agregações globais

## Consequências

### Positivas

- **Eficiência de recursos**: Aproveitamento máximo de infraestrutura compartilhada
- **Conformidade regulatória**: Capacidade de aplicar políticas específicas por jurisdição
- **Escalabilidade**: Adição de novos tenants e regiões sem redesenvolvimento
- **Observabilidade avançada**: Capacidade de monitorar e diagnosticar por contexto
- **Flexibilidade**: Capacidade de implementar diferentes políticas por tenant/região
- **Governança centralizada**: Visão unificada com capacidade de isolamento

### Negativas

- **Complexidade adicional**: Código mais complexo para gerenciar múltiplos contextos
- **Riscos de vazamento**: Necessidade de validação rigorosa para evitar vazamento cross-tenant
- **Sobrecarga de processamento**: Overhead para interpretação de contexto em cada requisição
- **Testes mais complexos**: Necessidade de testar múltiplas combinações de contexto

### Mitigação de Riscos

- Implementação de testes automatizados específicos para validação de isolamento
- Revisão de segurança dedicada para verificar isolamento multi-tenant
- Criação de middleware e decoradores padronizados para reduzir erros
- Monitoramento específico para detecção de acesso cross-tenant
- Auditoria periódica de configurações de isolamento de dados

## Conformidade com Standards

- **ISO/IEC 27001**: Controles para separação de ambientes e segregação de dados
- **PCI DSS 4.0**: Requisitos para isolamento de dados de cartão por entidade
- **GDPR/LGPD**: Separação de dados por controlador/processador
- **SOX**: Controles para segregação de acesso a dados financeiros
- **INNOVABIZ Platform Standards**: Compliance com requisitos de multi-tenancy

## Implementação na Observabilidade

A implementação na observabilidade inclui:

1. **Métricas Contextualizadas**:
   ```python
   audit_events_total = Counter(
       "audit_events_total",
       "Total de eventos de auditoria registrados",
       ["tenant", "region", "event_type", "severity"]
   )
   ```

2. **Health Checks por Contexto**:
   ```python
   @router.get("/health")
   async def health_check(request: Request):
       tenant = request.headers.get("X-Tenant-ID", "default")
       region = request.headers.get("X-Region", "global")
       # Verificações específicas de contexto
   ```

3. **Dashboards Grafana com Variáveis de Contexto**:
   - Variáveis para tenant, região e ambiente
   - Painéis filtráveis por múltiplas dimensões
   - Alertas configuráveis por contexto

4. **Propagação de Contexto em Alertas**:
   ```yaml
   - alert: HighAuditEventRate
     expr: sum(rate(audit_events_total{severity="high"}[5m])) by (tenant, region) > 100
     labels:
       tenant: '{{ $labels.tenant }}'
       region: '{{ $labels.region }}'
     annotations:
       summary: "Alta taxa de eventos de auditoria para {{ $labels.tenant }} em {{ $labels.region }}"
   ```

## Referências

1. INNOVABIZ Multi-Tenant Architecture Standards v3.1
2. Multi-Tenant Data Architecture Best Practices (Gartner, 2025)
3. Observability in Multi-Tenant Environments (O'Reilly, 2024)
4. PostgreSQL Row-Level Security Documentation - https://www.postgresql.org/docs/current/ddl-rowsecurity.html
5. FastAPI Dependency Injection for Tenant Context - https://fastapi.tiangolo.com/tutorial/dependencies/
6. Prometheus Multi-Tenant Design Patterns - https://prometheus.io/docs/practices/multi-tenant/