"""
INNOVABIZ - Motor de Autenticação Adaptativa
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Implementação do motor de avaliação de risco e decisão para
           autenticação adaptativa baseada em contexto e comportamento.
==================================================================
"""

import logging
import math
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple, Any

from .models import (
    RiskLevel,
    AuthenticationFactor,
    RiskSignal,
    DeviceFingerprint,
    LocationData,
    BehavioralProfile,
    RiskAssessment,
    AdaptivePolicy
)

# Import AR processors
from .ar_processors import (
    SpatialGestureProcessor,
    GazePatternProcessor,
    EnvironmentBasedProcessor,
    ARBiometricProcessor
)

# Configuração de logger
logger = logging.getLogger("innovabiz.iam.adaptive_auth")


class RiskEngine:
    """Motor de avaliação de risco para autenticação baseada em contexto."""
    
    def __init__(self, config: Optional[Dict] = None):
        """Inicializa o motor de risco com configuração opcional."""
        self.config = config or {}
        self.processors = self._init_signal_processors()
        logger.info("Motor de avaliação de risco inicializado")
    
    def _init_signal_processors(self) -> Dict:
        """Inicializa processadores de sinais de risco."""
        return {
            # Processadores padrão
            "ip_reputation": IPReputationProcessor(),
            "geo_velocity": GeoVelocityProcessor(),
            "device_analysis": DeviceAnalysisProcessor(),
            "behavioral": BehavioralProcessor(),
            "time_pattern": TimePatternProcessor(),
            "credential_anomaly": CredentialAnomalyProcessor(),
            
            # Processadores AR
            "ar_spatial_gesture": SpatialGestureProcessor(),
            "ar_gaze_pattern": GazePatternProcessor(),
            "ar_environment": EnvironmentBasedProcessor(),
            "ar_biometric": ARBiometricProcessor()
        }
    
    def evaluate(self, 
                 user_id: uuid.UUID, 
                 tenant_id: uuid.UUID,
                 auth_context: Dict[str, Any],
                 policy: AdaptivePolicy) -> RiskAssessment:
        """
        Avalia o risco de uma tentativa de autenticação e determina os requisitos.
        
        Args:
            user_id: ID do usuário que está tentando autenticar
            tenant_id: ID do tenant/organização
            auth_context: Contexto completo da autenticação (IP, dispositivo, localização, etc.)
            policy: Política de autenticação adaptativa a ser aplicada
            
        Returns:
            RiskAssessment: Avaliação de risco com nível, pontuação e requisitos
        """
        logger.debug(f"Iniciando avaliação de risco para usuário {user_id}")
        
        # Extrai dados do contexto de autenticação
        ip_address = auth_context.get("ip_address")
        if not ip_address:
            logger.warning("IP não fornecido no contexto de autenticação")
            ip_address = "0.0.0.0"
        
        # Processa impressão digital do dispositivo
        device_data = auth_context.get("device_data", {})
        device_fingerprint = self._process_device_data(user_id, device_data)
        
        # Processa dados de localização
        location_data = self._process_location_data(ip_address)
        
        # Coleta sinais de risco de todos os processadores
        risk_signals = []
        for processor_name, processor in self.processors.items():
            try:
                if processor_name == "ip_reputation" and policy.enable_geolocation_check:
                    signals = processor.process(ip_address, location_data)
                elif processor_name == "geo_velocity" and policy.enable_velocity_detection:
                    signals = processor.process(user_id, location_data)
                elif processor_name == "device_analysis" and policy.enable_device_fingerprinting:
                    signals = processor.process(user_id, device_fingerprint)
                elif processor_name == "behavioral" and policy.enable_behavioral_analysis:
                    signals = processor.process(user_id, auth_context)
                elif processor_name == "time_pattern":
                    signals = processor.process(user_id, datetime.utcnow())
                # Processadores AR
                elif processor_name == "ar_spatial_gesture" and policy.enable_ar_authentication:
                    signals = processor.process(user_id, auth_context)
                elif processor_name == "ar_gaze_pattern" and policy.enable_ar_authentication:
                    signals = processor.process(user_id, auth_context)
                elif processor_name == "ar_environment" and policy.enable_ar_authentication:
                    signals = processor.process(user_id, auth_context)
                elif processor_name == "ar_biometric" and policy.enable_ar_authentication:
                    signals = processor.process(user_id, auth_context)
                elif processor_name == "credential_anomaly":
                    signals = processor.process(user_id, auth_context.get("auth_method"))
                else:
                    signals = []
                
                risk_signals.extend(signals)
            except Exception as e:
                logger.error(f"Erro ao processar sinal {processor_name}: {str(e)}")
        
        # Calcula pontuação de risco com base nos sinais coletados
        risk_score = self._calculate_risk_score(risk_signals, policy)
        
        # Determina nível de risco com base na pontuação
        risk_level = self._determine_risk_level(risk_score, policy)
        
        # Determina fatores de autenticação necessários com base no nível de risco
        required_factors = self._determine_required_factors(risk_level, policy)
        
        # Cria justificativa para a avaliação de risco
        assessment_reason = self._create_assessment_reason(risk_signals, risk_level)
        
        # Cria objeto RiskAssessment
        assessment = RiskAssessment(
            user_id=user_id,
            session_id=auth_context.get("session_id"),
            ip_address=ip_address,
            device_fingerprint=device_fingerprint,
            location_data=location_data,
            risk_signals=risk_signals,
            authentication_context=auth_context,
            risk_level=risk_level,
            risk_score=risk_score,
            required_factors=required_factors,
            assessment_reason=assessment_reason
        )
        
        # Registra a avaliação para auditoria e análise
        self._log_assessment(assessment)
        
        return assessment
    
    def _process_device_data(self, user_id: uuid.UUID, device_data: Dict) -> DeviceFingerprint:
        """Processa dados do dispositivo e cria uma impressão digital."""
        # Implementação simplificada - em produção seria mais abrangente
        return DeviceFingerprint(
            device_id=device_data.get("device_id", str(uuid.uuid4())),
            user_agent=device_data.get("user_agent", "Unknown"),
            os_info=device_data.get("os_info", "Unknown"),
            browser_info=device_data.get("browser_info", "Unknown"),
            screen_resolution=device_data.get("screen_resolution", "0x0"),
            timezone=device_data.get("timezone", "UTC"),
            language=device_data.get("language", "en"),
            canvas_fingerprint=device_data.get("canvas_fingerprint"),
            webgl_fingerprint=device_data.get("webgl_fingerprint"),
            font_fingerprint=device_data.get("font_fingerprint"),
            hardware_concurrency=device_data.get("hardware_concurrency"),
            is_trusted_device=self._is_trusted_device(user_id, device_data.get("device_id", "")),
            last_seen=datetime.utcnow(),
            risk_score=self._calculate_device_risk(user_id, device_data)
        )
    
    def _is_trusted_device(self, user_id: uuid.UUID, device_id: str) -> bool:
        """Verifica se o dispositivo é confiável para o usuário."""
        # Em produção, consultaria uma tabela de dispositivos confiáveis
        # Implementação simplificada para demonstração
        return False
    
    def _calculate_device_risk(self, user_id: uuid.UUID, device_data: Dict) -> float:
        """Calcula o risco associado ao dispositivo."""
        # Lógica de avaliação de risco do dispositivo
        # Em produção, consideraria vários fatores como idade do dispositivo,
        # anomalias no browser fingerprint, consistência histórica, etc.
        return 0.3  # Valor de demonstração
    
    def _process_location_data(self, ip_address: str) -> LocationData:
        """Processa o IP e cria dados de localização."""
        # Em produção, usaria um serviço de geolocalização de IP
        # Implementação simulada para demonstração
        return LocationData(
            ip_address=ip_address,
            country_code="BR",
            region="SP",
            city="São Paulo",
            latitude=-23.5505,
            longitude=-46.6333,
            isp="Example ISP",
            is_vpn=False,
            is_proxy=False,
            is_hosting=False,
            is_tor=False
        )
    
    def _calculate_risk_score(self, signals: List[RiskSignal], policy: AdaptivePolicy) -> float:
        """Calcula pontuação de risco agregada com base nos sinais coletados."""
        if not signals:
            return 0.5  # Risco médio na ausência de sinais
            
        # Pesos diferentes para diferentes tipos de sinais
        # Em produção, esses pesos seriam configuráveis por política
        weights = {
            "ip_reputation": 0.2,
            "geo_velocity": 0.15,
            "device_trust": 0.15,
            "behavioral": 0.2,
            "time_pattern": 0.1,
            "new_location": 0.15,
            "failed_attempts": 0.2,
            "credential_anomaly": 0.2,
            # Pesos adicionais para outros tipos de sinais
        }
        
        # Calcula pontuação ponderada
        total_weight = 0
        weighted_score = 0
        
        for signal in signals:
            signal_type = signal.signal_type
            if isinstance(signal.signal_value, (int, float)):
                value = float(signal.signal_value)
            elif isinstance(signal.signal_value, bool):
                value = 1.0 if signal.signal_value else 0.0
            else:
                continue  # Ignora sinais com valores não numéricos
                
            # Multiplica pelo peso
            weight = weights.get(signal_type, 0.1)
            confidence = signal.confidence
            
            # Ajusta peso pela confiança
            adjusted_weight = weight * confidence
            weighted_score += value * adjusted_weight
            total_weight += adjusted_weight
        
        # Normaliza para valor entre 0 e 1
        if total_weight > 0:
            final_score = weighted_score / total_weight
        else:
            final_score = 0.5
            
        # Aplica sensibilidade configurada
        sensitivity = policy.anomaly_detection_sensitivity
        # Ajusta a curva de risco para refletir a sensibilidade
        # Uma sensibilidade maior amplifica riscos moderados
        if sensitivity != 0.5:
            final_score = self._adjust_for_sensitivity(final_score, sensitivity)
            
        return min(1.0, max(0.0, final_score))
    
    def _adjust_for_sensitivity(self, score: float, sensitivity: float) -> float:
        """Ajusta a pontuação de risco para refletir a sensibilidade configurada."""
        if sensitivity > 0.5:
            # Maior sensibilidade aumenta o score (mais conservador)
            factor = 2 * (sensitivity - 0.5)
            return score + (1 - score) * factor * score
        elif sensitivity < 0.5:
            # Menor sensibilidade diminui o score (mais permissivo)
            factor = 2 * (0.5 - sensitivity)
            return score - score * factor * (1 - score)
        else:
            return score
    
    def _determine_risk_level(self, risk_score: float, policy: AdaptivePolicy) -> RiskLevel:
        """Determina o nível de risco com base na pontuação e política."""
        if risk_score >= policy.risk_threshold_critical:
            return RiskLevel.CRITICAL
        elif risk_score >= policy.risk_threshold_high:
            return RiskLevel.HIGH
        elif risk_score >= policy.risk_threshold_medium:
            return RiskLevel.MEDIUM
        else:
            return RiskLevel.LOW
    
    def _determine_required_factors(self, risk_level: RiskLevel, policy: AdaptivePolicy) -> List[AuthenticationFactor]:
        """Determina os fatores de autenticação necessários com base no nível de risco."""
        if risk_level == RiskLevel.CRITICAL:
            return policy.factors_critical_risk
        elif risk_level == RiskLevel.HIGH:
            return policy.factors_high_risk
        elif risk_level == RiskLevel.MEDIUM:
            return policy.factors_medium_risk
        else:
            return policy.factors_low_risk
    
    def _create_assessment_reason(self, signals: List[RiskSignal], risk_level: RiskLevel) -> str:
        """Cria uma justificativa em linguagem natural para a avaliação de risco."""
        # Identifica os principais sinais que contribuíram para o risco
        key_signals = sorted(
            [s for s in signals if isinstance(s.signal_value, (int, float)) and float(s.signal_value) > 0.5],
            key=lambda x: float(x.signal_value) if isinstance(x.signal_value, (int, float)) else 0,
            reverse=True
        )[:3]
        
        if not key_signals:
            return f"Nível de risco {risk_level.value} determinado por análise geral."
        
        signal_reasons = []
        for signal in key_signals:
            if signal.signal_type == "ip_reputation":
                signal_reasons.append("reputação do endereço IP")
            elif signal.signal_type == "geo_velocity":
                signal_reasons.append("mudança rápida de localização geográfica")
            elif signal.signal_type == "device_trust":
                signal_reasons.append("dispositivo não reconhecido")
            elif signal.signal_type == "behavioral":
                signal_reasons.append("padrão de comportamento incomum")
            elif signal.signal_type == "time_pattern":
                signal_reasons.append("horário de acesso incomum")
            elif signal.signal_type == "new_location":
                signal_reasons.append("localização não reconhecida")
            elif signal.signal_type == "failed_attempts":
                signal_reasons.append("várias tentativas falhas de login")
            elif signal.signal_type == "credential_anomaly":
                signal_reasons.append("anomalia nas credenciais")
            else:
                signal_reasons.append(f"sinal de risco: {signal.signal_type}")
        
        if len(signal_reasons) == 1:
            reason = f"Nível de risco {risk_level.value} devido a {signal_reasons[0]}."
        elif len(signal_reasons) == 2:
            reason = f"Nível de risco {risk_level.value} devido a {signal_reasons[0]} e {signal_reasons[1]}."
        else:
            reason = f"Nível de risco {risk_level.value} devido a {signal_reasons[0]}, {signal_reasons[1]} e {signal_reasons[2]}."
        
        return reason
    
    def _log_assessment(self, assessment: RiskAssessment) -> None:
        """Registra a avaliação para auditoria e análise futura."""
        # Em produção, persistiria em banco de dados
        logger.info(
            f"Avaliação de risco: usuário={assessment.user_id}, "
            f"nível={assessment.risk_level.value}, "
            f"score={assessment.risk_score:.2f}, "
            f"fatores={[f.value for f in assessment.required_factors]}, "
            f"motivo={assessment.assessment_reason}"
        )


