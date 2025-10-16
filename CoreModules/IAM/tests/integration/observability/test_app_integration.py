"""
Testes de integração para verificar a observabilidade integrada à aplicação principal.

Este módulo valida a integração completa do módulo de observabilidade com uma aplicação
FastAPI, garantindo que métricas, health checks e endpoints diagnósticos funcionem
corretamente no contexto de uma aplicação real com multi-tenant e multi-região.

Author: INNOVABIZ DevOps Team
Date: 2025-07-31
"""

import asyncio
import json
import os
import time
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from fastapi import Depends, FastAPI, Header, Request, Response
from fastapi.testclient import TestClient
from prometheus_client import REGISTRY
from pydantic import BaseModel

from api.app.integrations.observability import ObservabilityIntegration, setup_observability
from api.app.metrics.audit_metrics import (
    audit_event_processing, retention_policy_application, compliance_check,
    track_audit_event, register_retention_policy, set_compliance_status
)


# Modelos para testes
class AuditEvent(BaseModel):
    """Modelo de evento de auditoria para testes."""
    event_id: str
    event_type: str
    entity_id: str
    user_id: str
    timestamp: int
    data: dict


class RetentionPolicy(BaseModel):
    """Modelo de política de retenção para testes."""
    policy_id: str
    policy_type: str
    retention_days: int
    applies_to: list[str]
    enabled: bool = True


# Funções auxiliares para extrair contexto
async def get_tenant_context(
    x_tenant_id: str = Header(None),
    x_region: str = Header(None),
    x_environment: str = Header("production"),
    x_correlation_id: str = Header(None)
):
    """Extrai o contexto multi-tenant das headers HTTP."""
    return {
        "tenant_id": x_tenant_id or "default",
        "region": x_region or "br-east",
        "environment": x_environment,
        "correlation_id": x_correlation_id or "test-correlation"
    }


# Fixture para criar uma aplicação de teste
@pytest.fixture
def test_app():
    """Cria uma aplicação FastAPI com observabilidade configurada para teste."""
    # Mock clients
    db_client = AsyncMock()
    redis_client = AsyncMock()
    kafka_client = AsyncMock()
    storage_client = AsyncMock()
    
    # Configurar mocks
    db_client.is_healthy.return_value = True
    redis_client.is_healthy.return_value = True
    kafka_client.is_healthy.return_value = True
    storage_client.is_healthy.return_value = True
    
    # Service mock
    audit_service = MagicMock()
    audit_service.process_event = AsyncMock()
    audit_service.apply_retention_policy = AsyncMock()
    audit_service.verify_compliance = AsyncMock()
    
    # Criar app
    app = FastAPI(title="IAM Audit Test App", version="1.0.0")
    
    # Configurar observabilidade
    obs = ObservabilityIntegration(
        app=app,
        service_name="iam-audit-integration-test",
        service_version="1.0.0",
        build_id="integration-test",
        commit_hash="abcdef123456",
        environment="test",
        region="global",
        db_client=db_client,
        redis_client=redis_client,
        kafka_client=kafka_client,
        storage_client=storage_client
    )
    
    obs.setup_metrics()
    obs.setup_metrics_endpoint()
    obs.setup_health_check_endpoint()
    obs.setup_diagnostic_endpoint()
    obs.setup_health_verification_methods()
    obs.setup_lifecycle_handlers()
    
    # Rotas para teste
    @app.post("/api/v1/audit/events")
    @audit_event_processing
    async def create_audit_event(
        event: AuditEvent,
        request: Request,
        context: dict = Depends(get_tenant_context)
    ):
        # Simular processamento
        await asyncio.sleep(0.1)
        
        # Rastreamento de evento
        track_audit_event(
            event_type=event.event_type,
            tenant_id=context["tenant_id"],
            region=context["region"],
            environment=context["environment"]
        )
        
        # Chamar serviço mock
        await audit_service.process_event(event, context)
        
        return {"status": "success", "event_id": event.event_id}
    
    @app.put("/api/v1/audit/retention/{policy_id}")
    @retention_policy_application
    async def apply_retention_policy(
        policy_id: str,
        policy: RetentionPolicy,
        request: Request,
        context: dict = Depends(get_tenant_context)
    ):
        # Registrar política
        register_retention_policy(
            policy_type=policy.policy_type,
            tenant_id=context["tenant_id"],
            region=context["region"],
            environment=context["environment"]
        )
        
        # Simular aplicação
        await asyncio.sleep(0.2)
        
        # Chamar serviço mock
        await audit_service.apply_retention_policy(policy_id, policy, context)
        
        return {"status": "success", "applied": True, "policy_id": policy_id}
    
    @app.get("/api/v1/audit/compliance/{framework}")
    @compliance_check
    async def check_compliance(
        framework: str,
        request: Request,
        context: dict = Depends(get_tenant_context)
    ):
        # Simular verificação
        await asyncio.sleep(0.1)
        
        # Processar verificação
        result = await audit_service.verify_compliance(framework, context)
        
        # Definir status (simulado como positivo)
        compliance_status = True
        
        # Atualizar métrica
        set_compliance_status(
            framework=framework,
            status=compliance_status,
            tenant_id=context["tenant_id"],
            region=context["region"],
            environment=context["environment"]
        )
        
        return {
            "status": "compliant" if compliance_status else "non-compliant",
            "framework": framework,
            "timestamp": int(time.time())
        }
    
    return {
        "app": app,
        "client": TestClient(app),
        "audit_service": audit_service,
        "db_client": db_client,
        "redis_client": redis_client
    }


