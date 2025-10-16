"""
Integração de Observabilidade para o Serviço de Auditoria IAM.

Este módulo implementa a integração das métricas Prometheus e outras
ferramentas de observabilidade com a aplicação principal FastAPI.
Segue os padrões INNOVABIZ e melhores práticas internacionais como
OpenTelemetry, Prometheus e Grafana.

Referências:
- ISO 20000 (Gestão de Serviços de TI)
- ITIL v4 (Observabilidade e Monitoramento)
- OpenTelemetry Specification v1.4
- Prometheus Operator Best Practices
- Gartner Recommendations for Observability 2025
"""
import os
import time
import logging
import structlog
from typing import Dict, Any, Optional
from fastapi import FastAPI, Request, Response
from prometheus_client import start_http_server, REGISTRY, PROCESS_COLLECTOR, PLATFORM_COLLECTOR
import socket

# Importar métricas e utilidades de instrumentação
from ..metrics import (
    init_metrics,
    setup_service_info,
    update_service_health,
    register_retention_policies,
    start_uptime_counter
)

# Configurar logger estruturado
logger = structlog.get_logger(__name__)


class ObservabilityIntegration:
    """
    Classe para integração de observabilidade com a aplicação principal.
    Implementa os padrões Multi-Contexto e Multi-Tenant da plataforma INNOVABIZ.
    """

    def __init__(
        self, 
        app: FastAPI,
        service_name: str = "iam-audit",
        version: str = "1.0.0",
        metrics_path: str = "/metrics",
        health_check_path: str = "/health",
    ):
        """
        Inicializa a integração de observabilidade.
        
        Args:
            app: Aplicação FastAPI principal
            service_name: Nome do serviço
            version: Versão do serviço
            metrics_path: Caminho para o endpoint de métricas
            health_check_path: Caminho para o health check
        """
        self.app = app
        self.service_name = service_name
        self.version = version
        self.metrics_path = metrics_path
        self.health_check_path = health_check_path
        self.start_time = time.time()
        
        # Armazenar informações de ambiente
        self.environment = os.getenv("ENVIRONMENT", "development")
        self.region = os.getenv("REGION", "global")
        self.build_id = os.getenv("BUILD_ID", "local")
        self.commit_hash = os.getenv("COMMIT_HASH", "dev")
        
        logger.info(
            "Inicializando integração de observabilidade",
            service=self.service_name,
            version=self.version,
            environment=self.environment
        )
        def setup(self):
        """
        Configura todos os componentes de observabilidade para a aplicação.
        
        Implementa a configuração completa do sistema de observabilidade:
        - Métricas Prometheus
        - Health Check
        - Informações do serviço
        - Contador de uptime
        - Endpoints de diagnóstico
        
        Segue os padrões INNOVABIZ para observabilidade multi-contexto.
        """
        # Inicializar métricas Prometheus no aplicativo FastAPI
        init_metrics(self.app)
        
        # Configurar informações estáticas do serviço
        setup_service_info(
            version=self.version,
            build_id=self.build_id,
            commit_hash=self.commit_hash,
            environment=self.environment,
            region=self.region
        )
        
        # Iniciar contador de tempo de atividade
        start_uptime_counter()
        
        # Registrar componentes do serviço como saudáveis por padrão
        self._setup_initial_health_state()
        
        # Adicionar endpoint de health check
        self._setup_health_endpoint()
        
        # Adicionar endpoint de diagnóstico para KrakenD API Gateway
        self._setup_diagnostic_endpoint()
        
        # Registrar handlers para início e término da aplicação
        self._setup_lifecycle_handlers()
        
        logger.info(
            "Integração de observabilidade configurada com sucesso",
            service=self.service_name,
            metrics_path=self.metrics_path,
            health_check_path=self.health_check_path
        )
        
        return self
        
    def _setup_initial_health_state(self):
        """Configura o estado inicial de saúde dos componentes."""
        # Componentes padrão do serviço de auditoria
        components = [
            "database", 
            "redis_cache", 
            "message_queue", 
            "storage_service",
            "krakend_gateway"
        ]
        
        # Definir todos os componentes como operacionais
        for component in components:
            update_service_health(component, 1)  # 1 = operacional
            
        logger.debug(
            "Estado inicial de saúde dos componentes configurado",
            components=components
        )    def _setup_health_endpoint(self):
        """
        Configura o endpoint de health check conforme padrões INNOVABIZ.
        Segue o Modelo de Capacidades de Observabilidade da plataforma.
        """
        @self.app.get(self.health_check_path)
        async def health_check(request: Request) -> Dict[str, Any]:
            """
            Endpoint de health check que verifica o estado dos componentes críticos.
            Implementa as recomendações do Gartner para Health Checks de Microserviços.
            
            Retorna:
                dict: Status de saúde do serviço e seus componentes
            """
            # Extrair informações de contexto multi-tenant e multi-regional dos headers
            tenant_id = request.headers.get("X-Tenant-ID", "global")
            regional_context = request.headers.get("X-Regional-Context", "global")
            
            # Registrar acesso ao health check com contexto
            logger.debug(
                "Health check acessado",
                tenant_id=tenant_id,
                regional_context=regional_context,
                client_ip=request.client.host if request.client else None
            )
            
            # Verificar componentes em tempo real (poderia consultar serviços externos)
            database_ok = self._check_database_health()
            cache_ok = self._check_cache_health()
            queue_ok = self._check_queue_health()
            storage_ok = self._check_storage_health()
            
            # Atualizar métricas de saúde com base nas verificações
            update_service_health("database", 1 if database_ok else 0)
            update_service_health("redis_cache", 1 if cache_ok else 0)
            update_service_health("message_queue", 1 if queue_ok else 0)
            update_service_health("storage_service", 1 if storage_ok else 0)
            
            # Verificar saúde geral do serviço
            all_ok = all([database_ok, cache_ok, queue_ok, storage_ok])
            
            # Preparar resposta detalhada no formato padrão INNOVABIZ
            status_response = {
                "status": "healthy" if all_ok else "degraded",
                "service": self.service_name,
                "version": self.version,
                "timestamp": time.time(),
                "region": self.region,
                "environment": self.environment,
                "uptime_seconds": round(time.time() - self.start_time),
                "components": {
                    "database": "healthy" if database_ok else "unhealthy",
                    "cache": "healthy" if cache_ok else "unhealthy",
                    "message_queue": "healthy" if queue_ok else "unhealthy",
                    "storage": "healthy" if storage_ok else "unhealthy"
                },
                "tenant_context": tenant_id,
                "regional_context": regional_context
            }
            
            # Definir código de status HTTP baseado na saúde
            response_status_code = 200 if all_ok else 503
            
            return Response(
                content=status_response,
                status_code=response_status_code,
                media_type="application/json"
            )    def _setup_diagnostic_endpoint(self):
        """
        Configura endpoint de diagnóstico para KrakenD API Gateway.
        Segue os padrões de integração da plataforma INNOVABIZ com API Gateway.
        """
        @self.app.get("/diagnostic")
        async def diagnostic(request: Request) -> Dict[str, Any]:
            """
            Endpoint de diagnóstico detalhado com métricas avançadas
            para integração com KrakenD API Gateway e ferramentas de monitoramento.
            
            Segue as recomendações da ISO 27001 (seção A.12.1.3) e do NIST Cybersecurity Framework.
            
            Retorna:
                dict: Informações detalhadas de diagnóstico do serviço
            """
            # Extrair informações de contexto
            tenant_id = request.headers.get("X-Tenant-ID", "global")
            regional_context = request.headers.get("X-Regional-Context", "global")
            
            # Obter hostname e IP do servidor
            hostname = socket.gethostname()
            try:
                host_ip = socket.gethostbyname(hostname)
            except:
                host_ip = "unknown"
                
            # Coletar métricas de sistema operacional
            import psutil
            cpu_percent = psutil.cpu_percent(interval=0.1)
            memory_info = psutil.virtual_memory()
            disk_info = psutil.disk_usage('/')
            
            # Construir resposta de diagnóstico detalhada
            diagnostic_info = {
                "service": {
                    "name": self.service_name,
                    "version": self.version,
                    "build_id": self.build_id,
                    "commit_hash": self.commit_hash,
                    "uptime_seconds": round(time.time() - self.start_time),
                    "environment": self.environment,
                    "region": self.region
                },
                "context": {
                    "tenant_id": tenant_id,
                    "regional_context": regional_context
                },
                "host": {
                    "hostname": hostname,
                    "ip": host_ip,
                    "cpu_percent": cpu_percent,
                    "memory": {
                        "total_mb": memory_info.total / (1024 * 1024),
                        "available_mb": memory_info.available / (1024 * 1024),
                        "percent_used": memory_info.percent
                    },
                    "disk": {
                        "total_gb": disk_info.total / (1024 * 1024 * 1024),
                        "free_gb": disk_info.free / (1024 * 1024 * 1024),
                        "percent_used": disk_info.percent
                    }
                },
                "metrics_endpoint": self.metrics_path,
                "health_check_endpoint": self.health_check_path,
                "timestamp": time.time()
            }
            
            return diagnostic_info    def _setup_lifecycle_handlers(self):
        """
        Configura handlers para eventos de ciclo de vida da aplicação.
        Segue as práticas do ITIL v4 para Gerenciamento de Eventos.
        """
        @self.app.on_event("startup")
        async def on_startup():
            """
            Handler para evento de inicialização da aplicação.
            Registra o início do serviço e inicializa componentes de observabilidade.
            """
            logger.info(
                "Serviço de auditoria iniciado", 
                service=self.service_name,
                version=self.version,
                environment=self.environment,
                region=self.region
            )
            
            # Carregar políticas de retenção do banco de dados
            # e registrar métricas iniciais
            try:
                policies_count = await self._load_retention_policies()
                register_retention_policies(policies_count)
                logger.info(
                    "Políticas de retenção carregadas", 
                    count=policies_count
                )
            except Exception as e:
                logger.error(
                    "Falha ao carregar políticas de retenção",
                    error=str(e),
                    exc_info=True
                )
        
        @self.app.on_event("shutdown")
        async def on_shutdown():
            """
            Handler para evento de encerramento da aplicação.
            Realiza limpeza e registra o encerramento do serviço.
            """
            logger.info(
                "Serviço de auditoria em processo de encerramento",
                service=self.service_name,
                uptime_seconds=round(time.time() - self.start_time)
            )
    
    # Métodos de verificação de saúde dos componentes
    # Implementados de acordo com as recomendações do NIST e ISO 27001
    
    def _check_database_health(self) -> bool:
        """
        Verifica a saúde da conexão com o banco de dados.
        
        Em uma implementação real, isso faria uma verificação no PostgreSQL
        usando uma query simples como "SELECT 1" para validar conectividade.
        
        Returns:
            bool: True se o banco de dados está operacional
        """
        # Implementação simulada - em produção faria uma verificação real
        try:
            # Simulação de verificação de banco de dados
            return True
        except Exception as e:
            logger.error(
                "Falha na verificação de saúde do banco de dados",
                error=str(e),
                exc_info=True
            )
            return False    def _check_cache_health(self) -> bool:
        """
        Verifica a saúde da conexão com o cache Redis.
        
        Em uma implementação real, isso testaria a conectividade com o Redis
        utilizando um comando PING e verificaria a latência da resposta.
        
        Segue as recomendações do Redis Enterprise para monitoramento de saúde.
        
        Returns:
            bool: True se o cache Redis está operacional
        """
        # Implementação simulada - em produção faria uma verificação real
        try:
            # Simulação de verificação de cache Redis
            return True
        except Exception as e:
            logger.error(
                "Falha na verificação de saúde do Redis",
                error=str(e),
                exc_info=True
            )
            return False
    
    def _check_queue_health(self) -> bool:
        """
        Verifica a saúde da conexão com a fila de mensagens Kafka.
        
        Em uma implementação real, isso verificaria a conectividade com o Kafka
        e o status dos consumidores e produtores de eventos de auditoria.
        
        Segue as recomendações do Confluent para monitoramento de saúde do Kafka.
        
        Returns:
            bool: True se o Kafka está operacional
        """
        # Implementação simulada - em produção faria uma verificação real
        try:
            # Simulação de verificação de Kafka
            return True
        except Exception as e:
            logger.error(
                "Falha na verificação de saúde do Kafka",
                error=str(e),
                exc_info=True
            )
            return False
    
    def _check_storage_health(self) -> bool:
        """
        Verifica a saúde do serviço de armazenamento.
        
        Em uma implementação real, isso testaria a disponibilidade do sistema
        de armazenamento de logs e eventos de auditoria (ex: S3, Azure Blob).
        
        Segue as recomendações de padrões de observabilidade para sistemas distribuídos.
        
        Returns:
            bool: True se o serviço de armazenamento está operacional
        """
        # Implementação simulada - em produção faria uma verificação real
        try:
            # Simulação de verificação de serviço de armazenamento
            return True
        except Exception as e:
            logger.error(
                "Falha na verificação de saúde do serviço de armazenamento",
                error=str(e),
                exc_info=True
            )
            return False    async def _load_retention_policies(self) -> int:
        """
        Carrega políticas de retenção do banco de dados.
        
        Em uma implementação real, isso consultaria o banco de dados
        para obter as políticas de retenção ativas para cada tenant
        e contexto regional.
        
        Segue os princípios de governança de dados do DMBOK e as
        regulamentações como LGPD, GDPR e SOX para gestão do ciclo
        de vida de dados de auditoria.
        
        Returns:
            int: Número de políticas de retenção ativas
        """
        # Implementação simulada - em produção consultaria o banco de dados
        # Simulação de políticas para múltiplos tenants e regiões
        policies = [
            {"tenant_id": "tenant1", "regional_context": "br-south", "type": "time_based", "days": 90},
            {"tenant_id": "tenant1", "regional_context": "eu-central", "type": "time_based", "days": 180},
            {"tenant_id": "tenant2", "regional_context": "br-south", "type": "time_based", "days": 30},
            {"tenant_id": "tenant3", "regional_context": "global", "type": "compliance_based", "days": 365},
        ]
        
        return len(policies)


# Função de conveniência para integrar a observabilidade com a aplicação FastAPI
def setup_observability(app: FastAPI) -> ObservabilityIntegration:
    """
    Configura todos os componentes de observabilidade para a aplicação FastAPI.
    
    Esta função é o ponto de entrada principal para integrar a observabilidade
    com um aplicativo FastAPI do serviço de auditoria IAM.
    
    Args:
        app: Aplicativo FastAPI principal
        
    Returns:
        ObservabilityIntegration: Instância configurada da integração de observabilidade
    
    Exemplo:
        ```python
        from fastapi import FastAPI
        from src.api.app.integrations.observability import setup_observability
        
        app = FastAPI(title="IAM Audit Service")
        setup_observability(app)
        ```
    """
    # Obter versão do pacote ou usar valor padrão
    try:
        import pkg_resources
        version = pkg_resources.get_distribution("innovabiz-iam-audit").version
    except:
        version = os.getenv("SERVICE_VERSION", "1.0.0")
    
    # Criar e configurar a integração de observabilidade
    return ObservabilityIntegration(
        app=app,
        service_name="iam-audit",
        version=version
    ).setup()