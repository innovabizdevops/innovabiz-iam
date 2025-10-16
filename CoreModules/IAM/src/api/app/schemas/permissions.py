from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field, validator
from uuid import UUID
from datetime import datetime

class PermissionBase(BaseModel):
    code: str
    name: str
    permission_scope: str
    resource_type: str
    description: Optional[str] = None
    
    @validator('permission_scope')
    def validate_permission_scope(cls, v):
        valid_scopes = [
            "organization", "application", "module", 
            "feature", "resource", "action"
        ]
        if v not in valid_scopes:
            raise ValueError(f"Escopo de permissão inválido. Deve ser um dos seguintes: {', '.join(valid_scopes)}")
        return v

class PermissionCreate(PermissionBase):
    metadata: Optional[Dict[str, Any]] = None
    is_system_permission: bool = False

class PermissionUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    resource_type: Optional[str] = None
    is_active: Optional[bool] = None
    metadata: Optional[Dict[str, Any]] = None

class PermissionSummary(BaseModel):
    id: UUID
    code: str
    name: str
    permission_scope: str
    resource_type: str
    
    class Config:
        orm_mode = True

class Permission(PermissionSummary):
    organization_id: UUID
    description: Optional[str] = None
    is_system_permission: bool
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    
    class Config:
        orm_mode = True

class AccessPolicy(BaseModel):
    id: UUID
    name: str
    description: Optional[str] = None
    policy_type: str
    organization_id: UUID
    created_at: datetime
    updated_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    policy_data: Dict[str, Any]
    is_active: bool
    
    @validator('policy_type')
    def validate_policy_type(cls, v):
        valid_types = ["rbac", "abac", "hybrid"]
        if v not in valid_types:
            raise ValueError(f"Tipo de política inválido. Deve ser um dos seguintes: {', '.join(valid_types)}")
        return v
    
    class Config:
        orm_mode = True

class AccessPolicyCreate(BaseModel):
    name: str
    description: Optional[str] = None
    policy_type: str
    policy_data: Dict[str, Any]
    is_active: bool = True
    metadata: Optional[Dict[str, Any]] = None
    
    @validator('policy_type')
    def validate_policy_type(cls, v):
        valid_types = ["rbac", "abac", "hybrid"]
        if v not in valid_types:
            raise ValueError(f"Tipo de política inválido. Deve ser um dos seguintes: {', '.join(valid_types)}")
        return v

class AccessPolicyUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    policy_data: Optional[Dict[str, Any]] = None
    is_active: Optional[bool] = None
    metadata: Optional[Dict[str, Any]] = None

class PermissionCheck(BaseModel):
    resource_type: str
    resource_id: Optional[str] = None
    action: str
    context: Optional[Dict[str, Any]] = None
