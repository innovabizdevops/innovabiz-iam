#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Pacote de Dashboard de Monitoramento de Anomalias Comportamentais

Este pacote fornece componentes para o dashboard de monitoramento de anomalias
comportamentais do sistema de detecção de fraudes do INNOVABIZ IAM/TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from .dashboard_manager import DashboardManager
from .dashboard_controller import get_dashboard_router
from .rules_integration import RulesDashboardIntegrator

__all__ = [
    'DashboardManager', 
    'get_dashboard_router', 
    'RulesDashboardIntegrator'
]