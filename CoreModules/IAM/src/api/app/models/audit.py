"""
INNOVABIZ IAM - Modelos de Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Modelos de dados para o sistema de auditoria multi-contexto
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA
"""

from enum import Enum
from typing import Dict, List, Optional, Any, Union
from datetime import datetime
from pydantic import BaseModel, Field, validator, root_validator


class AuditEventCategory(str, Enum):
    """Categorias de eventos de auditoria."""
    AUTHENTICATION = "AUTHENTICATION"           # Eventos de login, logout, etc.
    AUTHORIZATION = "AUTHORIZATION"             # Eventos de autorização de recursos
    USER_MANAGEMENT = "USER_MANAGEMENT"         # Gerenciamento de usuários
    DATA_ACCESS = "DATA_ACCESS"                 # Acesso a dados (leitura)
    DATA_MODIFICATION = "DATA_MODIFICATION"     # Modificação de dados
    CONFIGURATION = "CONFIGURATION"             # Alterações de configuração
    SECURITY = "SECURITY"                       # Eventos de segurança
    PRIVACY = "PRIVACY"                         # Eventos relacionados à privacidade
    SYSTEM = "SYSTEM"                           # Eventos de sistema
    APPLICATION = "APPLICATION"                 # Eventos da aplicação
    RESOURCE_MANAGEMENT = "RESOURCE_MANAGEMENT" # Gerenciamento de recursos
    API = "API"                                 # Chamadas de API
    PAYMENT = "PAYMENT"                         # Eventos de pagamento
    CARD_DATA = "CARD_DATA"                     # Eventos relacionados a cartões
    TRANSACTION = "TRANSACTION"                 # Transações financeiras
    RISK = "RISK"                               # Avaliação de risco
    CONSENT = "CONSENT"                         # Consentimento (GDPR/LGPD)
    EXTERNAL = "EXTERNAL"                       # Integrações externas
    OTHER = "OTHER"                             # Outros eventos


class AuditEventSeverity(str, Enum):
    """Níveis de severidade para eventos de auditoria."""
    CRITICAL = "CRITICAL"   # Eventos críticos de segurança/compliance
    HIGH = "HIGH"           # Alta severidade
    MEDIUM = "MEDIUM"       # Severidade média
    LOW = "LOW"             # Baixa severidade
    INFO = "INFO"           # Informacional


class ComplianceFramework(str, Enum):
    """Frameworks de compliance suportados."""
    GDPR = "GDPR"           # General Data Protection Regulation (UE)
    LGPD = "LGPD"           # Lei Geral de Proteção de Dados (Brasil)
    PCI_DSS = "PCI_DSS"     # Payment Card Industry Data Security Standard
    PSD2 = "PSD2"           # Payment Services Directive 2 (UE)
    BACEN = "BACEN"         # Banco Central do Brasil
    BNA = "BNA"             # Banco Nacional de Angola
    ISO_27001 = "ISO_27001" # ISO 27001 (Segurança da Informação)
    SOX = "SOX"             # Sarbanes-Oxley


class RegionalCompliance(BaseModel):
    """
    Requisitos de compliance por região/jurisdição.
    Usado para registrar quais frameworks de compliance
    se aplicam a um evento específico.
    """
    frameworks: List[ComplianceFramework] = Field(
        default_factory=list,
        description="Frameworks de compliance aplicáveis"
    )
    data_residency: Optional[str] = Field(
        None,
        description="Requisito de residência de dados"
    )
    data_retention: Optional[int] = Field(
        None,
        description="Período de retenção de dados em dias"
    )
    required_fields: List[str] = Field(
        default_factory=list,
        description="Campos obrigatórios para compliance"
    )
    sensitive_fields: List[str] = Field(
        default_factory=list,
        description="Campos sensíveis que precisam de proteção adicional"
    )


