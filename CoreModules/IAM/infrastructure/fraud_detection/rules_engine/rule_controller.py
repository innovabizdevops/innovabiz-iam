#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Controlador REST para o sistema de regras dinâmicas

Este módulo implementa o controlador REST para gerenciar o sistema de regras
dinâmicas, permitindo criar, atualizar, excluir e testar regras e conjuntos de regras.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import logging
import datetime
import json
import uuid
from typing import Dict, Any, List, Optional, Union, Tuple

from fastapi import FastAPI, Request, Response, Depends, HTTPException, Query, Body, Path
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import OAuth2PasswordBearer

from .rule_model import (
    Rule, RuleSet, RuleCondition, RuleGroup,
    RuleOperator, RuleLogicalOperator, RuleAction,
    RuleSeverity, RuleCategory, RuleValueType
)
from .rule_evaluator import RuleEvaluator, RuleEvaluationResult
from .rule_repository import RuleRepository

# Configuração do logger
logger = logging.getLogger("iam.trustguard.rules_engine.controller")

# Esquema de autenticação
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Repositório de regras (simulado)
rule_repository = RuleRepository()

# Avaliador de regras
rule_evaluator = RuleEvaluator()


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
            "roles": ["analyst", "admin"],
            "permissions": ["rules:read", "rules:write", "rules:execute"],
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
    
    # Verificar permissão wildcard (ex: rules:*)
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


# Criar aplicação FastAPI
app = FastAPI(
    title="API de Regras Dinâmicas",
    description="API para gerenciar e testar regras dinâmicas para detecção de anomalias comportamentais",
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


@app.get("/rules")
async def get_rules(
    region: Optional[str] = Query(None, description="Código da região"),
    tags: Optional[str] = Query(None, description="Tags separadas por vírgula"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém a lista de regras.
    
    Args:
        region: Código da região (opcional)
        tags: Tags separadas por vírgula (opcional)
        user_info: Informações do usuário autenticado
        
    Returns:
        Lista de regras
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"ler regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para ler regras"},
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
        
        # Processar tags
        tags_list = None
        if tags:
            tags_list = [tag.strip() for tag in tags.split(",") if tag.strip()]
        
        # Obter regras
        rules = rule_repository.get_rules(region=region, tags=tags_list)
        
        # Converter para dicionários
        rules_data = [rule.to_dict() for rule in rules]
        
        return JSONResponse(content=rules_data)
        
    except Exception as e:
        logger.error(f"Erro ao obter regras: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter regras: {str(e)}"},
            status_code=500
        )


@app.get("/rules/{rule_id}")
async def get_rule(
    rule_id: str = Path(..., description="ID da regra"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém uma regra pelo ID.
    
    Args:
        rule_id: ID da regra
        user_info: Informações do usuário autenticado
        
    Returns:
        Regra
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"ler regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para ler regras"},
                status_code=403
            )
        
        # Obter regra
        rule = rule_repository.get_rule(rule_id)
        
        # Verificar se a regra existe
        if not rule:
            logger.warning(f"Regra com ID {rule_id} não encontrada")
            return JSONResponse(
                content={"error": f"Regra com ID {rule_id} não encontrada"},
                status_code=404
            )
        
        # Verificar acesso à região
        if rule.region and not verify_region_access(rule.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {rule.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {rule.region}"},
                status_code=403
            )
        
        # Converter para dicionário
        rule_data = rule.to_dict()
        
        return JSONResponse(content=rule_data)
        
    except Exception as e:
        logger.error(f"Erro ao obter regra {rule_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter regra: {str(e)}"},
            status_code=500
        )


@app.post("/rules")
async def create_rule(
    rule_data: Dict[str, Any] = Body(..., description="Dados da regra"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Cria uma nova regra.
    
    Args:
        rule_data: Dados da regra
        user_info: Informações do usuário autenticado
        
    Returns:
        Regra criada
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"criar regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para criar regras"},
                status_code=403
            )
        
        # Verificar acesso à região
        region = rule_data.get("region")
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Criar regra
        rule = Rule.from_dict(rule_data)
        
        # Adicionar regra
        rule_id = rule_repository.add_rule(rule)
        
        # Obter regra atualizada
        rule = rule_repository.get_rule(rule_id)
        
        # Converter para dicionário
        rule_data = rule.to_dict()
        
        return JSONResponse(content=rule_data, status_code=201)
        
    except Exception as e:
        logger.error(f"Erro ao criar regra: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao criar regra: {str(e)}"},
            status_code=500
        )


