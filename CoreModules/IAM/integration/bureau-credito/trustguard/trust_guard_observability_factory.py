"""
Factory para integração de observabilidade com TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
from typing import Optional

from observability.rules.rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator
from observability.rules.rules_observability_monitor import RulesObservabilityMonitor, instrument_connector

from .trust_guard_connector import TrustGuardConnector
from .trust_guard_factory import TrustGuardFactory


class TrustGuardObservabilityFactory:
    """
    Factory para criar e instrumentar conectores TrustGuard com observabilidade.
    
    Integra a criação do conector TrustGuard com a configuração de observabilidade
    para fornecer monitoramento de métricas, logs e traces para as operações
    do TrustGuard.
    """
    
    @staticmethod
    def create_monitored_connector(
        component_name: str = "trustguard",
        logger_name: str = "innovabiz.iam.trustguard",
        api_url: Optional[str] = None,
        api_key: Optional[str] = None,
        tenant_id: Optional[str] = None,
        timeout: Optional[float] = None,
        cache_ttl: Optional[int] = None,
        cache_enabled: Optional[bool] = None,
    ) -> TrustGuardConnector:
        """
        Cria um conector TrustGuard instrumentado com observabilidade.
        
        Args:
            component_name: Nome do componente para telemetria
            logger_name: Nome do logger
            api_url: URL da API do TrustGuard (opcional, lê de env se não fornecido)
            api_key: Chave de API do TrustGuard (opcional, lê de env se não fornecido)
            tenant_id: ID do tenant (opcional, lê de env se não fornecido)
            timeout: Timeout para requisições HTTP (opcional, lê de env se não fornecido)
            cache_ttl: Tempo de vida do cache (opcional, lê de env se não fornecido)
            cache_enabled: Se o cache está habilitado (opcional, lê de env se não fornecido)
            
        Returns:
            TrustGuardConnector: Conector instrumentado com observabilidade
        """
        # Criar conector TrustGuard
        connector = TrustGuardFactory.create_connector(
            api_url=api_url,
            api_key=api_key,
            tenant_id=tenant_id,
            timeout=timeout,
            cache_ttl=cache_ttl,
            cache_enabled=cache_enabled,
        )
        
        # Criar configurador de observabilidade
        configurator = RulesObservabilityConfigurator.from_env(logger_name)
        
        # Criar monitor de observabilidade
        monitor = RulesObservabilityMonitor(configurator, component_name)
        
        # Instrumentar conector
        return instrument_connector(connector, monitor)
    
    @staticmethod
    async def create_monitored_connector_async(
        component_name: str = "trustguard",
        logger_name: str = "innovabiz.iam.trustguard",
        api_url: Optional[str] = None,
        api_key: Optional[str] = None,
        tenant_id: Optional[str] = None,
        timeout: Optional[float] = None,
        cache_ttl: Optional[int] = None,
        cache_enabled: Optional[bool] = None,
    ) -> TrustGuardConnector:
        """
        Cria um conector TrustGuard instrumentado com observabilidade de forma assíncrona.
        
        Esta versão assíncrona inicializa o conector de forma assíncrona.
        
        Args:
            component_name: Nome do componente para telemetria
            logger_name: Nome do logger
            api_url: URL da API do TrustGuard (opcional, lê de env se não fornecido)
            api_key: Chave de API do TrustGuard (opcional, lê de env se não fornecido)
            tenant_id: ID do tenant (opcional, lê de env se não fornecido)
            timeout: Timeout para requisições HTTP (opcional, lê de env se não fornecido)
            cache_ttl: Tempo de vida do cache (opcional, lê de env se não fornecido)
            cache_enabled: Se o cache está habilitado (opcional, lê de env se não fornecido)
            
        Returns:
            TrustGuardConnector: Conector instrumentado com observabilidade
        """
        # Criar conector TrustGuard de forma assíncrona
        connector = await TrustGuardFactory.create_connector_async(
            api_url=api_url,
            api_key=api_key,
            tenant_id=tenant_id,
            timeout=timeout,
            cache_ttl=cache_ttl,
            cache_enabled=cache_enabled,
        )
        
        # Criar configurador de observabilidade
        configurator = RulesObservabilityConfigurator.from_env(logger_name)
        
        # Criar monitor de observabilidade
        monitor = RulesObservabilityMonitor(configurator, component_name)
        
        # Instrumentar conector
        return instrument_connector(connector, monitor)


# Exemplo de uso:
# async def main():
#     connector = await TrustGuardObservabilityFactory.create_monitored_connector_async()
#     result = await connector.evaluate_access(access_request)
#     print(f"Decision: {result.decision}, Risk Level: {result.risk_level}")