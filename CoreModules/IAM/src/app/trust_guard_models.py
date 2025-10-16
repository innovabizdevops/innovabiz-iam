"""
Modelos de dados para o módulo TrustGuard.

Este módulo define as estruturas de dados e enumerações utilizadas
pelo motor de pontuação de confiança (TrustScore Engine) e seus componentes.
Implementa estruturas com suporte a multi-tenant e multi-contexto.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import uuid
from enum import Enum
from typing import Dict, List, Optional, Any, Union
from datetime import datetime
from pydantic import BaseModel, Field


class TrustDimension(str, Enum):
    """Dimensões de avaliação de confiança."""
    IDENTITY = "identity"
    BEHAVIORAL = "behavioral"
    FINANCIAL = "financial"
    CONTEXTUAL = "contextual"
    REPUTATION = "reputation"


class AnomalyType(str, Enum):
    """Tipos de anomalias que podem ser detectadas."""
    SCORE_DROP = "score_drop"
    IMPOSSIBLE_TRAVEL = "impossible_travel"
    UNUSUAL_LOCATION = "unusual_location"
    DEVICE_CHANGE = "device_change"
    UNUSUAL_BEHAVIOR = "unusual_behavior"
    FINANCIAL_ANOMALY = "financial_anomaly"
    IDENTITY_MISMATCH = "identity_mismatch"
    UNUSUAL_TRANSACTION = "unusual_transaction"
    CREDENTIAL_STUFFING = "credential_stuffing"
    ACCOUNT_TAKEOVER_ATTEMPT = "account_takeover_attempt"


class AnomalySeverity(str, Enum):
    """Níveis de severidade para anomalias detectadas."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class FactorType(str, Enum):
    """Tipos de fatores que influenciam pontuações de confiança."""
    POSITIVE = "positive"
    NEGATIVE = "negative"
    NEUTRAL = "neutral"
    REGIONAL = "regional"
    TEMPORAL = "temporal"


class DetectedAnomaly(BaseModel):
    """
    Modelo de anomalia detectada durante avaliação de confiança.
    """
    anomaly_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    type: AnomalyType
    description: str
    severity: AnomalySeverity
    confidence: float  # 0.0 a 1.0
    affected_dimensions: List[TrustDimension] = []
    metadata: Optional[Dict[str, Any]] = {}
    detected_at: datetime = Field(default_factory=datetime.now)


class TrustScoreFactorModel(BaseModel):
    """
    Modelo de fator que influencia uma pontuação de confiança.
    """
    factor_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    dimension: TrustDimension
    name: str
    description: str
    type: FactorType
    weight: float  # 0.0 a 1.0
    value: float  # -1.0 a 1.0 (negativo a positivo)
    regional_context: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = {}
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class TrustScoreResult(BaseModel):
    """
    Resultado de uma avaliação de pontuação de confiança.
    """
    user_id: str
    tenant_id: str
    context_id: Optional[str] = None
    overall_score: float  # 0.0 a 1.0
    dimension_scores: Dict[TrustDimension, float] = {}
    regional_context: Optional[str] = None
    confidence_level: float  # 0.0 a 1.0 (baixa a alta confiança na pontuação)
    evaluation_time_ms: int = 0
    timestamp: datetime = Field(default_factory=datetime.now)
    metadata: Dict[str, Any] = {}


class TrustScoreHistoryItem(BaseModel):
    """
    Item de histórico de pontuação de confiança para armazenamento compacto.
    """
    score: float
    dimension_scores: Dict[TrustDimension, float] = {}
    confidence_level: float
    region_code: Optional[str] = None
    context_id: Optional[str] = None
    timestamp: datetime
    anomaly_count: int = 0


class UserTrustProfile(BaseModel):
    """
    Perfil de confiança de um usuário com histórico e resumo de fatores.
    """
    user_id: str
    tenant_id: str
    latest_score: float
    trust_score_history: List[TrustScoreHistoryItem] = []
    history_summary: Dict[str, Any] = {}
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class TrustScoreRequest(BaseModel):
    """
    Solicitação de avaliação de pontuação de confiança.
    """
    user_id: str
    tenant_id: str
    context_id: Optional[str] = None
    session_id: Optional[str] = None
    regional_context: Optional[str] = None
    transaction_data: Optional[Dict[str, Any]] = None
    device_data: Optional[Dict[str, Any]] = None
    location_data: Optional[Dict[str, Any]] = None
    requested_dimensions: Optional[List[TrustDimension]] = None
    skip_factors: Optional[List[str]] = None