#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Módulo de Detecção de Anomalias baseado em ML para Análise Comportamental

Este módulo implementa algoritmos de Machine Learning para detecção avançada
de anomalias comportamentais, complementando as regras específicas por região.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import time
import uuid
import pickle
import logging
import datetime
import numpy as np
import pandas as pd
from enum import Enum
from typing import Dict, Any, List, Optional, Tuple, Union, Set
from dataclasses import dataclass, field

# Bibliotecas ML (precisam ser instaladas)
from sklearn.ensemble import IsolationForest, RandomForestClassifier
from sklearn.cluster import DBSCAN
from sklearn.preprocessing import StandardScaler, MinMaxScaler
from sklearn.pipeline import Pipeline
from sklearn.decomposition import PCA
from sklearn.model_selection import train_test_split
from sklearn.metrics import precision_score, recall_score, f1_score

# Configuração do logger
logger = logging.getLogger("iam.trustguard.ml_models.anomaly_detection")


class ModelType(Enum):
    """Tipos de modelos de ML para detecção de anomalias."""
    ISOLATION_FOREST = "isolation_forest"
    DBSCAN = "dbscan"
    AUTO_ENCODER = "auto_encoder"
    HYBRID = "hybrid"
    SEQUENTIAL = "sequential"


class FeatureCategory(Enum):
    """Categorias de características para modelos de ML."""
    AUTHENTICATION = "authentication"
    TRANSACTION = "transaction"
    SESSION = "session"
    DEVICE = "device"
    LOCATION = "location"
    TEMPORAL = "temporal"
    NETWORK = "network"
    USER_PROFILE = "user_profile"


