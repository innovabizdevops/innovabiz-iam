"""
INNOVABIZ - Servidor GraphQL para Validação IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação do servidor GraphQL para exposição das
           funcionalidades de validação, certificação e conformidade
           do módulo IAM através de API GraphQL.
==================================================================
"""

import os
import json
import logging
import uvicorn
from fastapi import FastAPI, Request, Response, Depends, HTTPException, Security
from fastapi.middleware.cors import CORSMiddleware
from starlette.graphql import GraphQLApp
from starlette.responses import JSONResponse
from typing import Dict, List, Any, Optional, Union
from graphql import GraphQLError

from .schema import schema
from ...auth.middleware import AuthMiddleware  # Middleware de autenticação IAM

# Configuração de logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("innovabiz.iam.validation.graphql")

# App FastAPI com GraphQL
app = FastAPI(
    title="INNOVABIZ IAM Validation API",
    description="API para validação, certificação e conformidade do módulo IAM",
    version="1.0.0"
)

# Configuração CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Em ambiente de produção, restringir para origens confiáveis
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Middleware de autenticação
auth_middleware = AuthMiddleware()


# Endpoint GraphQL
@app.route("/graphql", methods=["GET", "POST"])
async def graphql_endpoint(request: Request):
    """Endpoint GraphQL para validação IAM"""
    # Autenticação e autorização
    try:
        # Verificar token de autenticação
        await auth_middleware.authenticate(request)
        
        # Verificar autorização (assumindo que o middleware adicionou os dados do usuário)
        user_data = request.state.user
        if not user_data or not auth_middleware.has_permission(user_data, "iam:validation:read"):
            raise HTTPException(status_code=403, detail="Não autorizado para acessar validação IAM")
            
        # Para todas as mutações, requer permissão de escrita
        if request.method == "POST":
            body = await request.json()
            operation = body.get("operationName")
            
            # Verificar se é uma mutação
            if operation and any(op in operation for op in ["RunValidation", "GenerateCertificate", "ExportReport"]):
                if not auth_middleware.has_permission(user_data, "iam:validation:write"):
                    raise HTTPException(status_code=403, detail="Não autorizado para executar validação IAM")
        
        # Processar a solicitação GraphQL
        return await GraphQLApp(schema=schema)(request)
    
    except HTTPException as e:
        # Retornar erro HTTP
        return JSONResponse(
            status_code=e.status_code,
            content={"errors": [{"message": e.detail}]}
        )
    except GraphQLError as e:
        # Retornar erro GraphQL
        return JSONResponse(
            status_code=400,
            content={"errors": [{"message": str(e)}]}
        )
    except Exception as e:
        # Registrar erro não tratado
        logger.error(f"Erro não tratado: {str(e)}", exc_info=True)
        return JSONResponse(
            status_code=500,
            content={"errors": [{"message": "Erro interno do servidor"}]}
        )


# Endpoint de saúde
@app.get("/health", tags=["health"])
async def health_check():
    """Endpoint de verificação de saúde"""
    return {"status": "healthy", "service": "iam-validation-api"}


# Endpoint de versão
@app.get("/version", tags=["version"])
async def version():
    """Retorna informações de versão"""
    return {
        "service": "iam-validation-api",
        "version": "1.0.0",
        "framework": "GraphQL + FastAPI"
    }


def start_server(host: str = "0.0.0.0", port: int = 8000, reload: bool = False):
    """
    Inicia o servidor GraphQL
    
    Args:
        host: Host para o servidor
        port: Porta para o servidor
        reload: Habilitar reload automático para desenvolvimento
    """
    logger.info(f"Iniciando servidor GraphQL de validação IAM em {host}:{port}")
    
    # Configurações do Uvicorn
    uvicorn_config = {
        "app": "iam.validation.graphql.server:app",
        "host": host,
        "port": port,
        "reload": reload,
        "log_level": "info"
    }
    
    # Iniciar servidor
    uvicorn.run(**uvicorn_config)


if __name__ == "__main__":
    # Iniciar servidor diretamente, útil para desenvolvimento
    start_server(reload=True)
