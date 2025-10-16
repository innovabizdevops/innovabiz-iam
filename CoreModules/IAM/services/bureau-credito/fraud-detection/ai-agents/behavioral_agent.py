"""
Agente de Detecção Comportamental - Analisa padrões de comportamento para identificar anomalias
"""
from typing import Dict, Any, List, Optional
import numpy as np
from datetime import datetime, timedelta
import logging
from .base_agent import FraudDetectionAgent, AgentRegistry

logger = logging.getLogger(__name__)

class BehavioralDetectionAgent(FraudDetectionAgent):
    """Agente que analisa padrões comportamentais para detecção de fraudes"""
    
    def __init__(self, agent_id: str, config: Dict[str, Any]):
        super().__init__(agent_id, config)
        self.baseline_features = config.get("baseline_features", [])
        self.anomaly_threshold = config.get("anomaly_threshold", 0.75)
        self.min_history_points = config.get("min_history_points", 5)
        self.feature_weights = config.get("feature_weights", {})
        self.user_profiles: Dict[str, Dict[str, Any]] = {}
        
    def _initialize(self) -> None:
        """Configuração específica do agente comportamental"""
        logger.info(f"Inicializando agente comportamental: {self.agent_id}")
        
    def get_agent_type(self) -> str:
        """Retorna o tipo do agente"""
        return "behavioral_detection"
    
    def supports_continuous_learning(self) -> bool:
        """Este agente suporta aprendizado contínuo"""
        return True
        
    def _calculate_feature_importance(self, feature: str) -> float:
        """Calcula a importância de uma característica específica"""
        return self.feature_weights.get(feature, 1.0)
    
    def _get_user_profile(self, user_id: str) -> Dict[str, Any]:
        """Recupera o perfil do usuário, inicializando se necessário"""
        if user_id not in self.user_profiles:
            self.user_profiles[user_id] = {
                "history": [],
                "last_update": datetime.now(),
                "anomaly_score": 0.0,
                "baseline": {},
                "patterns": {}
            }
        return self.user_profiles[user_id]
    
    def _extract_features(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Extrai características relevantes dos dados de entrada"""
        features = {}
        
        # Características temporais
        current_time = datetime.now()
        hour_of_day = current_time.hour
        day_of_week = current_time.weekday()
        features["temporal_hour"] = hour_of_day
        features["temporal_day"] = day_of_week
        
        # Características de localização (se disponíveis)
        if "location" in data:
            features["location_country"] = data["location"].get("country", "unknown")
            features["location_region"] = data["location"].get("region", "unknown")
            
        # Características de dispositivo (se disponíveis)
        if "device" in data:
            features["device_type"] = data["device"].get("type", "unknown")
            features["device_os"] = data["device"].get("os", "unknown")
            
        # Características de transação (se disponíveis)
        if "transaction" in data:
            features["transaction_amount"] = data["transaction"].get("amount", 0.0)
            features["transaction_currency"] = data["transaction"].get("currency", "unknown")
            features["transaction_type"] = data["transaction"].get("type", "unknown")
            
        return features
    
    def _calculate_anomaly_score(self, 
                               user_id: str, 
                               features: Dict[str, Any]) -> float:
        """Calcula o score de anomalia para os recursos extraídos"""
        profile = self._get_user_profile(user_id)
        
        if not profile["history"] or len(profile["history"]) < self.min_history_points:
            # Sem histórico suficiente para calcular anomalias
            return 0.3  # Score moderado por padrão para novos usuários
            
        # Calcular desvio para cada característica
        feature_deviations = {}
        anomaly_score = 0.0
        total_weight = 0.0
        
        for feature_name, current_value in features.items():
            if feature_name not in self.baseline_features:
                continue
                
            # Extrair histórico desta característica
            feature_history = [
                h[feature_name] for h in profile["history"] 
                if feature_name in h
            ]
            
            if not feature_history:
                continue
                
            # Calcular desvio com base no tipo de dado
            deviation = 0.0
            weight = self._calculate_feature_importance(feature_name)
            
            if isinstance(current_value, (int, float)):
                # Características numéricas
                mean_val = np.mean(feature_history)
                std_val = np.std(feature_history) or 1.0  # Evitar divisão por zero
                z_score = abs(current_value - mean_val) / std_val
                deviation = min(1.0, z_score / 3.0)  # Normalizado para 0-1
                
            elif isinstance(current_value, str):
                # Características categóricas
                counts = {}
                for val in feature_history:
                    counts[val] = counts.get(val, 0) + 1
                    
                frequency = counts.get(current_value, 0) / len(feature_history)
                deviation = 1.0 - frequency  # Quanto menor a frequência, maior a anomalia
            
            feature_deviations[feature_name] = deviation
            anomaly_score += deviation * weight
            total_weight += weight
            
        # Calcular pontuação final
        if total_weight > 0:
            final_score = anomaly_score / total_weight
        else:
            final_score = 0.3  # Score moderado por padrão
            
        return final_score
    
    def _update_user_profile(self, 
                           user_id: str, 
                           features: Dict[str, Any]) -> None:
        """Atualiza o perfil do usuário com novos dados"""
        profile = self._get_user_profile(user_id)
        
        # Adicionar ponto ao histórico (limitando o tamanho)
        max_history = self.config.get("max_history_points", 100)
        profile["history"].append(features)
        if len(profile["history"]) > max_history:
            profile["history"] = profile["history"][-max_history:]
            
        # Atualizar timestamp
        profile["last_update"] = datetime.now()
        
        # Recalcular baseline para características numéricas
        baseline = {}
        for feature_name in self.baseline_features:
            feature_history = [
                h[feature_name] for h in profile["history"] 
                if feature_name in h and isinstance(h[feature_name], (int, float))
            ]
            
            if feature_history:
                baseline[feature_name] = {
                    "mean": np.mean(feature_history),
                    "std": np.std(feature_history),
                    "min": min(feature_history),
                    "max": max(feature_history)
                }
                
        profile["baseline"] = baseline
    
    def analyze(self, data: Dict[str, Any]) -> None:
        """Realiza análise comportamental dos dados"""
        if not self.context:
            logger.error("Contexto não disponível para análise")
            return
            
        # Extrair ID do usuário
        user_id = data.get("user_id")
        if not user_id:
            logger.warning("ID de usuário não encontrado nos dados")
            return
            
        # Extrair características
        features = self._extract_features(data)
        
        # Calcular score de anomalia
        anomaly_score = self._calculate_anomaly_score(user_id, features)
        
        # Adicionar insights ao contexto
        self.context.add_insight(
            self.agent_id, 
            "anomaly_score", 
            anomaly_score
        )
        
        # Adicionar fator de risco
        self.context.add_risk_factor("behavioral_anomaly", anomaly_score)
        
        # Detectar anomalias que excedem o limiar
        if anomaly_score > self.anomaly_threshold:
            severity = "high" if anomaly_score > 0.9 else "medium"
            self.context.add_fraud_indicator(
                indicator_type="behavioral_anomaly",
                severity=severity,
                description=f"Padrão comportamental anômalo detectado (score: {anomaly_score:.2f})",
                confidence=anomaly_score
            )
            
            # Adicionar detalhes específicos sobre as anomalias
            anomalous_features = []
            for feature_name, value in features.items():
                profile = self._get_user_profile(user_id)
                if feature_name in profile.get("baseline", {}):
                    baseline = profile["baseline"][feature_name]
                    if isinstance(value, (int, float)):
                        mean_val = baseline["mean"]
                        std_val = baseline["std"] or 1.0
                        z_score = abs(value - mean_val) / std_val
                        if z_score > 2.0:  # Mais de 2 desvios padrão
                            anomalous_features.append({
                                "feature": feature_name,
                                "current_value": value,
                                "typical_value": f"{mean_val:.2f} ± {std_val:.2f}",
                                "deviation": f"{z_score:.2f} σ"
                            })
            
            if anomalous_features:
                self.context.add_insight(
                    self.agent_id,
                    "anomalous_features",
                    anomalous_features
                )
        
        # Atualizar perfil do usuário com novos dados (aprendizado contínuo)
        self._update_user_profile(user_id, features)
    
    def train(self, training_data: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Treina o modelo com dados históricos"""
        trained_users = 0
        total_samples = 0
        
        for record in training_data:
            user_id = record.get("user_id")
            if not user_id:
                continue
                
            features = self._extract_features(record)
            self._update_user_profile(user_id, features)
            total_samples += 1
            
            # Contar usuários únicos treinados
            if user_id not in self.user_profiles:
                trained_users += 1
        
        return {
            "status": "success",
            "trained_users": trained_users,
            "total_samples": total_samples
        }


# Registrar o agente quando o módulo for importado
def register_agent(config: Dict[str, Any]) -> BehavioralDetectionAgent:
    """Cria e registra o agente comportamental"""
    agent_id = config.get("agent_id", "behavioral_default")
    agent = BehavioralDetectionAgent(agent_id, config)
    AgentRegistry().register_agent(agent)
    return agent