"""
INNOVABIZ - Processador de Padrões de Digitação
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de processador de autenticação baseado em
           padrões de digitação do usuário.
==================================================================
"""

import numpy as np
import datetime
import logging
import json
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass
from enum import Enum

# Configuração de logging
logger = logging.getLogger(__name__)

@dataclass
class KeystrokeFeatures:
    """Características extraídas de padrões de digitação"""
    dwell_times: List[float]  # Tempo entre pressionar e soltar uma tecla
    flight_times: List[float]  # Tempo entre soltar uma tecla e pressionar a próxima
    typing_speed: float  # Velocidade média de digitação (caracteres por minuto)
    error_rate: float  # Taxa de erros de digitação
    rhythm_consistency: float  # Consistência no ritmo de digitação
    pressure_pattern: Optional[List[float]] = None  # Padrão de pressão (se disponível)
    special_key_usage: Dict[str, int] = None  # Uso de teclas especiais

@dataclass
class TypingPatternContext:
    """Contexto para análise de padrão de digitação"""
    user_id: str
    session_id: str
    timestamp: datetime.datetime
    device_id: str
    keystroke_data: List[Dict[str, Any]]
    historical_profile: Optional[Dict[str, Any]] = None
    application_context: Optional[Dict[str, Any]] = None

