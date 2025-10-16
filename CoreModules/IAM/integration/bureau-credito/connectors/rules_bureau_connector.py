"""
Conector para integração entre o sistema de regras dinâmicas e o Bureau de Créditos.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Set, Tuple, Union

import httpx
from fastapi import Depends, FastAPI, HTTPException
from pydantic import BaseModel, Field, validator

# Importações do serviço de Bureau de Créditos
from ..services.bureau_credito_service import (
    BureauCreditoService,
    get_bureau_credito_service,
)
from ..models import (
    CreditReport,
    CreditRisk,
    CreditScore,
    FinancialProfile,
    FraudIndicator,
    IdentityVerification,
    RiskLevel,
    TransactionHistory,
)

# Importações do sistema de regras dinâmicas
# Nota: Ajustar caminho conforme estrutura do projeto
from infrastructure.fraud_detection.neuraflow.rule_enhancer import RuleEnhancer
from infrastructure.fraud_detection.rules_engine.evaluator import RuleEvaluator
from infrastructure.fraud_detection.rules_engine.models import (
    Rule,
    RuleSet,
    RuleEvaluationResult,
    Event,
)


class BureauDataType(str, Enum):
    """Tipos de dados do Bureau de Créditos"""
    CREDIT_SCORE = "credit_score"
    CREDIT_REPORT = "credit_report"
    CREDIT_RISK = "credit_risk"
    IDENTITY_VERIFICATION = "identity_verification"
    FRAUD_INDICATORS = "fraud_indicators"
    TRANSACTION_HISTORY = "transaction_history"
    FINANCIAL_PROFILE = "financial_profile"


class BureauRuleConfig(BaseModel):
    """Configuração para regra integrada com Bureau de Créditos"""
    data_type: BureauDataType = Field(..., description="Tipo de dado do Bureau")
    provider: str = Field(..., description="Provedor do Bureau")
    cache_ttl: int = Field(300, description="Tempo de vida do cache em segundos")
    region: Optional[str] = Field(None, description="Região para contexto regional")
    additional_params: Dict[str, Any] = Field(default_factory=dict, 
                                             description="Parâmetros adicionais")


class BureauRuleEvent(BaseModel):
    """Evento para avaliação de regra com Bureau de Créditos"""
    document_id: str = Field(..., description="Número do documento (CPF/CNPJ)")
    event_type: str = Field(..., description="Tipo de evento")
    event_data: Dict[str, Any] = Field(..., description="Dados do evento")
    timestamp: datetime = Field(default_factory=datetime.now, description="Data/hora do evento")
    context: Dict[str, Any] = Field(default_factory=dict, description="Contexto do evento")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados do evento")
    
    def to_event(self) -> Event:
        """
        Converte para o modelo Event do sistema de regras.
        
        Returns:
            Event: Evento convertido para o modelo do sistema de regras
        """
        return Event(
            id=self.document_id,
            type=self.event_type,
            data=self.event_data,
            timestamp=self.timestamp,
            context=self.context,
            metadata=self.metadata,
        )


class BureauRuleEnrichmentResult(BaseModel):
    """Resultado do enriquecimento de um evento com dados do Bureau"""
    event: BureauRuleEvent
    enriched_data: Dict[str, Any]
    data_source: BureauDataType
    provider: str
    timestamp: datetime = Field(default_factory=datetime.now)
    cache_hit: bool = False
    processing_time_ms: float


class BureauRulesConnector:
    """
    Conector para integração entre o sistema de regras dinâmicas e o Bureau de Créditos.
    
    Este conector permite:
    1. Enriquecer eventos com dados do Bureau de Créditos
    2. Avaliar regras com dados enriquecidos
    3. Criar regras baseadas em padrões do Bureau
    4. Monitorar alterações em scores de crédito
    """
    
    def __init__(
        self,
        bureau_service: BureauCreditoService,
        rule_enhancer: Optional[RuleEnhancer] = None,
        rule_evaluator: Optional[RuleEvaluator] = None,
        logger: Optional[logging.Logger] = None,
        cache_enabled: bool = True,
        cache_ttl: int = 300,  # 5 minutos
    ):
        """
        Inicializa o conector.
        
        Args:
            bureau_service: Serviço de Bureau de Créditos
            rule_enhancer: Enhancer de regras
            rule_evaluator: Avaliador de regras
            logger: Logger para registrar eventos
            cache_enabled: Indica se o cache está habilitado
            cache_ttl: Tempo de vida do cache em segundos
        """
        self.bureau_service = bureau_service
        self.rule_enhancer = rule_enhancer
        self.rule_evaluator = rule_evaluator
        self.logger = logger or logging.getLogger(__name__)
        self.cache_enabled = cache_enabled
        self.cache_ttl = cache_ttl
        
        # Cache para resultados de consultas ao Bureau
        # Formato: {data_type}:{provider}:{document_id} -> (timestamp, data)
        self._cache: Dict[str, Tuple[float, Any]] = {}
        
        self.logger.info("Bureau Rules Connector initialized")
    
    def _get_cache_key(
        self,
        data_type: BureauDataType,
        provider: str,
        document_id: str,
    ) -> str:
        """
        Obtém a chave de cache para uma consulta.
        
        Args:
            data_type: Tipo de dado do Bureau
            provider: Provedor do Bureau
            document_id: Número do documento
            
        Returns:
            str: Chave de cache
        """
        return f"{data_type}:{provider}:{document_id}"
    
    def _get_from_cache(
        self,
        data_type: BureauDataType,
        provider: str,
        document_id: str,
    ) -> Tuple[bool, Any]:
        """
        Obtém dados do cache.
        
        Args:
            data_type: Tipo de dado do Bureau
            provider: Provedor do Bureau
            document_id: Número do documento
            
        Returns:
            Tuple[bool, Any]: (cache_hit, data)
        """
        if not self.cache_enabled:
            return False, None
        
        cache_key = self._get_cache_key(data_type, provider, document_id)
        cache_entry = self._cache.get(cache_key)
        
        if not cache_entry:
            return False, None
        
        timestamp, data = cache_entry
        now = time.time()
        
        if now - timestamp > self.cache_ttl:
            # Cache expirado
            del self._cache[cache_key]
            return False, None
        
        return True, data
    
    def _set_cache(
        self,
        data_type: BureauDataType,
        provider: str,
        document_id: str,
        data: Any,
    ) -> None:
        """
        Armazena dados no cache.
        
        Args:
            data_type: Tipo de dado do Bureau
            provider: Provedor do Bureau
            document_id: Número do documento
            data: Dados a serem armazenados
        """
        if not self.cache_enabled:
            return
        
        cache_key = self._get_cache_key(data_type, provider, document_id)
        self._cache[cache_key] = (time.time(), data)
    
    async def get_credit_score(
        self,
        document_id: str,
        provider: str,
        region: Optional[str] = None,
    ) -> CreditScore:
        """
        Obtém o score de crédito de um cliente, usando cache se disponível.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            region: Região para contexto regional
            
        Returns:
            CreditScore: Informações de score de crédito
        """
        cache_hit, data = self._get_from_cache(
            BureauDataType.CREDIT_SCORE,
            provider,
            document_id,
        )
        
        if cache_hit:
            return data
        
        result = await self.bureau_service.get_credit_score(
            document_id=document_id,
            provider=provider,
            region=region,
        )
        
        self._set_cache(
            BureauDataType.CREDIT_SCORE,
            provider,
            document_id,
            result,
        )
        
        return result
    
    async def get_credit_risk(
        self,
        document_id: str,
        provider: str,
        transaction_amount: Optional[float] = None,
        transaction_type: Optional[str] = None,
        region: Optional[str] = None,
    ) -> CreditRisk:
        """
        Avalia o risco de crédito de um cliente, usando cache se disponível.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            transaction_amount: Valor da transação para avaliação contextual
            transaction_type: Tipo da transação para avaliação contextual
            region: Região para contexto regional
            
        Returns:
            CreditRisk: Avaliação de risco de crédito
        """
        # Para risco de crédito com transação, não usamos cache
        if transaction_amount or transaction_type:
            return await self.bureau_service.get_credit_risk(
                document_id=document_id,
                provider=provider,
                transaction_amount=transaction_amount,
                transaction_type=transaction_type,
                region=region,
            )
        
        cache_hit, data = self._get_from_cache(
            BureauDataType.CREDIT_RISK,
            provider,
            document_id,
        )
        
        if cache_hit:
            return data
        
        result = await self.bureau_service.get_credit_risk(
            document_id=document_id,
            provider=provider,
            region=region,
        )
        
        self._set_cache(
            BureauDataType.CREDIT_RISK,
            provider,
            document_id,
            result,
        )
        
        return result
    
    async def check_fraud_indicators(
        self,
        document_id: str,
        provider: str,
        ip_address: Optional[str] = None,
        device_id: Optional[str] = None,
        user_agent: Optional[str] = None,
        region: Optional[str] = None,
    ) -> FraudIndicator:
        """
        Verifica indicadores de fraude para um cliente, usando cache se disponível.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            ip_address: Endereço IP do cliente
            device_id: ID do dispositivo
            user_agent: User-Agent do navegador
            region: Região para contexto regional
            
        Returns:
            FraudIndicator: Indicadores de fraude
        """
        # Para indicadores de fraude com dados contextuais, não usamos cache
        if ip_address or device_id or user_agent:
            return await self.bureau_service.check_fraud_indicators(
                document_id=document_id,
                provider=provider,
                ip_address=ip_address,
                device_id=device_id,
                user_agent=user_agent,
                region=region,
            )
        
        cache_hit, data = self._get_from_cache(
            BureauDataType.FRAUD_INDICATORS,
            provider,
            document_id,
        )
        
        if cache_hit:
            return data
        
        result = await self.bureau_service.check_fraud_indicators(
            document_id=document_id,
            provider=provider,
            region=region,
        )
        
        self._set_cache(
            BureauDataType.FRAUD_INDICATORS,
            provider,
            document_id,
            result,
        )
        
        return result
    
    async def enrich_event(
        self,
        event: BureauRuleEvent,
        bureau_config: BureauRuleConfig,
    ) -> BureauRuleEnrichmentResult:
        """
        Enriquece um evento com dados do Bureau de Créditos.
        
        Args:
            event: Evento a ser enriquecido
            bureau_config: Configuração do Bureau
            
        Returns:
            BureauRuleEnrichmentResult: Resultado do enriquecimento
        """
        document_id = event.document_id
        provider = bureau_config.provider
        data_type = bureau_config.data_type
        
        start_time = time.time()
        cache_hit = False
        enriched_data = {}
        
        try:
            if data_type == BureauDataType.CREDIT_SCORE:
                cache_hit, data = self._get_from_cache(data_type, provider, document_id)
                
                if not cache_hit:
                    data = await self.bureau_service.get_credit_score(
                        document_id=document_id,
                        provider=provider,
                        region=bureau_config.region,
                    )
                    self._set_cache(data_type, provider, document_id, data)
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.CREDIT_RISK:
                # Obter parâmetros adicionais
                transaction_amount = bureau_config.additional_params.get("transaction_amount")
                transaction_type = bureau_config.additional_params.get("transaction_type")
                
                # Se tiver parâmetros específicos, não usa cache
                if transaction_amount or transaction_type:
                    data = await self.bureau_service.get_credit_risk(
                        document_id=document_id,
                        provider=provider,
                        transaction_amount=transaction_amount,
                        transaction_type=transaction_type,
                        region=bureau_config.region,
                    )
                else:
                    cache_hit, data = self._get_from_cache(data_type, provider, document_id)
                    
                    if not cache_hit:
                        data = await self.bureau_service.get_credit_risk(
                            document_id=document_id,
                            provider=provider,
                            region=bureau_config.region,
                        )
                        self._set_cache(data_type, provider, document_id, data)
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.FRAUD_INDICATORS:
                # Obter parâmetros adicionais
                ip_address = bureau_config.additional_params.get("ip_address")
                device_id = bureau_config.additional_params.get("device_id")
                user_agent = bureau_config.additional_params.get("user_agent")
                
                # Se tiver parâmetros específicos, não usa cache
                if ip_address or device_id or user_agent:
                    data = await self.bureau_service.check_fraud_indicators(
                        document_id=document_id,
                        provider=provider,
                        ip_address=ip_address,
                        device_id=device_id,
                        user_agent=user_agent,
                        region=bureau_config.region,
                    )
                else:
                    cache_hit, data = self._get_from_cache(data_type, provider, document_id)
                    
                    if not cache_hit:
                        data = await self.bureau_service.check_fraud_indicators(
                            document_id=document_id,
                            provider=provider,
                            region=bureau_config.region,
                        )
                        self._set_cache(data_type, provider, document_id, data)
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.IDENTITY_VERIFICATION:
                # Obter parâmetros adicionais
                name = bureau_config.additional_params.get("name")
                birth_date = bureau_config.additional_params.get("birth_date")
                mother_name = bureau_config.additional_params.get("mother_name")
                phone_number = bureau_config.additional_params.get("phone_number")
                
                # Para verificação de identidade, sempre consulta o serviço
                data = await self.bureau_service.verify_identity(
                    document_id=document_id,
                    provider=provider,
                    name=name,
                    birth_date=birth_date,
                    mother_name=mother_name,
                    phone_number=phone_number,
                    region=bureau_config.region,
                )
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.FINANCIAL_PROFILE:
                cache_hit, data = self._get_from_cache(data_type, provider, document_id)
                
                if not cache_hit:
                    data = await self.bureau_service.get_financial_profile(
                        document_id=document_id,
                        provider=provider,
                        region=bureau_config.region,
                    )
                    self._set_cache(data_type, provider, document_id, data)
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.TRANSACTION_HISTORY:
                # Obter parâmetros adicionais
                start_date = bureau_config.additional_params.get("start_date")
                end_date = bureau_config.additional_params.get("end_date")
                limit = bureau_config.additional_params.get("limit", 50)
                
                # Para histórico de transações, sempre consulta o serviço
                data = await self.bureau_service.get_transaction_history(
                    document_id=document_id,
                    provider=provider,
                    start_date=start_date,
                    end_date=end_date,
                    limit=limit,
                    region=bureau_config.region,
                )
                
                enriched_data = data.dict()
                
            elif data_type == BureauDataType.CREDIT_REPORT:
                # Obter parâmetros adicionais
                include_details = bureau_config.additional_params.get("include_details", False)
                
                cache_hit, data = self._get_from_cache(data_type, provider, document_id)
                
                if not cache_hit:
                    data = await self.bureau_service.get_credit_report(
                        document_id=document_id,
                        provider=provider,
                        include_details=include_details,
                        region=bureau_config.region,
                    )
                    self._set_cache(data_type, provider, document_id, data)
                
                enriched_data = data.dict()
            
            processing_time = (time.time() - start_time) * 1000  # ms
            
            return BureauRuleEnrichmentResult(
                event=event,
                enriched_data=enriched_data,
                data_source=data_type,
                provider=provider,
                cache_hit=cache_hit,
                processing_time_ms=processing_time,
            )
            
        except Exception as e:
            self.logger.error(f"Error enriching event: {str(e)}")
            processing_time = (time.time() - start_time) * 1000  # ms
            
            return BureauRuleEnrichmentResult(
                event=event,
                enriched_data={"error": str(e)},
                data_source=data_type,
                provider=provider,
                cache_hit=False,
                processing_time_ms=processing_time,
            )
    
    async def evaluate_rule_with_bureau_data(
        self,
        rule: Rule,
        event: BureauRuleEvent,
        bureau_config: BureauRuleConfig,
    ) -> RuleEvaluationResult:
        """
        Avalia uma regra com dados do Bureau de Créditos.
        
        Args:
            rule: Regra a ser avaliada
            event: Evento a ser avaliado
            bureau_config: Configuração do Bureau
            
        Returns:
            RuleEvaluationResult: Resultado da avaliação
        """
        if not self.rule_evaluator:
            raise ValueError("Rule evaluator not initialized")
        
        # Enriquecer evento com dados do Bureau
        enrichment_result = await self.enrich_event(event, bureau_config)
        
        # Criar evento enriquecido para o sistema de regras
        original_event = event.to_event()
        
        # Adicionar dados enriquecidos ao contexto do evento
        enriched_event = Event(
            id=original_event.id,
            type=original_event.type,
            data=original_event.data,
            timestamp=original_event.timestamp,
            context={
                **original_event.context,
                "bureau_data": enrichment_result.enriched_data,
                "bureau_provider": bureau_config.provider,
                "bureau_data_type": bureau_config.data_type,
            },
            metadata={
                **original_event.metadata,
                "bureau_enrichment": {
                    "provider": bureau_config.provider,
                    "data_type": bureau_config.data_type,
                    "cache_hit": enrichment_result.cache_hit,
                    "processing_time_ms": enrichment_result.processing_time_ms,
                },
            },
        )
        
        # Avaliar regra com dados enriquecidos
        result = await self.rule_evaluator.evaluate_rule(rule, enriched_event)
        return result
    
    async def evaluate_ruleset_with_bureau_data(
        self,
        ruleset: RuleSet,
        event: BureauRuleEvent,
        bureau_config: BureauRuleConfig,
    ) -> Dict[str, RuleEvaluationResult]:
        """
        Avalia um conjunto de regras com dados do Bureau de Créditos.
        
        Args:
            ruleset: Conjunto de regras a ser avaliado
            event: Evento a ser avaliado
            bureau_config: Configuração do Bureau
            
        Returns:
            Dict[str, RuleEvaluationResult]: Resultados da avaliação por ID de regra
        """
        if not self.rule_evaluator:
            raise ValueError("Rule evaluator not initialized")
        
        # Enriquecer evento com dados do Bureau
        enrichment_result = await self.enrich_event(event, bureau_config)
        
        # Criar evento enriquecido para o sistema de regras
        original_event = event.to_event()
        
        # Adicionar dados enriquecidos ao contexto do evento
        enriched_event = Event(
            id=original_event.id,
            type=original_event.type,
            data=original_event.data,
            timestamp=original_event.timestamp,
            context={
                **original_event.context,
                "bureau_data": enrichment_result.enriched_data,
                "bureau_provider": bureau_config.provider,
                "bureau_data_type": bureau_config.data_type,
            },
            metadata={
                **original_event.metadata,
                "bureau_enrichment": {
                    "provider": bureau_config.provider,
                    "data_type": bureau_config.data_type,
                    "cache_hit": enrichment_result.cache_hit,
                    "processing_time_ms": enrichment_result.processing_time_ms,
                },
            },
        )
        
        # Avaliar conjunto de regras com dados enriquecidos
        results = await self.rule_evaluator.evaluate_ruleset(ruleset, enriched_event)
        return results
    
    async def evaluate_rule_with_neuraflow_bureau_enhancement(
        self,
        rule: Rule,
        event: BureauRuleEvent,
        bureau_config: BureauRuleConfig,
    ) -> RuleEvaluationResult:
        """
        Avalia uma regra com dados do Bureau de Créditos e enriquecimento do NeuraFlow.
        
        Args:
            rule: Regra a ser avaliada
            event: Evento a ser avaliado
            bureau_config: Configuração do Bureau
            
        Returns:
            RuleEvaluationResult: Resultado da avaliação
        """
        if not self.rule_evaluator:
            raise ValueError("Rule evaluator not initialized")
            
        if not self.rule_enhancer:
            raise ValueError("Rule enhancer not initialized")
        
        # Enriquecer evento com dados do Bureau
        bureau_enrichment_result = await self.enrich_event(event, bureau_config)
        
        # Criar evento enriquecido para o sistema de regras
        original_event = event.to_event()
        
        # Adicionar dados enriquecidos ao contexto do evento
        bureau_enriched_event = Event(
            id=original_event.id,
            type=original_event.type,
            data=original_event.data,
            timestamp=original_event.timestamp,
            context={
                **original_event.context,
                "bureau_data": bureau_enrichment_result.enriched_data,
                "bureau_provider": bureau_config.provider,
                "bureau_data_type": bureau_config.data_type,
            },
            metadata={
                **original_event.metadata,
                "bureau_enrichment": {
                    "provider": bureau_config.provider,
                    "data_type": bureau_config.data_type,
                    "cache_hit": bureau_enrichment_result.cache_hit,
                    "processing_time_ms": bureau_enrichment_result.processing_time_ms,
                },
            },
        )
        
        # Enriquecer evento com NeuraFlow
        neuraflow_enhanced_event = await self.rule_enhancer.enhance_event(bureau_enriched_event)
        
        # Avaliar regra com dados duplamente enriquecidos
        result = await self.rule_evaluator.evaluate_rule(rule, neuraflow_enhanced_event)
        return result


# Factory para criação do conector
async def get_bureau_rules_connector(
    bureau_service: BureauCreditoService = Depends(get_bureau_credito_service),
    rule_enhancer: Optional[RuleEnhancer] = None,
    rule_evaluator: Optional[RuleEvaluator] = None,
) -> BureauRulesConnector:
    """
    Factory para criação do conector.
    
    Args:
        bureau_service: Serviço de Bureau de Créditos
        rule_enhancer: Enhancer de regras
        rule_evaluator: Avaliador de regras
        
    Returns:
        BureauRulesConnector: Conector configurado
    """
    logger = logging.getLogger("bureau_rules_connector")
    
    cache_enabled = os.environ.get("BUREAU_CACHE_ENABLED", "true").lower() == "true"
    cache_ttl = int(os.environ.get("BUREAU_CACHE_TTL", "300"))
    
    connector = BureauRulesConnector(
        bureau_service=bureau_service,
        rule_enhancer=rule_enhancer,
        rule_evaluator=rule_evaluator,
        logger=logger,
        cache_enabled=cache_enabled,
        cache_ttl=cache_ttl,
    )
    
    return connector