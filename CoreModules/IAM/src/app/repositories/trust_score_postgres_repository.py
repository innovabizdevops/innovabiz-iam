"""
Implementação PostgreSQL do repositório para histórico de pontuações de confiança.

Este módulo implementa a interface TrustScoreRepository usando PostgreSQL como
base de dados, fornecendo funcionalidades completas de persistência e recuperação
de dados de pontuação de confiança com suporte a multi-tenant e multi-contexto.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import asyncpg
from typing import Dict, List, Optional, Tuple, Any, Union, cast
from datetime import datetime, timedelta

from opentelemetry import trace
from opentelemetry.trace import Span

# Importar modelos e interface
from ..trust_guard_models import (
    TrustScoreResult,
    TrustScoreFactorModel,
    DetectedAnomaly,
    TrustDimension,
    UserTrustProfile,
    AnomalyType,
    AnomalySeverity,
    FactorType,
    TrustScoreHistoryItem
)
from .trust_score_repository import TrustScoreRepository

# Configuração de logging e tracing
logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)


class TrustScorePostgresRepository(TrustScoreRepository):
    """
    Implementação do repositório TrustScore usando PostgreSQL.
    
    Fornece persistência e recuperação eficiente de dados de pontuação
    de confiança com suporte completo a multi-tenant e multi-contexto.
    """
    
    def __init__(self, connection_pool: asyncpg.Pool):
        """
        Inicializa o repositório com um pool de conexões PostgreSQL.
        
        Args:
            connection_pool: Pool de conexões ao banco de dados PostgreSQL
        """
        self.pool = connection_pool
    
    async def save_trust_score_result(self, 
                                     result: TrustScoreResult,
                                     factors: Dict[TrustDimension, List[TrustScoreFactorModel]] = None,
                                     anomalies: List[DetectedAnomaly] = None) -> int:
        """
        Persiste o resultado de uma avaliação de confiança e seus componentes.
        
        Args:
            result: Resultado completo da avaliação de pontuação de confiança
            factors: Fatores que influenciaram cada dimensão (opcional)
            anomalies: Anomalias detectadas durante a avaliação (opcional)
            
        Returns:
            int: ID do registro de histórico criado
        """
        with tracer.start_as_current_span("save_trust_score_result") as span:
            # Preparar metadados para JSON
            metadata = {}
            if hasattr(result, 'metadata') and result.metadata:
                metadata = result.metadata
            
            async with self.pool.acquire() as conn:
                async with conn.transaction():
                    # Inserir registro principal
                    history_id = await conn.fetchval(
                        """
                        INSERT INTO trust_score_history 
                        (user_id, tenant_id, context_id, region_code, overall_score, 
                         confidence_level, evaluation_time_ms, created_at, metadata)
                        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                        RETURNING id
                        """,
                        result.user_id,
                        result.tenant_id,
                        result.context_id,
                        result.regional_context,
                        result.overall_score,
                        result.confidence_level,
                        result.evaluation_time_ms,
                        result.timestamp,
                        json.dumps(metadata)
                    )
                    
                    span.set_attribute("history_id", history_id)
                    span.set_attribute("user_id", result.user_id)
                    span.set_attribute("tenant_id", result.tenant_id)
                    
                    # Inserir pontuações por dimensão
                    if result.dimension_scores:
                        dimension_values = [(
                            history_id,
                            dim.value,  # Usando .value para enum
                            score,
                            result.timestamp
                        ) for dim, score in result.dimension_scores.items()]
                        
                        await conn.executemany(
                            """
                            INSERT INTO trust_score_dimension_history
                            (trust_score_history_id, dimension, score, created_at)
                            VALUES ($1, $2, $3, $4)
                            """,
                            dimension_values
                        )
                    
                    # Inserir fatores influenciadores
                    if factors:
                        factor_values = []
                        for dimension, factor_list in factors.items():
                            for factor in factor_list:
                                factor_values.append((
                                    history_id,
                                    factor.factor_id,
                                    dimension.value,  # Usando .value para enum
                                    factor.name,
                                    factor.description,
                                    factor.type.value,  # Usando .value para enum
                                    factor.weight,
                                    factor.value,
                                    json.dumps(factor.metadata) if factor.metadata else "{}",
                                    factor.created_at
                                ))
                        
                        if factor_values:
                            await conn.executemany(
                                """
                                INSERT INTO trust_score_factors
                                (trust_score_history_id, factor_id, dimension, factor_name,
                                 factor_description, factor_type, weight, value, metadata, created_at)
                                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                                """,
                                factor_values
                            )
                    
                    # Inserir anomalias detectadas
                    if anomalies:
                        anomaly_values = []
                        for anomaly in anomalies:
                            anomaly_values.append((
                                history_id,
                                anomaly.anomaly_id,
                                anomaly.type.value,  # Usando .value para enum
                                anomaly.description,
                                anomaly.severity.value,  # Usando .value para enum
                                anomaly.confidence,
                                json.dumps([d.value for d in anomaly.affected_dimensions]) if anomaly.affected_dimensions else "[]",
                                json.dumps(anomaly.metadata) if anomaly.metadata else "{}",
                                anomaly.detected_at
                            ))
                        
                        if anomaly_values:
                            await conn.executemany(
                                """
                                INSERT INTO trust_score_anomalies
                                (trust_score_history_id, anomaly_id, anomaly_type, description,
                                 severity, confidence, affected_dimensions, metadata, detected_at)
                                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                                """,
                                anomaly_values
                            )
                    
                    # Atualizar perfil de confiança do usuário com novo resultado
                    await self._update_user_profile_with_result(conn, result, factors, anomalies)
                    
                    return history_id
    
    async def _update_user_profile_with_result(self, conn, 
                                              result: TrustScoreResult,
                                              factors: Dict[TrustDimension, List[TrustScoreFactorModel]] = None,
                                              anomalies: List[DetectedAnomaly] = None):
        """
        Método auxiliar para atualizar o perfil de confiança do usuário com um novo resultado.
        
        Args:
            conn: Conexão com o banco de dados
            result: Resultado da avaliação de confiança
            factors: Fatores que influenciaram a pontuação
            anomalies: Anomalias detectadas
        """
        with tracer.start_as_current_span("_update_user_profile_with_result") as span:
            # Verificar se o perfil existe
            profile_exists = await conn.fetchval(
                """
                SELECT EXISTS(
                    SELECT 1 FROM user_trust_profiles 
                    WHERE user_id = $1 AND tenant_id = $2
                )
                """,
                result.user_id, result.tenant_id
            )
            
            # Criar estrutura de dados de histórico para o perfil
            history_item = TrustScoreHistoryItem(
                score=result.overall_score,
                dimension_scores=result.dimension_scores,
                confidence_level=result.confidence_level,
                region_code=result.regional_context,
                context_id=result.context_id,
                timestamp=result.timestamp,
                anomaly_count=len(anomalies) if anomalies else 0
            )
            
            # Preparar dados de fatores e anomalias para atualizar perfil
            top_factors = {}
            if factors:
                for dim, factor_list in factors.items():
                    # Filtrar apenas os fatores mais relevantes para cada dimensão
                    relevant_factors = sorted(factor_list, key=lambda f: f.weight * abs(f.value), reverse=True)[:3]
                    top_factors[dim.value] = [
                        {
                            "id": f.factor_id,
                            "name": f.name,
                            "type": f.type.value,
                            "weight": f.weight,
                            "value": f.value
                        }
                        for f in relevant_factors
                    ]
            
            # Extrair informações de dispositivo e localização para o histórico
            device_attributes = None
            location_data = None
            
            if hasattr(result, 'metadata') and result.metadata:
                if 'device_data' in result.metadata:
                    device_data = result.metadata['device_data']
                    device_id = device_data.get('device_id')
                    
                    if device_id:
                        device_attributes = {
                            "device_id": device_id,
                            "os": device_data.get('os'),
                            "browser": device_data.get('browser'),
                            "screen_resolution": device_data.get('screen_resolution'),
                            "timezone": device_data.get('timezone')
                        }
                
                if 'location_data' in result.metadata:
                    location_data = result.metadata['location_data']
            
            if profile_exists:
                # Atualizar perfil existente
                await conn.execute(
                    """
                    UPDATE user_trust_profiles 
                    SET 
                        latest_score = $3,
                        trust_score_history = array_append(
                            CASE 
                                WHEN array_length(trust_score_history, 1) >= 20 
                                THEN trust_score_history[2:array_length(trust_score_history, 1)]
                                ELSE trust_score_history
                            END,
                            $4::jsonb
                        ),
                        history_summary = jsonb_set(
                            COALESCE(history_summary, '{}'::jsonb),
                            '{top_factors}',
                            $5::jsonb,
                            true
                        ),
                        updated_at = NOW()
                    WHERE user_id = $1 AND tenant_id = $2
                    """,
                    result.user_id,
                    result.tenant_id,
                    result.overall_score,
                    json.dumps(history_item.__dict__),
                    json.dumps(top_factors)
                )
                
                # Atualizar dados de dispositivo se disponíveis
                if device_attributes:
                    await conn.execute(
                        """
                        UPDATE user_trust_profiles
                        SET history_summary = jsonb_set(
                            COALESCE(history_summary, '{}'::jsonb),
                            '{device_attributes}',
                            jsonb_set(
                                COALESCE(history_summary->'device_attributes', '{}'::jsonb),
                                $3::text[], 
                                $4::jsonb,
                                true
                            ),
                            true
                        )
                        WHERE user_id = $1 AND tenant_id = $2
                        """,
                        result.user_id,
                        result.tenant_id,
                        [device_attributes["device_id"]],
                        json.dumps(device_attributes)
                    )
                
                # Atualizar dados de localização se disponíveis
                if location_data:
                    # Adicionar à lista de localizações usuais se necessário
                    location_key = location_data.get('country') or location_data.get('city')
                    if location_key:
                        await conn.execute(
                            """
                            UPDATE user_trust_profiles
                            SET history_summary = jsonb_set(
                                COALESCE(history_summary, '{}'::jsonb),
                                '{usual_locations}',
                                CASE 
                                    WHEN history_summary->'usual_locations' ? $3 THEN history_summary->'usual_locations'
                                    ELSE jsonb_insert(
                                        COALESCE(history_summary->'usual_locations', '[]'::jsonb),
                                        '{0}',
                                        to_jsonb($3::text),
                                        true
                                    )
                                END,
                                true
                            )
                            WHERE user_id = $1 AND tenant_id = $2
                            """,
                            result.user_id,
                            result.tenant_id,
                            location_key
                        )
            else:
                # Criar novo perfil
                await conn.execute(
                    """
                    INSERT INTO user_trust_profiles 
                    (user_id, tenant_id, latest_score, trust_score_history, 
                     history_summary, created_at, updated_at)
                    VALUES ($1, $2, $3, ARRAY[$4::jsonb], $5, NOW(), NOW())
                    """,
                    result.user_id,
                    result.tenant_id,
                    result.overall_score,
                    json.dumps(history_item.__dict__),
                    json.dumps({
                        "top_factors": top_factors,
                        "device_attributes": {device_attributes["device_id"]: device_attributes} if device_attributes else {},
                        "usual_locations": [location_key] if location_data and location_key else []
                    })
                )
                
    async def get_user_trust_history(self, 
                                     user_id: str, 
                                     tenant_id: str,
                                     context_id: Optional[str] = None,
                                     region_code: Optional[str] = None,
                                     start_date: Optional[datetime] = None,
                                     end_date: Optional[datetime] = None,
                                     limit: int = 20,
                                     offset: int = 0) -> List[TrustScoreResult]:
        """
        Recupera o histórico de pontuação de confiança de um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: Filtrar por contexto específico (opcional)
            region_code: Filtrar por região específica (opcional)
            start_date: Data inicial para filtro (opcional)
            end_date: Data final para filtro (opcional)
            limit: Número máximo de registros a retornar
            offset: Deslocamento para paginação
            
        Returns:
            List[TrustScoreResult]: Lista de resultados de avaliação no período
        """
        with tracer.start_as_current_span("get_user_trust_history") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            
            # Construir query base
            query = """
                SELECT h.id, h.user_id, h.tenant_id, h.context_id, h.region_code,
                       h.overall_score, h.confidence_level, h.evaluation_time_ms,
                       h.created_at, h.metadata,
                       COALESCE(json_object_agg(d.dimension, d.score) 
                                FILTER (WHERE d.dimension IS NOT NULL), '{}') as dimension_scores
                FROM trust_score_history h
                LEFT JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                WHERE h.user_id = $1 AND h.tenant_id = $2
            """
            
            # Parâmetros da query
            params = [user_id, tenant_id]
            param_index = 3
            
            # Adicionar filtros opcionais
            if context_id:
                query += f" AND h.context_id = ${param_index}"
                params.append(context_id)
                param_index += 1
            
            if region_code:
                query += f" AND h.region_code = ${param_index}"
                params.append(region_code)
                param_index += 1
            
            if start_date:
                query += f" AND h.created_at >= ${param_index}"
                params.append(start_date)
                param_index += 1
            
            if end_date:
                query += f" AND h.created_at <= ${param_index}"
                params.append(end_date)
                param_index += 1
            
            # Agrupar por registro de histórico
            query += " GROUP BY h.id"
            
            # Ordenar e limitar resultados
            query += " ORDER BY h.created_at DESC"
            query += f" LIMIT ${param_index} OFFSET ${param_index + 1}"
            params.extend([limit, offset])
            
            async with self.pool.acquire() as conn:
                rows = await conn.fetch(query, *params)
                
                results = []
                for row in rows:
                    # Converter dimensões de string para enum
                    dimension_scores = {}
                    if row['dimension_scores']:
                        for dim_str, score in row['dimension_scores'].items():
                            try:
                                dimension = TrustDimension(dim_str)
                                dimension_scores[dimension] = score
                            except ValueError:
                                logger.warning(f"Dimensão desconhecida no histórico: {dim_str}")
                    
                    # Criar objeto de resultado
                    result = TrustScoreResult(
                        user_id=row['user_id'],
                        tenant_id=row['tenant_id'],
                        context_id=row['context_id'],
                        overall_score=row['overall_score'],
                        dimension_scores=dimension_scores,
                        regional_context=row['region_code'],
                        confidence_level=row['confidence_level'],
                        evaluation_time_ms=row['evaluation_time_ms'],
                        timestamp=row['created_at'],
                        metadata=row['metadata'] if row['metadata'] else {}
                    )
                    results.append(result)
                
                span.set_attribute("results_count", len(results))
                return results
    
    async def get_score_details(self, 
                              history_id: int) -> Tuple[TrustScoreResult, 
                                                       Dict[TrustDimension, List[TrustScoreFactorModel]], 
                                                       List[DetectedAnomaly]]:
        """
        Recupera detalhes completos de uma avaliação específica.
        
        Args:
            history_id: ID do registro de histórico
            
        Returns:
            Tuple: Resultado da avaliação, fatores por dimensão e anomalias detectadas
        """
        with tracer.start_as_current_span("get_score_details") as span:
            span.set_attribute("history_id", history_id)
            
            async with self.pool.acquire() as conn:
                # Recuperar registro principal de histórico
                history_row = await conn.fetchrow(
                    """
                    SELECT h.id, h.user_id, h.tenant_id, h.context_id, h.region_code,
                           h.overall_score, h.confidence_level, h.evaluation_time_ms,
                           h.created_at, h.metadata
                    FROM trust_score_history h
                    WHERE h.id = $1
                    """,
                    history_id
                )
                
                if not history_row:
                    raise ValueError(f"Registro de histórico não encontrado: {history_id}")
                
                # Recuperar pontuações por dimensão
                dimension_rows = await conn.fetch(
                    """
                    SELECT dimension, score
                    FROM trust_score_dimension_history
                    WHERE trust_score_history_id = $1
                    """,
                    history_id
                )
                
                dimension_scores = {}
                for row in dimension_rows:
                    try:
                        dimension = TrustDimension(row['dimension'])
                        dimension_scores[dimension] = row['score']
                    except ValueError:
                        logger.warning(f"Dimensão desconhecida: {row['dimension']}")
                
                # Construir objeto de resultado
                result = TrustScoreResult(
                    user_id=history_row['user_id'],
                    tenant_id=history_row['tenant_id'],
                    context_id=history_row['context_id'],
                    overall_score=history_row['overall_score'],
                    dimension_scores=dimension_scores,
                    regional_context=history_row['region_code'],
                    confidence_level=history_row['confidence_level'],
                    evaluation_time_ms=history_row['evaluation_time_ms'],
                    timestamp=history_row['created_at'],
                    metadata=history_row['metadata'] if history_row['metadata'] else {}
                )
                
                # Recuperar fatores
                factor_rows = await conn.fetch(
                    """
                    SELECT factor_id, dimension, factor_name, factor_description,
                           factor_type, weight, value, metadata, created_at
                    FROM trust_score_factors
                    WHERE trust_score_history_id = $1
                    """,
                    history_id
                )
                
                factors: Dict[TrustDimension, List[TrustScoreFactorModel]] = {}
                for row in factor_rows:
                    try:
                        dimension = TrustDimension(row['dimension'])
                        factor_type = FactorType(row['factor_type'])
                        
                        factor = TrustScoreFactorModel(
                            factor_id=row['factor_id'],
                            dimension=dimension,
                            name=row['factor_name'],
                            description=row['factor_description'],
                            type=factor_type,
                            weight=row['weight'],
                            value=row['value'],
                            metadata=row['metadata'] if row['metadata'] else {},
                            regional_context=history_row['region_code'],
                            created_at=row['created_at'],
                            updated_at=row['created_at']
                        )
                        
                        if dimension not in factors:
                            factors[dimension] = []
                        
                        factors[dimension].append(factor)
                    except ValueError as e:
                        logger.warning(f"Erro ao processar fator: {e}")
                
                # Recuperar anomalias
                anomaly_rows = await conn.fetch(
                    """
                    SELECT anomaly_id, anomaly_type, description, severity,
                           confidence, affected_dimensions, metadata, detected_at
                    FROM trust_score_anomalies
                    WHERE trust_score_history_id = $1
                    """,
                    history_id
                )
                
                anomalies: List[DetectedAnomaly] = []
                for row in anomaly_rows:
                    try:
                        anomaly_type = AnomalyType(row['anomaly_type'])
                        severity = AnomalySeverity(row['severity'])
                        
                        # Converter lista de strings de dimensões afetadas para enums
                        affected_dimensions = []
                        for dim_str in row['affected_dimensions']:
                            try:
                                affected_dimensions.append(TrustDimension(dim_str))
                            except ValueError:
                                logger.warning(f"Dimensão desconhecida na anomalia: {dim_str}")
                        
                        anomaly = DetectedAnomaly(
                            anomaly_id=row['anomaly_id'],
                            type=anomaly_type,
                            description=row['description'],
                            severity=severity,
                            confidence=row['confidence'],
                            affected_dimensions=affected_dimensions,
                            metadata=row['metadata'] if row['metadata'] else {},
                            detected_at=row['detected_at']
                        )
                        anomalies.append(anomaly)
                    except ValueError as e:
                        logger.warning(f"Erro ao processar anomalia: {e}")
                
                return result, factors, anomalies

    async def update_user_trust_profile(self, profile: UserTrustProfile) -> bool:
        """
        Atualiza o perfil de confiança de um usuário.
        
        Args:
            profile: Perfil de confiança atualizado
            
        Returns:
            bool: True se atualização foi bem-sucedida
        """
        with tracer.start_as_current_span("update_user_trust_profile") as span:
            span.set_attribute("user_id", profile.user_id)
            span.set_attribute("tenant_id", profile.tenant_id)
            
            async with self.pool.acquire() as conn:
                try:
                    # Verificar se o perfil existe
                    exists = await conn.fetchval(
                        """
                        SELECT EXISTS(
                            SELECT 1 FROM user_trust_profiles 
                            WHERE user_id = $1 AND tenant_id = $2
                        )
                        """,
                        profile.user_id, profile.tenant_id
                    )
                    
                    if exists:
                        # Atualizar perfil existente
                        await conn.execute(
                            """
                            UPDATE user_trust_profiles
                            SET latest_score = $3,
                                trust_score_history = $4,
                                history_summary = $5,
                                updated_at = $6
                            WHERE user_id = $1 AND tenant_id = $2
                            """,
                            profile.user_id,
                            profile.tenant_id,
                            profile.latest_score,
                            [json.dumps(item.__dict__) for item in profile.trust_score_history],
                            json.dumps(profile.history_summary),
                            datetime.now()
                        )
                    else:
                        # Inserir novo perfil
                        await conn.execute(
                            """
                            INSERT INTO user_trust_profiles
                            (user_id, tenant_id, latest_score, trust_score_history, 
                             history_summary, created_at, updated_at)
                            VALUES ($1, $2, $3, $4, $5, $6, $7)
                            """,
                            profile.user_id,
                            profile.tenant_id,
                            profile.latest_score,
                            [json.dumps(item.__dict__) for item in profile.trust_score_history],
                            json.dumps(profile.history_summary),
                            profile.created_at or datetime.now(),
                            datetime.now()
                        )
                    
                    return True
                except Exception as e:
                    logger.error(f"Erro ao atualizar perfil de confiança: {e}")
                    span.record_exception(e)
                    return False
                    
    async def delete_user_history(self, 
                                user_id: str, 
                                tenant_id: str,
                                context_id: Optional[str] = None,
                                older_than: Optional[datetime] = None) -> int:
        """
        Remove registros de histórico de um usuário, opcionalmente filtrando por data.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: Contexto específico (opcional)
            older_than: Remover apenas registros mais antigos que esta data (opcional)
            
        Returns:
            int: Número de registros removidos
        """
        with tracer.start_as_current_span("delete_user_history") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            
            # Construir query base
            query = """
                WITH deleted_ids AS (
                    DELETE FROM trust_score_history 
                    WHERE user_id = $1 AND tenant_id = $2
            """
            
            # Parâmetros da query
            params = [user_id, tenant_id]
            param_index = 3
            
            # Adicionar filtros opcionais
            if context_id:
                query += f" AND context_id = ${param_index}"
                params.append(context_id)
                param_index += 1
            
            if older_than:
                query += f" AND created_at < ${param_index}"
                params.append(older_than)
                param_index += 1
            
            # Finalizar query para retornar IDs removidos
            query += " RETURNING id) SELECT COUNT(*) FROM deleted_ids"
            
            async with self.pool.acquire() as conn:
                async with conn.transaction():
                    # Executar a exclusão e obter contagem
                    count = await conn.fetchval(query, *params)
                    
                    # Se foi removido histórico, atualizar perfil se necessário
                    if count > 0:
                        # Obter IDs de histórico restantes para o usuário
                        remaining_ids = await conn.fetch(
                            """
                            SELECT id, overall_score, created_at FROM trust_score_history
                            WHERE user_id = $1 AND tenant_id = $2
                            ORDER BY created_at DESC
                            LIMIT 20
                            """,
                            user_id, tenant_id
                        )
                        
                        if remaining_ids:
                            # Atualizar o latest_score do perfil se necessário
                            latest_id = remaining_ids[0]['id']
                            latest_score = remaining_ids[0]['overall_score']
                            
                            await conn.execute(
                                """
                                UPDATE user_trust_profiles
                                SET latest_score = $3, updated_at = NOW()
                                WHERE user_id = $1 AND tenant_id = $2
                                """,
                                user_id, tenant_id, latest_score
                            )
                    
                    span.set_attribute("deleted_count", count)
                    return count
    
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
    
    async def get_tenant_statistics(self, 
                                   tenant_id: str,
                                   region_code: Optional[str] = None,
                                   context_id: Optional[str] = None,
                                   dimension: Optional[TrustDimension] = None,
                                   period_start: Optional[datetime] = None,
                                   period_end: Optional[datetime] = None) -> Dict[str, Any]:
        """
        Recupera estatísticas agregadas de pontuação para um tenant.
        
        Args:
            tenant_id: ID do tenant
            region_code: Filtrar por região específica (opcional)
            context_id: Filtrar por contexto específico (opcional)
            dimension: Filtrar por dimensão específica (opcional)
            period_start: Início do período para estatísticas (opcional)
            period_end: Fim do período para estatísticas (opcional)
            
        Returns:
            Dict[str, Any]: Estatísticas calculadas para o tenant
        """
        with tracer.start_as_current_span("get_tenant_statistics") as span:
            span.set_attribute("tenant_id", tenant_id)
            
            # Definir período padrão se não especificado
            if not period_start:
                period_start = datetime.now() - timedelta(days=30)
            
            if not period_end:
                period_end = datetime.now()
            
            # Construir query base para histórico geral
            query_params = [tenant_id, period_start, period_end]
            param_index = 4
            
            base_conditions = "WHERE h.tenant_id = $1 AND h.created_at BETWEEN $2 AND $3"
            
            # Adicionar filtros opcionais
            if region_code:
                base_conditions += f" AND h.region_code = ${param_index}"
                query_params.append(region_code)
                param_index += 1
            
            if context_id:
                base_conditions += f" AND h.context_id = ${param_index}"
                query_params.append(context_id)
                param_index += 1
            
            # Query principal para estatísticas gerais
            query = f"""
                WITH filtered_history AS (
                    SELECT h.id, h.user_id, h.overall_score, h.confidence_level, h.created_at
                    FROM trust_score_history h
                    {base_conditions}
                ),
                user_stats AS (
                    SELECT 
                        user_id,
                        COUNT(*) as evaluation_count,
                        AVG(overall_score) as avg_score,
                        MIN(overall_score) as min_score,
                        MAX(overall_score) as max_score
                    FROM filtered_history
                    GROUP BY user_id
                )
                SELECT 
                    COUNT(DISTINCT fh.user_id) as total_users,
                    COUNT(*) as total_evaluations,
                    AVG(fh.overall_score) as avg_overall_score,
                    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY fh.overall_score) as median_score,
                    MIN(fh.overall_score) as min_overall_score,
                    MAX(fh.overall_score) as max_overall_score,
                    AVG(fh.confidence_level) as avg_confidence,
                    COUNT(*) FILTER (WHERE fh.overall_score < 0.4) as count_low_score,
                    COUNT(*) FILTER (WHERE fh.overall_score BETWEEN 0.4 AND 0.7) as count_medium_score,
                    COUNT(*) FILTER (WHERE fh.overall_score > 0.7) as count_high_score,
                    ARRAY(
                        SELECT json_build_object(
                            'user_id', us.user_id,
                            'evaluation_count', us.evaluation_count,
                            'avg_score', us.avg_score,
                            'min_score', us.min_score,
                            'max_score', us.max_score
                        )
                        FROM user_stats us
                        ORDER BY us.avg_score ASC
                        LIMIT 10
                    ) as lowest_scoring_users,
                    ARRAY(
                        SELECT json_build_object(
                            'date', DATE_TRUNC('day', fh.created_at),
                            'count', COUNT(*),
                            'avg_score', AVG(fh.overall_score)
                        )
                        FROM filtered_history fh
                        GROUP BY DATE_TRUNC('day', fh.created_at)
                        ORDER BY DATE_TRUNC('day', fh.created_at)
                    ) as daily_stats
                FROM filtered_history fh
            """
            
            async with self.pool.acquire() as conn:
                # Obter estatísticas gerais
                general_stats = await conn.fetchrow(query, *query_params)
                
                # Converter resultado para dicionário
                result = dict(general_stats)
                
                # Se dimensão específica solicitada, adicionar estatísticas por dimensão
                if dimension:
                    dim_query = f"""
                        SELECT 
                            AVG(d.score) as avg_score,
                            PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY d.score) as median_score,
                            MIN(d.score) as min_score,
                            MAX(d.score) as max_score,
                            COUNT(*) FILTER (WHERE d.score < 0.4) as count_low_score,
                            COUNT(*) FILTER (WHERE d.score BETWEEN 0.4 AND 0.7) as count_medium_score,
                            COUNT(*) FILTER (WHERE d.score > 0.7) as count_high_score
                        FROM trust_score_history h
                        JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                        {base_conditions} AND d.dimension = ${param_index}
                    """
                    query_params.append(dimension.value)
                    
                    dim_stats = await conn.fetchrow(dim_query, *query_params)
                    result['dimension_stats'] = dict(dim_stats)
                else:
                    # Estatísticas para todas as dimensões
                    all_dim_query = f"""
                        SELECT 
                            d.dimension,
                            AVG(d.score) as avg_score,
                            MIN(d.score) as min_score,
                            MAX(d.score) as max_score,
                            COUNT(*) as count
                        FROM trust_score_history h
                        JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                        {base_conditions}
                        GROUP BY d.dimension
                    """
                    
                    all_dim_rows = await conn.fetch(all_dim_query, *query_params)
                    result['dimensions'] = {row['dimension']: {
                        'avg_score': row['avg_score'],
                        'min_score': row['min_score'],
                        'max_score': row['max_score'],
                        'count': row['count']
                    } for row in all_dim_rows}
                
                # Estatísticas de anomalias
                anomaly_query = f"""
                    SELECT 
                        a.anomaly_type,
                        COUNT(*) as count,
                        AVG(a.confidence) as avg_confidence
                    FROM trust_score_history h
                    JOIN trust_score_anomalies a ON h.id = a.trust_score_history_id
                    {base_conditions}
                    GROUP BY a.anomaly_type
                    ORDER BY COUNT(*) DESC
                """
                
                anomaly_rows = await conn.fetch(anomaly_query, *query_params)
                result['anomalies'] = {row['anomaly_type']: {
                    'count': row['count'],
                    'avg_confidence': row['avg_confidence']
                } for row in anomaly_rows}
                
                return result
    
    async def get_user_trust_profile(self, 
                                    user_id: str, 
                                    tenant_id: str,
                                    context_id: Optional[str] = None) -> UserTrustProfile:
        """
        Recupera ou cria o perfil de confiança de um usuário.
        
        Consolida dados históricos em um perfil que pode ser utilizado
        para futuras avaliações de confiança.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: Contexto específico (opcional)
            
        Returns:
            UserTrustProfile: Perfil de confiança do usuário
        """
        with tracer.start_as_current_span("get_user_trust_profile") as span:
            span.set_attribute("user_id", user_id)
            span.set_attribute("tenant_id", tenant_id)
            
            async with self.pool.acquire() as conn:
                # Verificar se perfil existe
                query = """
                    SELECT 
                        user_id, tenant_id, latest_score, trust_score_history,
                        history_summary, created_at, updated_at
                    FROM user_trust_profiles
                    WHERE user_id = $1 AND tenant_id = $2
                """
                
                row = await conn.fetchrow(query, user_id, tenant_id)
                
                if not row:
                    # Se não existe, criar perfil inicial
                    logger.info(f"Criando perfil de confiança para usuário {user_id}")
                    
                    # Buscar últimas avaliações do usuário (até 10)
                    history_query = """
                        SELECT h.overall_score, h.confidence_level, h.context_id, 
                               h.region_code, h.created_at,
                               COALESCE(json_object_agg(d.dimension, d.score) 
                                FILTER (WHERE d.dimension IS NOT NULL), '{}') as dimension_scores
                        FROM trust_score_history h
                        LEFT JOIN trust_score_dimension_history d ON h.id = d.trust_score_history_id
                        WHERE h.user_id = $1 AND h.tenant_id = $2
                        GROUP BY h.id
                        ORDER BY h.created_at DESC
                        LIMIT 10
                    """
                    
                    history_rows = await conn.fetch(history_query, user_id, tenant_id)
                    
                    # Converter para objetos do modelo
                    history_items = []
                    latest_score = None
                    
                    for idx, h_row in enumerate(history_rows):
                        # Converter dimension_scores
                        dimension_scores = {}
                        if h_row['dimension_scores']:
                            for dim_str, score in h_row['dimension_scores'].items():
                                try:
                                    dimension = TrustDimension(dim_str)
                                    dimension_scores[dimension] = score
                                except ValueError:
                                    continue
                        
                        history_item = TrustScoreHistoryItem(
                            score=h_row['overall_score'],
                            dimension_scores=dimension_scores,
                            confidence_level=h_row['confidence_level'],
                            region_code=h_row['region_code'],
                            context_id=h_row['context_id'],
                            timestamp=h_row['created_at'],
                            anomaly_count=0  # Não temos essa informação aqui
                        )
                        
                        history_items.append(history_item)
                        
                        if idx == 0:
                            latest_score = h_row['overall_score']
                    
                    # Criar perfil vazio se não há histórico
                    if not latest_score:
                        latest_score = 0.5  # Pontuação padrão inicial
                    
                    # Criar e inserir perfil
                    profile = UserTrustProfile(
                        user_id=user_id,
                        tenant_id=tenant_id,
                        latest_score=latest_score,
                        trust_score_history=history_items,
                        history_summary={},
                        created_at=datetime.now(),
                        updated_at=datetime.now()
                    )
                    
                    # Persistir perfil
                    await conn.execute(
                        """
                        INSERT INTO user_trust_profiles
                        (user_id, tenant_id, latest_score, trust_score_history, 
                         history_summary, created_at, updated_at)
                        VALUES ($1, $2, $3, $4, $5, $6, $7)
                        """,
                        profile.user_id,
                        profile.tenant_id,
                        profile.latest_score,
                        [json.dumps(item.__dict__) for item in profile.trust_score_history],
                        json.dumps(profile.history_summary),
                        profile.created_at,
                        profile.updated_at
                    )
                    
                    return profile
                else:
                    # Converter perfil existente
                    history_items = []
                    
                    for item_json in row['trust_score_history']:
                        item_dict = json.loads(item_json)
                        
                        # Converter dimension_scores de string para enum
                        dimension_scores = {}
                        if 'dimension_scores' in item_dict and item_dict['dimension_scores']:
                            for dim_str, score in item_dict['dimension_scores'].items():
                                try:
                                    dimension = TrustDimension(dim_str)
                                    dimension_scores[dimension] = score
                                except ValueError:
                                    continue
                        
                        history_item = TrustScoreHistoryItem(
                            score=item_dict.get('score', 0),
                            dimension_scores=dimension_scores,
                            confidence_level=item_dict.get('confidence_level', 0),
                            region_code=item_dict.get('region_code'),
                            context_id=item_dict.get('context_id'),
                            timestamp=datetime.fromisoformat(item_dict.get('timestamp')) 
                                     if 'timestamp' in item_dict else datetime.now(),
                            anomaly_count=item_dict.get('anomaly_count', 0)
                        )
                        history_items.append(history_item)
                    
                    profile = UserTrustProfile(
                        user_id=row['user_id'],
                        tenant_id=row['tenant_id'],
                        latest_score=row['latest_score'],
                        trust_score_history=history_items,
                        history_summary=row['history_summary'] if row['history_summary'] else {},
                        created_at=row['created_at'],
                        updated_at=row['updated_at']
                    )
                    
                    if context_id and profile.history_summary:
                        # Filtrar dados para o contexto específico se fornecido
                        filtered_history = [
                            item for item in profile.trust_score_history
                            if item.context_id == context_id
                        ]
                        
                        if filtered_history:
                            profile.trust_score_history = filtered_history
                            # Atualizar latest_score para o contexto específico
                            profile.latest_score = filtered_history[0].score
                    
                    return profile