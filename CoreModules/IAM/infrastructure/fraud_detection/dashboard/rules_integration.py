#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Integração entre Dashboard de Anomalias e Sistema de Regras Dinâmicas

Este módulo implementa a integração entre o dashboard de monitoramento de anomalias
comportamentais e o sistema de regras dinâmicas, permitindo que o dashboard utilize
o sistema de regras para avaliar eventos comportamentais em tempo real.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import logging
import datetime
from typing import Dict, Any, List, Optional, Tuple, Union

from ..rules_engine.rule_model import (
    Rule, RuleSet, RuleOperator, RuleLogicalOperator,
    RuleAction, RuleSeverity, RuleCategory
)
from ..rules_engine.rule_evaluator import RuleEvaluator
from ..rules_engine.rule_repository import RuleRepository

# Configuração do logger
logger = logging.getLogger("iam.trustguard.dashboard.rules_integration")


class RulesDashboardIntegrator:
    """
    Integrador entre o dashboard de anomalias comportamentais e o sistema de regras dinâmicas.
    
    Esta classe fornece métodos para avaliação de eventos comportamentais utilizando o
    sistema de regras dinâmicas, e para transformação dos resultados em formato adequado
    para exibição no dashboard.
    """
    
    def __init__(
        self,
        rule_repository: Optional[RuleRepository] = None,
        rule_evaluator: Optional[RuleEvaluator] = None
    ):
        """
        Inicializa o integrador.
        
        Args:
            rule_repository: Repositório de regras (opcional)
            rule_evaluator: Avaliador de regras (opcional)
        """
        self.rule_repository = rule_repository or RuleRepository()
        self.rule_evaluator = rule_evaluator or RuleEvaluator()
        self._rule_cache = {}
        self._ruleset_cache = {}
        self._cache_timestamp = None
        self._cache_lock = asyncio.Lock()
        
        # Inicializar cache
        self._refresh_cache()
    
    def _refresh_cache(self) -> None:
        """
        Atualiza o cache de regras e conjuntos de regras.
        """
        try:
            # Atualizar cache de regras
            rules = self.rule_repository.get_rules()
            self._rule_cache = {rule.id: rule for rule in rules}
            
            # Atualizar cache de conjuntos de regras
            rulesets = self.rule_repository.get_rulesets()
            self._ruleset_cache = {ruleset.id: ruleset for ruleset in rulesets}
            
            # Registrar timestamp
            self._cache_timestamp = datetime.datetime.now()
            
            logger.info(
                f"Cache atualizado: {len(self._rule_cache)} regras, "
                f"{len(self._ruleset_cache)} conjuntos"
            )
        except Exception as e:
            logger.error(f"Erro ao atualizar cache de regras: {str(e)}")
    
    async def refresh_cache_async(self) -> None:
        """
        Atualiza o cache de regras e conjuntos de regras de forma assíncrona.
        """
        async with self._cache_lock:
            # Verificar se o cache precisa ser atualizado (a cada 5 minutos)
            if (
                self._cache_timestamp is None or
                (datetime.datetime.now() - self._cache_timestamp).total_seconds() > 300
            ):
                self._refresh_cache()
    
    async def evaluate_event(
        self, 
        event_data: Dict[str, Any],
        region: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None
    ) -> Tuple[List[Dict[str, Any]], Dict[str, Any]]:
        """
        Avalia um evento comportamental utilizando o sistema de regras dinâmicas.
        
        Args:
            event_data: Dados do evento comportamental
            region: Região para filtrar regras (opcional)
            context: Contexto adicional para avaliação (opcional)
            
        Returns:
            Tuple com lista de resultados de avaliação e resultados de ações
        """
        # Atualizar cache se necessário
        await self.refresh_cache_async()
        
        # Contexto padrão
        if context is None:
            context = {
                "timestamp": datetime.datetime.now().isoformat(),
                "source": "dashboard"
            }
        
        # Obter conjuntos de regras para a região
        rulesets = self.rule_repository.get_rulesets(region=region)
        
        if not rulesets:
            logger.warning(
                f"Nenhum conjunto de regras encontrado para a região {region}"
            )
            return [], {}
        
        # Resultados por conjunto
        all_results = []
        all_actions = {}
        
        # Avaliar cada conjunto
        for ruleset in rulesets:
            try:
                # Avaliar evento
                results, action_results = self.rule_evaluator.process_event(
                    ruleset, event_data, context
                )
                
                # Converter resultados para formato do dashboard
                for result in results:
                    if not result.matched:
                        continue
                        
                    dashboard_result = {
                        "rule_id": result.rule.id,
                        "ruleset_id": ruleset.id,
                        "rule_name": result.rule.name,
                        "ruleset_name": ruleset.name,
                        "score": result.score,
                        "severity": result.rule.severity.value if hasattr(result.rule.severity, "value") else result.rule.severity,
                        "category": result.rule.category.value if hasattr(result.rule.category, "value") else result.rule.category,
                        "timestamp": context.get("timestamp"),
                        "region": ruleset.region or "global",
                        "matched_fields": result.matched_fields,
                        "actions": result.actions,
                        "event_type": event_data.get("event_type", "unknown"),
                        "user_id": event_data.get("user_id"),
                        "session_id": event_data.get("session_id"),
                        "device_id": event_data.get("device_id"),
                        "ip_address": event_data.get("ip_address"),
                    }
                    
                    all_results.append(dashboard_result)
                
                # Consolidar resultados de ações
                for key, value in action_results.items():
                    if key in all_actions:
                        if isinstance(all_actions[key], list):
                            if isinstance(value, list):
                                all_actions[key].extend(value)
                            else:
                                all_actions[key].append(value)
                    else:
                        all_actions[key] = value
                
            except Exception as e:
                logger.error(
                    f"Erro ao avaliar evento com conjunto {ruleset.id}: {str(e)}"
                )
        
        return all_results, all_actions
    
    async def evaluate_batch(
        self,
        events: List[Dict[str, Any]],
        region: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None
    ) -> List[Tuple[Dict[str, Any], List[Dict[str, Any]], Dict[str, Any]]]:
        """
        Avalia um lote de eventos comportamentais utilizando o sistema de regras dinâmicas.
        
        Args:
            events: Lista de eventos comportamentais
            region: Região para filtrar regras (opcional)
            context: Contexto adicional para avaliação (opcional)
            
        Returns:
            Lista de tuplas (evento, resultados, ações)
        """
        results = []
        
        # Avaliar cada evento
        for event in events:
            try:
                # Contexto específico do evento
                event_context = context.copy() if context else {}
                event_context.update({
                    "event_id": event.get("id"),
                    "timestamp": event.get("timestamp") or datetime.datetime.now().isoformat(),
                    "source": "dashboard_batch"
                })
                
                # Avaliar evento
                event_results, event_actions = await self.evaluate_event(
                    event, region, event_context
                )
                
                # Adicionar aos resultados
                results.append((event, event_results, event_actions))
                
            except Exception as e:
                logger.error(f"Erro ao avaliar evento em lote: {str(e)}")
                results.append((event, [], {}))
        
        return results
    
    async def get_rule_statistics(
        self, 
        region: Optional[str] = None,
        days: int = 7
    ) -> Dict[str, Any]:
        """
        Obtém estatísticas de regras para exibição no dashboard.
        
        Args:
            region: Região para filtrar regras (opcional)
            days: Quantidade de dias para analisar (padrão: 7)
            
        Returns:
            Estatísticas de regras
        """
        # Atualizar cache se necessário
        await self.refresh_cache_async()
        
        # Obter regras para a região
        rules = self.rule_repository.get_rules(region=region)
        
        # Calcular estatísticas (simulado)
        # Em produção, consultaria banco de dados com eventos avaliados
        stats = {
            "total_rules": len(rules),
            "active_rules": sum(1 for rule in rules if rule.enabled),
            "rules_by_severity": {
                "critical": sum(1 for rule in rules if rule.severity == RuleSeverity.CRITICAL),
                "high": sum(1 for rule in rules if rule.severity == RuleSeverity.HIGH),
                "medium": sum(1 for rule in rules if rule.severity == RuleSeverity.MEDIUM),
                "low": sum(1 for rule in rules if rule.severity == RuleSeverity.LOW),
                "info": sum(1 for rule in rules if rule.severity == RuleSeverity.INFO),
            },
            "rules_by_category": {},
            "top_triggered_rules": [],
            "period_days": days,
            "region": region or "global"
        }
        
        # Categorias
        for rule in rules:
            category = rule.category.value if hasattr(rule.category, "value") else rule.category
            if category in stats["rules_by_category"]:
                stats["rules_by_category"][category] += 1
            else:
                stats["rules_by_category"][category] = 1
        
        # Top regras acionadas (simulado)
        # Em produção, obteria do banco de dados
        stats["top_triggered_rules"] = [
            {
                "rule_id": rule.id,
                "rule_name": rule.name,
                "severity": rule.severity.value if hasattr(rule.severity, "value") else rule.severity,
                "category": rule.category.value if hasattr(rule.category, "value") else rule.category,
                "trigger_count": 0,  # Simulado
                "last_triggered": None  # Simulado
            }
            for rule in rules[:5] if rule.enabled
        ]
        
        return stats