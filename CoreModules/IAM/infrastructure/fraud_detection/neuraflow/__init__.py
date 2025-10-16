"""
Módulo de integração com NeuraFlow para potencialização da detecção de anomalias
via modelos de machine learning e inteligência artificial avançados.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from .client import NeuraFlowClient
from .enhancer import RuleEnhancer
from .connector import NeuraFlowConnector
from .models import (
    NeuraFlowDetectionRequest, 
    NeuraFlowDetectionResponse,
    EnhancedEventData,
    ModelMetadata
)

__all__ = [
    'NeuraFlowClient',
    'RuleEnhancer',
    'NeuraFlowConnector',
    'NeuraFlowDetectionRequest',
    'NeuraFlowDetectionResponse',
    'EnhancedEventData',
    'ModelMetadata'
]