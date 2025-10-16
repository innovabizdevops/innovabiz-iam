"""
TrustGuard - Motor de Pontuação de Confiança Multi-Dimensional

Este módulo implementa o motor de pontuação de confiança TrustScore, que avalia
a confiabilidade de usuários em múltiplas dimensões para a plataforma INNOVABIZ.

O motor integra dados de identidade, comportamento, financeiros, contextuais e
de reputação para gerar uma pontuação de confiança abrangente, adaptada a
contextos regionais específicos (CPLP, SADC, PALOP, etc.).

Autor: Eduardo Jeremias
Data: 20/08/2025
Versão: 1.0.0
"""

import asyncio
import json
import logging
import time
import uuid
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional, Set, Tuple, Union

import numpy as np
from fastapi import HTTPException
from opentelemetry import metrics, trace
from pydantic import ValidationError

from .trust_guard_models import (AnomalyType, DetectedAnomaly, FactorType, 
                                RegionalConfig, TenantTrustConfig, 
                                TrustDimension, TrustLevel, TrustScoreFactorModel,
                                TrustScoreHistoryItem, TrustScoreRequest,
                                TrustScoreResponse, UserTrustProfile,
                                VerificationType, AnomalySeverity)


# Configuração de logging
logger = logging.getLogger("trustguard.score_engine")

# Configuração de tracing
tracer = trace.get_tracer("trustguard.score_engine")

# Configuração de métricas
meter = metrics.get_meter("trustguard.score_engine")

# Métricas
trust_score_counter = meter.create_counter(
    name="trustguard.score_evaluations",
    description="Número total de avaliações de pontuação de confiança",
    unit="1",
)

trust_score_histogram = meter.create_histogram(
    name="trustguard.score_values",
    description="Distribuição das pontuações de confiança",
    unit="1",
)

anomaly_detection_counter = meter.create_counter(
    name="trustguard.detected_anomalies",
    description="Número de anomalias detectadas",
    unit="1",
)

processing_time_histogram = meter.create_histogram(
    name="trustguard.processing_time_ms",
    description="Tempo de processamento das avaliações de confiança",
    unit="ms",
)


