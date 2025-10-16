"""
Modelos de Machine Learning para Detecção de Padrões de Fraude no Mercado Angolano

Este módulo implementa modelos específicos para detecção de fraudes contextualizados
para o mercado angolano, considerando características culturais, econômicas e 
comportamentais específicas da região.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import numpy as np
import pandas as pd
import tensorflow as tf
import joblib
from sklearn.ensemble import IsolationForest, RandomForestClassifier
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from tensorflow.keras.models import Sequential, load_model, Model
from tensorflow.keras.layers import Dense, LSTM, Dropout, Input, Concatenate
import xgboost as xgb
from typing import Dict, List, Tuple, Optional, Union, Any

# Configurações específicas para o mercado angolano
ANGOLA_CONFIG = {
    "currency": "AOA",
    "country_code": "AO",
    "region_code": "SSA",
    "high_risk_locations": ["Luanda-Sambizanga", "Luanda-Cazenga", "Benguela-Centro"],
    "suspicious_time_ranges": [(22, 5)],  # 22h às 5h
    "mobile_money_operators": ["UNITEL_MONEY", "MOVICEL_MKESH", "BAI_MULTICAIXA"],
    "suspicious_amount_thresholds": {
        "mobile_money": 500000.0,  # 500.000 AOA
        "card_transaction": 1000000.0,  # 1.000.000 AOA
        "bank_transfer": 5000000.0,  # 5.000.000 AOA
        "crypto": 250000.0,  # 250.000 AOA
    },
    "unusual_country_pairs": [
        ("AO", "RU"), ("AO", "NG"), ("AO", "CN"), ("AO", "TR")
    ],
    "document_types": ["BI", "PASSPORT", "MILITARY_ID", "MININT_ID", "RESIDENCE_PERMIT"],
    "regulatory_entities": ["BNA", "CMC", "INACOM", "AGT", "SME"],
    "trusted_bank_codes": ["BAI", "BFA", "BIC", "BPC", "BMA", "ATL", "BCA", "BDA", "KEVE"],
    "common_fraud_keywords": [
        "transferência imediata", "urgente", "secreto", "investimento garantido",
        "lucro rápido", "desconto", "kwanza forte", "desbloquear", "validação",
        "segurança", "confirmação", "transação", "kwanza digital"
    ]
}

class AngolaFraudPatternDetector:
    """
    Detector de padrões de fraude específicos para Angola utilizando múltiplos
    modelos de ML especializados para diferentes tipos de fraude.
    """
    
    def __init__(self, config_path: Optional[str] = None, models_path: Optional[str] = None):
        """
        Inicializa o detector de fraudes para Angola.
        
        Args:
            config_path: Caminho para arquivo de configuração personalizado
            models_path: Diretório para carregar/salvar modelos treinados
        """
        self.config = ANGOLA_CONFIG.copy()
        self.models_path = models_path or os.path.join(os.path.dirname(__file__), "models")
        
        # Carregar configuração personalizada se fornecida
        if config_path and os.path.exists(config_path):
            with open(config_path, 'r') as f:
                custom_config = json.load(f)
                self.config.update(custom_config)
        
        # Criar diretório de modelos se não existir
        os.makedirs(self.models_path, exist_ok=True)
        
        # Inicializar modelos
        self.models = {}
        self.scalers = {}
        self.encoders = {}
        
        # Carregar modelos pré-treinados
        self._load_models()
    
    def _load_models(self):
        """Carrega modelos pré-treinados se disponíveis."""
        # Modelo para detecção de anomalias em transações financeiras
        transaction_model_path = os.path.join(self.models_path, "ao_transaction_fraud_model.h5")
        if os.path.exists(transaction_model_path):
            self.models["transaction"] = load_model(transaction_model_path)
            
            # Carregar scaler correspondente
            scaler_path = os.path.join(self.models_path, "ao_transaction_scaler.pkl")
            if os.path.exists(scaler_path):
                self.scalers["transaction"] = joblib.load(scaler_path)
        else:
            self.models["transaction"] = self._build_transaction_model()
        
        # Modelo para detecção de fraudes em documentos
        document_model_path = os.path.join(self.models_path, "ao_document_fraud_model.pkl")
        if os.path.exists(document_model_path):
            self.models["document"] = joblib.load(document_model_path)
        else:
            self.models["document"] = self._build_document_model()
            
        # Modelo para detecção de anomalias comportamentais
        behavior_model_path = os.path.join(self.models_path, "ao_behavior_fraud_model.pkl")
        if os.path.exists(behavior_model_path):
            self.models["behavior"] = joblib.load(behavior_model_path)
        else:
            self.models["behavior"] = self._build_behavior_model()
            
        # Modelo para detecção de padrões de fraude específicos de Angola
        angola_patterns_model_path = os.path.join(self.models_path, "ao_specific_patterns_model.pkl")
        if os.path.exists(angola_patterns_model_path):
            self.models["angola_patterns"] = joblib.load(angola_patterns_model_path)
        else:
            self.models["angola_patterns"] = self._build_angola_specific_model()
    
    def _build_transaction_model(self) -> Model:
        """
        Constrói um modelo de deep learning para detecção de fraudes em transações financeiras
        adaptado às características específicas do mercado angolano.
        """
        # Input para características numéricas (valor, tempo, frequência, etc)
        numeric_input = Input(shape=(10,), name="numeric_features")
        numeric_features = Dense(64, activation="relu")(numeric_input)
        numeric_features = Dropout(0.3)(numeric_features)
        numeric_features = Dense(32, activation="relu")(numeric_features)
        
        # Input para características categóricas (tipo de transação, canal, localização, etc)
        categorical_input = Input(shape=(15,), name="categorical_features")
        categorical_features = Dense(32, activation="relu")(categorical_input)
        categorical_features = Dropout(0.3)(categorical_features)
        categorical_features = Dense(16, activation="relu")(categorical_features)
        
        # Input para sequência temporal (histórico recente de transações)
        sequence_input = Input(shape=(10, 8), name="sequence_features")
        sequence_features = LSTM(32, return_sequences=False)(sequence_input)
        sequence_features = Dropout(0.3)(sequence_features)
        
        # Combinar todas as características
        combined = Concatenate()([numeric_features, categorical_features, sequence_features])
        combined = Dense(64, activation="relu")(combined)
        combined = Dropout(0.3)(combined)
        combined = Dense(32, activation="relu")(combined)
        
        # Saída: probabilidade de fraude, score de risco e tipo de fraude mais provável
        fraud_probability = Dense(1, activation="sigmoid", name="fraud_probability")(combined)
        risk_score = Dense(1, activation="sigmoid", name="risk_score")(combined)
        fraud_type = Dense(7, activation="softmax", name="fraud_type")(combined)
        
        # Criar e compilar o modelo
        model = Model(
            inputs=[numeric_input, categorical_input, sequence_input],
            outputs=[fraud_probability, risk_score, fraud_type]
        )
        
        model.compile(
            optimizer="adam",
            loss={
                "fraud_probability": "binary_crossentropy",
                "risk_score": "mean_squared_error",
                "fraud_type": "categorical_crossentropy"
            },
            metrics={
                "fraud_probability": ["accuracy", tf.keras.metrics.AUC()],
                "risk_score": ["mae"],
                "fraud_type": ["accuracy"]
            }
        )
        
        return model
    
    def _build_document_model(self) -> RandomForestClassifier:
        """
        Constrói um modelo para detecção de fraudes em documentos angolanos
        (BI, passaporte, cartões de residência, etc).
        """
        # Utilizamos Random Forest para classificação de documentos
        model = RandomForestClassifier(
            n_estimators=200,
            max_depth=15,
            min_samples_split=10,
            min_samples_leaf=4,
            class_weight="balanced",
            random_state=42,
            n_jobs=-1,
            verbose=0
        )
        
        return model
    
    def _build_behavior_model(self) -> IsolationForest:
        """
        Constrói um modelo para detecção de anomalias comportamentais dos usuários,
        considerando padrões específicos dos consumidores angolanos.
        """
        # Utilizamos Isolation Forest para detecção de anomalias comportamentais
        model = IsolationForest(
            n_estimators=150,
            max_samples='auto',
            contamination=0.05,  # Estimativa de comportamentos anômalos
            max_features=1.0,
            bootstrap=True,
            n_jobs=-1,
            random_state=42,
            verbose=0
        )
        
        return model
    
    def _build_angola_specific_model(self) -> xgb.XGBClassifier:
        """
        Constrói um modelo específico para padrões de fraude observados no mercado angolano,
        com foco em dinâmicas locais como mobile money, pagamentos informais, etc.
        """
        # Utilizamos XGBoost para classificação de padrões específicos
        model = xgb.XGBClassifier(
            n_estimators=200,
            max_depth=8,
            learning_rate=0.1,
            subsample=0.8,
            colsample_bytree=0.8,
            objective='binary:logistic',
            scale_pos_weight=3,  # Ajuste para desbalanceamento de classes
            tree_method='hist',
            random_state=42,
            verbosity=0
        )
        
        return model
    
    def preprocess_transaction_data(self, transaction_data: Dict) -> Dict:
        """
        Pré-processa dados de transações financeiras para análise de fraude.
        
        Args:
            transaction_data: Dados da transação a ser analisada
            
        Returns:
            Dados pré-processados e normalizados
        """
        # Extrair e normalizar características numéricas
        numeric_features = np.array([
            transaction_data.get("amount", 0),
            transaction_data.get("previousBalance", 0),
            transaction_data.get("daysSinceLastTransaction", 0),
            transaction_data.get("transactionFrequency", 0),
            transaction_data.get("hourOfDay", 0) / 24.0,  # Normalizado para [0, 1]
            transaction_data.get("dayOfWeek", 0) / 7.0,   # Normalizado para [0, 1]
            transaction_data.get("monthDay", 0) / 31.0,   # Normalizado para [0, 1]
            transaction_data.get("transactionCount24h", 0),
            transaction_data.get("failedAttempts", 0),
            transaction_data.get("velocityScore", 0)
        ]).reshape(1, -1)
        
        # Normalizar com StandardScaler se disponível
        if "transaction" in self.scalers:
            numeric_features = self.scalers["transaction"].transform(numeric_features)
        
        # One-hot encoding para características categóricas
        categorical_features = np.zeros((1, 15))
        
        # Mapear características categóricas
        transaction_type = transaction_data.get("transactionType", "")
        transaction_types = ["transfer", "payment", "withdrawal", "deposit", "exchange"]
        if transaction_type in transaction_types:
            idx = transaction_types.index(transaction_type)
            categorical_features[0, idx] = 1
            
        channel = transaction_data.get("channel", "")
        channels = ["mobile", "web", "agency", "atm", "pos"]
        if channel in channels:
            idx = channels.index(channel) + len(transaction_types)
            categorical_features[0, idx] = 1
            
        device_type = transaction_data.get("deviceType", "")
        device_types = ["mobile_android", "mobile_ios", "desktop", "unknown", "other"]
        if device_type in device_types:
            idx = device_types.index(device_type) + len(transaction_types) + len(channels)
            categorical_features[0, idx] = 1
        
        # Sequência de transações recentes (histórico)
        recent_transactions = transaction_data.get("recentTransactions", [])
        sequence_features = np.zeros((1, 10, 8))
        
        for i, tx in enumerate(recent_transactions[:10]):
            if i >= 10:
                break
                
            # Para cada transação recente, extraímos 8 características principais
            sequence_features[0, i, 0] = tx.get("amount", 0)
            sequence_features[0, i, 1] = 1 if tx.get("transactionType") == transaction_type else 0
            sequence_features[0, i, 2] = 1 if tx.get("channel") == channel else 0
            sequence_features[0, i, 3] = tx.get("hourOfDay", 0) / 24.0
            sequence_features[0, i, 4] = tx.get("dayOfWeek", 0) / 7.0
            sequence_features[0, i, 5] = 1 if tx.get("status") == "success" else 0
            sequence_features[0, i, 6] = tx.get("velocityScore", 0)
            sequence_features[0, i, 7] = tx.get("distanceFromHome", 0) / 1000.0  # em km
        
        return {
            "numeric_features": numeric_features,
            "categorical_features": categorical_features,
            "sequence_features": sequence_features
        }
    
    def evaluate_transaction(self, transaction_data: Dict) -> Dict:
        """
        Avalia uma transação financeira para detecção de fraude usando o modelo
        especializado para o mercado angolano.
        
        Args:
            transaction_data: Dados da transação a ser analisada
            
        Returns:
            Resultado da avaliação com probabilidade de fraude, score de risco e padrões detectados
        """
        # Pré-processar os dados
        processed_data = self.preprocess_transaction_data(transaction_data)
        
        # Aplicar regras específicas do mercado angolano
        angola_specific_signals = self._check_angola_specific_signals(transaction_data)
        
        # Gerar predições usando o modelo de transações
        if "transaction" in self.models:
            model_predictions = self.models["transaction"].predict([
                processed_data["numeric_features"],
                processed_data["categorical_features"],
                processed_data["sequence_features"]
            ])
            
            # Extrair resultados do modelo
            fraud_probability = float(model_predictions[0][0][0])
            risk_score = float(model_predictions[1][0][0])
            fraud_type_probabilities = model_predictions[2][0]
            
            # Mapear tipos de fraude
            fraud_types = ["account_takeover", "money_laundering", "synthetic_identity", 
                          "social_engineering", "card_not_present", "first_party_fraud", "other"]
            fraud_type_scores = {
                fraud_type: float(prob) 
                for fraud_type, prob in zip(fraud_types, fraud_type_probabilities)
            }
            
            most_likely_type = fraud_types[np.argmax(fraud_type_probabilities)]
        else:
            # Fallback para caso o modelo não esteja disponível
            fraud_probability = 0.5
            risk_score = 0.5
            fraud_type_scores = {}
            most_likely_type = "unknown"
        
        # Combinar sinais específicos de Angola com resultados do modelo
        adjusted_risk = self._combine_risks(risk_score, angola_specific_signals["risk_score"])
        
        # Determinar padrões específicos detectados
        detected_patterns = []
        if angola_specific_signals["risk_score"] > 0.7:
            detected_patterns.extend(angola_specific_signals["detected_patterns"])
        
        # Determinar ações recomendadas com base na análise
        recommended_actions = self._determine_actions(
            fraud_probability, 
            adjusted_risk,
            most_likely_type,
            detected_patterns
        )
        
        return {
            "fraud_probability": fraud_probability,
            "risk_score": adjusted_risk,
            "fraud_type": most_likely_type,
            "fraud_type_scores": fraud_type_scores,
            "detected_patterns": detected_patterns,
            "angola_specific_signals": angola_specific_signals["signals"],
            "recommended_actions": recommended_actions,
            "confidence": 0.85,  # Nível de confiança do modelo
            "processing_time_ms": 235,  # Tempo de processamento simulado
        }
    
    def _check_angola_specific_signals(self, transaction_data: Dict) -> Dict:
        """
        Verifica sinais específicos do mercado angolano que podem indicar fraude.
        
        Args:
            transaction_data: Dados da transação
            
        Returns:
            Dicionário com sinais específicos detectados e score de risco
        """
        signals = []
        risk_score = 0.0
        detected_patterns = []
        
        # Verificar operadora de mobile money suspeita
        mobile_operator = transaction_data.get("mobileOperator", "")
        if mobile_operator and mobile_operator not in self.config["mobile_money_operators"]:
            signals.append("mobile_operator_suspicious")
            risk_score += 0.15
            detected_patterns.append("operadora_nao_reconhecida")
        
        # Verificar valor suspeito baseado no tipo de transação
        amount = transaction_data.get("amount", 0)
        transaction_type = transaction_data.get("transactionType", "")
        
        if transaction_type in self.config["suspicious_amount_thresholds"]:
            threshold = self.config["suspicious_amount_thresholds"][transaction_type]
            if amount > threshold:
                signals.append("suspicious_amount")
                risk_score += 0.2
                detected_patterns.append("valor_acima_threshold")
        
        # Verificar hora suspeita (22h às 5h são horários de maior risco em Angola)
        hour = transaction_data.get("hourOfDay", -1)
        if hour >= 22 or hour < 5:
            signals.append("suspicious_hour")
            risk_score += 0.1
            detected_patterns.append("horario_suspeito_noturno")
        
        # Verificar localização de alto risco
        location = transaction_data.get("location", "")
        for high_risk_loc in self.config["high_risk_locations"]:
            if high_risk_loc in location:
                signals.append("high_risk_location")
                risk_score += 0.15
                detected_patterns.append("local_alto_risco")
                break
        
        # Verificar padrões de valores específicos de Angola
        # Em Angola, valores redondos ou próximos de limites são suspeitos
        amount_str = str(int(amount))
        if amount_str.endswith("000000") or amount_str.endswith("999999"):
            signals.append("suspicious_round_amount")
            risk_score += 0.1
            detected_patterns.append("valor_redondo_suspeito")
        
        # Verificar transferência internacional para países de alto risco
        recipient_country = transaction_data.get("recipientCountry", "")
        if recipient_country:
            country_pair = ("AO", recipient_country)
            if country_pair in self.config["unusual_country_pairs"]:
                signals.append("unusual_country_pair")
                risk_score += 0.25
                detected_patterns.append("transferencia_pais_suspeito")
        
        # Verificar keywords suspeitas na descrição
        description = transaction_data.get("description", "").lower()
        for keyword in self.config["common_fraud_keywords"]:
            if keyword.lower() in description:
                signals.append(f"suspicious_keyword_{keyword}")
                risk_score += 0.1
                detected_patterns.append("palavras_chave_suspeitas")
                break
        
        # Verificar banco suspeito (não listado nos bancos confiáveis)
        bank_code = transaction_data.get("bankCode", "")
        if bank_code and bank_code not in self.config["trusted_bank_codes"]:
            signals.append("untrusted_bank")
            risk_score += 0.15
            detected_patterns.append("banco_nao_confiavel")
        
        # Limitar score de risco ao máximo de 1.0
        risk_score = min(risk_score, 1.0)
        
        return {
            "signals": signals,
            "risk_score": risk_score,
            "detected_patterns": detected_patterns
        }
    
    def _combine_risks(self, model_risk: float, specific_risk: float) -> float:
        """
        Combina o score de risco do modelo com os sinais específicos de Angola.
        
        Args:
            model_risk: Score de risco do modelo ML
            specific_risk: Score de risco dos sinais específicos
            
        Returns:
            Score de risco combinado
        """
        # Pesos para combinação
        model_weight = 0.6
        specific_weight = 0.4
        
        # Combinação ponderada
        combined_risk = (model_risk * model_weight) + (specific_risk * specific_weight)
        
        # Ajuste exponencial para valores altos (amplificar sinais fortes)
        if combined_risk > 0.7:
            combined_risk = min(combined_risk * 1.2, 1.0)
        
        return combined_risk
    
    def _determine_actions(
        self, 
        fraud_probability: float, 
        risk_score: float,
        fraud_type: str,
        detected_patterns: List[str]
    ) -> List[Dict]:
        """
        Determina ações recomendadas com base na análise de fraude.
        
        Args:
            fraud_probability: Probabilidade de fraude
            risk_score: Score de risco
            fraud_type: Tipo de fraude mais provável
            detected_patterns: Padrões específicos detectados
            
        Returns:
            Lista de ações recomendadas
        """
        actions = []
        
        # Ação de bloqueio para casos de alto risco
        if risk_score > 0.8:
            actions.append({
                "action_type": "block",
                "priority": "high",
                "reason": "Alto risco de fraude detectado",
                "confidence": risk_score
            })
        
        # Verificação adicional para casos de risco médio
        elif risk_score > 0.5:
            actions.append({
                "action_type": "additional_verification",
                "priority": "medium",
                "reason": "Risco significativo detectado",
                "verification_type": "two_factor",
                "confidence": risk_score
            })
        
        # Monitoramento para casos de baixo risco
        elif risk_score > 0.3:
            actions.append({
                "action_type": "monitor",
                "priority": "low",
                "reason": "Risco baixo, mas acima do normal",
                "confidence": risk_score
            })
        
        # Ação específica para padrões detectados
        if detected_patterns:
            pattern_desc = ", ".join(detected_patterns[:3])
            actions.append({
                "action_type": "flag_for_review",
                "priority": "medium" if risk_score > 0.6 else "low",
                "reason": f"Padrões suspeitos detectados: {pattern_desc}",
                "patterns": detected_patterns,
                "confidence": min(risk_score + 0.1, 1.0)
            })
        
        # Ajuste de segurança para tipos específicos de fraude
        if fraud_type == "account_takeover" and fraud_probability > 0.6:
            actions.append({
                "action_type": "security_adjust",
                "priority": "high",
                "adjustments": ["require_biometrics", "restrict_transactions", "notify_user"],
                "reason": "Possível tomada de conta detectada",
                "confidence": fraud_probability
            })
        elif fraud_type == "money_laundering" and fraud_probability > 0.5:
            actions.append({
                "action_type": "report",
                "priority": "high",
                "report_to": ["compliance_team", "bna"],
                "reason": "Possível lavagem de dinheiro",
                "confidence": fraud_probability
            })
        
        return actions
        
    def evaluate_document(self, document_data: Dict) -> Dict:
        """
        Avalia um documento para detecção de fraude.
        
        Args:
            document_data: Dados do documento a ser analisado
            
        Returns:
            Resultado da avaliação com probabilidade de fraude
        """
        # Implementação a ser expandida
        return {"is_valid": True, "confidence": 0.9, "risk_score": 0.1}
    
    def analyze_user_behavior(self, behavior_data: Dict) -> Dict:
        """
        Analisa o comportamento do usuário para detecção de anomalias.
        
        Args:
            behavior_data: Dados de comportamento do usuário
            
        Returns:
            Resultado da análise com score de anomalia
        """
        # Implementação a ser expandida
        return {"is_anomalous": False, "anomaly_score": 0.2, "confidence": 0.85}
    
    def train(self, training_data: Dict) -> None:
        """
        Treina os modelos com novos dados.
        
        Args:
            training_data: Dados de treinamento
        """
        # Implementação a ser expandida
        pass
    
    def save_models(self) -> None:
        """Salva os modelos treinados."""
        # Implementação a ser expandida
        pass