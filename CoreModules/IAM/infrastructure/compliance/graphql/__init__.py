"""
INNOVABIZ - Módulo GraphQL para Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Módulo de integração GraphQL para o serviço de validação
           de compliance IAM, incluindo validação HIPAA para o módulo
           Healthcare.
==================================================================
"""

import graphene
from graphql import GraphQLError

from .resolvers_query import Query
from .resolvers_mutation import Mutation

# Definir o schema GraphQL
schema = graphene.Schema(query=Query, mutation=Mutation)

__all__ = ['schema']
