"""
TrustGuard - Conector com Observabilidade Integrada

Conector para o serviço TrustGuard com observabilidade completa
multi-camada para monitoramento avançado.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - TrustGuard com Observabilidade
Data: 20/08/2025
"""

import os
import json
import time
import asyncio
import httpx
from typing import Dict, Any, Optional, List, Union, Tuple
from dataclasses import dataclass
from enum import Enum

# Importar sistema de observabilidade
from observability.core import (
    create_observer,
    trace_request,
    trace_risk,
    EventContext,
    EventSeverity,
    EventCategory
)


class TrustGuardRiskLevel(str, Enum):
    """Níveis de risco retornados pelo TrustGuard."""
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"
    UNKNOWN = "UNKNOWN"


@dataclass
class TrustGuardCredentials:
    """Credenciais para autenticação no TrustGuard."""
    api_key: str
    tenant_id: str
    client_id: Optional[str] = None
    market_context: Optional[str] = None
    region: Optional[str] = None


@dataclass
class TrustGuardAssessmentRequest:
    """Solicitação de avaliação para o TrustGuard."""
    entity_id: str
    event_type: str
    context_data: Dict[str, Any]
    transaction_id: Optional[str] = None
    amount: Optional[float] = None
    currency: Optional[str] = None
    market: Optional[str] = None
    channel: Optional[str] = None
    ip_address: Optional[str] = None
    device_id: Optional[str] = None


@dataclass
class TrustGuardAssessmentResponse:
    """Resposta de avaliação do TrustGuard."""
    transaction_id: str
    entity_id: str
    risk_level: TrustGuardRiskLevel
    risk_score: float
    reason_codes: List[str]
    recommendation: str
    additional_checks: List[str]
    evaluation_time_ms: int
    request_id: str
    timestamp: str
    raw_response: Dict[str, Any]
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'TrustGuardAssessmentResponse':
        """
        Cria uma resposta a partir de um dicionário.
        
        Args:
            data: Dicionário com dados da resposta
            
        Returns:
            TrustGuardAssessmentResponse: Objeto de resposta
        """
        try:
            risk_level_str = data.get('risk_level', TrustGuardRiskLevel.UNKNOWN)
            risk_level = TrustGuardRiskLevel(risk_level_str) if risk_level_str in [e.value for e in TrustGuardRiskLevel] else TrustGuardRiskLevel.UNKNOWN
            
            return cls(
                transaction_id=data.get('transaction_id', ''),
                entity_id=data.get('entity_id', ''),
                risk_level=risk_level,
                risk_score=float(data.get('risk_score', 0.0)),
                reason_codes=data.get('reason_codes', []),
                recommendation=data.get('recommendation', ''),
                additional_checks=data.get('additional_checks', []),
                evaluation_time_ms=int(data.get('evaluation_time_ms', 0)),
                request_id=data.get('request_id', ''),
                timestamp=data.get('timestamp', ''),
                raw_response=data
            )
        except (ValueError, TypeError) as e:
            # Se houver erro na conversão, criar resposta de erro
            return cls(
                transaction_id=data.get('transaction_id', ''),
                entity_id=data.get('entity_id', ''),
                risk_level=TrustGuardRiskLevel.UNKNOWN,
                risk_score=0.0,
                reason_codes=[f"Error parsing response: {str(e)}"],
                recommendation="ERROR",
                additional_checks=[],
                evaluation_time_ms=0,
                request_id=data.get('request_id', ''),
                timestamp=data.get('timestamp', ''),
                raw_response=data
            )


