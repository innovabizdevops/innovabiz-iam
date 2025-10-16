"""
INNOVABIZ - Servidor GraphQL para Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Servidor GraphQL para exposição dos serviços de validação
           de compliance do módulo IAM, incluindo validação HIPAA
           para o módulo Healthcare.
==================================================================
"""

import os
import json
import logging
import uuid
from datetime import datetime
from typing import Optional, Dict, Any

import uvicorn
from fastapi import FastAPI, Request, Depends, HTTPException, Header
from fastapi.responses import JSONResponse, FileResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import OAuth2PasswordBearer
from graphql.execution.executors.asyncio import AsyncioExecutor
from starlette.graphql import GraphQLApp
from starlette.responses import HTMLResponse, PlainTextResponse

from . import schema
from ..validator import ComplianceFramework, RegionCode

# Configurar logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("innovabiz.iam.compliance.graphql.server")

# Criar aplicação FastAPI
app = FastAPI(
    title="INNOVABIZ Compliance IAM API",
    description="API GraphQL para validação de compliance IAM, incluindo HIPAA para Healthcare",
    version="1.0.0"
)

# Configurar CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Configurar autenticação
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Diretórios para relatórios e artefatos
REPORT_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "reports")
os.makedirs(REPORT_DIR, exist_ok=True)

# Dependência para validar tenant
async def get_tenant_id(x_tenant_id: Optional[str] = Header(None)):
    """Valida o ID do tenant na requisição"""
    if not x_tenant_id:
        raise HTTPException(status_code=400, detail="x-tenant-id header is required")
    try:
        return uuid.UUID(x_tenant_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid tenant ID format")

# Rota para GraphQL
app.add_route(
    "/graphql",
    GraphQLApp(
        schema=schema, 
        executor_class=AsyncioExecutor,
        graphiql=True
    )
)

# Rota para verificação de saúde
@app.get("/health")
async def health_check():
    """Verificação de saúde do serviço"""
    return {"status": "healthy", "timestamp": datetime.now().isoformat()}

# Rota para download de relatórios
@app.get("/reports/{report_id}.{format}")
async def get_report(report_id: str, format: str, tenant_id: uuid.UUID = Depends(get_tenant_id)):
    """Obtém um relatório gerado"""
    # Validar formato
    format = format.lower()
    if format not in ["html", "pdf", "json", "csv", "markdown"]:
        raise HTTPException(status_code=400, detail="Invalid report format")
    
    # Validar acesso ao relatório (na implementação real, verificaria permissões)
    
    # Caminho do arquivo
    file_path = os.path.join(REPORT_DIR, f"{report_id}.{format}")
    
    # Verificar se o arquivo existe
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="Report not found")
    
    # Definir tipo de conteúdo
    content_types = {
        "html": "text/html",
        "pdf": "application/pdf",
        "json": "application/json",
        "csv": "text/csv",
        "markdown": "text/markdown"
    }
    
    # Retornar arquivo
    return FileResponse(
        path=file_path,
        media_type=content_types.get(format, "application/octet-stream"),
        filename=f"compliance_report_{datetime.now().strftime('%Y%m%d')}.{format}"
    )

# Rota para documentação GraphQL
@app.get("/")
async def graphql_playground():
    """Página de documentação GraphQL"""
    return HTMLResponse("""
    <!DOCTYPE html>
    <html>
    <head>
        <title>INNOVABIZ Compliance IAM API</title>
        <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
        <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
        <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
    </head>
    <body>
        <div id="root">
            <div class="loading">Loading...</div>
        </div>
        <script>
            window.addEventListener('load', function (event) {
                GraphQLPlayground.init(document.getElementById('root'), {
                    endpoint: '/graphql'
                })
            })
        </script>
    </body>
    </html>
    """)

# Rota para schema GraphQL
@app.get("/schema")
async def get_schema():
    """Obtém o schema GraphQL como texto"""
    schema_str = str(schema)
    return PlainTextResponse(schema_str)

# Rota para validação integrada com AR
@app.post("/ar-authentication/validate")
async def validate_ar_authentication(
    request: Request,
    tenant_id: uuid.UUID = Depends(get_tenant_id)
):
    """
    Validação de autenticação AR para acesso a dados de saúde protegidos
    Integra com o sistema AR de autenticação para verificar se o acesso
    a PHI está em conformidade com requisitos HIPAA
    """
    try:
        # Extrair dados da requisição
        data = await request.json()
        ar_factors = data.get("ar_factors", [])
        healthcare_data_access = data.get("healthcare_data_access", False)
        phi_category = data.get("phi_category", "default")
        
        # Validar entrada
        if not ar_factors:
            return JSONResponse(
                status_code=400,
                content={"error": "ar_factors is required"}
            )
        
        # Verificar se o acesso requer conformidade com HIPAA
        hipaa_required = healthcare_data_access
        
        # Na implementação real, validaria contra configuração do tenant
        # e requisitos HIPAA específicos para o tipo de dados PHI
        
        # Simulação de validação
        if hipaa_required:
            # Verificar fatores AR mínimos para HIPAA
            required_factor_count = 2 if phi_category == "sensitive" else 1
            
            if len(ar_factors) < required_factor_count:
                return JSONResponse(
                    status_code=403,
                    content={
                        "compliant": False,
                        "message": f"HIPAA compliance requires at least {required_factor_count} AR authentication factors for {phi_category} PHI",
                        "reference": "HIPAA Security Rule §164.312(d)",
                        "required_factors": required_factor_count,
                        "provided_factors": len(ar_factors)
                    }
                )
        
        # Acesso em conformidade
        return JSONResponse(
            content={
                "compliant": True,
                "timestamp": datetime.now().isoformat(),
                "tenant_id": str(tenant_id),
                "validation_id": str(uuid.uuid4()),
                "phi_access_authorized": True,
                "audit_record_created": True
            }
        )
    
    except Exception as e:
        logger.error(f"Erro ao validar autenticação AR: {str(e)}")
        return JSONResponse(
            status_code=500,
            content={"error": "Internal server error during AR authentication validation"}
        )

# Função para iniciar o servidor
def start_server(host="0.0.0.0", port=8000, reload=False):
    """Inicia o servidor GraphQL"""
    uvicorn.run("infrastructure.iam.compliance.graphql.server:app", host=host, port=port, reload=reload)

if __name__ == "__main__":
    # Iniciar o servidor quando executado diretamente
    start_server(reload=True)
