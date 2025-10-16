from fastapi import APIRouter, Depends, HTTPException, status, Path, Query
from sqlalchemy.orm import Session
from typing import Dict, Any, List, Optional
from uuid import UUID
from datetime import datetime, timedelta
import json

from app.db.session import get_db
from app.schemas.auth import (
    ArAuthRequest, ArAuthUpdateRequest, ArAuthResponse,
    MfaEnrollRequest, MfaMethodResponse
)
from app.services.auth_service import AuthService
from app.services.mfa_service import MFAService
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id
from app.routers.auth import get_current_user

router = APIRouter(prefix="/ar", tags=["AR/VR"])

@router.post("/methods", response_model=MfaMethodResponse, status_code=status.HTTP_201_CREATED)
async def register_ar_method(
    request: MfaEnrollRequest,
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Registra um novo método de autenticação AR/VR para o usuário atual
    """
    # Verificar se o método é tipo AR/VR
    if request.method_type not in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Método inválido para AR/VR. Deve ser um dos seguintes: ar_spatial_gesture, ar_gaze_pattern, ar_spatial_password"
        )
    
    try:
        # Registra o método AR/VR (utiliza o serviço MFA existente com tipo específico)
        ar_method = await MFAService.register_mfa_method(
            db=db,
            user_id=UUID(current_user["id"]),
            method_type=request.method_type,
            name=request.name,
            metadata=request.metadata
        )
        
        # Registra em log para auditoria
        try:
            db.execute(
                """
                SELECT iam.log_audit_event(
                    :tenant_id, 'ar_method_registration', 'ar_auth_method', :method_id,
                    'register', :user_id, :metadata::jsonb
                )
                """,
                {
                    "tenant_id": get_current_tenant_id(),
                    "method_id": ar_method["id"],
                    "user_id": current_user["id"],
                    "metadata": json.dumps({
                        "method_type": request.method_type,
                        "name": request.name
                    })
                }
            )
            db.commit()
        except Exception as e:
            logger.error(f"Error logging AR method registration: {str(e)}")
        
        return ar_method
    
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error registering AR method: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao registrar método AR/VR"
        )

@router.get("/methods", response_model=List[MfaMethodResponse])
async def list_ar_methods(
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> List[Dict[str, Any]]:
    """
    Lista os métodos de autenticação AR/VR do usuário atual
    """
    try:
        # Obtém os métodos MFA do tipo AR/VR
        all_methods = await MFAService.list_user_mfa_methods(
            db=db,
            user_id=UUID(current_user["id"])
        )
        
        # Filtra apenas os métodos AR/VR
        ar_methods = [
            method for method in all_methods
            if method["method_type"] in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]
        ]
        
        return ar_methods
    
    except Exception as e:
        logger.error(f"Error listing AR methods: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao listar métodos AR/VR"
        )

@router.delete("/methods/{method_id}", response_model=Dict[str, Any])
async def remove_ar_method(
    method_id: UUID = Path(..., description="ID do método AR/VR"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Remove um método de autenticação AR/VR do usuário atual
    """
    try:
        # Verifica se o método é do tipo AR/VR
        method = db.execute(
            """
            SELECT * FROM iam.user_mfa_methods
            WHERE id = :method_id AND user_id = :user_id
            """,
            {
                "method_id": str(method_id),
                "user_id": current_user["id"]
            }
        ).fetchone()
        
        if not method:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Método AR/VR não encontrado"
            )
        
        if method["method_type"] not in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="O método especificado não é um método AR/VR"
            )
        
        # Remove o método AR/VR
        success = await MFAService.remove_mfa_method(
            db=db,
            user_id=UUID(current_user["id"]),
            method_id=method_id
        )
        
        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Método AR/VR não encontrado ou não pode ser removido"
            )
        
        # Registra em log para auditoria
        try:
            db.execute(
                """
                SELECT iam.log_audit_event(
                    :tenant_id, 'ar_method_removal', 'ar_auth_method', :method_id,
                    'remove', :user_id, :metadata::jsonb
                )
                """,
                {
                    "tenant_id": get_current_tenant_id(),
                    "method_id": str(method_id),
                    "user_id": current_user["id"],
                    "metadata": json.dumps({
                        "method_type": method["method_type"]
                    })
                }
            )
            db.commit()
        except Exception as e:
            logger.error(f"Error logging AR method removal: {str(e)}")
        
        return {
            "success": True,
            "message": "Método AR/VR removido com sucesso"
        }
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error removing AR method: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao remover método AR/VR"
        )

