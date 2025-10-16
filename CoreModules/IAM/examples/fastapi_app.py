"""
Exemplo de aplicação FastAPI com integração completa de observabilidade.

Este exemplo demonstra como integrar o framework de observabilidade com uma
aplicação FastAPI, incluindo métricas, health checks e rastreamento.
"""

import asyncio
import random
import uuid
from typing import Dict, List, Optional, Any

import uvicorn
from fastapi import FastAPI, Depends, Header, HTTPException, Request
from pydantic import BaseModel

# Importação do módulo de observabilidade
import sys
import os

# Adiciona diretório pai ao PYTHONPATH para importar o módulo de observabilidade
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from src.observability import (
    configure_observability,
    ObservabilityConfig,
    HealthStatus,
    instrument_audit_event,
    instrument_function,
    traced
)


# Modelos para eventos de auditoria
class AuditEvent(BaseModel):
    """Modelo base para eventos de auditoria."""
    event_id: str = None
    event_type: str
    tenant_id: str
    user_id: str
    resource_id: Optional[str] = None
    action: str
    timestamp: str = None
    details: Dict[str, Any] = {}
    
    def __init__(self, **data):
        if 'event_id' not in data or not data['event_id']:
            data['event_id'] = str(uuid.uuid4())
        super().__init__(**data)


class LoginEvent(AuditEvent):
    """Evento de login de usuário."""
    def __init__(self, **data):
        data['event_type'] = 'login'
        data['action'] = 'login'
        super().__init__(**data)


# Criação da aplicação FastAPI
app = FastAPI(
    title="INNOVABIZ IAM Audit Service",
    description="Serviço de auditoria para IAM com observabilidade integrada",
    version="1.0.0"
)

# Configuração de observabilidade
observability = configure_observability(
    app,
    service_name="iam-audit-service",
    service_version="1.0.0",
    environment="development",
    default_tenant="innovabiz",
    default_region="eu-west-1"
)


# Serviço de Eventos de Auditoria
class AuditEventService:
    """Serviço para processamento de eventos de auditoria."""
    
    @instrument_audit_event(event_type="login")
    async def process_login_event(
        self,
        event: LoginEvent,
        tenant: str = "default",
        region: str = "global"
    ) -> Dict[str, Any]:
        """
        Processa um evento de login.
        
        Args:
            event: Evento de login
            tenant: ID do tenant
            region: Código da região
        """
        # Simulação de processamento
        await asyncio.sleep(random.uniform(0.01, 0.05))
        
        # Simulação de falha aleatória (5% de chance)
        if random.random() < 0.05:
            raise Exception("Falha ao processar evento de login")
        
        # Resultado do processamento
        return {
            "event_id": event.event_id,
            "status": "processed",
            "tenant": tenant,
            "region": region
        }
    
    @traced(name="audit.retention.apply")
    async def apply_retention_policy(
        self,
        tenant: str,
        region: str,
        policy_name: str,
        days: int
    ) -> Dict[str, Any]:
        """
        Aplica uma política de retenção para eventos de auditoria.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            policy_name: Nome da política de retenção
            days: Número de dias para retenção
        """
        # Simulação de processamento
        await asyncio.sleep(random.uniform(0.1, 0.3))
        
        # Número simulado de registros afetados
        affected_records = random.randint(10, 100)
        
        # Simulação de falha aleatória (3% de chance)
        if random.random() < 0.03:
            raise Exception(f"Falha ao aplicar política de retenção {policy_name}")
        
        # Resultado da operação
        return {
            "policy": policy_name,
            "tenant": tenant,
            "region": region,
            "days": days,
            "affected_records": affected_records
        }


# Injeção de dependência para o serviço
def get_audit_service():
    """Fornece uma instância do serviço de auditoria."""
    return AuditEventService()


# Rotas da aplicação
@app.get("/")
async def root():
    """Rota principal da aplicação."""
    return {
        "service": "INNOVABIZ IAM Audit Service",
        "version": "1.0.0",
        "status": "online"
    }


