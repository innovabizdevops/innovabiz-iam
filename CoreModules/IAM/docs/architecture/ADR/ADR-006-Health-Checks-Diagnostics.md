# ADR-006: Health Checks e Endpoints de Diagnóstico para IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service, como componente crítico de segurança e compliance da plataforma INNOVABIZ, requer mecanismos robustos para verificação de saúde e diagnóstico operacional. Estes mecanismos são fundamentais para:

- Verificação de disponibilidade por load balancers e orquestradores (Kubernetes)
- Detecção precoce de problemas com dependências e serviços externos
- Facilitar troubleshooting e análise de problemas
- Suporte a cenários multi-tenant e multi-regional
- Conformidade com requisitos de observabilidade da plataforma
- Integração com sistemas de monitoramento e alertas
- Atendimento a requisitos regulatórios de disponibilidade e resiliência

É necessário definir padrões consistentes para health checks e diagnósticos que atendam estes requisitos enquanto mantêm a segurança e performance do serviço.

## Decisão

Implementar uma arquitetura padronizada de **Health Checks e Endpoints de Diagnóstico** para o IAM Audit Service com os seguintes componentes:

### 1. Endpoints de Health Check

#### 1.1. Endpoint `/health`

Endpoint primário de health check com verificações básicas e rápidas:

```python
@router.get("/health", response_model=HealthResponse)
async def health_check(
    request: Request,
    tenant: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region")
):
    """
    Health check básico para verificar disponibilidade do serviço.
    Executa verificações rápidas (<100ms) para componentes essenciais.
    """
    tenant = tenant or settings.DEFAULT_TENANT
    region = region or settings.DEFAULT_REGION
    
    checks = {
        "database": await check_database_connection(tenant, region),
        "cache": await check_cache_connection(tenant, region),
    }
    
    status = "healthy" if all(c["status"] == "up" for c in checks.values()) else "degraded"
    
    return {
        "status": status,
        "version": settings.SERVICE_VERSION,
        "timestamp": datetime.utcnow().isoformat(),
        "tenant": tenant,
        "region": region,
        "checks": checks
    }
```

#### 1.2. Endpoint `/ready`

Endpoint para verificações de prontidão para receber tráfego:

```python
@router.get("/ready", response_model=ReadinessResponse)
async def readiness_check(
    request: Request,
    tenant: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region")
):
    """
    Readiness check para verificar capacidade de processar requisições.
    Usado por orquestradores para decisões de roteamento.
    """
    tenant = tenant or settings.DEFAULT_TENANT
    region = region or settings.DEFAULT_REGION
    
    checks = {
        "database": await check_database_connection(tenant, region),
        "cache": await check_cache_connection(tenant, region),
        "storage": await check_storage_connection(tenant, region),
        "kafka": await check_kafka_connection(tenant, region)
    }
    
    status = "ready" if all(c["status"] == "up" for c in checks.values()) else "not_ready"
    
    return {
        "status": status,
        "timestamp": datetime.utcnow().isoformat(),
        "tenant": tenant,
        "region": region,
        "checks": checks
    }
```

#### 1.3. Endpoint `/live`

Endpoint minimalista para verificação de processo vivo:

```python
@router.get("/live")
async def liveness_check():
    """
    Liveness check simples para verificar se o processo está respondendo.
    Usado por orquestradores para detectar processos travados.
    """
    return {"status": "alive", "timestamp": datetime.utcnow().isoformat()}
```

### 2. Endpoint de Diagnóstico

Endpoint detalhado para diagnóstico operacional, com verificações mais profundas:

