from datetime import datetime
from typing import Dict, List, Optional, Any
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.ext.asyncio import AsyncSession

from ..dependencies.database import get_db
from ..dependencies.context import get_tenant_context, get_regional_context
from ..schemas.audit_schemas import (
    AuditEventCreate,
    AuditEventResponse,
    AuditEventListResponse,
    AuditRetentionPolicyCreate,
    AuditRetentionPolicyResponse,
    AuditRetentionPolicyUpdate,
    AuditComplianceReportResponse,
    AuditStatisticsResponse,
    BatchEventCreate
)
from ..services.audit_service import AuditService
from ..models.audit_models import ComplianceFramework, ReportStatus, AuditEventCategory, AuditEventSeverity

router = APIRouter(
    prefix="/audit",
    tags=["Auditoria"],
    responses={404: {"description": "Recurso não encontrado"}}
)


@router.post("/events", response_model=AuditEventResponse, status_code=status.HTTP_201_CREATED)
async def create_audit_event(
    event: AuditEventCreate,
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Cria um novo evento de auditoria.
    
    Este endpoint permite registrar eventos de auditoria no sistema com 
    suporte a multi-contexto (multi-tenant e multi-regional).
    """
    try:
        audit_service = AuditService(db)
        
        # Assegura que o tenant_id e regional_context do middleware sejam usados
        # se não forem explicitamente fornecidos no evento
        if not event.tenant_id:
            event.tenant_id = tenant_id
        if not event.regional_context:
            event.regional_context = regional_context
            
        event_id = await audit_service.create_event(event)
        
        return {
            "id": event_id,
            "tenant_id": event.tenant_id,
            "regional_context": event.regional_context,
            "category": event.category,
            "action": event.action,
            "success": event.success,
            "created_at": datetime.now()
        }
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao criar evento de auditoria: {str(e)}"
        )


@router.post("/events/batch", status_code=status.HTTP_201_CREATED)
async def create_audit_events_batch(
    batch: BatchEventCreate,
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Cria múltiplos eventos de auditoria em lote.
    
    Este endpoint permite registrar vários eventos de auditoria em uma única chamada,
    melhorando a performance e reduzindo o overhead de rede.
    """
    try:
        audit_service = AuditService(db)
        
        # Assegura que eventos sem tenant_id ou regional_context usem os valores do middleware
        for event in batch.events:
            if not event.tenant_id:
                event.tenant_id = tenant_id
            if not event.regional_context:
                event.regional_context = regional_context
        
        result = await audit_service.process_batch_events(batch.events)
        
        return result
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao processar lote de eventos: {str(e)}"
        )


@router.get("/events/{event_id}", response_model=AuditEventResponse)
async def get_audit_event(
    event_id: UUID,
    tenant_id: str = Depends(get_tenant_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Recupera um evento de auditoria pelo ID.
    
    Este endpoint obtém os detalhes de um evento de auditoria específico,
    com isolamento multi-tenant para segurança.
    """
    try:
        audit_service = AuditService(db)
        event = await audit_service.get_event_by_id(event_id, tenant_id)
        
        if not event:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Evento de auditoria {event_id} não encontrado"
            )
            
        return event
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao recuperar evento de auditoria: {str(e)}"
        )


@router.get("/events", response_model=AuditEventListResponse)
async def list_audit_events(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: Optional[str] = Depends(get_regional_context),
    category: Optional[AuditEventCategory] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    resource_id: Optional[str] = None,
    resource_type: Optional[str] = None,
    user_id: Optional[str] = None,
    success: Optional[bool] = None,
    severity: Optional[AuditEventSeverity] = None,
    correlation_id: Optional[str] = None,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: AsyncSession = Depends(get_db)
):
    """
    Lista eventos de auditoria com diversos filtros.
    
    Este endpoint permite consultar eventos de auditoria com múltiplos
    critérios de filtragem e paginação.
    """
    try:
        audit_service = AuditService(db)
        
        events, total_count = await audit_service.get_events(
            tenant_id=tenant_id,
            regional_context=regional_context,
            category=category,
            start_date=start_date,
            end_date=end_date,
            resource_id=resource_id,
            resource_type=resource_type,
            user_id=user_id,
            success=success,
            severity=severity,
            correlation_id=correlation_id,
            limit=limit,
            offset=offset
        )
        
        return {
            "items": events,
            "total": total_count,
            "limit": limit,
            "offset": offset
        }
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao listar eventos de auditoria: {str(e)}"
        )@router.post("/retention-policies", response_model=AuditRetentionPolicyResponse, status_code=status.HTTP_201_CREATED)
async def create_retention_policy(
    policy: AuditRetentionPolicyCreate,
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Cria uma nova política de retenção para eventos de auditoria.
    
    Este endpoint permite definir políticas de retenção para automatizar
    a anonimização e exclusão de dados conforme requisitos de compliance.
    """
    try:
        audit_service = AuditService(db)
        
        # Assegura que o tenant_id e regional_context do middleware sejam usados
        # se não forem explicitamente fornecidos na política
        if not policy.tenant_id:
            policy.tenant_id = tenant_id
        if not policy.regional_context:
            policy.regional_context = regional_context
            
        policy_id = await audit_service.create_retention_policy(policy)
        
        created_policy = await audit_service.get_retention_policy_by_id(policy_id)
        return created_policy
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao criar política de retenção: {str(e)}"
        )


@router.get("/retention-policies", response_model=List[AuditRetentionPolicyResponse])
async def list_retention_policies(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: Optional[str] = Depends(get_regional_context),
    compliance_framework: Optional[ComplianceFramework] = None,
    active_only: bool = True,
    db: AsyncSession = Depends(get_db)
):
    """
    Lista políticas de retenção para um tenant.
    
    Este endpoint recupera políticas de retenção existentes com
    opção de filtragem por contexto regional e framework de compliance.
    """
    try:
        audit_service = AuditService(db)
        policies = await audit_service.get_retention_policies(
            tenant_id=tenant_id,
            regional_context=regional_context,
            compliance_framework=compliance_framework,
            active_only=active_only
        )
        
        return policies
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao listar políticas de retenção: {str(e)}"
        )


