"""
Repositório para gerenciamento do histórico de pontuações de confiança TrustScore.

Este módulo define a interface para persistência e recuperação de dados 
relacionados ao TrustScore, incluindo histórico, dimensões, fatores e anomalias.
Implementa suporte a múltiplos tenants e contextos regionais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from abc import ABC, abstractmethod
from typing import Dict, List, Optional, Tuple, Any, Union
from datetime import datetime, timedelta

# Importar modelos
from ..trust_guard_models import (
    TrustScoreResult,
    TrustScoreFactorModel,
    DetectedAnomaly,
    TrustDimension,
    UserTrustProfile
)


class TrustScoreRepository(ABC):
    """
    Interface para repositório de pontuações de confiança TrustScore.
    
    Define operações padrão para persistência e recuperação de histórico
    de confiança com suporte a múltiplos tenants e contextos.
    """
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
    async def update_user_trust_profile(self, 
                                       profile: UserTrustProfile) -> bool:
        """
        Atualiza o perfil de confiança de um usuário.
        
        Args:
            profile: Perfil de confiança atualizado
            
        Returns:
            bool: True se atualização foi bem-sucedida
        """
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
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
        pass