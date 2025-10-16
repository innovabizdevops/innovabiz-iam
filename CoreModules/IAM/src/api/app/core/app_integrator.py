"""
INNOVABIZ IAM - Application Integrator
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Módulo de integração para a aplicação IAM
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA

Este módulo implementa a integração dos componentes da aplicação IAM,
conectando o sistema de auditoria, middleware, observabilidade e contexto.
"""

import logging
import asyncio
from typing import Callable, Dict, Any, List, Optional
from fastapi import FastAPI, Request, Response, Depends
from fastapi.middleware.cors import CORSMiddleware

from .observability import setup_observability, logger, ObservabilityContext
from .context import context_manager, get_regional_context, get_tenant_context
from .audit_context_integrator import init_audit_context_integrator, get_audit_context_integrator
from ..middleware.audit_middleware import AuditMiddleware
from ..services.audit_service import AuditService, get_audit_service
from ..models.audit import AuditEventCreate, AuditEventSeverity, AuditEventCategory, RegionalContext


class AppIntegrator:
    """
    Integrador de componentes da aplicação IAM.
    
    Responsabilidades:
    1. Inicializar e conectar os componentes da aplicação
    2. Configurar middleware de auditoria
    3. Configurar observabilidade
    4. Integrar sistemas de contexto e auditoria
    """
    
    def __init__(
        self,
        app: FastAPI,
        audit_service: Optional[AuditService] = None,
        enable_audit: bool = True,
        enable_observability: bool = True
    ):
        """
        Inicializa o integrador da aplicação.
        
        Args:
            app: Aplicação FastAPI
            audit_service: Serviço de auditoria (opcional, será criado se não fornecido)
            enable_audit: Se True, habilita o sistema de auditoria
            enable_observability: Se True, habilita o sistema de observabilidade
        """
        self.app = app
        self.logger = logger.bind(module="app_integrator")
        self.enable_audit = enable_audit
        self.enable_observability = enable_observability
        
        # Inicializa o serviço de auditoria se não fornecido
        self.audit_service = audit_service or get_audit_service()
        
        # Observability context global
        self.observability_context = ObservabilityContext(
            service_name="innovabiz-iam",
            component="api"
        )
    
    async def setup(self):
        """
        Configura e integra todos os componentes da aplicação.
        """
        # Inicializa observabilidade
        if self.enable_observability:
            await self.setup_observability()
        
        # Inicializa auditoria
        if self.enable_audit:
            await self.setup_audit()
        
        # Adiciona middleware CORS
        self.setup_cors()
        
        # Adiciona handlers de eventos da aplicação
        self.setup_event_handlers()
        
        self.logger.info("Integração da aplicação concluída com sucesso")
    
    async def setup_observability(self):
        """
        Configura o sistema de observabilidade.
        """
        # Setup inicial de observabilidade (logging, tracing, metrics)
        setup_observability(self.app)
        
        self.logger.info("Sistema de observabilidade configurado")
    
    async def setup_audit(self):
        """
        Configura o sistema de auditoria e middleware.
        """
        # Inicializa o integrador de auditoria e contexto
        init_audit_context_integrator(self.audit_service)
        
        # Adiciona middleware de auditoria
        self.app.add_middleware(
            AuditMiddleware,
            audit_service=self.audit_service,
            skip_paths=["/health", "/metrics", "/docs", "/redoc", "/openapi.json"],
            enable_opentelemetry=self.enable_observability
        )
        
        self.logger.info("Sistema de auditoria e middleware configurados")
    
    def setup_cors(self):
        """
        Configura o middleware CORS.
        """
        self.app.add_middleware(
            CORSMiddleware,
            allow_origins=["*"],  # Configurar adequadamente para produção
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )
        
        self.logger.info("Middleware CORS configurado")
    
    def setup_event_handlers(self):
        """
        Configura os handlers de eventos da aplicação.
        """
        # Evento de startup
        @self.app.on_event("startup")
        async def startup_event():
            self.logger.info("Aplicação IAM iniciando", **self.observability_context.to_dict())
            
            # Registra evento de auditoria para inicialização
            if self.enable_audit:
                await self.register_startup_audit_event()
        
        # Evento de shutdown
        @self.app.on_event("shutdown")
        async def shutdown_event():
            self.logger.info("Aplicação IAM finalizando", **self.observability_context.to_dict())
            
            # Registra evento de auditoria para finalização
            if self.enable_audit:
                await self.register_shutdown_audit_event()
    
    async def register_startup_audit_event(self):
        """
        Registra um evento de auditoria para a inicialização da aplicação.
        """
        try:
            startup_event = AuditEventCreate(
                action="application_startup",
                category=AuditEventCategory.SYSTEM,
                severity=AuditEventSeverity.INFO,
                details={
                    "service": "innovabiz-iam",
                    "component": "api",
                    "version": self.app.version if hasattr(self.app, "version") else "unknown"
                },
                user_id="system",  # Eventos de sistema usam um usuário especial
                resource_id="innovabiz-iam-api",
                resource_type="application",
                tenant_id="system",  # Eventos de sistema usam um tenant especial
                regional_context=RegionalContext.BR,  # Contexto padrão, poderia ser obtido da configuração
                metadata={
                    "automated": True,
                    "lifecycle_event": "startup"
                }
            )
            
            await self.audit_service.create_audit_event(startup_event)
            self.logger.info("Evento de auditoria de inicialização registrado")
            
        except Exception as e:
            self.logger.error(
                "Falha ao registrar evento de auditoria de inicialização", 
                error=str(e),
                **self.observability_context.to_dict()
            )
    
    async def register_shutdown_audit_event(self):
        """
        Registra um evento de auditoria para a finalização da aplicação.
        """
        try:
            shutdown_event = AuditEventCreate(
                action="application_shutdown",
                category=AuditEventCategory.SYSTEM,
                severity=AuditEventSeverity.INFO,
                details={
                    "service": "innovabiz-iam",
                    "component": "api",
                    "version": self.app.version if hasattr(self.app, "version") else "unknown"
                },
                user_id="system",  # Eventos de sistema usam um usuário especial
                resource_id="innovabiz-iam-api",
                resource_type="application",
                tenant_id="system",  # Eventos de sistema usam um tenant especial
                regional_context=RegionalContext.BR,  # Contexto padrão, poderia ser obtido da configuração
                metadata={
                    "automated": True,
                    "lifecycle_event": "shutdown"
                }
            )
            
            # Criação síncrona do evento para garantir que seja registrado antes do shutdown
            # Usa um timeout curto para não atrasar o shutdown
            try:
                await asyncio.wait_for(
                    self.audit_service.create_audit_event(shutdown_event),
                    timeout=2.0
                )
                self.logger.info("Evento de auditoria de finalização registrado")
            except asyncio.TimeoutError:
                self.logger.warning(
                    "Timeout ao registrar evento de auditoria de finalização",
                    **self.observability_context.to_dict()
                )
            
        except Exception as e:
            self.logger.error(
                "Falha ao registrar evento de auditoria de finalização", 
                error=str(e),
                **self.observability_context.to_dict()
            )


# Factory para criar e configurar o integrador
async def setup_application_integration(
    app: FastAPI,
    audit_service: Optional[AuditService] = None,
    enable_audit: bool = True,
    enable_observability: bool = True
) -> AppIntegrator:
    """
    Configura a integração da aplicação.
    
    Args:
        app: Aplicação FastAPI
        audit_service: Serviço de auditoria (opcional)
        enable_audit: Se True, habilita o sistema de auditoria
        enable_observability: Se True, habilita o sistema de observabilidade
        
    Returns:
        Instância do integrador configurado
    """
    integrator = AppIntegrator(
        app=app,
        audit_service=audit_service,
        enable_audit=enable_audit,
        enable_observability=enable_observability
    )
    
    # Executa a configuração
    await integrator.setup()
    
    return integrator