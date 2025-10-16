#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Motor de avaliação de regras dinâmicas para detecção de anomalias comportamentais

Este módulo implementa o motor de avaliação das regras dinâmicas, responsável
por aplicar as regras aos eventos comportamentais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import re
import logging
import datetime
from typing import Dict, Any, List, Optional, Union, Tuple, Callable
import json

from .rule_model import (
    Rule, RuleSet, RuleCondition, RuleGroup, 
    RuleOperator, RuleLogicalOperator, RuleAction, 
    RuleSeverity, RuleCategory
)

# Configuração do logger
logger = logging.getLogger("iam.trustguard.rules_engine.evaluator")


class RuleEvaluationResult:
    """Resultado da avaliação de uma regra."""
    
    def __init__(
        self, 
        rule: Rule, 
        matched: bool, 
        score: float = 0.0,
        matched_fields: Optional[List[str]] = None,
        actions: Optional[List[str]] = None,
        context: Optional[Dict[str, Any]] = None,
    ):
        """
        Inicializa o resultado da avaliação.
        
        Args:
            rule: Regra avaliada
            matched: Se a regra foi ativada (True) ou não (False)
            score: Pontuação da avaliação
            matched_fields: Campos que ativaram a regra
            actions: Ações a serem executadas
            context: Contexto da avaliação
        """
        self.rule = rule
        self.matched = matched
        self.score = score
        self.matched_fields = matched_fields or []
        self.actions = actions or [a.value if hasattr(a, "value") else a for a in rule.actions] if matched else []
        self.context = context or {}
        self.timestamp = datetime.datetime.now()
    
    def to_dict(self) -> Dict[str, Any]:
        """
        Converte o resultado para dicionário.
        
        Returns:
            Dicionário com dados do resultado
        """
        return {
            "rule_id": self.rule.id,
            "rule_name": self.rule.name,
            "matched": self.matched,
            "score": self.score,
            "matched_fields": self.matched_fields,
            "actions": self.actions,
            "context": self.context,
            "timestamp": self.timestamp.isoformat(),
            "severity": self.rule.severity.value if hasattr(self.rule.severity, "value") else self.rule.severity,
            "category": self.rule.category.value if hasattr(self.rule.category, "value") else self.rule.category,
        }


