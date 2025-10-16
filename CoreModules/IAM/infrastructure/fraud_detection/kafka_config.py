#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Configuração de Infraestrutura Kafka para Processamento de Eventos de Fraude

Este módulo configura a infraestrutura Kafka necessária para o processamento
de eventos relacionados a detecção de fraude em tempo real na plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field

# Configuração de logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("kafka_fraud_detection")

@dataclass
class KafkaClusterConfig:
    """Configuração do cluster Kafka."""
    bootstrap_servers: List[str]
    security_protocol: str = "SASL_SSL"
    sasl_mechanism: str = "PLAIN"
    sasl_username: str = ""
    sasl_password: str = ""
    ssl_cafile: Optional[str] = None
    client_id: str = "innovabiz-fraud-detection"
    
    # Configurações de retry
    max_retries: int = 3
    retry_backoff_ms: int = 500
    
    # Configurações de performance
    batch_size: int = 16384
    linger_ms: int = 5
    compression_type: str = "snappy"
    
    # Métricas e monitoramento
    metrics_enabled: bool = True
    metrics_polling_interval_ms: int = 30000


@dataclass
class KafkaTopicConfig:
    """Configuração de tópico Kafka."""
    name: str
    partitions: int
    replication_factor: int
    retention_ms: int = 604800000  # 7 dias em milissegundos
    cleanup_policy: str = "delete"
    segment_bytes: int = 1073741824  # 1GB
    min_insync_replicas: int = 2
    max_message_bytes: int = 1000000
    

@dataclass
class FraudDetectionTopics:
    """Configuração dos tópicos específicos para detecção de fraude."""
    document_validation_events: KafkaTopicConfig
    authentication_events: KafkaTopicConfig
    transaction_events: KafkaTopicConfig
    session_events: KafkaTopicConfig
    alert_events: KafkaTopicConfig
    behavior_events: KafkaTopicConfig
    regional_fraud_patterns: KafkaTopicConfig
    dead_letter_queue: KafkaTopicConfig
    

@dataclass
class KafkaConsumerConfig:
    """Configuração de consumidor Kafka."""
    group_id: str
    auto_offset_reset: str = "earliest"
    enable_auto_commit: bool = False
    max_poll_records: int = 500
    max_poll_interval_ms: int = 300000  # 5 minutos
    session_timeout_ms: int = 30000
    heartbeat_interval_ms: int = 10000
    isolation_level: str = "read_committed"
    fetch_min_bytes: int = 1
    fetch_max_bytes: int = 52428800  # 50MB
    fetch_max_wait_ms: int = 500


@dataclass
class KafkaProducerConfig:
    """Configuração de produtor Kafka."""
    acks: str = "all"
    enable_idempotence: bool = True
    max_in_flight_requests_per_connection: int = 5
    delivery_timeout_ms: int = 120000  # 2 minutos
    request_timeout_ms: int = 30000
    transaction_timeout_ms: int = 60000  # 1 minuto


@dataclass
class StreamsConfig:
    """Configuração para Kafka Streams."""
    application_id: str
    state_dir: str
    commit_interval_ms: int = 30000
    cache_max_bytes_buffering: int = 10485760  # 10MB
    processing_guarantee: str = "exactly_once_v2"
    num_stream_threads: int = 4
    replication_factor: int = 3


@dataclass
class RegionalConfig:
    """Configuração específica para cada região."""
    region_code: str
    topics_prefix: str
    regional_bootstrap_servers: Optional[List[str]] = None
    regional_consumer_group_suffix: str = ""
    schema_registry_url: Optional[str] = None
    rate_limit_events_per_second: int = 1000
    regional_stream_store_name: str = ""
    additional_settings: Dict[str, Any] = field(default_factory=dict)


@dataclass
class FraudDetectionKafkaConfig:
    """Configuração completa do Kafka para detecção de fraude."""
    cluster: KafkaClusterConfig
    topics: FraudDetectionTopics
    consumer: KafkaConsumerConfig
    producer: KafkaProducerConfig
    streams: StreamsConfig
    regional_configs: Dict[str, RegionalConfig] = field(default_factory=dict)