@app.put("/rules/{rule_id}")
async def update_rule(
    rule_id: str = Path(..., description="ID da regra"),
    rule_data: Dict[str, Any] = Body(..., description="Dados da regra"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Atualiza uma regra existente.
    
    Args:
        rule_id: ID da regra
        rule_data: Dados da regra
        user_info: Informações do usuário autenticado
        
    Returns:
        Regra atualizada
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"atualizar regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para atualizar regras"},
                status_code=403
            )
        
        # Obter regra existente
        existing_rule = rule_repository.get_rule(rule_id)
        
        # Verificar se a regra existe
        if not existing_rule:
            logger.warning(f"Regra com ID {rule_id} não encontrada")
            return JSONResponse(
                content={"error": f"Regra com ID {rule_id} não encontrada"},
                status_code=404
            )
        
        # Verificar acesso à região existente
        if existing_rule.region and not verify_region_access(existing_rule.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {existing_rule.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {existing_rule.region}"},
                status_code=403
            )
        
        # Verificar acesso à nova região
        region = rule_data.get("region")
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Garantir que o ID é o mesmo
        rule_data["id"] = rule_id
        
        # Criar regra atualizada
        rule = Rule.from_dict(rule_data)
        
        # Atualizar regra
        success = rule_repository.update_rule(rule)
        
        if not success:
            return JSONResponse(
                content={"error": "Falha ao atualizar regra"},
                status_code=500
            )
        
        # Obter regra atualizada
        rule = rule_repository.get_rule(rule_id)
        
        # Converter para dicionário
        rule_data = rule.to_dict()
        
        return JSONResponse(content=rule_data)
        
    except Exception as e:
        logger.error(f"Erro ao atualizar regra {rule_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao atualizar regra: {str(e)}"},
            status_code=500
        )@app.delete("/rules/{rule_id}")
async def delete_rule(
    rule_id: str = Path(..., description="ID da regra"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Exclui uma regra existente.
    
    Args:
        rule_id: ID da regra
        user_info: Informações do usuário autenticado
        
    Returns:
        Confirmação de exclusão
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"excluir regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para excluir regras"},
                status_code=403
            )
        
        # Obter regra existente
        existing_rule = rule_repository.get_rule(rule_id)
        
        # Verificar se a regra existe
        if not existing_rule:
            logger.warning(f"Regra com ID {rule_id} não encontrada")
            return JSONResponse(
                content={"error": f"Regra com ID {rule_id} não encontrada"},
                status_code=404
            )
        
        # Verificar acesso à região
        if existing_rule.region and not verify_region_access(existing_rule.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {existing_rule.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {existing_rule.region}"},
                status_code=403
            )
        
        # Excluir regra
        success = rule_repository.delete_rule(rule_id)
        
        if not success:
            return JSONResponse(
                content={"error": "Falha ao excluir regra"},
                status_code=500
            )
        
        return JSONResponse(content={"message": f"Regra {rule_id} excluída com sucesso"})
        
    except Exception as e:
        logger.error(f"Erro ao excluir regra {rule_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao excluir regra: {str(e)}"},
            status_code=500
        )