# Limpar o registro do Prometheus entre testes
@pytest.fixture(autouse=True)
def clear_prometheus_registry():
    """Limpa o registro do Prometheus entre testes."""
    for name in list(REGISTRY._names_to_collectors.keys()):
        if name.startswith('audit_') or name.startswith('http_'):
            REGISTRY.unregister(REGISTRY._names_to_collectors[name])
    yield


@pytest.mark.integration
class TestObservabilityAppIntegration:
    """Testes de integração para o módulo de observabilidade."""
    
    def test_metrics_endpoint(self, test_app):
        """Verifica se o endpoint de métricas está funcionando na aplicação."""
        # Act
        response = test_app["client"].get("/metrics")
        
        # Assert
        assert response.status_code == 200
        assert "application/openmetrics-text" in response.headers["content-type"]
        assert "# HELP audit_service_info" in response.text
    
    def test_health_endpoint(self, test_app):
        """Verifica se o endpoint de saúde está funcionando na aplicação."""
        # Act
        response = test_app["client"].get("/health")
        
        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "iam-audit-integration-test"
        assert len(data["checks"]) == 4  # DB, Redis, Kafka, Storage
    
    def test_diagnostic_endpoint(self, test_app):
        """Verifica se o endpoint de diagnóstico está funcionando na aplicação."""
        # Act
        response = test_app["client"].get("/diagnostic")
        
        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "iam-audit-integration-test"
        assert data["version"] == "1.0.0"
        assert data["build"] == "integration-test"
        assert data["commit"] == "abcdef123456"
        assert data["environment"] == "test"
    
    def test_audit_event_processing_decorator(self, test_app):
        """Verifica se o decorador de processamento de eventos está funcionando."""
        # Arrange
        event = {
            "event_id": "evt-12345",
            "event_type": "user.login",
            "entity_id": "user-1",
            "user_id": "admin-1",
            "timestamp": int(time.time()),
            "data": {"ip": "192.168.1.1", "user_agent": "Mozilla/5.0"}
        }
        
        # Act
        response = test_app["client"].post(
            "/api/v1/audit/events",
            json=event,
            headers={
                "X-Tenant-ID": "tenant-abc",
                "X-Region": "br-south",
                "X-Environment": "production"
            }
        )
        
        # Assert
        assert response.status_code == 200
        assert response.json()["status"] == "success"
        
        # Verificar se o serviço foi chamado com o contexto correto
        test_app["audit_service"].process_event.assert_called_once()
        context_arg = test_app["audit_service"].process_event.call_args[0][1]
        assert context_arg["tenant_id"] == "tenant-abc"
        assert context_arg["region"] == "br-south"
        assert context_arg["environment"] == "production"
    
    def test_retention_policy_decorator(self, test_app):
        """Verifica se o decorador de política de retenção está funcionando."""
        # Arrange
        policy = {
            "policy_id": "pol-12345",
            "policy_type": "legal_hold",
            "retention_days": 365,
            "applies_to": ["user.login", "data.access"],
            "enabled": True
        }
        
        # Act
        response = test_app["client"].put(
            "/api/v1/audit/retention/pol-12345",
            json=policy,
            headers={
                "X-Tenant-ID": "tenant-xyz",
                "X-Region": "eu-central",
                "X-Environment": "staging"
            }
        )
        
        # Assert
        assert response.status_code == 200
        assert response.json()["status"] == "success"
        assert response.json()["applied"] is True
        
        # Verificar se o serviço foi chamado com o contexto correto
        test_app["audit_service"].apply_retention_policy.assert_called_once()
        context_arg = test_app["audit_service"].apply_retention_policy.call_args[0][2]
        assert context_arg["tenant_id"] == "tenant-xyz"
        assert context_arg["region"] == "eu-central"
        assert context_arg["environment"] == "staging"
    
    def test_compliance_check_decorator(self, test_app):
        """Verifica se o decorador de verificação de compliance está funcionando."""
        # Act
        response = test_app["client"].get(
            "/api/v1/audit/compliance/LGPD",
            headers={
                "X-Tenant-ID": "tenant-123",
                "X-Region": "br-east",
                "X-Environment": "production",
                "X-Correlation-ID": "corr-abc-123"
            }
        )
        
        # Assert
        assert response.status_code == 200
        assert response.json()["status"] == "compliant"
        assert response.json()["framework"] == "LGPD"
        
        # Verificar se o serviço foi chamado com o contexto correto
        test_app["audit_service"].verify_compliance.assert_called_once()
        context_arg = test_app["audit_service"].verify_compliance.call_args[0][1]
        assert context_arg["tenant_id"] == "tenant-123"
        assert context_arg["region"] == "br-east"
        assert context_arg["environment"] == "production"
        assert context_arg["correlation_id"] == "corr-abc-123"
    
    def test_health_check_with_db_failure(self, test_app):
        """Verifica o comportamento do health check quando o banco de dados falha."""
        # Arrange - simular falha no banco de dados
        test_app["db_client"].is_healthy.return_value = False
        
        # Act
        response = test_app["client"].get("/health")
        
        # Assert
        assert response.status_code == 503
        data = response.json()
        assert data["status"] == "unhealthy"
        db_check = next(check for check in data["checks"] if check["name"] == "database")
        assert db_check["status"] is False
    
    def test_multi_tenant_metric_isolation(self, test_app):
        """Verifica se as métricas são isoladas corretamente por tenant."""
        # Arrange
        event = {
            "event_id": "evt-tenant-a",
            "event_type": "user.login",
            "entity_id": "user-1",
            "user_id": "admin-1",
            "timestamp": int(time.time()),
            "data": {"ip": "192.168.1.1"}
        }
        
        # Act - enviar mesmo evento para dois tenants diferentes
        test_app["client"].post(
            "/api/v1/audit/events",
            json=event,
            headers={
                "X-Tenant-ID": "tenant-a",
                "X-Region": "us-east"
            }
        )
        
        event["event_id"] = "evt-tenant-b"
        test_app["client"].post(
            "/api/v1/audit/events",
            json=event,
            headers={
                "X-Tenant-ID": "tenant-b",
                "X-Region": "us-east"
            }
        )
        
        # Verificar métricas via endpoint
        response = test_app["client"].get("/metrics")
        metrics_output = response.text
        
        # Assert
        # Verificar que as métricas estão sendo registradas separadamente por tenant
        assert 'tenant_id="tenant-a"' in metrics_output
        assert 'tenant_id="tenant-b"' in metrics_output
        assert 'event_type="user.login"' in metrics_output
    
    def test_multi_region_metrics(self, test_app):
        """Verifica se as métricas são corretamente registradas por região."""
        # Arrange
        policy = {
            "policy_id": "pol-region-test",
            "policy_type": "data_privacy",
            "retention_days": 180,
            "applies_to": ["pii.access"],
            "enabled": True
        }
        
        # Act - aplicar mesma política em duas regiões diferentes
        test_app["client"].put(
            "/api/v1/audit/retention/pol-region-test",
            json=policy,
            headers={
                "X-Tenant-ID": "global-tenant",
                "X-Region": "eu-west"
            }
        )
        
        test_app["client"].put(
            "/api/v1/audit/retention/pol-region-test",
            json=policy,
            headers={
                "X-Tenant-ID": "global-tenant",
                "X-Region": "ap-south"
            }
        )
        
        # Verificar métricas via endpoint
        response = test_app["client"].get("/metrics")
        metrics_output = response.text
        
        # Assert
        # Verificar que as métricas estão sendo registradas separadamente por região
        assert 'region="eu-west"' in metrics_output
        assert 'region="ap-south"' in metrics_output
        assert 'policy_type="data_privacy"' in metrics_output


if __name__ == "__main__":
    pytest.main(["--verbose", "test_app_integration.py"])