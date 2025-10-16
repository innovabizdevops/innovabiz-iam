"""
Orquestrador de Agentes - Coordena a execução de múltiplos agentes para detecção de fraude
"""
from typing import Dict, Any, List, Optional, Tuple
import logging
import threading
import time
import json
from datetime import datetime
from .base_agent import FraudDetectionAgent, AgentContext, AgentRegistry

logger = logging.getLogger(__name__)

class AgentOrchestrator:
    """Orquestrador que coordena múltiplos agentes de detecção de fraude"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.min_agents = config.get("min_agents", 1)
        self.timeout_seconds = config.get("timeout_seconds", 10)
        self.decision_threshold = config.get("decision_threshold", 0.7)
        self.registry = AgentRegistry()
        self.default_agent_configs = config.get("default_agents", [])
        self._load_default_agents()
        
    def _load_default_agents(self) -> None:
        """Carrega agentes padrão a partir da configuração"""
        from . import behavioral_agent, rules_agent, ml_agent
        
        for agent_config in self.default_agent_configs:
            agent_type = agent_config.get("type")
            
            if agent_type == "behavioral":
                behavioral_agent.register_agent(agent_config)
            elif agent_type == "rules":
                rules_agent.register_agent(agent_config)
            elif agent_type == "ml":
                ml_agent.register_agent(agent_config)
            else:
                logger.warning(f"Tipo de agente desconhecido: {agent_type}")
                
        logger.info(f"Carregados {len(self.default_agent_configs)} agentes padrão")
    
    def analyze(self, data: Dict[str, Any], 
               context: Optional[AgentContext] = None) -> Dict[str, Any]:
        """
        Executa análise de fraude usando todos os agentes disponíveis
        
        Args:
            data: Dados a serem analisados
            context: Contexto opcional pré-existente
            
        Returns:
            Resultado da análise incluindo score de confiança, decisão, etc.
        """
        # Criar ou usar contexto fornecido
        ctx = context or AgentContext()
        
        # Listar agentes habilitados e ordenar por prioridade
        agents = sorted(
            self.registry.get_enabled_agents(),
            key=lambda a: a.priority,
            reverse=True  # Maior prioridade primeiro
        )
        
        if len(agents) < self.min_agents:
            logger.error(f"Número insuficiente de agentes ({len(agents)}/{self.min_agents})")
            return {
                "status": "error",
                "message": f"Número insuficiente de agentes ({len(agents)}/{self.min_agents})",
                "timestamp": datetime.now().isoformat()
            }
        
        # Preparar todos os agentes com o contexto compartilhado
        for agent in agents:
            agent.set_context(ctx)
        
        # Executar agentes (com timeout)
        results = self._run_agents_with_timeout(agents, data)
        
        # Processar resultados e tomar decisão final
        total_risk = ctx.get_risk_score()
        fraud_indicators = ctx.fraud_indicators
        
        # Determinar decisão final
        decision = "approve"
        if total_risk > self.decision_threshold:
            decision = "reject"
        elif total_risk > self.decision_threshold * 0.7:
            decision = "review"
            
        # Calcular confiança na decisão
        decision_confidence = 0.5
        if decision == "approve":
            decision_confidence = 1.0 - total_risk
        elif decision == "reject":
            decision_confidence = total_risk
        else:  # review
            # Para casos ambíguos, a confiança é mais baixa
            decision_confidence = 0.5 - abs(0.5 - total_risk)
            
        # Construir resultado final
        ctx.decision = decision
        ctx.confidence = decision_confidence
        
        # Retornar resultados da análise
        return {
            "status": "success",
            "transaction_id": ctx.transaction_id,
            "timestamp": datetime.now().isoformat(),
            "risk_score": total_risk,
            "decision": decision,
            "confidence": decision_confidence,
            "fraud_indicators": fraud_indicators,
            "insights": ctx.insights,
            "agents_executed": [agent.agent_id for agent in agents],
            "execution_time_ms": (datetime.now() - ctx.start_time).total_seconds() * 1000
        }
    
    def _run_agents_with_timeout(self, 
                              agents: List[FraudDetectionAgent],
                              data: Dict[str, Any]) -> Dict[str, Any]:
        """Executa todos os agentes com um timeout global"""
        results = {}
        threads = []
        thread_errors = {}
        
        # Criar thread para cada agente
        for agent in agents:
            if not agent.is_applicable(data):
                logger.debug(f"Agente {agent.agent_id} não é aplicável para os dados")
                continue
                
            t = threading.Thread(
                target=self._run_agent_thread,
                args=(agent, data, results, thread_errors)
            )
            threads.append(t)
            t.daemon = True
            t.start()
        
        # Aguardar conclusão com timeout
        start_time = time.time()
        for t in threads:
            remaining_time = self.timeout_seconds - (time.time() - start_time)
            if remaining_time <= 0:
                logger.warning("Timeout atingido durante execução dos agentes")
                break
                
            t.join(timeout=remaining_time)
        
        # Verificar erros
        for agent_id, error in thread_errors.items():
            logger.error(f"Erro no agente {agent_id}: {error}")
            
        return results
    
    def _run_agent_thread(self, 
                        agent: FraudDetectionAgent,
                        data: Dict[str, Any],
                        results: Dict[str, Any],
                        errors: Dict[str, str]) -> None:
        """Função auxiliar para execução de agente em thread"""
        try:
            start_time = time.time()
            agent.analyze(data)
            execution_time = time.time() - start_time
            
            results[agent.agent_id] = {
                "success": True,
                "execution_time": execution_time
            }
            
            logger.debug(f"Agente {agent.agent_id} concluído em {execution_time:.2f}s")
            
        except Exception as e:
            errors[agent.agent_id] = str(e)
            results[agent.agent_id] = {
                "success": False,
                "error": str(e)
            }
    
    def train_agents(self, training_data: Dict[str, List[Dict[str, Any]]]) -> Dict[str, Any]:
        """Treina agentes que suportam aprendizado contínuo"""
        results = {}
        
        for agent_id, agent_data in training_data.items():
            agent = self.registry.get_agent(agent_id)
            if not agent:
                results[agent_id] = {
                    "status": "error",
                    "message": f"Agente não encontrado: {agent_id}"
                }
                continue
                
            if not agent.supports_continuous_learning():
                results[agent_id] = {
                    "status": "error",
                    "message": f"Agente {agent_id} não suporta aprendizado contínuo"
                }
                continue
                
            try:
                train_result = agent.train(agent_data)
                results[agent_id] = {
                    "status": "success",
                    "result": train_result
                }
            except Exception as e:
                results[agent_id] = {
                    "status": "error",
                    "message": str(e)
                }
                
        return results
    
    def get_agent_status(self) -> Dict[str, Any]:
        """Retorna status de todos os agentes registrados"""
        agents = self.registry.get_all_agents()
        
        return {
            "total_agents": len(agents),
            "enabled_agents": len([a for a in agents.values() if a.enabled]),
            "agent_types": {
                agent_type: len([a for a in agents.values() 
                               if a.get_agent_type() == agent_type])
                for agent_type in set(a.get_agent_type() for a in agents.values())
            },
            "agents": [
                {
                    "id": agent.agent_id,
                    "type": agent.get_agent_type(),
                    "enabled": agent.enabled,
                    "priority": agent.priority,
                    "supports_learning": agent.supports_continuous_learning()
                }
                for agent in agents.values()
            ]
        }