```python
@router.get("/diagnostic", response_model=DiagnosticResponse)
async def diagnostic(
    request: Request,
    include_deps: bool = Query(True),
    include_metrics: bool = Query(True),
    include_config: bool = Query(False),
    tenant: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region"),
    current_user: User = Depends(get_current_admin_user)  # Requer autenticação de admin
):
    """
    Diagnóstico detalhado do serviço com verificações profundas.
    Requer autenticação de administrador.
    """
    tenant = tenant or settings.DEFAULT_TENANT
    region = region or settings.DEFAULT_REGION
    
    result = {
        "service": {
            "name": "iam-audit-service",
            "version": settings.SERVICE_VERSION,
            "build_id": settings.BUILD_ID,
            "commit_hash": settings.COMMIT_HASH,
            "start_time": service_start_time.isoformat(),
            "uptime_seconds": (datetime.utcnow() - service_start_time).total_seconds()
        },
        "context": {
            "tenant": tenant,
            "region": region,
            "environment": settings.ENVIRONMENT
        },
        "timestamp": datetime.utcnow().isoformat()
    }
    
    if include_deps:
        result["dependencies"] = await check_all_dependencies(tenant, region)
    
    if include_metrics:
        result["metrics"] = await collect_service_metrics(tenant, region)
    
    if include_config:
        result["config"] = get_sanitized_config()  # Versão sanitizada da configuração
    
    return result
```

### 3. Verificadores de Dependências

Framework extensível para verificação de dependências:

```python
class HealthChecker:
    async def check_database(self, tenant: str, region: str) -> HealthCheckResult:
        """Verifica conexão e funcionalidade básica do banco de dados."""
        start_time = time.time()
        try:
            # Executa query simples para verificar conexão
            async with db_session() as session:
                result = await session.execute(text("SELECT 1"))
                assert result.scalar() == 1
                
            # Verifica permissões específicas do tenant
            await self._check_tenant_db_permissions(tenant, region)
            
            duration_ms = (time.time() - start_time) * 1000
            return {
                "status": "up",
                "latency_ms": duration_ms,
                "message": "Database connection successful"
            }
        except Exception as e:
            duration_ms = (time.time() - start_time) * 1000
            metrics.dependency_check_failed.labels(
                tenant=tenant, 
                region=region,
                dependency="database"
            ).inc()
            return {
                "status": "down",
                "latency_ms": duration_ms,
                "error": str(e),
                "message": "Database connection failed"
            }
    
    async def check_redis(self, tenant: str, region: str) -> HealthCheckResult:
        """Verifica conexão e funcionalidade do Redis."""
        # Implementação similar...
    
    async def check_kafka(self, tenant: str, region: str) -> HealthCheckResult:
        """Verifica conexão e funcionalidade do Kafka."""
        # Implementação similar...
    
    async def check_storage(self, tenant: str, region: str) -> HealthCheckResult:
        """Verifica conexão e funcionalidade do storage."""
        # Implementação similar...
```

### 4. Integração com Sistema de Observabilidade

Integração automática com métricas Prometheus:

```python
# Métricas para health checks
health_check_success = Counter(
    "iam_audit_health_check_success_total",
    "Total de health checks bem-sucedidos",
    ["tenant", "region", "check_type", "component"]
)

health_check_failure = Counter(
    "iam_audit_health_check_failure_total",
    "Total de health checks com falha",
    ["tenant", "region", "check_type", "component"]
)

health_check_latency = Histogram(
    "iam_audit_health_check_duration_seconds",
    "Duração dos health checks",
    ["tenant", "region", "check_type", "component"],
    buckets=[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]
)

# Decorator para instrumentação automática
def instrument_health_check(check_type: str, component: str):
    def decorator(func):
        @wraps(func)
        async def wrapper(self, tenant: str, region: str, *args, **kwargs):
            start_time = time.time()
            try:
                result = await func(self, tenant, region, *args, **kwargs)
                duration = time.time() - start_time
                
                if result["status"] == "up":
                    health_check_success.labels(
                        tenant=tenant,
                        region=region,
                        check_type=check_type,
                        component=component
                    ).inc()
                else:
                    health_check_failure.labels(
                        tenant=tenant,
                        region=region,
                        check_type=check_type,
                        component=component
                    ).inc()
                
                health_check_latency.labels(
                    tenant=tenant,
                    region=region,
                    check_type=check_type,
                    component=component
                ).observe(duration)
                
                return result
            except Exception as e:
                duration = time.time() - start_time
                health_check_failure.labels(
                    tenant=tenant,
                    region=region,
                    check_type=check_type,
                    component=component
                ).inc()
                health_check_latency.labels(
                    tenant=tenant,
                    region=region,
                    check_type=check_type,
                    component=component
                ).observe(duration)
                
                # Re-lança exceção para tratamento superior
                raise
        return wrapper
    return decorator
```