@app.get("/rulesets")
async def get_rulesets(
    region: Optional[str] = Query(None, description="Código da região"),
    tags: Optional[str] = Query(None, description="Tags separadas por vírgula"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém a lista de conjuntos de regras.
    
    Args:
        region: Código da região (opcional)
        tags: Tags separadas por vírgula (opcional)
        user_info: Informações do usuário autenticado
        
    Returns:
        Lista de conjuntos de regras
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"ler conjuntos de regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para ler conjuntos de regras"},
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
        
        # Processar tags
        tags_list = None
        if tags:
            tags_list = [tag.strip() for tag in tags.split(",") if tag.strip()]
        
        # Obter conjuntos de regras
        rulesets = rule_repository.get_rulesets(region=region, tags=tags_list)
        
        # Converter para dicionários
        rulesets_data = [ruleset.to_dict() for ruleset in rulesets]
        
        return JSONResponse(content=rulesets_data)
        
    except Exception as e:
        logger.error(f"Erro ao obter conjuntos de regras: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter conjuntos de regras: {str(e)}"},
            status_code=500
        )


@app.get("/rulesets/{ruleset_id}")
async def get_ruleset(
    ruleset_id: str = Path(..., description="ID do conjunto de regras"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Obtém um conjunto de regras pelo ID.
    
    Args:
        ruleset_id: ID do conjunto de regras
        user_info: Informações do usuário autenticado
        
    Returns:
        Conjunto de regras
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"ler conjuntos de regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para ler conjuntos de regras"},
                status_code=403
            )
        
        # Obter conjunto de regras
        ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Verificar se o conjunto existe
        if not ruleset:
            logger.warning(f"Conjunto de regras com ID {ruleset_id} não encontrado")
            return JSONResponse(
                content={"error": f"Conjunto de regras com ID {ruleset_id} não encontrado"},
                status_code=404
            )
        
        # Verificar acesso à região
        if ruleset.region and not verify_region_access(ruleset.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {ruleset.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {ruleset.region}"},
                status_code=403
            )
        
        # Converter para dicionário
        ruleset_data = ruleset.to_dict()
        
        return JSONResponse(content=ruleset_data)
        
    except Exception as e:
        logger.error(f"Erro ao obter conjunto de regras {ruleset_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter conjunto de regras: {str(e)}"},
            status_code=500
        )


@app.post("/rulesets")
async def create_ruleset(
    ruleset_data: Dict[str, Any] = Body(..., description="Dados do conjunto de regras"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Cria um novo conjunto de regras.
    
    Args:
        ruleset_data: Dados do conjunto de regras
        user_info: Informações do usuário autenticado
        
    Returns:
        Conjunto de regras criado
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"criar conjuntos de regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para criar conjuntos de regras"},
                status_code=403
            )
        
        # Verificar acesso à região
        region = ruleset_data.get("region")
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Criar conjunto de regras
        ruleset = RuleSet.from_dict(ruleset_data)
        
        # Adicionar conjunto de regras
        ruleset_id = rule_repository.add_ruleset(ruleset)
        
        # Obter conjunto atualizado
        ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Converter para dicionário
        ruleset_data = ruleset.to_dict()
        
        return JSONResponse(content=ruleset_data, status_code=201)
        
    except Exception as e:
        logger.error(f"Erro ao criar conjunto de regras: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao criar conjunto de regras: {str(e)}"},
            status_code=500
        )


