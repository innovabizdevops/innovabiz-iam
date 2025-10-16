from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Any, Union
from uuid import UUID

from pydantic import BaseModel, Field, validator, root_validator
from pydantic.networks import IPvAnyAddress

from ..models.audit_models import (
    AuditEventCategory, 
    AuditEventSeverity, 
    ComplianceFramework,
    ReportStatus,
    RetentionPolicyType
)


class HttpDetailsModel(BaseModel):
    """Modelo para detalhes de requisições HTTP em eventos de auditoria."""
    method: Optional[str] = Field(None, description="Método HTTP (GET, POST, etc.)")
    url: Optional[str] = Field(None, description="URL da requisição")
    status_code: Optional[int] = Field(None, description="Código de status HTTP")
    user_agent: Optional[str] = Field(None, description="User-Agent do cliente")
    request_id: Optional[str] = Field(None, description="ID de rastreamento da requisição")
    path: Optional[str] = Field(None, description="Caminho da requisição")
    query_params: Optional[Dict[str, str]] = Field(None, description="Parâmetros de consulta")
    headers: Optional[Dict[str, str]] = Field(None, description="Cabeçalhos da requisição (filtrados)")
    
    class Config:
        schema_extra = {
            "example": {
                "method": "POST",
                "url": "/api/users",
                "status_code": 201,
                "user_agent": "Mozilla/5.0",
                "request_id": "req-123-abc",
                "path": "/api/users",
                "query_params": {"tenant": "tenant1"},
                "headers": {"content-type": "application/json"}
            }
        }


class AuditEventBase(BaseModel):
    """Modelo base para eventos de auditoria."""
    category: AuditEventCategory = Field(..., description="Categoria do evento de auditoria")
    action: str = Field(..., description="Ação realizada", min_length=1, max_length=100)
    description: str = Field(..., description="Descrição do evento", min_length=1, max_length=500)
    resource_type: Optional[str] = Field(None, description="Tipo do recurso afetado")
    resource_id: Optional[str] = Field(None, description="ID do recurso afetado")
    resource_name: Optional[str] = Field(None, description="Nome do recurso afetado")
    severity: AuditEventSeverity = Field(default=AuditEventSeverity.INFO, description="Severidade do evento")
    success: bool = Field(..., description="Se a ação foi bem-sucedida")
    error_message: Optional[str] = Field(None, description="Mensagem de erro, se houver")
    details: Optional[Dict[str, Any]] = Field(default={}, description="Detalhes adicionais do evento")
    tags: Optional[List[str]] = Field(default=[], description="Tags para categorização")
    
    # Campos de contexto multi-tenant e multi-regional
    tenant_id: Optional[str] = Field(None, description="ID do tenant")
    regional_context: Optional[str] = Field(None, description="Código do contexto regional (BR, US, EU, AO)")
    country_code: Optional[str] = Field(None, description="Código ISO do país")
    language: Optional[str] = Field(None, description="Código do idioma (pt-BR, en-US, etc.)")
    
    # Campos de usuário e correlação
    user_id: Optional[str] = Field(None, description="ID do usuário que realizou a ação")
    user_name: Optional[str] = Field(None, description="Nome do usuário que realizou a ação")
    correlation_id: Optional[str] = Field(None, description="ID de correlação para rastreamento")
    
    # Campos de rede e HTTP
    source_ip: Optional[IPvAnyAddress] = Field(None, description="Endereço IP de origem")
    http_details: Optional[HttpDetailsModel] = Field(None, description="Detalhes da requisição HTTP")
    
    @validator('regional_context')
    def validate_regional_context(cls, v):
        """Valida o código do contexto regional."""
        if v and v not in ["BR", "US", "EU", "AO"]:
            raise ValueError("Contexto regional deve ser BR, US, EU ou AO")
        return v
    
    @validator('country_code')
    def validate_country_code(cls, v):
        """Valida o código ISO do país."""
        if v and len(v) != 2:
            raise ValueError("Código do país deve seguir o padrão ISO de 2 letras")
        return v if v else None
    
    @root_validator
    def validate_error_message(cls, values):
        """Valida que a mensagem de erro está presente quando success é False."""
        success = values.get('success')
        error_message = values.get('error_message')
        
        if success is False and not error_message:
            raise ValueError("Mensagem de erro é obrigatória quando success é False")
        
        if success is True and error_message:
            values['error_message'] = None  # Remove mensagem de erro quando success é True
        
        return values
    
    class Config:
        use_enum_values = True
        schema_extra = {
            "example": {
                "category": "USER_MANAGEMENT",
                "action": "CREATE_USER",
                "description": "Criação de novo usuário",
                "resource_type": "USER",
                "resource_id": "user-123",
                "resource_name": "john.doe",
                "severity": "INFO",
                "success": True,
                "details": {"role": "admin", "department": "IT"},
                "tags": ["user", "creation"],
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "country_code": "BR",
                "language": "pt-BR",
                "user_id": "admin-456",
                "user_name": "Admin User",
                "correlation_id": "corr-789",
                "source_ip": "192.168.1.1"
            }
        }


