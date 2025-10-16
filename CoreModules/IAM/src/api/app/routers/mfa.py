from fastapi import APIRouter, Depends, HTTPException, status, Path, Query
from sqlalchemy.orm import Session
from typing import Dict, Any, List, Optional
from uuid import UUID

from app.db.session import get_db
from app.schemas.auth import (
    MfaInitiateRequest, MfaInitiateResponse, MfaVerifyRequest,
    MfaEnrollRequest, MfaMethodResponse
)
from app.services.auth_service import AuthService
from app.services.mfa_service import MFAService
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id
from app.routers.auth import get_current_user

router = APIRouter(prefix="/mfa", tags=["MFA"])

@router.post("/initiate", response_model=MfaInitiateResponse)
async def initiate_mfa(
    request: MfaInitiateRequest,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """
    Inicia o processo de autenticação multi-fator
    """
    try:
        # Decodifica o token MFA para obter o usuário
        payload = AuthService.decode_token(request.auth_token)
        
        if payload.get("token_type") != "mfa":
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Token inválido para iniciar processo MFA",
            )
        
        user_id = payload.get("user_id")
        if not user_id:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Token não contém ID de usuário",
            )
        
        # Obtém o usuário
        user = db.execute(
            "SELECT * FROM iam.users WHERE id = :user_id",
            {"user_id": user_id}
        ).fetchone()
        
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Usuário não encontrado",
            )
        
        # Obtém os métodos MFA do usuário
        mfa_methods = await MFAService.list_user_mfa_methods(db, UUID(user_id))
        
        # Verifica se o método solicitado está disponível
        method = next((m for m in mfa_methods if m["method_type"] == request.method_type and m["status"] == "enabled"), None)
        
        if not method:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Método MFA '{request.method_type}' não está disponível ou habilitado",
            )
        
        # Preparar desafio de acordo com o tipo de método
        challenge = {}
        verification_url = None
        
        if request.method_type == "totp":
            # Para TOTP, não é necessário gerar um desafio, usuário já tem o aplicativo
            challenge = {
                "message": "Digite o código de 6 dígitos do seu aplicativo autenticador"
            }
        
        elif request.method_type == "backup_codes":
            challenge = {
                "message": "Digite um dos seus códigos de backup"
            }
        
        elif request.method_type == "sms":
            # Em uma implementação real, geraria e enviaria um código SMS
            # Para exemplo, simularemos o envio
            code = "123456"  # Na prática, seria gerado aleatoriamente
            phone = method["method_data"]["phone_number"]
            
            # Simular armazenamento do código (em produção, seria armazenado no banco de dados)
            # E envio do SMS (em produção, usaria um provedor SMS real)
            logger.info(f"[MOCK] Enviando código SMS para {phone}: {code}")
            
            # Armazenar o código na tabela de verificação
            db.execute(
                """
                INSERT INTO iam.mfa_verification_codes (user_id, method_id, code, expires_at)
                VALUES (:user_id, :method_id, :code, NOW() + interval '5 minutes')
                """,
                {
                    "user_id": user_id,
                    "method_id": method["id"],
                    "code": code
                }
            )
            db.commit()
            
            challenge = {
                "message": f"Um código de verificação foi enviado para o número terminado em {phone[-4:]}",
                "phone_masked": f"xxxx-xxx-{phone[-4:]}"
            }
        
        elif request.method_type == "email":
            # Em uma implementação real, geraria e enviaria um código por email
            # Para exemplo, simularemos o envio
            code = "123456"  # Na prática, seria gerado aleatoriamente
            email = method["method_data"]["email"]
            
            # Simular armazenamento do código e envio do email
            logger.info(f"[MOCK] Enviando código Email para {email}: {code}")
            
            # Armazenar o código na tabela de verificação
            db.execute(
                """
                INSERT INTO iam.mfa_verification_codes (user_id, method_id, code, expires_at)
                VALUES (:user_id, :method_id, :code, NOW() + interval '5 minutes')
                """,
                {
                    "user_id": user_id,
                    "method_id": method["id"],
                    "code": code
                }
            )
            db.commit()
            
            challenge = {
                "message": f"Um código de verificação foi enviado para {email[:3]}...{email.split('@')[0][-2:]}@{email.split('@')[1]}",
                "email_masked": f"{email[:3]}...{email.split('@')[0][-2:]}@{email.split('@')[1]}"
            }
        
        elif request.method_type in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]:
            # Para métodos AR/VR, o desafio seria específico para o dispositivo
            challenge = {
                "message": f"Realize o seu padrão de {request.method_type.replace('ar_', '').replace('_', ' ')} no dispositivo AR",
                "ar_method_type": request.method_type,
                "session_id": str(UUID(user_id))  # Na prática, seria um ID de sessão único
            }
            
            # URL para iniciar o processo no dispositivo AR/VR
            verification_url = f"ar-auth://{request.method_type}/verify?session={str(UUID(user_id))}"
        
        # Cria um novo token MFA com informações do método
        mfa_token = AuthService.create_mfa_token({
            "user_id": user_id,
            "tenant_id": get_current_tenant_id(),
            "organization_id": str(user["organization_id"]),
            "method_id": method["id"],
            "method_type": request.method_type
        })
        
        # Prepara a resposta
        response = {
            "mfa_token": mfa_token,
            "method_type": request.method_type,
            "expires_in": settings.MFA_TOKEN_EXPIRE_MINUTES * 60,
            "challenge": challenge,
            "verification_url": verification_url
        }
        
        return response
    
    except Exception as e:
        logger.error(f"Error initiating MFA: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao iniciar processo MFA"
        )