@router.get("/retention-policies/{policy_id}", response_model=AuditRetentionPolicyResponse)
async def get_retention_policy(
    policy_id: UUID,
    tenant_id: str = Depends(get_tenant_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Recupera uma política de retenção pelo ID.
    
    Este endpoint obtém os detalhes de uma política de retenção específica,
    com isolamento multi-tenant para segurança.
    """
    try:
        audit_service = AuditService(db)
        policy = await audit_service.get_retention_policy_by_id(policy_id, tenant_id)
        
        if not policy:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Política de retenção {policy_id} não encontrada"
            )
            
        return policy
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao recuperar política de retenção: {str(e)}"
        )


@router.patch("/retention-policies/{policy_id}", response_model=AuditRetentionPolicyResponse)
async def update_retention_policy(
    policy_id: UUID,
    policy_update: AuditRetentionPolicyUpdate,
    tenant_id: str = Depends(get_tenant_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Atualiza uma política de retenção existente.
    
    Este endpoint permite modificar parâmetros de uma política de retenção,
    como período de retenção, campos a serem anonimizados e status ativo.
    """
    try:
        audit_service = AuditService(db)
        
        # Verifica se a política existe e pertence ao tenant
        existing_policy = await audit_service.get_retention_policy_by_id(policy_id, tenant_id)
        if not existing_policy:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Política de retenção {policy_id} não encontrada"
            )
        
        # Atualiza a política
        updated_policy = await audit_service.update_retention_policy(policy_id, policy_update)
        return updated_policy
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao atualizar política de retenção: {str(e)}"
        )


@router.post("/apply-retention", status_code=status.HTTP_200_OK)
async def apply_retention_policies(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: Optional[str] = Depends(get_regional_context),
    batch_size: int = 100,
    dry_run: bool = False,
    db: AsyncSession = Depends(get_db)
):
    """
    Aplica políticas de retenção para eventos de auditoria.
    
    Este endpoint executa a aplicação de políticas de retenção,
    realizando anonimização e exclusão de dados conforme configurado.
    O modo dry_run permite simular a execução sem alterar dados.
    """
    try:
        audit_service = AuditService(db)
        result = await audit_service.apply_retention_policies(
            tenant_id=tenant_id,
            regional_context=regional_context,
            batch_size=batch_size,
            dry_run=dry_run
        )
        
        return result
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao aplicar políticas de retenção: {str(e)}"
        )@router.post("/events/{event_id}/mask", status_code=status.HTTP_200_OK)
