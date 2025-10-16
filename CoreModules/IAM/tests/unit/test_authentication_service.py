#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Testes unitários para o serviço de autenticação do módulo IAM.

Este módulo contém testes para validar o funcionamento do serviço
de autenticação e suas principais funcionalidades.

Author: InnovaBiz Tech Team
Date: 29/04/2025
"""

import unittest
import uuid
from datetime import datetime, timedelta
from unittest.mock import Mock, MagicMock, patch

from backend.iam.domain.models.identity import User, UserStatus
from backend.iam.domain.models.session import Session, SessionStatus
from backend.iam.domain.models.authentication_types import AuthenticationMethod, AuthenticationStatus
from backend.iam.domain.services.authentication_service import AuthenticationService


class TestAuthenticationService(unittest.TestCase):
    """Testes para o serviço de autenticação."""
    
    def setUp(self):
        """Configuração de cada teste."""
        # Criar mocks para os repositórios
        self.user_repository = Mock()
        self.session_repository = Mock()
        self.credential_repository = Mock()
        
        # Instanciar o serviço de autenticação com os mocks
        self.auth_service = AuthenticationService(
            user_repository=self.user_repository,
            session_repository=self.session_repository,
            credential_repository=self.credential_repository
        )
        
        # Criar um usuário de teste
        self.test_user = User(
            username="testuser",
            email="testuser@example.com",
            id="user-1234"
        )
        self.test_user.status = UserStatus.ACTIVE.value
        
        # Mock para o método verify_password
        self.test_user.verify_password = Mock()
        self.test_user.record_login = Mock()
        
        # Configurar o repositório de usuários para retornar o usuário de teste
        self.user_repository.find_by_username.return_value = self.test_user
        self.user_repository.find_by_email.return_value = self.test_user
        self.user_repository.find_by_id.return_value = self.test_user
        
        # Configurar o repositório de sessões
        self.session_repository.create.side_effect = lambda x: x
        self.session_repository.update.side_effect = lambda x: x
        
    def test_authenticate_success(self):
        """Testa autenticação bem-sucedida."""
        # Configurar mock para verificação de senha bem-sucedida
        self.test_user.verify_password.return_value = True
        
        # Chamar o método de autenticação
        status, result = self.auth_service.authenticate(
            "testuser", 
            "correctpassword", 
            "127.0.0.1", 
            "Mozilla/5.0"
        )
        
        # Verificar resultado
        self.assertEqual(status, AuthenticationStatus.SUCCESS)
        self.assertIsNotNone(result)
        self.assertEqual(result["user_id"], "user-1234")
        
        # Verificar se os métodos esperados foram chamados
        self.user_repository.find_by_username.assert_called_once_with("testuser")
        self.test_user.verify_password.assert_called_once_with("correctpassword")
        self.test_user.record_login.assert_called_once()
        self.session_repository.create.assert_called_once()
        
    def test_authenticate_invalid_credentials(self):
        """Testa autenticação com credenciais inválidas."""
        # Configurar mock para verificação de senha falha
        self.test_user.verify_password.return_value = False
        
        # Chamar o método de autenticação
        status, result = self.auth_service.authenticate(
            "testuser", 
            "wrongpassword", 
            "127.0.0.1", 
            "Mozilla/5.0"
        )
        
        # Verificar resultado
        self.assertEqual(status, AuthenticationStatus.INVALID_CREDENTIALS)
        self.assertIsNone(result)
        
        # Verificar se os métodos esperados foram chamados
        self.user_repository.find_by_username.assert_called_once_with("testuser")
        self.test_user.verify_password.assert_called_once_with("wrongpassword")
        self.test_user.record_login.assert_called_once_with(False, "127.0.0.1")
        self.session_repository.create.assert_not_called()
        
    def test_authenticate_inactive_account(self):
        """Testa autenticação com conta inativa."""
        # Configurar usuário como inativo
        self.test_user.status = UserStatus.INACTIVE.value
        
        # Chamar o método de autenticação
        status, result = self.auth_service.authenticate(
            "testuser", 
            "anypassword", 
            "127.0.0.1", 
            "Mozilla/5.0"
        )
        
        # Verificar resultado
        self.assertEqual(status, AuthenticationStatus.ACCOUNT_DISABLED)
        self.assertIsNone(result)
        
        # Verificar se os métodos esperados foram chamados
        self.user_repository.find_by_username.assert_called_once_with("testuser")
        self.test_user.verify_password.assert_not_called()
        self.session_repository.create.assert_not_called()
        
    def test_authenticate_locked_account(self):
        """Testa autenticação com conta bloqueada."""
        # Configurar usuário como bloqueado
        self.test_user.status = UserStatus.LOCKED.value
        self.test_user.account_locked_until = datetime.now() + timedelta(minutes=30)
        
        # Chamar o método de autenticação
        status, result = self.auth_service.authenticate(
            "testuser", 
            "anypassword", 
            "127.0.0.1", 
            "Mozilla/5.0"
        )
        
        # Verificar resultado
        self.assertEqual(status, AuthenticationStatus.ACCOUNT_LOCKED)
        self.assertIsNone(result)
        
        # Verificar se os métodos esperados foram chamados
        self.user_repository.find_by_username.assert_called_once_with("testuser")
        self.test_user.verify_password.assert_not_called()
        self.session_repository.create.assert_not_called()
        
    def test_authenticate_mfa_required(self):
        """Testa autenticação quando MFA é requerido."""
        # Configurar mock para verificação de senha bem-sucedida
        self.test_user.verify_password.return_value = True
        
        # Adicionar fator MFA ao usuário
        self.test_user.auth_factors = [
            {
                "type": "totp",
                "identifier": "authenticator",
                "enabled": True,
                "created_at": datetime.now().isoformat()
            }
        ]
        
        # Chamar o método de autenticação
        status, result = self.auth_service.authenticate(
            "testuser", 
            "correctpassword", 
            "127.0.0.1", 
            "Mozilla/5.0"
        )
        
        # Verificar resultado
        self.assertEqual(status, AuthenticationStatus.MFA_REQUIRED)
        self.assertIsNotNone(result)
        self.assertEqual(result["user_id"], "user-1234")
        self.assertIn("mfa_methods", result)
        self.assertEqual(result["mfa_methods"], ["totp"])
        
        # Verificar se os métodos esperados foram chamados
        self.user_repository.find_by_username.assert_called_once_with("testuser")
        self.test_user.verify_password.assert_called_once_with("correctpassword")
        self.test_user.record_login.assert_called_once()
        self.session_repository.create.assert_called_once()
        
    def test_verify_token_valid(self):
        """Testa verificação de token válido."""
        # Criar uma sessão de teste
        test_session = Session(user_id="user-1234", id="session-1234")
        test_session.access_token = "hash-of-valid-token"
        test_session.status = SessionStatus.ACTIVE.value
        test_session.is_valid = Mock(return_value=True)
        test_session.update_activity = Mock()
        
        # Configurar o repositório de sessões
        self.session_repository.find_by_id.return_value = test_session
        
        # Mockar a função validate_token
        with patch('backend.iam.domain.services.authentication_service.validate_token') as mock_validate:
            # Configurar retorno do validate_token
            mock_validate.return_value = {
                "sub": "user-1234",
                "sid": "session-1234",
                "exp": (datetime.now() + timedelta(hours=1)).timestamp()
            }
            
            # Mockar a função hashlib.sha256
            with patch('backend.iam.domain.services.authentication_service.hashlib.sha256') as mock_hash:
                # Configurar o hash para retornar um objeto com hexdigest
                mock_hash_obj = Mock()
                mock_hash_obj.hexdigest.return_value = "hash-of-valid-token"
                mock_hash.return_value = mock_hash_obj
                
                # Chamar o método de verificação de token
                valid, result = self.auth_service.verify_token("valid-token")
                
                # Verificar resultado
                self.assertTrue(valid)
                self.assertIsNotNone(result)
                self.assertEqual(result["user_id"], "user-1234")
                self.assertEqual(result["session_id"], "session-1234")
                
                # Verificar se os métodos esperados foram chamados
                mock_validate.assert_called_once()
                self.session_repository.find_by_id.assert_called_once_with("session-1234")
                test_session.update_activity.assert_called_once()
                self.user_repository.find_by_id.assert_called_once_with("user-1234")
    
    def test_verify_token_invalid(self):
        """Testa verificação de token inválido."""
        # Mockar a função validate_token para lançar exceção
        with patch('backend.iam.domain.services.authentication_service.validate_token') as mock_validate:
            # Configurar validate_token para lançar exceção
            mock_validate.side_effect = Exception("Token inválido")
            
            # Chamar o método de verificação de token
            valid, result = self.auth_service.verify_token("invalid-token")
            
            # Verificar resultado
            self.assertFalse(valid)
            self.assertIsNone(result)
            
            # Verificar se os métodos esperados foram chamados
            mock_validate.assert_called_once()
            self.session_repository.find_by_id.assert_not_called()
    
    def test_logout(self):
        """Testa logout (revogação de sessão)."""
        # Criar uma sessão de teste
        test_session = Session(user_id="user-1234", id="session-1234")
        test_session.revoke = Mock()
        
        # Configurar o repositório de sessões
        self.session_repository.find_by_id.return_value = test_session
        
        # Mockar a função validate_token
        with patch('backend.iam.domain.services.authentication_service.validate_token') as mock_validate:
            # Configurar retorno do validate_token
            mock_validate.return_value = {
                "sub": "user-1234",
                "sid": "session-1234"
            }
            
            # Chamar o método de logout
            success = self.auth_service.logout("valid-token")
            
            # Verificar resultado
            self.assertTrue(success)
            
            # Verificar se os métodos esperados foram chamados
            mock_validate.assert_called_once()
            self.session_repository.find_by_id.assert_called_once_with("session-1234")
            test_session.revoke.assert_called_once()
            self.session_repository.update.assert_called_once_with(test_session)


if __name__ == '__main__':
    unittest.main()
