#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Sistema de Regras Dinâmicas para Detecção de Anomalias Comportamentais

Este pacote implementa um sistema completo de regras dinâmicas para detecção de
anomalias comportamentais no módulo TrustGuard do INNOVABIZ IAM, permitindo
configurar, testar e avaliar regras de segurança em tempo real.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from .rule_model import (
    Rule, RuleSet, RuleCondition, RuleGroup, 
    RuleOperator, RuleLogicalOperator, RuleAction, 
    RuleSeverity, RuleCategory, RuleValueType
)
from .rule_evaluator import RuleEvaluator, RuleEvaluationResult
from .rule_repository import RuleRepository
from .rule_controller import get_rules_engine_router

__all__ = [
    # Classes de modelo
    'Rule', 'RuleSet', 'RuleCondition', 'RuleGroup',
    
    # Enums
    'RuleOperator', 'RuleLogicalOperator', 'RuleAction', 'RuleSeverity', 'RuleCategory', 'RuleValueType',
    
    # Avaliador de regras
    'RuleEvaluator', 'RuleEvaluationResult',
    
    # Repositório de regras
    'RuleRepository',
    
    # Controlador API
    'get_rules_engine_router',
]