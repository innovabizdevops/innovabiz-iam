"""
INNOVABIZ IAM - Configuração de Testes
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Configurações para testes automatizados do módulo de auditoria multi-contexto
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA, SOX
"""

import os
import sys
import pytest
from typing import Dict, Any, Generator
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import NullPool

# Garante que o diretório raiz da aplicação está no PYTHONPATH
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../api')))

# Importações da aplicação
from app.db.base import Base
from app.db.session import get_db

# Configurações para ambiente de teste
TEST_SQLALCHEMY_DATABASE_URL = "postgresql+asyncpg://test:test@localhost:5432/innovabiz_iam_test"

# Engine do SQLAlchemy para testes (com NullPool para garantir limpeza após testes)
test_engine = create_async_engine(
    TEST_SQLALCHEMY_DATABASE_URL,
    poolclass=NullPool,
    echo=False
)

# Configuração de sessão de teste
TestingSessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=test_engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# Contextos multi-regionais para testes
TEST_REGIONAL_CONTEXTS = {
    "BR": {
        "country_code": "BR",
        "currency": "BRL",
        "language": "pt-BR",
        "compliance": ["LGPD", "BACEN", "PCI_DSS"],
    },
    "US": {
        "country_code": "US",
        "currency": "USD", 
        "language": "en-US",
        "compliance": ["CCPA", "SOX", "PCI_DSS"],
    },
    "EU": {
        "country_code": "EU",
        "currency": "EUR",
        "language": "en-GB",
        "compliance": ["GDPR", "PSD2", "PCI_DSS"],
    },
    "AO": {
        "country_code": "AO",
        "currency": "AOA",
        "language": "pt-AO",
        "compliance": ["BNA", "PCI_DSS"],
    }
}

# Tenants de teste para diferentes contextos
TEST_TENANTS = {
    "tenant_br_1": {"name": "Tenant Brasil 1", "regional_context": "BR"},
    "tenant_us_1": {"name": "Tenant USA 1", "regional_context": "US"},
    "tenant_eu_1": {"name": "Tenant Europe 1", "regional_context": "EU"},
    "tenant_ao_1": {"name": "Tenant Angola 1", "regional_context": "AO"},
    "tenant_global": {"name": "Global Tenant", "regional_context": None},
}

# Override das variáveis de ambiente para testes
os.environ["AUDIT_BATCH_SIZE"] = "10"
os.environ["AUDIT_BATCH_INTERVAL"] = "1"
os.environ["AUDIT_DEFAULT_RETENTION_DAYS"] = "30"