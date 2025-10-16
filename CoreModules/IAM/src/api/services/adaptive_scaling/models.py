"""
Modelos para o serviço de escalonamento adaptativo baseado em TrustScore.

Este módulo define os modelos de dados utilizados no escalonamento
adaptativo de segurança e experiência baseado em pontuações de confiança.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from enum import Enum
from typing import Dict, List, Optional, Union
from datetime import datetime
from pydantic import BaseModel, Field


class ScalingDirection(str, Enum):
    """Direção do escalonamento adaptativo."""
    UP = "up"  # Aumentar restrições/segurança
    DOWN = "down"  # Diminuir restrições/segurança
    MAINTAIN = "maintain"  # Manter nível atual


class SecurityLevel(str, Enum):
    """Níveis de segurança disponíveis."""
    MINIMAL = "minimal"
    LOW = "low"
    STANDARD = "standard"
    HIGH = "high"
    VERY_HIGH = "very_high"
    MAXIMUM = "maximum"


class SecurityMechanism(str, Enum):
    """Mecanismos de segurança que podem ser ajustados."""
    AUTH_FACTORS = "auth_factors"
    SESSION_TIMEOUT = "session_timeout"
    TRANSACTION_LIMITS = "transaction_limits"
    DEVICE_VERIFICATION = "device_verification"
    LOCATION_VERIFICATION = "location_verification"
    BIOMETRIC_REQUIREMENT = "biometric_requirement"
    BEHAVIORAL_ANALYSIS = "behavioral_analysis"
    CONTEXTUAL_AWARENESS = "contextual_awareness"
    CREDENTIAL_COMPLEXITY = "credential_complexity"
    PRIVILEGED_ACCESS = "privileged_access"


class ScalingTrigger(BaseModel):
    """Modelo para gatilhos de escalonamento."""
    id: str = Field(..., description="Identificador único do gatilho")
    name: str = Field(..., description="Nome do gatilho")
    description: str = Field(..., description="Descrição do gatilho")
    dimension: str = Field(..., description="Dimensão de confiança relacionada")
    condition_type: str = Field(..., description="Tipo de condição (threshold, delta, anomaly)")
    threshold_value: float = Field(..., description="Valor limite para acionamento")
    comparison: str = Field(..., description="Operador de comparação (lt, lte, gt, gte, eq)")
    scaling_direction: ScalingDirection = Field(..., description="Direção do escalonamento")
    priority: int = Field(default=1, description="Prioridade do gatilho (maior número = maior prioridade)")
    cooldown_minutes: int = Field(default=60, description="Período mínimo entre acionamentos")
    tenant_specific: bool = Field(default=False, description="Se é específico para um tenant")
    tenant_id: Optional[str] = Field(default=None, description="ID do tenant, se específico")
    region_specific: bool = Field(default=False, description="Se é específico para uma região")
    region_code: Optional[str] = Field(default=None, description="Código da região, se específico")
    context_specific: bool = Field(default=False, description="Se é específico para um contexto")
    context_id: Optional[str] = Field(default=None, description="ID do contexto, se específico")
    enabled: bool = Field(default=True, description="Se o gatilho está ativo")
    created_at: datetime = Field(default_factory=datetime.now, description="Data de criação")
    updated_at: datetime = Field(default_factory=datetime.now, description="Data de atualização")


class SecurityAdjustment(BaseModel):
    """Modelo para ajuste de segurança a ser aplicado."""
    mechanism: SecurityMechanism = Field(..., description="Mecanismo de segurança a ajustar")
    current_level: SecurityLevel = Field(..., description="Nível atual de segurança")
    new_level: SecurityLevel = Field(..., description="Novo nível de segurança")
    parameters: Dict[str, Union[str, int, float, bool]] = Field(default_factory=dict, description="Parâmetros específicos do ajuste")
    reason: str = Field(..., description="Motivo do ajuste")
    expires_at: Optional[datetime] = Field(default=None, description="Expiração do ajuste, se temporário")


class ScalingPolicy(BaseModel):
    """Modelo para política de escalonamento."""
    id: str = Field(..., description="Identificador único da política")
    name: str = Field(..., description="Nome da política")
    description: str = Field(..., description="Descrição da política")
    tenant_id: Optional[str] = Field(default=None, description="ID do tenant, se específico")
    region_code: Optional[str] = Field(default=None, description="Código da região, se específico")
    context_id: Optional[str] = Field(default=None, description="ID do contexto, se específico")
    trigger_ids: List[str] = Field(..., description="IDs dos gatilhos associados")
    adjustment_map: Dict[str, Dict[str, SecurityLevel]] = Field(..., description="Mapeamento de mecanismos/níveis por condição")
    cooldown_minutes: int = Field(default=30, description="Período mínimo entre aplicações")
    enabled: bool = Field(default=True, description="Se a política está ativa")
    priority: int = Field(default=1, description="Prioridade da política (maior número = maior prioridade)")
    created_at: datetime = Field(default_factory=datetime.now, description="Data de criação")
    updated_at: datetime = Field(default_factory=datetime.now, description="Data de atualização")


class ScalingEvent(BaseModel):
    """Modelo para evento de escalonamento."""
    id: str = Field(..., description="Identificador único do evento")
    user_id: str = Field(..., description="ID do usuário afetado")
    tenant_id: str = Field(..., description="ID do tenant")
    context_id: Optional[str] = Field(default=None, description="ID do contexto")
    region_code: Optional[str] = Field(default=None, description="Código da região")
    trigger_id: str = Field(..., description="ID do gatilho acionado")
    policy_id: str = Field(..., description="ID da política aplicada")
    trust_score: float = Field(..., description="Pontuação de confiança no momento")
    dimension_scores: Dict[str, float] = Field(..., description="Pontuações por dimensão")
    scaling_direction: ScalingDirection = Field(..., description="Direção do escalonamento")
    adjustments: List[SecurityAdjustment] = Field(..., description="Ajustes de segurança aplicados")
    event_time: datetime = Field(default_factory=datetime.now, description="Data/hora do evento")
    expires_at: Optional[datetime] = Field(default=None, description="Expiração dos ajustes, se temporários")
    metadata: Dict[str, Union[str, int, float, bool, Dict]] = Field(default_factory=dict, description="Metadados adicionais")


class AdaptiveConfig(BaseModel):
    """Configuração para o sistema adaptativo."""
    enabled: bool = Field(default=True, description="Se o sistema adaptativo está habilitado")
    default_cooldown_minutes: int = Field(default=60, description="Período padrão de cooldown entre escalamentos")
    max_consecutive_escalations: int = Field(default=3, description="Máximo de escalações consecutivas")
    allow_auto_downgrade: bool = Field(default=True, description="Permite redução automática de nível")
    downgrade_delay_minutes: int = Field(default=1440, description="Atraso para redução automática (24h)")
    log_all_evaluations: bool = Field(default=False, description="Registrar todas avaliações, mesmo sem mudança")
    notify_user_on_change: bool = Field(default=True, description="Notificar usuário sobre mudanças")
    override_manual_settings: bool = Field(default=False, description="Sobrescrever configurações manuais")