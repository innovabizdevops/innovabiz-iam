"""
INNOVABIZ - Sistema de Observabilidade Multi-Camada

Módulo de observabilidade unificado para toda a plataforma INNOVABIZ.
Fornece ferramentas para monitoramento, logs, métricas e tracing
em várias camadas da aplicação (infraestrutura, aplicação, negócio, segurança, usuário).

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - Sistema de Observabilidade Multi-Camada
Data: 20/08/2025
"""

from .multi_layer_config import (
    MultiLayerObservabilityConfig,
    MultiLayerObservabilityConfigurator,
    ObservabilityLevel
)

from .multi_layer_monitor import (
    EventContext,
    EventCategory,
    EventSeverity,
    MultiLayerObservabilityMonitor
)

from .multi_layer_monitor_part3 import SecurityLayerMonitor
from .multi_layer_monitor_part4 import BusinessLayerMonitor, UnifiedMonitor

# Exportar funções utilitárias para fácil importação
def create_monitor_from_env(
    module_name: str = None,
    component_name: str = None,
    logger_name: str = "innovabiz.observability"
):
    """
    Cria um monitor unificado a partir de variáveis de ambiente.
    
    Args:
        module_name: Nome do módulo (opcional, se não fornecido usa OTEL_SERVICE_NAME da env)
        component_name: Nome do componente específico (opcional)
        logger_name: Base para nomes de loggers
    
    Returns:
        UnifiedMonitor: Monitor unificado pronto para uso
    """
    config = MultiLayerObservabilityConfig()
    
    if module_name:
        config.module_name = module_name
    if component_name:
        config.component_name = component_name
    
    # Criar configurador
    configurator = MultiLayerObservabilityConfigurator(config, logger_name)
    configurator.configure_from_env()
    
    # Criar e retornar monitor unificado
    return UnifiedMonitor(configurator)


# Decoradores rápidos para uso comum
def trace_request(operation_name: str = None):
    """
    Decorador para rastrear requisições HTTP/API.
    
    Args:
        operation_name: Nome da operação (opcional)
    
    Returns:
        Callable: Decorador configurado
    """
    monitor = create_monitor_from_env()
    return monitor.base_monitor.trace_request(operation_name)


def trace_auth(auth_type: str = "password"):
    """
    Decorador para rastrear autenticação.
    
    Args:
        auth_type: Tipo de autenticação (password, mfa, biometric, etc.)
    
    Returns:
        Callable: Decorador configurado
    """
    monitor = create_monitor_from_env()
    return monitor.security.trace_authentication(auth_type=auth_type)


def trace_transaction(transaction_type: str):
    """
    Decorador para rastrear transações de negócio.
    
    Args:
        transaction_type: Tipo da transação
    
    Returns:
        Callable: Decorador configurado
    """
    monitor = create_monitor_from_env()
    return monitor.business.trace_business_transaction(transaction_type=transaction_type)


def trace_risk(risk_type: str = "access"):
    """
    Decorador para rastrear avaliações de risco.
    
    Args:
        risk_type: Tipo de avaliação de risco
    
    Returns:
        Callable: Decorador configurado
    """
    monitor = create_monitor_from_env()
    return monitor.security.trace_risk_assessment(risk_type=risk_type)


# Aliases comuns para facilidade de uso
create_observer = create_monitor_from_env

# Definir o que é exportado ao importar *
__all__ = [
    # Classes principais
    'MultiLayerObservabilityConfig',
    'MultiLayerObservabilityConfigurator',
    'MultiLayerObservabilityMonitor',
    'SecurityLayerMonitor',
    'BusinessLayerMonitor',
    'UnifiedMonitor',
    'ObservabilityLevel',
    'EventContext',
    'EventCategory',
    'EventSeverity',
    
    # Funções de utilidade
    'create_monitor_from_env',
    'create_observer',
    
    # Decoradores
    'trace_request',
    'trace_auth',
    'trace_transaction',
    'trace_risk',
]