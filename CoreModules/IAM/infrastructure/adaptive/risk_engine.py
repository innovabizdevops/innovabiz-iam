"""
INNOVABIZ - Motor de Avaliação de Risco para Autenticação Adaptativa
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Motor de avaliação de risco para autenticação adaptativa,
           baseado nos benchmarks da Gartner e Forrester.
==================================================================
"""

import json
import logging
import datetime
from enum import Enum, auto
from typing import Dict, List, Any, Optional, Set, Tuple
from dataclasses import dataclass, field
from pathlib import Path
import ipaddress
import re
import uuid

# Configuração de logging
logger = logging.getLogger(__name__)


class RiskLevel(Enum):
    """Níveis de risco para autenticação adaptativa."""
    LOW = auto()
    MEDIUM = auto()
    HIGH = auto()
    CRITICAL = auto()


class RiskSignalCategory(Enum):
    """Categorias de sinais de risco."""
    LOCATION = auto()
    DEVICE = auto()
    BEHAVIOR = auto()
    TIME = auto()
    NETWORK = auto()
    RESOURCE = auto()
    HISTORY = auto()
    THREAT_INTEL = auto()


@dataclass
class RiskSignal:
    """Representação de um sinal de risco."""
    name: str
    category: RiskSignalCategory
    weight: float
    value: Optional[Any] = None
    confidence: float = 1.0
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class RiskAssessment:
    """Resultado da avaliação de risco."""
    level: RiskLevel
    score: float
    signals: List[RiskSignal]
    timestamp: datetime.datetime = field(default_factory=datetime.datetime.now)
    session_id: str = field(default_factory=lambda: str(uuid.uuid4()))
    required_factors: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)


