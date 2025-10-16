"""
Configuração de métricas para o serviço de auditoria do IAM.

Este módulo implementa instrumentação Prometheus para monitoramento
completo do serviço de auditoria, incluindo contadores, histogramas e
gauges para eventos, políticas de retenção e contexto multi-regional.
"""
from prometheus_client import Counter, Histogram, Gauge, Info
import time
from typing import Callable, Dict, Any, Optional
from fastapi import Request, Response
from functools import wraps
import structlog

logger = structlog.get_logger(__name__)

# Métricas de auditoria - eventos
AUDIT_EVENTS_TOTAL = Counter(
    "audit_events_total",
    "Total de eventos de auditoria processados",
    ["event_type", "tenant_id", "regional_context", "severity"]
)

AUDIT_EVENT_PROCESSING_DURATION = Histogram(
    "audit_event_processing_duration",
    "Duração do processamento de eventos de auditoria em segundos",
    ["tenant_id", "regional_context"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0)
)

AUDIT_EVENT_SIZE_BYTES = Histogram(
    "audit_event_size_bytes",
    "Tamanho dos eventos de auditoria em bytes",
    ["tenant_id", "regional_context"],
    buckets=(100, 500, 1000, 5000, 10000, 50000, 100000)
)

# Métricas de retenção
AUDIT_RETENTION_POLICIES_ACTIVE = Gauge(
    "audit_retention_policies_active",
    "Número de políticas de retenção ativas",
    ["tenant_id", "regional_context"]
)

AUDIT_RETENTION_EVENTS_PROCESSED_TOTAL = Counter(
    "audit_retention_events_processed_total",
    "Total de eventos processados por políticas de retenção",
    ["tenant_id", "regional_context", "policy_type"]
)

AUDIT_RETENTION_POLICY_EXECUTION_DURATION = Histogram(
    "audit_retention_policy_execution_duration",
    "Duração da execução de políticas de retenção em segundos",
    ["tenant_id", "regional_context", "policy_type"],
    buckets=(0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0, 600.0)
)

AUDIT_RETENTION_POLICY_SUCCESS = Counter(
    "audit_retention_policy_success",
    "Contagem de execuções bem-sucedidas de políticas de retenção",
    ["tenant_id", "regional_context", "policy_type"]
)

AUDIT_RETENTION_POLICY_FAILURE = Counter(
    "audit_retention_policy_failure",
    "Contagem de falhas na execução de políticas de retenção",
    ["tenant_id", "regional_context", "policy_type", "error_type"]
)
# Métricas de conformidade
AUDIT_COMPLIANCE_EVENTS_TOTAL = Counter(
    "audit_compliance_events_total",
    "Total de eventos de conformidade processados",
    ["tenant_id", "regional_context", "framework", "result"]
)

AUDIT_COMPLIANCE_CHECK_DURATION = Histogram(
    "audit_compliance_check_duration",
    "Duração das verificações de conformidade em segundos",
    ["tenant_id", "regional_context", "framework"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0)
)

AUDIT_REGIONAL_COMPLIANCE_STATUS = Gauge(
    "audit_regional_compliance_status",
    "Status de conformidade por região (1=conformidade, 0=não conformidade)",
    ["tenant_id", "regional_context", "framework", "regulation"]
)

# Métricas HTTP
HTTP_REQUESTS_TOTAL = Counter(
    "http_requests_total",
    "Total de requisições HTTP",
    ["method", "endpoint", "tenant_id", "regional_context"]
)

HTTP_REQUEST_DURATION_SECONDS = Histogram(
    "http_request_duration_seconds",
    "Duração das requisições HTTP em segundos",
    ["method", "endpoint", "tenant_id", "regional_context"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0)
)

HTTP_RESPONSE_SIZE_BYTES = Histogram(
    "http_response_size_bytes",
    "Tamanho das respostas HTTP em bytes",
    ["method", "endpoint", "tenant_id", "regional_context"],
    buckets=(100, 500, 1000, 5000, 10000, 50000, 100000)
)# Métricas de Status do Serviço
SERVICE_INFO = Info(
    "audit_service_info",
    "Informações sobre o serviço de auditoria"
)

SERVICE_UPTIME_SECONDS = Gauge(
    "audit_service_uptime_seconds", 
    "Tempo de atividade do serviço em segundos"
)

SERVICE_HEALTH_STATUS = Gauge(
    "audit_service_health_status",
    "Status de saúde do serviço (1=saudável, 0=não saudável)",
    ["component"]
)

