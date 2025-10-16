"""
Serviço para integração com Bureau de Créditos.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Tuple, Union

import httpx
from fastapi import Depends, HTTPException
from pydantic import BaseModel, Field, validator

from ..models import (
    CreditReport, 
    CreditScore, 
    CreditRisk,
    FinancialProfile, 
    TransactionHistory,
    FraudIndicator,
    IdentityVerification
)


class BureauCredentialsType(str, Enum):
    """Tipos de credenciais para Bureau de Créditos"""
    API_KEY = "api_key"
    OAUTH = "oauth"
    CERTIFICATE = "certificate"
    JWT = "jwt"


class BureauProvider(str, Enum):
    """Provedores de Bureau de Créditos suportados"""
    SERASA = "serasa"
    SPC = "spc"
    QUOD = "quod"
    TRANSUNION = "transunion"
    EXPERIAN = "experian"
    EQUIFAX = "equifax"
    CREDLINK = "credlink"
    TRUSTGUARD = "trustguard"  # Provedor interno INNOVABIZ


class BureauCredentials(BaseModel):
    """Credenciais para autenticação com Bureau de Créditos"""
    type: BureauCredentialsType
    api_key: Optional[str] = None
    client_id: Optional[str] = None
    client_secret: Optional[str] = None
    certificate_path: Optional[str] = None
    jwt_token: Optional[str] = None
    
    @validator("api_key")
    def validate_api_key(cls, v, values):
        if values.get("type") == BureauCredentialsType.API_KEY and not v:
            raise ValueError("API key é obrigatório para o tipo de credencial API_KEY")
        return v
    
    @validator("client_id", "client_secret")
    def validate_oauth(cls, v, values):
        if values.get("type") == BureauCredentialsType.OAUTH and not v:
            raise ValueError("client_id e client_secret são obrigatórios para o tipo de credencial OAUTH")
        return v
    
    @validator("certificate_path")
    def validate_certificate(cls, v, values):
        if values.get("type") == BureauCredentialsType.CERTIFICATE and not v:
            raise ValueError("certificate_path é obrigatório para o tipo de credencial CERTIFICATE")
        return v
    
    @validator("jwt_token")
    def validate_jwt(cls, v, values):
        if values.get("type") == BureauCredentialsType.JWT and not v:
            raise ValueError("jwt_token é obrigatório para o tipo de credencial JWT")
        return v


class BureauCredentialsManager:
    """Gerenciador de credenciais para Bureau de Créditos"""
    
    def __init__(
        self,
        credentials_path: Optional[str] = None,
        credentials: Optional[Dict[str, BureauCredentials]] = None,
        logger: Optional[logging.Logger] = None,
    ):
        """
        Inicializa o gerenciador de credenciais.
        
        Args:
            credentials_path: Caminho para arquivo de credenciais
            credentials: Dicionário de credenciais
            logger: Logger para registrar eventos
        """
        self.logger = logger or logging.getLogger(__name__)
        self.credentials = credentials or {}
        
        if credentials_path:
            self.load_credentials(credentials_path)
    
    def load_credentials(self, credentials_path: str) -> None:
        """Carrega credenciais de arquivo."""
        try:
            with open(credentials_path, "r") as f:
                data = json.load(f)
                
                for provider, creds in data.items():
                    self.credentials[provider] = BureauCredentials(**creds)
                    
            self.logger.info(f"Loaded credentials for {len(self.credentials)} bureau providers")
                
        except Exception as e:
            self.logger.error(f"Failed to load bureau credentials: {str(e)}")
    
    def get_credentials(self, provider: str) -> Optional[BureauCredentials]:
        """
        Obtém credenciais para um provedor específico.
        
        Args:
            provider: Nome do provedor
            
        Returns:
            Optional[BureauCredentials]: Credenciais do provedor ou None
        """
        return self.credentials.get(provider)
    
    def set_credentials(self, provider: str, credentials: BureauCredentials) -> None:
        """
        Define credenciais para um provedor específico.
        
        Args:
            provider: Nome do provedor
            credentials: Credenciais para o provedor
        """
        self.credentials[provider] = credentials
        self.logger.info(f"Set credentials for bureau provider '{provider}'")


class BureauCreditoService:
    """
    Serviço para integração com Bureau de Créditos.
    
    Esta classe fornece métodos para interagir com diferentes provedores
    de Bureau de Créditos, permitindo:
    
    1. Verificar score de crédito
    2. Obter relatório de crédito completo
    3. Verificar risco de crédito
    4. Validar identidade
    5. Verificar indicadores de fraude
    6. Obter histórico de transações
    """
    
    def __init__(
        self,
        credentials_manager: BureauCredentialsManager,
        logger: Optional[logging.Logger] = None,
        timeout: int = 10,
    ):
        """
        Inicializa o serviço de Bureau de Créditos.
        
        Args:
            credentials_manager: Gerenciador de credenciais
            logger: Logger para registrar eventos
            timeout: Tempo limite para requisições em segundos
        """
        self.credentials_manager = credentials_manager
        self.logger = logger or logging.getLogger(__name__)
        self.timeout = timeout
        self.clients: Dict[str, httpx.AsyncClient] = {}
        
        self.base_urls = {
            BureauProvider.SERASA: "https://api.serasa.com.br/v1",
            BureauProvider.SPC: "https://api.spcbrasil.org/v1",
            BureauProvider.QUOD: "https://api.quod.com.br/v1",
            BureauProvider.TRANSUNION: "https://api.transunion.com/v1",
            BureauProvider.EXPERIAN: "https://api.experian.com/v1",
            BureauProvider.EQUIFAX: "https://api.equifax.com/v1",
            BureauProvider.CREDLINK: "https://api.credlink.com.br/v1",
            BureauProvider.TRUSTGUARD: "http://trustguard-api.innovabiz.io/v1",
        }
        
        self.logger.info("Bureau de Crédito service initialized")
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
    
    async def close(self):
        """Fecha todos os clientes HTTP."""
        for provider, client in self.clients.items():
            await client.aclose()
            self.logger.debug(f"Closed HTTP client for {provider}")
        
        self.clients = {}
    
    async def get_client(self, provider: str) -> httpx.AsyncClient:
        """
        Obtém cliente HTTP para um provedor específico.
        
        Args:
            provider: Nome do provedor
            
        Returns:
            httpx.AsyncClient: Cliente HTTP configurado
            
        Raises:
            ValueError: Se as credenciais do provedor não estiverem disponíveis
        """
        if provider in self.clients:
            return self.clients[provider]
        
        credentials = self.credentials_manager.get_credentials(provider)
        
        if not credentials:
            raise ValueError(f"Credenciais não disponíveis para o provedor '{provider}'")
        
        headers = {}
        
        if credentials.type == BureauCredentialsType.API_KEY:
            headers["X-API-Key"] = credentials.api_key
            
        elif credentials.type == BureauCredentialsType.JWT:
            headers["Authorization"] = f"Bearer {credentials.jwt_token}"
        
        client = httpx.AsyncClient(
            timeout=self.timeout,
            headers=headers,
            base_url=self.base_urls.get(provider, ""),
        )
        
        self.clients[provider] = client
        return client
    
    async def _make_request(
        self,
        provider: str,
        method: str,
        endpoint: str,
        json_data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        """
        Realiza uma requisição HTTP para o provedor de Bureau de Créditos.
        
        Args:
            provider: Nome do provedor
            method: Método HTTP (GET, POST, etc.)
            endpoint: Endpoint da API
            json_data: Dados JSON para o corpo da requisição
            params: Parâmetros de consulta
            headers: Cabeçalhos adicionais
            
        Returns:
            Dict[str, Any]: Resposta da API
            
        Raises:
            HTTPException: Em caso de erro na requisição
        """
        client = await self.get_client(provider)
        request_headers = headers or {}
        
        start_time = time.time()
        self.logger.debug(f"Making {method} request to {provider} - {endpoint}")
        
        try:
            response = await client.request(
                method=method,
                url=endpoint,
                json=json_data,
                params=params,
                headers=request_headers,
            )
            
            elapsed = (time.time() - start_time) * 1000
            self.logger.debug(f"Request completed in {elapsed:.2f}ms with status {response.status_code}")
            
            if response.status_code >= 400:
                try:
                    error_data = response.json()
                    detail = error_data.get("detail", "Unknown error")
                except Exception:
                    detail = response.text or "Unknown error"
                
                raise HTTPException(
                    status_code=response.status_code,
                    detail=f"Erro na requisição ao {provider}: {detail}",
                )
            
            return response.json()
            
        except httpx.TimeoutException:
            self.logger.error(f"Request to {provider} timed out after {self.timeout}s")
            raise HTTPException(
                status_code=408,
                detail=f"Timeout na requisição ao {provider} após {self.timeout}s",
            )
            
        except httpx.RequestError as e:
            self.logger.error(f"Request error: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Erro na requisição ao {provider}: {str(e)}",
            )
    
    async def get_credit_score(
        self,
        document_id: str,
        provider: str,
        region: Optional[str] = None,
    ) -> CreditScore:
        """
        Obtém o score de crédito de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            region: Região para contexto regional
            
        Returns:
            CreditScore: Informações de score de crédito
        """
        try:
            params = {"document": document_id}
            
            if region:
                params["region"] = region
            
            result = await self._make_request(
                provider=provider,
                method="GET",
                endpoint="/credit/score",
                params=params,
            )
            
            return CreditScore(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get credit score: {str(e)}")
            raise
    
    async def get_credit_report(
        self,
        document_id: str,
        provider: str,
        include_details: bool = False,
        region: Optional[str] = None,
    ) -> CreditReport:
        """
        Obtém o relatório de crédito completo de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            include_details: Indica se deve incluir detalhes completos
            region: Região para contexto regional
            
        Returns:
            CreditReport: Relatório de crédito completo
        """
        try:
            params = {"document": document_id, "include_details": include_details}
            
            if region:
                params["region"] = region
            
            result = await self._make_request(
                provider=provider,
                method="GET",
                endpoint="/credit/report",
                params=params,
            )
            
            return CreditReport(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get credit report: {str(e)}")
            raise
    
    async def get_credit_risk(
        self,
        document_id: str,
        provider: str,
        transaction_amount: Optional[float] = None,
        transaction_type: Optional[str] = None,
        region: Optional[str] = None,
    ) -> CreditRisk:
        """
        Avalia o risco de crédito de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            transaction_amount: Valor da transação para avaliação contextual
            transaction_type: Tipo da transação para avaliação contextual
            region: Região para contexto regional
            
        Returns:
            CreditRisk: Avaliação de risco de crédito
        """
        try:
            data = {
                "document": document_id,
                "transaction_amount": transaction_amount,
                "transaction_type": transaction_type,
                "region": region,
            }
            
            result = await self._make_request(
                provider=provider,
                method="POST",
                endpoint="/credit/risk",
                json_data={k: v for k, v in data.items() if v is not None},
            )
            
            return CreditRisk(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get credit risk: {str(e)}")
            raise
    
    async def verify_identity(
        self,
        document_id: str,
        provider: str,
        name: str,
        birth_date: Optional[str] = None,
        mother_name: Optional[str] = None,
        phone_number: Optional[str] = None,
        region: Optional[str] = None,
    ) -> IdentityVerification:
        """
        Verifica a identidade de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            name: Nome completo
            birth_date: Data de nascimento (formato YYYY-MM-DD)
            mother_name: Nome da mãe
            phone_number: Número de telefone
            region: Região para contexto regional
            
        Returns:
            IdentityVerification: Resultado da verificação de identidade
        """
        try:
            data = {
                "document": document_id,
                "name": name,
                "birth_date": birth_date,
                "mother_name": mother_name,
                "phone_number": phone_number,
                "region": region,
            }
            
            result = await self._make_request(
                provider=provider,
                method="POST",
                endpoint="/identity/verify",
                json_data={k: v for k, v in data.items() if v is not None},
            )
            
            return IdentityVerification(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to verify identity: {str(e)}")
            raise
    
    async def check_fraud_indicators(
        self,
        document_id: str,
        provider: str,
        ip_address: Optional[str] = None,
        device_id: Optional[str] = None,
        user_agent: Optional[str] = None,
        region: Optional[str] = None,
    ) -> FraudIndicator:
        """
        Verifica indicadores de fraude para um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            ip_address: Endereço IP do cliente
            device_id: ID do dispositivo
            user_agent: User-Agent do navegador
            region: Região para contexto regional
            
        Returns:
            FraudIndicator: Indicadores de fraude
        """
        try:
            data = {
                "document": document_id,
                "ip_address": ip_address,
                "device_id": device_id,
                "user_agent": user_agent,
                "region": region,
            }
            
            result = await self._make_request(
                provider=provider,
                method="POST",
                endpoint="/fraud/indicators",
                json_data={k: v for k, v in data.items() if v is not None},
            )
            
            return FraudIndicator(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to check fraud indicators: {str(e)}")
            raise
    
    async def get_transaction_history(
        self,
        document_id: str,
        provider: str,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        limit: int = 50,
        region: Optional[str] = None,
    ) -> TransactionHistory:
        """
        Obtém o histórico de transações de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            start_date: Data inicial (formato YYYY-MM-DD)
            end_date: Data final (formato YYYY-MM-DD)
            limit: Limite de transações
            region: Região para contexto regional
            
        Returns:
            TransactionHistory: Histórico de transações
        """
        try:
            params = {
                "document": document_id,
                "limit": limit,
            }
            
            if start_date:
                params["start_date"] = start_date
                
            if end_date:
                params["end_date"] = end_date
                
            if region:
                params["region"] = region
            
            result = await self._make_request(
                provider=provider,
                method="GET",
                endpoint="/transactions/history",
                params=params,
            )
            
            return TransactionHistory(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get transaction history: {str(e)}")
            raise
    
    async def get_financial_profile(
        self,
        document_id: str,
        provider: str,
        region: Optional[str] = None,
    ) -> FinancialProfile:
        """
        Obtém o perfil financeiro de um cliente.
        
        Args:
            document_id: Número do documento (CPF/CNPJ)
            provider: Nome do provedor de Bureau de Créditos
            region: Região para contexto regional
            
        Returns:
            FinancialProfile: Perfil financeiro do cliente
        """
        try:
            params = {"document": document_id}
            
            if region:
                params["region"] = region
            
            result = await self._make_request(
                provider=provider,
                method="GET",
                endpoint="/financial/profile",
                params=params,
            )
            
            return FinancialProfile(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get financial profile: {str(e)}")
            raise


# Factory para criação do serviço
async def get_bureau_credito_service() -> BureauCreditoService:
    """
    Factory para criação do serviço de Bureau de Créditos.
    
    Returns:
        BureauCreditoService: Serviço configurado
    """
    logger = logging.getLogger("bureau_credito_service")
    
    # Caminho para o arquivo de credenciais
    credentials_path = os.environ.get(
        "BUREAU_CREDENTIALS_PATH",
        "/etc/innovabiz/bureau/credentials.json",
    )
    
    # Cria gerenciador de credenciais
    credentials_manager = BureauCredentialsManager(
        credentials_path=credentials_path,
        logger=logger,
    )
    
    # Cria serviço
    service = BureauCreditoService(
        credentials_manager=credentials_manager,
        logger=logger,
        timeout=int(os.environ.get("BUREAU_TIMEOUT", "10")),
    )
    
    try:
        yield service
    finally:
        await service.close()