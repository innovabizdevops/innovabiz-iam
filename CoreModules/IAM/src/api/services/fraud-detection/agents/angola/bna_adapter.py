"""
Adaptador para Banco Nacional de Angola (BNA)

Este módulo implementa o adaptador para integração com o Banco Nacional de Angola (BNA),
permitindo validações regulatórias, verificações AML/KYC e consulta de listas restritivas
para detecção avançada de fraudes no contexto financeiro angolano.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
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
import re
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Optional, Union, Any
from .angola_data_adapters import AngolaDataAdapter

logger = logging.getLogger("bna_adapter")

class BNAAdapter(AngolaDataAdapter):
    """
    Adaptador para integração com o Banco Nacional de Angola (BNA),
    órgão regulador do sistema financeiro angolano.
    
    Fornece acesso a validações regulatórias, listas restritivas, 
    verificação de entidades financeiras e serviços de validação AML/KYC.
    """
    
    # Endpoints da API
    API_ENDPOINTS = {
        "v1": {
            "auth": "/api/v1/auth/token",
            "validate_entity": "/api/v1/entity/validate",
            "aml_check": "/api/v1/compliance/aml",
            "restricted_list": "/api/v1/compliance/restricted",
            "sanctioned_list": "/api/v1/compliance/sanctions",
            "pep_check": "/api/v1/compliance/pep",
            "bank_validation": "/api/v1/banks/validate",
            "account_validation": "/api/v1/accounts/validate",
            "license_validation": "/api/v1/licenses/validate"
        },
        "v2": {
            "auth": "/api/v2/auth/token",
            "validate_entity": "/api/v2/entity/validate",
            "aml_check": "/api/v2/compliance/aml",
            "restricted_list": "/api/v2/compliance/restricted",
            "sanctioned_list": "/api/v2/compliance/sanctions",
            "pep_check": "/api/v2/compliance/pep",
            "bank_validation": "/api/v2/banks/validate",
            "account_validation": "/api/v2/accounts/validate",
            "license_validation": "/api/v2/licenses/validate",
            "financial_activity": "/api/v2/entity/activity",
            "risk_assessment": "/api/v2/entity/risk",
            "regulatory_status": "/api/v2/entity/status"
        }
    }
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        api_version: str = "v2"
    ):
        """
        Inicializa o adaptador para o Banco Nacional de Angola (BNA).
        
        Args:
            config_path: Caminho para arquivo de configuração
            credentials_path: Caminho para arquivo de credenciais
            cache_dir: Diretório para cache de dados
            api_version: Versão da API a ser utilizada (v1 ou v2)
        """
        super().__init__(config_path, credentials_path, cache_dir)
        
        # Configurações da API
        self.api_version = api_version
        self.api_base_url = self.config.get("bna_api_base_url", "https://api-partner.bna.ao")
        self.api_timeout = self.config.get("bna_api_timeout", 60)  # timeout em segundos
        
        # Autenticação
        self.api_key = self.credentials.get("bna_api_key", "")
        self.api_secret = self.credentials.get("bna_api_secret", "")
        self.client_id = self.credentials.get("bna_client_id", "")
        self.auth_token = ""
        self.token_expires_at = None
        
        # Certificado digital (necessário para algumas operações do BNA)
        self.certificate_path = self.credentials.get("bna_certificate_path", "")
        self.certificate_password = self.credentials.get("bna_certificate_password", "")
        self.use_certificate = self.config.get("use_certificate", False)
        
        # Headers padrão
        self.default_headers = {
            "Content-Type": "application/json",
            "User-Agent": "INNOVABIZ-IAM-TrustGuard/1.0",
            "Accept": "application/json",
            "X-API-Source": "INNOVABIZ-TrustGuard",
            "X-API-Version": self.api_version
        }
        
        # Rate limiting
        self.rate_limit = {
            "max_requests_per_minute": self.config.get("bna_max_requests_per_minute", 20),
            "last_request_time": None,
            "request_count": 0
        }
        
        # Endpoints
        self.endpoints = self.API_ENDPOINTS.get(api_version, self.API_ENDPOINTS["v2"])
        
        # Verificar se temos dados de configuração válidos
        if not self.api_key or not self.api_secret:
            logger.warning("API Key e/ou Secret não configurados para o BNA")
            
    def connect(self) -> bool:
        """
        Estabelece conexão com a API do BNA e obtém token de autenticação.
        
        Returns:
            True se conectado com sucesso, False caso contrário
        """
        # Verificar se já temos um token válido
        if self._is_token_valid():
            return True
            
        try:
            # Gerar HMAC para autenticação
            timestamp = str(int(time.time()))
            nonce = str(uuid.uuid4())
            signature_message = f"{self.api_key}:{timestamp}:{nonce}"
            
            signature = hmac.new(
                self.api_secret.encode('utf-8'),
                signature_message.encode('utf-8'),
                hashlib.sha256
            ).hexdigest()
            
            # Payload para autenticação
            auth_payload = {
                "clientId": self.client_id,
                "apiKey": self.api_key,
                "timestamp": timestamp,
                "nonce": nonce,
                "signature": signature
            }
            
            # Fazer requisição de autenticação
            auth_url = f"{self.api_base_url}{self.endpoints['auth']}"
            headers = self.default_headers.copy()
            
            # Configurar certificado digital se necessário
            cert = None
            if self.use_certificate and self.certificate_path:
                cert = (self.certificate_path, self.certificate_password)
                
            response = self.session.post(
                auth_url,
                headers=headers,
                json=auth_payload,
                timeout=self.api_timeout,
                cert=cert
            )
            
            if response.status_code == 200:
                token_data = response.json()
                
                # Armazenar token e calcular expiração
                self.auth_token = token_data.get("token", "")
                
                # Calcular tempo de expiração
                expires_in = token_data.get("expiresIn", 3600)  # Default: 1 hora
                self.token_expires_at = datetime.now() + timedelta(seconds=expires_in)
                
                # Atualizar status de conexão
                self.connection_status["is_connected"] = True
                self.connection_status["last_connection_time"] = datetime.now().isoformat()
                
                logger.info("Conexão estabelecida com sucesso com a API do BNA")
                return True
            else:
                error_msg = f"Falha na autenticação com o BNA: {response.status_code} - {response.text}"
                self.handle_error("AuthenticationError", error_msg)
                return False
                
        except requests.exceptions.RequestException as e:
            error_msg = f"Erro de requisição ao BNA: {str(e)}"
            self.handle_error("RequestError", error_msg)
            return False
        except Exception as e:
            error_msg = f"Erro inesperado ao conectar ao BNA: {str(e)}"
            self.handle_error("UnexpectedError", error_msg)
            return False
    
    def _is_token_valid(self) -> bool:
        """
        Verifica se o token de autenticação atual é válido.
        
        Returns:
            True se o token for válido, False caso contrário
        """
        if not self.auth_token or not self.token_expires_at:
            return False
            
        # Verificar se o token ainda é válido (com margem de segurança de 5 minutos)
        safety_margin = timedelta(minutes=5)
        return datetime.now() < (self.token_expires_at - safety_margin)
    
    def _check_rate_limit(self) -> bool:
        """
        Verifica e gerencia rate limiting para não exceder limites da API do BNA.
        
        Returns:
            True se a requisição pode prosseguir, False caso contrário
        """
        current_time = time.time()
        
        # Primeiro request ou minuto diferente
        if not self.rate_limit["last_request_time"] or \
           (current_time - self.rate_limit["last_request_time"]) >= 60:
            self.rate_limit["request_count"] = 1
            self.rate_limit["last_request_time"] = current_time
            return True
            
        # Incrementar contador
        self.rate_limit["request_count"] += 1
        
        # Verificar se excedemos o limite
        if self.rate_limit["request_count"] > self.rate_limit["max_requests_per_minute"]:
            time_to_wait = 60 - (current_time - self.rate_limit["last_request_time"])
            if time_to_wait > 0:
                logger.warning(f"Limite de requisições ao BNA excedido. Aguardando {time_to_wait:.2f} segundos.")
                time.sleep(time_to_wait)
                # Resetar contador
                self.rate_limit["request_count"] = 1
                self.rate_limit["last_request_time"] = time.time()
                
        return True
    
    def fetch_data(self, endpoint_name: str, payload: Dict, use_cache: bool = True, cache_max_age: int = 24) -> Dict:
        """
        Busca dados da API do BNA.
        
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
                "error": f"Endpoint não especificado ou inválido para a API do BNA",
                "valid_endpoints": list(self.endpoints.keys())
            }
        
        # Verificar cache
        if use_cache:
            cache_key = f"bna_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}"
            cached_data = self.get_cached_data(cache_key, cache_max_age)
            
            if cached_data:
                logger.info(f"Dados recuperados do cache para BNA - {endpoint_name}")
                return {
                    "success": True,
                    "data": cached_data,
                    "source": "cache",
                    "cached_at": datetime.now().isoformat()
                }
        
        # Verificar conexão
        if not self._is_token_valid():
            if not self.connect():
                return {
                    "success": False,
                    "error": "Não foi possível estabelecer conexão com a API do BNA",
                    "connection_status": self.connection_status
                }
        
        # Verificar rate limiting
        self._check_rate_limit()
        
        try:
            # Preparar URL e headers
            endpoint_url = f"{self.api_base_url}{self.endpoints[endpoint_name]}"
            headers = self.default_headers.copy()
            headers["Authorization"] = f"Bearer {self.auth_token}"
            
            # Adicionar identificação de requisição
            request_id = str(uuid.uuid4())
            headers["X-Request-ID"] = request_id
            
            # Configurar certificado digital se necessário
            cert = None
            if self.use_certificate and self.certificate_path:
                cert = (self.certificate_path, self.certificate_password)
                
            # Fazer a requisição
            response = self.session.post(
                endpoint_url,
                headers=headers,
                json=payload,
                timeout=self.api_timeout,
                cert=cert
            )
            
            # Verificar resposta
            if response.status_code == 200:
                data = response.json()
                
                # Salvar em cache se solicitado
                if use_cache:
                    self.cache_data(f"bna_{endpoint_name}_{hash(json.dumps(payload, sort_keys=True))}", data)
                
                return {
                    "success": True,
                    "data": data,
                    "source": "api",
                    "request_id": request_id,
                    "timestamp": datetime.now().isoformat()
                }
            elif response.status_code == 401:
                # Token expirou, tentar reconectar
                logger.warning("Token expirado para a API do BNA, reconectando...")
                self.auth_token = ""
                self.token_expires_at = None
                
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
                error_msg = f"Erro na resposta do BNA: {response.status_code} - {response.text}"
                self.handle_error("ResponseError", error_msg)
                
                return {
                    "success": False,
                    "error": error_msg,
                    "status_code": response.status_code,
                    "request_id": request_id
                }
                
        except Exception as e:
            error_msg = f"Erro ao buscar dados do BNA: {str(e)}"
            self.handle_error("FetchError", error_msg)
            
            return {
                "success": False,
                "error": error_msg
            }    
    # Métodos de validação regulatória
    def validate_entity(self, entity_type: str, entity_id: str, entity_data: Dict = None, use_cache: bool = True) -> Dict:
        """
        Valida uma entidade (pessoa física ou jurídica) junto ao BNA.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_data: Dados adicionais da entidade para validação cruzada
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da validação da entidade
        """
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Validar formato do identificador
        if entity_type == "individual" and not re.match(r'^\d{9}[A-Z]{2}\d{3}$', entity_id):
            return {
                "success": False,
                "error": "Formato de BI inválido. Formato esperado: 9 dígitos + 2 letras + 3 dígitos",
                "validation": False
            }
            
        if entity_type == "company" and not re.match(r'^\d{10}$', entity_id):
            return {
                "success": False,
                "error": "Formato de NIF inválido. Formato esperado: 10 dígitos",
                "validation": False
            }
            
        # Preparar payload para validação
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "validationLevel": "standard"
        }
        
        # Adicionar dados adicionais se fornecidos
        if entity_data:
            payload["entityData"] = entity_data
            
        # Fazer requisição ao BNA
        result = self.fetch_data("validate_entity", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        validation_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "validation": validation_data.get("isValid", False),
            "validation_level": validation_data.get("validationLevel", "standard"),
            "validation_details": {
                "entity_exists": validation_data.get("entityExists", False),
                "entity_active": validation_data.get("entityActive", False),
                "entity_status": validation_data.get("entityStatus", "unknown"),
                "registration_date": validation_data.get("registrationDate"),
                "last_update": validation_data.get("lastUpdate")
            },
            "regulatory_status": validation_data.get("regulatoryStatus", "unknown"),
            "warning_flags": validation_data.get("warningFlags", []),
            "validation_timestamp": datetime.now().isoformat()
        }
        
    def check_aml_status(self, entity_type: str, entity_id: str, entity_name: str = None, use_cache: bool = True) -> Dict:
        """
        Realiza verificação AML (Anti-Money Laundering) para uma entidade.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_name: Nome da entidade (opcional, para validação adicional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da verificação AML
        """
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Preparar payload para verificação AML
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "checkLevel": "comprehensive"
        }
        
        # Adicionar nome se fornecido
        if entity_name:
            payload["entityName"] = entity_name
            
        # Fazer requisição ao BNA
        result = self.fetch_data("aml_check", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        aml_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "entity_name": entity_name,
            "aml_status": {
                "clear": aml_data.get("isClear", True),
                "risk_level": aml_data.get("riskLevel", "low"),
                "alerts_count": aml_data.get("alertsCount", 0)
            },
            "risk_factors": aml_data.get("riskFactors", []),
            "monitoring_recommendations": aml_data.get("recommendations", []),
            "has_suspicious_activity": aml_data.get("suspiciousActivity", False),
            "regulatory_notices": aml_data.get("regulatoryNotices", []),
            "check_timestamp": datetime.now().isoformat()
        }
        
    def check_restricted_list(self, entity_type: str, entity_id: str, entity_name: str = None, use_cache: bool = True) -> Dict:
        """
        Verifica se uma entidade está em alguma lista restritiva do BNA.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_name: Nome da entidade (opcional, para validação adicional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da verificação em listas restritivas
        """
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Preparar payload para verificação de lista restritiva
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "includeHistory": True
        }
        
        # Adicionar nome se fornecido
        if entity_name:
            payload["entityName"] = entity_name
            
        # Fazer requisição ao BNA
        result = self.fetch_data("restricted_list", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        restricted_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "entity_name": entity_name,
            "is_restricted": restricted_data.get("isRestricted", False),
            "restriction_details": restricted_data.get("restrictionDetails", []),
            "restriction_history": restricted_data.get("restrictionHistory", []),
            "restriction_level": restricted_data.get("restrictionLevel", "none"),
            "last_check_date": restricted_data.get("lastCheckDate"),
            "check_timestamp": datetime.now().isoformat()
        }
        
    def check_sanctioned_list(self, entity_type: str, entity_id: str, entity_name: str = None, use_cache: bool = True) -> Dict:
        """
        Verifica se uma entidade está em alguma lista de sanções reconhecida pelo BNA.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_name: Nome da entidade (opcional, para validação adicional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da verificação em listas de sanções
        """
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Preparar payload para verificação de lista de sanções
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "checkDomestic": True,
            "checkInternational": True
        }
        
        # Adicionar nome se fornecido
        if entity_name:
            payload["entityName"] = entity_name
            
        # Fazer requisição ao BNA
        result = self.fetch_data("sanctioned_list", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        sanctions_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "entity_name": entity_name,
            "is_sanctioned": sanctions_data.get("isSanctioned", False),
            "sanctions_details": sanctions_data.get("sanctionsDetails", []),
            "domestic_sanctions": sanctions_data.get("domesticSanctions", []),
            "international_sanctions": sanctions_data.get("internationalSanctions", []),
            "sanction_categories": sanctions_data.get("sanctionCategories", []),
            "sanction_severity": sanctions_data.get("sanctionSeverity", "none"),
            "check_timestamp": datetime.now().isoformat()
        }
        
    def check_pep_status(self, entity_type: str, entity_id: str, entity_name: str = None, use_cache: bool = True) -> Dict:
        """
        Verifica se uma entidade é uma Pessoa Politicamente Exposta (PEP).
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_name: Nome da entidade (opcional, para validação adicional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da verificação de status PEP
        """
        # Validar parâmetros
        if entity_type != "individual":
            return {
                "success": False,
                "error": "Verificação PEP só está disponível para pessoas físicas (entity_type = 'individual')",
                "validation": False
            }
            
        # Preparar payload para verificação PEP
        payload = {
            "entityId": entity_id,
            "checkFamilyMembers": True,
            "checkAssociates": True
        }
        
        # Adicionar nome se fornecido
        if entity_name:
            payload["entityName"] = entity_name
            
        # Fazer requisição ao BNA
        result = self.fetch_data("pep_check", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        pep_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_id": entity_id,
            "entity_name": entity_name,
            "is_pep": pep_data.get("isPep", False),
            "pep_category": pep_data.get("pepCategory", ""),
            "pep_since": pep_data.get("pepSince", ""),
            "pep_until": pep_data.get("pepUntil", ""),
            "pep_position": pep_data.get("pepPosition", ""),
            "pep_organization": pep_data.get("pepOrganization", ""),
            "related_peps": pep_data.get("relatedPeps", []),
            "risk_level": pep_data.get("riskLevel", "low"),
            "enhanced_due_diligence_required": pep_data.get("enhancedDueDiligenceRequired", False),
            "check_timestamp": datetime.now().isoformat()
        }    
    def validate_bank_account(self, bank_code: str, account_number: str, account_holder: Dict = None, use_cache: bool = True) -> Dict:
        """
        Valida uma conta bancária junto ao BNA.
        
        Args:
            bank_code: Código do banco (3 dígitos)
            account_number: Número da conta (IBAN angolano)
            account_holder: Dados do titular da conta (opcional, para validação cruzada)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da validação da conta bancária
        """
        # Validar formato do código do banco
        if not re.match(r'^\d{3}$', bank_code):
            return {
                "success": False,
                "error": "Formato de código bancário inválido. Formato esperado: 3 dígitos",
                "validation": False
            }
            
        # Validar formato do IBAN angolano
        if not re.match(r'^AO\d{2}00\d{3}\d{20}$', account_number):
            return {
                "success": False,
                "error": "Formato de IBAN inválido. Formato esperado: IBAN angolano (AO06 seguido de 25 dígitos)",
                "validation": False
            }
            
        # Preparar payload para validação de conta
        payload = {
            "bankCode": bank_code,
            "accountNumber": account_number,
            "validationLevel": "full"
        }
        
        # Adicionar dados do titular se fornecidos
        if account_holder:
            payload["accountHolder"] = account_holder
            
        # Fazer requisição ao BNA
        result = self.fetch_data("account_validation", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        validation_data = result.get("data", {})
        
        return {
            "success": True,
            "bank_code": bank_code,
            "account_number": account_number,
            "is_valid": validation_data.get("isValid", False),
            "account_status": validation_data.get("accountStatus", "unknown"),
            "bank_name": validation_data.get("bankName", ""),
            "branch_code": validation_data.get("branchCode", ""),
            "account_type": validation_data.get("accountType", ""),
            "holder_match": validation_data.get("holderMatch", False) if account_holder else None,
            "holder_match_score": validation_data.get("holderMatchScore", 0) if account_holder else None,
            "account_age_days": validation_data.get("accountAgeDays", 0),
            "warning_flags": validation_data.get("warningFlags", []),
            "validation_timestamp": datetime.now().isoformat()
        }
        
    def validate_banking_license(self, license_id: str, entity_id: str = None, use_cache: bool = True) -> Dict:
        """
        Valida uma licença bancária ou financeira emitida pelo BNA.
        
        Args:
            license_id: Identificador da licença
            entity_id: NIF da entidade detentora da licença (opcional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da validação da licença
        """
        # Preparar payload para validação de licença
        payload = {
            "licenseId": license_id
        }
        
        # Adicionar NIF se fornecido
        if entity_id:
            payload["entityId"] = entity_id
            
        # Fazer requisição ao BNA
        result = self.fetch_data("license_validation", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        license_data = result.get("data", {})
        
        return {
            "success": True,
            "license_id": license_id,
            "entity_id": entity_id,
            "is_valid": license_data.get("isValid", False),
            "license_type": license_data.get("licenseType", ""),
            "license_category": license_data.get("licenseCategory", ""),
            "license_status": license_data.get("licenseStatus", ""),
            "issue_date": license_data.get("issueDate", ""),
            "expiry_date": license_data.get("expiryDate", ""),
            "authorized_activities": license_data.get("authorizedActivities", []),
            "restrictions": license_data.get("restrictions", []),
            "regulatory_notes": license_data.get("regulatoryNotes", []),
            "validation_timestamp": datetime.now().isoformat()
        }
        
    def check_financial_activity(self, entity_type: str, entity_id: str, time_period_months: int = 12, use_cache: bool = True) -> Dict:
        """
        Obtém informações sobre a atividade financeira de uma entidade.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            time_period_months: Período de tempo em meses para buscar dados (padrão: 12 meses)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Informações sobre a atividade financeira da entidade
        """
        # Verificar se este endpoint existe na versão da API
        if "financial_activity" not in self.endpoints:
            return {
                "success": False,
                "error": f"Endpoint 'financial_activity' não disponível na versão {self.api_version} da API",
                "available_endpoints": list(self.endpoints.keys())
            }
            
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Preparar payload para busca de atividade financeira
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "timePeriodMonths": time_period_months,
            "includeDetails": True
        }
            
        # Fazer requisição ao BNA
        result = self.fetch_data("financial_activity", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        activity_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "time_period_months": time_period_months,
            "financial_profile": {
                "activity_level": activity_data.get("activityLevel", "medium"),
                "transaction_volume": activity_data.get("transactionVolume", 0),
                "transaction_value": activity_data.get("transactionValue", 0),
                "cross_border_percentage": activity_data.get("crossBorderPercentage", 0),
                "typical_transaction_size": activity_data.get("typicalTransactionSize", 0)
            },
            "risk_indicators": {
                "unusual_patterns": activity_data.get("unusualPatterns", False),
                "sudden_changes": activity_data.get("suddenChanges", False),
                "suspicious_jurisdictions": activity_data.get("suspiciousJurisdictions", False),
                "high_risk_activities": activity_data.get("highRiskActivities", False),
            },
            "activity_sectors": activity_data.get("activitySectors", []),
            "main_counterparties": activity_data.get("mainCounterparties", []),
            "financial_institutions": activity_data.get("financialInstitutions", []),
            "analysis_timestamp": datetime.now().isoformat()
        }
        
    def assess_risk(self, entity_type: str, entity_id: str, assessment_type: str = "comprehensive", use_cache: bool = False) -> Dict:
        """
        Realiza uma avaliação de risco regulatório e financeiro para uma entidade.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            assessment_type: Tipo de avaliação ("basic", "standard" ou "comprehensive")
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado da avaliação de risco
        """
        # Verificar se este endpoint existe na versão da API
        if "risk_assessment" not in self.endpoints:
            return {
                "success": False,
                "error": f"Endpoint 'risk_assessment' não disponível na versão {self.api_version} da API",
                "available_endpoints": list(self.endpoints.keys())
            }
            
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        if assessment_type not in ["basic", "standard", "comprehensive"]:
            return {
                "success": False,
                "error": "Tipo de avaliação inválido. Use 'basic', 'standard' ou 'comprehensive'.",
                "validation": False
            }
            
        # Preparar payload para avaliação de risco
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "assessmentType": assessment_type,
            "includeRecommendations": True
        }
            
        # Fazer requisição ao BNA
        result = self.fetch_data("risk_assessment", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        risk_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "assessment_type": assessment_type,
            "risk_scores": {
                "overall_risk": risk_data.get("overallRisk", 0),
                "financial_risk": risk_data.get("financialRisk", 0),
                "regulatory_risk": risk_data.get("regulatoryRisk", 0),
                "compliance_risk": risk_data.get("complianceRisk", 0),
                "fraud_risk": risk_data.get("fraudRisk", 0)
            },
            "risk_level": risk_data.get("riskLevel", "medium"),
            "risk_factors": risk_data.get("riskFactors", []),
            "mitigating_factors": risk_data.get("mitigatingFactors", []),
            "recommendations": risk_data.get("recommendations", []),
            "required_actions": risk_data.get("requiredActions", []),
            "assessment_timestamp": datetime.now().isoformat()
        }
        
    def check_regulatory_status(self, entity_type: str, entity_id: str, use_cache: bool = True) -> Dict:
        """
        Verifica o status regulatório de uma entidade junto ao BNA.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Informações sobre o status regulatório da entidade
        """
        # Verificar se este endpoint existe na versão da API
        if "regulatory_status" not in self.endpoints:
            return {
                "success": False,
                "error": f"Endpoint 'regulatory_status' não disponível na versão {self.api_version} da API",
                "available_endpoints": list(self.endpoints.keys())
            }
            
        # Validar parâmetros
        if entity_type not in ["individual", "company"]:
            return {
                "success": False,
                "error": "Tipo de entidade inválido. Use 'individual' para pessoas físicas ou 'company' para empresas.",
                "validation": False
            }
            
        # Preparar payload para verificação de status regulatório
        payload = {
            "entityType": entity_type,
            "entityId": entity_id,
            "includeHistory": True
        }
            
        # Fazer requisição ao BNA
        result = self.fetch_data("regulatory_status", payload, use_cache)
        
        # Se ocorreu erro na requisição, retornar o erro
        if not result.get("success"):
            return result
            
        # Extrair dados relevantes
        status_data = result.get("data", {})
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "regulatory_status": status_data.get("status", "unknown"),
            "compliance_level": status_data.get("complianceLevel", "unknown"),
            "licenses": status_data.get("licenses", []),
            "authorizations": status_data.get("authorizations", []),
            "open_investigations": status_data.get("openInvestigations", 0),
            "regulatory_actions": status_data.get("regulatoryActions", []),
            "status_history": status_data.get("statusHistory", []),
            "reporting_obligations": status_data.get("reportingObligations", []),
            "status_timestamp": datetime.now().isoformat()
        }    
    # Métodos de análise combinada
    def comprehensive_entity_check(self, entity_type: str, entity_id: str, entity_name: str = None, use_cache: bool = True) -> Dict:
        """
        Realiza uma verificação abrangente da entidade, combinando múltiplas verificações.
        
        Args:
            entity_type: Tipo de entidade ("individual" ou "company")
            entity_id: Identificador da entidade (BI para individual, NIF para company)
            entity_name: Nome da entidade (opcional, para validação adicional)
            use_cache: Se deve usar cache para esta requisição
            
        Returns:
            Resultado combinado das verificações
        """
        results = {}
        
        # Validação da entidade
        results["entity_validation"] = self.validate_entity(entity_type, entity_id, use_cache=use_cache)
        
        # Verificação AML
        results["aml_check"] = self.check_aml_status(entity_type, entity_id, entity_name, use_cache=use_cache)
        
        # Verificação em listas restritivas
        results["restricted_list"] = self.check_restricted_list(entity_type, entity_id, entity_name, use_cache=use_cache)
        
        # Verificação em listas de sanções
        results["sanctions"] = self.check_sanctioned_list(entity_type, entity_id, entity_name, use_cache=use_cache)
        
        # Verificação PEP (apenas para pessoas físicas)
        if entity_type == "individual":
            results["pep_status"] = self.check_pep_status(entity_type, entity_id, entity_name, use_cache=use_cache)
        
        # Verificar status regulatório (se disponível na versão da API)
        if "regulatory_status" in self.endpoints:
            results["regulatory_status"] = self.check_regulatory_status(entity_type, entity_id, use_cache=use_cache)
        
        # Verificar atividade financeira (se disponível na versão da API)
        if "financial_activity" in self.endpoints:
            results["financial_activity"] = self.check_financial_activity(entity_type, entity_id, use_cache=use_cache)
        
        # Avaliação de risco (se disponível na versão da API)
        if "risk_assessment" in self.endpoints:
            results["risk_assessment"] = self.assess_risk(entity_type, entity_id, use_cache=use_cache)
        
        # Cálculo da pontuação de risco global
        risk_score = self._calculate_combined_risk_score(results)
        
        # Decisão recomendada com base no score
        recommendation = self._get_recommendation_from_risk_score(risk_score)
        
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "entity_name": entity_name,
            "risk_score": risk_score,
            "risk_level": recommendation["risk_level"],
            "recommended_action": recommendation["action"],
            "verification_details": results,
            "warning_flags": self._extract_warning_flags(results),
            "analysis_timestamp": datetime.now().isoformat()
        }
    
    def _calculate_combined_risk_score(self, results: Dict) -> float:
        """
        Calcula um score combinado de risco com base nos resultados das verificações.
        
        Args:
            results: Resultados das várias verificações
            
        Returns:
            Score de risco (0-100, onde maior = mais risco)
        """
        score = 0
        components = []
        
        # Validação da entidade
        if "entity_validation" in results and results["entity_validation"].get("success"):
            validation = results["entity_validation"]
            if not validation.get("validation"):
                score += 30
                components.append(("entity_invalid", 30))
            
            if "warning_flags" in validation and validation["warning_flags"]:
                flag_score = min(len(validation["warning_flags"]) * 5, 15)
                score += flag_score
                components.append(("entity_warning_flags", flag_score))
        
        # Verificação AML
        if "aml_check" in results and results["aml_check"].get("success"):
            aml = results["aml_check"]
            if "aml_status" in aml and not aml["aml_status"].get("clear", True):
                risk_level = aml["aml_status"].get("risk_level", "low")
                risk_score = {"low": 10, "medium": 20, "high": 40, "critical": 60}.get(risk_level, 0)
                score += risk_score
                components.append(("aml_risk", risk_score))
        
        # Listas restritivas
        if "restricted_list" in results and results["restricted_list"].get("success"):
            restricted = results["restricted_list"]
            if restricted.get("is_restricted", False):
                restriction_level = restricted.get("restriction_level", "low")
                risk_score = {"low": 15, "medium": 30, "high": 45, "critical": 70}.get(restriction_level, 0)
                score += risk_score
                components.append(("restricted_list", risk_score))
        
        # Sanções
        if "sanctions" in results and results["sanctions"].get("success"):
            sanctions = results["sanctions"]
            if sanctions.get("is_sanctioned", False):
                sanction_severity = sanctions.get("sanction_severity", "low")
                risk_score = {"low": 20, "medium": 40, "high": 60, "critical": 80}.get(sanction_severity, 0)
                score += risk_score
                components.append(("sanctions", risk_score))
        
        # Status PEP
        if "pep_status" in results and results["pep_status"].get("success"):
            pep = results["pep_status"]
            if pep.get("is_pep", False):
                risk_level = pep.get("risk_level", "low")
                risk_score = {"low": 10, "medium": 20, "high": 30}.get(risk_level, 0)
                score += risk_score
                components.append(("pep_status", risk_score))
        
        # Atividade financeira
        if "financial_activity" in results and results["financial_activity"].get("success"):
            activity = results["financial_activity"]
            if "risk_indicators" in activity:
                indicators = activity["risk_indicators"]
                risk_count = sum(1 for v in indicators.values() if v)
                risk_score = min(risk_count * 10, 30)
                score += risk_score
                components.append(("financial_activity", risk_score))
        
        # Avaliação de risco
        if "risk_assessment" in results and results["risk_assessment"].get("success"):
            assessment = results["risk_assessment"]
            if "risk_scores" in assessment:
                overall_risk = assessment["risk_scores"].get("overall_risk", 0)
                # Normalizar para escala 0-100 se necessário
                if overall_risk > 0 and overall_risk <= 1:
                    overall_risk *= 100
                score = max(score, overall_risk)  # Usar o maior valor entre score calculado e avaliação de risco
                components.append(("risk_assessment", overall_risk))
        
        # Limitar score final a 100
        return min(score, 100)
    
    def _get_recommendation_from_risk_score(self, risk_score: float) -> Dict:
        """
        Determina a recomendação baseada no score de risco.
        
        Args:
            risk_score: Score de risco (0-100)
            
        Returns:
            Dicionário com nível de risco e ação recomendada
        """
        if risk_score < 20:
            return {
                "risk_level": "low",
                "action": "allow"
            }
        elif risk_score < 40:
            return {
                "risk_level": "medium",
                "action": "monitor"
            }
        elif risk_score < 70:
            return {
                "risk_level": "high",
                "action": "verify"
            }
        else:
            return {
                "risk_level": "critical",
                "action": "block"
            }
    
    def _extract_warning_flags(self, results: Dict) -> List[Dict]:
        """
        Extrai e consolida todas as flags de aviso dos resultados.
        
        Args:
            results: Resultados das várias verificações
            
        Returns:
            Lista consolidada de flags de aviso
        """
        warning_flags = []
        
        # Validação da entidade
        if "entity_validation" in results and results["entity_validation"].get("success"):
            validation = results["entity_validation"]
            if not validation.get("validation"):
                warning_flags.append({
                    "source": "entity_validation",
                    "flag": "invalid_entity",
                    "description": f"Entidade inválida ou não registrada no BNA"
                })
            
            if "warning_flags" in validation:
                for flag in validation.get("warning_flags", []):
                    warning_flags.append({
                        "source": "entity_validation",
                        "flag": flag.get("code", "warning"),
                        "description": flag.get("description", "Flag de aviso na validação da entidade")
                    })
        
        # Verificação AML
        if "aml_check" in results and results["aml_check"].get("success"):
            aml = results["aml_check"]
            if "aml_status" in aml and not aml["aml_status"].get("clear", True):
                warning_flags.append({
                    "source": "aml_check",
                    "flag": "aml_risk",
                    "description": f"Risco AML nível {aml['aml_status'].get('risk_level', 'desconhecido')}"
                })
                
            # Adicionar fatores de risco específicos
            for factor in aml.get("risk_factors", []):
                warning_flags.append({
                    "source": "aml_check",
                    "flag": factor.get("code", "risk_factor"),
                    "description": factor.get("description", "Fator de risco AML")
                })
        
        # Lista restritiva
        if "restricted_list" in results and results["restricted_list"].get("success"):
            restricted = results["restricted_list"]
            if restricted.get("is_restricted", False):
                warning_flags.append({
                    "source": "restricted_list",
                    "flag": "on_restricted_list",
                    "description": f"Entidade em lista restritiva do BNA"
                })
                
                # Adicionar detalhes da restrição
                for detail in restricted.get("restriction_details", []):
                    warning_flags.append({
                        "source": "restricted_list",
                        "flag": detail.get("code", "restriction"),
                        "description": detail.get("description", "Detalhe de restrição")
                    })
        
        # Lista de sanções
        if "sanctions" in results and results["sanctions"].get("success"):
            sanctions = results["sanctions"]
            if sanctions.get("is_sanctioned", False):
                warning_flags.append({
                    "source": "sanctions",
                    "flag": "on_sanctions_list",
                    "description": f"Entidade em lista de sanções"
                })
                
                # Adicionar categorias de sanções
                for category in sanctions.get("sanction_categories", []):
                    warning_flags.append({
                        "source": "sanctions",
                        "flag": f"sanction_{category}",
                        "description": f"Sanção na categoria: {category}"
                    })
        
        # Status PEP
        if "pep_status" in results and results["pep_status"].get("success"):
            pep = results["pep_status"]
            if pep.get("is_pep", False):
                warning_flags.append({
                    "source": "pep_status",
                    "flag": "pep_identified",
                    "description": f"Pessoa Politicamente Exposta identificada, categoria: {pep.get('pep_category', 'não especificada')}"
                })
        
        # Atividade financeira
        if "financial_activity" in results and results["financial_activity"].get("success"):
            activity = results["financial_activity"]
            if "risk_indicators" in activity:
                indicators = activity["risk_indicators"]
                for indicator, value in indicators.items():
                    if value:
                        warning_flags.append({
                            "source": "financial_activity",
                            "flag": indicator,
                            "description": f"Indicador de risco na atividade financeira: {indicator}"
                        })
        
        return warning_flags
    # Exemplos de utilização e aplicações
    @staticmethod
    def usage_examples() -> str:
        """
        Retorna exemplos de uso do adaptador BNA.
        
        Returns:
            String com exemplos de código para usar o adaptador
        """
        examples = """
# Exemplo 1: Verificação básica de pessoa física
from fraud_detection.agents.angola.bna_adapter import BNAAdapter

# Inicializar o adaptador
bna = BNAAdapter(
    config_path="/path/to/config.json",
    credentials_path="/path/to/credentials.json",
    api_version="v2"
)

# Conectar à API do BNA
bna.connect()

# Validar uma pessoa física
result = bna.validate_entity(
    entity_type="individual",
    entity_id="123456789AB123"  # BI angolano
)

# Verificar status AML
aml_result = bna.check_aml_status(
    entity_type="individual", 
    entity_id="123456789AB123",
    entity_name="João Silva"
)

# Verificar status PEP
pep_result = bna.check_pep_status(
    entity_type="individual",
    entity_id="123456789AB123"
)

# Exemplo 2: Verificação de empresa
# Validar uma empresa
company_result = bna.validate_entity(
    entity_type="company",
    entity_id="1234567890"  # NIF angolano
)

# Verificar licença financeira
license_result = bna.validate_banking_license(
    license_id="BNA-LIC-2023-12345",
    entity_id="1234567890"
)

# Exemplo 3: Verificação abrangente
# Realizar verificação completa de uma entidade
comprehensive_result = bna.comprehensive_entity_check(
    entity_type="individual",
    entity_id="123456789AB123",
    entity_name="João Silva"
)

# Verificar score de risco e ação recomendada
risk_score = comprehensive_result.get("risk_score")
risk_level = comprehensive_result.get("risk_level")
recommended_action = comprehensive_result.get("recommended_action")

print(f"Score de risco: {risk_score}")
print(f"Nível de risco: {risk_level}")
print(f"Ação recomendada: {recommended_action}")

# Exemplo 4: Validação de conta bancária
account_result = bna.validate_bank_account(
    bank_code="045",
    account_number="AO06004500000123456789012",
    account_holder={
        "name": "João Silva",
        "id": "123456789AB123",
        "id_type": "BI"
    }
)

# Verificar se a conta é válida e corresponde ao titular
if account_result.get("is_valid") and account_result.get("holder_match"):
    print("Conta bancária válida e corresponde ao titular")
else:
    print("Problema com a validação da conta bancária")
"""
        
        return examples

