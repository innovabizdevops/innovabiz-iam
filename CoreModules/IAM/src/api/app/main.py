"""
INNOVABIZ IAM - API Principal
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Aplicação FastAPI principal do módulo IAM com auditoria e observabilidade
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA
"""

import os
import sys
import asyncio
import logging
from typing import Dict, Any, Optional
from fastapi import FastAPI, Request, Response, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

# Importações dos componentes internos
from .core.observability import logger
from .core.app_integrator import setup_application_integration
from .core.events import setup_event_handlers
from .routers import audit
from .services.audit_service import get_audit_service, init_audit_service
from .core.context import get_regional_context, get_tenant_context, RegionalContext

# Configurações da aplicação
APP_NAME = "innovabiz-iam"
APP_VERSION = "1.0.0"
APP_DESCRIPTION = """
INNOVABIZ IAM - Sistema de Identidade e Acesso com Auditoria Multi-Contexto

Compatibilidade:
- Multi-tenant
- Multi-regional (BR, US, EU, AO)
- Multi-idioma
- Multi-compliance (GDPR, LGPD, PCI DSS, PSD2, BACEN, BNA)

© 2025 INNOVABIZ
"""

# Inicializa a aplicação FastAPI
app = FastAPI(
    title=APP_NAME,
    description=APP_DESCRIPTION,
    version=APP_VERSION,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json"
)

# Adiciona atributo de versão para uso nos componentes
app.version = APP_VERSION

# Configura os handlers de eventos de startup e shutdown
setup_event_handlers(app)

# Cria o logger estruturado para a aplicação principal
main_logger = logger.bind(service=APP_NAME, version=APP_VERSION)

# Endpoints de saúde e prontidão

@app.get("/health", tags=["Monitoring"])
async def health_check():
    """
    Endpoint para verificar a saúde do serviço.
    
    Este endpoint não requer autenticação e é usado pelos orquestradores
    para verificar se o serviço está funcionando corretamente.
    """
    return {"status": "healthy", "service": APP_NAME, "version": APP_VERSION}

@app.get("/ready", tags=["Monitoring"])
async def ready_check():
    """
    Endpoint para verificar a prontidão do serviço.
    
    Este endpoint verifica se todas as dependências estão acessíveis
    e se o serviço está pronto para receber tráfego.
    """
    dependencies_status = {}
    all_ready = True
    
    # Verifica conexão com MongoDB
    try:
        audit_service = get_audit_service()
        mongodb_ready = await audit_service.check_database_connection()
        dependencies_status["mongodb"] = "ready" if mongodb_ready else "not_ready"
        if not mongodb_ready:
            all_ready = False
    except Exception as e:
        dependencies_status["mongodb"] = f"error: {str(e)}"
        all_ready = False
    
    # Verifica conexão com Redis (se aplicável)
    try:
        audit_service = get_audit_service()
        redis_ready = await audit_service.check_redis_connection()
        dependencies_status["redis"] = "ready" if redis_ready else "not_ready"
        if not redis_ready:
            all_ready = False
    except Exception as e:
        dependencies_status["redis"] = f"error: {str(e)}"
        # Redis pode ser opcional, então não alteramos o status geral
    
    status_code = status.HTTP_200_OK if all_ready else status.HTTP_503_SERVICE_UNAVAILABLE
    
    return JSONResponse(
        content={
            "status": "ready" if all_ready else "not_ready",
            "service": APP_NAME,
            "version": APP_VERSION,
            "dependencies": dependencies_status
        },
        status_code=status_code
    )

@app.get("/live", tags=["Monitoring"])
async def liveness_check():
    """
    Endpoint para verificar se o serviço está em execução.
    
    Este endpoint é mais simples que o ready check e apenas
    verifica se a aplicação está respondendo.
    """
    return {"status": "live", "service": APP_NAME, "version": APP_VERSION}

@app.get("/info", tags=["Monitoring"])
async def info():
    """
    Endpoint para obter informações sobre o serviço.
    """
    return {
        "service": APP_NAME,
        "version": APP_VERSION,
        "description": "INNOVABIZ IAM - Sistema de Identidade e Acesso",
        "environment": os.environ.get("ENVIRONMENT", "development"),
        "regional_contexts": [context.value for context in RegionalContext],
    }

@app.get("/context", tags=["Monitoring"])
async def get_context(
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context)
):
    """
    Endpoint para verificar o contexto atual.
    
    Este endpoint é útil para debugging e verificação de como
    o sistema está detectando o contexto regional e tenant.
    """
    return {
        "regional_context": regional_context.value,
        "tenant_id": tenant_id
    }

# Função de inicialização assíncrona
async def initialize() -> None:
    """
    Inicializa os componentes assíncronos da aplicação.
    """
    try:
        # Inicializa o serviço de auditoria
        await init_audit_service()
        
        # Configura a integração da aplicação
        await setup_application_integration(
            app=app,
            audit_service=get_audit_service(),
            enable_audit=True,
            enable_observability=True
        )
        
        main_logger.info("Aplicação IAM inicializada com sucesso")
    except Exception as e:
        main_logger.error(f"Falha ao inicializar a aplicação: {str(e)}")
        raise e

