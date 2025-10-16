"""
Integrador entre o sistema de regras dinâmicas, o Bureau de Créditos e o TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional, Set, Tuple, Union

from fastapi import Depends, FastAPI, HTTPException, Header, Security
from pydantic import BaseModel, Field, validator

# Importações do TrustGuard
from .trust_guard_connector import TrustGuardConnector
from .trust_guard_factory import get_trust_guard_connector
from .trust_guard_models import (
    AccessDecision,
    AccessDecisionResponse,
    AccessRequest,
    RiskLevel,
    UserContext,
    UserRisk,
)

# Importações do Bureau de Créditos
from ..services.bureau_credito_service import (
    BureauCreditoService, 
    get_bureau_credito_service
)
from ..connectors.rules_bureau_connector import (
    BureauDataType,
    BureauRuleConfig,
    BureauRuleEvent,
    BureauRulesConnector,
    get_bureau_rules_connector,
)

# Importações do sistema de regras dinâmicas
from infrastructure.fraud_detection.rules_engine.evaluator import RuleEvaluator
from infrastructure.fraud_detection.rules_engine.models import (
    Event,
    Rule,
    RuleEvaluationResult,
    RuleSet,
)
from infrastructure.fraud_detection.neuraflow.rule_enhancer import RuleEnhancer
from infrastructure.fraud_detection.neuraflow.neuraflow_connector import NeuraFlowConnector
from infrastructure.fraud_detection.neuraflow.models import EnhancementConfig


class RulesTrustGuardEvent(BaseModel):
    """Evento para integração entre regras, Bureau de Créditos e TrustGuard"""
    user_id: str = Field(..., description="ID do usuário")
    document_id: str = Field(..., description="Número do documento (CPF/CNPJ)")
    event_type: str = Field(..., description="Tipo de evento")
    event_data: Dict[str, Any] = Field(..., description="Dados do evento")
    timestamp: datetime = Field(default_factory=datetime.now, description="Data/hora do evento")
    ip_address: Optional[str] = Field(None, description="Endereço IP")
    device_id: Optional[str] = Field(None, description="ID do dispositivo")
    user_agent: Optional[str] = Field(None, description="User-Agent")
    session_id: Optional[str] = Field(None, description="ID da sessão")
    resource_id: Optional[str] = Field(None, description="ID do recurso acessado")
    resource_type: Optional[str] = Field(None, description="Tipo do recurso acessado")
    action: Optional[str] = Field(None, description="Ação realizada")
    context: Dict[str, Any] = Field(default_factory=dict, description="Contexto adicional")


class RulesTrustGuardEvaluationResult(BaseModel):
    """Resultado da avaliação integrada entre regras, Bureau de Créditos e TrustGuard"""
    event_id: str = Field(..., description="ID do evento avaliado")
    timestamp: datetime = Field(default_factory=datetime.now, description="Data/hora da avaliação")
    trust_guard_decision: Optional[AccessDecisionResponse] = Field(None, 
                                                                description="Decisão do TrustGuard")
    rule_evaluation_results: Dict[str, RuleEvaluationResult] = Field(default_factory=dict, 
                                                                  description="Resultados da avaliação das regras")
    bureau_risk_level: Optional[RiskLevel] = Field(None, 
                                                description="Nível de risco do Bureau de Créditos")
    neuraflow_enhanced: bool = Field(False, 
                                   description="Indica se foi aprimorado pelo NeuraFlow")
    processing_time_ms: float = Field(..., description="Tempo de processamento em milissegundos")
    final_risk_level: RiskLevel = Field(..., description="Nível de risco final")
    risk_factors: List[Dict[str, Any]] = Field(default_factory=list, 
                                             description="Fatores de risco identificados")


class RulesTrustGuardIntegrator:
    """
    Integrador entre o sistema de regras dinâmicas, o Bureau de Créditos e o TrustGuard.
    
    Esta classe fornece:
    1. Avaliação integrada de eventos usando os três sistemas
    2. Atualização de risco do usuário no TrustGuard com base nas regras e Bureau
    3. Autorização avançada baseada em regras dinâmicas
    4. Enriquecimento de eventos com dados do Bureau e análise do NeuraFlow
    """
    
    def __init__(
        self,
        trust_guard_connector: TrustGuardConnector,
        bureau_rules_connector: BureauRulesConnector,
        neuraflow_connector: Optional[NeuraFlowConnector] = None,
        rule_evaluator: Optional[RuleEvaluator] = None,
        logger: Optional[logging.Logger] = None,
    ):
        """
        Inicializa o integrador.
        
        Args:
            trust_guard_connector: Conector do TrustGuard
            bureau_rules_connector: Conector de regras do Bureau
            neuraflow_connector: Conector do NeuraFlow
            rule_evaluator: Avaliador de regras
            logger: Logger para registrar eventos
        """
        self.trust_guard = trust_guard_connector
        self.bureau_rules = bureau_rules_connector
        self.neuraflow = neuraflow_connector
        self.rule_evaluator = rule_evaluator
        self.logger = logger or logging.getLogger(__name__)
        
        self.logger.info("Rules TrustGuard integrator initialized")
        
        # Mapeia níveis de risco do Bureau para TrustGuard
        self.risk_level_mapping = {
            "very_low": RiskLevel.VERY_LOW,
            "low": RiskLevel.LOW,
            "medium": RiskLevel.MEDIUM,
            "high": RiskLevel.HIGH,
            "very_high": RiskLevel.VERY_HIGH,
        }
    
    def _map_risk_level(self, risk_level: str) -> RiskLevel:
        """
        Mapeia um nível de risco string para o enum RiskLevel.
        
        Args:
            risk_level: Nível de risco em string
            
        Returns:
            RiskLevel: Nível de risco convertido
        """
        return self.risk_level_mapping.get(risk_level.lower(), RiskLevel.MEDIUM)
    
    def _create_bureau_rule_event(
        self,
        event: RulesTrustGuardEvent,
    ) -> BureauRuleEvent:
        """
        Cria um evento para o conector de regras do Bureau.
        
        Args:
            event: Evento original
            
        Returns:
            BureauRuleEvent: Evento convertido
        """
        return BureauRuleEvent(
            document_id=event.document_id,
            event_type=event.event_type,
            event_data=event.event_data,
            timestamp=event.timestamp,
            context={
                **event.context,
                "ip_address": event.ip_address,
                "device_id": event.device_id,
                "user_agent": event.user_agent,
                "session_id": event.session_id,
                "resource_id": event.resource_id,
                "resource_type": event.resource_type,
                "action": event.action,
            },
            metadata={
                "user_id": event.user_id,
                "original_event_id": str(event.timestamp.timestamp()) + "-" + event.user_id,
            },
        )
    
    def _create_trust_guard_user_context(
        self,
        event: RulesTrustGuardEvent,
    ) -> UserContext:
        """
        Cria um contexto de usuário para o TrustGuard.
        
        Args:
            event: Evento original
            
        Returns:
            UserContext: Contexto de usuário convertido
        """
        return UserContext(
            user_id=event.user_id,
            ip_address=event.ip_address,
            device_id=event.device_id,
            user_agent=event.user_agent,
            session_id=event.session_id,
            auth_methods=[],  # Será preenchido pelo TrustGuard
            auth_level="LOW",  # Valor padrão, será atualizado pelo TrustGuard
            roles=[],  # Será preenchido pelo TrustGuard
            groups=[],  # Será preenchido pelo TrustGuard
            permissions=[],  # Será preenchido pelo TrustGuard
            attributes={
                **event.context,
                "document_id": event.document_id,
                "event_type": event.event_type,
            },
        )
    
    def _extract_risk_factors(
        self,
        rule_results: Dict[str, RuleEvaluationResult],
    ) -> List[Dict[str, Any]]:
        """
        Extrai fatores de risco dos resultados da avaliação de regras.
        
        Args:
            rule_results: Resultados da avaliação de regras
            
        Returns:
            List[Dict[str, Any]]: Fatores de risco extraídos
        """
        risk_factors = []
        
        for rule_id, result in rule_results.items():
            if result.triggered:
                risk_factors.append({
                    "rule_id": rule_id,
                    "rule_name": result.rule_name,
                    "risk_level": result.metadata.get("risk_level", "medium"),
                    "description": result.metadata.get("description", "Regra dinâmica acionada"),
                    "score_impact": result.metadata.get("score_impact", 0.1),
                    "timestamp": datetime.now().isoformat(),
                })
        
        return risk_factors
    
    def _calculate_final_risk_level(
        self,
        rule_results: Dict[str, RuleEvaluationResult],
        bureau_risk_level: Optional[RiskLevel] = None,
    ) -> RiskLevel:
        """
        Calcula o nível de risco final com base nos resultados das regras e Bureau.
        
        Args:
            rule_results: Resultados da avaliação de regras
            bureau_risk_level: Nível de risco do Bureau
            
        Returns:
            RiskLevel: Nível de risco final
        """
        # Pontuação para cada nível de risco
        risk_scores = {
            "very_low": 1,
            "low": 2,
            "medium": 3,
            "high": 4,
            "very_high": 5,
        }
        
        # Peso dos fatores (regras têm peso maior que Bureau)
        rule_weight = 0.7
        bureau_weight = 0.3
        
        # Pontuação máxima de regras
        max_rule_score = 0
        
        for result in rule_results.values():
            if result.triggered:
                risk_level = result.metadata.get("risk_level", "medium").lower()
                score = risk_scores.get(risk_level, 3)
                max_rule_score = max(max_rule_score, score)
        
        # Se não houver regras acionadas, considera risco baixo
        if max_rule_score == 0:
            max_rule_score = 2  # low
        
        # Pontuação do Bureau
        bureau_score = 0
        
        if bureau_risk_level:
            bureau_score = risk_scores.get(bureau_risk_level.lower(), 3)
        
        # Cálculo ponderado
        if bureau_score > 0:
            final_score = rule_weight * max_rule_score + bureau_weight * bureau_score
        else:
            final_score = max_rule_score
        
        # Mapeamento para nível de risco
        if final_score < 1.5:
            return RiskLevel.VERY_LOW
        elif final_score < 2.5:
            return RiskLevel.LOW
        elif final_score < 3.5:
            return RiskLevel.MEDIUM
        elif final_score < 4.5:
            return RiskLevel.HIGH
        else:
            return RiskLevel.VERY_HIGH
    
    async def evaluate_event(
        self,
        event: RulesTrustGuardEvent,
        ruleset_id: str,
        bureau_config: Optional[BureauRuleConfig] = None,
        update_trust_guard: bool = True,
    ) -> RulesTrustGuardEvaluationResult:
        """
        Avalia um evento de forma integrada entre regras, Bureau e TrustGuard.
        
        Args:
            event: Evento a ser avaliado
            ruleset_id: ID do conjunto de regras a ser utilizado
            bureau_config: Configuração do Bureau (opcional)
            update_trust_guard: Indica se deve atualizar o TrustGuard
            
        Returns:
            RulesTrustGuardEvaluationResult: Resultado da avaliação integrada
        """
        start_time = time.time()
        event_id = f"{event.timestamp.timestamp()}-{event.user_id}"
        
        self.logger.info(f"Evaluating event {event_id} with ruleset {ruleset_id}")
        
        # Resultados da avaliação
        rule_results = {}
        bureau_risk_level = None
        trust_guard_decision = None
        neuraflow_enhanced = False
        
        try:
            # 1. Buscar o conjunto de regras
            if not self.rule_evaluator:
                raise ValueError("Rule evaluator not initialized")
            
            ruleset = await self.rule_evaluator.get_ruleset(ruleset_id)
            
            if not ruleset:
                raise ValueError(f"Ruleset {ruleset_id} not found")
            
            # 2. Avaliar com Bureau de Créditos (se configurado)
            if bureau_config and self.bureau_rules:
                bureau_event = self._create_bureau_rule_event(event)
                
                try:
                    bureau_results = await self.bureau_rules.evaluate_ruleset_with_bureau_data(
                        ruleset=ruleset,
                        event=bureau_event,
                        bureau_config=bureau_config,
                    )
                    
                    rule_results = bureau_results
                    
                    # Extrair nível de risco do Bureau
                    if "bureau_data" in bureau_event.context:
                        bureau_data = bureau_event.context["bureau_data"]
                        
                        if "risk_level" in bureau_data:
                            bureau_risk_level = self._map_risk_level(bureau_data["risk_level"])
                    
                except Exception as e:
                    self.logger.error(f"Error evaluating with Bureau: {str(e)}")
                    
                    # Continua com a avaliação padrão em caso de erro
                    standard_event = Event(
                        id=event_id,
                        type=event.event_type,
                        data=event.event_data,
                        timestamp=event.timestamp,
                        context=event.context,
                    )
                    
                    rule_results = await self.rule_evaluator.evaluate_ruleset(ruleset, standard_event)
            
            # 3. Avaliar com NeuraFlow (se disponível)
            elif self.neuraflow and self.rule_evaluator:
                try:
                    # Criar evento padrão
                    standard_event = Event(
                        id=event_id,
                        type=event.event_type,
                        data=event.event_data,
                        timestamp=event.timestamp,
                        context={
                            **event.context,
                            "user_id": event.user_id,
                            "document_id": event.document_id,
                            "ip_address": event.ip_address,
                            "device_id": event.device_id,
                            "user_agent": event.user_agent,
                        },
                    )
                    
                    # Configurar enriquecimento
                    enhancement_config = EnhancementConfig(
                        enhance_user_behavior=True,
                        enhance_device_info=bool(event.device_id),
                        enhance_location=bool(event.ip_address),
                        enhance_transaction_risk=True,
                    )
                    
                    # Enriquecer evento com NeuraFlow
                    enhanced_event = await self.neuraflow.enhance_event(
                        event=standard_event,
                        config=enhancement_config,
                    )
                    
                    # Avaliar regras com evento enriquecido
                    rule_results = await self.rule_evaluator.evaluate_ruleset(ruleset, enhanced_event)
                    neuraflow_enhanced = True
                    
                except Exception as e:
                    self.logger.error(f"Error enhancing with NeuraFlow: {str(e)}")
                    
                    # Continua com a avaliação padrão em caso de erro
                    standard_event = Event(
                        id=event_id,
                        type=event.event_type,
                        data=event.event_data,
                        timestamp=event.timestamp,
                        context=event.context,
                    )
                    
                    rule_results = await self.rule_evaluator.evaluate_ruleset(ruleset, standard_event)
            
            # 4. Avaliação padrão de regras (fallback)
            else:
                standard_event = Event(
                    id=event_id,
                    type=event.event_type,
                    data=event.event_data,
                    timestamp=event.timestamp,
                    context=event.context,
                )
                
                rule_results = await self.rule_evaluator.evaluate_ruleset(ruleset, standard_event)
            
            # 5. Extrair fatores de risco das regras acionadas
            risk_factors = self._extract_risk_factors(rule_results)
            
            # 6. Calcular nível de risco final
            final_risk_level = self._calculate_final_risk_level(
                rule_results=rule_results,
                bureau_risk_level=bureau_risk_level,
            )
            
            # 7. Atualizar o TrustGuard (se solicitado)
            if update_trust_guard and self.trust_guard:
                try:
                    # Atualizar perfil de risco do usuário
                    await self.trust_guard.update_user_risk(
                        user_id=event.user_id,
                        risk_level=final_risk_level,
                        risk_score=self._risk_level_to_score(final_risk_level),
                        risk_factors=risk_factors,
                    )
                    
                    # Se houver ID de sessão, atualizar contexto da sessão
                    if event.session_id:
                        await self.trust_guard.update_session(
                            session_id=event.session_id,
                            updates={
                                "risk_level": final_risk_level,
                                "context": {
                                    "last_evaluated_event": event_id,
                                    "rules_evaluated": len(rule_results),
                                    "rules_triggered": sum(1 for r in rule_results.values() if r.triggered),
                                    "neuraflow_enhanced": neuraflow_enhanced,
                                    "bureau_data_used": bureau_config is not None,
                                },
                            },
                        )
                    
                except Exception as e:
                    self.logger.error(f"Error updating TrustGuard: {str(e)}")
            
            # 8. Medir tempo de processamento
            processing_time = (time.time() - start_time) * 1000  # ms
            
            # 9. Construir resultado
            return RulesTrustGuardEvaluationResult(
                event_id=event_id,
                trust_guard_decision=trust_guard_decision,
                rule_evaluation_results=rule_results,
                bureau_risk_level=bureau_risk_level,
                neuraflow_enhanced=neuraflow_enhanced,
                processing_time_ms=processing_time,
                final_risk_level=final_risk_level,
                risk_factors=risk_factors,
            )
            
        except Exception as e:
            self.logger.error(f"Error in integrated evaluation: {str(e)}")
            
            # Medir tempo de processamento mesmo em caso de erro
            processing_time = (time.time() - start_time) * 1000  # ms
            
            # Retornar resultado com erro
            return RulesTrustGuardEvaluationResult(
                event_id=event_id,
                trust_guard_decision=None,
                rule_evaluation_results={},
                bureau_risk_level=None,
                neuraflow_enhanced=False,
                processing_time_ms=processing_time,
                final_risk_level=RiskLevel.MEDIUM,  # Valor padrão em caso de erro
                risk_factors=[{
                    "rule_id": "error",
                    "rule_name": "Error",
                    "risk_level": "medium",
                    "description": f"Error in evaluation: {str(e)}",
                    "score_impact": 0.0,
                    "timestamp": datetime.now().isoformat(),
                }],
            )
    
    def _risk_level_to_score(self, risk_level: RiskLevel) -> float:
        """
        Converte um nível de risco para um score numérico.
        
        Args:
            risk_level: Nível de risco
            
        Returns:
            float: Score de risco (0.0 - 1.0)
        """
        mapping = {
            RiskLevel.VERY_LOW: 0.1,
            RiskLevel.LOW: 0.3,
            RiskLevel.MEDIUM: 0.5,
            RiskLevel.HIGH: 0.7,
            RiskLevel.VERY_HIGH: 0.9,
        }
        
        return mapping.get(risk_level, 0.5)
    
    async def evaluate_access_request(
        self,
        event: RulesTrustGuardEvent,
        ruleset_id: str,
        bureau_config: Optional[BureauRuleConfig] = None,
    ) -> AccessDecisionResponse:
        """
        Avalia uma solicitação de acesso com regras, Bureau e TrustGuard.
        
        Args:
            event: Evento de acesso
            ruleset_id: ID do conjunto de regras a ser utilizado
            bureau_config: Configuração do Bureau (opcional)
            
        Returns:
            AccessDecisionResponse: Decisão de acesso
        """
        if not self.trust_guard:
            raise ValueError("TrustGuard connector not initialized")
        
        # 1. Avaliar evento com regras e Bureau
        evaluation_result = await self.evaluate_event(
            event=event,
            ruleset_id=ruleset_id,
            bureau_config=bureau_config,
            update_trust_guard=True,  # Atualiza o TrustGuard automaticamente
        )
        
        # 2. Criar solicitação de acesso para o TrustGuard
        user_context = self._create_trust_guard_user_context(event)
        
        # Adicionar nível de risco ao contexto
        user_context.risk_profile = UserRisk(
            user_id=event.user_id,
            risk_level=evaluation_result.final_risk_level,
            risk_score=self._risk_level_to_score(evaluation_result.final_risk_level),
            risk_factors=evaluation_result.risk_factors,
        )
        
        # Criar contexto de recurso
        from .trust_guard_models import ResourceContext, ResourceType, ActionType
        
        resource_context = ResourceContext(
            resource_id=event.resource_id or f"resource-{event.event_type}",
            resource_type=ResourceType(event.resource_type.lower()) if event.resource_type else ResourceType.SERVICE,
            resource_path=event.context.get("resource_path"),
            action=ActionType(event.action.lower()) if event.action else ActionType.READ,
            metadata={
                "event_type": event.event_type,
                "rules_evaluated": len(evaluation_result.rule_evaluation_results),
                "rules_triggered": sum(
                    1 for r in evaluation_result.rule_evaluation_results.values() 
                    if r.triggered
                ),
            },
        )
        
        # Criar solicitação de acesso
        access_request = AccessRequest(
            user_context=user_context,
            resource_context=resource_context,
            session_data={
                "session_id": event.session_id,
            } if event.session_id else {},
            transaction_data=event.event_data,
            environment={
                "source_ip": event.ip_address,
                "device_id": event.device_id,
                "user_agent": event.user_agent,
                "timestamp": event.timestamp.isoformat(),
            },
        )
        
        # 3. Enviar solicitação de acesso ao TrustGuard
        try:
            decision = await self.trust_guard.evaluate_access(access_request)
            return decision
            
        except Exception as e:
            self.logger.error(f"Error evaluating access with TrustGuard: {str(e)}")
            
            # Fallback para decisão baseada apenas nas regras
            from .trust_guard_models import AccessDecisionResponse, AccessDecision
            
            # Se alguma regra de alto risco foi acionada, nega o acesso
            high_risk_triggered = False
            
            for result in evaluation_result.rule_evaluation_results.values():
                if result.triggered:
                    risk_level = result.metadata.get("risk_level", "").lower()
                    if risk_level in ["high", "very_high"]:
                        high_risk_triggered = True
                        break
            
            decision = AccessDecision.DENY if high_risk_triggered else AccessDecision.ALLOW
            
            return AccessDecisionResponse(
                request_id=str(uuid4()),
                decision=decision,
                reason="Fallback decision due to TrustGuard error",
                risk_level=evaluation_result.final_risk_level,
                auth_level=user_context.auth_level,
            )


# Factory para criação do integrador
async def get_rules_trustguard_integrator(
    trust_guard_connector: TrustGuardConnector = Depends(get_trust_guard_connector),
    bureau_rules_connector: BureauRulesConnector = Depends(get_bureau_rules_connector),
    neuraflow_connector: Optional[NeuraFlowConnector] = None,
    rule_evaluator: Optional[RuleEvaluator] = None,
) -> RulesTrustGuardIntegrator:
    """
    Factory para criação do integrador.
    
    Args:
        trust_guard_connector: Conector do TrustGuard
        bureau_rules_connector: Conector de regras do Bureau
        neuraflow_connector: Conector do NeuraFlow
        rule_evaluator: Avaliador de regras
        
    Returns:
        RulesTrustGuardIntegrator: Integrador configurado
    """
    logger = logging.getLogger("rules_trustguard_integrator")
    
    integrator = RulesTrustGuardIntegrator(
        trust_guard_connector=trust_guard_connector,
        bureau_rules_connector=bureau_rules_connector,
        neuraflow_connector=neuraflow_connector,
        rule_evaluator=rule_evaluator,
        logger=logger,
    )
    
    return integrator