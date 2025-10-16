#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Testes unitários para os métodos de autenticação biométrica.

Este módulo testa os métodos de autenticação WebAuthn e impressão digital.

Author: InnovaBiz Tech Team
Date: 29/04/2025
"""

import unittest
from unittest.mock import MagicMock, patch
import json
import base64
import os
from datetime import datetime, timedelta

from backend.iam.domain.services.auth_methods.webauthn import WebAuthnAuthentication
from backend.iam.domain.services.auth_methods.fingerprint import FingerprintAuthentication
from backend.iam.domain.models.authentication_types import AuthenticationMethod


class TestWebAuthnAuthentication(unittest.TestCase):
    """Testes para autenticação WebAuthn."""
    
    def setUp(self):
        """Configurar ambiente de teste."""
        self.user_repository = MagicMock()
        self.credential_repository = MagicMock()
        
        # Criar instância de autenticação
        self.auth = WebAuthnAuthentication(
            user_repository=self.user_repository,
            credential_repository=self.credential_repository
        )
        
        # Configurar usuário de teste
        self.user_id = "user123"
        self.user = MagicMock()
        self.user.user_id = self.user_id
        self.user.username = "testuser"
        self.user.get_full_name.return_value = "Test User"
        self.user.auth_factors = []
        
        # Configurar repositório para retornar o usuário
        self.user_repository.find_by_id.return_value = self.user
    
    def test_method_type(self):
        """Testar tipo do método de autenticação."""
        self.assertEqual(self.auth.method_type, AuthenticationMethod.WEBAUTHN)
    
    @patch('webauthn.registration.generate_registration_options')
    def test_initialize(self, mock_generate_options):
        """Testar inicialização do registro WebAuthn."""
        # Configurar mock
        mock_generate_options.return_value = {"publicKey": {"challenge": "challenge123"}}
        
        # Chamar método
        success, data = self.auth.initialize(self.user_id, "My Device")
        
        # Verificar resultado
        self.assertTrue(success)
        self.assertIn("options", data)
        
        # Verificar que o método de geração de opções foi chamado
        mock_generate_options.assert_called_once()
        
        # Verificar que o desafio foi armazenado
        self.assertIn(self.user_id, self.auth._challenges)
    
    def test_initialize_invalid_user(self):
        """Testar inicialização com usuário inválido."""
        # Configurar mock para retornar None
        self.user_repository.find_by_id.return_value = None
        
        # Chamar método
        success, data = self.auth.initialize("invalid_user", "My Device")
        
        # Verificar resultado
        self.assertFalse(success)
        self.assertIn("error", data)
    
    @patch('webauthn.authentication.generate_authentication_options')
    def test_initialize_authentication(self, mock_generate_options):
        """Testar inicialização da autenticação WebAuthn."""
        # Configurar mock
        mock_generate_options.return_value = {"publicKey": {"challenge": "challenge123"}}
        
        # Configurar fator de autenticação no usuário
        self.user.auth_factors = [{
            "type": AuthenticationMethod.WEBAUTHN.value,
            "enabled": True,
            "identifier": "credential123",
            "secret": json.dumps({
                "id": "credential123",
                "public_key": base64.b64encode(b"public_key").decode('utf-8')
            })
        }]
        
        # Chamar método
        success, data = self.auth.initialize_authentication(self.user_id)
        
        # Verificar resultado
        self.assertTrue(success)
        self.assertIn("options", data)
        
        # Verificar que o método de geração de opções foi chamado
        mock_generate_options.assert_called_once()
        
        # Verificar que o desafio foi armazenado
        self.assertIn(self.user_id, self.auth._challenges)
    
    def test_is_enrolled(self):
        """Testar verificação de registro."""
        # Caso: Usuário não tem fatores
        self.assertFalse(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem fator WebAuthn
        self.user.auth_factors = [{
            "type": AuthenticationMethod.WEBAUTHN.value,
            "enabled": True,
            "identifier": "credential123"
        }]
        self.assertTrue(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem fator, mas desabilitado
        self.user.auth_factors = [{
            "type": AuthenticationMethod.WEBAUTHN.value,
            "enabled": False,
            "identifier": "credential123"
        }]
        self.assertFalse(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem outro tipo de fator
        self.user.auth_factors = [{
            "type": AuthenticationMethod.PASSWORD.value,
            "enabled": True,
            "identifier": "credential123"
        }]
        self.assertFalse(self.auth.is_enrolled(self.user_id))


class TestFingerprintAuthentication(unittest.TestCase):
    """Testes para autenticação por impressão digital."""
    
    def setUp(self):
        """Configurar ambiente de teste."""
        self.user_repository = MagicMock()
        self.credential_repository = MagicMock()
        
        # Criar instância de autenticação
        self.auth = FingerprintAuthentication(
            user_repository=self.user_repository,
            credential_repository=self.credential_repository
        )
        
        # Configurar usuário de teste
        self.user_id = "user123"
        self.user = MagicMock()
        self.user.user_id = self.user_id
        self.user.username = "testuser"
        self.user.auth_factors = []
        
        # Configurar repositório para retornar o usuário
        self.user_repository.find_by_id.return_value = self.user
        
        # Configurar ID do dispositivo
        self.device_id = "device123"
    
    def test_method_type(self):
        """Testar tipo do método de autenticação."""
        self.assertEqual(self.auth.method_type, AuthenticationMethod.BIOMETRIC)
    
    def test_initialize(self):
        """Testar inicialização do registro de impressão digital."""
        # Dados do dispositivo
        device_info = {
            "name": "My Phone",
            "model": "Pixel 7",
            "os": "Android 14"
        }
        
        # Chamar método
        success, data = self.auth.initialize(self.user_id, self.device_id, device_info)
        
        # Verificar resultado
        self.assertTrue(success)
        self.assertIn("device_id", data)
        self.assertIn("device_key", data)
        
        # Verificar que o fator foi adicionado ao usuário
        self.user.add_auth_factor.assert_called_once()
        call_args = self.user.add_auth_factor.call_args[1]
        self.assertEqual(call_args["factor_type"], AuthenticationMethod.BIOMETRIC.value)
        self.assertEqual(call_args["identifier"], self.device_id)
        
        # Verificar que o usuário foi atualizado no repositório
        self.user_repository.update.assert_called_once_with(self.user)
    
    def test_initialize_invalid_user(self):
        """Testar inicialização com usuário inválido."""
        # Configurar mock para retornar None
        self.user_repository.find_by_id.return_value = None
        
        # Chamar método
        success, data = self.auth.initialize("invalid_user", self.device_id)
        
        # Verificar resultado
        self.assertFalse(success)
        self.assertIn("error", data)
    
    def test_initialize_authentication(self):
        """Testar inicialização da autenticação por impressão digital."""
        # Configurar fator de autenticação no usuário
        device_key = base64.b64encode(os.urandom(32)).decode('utf-8')
        self.user.auth_factors = [{
            "type": AuthenticationMethod.BIOMETRIC.value,
            "enabled": True,
            "identifier": self.device_id,
            "secret": json.dumps({
                "key": device_key,
                "device": {
                    "device_id": self.device_id,
                    "device_name": "My Phone"
                }
            })
        }]
        
        # Chamar método
        success, data = self.auth.initialize_authentication(self.user_id, self.device_id)
        
        # Verificar resultado
        self.assertTrue(success)
        self.assertIn("challenge", data)
        
        # Verificar que o desafio foi armazenado
        self.assertIn(f"{self.user_id}:{self.device_id}", self.auth._challenges)
    
    def test_is_enrolled(self):
        """Testar verificação de registro."""
        # Caso: Usuário não tem fatores
        self.assertFalse(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem fator biométrico
        self.user.auth_factors = [{
            "type": AuthenticationMethod.BIOMETRIC.value,
            "enabled": True,
            "identifier": self.device_id
        }]
        self.assertTrue(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem fator, mas desabilitado
        self.user.auth_factors = [{
            "type": AuthenticationMethod.BIOMETRIC.value,
            "enabled": False,
            "identifier": self.device_id
        }]
        self.assertFalse(self.auth.is_enrolled(self.user_id))
        
        # Caso: Usuário tem outro tipo de fator
        self.user.auth_factors = [{
            "type": AuthenticationMethod.PASSWORD.value,
            "enabled": True,
            "identifier": "credential123"
        }]
        self.assertFalse(self.auth.is_enrolled(self.user_id))
    
    def test_get_registered_devices(self):
        """Testar listagem de dispositivos registrados."""
        # Configurar fatores de autenticação no usuário
        device1 = self.device_id
        device2 = "device456"
        
        self.user.auth_factors = [
            {
                "type": AuthenticationMethod.BIOMETRIC.value,
                "enabled": True,
                "identifier": device1,
                "secret": json.dumps({
                    "key": "secret1",
                    "device": {
                        "device_id": device1,
                        "device_name": "My Phone",
                        "device_model": "Pixel 7",
                        "device_os": "Android 14",
                        "registered_at": datetime.now().isoformat(),
                        "last_used_at": None
                    }
                })
            },
            {
                "type": AuthenticationMethod.BIOMETRIC.value,
                "enabled": True,
                "identifier": device2,
                "secret": json.dumps({
                    "key": "secret2",
                    "device": {
                        "device_id": device2,
                        "device_name": "My Tablet",
                        "device_model": "iPad Pro",
                        "device_os": "iOS 17",
                        "registered_at": datetime.now().isoformat(),
                        "last_used_at": None
                    }
                })
            },
            {
                "type": AuthenticationMethod.PASSWORD.value,
                "enabled": True,
                "identifier": "password"
            }
        ]
        
        # Chamar método
        devices = self.auth.get_registered_devices(self.user_id)
        
        # Verificar resultado
        self.assertEqual(len(devices), 2)
        device_ids = [d["device_id"] for d in devices]
        self.assertIn(device1, device_ids)
        self.assertIn(device2, device_ids)
        
        # Verificar que a chave secreta não está incluída
        for device in devices:
            self.assertNotIn("key", device)
    
    def test_remove_device(self):
        """Testar remoção de dispositivo."""
        # Configurar fatores de autenticação no usuário
        device1 = self.device_id
        device2 = "device456"
        
        self.user.auth_factors = [
            {
                "type": AuthenticationMethod.BIOMETRIC.value,
                "enabled": True,
                "identifier": device1,
                "secret": json.dumps({
                    "key": "secret1",
                    "device": {"device_id": device1}
                })
            },
            {
                "type": AuthenticationMethod.BIOMETRIC.value,
                "enabled": True,
                "identifier": device2,
                "secret": json.dumps({
                    "key": "secret2",
                    "device": {"device_id": device2}
                })
            }
        ]
        
        # Chamar método para remover device1
        result = self.auth.remove_device(self.user_id, device1)
        
        # Verificar resultado
        self.assertTrue(result)
        
        # Verificar que o usuário foi atualizado
        self.user_repository.update.assert_called_once_with(self.user)
        
        # Verificar que apenas device2 permanece
        self.assertEqual(len(self.user.auth_factors), 1)
        self.assertEqual(self.user.auth_factors[0]["identifier"], device2)