class TrustGuardObservabilityConnector:
    """
    Conector para o TrustGuard com observabilidade integrada.
    
    Fornece métodos para avaliar riscos e verificar entidades com
    monitoramento completo de logs, métricas e tracing em múltiplas camadas.
    """
    
    def __init__(
        self, 
        credentials: TrustGuardCredentials,
        base_url: Optional[str] = None,
        timeout_seconds: float = 10.0,
        cache_ttl_seconds: int = 300,
        retry_attempts: int = 3,
        retry_backoff_seconds: float = 1.0,
        disable_observability: bool = False
    ):
        """
        Inicializa o conector TrustGuard.
        
        Args:
            credentials: Credenciais para autenticação
            base_url: URL base da API do TrustGuard
            timeout_seconds: Tempo limite para requisições em segundos
            cache_ttl_seconds: Tempo de vida do cache em segundos
            retry_attempts: Número de tentativas de retry
            retry_backoff_seconds: Tempo de espera entre retentativas
            disable_observability: Desativa a observabilidade para testes
        """
        self.credentials = credentials
        self.base_url = base_url or os.environ.get('TRUST_GUARD_BASE_URL', 'https://api.trustguard.innovabiz.com/v1')
        self.timeout_seconds = timeout_seconds
        self.cache_ttl_seconds = cache_ttl_seconds
        self.retry_attempts = retry_attempts
        self.retry_backoff_seconds = retry_backoff_seconds
        
        # Cache para respostas recentes
        self._cache: Dict[str, Tuple[TrustGuardAssessmentResponse, float]] = {}
        
        # Configurar observabilidade
        if not disable_observability:
            self.observer = create_observer(
                module_name="innovabiz.trustguard",
                component_name="connector"
            )
        else:
            self.observer = None
        
        self.http_client = httpx.AsyncClient(
            timeout=timeout_seconds,
            headers=self._build_auth_headers()
        )
    
    def _build_auth_headers(self) -> Dict[str, str]:
        """
        Constrói os headers de autenticação.
        
        Returns:
            Dict[str, str]: Headers para requisições autenticadas
        """
        headers = {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "X-API-Key": self.credentials.api_key,
            "X-Tenant-ID": self.credentials.tenant_id
        }
        
        if self.credentials.client_id:
            headers["X-Client-ID"] = self.credentials.client_id
        
        if self.credentials.market_context:
            headers["X-Market-Context"] = self.credentials.market_context
        
        if self.credentials.region:
            headers["X-Region"] = self.credentials.region
        
        return headers
    
    def _get_cache_key(self, request: TrustGuardAssessmentRequest) -> str:
        """
        Gera uma chave de cache para a requisição.
        
        Args:
            request: Requisição de avaliação
            
        Returns:
            str: Chave de cache
        """
        # Chave baseada na combinação de identificadores únicos
        key_parts = [
            request.entity_id,
            request.event_type,
            request.transaction_id or "",
            str(request.amount or 0),
            request.currency or "",
            request.market or ""
        ]
        
        return ":".join(key_parts)
    
    def _get_from_cache(self, request: TrustGuardAssessmentRequest) -> Optional[TrustGuardAssessmentResponse]:
        """
        Tenta recuperar uma resposta do cache.
        
        Args:
            request: Requisição de avaliação
            
        Returns:
            Optional[TrustGuardAssessmentResponse]: Resposta em cache ou None
        """
        cache_key = self._get_cache_key(request)
        
        if cache_key in self._cache:
            cached_response, timestamp = self._cache[cache_key]
            current_time = time.time()
            
            # Verificar se o cache ainda é válido
            if current_time - timestamp < self.cache_ttl_seconds:
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Cache hit para avaliação de entidade {request.entity_id}",
                        severity=EventSeverity.DEBUG,
                        category=EventCategory.INFRASTRUCTURE,
                        metadata={"entity_id": request.entity_id, "event_type": request.event_type}
                    )
                return cached_response
            
            # Remover entradas expiradas
            del self._cache[cache_key]
        
        return None
    
    def _store_in_cache(self, request: TrustGuardAssessmentRequest, response: TrustGuardAssessmentResponse) -> None:
        """
        Armazena uma resposta no cache.
        
        Args:
            request: Requisição original
            response: Resposta a ser armazenada
        """
        cache_key = self._get_cache_key(request)
        self._cache[cache_key] = (response, time.time())
        
        # Log de debug para cache
        if self.observer:
            self.observer.base_monitor.log_event(
                message=f"Armazenando resposta em cache para entidade {request.entity_id}",
                severity=EventSeverity.DEBUG,
                category=EventCategory.INFRASTRUCTURE,
                metadata={"entity_id": request.entity_id, "event_type": request.event_type}
            )
        
        # Limpar cache se estiver muito grande (limite arbitrário de 1000 itens)
        if len(self._cache) > 1000:
            # Remover os itens mais antigos
            sorted_keys = sorted(
                self._cache.keys(),
                key=lambda k: self._cache[k][1]
            )
            
            # Remover os 20% mais antigos
            items_to_remove = int(len(self._cache) * 0.2)
            for key in sorted_keys[:items_to_remove]:
                del self._cache[key]
    
    @trace_risk(risk_type="entity")
    async def assess_risk(self, request: TrustGuardAssessmentRequest) -> TrustGuardAssessmentResponse:
        """
        Avalia o risco de uma entidade ou transação.
        
        Args:
            request: Dados para avaliação de risco
            
        Returns:
            TrustGuardAssessmentResponse: Resultado da avaliação
            
        Raises:
            Exception: Se houver erro na comunicação ou resposta inválida
        """
        # Verificar cache primeiro se disponível
        cached_response = self._get_from_cache(request)
        if cached_response:
            return cached_response
        
        # Criar contexto para evento com dados relevantes
        event_context = EventContext(
            user_id=request.entity_id,
            transaction_id=request.transaction_id,
            tenant_id=self.credentials.tenant_id,
            attributes={
                "event_type": request.event_type,
                "market": request.market or self.credentials.market_context or "unknown"
            }
        )
        
        # Logs e métricas de entrada
        if self.observer:
            self.observer.base_monitor.log_event(
                message=f"Iniciando avaliação de risco para entidade {request.entity_id} - tipo {request.event_type}",
                context=event_context,
                severity=EventSeverity.INFO,
                category=EventCategory.SECURITY
            )
        
        # Preparar payload
        payload = {
            "entity_id": request.entity_id,
            "event_type": request.event_type,
            "context_data": request.context_data,
        }
        
        # Adicionar campos opcionais se presentes
        for field in ["transaction_id", "amount", "currency", "market", "channel", "ip_address", "device_id"]:
            value = getattr(request, field)
            if value is not None:
                payload[field] = value
        
        # Se não há transaction_id, gerar um
        if not payload.get("transaction_id"):
            payload["transaction_id"] = f"tx-{int(time.time())}-{request.entity_id[:8]}"
        
        # Realizar requisição com retry
        attempts = 0
        last_error = None
        
        while attempts < self.retry_attempts:
            try:
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Enviando requisição para TrustGuard (tentativa {attempts+1}/{self.retry_attempts})",
                        context=event_context,
                        severity=EventSeverity.DEBUG,
                        category=EventCategory.INFRASTRUCTURE
                    )
                
                # Fazer a requisição
                url = f"{self.base_url}/assessment"
                response = await self.http_client.post(url, json=payload)
                
                # Verificar resposta HTTP
                if response.status_code == 200:
                    response_data = response.json()
                    assessment_response = TrustGuardAssessmentResponse.from_dict(response_data)
                    
                    # Registrar métricas de tempo de avaliação
                    eval_time_ms = assessment_response.evaluation_time_ms
                    risk_score = assessment_response.risk_score
                    risk_level = assessment_response.risk_level
                    
                    # Armazenar em cache
                    self._store_in_cache(request, assessment_response)
                    
                    # Log de resultado
                    if self.observer:
                        self.observer.base_monitor.log_event(
                            message=f"Avaliação concluída: entidade={request.entity_id}, risco={risk_level.value}, pontuação={risk_score}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.SECURITY,
                            metadata={
                                "risk_level": risk_level.value,
                                "risk_score": risk_score,
                                "evaluation_time_ms": eval_time_ms,
                                "reason_codes": assessment_response.reason_codes
                            }
                        )
                    
                    return assessment_response
                
                # Tratar erros HTTP
                error_msg = f"Erro HTTP {response.status_code}: {response.text}"
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Erro na avaliação: {error_msg}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.INFRASTRUCTURE
                    )
                
                last_error = Exception(error_msg)
                
                # Decidir se faz retry com base no status code
                if response.status_code >= 500 or response.status_code == 429:
                    await asyncio.sleep(self.retry_backoff_seconds * (2 ** attempts))
                    attempts += 1
                    continue
                else:
                    # Erros 4xx não devem ser retentados (exceto 429)
                    break
                
            except Exception as e:
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Exceção na comunicação com TrustGuard: {str(e)}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.INFRASTRUCTURE,
                        exception=e
                    )
                
                last_error = e
                await asyncio.sleep(self.retry_backoff_seconds * (2 ** attempts))
                attempts += 1
        
        # Se chegou aqui, todas as tentativas falharam
        if last_error:
            raise last_error
        
        # Caso improvável de chegar aqui sem erro definido
        raise Exception("Falha na avaliação de risco - causa desconhecida")
    
    @trace_risk(risk_type="verification")
    async def verify_entity(
        self, 
        entity_id: str,
        verification_type: str,
        verification_data: Dict[str, Any],
        market: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Verifica atributos de uma entidade contra o bureau de créditos.
        
        Args:
            entity_id: Identificador da entidade
            verification_type: Tipo de verificação (identity, address, etc)
            verification_data: Dados para verificação
            market: Mercado alvo (opcional)
            
        Returns:
            Dict[str, Any]: Resultado da verificação
            
        Raises:
            Exception: Se houver erro na comunicação ou resposta inválida
        """
        # Criar contexto para evento
        event_context = EventContext(
            user_id=entity_id,
            tenant_id=self.credentials.tenant_id,
            attributes={
                "verification_type": verification_type,
                "market": market or self.credentials.market_context or "unknown"
            }
        )
        
        # Logs e métricas de entrada
        if self.observer:
            self.observer.base_monitor.log_event(
                message=f"Iniciando verificação de entidade {entity_id} - tipo {verification_type}",
                context=event_context,
                severity=EventSeverity.INFO,
                category=EventCategory.SECURITY
            )
        
        # Preparar payload
        payload = {
            "entity_id": entity_id,
            "verification_type": verification_type,
            "verification_data": verification_data
        }
        
        if market:
            payload["market"] = market
        
        # Realizar requisição com retry
        attempts = 0
        last_error = None
        
        while attempts < self.retry_attempts:
            try:
                # Fazer a requisição
                url = f"{self.base_url}/verify"
                response = await self.http_client.post(url, json=payload)
                
                # Verificar resposta HTTP
                if response.status_code == 200:
                    verification_result = response.json()
                    
                    # Log de resultado
                    if self.observer:
                        self.observer.base_monitor.log_event(
                            message=f"Verificação concluída: entidade={entity_id}, tipo={verification_type}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.SECURITY,
                            metadata={
                                "verification_success": verification_result.get("success", False),
                                "verification_score": verification_result.get("score", 0),
                                "verification_type": verification_type
                            }
                        )
                    
                    return verification_result
                
                # Tratar erros HTTP
                error_msg = f"Erro HTTP {response.status_code}: {response.text}"
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Erro na verificação: {error_msg}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.INFRASTRUCTURE
                    )
                
                last_error = Exception(error_msg)
                
                # Decisão de retry
                if response.status_code >= 500 or response.status_code == 429:
                    await asyncio.sleep(self.retry_backoff_seconds * (2 ** attempts))
                    attempts += 1
                    continue
                else:
                    break
                
            except Exception as e:
                if self.observer:
                    self.observer.base_monitor.log_event(
                        message=f"Exceção na verificação com TrustGuard: {str(e)}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.INFRASTRUCTURE,
                        exception=e
                    )
                
                last_error = e
                await asyncio.sleep(self.retry_backoff_seconds * (2 ** attempts))
                attempts += 1
        
        # Se chegou aqui, todas as tentativas falharam
        if last_error:
            raise last_error
        
        # Caso improvável de chegar aqui sem erro definido
        raise Exception("Falha na verificação de entidade - causa desconhecida")
    
    async def close(self):
        """Fecha o cliente HTTP e libera recursos."""
        await self.http_client.aclose()
        
        if self.observer:
            self.observer.base_monitor.log_event(
                message="Conector TrustGuard fechado",
                severity=EventSeverity.INFO,
                category=EventCategory.INFRASTRUCTURE
            )
    
    async def __aenter__(self):
        """Suporte para uso em contexto assíncrono."""
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Fechamento automático ao sair do contexto."""
        await self.close()