from fastapi import APIRouter, Depends, HTTPException, status, Request, Response, Body
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from typing import Dict, Any, Optional, List
from datetime import datetime, timedelta
import uuid

from app.db.session import get_db
from app.schemas.auth import (
    LoginRequest, LoginResponse, MfaInitiateRequest, MfaInitiateResponse,
    MfaVerifyRequest, RefreshTokenRequest, Token,
    ArAuthRequest, ArAuthUpdateRequest, ArAuthResponse
)
from app.services.auth_service import AuthService
from app.services.mfa_service import MFAService
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id

router = APIRouter(prefix="/auth", tags=["Authentication"])

# Configuração do OAuth2
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="auth/login")

# Helper para obter o usuário atual a partir do token
async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """
    Dependency para obter o usuário atual a partir do token JWT
    """
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Credenciais inválidas",
        headers={"WWW-Authenticate": "Bearer"},
    )
    
    try:
        payload = AuthService.decode_token(token)
        user_id = payload.get("user_id")
        if user_id is None:
            raise credentials_exception
        
        # Chamada SQL para obter o usuário pelo ID
        # (Implementação simplificada, na prática usaria uma função específica)
        user = db.execute(
            "SELECT * FROM iam.users WHERE id = :user_id",
            {"user_id": user_id}
        ).fetchone()
        
        if user is None:
            raise credentials_exception
        
        # Verificar se o usuário está ativo
        if user["status"] != "active":
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Usuário inativo ou suspenso",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        return dict(user)
    except Exception as e:
        logger.error(f"Error authenticating user from token: {str(e)}")
        raise credentials_exception

@router.post("/login", response_model=LoginResponse)
async def login(
    login_data: LoginRequest,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """
    Autentica um usuário e retorna tokens de acesso e atualização
    """
    # Autentica o usuário
    user = await AuthService.authenticate_user(
        db=db,
        username=login_data.username,
        password=login_data.password,
        organization_code=login_data.organization_code
    )
    
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Username, senha ou organização inválidos",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    # Verifica se o usuário está ativo
    if user["status"] != "active":
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Usuário inativo ou suspenso",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    # Obtém a organização
    organization = db.execute(
        "SELECT * FROM iam.organizations WHERE id = :org_id",
        {"org_id": str(user["organization_id"])}
    ).fetchone()
    
    if not organization:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao obter dados da organização",
        )
    
    # Verifica se o usuário tem MFA configurado
    user_mfa_methods = await MFAService.list_user_mfa_methods(db, uuid.UUID(user["id"]))
    mfa_required = False
    mfa_token = None
    
    # Se o usuário tem métodos MFA ativos, exige verificação
    if user_mfa_methods and any(m["status"] == "enabled" for m in user_mfa_methods):
        mfa_required = True
        
        # Cria um token MFA para a verificação
        mfa_token = AuthService.create_mfa_token({
            "user_id": str(user["id"]),
            "tenant_id": get_current_tenant_id(),
            "organization_id": str(user["organization_id"]),
            "auth_type": "mfa_required"
        })
        
        # Para evitar vazamento de informações, não gera tokens de acesso ainda
        access_token = mfa_token
        refresh_token = None
        expires_in = settings.MFA_TOKEN_EXPIRE_MINUTES * 60
    else:
        # Sem MFA, gera tokens normalmente
        
        # Obtém os papéis e permissões do usuário
        roles = await AuthService.get_user_roles(db, uuid.UUID(user["id"]))
        permissions = await AuthService.get_user_permissions(db, uuid.UUID(user["id"]))
        
        # Cria tokens
        token_data = {
            "sub": user["username"],
            "user_id": str(user["id"]),
            "tenant_id": get_current_tenant_id(),
            "organization_id": str(user["organization_id"]),
            "roles": [r["code"] for r in roles],
            "permissions": permissions,
            "is_mfa_verified": False,
            "session_id": str(uuid.uuid4())
        }
        
        # Definindo tempo de expiração
        expires_delta = timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
        if login_data.remember_me:
            expires_delta = timedelta(days=7)  # 1 semana para "lembrar de mim"
        
        access_token = AuthService.create_access_token(
            data=token_data,
            expires_delta=expires_delta
        )
        
        # Token de atualização apenas se solicitado (remember_me)
        refresh_token = None
        if login_data.remember_me:
            refresh_token = AuthService.create_refresh_token(token_data)
        
        expires_in = expires_delta.total_seconds()
    
    # Registra o login (para auditoria)
    try:
        db.execute(
            "SELECT iam.log_user_login(:user_id, :tenant_id, :success, :ip_address, :user_agent, :metadata)",
            {
                "user_id": str(user["id"]),
                "tenant_id": get_current_tenant_id(),
                "success": True,
                "ip_address": None,  # Seria obtido da requisição em uma implementação completa
                "user_agent": None,  # Seria obtido da requisição em uma implementação completa
                "metadata": {"mfa_required": mfa_required}
            }
        )
        db.commit()
    except Exception as e:
        logger.error(f"Error logging user login: {str(e)}")
    
    # Prepara a resposta
    response = {
        "access_token": access_token,
        "token_type": "Bearer",
        "expires_in": int(expires_in),
        "refresh_token": refresh_token,
        "user": {
            "id": user["id"],
            "username": user["username"],
            "email": user["email"],
            "full_name": user["full_name"],
            "status": user["status"]
        },
        "organization": {
            "id": organization["id"],
            "code": organization["code"],
            "name": organization["name"]
        },
        "mfa_required": mfa_required,
        "mfa_token": mfa_token
    }
    
    return response
