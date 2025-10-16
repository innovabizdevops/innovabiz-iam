#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Consumidor Base para Eventos de Fraude

Este módulo implementa a classe base para todos os consumidores de eventos
do sistema de detecção de fraudes do TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import time
import uuid
import logging
import datetime
import threading
import signal
import os
import sys
from typing import Dict, Any, List, Optional, Union, Callable, Set
from abc import ABC, abstractmethod
from dataclasses import dataclass, asdict, field
from contextlib import contextmanager
from confluent_kafka import Consumer, KafkaException, TopicPartition, OFFSET_BEGINNING, OFFSET_END

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
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../..')))
    try:
        from fraud_detection.kafka_config import (
            load_config_from_env, 
            load_config_from_file, 
            get_default_config,
            FraudDetectionKafkaConfig
        )
    except ImportError:
        # Mensagem de erro para ajudar na resolução de problemas
        sys.stderr.write("ERRO: Não foi possível importar os módulos de configuração Kafka.\n")
        sys.stderr.write("Verifique se o arquivo kafka_config.py está disponível.\n")
        sys.exit(1)

# Configuração de logging
logger = logging.getLogger("fraud_detection_consumers")
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)


class ProcessingResult:
    """Classe para representar o resultado do processamento de um evento."""
    
    def __init__(self, success: bool, message: str = "", data: Any = None, error: Exception = None):
        """
        Inicializa um resultado de processamento.
        
        Args:
            success: Indica se o processamento foi bem-sucedido
            message: Mensagem descritiva sobre o resultado
            data: Dados opcionais retornados pelo processamento
            error: Exceção caso o processamento tenha falhado
        """
        self.success = success
        self.message = message
        self.data = data
        self.error = error
        self.timestamp = datetime.datetime.now().isoformat()
    
    @classmethod
    def success_result(cls, message: str = "Processamento bem-sucedido", data: Any = None) -> 'ProcessingResult':
        """Cria um resultado de sucesso."""
        return cls(True, message, data)
    
    @classmethod
    def failure_result(cls, message: str = "Falha no processamento", error: Exception = None) -> 'ProcessingResult':
        """Cria um resultado de falha."""
        return cls(False, message, None, error)
    
    def __str__(self) -> str:
        """Retorna uma representação de string do resultado."""
        status = "SUCESSO" if self.success else "FALHA"
        result = f"[{status}] {self.message}"
        if self.error:
            result += f" - Erro: {str(self.error)}"
        return result


