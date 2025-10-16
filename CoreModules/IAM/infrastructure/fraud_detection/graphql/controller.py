#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Controlador GraphQL para análises comportamentais

Este módulo implementa o controlador da API GraphQL para análises comportamentais,
conectando o schema com os resolvers e fornecendo endpoints para consultas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
from typing import Dict, Any, Optional

from fastapi import FastAPI, Request, Response, Depends, HTTPException
from fastapi.responses import JSONResponse
from fastapi.security import OAuth2PasswordBearer
from graphql import GraphQLError
from starlette.middleware.cors import CORSMiddleware

from .schema import schema
from .resolvers import bind_resolvers_to_schema

# Configuração do logger
logger = logging.getLogger("iam.trustguard.graphql.controller")

# Aplicar resolvers ao schema
bind_resolvers_to_schema(schema)

# Esquema de autenticação
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")


# Middleware de autenticação
async def verify_token(token: str = Depends(oauth2_scheme)) -> Dict[str, Any]:
    """
    Verifica o token de acesso e retorna informações do usuário.
    
    Args:
        token: Token de acesso
        
    Returns:
        Informações do usuário autenticado
        
    Raises:
        HTTPException: Se o token for inválido
    """
    try:
        # Em produção, validar o token com o IAM
        # Este é um exemplo simplificado
        if not token or len(token) < 8:
            logger.warning(f"Token inválido: {token}")
            raise HTTPException(
                status_code=401,
                detail="Token de acesso inválido",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        # Informações do usuário (simuladas)
        return {
            "user_id": "user123",
            "username": "usuario.exemplo",
            "roles": ["analyst", "viewer"],
            "permissions": ["behavior:read", "alerts:read"]
        }
    except Exception as e:
        logger.error(f"Erro ao verificar token: {str(e)}")
        raise HTTPException(
            status_code=401,
            detail="Falha na autenticação",
            headers={"WWW-Authenticate": "Bearer"},
        )


# Middleware de autorização
def verify_permission(required_permission: str, user_info: Dict[str, Any]) -> bool:
    """
    Verifica se o usuário tem a permissão necessária.
    
    Args:
        required_permission: Permissão necessária
        user_info: Informações do usuário
        
    Returns:
        True se o usuário tem permissão, False caso contrário
    """
    # Verificar permissões do usuário
    permissions = user_info.get("permissions", [])
    
    # Verificar permissão específica
    if required_permission in permissions:
        return True
    
    # Verificar permissão wildcard (ex: behavior:*)
    permission_prefix = required_permission.split(":")[0] + ":*"
    if permission_prefix in permissions:
        return True
    
    # Verificar permissão total
    if "*:*" in permissions:
        return True
    
    return False


# Criar aplicação FastAPI
app = FastAPI(
    title="API GraphQL para Análise Comportamental",
    description="API GraphQL para consulta de eventos, alertas e análises comportamentais",
    version="1.0.0"
)

# Configurar CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Em produção, restringir para origens específicas
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.post("/graphql")
async def graphql_endpoint(request: Request, user_info: Dict[str, Any] = Depends(verify_token)) -> JSONResponse:
    """
    Endpoint GraphQL para consultas de análise comportamental.
    
    Args:
        request: Requisição HTTP
        user_info: Informações do usuário autenticado
        
    Returns:
        Resposta JSON com resultado da consulta GraphQL
    """
    try:
        # Ler corpo da requisição
        content_type = request.headers.get("Content-Type", "")
        if "application/json" in content_type:
            data = await request.json()
        else:
            data = await request.form()
        
        # Extrair query e variáveis
        query = data.get("query")
        variables = data.get("variables")
        operation_name = data.get("operationName")
        
        if not query:
            logger.warning("Requisição GraphQL sem query")
            return JSONResponse(
                content={"errors": [{"message": "Query não fornecida"}]},
                status_code=400
            )
        
        # Verificar se é uma query de introspeção
        is_introspection = "IntrospectionQuery" in query or "__schema" in query
        
        # Verificar permissão para consultas que não são de introspeção
        if not is_introspection and not verify_permission("behavior:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('user_id')} sem permissão para "
                f"consulta comportamental: {query[:100]}..."
            )
            return JSONResponse(
                content={"errors": [{"message": "Sem permissão para esta consulta"}]},
                status_code=403
            )
        
        # Executar consulta GraphQL
        context = {"user_info": user_info, "request": request}
        
        result = schema.execute(
            query,
            variable_values=variables,
            context_value=context,
            operation_name=operation_name
        )
        
        # Processar erros
        if result.errors:
            errors = [
                {"message": str(error), "path": error.path}
                if hasattr(error, "path") else {"message": str(error)}
                for error in result.errors
            ]
            
            logger.error(f"Erros na consulta GraphQL: {errors}")
            
            return JSONResponse(
                content={"errors": errors, "data": result.data},
                status_code=200  # GraphQL retorna 200 mesmo com erros
            )
        
        # Retornar resultado com dados
        return JSONResponse(content={"data": result.data})
        
    except Exception as e:
        logger.error(f"Erro ao processar consulta GraphQL: {str(e)}")
        return JSONResponse(
            content={"errors": [{"message": f"Erro interno: {str(e)}"}]},
            status_code=500
        )


@app.get("/graphql/schema")
async def get_schema(user_info: Dict[str, Any] = Depends(verify_token)) -> JSONResponse:
    """
    Endpoint para obter o schema GraphQL.
    
    Args:
        user_info: Informações do usuário autenticado
        
    Returns:
        Schema GraphQL em formato JSON
    """
    try:
        # Verificar permissão
        if not verify_permission("behavior:read", user_info):
            logger.warning(
                f"Usuário {user_info.get('user_id')} sem permissão para "
                f"visualizar schema GraphQL"
            )
            return JSONResponse(
                content={"error": "Sem permissão para visualizar schema GraphQL"},
                status_code=403
            )
        
        # Consulta de introspeção para obter o schema
        introspection_query = """
        query IntrospectionQuery {
          __schema {
            queryType {
              name
            }
            types {
              name
              kind
              description
              fields {
                name
                description
                args {
                  name
                  description
                  type {
                    name
                    kind
                    ofType {
                      name
                      kind
                    }
                  }
                  defaultValue
                }
                type {
                  name
                  kind
                  ofType {
                    name
                    kind
                    ofType {
                      name
                      kind
                    }
                  }
                }
              }
              interfaces {
                name
              }
              enumValues {
                name
                description
              }
              possibleTypes {
                name
              }
            }
          }
        }
        """
        
        # Executar consulta de introspeção
        result = schema.execute(introspection_query)
        
        if result.errors:
            logger.error(f"Erro ao obter schema GraphQL: {result.errors}")
            return JSONResponse(
                content={"error": "Erro ao obter schema GraphQL"},
                status_code=500
            )
        
        return JSONResponse(content=result.data)
        
    except Exception as e:
        logger.error(f"Erro ao obter schema GraphQL: {str(e)}")
        return JSONResponse(
            content={"error": f"Erro ao obter schema GraphQL: {str(e)}"},
            status_code=500
        )


def get_graphql_router():
    """
    Obtém o router FastAPI para a API GraphQL.
    
    Returns:
        Router FastAPI
    """
    return app