# Middleware para métricas HTTP
async def metrics_middleware(request: Request, call_next) -> Response:
    """
    Middleware FastAPI para coletar métricas de requisições HTTP.
    
    Args:
        request: Objeto de requisição FastAPI
        call_next: Próxima função na cadeia de middleware
        
    Returns:
        Response: Objeto de resposta HTTP
    """
    # Extrair informações de contexto
    tenant_id = request.headers.get("X-Tenant-ID", "unknown")
    regional_context = request.headers.get("X-Regional-Context", "unknown")
    
    method = request.method
    endpoint = request.url.path
    
    # Registrar início da requisição
    start_time = time.time()
    
    # Processar requisição
    try:
        response = await call_next(request)
        status_code = response.status_code
    except Exception as e:
        # Registrar exceção
        logger.exception(f"Exceção ao processar requisição: {str(e)}")
        raise
    finally:
        # Calcular duração
        duration = time.time() - start_time
        
        # Incrementar contador de requisições
        HTTP_REQUESTS_TOTAL.labels(
            method=method,
            endpoint=endpoint,
            tenant_id=tenant_id,
            regional_context=regional_context
        ).inc()
        
        # Observar duração da requisição
        HTTP_REQUEST_DURATION_SECONDS.labels(
            method=method,
            endpoint=endpoint,
            tenant_id=tenant_id,
            regional_context=regional_context
        ).observe(duration)
        
    # Registrar tamanho da resposta se possível
    response_body_size = getattr(response, "body_size", 0)
    if response_body_size:
        HTTP_RESPONSE_SIZE_BYTES.labels(
            method=method,
            endpoint=endpoint,
            tenant_id=tenant_id,
            regional_context=regional_context
        ).observe(response_body_size)
    
    return response# Decoradores para instrumentação de métodos
def instrument_audit_event_processing(func: Callable) -> Callable:
    """
    Decorador para instrumentar o processamento de eventos de auditoria.
    
    Args:
        func: Função a ser instrumentada
        
    Returns:
        Callable: Função instrumentada
    """
    @wraps(func)
    async def wrapper(*args, **kwargs):
        # Extrair informações do contexto (assumindo padrão de argumentos nomeados)
        event_type = kwargs.get("event_type", "unknown")
        tenant_id = kwargs.get("tenant_id", "unknown")
        regional_context = kwargs.get("regional_context", "unknown")
        severity = kwargs.get("severity", "unknown")
        
        # Se os argumentos não forem nomeados, tentar extrair do primeiro argumento (evento)
        if args and hasattr(args[0], "get"):
            event = args[0]
            event_type = event_type if event_type != "unknown" else event.get("event_type", "unknown")
            tenant_id = tenant_id if tenant_id != "unknown" else event.get("tenant_id", "unknown")
            regional_context = regional_context if regional_context != "unknown" else event.get("regional_context", "unknown")
            severity = severity if severity != "unknown" else event.get("severity", "unknown")
        
        # Medir tamanho do evento em bytes (se possível)
        event_size = 0
        if args and hasattr(args[0], "__sizeof__"):
            event_size = args[0].__sizeof__()
        elif kwargs.get("event") and hasattr(kwargs["event"], "__sizeof__"):
            event_size = kwargs["event"].__sizeof__()
            
        # Registrar início do processamento
        start_time = time.time()
        
        try:
            # Executar função original
            result = await func(*args, **kwargs) if hasattr(func, "__await__") else func(*args, **kwargs)
            
            # Incrementar contador de eventos processados
            AUDIT_EVENTS_TOTAL.labels(
                event_type=event_type,
                tenant_id=tenant_id,
                regional_context=regional_context,
                severity=severity
            ).inc()
            
            # Registrar duração do processamento
            duration = time.time() - start_time
            AUDIT_EVENT_PROCESSING_DURATION.labels(
                tenant_id=tenant_id,
                regional_context=regional_context
            ).observe(duration)
            
            # Registrar tamanho do evento
            if event_size > 0:
                AUDIT_EVENT_SIZE_BYTES.labels(
                    tenant_id=tenant_id,
                    regional_context=regional_context
                ).observe(event_size)
            
            return result
        except Exception as e:
            # Registrar falha (poderia haver métricas específicas para falhas)
            logger.exception(f"Falha ao processar evento de auditoria: {str(e)}")
            raise
            
    return wrapperdef instrument_retention_policy(func: Callable) -> Callable:
    """
    Decorador para instrumentar a execução de políticas de retenção.
    
    Args:
        func: Função a ser instrumentada
        
    Returns:
        Callable: Função instrumentada
    """
    @wraps(func)
    async def wrapper(*args, **kwargs):
        # Extrair informações do contexto
        tenant_id = kwargs.get("tenant_id", "unknown")
        regional_context = kwargs.get("regional_context", "unknown")
        policy_type = kwargs.get("policy_type", "unknown")
        
        # Registrar início da execução
        start_time = time.time()
        
        try:
            # Executar função original
            result = await func(*args, **kwargs) if hasattr(func, "__await__") else func(*args, **kwargs)
            
            # Registrar sucesso
            AUDIT_RETENTION_POLICY_SUCCESS.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                policy_type=policy_type
            ).inc()
            
            # Atualizar contador de eventos processados
            if hasattr(result, "get"):
                processed_count = result.get("processed_count", 0)
                if processed_count > 0:
                    AUDIT_RETENTION_EVENTS_PROCESSED_TOTAL.labels(
                        tenant_id=tenant_id,
                        regional_context=regional_context,
                        policy_type=policy_type
                    ).inc(processed_count)
            
            # Registrar duração da execução
            duration = time.time() - start_time
            AUDIT_RETENTION_POLICY_EXECUTION_DURATION.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                policy_type=policy_type
            ).observe(duration)
            
            return result
        except Exception as e:
            # Registrar falha
            error_type = type(e).__name__
            AUDIT_RETENTION_POLICY_FAILURE.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                policy_type=policy_type,
                error_type=error_type
            ).inc()
            
            logger.exception(f"Falha ao executar política de retenção: {str(e)}")
            raise
            
    return wrapperdef instrument_compliance_check(func: Callable) -> Callable:
    """
    Decorador para instrumentar verificações de conformidade.
    
    Args:
        func: Função a ser instrumentada
        
    Returns:
        Callable: Função instrumentada
    """
    @wraps(func)
    async def wrapper(*args, **kwargs):
        # Extrair informações do contexto
        tenant_id = kwargs.get("tenant_id", "unknown")
        regional_context = kwargs.get("regional_context", "unknown")
        framework = kwargs.get("framework", "unknown")
        regulation = kwargs.get("regulation", "unknown")
        
        # Registrar início da verificação
        start_time = time.time()
        
        try:
            # Executar função original
            result = await func(*args, **kwargs) if hasattr(func, "__await__") else func(*args, **kwargs)
            
            # Determinar resultado da verificação
            compliance_result = "compliant"
            if isinstance(result, dict):
                compliance_result = "compliant" if result.get("compliant", False) else "non_compliant"
            elif isinstance(result, bool):
                compliance_result = "compliant" if result else "non_compliant"
            
            # Incrementar contador de eventos de conformidade
            AUDIT_COMPLIANCE_EVENTS_TOTAL.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                framework=framework,
                result=compliance_result
            ).inc()
            
            # Atualizar status de conformidade regional
            status_value = 1.0 if compliance_result == "compliant" else 0.0
            AUDIT_REGIONAL_COMPLIANCE_STATUS.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                framework=framework,
                regulation=regulation
            ).set(status_value)
            
            # Registrar duração da verificação
            duration = time.time() - start_time
            AUDIT_COMPLIANCE_CHECK_DURATION.labels(
                tenant_id=tenant_id,
                regional_context=regional_context,
                framework=framework
            ).observe(duration)
            
            return result
        except Exception as e:
            # Registrar falha (poderia haver métricas específicas para falhas)
            logger.exception(f"Falha na verificação de conformidade: {str(e)}")
            raise
            
    return wrapper# Funções de utilidade para inicialização e configuração

