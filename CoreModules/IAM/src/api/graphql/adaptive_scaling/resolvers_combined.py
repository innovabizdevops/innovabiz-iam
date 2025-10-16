"""
Resolvers GraphQL combinados para o serviço de escalonamento adaptativo.

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

from .resolvers import AdaptiveScalingResolvers
from .resolvers_mutations import AdaptiveScalingMutations
from .resolvers_mutations2 import AdaptiveScalingPolicyMutations
from ...services.adaptive_scaling.adaptive_scaling_service import AdaptiveScalingService
from ....app.repositories.trust_score_repository import TrustScoreRepository

# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingGraphQLResolvers:
    """
    Implementação completa dos resolvers GraphQL para o serviço de escalonamento adaptativo.
    Combina todas as operações de query e mutation.
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
        
        # Criar instâncias dos resolvers especializados
        self._query_resolvers = AdaptiveScalingResolvers(
            adaptive_scaling_service=adaptive_scaling_service,
            trust_repository=trust_repository
        )
        
        self._mutation_resolvers = AdaptiveScalingMutations(
            adaptive_scaling_service=adaptive_scaling_service,
            db_pool=self.db_pool
        )
        
        self._policy_mutation_resolvers = AdaptiveScalingPolicyMutations(
            adaptive_scaling_service=adaptive_scaling_service,
            db_pool=self.db_pool
        )
    
    def set_notification_service(self, notification_service):
        """
        Define o serviço de notificações.
        
        Args:
            notification_service: Serviço de notificações
        """
        self.notification_service = notification_service
        self._query_resolvers.notification_service = notification_service
        self._mutation_resolvers.notification_service = notification_service
        self._policy_mutation_resolvers.notification_service = notification_service
    
    def set_audit_service(self, audit_service):
        """
        Define o serviço de auditoria.
        
        Args:
            audit_service: Serviço de auditoria
        """
        self.audit_service = audit_service
        self._query_resolvers.audit_service = audit_service
        self._mutation_resolvers.audit_service = audit_service
        self._policy_mutation_resolvers.audit_service = audit_service
    
    #
    # Resolvers de Query
    #
    
    async def get_user_security_profile(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_user_security_profile(*args, **kwargs)
    
    async def get_current_security_level(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_current_security_level(*args, **kwargs)
    
    async def get_scaling_events(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_events(*args, **kwargs)
    
    async def get_scaling_event_by_id(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_event_by_id(*args, **kwargs)
    
    async def get_scaling_triggers(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_triggers(*args, **kwargs)
    
    async def get_scaling_trigger_by_id(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_trigger_by_id(*args, **kwargs)
    
    async def get_scaling_policies(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_policies(*args, **kwargs)
    
    async def get_scaling_policy_by_id(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_scaling_policy_by_id(*args, **kwargs)
    
    async def get_adaptive_scaling_status(self, *args, **kwargs):
        """Proxy para o resolver de query."""
        return await self._query_resolvers.get_adaptive_scaling_status(*args, **kwargs)
    
    #
    # Resolvers de Mutation
    #
    
    async def apply_security_adjustment(self, *args, **kwargs):
        """Proxy para o resolver de mutation."""
        return await self._mutation_resolvers.apply_security_adjustment(*args, **kwargs)
    
    async def upsert_scaling_trigger(self, *args, **kwargs):
        """Proxy para o resolver de mutation."""
        return await self._mutation_resolvers.upsert_scaling_trigger(*args, **kwargs)
    
    async def set_scaling_trigger_enabled(self, *args, **kwargs):
        """Proxy para o resolver de mutation."""
        return await self._mutation_resolvers.set_scaling_trigger_enabled(*args, **kwargs)
    
    async def delete_scaling_trigger(self, *args, **kwargs):
        """Proxy para o resolver de mutation."""
        return await self._mutation_resolvers.delete_scaling_trigger(*args, **kwargs)
    
    async def upsert_scaling_policy(self, *args, **kwargs):
        """Proxy para o resolver de mutation de política."""
        return await self._policy_mutation_resolvers.upsert_scaling_policy(*args, **kwargs)
    
    async def set_scaling_policy_enabled(self, *args, **kwargs):
        """Proxy para o resolver de mutation de política."""
        return await self._policy_mutation_resolvers.set_scaling_policy_enabled(*args, **kwargs)
    
    async def delete_scaling_policy(self, *args, **kwargs):
        """Proxy para o resolver de mutation de política."""
        return await self._policy_mutation_resolvers.delete_scaling_policy(*args, **kwargs)
    
    async def revoke_scaling_event(self, *args, **kwargs):
        """Proxy para o resolver de mutation de política."""
        return await self._policy_mutation_resolvers.revoke_scaling_event(*args, **kwargs)