@router.post("/verify", response_model=Dict[str, Any])
async def verify_mfa(
    request: MfaVerifyRequest,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """
    Verifica um código MFA e retorna tokens de acesso em caso de sucesso
    """
    try:
        # Decodifica o token MFA
        payload = AuthService.decode_token(request.mfa_token)
        
        if payload.get("token_type") != "mfa":
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Token inválido para verificação MFA",
            )
        
        user_id = payload.get("user_id")
        method_id = payload.get("method_id")
        method_type = payload.get("method_type")
        
        if not user_id or not method_id or not method_type:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Token MFA inválido ou incompleto",
            )
        
        # Verifica o código MFA
        is_valid = await MFAService.verify_mfa(
            db=db,
            user_id=UUID(user_id),
            method_id=UUID(method_id),
            verification_code=request.verification_code
        )
        
        if not is_valid:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Código de verificação inválido ou expirado",
            )
        
        # Código válido, gerar tokens de acesso
        
        # Obtém o usuário
        user = db.execute(
            "SELECT * FROM iam.users WHERE id = :user_id",
            {"user_id": user_id}
        ).fetchone()
        
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Usuário não encontrado",
            )
        
        # Obtém a organização
        organization = db.execute(
            "SELECT * FROM iam.organizations WHERE id = :org_id",
            {"org_id": str(user["organization_id"])}
        ).fetchone()
        
        # Obtém os papéis e permissões do usuário
        roles = await AuthService.get_user_roles(db, UUID(user_id))
        permissions = await AuthService.get_user_permissions(db, UUID(user_id))
        
        # Cria tokens
        session_id = str(UUID.uuid4())
        token_data = {
            "sub": user["username"],
            "user_id": user_id,
            "tenant_id": get_current_tenant_id(),
            "organization_id": str(user["organization_id"]),
            "roles": [r["code"] for r in roles],
            "permissions": permissions,
            "is_mfa_verified": True,
            "mfa_method": method_type,
            "session_id": session_id
        }
        
        access_token = AuthService.create_access_token(
            data=token_data,
            expires_delta=timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
        )
        
        refresh_token = AuthService.create_refresh_token(token_data)
        
        # Se solicitado, registra o dispositivo como confiável
        if request.trust_device:
            # Em uma implementação real, armazenaria um token de dispositivo
            # confiável associado ao usuário
            pass
        
        # Registra a verificação MFA bem-sucedida
        try:
            db.execute(
                "UPDATE iam.user_mfa_methods SET last_used = NOW() WHERE id = :method_id",
                {"method_id": method_id}
            )
            
            db.execute(
                """
                INSERT INTO iam.user_sessions (id, user_id, session_type, started_at, expires_at, is_mfa_verified, mfa_method_id)
                VALUES (:session_id, :user_id, 'web', NOW(), NOW() + interval '1 day', TRUE, :method_id)
                """,
                {
                    "session_id": session_id,
                    "user_id": user_id,
                    "method_id": method_id
                }
            )
            
            db.commit()
        except Exception as e:
            logger.error(f"Error registering MFA verification: {str(e)}")
        
        # Prepara a resposta
        response = {
            "access_token": access_token,
            "token_type": "Bearer",
            "expires_in": settings.ACCESS_TOKEN_EXPIRE_MINUTES * 60,
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
            "session_id": session_id,
            "mfa_verified": True,
            "mfa_method": method_type
        }
        
        return response
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error verifying MFA: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao verificar código MFA"
        )

@router.get("/methods", response_model=List[MfaMethodResponse])
async def list_mfa_methods(
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> List[Dict[str, Any]]:
    """
    Lista os métodos MFA do usuário atual
    """
    try:
        # Obtém os métodos MFA do usuário
        mfa_methods = await MFAService.list_user_mfa_methods(
            db=db,
            user_id=UUID(current_user["id"])
        )
        
        return mfa_methods
    
    except Exception as e:
        logger.error(f"Error listing MFA methods: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao listar métodos MFA"
        )

@router.post("/methods", response_model=MfaMethodResponse, status_code=status.HTTP_201_CREATED)
async def register_mfa_method(
    request: MfaEnrollRequest,
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Registra um novo método MFA para o usuário atual
    """
    try:
        # Registra o método MFA
        mfa_method = await MFAService.register_mfa_method(
            db=db,
            user_id=UUID(current_user["id"]),
            method_type=request.method_type,
            name=request.name,
            metadata={
                "phone_number": request.phone_number,
                "email": request.email,
                **(request.metadata or {})
            }
        )
        
        return mfa_method
    
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error registering MFA method: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao registrar método MFA"
        )

@router.delete("/methods/{method_id}", response_model=Dict[str, Any])
async def remove_mfa_method(
    method_id: UUID = Path(..., description="ID do método MFA"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Remove um método MFA do usuário atual
    """
    try:
        # Remove o método MFA
        success = await MFAService.remove_mfa_method(
            db=db,
            user_id=UUID(current_user["id"]),
            method_id=method_id
        )
        
        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Método MFA não encontrado ou não pode ser removido"
            )
        
        return {
            "success": True,
            "message": "Método MFA removido com sucesso"
        }
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error removing MFA method: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao remover método MFA"
        )