### 5. Modelos de Resposta

Modelos Pydantic para garantir respostas consistentes:

```python
class HealthCheckResult(BaseModel):
    status: str  # "up" ou "down"
    latency_ms: float
    message: str
    error: Optional[str] = None
    details: Optional[Dict[str, Any]] = None

class HealthResponse(BaseModel):
    status: str  # "healthy", "degraded", "unhealthy"
    version: str
    timestamp: str
    tenant: str
    region: str
    checks: Dict[str, HealthCheckResult]

class ReadinessResponse(BaseModel):
    status: str  # "ready", "not_ready"
    timestamp: str
    tenant: str
    region: str
    checks: Dict[str, HealthCheckResult]

class DiagnosticResponse(BaseModel):
    service: Dict[str, Any]
    context: Dict[str, str]
    timestamp: str
    dependencies: Optional[Dict[str, HealthCheckResult]] = None
    metrics: Optional[Dict[str, Any]] = None
    config: Optional[Dict[str, Any]] = None
```

## Alternativas Consideradas

### 1. Health Checks Minimalistas

**Prós:**
- Menor overhead
- Simplicidade de implementação
- Menos código para manter

**Contras:**
- Detecção tardia de problemas
- Visibilidade limitada para operações
- Troubleshooting mais difícil
- Não atende requisitos de compliance

### 2. Solução Proprietária de Health Checks

**Prós:**
- Recursos avançados out-of-the-box
- Dashboards e relatórios prontos

**Contras:**
- Dependência de terceiros
- Potencial vendor lock-in
- Necessidade de integração adicional
- Custos extras

### 3. Health Checks em Sistema Externo

**Prós:**
- Separação de concerns
- Menor impacto no serviço principal

**Contras:**
- Detecção mais lenta de problemas
- Menos informações de contexto
- Maior complexidade operacional
- Dificuldade para diagnósticos detalhados

## Consequências

### Positivas

- **Detecção precoce**: Identificação rápida de problemas antes de impacto em usuários
- **Diagnóstico facilitado**: Informações detalhadas para troubleshooting
- **Integração com orquestradores**: Suporte a decisões automáticas de roteamento e failover
- **Visibilidade contextualizada**: Health checks por tenant e região
- **Compliance**: Atendimento a requisitos de disponibilidade e resiliência
- **Consistência**: Padrão uniforme para health checks em toda a plataforma

### Negativas

- **Overhead operacional**: Impacto mínimo de performance para verificações constantes
- **Complexidade adicional**: Framework adicional para manter
- **Potencial para falsos positivos/negativos**: Necessidade de calibração cuidadosa
- **Exposição de informações sensíveis**: Necessidade de controle de acesso rigoroso

### Mitigação de Riscos

- Implementar rate limiting para health checks pesados
- Estabelecer timeouts adequados para prevenir bloqueios
- Criar diferentes níveis de verificação (light, standard, deep)
- Configurar autenticação para endpoints de diagnóstico
- Sanitizar informações sensíveis em resposta de diagnóstico
- Implementar circuit breakers para evitar sobrecarga de dependências

## Conformidade com Padrões

- **Kubernetes Probe Pattern**: Liveness, Readiness e Startup probes
- **ISO/IEC 25010**: Atributos de qualidade para disponibilidade e confiabilidade
- **Princípios SRE**: Monitoramento de saúde de serviço
- **PCI DSS 4.0**: Requisitos de disponibilidade e monitoramento (10.2, 11.4)
- **INNOVABIZ Platform Observability Standards v2.5**

## Implementação

A implementação inclui:

1. **Módulo `health.checkers`**:
   - Classes para verificação de dependências
   - Implementação de verificações específicas por componente
   - Instrumentação automática via decoradores

