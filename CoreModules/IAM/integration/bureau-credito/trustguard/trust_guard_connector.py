"""
Conector TrustGuard para integração com o sistema de regras dinâmicas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import asyncio
import json
import logging
import os
import time
from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Dict, List, Optional, Set, Tuple, Union
from uuid import uuid4

import httpx
from fastapi import Depends, FastAPI, HTTPException, Header, Security
from pydantic import BaseModel, Field, validator

# Importações de modelos do TrustGuard
from .trust_guard_models import (
    AccessDecision,
    AccessDecisionResponse,
    AccessRequest,
    AuthenticationLevel,
    AuthenticationRequest,
    AuthenticationResponse,
    AuthMethod,
    Policy,
    PolicyCondition,
    PolicyEffect,
    ResourceContext,
    RiskLevel,
    SessionInfo,
    SessionStatus,
    TrustScore,
    UserContext,
    UserRisk,
)

# Importações para integração com o sistema de regras dinâmicas
from infrastructure.fraud_detection.rules_engine.evaluator import RuleEvaluator
from infrastructure.fraud_detection.rules_engine.models import (
    Event,
    Rule,
    RuleEvaluationResult,
    RuleSet,
)
from infrastructure.fraud_detection.neuraflow.rule_enhancer import RuleEnhancer


class TrustGuardConfig(BaseModel):
    """Configuração do TrustGuard"""
    api_url: str = Field(..., description="URL da API do TrustGuard")
    api_key: str = Field(..., description="Chave de API do TrustGuard")
    tenant_id: str = Field(..., description="ID do tenant")
    timeout: int = Field(10, description="Timeout para requisições em segundos")
    cache_ttl: int = Field(300, description="Tempo de vida do cache em segundos")
    cache_enabled: bool = Field(True, description="Indica se o cache está habilitado")
    default_risk_level: RiskLevel = Field(RiskLevel.MEDIUM, 
                                        description="Nível de risco padrão")
    default_auth_level: AuthenticationLevel = Field(AuthenticationLevel.LOW, 
                                                 description="Nível de autenticação padrão")


class TrustGuardConnector:
    """
    Conector para integração com o TrustGuard para controle de acesso avançado.
    
    O TrustGuard é um módulo do INNOVABIZ que fornece:
    
    1. Autenticação adaptativa baseada em risco
    2. Autorização avançada baseada em políticas
    3. Avaliação de risco contextual
    4. Gestão de sessão segura
    5. Auditoria detalhada
    """
    
    def __init__(
        self,
        config: TrustGuardConfig,
        rule_evaluator: Optional[RuleEvaluator] = None,
        rule_enhancer: Optional[RuleEnhancer] = None,
        logger: Optional[logging.Logger] = None,
    ):
        """
        Inicializa o conector do TrustGuard.
        
        Args:
            config: Configuração do TrustGuard
            rule_evaluator: Avaliador de regras
            rule_enhancer: Enhancer de regras
            logger: Logger para registrar eventos
        """
        self.config = config
        self.rule_evaluator = rule_evaluator
        self.rule_enhancer = rule_enhancer
        self.logger = logger or logging.getLogger(__name__)
        self.client: Optional[httpx.AsyncClient] = None
        
        # Cache para decisões de acesso
        # Formato: {request_id} -> (timestamp, decision)
        self._cache: Dict[str, Tuple[float, AccessDecisionResponse]] = {}
        
        self.logger.info("TrustGuard connector initialized")
    
    async def __aenter__(self):
        await self.init_client()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
    
    async def init_client(self) -> httpx.AsyncClient:
        """
        Inicializa o cliente HTTP.
        
        Returns:
            httpx.AsyncClient: Cliente HTTP configurado
        """
        if self.client is None:
            self.client = httpx.AsyncClient(
                timeout=self.config.timeout,
                headers={
                    "X-API-Key": self.config.api_key,
                    "X-Tenant-ID": self.config.tenant_id,
                    "Content-Type": "application/json",
                },
                base_url=self.config.api_url,
            )
        
        return self.client
    
    async def close(self):
        """Fecha o cliente HTTP."""
        if self.client:
            await self.client.aclose()
            self.client = None
            self.logger.debug("HTTP client closed")
    
    async def _make_request(
        self,
        method: str,
        endpoint: str,
        json_data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        """
        Realiza uma requisição HTTP para o TrustGuard.
        
        Args:
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
        if self.client is None:
            await self.init_client()
        
        request_headers = headers or {}
        
        start_time = time.time()
        self.logger.debug(f"Making {method} request to TrustGuard - {endpoint}")
        
        try:
            response = await self.client.request(
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
                    detail=f"Erro na requisição ao TrustGuard: {detail}",
                )
            
            return response.json()
            
        except httpx.TimeoutException:
            self.logger.error(f"Request to TrustGuard timed out after {self.config.timeout}s")
            raise HTTPException(
                status_code=408,
                detail=f"Timeout na requisição ao TrustGuard após {self.config.timeout}s",
            )
            
        except httpx.RequestError as e:
            self.logger.error(f"Request error: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Erro na requisição ao TrustGuard: {str(e)}",
            )
    
    def _get_from_cache(
        self,
        request_id: str,
    ) -> Tuple[bool, Optional[AccessDecisionResponse]]:
        """
        Obtém uma decisão de acesso do cache.
        
        Args:
            request_id: ID da solicitação
            
        Returns:
            Tuple[bool, Optional[AccessDecisionResponse]]: (cache_hit, decision)
        """
        if not self.config.cache_enabled:
            return False, None
        
        cache_entry = self._cache.get(request_id)
        
        if not cache_entry:
            return False, None
        
        timestamp, decision = cache_entry
        now = time.time()
        
        if now - timestamp > self.config.cache_ttl:
            # Cache expirado
            del self._cache[request_id]
            return False, None
        
        return True, decision
    
    def _set_cache(
        self,
        request_id: str,
        decision: AccessDecisionResponse,
    ) -> None:
        """
        Armazena uma decisão de acesso no cache.
        
        Args:
            request_id: ID da solicitação
            decision: Decisão de acesso
        """
        if not self.config.cache_enabled:
            return
        
        self._cache[request_id] = (time.time(), decision)
    
    async def evaluate_access(
        self,
        request: AccessRequest,
    ) -> AccessDecisionResponse:
        """
        Avalia uma solicitação de acesso.
        
        Args:
            request: Solicitação de acesso
            
        Returns:
            AccessDecisionResponse: Resposta com a decisão de acesso
        """
        cache_hit, cached_decision = self._get_from_cache(request.request_id)
        
        if cache_hit:
            self.logger.debug(f"Cache hit for request {request.request_id}")
            return cached_decision
        
        try:
            result = await self._make_request(
                method="POST",
                endpoint="/access/evaluate",
                json_data=request.dict(),
            )
            
            decision = AccessDecisionResponse(**result)
            self._set_cache(request.request_id, decision)
            return decision
            
        except Exception as e:
            self.logger.error(f"Failed to evaluate access: {str(e)}")
            raise
    
    async def create_session(
        self,
        auth_response: AuthenticationResponse,
        session_ttl: int = 3600,  # 1 hora
    ) -> SessionInfo:
        """
        Cria uma nova sessão de usuário.
        
        Args:
            auth_response: Resposta de autenticação
            session_ttl: Tempo de vida da sessão em segundos
            
        Returns:
            SessionInfo: Informações da sessão criada
        """
        if not auth_response.success:
            raise ValueError("Cannot create session from failed authentication")
        
        try:
            result = await self._make_request(
                method="POST",
                endpoint="/sessions",
                json_data={
                    "user_id": auth_response.user_id,
                    "auth_level": auth_response.auth_level,
                    "ttl": session_ttl,
                    "auth_response_id": auth_response.request_id,
                },
            )
            
            return SessionInfo(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to create session: {str(e)}")
            raise
    
    async def validate_session(
        self,
        session_id: str,
    ) -> SessionInfo:
        """
        Valida uma sessão de usuário.
        
        Args:
            session_id: ID da sessão
            
        Returns:
            SessionInfo: Informações atualizadas da sessão
            
        Raises:
            HTTPException: Se a sessão for inválida
        """
        try:
            result = await self._make_request(
                method="GET",
                endpoint=f"/sessions/{session_id}",
            )
            
            session_info = SessionInfo(**result)
            
            if session_info.status != SessionStatus.ACTIVE:
                raise HTTPException(
                    status_code=401,
                    detail=f"Session is not active: {session_info.status}",
                )
            
            return session_info
            
        except HTTPException:
            raise
            
        except Exception as e:
            self.logger.error(f"Failed to validate session: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Failed to validate session: {str(e)}",
            )
    
    async def update_session(
        self,
        session_id: str,
        updates: Dict[str, Any],
    ) -> SessionInfo:
        """
        Atualiza uma sessão de usuário.
        
        Args:
            session_id: ID da sessão
            updates: Atualizações a serem aplicadas
            
        Returns:
            SessionInfo: Informações atualizadas da sessão
        """
        try:
            result = await self._make_request(
                method="PATCH",
                endpoint=f"/sessions/{session_id}",
                json_data=updates,
            )
            
            return SessionInfo(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to update session: {str(e)}")
            raise
    
    async def revoke_session(
        self,
        session_id: str,
        reason: Optional[str] = None,
    ) -> bool:
        """
        Revoga uma sessão de usuário.
        
        Args:
            session_id: ID da sessão
            reason: Motivo da revogação
            
        Returns:
            bool: True se a sessão foi revogada com sucesso
        """
        try:
            await self._make_request(
                method="DELETE",
                endpoint=f"/sessions/{session_id}",
                params={"reason": reason} if reason else None,
            )
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to revoke session: {str(e)}")
            return False
    
    async def authenticate(
        self,
        request: AuthenticationRequest,
    ) -> AuthenticationResponse:
        """
        Autentica um usuário.
        
        Args:
            request: Solicitação de autenticação
            
        Returns:
            AuthenticationResponse: Resposta da autenticação
        """
        try:
            result = await self._make_request(
                method="POST",
                endpoint="/auth/authenticate",
                json_data=request.dict(),
            )
            
            return AuthenticationResponse(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to authenticate: {str(e)}")
            raise
    
    async def get_user_risk(
        self,
        user_id: str,
    ) -> UserRisk:
        """
        Obtém o perfil de risco de um usuário.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            UserRisk: Perfil de risco do usuário
        """
        try:
            result = await self._make_request(
                method="GET",
                endpoint=f"/users/{user_id}/risk",
            )
            
            return UserRisk(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get user risk: {str(e)}")
            raise
    
    async def update_user_risk(
        self,
        user_id: str,
        risk_level: RiskLevel,
        risk_score: float,
        risk_factors: List[Dict[str, Any]] = None,
    ) -> UserRisk:
        """
        Atualiza o perfil de risco de um usuário.
        
        Args:
            user_id: ID do usuário
            risk_level: Nível de risco
            risk_score: Score de risco
            risk_factors: Fatores de risco
            
        Returns:
            UserRisk: Perfil de risco atualizado
        """
        risk_factors = risk_factors or []
        
        try:
            result = await self._make_request(
                method="PUT",
                endpoint=f"/users/{user_id}/risk",
                json_data={
                    "user_id": user_id,
                    "risk_level": risk_level,
                    "risk_score": risk_score,
                    "risk_factors": risk_factors,
                },
            )
            
            return UserRisk(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to update user risk: {str(e)}")
            raise
    
    async def create_policy(
        self,
        policy: Policy,
    ) -> Policy:
        """
        Cria uma nova política de acesso.
        
        Args:
            policy: Política a ser criada
            
        Returns:
            Policy: Política criada
        """
        try:
            result = await self._make_request(
                method="POST",
                endpoint="/policies",
                json_data=policy.dict(),
            )
            
            return Policy(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to create policy: {str(e)}")
            raise
    
    async def get_policy(
        self,
        policy_id: str,
    ) -> Policy:
        """
        Obtém uma política de acesso.
        
        Args:
            policy_id: ID da política
            
        Returns:
            Policy: Política encontrada
        """
        try:
            result = await self._make_request(
                method="GET",
                endpoint=f"/policies/{policy_id}",
            )
            
            return Policy(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to get policy: {str(e)}")
            raise
    
    async def update_policy(
        self,
        policy_id: str,
        updates: Dict[str, Any],
    ) -> Policy:
        """
        Atualiza uma política de acesso.
        
        Args:
            policy_id: ID da política
            updates: Atualizações a serem aplicadas
            
        Returns:
            Policy: Política atualizada
        """
        try:
            result = await self._make_request(
                method="PATCH",
                endpoint=f"/policies/{policy_id}",
                json_data=updates,
            )
            
            return Policy(**result)
            
        except Exception as e:
            self.logger.error(f"Failed to update policy: {str(e)}")
            raise
    
    async def delete_policy(
        self,
        policy_id: str,
    ) -> bool:
        """
        Exclui uma política de acesso.
        
        Args:
            policy_id: ID da política
            
        Returns:
            bool: True se a política foi excluída com sucesso
        """
        try:
            await self._make_request(
                method="DELETE",
                endpoint=f"/policies/{policy_id}",
            )
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to delete policy: {str(e)}")
            return False
    
    async def list_policies(
        self,
        active_only: bool = True,
        resource_type: Optional[str] = None,
        action: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[Policy]:
        """
        Lista políticas de acesso.
        
        Args:
            active_only: Listar apenas políticas ativas
            resource_type: Filtrar por tipo de recurso
            action: Filtrar por ação
            limit: Limite de resultados
            offset: Deslocamento para paginação
            
        Returns:
            List[Policy]: Lista de políticas
        """
        params = {
            "active_only": str(active_only).lower(),
            "limit": limit,
            "offset": offset,
        }
        
        if resource_type:
            params["resource_type"] = resource_type
            
        if action:
            params["action"] = action
        
        try:
            result = await self._make_request(
                method="GET",
                endpoint="/policies",
                params=params,
            )
            
            return [Policy(**item) for item in result.get("items", [])]
            
        except Exception as e:
            self.logger.error(f"Failed to list policies: {str(e)}")
            raise