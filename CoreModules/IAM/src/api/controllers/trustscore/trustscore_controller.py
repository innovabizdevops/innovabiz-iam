"""
Controlador para a API GraphQL do TrustScore.

Este módulo implementa o controlador FastAPI que expõe
os endpoints GraphQL para consulta de pontuações de confiança.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import asyncpg
from typing import Dict, Any, Optional, List
from fastapi import APIRouter, Depends, Request, Response, HTTPException, status
from graphql import GraphQLSchema, graphql_sync, execute, parse

from ....app.repositories.trust_score_postgres_repository import TrustScorePostgresRepository
from ....app.services.trust_score_query_service import TrustScoreQueryService
from ...graphql.trustscore import register_trustscore_graphql
from ...middleware.auth import get_current_user, verify_permissions
from ...config import get_db_pool

# Configuração de logging
logger = logging.getLogger(__name__)

# Criação do router
router = APIRouter(
    prefix="/api/v1/graphql/trustscore",
    tags=["trustscore", "graphql"]
)


class TrustScoreGraphQLController:
    """Controlador para a API GraphQL do TrustScore."""
    
    def __init__(self):
        """Inicializa o controlador."""
        self._schema = None
        self._resolvers_map = {}
        self._initialized = False
    
    async def initialize(self, base_schema: GraphQLSchema) -> None:
        """
        Inicializa o controlador com o esquema GraphQL base.
        
        Args:
            base_schema: Esquema GraphQL base
        """
        try:
            if self._initialized:
                return
            
            # Obter pool de conexão ao banco de dados
            db_pool = await get_db_pool()
            
            # Registrar tipos e resolvers do TrustScore no esquema
            self._schema = await register_trustscore_graphql(
                base_schema=base_schema,
                db_pool=db_pool,
                resolvers_map=self._resolvers_map
            )
            
            self._initialized = True
            logger.info("Controlador GraphQL do TrustScore inicializado com sucesso")
        except Exception as e:
            logger.error(f"Erro ao inicializar controlador GraphQL do TrustScore: {e}")
            raise
    
    @property
    def schema(self) -> GraphQLSchema:
        """Retorna o esquema GraphQL configurado."""
        if not self._initialized:
            raise RuntimeError("Controlador GraphQL do TrustScore não inicializado")
        return self._schema
    
    @property
    def resolvers_map(self) -> Dict[str, Any]:
        """Retorna o mapa de resolvers configurado."""
        if not self._initialized:
            raise RuntimeError("Controlador GraphQL do TrustScore não inicializado")
        return self._resolvers_map


# Instância global do controlador
controller = TrustScoreGraphQLController()


async def execute_graphql_query(
    query: str,
    variables: Dict[str, Any] = None,
    operation_name: Optional[str] = None,
    context: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Executa uma consulta GraphQL.
    
    Args:
        query: Consulta GraphQL
        variables: Variáveis da consulta
        operation_name: Nome da operação
        context: Contexto de execução
        
    Returns:
        Dict: Resultado da execução da consulta
    """
    if not controller._initialized:
        raise RuntimeError("Controlador GraphQL do TrustScore não inicializado")
    
    try:
        # Executar consulta GraphQL de forma assíncrona
        result = await execute(
            schema=controller.schema,
            document=parse(query),
            variable_values=variables,
            operation_name=operation_name,
            context_value=context,
            root_value=None
        )
        
        # Verificar erros
        if result.errors:
            error_messages = [str(error) for error in result.errors]
            logger.error(f"Erros na execução de consulta GraphQL: {error_messages}")
            
            # Estruturar resposta com erros
            return {
                "data": result.data,
                "errors": [
                    {
                        "message": str(error),
                        "locations": [
                            {
                                "line": location.line,
                                "column": location.column
                            } 
                            for location in error.locations
                        ] if error.locations else None,
                        "path": error.path
                    }
                    for error in result.errors
                ]
            }
        
        # Retornar dados sem erros
        return {
            "data": result.data
        }
    except Exception as e:
        logger.error(f"Erro ao executar consulta GraphQL: {e}")
        return {
            "errors": [{"message": str(e)}]
        }


@router.post("")
async def graphql_endpoint(
    request: Request,
    response: Response,
    user = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Endpoint GraphQL principal para consultas do TrustScore.
    
    Args:
        request: Objeto de requisição FastAPI
        response: Objeto de resposta FastAPI
        user: Usuário autenticado (via dependência)
        
    Returns:
        Dict: Resultado da execução da consulta GraphQL
    """
    try:
        # Verificar autenticação
        if not user:
            response.status_code = status.HTTP_401_UNAUTHORIZED
            return {"errors": [{"message": "Não autenticado"}]}
        
        # Extrair dados da requisição
        request_data = await request.json()
        query = request_data.get("query")
        variables = request_data.get("variables", {})
        operation_name = request_data.get("operationName")
        
        if not query:
            response.status_code = status.HTTP_400_BAD_REQUEST
            return {"errors": [{"message": "Consulta GraphQL não fornecida"}]}
        
        # Preparar contexto de execução
        context = {
            "request": request,
            "user": user
        }
        
        # Executar consulta GraphQL
        result = await execute_graphql_query(
            query=query,
            variables=variables,
            operation_name=operation_name,
            context=context
        )
        
        # Verificar erros na resposta
        if "errors" in result:
            response.status_code = status.HTTP_400_BAD_REQUEST
        
        return result
    except Exception as e:
        logger.error(f"Erro no endpoint GraphQL: {e}")
        response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
        return {"errors": [{"message": str(e)}]}


@router.get("/schema")
async def get_schema(
    request: Request,
    response: Response,
    user = Depends(get_current_user),
    _verify = Depends(verify_permissions(["admin:schema:view"]))
) -> Dict[str, Any]:
    """
    Endpoint para obter o esquema GraphQL.
    
    Args:
        request: Objeto de requisição FastAPI
        response: Objeto de resposta FastAPI
        user: Usuário autenticado (via dependência)
        _verify: Verificação de permissões (via dependência)
        
    Returns:
        Dict: Esquema GraphQL em formato SDL
    """
    try:
        # Verificar inicialização
        if not controller._initialized:
            response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
            return {"errors": [{"message": "Serviço GraphQL não inicializado"}]}
        
        # Retornar esquema em formato SDL
        from graphql import print_schema
        schema_sdl = print_schema(controller.schema)
        
        return {
            "data": {
                "schema": schema_sdl
            }
        }
    except Exception as e:
        logger.error(f"Erro ao obter esquema GraphQL: {e}")
        response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
        return {"errors": [{"message": str(e)}]}