class TrustScoreEngine:
    """
    Motor principal de avaliação de pontuação de confiança.
    
    Responsável por calcular pontuações de confiança multi-dimensionais para usuários
    baseado em diversos fatores e adaptado a contextos regionais específicos.
    """
    
    def __init__(self, db_connector=None, cache_provider=None, config_provider=None,
                 anomaly_detector=None, regional_config_provider=None):
        """
        Inicializa o motor de pontuação de confiança.
        
        Args:
            db_connector: Conector para banco de dados para armazenamento persistente
            cache_provider: Provedor de cache para melhorar performance
            config_provider: Provedor de configuração para configurações de tenant
            anomaly_detector: Detector de anomalias para identificar comportamentos suspeitos
            regional_config_provider: Provedor de configurações regionais
        """
        self.db_connector = db_connector
        self.cache_provider = cache_provider
        self.config_provider = config_provider
        self.anomaly_detector = anomaly_detector
        self.regional_config_provider = regional_config_provider
        self.executor = ThreadPoolExecutor(max_workers=10)
        
        # Inicializar mapeamento de dimensões para avaliadores
        self._dimension_evaluators = {
            TrustDimension.IDENTITY: self._evaluate_identity_dimension,
            TrustDimension.BEHAVIORAL: self._evaluate_behavioral_dimension,
            TrustDimension.FINANCIAL: self._evaluate_financial_dimension,
            TrustDimension.CONTEXTUAL: self._evaluate_contextual_dimension,
            TrustDimension.REPUTATION: self._evaluate_reputation_dimension,
            TrustDimension.DOCUMENT: self._evaluate_document_dimension,
            TrustDimension.DEVICE: self._evaluate_device_dimension,
            TrustDimension.BIOMETRIC: self._evaluate_biometric_dimension,
            TrustDimension.TRANSACTION: self._evaluate_transaction_dimension,
            TrustDimension.REGIONAL: self._evaluate_regional_dimension,
        }
        
        # Carregar configurações padrão
        self._default_config = self._load_default_config()
        logger.info("TrustScoreEngine inicializado com configurações padrão")
    
    def _load_default_config(self) -> Dict[str, Any]:
        """
        Carrega configurações padrão para o motor de pontuação.
        
        Returns:
            Dict[str, Any]: Configurações padrão
        """
        return {
            "dimension_weights": {
                TrustDimension.IDENTITY: 0.3,
                TrustDimension.BEHAVIORAL: 0.2,
                TrustDimension.FINANCIAL: 0.25,
                TrustDimension.CONTEXTUAL: 0.15,
                TrustDimension.REPUTATION: 0.1,
                TrustDimension.DOCUMENT: 0.0,  # Estas dimensões são avaliadas dentro das principais
                TrustDimension.DEVICE: 0.0,
                TrustDimension.BIOMETRIC: 0.0,
                TrustDimension.TRANSACTION: 0.0,
                TrustDimension.REGIONAL: 0.0,
            },
            "score_thresholds": {
                TrustLevel.CRITICAL: 20,
                TrustLevel.LOW: 40,
                TrustLevel.MEDIUM: 60,
                TrustLevel.HIGH: 80,
                TrustLevel.VERY_HIGH: 90,
                TrustLevel.PREMIUM: 95
            },
            "anomaly_detection": {
                "enabled": True,
                "confidence_threshold": 0.7,
                "max_anomalies": 10
            },
            "factor_config": {
                "max_top_factors": 5,
                "negative_factor_weight": 1.5  # Peso extra para fatores negativos
            },
            "caching": {
                "enabled": True,
                "ttl_seconds": 300
            }
        }
    
    async def evaluate_trust_score(self, request: TrustScoreRequest) -> TrustScoreResponse:
        """
        Avalia a pontuação de confiança para um usuário específico.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            
        Returns:
            TrustScoreResponse: Resposta com pontuação de confiança e detalhes
        """
        start_time = time.time()
        
        # Criar contexto de trace
        with tracer.start_as_current_span("evaluate_trust_score") as span:
            span.set_attribute("user_id", request.user_id)
            span.set_attribute("tenant_id", request.tenant_id)
            span.set_attribute("event_type", request.event_type)
            span.set_attribute("region_code", request.region_code or "global")
            
            try:
                # Verificar cache se habilitado
                if self._default_config["caching"]["enabled"] and self.cache_provider:
                    cache_key = f"trust_score:{request.user_id}:{request.tenant_id}:{request.event_type}"
                    cached_response = await self.cache_provider.get(cache_key)
                    if cached_response:
                        logger.info(f"Pontuação de confiança recuperada do cache para usuário {request.user_id}")
                        span.set_attribute("cache_hit", True)
                        return TrustScoreResponse(**json.loads(cached_response))
                    span.set_attribute("cache_hit", False)
                
                # Carregar configurações específicas do tenant
                tenant_config = await self._load_tenant_config(request.tenant_id)
                
                # Carregar configuração regional se aplicável
                regional_config = None
                if request.region_code:
                    regional_config = await self._load_regional_config(request.region_code)
                    
                # Carregar perfil de confiança do usuário
                user_profile = await self._load_user_trust_profile(request.user_id, request.tenant_id)
                
                # Determinar dimensões a serem avaliadas
                dimensions_to_evaluate = request.dimensions or list(self._dimension_evaluators.keys())
                
                # Filtrar apenas dimensões principais que têm peso
                dimensions_to_evaluate = [d for d in dimensions_to_evaluate 
                                          if tenant_config["dimension_weights"].get(d, 0) > 0]
                
                # Avaliar cada dimensão
                dimension_results = {}
                dimension_factors = {}
                
                # Executar avaliações de dimensão em paralelo
                evaluation_tasks = []
                for dimension in dimensions_to_evaluate:
                    if dimension in self._dimension_evaluators:
                        evaluation_tasks.append(
                            self._evaluate_dimension(
                                dimension, request, user_profile, tenant_config, regional_config
                            )
                        )
                
                # Aguardar conclusão de todas as avaliações
                dimension_results_list = await asyncio.gather(*evaluation_tasks)
                
                # Processar resultados
                for dimension, score, factors in dimension_results_list:
                    dimension_results[dimension] = score
                    dimension_factors[dimension] = factors
                
                # Calcular pontuação geral
                overall_score, top_factors = self._calculate_overall_score(
                    dimension_results, dimension_factors, tenant_config, regional_config
                )
                
                # Determinar nível de confiança
                trust_level = self._determine_trust_level(overall_score, tenant_config)
                
                # Detectar anomalias se habilitado
                detected_anomalies = []
                if tenant_config["anomaly_detection"]["enabled"]:
                    detected_anomalies = await self._detect_anomalies(
                        request, user_profile, dimension_results, tenant_config, regional_config
                    )
                
                # Criar resposta
                response = TrustScoreResponse(
                    request_id=request.request_id,
                    user_id=request.user_id,
                    tenant_id=request.tenant_id,
                    transaction_id=request.transaction_id,
                    score=overall_score,
                    trust_level=trust_level,
                    dimension_scores=dimension_results,
                    regional_context=request.region_code,
                    top_factors=top_factors,
                    detected_anomalies=detected_anomalies,
                    confidence_level=self._calculate_confidence_level(dimension_results, detected_anomalies),
                    evaluation_time_ms=int((time.time() - start_time) * 1000),
                    timestamp=datetime.now()
                )
                
                # Salvar histórico
                await self._save_trust_score_history(request, response, dimension_factors)
                
                # Atualizar cache se habilitado
                if self._default_config["caching"]["enabled"] and self.cache_provider:
                    ttl = self._default_config["caching"]["ttl_seconds"]
                    await self.cache_provider.set(cache_key, response.json(), ttl)
                
                # Registrar métricas
                trust_score_counter.add(1, {"tenant_id": request.tenant_id, 
                                          "event_type": request.event_type,
                                          "region": request.region_code or "global"})
                
                trust_score_histogram.record(overall_score, {"trust_level": trust_level,
                                                           "tenant_id": request.tenant_id})
                
                processing_time_histogram.record(
                    response.evaluation_time_ms,
                    {"tenant_id": request.tenant_id, "event_type": request.event_type}
                )
                
                # Registrar no log
                logger.info(
                    f"Avaliação de confiança concluída para usuário {request.user_id}",
                    extra={
                        "user_id": request.user_id,
                        "tenant_id": request.tenant_id,
                        "score": overall_score,
                        "trust_level": trust_level,
                        "anomalies_count": len(detected_anomalies),
                        "processing_time_ms": response.evaluation_time_ms
                    }
                )
                
                return response
                
            except Exception as e:
                span.record_exception(e)
                logger.error(
                    f"Erro ao avaliar pontuação de confiança: {str(e)}",
                    exc_info=True,
                    extra={"user_id": request.user_id, "tenant_id": request.tenant_id}
                )
                raise HTTPException(
                    status_code=500,
                    detail=f"Erro ao processar avaliação de confiança: {str(e)}"
                )
    
    async def _evaluate_dimension(self, dimension: TrustDimension, 
                                 request: TrustScoreRequest,
                                 user_profile: UserTrustProfile,
                                 tenant_config: Dict[str, Any],
                                 regional_config: Optional[Dict[str, Any]]) -> Tuple[TrustDimension, int, List[TrustScoreFactorModel]]:
        """
        Avalia uma dimensão específica de confiança.
        
        Args:
            dimension: Dimensão a ser avaliada
            request: Solicitação original de avaliação
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[TrustDimension, int, List[TrustScoreFactorModel]]: 
                Dimensão, pontuação e fatores relevantes
        """
        with tracer.start_as_current_span(f"evaluate_dimension_{dimension}") as span:
            span.set_attribute("dimension", dimension)
            
            evaluator = self._dimension_evaluators.get(dimension)
            if not evaluator:
                logger.warning(f"Avaliador não encontrado para dimensão: {dimension}")
                return dimension, 50, []
            
            score, factors = await evaluator(request, user_profile, tenant_config, regional_config)
            
            span.set_attribute("dimension_score", score)
            span.set_attribute("factors_count", len(factors))
            
            return dimension, score, factors
    
    async def _evaluate_identity_dimension(self, request: TrustScoreRequest,
                                         user_profile: UserTrustProfile,
                                         tenant_config: Dict[str, Any],
                                         regional_config: Optional[Dict[str, Any]]) -> Tuple[int, List[TrustScoreFactorModel]]:
        """
        Avalia a dimensão de identidade do usuário.
        
        Considera verificações de identidade, biometria, documentos e outras credenciais.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[int, List[TrustScoreFactorModel]]: Pontuação e fatores relevantes
        """
        with tracer.start_as_current_span("evaluate_identity_dimension") as span:
            factors = []
            base_score = 50  # Pontuação inicial
            
            # Verificar documentos verificados
            if VerificationType.DOCUMENT in user_profile.completed_verifications:
                doc_factor = TrustScoreFactorModel(
                    factor_id=f"doc_verification_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.IDENTITY,
                    name="Documento Verificado",
                    description="Usuário possui documento de identidade verificado",
                    type=FactorType.POSITIVE,
                    weight=0.8,
                    value=1.0,
                    metadata={"verification_method": "manual_review"},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(doc_factor)
                base_score += 15
            
            # Verificar biometria
            if VerificationType.BIOMETRIC in user_profile.completed_verifications:
                biometric_factor = TrustScoreFactorModel(
                    factor_id=f"biometric_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.IDENTITY,
                    name="Biometria Verificada",
                    description="Usuário possui verificação biométrica concluída",
                    type=FactorType.POSITIVE,
                    weight=0.9,
                    value=1.0,
                    metadata={"biometric_type": "facial"},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(biometric_factor)
                base_score += 20
            
            # Verificar falhas recentes de verificação
            if VerificationType.DOCUMENT in user_profile.failed_verifications:
                failed_doc_factor = TrustScoreFactorModel(
                    factor_id=f"failed_doc_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.IDENTITY,
                    name="Falha em Verificação de Documento",
                    description="Usuário teve falha recente em verificação de documento",
                    type=FactorType.NEGATIVE,
                    weight=0.7,
                    value=1.0,
                    metadata={"failure_reason": "document_inconsistency"},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(failed_doc_factor)
                base_score -= 25
            
            # Verificar email confirmado
            if VerificationType.EMAIL in user_profile.completed_verifications:
                email_factor = TrustScoreFactorModel(
                    factor_id=f"verified_email_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.IDENTITY,
                    name="Email Verificado",
                    description="Usuário confirmou endereço de email",
                    type=FactorType.POSITIVE,
                    weight=0.5,
                    value=1.0,
                    metadata={},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(email_factor)
                base_score += 5
            
            # Verificar telefone confirmado
            if VerificationType.PHONE in user_profile.completed_verifications:
                phone_factor = TrustScoreFactorModel(
                    factor_id=f"verified_phone_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.IDENTITY,
                    name="Telefone Verificado",
                    description="Usuário confirmou número de telefone",
                    type=FactorType.POSITIVE,
                    weight=0.6,
                    value=1.0,
                    metadata={},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(phone_factor)
                base_score += 8
            
            # Aplicar ajustes regionais se disponíveis
            if regional_config:
                if "identity_score_adjustment" in regional_config:
                    base_score += regional_config["identity_score_adjustment"]
                    
                # Verificar requisitos específicos da região
                if "required_verifications" in regional_config:
                    for req_verification in regional_config["required_verifications"]:
                        if req_verification not in user_profile.completed_verifications:
                            missing_factor = TrustScoreFactorModel(
                                factor_id=f"missing_{req_verification}_{uuid.uuid4().hex[:8]}",
                                dimension=TrustDimension.IDENTITY,
                                name=f"Verificação {req_verification} Pendente",
                                description=f"Verificação obrigatória para região {request.region_code} pendente",
                                type=FactorType.NEGATIVE,
                                weight=0.7,
                                value=1.0,
                                metadata={"region_code": request.region_code},
                                regional_context=request.region_code,
                                created_at=datetime.now(),
                                updated_at=datetime.now()
                            )
                            factors.append(missing_factor)
                            base_score -= 15
            
            # Garantir que pontuação esteja no intervalo [0, 100]
            final_score = max(0, min(100, base_score))
            span.set_attribute("identity_score", final_score)
            
            return final_score, factors
    
    async def _evaluate_financial_dimension(self, request: TrustScoreRequest,
                                          user_profile: UserTrustProfile,
                                          tenant_config: Dict[str, Any],
                                          regional_config: Optional[Dict[str, Any]]) -> Tuple[int, List[TrustScoreFactorModel]]:
        """
        Avalia a dimensão financeira do usuário.
        
        Considera histórico de crédito, transações, padrões de pagamento e outros fatores financeiros.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[int, List[TrustScoreFactorModel]]: Pontuação e fatores relevantes
        """
        with tracer.start_as_current_span("evaluate_financial_dimension") as span:
            factors = []
            base_score = 50  # Pontuação inicial
            
            # Verificar se temos dados financeiros no contexto
            if not request.context_data or "financial_data" not in request.context_data:
                # Sem dados suficientes, retornar pontuação neutra
                logger.warning(f"Dados financeiros insuficientes para usuário {request.user_id}")
                span.set_attribute("financial_data_available", False)
                
                insufficient_factor = TrustScoreFactorModel(
                    factor_id=f"insufficient_financial_data_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.FINANCIAL,
                    name="Dados Financeiros Insuficientes",
                    description="Informações financeiras insuficientes para avaliação completa",
                    type=FactorType.NEUTRAL,
                    weight=1.0,
                    value=1.0,
                    metadata={},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(insufficient_factor)
                return base_score, factors
            
            span.set_attribute("financial_data_available", True)
            financial_data = request.context_data.get("financial_data", {})
            
            # Verificar score de crédito
            if "credit_score" in financial_data:
                credit_score = financial_data["credit_score"]
                span.set_attribute("credit_score", credit_score)
                
                # Mapear score de crédito para ajuste de pontuação
                if credit_score >= 800:
                    base_score += 30
                    factor_type = FactorType.POSITIVE
                    description = "Excelente histórico de crédito"
                elif credit_score >= 700:
                    base_score += 20
                    factor_type = FactorType.POSITIVE
                    description = "Bom histórico de crédito"
                elif credit_score >= 600:
                    base_score += 5
                    factor_type = FactorType.POSITIVE
                    description = "Histórico de crédito aceitável"
                elif credit_score >= 500:
                    base_score -= 10
                    factor_type = FactorType.NEGATIVE
                    description = "Histórico de crédito abaixo da média"
                else:
                    base_score -= 20
                    factor_type = FactorType.NEGATIVE
                    description = "Histórico de crédito ruim"
                
                credit_factor = TrustScoreFactorModel(
                    factor_id=f"credit_score_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.FINANCIAL,
                    name="Score de Crédito",
                    description=description,
                    type=factor_type,
                    weight=0.8,
                    value=credit_score / 1000.0,  # Normalizar para [0,1]
                    metadata={"raw_score": credit_score},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(credit_factor)
            
            # Verificar histórico de pagamentos
            if "payment_history" in financial_data:
                payment_history = financial_data["payment_history"]
                late_payments = payment_history.get("late_payments", 0)
                on_time_payments = payment_history.get("on_time_payments", 0)
                
                payment_ratio = 1.0
                if (late_payments + on_time_payments) > 0:
                    payment_ratio = on_time_payments / (late_payments + on_time_payments)
                
                span.set_attribute("payment_ratio", payment_ratio)
                
                if payment_ratio > 0.95:
                    base_score += 15
                    factor_type = FactorType.POSITIVE
                    description = "Excelente histórico de pagamentos em dia"
                elif payment_ratio > 0.85:
                    base_score += 10
                    factor_type = FactorType.POSITIVE
                    description = "Bom histórico de pagamentos"
                elif payment_ratio > 0.7:
                    base_score += 0
                    factor_type = FactorType.NEUTRAL
                    description = "Histórico de pagamentos aceitável"
                else:
                    base_score -= 15
                    factor_type = FactorType.NEGATIVE
                    description = "Histórico de pagamentos com atrasos frequentes"
                
                payment_factor = TrustScoreFactorModel(
                    factor_id=f"payment_history_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.FINANCIAL,
                    name="Histórico de Pagamentos",
                    description=description,
                    type=factor_type,
                    weight=0.7,
                    value=payment_ratio,
                    metadata={
                        "late_payments": late_payments,
                        "on_time_payments": on_time_payments
                    },
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(payment_factor)
            
            # Verificar relação dívida/renda
            if "debt_to_income" in financial_data:
                dti = financial_data["debt_to_income"]
                span.set_attribute("debt_to_income", dti)
                
                if dti < 0.3:
                    base_score += 10
                    factor_type = FactorType.POSITIVE
                    description = "Excelente relação dívida/renda"
                elif dti < 0.4:
                    base_score += 5
                    factor_type = FactorType.POSITIVE
                    description = "Boa relação dívida/renda"
                elif dti < 0.5:
                    base_score += 0
                    factor_type = FactorType.NEUTRAL
                    description = "Relação dívida/renda aceitável"
                else:
                    base_score -= 10
                    factor_type = FactorType.NEGATIVE
                    description = "Alta relação dívida/renda"
                
                dti_factor = TrustScoreFactorModel(
                    factor_id=f"debt_to_income_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.FINANCIAL,
                    name="Relação Dívida/Renda",
                    description=description,
                    type=factor_type,
                    weight=0.6,
                    value=1.0 - min(dti, 1.0),  # Inverter para normalizar
                    metadata={"ratio": dti},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(dti_factor)
            
            # Aplicar ajustes regionais se disponíveis
            if regional_config:
                if "financial_score_adjustment" in regional_config:
                    base_score += regional_config["financial_score_adjustment"]
            
            # Garantir que pontuação esteja no intervalo [0, 100]
            final_score = max(0, min(100, base_score))
            span.set_attribute("financial_score", final_score)
            
            return final_score, factors
    
    async def _evaluate_behavioral_dimension(self, request: TrustScoreRequest,
                                          user_profile: UserTrustProfile,
                                          tenant_config: Dict[str, Any],
                                          regional_config: Optional[Dict[str, Any]]) -> Tuple[int, List[TrustScoreFactorModel]]:
        """
        Avalia a dimensão comportamental do usuário.
        
        Considera padrões de uso, consistência de comportamento, e anomalias comportamentais.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[int, List[TrustScoreFactorModel]]: Pontuação e fatores relevantes
        """
        with tracer.start_as_current_span("evaluate_behavioral_dimension") as span:
            factors = []
            base_score = 50  # Pontuação inicial
            
            # Verificar presença de dispositivo conhecido
            device_data = request.device_data or {}
            if device_data:
                if "device_id" in device_data and "known_devices" in user_profile.history_summary:
                    device_id = device_data["device_id"]
                    known_devices = user_profile.history_summary.get("known_devices", [])
                    
                    if device_id in known_devices:
                        base_score += 10
                        device_factor = TrustScoreFactorModel(
                            factor_id=f"known_device_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.BEHAVIORAL,
                            name="Dispositivo Conhecido",
                            description="Acesso de dispositivo previamente utilizado",
                            type=FactorType.POSITIVE,
                            weight=0.7,
                            value=1.0,
                            metadata={"device_id": device_id},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(device_factor)
                    else:
                        base_score -= 5
                        new_device_factor = TrustScoreFactorModel(
                            factor_id=f"new_device_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.BEHAVIORAL,
                            name="Novo Dispositivo",
                            description="Acesso de dispositivo não reconhecido",
                            type=FactorType.NEUTRAL,
                            weight=0.6,
                            value=1.0,
                            metadata={"device_id": device_id},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(new_device_factor)
            
            # Verificar anomalias comportamentais recentes
            for anomaly in user_profile.last_anomalies:
                if anomaly.type == AnomalyType.UNUSUAL_BEHAVIOR:
                    base_score -= 15
                    anomaly_factor = TrustScoreFactorModel(
                        factor_id=f"behavioral_anomaly_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.BEHAVIORAL,
                        name="Anomalia Comportamental Recente",
                        description=f"Detecção recente de {anomaly.description}",
                        type=FactorType.NEGATIVE,
                        weight=0.8,
                        value=1.0,
                        metadata={"anomaly_id": anomaly.anomaly_id},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(anomaly_factor)
            
            # Verificar consistência de padrões de uso
            if "usage_patterns" in user_profile.history_summary:
                usage_patterns = user_profile.history_summary["usage_patterns"]
                
                if usage_patterns.get("consistency_score", 0) > 0.8:
                    base_score += 15
                    consistency_factor = TrustScoreFactorModel(
                        factor_id=f"usage_consistency_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.BEHAVIORAL,
                        name="Padrão de Uso Consistente",
                        description="Usuário demonstra padrões de uso consistentes",
                        type=FactorType.POSITIVE,
                        weight=0.75,
                        value=usage_patterns.get("consistency_score", 0),
                        metadata={},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(consistency_factor)
            
            # Verificar tempo de atividade da conta
            if "account_activity_days" in user_profile.history_summary:
                activity_days = user_profile.history_summary["account_activity_days"]
                
                if activity_days > 365:
                    base_score += 15
                    activity_factor = TrustScoreFactorModel(
                        factor_id=f"long_term_activity_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.BEHAVIORAL,
                        name="Atividade de Longo Prazo",
                        description="Conta ativa por mais de 1 ano",
                        type=FactorType.POSITIVE,
                        weight=0.8,
                        value=1.0,
                        metadata={"days": activity_days},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(activity_factor)
                elif activity_days < 30:
                    base_score -= 10
                    new_account_factor = TrustScoreFactorModel(
                        factor_id=f"new_account_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.BEHAVIORAL,
                        name="Conta Nova",
                        description="Conta ativa por menos de 30 dias",
                        type=FactorType.NEUTRAL,
                        weight=0.7,
                        value=1.0,
                        metadata={"days": activity_days},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(new_account_factor)
            
            # Aplicar ajustes regionais se disponíveis
            if regional_config:
                if "behavioral_score_adjustment" in regional_config:
                    base_score += regional_config["behavioral_score_adjustment"]
            
            # Garantir que pontuação esteja no intervalo [0, 100]
            final_score = max(0, min(100, base_score))
            span.set_attribute("behavioral_score", final_score)
            
            return final_score, factors
            
    async def _detect_anomalies(self, request: TrustScoreRequest,
                              user_profile: UserTrustProfile,
                              dimension_scores: Dict[TrustDimension, int],
                              tenant_config: Dict[str, Any],
                              regional_config: Optional[Dict[str, Any]]) -> List[DetectedAnomaly]:
        """
        Detecta anomalias nos padrões de uso e pontuações.
        
        Identifica comportamentos suspeitos baseados em histórico e variações súbitas de padrões.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            dimension_scores: Pontuações por dimensão
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            List[DetectedAnomaly]: Lista de anomalias detectadas
        """
        with tracer.start_as_current_span("detect_anomalies") as span:
            anomalies = []
            
            # Verificar histórico de pontuações anteriores
            if user_profile.trust_score_history:
                # Obter pontuações recentes
                recent_scores = [item.score for item in user_profile.trust_score_history[-5:]]
                current_score = sum(dimension_scores.values()) / len(dimension_scores)
                
                # Verificar se há queda acentuada na pontuação
                if recent_scores and recent_scores[-1] > 0:
                    avg_recent_score = sum(recent_scores) / len(recent_scores)
                    score_drop = avg_recent_score - current_score
                    
                    if score_drop > 20:  # Queda significativa
                        anomaly = DetectedAnomaly(
                            anomaly_id=f"score_drop_{uuid.uuid4().hex[:8]}",
                            type=AnomalyType.SCORE_VARIATION,
                            description="Queda acentuada na pontuação de confiança",
                            severity=AnomalySeverity.HIGH,
                            confidence=0.85,
                            affected_dimensions=[d for d, s in dimension_scores.items() 
                                               if s < 50],  # Dimensões com pontuação baixa
                            metadata={
                                "avg_recent_score": avg_recent_score,
                                "current_score": current_score,
                                "score_drop": score_drop
                            },
                            detected_at=datetime.now()
                        )
                        anomalies.append(anomaly)
                        anomaly_detection_counter.add(1, {"anomaly_type": "score_drop",
                                                         "tenant_id": request.tenant_id})
            
            # Verificar anomalias de localização
            if user_profile.history_summary.get("usual_locations") and request.location_data:
                usual_locations = user_profile.history_summary["usual_locations"]
                current_location = request.location_data.get("country") or request.location_data.get("city")
                
                # Se a localização atual não estiver entre as usuais
                if current_location and current_location not in usual_locations:
                    # Verificar velocidade impossível (ex: acesso em dois países em curto período)
                    if user_profile.last_activity and user_profile.last_activity.location_data:
                        last_location = (user_profile.last_activity.location_data.get("country") or 
                                        user_profile.last_activity.location_data.get("city"))
                        last_time = user_profile.last_activity.timestamp
                        time_diff = datetime.now() - last_time
                        
                        # Se último acesso foi há menos de 2 horas e de localização diferente
                        if (last_location and last_location != current_location and 
                            time_diff < timedelta(hours=2)):
                            
                            anomaly = DetectedAnomaly(
                                anomaly_id=f"impossible_travel_{uuid.uuid4().hex[:8]}",
                                type=AnomalyType.LOCATION_ANOMALY,
                                description="Velocidade de deslocamento impossível entre acessos",
                                severity=AnomalySeverity.CRITICAL,
                                confidence=0.95,
                                affected_dimensions=[TrustDimension.CONTEXTUAL],
                                metadata={
                                    "current_location": current_location,
                                    "last_location": last_location,
                                    "time_diff_minutes": time_diff.total_seconds() / 60
                                },
                                detected_at=datetime.now()
                            )
                            anomalies.append(anomaly)
                            anomaly_detection_counter.add(1, {"anomaly_type": "impossible_travel",
                                                            "tenant_id": request.tenant_id})
            
            # Verificar anomalias de dispositivo
            device_data = request.device_data or {}
            if device_data.get("device_id") and user_profile.history_summary.get("device_attributes"):
                device_id = device_data["device_id"]
                known_attributes = user_profile.history_summary.get("device_attributes", {})
                
                # Se dispositivo conhecido mas atributos diferentes
                if device_id in known_attributes:
                    expected_attrs = known_attributes[device_id]
                    differences = []
                    
                    # Comparar atributos do dispositivo
                    for key in ["os", "browser", "screen_resolution", "timezone"]:
                        if (key in expected_attrs and key in device_data and 
                            expected_attrs[key] != device_data[key]):
                            differences.append(key)
                    
                    if differences:
                        anomaly = DetectedAnomaly(
                            anomaly_id=f"device_attribute_change_{uuid.uuid4().hex[:8]}",
                            type=AnomalyType.DEVICE_ANOMALY,
                            description="Alterações em atributos de dispositivo conhecido",
                            severity=AnomalySeverity.MEDIUM,
                            confidence=0.8,
                            affected_dimensions=[TrustDimension.BEHAVIORAL, TrustDimension.DEVICE],
                            metadata={
                                "device_id": device_id,
                                "changed_attributes": differences
                            },
                            detected_at=datetime.now()
                        )
                        anomalies.append(anomaly)
                        anomaly_detection_counter.add(1, {"anomaly_type": "device_attribute_change",
                                                        "tenant_id": request.tenant_id})
            
            # Verificar anomalias financeiras/transacionais
            if request.context_data and "financial_data" in request.context_data:
                financial_data = request.context_data.get("financial_data", {})
                
                # Verificar transação com valor anormal
                if "transaction_amount" in financial_data and "average_transaction" in user_profile.history_summary:
                    tx_amount = financial_data["transaction_amount"]
                    avg_amount = user_profile.history_summary["average_transaction"]
                    
                    if tx_amount > avg_amount * 5:  # Transação 5x acima da média
                        anomaly = DetectedAnomaly(
                            anomaly_id=f"unusual_transaction_amount_{uuid.uuid4().hex[:8]}",
                            type=AnomalyType.FINANCIAL_ANOMALY,
                            description="Transação com valor muito acima da média do usuário",
                            severity=AnomalySeverity.HIGH,
                            confidence=0.85,
                            affected_dimensions=[TrustDimension.FINANCIAL, TrustDimension.TRANSACTION],
                            metadata={
                                "transaction_amount": tx_amount,
                                "average_amount": avg_amount,
                                "multiple": tx_amount / avg_amount
                            },
                            detected_at=datetime.now()
                        )
                        anomalies.append(anomaly)
                        anomaly_detection_counter.add(1, {"anomaly_type": "unusual_amount",
                                                        "tenant_id": request.tenant_id})
            
            # Limitar número máximo de anomalias
            max_anomalies = tenant_config["anomaly_detection"]["max_anomalies"]
            if len(anomalies) > max_anomalies:
                # Ordenar por severidade e confiança
                anomalies.sort(key=lambda a: (a.severity, a.confidence), reverse=True)
                anomalies = anomalies[:max_anomalies]
            
            # Registrar métricas
            span.set_attribute("anomalies_detected", len(anomalies))
            for anomaly in anomalies:
                span.set_attribute(f"anomaly_{anomaly.type}", 1)
            
            return anomalies
            
    def _calculate_confidence_level(self, dimension_scores: Dict[TrustDimension, int], 
                                  detected_anomalies: List[DetectedAnomaly]) -> float:
        """
        Calcula o nível de confiança na pontuação de confiança geral.
        
        Baseado na cobertura das dimensões e anomalias detectadas.
        
        Args:
            dimension_scores: Pontuações por dimensão
            detected_anomalies: Anomalias detectadas
            
        Returns:
            float: Nível de confiança entre 0.0 e 1.0
        """
        with tracer.start_as_current_span("calculate_confidence_level") as span:
            # Dimensões principais que devem estar presentes para alta confiança
            key_dimensions = [
                TrustDimension.IDENTITY, 
                TrustDimension.BEHAVIORAL, 
                TrustDimension.FINANCIAL
            ]
            
            # Verificar cobertura de dimensões principais
            dimension_coverage = sum(1 for dim in key_dimensions if dim in dimension_scores) / len(key_dimensions)
            
            # Base de confiança baseada na cobertura de dimensões
            base_confidence = 0.5 + (dimension_coverage * 0.5)  # 0.5 a 1.0
            
            # Reduzir confiança com base em anomalias detectadas
            anomaly_penalty = 0.0
            if detected_anomalies:
                # Penalidades por severidade de anomalia
                severity_weights = {
                    AnomalySeverity.LOW: 0.05,
                    AnomalySeverity.MEDIUM: 0.1,
                    AnomalySeverity.HIGH: 0.15,
                    AnomalySeverity.CRITICAL: 0.25
                }
                
                # Somar penalidades
                for anomaly in detected_anomalies:
                    anomaly_penalty += severity_weights.get(anomaly.severity, 0.1) * anomaly.confidence
                
                # Limitar penalidade máxima
                anomaly_penalty = min(anomaly_penalty, 0.8)
            
            # Calcular confiança final
            confidence_level = max(0.1, base_confidence - anomaly_penalty)
            
            span.set_attribute("dimension_coverage", dimension_coverage)
            span.set_attribute("anomaly_penalty", anomaly_penalty)
            span.set_attribute("confidence_level", confidence_level)
            
            return confidence_level
            
    async def _evaluate_reputation_dimension(self, request: TrustScoreRequest,
                                          user_profile: UserTrustProfile,
                                          tenant_config: Dict[str, Any],
                                          regional_config: Optional[Dict[str, Any]]) -> Tuple[int, List[TrustScoreFactorModel]]:
        """
        Avalia a dimensão de reputação do usuário.
        
        Considera feedback, histórico de denúncias, reputação em outros sistemas, etc.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[int, List[TrustScoreFactorModel]]: Pontuação e fatores relevantes
        """
        with tracer.start_as_current_span("evaluate_reputation_dimension") as span:
            factors = []
            base_score = 50  # Pontuação inicial
            
            # Verificar presença de dados de reputação
            if not request.context_data or not "reputation_data" in request.context_data:
                # Sem dados de reputação, retornar pontuação neutra
                neutral_factor = TrustScoreFactorModel(
                    factor_id=f"insufficient_reputation_data_{uuid.uuid4().hex[:8]}",
                    dimension=TrustDimension.REPUTATION,
                    name="Dados de Reputação Insuficientes",
                    description="Informações de reputação insuficientes para avaliação completa",
                    type=FactorType.NEUTRAL,
                    weight=1.0,
                    value=1.0,
                    metadata={},
                    regional_context=request.region_code,
                    created_at=datetime.now(),
                    updated_at=datetime.now()
                )
                factors.append(neutral_factor)
                return base_score, factors
            
            reputation_data = request.context_data.get("reputation_data", {})
            
            # Verificar feedback positivo
            if "positive_feedback" in reputation_data and "total_feedback" in reputation_data:
                positive = reputation_data["positive_feedback"]
                total = reputation_data["total_feedback"]
                
                if total > 0:
                    positive_ratio = positive / total
                    
                    if positive_ratio > 0.9 and total >= 10:
                        base_score += 20
                        feedback_factor = TrustScoreFactorModel(
                            factor_id=f"excellent_feedback_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.REPUTATION,
                            name="Excelente Feedback",
                            description="Usuário possui excelente histórico de feedback positivo",
                            type=FactorType.POSITIVE,
                            weight=0.8,
                            value=positive_ratio,
                            metadata={"positive": positive, "total": total},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(feedback_factor)
                    elif positive_ratio > 0.7:
                        base_score += 10
                        feedback_factor = TrustScoreFactorModel(
                            factor_id=f"good_feedback_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.REPUTATION,
                            name="Bom Feedback",
                            description="Usuário possui bom histórico de feedback positivo",
                            type=FactorType.POSITIVE,
                            weight=0.7,
                            value=positive_ratio,
                            metadata={"positive": positive, "total": total},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(feedback_factor)
                    elif positive_ratio < 0.5 and total >= 5:
                        base_score -= 15
                        negative_feedback_factor = TrustScoreFactorModel(
                            factor_id=f"negative_feedback_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.REPUTATION,
                            name="Feedback Negativo",
                            description="Usuário possui histórico significativo de feedback negativo",
                            type=FactorType.NEGATIVE,
                            weight=0.75,
                            value=1.0 - positive_ratio,
                            metadata={"positive": positive, "total": total},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(negative_feedback_factor)
            
            # Verificar denúncias
            if "report_count" in reputation_data:
                report_count = reputation_data["report_count"]
                
                if report_count > 0:
                    base_score -= 10 * min(report_count, 5)  # Limite de -50 pontos
                    report_factor = TrustScoreFactorModel(
                        factor_id=f"user_reports_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.REPUTATION,
                        name="Denúncias",
                        description=f"Usuário possui {report_count} denúncias",
                        type=FactorType.NEGATIVE,
                        weight=0.8,
                        value=min(report_count / 10.0, 1.0),
                        metadata={"count": report_count},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(report_factor)
            
            # Verificar badges/conquistas
            if "badges" in reputation_data:
                badges = reputation_data["badges"]
                
                if badges and len(badges) > 0:
                    badges_count = len(badges)
                    base_score += min(badges_count * 5, 20)  # Limite de +20 pontos
                    
                    badges_factor = TrustScoreFactorModel(
                        factor_id=f"user_badges_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.REPUTATION,
                        name="Badges e Conquistas",
                        description=f"Usuário possui {badges_count} badges/conquistas",
                        type=FactorType.POSITIVE,
                        weight=0.6,
                        value=min(badges_count / 10.0, 1.0),
                        metadata={"badges": badges},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(badges_factor)
            
            # Verificar reputação externa
            if "external_reputation_score" in reputation_data:
                external_score = reputation_data["external_reputation_score"]
                
                if external_score > 80:
                    base_score += 15
                    external_factor = TrustScoreFactorModel(
                        factor_id=f"external_reputation_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.REPUTATION,
                        name="Excelente Reputação Externa",
                        description="Usuário possui excelente reputação em sistemas externos",
                        type=FactorType.POSITIVE,
                        weight=0.7,
                        value=external_score / 100.0,
                        metadata={"score": external_score},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(external_factor)
                elif external_score < 40:
                    base_score -= 15
                    external_factor = TrustScoreFactorModel(
                        factor_id=f"poor_external_reputation_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.REPUTATION,
                        name="Baixa Reputação Externa",
                        description="Usuário possui baixa reputação em sistemas externos",
                        type=FactorType.NEGATIVE,
                        weight=0.7,
                        value=(100.0 - external_score) / 100.0,
                        metadata={"score": external_score},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(external_factor)
            
            # Aplicar ajustes regionais se disponíveis
            if regional_config:
                if "reputation_score_adjustment" in regional_config:
                    base_score += regional_config["reputation_score_adjustment"]
            
            # Garantir que pontuação esteja no intervalo [0, 100]
            final_score = max(0, min(100, base_score))
            span.set_attribute("reputation_score", final_score)
            
            return final_score, factors
            
    async def _evaluate_contextual_dimension(self, request: TrustScoreRequest,
                                           user_profile: UserTrustProfile,
                                           tenant_config: Dict[str, Any],
                                           regional_config: Optional[Dict[str, Any]]) -> Tuple[int, List[TrustScoreFactorModel]]:
        """
        Avalia a dimensão contextual do usuário.
        
        Considera informações de localização, horário de acesso, dispositivo e contexto da transação.
        
        Args:
            request: Solicitação de avaliação de pontuação de confiança
            user_profile: Perfil de confiança do usuário
            tenant_config: Configuração específica do tenant
            regional_config: Configuração regional (opcional)
            
        Returns:
            Tuple[int, List[TrustScoreFactorModel]]: Pontuação e fatores relevantes
        """
        with tracer.start_as_current_span("evaluate_contextual_dimension") as span:
            factors = []
            base_score = 50  # Pontuação inicial
            
            # Verificar dados de localização
            location_data = request.location_data or {}
            if location_data:
                # Verificar se a localização é consistente com o histórico
                if "usual_locations" in user_profile.history_summary:
                    usual_locations = user_profile.history_summary.get("usual_locations", [])
                    current_location = location_data.get("country") or location_data.get("city")
                    
                    if current_location and usual_locations and current_location in usual_locations:
                        base_score += 15
                        location_factor = TrustScoreFactorModel(
                            factor_id=f"familiar_location_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.CONTEXTUAL,
                            name="Localização Familiar",
                            description="Acesso de localização previamente utilizada",
                            type=FactorType.POSITIVE,
                            weight=0.7,
                            value=1.0,
                            metadata={"location": current_location},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(location_factor)
                    elif current_location:
                        # Localização não reconhecida
                        base_score -= 10
                        new_location_factor = TrustScoreFactorModel(
                            factor_id=f"new_location_{uuid.uuid4().hex[:8]}",
                            dimension=TrustDimension.CONTEXTUAL,
                            name="Nova Localização",
                            description="Acesso de localização não reconhecida",
                            type=FactorType.NEUTRAL,
                            weight=0.6,
                            value=1.0,
                            metadata={"location": current_location},
                            regional_context=request.region_code,
                            created_at=datetime.now(),
                            updated_at=datetime.now()
                        )
                        factors.append(new_location_factor)
            
            # Verificar horário de acesso
            current_hour = datetime.now().hour
            if "usual_access_hours" in user_profile.history_summary:
                usual_hours = user_profile.history_summary.get("usual_access_hours", [])
                
                if usual_hours and current_hour in usual_hours:
                    base_score += 10
                    time_factor = TrustScoreFactorModel(
                        factor_id=f"usual_time_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.CONTEXTUAL,
                        name="Horário Usual",
                        description="Acesso em horário habitual",
                        type=FactorType.POSITIVE,
                        weight=0.5,
                        value=1.0,
                        metadata={"hour": current_hour},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(time_factor)
                elif current_hour < 6 or current_hour > 22:  # Horários incomuns
                    base_score -= 5
                    unusual_time_factor = TrustScoreFactorModel(
                        factor_id=f"unusual_time_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.CONTEXTUAL,
                        name="Horário Incomum",
                        description="Acesso em horário não usual",
                        type=FactorType.NEUTRAL,
                        weight=0.4,
                        value=1.0,
                        metadata={"hour": current_hour},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(unusual_time_factor)
            
            # Verificar tipo de evento/transação
            if "high_risk_events" in tenant_config:
                high_risk_events = tenant_config.get("high_risk_events", [])
                
                if request.event_type in high_risk_events:
                    base_score -= 10
                    high_risk_factor = TrustScoreFactorModel(
                        factor_id=f"high_risk_event_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.CONTEXTUAL,
                        name="Evento de Alto Risco",
                        description=f"O evento {request.event_type} é considerado de alto risco",
                        type=FactorType.NEGATIVE,
                        weight=0.7,
                        value=1.0,
                        metadata={"event_type": request.event_type},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(high_risk_factor)
            
            # Verificar anomalias de IP/rede
            device_data = request.device_data or {}
            if device_data and "ip_address" in device_data:
                ip_address = device_data["ip_address"]
                
                # Verificar se o IP está em uma lista de IPs suspeitos
                if "suspicious_ips" in tenant_config and ip_address in tenant_config.get("suspicious_ips", []):
                    base_score -= 20
                    ip_factor = TrustScoreFactorModel(
                        factor_id=f"suspicious_ip_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.CONTEXTUAL,
                        name="IP Suspeito",
                        description="Acesso de endereço IP considerado suspeito",
                        type=FactorType.NEGATIVE,
                        weight=0.8,
                        value=1.0,
                        metadata={"ip_address": ip_address},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(ip_factor)
                    
                # Verificar VPN/proxy
                if "is_vpn" in device_data and device_data["is_vpn"]:
                    base_score -= 5
                    vpn_factor = TrustScoreFactorModel(
                        factor_id=f"vpn_usage_{uuid.uuid4().hex[:8]}",
                        dimension=TrustDimension.CONTEXTUAL,
                        name="Uso de VPN/Proxy",
                        description="Acesso através de VPN ou proxy detectado",
                        type=FactorType.NEUTRAL,
                        weight=0.6,
                        value=1.0,
                        metadata={},
                        regional_context=request.region_code,
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    factors.append(vpn_factor)
            
            # Aplicar ajustes regionais se disponíveis
            if regional_config:
                if "contextual_score_adjustment" in regional_config:
                    base_score += regional_config["contextual_score_adjustment"]
            
            # Garantir que pontuação esteja no intervalo [0, 100]
            final_score = max(0, min(100, base_score))
            span.set_attribute("contextual_score", final_score)
            
            return final_score, factors