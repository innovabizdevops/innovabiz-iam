"""
INNOVABIZ IAM - Middleware de Auditoria Multi-Contexto
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Middleware para captura automática e enriquecimento de eventos de auditoria
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA
"""

import time
import uuid
import ipaddress
import structlog
import json
from typing import Callable, Dict, List, Optional, Any, Set
from fastapi import Request, Response
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.types import ASGIApp, Receive, Scope, Send
from starlette.datastructures import MutableHeaders

from app.models.audit import (
    AuditEventCategory, 
    AuditEventSeverity, 
    AuditEventCreate, 
    AuditHttpDetails,
    RegionalCompliance,
    ComplianceFramework
)

# Configuração do logger estruturado
logger = structlog.get_logger(__name__)

# Constantes para configuração
DEFAULT_TENANT_HEADER = "X-Tenant-ID"
DEFAULT_REGIONAL_CONTEXT_HEADER = "X-Regional-Context"
DEFAULT_LANG_HEADER = "Accept-Language"
DEFAULT_CORRELATION_ID_HEADER = "X-Correlation-ID"

# Paths que não devem ser auditados
DEFAULT_EXCLUDED_PATHS = {
    "/health", 
    "/ready", 
    "/live", 
    "/metrics", 
    "/docs", 
    "/redoc", 
    "/openapi.json"
}

# Headers sensíveis que devem ser removidos ou mascarados
SENSITIVE_HEADERS = {
    "authorization", 
    "cookie", 
    "x-api-key", 
    "proxy-authorization"
}

# Configurações de compliance por região
REGION_COMPLIANCE_CONFIG = {
    "BR": {
        "frameworks": [
            ComplianceFramework.LGPD,
            ComplianceFramework.BACEN,
            ComplianceFramework.PCI_DSS
        ],
        "data_retention": 730,  # 2 anos
        "required_fields": ["user_id", "tenant_id", "regional_context"]
    },
    "AO": {
        "frameworks": [
            ComplianceFramework.BNA,
            ComplianceFramework.PCI_DSS
        ],
        "data_retention": 1825,  # 5 anos
        "required_fields": ["user_id", "tenant_id", "regional_context"]
    },
    "EU": {
        "frameworks": [
            ComplianceFramework.GDPR,
            ComplianceFramework.PSD2,
            ComplianceFramework.PCI_DSS
        ],
        "data_retention": 365,  # 1 ano
        "required_fields": ["user_id", "tenant_id", "regional_context"]
    },
    "US": {
        "frameworks": [
            ComplianceFramework.PCI_DSS,
            ComplianceFramework.SOX
        ],
        "data_retention": 2555,  # 7 anos
        "required_fields": ["user_id", "tenant_id", "regional_context"]
    },
    # Default para qualquer outra região
    "DEFAULT": {
        "frameworks": [ComplianceFramework.PCI_DSS],
        "data_retention": 365,  # 1 ano
        "required_fields": ["user_id", "tenant_id"]
    }
}