@dataclass
class ModelMetadata:
    """Metadados sobre um modelo de ML treinado."""
    model_id: str
    model_type: ModelType
    version: str
    created_at: datetime.datetime
    updated_at: datetime.datetime
    training_data_size: int
    feature_count: int
    performance_metrics: Dict[str, float]
    feature_importance: Optional[Dict[str, float]] = None
    regions: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte metadados para dicionário."""
        return {
            "model_id": self.model_id,
            "model_type": self.model_type.value,
            "version": self.version,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
            "training_data_size": self.training_data_size,
            "feature_count": self.feature_count,
            "performance_metrics": self.performance_metrics,
            "feature_importance": self.feature_importance,
            "regions": self.regions
        }


class FeatureExtractor:
    """
    Extrai e processa características de eventos para uso em modelos de ML.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o extrator de características.
        
        Args:
            config: Configuração para extração de características
        """
        self.config = config
        self.feature_categories = config.get("feature_categories", [])
        self.feature_transformations = config.get("feature_transformations", {})
        self.categorical_encodings = config.get("categorical_encodings", {})
        
        # Inicializar transformadores
        self.scaler = StandardScaler()
        self.has_fit_transformers = False
        
        logger.info("Extrator de características inicializado")
    
    def extract_features(self, event_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extrai características de um evento para alimentar modelos ML.
        
        Args:
            event_data: Dados do evento (autenticação, transação, etc.)
            
        Returns:
            Características extraídas
        """
        features = {}
        
        try:
            # Extrair características de autenticação
            if FeatureCategory.AUTHENTICATION.value in self.feature_categories:
                auth_features = self._extract_authentication_features(event_data)
                features.update(auth_features)
            
            # Extrair características de transação
            if FeatureCategory.TRANSACTION.value in self.feature_categories:
                transaction_features = self._extract_transaction_features(event_data)
                features.update(transaction_features)
            
            # Extrair características de sessão
            if FeatureCategory.SESSION.value in self.feature_categories:
                session_features = self._extract_session_features(event_data)
                features.update(session_features)
            
            # Extrair características de dispositivo
            if FeatureCategory.DEVICE.value in self.feature_categories:
                device_features = self._extract_device_features(event_data)
                features.update(device_features)
            
            # Extrair características de localização
            if FeatureCategory.LOCATION.value in self.feature_categories:
                location_features = self._extract_location_features(event_data)
                features.update(location_features)
            
            # Extrair características temporais
            if FeatureCategory.TEMPORAL.value in self.feature_categories:
                temporal_features = self._extract_temporal_features(event_data)
                features.update(temporal_features)
            
            # Extrair características de rede
            if FeatureCategory.NETWORK.value in self.feature_categories:
                network_features = self._extract_network_features(event_data)
                features.update(network_features)
            
            # Extrair características de perfil de usuário
            if FeatureCategory.USER_PROFILE.value in self.feature_categories:
                profile_features = self._extract_user_profile_features(event_data)
                features.update(profile_features)
            
            # Aplicar transformações personalizadas
            features = self._apply_transformations(features)
            
            logger.debug(f"Extraídas {len(features)} características do evento")
            return features
            
        except Exception as e:
            logger.error(f"Erro ao extrair características: {str(e)}")
            # Retorna conjunto mínimo de características em caso de erro
            return {
                "event_type": event_data.get("event_type", "unknown"),
                "timestamp": datetime.datetime.now().timestamp(),
                "error_in_extraction": 1.0
            }
    
    def _extract_authentication_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas à autenticação.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de autenticação
        """
        auth_data = event_data.get("authentication", {})
        features = {}
        
        # Tipo de autenticação (converter para numérico)
        auth_type = auth_data.get("auth_type", "unknown")
        auth_type_mapping = self.categorical_encodings.get("auth_type", {})
        features["auth_type_code"] = float(auth_type_mapping.get(auth_type, 0.0))
        
        # Nível de autenticação (1FA, 2FA, MFA)
        features["auth_level"] = float(auth_data.get("auth_level", 1.0))
        
        # Tempo de autenticação em segundos
        features["auth_duration_sec"] = float(auth_data.get("duration_ms", 0.0) / 1000.0)
        
        # Número de tentativas
        features["auth_attempts"] = float(auth_data.get("attempts", 1.0))
        
        # Resultado (sucesso=1, falha=0)
        features["auth_success"] = 1.0 if auth_data.get("success", False) else 0.0
        
        # Fator de confiança da autenticação
        features["auth_confidence"] = float(auth_data.get("confidence", 1.0))
        
        # Hora do dia normalizada (0-1)
        auth_time = auth_data.get("timestamp", time.time())
        hour_of_day = datetime.datetime.fromtimestamp(auth_time).hour
        features["auth_hour_normalized"] = hour_of_day / 24.0
        
        return features
    
    def _extract_transaction_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas a transações.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de transação
        """
        tx_data = event_data.get("transaction", {})
        features = {}
        
        if not tx_data:
            # Retorna features vazias se não for evento de transação
            return features
        
        # Valor da transação (normalizado por config)
        tx_amount = float(tx_data.get("amount", 0.0))
        max_amount = self.config.get("normalization", {}).get("max_transaction_amount", 10000.0)
        features["tx_amount_normalized"] = min(tx_amount / max_amount, 1.0)
        
        # Tipo de transação
        tx_type = tx_data.get("type", "unknown")
        tx_type_mapping = self.categorical_encodings.get("transaction_type", {})
        features["tx_type_code"] = float(tx_type_mapping.get(tx_type, 0.0))
        
        # Velocidade de transação (valor/tempo desde última transação)
        last_tx_time = float(tx_data.get("last_transaction_time", 0.0))
        current_time = float(tx_data.get("timestamp", time.time()))
        time_diff = max(current_time - last_tx_time, 1.0)  # Evitar divisão por zero
        features["tx_velocity"] = tx_amount / time_diff
        
        # Número de transações recentes
        features["tx_count_24h"] = float(tx_data.get("tx_count_24h", 0.0))
        
        # Distância em relação ao padrão de gasto
        features["tx_distance_from_pattern"] = float(tx_data.get("distance_from_pattern", 0.0))
        
        # Novo beneficiário
        features["tx_new_beneficiary"] = 1.0 if tx_data.get("is_new_beneficiary", False) else 0.0
        
        return features
    
    def _extract_session_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas à sessão.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de sessão
        """
        session_data = event_data.get("session", {})
        features = {}
        
        # Duração da sessão
        session_start = float(session_data.get("start_time", 0.0))
        current_time = float(event_data.get("timestamp", time.time()))
        session_duration = current_time - session_start
        features["session_duration_sec"] = session_duration
        
        # Número de ações na sessão
        features["session_action_count"] = float(session_data.get("action_count", 0.0))
        
        # Taxa de ações por minuto
        action_rate = 0.0
        if session_duration > 0:
            action_rate = features["session_action_count"] / (session_duration / 60.0)
        features["session_action_rate"] = action_rate
        
        # Navegação entre páginas
        features["session_page_count"] = float(session_data.get("page_count", 0.0))
        
        # Uso de recursos sensíveis
        features["session_sensitive_resource_access"] = float(session_data.get("sensitive_resource_access", 0.0))
        
        return features
    
    def _extract_device_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas ao dispositivo.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de dispositivo
        """
        device_data = event_data.get("device", {})
        features = {}
        
        # Tipo de dispositivo
        device_type = device_data.get("type", "unknown")
        device_mapping = self.categorical_encodings.get("device_type", {})
        features["device_type_code"] = float(device_mapping.get(device_type, 0.0))
        
        # Sistema operacional
        os_type = device_data.get("os", "unknown")
        os_mapping = self.categorical_encodings.get("os_type", {})
        features["device_os_code"] = float(os_mapping.get(os_type, 0.0))
        
        # Versão do navegador
        browser_type = device_data.get("browser", "unknown")
        browser_mapping = self.categorical_encodings.get("browser_type", {})
        features["device_browser_code"] = float(browser_mapping.get(browser_type, 0.0))
        
        # Dispositivo rooteado/jailbreak
        features["device_is_rooted"] = 1.0 if device_data.get("is_rooted", False) else 0.0
        
        # Emulador
        features["device_is_emulator"] = 1.0 if device_data.get("is_emulator", False) else 0.0
        
        # ID de dispositivo alterado recentemente
        features["device_id_changed"] = 1.0 if device_data.get("id_changed_recently", False) else 0.0
        
        # Novo dispositivo para o usuário
        features["device_is_new"] = 1.0 if device_data.get("is_new", False) else 0.0
        
        # Número de dispositivos usados pelo usuário
        features["device_count_for_user"] = float(device_data.get("device_count_for_user", 1.0))
        
        return features
    
    def _extract_location_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas à localização.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de localização
        """
        location_data = event_data.get("location", {})
        features = {}
        
        # Distância da última localização conhecida (km)
        features["location_distance_from_last_km"] = float(location_data.get("distance_from_last_km", 0.0))
        
        # Velocidade de mudança de localização (km/h)
        distance_km = float(location_data.get("distance_from_last_km", 0.0))
        time_diff_hours = float(location_data.get("time_diff_from_last_h", 1.0))
        time_diff_hours = max(time_diff_hours, 0.001)  # Evitar divisão por zero
        features["location_speed_kmh"] = distance_km / time_diff_hours
        
        # Risco da região
        features["location_risk_score"] = float(location_data.get("risk_score", 0.0))
        
        # Localização inconsistente com o IP
        features["location_ip_mismatch"] = 1.0 if location_data.get("ip_location_mismatch", False) else 0.0
        
        # Tipo de localização (residencial, comercial, etc.)
        location_type = location_data.get("type", "unknown")
        location_mapping = self.categorical_encodings.get("location_type", {})
        features["location_type_code"] = float(location_mapping.get(location_type, 0.0))
        
        # VPN/Proxy detectado
        features["location_is_vpn"] = 1.0 if location_data.get("is_vpn", False) else 0.0
        
        # Localização na lista de alto risco
        features["location_is_high_risk"] = 1.0 if location_data.get("is_high_risk", False) else 0.0
        
        return features
    
    def _extract_temporal_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características temporais do evento.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características temporais
        """
        timestamp = event_data.get("timestamp", time.time())
        dt = datetime.datetime.fromtimestamp(timestamp)
        features = {}
        
        # Hora do dia (normalizada 0-1)
        hour = dt.hour + (dt.minute / 60.0)
        features["temporal_hour_normalized"] = hour / 24.0
        
        # Dia da semana (normalizado 0-1)
        features["temporal_day_of_week"] = float(dt.weekday()) / 6.0
        
        # Fim de semana
        features["temporal_is_weekend"] = 1.0 if dt.weekday() >= 5 else 0.0
        
        # Horário comercial
        is_business_hours = 9 <= dt.hour < 18 and dt.weekday() < 5
        features["temporal_is_business_hours"] = 1.0 if is_business_hours else 0.0
        
        # Madrugada
        is_late_night = 0 <= dt.hour < 5
        features["temporal_is_late_night"] = 1.0 if is_late_night else 0.0
        
        # Dia do mês (normalizado)
        features["temporal_day_of_month"] = float(dt.day) / 31.0
        
        return features
    
    def _extract_network_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas à rede.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de rede
        """
        network_data = event_data.get("network", {})
        features = {}
        
        # Tipo de conexão
        connection_type = network_data.get("connection_type", "unknown")
        connection_mapping = self.categorical_encodings.get("connection_type", {})
        features["network_type_code"] = float(connection_mapping.get(connection_type, 0.0))
        
        # Operadora
        carrier = network_data.get("carrier", "unknown")
        carrier_mapping = self.categorical_encodings.get("carrier", {})
        features["network_carrier_code"] = float(carrier_mapping.get(carrier, 0.0))
        
        # IP em lista de risco
        features["network_ip_risk"] = float(network_data.get("ip_risk_score", 0.0))
        
        # Múltiplos usuários no mesmo IP
        features["network_users_same_ip"] = float(network_data.get("user_count_same_ip", 1.0))
        
        # ASN em lista de risco
        features["network_asn_risk"] = float(network_data.get("asn_risk_score", 0.0))
        
        # Mudança de IP
        features["network_ip_changed"] = 1.0 if network_data.get("ip_changed", False) else 0.0
        
        return features
    
    def _extract_user_profile_features(self, event_data: Dict[str, Any]) -> Dict[str, float]:
        """
        Extrai características relacionadas ao perfil do usuário.
        
        Args:
            event_data: Dados do evento
            
        Returns:
            Características de perfil
        """
        profile_data = event_data.get("user_profile", {})
        features = {}
        
        # Idade da conta (em dias)
        account_age_days = float(profile_data.get("account_age_days", 0.0))
        max_age = 365 * 10  # 10 anos como máximo para normalização
        features["profile_account_age_normalized"] = min(account_age_days / max_age, 1.0)
        
        # Nível de acesso
        features["profile_access_level"] = float(profile_data.get("access_level", 1.0))
        
        # Score de risco histórico
        features["profile_historical_risk"] = float(profile_data.get("historical_risk_score", 0.0))
        
        # Histórico de alertas
        features["profile_alert_count_30d"] = float(profile_data.get("alert_count_30d", 0.0))
        
        # Segmento de cliente
        segment = profile_data.get("segment", "unknown")
        segment_mapping = self.categorical_encodings.get("customer_segment", {})
        features["profile_segment_code"] = float(segment_mapping.get(segment, 0.0))
        
        # Número de produtos/serviços contratados
        features["profile_product_count"] = float(profile_data.get("product_count", 0.0))
        
        # Score de crédito (normalizado)
        credit_score = float(profile_data.get("credit_score", 0.0))
        max_score = self.config.get("normalization", {}).get("max_credit_score", 1000.0)
        features["profile_credit_score_normalized"] = credit_score / max_score
        
        return features
    
    def _apply_transformations(self, features: Dict[str, float]) -> Dict[str, float]:
        """
        Aplica transformações específicas aos recursos extraídos.
        
        Args:
            features: Características extraídas
            
        Returns:
            Características transformadas
        """
        transformed = features.copy()
        
        # Aplicar transformações configuradas
        for feature_name, transformation in self.feature_transformations.items():
            if feature_name in transformed:
                value = transformed[feature_name]
                
                if transformation == "log":
                    if value > 0:
                        transformed[feature_name] = np.log(1 + value)
                elif transformation == "sqrt":
                    if value >= 0:
                        transformed[feature_name] = np.sqrt(value)
                elif transformation == "square":
                    transformed[feature_name] = value * value
                elif transformation == "inverse":
                    if value != 0:
                        transformed[feature_name] = 1.0 / value
        
        return transformed
    
    def transform_to_vector(self, features: Dict[str, float]) -> np.ndarray:
        """
        Converte o dicionário de características em um vetor para alimentar o modelo.
        
        Args:
            features: Características extraídas
            
        Returns:
            Vetor de características
        """
        # Criar array NumPy a partir do dicionário
        # Ordenar por chave para garantir consistência
        sorted_features = sorted(features.items(), key=lambda x: x[0])
        feature_vector = np.array([value for _, value in sorted_features])
        
        return feature_vector


class AnomalyDetector:
    """
    Detector de anomalias comportamentais baseado em ML.
    
    Utiliza múltiplos algoritmos para detecção de comportamento anômalo,
    complementando as regras específicas por região.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o detector de anomalias.
        
        Args:
            config: Configuração para o detector
        """
        self.config = config
        self.model_type = ModelType(config.get("model_type", ModelType.ISOLATION_FOREST.value))
        self.models = {}
        self.model_metadata = {}
        self.feature_extractor = FeatureExtractor(config.get("feature_extractor", {}))
        self.models_path = config.get("models_path", "./models")
        self.threshold = config.get("anomaly_threshold", 0.8)
        self.default_region = config.get("default_region", "global")
        
        # Criar diretório de modelos se não existir
        if not os.path.exists(self.models_path):
            os.makedirs(self.models_path)
        
        # Carregar modelos existentes
        self._load_models()
        
        logger.info(f"Detector de anomalias inicializado com modelo {self.model_type.value}")
    
    def _load_models(self) -> None:
        """Carrega modelos salvos do disco."""
        try:
            # Buscar todos os arquivos .pkl no diretório de modelos
            for filename in os.listdir(self.models_path):
                if filename.endswith(".pkl") and "_model_" in filename:
                    model_path = os.path.join(self.models_path, filename)
                    # Extrair região do nome do arquivo
                    region = filename.split("_model_")[0]
                    
                    with open(model_path, "rb") as f:
                        self.models[region] = pickle.load(f)
                    
                    # Carregar metadados
                    metadata_path = os.path.join(self.models_path, f"{region}_metadata.json")
                    if os.path.exists(metadata_path):
                        with open(metadata_path, "r") as f:
                            metadata_dict = json.load(f)
                            metadata = ModelMetadata(
                                model_id=metadata_dict["model_id"],
                                model_type=ModelType(metadata_dict["model_type"]),
                                version=metadata_dict["version"],
                                created_at=datetime.datetime.fromisoformat(metadata_dict["created_at"]),
                                updated_at=datetime.datetime.fromisoformat(metadata_dict["updated_at"]),
                                training_data_size=metadata_dict["training_data_size"],
                                feature_count=metadata_dict["feature_count"],
                                performance_metrics=metadata_dict["performance_metrics"],
                                feature_importance=metadata_dict.get("feature_importance"),
                                regions=[region]
                            )
                            self.model_metadata[region] = metadata
                    
                    logger.info(f"Modelo carregado para região: {region}")
            
            logger.info(f"Total de {len(self.models)} modelos carregados")
        except Exception as e:
            logger.error(f"Erro ao carregar modelos: {str(e)}")
    
    def _save_model(self, model: Any, region: str, metadata: ModelMetadata) -> None:
        """
        Salva um modelo treinado no disco.
        
        Args:
            model: Modelo treinado
            region: Código da região
            metadata: Metadados do modelo
        """
        try:
            # Salvar modelo
            model_path = os.path.join(self.models_path, f"{region}_model_{metadata.version}.pkl")
            with open(model_path, "wb") as f:
                pickle.dump(model, f)
            
            # Salvar metadados
            metadata_path = os.path.join(self.models_path, f"{region}_metadata.json")
            with open(metadata_path, "w") as f:
                json.dump(metadata.to_dict(), f, indent=2)
            
            logger.info(f"Modelo para região {region} salvo com sucesso: {model_path}")
        except Exception as e:
            logger.error(f"Erro ao salvar modelo para região {region}: {str(e)}")
    
    def train(self, training_data: List[Dict[str, Any]], region: str) -> ModelMetadata:
        """
        Treina um novo modelo para a região especificada.
        
        Args:
            training_data: Dados de treinamento
            region: Código da região
            
        Returns:
            Metadados do modelo treinado
        """
        if len(training_data) < 100:
            logger.warning(f"Conjunto de dados de treinamento muito pequeno para região {region}")
        
        # Extrair características
        logger.info(f"Extraindo características de {len(training_data)} eventos para treinamento")
        features_list = []
        for event in training_data:
            features = self.feature_extractor.extract_features(event)
            features_list.append(features)
        
        # Converter para DataFrame
        df = pd.DataFrame(features_list)
        
        # Remover colunas com valores nulos
        df = df.dropna(axis=1)
        
        # Verificar se há dados suficientes após limpeza
        if df.empty or len(df) < 50:
            logger.error("Dados insuficientes após limpeza")
            raise ValueError("Dados de treinamento insuficientes após limpeza")
        
        # Dividir em treino e teste (80/20)
        X_train, X_test = train_test_split(df, test_size=0.2, random_state=42)
        
        # Criar e treinar o modelo
        model = self._create_model()
        
        logger.info(f"Treinando modelo {self.model_type.value} para região {region}")
        model.fit(X_train)
        
        # Avaliar modelo
        metrics = self._evaluate_model(model, X_test)
        
        # Criar metadados
        model_id = str(uuid.uuid4())
        now = datetime.datetime.now()
        version = now.strftime("%Y%m%d%H%M")
        
        # Obter importância de features (se disponível)
        feature_importance = None
        if hasattr(model, "feature_importances_") or (hasattr(model, "best_estimator_") and hasattr(model.best_estimator_, "feature_importances_")):
            try:
                if hasattr(model, "feature_importances_"):
                    importances = model.feature_importances_
                else:
                    importances = model.best_estimator_.feature_importances_
                
                feature_importance = {feature: float(importance) for feature, importance in zip(df.columns, importances)}
            except:
                pass
        
        metadata = ModelMetadata(
            model_id=model_id,
            model_type=self.model_type,
            version=version,
            created_at=now,
            updated_at=now,
            training_data_size=len(training_data),
            feature_count=len(df.columns),
            performance_metrics=metrics,
            feature_importance=feature_importance,
            regions=[region]
        )
        
        # Salvar modelo e metadados
        self.models[region] = model
        self.model_metadata[region] = metadata
        self._save_model(model, region, metadata)
        
        logger.info(f"Modelo para região {region} treinado com sucesso. Métricas: {metrics}")
        
        return metadata
    
    def _create_model(self) -> Any:
        """
        Cria uma instância do modelo especificado na configuração.
        
        Returns:
            Modelo ML inicializado
        """
        if self.model_type == ModelType.ISOLATION_FOREST:
            model = IsolationForest(
                n_estimators=self.config.get("n_estimators", 100),
                max_samples=self.config.get("max_samples", "auto"),
                contamination=self.config.get("contamination", 0.1),
                random_state=42
            )
        
        elif self.model_type == ModelType.DBSCAN:
            model = Pipeline([
                ('scaler', StandardScaler()),
                ('dbscan', DBSCAN(
                    eps=self.config.get("eps", 0.5),
                    min_samples=self.config.get("min_samples", 5)
                ))
            ])
        
        elif self.model_type == ModelType.HYBRID:
            # Modelo híbrido que combina múltiplos detectores de anomalias
            # Implementação simplificada usando apenas IsolationForest
            model = IsolationForest(
                n_estimators=self.config.get("n_estimators", 100),
                max_samples=self.config.get("max_samples", "auto"),
                contamination=self.config.get("contamination", 0.1),
                random_state=42
            )
            
        else:
            # Modelo padrão
            model = IsolationForest(contamination=0.1, random_state=42)
        
        return model
    
    def _evaluate_model(self, model: Any, X_test: pd.DataFrame) -> Dict[str, float]:
        """
        Avalia o modelo treinado.
        
        Args:
            model: Modelo treinado
            X_test: Conjunto de teste
            
        Returns:
            Métricas de desempenho
        """
        metrics = {}
        
        try:
            # Predição no conjunto de teste
            if self.model_type == ModelType.ISOLATION_FOREST:
                # Isolation Forest retorna scores (-1 para outliers, 1 para inliers)
                raw_scores = model.decision_function(X_test)
                # Normalizar para [0, 1], onde 1 é mais anômalo
                normalized_scores = 1 - ((raw_scores + 1) / 2)
                
                # Adicionar algumas métricas básicas
                metrics["mean_anomaly_score"] = float(np.mean(normalized_scores))
                metrics["median_anomaly_score"] = float(np.median(normalized_scores))
                metrics["std_anomaly_score"] = float(np.std(normalized_scores))
                metrics["max_anomaly_score"] = float(np.max(normalized_scores))
                metrics["min_anomaly_score"] = float(np.min(normalized_scores))
                
            elif self.model_type == ModelType.DBSCAN:
                # DBSCAN rotula outliers como -1 e inliers como clusters >= 0
                labels = model.named_steps['dbscan'].fit_predict(model.named_steps['scaler'].fit_transform(X_test))
                outlier_ratio = np.sum(labels == -1) / len(labels)
                
                metrics["outlier_ratio"] = float(outlier_ratio)
                metrics["n_clusters"] = int(len(np.unique(labels[labels >= 0])))
            
            # Adicionar métricas genéricas
            metrics["sample_size"] = int(len(X_test))
            metrics["feature_count"] = int(X_test.shape[1])
            metrics["timestamp"] = datetime.datetime.now().isoformat()
            
        except Exception as e:
            logger.error(f"Erro ao avaliar modelo: {str(e)}")
            metrics["error"] = str(e)
        
        return metrics
    
    def detect(self, event_data: Dict[str, Any], region: str = None) -> Dict[str, Any]:
        """
        Detecta anomalias em um evento utilizando o modelo apropriado.
        
        Args:
            event_data: Dados do evento a ser analisado
            region: Código da região (opcional, usa padrão se não especificado)
            
        Returns:
            Resultado da detecção com score de anomalia e detalhes
        """
        # Determinar a região a ser usada
        use_region = region or event_data.get("region", self.default_region)
        
        # Se não houver modelo específico para a região, usar modelo global
        if use_region not in self.models and "global" in self.models:
            logger.info(f"Modelo específico para região {use_region} não encontrado. Usando modelo global.")
            use_region = "global"
        
        # Se ainda não houver modelo, retornar erro
        if use_region not in self.models:
            logger.error(f"Modelo não disponível para região {use_region} e não há modelo global.")
            return {
                "error": f"Modelo não disponível para região {use_region}",
                "is_anomaly": False,
                "anomaly_score": 0.0,
                "confidence": 0.0,
                "details": {}
            }
        
        try:
            # Extrair características do evento
            features = self.feature_extractor.extract_features(event_data)
            
            # Verificar se há características suficientes
            if len(features) < 3:
                logger.warning("Poucas características extraídas do evento")
                return {
                    "warning": "Poucas características extraídas para análise confiável",
                    "is_anomaly": False,
                    "anomaly_score": 0.0,
                    "confidence": 0.1,
                    "details": features
                }
            
            # Converter para formato adequado para o modelo
            model = self.models[use_region]
            
            # Converter dicionário para DataFrame
            df = pd.DataFrame([features])
            
            # Remover colunas com valores nulos
            df = df.dropna(axis=1)
            
            # Calcular score de anomalia
            result = {}
            
            if self.model_type == ModelType.ISOLATION_FOREST:
                # Isolation Forest retorna scores (-1 para outliers, 1 para inliers)
                raw_score = model.decision_function(df)[0]
                # Normalizar para [0, 1], onde 1 é mais anômalo
                anomaly_score = 1 - ((raw_score + 1) / 2)
                # Determinar se é anomalia com base no threshold
                is_anomaly = anomaly_score >= self.threshold
                
                result = {
                    "is_anomaly": is_anomaly,
                    "anomaly_score": float(anomaly_score),
                    "raw_score": float(raw_score),
                    "confidence": 0.7 + (0.3 * abs(anomaly_score - 0.5) / 0.5),
                    "threshold": self.threshold,
                    "model_type": self.model_type.value,
                    "region": use_region
                }
                
            elif self.model_type == ModelType.DBSCAN:
                # DBSCAN rotula outliers como -1 e inliers como clusters >= 0
                scaler = model.named_steps['scaler']
                dbscan = model.named_steps['dbscan']
                
                # Aplicar transformação
                X_scaled = scaler.transform(df)
                label = dbscan.fit_predict(X_scaled)[0]
                
                # -1 é outlier
                is_anomaly = label == -1
                
                # Calcular distância ao cluster mais próximo como score
                if is_anomaly:
                    anomaly_score = 0.9  # Alta anomalia
                else:
                    # Estimar score pela distância relativa
                    anomaly_score = 0.2  # Baixa anomalia para inliers
                
                result = {
                    "is_anomaly": is_anomaly,
                    "anomaly_score": float(anomaly_score),
                    "cluster_label": int(label),
                    "confidence": 0.8 if is_anomaly else 0.7,
                    "model_type": self.model_type.value,
                    "region": use_region
                }
            
            # Adicionar principais características que contribuíram para o score
            # (Simplificado para esta implementação)
            top_features = {}
            try:
                # Ordenar features pelo valor absoluto
                sorted_features = sorted(features.items(), key=lambda x: abs(x[1]), reverse=True)
                # Selecionar top 5
                top_features = {k: v for k, v in sorted_features[:5]}
            except:
                pass
            
            result["contributing_factors"] = top_features
            result["timestamp"] = datetime.datetime.now().isoformat()
            
            return result
        
        except Exception as e:
            logger.error(f"Erro na detecção de anomalias: {str(e)}")
            return {
                "error": str(e),
                "is_anomaly": False,
                "anomaly_score": 0.0,
                "confidence": 0.0,
                "exception": type(e).__name__
            }
    
    def batch_detect(self, events: List[Dict[str, Any]], region: str = None) -> List[Dict[str, Any]]:
        """
        Processa um lote de eventos para detecção de anomalias.
        
        Args:
            events: Lista de eventos para análise
            region: Código da região (opcional)
            
        Returns:
            Lista de resultados da detecção
        """
        results = []
        for event in events:
            result = self.detect(event, region)
            results.append(result)
        
        return results


class AnomalyDetectorFactory:
    """
    Factory para criar instâncias do detector de anomalias.
    """
    
    @staticmethod
    def create(config: Dict[str, Any]) -> AnomalyDetector:
        """
        Cria uma instância do detector de anomalias com as configurações fornecidas.
        
        Args:
            config: Configurações para o detector
            
        Returns:
            Instância do AnomalyDetector
        """
        return AnomalyDetector(config)