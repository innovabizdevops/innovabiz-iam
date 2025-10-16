"""
Testes de integração para os endpoints de auditoria.

Este módulo implementa testes de integração completos para validar
o comportamento dos endpoints REST da API de auditoria, verificando
isolamento multi-tenant, contexto regional e autorização.
"""
import json
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from sqlalchemy.orm import sessionmaker

from app.api import deps
from app.db.base import Base
from app.models.audit_models import (
    AuditEventCategory,
    AuditEventSeverity,
    ComplianceFramework,
    RetentionPolicyType
)
from app.routers.audit_router import router as audit_router

# Configuração do banco de dados de teste
TEST_SQLALCHEMY_DATABASE_URL = "postgresql+asyncpg://test:test@localhost/test_innovabiz_iam"


# Fixtures para os testes
@pytest.fixture
async def db_engine():
    """Cria engine para banco de dados de teste."""
    engine = create_async_engine(
        TEST_SQLALCHEMY_DATABASE_URL,
        echo=False,
        future=True
    )
    
    # Recria todas as tabelas para cada teste
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
        await conn.run_sync(Base.metadata.create_all)
    
    yield engine
    
    # Cleanup após os testes
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
    
    await engine.dispose()


@pytest.fixture
async def db_session(db_engine):
    """Cria uma sessão de banco de dados para teste."""
    async_session = sessionmaker(
        db_engine,
        class_=AsyncSession,
        expire_on_commit=False,
        autoflush=False
    )

    async with async_session() as session:
        yield session


@pytest.fixture
def app(db_session):
    """Configura a aplicação FastAPI para testes."""
    app = FastAPI()
    
    # Override da dependência para usar sessão de teste
    async def get_db_override():
        yield db_session
    
    # Override da dependência para contexto multi-tenant
    async def get_tenant_id_override():
        return "tenant1"
    
    # Override da dependência para contexto regional
    async def get_regional_context_override():
        return "BR"
    
    app.dependency_overrides[deps.get_db] = get_db_override
    app.dependency_overrides[deps.get_tenant_id] = get_tenant_id_override
    app.dependency_overrides[deps.get_regional_context] = get_regional_context_override
    
    app.include_router(audit_router, prefix="/audit")
    
    return app


@pytest.fixture
def client(app):
    """Retorna um cliente de teste para a API."""
    return TestClient(app)


@pytest.fixture
def valid_audit_event():
    """Retorna um evento de auditoria válido para os testes."""
    return {
        "category": "USER_MANAGEMENT",
        "action": "CREATE_USER",
        "description": "Criação de novo usuário",
        "resource_type": "USER",
        "resource_id": "user-123",
        "resource_name": "john.doe",
        "severity": "INFO",
        "success": True,
        "details": {"role": "admin", "department": "IT"},
        "tags": ["user", "creation"],
        "tenant_id": "tenant1",  # Será sobrescrito pelo contexto
        "regional_context": "BR",  # Será sobrescrito pelo contexto
        "country_code": "BR",
        "language": "pt-BR",
        "user_id": "admin-456",
        "user_name": "Admin User",
        "correlation_id": "corr-789",
        "source_ip": "192.168.1.1",
        "http_details": {
            "method": "POST",
            "url": "/api/users",
            "status_code": 201,
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123-abc",
            "path": "/api/users",
            "query_params": {"tenant": "tenant1"}
        }
    }


@pytest.fixture
def valid_retention_policy():
    """Retorna uma política de retenção válida para os testes."""
    return {
        "tenant_id": "tenant1",  # Será sobrescrito pelo contexto
        "regional_context": "BR",  # Será sobrescrito pelo contexto
        "compliance_framework": "LGPD",
        "retention_period_days": 730,  # 2 anos
        "policy_type": "ANONYMIZATION",
        "fields_to_anonymize": ["user_id", "user_name", "source_ip"],
        "description": "Política LGPD para anonimização de dados pessoais",
        "active": True
    }


# Testes de endpoints
@pytest.mark.asyncio
async def test_create_audit_event(client, valid_audit_event):
    """
    Testa a criação de um evento de auditoria via API.
    
    Verifica se o evento é criado corretamente e se o ID é retornado.
    """
    # Act
    response = client.post("/audit/events", json=valid_audit_event)
    
    # Assert
    assert response.status_code == 201
    assert "id" in response.json()
    assert "tenant_id" in response.json()
    assert "regional_context" in response.json()
    assert response.json()["tenant_id"] == "tenant1"
    assert response.json()["regional_context"] == "BR"