def load_config_from_env() -> FraudDetectionKafkaConfig:
    """
    Carrega a configuração Kafka a partir de variáveis de ambiente.
    
    Returns:
        FraudDetectionKafkaConfig: Configuração completa para o sistema Kafka
    """
    # Carregar variáveis de ambiente
    bootstrap_servers = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092").split(",")
    sasl_username = os.getenv("KAFKA_SASL_USERNAME", "")
    sasl_password = os.getenv("KAFKA_SASL_PASSWORD", "")
    ssl_cafile = os.getenv("KAFKA_SSL_CAFILE", None)
    
    # Configuração do cluster
    cluster_config = KafkaClusterConfig(
        bootstrap_servers=bootstrap_servers,
        sasl_username=sasl_username,
        sasl_password=sasl_password,
        ssl_cafile=ssl_cafile,
        client_id=os.getenv("KAFKA_CLIENT_ID", "innovabiz-fraud-detection"),
        security_protocol=os.getenv("KAFKA_SECURITY_PROTOCOL", "SASL_SSL"),
        sasl_mechanism=os.getenv("KAFKA_SASL_MECHANISM", "PLAIN"),
    )
    
    # Configuração de tópicos
    replication_factor = int(os.getenv("KAFKA_REPLICATION_FACTOR", "3"))
    topics_config = FraudDetectionTopics(
        document_validation_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_DOCUMENT_VALIDATION", "document-validation-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_DOCUMENT_VALIDATION_PARTITIONS", "8")),
            replication_factor=replication_factor
        ),
        authentication_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_AUTHENTICATION", "authentication-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_AUTHENTICATION_PARTITIONS", "8")),
            replication_factor=replication_factor
        ),
        transaction_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_TRANSACTION", "transaction-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_TRANSACTION_PARTITIONS", "16")),
            replication_factor=replication_factor
        ),
        session_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_SESSION", "session-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_SESSION_PARTITIONS", "8")),
            replication_factor=replication_factor
        ),
        alert_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_ALERT", "alert-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_ALERT_PARTITIONS", "4")),
            replication_factor=replication_factor
        ),
        behavior_events=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_BEHAVIOR", "behavior-events"),
            partitions=int(os.getenv("KAFKA_TOPIC_BEHAVIOR_PARTITIONS", "8")),
            replication_factor=replication_factor
        ),
        regional_fraud_patterns=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_REGIONAL_PATTERNS", "regional-fraud-patterns"),
            partitions=int(os.getenv("KAFKA_TOPIC_REGIONAL_PATTERNS_PARTITIONS", "4")),
            replication_factor=replication_factor
        ),
        dead_letter_queue=KafkaTopicConfig(
            name=os.getenv("KAFKA_TOPIC_DLQ", "fraud-detection-dlq"),
            partitions=int(os.getenv("KAFKA_TOPIC_DLQ_PARTITIONS", "4")),
            replication_factor=replication_factor,
            retention_ms=int(os.getenv("KAFKA_TOPIC_DLQ_RETENTION_MS", "2592000000"))  # 30 dias
        )
    )
    
    # Configuração de consumidor
    consumer_config = KafkaConsumerConfig(
        group_id=os.getenv("KAFKA_CONSUMER_GROUP_ID", "fraud-detection-group"),
        auto_offset_reset=os.getenv("KAFKA_CONSUMER_AUTO_OFFSET_RESET", "earliest"),
        enable_auto_commit=os.getenv("KAFKA_CONSUMER_ENABLE_AUTO_COMMIT", "False").lower() == "true",
        max_poll_records=int(os.getenv("KAFKA_CONSUMER_MAX_POLL_RECORDS", "500")),
        max_poll_interval_ms=int(os.getenv("KAFKA_CONSUMER_MAX_POLL_INTERVAL_MS", "300000")),
    )
    
    # Configuração de produtor
    producer_config = KafkaProducerConfig(
        acks=os.getenv("KAFKA_PRODUCER_ACKS", "all"),
        enable_idempotence=os.getenv("KAFKA_PRODUCER_ENABLE_IDEMPOTENCE", "True").lower() == "true",
        max_in_flight_requests_per_connection=int(os.getenv("KAFKA_PRODUCER_MAX_IN_FLIGHT", "5")),
    )
    
    # Configuração de Streams
    streams_config = StreamsConfig(
        application_id=os.getenv("KAFKA_STREAMS_APP_ID", "fraud-detection-streams"),
        state_dir=os.getenv("KAFKA_STREAMS_STATE_DIR", "/tmp/kafka-streams"),
        commit_interval_ms=int(os.getenv("KAFKA_STREAMS_COMMIT_INTERVAL_MS", "30000")),
        num_stream_threads=int(os.getenv("KAFKA_STREAMS_NUM_THREADS", "4")),
    )
    
    # Configurações regionais
    regional_configs = {}
    
    # Carregar configurações para Angola
    regional_configs["AO"] = RegionalConfig(
        region_code="AO",
        topics_prefix="ao-",
        regional_consumer_group_suffix="-angola",
        regional_stream_store_name="angola-fraud-store",
        rate_limit_events_per_second=int(os.getenv("KAFKA_RATE_LIMIT_ANGOLA", "500"))
    )
    
    # Carregar configurações para Brasil
    regional_configs["BR"] = RegionalConfig(
        region_code="BR",
        topics_prefix="br-",
        regional_consumer_group_suffix="-brasil",
        regional_stream_store_name="brasil-fraud-store",
        rate_limit_events_per_second=int(os.getenv("KAFKA_RATE_LIMIT_BRASIL", "1000"))
    )
    
    # Carregar configurações para Moçambique
    regional_configs["MZ"] = RegionalConfig(
        region_code="MZ",
        topics_prefix="mz-",
        regional_consumer_group_suffix="-mocambique",
        regional_stream_store_name="mocambique-fraud-store",
        rate_limit_events_per_second=int(os.getenv("KAFKA_RATE_LIMIT_MOCAMBIQUE", "300"))
    )
    
    # Carregar configurações para Portugal
    regional_configs["PT"] = RegionalConfig(
        region_code="PT",
        topics_prefix="pt-",
        regional_consumer_group_suffix="-portugal",
        regional_stream_store_name="portugal-fraud-store",
        rate_limit_events_per_second=int(os.getenv("KAFKA_RATE_LIMIT_PORTUGAL", "800")),
        additional_settings={
            "gdpr_compliant": True,
            "data_retention_days": 30,
            "anonymization_enabled": True
        }
    )
    
    # Retornar configuração completa
    return FraudDetectionKafkaConfig(
        cluster=cluster_config,
        topics=topics_config,
        consumer=consumer_config,
        producer=producer_config,
        streams=streams_config,
        regional_configs=regional_configs
    )


