"""
Aprimorador de regras que integra o motor de regras dinâmicas com NeuraFlow.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
from typing import Any, Dict, List, Optional, Set, Tuple, Union

from ..rules_engine.rule_types import Rule, RuleSet, RuleCondition, RuleGroup, RuleOperator
from ..rules_engine.evaluator import RuleEvaluator
from .models import (
    EnhancedEventData,
    EnhancedFeature,
    EnhancementConfig,
    EnhancementType,
    ModelType,
    NeuraFlowDetectionResponse
)
from .client import NeuraFlowClient


class RuleEnhancer:
    """
    Aprimorador de regras que integra o motor de regras dinâmicas com NeuraFlow.
    
    Esta classe funciona como uma ponte entre o motor de regras dinâmicas
    e os modelos avançados de ML/AI do NeuraFlow, permitindo:
    
    1. Enriquecer dados de eventos com insights de ML/AI
    2. Gerar novas regras com base em padrões detectados por ML/AI
    3. Otimizar regras existentes com base em análises avançadas
    4. Contextualizar avaliações de regras com escores de risco de ML/AI
    """
    
    def __init__(
        self,
        neuraflow_client: NeuraFlowClient,
        rule_evaluator: RuleEvaluator,
        logger: Optional[logging.Logger] = None,
        enhancement_config: Optional[EnhancementConfig] = None,
        cache_ttl: int = 300,  # 5 minutos
    ):
        """
        Inicializa o aprimorador de regras.
        
        Args:
            neuraflow_client: Cliente para API do NeuraFlow
            rule_evaluator: Avaliador de regras
            logger: Logger para registrar eventos
            enhancement_config: Configuração padrão para aprimoramento
            cache_ttl: Tempo de vida do cache em segundos
        """
        self.client = neuraflow_client
        self.evaluator = rule_evaluator
        self.logger = logger or logging.getLogger(__name__)
        self.cache_ttl = cache_ttl
        self.feature_cache = {}
        self.enhancement_config = enhancement_config or EnhancementConfig(
            enhancement_types=[
                EnhancementType.FEATURE_EXTRACTION,
                EnhancementType.CONTEXT_ENRICHMENT,
                EnhancementType.RISK_SCORING
            ],
            confidence_threshold=0.7,
        )
        
        self.logger.info("Rule enhancer initialized with NeuraFlow integration")
    
    async def enhance_event_data(
        self,
        event_data: Dict[str, Any],
        config: Optional[EnhancementConfig] = None,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> EnhancedEventData:
        """
        Aprimora dados de evento com insights de ML/AI do NeuraFlow.
        
        Args:
            event_data: Dados do evento para aprimoramento
            config: Configuração específica para este aprimoramento
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            EnhancedEventData: Dados aprimorados do evento
        """
        enhancement_config = config or self.enhancement_config
        
        cache_key = self._generate_cache_key(event_data, enhancement_config)
        cached_data = self._get_from_cache(cache_key)
        
        if cached_data:
            self.logger.debug("Enhanced event data found in cache")
            return cached_data
        
        try:
            enhanced_data = await self.client.enhance_event_data(
                event_data=event_data,
                config=enhancement_config,
                tenant_id=tenant_id,
                region=region,
            )
            
            self._add_to_cache(cache_key, enhanced_data)
            self.logger.info(f"Event data enhanced with {len(enhanced_data.enhanced_features)} features")
            
            return enhanced_data
            
        except Exception as e:
            self.logger.error(f"Failed to enhance event data: {str(e)}")
            # Retorna os dados originais em caso de erro
            return EnhancedEventData(
                original_event=event_data,
                enhanced_features={},
                risk_indicators={},
                context_enrichment={},
                behavioral_patterns={},
                regional_factors={},
                temporal_analysis={},
            )
    
    async def evaluate_with_enhancement(
        self,
        rule: Union[Rule, RuleSet],
        event_data: Dict[str, Any],
        config: Optional[EnhancementConfig] = None,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Avalia regra ou conjunto de regras com dados aprimorados.
        
        Args:
            rule: Regra ou conjunto de regras para avaliar
            event_data: Dados do evento original
            config: Configuração para aprimoramento
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Dict[str, Any]: Resultado da avaliação aprimorada
        """
        try:
            enhanced_data = await self.enhance_event_data(
                event_data=event_data,
                config=config,
                tenant_id=tenant_id,
                region=region,
            )
            
            # Mescla dados originais com dados aprimorados para avaliação
            evaluation_data = {**event_data}
            
            # Adiciona features aprimoradas aos dados de avaliação
            for feature_name, feature in enhanced_data.enhanced_features.items():
                evaluation_data[f"enhanced.{feature_name}"] = feature.value
                evaluation_data[f"enhanced.{feature_name}.confidence"] = feature.confidence
                evaluation_data[f"enhanced.{feature_name}.importance"] = feature.importance
            
            # Adiciona indicadores de risco aos dados de avaliação
            for risk_name, risk_value in enhanced_data.risk_indicators.items():
                evaluation_data[f"risk.{risk_name}"] = risk_value
            
            # Adiciona dados de contexto aos dados de avaliação
            for context_name, context_value in enhanced_data.context_enrichment.items():
                evaluation_data[f"context.{context_name}"] = context_value
            
            # Adiciona dados comportamentais aos dados de avaliação
            for pattern_name, pattern_value in enhanced_data.behavioral_patterns.items():
                evaluation_data[f"behavior.{pattern_name}"] = pattern_value
            
            # Adiciona fatores regionais aos dados de avaliação
            for factor_name, factor_value in enhanced_data.regional_factors.items():
                evaluation_data[f"region.{factor_name}"] = factor_value
            
            # Adiciona análise temporal aos dados de avaliação
            for time_name, time_value in enhanced_data.temporal_analysis.items():
                evaluation_data[f"temporal.{time_name}"] = time_value
            
            # Avalia regra ou conjunto com dados aprimorados
            if isinstance(rule, Rule):
                result = await self.evaluator.evaluate_rule(rule, evaluation_data)
                
                # Adiciona metadados de aprimoramento
                result["enhanced"] = True
                result["enhancement_metadata"] = {
                    "feature_count": len(enhanced_data.enhanced_features),
                    "risk_indicators": list(enhanced_data.risk_indicators.keys()),
                    "context_fields": list(enhanced_data.context_enrichment.keys()),
                }
                
                return result
            else:  # RuleSet
                result = await self.evaluator.evaluate_ruleset(rule, evaluation_data)
                
                # Adiciona metadados de aprimoramento
                result["enhanced"] = True
                result["enhancement_metadata"] = {
                    "feature_count": len(enhanced_data.enhanced_features),
                    "risk_indicators": list(enhanced_data.risk_indicators.keys()),
                    "context_fields": list(enhanced_data.context_enrichment.keys()),
                }
                
                return result
        
        except Exception as e:
            self.logger.error(f"Enhanced evaluation failed: {str(e)}")
            # Fallback para avaliação tradicional em caso de erro
            if isinstance(rule, Rule):
                result = await self.evaluator.evaluate_rule(rule, event_data)
            else:  # RuleSet
                result = await self.evaluator.evaluate_ruleset(rule, event_data)
            
            result["enhanced"] = False
            result["enhancement_error"] = str(e)
            
            return result
    
    async def generate_rule_suggestions(
        self,
        event_data: Dict[str, Any],
        detection_result: Optional[NeuraFlowDetectionResponse] = None,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
        confidence_threshold: float = 0.8,
    ) -> List[Rule]:
        """
        Gera sugestões de regras com base em dados do evento e detecções de ML/AI.
        
        Args:
            event_data: Dados do evento para análise
            detection_result: Resultado de detecção do NeuraFlow (opcional)
            tenant_id: ID do tenant
            region: Região
            confidence_threshold: Limite mínimo de confiança para sugestões
            
        Returns:
            List[Rule]: Lista de regras sugeridas
        """
        try:
            # Se não for fornecido um resultado de detecção, obtenha um
            if not detection_result:
                detection_result = await self.client.detect_anomalies(
                    event_data=event_data,
                    model_type=ModelType.PATTERN_RECOGNITION,
                    tenant_id=tenant_id,
                    region=region,
                    confidence_threshold=confidence_threshold,
                )
            
            # Apenas considere resultados com confiança acima do threshold
            valid_results = [r for r in detection_result.results if r.confidence >= confidence_threshold]
            
            if not valid_results:
                self.logger.info("No high-confidence patterns detected for rule suggestions")
                return []
            
            suggested_rules = []
            
            for result in valid_results:
                # Extrai características importantes para gerar condições de regras
                top_features = sorted(
                    result.features.items(),
                    key=lambda x: x[1],
                    reverse=True,
                )[:5]  # Top 5 features
                
                conditions = []
                
                for feature_name, importance in top_features:
                    if feature_name in event_data:
                        feature_value = event_data[feature_name]
                        
                        # Cria condição com base no tipo de dado
                        if isinstance(feature_value, (int, float)):
                            # Para valores numéricos, cria condição de faixa
                            # baseado no valor do evento com margem de 10%
                            margin = abs(feature_value * 0.1)
                            min_value = feature_value - margin
                            max_value = feature_value + margin
                            
                            conditions.append(RuleCondition(
                                field=feature_name,
                                operator=RuleOperator.GREATER_THAN_OR_EQUAL,
                                value=min_value,
                            ))
                            
                            conditions.append(RuleCondition(
                                field=feature_name,
                                operator=RuleOperator.LESS_THAN_OR_EQUAL,
                                value=max_value,
                            ))
                            
                        elif isinstance(feature_value, str):
                            # Para strings, cria condição de igualdade exata
                            conditions.append(RuleCondition(
                                field=feature_name,
                                operator=RuleOperator.EQUALS,
                                value=feature_value,
                            ))
                            
                        elif isinstance(feature_value, bool):
                            # Para booleanos, cria condição de igualdade
                            conditions.append(RuleCondition(
                                field=feature_name,
                                operator=RuleOperator.EQUALS,
                                value=feature_value,
                            ))
                
                if conditions:
                    # Cria grupo de condições
                    rule_group = RuleGroup(
                        conditions=conditions,
                        logical_operator="AND",
                    )
                    
                    # Cria regra sugerida
                    suggested_rule = Rule(
                        name=f"NeuraFlow Suggested Rule: {result.model_type}",
                        description=f"Regra gerada automaticamente com base em padrões detectados por IA/ML",
                        conditions=rule_group,
                        actions=["log", "notify"],
                        category="ML_GENERATED",
                        severity="medium",
                        score=int(result.score * 100),  # Converte score para escala 0-100
                        metadata={
                            "source": "neuraflow",
                            "model_id": result.model_id,
                            "model_type": result.model_type,
                            "confidence": result.confidence,
                            "generated_at": detection_result.timestamp,
                        },
                        status="inactive",  # Inativa por padrão, requer revisão humana
                        tags=["neuraflow", "ml-generated", result.model_type],
                        created_by="neuraflow-enhancer",
                    )
                    
                    suggested_rules.append(suggested_rule)
            
            self.logger.info(f"Generated {len(suggested_rules)} rule suggestions from ML/AI patterns")
            return suggested_rules
            
        except Exception as e:
            self.logger.error(f"Failed to generate rule suggestions: {str(e)}")
            return []
    
    async def optimize_rule(
        self,
        rule: Rule,
        sample_events: List[Dict[str, Any]],
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Tuple[Rule, Dict[str, Any]]:
        """
        Otimiza uma regra com base em eventos de exemplo e análise de ML/AI.
        
        Args:
            rule: Regra a ser otimizada
            sample_events: Lista de eventos de exemplo para análise
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Tuple[Rule, Dict[str, Any]]: Regra otimizada e metadados de otimização
        """
        if not sample_events:
            return rule, {"optimized": False, "reason": "No sample events provided"}
        
        try:
            # Analisa cada evento de exemplo com NeuraFlow
            enhanced_events = []
            
            for event in sample_events:
                enhanced_data = await self.enhance_event_data(
                    event_data=event,
                    tenant_id=tenant_id,
                    region=region,
                )
                enhanced_events.append(enhanced_data)
            
            # Identifica características importantes presentes em todos os eventos
            common_features = self._find_common_important_features(enhanced_events)
            
            if not common_features:
                return rule, {"optimized": False, "reason": "No common important features found"}
            
            # Cria uma cópia da regra original para otimização
            optimized_rule = rule.copy(deep=True)
            
            # Extrai todas as condições existentes
            existing_fields = self._extract_condition_fields(optimized_rule)
            
            # Adiciona novas condições baseadas em features comuns
            new_conditions = []
            
            for feature_name, feature_info in common_features.items():
                if feature_name not in existing_fields:
                    # Adiciona condição com base no tipo de dado
                    value_type = feature_info["type"]
                    value = feature_info["value"]
                    
                    if value_type in ("int", "float"):
                        # Para valores numéricos, usa faixa de valores
                        min_value = feature_info["min_value"]
                        max_value = feature_info["max_value"]
                        
                        new_conditions.append(RuleCondition(
                            field=f"enhanced.{feature_name}",
                            operator=RuleOperator.GREATER_THAN_OR_EQUAL,
                            value=min_value,
                        ))
                        
                        new_conditions.append(RuleCondition(
                            field=f"enhanced.{feature_name}",
                            operator=RuleOperator.LESS_THAN_OR_EQUAL,
                            value=max_value,
                        ))
                        
                    elif value_type == "str":
                        # Para strings, usa igualdade exata
                        new_conditions.append(RuleCondition(
                            field=f"enhanced.{feature_name}",
                            operator=RuleOperator.EQUALS,
                            value=value,
                        ))
                        
                    elif value_type == "bool":
                        # Para booleanos, usa igualdade
                        new_conditions.append(RuleCondition(
                            field=f"enhanced.{feature_name}",
                            operator=RuleOperator.EQUALS,
                            value=value,
                        ))
            
            # Adiciona novas condições à regra
            if isinstance(optimized_rule.conditions, RuleGroup):
                for condition in new_conditions:
                    optimized_rule.conditions.conditions.append(condition)
            else:
                # Se não for um grupo, cria um grupo com as condições existentes e as novas
                existing_condition = optimized_rule.conditions
                new_group = RuleGroup(
                    conditions=[existing_condition] + new_conditions,
                    logical_operator="AND",
                )
                optimized_rule.conditions = new_group
            
            # Atualiza metadados da regra
            if not optimized_rule.metadata:
                optimized_rule.metadata = {}
                
            optimized_rule.metadata["neuraflow_optimized"] = True
            optimized_rule.metadata["optimization_timestamp"] = self._get_timestamp()
            optimized_rule.metadata["new_conditions_count"] = len(new_conditions)
            
            if not optimized_rule.tags:
                optimized_rule.tags = []
                
            if "neuraflow-optimized" not in optimized_rule.tags:
                optimized_rule.tags.append("neuraflow-optimized")
            
            # Atualiza nome e descrição
            optimized_rule.name = f"{optimized_rule.name} (Optimized)"
            optimized_rule.description = f"{optimized_rule.description}\n\nOtimizado com NeuraFlow em {self._get_timestamp()}"
            
            self.logger.info(f"Rule optimized with {len(new_conditions)} new conditions from ML/AI analysis")
            
            optimization_metadata = {
                "optimized": True,
                "new_conditions": len(new_conditions),
                "common_features": list(common_features.keys()),
                "sample_count": len(sample_events),
            }
            
            return optimized_rule, optimization_metadata
            
        except Exception as e:
            self.logger.error(f"Failed to optimize rule: {str(e)}")
            return rule, {"optimized": False, "error": str(e)}
    
    def _generate_cache_key(self, event_data: Dict[str, Any], config: EnhancementConfig) -> str:
        """Gera chave de cache para dados de evento e configuração."""
        event_hash = hash(json.dumps(event_data, sort_keys=True))
        config_hash = hash(json.dumps(config.dict(), sort_keys=True))
        return f"{event_hash}:{config_hash}"
    
    def _get_from_cache(self, key: str) -> Optional[EnhancedEventData]:
        """Obtém dados aprimorados do cache."""
        if key in self.feature_cache:
            entry = self.feature_cache[key]
            if entry["timestamp"] + self.cache_ttl > self._get_timestamp_seconds():
                return entry["data"]
            else:
                # Remove entrada expirada
                del self.feature_cache[key]
        return None
    
    def _add_to_cache(self, key: str, data: EnhancedEventData) -> None:
        """Adiciona dados aprimorados ao cache."""
        self.feature_cache[key] = {
            "data": data,
            "timestamp": self._get_timestamp_seconds(),
        }
        
        # Limpa cache se ficar muito grande
        if len(self.feature_cache) > 1000:
            # Remove entradas mais antigas
            sorted_keys = sorted(
                self.feature_cache.keys(),
                key=lambda k: self.feature_cache[k]["timestamp"],
            )
            for old_key in sorted_keys[:200]:  # Remove 20% das entradas mais antigas
                del self.feature_cache[old_key]
    
    def _get_timestamp_seconds(self) -> int:
        """Obtém timestamp atual em segundos."""
        import time
        return int(time.time())
    
    def _get_timestamp(self) -> str:
        """Obtém timestamp atual formatado."""
        import datetime
        return datetime.datetime.now().isoformat()
    
    def _extract_condition_fields(self, rule: Rule) -> Set[str]:
        """Extrai todos os campos de condição de uma regra."""
        fields = set()
        
        if isinstance(rule.conditions, RuleCondition):
            fields.add(rule.conditions.field)
            
        elif isinstance(rule.conditions, RuleGroup):
            for condition in rule.conditions.conditions:
                if isinstance(condition, RuleCondition):
                    fields.add(condition.field)
                elif isinstance(condition, RuleGroup):
                    # Recursivamente extrai campos de subgrupos
                    for subcond in condition.conditions:
                        if isinstance(subcond, RuleCondition):
                            fields.add(subcond.field)
        
        return fields
    
    def _find_common_important_features(
        self,
        enhanced_events: List[EnhancedEventData],
    ) -> Dict[str, Dict[str, Any]]:
        """
        Identifica características importantes presentes em todos os eventos.
        
        Args:
            enhanced_events: Lista de eventos aprimorados
            
        Returns:
            Dict[str, Dict[str, Any]]: Dicionário de características comuns com metadados
        """
        if not enhanced_events:
            return {}
        
        # Conta ocorrências de cada feature
        feature_occurrences = {}
        feature_values = {}
        feature_types = {}
        feature_min_values = {}
        feature_max_values = {}
        
        for event in enhanced_events:
            for feature_name, feature in event.enhanced_features.items():
                if feature.importance >= 0.5:  # Apenas considere features importantes
                    if feature_name not in feature_occurrences:
                        feature_occurrences[feature_name] = 0
                        feature_values[feature_name] = []
                    
                    feature_occurrences[feature_name] += 1
                    feature_values[feature_name].append(feature.value)
                    
                    # Determina tipo da feature
                    value_type = type(feature.value).__name__
                    if feature_name not in feature_types:
                        feature_types[feature_name] = value_type
                    
                    # Para valores numéricos, registra mínimo e máximo
                    if isinstance(feature.value, (int, float)):
                        if feature_name not in feature_min_values:
                            feature_min_values[feature_name] = feature.value
                            feature_max_values[feature_name] = feature.value
                        else:
                            feature_min_values[feature_name] = min(
                                feature_min_values[feature_name],
                                feature.value
                            )
                            feature_max_values[feature_name] = max(
                                feature_max_values[feature_name],
                                feature.value
                            )
        
        # Identifica features presentes em todos os eventos
        total_events = len(enhanced_events)
        common_features = {}
        
        for feature_name, occurrences in feature_occurrences.items():
            if occurrences == total_events:
                # Determina valor mais comum para a feature
                if feature_types[feature_name] in ("int", "float"):
                    common_features[feature_name] = {
                        "type": feature_types[feature_name],
                        "value": sum(feature_values[feature_name]) / len(feature_values[feature_name]),  # média
                        "min_value": feature_min_values[feature_name],
                        "max_value": feature_max_values[feature_name],
                    }
                else:
                    # Para valores não numéricos, usa o valor mais frequente
                    from collections import Counter
                    value_counts = Counter(feature_values[feature_name])
                    most_common_value = value_counts.most_common(1)[0][0]
                    
                    common_features[feature_name] = {
                        "type": feature_types[feature_name],
                        "value": most_common_value,
                    }
        
        return common_features