class BaseEventConsumer(ABC):
    """
    Classe base para consumidores de eventos.
    
    Esta é uma classe abstrata que define o comportamento comum a todos os
    consumidores especializados. Ela gerencia a conexão com o Kafka, o consumo
    de eventos e o processamento básico.
    """
    
    def __init__(
        self,
        consumer_group_id: str,
        topics: List[str],
        config_file: Optional[str] = None,
        auto_offset_reset: str = "earliest",
        enable_auto_commit: bool = False,
        auto_commit_interval_ms: int = 5000,
        session_timeout_ms: int = 30000,
        heartbeat_interval_ms: int = 10000,
        max_poll_interval_ms: int = 300000,
        max_poll_records: int = 500,
        fetch_max_bytes: int = 52428800,  # 50MB
        fetch_min_bytes: int = 1,
        fetch_max_wait_ms: int = 500,
        max_partition_fetch_bytes: int = 1048576,  # 1MB
        metadata_max_age_ms: int = 300000,
        request_timeout_ms: int = 40000,
        isolation_level: str = "read_committed",
        region_code: Optional[str] = None
    ):
        """
        Inicializa o consumidor base.
        
        Args:
            consumer_group_id: ID do grupo de consumidores
            topics: Lista de tópicos para consumir
            config_file: Caminho para o arquivo de configuração
            auto_offset_reset: Estratégia de reset de offset ("earliest", "latest")
            enable_auto_commit: Ativar commit automático de offset
            auto_commit_interval_ms: Intervalo para auto-commit (ms)
            session_timeout_ms: Tempo limite da sessão (ms)
            heartbeat_interval_ms: Intervalo de heartbeat (ms)
            max_poll_interval_ms: Intervalo máximo entre polls (ms)
            max_poll_records: Máximo de registros retornados por poll
            fetch_max_bytes: Máximo de bytes a buscar por requisição
            fetch_min_bytes: Mínimo de bytes a buscar por requisição
            fetch_max_wait_ms: Tempo máximo de espera para fetch (ms)
            max_partition_fetch_bytes: Máximo de bytes por partição
            metadata_max_age_ms: Tempo máximo para metadados (ms)
            request_timeout_ms: Tempo limite para requisições (ms)
            isolation_level: Nível de isolamento
            region_code: Código da região para filtragem de eventos
        """
        self.consumer_group_id = consumer_group_id
        self.topics = topics
        self.running = False
        self.consumer = None
        self.region_code = region_code
        self.last_processed_offsets = {}
        self.processing_stats = {
            "total_processed": 0,
            "success_count": 0,
            "failure_count": 0,
            "start_time": None,
            "last_process_time": None,
            "processing_times": [],
            "errors_by_type": {}
        }
        
        # Carregar configuração
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
        
        # Configurações de consumo
        self.consumer_config = {
            'bootstrap.servers': ','.join(self.config.cluster.bootstrap_servers),
            'group.id': consumer_group_id,
            'client.id': f"{self.config.cluster.client_id}-{consumer_group_id}-{str(uuid.uuid4())[:8]}",
            'security.protocol': self.config.cluster.security_protocol,
            'sasl.mechanism': self.config.cluster.sasl_mechanism,
            'sasl.username': self.config.cluster.sasl_username,
            'sasl.password': self.config.cluster.sasl_password,
            'auto.offset.reset': auto_offset_reset,
            'enable.auto.commit': str(enable_auto_commit).lower(),
            'auto.commit.interval.ms': auto_commit_interval_ms,
            'session.timeout.ms': session_timeout_ms,
            'heartbeat.interval.ms': heartbeat_interval_ms,
            'max.poll.interval.ms': max_poll_interval_ms,
            'max.poll.records': max_poll_records,
            'fetch.max.bytes': fetch_max_bytes,
            'fetch.min.bytes': fetch_min_bytes,
            'fetch.max.wait.ms': fetch_max_wait_ms,
            'max.partition.fetch.bytes': max_partition_fetch_bytes,
            'metadata.max.age.ms': metadata_max_age_ms,
            'request.timeout.ms': request_timeout_ms,
            'isolation.level': isolation_level
        }
        
        # Configurar tratamento de sinais
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
        
        # Metadados
        self.consumer_metadata = {
            "consumer_id": self.consumer_config['client.id'],
            "group_id": consumer_group_id,
            "topics": topics,
            "region_code": region_code,
            "start_time": None,
            "hostname": os.uname().nodename if hasattr(os, 'uname') else "unknown",
            "process_id": os.getpid(),
        }
    
    def _signal_handler(self, sig, frame):
        """
        Manipulador de sinais para parar o consumidor graciosamente.
        
        Args:
            sig: Sinal recebido
            frame: Frame de execução atual
        """
        logger.info(f"Sinal recebido ({sig}). Encerrando consumidor...")
        self.stop()
    
    def initialize(self):
        """Inicializa o consumidor Kafka."""
        try:
            # Criar consumidor
            self.consumer = Consumer(self.consumer_config)
            
            # Inscrever nos tópicos
            self.consumer.subscribe(self.topics, on_assign=self._on_assign, on_revoke=self._on_revoke)
            
            logger.info(f"Consumidor Kafka inicializado com sucesso para {self.config.cluster.bootstrap_servers}")
            logger.info(f"Inscrito nos tópicos: {', '.join(self.topics)}")
            logger.info(f"Grupo de consumidores: {self.consumer_group_id}")
        except Exception as e:
            logger.error(f"Erro ao inicializar consumidor Kafka: {str(e)}")
            raise
    
    def _on_assign(self, consumer, partitions):
        """
        Callback chamado quando partições são atribuídas ao consumidor.
        
        Args:
            consumer: Instância do consumidor
            partitions: Lista de partições atribuídas
        """
        logger.info(f"Partições atribuídas: {[f'{p.topic}[{p.partition}]' for p in partitions]}")
        for partition in partitions:
            # Armazenar o offset inicial para cada partição
            self.last_processed_offsets[f"{partition.topic}:{partition.partition}"] = partition.offset
    
    def _on_revoke(self, consumer, partitions):
        """
        Callback chamado quando partições são revogadas do consumidor.
        
        Args:
            consumer: Instância do consumidor
            partitions: Lista de partições revogadas
        """
        logger.info(f"Partições revogadas: {[f'{p.topic}[{p.partition}]' for p in partitions]}")
        # Commit de offsets antes da revogação para garantir progresso
        if not self.consumer_config.get('enable.auto.commit', False):
            try:
                consumer.commit()
                logger.info("Offsets committed on revoke")
            except KafkaException as e:
                logger.warning(f"Falha ao commitar offsets na revogação: {e}")
    
    def start(self):
        """Inicia o consumidor e começa a processar mensagens."""
        if self.running:
            logger.warning("Consumidor já está em execução.")
            return
        
        logger.info(f"Iniciando consumidor {self.consumer_group_id} para os tópicos {self.topics}")
        self.initialize()
        self.running = True
        self.processing_stats["start_time"] = datetime.datetime.now().isoformat()
        self.consumer_metadata["start_time"] = self.processing_stats["start_time"]
        
        try:
            # Notificar sobre o início
            self.on_consumer_start()
            
            # Loop principal
            while self.running:
                try:
                    # Poll para novos eventos
                    msg = self.consumer.poll(timeout=1.0)
                    
                    # Se não houver mensagem, continue o loop
                    if msg is None:
                        continue
                    
                    # Se houver erro na mensagem, log e continue
                    if msg.error():
                        logger.error(f"Erro no consumo: {msg.error()}")
                        continue
                    
                    # Processar a mensagem
                    self._process_message(msg)
                    
                    # Commit do offset se auto-commit estiver desativado
                    if not self.consumer_config.get('enable.auto.commit', False):
                        self.consumer.commit(msg)
                        
                except KafkaException as ke:
                    logger.error(f"Erro Kafka durante o consumo: {str(ke)}")
                except Exception as e:
                    logger.exception(f"Erro inesperado durante o consumo: {str(e)}")
                    # Pausa curta para evitar loop de erros rápido
                    time.sleep(1)
        
        except KeyboardInterrupt:
            logger.info("Consumidor interrompido pelo usuário.")
        finally:
            self.stop()
    
    def _process_message(self, msg):
        """
        Processa uma mensagem recebida do Kafka.
        
        Args:
            msg: Mensagem do Kafka
        """
        start_time = time.time()
        
        # Converter a mensagem para objeto Python
        try:
            # Extrair informações da mensagem
            topic = msg.topic()
            partition = msg.partition()
            offset = msg.offset()
            key = msg.key().decode('utf-8') if msg.key() else None
            
            # Atualizar offset para esta partição
            partition_key = f"{topic}:{partition}"
            self.last_processed_offsets[partition_key] = offset
            
            # Deserializar valor
            value = json.loads(msg.value().decode('utf-8'))
            
            # Filtrar por região, se necessário
            if self.region_code:
                metadata = value.get('metadata', {})
                event_region = metadata.get('region_code', '')
                
                # Se a mensagem não corresponder à região, ignorar
                if event_region and event_region != self.region_code:
                    logger.debug(f"Ignorando evento da região {event_region}, esperado {self.region_code}")
                    return
            
            # Registrar informações de processamento
            logger.debug(f"Processando mensagem: tópico={topic}, partição={partition}, offset={offset}, key={key}")
            
            # Chamar o método abstrato para processamento específico
            result = self.process_event(topic, value)
            
            # Atualizar estatísticas
            self.processing_stats["total_processed"] += 1
            self.processing_stats["last_process_time"] = datetime.datetime.now().isoformat()
            
            if result.success:
                self.processing_stats["success_count"] += 1
            else:
                self.processing_stats["failure_count"] += 1
                error_type = type(result.error).__name__ if result.error else "UnknownError"
                self.processing_stats["errors_by_type"][error_type] = self.processing_stats["errors_by_type"].get(error_type, 0) + 1
                logger.warning(f"Falha ao processar evento: {result.message}")
            
            # Calcular e armazenar o tempo de processamento
            processing_time = time.time() - start_time
            self.processing_stats["processing_times"].append(processing_time)
            # Manter apenas os últimos 100 tempos de processamento
            if len(self.processing_stats["processing_times"]) > 100:
                self.processing_stats["processing_times"].pop(0)
            
            return result
            
        except json.JSONDecodeError as je:
            logger.error(f"Erro ao decodificar JSON: {str(je)}")
            self.processing_stats["failure_count"] += 1
            error_type = "JSONDecodeError"
            self.processing_stats["errors_by_type"][error_type] = self.processing_stats["errors_by_type"].get(error_type, 0) + 1
            return ProcessingResult.failure_result(f"Erro ao decodificar JSON: {str(je)}", je)
            
        except Exception as e:
            logger.exception(f"Erro ao processar mensagem: {str(e)}")
            self.processing_stats["failure_count"] += 1
            error_type = type(e).__name__
            self.processing_stats["errors_by_type"][error_type] = self.processing_stats["errors_by_type"].get(error_type, 0) + 1
            return ProcessingResult.failure_result(f"Erro ao processar mensagem: {str(e)}", e)
    
    def stop(self):
        """Para o consumidor graciosamente."""
        if not self.running:
            return
        
        logger.info("Parando o consumidor...")
        self.running = False
        
        if self.consumer:
            try:
                # Commit final de offsets
                if not self.consumer_config.get('enable.auto.commit', False):
                    self.consumer.commit()
                
                # Fechar o consumidor
                self.consumer.close()
                
                # Notificar sobre a parada
                self.on_consumer_stop()
                
                logger.info("Consumidor parado com sucesso")
                logger.info(f"Estatísticas de processamento: " +
                           f"Total={self.processing_stats['total_processed']}, " +
                           f"Sucesso={self.processing_stats['success_count']}, " +
                           f"Falha={self.processing_stats['failure_count']}")
                
            except Exception as e:
                logger.error(f"Erro ao parar consumidor: {str(e)}")
    
    def get_stats(self) -> Dict[str, Any]:
        """
        Retorna estatísticas sobre o consumo.
        
        Returns:
            Dict: Estatísticas de processamento
        """
        stats = self.processing_stats.copy()
        
        # Adicionar estatísticas calculadas
        if stats["total_processed"] > 0:
            stats["success_rate"] = stats["success_count"] / stats["total_processed"]
            stats["failure_rate"] = stats["failure_count"] / stats["total_processed"]
        else:
            stats["success_rate"] = 0
            stats["failure_rate"] = 0
        
        # Calcular média de tempo de processamento
        if stats["processing_times"]:
            stats["avg_processing_time"] = sum(stats["processing_times"]) / len(stats["processing_times"])
            stats["max_processing_time"] = max(stats["processing_times"])
            stats["min_processing_time"] = min(stats["processing_times"])
        else:
            stats["avg_processing_time"] = 0
            stats["max_processing_time"] = 0
            stats["min_processing_time"] = 0
        
        # Adicionar metadados do consumidor
        stats["consumer_metadata"] = self.consumer_metadata
        
        # Adicionar informações de lag, se disponível
        if self.consumer:
            try:
                topic_partitions = []
                for topic in self.topics:
                    metadata = self.consumer.list_topics(topic)
                    if topic in metadata.topics:
                        for partition_id in metadata.topics[topic].partitions:
                            topic_partitions.append(TopicPartition(topic, partition_id))
                
                # Obter offsets finais para calcular lag
                end_offsets = self.consumer.get_watermark_offsets(topic_partitions)
                consumer_positions = {}
                lags = {}
                
                for tp in topic_partitions:
                    position = self.consumer.position([tp])[0]
                    consumer_positions[f"{tp.topic}:{tp.partition}"] = position.offset
                    
                    # Verificar se temos o offset final
                    if (tp.topic, tp.partition) in end_offsets:
                        low, high = end_offsets[(tp.topic, tp.partition)]
                        lags[f"{tp.topic}:{tp.partition}"] = high - position.offset
                
                stats["consumer_positions"] = consumer_positions
                stats["lags"] = lags
            
            except Exception as e:
                logger.warning(f"Não foi possível obter informações de lag: {str(e)}")
        
        return stats
    
    def reset_offsets(self, to_beginning: bool = True):
        """
        Reseta os offsets do consumidor para o início ou fim dos tópicos.
        
        Args:
            to_beginning: Se True, reseta para o início. Se False, reseta para o fim.
        """
        if not self.consumer:
            logger.error("Consumidor não inicializado")
            return
        
        # Verificar se o consumidor está inscrito
        if not self.topics:
            logger.error("Nenhum tópico inscrito")
            return
        
        try:
            # Obter partições para cada tópico
            topic_partitions = []
            for topic in self.topics:
                metadata = self.consumer.list_topics(topic)
                if topic in metadata.topics:
                    for partition_id in metadata.topics[topic].partitions:
                        tp = TopicPartition(topic, partition_id, 
                                           OFFSET_BEGINNING if to_beginning else OFFSET_END)
                        topic_partitions.append(tp)
            
            # Resetar offsets
            self.consumer.assign(topic_partitions)
            
            # Commitar os offsets resetados
            self.consumer.commit()
            
            # Re-inscrever nos tópicos
            self.consumer.unassign()
            self.consumer.subscribe(self.topics)
            
            position = "início" if to_beginning else "fim"
            logger.info(f"Offsets resetados para o {position} dos tópicos: {', '.join(self.topics)}")
            
        except Exception as e:
            logger.error(f"Erro ao resetar offsets: {str(e)}")
    
    def commit_offsets(self):
        """Força o commit dos offsets atuais."""
        if not self.consumer:
            logger.error("Consumidor não inicializado")
            return
        
        try:
            self.consumer.commit()
            logger.info("Offsets committed manually")
        except Exception as e:
            logger.error(f"Erro ao commitar offsets: {str(e)}")
    
    def on_consumer_start(self):
        """
        Método chamado quando o consumidor é iniciado.
        
        Pode ser sobrescrito por classes derivadas para executar
        ações específicas no início do consumo.
        """
        pass
    
    def on_consumer_stop(self):
        """
        Método chamado quando o consumidor é parado.
        
        Pode ser sobrescrito por classes derivadas para executar
        ações específicas no fim do consumo.
        """
        pass
    
    @abstractmethod
    def process_event(self, topic: str, event: Dict[str, Any]) -> ProcessingResult:
        """
        Processa um evento de fraude.
        
        Este é um método abstrato que deve ser implementado por cada consumidor
        especializado para processar os diferentes tipos de eventos.
        
        Args:
            topic: Tópico de onde o evento foi recebido
            event: Evento a ser processado
            
        Returns:
            ProcessingResult: Resultado do processamento
        """
        pass


# Função auxiliar para iniciar consumidores em threads separadas
def start_consumer_in_thread(consumer: BaseEventConsumer) -> threading.Thread:
    """
    Inicia um consumidor em uma thread separada.
    
    Args:
        consumer: Instância do consumidor a ser iniciado
        
    Returns:
        Thread: Thread do consumidor
    """
    thread = threading.Thread(target=consumer.start)
    thread.daemon = True
    thread.start()
    return thread


# Função para parar consumidores graciosamente
def stop_consumers(consumers: List[BaseEventConsumer], threads: List[threading.Thread] = None, timeout: int = 30):
    """
    Para múltiplos consumidores graciosamente.
    
    Args:
        consumers: Lista de consumidores para parar
        threads: Lista de threads dos consumidores (opcional)
        timeout: Tempo máximo de espera em segundos
    """
    for consumer in consumers:
        consumer.stop()
    
    if threads:
        for thread in threads:
            thread.join(timeout)