class AuditEventCreate(AuditEventBase):
    """Modelo para criação de um evento de auditoria."""
    passclass AuditEventResponse(BaseModel):
    """Modelo de resposta para eventos de auditoria."""
    id: UUID
    tenant_id: str
    regional_context: str
    category: str
    action: str
    description: str
    resource_type: Optional[str]
    resource_id: Optional[str]
    resource_name: Optional[str]
    severity: str
    success: bool
    error_message: Optional[str]
    details: Optional[Dict[str, Any]]
    tags: List[str]
    user_id: Optional[str]
    user_name: Optional[str]
    correlation_id: Optional[str]
    source_ip: Optional[str]
    country_code: Optional[str]
    language: Optional[str]
    http_details: Optional[Dict[str, Any]]
    compliance_frameworks: List[str]
    masked_fields: List[str]
    anonymized_fields: List[str]
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True
        schema_extra = {
            "example": {
                "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "category": "USER_MANAGEMENT",
                "action": "CREATE_USER",
                "description": "Criação de novo usuário",
                "resource_type": "USER",
                "resource_id": "user-123",
                "resource_name": "john.doe",
                "severity": "INFO",
                "success": True,
                "error_message": None,
                "details": {"role": "admin", "department": "IT"},
                "tags": ["user", "creation"],
                "user_id": "admin-456",
                "user_name": "Admin User",
                "correlation_id": "corr-789",
                "source_ip": "192.168.1.1",
                "country_code": "BR",
                "language": "pt-BR",
                "http_details": {"method": "POST", "path": "/api/users"},
                "compliance_frameworks": ["LGPD", "ISO_27001"],
                "masked_fields": [],
                "anonymized_fields": [],
                "created_at": "2023-07-21T15:30:45.123Z",
                "updated_at": "2023-07-21T15:30:45.123Z"
            }
        }


class AuditEventListResponse(BaseModel):
    """Modelo de resposta para listagem paginada de eventos de auditoria."""
    items: List[AuditEventResponse]
    total: int
    limit: int
    offset: int
    
    class Config:
        schema_extra = {
            "example": {
                "items": [
                    {
                        "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
                        "tenant_id": "tenant1",
                        "regional_context": "BR",
                        "category": "USER_MANAGEMENT",
                        "action": "CREATE_USER",
                        "description": "Criação de novo usuário",
                        "severity": "INFO",
                        "success": True,
                        "created_at": "2023-07-21T15:30:45.123Z",
                        "updated_at": "2023-07-21T15:30:45.123Z"
                    }
                ],
                "total": 42,
                "limit": 10,
                "offset": 0
            }
        }


