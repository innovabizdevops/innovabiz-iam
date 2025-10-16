"""
Adaptador para Serviço de Identificação Civil de Angola

Este módulo implementa o adaptador para integração com o Serviço de Identificação
Civil de Angola (SENICA), permitindo validação de documentos de identidade,
verificação biométrica e consultas sobre o status de documentos emitidos.

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

logger = logging.getLogger("civil_id_adapter")

class CivilIdAdapter(AngolaDataAdapter):
    """
    Adaptador para integração com o Serviço de Identificação Civil de Angola,
    permitindo validações de identidade, verificação de documentos e consultas biométricas.
    """
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        api_version: str = "v2"
    ):
        """
        Inicializa o adaptador para o Serviço de Identificação Civil.
        
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
            "https://api-sandbox.senica.gov.ao"
        )
        self.api_timeout = self.config.get("api_timeout", 30)  # timeout em segundos
        
        # Credenciais
        self.api_key = self.credentials.get("civil_id_api_key", "")
        self.api_secret = self.credentials.get("civil_id_api_secret", "")
        self.institution_id = self.credentials.get("institution_id", "")
        self.certificate_path = self.credentials.get("certificate_path", "")
        
        # Endpoints
        self.endpoints = {
            "auth": f"/api/{self.api_version}/auth/token",
            "validate_bi": f"/api/{self.api_version}/id/validate",
            "verify_photo": f"/api/{self.api_version}/biometrics/photo",
            "verify_fingerprint": f"/api/{self.api_version}/biometrics/fingerprint",
            "document_status": f"/api/{self.api_version}/document/status",
            "address_history": f"/api/{self.api_version}/person/address/history",
            "validate_name": f"/api/{self.api_version}/person/name/validate",
            "check_deceased": f"/api/{self.api_version}/person/status",
            "face_comparison": f"/api/{self.api_version}/biometrics/face/compare"
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
            "max_requests_per_minute": self.config.get("max_requests_per_minute", 30),
            "last_request_time": None,
            "request_count": 0
        }
        
        # Criptografia e segurança
        self.encryption_enabled = self.config.get("encryption_enabled", True)
        self.hash_algorithm = self.config.get("hash_algorithm", "sha256")
    
    def connect(self) -> bool:
        """
        Estabelece conexão com a API do Serviço de Identificação Civil e obtém token de autenticação.
        
        Returns:
            True se conectado com sucesso, False caso contrário
        """
        # Verificar se já temos um token válido
        if self._is_token_valid():
            return True
            
        try:
            # Verificar se as credenciais foram fornecidas
            if not self.api_key or not self.api_secret:
                error_msg = "API Key e Secret não fornecidos para Serviço de Identificação Civil"
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
                
                logger.info("Conexão estabelecida com sucesso com Serviço de Identificação Civil")
                return True
            else:
                error_msg = f"Falha na autenticação com Serviço de Identificação Civil: {response.status_code} - {response.text}"
                self.handle_error("AuthenticationError", error_msg)
                return False
                
        except requests.exceptions.RequestException as e:
            error_msg = f"Erro de requisição ao Serviço de Identificação Civil: {str(e)}"
            self.handle_error("RequestError", error_msg)
            return False
        except Exception as e:
            error_msg = f"Erro inesperado ao conectar ao Serviço de Identificação Civil: {str(e)}"
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
        Gera token de autenticação para API do Serviço de Identificação Civil usando HMAC.
        
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
        Busca dados da API do Serviço de Identificação Civil.
        
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
            cache_key = f"civil_id_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}"
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
                    "error": "Não foi possível estabelecer conexão com o Serviço de Identificação Civil",
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
                    self.cache_data(f"civil_id_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}", data)
                
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
                error_msg = f"Erro na resposta do Serviço de Identificação Civil: {response.status_code} - {response.text}"
                self.handle_error("ResponseError", error_msg)
                
                return {
                    "success": False,
                    "error": error_msg,
                    "status_code": response.status_code,
                    "request_id": request_id
                }
                
        except Exception as e:
            error_msg = f"Erro ao buscar dados do Serviço de Identificação Civil: {str(e)}"
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
        
    def validate_bi(self, bi_data: Dict) -> Dict:
        """
        Valida um Bilhete de Identidade (BI) angolano.
        
        Args:
            bi_data: Dados do BI
                - bi_number: Número do BI
                - full_name: Nome completo
                - birth_date: Data de nascimento (formato YYYY-MM-DD)
                - expiry_date: Data de validade (opcional)
                
        Returns:
            Resultado da validação do BI
        """
        # Validar dados obrigatórios
        required_fields = ["bi_number", "full_name"]
        for field in required_fields:
            if field not in bi_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = bi_data.copy()
        payload["request_type"] = "validate_bi"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar validação do BI
        return self.fetch_data("validate_bi", payload, use_cache=True, cache_max_age=24)  # Cache de 24 horas
    
    def verify_photo(self, photo_data: Dict) -> Dict:
        """
        Verifica a foto de um documento contra o registro oficial.
        
        Args:
            photo_data: Dados da foto
                - bi_number: Número do BI
                - photo: Imagem em base64
                - photo_format: Formato da imagem (jpeg, png)
                
        Returns:
            Resultado da verificação da foto
        """
        # Validar dados obrigatórios
        required_fields = ["bi_number", "photo", "photo_format"]
        for field in required_fields:
            if field not in photo_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = photo_data.copy()
        payload["request_type"] = "verify_photo"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar verificação da foto
        return self.fetch_data("verify_photo", payload, use_cache=False)  # Não usar cache para verificações biométricas
    
    def verify_fingerprint(self, fingerprint_data: Dict) -> Dict:
        """
        Verifica uma impressão digital contra o registro oficial.
        
        Args:
            fingerprint_data: Dados da impressão digital
                - bi_number: Número do BI
                - fingerprint: Impressão digital em base64
                - finger_position: Posição do dedo (thumb, index, middle, ring, little)
                - hand: Mão (left, right)
                - template_format: Formato do template (ANSI, ISO, WSQ)
                
        Returns:
            Resultado da verificação da impressão digital
        """
        # Validar dados obrigatórios
        required_fields = ["bi_number", "fingerprint", "finger_position", "hand"]
        for field in required_fields:
            if field not in fingerprint_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Validar posição do dedo
        valid_positions = ["thumb", "index", "middle", "ring", "little"]
        if fingerprint_data["finger_position"] not in valid_positions:
            return {
                "success": False,
                "error": f"Posição do dedo inválida: {fingerprint_data['finger_position']}",
                "valid_positions": valid_positions
            }
            
        # Validar mão
        valid_hands = ["left", "right"]
        if fingerprint_data["hand"] not in valid_hands:
            return {
                "success": False,
                "error": f"Mão inválida: {fingerprint_data['hand']}",
                "valid_hands": valid_hands
            }
                
        # Adicionar campos adicionais
        payload = fingerprint_data.copy()
        payload["request_type"] = "verify_fingerprint"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar verificação da impressão digital
        return self.fetch_data("verify_fingerprint", payload, use_cache=False)  # Não usar cache para verificações biométricas
    
    def check_document_status(self, document_data: Dict) -> Dict:
        """
        Verifica o status de um documento (válido, expirado, cancelado, etc.).
        
        Args:
            document_data: Dados do documento
                - document_type: Tipo de documento (BI, Passaporte, etc.)
                - document_number: Número do documento
                
        Returns:
            Status atual do documento
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
        payload["request_type"] = "document_status"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar status do documento
        return self.fetch_data("document_status", payload, use_cache=True, cache_max_age=1)  # Cache de 1 hora
    
    def get_address_history(self, person_data: Dict) -> Dict:
        """
        Obtém histórico de endereços de uma pessoa.
        
        Args:
            person_data: Dados da pessoa
                - bi_number: Número do BI
                - start_date: Data inicial (opcional, formato YYYY-MM-DD)
                - end_date: Data final (opcional, formato YYYY-MM-DD)
                
        Returns:
            Histórico de endereços
        """
        # Validar dados obrigatórios
        if "bi_number" not in person_data:
            return {
                "success": False,
                "error": "Campo obrigatório 'bi_number' não fornecido",
                "required_fields": ["bi_number"]
            }
                
        # Adicionar campos adicionais
        payload = person_data.copy()
        payload["request_type"] = "address_history"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar histórico de endereços
        return self.fetch_data("address_history", payload, use_cache=True, cache_max_age=24)  # Cache de 24 horas
    
    def validate_name(self, name_data: Dict) -> Dict:
        """
        Valida se um nome corresponde a um documento específico.
        
        Args:
            name_data: Dados do nome
                - document_type: Tipo de documento (BI, Passaporte, etc.)
                - document_number: Número do documento
                - full_name: Nome completo para verificação
                
        Returns:
            Resultado da validação do nome
        """
        # Validar dados obrigatórios
        required_fields = ["document_type", "document_number", "full_name"]
        for field in required_fields:
            if field not in name_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = name_data.copy()
        payload["request_type"] = "validate_name"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar validação do nome
        return self.fetch_data("validate_name", payload, use_cache=True, cache_max_age=24)  # Cache de 24 horas
    
    def check_deceased_status(self, person_data: Dict) -> Dict:
        """
        Verifica se uma pessoa consta como falecida nos registros civis.
        
        Args:
            person_data: Dados da pessoa
                - document_type: Tipo de documento (BI, Passaporte, etc.)
                - document_number: Número do documento
                - full_name: Nome completo (opcional)
                - birth_date: Data de nascimento (opcional, formato YYYY-MM-DD)
                
        Returns:
            Status da pessoa (viva ou falecida, com data se aplicável)
        """
        # Validar dados obrigatórios
        required_fields = ["document_type", "document_number"]
        for field in required_fields:
            if field not in person_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = person_data.copy()
        payload["request_type"] = "check_deceased"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Buscar status da pessoa
        return self.fetch_data("check_deceased", payload, use_cache=True, cache_max_age=1)  # Cache de 1 hora
    
    def compare_faces(self, face_data: Dict) -> Dict:
        """
        Compara duas imagens faciais para verificar se são da mesma pessoa.
        
        Args:
            face_data: Dados das faces
                - face_image1: Primeira imagem facial em base64
                - face_image2: Segunda imagem facial em base64
                - image_format: Formato das imagens (jpeg, png)
                - threshold: Limite de confiança para correspondência (0.0-1.0, opcional)
                
        Returns:
            Resultado da comparação facial com pontuação de similaridade
        """
        # Validar dados obrigatórios
        required_fields = ["face_image1", "face_image2", "image_format"]
        for field in required_fields:
            if field not in face_data:
                return {
                    "success": False,
                    "error": f"Campo obrigatório não fornecido: {field}",
                    "required_fields": required_fields
                }
                
        # Adicionar campos adicionais
        payload = face_data.copy()
        payload["request_type"] = "face_comparison"
        payload["institution_id"] = self.institution_id
        payload["request_timestamp"] = datetime.now().isoformat()
        
        # Se não foi fornecido threshold, usar o padrão
        if "threshold" not in payload:
            payload["threshold"] = 0.8  # 80% de confiança
        
        # Buscar comparação facial
        return self.fetch_data("face_comparison", payload, use_cache=False)  # Não usar cache para verificações biométricas