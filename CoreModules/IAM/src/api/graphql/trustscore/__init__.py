"""
Módulo de inicialização da API GraphQL do TrustScore.

Este módulo configura e registra os resolvers do TrustScore no esquema GraphQL principal.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import logging
from pathlib import Path
from typing import Dict, Any, Optional

from graphql import (
    GraphQLSchema,
    build_schema,
    extend_schema,
    parse
)

from ....app.services.trust_score_query_service import TrustScoreQueryService
from ....app.repositories.trust_score_repository import TrustScoreRepository
from ....app.repositories.trust_score_postgres_repository import TrustScorePostgresRepository
from .resolvers import TrustScoreResolvers

# Configuração de logging
logger = logging.getLogger(__name__)


def get_schema_path(filename: str) -> str:
    """Retorna o caminho absoluto para um arquivo de esquema GraphQL."""
    current_dir = Path(__file__).parent.absolute()
    return os.path.join(current_dir, filename)


def load_schema_from_file(file_path: str) -> str:
    """Carrega um esquema GraphQL de um arquivo."""
    with open(file_path, 'r') as schema_file:
        return schema_file.read()


async def register_trustscore_graphql(
    base_schema: GraphQLSchema,
    db_pool,
    resolvers_map: Dict[str, Any]
) -> GraphQLSchema:
    """
    Registra os tipos, schema e resolvers do TrustScore no esquema GraphQL principal.
    
    Args:
        base_schema: Esquema GraphQL base
        db_pool: Pool de conexões de banco de dados
        resolvers_map: Mapa de resolvers existentes
        
    Returns:
        GraphQLSchema: Esquema GraphQL atualizado
    """
    try:
        logger.info("Registrando API GraphQL do TrustScore")
        
        # Inicializar repositório e serviços
        repository = TrustScorePostgresRepository(db_pool)
        query_service = TrustScoreQueryService(repository, db_pool)
        
        # Inicializar resolvers
        trust_resolvers = TrustScoreResolvers(query_service, repository)
        
        # Carregar arquivos de esquema
        types_sdl = load_schema_from_file(get_schema_path("types.graphql"))
        schema_sdl = load_schema_from_file(get_schema_path("schema.graphql"))
        
        # Estender o esquema base com os tipos e consultas do TrustScore
        extended_schema = extend_schema(
            base_schema,
            parse(types_sdl + "\n" + schema_sdl)
        )
        
        # Registrar os resolvers
        resolvers_map.update({
            "Query": {
                "userTrustProfile": trust_resolvers.get_user_trust_profile,
                "trustScoreTimeline": trust_resolvers.get_trust_score_timeline,
                "regionalComparison": trust_resolvers.get_regional_comparison,
                "anomalyDetails": trust_resolvers.get_anomaly_details,
                "tenantStatistics": trust_resolvers.get_tenant_statistics,
                "trustScoreResults": trust_resolvers.get_trust_score_results
            }
        })
        
        logger.info("API GraphQL do TrustScore registrada com sucesso")
        return extended_schema
    except Exception as e:
        logger.error(f"Erro ao registrar API GraphQL do TrustScore: {e}")
        raise