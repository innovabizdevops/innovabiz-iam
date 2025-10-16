"""
Adaptador para Bureau de Crédito de Angola

Este módulo implementa o adaptador para integração com o Bureau de Crédito angolano,
permitindo consultas de histórico creditício, validações de identidade financeira e
verificação de risco para detecção avançada de fraudes.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import requests
import base64
import time
import uuid
import hashlib
import hmac
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Optional, Union, Any
from .angola_data_adapters import AngolaDataAdapter

logger = logging.getLogger("credit_bureau_adapter")

class CreditBureauAdapter(AngolaDataAdapter):
    """
    Adaptador para integração com o Bureau de Crédito de Angola,
    permitindo consultas de histórico creditício e avaliações de risco.
    """
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        api_version: str = "v2"
    ):
        """
        Inicializa o adaptador para Bureau de Crédito.
        
        Args:
            config_path: Caminho para arquivo de configuração
            credentials_path: Caminho para arquivo de credenciais
            cache_dir: Diretório para cache de dados
            api_version: Versão da API a ser utilizada
        """
        super().__init__(config_path, credentials_path, cache_dir)
        
        # Configurações da API
        self.api_version = api_version
        self.api_base_url = self.config.get(
            "api_base_url", 
            "https://api-sandbox.bureau-credito.co.ao"
        )
        self.api_timeout = self.config.get("api_timeout", 30)  # timeout em segundos
        
        # Credenciais
        self.api_key = self.credentials.get("bureau_api_key", "")
        self.api_secret = self.credentials.get("bureau_api_secret", "")
        self.institution_id = self.credentials.get("institution_id", "")
        self.certificate_path = self.credentials.get("certificate_path", "")
        
        # Endpoints
        self.endpoints = {
            "auth": f"/api/{self.api_version}/auth/token",
            "credit_report": f"/api/{self.api_version}/credit/report",
            "credit_score": f"/api/{self.api_version}/credit/score",
            "identity_check": f"/api/{self.api_version}/identity/verify",
            "fraud_alerts": f"/api/{self.api_version}/alerts/fraud",
            "payment_history": f"/api/{self.api_version}/history/payments",
            "liability_check": f"/api/{self.api_version}/liability/check",
            "document_validation": f"/api/{self.api_version}/document/validate"
        }
        
        # Token de autenticação
        self.auth_token = {
            "token": "",
            "expires_at": None
        }
        
        # Headers padrão
        self.default_headers = {
            "Content-Type": "application/json",
            "User-Agent": "INNOVABIZ-IAM-TrustGuard/1.0",
            "Accept": "application/json"
        }
        
        # Rate limiting
        self.rate_limit = {
            "max_requests_per_minute": self.config.get("max_requests_per_minute", 60),
            "last_request_time": None,
            "request_count": 0
        }
        
        # Criptografia e segurança
        self.encryption_enabled = self.config.get("encryption_enabled", True)
        self.hash_algorithm = self.config.get("hash_algorithm", "sha256")
    
    def connect(self) -> bool:
        """
        Estabelece conexão com a API do Bureau de Crédito e obtém token de autenticação.
        
        Returns:
            True se conectado com sucesso, False caso contrário
        """
        # Verificar se já temos um token válido
        if self._is_token_valid():
            return True
            
        try:
            # Verificar se as credenciais foram fornecidas
            if not self.api_key or not self.api_secret:
                error_msg = "API Key e Secret não fornecidos para Bureau de Crédito"
                self.handle_error("CredentialsError", error_msg)
                return False
            
            # Gerar token de autenticação
            auth_token = self._generate_auth_token()
            
            # Preparar payload para autenticação
            payload = {
                "institution_id": self.institution_id,
                "api_key": self.api_key,
                "timestamp": str(int(time.time()))
            }
            
            # Fazer requisição de autenticação
            auth_url = f"{self.api_base_url}{self.endpoints['auth']}"
            headers = self.default_headers.copy()
            headers["Authorization"] = f"Bearer {auth_token}"
            
            response = self.session.post(
                auth_url,
                headers=headers,
                json=payload,
                timeout=self.api_timeout
            )
            
            if response.status_code == 200:
                token_data = response.json()
                
                # Armazenar token e calcular expiração
                self.auth_token["token"] = token_data["access_token"]
                
                # Calcular tempo de expiração
                expires_in = token_data.get("expires_in", 3600)  # Default: 1 hora
                self.auth_token["expires_at"] = datetime.now() + timedelta(seconds=expires_in)
                
                # Atualizar status de conexão
                self.connection_status["is_connected"] = True
                self.connection_status["last_connection_time"] = datetime.now().isoformat()
                
                logger.info("Conexão estabelecida com sucesso com Bureau de Crédito")
                return True
            else:
                error_msg = f"Falha na autenticação com Bureau de Crédito: {response.status_code} - {response.text}"
                self.handle_error("AuthenticationError", error_msg)
                return False
                
        except requests.exceptions.RequestException as e:
            error_msg = f"Erro de requisição ao Bureau de Crédito: {str(e)}"
            self.handle_error("RequestError", error_msg)
            return False
        except Exception as e:
            error_msg = f"Erro inesperado ao conectar ao Bureau de Crédito: {str(e)}"
            self.handle_error("UnexpectedError", error_msg)
            return False
    
    def _is_token_valid(self) -> bool:
        """
        Verifica se o token de autenticação atual é válido.
        
        Returns:
            True se o token for válido, False caso contrário
        """
        if not self.auth_token["token"] or not self.auth_token["expires_at"]:
            return False
            
        # Verificar se o token ainda é válido (com margem de segurança de 5 minutos)
        safety_margin = timedelta(minutes=5)
        return datetime.now() < (self.auth_token["expires_at"] - safety_margin)
    
    def _generate_auth_token(self) -> str:
        """
        Gera token de autenticação para API do Bureau de Crédito usando HMAC.
        
        Returns:
            Token de autenticação
        """
        timestamp = str(int(time.time()))
        nonce = str(uuid.uuid4())
        message = f"{self.api_key}:{timestamp}:{nonce}"
        
        signature = hmac.new(
            self.api_secret.encode('utf-8'),
            message.encode('utf-8'),
            getattr(hashlib, self.hash_algorithm)
        ).hexdigest()
        
        token_data = {
            "apiKey": self.api_key,
            "timestamp": timestamp,
            "nonce": nonce,
            "signature": signature
        }
        
        return base64.b64encode(json.dumps(token_data).encode('utf-8')).decode('utf-8')
    
    def _check_rate_limit(self) -> bool:
        """
        Verifica e gerencia rate limiting para não exceder limites da API.
        
        Returns:
            True se a requisição pode prosseguir, False caso contrário
        """
        current_time = time.time()
        max_requests = self.rate_limit["max_requests_per_minute"]
        
        # Primeiro request ou minuto diferente
        if not self.rate_limit["last_request_time"] or \
           (current_time - self.rate_limit["last_request_time"]) >= 60:
            self.rate_limit["request_count"] = 1
            self.rate_limit["last_request_time"] = current_time
            return True
            
        # Incrementar contador
        self.rate_limit["request_count"] += 1
        
        # Verificar se excedemos o limite
        if self.rate_limit["request_count"] > max_requests:
            time_to_wait = 60 - (current_time - self.rate_limit["last_request_time"])
            if time_to_wait > 0:
                logger.warning(f"Limite de requisições excedido. Aguardando {time_to_wait:.2f} segundos.")
                time.sleep(time_to_wait)
                # Resetar contador
                self.rate_limit["request_count"] = 1
                self.rate_limit["last_request_time"] = time.time()
                
        return True
        
    def fetch_data(self, endpoint_name: str, payload: Dict, use_cache: bool = True, cache_max_age: int = 24) -> Dict:
        """
        Busca dados da API do Bureau de Crédito.
        
        Args:
            endpoint_name: Nome do endpoint a ser chamado
            payload: Dados a serem enviados na requisição
            use_cache: Se deve usar cache para esta requisição
            cache_max_age: Idade máxima do cache em horas
            
        Returns:
            Dados obtidos ou informações de erro
        """
        # Verificar endpoint
        if endpoint_name not in self.endpoints:
            return {
                "success": False,
                "error": "Endpoint não especificado ou inválido",
                "valid_endpoints": list(self.endpoints.keys())
            }
        
        # Verificar cache
        if use_cache:
            cache_key = f"bureau_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}"
            cached_data = self.get_cached_data(cache_key, cache_max_age)
            
            if cached_data:
                logger.info(f"Dados recuperados do cache para {endpoint_name}")
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
                    "error": "Não foi possível estabelecer conexão com o Bureau de Crédito",
                    "connection_status": self.connection_status
                }
        
        # Verificar rate limiting
        self._check_rate_limit()
        
        try:
            # Preparar URL e headers
            endpoint_url = f"{self.api_base_url}{self.endpoints[endpoint_name]}"
            headers = self.default_headers.copy()
            headers["Authorization"] = f"Bearer {self.auth_token['token']}"
            
            # Adicionar identificação de requisição
            request_id = str(uuid.uuid4())
            headers["X-Request-ID"] = request_id
            
            # Adicionar headers de segurança
            if self.encryption_enabled:
                timestamp = str(int(time.time()))
                headers["X-Timestamp"] = timestamp
                headers["X-Signature"] = self._generate_payload_signature(payload, timestamp)
            
            # Fazer a requisição
            response = self.session.post(
                endpoint_url,
                headers=headers,
                json=payload,
                timeout=self.api_timeout
            )
            
            # Verificar resposta
            if response.status_code == 200:
                data = response.json()
                
                # Salvar em cache se solicitado
                if use_cache:
                    self.cache_data(f"bureau_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}", data)
                
                return {
                    "success": True,
                    "data": data,
                    "source": "api",
                    "request_id": request_id,
                    "timestamp": datetime.now().isoformat()
                }
            elif response.status_code == 401:
                # Token expirou, tentar reconectar
                logger.warning("Token expirado, reconectando...")
                self.auth_token = {"token": "", "expires_at": None}
                
                if self.connect():
                    # Tentar novamente com o novo token
                    return self.fetch_data(endpoint_name, payload, use_cache, cache_max_age)
                else:
                    return {
                        "success": False,
                        "error": "Falha na renovação do token de autenticação",
                        "status_code": 401
                    }
            else:
                error_msg = f"Erro na resposta do Bureau de Crédito: {response.status_code} - {response.text}"
                self.handle_error("ResponseError", error_msg)
                
                return {
                    "success": False,
                    "error": error_msg,
                    "status_code": response.status_code,
                    "request_id": request_id
                }
                
        except Exception as e:
            error_msg = f"Erro ao buscar dados do Bureau de Crédito: {str(e)}"
            self.handle_error("FetchError", error_msg)
            
            return {
                "success": False,
                "error": error_msg
            }
    
    def _generate_payload_signature(self, payload: Dict, timestamp: str) -> str:
        """
        Gera assinatura para o payload da requisição.
        
        Args:
            payload: Dados a serem enviados
            timestamp: Timestamp da requisição
            
        Returns:
            Assinatura do payload
        """
        # Converter payload para string ordenada
        payload_str = json.dumps(payload, sort_keys=True)
        message = f"{payload_str}:{timestamp}:{self.api_key}"
        
        # Gerar HMAC com a chave secreta
        signature = hmac.new(
            self.api_secret.encode('utf-8'),
            message.encode('utf-8'),
            getattr(hashlib, self.hash_algorithm)
        ).hexdigest()
        
        return signature
    
    def get_credit_score(self, subject_data: Dict) -> Dict:
        """
        Obtém o score de crédito de um indivíduo ou empresa.
        
        Args:
            subject_data: Dados do sujeito
                - subject_type: Tipo (individual ou organization)
                - document_type: Tipo de documento (BI, NIF, etc.)
                - document_number: Número do documento
                - full_name: Nome completo (obrigatório para indivíduos)
                
        Returns:
            Score de crédito e detalhes associados
        """
        # Validar dados obrigatórios
        required_fields = ["subject_type", "document_type", "document_number"]
        for field in required_fields:
            if field not in subject_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
        
        # Validações específicas por tipo de sujeito
        subject_type = subject_data["subject_type"]
        if subject_type == "individual" and "full_name" not in subject_data:
            return {
                "success": False,
                "error": "Campo obrigatório 'full_name' não fornecido para indivíduos",
                "required_fields": required_fields + ["full_name"]
            }
            
        # Adicionar campos adicionais
        payload = subject_data.copy()
        payload["request_type"] = "credit_score"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar score de crédito
        return self.fetch_data("credit_score", payload, use_cache=True, cache_max_age=1)  # Cache de 1 hora
    
    def get_credit_report(self, subject_data: Dict, report_type: str = "standard") -> Dict:
        """
        Obtém relatório de crédito completo de um indivíduo ou empresa.
        
        Args:
            subject_data: Dados do sujeito
                - subject_type: Tipo (individual ou organization)
                - document_type: Tipo de documento (BI, NIF, etc.)
                - document_number: Número do documento
                - full_name: Nome completo (obrigatório para indivíduos)
            report_type: Tipo de relatório (standard, detailed, summary)
                
        Returns:
            Relatório de crédito completo
        """
        # Validar dados obrigatórios
        required_fields = ["subject_type", "document_type", "document_number"]
        for field in required_fields:
            if field not in subject_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
        
        # Validar tipo de relatório
        valid_report_types = ["standard", "detailed", "summary"]
        if report_type not in valid_report_types:
            return {
                "success": False,
                "error": f"Tipo de relatório inválido: {report_type}",
                "valid_types": valid_report_types
            }
            
        # Adicionar campos adicionais
        payload = subject_data.copy()
        payload["report_type"] = report_type
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar relatório de crédito
        return self.fetch_data("credit_report", payload, use_cache=True, cache_max_age=1)  # Cache de 1 hora
    
    def verify_identity(self, identity_data: Dict) -> Dict:
        """
        Verifica a identidade de um indivíduo contra registros do Bureau de Crédito.
        
        Args:
            identity_data: Dados de identidade
                - document_type: Tipo de documento (BI, Passaporte, etc.)
                - document_number: Número do documento
                - full_name: Nome completo
                - birth_date: Data de nascimento (formato YYYY-MM-DD)
                - mother_name: Nome da mãe (opcional, mas recomendado)
                
        Returns:
            Resultado da verificação de identidade
        """
        # Validar dados obrigatórios
        required_fields = ["document_type", "document_number", "full_name"]
        for field in required_fields:
            if field not in identity_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = identity_data.copy()
        payload["request_type"] = "identity_verification"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar verificação de identidade
        return self.fetch_data("identity_check", payload, use_cache=False)  # Não usar cache para verificações de identidade
    
    def check_fraud_alerts(self, subject_data: Dict) -> Dict:
        """
        Verifica se existem alertas de fraude associados a um indivíduo ou empresa.
        
        Args:
            subject_data: Dados do sujeito
                - subject_type: Tipo (individual ou organization)
                - document_type: Tipo de documento (BI, NIF, etc.)
                - document_number: Número do documento
                
        Returns:
            Lista de alertas de fraude ativos
        """
        # Validar dados obrigatórios
        required_fields = ["subject_type", "document_type", "document_number"]
        for field in required_fields:
            if field not in subject_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = subject_data.copy()
        payload["request_type"] = "fraud_alerts"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar alertas de fraude
        return self.fetch_data("fraud_alerts", payload, use_cache=True, cache_max_age=1)  # Cache de 1 hora
    
    def get_payment_history(self, subject_data: Dict, history_params: Dict = None) -> Dict:
        """
        Obtém histórico de pagamentos de um indivíduo ou empresa.
        
        Args:
            subject_data: Dados do sujeito
                - subject_type: Tipo (individual ou organization)
                - document_type: Tipo de documento (BI, NIF, etc.)
                - document_number: Número do documento
            history_params: Parâmetros adicionais
                - start_date: Data inicial (formato YYYY-MM-DD)
                - end_date: Data final (formato YYYY-MM-DD)
                - payment_types: Lista de tipos de pagamento
                
        Returns:
            Histórico de pagamentos
        """
        # Validar dados obrigatórios
        required_fields = ["subject_type", "document_type", "document_number"]
        for field in required_fields:
            if field not in subject_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Parâmetros padrão
        params = history_params or {}
        if "start_date" not in params:
            params["start_date"] = (datetime.now() - timedelta(days=365)).strftime("%Y-%m-%d")
        if "end_date" not in params:
            params["end_date"] = datetime.now().strftime("%Y-%m-%d")
        if "payment_types" not in params:
            params["payment_types"] = ["loan", "credit_card", "utility", "rental", "mortgage"]
            
        # Adicionar campos adicionais
        payload = subject_data.copy()
        payload.update(params)
        payload["request_type"] = "payment_history"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar histórico de pagamentos
        return self.fetch_data("payment_history", payload, use_cache=True, cache_max_age=6)  # Cache de 6 horas
    
    def check_liabilities(self, subject_data: Dict) -> Dict:
        """
        Verifica responsabilidades financeiras de um indivíduo ou empresa.
        
        Args:
            subject_data: Dados do sujeito
                - subject_type: Tipo (individual ou organization)
                - document_type: Tipo de documento (BI, NIF, etc.)
                - document_number: Número do documento
                
        Returns:
            Lista de responsabilidades financeiras ativas
        """
        # Validar dados obrigatórios
        required_fields = ["subject_type", "document_type", "document_number"]
        for field in required_fields:
            if field not in subject_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = subject_data.copy()
        payload["request_type"] = "liability_check"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar responsabilidades financeiras
        return self.fetch_data("liability_check", payload, use_cache=True, cache_max_age=24)  # Cache de 24 horas
    
    def validate_document(self, document_data: Dict) -> Dict:
        """
        Valida a autenticidade de um documento.
        
        Args:
            document_data: Dados do documento
                - document_type: Tipo de documento (BI, NIF, Passaporte, etc.)
                - document_number: Número do documento
                - issue_date: Data de emissão (formato YYYY-MM-DD)
                - expiry_date: Data de validade (formato YYYY-MM-DD)
                - issuing_authority: Autoridade emissora
                
        Returns:
            Resultado da validação do documento
        """
        # Validar dados obrigatórios
        required_fields = ["document_type", "document_number"]
        for field in required_fields:
            if field not in document_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = document_data.copy()
        payload["request_type"] = "document_validation"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar validação de documento
        return self.fetch_data("document_validation", payload, use_cache=False)  # Não usar cache para validações de documento