# Handler para inicialização assíncrona
@app.on_event("startup")
async def startup_event() -> None:
    """
    Evento executado na inicialização da aplicação.
    """
    await initialize()

# Handler para finalização da aplicação
@app.on_event("shutdown")
async def shutdown_event() -> None:
    """
    Evento executado no encerramento da aplicação.
    """
    main_logger.info("Finalizando aplicação IAM")
    
    # Fecha conexões e limpa recursos
    try:
        audit_service = get_audit_service()
        if audit_service:
            await audit_service.close()
    except Exception as e:
        main_logger.error(f"Erro ao finalizar serviço de auditoria: {str(e)}")

# Handler de exceções global
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    """
    Handler global de exceções.
    
    Garante que todas as exceções não tratadas sejam registradas
    e retornem uma resposta adequada.
    """
    # Obtém caminho e método para logging
    path = request.url.path
    method = request.method
    
    # Gera ID único para o erro para facilitar troubleshooting
    import uuid
    error_id = str(uuid.uuid4())
    
    # Determina o código de status HTTP apropriado
    if isinstance(exc, HTTPException):
        status_code = exc.status_code
    else:
        status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
    
    # Log detalhado do erro
    main_logger.error(
        f"Erro não tratado: {str(exc)}",
        error_id=error_id,
        path=path,
        method=method,
        status_code=status_code,
        error_type=type(exc).__name__,
        error_details=str(exc)
    )
    
    # Retorna resposta com detalhes úteis para o cliente
    return JSONResponse(
        status_code=status_code,
        content={
            "error": "Ocorreu um erro ao processar a requisição",
            "error_id": error_id,
            "details": str(exc) if isinstance(exc, HTTPException) else "Erro interno do servidor",
            "status_code": status_code
        }
    )

# Inclui os routers da aplicação
app.include_router(audit.router, prefix="/api/v1/audit", tags=["Audit"])

# Adiciona middleware de CORS (já configurado pelo AppIntegrator, mas deixamos explícito aqui)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configurar adequadamente para produção
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Ponto de entrada para execução direta
if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8000"))
    host = os.environ.get("HOST", "0.0.0.0")
    
    main_logger.info(f"Iniciando servidor Uvicorn em {host}:{port}")
    
    # Configuração do Uvicorn
    uvicorn.run(
        "app.main:app",
        host=host,
        port=port,
        reload=os.environ.get("ENVIRONMENT", "development") == "development",
        workers=int(os.environ.get("WORKERS", "1")),
        access_log=True
    )# Importação do router de auditoria
from .routers import audit_router

# Registro do router de auditoria na aplicação
app.include_router(audit_router.router)

