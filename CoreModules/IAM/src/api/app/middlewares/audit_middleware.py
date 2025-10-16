import json
import time
import uuid
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id
from app.db.session import get_db_context
from sqlalchemy import text

class AuditMiddleware(BaseHTTPMiddleware):
    """
    Middleware para registrar eventos de auditoria para todas as requisições.
    Implementa requisitos de auditoria para conformidade com GDPR, LGPD, HIPAA e PNDSB.
    """
    
    async def dispatch(self, request: Request, call_next) -> Response:
        # Gera ID único para a requisição
        request_id = str(uuid.uuid4())
        request.state.request_id = request_id
        
        # Extrai informações da requisição
        method = request.method
        url = str(request.url)
        client_ip = request.client.host if request.client else None
        user_agent = request.headers.get("User-Agent", "")
        
        # Extrai o usuário autenticado do token JWT (se disponível)
        user_id = self._extract_user_id(request)
        
        # Registra o tempo de início da requisição
        start_time = time.time()
        
        # Prepara os dados de entrada para o log de auditoria
        request_body = None
        content_type = request.headers.get("Content-Type", "")
        if "application/json" in content_type and method in ["POST", "PUT", "PATCH"]:
            try:
                # Clona o corpo da requisição para poder lê-lo
                body_bytes = await request.body()
                request.state.body_bytes = body_bytes  # Armazena para uso posterior
                
                # Tenta fazer o parse do JSON (pode falhar se o formato for inválido)
                request_body = json.loads(body_bytes.decode())
                
                # Remove dados sensíveis para o log (ex: senhas)
                if isinstance(request_body, dict):
                    if "password" in request_body:
                        request_body["password"] = "********"
                    if "current_password" in request_body:
                        request_body["current_password"] = "********"
                    if "verification_code" in request_body:
                        request_body["verification_code"] = "********"
            except Exception as e:
                logger.warning(f"Could not parse request body as JSON: {str(e)}")
        
        response = None
        audit_event = {
            "request_id": request_id,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "user_id": user_id,
            "tenant_id": get_current_tenant_id(),
            "method": method,
            "url": url,
            "client_ip": client_ip,
            "user_agent": user_agent,
            "request_body": request_body if settings.DEBUG else None,  # Em produção, não registra o corpo
            "status_code": None,
            "response_time_ms": None,
            "error": None
        }
        
        try:
            # Executa o resto da aplicação
            response = await call_next(request)
            
            # Registra informações da resposta
            audit_event["status_code"] = response.status_code
            audit_event["response_time_ms"] = int((time.time() - start_time) * 1000)
            
            # Armazena o evento de auditoria
            await self._store_audit_event(audit_event)
            
            return response
        except Exception as e:
            # Registra a exceção
            error_detail = str(e)
            audit_event["error"] = error_detail
            audit_event["status_code"] = 500
            audit_event["response_time_ms"] = int((time.time() - start_time) * 1000)
            
            # Armazena o evento de auditoria mesmo em caso de erro
            await self._store_audit_event(audit_event)
            
            # Re-lança a exceção para ser tratada por outros handlers
            raise
    
    def _extract_user_id(self, request: Request) -> str:
        """
        Extrai o ID do usuário do token JWT ou de outros mecanismos de autenticação.
        Este é um placeholder simplificado - a implementação real depende do seu mecanismo de autenticação.
        """
        # Se você tiver um middleware de autenticação que já extrai essa informação,
        # pode recuperá-la de request.state
        if hasattr(request.state, "user_id"):
            return request.state.user_id
        
        # Implementação simplificada - em uma implementação real,
        # você precisa decodificar o token JWT e extrair o user_id
        authorization = request.headers.get("Authorization", "")
        if authorization.startswith("Bearer "):
            # Em uma implementação real, decodifique o token e extraia o user_id
            # Por enquanto, retorna None para indicar usuário não autenticado
            return None
        
        return None
    
    async def _store_audit_event(self, audit_event: dict) -> None:
        """
        Armazena o evento de auditoria no banco de dados.
        Na implementação real, isso seria feito de forma assíncrona para não bloquear a resposta.
        """
        if not settings.AUDIT_ENABLED:
            # Se a auditoria estiver desabilitada, apenas registra em log se estiver em modo debug
            if settings.DEBUG:
                logger.debug(f"Audit event: {json.dumps(audit_event)}")
            return
        
        try:
            # Nesta implementação, chamamos diretamente a função SQL que criamos
            # para registrar eventos de auditoria
            with get_db_context() as db:
                event_type = "api_request"
                resource_type = "api_endpoint"
                action = audit_event["method"]
                resource_id = audit_event["url"]
                actor_id = audit_event["user_id"] or "anonymous"
                metadata = json.dumps({
                    "request_id": audit_event["request_id"],
                    "client_ip": audit_event["client_ip"],
                    "user_agent": audit_event["user_agent"],
                    "status_code": audit_event["status_code"],
                    "response_time_ms": audit_event["response_time_ms"],
                    "error": audit_event["error"]
                })
                
                # Chamada para a função de log de auditoria que implementamos no banco de dados
                sql = text("""
                    SELECT iam.log_audit_event(
                        :tenant_id, :event_type, :resource_type, :resource_id,
                        :action, :actor_id, :metadata
                    )
                """)
                
                db.execute(
                    sql,
                    {
                        "tenant_id": audit_event["tenant_id"],
                        "event_type": event_type,
                        "resource_type": resource_type,
                        "resource_id": resource_id,
                        "action": action,
                        "actor_id": actor_id,
                        "metadata": metadata
                    }
                )
                db.commit()
                
        except Exception as e:
            # Em caso de erro ao registrar a auditoria, registra um log, mas não falha a requisição
            logger.error(f"Error storing audit event: {str(e)}")
            # Em produção, poderia enviar para uma fila para processamento posterior
