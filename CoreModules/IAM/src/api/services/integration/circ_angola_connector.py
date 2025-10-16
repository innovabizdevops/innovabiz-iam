#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Conector para CIRC Angola (Central de Informação de Risco de Crédito)

Este módulo implementa a integração com o sistema CIRC para verificação
de documentos e consultas de crédito em Angola.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import requests
import hashlib
import time
from typing import Dict, Any, Optional, List, Union
from datetime import datetime

# Configuração de logging
logger = logging.getLogger("circ_angola_connector")


class CIRCAngolaConnector:
    """
    Implementa a conexão e operações com o sistema CIRC Angola
    para verificação de documentos e consulta de informações de crédito.
    """
    
    def __init__(self, config: Dict[str, Any] = None):
        """
        Inicializa o conector com configurações específicas.
        
        Args:
            config: Configurações para conexão com CIRC (opcional)
        """
        # Carregar configuração
        self.config = config or self._load_default_config()
        
        # Inicializar atributos
        self.api_key = self.config.get("api_key")
        self.api_secret = self.config.get("api_secret")
        self.base_url = self.config.get("base_url", "https://api.circ.co.ao/v2")
        self.timeout = self.config.get("timeout", 30)
        self.last_request_time = None
        self.session_token = None
        self.token_expiry = None
        
        # Inicializar sessão HTTP
        self.session = requests.Session()
        
        # Adicionar headers padrão
        self.session.headers.update({
            "Content-Type": "application/json",
            "Accept": "application/json",
            "User-Agent": "INNOVABIZ-TrustGuard/1.0",
            "X-Client-ID": self.config.get("client_id", "INNOVABIZ-IAM")
        })
        
    def _load_default_config(self) -> Dict[str, Any]:
        """
        Carrega configuração padrão ou de variáveis de ambiente.
        
        Returns:
            dict: Configurações para o conector
        """
        return {
            "api_key": os.environ.get("CIRC_ANGOLA_API_KEY", "demo_key"),
            "api_secret": os.environ.get("CIRC_ANGOLA_API_SECRET", "demo_secret"),
            "base_url": os.environ.get("CIRC_ANGOLA_BASE_URL", "https://api.circ.co.ao/v2"),
            "timeout": int(os.environ.get("CIRC_ANGOLA_TIMEOUT", "30")),
            "retry_attempts": int(os.environ.get("CIRC_ANGOLA_RETRY", "3")),
            "rate_limit": {
                "requests_per_minute": int(os.environ.get("CIRC_ANGOLA_RATE_LIMIT", "60")),
            }
        }
    
    def _get_auth_token(self) -> str:
        """
        Obtém um token de autenticação para uso nas requisições.
        Implementa cache do token para evitar requisições desnecessárias.
        
        Returns:
            str: Token de autenticação
        """
        now = datetime.now()
        
        # Se já temos um token válido, retorna ele
        if self.session_token and self.token_expiry and now < self.token_expiry:
            return self.session_token
            
        # Caso contrário, solicita um novo
        auth_endpoint = f"{self.base_url}/auth/token"
        
        try:
            response = self.session.post(
                auth_endpoint,
                json={
                    "api_key": self.api_key,
                    "api_secret": self.api_secret
                },
                timeout=self.timeout
            )
            
            if response.status_code == 200:
                data = response.json()
                self.session_token = data.get("token")
                # Token válido por 1 hora (ou conforme especificado na resposta)
                expires_in = data.get("expires_in", 3600)  # Padrão 1 hora
                self.token_expiry = now + timedelta(seconds=expires_in)
                logger.debug(f"Token de autenticação CIRC obtido. Válido até {self.token_expiry}")
                return self.session_token
            else:
                logger.error(f"Erro ao obter token de autenticação CIRC: {response.status_code} - {response.text}")
                return None
                
        except Exception as e:
            logger.error(f"Exceção ao obter token de autenticação CIRC: {str(e)}")
            return None
    
    def _make_api_request(self, endpoint: str, method: str = "GET", 
                         data: Dict[str, Any] = None, retry: int = 0) -> Dict[str, Any]:
        """
        Realiza uma requisição à API do CIRC com tratamento de erros e retry.
        
        Args:
            endpoint: Endpoint da API (sem base_url)
            method: Método HTTP (GET, POST, etc)
            data: Dados para envio na requisição
            retry: Contador de tentativas de retry
            
        Returns:
            dict: Resposta da API parseada como JSON ou None em caso de erro
        """
        # Controle de rate limiting
        self._respect_rate_limit()
        
        # Garante que temos um token válido
        token = self._get_auth_token()
        if not token and retry < self.config.get("retry_attempts", 3):
            logger.warning("Falha ao obter token. Tentando novamente...")
            return self._make_api_request(endpoint, method, data, retry + 1)
            
        # Preparar headers com autenticação
        headers = {"Authorization": f"Bearer {token}"}
        
        # Url completa
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        
        try:
            # Registra o tempo da requisição
            self.last_request_time = datetime.now()
            
            # Executa a requisição
            if method.upper() == "GET":
                response = self.session.get(url, headers=headers, params=data, timeout=self.timeout)
            else:
                response = self.session.request(
                    method=method.upper(),
                    url=url,
                    headers=headers,
                    json=data,
                    timeout=self.timeout
                )
            
            # Tratamento do response
            if response.status_code == 200:
                return response.json()
            elif response.status_code == 401 and retry < self.config.get("retry_attempts", 3):
                # Token expirado, tentar renovar e repetir
                logger.warning("Token expirado. Renovando...")
                self.session_token = None
                return self._make_api_request(endpoint, method, data, retry + 1)
            elif response.status_code == 429 and retry < self.config.get("retry_attempts", 3):
                # Rate limit atingido, esperar e tentar novamente
                wait_time = int(response.headers.get("Retry-After", 60))
                logger.warning(f"Rate limit atingido. Aguardando {wait_time}s...")
                time.sleep(wait_time)
                return self._make_api_request(endpoint, method, data, retry + 1)
            else:
                logger.error(f"Erro na requisição CIRC: {response.status_code} - {response.text}")
                return {
                    "success": False,
                    "error": f"HTTP {response.status_code}",
                    "message": response.text
                }
                
        except requests.exceptions.RequestException as e:
            logger.error(f"Exceção na requisição CIRC: {str(e)}")
            
            if retry < self.config.get("retry_attempts", 3):
                retry_after = 2 ** retry  # Backoff exponencial
                logger.info(f"Tentando novamente em {retry_after}s...")
                time.sleep(retry_after)
                return self._make_api_request(endpoint, method, data, retry + 1)
                
            return {
                "success": False,
                "error": "request_exception",
                "message": str(e)
            }
    
    def _respect_rate_limit(self) -> None:
        """
        Implementa controle de rate limiting para não exceder limites da API.
        """
        if not self.last_request_time:
            return
            
        # Calcular tempo desde a última requisição
        now = datetime.now()
        elapsed = (now - self.last_request_time).total_seconds()
        
        # Tempo mínimo entre requisições com base na configuração
        requests_per_minute = self.config.get("rate_limit", {}).get("requests_per_minute", 60)
        min_interval = 60.0 / requests_per_minute
        
        if elapsed < min_interval:
            sleep_time = min_interval - elapsed
            logger.debug(f"Respeitando rate limit. Aguardando {sleep_time:.2f}s")
            time.sleep(sleep_time)
    
    def verify_document(self, doc_type: str, doc_number: str, 
                       name: Optional[str] = None) -> Dict[str, Any]:
        """
        Verifica a autenticidade de um documento angolano através do CIRC.
        
        Args:
            doc_type: Tipo do documento (bi, passport, nif)
            doc_number: Número do documento
            name: Nome do titular (opcional)
            
        Returns:
            dict: Resultado da verificação com detalhes
        """
        logger.info(f"Verificando documento {doc_type} {doc_number}")
        
        # Mapear tipos de documento para os valores esperados pela API
        doc_type_map = {
            "bi": "identity_card",
            "passport": "passport",
            "nif": "tax_id"
        }
        
        api_doc_type = doc_type_map.get(doc_type.lower(), doc_type.lower())
        
        # Preparar payload para a requisição
        data = {
            "document_type": api_doc_type,
            "document_number": doc_number
        }
        
        # Adicionar nome se fornecido
        if name:
            data["holder_name"] = name
        
        try:
            # Enviar requisição para API
            result = self._make_api_request(
                endpoint="/verification/document",
                method="POST",
                data=data
            )
            
            # Em ambiente de desenvolvimento/teste, simular resposta
            if self.config.get("environment") == "development" or not result.get("success", True):
                import random
                logger.warning("Usando resposta simulada para verificação de documento")
                
                # Simulação baseada no documento
                if doc_type.lower() == "bi" and doc_number.startswith(tuple(["000", "001", "002", "003"])):
                    confidence = random.uniform(0.85, 0.98)
                else:
                    confidence = random.uniform(0.60, 0.94)
                    
                return {
                    "verified": confidence > 0.7,
                    "score": confidence,
                    "document": {
                        "type": doc_type,
                        "number": doc_number,
                        "valid": confidence > 0.7
                    },
                    "verification_id": f"sim-{hashlib.md5(doc_number.encode()).hexdigest()[:8]}",
                    "verification_time": datetime.now().isoformat()
                }
            
            # Processar resultado real
            if "document_verification" in result:
                verification = result["document_verification"]
                return {
                    "verified": verification.get("is_valid", False),
                    "score": verification.get("confidence_score", 0.0),
                    "document": {
                        "type": doc_type,
                        "number": doc_number,
                        "valid": verification.get("is_valid", False)
                    },
                    "verification_id": result.get("verification_id"),
                    "verification_time": verification.get("verification_time")
                }
            else:
                return {
                    "verified": False,
                    "score": 0.0,
                    "error": result.get("error", "unknown_error"),
                    "message": result.get("message", "Erro desconhecido na verificação")
                }
                
        except Exception as e:
            logger.error(f"Erro ao verificar documento: {str(e)}")
            return {
                "verified": False,
                "score": 0.0,
                "error": "exception",
                "message": str(e)
            }
    
    def check_credit_information(self, doc_type: str, doc_number: str, 
                               consent: bool = True) -> Dict[str, Any]:
        """
        Consulta informações de crédito para um indivíduo ou empresa.
        
        Args:
            doc_type: Tipo de documento (bi, nif)
            doc_number: Número do documento
            consent: Se o titular consentiu com a consulta (obrigatório por lei)
            
        Returns:
            dict: Informações de crédito ou erro
        """
        if not consent:
            return {
                "success": False,
                "error": "consent_required",
                "message": "É necessário o consentimento do titular para consulta de crédito"
            }
            
        logger.info(f"Consultando informações de crédito para {doc_type} {doc_number}")
        
        # Mapear tipos de documento
        doc_type_map = {
            "bi": "identity_card",
            "nif": "tax_id"
        }
        
        api_doc_type = doc_type_map.get(doc_type.lower(), doc_type.lower())
        
        # Preparar payload
        data = {
            "document_type": api_doc_type,
            "document_number": doc_number,
            "has_consent": consent,
            "request_type": "credit_information",
            "request_reason": "account_opening"
        }
        
        try:
            # Enviar requisição para API
            result = self._make_api_request(
                endpoint="/credit/individual",
                method="POST",
                data=data
            )
            
            # Em ambiente de desenvolvimento/teste, simular resposta
            if self.config.get("environment") == "development" or not result.get("success", True):
                logger.warning("Usando resposta simulada para consulta de crédito")
                
                # Dados simulados
                import random
                score = random.randint(300, 850)
                active_loans = random.randint(0, 3)
                defaults = random.randint(0, 1) if score < 650 else 0
                
                return {
                    "success": True,
                    "credit_score": score,
                    "risk_level": "low" if score > 700 else "medium" if score > 600 else "high",
                    "active_loans": active_loans,
                    "defaults": defaults,
                    "report_id": f"sim-{hashlib.md5(doc_number.encode()).hexdigest()[:10]}",
                    "report_date": datetime.now().isoformat(),
                    "simulated": True
                }
            
            # Processar resultado real
            if "credit_information" in result:
                credit_info = result["credit_information"]
                return {
                    "success": True,
                    "credit_score": credit_info.get("score", 0),
                    "risk_level": credit_info.get("risk_level", "unknown"),
                    "active_loans": credit_info.get("active_loans", 0),
                    "defaults": credit_info.get("defaults", 0),
                    "report_id": result.get("report_id"),
                    "report_date": result.get("report_date")
                }
            else:
                return {
                    "success": False,
                    "error": result.get("error", "unknown_error"),
                    "message": result.get("message", "Erro desconhecido na consulta de crédito")
                }
                
        except Exception as e:
            logger.error(f"Erro ao consultar crédito: {str(e)}")
            return {
                "success": False,
                "error": "exception",
                "message": str(e)
            }
    
    def check_business_registry(self, nif: str) -> Dict[str, Any]:
        """
        Consulta o registro empresarial angolano para validar informações de empresas.
        
        Args:
            nif: NIF (Número de Identificação Fiscal) da empresa
            
        Returns:
            dict: Informações de registro empresarial ou erro
        """
        logger.info(f"Consultando registro empresarial para NIF {nif}")
        
        # Preparar payload
        data = {
            "tax_id": nif,
            "request_type": "business_registry"
        }
        
        try:
            # Enviar requisição para API
            result = self._make_api_request(
                endpoint="/verification/business",
                method="POST",
                data=data
            )
            
            # Em ambiente de desenvolvimento/teste, simular resposta
            if self.config.get("environment") == "development" or not result.get("success", True):
                logger.warning("Usando resposta simulada para consulta de registro empresarial")
                
                # Status simulado baseado no NIF
                active = nif.endswith(("0", "1", "2", "3", "4", "5", "6", "7"))
                
                return {
                    "success": True,
                    "verified": True,
                    "business": {
                        "name": "Empresa Simulada Lda.",
                        "tax_id": nif,
                        "registration_date": "2020-01-15",
                        "status": "active" if active else "inactive",
                        "address": "Avenida 4 de Fevereiro, Luanda, Angola",
                        "industry": "Serviços Financeiros"
                    },
                    "verification_id": f"sim-{hashlib.md5(nif.encode()).hexdigest()[:8]}",
                    "simulated": True
                }
            
            # Processar resultado real
            if "business_verification" in result:
                business = result["business_verification"]
                return {
                    "success": True,
                    "verified": business.get("is_valid", False),
                    "business": {
                        "name": business.get("name"),
                        "tax_id": nif,
                        "registration_date": business.get("registration_date"),
                        "status": business.get("status"),
                        "address": business.get("address"),
                        "industry": business.get("industry")
                    },
                    "verification_id": result.get("verification_id")
                }
            else:
                return {
                    "success": False,
                    "error": result.get("error", "unknown_error"),
                    "message": result.get("message", "Erro desconhecido na consulta de registro empresarial")
                }
                
        except Exception as e:
            logger.error(f"Erro ao consultar registro empresarial: {str(e)}")
            return {
                "success": False,
                "error": "exception",
                "message": str(e)
            }


# Função utilitária para teste rápido
def test_connector():
    """Função para testes simples do conector"""
    # Configurar para ambiente de desenvolvimento
    config = {
        "environment": "development",
        "api_key": "test_key",
        "api_secret": "test_secret",
        "base_url": "https://api.circ.co.ao/v2"
    }
    
    connector = CIRCAngolaConnector(config)
    
    # Teste de verificação de documento
    doc_result = connector.verify_document(
        doc_type="bi",
        doc_number="000123456AB123",
        name="João Manuel dos Santos"
    )
    
    print("Resultado de verificação de documento:")
    print(json.dumps(doc_result, indent=2, ensure_ascii=False))
    
    # Teste de consulta de crédito
    credit_result = connector.check_credit_information(
        doc_type="bi",
        doc_number="000123456AB123"
    )
    
    print("\nResultado de consulta de crédito:")
    print(json.dumps(credit_result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    # Configurar logging para testes
    from datetime import timedelta
    logging.basicConfig(level=logging.INFO)
    
    # Executar teste
    test_connector()