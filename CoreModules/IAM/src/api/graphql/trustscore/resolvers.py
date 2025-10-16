"""
Resolvers GraphQL para o módulo TrustScore.

Este módulo implementa os resolvers para consultas e mutações relacionadas
ao TrustScore, permitindo acesso otimizado aos dados de pontuação de confiança
através da API GraphQL centralizada.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import asyncio
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional, Union
from graphql import GraphQLResolveInfo

from ....app.services.trust_score_query_service import TrustScoreQueryService
from ....app.repositories.trust_score_repository import TrustScoreRepository
from ....app.trust_guard_models import TrustScoreResult, UserTrustProfile

# Configuração de logging
logger = logging.getLogger(__name__)


class TrustScoreResolvers:
    """
    Implementação dos resolvers GraphQL para o TrustScore.
    """
    
    def __init__(
        self,
        query_service: TrustScoreQueryService,
        repository: TrustScoreRepository
    ):
        """
        Inicializa os resolvers com as dependências necessárias.
        
        Args:
            query_service: Serviço de consulta otimizada para TrustScore
            repository: Repositório base para consultas diretas
        """
        self.query_service = query_service
        self.repository = repository
    
    async def get_user_trust_profile(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        user_id: str,
        tenant_id: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Resolver para obter o perfil de confiança de um usuário.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            
        Returns:
            Dict: Perfil de confiança formatado
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read")
            
            # Buscar perfil via repositório
            profile = await self.repository.get_user_trust_profile(
                user_id=user_id,
                tenant_id=tenant_id,
                context_id=context_id
            )
            
            # Converter para formato GraphQL
            return {
                "userId": profile.user_id,
                "tenantId": profile.tenant_id,
                "latestScore": profile.latest_score,
                "trustScoreHistory": [
                    {
                        "score": item.score,
                        "dimensionScores": item.dimension_scores,
                        "confidenceLevel": item.confidence_level,
                        "regionCode": item.region_code,
                        "contextId": item.context_id,
                        "timestamp": item.timestamp,
                        "anomalyCount": item.anomaly_count
                    }
                    for item in profile.trust_score_history
                ],
                "historySummary": profile.history_summary,
                "createdAt": profile.created_at,
                "updatedAt": profile.updated_at
            }
        except Exception as e:
            logger.error(f"Erro ao buscar perfil de confiança: {e}")
            raise
    
    async def get_trust_score_timeline(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        user_id: str,
        tenant_id: str,
        days: int = 30,
        context_id: Optional[str] = None,
        include_anomalies: bool = True
    ) -> Dict[str, Any]:
        """
        Resolver para obter linha do tempo de pontuações de confiança.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            user_id: ID do usuário
            tenant_id: ID do tenant
            days: Número de dias de histórico
            context_id: ID do contexto (opcional)
            include_anomalies: Se deve incluir anomalias
            
        Returns:
            Dict: Linha do tempo formatada
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read")
            
            # Buscar timeline via serviço otimizado
            timeline = await self.query_service.get_user_trust_timeline(
                user_id=user_id,
                tenant_id=tenant_id,
                days=days,
                context_id=context_id,
                include_anomalies=include_anomalies
            )
            
            # Transformar para o formato esperado pelo schema GraphQL
            return {
                "userId": timeline["user_id"],
                "tenantId": timeline["tenant_id"],
                "latestScore": timeline["latest_score"],
                "timeline": timeline["timeline"],
                "dimensions": timeline["dimensions"],
                "anomalies": timeline.get("anomalies", []),
                "profileSummary": timeline["profile_summary"],
                "generatedAt": timeline["generated_at"]
            }
        except Exception as e:
            logger.error(f"Erro ao buscar linha do tempo: {e}")
            raise
    
    async def get_regional_comparison(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        user_id: str,
        tenant_id: str,
        region_code: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Resolver para obter comparação regional de pontuação de confiança.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            user_id: ID do usuário
            tenant_id: ID do tenant
            region_code: Código da região
            context_id: ID do contexto (opcional)
            
        Returns:
            Dict: Comparação regional formatada
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read")
            
            # Buscar comparação regional via serviço otimizado
            comparison = await self.query_service.get_regional_comparison(
                user_id=user_id,
                tenant_id=tenant_id,
                region_code=region_code,
                context_id=context_id
            )
            
            # Retornar dados no formato esperado pelo schema GraphQL
            return {
                "user": {
                    "id": comparison["user"]["id"],
                    "avgScore": comparison["user"]["avg_score"],
                    "evaluationCount": comparison["user"]["evaluation_count"],
                    "dimensions": comparison["user"]["dimensions"],
                    "percentile": comparison["user"]["percentile"]
                },
                "region": {
                    "code": comparison["region"]["code"],
                    "avgScore": comparison["region"]["avg_score"],
                    "medianScore": comparison["region"]["median_score"],
                    "userCount": comparison["region"]["user_count"],
                    "evaluationCount": comparison["region"]["evaluation_count"],
                    "dimensions": comparison["region"]["dimensions"]
                }
            }
        except Exception as e:
            logger.error(f"Erro ao buscar comparação regional: {e}")
            raise
    
    async def get_anomaly_details(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        user_id: str,
        tenant_id: str,
        days: int = 90,
        anomaly_types: Optional[List[str]] = None,
        min_severity: Optional[str] = None,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Resolver para obter detalhes de anomalias de um usuário.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            user_id: ID do usuário
            tenant_id: ID do tenant
            days: Histórico em dias
            anomaly_types: Lista de tipos de anomalias para filtrar
            min_severity: Severidade mínima a considerar
            context_id: ID do contexto (opcional)
            
        Returns:
            Dict: Detalhes de anomalias formatados
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read")
            
            # Buscar detalhes via serviço otimizado
            details = await self.query_service.get_anomaly_details(
                user_id=user_id,
                tenant_id=tenant_id,
                days=days,
                anomaly_types=anomaly_types,
                min_severity=min_severity,
                context_id=context_id
            )
            
            # Retornar no formato esperado pelo schema GraphQL
            return {
                "userId": details["user_id"],
                "tenantId": details["tenant_id"],
                "period": details["period"],
                "anomalies": details["anomalies"],
                "statistics": details["statistics"],
                "temporalPatterns": details["temporal_patterns"],
                "totalAnomalies": details["total_anomalies"]
            }
        except Exception as e:
            logger.error(f"Erro ao buscar detalhes de anomalias: {e}")
            raise
    
    async def get_tenant_statistics(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        tenant_id: str,
        days: int = 30,
        region_code: Optional[str] = None,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Resolver para obter estatísticas agregadas por tenant.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            tenant_id: ID do tenant
            days: Histórico em dias
            region_code: Código da região (opcional)
            context_id: ID do contexto (opcional)
            
        Returns:
            Dict: Estatísticas do tenant formatadas
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read_stats")
            
            # Calcular período de análise
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)
            
            # Buscar estatísticas via repositório
            stats = await self.repository.get_tenant_statistics(
                tenant_id=tenant_id,
                start_date=start_date,
                end_date=end_date,
                region_code=region_code,
                context_id=context_id
            )
            
            # Retornar no formato esperado pelo schema GraphQL
            return {
                "tenantId": tenant_id,
                "userCount": stats["user_count"],
                "evaluationCount": stats["evaluation_count"],
                "averageScore": stats["average_score"],
                "dimensionStats": stats["dimension_stats"],
                "anomalyStats": stats["anomaly_stats"],
                "regionalDistribution": stats.get("regional_distribution"),
                "periodSummary": {
                    "start": start_date.isoformat(),
                    "end": end_date.isoformat(),
                    "days": days
                }
            }
        except Exception as e:
            logger.error(f"Erro ao buscar estatísticas do tenant: {e}")
            raise
    
    async def get_trust_score_results(
        self,
        obj: Any,
        info: GraphQLResolveInfo,
        user_id: Optional[str] = None,
        tenant_id: str = None,
        context_id: Optional[str] = None,
        region_code: Optional[str] = None,
        limit: int = 10,
        offset: int = 0,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        sort_by: str = "timestamp",
        sort_direction: str = "desc"
    ) -> Dict[str, Any]:
        """
        Resolver para listar resultados de pontuação de confiança paginados.
        
        Args:
            obj: Objeto raiz GraphQL
            info: Informações de contexto da consulta
            user_id: ID do usuário (opcional)
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            region_code: Código da região (opcional)
            limit: Limite de resultados por página
            offset: Deslocamento para paginação
            start_date: Data inicial (opcional)
            end_date: Data final (opcional)
            sort_by: Campo para ordenação
            sort_direction: Direção da ordenação (asc/desc)
            
        Returns:
            Dict: Resultados paginados
        """
        try:
            # Verificar autenticação e autorização
            await self._verify_access(info, tenant_id, "trust_score:read")
            
            # Converter datas se fornecidas
            start_date_obj = None
            if start_date:
                start_date_obj = datetime.fromisoformat(start_date)
                
            end_date_obj = None
            if end_date:
                end_date_obj = datetime.fromisoformat(end_date)
            
            # Buscar histórico via repositório
            history_results = await self.repository.get_user_trust_history(
                user_id=user_id,
                tenant_id=tenant_id,
                context_id=context_id,
                region_code=region_code,
                limit=limit,
                offset=offset,
                start_date=start_date_obj,
                end_date=end_date_obj,
                sort_by=sort_by,
                sort_direction=sort_direction
            )
            
            # Formatar resultados para o schema GraphQL
            edges = []
            for idx, result in enumerate(history_results['items']):
                cursor = f"{offset + idx}"
                
                # Buscar detalhes completos se necessário
                score_details = None
                if history_results.get('include_details'):
                    score_details = result
                else:
                    score_details = await self.repository.get_score_details(result['id'])
                
                # Formatar para o schema GraphQL
                node = {
                    "userId": score_details["user_id"],
                    "tenantId": score_details["tenant_id"],
                    "contextId": score_details.get("context_id"),
                    "overallScore": score_details["overall_score"],
                    "dimensionScores": score_details["dimension_scores"],
                    "regionalContext": score_details.get("region_code"),
                    "confidenceLevel": score_details["confidence_level"],
                    "evaluationTimeMs": score_details.get("evaluation_time_ms", 0),
                    "timestamp": score_details["created_at"],
                    "factors": score_details.get("factors", []),
                    "anomalies": score_details.get("anomalies", []),
                    "metadata": score_details.get("metadata", {})
                }
                
                edges.append({
                    "node": node,
                    "cursor": cursor
                })
            
            # Montar informações de paginação
            has_next = history_results['total_count'] > (offset + limit)
            has_previous = offset > 0
            
            page_info = {
                "hasNextPage": has_next,
                "hasPreviousPage": has_previous,
                "startCursor": str(offset) if edges else None,
                "endCursor": str(offset + len(edges) - 1) if edges else None
            }
            
            # Retornar conexão paginada
            return {
                "edges": edges,
                "pageInfo": page_info,
                "totalCount": history_results['total_count']
            }
        except Exception as e:
            logger.error(f"Erro ao listar resultados de pontuação: {e}")
            raise
    
    async def _verify_access(
        self,
        info: GraphQLResolveInfo,
        tenant_id: str,
        permission: str
    ) -> None:
        """
        Verifica se o usuário tem acesso aos dados solicitados.
        
        Args:
            info: Informações de contexto da consulta
            tenant_id: ID do tenant
            permission: Permissão necessária
            
        Raises:
            Exception: Se usuário não tiver autorização
        """
        # Obter contexto da requisição
        context = info.context
        
        # Verificar se usuário está autenticado
        if not hasattr(context, 'user') or not context.user:
            raise Exception("Usuário não autenticado")
        
        # Verificar acesso ao tenant
        if not context.user.has_tenant_access(tenant_id):
            raise Exception(f"Usuário não tem acesso ao tenant {tenant_id}")
        
        # Verificar permissão específica
        if not context.user.has_permission(permission):
            raise Exception(f"Usuário não tem permissão {permission}")