class AuditMiddleware(BaseHTTPMiddleware):
    """
    Middleware de auditoria para FastAPI que captura automaticamente eventos HTTP.
    
    Funcionalidades:
    - Captura automática de eventos HTTP
    - Integração com contexto regional e tenant
    - Aplicação de políticas de compliance baseadas em região
    - Geração de IDs de correlação para rastreamento
    - Enriquecimento de eventos com informações de contexto
    - Mascaramento de dados sensíveis
    - Logging estruturado
    """
    
    def __init__(
        self, 
        app: ASGIApp, 
        audit_service: Any = None,
        exclude_paths: Optional[Set[str]] = None,
        tenant_header: str = DEFAULT_TENANT_HEADER,
        regional_context_header: str = DEFAULT_REGIONAL_CONTEXT_HEADER,
        lang_header: str = DEFAULT_LANG_HEADER,
        correlation_id_header: str = DEFAULT_CORRELATION_ID_HEADER,
        record_request_body: bool = False,
        record_response_body: bool = False,
        enable_audit_log: bool = True,
        default_tenant: str = None,
        default_regional_context: str = None
    ):
        """
        Inicializa o middleware de auditoria.
        
        Args:
            app: A aplicação ASGI
            audit_service: Serviço de auditoria para registro de eventos
            exclude_paths: Conjunto de paths que não devem ser auditados
            tenant_header: Nome do header para identificação do tenant
            regional_context_header: Nome do header para contexto regional
            lang_header: Nome do header para idioma
            correlation_id_header: Nome do header para ID de correlação
            record_request_body: Se deve registrar o corpo da requisição (use com cautela)
            record_response_body: Se deve registrar o corpo da resposta (use com cautela)
            enable_audit_log: Se o registro de auditoria está habilitado
            default_tenant: Tenant padrão se não especificado no header
            default_regional_context: Contexto regional padrão se não especificado no header
        """
        super().__init__(app)
        self.audit_service = audit_service
        self.exclude_paths = exclude_paths or DEFAULT_EXCLUDED_PATHS
        self.tenant_header = tenant_header
        self.regional_context_header = regional_context_header
        self.lang_header = lang_header
        self.correlation_id_header = correlation_id_header
        self.record_request_body = record_request_body
        self.record_response_body = record_response_body
        self.enable_audit_log = enable_audit_log
        self.default_tenant = default_tenant
        self.default_regional_context = default_regional_context
        
        logger.info(
            "AuditMiddleware inicializado",
            exclude_paths=self.exclude_paths,
            tenant_header=self.tenant_header,
            regional_context_header=self.regional_context_header,
            record_request_body=self.record_request_body,
            record_response_body=self.record_response_body,
            enable_audit_log=self.enable_audit_log
        )
    
    async def dispatch(
        self, request: Request, call_next: Callable
    ) -> Response:
        """
        Processa uma requisição e captura informações para auditoria.
        
        Args:
            request: A requisição HTTP
            call_next: Callback para processar a próxima middleware/handler
            
        Returns:
            Response: A resposta HTTP
        """
        # Não auditar caminhos excluídos
        if self._should_skip_audit(request.url.path):
            return await call_next(request)
        
        # Início da requisição
        start_time = time.time()
        
        # Extrair contexto e preparar dados
        correlation_id = self._get_or_generate_correlation_id(request)
        tenant_id = self._get_tenant_id(request)
        regional_context = self._get_regional_context(request)
        language = self._get_language(request)
        
        # Injetar correlation_id no header de response
        async def send_wrapper(message: Dict[str, Any]):
            if message["type"] == "http.response.start":
                headers = MutableHeaders(scope=message)
                headers.append(self.correlation_id_header, correlation_id)
            await send(message)
        
        # Processando a requisição e capturando a resposta
        response = None
        error = None
        status_code = 500
        
        try:
            # Passar correlation_id, tenant_id e regional_context como state
            request.state.correlation_id = correlation_id
            request.state.tenant_id = tenant_id
            request.state.regional_context = regional_context
            request.state.language = language
            
            # Processar a requisição
            response = await call_next(request)
            status_code = response.status_code
            return response
            
        except Exception as exc:
            error = exc
            logger.exception(
                "Erro ao processar requisição",
                correlation_id=correlation_id,
                tenant_id=tenant_id,
                regional_context=regional_context,
                path=request.url.path,
                method=request.method,
                error=str(exc)
            )
            # Recriar resposta em caso de erro
            if isinstance(exc, Response):
                response = exc
                status_code = response.status_code
            else:
                response = JSONResponse(
                    status_code=500,
                    content={"detail": "Internal server error"}
                )
                status_code = 500
            
            raise
            
        finally:
            # Registrar evento de auditoria (mesmo em caso de erro)
            if self.enable_audit_log and self.audit_service:
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Capturar headers (removendo informações sensíveis)
                headers = self._sanitize_headers(dict(request.headers.items()))
                
                # Capturar informações HTTP
                http_details = AuditHttpDetails(
                    method=request.method,
                    path=request.url.path,
                    query_params=dict(request.query_params),
                    headers=headers,
                    status_code=status_code,
                    request_id=correlation_id,
                    client_ip=self._get_client_ip(request),
                    user_agent=headers.get("user-agent"),
                    duration_ms=duration_ms
                )
                
                # Determinar categoria baseada no path
                category = self._determine_category(request.url.path)
                
                # Determinar severidade baseada no status code
                severity = self._determine_severity(status_code)
                
                # Extrair user_id (se disponível)
                user_id = await self._extract_user_id(request)
                
                # Gerar compliance info baseado no contexto regional
                compliance = self._generate_compliance_info(regional_context)
                
                # Criar evento de auditoria
                audit_event = AuditEventCreate(
                    category=category,
                    severity=severity,
                    action=f"http.{request.method.lower()}",
                    user_id=user_id,
                    resource_type="http_endpoint",
                    resource_id=request.url.path,
                    description=f"{request.method} {request.url.path}",
                    success=200 <= status_code < 400,
                    error_message=str(error) if error else None,
                    tenant_id=tenant_id,
                    regional_context=regional_context,
                    language=language,
                    correlation_id=correlation_id,
                    http_details=http_details,
                    source_ip=self._get_client_ip(request),
                    source_system="api",
                    compliance=compliance,
                    compliance_tags=self._determine_compliance_tags(request.url.path)
                )
                
                # Registrar evento de auditoria de forma assíncrona
                try:
                    await self.audit_service.create_event(audit_event)
                except Exception as e:
                    logger.error(
                        "Erro ao registrar evento de auditoria",
                        error=str(e),
                        correlation_id=correlation_id
                    )
            
            # Log final (mesmo sem serviço de auditoria)
            logger.info(
                "Requisição processada",
                method=request.method,
                path=request.url.path,
                status_code=status_code,
                duration_ms=int((time.time() - start_time) * 1000),
                correlation_id=correlation_id,
                tenant_id=tenant_id,
                regional_context=regional_context
            )
            
            return response
    
    def _should_skip_audit(self, path: str) -> bool:
        """Verifica se um path deve ser excluído da auditoria."""
        for excluded_path in self.exclude_paths:
            if path == excluded_path or path.startswith(excluded_path):
                return True
        return False
    
    def _get_or_generate_correlation_id(self, request: Request) -> str:
        """Obtém ou gera um ID de correlação para a requisição."""
        correlation_id = request.headers.get(self.correlation_id_header)
        if not correlation_id:
            correlation_id = str(uuid.uuid4())
        return correlation_id
    
    def _get_tenant_id(self, request: Request) -> Optional[str]:
        """Obtém o ID do tenant da requisição."""
        tenant_id = request.headers.get(self.tenant_header)
        return tenant_id or self.default_tenant
    
    def _get_regional_context(self, request: Request) -> str:
        """Obtém o contexto regional da requisição."""
        regional_context = request.headers.get(self.regional_context_header)
        return regional_context or self.default_regional_context or "DEFAULT"
    
    def _get_language(self, request: Request) -> Optional[str]:
        """Obtém o idioma da requisição."""
        lang_header = request.headers.get(self.lang_header)
        if lang_header:
            # Extrair código de idioma principal (ex: pt-BR;q=0.9 -> pt-BR)
            return lang_header.split(',')[0].split(';')[0].strip()
        return None
    
    def _get_client_ip(self, request: Request) -> Optional[str]:
        """
        Obtém o IP do cliente, considerando headers de proxy.
        Se necessário, aplica técnicas de anonimização para GDPR/LGPD.
        """
        # Tenta obter de headers comuns para proxies
        client_ip = (
            request.headers.get("x-forwarded-for") or
            request.headers.get("x-real-ip") or
            request.client.host if request.client else None
        )
        
        # Se tiver múltiplos IPs (x-forwarded-for pode ter vários), pega o primeiro
        if client_ip and "," in client_ip:
            client_ip = client_ip.split(",")[0].strip()
            
        # Anonimizar se necessário (depende da política de privacidade)
        regional_context = self._get_regional_context(request)
        if regional_context in ["EU", "BR"] and client_ip:
            try:
                # Para IPv4, mantém os primeiros 3 octetos
                if ":" not in client_ip:  # IPv4
                    parts = client_ip.split(".")
                    if len(parts) == 4:
                        client_ip = f"{parts[0]}.{parts[1]}.{parts[2]}.*"
                # Para IPv6, mantém os primeiros 4 grupos
                else:
                    addr = ipaddress.IPv6Address(client_ip)
                    masked = addr.exploded.split(":")
                    if len(masked) == 8:
                        client_ip = f"{masked[0]}:{masked[1]}:{masked[2]}:{masked[3]}:*:*:*:*"
            except Exception:
                # Em caso de erro na anonimização, retorna parcialmente mascarado
                client_ip = client_ip[:6] + "***"
                
        return client_ip
        
    async def _extract_user_id(self, request: Request) -> Optional[str]:
        """
        Extrai o ID do usuário da requisição.
        Tenta diferentes fontes como token JWT, sessão, etc.
        """
        # Verificar se já existe no state (definido por outro middleware)
        if hasattr(request.state, "user_id"):
            return request.state.user_id
            
        # Tentar extrair de token JWT no header Authorization
        auth_header = request.headers.get("authorization")
        if auth_header and auth_header.lower().startswith("bearer "):
            try:
                # Implementação simplificada - em produção, usar biblioteca JWT
                import base64
                import json
                
                token = auth_header.split(" ")[1]
                # Pegar a parte do payload (segundo segmento)
                payload_part = token.split(".")[1]
                # Adicionar padding se necessário
                padding = "=" * (4 - len(payload_part) % 4)
                payload_bytes = base64.urlsafe_b64decode(payload_part + padding)
                payload = json.loads(payload_bytes)
                # Tentar extrair user_id de claims comuns
                user_id = payload.get("sub") or payload.get("user_id")
                if user_id:
                    return str(user_id)
            except Exception as e:
                logger.warning("Erro ao extrair user_id do token JWT", error=str(e))
        
        # Outras implementações específicas podem ser adicionadas
        # Ex: cookies de sessão, basic auth, etc.
        
        return None
    
    def _sanitize_headers(self, headers: Dict[str, str]) -> Dict[str, str]:
        """Remove ou mascara headers sensíveis."""
        result = {}
        for key, value in headers.items():
            key_lower = key.lower()
            if key_lower in SENSITIVE_HEADERS:
                # Mascara headers sensíveis (ex: Authorization: Bearer eyJ... -> Bearer ****)
                if key_lower == "authorization":
                    parts = value.split(" ", 1)
                    if len(parts) > 1:
                        result[key] = f"{parts[0]} ****"
                    else:
                        result[key] = "****"
                else:
                    result[key] = "****"
            else:
                result[key] = value
        return result
    
    def _determine_category(self, path: str) -> AuditEventCategory:
        """Determina a categoria do evento baseada no path."""
        path_lower = path.lower()
        
        if "/auth" in path_lower or "/login" in path_lower or "/logout" in path_lower:
            return AuditEventCategory.AUTHENTICATION
        elif "/user" in path_lower or "/users" in path_lower:
            return AuditEventCategory.USER_MANAGEMENT
        elif "/role" in path_lower or "/permission" in path_lower:
            return AuditEventCategory.AUTHORIZATION
        elif "/config" in path_lower or "/settings" in path_lower:
            return AuditEventCategory.CONFIGURATION
        elif "/audit" in path_lower:
            return AuditEventCategory.SECURITY
        elif "/payment" in path_lower:
            return AuditEventCategory.PAYMENT
        elif "/transaction" in path_lower:
            return AuditEventCategory.TRANSACTION
        elif "/card" in path_lower:
            return AuditEventCategory.CARD_DATA
        elif "/consent" in path_lower or "/privacy" in path_lower:
            return AuditEventCategory.PRIVACY
        elif "/risk" in path_lower:
            return AuditEventCategory.RISK
        elif path_lower.startswith("/api/"):
            return AuditEventCategory.API
        else:
            return AuditEventCategory.APPLICATION
    
    def _determine_severity(self, status_code: int) -> AuditEventSeverity:
        """Determina a severidade baseada no código de status."""
        if status_code >= 500:
            return AuditEventSeverity.HIGH
        elif status_code >= 400:
            return AuditEventSeverity.MEDIUM
        elif status_code >= 300:
            return AuditEventSeverity.LOW
        else:
            return AuditEventSeverity.INFO
    
    def _generate_compliance_info(self, regional_context: str) -> RegionalCompliance:
        """
        Gera informações de compliance baseadas no contexto regional.
        """
        # Obter configuração da região ou usar default
        config = REGION_COMPLIANCE_CONFIG.get(
            regional_context, 
            REGION_COMPLIANCE_CONFIG["DEFAULT"]
        )
        
        return RegionalCompliance(
            frameworks=config.get("frameworks", []),
            data_residency=regional_context,
            data_retention=config.get("data_retention"),
            required_fields=config.get("required_fields", []),
            sensitive_fields=config.get("sensitive_fields", [])
        )
    
    def _determine_compliance_tags(self, path: str) -> List[str]:
        """
        Determina tags de compliance baseadas no path e no tipo de dados acessados.
        """
        path_lower = path.lower()
        tags = []
        
        # Adicionar tags baseadas no conteúdo do path
        if "/user" in path_lower or "/profile" in path_lower:
            tags.append("PII")  # Personally Identifiable Information
        
        if "/payment" in path_lower or "/card" in path_lower:
            tags.append("PCI")  # Payment Card Industry
        
        if "/health" in path_lower or "/medical" in path_lower:
            tags.append("PHI")  # Protected Health Information
        
        if "/consent" in path_lower or "/privacy" in path_lower:
            tags.append("CONSENT")  # Consentimento
            
        if "/document" in path_lower or "/kyc" in path_lower:
            tags.append("KYC")  # Know Your Customer
            
        return tags


