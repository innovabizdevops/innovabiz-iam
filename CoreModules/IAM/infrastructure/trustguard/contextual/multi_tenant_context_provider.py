"""
Provedor de Contexto Multi-Tenant para o TrustGuard

Este módulo implementa um provedor de contexto que permite análise contextual
avançada por tenant, permitindo personalização de regras e comportamentos
de segurança por mercado, região e cliente.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Set, Tuple
import json
import datetime
import hashlib
from dataclasses import dataclass

from ...observability.core.multi_layer_monitor import MultiLayerMonitor
from ...observability.logging.hook_logger import Logger
from ...observability.metrics.hook_metrics import MetricsCollector
from ...constants.constants import MERCADOS, REGIOES, NIVEIS_RISCO

@dataclass
class TenantConfig:
    """Configuração de um tenant específico."""
    tenant_id: str
    nome: str
    mercados: List[str]
    regioes: List[str]
    nivel_seguranca_padrao: str
    fatores_autenticacao_obrigatorios: List[str]
    limites_transacao: Dict[str, float]
    compliance_schemas: List[str]
    configuracao_personalizada: Dict[str, Any]
    regras_avaliacao: Dict[str, Dict[str, Any]]

@dataclass
class ContextoAnalise:
    """Contexto para análise de segurança."""
    tenant_id: str
    usuario_id: str
    ip_origem: str
    dispositivo_hash: str
    localizacao: Dict[str, str]
    timestamp: datetime.datetime
    canal: str
    tipo_transacao: Optional[str] = None
    valor_transacao: Optional[float] = None
    destinatario: Optional[Dict[str, Any]] = None
    historico_recente: Optional[List[Dict[str, Any]]] = None
    metadados: Optional[Dict[str, Any]] = None

class MultiTenantContextProvider:
    """
    Provedor de contexto multi-tenant para o TrustGuard.
    
    Responsável por carregar, gerenciar e disponibilizar configurações específicas
    de cada tenant para análise contextual de segurança.
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o provedor de contexto multi-tenant.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        self.tenants: Dict[str, TenantConfig] = {}
        self.observability = observability_monitor
        self.logger = observability_monitor.get_logger() if observability_monitor else Logger()
        self.metrics = observability_monitor.get_metrics_collector() if observability_monitor else MetricsCollector()
        
        self.logger.info("MultiTenantContextProvider inicializado")
    
    def registrar_tenant(self, config: TenantConfig) -> bool:
        """
        Registra um novo tenant no sistema.
        
        Args:
            config: Configuração do tenant
            
        Returns:
            bool: True se o tenant foi registrado com sucesso
        """
        try:
            if config.tenant_id in self.tenants:
                self.logger.warn(f"Tenant {config.tenant_id} já existe. Atualizando configuração.")
                
            # Validar configuração
            self._validar_configuracao_tenant(config)
            
            # Registrar tenant
            self.tenants[config.tenant_id] = config
            
            self.logger.info(f"Tenant {config.tenant_id} registrado com sucesso")
            self.metrics.incrementCounter("trustguard.tenant.register", 
                                        {"tenant_id": config.tenant_id, "success": "true"})
            
            return True
        except Exception as e:
            self.logger.error(f"Erro ao registrar tenant {config.tenant_id}: {str(e)}")
            self.metrics.incrementCounter("trustguard.tenant.register", 
                                        {"tenant_id": config.tenant_id, "success": "false", "error": str(e)})
            return False
    
    def _validar_configuracao_tenant(self, config: TenantConfig) -> None:
        """
        Valida a configuração de um tenant.
        
        Args:
            config: Configuração do tenant
            
        Raises:
            ValueError: Se a configuração for inválida
        """
        # Validar mercados
        for mercado in config.mercados:
            if mercado not in MERCADOS:
                raise ValueError(f"Mercado inválido: {mercado}")
        
        # Validar regiões
        for regiao in config.regioes:
            if regiao not in REGIOES:
                raise ValueError(f"Região inválida: {regiao}")
        
        # Validar nível de segurança
        if config.nivel_seguranca_padrao not in NIVEIS_RISCO:
            raise ValueError(f"Nível de segurança inválido: {config.nivel_seguranca_padrao}")
        
        # Validar fatores de autenticação
        fatores_validos = {"senha", "otp", "biometria", "certificado", "token_fisico", "geolocalizacao", "comportamental"}
        for fator in config.fatores_autenticacao_obrigatorios:
            if fator not in fatores_validos:
                raise ValueError(f"Fator de autenticação inválido: {fator}")
    
    def obter_tenant(self, tenant_id: str) -> Optional[TenantConfig]:
        """
        Obtém a configuração de um tenant específico.
        
        Args:
            tenant_id: ID do tenant
            
        Returns:
            Optional[TenantConfig]: Configuração do tenant ou None se não existir
        """
        if tenant_id in self.tenants:
            self.metrics.incrementCounter("trustguard.tenant.access", {"tenant_id": tenant_id})
            return self.tenants[tenant_id]
        else:
            self.logger.warn(f"Tentativa de acesso a tenant inexistente: {tenant_id}")
            self.metrics.incrementCounter("trustguard.tenant.access_error", {"tenant_id": tenant_id, "error": "not_found"})
            return None
    
    def enriquecer_contexto(self, contexto: ContextoAnalise) -> ContextoAnalise:
        """
        Enriquece o contexto de análise com informações específicas do tenant.
        
        Args:
            contexto: Contexto inicial de análise
            
        Returns:
            ContextoAnalise: Contexto enriquecido
        """
        tenant_config = self.obter_tenant(contexto.tenant_id)
        if not tenant_config:
            self.logger.error(f"Tenant não encontrado para enriquecimento de contexto: {contexto.tenant_id}")
            return contexto
        
        # Adicionar informações de mercado e região ao metadados
        if contexto.metadados is None:
            contexto.metadados = {}
        
        # Adicionar configurações específicas do tenant aos metadados
        contexto.metadados.update({
            "tenant_mercados": tenant_config.mercados,
            "tenant_regioes": tenant_config.regioes,
            "tenant_nivel_seguranca_padrao": tenant_config.nivel_seguranca_padrao,
            "tenant_compliance_schemas": tenant_config.compliance_schemas,
        })
        
        # Enriquecer com geolocalização e dados específicos da região
        if "pais" in contexto.localizacao:
            pais = contexto.localizacao["pais"]
            if pais in tenant_config.configuracao_personalizada.get("paises_config", {}):
                contexto.metadados["pais_config"] = tenant_config.configuracao_personalizada["paises_config"][pais]
        
        # Enriquecer com informações de transação específicas do tenant
        if contexto.tipo_transacao and contexto.tipo_transacao in tenant_config.configuracao_personalizada.get("transacoes_config", {}):
            contexto.metadados["transacao_config"] = tenant_config.configuracao_personalizada["transacoes_config"][contexto.tipo_transacao]
        
        return contexto
    
    def calcular_hash_contexto(self, contexto: ContextoAnalise) -> str:
        """
        Calcula um hash único para o contexto de análise.
        
        Args:
            contexto: Contexto de análise
            
        Returns:
            str: Hash do contexto
        """
        # Criar representação serializável do contexto
        ctx_dict = {
            "tenant_id": contexto.tenant_id,
            "usuario_id": contexto.usuario_id,
            "ip_origem": contexto.ip_origem,
            "dispositivo_hash": contexto.dispositivo_hash,
            "localizacao": contexto.localizacao,
            "timestamp": contexto.timestamp.isoformat(),
            "canal": contexto.canal,
            "tipo_transacao": contexto.tipo_transacao,
            "valor_transacao": contexto.valor_transacao,
        }
        
        # Calcular hash SHA-256
        ctx_str = json.dumps(ctx_dict, sort_keys=True)
        return hashlib.sha256(ctx_str.encode()).hexdigest()
    
    def compativel_com_mercado(self, tenant_id: str, mercado: str) -> bool:
        """
        Verifica se um tenant é compatível com um determinado mercado.
        
        Args:
            tenant_id: ID do tenant
            mercado: Código do mercado
            
        Returns:
            bool: True se o tenant é compatível com o mercado
        """
        tenant = self.obter_tenant(tenant_id)
        if not tenant:
            return False
        
        return mercado in tenant.mercados
    
    def obter_tenants_por_mercado(self, mercado: str) -> List[str]:
        """
        Obtém a lista de tenants compatíveis com um determinado mercado.
        
        Args:
            mercado: Código do mercado
            
        Returns:
            List[str]: Lista de IDs de tenants compatíveis
        """
        return [tid for tid, tenant in self.tenants.items() if mercado in tenant.mercados]
    
    def obter_tenants_por_regiao(self, regiao: str) -> List[str]:
        """
        Obtém a lista de tenants compatíveis com uma determinada região.
        
        Args:
            regiao: Código da região
            
        Returns:
            List[str]: Lista de IDs de tenants compatíveis
        """
        return [tid for tid, tenant in self.tenants.items() if regiao in tenant.regioes]
    
    def obter_regras_avaliacao(self, tenant_id: str, tipo_regra: str) -> Dict[str, Any]:
        """
        Obtém as regras de avaliação específicas de um tenant para um tipo de regra.
        
        Args:
            tenant_id: ID do tenant
            tipo_regra: Tipo de regra (autenticacao, transacao, acesso, etc.)
            
        Returns:
            Dict[str, Any]: Configuração das regras ou dicionário vazio se não encontrado
        """
        tenant = self.obter_tenant(tenant_id)
        if not tenant:
            return {}
        
        return tenant.regras_avaliacao.get(tipo_regra, {})
