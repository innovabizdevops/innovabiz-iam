"""
Agente baseado em regras - Implementa detecção de fraude usando regras configuráveis
"""
from typing import Dict, Any, List, Optional, Callable
import logging
import json
import re
from datetime import datetime, timedelta
from .base_agent import FraudDetectionAgent, AgentRegistry

logger = logging.getLogger(__name__)

class Rule:
    """Representa uma regra individual para detecção de fraude"""
    
    def __init__(self, rule_id: str, config: Dict[str, Any]):
        self.rule_id = rule_id
        self.name = config.get("name", rule_id)
        self.description = config.get("description", "")
        self.severity = config.get("severity", "medium")
        self.condition_type = config.get("condition_type", "simple")
        self.condition = config.get("condition", {})
        self.risk_score = config.get("risk_score", 0.5)
        self.tags = config.get("tags", [])
        self.enabled = config.get("enabled", True)
    
    def evaluate(self, data: Dict[str, Any]) -> bool:
        """Avalia se a regra foi violada com base nos dados"""
        if self.condition_type == "simple":
            return self._evaluate_simple_condition(data)
        elif self.condition_type == "complex":
            return self._evaluate_complex_condition(data)
        elif self.condition_type == "regex":
            return self._evaluate_regex_condition(data)
        elif self.condition_type == "threshold":
            return self._evaluate_threshold_condition(data)
        else:
            logger.warning(f"Tipo de condição desconhecido: {self.condition_type}")
            return False
    
    def _evaluate_simple_condition(self, data: Dict[str, Any]) -> bool:
        """Avalia uma condição simples baseada em igualdade ou comparação"""
        field = self.condition.get("field")
        if not field:
            return False
            
        # Suporte a campo aninhado com notação de ponto
        value = self._get_nested_value(data, field)
        if value is None:
            return False
            
        operator = self.condition.get("operator", "eq")
        expected = self.condition.get("value")
        
        if operator == "eq":
            return value == expected
        elif operator == "neq":
            return value != expected
        elif operator == "gt":
            return value > expected
        elif operator == "lt":
            return value < expected
        elif operator == "gte":
            return value >= expected
        elif operator == "lte":
            return value <= expected
        elif operator == "in":
            return value in expected
        elif operator == "contains":
            return expected in value
        elif operator == "exists":
            return True  # O valor existe se chegamos aqui
        elif operator == "not_exists":
            return False  # Nunca será alcançado pois retornamos False se value é None
        else:
            logger.warning(f"Operador desconhecido: {operator}")
            return False
            
    def _evaluate_complex_condition(self, data: Dict[str, Any]) -> bool:
        """Avalia uma condição complexa baseada em operadores lógicos AND/OR"""
        logical_op = self.condition.get("logical_operator", "AND")
        conditions = self.condition.get("conditions", [])
        
        if not conditions:
            return False
            
        results = []
        for sub_condition in conditions:
            sub_rule = Rule(f"{self.rule_id}_sub", {
                "condition_type": sub_condition.get("type", "simple"),
                "condition": sub_condition
            })
            results.append(sub_rule.evaluate(data))
            
        if logical_op == "AND":
            return all(results)
        elif logical_op == "OR":
            return any(results)
        else:
            logger.warning(f"Operador lógico desconhecido: {logical_op}")
            return False
            
    def _evaluate_regex_condition(self, data: Dict[str, Any]) -> bool:
        """Avalia uma condição baseada em expressões regulares"""
        field = self.condition.get("field")
        pattern = self.condition.get("pattern")
        
        if not field or not pattern:
            return False
            
        value = self._get_nested_value(data, field)
        if value is None or not isinstance(value, str):
            return False
            
        try:
            regex = re.compile(pattern)
            return bool(regex.search(value))
        except re.error as e:
            logger.error(f"Erro na expressão regular: {e}")
            return False
            
    def _evaluate_threshold_condition(self, data: Dict[str, Any]) -> bool:
        """Avalia uma condição de limiar com múltiplos campos"""
        threshold = self.condition.get("threshold", 0)
        weights = self.condition.get("weights", {})
        total_score = 0
        
        for field, weight in weights.items():
            value = self._get_nested_value(data, field)
            if value is not None:
                if isinstance(value, bool):
                    value = 1 if value else 0
                if isinstance(value, (int, float)):
                    total_score += value * weight
                    
        return total_score >= threshold
    
    def _get_nested_value(self, data: Dict[str, Any], field_path: str) -> Any:
        """Recupera um valor aninhado usando notação de ponto"""
        if not field_path:
            return None
            
        parts = field_path.split(".")
        current = data
        
        for part in parts:
            if isinstance(current, dict) and part in current:
                current = current[part]
            else:
                return None
                
        return current


