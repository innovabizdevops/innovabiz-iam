"""
INNOVABIZ - Módulo de Autenticação Adaptativa
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Modelos de dados para o sistema de autenticação adaptativa
           baseada em risco, fornecendo autenticação contextual dinâmica.
==================================================================
"""

from datetime import datetime, timedelta
from enum import Enum
import uuid
from typing import Dict, List, Optional, Union
from pydantic import BaseModel, Field, validator


class RiskLevel(str, Enum):
    """Níveis de risco para autenticação adaptativa."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class AuthenticationFactor(str, Enum):
    """Tipos de fatores de autenticação disponíveis."""
    PASSWORD = "password"
    TOTP = "totp"
    SMS = "sms"
    EMAIL = "email"
    PUSH = "push"
    BIOMETRIC = "biometric"
    CERTIFICATE = "certificate"
    HARDWARE_TOKEN = "hardware_token"
    
    # Fatores de autenticação AR
    AR_SPATIAL_GESTURE = "ar_spatial_gesture"
    AR_GAZE_PATTERN = "ar_gaze_pattern"
    AR_ENVIRONMENT = "ar_environment"
    AR_BIOMETRIC = "ar_biometric"


class RiskSignal(BaseModel):
    """Sinal de risco usado para avaliação de autenticação adaptativa."""
    signal_type: str
    signal_value: Union[str, int, float, bool]
    confidence: float = Field(ge=0.0, le=1.0)
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class DeviceFingerprint(BaseModel):
    """Impressão digital do dispositivo para verificação de consistência."""
    device_id: str
    user_agent: str
    os_info: str
    browser_info: str
    screen_resolution: str
    timezone: str
    language: str
    canvas_fingerprint: Optional[str] = None
    webgl_fingerprint: Optional[str] = None
    font_fingerprint: Optional[str] = None
    hardware_concurrency: Optional[int] = None
    is_trusted_device: bool = False
    last_seen: datetime = Field(default_factory=datetime.utcnow)
    risk_score: float = 0.0

    class Config:
        orm_mode = True


class LocationData(BaseModel):
    """Dados de localização para avaliação de risco geográfico."""
    ip_address: str
    country_code: str
    region: str
    city: str
    latitude: float
    longitude: float
    isp: Optional[str] = None
    is_vpn: bool = False
    is_proxy: bool = False
    is_hosting: bool = False
    is_tor: bool = False
    confidence: float = 1.0
    
    class Config:
        orm_mode = True


class BehavioralProfile(BaseModel):
    """Perfil comportamental do usuário para detecção de anomalias."""
    user_id: uuid.UUID
    typical_login_times: List[Dict[str, Union[str, int]]]
    typical_devices: List[str]
    typical_locations: List[Dict[str, str]]
    typical_session_duration: int  # em minutos
    typical_action_patterns: Dict[str, int]
    behavioral_score: float = 0.0
    last_updated: datetime = Field(default_factory=datetime.utcnow)
    
    class Config:
        orm_mode = True


class RiskAssessment(BaseModel):
    """Avaliação de risco para uma tentativa de autenticação."""
    assessment_id: uuid.UUID = Field(default_factory=uuid.uuid4)
    user_id: uuid.UUID
    session_id: Optional[uuid.UUID] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    ip_address: str
    device_fingerprint: DeviceFingerprint
    location_data: LocationData
    risk_signals: List[RiskSignal]
    authentication_context: Dict[str, Union[str, int, bool]] = {}
    risk_level: RiskLevel
    risk_score: float = Field(ge=0.0, le=1.0)
    required_factors: List[AuthenticationFactor] = []
    assessment_reason: str
    
    class Config:
        orm_mode = True


class AdaptivePolicy(BaseModel):
    """Política de autenticação adaptativa configurável por tenant/aplicação."""
    policy_id: uuid.UUID = Field(default_factory=uuid.uuid4)
    tenant_id: uuid.UUID
    application_id: Optional[uuid.UUID] = None
    name: str
    description: str
    is_enabled: bool = True
    risk_threshold_medium: float = 0.3
    risk_threshold_high: float = 0.6
    risk_threshold_critical: float = 0.8
    
    # Fatores exigidos por nível de risco
    factors_low_risk: List[AuthenticationFactor] = [AuthenticationFactor.PASSWORD]
    factors_medium_risk: List[AuthenticationFactor] = [AuthenticationFactor.PASSWORD, AuthenticationFactor.TOTP]
    factors_high_risk: List[AuthenticationFactor] = [AuthenticationFactor.PASSWORD, AuthenticationFactor.TOTP]
    factors_critical_risk: List[AuthenticationFactor] = [
        AuthenticationFactor.PASSWORD, 
        AuthenticationFactor.TOTP, 
        AuthenticationFactor.PUSH
    ]
    
    # Flags de controle de funcionalidade
    enable_geolocation_check: bool = True
    enable_device_fingerprinting: bool = True
    enable_behavioral_analysis: bool = True
    enable_velocity_detection: bool = True
    enable_impossible_travel_detection: bool = True
    
    # Flags para autenticação AR
    enable_ar_authentication: bool = False  # Desativado por padrão
    enable_ar_spatial_gesture: bool = False
    enable_ar_gaze_pattern: bool = False
    enable_ar_environment: bool = False
    enable_ar_biometric: bool = False
    
    # Limiares específicos
    max_failed_attempts: int = 5
    suspicious_country_multiplier: float = 2.0
    new_device_risk_score: float = 0.5
    new_location_risk_score: float = 0.4
    unusual_time_risk_score: float = 0.3
    
    # Configurações avançadas
    anomaly_detection_sensitivity: float = 0.7
    geo_velocity_threshold_kmh: int = 900  # km/h
    behavioral_baseline_days: int = 30
    trusted_device_expiry_days: int = 90
    
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    created_by: Optional[uuid.UUID] = None
    
    class Config:
        orm_mode = True
        
    @validator('factors_medium_risk', 'factors_high_risk', 'factors_critical_risk')
    def validate_factor_progression(cls, v, values):
        """Validar que níveis de risco mais altos exigem mais fatores."""
        if 'factors_low_risk' in values and len(v) < len(values['factors_low_risk']):
            raise ValueError("Níveis de risco mais altos devem exigir pelo menos o mesmo número de fatores")
        return v