class BatchEventCreate(BaseModel):
    """Modelo para criação de lote de eventos de auditoria."""
    events: List[AuditEventCreate] = Field(..., min_items=1, max_items=100)
    
    class Config:
        schema_extra = {
            "example": {
                "events": [
                    {
                        "category": "USER_MANAGEMENT",
                        "action": "CREATE_USER",
                        "description": "Criação de novo usuário",
                        "resource_type": "USER",
                        "resource_id": "user-123",
                        "success": True,
                        "tenant_id": "tenant1",
                        "regional_context": "BR"
                    },
                    {
                        "category": "AUTHENTICATION",
                        "action": "USER_LOGIN",
                        "description": "Login de usuário",
                        "resource_type": "USER",
                        "resource_id": "user-123",
                        "success": True,
                        "tenant_id": "tenant1",
                        "regional_context": "BR"
                    }
                ]
            }
        }


class AuditRetentionPolicyBase(BaseModel):
    """Modelo base para políticas de retenção de auditoria."""
    tenant_id: Optional[str] = Field(None, description="ID do tenant")
    regional_context: Optional[str] = Field(None, description="Código do contexto regional (BR, US, EU, AO)")
    compliance_framework: ComplianceFramework = Field(
        ..., 
        description="Framework de compliance associado"
    )
    retention_period_days: int = Field(
        ..., 
        description="Período de retenção em dias",
        ge=1,
        le=3650  # Máximo de 10 anos
    )
    policy_type: RetentionPolicyType = Field(
        default=RetentionPolicyType.ANONYMIZATION,
        description="Tipo de política de retenção"
    )
    fields_to_anonymize: Optional[List[str]] = Field(
        default=[],
        description="Campos a serem anonimizados após o período de retenção"
    )
    description: Optional[str] = Field(
        None, 
        description="Descrição da política",
        max_length=500
    )
    active: bool = Field(
        default=True,
        description="Se a política está ativa"
    )
    
    @validator('regional_context')
    def validate_regional_context(cls, v):
        """Valida o código do contexto regional."""
        if v and v not in ["BR", "US", "EU", "AO"]:
            raise ValueError("Contexto regional deve ser BR, US, EU ou AO")
        return v
    
    @root_validator
    def validate_anonymization_fields(cls, values):
        """Valida campos de anonimização conforme o tipo de política."""
        policy_type = values.get('policy_type')
        fields_to_anonymize = values.get('fields_to_anonymize', [])
        
        if policy_type == RetentionPolicyType.ANONYMIZATION and not fields_to_anonymize:
            raise ValueError("Campos de anonimização são obrigatórios para políticas de anonimização")
        
        if policy_type == RetentionPolicyType.DELETION and fields_to_anonymize:
            values['fields_to_anonymize'] = []  # Remove campos para políticas de exclusão
        
        return values
    
    class Config:
        use_enum_values = True
        schema_extra = {
            "example": {
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "compliance_framework": "LGPD",
                "retention_period_days": 730,  # 2 anos
                "policy_type": "ANONYMIZATION",
                "fields_to_anonymize": ["user_id", "user_name", "source_ip"],
                "description": "Política LGPD para anonimização de dados pessoais",
                "active": True
            }
        }class AuditRetentionPolicyCreate(AuditRetentionPolicyBase):
    """Modelo para criação de políticas de retenção de auditoria."""
    pass


class AuditRetentionPolicyUpdate(BaseModel):
    """Modelo para atualização de políticas de retenção de auditoria."""
    retention_period_days: Optional[int] = Field(
        None, 
        description="Período de retenção em dias",
        ge=1,
        le=3650  # Máximo de 10 anos
    )
    policy_type: Optional[RetentionPolicyType] = Field(
        None,
        description="Tipo de política de retenção"
    )
    fields_to_anonymize: Optional[List[str]] = Field(
        None,
        description="Campos a serem anonimizados após o período de retenção"
    )
    description: Optional[str] = Field(
        None, 
        description="Descrição da política",
        max_length=500
    )
    active: Optional[bool] = Field(
        None,
        description="Se a política está ativa"
    )
    
    @root_validator
    def validate_anonymization_fields(cls, values):
        """Valida campos de anonimização conforme o tipo de política."""
        policy_type = values.get('policy_type')
        fields_to_anonymize = values.get('fields_to_anonymize')
        
        if (policy_type == RetentionPolicyType.ANONYMIZATION and 
            fields_to_anonymize is not None and len(fields_to_anonymize) == 0):
            raise ValueError("Campos de anonimização não podem estar vazios para políticas de anonimização")
        
        if policy_type == RetentionPolicyType.DELETION and fields_to_anonymize:
            values['fields_to_anonymize'] = []  # Remove campos para políticas de exclusão
        
        return values
    
    class Config:
        use_enum_values = True
        schema_extra = {
            "example": {
                "retention_period_days": 1095,  # Atualiza para 3 anos
                "description": "Política LGPD atualizada para anonimização de dados pessoais",
                "active": True
            }
        }


