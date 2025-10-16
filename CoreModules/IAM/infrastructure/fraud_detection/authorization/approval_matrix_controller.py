#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Controlador para Matriz de Autorização e Aprovação para Alertas Comportamentais

Este módulo contém o controlador que gerencia o fluxo de aprovação de alertas,
expondo APIs para criar e processar solicitações de aprovação.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import logging
import datetime
from typing import Dict, Any, List, Optional

from fastapi import FastAPI, HTTPException, Depends, Query, Path, Body, status
from fastapi.security import OAuth2PasswordBearer
from pydantic import BaseModel, Field

from .approval_matrix import ApprovalMatrix, ApprovalMatrixFactory, ApprovalAction, AlertSeverity, AlertCategory, ApprovalLevel
from .approval_matrix_config import get_approval_matrix_config

# Configuração do logger
logger = logging.getLogger("iam.trustguard.authorization.controller")


# Modelos de dados Pydantic
class AlertData(BaseModel):
    """Dados do alerta comportamental."""
    alert_id: str = Field(..., description="ID único do alerta")
    severity: str = Field(..., description="Severidade do alerta (low, medium, high, critical)")
    category: str = Field(..., description="Categoria do alerta")
    transaction_amount: float = Field(0.0, description="Valor da transação associada")
    risk_score: float = Field(..., description="Score de risco (0.0 a 1.0)")
    region: str = Field(..., description="Código da região")
    customer_segment: Optional[str] = Field(None, description="Segmento do cliente")
    event_timestamp: str = Field(..., description="Timestamp do evento original")
    alert_source: str = Field(..., description="Fonte que gerou o alerta")
    details: Dict[str, Any] = Field({}, description="Detalhes adicionais do alerta")


class UserData(BaseModel):
    """Dados do usuário para aprovação."""
    user_id: str = Field(..., description="ID do usuário")
    name: str = Field(..., description="Nome do usuário")
    approval_level: str = Field(..., description="Nível de aprovação do usuário")
    roles: List[str] = Field([], description="Papéis do usuário")


class ApprovalRequest(BaseModel):
    """Solicitação para criação de aprovação."""
    alert_data: AlertData = Field(..., description="Dados do alerta")


class ApprovalActionRequest(BaseModel):
    """Solicitação para ação de aprovação."""
    user_data: UserData = Field(..., description="Dados do usuário")
    action: str = Field(..., description="Ação de aprovação")
    comments: Optional[str] = Field(None, description="Comentários opcionais")


class ApprovalResponse(BaseModel):
    """Resposta para operação de aprovação."""
    success: bool = Field(..., description="Status da operação")
    request_id: str = Field(..., description="ID da solicitação de aprovação")
    status: Optional[str] = Field(None, description="Status atual da solicitação")
    error: Optional[str] = Field(None, description="Mensagem de erro, se houver")
    details: Optional[Dict[str, Any]] = Field(None, description="Detalhes adicionais")


# Instância única da matriz de aprovação (para simplificação)
# Em produção, isso seria gerenciado com injeção de dependência
approval_matrices = {}


def get_approval_matrix(region: str = None) -> ApprovalMatrix:
    """
    Obtém ou cria uma instância da matriz de aprovação para uma região.
    
    Args:
        region: Código da região
        
    Returns:
        Instância da matriz de aprovação
    """
    if region not in approval_matrices:
        config = get_approval_matrix_config(region)
        approval_matrices[region] = ApprovalMatrixFactory.create(config)
    
    return approval_matrices[region]


# API para controlador de matriz de aprovação
router = FastAPI(
    title="API de Aprovação de Alertas Comportamentais",
    description="API para gerenciamento de aprovações de alertas comportamentais",
    version="1.0.0"
)

# Esquema de autenticação (simplificado)
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")


async def get_current_user(token: str = Depends(oauth2_scheme)) -> UserData:
    """
    Obtém o usuário atual a partir do token.
    
    Args:
        token: Token de acesso
        
    Returns:
        Dados do usuário
    
    Raises:
        HTTPException: Se o token for inválido
    """
    # Implementação simplificada - em produção, validaria token com IAM
    # e obteria informações reais do usuário
    try:
        # Usuário de exemplo para ilustração
        return UserData(
            user_id="user123",
            name="Exemplo Usuario",
            approval_level="l3_supervisor",
            roles=["fraud_analyst", "supervisor"]
        )
    except Exception as e:
        logger.error(f"Erro ao validar token: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token de acesso inválido",
            headers={"WWW-Authenticate": "Bearer"},
        )


@router.post(
    "/v1/approval-requests",
    response_model=ApprovalResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Criar solicitação de aprovação",
    description="Cria uma nova solicitação de aprovação para um alerta comportamental"
)
async def create_approval_request(
    request: ApprovalRequest,
    current_user: UserData = Depends(get_current_user)
) -> ApprovalResponse:
    """
    Cria uma solicitação de aprovação para um alerta comportamental.
    
    Args:
        request: Dados para criação da solicitação
        current_user: Usuário atual
        
    Returns:
        Resposta com status da operação
    """
    try:
        # Obter matriz de aprovação para a região do alerta
        region = request.alert_data.region
        approval_matrix = get_approval_matrix(region)
        
        # Criar solicitação de aprovação
        approval_request = approval_matrix.create_approval_request(
            request.alert_data.dict()
        )
        
        # Converter resposta
        return ApprovalResponse(
            success=True,
            request_id=approval_request["request_id"],
            status=approval_request["status"],
            details={
                "created_at": approval_request["created_at"],
                "required_approval_level": approval_request["required_approval_level"],
                "can_auto_approve": approval_request.get("can_auto_approve", False)
            }
        )
        
    except Exception as e:
        logger.error(f"Erro ao criar solicitação de aprovação: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao criar solicitação de aprovação: {str(e)}"
        )


