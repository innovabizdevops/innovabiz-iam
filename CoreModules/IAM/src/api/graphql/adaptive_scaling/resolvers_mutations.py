"""
Resolvers GraphQL para mutations do serviço de escalonamento adaptativo.

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
from ...services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from ...services.adaptive_scaling.models import (
    SecurityLevel,
    SecurityMechanism,
    ScalingDirection,
    SecurityAdjustment
)

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingMutations:
    """
    Resolvers para mutations do serviço de escalonamento adaptativo.
    """
    
    def __init__(self, adaptive_scaling_service, db_pool, audit_service=None, notification_service=None):
        """
        Inicializa os resolvers de mutations.
        
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
    
    #
    # Resolvers de Mutation
    #
    
    async def apply_security_adjustment(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        tenant_id: str,
        user_id: str,
        adjustment: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Aplica um ajuste de segurança manual para um usuário.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            tenant_id: ID do tenant
            user_id: ID do usuário
            adjustment: Detalhes do ajuste de segurança
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.apply_security_adjustment") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                span.set_attribute("security.mechanism", adjustment["mechanism"])
                span.set_attribute("security.level", adjustment["level"])
                if adjustment.get("contextId"):
                    span.set_attribute("context.id", adjustment["contextId"])
                
                # Verificar permissão
                await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                
                # Converter para objetos do modelo
                mechanism = SecurityMechanism(adjustment["mechanism"])
                level = SecurityLevel(adjustment["level"])
                context_id = adjustment.get("contextId")
                expires_at = adjustment.get("expiresAt")
                parameters = adjustment.get("parameters", {})
                reason = adjustment.get("reason", "Manual adjustment by administrator")
                
                # Definir expiração padrão se não especificada (30 dias)
                if not expires_at:
                    expires_at = datetime.now() + timedelta(days=30)
                
                # Criar objeto de ajuste de segurança
                security_adjustment = SecurityAdjustment(
                    mechanism=mechanism,
                    current_level=None,  # Será determinado pelo serviço
                    new_level=level,
                    reason=reason,
                    parameters=parameters,
                    expires_at=expires_at
                )
                
                # Aplicar ajuste de segurança
                result = await self.adaptive_scaling_service.apply_manual_security_adjustment(
                    user_id=user_id,
                    tenant_id=tenant_id,
                    adjustment=security_adjustment,
                    context_id=context_id,
                    admin_id=info.context.user.get("id"),
                    source="api"
                )
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type="security_adjustment",
                        tenant_id=tenant_id,
                        user_id=user_id,
                        actor_id=info.context.user.get("id"),
                        details={
                            "mechanism": adjustment["mechanism"],
                            "old_level": result.get("previous_level"),
                            "new_level": adjustment["level"],
                            "reason": reason,
                            "expires_at": expires_at.isoformat() if expires_at else None,
                            "context_id": context_id
                        }
                    )
                
                # Enviar notificação, se disponível
                if self.notification_service:
                    await self.notification_service.send_notification(
                        tenant_id=tenant_id,
                        user_id=user_id,
                        notification_type="security_level_changed",
                        title="Nível de segurança alterado",
                        message=f"Seu nível de segurança para {adjustment['mechanism']} foi alterado para {adjustment['level']}",
                        details={
                            "mechanism": adjustment["mechanism"],
                            "level": adjustment["level"],
                            "expires_at": expires_at.isoformat() if expires_at else None
                        }
                    )
                
                # Construir resposta
                return {
                    "success": True,
                    "userId": user_id,
                    "tenantId": tenant_id,
                    "contextId": context_id,
                    "mechanism": adjustment["mechanism"],
                    "previousLevel": result.get("previous_level"),
                    "newLevel": adjustment["level"],
                    "eventId": result.get("event_id"),
                    "expiresAt": expires_at
                }
            except Exception as e:
                logger.error(f"Erro ao aplicar ajuste de segurança: {e}")
                span.record_exception(e)
                raise
    
    async def upsert_scaling_trigger(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        input: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Cria ou atualiza um gatilho de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            input: Dados do gatilho
            
        Returns:
            Gatilho criado ou atualizado
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.upsert_scaling_trigger") as span:
            try:
                # Verificar se é uma criação ou atualização
                trigger_id = input.get("id")
                is_update = trigger_id is not None
                
                # Verificar permissão apropriada
                tenant_id = input.get("tenantId")
                
                if tenant_id:
                    # Gatilho específico de tenant
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    # Gatilho global
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Adicionar atributos ao span
                if is_update:
                    span.set_attribute("trigger.id", trigger_id)
                    span.set_attribute("operation", "update")
                else:
                    span.set_attribute("operation", "create")
                
                if tenant_id:
                    span.set_attribute("tenant.id", tenant_id)
                
                # Converter dados de entrada para formato do banco
                trigger_data = {
                    "name": input["name"],
                    "enabled": input.get("enabled", True),
                    "tenant_specific": tenant_id is not None,
                    "tenant_id": tenant_id,
                    "region_specific": input.get("regionCode") is not None,
                    "region_code": input.get("regionCode"),
                    "context_specific": input.get("contextId") is not None,
                    "context_id": input.get("contextId"),
                    "condition_type": input["conditionType"],
                    "dimension": input["dimension"],
                    "comparison": input["comparison"],
                    "threshold_value": input["thresholdValue"],
                    "scaling_direction": input["scalingDirection"],
                    "description": input.get("description", ""),
                    "metadata": json.dumps(input.get("metadata", {}))
                }
                
                # Executar a operação de banco
                if is_update:
                    # Verificar se o gatilho existe
                    existing = await self.db_pool.fetchrow(
                        "SELECT id, tenant_id FROM scaling_triggers WHERE id = $1",
                        trigger_id
                    )
                    
                    if not existing:
                        raise ValueError(f"Gatilho com ID {trigger_id} não encontrado")
                    
                    # Verificar se tem permissão para editar este gatilho específico
                    existing_tenant_id = existing["tenant_id"]
                    if existing_tenant_id and existing_tenant_id != tenant_id:
                        await check_permissions(info.context, "adaptive_scaling:admin")
                    
                    # Atualizar gatilho
                    await self.db_pool.execute(
                        """
                        UPDATE scaling_triggers
                        SET name = $2, enabled = $3, tenant_specific = $4, tenant_id = $5,
                            region_specific = $6, region_code = $7, context_specific = $8, context_id = $9,
                            condition_type = $10, dimension = $11, comparison = $12, threshold_value = $13,
                            scaling_direction = $14, description = $15, metadata = $16
                        WHERE id = $1
                        """,
                        trigger_id, trigger_data["name"], trigger_data["enabled"],
                        trigger_data["tenant_specific"], trigger_data["tenant_id"],
                        trigger_data["region_specific"], trigger_data["region_code"],
                        trigger_data["context_specific"], trigger_data["context_id"],
                        trigger_data["condition_type"], trigger_data["dimension"],
                        trigger_data["comparison"], trigger_data["threshold_value"],
                        trigger_data["scaling_direction"], trigger_data["description"],
                        trigger_data["metadata"]
                    )
                else:
                    # Criar novo ID
                    trigger_id = str(uuid.uuid4())
                    
                    # Criar gatilho
                    await self.db_pool.execute(
                        """
                        INSERT INTO scaling_triggers(
                            id, name, enabled, tenant_specific, tenant_id,
                            region_specific, region_code, context_specific, context_id,
                            condition_type, dimension, comparison, threshold_value,
                            scaling_direction, description, metadata
                        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
                        """,
                        trigger_id, trigger_data["name"], trigger_data["enabled"],
                        trigger_data["tenant_specific"], trigger_data["tenant_id"],
                        trigger_data["region_specific"], trigger_data["region_code"],
                        trigger_data["context_specific"], trigger_data["context_id"],
                        trigger_data["condition_type"], trigger_data["dimension"],
                        trigger_data["comparison"], trigger_data["threshold_value"],
                        trigger_data["scaling_direction"], trigger_data["description"],
                        trigger_data["metadata"]
                    )
                
                # Atualizar cache do serviço
                await self.adaptive_scaling_service.refresh_triggers_cache()
                
                # Obter dados atualizados
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
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type=f"scaling_trigger_{is_update and 'updated' or 'created'}",
                        tenant_id=tenant_id or "global",
                        user_id=None,
                        actor_id=info.context.user.get("id"),
                        details={
                            "trigger_id": trigger_id,
                            "trigger_name": input["name"],
                            "dimension": input["dimension"],
                            "condition": f"{input['comparison']} {input['thresholdValue']}",
                            "scaling_direction": input["scalingDirection"],
                            "context_id": input.get("contextId"),
                            "region_code": input.get("regionCode")
                        }
                    )
                
                # Converter para formato GraphQL e retornar
                result = dict(row)
                result["metadata"] = json.loads(result["metadata"]) if result["metadata"] else {}
                
                return {
                    "id": result["id"],
                    "name": result["name"],
                    "enabled": result["enabled"],
                    "tenantSpecific": result["tenant_specific"],
                    "tenantId": result["tenant_id"],
                    "regionSpecific": result["region_specific"],
                    "regionCode": result["region_code"],
                    "contextSpecific": result["context_specific"],
                    "contextId": result["context_id"],
                    "conditionType": result["condition_type"],
                    "dimension": result["dimension"],
                    "comparison": result["comparison"],
                    "thresholdValue": result["threshold_value"],
                    "scalingDirection": result["scaling_direction"],
                    "description": result["description"],
                    "metadata": result["metadata"]
                }
            except Exception as e:
                logger.error(f"Erro ao criar/atualizar gatilho de escalonamento: {e}")
                span.record_exception(e)
                raise
    
    async def set_scaling_trigger_enabled(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        trigger_id: str,
        enabled: bool
    ) -> Dict[str, Any]:
        """
        Ativa ou desativa um gatilho de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            trigger_id: ID do gatilho
            enabled: Estado de ativação
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.set_scaling_trigger_enabled") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("trigger.id", trigger_id)
                span.set_attribute("enabled", enabled)
                
                # Verificar se o gatilho existe
                trigger = await self.db_pool.fetchrow(
                    "SELECT id, tenant_id FROM scaling_triggers WHERE id = $1",
                    trigger_id
                )
                
                if not trigger:
                    raise ValueError(f"Gatilho com ID {trigger_id} não encontrado")
                
                # Verificar permissão
                tenant_id = trigger["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Atualizar estado
                await self.db_pool.execute(
                    "UPDATE scaling_triggers SET enabled = $2 WHERE id = $1",
                    trigger_id, enabled
                )
                
                # Atualizar cache do serviço
                await self.adaptive_scaling_service.refresh_triggers_cache()
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type="scaling_trigger_state_changed",
                        tenant_id=tenant_id or "global",
                        user_id=None,
                        actor_id=info.context.user.get("id"),
                        details={
                            "trigger_id": trigger_id,
                            "enabled": enabled
                        }
                    )
                
                return {
                    "success": True,
                    "triggerId": trigger_id,
                    "enabled": enabled
                }
            except Exception as e:
                logger.error(f"Erro ao alterar estado do gatilho: {e}")
                span.record_exception(e)
                raise
    
    async def delete_scaling_trigger(
        self, 
        _: Any, 
        info: GraphQLResolveInfo,
        trigger_id: str
    ) -> Dict[str, Any]:
        """
        Remove um gatilho de escalonamento.
        
        Args:
            _: Objeto raiz (não utilizado)
            info: Informações de contexto da requisição GraphQL
            trigger_id: ID do gatilho
            
        Returns:
            Resultado da operação
        """
        with tracer.start_as_current_span("adaptive_scaling_mutations.delete_scaling_trigger") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("trigger.id", trigger_id)
                
                # Verificar se o gatilho existe
                trigger = await self.db_pool.fetchrow(
                    "SELECT id, tenant_id, name FROM scaling_triggers WHERE id = $1",
                    trigger_id
                )
                
                if not trigger:
                    raise ValueError(f"Gatilho com ID {trigger_id} não encontrado")
                
                # Verificar permissão
                tenant_id = trigger["tenant_id"]
                if tenant_id:
                    await check_permissions(info.context, "adaptive_scaling:write", tenant_id)
                else:
                    await check_permissions(info.context, "adaptive_scaling:admin")
                
                # Verificar se o gatilho está em uso por alguma política
                is_used = await self.db_pool.fetchval(
                    "SELECT EXISTS(SELECT 1 FROM scaling_policies WHERE trigger_ids @> $1)",
                    json.dumps([trigger_id])
                )
                
                if is_used:
                    raise ValueError("Este gatilho está sendo usado por uma ou mais políticas e não pode ser excluído")
                
                # Remover gatilho
                await self.db_pool.execute(
                    "DELETE FROM scaling_triggers WHERE id = $1",
                    trigger_id
                )
                
                # Atualizar cache do serviço
                await self.adaptive_scaling_service.refresh_triggers_cache()
                
                # Registrar na auditoria, se disponível
                if self.audit_service:
                    await self.audit_service.log_security_event(
                        event_type="scaling_trigger_deleted",
                        tenant_id=tenant_id or "global",
                        user_id=None,
                        actor_id=info.context.user.get("id"),
                        details={
                            "trigger_id": trigger_id,
                            "trigger_name": trigger["name"]
                        }
                    )
                
                return {
                    "success": True,
                    "triggerId": trigger_id
                }
            except ValueError as e:
                # Erros de validação esperados
                logger.warning(f"Erro de validação ao excluir gatilho: {e}")
                return {
                    "success": False,
                    "triggerId": trigger_id,
                    "error": str(e)
                }
            except Exception as e:
                logger.error(f"Erro ao excluir gatilho: {e}")
                span.record_exception(e)
                raise