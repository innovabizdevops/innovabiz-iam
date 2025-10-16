"""
Módulo de integração para escalonamento adaptativo GraphQL.

Este módulo fornece funções para registrar resolvers GraphQL do escalonamento adaptativo
com o servidor GraphQL principal, carregar o esquema GraphQL e inicializar 
os componentes necessários.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
from typing import Dict, Any, Optional
import asyncio

from ariadne import load_schema_from_path, make_executable_schema
from opentelemetry import trace

from .resolvers_combined import AdaptiveScalingGraphQLResolvers
from ...services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from ....app.repositories.trust_score_repository import TrustScoreRepository

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingGraphQLIntegration:
    """
    Classe de integração para o escalonamento adaptativo GraphQL.
    """

    def __init__(
        self,
        app_context: Dict[str, Any],
        schema_path: Optional[str] = None
    ):
        """
        Inicializa a integração.
        
        Args:
            app_context: Contexto da aplicação com pools e serviços
            schema_path: Caminho para o arquivo de esquema GraphQL (opcional)
        """
        self.app_context = app_context
        self.schema_path = schema_path or os.path.join(
            os.path.dirname(__file__), "schema.graphql"
        )
        self.resolvers = None
        
        # Verificar serviços necessários
        if "adaptive_scaling_service" not in app_context:
            raise ValueError("Serviço de escalonamento adaptativo não encontrado no contexto da aplicação")
        
        if "trust_score_repository" not in app_context:
            raise ValueError("Repositório de TrustScore não encontrado no contexto da aplicação")

    def initialize(self):
        """
        Inicializa a integração.
        
        Returns:
            AdaptiveScalingGraphQLResolvers: Objeto de resolvers
        """
        with tracer.start_as_current_span("adaptive_scaling.initialize_integration"):
            try:
                adaptive_scaling_service = self.app_context["adaptive_scaling_service"]
                trust_repository = self.app_context["trust_score_repository"]
                
                # Criar instância dos resolvers
                self.resolvers = AdaptiveScalingGraphQLResolvers(
                    adaptive_scaling_service=adaptive_scaling_service,
                    trust_repository=trust_repository
                )
                
                # Configurar serviços opcionais
                if "notification_service" in self.app_context:
                    self.resolvers.set_notification_service(
                        self.app_context["notification_service"]
                    )
                
                if "audit_service" in self.app_context:
                    self.resolvers.set_audit_service(
                        self.app_context["audit_service"]
                    )
                
                logger.info("Integração GraphQL de escalonamento adaptativo inicializada com sucesso")
                return self.resolvers
                
            except Exception as e:
                logger.error(f"Erro ao inicializar integração GraphQL de escalonamento adaptativo: {e}")
                raise

    def load_schema(self):
        """
        Carrega o esquema GraphQL.
        
        Returns:
            str: Esquema GraphQL
        """
        with tracer.start_as_current_span("adaptive_scaling.load_schema"):
            try:
                if not os.path.exists(self.schema_path):
                    raise FileNotFoundError(f"Arquivo de esquema GraphQL não encontrado: {self.schema_path}")
                
                schema = load_schema_from_path(self.schema_path)
                logger.info(f"Esquema GraphQL de escalonamento adaptativo carregado: {self.schema_path}")
                return schema
                
            except Exception as e:
                logger.error(f"Erro ao carregar esquema GraphQL de escalonamento adaptativo: {e}")
                raise

    async def register_resolvers(self, graphql_server):
        """
        Registra os resolvers com o servidor GraphQL.
        
        Args:
            graphql_server: Instância do servidor GraphQL
        """
        with tracer.start_as_current_span("adaptive_scaling.register_resolvers"):
            try:
                if not self.resolvers:
                    self.initialize()
                
                # Registrar resolvers de Query
                graphql_server.add_query_resolver(
                    "userSecurityProfile", 
                    self.resolvers.get_user_security_profile
                )
                
                graphql_server.add_query_resolver(
                    "currentSecurityLevel", 
                    self.resolvers.get_current_security_level
                )
                
                graphql_server.add_query_resolver(
                    "scalingEvents", 
                    self.resolvers.get_scaling_events
                )
                
                graphql_server.add_query_resolver(
                    "scalingEventById", 
                    self.resolvers.get_scaling_event_by_id
                )
                
                graphql_server.add_query_resolver(
                    "scalingTriggers", 
                    self.resolvers.get_scaling_triggers
                )
                
                graphql_server.add_query_resolver(
                    "scalingTriggerById", 
                    self.resolvers.get_scaling_trigger_by_id
                )
                
                graphql_server.add_query_resolver(
                    "scalingPolicies", 
                    self.resolvers.get_scaling_policies
                )
                
                graphql_server.add_query_resolver(
                    "scalingPolicyById", 
                    self.resolvers.get_scaling_policy_by_id
                )
                
                graphql_server.add_query_resolver(
                    "adaptiveScalingStatus", 
                    self.resolvers.get_adaptive_scaling_status
                )
                
                # Registrar resolvers de Mutation
                graphql_server.add_mutation_resolver(
                    "applySecurityAdjustment", 
                    self.resolvers.apply_security_adjustment
                )
                
                graphql_server.add_mutation_resolver(
                    "upsertScalingTrigger", 
                    self.resolvers.upsert_scaling_trigger
                )
                
                graphql_server.add_mutation_resolver(
                    "setScalingTriggerEnabled", 
                    self.resolvers.set_scaling_trigger_enabled
                )
                
                graphql_server.add_mutation_resolver(
                    "deleteScalingTrigger", 
                    self.resolvers.delete_scaling_trigger
                )
                
                graphql_server.add_mutation_resolver(
                    "upsertScalingPolicy", 
                    self.resolvers.upsert_scaling_policy
                )
                
                graphql_server.add_mutation_resolver(
                    "setScalingPolicyEnabled", 
                    self.resolvers.set_scaling_policy_enabled
                )
                
                graphql_server.add_mutation_resolver(
                    "deleteScalingPolicy", 
                    self.resolvers.delete_scaling_policy
                )
                
                graphql_server.add_mutation_resolver(
                    "revokeScalingEvent", 
                    self.resolvers.revoke_scaling_event
                )
                
                # Reiniciar o servidor para aplicar as alterações (se necessário)
                if hasattr(graphql_server, 'refresh_schema'):
                    await graphql_server.refresh_schema()
                
                logger.info("Resolvers GraphQL de escalonamento adaptativo registrados com sucesso")
                
            except Exception as e:
                logger.error(f"Erro ao registrar resolvers GraphQL de escalonamento adaptativo: {e}")
                raise

    @staticmethod
    async def initialize_adaptive_scaling(app_context):
        """
        Inicializa o módulo de escalonamento adaptativo e registra resolvers GraphQL.
        
        Args:
            app_context: Contexto da aplicação com pools e serviços
            
        Returns:
            dict: Contexto atualizado com serviços de escalonamento adaptativo
        """
        with tracer.start_as_current_span("adaptive_scaling.initialize_module"):
            try:
                # Verificar se já está inicializado
                if "adaptive_scaling_service" in app_context:
                    logger.info("Serviço de escalonamento adaptativo já inicializado")
                    return app_context
                
                # Verificar dependências
                if "db_pool" not in app_context:
                    raise ValueError("Pool de banco de dados não encontrado no contexto da aplicação")
                
                if "trust_score_repository" not in app_context:
                    # Criar repositório de TrustScore se não existir
                    logger.info("Criando repositório de TrustScore")
                    app_context["trust_score_repository"] = TrustScoreRepository(
                        app_context["db_pool"]
                    )
                
                # Criar serviço de escalonamento adaptativo
                logger.info("Criando serviço de escalonamento adaptativo")
                app_context["adaptive_scaling_service"] = AdaptiveScalingService(
                    db_pool=app_context["db_pool"],
                    trust_repository=app_context["trust_score_repository"],
                    notification_service=app_context.get("notification_service"),
                    audit_service=app_context.get("audit_service"),
                    cache_service=app_context.get("cache_service")
                )
                
                # Inicializar serviço (pré-carregar configurações, políticas, etc.)
                await app_context["adaptive_scaling_service"].initialize()
                
                # Inicializar e registrar resolvers GraphQL
                if "graphql_server" in app_context:
                    integration = AdaptiveScalingGraphQLIntegration(app_context)
                    await integration.register_resolvers(app_context["graphql_server"])
                    app_context["adaptive_scaling_graphql"] = integration
                else:
                    logger.warning("Servidor GraphQL não encontrado no contexto da aplicação")
                
                logger.info("Módulo de escalonamento adaptativo inicializado com sucesso")
                return app_context
                
            except Exception as e:
                logger.error(f"Erro ao inicializar módulo de escalonamento adaptativo: {e}")
                raise