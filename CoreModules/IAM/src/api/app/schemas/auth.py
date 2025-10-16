from typing import Optional, List, Dict, Any
from pydantic import BaseModel, EmailStr, Field, validator
from uuid import UUID
from datetime import datetime

class Token(BaseModel):
    access_token: str
    token_type: str
    expires_in: int
    refresh_token: Optional[str] = None

class TokenPayload(BaseModel):
    sub: Optional[str] = None
    user_id: UUID
    tenant_id: UUID
    organization_id: UUID
    exp: Optional[int] = None
    roles: List[str] = []
    permissions: List[str] = []
    is_mfa_verified: bool = False
    session_id: Optional[UUID] = None

class LoginRequest(BaseModel):
    username: str
    password: str
    organization_code: str
    remember_me: bool = False

class LoginResponse(BaseModel):
    access_token: str
    token_type: str = "Bearer"
    expires_in: int
    refresh_token: Optional[str] = None
    user: Dict[str, Any]
    organization: Dict[str, Any]
    mfa_required: bool = False
    mfa_token: Optional[str] = None

class MfaInitiateRequest(BaseModel):
    auth_token: str
    method_type: str = Field(..., description="Tipo de método MFA: totp, sms, email, push_notification, ar_spatial_gesture, ar_gaze_pattern, ar_spatial_password, backup_codes")
    
    @validator('method_type')
    def validate_method_type(cls, v):
        valid_methods = [
            "totp", "sms", "email", "push_notification", 
            "ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password", 
            "backup_codes"
        ]
        if v not in valid_methods:
            raise ValueError(f"Método MFA inválido. Deve ser um dos seguintes: {', '.join(valid_methods)}")
        return v

class MfaInitiateResponse(BaseModel):
    mfa_token: str
    method_type: str
    expires_in: int
    challenge: Optional[Dict[str, Any]] = None
    verification_url: Optional[str] = None

class MfaVerifyRequest(BaseModel):
    mfa_token: str
    verification_code: str
    trust_device: bool = False

class MfaEnrollRequest(BaseModel):
    method_type: str
    name: str
    phone_number: Optional[str] = None
    email: Optional[EmailStr] = None
    metadata: Optional[Dict[str, Any]] = None

    @validator('method_type')
    def validate_method_type(cls, v):
        valid_methods = [
            "totp", "sms", "email", "push_notification", "biometric", 
            "security_key", "backup_codes", "ar_spatial_gesture", 
            "ar_gaze_pattern", "ar_spatial_password"
        ]
        if v not in valid_methods:
            raise ValueError(f"Método MFA inválido. Deve ser um dos seguintes: {', '.join(valid_methods)}")
        return v
    
    @validator('phone_number')
    def validate_phone_number(cls, v, values):
        if values.get('method_type') == 'sms' and not v:
            raise ValueError("Número de telefone obrigatório para método SMS")
        return v
    
    @validator('email')
    def validate_email(cls, v, values):
        if values.get('method_type') == 'email' and not v:
            raise ValueError("Email obrigatório para método Email")
        return v

class MfaMethodResponse(BaseModel):
    id: UUID
    user_id: UUID
    method_type: str
    name: str
    status: str
    last_used: Optional[datetime] = None
    created_at: datetime
    updated_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    setup_details: Optional[Dict[str, Any]] = None

class RefreshTokenRequest(BaseModel):
    refresh_token: str

class ArAuthRequest(BaseModel):
    device_id: str
    initial_confidence: float = Field(1.0, ge=0.0, le=1.0)
    session_duration_hours: int = Field(4, ge=1, le=24)
    metadata: Optional[Dict[str, Any]] = None

class ArAuthUpdateRequest(BaseModel):
    session_id: UUID
    confidence_update: float = Field(..., ge=-1.0, le=1.0)
    reason: Optional[str] = None

class ArAuthResponse(BaseModel):
    session_id: UUID
    user_id: UUID
    device_id: str
    confidence_score: float
    expires_at: datetime
    last_verification: datetime
    created_at: datetime
    metadata: Optional[Dict[str, Any]] = None
