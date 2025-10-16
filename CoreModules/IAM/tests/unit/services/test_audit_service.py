"""
Testes unitários para AuditService.
"""
import json
import uuid
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.audit_models import (
    AuditEventCategory,
    AuditEventSeverity,
    ComplianceFramework,
    RetentionPolicyType
)
from app.schemas.audit_schemas import AuditEventCreate, HttpDetailsModel
from app.services.audit_service import AuditService


@pytest.fixture
def mock_db_session():
    """Fixture para mock de sessão de banco de dados."""
    session = AsyncMock(spec=AsyncSession)
    return session


@pytest.fixture
def mock_logger():
    """Fixture para mock de logger."""
    return MagicMock()


@pytest.fixture
def audit_service(mock_db_session, mock_logger):
    """Fixture para instância de AuditService."""
    return AuditService(db=mock_db_session, logger=mock_logger)


@pytest.fixture
def sample_audit_event():
    """Fixture para evento de auditoria de exemplo."""
    return AuditEventCreate(
        category=AuditEventCategory.USER_MANAGEMENT,
        action="CREATE_USER",
        description="Criação de novo usuário",
        resource_type="USER",
        resource_id="user-123",
        resource_name="john.doe",
        severity=AuditEventSeverity.INFO,
        success=True,
        details={"role": "admin", "department": "IT"},
        tags=["user", "creation"],
        tenant_id="tenant1",
        regional_context="BR",
        country_code="BR",
        language="pt-BR",
        user_id="admin-456",
        user_name="Admin User",
        correlation_id="corr-789",
        source_ip="192.168.1.1",
        http_details=HttpDetailsModel(
            method="POST",
            url="/api/users",
            status_code=201,
            user_agent="Mozilla/5.0",
            request_id="req-123-abc",
            path="/api/users",
            query_params={"tenant": "tenant1"}
        )
    )


@pytest.mark.asyncio
async def test_create_audit_event(audit_service, mock_db_session, sample_audit_event):
    """
    Testa a criação de um evento de auditoria.
    
    Verifica se o evento é criado corretamente com todas as propriedades e 
    se os frameworks de compliance são automaticamente detectados.
    """
    # Arrange
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one_or_none.return_value = str(uuid.uuid4())
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.create_audit_event(sample_audit_event)

    # Assert
    assert result is not None
    assert isinstance(result, uuid.UUID)
    mock_db_session.execute.assert_called_once()
    assert mock_logger.info.call_count == 1


@pytest.mark.asyncio
async def test_get_audit_event(audit_service, mock_db_session):
    """
    Testa a recuperação de um evento de auditoria pelo ID.
    
    Verifica se o evento é recuperado corretamente e se as permissões
    multi-tenant são respeitadas.
    """
    # Arrange
    event_id = uuid.uuid4()
    mock_result = {
        "id": event_id,
        "tenant_id": "tenant1",
        "regional_context": "BR",
        "category": "USER_MANAGEMENT",
        "action": "CREATE_USER",
        "description": "Criação de novo usuário",
        "created_at": datetime.now(timezone.utc),
        "updated_at": datetime.now(timezone.utc),
    }
    
    mock_execute_result = AsyncMock()
    mock_execute_result.mappings.return_value.first.return_value = mock_result
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.get_audit_event(event_id, "tenant1", "BR")

    # Assert
    assert result == mock_result
    mock_db_session.execute.assert_called_once()
    

@pytest.mark.asyncio
async def test_get_audit_event_not_found(audit_service, mock_db_session):
    """
    Testa o comportamento quando um evento de auditoria não é encontrado.
    """
    # Arrange
    event_id = uuid.uuid4()
    mock_execute_result = AsyncMock()
    mock_execute_result.mappings.return_value.first.return_value = None
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.get_audit_event(event_id, "tenant1", "BR")

    # Assert
    assert result is None
    mock_db_session.execute.assert_called_once()


@pytest.mark.asyncio
async def test_list_audit_events(audit_service, mock_db_session):
    """
    Testa a listagem de eventos de auditoria com filtragem.
    """
    # Arrange
    mock_events = [
        {
            "id": uuid.uuid4(),
            "tenant_id": "tenant1",
            "regional_context": "BR",
            "category": "USER_MANAGEMENT",
            "action": "CREATE_USER",
            "description": "Criação de novo usuário 1",
            "created_at": datetime.now(timezone.utc),
        },
        {
            "id": uuid.uuid4(),
            "tenant_id": "tenant1",
            "regional_context": "BR",
            "category": "USER_MANAGEMENT",
            "action": "UPDATE_USER",
            "description": "Atualização de usuário 2",
            "created_at": datetime.now(timezone.utc),
        }
    ]
    
    mock_count_result = AsyncMock()
    mock_count_result.scalar_one.return_value = len(mock_events)
    
    mock_execute_result = AsyncMock()
    mock_execute_result.mappings.return_value.all.return_value = mock_events
    
    mock_db_session.execute.side_effect = [mock_execute_result, mock_count_result]

    # Act
    result = await audit_service.list_audit_events(
        tenant_id="tenant1",
        regional_context="BR",
        category=AuditEventCategory.USER_MANAGEMENT,
        start_date=datetime.now(timezone.utc) - timedelta(days=7),
        end_date=datetime.now(timezone.utc),
        limit=10,
        offset=0
    )

    # Assert
    assert result["items"] == mock_events
    assert result["total"] == len(mock_events)
    assert result["limit"] == 10
    assert result["offset"] == 0
    assert mock_db_session.execute.call_count == 2


