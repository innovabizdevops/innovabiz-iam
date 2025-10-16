"""
Testes unitários para o módulo de métricas de auditoria.

Valida o funcionamento dos decoradores de instrumentação e métricas
para o serviço de auditoria multi-contexto do IAM.
"""
import pytest
import asyncio
from unittest.mock import patch, MagicMock
from prometheus_client import REGISTRY

# Importar componentes a serem testados
from src.api.app.metrics.audit_metrics import (
    instrument_audit_event_processing,
    instrument_retention_policy,
    instrument_compliance_check,
    metrics_middleware,
    setup_service_info
)

# Fixtures para testes
@pytest.fixture
def audit_event_data():
    """Fixture para dados de evento de auditoria."""
    return {
        "event_id": "test-event-123",
        "event_type": "login",
        "action": "user_login",
        "actor_id": "user-456",
        "resource_id": "system-789",
        "resource_type": "application",
        "timestamp": 1690000000.0,
        "tenant_id": "tenant-abc",
        "regional_context": "br-south",
        "severity": "INFO",
        "details": {"ip": "192.168.1.1", "browser": "Chrome"}
    }

@pytest.fixture
def retention_policy_data():
    """Fixture para dados de política de retenção."""
    return {
        "policy_id": "policy-123",
        "policy_type": "time_based",
        "retention_period_days": 90,
        "tenant_id": "tenant-abc",
        "regional_context": "br-south",
        "description": "Política de retenção padrão para Brasil"
    }@pytest.fixture
def compliance_check_data():
    """Fixture para dados de verificação de conformidade."""
    return {
        "tenant_id": "tenant-abc",
        "regional_context": "br-south",
        "framework": "LGPD",
        "regulation": "Art-14",
        "resource_ids": ["resource-123", "resource-456"]
    }

@pytest.fixture
def mock_request():
    """Fixture para mock de requisição FastAPI."""
    mock_req = MagicMock()
    mock_req.method = "POST"
    mock_req.url.path = "/api/v1/audit/events"
    mock_req.headers = {
        "X-Tenant-ID": "tenant-abc",
        "X-Regional-Context": "br-south"
    }
    return mock_req

@pytest.fixture
def mock_response():
    """Fixture para mock de resposta FastAPI."""
    mock_res = MagicMock()
    mock_res.status_code = 201
    mock_res.body_size = 256
    return mock_res


# Testes para o decorador de processamento de eventos
@pytest.mark.asyncio
async def test_instrument_audit_event_processing_async(audit_event_data):
    """Testa o decorador de instrumentação para processamento de eventos assíncrono."""
    # Criar função de teste decorada
    @instrument_audit_event_processing
    async def test_process_event(event):
        await asyncio.sleep(0.01)  # Simular processamento
        return {"status": "processed", "event_id": event["event_id"]}
    
    # Executar função instrumentada
    result = await test_process_event(audit_event_data)
    
    # Verificar resultado
    assert result["status"] == "processed"
    assert result["event_id"] == audit_event_data["event_id"]
    
    # Verificar se as métricas foram registradas (via REGISTRY)
    # Nota: Em testes reais, usaríamos um coletor personalizado para verificação mais precisa
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "audit_events_total" for m in metric_samples)
    assert any(m.name == "audit_event_processing_duration" for m in metric_samples)# Teste para o decorador de política de retenção
@pytest.mark.asyncio
async def test_instrument_retention_policy(retention_policy_data):
    """Testa o decorador de instrumentação para políticas de retenção."""
    # Criar função de teste decorada
    @instrument_retention_policy
    async def test_apply_retention_policy(policy_type, tenant_id, regional_context, **kwargs):
        await asyncio.sleep(0.01)  # Simular processamento
        return {
            "status": "success", 
            "processed_count": 25, 
            "policy_type": policy_type
        }
    
    # Executar função instrumentada
    result = await test_apply_retention_policy(
        policy_type=retention_policy_data["policy_type"],
        tenant_id=retention_policy_data["tenant_id"],
        regional_context=retention_policy_data["regional_context"],
        retention_period_days=retention_policy_data["retention_period_days"]
    )
    
    # Verificar resultado
    assert result["status"] == "success"
    assert result["processed_count"] == 25
    assert result["policy_type"] == retention_policy_data["policy_type"]
    
    # Verificar métricas
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "audit_retention_policy_success" for m in metric_samples)
    assert any(m.name == "audit_retention_events_processed_total" for m in metric_samples)
    assert any(m.name == "audit_retention_policy_execution_duration" for m in metric_samples)


# Teste para o decorador de verificação de conformidade
@pytest.mark.asyncio
async def test_instrument_compliance_check(compliance_check_data):
    """Testa o decorador de instrumentação para verificações de conformidade."""
    # Criar função de teste decorada
    @instrument_compliance_check
    async def test_verify_compliance(tenant_id, regional_context, framework, regulation, resource_ids):
        await asyncio.sleep(0.01)  # Simular processamento
        return {"compliant": True, "framework": framework, "regulation": regulation}
    
    # Executar função instrumentada
    result = await test_verify_compliance(**compliance_check_data)
    
    # Verificar resultado
    assert result["compliant"] is True
    assert result["framework"] == compliance_check_data["framework"]
    assert result["regulation"] == compliance_check_data["regulation"]
    
    # Verificar métricas
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "audit_compliance_events_total" for m in metric_samples)
    assert any(m.name == "audit_compliance_check_duration" for m in metric_samples)
    assert any(m.name == "audit_regional_compliance_status" for m in metric_samples)# Teste para o middleware de métricas HTTP
@pytest.mark.asyncio
async def test_metrics_middleware(mock_request, mock_response):
    """Testa o middleware de métricas HTTP."""
    # Mock para a função call_next que retorna uma resposta
    async def mock_call_next(request):
        return mock_response
    
    # Executar middleware
    response = await metrics_middleware(mock_request, mock_call_next)
    
    # Verificar resultado
    assert response == mock_response
    
    # Verificar métricas
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "http_requests_total" for m in metric_samples)
    assert any(m.name == "http_request_duration_seconds" for m in metric_samples)


# Teste para o middleware de métricas HTTP com exceção
@pytest.mark.asyncio
async def test_metrics_middleware_exception(mock_request):
    """Testa o middleware de métricas HTTP quando ocorre uma exceção."""
    # Mock para a função call_next que lança uma exceção
    async def mock_call_next_exception(request):
        raise ValueError("Erro de teste")
    
    # Executar middleware (deve propagar a exceção)
    with pytest.raises(ValueError, match="Erro de teste"):
        await metrics_middleware(mock_request, mock_call_next_exception)
    
    # Verificar se as métricas ainda foram registradas
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "http_requests_total" for m in metric_samples)
    assert any(m.name == "http_request_duration_seconds" for m in metric_samples)


# Teste para configuração de informações do serviço
def test_setup_service_info():
    """Testa a configuração de informações estáticas do serviço."""
    # Configurar informações do serviço
    setup_service_info(
        version="1.0.0",
        build_id="test-build-123",
        commit_hash="test-commit-456",
        environment="test",
        region="global"
    )
    
    # Verificar se a métrica de informação foi registrada
    metric_samples = list(REGISTRY.collect())
    assert any(m.name == "audit_service_info" for m in metric_samples)
    
    # Em um teste real, poderíamos verificar os valores das labels da métrica Info