if __name__ == "__main__":
    # Exemplo de configuração e utilização
    import sys
    import os
    import json
    
    # Configurar logging
    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    
    # Verificar se foi passado um caminho para o arquivo de configuração
    config_path = sys.argv[1] if len(sys.argv) > 1 else None
    credentials_path = sys.argv[2] if len(sys.argv) > 2 else None
    
    # Inicializar o adaptador
    try:
        bna = BNAAdapter(config_path, credentials_path)
        if bna.connect():
            print("Conexão com BNA estabelecida com sucesso!")
            
            # Exemplo de verificação de entidade
            print("\nRealizando verificação de entidade...")
            entity_result = bna.validate_entity(
                entity_type="individual", 
                entity_id="123456789AB123"  # Exemplo de BI
            )
            print(f"Resultado: {json.dumps(entity_result, indent=2)}")
            
            # Exemplo de verificação AML
            print("\nRealizando verificação AML...")
            aml_result = bna.check_aml_status(
                entity_type="individual", 
                entity_id="123456789AB123"
            )
            print(f"Resultado: {json.dumps(aml_result, indent=2)}")
            
        else:
            print("Não foi possível estabelecer conexão com o BNA.")
    except Exception as e:
        print(f"Erro ao inicializar o adaptador BNA: {str(e)}")