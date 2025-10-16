"""
Implementação de métodos adicionais para o repositório TrustScore PostgreSQL.

Este arquivo contém os métodos restantes que serão combinados com o repositório
principal para completar a implementação do TrustScore.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta

from opentelemetry import trace

# Importar modelos
from ..trust_guard_models import (
    UserTrustProfile,
    TrustDimension
)

# Configuração de tracing
tracer = trace.get_tracer(__name__)
logger = logging.getLogger(__name__)


# Métodos adicionais para TrustScorePostgresRepository
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