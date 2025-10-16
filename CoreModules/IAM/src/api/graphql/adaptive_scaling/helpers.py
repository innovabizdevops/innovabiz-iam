"""
Funções auxiliares para os resolvers GraphQL do serviço de escalonamento adaptativo.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime

from ...common.errors import NotAuthorizedError, ResourceNotFoundError

# Configuração de logging
logger = logging.getLogger(__name__)


async def check_permissions(context: Any, permission: str, tenant_id: Optional[str] = None) -> None:
    """
    Verifica se o usuário tem a permissão necessária.
    
    Args:
        context: Contexto da requisição GraphQL
        permission: Permissão a ser verificada
        tenant_id: ID do tenant, se aplicável
    
    Raises:
        NotAuthorizedError: Se o usuário não tiver permissão
    """
    if not hasattr(context, "user") or not context.user:
        raise NotAuthorizedError("Usuário não autenticado")
        
    # Implementação simplificada, adapte conforme seu sistema de permissões
    user = context.user
    
    # Se o usuário for superadmin, permite tudo
    if user.get("role") == "superadmin":
        return
        
    # Verifica permissões específicas do tenant
    if tenant_id:
        tenant_permissions = user.get("tenant_permissions", {}).get(tenant_id, [])
        if permission in tenant_permissions or "adaptive_scaling:admin" in tenant_permissions:
            return
            
    # Verifica permissões globais
    global_permissions = user.get("permissions", [])
    if permission in global_permissions or "adaptive_scaling:admin" in global_permissions:
        return
        
    # Se chegou aqui, não tem permissão
    logger.warning(
        f"Acesso negado: usuário {user.get('username')} tentou acessar {permission} "
        f"para tenant {tenant_id if tenant_id else 'global'}"
    )
    raise NotAuthorizedError(f"Sem permissão para {permission}")


def transform_event_to_graphql(row: Dict[str, Any]) -> Dict[str, Any]:
    """
    Transforma um registro de evento do formato do banco para o formato GraphQL.
    """
    # Converter campos JSON
    dimension_scores = json.loads(row["dimension_scores"]) if row["dimension_scores"] else {}
    adjustments = json.loads(row["adjustments"]) if row["adjustments"] else []
    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
    
    # Transformar ajustes
    transformed_adjustments = []
    for adj in adjustments:
        transformed_adjustments.append({
            "mechanism": adj["mechanism"],
            "currentLevel": adj["current_level"],
            "newLevel": adj["new_level"],
            "reason": adj.get("reason"),
            "parameters": adj.get("parameters"),
            "expiresAt": adj.get("expires_at")
        })
    
    # Construir resposta
    return {
        "id": row["id"],
        "userId": row["user_id"],
        "tenantId": row["tenant_id"],
        "contextId": row["context_id"],
        "regionCode": row["region_code"],
        "triggerId": row["trigger_id"],
        "policyId": row["policy_id"],
        "trustScore": row["trust_score"],
        "dimensionScores": dimension_scores,
        "scalingDirection": row["scaling_direction"],
        "adjustments": transformed_adjustments,
        "eventTime": row["event_time"],
        "expiresAt": row["expires_at"],
        "metadata": metadata
    }


def transform_trigger_to_graphql(row: Dict[str, Any]) -> Dict[str, Any]:
    """
    Transforma um registro de gatilho do formato do banco para o formato GraphQL.
    """
    # Converter campos JSON
    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
    
    # Construir resposta
    return {
        "id": row["id"],
        "name": row["name"],
        "enabled": row["enabled"],
        "tenantSpecific": row["tenant_specific"],
        "tenantId": row["tenant_id"],
        "regionSpecific": row["region_specific"],
        "regionCode": row["region_code"],
        "contextSpecific": row["context_specific"],
        "contextId": row["context_id"],
        "conditionType": row["condition_type"],
        "dimension": row["dimension"],
        "comparison": row["comparison"],
        "thresholdValue": row["threshold_value"],
        "scalingDirection": row["scaling_direction"],
        "description": row["description"],
        "metadata": metadata
    }


def transform_policy_to_graphql(row: Dict[str, Any]) -> Dict[str, Any]:
    """
    Transforma um registro de política do formato do banco para o formato GraphQL.
    """
    # Converter campos JSON
    trigger_ids = json.loads(row["trigger_ids"]) if row["trigger_ids"] else []
    adjustment_map = json.loads(row["adjustment_map"]) if row["adjustment_map"] else {}
    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
    
    # Construir resposta
    return {
        "id": row["id"],
        "name": row["name"],
        "enabled": row["enabled"],
        "priority": row["priority"],
        "tenantId": row["tenant_id"],
        "regionCode": row["region_code"],
        "contextId": row["context_id"],
        "triggerIds": trigger_ids,
        "adjustmentMap": adjustment_map,
        "description": row["description"],
        "createdAt": row["created_at"],
        "updatedAt": row["updated_at"],
        "metadata": metadata
    }


async def transform_security_profile(
    profile: Dict[str, Any], 
    get_recent_events_func
) -> Dict[str, Any]:
    """
    Transforma o perfil de segurança do formato interno para o formato GraphQL.
    
    Args:
        profile: Perfil de segurança
        get_recent_events_func: Função para obter eventos recentes
    """
    # Obter os eventos de escalonamento recentes
    events = await get_recent_events_func(
        profile.get("user_id"), 
        profile.get("tenant_id"),
        profile.get("context_id")
    )
    
    # Transformar níveis de segurança
    security_levels = []
    for mechanism, data in profile.get("security_levels", {}).items():
        security_levels.append({
            "mechanism": mechanism,
            "level": data.get("level"),
            "parameters": data.get("parameters", {}),
            "updatedAt": data.get("updated_at") or datetime.now(),
            "expiresAt": data.get("expires_at"),
            "metadata": data.get("metadata", {})
        })
        
    # Construir resposta
    return {
        "userId": profile.get("user_id"),
        "tenantId": profile.get("tenant_id"),
        "contextId": profile.get("context_id"),
        "securityLevels": security_levels,
        "lastModified": profile.get("last_modified"),
        "scalingEvents": events
    }


class DatabaseQueries:
    """
    Classe com métodos para consultas ao banco de dados.
    """
    
    @staticmethod
    async def get_security_level_details(
        db_pool,
        user_id: str, 
        tenant_id: str, 
        mechanism: str,
        context_id: Optional[str]
    ) -> Tuple[Dict[str, Any], Dict[str, Any], datetime, Optional[datetime]]:
        """
        Obtém detalhes adicionais de um nível de segurança.
        
        Returns:
            Tuple com parâmetros, metadata, data de atualização e data de expiração
        """
        # Buscar no repositório os detalhes do nível de segurança
        try:
            details = await db_pool.fetchrow(
                """
                SELECT parameters, metadata, updated_at, expires_at 
                FROM security_levels 
                WHERE user_id = $1 AND tenant_id = $2 AND mechanism = $3
                AND (context_id = $4 OR (context_id IS NULL AND $4 IS NULL))
                """,
                user_id, tenant_id, mechanism, context_id
            )
            
            if details:
                return (
                    json.loads(details["parameters"]) if details["parameters"] else {},
                    json.loads(details["metadata"]) if details["metadata"] else {},
                    details["updated_at"],
                    details["expires_at"]
                )
            return ({}, {}, datetime.now(), None)
        except Exception as e:
            logger.error(f"Erro ao buscar detalhes do nível de segurança: {e}")
            return ({}, {}, datetime.now(), None)
    
    @staticmethod
    async def fetch_scaling_events_page(db_pool, filter_params: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca eventos de escalonamento paginados com base nos filtros.
        """
        # Construir consulta SQL base
        base_query = """
        SELECT id, user_id, tenant_id, context_id, region_code, 
               trigger_id, policy_id, trust_score, dimension_scores,
               scaling_direction, adjustments, event_time, expires_at, metadata
        FROM scaling_events
        WHERE tenant_id = $1
        """
        
        # Parâmetros da consulta
        params = [filter_params["tenant_id"]]
        param_index = 2
        
        # Adicionar filtros condicionais
        if filter_params.get("user_id"):
            base_query += f" AND user_id = ${param_index}"
            params.append(filter_params["user_id"])
            param_index += 1
            
        if filter_params.get("context_id"):
            base_query += f" AND (context_id = ${param_index} OR context_id IS NULL)"
            params.append(filter_params["context_id"])
            param_index += 1
            
        if filter_params.get("region_code"):
            base_query += f" AND (region_code = ${param_index} OR region_code IS NULL)"
            params.append(filter_params["region_code"])
            param_index += 1
            
        if filter_params.get("from_date"):
            base_query += f" AND event_time >= ${param_index}"
            params.append(filter_params["from_date"])
            param_index += 1
            
        if filter_params.get("to_date"):
            base_query += f" AND event_time <= ${param_index}"
            params.append(filter_params["to_date"])
            param_index += 1
            
        if filter_params.get("scaling_direction"):
            base_query += f" AND scaling_direction = ${param_index}"
            params.append(filter_params["scaling_direction"])
            param_index += 1
            
        # Contar total de registros
        count_query = f"SELECT COUNT(*) as total FROM ({base_query}) as count_query"
        total_count = await db_pool.fetchval(count_query, *params)
        
        # Adicionar ordenação e paginação
        page = filter_params.get("page", 1)
        page_size = filter_params.get("page_size", 20)
        offset = (page - 1) * page_size
        
        base_query += " ORDER BY event_time DESC LIMIT $%d OFFSET $%d" % (param_index, param_index + 1)
        params.extend([page_size, offset])
        
        # Executar consulta paginada
        rows = await db_pool.fetch(base_query, *params)
        
        # Transformar resultados
        items = []
        for row in rows:
            items.append(transform_event_to_graphql(dict(row)))
            
        # Calcular informações de paginação
        page_count = (total_count + page_size - 1) // page_size
        
        return {
            "items": items,
            "totalCount": total_count,
            "pageCount": page_count,
            "currentPage": page,
            "hasNextPage": page < page_count
        }
    
    @staticmethod
    async def fetch_scaling_event(db_pool, event_id: str) -> Optional[Dict[str, Any]]:
        """
        Busca um evento de escalonamento por ID.
        """
        row = await db_pool.fetchrow(
            """
            SELECT id, user_id, tenant_id, context_id, region_code, 
                   trigger_id, policy_id, trust_score, dimension_scores,
                   scaling_direction, adjustments, event_time, expires_at, metadata
            FROM scaling_events
            WHERE id = $1
            """,
            event_id
        )
        
        if not row:
            return None
            
        return transform_event_to_graphql(dict(row))
    
    @staticmethod
    async def fetch_recent_scaling_events(
        db_pool,
        user_id: str, 
        tenant_id: str, 
        context_id: Optional[str],
        limit: int = 5
    ) -> List[Dict[str, Any]]:
        """
        Busca os eventos de escalonamento mais recentes para um usuário.
        """
        rows = await db_pool.fetch(
            """
            SELECT id, user_id, tenant_id, context_id, region_code, 
                   trigger_id, policy_id, trust_score, dimension_scores,
                   scaling_direction, adjustments, event_time, expires_at, metadata
            FROM scaling_events
            WHERE user_id = $1 AND tenant_id = $2
            AND (context_id = $3 OR (context_id IS NULL AND $3 IS NULL))
            ORDER BY event_time DESC
            LIMIT $4
            """,
            user_id, tenant_id, context_id, limit
        )
        
        return [transform_event_to_graphql(dict(row)) for row in rows]