def setup_service_info(
    version: str, 
    build_id: str, 
    commit_hash: str, 
    environment: str,
    region: str
) -> None:
    """
    Configura informações estáticas do serviço.
    
    Args:
        version: Versão do serviço
        build_id: ID da compilação
        commit_hash: Hash do commit
        environment: Ambiente (prod, staging, qa, dev)
        region: Região de implantação
    """
    SERVICE_INFO.info({
        "version": version,
        "build_id": build_id,
        "commit_hash": commit_hash,
        "environment": environment,
        "region": region,
        "service": "iam-audit"
    })


def update_service_health(component: str, is_healthy: bool) -> None:
    """
    Atualiza o status de saúde de um componente do serviço.
    
    Args:
        component: Nome do componente (db, cache, kafka, etc.)
        is_healthy: Status de saúde (True=saudável, False=não saudável)
    """
    value = 1.0 if is_healthy else 0.0
    SERVICE_HEALTH_STATUS.labels(component=component).set(value)


def register_retention_policies(
    tenant_id: str, 
    regional_context: str, 
    active_count: int
) -> None:
    """
    Registra o número de políticas de retenção ativas.
    
    Args:
        tenant_id: ID do tenant
        regional_context: Contexto regional
        active_count: Número de políticas ativas
    """
    AUDIT_RETENTION_POLICIES_ACTIVE.labels(
        tenant_id=tenant_id,
        regional_context=regional_context
    ).set(active_count)


def start_uptime_counter() -> None:
    """
    Inicia a contagem de uptime do serviço. Esta função deve ser chamada
    no início da aplicação.
    """
    def uptime_tracker():
        start = time.time()
        while True:
            SERVICE_UPTIME_SECONDS.set(time.time() - start)
            time.sleep(1)
    
    import threading
    threading.Thread(target=uptime_tracker, daemon=True).start()


def init_metrics(app) -> None:
    """
    Inicializa métricas para o aplicativo FastAPI.
    
    Args:
        app: Instância do aplicativo FastAPI
    """
    from prometheus_client import make_asgi_app
    
    # Criar endpoint Prometheus
    metrics_app = make_asgi_app()
    app.mount("/metrics", metrics_app)
    
    # Registrar middleware de métricas
    app.middleware("http")(metrics_middleware)
    
    # Iniciar contador de uptime
    start_uptime_counter()
    
    logger.info("Métricas Prometheus inicializadas com sucesso")