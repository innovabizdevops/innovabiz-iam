"""
Modelos de dados para integração com TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Dict, List, Optional, Set, Tuple, Union
from uuid import UUID, uuid4

from pydantic import BaseModel, Field, validator


class AuthenticationLevel(str, Enum):
    """Níveis de autenticação do TrustGuard"""
    VERY_LOW = "very_low"       # Muito baixo
    LOW = "low"                 # Baixo
    MEDIUM = "medium"           # Médio
    HIGH = "high"               # Alto
    VERY_HIGH = "very_high"     # Muito alto


class AccessDecision(str, Enum):
    """Decisões de acesso do TrustGuard"""
    ALLOW = "allow"             # Permitir
    DENY = "deny"               # Negar
    CHALLENGE = "challenge"     # Desafiar
    STEP_UP = "step_up"         # Aumentar autenticação
    MONITOR = "monitor"         # Monitorar
    RESTRICT = "restrict"       # Restringir


class RiskLevel(str, Enum):
    """Níveis de risco do TrustGuard"""
    VERY_LOW = "very_low"       # Muito baixo
    LOW = "low"                 # Baixo
    MEDIUM = "medium"           # Médio
    HIGH = "high"               # Alto
    VERY_HIGH = "very_high"     # Muito alto


class SessionStatus(str, Enum):
    """Status de sessão do TrustGuard"""
    ACTIVE = "active"           # Ativa
    EXPIRED = "expired"         # Expirada
    REVOKED = "revoked"         # Revogada
    LOCKED = "locked"           # Bloqueada
    SUSPICIOUS = "suspicious"   # Suspeita


class AuthMethod(str, Enum):
    """Métodos de autenticação suportados"""
    PASSWORD = "password"                       # Senha
    OTP = "otp"                                 # One-time Password
    TOTP = "totp"                               # Time-based OTP
    FIDO2 = "fido2"                             # FIDO2/WebAuthn
    BIOMETRIC = "biometric"                     # Biometria
    FACIAL_RECOGNITION = "facial_recognition"   # Reconhecimento facial
    FINGERPRINT = "fingerprint"                 # Impressão digital
    VOICE = "voice"                             # Reconhecimento de voz
    BEHAVIORAL = "behavioral"                   # Comportamental
    SMS = "sms"                                 # SMS
    EMAIL = "email"                             # Email
    SOCIAL = "social"                           # Social
    OAUTH = "oauth"                             # OAuth
    SAML = "saml"                               # SAML
    X509 = "x509"                               # Certificado X.509
    JWT = "jwt"                                 # JWT
    API_KEY = "api_key"                         # API Key


class ResourceType(str, Enum):
    """Tipos de recursos protegidos"""
    API = "api"                 # API
    WEB_PAGE = "web_page"       # Página Web
    ENDPOINT = "endpoint"       # Endpoint
    SERVICE = "service"         # Serviço
    FILE = "file"               # Arquivo
    DATABASE = "database"       # Banco de dados
    FUNCTION = "function"       # Função
    APPLICATION = "application" # Aplicação


class ActionType(str, Enum):
    """Tipos de ações"""
    READ = "read"               # Leitura
    WRITE = "write"             # Escrita
    UPDATE = "update"           # Atualização
    DELETE = "delete"           # Exclusão
    EXECUTE = "execute"         # Execução
    ADMIN = "admin"             # Administração
    QUERY = "query"             # Consulta


class PolicyEffect(str, Enum):
    """Efeitos de políticas de acesso"""
    ALLOW = "allow"             # Permitir
    DENY = "deny"               # Negar
    CHALLENGE = "challenge"     # Desafiar


class TrustScore(BaseModel):
    """Score de confiança do TrustGuard"""
    score: float = Field(..., description="Score de confiança (0.0 - 1.0)")
    confidence: float = Field(..., description="Confiança no score (0.0 - 1.0)")
    factors: List[Dict[str, Any]] = Field(default_factory=list, 
                                         description="Fatores que influenciam o score")
    timestamp: datetime = Field(default_factory=datetime.now, 
                              description="Timestamp da avaliação")
    
    @validator("score", "confidence")
    def validate_score(cls, v):
        if not 0 <= v <= 1:
            raise ValueError("Score must be between 0.0 and 1.0")
        return v


class UserRisk(BaseModel):
    """Avaliação de risco do usuário"""
    user_id: str = Field(..., description="ID do usuário")
    risk_level: RiskLevel = Field(..., description="Nível de risco")
    risk_score: float = Field(..., description="Score de risco (0.0 - 1.0)")
    risk_factors: List[Dict[str, Any]] = Field(default_factory=list, 
                                             description="Fatores de risco")
    last_update: datetime = Field(default_factory=datetime.now, 
                                description="Timestamp da última atualização")
    
    @validator("risk_score")
    def validate_risk_score(cls, v):
        if not 0 <= v <= 1:
            raise ValueError("Risk score must be between 0.0 and 1.0")
        return v


class UserContext(BaseModel):
    """Contexto do usuário para avaliação de acesso"""
    user_id: str = Field(..., description="ID do usuário")
    ip_address: Optional[str] = Field(None, description="Endereço IP")
    device_id: Optional[str] = Field(None, description="ID do dispositivo")
    user_agent: Optional[str] = Field(None, description="User-Agent")
    geo_location: Optional[Dict[str, Any]] = Field(None, description="Geolocalização")
    timestamp: datetime = Field(default_factory=datetime.now, description="Timestamp")
    session_id: Optional[str] = Field(None, description="ID da sessão")
    auth_methods: List[AuthMethod] = Field(default_factory=list, 
                                         description="Métodos de autenticação usados")
    auth_level: AuthenticationLevel = Field(AuthenticationLevel.LOW, 
                                         description="Nível de autenticação")
    roles: List[str] = Field(default_factory=list, description="Papéis do usuário")
    groups: List[str] = Field(default_factory=list, description="Grupos do usuário")
    permissions: List[str] = Field(default_factory=list, description="Permissões do usuário")
    attributes: Dict[str, Any] = Field(default_factory=dict, description="Atributos do usuário")
    risk_profile: Optional[UserRisk] = Field(None, description="Perfil de risco do usuário")
    trust_score: Optional[TrustScore] = Field(None, description="Score de confiança")


class ResourceContext(BaseModel):
    """Contexto do recurso para avaliação de acesso"""
    resource_id: str = Field(..., description="ID do recurso")
    resource_type: ResourceType = Field(..., description="Tipo do recurso")
    resource_path: Optional[str] = Field(None, description="Caminho do recurso")
    action: ActionType = Field(..., description="Ação solicitada")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados do recurso")
    sensitivity: Optional[str] = Field(None, description="Nível de sensibilidade")
    owner: Optional[str] = Field(None, description="Proprietário do recurso")
    tags: List[str] = Field(default_factory=list, description="Tags do recurso")
    required_auth_level: AuthenticationLevel = Field(AuthenticationLevel.LOW, 
                                                  description="Nível de autenticação necessário")


class AccessRequest(BaseModel):
    """Solicitação de acesso a um recurso"""
    request_id: str = Field(default_factory=lambda: str(uuid4()), description="ID da solicitação")
    user_context: UserContext = Field(..., description="Contexto do usuário")
    resource_context: ResourceContext = Field(..., description="Contexto do recurso")
    timestamp: datetime = Field(default_factory=datetime.now, description="Timestamp")
    session_data: Dict[str, Any] = Field(default_factory=dict, description="Dados da sessão")
    transaction_data: Optional[Dict[str, Any]] = Field(None, description="Dados da transação")
    environment: Dict[str, Any] = Field(default_factory=dict, description="Dados do ambiente")


class PolicyCondition(BaseModel):
    """Condição para uma política de acesso"""
    attribute: str = Field(..., description="Atributo a ser avaliado")
    operator: str = Field(..., description="Operador (eq, neq, lt, gt, in, etc.)")
    value: Any = Field(..., description="Valor para comparação")


class Policy(BaseModel):
    """Política de acesso no TrustGuard"""
    policy_id: str = Field(default_factory=lambda: str(uuid4()), description="ID da política")
    name: str = Field(..., description="Nome da política")
    description: Optional[str] = Field(None, description="Descrição da política")
    effect: PolicyEffect = Field(..., description="Efeito da política")
    priority: int = Field(0, description="Prioridade da política (maior = mais prioritário)")
    conditions: List[PolicyCondition] = Field(default_factory=list, 
                                            description="Condições da política")
    resource_types: List[ResourceType] = Field(default_factory=list, 
                                             description="Tipos de recursos aplicáveis")
    actions: List[ActionType] = Field(default_factory=list, description="Ações aplicáveis")
    user_attributes: Dict[str, List[str]] = Field(default_factory=dict, 
                                                description="Atributos de usuário aplicáveis")
    resource_attributes: Dict[str, List[str]] = Field(default_factory=dict, 
                                                    description="Atributos de recurso aplicáveis")
    environment_attributes: Dict[str, List[str]] = Field(default_factory=dict, 
                                                      description="Atributos de ambiente aplicáveis")
    created_at: datetime = Field(default_factory=datetime.now, description="Data de criação")
    updated_at: datetime = Field(default_factory=datetime.now, description="Data de atualização")
    version: int = Field(1, description="Versão da política")
    active: bool = Field(True, description="Política está ativa")
    owner: Optional[str] = Field(None, description="Proprietário da política")


class AccessDecisionResponse(BaseModel):
    """Resposta de decisão de acesso"""
    request_id: str = Field(..., description="ID da solicitação")
    decision: AccessDecision = Field(..., description="Decisão de acesso")
    timestamp: datetime = Field(default_factory=datetime.now, description="Timestamp")
    reason: Optional[str] = Field(None, description="Razão da decisão")
    policies_evaluated: List[str] = Field(default_factory=list, 
                                        description="IDs das políticas avaliadas")
    risk_level: RiskLevel = Field(RiskLevel.LOW, description="Nível de risco avaliado")
    auth_level: AuthenticationLevel = Field(AuthenticationLevel.LOW, 
                                         description="Nível de autenticação atual")
    required_auth_level: Optional[AuthenticationLevel] = Field(None, 
                                                            description="Nível de autenticação necessário")
    session_ttl: Optional[int] = Field(None, description="Tempo de vida da sessão em segundos")
    obligations: List[Dict[str, Any]] = Field(default_factory=list, 
                                            description="Obrigações a serem cumpridas")
    advice: List[Dict[str, Any]] = Field(default_factory=list, 
                                       description="Recomendações não obrigatórias")


class SessionInfo(BaseModel):
    """Informações sobre uma sessão de usuário"""
    session_id: str = Field(..., description="ID da sessão")
    user_id: str = Field(..., description="ID do usuário")
    status: SessionStatus = Field(SessionStatus.ACTIVE, description="Status da sessão")
    created_at: datetime = Field(default_factory=datetime.now, description="Data de criação")
    expires_at: datetime = Field(..., description="Data de expiração")
    last_activity: datetime = Field(default_factory=datetime.now, description="Última atividade")
    auth_methods: List[AuthMethod] = Field(default_factory=list, 
                                         description="Métodos de autenticação usados")
    auth_level: AuthenticationLevel = Field(AuthenticationLevel.LOW, 
                                         description="Nível de autenticação")
    ip_address: Optional[str] = Field(None, description="Endereço IP")
    device_id: Optional[str] = Field(None, description="ID do dispositivo")
    user_agent: Optional[str] = Field(None, description="User-Agent")
    geo_location: Optional[Dict[str, Any]] = Field(None, description="Geolocalização")
    risk_level: RiskLevel = Field(RiskLevel.LOW, description="Nível de risco")
    trust_score: Optional[TrustScore] = Field(None, description="Score de confiança")
    context: Dict[str, Any] = Field(default_factory=dict, description="Contexto da sessão")
    
    @validator("expires_at")
    def validate_expires_at(cls, v, values):
        if "created_at" in values and v <= values["created_at"]:
            raise ValueError("Expiration date must be after creation date")
        return v


class AuthenticationRequest(BaseModel):
    """Solicitação de autenticação"""
    user_id: str = Field(..., description="ID do usuário")
    auth_method: AuthMethod = Field(..., description="Método de autenticação")
    credentials: Dict[str, Any] = Field(..., description="Credenciais")
    ip_address: Optional[str] = Field(None, description="Endereço IP")
    device_id: Optional[str] = Field(None, description="ID do dispositivo")
    user_agent: Optional[str] = Field(None, description="User-Agent")
    geo_location: Optional[Dict[str, Any]] = Field(None, description="Geolocalização")
    request_id: str = Field(default_factory=lambda: str(uuid4()), description="ID da solicitação")
    timestamp: datetime = Field(default_factory=datetime.now, description="Timestamp")
    context: Dict[str, Any] = Field(default_factory=dict, description="Contexto da autenticação")


class AuthenticationResponse(BaseModel):
    """Resposta de autenticação"""
    request_id: str = Field(..., description="ID da solicitação")
    success: bool = Field(..., description="Sucesso da autenticação")
    user_id: str = Field(..., description="ID do usuário")
    timestamp: datetime = Field(default_factory=datetime.now, description="Timestamp")
    session_id: Optional[str] = Field(None, description="ID da sessão gerada")
    auth_level: AuthenticationLevel = Field(..., description="Nível de autenticação")
    token: Optional[str] = Field(None, description="Token de autenticação")
    token_type: Optional[str] = Field(None, description="Tipo do token")
    expires_in: Optional[int] = Field(None, description="Tempo de expiração do token em segundos")
    refresh_token: Optional[str] = Field(None, description="Token de atualização")
    error: Optional[str] = Field(None, description="Mensagem de erro")
    error_code: Optional[str] = Field(None, description="Código de erro")
    additional_factors: List[AuthMethod] = Field(default_factory=list, 
                                              description="Fatores adicionais necessários")
    risk_level: RiskLevel = Field(RiskLevel.LOW, description="Nível de risco avaliado")