from typing import Optional, List, Dict, Any
from pydantic import BaseModel, EmailStr, Field, validator
from uuid import UUID
from datetime import datetime

class UserBase(BaseModel):
    username: str
    email: EmailStr
    full_name: str
    
class UserCreate(UserBase):
    password: str
    status: str = "active"
    preferences: Optional[Dict[str, Any]] = None
    metadata: Optional[Dict[str, Any]] = None
    initial_roles: Optional[List[UUID]] = None
    
    @validator('status')
    def validate_status(cls, v):
        valid_status = ["active", "inactive"]
        if v not in valid_status:
            raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v

class UserUpdate(BaseModel):
    full_name: Optional[str] = None
    email: Optional[EmailStr] = None
    status: Optional[str] = None
    password: Optional[str] = None
    preferences: Optional[Dict[str, Any]] = None
    metadata: Optional[Dict[str, Any]] = None
    
    @validator('status')
    def validate_status(cls, v):
        if v is not None:
            valid_status = ["active", "inactive", "suspended", "locked"]
            if v not in valid_status:
                raise ValueError(f"Status inválido. Deve ser um dos seguintes: {', '.join(valid_status)}")
        return v

class UserSummary(BaseModel):
    id: UUID
    username: str
    email: EmailStr
    full_name: str
    status: str
    created_at: datetime
    last_login: Optional[datetime] = None
    
    class Config:
        orm_mode = True

class User(UserSummary):
    organization_id: UUID
    updated_at: Optional[datetime] = None
    preferences: Optional[Dict[str, Any]] = None
    metadata: Optional[Dict[str, Any]] = None
    roles: Optional[List[Dict[str, Any]]] = None
    mfa_methods: Optional[List[Dict[str, Any]]] = None
    federated_identities: Optional[List[Dict[str, Any]]] = None
    
    class Config:
        orm_mode = True

class UserPasswordChange(BaseModel):
    current_password: str
    new_password: str
    confirm_password: str
    
    @validator('confirm_password')
    def passwords_match(cls, v, values, **kwargs):
        if 'new_password' in values and v != values['new_password']:
            raise ValueError('Senhas não conferem')
        return v

class UserRoleAssignment(BaseModel):
    role_id: UUID
    expires_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None

class UserRoleAssignmentResponse(BaseModel):
    id: UUID
    user_id: UUID
    role_id: UUID
    role_name: str
    organization_id: UUID
    created_at: datetime
    expires_at: Optional[datetime] = None
    created_by: Optional[UUID] = None
    metadata: Optional[Dict[str, Any]] = None
    
    class Config:
        orm_mode = True
