# ADR-005: Framework de Decoradores e Middleware para Instrumentação Automática

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service requer instrumentação consistente e completa para garantir observabilidade de alta qualidade. Entretanto, a instrumentação manual pode ser:

- Propensa a erros e inconsistências
- Trabalhosa e repetitiva
- Sujeita a esquecimentos em pontos críticos
- Difícil de manter de forma padronizada
- Desafiadora para garantir cobertura completa

É necessária uma abordagem que torne a instrumentação automática, padronizada e de baixo impacto para os desenvolvedores, garantindo consistência nas métricas coletadas em todo o serviço de auditoria.

## Decisão

Implementar um **Framework de Decoradores e Middleware** para instrumentação automática do IAM Audit Service com as seguintes características:

1. **Decoradores Python** para instrumentação de funções:
   - Medição automática de latência
   - Contagem de chamadas e erros
   - Enriquecimento com contexto (tenant, região)
   - Suporte a métricas específicas de domínio

2. **Middleware FastAPI** para instrumentação de endpoints HTTP:
   - Métricas padrão RED (Request Rate, Error Rate, Duration)
   - Contexto automático extraído de headers
   - Categorização por path, método e status
   - Middleware modular e extensível

3. **Configuração centralizada** com capacidade de:
   - Ativar/desativar componentes específicos
   - Customizar buckets de histogramas
   - Definir labels padrão para contexto
   - Configurar comportamentos específicos

### Componentes Principais

#### 1. Decoradores de Função

```python
# Decorador para instrumentar eventos de auditoria
@metrics.audit_event(event_type="user_access")
async def process_user_access(user_id: str, resource: str):
    # Lógica de processamento...
    pass

# Decorador para instrumentar operações de retenção
@metrics.retention_operation(policy_type="gdpr_compliance")
async def execute_retention_policy(tenant_id: str):
    # Lógica de retenção...
    pass

# Decorador para instrumentar verificações de compliance
@metrics.compliance_check(compliance_type="pci_dss_4")
async def verify_compliance(audit_records: List[AuditRecord]):
    # Lógica de verificação...
    pass

# Decorador genérico com timing automático
@metrics.timed(name="custom_operation", labels={"criticality": "high"})
async def perform_critical_operation():
    # Operação crítica...
    pass
```

#### 2. Middleware HTTP

```python
# Configuração de middleware na aplicação FastAPI
app = FastAPI()
observability = ObservabilityIntegration()

# Middleware de métricas HTTP
app.add_middleware(
    metrics.HTTPMetricsMiddleware,
    exclude_paths=["/metrics", "/health"],
    label_extractors={
        "tenant": lambda request: request.headers.get("X-Tenant-ID", "default"),
        "region": lambda request: request.headers.get("X-Region", "global"),
    }
)

# Middleware de contexto
app.add_middleware(
    metrics.ContextMiddleware,
    header_mapping={
        "X-Tenant-ID": "tenant",
        "X-Region": "region",
        "X-Environment": "environment"
    }
)

# Exposição de endpoint de métricas
app.include_router(observability.router)
```

#### 3. Configuração e Integração

```python
# Configuração centralizada via ObservabilityConfig
config = ObservabilityConfig(
    service_name="iam-audit-service",
    namespace="innovabiz",
    default_labels={
        "version": "1.0.0",
        "environment": "production"
    },
    histogram_buckets={
        "http_request_duration_seconds": [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0],
        "audit_event_processing_seconds": [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5]
    }
)

# Integração simplificada com aplicação
observability = ObservabilityIntegration(config)
observability.instrument_app(app)
```

### Extensibilidade e Customização

O framework deve permitir extensões e customizações para atender requisitos específicos:

```python
# Criação de decoradores customizados
custom_metric = metrics.create_counter(
    "iam_audit_custom_operation_total",
    "Contador para operações customizadas",
    ["tenant", "region", "operation_type"]
)

# Registro de métricas customizadas
@app.get("/custom-endpoint")
async def custom_endpoint(request: Request):
    tenant = request.state.context.tenant
    region = request.state.context.region
    
    # Operação de negócio
    result = await business_logic()
    
    # Registro manual de métrica
    custom_metric.labels(tenant=tenant, region=region, operation_type="special").inc()
    return result
```

## Alternativas Consideradas

### 1. Instrumentação Puramente Manual

**Prós:**
- Controle total sobre cada ponto de instrumentação
- Sem overhead de framework

**Contras:**
- Alta probabilidade de inconsistências
- Código repetitivo e boilerplate
- Risco de cobertura incompleta
- Manutenção mais complexa

### 2. Aspectos com AspectPython ou frameworks similares

