#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
INNOVABIZ IAM - Módulo Regional de Análise Comportamental

Este pacote contém implementações específicas por região/país para
análise comportamental e detecção de fraudes, adaptadas aos padrões,
regulações e comportamentos de usuários de cada localidade.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from .angola_behavioral_patterns import create_angola_analyzer
from .brazil_behavioral_patterns import create_brazil_analyzer
from .mozambique_behavioral_patterns import create_mozambique_analyzer

__all__ = [
    'create_angola_analyzer',
    'create_brazil_analyzer',
    'create_mozambique_analyzer',
]