"""
Monitor de observabilidade para o sistema de regras dinâmicas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/Observabilidade
Data: 21/08/2025
"""

import json
import logging
import time
from datetime import datetime
from functools import wraps
from typing import Any, Dict, List, Optional, Union, Callable

from opentelemetry import trace, metrics
from opentelemetry.trace import Span, SpanKind
from opentelemetry.metrics import Meter, Counter, UpDownCounter, Histogram

from .rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator


class RulesObservabilityMonitor:
    """
    Monitor de observabilidade para o sistema de regras dinâmicas.
    
    Provê monitoramento e observabilidade para:
    1. Avaliação de regras
    2. Integração com Bureau de Créditos
    3. Integração com TrustGuard
    4. Desempenho e comportamento do sistema
    """
    
    def __init__(
        self,
        configurator: RulesObservabilityConfigurator,
        component_name: str = "rules-engine",
    ):
        """
        Inicializa o monitor de observabilidade.
        
        Args:
            configurator: Configurador de observabilidade
            component_name: Nome do componente
        """
        self.configurator = configurator
        self.component_name = component_name
        
        # Configurar observabilidade
        self.logger = configurator.setup_logger()
        self.tracer = configurator.setup_tracer()
        self.meter = configurator.setup_meter()
        
        # Configurar métricas
        self._setup_metrics()
    
    def _setup_metrics(self):
        """Configura métricas para monitoramento."""
        # Contadores para regras
        self.rule_evaluation_counter = self.meter.create_counter(
            name="rules.evaluation.count",
            description="Número de avaliações de regras",
            unit="1",
        )
        
        self.rule_triggered_counter = self.meter.create_counter(
            name="rules.triggered.count",
            description="Número de regras acionadas",
            unit="1",
        )
        
        # Histograma para tempo de avaliação
        self.rule_evaluation_time = self.meter.create_histogram(
            name="rules.evaluation.duration",
            description="Duração da avaliação de regras",
            unit="ms",
        )
        
        # Contador para chamadas ao Bureau de Créditos
        self.bureau_request_counter = self.meter.create_counter(
            name="bureau.request.count",
            description="Número de requisições ao Bureau de Créditos",
            unit="1",
        )
        
        self.bureau_error_counter = self.meter.create_counter(
            name="bureau.error.count",
            description="Número de erros em requisições ao Bureau de Créditos",
            unit="1",
        )
        
        # Histograma para tempo de resposta do Bureau
        self.bureau_response_time = self.meter.create_histogram(
            name="bureau.response.duration",
            description="Duração da resposta do Bureau de Créditos",
            unit="ms",
        )
        
        # Contadores para TrustGuard
        self.trustguard_request_counter = self.meter.create_counter(
            name="trustguard.request.count",
            description="Número de requisições ao TrustGuard",
            unit="1",
        )
        
        self.trustguard_error_counter = self.meter.create_counter(
            name="trustguard.error.count",
            description="Número de erros em requisições ao TrustGuard",
            unit="1",
        )
        
        # Histograma para tempo de resposta do TrustGuard
        self.trustguard_response_time = self.meter.create_histogram(
            name="trustguard.response.duration",
            description="Duração da resposta do TrustGuard",
            unit="ms",
        )
        
        # Contador para decisões de acesso
        self.access_decision_counter = self.meter.create_counter(
            name="access.decision.count",
            description="Número de decisões de acesso",
            unit="1",
        )
        
        # Contadores para cache
        self.cache_hit_counter = self.meter.create_counter(
            name="cache.hit.count",
            description="Número de acertos de cache",
            unit="1",
        )
        
        self.cache_miss_counter = self.meter.create_counter(
            name="cache.miss.count",
            description="Número de erros de cache",
            unit="1",
        )
    
    def trace_rule_evaluation(self, func):
        """
        Decorador para instrumentar avaliação de regras.
        
        Args:
            func: Função a ser instrumentada
            
        Returns:
            Callable: Função decorada
        """
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Extrair informações para telemetria
            ruleset_id = None
            event_type = None
            num_rules = 0
            
            # Extrair ruleset_id dos argumentos
            if len(args) > 1 and hasattr(args[1], "id"):
                ruleset_id = args[1].id
                if hasattr(args[1], "rules"):
                    num_rules = len(args[1].rules)
            
            # Extrair event_type dos argumentos
            if len(args) > 2 and hasattr(args[2], "type"):
                event_type = args[2].type
            
            # Criar span para tracing
            with self.tracer.start_as_current_span(
                name=f"{self.component_name}.evaluate_rules",
                kind=SpanKind.INTERNAL,
                attributes={
                    "ruleset.id": ruleset_id or "unknown",
                    "ruleset.num_rules": num_rules,
                    "event.type": event_type or "unknown",
                },
            ) as span:
                try:
                    # Executar função original
                    result = await func(*args, **kwargs)
                    
                    # Calcular métricas
                    elapsed_time = (time.time() - start_time) * 1000  # ms
                    
                    # Registrar métricas
                    self.rule_evaluation_counter.add(1, {"ruleset_id": ruleset_id or "unknown"})
                    self.rule_evaluation_time.record(elapsed_time, {"ruleset_id": ruleset_id or "unknown"})
                    
                    # Registrar regras acionadas
                    triggered_rules = sum(1 for r in result.values() if r.triggered)
                    self.rule_triggered_counter.add(triggered_rules, {"ruleset_id": ruleset_id or "unknown"})
                    
                    # Adicionar atributos ao span
                    span.set_attribute("rules.total", num_rules)
                    span.set_attribute("rules.triggered", triggered_rules)
                    span.set_attribute("rules.evaluation.duration_ms", elapsed_time)
                    
                    # Log
                    self.logger.info(
                        f"Avaliação de regras completada: ruleset_id={ruleset_id}, "
                        f"event_type={event_type}, regras_acionadas={triggered_rules}/{num_rules}, "
                        f"tempo={elapsed_time:.2f}ms"
                    )
                    
                    return result
                    
                except Exception as e:
                    # Registrar erro
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", str(e))
                    
                    self.logger.error(
                        f"Erro na avaliação de regras: ruleset_id={ruleset_id}, "
                        f"event_type={event_type}, erro={str(e)}"
                    )
                    
                    # Re-lançar a exceção
                    raise
        
        return wrapper
    
    def trace_bureau_request(self, func):
        """
        Decorador para instrumentar requisições ao Bureau de Créditos.
        
        Args:
            func: Função a ser instrumentada
            
        Returns:
            Callable: Função decorada
        """
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Extrair informações para telemetria
            provider = kwargs.get("provider", "unknown")
            data_type = kwargs.get("data_type", "unknown")
            document_id = kwargs.get("document_id", "unknown")
            
            # Criar span para tracing
            with self.tracer.start_as_current_span(
                name=f"{self.component_name}.bureau_request",
                kind=SpanKind.CLIENT,
                attributes={
                    "bureau.provider": provider,
                    "bureau.data_type": data_type,
                    "bureau.document_id": document_id,
                },
            ) as span:
                try:
                    # Executar função original
                    result = await func(*args, **kwargs)
                    
                    # Calcular métricas
                    elapsed_time = (time.time() - start_time) * 1000  # ms
                    
                    # Registrar métricas
                    self.bureau_request_counter.add(1, {
                        "provider": provider,
                        "data_type": data_type,
                    })
                    
                    self.bureau_response_time.record(elapsed_time, {
                        "provider": provider,
                        "data_type": data_type,
                    })
                    
                    # Adicionar atributos ao span
                    span.set_attribute("bureau.response.duration_ms", elapsed_time)
                    span.set_attribute("bureau.response.status", "success")
                    
                    # Log
                    self.logger.info(
                        f"Requisição ao Bureau completada: provider={provider}, "
                        f"data_type={data_type}, document_id={document_id}, "
                        f"tempo={elapsed_time:.2f}ms"
                    )
                    
                    return result
                    
                except Exception as e:
                    # Registrar erro
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", str(e))
                    span.set_attribute("bureau.response.status", "error")
                    
                    self.bureau_error_counter.add(1, {
                        "provider": provider,
                        "data_type": data_type,
                        "error_type": type(e).__name__,
                    })
                    
                    self.logger.error(
                        f"Erro na requisição ao Bureau: provider={provider}, "
                        f"data_type={data_type}, document_id={document_id}, erro={str(e)}"
                    )
                    
                    # Re-lançar a exceção
                    raise
        
        return wrapper
    
    def trace_trustguard_request(self, func):
        """
        Decorador para instrumentar requisições ao TrustGuard.
        
        Args:
            func: Função a ser instrumentada
            
        Returns:
            Callable: Função decorada
        """
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Extrair informações para telemetria
            endpoint = None
            user_id = None
            
            # Tentar extrair endpoint da chamada
            if func.__name__ == "_make_request" and len(args) > 2:
                endpoint = args[2]
            elif "endpoint" in kwargs:
                endpoint = kwargs["endpoint"]
            
            # Tentar extrair user_id
            for arg in args:
                if hasattr(arg, "user_id"):
                    user_id = arg.user_id
                    break
            
            if user_id is None and "user_id" in kwargs:
                user_id = kwargs["user_id"]
            
            # Criar span para tracing
            with self.tracer.start_as_current_span(
                name=f"{self.component_name}.trustguard_request",
                kind=SpanKind.CLIENT,
                attributes={
                    "trustguard.endpoint": endpoint or "unknown",
                    "trustguard.user_id": user_id or "unknown",
                },
            ) as span:
                try:
                    # Executar função original
                    result = await func(*args, **kwargs)
                    
                    # Calcular métricas
                    elapsed_time = (time.time() - start_time) * 1000  # ms
                    
                    # Registrar métricas
                    self.trustguard_request_counter.add(1, {
                        "endpoint": endpoint or "unknown",
                        "method": func.__name__,
                    })
                    
                    self.trustguard_response_time.record(elapsed_time, {
                        "endpoint": endpoint or "unknown",
                        "method": func.__name__,
                    })
                    
                    # Adicionar atributos ao span
                    span.set_attribute("trustguard.response.duration_ms", elapsed_time)
                    span.set_attribute("trustguard.response.status", "success")
                    
                    # Log
                    self.logger.info(
                        f"Requisição ao TrustGuard completada: endpoint={endpoint}, "
                        f"method={func.__name__}, user_id={user_id}, "
                        f"tempo={elapsed_time:.2f}ms"
                    )
                    
                    return result
                    
                except Exception as e:
                    # Registrar erro
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", str(e))
                    span.set_attribute("trustguard.response.status", "error")
                    
                    self.trustguard_error_counter.add(1, {
                        "endpoint": endpoint or "unknown",
                        "method": func.__name__,
                        "error_type": type(e).__name__,
                    })
                    
                    self.logger.error(
                        f"Erro na requisição ao TrustGuard: endpoint={endpoint}, "
                        f"method={func.__name__}, user_id={user_id}, erro={str(e)}"
                    )
                    
                    # Re-lançar a exceção
                    raise
        
        return wrapper
    
    def trace_access_evaluation(self, func):
        """
        Decorador para instrumentar avaliações de acesso.
        
        Args:
            func: Função a ser instrumentada
            
        Returns:
            Callable: Função decorada
        """
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Extrair informações para telemetria
            request_id = None
            user_id = None
            resource_id = None
            resource_type = None
            action = None
            
            # Extrair informações da solicitação de acesso
            request = None
            if len(args) > 1 and hasattr(args[1], "request_id"):
                request = args[1]
            elif "request" in kwargs and hasattr(kwargs["request"], "request_id"):
                request = kwargs["request"]
            
            if request:
                request_id = request.request_id
                
                if hasattr(request, "user_context") and request.user_context:
                    user_id = request.user_context.user_id
                
                if hasattr(request, "resource_context") and request.resource_context:
                    resource_id = request.resource_context.resource_id
                    resource_type = request.resource_context.resource_type
                    action = request.resource_context.action
            
            # Criar span para tracing
            with self.tracer.start_as_current_span(
                name=f"{self.component_name}.evaluate_access",
                kind=SpanKind.INTERNAL,
                attributes={
                    "access.request_id": request_id or "unknown",
                    "access.user_id": user_id or "unknown",
                    "access.resource_id": resource_id or "unknown",
                    "access.resource_type": str(resource_type) if resource_type else "unknown",
                    "access.action": str(action) if action else "unknown",
                },
            ) as span:
                try:
                    # Executar função original
                    result = await func(*args, **kwargs)
                    
                    # Calcular métricas
                    elapsed_time = (time.time() - start_time) * 1000  # ms
                    
                    # Registrar métricas
                    self.access_decision_counter.add(1, {
                        "decision": str(result.decision),
                        "risk_level": str(result.risk_level),
                        "resource_type": str(resource_type) if resource_type else "unknown",
                        "action": str(action) if action else "unknown",
                    })
                    
                    self.trustguard_response_time.record(elapsed_time, {
                        "method": "evaluate_access",
                    })
                    
                    # Adicionar atributos ao span
                    span.set_attribute("access.decision", str(result.decision))
                    span.set_attribute("access.risk_level", str(result.risk_level))
                    span.set_attribute("access.response.duration_ms", elapsed_time)
                    
                    # Log
                    self.logger.info(
                        f"Avaliação de acesso completada: request_id={request_id}, "
                        f"user_id={user_id}, resource_id={resource_id}, "
                        f"decision={result.decision}, risk_level={result.risk_level}, "
                        f"tempo={elapsed_time:.2f}ms"
                    )
                    
                    return result
                    
                except Exception as e:
                    # Registrar erro
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", str(e))
                    
                    self.trustguard_error_counter.add(1, {
                        "method": "evaluate_access",
                        "error_type": type(e).__name__,
                    })
                    
                    self.logger.error(
                        f"Erro na avaliação de acesso: request_id={request_id}, "
                        f"user_id={user_id}, resource_id={resource_id}, erro={str(e)}"
                    )
                    
                    # Re-lançar a exceção
                    raise
        
        return wrapper
    
    def trace_cache_operation(self, func):
        """
        Decorador para instrumentar operações de cache.
        
        Args:
            func: Função a ser instrumentada
            
        Returns:
            Callable: Função decorada
        """
        @wraps(func)
        def wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Extrair informações para telemetria
            cache_key = None
            
            # Extrair chave de cache
            if len(args) > 1:
                cache_key = args[1]
            elif "key" in kwargs:
                cache_key = kwargs["key"]
            
            # Criar span para tracing
            with self.tracer.start_as_current_span(
                name=f"{self.component_name}.cache_operation",
                kind=SpanKind.INTERNAL,
                attributes={
                    "cache.key": str(cache_key),
                    "cache.operation": func.__name__,
                },
            ) as span:
                try:
                    # Executar função original
                    result = func(*args, **kwargs)
                    
                    # Calcular métricas
                    elapsed_time = (time.time() - start_time) * 1000  # ms
                    
                    # Para _get_from_cache, registrar hit/miss
                    if func.__name__ == "_get_from_cache":
                        hit, _ = result
                        if hit:
                            self.cache_hit_counter.add(1)
                            span.set_attribute("cache.hit", True)
                        else:
                            self.cache_miss_counter.add(1)
                            span.set_attribute("cache.hit", False)
                    
                    # Adicionar atributos ao span
                    span.set_attribute("cache.operation.duration_ms", elapsed_time)
                    
                    return result
                    
                except Exception as e:
                    # Registrar erro
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", str(e))
                    
                    self.logger.error(
                        f"Erro na operação de cache: operation={func.__name__}, "
                        f"key={cache_key}, erro={str(e)}"
                    )
                    
                    # Re-lançar a exceção
                    raise
        
        return wrapper
    
    def record_metric(self, name: str, value: float, attributes: Dict[str, str] = None):
        """
        Registra uma métrica personalizada.
        
        Args:
            name: Nome da métrica
            value: Valor da métrica
            attributes: Atributos para a métrica
        """
        attributes = attributes or {}
        
        # Verificar se a métrica existe
        if not hasattr(self, name):
            # Criar uma nova métrica
            self.logger.warning(f"Métrica {name} não encontrada, criando nova")
            setattr(self, name, self.meter.create_counter(
                name=name,
                description=f"Métrica personalizada: {name}",
                unit="1",
            ))
        
        # Registrar o valor
        getattr(self, name).add(value, attributes)
    
    def start_span(self, name: str, attributes: Dict[str, str] = None):
        """
        Inicia um span personalizado.
        
        Args:
            name: Nome do span
            attributes: Atributos para o span
            
        Returns:
            Span: Span criado
        """
        attributes = attributes or {}
        return self.tracer.start_as_current_span(
            name=f"{self.component_name}.{name}",
            kind=SpanKind.INTERNAL,
            attributes=attributes,
        )


def instrument_connector(connector, monitor):
    """
    Instrumenta um conector com monitoramento e observabilidade.
    
    Args:
        connector: Conector a ser instrumentado
        monitor: Monitor de observabilidade
    """
    # Instrumentar métodos do TrustGuard
    connector._make_request = monitor.trace_trustguard_request(connector._make_request)
    connector.evaluate_access = monitor.trace_access_evaluation(connector.evaluate_access)
    connector._get_from_cache = monitor.trace_cache_operation(connector._get_from_cache)
    connector._set_cache = monitor.trace_cache_operation(connector._set_cache)
    
    return connector