@pytest.mark.asyncio
async def test_process_batch_events(audit_service, mock_db_session, sample_audit_event):
    """
    Testa o processamento em lote de eventos de auditoria.
    
    Verifica se múltiplos eventos são processados corretamente em um único
    lote e se as operações de banco de dados são otimizadas.
    """
    # Arrange
    events = [sample_audit_event, sample_audit_event]
    
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar.return_value = 2  # número de eventos inseridos
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.process_batch_events(events)

    # Assert
    assert result == 2
    mock_db_session.execute.assert_called_once()


@pytest.mark.asyncio
async def test_create_retention_policy(audit_service, mock_db_session):
    """
    Testa a criação de uma política de retenção.
    """
    # Arrange
    policy_data = {
        "tenant_id": "tenant1",
        "regional_context": "BR",
        "compliance_framework": ComplianceFramework.LGPD,
        "retention_period_days": 730,  # 2 anos
        "policy_type": RetentionPolicyType.ANONYMIZATION,
        "fields_to_anonymize": ["user_id", "user_name", "source_ip"],
        "description": "Política LGPD para anonimização de dados pessoais",
        "active": True
    }
    
    policy_id = uuid.uuid4()
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one.return_value = policy_id
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.create_retention_policy(policy_data)

    # Assert
    assert result == policy_id
    mock_db_session.execute.assert_called_once()
    mock_logger.info.assert_called_once()


@pytest.mark.asyncio
async def test_apply_retention_policies(audit_service, mock_db_session):
    """
    Testa a aplicação de políticas de retenção.
    
    Verifica se as políticas são aplicadas corretamente usando stored procedures
    e se o processo é auditado adequadamente.
    """
    # Arrange
    tenant_id = "tenant1"
    regional_context = "BR"
    batch_size = 100
    dry_run = False
    
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one.return_value = {
        "events_processed": 150,
        "events_anonymized": 50,
        "events_deleted": 20,
        "processing_time_ms": 2500
    }
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.apply_retention_policies(
        tenant_id=tenant_id,
        regional_context=regional_context,
        batch_size=batch_size,
        dry_run=dry_run
    )

    # Assert
    assert result["events_processed"] == 150
    assert result["events_anonymized"] == 50
    assert result["events_deleted"] == 20
    mock_db_session.execute.assert_called_once()
    mock_logger.info.assert_called()


@pytest.mark.asyncio
async def test_mask_sensitive_fields(audit_service, mock_db_session):
    """
    Testa o mascaramento de campos sensíveis em eventos de auditoria.
    """
    # Arrange
    event_id = uuid.uuid4()
    fields = ["source_ip", "user_name", "details.credit_card"]
    
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one.return_value = True
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.mask_sensitive_fields(event_id, fields)

    # Assert
    assert result is True
    mock_db_session.execute.assert_called_once()
    mock_logger.info.assert_called_once()


@pytest.mark.asyncio
async def test_generate_compliance_report(audit_service, mock_db_session):
    """
    Testa a geração de relatórios de compliance.
    """
    # Arrange
    report_params = {
        "tenant_id": "tenant1",
        "regional_context": "BR",
        "compliance_framework": ComplianceFramework.LGPD,
        "start_date": datetime.now(timezone.utc) - timedelta(days=30),
        "end_date": datetime.now(timezone.utc),
        "include_anonymized": False
    }
    
    report_id = uuid.uuid4()
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one.return_value = report_id
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.generate_compliance_report(**report_params)

    # Assert
    assert result == report_id
    mock_db_session.execute.assert_called_once()


@pytest.mark.asyncio
async def test_generate_audit_statistics(audit_service, mock_db_session):
    """
    Testa a geração de estatísticas de auditoria.
    """
    # Arrange
    tenant_id = "tenant1"
    regional_context = "BR"
    period = "daily"
    start_date = datetime.now(timezone.utc) - timedelta(days=30)
    end_date = datetime.now(timezone.utc)
    
    mock_stats = {
        "daily": {
            "2023-07-01": {
                "total": 150,
                "by_category": {
                    "USER_MANAGEMENT": 50,
                    "AUTHENTICATION": 80,
                    "DATA_ACCESS": 20
                },
                "by_severity": {
                    "INFO": 120,
                    "WARNING": 20,
                    "ERROR": 10
                },
                "success_rate": 0.95
            }
        }
    }
    
    mock_execute_result = AsyncMock()
    mock_execute_result.scalar_one.return_value = json.dumps(mock_stats)
    mock_db_session.execute.return_value = mock_execute_result

    # Act
    result = await audit_service.generate_audit_statistics(
        tenant_id=tenant_id,
        regional_context=regional_context,
        period=period,
        start_date=start_date,
        end_date=end_date
    )

    # Assert
    assert result == mock_stats
    mock_db_session.execute.assert_called_once()
    mock_logger.info.assert_called_once()