# Inicialização de jobs agendados para o sistema de auditoria
@app.on_event("startup")
async def setup_audit_scheduled_jobs():
    """
    Configura jobs agendados para o sistema de auditoria multi-contexto.
    
    Inclui jobs para aplicação de políticas de retenção e geração de estatísticas
    para todos os tenants e contextos regionais suportados.
    """
    logger.info("Configurando jobs agendados para o sistema de auditoria multi-contexto")
    
    # Scheduler já configurado na inicialização da aplicação
    
    # Job para aplicação de políticas de retenção (execução diária às 01:00)
    @scheduler.scheduled_job("cron", id="apply_retention_policies", hour=1, minute=0)
    async def apply_retention_policies_job():
        try:
            logger.info("Iniciando job de aplicação de políticas de retenção")
            
            # Obtém uma nova sessão do banco de dados
            async with get_async_session() as session:
                audit_service = AuditService(session)
                
                # Lista de contextos regionais suportados
                regional_contexts = ["BR", "US", "EU", "AO"]
                
                # Lista de tenants ativos
                # Em um ambiente real, essa lista seria obtida de um serviço de tenants
                active_tenants = await get_active_tenants(session)
                
                total_processed = 0
                
                # Processa cada combinação de tenant e contexto regional
                for tenant_id in active_tenants:
                    for regional_context in regional_contexts:
                        try:
                            logger.info(
                                "Aplicando políticas de retenção", 
                                tenant_id=tenant_id, 
                                regional_context=regional_context
                            )
                            
                            # Aplica políticas de retenção para o tenant e contexto regional
                            result = await audit_service.apply_retention_policies(
                                tenant_id=tenant_id,
                                regional_context=regional_context,
                                batch_size=500,
                                dry_run=False
                            )
                            
                            # Contabiliza eventos processados
                            anonymized = result.get("statistics", {}).get("anonymized", 0)
                            deleted = result.get("statistics", {}).get("deleted", 0)
                            total_processed += anonymized + deleted
                            
                        except Exception as e:
                            logger.error(
                                "Falha ao aplicar políticas para tenant/região",
                                tenant_id=tenant_id,
                                regional_context=regional_context,
                                error=str(e)
                            )
                
                logger.info(
                    "Job de aplicação de políticas de retenção concluído",
                    total_processed=total_processed
                )
                
        except Exception as e:
            logger.error(
                "Falha no job de aplicação de políticas de retenção",
                error=str(e)
            )
    
    # Job para geração de estatísticas de auditoria (execução diária às 02:00)
    @scheduler.scheduled_job("cron", id="generate_audit_statistics", hour=2, minute=0)
    async def generate_audit_statistics_job():
        try:
            logger.info("Iniciando job de geração de estatísticas de auditoria")
            
            # Obtém uma nova sessão do banco de dados
            async with get_async_session() as session:
                audit_service = AuditService(session)
                
                # Lista de contextos regionais suportados
                regional_contexts = ["BR", "US", "EU", "AO"]
                
                # Períodos para geração de estatísticas
                periods = ["daily", "weekly", "monthly"]
                
                # Lista de tenants ativos
                active_tenants = await get_active_tenants(session)
                
                total_generated = 0
                
                # Processa cada combinação de tenant e contexto regional
                for tenant_id in active_tenants:
                    for regional_context in regional_contexts:
                        for period in periods:
                            try:
                                logger.info(
                                    "Gerando estatísticas de auditoria", 
                                    tenant_id=tenant_id, 
                                    regional_context=regional_context,
                                    period=period
                                )
                                
                                # Gera estatísticas para o tenant, contexto regional e período
                                result = await audit_service.generate_audit_statistics(
                                    tenant_id=tenant_id,
                                    regional_context=regional_context,
                                    period=period,
                                    update_existing=True
                                )
                                
                                # Contabiliza estatísticas geradas
                                stat_count = len(result.get("statistics", {}))
                                total_generated += stat_count
                                
                            except Exception as e:
                                logger.error(
                                    "Falha ao gerar estatísticas para tenant/região/período",
                                    tenant_id=tenant_id,
                                    regional_context=regional_context,
                                    period=period,
                                    error=str(e)
                                )
                
                logger.info(
                    "Job de geração de estatísticas de auditoria concluído",
                    total_generated=total_generated
                )
                
        except Exception as e:
            logger.error(
                "Falha no job de geração de estatísticas de auditoria",
                error=str(e)
            )
    
    # Job para geração de relatórios de compliance automáticos (execução semanal aos domingos às 03:00)
    @scheduler.scheduled_job("cron", id="generate_compliance_reports", day_of_week=6, hour=3, minute=0)
    async def generate_compliance_reports_job():
        try:
            logger.info("Iniciando job de geração de relatórios de compliance automáticos")
            
            # Obtém uma nova sessão do banco de dados
            async with get_async_session() as session:
                audit_service = AuditService(session)
                
                # Mapeamento de regiões para frameworks de compliance obrigatórios
                region_compliance = {
                    "BR": [ComplianceFramework.LGPD, ComplianceFramework.BACEN],
                    "US": [ComplianceFramework.SOX, ComplianceFramework.PCI_DSS],
                    "EU": [ComplianceFramework.GDPR, ComplianceFramework.PSD2],
                    "AO": [ComplianceFramework.BNA]
                }
                
                # Data de início e fim para relatórios semanais
                end_date = datetime.now()
                start_date = end_date - timedelta(days=7)
                
                # Lista de tenants ativos
                active_tenants = await get_active_tenants(session)
                
                total_generated = 0
                
                # Processa cada combinação de tenant, contexto regional e framework
                for tenant_id in active_tenants:
                    for region, frameworks in region_compliance.items():
                        for framework in frameworks:
                            try:
                                logger.info(
                                    "Gerando relatório de compliance automático", 
                                    tenant_id=tenant_id, 
                                    regional_context=region,
                                    framework=framework.value
                                )
                                
                                # Gera relatório de compliance
                                await audit_service.generate_compliance_report(
                                    tenant_id=tenant_id,
                                    regional_context=region,
                                    compliance_framework=framework,
                                    start_date=start_date,
                                    end_date=end_date,
                                    report_type="standard",
                                    report_format="json",
                                    user_id=None,  # Relatório gerado pelo sistema
                                    include_anonymized=False
                                )
                                
                                total_generated += 1
                                
                            except Exception as e:
                                logger.error(
                                    "Falha ao gerar relatório de compliance",
                                    tenant_id=tenant_id,
                                    regional_context=region,
                                    framework=framework.value,
                                    error=str(e)
                                )
                
                logger.info(
                    "Job de geração de relatórios de compliance concluído",
                    total_generated=total_generated
                )
                
        except Exception as e:
            logger.error(
                "Falha no job de geração de relatórios de compliance",
                error=str(e)
            )

# Função auxiliar para obter tenants ativos
async def get_active_tenants(session) -> List[str]:
    """
    Obtém a lista de tenants ativos no sistema.
    
    Em um ambiente real, isso seria obtido de um serviço de gerenciamento de tenants.
    Para fins de demonstração, usamos uma lista fixa de tenants.
    """
    # Simulação de tenants ativos
    # Em um ambiente de produção, isso seria obtido de uma tabela ou serviço de tenants
    return ["tenant1", "tenant2", "tenant3", "tenant4", "innovabiz"]