"""
Integração do serviço de escalonamento adaptativo com o servidor GraphQL.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import logging
from typing import Dict, Any

from graphql import GraphQLSchema
from ariadne import load_schema_from_path, make_executable_schema
from ariadne.asgi import GraphQL

from .resolvers import AdaptiveScalingResolvers
from ...services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from ...services.notification.notification_service import NotificationService
from ...services.audit.audit_service import AuditService
from ....app.repositories.trust_score_repository import TrustScoreRepository
from ....app.repositories.adaptive_scaling_repository import AdaptiveScalingRepository

# Configuração de logging
logger = logging.getLogger(__name__)

def register_adaptive_scaling_resolvers(
    schema: GraphQLSchema,
    resolvers_map: Dict[str, Dict[str, Any]],
    services: Dict[str, Any]
) -> None:
    """
    Registra os resolvers do serviço de escalonamento adaptativo no esquema GraphQL.
    
    Args:
        schema: Esquema GraphQL existente
        resolvers_map: Mapa de resolvers existente
        services: Serviços disponíveis para injeção
    """
    try:
        logger.info("Registrando resolvers do serviço de escalonamento adaptativo")
        
        # Obter serviços necessários
        adaptive_scaling_service = services.get("adaptive_scaling_service")
        trust_repository = services.get("trust_score_repository")
        notification_service = services.get("notification_service")
        audit_service = services.get("audit_service")
        
        # Verificar se serviços obrigatórios estão disponíveis
        if not adaptive_scaling_service or not trust_repository:
            logger.error(
                "Serviços necessários não estão disponíveis. "
                "AdaptiveScalingService e TrustScoreRepository são obrigatórios."
            )
            return
        
        # Criar resolvers
        resolvers = AdaptiveScalingResolvers(
            adaptive_scaling_service=adaptive_scaling_service,
            trust_repository=trust_repository
        )
        
        # Adicionar serviços opcionais se disponíveis
        if notification_service:
            resolvers.notification_service = notification_service
        
        if audit_service:
            resolvers.audit_service = audit_service
        
        # Registrar resolvers de Query
        query_resolvers = resolvers_map.get("Query", {})
        query_resolvers["userSecurityProfile"] = resolvers.get_user_security_profile
        query_resolvers["currentSecurityLevel"] = resolvers.get_current_security_level
        query_resolvers["scalingEvents"] = resolvers.get_scaling_events
        query_resolvers["scalingEventById"] = resolvers.get_scaling_event_by_id
        query_resolvers["scalingTriggers"] = resolvers.get_scaling_triggers
        query_resolvers["scalingTriggerById"] = resolvers.get_scaling_trigger_by_id
        query_resolvers["scalingPolicies"] = resolvers.get_scaling_policies
        query_resolvers["scalingPolicyById"] = resolvers.get_scaling_policy_by_id
        query_resolvers["adaptiveScalingStatus"] = resolvers.get_adaptive_scaling_status
        resolvers_map["Query"] = query_resolvers
        
        # Registrar resolvers de Mutation
        mutation_resolvers = resolvers_map.get("Mutation", {})
        mutation_resolvers["applySecurityAdjustment"] = resolvers.apply_security_adjustment
        mutation_resolvers["upsertScalingTrigger"] = resolvers.upsert_scaling_trigger
        mutation_resolvers["upsertScalingPolicy"] = resolvers.upsert_scaling_policy
        mutation_resolvers["setScalingTriggerEnabled"] = resolvers.set_scaling_trigger_enabled
        mutation_resolvers["setScalingPolicyEnabled"] = resolvers.set_scaling_policy_enabled
        mutation_resolvers["deleteScalingTrigger"] = resolvers.delete_scaling_trigger
        mutation_resolvers["deleteScalingPolicy"] = resolvers.delete_scaling_policy
        mutation_resolvers["revokeScalingEvent"] = resolvers.revoke_scaling_event
        resolvers_map["Mutation"] = mutation_resolvers
        
        logger.info("Resolvers do serviço de escalonamento adaptativo registrados com sucesso")
    except Exception as e:
        logger.error(f"Erro ao registrar resolvers do serviço de escalonamento adaptativo: {e}")
        raise

def load_adaptive_scaling_schema() -> str:
    """
    Carrega o esquema GraphQL do serviço de escalonamento adaptativo.
    
    Returns:
        Conteúdo do esquema GraphQL
    """
    try:
        # Determinar caminho do arquivo de schema
        current_dir = os.path.dirname(os.path.abspath(__file__))
        schema_path = os.path.join(current_dir, "schema.graphql")
        
        # Carregar conteúdo do schema
        with open(schema_path, "r") as schema_file:
            schema_content = schema_file.read()
        
        return schema_content
    except Exception as e:
        logger.error(f"Erro ao carregar esquema GraphQL do escalonamento adaptativo: {e}")
        raise

def initialize_adaptive_scaling_module(app_context: Dict[str, Any]) -> None:
    """
    Inicializa o módulo de escalonamento adaptativo.
    
    Args:
        app_context: Contexto da aplicação com serviços e configurações
    """
    try:
        logger.info("Inicializando módulo de escalonamento adaptativo")
        
        # Obter pool de conexões
        db_pool = app_context.get("db_pool")
        if not db_pool:
            logger.error("Pool de conexões não disponível")
            return
        
        # Criar repositório
        adaptive_scaling_repository = AdaptiveScalingRepository(db_pool)
        app_context["adaptive_scaling_repository"] = adaptive_scaling_repository
        
        # Obter serviços existentes
        trust_repository = app_context.get("trust_score_repository")
        notification_service = app_context.get("notification_service")
        
        if not trust_repository:
            logger.error("Repositório de confiança não disponível")
            return
        
        # Criar serviço
        config = app_context.get("config", {}).get("adaptive_scaling", {})
        adaptive_scaling_service = AdaptiveScalingService(
            db_pool=db_pool,
            trust_repository=trust_repository,
            notification_service=notification_service,
            config=config
        )
        
        # Inicializar serviço
        logger.info("Inicializando serviço de escalonamento adaptativo")
        app_context["adaptive_scaling_service"] = adaptive_scaling_service
        
        # Inicializar cache do serviço
        logger.info("Carregando cache do serviço de escalonamento adaptativo")
        # A inicialização do cache deve ser feita de forma assíncrona durante a inicialização da aplicação
        
        logger.info("Módulo de escalonamento adaptativo inicializado com sucesso")
    except Exception as e:
        logger.error(f"Erro ao inicializar módulo de escalonamento adaptativo: {e}")
        raise