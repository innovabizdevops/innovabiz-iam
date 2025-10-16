"""
Adaptador para Operadoras de Mobile Money em Angola

Este módulo implementa adaptadores para integração com as principais operadoras
de mobile money em Angola (Unitel Money, Movicel MKesh, BAI Multicaixa), permitindo
verificação de transações, validação de contas e detecção de padrões suspeitos.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import requests
import hashlib
import hmac
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Optional, Union, Any
from .angola_data_adapters import AngolaDataAdapter

logger = logging.getLogger("mobile_money_adapter")

class MobileMoneyAdapter(AngolaDataAdapter):
    """
    Adaptador base para operadoras de mobile money em Angola,
    provendo funcionalidades comuns para todos os provedores.
    """
    
    def __init__(
        self,
        provider: str,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        api_version: str = "v1"
    ):
        """
        Inicializa o adaptador para mobile money.
        
        Args:
            provider: Operadora de mobile money ("UNITEL", "MOVICEL", "BAI")
            config_path: Caminho para arquivo de configuração
            credentials_path: Caminho para arquivo de credenciais
            cache_dir: Diretório para cache de dados
            api_version: Versão da API a ser utilizada
        """
        super().__init__(config_path, credentials_path, cache_dir)
        
        # Verificar provedor válido
        valid_providers = ["UNITEL", "MOVICEL", "BAI"]
        if provider.upper() not in valid_providers:
            raise ValueError(f"Provedor inválido. Use um dos seguintes: {', '.join(valid_providers)}")
        
        self.provider = provider.upper()
        self.api_version = api_version
        
        # Configurações específicas por provedor
        provider_configs = {
            "UNITEL": {
                "api_base_url": "https://api-sandbox.unitel.money",
                "auth_endpoint": "/auth/token",
                "transaction_endpoint": "/transactions",
                "verification_endpoint": "/verification",
                "account_endpoint": "/accounts",
                "limits_endpoint": "/limits",
                "status_endpoint": "/status"
            },
            "MOVICEL": {
                "api_base_url": "https://api-test.mkesh.co.ao",
                "auth_endpoint": "/oauth/token",
                "transaction_endpoint": "/api/transactions",
                "verification_endpoint": "/api/verification",
                "account_endpoint": "/api/accounts",
                "limits_endpoint": "/api/limits",
                "status_endpoint": "/api/status"
            },
            "BAI": {
                "api_base_url": "https://api-sandbox.multicaixa.co.ao",
                "auth_endpoint": "/oauth2/token",
                "transaction_endpoint": "/api/v1/transactions",
                "verification_endpoint": "/api/v1/verification",
                "account_endpoint": "/api/v1/accounts",
                "limits_endpoint": "/api/v1/limits",
                "status_endpoint": "/api/v1/status"
            }
        }
        
        # Aplicar configurações do provedor
        self.provider_config = provider_configs[self.provider]
        self.api_base_url = self.config.get(
            "api_base_url", 
            self.provider_config["api_base_url"]
        )
        
        # Aplicar endpoints
        self.endpoints = self.provider_config
        
        # Credenciais específicas
        self.client_id = self.credentials.get(f"{self.provider.lower()}_client_id", "")
        self.client_secret = self.credentials.get(f"{self.provider.lower()}_client_secret", "")
        self.api_key = self.credentials.get(f"{self.provider.lower()}_api_key", "")
        
        # Token de autenticação
        self.auth_token = {
            "access_token": "",
            "expires_at": None,
            "token_type": "Bearer"
        }
        
        # Headers padrão
        self.default_headers = {
            "Content-Type": "application/json",
            "User-Agent": f"INNOVABIZ-IAM-TrustGuard/{self.provider}/1.0",
            "Accept": "application/json"
        }
        
        # Rate limiting
        self.rate_limit = {
            "max_requests_per_minute": self.config.get("max_requests_per_minute", 100),
            "last_request_time": None,
            "request_count": 0
        }
    
    def connect(self) -> bool:
        """
        Estabelece conexão com a API da operadora e obtém token de autenticação.
        
        Returns:
            True se conectado com sucesso, False caso contrário
        """
        # Verificar se já temos um token válido
        if self._is_token_valid():
            return True
            
        try:
            # Verificar se as credenciais foram fornecidas
            if not self.client_id or not self.client_secret:
                error_msg = f"Client ID e Secret não fornecidos para {self.provider}"
                self.handle_error("CredentialsError", error_msg)
                return False
                
            # Montar URL e payload para autenticação
            auth_url = f"{self.api_base_url}{self.endpoints['auth_endpoint']}"
            
            # Payload específico por provedor
            if self.provider == "UNITEL":
                payload = {
                    "grant_type": "client_credentials",
                    "client_id": self.client_id,
                    "client_secret": self.client_secret
                }
            elif self.provider == "MOVICEL":
                # Mkesh usa Basic Auth no header
                headers = self.default_headers.copy()
                auth_string = f"{self.client_id}:{self.client_secret}"
                headers["Authorization"] = f"Basic {base64.b64encode(auth_string.encode()).decode()}"
                payload = {"grant_type": "client_credentials"}
            elif self.provider == "BAI":
                payload = {
                    "grant_type": "client_credentials",
                    "client_id": self.client_id,
                    "client_secret": self.client_secret,
                    "scope": "transactions account verification"
                }
            
            # Fazer requisição de autenticação
            response = self.session.post(
                auth_url,
                headers=self.default_headers,
                json=payload,
                timeout=30
            )
            
            if response.status_code == 200:
                token_data = response.json()
                
                # Armazenar token e calcular expiração
                self.auth_token["access_token"] = token_data["access_token"]
                self.auth_token["token_type"] = token_data.get("token_type", "Bearer")
                
                # Calcular tempo de expiração
                expires_in = token_data.get("expires_in", 3600)  # Default: 1 hora
                self.auth_token["expires_at"] = datetime.now() + timedelta(seconds=expires_in)
                
                # Atualizar status de conexão
                self.connection_status["is_connected"] = True
                self.connection_status["last_connection_time"] = datetime.now().isoformat()
                
                logger.info(f"Conexão estabelecida com sucesso com {self.provider} Money")
                return True
            else:
                error_msg = f"Falha na autenticação com {self.provider}: {response.status_code} - {response.text}"
                self.handle_error("AuthenticationError", error_msg)
                return False
                
        except Exception as e:
            error_msg = f"Erro ao conectar com {self.provider}: {str(e)}"
            self.handle_error("ConnectionError", error_msg)
            return False
    
    def _is_token_valid(self) -> bool:
        """
        Verifica se o token de autenticação atual é válido.
        
        Returns:
            True se o token for válido, False caso contrário
        """
        if not self.auth_token["access_token"] or not self.auth_token["expires_at"]:
            return False
            
        # Verificar se o token ainda é válido (com margem de segurança de 5 minutos)
        safety_margin = timedelta(minutes=5)
        return datetime.now() < (self.auth_token["expires_at"] - safety_margin)
    
    def _get_auth_header(self) -> Dict:
        """
        Obtém o header de autenticação atualizado.
        
        Returns:
            Header com token de autenticação
        """
        if not self._is_token_valid():
            self.connect()
            
        return {
            "Authorization": f"{self.auth_token['token_type']} {self.auth_token['access_token']}"
        }
    
    def fetch_data(self, query_params: Dict) -> Dict:
        """
        Busca dados da API da operadora de mobile money.
        
        Args:
            query_params: Dicionário com parâmetros de consulta, incluindo:
                - endpoint_key: Chave do endpoint a ser consultado
                - path: Caminho adicional após o endpoint base
                - method: Método HTTP (GET, POST, etc.)
                - parameters: Parâmetros específicos da consulta
                - use_cache: Se deve tentar usar dados em cache
                
        Returns:
            Dicionário com dados obtidos ou informações de erro
        """
        # Verificar parâmetros obrigatórios
        endpoint_key = query_params.get("endpoint_key")
        if not endpoint_key or endpoint_key not in self.endpoints:
            return {
                "success": False,
                "error": "Endpoint não especificado ou inválido",
                "valid_endpoints": list(self.endpoints.keys())
            }
        
        # Extrair parâmetros
        method = query_params.get("method", "GET")
        path = query_params.get("path", "")
        parameters = query_params.get("parameters", {})
        use_cache = query_params.get("use_cache", True)
        cache_max_age = query_params.get("cache_max_age", 24)  # horas
        
        # Verificar cache para requisições GET
        if method == "GET" and use_cache:
            cache_key = f"{self.provider}_{endpoint_key}_{path}_{hash(json.dumps(parameters, sort_keys=True))}"
            cached_data = self.get_cached_data(cache_key, cache_max_age)
            
            if cached_data:
                logger.info(f"Dados recuperados do cache para {endpoint_key}")
                return {
                    "success": True,
                    "data": cached_data,
                    "source": "cache",
                    "cached_at": datetime.now().isoformat()
                }
        
        # Verificar conexão
        if not self.connection_status["is_connected"]:
            if not self.connect():
                return {
                    "success": False,
                    "error": f"Não foi possível estabelecer conexão com {self.provider}",
                    "connection_status": self.connection_status
                }
        
        try:
            # Construir URL completa
            endpoint_base = self.endpoints[endpoint_key]
            url = f"{self.api_base_url}{endpoint_base}{path}"
            
            # Preparar headers
            headers = self.default_headers.copy()
            headers.update(self._get_auth_header())
            
            # Adicionar API Key se disponível
            if self.api_key:
                headers["X-API-Key"] = self.api_key
            
            # Adicionar ID de rastreamento para auditoria
            trace_id = str(uuid.uuid4())
            headers["X-Trace-ID"] = trace_id
            
            # Adicionar timestamp para assinatura de requisição
            timestamp = str(int(time.time()))
            headers["X-Timestamp"] = timestamp
            
            # Fazer a requisição de acordo com o método
            if method == "GET":
                response = self.session.get(url, headers=headers, params=parameters, timeout=30)
            elif method == "POST":
                response = self.session.post(url, headers=headers, json=parameters, timeout=30)
            else:
                return {
                    "success": False,
                    "error": f"Método HTTP não suportado: {method}"
                }
            
            # Processar resposta
            if response.status_code in [200, 201]:
                data = response.json()
                
                # Armazenar em cache para requisições GET
                if method == "GET" and use_cache:
                    self.cache_data(cache_key, data)
                
                return {
                    "success": True,
                    "data": data,
                    "trace_id": trace_id,
                    "source": "api",
                    "timestamp": datetime.now().isoformat()
                }
            else:
                error_msg = f"Erro na resposta de {self.provider}: {response.status_code} - {response.text}"
                self.handle_error("ResponseError", error_msg)
                
                return {
                    "success": False,
                    "error": error_msg,
                    "status_code": response.status_code,
                    "trace_id": trace_id
                }
                
        except Exception as e:
            error_msg = f"Erro ao buscar dados de {self.provider}: {str(e)}"
            self.handle_error("FetchError", error_msg)
            
            return {
                "success": False,
                "error": error_msg
            }
    
    def verify_account(self, account_data: Dict) -> Dict:
        """
        Verifica se uma conta de mobile money é válida.
        
        Args:
            account_data: Dados da conta a ser verificada
                - phone_number: Número de telefone
                - account_id: ID da conta (opcional, dependendo do provedor)
                
        Returns:
            Resultado da verificação
        """
        if "phone_number" not in account_data:
            return {
                "success": False,
                "error": "Número de telefone não fornecido"
            }
        
        # Normalizar número de telefone (remover prefixo internacional se presente)
        phone = account_data["phone_number"]
        if phone.startswith("+244"):
            phone = phone[4:]
        elif phone.startswith("00244"):
            phone = phone[5:]
            
        # Preparar parâmetros de consulta
        query_params = {
            "endpoint_key": "verification_endpoint",
            "path": "/account",
            "method": "POST",
            "parameters": {
                "phone_number": phone,
                "account_id": account_data.get("account_id", ""),
                "request_id": str(uuid.uuid4())
            },
            "use_cache": False  # Não usar cache para verificações de conta
        }
        
        return self.fetch_data(query_params)
    
    def check_transaction_status(self, transaction_id: str) -> Dict:
        """
        Verifica o status de uma transação.
        
        Args:
            transaction_id: ID da transação
            
        Returns:
            Status da transação
        """
        query_params = {
            "endpoint_key": "transaction_endpoint",
            "path": f"/{transaction_id}",
            "method": "GET",
            "use_cache": False  # Não usar cache para status de transação
        }
        
        return self.fetch_data(query_params)
    
    def get_transaction_history(self, account_data: Dict, history_params: Dict = None) -> Dict:
        """
        Obtém histórico de transações de uma conta.
        
        Args:
            account_data: Dados da conta
                - phone_number: Número de telefone
                - account_id: ID da conta (opcional)
            history_params: Parâmetros adicionais
                - start_date: Data inicial (formato YYYY-MM-DD)
                - end_date: Data final (formato YYYY-MM-DD)
                - limit: Número máximo de transações
                
        Returns:
            Histórico de transações
        """
        if "phone_number" not in account_data:
            return {
                "success": False,
                "error": "Número de telefone não fornecido"
            }
            
        # Normalizar número de telefone
        phone = account_data["phone_number"]
        if phone.startswith("+244"):
            phone = phone[4:]
        elif phone.startswith("00244"):
            phone = phone[5:]
            
        # Parâmetros padrão
        params = history_params or {}
        if "start_date" not in params:
            params["start_date"] = (datetime.now() - timedelta(days=30)).strftime("%Y-%m-%d")
        if "end_date" not in params:
            params["end_date"] = datetime.now().strftime("%Y-%m-%d")
        if "limit" not in params:
            params["limit"] = 50
            
        # Adicionar dados da conta
        params["phone_number"] = phone
        if "account_id" in account_data:
            params["account_id"] = account_data["account_id"]
            
        query_params = {
            "endpoint_key": "transaction_endpoint",
            "path": "/history",
            "method": "GET",
            "parameters": params,
            "use_cache": True,
            "cache_max_age": 1  # Atualização a cada hora
        }
        
        return self.fetch_data(query_params)
    
    def check_account_limits(self, account_data: Dict) -> Dict:
        """
        Verifica limites de transação de uma conta.
        
        Args:
            account_data: Dados da conta
                - phone_number: Número de telefone
                - account_id: ID da conta (opcional)
                
        Returns:
            Limites da conta
        """
        if "phone_number" not in account_data:
            return {
                "success": False,
                "error": "Número de telefone não fornecido"
            }
            
        # Normalizar número
        phone = account_data["phone_number"]
        if phone.startswith("+244"):
            phone = phone[4:]
        elif phone.startswith("00244"):
            phone = phone[5:]
            
        params = {
            "phone_number": phone
        }
        if "account_id" in account_data:
            params["account_id"] = account_data["account_id"]
            
        query_params = {
            "endpoint_key": "limits_endpoint",
            "method": "GET",
            "parameters": params,
            "use_cache": True,
            "cache_max_age": 24  # Atualização diária
        }
        
        return self.fetch_data(query_params)
    
    def get_provider_status(self) -> Dict:
        """
        Verifica o status do serviço da operadora.
        
        Returns:
            Status do serviço
        """
        query_params = {
            "endpoint_key": "status_endpoint",
            "method": "GET",
            "use_cache": False
        }
        
        return self.fetch_data(query_params)