async def mask_event_fields(
    event_id: UUID,
    fields_to_mask: List[str],
    tenant_id: str = Depends(get_tenant_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Mascara campos sensíveis em um evento de auditoria.
    
    Este endpoint permite mascarar campos específicos em um evento de auditoria,
    útil para proteger informações sensíveis em logs de auditoria.
    """
    try:
        audit_service = AuditService(db)
        
        # Verifica se o evento existe e pertence ao tenant
        event = await audit_service.get_event_by_id(event_id, tenant_id)
        if not event:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Evento de auditoria {event_id} não encontrado"
            )
        
        # Mascara os campos
        success = await audit_service.mask_sensitive_fields(event_id, fields_to_mask)
        
        if not success:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Falha ao mascarar campos do evento"
            )
        
        return {
            "event_id": str(event_id),
            "masked_fields": fields_to_mask,
            "success": success
        }
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao mascarar campos do evento: {str(e)}"
        )


@router.post("/compliance-reports", response_model=Dict[str, Any], status_code=status.HTTP_201_CREATED)
async def generate_compliance_report(
    compliance_framework: ComplianceFramework,
    start_date: datetime,
    end_date: datetime,
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    report_type: str = "standard",
    report_format: str = "json",
    include_anonymized: bool = False,
    user_id: Optional[str] = None,
    db: AsyncSession = Depends(get_db)
):
    """
    Gera um relatório de compliance para um framework específico.
    
    Este endpoint cria um relatório detalhado de compliance para um determinado período,
    com estatísticas e métricas de eventos de auditoria relevantes para o framework.
    """
    try:
        audit_service = AuditService(db)
        
        # Verifica se datas são válidas
        if start_date >= end_date:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="A data inicial deve ser anterior à data final"
            )
        
        # Gera o relatório
        report_data = await audit_service.generate_compliance_report(
            tenant_id=tenant_id,
            regional_context=regional_context,
            compliance_framework=compliance_framework,
            start_date=start_date,
            end_date=end_date,
            report_type=report_type,
            report_format=report_format,
            user_id=user_id,
            include_anonymized=include_anonymized
        )
        
        return report_data
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao gerar relatório de compliance: {str(e)}"
        )


@router.get("/compliance-reports", response_model=List[AuditComplianceReportResponse])
async def list_compliance_reports(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: Optional[str] = Depends(get_regional_context),
    compliance_framework: Optional[ComplianceFramework] = None,
    status: Optional[ReportStatus] = None,
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: AsyncSession = Depends(get_db)
):
    """
    Lista relatórios de compliance disponíveis.
    
    Este endpoint recupera relatórios de compliance existentes com
    opções de filtragem por contexto regional, framework e status.
    """
    try:
        audit_service = AuditService(db)
        reports = await audit_service.get_compliance_reports(
            tenant_id=tenant_id,
            regional_context=regional_context,
            compliance_framework=compliance_framework,
            status=status,
            limit=limit,
            offset=offset
        )
        
        return reports
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao listar relatórios de compliance: {str(e)}"
        )


@router.get("/compliance-reports/{report_id}", response_model=AuditComplianceReportResponse)
async def get_compliance_report(
    report_id: UUID,
    tenant_id: str = Depends(get_tenant_context),
    db: AsyncSession = Depends(get_db)
):
    """
    Recupera um relatório de compliance pelo ID.
    
    Este endpoint obtém os detalhes de um relatório de compliance específico,
    com isolamento multi-tenant para segurança.
    """
    try:
        audit_service = AuditService(db)
        report = await audit_service.get_compliance_report_by_id(report_id, tenant_id)
        
        if not report:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Relatório de compliance {report_id} não encontrado"
            )
            
        return report
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao recuperar relatório de compliance: {str(e)}"
        )


@router.post("/statistics", response_model=AuditStatisticsResponse, status_code=status.HTTP_201_CREATED)
async def generate_audit_statistics(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    period: str = Query("daily", regex="^(daily|weekly|monthly)$"),
    update_existing: bool = True,
    db: AsyncSession = Depends(get_db)
):
    """
    Gera estatísticas de auditoria para um período específico.
    
    Este endpoint cria estatísticas agregadas de eventos de auditoria,
    úteis para dashboards e análises de tendências.
    """
    try:
        audit_service = AuditService(db)
        
        # Gera estatísticas
        stats_data = await audit_service.generate_audit_statistics(
            tenant_id=tenant_id,
            regional_context=regional_context,
            period=period,
            update_existing=update_existing
        )
        
        return stats_data
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao gerar estatísticas de auditoria: {str(e)}"
        )


@router.get("/statistics", response_model=AuditStatisticsResponse)
async def get_audit_statistics(
    tenant_id: str = Depends(get_tenant_context),
    regional_context: str = Depends(get_regional_context),
    period: str = Query("daily", regex="^(daily|weekly|monthly)$"),
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    db: AsyncSession = Depends(get_db)
):
    """
    Recupera estatísticas de auditoria existentes.
    
    Este endpoint obtém estatísticas agregadas previamente geradas
    para um determinado período e intervalo de datas.
    """
    try:
        audit_service = AuditService(db)
        
        # Recupera estatísticas
        stats_data = await audit_service.get_audit_statistics(
            tenant_id=tenant_id,
            regional_context=regional_context,
            period=period,
            start_date=start_date,
            end_date=end_date
        )
        
        if not stats_data:
            return {
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "period": period,
                "generated_at": None,
                "statistics": {}
            }
            
        return stats_data
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Falha ao recuperar estatísticas de auditoria: {str(e)}"
        )