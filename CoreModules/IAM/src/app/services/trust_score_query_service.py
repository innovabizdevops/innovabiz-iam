"""
Serviço de consulta otimizada para o histórico de pontuações de confiança.

Este módulo implementa métodos otimizados para consulta e análise de dados
de pontuação de confiança, incluindo cache, materialização de visões e
estratégias de particionamento para consultas de alto desempenho.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import asyncio
import asyncpg
from typing import Dict, List, Optional, Any, Tuple, Union, Set
from datetime import datetime, timedelta
from cachetools import TTLCache, LRUCache
from opentelemetry import trace
from opentelemetry.trace import Span

# Importar repositório e modelos
from ..repositories.trust_score_repository import TrustScoreRepository
from ..trust_guard_models import (
    TrustScoreResult,
    TrustDimension,
    DetectedAnomaly,
    UserTrustProfile,
    AnomalyType,
    TrustScoreHistoryItem
)

# Configuração de logging e tracing
logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)


class TrustScoreQueryService:
    """
    Serviço para consultas otimizadas sobre históricos de pontuações de confiança.
    
    Implementa estratégias de cache, materialização e particionamento para 
    consultas de alta performance sobre grandes volumes de dados de confiança.
    """
    
    def __init__(
        self,
        repository: TrustScoreRepository,
        connection_pool: asyncpg.Pool,
        cache_ttl: int = 300,  # 5 minutos
        cache_max_size: int = 1000
    ):
        """
        Inicializa o serviço de consulta.
        
        Args:
            repository: Repositório base para consultas
            connection_pool: Pool de conexões para consultas SQL diretas
            cache_ttl: Tempo de vida do cache em segundos
            cache_max_size: Tamanho máximo do cache
        """
        self.repository = repository
        self.pool = connection_pool
        
        # Cache para perfis de usuário
        self.profile_cache = TTLCache(maxsize=cache_max_size, ttl=cache_ttl)
        
        # Cache para estatísticas de tenant
        self.tenant_stats_cache = TTLCache(maxsize=100, ttl=600)  # 10 minutos
        
        # Cache para resultados de consultas frequentes
        self.query_cache = LRUCache(maxsize=500)
        
        # Set para controle de atualizações pendentes de materialização
        self.materialization_pending: Set[str] = set()
    
    async def _get_user_profile_cached(
        self,
        user_id: str,
        tenant_id: str,
        context_id: Optional[str] = None
    ) -> UserTrustProfile:
        """
        Recupera perfil do usuário com suporte a cache.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: Contexto específico (opcional)
            
        Returns:
            UserTrustProfile: Perfil de confiança do usuário
        """
        cache_key = f"profile:{user_id}:{tenant_id}:{context_id}"
        
        if cache_key in self.profile_cache:
            return self.profile_cache[cache_key]
            
        profile = await self.repository.get_user_trust_profile(user_id, tenant_id, context_id)
        self.profile_cache[cache_key] = profile
        return profile
            
    async def get_user_trust_timeline(
        self,
        user_id: str,
        tenant_id: str,
        days: int = 30,
        context_id: Optional[str] = None,
        include_anomalies: bool = True,
        force_refresh: bool = False
    ) -> Dict[str, Any]:
        """
        Recupera linha do tempo otimizada de pontuações para um usuário específico.
        
        Utiliza otimizações como cache, pré-agregação e materialização para
        entregar histórico com alta performance.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            days: Número de dias no histórico
            context_id: Contexto específico (opcional)
            include_anomalies: Se deve incluir detalhes de anomalias
            force_refresh: Forçar atualização ignorando cache
            
        Returns:
            Dict[str, Any]: Linha do tempo com pontuações, tendências e estatísticas
        """
        with tracer.start_as_current_span("get_user_trust_timeline") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            span.set_attribute("days", days)
            
            cache_key = f"timeline:{user_id}:{tenant_id}:{days}:{context_id}:{include_anomalies}"
            
            if not force_refresh and cache_key in self.query_cache:
                span.set_attribute("cache_hit", True)
                return self.query_cache[cache_key]
            
            span.set_attribute("cache_hit", False)
            
            # Recuperar perfil do usuário (pode vir do cache interno do método)
            profile = await self._get_user_profile_cached(user_id, tenant_id, context_id)
            
            # Calcular período de análise
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)
            
            async with self.pool.acquire() as conn:
                # Consulta otimizada para pontuações por dia 
                # (usa função de agregação pré-computada)
                daily_scores = await conn.fetch(
                    """
                    SELECT * FROM get_user_daily_scores($1, $2, $3, $4, $5)
                    """,
                    user_id, tenant_id, start_date, end_date, context_id
                )
                
                # Converter para formato de saída
                timeline_data = [{
                    'date': row['date'].isoformat(),
                    'avg_score': row['avg_score'],
                    'min_score': row['min_score'],
                    'max_score': row['max_score'],
                    'evaluations': row['count']
                } for row in daily_scores]
                
                # Recuperar destaques de dimensão
                dimension_highlights = await conn.fetch(
                    """
                    SELECT 
                        d.dimension, 
                        AVG(d.score) as avg_score,
                        MIN(d.score) as min_score,
                        MAX(d.score) as max_score,
                        VARIANCE(d.score) as variance
                    FROM trust_score_history h
                    JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                    WHERE h.user_id = $1 
                    AND h.tenant_id = $2 
                    AND h.created_at BETWEEN $3 AND $4
                    """ + (f"AND h.context_id = ${5}" if context_id else ""),
                    *([user_id, tenant_id, start_date, end_date] + ([context_id] if context_id else []))
                )
                
                dimensions = {row['dimension']: {
                    'avg_score': row['avg_score'],
                    'min_score': row['min_score'],
                    'max_score': row['max_score'],
                    'variance': row['variance']
                } for row in dimension_highlights}
                
                # Recuperar anomalias mais recentes se solicitado
                recent_anomalies = []
                if include_anomalies:
                    anomaly_rows = await conn.fetch(
                        """
                        SELECT a.anomaly_id, a.anomaly_type, a.description,
                               a.severity, a.confidence, a.affected_dimensions,
                               a.detected_at
                        FROM trust_score_history h
                        JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
                        WHERE h.user_id = $1 
                        AND h.tenant_id = $2 
                        AND h.created_at BETWEEN $3 AND $4
                        """ + (f"AND h.context_id = ${5}" if context_id else "") + """
                        ORDER BY a.detected_at DESC
                        LIMIT 10
                        """,
                        *([user_id, tenant_id, start_date, end_date] + ([context_id] if context_id else []))
                    )
                    
                    # Formatar anomalias
                    for row in anomaly_rows:
                        recent_anomalies.append({
                            'anomaly_id': row['anomaly_id'],
                            'type': row['anomaly_type'],
                            'description': row['description'],
                            'severity': row['severity'],
                            'confidence': row['confidence'],
                            'affected_dimensions': row['affected_dimensions'],
                            'detected_at': row['detected_at'].isoformat()
                        })
            
            # Compilar resultado final
            result = {
                'user_id': user_id,
                'tenant_id': tenant_id,
                'latest_score': profile.latest_score,
                'timeline': timeline_data,
                'dimensions': dimensions,
                'anomalies': recent_anomalies,
                'profile_summary': profile.history_summary if profile.history_summary else {},
                'generated_at': datetime.now().isoformat()
            }
            
            # Armazenar em cache
            self.query_cache[cache_key] = result
            
            return result
            
    async def get_regional_comparison(
        self,
        user_id: str,
        tenant_id: str,
        region_code: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Compara pontuação de confiança de um usuário com a média regional.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            region_code: Código da região para comparação
            context_id: Contexto específico (opcional)
            
        Returns:
            Dict[str, Any]: Dados comparativos entre usuário e região
        """
        with tracer.start_as_current_span("get_regional_comparison") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            span.set_attribute("region_code", region_code)
            
            # Chave de cache
            cache_key = f"region_compare:{user_id}:{tenant_id}:{region_code}:{context_id}"
            
            if cache_key in self.query_cache:
                span.set_attribute("cache_hit", True)
                return self.query_cache[cache_key]
            
            span.set_attribute("cache_hit", False)
            
            async with self.pool.acquire() as conn:
                # Recuperar pontuação média do usuário
                user_avg = await conn.fetchrow(
                    """
                    SELECT 
                        AVG(h.overall_score) as avg_score,
                        COUNT(*) as evaluation_count
                    FROM trust_score_history h
                    WHERE h.user_id = $1 
                    AND h.tenant_id = $2 
                    AND h.region_code = $3
                    """ + (f"AND h.context_id = ${4}" if context_id else ""),
                    *([user_id, tenant_id, region_code] + ([context_id] if context_id else []))
                )
                
                # Recuperar estatísticas regionais
                region_stats = await conn.fetchrow(
                    """
                    SELECT 
                        AVG(h.overall_score) as avg_score,
                        PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY h.overall_score) as median_score,
                        COUNT(DISTINCT h.user_id) as user_count,
                        COUNT(*) as evaluation_count
                    FROM trust_score_history h
                    WHERE h.tenant_id = $1 
                    AND h.region_code = $2
                    """ + (f"AND h.context_id = ${3}" if context_id else ""),
                    *([tenant_id, region_code] + ([context_id] if context_id else []))
                )
                
                # Recuperar estatísticas por dimensão para o usuário
                user_dims = await conn.fetch(
                    """
                    SELECT 
                        d.dimension,
                        AVG(d.score) as avg_score
                    FROM trust_score_history h
                    JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                    WHERE h.user_id = $1 
                    AND h.tenant_id = $2 
                    AND h.region_code = $3
                    """ + (f"AND h.context_id = ${4}" if context_id else "") + """
                    GROUP BY d.dimension
                    """,
                    *([user_id, tenant_id, region_code] + ([context_id] if context_id else []))
                )
                
                user_dimensions = {row['dimension']: row['avg_score'] for row in user_dims}
                
                # Recuperar estatísticas por dimensão para a região
                region_dims = await conn.fetch(
                    """
                    SELECT 
                        d.dimension,
                        AVG(d.score) as avg_score
                    FROM trust_score_history h
                    JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                    WHERE h.tenant_id = $1 
                    AND h.region_code = $2
                    """ + (f"AND h.context_id = ${3}" if context_id else "") + """
                    GROUP BY d.dimension
                    """,
                    *([tenant_id, region_code] + ([context_id] if context_id else []))
                )
                
                region_dimensions = {row['dimension']: row['avg_score'] for row in region_dims}
                
                # Calcular percentil do usuário
                user_percentile = await conn.fetchval(
                    """
                    WITH user_score AS (
                        SELECT AVG(h.overall_score) as score
                        FROM trust_score_history h
                        WHERE h.user_id = $1 
                        AND h.tenant_id = $2 
                        AND h.region_code = $3
                        """ + (f"AND h.context_id = ${4}" if context_id else "") + """
                    ),
                    scores_cdf AS (
                        SELECT 
                            h.overall_score,
                            PERCENT_RANK() OVER (ORDER BY h.overall_score) as percentile
                        FROM trust_score_history h
                        WHERE h.tenant_id = $2
                        AND h.region_code = $3
                        """ + (f"AND h.context_id = ${4}" if context_id else "") + """
                    )
                    SELECT MAX(s.percentile) * 100
                    FROM scores_cdf s, user_score u
                    WHERE s.overall_score <= u.score
                    """,
                    *([user_id, tenant_id, region_code] + ([context_id] if context_id else []))
                )
            
            # Compilar resultado
            result = {
                'user': {
                    'id': user_id,
                    'avg_score': user_avg['avg_score'] if user_avg else None,
                    'evaluation_count': user_avg['evaluation_count'] if user_avg else 0,
                    'dimensions': user_dimensions,
                    'percentile': user_percentile or 0
                },
                'region': {
                    'code': region_code,
                    'avg_score': region_stats['avg_score'] if region_stats else None,
                    'median_score': region_stats['median_score'] if region_stats else None,
                    'user_count': region_stats['user_count'] if region_stats else 0,
                    'evaluation_count': region_stats['evaluation_count'] if region_stats else 0,
                    'dimensions': region_dimensions
                }
            }
            
            # Armazenar em cache
            self.query_cache[cache_key] = result
            
            return result
            
    async def get_anomaly_details(
        self,
        user_id: str,
        tenant_id: str,
        days: int = 90,
        anomaly_types: Optional[List[str]] = None,
        min_severity: Optional[str] = None,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Recupera detalhes avançados de anomalias para análise de segurança.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            days: Histórico em dias
            anomaly_types: Lista de tipos de anomalias para filtrar
            min_severity: Severidade mínima a considerar
            context_id: Contexto específico (opcional)
            
        Returns:
            Dict[str, Any]: Dados detalhados de anomalias com padrões e tendências
        """
        with tracer.start_as_current_span("get_anomaly_details") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            span.set_attribute("days", days)
            
            # Chave de cache
            cache_key = f"anomaly_details:{user_id}:{tenant_id}:{days}:{anomaly_types}:{min_severity}:{context_id}"
            
            if cache_key in self.query_cache:
                span.set_attribute("cache_hit", True)
                return self.query_cache[cache_key]
            
            span.set_attribute("cache_hit", False)
            
            # Calcular período
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)
            
            # Parâmetros base
            params = [user_id, tenant_id, start_date, end_date]
            param_index = 5
            
            # Construir cláusula WHERE base
            where_clause = "WHERE h.user_id = $1 AND h.tenant_id = $2 AND a.detected_at BETWEEN $3 AND $4"
            
            if context_id:
                where_clause += f" AND h.context_id = ${param_index}"
                params.append(context_id)
                param_index += 1
                
            if min_severity:
                severity_ranks = {
                    'low': 1,
                    'medium': 2,
                    'high': 3,
                    'critical': 4
                }
                severity_filter = []
                min_rank = severity_ranks.get(min_severity.lower(), 1)
                
                for sev, rank in severity_ranks.items():
                    if rank >= min_rank:
                        severity_filter.append(sev)
                
                if severity_filter:
                    placeholders = ", ".join([f"${param_index + i}" for i in range(len(severity_filter))])
                    where_clause += f" AND a.severity IN ({placeholders})"
                    params.extend(severity_filter)
                    param_index += len(severity_filter)
            
            if anomaly_types:
                placeholders = ", ".join([f"${param_index + i}" for i in range(len(anomaly_types))])
                where_clause += f" AND a.anomaly_type IN ({placeholders})"
                params.extend(anomaly_types)
                param_index += len(anomaly_types)
            
            async with self.pool.acquire() as conn:
                # Recuperar anomalias com detalhes
                anomalies = await conn.fetch(
                    f"""
                    SELECT 
                        a.anomaly_id,
                        a.anomaly_type,
                        a.description,
                        a.severity,
                        a.confidence,
                        a.affected_dimensions,
                        a.detected_at,
                        a.metadata,
                        h.overall_score,
                        h.region_code,
                        h.context_id
                    FROM trust_score_history h
                    JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
                    {where_clause}
                    ORDER BY a.detected_at DESC
                    """,
                    *params
                )
                
                # Agrupar anomalias por tipo
                anomalies_by_type = {}
                for row in anomalies:
                    anomaly_type = row['anomaly_type']
                    if anomaly_type not in anomalies_by_type:
                        anomalies_by_type[anomaly_type] = []
                    
                    anomalies_by_type[anomaly_type].append({
                        'anomaly_id': row['anomaly_id'],
                        'description': row['description'],
                        'severity': row['severity'],
                        'confidence': row['confidence'],
                        'affected_dimensions': row['affected_dimensions'],
                        'detected_at': row['detected_at'].isoformat(),
                        'overall_score': row['overall_score'],
                        'region_code': row['region_code'],
                        'context_id': row['context_id'],
                        'metadata': row['metadata']
                    })
                
                # Calcular estatísticas por tipo de anomalia
                anomaly_stats = {}
                for anomaly_type, anomalies_list in anomalies_by_type.items():
                    if not anomalies_list:
                        continue
                        
                    # Extrai severidades
                    severities = [a['severity'] for a in anomalies_list]
                    severity_counts = {}
                    for sev in severities:
                        if sev not in severity_counts:
                            severity_counts[sev] = 0
                        severity_counts[sev] += 1
                    
                    # Calcula confiança média
                    avg_confidence = sum(a['confidence'] for a in anomalies_list) / len(anomalies_list)
                    
                    # Calcula distribuição temporal
                    first_date = min(datetime.fromisoformat(a['detected_at']) for a in anomalies_list)
                    last_date = max(datetime.fromisoformat(a['detected_at']) for a in anomalies_list)
                    
                    anomaly_stats[anomaly_type] = {
                        'count': len(anomalies_list),
                        'severity_distribution': severity_counts,
                        'avg_confidence': avg_confidence,
                        'first_detected': first_date.isoformat(),
                        'last_detected': last_date.isoformat()
                    }
                
                # Obter padrões temporais de anomalias
                time_patterns = await conn.fetch(
                    f"""
                    SELECT 
                        DATE_TRUNC('day', a.detected_at) as date,
                        a.anomaly_type,
                        COUNT(*) as count
                    FROM trust_score_history h
                    JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
                    {where_clause}
                    GROUP BY DATE_TRUNC('day', a.detected_at), a.anomaly_type
                    ORDER BY DATE_TRUNC('day', a.detected_at), a.anomaly_type
                    """,
                    *params
                )
                
                # Formatar padrões temporais
                temporal_patterns = {}
                for row in time_patterns:
                    anomaly_type = row['anomaly_type']
                    if anomaly_type not in temporal_patterns:
                        temporal_patterns[anomaly_type] = []
                    
                    temporal_patterns[anomaly_type].append({
                        'date': row['date'].isoformat(),
                        'count': row['count']
                    })
            
            # Compilar resultado
            result = {
                'user_id': user_id,
                'tenant_id': tenant_id,
                'period': {
                    'start': start_date.isoformat(),
                    'end': end_date.isoformat(),
                    'days': days
                },
                'anomalies': anomalies_by_type,
                'statistics': anomaly_stats,
                'temporal_patterns': temporal_patterns,
                'total_anomalies': sum(len(items) for items in anomalies_by_type.values())
            }
            
            # Armazenar em cache
            self.query_cache[cache_key] = result
            
            return result
    
    async def update_materialized_views(self) -> bool:
        """
        Atualiza visões materializadas para consultas de alta performance.
        
        Este método deve ser executado periodicamente para manter as visões
        materializadas atualizadas com dados recentes.
        
        Returns:
            bool: True se atualização foi bem-sucedida
        """
        with tracer.start_as_current_span("update_materialized_views") as span:
            try:
                async with self.pool.acquire() as conn:
                    # Atualizar visão materializada de estatísticas diárias
                    await conn.execute("REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trust_score_daily_stats")
                    
                    # Atualizar visão materializada de estatísticas por região
                    await conn.execute("REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trust_score_regional_stats")
                    
                    # Atualizar visão materializada de anomalias
                    await conn.execute("REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trust_score_anomaly_stats")
                    
                    # Limpar cache após atualização
                    self.query_cache.clear()
                    
                    return True
            except Exception as e:
                logger.error(f"Erro ao atualizar visões materializadas: {e}")
                span.record_exception(e)
                return False