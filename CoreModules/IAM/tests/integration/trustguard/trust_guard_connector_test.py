"""
Testes unitários para o conector do TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import os
import unittest
from datetime import datetime
from typing import Dict, Any, List
from unittest.mock import AsyncMock, MagicMock, patch

import httpx
import pytest
from fastapi import HTTPException

from integration.bureau_credito.trustguard.trust_guard_connector import (
    TrustGuardConfig,
    TrustGuardConnector,
)
from integration.bureau_credito.trustguard.trust_guard_models import (
    AccessDecision,
    AccessDecisionResponse,
    AccessRequest,
    AuthenticationLevel,
    AuthenticationRequest,
    AuthenticationResponse,
    AuthMethod,
    Policy,
    ResourceContext,
    RiskLevel,
    SessionInfo,
    SessionStatus,
    TrustScore,
    UserContext,
    UserRisk,
)


class TestTrustGuardConnector(unittest.TestCase):
    """Testes para o conector do TrustGuard."""

    def setUp(self):
        """Configura o ambiente de teste."""
        self.config = TrustGuardConfig(
            api_url="http://mock-trustguard-api.com/v1",
            api_key="mock-api-key",
            tenant_id="mock-tenant",
            timeout=5,
            cache_ttl=300,
            cache_enabled=True,
        )
        
        # Mock do logger
        self.mock_logger = MagicMock()
        
        # Mock do avaliador de regras
        self.mock_rule_evaluator = MagicMock()
        
        # Mock do enhancer de regras
        self.mock_rule_enhancer = MagicMock()
        
        # Mock do cliente HTTP
        self.mock_client = AsyncMock()
        
        # Patch para o init_client
        self.client_patcher = patch.object(
            TrustGuardConnector, 'init_client', 
            return_value=self.mock_client
        )
        
        self.mock_init_client = self.client_patcher.start()
        
        # Cria o conector
        self.connector = TrustGuardConnector(
            config=self.config,
            rule_evaluator=self.mock_rule_evaluator,
            rule_enhancer=self.mock_rule_enhancer,
            logger=self.mock_logger,
        )
        
        # Atribui o cliente mockado
        self.connector.client = self.mock_client
    
    def tearDown(self):
        """Limpa o ambiente após os testes."""
        self.client_patcher.stop()
    
    @pytest.mark.asyncio
    async def test_init_client(self):
        """Testa a inicialização do cliente HTTP."""
        # Restaura o método original
        self.client_patcher.stop()
        
        # Patch do construtor do httpx.AsyncClient
        with patch('httpx.AsyncClient') as mock_async_client:
            # Configura o mock
            mock_client_instance = AsyncMock()
            mock_async_client.return_value = mock_client_instance
            
            # Chama o método
            client = await self.connector.init_client()
            
            # Verifica se o cliente foi criado com os parâmetros corretos
            mock_async_client.assert_called_once()
            call_kwargs = mock_async_client.call_args.kwargs
            assert call_kwargs['timeout'] == self.config.timeout
            assert call_kwargs['headers']['X-API-Key'] == self.config.api_key
            assert call_kwargs['headers']['X-Tenant-ID'] == self.config.tenant_id
            assert call_kwargs['base_url'] == self.config.api_url
            
            # Verifica se o cliente foi retornado corretamente
            assert client == mock_client_instance
        
        # Restaura o patch
        self.client_patcher = patch.object(
            TrustGuardConnector, 'init_client', 
            return_value=self.mock_client
        )
        self.mock_init_client = self.client_patcher.start()
    
    @pytest.mark.asyncio
    async def test_close(self):
        """Testa o fechamento do cliente HTTP."""
        # Configura o mock
        self.connector.client = AsyncMock()
        
        # Chama o método
        await self.connector.close()
        
        # Verifica se o cliente foi fechado
        self.connector.client.aclose.assert_called_once()
        
        # Verifica se o cliente foi removido
        assert self.connector.client is None
    
    @pytest.mark.asyncio
    async def test_make_request_success(self):
        """Testa uma requisição HTTP bem-sucedida."""
        # Configura o mock
        mock_response = AsyncMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"result": "success"}
        
        self.mock_client.request.return_value = mock_response
        
        # Chama o método
        result = await self.connector._make_request(
            method="GET",
            endpoint="/test",
            json_data={"key": "value"},
            params={"param": "value"},
            headers={"Custom-Header": "value"},
        )
        
        # Verifica se a requisição foi feita com os parâmetros corretos
        self.mock_client.request.assert_called_once_with(
            method="GET",
            url="/test",
            json={"key": "value"},
            params={"param": "value"},
            headers={"Custom-Header": "value"},
        )
        
        # Verifica se o resultado está correto
        assert result == {"result": "success"}
    
    @pytest.mark.asyncio
    async def test_make_request_error(self):
        """Testa uma requisição HTTP com erro."""
        # Configura o mock
        mock_response = AsyncMock()
        mock_response.status_code = 400
        mock_response.json.return_value = {"detail": "Bad request"}
        
        self.mock_client.request.return_value = mock_response
        
        # Chama o método e verifica se a exceção é lançada
        with pytest.raises(HTTPException) as excinfo:
            await self.connector._make_request(
                method="GET",
                endpoint="/test",
            )
        
        # Verifica se a exceção tem o código e a mensagem corretos
        assert excinfo.value.status_code == 400
        assert "Bad request" in excinfo.value.detail
    
    @pytest.mark.asyncio
    async def test_make_request_timeout(self):
        """Testa uma requisição HTTP com timeout."""
        # Configura o mock
        self.mock_client.request.side_effect = httpx.TimeoutException("Timeout")
        
        # Chama o método e verifica se a exceção é lançada
        with pytest.raises(HTTPException) as excinfo:
            await self.connector._make_request(
                method="GET",
                endpoint="/test",
            )
        
        # Verifica se a exceção tem o código e a mensagem corretos
        assert excinfo.value.status_code == 408
        assert "Timeout" in excinfo.value.detail
    
    def test_cache_operations(self):
        """Testa operações de cache."""
        # Teste para _get_from_cache quando não há entrada no cache
        hit, decision = self.connector._get_from_cache("test-id")
        assert hit is False
        assert decision is None
        
        # Criar uma decisão de teste
        test_decision = AccessDecisionResponse(
            request_id="test-id",
            decision=AccessDecision.ALLOW,
            reason="Test",
            risk_level=RiskLevel.LOW,
            auth_level=AuthenticationLevel.LOW,
        )
        
        # Teste para _set_cache
        self.connector._set_cache("test-id", test_decision)
        
        # Teste para _get_from_cache quando há entrada no cache
        hit, decision = self.connector._get_from_cache("test-id")
        assert hit is True
        assert decision == test_decision
    
    @pytest.mark.asyncio
    async def test_evaluate_access(self):
        """Testa a avaliação de acesso."""
        # Criar uma solicitação de teste
        test_request = AccessRequest(
            request_id="test-id",
            user_context=UserContext(
                user_id="test-user",
                auth_methods=[AuthMethod.PASSWORD],
                auth_level=AuthenticationLevel.LOW,
                roles=["user"],
                groups=["default"],
                permissions=["read"],
                attributes={},
            ),
            resource_context=ResourceContext(
                resource_id="test-resource",
                resource_type="api",
                action="read",
            ),
            session_data={},
            environment={},
            transaction_data={},
        )
        
        # Criar uma resposta de teste
        test_response = {
            "request_id": "test-id",
            "decision": "ALLOW",
            "reason": "Test",
            "risk_level": "LOW",
            "auth_level": "LOW",
        }
        
        # Configurar o mock para _make_request
        with patch.object(
            self.connector, '_make_request', 
            return_value=test_response
        ) as mock_make_request:
            # Chamar o método
            result = await self.connector.evaluate_access(test_request)
            
            # Verificar se _make_request foi chamado com os parâmetros corretos
            mock_make_request.assert_called_once_with(
                method="POST",
                endpoint="/access/evaluate",
                json_data=test_request.dict(),
            )
            
            # Verificar o resultado
            assert result.request_id == "test-id"
            assert result.decision == AccessDecision.ALLOW
            assert result.reason == "Test"
            assert result.risk_level == RiskLevel.LOW
            assert result.auth_level == AuthenticationLevel.LOW
    
    @pytest.mark.asyncio
    async def test_create_session(self):
        """Testa a criação de sessão."""
        # Criar uma resposta de autenticação de teste
        test_auth_response = AuthenticationResponse(
            request_id="test-id",
            user_id="test-user",
            success=True,
            auth_level=AuthenticationLevel.LOW,
            auth_methods=[AuthMethod.PASSWORD],
        )
        
        # Criar uma resposta de teste
        test_response = {
            "session_id": "test-session",
            "user_id": "test-user",
            "auth_level": "LOW",
            "status": "ACTIVE",
            "created_at": "2025-08-21T12:00:00Z",
            "expires_at": "2025-08-21T13:00:00Z",
        }
        
        # Configurar o mock para _make_request
        with patch.object(
            self.connector, '_make_request', 
            return_value=test_response
        ) as mock_make_request:
            # Chamar o método
            result = await self.connector.create_session(test_auth_response, 3600)
            
            # Verificar se _make_request foi chamado com os parâmetros corretos
            mock_make_request.assert_called_once()
            call_args = mock_make_request.call_args[1]
            assert call_args['method'] == "POST"
            assert call_args['endpoint'] == "/sessions"
            assert call_args['json_data']['user_id'] == "test-user"
            assert call_args['json_data']['auth_level'] == AuthenticationLevel.LOW
            assert call_args['json_data']['ttl'] == 3600
            
            # Verificar o resultado
            assert result.session_id == "test-session"
            assert result.user_id == "test-user"
            assert result.auth_level == AuthenticationLevel.LOW
            assert result.status == SessionStatus.ACTIVE
    
    @pytest.mark.asyncio
    async def test_validate_session_success(self):
        """Testa a validação de sessão com sucesso."""
        # Criar uma resposta de teste
        test_response = {
            "session_id": "test-session",
            "user_id": "test-user",
            "auth_level": "LOW",
            "status": "ACTIVE",
            "created_at": "2025-08-21T12:00:00Z",
            "expires_at": "2025-08-21T13:00:00Z",
        }
        
        # Configurar o mock para _make_request
        with patch.object(
            self.connector, '_make_request', 
            return_value=test_response
        ) as mock_make_request:
            # Chamar o método
            result = await self.connector.validate_session("test-session")
            
            # Verificar se _make_request foi chamado com os parâmetros corretos
            mock_make_request.assert_called_once_with(
                method="GET",
                endpoint="/sessions/test-session",
            )
            
            # Verificar o resultado
            assert result.session_id == "test-session"
            assert result.user_id == "test-user"
            assert result.auth_level == AuthenticationLevel.LOW
            assert result.status == SessionStatus.ACTIVE
    
    @pytest.mark.asyncio
    async def test_validate_session_inactive(self):
        """Testa a validação de sessão inativa."""
        # Criar uma resposta de teste
        test_response = {
            "session_id": "test-session",
            "user_id": "test-user",
            "auth_level": "LOW",
            "status": "EXPIRED",
            "created_at": "2025-08-21T12:00:00Z",
            "expires_at": "2025-08-21T13:00:00Z",
        }
        
        # Configurar o mock para _make_request
        with patch.object(
            self.connector, '_make_request', 
            return_value=test_response
        ) as mock_make_request:
            # Chamar o método e verificar se a exceção é lançada
            with pytest.raises(HTTPException) as excinfo:
                await self.connector.validate_session("test-session")
            
            # Verificar se _make_request foi chamado com os parâmetros corretos
            mock_make_request.assert_called_once_with(
                method="GET",
                endpoint="/sessions/test-session",
            )
            
            # Verificar a exceção
            assert excinfo.value.status_code == 401
            assert "not active" in excinfo.value.detail


if __name__ == "__main__":
    unittest.main()