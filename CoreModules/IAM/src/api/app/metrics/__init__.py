"""
Pacote de métricas para o serviço de auditoria do IAM.

Este pacote fornece instrumentação Prometheus para monitoramento
e observabilidade do serviço de auditoria multi-contexto do IAM.
"""
from .audit_metrics import (
    # Middleware
    metrics_middleware,
    
    # Decoradores
    instrument_audit_event_processing,
    instrument_retention_policy,
    instrument_compliance_check,
    
    # Funções de utilidade
    setup_service_info,
    update_service_health,
    register_retention_policies,
    start_uptime_counter,
    init_metrics
)

__all__ = [
    'metrics_middleware',
    'instrument_audit_event_processing',
    'instrument_retention_policy',
    'instrument_compliance_check',
    'setup_service_info',
    'update_service_health',
    'register_retention_policies',
    'start_uptime_counter',
    'init_metrics'
]