2. **Módulo `health.routers`**:
   - Endpoints FastAPI para health, readiness e liveness
   - Endpoint protegido para diagnóstico detalhado
   - Suporte a parâmetros de customização

3. **Módulo `health.models`**:
   - Modelos Pydantic para respostas padronizadas
   - Schemas para validação de inputs e outputs
   - Constantes para status e mensagens

4. **Integração com `ObservabilityIntegration`**:
   - Registro automático das rotas de health check
   - Coleta de métricas de health check
   - Alertas para falhas recorrentes

## Exemplos de Uso

### Exemplo 1: Verificação Básica de Health Check

```bash
$ curl https://iam-audit-service.innovabiz.com/health
{
  "status": "healthy",
  "version": "1.2.3",
  "timestamp": "2025-07-31T14:32:16Z",
  "tenant": "default",
  "region": "br-east",
  "checks": {
    "database": {
      "status": "up",
      "latency_ms": 23.5,
      "message": "Database connection successful"
    },
    "cache": {
      "status": "up",
      "latency_ms": 5.2,
      "message": "Cache connection successful"
    }
  }
}
```

### Exemplo 2: Verificação de Health Check Multi-Tenant

```bash
$ curl -H "X-Tenant-ID: client123" -H "X-Region: us-east" \
  https://iam-audit-service.innovabiz.com/health
{
  "status": "healthy",
  "version": "1.2.3",
  "timestamp": "2025-07-31T14:33:05Z",
  "tenant": "client123",
  "region": "us-east",
  "checks": {
    "database": {
      "status": "up",
      "latency_ms": 45.3,
      "message": "Database connection successful"
    },
    "cache": {
      "status": "up",
      "latency_ms": 12.8,
      "message": "Cache connection successful"
    }
  }
}
```

### Exemplo 3: Diagnóstico Detalhado

```bash
$ curl -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: client123" \
  "https://iam-audit-service.innovabiz.com/diagnostic?include_metrics=true"
{
  "service": {
    "name": "iam-audit-service",
    "version": "1.2.3",
    "build_id": "20250731-1420",
    "commit_hash": "a7c3e9b",
    "start_time": "2025-07-31T12:00:00Z",
    "uptime_seconds": 9000
  },
  "context": {
    "tenant": "client123",
    "region": "us-east",
    "environment": "production"
  },
  "timestamp": "2025-07-31T14:35:00Z",
  "dependencies": {
    "database": {
      "status": "up",
      "latency_ms": 42.1,
      "message": "Database connection successful",
      "details": {
        "version": "PostgreSQL 16.3",
        "connections": 5,
        "active_queries": 2
      }
    },
    "cache": {
      "status": "up",
      "latency_ms": 11.3,
      "message": "Cache connection successful",
      "details": {
        "version": "Redis 7.2",
        "used_memory": "256MB",
        "hit_ratio": 0.95
      }
    },
    "kafka": {
      "status": "up",
      "latency_ms": 37.8,
      "message": "Kafka connection successful",
      "details": {
        "version": "3.5.1",
        "topics": 5,
        "lag": 0
      }
    },
    "storage": {
      "status": "up",
      "latency_ms": 68.2,
      "message": "Storage connection successful"
    }
  },
  "metrics": {
    "audit_events_total": 1250367,
    "http_requests_total": 523678,
    "error_rate": 0.0023,
    "p99_latency_ms": 87.3
  }
}
```

## Referências

1. Kubernetes Probes - https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
2. FastAPI Health Check Patterns - https://fastapi.tiangolo.com/advanced/healthchecks/
3. Google SRE Book: Monitoring Distributed Systems - https://sre.google/sre-book/monitoring-distributed-systems/
4. Health Check Response Format for HTTP APIs - https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check
5. INNOVABIZ Platform Observability Standards v2.5 (Internal Document)
6. Site Reliability Engineering: Implementing Health Checks - https://docs.microsoft.com/en-us/azure/architecture/patterns/health-endpoint-monitoring