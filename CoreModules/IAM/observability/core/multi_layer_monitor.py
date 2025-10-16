"""
Monitor de Observabilidade Multi-Camada para a plataforma INNOVABIZ.

Este módulo implementa o monitor de observabilidade multi-camada,
fornecendo decoradores e utilitários para instrumentar componentes
da plataforma com telemetria unificada.

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
from typing import Dict, List, Optional, Union, Any, Callable, TypeVar, cast
from datetime import datetime
from enum import Enum

from .multi_layer_config import (
    MultiLayerObservabilityConfig,
    MultiLayerObservabilityConfigurator,
    ObservabilityLevel,
    SecurityContext,
    TelemetryTag
)

# Tipo genérico para funções
F = TypeVar('F', bound=Callable[..., Any])
AsyncF = TypeVar('AsyncF', bound=Callable[..., Any])

# Constantes para nomes de métricas
METRIC_PREFIX = "innovabiz"

# Métricas de aplicação
APP_REQUEST_COUNTER = f"{METRIC_PREFIX}.app.request.count"
APP_REQUEST_DURATION = f"{METRIC_PREFIX}.app.request.duration"
APP_ERROR_COUNTER = f"{METRIC_PREFIX}.app.error.count"
APP_ACTIVE_SESSIONS = f"{METRIC_PREFIX}.app.sessions.active"

# Métricas de negócio
BUSINESS_TRANSACTION_COUNTER = f"{METRIC_PREFIX}.business.transaction.count"
BUSINESS_TRANSACTION_VALUE = f"{METRIC_PREFIX}.business.transaction.value"
BUSINESS_PROCESS_DURATION = f"{METRIC_PREFIX}.business.process.duration"

# Métricas de segurança
SECURITY_AUTH_ATTEMPT_COUNTER = f"{METRIC_PREFIX}.security.auth.attempt.count"
SECURITY_AUTH_SUCCESS_COUNTER = f"{METRIC_PREFIX}.security.auth.success.count"
SECURITY_AUTH_FAILURE_COUNTER = f"{METRIC_PREFIX}.security.auth.failure.count"
SECURITY_RISK_ASSESSMENT = f"{METRIC_PREFIX}.security.risk.level"
SECURITY_THREAT_COUNTER = f"{METRIC_PREFIX}.security.threat.count"

# Métricas de integração
INTEGRATION_REQUEST_COUNTER = f"{METRIC_PREFIX}.integration.request.count"
INTEGRATION_REQUEST_DURATION = f"{METRIC_PREFIX}.integration.request.duration"
INTEGRATION_ERROR_COUNTER = f"{METRIC_PREFIX}.integration.error.count"

# Métricas de usuário
USER_INTERACTION_COUNTER = f"{METRIC_PREFIX}.user.interaction.count"
USER_JOURNEY_DURATION = f"{METRIC_PREFIX}.user.journey.duration"
USER_ERROR_COUNTER = f"{METRIC_PREFIX}.user.error.count"
USER_SATISFACTION = f"{METRIC_PREFIX}.user.satisfaction.score"

# Enums para classificação de eventos
class EventSeverity(str, Enum):
    """Níveis de severidade para eventos."""
    DEBUG = "debug"
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


class EventCategory(str, Enum):
    """Categorias de eventos para classificação."""
    SECURITY = "security"
    PERFORMANCE = "performance"
    AVAILABILITY = "availability"
    FUNCTIONALITY = "functionality"
    USABILITY = "usability"
    BUSINESS = "business"
    COMPLIANCE = "compliance"
    INFRASTRUCTURE = "infrastructure"
    INTEGRATION = "integration"


class EventContext:
    """Contexto de evento para enriquecimento de logs e métricas."""
    
    def __init__(
        self,
        tenant_id: str,
        module_name: str,
        component_name: Optional[str] = None,
        user_id: Optional[str] = None,
        transaction_id: Optional[str] = None,
        request_id: Optional[str] = None,
        correlation_id: Optional[str] = None,
        region: Optional[str] = None,
        market_context: Optional[str] = None,
    ):
        self.tenant_id = tenant_id
        self.module_name = module_name
        self.component_name = component_name
        self.user_id = user_id
        self.transaction_id = transaction_id
        self.request_id = request_id
        self.correlation_id = correlation_id or request_id
        self.region = region
        self.market_context = market_context
        self.timestamp = datetime.utcnow().isoformat() + "Z"
        self.attributes = {}
    
    def add_attribute(self, key: str, value: Any) -> 'EventContext':
        """Adiciona um atributo ao contexto."""
        self.attributes[key] = value
        return self
    
    def add_attributes(self, attributes: Dict[str, Any]) -> 'EventContext':
        """Adiciona múltiplos atributos ao contexto."""
        self.attributes.update(attributes)
        return self
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o contexto para um dicionário."""
        result = {
            "tenant_id": self.tenant_id,
            "module_name": self.module_name,
            "timestamp": self.timestamp,
        }
        
        # Adicionar atributos opcionais se presentes
        if self.component_name:
            result["component_name"] = self.component_name
        if self.user_id:
            result["user_id"] = self.user_id
        if self.transaction_id:
            result["transaction_id"] = self.transaction_id
        if self.request_id:
            result["request_id"] = self.request_id
        if self.correlation_id:
            result["correlation_id"] = self.correlation_id
        if self.region:
            result["region"] = self.region
        if self.market_context:
            result["market_context"] = self.market_context
        
        # Adicionar atributos extras
        result.update(self.attributes)
        
        return result