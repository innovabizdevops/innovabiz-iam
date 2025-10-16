#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Testes Unitários para Padrões Comportamentais Específicos de Angola

Este módulo contém testes unitários para a implementação de análise comportamental
específica para Angola, verificando a funcionalidade de detecção de anomalias regionais,
validação de números de telefone, análise de risco de localização e integração com
Bureau de Crédito.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import unittest
import datetime
import re
from unittest.mock import MagicMock, patch
from typing import Dict, Any, List

# Importar módulos a serem testados
from infrastructure.fraud_detection.event_consumers.regional.angola_behavioral_patterns import (
    AngolanBehaviorPatterns,
    create_angola_analyzer
)


class TestAngolanBehaviorPatterns(unittest.TestCase):
    """Testes unitários para padrões comportamentais específicos de Angola."""

    def setUp(self):
        """Configuração para cada teste."""
        # Configuração base para testes
        self.config = {
            "custom_rules": {
                "high_risk_threshold": 0.8,
                "medium_risk_threshold": 0.6
            }
        }
        
        # Criar instância do analisador
        self.analyzer = create_angola_analyzer(self.config)
    
    def test_initialization(self):
        """Teste de inicialização correta."""
        # Verificar se as zonas de alto risco foram carregadas
        self.assertTrue(len(self.analyzer.high_risk_zones) > 0)
        self.assertIn("cabinda", self.analyzer.high_risk_zones)
        self.assertIn("lunda norte", self.analyzer.high_risk_zones)
        
        # Verificar se as zonas urbanas foram carregadas
        self.assertTrue(len(self.analyzer.urban_zones) > 0)
        self.assertIn("luanda", self.analyzer.urban_zones)
        self.assertIn("benguela", self.analyzer.urban_zones)
        
        # Verificar operadoras de telecomunicações
        self.assertTrue(len(self.analyzer.trusted_telcos) > 0)
        self.assertIn("unitel", self.analyzer.trusted_telcos)
        self.assertIn("movicel", self.analyzer.trusted_telcos)
        
        # Verificar padrões de fraude
        self.assertTrue(len(self.analyzer.fraud_patterns) > 0)
        self.assertIn("mobile_money_fraud", self.analyzer.fraud_patterns)
    
    def test_analyze_location_risk_high_risk_zone(self):
        """Teste de análise de risco para zonas de alto risco."""
        # Localização em zona de alto risco (Cabinda)
        location_data = {
            "city": "Cabinda",
            "province": "Cabinda",
            "country_code": "AO",
            "latitude": -5.5500,
            "longitude": 12.2000
        }
        
        # Analisar risco
        result = self.analyzer.analyze_location_risk(location_data)
        
        # Verificações
        self.assertTrue(result["is_high_risk_zone"])
        self.assertGreater(result["risk_score"], 0.5)  # Score deve ser aumentado
        self.assertIn("high_risk_zone", result["risk_factors"])
        self.assertTrue(len(result["insights"]) > 0)
    
    def test_analyze_location_risk_urban_zone(self):
        """Teste de análise de risco para zonas urbanas (menor risco)."""
        # Localização em zona urbana (Luanda)
        location_data = {
            "city": "Luanda",
            "province": "Luanda",
            "country_code": "AO",
            "latitude": -8.8383,
            "longitude": 13.2344
        }
        
        # Analisar risco
        result = self.analyzer.analyze_location_risk(location_data)
        
        # Verificações
        self.assertTrue(result["is_urban_zone"])
        self.assertLessEqual(result["risk_score"], 0.5)  # Score deve ser reduzido
        self.assertFalse(result["is_high_risk_zone"])
        self.assertTrue(len(result["insights"]) > 0)
        self.assertNotIn("high_risk_zone", result.get("risk_factors", []))
    
    def test_analyze_location_risk_outside_angola(self):
        """Teste de análise de risco para localização fora de Angola."""
        # Localização fora de Angola
        location_data = {
            "city": "Maputo",
            "province": "Maputo",
            "country_code": "MZ",  # Moçambique
            "latitude": -25.9692,
            "longitude": 32.5732
        }
        
        # Analisar risco
        result = self.analyzer.analyze_location_risk(location_data)
        
        # Verificações
        self.assertGreater(result["risk_score"], 0.6)  # Score deve ser alto
        self.assertIn("access_from_outside_angola", result["risk_factors"])
    
    def test_is_rapid_location_change(self):
        """Teste de detecção de mudança rápida de localização."""
        # Localização anterior
        prev_location = {
            "city": "Luanda",
            "province": "Luanda",
            "country_code": "AO"
        }
        
        # Nova localização (província diferente)
        new_location = {
            "city": "Lubango",
            "province": "Huíla",
            "country_code": "AO"
        }
        
        # Verificar mudança rápida
        result = self.analyzer._is_rapid_location_change(prev_location, new_location)
        self.assertTrue(result)
        
        # Testar mudança internacional
        international_location = {
            "city": "Lisboa",
            "province": "Lisboa",
            "country_code": "PT"
        }
        result = self.analyzer._is_rapid_location_change(prev_location, international_location)
        self.assertTrue(result)
        
        # Testar mesma província (não é mudança rápida)
        same_province = {
            "city": "Viana",
            "province": "Luanda",
            "country_code": "AO"
        }
        result = self.analyzer._is_rapid_location_change(prev_location, same_province)
        self.assertFalse(result)
    
    def test_analyze_mobile_money_behavior_normal(self):
        """Teste de análise comportamental Mobile Money para transação normal."""
        # Dados de transação normal
        transaction_data = {
            "transaction_id": "tx123",
            "amount": 5000,  # 5.000 Kwanzas (abaixo do limite)
            "transaction_type": "transfer",
            "timestamp": datetime.datetime.now(),
            "recipient_id": "recipient123"
        }
        
        # Histórico do usuário com destinatário conhecido
        user_history = {
            "daily_transactions": [{"id": "tx100", "amount": 2000}],
            "daily_volume": 2000,
            "monthly_volume": 10000,
            "known_recipients": {"recipient123"}
        }
        
        # Analisar comportamento
        result = self.analyzer.analyze_mobile_money_behavior(transaction_data, user_history)
        
        # Verificações
        self.assertLessEqual(result["risk_score"], 0.5)  # Risco baixo
        self.assertEqual(result["recommendation"], "allow")
        self.assertEqual(len(result["risk_factors"]), 0)  # Sem fatores de risco
    
    def test_analyze_mobile_money_behavior_high_value(self):
        """Teste de análise comportamental Mobile Money para transação de alto valor."""
        # Dados de transação de alto valor
        transaction_data = {
            "transaction_id": "tx123",
            "amount": 80000,  # 80.000 Kwanzas (acima do limite)
            "transaction_type": "transfer",
            "timestamp": datetime.datetime.now(),
            "recipient_id": "new_recipient"
        }
        
        # Histórico do usuário
        user_history = {
            "daily_transactions": [{"id": "tx100", "amount": 10000}],
            "daily_volume": 10000,
            "monthly_volume": 50000,
            "known_recipients": {"recipient123", "recipient456"}  # Não inclui o destinatário atual
        }
        
        # Analisar comportamento
        result = self.analyzer.analyze_mobile_money_behavior(transaction_data, user_history)
        
        # Verificações
        self.assertGreater(result["risk_score"], 0.5)  # Risco elevado
        self.assertIn("large_single_transaction", result["risk_factors"])
        self.assertIn("exceeded_daily_limit", result["risk_factors"])
        self.assertNotEqual(result["recommendation"], "allow")
    
    def test_analyze_mobile_money_behavior_rapid_cash_out(self):
        """Teste de análise comportamental Mobile Money para padrão suspeito de cash-out rápido."""
        # Dados de transação de cash-out
        transaction_data = {
            "transaction_id": "tx123",
            "amount": 5000,
            "transaction_type": "cash_out",
            "timestamp": datetime.datetime.now(),
            "agent_id": "agent123"
        }
        
        # Histórico com cash-in recente (menos de 10 minutos atrás)
        recent_timestamp = datetime.datetime.now() - datetime.timedelta(minutes=5)
        user_history = {
            "daily_transactions": [{"id": "tx100", "amount": 5000, "type": "cash_in"}],
            "daily_volume": 5000,
            "monthly_volume": 20000,
            "recent_cash_in": {
                "timestamp": recent_timestamp,
                "amount": 5000,
                "agent_id": "agent123"  # Mesmo agente
            }
        }
        
        # Analisar comportamento
        result = self.analyzer.analyze_mobile_money_behavior(transaction_data, user_history)
        
        # Verificações
        self.assertGreater(result["risk_score"], 0.6)  # Risco elevado
        self.assertIn("rapid_cash_in_cash_out", result["risk_factors"])
        self.assertIn("same_agent_cash_in_out", result["risk_factors"])
        self.assertEqual(result["recommendation"], "review")
    
    def test_validate_angola_phone_valid(self):
        """Teste de validação de número de telefone angolano válido."""
        # Número válido Unitel
        phone_number = "+244912345678"  # Formato Unitel (91)
        
        # Validar número
        result = self.analyzer.validate_angola_phone(phone_number)
        
        # Verificações
        self.assertTrue(result["is_valid"])
        self.assertEqual(result["operator"], "unitel")
        self.assertTrue(len(result["insights"]) > 0)
    
    def test_validate_angola_phone_invalid(self):
        """Teste de validação de número de telefone angolano inválido."""
        # Número inválido
        phone_number = "+244712345678"  # Prefixo inexistente
        
        # Validar número
        result = self.analyzer.validate_angola_phone(phone_number)
        
        # Verificações
        self.assertFalse(result["is_valid"])
        self.assertIsNone(result["operator"])
    
    def test_validate_angola_phone_normalized(self):
        """Teste de normalização de número de telefone."""
        # Número com formato diferente mas válido
        phone_number = "00244912345678"  # Com prefixo internacional 00
        
        # Validar número
        result = self.analyzer.validate_angola_phone(phone_number)
        
        # Verificações
        self.assertTrue(result["is_valid"])
        self.assertEqual(result["operator"], "unitel")
    
    def test_analyze_device_context_common_device(self):
        """Teste de análise de contexto de dispositivo comum em Angola."""
        # Dispositivo comum em Angola
        device_data = {
            "id": "device123",
            "model": "tecno spark 7",  # Marca comum em Angola
            "os": "android",
            "browser": {
                "name": "chrome",
                "version": 95
            },
            "network": "unitel"  # Operadora confiável
        }
        
        # Histórico do usuário
        user_history = {
            "usual_device_os": "android"
        }
        
        # Analisar contexto do dispositivo
        result = self.analyzer.analyze_device_context(device_data, user_history)
        
        # Verificações
        self.assertTrue(result["is_common_device"])
        self.assertLessEqual(result["risk_score"], 0.4)  # Risco baixo
        self.assertEqual(len(result["risk_factors"]), 0)
    
    def test_analyze_device_context_unusual_device(self):
        """Teste de análise de contexto de dispositivo incomum em Angola."""
        # Dispositivo não comum em Angola
        device_data = {
            "id": "device456",
            "model": "iphone 13",  # Menos comum em Angola
            "os": "ios",
            "browser": {
                "name": "safari",
                "version": 15
            },
            "network": "unknown_network"  # Rede desconhecida
        }
        
        # Histórico do usuário
        user_history = {
            "usual_device_os": "android"  # Diferente do atual
        }
        
        # Analisar contexto do dispositivo
        result = self.analyzer.analyze_device_context(device_data, user_history)
        
        # Verificações
        self.assertFalse(result["is_common_device"])
        self.assertGreater(result["risk_score"], 0.3)
        self.assertIn("os_platform_change", result["risk_factors"])
        self.assertIn("unknown_network", result["risk_factors"])
    
    def test_analyze_bureau_credito_integration_good_score(self):
        """Teste de análise de dados do Bureau de Crédito com bom score."""
        # Dados do usuário
        user_data = {
            "user_id": "user123",
            "full_name": "João Silva"
        }
        
        # Dados do bureau com bom score
        bureau_data = {
            "credit_score": 800,  # Bom score (0-1000)
            "payment_defaults": 0,  # Sem inadimplências
            "active_loans": 1,     # Poucos empréstimos
            "recent_inquiries": 1  # Poucas consultas
        }
        
        # Analisar dados do Bureau
        result = self.analyzer.analyze_bureau_credito_integration(user_data, bureau_data)
        
        # Verificações
        self.assertLess(result["risk_score"], 0.5)
        self.assertEqual(result["credit_status"], "good")
        self.assertEqual(len(result["risk_factors"]), 0)
    
    def test_analyze_bureau_credito_integration_bad_score(self):
        """Teste de análise de dados do Bureau de Crédito com score ruim."""
        # Dados do usuário
        user_data = {
            "user_id": "user456",
            "full_name": "Pedro Santos"
        }
        
        # Dados do bureau com score ruim
        bureau_data = {
            "credit_score": 250,  # Score baixo
            "payment_defaults": 3,  # Várias inadimplências
            "active_loans": 5,     # Muitos empréstimos
            "recent_inquiries": 10  # Muitas consultas
        }
        
        # Analisar dados do Bureau
        result = self.analyzer.analyze_bureau_credito_integration(user_data, bureau_data)
        
        # Verificações
        self.assertGreater(result["risk_score"], 0.7)
        self.assertEqual(result["credit_status"], "bad")
        self.assertIn("low_credit_score", result["risk_factors"])
        self.assertIn("payment_defaults", result["risk_factors"])
        self.assertIn("multiple_active_loans", result["risk_factors"])
        self.assertIn("multiple_credit_inquiries", result["risk_factors"])
    
    def test_get_regional_rules(self):
        """Teste de obtenção de regras regionais específicas para Angola."""
        # Obter regras regionais
        rules = self.analyzer.get_regional_rules()
        
        # Verificações
        self.assertIn("auth_thresholds", rules)
        self.assertIn("session_thresholds", rules)
        self.assertIn("transaction_thresholds", rules)
        self.assertIn("location_patterns", rules)
        self.assertIn("mobile_device_patterns", rules)
        self.assertIn("connection_patterns", rules)
        self.assertIn("bureau_credit_thresholds", rules)
        
        # Verificar regras específicas para Angola
        self.assertEqual(rules["auth_thresholds"]["max_failed_attempts"], 4)
        self.assertEqual(rules["session_thresholds"]["max_concurrent_sessions"], 2)
        self.assertIn("high_risk_zones", rules["location_patterns"])
        self.assertTrue(isinstance(rules["location_patterns"]["high_risk_zones"], list))


if __name__ == "__main__":
    unittest.main()