"""
Conector entre o sistema de regras dinâmicas e o NeuraFlow.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import os
from typing import Any, Dict, List, Optional, Tuple, Union

from fastapi import Depends, HTTPException, Security
from pydantic import BaseModel, Field

from ...common.security.audit import AuditLogger
from ..rules_engine.rule_types import Rule, RuleSet
from ..rules_engine.repository import RuleRepository
from ..rules_engine.evaluator import RuleEvaluator
from ..rules_engine.statistics import RuleStatisticsService
from .client import NeuraFlowClient
from .enhancer import RuleEnhancer
from .models import (
    EnhancementConfig,
    EnhancementType,
    ModelType,
    NeuraFlowDetectionResponse
)


class NeuraFlowConfig(BaseModel):
    """Configuração para o conector NeuraFlow"""
    api_key: str = Field(..., description="Chave de API para autenticação com NeuraFlow")
    base_url: str = Field("https://api.neuraflow.innovabiz.io/v1", 
                         description="URL base da API NeuraFlow")
    timeout: int = Field(10, description="Tempo limite para requisições em segundos")
    enhancement_types: List[EnhancementType] = Field(
        default_factory=lambda: [
            EnhancementType.FEATURE_EXTRACTION,
            EnhancementType.CONTEXT_ENRICHMENT,
            EnhancementType.RISK_SCORING,
            EnhancementType.BEHAVIORAL_ANALYSIS,
        ],
        description="Tipos de aprimoramento padrão"
    )
    confidence_threshold: float = Field(0.7, description="Limite de confiança padrão")
    cache_ttl: int = Field(300, description="Tempo de vida do cache em segundos")
    auto_optimize_rules: bool = Field(False, 
                                     description="Otimizar regras automaticamente")
    auto_suggest_rules: bool = Field(False, 
                                    description="Sugerir regras automaticamente")
    enabled_models: List[str] = Field(
        default_factory=list,
        description="IDs de modelos específicos a serem usados"
    )
    max_processing_time_ms: int = Field(
        200, 
        description="Tempo máximo de processamento em milissegundos"
    )


class NeuraFlowConnector:
    """
    Conector entre o sistema de regras dinâmicas e o NeuraFlow.
    
    Esta classe fornece uma interface para interagir com o NeuraFlow
    a partir do sistema de regras dinâmicas, permitindo:
    
    1. Inicializar e configurar a conexão com NeuraFlow
    2. Avaliar eventos com aprimoramento de ML/AI
    3. Otimizar regras existentes usando insights de ML/AI
    4. Gerar sugestões de novas regras baseadas em padrões detectados
    5. Fornecer métricas e estatísticas sobre a integração
    """
    
    def __init__(
        self,
        config: NeuraFlowConfig,
        rule_repository: RuleRepository,
        rule_evaluator: RuleEvaluator,
        statistics_service: RuleStatisticsService,
        audit_logger: AuditLogger,
        logger: Optional[logging.Logger] = None,
    ):
        """
        Inicializa o conector NeuraFlow.
        
        Args:
            config: Configuração para o conector
            rule_repository: Repositório de regras
            rule_evaluator: Avaliador de regras
            statistics_service: Serviço de estatísticas de regras
            audit_logger: Logger de auditoria
            logger: Logger para registrar eventos
        """
        self.config = config
        self.rule_repository = rule_repository
        self.rule_evaluator = rule_evaluator
        self.statistics_service = statistics_service
        self.audit_logger = audit_logger
        self.logger = logger or logging.getLogger(__name__)
        
        # Inicializa cliente NeuraFlow
        self.client = NeuraFlowClient(
            api_key=config.api_key,
            base_url=config.base_url,
            timeout=config.timeout,
            logger=self.logger,
        )
        
        # Inicializa enhancer
        enhancement_config = EnhancementConfig(
            enhancement_types=config.enhancement_types,
            confidence_threshold=config.confidence_threshold,
            max_processing_time_ms=config.max_processing_time_ms,
        )
        
        self.enhancer = RuleEnhancer(
            neuraflow_client=self.client,
            rule_evaluator=rule_evaluator,
            logger=self.logger,
            enhancement_config=enhancement_config,
            cache_ttl=config.cache_ttl,
        )
        
        self.logger.info(f"NeuraFlow connector initialized with {len(config.enhancement_types)} enhancement types")
    
    async def close(self):
        """Fecha conexões e recursos."""
        await self.client.close()
        self.logger.info("NeuraFlow connector closed")
    
    async def evaluate_rule_with_enhancement(
        self,
        rule_id: str,
        event_data: Dict[str, Any],
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Avalia uma regra com aprimoramento de ML/AI.
        
        Args:
            rule_id: ID da regra a ser avaliada
            event_data: Dados do evento para avaliação
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Dict[str, Any]: Resultado da avaliação aprimorada
        """
        rule = await self.rule_repository.get_rule_by_id(rule_id)
        
        if not rule:
            raise ValueError(f"Regra não encontrada: {rule_id}")
        
        result = await self.enhancer.evaluate_with_enhancement(
            rule=rule,
            event_data=event_data,
            tenant_id=tenant_id,
            region=region,
        )
        
        # Registra estatísticas de avaliação
        await self.statistics_service.record_rule_evaluation(
            rule_id=rule_id,
            matched=result.get("matched", False),
            score=result.get("score", 0),
            evaluation_time_ms=result.get("evaluation_time_ms", 0),
            enhanced=True,
            tenant_id=tenant_id,
            region=region,
        )
        
        # Registra auditoria
        await self.audit_logger.log_rule_evaluation(
            rule_id=rule_id,
            result=result,
            enhanced=True,
            tenant_id=tenant_id,
            region=region,
        )
        
        return result
    
    async def evaluate_ruleset_with_enhancement(
        self,
        ruleset_id: str,
        event_data: Dict[str, Any],
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Avalia um conjunto de regras com aprimoramento de ML/AI.
        
        Args:
            ruleset_id: ID do conjunto de regras a ser avaliado
            event_data: Dados do evento para avaliação
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Dict[str, Any]: Resultado da avaliação aprimorada
        """
        ruleset = await self.rule_repository.get_ruleset_by_id(ruleset_id)
        
        if not ruleset:
            raise ValueError(f"Conjunto de regras não encontrado: {ruleset_id}")
        
        result = await self.enhancer.evaluate_with_enhancement(
            rule=ruleset,
            event_data=event_data,
            tenant_id=tenant_id,
            region=region,
        )
        
        # Registra estatísticas de avaliação
        await self.statistics_service.record_ruleset_evaluation(
            ruleset_id=ruleset_id,
            matched=result.get("matched", False),
            score=result.get("score", 0),
            evaluation_time_ms=result.get("evaluation_time_ms", 0),
            rules_matched=result.get("rules_matched", 0),
            total_rules=result.get("total_rules", 0),
            enhanced=True,
            tenant_id=tenant_id,
            region=region,
        )
        
        # Registra auditoria
        await self.audit_logger.log_ruleset_evaluation(
            ruleset_id=ruleset_id,
            result=result,
            enhanced=True,
            tenant_id=tenant_id,
            region=region,
        )
        
        return result
    
    async def optimize_rule(
        self,
        rule_id: str,
        sample_count: int = 10,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Tuple[Rule, Dict[str, Any]]:
        """
        Otimiza uma regra usando ML/AI.
        
        Args:
            rule_id: ID da regra a ser otimizada
            sample_count: Número de amostras a usar para otimização
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Tuple[Rule, Dict[str, Any]]: Regra otimizada e metadados de otimização
        """
        rule = await self.rule_repository.get_rule_by_id(rule_id)
        
        if not rule:
            raise ValueError(f"Regra não encontrada: {rule_id}")
        
        # Obtém amostras de eventos que correspondem à regra
        sample_events = await self.statistics_service.get_matching_event_samples(
            rule_id=rule_id,
            limit=sample_count,
            tenant_id=tenant_id,
            region=region,
        )
        
        if not sample_events:
            return rule, {
                "optimized": False, 
                "reason": "Nenhuma amostra de evento correspondente disponível"
            }
        
        # Otimiza regra com base nas amostras
        optimized_rule, optimization_metadata = await self.enhancer.optimize_rule(
            rule=rule,
            sample_events=sample_events,
            tenant_id=tenant_id,
            region=region,
        )
        
        if optimization_metadata.get("optimized", False):
            # Salva regra otimizada
            optimized_rule.id = None  # Força criação de nova regra
            optimized_rule.status = "inactive"  # Inativa por padrão, requer revisão
            
            # Registra auditoria
            await self.audit_logger.log_rule_optimization(
                original_rule_id=rule_id,
                optimized_rule=optimized_rule,
                metadata=optimization_metadata,
                tenant_id=tenant_id,
                region=region,
            )
            
            # Salva regra otimizada no repositório
            saved_rule = await self.rule_repository.create_rule(
                rule=optimized_rule,
                tenant_id=tenant_id,
                region=region,
            )
            
            return saved_rule, optimization_metadata
        
        return rule, optimization_metadata
    
    async def generate_rule_suggestions(
        self,
        event_data: Dict[str, Any],
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> List[Rule]:
        """
        Gera sugestões de regras com base em dados do evento.
        
        Args:
            event_data: Dados do evento para análise
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            List[Rule]: Lista de regras sugeridas
        """
        suggested_rules = await self.enhancer.generate_rule_suggestions(
            event_data=event_data,
            tenant_id=tenant_id,
            region=region,
            confidence_threshold=self.config.confidence_threshold,
        )
        
        if suggested_rules:
            # Registra auditoria
            await self.audit_logger.log_rule_suggestion(
                suggested_rules=suggested_rules,
                event_data=event_data,
                tenant_id=tenant_id,
                region=region,
            )
            
            # Se auto_suggest_rules estiver habilitado, salva as regras sugeridas
            if self.config.auto_suggest_rules:
                saved_rules = []
                
                for rule in suggested_rules:
                    saved_rule = await self.rule_repository.create_rule(
                        rule=rule,
                        tenant_id=tenant_id,
                        region=region,
                    )
                    saved_rules.append(saved_rule)
                
                return saved_rules
        
        return suggested_rules
    
    async def get_neuraflow_models(
        self,
        model_type: Optional[ModelType] = None,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Obtém modelos disponíveis no NeuraFlow.
        
        Args:
            model_type: Tipo de modelo a ser filtrado
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Dict[str, Any]: Informações sobre modelos disponíveis
        """
        try:
            model_info = await self.client.get_available_models(
                model_type=model_type,
                tenant_id=tenant_id,
                region=region,
            )
            
            return model_info.dict()
            
        except Exception as e:
            self.logger.error(f"Failed to get NeuraFlow models: {str(e)}")
            return {
                "error": str(e),
                "available_models": [],
                "recommended_models": {},
                "region_specific_models": {},
                "global_models": [],
            }
    
    async def get_connector_metrics(
        self,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Obtém métricas sobre o conector NeuraFlow.
        
        Args:
            tenant_id: ID do tenant
            region: Região
            
        Returns:
            Dict[str, Any]: Métricas do conector
        """
        # Obtém estatísticas de avaliações aprimoradas
        enhanced_stats = await self.statistics_service.get_enhanced_evaluation_statistics(
            tenant_id=tenant_id,
            region=region,
        )
        
        # Obtém estatísticas de otimizações de regras
        optimization_stats = await self.statistics_service.get_optimization_statistics(
            tenant_id=tenant_id,
            region=region,
        )
        
        # Obtém estatísticas de sugestões de regras
        suggestion_stats = await self.statistics_service.get_suggestion_statistics(
            tenant_id=tenant_id,
            region=region,
        )
        
        # Retorna métricas consolidadas
        return {
            "enhanced_evaluations": enhanced_stats,
            "rule_optimizations": optimization_stats,
            "rule_suggestions": suggestion_stats,
            "enhancement_types": [et.value for et in self.config.enhancement_types],
            "auto_optimize_rules": self.config.auto_optimize_rules,
            "auto_suggest_rules": self.config.auto_suggest_rules,
            "confidence_threshold": self.config.confidence_threshold,
        }


async def get_neuraflow_connector(
    rule_repository: RuleRepository = Depends(),
    rule_evaluator: RuleEvaluator = Depends(),
    statistics_service: RuleStatisticsService = Depends(),
    audit_logger: AuditLogger = Depends(),
) -> NeuraFlowConnector:
    """
    Dependência do FastAPI para obter o conector NeuraFlow configurado.
    
    Args:
        rule_repository: Repositório de regras
        rule_evaluator: Avaliador de regras
        statistics_service: Serviço de estatísticas de regras
        audit_logger: Logger de auditoria
        
    Returns:
        NeuraFlowConnector: Conector configurado
    """
    # Carrega configuração do ambiente
    config = NeuraFlowConfig(
        api_key=os.environ.get("NEURAFLOW_API_KEY", ""),
        base_url=os.environ.get("NEURAFLOW_API_URL", "https://api.neuraflow.innovabiz.io/v1"),
        timeout=int(os.environ.get("NEURAFLOW_TIMEOUT", "10")),
        confidence_threshold=float(os.environ.get("NEURAFLOW_CONFIDENCE_THRESHOLD", "0.7")),
        cache_ttl=int(os.environ.get("NEURAFLOW_CACHE_TTL", "300")),
        auto_optimize_rules=os.environ.get("NEURAFLOW_AUTO_OPTIMIZE", "false").lower() == "true",
        auto_suggest_rules=os.environ.get("NEURAFLOW_AUTO_SUGGEST", "false").lower() == "true",
        max_processing_time_ms=int(os.environ.get("NEURAFLOW_MAX_PROCESSING_TIME", "200")),
    )
    
    # Configura tipos de aprimoramento
    enhancement_types_str = os.environ.get(
        "NEURAFLOW_ENHANCEMENT_TYPES",
        "feature_extraction,context_enrichment,risk_scoring,behavioral_analysis"
    )
    
    if enhancement_types_str:
        config.enhancement_types = [
            EnhancementType(et.strip())
            for et in enhancement_types_str.split(",")
            if et.strip()
        ]
    
    # Configura modelos habilitados
    enabled_models_str = os.environ.get("NEURAFLOW_ENABLED_MODELS", "")
    
    if enabled_models_str:
        config.enabled_models = [
            model_id.strip()
            for model_id in enabled_models_str.split(",")
            if model_id.strip()
        ]
    
    # Cria e retorna conector
    connector = NeuraFlowConnector(
        config=config,
        rule_repository=rule_repository,
        rule_evaluator=rule_evaluator,
        statistics_service=statistics_service,
        audit_logger=audit_logger,
    )
    
    try:
        yield connector
    finally:
        await connector.close()