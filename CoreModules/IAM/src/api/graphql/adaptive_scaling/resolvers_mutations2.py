"""
Resolvers GraphQL adicionais para mutations do serviço de escalonamento adaptativo.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import uuid
import logging
from typing import Dict, List, Any, Optional, Union
from datetime import datetime, timedelta

from opentelemetry import trace
from graphql import GraphQLResolveInfo

from .helpers import check_permissions

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingPolicyMutations:
    """
    Resolvers para mutations relacionadas a políticas de escalonamento.
    """
    
    def __init__(self, adaptive_scaling_service, db_pool, audit_service=None, notification_service=None):
        """
        Inicializa os resolvers de mutations para políticas.
        
        Args:
            adaptive_scaling_service: Serviço de escalonamento adaptativo
            db_pool: Pool de conexões ao banco de dados
            audit_service: Serviço de auditoria (opcional)
            notification_service: Serviço de notificações (opcional)
        """
        self.adaptive_scaling_service = adaptive_scaling_service
        self.db_pool = db_pool
        self.audit_service = audit_service
        self.notification_service = notification_service
    
    async def upsert_scaling_policy(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        input: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Cria ou atualiza uma política de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            input: Dados da política
            
        Returns:
            Política criada ou atualizada
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.upsert_scaling_policy") as span:
            try:
                # Verificar se é uma criação ou atualização
                policy_id = input.get("id")
                is_update = policy_id is not None
                
                # Verificar permissão apropriada
                tenant_id = input.get("tenantId")
                
                if tenant_id:
                    # Política específica de tenant
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    # Política global
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Adicionar atributos ao span
                if is_update:
                    span.set_attribute("policy.id", policy_id)
                    span.set_attribute("operation", "update")
                else:
                    span.set_attribute("operation", "create")
                
                if tenant_id:
                    span.set_attribute("tenant.id", tenant_id)
                
                # Validar IDs de gatilhos
                trigger_ids = input.get("triggerIds", [])
                if trigger_ids:
                    for trigger_id in trigger_ids:
                        # Verificar se o gatilho existe
                        trigger_exists = await self.db_pool.fetchval(
                            "SELECT EXISTS(SELECT 1 FROM scaling_triggers WHERE id = $1)",
                            trigger_id
                        )
                        
                        if not trigger_exists:
                            raise ValueError(f"Gatilho com ID {trigger_id} não encontrado")
                
                # Converter dados de entrada para formato do banco
                policy_data = {
                    "name": input["name"],
                    "enabled": input.get("enabled", True),
                    "priority": input["priority"],
                    "tenant_id": tenant_id,
                    "region_code": input.get("regionCode"),
                    "context_id": input.get("contextId"),
                    "trigger_ids": json.dumps(trigger_ids),
                    "adjustment_map": json.dumps(input["adjustmentMap"]),
                    "description": input.get("description", ""),
                    "created_at": datetime.now(),
                    "updated_at": datetime.now(),
                    "metadata": json.dumps(input.get("metadata", {}))
                }
                
                # Executar a operação de banco
                if is_update:
                    # Verificar se a política existe
                    existing = await self.db_pool.fetchrow(
                        "SELECT id, tenant_id FROM scaling_policies WHERE id = $1",
                        policy_id
                    )
                    
                    if not existing:
                        raise ValueError(f"Política com ID {policy_id} não encontrada")
                    
                    # Verificar se tem permissão para editar esta política específica
                    existing_tenant_id = existing["tenant_id"]
                    if existing_tenant_id and existing_tenant_id != tenant_id:
                        await check_permissions(info.context, "adaptive_scaling:admin")
                    
                    # Preservar data de criação original
                    created_at = await self.db_pool.fetchval(
                        "SELECT created_at FROM scaling_policies WHERE id = $1",
                        policy_id
                    )
                    policy_data["created_at"] = created_at
                    
                    # Atualizar política
                    await self.db_pool.execute(
                        """
                        UPDATE scaling_policies
                        SET name = $2, enabled = $3, priority = $4, tenant_id = $5,
                            region_code = $6, context_id = $7, trigger_ids = $8,
                            adjustment_map = $9, description = $10, created_at = $11,
                            updated_at = $12, metadata = $13
                        WHERE id = $1
                        """,
                        policy_id, policy_data["name"], policy_data["enabled"],
                        policy_data["priority"], policy_data["tenant_id"],
                        policy_data["region_code"], policy_data["context_id"],
                        policy_data["trigger_ids"], policy_data["adjustment_map"],
                        policy_data["description"], policy_data["created_at"],
                        policy_data["updated_at"], policy_data["metadata"]
                    )
                else:
                    # Criar novo ID
                    policy_id = str(uuid.uuid4())
                    
                    # Criar política
                    await self.db_pool.execute(
                        """
                        INSERT INTO scaling_policies(
                            id, name, enabled, priority, tenant_id, region_code, context_id,
                            trigger_ids, adjustment_map, description, created_at, updated_at, metadata
                        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
                        """,
                        policy_id, policy_data["name"], policy_data["enabled"],
                        policy_data["priority"], policy_data["tenant_id"],
                        policy_data["region_code"], policy_data["context_id"],
                        policy_data["trigger_ids"], policy_data["adjustment_map"],
                        policy_data["description"], policy_data["created_at"],
                        policy_data["updated_at"], policy_data["metadata"]
                    )
                
                # Atualizar cache do serviço
                await self.adaptive_scaling_service.refresh_policies_cache()
                
                # Obter dados atualizados
                row = await self.db_pool.fetchrow(
                    """
                    SELECT id, name, enabled, priority, tenant_id, region_code, context_id,
                           trigger_ids, adjustment_map, description, created_at, updated_at, metadata
                    FROM scaling_policies
                    WHERE id = $1
                    """,
                    policy_id
                )
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type=f"scaling_policy_{is_update and 'updated' or 'created'}",
                        tenant_id=tenant_id or "global",
                        user_id=None,
                        actor_id=info.context.user.get("id"),
                        details={
                            "policy_id": policy_id,
                            "policy_name": input["name"],
                            "priority": input["priority"],
                            "trigger_count": len(trigger_ids),
                            "context_id": input.get("contextId"),
                            "region_code": input.get("regionCode")
                        }
                    )
                
                # Converter para formato GraphQL e retornar
                result = dict(row)
                result["trigger_ids"] = json.loads(result["trigger_ids"]) if result["trigger_ids"] else []
                result["adjustment_map"] = json.loads(result["adjustment_map"]) if result["adjustment_map"] else {}
                result["metadata"] = json.loads(result["metadata"]) if result["metadata"] else {}
                
                return {
                    "id": result["id"],
                    "name": result["name"],
                    "enabled": result["enabled"],
                    "priority": result["priority"],
                    "tenantId": result["tenant_id"],
                    "regionCode": result["region_code"],
                    "contextId": result["context_id"],
                    "triggerIds": result["trigger_ids"],
                    "adjustmentMap": result["adjustment_map"],
                    "description": result["description"],
                    "createdAt": result["created_at"],
                    "updatedAt": result["updated_at"],
                    "metadata": result["metadata"]
                }
            except Exception as e:
                logger.error(f"Erro ao criar/atualizar política de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def set_scaling_policy_enabled(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        policy_id: str,
        enabled: bool
    ) -> Dict[str, Any]:
        """
        Ativa ou desativa uma política de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            policy_id: ID da política
            enabled: Estado de ativação
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.set_scaling_policy_enabled") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("policy.id", policy_id)
                span.set_attribute("enabled", enabled)
                
                # Verificar se a política existe
                policy = await self.db_pool.fetchrow(
                    "SELECT id, tenant_id FROM scaling_policies WHERE id = $1",
                    policy_id
                )
                
                if not policy:
                    raise ValueError(f"Política com ID {policy_id} não encontrada")
                
                # Verificar permissão
                tenant_id = policy["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Atualizar estado
                await self.db_pool.execute(
                    "UPDATE scaling_policies SET enabled = $2 WHERE id = $1",
                    policy_id, enabled
                )
                
                # Atualizar cache do serviço
                await self.adaptive_scaling_service.refresh_policies_cache()
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type="scaling_policy_state_changed",
                        tenant_id=tenant_id or "global",
                        user_id=None,
                        actor_id=info.context.user.get("id"),
                        details={
                            "policy_id": policy_id,
                            "enabled": enabled
                        }
                    )
                
                return {
                    "success": True,
                    "policyId": policy_id,
                    "enabled": enabled
                }
            except Exception as e:
                logger.error(f"Erro ao alterar estado da política: {e}")
                span.record_exception(e)
                raise
    
    async def delete_scaling_policy(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        policy_id: str
    ) -> Dict[str, Any]:
        """
        Remove uma política de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            policy_id: ID da política
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.delete_scaling_policy") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("policy.id", policy_id)
                
                # Verificar se a política existe
                policy = await self.db_pool.fetchrow(
                    "SELECT id, tenant_id, name FROM scaling_policies WHERE id = $1",
                    policy_id
                )
                
                if not policy:
                    raise ValueError(f"Política com ID {policy_id} não encontrada")
                
                # Verificar permissão
                tenant_id = policy["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Verificar se há eventos recentes que usaram esta política
                events_count = await self.db_pool.fetchval(
                    """
                    SELECT COUNT(*)
                    FROM scaling_events
                    WHERE policy_id = $1 AND event_time > NOW() - INTERVAL '30 days'
                    """,
                    policy_id
                )
                
                # Se houver eventos recentes, apenas desativar a política em vez de excluir
                if events_count > 0:
                    await self.db_pool.execute(
                        "UPDATE scaling_policies SET enabled = false WHERE id = $1",
                        policy_id
                    )
                    
                    logger.info(f"Política {policy_id} desativada em vez de excluída devido a {events_count} eventos recentes")
                    
                    # Registrar na auditoria, se disponível
                    if self.audit_service:
                        await self.audit_service.log_security_event(
                            event_type="scaling_policy_disabled",
                            tenant_id=tenant_id or "global",
                            user_id=None,
                            actor_id=info.context.user.get("id"),
                            details={
                                "policy_id": policy_id,
                                "policy_name": policy["name"],
                                "reason": f"Desativada em vez de excluída devido a {events_count} eventos recentes"
                            }
                        )
                    
                    return {
                        "success": True,
                        "policyId": policy_id,
                        "message": f"Política desativada em vez de excluída devido a {events_count} eventos recentes"
                    }
                else:
                    # Remover política
                    await self.db_pool.execute(
                        "DELETE FROM scaling_policies WHERE id = $1",
                        policy_id
                    )
                    
                    # Atualizar cache do serviço
                    await self.adaptive_scaling_service.refresh_policies_cache()
                    
                    # Registrar na auditoria, se disponível
                    if self.audit_service:
                        await self.audit_service.log_security_event(
                            event_type="scaling_policy_deleted",
                            tenant_id=tenant_id or "global",
                            user_id=None,
                            actor_id=info.context.user.get("id"),
                            details={
                                "policy_id": policy_id,
                                "policy_name": policy["name"]
                            }
                        )
                    
                    return {
                        "success": True,
                        "policyId": policy_id
                    }
            except ValueError as e:
                # Erros de validação esperados
                logger.warning(f"Erro de validação ao excluir política: {e}")
                return {
                    "success": False,
                    "policyId": policy_id,
                    "error": str(e)
                }
            except Exception as e:
                logger.error(f"Erro ao excluir política: {e}")
                span.record_exception(e)
                raise
    
    async def revoke_scaling_event(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        event_id: str,
        reason: str
    ) -> Dict[str, Any]:
        """
        Revoga um evento de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            event_id: ID do evento
            reason: Razão da revogação
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.revoke_scaling_event") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("event.id", event_id)
                span.set_attribute("reason", reason)
                
                # Buscar evento
                event = await self.db_pool.fetchrow(
                    """
                    SELECT id, user_id, tenant_id, scaling_direction, expires_at
                    FROM scaling_events
                    WHERE id = $1
                    """,
                    event_id
                )
                
                if not event:
                    raise ValueError(f"Evento com ID {event_id} não encontrado")
                
                # Verificar se o evento já expirou
                if event["expires_at"] and event["expires_at"] < datetime.now():
                    raise ValueError("Este evento já expirou e não pode ser revogado")
                
                # Verificar permissão
                tenant_id = event["tenant_id"]
                await check_permissions(info.context, "adaptive_scaling:admin", tenant_id)
                
                # Revogar evento
                success = await self.adaptive_scaling_service.revoke_scaling_event(
                    event_id=event_id,
                    reason=reason,
                    admin_id=info.context.user.get("id")
                )
                
                if success:
                    # Registrar na auditoria, se disponível
                    if self.audit_service:
                        await self.audit_service.log_security_event(
                            event_type="scaling_event_revoked",
                            tenant_id=tenant_id,
                            user_id=event["user_id"],
                            actor_id=info.context.user.get("id"),
                            details={
                                "event_id": event_id,
                                "reason": reason,
                                "scaling_direction": event["scaling_direction"]
                            }
                        )
                    
                    # Enviar notificação, se disponível
                    if self.notification_service:
                        await self.notification_service.send_notification(
                            tenant_id=tenant_id,
                            user_id=event["user_id"],
                            notification_type="security_event_revoked",
                            title="Evento de segurança revogado",
                            message=f"Um ajuste de segurança foi revogado: {reason}",
                            details={
                                "event_id": event_id,
                                "reason": reason
                            }
                        )
                    
                    return {
                        "success": True,
                        "eventId": event_id,
                        "userId": event["user_id"],
                        "tenantId": tenant_id
                    }
                else:
                    return {
                        "success": False,
                        "eventId": event_id,
                        "error": "Falha ao revogar evento"
                    }
            except ValueError as e:
                # Erros de validação esperados
                logger.warning(f"Erro de validação ao revogar evento: {e}")
                return {
                    "success": False,
                    "eventId": event_id,
                    "error": str(e)
                }
            except Exception as e:
                logger.error(f"Erro ao revogar evento: {e}")
                span.record_exception(e)
                raise