@app.post("/audit/events", status_code=201)
@traced(name="api.audit.create_event")
async def create_audit_event(
    event: AuditEvent,
    request: Request,
    audit_service: AuditEventService = Depends(get_audit_service),
    tenant_id: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region")
):
    """
    Cria um novo evento de auditoria.
    
    Args:
        event: Evento de auditoria
        request: Objeto de requisição
        audit_service: Serviço de eventos de auditoria
        tenant_id: ID do tenant (via header)
        region: Código da região (via header)
    """
    # Usa tenant e região dos headers, se disponíveis
    tenant = tenant_id or "default"
    region = region or "global"
    
    # Sobrescreve tenant no evento, se necessário
    if event.tenant_id != tenant:
        event.tenant_id = tenant
    
    # Processa diferentes tipos de eventos
    if event.event_type == "login":
        login_event = LoginEvent(
            user_id=event.user_id,
            tenant_id=event.tenant_id,
            details=event.details,
            resource_id=event.resource_id
        )
        result = await audit_service.process_login_event(login_event, tenant, region)
        return result
    else:
        # Evento genérico
        return {
            "event_id": event.event_id,
            "status": "received",
            "tenant": tenant,
            "region": region
        }


@app.post("/audit/retention/apply")
@traced(name="api.audit.apply_retention")
async def apply_retention(
    request: Request,
    policy_name: str,
    days: int,
    audit_service: AuditEventService = Depends(get_audit_service),
    tenant_id: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region")
):
    """
    Aplica uma política de retenção para eventos de auditoria.
    
    Args:
        request: Objeto de requisição
        policy_name: Nome da política de retenção
        days: Número de dias para retenção
        audit_service: Serviço de eventos de auditoria
        tenant_id: ID do tenant (via header)
        region: Código da região (via header)
    """
    # Usa tenant e região dos headers, se disponíveis
    tenant = tenant_id or "default"
    region = region or "global"
    
    # Valida parâmetros
    if days <= 0:
        raise HTTPException(status_code=400, detail="Dias de retenção deve ser maior que zero")
    
    # Aplica política de retenção
    result = await audit_service.apply_retention_policy(tenant, region, policy_name, days)
    return result


@app.get("/audit/simulate-error")
async def simulate_error():
    """Simula um erro para testar instrumentação de exceções."""
    # Gera um erro aleatório
    error_types = [
        ValueError("Valor inválido"),
        KeyError("Chave não encontrada"),
        RuntimeError("Erro de execução"),
        NotImplementedError("Funcionalidade não implementada")
    ]
    raise random.choice(error_types)


# Rota para simular carga variada para métricas
@app.get("/audit/simulate-load")
async def simulate_load(
    request: Request,
    events: int = 10,
    tenant_id: str = Header(None, alias="X-Tenant-ID"),
    region: str = Header(None, alias="X-Region")
):
    """
    Simula carga variada para gerar métricas.
    
    Args:
        request: Objeto de requisição
        events: Número de eventos a simular
        tenant_id: ID do tenant (via header)
        region: Código da região (via header)
    """
    # Usa tenant e região dos headers, se disponíveis
    tenant = tenant_id or "default"
    region = region or "global"
    
    # Limita número de eventos
    if events > 100:
        events = 100
    
    # Simula processamento de eventos
    audit_service = get_audit_service()
    results = []
    
    for i in range(events):
        event_type = random.choice(["login", "logout", "access", "create", "update", "delete"])
        user_id = f"user-{random.randint(1, 100)}"
        
        event = AuditEvent(
            event_type=event_type,
            tenant_id=tenant,
            user_id=user_id,
            action=event_type,
            resource_id=f"resource-{random.randint(1, 50)}",
            details={"source": "simulation"}
        )
        
        if event_type == "login":
            try:
                login_event = LoginEvent(
                    user_id=user_id,
                    tenant_id=tenant,
                    details={"source": "simulation"}
                )
                result = await audit_service.process_login_event(login_event, tenant, region)
                results.append(result)
            except Exception as e:
                results.append({"error": str(e), "event_type": "login"})
        else:
            # Simula processamento genérico
            await asyncio.sleep(random.uniform(0.01, 0.05))
            results.append({
                "event_id": event.event_id,
                "status": "processed",
                "event_type": event_type
            })
    
    # Aplica política de retenção aleatória
    if random.random() < 0.3:
        policy_name = random.choice(["gdpr", "pci-dss", "sox", "hipaa"])
        days = random.randint(30, 365)
        try:
            retention_result = await audit_service.apply_retention_policy(
                tenant, region, policy_name, days
            )
            results.append({"retention": retention_result})
        except Exception as e:
            results.append({"error": str(e), "operation": "retention"})
    
    return {
        "processed_events": len(results),
        "tenant": tenant,
        "region": region,
        "results": results[:5]  # Limita resultados para evitar resposta grande
    }


if __name__ == "__main__":
    # Inicia o servidor
    uvicorn.run(app, host="0.0.0.0", port=8000)