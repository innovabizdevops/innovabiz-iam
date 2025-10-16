"""
INNOVABIZ IAM - Testes de Integração para Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Testes de integração para o serviço de auditoria multi-contexto
"""

import pytest
import uuid
import asyncio
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Union

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select

from app.models.audit import (
    AuditEventCreate, AuditEventEntity, AuditRetentionPolicy, 
    AuditComplianceReport, AuditStatistics, ComplianceFramework,
    AuditEventCategory, AuditEventSeverity
)
from app.services.audit_service import AuditService
from tests.conftest.config import TEST_TENANTS, TEST_REGIONAL_CONTEXTS


@pytest.mark.asyncio
class TestAuditIntegration:
    """
    Testes de integração para o serviço de auditoria multi-contexto.
    Valida a integração com o banco de dados PostgreSQL.
    """
    
    async def test_database_crud_operations(self, db_session: AsyncSession, audit_service: AuditService, sample_audit_events):
        """
        Testa operações CRUD completas com o banco de dados.
        """
        # 1. CREATE - Cria um evento de auditoria
        event_data = sample_audit_events["br_login_success"]
        created_event = await audit_service.create_audit_event(event_data)
        
        # Confirma que foi persistido no banco
        stmt = select(AuditEventEntity).where(AuditEventEntity.id == created_event.id)
        result = await db_session.execute(stmt)
        db_event = result.scalars().first()
        
        assert db_event is not None
        assert db_event.id == created_event.id
        assert db_event.tenant_id == event_data.tenant_id
        assert db_event.regional_context == "BR"
        assert "LGPD" in db_event.compliance_frameworks
        
        # 2. READ - Recupera o evento pelo serviço
        retrieved_event = await audit_service.get_audit_event_by_id(created_event.id)
        assert retrieved_event is not None
        assert retrieved_event.id == created_event.id
        
        # 3. UPDATE - Atualiza campos do evento (via marcação de mascaramento)
        await audit_service.mask_sensitive_fields(
            event_id=created_event.id,
            field_names=["source_ip", "details.method"]
        )
        
        # Verifica se o mascaramento foi aplicado
        updated_event = await audit_service.get_audit_event_by_id(created_event.id)
        assert "source_ip" in updated_event.masked_fields
        assert "details.method" in updated_event.masked_fields
        
        # 4. DELETE - Não implementamos delete físico para eventos de auditoria (imutabilidade)
        # Podemos apenas marcar como expirados ou anonimizados para compliance
        
    async def test_retention_policy_database_integration(
        self, db_session: AsyncSession, audit_service: AuditService
    ):
        """
        Testa integração das políticas de retenção com o banco de dados.
        """
        # Cria políticas de retenção para diferentes contextos regionais
        policies = {}
        
        # Brasil - LGPD (2 anos)
        policies["br_policy"] = await audit_service.create_retention_policy(
            tenant_id=TEST_TENANTS["BR"][0],
            regional_context="BR",
            retention_days=730,  # 2 anos
            compliance_framework=ComplianceFramework.LGPD,
            category=AuditEventCategory.AUTHENTICATION,
            description="Política de retenção LGPD - Autenticação",
            automatic_anonymization=True,
            anonymization_fields=["source_ip", "user_name"]
        )
        
        # EUA - SOX (7 anos)
        policies["us_policy"] = await audit_service.create_retention_policy(
            tenant_id=TEST_TENANTS["US"][0],
            regional_context="US",
            retention_days=2555,  # 7 anos
            compliance_framework=ComplianceFramework.SOX,
            category=AuditEventCategory.FINANCIAL,
            description="SOX Retention Policy - Financial",
            automatic_anonymization=False,
            anonymization_fields=[]
        )
        
        # Europa - GDPR (5 anos)
        policies["eu_policy"] = await audit_service.create_retention_policy(
            tenant_id=TEST_TENANTS["EU"][0],
            regional_context="EU",
            retention_days=1825,  # 5 anos
            compliance_framework=ComplianceFramework.GDPR,
            category=AuditEventCategory.CONSENT,
            description="GDPR Retention Policy - Consent",
            automatic_anonymization=True,
            anonymization_fields=["source_ip", "user_id", "details"]
        )
        
        # Angola - BNA (10 anos)
        policies["ao_policy"] = await audit_service.create_retention_policy(
            tenant_id=TEST_TENANTS["AO"][0],
            regional_context="AO",
            retention_days=3650,  # 10 anos
            compliance_framework=ComplianceFramework.BNA,
            category=AuditEventCategory.FINANCIAL,
            description="BNA Retention Policy - Financial",
            automatic_anonymization=False,
            anonymization_fields=[]
        )
        
        # Verifica se todas as políticas foram persistidas no banco de dados
        for key, policy in policies.items():
            stmt = select(AuditRetentionPolicy).where(AuditRetentionPolicy.id == policy.id)
            result = await db_session.execute(stmt)
            db_policy = result.scalars().first()
            
            assert db_policy is not None
            assert db_policy.id == policy.id
            assert db_policy.tenant_id == policy.tenant_id
            assert db_policy.regional_context == policy.regional_context
            assert db_policy.compliance_framework == policy.compliance_framework
            assert db_policy.retention_days == policy.retention_days
    
    async def test_multi_tenant_isolation(self, db_session: AsyncSession, audit_service: AuditService, sample_audit_events):
        """
        Testa o isolamento multi-tenant para eventos de auditoria.
        """
        # Cria eventos para diferentes tenants
        br_tenant = TEST_TENANTS["BR"][0]
        us_tenant = TEST_TENANTS["US"][0]
        
        br_event = sample_audit_events["br_login_success"]
        us_event = sample_audit_events["us_data_access"]
        
        await audit_service.create_audit_event(br_event)
        await audit_service.create_audit_event(us_event)
        
        # Recupera eventos para tenant BR
        br_tenant_events = await audit_service.get_audit_events(tenant_id=br_tenant)
        assert all(event.tenant_id == br_tenant for event in br_tenant_events)
        assert len(br_tenant_events) == 1
        
        # Recupera eventos para tenant US
        us_tenant_events = await audit_service.get_audit_events(tenant_id=us_tenant)
        assert all(event.tenant_id == us_tenant for event in us_tenant_events)
        assert len(us_tenant_events) == 1
    
    async def test_compliance_report_integration(self, db_session: AsyncSession, audit_service: AuditService, sample_audit_events):
        """
        Testa a integração completa de relatórios de compliance.
        """
        # Cria eventos de auditoria para diferentes contextos regionais
        for event_key in ["br_login_success", "us_data_access", "eu_consent_update", "ao_transaction"]:
            await audit_service.create_audit_event(sample_audit_events[event_key])
        
        # Gera relatório de compliance para Brasil - LGPD
        start_date = datetime.utcnow() - timedelta(days=7)
        end_date = datetime.utcnow()
        
        report = await audit_service.create_compliance_report(
            tenant_id=TEST_TENANTS["BR"][0],
            regional_context="BR",
            compliance_framework=ComplianceFramework.LGPD,
            start_date=start_date,
            end_date=end_date,
            report_name="Relatório Semanal LGPD",
            report_description="Relatório de compliance semanal LGPD"
        )
        
        # Verifica persistência no banco de dados
        stmt = select(AuditComplianceReport).where(AuditComplianceReport.id == report.id)
        result = await db_session.execute(stmt)
        db_report = result.scalars().first()
        
        assert db_report is not None
        assert db_report.id == report.id
        assert db_report.tenant_id == TEST_TENANTS["BR"][0]
        assert db_report.regional_context == "BR"
        assert db_report.compliance_framework == ComplianceFramework.LGPD
        assert db_report.event_count > 0
        assert db_report.start_date == start_date
        assert db_report.end_date == end_date
    
    async def test_statistics_aggregation(self, db_session: AsyncSession, audit_service: AuditService, sample_audit_events):
        """
        Testa a agregação de estatísticas de auditoria.
        """
        # Cria múltiplos eventos para teste de agregação
        events = []
        for i in range(5):
            event = sample_audit_events["br_login_success"]
            events.append(await audit_service.create_audit_event(event))
        
        for i in range(3):
            event = sample_audit_events["us_data_access"]
            events.append(await audit_service.create_audit_event(event))
        
        for i in range(2):
            event = sample_audit_events["error_event"]
            events.append(await audit_service.create_audit_event(event))
        
        # Gera estatísticas para tenant BR
        br_stats = await audit_service.generate_audit_statistics(
            tenant_id=TEST_TENANTS["BR"][0],
            start_date=datetime.utcnow() - timedelta(days=1),
            end_date=datetime.utcnow(),
            group_by=["category", "success"]
        )
        
        # Verifica persistência no banco de dados
        stmt = select(AuditStatistics).where(AuditStatistics.tenant_id == TEST_TENANTS["BR"][0])
        result = await db_session.execute(stmt)
        db_stats = result.scalars().all()
        
        assert len(db_stats) > 0
        for stat in db_stats:
            assert stat.tenant_id == TEST_TENANTS["BR"][0]
            assert stat.statistics_data is not None
            assert "total_events" in stat.statistics_data
            assert stat.period_start <= datetime.utcnow()
            assert stat.period_end >= datetime.utcnow() - timedelta(days=1)
    
    async def test_batch_processing_integration(self, db_session: AsyncSession, audit_service: AuditService, sample_audit_events):
        """
        Testa o processamento em lote integrado com o banco de dados.
        """
        # Prepara lote de eventos
        batch_events = []
        for i in range(10):
            if i % 2 == 0:
                event = sample_audit_events["br_login_success"]
            else:
                event = sample_audit_events["us_data_access"]
            
            # Modifica o correlation_id para simular eventos únicos
            event_copy = event.copy(deep=True)
            event_copy.correlation_id = f"{event.correlation_id}-{i}"
            batch_events.append(event_copy)
        
        # Processa em lote
        created_events = await audit_service.batch_create_audit_events(batch_events)
        assert len(created_events) == len(batch_events)
        
        # Verifica persistência no banco de dados
        for event in created_events:
            stmt = select(AuditEventEntity).where(AuditEventEntity.id == event.id)
            result = await db_session.execute(stmt)
            db_event = result.scalars().first()
            
            assert db_event is not None
            assert db_event.id == event.id
            
        # Verifica contagem por tenant
        br_count = await db_session.execute(
            select(AuditEventEntity).where(
                AuditEventEntity.tenant_id == TEST_TENANTS["BR"][0]
            )
        )
        us_count = await db_session.execute(
            select(AuditEventEntity).where(
                AuditEventEntity.tenant_id == TEST_TENANTS["US"][0]
            )
        )
        
        br_events = br_count.scalars().all()
        us_events = us_count.scalars().all()
        
        # Metade dos eventos é BR e metade US
        assert len(br_events) == 5
        assert len(us_events) == 5