class RiskProcessorInterface:
    """Interface para processadores de sinais de risco."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Processa o contexto e retorna um sinal de risco.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        raise NotImplementedError("Implemente na subclasse")


class RiskEngine:
    """Motor de avaliação de risco para autenticação adaptativa."""
    
    def __init__(self, config_path: Optional[Path] = None):
        """
        Inicializa o motor de avaliação de risco.
        
        Args:
            config_path: Caminho para arquivo de configuração
        """
        self.processors: List[RiskProcessorInterface] = []
        self.config = {}
        
        if config_path and config_path.exists():
            with open(config_path, 'r') as f:
                self.config = json.load(f)
        
        self._load_default_config()
        self._initialize_processors()
    
    def _load_default_config(self):
        """Carrega configuração padrão se não fornecida."""
        if not self.config:
            self.config = {
                "thresholds": {
                    "low": 25,
                    "medium": 50,
                    "high": 75,
                    "critical": 90
                },
                "factor_requirements": {
                    "low": ["password"],
                    "medium": ["password", "totp"],
                    "high": ["password", "totp", "push"],
                    "critical": ["password", "totp", "biometric_advanced"]
                },
                "signal_weights": {
                    "location": 0.25,
                    "device": 0.2,
                    "behavior": 0.2,
                    "time": 0.1,
                    "network": 0.15,
                    "resource": 0.1
                }
            }
    
    def _initialize_processors(self):
        """Inicializa os processadores de sinais de risco."""
        # Instancie os processadores conforme necessário
        # self.processors.append(LocationRiskProcessor())
        # ...
        pass
    
    def register_processor(self, processor: RiskProcessorInterface):
        """
        Registra um processador de sinais de risco.
        
        Args:
            processor: Processador a ser registrado
        """
        self.processors.append(processor)
    
    def evaluate_risk(self, auth_context: Dict[str, Any]) -> RiskAssessment:
        """
        Avalia o risco da tentativa de autenticação.
        
        Args:
            auth_context: Contexto da autenticação
            
        Returns:
            Avaliação de risco resultante
        """
        signals = []
        total_score = 0.0
        total_weight = 0.0
        
        # Processar cada sinal de risco
        for processor in self.processors:
            try:
                signal = processor.process(auth_context)
                signals.append(signal)
                
                # Ponderação do sinal
                weighted_signal = signal.value * signal.weight * signal.confidence
                total_score += weighted_signal
                total_weight += signal.weight
            except Exception as e:
                logger.error(f"Erro ao processar sinal: {str(e)}")
        
        # Calcular pontuação final
        if total_weight > 0:
            final_score = (total_score / total_weight) * 100
        else:
            final_score = 0
        
        # Determinar nível de risco
        risk_level = self._determine_risk_level(final_score)
        
        # Determinar fatores de autenticação necessários
        required_factors = self._determine_required_factors(risk_level)
        
        return RiskAssessment(
            level=risk_level,
            score=final_score,
            signals=signals,
            required_factors=required_factors
        )
    
    def _determine_risk_level(self, score: float) -> RiskLevel:
        """
        Determina o nível de risco com base na pontuação.
        
        Args:
            score: Pontuação de risco
            
        Returns:
            Nível de risco
        """
        thresholds = self.config.get("thresholds", {})
        
        if score >= thresholds.get("critical", 90):
            return RiskLevel.CRITICAL
        elif score >= thresholds.get("high", 75):
            return RiskLevel.HIGH
        elif score >= thresholds.get("medium", 50):
            return RiskLevel.MEDIUM
        else:
            return RiskLevel.LOW
    
    def _determine_required_factors(self, risk_level: RiskLevel) -> List[str]:
        """
        Determina os fatores de autenticação necessários com base no nível de risco.
        
        Args:
            risk_level: Nível de risco
            
        Returns:
            Lista de fatores de autenticação necessários
        """
        factor_requirements = self.config.get("factor_requirements", {})
        
        if risk_level == RiskLevel.CRITICAL:
            return factor_requirements.get("critical", ["password", "totp", "biometric_advanced"])
        elif risk_level == RiskLevel.HIGH:
            return factor_requirements.get("high", ["password", "totp", "push"])
        elif risk_level == RiskLevel.MEDIUM:
            return factor_requirements.get("medium", ["password", "totp"])
        else:
            return factor_requirements.get("low", ["password"])


# Implementações específicas de processadores de risco
class LocationRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado em localização."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base na localização.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        # Exemplo de lógica para avaliação de risco de localização
        location = context.get("location", {})
        ip_address = location.get("ip_address")
        country = location.get("country")
        city = location.get("city")
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.8  # Confiança padrão
        
        # Verificar se é uma localização conhecida
        known_locations = context.get("user_profile", {}).get("known_locations", [])
        location_known = False
        
        for known in known_locations:
            if known.get("country") == country and known.get("city") == city:
                location_known = True
                break
        
        if not location_known:
            risk_value += 0.6
        
        # Verificar se o país é de alto risco
        high_risk_countries = ["Unknown"]  # Lista de países de alto risco
        if country in high_risk_countries:
            risk_value += 0.8
            confidence = 0.9
        
        # Verificar se o IP está em listas de bloqueio
        if ip_address:
            # Simulação de verificação de IP em lista de bloqueio
            # Na implementação real, isso consultaria um serviço de reputação de IP
            if ip_address.startswith("10.") or ip_address.startswith("192.168."):
                risk_value += 0.0  # IP interno, baixo risco
            elif ip_address.startswith("185.") or ip_address.startswith("31."):
                risk_value += 0.7  # Simulação de IP suspeito
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="location_risk",
            category=RiskSignalCategory.LOCATION,
            weight=0.25,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "ip_address": ip_address,
                "country": country,
                "city": city,
                "known_location": location_known
            }
        )


class DeviceRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado em dispositivo."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base no dispositivo.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        device = context.get("device", {})
        device_id = device.get("id")
        os_info = device.get("os", "")
        browser = device.get("browser", "")
        is_mobile = device.get("is_mobile", False)
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.85  # Confiança padrão
        
        # Verificar se é um dispositivo conhecido
        known_devices = context.get("user_profile", {}).get("known_devices", [])
        device_known = any(d.get("id") == device_id for d in known_devices)
        
        if not device_known:
            risk_value += 0.7
        
        # Verificar se há indicadores de spoofing de dispositivo
        user_agent = device.get("user_agent", "")
        if user_agent:
            # Detecção simplificada de inconsistências no user agent
            if "iPhone" in user_agent and "Windows" in os_info:
                risk_value += 0.8
                confidence = 0.95
            elif "Android" in user_agent and "iOS" in os_info:
                risk_value += 0.8
                confidence = 0.95
        
        # Verificar se o dispositivo tem software de segurança
        has_security_software = device.get("has_security_software", False)
        if not has_security_software:
            risk_value += 0.3
        
        # Dispositivos móveis podem ter considerações específicas
        if is_mobile:
            if device.get("is_rooted", False) or device.get("is_jailbroken", False):
                risk_value += 0.6
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="device_risk",
            category=RiskSignalCategory.DEVICE,
            weight=0.2,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "device_id": device_id,
                "os": os_info,
                "browser": browser,
                "is_mobile": is_mobile,
                "known_device": device_known
            }
        )


class BehaviorRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado em comportamento do usuário."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base no comportamento do usuário.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        user_id = context.get("user_id", "")
        current_behavior = context.get("behavior", {})
        typing_pattern = current_behavior.get("typing_pattern", {})
        mouse_movement = current_behavior.get("mouse_movement", {})
        access_time = current_behavior.get("access_time")
        transaction_amount = current_behavior.get("transaction_amount")
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.75  # Confiança padrão para análise comportamental
        
        # Perfil comportamental do usuário
        user_profile = context.get("user_profile", {})
        usual_access_times = user_profile.get("usual_access_times", [])
        usual_transaction_amounts = user_profile.get("usual_transaction_amounts", [])
        usual_typing_patterns = user_profile.get("typing_patterns", [])
        
        # Avaliar padrão de acesso temporal
        if access_time and usual_access_times:
            current_hour = datetime.datetime.fromisoformat(access_time).hour
            if not any(abs(current_hour - usual.get("hour", 0)) <= 2 for usual in usual_access_times):
                risk_value += 0.4
        
        # Avaliar valores de transação
        if transaction_amount is not None and usual_transaction_amounts:
            avg_amount = sum(amt.get("amount", 0) for amt in usual_transaction_amounts) / len(usual_transaction_amounts)
            if transaction_amount > avg_amount * 3:  # Transação 3x maior que a média
                risk_value += 0.5
        
        # Avaliar padrão de digitação (simplificado)
        if typing_pattern and usual_typing_patterns:
            # Implementação simplificada - na prática, usaria algoritmos de ML para comparar padrões
            typing_match = False
            for pattern in usual_typing_patterns:
                # Simular comparação de padrões
                if pattern.get("speed", 0) == typing_pattern.get("speed", -1):
                    typing_match = True
                    break
            
            if not typing_match:
                risk_value += 0.6
                confidence = 0.85  # Maior confiança para biometria comportamental
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="behavior_risk",
            category=RiskSignalCategory.BEHAVIOR,
            weight=0.2,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "access_time_match": access_time and usual_access_times,
                "transaction_amount": transaction_amount,
                "typing_pattern_analyzed": bool(typing_pattern and usual_typing_patterns)
            }
        )


class NetworkRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado em rede."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base na rede.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        network = context.get("network", {})
        ip_address = network.get("ip")
        connection_type = network.get("connection_type", "unknown")
        vpn_detected = network.get("vpn_detected", False)
        tor_detected = network.get("tor_detected", False)
        proxy_detected = network.get("proxy_detected", False)
        asn = network.get("asn")
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.8  # Confiança padrão
        
        # Verificar uso de VPN/Tor/Proxy
        if vpn_detected or tor_detected or proxy_detected:
            risk_value += 0.6
            
            # Tor geralmente apresenta maior risco que VPN comum
            if tor_detected:
                risk_value += 0.2
                confidence = 0.95
        
        # Verificar tipo de conexão
        if connection_type == "public_wifi":
            risk_value += 0.4
        elif connection_type == "mobile":
            risk_value += 0.2
        
        # Verificar ASN (Autonomous System Number)
        # Na prática, teria uma lista de ASNs de alto risco
        high_risk_asns = ["64496", "64511"]  # ASNs fictícios para demonstração
        if asn in high_risk_asns:
            risk_value += 0.5
            confidence = 0.9
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="network_risk",
            category=RiskSignalCategory.NETWORK,
            weight=0.15,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "ip_address": ip_address,
                "connection_type": connection_type,
                "vpn_detected": vpn_detected,
                "tor_detected": tor_detected,
                "proxy_detected": proxy_detected,
                "asn": asn
            }
        )


class ResourceRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado no recurso acessado."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base no recurso sendo acessado.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        resource = context.get("resource", {})
        resource_type = resource.get("type", "unknown")
        resource_id = resource.get("id")
        sensitivity = resource.get("sensitivity", "low")
        operation = resource.get("operation", "read")
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.9  # Alta confiança para classificação de recursos
        
        # Verificar sensibilidade do recurso
        if sensitivity == "critical":
            risk_value += 0.9
        elif sensitivity == "high":
            risk_value += 0.7
        elif sensitivity == "medium":
            risk_value += 0.4
        else:  # "low"
            risk_value += 0.1
        
        # Verificar operação
        if operation == "delete":
            risk_value += 0.3
        elif operation == "write" or operation == "update":
            risk_value += 0.2
        
        # Verificar tipo de recurso
        high_risk_resources = ["financial", "medical", "personal_data", "admin"]
        if resource_type in high_risk_resources:
            risk_value += 0.4
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="resource_risk",
            category=RiskSignalCategory.RESOURCE,
            weight=0.1,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "resource_id": resource_id,
                "resource_type": resource_type,
                "sensitivity": sensitivity,
                "operation": operation
            }
        )


class TimeRiskProcessor(RiskProcessorInterface):
    """Processador de risco baseado em tempo."""
    
    def process(self, context: Dict[str, Any]) -> RiskSignal:
        """
        Avalia o risco com base no tempo de acesso.
        
        Args:
            context: Contexto da autenticação
            
        Returns:
            Sinal de risco processado
        """
        now = datetime.datetime.now()
        user_tz = context.get("user_timezone", "UTC")
        user_profile = context.get("user_profile", {})
        
        # Tentar obter a hora no fuso horário do usuário
        try:
            # Na implementação real, usaríamos pytz ou datetime com timezone
            # Para simplificar, assumimos que a hora atual já está no fuso do usuário
            current_hour = now.hour
        except Exception as e:
            logger.error(f"Erro ao calcular hora no fuso do usuário: {str(e)}")
            current_hour = now.hour
        
        # Valor de risco inicial
        risk_value = 0.0
        confidence = 0.7  # Confiança média para análise de tempo
        
        # Verificar se é horário de trabalho (simplificado)
        is_business_hours = 9 <= current_hour <= 18
        is_business_day = now.weekday() < 5  # Segunda a Sexta
        
        # Obter padrões habituais de acesso do usuário
        usual_hours = user_profile.get("usual_access_hours", [])
        
        # Verificar se o acesso está fora dos padrões habituais do usuário
        unusual_time = True
        if usual_hours:
            for hour_range in usual_hours:
                start = hour_range.get("start", 0)
                end = hour_range.get("end", 0)
                if start <= current_hour <= end:
                    unusual_time = False
                    break
        else:
            # Se não houver dados sobre padrões habituais, usar horário comercial
            if is_business_hours and is_business_day:
                unusual_time = False
        
        # Avaliação de risco com base no tempo
        if unusual_time:
            # Maior risco para acessos noturnos
            if current_hour >= 22 or current_hour <= 5:
                risk_value += 0.7
            else:
                risk_value += 0.4
        
        # Verificar se é um dia especial (feriado, fim de semana)
        if not is_business_day:
            risk_value += 0.3
        
        # Limitar valor de risco entre 0 e 1
        risk_value = min(max(risk_value, 0.0), 1.0)
        
        return RiskSignal(
            name="time_risk",
            category=RiskSignalCategory.TIME,
            weight=0.1,  # Peso conforme configuração padrão
            value=risk_value,
            confidence=confidence,
            metadata={
                "current_hour": current_hour,
                "is_business_hours": is_business_hours,
                "is_business_day": is_business_day,
                "unusual_time": unusual_time
            }
        )


# Integração com OpenTelemetry para observabilidade
try:
    from opentelemetry import trace, metrics
    from opentelemetry.trace import Status, StatusCode
    
    HAS_TELEMETRY = True
except ImportError:
    HAS_TELEMETRY = False


class TelemetryEnabledRiskEngine(RiskEngine):
    """Versão do motor de risco com suporte a telemetria."""
    
    def __init__(self, config_path: Optional[Path] = None):
        """
        Inicializa o motor de avaliação de risco com telemetria.
        
        Args:
            config_path: Caminho para arquivo de configuração
        """
        super().__init__(config_path)
        
        if HAS_TELEMETRY:
            self.tracer = trace.get_tracer("innovabiz.iam.risk_engine")
            self.meter = metrics.get_meter("innovabiz.iam.risk_engine")
            
            # Métricas
            self.risk_assessment_counter = self.meter.create_counter(
                name="risk_assessments",
                description="Número de avaliações de risco realizadas",
                unit="1"
            )
            
            self.risk_level_counter = self.meter.create_counter(
                name="risk_levels",
                description="Distribuição de níveis de risco",
                unit="1"
            )
            
            self.risk_score_histogram = self.meter.create_histogram(
                name="risk_scores",
                description="Distribuição de pontuações de risco",
                unit="1"
            )
    
    def evaluate_risk(self, auth_context: Dict[str, Any]) -> RiskAssessment:
        """
        Avalia o risco da tentativa de autenticação com telemetria.
        
        Args:
            auth_context: Contexto da autenticação
            
        Returns:
            Avaliação de risco resultante
        """
        if not HAS_TELEMETRY:
            return super().evaluate_risk(auth_context)
        
        with self.tracer.start_as_current_span("evaluate_risk") as span:
            try:
                # Adicionar atributos ao span
                user_id = auth_context.get("user_id", "unknown")
                span.set_attribute("user_id", user_id)
                span.set_attribute("auth_context.source_ip", auth_context.get("location", {}).get("ip_address", "unknown"))
                span.set_attribute("auth_context.resource_type", auth_context.get("resource", {}).get("type", "unknown"))
                
                # Executar avaliação de risco
                assessment = super().evaluate_risk(auth_context)
                
                # Registrar métricas
                self.risk_assessment_counter.add(1, {"user_id": user_id})
                self.risk_level_counter.add(1, {"risk_level": assessment.level.name})
                self.risk_score_histogram.record(assessment.score)
                
                # Adicionar mais atributos ao span com o resultado
                span.set_attribute("risk_level", assessment.level.name)
                span.set_attribute("risk_score", assessment.score)
                span.set_status(Status(StatusCode.OK))
                
                return assessment
            except Exception as e:
                span.record_exception(e)
                span.set_status(Status(StatusCode.ERROR, str(e)))
                raise


# Fábrica para criação do motor de risco
def create_risk_engine(config_path: Optional[Path] = None, enable_telemetry: bool = True) -> RiskEngine:
    """
    Cria uma instância do motor de avaliação de risco.
    
    Args:
        config_path: Caminho para arquivo de configuração
        enable_telemetry: Se deve habilitar telemetria
    
    Returns:
        Instância do motor de avaliação de risco
    """
    if enable_telemetry and HAS_TELEMETRY:
        engine = TelemetryEnabledRiskEngine(config_path)
    else:
        engine = RiskEngine(config_path)
    
    # Registrar processadores padrão
    engine.register_processor(LocationRiskProcessor())
    engine.register_processor(DeviceRiskProcessor())
    engine.register_processor(BehaviorRiskProcessor())
    engine.register_processor(NetworkRiskProcessor())
    engine.register_processor(ResourceRiskProcessor())
    engine.register_processor(TimeRiskProcessor())
    
    return engine
