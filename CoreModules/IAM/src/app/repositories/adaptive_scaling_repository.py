"""
Repositório para operações de banco de dados do serviço de escalonamento adaptativo.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import uuid
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime

from opentelemetry import trace
from asyncpg.pool import Pool

from ..repositories.base_repository import BaseRepository
from ...services.adaptive_scaling.models import (
    ScalingDirection,
    SecurityLevel,
    SecurityMechanism,
    ScalingTrigger, 
    ScalingPolicy,
    ScalingEvent,
    SecurityAdjustment
)

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingRepository(BaseRepository):
    """
    Repositório para operações de banco de dados relacionadas ao escalonamento adaptativo.
    """
    
    def __init__(self, db_pool: Pool):
        """
        Inicializa o repositório com um pool de conexões.
        
        Args:
            db_pool: Pool de conexões asyncpg
        """
        super().__init__(db_pool)
    
    async def get_security_profile(
        self, 
        user_id: str, 
        tenant_id: str, 
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Obtém o perfil de segurança completo de um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            
        Returns:
            Dicionário com o perfil de segurança
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.get_security_profile") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Buscar níveis de segurança atuais
                levels_rows = await self.db_pool.fetch(
                    """
                    SELECT mechanism, level, parameters, updated_at, expires_at, metadata 
                    FROM security_levels 
                    WHERE user_id = $1 AND tenant_id = $2
                    AND (context_id = $3 OR (context_id IS NULL AND $3 IS NULL))
                    """,
                    user_id, tenant_id, context_id
                )
                
                # Transformar para o formato de resposta
                security_levels = {}
                last_modified = None
                
                for row in levels_rows:
                    mechanism = row["mechanism"]
                    updated_at = row["updated_at"]
                    
                    # Atualizar última modificação
                    if not last_modified or updated_at > last_modified:
                        last_modified = updated_at
                    
                    # Converter campos JSON
                    parameters = json.loads(row["parameters"]) if row["parameters"] else {}
                    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
                    
                    # Adicionar ao dicionário de níveis
                    security_levels[mechanism] = {
                        "level": row["level"],
                        "parameters": parameters,
                        "updated_at": updated_at,
                        "expires_at": row["expires_at"],
                        "metadata": metadata
                    }
                
                # Construir resposta
                return {
                    "user_id": user_id,
                    "tenant_id": tenant_id,
                    "context_id": context_id,
                    "security_levels": security_levels,
                    "last_modified": last_modified
                }
            except Exception as e:
                logger.error(f"Erro ao obter perfil de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def get_security_level(
        self, 
        user_id: str, 
        tenant_id: str, 
        mechanism: SecurityMechanism,
        context_id: Optional[str] = None
    ) -> SecurityLevel:
        """
        Obtém o nível de segurança atual para um mecanismo específico.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            mechanism: Mecanismo de segurança
            context_id: ID do contexto (opcional)
            
        Returns:
            Nível de segurança atual
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.get_security_level") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                span.set_attribute("security.mechanism", mechanism.value)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Buscar nível de segurança
                level_str = await self.db_pool.fetchval(
                    """
                    SELECT level
                    FROM security_levels 
                    WHERE user_id = $1 AND tenant_id = $2 AND mechanism = $3
                    AND (context_id = $4 OR (context_id IS NULL AND $4 IS NULL))
                    """,
                    user_id, tenant_id, mechanism.value, context_id
                )
                
                if not level_str:
                    # Buscar nível padrão no tenant (sem contexto)
                    if context_id:
                        level_str = await self.db_pool.fetchval(
                            """
                            SELECT default_level
                            FROM security_mechanism_defaults 
                            WHERE tenant_id = $1 AND mechanism = $2 AND context_id IS NULL
                            """,
                            tenant_id, mechanism.value
                        )
                
                if not level_str:
                    # Buscar nível padrão global
                    level_str = await self.db_pool.fetchval(
                        """
                        SELECT default_level
                        FROM security_mechanism_defaults 
                        WHERE tenant_id IS NULL AND mechanism = $1 AND context_id IS NULL
                        """,
                        mechanism.value
                    )
                
                if not level_str:
                    # Valor padrão mais seguro
                    level_str = "STANDARD"
                
                # Converter para enum
                return SecurityLevel(level_str)
            except Exception as e:
                logger.error(f"Erro ao obter nível de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def get_active_triggers(
        self,
        tenant_id: Optional[str] = None,
        context_id: Optional[str] = None,
        region_code: Optional[str] = None
    ) -> List[ScalingTrigger]:
        """
        Obtém todos os gatilhos de escalonamento ativos.
        
        Args:
            tenant_id: ID do tenant para filtrar (opcional)
            context_id: ID do contexto para filtrar (opcional)
            region_code: Código da região para filtrar (opcional)
            
        Returns:
            Lista de gatilhos de escalonamento ativos
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.get_active_triggers") as span:
            try:
                # Construir consulta SQL base
                base_query = """
                SELECT id, name, enabled, tenant_specific, tenant_id, region_specific, region_code,
                       context_specific, context_id, condition_type, dimension, comparison,
                       threshold_value, scaling_direction, description, metadata
                FROM scaling_triggers
                WHERE enabled = true
                """
                
                # Parâmetros da consulta
                params = []
                param_index = 1
                
                # Adicionar filtros condicionais
                if tenant_id:
                    base_query += f" AND (tenant_id = ${param_index} OR tenant_id IS NULL)"
                    params.append(tenant_id)
                    param_index += 1
                    
                if context_id:
                    base_query += f" AND (context_id = ${param_index} OR context_id IS NULL)"
                    params.append(context_id)
                    param_index += 1
                    
                if region_code:
                    base_query += f" AND (region_code = ${param_index} OR region_code IS NULL)"
                    params.append(region_code)
                    param_index += 1
                
                # Executar consulta
                rows = await self.db_pool.fetch(base_query, *params)
                
                # Transformar resultados para objetos do modelo
                triggers = []
                for row in rows:
                    # Converter metadata de JSON
                    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
                    
                    # Criar objeto ScalingTrigger
                    trigger = ScalingTrigger(
                        id=row["id"],
                        name=row["name"],
                        enabled=row["enabled"],
                        tenant_specific=row["tenant_specific"],
                        tenant_id=row["tenant_id"],
                        region_specific=row["region_specific"],
                        region_code=row["region_code"],
                        context_specific=row["context_specific"],
                        context_id=row["context_id"],
                        condition_type=row["condition_type"],
                        dimension=row["dimension"],
                        comparison=row["comparison"],
                        threshold_value=row["threshold_value"],
                        scaling_direction=ScalingDirection(row["scaling_direction"]),
                        description=row["description"],
                        metadata=metadata
                    )
                    
                    triggers.append(trigger)
                
                return triggers
            except Exception as e:
                logger.error(f"Erro ao obter gatilhos ativos: {e}")
                span.record_exception(e)
                raise
    
    async def get_active_policies(
        self,
        tenant_id: Optional[str] = None,
        context_id: Optional[str] = None,
        region_code: Optional[str] = None
    ) -> List[ScalingPolicy]:
        """
        Obtém todas as políticas de escalonamento ativas.
        
        Args:
            tenant_id: ID do tenant para filtrar (opcional)
            context_id: ID do contexto para filtrar (opcional)
            region_code: Código da região para filtrar (opcional)
            
        Returns:
            Lista de políticas de escalonamento ativas
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.get_active_policies") as span:
            try:
                # Construir consulta SQL base
                base_query = """
                SELECT id, name, enabled, priority, tenant_id, region_code, context_id,
                       trigger_ids, adjustment_map, description, created_at, updated_at, metadata
                FROM scaling_policies
                WHERE enabled = true
                """
                
                # Parâmetros da consulta
                params = []
                param_index = 1
                
                # Adicionar filtros condicionais
                if tenant_id:
                    base_query += f" AND (tenant_id = ${param_index} OR tenant_id IS NULL)"
                    params.append(tenant_id)
                    param_index += 1
                    
                if context_id:
                    base_query += f" AND (context_id = ${param_index} OR context_id IS NULL)"
                    params.append(context_id)
                    param_index += 1
                    
                if region_code:
                    base_query += f" AND (region_code = ${param_index} OR region_code IS NULL)"
                    params.append(region_code)
                    param_index += 1
                
                # Adicionar ordenação por prioridade (mais alta primeiro)
                base_query += " ORDER BY priority DESC"
                
                # Executar consulta
                rows = await self.db_pool.fetch(base_query, *params)
                
                # Transformar resultados para objetos do modelo
                policies = []
                for row in rows:
                    # Converter campos JSON
                    trigger_ids = json.loads(row["trigger_ids"]) if row["trigger_ids"] else []
                    adjustment_map = json.loads(row["adjustment_map"]) if row["adjustment_map"] else {}
                    metadata = json.loads(row["metadata"]) if row["metadata"] else {}
                    
                    # Criar objeto ScalingPolicy
                    policy = ScalingPolicy(
                        id=row["id"],
                        name=row["name"],
                        enabled=row["enabled"],
                        priority=row["priority"],
                        tenant_id=row["tenant_id"],
                        region_code=row["region_code"],
                        context_id=row["context_id"],
                        trigger_ids=trigger_ids,
                        adjustment_map=adjustment_map,
                        description=row["description"],
                        created_at=row["created_at"],
                        updated_at=row["updated_at"],
                        metadata=metadata
                    )
                    
                    policies.append(policy)
                
                return policies
            except Exception as e:
                logger.error(f"Erro ao obter políticas ativas: {e}")
                span.record_exception(e)
                raise
    
    async def create_scaling_event(self, event: ScalingEvent) -> ScalingEvent:
        """
        Cria um novo evento de escalonamento.
        
        Args:
            event: Evento de escalonamento a ser criado
            
        Returns:
            Evento criado com ID atualizado
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.create_scaling_event") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", event.user_id)
                span.set_attribute("tenant.id", event.tenant_id)
                if event.context_id:
                    span.set_attribute("context.id", event.context_id)
                
                # Converter objetos para JSON
                dimension_scores_json = json.dumps(event.dimension_scores) if event.dimension_scores else None
                adjustments_json = json.dumps([a.dict() for a in event.adjustments]) if event.adjustments else None
                metadata_json = json.dumps(event.metadata) if event.metadata else None
                
                # Inserir no banco de dados
                row = await self.db_pool.fetchrow(
                    """
                    INSERT INTO scaling_events(
                        id, user_id, tenant_id, context_id, region_code,
                        trigger_id, policy_id, trust_score, dimension_scores,
                        scaling_direction, adjustments, event_time, expires_at, metadata
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
                    RETURNING id
                    """,
                    event.id or str(uuid.uuid4()),
                    event.user_id,
                    event.tenant_id,
                    event.context_id,
                    event.region_code,
                    event.trigger_id,
                    event.policy_id,
                    event.trust_score,
                    dimension_scores_json,
                    event.scaling_direction.value if event.scaling_direction else None,
                    adjustments_json,
                    event.event_time or datetime.now(),
                    event.expires_at,
                    metadata_json
                )
                
                # Atualizar ID do evento
                event.id = row["id"]
                
                return event
            except Exception as e:
                logger.error(f"Erro ao criar evento de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def update_security_level(
        self,
        user_id: str,
        tenant_id: str,
        mechanism: SecurityMechanism,
        level: SecurityLevel,
        parameters: Optional[Dict[str, Any]] = None,
        expires_at: Optional[datetime] = None,
        metadata: Optional[Dict[str, Any]] = None,
        context_id: Optional[str] = None
    ) -> bool:
        """
        Atualiza ou cria um nível de segurança para um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            mechanism: Mecanismo de segurança
            level: Novo nível de segurança
            parameters: Parâmetros do nível de segurança (opcional)
            expires_at: Data de expiração (opcional)
            metadata: Metadata adicional (opcional)
            context_id: ID do contexto (opcional)
            
        Returns:
            True se a operação foi bem-sucedida
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.update_security_level") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                span.set_attribute("security.mechanism", mechanism.value)
                span.set_attribute("security.level", level.value)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Converter para JSON
                parameters_json = json.dumps(parameters) if parameters else None
                metadata_json = json.dumps(metadata) if metadata else None
                
                # Verificar se o nível já existe
                exists = await self.db_pool.fetchval(
                    """
                    SELECT EXISTS(
                        SELECT 1 FROM security_levels 
                        WHERE user_id = $1 AND tenant_id = $2 AND mechanism = $3
                        AND (context_id = $4 OR (context_id IS NULL AND $4 IS NULL))
                    )
                    """,
                    user_id, tenant_id, mechanism.value, context_id
                )
                
                if exists:
                    # Atualizar nível existente
                    await self.db_pool.execute(
                        """
                        UPDATE security_levels 
                        SET level = $5, parameters = $6, updated_at = $7, expires_at = $8, metadata = $9
                        WHERE user_id = $1 AND tenant_id = $2 AND mechanism = $3
                        AND (context_id = $4 OR (context_id IS NULL AND $4 IS NULL))
                        """,
                        user_id, tenant_id, mechanism.value, context_id,
                        level.value, parameters_json, datetime.now(), expires_at, metadata_json
                    )
                else:
                    # Inserir novo nível
                    await self.db_pool.execute(
                        """
                        INSERT INTO security_levels(
                            user_id, tenant_id, context_id, mechanism, level, 
                            parameters, updated_at, expires_at, metadata
                        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                        """,
                        user_id, tenant_id, context_id, mechanism.value, level.value,
                        parameters_json, datetime.now(), expires_at, metadata_json
                    )
                
                return True
            except Exception as e:
                logger.error(f"Erro ao atualizar nível de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def revoke_scaling_event(self, event_id: str, reason: str) -> bool:
        """
        Revoga um evento de escalonamento (expira imediatamente).
        
        Args:
            event_id: ID do evento a ser revogado
            reason: Razão da revogação
            
        Returns:
            True se a operação foi bem-sucedida
        """
        with tracer.start_as_current_span("adaptive_scaling_repository.revoke_scaling_event") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("event.id", event_id)
                span.set_attribute("reason", reason)
                
                # Definir expiração para agora
                now = datetime.now()
                
                # Atualizar expiração do evento
                result = await self.db_pool.execute(
                    """
                    UPDATE scaling_events
                    SET expires_at = $2,
                        metadata = jsonb_set(
                            COALESCE(metadata::jsonb, '{}'::jsonb), 
                            '{revoked}', 'true'::jsonb
                        )::json,
                        metadata = jsonb_set(
                            COALESCE(metadata::jsonb, '{}'::jsonb), 
                            '{revoke_reason}', $3::jsonb
                        )::json,
                        metadata = jsonb_set(
                            COALESCE(metadata::jsonb, '{}'::jsonb), 
                            '{revoked_at}', $4::jsonb
                        )::json
                    WHERE id = $1
                    """,
                    event_id, now, json.dumps(reason), json.dumps(now.isoformat())
                )
                
                return result == "UPDATE 1"
            except Exception as e:
                logger.error(f"Erro ao revogar evento de escalonamento: {e}")
                span.record_exception(e)
                raise