# Classes de processadores de sinais individuais
# Estas seriam implementadas completamente em produção

class IPReputationProcessor:
    """Processador para avaliar a reputação do endereço IP."""
    
    def process(self, ip_address: str, location_data: LocationData) -> List[RiskSignal]:
        # Implementação simulada para demonstração
        # Em produção, consultaria bases de dados de reputação de IP
        signals = []
        
        # Sinal de risco para proxy/VPN/Tor
        if location_data.is_proxy or location_data.is_vpn or location_data.is_tor:
            signals.append(RiskSignal(
                signal_type="ip_reputation",
                signal_value=0.8,
                confidence=0.9,
                timestamp=datetime.utcnow()
            ))
        
        # Lista de países de alto risco (exemplo)
        high_risk_countries = ["KP", "IR", "SY"]
        if location_data.country_code in high_risk_countries:
            signals.append(RiskSignal(
                signal_type="ip_reputation",
                signal_value=0.9,
                confidence=0.95,
                timestamp=datetime.utcnow()
            ))
        
        return signals


class GeoVelocityProcessor:
    """Processador para detectar velocidade geográfica impossível entre logins."""
    
    def process(self, user_id: uuid.UUID, location_data: LocationData) -> List[RiskSignal]:
        # Implementação simulada para demonstração
        # Em produção, consultaria histórico de localizações do usuário
        signals = []
        
        # Simulação: 10% de chance de detectar viagem impossível
        if hash(str(user_id) + str(datetime.utcnow().day)) % 10 == 0:
            signals.append(RiskSignal(
                signal_type="geo_velocity",
                signal_value=0.95,
                confidence=0.85,
                timestamp=datetime.utcnow()
            ))
        
        return signals