class AuditHttpDetails(BaseModel):
    """
    Detalhes de requisições HTTP para eventos relacionados.
    Usado para eventos de API e HTTP.
    """
    method: str = Field(..., description="Método HTTP")
    path: str = Field(..., description="Caminho da requisição")
    query_params: Optional[Dict[str, str]] = Field(
        None, 
        description="Parâmetros de query (sem dados sensíveis)"
    )
    headers: Optional[Dict[str, str]] = Field(
        None, 
        description="Headers relevantes (sem dados sensíveis)"
    )
    status_code: Optional[int] = Field(
        None, 
        description="Código de status HTTP da resposta"
    )
    request_id: Optional[str] = Field(
        None,
        description="ID único da requisição para correlação"
    )
    client_ip: Optional[str] = Field(
        None,
        description="IP do cliente (anonimizado se necessário)"
    )
    user_agent: Optional[str] = Field(
        None,
        description="User-Agent do cliente"
    )
    duration_ms: Optional[int] = Field(
        None,
        description="Duração da requisição em milissegundos"
    )


class AuditEventBase(BaseModel):
    """Modelo base para eventos de auditoria."""
    category: AuditEventCategory = Field(
        ...,
        description="Categoria do evento de auditoria"
    )
    severity: AuditEventSeverity = Field(
        default=AuditEventSeverity.INFO,
        description="Nível de severidade do evento"
    )
    action: str = Field(
        ...,
        description="Ação realizada (ex: user.create, data.read)"
    )
    timestamp: Optional[datetime] = Field(
        None,
        description="Timestamp UTC do evento (preenchido automaticamente se omitido)"
    )
    user_id: Optional[str] = Field(
        None,
        description="ID do usuário que realizou a ação"
    )
    user_name: Optional[str] = Field(
        None,
        description="Nome do usuário que realizou a ação"
    )
    resource_type: Optional[str] = Field(
        None,
        description="Tipo do recurso afetado (ex: user, role, permission)"
    )
    resource_id: Optional[str] = Field(
        None,
        description="ID do recurso afetado"
    )
    resource_name: Optional[str] = Field(
        None,
        description="Nome do recurso afetado"
    )
    description: str = Field(
        ...,
        description="Descrição detalhada do evento"
    )
    success: bool = Field(
        True,
        description="Indica se a ação foi bem-sucedida"
    )
    error_message: Optional[str] = Field(
        None,
        description="Mensagem de erro se a ação falhou"
    )
    tenant_id: Optional[str] = Field(
        None,
        description="ID do tenant (isolamento multi-tenant)"
    )
    regional_context: Optional[str] = Field(
        None,
        description="Contexto regional (BR, US, EU, AO)"
    )
    country_code: Optional[str] = Field(
        None,
        description="Código ISO do país"
    )
    language: Optional[str] = Field(
        None,
        description="Código de idioma (ex: pt-BR, en-US)"
    )
    tags: List[str] = Field(
        default_factory=list,
        description="Tags para classificação e busca"
    )
    compliance: Optional[RegionalCompliance] = Field(
        None,
        description="Informações de compliance aplicáveis"
    )
    compliance_tags: List[str] = Field(
        default_factory=list,
        description="Tags de compliance (ex: PII, PCI, GDPR)"
    )
    correlation_id: Optional[str] = Field(
        None,
        description="ID de correlação para eventos relacionados"
    )
    http_details: Optional[AuditHttpDetails] = Field(
        None,
        description="Detalhes HTTP (para eventos de API)"
    )
    source_ip: Optional[str] = Field(
        None,
        description="IP de origem da ação (anonimizado se necessário)"
    )
    source_system: Optional[str] = Field(
        None,
        description="Sistema de origem (ex: web, mobile, api)"
    )

    @validator("timestamp", pre=True, always=True)
    def set_timestamp_default(cls, v):
        """Define timestamp atual se não fornecido."""
        if v is None:
            return datetime.utcnow()
        return v

    @validator("tags", "compliance_tags", pre=True)
    def ensure_list(cls, v):
        """Garante que tags sempre sejam listas."""
        if v is None:
            return []
        return v


