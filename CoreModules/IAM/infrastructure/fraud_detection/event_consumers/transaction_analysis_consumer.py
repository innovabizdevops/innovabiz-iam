#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Consumidor de Eventos de Análise de Transações

Este módulo implementa o consumidor especializado em processar eventos
de transações para detecção de fraudes em tempo real.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import datetime
import uuid
import threading
from typing import Dict, Any, List, Optional, Union, Tuple
import pandas as pd
import numpy as np
from collections import defaultdict
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
logger = logging.getLogger("transaction_analysis_consumer")
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)


class TransactionAnalysisConsumer(BaseEventConsumer):
    """
    Consumidor especializado em processar eventos de transações financeiras.
    
    Este consumidor analisa transações em tempo real em busca de padrões suspeitos,
    mudanças de comportamento e sinais de fraude, usando heurísticas, regras
    configuráveis e modelos de machine learning.
    """
    
    def __init__(
        self, 
        consumer_group_id: str = "transaction_analysis_consumer", 
        config_file: Optional[str] = None,
        region_code: Optional[str] = None,
        rules_path: Optional[str] = None,
        model_path: Optional[str] = None,
        user_history_db: Optional[str] = None,
        max_workers: int = 4,
        memory_window_seconds: int = 3600  # Janela de memória de 1 hora
    ):
        """
        Inicializa o consumidor de análise de transações.
        
        Args:
            consumer_group_id: ID do grupo de consumidores
            config_file: Caminho para o arquivo de configuração
            region_code: Código da região para filtrar eventos (AO, BR, MZ, PT)
            rules_path: Caminho para as regras de análise de transações
            model_path: Caminho para o modelo de ML
            user_history_db: Conexão para o banco de dados de histórico de usuários
            max_workers: Número máximo de workers para processamento paralelo
            memory_window_seconds: Janela de memória para análise de transações recentes
        """
        # Tópicos a serem consumidos
        topics = [
            "fraud_detection.payment_transactions", 
            "fraud_detection.account_transactions",
            "payment_gateway.transaction_events",
            "mobile_money.transaction_events",
            "ecommerce.transaction_events"
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
        self.rules_path = rules_path
        self.model_path = model_path
        self.user_history_db = user_history_db
        self.max_workers = max_workers
        self.memory_window_seconds = memory_window_seconds
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        
        # Memória de transações recentes
        self.recent_transactions = defaultdict(list)  # user_id -> lista de transações
        self.memory_lock = threading.Lock()
        
        # Cache de perfis de usuário
        self.user_profiles = {}
        self.profile_lock = threading.Lock()
        
        # Estatísticas de análise de transações
        self.transaction_stats = {
            "total_transactions": 0,
            "suspicious_transactions": 0,
            "high_risk_transactions": 0,
            "blocked_transactions": 0,
            "transaction_types": {},
            "transaction_channels": {},
            "fraud_signals": {},
            "transaction_volumes": defaultdict(float),
            "transaction_counts": defaultdict(int)
        }
        
        # Carregar regras e modelos
        self.rules = self._load_rules()
        self.model = self._load_model()
        
        # Iniciar thread de limpeza de memória
        self.cleanup_thread = threading.Thread(target=self._cleanup_memory, daemon=True)
        self.cleanup_thread.start()
        
        logger.info(f"Consumidor de análise de transações inicializado para região: {region_code or 'Todas'}")
    
    def _load_rules(self) -> Dict:
        """
        Carrega as regras de análise de transações.
        
        Returns:
            Dict: Regras de análise por tipo de transação e região
        """
        if not self.rules_path:
            # Regras padrão
            return {
                "default": {
                    "velocity_thresholds": {
                        "max_tx_per_hour": 10,
                        "max_tx_per_day": 50,
                        "max_amount_per_day": 10000,
                        "max_different_merchants_per_hour": 5
                    },
                    "amount_thresholds": {
                        "suspicious_threshold": 500,
                        "high_risk_threshold": 2000
                    },
                    "location_rules": {
                        "max_distance_km_per_hour": 500,
                        "suspicious_countries": ["NG", "UG", "TR", "RU"]
                    },
                    "behavior_rules": {
                        "unusual_time_window_start": 23,  # 23:00
                        "unusual_time_window_end": 5,     # 05:00
                        "unusual_device_weight": 0.7,
                        "unusual_location_weight": 0.8,
                        "unusual_amount_weight": 0.6,
                        "unusual_merchant_weight": 0.5,
                        "unusual_time_weight": 0.4
                    }
                },
                "AO": {
                    "velocity_thresholds": {
                        "max_tx_per_hour": 8,
                        "max_tx_per_day": 40,
                        "max_amount_per_day": 500000,  # Kwanzas
                        "max_different_merchants_per_hour": 4
                    },
                    "amount_thresholds": {
                        "suspicious_threshold": 250000,  # Kwanzas
                        "high_risk_threshold": 1000000   # Kwanzas
                    },
                    "location_rules": {
                        "max_distance_km_per_hour": 300,
                        "suspicious_countries": ["NG", "UG", "TR", "RU", "ZA", "NA"]
                    },
                    "device_rules": {
                        "trusted_devices_required": False,
                        "max_devices_per_user": 5
                    },
                    "specific_merchants_limits": {
                        "gambling": {"max_per_day": 50000},
                        "crypto": {"max_per_day": 100000},
                        "foreign_exchange": {"max_per_day": 200000}
                    }
                },
                "BR": {
                    "velocity_thresholds": {
                        "max_tx_per_hour": 12,
                        "max_tx_per_day": 60,
                        "max_amount_per_day": 20000,  # Reais
                        "max_different_merchants_per_hour": 6
                    },
                    "amount_thresholds": {
                        "suspicious_threshold": 5000,   # Reais
                        "high_risk_threshold": 10000    # Reais
                    },
                    "pix_specific_rules": {
                        "max_pix_per_hour": 20,
                        "max_pix_per_day": 100,
                        "max_pix_amount_per_day": 30000,
                        "max_new_pix_keys_per_day": 3,
                        "suspicious_pix_pattern": "multiple_keys_short_period"
                    },
                    "location_rules": {
                        "max_distance_km_per_hour": 800,
                        "suspicious_states": ["RR", "AC", "AP"],
                        "suspicious_borders": True
                    }
                },
                "MZ": {
                    "velocity_thresholds": {
                        "max_tx_per_hour": 8,
                        "max_tx_per_day": 40,
                        "max_amount_per_day": 50000,  # Meticais
                        "max_different_merchants_per_hour": 4
                    },
                    "amount_thresholds": {
                        "suspicious_threshold": 20000,  # Meticais
                        "high_risk_threshold": 40000    # Meticais
                    },
                    "location_rules": {
                        "max_distance_km_per_hour": 300,
                        "suspicious_countries": ["ZA", "ZW", "TZ", "MW"]
                    },
                    "mobile_money_rules": {
                        "max_cash_out_per_day": 20000,
                        "max_transfers_per_day": 10,
                        "suspicious_pattern": "multiple_cashin_cashout"
                    }
                },
                "PT": {
                    "velocity_thresholds": {
                        "max_tx_per_hour": 10,
                        "max_tx_per_day": 50,
                        "max_amount_per_day": 5000,  # Euros
                        "max_different_merchants_per_hour": 5
                    },
                    "amount_thresholds": {
                        "suspicious_threshold": 1000,  # Euros
                        "high_risk_threshold": 3000    # Euros
                    },
                    "location_rules": {
                        "max_distance_km_per_hour": 800,
                        "suspicious_countries": ["NG", "UG", "TR", "RU", "BY"]
                    },
                    "behavior_rules": {
                        "unusual_time_window_start": 0,  # 00:00
                        "unusual_time_window_end": 5,    # 05:00
                        "gdpr_compliance": True
                    }
                }
            }
        
        try:
            # Carregar regras de arquivo
            import yaml
            with open(self.rules_path, 'r') as file:
                return yaml.safe_load(file)
        except Exception as e:
            logger.error(f"Erro ao carregar regras do arquivo: {str(e)}. Usando regras padrão.")
            return self._load_rules()
    
    def _load_model(self):
        """
        Carrega o modelo de ML para detecção de fraudes em transações.
        
        Returns:
            object: Modelo de ML carregado
        """
        if not self.model_path:
            # Modelo simulado simples
            class DummyTransactionModel:
                def predict_proba(self, features):
                    # Simula probabilidades baseado em features
                    import random
                    # Aumenta probabilidade para valores grandes ou transações de alto volume
                    amount = features.get('amount', 0)
                    tx_count = features.get('tx_count_1h', 0)
                    unusual_factors = features.get('unusual_factors', 0)
                    
                    base_prob = 0.1
                    if amount > 1000:
                        base_prob += 0.2
                    if tx_count > 5:
                        base_prob += 0.2
                    if unusual_factors > 2:
                        base_prob += 0.3
                        
                    # Adiciona aleatoriedade
                    return [min(0.95, max(0.05, base_prob + random.uniform(-0.1, 0.1)))]
            
            return DummyTransactionModel()
        
        try:
            # Carregar modelo real
            import joblib
            return joblib.load(self.model_path)
        except Exception as e:
            logger.error(f"Erro ao carregar modelo: {str(e)}. Usando modelo simplificado.")
            return self._load_model()
    
    def on_consumer_start(self):
        """Ações ao iniciar o consumidor."""
        logger.info("Consumidor de análise de transações iniciado. Processando eventos...")
    
    def on_consumer_stop(self):
        """Ações ao parar o consumidor."""
        self.executor.shutdown(wait=True)
        logger.info("Executor de threads encerrado.")
        
        # Log de estatísticas
        logger.info(f"Estatísticas de análise de transações:")
        logger.info(f"Total de transações processadas: {self.transaction_stats['total_transactions']}")
        logger.info(f"Transações suspeitas: {self.transaction_stats['suspicious_transactions']}")
        logger.info(f"Transações de alto risco: {self.transaction_stats['high_risk_transactions']}")
        logger.info(f"Transações bloqueadas: {self.transaction_stats['blocked_transactions']}")
        
        if self.transaction_stats['total_transactions'] > 0:
            suspicious_rate = (self.transaction_stats['suspicious_transactions'] / self.transaction_stats['total_transactions']) * 100
            logger.info(f"Taxa de transações suspeitas: {suspicious_rate:.2f}%")
    
    def _cleanup_memory(self):
        """
        Limpa periodicamente a memória de transações antigas.
        """
        while True:
            try:
                import time
                time.sleep(60)  # Verificar a cada minuto
                
                now = datetime.datetime.now()
                cutoff_time = now - datetime.timedelta(seconds=self.memory_window_seconds)
                
                with self.memory_lock:
                    for user_id, transactions in list(self.recent_transactions.items()):
                        # Filtrar apenas transações dentro da janela de tempo
                        recent_txs = [tx for tx in transactions if tx.get('timestamp', now) >= cutoff_time]
                        
                        if not recent_txs:
                            # Remover entrada se não houver transações recentes
                            del self.recent_transactions[user_id]
                        else:
                            # Atualizar lista para conter apenas transações recentes
                            self.recent_transactions[user_id] = recent_txs
            except Exception as e:
                logger.error(f"Erro durante limpeza de memória: {str(e)}")
    
    def process_event(self, topic: str, event: Dict[str, Any]) -> ProcessingResult:
        """
        Processa um evento de transação.
        
        Args:
            topic: Tópico de onde o evento foi recebido
            event: Evento de transação
            
        Returns:
            ProcessingResult: Resultado do processamento
        """
        try:
            # Padronização do formato da transação (diferentes fontes podem ter formatos diferentes)
            transaction = self._normalize_transaction_format(topic, event)
            
            # Extrair informações básicas
            tx_id = transaction.get('transaction_id', 'unknown')
            user_id = transaction.get('user_id', 'unknown')
            amount = transaction.get('amount', 0.0)
            currency = transaction.get('currency', 'unknown')
            tx_type = transaction.get('transaction_type', 'unknown')
            channel = transaction.get('channel', 'unknown')
            merchant = transaction.get('merchant', {})
            
            # Log de processamento
            logger.debug(f"Processando transação {tx_id} do usuário {user_id} " +
                         f"({amount} {currency}, tipo: {tx_type}, canal: {channel})")
            
            # Atualizar estatísticas
            self.transaction_stats['total_transactions'] += 1
            self.transaction_stats['transaction_types'][tx_type] = self.transaction_stats['transaction_types'].get(tx_type, 0) + 1
            self.transaction_stats['transaction_channels'][channel] = self.transaction_stats['transaction_channels'].get(channel, 0) + 1
            self.transaction_stats['transaction_volumes'][currency] += amount
            self.transaction_stats['transaction_counts'][f"{tx_type}_{currency}"] += 1
            
            # Obter transações recentes do usuário
            recent_txs = self._get_recent_user_transactions(user_id)
            
            # Atualizar memória de transações recentes
            if 'timestamp' not in transaction:
                transaction['timestamp'] = datetime.datetime.now()
            self._update_transaction_memory(user_id, transaction)
            
            # Analisar transação em busca de sinais de fraude
            fraud_signals, risk_score, is_suspicious, is_high_risk = self._analyze_transaction(transaction, recent_txs)
            
            # Registrar resultados da análise
            if is_suspicious:
                self.transaction_stats['suspicious_transactions'] += 1
                
                if is_high_risk:
                    self.transaction_stats['high_risk_transactions'] += 1
                    
                    # Verificar se deve bloquear a transação
                    if risk_score > 0.85:  # Limite para bloqueio automático
                        self.transaction_stats['blocked_transactions'] += 1
                
                # Registrar sinais de fraude identificados
                for signal in fraud_signals:
                    signal_type = signal.get('type', 'unknown')
                    self.transaction_stats['fraud_signals'][signal_type] = self.transaction_stats['fraud_signals'].get(signal_type, 0) + 1
                
                # Se for uma transação suspeita, gerar alerta
                if fraud_signals:
                    self._generate_fraud_alert(transaction, fraud_signals, risk_score)
            
            # Resultado do processamento
            return ProcessingResult.success_result(
                f"Transação {tx_id} processada com sucesso",
                data={
                    "transaction_id": tx_id,
                    "user_id": user_id,
                    "amount": amount,
                    "currency": currency,
                    "transaction_type": tx_type,
                    "channel": channel,
                    "fraud_signals": fraud_signals,
                    "risk_score": risk_score,
                    "is_suspicious": is_suspicious,
                    "is_high_risk": is_high_risk
                }
            )
            
        except KeyError as ke:
            error_msg = f"Erro ao processar evento: campo obrigatório não encontrado - {str(ke)}"
            logger.error(error_msg)
            return ProcessingResult.failure_result(error_msg, ke)
        
        except Exception as e:
            error_msg = f"Erro inesperado ao processar evento de transação: {str(e)}"
            logger.exception(error_msg)
            return ProcessingResult.failure_result(error_msg, e)    def _normalize_transaction_format(self, topic: str, event: Dict[str, Any]) -> Dict[str, Any]:
        """
        Padroniza o formato da transação de diferentes fontes.
        
        Args:
            topic: Tópico de origem da transação
            event: Evento original da transação
            
        Returns:
            Dict: Transação em formato padronizado
        """
        normalized = {}
        
        try:
            # Mapear campos comuns
            normalized['transaction_id'] = event.get('transaction_id') or event.get('tx_id') or event.get('id') or str(uuid.uuid4())
            normalized['user_id'] = event.get('user_id') or event.get('client_id') or event.get('customer_id') or 'unknown'
            normalized['amount'] = float(event.get('amount') or event.get('value') or event.get('tx_amount') or 0.0)
            normalized['currency'] = event.get('currency') or event.get('currency_code') or 'unknown'
            
            # Mapear detalhes da transação
            normalized['transaction_type'] = event.get('transaction_type') or event.get('type') or event.get('tx_type') or 'unknown'
            normalized['channel'] = event.get('channel') or event.get('source') or 'unknown'
            normalized['status'] = event.get('status') or 'unknown'
            
            # Copiar metadados se existirem
            if 'metadata' in event:
                normalized['metadata'] = event['metadata']
            else:
                normalized['metadata'] = {}
                
            # Adicionar origem da transação
            normalized['metadata']['source_topic'] = topic
                
            # Extrair informações de timestamp
            if 'timestamp' in event:
                normalized['timestamp'] = self._parse_timestamp(event['timestamp'])
            elif 'transaction_date' in event:
                normalized['timestamp'] = self._parse_timestamp(event['transaction_date'])
            elif 'date' in event:
                normalized['timestamp'] = self._parse_timestamp(event['date'])
            else:
                normalized['timestamp'] = datetime.datetime.now()
                
            # Extrair informações de localização
            normalized['location'] = event.get('location') or {}
            if 'location' not in normalized and ('country_code' in event or 'country' in event):
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
                
            # Extrair informações do comerciante/beneficiário
            normalized['merchant'] = event.get('merchant') or {}
            if 'merchant' not in event and ('merchant_id' in event or 'merchant_name' in event):
                normalized['merchant'] = {
                    'merchant_id': event.get('merchant_id'),
                    'merchant_name': event.get('merchant_name'),
                    'merchant_category': event.get('merchant_category') or event.get('category'),
                    'merchant_country': event.get('merchant_country')
                }
                
            # Formatar campos específicos com base no tópico de origem
            if 'payment_gateway' in topic:
                # Campos específicos do Payment Gateway
                normalized['payment_method'] = event.get('payment_method')
                normalized['card_details'] = {
                    'last4': event.get('card_last4') or event.get('last4'),
                    'bin': event.get('card_bin') or event.get('bin'),
                    'card_type': event.get('card_type'),
                    'issuer': event.get('card_issuer') or event.get('issuer')
                }
                
            elif 'mobile_money' in topic:
                # Campos específicos de Mobile Money
                normalized['wallet_id'] = event.get('wallet_id')
                normalized['mobile_number'] = event.get('mobile_number') or event.get('msisdn')
                normalized['transaction_fee'] = event.get('fee') or 0.0
                normalized['operator'] = event.get('operator')
                
            elif 'ecommerce' in topic:
                # Campos específicos de E-commerce
                normalized['order_id'] = event.get('order_id')
                normalized['products'] = event.get('products') or []
                normalized['shipping_address'] = event.get('shipping_address')
                normalized['billing_address'] = event.get('billing_address')
                
            # Copiar campos adicionais
            for key, value in event.items():
                if key not in normalized and not key.startswith('_'):
                    normalized[key] = value
                    
            return normalized
            
        except Exception as e:
            logger.error(f"Erro ao normalizar formato da transação: {str(e)}")
            # Em caso de erro, retornar o evento original com campos mínimos
            return {
                'transaction_id': event.get('transaction_id') or str(uuid.uuid4()),
                'user_id': event.get('user_id') or 'unknown',
                'amount': float(event.get('amount') or 0.0),
                'currency': event.get('currency') or 'unknown',
                'transaction_type': event.get('transaction_type') or 'unknown',
                'channel': event.get('channel') or 'unknown',
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
    
    def _update_transaction_memory(self, user_id: str, transaction: Dict[str, Any]):
        """
        Atualiza a memória de transações recentes para um usuário.
        
        Args:
            user_id: ID do usuário
            transaction: Transação a ser adicionada à memória
        """
        with self.memory_lock:
            # Adicionar transação à lista do usuário
            self.recent_transactions[user_id].append(transaction)
            
            # Remover transações antigas (fora da janela de memória)
            now = datetime.datetime.now()
            cutoff_time = now - datetime.timedelta(seconds=self.memory_window_seconds)
            
            self.recent_transactions[user_id] = [
                tx for tx in self.recent_transactions[user_id] 
                if tx.get('timestamp', now) >= cutoff_time
            ]
    
    def _get_recent_user_transactions(self, user_id: str) -> List[Dict[str, Any]]:
        """
        Recupera transações recentes de um usuário na janela de memória.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            List: Lista de transações recentes do usuário
        """
        with self.memory_lock:
            # Retornar cópia das transações recentes
            return self.recent_transactions.get(user_id, [])[:]
    
    def _get_user_profile(self, user_id: str) -> Dict[str, Any]:
        """
        Obtém ou cria o perfil de um usuário para análise de comportamento.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            Dict: Perfil do usuário com dados históricos
        """
        with self.profile_lock:
            if user_id in self.user_profiles:
                return self.user_profiles[user_id]
            
            # Se não temos perfil em cache, tentar carregar do banco de dados
            profile = self._load_user_profile_from_db(user_id)
            
            # Se não houver perfil no banco, criar um vazio
            if not profile:
                profile = {
                    'user_id': user_id,
                    'created_at': datetime.datetime.now(),
                    'updated_at': datetime.datetime.now(),
                    'usual_merchants': [],
                    'usual_locations': [],
                    'usual_devices': [],
                    'transaction_patterns': {
                        'average_amount': 0,
                        'std_dev_amount': 0,
                        'usual_hours': [],
                        'usual_days': [],
                        'transaction_frequency': {
                            'hourly': 0,
                            'daily': 0,
                            'weekly': 0,
                            'monthly': 0
                        }
                    },
                    'risk_score': 0.5,  # Score de risco inicial neutro
                    'last_transactions': []
                }
            
            # Armazenar no cache
            self.user_profiles[user_id] = profile
            return profile
    
    def _load_user_profile_from_db(self, user_id: str) -> Dict[str, Any]:
        """
        Carrega o perfil do usuário do banco de dados.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            Dict: Perfil do usuário ou None se não existir
        """
        if not self.user_history_db:
            return None
            
        try:
            # Implementação depende do tipo de banco de dados usado
            # Exemplo simples simulando uma consulta
            logger.debug(f"Carregando perfil do usuário {user_id} do banco de dados")
            
            # Aqui seria implementada a lógica real de consulta ao banco
            # Por enquanto, retornamos None para simular usuário sem perfil
            return None
            
        except Exception as e:
            logger.error(f"Erro ao carregar perfil do usuário do banco: {str(e)}")
            return None
    
    def _update_user_profile(self, user_id: str, transaction: Dict[str, Any], fraud_signals: List[Dict]):
        """
        Atualiza o perfil do usuário com base na transação atual.
        
        Args:
            user_id: ID do usuário
            transaction: Transação atual
            fraud_signals: Sinais de fraude detectados
        """
        # Esta função seria chamada após o processamento da transação
        # para atualizar o perfil comportamental do usuário
        pass  # Implementação completa em versão futura
    
    def _analyze_transaction(self, transaction: Dict[str, Any], recent_txs: List[Dict[str, Any]]) -> Tuple[List, float, bool, bool]:
        """
        Analisa uma transação em busca de sinais de fraude.
        
        Args:
            transaction: Transação normalizada
            recent_txs: Transações recentes do mesmo usuário
            
        Returns:
            Tuple: (sinais de fraude, score de risco, flag de suspeito, flag de alto risco)
        """
        fraud_signals = []
        is_suspicious = False
        is_high_risk = False
        risk_score = 0.0
        
        try:
            # Extrair informações básicas da transação
            user_id = transaction.get('user_id', 'unknown')
            amount = transaction.get('amount', 0.0)
            currency = transaction.get('currency', 'unknown')
            tx_type = transaction.get('transaction_type', 'unknown')
            channel = transaction.get('channel', 'unknown')
            timestamp = transaction.get('timestamp', datetime.datetime.now())
            location = transaction.get('location', {})
            device = transaction.get('device', {})
            merchant = transaction.get('merchant', {})
            
            # Obter regras relevantes com base na região/país
            country_code = location.get('country_code', 'unknown')
            region_code = self._get_region_code(country_code)
            rules = self._get_rules_for_region(region_code)
            
            # 1. Análise de velocidade (volume/frequência de transações)
            velocity_signals = self._analyze_velocity(transaction, recent_txs, rules)
            fraud_signals.extend(velocity_signals)
            
            # 2. Análise de valor da transação
            amount_signals = self._analyze_amount(transaction, recent_txs, rules)
            fraud_signals.extend(amount_signals)
            
            # 3. Análise de localização
            location_signals = self._analyze_location(transaction, recent_txs, rules)
            fraud_signals.extend(location_signals)
            
            # 4. Análise de dispositivo
            device_signals = self._analyze_device(transaction, recent_txs, rules)
            fraud_signals.extend(device_signals)
            
            # 5. Análise de comportamento
            behavior_signals = self._analyze_behavior(transaction, recent_txs, rules)
            fraud_signals.extend(behavior_signals)
            
            # 6. Análises específicas por tipo de transação
            if tx_type == 'pix' and 'pix_specific_rules' in rules:
                pix_signals = self._analyze_pix_transaction(transaction, recent_txs, rules)
                fraud_signals.extend(pix_signals)
                
            elif 'mobile_money' in channel.lower() and 'mobile_money_rules' in rules:
                mm_signals = self._analyze_mobile_money(transaction, recent_txs, rules)
                fraud_signals.extend(mm_signals)
                
            # 7. Análise baseada em ML
            ml_signals, ml_score = self._analyze_with_ml(transaction, recent_txs, fraud_signals)
            fraud_signals.extend(ml_signals)
            
            # Calcular pontuação de risco final com base em todos os sinais
            risk_factors = [signal.get('risk_factor', 0) for signal in fraud_signals]
            if risk_factors:
                base_risk = sum(risk_factors) / len(risk_factors)
                # Ajustar com score do ML (peso de 40%)
                risk_score = (base_risk * 0.6) + (ml_score * 0.4)
            else:
                risk_score = ml_score
            
            # Limitar entre 0 e 1
            risk_score = min(1.0, max(0.0, risk_score))
            
            # Determinar níveis de risco
            is_suspicious = risk_score >= 0.6 or len(fraud_signals) >= 2
            is_high_risk = risk_score >= 0.8 or len(fraud_signals) >= 4
            
            # Registrar nível de risco na transação para referência
            transaction['risk_score'] = risk_score
            transaction['fraud_signals'] = [signal.get('type') for signal in fraud_signals]
            transaction['is_suspicious'] = is_suspicious
            transaction['is_high_risk'] = is_high_risk
            
            return fraud_signals, risk_score, is_suspicious, is_high_risk
            
        except Exception as e:
            logger.error(f"Erro ao analisar transação: {str(e)}")
            return [], 0.5, False, False
    
    def _get_region_code(self, country_code: str) -> str:
        """
        Mapeia código de país para região.
        
        Args:
            country_code: Código ISO do país
            
        Returns:
            str: Código da região (AO, BR, MZ, PT) ou None
        """
        # Mapeamento simplificado de países para regiões suportadas
        region_mapping = {
            'AO': 'AO',  # Angola
            'BR': 'BR',  # Brasil
            'MZ': 'MZ',  # Moçambique
            'PT': 'PT',  # Portugal
            # Outros países mapeados para regiões apropriadas
            'CV': 'PT',  # Cabo Verde -> regras de Portugal
            'GW': 'PT',  # Guiné-Bissau -> regras de Portugal
            'ST': 'PT',  # São Tomé e Príncipe -> regras de Portugal
            'TL': 'PT',  # Timor-Leste -> regras de Portugal
        }
        
        return region_mapping.get(country_code)
    
    def _get_rules_for_region(self, region_code: str) -> Dict:
        """
        Obtém regras específicas para uma região.
        
        Args:
            region_code: Código da região
            
        Returns:
            Dict: Regras para a região ou regras padrão
        """
        if region_code and region_code in self.rules:
            return self.rules[region_code]
        return self.rules['default']
    
    def _generate_fraud_alert(self, transaction: Dict[str, Any], fraud_signals: List[Dict[str, Any]], risk_score: float):
        """
        Gera um alerta de fraude baseado em sinais detectados.
        
        Args:
            transaction: Transação original
            fraud_signals: Sinais de fraude detectados
            risk_score: Pontuação de risco calculada
        """
        try:
            # Determinar o nível de severidade com base no score de risco
            if risk_score >= 0.85:
                severity = 'high'
            elif risk_score >= 0.7:
                severity = 'medium'
            else:
                severity = 'low'
            
            # Extrair informações básicas
            tx_id = transaction.get('transaction_id', 'unknown')
            user_id = transaction.get('user_id', 'unknown')
            amount = transaction.get('amount', 0.0)
            currency = transaction.get('currency', 'unknown')
            
            # Criar mensagem de alerta
            alert_message = (
                f"Alerta de fraude em transação detectado: " +
                f"{tx_id} do usuário {user_id} " +
                f"({amount} {currency})"
            )
            
            # Detalhes do alerta
            alert_details = {
                'transaction_id': tx_id,
                'user_id': user_id,
                'amount': amount,
                'currency': currency,
                'transaction_type': transaction.get('transaction_type', 'unknown'),
                'channel': transaction.get('channel', 'unknown'),
                'merchant': transaction.get('merchant', {}),
                'location': transaction.get('location', {}),
                'device': transaction.get('device', {}),
                'timestamp': transaction.get('timestamp', datetime.datetime.now()).isoformat(),
                'fraud_signals': fraud_signals,
                'risk_score': risk_score,
                'detection_timestamp': datetime.datetime.now().isoformat(),
                'severity': severity
            }
            
            # Publicar alerta (em implementação futura, integração com EventProducer)
            logger.warning(f"ALERTA DE FRAUDE EM TRANSAÇÃO: {alert_message}")
            logger.warning(f"Severidade: {severity.upper()}, Score: {risk_score:.2f}")
            logger.warning(f"Sinais detectados: {len(fraud_signals)}")
            
            # TODO: Integrar com EventProducer para publicar alerta
            # from infrastructure.fraud_detection.event_producers import EventProducer, AlertEvent, EventType, EventSeverity, EventMetadata
            # event_producer = EventProducer()
            # alert_metadata = EventMetadata(
            #     region_code=transaction.get('metadata', {}).get('region_code', ''),
            #     tenant_id=transaction.get('metadata', {}).get('tenant_id', ''),
            #     user_id=user_id,
            #     source_module="transaction_analysis_consumer"
            # )
            # alert_event = AlertEvent(
            #     metadata=alert_metadata,
            #     alert_id=str(uuid.uuid4()),
            #     alert_type="transaction_fraud",
            #     alert_severity=EventSeverity.HIGH if severity == 'high' else (EventSeverity.MEDIUM if severity == 'medium' else EventSeverity.LOW),
            #     source_events=[{'transaction_id': tx_id}],
            #     alert_message=alert_message,
            #     alert_details=alert_details
            # )
            # event_producer.produce_event(alert_event, EventType.ALERT)
            # event_producer.flush()
            
        except Exception as e:
            logger.error(f"Erro ao gerar alerta de fraude: {str(e)}")


# Função de exemplo para iniciar o consumidor
def start_transaction_analysis_consumer(region_code=None):
    """
    Inicia o consumidor de análise de transações.
    
    Args:
        region_code: Código da região para filtrar eventos
    """
    consumer = TransactionAnalysisConsumer(
        consumer_group_id=f"transaction_analysis_consumer_{region_code or 'all'}",
        region_code=region_code
    )
    consumer.start()
    
    
if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Consumidor de eventos de análise de transações")
    parser.add_argument("--region", help="Código da região (AO, BR, MZ, PT)", default=None)
    parser.add_argument("--config", help="Caminho para o arquivo de configuração", default=None)
    parser.add_argument("--model", help="Caminho para o modelo de ML", default=None)
    parser.add_argument("--rules", help="Caminho para regras de validação", default=None)
    
    args = parser.parse_args()
    
    consumer = TransactionAnalysisConsumer(
        region_code=args.region,
        config_file=args.config,
        model_path=args.model,
        rules_path=args.rules
    )
    
    try:
        consumer.start()
    except KeyboardInterrupt:
        logger.info("Consumidor interrompido pelo usuário.")
    except Exception as e:
        logger.error(f"Erro ao executar consumidor: {str(e)}")
    finally:
        consumer.stop()