class AuditContextMiddleware(BaseHTTPMiddleware):
    """
    Middleware para gerenciar contexto regional e tenant para auditoria.
    
    Este middleware adiciona informações de contexto ao estado da requisição,
    permitindo que handlers e outros middlewares acessem esses dados.
    """
    
    def __init__(
        self,
        app: ASGIApp,
        tenant_header: str = DEFAULT_TENANT_HEADER,
        regional_context_header: str = DEFAULT_REGIONAL_CONTEXT_HEADER,
        lang_header: str = DEFAULT_LANG_HEADER,
        correlation_id_header: str = DEFAULT_CORRELATION_ID_HEADER,
        default_tenant: str = None,
        default_regional_context: str = None,
        default_language: str = "en-US"
    ):
        """
        Inicializa o middleware de contexto de auditoria.
        
        Args:
            app: A aplicação ASGI
            tenant_header: Nome do header para identificação do tenant
            regional_context_header: Nome do header para contexto regional
            lang_header: Nome do header para idioma
            correlation_id_header: Nome do header para ID de correlação
            default_tenant: Tenant padrão se não especificado no header
            default_regional_context: Contexto regional padrão se não especificado no header
            default_language: Idioma padrão se não especificado no header
        """
        super().__init__(app)
        self.tenant_header = tenant_header
        self.regional_context_header = regional_context_header
        self.lang_header = lang_header
        self.correlation_id_header = correlation_id_header
        self.default_tenant = default_tenant
        self.default_regional_context = default_regional_context
        self.default_language = default_language
        
    async def dispatch(
        self, request: Request, call_next: Callable
    ) -> Response:
        """
        Processa uma requisição e adiciona informações de contexto ao estado.
        
        Args:
            request: A requisição HTTP
            call_next: Callback para processar a próxima middleware/handler
            
        Returns:
            Response: A resposta HTTP
        """
        # Extrair ou gerar correlation_id
        correlation_id = request.headers.get(self.correlation_id_header)
        if not correlation_id:
            correlation_id = str(uuid.uuid4())
            
        # Extrair tenant_id
        tenant_id = request.headers.get(self.tenant_header) or self.default_tenant
            
        # Extrair regional_context
        regional_context = request.headers.get(self.regional_context_header) or self.default_regional_context
            
        # Extrair language
        language_header = request.headers.get(self.lang_header)
        language = self.default_language
        if language_header:
            # Extrair código de idioma principal (ex: pt-BR;q=0.9 -> pt-BR)
            language = language_header.split(',')[0].split(';')[0].strip()
            
        # Adicionar informações ao estado da requisição
        request.state.correlation_id = correlation_id
        request.state.tenant_id = tenant_id
        request.state.regional_context = regional_context
        request.state.language = language
        
        # Adicionar correlation_id ao header de resposta
        response = await call_next(request)
        response.headers[self.correlation_id_header] = correlation_id
        
        # Se houver tenant_id e regional_context, adicionar aos headers de resposta
        if tenant_id:
            response.headers[self.tenant_header] = tenant_id
        if regional_context:
            response.headers[self.regional_context_header] = regional_context
            
        return response