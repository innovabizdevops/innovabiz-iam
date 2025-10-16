import uuid
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
from app.core.config import settings
from app.core.logger import logger
from contextvars import ContextVar

# Variáveis de contexto para o ID do tenant
tenant_id_var: ContextVar[str] = ContextVar('tenant_id', default=settings.DEFAULT_TENANT_ID)

class TenantMiddleware(BaseHTTPMiddleware):
    """
    Middleware para gerenciar o contexto multi-tenant nas requisições.
    Extrai o ID do tenant dos headers, parâmetros da requisição ou token JWT.
    Define o ID no contexto da requisição e também no contexto de execução.
    """
    
    async def dispatch(self, request: Request, call_next) -> Response:
        tenant_id = self._get_tenant_id(request)
        # Define o tenant_id no contexto
        request.state.tenant_id = tenant_id
        token = tenant_id_var.set(tenant_id)
        
        try:
            # Executa o resto da aplicação com o tenant_id no contexto
            response = await call_next(request)
            return response
        except Exception as e:
            logger.error(f"Error processing request in tenant middleware: {str(e)}")
            raise
        finally:
            # Restaura o contexto anterior (importante para evitar vazamento entre requisições)
            tenant_id_var.reset(token)

    def _get_tenant_id(self, request: Request) -> str:
        """
        Extrai o ID do tenant da requisição. Prioridade:
        1. Cabeçalho X-Tenant-ID
        2. Parâmetro de query tenant_id
        3. JWT Token (authorization header com propriedade tenant_id)
        4. Tenant padrão
        """
        # Verifica se existe um tenant_id no cabeçalho
        tenant_id = request.headers.get('X-Tenant-ID')
        if tenant_id:
            try:
                # Valida UUID
                return str(uuid.UUID(tenant_id))
            except ValueError:
                logger.warning(f"Invalid tenant_id in header: {tenant_id}")
        
        # Verifica se existe um tenant_id nos parâmetros de query
        tenant_id = request.query_params.get('tenant_id')
        if tenant_id:
            try:
                # Valida UUID
                return str(uuid.UUID(tenant_id))
            except ValueError:
                logger.warning(f"Invalid tenant_id in query parameters: {tenant_id}")
        
        # Seria necessário implementar a extração do JWT token e obter o tenant_id
        # Essa implementação seria mais complexa e dependeria da lógica de autenticação
        # Por enquanto, retornamos o tenant padrão
        
        return settings.DEFAULT_TENANT_ID

def get_current_tenant_id() -> str:
    """
    Retorna o ID do tenant atual no contexto de execução.
    Pode ser usado em qualquer parte da aplicação para obter o tenant atual.
    """
    return tenant_id_var.get()