class RuleBasedAgent(FraudDetectionAgent):
    """Agente que utiliza um conjunto de regras para detecção de fraudes"""
    
    def __init__(self, agent_id: str, config: Dict[str, Any]):
        super().__init__(agent_id, config)
        self.rules: Dict[str, Rule] = {}
        self._load_rules()
        
    def _initialize(self) -> None:
        """Inicialização específica para agente baseado em regras"""
        logger.info(f"Inicializando agente baseado em regras: {self.agent_id}")
        
    def get_agent_type(self) -> str:
        """Retorna o tipo do agente"""
        return "rule_based"
        
    def _load_rules(self) -> None:
        """Carrega as regras a partir da configuração"""
        rules_config = self.config.get("rules", [])
        for rule_config in rules_config:
            rule_id = rule_config.get("id")
            if rule_id:
                self.rules[rule_id] = Rule(rule_id, rule_config)
                logger.debug(f"Regra carregada: {rule_id}")
                
        rules_file = self.config.get("rules_file")
        if rules_file:
            try:
                with open(rules_file, "r") as f:
                    file_rules = json.load(f)
                    for rule_config in file_rules:
                        rule_id = rule_config.get("id")
                        if rule_id:
                            self.rules[rule_id] = Rule(rule_id, rule_config)
                            logger.debug(f"Regra carregada do arquivo: {rule_id}")
            except Exception as e:
                logger.error(f"Erro ao carregar regras do arquivo {rules_file}: {e}")
    
    def add_rule(self, rule_config: Dict[str, Any]) -> Optional[str]:
        """Adiciona uma nova regra ao agente"""
        rule_id = rule_config.get("id")
        if not rule_id:
            rule_id = f"rule_{len(self.rules) + 1}"
            rule_config["id"] = rule_id
            
        self.rules[rule_id] = Rule(rule_id, rule_config)
        logger.info(f"Nova regra adicionada: {rule_id}")
        return rule_id
        
    def remove_rule(self, rule_id: str) -> bool:
        """Remove uma regra do agente"""
        if rule_id in self.rules:
            del self.rules[rule_id]
            logger.info(f"Regra removida: {rule_id}")
            return True
        return False
        
    def get_rules(self) -> List[Dict[str, Any]]:
        """Retorna todas as regras configuradas"""
        return [
            {
                "id": rule.rule_id,
                "name": rule.name,
                "description": rule.description,
                "severity": rule.severity,
                "enabled": rule.enabled
            }
            for rule in self.rules.values()
        ]
        
    def analyze(self, data: Dict[str, Any]) -> None:
        """Analisa os dados em busca de violações das regras"""
        if not self.context:
            logger.error("Contexto não disponível para análise")
            return
            
        violated_rules = []
        total_risk_score = 0.0
        
        # Avaliar cada regra
        for rule_id, rule in self.rules.items():
            if not rule.enabled:
                continue
                
            try:
                if rule.evaluate(data):
                    violated_rules.append(rule)
                    total_risk_score += rule.risk_score
            except Exception as e:
                logger.error(f"Erro ao avaliar regra {rule_id}: {e}")
                
        # Adicionar insights ao contexto
        if violated_rules:
            violations = [
                {
                    "rule_id": rule.rule_id,
                    "name": rule.name,
                    "severity": rule.severity,
                    "risk_score": rule.risk_score
                }
                for rule in violated_rules
            ]
            
            self.context.add_insight(
                self.agent_id,
                "rule_violations",
                violations
            )
            
            # Normalizar o score de risco (máximo de 1.0)
            normalized_risk = min(1.0, total_risk_score)
            
            # Adicionar fator de risco
            self.context.add_risk_factor("rule_violations", normalized_risk)
            
            # Adicionar indicadores de fraude para violações de alta severidade
            for rule in violated_rules:
                if rule.severity in ["high", "critical"]:
                    self.context.add_fraud_indicator(
                        indicator_type="rule_violation",
                        severity=rule.severity,
                        description=f"Violação de regra: {rule.name}",
                        confidence=rule.risk_score
                    )
                    
        # Adicionar métricas gerais
        self.context.add_insight(
            self.agent_id,
            "rules_summary",
            {
                "total_rules": len(self.rules),
                "evaluated_rules": len([r for r in self.rules.values() if r.enabled]),
                "violated_rules": len(violated_rules),
                "risk_score": normalized_risk if violated_rules else 0.0
            }
        )


# Registrar o agente quando o módulo for importado
def register_agent(config: Dict[str, Any]) -> RuleBasedAgent:
    """Cria e registra o agente baseado em regras"""
    agent_id = config.get("agent_id", "rules_default")
    agent = RuleBasedAgent(agent_id, config)
    AgentRegistry().register_agent(agent)
    return agent