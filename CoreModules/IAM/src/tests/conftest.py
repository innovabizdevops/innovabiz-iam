"""
INNOVABIZ IAM - Fixtures para Testes
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Fixtures do pytest para testes de auditoria multi-contexto
"""

import pytest
import asyncio
import structlog
from typing import Dict, Any, Generator, AsyncGenerator
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession
from sqlalchemy.orm import sessionmaker

from tests.conftest.config import test_engine, TestingSessionLocal, TEST_TENANTS, TEST_REGIONAL_CONTEXTS
from app.db.base import Base
from app.models.audit import AuditEventCategory, AuditEventSeverity, ComplianceFramework, AuditEventCreate
from app.services.audit_service import AuditService

# Configura o logger para testes
logger = structlog.get_logger(__name__)


@pytest.fixture(scope="session")
def event_loop():
    """
    Cria um event loop para testes assíncronos.
    Necessário para testes assíncronos com pytest.
    """
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture(scope="function")
async def db_session() -> AsyncGenerator[AsyncSession, None]:
    """
    Fixture que fornece uma sessão de banco de dados para cada teste.
    Garante que todas as operações são revertidas após o teste.
    """
    async with test_engine.begin() as conn:
        # Recria todas as tabelas para cada teste (garantia de isolamento)
        await conn.run_sync(Base.metadata.create_all)
        
    async with TestingSessionLocal() as session:
        try:
            yield session
        finally:
            await session.rollback()
            await session.close()
    
    # Limpa todas as tabelas após o teste
    async with test_engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)


@pytest.fixture(scope="function")
async def audit_service(db_session: AsyncSession) -> AuditService:
    """
    Fixture que fornece uma instância isolada do serviço de auditoria para testes.
    """
    service = AuditService(db_session=db_session)
    return service


@pytest.fixture
def sample_audit_events() -> Dict[str, AuditEventCreate]:
    """
    Fornece eventos de auditoria de exemplo para todos os contextos regionais.
    """
    events = {}
    
    # Evento padrão Brasil - LGPD
    events["br_login_success"] = AuditEventCreate(
        category=AuditEventCategory.AUTHENTICATION,
        action="LOGIN",
        description="Login bem sucedido",
        resource_type="USER",
        resource_id="user123",
        resource_name="João Silva",
        severity=AuditEventSeverity.INFO,
        success=True,
        details={"method": "password", "mfa_used": True},
        tags=["authentication", "login", "success"],
        tenant_id="tenant_br_1",
        regional_context="BR",
        country_code="BR",
        language="pt-BR",
        user_id="user123",
        user_name="João Silva",
        correlation_id="corr-br-123456",
        source_ip="200.158.10.20",
    )
    
    # Evento Estados Unidos - SOX
    events["us_data_access"] = AuditEventCreate(
        category=AuditEventCategory.DATA_ACCESS,
        action="VIEW_FINANCIAL_REPORT",
        description="Financial report accessed",
        resource_type="FINANCIAL_REPORT",
        resource_id="report456",
        resource_name="Q2 2025 Financial Report",
        severity=AuditEventSeverity.INFO,
        success=True,
        details={"report_type": "quarterly", "department": "finance"},
        tags=["financial", "report", "sox"],
        tenant_id="tenant_us_1",
        regional_context="US",
        country_code="US",
        language="en-US",
        user_id="user456",
        user_name="John Smith",
        correlation_id="corr-us-654321",
        source_ip="104.28.42.30",
    )
    
    # Evento Europa - GDPR
    events["eu_consent_update"] = AuditEventCreate(
        category=AuditEventCategory.CONSENT,
        action="UPDATE_CONSENT",
        description="User updated privacy preferences",
        resource_type="USER_CONSENT",
        resource_id="consent789",
        resource_name="Privacy Preferences",
        severity=AuditEventSeverity.INFO,
        success=True,
        details={"marketing_emails": False, "data_processing": True},
        tags=["gdpr", "consent", "privacy"],
        tenant_id="tenant_eu_1",
        regional_context="EU",
        country_code="DE",
        language="de-DE",
        user_id="user789",
        user_name="Hans Müller",
        correlation_id="corr-eu-789012",
        source_ip="81.169.145.88",
    )
    
    # Evento Angola - BNA
    events["ao_transaction"] = AuditEventCreate(
        category=AuditEventCategory.FINANCIAL,
        action="MONEY_TRANSFER",
        description="Transferência bancária",
        resource_type="TRANSACTION",
        resource_id="trans101112",
        resource_name="Transferência para conta empresarial",
        severity=AuditEventSeverity.INFO,
        success=True,
        details={"amount": 250000, "currency": "AOA", "destination": "business"},
        tags=["bna", "transfer", "business"],
        tenant_id="tenant_ao_1",
        regional_context="AO",
        country_code="AO",
        language="pt-AO",
        user_id="user101112",
        user_name="António Santos",
        correlation_id="corr-ao-101112",
        source_ip="197.149.89.30",
    )
    
    # Evento de erro para testes
    events["error_event"] = AuditEventCreate(
        category=AuditEventCategory.SECURITY,
        action="PERMISSION_DENIED",
        description="Acesso negado ao recurso",
        resource_type="ADMIN_PANEL",
        resource_id="admin",
        resource_name="Painel de Administração",
        severity=AuditEventSeverity.ERROR,
        success=False,
        error_message="Usuário não possui permissão para acessar este recurso",
        details={"required_role": "ADMIN", "user_role": "USER"},
        tags=["security", "permission", "denied"],
        tenant_id="tenant_global",
        regional_context=None,
        country_code=None,
        language="en-US",
        user_id="user999",
        user_name="Test User",
        correlation_id="corr-error-999",
        source_ip="192.168.1.1",
    )
    
    return events