def save_config_to_file(config: FraudDetectionKafkaConfig, filepath: str) -> bool:
    """
    Salva a configuração em um arquivo JSON.
    
    Args:
        config: Configuração a ser salva
        filepath: Caminho do arquivo
        
    Returns:
        bool: True se a operação foi bem-sucedida, False caso contrário
    """
    try:
        # Converter configuração para dicionário
        config_dict = {
            "cluster": {
                "bootstrap_servers": config.cluster.bootstrap_servers,
                "security_protocol": config.cluster.security_protocol,
                "sasl_mechanism": config.cluster.sasl_mechanism,
                "client_id": config.cluster.client_id,
                # Não incluir credenciais sensíveis
                "metrics_enabled": config.cluster.metrics_enabled,
                "metrics_polling_interval_ms": config.cluster.metrics_polling_interval_ms
            },
            "topics": {
                "document_validation_events": vars(config.topics.document_validation_events),
                "authentication_events": vars(config.topics.authentication_events),
                "transaction_events": vars(config.topics.transaction_events),
                "session_events": vars(config.topics.session_events),
                "alert_events": vars(config.topics.alert_events),
                "behavior_events": vars(config.topics.behavior_events),
                "regional_fraud_patterns": vars(config.topics.regional_fraud_patterns),
                "dead_letter_queue": vars(config.topics.dead_letter_queue)
            },
            "consumer": vars(config.consumer),
            "producer": vars(config.producer),
            "streams": vars(config.streams),
            "regional_configs": {
                region: {
                    "region_code": region_config.region_code,
                    "topics_prefix": region_config.topics_prefix,
                    "regional_consumer_group_suffix": region_config.regional_consumer_group_suffix,
                    "regional_stream_store_name": region_config.regional_stream_store_name,
                    "rate_limit_events_per_second": region_config.rate_limit_events_per_second,
                    "additional_settings": region_config.additional_settings
                }
                for region, region_config in config.regional_configs.items()
            }
        }
        
        # Salvar em arquivo
        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(config_dict, f, indent=2)
            
        logger.info(f"Configuração salva com sucesso em {filepath}")
        return True
    except Exception as e:
        logger.error(f"Erro ao salvar configuração: {str(e)}")
        return False


