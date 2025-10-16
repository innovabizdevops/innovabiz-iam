"""
Pacote de Agentes IA para Detecção de Fraudes Adaptativas
"""
from .base_agent import FraudDetectionAgent, AgentContext, AgentRegistry
from .orchestrator import AgentOrchestrator

__version__ = "0.1.0"

# Exportar classes principais
__all__ = [
    "FraudDetectionAgent",
    "AgentContext",
    "AgentRegistry",
    "AgentOrchestrator"
]

# Função para configurar o orquestrador a partir de um arquivo de configuração
def setup_orchestrator_from_config(config_file=None, config_dict=None):
    """
    Configura um orquestrador a partir de um arquivo de configuração ou dicionário
    
    Args:
        config_file: Caminho para o arquivo de configuração JSON
        config_dict: Dicionário de configuração
        
    Returns:
        AgentOrchestrator: Uma instância configurada do orquestrador
    """
    import json
    import os
    import logging
    
    logger = logging.getLogger(__name__)
    
    if config_file and os.path.exists(config_file):
        try:
            with open(config_file, 'r') as f:
                config = json.load(f)
            logger.info(f"Configuração carregada do arquivo: {config_file}")
        except Exception as e:
            logger.error(f"Erro ao carregar configuração: {e}")
            config = {}
    elif config_dict:
        config = config_dict
    else:
        logger.warning("Nenhuma configuração fornecida. Usando configuração padrão.")
        config = {
            "min_agents": 1,
            "timeout_seconds": 10,
            "decision_threshold": 0.7,
            "default_agents": [
                {
                    "type": "rules",
                    "agent_id": "default_rules",
                    "priority": 80,
                    "rules": [
                        {
                            "id": "high_risk_country",
                            "name": "País de Alto Risco",
                            "severity": "high",
                            "risk_score": 0.8,
                            "condition_type": "simple",
                            "condition": {
                                "field": "location.country",
                                "operator": "in",
                                "value": ["País Alto Risco 1", "País Alto Risco 2"]
                            }
                        }
                    ]
                },
                {
                    "type": "behavioral",
                    "agent_id": "default_behavioral",
                    "priority": 70,
                    "baseline_features": [
                        "transaction_amount",
                        "temporal_hour",
                        "device_type"
                    ],
                    "anomaly_threshold": 0.75
                }
            ]
        }
    
    # Criar e retornar orquestrador
    return AgentOrchestrator(config)