class DeviceAnalysisProcessor:
    """Processador para analisar confiabilidade do dispositivo."""
    
    def process(self, user_id: uuid.UUID, device: DeviceFingerprint) -> List[RiskSignal]:
        signals = []
        
        # Dispositivo não confiável
        if not device.is_trusted_device:
            signals.append(RiskSignal(
                signal_type="device_trust",
                signal_value=0.7,
                confidence=0.9,
                timestamp=datetime.utcnow()
            ))
        
        # Outros sinais de análise de dispositivo seriam adicionados aqui
        
        return signals


class BehavioralProcessor:
    """Processador para analisar padrões comportamentais do usuário."""
    
    def process(self, user_id: uuid.UUID, auth_context: Dict) -> List[RiskSignal]:
        # Implementação simulada para demonstração
        # Em produção, consultaria perfil comportamental do usuário
        signals = []
        
        # Simulação: 15% de chance de detectar comportamento anômalo
        if hash(str(user_id) + str(datetime.utcnow().hour)) % 20 < 3:
            signals.append(RiskSignal(
                signal_type="behavioral",
                signal_value=0.75,
                confidence=0.8,
                timestamp=datetime.utcnow()
            ))
        
        return signals


class TimePatternProcessor:
    """Processador para analisar padrões temporais de acesso."""
    
    def process(self, user_id: uuid.UUID, current_time: datetime) -> List[RiskSignal]:
        signals = []
        
        # Exemplo: horário noturno (2-5 da manhã) é mais suspeito
        hour = current_time.hour
        if 2 <= hour <= 5:
            signals.append(RiskSignal(
                signal_type="time_pattern",
                signal_value=0.6,
                confidence=0.7,
                timestamp=datetime.utcnow()
            ))
        
        return signals


class CredentialAnomalyProcessor:
    """Processador para detectar anomalias em credenciais."""
    
    def process(self, user_id: uuid.UUID, auth_method: Optional[str]) -> List[RiskSignal]:
        signals = []
        
        # Exemplos de sinais que seriam implementados em produção:
        # - Detecção de padrão de senha comum
        # - Credenciais vazadas em breaches conhecidos
        # - Padrão de digitação inconsistente
        
        return signals
