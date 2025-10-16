"""
Testes unitários para os resolvers GraphQL do TrustScore.

Este módulo implementa testes automatizados para verificar o
correto funcionamento dos resolvers GraphQL do módulo TrustScore.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import pytest
import json
import asyncio
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock, patch

from graphql import GraphQLResolveInfo

from src.api.graphql.trustscore.resolvers import TrustScoreResolvers
from src.app.trust_guard_models import (
    UserTrustProfile, 
    TrustScoreHistoryItem,
    TrustScoreResult
)


@pytest.fixture
def mock_query_service():
    """Fixture para mock do serviço de consulta."""
    service = AsyncMock()
    
    # Mock para timeline de pontuação
    service.get_user_trust_timeline.return_value = {
        "user_id": "user123",
        "tenant_id": "tenant456",
        "latest_score": 0.85,
        "timeline": [
            {
                "date": "2025-08-19",
                "avg_score": 0.83,
                "min_score": 0.79,
                "max_score": 0.87,
                "evaluations": 5
            },
            {
                "date": "2025-08-18",
                "avg_score": 0.81,
                "min_score": 0.77,
                "max_score": 0.86,
                "evaluations": 4
            }
        ],
        "dimensions": {
            "identity": {
                "avg_score": 0.9,
                "min_score": 0.85,
                "max_score": 0.95,
                "variance": 0.001
            },
            "behavioral": {
                "avg_score": 0.82,
                "min_score": 0.75,
                "max_score": 0.88,
                "variance": 0.002
            }
        },
        "anomalies": [
            {
                "anomaly_id": "anom1",
                "type": "unusual_location",
                "description": "Acesso de local não usual",
                "severity": "medium",
                "confidence": 0.75,
                "affected_dimensions": ["identity", "behavioral"],
                "detected_at": "2025-08-19T10:30:00"
            }
        ],
        "profile_summary": {
            "trend": "stable",
            "avg_30d": 0.82,
            "anomalies_30d": 2
        },
        "generated_at": "2025-08-20T08:45:00"
    }
    
    # Mock para comparação regional
    service.get_regional_comparison.return_value = {
        "user": {
            "id": "user123",
            "avg_score": 0.85,
            "evaluation_count": 42,
            "dimensions": {"identity": 0.9, "behavioral": 0.82},
            "percentile": 78.5
        },
        "region": {
            "code": "AO",
            "avg_score": 0.76,
            "median_score": 0.79,
            "user_count": 1250,
            "evaluation_count": 28450,
            "dimensions": {"identity": 0.81, "behavioral": 0.75}
        }
    }
    
    # Mock para detalhes de anomalias
    service.get_anomaly_details.return_value = {
        "user_id": "user123",
        "tenant_id": "tenant456",
        "period": {
            "start": "2025-05-22T08:45:00",
            "end": "2025-08-20T08:45:00",
            "days": 90
        },
        "anomalies": {
            "unusual_location": [
                {
                    "anomaly_id": "anom1",
                    "description": "Acesso de local não usual",
                    "severity": "medium",
                    "confidence": 0.75,
                    "affected_dimensions": ["identity", "behavioral"],
                    "detected_at": "2025-08-19T10:30:00",
                    "overall_score": 0.82,
                    "region_code": "AO",
                    "context_id": "mobile",
                    "metadata": {"location": "Luanda", "device": "new_device"}
                }
            ]
        },
        "statistics": {
            "unusual_location": {
                "count": 1,
                "severity_distribution": {"medium": 1},
                "avg_confidence": 0.75,
                "first_detected": "2025-08-19T10:30:00",
                "last_detected": "2025-08-19T10:30:00"
            }
        },
        "temporal_patterns": {
            "unusual_location": [
                {
                    "date": "2025-08-19",
                    "count": 1
                }
            ]
        },
        "total_anomalies": 1
    }
    
    return service


@pytest.fixture
def mock_repository():
    """Fixture para mock do repositório."""
    repo = AsyncMock()
    
    # Mock para perfil de usuário
    history_items = [
        TrustScoreHistoryItem(
            score=0.85,
            dimension_scores={"identity": 0.9, "behavioral": 0.82},
            confidence_level=0.95,
            region_code="AO",
            context_id="mobile",
            timestamp=datetime.now(),
            anomaly_count=1
        ),
        TrustScoreHistoryItem(
            score=0.83,
            dimension_scores={"identity": 0.88, "behavioral": 0.80},
            confidence_level=0.93,
            region_code="AO",
            context_id="mobile",
            timestamp=datetime.now() - timedelta(days=1),
            anomaly_count=0
        )
    ]
    
    profile = UserTrustProfile(
        user_id="user123",
        tenant_id="tenant456",
        latest_score=0.85,
        trust_score_history=history_items,
        history_summary={
            "trend": "stable",
            "avg_30d": 0.82,
            "anomalies_30d": 2
        },
        created_at=datetime.now() - timedelta(days=30),
        updated_at=datetime.now()
    )
    
    repo.get_user_trust_profile.return_value = profile
    
    # Mock para estatísticas de tenant
    repo.get_tenant_statistics.return_value = {
        "user_count": 1250,
        "evaluation_count": 28450,
        "average_score": 0.76,
        "dimension_stats": {
            "identity": {
                "avg": 0.81,
                "min": 0.65,
                "max": 0.98
            },
            "behavioral": {
                "avg": 0.75,
                "min": 0.60,
                "max": 0.92
            }
        },
        "anomaly_stats": {
            "unusual_location": 145,
            "impossible_travel": 22
        },
        "regional_distribution": {
            "AO": 950,
            "MZ": 300
        }
    }
    
    # Mock para histórico de pontuação
    repo.get_user_trust_history.return_value = {
        "items": [
            {
                "id": "score1",
                "user_id": "user123",
                "tenant_id": "tenant456",
                "context_id": "mobile",
                "overall_score": 0.85,
                "dimension_scores": {"identity": 0.9, "behavioral": 0.82},
                "confidence_level": 0.95,
                "region_code": "AO",
                "created_at": datetime.now(),
                "evaluation_time_ms": 45,
                "metadata": {}
            },
            {
                "id": "score2",
                "user_id": "user123",
                "tenant_id": "tenant456",
                "context_id": "mobile",
                "overall_score": 0.83,
                "dimension_scores": {"identity": 0.88, "behavioral": 0.80},
                "confidence_level": 0.93,
                "region_code": "AO",
                "created_at": datetime.now() - timedelta(days=1),
                "evaluation_time_ms": 42,
                "metadata": {}
            }
        ],
        "total_count": 2,
        "include_details": True
    }
    
    # Mock para detalhes de pontuação
    repo.get_score_details.return_value = {
        "id": "score1",
        "user_id": "user123",
        "tenant_id": "tenant456",
        "context_id": "mobile",
        "overall_score": 0.85,
        "dimension_scores": {"identity": 0.9, "behavioral": 0.82},
        "confidence_level": 0.95,
        "region_code": "AO",
        "created_at": datetime.now(),
        "evaluation_time_ms": 45,
        "factors": [
            {
                "factor_id": "factor1",
                "dimension": "identity",
                "name": "Histórico de autenticação estável",
                "description": "Padrão consistente de autenticação",
                "type": "positive",
                "weight": 0.7,
                "value": 0.8
            }
        ],
        "anomalies": [],
        "metadata": {}
    }
    
    return repo


@pytest.fixture
def mock_info():
    """Fixture para mock de informações de contexto GraphQL."""
    info = MagicMock(spec=GraphQLResolveInfo)
    
    # Mock para contexto de requisição com usuário autenticado
    context = MagicMock()
    context.user = MagicMock()
    context.user.has_tenant_access.return_value = True
    context.user.has_permission.return_value = True
    
    info.context = context
    return info


class TestTrustScoreResolvers:
    """Testes para os resolvers GraphQL do TrustScore."""
    
    @pytest.mark.asyncio
    async def test_get_user_trust_profile(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para obter perfil de confiança."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_user_trust_profile(
            None,  # obj
            mock_info,
            user_id="user123",
            tenant_id="tenant456",
            context_id="mobile"
        )
        
        # Verificar resultado
        assert result["userId"] == "user123"
        assert result["tenantId"] == "tenant456"
        assert result["latestScore"] == 0.85
        assert len(result["trustScoreHistory"]) == 2
        assert isinstance(result["historySummary"], dict)
        
        # Verificar chamadas do repositório
        mock_repository.get_user_trust_profile.assert_called_once_with(
            user_id="user123",
            tenant_id="tenant456",
            context_id="mobile"
        )
    
    @pytest.mark.asyncio
    async def test_get_trust_score_timeline(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para obter linha do tempo de pontuações."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_trust_score_timeline(
            None,  # obj
            mock_info,
            user_id="user123",
            tenant_id="tenant456",
            days=30,
            context_id="mobile",
            include_anomalies=True
        )
        
        # Verificar resultado
        assert result["userId"] == "user123"
        assert result["tenantId"] == "tenant456"
        assert result["latestScore"] == 0.85
        assert len(result["timeline"]) == 2
        assert "dimensions" in result
        assert len(result["anomalies"]) == 1
        assert "profileSummary" in result
        assert "generatedAt" in result
        
        # Verificar chamadas do serviço
        mock_query_service.get_user_trust_timeline.assert_called_once_with(
            user_id="user123",
            tenant_id="tenant456",
            days=30,
            context_id="mobile",
            include_anomalies=True
        )
    
    @pytest.mark.asyncio
    async def test_get_regional_comparison(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para obter comparação regional."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_regional_comparison(
            None,  # obj
            mock_info,
            user_id="user123",
            tenant_id="tenant456",
            region_code="AO",
            context_id="mobile"
        )
        
        # Verificar resultado
        assert result["user"]["id"] == "user123"
        assert result["user"]["avgScore"] == 0.85
        assert result["user"]["percentile"] == 78.5
        assert result["region"]["code"] == "AO"
        assert result["region"]["avgScore"] == 0.76
        assert result["region"]["userCount"] == 1250
        
        # Verificar chamadas do serviço
        mock_query_service.get_regional_comparison.assert_called_once_with(
            user_id="user123",
            tenant_id="tenant456",
            region_code="AO",
            context_id="mobile"
        )
    
    @pytest.mark.asyncio
    async def test_get_anomaly_details(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para obter detalhes de anomalias."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_anomaly_details(
            None,  # obj
            mock_info,
            user_id="user123",
            tenant_id="tenant456",
            days=90,
            anomaly_types=["unusual_location"],
            min_severity="medium",
            context_id="mobile"
        )
        
        # Verificar resultado
        assert result["userId"] == "user123"
        assert result["tenantId"] == "tenant456"
        assert result["period"]["days"] == 90
        assert "unusual_location" in result["anomalies"]
        assert result["statistics"]["unusual_location"]["count"] == 1
        assert result["totalAnomalies"] == 1
        
        # Verificar chamadas do serviço
        mock_query_service.get_anomaly_details.assert_called_once_with(
            user_id="user123",
            tenant_id="tenant456",
            days=90,
            anomaly_types=["unusual_location"],
            min_severity="medium",
            context_id="mobile"
        )
    
    @pytest.mark.asyncio
    async def test_get_tenant_statistics(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para obter estatísticas de tenant."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_tenant_statistics(
            None,  # obj
            mock_info,
            tenant_id="tenant456",
            days=30,
            region_code="AO",
            context_id="mobile"
        )
        
        # Verificar resultado
        assert result["tenantId"] == "tenant456"
        assert result["userCount"] == 1250
        assert result["evaluationCount"] == 28450
        assert result["averageScore"] == 0.76
        assert "dimensionStats" in result
        assert "anomalyStats" in result
        assert "regionalDistribution" in result
        assert result["periodSummary"]["days"] == 30
        
        # Verificar chamadas do repositório
        mock_repository.get_tenant_statistics.assert_called_once()
        call_args = mock_repository.get_tenant_statistics.call_args[1]
        assert call_args["tenant_id"] == "tenant456"
        assert call_args["region_code"] == "AO"
        assert call_args["context_id"] == "mobile"
    
    @pytest.mark.asyncio
    async def test_get_trust_score_results(self, mock_repository, mock_query_service, mock_info):
        """Testa o resolver para listar resultados de pontuação paginados."""
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Chamar resolver
        result = await resolvers.get_trust_score_results(
            None,  # obj
            mock_info,
            user_id="user123",
            tenant_id="tenant456",
            context_id="mobile",
            limit=10,
            offset=0
        )
        
        # Verificar resultado
        assert len(result["edges"]) == 2
        assert result["totalCount"] == 2
        assert result["pageInfo"]["hasNextPage"] == False
        assert result["pageInfo"]["hasPreviousPage"] == False
        
        # Verificar estrutura de um edge
        edge = result["edges"][0]
        assert "node" in edge
        assert "cursor" in edge
        assert edge["node"]["userId"] == "user123"
        assert edge["node"]["tenantId"] == "tenant456"
        assert edge["node"]["overallScore"] == 0.85
        
        # Verificar chamadas do repositório
        mock_repository.get_user_trust_history.assert_called_once()
        call_args = mock_repository.get_user_trust_history.call_args[1]
        assert call_args["user_id"] == "user123"
        assert call_args["tenant_id"] == "tenant456"
        assert call_args["context_id"] == "mobile"
        assert call_args["limit"] == 10
        assert call_args["offset"] == 0
    
    @pytest.mark.asyncio
    async def test_unauthorized_access(self, mock_repository, mock_query_service, mock_info):
        """Testa rejeição de acesso não autorizado."""
        # Configurar comportamento de usuário sem acesso
        mock_info.context.user.has_tenant_access.return_value = False
        
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Verificar exceção ao tentar acessar dados
        with pytest.raises(Exception) as excinfo:
            await resolvers.get_user_trust_profile(
                None,
                mock_info,
                user_id="user123",
                tenant_id="tenant456"
            )
        
        assert "não tem acesso ao tenant" in str(excinfo.value)
    
    @pytest.mark.asyncio
    async def test_unauthorized_permission(self, mock_repository, mock_query_service, mock_info):
        """Testa rejeição por falta de permissão específica."""
        # Configurar comportamento de usuário sem permissão
        mock_info.context.user.has_tenant_access.return_value = True
        mock_info.context.user.has_permission.return_value = False
        
        # Configurar resolvers
        resolvers = TrustScoreResolvers(mock_query_service, mock_repository)
        
        # Verificar exceção ao tentar acessar dados sem permissão
        with pytest.raises(Exception) as excinfo:
            await resolvers.get_tenant_statistics(
                None,
                mock_info,
                tenant_id="tenant456"
            )
        
        assert "não tem permissão" in str(excinfo.value)