"""
INNOVABIZ IAM - Router de Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Endpoints para gerenciamento e consulta de eventos de auditoria
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA
"""

from typing import List, Optional, Dict, Any
from datetime import datetime, timedelta
from fastapi import APIRouter, Depends, HTTPException, Query, Path, status, Request
from fastapi.responses import JSONResponse
from fastapi.security import OAuth2PasswordBearer
from pydantic import BaseModel, Field

from ..models.audit import (
    AuditEventRead, 
    AuditEventCreate,
    AuditEventCategory,
    AuditEventSeverity,
    AuditEventFilter
)
from ..services.audit_service import AuditService, get_audit_service
from ..core.observability import logger
from ..core.context import get_regional_context, get_tenant_context, RegionalContext
from ..core.audit_context_integrator import get_audit_context_integrator

# Configuração do router
router = APIRouter()

# OAuth2 bearer token para autenticação
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Logger para este módulo
audit_logger = logger.bind(module="audit_router")

# Modelos para respostas paginadas
class PaginatedResponse(BaseModel):
    """Modelo para respostas paginadas."""
    items: List[AuditEventRead]
    total: int
    page: int
    page_size: int
    total_pages: int


# Endpoints para auditoria

@router.get(
    "", 
    response_model=PaginatedResponse,
    summary="Listar eventos de auditoria",
    description="Obtém eventos de auditoria com filtros, ordenação e paginação"
)
async def list_audit_events(
    request: Request,
    category: Optional[AuditEventCategory] = Query(None, description="Filtrar por categoria de evento"),
    severity: Optional[AuditEventSeverity] = Query(None, description="Filtrar por severidade do evento"),
    start_date: Optional[datetime] = Query(None, description="Data inicial para filtro (formato ISO)"),
    end_date: Optional[datetime] = Query(None, description="Data final para filtro (formato ISO)"),
    user_id: Optional[str] = Query(None, description="Filtrar por ID do usuário"),
    resource_id: Optional[str] = Query(None, description="Filtrar por ID do recurso"),
    action: Optional[str] = Query(None, description="Filtrar por ação realizada"),
    status_code: Optional[int] = Query(None, description="Filtrar por código de status HTTP"),
    sort_by: str = Query("timestamp", description="Campo para ordenação"),
    sort_order: str = Query("desc", description="Ordem da ordenação (asc, desc)"),
    page: int = Query(1, ge=1, description="Número da página"),
    page_size: int = Query(20, ge=1, le=100, description="Itens por página"),
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context),
    audit_service: AuditService = Depends(get_audit_service),
    token: str = Depends(oauth2_scheme)
):
    """
    Lista eventos de auditoria com suporte a filtros avançados, ordenação e paginação.
    
    Os eventos são filtrados automaticamente pelo contexto regional e tenant_id do usuário.
    Compliance com GDPR/LGPD é aplicado automaticamente (mascaramento de dados sensíveis).
    """
    try:
        # Cria filtro a partir dos parâmetros
        filters = AuditEventFilter(
            category=category,
            severity=severity,
            start_date=start_date,
            end_date=end_date,
            user_id=user_id,
            resource_id=resource_id,
            action=action,
            status_code=status_code,
            regional_context=regional_context.value,
            tenant_id=tenant_id,
        )
        
        # Obtém eventos do serviço de auditoria
        events, total = await audit_service.find_events(
            filters=filters,
            sort_by=sort_by,
            sort_order=sort_order,
            page=page,
            page_size=page_size
        )
        
        # Calcula total de páginas
        total_pages = (total + page_size - 1) // page_size
        
        # Registra a consulta no log
        audit_logger.info(
            "Consulta de eventos de auditoria",
            filters=filters.dict(),
            total_results=total,
            page=page,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        # Retorna resposta paginada
        return PaginatedResponse(
            items=events,
            total=total,
            page=page,
            page_size=page_size,
            total_pages=total_pages
        )
        
    except Exception as e:
        audit_logger.error(
            f"Erro ao listar eventos de auditoria: {str(e)}",
            error=str(e),
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao listar eventos de auditoria: {str(e)}"
        )


@router.get(
    "/{event_id}", 
    response_model=AuditEventRead,
    summary="Obter evento de auditoria por ID",
    description="Obtém detalhes de um evento de auditoria específico"
)
async def get_audit_event(
    request: Request,
    event_id: str = Path(..., description="ID do evento de auditoria"),
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context),
    audit_service: AuditService = Depends(get_audit_service),
    token: str = Depends(oauth2_scheme)
):
    """
    Obtém detalhes de um evento de auditoria específico.
    
    Verifica automaticamente se o evento pertence ao tenant e contexto regional do usuário.
    """
    try:
        # Busca evento pelo ID
        event = await audit_service.get_event_by_id(
            event_id=event_id,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        # Verifica se o evento existe e pertence ao tenant correto
        if not event:
            audit_logger.warning(
                "Tentativa de acesso a evento não encontrado",
                event_id=event_id,
                regional_context=regional_context.value,
                tenant_id=tenant_id
            )
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Evento de auditoria não encontrado"
            )
        
        # Log de acesso ao evento
        audit_logger.info(
            "Acesso a evento de auditoria",
            event_id=event_id,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        return event
        
    except HTTPException:
        # Repassa exceções HTTP já formatadas
        raise
        
    except Exception as e:
        audit_logger.error(
            f"Erro ao buscar evento de auditoria: {str(e)}",
            error=str(e),
            event_id=event_id,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao buscar evento de auditoria: {str(e)}"
        )


@router.post(
    "/manual", 
    response_model=AuditEventRead,
    status_code=status.HTTP_201_CREATED,
    summary="Criar evento de auditoria manual",
    description="Cria um novo evento de auditoria manualmente"
)
async def create_audit_event(
    request: Request,
    event: AuditEventCreate,
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context),
    audit_service: AuditService = Depends(get_audit_service),
    audit_integrator = Depends(get_audit_context_integrator),
    token: str = Depends(oauth2_scheme)
):
    """
    Cria um novo evento de auditoria manualmente.
    
    Útil para registrar eventos que não são capturados automaticamente,
    como ações realizadas fora do sistema ou em processos batch.
    """
    try:
        # Enriquece o evento com contexto
        enriched_event = await audit_integrator.enrich_audit_event(
            event=event,
            regional_context=regional_context,
            tenant_id=tenant_id
        )
        
        # Cria o evento no serviço de auditoria
        created_event = await audit_service.create_event(enriched_event)
        
        # Log da criação do evento
        audit_logger.info(
            "Evento de auditoria criado manualmente",
            event_id=created_event.id,
            category=created_event.category,
            severity=created_event.severity,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        return created_event
        
    except Exception as e:
        audit_logger.error(
            f"Erro ao criar evento de auditoria manual: {str(e)}",
            error=str(e),
            category=getattr(event, 'category', None),
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao criar evento de auditoria: {str(e)}"
        )


@router.get(
    "/stats/summary",
    summary="Estatísticas de auditoria",
    description="Obtém estatísticas sobre os eventos de auditoria"
)
async def get_audit_stats(
    request: Request,
    days: int = Query(30, ge=1, le=365, description="Período em dias para as estatísticas"),
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context),
    audit_service: AuditService = Depends(get_audit_service),
    token: str = Depends(oauth2_scheme)
):
    """
    Obtém estatísticas sobre os eventos de auditoria.
    
    Inclui contagens por categoria e severidade, tendências e métricas de compliance.
    """
    try:
        # Define data de início para o período solicitado
        start_date = datetime.utcnow() - timedelta(days=days)
        
        # Obtém as estatísticas
        stats = await audit_service.get_audit_statistics(
            regional_context=regional_context.value,
            tenant_id=tenant_id,
            start_date=start_date
        )
        
        # Log da consulta de estatísticas
        audit_logger.info(
            "Consulta de estatísticas de auditoria",
            days=days,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        return stats
        
    except Exception as e:
        audit_logger.error(
            f"Erro ao obter estatísticas de auditoria: {str(e)}",
            error=str(e),
            days=days,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao obter estatísticas de auditoria: {str(e)}"
        )


@router.get(
    "/compliance/report",
    summary="Relatório de compliance",
    description="Gera um relatório de compliance baseado nos eventos de auditoria"
)
async def get_compliance_report(
    request: Request,
    start_date: Optional[datetime] = Query(None, description="Data inicial para o relatório"),
    end_date: Optional[datetime] = Query(None, description="Data final para o relatório"),
    framework: str = Query(None, description="Framework de compliance (ex: GDPR, LGPD, PCI-DSS)"),
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context),
    audit_service: AuditService = Depends(get_audit_service),
    token: str = Depends(oauth2_scheme)
):
    """
    Gera um relatório de compliance baseado nos eventos de auditoria.
    
    Analisa os eventos de auditoria e gera um relatório de conformidade
    com os frameworks especificados.
    """
    try:
        # Define datas padrão se não informadas
        if not start_date:
            start_date = datetime.utcnow() - timedelta(days=30)
        if not end_date:
            end_date = datetime.utcnow()
            
        # Se não informado framework, usa o padrão para o contexto regional
        if not framework:
            framework = {
                RegionalContext.BR: "LGPD",
                RegionalContext.EU: "GDPR",
                RegionalContext.US: "PCI-DSS",
                RegionalContext.AO: "BNA"
            }.get(regional_context, "GDPR")
        
        # Gera o relatório de compliance
        report = await audit_service.generate_compliance_report(
            regional_context=regional_context.value,
            tenant_id=tenant_id,
            start_date=start_date,
            end_date=end_date,
            framework=framework
        )
        
        # Log da geração do relatório
        audit_logger.info(
            "Relatório de compliance gerado",
            framework=framework,
            start_date=start_date.isoformat(),
            end_date=end_date.isoformat(),
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        
        return report
        
    except Exception as e:
        audit_logger.error(
            f"Erro ao gerar relatório de compliance: {str(e)}",
            error=str(e),
            framework=framework,
            regional_context=regional_context.value,
            tenant_id=tenant_id
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao gerar relatório de compliance: {str(e)}"
        )