class AuditRetentionPolicyResponse(BaseModel):
    """Modelo de resposta para políticas de retenção de auditoria."""
    id: UUID
    tenant_id: str
    regional_context: Optional[str]
    compliance_framework: str
    retention_period_days: int
    policy_type: str
    fields_to_anonymize: List[str]
    description: Optional[str]
    active: bool
    created_at: datetime
    updated_at: datetime
    created_by: Optional[str]
    updated_by: Optional[str]
    
    class Config:
        orm_mode = True
        schema_extra = {
            "example": {
                "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "compliance_framework": "LGPD",
                "retention_period_days": 730,
                "policy_type": "ANONYMIZATION",
                "fields_to_anonymize": ["user_id", "user_name", "source_ip"],
                "description": "Política LGPD para anonimização de dados pessoais",
                "active": True,
                "created_at": "2023-07-21T15:30:45.123Z",
                "updated_at": "2023-07-21T15:30:45.123Z",
                "created_by": "admin-user",
                "updated_by": null
            }
        }


class AuditComplianceReportResponse(BaseModel):
    """Modelo de resposta para relatórios de compliance de auditoria."""
    id: UUID
    tenant_id: str
    regional_context: str
    compliance_framework: str
    report_type: str
    report_format: str
    start_date: datetime
    end_date: datetime
    status: str
    report_data: Optional[Dict[str, Any]]
    error_message: Optional[str]
    include_anonymized: bool
    created_at: datetime
    completed_at: Optional[datetime]
    created_by: Optional[str]
    
    class Config:
        orm_mode = True
        schema_extra = {
            "example": {
                "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "compliance_framework": "LGPD",
                "report_type": "standard",
                "report_format": "json",
                "start_date": "2023-07-01T00:00:00Z",
                "end_date": "2023-07-31T23:59:59Z",
                "status": "COMPLETED",
                "report_data": {
                    "total_events": 1250,
                    "events_by_category": {
                        "USER_MANAGEMENT": 320,
                        "AUTHENTICATION": 720,
                        "DATA_ACCESS": 210
                    },
                    "events_by_severity": {
                        "INFO": 950,
                        "WARNING": 280,
                        "ERROR": 20
                    }
                },
                "error_message": null,
                "include_anonymized": false,
                "created_at": "2023-08-01T10:15:30Z",
                "completed_at": "2023-08-01T10:16:45Z",
                "created_by": "admin-user"
            }
        }


class AuditStatisticsResponse(BaseModel):
    """Modelo de resposta para estatísticas de auditoria."""
    tenant_id: str
    regional_context: str
    period: str
    generated_at: Optional[str]
    statistics: Dict[str, Any]
    
    class Config:
        schema_extra = {
            "example": {
                "tenant_id": "tenant1",
                "regional_context": "BR",
                "period": "daily",
                "generated_at": "2023-07-21T15:30:45.123Z",
                "statistics": {
                    "2023-07-21": {
                        "total": 450,
                        "by_category": {
                            "USER_MANAGEMENT": 120,
                            "AUTHENTICATION": 280,
                            "DATA_ACCESS": 50
                        },
                        "by_severity": {
                            "INFO": 350,
                            "WARNING": 80,
                            "ERROR": 20
                        },
                        "success_rate": 0.95
                    }
                }
            }
        }