@app.put("/rulesets/{ruleset_id}")
async def update_ruleset(
    ruleset_id: str = Path(..., description="ID do conjunto de regras"),
    ruleset_data: Dict[str, Any] = Body(..., description="Dados do conjunto de regras"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Atualiza um conjunto de regras existente.
    
    Args:
        ruleset_id: ID do conjunto de regras
        ruleset_data: Dados do conjunto de regras
        user_info: Informações do usuário autenticado
        
    Returns:
        Conjunto de regras atualizado
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"atualizar conjuntos de regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para atualizar conjuntos de regras"},
                status_code=403
            )
        
        # Obter conjunto existente
        existing_ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Verificar se o conjunto existe
        if not existing_ruleset:
            logger.warning(f"Conjunto de regras com ID {ruleset_id} não encontrado")
            return JSONResponse(
                content={"error": f"Conjunto de regras com ID {ruleset_id} não encontrado"},
                status_code=404
            )
        
        # Verificar acesso à região existente
        if existing_ruleset.region and not verify_region_access(existing_ruleset.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {existing_ruleset.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {existing_ruleset.region}"},
                status_code=403
            )
        
        # Verificar acesso à nova região
        region = ruleset_data.get("region")
        if region and not verify_region_access(region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {region}"},
                status_code=403
            )
        
        # Garantir que o ID é o mesmo
        ruleset_data["id"] = ruleset_id
        
        # Criar conjunto atualizado
        ruleset = RuleSet.from_dict(ruleset_data)
        
        # Atualizar conjunto
        success = rule_repository.update_ruleset(ruleset)
        
        if not success:
            return JSONResponse(
                content={"error": "Falha ao atualizar conjunto de regras"},
                status_code=500
            )
        
        # Obter conjunto atualizado
        ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Converter para dicionário
        ruleset_data = ruleset.to_dict()
        
        return JSONResponse(content=ruleset_data)
        
    except Exception as e:
        logger.error(f"Erro ao atualizar conjunto de regras {ruleset_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao atualizar conjunto de regras: {str(e)}"},
            status_code=500
        )


