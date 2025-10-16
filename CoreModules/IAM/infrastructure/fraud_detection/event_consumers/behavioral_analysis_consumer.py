#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Consumidor de Eventos para Análise Comportamental

Este módulo implementa o consumidor especializado em processar eventos
para análise comportamental e detecção de anomalias.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import datetime
import uuid
import threading
from typing import Dict, Any, List, Optional, Union, Tuple, Set
import pandas as pd
import numpy as np
from collections import defaultdict, Counter
from concurrent.futures import ThreadPoolExecutor

# Importação do consumidor base
try:
    from infrastructure.fraud_detection.event_consumers.base_consumer import BaseEventConsumer, ProcessingResult
except ImportError:
    import os
    import sys
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
    from event_consumers.base_consumer import BaseEventConsumer, ProcessingResult

# Configuração de logging
logger = logging.getLogger("behavioral_analysis_consumer")
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)


class BehavioralAnalysisConsumer(BaseEventConsumer):
    """
    Consumidor especializado em análise comportamental de usuários.
    
    Este consumidor analisa padrões de comportamento dos usuários 
    para identificar anomalias e possíveis fraudes, com adaptações 
    específicas para diferentes contextos regionais.
    """
    
    def __init__(
        self, 
        consumer_group_id: str = "behavioral_analysis_consumer", 
        config_file: Optional[str] = None,
        region_code: Optional[str] = None,
        model_path: Optional[str] = None,
        user_profile_db: Optional[str] = None,
        behavioral_rules_path: Optional[str] = None,
        max_workers: int = 4,
        history_window_days: int = 30,
        profile_update_interval_hours: int = 24
    ):
        """
        Inicializa o consumidor de análise comportamental.
        
        Args:
            consumer_group_id: ID do grupo de consumidores
            config_file: Caminho para o arquivo de configuração
            region_code: Código da região para filtrar eventos (AO, BR, MZ, PT)
            model_path: Caminho para o modelo de ML
            user_profile_db: Conexão para o banco de dados de perfis de usuário
            behavioral_rules_path: Caminho para as regras de comportamento
            max_workers: Número máximo de workers para processamento paralelo
            history_window_days: Janela de dias para análise histórica
            profile_update_interval_hours: Intervalo para atualização de perfis em horas
        """
        # Tópicos a serem consumidos (diversos eventos comportamentais)
        topics = [
            "fraud_detection.user_events",
            "fraud_detection.authentication_events",
            "fraud_detection.session_events",
            "fraud_detection.device_events",
            "iam.user_activity",
            "iam.authentication_attempts",
            "payment_gateway.user_events",
            "mobile_money.user_events",
            "ecommerce.user_events",
            "bureau_creditos.user_events"
        ]
        
        # Se houver região específica, ajustar tópicos
        if region_code:
            region_prefixes = {
                "AO": "angola.",
                "BR": "brasil.",
                "MZ": "mocambique.",
                "PT": "portugal."
            }
            if region_code in region_prefixes:
                topics = [f"{region_prefixes[region_code]}{topic}" for topic in topics]
        
        # Inicializar consumidor base
        super().__init__(
            consumer_group_id=consumer_group_id,
            topics=topics,
            config_file=config_file,
            region_code=region_code,
            enable_auto_commit=False,  # Controle manual de commit
            isolation_level="read_committed"
        )
        
        # Variáveis específicas deste consumidor
        self.region_code = region_code
        self.model_path = model_path
        self.user_profile_db = user_profile_db
        self.behavioral_rules_path = behavioral_rules_path
        self.max_workers = max_workers
        self.history_window_days = history_window_days
        self.profile_update_interval_hours = profile_update_interval_hours
        
        # Executor para processamento paralelo
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        
        # Cache de perfis de usuário
        self.user_profiles = {}
        self.profile_locks = defaultdict(threading.Lock)  # Locks por usuário
        self.global_profile_lock = threading.Lock()
        
        # Cache de sessões ativas
        self.active_sessions = {}
        self.session_lock = threading.Lock()
        
        # Contador de eventos por usuário
        self.event_counters = defaultdict(Counter)
        self.counter_lock = threading.Lock()
        
        # Conjunto de dispositivos conhecidos por usuário
        self.known_devices = defaultdict(set)
        self.devices_lock = threading.Lock()
        
        # Registros de localização por usuário
        self.user_locations = defaultdict(list)
        self.locations_lock = threading.Lock()
        
        # Estatísticas do consumidor
        self.behavior_stats = {
            "total_events": 0,
            "anomalous_events": 0,
            "suspicious_users": set(),
            "event_types": Counter(),
            "anomaly_types": Counter(),
            "alerts_generated": 0,
            "profiles_created": 0,
            "profiles_updated": 0
        }
        
        # Carregar regras comportamentais e modelos
        self.behavioral_rules = self._load_behavioral_rules()
        self.model = self._load_model()
        
        # Iniciar thread de atualização de perfis
        self.update_thread = threading.Thread(target=self._schedule_profile_updates, daemon=True)
        self.update_thread.start()
        
        logger.info(f"Consumidor de análise comportamental inicializado para região: {region_code or 'Todas'}")
    
    def _load_behavioral_rules(self) -> Dict:
        """
        Carrega as regras de análise comportamental.
        
        Returns:
            Dict: Regras de comportamento por região e contexto
        """
        if not self.behavioral_rules_path:
            # Regras padrão
            return {
                "default": {
                    "session_thresholds": {
                        "max_concurrent_sessions": 3,
                        "max_session_time_hours": 12,
                        "max_idle_time_minutes": 30
                    },
                    "authentication_thresholds": {
                        "max_failed_attempts": 5,
                        "lockout_period_minutes": 15,
                        "max_password_changes_per_day": 2,
                        "max_device_changes_per_week": 3
                    },
                    "navigation_patterns": {
                        "unusual_page_sequence_threshold": 0.7,
                        "max_navigation_speed_pages_per_minute": 30,
                        "suspicious_api_call_patterns": [
                            "multiple_account_access",
                            "rapid_permission_changes",
                            "sensitive_data_bulk_access"
                        ]
                    },
                    "temporal_patterns": {
                        "unusual_time_weight": 0.6,
                        "unusual_day_weight": 0.5,
                        "working_hours_start": 8,  # 8:00
                        "working_hours_end": 18    # 18:00
                    },
                    "location_patterns": {
                        "location_change_speed_threshold_kmh": 800,  # Velocidade suspeita entre logins
                        "max_different_countries_per_day": 2
                    },
                    "behavior_scoring": {
                        "baseline_deviation_threshold": 0.3,  # Desvio aceitável do baseline
                        "rapid_behavior_change_threshold": 0.5,  # Mudança rápida em score
                        "anomaly_score_threshold": 0.7  # Score para considerar anomalia
                    }
                },
                "AO": {
                    "session_thresholds": {
                        "max_concurrent_sessions": 2,
                        "max_session_time_hours": 8,
                        "max_idle_time_minutes": 20
                    },
                    "authentication_thresholds": {
                        "max_failed_attempts": 4,
                        "lockout_period_minutes": 20,
                        "max_password_changes_per_day": 1,
                        "max_device_changes_per_week": 2
                    },
                    "location_patterns": {
                        "location_change_speed_threshold_kmh": 500,
                        "trusted_locations": ["Luanda", "Benguela", "Huambo"],
                        "high_risk_locations": ["areas_fronteiriças", "províncias_remotas"]
                    },
                    "mobile_device_patterns": {
                        "primary_device_trust_level": "high",
                        "multi_device_threshold": 3,
                        "suspicious_device_patterns": [
                            "multiple_device_change",
                            "rooted_device",
                            "emulator_detection"
                        ]
                    },
                    "connection_patterns": {
                        "trusted_networks": ["Unitel", "Movicel", "Angola Telecom"],
                        "suspicious_networks": ["unknown_vpn", "tor_exit_nodes"]
                    }
                },
                "BR": {
                    "session_thresholds": {
                        "max_concurrent_sessions": 3,
                        "max_session_time_hours": 10,
                        "max_idle_time_minutes": 25
                    },
                    "authentication_thresholds": {
                        "max_failed_attempts": 5,
                        "lockout_period_minutes": 15,
                        "max_password_changes_per_day": 2,
                        "max_device_changes_per_week": 3
                    },
                    "location_patterns": {
                        "location_change_speed_threshold_kmh": 800,
                        "trusted_locations": ["São Paulo", "Rio de Janeiro", "Brasília", "Belo Horizonte"],
                        "high_risk_locations": ["fronteiras_paraguai", "fronteiras_bolivia"]
                    },
                    "pix_behavior_patterns": {
                        "max_key_registrations_per_day": 3,
                        "max_pix_recipients_per_day": 15,
                        "suspicious_patterns": [
                            "multiple_keys_single_account",
                            "rapid_key_generation_deletion",
                            "unusual_transaction_hours"
                        ]
                    },
                    "mobile_device_patterns": {
                        "primary_device_trust_level": "high",
                        "multi_device_threshold": 4,
                        "device_change_cooling_period_hours": 24
                    },
                    "lgpd_compliance": {
                        "data_access_monitoring": True,
                        "sensitive_data_patterns": [
                            "bulk_data_export",
                            "personal_data_aggregation",
                            "unusual_data_access_sequence"
                        ]
                    }
                },
                "MZ": {
                    "session_thresholds": {
                        "max_concurrent_sessions": 2,
                        "max_session_time_hours": 8,
                        "max_idle_time_minutes": 20
                    },
                    "authentication_thresholds": {
                        "max_failed_attempts": 4,
                        "lockout_period_minutes": 20,
                        "max_password_changes_per_day": 1,
                        "max_device_changes_per_week": 2
                    },
                    "location_patterns": {
                        "location_change_speed_threshold_kmh": 500,
                        "trusted_locations": ["Maputo", "Beira", "Nampula"],
                        "high_risk_locations": ["fronteiras_africa_do_sul", "fronteiras_tanzania"]
                    },
                    "mobile_money_behavior": {
                        "max_cashin_transactions_per_day": 10,
                        "max_cashout_transactions_per_day": 5,
                        "max_p2p_transfers_per_day": 15,
                        "suspicious_patterns": [
                            "cashin_cashout_same_agent",
                            "circular_transfers",
                            "agent_splitting"
                        ]
                    },
                    "connection_patterns": {
                        "trusted_networks": ["Vodacom", "Tmcel", "Movitel"],
                        "network_change_threshold_per_day": 3
                    }
                },
                "PT": {
                    "session_thresholds": {
                        "max_concurrent_sessions": 3,
                        "max_session_time_hours": 10,
                        "max_idle_time_minutes": 30
                    },
                    "authentication_thresholds": {
                        "max_failed_attempts": 5,
                        "lockout_period_minutes": 15,
                        "max_password_changes_per_day": 2,
                        "max_device_changes_per_week": 3
                    },
                    "location_patterns": {
                        "location_change_speed_threshold_kmh": 800,
                        "trusted_locations": ["Lisboa", "Porto", "Braga", "Coimbra"],
                        "high_risk_locations": ["non_eu_countries"]
                    },
                    "gdpr_compliance": {
                        "data_access_monitoring": True,
                        "data_export_monitoring": True,
                        "right_to_be_forgotten_monitoring": True,
                        "sensitive_data_patterns": [
                            "bulk_data_access",
                            "cross_border_data_transfer",
                            "profile_building_actions"
                        ]
                    },
                    "psd2_monitoring": {
                        "sca_bypass_attempts": True,
                        "suspicious_consent_management": True,
                        "unusual_third_party_access": True
                    },
                    "european_payment_patterns": {
                        "sepa_monitoring": True,
                        "unusual_iban_patterns": True,
                        "cross_border_payment_scrutiny": True
                    }
                }
            }
        
        try:
            # Carregar regras de arquivo
            import yaml
            with open(self.behavioral_rules_path, 'r') as file:
                return yaml.safe_load(file)
        except Exception as e:
            logger.error(f"Erro ao carregar regras de comportamento: {str(e)}. Usando regras padrão.")
            return self._load_behavioral_rules()
    
    def _load_model(self):
        """
        Carrega o modelo de ML para detecção de anomalias comportamentais.
        
        Returns:
            object: Modelo de ML carregado
        """
        if not self.model_path:
            # Modelo simulado simples
            class DummyBehavioralModel:
                def predict_anomaly_score(self, features, baseline):
                    # Simula score de anomalia baseado na diferença entre comportamento atual e baseline
                    import random
                    
                    # Calcula desvios básicos
                    deviations = []
                    for key in baseline:
                        if key in features and isinstance(baseline[key], (int, float)) and isinstance(features[key], (int, float)):
                            if baseline[key] != 0:  # Evitar divisão por zero
                                deviation = abs(features[key] - baseline[key]) / baseline[key]
                                deviations.append(min(1.0, deviation))  # Limitar a 100% de desvio
                    
                    # Se não houver desvios calculáveis, use um valor aleatório baixo
                    if not deviations:
                        return random.uniform(0.1, 0.3)
                    
                    # Média dos desvios + componente aleatório
                    base_score = sum(deviations) / len(deviations)
                    return min(0.95, max(0.05, base_score + random.uniform(-0.1, 0.2)))
                
                def update_baseline(self, baseline, features, learning_rate=0.1):
                    # Atualiza o baseline com novos dados (aprendizado contínuo)
                    updated = baseline.copy()
                    for key, value in features.items():
                        if key in baseline and isinstance(baseline[key], (int, float)) and isinstance(value, (int, float)):
                            updated[key] = baseline[key] * (1 - learning_rate) + value * learning_rate
                        else:
                            updated[key] = value
                    return updated
            
            return DummyBehavioralModel()
        
        try:
            # Carregar modelo real
            import joblib
            return joblib.load(self.model_path)
        except Exception as e:
            logger.error(f"Erro ao carregar modelo comportamental: {str(e)}. Usando modelo simplificado.")
            return self._load_model()
    
    def _schedule_profile_updates(self):
        """
        Agenda atualizações periódicas de perfis de usuário.
        """
        import time
        while True:
            try:
                # Aguardar pelo intervalo configurado
                time.sleep(self.profile_update_interval_hours * 3600)
                
                # Atualizar perfis em cache
                with self.global_profile_lock:
                    user_ids = list(self.user_profiles.keys())
                
                logger.info(f"Iniciando atualização programada de {len(user_ids)} perfis de usuário")
                
                # Processar em lotes para não sobrecarregar
                batch_size = 50
                for i in range(0, len(user_ids), batch_size):
                    batch = user_ids[i:i+batch_size]
                    futures = []
                    
                    for user_id in batch:
                        futures.append(
                            self.executor.submit(self._update_user_profile_in_db, user_id)
                        )
                    
                    # Aguardar conclusão do lote
                    for future in futures:
                        try:
                            future.result(timeout=30)  # 30s timeout por perfil
                        except Exception as e:
                            logger.error(f"Erro na atualização programada de perfil: {str(e)}")
                
                logger.info(f"Atualização programada de perfis concluída")
                    
            except Exception as e:
                logger.error(f"Erro na thread de atualização de perfis: {str(e)}")
    
    def _update_user_profile_in_db(self, user_id):
        """
        Atualiza o perfil de um usuário no banco de dados.
        
        Args:
            user_id: ID do usuário
        """
        try:
            with self.profile_locks[user_id]:
                if user_id in self.user_profiles:
                    profile = self.user_profiles[user_id]
                    
                    # Simular atualização no banco de dados
                    logger.debug(f"Perfil do usuário {user_id} atualizado no banco de dados")
                    
                    # Em uma implementação real, seria feita uma chamada ao banco
                    # db_client.update_user_profile(user_id, profile)
                    
                    # Registrar atualização nas estatísticas
                    self.behavior_stats["profiles_updated"] += 1
        except Exception as e:
            logger.error(f"Erro ao atualizar perfil do usuário {user_id} no banco: {str(e)}")
    
    def on_consumer_start(self):
        """Ações ao iniciar o consumidor."""
        logger.info("Consumidor de análise comportamental iniciado. Processando eventos...")
    
    def on_consumer_stop(self):
        """Ações ao parar o consumidor."""
        self.executor.shutdown(wait=True)
        logger.info("Executor de threads encerrado.")
        
        # Log de estatísticas
        logger.info(f"Estatísticas de análise comportamental:")
        logger.info(f"Total de eventos processados: {self.behavior_stats['total_events']}")
        logger.info(f"Eventos anômalos: {self.behavior_stats['anomalous_events']}")
        logger.info(f"Usuários suspeitos: {len(self.behavior_stats['suspicious_users'])}")
        logger.info(f"Alertas gerados: {self.behavior_stats['alerts_generated']}")
        logger.info(f"Perfis criados/atualizados: {self.behavior_stats['profiles_created']}/{self.behavior_stats['profiles_updated']}")
        
        if self.behavior_stats['total_events'] > 0:
            anomaly_rate = (self.behavior_stats['anomalous_events'] / self.behavior_stats['total_events']) * 100
            logger.info(f"Taxa de anomalias: {anomaly_rate:.2f}%")    def process_event(self, topic: str, event: Dict[str, Any]) -> ProcessingResult:
        """
        Processa um evento comportamental.
        
        Args:
            topic: Tópico de onde o evento foi recebido
            event: Evento comportamental
            
        Returns:
            ProcessingResult: Resultado do processamento
        """
        try:
            # Normalizar formato do evento
            normalized_event = self._normalize_event_format(topic, event)
            
            # Extrair informações básicas
            event_id = normalized_event.get('event_id', str(uuid.uuid4()))
            user_id = normalized_event.get('user_id', 'unknown')
            event_type = normalized_event.get('event_type', 'unknown')
            timestamp = normalized_event.get('timestamp', datetime.datetime.now())
            
            # Log de processamento
            logger.debug(f"Processando evento comportamental {event_id} do usuário {user_id} (tipo: {event_type})")
            
            # Atualizar estatísticas
            with self.counter_lock:
                self.behavior_stats['total_events'] += 1
                self.behavior_stats['event_types'][event_type] += 1
                self.event_counters[user_id][event_type] += 1
            
            # Obter ou criar perfil do usuário
            user_profile = self._get_or_create_user_profile(user_id)
            
            # Atualizar dados contextuais no perfil
            self._update_contextual_data(user_id, normalized_event)
            
            # Analisar evento em busca de anomalias comportamentais
            anomalies, anomaly_score = self._analyze_behavioral_anomalies(user_id, normalized_event, user_profile)
            
            # Atualizar perfil do usuário com novo evento
            self._update_user_profile(user_id, normalized_event, anomalies)
            
            # Se foram encontradas anomalias, registrar e gerar alertas
            if anomalies:
                with self.counter_lock:
                    self.behavior_stats['anomalous_events'] += 1
                    self.behavior_stats['suspicious_users'].add(user_id)
                    
                    # Registrar tipos de anomalias
                    for anomaly in anomalies:
                        anomaly_type = anomaly.get('type', 'unknown')
                        self.behavior_stats['anomaly_types'][anomaly_type] += 1
                
                # Gerar alerta de anomalia comportamental
                if anomaly_score >= self.behavioral_rules['default']['behavior_scoring']['anomaly_score_threshold']:
                    self._generate_behavioral_alert(user_id, normalized_event, anomalies, anomaly_score)
                    self.behavior_stats['alerts_generated'] += 1
            
            # Resultado do processamento
            return ProcessingResult.success_result(
                f"Evento {event_id} processado com sucesso",
                data={
                    "event_id": event_id,
                    "user_id": user_id,
                    "event_type": event_type,
                    "anomalies": anomalies,
                    "anomaly_score": anomaly_score
                }
            )
            
        except KeyError as ke:
            error_msg = f"Erro ao processar evento: campo obrigatório não encontrado - {str(ke)}"
            logger.error(error_msg)
            return ProcessingResult.failure_result(error_msg, ke)
        
        except Exception as e:
            error_msg = f"Erro inesperado ao processar evento comportamental: {str(e)}"
            logger.exception(error_msg)
            return ProcessingResult.failure_result(error_msg, e)
    
    def _normalize_event_format(self, topic: str, event: Dict[str, Any]) -> Dict[str, Any]:
        """
        Padroniza o formato do evento de diferentes fontes.
        
        Args:
            topic: Tópico de origem do evento
            event: Evento original
            
        Returns:
            Dict: Evento em formato padronizado
        """
        normalized = {}
        
        try:
            # Mapear campos comuns
            normalized['event_id'] = event.get('event_id') or event.get('id') or str(uuid.uuid4())
            normalized['user_id'] = event.get('user_id') or event.get('subject_id') or event.get('client_id') or 'unknown'
            
            # Determinar tipo de evento com base no tópico e dados
            if 'event_type' in event:
                normalized['event_type'] = event['event_type']
            elif 'type' in event:
                normalized['event_type'] = event['type']
            else:
                # Inferir tipo baseado no tópico
                if 'authentication' in topic:
                    normalized['event_type'] = 'authentication'
                elif 'session' in topic:
                    normalized['event_type'] = 'session'
                elif 'device' in topic:
                    normalized['event_type'] = 'device'
                elif 'activity' in topic:
                    normalized['event_type'] = 'user_activity'
                else:
                    normalized['event_type'] = 'user_event'
            
            # Copiar metadados se existirem
            if 'metadata' in event:
                normalized['metadata'] = event['metadata']
            else:
                normalized['metadata'] = {}
                
            # Adicionar origem do evento
            normalized['metadata']['source_topic'] = topic
                
            # Extrair informações de timestamp
            if 'timestamp' in event:
                normalized['timestamp'] = self._parse_timestamp(event['timestamp'])
            elif 'event_time' in event:
                normalized['timestamp'] = self._parse_timestamp(event['event_time'])
            elif 'created_at' in event:
                normalized['timestamp'] = self._parse_timestamp(event['created_at'])
            else:
                normalized['timestamp'] = datetime.datetime.now()
            
            # Extrair informações específicas baseadas no tipo de evento
            if normalized['event_type'] == 'authentication':
                normalized['auth_result'] = event.get('result') or event.get('auth_result')
                normalized['auth_method'] = event.get('method') or event.get('auth_method')
                normalized['auth_factor'] = event.get('factor') or event.get('auth_factor')
                normalized['successful'] = event.get('successful', True)
                normalized['failure_reason'] = event.get('failure_reason')
                
            elif normalized['event_type'] == 'session':
                normalized['session_id'] = event.get('session_id')
                normalized['session_action'] = event.get('action') or 'unknown'  # start, refresh, end
                normalized['session_duration'] = event.get('duration')
                normalized['idle_time'] = event.get('idle_time')
                
            elif normalized['event_type'] == 'device':
                normalized['device_id'] = event.get('device_id')
                normalized['device_action'] = event.get('action') or 'unknown'  # register, verify, remove
                normalized['device_type'] = event.get('device_type') or event.get('type')
                normalized['device_info'] = event.get('device_info') or event.get('info') or {}
                
            elif normalized['event_type'] == 'user_activity':
                normalized['activity'] = event.get('activity') or event.get('action')
                normalized['resource'] = event.get('resource')
                normalized['resource_type'] = event.get('resource_type')
                normalized['result'] = event.get('result')
                
            # Extrair informações de localização
            normalized['location'] = event.get('location') or {}
            if 'location' not in event and ('country_code' in event or 'country' in event):
                normalized['location'] = {
                    'country_code': event.get('country_code') or event.get('country'),
                    'city': event.get('city'),
                    'coordinates': event.get('coordinates')
                }
                
            # Extrair informações do dispositivo
            normalized['device'] = event.get('device') or {}
            if 'device' not in event and ('device_id' in event or 'device_type' in event):
                normalized['device'] = {
                    'device_id': event.get('device_id'),
                    'device_type': event.get('device_type'),
                    'ip_address': event.get('ip_address') or event.get('ip'),
                    'user_agent': event.get('user_agent')
                }
                
            # Extrair informações de contexto
            normalized['context'] = event.get('context') or {}
            if 'context' not in event:
                context_fields = {}
                # Copiar campos que podem ser de contexto
                for field in ['app_version', 'channel', 'platform', 'source', 'referrer']:
                    if field in event:
                        context_fields[field] = event[field]
                normalized['context'] = context_fields
                
            # Copiar campos adicionais
            for key, value in event.items():
                if key not in normalized and not key.startswith('_'):
                    normalized[key] = value
                    
            return normalized
            
        except Exception as e:
            logger.error(f"Erro ao normalizar formato do evento: {str(e)}")
            # Em caso de erro, retornar o evento original com campos mínimos
            return {
                'event_id': event.get('event_id') or str(uuid.uuid4()),
                'user_id': event.get('user_id') or 'unknown',
                'event_type': event.get('event_type') or 'unknown',
                'timestamp': datetime.datetime.now(),
                'original_event': event
            }
    
    def _parse_timestamp(self, timestamp_value) -> datetime.datetime:
        """
        Converte diferentes formatos de timestamp para datetime.
        
        Args:
            timestamp_value: Valor de timestamp em vários formatos possíveis
            
        Returns:
            datetime.datetime: Objeto datetime parseado
        """
        if isinstance(timestamp_value, datetime.datetime):
            return timestamp_value
            
        if isinstance(timestamp_value, (int, float)):
            # Timestamp em segundos ou milissegundos
            if timestamp_value > 1000000000000:  # Provavelmente em milissegundos
                return datetime.datetime.fromtimestamp(timestamp_value / 1000)
            else:  # Provavelmente em segundos
                return datetime.datetime.fromtimestamp(timestamp_value)
                
        if isinstance(timestamp_value, str):
            try:
                # Tentar vários formatos comuns
                formats = [
                    "%Y-%m-%dT%H:%M:%S.%fZ",  # ISO 8601 com fração de segundos e Z
                    "%Y-%m-%dT%H:%M:%SZ",      # ISO 8601 sem fração de segundos
                    "%Y-%m-%dT%H:%M:%S",       # ISO 8601 sem timezone
                    "%Y-%m-%d %H:%M:%S",       # Formato comum de SQL
                    "%d/%m/%Y %H:%M:%S",       # Formato dd/mm/yyyy
                    "%m/%d/%Y %H:%M:%S"        # Formato mm/dd/yyyy
                ]
                
                for fmt in formats:
                    try:
                        return datetime.datetime.strptime(timestamp_value, fmt)
                    except ValueError:
                        continue
                        
                # Se nenhum formato funcionar, tentar parser genérico
                from dateutil import parser
                return parser.parse(timestamp_value)
                
            except Exception:
                # Falha ao parse, retornar data/hora atual
                logger.warning(f"Não foi possível fazer parse do timestamp: {timestamp_value}")
                return datetime.datetime.now()
                
        # Para qualquer outro caso, retornar data/hora atual
        return datetime.datetime.now()
    
    def _get_or_create_user_profile(self, user_id: str) -> Dict[str, Any]:
        """
        Obtém o perfil do usuário do cache ou cria um novo.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            Dict: Perfil do usuário
        """
        with self.global_profile_lock:
            if user_id in self.user_profiles:
                return self.user_profiles[user_id]
        
        # Se não está no cache, tentar carregar do banco ou criar novo
        with self.profile_locks[user_id]:
            # Verificar novamente após adquirir o lock específico do usuário
            if user_id in self.user_profiles:
                return self.user_profiles[user_id]
                
            # Tentar carregar do banco
            profile = self._load_user_profile_from_db(user_id)
            
            # Se não existe no banco, criar perfil padrão
            if not profile:
                profile = {
                    'user_id': user_id,
                    'created_at': datetime.datetime.now().isoformat(),
                    'updated_at': datetime.datetime.now().isoformat(),
                    'behavior_baseline': {
                        'usual_auth_times': [],  # Horários comuns de autenticação
                        'usual_days': [],        # Dias da semana comuns
                        'usual_locations': [],   # Localizações comuns
                        'usual_devices': [],     # Dispositivos comuns
                        'usual_activities': {},  # Frequência de atividades
                        'usual_session_length': 0,  # Duração média de sessão
                        'usual_idle_time': 0,    # Tempo médio de inatividade
                        'activity_patterns': {}, # Padrões de navegação/API
                        'auth_stats': {
                            'success_rate': 1.0,
                            'avg_attempts': 1,
                            'password_changes': 0
                        },
                        'device_stats': {
                            'device_count': 0,
                            'device_changes': 0,
                            'primary_device': None
                        }
                    },
                    'risk_indicators': {
                        'overall_risk_score': 0.5,  # Score de risco inicial médio
                        'authentication_risk': 0.5,
                        'session_risk': 0.5,
                        'location_risk': 0.5,
                        'device_risk': 0.5,
                        'activity_risk': 0.5,
                        'recent_anomalies': []
                    },
                    'recent_events': [],
                    'region_specific': {
                        'region_code': self.region_code or 'default'
                    }
                }
                
                # Registrar criação nas estatísticas
                self.behavior_stats['profiles_created'] += 1
                logger.info(f"Perfil comportamental criado para usuário {user_id}")
            
            # Armazenar no cache
            with self.global_profile_lock:
                self.user_profiles[user_id] = profile
                
            return profile
    
    def _load_user_profile_from_db(self, user_id: str) -> Dict[str, Any]:
        """
        Carrega o perfil comportamental do usuário do banco de dados.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            Dict: Perfil do usuário ou None se não existir
        """
        if not self.user_profile_db:
            return None
            
        try:
            # Implementação depende do tipo de banco de dados usado
            # Exemplo simples simulando uma consulta
            logger.debug(f"Carregando perfil comportamental do usuário {user_id} do banco de dados")
            
            # Aqui seria implementada a lógica real de consulta ao banco
            # Por enquanto, retornamos None para simular usuário sem perfil
            return None
            
        except Exception as e:
            logger.error(f"Erro ao carregar perfil comportamental do usuário do banco: {str(e)}")
            return None
    
    def _update_contextual_data(self, user_id: str, event: Dict[str, Any]):
        """
        Atualiza dados contextuais específicos baseados no tipo de evento.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
        """
        try:
            event_type = event.get('event_type')
            
            # Atualizar registro de dispositivos
            if 'device' in event and event['device'].get('device_id'):
                device_id = event['device'].get('device_id')
                with self.devices_lock:
                    self.known_devices[user_id].add(device_id)
            
            # Atualizar localização
            if 'location' in event and event['location'].get('coordinates'):
                with self.locations_lock:
                    self.user_locations[user_id].append({
                        'coordinates': event['location'].get('coordinates'),
                        'timestamp': event.get('timestamp')
                    })
                    
                    # Limitar número de localizações armazenadas
                    max_locations = 20
                    if len(self.user_locations[user_id]) > max_locations:
                        self.user_locations[user_id] = self.user_locations[user_id][-max_locations:]
            
            # Atualizar sessões ativas
            if event_type == 'session':
                session_id = event.get('session_id')
                session_action = event.get('session_action')
                
                if not session_id:
                    return
                    
                with self.session_lock:
                    session_key = f"{user_id}:{session_id}"
                    
                    if session_action == 'start':
                        self.active_sessions[session_key] = {
                            'start_time': event.get('timestamp'),
                            'last_activity': event.get('timestamp'),
                            'device': event.get('device'),
                            'location': event.get('location')
                        }
                    elif session_action == 'refresh':
                        if session_key in self.active_sessions:
                            self.active_sessions[session_key]['last_activity'] = event.get('timestamp')
                    elif session_action == 'end':
                        if session_key in self.active_sessions:
                            del self.active_sessions[session_key]
            
        except Exception as e:
            logger.error(f"Erro ao atualizar dados contextuais: {str(e)}")
    
    def _update_user_profile(self, user_id: str, event: Dict[str, Any], anomalies: List[Dict]):
        """
        Atualiza o perfil do usuário com base no novo evento.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            anomalies: Anomalias detectadas, se houver
        """
        try:
            with self.profile_locks[user_id]:
                if user_id not in self.user_profiles:
                    return
                
                profile = self.user_profiles[user_id]
                baseline = profile['behavior_baseline']
                
                # Atualizar timestamp
                profile['updated_at'] = datetime.datetime.now().isoformat()
                
                # Adicionar evento à lista de eventos recentes
                event_summary = {
                    'event_id': event.get('event_id'),
                    'event_type': event.get('event_type'),
                    'timestamp': event.get('timestamp').isoformat() if isinstance(event.get('timestamp'), datetime.datetime) else event.get('timestamp'),
                    'has_anomalies': len(anomalies) > 0
                }
                
                profile['recent_events'].insert(0, event_summary)
                # Limitar número de eventos recentes
                profile['recent_events'] = profile['recent_events'][:20]
                
                # Adicionar anomalias à lista de anomalias recentes
                if anomalies:
                    for anomaly in anomalies:
                        anomaly_record = {
                            'type': anomaly.get('type'),
                            'description': anomaly.get('description'),
                            'severity': anomaly.get('severity'),
                            'timestamp': datetime.datetime.now().isoformat(),
                            'event_id': event.get('event_id')
                        }
                        profile['risk_indicators']['recent_anomalies'].insert(0, anomaly_record)
                    
                    # Limitar número de anomalias recentes
                    profile['risk_indicators']['recent_anomalies'] = profile['risk_indicators']['recent_anomalies'][:10]
                
                # Atualizar baseline específico por tipo de evento
                event_type = event.get('event_type')
                
                if event_type == 'authentication':
                    self._update_auth_baseline(user_id, event, baseline)
                elif event_type == 'session':
                    self._update_session_baseline(user_id, event, baseline)
                elif event_type == 'device':
                    self._update_device_baseline(user_id, event, baseline)
                elif event_type == 'user_activity':
                    self._update_activity_baseline(user_id, event, baseline)
                
                # Atualizar padrões temporais (horários/dias)
                timestamp = event.get('timestamp')
                if isinstance(timestamp, datetime.datetime):
                    # Registrar hora do dia (0-23)
                    hour = timestamp.hour
                    hour_counts = baseline.get('usual_auth_times', [0] * 24)
                    if len(hour_counts) < 24:
                        hour_counts = [0] * 24
                    hour_counts[hour] += 1
                    baseline['usual_auth_times'] = hour_counts
                    
                    # Registrar dia da semana (0-6, segunda a domingo)
                    weekday = timestamp.weekday()
                    day_counts = baseline.get('usual_days', [0] * 7)
                    if len(day_counts) < 7:
                        day_counts = [0] * 7
                    day_counts[weekday] += 1
                    baseline['usual_days'] = day_counts
                
                # Atualizar localizações usuais
                if 'location' in event and event['location'].get('city') or event['location'].get('country_code'):
                    location_key = f"{event['location'].get('city') or ''}:{event['location'].get('country_code') or ''}"
                    if location_key != ':':  # Evitar chaves vazias
                        usual_locations = baseline.get('usual_locations', [])
                        location_exists = False
                        
                        for loc in usual_locations:
                            if loc['key'] == location_key:
                                loc['count'] += 1
                                location_exists = True
                                break
                                
                        if not location_exists:
                            usual_locations.append({
                                'key': location_key,
                                'city': event['location'].get('city'),
                                'country_code': event['location'].get('country_code'),
                                'count': 1
                            })
                            
                        # Ordenar por contagem e limitar
                        baseline['usual_locations'] = sorted(usual_locations, key=lambda x: x['count'], reverse=True)[:10]
                
                # Atualizar dispositivos usuais
                if 'device' in event and event['device'].get('device_id'):
                    device_id = event['device'].get('device_id')
                    usual_devices = baseline.get('usual_devices', [])
                    device_exists = False
                    
                    for dev in usual_devices:
                        if dev['device_id'] == device_id:
                            dev['count'] += 1
                            dev['last_seen'] = datetime.datetime.now().isoformat()
                            device_exists = True
                            break
                            
                    if not device_exists:
                        usual_devices.append({
                            'device_id': device_id,
                            'device_type': event['device'].get('device_type'),
                            'user_agent': event['device'].get('user_agent'),
                            'count': 1,
                            'first_seen': datetime.datetime.now().isoformat(),
                            'last_seen': datetime.datetime.now().isoformat()
                        })
                        
                    # Ordenar por contagem e limitar
                    baseline['usual_devices'] = sorted(usual_devices, key=lambda x: x['count'], reverse=True)[:10]
                
        except Exception as e:
            logger.error(f"Erro ao atualizar perfil do usuário {user_id}: {str(e)}")
    
    # Métodos auxiliares para atualizar partes específicas do baseline
    def _update_auth_baseline(self, user_id, event, baseline):
        """Atualiza o baseline de autenticação"""
        pass  # Implementação detalhada seria adicionada aqui
        
    def _update_session_baseline(self, user_id, event, baseline):
        """Atualiza o baseline de sessão"""
        pass  # Implementação detalhada seria adicionada aqui
        
    def _update_device_baseline(self, user_id, event, baseline):
        """Atualiza o baseline de dispositivos"""
        pass  # Implementação detalhada seria adicionada aqui
        
    def _update_activity_baseline(self, user_id, event, baseline):
        """Atualiza o baseline de atividades"""
        pass  # Implementação detalhada seria adicionada aqui    def _analyze_behavioral_anomalies(self, user_id: str, event: Dict[str, Any], user_profile: Dict[str, Any]) -> Tuple[List[Dict], float]:
        """
        Analisa um evento em busca de anomalias comportamentais.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            user_profile: Perfil comportamental do usuário
            
        Returns:
            Tuple[List[Dict], float]: Lista de anomalias detectadas e score de anomalia
        """
        anomalies = []
        anomaly_scores = []
        
        try:
            event_type = event.get('event_type')
            region_code = self.region_code or 'default'
            
            # Obter regras específicas da região ou padrão
            rules = self.behavioral_rules.get(region_code, self.behavioral_rules['default'])
            
            # Extrair características comportamentais do evento
            event_features = self._extract_behavioral_features(user_id, event)
            
            # Obter baseline do perfil
            baseline = user_profile.get('behavior_baseline', {})
            
            # 1. Análise específica por tipo de evento
            if event_type == 'authentication':
                auth_anomalies = self._detect_authentication_anomalies(user_id, event, baseline, rules)
                anomalies.extend(auth_anomalies)
                
            elif event_type == 'session':
                session_anomalies = self._detect_session_anomalies(user_id, event, baseline, rules)
                anomalies.extend(session_anomalies)
                
            elif event_type == 'device':
                device_anomalies = self._detect_device_anomalies(user_id, event, baseline, rules)
                anomalies.extend(device_anomalies)
                
            elif event_type == 'user_activity':
                activity_anomalies = self._detect_activity_anomalies(user_id, event, baseline, rules)
                anomalies.extend(activity_anomalies)
            
            # 2. Análise temporal (horário incomum, dia incomum)
            temporal_anomalies = self._detect_temporal_anomalies(user_id, event, baseline, rules)
            anomalies.extend(temporal_anomalies)
            
            # 3. Análise de localização
            location_anomalies = self._detect_location_anomalies(user_id, event, baseline, rules)
            anomalies.extend(location_anomalies)
            
            # 4. Análise regional específica
            if region_code != 'default':
                regional_anomalies = self._detect_regional_specific_anomalies(user_id, event, baseline, rules, region_code)
                anomalies.extend(regional_anomalies)
            
            # 5. Análise baseada em modelo ML
            ml_anomaly_score = 0
            if event_features and baseline:
                ml_anomaly_score = self.model.predict_anomaly_score(event_features, baseline)
                
                if ml_anomaly_score > rules.get('behavior_scoring', {}).get('anomaly_score_threshold', 0.7):
                    anomalies.append({
                        'type': 'ml_behavior_anomaly',
                        'description': 'Comportamento anómalo detectado pelo modelo de ML',
                        'severity': 'medium' if ml_anomaly_score < 0.85 else 'high',
                        'score': ml_anomaly_score,
                        'related_features': list(event_features.keys())[:5]  # Primeiras 5 características
                    })
            
            # Calcular score de anomalia final
            anomaly_scores = [a.get('score', 0.5) for a in anomalies if 'score' in a]
            if ml_anomaly_score > 0:
                anomaly_scores.append(ml_anomaly_score)
                
            final_anomaly_score = max(anomaly_scores) if anomaly_scores else 0
            
            return anomalies, final_anomaly_score
            
        except Exception as e:
            logger.error(f"Erro na análise comportamental do usuário {user_id}: {str(e)}")
            return [], 0
    
    def _extract_behavioral_features(self, user_id: str, event: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extrai características comportamentais de um evento.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            
        Returns:
            Dict: Características comportamentais
        """
        features = {}
        
        try:
            event_type = event.get('event_type')
            
            # Extrair características comuns
            timestamp = event.get('timestamp')
            if isinstance(timestamp, datetime.datetime):
                features['hour'] = timestamp.hour
                features['day_of_week'] = timestamp.weekday()
                features['weekend'] = 1 if timestamp.weekday() >= 5 else 0
                
                # Calcular hora relativa ao horário comercial (8-18h)
                # 0 = meio do horário comercial, valores maiores = mais fora do horário
                if 8 <= timestamp.hour < 18:
                    # Dentro do horário comercial
                    mid_business = 13  # Meio-dia é o "centro" do horário comercial
                    features['business_hour_distance'] = abs(timestamp.hour - mid_business) / 5  # Normalizado para 0-1
                else:
                    # Fora do horário comercial
                    distance_to_business = min(
                        (timestamp.hour - 18) % 24,  # Distância após 18h
                        (8 - timestamp.hour) % 24    # Distância antes das 8h
                    )
                    features['business_hour_distance'] = min(1.0, distance_to_business / 6)  # Normalizado, max 1.0
            
            # Extrair características específicas do tipo de evento
            if event_type == 'authentication':
                features['auth_success'] = 1 if event.get('successful', True) else 0
                features['auth_method'] = hash(str(event.get('auth_method'))) % 100  # Hash para valor numérico
                features['has_2fa'] = 1 if event.get('auth_factor') == 'second' else 0
                
                # Contagens de autenticação
                with self.counter_lock:
                    user_counters = self.event_counters.get(user_id, Counter())
                    features['auth_count_24h'] = user_counters.get('authentication', 0)
                    features['failed_auth_count'] = user_counters.get('failed_authentication', 0)
                
            elif event_type == 'session':
                session_action = event.get('session_action')
                features['session_start'] = 1 if session_action == 'start' else 0
                features['session_refresh'] = 1 if session_action == 'refresh' else 0
                features['session_end'] = 1 if session_action == 'end' else 0
                
                if event.get('session_duration'):
                    features['session_duration'] = min(event.get('session_duration'), 86400) / 3600  # Normalizado para horas, max 24h
                
                if event.get('idle_time'):
                    features['idle_time'] = min(event.get('idle_time'), 3600) / 60  # Normalizado para minutos, max 60min
                
                # Número de sessões ativas
                with self.session_lock:
                    user_sessions = [s for s in self.active_sessions if s.startswith(f"{user_id}:")]
                    features['active_session_count'] = len(user_sessions)
            
            elif event_type == 'device':
                features['device_known'] = 0
                
                # Verificar se é um dispositivo conhecido
                device_id = event.get('device', {}).get('device_id')
                if device_id:
                    with self.devices_lock:
                        if device_id in self.known_devices.get(user_id, set()):
                            features['device_known'] = 1
                
                # Extrair tipo de dispositivo
                device_type = event.get('device', {}).get('device_type', '').lower()
                features['device_mobile'] = 1 if 'mobile' in device_type or 'phone' in device_type or 'android' in device_type or 'ios' in device_type else 0
                features['device_desktop'] = 1 if 'desktop' in device_type or 'laptop' in device_type or 'mac' in device_type or 'windows' in device_type else 0
                features['device_tablet'] = 1 if 'tablet' in device_type or 'ipad' in device_type else 0
                features['device_other'] = 1 if features['device_mobile'] == 0 and features['device_desktop'] == 0 and features['device_tablet'] == 0 else 0
            
            elif event_type == 'user_activity':
                activity = event.get('activity', '')
                resource = event.get('resource', '')
                resource_type = event.get('resource_type', '')
                
                # Hash de características para valores numéricos
                features['activity_hash'] = hash(str(activity)) % 100
                features['resource_hash'] = hash(str(resource)) % 100
                features['resource_type_hash'] = hash(str(resource_type)) % 100
                
                # Marcadores para tipos comuns de atividades
                features['is_data_access'] = 1 if 'view' in activity or 'read' in activity or 'get' in activity or 'list' in activity else 0
                features['is_data_modify'] = 1 if 'edit' in activity or 'update' in activity or 'write' in activity or 'put' in activity else 0
                features['is_data_delete'] = 1 if 'delete' in activity or 'remove' in activity else 0
                features['is_admin_action'] = 1 if 'admin' in activity or 'config' in activity or 'settings' in activity else 0
                features['is_sensitive_resource'] = 1 if 'password' in resource or 'credential' in resource or 'key' in resource or 'secret' in resource else 0
            
            # Extrair características de localização
            if 'location' in event:
                location = event['location']
                country_code = location.get('country_code', '')
                city = location.get('city', '')
                
                if country_code:
                    features['location_country_hash'] = hash(country_code) % 100
                
                if city:
                    features['location_city_hash'] = hash(city) % 100
                    
                # Calcular "distância" para outras localizações recentes do mesmo usuário
                with self.locations_lock:
                    recent_locations = self.user_locations.get(user_id, [])
                    if recent_locations and 'coordinates' in location:
                        current_coords = location['coordinates']
                        if len(recent_locations) > 0 and 'coordinates' in recent_locations[-1]:
                            last_coords = recent_locations[-1]['coordinates']
                            # Simulação simples de distância (poderia usar haversine para cálculo real)
                            features['location_distance'] = self._calculate_location_distance(current_coords, last_coords)
            
            # Extrair características de dispositivo
            if 'device' in event:
                device = event['device']
                ip = device.get('ip_address', '')
                user_agent = device.get('user_agent', '')
                
                if ip:
                    features['ip_hash'] = hash(ip) % 100
                
                if user_agent:
                    # Identificar características do user agent
                    features['is_mobile_ua'] = 1 if 'mobile' in user_agent.lower() or 'android' in user_agent.lower() or 'iphone' in user_agent.lower() else 0
                    features['is_browser'] = 1 if 'chrome' in user_agent.lower() or 'firefox' in user_agent.lower() or 'safari' in user_agent.lower() or 'edge' in user_agent.lower() else 0
                    features['is_bot'] = 1 if 'bot' in user_agent.lower() or 'crawler' in user_agent.lower() or 'spider' in user_agent.lower() else 0
                    features['is_old_browser'] = 1 if 'msie' in user_agent.lower() or 'trident' in user_agent.lower() else 0
            
            return features
            
        except Exception as e:
            logger.error(f"Erro ao extrair características comportamentais: {str(e)}")
            return {}
    
    def _calculate_location_distance(self, coords1, coords2):
        """
        Calcula distância aproximada entre coordenadas.
        Implementação simples - uma implementação real usaria haversine.
        """
        try:
            if isinstance(coords1, (list, tuple)) and isinstance(coords2, (list, tuple)) and len(coords1) >= 2 and len(coords2) >= 2:
                # Diferença simples de lat/long - não é geograficamente preciso mas serve para detecção de anomalias
                lat_diff = abs(coords1[0] - coords2[0])
                long_diff = abs(coords1[1] - coords2[1])
                return (lat_diff**2 + long_diff**2)**0.5
            return 0
        except Exception:
            return 0
    
    def _detect_authentication_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """
        Detecta anomalias específicas de autenticação.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            baseline: Baseline comportamental do usuário
            rules: Regras comportamentais
            
        Returns:
            List[Dict]: Lista de anomalias detectadas
        """
        anomalies = []
        
        try:
            auth_thresholds = rules.get('authentication_thresholds', {})
            
            # Verificar falhas de autenticação consecutivas
            if not event.get('successful', True):
                # Consultar contagem de falhas recentes
                with self.counter_lock:
                    failed_count = self.event_counters.get(user_id, Counter()).get('failed_authentication', 0)
                    
                    if failed_count >= auth_thresholds.get('max_failed_attempts', 5):
                        anomalies.append({
                            'type': 'auth_brute_force',
                            'description': 'Possível tentativa de força bruta detectada',
                            'severity': 'high',
                            'score': min(1.0, 0.7 + (failed_count / 20))  # Aumenta com mais tentativas
                        })
            
            # Verificar método de autenticação incomum
            usual_auth_methods = baseline.get('auth_stats', {}).get('usual_methods', [])
            current_method = event.get('auth_method')
            
            if current_method and usual_auth_methods and current_method not in [m['method'] for m in usual_auth_methods]:
                anomalies.append({
                    'type': 'unusual_auth_method',
                    'description': 'Método de autenticação incomum para este usuário',
                    'severity': 'medium',
                    'score': 0.65
                })
            
            # Verificar autenticação após redefinição recente de senha
            password_changes = baseline.get('auth_stats', {}).get('password_changes', 0)
            max_changes = auth_thresholds.get('max_password_changes_per_day', 2)
            
            if password_changes > max_changes:
                anomalies.append({
                    'type': 'excessive_password_resets',
                    'description': 'Múltiplas alterações de senha em um curto período',
                    'severity': 'medium',
                    'score': min(1.0, 0.6 + (password_changes - max_changes) * 0.1)
                })
            
            # Verificar ausência de segundo fator em conta com MFA ativado
            if event.get('successful', True) and not event.get('auth_factor') and baseline.get('auth_stats', {}).get('mfa_enabled', False):
                anomalies.append({
                    'type': 'missing_2fa',
                    'description': 'Autenticação sem segundo fator em conta com MFA ativado',
                    'severity': 'high',
                    'score': 0.85
                })
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de autenticação: {str(e)}")
            return []
    
    def _detect_session_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """
        Detecta anomalias específicas de sessão.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            baseline: Baseline comportamental do usuário
            rules: Regras comportamentais
            
        Returns:
            List[Dict]: Lista de anomalias detectadas
        """
        anomalies = []
        
        try:
            session_thresholds = rules.get('session_thresholds', {})
            
            # Verificar múltiplas sessões concorrentes
            with self.session_lock:
                user_sessions = [s for s in self.active_sessions if s.startswith(f"{user_id}:")]
                max_sessions = session_thresholds.get('max_concurrent_sessions', 3)
                
                if len(user_sessions) > max_sessions:
                    anomalies.append({
                        'type': 'concurrent_sessions',
                        'description': f'Múltiplas sessões ativas ({len(user_sessions)})',
                        'severity': 'medium',
                        'score': min(1.0, 0.6 + (len(user_sessions) - max_sessions) * 0.1)
                    })
            
            # Verificar sessão muito longa
            if event.get('session_action') == 'end' and event.get('session_duration'):
                duration_hours = event.get('session_duration') / 3600  # Converter para horas
                max_duration = session_thresholds.get('max_session_time_hours', 12)
                
                if duration_hours > max_duration:
                    anomalies.append({
                        'type': 'long_session',
                        'description': f'Sessão excessivamente longa ({duration_hours:.1f} horas)',
                        'severity': 'low',
                        'score': min(1.0, 0.5 + (duration_hours - max_duration) / max_duration)
                    })
            
            # Verificar tempo de inatividade excessivo
            if event.get('idle_time'):
                idle_minutes = event.get('idle_time') / 60  # Converter para minutos
                max_idle = session_thresholds.get('max_idle_time_minutes', 30)
                
                if idle_minutes > max_idle:
                    anomalies.append({
                        'type': 'excessive_idle',
                        'description': f'Tempo de inatividade excessivo ({idle_minutes:.1f} minutos)',
                        'severity': 'low',
                        'score': min(1.0, 0.4 + (idle_minutes - max_idle) / (max_idle * 2))
                    })
            
            # Verificar mudanças de dispositivo durante a sessão
            if event.get('session_id') and event.get('device', {}).get('device_id'):
                session_key = f"{user_id}:{event.get('session_id')}"
                
                with self.session_lock:
                    if session_key in self.active_sessions:
                        original_device = self.active_sessions[session_key].get('device', {}).get('device_id')
                        current_device = event.get('device', {}).get('device_id')
                        
                        if original_device and current_device and original_device != current_device:
                            anomalies.append({
                                'type': 'session_device_change',
                                'description': 'Mudança de dispositivo durante a mesma sessão',
                                'severity': 'high',
                                'score': 0.8
                            })
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de sessão: {str(e)}")
            return []
    
    def _detect_device_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """Detecta anomalias relacionadas a dispositivos"""
        anomalies = []
        
        try:
            # Implementação completa seria adicionada aqui
            pass
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de dispositivo: {str(e)}")
            return []
    
    def _detect_activity_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """Detecta anomalias relacionadas a atividades de usuário"""
        anomalies = []
        
        try:
            # Implementação completa seria adicionada aqui
            pass
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de atividade: {str(e)}")
            return []
    
    def _detect_temporal_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """Detecta anomalias temporais (horário incomum, dia incomum)"""
        anomalies = []
        
        try:
            # Implementação completa seria adicionada aqui
            pass
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias temporais: {str(e)}")
            return []
    
    def _detect_location_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any]) -> List[Dict]:
        """Detecta anomalias de localização"""
        anomalies = []
        
        try:
            # Implementação completa seria adicionada aqui
            pass
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de localização: {str(e)}")
            return []
    
    def _detect_regional_specific_anomalies(self, user_id: str, event: Dict[str, Any], baseline: Dict[str, Any], rules: Dict[str, Any], region_code: str) -> List[Dict]:
        """Detecta anomalias específicas por região"""
        anomalies = []
        
        try:
            # Implementação específica para cada região (AO, BR, MZ, PT)
            if region_code == 'AO':
                # Angola - Implementação específica seria adicionada aqui
                pass
            elif region_code == 'BR':
                # Brasil - Implementação específica seria adicionada aqui
                pass
            elif region_code == 'MZ':
                # Moçambique - Implementação específica seria adicionada aqui
                pass
            elif region_code == 'PT':
                # Portugal - Implementação específica seria adicionada aqui
                pass
            
            return anomalies
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias regionais ({region_code}): {str(e)}")
            return []
    
    def _generate_behavioral_alert(self, user_id: str, event: Dict[str, Any], anomalies: List[Dict], anomaly_score: float):
        """
        Gera um alerta comportamental para anomalias detectadas.
        
        Args:
            user_id: ID do usuário
            event: Evento normalizado
            anomalies: Anomalias detectadas
            anomaly_score: Score de anomalia
        """
        try:
            # Determinar nível de severidade do alerta
            max_severity = 'low'
            for anomaly in anomalies:
                severity = anomaly.get('severity', 'low')
                if severity == 'high':
                    max_severity = 'high'
                    break
                elif severity == 'medium' and max_severity != 'high':
                    max_severity = 'medium'
            
            # Gerar ID único para o alerta
            alert_id = f"beh-{str(uuid.uuid4())[:8]}"
            
            # Criar descrição
            anomaly_types = [a.get('type', 'unknown') for a in anomalies]
            anomaly_descriptions = [a.get('description', 'Anomalia não especificada') for a in anomalies]
            
            # Construir corpo do alerta
            alert_data = {
                'alert_id': alert_id,
                'type': 'behavioral',
                'timestamp': datetime.datetime.now().isoformat(),
                'user_id': user_id,
                'severity': max_severity,
                'anomaly_score': anomaly_score,
                'anomaly_types': anomaly_types,
                'anomaly_descriptions': anomaly_descriptions,
                'event_id': event.get('event_id'),
                'event_type': event.get('event_type'),
                'context': {
                    'location': event.get('location'),
                    'device': event.get('device'),
                    'timestamp': event.get('timestamp').isoformat() if isinstance(event.get('timestamp'), datetime.datetime) else event.get('timestamp')
                },
                'region_code': self.region_code or 'default'
            }
            
            # Log do alerta
            logger.warning(f"ALERTA COMPORTAMENTAL: {max_severity.upper()} - Usuário {user_id} - Score {anomaly_score:.2f} - {', '.join(anomaly_descriptions)}")
            
            # Enviar alerta para tópico do Kafka
            self._send_behavioral_alert(alert_data)
            
            # Enviar evento para serviços de compliance se necessário
            if max_severity == 'high' or anomaly_score > 0.8:
                self._notify_compliance_services(alert_data)
                
        except Exception as e:
            logger.error(f"Erro ao gerar alerta comportamental: {str(e)}")
    
    def _send_behavioral_alert(self, alert_data: Dict[str, Any]):
        """
        Envia um alerta comportamental para o tópico Kafka.
        
        Args:
            alert_data: Dados do alerta
        """
        try:
            alert_topic = "iam.behavioral.alerts"
            
            # Adicionar prefixo de região se aplicável
            if self.region_code:
                region_prefixes = {
                    "AO": "angola.",
                    "BR": "brasil.",
                    "MZ": "mocambique.",
                    "PT": "portugal."
                }
                if self.region_code in region_prefixes:
                    alert_topic = f"{region_prefixes[self.region_code]}{alert_topic}"
            
            # Serializar alerta
            alert_json = json.dumps(alert_data)
            
            # Enviar para o Kafka usando o produtor do consumidor base
            if hasattr(self, 'producer') and self.producer:
                self.producer.produce(
                    topic=alert_topic,
                    key=alert_data['user_id'],
                    value=alert_json
                )
                self.producer.flush(timeout=5)
                logger.info(f"Alerta comportamental enviado para {alert_topic}")
            else:
                logger.warning("Produtor Kafka não disponível, alerta não enviado")
                
        except Exception as e:
            logger.error(f"Erro ao enviar alerta comportamental: {str(e)}")
    
    def _notify_compliance_services(self, alert_data: Dict[str, Any]):
        """
        Notifica serviços de compliance sobre alerta de alta severidade.
        
        Args:
            alert_data: Dados do alerta
        """
        try:
            # Implementação básica da notificação - em um sistema real, isso poderia
            # chamar uma API REST, enviar para uma fila de mensagens adicional, etc.
            logger.info(f"Notificando serviços de compliance sobre alerta de severidade {alert_data['severity']}")
            
            # Exemplo: se integrado com serviço de compliance, chamaria uma API:
            # requests.post("http://compliance-service/api/behavioral-alerts", json=alert_data)
            
        except Exception as e:
            logger.error(f"Erro ao notificar serviços de compliance: {str(e)}")


# Inicialização direta para testes
if __name__ == "__main__":
    # Configurar logging para testes
    logging.basicConfig(level=logging.INFO)
    
    # Criar instância do consumidor
    consumer = BehavioralAnalysisConsumer(
        consumer_group_id="behavioral_analysis_test",
        region_code="AO"  # Testes para Angola
    )
    
    # Iniciar consumidor (para uso real)
    try:
        consumer.start()
    except KeyboardInterrupt:
        logger.info("Consumidor interrompido pelo usuário")
    finally:
        consumer.stop()