@router.post("/continuous-auth/start", response_model=ArAuthResponse)
async def start_ar_continuous_auth(
    request: ArAuthRequest,
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Inicia uma sessão de autenticação contínua para AR/VR
    """
    try:
        # Verifica se o usuário tem pelo menos um método AR/VR registrado
        ar_methods = await list_ar_methods(db=db, current_user=current_user)
        
        if not ar_methods:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Usuário não possui métodos AR/VR registrados"
            )
        
        # Cria uma nova sessão de autenticação contínua
        session_id = UUID.uuid4()
        now = datetime.utcnow()
        expires_at = now + timedelta(hours=request.session_duration_hours)
        
        # Insere a sessão no banco de dados
        db.execute(
            """
            INSERT INTO iam.ar_continuous_auth_sessions (
                id, user_id, device_id, confidence_score, 
                last_verification, created_at, expires_at, metadata
            )
            VALUES (
                :session_id, :user_id, :device_id, :confidence_score,
                :last_verification, :created_at, :expires_at, :metadata::jsonb
            )
            """,
            {
                "session_id": str(session_id),
                "user_id": current_user["id"],
                "device_id": request.device_id,
                "confidence_score": request.initial_confidence,
                "last_verification": now,
                "created_at": now,
                "expires_at": expires_at,
                "metadata": json.dumps(request.metadata or {})
            }
        )
        
        # Registra em log para auditoria
        db.execute(
            """
            SELECT iam.log_audit_event(
                :tenant_id, 'ar_continuous_auth_start', 'ar_auth_session', :session_id,
                'start', :user_id, :metadata::jsonb
            )
            """,
            {
                "tenant_id": get_current_tenant_id(),
                "session_id": str(session_id),
                "user_id": current_user["id"],
                "metadata": json.dumps({
                    "device_id": request.device_id,
                    "initial_confidence": request.initial_confidence,
                    "session_duration_hours": request.session_duration_hours
                })
            }
        )
        
        db.commit()
        
        # Prepara a resposta
        response = {
            "session_id": session_id,
            "user_id": UUID(current_user["id"]),
            "device_id": request.device_id,
            "confidence_score": request.initial_confidence,
            "expires_at": expires_at,
            "last_verification": now,
            "created_at": now,
            "metadata": request.metadata
        }
        
        return response
    
    except HTTPException:
        raise
    except Exception as e:
        db.rollback()
        logger.error(f"Error starting AR continuous auth: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao iniciar autenticação contínua AR/VR"
        )

@router.put("/continuous-auth/update", response_model=Dict[str, Any])
async def update_ar_auth_confidence(
    request: ArAuthUpdateRequest,
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Atualiza a confiança de uma sessão de autenticação contínua AR/VR
    """
    try:
        # Verifica se a sessão existe e pertence ao usuário
        session = db.execute(
            """
            SELECT * FROM iam.ar_continuous_auth_sessions
            WHERE id = :session_id AND user_id = :user_id
            """,
            {
                "session_id": str(request.session_id),
                "user_id": current_user["id"]
            }
        ).fetchone()
        
        if not session:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Sessão de autenticação AR/VR não encontrada"
            )
        
        # Verifica se a sessão não expirou
        now = datetime.utcnow()
        if session["expires_at"] < now:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Sessão de autenticação AR/VR expirada"
            )
        
        # Calcula a nova confiança
        current_confidence = session["confidence_score"]
        new_confidence = current_confidence + request.confidence_update
        
        # Limita entre 0 e 1
        new_confidence = max(0.0, min(1.0, new_confidence))
        
        # Atualiza a sessão
        db.execute(
            """
            UPDATE iam.ar_continuous_auth_sessions
            SET confidence_score = :confidence_score,
                last_verification = NOW()
            WHERE id = :session_id
            """,
            {
                "session_id": str(request.session_id),
                "confidence_score": new_confidence
            }
        )
        
        # Se a confiança cair abaixo do limiar, registra em eventos de segurança
        if new_confidence < settings.AR_VR_CONTINUOUS_AUTH_THRESHOLD:
            db.execute(
                """
                INSERT INTO iam.security_events (
                    user_id, event_type, severity, details
                )
                VALUES (
                    :user_id, 'ar_auth_confidence_low', 'high',
                    :details::jsonb
                )
                """,
                {
                    "user_id": current_user["id"],
                    "details": json.dumps({
                        "session_id": str(request.session_id),
                        "confidence_score": new_confidence,
                        "threshold": settings.AR_VR_CONTINUOUS_AUTH_THRESHOLD,
                        "reason": request.reason
                    })
                }
            )
        
        # Registra em log para auditoria
        db.execute(
            """
            SELECT iam.log_audit_event(
                :tenant_id, 'ar_continuous_auth_update', 'ar_auth_session', :session_id,
                'update', :user_id, :metadata::jsonb
            )
            """,
            {
                "tenant_id": get_current_tenant_id(),
                "session_id": str(request.session_id),
                "user_id": current_user["id"],
                "metadata": json.dumps({
                    "previous_confidence": current_confidence,
                    "new_confidence": new_confidence,
                    "confidence_update": request.confidence_update,
                    "reason": request.reason
                })
            }
        )
        
        db.commit()
        
        # Prepara a resposta
        response = {
            "session_id": request.session_id,
            "new_confidence": new_confidence,
            "previous_confidence": current_confidence,
            "update_time": now.isoformat()
        }
        
        return response
    
    except HTTPException:
        raise
    except Exception as e:
        db.rollback()
        logger.error(f"Error updating AR auth confidence: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao atualizar confiança de autenticação AR/VR"
        )

@router.get("/continuous-auth/session/{session_id}", response_model=ArAuthResponse)
async def get_ar_auth_session(
    session_id: UUID = Path(..., description="ID da sessão de autenticação AR/VR"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Obtém os detalhes de uma sessão de autenticação contínua AR/VR
    """
    try:
        # Verifica se a sessão existe e pertence ao usuário
        session = db.execute(
            """
            SELECT * FROM iam.ar_continuous_auth_sessions
            WHERE id = :session_id AND user_id = :user_id
            """,
            {
                "session_id": str(session_id),
                "user_id": current_user["id"]
            }
        ).fetchone()
        
        if not session:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Sessão de autenticação AR/VR não encontrada"
            )
        
        # Converte o resultado em dicionário
        session_dict = dict(session)
        
        # Converte campos UUID para objetos UUID
        session_dict["id"] = UUID(session_dict["id"])
        session_dict["user_id"] = UUID(session_dict["user_id"])
        
        # Converte campos datetime para ISO format
        for field in ["last_verification", "created_at", "expires_at"]:
            if session_dict[field]:
                session_dict[field] = session_dict[field].isoformat()
        
        return session_dict
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error getting AR auth session: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao obter sessão de autenticação AR/VR"
        )
