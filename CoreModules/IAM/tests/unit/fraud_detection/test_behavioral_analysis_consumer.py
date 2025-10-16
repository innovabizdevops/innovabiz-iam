#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Testes Unitários para o Consumidor de Análise Comportamental

Este módulo contém testes unitários para o consumidor de análise comportamental,
verificando a funcionalidade de normalização de eventos, análise de anomalias,
geração de alertas, e outras funcionalidades críticas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import unittest
import datetime
import json
from unittest.mock import MagicMock, patch
import pandas as pd
import numpy as np
from typing import Dict, Any, List

# Importar módulos a serem testados
from infrastructure.fraud_detection.event_consumers.behavioral_analysis_consumer import (
    BehavioralAnalysisConsumer,
    ProcessingResult
)


class TestBehavioralAnalysisConsumer(unittest.TestCase):
    """Testes unitários para o consumidor de análise comportamental."""

    def setUp(self):
        """Configuração para cada teste."""
        # Mock da configuração
        self.config = {
            "topics": {
                "global": "global.events",
                "regional": {
                    "ao": "ao.events",
                    "br": "br.events",
                    "pt": "pt.events"
                }
            },
            "max_workers": 5,
            "region": "ao",  # Angola como região padrão para testes
            "cache_ttl": 3600,
            "model_path": "models/anomaly_detection.joblib",
            "alert_topic_pattern": "{region}.alerts.behavioral",
            "rules_path": "config/behavioral_rules.yaml",
            "thresholds": {
                "anomaly_score": 0.7,
                "suspicious_score": 0.5
            }
        }
        
        # Mock do Kafka producer
        self.mock_producer = MagicMock()
        
        # Mock do cache
        self.mock_cache = MagicMock()
        self.mock_cache.get.return_value = None  # Simular cache miss por padrão
        
        # Criação da instância do consumidor com dependências mockadas
        with patch('infrastructure.fraud_detection.event_consumers.behavioral_analysis_consumer.KafkaProducer') as mock_kafka_producer:
            mock_kafka_producer.return_value = self.mock_producer
            
            self.consumer = BehavioralAnalysisConsumer(
                config=self.config,
                cache=self.mock_cache
            )
            
        # Mock para o modelo ML
        self.consumer._load_ml_model = MagicMock()
        self.consumer.model = MagicMock()
        self.consumer.model.predict_proba.return_value = np.array([[0.8, 0.2]])
    
    def test_normalize_auth_event(self):
        """Teste de normalização de evento de autenticação."""
        # Dados de evento de autenticação simulados
        auth_event = {
            "event_id": "auth_123",
            "type": "authentication",
            "user": {
                "id": "user123",
                "email": "user@example.com"
            },
            "timestamp": "2025-08-20T10:30:00Z",
            "success": True,
            "ip_address": "192.168.1.1",
            "device": {
                "id": "device123",
                "os": "Android 12",
                "model": "Samsung Galaxy S22",
                "browser": {
                    "name": "Chrome",
                    "version": "100.0.0"
                }
            },
            "location": {
                "latitude": -8.8383,  # Coordenadas de Luanda, Angola
                "longitude": 13.2344,
                "city": "Luanda",
                "country_code": "AO"
            }
        }
        
        # Normalizar evento
        topic = "ao.events"
        normalized = self.consumer._normalize_event_format(topic, auth_event)
        
        # Verificações
        self.assertEqual(normalized["event_id"], "auth_123")
        self.assertEqual(normalized["user_id"], "user123")
        self.assertEqual(normalized["event_type"], "authentication")
        self.assertEqual(normalized["device_id"], "device123")
        self.assertIn("timestamp", normalized)
        self.assertIn("location", normalized)
        self.assertEqual(normalized["location"]["country_code"], "AO")
        self.assertEqual(normalized["region"], "ao")
        self.assertEqual(normalized["context"]["authentication"]["success"], True)
    
    def test_normalize_transaction_event(self):
        """Teste de normalização de evento de transação."""
        # Dados de evento de transação simulados
        transaction_event = {
            "transaction_id": "tx_456",
            "user_identifier": "user123",
            "timestamp": "2025-08-20T14:45:00Z",
            "amount": 1000,
            "currency": "AOA",  # Kwanza angolano
            "transaction_type": "payment",
            "status": "completed",
            "device_info": {
                "device_id": "device123",
                "os_version": "Android 12",
                "ip": "196.223.35.87"  # IP de Angola
            },
            "merchant": {
                "id": "merchant_789",
                "name": "Loja Angola",
                "category": "retail"
            },
            "location": {
                "lat": -8.8383,
                "long": 13.2344,
                "city": "Luanda",
                "country": "AO"
            }
        }
        
        # Normalizar evento
        topic = "ao.events.payments"
        normalized = self.consumer._normalize_event_format(topic, transaction_event)
        
        # Verificações
        self.assertEqual(normalized["event_id"], "tx_456")
        self.assertEqual(normalized["user_id"], "user123")
        self.assertEqual(normalized["event_type"], "transaction")
        self.assertEqual(normalized["device_id"], "device123")
        self.assertIn("timestamp", normalized)
        self.assertEqual(normalized["region"], "ao")
        self.assertEqual(normalized["context"]["transaction"]["amount"], 1000)
        self.assertEqual(normalized["context"]["transaction"]["currency"], "AOA")
        self.assertEqual(normalized["context"]["transaction"]["status"], "completed")
    
    def test_get_user_profile_new_user(self):
        """Teste de obtenção de perfil para novo usuário."""
        # Mock para cache miss
        self.mock_cache.get.return_value = None
        
        # Obter perfil
        user_id = "new_user_123"
        profile = self.consumer._get_user_profile(user_id)
        
        # Verificar que um novo perfil foi criado
        self.assertIsNotNone(profile)
        self.assertEqual(profile["user_id"], user_id)
        self.assertIn("created_at", profile)
        self.assertIn("auth_stats", profile)
        self.assertIn("transaction_stats", profile)
        self.assertIn("device_stats", profile)
        self.assertIn("location_stats", profile)
        
        # Verificar que o cache foi atualizado
        self.mock_cache.set.assert_called_once()
    
    def test_get_user_profile_existing_user(self):
        """Teste de obtenção de perfil para usuário existente."""
        # Mock para perfil existente no cache
        existing_profile = {
            "user_id": "existing_user",
            "created_at": "2025-08-01T00:00:00Z",
            "updated_at": "2025-08-19T23:59:59Z",
            "auth_stats": {"login_count": 10},
            "transaction_stats": {"total_count": 5},
            "device_stats": {"device_ids": ["device1", "device2"]},
            "location_stats": {"locations": ["Luanda", "Benguela"]}
        }
        self.mock_cache.get.return_value = existing_profile
        
        # Obter perfil
        user_id = "existing_user"
        profile = self.consumer._get_user_profile(user_id)
        
        # Verificar que o perfil existente foi retornado
        self.assertEqual(profile, existing_profile)
        
        # Verificar que o cache não foi atualizado
        self.mock_cache.set.assert_not_called()
    
    @patch('infrastructure.fraud_detection.event_consumers.behavioral_analysis_consumer.datetime')
    def test_analyze_auth_anomalies(self, mock_datetime):
        """Teste de análise de anomalias de autenticação."""
        # Configurar mock de datetime
        mock_now = datetime.datetime(2025, 8, 20, 10, 30, 0)
        mock_datetime.datetime.now.return_value = mock_now
        mock_datetime.datetime.fromisoformat.side_effect = lambda x: datetime.datetime.fromisoformat(x)
        
        # Criar evento normalizado
        event = {
            "event_id": "auth_123",
            "user_id": "user456",
            "event_type": "authentication",
            "timestamp": "2025-08-20T10:30:00Z",
            "device_id": "new_device_789",  # Dispositivo não usual
            "ip": "196.46.30.100",  # IP diferente do usual
            "region": "ao",
            "location": {
                "city": "Benguela",  # Cidade diferente do usual
                "country_code": "AO"
            },
            "context": {
                "authentication": {
                    "success": True,
                    "method": "password",
                    "attempts": 1
                }
            }
        }
        
        # Mock para perfil do usuário com padrões usuais diferentes
        user_profile = {
            "user_id": "user456",
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-08-19T23:59:59Z",
            "auth_stats": {
                "login_count": 50,
                "usual_devices": ["usual_device_1", "usual_device_2"],
                "usual_locations": ["Luanda"],
                "usual_ips": ["196.223.35.87"],
                "usual_auth_times": [8, 9, 17, 18],  # Horas do dia usuais
                "failed_attempts": 0
            }
        }
        
        # Analisar anomalias
        anomalies = self.consumer._analyze_auth_anomalies(event, user_profile)
        
        # Verificar detecção de anomalias
        self.assertGreater(len(anomalies), 0)
        self.assertIn("new_device", anomalies)
        self.assertIn("unusual_location", anomalies)
        
        # Verificar scores de anomalias
        self.assertGreater(anomalies["new_device"]["score"], 0.5)
        self.assertGreater(anomalies["unusual_location"]["score"], 0.5)
    
    def test_update_user_profile(self):
        """Teste de atualização de perfil do usuário."""
        # Perfil inicial
        initial_profile = {
            "user_id": "test_user",
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-08-19T23:59:59Z",
            "auth_stats": {
                "login_count": 10,
                "usual_devices": ["device1"],
                "usual_locations": ["Luanda"],
                "usual_ips": ["196.223.35.87"],
                "usual_auth_times": [8, 17],
                "failed_attempts": 0
            },
            "transaction_stats": {
                "total_count": 5,
                "total_amount": 5000,
                "average_amount": 1000
            }
        }
        
        # Evento para atualização
        event = {
            "event_id": "auth_123",
            "user_id": "test_user",
            "event_type": "authentication",
            "timestamp": "2025-08-20T10:30:00Z",
            "device_id": "device2",  # Novo dispositivo
            "ip": "196.46.30.100",  # Novo IP
            "region": "ao",
            "location": {
                "city": "Benguela",  # Nova localização
                "country_code": "AO"
            },
            "context": {
                "authentication": {
                    "success": True,
                    "method": "password",
                    "attempts": 1
                }
            }
        }
        
        # Atualizar perfil
        updated_profile = self.consumer._update_user_profile(event, initial_profile)
        
        # Verificar atualizações
        self.assertEqual(updated_profile["user_id"], "test_user")
        self.assertNotEqual(updated_profile["updated_at"], initial_profile["updated_at"])
        
        # Verificar estatísticas de autenticação
        auth_stats = updated_profile["auth_stats"]
        self.assertEqual(auth_stats["login_count"], 11)  # +1
        self.assertIn("device1", auth_stats["usual_devices"])
        self.assertIn("device2", auth_stats["usual_devices"])  # Novo dispositivo adicionado
        self.assertIn("Benguela", auth_stats["usual_locations"])  # Nova localização adicionada
        self.assertIn("196.46.30.100", auth_stats["usual_ips"])  # Novo IP adicionado
        self.assertIn(10, auth_stats["usual_auth_times"])  # Nova hora adicionada
    
    def test_generate_alert(self):
        """Teste de geração de alerta."""
        # Evento com anomalias
        event = {
            "event_id": "auth_123",
            "user_id": "alert_user",
            "event_type": "authentication",
            "timestamp": "2025-08-20T10:30:00Z",
            "device_id": "suspicious_device",
            "ip": "196.46.30.100",
            "region": "ao",
            "location": {
                "city": "Benguela",
                "country_code": "AO"
            },
            "context": {
                "authentication": {
                    "success": True,
                    "method": "password",
                    "attempts": 1
                }
            }
        }
        
        # Anomalias detectadas
        anomalies = {
            "new_device": {
                "score": 0.8,
                "description": "Novo dispositivo não reconhecido"
            },
            "unusual_location": {
                "score": 0.75,
                "description": "Localização não usual"
            }
        }
        
        # Gerar alerta
        alert = self.consumer._generate_behavioral_alert(event, anomalies, 0.8)
        
        # Verificar alerta
        self.assertEqual(alert["user_id"], "alert_user")
        self.assertEqual(alert["event_id"], "auth_123")
        self.assertEqual(alert["event_type"], "authentication")
        self.assertIn("alert_id", alert)
        self.assertIn("timestamp", alert)
        self.assertEqual(alert["region"], "ao")
        self.assertEqual(alert["risk_score"], 0.8)
        self.assertIn("anomalies", alert)
        self.assertEqual(len(alert["anomalies"]), 2)
        self.assertIn("new_device", alert["anomalies"])
        self.assertIn("unusual_location", alert["anomalies"])
    
    @patch('infrastructure.fraud_detection.event_consumers.behavioral_analysis_consumer.json')
    def test_send_alert(self, mock_json):
        """Teste de envio de alerta."""
        # Mock para json.dumps
        mock_json.dumps.return_value = '{"test": "alert"}'
        
        # Alerta simulado
        alert = {
            "alert_id": "alert_123",
            "user_id": "alert_user",
            "event_id": "auth_123",
            "event_type": "authentication",
            "timestamp": "2025-08-20T10:30:00Z",
            "region": "ao",
            "risk_score": 0.8,
            "anomalies": {
                "new_device": {
                    "score": 0.8,
                    "description": "Novo dispositivo não reconhecido"
                }
            }
        }
        
        # Enviar alerta
        self.consumer._send_behavioral_alert(alert)
        
        # Verificar chamada ao produtor Kafka
        self.mock_producer.send.assert_called_once()
        topic_arg = self.mock_producer.send.call_args[0][0]
        self.assertEqual(topic_arg, "ao.alerts.behavioral")
    
    @patch('infrastructure.fraud_detection.event_consumers.behavioral_analysis_consumer.BehavioralAnalysisConsumer._analyze_behavioral_anomalies')
    def test_process_event(self, mock_analyze):
        """Teste do método principal de processamento de eventos."""
        # Mock para análise de anomalias
        mock_anomalies = {
            "new_device": {"score": 0.8, "description": "Novo dispositivo"},
            "unusual_location": {"score": 0.75, "description": "Localização não usual"}
        }
        mock_analyze.return_value = (mock_anomalies, 0.8)
        
        # Evento simulado
        event = {
            "event_id": "auth_123",
            "type": "authentication",
            "user": {"id": "process_user"},
            "timestamp": "2025-08-20T10:30:00Z",
            "device": {"id": "device123"},
            "location": {"country_code": "AO"}
        }
        
        # Mock para get_user_profile
        self.consumer._get_user_profile = MagicMock()
        self.consumer._get_user_profile.return_value = {
            "user_id": "process_user",
            "auth_stats": {}
        }
        
        # Mock para update_user_profile
        self.consumer._update_user_profile = MagicMock()
        
        # Mock para generate_behavioral_alert
        self.consumer._generate_behavioral_alert = MagicMock()
        
        # Mock para send_behavioral_alert
        self.consumer._send_behavioral_alert = MagicMock()
        
        # Processar evento
        topic = "ao.events"
        result = self.consumer.process_event(topic, event)
        
        # Verificar resultado
        self.assertIsInstance(result, ProcessingResult)
        self.assertTrue(result.success)
        self.assertEqual(result.event_id, "auth_123")
        
        # Verificar chamadas aos métodos mock
        self.consumer._get_user_profile.assert_called_once()
        self.consumer._update_user_profile.assert_called_once()
        mock_analyze.assert_called_once()
        self.consumer._generate_behavioral_alert.assert_called_once()
        self.consumer._send_behavioral_alert.assert_called_once()


if __name__ == "__main__":
    unittest.main()