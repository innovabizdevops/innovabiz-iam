#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Testes Unitários para Adaptador de Bureau de Créditos

Este módulo contém testes unitários para os adaptadores de Bureau de Créditos,
verificando a funcionalidade de consultas de relatórios de crédito, consultas de score,
e normalização de dados para diferentes regiões.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import unittest
import json
import datetime
from unittest.mock import MagicMock, patch, ANY
import requests
from typing import Dict, Any, List

# Importar módulos a serem testados
from infrastructure.fraud_detection.event_consumers.adapters.bureau_credito_adapter import (
    BaseBureauCreditoAdapter,
    AngolaBureauCreditoAdapter,
    BrazilBureauCreditoAdapter,
    create_bureau_adapter
)


class TestBaseBureauCreditoAdapter(unittest.TestCase):
    """Testes unitários para a classe base do adaptador de Bureau de Créditos."""

    def setUp(self):
        """Configuração para cada teste."""
        # Criar classe concreta derivada para testar métodos não abstratos
        class ConcreteBureauAdapter(BaseBureauCreditoAdapter):
            def _setup_credentials(self):
                pass
                
            def get_credit_report(self, user_data):
                return {"mock": "report"}
                
            def check_credit_score(self, user_id):
                return {"mock": "score"}
        
        # Configuração para testes
        self.config = {
            "base_url": "https://api.bureau-test.com",
            "api_key": "test_key",
            "api_secret": "test_secret",
            "timeout": 10
        }
        
        # Instância para testes
        self.adapter = ConcreteBureauAdapter(self.config)
    
    def test_initialization(self):
        """Teste de inicialização correta."""
        self.assertEqual(self.adapter.base_url, "https://api.bureau-test.com")
        self.assertEqual(self.adapter.api_key, "test_key")
        self.assertEqual(self.adapter.api_secret, "test_secret")
        self.assertEqual(self.adapter.timeout, 10)
        self.assertEqual(self.adapter.headers["Content-Type"], "application/json")
        self.assertEqual(self.adapter.headers["Accept"], "application/json")
    
    def test_validate_response_success(self):
        """Teste de validação de resposta bem-sucedida."""
        # Criar mock de resposta
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"data": "test"}
        
        # Validar resposta
        result = self.adapter.validate_response(mock_response)
        
        # Verificações
        self.assertTrue(result["success"])
        self.assertEqual(result["data"], {"data": "test"})
        self.assertIsNone(result["error"])
        self.assertEqual(result["status_code"], 200)
    
    def test_validate_response_error(self):
        """Teste de validação de resposta com erro."""
        # Criar mock de resposta
        mock_response = MagicMock()
        mock_response.status_code = 403
        mock_response.text = "Forbidden"
        
        # Validar resposta
        result = self.adapter.validate_response(mock_response)
        
        # Verificações
        self.assertFalse(result["success"])
        self.assertIsNone(result["data"])
        self.assertEqual(result["error"]["code"], 403)
        self.assertEqual(result["error"]["message"], "Forbidden")
    
    def test_validate_response_parse_error(self):
        """Teste de validação de resposta com erro de parsing."""
        # Criar mock de resposta
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.side_effect = ValueError("Invalid JSON")
        
        # Validar resposta
        result = self.adapter.validate_response(mock_response)
        
        # Verificações
        self.assertFalse(result["success"])
        self.assertIsNone(result["data"])
        self.assertEqual(result["error"]["code"], "PARSE_ERROR")
        self.assertEqual(result["error"]["message"], "Invalid JSON")
    
    def test_generate_hmac_signature(self):
        """Teste de geração de assinatura HMAC."""
        # Dados para assinar
        payload = {"user_id": "test", "timestamp": 123456789}
        secret = "my_secret_key"
        
        # Gerar assinatura
        signature = self.adapter._generate_hmac_signature(payload, secret)
        
        # Verificar que a assinatura é uma string e não está vazia
        self.assertTrue(isinstance(signature, str))
        self.assertTrue(len(signature) > 0)
        
        # Verificar que assinaturas para o mesmo payload são iguais
        signature2 = self.adapter._generate_hmac_signature(payload, secret)
        self.assertEqual(signature, signature2)
        
        # Verificar que mudanças no payload ou secret geram assinaturas diferentes
        different_payload = {"user_id": "different", "timestamp": 123456789}
        different_signature = self.adapter._generate_hmac_signature(different_payload, secret)
        self.assertNotEqual(signature, different_signature)
        
        different_secret = "different_secret"
        different_signature = self.adapter._generate_hmac_signature(payload, different_secret)
        self.assertNotEqual(signature, different_signature)


