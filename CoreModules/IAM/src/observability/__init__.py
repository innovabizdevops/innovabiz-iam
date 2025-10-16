"""
Módulo de observabilidade para IAM Audit Service da plataforma INNOVABIZ.

Este pacote implementa recursos abrangentes de observabilidade, incluindo:
- Métricas Prometheus
- Health checks e diagnósticos
- Rastreamento distribuído com OpenTelemetry
- Integração multi-tenant e multi-regional

O design segue as decisões arquiteturais documentadas nos ADRs:
- ADR-002: Observabilidade com Prometheus e Grafana
- ADR-004: Padrões de Métricas de Observabilidade
- ADR-005: Framework de Decoradores e Middleware
- ADR-006: Health Checks e Diagnósticos
- ADR-007: Sistema de Alertas e Dashboards
- ADR-008: Rastreamento com OpenTelemetry

Autor: INNOVABIZ DevOps Team
"""

from .config import ObservabilityConfig
from .integration import ObservabilityIntegration, ContextInfo
from .metrics import MetricsManager, instrument_audit_event, instrument_function
from .health import HealthChecker, HealthResponse, DiagnosticResponse
from .tracing import TracingIntegration, traced

__all__ = [
    # Classes principais
    'ObservabilityIntegration',
    'ObservabilityConfig',
    'MetricsManager',
    'HealthChecker',
    'TracingIntegration',
    'ContextInfo',
    
    # Modelos e respostas
    'HealthResponse',
    'DiagnosticResponse',
    
    # Decoradores e funções de instrumentação
    'instrument_audit_event',
    'instrument_function',
    'traced',
]

__version__ = '0.1.0'