@app.delete("/rulesets/{ruleset_id}")
async def delete_ruleset(
    ruleset_id: str = Path(..., description="ID do conjunto de regras"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Exclui um conjunto de regras existente.
    
    Args:
        ruleset_id: ID do conjunto de regras
        user_info: Informações do usuário autenticado
        
    Returns:
        Confirmação de exclusão
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:write", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"excluir conjuntos de regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para excluir conjuntos de regras"},
                status_code=403
            )
        
        # Obter conjunto existente
        existing_ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Verificar se o conjunto existe
        if not existing_ruleset:
            logger.warning(f"Conjunto de regras com ID {ruleset_id} não encontrado")
            return JSONResponse(
                content={"error": f"Conjunto de regras com ID {ruleset_id} não encontrado"},
                status_code=404
            )
        
        # Verificar acesso à região
        if existing_ruleset.region and not verify_region_access(existing_ruleset.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {existing_ruleset.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {existing_ruleset.region}"},
                status_code=403
            )
        
        # Excluir conjunto
        success = rule_repository.delete_ruleset(ruleset_id)
        
        if not success:
            return JSONResponse(
                content={"error": "Falha ao excluir conjunto de regras"},
                status_code=500
            )
        
        return JSONResponse(
            content={"message": f"Conjunto de regras {ruleset_id} excluído com sucesso"}
        )
        
    except Exception as e:
        logger.error(f"Erro ao excluir conjunto de regras {ruleset_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao excluir conjunto de regras: {str(e)}"},
            status_code=500
        )


@app.post("/rules/{rule_id}/test")
async def test_rule(
    rule_id: str = Path(..., description="ID da regra"),
    event_data: Dict[str, Any] = Body(..., description="Dados do evento para teste"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Testa uma regra com dados de evento.
    
    Args:
        rule_id: ID da regra
        event_data: Dados do evento para teste
        user_info: Informações do usuário autenticado
        
    Returns:
        Resultado da avaliação
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:execute", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"executar regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para executar regras"},
                status_code=403
            )
        
        # Obter regra
        rule = rule_repository.get_rule(rule_id)
        
        # Verificar se a regra existe
        if not rule:
            logger.warning(f"Regra com ID {rule_id} não encontrada")
            return JSONResponse(
                content={"error": f"Regra com ID {rule_id} não encontrada"},
                status_code=404
            )
        
        # Verificar acesso à região
        if rule.region and not verify_region_access(rule.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {rule.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {rule.region}"},
                status_code=403
            )
        
        # Contexto de execução
        context = {
            "user_id": user_info.get("user_id"),
            "username": user_info.get("username"),
            "test_mode": True,
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        # Avaliar regra
        result = rule_evaluator.evaluate_rule(rule, event_data, context)
        
        # Executar ações (simulado)
        action_results = {}
        if result.matched:
            action_results = rule_evaluator.execute_actions(result, event_data, context)
        
        # Construir resposta
        response = {
            "rule_id": rule.id,
            "rule_name": rule.name,
            "matched": result.matched,
            "score": result.score,
            "severity": rule.severity.value if hasattr(rule.severity, "value") else rule.severity,
            "category": rule.category.value if hasattr(rule.category, "value") else rule.category,
            "actions": result.actions,
            "matched_fields": result.matched_fields,
            "action_results": action_results,
            "context": context
        }
        
        return JSONResponse(content=response)
        
    except Exception as e:
        logger.error(f"Erro ao testar regra {rule_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao testar regra: {str(e)}"},
            status_code=500
        )


@app.post("/rulesets/{ruleset_id}/test")
async def test_ruleset(
    ruleset_id: str = Path(..., description="ID do conjunto de regras"),
    event_data: Dict[str, Any] = Body(..., description="Dados do evento para teste"),
    user_info: Dict[str, Any] = Depends(verify_token)
) -> JSONResponse:
    """
    Testa um conjunto de regras com dados de evento.
    
    Args:
        ruleset_id: ID do conjunto de regras
        event_data: Dados do evento para teste
        user_info: Informações do usuário autenticado
        
    Returns:
        Resultados da avaliação
    """
    try:
        # Verificar permissão
        if not verify_permission("rules:execute", user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem permissão para "
                f"executar regras"
            )
            return JSONResponse(
                content={"error": "Sem permissão para executar regras"},
                status_code=403
            )
        
        # Obter conjunto de regras
        ruleset = rule_repository.get_ruleset(ruleset_id)
        
        # Verificar se o conjunto existe
        if not ruleset:
            logger.warning(f"Conjunto de regras com ID {ruleset_id} não encontrado")
            return JSONResponse(
                content={"error": f"Conjunto de regras com ID {ruleset_id} não encontrado"},
                status_code=404
            )
        
        # Verificar acesso à região
        if ruleset.region and not verify_region_access(ruleset.region, user_info):
            logger.warning(
                f"Usuário {user_info.get('username')} sem acesso à "
                f"região {ruleset.region}"
            )
            return JSONResponse(
                content={"error": f"Sem acesso à região {ruleset.region}"},
                status_code=403
            )
        
        # Contexto de execução
        context = {
            "user_id": user_info.get("user_id"),
            "username": user_info.get("username"),
            "test_mode": True,
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        # Avaliar conjunto
        results, action_results = rule_evaluator.process_event(ruleset, event_data, context)
        
        # Construir resposta
        response = {
            "ruleset_id": ruleset.id,
            "ruleset_name": ruleset.name,
            "results": [
                {
                    "rule_id": result.rule.id,
                    "rule_name": result.rule.name,
                    "matched": result.matched,
                    "score": result.score,
                    "severity": result.rule.severity.value if hasattr(result.rule.severity, "value") else result.rule.severity,
                    "category": result.rule.category.value if hasattr(result.rule.category, "value") else result.rule.category,
                    "actions": result.actions,
                    "matched_fields": result.matched_fields
                }
                for result in results
            ],
            "action_results": action_results,
            "context": context,
            "matched_count": sum(1 for result in results if result.matched),
            "total_score": sum(result.score for result in results if result.matched)
        }
        
        return JSONResponse(content=response)
        
    except Exception as e:
        logger.error(f"Erro ao testar conjunto de regras {ruleset_id}: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao testar conjunto de regras: {str(e)}"},
            status_code=500
        )


def get_rules_engine_router():
    """
    Obtém o router FastAPI para a API de regras.
    
    Returns:
        Router FastAPI
    """
    return app