def load_config_from_file(filepath: str) -> Optional[FraudDetectionKafkaConfig]:
    """
    Carrega a configuração de um arquivo JSON.
    
    Args:
        filepath: Caminho do arquivo
        
    Returns:
        Optional[FraudDetectionKafkaConfig]: Configuração carregada ou None em caso de erro
    """
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            config_dict = json.load(f)
            
        # Converter dicionário para objetos de configuração
        cluster_config = KafkaClusterConfig(
            bootstrap_servers=config_dict["cluster"]["bootstrap_servers"],
            security_protocol=config_dict["cluster"].get("security_protocol", "SASL_SSL"),
            sasl_mechanism=config_dict["cluster"].get("sasl_mechanism", "PLAIN"),
            client_id=config_dict["cluster"].get("client_id", "innovabiz-fraud-detection"),
            metrics_enabled=config_dict["cluster"].get("metrics_enabled", True),
            metrics_polling_interval_ms=config_dict["cluster"].get("metrics_polling_interval_ms", 30000)
        )
        
        topics_config = FraudDetectionTopics(
            document_validation_events=KafkaTopicConfig(**config_dict["topics"]["document_validation_events"]),
            authentication_events=KafkaTopicConfig(**config_dict["topics"]["authentication_events"]),
            transaction_events=KafkaTopicConfig(**config_dict["topics"]["transaction_events"]),
            session_events=KafkaTopicConfig(**config_dict["topics"]["session_events"]),
            alert_events=KafkaTopicConfig(**config_dict["topics"]["alert_events"]),
            behavior_events=KafkaTopicConfig(**config_dict["topics"]["behavior_events"]),
            regional_fraud_patterns=KafkaTopicConfig(**config_dict["topics"]["regional_fraud_patterns"]),
            dead_letter_queue=KafkaTopicConfig(**config_dict["topics"]["dead_letter_queue"])
        )
        
        consumer_config = KafkaConsumerConfig(**config_dict["consumer"])
        producer_config = KafkaProducerConfig(**config_dict["producer"])
        streams_config = StreamsConfig(**config_dict["streams"])
        
        regional_configs = {}
        for region, region_dict in config_dict["regional_configs"].items():
            regional_configs[region] = RegionalConfig(
                region_code=region_dict["region_code"],
                topics_prefix=region_dict["topics_prefix"],
                regional_consumer_group_suffix=region_dict["regional_consumer_group_suffix"],
                regional_stream_store_name=region_dict["regional_stream_store_name"],
                rate_limit_events_per_second=region_dict["rate_limit_events_per_second"],
                additional_settings=region_dict.get("additional_settings", {})
            )
        
        # Retornar configuração completa
        return FraudDetectionKafkaConfig(
            cluster=cluster_config,
            topics=topics_config,
            consumer=consumer_config,
            producer=producer_config,
            streams=streams_config,
            regional_configs=regional_configs
        )
    except Exception as e:
        logger.error(f"Erro ao carregar configuração: {str(e)}")
        return None


