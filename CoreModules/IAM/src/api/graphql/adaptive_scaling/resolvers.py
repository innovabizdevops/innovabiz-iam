"""
Resolvers GraphQL para o serviço de escalonamento adaptativo.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
from typing import Dict, List, Any, Optional, Union
from datetime import datetime

from opentelemetry import trace
from graphql import GraphQLResolveInfo

from .helpers import (
    check_permissions,
    transform_event_to_graphql,
    transform_trigger_to_graphql,
    transform_policy_to_graphql,
    transform_security_profile,
    DatabaseQueries
)
from ...services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from ...services.adaptive_scaling.models import (
    SecurityLevel,
    SecurityMechanism,
    ScalingDirection,
    SecurityAdjustment
)
from ....app.repositories.trust_score_repository import TrustScoreRepository

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingResolvers:
    """
    Implementação dos resolvers GraphQL para o serviço de escalonamento adaptativo.
    """
    
    def __init__(
        self,
        adaptive_scaling_service: AdaptiveScalingService,
        trust_repository: TrustScoreRepository
    ):
        """
        Inicializa os resolvers.
        
        Args:
            adaptive_scaling_service: Serviço de escalonamento adaptativo
            trust_repository: Repositório de TrustScore
        """
        self.adaptive_scaling_service = adaptive_scaling_service
        self.trust_repository = trust_repository
        self.notification_service = None
        self.audit_service = None
        self.db_pool = adaptive_scaling_service.db_pool
    
    #
    # Resolvers de Query
    #
    
    async def get_user_security_profile(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        user_id: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Obtém o perfil de segurança completo de um usuário.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            user_id: ID do usuário
            context_id: ID do contexto (opcional)
            
        Returns:
            Perfil de segurança do usuário
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_user_security_profile") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                
                # Obter perfil de segurança
                profile = await self.adaptive_scaling_service.get_security_profile(
                    user_id=user_id,
                    tenant_id=tenant_id,
                    context_id=context_id
                )
                
                # Transformar para o formato GraphQL
                return await transform_security_profile(
                    profile,
                    lambda user_id, tenant_id, context_id: self._get_recent_events(user_id, tenant_id, context_id)
                )
            except Exception as e:
                logger.error(f"Erro ao obter perfil de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def _get_recent_events(
        self,
        user_id: str,
        tenant_id: str,
        context_id: Optional[str] = None,
        limit: int = 5
    ) -> List[Dict[str, Any]]:
        """
        Obtém os eventos recentes para um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            limit: Limite de eventos
            
        Returns:
            Lista de eventos recentes
        """
        return await DatabaseQueries.fetch_recent_scaling_events(
            self.db_pool,
            user_id,
            tenant_id,
            context_id,
            limit
        )
    
    async def get_current_security_level(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        user_id: str,
        mechanism: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Obtém o nível de segurança atual para um mecanismo específico.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            user_id: ID do usuário
            mechanism: Mecanismo de segurança
            context_id: ID do contexto (opcional)
            
        Returns:
            Nível de segurança atual
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_current_security_level") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                span.set_attribute("security.mechanism", mechanism)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                
                # Converter string para enum
                mechanism_enum = SecurityMechanism(mechanism)
                
                # Obter nível de segurança atual
                level = await self.adaptive_scaling_service.get_security_level(
                    user_id=user_id,
                    tenant_id=tenant_id,
                    mechanism=mechanism_enum,
                    context_id=context_id
                )
                
                # Obter detalhes adicionais
                parameters, metadata, updated_at, expires_at = await DatabaseQueries.get_security_level_details(
                    self.db_pool,
                    user_id,
                    tenant_id,
                    mechanism,
                    context_id
                )
                
                # Construir resposta
                return {
                    "mechanism": mechanism,
                    "level": level.value,
                    "parameters": parameters,
                    "updatedAt": updated_at,
                    "expiresAt": expires_at,
                    "metadata": metadata
                }
            except Exception as e:
                logger.error(f"Erro ao obter nível de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_events(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        filter: Optional[Dict[str, Any]] = None,
        pagination: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Obtém eventos de escalonamento paginados.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            filter: Filtros a serem aplicados (opcional)
            pagination: Configurações de paginação (opcional)
            
        Returns:
            Eventos de escalonamento paginados
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_events") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("tenant.id", tenant_id)
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                
                # Montar parâmetros de filtro
                filter_params = {"tenant_id": tenant_id}
                
                if filter:
                    # Adicionar filtros adicionais
                    filter_params.update(filter)
                
                if pagination:
                    # Adicionar parâmetros de paginação
                    filter_params["page"] = pagination.get("page", 1)
                    filter_params["page_size"] = pagination.get("pageSize", 20)
                
                # Buscar eventos paginados
                return await DatabaseQueries.fetch_scaling_events_page(self.db_pool, filter_params)
            except Exception as e:
                logger.error(f"Erro ao obter eventos de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_event_by_id(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        event_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        Obtém um evento de escalonamento pelo seu ID.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            event_id: ID do evento
            
        Returns:
            Evento de escalonamento ou None se não encontrado
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_event_by_id") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("event.id", event_id)
                
                # Buscar evento
                event = await DatabaseQueries.fetch_scaling_event(self.db_pool, event_id)
                
                # Se não encontrou, retorna None
                if not event:
                    return None
                
                # Verificar permissão para o tenant do evento
                await check_permissions(info.context, "adaptive_scaling:read", event["tenantId"])
                
                return event
            except Exception as e:
                logger.error(f"Erro ao obter evento de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_triggers(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        context_id: Optional[str] = None,
        region_code: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        """
        Obtém todos os gatilhos de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            region_code: Código da região (opcional)
            
        Returns:
            Lista de gatilhos de escalonamento
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_triggers") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                if region_code:
                    span.set_attribute("region.code", region_code)
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                
                # Construir consulta SQL
                query = """
                SELECT id, name, enabled, tenant_specific, tenant_id, region_specific, region_code,
                       context_specific, context_id, condition_type, dimension, comparison,
                       threshold_value, scaling_direction, description, metadata
                FROM scaling_triggers
                WHERE tenant_id = $1 OR tenant_id IS NULL
                """
                
                # Parâmetros da consulta
                params = [tenant_id]
                param_index = 2
                
                # Adicionar filtros adicionais
                if context_id:
                    query += f" AND (context_id = ${param_index} OR context_id IS NULL)"
                    params.append(context_id)
                    param_index += 1
                
                if region_code:
                    query += f" AND (region_code = ${param_index} OR region_code IS NULL)"
                    params.append(region_code)
                    param_index += 1
                
                # Executar consulta
                rows = await self.db_pool.fetch(query, *params)
                
                # Transformar resultados
                triggers = []
                for row in rows:
                    triggers.append(transform_trigger_to_graphql(dict(row)))
                
                return triggers
            except Exception as e:
                logger.error(f"Erro ao obter gatilhos de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_trigger_by_id(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        trigger_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        Obtém um gatilho de escalonamento pelo seu ID.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            trigger_id: ID do gatilho
            
        Returns:
            Gatilho de escalonamento ou None se não encontrado
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_trigger_by_id") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("trigger.id", trigger_id)
                
                # Buscar gatilho
                row = await self.db_pool.fetchrow(
                    """
                    SELECT id, name, enabled, tenant_specific, tenant_id, region_specific, region_code,
                           context_specific, context_id, condition_type, dimension, comparison,
                           threshold_value, scaling_direction, description, metadata
                    FROM scaling_triggers
                    WHERE id = $1
                    """,
                    trigger_id
                )
                
                # Se não encontrou, retorna None
                if not row:
                    return None
                
                # Converter para dicionário
                trigger = dict(row)
                
                # Verificar permissão para o tenant do gatilho
                tenant_id = trigger["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                else:
                    # Se o gatilho for global, verificar permissão admin
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Transformar para o formato GraphQL
                return transform_trigger_to_graphql(trigger)
            except Exception as e:
                logger.error(f"Erro ao obter gatilho de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_policies(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        context_id: Optional[str] = None,
        region_code: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        """
        Obtém todas as políticas de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            region_code: Código da região (opcional)
            
        Returns:
            Lista de políticas de escalonamento
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_policies") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                if region_code:
                    span.set_attribute("region.code", region_code)
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                
                # Construir consulta SQL
                query = """
                SELECT id, name, enabled, priority, tenant_id, region_code, context_id,
                       trigger_ids, adjustment_map, description, created_at, updated_at, metadata
                FROM scaling_policies
                WHERE tenant_id = $1 OR tenant_id IS NULL
                """
                
                # Parâmetros da consulta
                params = [tenant_id]
                param_index = 2
                
                # Adicionar filtros adicionais
                if context_id:
                    query += f" AND (context_id = ${param_index} OR context_id IS NULL)"
                    params.append(context_id)
                    param_index += 1
                
                if region_code:
                    query += f" AND (region_code = ${param_index} OR region_code IS NULL)"
                    params.append(region_code)
                    param_index += 1
                
                # Adicionar ordenação por prioridade
                query += " ORDER BY priority DESC"
                
                # Executar consulta
                rows = await self.db_pool.fetch(query, *params)
                
                # Transformar resultados
                policies = []
                for row in rows:
                    policies.append(transform_policy_to_graphql(dict(row)))
                
                return policies
            except Exception as e:
                logger.error(f"Erro ao obter políticas de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_scaling_policy_by_id(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        policy_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        Obtém uma política de escalonamento pelo seu ID.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            policy_id: ID da política
            
        Returns:
            Política de escalonamento ou None se não encontrada
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_scaling_policy_by_id") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("policy.id", policy_id)
                
                # Buscar política
                row = await self.db_pool.fetchrow(
                    """
                    SELECT id, name, enabled, priority, tenant_id, region_code, context_id,
                           trigger_ids, adjustment_map, description, created_at, updated_at, metadata
                    FROM scaling_policies
                    WHERE id = $1
                    """,
                    policy_id
                )
                
                # Se não encontrou, retorna None
                if not row:
                    return None
                
                # Converter para dicionário
                policy = dict(row)
                
                # Verificar permissão para o tenant da política
                tenant_id = policy["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:read", tenant_id)
                else:
                    # Se a política for global, verificar permissão admin
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Transformar para o formato GraphQL
                return transform_policy_to_graphql(policy)
            except Exception as e:
                logger.error(f"Erro ao obter política de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def get_adaptive_scaling_status(
        self, 
        _: Any, 
        info: GraphQLResolveInfo
    ) -> Dict[str, Any]:
        """
        Obtém o status atual do serviço de escalonamento adaptativo.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            
        Returns:
            Status atual do serviço
        """
        with tracer.start_as_current_span("adaptive_scaling_resolvers.get_adaptive_scaling_status") as span:
            try:
                # Verificar permissão de administrador
                await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Obter estatísticas do serviço
                cache_stats = self.adaptive_scaling_service.get_cache_stats()
                
                # Consultar contagens no banco de dados
                triggers_count = await self.db_pool.fetchval("SELECT COUNT(*) FROM scaling_triggers")
                policies_count = await self.db_pool.fetchval("SELECT COUNT(*) FROM scaling_policies")
                events_count = await self.db_pool.fetchval("SELECT COUNT(*) FROM scaling_events")
                
                # Obter estatísticas de processamento
                processing_stats = self.adaptive_scaling_service.get_processing_stats()
                
                # Construir resposta
                return {
                    "isActive": True,
                    "cacheStatus": {
                        "triggersInCache": cache_stats.get("triggers_count", 0),
                        "policiesInCache": cache_stats.get("policies_count", 0),
                        "lastCacheRefresh": cache_stats.get("last_refresh")
                    },
                    "databaseStatus": {
                        "triggersCount": triggers_count,
                        "policiesCount": policies_count,
                        "eventsCount": events_count
                    },
                    "processingStats": {
                        "evaluationsLastHour": processing_stats.get("evaluations_last_hour", 0),
                        "triggersActivatedLastDay": processing_stats.get("triggers_activated_last_day", 0),
                        "adjustmentsAppliedLastDay": processing_stats.get("adjustments_applied_last_day", 0),
                        "averageProcessingTimeMs": processing_stats.get("avg_processing_time_ms", 0)
                    },
                    "version": "1.0.0"
                }
            except Exception as e:
                logger.error(f"Erro ao obter status do serviço: {e}")
                span.record_exception(e)
                raise