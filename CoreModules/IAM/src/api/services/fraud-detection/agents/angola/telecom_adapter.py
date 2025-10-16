"""
Adaptador para Operadoras de Telecomunicações de Angola

Este módulo implementa o adaptador para integração com as operadoras de telecomunicações
de Angola (Unitel, Movicel, Angola Telecom e TV Cabo), permitindo validações de
números telefônicos, informações de geolocalização e padrões de comunicação para
detecção avançada de fraudes.

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

logger = logging.getLogger("telecom_adapter")

class TelecomAdapter(AngolaDataAdapter):
    """
    Adaptador para integração com operadoras de telecomunicações de Angola,
    incluindo Unitel, Movicel, Angola Telecom e TV Cabo.
    """
    
    # Operadoras suportadas
    SUPPORTED_PROVIDERS = ["unitel", "movicel", "angola_telecom", "tvcabo"]
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        api_version: str = "v2"
    ):
        """
        Inicializa o adaptador para operadoras de telecomunicações.
        
        Args:
            config_path: Caminho para arquivo de configuração
            credentials_path: Caminho para arquivo de credenciais
            cache_dir: Diretório para cache de dados
            api_version: Versão da API a ser utilizada
        """
        super().__init__(config_path, credentials_path, cache_dir)
        
        # Configurações da API
        self.api_version = api_version
        self.api_timeout = self.config.get("api_timeout", 30)  # timeout em segundos
        
        # APIs das operadoras
        self.provider_apis = {
            "unitel": {
                "base_url": self.config.get("unitel_api_base_url", "https://api-partner.unitel.ao"),
                "auth_endpoint": f"/api/{self.api_version}/auth/token",
                "verify_endpoint": f"/api/{self.api_version}/subscriber/verify",
                "location_endpoint": f"/api/{self.api_version}/subscriber/location",
                "usage_endpoint": f"/api/{self.api_version}/subscriber/usage",
                "devices_endpoint": f"/api/{self.api_version}/subscriber/devices"
            },
            "movicel": {
                "base_url": self.config.get("movicel_api_base_url", "https://api-partners.movicel.co.ao"),
                "auth_endpoint": f"/api/{self.api_version}/auth/token",
                "verify_endpoint": f"/api/{self.api_version}/subscriber/verify",
                "location_endpoint": f"/api/{self.api_version}/subscriber/location",
                "usage_endpoint": f"/api/{self.api_version}/subscriber/usage",
                "devices_endpoint": f"/api/{self.api_version}/subscriber/devices"
            },
            "angola_telecom": {
                "base_url": self.config.get("angola_telecom_api_base_url", "https://api-partners.angolatelecom.ao"),
                "auth_endpoint": f"/api/{self.api_version}/auth/token",
                "verify_endpoint": f"/api/{self.api_version}/subscriber/verify",
                "usage_endpoint": f"/api/{self.api_version}/subscriber/usage"
            },
            "tvcabo": {
                "base_url": self.config.get("tvcabo_api_base_url", "https://api-partners.tvcabo.ao"),
                "auth_endpoint": f"/api/{self.api_version}/auth/token",
                "verify_endpoint": f"/api/{self.api_version}/subscriber/verify",
                "usage_endpoint": f"/api/{self.api_version}/subscriber/usage",
                "devices_endpoint": f"/api/{self.api_version}/subscriber/devices"
            }
        }
        
        # Credenciais
        self.credentials_by_provider = {
            "unitel": {
                "api_key": self.credentials.get("unitel_api_key", ""),
                "api_secret": self.credentials.get("unitel_api_secret", ""),
                "client_id": self.credentials.get("unitel_client_id", "")
            },
            "movicel": {
                "api_key": self.credentials.get("movicel_api_key", ""),
                "api_secret": self.credentials.get("movicel_api_secret", ""),
                "client_id": self.credentials.get("movicel_client_id", "")
            },
            "angola_telecom": {
                "api_key": self.credentials.get("angola_telecom_api_key", ""),
                "api_secret": self.credentials.get("angola_telecom_api_secret", ""),
                "client_id": self.credentials.get("angola_telecom_client_id", "")
            },
            "tvcabo": {
                "api_key": self.credentials.get("tvcabo_api_key", ""),
                "api_secret": self.credentials.get("tvcabo_api_secret", ""),
                "client_id": self.credentials.get("tvcabo_client_id", "")
            }
        }
        
        # Token de autenticação para cada operadora
        self.auth_tokens = {provider: {"token": "", "expires_at": None} for provider in self.SUPPORTED_PROVIDERS}
        
        # Headers padrão
        self.default_headers = {
            "Content-Type": "application/json",
            "User-Agent": "INNOVABIZ-IAM-TrustGuard/1.0",
            "Accept": "application/json"
        }
        
        # Rate limiting por operadora
        self.rate_limits = {
            provider: {
                "max_requests_per_minute": self.config.get(f"{provider}_max_requests_per_minute", 30),
                "last_request_time": None,
                "request_count": 0
            } for provider in self.SUPPORTED_PROVIDERS
        }
        
        # Criptografia e segurança
        self.encryption_enabled = self.config.get("encryption_enabled", True)
        self.hash_algorithm = self.config.get("hash_algorithm", "sha256")    
    def connect(self, provider: str) -> bool:
        """
        Estabelece conexão com a API da operadora especificada e obtém token de autenticação.
        
        Args:
            provider: Nome da operadora (unitel, movicel, angola_telecom, tvcabo)
            
        Returns:
            True se conectado com sucesso, False caso contrário
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            error_msg = f"Operadora não suportada: {provider}. Operadoras válidas: {self.SUPPORTED_PROVIDERS}"
            self.handle_error("ProviderError", error_msg)
            return False
        
        # Verificar se já temos um token válido para esta operadora
        if self._is_token_valid(provider):
            return True
            
        try:
            # Obter credenciais e endpoints para esta operadora
            provider_creds = self.credentials_by_provider[provider]
            provider_api = self.provider_apis[provider]
            
            # Verificar se as credenciais foram fornecidas
            if not provider_creds["api_key"] or not provider_creds["api_secret"]:
                error_msg = f"API Key e Secret não fornecidos para a operadora {provider}"
                self.handle_error("CredentialsError", error_msg)
                return False
            
            # Gerar token de autenticação
            auth_token = self._generate_auth_token(provider)
            
            # Preparar payload para autenticação
            payload = {
                "client_id": provider_creds["client_id"],
                "api_key": provider_creds["api_key"],
                "timestamp": str(int(time.time()))
            }
            
            # Fazer requisição de autenticação
            auth_url = f"{provider_api['base_url']}{provider_api['auth_endpoint']}"
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
                self.auth_tokens[provider]["token"] = token_data["access_token"]
                
                # Calcular tempo de expiração
                expires_in = token_data.get("expires_in", 3600)  # Default: 1 hora
                self.auth_tokens[provider]["expires_at"] = datetime.now() + timedelta(seconds=expires_in)
                
                # Atualizar status de conexão
                self.connection_status["is_connected"] = True
                self.connection_status["last_connection_time"] = datetime.now().isoformat()
                
                logger.info(f"Conexão estabelecida com sucesso com a operadora {provider}")
                return True
            else:
                error_msg = f"Falha na autenticação com a operadora {provider}: {response.status_code} - {response.text}"
                self.handle_error("AuthenticationError", error_msg)
                return False
                
        except requests.exceptions.RequestException as e:
            error_msg = f"Erro de requisição à operadora {provider}: {str(e)}"
            self.handle_error("RequestError", error_msg)
            return False
        except Exception as e:
            error_msg = f"Erro inesperado ao conectar à operadora {provider}: {str(e)}"
            self.handle_error("UnexpectedError", error_msg)
            return False    
    def _is_token_valid(self, provider: str) -> bool:
        """
        Verifica se o token de autenticação atual é válido para a operadora especificada.
        
        Args:
            provider: Nome da operadora
            
        Returns:
            True se o token for válido, False caso contrário
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            return False
            
        token_data = self.auth_tokens.get(provider, {})
        
        if not token_data.get("token") or not token_data.get("expires_at"):
            return False
            
        # Verificar se o token ainda é válido (com margem de segurança de 5 minutos)
        safety_margin = timedelta(minutes=5)
        return datetime.now() < (token_data["expires_at"] - safety_margin)
    
    def _generate_auth_token(self, provider: str) -> str:
        """
        Gera token de autenticação para API da operadora especificada usando HMAC.
        
        Args:
            provider: Nome da operadora
            
        Returns:
            Token de autenticação
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            raise ValueError(f"Operadora não suportada: {provider}")
            
        # Obter credenciais para esta operadora
        provider_creds = self.credentials_by_provider[provider]
        
        timestamp = str(int(time.time()))
        nonce = str(uuid.uuid4())
        message = f"{provider_creds['api_key']}:{timestamp}:{nonce}"
        
        signature = hmac.new(
            provider_creds['api_secret'].encode('utf-8'),
            message.encode('utf-8'),
            getattr(hashlib, self.hash_algorithm)
        ).hexdigest()
        
        token_data = {
            "apiKey": provider_creds['api_key'],
            "timestamp": timestamp,
            "nonce": nonce,
            "signature": signature
        }
        
        return base64.b64encode(json.dumps(token_data).encode('utf-8')).decode('utf-8')
    
    def _check_rate_limit(self, provider: str) -> bool:
        """
        Verifica e gerencia rate limiting para não exceder limites da API para a operadora especificada.
        
        Args:
            provider: Nome da operadora
            
        Returns:
            True se a requisição pode prosseguir, False caso contrário
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            return False
            
        rate_limit_data = self.rate_limits[provider]
        current_time = time.time()
        max_requests = rate_limit_data["max_requests_per_minute"]
        
        # Primeiro request ou minuto diferente
        if not rate_limit_data["last_request_time"] or \
           (current_time - rate_limit_data["last_request_time"]) >= 60:
            rate_limit_data["request_count"] = 1
            rate_limit_data["last_request_time"] = current_time
            return True
            
        # Incrementar contador
        rate_limit_data["request_count"] += 1
        
        # Verificar se excedemos o limite
        if rate_limit_data["request_count"] > max_requests:
            time_to_wait = 60 - (current_time - rate_limit_data["last_request_time"])
            if time_to_wait > 0:
                logger.warning(f"Limite de requisições excedido para a operadora {provider}. Aguardando {time_to_wait:.2f} segundos.")
                time.sleep(time_to_wait)
                # Resetar contador
                rate_limit_data["request_count"] = 1
                rate_limit_data["last_request_time"] = time.time()
                
        return True    
    def fetch_data(self, provider: str, endpoint_name: str, payload: Dict, use_cache: bool = True, cache_max_age: int = 24) -> Dict:
        """
        Busca dados da API da operadora especificada.
        
        Args:
            provider: Nome da operadora
            endpoint_name: Nome do endpoint a ser chamado
            payload: Dados a serem enviados na requisição
            use_cache: Se deve usar cache para esta requisição
            cache_max_age: Idade máxima do cache em horas
            
        Returns:
            Dados obtidos ou informações de erro
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            return {
                "success": False,
                "error": f"Operadora não suportada: {provider}",
                "valid_providers": self.SUPPORTED_PROVIDERS
            }
            
        # Obter API da operadora
        provider_api = self.provider_apis.get(provider, {})
        
        # Verificar endpoint
        if endpoint_name not in provider_api:
            return {
                "success": False,
                "error": f"Endpoint não especificado ou inválido para a operadora {provider}",
                "valid_endpoints": list(provider_api.keys())
            }
        
        # Verificar cache
        if use_cache:
            cache_key = f"telecom_{provider}_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}"
            cached_data = self.get_cached_data(cache_key, cache_max_age)
            
            if cached_data:
                logger.info(f"Dados recuperados do cache para {provider} - {endpoint_name}")
                return {
                    "success": True,
                    "data": cached_data,
                    "source": "cache",
                    "cached_at": datetime.now().isoformat()
                }
        
        # Verificar conexão
        if not self._is_token_valid(provider):
            if not self.connect(provider):
                return {
                    "success": False,
                    "error": f"Não foi possível estabelecer conexão com a operadora {provider}",
                    "connection_status": self.connection_status
                }
        
        # Verificar rate limiting
        self._check_rate_limit(provider)
        
        try:
            # Preparar URL e headers
            endpoint_url = f"{provider_api['base_url']}{provider_api[endpoint_name]}"
            headers = self.default_headers.copy()
            headers["Authorization"] = f"Bearer {self.auth_tokens[provider]['token']}"
            
            # Adicionar identificação de requisição
            request_id = str(uuid.uuid4())
            headers["X-Request-ID"] = request_id
            
            # Adicionar headers de segurança
            if self.encryption_enabled:
                timestamp = str(int(time.time()))
                headers["X-Timestamp"] = timestamp
                headers["X-Signature"] = self._generate_payload_signature(provider, payload, timestamp)
            
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
                    self.cache_data(f"telecom_{provider}_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}", data)
                
                return {
                    "success": True,
                    "data": data,
                    "source": "api",
                    "request_id": request_id,
                    "timestamp": datetime.now().isoformat()
                }
            elif response.status_code == 401:
                # Token expirou, tentar reconectar
                logger.warning(f"Token expirado para a operadora {provider}, reconectando...")
                self.auth_tokens[provider] = {"token": "", "expires_at": None}
                
                if self.connect(provider):
                    # Tentar novamente com o novo token
                    return self.fetch_data(provider, endpoint_name, payload, use_cache, cache_max_age)
                else:
                    return {
                        "success": False,
                        "error": "Falha na renovação do token de autenticação",
                        "status_code": 401
                    }
            else:
                error_msg = f"Erro na resposta da operadora {provider}: {response.status_code} - {response.text}"
                self.handle_error("ResponseError", error_msg)
                
                return {
                    "success": False,
                    "error": error_msg,
                    "status_code": response.status_code,
                    "request_id": request_id
                }
                
        except Exception as e:
            error_msg = f"Erro ao buscar dados da operadora {provider}: {str(e)}"
            self.handle_error("FetchError", error_msg)
            
            return {
                "success": False,
                "error": error_msg
            }
    
    def _generate_payload_signature(self, provider: str, payload: Dict, timestamp: str) -> str:
        """
        Gera assinatura para o payload da requisição para a operadora especificada.
        
        Args:
            provider: Nome da operadora
            payload: Dados a serem enviados
            timestamp: Timestamp da requisição
            
        Returns:
            Assinatura do payload
        """
        # Verificar se a operadora é suportada
        if provider not in self.SUPPORTED_PROVIDERS:
            raise ValueError(f"Operadora não suportada: {provider}")
            
        # Obter credenciais para esta operadora
        provider_creds = self.credentials_by_provider[provider]
        
        # Converter payload para string ordenada
        payload_str = json.dumps(payload, sort_keys=True)
        message = f"{payload_str}:{timestamp}:{provider_creds['api_key']}"
        
        # Gerar HMAC com a chave secreta
        signature = hmac.new(
            provider_creds['api_secret'].encode('utf-8'),
            message.encode('utf-8'),
            getattr(hashlib, self.hash_algorithm)
        ).hexdigest()
        
        return signature    # Métodos específicos de validação
    def verify_phone_number(self, phone_number: str, use_cache: bool = True) -> Dict:
        """
        Verifica se um número de telefone é válido e está ativo em Angola.
        Consulta automaticamente a operadora correta com base no prefixo do número.
        
        Args:
            phone_number: Número de telefone a ser verificado (formato internacional, ex: "+244923456789")
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Informações de validade do número, operadora e status de atividade
        """
        # Validar formato do número
        if not phone_number.startswith("+244") or len(phone_number) != 13:
            return {
                "success": False,
                "error": "Formato de número inválido. Use o formato internacional, ex: +244923456789",
                "validation": False
            }
        
        # Identificar operadora com base no prefixo
        operator_prefixes = {
            "unitel": ["921", "922", "923", "924", "925", "991", "992", "993", "994"],
            "movicel": ["910", "911", "912", "913", "914", "915", "916", "917", "990"],
            "angola_telecom": ["222", "223", "224", "225", "226", "227", "228", "229"]
        }
        
        provider = None
        prefix = phone_number[4:7]
        
        for operator, prefixes in operator_prefixes.items():
            if prefix in prefixes:
                provider = operator
                break
                
        if not provider:
            return {
                "success": False,
                "error": f"Não foi possível identificar a operadora para o prefixo {prefix}",
                "validation": False
            }
            
        # Preparar payload para verificação
        payload = {
            "phoneNumber": phone_number,
            "verifyType": "basic"
        }
        
        # Fazer requisição à operadora
        result = self.fetch_data(provider, "verify_endpoint", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        subscriber_data = result.get("data", {})
        
        return {
            "success": True,
            "provider": provider,
            "validation": True,
            "phone_number": phone_number,
            "is_active": subscriber_data.get("is_active", False),
            "subscription_type": subscriber_data.get("subscription_type", "unknown"),
            "account_age_days": subscriber_data.get("account_age_days", 0),
            "is_prepaid": subscriber_data.get("is_prepaid", True),
            "risk_score": subscriber_data.get("risk_score", 50),
            "has_recent_sim_change": subscriber_data.get("recent_sim_change", False),
            "last_activity": subscriber_data.get("last_activity_timestamp"),
            "verification_timestamp": datetime.now().isoformat()
        }
        
    def get_location_by_number(self, phone_number: str, use_cache: bool = True) -> Dict:
        """
        Obtém a localização atual aproximada de um número de telefone (baseado na torre de celular).
        
        Args:
            phone_number: Número de telefone a ser localizado (formato internacional, ex: "+244923456789")
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Informações de localização do número, incluindo coordenadas, província, município
        """
        # Validar formato do número
        if not phone_number.startswith("+244") or len(phone_number) != 13:
            return {
                "success": False,
                "error": "Formato de número inválido. Use o formato internacional, ex: +244923456789",
                "validation": False
            }
        
        # Identificar operadora com base no prefixo
        operator_prefixes = {
            "unitel": ["921", "922", "923", "924", "925", "991", "992", "993", "994"],
            "movicel": ["910", "911", "912", "913", "914", "915", "916", "917", "990"],
            "angola_telecom": ["222", "223", "224", "225", "226", "227", "228", "229"]
        }
        
        provider = None
        prefix = phone_number[4:7]
        
        for operator, prefixes in operator_prefixes.items():
            if prefix in prefixes:
                provider = operator
                break
                
        if not provider:
            return {
                "success": False,
                "error": f"Não foi possível identificar a operadora para o prefixo {prefix}",
                "validation": False
            }
        
        # Verificar se a operadora suporta localização
        if provider == "angola_telecom" or "location_endpoint" not in self.provider_apis[provider]:
            return {
                "success": False,
                "error": f"Operadora {provider} não suporta serviço de localização",
                "validation": False
            }
            
        # Preparar payload para localização
        payload = {
            "phoneNumber": phone_number,
            "locationType": "cell_tower",
            "accuracy": "high"
        }
        
        # Fazer requisição à operadora
        result = self.fetch_data(provider, "location_endpoint", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        location_data = result.get("data", {})
        
        return {
            "success": True,
            "provider": provider,
            "phone_number": phone_number,
            "latitude": location_data.get("latitude"),
            "longitude": location_data.get("longitude"),
            "accuracy_meters": location_data.get("accuracy_meters", 1000),
            "cell_tower_id": location_data.get("cell_id"),
            "province": location_data.get("province"),
            "municipality": location_data.get("municipality"),
            "neighborhood": location_data.get("neighborhood"),
            "last_updated": location_data.get("timestamp"),
            "location_timestamp": datetime.now().isoformat()
        }
        
    def get_usage_pattern(self, phone_number: str, time_period_days: int = 30, use_cache: bool = True) -> Dict:
        """
        Obtém padrões de uso e comportamento para um número de telefone.
        
        Args:
            phone_number: Número de telefone (formato internacional, ex: "+244923456789")
            time_period_days: Período de tempo em dias para buscar dados (padrão: 30 dias)
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Informações sobre padrões de uso, incluindo estatísticas e sinais de fraude
        """
        # Validar formato do número
        if not phone_number.startswith("+244") or len(phone_number) != 13:
            return {
                "success": False,
                "error": "Formato de número inválido. Use o formato internacional, ex: +244923456789",
                "validation": False
            }
        
        # Identificar operadora com base no prefixo
        operator_prefixes = {
            "unitel": ["921", "922", "923", "924", "925", "991", "992", "993", "994"],
            "movicel": ["910", "911", "912", "913", "914", "915", "916", "917", "990"],
            "angola_telecom": ["222", "223", "224", "225", "226", "227", "228", "229"]
        }
        
        provider = None
        prefix = phone_number[4:7]
        
        for operator, prefixes in operator_prefixes.items():
            if prefix in prefixes:
                provider = operator
                break
                
        if not provider:
            return {
                "success": False,
                "error": f"Não foi possível identificar a operadora para o prefixo {prefix}",
                "validation": False
            }
            
        # Preparar payload para dados de uso
        payload = {
            "phoneNumber": phone_number,
            "timePeriodDays": time_period_days,
            "includeStats": True
        }
        
        # Fazer requisição à operadora
        result = self.fetch_data(provider, "usage_endpoint", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        usage_data = result.get("data", {})
        
        return {
            "success": True,
            "provider": provider,
            "phone_number": phone_number,
            "time_period_days": time_period_days,
            "usage_statistics": {
                "calls_count": usage_data.get("calls_count", 0),
                "sms_count": usage_data.get("sms_count", 0),
                "data_usage_mb": usage_data.get("data_usage_mb", 0),
                "international_calls": usage_data.get("international_calls", 0),
                "roaming_days": usage_data.get("roaming_days", 0),
                "unique_contacts": usage_data.get("unique_contacts", 0),
                "average_call_duration": usage_data.get("avg_call_duration_seconds", 0),
                "night_activity_percent": usage_data.get("night_activity_percent", 0)
            },
            "fraud_signals": {
                "unusual_location_changes": usage_data.get("unusual_location_changes", False),
                "suspicious_international_activity": usage_data.get("suspicious_international", False),
                "abnormal_usage_pattern": usage_data.get("abnormal_pattern", False),
                "sim_swap_detected": usage_data.get("recent_sim_swap", False),
                "fraud_score": usage_data.get("fraud_score", 0)
            },
            "analysis_timestamp": datetime.now().isoformat()
        }    
    def get_device_info(self, phone_number: str, use_cache: bool = True) -> Dict:
        """
        Obtém informações sobre os dispositivos usados com determinado número de telefone.
        
        Args:
            phone_number: Número de telefone (formato internacional, ex: "+244923456789")
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Informações sobre dispositivos associados ao número
        """
        # Validar formato do número
        if not phone_number.startswith("+244") or len(phone_number) != 13:
            return {
                "success": False,
                "error": "Formato de número inválido. Use o formato internacional, ex: +244923456789",
                "validation": False
            }
        
        # Identificar operadora com base no prefixo
        operator_prefixes = {
            "unitel": ["921", "922", "923", "924", "925", "991", "992", "993", "994"],
            "movicel": ["910", "911", "912", "913", "914", "915", "916", "917", "990"],
            "angola_telecom": ["222", "223", "224", "225", "226", "227", "228", "229"]
        }
        
        provider = None
        prefix = phone_number[4:7]
        
        for operator, prefixes in operator_prefixes.items():
            if prefix in prefixes:
                provider = operator
                break
                
        if not provider:
            return {
                "success": False,
                "error": f"Não foi possível identificar a operadora para o prefixo {prefix}",
                "validation": False
            }
        
        # Verificar se a operadora suporta informações de dispositivos
        if provider == "angola_telecom" or "devices_endpoint" not in self.provider_apis[provider]:
            return {
                "success": False,
                "error": f"Operadora {provider} não suporta serviço de informações de dispositivos",
                "validation": False
            }
            
        # Preparar payload para dados de dispositivos
        payload = {
            "phoneNumber": phone_number,
            "includeHistory": True,
            "historyMonths": 3
        }
        
        # Fazer requisição à operadora
        result = self.fetch_data(provider, "devices_endpoint", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        devices_data = result.get("data", {})
        
        return {
            "success": True,
            "provider": provider,
            "phone_number": phone_number,
            "current_device": {
                "imei": devices_data.get("current_imei"),
                "model": devices_data.get("current_model"),
                "manufacturer": devices_data.get("current_manufacturer"),
                "os": devices_data.get("current_os"),
                "first_seen": devices_data.get("current_first_seen"),
                "last_seen": devices_data.get("current_last_seen")
            },
            "device_changes": devices_data.get("device_changes", []),
            "multiple_devices_detected": devices_data.get("multiple_devices", False),
            "unusual_device_pattern": devices_data.get("unusual_pattern", False),
            "fraud_risk_score": devices_data.get("fraud_risk", 0),
            "analysis_timestamp": datetime.now().isoformat()
        }
    
    def verify_phone_ownership(self, phone_number: str, name: str, bi_number: str = None, use_cache: bool = True) -> Dict:
        """
        Verifica se um número de telefone pertence à pessoa informada.
        
        Args:
            phone_number: Número de telefone (formato internacional, ex: "+244923456789")
            name: Nome completo do titular
            bi_number: Número do BI (opcional)
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Resultado da verificação de propriedade do número
        """
        # Validar formato do número
        if not phone_number.startswith("+244") or len(phone_number) != 13:
            return {
                "success": False,
                "error": "Formato de número inválido. Use o formato internacional, ex: +244923456789",
                "validation": False
            }
        
        # Identificar operadora com base no prefixo
        operator_prefixes = {
            "unitel": ["921", "922", "923", "924", "925", "991", "992", "993", "994"],
            "movicel": ["910", "911", "912", "913", "914", "915", "916", "917", "990"],
            "angola_telecom": ["222", "223", "224", "225", "226", "227", "228", "229"]
        }
        
        provider = None
        prefix = phone_number[4:7]
        
        for operator, prefixes in operator_prefixes.items():
            if prefix in prefixes:
                provider = operator
                break
                
        if not provider:
            return {
                "success": False,
                "error": f"Não foi possível identificar a operadora para o prefixo {prefix}",
                "validation": False
            }
            
        # Preparar payload para verificação
        payload = {
            "phoneNumber": phone_number,
            "verifyType": "ownership",
            "ownerName": name
        }
        
        # Adicionar BI se fornecido
        if bi_number:
            payload["identityNumber"] = bi_number
            
        # Fazer requisição à operadora
        result = self.fetch_data(provider, "verify_endpoint", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        verify_data = result.get("data", {})
        
        return {
            "success": True,
            "provider": provider,
            "phone_number": phone_number,
            "name_provided": name,
            "bi_provided": bi_number,
            "is_owner": verify_data.get("is_owner", False),
            "name_match_score": verify_data.get("name_match_score", 0),
            "identity_match": verify_data.get("identity_match", False),
            "registration_date": verify_data.get("registration_date"),
            "last_ownership_change": verify_data.get("last_ownership_change"),
            "verification_timestamp": datetime.now().isoformat()
        }
    
    def check_fraud_signals(self, phone_number: str, use_cache: bool = False) -> Dict:
        """
        Verifica sinais de fraude para um número de telefone combinando múltiplas fontes de dados.
        
        Args:
            phone_number: Número de telefone (formato internacional, ex: "+244923456789")
            use_cache: Se deve usar cache para esta requisição
        
        Returns:
            Análise de risco de fraude com base em múltiplos sinais
        """
        # Coletar informações de várias fontes
        verification = self.verify_phone_number(phone_number, use_cache)
        location = self.get_location_by_number(phone_number, use_cache)
        usage = self.get_usage_pattern(phone_number, 30, use_cache)
        
        # Tentar obter informações de dispositivos se disponível
        try:
            device_info = self.get_device_info(phone_number, use_cache)
        except Exception:
            device_info = {"success": False}
        
        # Se alguma das verificações básicas falhar, retornar erro
        if not verification.get("success"):
            return verification
            
        # Combinar sinais de fraude
        fraud_signals = []
        risk_score = 0
        
        # Verificar sinais de fraude do número
        if not verification.get("is_active", True):
            fraud_signals.append("Número inativo")
            risk_score += 30
            
        if verification.get("has_recent_sim_change", False):
            fraud_signals.append("Troca recente de SIM card")
            risk_score += 40
            
        # Verificar sinais de fraude da localização
        if location.get("success"):
            province = location.get("province", "")
            if "Luanda" not in province and usage.get("fraud_signals", {}).get("unusual_location_changes", False):
                fraud_signals.append("Localização incomum e mudanças frequentes de localização")
                risk_score += 25
                
        # Verificar sinais de fraude de uso
        if usage.get("success"):
            fraud_usage = usage.get("fraud_signals", {})
            
            if fraud_usage.get("suspicious_international_activity", False):
                fraud_signals.append("Atividade internacional suspeita")
                risk_score += 35
                
            if fraud_usage.get("abnormal_usage_pattern", False):
                fraud_signals.append("Padrão de uso anormal")
                risk_score += 20
                
            if fraud_usage.get("sim_swap_detected", False):
                fraud_signals.append("SIM swap detectado")
                risk_score += 50
                
            risk_score += min(fraud_usage.get("fraud_score", 0), 50)  # Limitar contribuição máxima
            
        # Verificar sinais de fraude de dispositivos
        if device_info.get("success"):
            if device_info.get("unusual_device_pattern", False):
                fraud_signals.append("Padrão incomum de troca de dispositivos")
                risk_score += 30
                
            if device_info.get("multiple_devices_detected", False):
                fraud_signals.append("Múltiplos dispositivos utilizados simultaneamente")
                risk_score += 25
                
            risk_score += min(device_info.get("fraud_risk_score", 0), 40)  # Limitar contribuição máxima
            
        # Normalizar score final para 0-100
        final_risk_score = min(int(risk_score), 100)
        
        # Determinar nível de risco
        risk_level = "baixo"
        if final_risk_score >= 75:
            risk_level = "muito alto"
        elif final_risk_score >= 50:
            risk_level = "alto"
        elif final_risk_score >= 25:
            risk_level = "médio"
            
        # Determinar ação recomendada
        action = "permitir"
        if final_risk_score >= 75:
            action = "bloquear"
        elif final_risk_score >= 50:
            action = "verificação adicional"
        elif final_risk_score >= 25:
            action = "monitorar"
            
        return {
            "success": True,
            "phone_number": phone_number,
            "risk_score": final_risk_score,
            "risk_level": risk_level,
            "fraud_signals": fraud_signals,
            "recommended_action": action,
            "verification_data": {
                "is_active": verification.get("is_active", False),
                "subscription_type": verification.get("subscription_type", "unknown"),
                "account_age_days": verification.get("account_age_days", 0)
            },
            "location_data": {
                "province": location.get("province", "unknown") if location.get("success") else "unknown",
                "municipality": location.get("municipality", "unknown") if location.get("success") else "unknown"
            },
            "analysis_timestamp": datetime.now().isoformat()
        }


# Exemplo de uso
if __name__ == "__main__":
    # Inicializar adaptador
    adapter = TelecomAdapter()
    
    # Verificar um número de telefone
    result = adapter.verify_phone_number("+244923456789")
    print(f"Verificação de número: {result}")
    
    # Verificar localização
    location = adapter.get_location_by_number("+244923456789")
    print(f"Localização: {location}")
    
    # Verificar padrões de uso
    usage = adapter.get_usage_pattern("+244923456789")
    print(f"Padrões de uso: {usage}")
    
    # Verificar sinais de fraude
    fraud_check = adapter.check_fraud_signals("+244923456789")
    print(f"Análise de fraude: {fraud_check}")