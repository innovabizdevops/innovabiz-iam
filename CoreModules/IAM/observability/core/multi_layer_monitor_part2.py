"""
Monitor de Observabilidade Multi-Camada para a plataforma INNOVABIZ - Parte 2.

Implementação da classe principal do monitor de observabilidade.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - Sistema de Observabilidade Multi-Camada
Data: 20/08/2025
"""

import os
import time
import json
import inspect
import functools
import traceback
import asyncio
from typing import Dict, List, Optional, Union, Any, Callable, TypeVar, cast
from datetime import datetime
from enum import Enum

from .multi_layer_config import (
    MultiLayerObservabilityConfig,
    MultiLayerObservabilityConfigurator,
    ObservabilityLevel
)
from .multi_layer_monitor import (
    F, AsyncF, EventContext, EventSeverity, EventCategory,
    METRIC_PREFIX, APP_REQUEST_COUNTER, APP_REQUEST_DURATION, APP_ERROR_COUNTER,
    BUSINESS_TRANSACTION_COUNTER, SECURITY_AUTH_ATTEMPT_COUNTER, INTEGRATION_REQUEST_COUNTER
)


class MultiLayerObservabilityMonitor:
    """
    Monitor de observabilidade multi-camada.
    
    Fornece decoradores e utilitários para instrumentar componentes
    da plataforma com telemetria unificada, incluindo logs estruturados,
    métricas e traces para múltiplas camadas da aplicação.
    """
    
    def __init__(self, configurator: MultiLayerObservabilityConfigurator, component_name: str = None):
        """
        Inicializa o monitor com o configurador fornecido.
        
        Args:
            configurator: Configurador de observabilidade
            component_name: Nome do componente (opcional, sobrescreve o configurador)
        """
        self.configurator = configurator
        self.config = configurator.config
        self.logger = configurator.logger
        self.tracer = configurator.tracer
        self.meter = configurator.meter
        
        # Sobrescrever o nome do componente se fornecido
        if component_name:
            self.config.component_name = component_name
        
        # Inicializar contadores e histogramas comuns
        if self.meter:
            # Métricas de aplicação
            self.app_request_counter = self.meter.create_counter(
                APP_REQUEST_COUNTER,
                description="Número de requisições da aplicação",
                unit="1"
            )
            self.app_request_duration = self.meter.create_histogram(
                APP_REQUEST_DURATION,
                description="Duração de requisições da aplicação",
                unit="ms"
            )
            self.app_error_counter = self.meter.create_counter(
                APP_ERROR_COUNTER,
                description="Número de erros da aplicação",
                unit="1"
            )
            
            # Métricas de negócio
            self.business_transaction_counter = self.meter.create_counter(
                BUSINESS_TRANSACTION_COUNTER,
                description="Número de transações de negócio",
                unit="1"
            )
            
            # Métricas de segurança
            self.security_auth_attempt_counter = self.meter.create_counter(
                SECURITY_AUTH_ATTEMPT_COUNTER,
                description="Número de tentativas de autenticação",
                unit="1"
            )
            
            # Métricas de integração
            self.integration_request_counter = self.meter.create_counter(
                INTEGRATION_REQUEST_COUNTER,
                description="Número de requisições de integração",
                unit="1"
            )
    
    def create_event_context(
        self,
        user_id: Optional[str] = None,
        transaction_id: Optional[str] = None,
        request_id: Optional[str] = None,
        correlation_id: Optional[str] = None,
        additional_attributes: Optional[Dict[str, Any]] = None
    ) -> EventContext:
        """
        Cria um contexto de evento com os atributos fornecidos.
        
        Args:
            user_id: ID do usuário
            transaction_id: ID da transação
            request_id: ID da requisição
            correlation_id: ID de correlação
            additional_attributes: Atributos adicionais
        
        Returns:
            EventContext: Contexto de evento
        """
        context = EventContext(
            tenant_id=self.config.tenant_id,
            module_name=self.config.module_name,
            component_name=self.config.component_name,
            user_id=user_id,
            transaction_id=transaction_id,
            request_id=request_id,
            correlation_id=correlation_id,
            region=self.config.region,
            market_context=self.config.market_context,
        )
        
        if additional_attributes:
            context.add_attributes(additional_attributes)
        
        return context
    
    def log_event(
        self,
        message: str,
        context: EventContext,
        severity: EventSeverity = EventSeverity.INFO,
        category: EventCategory = EventCategory.FUNCTIONALITY,
        exception: Optional[Exception] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> None:
        """
        Registra um evento com contexto enriquecido.
        
        Args:
            message: Mensagem do evento
            context: Contexto do evento
            severity: Severidade do evento
            category: Categoria do evento
            exception: Exceção associada (se houver)
            metadata: Metadados adicionais
        """
        # Preparar metadados do evento
        event_data = {
            "event_category": category,
            **context.to_dict()
        }
        
        # Adicionar metadados extras se fornecidos
        if metadata:
            event_data["metadata"] = metadata
        
        # Mapear severidade para nível de log
        log_level = {
            EventSeverity.DEBUG: logging.DEBUG,
            EventSeverity.INFO: logging.INFO,
            EventSeverity.WARNING: logging.WARNING,
            EventSeverity.ERROR: logging.ERROR,
            EventSeverity.CRITICAL: logging.CRITICAL
        }.get(severity, logging.INFO)
        
        # Registrar o evento
        if exception:
            self.logger.log(log_level, message, exc_info=exception, extra=event_data)
        else:
            self.logger.log(log_level, message, extra=event_data)
    
    def trace_application_request(
        self,
        span_name: Optional[str] = None,
        record_exception: bool = True
    ) -> Callable[[F], F]:
        """
        Decorador para rastrear requisições da camada de aplicação.
        
        Args:
            span_name: Nome personalizado para o span
            record_exception: Se deve registrar exceções
            
        Returns:
            Callable: Decorador para função
        """
        def decorator(func: F) -> F:
            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                # Obter o nome do span
                name = span_name or f"{func.__module__}.{func.__qualname__}"
                
                # Obter contexto da requisição (implementação específica do framework)
                request_context = self._extract_request_context(args, kwargs)
                
                # Iniciar métricas
                start_time = time.time()
                
                # Atributos padrão para span e métricas
                attributes = {
                    "tenant_id": self.config.tenant_id,
                    "module_name": self.config.module_name,
                    "service.name": self.config.service_name,
                    "application.layer": "application"
                }
                
                if self.config.component_name:
                    attributes["component_name"] = self.config.component_name
                
                if request_context:
                    attributes.update(request_context)
                
                # Criar e iniciar span se tracing estiver habilitado
                span = None
                if self.tracer and self.config.tracing_enabled:
                    span = self.tracer.start_span(name, attributes=attributes)
                
                try:
                    # Incrementar contador de requisições
                    if self.meter and hasattr(self, "app_request_counter"):
                        self.app_request_counter.add(1, attributes)
                    
                    # Chamar função original
                    result = func(*args, **kwargs)
                    
                    # Registrar sucesso
                    if span:
                        span.set_attribute("request.status", "success")
                    
                    return result
                
                except Exception as e:
                    # Registrar falha
                    if span and record_exception:
                        span.set_attribute("request.status", "error")
                        span.set_attribute("error.type", e.__class__.__name__)
                        span.set_attribute("error.message", str(e))
                        span.record_exception(e)
                    
                    # Incrementar contador de erros
                    if self.meter and hasattr(self, "app_error_counter"):
                        error_attrs = {**attributes, "error.type": e.__class__.__name__}
                        self.app_error_counter.add(1, error_attrs)
                    
                    # Registrar no log
                    self.logger.error(
                        f"Erro na requisição da aplicação: {name}",
                        exc_info=e,
                        extra={"attributes": attributes}
                    )
                    
                    # Re-lançar exceção
                    raise
                
                finally:
                    # Finalizar métricas de duração
                    duration_ms = (time.time() - start_time) * 1000
                    if self.meter and hasattr(self, "app_request_duration"):
                        self.app_request_duration.record(duration_ms, attributes)
                    
                    # Finalizar span
                    if span:
                        span.__exit__(None, None, None)
            
            # Suporte para funções assíncronas
            if asyncio.iscoroutinefunction(func):
                @functools.wraps(func)
                async def async_wrapper(*args, **kwargs):
                    # Obter o nome do span
                    name = span_name or f"{func.__module__}.{func.__qualname__}"
                    
                    # Obter contexto da requisição
                    request_context = self._extract_request_context(args, kwargs)
                    
                    # Iniciar métricas
                    start_time = time.time()
                    
                    # Atributos padrão para span e métricas
                    attributes = {
                        "tenant_id": self.config.tenant_id,
                        "module_name": self.config.module_name,
                        "service.name": self.config.service_name,
                        "application.layer": "application"
                    }
                    
                    if self.config.component_name:
                        attributes["component_name"] = self.config.component_name
                    
                    if request_context:
                        attributes.update(request_context)
                    
                    # Criar e iniciar span se tracing estiver habilitado
                    span = None
                    if self.tracer and self.config.tracing_enabled:
                        span = self.tracer.start_span(name, attributes=attributes)
                    
                    try:
                        # Incrementar contador de requisições
                        if self.meter and hasattr(self, "app_request_counter"):
                            self.app_request_counter.add(1, attributes)
                        
                        # Chamar função original
                        result = await func(*args, **kwargs)
                        
                        # Registrar sucesso
                        if span:
                            span.set_attribute("request.status", "success")
                        
                        return result
                    
                    except Exception as e:
                        # Registrar falha
                        if span and record_exception:
                            span.set_attribute("request.status", "error")
                            span.set_attribute("error.type", e.__class__.__name__)
                            span.set_attribute("error.message", str(e))
                            span.record_exception(e)
                        
                        # Incrementar contador de erros
                        if self.meter and hasattr(self, "app_error_counter"):
                            error_attrs = {**attributes, "error.type": e.__class__.__name__}
                            self.app_error_counter.add(1, error_attrs)
                        
                        # Registrar no log
                        self.logger.error(
                            f"Erro na requisição assíncrona da aplicação: {name}",
                            exc_info=e,
                            extra={"attributes": attributes}
                        )
                        
                        # Re-lançar exceção
                        raise
                    
                    finally:
                        # Finalizar métricas de duração
                        duration_ms = (time.time() - start_time) * 1000
                        if self.meter and hasattr(self, "app_request_duration"):
                            self.app_request_duration.record(duration_ms, attributes)
                        
                        # Finalizar span
                        if span:
                            span.__exit__(None, None, None)
                
                return cast(F, async_wrapper)
            
            return cast(F, wrapper)
        
        return decorator
    
    def _extract_request_context(self, args, kwargs) -> Dict[str, Any]:
        """
        Extrai contexto da requisição dos argumentos da função.
        
        Esta é uma implementação genérica que pode ser sobrescrita
        para frameworks específicos (FastAPI, Flask, Django, etc.).
        
        Args:
            args: Argumentos posicionais
            kwargs: Argumentos nomeados
            
        Returns:
            Dict[str, Any]: Contexto da requisição
        """
        context = {}
        
        # Tentar extrair de argumentos comuns
        if kwargs.get("request_id"):
            context["request.id"] = kwargs["request_id"]
        
        if kwargs.get("user_id"):
            context["user.id"] = kwargs["user_id"]
        
        if kwargs.get("tenant_id"):
            context["tenant.id"] = kwargs["tenant_id"]
        
        # Tentar extrair de objetos de requisição comuns (Django, Flask, FastAPI)
        for arg in args:
            # Detectar objetos de request comuns
            if hasattr(arg, "path") and hasattr(arg, "method"):
                # Parece um objeto de requisição HTTP
                context["http.method"] = getattr(arg, "method", None)
                context["http.path"] = getattr(arg, "path", None)
                
                # Tentar extrair headers comuns
                headers = getattr(arg, "headers", {})
                if isinstance(headers, dict):
                    user_agent = headers.get("user-agent") or headers.get("User-Agent")
                    if user_agent:
                        context["http.user_agent"] = user_agent
                
                # Tentar extrair ID de usuário se autenticado
                user = getattr(arg, "user", None)
                if user and hasattr(user, "id"):
                    context["user.id"] = str(user.id)
        
        return context