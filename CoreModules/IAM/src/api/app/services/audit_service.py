"""
INNOVABIZ IAM - Serviço de Auditoria Multi-Contexto
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Implementação do serviço de auditoria com suporte a multi-tenant, multi-regional e compliance
"""

import uuid
import logging
import structlog
import asyncio
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional, Set, Tuple, Union

from sqlalchemy import text, func, and_, or_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.sql import func
from sqlalchemy.orm import selectinload

from app.db.models.audit import (
    AuditEventCreate, AuditEventEntity, AuditEventCategory, 
    AuditEventSeverity, AuditRetentionPolicy, AuditComplianceReport,
    AuditStatistics, ComplianceFramework, ReportStatus
)
from app.core.config import settings

# Configuração do logger estruturado
logger = structlog.get_logger(__name__)


class AuditService:
    """
    Serviço de Auditoria Multi-Contexto para o INNOVABIZ IAM.
    
    Responsável por:
    - Registro de eventos de auditoria
    - Aplicação de políticas de retenção e anonimização
    - Geração de relatórios de compliance
    - Estatísticas e análises de auditoria
    
    Suporta:
    - Multi-tenant: Isolamento por tenant com chaves de particionamento
    - Multi-regional: Diferentes contextos regionais (BR, US, EU, AO)
    - Multi-regulatório: Compliance com LGPD, GDPR, SOX, PCI DSS, etc.
    """
    
    def __init__(self, db_session: AsyncSession):
        """
        Inicializa o serviço de auditoria.
        
        Args:
            db_session: Sessão assíncrona do SQLAlchemy
        """
        self.db_session = db_session
        self.logger = logger.bind(service="audit_service")
        
        # Mapeamento de contextos regionais para frameworks de compliance
        self.regional_compliance_map = {
            "BR": ["LGPD", "BACEN"],
            "US": ["SOX", "NIST"],
            "EU": ["GDPR", "PSD2"],
            "AO": ["BNA"]
        }
        
        # Configurações padrão de retenção (em dias)
        self.default_retention_days = getattr(
            settings, "AUDIT_DEFAULT_RETENTION_DAYS", 730
        )  # 2 anos por padrão
        
        # Configurações para processamento em lote
        self.batch_size = getattr(settings, "AUDIT_BATCH_SIZE", 100)
        self.batch_interval = getattr(settings, "AUDIT_BATCH_INTERVAL", 60)  # segundos
    
    async def create_audit_event(self, event_data: AuditEventCreate) -> AuditEventEntity:
        """
        Cria um novo evento de auditoria no sistema.
        
        Args:
            event_data: Dados do evento de auditoria a ser criado
            
        Returns:
            Evento de auditoria criado
        """
        try:
            # Gera ID UUID se não fornecido
            event_id = uuid.uuid4()
            
            # Detecta frameworks de compliance aplicáveis
            compliance_frameworks = self._detect_compliance_frameworks(event_data.regional_context)
            
            # Gera chave de particionamento
            partition_key = self._generate_partition_key(
                tenant_id=event_data.tenant_id,
                regional_context=event_data.regional_context
            )
            
            # Cria a entidade de evento de auditoria
            audit_event = AuditEventEntity(
                id=event_id,
                category=event_data.category,
                action=event_data.action,
                description=event_data.description,
                resource_type=event_data.resource_type,
                resource_id=event_data.resource_id,
                resource_name=event_data.resource_name,
                severity=event_data.severity,
                success=event_data.success,
                error_message=event_data.error_message,
                details=event_data.details or {},
                tags=event_data.tags or [],
                tenant_id=event_data.tenant_id,
                regional_context=event_data.regional_context,
                country_code=event_data.country_code,
                language=event_data.language,
                user_id=event_data.user_id,
                user_name=event_data.user_name,
                correlation_id=event_data.correlation_id,
                source_ip=event_data.source_ip,
                http_details=event_data.http_details.dict() if event_data.http_details else None,
                compliance_frameworks=compliance_frameworks,
                masked_fields=[],
                anonymized_fields=[],
                partition_key=partition_key
            )
            
            # Persiste no banco de dados
            self.db_session.add(audit_event)
            await self.db_session.flush()
            await self.db_session.commit()
            
            self.logger.info(
                "audit_event_created",
                event_id=str(audit_event.id),
                tenant_id=audit_event.tenant_id,
                category=str(audit_event.category),
                action=audit_event.action,
                regional_context=audit_event.regional_context
            )
            
            return audit_event
        
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "audit_event_creation_failed",
                error=str(e),
                tenant_id=event_data.tenant_id,
                category=str(event_data.category),
                action=event_data.action
            )
            raise
    
    async def get_audit_event_by_id(self, event_id: uuid.UUID) -> Optional[AuditEventEntity]:
        """
        Recupera um evento de auditoria específico pelo seu ID.
        
        Args:
            event_id: ID único do evento de auditoria
            
        Returns:
            Evento de auditoria encontrado ou None
        """
        query = select(AuditEventEntity).where(AuditEventEntity.id == event_id)
        result = await self.db_session.execute(query)
        return result.scalars().first()
    
    async def get_audit_events_by_tenant(
        self,
        tenant_id: str,
        regional_context: Optional[str] = None,
        category: Optional[AuditEventCategory] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        success: Optional[bool] = None,
        limit: int = 100,
        offset: int = 0
    ) -> List[AuditEventEntity]:
        """
        Recupera eventos de auditoria para um tenant específico com opções de filtragem.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Filtro de contexto regional (opcional)
            category: Filtro de categoria (opcional)
            start_date: Data inicial para filtro (opcional)
            end_date: Data final para filtro (opcional)
            success: Filtro por sucesso/falha (opcional)
            limit: Número máximo de resultados
            offset: Deslocamento para paginação
            
        Returns:
            Lista de eventos de auditoria
        """
        query = select(AuditEventEntity).where(AuditEventEntity.tenant_id == tenant_id)
        
        # Aplica filtros adicionais
        if regional_context:
            query = query.where(AuditEventEntity.regional_context == regional_context)
        
        if category:
            query = query.where(AuditEventEntity.category == category)
        
        if start_date:
            query = query.where(AuditEventEntity.created_at >= start_date)
        
        if end_date:
            query = query.where(AuditEventEntity.created_at <= end_date)
        
        if success is not None:
            query = query.where(AuditEventEntity.success == success)
        
        # Ordenação e paginação
        query = query.order_by(AuditEventEntity.created_at.desc())
        query = query.offset(offset).limit(limit)
        
        result = await self.db_session.execute(query)
        return result.scalars().all()
    
    async def get_audit_events_by_correlation_id(self, correlation_id: str) -> List[AuditEventEntity]:
        """
        Recupera eventos de auditoria pelo ID de correlação (para rastreamento de fluxos).
        
        Args:
            correlation_id: ID de correlação para rastrear eventos relacionados
            
        Returns:
            Lista de eventos de auditoria com o mesmo ID de correlação
        """
        query = select(AuditEventEntity).where(
            AuditEventEntity.correlation_id == correlation_id
        ).order_by(AuditEventEntity.created_at)
        
        result = await self.db_session.execute(query)
        return result.scalars().all()
    
    async def create_retention_policy(
        self,
        tenant_id: str,
        regional_context: Optional[str],
        retention_days: int,
        compliance_framework: ComplianceFramework,
        category: Optional[AuditEventCategory],
        description: str,
        automatic_anonymization: bool = False,
        anonymization_fields: List[str] = []
    ) -> AuditRetentionPolicy:
        """
        Cria uma nova política de retenção para eventos de auditoria.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional (ou None para global)
            retention_days: Número de dias para reter os eventos
            compliance_framework: Framework de compliance aplicável
            category: Categoria de evento (ou None para todas)
            description: Descrição da política
            automatic_anonymization: Se deve anonimizar automaticamente campos sensíveis
            anonymization_fields: Lista de campos a serem anonimizados
            
        Returns:
            Política de retenção criada
        """
        try:
            policy = AuditRetentionPolicy(
                id=uuid.uuid4(),
                tenant_id=tenant_id,
                regional_context=regional_context,
                retention_days=retention_days,
                compliance_framework=compliance_framework,
                category=category,
                description=description,
                automatic_anonymization=automatic_anonymization,
                anonymization_fields=anonymization_fields,
                active=True
            )
            
            self.db_session.add(policy)
            await self.db_session.flush()
            await self.db_session.commit()
            
            self.logger.info(
                "retention_policy_created",
                policy_id=str(policy.id),
                tenant_id=policy.tenant_id,
                regional_context=policy.regional_context,
                compliance_framework=str(policy.compliance_framework)
            )
            
            return policy
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "retention_policy_creation_failed",
                error=str(e),
                tenant_id=tenant_id,
                compliance_framework=str(compliance_framework)
            )
            raise
    
    async def apply_retention_policies(self, tenant_id: str) -> Dict[str, int]:
        """
        Aplica todas as políticas de retenção ativas para um tenant.
        
        Args:
            tenant_id: ID do tenant
            
        Returns:
            Dicionário com contagem de eventos anonimizados/excluídos por política
        """
        result = {}
        
        # Busca todas as políticas ativas para o tenant
        query = select(AuditRetentionPolicy).where(
            and_(
                AuditRetentionPolicy.tenant_id == tenant_id,
                AuditRetentionPolicy.active == True
            )
        )
        policies = await self.db_session.execute(query)
        
        # Para cada política, chama a função do banco de dados para aplicá-la
        for policy in policies.scalars():
            try:
                # Usa função armazenada do PostgreSQL para aplicar política
                stmt = text("CALL apply_retention_policy(:policy_id)")
                await self.db_session.execute(stmt, {"policy_id": policy.id})
                await self.db_session.commit()
                
                # Recupera estatísticas da última aplicação
                stats_query = select(AuditStatistics).where(
                    and_(
                        AuditStatistics.tenant_id == tenant_id,
                        AuditStatistics.statistics_type == 'RETENTION_POLICY_APPLICATION'
                    )
                ).order_by(AuditStatistics.created_at.desc()).limit(1)
                
                stats_result = await self.db_session.execute(stats_query)
                stats = stats_result.scalars().first()
                
                if stats:
                    result[str(policy.id)] = {
                        "anonymized_count": stats.statistics_data.get("anonymized_count", 0),
                        "deleted_count": stats.statistics_data.get("deleted_count", 0),
                        "compliance_framework": str(policy.compliance_framework)
                    }
            except Exception as e:
                self.logger.error(
                    "retention_policy_application_failed",
                    error=str(e),
                    policy_id=str(policy.id),
                    tenant_id=tenant_id
                )
                result[str(policy.id)] = {"error": str(e)}
                await self.db_session.rollback()
        
        return result
    
    async def mask_sensitive_fields(
        self,
        event_id: uuid.UUID,
        field_names: List[str]
    ) -> bool:
        """
        Mascara campos sensíveis em um evento de auditoria.
        
        Args:
            event_id: ID do evento de auditoria
            field_names: Lista de nomes dos campos a serem mascarados
            
        Returns:
            True se o mascaramento foi bem-sucedido
        """
        try:
            # Usa função armazenada do PostgreSQL para mascaramento
            stmt = text("SELECT mask_sensitive_fields(:event_id, :field_names)")
            await self.db_session.execute(
                stmt, 
                {"event_id": event_id, "field_names": field_names}
            )
            await self.db_session.commit()
            
            self.logger.info(
                "fields_masked",
                event_id=str(event_id),
                field_count=len(field_names)
            )
            
            return True
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "field_masking_failed",
                error=str(e),
                event_id=str(event_id)
            )
            return False
    
    async def generate_compliance_report(
        self,
        tenant_id: str,
        regional_context: Optional[str],
        compliance_framework: ComplianceFramework,
        start_date: datetime,
        end_date: datetime,
        report_name: str,
        report_description: Optional[str] = None,
        created_by: Optional[str] = None
    ) -> AuditComplianceReport:
        """
        Gera um novo relatório de compliance para eventos de auditoria.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional (ou None para global)
            compliance_framework: Framework de compliance aplicável
            start_date: Data inicial do relatório
            end_date: Data final do relatório
            report_name: Nome do relatório
            report_description: Descrição do relatório (opcional)
            created_by: ID do usuário que criou o relatório (opcional)
            
        Returns:
            Relatório de compliance gerado
        """
        try:
            # Cria o registro do relatório com status PENDING
            report = AuditComplianceReport(
                id=uuid.uuid4(),
                tenant_id=tenant_id,
                regional_context=regional_context,
                compliance_framework=compliance_framework,
                report_name=report_name,
                report_description=report_description,
                status=ReportStatus.PENDING,
                start_date=start_date,
                end_date=end_date,
                created_by=created_by
            )
            
            self.db_session.add(report)
            await self.db_session.flush()
            
            # Conta eventos para o relatório
            count_query = select(func.count()).select_from(AuditEventEntity).where(
                and_(
                    AuditEventEntity.tenant_id == tenant_id,
                    AuditEventEntity.created_at.between(start_date, end_date),
                    AuditEventEntity.compliance_frameworks.contains([str(compliance_framework)])
                )
            )
            
            if regional_context:
                count_query = count_query.where(AuditEventEntity.regional_context == regional_context)
            
            result = await self.db_session.execute(count_query)
            event_count = result.scalar()
            report.event_count = event_count
            
            # Gera os dados do relatório
            report_data = await self._generate_compliance_report_data(
                tenant_id=tenant_id,
                regional_context=regional_context,
                compliance_framework=compliance_framework,
                start_date=start_date,
                end_date=end_date
            )
            
            report.report_data = report_data
            report.status = ReportStatus.COMPLETED
            
            await self.db_session.commit()
            
            self.logger.info(
                "compliance_report_generated",
                report_id=str(report.id),
                tenant_id=report.tenant_id,
                compliance_framework=str(report.compliance_framework),
                event_count=report.event_count
            )
            
            return report
        
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "compliance_report_generation_failed",
                error=str(e),
                tenant_id=tenant_id,
                compliance_framework=str(compliance_framework)
            )
            raise
    
    async def generate_audit_statistics(
        self,
        tenant_id: str,
        regional_context: Optional[str],
        start_date: datetime,
        end_date: datetime,
        statistics_type: str = "PERIODIC_SUMMARY",
        group_by: List[str] = ["category", "success", "severity"]
    ) -> AuditStatistics:
        """
        Gera estatísticas agregadas para eventos de auditoria.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional (ou None para global)
            start_date: Data inicial para estatísticas
            end_date: Data final para estatísticas
            statistics_type: Tipo de estatística
            group_by: Campos para agrupamento das estatísticas
            
        Returns:
            Registro de estatísticas gerado
        """
        try:
            # Usa função armazenada do PostgreSQL para gerar estatísticas
            stmt = text(
                "SELECT generate_audit_statistics(:tenant_id, :regional_context, :start_date, :end_date, :group_by)"
            )
            result = await self.db_session.execute(
                stmt,
                {
                    "tenant_id": tenant_id,
                    "regional_context": regional_context,
                    "start_date": start_date,
                    "end_date": end_date,
                    "group_by": group_by
                }
            )
            statistics_id = result.scalar()
            
            # Recupera o registro de estatísticas
            query = select(AuditStatistics).where(AuditStatistics.id == statistics_id)
            statistics_result = await self.db_session.execute(query)
            statistics = statistics_result.scalars().first()
            
            self.logger.info(
                "audit_statistics_generated",
                statistics_id=str(statistics_id),
                tenant_id=tenant_id,
                regional_context=regional_context,
                period=f"{start_date.isoformat()} to {end_date.isoformat()}"
            )
            
            return statistics
        
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "audit_statistics_generation_failed",
                error=str(e),
                tenant_id=tenant_id,
                period=f"{start_date.isoformat()} to {end_date.isoformat()}"
            )
            raise    async def process_batch_events(
        self,
        events: List[AuditEventCreate]
    ) -> Dict[str, Any]:
        """
        Processa um lote de eventos de auditoria de forma eficiente.
        
        Args:
            events: Lista de eventos de auditoria a serem processados
            
        Returns:
            Dicionário com resultados do processamento em lote
        """
        successful_events = 0
        failed_events = 0
        event_ids = []
        
        try:
            # Processar eventos em lote
            for event_data in events:
                try:
                    # Gera ID UUID
                    event_id = uuid.uuid4()
                    
                    # Detecta frameworks de compliance aplicáveis
                    compliance_frameworks = self._detect_compliance_frameworks(event_data.regional_context)
                    
                    # Gera chave de particionamento
                    partition_key = self._generate_partition_key(
                        tenant_id=event_data.tenant_id,
                        regional_context=event_data.regional_context
                    )
                    
                    # Cria a entidade de evento de auditoria
                    audit_event = AuditEventEntity(
                        id=event_id,
                        category=event_data.category,
                        action=event_data.action,
                        description=event_data.description,
                        resource_type=event_data.resource_type,
                        resource_id=event_data.resource_id,
                        resource_name=event_data.resource_name,
                        severity=event_data.severity,
                        success=event_data.success,
                        error_message=event_data.error_message,
                        details=event_data.details or {},
                        tags=event_data.tags or [],
                        tenant_id=event_data.tenant_id,
                        regional_context=event_data.regional_context,
                        country_code=event_data.country_code,
                        language=event_data.language,
                        user_id=event_data.user_id,
                        user_name=event_data.user_name,
                        correlation_id=event_data.correlation_id,
                        source_ip=event_data.source_ip,
                        http_details=event_data.http_details.dict() if event_data.http_details else None,
                        compliance_frameworks=compliance_frameworks,
                        masked_fields=[],
                        anonymized_fields=[],
                        partition_key=partition_key
                    )
                    
                    # Adiciona evento para insert em lote
                    self.db_session.add(audit_event)
                    event_ids.append(str(event_id))
                    successful_events += 1
                    
                except Exception as e:
                    self.logger.warning(
                        "batch_event_processing_item_failed",
                        error=str(e),
                        tenant_id=event_data.tenant_id,
                        action=event_data.action
                    )
                    failed_events += 1
            
            # Commit do lote completo
            await self.db_session.commit()
            
            self.logger.info(
                "batch_events_processed",
                successful_count=successful_events,
                failed_count=failed_events,
                total_count=len(events)
            )
            
            return {
                "successful_count": successful_events,
                "failed_count": failed_events,
                "total_count": len(events),
                "event_ids": event_ids
            }
            
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "batch_events_processing_failed",
                error=str(e),
                event_count=len(events)
            )
            return {
                "successful_count": 0,
                "failed_count": len(events),
                "total_count": len(events),
                "error": str(e)
            }
    
    async def get_retention_policies(
        self,
        tenant_id: str,
        regional_context: Optional[str] = None,
        compliance_framework: Optional[ComplianceFramework] = None,
        active_only: bool = True
    ) -> List[AuditRetentionPolicy]:
        """
        Recupera políticas de retenção para um tenant.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Filtro de contexto regional (opcional)
            compliance_framework: Filtro de framework de compliance (opcional)
            active_only: Se deve retornar apenas políticas ativas
            
        Returns:
            Lista de políticas de retenção
        """
        query = select(AuditRetentionPolicy).where(AuditRetentionPolicy.tenant_id == tenant_id)
        
        if regional_context:
            query = query.where(
                or_(
                    AuditRetentionPolicy.regional_context == regional_context,
                    AuditRetentionPolicy.regional_context == None  # Políticas globais
                )
            )
        
        if compliance_framework:
            query = query.where(AuditRetentionPolicy.compliance_framework == compliance_framework)
        
        if active_only:
            query = query.where(AuditRetentionPolicy.active == True)
        
        result = await self.db_session.execute(query)
        return result.scalars().all()
    
    async def get_compliance_reports(
        self,
        tenant_id: str,
        regional_context: Optional[str] = None,
        compliance_framework: Optional[ComplianceFramework] = None,
        status: Optional[ReportStatus] = None,
        limit: int = 20,
        offset: int = 0
    ) -> List[AuditComplianceReport]:
        """
        Recupera relatórios de compliance para um tenant.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Filtro de contexto regional (opcional)
            compliance_framework: Filtro de framework de compliance (opcional)
            status: Filtro de status do relatório (opcional)
            limit: Número máximo de resultados
            offset: Deslocamento para paginação
            
        Returns:
            Lista de relatórios de compliance
        """
        query = select(AuditComplianceReport).where(AuditComplianceReport.tenant_id == tenant_id)
        
        if regional_context:
            query = query.where(AuditComplianceReport.regional_context == regional_context)
        
        if compliance_framework:
            query = query.where(AuditComplianceReport.compliance_framework == compliance_framework)
        
        if status:
            query = query.where(AuditComplianceReport.status == status)
        
        # Ordenação e paginação
        query = query.order_by(AuditComplianceReport.created_at.desc())
        query = query.offset(offset).limit(limit)
        
        result = await self.db_session.execute(query)
        return result.scalars().all()    async def mask_sensitive_fields(
        self,
        event_id: UUID,
        fields_to_mask: List[str]
    ) -> bool:
        """
        Mascara campos sensíveis em um evento de auditoria utilizando
        procedimento armazenado no banco de dados.
        
        Args:
            event_id: ID do evento de auditoria
            fields_to_mask: Lista de campos a serem mascarados
            
        Returns:
            Booleano indicando sucesso da operação
        """
        try:
            # Executa procedimento armazenado para mascarar campos
            stmt = text("""
                CALL mask_audit_event_fields(:event_id, :fields);
            """)
            
            await self.db_session.execute(
                stmt, 
                {"event_id": str(event_id), "fields": fields_to_mask}
            )
            
            await self.db_session.commit()
            
            self.logger.info(
                "audit_event_fields_masked",
                event_id=str(event_id),
                field_count=len(fields_to_mask)
            )
            
            return True
            
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "audit_event_field_masking_failed",
                event_id=str(event_id),
                error=str(e)
            )
            return False
    
    async def generate_compliance_report(
        self,
        tenant_id: str,
        regional_context: str,
        compliance_framework: ComplianceFramework,
        start_date: datetime,
        end_date: datetime,
        report_type: str = "standard",
        report_format: str = "json",
        user_id: Optional[str] = None,
        include_anonymized: bool = False
    ) -> Dict[str, Any]:
        """
        Gera um relatório de compliance com estatísticas de eventos de auditoria
        para um determinado framework de compliance.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional
            compliance_framework: Framework de compliance
            start_date: Data inicial do período do relatório
            end_date: Data final do período do relatório
            report_type: Tipo de relatório (standard, detailed, summary)
            report_format: Formato do relatório (json, csv, pdf)
            user_id: ID do usuário que solicitou o relatório
            include_anonymized: Se deve incluir eventos anonimizados
            
        Returns:
            Dicionário com detalhes do relatório gerado
        """
        try:
            # Gera um ID para o relatório
            report_id = uuid.uuid4()
            
            # Cria registro de relatório
            report = AuditComplianceReport(
                id=report_id,
                tenant_id=tenant_id,
                regional_context=regional_context,
                compliance_framework=compliance_framework,
                report_type=report_type,
                report_format=report_format,
                start_date=start_date,
                end_date=end_date,
                status=ReportStatus.PROCESSING,
                created_by=user_id,
                include_anonymized=include_anonymized
            )
            
            self.db_session.add(report)
            await self.db_session.commit()
            
            # Conta eventos por categoria
            query = select(
                AuditEventEntity.category,
                func.count(AuditEventEntity.id).label('count')
            ).where(
                AuditEventEntity.tenant_id == tenant_id,
                AuditEventEntity.regional_context == regional_context,
                func.array_position(AuditEventEntity.compliance_frameworks, compliance_framework.value) > 0,
                AuditEventEntity.created_at >= start_date,
                AuditEventEntity.created_at <= end_date
            )
            
            if not include_anonymized:
                # Exclui eventos anonimizados
                query = query.where(
                    or_(
                        AuditEventEntity.anonymized_fields == None,
                        func.array_length(AuditEventEntity.anonymized_fields, 1) == 0
                    )
                )
                
            # Agrupamento por categoria
            query = query.group_by(AuditEventEntity.category)
            
            # Executa query
            result = await self.db_session.execute(query)
            events_by_category = {row.category: row.count for row in result}
            
            # Estatísticas por severidade
            query = select(
                AuditEventEntity.severity,
                func.count(AuditEventEntity.id).label('count')
            ).where(
                AuditEventEntity.tenant_id == tenant_id,
                AuditEventEntity.regional_context == regional_context,
                func.array_position(AuditEventEntity.compliance_frameworks, compliance_framework.value) > 0,
                AuditEventEntity.created_at >= start_date,
                AuditEventEntity.created_at <= end_date
            ).group_by(AuditEventEntity.severity)
            
            result = await self.db_session.execute(query)
            events_by_severity = {row.severity: row.count for row in result}
            
            # Estatísticas de sucesso/falha
            query = select(
                AuditEventEntity.success,
                func.count(AuditEventEntity.id).label('count')
            ).where(
                AuditEventEntity.tenant_id == tenant_id,
                AuditEventEntity.regional_context == regional_context,
                func.array_position(AuditEventEntity.compliance_frameworks, compliance_framework.value) > 0,
                AuditEventEntity.created_at >= start_date,
                AuditEventEntity.created_at <= end_date
            ).group_by(AuditEventEntity.success)
            
            result = await self.db_session.execute(query)
            events_by_success = {str(row.success): row.count for row in result}
            
            # Dados do relatório
            report_data = {
                "report_id": str(report_id),
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "compliance_framework": compliance_framework.value,
                "period": {
                    "start_date": start_date.isoformat(),
                    "end_date": end_date.isoformat()
                },
                "statistics": {
                    "total_events": sum(events_by_category.values()),
                    "events_by_category": events_by_category,
                    "events_by_severity": events_by_severity,
                    "events_by_success": events_by_success
                },
                "generation_date": datetime.now().isoformat(),
                "report_type": report_type,
                "report_format": report_format
            }
            
            # Atualiza relatório com os dados gerados
            await self.db_session.execute(
                update(AuditComplianceReport).where(
                    AuditComplianceReport.id == report_id
                ).values(
                    status=ReportStatus.COMPLETED,
                    report_data=report_data,
                    completed_at=datetime.now()
                )
            )
            
            await self.db_session.commit()
            
            self.logger.info(
                "compliance_report_generated",
                report_id=str(report_id),
                tenant_id=tenant_id,
                regional_context=regional_context,
                compliance_framework=compliance_framework.value,
                event_count=sum(events_by_category.values())
            )
            
            return report_data
            
        except Exception as e:
            await self.db_session.rollback()
            
            # Atualiza status do relatório para falha
            if 'report_id' in locals():
                await self.db_session.execute(
                    update(AuditComplianceReport).where(
                        AuditComplianceReport.id == report_id
                    ).values(
                        status=ReportStatus.FAILED,
                        error_message=str(e)
                    )
                )
                
                await self.db_session.commit()
            
            self.logger.error(
                "compliance_report_generation_failed",
                tenant_id=tenant_id,
                regional_context=regional_context,
                compliance_framework=compliance_framework.value,
                error=str(e)
            )
            
            raise AuditServiceError(f"Falha ao gerar relatório de compliance: {e}")    async def apply_retention_policies(
        self,
        tenant_id: str,
        regional_context: Optional[str] = None,
        batch_size: int = 100,
        dry_run: bool = False
    ) -> Dict[str, Any]:
        """
        Aplica políticas de retenção para eventos de auditoria, incluindo
        anonimização e exclusão conforme regulamentos de compliance.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional opcional para aplicar políticas específicas
            batch_size: Tamanho do lote para processamento
            dry_run: Se verdadeiro, apenas simula a aplicação sem alterar dados
            
        Returns:
            Dicionário com estatísticas da aplicação das políticas
        """
        try:
            # Obtém políticas de retenção ativas
            query = select(AuditRetentionPolicy).where(
                AuditRetentionPolicy.tenant_id == tenant_id,
                AuditRetentionPolicy.active == True
            )
            
            if regional_context:
                query = query.where(
                    or_(
                        AuditRetentionPolicy.regional_context == regional_context,
                        AuditRetentionPolicy.regional_context == None  # Políticas globais
                    )
                )
                
            result = await self.db_session.execute(query)
            policies = result.scalars().all()
            
            if not policies:
                self.logger.warning(
                    "no_active_retention_policies",
                    tenant_id=tenant_id,
                    regional_context=regional_context
                )
                return {"message": "Nenhuma política de retenção ativa encontrada", "applied": 0}
            
            # Executa procedimento armazenado para aplicar políticas
            stmt = text("""
                SELECT apply_audit_retention_policies(
                    :tenant_id, 
                    :regional_context,
                    :batch_size,
                    :dry_run
                );
            """)
            
            params = {
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "batch_size": batch_size,
                "dry_run": dry_run
            }
            
            result = await self.db_session.execute(stmt, params)
            result_row = result.fetchone()
            
            # Extrai estatísticas do resultado
            stats = json.loads(result_row[0]) if result_row else {"anonymized": 0, "deleted": 0}
            
            if not dry_run:
                await self.db_session.commit()
                
                self.logger.info(
                    "retention_policies_applied",
                    tenant_id=tenant_id,
                    regional_context=regional_context,
                    anonymized_count=stats.get("anonymized", 0),
                    deleted_count=stats.get("deleted", 0)
                )
            else:
                self.logger.info(
                    "retention_policies_dry_run",
                    tenant_id=tenant_id,
                    regional_context=regional_context,
                    would_anonymize=stats.get("anonymized", 0),
                    would_delete=stats.get("deleted", 0)
                )
            
            return {
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "dry_run": dry_run,
                "policies_applied": len(policies),
                "statistics": stats
            }
            
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "retention_policy_application_failed",
                tenant_id=tenant_id,
                regional_context=regional_context,
                error=str(e)
            )
            raise AuditServiceError(f"Falha ao aplicar políticas de retenção: {e}")
    
    async def generate_audit_statistics(
        self,
        tenant_id: str,
        regional_context: str,
        period: str = "daily",
        update_existing: bool = True
    ) -> Dict[str, Any]:
        """
        Gera estatísticas agregadas de auditoria por período para um tenant
        e contexto regional específicos.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional
            period: Período de agregação (daily, weekly, monthly)
            update_existing: Se deve atualizar estatísticas existentes
            
        Returns:
            Dicionário com estatísticas geradas
        """
        try:
            # Executa função armazenada no banco para gerar estatísticas
            stmt = text("""
                SELECT generate_audit_statistics(
                    :tenant_id, 
                    :regional_context,
                    :period,
                    :update_existing
                );
            """)
            
            params = {
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "period": period,
                "update_existing": update_existing
            }
            
            result = await self.db_session.execute(stmt, params)
            result_row = result.fetchone()
            
            # Extrai estatísticas do resultado
            stats = json.loads(result_row[0]) if result_row else {}
            
            await self.db_session.commit()
            
            self.logger.info(
                "audit_statistics_generated",
                tenant_id=tenant_id,
                regional_context=regional_context,
                period=period,
                statistic_count=len(stats)
            )
            
            return {
                "tenant_id": tenant_id,
                "regional_context": regional_context,
                "period": period,
                "generated_at": datetime.now().isoformat(),
                "statistics": stats
            }
            
        except Exception as e:
            await self.db_session.rollback()
            self.logger.error(
                "audit_statistics_generation_failed",
                tenant_id=tenant_id,
                regional_context=regional_context,
                period=period,
                error=str(e)
            )
            raise AuditServiceError(f"Falha ao gerar estatísticas de auditoria: {e}")
    
    # Métodos internos auxiliares
    
    def _detect_compliance_frameworks(self, regional_context: str) -> List[str]:
        """
        Detecta frameworks de compliance aplicáveis com base no contexto regional.
        
        Args:
            regional_context: Código do contexto regional (ex: BR, US, EU, AO)
            
        Returns:
            Lista de frameworks de compliance aplicáveis
        """
        frameworks = []
        
        # Mapeamento de contextos regionais para frameworks
        region_frameworks = {
            "BR": [ComplianceFramework.LGPD.value, ComplianceFramework.BACEN.value],
            "US": [ComplianceFramework.SOX.value, ComplianceFramework.PCI_DSS.value],
            "EU": [ComplianceFramework.GDPR.value, ComplianceFramework.PSD2.value],
            "AO": [ComplianceFramework.BNA.value]
        }
        
        # Adiciona frameworks específicos da região
        if regional_context in region_frameworks:
            frameworks.extend(region_frameworks[regional_context])
        
        # Adiciona frameworks globais para todos os contextos
        frameworks.append(ComplianceFramework.ISO_27001.value)
        
        return frameworks
    
    def _generate_partition_key(self, tenant_id: str, regional_context: str) -> str:
        """
        Gera uma chave de particionamento para eventos de auditoria.
        
        Args:
            tenant_id: ID do tenant
            regional_context: Contexto regional
            
        Returns:
            Chave de particionamento no formato tenant_id:regional_context:year_month
        """
        current_date = datetime.now()
        year_month = current_date.strftime("%Y%m")
        return f"{tenant_id}:{regional_context}:{year_month}"