class AuditEventCreate(AuditEventBase):
    """
    Modelo para criação de eventos de auditoria.
    Usado ao registrar um novo evento.
    """
    details: Optional[Dict[str, Any]] = Field(
        None,
        description="Detalhes adicionais específicos do evento (pode conter dados estruturados)"
    )


class AuditEventRead(AuditEventBase):
    """
    Modelo para leitura de eventos de auditoria.
    Usado ao retornar eventos das APIs.
    """
    id: str = Field(..., description="ID único do evento")
    details: Optional[Dict[str, Any]] = Field(
        None,
        description="Detalhes adicionais específicos do evento"
    )
    masked_fields: List[str] = Field(
        default_factory=list,
        description="Lista de campos que foram mascarados por motivos de privacidade"
    )

    class Config:
        orm_mode = True


class AuditEventFilter(BaseModel):
    """
    Filtros para consulta de eventos de auditoria.
    Usado para filtrar eventos nas APIs de consulta.
    """
    category: Optional[AuditEventCategory] = Field(
        None,
        description="Filtrar por categoria"
    )
    severity: Optional[AuditEventSeverity] = Field(
        None,
        description="Filtrar por severidade"
    )
    action: Optional[str] = Field(
        None,
        description="Filtrar por ação"
    )
    start_date: Optional[datetime] = Field(
        None,
        description="Data inicial para filtro (inclusiva)"
    )
    end_date: Optional[datetime] = Field(
        None,
        description="Data final para filtro (inclusiva)"
    )
    user_id: Optional[str] = Field(
        None,
        description="Filtrar por usuário"
    )
    resource_type: Optional[str] = Field(
        None,
        description="Filtrar por tipo de recurso"
    )
    resource_id: Optional[str] = Field(
        None,
        description="Filtrar por ID de recurso"
    )
    success: Optional[bool] = Field(
        None,
        description="Filtrar por sucesso/falha"
    )
    tenant_id: Optional[str] = Field(
        None,
        description="Filtrar por tenant"
    )
    regional_context: Optional[str] = Field(
        None,
        description="Filtrar por contexto regional"
    )
    country_code: Optional[str] = Field(
        None,
        description="Filtrar por código de país"
    )
    tag: Optional[str] = Field(
        None,
        description="Filtrar por tag"
    )
    compliance_tag: Optional[str] = Field(
        None,
        description="Filtrar por tag de compliance"
    )
    correlation_id: Optional[str] = Field(
        None,
        description="Filtrar por ID de correlação"
    )
    status_code: Optional[int] = Field(
        None,
        description="Filtrar por código de status HTTP"
    )

    @root_validator(pre=True)
    def validate_date_range(cls, values):
        """Valida que start_date é anterior a end_date."""
        start_date = values.get("start_date")
        end_date = values.get("end_date")
        
        if start_date and end_date and start_date > end_date:
            raise ValueError("start_date deve ser anterior a end_date")
            
        return values


class ComplianceStatus(str, Enum):
    """Status de compliance para relatórios."""
    COMPLIANT = "COMPLIANT"               # Totalmente conforme
    PARTIALLY_COMPLIANT = "PARTIALLY_COMPLIANT"  # Parcialmente conforme
    NON_COMPLIANT = "NON_COMPLIANT"       # Não conforme
    UNKNOWN = "UNKNOWN"                   # Status desconhecido


class ComplianceReportSummary(BaseModel):
    """Resumo de relatório de compliance."""
    framework: ComplianceFramework = Field(..., description="Framework avaliado")
    status: ComplianceStatus = Field(..., description="Status de compliance")
    score: int = Field(..., ge=0, le=100, description="Pontuação de compliance (0-100)")
    issues: List[str] = Field(default_factory=list, description="Problemas encontrados")
    recommendations: List[str] = Field(default_factory=list, description="Recomendações")
    period_start: datetime = Field(..., description="Início do período avaliado")
    period_end: datetime = Field(..., description="Fim do período avaliado")
    events_analyzed: int = Field(..., ge=0, description="Quantidade de eventos analisados")