class TestAngolaBureauCreditoAdapter(unittest.TestCase):
    """Testes unitários para o adaptador de Bureau de Créditos de Angola."""

    def setUp(self):
        """Configuração para cada teste."""
        # Configuração para testes
        self.config = {
            "base_url": "https://api.bureau-angola.co.ao",
            "api_key": "angola_key",
            "api_secret": "angola_secret",
            "api_version": "2.0",
            "banking_license": "AO-BNA-12345",
            "client_reference": "innovabiz-test",
            "entity_name": "InnovaBiz IAM Test",
            "purpose_code": "FRAUD_DETECTION"
        }
        
        # Instância para testes
        self.adapter = AngolaBureauCreditoAdapter(self.config)
    
    def test_setup_credentials(self):
        """Teste de configuração de credenciais específicas para Angola."""
        # Verificar headers específicos
        self.assertEqual(self.adapter.headers["X-Angola-Bureau-ApiKey"], "angola_key")
        self.assertEqual(self.adapter.headers["X-Angola-Bureau-Version"], "2.0")
        self.assertEqual(self.adapter.headers["X-Angola-Banking-License"], "AO-BNA-12345")
        
        # Verificar região
        self.assertEqual(self.adapter.region, "AO")
        
        # Verificar endpoints
        self.assertTrue(self.adapter.endpoints["credit_report"].startswith("https://api.bureau-angola.co.ao"))
        self.assertTrue(self.adapter.endpoints["credit_score"].startswith("https://api.bureau-angola.co.ao"))
        self.assertIn("angola", self.adapter.endpoints["credit_report"])
    
    @patch('infrastructure.fraud_detection.event_consumers.adapters.bureau_credito_adapter.requests.post')
    def test_get_credit_report(self, mock_post):
        """Teste de obtenção de relatório de crédito."""
        # Configurar mock de resposta
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "credit_score": {
                "score": 750,
                "range": {"min": 0, "max": 1000},
                "risk_category": "low"
            },
            "payment_history": [
                {"id": "loan1", "status": "PAID", "date": "2025-01-15"},
                {"id": "loan2", "status": "ACTIVE", "date": "2025-06-01"}
            ],
            "loans": [
                {"id": "loan2", "status": "ACTIVE", "amount": 50000, "currency": "AOA"}
            ]
        }
        mock_post.return_value = mock_response
        
        # Dados do usuário
        user_data = {
            "user_id": "angola_user",
            "full_name": "António Silva",
            "id_bilhete": "123456789AO",
            "nif": "123456789",
            "phone_number": "+244912345678"
        }
        
        # Obter relatório
        result = self.adapter.get_credit_report(user_data)
        
        # Verificações
        self.assertTrue(result["success"])
        self.assertIsNotNone(result["data"])
        
        # Verificar que o post foi chamado com os argumentos corretos
        mock_post.assert_called_once()
        args, kwargs = mock_post.call_args
        
        # Verificar URL do endpoint
        self.assertEqual(args[0], self.adapter.endpoints["credit_report"])
        
        # Verificar headers
        self.assertIn("X-Angola-Bureau-Timestamp", kwargs["headers"])
        self.assertIn("X-Angola-Bureau-Signature", kwargs["headers"])
        
        # Verificar payload
        payload = kwargs["json"]
        self.assertEqual(payload["region"], "AO")
        self.assertEqual(payload["subject"]["user_id"], "angola_user")
        self.assertEqual(payload["subject"]["full_name"], "António Silva")
        self.assertIn("identification", payload["subject"])
        self.assertEqual(len(payload["subject"]["identification"]), 2)  # BI e NIF
    
    @patch('infrastructure.fraud_detection.event_consumers.adapters.bureau_credito_adapter.requests.post')
    def test_check_credit_score(self, mock_post):
        """Teste de consulta de score de crédito."""
        # Configurar mock de resposta
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "score": 820,
            "range": {"min": 0, "max": 1000},
            "risk_category": "low",
            "evaluation_date": "2025-08-20T10:15:30Z"
        }
        mock_post.return_value = mock_response
        
        # Consultar score
        result = self.adapter.check_credit_score("angola_user_123")
        
        # Verificações
        self.assertTrue(result["success"])
        self.assertIsNotNone(result["data"])
        self.assertEqual(result["data"]["credit_score"], 820)
        self.assertEqual(result["data"]["risk_category"], "low")
        
        # Verificar que o post foi chamado com os argumentos corretos
        mock_post.assert_called_once()
        args, kwargs = mock_post.call_args
        
        # Verificar URL do endpoint
        self.assertEqual(args[0], self.adapter.endpoints["credit_score"])
        
        # Verificar payload
        payload = kwargs["json"]
        self.assertEqual(payload["user_id"], "angola_user_123")
        self.assertEqual(payload["region"], "AO")
        self.assertEqual(payload["request_type"], "score_only")
    
    def test_prepare_angola_payload(self):
        """Teste de preparação de payload específico para Angola."""
        # Dados do usuário
        user_data = {
            "user_id": "angola_user_456",
            "full_name": "Maria Costa",
            "id_bilhete": "987654321AO",
            "nif": "987654321",
            "passport": "AB123456",
            "phone_number": "+244923456789",
            "address": "Rua 1, Luanda",
            "date_of_birth": "1990-01-01"
        }
        
        # Preparar payload
        payload = self.adapter._prepare_angola_payload(user_data)
        
        # Verificações
        self.assertEqual(payload["region"], "AO")
        self.assertEqual(payload["requesting_entity"], "InnovaBiz IAM Test")
        self.assertEqual(payload["purpose_code"], "FRAUD_PREVENTION")
        self.assertEqual(payload["subject"]["user_id"], "angola_user_456")
        self.assertEqual(payload["subject"]["full_name"], "Maria Costa")
        
        # Verificar documentos de identificação
        identification = payload["subject"]["identification"]
        self.assertEqual(len(identification), 3)  # BI, NIF, Passport
        
        # Verificar que cada tipo de documento está presente
        id_types = [doc["type"] for doc in identification]
        self.assertIn("BI", id_types)
        self.assertIn("NIF", id_types)
        self.assertIn("PASSPORT", id_types)
        
        # Verificar dados adicionais
        self.assertEqual(payload["subject"]["phone_number"], "+244923456789")
        self.assertEqual(payload["subject"]["address"], "Rua 1, Luanda")
        self.assertEqual(payload["subject"]["date_of_birth"], "1990-01-01")
    
    def test_normalize_angola_report(self):
        """Teste de normalização de relatório do Bureau de Angola."""
        # Dados brutos do relatório
        report_data = {
            "credit_score": {
                "score": 650,
                "range": {"min": 0, "max": 1000},
                "risk_category": "medium"
            },
            "payment_history": [
                {"id": "loan1", "status": "PAID", "date": "2024-12-15"},
                {"id": "loan2", "status": "DEFAULT", "date": "2025-02-01"},
                {"id": "loan3", "status": "DEFAULT", "date": "2025-04-15"}
            ],
            "loans": [
                {"id": "loan2", "status": "ACTIVE", "amount": 30000, "currency": "AOA"},
                {"id": "loan3", "status": "ACTIVE", "amount": 50000, "currency": "AOA"},
                {"id": "loan4", "status": "ACTIVE", "amount": 20000, "currency": "AOA"}
            ],
            "inquiries": [
                {"inquiry_id": "inq1", "inquiry_date": "2025-05-15T10:30:00"},
                {"inquiry_id": "inq2", "inquiry_date": "2025-06-20T14:45:00"},
                {"inquiry_id": "inq3", "inquiry_date": "2025-07-10T09:15:00"},
                {"inquiry_id": "inq4", "inquiry_date": "2025-08-05T11:00:00"},
                {"inquiry_id": "inq5", "inquiry_date": "2025-08-18T16:20:00"}
            ],
            "account_summary": {
                "total_accounts": 5,
                "open_accounts": 3,
                "closed_accounts": 2
            },
            "risk_indicators": [
                {"code": "MULTIPLE_LOANS", "severity": "medium"},
                {"code": "PAYMENT_DEFAULTS", "severity": "high"}
            ],
            "score_history": [
                {"date": "2025-01-01", "score": 720},
                {"date": "2025-04-01", "score": 680},
                {"date": "2025-07-01", "score": 650}
            ]
        }
        
        # Normalizar relatório
        normalized = self.adapter._normalize_angola_report(report_data)
        
        # Verificações
        self.assertEqual(normalized["credit_score"], 650)
        self.assertEqual(normalized["payment_defaults"], 2)  # 2 inadimplências
        self.assertEqual(normalized["active_loans"], 3)  # 3 empréstimos ativos
        self.assertEqual(normalized["recent_inquiries"], 5)  # 5 consultas recentes
        self.assertEqual(normalized["account_summary"]["total_accounts"], 5)
        self.assertEqual(len(normalized["risk_indicators"]), 2)
        self.assertEqual(len(normalized["score_history"]), 3)
        self.assertEqual(len(normalized["payment_history"]), 3)


class TestCreateBureauAdapter(unittest.TestCase):
    """Testes unitários para a fábrica de adaptadores."""

    def setUp(self):
        """Configuração para cada teste."""
        self.config = {"base_url": "https://api.test.com", "api_key": "test_key"}
    
    def test_create_angola_adapter(self):
        """Teste de criação de adaptador para Angola."""
        adapter = create_bureau_adapter("AO", self.config)
        self.assertIsInstance(adapter, AngolaBureauCreditoAdapter)
        self.assertEqual(adapter.region, "AO")
    
    def test_create_brazil_adapter(self):
        """Teste de criação de adaptador para Brasil."""
        adapter = create_bureau_adapter("BR", self.config)
        self.assertIsInstance(adapter, BrazilBureauCreditoAdapter)
    
    def test_create_unsupported_region(self):
        """Teste de criação de adaptador para região não suportada."""
        with self.assertRaises(ValueError):
            create_bureau_adapter("XX", self.config)


if __name__ == "__main__":
    unittest.main()