"""
Segunda parte do monitor de observabilidade para o sistema de regras dinâmicas.

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


class RulesObservabilityMonitorPart2:
    """
    Continuação das funções do monitor de observabilidade para o sistema de regras dinâmicas.
    
    Este código será integrado ao arquivo principal.
    """
    
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