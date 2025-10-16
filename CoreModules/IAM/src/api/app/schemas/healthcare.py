from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field, validator
from uuid import UUID
from datetime import datetime

class HealthcareComplianceCheck(BaseModel):
    name: str
    requirement: str
    status: str
    details: Optional[str] = None
    
    @validator('status')
    def validate_status(cls, v):
        valid_status = ["passed", "failed", "warning", "not_applicable"]
        if v not in valid_status:
            raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v

class HealthcareComplianceResult(BaseModel):
    regulation: str
    timestamp: datetime
    validator: str
    status: str
    score: int = Field(..., ge=0, le=100)
    checks: List[HealthcareComplianceCheck]
    remediation_plan: Optional[str] = None
    
    @validator('regulation')
    def validate_regulation(cls, v):
        valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb", "all"]
        if v not in valid_regulations:
            raise ValueError(f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}")
        return v
    
    @validator('status')
    def validate_status(cls, v):
        valid_status = ["passed", "failed", "warning", "not_applicable"]
        if v not in valid_status:
            raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v

class HealthcareComplianceValidationSummary(BaseModel):
    id: UUID
    validation_timestamp: datetime
    regulation: str
    validator_name: str
    status: str
    score: int = Field(..., ge=0, le=100)
    validated_by: Optional[UUID] = None
    
    @validator('regulation')
    def validate_regulation(cls, v):
        valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb"]
        if v not in valid_regulations:
            raise ValueError(f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}")
        return v
    
    @validator('status')
    def validate_status(cls, v):
        valid_status = ["passed", "failed", "warning", "not_applicable"]
        if v not in valid_status:
            raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v
    
    class Config:
        orm_mode = True

class HealthcareComplianceValidationRequest(BaseModel):
    regulation: str
    parameters: Optional[Dict[str, Any]] = None
    
    @validator('regulation')
    def validate_regulation(cls, v):
        valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb", "all"]
        if v not in valid_regulations:
            raise ValueError(f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}")
        return v

class HealthcareComplianceHistoryFilter(BaseModel):
    page: int = 1
    per_page: int = Field(20, ge=1, le=100)
    regulation: Optional[str] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    status: Optional[str] = None
    
    @validator('regulation')
    def validate_regulation(cls, v):
        if v is not None:
            valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb", "all"]
            if v not in valid_regulations:
                raise ValueError(f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}")
        return v
    
    @validator('status')
    def validate_status(cls, v):
        if v is not None:
            valid_status = ["passed", "failed", "warning", "not_applicable"]
            if v not in valid_status:
                raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v

class ComplianceRequirement(BaseModel):
    id: UUID
    regulation: str
    code: str
    name: str
    description: str
    validation_method: str
    severity: str
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None
    
    @validator('regulation')
    def validate_regulation(cls, v):
        valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb"]
        if v not in valid_regulations:
            raise ValueError(f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}")
        return v
    
    @validator('validation_method')
    def validate_validation_method(cls, v):
        valid_methods = ["automatic", "manual", "hybrid"]
        if v not in valid_methods:
            raise ValueError(f"Método de validação inválido. Deve ser um dos seguintes: {', '.join(valid_methods)}")
        return v
    
    @validator('severity')
    def validate_severity(cls, v):
        valid_severities = ["critical", "high", "medium", "low"]
        if v not in valid_severities:
            raise ValueError(f"Severidade inválida. Deve ser uma das seguintes: {', '.join(valid_severities)}")
        return v
    
    class Config:
        orm_mode = True
