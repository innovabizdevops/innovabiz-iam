#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Produtores de Eventos para Sinais de Fraude

Este módulo implementa os produtores de eventos para sinais de fraude
que serão enviados para o Kafka e processados em tempo real.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import time
import uuid
import logging
import datetime
from typing import Dict, Any, List, Optional, Union
from enum import Enum
from dataclasses import dataclass, asdict, field
from contextlib import contextmanager
from confluent_kafka import Producer, KafkaException

# Importação de módulos específicos do projeto
try:
    from infrastructure.fraud_detection.kafka_config import (
        load_config_from_env, 
        load_config_from_file, 
        get_default_config,
        FraudDetectionKafkaConfig
    )
except ImportError:
    # Fallback para testes ou uso isolado
    import os
    import sys
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
    from fraud_detection.kafka_config import (
        load_config_from_env, 
        load_config_from_file, 
        get_default_config,
        FraudDetectionKafkaConfig
    )

# Configuração de logging
logger = logging.getLogger("fraud_detection_producers")
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)


class EventSeverity(str, Enum):
    """Severidade do evento de fraude."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class EventType(str, Enum):
    """Tipo de evento de fraude."""
    DOCUMENT_VALIDATION = "document_validation"
    AUTHENTICATION = "authentication"
    TRANSACTION = "transaction"
    SESSION = "session"
    BEHAVIOR = "behavior"
    REGIONAL_PATTERN = "regional_pattern"
    ALERT = "alert"


@dataclass
class EventMetadata:
    """Metadados comuns a todos os eventos."""
    event_id: str = field(default_factory=lambda: str(uuid.uuid4()))
    timestamp: str = field(default_factory=lambda: datetime.datetime.now().isoformat())
    region_code: str = ""
    tenant_id: str = ""
    user_id: Optional[str] = None
    session_id: Optional[str] = None
    request_id: Optional[str] = None
    source_system: str = "innovabiz-trustguard"
    source_module: str = "fraud_detection"
    environment: str = "production"
    version: str = "1.0.0"
    gdpr_compliant: bool = True
    additional_info: Dict[str, Any] = field(default_factory=dict)


@dataclass
class DocumentValidationEvent:
    """Evento de validação documental."""
    metadata: EventMetadata
    document_type: str
    document_number: str
    validation_result: bool
    confidence_score: float
    country_code: str
    validation_steps: List[Dict[str, Any]]
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    risk_level: str = "low"
    fraud_signals: List[Dict[str, Any]] = field(default_factory=list)
    verification_service: Optional[str] = None
    verification_timestamp: Optional[str] = None
    check_reference: Optional[str] = None
    document_issuer: Optional[str] = None
    document_issue_date: Optional[str] = None
    document_expiry_date: Optional[str] = None
    holder_name: Optional[str] = None
    
    # Campos anonimizados ou criptografados para conformidade GDPR
    anonymized_fields: List[str] = field(default_factory=list)


@dataclass
class AuthenticationEvent:
    """Evento de autenticação."""
    metadata: EventMetadata
    authentication_result: bool
    authentication_method: str
    ip_address: str
    device_id: Optional[str] = None
    geolocation: Optional[Dict[str, Any]] = None
    auth_factors: List[str] = field(default_factory=list)
    risk_score: float = 0.0
    risk_level: str = "low"
    failure_reason: Optional[str] = None
    attempt_number: int = 1
    auth_patterns: Dict[str, Any] = field(default_factory=dict)
    mfa_used: bool = False
    mfa_type: Optional[str] = None
    biometrics_used: bool = False
    biometrics_type: Optional[str] = None
    anomaly_signals: List[Dict[str, Any]] = field(default_factory=list)


@dataclass
class TransactionEvent:
    """Evento de transação."""
    metadata: EventMetadata
    transaction_id: str
    transaction_type: str
    transaction_amount: float
    transaction_currency: str
    merchant_id: Optional[str] = None
    merchant_name: Optional[str] = None
    merchant_category: Optional[str] = None
    payment_method: Optional[str] = None
    payment_channel: Optional[str] = None
    destination_account: Optional[str] = None
    risk_score: float = 0.0
    risk_level: str = "low"
    fraud_signals: List[Dict[str, Any]] = field(default_factory=list)
    is_cross_border: bool = False
    is_high_value: bool = False
    is_high_frequency: bool = False
    velocity_metrics: Dict[str, Any] = field(default_factory=dict)
    aml_signals: Dict[str, Any] = field(default_factory=dict)
    transaction_status: str = "pending"


@dataclass
class SessionEvent:
    """Evento de sessão."""
    metadata: EventMetadata
    session_id: str
    ip_address: str
    device_info: Dict[str, Any]
    geolocation: Optional[Dict[str, Any]] = None
    user_agent: Optional[str] = None
    session_start: str = field(default_factory=lambda: datetime.datetime.now().isoformat())
    session_end: Optional[str] = None
    session_duration: Optional[int] = None
    session_events: List[Dict[str, Any]] = field(default_factory=list)
    risk_score: float = 0.0
    risk_level: str = "low"
    fraud_signals: List[Dict[str, Any]] = field(default_factory=list)
    session_type: str = "normal"
    is_mobile: bool = False
    is_new_device: bool = False
    is_suspicious_location: bool = False
    behavioral_metrics: Dict[str, Any] = field(default_factory=dict)


@dataclass
class BehaviorEvent:
    """Evento de comportamento."""
    metadata: EventMetadata
    behavior_type: str
    behavior_data: Dict[str, Any]
    confidence_score: float
    risk_score: float = 0.0
    risk_level: str = "low"
    abnormal_patterns: List[Dict[str, Any]] = field(default_factory=list)
    behavioral_profile_id: Optional[str] = None
    triggered_rules: List[str] = field(default_factory=list)
    region_specific_flags: Dict[str, Any] = field(default_factory=dict)
    user_journey_stage: Optional[str] = None
    activity_context: Dict[str, Any] = field(default_factory=dict)


@dataclass
class AlertEvent:
    """Evento de alerta."""
    metadata: EventMetadata
    alert_id: str
    alert_type: str
    alert_severity: EventSeverity
    source_events: List[Dict[str, str]]
    alert_message: str
    alert_details: Dict[str, Any]
    triggered_rules: List[str] = field(default_factory=list)
    requires_action: bool = False
    assigned_to: Optional[str] = None
    escalation_level: int = 0
    time_to_respond: Optional[int] = None  # Em segundos
    related_cases: List[str] = field(default_factory=list)
    recommended_actions: List[str] = field(default_factory=list)
    resolution_status: str = "pending"


@dataclass
class RegionalPatternEvent:
    """Evento de padrão regional de fraude."""
    metadata: EventMetadata
    pattern_id: str
    pattern_name: str
    region_code: str
    pattern_type: str
    pattern_confidence: float
    detected_indicators: List[Dict[str, Any]]
    pattern_frequency: Dict[str, Any]
    risk_level: str = "medium"
    affected_services: List[str] = field(default_factory=list)
    pattern_duration: Dict[str, Any] = field(default_factory=dict)
    remediation_suggestions: List[str] = field(default_factory=list)
    is_emerging_threat: bool = False
    correlation_data: Dict[str, Any] = field(default_factory=dict)


class EventProducer:
    """
    Produtor de eventos para o sistema de detecção de fraudes.
    
    Esta classe é responsável por enviar eventos para o Kafka para processamento
    em tempo real por consumidores especializados.
    """
    
    def __init__(self, config_file: Optional[str] = None):
        """
        Inicializa o produtor de eventos com a configuração especificada.
        
        Args:
            config_file: Caminho para o arquivo de configuração. Se None,
                        carrega a configuração das variáveis de ambiente.
        """
        if config_file:
            self.config = load_config_from_file(config_file)
            if not self.config:
                logger.warning("Falha ao carregar configuração do arquivo. Usando padrão.")
                self.config = get_default_config()
        else:
            try:
                self.config = load_config_from_env()
            except Exception as e:
                logger.warning(f"Falha ao carregar configuração de ambiente: {str(e)}. Usando padrão.")
                self.config = get_default_config()
                
        self.producer = None
        self.initialize_producer()
        
    def initialize_producer(self):
        """Inicializa o produtor Kafka com as configurações carregadas."""
        try:
            # Configuração do produtor Kafka
            producer_config = {
                'bootstrap.servers': ','.join(self.config.cluster.bootstrap_servers),
                'client.id': self.config.cluster.client_id,
                'security.protocol': self.config.cluster.security_protocol,
                'sasl.mechanism': self.config.cluster.sasl_mechanism,
                'sasl.username': self.config.cluster.sasl_username,
                'sasl.password': self.config.cluster.sasl_password,
                'acks': self.config.producer.acks,
                'enable.idempotence': str(self.config.producer.enable_idempotence).lower(),
                'max.in.flight.requests.per.connection': self.config.producer.max_in_flight_requests_per_connection,
                'linger.ms': self.config.cluster.linger_ms,
                'compression.type': self.config.cluster.compression_type,
                'batch.size': self.config.cluster.batch_size,
                'delivery.timeout.ms': self.config.producer.delivery_timeout_ms,
                'request.timeout.ms': self.config.producer.request_timeout_ms,
                'transaction.timeout.ms': self.config.producer.transaction_timeout_ms,
                'retries': self.config.cluster.max_retries,
                'retry.backoff.ms': self.config.cluster.retry_backoff_ms
            }
            
            # Criar produtor
            self.producer = Producer(producer_config)
            logger.info(f"Produtor Kafka inicializado com sucesso para {self.config.cluster.bootstrap_servers}")
        except Exception as e:
            logger.error(f"Erro ao inicializar produtor Kafka: {str(e)}")
            raise
    
    def delivery_report(self, err, msg):
        """
        Callback chamado para cada mensagem produzida para indicar se houve erro ou sucesso.
        
        Args:
            err: Erro (se houver)
            msg: Mensagem produzida
        """
        if err is not None:
            logger.error(f"Falha na entrega da mensagem: {str(err)}")
        else:
            topic = msg.topic()
            partition = msg.partition()
            offset = msg.offset()
            key = msg.key().decode('utf-8') if msg.key() else None
            logger.debug(f"Mensagem entregue com sucesso: tópico={topic}, partição={partition}, offset={offset}, key={key}")
    
    def get_topic_name(self, event_type: EventType, region_code: str = None) -> str:
        """
        Obtém o nome do tópico Kafka para o tipo de evento e região especificados.
        
        Args:
            event_type: Tipo do evento
            region_code: Código da região (opcional)
            
        Returns:
            str: Nome do tópico Kafka
        """
        topics = self.config.topics
        topic_map = {
            EventType.DOCUMENT_VALIDATION: topics.document_validation_events.name,
            EventType.AUTHENTICATION: topics.authentication_events.name,
            EventType.TRANSACTION: topics.transaction_events.name,
            EventType.SESSION: topics.session_events.name,
            EventType.BEHAVIOR: topics.behavior_events.name,
            EventType.REGIONAL_PATTERN: topics.regional_fraud_patterns.name,
            EventType.ALERT: topics.alert_events.name
        }
        
        base_topic = topic_map.get(event_type)
        if not base_topic:
            raise ValueError(f"Tipo de evento não suportado: {event_type}")
        
        # Se a região for especificada, usar o prefixo regional correspondente
        if region_code and region_code in self.config.regional_configs:
            region_prefix = self.config.regional_configs[region_code].topics_prefix
            return f"{region_prefix}{base_topic}"
        
        return base_topic
    
    @contextmanager
    def transaction(self):
        """
        Gerenciador de contexto para transações Kafka.
        
        Permite o envio de múltiplos eventos em uma única transação atômica.
        """
        try:
            self.producer.init_transactions()
            self.producer.begin_transaction()
            yield
            self.producer.commit_transaction()
        except KafkaException as e:
            logger.error(f"Erro na transação Kafka: {str(e)}")
            self.producer.abort_transaction()
            raise
    
    def produce_event(self, event: Union[DocumentValidationEvent, AuthenticationEvent, 
                                         TransactionEvent, SessionEvent, BehaviorEvent, 
                                         AlertEvent, RegionalPatternEvent],
                     event_type: EventType, key: Optional[str] = None):
        """
        Produz um evento no tópico Kafka correspondente.
        
        Args:
            event: Objeto de evento a ser enviado
            event_type: Tipo do evento
            key: Chave para particionamento (opcional)
        """
        try:
            # Converter o evento para dicionário
            event_dict = asdict(event)
            
            # Obter o nome do tópico
            region_code = event.metadata.region_code if hasattr(event.metadata, 'region_code') else None
            topic_name = self.get_topic_name(event_type, region_code)
            
            # Serializar o evento
            event_json = json.dumps(event_dict, default=str).encode('utf-8')
            
            # Definir a chave se não fornecida
            if key is None:
                key = event.metadata.event_id
            
            # Produzir a mensagem
            self.producer.produce(
                topic=topic_name,
                key=key.encode('utf-8') if key else None,
                value=event_json,
                callback=self.delivery_report
            )
            
            # Tentar enviar as mensagens imediatamente
            self.producer.poll(0)
            
        except Exception as e:
            logger.error(f"Erro ao produzir evento: {str(e)}")
            # Salvar no dead-letter-queue
            self._save_to_dlq(event, event_type, str(e))
            raise
    
    def produce_bulk_events(self, events: List[tuple]):
        """
        Produz múltiplos eventos em uma única transação.
        
        Args:
            events: Lista de tuplas (evento, event_type, key) a serem enviados
        """
        with self.transaction():
            for event, event_type, key in events:
                self.produce_event(event, event_type, key)
    
    def flush(self, timeout: int = 30):
        """
        Garante que todas as mensagens pendentes sejam enviadas para o broker.
        
        Args:
            timeout: Tempo máximo de espera em segundos
        """
        self.producer.flush(timeout)
    
    def _save_to_dlq(self, event, event_type, error_message):
        """
        Salva um evento que falhou no dead-letter-queue.
        
        Args:
            event: Evento original
            event_type: Tipo do evento
            error_message: Mensagem de erro
        """
        try:
            # Converter o evento para dicionário
            event_dict = asdict(event)
            
            # Adicionar informações de erro
            dlq_payload = {
                "original_event": event_dict,
                "event_type": event_type.value if isinstance(event_type, EventType) else event_type,
                "error_message": error_message,
                "timestamp": datetime.datetime.now().isoformat(),
                "retry_count": 0
            }
            
            # Serializar o payload
            dlq_json = json.dumps(dlq_payload, default=str).encode('utf-8')
            
            # Produzir a mensagem no DLQ
            self.producer.produce(
                topic=self.config.topics.dead_letter_queue.name,
                value=dlq_json,
                callback=self.delivery_report
            )
            
            # Tentar enviar as mensagens imediatamente
            self.producer.poll(0)
            
        except Exception as e:
            logger.error(f"Erro ao salvar no DLQ: {str(e)}")
    
    def close(self):
        """Fecha o produtor e libera os recursos."""
        if self.producer:
            self.flush()
            # Não há método close() para o produtor da confluent_kafka,
            # o garbage collector vai lidar com isso


# Funções auxiliares para criar eventos

def create_document_validation_event(
    document_type: str,
    document_number: str,
    validation_result: bool,
    confidence_score: float,
    country_code: str,
    validation_steps: List[Dict[str, Any]],
    region_code: str,
    tenant_id: str,
    user_id: Optional[str] = None,
    errors: List[str] = None,
    warnings: List[str] = None,
    risk_level: str = "low",
    **kwargs
) -> DocumentValidationEvent:
    """
    Cria um evento de validação documental pronto para envio.
    
    Args:
        document_type: Tipo de documento (ex: "nif", "cc", "passport")
        document_number: Número do documento
        validation_result: Resultado da validação (True/False)
        confidence_score: Pontuação de confiança (0.0-1.0)
        country_code: Código do país do documento
        validation_steps: Lista de etapas de validação e seus resultados
        region_code: Código da região (AO, BR, MZ, PT)
        tenant_id: ID do tenant
        user_id: ID do usuário (opcional)
        errors: Lista de erros encontrados (opcional)
        warnings: Lista de avisos (opcional)
        risk_level: Nível de risco (low, medium, high, critical)
        **kwargs: Parâmetros adicionais
        
    Returns:
        DocumentValidationEvent: Evento pronto para envio
    """
    metadata = EventMetadata(
        region_code=region_code,
        tenant_id=tenant_id,
        user_id=user_id,
        source_module="document_validation",
        additional_info=kwargs.get("additional_info", {})
    )
    
    return DocumentValidationEvent(
        metadata=metadata,
        document_type=document_type,
        document_number=document_number,
        validation_result=validation_result,
        confidence_score=confidence_score,
        country_code=country_code,
        validation_steps=validation_steps,
        errors=errors or [],
        warnings=warnings or [],
        risk_level=risk_level,
        fraud_signals=kwargs.get("fraud_signals", []),
        verification_service=kwargs.get("verification_service"),
        verification_timestamp=kwargs.get("verification_timestamp"),
        check_reference=kwargs.get("check_reference"),
        document_issuer=kwargs.get("document_issuer"),
        document_issue_date=kwargs.get("document_issue_date"),
        document_expiry_date=kwargs.get("document_expiry_date"),
        holder_name=kwargs.get("holder_name"),
        anonymized_fields=kwargs.get("anonymized_fields", [])
    )


def create_authentication_event(
    authentication_result: bool,
    authentication_method: str,
    ip_address: str,
    region_code: str,
    tenant_id: str,
    user_id: Optional[str] = None,
    device_id: Optional[str] = None,
    risk_score: float = 0.0,
    **kwargs
) -> AuthenticationEvent:
    """
    Cria um evento de autenticação pronto para envio.
    
    Args:
        authentication_result: Resultado da autenticação (True/False)
        authentication_method: Método utilizado (ex: "password", "biometric", "mfa")
        ip_address: Endereço IP do cliente
        region_code: Código da região (AO, BR, MZ, PT)
        tenant_id: ID do tenant
        user_id: ID do usuário (opcional)
        device_id: ID do dispositivo (opcional)
        risk_score: Pontuação de risco (0.0-1.0)
        **kwargs: Parâmetros adicionais
        
    Returns:
        AuthenticationEvent: Evento pronto para envio
    """
    metadata = EventMetadata(
        region_code=region_code,
        tenant_id=tenant_id,
        user_id=user_id,
        session_id=kwargs.get("session_id"),
        source_module="authentication",
        additional_info=kwargs.get("additional_info", {})
    )
    
    # Determinar o nível de risco com base na pontuação
    if risk_score >= 0.8:
        risk_level = "critical"
    elif risk_score >= 0.6:
        risk_level = "high"
    elif risk_score >= 0.3:
        risk_level = "medium"
    else:
        risk_level = "low"
    
    return AuthenticationEvent(
        metadata=metadata,
        authentication_result=authentication_result,
        authentication_method=authentication_method,
        ip_address=ip_address,
        device_id=device_id,
        geolocation=kwargs.get("geolocation"),
        auth_factors=kwargs.get("auth_factors", []),
        risk_score=risk_score,
        risk_level=risk_level,
        failure_reason=kwargs.get("failure_reason"),
        attempt_number=kwargs.get("attempt_number", 1),
        auth_patterns=kwargs.get("auth_patterns", {}),
        mfa_used=kwargs.get("mfa_used", False),
        mfa_type=kwargs.get("mfa_type"),
        biometrics_used=kwargs.get("biometrics_used", False),
        biometrics_type=kwargs.get("biometrics_type"),
        anomaly_signals=kwargs.get("anomaly_signals", [])
    )


# Exemplo de uso do produtor
def example_usage():
    """Exemplo de uso do produtor de eventos."""
    try:
        # Inicializar o produtor
        producer = EventProducer()
        
        # Criar evento de validação documental
        doc_event = create_document_validation_event(
            document_type="nif",
            document_number="123456789",
            validation_result=True,
            confidence_score=0.95,
            country_code="PT",
            validation_steps=[
                {"name": "format_check", "result": True, "details": "Formato válido"},
                {"name": "checksum_validation", "result": True, "details": "Dígito verificador válido"},
                {"name": "external_validation", "result": True, "details": "Verificado na autoridade tributária"}
            ],
            region_code="PT",
            tenant_id="tenant123",
            user_id="user456",
            verification_service="ServiçoFinançasPortugal",
            holder_name="Maria Silva",
            anonymized_fields=["holder_name"]
        )
        
        # Produzir o evento
        producer.produce_event(doc_event, EventType.DOCUMENT_VALIDATION)
        
        # Garantir que o evento foi enviado
        producer.flush()
        logger.info("Evento de validação documental enviado com sucesso")
        
        # Criar evento de autenticação
        auth_event = create_authentication_event(
            authentication_result=False,
            authentication_method="password",
            ip_address="192.168.1.1",
            region_code="PT",
            tenant_id="tenant123",
            user_id="user456",
            device_id="device789",
            risk_score=0.7,
            geolocation={"country": "Portugal", "city": "Lisboa", "coordinates": {"lat": 38.7223, "lon": -9.1393}},
            auth_factors=["password"],
            failure_reason="Senha incorreta",
            attempt_number=3,
            mfa_used=False
        )
        
        # Produzir o evento
        producer.produce_event(auth_event, EventType.AUTHENTICATION)
        
        # Garantir que o evento foi enviado
        producer.flush()
        logger.info("Evento de autenticação enviado com sucesso")
        
        # Fechar o produtor
        producer.close()
        
    except Exception as e:
        logger.error(f"Erro no exemplo de uso: {str(e)}")


if __name__ == "__main__":
    # Executar exemplo de uso
    example_usage()