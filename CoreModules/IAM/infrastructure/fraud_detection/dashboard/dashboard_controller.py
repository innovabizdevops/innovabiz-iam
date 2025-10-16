#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Controlador da API do Dashboard de Monitoramento de Anomalias Comportamentais

Este módulo implementa o controlador da API REST para o dashboard de monitoramento,
expondo endpoints para obtenção de dados de visualização e configurações.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import logging
import datetime
import json
from typing import Dict, Any, List, Optional, Union

from fastapi import FastAPI, Request, Response, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import OAuth2PasswordBearer

from .dashboard_manager import DashboardManager
from ..authorization.approval_matrix import ApprovalMatrix

# Configuração do logger
logger = logging.getLogger("iam.trustguard.dashboard.controller")

# Esquema de autenticação
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Gerenciadores de dashboard por região
dashboard_managers = {}


# Middleware de autenticação
async def verify_token(token: str = Depends(oauth2_scheme)) -> Dict[str, Any]:
    """
    Verifica o token de acesso e retorna informações do usuário.
    
    Args:
        token: Token de acesso
        
    Returns:
        Informações do usuário autenticado
        
    Raises:
        HTTPException: Se o token for inválido
    """
    try:
        # Em produção, validar o token com o IAM
        # Este é um exemplo simplificado
        if not token or len(token) < 8:
            logger.warning(f"Token inválido: {token}")
            raise HTTPException(
                status_code=401,
                detail="Token de acesso inválido",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        # Informações do usuário (simuladas)
        return {
            "user_id": "user123",
            "username": "usuario.exemplo",
            "roles": ["analyst", "risk_manager"],
            "permissions": ["behavior:read", "alerts:read", "dashboard:access"],
            "region_access": ["global", "BR", "MZ", "AO"]
        }
    except Exception as e:
        logger.error(f"Erro ao verificar token: {str(e)}")
        raise HTTPException(
            status_code=401,
            detail="Falha na autenticação",
            headers={"WWW-Authenticate": "Bearer"},
        )


# Middleware de autorização
def verify_permission(required_permission: str, user_info: Dict[str, Any]) -> bool:
    """
    Verifica se o usuário tem a permissão necessária.
    
    Args:
        required_permission: Permissão necessária
        user_info: Informações do usuário
        
    Returns:
        True se o usuário tem permissão, False caso contrário
    """
    # Verificar permissões do usuário
    permissions = user_info.get("permissions", [])
    
    # Verificar permissão específica
    if required_permission in permissions:
        return True
    
    # Verificar permissão wildcard (ex: dashboard:*)
    permission_prefix = required_permission.split(":")[0] + ":*"
    if permission_prefix in permissions:
        return True
    
    # Verificar permissão total
    if "*:*" in permissions:
        return True
    
    return False


# Verificar acesso à região
def verify_region_access(region: str, user_info: Dict[str, Any]) -> bool:
    """
    Verifica se o usuário tem acesso à região especificada.
    
    Args:
        region: Código da região
        user_info: Informações do usuário
        
    Returns:
        True se o usuário tem acesso à região, False caso contrário
    """
    # Verificar acesso à região
    region_access = user_info.get("region_access", [])
    
    # Acesso global
    if "global" in region_access:
        return True
    
    # Acesso à região específica
    if region in region_access:
        return True
    
    return False


# Obter gerenciador de dashboard para uma região
def get_dashboard_manager(region: Optional[str] = None) -> DashboardManager:
    """
    Obtém o gerenciador de dashboard para uma região específica.
    
    Args:
        region: Código da região
        
    Returns:
        Gerenciador de dashboard
    """
    if not region:
        region = "global"
        
    # Criar gerenciador se não existir
    if region not in dashboard_managers:
        dashboard_managers[region] = DashboardManager(region=region)
        
    return dashboard_managers[region]


# Criar aplicação FastAPI
app = FastAPI(
    title="API de Dashboard de Monitoramento de Anomalias",
    description="API para o dashboard de monitoramento de anomalias comportamentais",
    version="1.0.0"
)

# Configurar CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Em produção, restringir para origens específicas
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/dashboard/config")
async def get_dashboard_config(
    region: str = Query(None, description="Código da região"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém a configuração do dashboard.
    
    Args:
        region: Código da região
        user_info: Informações do usuário autenticado
        
    Returns:
        Configuração do dashboard
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Obter gerenciador de dashboard
        dashboard_manager = get_dashboard_manager(region)
        
        # Obter configuração
        config = dashboard_manager.get_dashboard_configuration()
        
        return JSONResponse(content=config)
        
    except Exception as e:
        logger.error(f"Erro ao obter configuração do dashboard: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter configuração: {str(e)}"},
            status_code=500
        )


@app.get("/dashboard/panel/{panel_id}")
async def get_panel_data(
    panel_id: str,
    region: str = Query(None, description="Código da região"),
    start_date: str = Query(None, description="Data inicial (ISO)"),
    end_date: str = Query(None, description="Data final (ISO)"),
    timespan: str = Query(None, description="Período (1h, 6h, 12h, 24h, 7d, 30d)"),
    user_id: str = Query(None, description="Filtrar por ID de usuário"),
    severity: str = Query(None, description="Filtrar por severidade"),
    category: str = Query(None, description="Filtrar por categoria"),
    status: str = Query(None, description="Filtrar por status"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém dados de um painel específico.
    
    Args:
        panel_id: ID do painel
        region: Código da região
        start_date: Data inicial
        end_date: Data final
        timespan: Período
        user_id: Filtrar por ID de usuário
        severity: Filtrar por severidade
        category: Filtrar por categoria
        status: Filtrar por status
        user_info: Informações do usuário autenticado
        
    Returns:
        Dados do painel
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Obter gerenciador de dashboard
        dashboard_manager = get_dashboard_manager(region)
        
        # Construir filtros
        filters = {}
        
        if start_date:
            filters["start_date"] = start_date
            
        if end_date:
            filters["end_date"] = end_date
            
        if timespan:
            filters["timespan"] = timespan
            
        if user_id:
            filters["user_id"] = user_id
            
        if severity:
            filters["severity"] = severity
            
        if category:
            filters["category"] = category
            
        if status:
            filters["status"] = status
        
        # Obter dados do painel
        data = dashboard_manager.get_panel_data(panel_id, filters)
        
        return JSONResponse(content=data)
        
    except Exception as e:
        logger.error(f"Erro ao obter dados do painel {panel_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter dados do painel: {str(e)}"},
            status_code=500
        )


@app.get("/dashboard/alerts")
async def get_active_alerts(
    region: str = Query(None, description="Código da região"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém alertas ativos do dashboard.
    
    Args:
        region: Código da região
        user_info: Informações do usuário autenticado
        
    Returns:
        Lista de alertas ativos
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Simular alguns alertas ativos
        alerts = [
            {
                "id": "alert1",
                "name": "Taxa de Anomalias Alta",
                "description": "A taxa de anomalias está acima do limite de 15%",
                "current_value": 0.18,
                "threshold": 0.15,
                "triggered_at": (datetime.datetime.now() - datetime.timedelta(minutes=12)).isoformat(),
                "severity": "high",
            },
            {
                "id": "alert2",
                "name": "Transações de Alto Risco",
                "description": "Detectadas transações com score de risco muito alto",
                "current_value": 0.92,
                "threshold": 0.85,
                "triggered_at": (datetime.datetime.now() - datetime.timedelta(minutes=5)).isoformat(),
                "severity": "critical",
            },
        ]
        
        return JSONResponse(content=alerts)
        
    except Exception as e:
        logger.error(f"Erro ao obter alertas do dashboard: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter alertas: {str(e)}"},
            status_code=500
        )


@app.get("/dashboard/overview")
async def get_dashboard_overview(
    region: str = Query(None, description="Código da região"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém visão geral do dashboard (métricas principais).
    
    Args:
        region: Código da região
        user_info: Informações do usuário autenticado
        
    Returns:
        Visão geral do dashboard
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Obter gerenciador de dashboard
        dashboard_manager = get_dashboard_manager(region)
        
        # Obter dados do painel de visão geral
        data = dashboard_manager.get_panel_data("overview")
        
        # Adicionar informações sobre alertas ativos
        data["active_alerts_count"] = 2
        
        return JSONResponse(content=data)
        
    except Exception as e:
        logger.error(f"Erro ao obter visão geral do dashboard: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter visão geral: {str(e)}"},
            status_code=500
        )


@app.get("/dashboard/regions")
async def get_available_regions(
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém regiões disponíveis para o usuário.
    
    Args:
        user_info: Informações do usuário autenticado
        
    Returns:
        Lista de regiões disponíveis
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Obter regiões disponíveis para o usuário
        region_access = user_info.get("region_access", [])
        
        # Se tiver acesso global, retornar todas as regiões
        if "global" in region_access:
            regions = [
                {"code": "BR", "name": "Brasil"},
                {"code": "MZ", "name": "Moçambique"},
                {"code": "AO", "name": "Angola"},
                {"code": "PT", "name": "Portugal"},
                {"code": "CV", "name": "Cabo Verde"},
                {"code": "global", "name": "Global"},
            ]
        else:
            # Filtrar regiões acessíveis
            region_names = {
                "BR": "Brasil",
                "MZ": "Moçambique",
                "AO": "Angola",
                "PT": "Portugal",
                "CV": "Cabo Verde",
                "global": "Global",
            }
            
            regions = [
                {"code": code, "name": region_names.get(code, code)}
                for code in region_access
            ]
        
        return JSONResponse(content=regions)
        
    except Exception as e:
        logger.error(f"Erro ao obter regiões disponíveis: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter regiões: {str(e)}"},
            status_code=500
        )


def get_dashboard_router():
    """
    Obtém o router FastAPI para a API do dashboard.
    
    Returns:
        Router FastAPI
    """
    return appfrom .rules_integration import RulesDashboardIntegrator

# Integradores de regras por região
rules_integrators = {}


# Obter integrador de regras para uma região
def get_rules_integrator(region: Optional[str] = None) -> RulesDashboardIntegrator:
    """
    Obtém o integrador de regras para uma região específica.
    
    Args:
        region: Código da região
        
    Returns:
        Integrador de regras
    """
    if not region:
        region = "global"
        
    # Criar integrador se não existir
    if region not in rules_integrators:
        rules_integrators[region] = RulesDashboardIntegrator()
        
    return rules_integrators[region]


@app.post("/dashboard/evaluate")
async def evaluate_event(
    event_data: Dict[str, Any],
    region: str = Query(None, description="Código da região"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Avalia um evento comportamental utilizando o sistema de regras dinâmicas.
    
    Args:
        event_data: Dados do evento comportamental
        region: Código da região
        user_info: Informações do usuário autenticado
        
    Returns:
        Resultado da avaliação
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Contexto de avaliação
        context = {
            "user_id": user_info.get("user_id"),
            "username": user_info.get("username"),
            "source": "dashboard_api",
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        # Obter integrador de regras
        rules_integrator = get_rules_integrator(region)
        
        # Avaliar evento
        results, actions = await rules_integrator.evaluate_event(
            event_data, region, context
        )
        
        # Construir resposta
        response = {
            "event_id": event_data.get("id", "unknown"),
            "event_type": event_data.get("event_type", "unknown"),
            "results": results,
            "actions": actions,
            "matched_count": len(results),
            "timestamp": context.get("timestamp"),
            "region": region or "global"
        }
        
        return JSONResponse(content=response)
        
    except Exception as e:
        logger.error(f"Erro ao avaliar evento: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao avaliar evento: {str(e)}"},
            status_code=500
        )


@app.post("/dashboard/evaluate/batch")
async def evaluate_batch(
    events: List[Dict[str, Any]],
    region: str = Query(None, description="Código da região"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Avalia um lote de eventos comportamentais utilizando o sistema de regras dinâmicas.
    
    Args:
        events: Lista de eventos comportamentais
        region: Código da região
        user_info: Informações do usuário autenticado
        
    Returns:
        Resultados da avaliação
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Limitar tamanho do lote para evitar sobrecarga
        if len(events) > 100:
            return JSONResponse(
                content={"error": "Lote muito grande (máximo: 100 eventos)"},
                status_code=400
            )
        
        # Contexto de avaliação
        context = {
            "user_id": user_info.get("user_id"),
            "username": user_info.get("username"),
            "source": "dashboard_api_batch",
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        # Obter integrador de regras
        rules_integrator = get_rules_integrator(region)
        
        # Avaliar lote de eventos
        batch_results = await rules_integrator.evaluate_batch(
            events, region, context
        )
        
        # Processar resultados
        results = []
        for event, event_results, event_actions in batch_results:
            results.append({
                "event_id": event.get("id", "unknown"),
                "event_type": event.get("event_type", "unknown"),
                "results": event_results,
                "actions": event_actions,
                "matched_count": len(event_results)
            })
        
        # Construir resposta
        response = {
            "batch_size": len(events),
            "timestamp": context.get("timestamp"),
            "region": region or "global",
            "results": results,
            "total_matches": sum(r["matched_count"] for r in results)
        }
        
        return JSONResponse(content=response)
        
    except Exception as e:
        logger.error(f"Erro ao avaliar lote de eventos: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao avaliar lote de eventos: {str(e)}"},
            status_code=500
        )


@app.get("/dashboard/rule_statistics")
async def get_rule_statistics(
    region: str = Query(None, description="Código da região"),
    days: int = Query(7, description="Quantidade de dias para analisar"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém estatísticas de regras para exibição no dashboard.
    
    Args:
        region: Código da região
        days: Quantidade de dias para analisar
        user_info: Informações do usuário autenticado
        
    Returns:
        Estatísticas de regras
    """
    try:
        # Verificar permissão
        if not verify_permission("dashboard:access", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"acessar dashboard"
            )
            return JSONResponse(
                content={"error": "Sem permissão para acessar dashboard"},
                status_code=403
            )
        
        # Verificar acesso à região
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Limitar período
        if days > 90:
            days = 90
        
        # Obter integrador de regras
        rules_integrator = get_rules_integrator(region)
        
        # Obter estatísticas
        stats = await rules_integrator.get_rule_statistics(region, days)
        
        return JSONResponse(content=stats)
        
    except Exception as e:
        logger.error(f"Erro ao obter estatísticas de regras: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter estatísticas de regras: {str(e)}"},
            status_code=500
        )