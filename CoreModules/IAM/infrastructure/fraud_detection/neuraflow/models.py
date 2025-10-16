"""
Modelos de dados para integração com NeuraFlow

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from enum import Enum
from typing import Dict, List, Optional, Any, Union
from pydantic import BaseModel, Field


class ModelType(str, Enum):
    """Tipos de modelos disponíveis no NeuraFlow"""
    ANOMALY_DETECTION = "anomaly_detection"
    FRAUD_PREDICTION = "fraud_prediction"
    BEHAVIOR_ANALYSIS = "behavior_analysis"
    RISK_SCORING = "risk_scoring"
    ENTITY_CLASSIFICATION = "entity_classification"
    PATTERN_RECOGNITION = "pattern_recognition"
    ADAPTIVE_AUTHENTICATION = "adaptive_authentication"


class ModelMetadata(BaseModel):
    """Metadados do modelo NeuraFlow"""
    model_id: str
    model_type: ModelType
    version: str
    region: Optional[str] = None
    description: Optional[str] = None
    created_at: str
    updated_at: str
    accuracy: float
    confidence_threshold: float
    tags: List[str] = Field(default_factory=list)


class NeuraFlowDetectionRequest(BaseModel):
    """Solicitação de detecção para o NeuraFlow"""
    event_data: Dict[str, Any]
    model_id: Optional[str] = None
    model_type: Optional[ModelType] = None
    region: Optional[str] = None
    context: Optional[Dict[str, Any]] = None
    request_id: Optional[str] = None
    tenant_id: str
    confidence_threshold: Optional[float] = None
    feature_flags: Dict[str, bool] = Field(default_factory=dict)


class ModelResult(BaseModel):
    """Resultado da avaliação de um modelo individual"""
    model_id: str
    model_type: ModelType
    score: float
    confidence: float
    decision: str
    features: Dict[str, float] = Field(default_factory=dict)
    explanation: Optional[Dict[str, Any]] = None
    processing_time_ms: float


class NeuraFlowDetectionResponse(BaseModel):
    """Resposta de detecção do NeuraFlow"""
    request_id: str
    timestamp: str
    results: List[ModelResult]
    enhanced_data: Dict[str, Any]
    aggregated_score: float
    decision: str
    processing_time_ms: float
    tenant_id: str
    region: Optional[str] = None
    trace_id: str


class EnhancedFeature(BaseModel):
    """Feature aprimorada com dados de IA/ML"""
    name: str
    value: Any
    confidence: float
    source: str
    importance: float = 0.0
    related_features: List[str] = Field(default_factory=list)


class EnhancedEventData(BaseModel):
    """Dados de evento aprimorados pelo NeuraFlow"""
    original_event: Dict[str, Any]
    enhanced_features: Dict[str, EnhancedFeature]
    risk_indicators: Dict[str, float]
    context_enrichment: Dict[str, Any]
    behavioral_patterns: Dict[str, Any]
    regional_factors: Dict[str, Any]
    temporal_analysis: Dict[str, Any]


class EnhancementType(str, Enum):
    """Tipos de aprimoramentos disponíveis"""
    FEATURE_EXTRACTION = "feature_extraction"
    CONTEXT_ENRICHMENT = "context_enrichment"
    RISK_SCORING = "risk_scoring"
    BEHAVIORAL_ANALYSIS = "behavioral_analysis"
    ANOMALY_DETECTION = "anomaly_detection"
    PATTERN_RECOGNITION = "pattern_recognition"


class EnhancementConfig(BaseModel):
    """Configuração para aprimoramento de dados"""
    enhancement_types: List[EnhancementType]
    models: List[str] = Field(default_factory=list)
    confidence_threshold: float = 0.7
    max_processing_time_ms: int = 200
    feature_flags: Dict[str, bool] = Field(default_factory=dict)


class NeuraFlowModelInfo(BaseModel):
    """Informações sobre modelos disponíveis no NeuraFlow"""
    available_models: List[ModelMetadata]
    recommended_models: Dict[ModelType, List[str]]
    region_specific_models: Dict[str, List[str]]
    global_models: List[str]