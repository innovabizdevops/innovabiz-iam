"""
INNOVABIZ IAM - Eventos do Ciclo de Vida da Aplicação
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Eventos de startup e shutdown para inicializar recursos e garantir limpeza adequada
"""

from typing import Callable
import asyncio
import structlog
from fastapi import FastAPI
from app.services.audit_service import get_audit_service

logger = structlog.get_logger(__name__)


async def start_audit_batch_processor() -> None:
    """
    Inicializa o processador de lotes de auditoria.
    
    Este processo assíncrono será executado em background durante todo o ciclo
    de vida da aplicação, processando eventos de auditoria em lote para melhorar
    performance e reduzir impacto no banco de dados.
    """
    logger.info("Inicializando processador de lotes de auditoria...")
    
    # Obtém instância do serviço de auditoria
    audit_service = await get_audit_service()
    
    # Inicia o processador de lotes
    await audit_service.start_batch_processor()
    
    logger.info("Processador de lotes de auditoria iniciado com sucesso")


async def stop_audit_batch_processor() -> None:
    """
    Finaliza o processador de lotes de auditoria.
    
    Garante que todos os eventos pendentes sejam processados antes do desligamento
    da aplicação para evitar perda de dados.
    """
    logger.info("Finalizando processador de lotes de auditoria...")
    
    # Obtém instância do serviço de auditoria
    audit_service = await get_audit_service()
    
    # Finaliza o processador de lotes, garantindo processamento de eventos pendentes
    await audit_service.stop_batch_processor()
    
    logger.info("Processador de lotes de auditoria finalizado com sucesso")


def create_start_app_handler(app: FastAPI) -> Callable:
    """
    Cria e retorna uma função handler para eventos de inicialização da aplicação.
    
    Args:
        app: Instância da aplicação FastAPI
        
    Returns:
        Função handler para eventos de inicialização
    """
    async def start_app() -> None:
        logger.info("Inicializando serviços na inicialização da aplicação...")
        
        # Iniciar processador de lotes de auditoria
        await start_audit_batch_processor()
        
        logger.info("Todos os serviços inicializados com sucesso")
    
    return start_app


def create_stop_app_handler(app: FastAPI) -> Callable:
    """
    Cria e retorna uma função handler para eventos de desligamento da aplicação.
    
    Args:
        app: Instância da aplicação FastAPI
        
    Returns:
        Função handler para eventos de desligamento
    """
    async def stop_app() -> None:
        logger.info("Finalizando serviços no desligamento da aplicação...")
        
        # Finalizar processador de lotes de auditoria
        await stop_audit_batch_processor()
        
        logger.info("Todos os serviços finalizados com sucesso")
    
    return stop_app


def setup_event_handlers(app: FastAPI) -> None:
    """
    Configura os handlers de eventos para a inicialização e desligamento da aplicação.
    
    Args:
        app: Instância da aplicação FastAPI
    """
    app.add_event_handler("startup", create_start_app_handler(app))
    app.add_event_handler("shutdown", create_stop_app_handler(app))
    
    logger.info("Handlers de eventos configurados com sucesso")