class RuleEvaluator:
    """
    Motor de avaliação de regras dinâmicas.
    
    Responsável por avaliar eventos comportamentais usando regras dinâmicas.
    """
    
    def __init__(self, default_region: Optional[str] = None):
        """
        Inicializa o avaliador de regras.
        
        Args:
            default_region: Região padrão para avaliação
        """
        self.default_region = default_region
        
        # Registro de funções de acesso a campos
        self.field_accessors: Dict[str, Callable] = {}
        
        # Registro de funções personalizadas
        self.custom_functions: Dict[str, Callable] = {}
        
        # Registro de funções de ação
        self.action_handlers: Dict[str, Callable] = {}
        
        # Registrar funções padrão
        self._register_default_functions()
        
        logger.info("RuleEvaluator inicializado")
    
    def _register_default_functions(self) -> None:
        """Registra funções padrão para acesso a campos e ações."""
        # Função padrão para acessar campos via notação de ponto
        self.register_field_accessor("default", self._default_field_accessor)
        
        # Funções de manipulação de data/hora
        self.register_custom_function("now", lambda: datetime.datetime.now())
        self.register_custom_function("today", lambda: datetime.date.today())
        self.register_custom_function("utcnow", lambda: datetime.datetime.utcnow())
        self.register_custom_function("timestamp", lambda: datetime.datetime.now().timestamp())
        
        # Funções de manipulação de strings
        self.register_custom_function("lower", lambda s: s.lower() if isinstance(s, str) else s)
        self.register_custom_function("upper", lambda s: s.upper() if isinstance(s, str) else s)
        self.register_custom_function("len", lambda v: len(v) if hasattr(v, "__len__") else 0)
        
        # Funções de manipulação de listas
        self.register_custom_function("count", lambda l, v: l.count(v) if hasattr(l, "count") else 0)
        self.register_custom_function("sum", lambda l: sum(l) if hasattr(l, "__iter__") else 0)
        self.register_custom_function("avg", lambda l: sum(l) / len(l) if hasattr(l, "__iter__") and len(l) > 0 else 0)
        
        # Ações padrão
        self.register_action_handler(RuleAction.LOG.value, self._default_log_action)
        self.register_action_handler(RuleAction.ALERT.value, self._default_alert_action)
    
    def register_field_accessor(self, name: str, accessor: Callable) -> None:
        """
        Registra uma função de acesso a campos.
        
        Args:
            name: Nome do acessor
            accessor: Função de acesso
        """
        self.field_accessors[name] = accessor
        logger.debug(f"Registrado acessor de campos: {name}")
    
    def register_custom_function(self, name: str, function: Callable) -> None:
        """
        Registra uma função personalizada.
        
        Args:
            name: Nome da função
            function: Função personalizada
        """
        self.custom_functions[name] = function
        logger.debug(f"Registrada função personalizada: {name}")
    
    def register_action_handler(self, action: str, handler: Callable) -> None:
        """
        Registra um manipulador de ação.
        
        Args:
            action: Nome da ação
            handler: Função manipuladora
        """
        self.action_handlers[action] = handler
        logger.debug(f"Registrado manipulador de ação: {action}")
    
    def _default_field_accessor(self, data: Dict[str, Any], field_path: str) -> Any:
        """
        Acessa um campo nos dados usando notação de ponto.
        
        Args:
            data: Dados do evento
            field_path: Caminho do campo usando notação de ponto
            
        Returns:
            Valor do campo ou None se não encontrado
        """
        # Verificar se é uma chamada de função
        if "(" in field_path and field_path.endswith(")"):
            # Formato: function_name(arg1, arg2, ...)
            function_call = field_path.split("(", 1)
            function_name = function_call[0].strip()
            args_str = function_call[1][:-1]  # Remover ')'
            
            # Verificar se a função existe
            if function_name in self.custom_functions:
                # Parsear argumentos
                args = []
                if args_str.strip():
                    # Implementação simplificada de parsing de argumentos
                    # Em produção, usar um parser mais robusto
                    for arg in args_str.split(","):
                        arg = arg.strip()
                        # Tentar converter para número ou booleano
                        if arg.lower() == "true":
                            args.append(True)
                        elif arg.lower() == "false":
                            args.append(False)
                        elif arg.lower() == "null":
                            args.append(None)
                        else:
                            try:
                                if "." in arg:
                                    args.append(float(arg))
                                else:
                                    args.append(int(arg))
                            except ValueError:
                                # Se não for número, verificar se é um campo
                                if arg.startswith("$"):
                                    # Referência a outro campo
                                    field_ref = arg[1:]  # Remover '$'
                                    args.append(self._default_field_accessor(data, field_ref))
                                else:
                                    # String literal
                                    args.append(arg.strip('"\''))
                
                # Chamar a função
                try:
                    return self.custom_functions[function_name](*args)
                except Exception as e:
                    logger.warning(f"Erro ao chamar função {function_name}: {str(e)}")
                    return None
            
            logger.warning(f"Função não encontrada: {function_name}")
            return None
        
        # Acesso normal a campo
        current = data
        for part in field_path.split("."):
            # Suporte a índices de lista: field[0]
            if "[" in part and part.endswith("]"):
                field_name, index_str = part.split("[", 1)
                index = int(index_str[:-1])  # Remover ']'
                
                if field_name:
                    # Acessar campo e depois índice: items[0]
                    if isinstance(current, dict) and field_name in current:
                        current = current[field_name]
                    else:
                        return None
                
                # Acessar índice
                if isinstance(current, (list, tuple)) and 0 <= index < len(current):
                    current = current[index]
                else:
                    return None
            elif isinstance(current, dict) and part in current:
                current = current[part]
            else:
                return None
                
        return current
    
    def _default_log_action(self, result: RuleEvaluationResult, context: Dict[str, Any]) -> None:
        """
        Ação padrão para log.
        
        Args:
            result: Resultado da avaliação
            context: Contexto adicional
        """
        logger.info(
            f"Regra ativada: {result.rule.name} ({result.rule.id})"
            f" - Severidade: {result.rule.severity}"
            f" - Categoria: {result.rule.category}"
        )
    
    def _default_alert_action(self, result: RuleEvaluationResult, context: Dict[str, Any]) -> None:
        """
        Ação padrão para alerta.
        
        Args:
            result: Resultado da avaliação
            context: Contexto adicional
        """
        # Em produção, integrar com sistema de alertas
        logger.info(
            f"ALERTA: Regra ativada: {result.rule.name} ({result.rule.id})"
            f" - Severidade: {result.rule.severity}"
            f" - Categoria: {result.rule.category}"
        )
    
    def evaluate_condition(self, condition: RuleCondition, data: Dict[str, Any], context: Dict[str, Any]) -> bool:
        """
        Avalia uma condição de regra.
        
        Args:
            condition: Condição a ser avaliada
            data: Dados do evento
            context: Contexto da avaliação
            
        Returns:
            True se a condição é satisfeita, False caso contrário
        """
        try:
            # Obter valor do campo
            field_value = self._default_field_accessor(data, condition.field)
            
            # Valor para comparação
            comparison_value = condition.value
            
            # Avaliar com base no operador
            if condition.operator == RuleOperator.EQUAL:
                return field_value == comparison_value
                
            elif condition.operator == RuleOperator.NOT_EQUAL:
                return field_value != comparison_value
                
            elif condition.operator == RuleOperator.GREATER_THAN:
                return field_value > comparison_value
                
            elif condition.operator == RuleOperator.GREATER_EQUAL:
                return field_value >= comparison_value
                
            elif condition.operator == RuleOperator.LESS_THAN:
                return field_value < comparison_value
                
            elif condition.operator == RuleOperator.LESS_EQUAL:
                return field_value <= comparison_value
                
            elif condition.operator == RuleOperator.IN:
                return field_value in comparison_value
                
            elif condition.operator == RuleOperator.NOT_IN:
                return field_value not in comparison_value
                
            elif condition.operator == RuleOperator.CONTAINS:
                if isinstance(field_value, str) and isinstance(comparison_value, str):
                    return comparison_value in field_value
                elif isinstance(field_value, (list, tuple, set)):
                    return comparison_value in field_value
                return False
                
            elif condition.operator == RuleOperator.STARTS_WITH:
                return isinstance(field_value, str) and field_value.startswith(comparison_value)
                
            elif condition.operator == RuleOperator.ENDS_WITH:
                return isinstance(field_value, str) and field_value.endswith(comparison_value)
                
            elif condition.operator == RuleOperator.REGEX:
                return isinstance(field_value, str) and bool(re.match(comparison_value, field_value))
                
            elif condition.operator == RuleOperator.EXISTS:
                return field_value is not None
                
            elif condition.operator == RuleOperator.NOT_EXISTS:
                return field_value is None
                
            elif condition.operator == RuleOperator.BETWEEN:
                if isinstance(comparison_value, (list, tuple)) and len(comparison_value) == 2:
                    return comparison_value[0] <= field_value <= comparison_value[1]
                return False
                
            elif condition.operator == RuleOperator.NOT_BETWEEN:
                if isinstance(comparison_value, (list, tuple)) and len(comparison_value) == 2:
                    return field_value < comparison_value[0] or field_value > comparison_value[1]
                return False
                
            elif condition.operator == RuleOperator.ALL:
                if isinstance(field_value, (list, tuple)):
                    return all(item == comparison_value for item in field_value)
                return False
                
            elif condition.operator == RuleOperator.ANY:
                if isinstance(field_value, (list, tuple)):
                    return any(item == comparison_value for item in field_value)
                return False
                
            elif condition.operator == RuleOperator.NONE:
                if isinstance(field_value, (list, tuple)):
                    return all(item != comparison_value for item in field_value)
                return True
                
            logger.warning(f"Operador não suportado: {condition.operator}")
            return False
            
        except Exception as e:
            logger.warning(f"Erro ao avaliar condição: {str(e)}")
            return False
    
    def evaluate_group(self, group: RuleGroup, data: Dict[str, Any], context: Dict[str, Any]) -> bool:
        """
        Avalia um grupo de condições.
        
        Args:
            group: Grupo a ser avaliado
            data: Dados do evento
            context: Contexto da avaliação
            
        Returns:
            True se o grupo é satisfeito, False caso contrário
        """
        try:
            # Avaliar condições do grupo
            results = []
            for condition in group.conditions:
                if isinstance(condition, RuleCondition):
                    result = self.evaluate_condition(condition, data, context)
                elif isinstance(condition, RuleGroup):
                    result = self.evaluate_group(condition, data, context)
                else:
                    logger.warning(f"Tipo de condição não suportado: {type(condition)}")
                    result = False
                    
                results.append(result)
            
            # Aplicar operador lógico
            if group.operator == RuleLogicalOperator.AND:
                return all(results)
                
            elif group.operator == RuleLogicalOperator.OR:
                return any(results)
                
            elif group.operator == RuleLogicalOperator.NOT:
                # NOT só deve ter uma condição
                if len(results) == 1:
                    return not results[0]
                else:
                    logger.warning("Operador NOT deve ter exatamente uma condição")
                    return False
                    
            logger.warning(f"Operador lógico não suportado: {group.operator}")
            return False
            
        except Exception as e:
            logger.warning(f"Erro ao avaliar grupo: {str(e)}")
            return False
    
    def evaluate_rule(self, rule: Rule, data: Dict[str, Any], context: Optional[Dict[str, Any]] = None) -> RuleEvaluationResult:
        """
        Avalia uma regra para um evento comportamental.
        
        Args:
            rule: Regra a ser avaliada
            data: Dados do evento
            context: Contexto da avaliação (opcional)
            
        Returns:
            Resultado da avaliação
        """
        if context is None:
            context = {}
            
        # Verificar se a regra está habilitada
        if not rule.enabled:
            return RuleEvaluationResult(rule, False, 0.0, [], [], context)
            
        # Verificar região se especificada
        if rule.region and "region" in data:
            if rule.region != data["region"] and rule.region != "global":
                return RuleEvaluationResult(rule, False, 0.0, [], [], context)
                
        try:
            # Avaliar condição
            matched = False
            matched_fields = []
            
            if isinstance(rule.condition, RuleCondition):
                matched = self.evaluate_condition(rule.condition, data, context)
                if matched:
                    matched_fields.append(rule.condition.field)
                    
            elif isinstance(rule.condition, RuleGroup):
                matched = self.evaluate_group(rule.condition, data, context)
                
            else:
                logger.warning(f"Tipo de condição não suportado: {type(rule.condition)}")
            
            # Calcular score (simulado)
            # Em produção, usar uma lógica mais sofisticada
            score = 0.0
            if matched:
                # Base score com base na severidade
                severity_scores = {
                    RuleSeverity.LOW: 0.2,
                    RuleSeverity.MEDIUM: 0.5,
                    RuleSeverity.HIGH: 0.8,
                    RuleSeverity.CRITICAL: 1.0,
                }
                base_score = severity_scores.get(rule.severity, 0.5)
                
                # Aplicar multiplicador
                score = base_score * rule.score_multiplier
                
            return RuleEvaluationResult(
                rule=rule, 
                matched=matched,
                score=score,
                matched_fields=matched_fields,
                context=context
            )
            
        except Exception as e:
            logger.error(f"Erro ao avaliar regra {rule.id}: {str(e)}")
            return RuleEvaluationResult(rule, False, 0.0, [], [], context)
    
    def evaluate_ruleset(
        self, ruleset: RuleSet, data: Dict[str, Any], context: Optional[Dict[str, Any]] = None
    ) -> List[RuleEvaluationResult]:
        """
        Avalia um conjunto de regras para um evento comportamental.
        
        Args:
            ruleset: Conjunto de regras
            data: Dados do evento
            context: Contexto da avaliação (opcional)
            
        Returns:
            Lista de resultados da avaliação
        """
        if context is None:
            context = {}
            
        results = []
        
        # Verificar se o ruleset está habilitado
        if not ruleset.enabled:
            return results
            
        # Verificar região se especificada
        if ruleset.region and "region" in data:
            if ruleset.region != data["region"] and ruleset.region != "global":
                return results
        
        # Avaliar cada regra
        for rule in ruleset.rules:
            result = self.evaluate_rule(rule, data, context)
            results.append(result)
            
        return results
    
    def execute_actions(
        self, result: RuleEvaluationResult, data: Dict[str, Any], context: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Executa as ações de uma regra ativada.
        
        Args:
            result: Resultado da avaliação
            data: Dados do evento
            context: Contexto da avaliação (opcional)
            
        Returns:
            Resultado das ações
        """
        if context is None:
            context = {}
            
        action_results = {}
        
        # Somente executar ações se a regra foi ativada
        if not result.matched:
            return action_results
            
        # Executar cada ação
        for action in result.actions:
            try:
                # Verificar se há um handler registrado
                if action in self.action_handlers:
                    handler = self.action_handlers[action]
                    action_context = {**context, "data": data, "result": result.to_dict()}
                    handler_result = handler(result, action_context)
                    action_results[action] = handler_result
                else:
                    logger.warning(f"Nenhum handler registrado para ação: {action}")
                    
            except Exception as e:
                logger.error(f"Erro ao executar ação {action}: {str(e)}")
                action_results[action] = {"error": str(e)}
                
        return action_results
    
    def process_event(
        self, ruleset: RuleSet, data: Dict[str, Any], context: Optional[Dict[str, Any]] = None
    ) -> Tuple[List[RuleEvaluationResult], Dict[str, Any]]:
        """
        Processa um evento aplicando um conjunto de regras.
        
        Args:
            ruleset: Conjunto de regras
            data: Dados do evento
            context: Contexto da avaliação (opcional)
            
        Returns:
            Tupla com lista de resultados e resultados das ações
        """
        if context is None:
            context = {}
            
        # Avaliar regras
        results = self.evaluate_ruleset(ruleset, data, context)
        
        # Executar ações para regras ativadas
        action_results = {}
        for result in results:
            if result.matched:
                action_result = self.execute_actions(result, data, context)
                action_results[result.rule.id] = action_result
                
        return results, action_results