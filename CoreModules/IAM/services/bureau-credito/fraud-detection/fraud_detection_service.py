"""
Serviço de Detecção de Fraudes - Integra agentes IA com a API do Bureau de Crédito
"""
import logging
import json
import os
from typing import Dict, Any, List, Optional
from datetime import datetime

from ai-agents import AgentOrchestrator, AgentContext, setup_orchestrator_from_config

logger = logging.getLogger(__name__)

class FraudDetectionService:
    """
    Serviço principal de detecção de fraudes que integra os agentes IA
    com o Bureau de Crédito
    """
    
    def __init__(self, config_path: Optional[str] = None):
        self.config_path = config_path
        self.config = self._load_config()
        self.orchestrator = setup_orchestrator_from_config(config_dict=self.config)
        logger.info("Serviço de detecção de fraudes inicializado")
        
    def _load_config(self) -> Dict[str, Any]:
        """Carrega configuração do serviço"""
        default_config = {
            "min_agents": 2,
            "timeout_seconds": 10,
            "decision_threshold": 0.7,
            "log_level": "INFO",
            "telemetry_enabled": True
        }
        
        if not self.config_path or not os.path.exists(self.config_path):
            logger.warning(f"Arquivo de configuração não encontrado: {self.config_path}")
            logger.info("Usando configuração padrão")
            return default_config
            
        try:
            with open(self.config_path, 'r') as f:
                config = json.load(f)
            logger.info(f"Configuração carregada de: {self.config_path}")
            return {**default_config, **config}  # Mesclar com padrões
        except Exception as e:
            logger.error(f"Erro ao carregar configuração: {e}")
            return default_config
    
    def analyze_credit_request(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa uma solicitação de crédito para detectar fraudes
        
        Args:
            request_data: Dados da solicitação de crédito
            
        Returns:
            Resultado da análise incluindo score de risco, decisão, etc.
        """
        logger.info(f"Analisando solicitação de crédito: {request_data.get('request_id', 'N/A')}")
        
        # Enriquecer dados com informações adicionais se necessário
        enriched_data = self._enrich_request_data(request_data)
        
        # Criar contexto específico para solicitação de crédito
        context = AgentContext()
        context.metadata["request_type"] = "credit_request"
        context.metadata["request_id"] = request_data.get("request_id", "unknown")
        context.metadata["tenant_id"] = request_data.get("tenant_id", "unknown")
        
        # Executar análise com orquestrador de agentes
        result = self.orchestrator.analyze(enriched_data, context)
        
        # Adicionar telemetria se habilitado
        if self.config.get("telemetry_enabled", True):
            self._record_telemetry(request_data, result)
            
        return result
    
    def analyze_transaction(self, transaction_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa uma transação para detectar fraudes
        
        Args:
            transaction_data: Dados da transação
            
        Returns:
            Resultado da análise incluindo score de risco, decisão, etc.
        """
        logger.info(f"Analisando transação: {transaction_data.get('transaction_id', 'N/A')}")
        
        # Enriquecer dados com informações adicionais se necessário
        enriched_data = self._enrich_transaction_data(transaction_data)
        
        # Criar contexto específico para transação
        context = AgentContext()
        context.metadata["request_type"] = "transaction"
        context.metadata["transaction_id"] = transaction_data.get("transaction_id", "unknown")
        context.metadata["tenant_id"] = transaction_data.get("tenant_id", "unknown")
        
        # Executar análise com orquestrador de agentes
        result = self.orchestrator.analyze(enriched_data, context)
        
        # Adicionar telemetria se habilitado
        if self.config.get("telemetry_enabled", True):
            self._record_telemetry(transaction_data, result)
            
        return result
    
    def _enrich_request_data(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Enriquece dados de solicitação de crédito com informações adicionais
        como geolocalização, histórico, etc.
        """
        enriched_data = request_data.copy()
        
        # Adicionar timestamp atual se não existir
        if "timestamp" not in enriched_data:
            enriched_data["timestamp"] = datetime.now().isoformat()
            
        # TODO: Implementar enriquecimento adicional
        # - Histórico do usuário
        # - Dados demográficos
        # - Informações geográficas
        
        return enriched_data
    
    def _enrich_transaction_data(self, transaction_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Enriquece dados de transação com informações adicionais
        como geolocalização, histórico, etc.
        """
        enriched_data = transaction_data.copy()
        
        # Adicionar timestamp atual se não existir
        if "timestamp" not in enriched_data:
            enriched_data["timestamp"] = datetime.now().isoformat()
            
        # TODO: Implementar enriquecimento adicional
        # - Histórico de transações
        # - Padrões de gasto
        # - Informações de dispositivo/localização
        
        return enriched_data
    
    def _record_telemetry(self, request_data: Dict[str, Any], result: Dict[str, Any]) -> None:
        """Registra dados de telemetria para análise posterior"""
        try:
            telemetry_data = {
                "timestamp": datetime.now().isoformat(),
                "request_type": result.get("metadata", {}).get("request_type", "unknown"),
                "tenant_id": request_data.get("tenant_id", "unknown"),
                "risk_score": result.get("risk_score", 0),
                "decision": result.get("decision", "unknown"),
                "fraud_indicators_count": len(result.get("fraud_indicators", [])),
                "execution_time_ms": result.get("execution_time_ms", 0),
                "agents_executed": result.get("agents_executed", [])
            }
            
            # TODO: Enviar telemetria para sistema de monitoramento
            logger.debug(f"Telemetria: {json.dumps(telemetry_data)}")
            
        except Exception as e:
            logger.error(f"Erro ao registrar telemetria: {e}")
    
    def train_models(self, training_data: Dict[str, List[Dict[str, Any]]]) -> Dict[str, Any]:
        """
        Treina modelos de agentes com dados históricos
        
        Args:
            training_data: Dicionário mapeando IDs de agentes para seus dados de treinamento
            
        Returns:
            Resultados do treinamento por agente
        """
        return self.orchestrator.train_agents(training_data)
    
    def get_service_status(self) -> Dict[str, Any]:
        """Retorna o status atual do serviço de detecção de fraudes"""
        agent_status = self.orchestrator.get_agent_status()
        
        return {
            "status": "online",
            "version": "0.1.0",
            "config": {
                "min_agents": self.config.get("min_agents"),
                "timeout_seconds": self.config.get("timeout_seconds"),
                "decision_threshold": self.config.get("decision_threshold"),
                "telemetry_enabled": self.config.get("telemetry_enabled")
            },
            "agent_status": agent_status,
            "uptime_seconds": 0  # TODO: Implementar tracking de uptime
        }