def get_default_config() -> FraudDetectionKafkaConfig:
    """
    Retorna uma configuração padrão para ambientes de desenvolvimento.
    
    Returns:
        FraudDetectionKafkaConfig: Configuração padrão
    """
    # Configuração do cluster
    cluster_config = KafkaClusterConfig(
        bootstrap_servers=["localhost:9092"],
        security_protocol="PLAINTEXT",  # Sem segurança para desenvolvimento local
        sasl_mechanism="PLAIN",
        sasl_username="",
        sasl_password="",
        ssl_cafile=None,
        client_id="innovabiz-fraud-detection-dev"
    )
    
    # Configuração de tópicos com replicação=1 para desenvolvimento local
    topics_config = FraudDetectionTopics(
        document_validation_events=KafkaTopicConfig(
            name="document-validation-events",
            partitions=3,
            replication_factor=1
        ),
        authentication_events=KafkaTopicConfig(
            name="authentication-events",
            partitions=3,
            replication_factor=1
        ),
        transaction_events=KafkaTopicConfig(
            name="transaction-events",
            partitions=3,
            replication_factor=1
        ),
        session_events=KafkaTopicConfig(
            name="session-events",
            partitions=3,
            replication_factor=1
        ),
        alert_events=KafkaTopicConfig(
            name="alert-events",
            partitions=2,
            replication_factor=1
        ),
        behavior_events=KafkaTopicConfig(
            name="behavior-events",
            partitions=3,
            replication_factor=1
        ),
        regional_fraud_patterns=KafkaTopicConfig(
            name="regional-fraud-patterns",
            partitions=2,
            replication_factor=1
        ),
        dead_letter_queue=KafkaTopicConfig(
            name="fraud-detection-dlq",
            partitions=1,
            replication_factor=1
        )
    )
    
    # Configurações padrão para o ambiente de desenvolvimento
    consumer_config = KafkaConsumerConfig(
        group_id="fraud-detection-dev-group",
        auto_offset_reset="earliest",
        enable_auto_commit=False
    )
    
    producer_config = KafkaProducerConfig(
        acks="all",
        enable_idempotence=True
    )
    
    streams_config = StreamsConfig(
        application_id="fraud-detection-streams-dev",
        state_dir="/tmp/kafka-streams-dev",
        num_stream_threads=2,
        replication_factor=1
    )
    
    # Configurações regionais simplificadas para desenvolvimento
    regional_configs = {
        "AO": RegionalConfig(
            region_code="AO",
            topics_prefix="ao-dev-",
            regional_stream_store_name="angola-dev-store"
        ),
        "BR": RegionalConfig(
            region_code="BR",
            topics_prefix="br-dev-",
            regional_stream_store_name="brasil-dev-store"
        ),
        "MZ": RegionalConfig(
            region_code="MZ",
            topics_prefix="mz-dev-",
            regional_stream_store_name="mocambique-dev-store"
        ),
        "PT": RegionalConfig(
            region_code="PT",
            topics_prefix="pt-dev-",
            regional_stream_store_name="portugal-dev-store",
            additional_settings={
                "gdpr_compliant": True,
                "anonymization_enabled": True
            }
        )
    }
    
    return FraudDetectionKafkaConfig(
        cluster=cluster_config,
        topics=topics_config,
        consumer=consumer_config,
        producer=producer_config,
        streams=streams_config,
        regional_configs=regional_configs
    )


def setup_environment():
    """Configura o ambiente com as variáveis necessárias para testes."""
    # Criar diretórios necessários
    os.makedirs("logs", exist_ok=True)
    os.makedirs("config", exist_ok=True)
    
    # Salvar configuração padrão para desenvolvimento
    default_config = get_default_config()
    save_config_to_file(default_config, "config/kafka_fraud_detection_dev.json")
    
    # Salvar configuração baseada em variáveis de ambiente para produção
    try:
        env_config = load_config_from_env()
        save_config_to_file(env_config, "config/kafka_fraud_detection_prod.json")
    except Exception as e:
        logger.warning(f"Falha ao carregar configuração de variáveis de ambiente: {str(e)}")


if __name__ == "__main__":
    # Configurar ambiente para testes
    setup_environment()
    
    # Carregar configuração de desenvolvimento
    dev_config = load_config_from_file("config/kafka_fraud_detection_dev.json")
    if dev_config:
        print("Configuração de desenvolvimento carregada com sucesso:")
        print(f"Bootstrap Servers: {dev_config.cluster.bootstrap_servers}")
        print(f"Tópicos configurados: {[t.name for t in [dev_config.topics.document_validation_events, dev_config.topics.authentication_events, dev_config.topics.transaction_events]]}")
        print(f"Regiões configuradas: {list(dev_config.regional_configs.keys())}")