class TypingPatternResult:
    """Resultado da análise de padrão de digitação"""
    def __init__(
        self,
        match_score: float,
        risk_level: str,
        features: KeystrokeFeatures,
        anomalies: List[str],
        confidence: float
    ):
        self.match_score = match_score  # 0.0 a 1.0
        self.risk_level = risk_level  # "low", "medium", "high", "critical"
        self.features = features
        self.anomalies = anomalies
        self.confidence = confidence  # 0.0 a 1.0
        self.timestamp = datetime.datetime.now()
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o resultado para dicionário"""
        return {
            "match_score": self.match_score,
            "risk_level": self.risk_level,
            "confidence": self.confidence,
            "anomalies": self.anomalies,
            "timestamp": self.timestamp.isoformat()
        }

class TypingPatternProcessor:
    """Processador para análise de padrões de digitação"""
    
    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """
        Inicializa o processador de padrões de digitação.
        
        Args:
            config: Configuração opcional para o processador
        """
        self.config = config or {}
        self.name = "typing_pattern"
        
        # Configurações padrão
        self.min_keystrokes = self.config.get("min_keystrokes", 20)
        self.match_threshold = self.config.get("match_threshold", 0.75)
        self.confidence_threshold = self.config.get("confidence_threshold", 0.65)
        self.feature_weights = self.config.get("feature_weights", {
            "dwell_times": 0.3,
            "flight_times": 0.3,
            "typing_speed": 0.15,
            "error_rate": 0.1,
            "rhythm_consistency": 0.15
        })
        
        # Inicialização de modelos
        self._initialize_models()
        
        logger.info(f"Processador de padrões de digitação inicializado com configuração: {self.config}")
    
    def _initialize_models(self):
        """Inicializa modelos de análise de padrões"""
        # Em uma implementação real, aqui seriam inicializados modelos 
        # de machine learning para análise de padrões
        self.models = {
            "anomaly_detection": None,  # Modelo para detecção de anomalias
            "user_profile": None,  # Modelo para profile de usuário
            "pattern_matching": None  # Modelo para correspondência de padrões
        }
    
    def process(self, context: TypingPatternContext) -> TypingPatternResult:
        """
        Processa os dados de padrão de digitação.
        
        Args:
            context: Contexto de padrão de digitação
            
        Returns:
            Resultado da análise
        """
        logger.debug(f"Processando padrão de digitação para usuário {context.user_id}")
        
        # Verificar dados suficientes
        if len(context.keystroke_data) < self.min_keystrokes:
            logger.warning(f"Dados insuficientes para análise confiável: {len(context.keystroke_data)} < {self.min_keystrokes}")
            return TypingPatternResult(
                match_score=0.0,
                risk_level="high",
                features=self._extract_features(context.keystroke_data),
                anomalies=["insufficient_data"],
                confidence=0.3
            )
        
        # Extrair características
        features = self._extract_features(context.keystroke_data)
        
        # Se houver perfil histórico, comparar com ele
        if context.historical_profile:
            match_score, anomalies = self._compare_with_profile(features, context.historical_profile)
            confidence = self._calculate_confidence(features, context.historical_profile)
        else:
            # Sem perfil histórico, baixa confiança
            logger.warning(f"Sem perfil histórico para usuário {context.user_id}")
            match_score = 0.5  # Neutro
            anomalies = ["no_historical_profile"]
            confidence = 0.4
        
        # Determinar nível de risco
        risk_level = self._determine_risk_level(match_score, anomalies, confidence)
        
        # Criar resultado
        result = TypingPatternResult(
            match_score=match_score,
            risk_level=risk_level,
            features=features,
            anomalies=anomalies,
            confidence=confidence
        )
        
        logger.debug(f"Análise de padrão de digitação concluída: {result.to_dict()}")
        return result
    
    def _extract_features(self, keystroke_data: List[Dict[str, Any]]) -> KeystrokeFeatures:
        """
        Extrai características dos dados de digitação.
        
        Args:
            keystroke_data: Dados brutos de teclas pressionadas/liberadas
            
        Returns:
            Características extraídas
        """
        # Em uma implementação real, aqui seriam extraídas as características
        # a partir dos dados brutos de keystrokes
        
        # Simulação para o propósito desta implementação
        dwell_times = [k.get("dwellTime", 0) for k in keystroke_data if "dwellTime" in k]
        flight_times = [k.get("flightTime", 0) for k in keystroke_data if "flightTime" in k]
        
        # Calcular velocidade média (caracteres por minuto)
        if dwell_times and flight_times:
            total_time = sum(dwell_times) + sum(flight_times)
            char_count = len(keystroke_data)
            typing_speed = (char_count / total_time) * 60000 if total_time > 0 else 0
        else:
            typing_speed = 0
        
        # Calcular taxa de erros (backspaces / total keys)
        backspaces = sum(1 for k in keystroke_data if k.get("key") == "Backspace")
        error_rate = backspaces / len(keystroke_data) if keystroke_data else 0
        
        # Calcular consistência de ritmo
        if dwell_times and len(dwell_times) > 1:
            dwell_std = np.std(dwell_times)
            dwell_mean = np.mean(dwell_times)
            rhythm_consistency = 1.0 - min(1.0, (dwell_std / dwell_mean) if dwell_mean > 0 else 0)
        else:
            rhythm_consistency = 0.5  # Valor neutro
        
        # Contar uso de teclas especiais
        special_keys = {}
        for k in keystroke_data:
            key = k.get("key", "")
            if key in ["Shift", "Control", "Alt", "Meta", "Tab", "Enter"]:
                special_keys[key] = special_keys.get(key, 0) + 1
        
        return KeystrokeFeatures(
            dwell_times=dwell_times,
            flight_times=flight_times,
            typing_speed=typing_speed,
            error_rate=error_rate,
            rhythm_consistency=rhythm_consistency,
            pressure_pattern=None,  # Não disponível na maioria dos dispositivos
            special_key_usage=special_keys
        )
    
    def _compare_with_profile(
        self, 
        features: KeystrokeFeatures, 
        profile: Dict[str, Any]
    ) -> Tuple[float, List[str]]:
        """
        Compara características extraídas com o perfil do usuário.
        
        Args:
            features: Características extraídas
            profile: Perfil histórico do usuário
            
        Returns:
            Pontuação de correspondência e lista de anomalias
        """
        anomalies = []
        scores = {}
        
        # Comparar tempo de permanência (dwell time)
        profile_dwell = profile.get("avg_dwell_time", 0)
        current_dwell = np.mean(features.dwell_times) if features.dwell_times else 0
        
        if profile_dwell > 0:
            dwell_diff = abs(current_dwell - profile_dwell) / profile_dwell
            scores["dwell_times"] = max(0, 1.0 - dwell_diff)
            
            if dwell_diff > 0.5:
                anomalies.append("abnormal_dwell_time")
        else:
            scores["dwell_times"] = 0.5  # Neutro
        
        # Comparar tempo de voo (flight time)
        profile_flight = profile.get("avg_flight_time", 0)
        current_flight = np.mean(features.flight_times) if features.flight_times else 0
        
        if profile_flight > 0:
            flight_diff = abs(current_flight - profile_flight) / profile_flight
            scores["flight_times"] = max(0, 1.0 - flight_diff)
            
            if flight_diff > 0.5:
                anomalies.append("abnormal_flight_time")
        else:
            scores["flight_times"] = 0.5  # Neutro
        
        # Comparar velocidade de digitação
        profile_speed = profile.get("typing_speed", 0)
        if profile_speed > 0:
            speed_diff = abs(features.typing_speed - profile_speed) / profile_speed
            scores["typing_speed"] = max(0, 1.0 - speed_diff)
            
            if speed_diff > 0.5:
                anomalies.append("abnormal_typing_speed")
        else:
            scores["typing_speed"] = 0.5  # Neutro
        
        # Comparar taxa de erros
        profile_error = profile.get("error_rate", 0)
        error_diff = abs(features.error_rate - profile_error)
        scores["error_rate"] = max(0, 1.0 - (error_diff * 5))  # Erro tem peso maior
        
        if error_diff > 0.2:
            anomalies.append("abnormal_error_rate")
        
        # Comparar consistência de ritmo
        profile_rhythm = profile.get("rhythm_consistency", 0)
        rhythm_diff = abs(features.rhythm_consistency - profile_rhythm)
        scores["rhythm_consistency"] = max(0, 1.0 - (rhythm_diff * 2))
        
        if rhythm_diff > 0.3:
            anomalies.append("abnormal_rhythm")
        
        # Calcular pontuação ponderada
        weighted_score = 0
        total_weight = 0
        
        for feature, score in scores.items():
            weight = self.feature_weights.get(feature, 0.1)
            weighted_score += score * weight
            total_weight += weight
        
        if total_weight > 0:
            match_score = weighted_score / total_weight
        else:
            match_score = 0.5  # Valor neutro
        
        return match_score, anomalies
    
    def _calculate_confidence(
        self, 
        features: KeystrokeFeatures, 
        profile: Dict[str, Any]
    ) -> float:
        """
        Calcula a confiança na análise.
        
        Args:
            features: Características extraídas
            profile: Perfil histórico do usuário
            
        Returns:
            Nível de confiança (0.0 a 1.0)
        """
        # Fatores que afetam a confiança
        factors = []
        
        # Quantidade de dados
        sample_size = len(features.dwell_times)
        if sample_size >= 100:
            factors.append(1.0)
        elif sample_size >= 50:
            factors.append(0.9)
        elif sample_size >= 30:
            factors.append(0.8)
        else:
            factors.append(0.5)
        
        # Qualidade do perfil histórico
        profile_samples = profile.get("sample_count", 0)
        if profile_samples >= 1000:
            factors.append(1.0)
        elif profile_samples >= 500:
            factors.append(0.9)
        elif profile_samples >= 100:
            factors.append(0.7)
        else:
            factors.append(0.5)
        
        # Consistência dos dados
        if features.dwell_times:
            dwell_std = np.std(features.dwell_times)
            dwell_mean = np.mean(features.dwell_times)
            if dwell_mean > 0:
                variation_coef = dwell_std / dwell_mean
                if variation_coef < 0.1:
                    factors.append(1.0)
                elif variation_coef < 0.2:
                    factors.append(0.8)
                elif variation_coef < 0.3:
                    factors.append(0.6)
                else:
                    factors.append(0.4)
            else:
                factors.append(0.5)
        else:
            factors.append(0.5)
        
        # Calcular média dos fatores
        return sum(factors) / len(factors) if factors else 0.5
    
    def _determine_risk_level(
        self, 
        match_score: float, 
        anomalies: List[str], 
        confidence: float
    ) -> str:
        """
        Determina o nível de risco com base na pontuação de correspondência.
        
        Args:
            match_score: Pontuação de correspondência (0.0 a 1.0)
            anomalies: Lista de anomalias detectadas
            confidence: Nível de confiança na análise
            
        Returns:
            Nível de risco: "low", "medium", "high" ou "critical"
        """
        # Ajustar pontuação com base na confiança
        adjusted_score = match_score * confidence
        
        # Verificação de anomalias críticas
        critical_anomalies = ["no_historical_profile", "insufficient_data"]
        if any(a in critical_anomalies for a in anomalies):
            return "high"  # Alto risco por padrão em caso de dados insuficientes
        
        # Determinação do nível de risco
        if adjusted_score >= 0.9:
            return "low"
        elif adjusted_score >= 0.75:
            return "low" if len(anomalies) == 0 else "medium"
        elif adjusted_score >= 0.6:
            return "medium"
        elif adjusted_score >= 0.4:
            return "high"
        else:
            return "critical"
    
    def update_profile(
        self, 
        user_id: str, 
        features: KeystrokeFeatures, 
        profile: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Atualiza o perfil do usuário com novas características.
        
        Args:
            user_id: ID do usuário
            features: Novas características extraídas
            profile: Perfil existente
            
        Returns:
            Perfil atualizado
        """
        # Obter contador de amostras
        sample_count = profile.get("sample_count", 0) + 1
        
        # Calcular médias atualizadas
        def update_avg(current_avg, new_value, count):
            if current_avg is None or count <= 1:
                return new_value
            return ((current_avg * (count - 1)) + new_value) / count
        
        # Atualizar média de tempos de permanência
        current_dwell = np.mean(features.dwell_times) if features.dwell_times else 0
        updated_profile = {
            "user_id": user_id,
            "sample_count": sample_count,
            "last_updated": datetime.datetime.now().isoformat(),
            "avg_dwell_time": update_avg(profile.get("avg_dwell_time"), current_dwell, sample_count),
            "avg_flight_time": update_avg(
                profile.get("avg_flight_time"), 
                np.mean(features.flight_times) if features.flight_times else 0, 
                sample_count
            ),
            "typing_speed": update_avg(profile.get("typing_speed"), features.typing_speed, sample_count),
            "error_rate": update_avg(profile.get("error_rate"), features.error_rate, sample_count),
            "rhythm_consistency": update_avg(
                profile.get("rhythm_consistency"), 
                features.rhythm_consistency, 
                sample_count
            )
        }
        
        # Manter dados históricos para análise de tendências
        historical_data = profile.get("historical_data", [])
        
        # Limitar tamanho do histórico para não crescer indefinidamente
        max_history = 100
        if len(historical_data) >= max_history:
            historical_data = historical_data[-(max_history-1):]
        
        # Adicionar novo ponto ao histórico
        historical_data.append({
            "timestamp": datetime.datetime.now().isoformat(),
            "dwell_time": current_dwell,
            "flight_time": np.mean(features.flight_times) if features.flight_times else 0,
            "typing_speed": features.typing_speed,
            "error_rate": features.error_rate,
            "rhythm_consistency": features.rhythm_consistency
        })
        
        updated_profile["historical_data"] = historical_data
        
        logger.debug(f"Perfil atualizado para usuário {user_id}")
        return updated_profile
