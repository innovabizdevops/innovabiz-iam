"""
Agente ML - Implementa detecção de fraude usando modelos de Machine Learning
"""
from typing import Dict, Any, List, Optional, Tuple
import logging
import numpy as np
import pickle
import os
from datetime import datetime
from .base_agent import FraudDetectionAgent, AgentRegistry

try:
    from sklearn.ensemble import IsolationForest, RandomForestClassifier
    from sklearn.preprocessing import StandardScaler
    SKLEARN_AVAILABLE = True
except ImportError:
    SKLEARN_AVAILABLE = False
    logging.warning("scikit-learn não está disponível. MLAgent terá funcionalidade limitada.")

logger = logging.getLogger(__name__)

class MLAgent(FraudDetectionAgent):
    """Agente que utiliza modelos de ML para detecção de fraudes"""
    
    def __init__(self, agent_id: str, config: Dict[str, Any]):
        super().__init__(agent_id, config)
        self.model_type = config.get("model_type", "isolation_forest")
        self.model_path = config.get("model_path")
        self.feature_mapping = config.get("feature_mapping", {})
        self.required_features = config.get("required_features", [])
        self.scaler = None
        self.model = None
        
    def _initialize(self) -> None:
        """Inicialização específica para agente ML"""
        logger.info(f"Inicializando agente ML: {self.agent_id}")
        if not SKLEARN_AVAILABLE:
            logger.error("scikit-learn não está instalado. Agente ML desabilitado.")
            self.enabled = False
            return
            
        # Carregar modelo se existir
        if self.model_path and os.path.exists(self.model_path):
            try:
                with open(self.model_path, 'rb') as f:
                    saved_data = pickle.load(f)
                    self.model = saved_data.get('model')
                    self.scaler = saved_data.get('scaler')
                    logger.info(f"Modelo carregado de {self.model_path}")
            except Exception as e:
                logger.error(f"Erro ao carregar modelo de {self.model_path}: {e}")
                self._initialize_model()
        else:
            self._initialize_model()
            
    def _initialize_model(self) -> None:
        """Inicializa um novo modelo ML"""
        if not SKLEARN_AVAILABLE:
            return
            
        if self.model_type == "isolation_forest":
            self.model = IsolationForest(
                contamination=0.05,
                random_state=42,
                n_estimators=100
            )
            logger.info("Novo modelo Isolation Forest inicializado")
        elif self.model_type == "random_forest":
            self.model = RandomForestClassifier(
                n_estimators=100,
                random_state=42,
                class_weight='balanced'
            )
            logger.info("Novo modelo Random Forest inicializado")
        else:
            logger.error(f"Tipo de modelo desconhecido: {self.model_type}")
            
        self.scaler = StandardScaler()
        
    def get_agent_type(self) -> str:
        """Retorna o tipo do agente"""
        return "machine_learning"
        
    def supports_continuous_learning(self) -> bool:
        """Este agente suporta aprendizado contínuo"""
        return True
        
    def _extract_features(self, data: Dict[str, Any]) -> Optional[np.ndarray]:
        """Extrai características para o modelo ML a partir dos dados"""
        # Usar mapeamento de características ou extrair diretamente
        features = {}
        
        if self.feature_mapping:
            # Extrair baseado no mapeamento configurado
            for feature_name, data_path in self.feature_mapping.items():
                value = self._get_nested_value(data, data_path)
                if value is not None:
                    features[feature_name] = value
        else:
            # Extração padrão baseada em campos esperados
            for feature in self.required_features:
                value = self._get_nested_value(data, feature)
                if value is not None:
                    features[feature] = value
        
        # Verificar se temos todas as características necessárias
        missing_features = [f for f in self.required_features if f not in features]
        if missing_features:
            logger.warning(f"Características ausentes: {missing_features}")
            return None
            
        # Converter para valores numéricos
        feature_values = []
        feature_names = []
        
        for name, value in features.items():
            # Converter tipos não numéricos
            if isinstance(value, bool):
                feature_values.append(1 if value else 0)
            elif isinstance(value, (int, float)):
                feature_values.append(float(value))
            elif isinstance(value, str):
                # Hash simples para valor categórico
                feature_values.append(hash(value) % 1000 / 1000)
            else:
                continue  # Ignorar tipos não suportados
                
            feature_names.append(name)
            
        if not feature_values:
            return None
            
        return np.array(feature_values).reshape(1, -1), feature_names
        
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
        
    def analyze(self, data: Dict[str, Any]) -> None:
        """Analisa os dados usando o modelo ML"""
        if not SKLEARN_AVAILABLE or not self.model or not self.context:
            logger.error("Modelo ML não disponível para análise")
            return
            
        # Extrair características
        features_result = self._extract_features(data)
        if not features_result:
            logger.warning("Não foi possível extrair características para análise ML")
            return
            
        features, feature_names = features_result
            
        # Normalizar dados se o scaler estiver disponível
        if self.scaler:
            try:
                features = self.scaler.transform(features)
            except Exception as e:
                logger.error(f"Erro ao normalizar características: {e}")
                # Continuar com dados não normalizados
                
        # Analisar com o modelo
        try:
            if self.model_type == "isolation_forest":
                # -1 para anomalias, 1 para normais
                raw_score = self.model.decision_function(features)[0]
                # Converter para pontuação de 0 a 1 (quanto maior, mais anômalo)
                anomaly_score = 1 - (raw_score + 0.5) / 1.5
                anomaly_score = max(0, min(1, anomaly_score))
                
                self.context.add_insight(
                    self.agent_id,
                    "anomaly_score",
                    anomaly_score
                )
                
                # Adicionar fator de risco
                self.context.add_risk_factor("ml_anomaly", anomaly_score)
                
                # Adicionar indicador de fraude para anomalias significativas
                if anomaly_score > 0.8:
                    self.context.add_fraud_indicator(
                        indicator_type="ml_anomaly",
                        severity="high",
                        description=f"Anomalia detectada por ML (score: {anomaly_score:.2f})",
                        confidence=anomaly_score
                    )
                elif anomaly_score > 0.6:
                    self.context.add_fraud_indicator(
                        indicator_type="ml_anomaly",
                        severity="medium",
                        description=f"Possível anomalia detectada por ML (score: {anomaly_score:.2f})",
                        confidence=anomaly_score
                    )
            
            elif self.model_type == "random_forest":
                # Probabilidade de classe positiva (fraude)
                try:
                    fraud_probability = self.model.predict_proba(features)[0][1]
                    
                    self.context.add_insight(
                        self.agent_id,
                        "fraud_probability",
                        fraud_probability
                    )
                    
                    # Adicionar fator de risco
                    self.context.add_risk_factor("ml_fraud", fraud_probability)
                    
                    # Adicionar indicador de fraude baseado na probabilidade
                    if fraud_probability > 0.8:
                        self.context.add_fraud_indicator(
                            indicator_type="ml_fraud_detection",
                            severity="high",
                            description=f"Alta probabilidade de fraude detectada (score: {fraud_probability:.2f})",
                            confidence=fraud_probability
                        )
                    elif fraud_probability > 0.6:
                        self.context.add_fraud_indicator(
                            indicator_type="ml_fraud_detection",
                            severity="medium",
                            description=f"Média probabilidade de fraude detectada (score: {fraud_probability:.2f})",
                            confidence=fraud_probability
                        )
                except:
                    # Modelo pode não estar treinado para predict_proba
                    prediction = self.model.predict(features)[0]
                    if prediction == 1:  # Classe positiva (fraude)
                        self.context.add_fraud_indicator(
                            indicator_type="ml_fraud_detection",
                            severity="high",
                            description="Fraude detectada pelo modelo ML",
                            confidence=0.85
                        )
                        self.context.add_risk_factor("ml_fraud", 0.85)
            
        except Exception as e:
            logger.error(f"Erro ao analisar com modelo ML: {e}")
    
    def train(self, training_data: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Treina o modelo com dados históricos"""
        if not SKLEARN_AVAILABLE:
            return {"status": "error", "message": "scikit-learn não disponível"}
            
        if not training_data:
            return {"status": "error", "message": "Dados de treinamento vazios"}
            
        X = []
        y = []
        all_feature_names = set()
        
        # Extrair características e rótulos de todos os registros
        for record in training_data:
            features_result = self._extract_features(record)
            if not features_result:
                continue
                
            features, feature_names = features_result
            X.append(features.flatten())
            all_feature_names.update(feature_names)
            
            # Para classificadores supervisionados, precisamos de rótulos
            if self.model_type == "random_forest":
                label = record.get("is_fraud", record.get("fraud", 0))
                y.append(1 if label else 0)
        
        if not X:
            return {"status": "error", "message": "Nenhuma característica extraída"}
            
        X = np.array(X)
        
        try:
            # Treinar o scaler
            self.scaler = StandardScaler()
            X_scaled = self.scaler.fit_transform(X)
            
            # Treinar o modelo apropriado
            if self.model_type == "isolation_forest":
                self.model = IsolationForest(
                    contamination=0.05,
                    random_state=42,
                    n_estimators=100
                )
                self.model.fit(X_scaled)
            elif self.model_type == "random_forest":
                if not y:
                    return {"status": "error", "message": "Dados de treinamento sem rótulos"}
                    
                self.model = RandomForestClassifier(
                    n_estimators=100,
                    random_state=42,
                    class_weight='balanced'
                )
                self.model.fit(X_scaled, y)
                
            # Salvar o modelo treinado se caminho foi especificado
            if self.model_path:
                os.makedirs(os.path.dirname(self.model_path), exist_ok=True)
                with open(self.model_path, 'wb') as f:
                    pickle.dump({
                        'model': self.model,
                        'scaler': self.scaler,
                        'trained_at': datetime.now().isoformat(),
                        'feature_names': list(all_feature_names)
                    }, f)
                    
            return {
                "status": "success",
                "samples": len(X),
                "features": len(all_feature_names),
                "model_type": self.model_type
            }
        
        except Exception as e:
            logger.error(f"Erro ao treinar modelo ML: {e}")
            return {"status": "error", "message": str(e)}


# Registrar o agente quando o módulo for importado
def register_agent(config: Dict[str, Any]) -> MLAgent:
    """Cria e registra o agente ML"""
    agent_id = config.get("agent_id", "ml_default")
    agent = MLAgent(agent_id, config)
    AgentRegistry().register_agent(agent)
    return agent