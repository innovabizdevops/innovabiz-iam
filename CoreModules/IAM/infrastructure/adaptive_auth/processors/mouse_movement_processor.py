"""
INNOVABIZ - Processador de Movimentos do Mouse
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de processador de autenticação baseado em
           análise de movimentos do mouse para detecção de padrões
           comportamentais e autenticação contínua.
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
class MouseMovementFeatures:
    """Características extraídas de movimentos do mouse"""
    velocity_profile: List[float]  # Perfil de velocidade ao longo do tempo
    acceleration_profile: List[float]  # Perfil de aceleração
    jerk_profile: Optional[List[float]] = None  # Taxa de variação da aceleração
    curvature: List[float] = None  # Curvatura das trajetórias
    direction_changes: int = 0  # Número de mudanças de direção
    click_patterns: Dict[str, Any] = None  # Padrões de clique
    hover_patterns: List[Dict[str, Any]] = None  # Padrões de hover
    scroll_patterns: List[Dict[str, Any]] = None  # Padrões de rolagem

@dataclass
class MouseMovementContext:
    """Contexto para análise de movimentos do mouse"""
    user_id: str
    session_id: str
    timestamp: datetime.datetime
    device_id: str
    movement_data: List[Dict[str, Any]]
    historical_profile: Optional[Dict[str, Any]] = None
    application_context: Optional[Dict[str, Any]] = None
    screen_resolution: Optional[Tuple[int, int]] = None

class MouseMovementResult:
    """Resultado da análise de movimentos do mouse"""
    def __init__(
        self,
        match_score: float,
        risk_level: str,
        features: MouseMovementFeatures,
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

class MouseMovementProcessor:
    """Processador para análise de movimentos do mouse"""
    
    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """
        Inicializa o processador de movimentos do mouse.
        
        Args:
            config: Configuração opcional para o processador
        """
        self.config = config or {}
        self.name = "mouse_movement"
        
        # Configurações padrão
        self.min_movements = self.config.get("min_movements", 50)
        self.match_threshold = self.config.get("match_threshold", 0.75)
        self.confidence_threshold = self.config.get("confidence_threshold", 0.65)
        self.sampling_rate = self.config.get("sampling_rate", 10)  # Hz
        self.feature_weights = self.config.get("feature_weights", {
            "velocity": 0.25,
            "acceleration": 0.2,
            "curvature": 0.2,
            "direction_changes": 0.15,
            "click_patterns": 0.2
        })
        
        # Inicialização de modelos
        self._initialize_models()
        
        logger.info(f"Processador de movimentos do mouse inicializado com configuração: {self.config}")
    
    def _initialize_models(self):
        """Inicializa modelos de análise de movimentos"""
        # Em uma implementação real, aqui seriam inicializados modelos 
        # de machine learning para análise de movimentos
        self.models = {
            "anomaly_detection": None,  # Modelo para detecção de anomalias
            "user_profile": None,  # Modelo para profile de usuário
            "trajectory_analysis": None  # Modelo para análise de trajetórias
        }
    
    def process(self, context: MouseMovementContext) -> MouseMovementResult:
        """
        Processa os dados de movimentos do mouse.
        
        Args:
            context: Contexto de movimentos do mouse
            
        Returns:
            Resultado da análise
        """
        logger.debug(f"Processando movimentos do mouse para usuário {context.user_id}")
        
        # Verificar dados suficientes
        if len(context.movement_data) < self.min_movements:
            logger.warning(f"Dados insuficientes para análise confiável: {len(context.movement_data)} < {self.min_movements}")
            return MouseMovementResult(
                match_score=0.0,
                risk_level="high",
                features=self._extract_features(context.movement_data, context.screen_resolution),
                anomalies=["insufficient_data"],
                confidence=0.3
            )
        
        # Extrair características
        features = self._extract_features(context.movement_data, context.screen_resolution)
        
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
        result = MouseMovementResult(
            match_score=match_score,
            risk_level=risk_level,
            features=features,
            anomalies=anomalies,
            confidence=confidence
        )
        
        logger.debug(f"Análise de movimentos do mouse concluída: {result.to_dict()}")
        return result
    
    def _extract_features(
        self, 
        movement_data: List[Dict[str, Any]],
        screen_resolution: Optional[Tuple[int, int]] = None
    ) -> MouseMovementFeatures:
        """
        Extrai características dos dados de movimento do mouse.
        
        Args:
            movement_data: Dados brutos de movimento do mouse
            screen_resolution: Resolução da tela para normalização
            
        Returns:
            Características extraídas
        """
        # Em uma implementação real, aqui seriam extraídas as características
        # a partir dos dados brutos de movimento
        
        # Extrair posições e timestamps
        positions = []
        timestamps = []
        
        for entry in movement_data:
            if "x" in entry and "y" in entry and "timestamp" in entry:
                positions.append((entry["x"], entry["y"]))
                timestamps.append(entry["timestamp"])
        
        if not positions or len(positions) < 2:
            # Retornar características vazias se não houver dados suficientes
            return MouseMovementFeatures(
                velocity_profile=[],
                acceleration_profile=[],
                curvature=[]
            )
        
        # Normalizar posições se houver informação de resolução
        if screen_resolution and screen_resolution[0] > 0 and screen_resolution[1] > 0:
            positions = [(x / screen_resolution[0], y / screen_resolution[1]) for x, y in positions]
        
        # Calcular velocidades entre pontos consecutivos
        velocities = []
        for i in range(1, len(positions)):
            # Calcular distância euclidiana
            dx = positions[i][0] - positions[i-1][0]
            dy = positions[i][1] - positions[i-1][1]
            distance = np.sqrt(dx*dx + dy*dy)
            
            # Calcular tempo decorrido
            try:
                dt = (timestamps[i] - timestamps[i-1]) / 1000.0  # em segundos
                if dt > 0:
                    velocities.append(distance / dt)
                else:
                    velocities.append(0)
            except (TypeError, ValueError):
                velocities.append(0)
        
        # Calcular acelerações
        accelerations = []
        for i in range(1, len(velocities)):
            try:
                dt = (timestamps[i+1] - timestamps[i]) / 1000.0  # em segundos
                if dt > 0:
                    accelerations.append((velocities[i] - velocities[i-1]) / dt)
                else:
                    accelerations.append(0)
            except (TypeError, ValueError, IndexError):
                accelerations.append(0)
        
        # Calcular curvaturas
        curvatures = []
        for i in range(1, len(positions) - 1):
            try:
                # Vetores de movimento
                v1 = (positions[i][0] - positions[i-1][0], positions[i][1] - positions[i-1][1])
                v2 = (positions[i+1][0] - positions[i][0], positions[i+1][1] - positions[i][1])
                
                # Normalizar vetores
                v1_mag = np.sqrt(v1[0]*v1[0] + v1[1]*v1[1])
                v2_mag = np.sqrt(v2[0]*v2[0] + v2[1]*v2[1])
                
                if v1_mag > 0 and v2_mag > 0:
                    v1_norm = (v1[0]/v1_mag, v1[1]/v1_mag)
                    v2_norm = (v2[0]/v2_mag, v2[1]/v2_mag)
                    
                    # Produto escalar (coseno do ângulo)
                    dot_product = v1_norm[0]*v2_norm[0] + v1_norm[1]*v2_norm[1]
                    
                    # Limitar para evitar erros numéricos
                    dot_product = max(-1.0, min(1.0, dot_product))
                    
                    # Ângulo em radianos
                    angle = np.arccos(dot_product)
                    
                    # Curvatura (ângulo / distância)
                    dist = (v1_mag + v2_mag) / 2
                    curvature = angle / dist if dist > 0 else 0
                    
                    curvatures.append(curvature)
                else:
                    curvatures.append(0)
            except (IndexError, ValueError):
                curvatures.append(0)
        
        # Contar mudanças de direção
        direction_changes = 0
        prev_direction = None
        
        for i in range(1, len(positions)):
            dx = positions[i][0] - positions[i-1][0]
            dy = positions[i][1] - positions[i-1][1]
            
            # Determinar direção atual (dividida em 8 setores)
            angle = np.arctan2(dy, dx)
            sector = int((angle + np.pi) / (np.pi/4)) % 8
            
            if prev_direction is not None and sector != prev_direction:
                direction_changes += 1
            
            prev_direction = sector
        
        # Analisar padrões de clique (se disponíveis)
        click_patterns = {}
        clicks = [m for m in movement_data if m.get("type") == "click"]
        
        if clicks:
            # Contagem de tipos de clique
            click_patterns["single_count"] = len([c for c in clicks if c.get("click_type") == "single"])
            click_patterns["double_count"] = len([c for c in clicks if c.get("click_type") == "double"])
            click_patterns["right_count"] = len([c for c in clicks if c.get("button") == "right"])
            
            # Intervalos entre cliques
            click_intervals = []
            for i in range(1, len(clicks)):
                try:
                    interval = clicks[i].get("timestamp", 0) - clicks[i-1].get("timestamp", 0)
                    if interval > 0:
                        click_intervals.append(interval)
                except (TypeError, ValueError):
                    pass
            
            if click_intervals:
                click_patterns["avg_interval"] = np.mean(click_intervals)
                click_patterns["std_interval"] = np.std(click_intervals)
        
        # Analisar padrões de hover (se disponíveis)
        hover_patterns = []
        hovers = [m for m in movement_data if m.get("type") == "hover"]
        
        if hovers:
            for hover in hovers:
                hover_patterns.append({
                    "duration": hover.get("duration", 0),
                    "x": hover.get("x", 0),
                    "y": hover.get("y", 0),
                    "element": hover.get("element", "")
                })
        
        # Analisar padrões de scroll (se disponíveis)
        scroll_patterns = []
        scrolls = [m for m in movement_data if m.get("type") == "scroll"]
        
        if scrolls:
            for scroll in scrolls:
                scroll_patterns.append({
                    "delta_y": scroll.get("delta_y", 0),
                    "speed": scroll.get("speed", 0),
                    "timestamp": scroll.get("timestamp", 0)
                })
        
        return MouseMovementFeatures(
            velocity_profile=velocities,
            acceleration_profile=accelerations,
            jerk_profile=None,  # Não calculado nesta implementação
            curvature=curvatures,
            direction_changes=direction_changes,
            click_patterns=click_patterns,
            hover_patterns=hover_patterns,
            scroll_patterns=scroll_patterns
        )
    
    def _compare_with_profile(
        self, 
        features: MouseMovementFeatures, 
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
        
        # Comparar velocidade média
        if features.velocity_profile:
            current_velocity = np.mean(features.velocity_profile)
            profile_velocity = profile.get("avg_velocity", 0)
            
            if profile_velocity > 0:
                velocity_diff = abs(current_velocity - profile_velocity) / profile_velocity
                scores["velocity"] = max(0, 1.0 - velocity_diff)
                
                if velocity_diff > 0.5:
                    anomalies.append("abnormal_velocity")
            else:
                scores["velocity"] = 0.5  # Neutro
        else:
            scores["velocity"] = 0.5  # Neutro
        
        # Comparar aceleração média
        if features.acceleration_profile:
            current_accel = np.mean(features.acceleration_profile)
            profile_accel = profile.get("avg_acceleration", 0)
            
            if profile_accel != 0:  # Evitar divisão por zero
                accel_diff = abs(current_accel - profile_accel) / (abs(profile_accel) + 0.001)
                scores["acceleration"] = max(0, 1.0 - min(1.0, accel_diff))
                
                if accel_diff > 0.5:
                    anomalies.append("abnormal_acceleration")
            else:
                scores["acceleration"] = 0.5  # Neutro
        else:
            scores["acceleration"] = 0.5  # Neutro
        
        # Comparar curvatura média
        if features.curvature:
            current_curvature = np.mean(features.curvature)
            profile_curvature = profile.get("avg_curvature", 0)
            
            if profile_curvature > 0:
                curvature_diff = abs(current_curvature - profile_curvature) / profile_curvature
                scores["curvature"] = max(0, 1.0 - curvature_diff)
                
                if curvature_diff > 0.5:
                    anomalies.append("abnormal_curvature")
            else:
                scores["curvature"] = 0.5  # Neutro
        else:
            scores["curvature"] = 0.5  # Neutro
        
        # Comparar taxa de mudanças de direção
        profile_changes = profile.get("avg_direction_changes", 0)
        if profile_changes > 0 and features.direction_changes > 0:
            changes_ratio = features.direction_changes / len(features.velocity_profile) if features.velocity_profile else 0
            profile_ratio = profile_changes
            
            changes_diff = abs(changes_ratio - profile_ratio) / profile_ratio if profile_ratio > 0 else 1.0
            scores["direction_changes"] = max(0, 1.0 - changes_diff)
            
            if changes_diff > 0.5:
                anomalies.append("abnormal_direction_changes")
        else:
            scores["direction_changes"] = 0.5  # Neutro
        
        # Comparar padrões de clique
        if features.click_patterns and "avg_interval" in features.click_patterns:
            current_interval = features.click_patterns.get("avg_interval", 0)
            profile_interval = profile.get("click_patterns", {}).get("avg_interval", 0)
            
            if profile_interval > 0:
                interval_diff = abs(current_interval - profile_interval) / profile_interval
                scores["click_patterns"] = max(0, 1.0 - interval_diff)
                
                if interval_diff > 0.5:
                    anomalies.append("abnormal_click_pattern")
            else:
                scores["click_patterns"] = 0.5  # Neutro
        else:
            scores["click_patterns"] = 0.5  # Neutro
        
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
        features: MouseMovementFeatures, 
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
        sample_size = len(features.velocity_profile)
        if sample_size >= 100:
            factors.append(1.0)
        elif sample_size >= 50:
            factors.append(0.9)
        elif sample_size >= 20:
            factors.append(0.7)
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
        if features.velocity_profile:
            velocity_std = np.std(features.velocity_profile)
            velocity_mean = np.mean(features.velocity_profile)
            if velocity_mean > 0:
                variation_coef = velocity_std / velocity_mean
                if variation_coef < 0.3:
                    factors.append(0.9)
                elif variation_coef < 0.5:
                    factors.append(0.7)
                elif variation_coef < 0.8:
                    factors.append(0.5)
                else:
                    factors.append(0.3)
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
            return "low" if len(anomalies) <= 1 else "medium"
        elif adjusted_score >= 0.6:
            return "medium"
        elif adjusted_score >= 0.4:
            return "high"
        else:
            return "critical"
    
    def update_profile(
        self, 
        user_id: str, 
        features: MouseMovementFeatures, 
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
        
        # Calcular valores médios atuais
        current_velocity = np.mean(features.velocity_profile) if features.velocity_profile else 0
        current_accel = np.mean(features.acceleration_profile) if features.acceleration_profile else 0
        current_curvature = np.mean(features.curvature) if features.curvature else 0
        
        # Calcular taxa de mudanças de direção
        changes_ratio = features.direction_changes / len(features.velocity_profile) if features.velocity_profile and len(features.velocity_profile) > 0 else 0
        
        # Atualizar perfil
        updated_profile = {
            "user_id": user_id,
            "sample_count": sample_count,
            "last_updated": datetime.datetime.now().isoformat(),
            "avg_velocity": update_avg(profile.get("avg_velocity"), current_velocity, sample_count),
            "avg_acceleration": update_avg(profile.get("avg_acceleration"), current_accel, sample_count),
            "avg_curvature": update_avg(profile.get("avg_curvature"), current_curvature, sample_count),
            "avg_direction_changes": update_avg(profile.get("avg_direction_changes"), changes_ratio, sample_count),
        }
        
        # Atualizar padrões de clique
        if features.click_patterns:
            click_profile = profile.get("click_patterns", {})
            updated_click_patterns = {}
            
            for key, value in features.click_patterns.items():
                updated_click_patterns[key] = update_avg(click_profile.get(key), value, sample_count)
            
            updated_profile["click_patterns"] = updated_click_patterns
        
        # Manter dados históricos para análise de tendências
        historical_data = profile.get("historical_data", [])
        
        # Limitar tamanho do histórico para não crescer indefinidamente
        max_history = 50
        if len(historical_data) >= max_history:
            historical_data = historical_data[-(max_history-1):]
        
        # Adicionar novo ponto ao histórico
        historical_data.append({
            "timestamp": datetime.datetime.now().isoformat(),
            "velocity": current_velocity,
            "acceleration": current_accel,
            "curvature": current_curvature,
            "direction_changes_ratio": changes_ratio
        })
        
        updated_profile["historical_data"] = historical_data
        
        logger.debug(f"Perfil de movimento do mouse atualizado para usuário {user_id}")
        return updated_profile