@router.post(
    "/v1/approval-requests/{request_id}/actions",
    response_model=ApprovalResponse,
    status_code=status.HTTP_200_OK,
    summary="Executar ação de aprovação",
    description="Executa uma ação (aprovar, rejeitar, etc.) em uma solicitação de aprovação"
)
async def process_approval_action(
    request_id: str = Path(..., description="ID da solicitação de aprovação"),
    action_request: ApprovalActionRequest = Body(...),
    region: str = Query(None, description="Região (opcional)"),
    current_user: UserData = Depends(get_current_user)
) -> ApprovalResponse:
    """
    Processa uma ação de aprovação/rejeição para um alerta.
    
    Args:
        request_id: ID da solicitação
        action_request: Dados da ação
        region: Região opcional
        current_user: Usuário atual
        
    Returns:
        Resposta com status da operação
    """
    try:
        # Verificar se usuário tem permissão
        # Em produção, validaria se o usuário atual pode agir como o usuário especificado
        if current_user.user_id != action_request.user_data.user_id:
            logger.warning(f"Tentativa de impersonificação: {current_user.user_id} -> {action_request.user_data.user_id}")
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Não autorizado a agir em nome de outro usuário"
            )
        
        # Obter matriz de aprovação
        approval_matrix = get_approval_matrix(region)
        
        # Converter ação para enum
        try:
            action_enum = ApprovalAction(action_request.action)
        except ValueError:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Ação inválida: {action_request.action}"
            )
        
        # Processar ação
        result = approval_matrix.process_approval_action(
            request_id=request_id,
            user_data=action_request.user_data.dict(),
            action=action_enum,
            comments=action_request.comments or ""
        )
        
        # Verificar resultado
        if not result.get("success", False):
            error_msg = result.get("error", "Erro desconhecido ao processar ação")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=error_msg
            )
        
        # Converter resposta
        return ApprovalResponse(
            success=result["success"],
            request_id=result["request_id"],
            status=result["status"],
            details={k: v for k, v in result.items() if k not in ["success", "request_id", "status", "error"]}
        )
        
    except HTTPException:
        raise
        
    except Exception as e:
        logger.error(f"Erro ao processar ação de aprovação: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao processar ação de aprovação: {str(e)}"
        )


@router.get(
    "/v1/approval-requests/{request_id}",
    response_model=Dict[str, Any],
    status_code=status.HTTP_200_OK,
    summary="Obter solicitação de aprovação",
    description="Obtém detalhes de uma solicitação de aprovação"
)
async def get_approval_request(
    request_id: str = Path(..., description="ID da solicitação de aprovação"),
    region: str = Query(None, description="Região (opcional)"),
    current_user: UserData = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Obtém detalhes de uma solicitação de aprovação.
    
    Args:
        request_id: ID da solicitação
        region: Região opcional
        current_user: Usuário atual
        
    Returns:
        Detalhes da solicitação de aprovação
    """
    try:
        # Obter matriz de aprovação
        approval_matrix = get_approval_matrix(region)
        
        # Obter solicitação
        approval_request = approval_matrix.get_approval_request(request_id)
        
        if not approval_request:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Solicitação de aprovação não encontrada: {request_id}"
            )
        
        return approval_request
        
    except HTTPException:
        raise
        
    except Exception as e:
        logger.error(f"Erro ao obter solicitação de aprovação: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao obter solicitação de aprovação: {str(e)}"
        )


@router.get(
    "/v1/approval-requests",
    response_model=List[Dict[str, Any]],
    status_code=status.HTTP_200_OK,
    summary="Listar solicitações de aprovação pendentes",
    description="Lista solicitações de aprovação pendentes, opcionalmente filtradas por nível"
)
async def list_pending_approvals(
    approval_level: Optional[str] = Query(None, description="Filtrar por nível de aprovação"),
    region: str = Query(None, description="Região (opcional)"),
    limit: int = Query(100, description="Limite de resultados"),
    offset: int = Query(0, description="Offset para paginação"),
    current_user: UserData = Depends(get_current_user)
) -> List[Dict[str, Any]]:
    """
    Lista solicitações de aprovação pendentes.
    
    Args:
        approval_level: Nível de aprovação para filtro
        region: Região opcional
        limit: Limite de resultados
        offset: Offset para paginação
        current_user: Usuário atual
        
    Returns:
        Lista de solicitações pendentes
    """
    try:
        # Obter matriz de aprovação
        approval_matrix = get_approval_matrix(region)
        
        # Converter nível de aprovação
        level_enum = None
        if approval_level:
            try:
                level_enum = ApprovalLevel(approval_level)
            except ValueError:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"Nível de aprovação inválido: {approval_level}"
                )
        
        # Obter solicitações pendentes
        return approval_matrix.get_pending_approvals(
            approval_level=level_enum,
            limit=limit,
            offset=offset
        )
        
    except HTTPException:
        raise
        
    except Exception as e:
        logger.error(f"Erro ao listar solicitações pendentes: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao listar solicitações pendentes: {str(e)}"
        )