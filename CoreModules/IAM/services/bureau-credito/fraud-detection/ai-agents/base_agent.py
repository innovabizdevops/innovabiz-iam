"""
Base Agent Module - Fundamental para todos os agentes IA de detecção de fraudes
"""
from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional
import logging
from datetime import datetime
import uuid

logger = logging.getLogger(__name__)

class AgentContext:
    """Contexto compartilhado entre os agentes durante análise de fraude"""
    def __init__(self):
        self.transaction_id = str(uuid.uuid4())
        self.start_time = datetime.now()
        self.insights: Dict[str, Any] = {}
        self.risk_factors: Dict[str, float] = {}
        self.fraud_indicators: List[Dict[str, Any]] = []
        self.trust_score: float = 0.0
        self.decision: Optional[str] = None
        self.confidence: float = 0.0
        self.metadata: Dict[str, Any] = {}
    
    def add_insight(self, agent_id: str, key: str, value: Any) -> None:
        """Adiciona uma nova insight ao contexto"""
        if agent_id not in self.insights:
            self.insights[agent_id] = {}
        self.insights[agent_id][key] = value
    
    def add_risk_factor(self, name: str, score: float) -> None:
        """Adiciona um fator de risco com pontuação associada"""
        self.risk_factors[name] = score
    
    def add_fraud_indicator(self, 
                           indicator_type: str, 
                           severity: str, 
                           description: str,
                           confidence: float) -> None:
        """Adiciona um indicador de fraude ao contexto"""
        self.fraud_indicators.append({
            "type": indicator_type,
            "severity": severity,
            "description": description,
            "confidence": confidence,
            "timestamp": datetime.now().isoformat()
        })
    
    def get_risk_score(self) -> float:
        """Calcula pontuação de risco com base em todos os fatores"""
        if not self.risk_factors:
            return 0.0
        return sum(self.risk_factors.values()) / len(self.risk_factors)


class FraudDetectionAgent(ABC):
    """Classe abstrata base para todos os agentes de detecção de fraude"""
    
    def __init__(self, agent_id: str, config: Dict[str, Any]):
        self.agent_id = agent_id
        self.config = config
        self.enabled = config.get("enabled", True)
        self.priority = config.get("priority", 50)  # Prioridade padrão (0-100)
        self.context: Optional[AgentContext] = None
        self._initialize()
    
    def _initialize(self) -> None:
        """Inicialização customizada para o agente"""
        pass
    
    def set_context(self, context: AgentContext) -> None:
        """Configura o contexto para o agente"""
        self.context = context
    
    @abstractmethod
    def analyze(self, data: Dict[str, Any]) -> None:
        """Método principal para análise de dados"""
        pass
    
    @abstractmethod
    def get_agent_type(self) -> str:
        """Retorna o tipo do agente"""
        pass
    
    def is_applicable(self, data: Dict[str, Any]) -> bool:
        """Determina se este agente é aplicável para os dados fornecidos"""
        return True
    
    def supports_continuous_learning(self) -> bool:
        """Indica se o agente suporta aprendizado contínuo"""
        return False
    
    def train(self, training_data: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Treinamento do modelo (para agentes com aprendizado)"""
        if not self.supports_continuous_learning():
            raise NotImplementedError(
                f"Agent {self.agent_id} does not support continuous learning"
            )
        return {"status": "not_implemented"}


class AgentRegistry:
    """Registro central de agentes de detecção de fraude"""
    
    _instance = None
    _agents: Dict[str, FraudDetectionAgent] = {}
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(AgentRegistry, cls).__new__(cls)
        return cls._instance
    
    def register_agent(self, agent: FraudDetectionAgent) -> None:
        """Registra um agente no registro"""
        if agent.agent_id in self._agents:
            logger.warning(f"Substituindo agente existente com ID: {agent.agent_id}")
        self._agents[agent.agent_id] = agent
        logger.info(f"Agente {agent.agent_id} ({agent.get_agent_type()}) registrado com sucesso")
    
    def get_agent(self, agent_id: str) -> Optional[FraudDetectionAgent]:
        """Recupera um agente pelo ID"""
        return self._agents.get(agent_id)
    
    def get_all_agents(self) -> Dict[str, FraudDetectionAgent]:
        """Recupera todos os agentes registrados"""
        return self._agents
    
    def get_enabled_agents(self) -> List[FraudDetectionAgent]:
        """Recupera apenas agentes habilitados"""
        return [agent for agent in self._agents.values() if agent.enabled]
    
    def get_agents_by_type(self, agent_type: str) -> List[FraudDetectionAgent]:
        """Recupera agentes pelo tipo"""
        return [
            agent for agent in self._agents.values() 
            if agent.get_agent_type() == agent_type
        ]