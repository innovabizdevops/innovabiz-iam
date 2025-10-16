"""
Monitor de Observabilidade Multi-Camada para a plataforma INNOVABIZ - Parte 3.

Implementação de monitores específicos para camadas de segurança e negócio.

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
    SECURITY_AUTH_ATTEMPT_COUNTER, SECURITY_AUTH_SUCCESS_COUNTER,
    SECURITY_AUTH_FAILURE_COUNTER, SECURITY_RISK_ASSESSMENT,
    BUSINESS_TRANSACTION_COUNTER, BUSINESS_TRANSACTION_VALUE,
    BUSINESS_PROCESS_DURATION
)


class SecurityLayerMonitor:
    """
    Monitor específico para a camada de segurança da plataforma.
    
    Fornece decoradores e utilitários para monitorar eventos de segurança,
    autenticação, autorização e avaliação de risco.
    """
    
    def __init__(self, parent_monitor):
        """
        Inicializa o monitor de segurança.
        
        Args:
            parent_monitor: Monitor principal de observabilidade
        """
        self.parent = parent_monitor
        self.config = parent_monitor.config
        self.logger = parent_monitor.logger
        self.tracer = parent_monitor.tracer
        self.meter = parent_monitor.meter
        
        # Inicializar contadores e histogramas específicos de segurança
        if self.meter:
            self.auth_attempt_counter = self.meter.create_counter(
                SECURITY_AUTH_ATTEMPT_COUNTER,
                description="Número de tentativas de autenticação",
                unit="1"
            )
            self.auth_success_counter = self.meter.create_counter(
                SECURITY_AUTH_SUCCESS_COUNTER,
                description="Número de autenticações bem-sucedidas",
                unit="1"
            )
            self.auth_failure_counter = self.meter.create_counter(
                SECURITY_AUTH_FAILURE_COUNTER,
                description="Número de falhas de autenticação",
                unit="1"
            )
            self.risk_assessment = self.meter.create_histogram(
                SECURITY_RISK_ASSESSMENT,
                description="Nível de risco avaliado",
                unit="score"
            )
    
    def trace_authentication(
        self,
        auth_type: str = "password",
        mask_credentials: bool = True
    ) -> Callable[[F], F]:
        """
        Decorador para rastrear tentativas de autenticação.
        
        Args:
            auth_type: Tipo de autenticação (password, mfa, biometric, etc.)
            mask_credentials: Se deve mascarar credenciais em logs e traces
            
        Returns:
            Callable: Decorador para função
        """
        def decorator(func: F) -> F:
            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                # Determinar o nome do span
                name = f"auth.{auth_type}"
                
                # Extrair identificador de usuário e tenant
                user_identifier = self._extract_user_identifier(args, kwargs)
                
                # Atributos base para telemetria
                attributes = {
                    "tenant_id": self.config.tenant_id,
                    "module_name": self.config.module_name,
                    "auth.type": auth_type,
                    "security.layer": "authentication"
                }
                
                if user_identifier:
                    attributes["user.identifier"] = user_identifier
                
                if self.config.component_name:
                    attributes["component_name"] = self.config.component_name
                
                # Iniciar medição de tempo
                start_time = time.time()
                
                # Criar evento de contexto
                event_context = self.parent.create_event_context(
                    user_id=user_identifier,
                    additional_attributes={"auth.type": auth_type}
                )
                
                # Criar e iniciar span se tracing estiver habilitado
                span = None
                if self.tracer and self.config.tracing_enabled:
                    span = self.tracer.start_span(name, attributes=attributes)
                
                try:
                    # Registrar tentativa de autenticação
                    if self.meter and hasattr(self, "auth_attempt_counter"):
                        self.auth_attempt_counter.add(1, attributes)
                    
                    self.parent.log_event(
                        message=f"Tentativa de autenticação: tipo={auth_type}",
                        context=event_context,
                        severity=EventSeverity.INFO,
                        category=EventCategory.SECURITY
                    )
                    
                    # Mascarar credenciais se necessário
                    if mask_credentials:
                        masked_args, masked_kwargs = self._mask_credentials(args, kwargs)
                        result = func(*masked_args, **masked_kwargs)
                    else:
                        result = func(*args, **kwargs)
                    
                    # Autenticação bem-sucedida
                    if self.meter and hasattr(self, "auth_success_counter"):
                        self.auth_success_counter.add(1, attributes)
                    
                    if span:
                        span.set_attribute("auth.success", True)
                    
                    self.parent.log_event(
                        message=f"Autenticação bem-sucedida: tipo={auth_type}",
                        context=event_context,
                        severity=EventSeverity.INFO,
                        category=EventCategory.SECURITY
                    )
                    
                    return result
                
                except Exception as e:
                    # Autenticação falhou
                    if self.meter and hasattr(self, "auth_failure_counter"):
                        self.auth_failure_counter.add(1, {
                            **attributes,
                            "error.type": e.__class__.__name__
                        })
                    
                    if span:
                        span.set_attribute("auth.success", False)
                        span.set_attribute("error.type", e.__class__.__name__)
                        span.set_attribute("error.message", str(e))
                        span.record_exception(e)
                    
                    # Registrar falha de autenticação com detalhes apropriados
                    error_msg = str(e)
                    # Evitar vazamento de informações sensíveis no log
                    sanitized_error = self._sanitize_error_message(error_msg)
                    
                    self.parent.log_event(
                        message=f"Falha de autenticação: tipo={auth_type}, erro={sanitized_error}",
                        context=event_context,
                        severity=EventSeverity.WARNING,
                        category=EventCategory.SECURITY
                    )
                    
                    # Re-lançar exceção
                    raise
                
                finally:
                    # Finalizar span
                    if span:
                        span.__exit__(None, None, None)
            
            # Suporte para funções assíncronas
            if asyncio.iscoroutinefunction(func):
                @functools.wraps(func)
                async def async_wrapper(*args, **kwargs):
                    # Determinar o nome do span
                    name = f"auth.{auth_type}"
                    
                    # Extrair identificador de usuário e tenant
                    user_identifier = self._extract_user_identifier(args, kwargs)
                    
                    # Atributos base para telemetria
                    attributes = {
                        "tenant_id": self.config.tenant_id,
                        "module_name": self.config.module_name,
                        "auth.type": auth_type,
                        "security.layer": "authentication"
                    }
                    
                    if user_identifier:
                        attributes["user.identifier"] = user_identifier
                    
                    if self.config.component_name:
                        attributes["component_name"] = self.config.component_name
                    
                    # Iniciar medição de tempo
                    start_time = time.time()
                    
                    # Criar evento de contexto
                    event_context = self.parent.create_event_context(
                        user_id=user_identifier,
                        additional_attributes={"auth.type": auth_type}
                    )
                    
                    # Criar e iniciar span se tracing estiver habilitado
                    span = None
                    if self.tracer and self.config.tracing_enabled:
                        span = self.tracer.start_span(name, attributes=attributes)
                    
                    try:
                        # Registrar tentativa de autenticação
                        if self.meter and hasattr(self, "auth_attempt_counter"):
                            self.auth_attempt_counter.add(1, attributes)
                        
                        self.parent.log_event(
                            message=f"Tentativa de autenticação assíncrona: tipo={auth_type}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.SECURITY
                        )
                        
                        # Mascarar credenciais se necessário
                        if mask_credentials:
                            masked_args, masked_kwargs = self._mask_credentials(args, kwargs)
                            result = await func(*masked_args, **masked_kwargs)
                        else:
                            result = await func(*args, **kwargs)
                        
                        # Autenticação bem-sucedida
                        if self.meter and hasattr(self, "auth_success_counter"):
                            self.auth_success_counter.add(1, attributes)
                        
                        if span:
                            span.set_attribute("auth.success", True)
                        
                        self.parent.log_event(
                            message=f"Autenticação assíncrona bem-sucedida: tipo={auth_type}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.SECURITY
                        )
                        
                        return result
                    
                    except Exception as e:
                        # Autenticação falhou
                        if self.meter and hasattr(self, "auth_failure_counter"):
                            self.auth_failure_counter.add(1, {
                                **attributes,
                                "error.type": e.__class__.__name__
                            })
                        
                        if span:
                            span.set_attribute("auth.success", False)
                            span.set_attribute("error.type", e.__class__.__name__)
                            span.set_attribute("error.message", str(e))
                            span.record_exception(e)
                        
                        # Registrar falha de autenticação com detalhes apropriados
                        error_msg = str(e)
                        # Evitar vazamento de informações sensíveis no log
                        sanitized_error = self._sanitize_error_message(error_msg)
                        
                        self.parent.log_event(
                            message=f"Falha de autenticação assíncrona: tipo={auth_type}, erro={sanitized_error}",
                            context=event_context,
                            severity=EventSeverity.WARNING,
                            category=EventCategory.SECURITY
                        )
                        
                        # Re-lançar exceção
                        raise
                    
                    finally:
                        # Finalizar span
                        if span:
                            span.__exit__(None, None, None)
                
                return cast(F, async_wrapper)
            
            return cast(F, wrapper)
        
        return decorator
    
    def trace_risk_assessment(
        self,
        risk_type: str = "access"
    ) -> Callable[[F], F]:
        """
        Decorador para rastrear avaliações de risco.
        
        Args:
            risk_type: Tipo de avaliação de risco (access, transaction, etc.)
            
        Returns:
            Callable: Decorador para função
        """
        def decorator(func: F) -> F:
            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                # Determinar o nome do span
                name = f"risk.assessment.{risk_type}"
                
                # Extrair identificador de usuário ou entidade
                entity_id = self._extract_entity_identifier(args, kwargs)
                
                # Atributos base para telemetria
                attributes = {
                    "tenant_id": self.config.tenant_id,
                    "module_name": self.config.module_name,
                    "risk.type": risk_type,
                    "security.layer": "risk_assessment"
                }
                
                if entity_id:
                    attributes["entity.id"] = entity_id
                
                if self.config.component_name:
                    attributes["component_name"] = self.config.component_name
                
                # Criar evento de contexto
                event_context = self.parent.create_event_context(
                    user_id=entity_id,
                    additional_attributes={"risk.type": risk_type}
                )
                
                # Criar e iniciar span se tracing estiver habilitado
                span = None
                if self.tracer and self.config.tracing_enabled:
                    span = self.tracer.start_span(name, attributes=attributes)
                
                try:
                    # Chamar função original
                    result = func(*args, **kwargs)
                    
                    # Extrair nível de risco do resultado
                    risk_level = self._extract_risk_level(result)
                    risk_score = self._extract_risk_score(result)
                    
                    # Registrar métricas de risco
                    if risk_score is not None and self.meter and hasattr(self, "risk_assessment"):
                        self.risk_assessment.record(risk_score, {
                            **attributes,
                            "risk.level": risk_level if risk_level else "unknown"
                        })
                    
                    if span:
                        if risk_level:
                            span.set_attribute("risk.level", risk_level)
                        if risk_score is not None:
                            span.set_attribute("risk.score", risk_score)
                    
                    # Log da avaliação de risco
                    self.parent.log_event(
                        message=f"Avaliação de risco: tipo={risk_type}, nível={risk_level}, pontuação={risk_score}",
                        context=event_context,
                        severity=EventSeverity.INFO,
                        category=EventCategory.SECURITY,
                        metadata={"risk_level": risk_level, "risk_score": risk_score}
                    )
                    
                    return result
                
                except Exception as e:
                    # Registrar erro na avaliação de risco
                    if span:
                        span.set_attribute("error.type", e.__class__.__name__)
                        span.set_attribute("error.message", str(e))
                        span.record_exception(e)
                    
                    self.parent.log_event(
                        message=f"Erro na avaliação de risco: tipo={risk_type}, erro={str(e)}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.SECURITY,
                        exception=e
                    )
                    
                    # Re-lançar exceção
                    raise
                
                finally:
                    # Finalizar span
                    if span:
                        span.__exit__(None, None, None)
            
            # Suporte para funções assíncronas
            if asyncio.iscoroutinefunction(func):
                @functools.wraps(func)
                async def async_wrapper(*args, **kwargs):
                    # Implementação assíncrona similar à síncrona
                    name = f"risk.assessment.{risk_type}"
                    entity_id = self._extract_entity_identifier(args, kwargs)
                    
                    attributes = {
                        "tenant_id": self.config.tenant_id,
                        "module_name": self.config.module_name,
                        "risk.type": risk_type,
                        "security.layer": "risk_assessment"
                    }
                    
                    if entity_id:
                        attributes["entity.id"] = entity_id
                    
                    if self.config.component_name:
                        attributes["component_name"] = self.config.component_name
                    
                    event_context = self.parent.create_event_context(
                        user_id=entity_id,
                        additional_attributes={"risk.type": risk_type}
                    )
                    
                    span = None
                    if self.tracer and self.config.tracing_enabled:
                        span = self.tracer.start_span(name, attributes=attributes)
                    
                    try:
                        result = await func(*args, **kwargs)
                        
                        risk_level = self._extract_risk_level(result)
                        risk_score = self._extract_risk_score(result)
                        
                        if risk_score is not None and self.meter and hasattr(self, "risk_assessment"):
                            self.risk_assessment.record(risk_score, {
                                **attributes,
                                "risk.level": risk_level if risk_level else "unknown"
                            })
                        
                        if span:
                            if risk_level:
                                span.set_attribute("risk.level", risk_level)
                            if risk_score is not None:
                                span.set_attribute("risk.score", risk_score)
                        
                        self.parent.log_event(
                            message=f"Avaliação de risco assíncrona: tipo={risk_type}, nível={risk_level}, pontuação={risk_score}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.SECURITY,
                            metadata={"risk_level": risk_level, "risk_score": risk_score}
                        )
                        
                        return result
                    
                    except Exception as e:
                        if span:
                            span.set_attribute("error.type", e.__class__.__name__)
                            span.set_attribute("error.message", str(e))
                            span.record_exception(e)
                        
                        self.parent.log_event(
                            message=f"Erro na avaliação de risco assíncrona: tipo={risk_type}, erro={str(e)}",
                            context=event_context,
                            severity=EventSeverity.ERROR,
                            category=EventCategory.SECURITY,
                            exception=e
                        )
                        
                        raise
                    
                    finally:
                        if span:
                            span.__exit__(None, None, None)
                
                return cast(F, async_wrapper)
            
            return cast(F, wrapper)
        
        return decorator
    
    def _extract_user_identifier(self, args, kwargs) -> Optional[str]:
        """Extrai identificador de usuário dos argumentos da função."""
        # Verificar argumentos nomeados comuns
        for key in ["username", "email", "user_id", "userId"]:
            if key in kwargs:
                return str(kwargs[key])
        
        # Verificar objetos de usuário em argumentos posicionais
        for arg in args:
            if hasattr(arg, "username"):
                return getattr(arg, "username")
            elif hasattr(arg, "email"):
                return getattr(arg, "email")
            elif hasattr(arg, "user_id"):
                return str(getattr(arg, "user_id"))
        
        return None
    
    def _extract_entity_identifier(self, args, kwargs) -> Optional[str]:
        """Extrai identificador de entidade dos argumentos da função."""
        # Verificar argumentos nomeados comuns
        for key in ["entity_id", "entityId", "user_id", "userId", "transaction_id", "transactionId"]:
            if key in kwargs:
                return str(kwargs[key])
        
        # Verificar objetos em argumentos posicionais
        for arg in args:
            for attr in ["id", "entity_id", "user_id", "transaction_id"]:
                if hasattr(arg, attr):
                    return str(getattr(arg, attr))
        
        return None
    
    def _extract_risk_level(self, result) -> Optional[str]:
        """Extrai nível de risco do resultado da função."""
        if hasattr(result, "risk_level"):
            return getattr(result, "risk_level")
        elif hasattr(result, "riskLevel"):
            return getattr(result, "riskLevel")
        elif isinstance(result, dict):
            return result.get("risk_level") or result.get("riskLevel")
        
        return None
    
    def _extract_risk_score(self, result) -> Optional[float]:
        """Extrai pontuação de risco do resultado da função."""
        if hasattr(result, "risk_score"):
            return getattr(result, "risk_score")
        elif hasattr(result, "riskScore"):
            return getattr(result, "riskScore")
        elif isinstance(result, dict):
            score = result.get("risk_score") or result.get("riskScore")
            if score is not None:
                try:
                    return float(score)
                except (ValueError, TypeError):
                    pass
        
        # Tentar converter nível de risco para pontuação
        risk_level = self._extract_risk_level(result)
        if risk_level:
            risk_map = {
                "LOW": 0.2,
                "MEDIUM": 0.5,
                "HIGH": 0.8,
                "CRITICAL": 1.0,
                "VERY_LOW": 0.1,
                "VERY_HIGH": 0.9,
            }
            return risk_map.get(risk_level.upper())
        
        return None
    
    def _mask_credentials(self, args, kwargs):
        """Mascara credenciais sensíveis nos argumentos da função."""
        # Clonar kwargs e mascarar campos sensíveis
        masked_kwargs = kwargs.copy()
        sensitive_fields = ["password", "senha", "secret", "token", "api_key", "apiKey", "access_key", "accessKey"]
        
        for field in sensitive_fields:
            if field in masked_kwargs:
                masked_kwargs[field] = "********"
        
        # Não modificar args por ser uma tupla imutável
        # Em um caso real, poderíamos procurar por objetos com atributos sensíveis
        return args, masked_kwargs
    
    def _sanitize_error_message(self, error_msg: str) -> str:
        """Sanitiza mensagem de erro para remover informações sensíveis."""
        # Padrões de sanitização para campos sensíveis em mensagens de erro
        patterns = [
            (r"password\s*=\s*['\"](.*?)['\"]", "password='********'"),
            (r"token\s*=\s*['\"](.*?)['\"]", "token='********'"),
            (r"api[-_]key\s*=\s*['\"](.*?)['\"]", "api_key='********'"),
            # Adicionar mais padrões conforme necessário
        ]
        
        sanitized_msg = error_msg
        for pattern, replacement in patterns:
            import re
            sanitized_msg = re.sub(pattern, replacement, sanitized_msg)
        
        return sanitized_msg