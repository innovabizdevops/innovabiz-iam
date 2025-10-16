"""
Implementação final de métodos para o repositório TrustScore PostgreSQL.

Este arquivo contém os métodos de análise de tendências e anomalias para
o repositório TrustScore PostgreSQL.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta

from opentelemetry import trace

# Configuração de tracing
tracer = trace.get_tracer(__name__)


async def get_trust_score_trends(self,
                               user_id: str,
                               tenant_id: str,
                               days: int = 30,
                               context_id: Optional[str] = None) -> Dict[str, List[Dict[str, Any]]]:
    """
    Obtém tendências de pontuação de confiança para análise temporal.
    
    Args:
        user_id: ID do usuário
        tenant_id: ID do tenant
        days: Número de dias para análise de tendência
        context_id: Contexto específico (opcional)
        
    Returns:
        Dict[str, List[Dict[str, Any]]]: Dados de tendência por dimensão
    """
    with tracer.start_as_current_span("get_trust_score_trends") as span:
        span.set_attribute("user_id", user_id)
        span.set_attribute("tenant_id", tenant_id)
        span.set_attribute("days", days)
        
        # Definir período de análise
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)
        
        # Construir query base
        query_params = [user_id, tenant_id, start_date, end_date]
        param_index = 5
        
        # Construir a cláusula WHERE base
        where_clause = "WHERE h.user_id = $1 AND h.tenant_id = $2 AND h.created_at BETWEEN $3 AND $4"
        
        if context_id:
            where_clause += f" AND h.context_id = ${param_index}"
            query_params.append(context_id)
            param_index += 1
        
        # Query para pontuações gerais ao longo do tempo
        overall_query = f"""
            SELECT 
                DATE_TRUNC('day', h.created_at) as date,
                AVG(h.overall_score) as avg_score,
                MIN(h.overall_score) as min_score,
                MAX(h.overall_score) as max_score,
                COUNT(*) as evaluation_count
            FROM trust_score_history h
            {where_clause}
            GROUP BY DATE_TRUNC('day', h.created_at)
            ORDER BY DATE_TRUNC('day', h.created_at)
        """
        
        # Query para pontuações por dimensão ao longo do tempo
        dimension_query = f"""
            SELECT 
                DATE_TRUNC('day', h.created_at) as date,
                d.dimension,
                AVG(d.score) as avg_score,
                MIN(d.score) as min_score,
                MAX(d.score) as max_score,
                COUNT(*) as evaluation_count
            FROM trust_score_history h
            JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
            {where_clause}
            GROUP BY DATE_TRUNC('day', h.created_at), d.dimension
            ORDER BY DATE_TRUNC('day', h.created_at), d.dimension
        """
        
        # Query para contagem de anomalias ao longo do tempo
        anomaly_query = f"""
            SELECT 
                DATE_TRUNC('day', h.created_at) as date,
                a.anomaly_type,
                COUNT(*) as count
            FROM trust_score_history h
            JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
            {where_clause}
            GROUP BY DATE_TRUNC('day', h.created_at), a.anomaly_type
            ORDER BY DATE_TRUNC('day', h.created_at), a.anomaly_type
        """
        
        async with self.pool.acquire() as conn:
            # Executar queries
            overall_rows = await conn.fetch(overall_query, *query_params)
            dimension_rows = await conn.fetch(dimension_query, *query_params)
            anomaly_rows = await conn.fetch(anomaly_query, *query_params)
            
            # Processar resultados
            overall_trend = [{
                'date': row['date'].isoformat(),
                'avg_score': row['avg_score'],
                'min_score': row['min_score'],
                'max_score': row['max_score'],
                'count': row['evaluation_count']
            } for row in overall_rows]
            
            # Agrupar por dimensão
            dimension_trends = {}
            for row in dimension_rows:
                dim_str = row['dimension']
                if dim_str not in dimension_trends:
                    dimension_trends[dim_str] = []
                
                dimension_trends[dim_str].append({
                    'date': row['date'].isoformat(),
                    'avg_score': row['avg_score'],
                    'min_score': row['min_score'],
                    'max_score': row['max_score'],
                    'count': row['evaluation_count']
                })
            
            # Agrupar anomalias por tipo
            anomaly_trends = {}
            for row in anomaly_rows:
                anomaly_type = row['anomaly_type']
                if anomaly_type not in anomaly_trends:
                    anomaly_trends[anomaly_type] = []
                
                anomaly_trends[anomaly_type].append({
                    'date': row['date'].isoformat(),
                    'count': row['count']
                })
            
            return {
                'overall': overall_trend,
                'dimensions': dimension_trends,
                'anomalies': anomaly_trends
            }


async def get_anomaly_frequency(self,
                              tenant_id: str,
                              days: int = 30,
                              region_code: Optional[str] = None) -> Dict[str, int]:
    """
    Obtém frequência de tipos de anomalias detectadas no período.
    
    Args:
        tenant_id: ID do tenant
        days: Número de dias para análise
        region_code: Código da região (opcional)
        
    Returns:
        Dict[str, int]: Contagem de anomalias por tipo
    """
    with tracer.start_as_current_span("get_anomaly_frequency") as span:
        span.set_attribute("tenant_id", tenant_id)
        span.set_attribute("days", days)
        
        # Definir período de análise
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)
        
        # Construir query
        query_params = [tenant_id, start_date, end_date]
        param_index = 4
        
        query = """
            SELECT 
                a.anomaly_type,
                COUNT(*) as count,
                AVG(a.confidence) as avg_confidence,
                MAX(a.detected_at) as last_detected
            FROM trust_score_history h
            JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
            WHERE h.tenant_id = $1 AND h.created_at BETWEEN $2 AND $3
        """
        
        if region_code:
            query += f" AND h.region_code = ${param_index}"
            query_params.append(region_code)
            param_index += 1
        
        query += " GROUP BY a.anomaly_type ORDER BY COUNT(*) DESC"
        
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(query, *query_params)
            
            result = {}
            for row in rows:
                result[row['anomaly_type']] = {
                    'count': row['count'],
                    'avg_confidence': row['avg_confidence'],
                    'last_detected': row['last_detected'].isoformat()
                }
            
            return result