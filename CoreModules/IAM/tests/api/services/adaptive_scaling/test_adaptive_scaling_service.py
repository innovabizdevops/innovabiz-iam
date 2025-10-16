"""
Testes unitários para o serviço de escalonamento adaptativo.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import uuid
import pytest
import asyncio
from unittest.mock import MagicMock, patch, AsyncMock
from datetime import datetime, timedelta

from src.api.services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from src.api.services.adaptive_scaling.models import (
    ScalingDirection,
    SecurityLevel,
    SecurityMechanism,
    ScalingTrigger, 
    ScalingPolicy,
    ScalingEvent,
    SecurityAdjustment,
    AdaptiveConfig
)
from src.app.trust_guard_models import TrustScoreResult, DetectedAnomaly

# Mock para o pool de banco de dados
class MockPool:
    def __init__(self, conn=None):
        self.conn = conn or MockConnection()
        
    async def acquire(self):
        return self.conn
        
    async def release(self, conn):
        pass

# Mock para conexão de banco de dados        
class MockConnection:
    def __init__(self, fetch_results=None, fetchrow_results=None, execute_results=None):
        self.fetch_results = fetch_results or {}
        self.fetchrow_results = fetchrow_results or {}
        self.execute_calls = []
        self.execute_results = execute_results or {}
        
    async def fetch(self, query, *args, **kwargs):
        key = self._make_key(query, args)
        return self.fetch_results.get(key, [])
        
    async def fetchrow(self, query, *args, **kwargs):
        key = self._make_key(query, args)
        return self.fetchrow_results.get(key, None)
        
    async def execute(self, query, *args, **kwargs):
        self.execute_calls.append((query, args))
        key = self._make_key(query, args)
        return self.execute_results.get(key, None)
        
    def _make_key(self, query, args):
        # Simplificação para fins de teste
        return str(query)[:30] + str(len(args))
        
    def __await__(self):
        async def _await_self():
            return self
        return _await_self().__await__()
        
    async def __aenter__(self):
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        pass


@pytest.fixture
def mock_db_pool():
    """Fixture para criar um mock do pool de banco de dados"""
    conn = MockConnection()
    return MockPool(conn)

@pytest.fixture
def mock_trust_repository():
    """Fixture para criar um mock do repositório de trust score"""
    return AsyncMock()

@pytest.fixture
def mock_trust_query_service():
    """Fixture para criar um mock do serviço de consulta de trust score"""
    return AsyncMock()

@pytest.fixture
def mock_notification_service():
    """Fixture para criar um mock do serviço de notificação"""
    return AsyncMock()

@pytest.fixture
def sample_trust_score_result():
    """Fixture para criar um resultado de trust score para testes"""
    return TrustScoreResult(
        user_id="test_user_123",
        tenant_id="tenant_456",
        context_id="context_789",
        regional_context="AO",
        calculation_id=str(uuid.uuid4()),
        calculation_time=datetime.now(),
        overall_score=0.75,
        confidence_interval=0.05,
        dimension_scores={
            "identity": 0.8,
            "behavioral": 0.7,
            "financial": 0.75,
            "historical": 0.85,
            "contextual": 0.65
        },
        factors=[],
        anomalies=[
            DetectedAnomaly(
                anomaly_id=str(uuid.uuid4()),
                anomaly_type="unusual_location",
                severity="medium",
                affected_dimensions=["contextual", "behavioral"],
                description="Login de local não usual",
                detection_time=datetime.now(),
                confidence=0.85,
                metadata={"location": "Unknown Location"}
            )
        ],
        data_sources=["iam", "transactions"],
        metadata={}
    )


@pytest.mark.asyncio
async def test_initialize_success(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service, 
    mock_notification_service
):
    """Testa inicialização bem-sucedida do serviço"""
    # Configurar mock para retornar configuração
    config_row = {"config": json.dumps({"enabled": True, "default_cooldown_minutes": 15})}
    mock_db_pool.conn.fetchrow_results = {
        "SELECT config FROM adaptive_sc0": config_row
    }
    
    # Configurar mocks para triggers e políticas
    triggers = [
        {"id": "trigger1", "enabled": True, "tenant_id": None, 
         "tenant_specific": False, "region_code": None, "region_specific": False,
         "context_id": None, "context_specific": False, "condition_type": "threshold",
         "dimension": "overall", "comparison": "lt", "threshold_value": 0.6,
         "scaling_direction": "up", "name": "Low Score Trigger"}
    ]
    
    policies = [
        {"id": "policy1", "name": "Security Policy", "enabled": True, 
         "priority": 10, "tenant_id": None, "region_code": None, "context_id": None,
         "trigger_ids": json.dumps(["trigger1"]), 
         "adjustment_map": json.dumps({
             "up": {"auth_factors": "high", "session_timeout": "high"},
             "down": {"auth_factors": "standard", "session_timeout": "standard"}
         })}
    ]
    
    # Adicionar ao mock
    mock_db_pool.conn.fetch_results = {
        "SELECT * FROM scaling_trigg0": triggers,
        "SELECT * FROM scaling_polic0": policies
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service,
        notification_service=mock_notification_service
    )
    
    # Executar inicialização
    await service.initialize()
    
    # Verificações
    assert service.config.enabled == True
    assert service.config.default_cooldown_minutes == 15
    assert "trigger1" in service._trigger_cache
    assert "policy1" in service._policy_cache


@pytest.mark.asyncio
async def test_evaluate_trust_score_no_triggers(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service,
    sample_trust_score_result
):
    """Testa avaliação sem gatilhos acionados"""
    # Criar serviço com mocks vazios
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Configurar cache vazio
    service._trigger_cache = {}
    service._policy_cache = {}
    
    # Executar avaliação
    result = await service.evaluate_trust_score(sample_trust_score_result)
    
    # Verificar que nenhum evento foi gerado
    assert result is None


@pytest.mark.asyncio
async def test_evaluate_trust_score_with_trigger(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service,
    sample_trust_score_result
):
    """Testa avaliação com gatilho acionado"""
    # Configurar mock para níveis de segurança atuais
    mock_db_pool.conn.fetchrow_results = {
        "SELECT level FROM security_l0": {"level": "standard"},
        "SELECT level FROM security_d0": {"level": "standard"}
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Habilitar serviço
    service.config = AdaptiveConfig(enabled=True)
    
    # Configurar gatilho
    trigger = ScalingTrigger(
        id="trigger1",
        name="Low Trust Score",
        enabled=True,
        tenant_specific=False,
        region_specific=False,
        context_specific=False,
        condition_type="threshold",
        dimension="overall",
        comparison="lt",
        threshold_value=0.8,  # Score abaixo de 0.8 aciona
        scaling_direction=ScalingDirection.UP
    )
    
    # Configurar política
    policy = ScalingPolicy(
        id="policy1",
        name="Enhanced Security",
        enabled=True,
        priority=10,
        trigger_ids=["trigger1"],
        adjustment_map={
            "up": {
                "auth_factors": "high",
                "session_timeout": "high"
            },
            "down": {
                "auth_factors": "standard",
                "session_timeout": "standard"
            }
        }
    )
    
    # Configurar caches
    service._trigger_cache = {"trigger1": trigger}
    service._policy_cache = {"policy1": policy}
    
    # Modificar sample para score baixo
    sample_trust_score_result.overall_score = 0.75  # Abaixo do threshold de 0.8
    
    # Executar avaliação
    result = await service.evaluate_trust_score(sample_trust_score_result)
    
    # Verificar resultado
    assert result is not None
    assert result.scaling_direction == ScalingDirection.UP
    assert result.policy_id == "policy1"
    assert result.trigger_id == "trigger1"
    assert len(result.adjustments) == 2
    
    # Verificar que os ajustes foram salvos
    assert len(mock_db_pool.conn.execute_calls) >= 1


@pytest.mark.asyncio
async def test_get_current_security_level(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service
):
    """Testa obtenção do nível de segurança atual"""
    # Configurar mock para retornar um nível
    mock_db_pool.conn.fetchrow_results = {
        "SELECT level FROM security_l0": {"level": "high"}
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Obter nível de segurança
    level = await service.get_current_security_level(
        user_id="test_user",
        tenant_id="test_tenant",
        mechanism=SecurityMechanism.AUTH_FACTORS
    )
    
    # Verificar resultado
    assert level == SecurityLevel.HIGH


@pytest.mark.asyncio
async def test_get_current_security_level_not_found(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service
):
    """Testa obtenção do nível de segurança quando não encontrado"""
    # Configurar mock para retornar None (não encontrado)
    mock_db_pool.conn.fetchrow_results = {
        "SELECT level FROM security_l0": None,
        "SELECT level FROM security_d0": {"level": "standard"}
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Sobrescrever método interno para retornar nível padrão
    async def _get_default_security_level(*args, **kwargs):
        return SecurityLevel.STANDARD
        
    service._get_default_security_level = _get_default_security_level
    
    # Obter nível de segurança
    level = await service.get_current_security_level(
        user_id="test_user",
        tenant_id="test_tenant",
        mechanism=SecurityMechanism.AUTH_FACTORS
    )
    
    # Verificar resultado
    assert level == SecurityLevel.STANDARD


@pytest.mark.asyncio
async def test_notify_user_of_changes(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service, 
    mock_notification_service
):
    """Testa notificação de usuário sobre mudanças"""
    # Criar serviço com mock de notificação
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service,
        notification_service=mock_notification_service
    )
    
    # Criar evento de escalonamento
    event = ScalingEvent(
        id="event123",
        user_id="user456",
        tenant_id="tenant789",
        context_id="context123",
        region_code="AO",
        trigger_id="trigger1",
        policy_id="policy1",
        trust_score=0.75,
        dimension_scores={"overall": 0.75},
        scaling_direction=ScalingDirection.UP,
        adjustments=[
            SecurityAdjustment(
                mechanism=SecurityMechanism.AUTH_FACTORS,
                current_level=SecurityLevel.STANDARD,
                new_level=SecurityLevel.HIGH,
                parameters={},
                reason="TrustScore baixo"
            )
        ],
        event_time=datetime.now(),
        expires_at=datetime.now() + timedelta(hours=24)
    )
    
    # Executar notificação
    await service._notify_user_of_changes(event)
    
    # Verificar que o serviço de notificação foi chamado
    mock_notification_service.send_user_notification.assert_called_once()
    
    # Verificar argumentos
    call_args = mock_notification_service.send_user_notification.call_args
    assert call_args[1]["user_id"] == "user456"
    assert call_args[1]["tenant_id"] == "tenant789"
    assert "aumentados" in call_args[1]["title"]
    assert "Fatores de autenticação" in call_args[1]["body"]


@pytest.mark.asyncio
async def test_check_triggers(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service, 
    sample_trust_score_result
):
    """Testa verificação de gatilhos"""
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Configurar gatilhos no cache
    service._trigger_cache = {
        "trigger1": ScalingTrigger(
            id="trigger1",
            name="Low Trust Score",
            enabled=True,
            tenant_specific=False,
            region_specific=False,
            context_specific=False,
            condition_type="threshold",
            dimension="overall",
            comparison="lt",
            threshold_value=0.8,  # Score abaixo de 0.8 aciona
            scaling_direction=ScalingDirection.UP
        ),
        "trigger2": ScalingTrigger(
            id="trigger2",
            name="Low Identity Score",
            enabled=True,
            tenant_specific=False,
            region_specific=False,
            context_specific=False,
            condition_type="threshold",
            dimension="identity",
            comparison="lt",
            threshold_value=0.9,  # Score de identidade abaixo de 0.9 aciona
            scaling_direction=ScalingDirection.UP
        ),
        "trigger3": ScalingTrigger(
            id="trigger3",
            name="Tenant Specific",
            enabled=True,
            tenant_specific=True,
            tenant_id="another_tenant",  # Não deve acionar pois o tenant é diferente
            region_specific=False,
            context_specific=False,
            condition_type="threshold",
            dimension="overall",
            comparison="lt",
            threshold_value=0.9,
            scaling_direction=ScalingDirection.UP
        ),
        "trigger4": ScalingTrigger(
            id="trigger4",
            name="Anomaly Detection",
            enabled=True,
            tenant_specific=False,
            region_specific=False,
            context_specific=False,
            condition_type="anomaly",
            dimension="contextual",
            comparison="gt",
            threshold_value=0,  # Qualquer anomalia aciona
            scaling_direction=ScalingDirection.UP
        )
    }
    
    # Configurar políticas no cache
    service._policy_cache = {
        "policy1": ScalingPolicy(
            id="policy1",
            name="Enhanced Security",
            enabled=True,
            priority=10,
            trigger_ids=["trigger1", "trigger2", "trigger4"],
            adjustment_map={
                "up": {
                    "auth_factors": "high",
                    "session_timeout": "high"
                }
            }
        ),
        "policy2": ScalingPolicy(
            id="policy2",
            name="Tenant Specific",
            enabled=True,
            priority=5,
            tenant_id="another_tenant",  # Não deve acionar pois o tenant é diferente
            trigger_ids=["trigger3"],
            adjustment_map={
                "up": {
                    "auth_factors": "high"
                }
            }
        )
    }
    
    # Executar verificação de gatilhos
    triggered = await service._check_triggers(sample_trust_score_result)
    
    # Deve acionar trigger1, trigger2 e trigger4
    expected_triggers = {
        ("policy1", "trigger1", ScalingDirection.UP),
        ("policy1", "trigger2", ScalingDirection.UP),
        ("policy1", "trigger4", ScalingDirection.UP)
    }
    
    assert set(triggered) == expected_triggers
    assert len(triggered) == 3


@pytest.mark.asyncio
async def test_determine_security_adjustments(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service,
    sample_trust_score_result
):
    """Testa determinação de ajustes de segurança"""
    # Configurar mock para níveis de segurança atuais
    mock_db_pool.conn.fetchrow_results = {
        "SELECT level FROM security_l0": {"level": "standard"}
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Criar política de teste
    policy = ScalingPolicy(
        id="policy1",
        name="Enhanced Security",
        enabled=True,
        priority=10,
        trigger_ids=["trigger1"],
        adjustment_map={
            "up": {
                "auth_factors": "high",
                "session_timeout": "high"
            }
        }
    )
    
    # Executar determinação de ajustes
    adjustments = await service._determine_security_adjustments(
        sample_trust_score_result, 
        policy, 
        ScalingDirection.UP
    )
    
    # Verificar ajustes
    assert len(adjustments) == 2
    
    # Verificar mecanismos ajustados
    mechanisms = [adj.mechanism for adj in adjustments]
    assert SecurityMechanism.AUTH_FACTORS in mechanisms
    assert SecurityMechanism.SESSION_TIMEOUT in mechanisms
    
    # Verificar níveis ajustados
    for adj in adjustments:
        assert adj.new_level == SecurityLevel.HIGH
        assert adj.current_level == SecurityLevel.STANDARD


@pytest.mark.asyncio
async def test_get_user_security_profile(
    mock_db_pool, 
    mock_trust_repository, 
    mock_trust_query_service
):
    """Testa obtenção de perfil completo de segurança"""
    # Configurar mock para retornar níveis de segurança
    security_levels = [
        {
            "mechanism": "auth_factors", 
            "level": "high",
            "context_id": None,
            "updated_at": datetime.now(),
            "expires_at": datetime.now() + timedelta(hours=24),
            "metadata": json.dumps({"reason": "TrustScore"})
        },
        {
            "mechanism": "session_timeout", 
            "level": "high",
            "context_id": None,
            "updated_at": datetime.now(),
            "expires_at": datetime.now() + timedelta(hours=24),
            "metadata": json.dumps({"reason": "TrustScore"})
        }
    ]
    
    # Configurar mock para retornar eventos recentes
    scaling_events = [
        {
            "id": "event123",
            "trigger_id": "trigger1",
            "policy_id": "policy1",
            "scaling_direction": "up",
            "event_time": datetime.now(),
            "expires_at": datetime.now() + timedelta(hours=24)
        }
    ]
    
    # Configurar mocks
    mock_db_pool.conn.fetch_results = {
        "SELECT mechanism, level, co0": security_levels,
        "SELECT id, trigger_id, pol0": scaling_events
    }
    
    # Criar serviço
    service = AdaptiveScalingService(
        db_pool=mock_db_pool,
        trust_repository=mock_trust_repository,
        trust_query_service=mock_trust_query_service
    )
    
    # Obter perfil de segurança
    profile = await service.get_user_security_profile(
        user_id="test_user",
        tenant_id="test_tenant"
    )
    
    # Verificar resultado
    assert profile["user_id"] == "test_user"
    assert profile["tenant_id"] == "test_tenant"
    assert len(profile["security_levels"]) == 1  # Um contexto (default)
    assert len(profile["scaling_events"]) == 1
    assert "auth_factors" in profile["security_levels"]["default"]
    assert "session_timeout" in profile["security_levels"]["default"]
    
    # Verificar níveis
    assert profile["security_levels"]["default"]["auth_factors"]["level"] == "high"
    assert profile["security_levels"]["default"]["session_timeout"]["level"] == "high"