@pytest.mark.asyncio
async def test_create_event_batch(client, valid_audit_event):
    """
    Testa a criação de eventos em lote via API.
    
    Verifica se múltiplos eventos são processados corretamente.
    """
    # Arrange
    batch_events = {
        "events": [valid_audit_event, valid_audit_event]
    }
    
    # Act
    response = client.post("/audit/events/batch", json=batch_events)
    
    # Assert
    assert response.status_code == 201
    assert "count" in response.json()
    assert response.json()["count"] == 2


@pytest.mark.asyncio
async def test_get_audit_event(client, valid_audit_event):
    """
    Testa a obtenção de um evento específico por ID.
    
    Verifica se o evento pode ser recuperado após a criação.
    """
    # Arrange - Cria um evento primeiro
    create_response = client.post("/audit/events", json=valid_audit_event)
    assert create_response.status_code == 201
    event_id = create_response.json()["id"]
    
    # Act - Tenta recuperar o evento criado
    response = client.get(f"/audit/events/{event_id}")
    
    # Assert
    assert response.status_code == 200
    assert response.json()["id"] == event_id
    assert response.json()["action"] == valid_audit_event["action"]
    assert response.json()["description"] == valid_audit_event["description"]


@pytest.mark.asyncio
async def test_list_audit_events(client, valid_audit_event):
    """
    Testa a listagem de eventos com filtros via API.
    
    Verifica se a paginação e filtragem funcionam corretamente.
    """
    # Arrange - Cria múltiplos eventos
    for _ in range(3):
        client.post("/audit/events", json=valid_audit_event)
    
    # Act - Lista eventos com filtro
    response = client.get(
        "/audit/events",
        params={
            "category": "USER_MANAGEMENT",
            "limit": 2,
            "offset": 0
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert "items" in response.json()
    assert "total" in response.json()
    assert "limit" in response.json()
    assert "offset" in response.json()
    assert response.json()["total"] >= 3  # Deve ter ao menos os 3 que criamos
    assert len(response.json()["items"]) == 2  # Limite respeitado
    
    # Act - Segunda página
    response = client.get(
        "/audit/events",
        params={
            "category": "USER_MANAGEMENT",
            "limit": 2,
            "offset": 2
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert len(response.json()["items"]) >= 1  # Ao menos 1 na segunda página


@pytest.mark.asyncio
async def test_create_retention_policy(client, valid_retention_policy):
    """
    Testa a criação de política de retenção via API.
    
    Verifica se a política é criada corretamente e se o ID é retornado.
    """
    # Act
    response = client.post("/audit/retention-policies", json=valid_retention_policy)
    
    # Assert
    assert response.status_code == 201
    assert "id" in response.json()
    assert response.json()["tenant_id"] == "tenant1"
    assert response.json()["regional_context"] == "BR"
    assert response.json()["compliance_framework"] == "LGPD"
    assert response.json()["retention_period_days"] == 730


@pytest.mark.asyncio
async def test_get_retention_policy(client, valid_retention_policy):
    """
    Testa a obtenção de uma política de retenção específica por ID.
    
    Verifica se a política pode ser recuperada após a criação.
    """
    # Arrange - Cria uma política primeiro
    create_response = client.post("/audit/retention-policies", json=valid_retention_policy)
    assert create_response.status_code == 201
    policy_id = create_response.json()["id"]
    
    # Act - Tenta recuperar a política criada
    response = client.get(f"/audit/retention-policies/{policy_id}")
    
    # Assert
    assert response.status_code == 200
    assert response.json()["id"] == policy_id
    assert response.json()["compliance_framework"] == valid_retention_policy["compliance_framework"]
    assert response.json()["retention_period_days"] == valid_retention_policy["retention_period_days"]


@pytest.mark.asyncio
async def test_update_retention_policy(client, valid_retention_policy):
    """
    Testa a atualização de uma política de retenção via API.
    
    Verifica se os campos são atualizados corretamente.
    """
    # Arrange - Cria uma política primeiro
    create_response = client.post("/audit/retention-policies", json=valid_retention_policy)
    assert create_response.status_code == 201
    policy_id = create_response.json()["id"]
    
    # Prepara dados de atualização
    update_data = {
        "retention_period_days": 1095,  # Altera para 3 anos
        "description": "Política LGPD atualizada para anonimização de dados pessoais"
    }
    
    # Act - Atualiza a política
    response = client.patch(f"/audit/retention-policies/{policy_id}", json=update_data)
    
    # Assert
    assert response.status_code == 200
    assert response.json()["retention_period_days"] == 1095
    assert response.json()["description"] == update_data["description"]


@pytest.mark.asyncio
async def test_list_retention_policies(client, valid_retention_policy):
    """
    Testa a listagem de políticas de retenção via API.
    
    Verifica se a paginação e filtragem funcionam corretamente.
    """
    # Arrange - Cria múltiplas políticas
    client.post("/audit/retention-policies", json=valid_retention_policy)
    
    # Cria política para outro framework
    policy2 = valid_retention_policy.copy()
    policy2["compliance_framework"] = "ISO_27001"
    client.post("/audit/retention-policies", json=policy2)
    
    # Act - Lista políticas com filtro
    response = client.get(
        "/audit/retention-policies",
        params={
            "compliance_framework": "LGPD",
            "active": True
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert "items" in response.json()
    assert "total" in response.json()
    assert len(response.json()["items"]) >= 1
    assert response.json()["items"][0]["compliance_framework"] == "LGPD"


@pytest.mark.asyncio
async def test_apply_retention_policies(client, valid_retention_policy):
    """
    Testa a aplicação de políticas de retenção via API.
    
    Verifica se as políticas são aplicadas corretamente.
    """
    # Arrange - Cria uma política
    client.post("/audit/retention-policies", json=valid_retention_policy)
    
    # Act - Aplica as políticas em dry run
    response = client.post(
        "/audit/apply-retention",
        json={
            "dry_run": True,
            "batch_size": 50
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert "events_processed" in response.json()
    assert "events_anonymized" in response.json()
    assert "events_deleted" in response.json()
    assert "processing_time_ms" in response.json()


@pytest.mark.asyncio
async def test_mask_sensitive_fields(client, valid_audit_event):
    """
    Testa o mascaramento de campos sensíveis via API.
    
    Verifica se os campos são mascarados corretamente.
    """
    # Arrange - Cria um evento primeiro
    create_response = client.post("/audit/events", json=valid_audit_event)
    assert create_response.status_code == 201
    event_id = create_response.json()["id"]
    
    # Act - Mascara campos sensíveis
    response = client.post(
        f"/audit/events/{event_id}/mask",
        json={
            "fields": ["source_ip", "user_name", "details.credit_card"]
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert response.json()["success"] is True
    
    # Verifica se os campos foram mascarados
    get_response = client.get(f"/audit/events/{event_id}")
    assert get_response.status_code == 200
    assert "masked_fields" in get_response.json()
    assert "source_ip" in get_response.json()["masked_fields"]
    assert "user_name" in get_response.json()["masked_fields"]


@pytest.mark.asyncio
async def test_generate_compliance_report(client, valid_audit_event, valid_retention_policy):
    """
    Testa a geração de relatórios de compliance via API.
    
    Verifica se o relatório é gerado corretamente.
    """
    # Arrange - Cria eventos e políticas
    client.post("/audit/events", json=valid_audit_event)
    client.post("/audit/retention-policies", json=valid_retention_policy)
    
    # Act - Gera relatório de compliance
    report_request = {
        "compliance_framework": "LGPD",
        "start_date": (datetime.now() - timedelta(days=30)).isoformat(),
        "end_date": datetime.now().isoformat(),
        "report_format": "json",
        "include_anonymized": False
    }
    
    response = client.post("/audit/compliance-reports", json=report_request)
    
    # Assert
    assert response.status_code == 201
    assert "id" in response.json()
    assert response.json()["compliance_framework"] == "LGPD"
    assert response.json()["status"] in ["PENDING", "PROCESSING", "COMPLETED"]


@pytest.mark.asyncio
async def test_get_audit_statistics(client, valid_audit_event):
    """
    Testa a obtenção de estatísticas de auditoria via API.
    
    Verifica se as estatísticas são geradas corretamente.
    """
    # Arrange - Cria eventos
    for _ in range(5):
        client.post("/audit/events", json=valid_audit_event)
    
    # Act - Solicita estatísticas
    response = client.get(
        "/audit/statistics",
        params={
            "period": "daily",
            "start_date": (datetime.now() - timedelta(days=7)).date().isoformat(),
            "end_date": datetime.now().date().isoformat()
        }
    )
    
    # Assert
    assert response.status_code == 200
    assert "statistics" in response.json()
    assert "tenant_id" in response.json()
    assert "regional_context" in response.json()
    assert "period" in response.json()
    assert response.json()["period"] == "daily"