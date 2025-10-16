"""
Cliente HTTP para API do NeuraFlow.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import time
from typing import Any, Dict, List, Optional, Union

import httpx
from pydantic import ValidationError

from .models import (
    EnhancementConfig,
    EnhancedEventData,
    ModelMetadata,
    ModelType,
    NeuraFlowDetectionRequest,
    NeuraFlowDetectionResponse,
    NeuraFlowModelInfo
)


class NeuraFlowAPIError(Exception):
    """Exceção para erros da API NeuraFlow."""

    def __init__(self, status_code: int, detail: str, trace_id: Optional[str] = None):
        self.status_code = status_code
        self.detail = detail
        self.trace_id = trace_id
        super().__init__(f"NeuraFlow API Error ({status_code}): {detail}")


class NeuraFlowClient:
    """Cliente para comunicação com a API NeuraFlow."""

    def __init__(
        self,
        api_key: str,
        base_url: str = "https://api.neuraflow.innovabiz.io/v1",
        timeout: int = 10,
        logger: Optional[logging.Logger] = None,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ):
        """
        Inicializa o cliente NeuraFlow.
        
        Args:
            api_key: Chave de API para autenticação
            base_url: URL base da API NeuraFlow
            timeout: Tempo limite para requisições em segundos
            logger: Logger para registrar eventos
            tenant_id: ID do tenant padrão
            region: Região padrão
        """
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.logger = logger or logging.getLogger(__name__)
        self.tenant_id = tenant_id
        self.region = region
        
        self.client = httpx.AsyncClient(
            timeout=timeout,
            headers={
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/json",
                "X-NeuraFlow-SDK-Version": "1.0.0",
            },
        )
        
        self.logger.info(f"NeuraFlow client initialized with base URL: {base_url}")
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
    
    async def close(self):
        """Fecha o cliente HTTP."""
        if self.client:
            await self.client.aclose()
            self.logger.debug("NeuraFlow client closed")
    
    async def _make_request(
        self,
        method: str,
        endpoint: str,
        json_data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        """
        Realiza uma requisição HTTP à API NeuraFlow.
        
        Args:
            method: Método HTTP (GET, POST, etc.)
            endpoint: Endpoint da API
            json_data: Dados JSON para o corpo da requisição
            params: Parâmetros de consulta
            headers: Cabeçalhos adicionais
            
        Returns:
            Dict[str, Any]: Resposta da API
            
        Raises:
            NeuraFlowAPIError: Em caso de erro na requisição
        """
        url = f"{self.base_url}/{endpoint}"
        request_headers = {
            "X-NeuraFlow-Request-ID": f"nf-req-{int(time.time() * 1000)}",
        }
        
        if headers:
            request_headers.update(headers)
        
        if self.tenant_id:
            request_headers["X-NeuraFlow-Tenant-ID"] = self.tenant_id
            
        if self.region:
            request_headers["X-NeuraFlow-Region"] = self.region
        
        start_time = time.time()
        self.logger.debug(f"Making {method} request to {url}")
        
        try:
            response = await self.client.request(
                method=method,
                url=url,
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
                    trace_id = error_data.get("trace_id")
                except Exception:
                    detail = response.text or "Unknown error"
                    trace_id = None
                
                raise NeuraFlowAPIError(response.status_code, detail, trace_id)
            
            return response.json()
            
        except httpx.TimeoutException:
            self.logger.error(f"Request to {url} timed out after {self.timeout}s")
            raise NeuraFlowAPIError(408, f"Request timed out after {self.timeout}s")
            
        except httpx.RequestError as e:
            self.logger.error(f"Request error: {str(e)}")
            raise NeuraFlowAPIError(0, f"Request error: {str(e)}")
    
    async def detect_anomalies(
        self,
        event_data: Dict[str, Any],
        model_id: Optional[str] = None,
        model_type: Optional[ModelType] = None,
        region: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None,
        confidence_threshold: Optional[float] = None,
        tenant_id: Optional[str] = None,
    ) -> NeuraFlowDetectionResponse:
        """
        Detecta anomalias nos dados do evento usando o NeuraFlow.
        
        Args:
            event_data: Dados do evento para análise
            model_id: ID do modelo específico a ser usado
            model_type: Tipo de modelo a ser usado
            region: Região para contexto regional
            context: Dados de contexto adicionais
            confidence_threshold: Limite de confiança para detecção
            tenant_id: ID do tenant (sobrescreve o padrão)
            
        Returns:
            NeuraFlowDetectionResponse: Resultado da detecção
        """
        request = NeuraFlowDetectionRequest(
            event_data=event_data,
            model_id=model_id,
            model_type=model_type.value if model_type else None,
            region=region or self.region,
            context=context,
            tenant_id=tenant_id or self.tenant_id,
            confidence_threshold=confidence_threshold,
        )
        
        result = await self._make_request("POST", "detect", json_data=request.dict(exclude_none=True))
        
        try:
            return NeuraFlowDetectionResponse(**result)
        except ValidationError as e:
            self.logger.error(f"Failed to parse NeuraFlow response: {e}")
            raise NeuraFlowAPIError(0, f"Invalid response format: {e}")
    
    async def enhance_event_data(
        self,
        event_data: Dict[str, Any],
        config: EnhancementConfig,
        tenant_id: Optional[str] = None,
        region: Optional[str] = None,
    ) -> EnhancedEventData:
        """
        Aprimora dados de evento com análises adicionais.
        
        Args:
            event_data: Dados do evento para aprimoramento
            config: Configuração de aprimoramento
            tenant_id: ID do tenant (sobrescreve o padrão)
            region: Região (sobrescreve o padrão)
            
        Returns:
            EnhancedEventData: Dados aprimorados do evento
        """
        request_data = {
            "event_data": event_data,
            "enhancement_config": config.dict(exclude_none=True),
            "tenant_id": tenant_id or self.tenant_id,
            "region": region or self.region,
        }
        
        result = await self._make_request("POST", "enhance", json_data=request_data)
        
        try:
            return EnhancedEventData(**result)
        except ValidationError as e:
            self.logger.error(f"Failed to parse enhancement response: {e}")
            raise NeuraFlowAPIError(0, f"Invalid response format: {e}")
    
    async def get_available_models(
        self,
        model_type: Optional[ModelType] = None,
        region: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> NeuraFlowModelInfo:
        """
        Obtém modelos disponíveis no NeuraFlow.
        
        Args:
            model_type: Filtrar por tipo de modelo
            region: Filtrar por região
            tenant_id: ID do tenant (sobrescreve o padrão)
            
        Returns:
            NeuraFlowModelInfo: Informações sobre modelos disponíveis
        """
        params = {}
        if model_type:
            params["model_type"] = model_type.value
            
        if region:
            params["region"] = region
            
        headers = {}
        if tenant_id:
            headers["X-NeuraFlow-Tenant-ID"] = tenant_id
        
        result = await self._make_request("GET", "models", params=params, headers=headers)
        
        try:
            return NeuraFlowModelInfo(**result)
        except ValidationError as e:
            self.logger.error(f"Failed to parse models response: {e}")
            raise NeuraFlowAPIError(0, f"Invalid response format: {e}")
    
    async def get_model_details(
        self,
        model_id: str,
        tenant_id: Optional[str] = None,
    ) -> ModelMetadata:
        """
        Obtém detalhes de um modelo específico.
        
        Args:
            model_id: ID do modelo
            tenant_id: ID do tenant (sobrescreve o padrão)
            
        Returns:
            ModelMetadata: Metadados do modelo
        """
        headers = {}
        if tenant_id:
            headers["X-NeuraFlow-Tenant-ID"] = tenant_id
            
        result = await self._make_request("GET", f"models/{model_id}", headers=headers)
        
        try:
            return ModelMetadata(**result)
        except ValidationError as e:
            self.logger.error(f"Failed to parse model details response: {e}")
            raise NeuraFlowAPIError(0, f"Invalid response format: {e}")