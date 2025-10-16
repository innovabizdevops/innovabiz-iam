"""
INNOVABIZ IAM - Testes Unitários para Serviço de Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Testes unitários para o serviço de auditoria multi-contexto do módulo IAM
"""

import pytest
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional

from app.models.audit import (
    AuditEventCreate, AuditEventEntity, AuditEventCategory, 
    AuditEventSeverity, AuditRetentionPolicy, ComplianceFramework
)
from app.services.audit_service import AuditService


@pytest.mark.asyncio
class TestAuditService:
    """
    Testes unitários para o serviço de auditoria multi-contexto.
    """

    async def test_create_audit_event(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a criação de eventos de auditoria básicos.
        """
        # Cria um evento de auditoria para contexto Brasil
        event_data = sample_audit_events["br_login_success"]
        created_event = await audit_service.create_audit_event(event_data)
        
        # Verifica se o evento foi criado corretamente
        assert created_event is not None
        assert created_event.id is not None
        assert created_event.category == event_data.category
        assert created_event.action == event_data.action
        assert created_event.description == event_data.description
        assert created_event.tenant_id == event_data.tenant_id
        assert created_event.regional_context == "BR"
        assert created_event.compliance_frameworks is not None
        assert "LGPD" in created_event.compliance_frameworks
        assert created_event.masked_fields == []
        assert created_event.anonymized_fields == []
        assert created_event.partition_key is not None
        assert created_event.partition_key.startswith(f"{event_data.tenant_id}:BR:")
    
    async def test_create_audit_event_multi_regional(
        self, audit_service: AuditService, sample_audit_events
    ):
        """
        Testa a criação de eventos de auditoria para múltiplas regiões.
        """
        # Cria eventos para diferentes contextos regionais
        events = {}
        for region_key in ["br_login_success", "us_data_access", "eu_consent_update", "ao_transaction"]:
            event_data = sample_audit_events[region_key]
            created_event = await audit_service.create_audit_event(event_data)
            events[region_key] = created_event
        
        # Verifica Brasil
        br_event = events["br_login_success"]
        assert br_event.regional_context == "BR"
        assert "LGPD" in br_event.compliance_frameworks
        assert br_event.language == "pt-BR"
        
        # Verifica EUA
        us_event = events["us_data_access"]
        assert us_event.regional_context == "US"
        assert "SOX" in us_event.compliance_frameworks
        assert us_event.language == "en-US"
        
        # Verifica Europa
        eu_event = events["eu_consent_update"]
        assert eu_event.regional_context == "EU"
        assert "GDPR" in eu_event.compliance_frameworks
        assert eu_event.language == "de-DE"
        
        # Verifica Angola
        ao_event = events["ao_transaction"]
        assert ao_event.regional_context == "AO"
        assert "BNA" in ao_event.compliance_frameworks
        assert ao_event.language == "pt-AO"
    
    async def test_get_audit_event_by_id(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a recuperação de um evento de auditoria por ID.
        """
        # Cria um evento
        event_data = sample_audit_events["br_login_success"]
        created_event = await audit_service.create_audit_event(event_data)
        
        # Recupera o evento pelo ID
        retrieved_event = await audit_service.get_audit_event_by_id(created_event.id)
        
        # Verifica se os dados são consistentes
        assert retrieved_event is not None
        assert retrieved_event.id == created_event.id
        assert retrieved_event.category == created_event.category
        assert retrieved_event.action == created_event.action
        assert retrieved_event.description == created_event.description
        assert retrieved_event.tenant_id == created_event.tenant_id
        assert retrieved_event.regional_context == created_event.regional_context
    
    async def test_get_audit_events_by_tenant(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a recuperação de eventos de auditoria por tenant.
        """
        # Cria múltiplos eventos para diferentes tenants
        for event_key in ["br_login_success", "us_data_access"]:
            await audit_service.create_audit_event(sample_audit_events[event_key])
        
        # Recupera eventos por tenant - Brasil
        br_events = await audit_service.get_audit_events(tenant_id="tenant_br_1", limit=10)
        assert len(br_events) == 1
        assert br_events[0].tenant_id == "tenant_br_1"
        assert br_events[0].regional_context == "BR"
        
        # Recupera eventos por tenant - EUA
        us_events = await audit_service.get_audit_events(tenant_id="tenant_us_1", limit=10)
        assert len(us_events) == 1
        assert us_events[0].tenant_id == "tenant_us_1"
        assert us_events[0].regional_context == "US"

    async def test_get_audit_events_by_category(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a recuperação de eventos de auditoria por categoria.
        """
        # Cria múltiplos eventos para diferentes categorias
        for event_key in ["br_login_success", "us_data_access", "eu_consent_update"]:
            await audit_service.create_audit_event(sample_audit_events[event_key])
        
        # Recupera eventos de autenticação
        auth_events = await audit_service.get_audit_events(
            category=AuditEventCategory.AUTHENTICATION, limit=10
        )
        assert len(auth_events) == 1
        assert auth_events[0].category == AuditEventCategory.AUTHENTICATION
        assert auth_events[0].action == "LOGIN"
        
        # Recupera eventos de acesso a dados
        data_events = await audit_service.get_audit_events(
            category=AuditEventCategory.DATA_ACCESS, limit=10
        )
        assert len(data_events) == 1
        assert data_events[0].category == AuditEventCategory.DATA_ACCESS
        assert data_events[0].action == "VIEW_FINANCIAL_REPORT"
        
        # Recupera eventos de consentimento
        consent_events = await audit_service.get_audit_events(
            category=AuditEventCategory.CONSENT, limit=10
        )
        assert len(consent_events) == 1
        assert consent_events[0].category == AuditEventCategory.CONSENT
        assert consent_events[0].action == "UPDATE_CONSENT"

    async def test_create_retention_policy(self, audit_service: AuditService):
        """
        Testa a criação de políticas de retenção de auditoria.
        """
        # Cria uma política de retenção para o tenant Brasil
        retention_policy = await audit_service.create_retention_policy(
            tenant_id="tenant_br_1",
            regional_context="BR",
            retention_days=730,  # 2 anos (LGPD)
            compliance_framework=ComplianceFramework.LGPD,
            category=AuditEventCategory.AUTHENTICATION,
            description="Política de retenção para logs de autenticação - LGPD",
            automatic_anonymization=True,
            anonymization_fields=["source_ip", "user_name"]
        )
        
        # Verifica se a política foi criada corretamente
        assert retention_policy is not None
        assert retention_policy.id is not None
        assert retention_policy.tenant_id == "tenant_br_1"
        assert retention_policy.regional_context == "BR"
        assert retention_policy.retention_days == 730
        assert retention_policy.compliance_framework == ComplianceFramework.LGPD
        assert retention_policy.automatic_anonymization is True
        assert "source_ip" in retention_policy.anonymization_fields
        assert "user_name" in retention_policy.anonymization_fields
        
        # Cria uma política de retenção para o tenant EUA
        retention_policy = await audit_service.create_retention_policy(
            tenant_id="tenant_us_1",
            regional_context="US",
            retention_days=2555,  # 7 anos (SOX)
            compliance_framework=ComplianceFramework.SOX,
            category=AuditEventCategory.FINANCIAL,
            description="Política de retenção para logs financeiros - SOX",
            automatic_anonymization=False,
            anonymization_fields=[]
        )
        
        # Verifica se a política foi criada corretamente
        assert retention_policy is not None
        assert retention_policy.id is not None
        assert retention_policy.tenant_id == "tenant_us_1"
        assert retention_policy.regional_context == "US"
        assert retention_policy.retention_days == 2555
        assert retention_policy.compliance_framework == ComplianceFramework.SOX
        assert retention_policy.automatic_anonymization is False
        assert len(retention_policy.anonymization_fields) == 0

    async def test_apply_retention_policy(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a aplicação de políticas de retenção de auditoria.
        """
        # Cria eventos de auditoria para teste
        br_event = await audit_service.create_audit_event(sample_audit_events["br_login_success"])
        
        # Cria uma política de retenção com anonimização
        retention_policy = await audit_service.create_retention_policy(
            tenant_id="tenant_br_1",
            regional_context="BR",
            retention_days=730,
            compliance_framework=ComplianceFramework.LGPD,
            category=AuditEventCategory.AUTHENTICATION,
            description="Política de retenção para logs de autenticação - LGPD",
            automatic_anonymization=True,
            anonymization_fields=["source_ip", "user_name"]
        )
        
        # Aplica a política de retenção manualmente
        await audit_service.apply_retention_policy(br_event.id, retention_policy.id)
        
        # Recupera o evento atualizado
        updated_event = await audit_service.get_audit_event_by_id(br_event.id)
        
        # Verifica se a anonimização foi aplicada
        assert updated_event is not None
        assert "source_ip" in updated_event.anonymized_fields
        assert "user_name" in updated_event.anonymized_fields
        assert updated_event.details.get("source_ip") == "[ANONYMIZED]"
        assert updated_event.details.get("user_name") == "[ANONYMIZED]"

    async def test_batch_processing(self, audit_service: AuditService, sample_audit_events):
        """
        Testa o processamento em lote de eventos de auditoria.
        """
        # Cria uma lista de eventos para processamento em lote
        events_to_create = [
            sample_audit_events["br_login_success"],
            sample_audit_events["us_data_access"],
            sample_audit_events["eu_consent_update"],
            sample_audit_events["ao_transaction"]
        ]
        
        # Processa os eventos em lote
        created_events = await audit_service.batch_create_audit_events(events_to_create)
        
        # Verifica se os eventos foram criados corretamente
        assert len(created_events) == 4
        
        # Verifica se cada evento tem seus respectivos contextos regionais e compliance
        regions = ["BR", "US", "EU", "AO"]
        frameworks = ["LGPD", "SOX", "GDPR", "BNA"]
        
        for i, event in enumerate(created_events):
            assert event.regional_context == regions[i]
            assert frameworks[i] in event.compliance_frameworks

    async def test_create_compliance_report(self, audit_service: AuditService, sample_audit_events):
        """
        Testa a geração de relatório de compliance.
        """
        # Cria eventos para diferentes contextos regionais
        for event_key in ["br_login_success", "us_data_access", "eu_consent_update", "ao_transaction"]:
            await audit_service.create_audit_event(sample_audit_events[event_key])
        
        # Gera relatório de compliance para Brasil - LGPD
        report = await audit_service.create_compliance_report(
            tenant_id="tenant_br_1",
            regional_context="BR",
            compliance_framework=ComplianceFramework.LGPD,
            start_date=datetime.utcnow() - timedelta(days=7),
            end_date=datetime.utcnow(),
            report_name="Relatório Semanal LGPD",
            report_description="Relatório de auditoria semanal para compliance LGPD"
        )
        
        # Verifica se o relatório foi gerado corretamente
        assert report is not None
        assert report.id is not None
        assert report.tenant_id == "tenant_br_1"
        assert report.regional_context == "BR"
        assert report.compliance_framework == ComplianceFramework.LGPD
        assert report.event_count > 0
        assert report.report_name == "Relatório Semanal LGPD"
        assert report.status == "COMPLETED"
        
        # Gera relatório de compliance para EUA - SOX
        report = await audit_service.create_compliance_report(
            tenant_id="tenant_us_1",
            regional_context="US",
            compliance_framework=ComplianceFramework.SOX,
            start_date=datetime.utcnow() - timedelta(days=30),
            end_date=datetime.utcnow(),
            report_name="Monthly SOX Audit Report",
            report_description="Monthly audit report for SOX compliance"
        )
        
        # Verifica se o relatório foi gerado corretamente
        assert report is not None
        assert report.id is not None
        assert report.tenant_id == "tenant_us_1"
        assert report.regional_context == "US"
        assert report.compliance_framework == ComplianceFramework.SOX
        assert report.event_count > 0
        assert report.report_name == "Monthly SOX Audit Report"
        assert report.status == "COMPLETED"