**Prós:**
- Separação completa de concerns
- Flexibilidade total de instrumentação

**Contras:**
- Maior complexidade conceptual
- Potenciais desafios de performance
- Curva de aprendizado mais íngreme
- Menor previsibilidade de comportamento

### 3. Monkeypatching Global de Funções

**Prós:**
- Instrumentação transparente
- Sem alterações no código de negócio

**Contras:**
- Difícil manutenção
- Potenciais conflitos com outras bibliotecas
- Difícil customização por endpoint
- Problemas com atualização de bibliotecas

## Consequências

### Positivas

- **Consistência**: Instrumentação padronizada em todo o serviço
- **Cobertura completa**: Garantia de que todos os endpoints e operações críticas são monitorados
- **Produtividade**: Redução significativa de código boilerplate
- **Manutenibilidade**: Centralização da lógica de instrumentação
- **Flexibilidade**: Suporte para customização quando necessário
- **Context-awareness**: Propagação automática de contexto multi-tenant e multi-regional

### Negativas

- **Overhead inicial**: Pequeno impacto de performance pela camada adicional de decoradores
- **Abstração**: Potencial "magia" que pode confundir novos desenvolvedores
- **Complexidade adicional**: Necessidade de entender o framework de decoradores
- **Risco de shadow instrumentation**: Potencial para instrumentação invisível ou inesperada

### Mitigação de Riscos

- Implementar testes de performance para validar o overhead dos decoradores
- Criar documentação clara sobre o uso do framework
- Estabelecer convenções de nomeação consistentes
- Fornecer exemplos e templates para casos comuns
- Implementar validação em CI para garantir uso correto dos decoradores
- Manter logs detalhados durante o processo de inicialização para transparência

## Conformidade com Padrões

- **SOLID Principles**: Especialmente Single Responsibility e Open/Closed
- **DRY (Don't Repeat Yourself)**: Eliminação de código repetitivo
- **Python Decorator Best Practices**: Uso idiomático de decoradores
- **FastAPI Middleware Standards**: Integração correta com o ciclo de vida do framework
- **INNOVABIZ Platform Standards v3.2**: Alinhamento com padrões da plataforma

## Implementação

A implementação inclui:

1. **Módulo `observability.decorators`**:
   - Decoradores especializados para diferentes tipos de operações
   - Factory functions para criação de decoradores customizados
   - Utilitários para extração de contexto

2. **Módulo `observability.middleware`**:
   - Middleware para métricas HTTP
   - Middleware para propagação de contexto
   - Middleware para verificação de health

3. **Módulo `observability.config`**:
   - Classes de configuração baseadas em Pydantic
   - Funcionalidades de override via environment variables
   - Validação e defaults sensíveis

4. **Classe `ObservabilityIntegration`**:
   - Interface unificada para integração com aplicações
   - Gerenciamento de ciclo de vida de métricas
   - Configuração de endpoints padrão

## Exemplos de Uso

### Exemplo 1: Instrumentação de Endpoint de Eventos de Auditoria

```python
@router.post("/events")
@metrics.audit_event(event_type="generic")
async def create_audit_event(
    event: AuditEventCreate,
    request: Request,
    db: AsyncSession = Depends(get_db)
):
    tenant = request.state.context.tenant
    region = request.state.context.region
    
    try:
        result = await audit_service.create_event(event, tenant, region, db)
        return result
    except ValidationError as e:
        # Erro é automaticamente capturado pelo decorador
        # e registrado como métrica de erro
        raise HTTPException(status_code=400, detail=str(e))
```

### Exemplo 2: Instrumentação de Job de Retenção

```python
@metrics.retention_operation(policy_type="standard")
async def execute_retention_job(tenant_id: str, region: str):
    """
    Executa o job de retenção para um tenant específico.
    Métricas serão coletadas automaticamente:
    - iam_audit_retention_execution_seconds (duração)
    - iam_audit_retention_purge_total (contagem de registros)
    - iam_audit_retention_execution_total (contagem de execuções)
    """
    records = await retention_service.apply_policies(tenant_id, region)
    return records
```

## Referências

1. Python Decorators Documentation - https://docs.python.org/3/glossary.html#term-decorator
2. FastAPI Middleware - https://fastapi.tiangolo.com/tutorial/middleware/
3. Prometheus Python Client - https://github.com/prometheus/client_python
4. RED Method for Monitoring Microservices - https://www.weave.works/blog/the-red-method-key-metrics-for-microservices
5. INNOVABIZ Observability Standards v2.5 (Internal Document)
6. Aspect-Oriented Programming Concepts - https://en.wikipedia.org/wiki/Aspect-oriented_programming