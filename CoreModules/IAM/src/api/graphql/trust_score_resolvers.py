import json
import logging
from typing import Dict, Any, List, Optional

import graphene
from graphql import GraphQLError
from graphene import ObjectType, Field, String, Float, List as GrapheneList, ID, Int, JSONString

from .common.auth import require_permission, get_current_user
from .common.context import get_trace_context
from ..app.trust_score_engine import TrustScoreEngine, TrustScoreResult, TrustScoreCategory
from ..app.trust_guard_service import TrustGuardService

logger = logging.getLogger(__name__)


class TrustFactorScoreType(ObjectType):
    """Tipo para representar pontuações individuais por fator."""
    factor = String(required=True, description="Nome do fator avaliado")
    score = Float(required=True, description="Pontuação do fator (0-100)")
    weight = Float(description="Peso do fator no cálculo final (%)")
    description = String(description="Descrição do significado da pontuação")


class TrustScoreContextType(ObjectType):
    """Tipo para representar o contexto de uma avaliação de confiabilidade."""
    total_verifications = Int(description="Total de verificações consideradas")
    document_verifications = Int(description="Número de verificações de documento")
    biometric_verifications = Int(description="Número de verificações biométricas")
    additional_data = JSONString(description="Dados contextuais adicionais")


class TrustScoreType(ObjectType):
    """Tipo para representar uma pontuação de confiabilidade."""
    user_id = ID(required=True, description="ID do usuário avaliado")
    score = Float(required=True, description="Pontuação geral de confiabilidade (0-100)")
    category = String(required=True, description="Categoria da pontuação (VERY_HIGH, HIGH, MEDIUM, LOW, VERY_LOW)")
    factor_scores = GrapheneList(TrustFactorScoreType, description="Pontuações por fator")
    recommendations = GrapheneList(String, description="Recomendações baseadas na pontuação")
    timestamp = String(required=True, description="Data e hora da avaliação (ISO format)")
    expires_at = String(description="Data e hora de expiração da avaliação (ISO format)")
    verification_ids = GrapheneList(ID, description="IDs das verificações consideradas")
    context = Field(TrustScoreContextType, description="Contexto da avaliação")


class TrustScoreMutations(ObjectType):
    """Mutações relacionadas ao sistema de pontuação de confiabilidade."""
    
    calculate_trust_score = Field(
        TrustScoreType,
        user_id=String(required=True),
        verification_ids=GrapheneList(String),
        include_user_history=Boolean(default_value=True),
        context_data=JSONString(),
        description="Calcula a pontuação de confiabilidade para um usuário"
    )
    
    @require_permission("iam:trustscore:calculate")
    def resolve_calculate_trust_score(
        self, 
        info, 
        user_id: str, 
        verification_ids: Optional[List[str]] = None,
        include_user_history: bool = True,
        context_data: Optional[str] = None
    ):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            current_user = get_current_user(info.context)
            
            logger.info(
                f"Calculando pontuação de confiabilidade para usuário {user_id} por {current_user['id']}"
            )
            
            # Inicializar serviços necessários
            trust_guard_service = TrustGuardService()
            trust_score_engine = TrustScoreEngine()
            
            # Buscar verificações do usuário
            verifications = []
            if verification_ids:
                # Buscar verificações específicas
                for verification_id in verification_ids:
                    try:
                        verification = trust_guard_service.get_verification_status(ctx, verification_id)
                        verifications.append(verification)
                    except Exception as e:
                        logger.warning(f"Falha ao buscar verificação {verification_id}: {str(e)}")
            else:
                # Buscar histórico de verificações
                try:
                    verifications = trust_guard_service.get_user_verification_history(ctx, user_id, limit=10)
                except Exception as e:
                    logger.warning(f"Falha ao buscar histórico de verificações: {str(e)}")
            
            # Buscar histórico do usuário se solicitado
            user_history = None
            if include_user_history:
                try:
                    # Esta função seria implementada para obter o histórico de atividades do usuário
                    # Omitida aqui para simplicidade
                    user_history = self._get_user_history(ctx, user_id)
                except Exception as e:
                    logger.warning(f"Falha ao buscar histórico do usuário: {str(e)}")
            
            # Processar dados de contexto se fornecidos
            context = {}
            if context_data:
                try:
                    context = json.loads(context_data)
                except Exception as e:
                    logger.warning(f"Falha ao processar dados de contexto: {str(e)}")
            
            # Calcular pontuação
            if not verifications:
                raise GraphQLError(f"Nenhuma verificação encontrada para o usuário {user_id}")
                
            result = trust_score_engine.calculate_score(
                user_id=user_id,
                verifications=verifications,
                user_history=user_history,
                context_data=context
            )
            
            # Mapear resultado para o formato GraphQL
            factor_scores = []
            for factor, score in result.factor_scores.items():
                factor_scores.append({
                    "factor": factor,
                    "score": score,
                    "weight": trust_score_engine.factor_weights.get(factor, 0),
                    "description": self._get_factor_description(factor)
                })
            
            # Preparar dados de contexto
            result_context = {
                "total_verifications": result.context.get("total_verifications", 0),
                "document_verifications": result.context.get("document_verifications", 0),
                "biometric_verifications": result.context.get("biometric_verifications", 0),
                "additional_data": json.dumps(result.context.get("calculation_metadata", {}))
            }
            
            # Registrar cálculo no histórico
            self._record_score_calculation(ctx, result)
            
            logger.info(
                f"Pontuação de confiabilidade calculada para usuário {user_id}: "
                f"{result.score} ({result.category})"
            )
            
            return {
                "user_id": result.user_id,
                "score": result.score,
                "category": result.category,
                "factor_scores": factor_scores,
                "recommendations": result.recommendations,
                "timestamp": result.timestamp.isoformat(),
                "expires_at": result.expires_at.isoformat() if result.expires_at else None,
                "verification_ids": result.verification_ids,
                "context": result_context
            }
            
        except Exception as e:
            logger.error(f"Erro ao calcular pontuação de confiabilidade: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha ao calcular pontuação de confiabilidade: {str(e)}")
    
    def _get_factor_description(self, factor: str) -> str:
        """Retorna a descrição para um fator de pontuação."""
        descriptions = {
            "DOCUMENT_VERIFICATION": "Avaliação de autenticidade e validade de documentos oficiais",
            "BIOMETRIC_VERIFICATION": "Correspondência entre dados biométricos e identidade declarada",
            "COMPLIANCE_CHECK": "Avaliação de conformidade regulatória e status em listas de observação",
            "ACTIVITY_HISTORY": "Análise de padrões históricos de atividade e comportamento",
            "DEVICE_REPUTATION": "Avaliação da reputação e confiabilidade dos dispositivos utilizados",
            "LIVENESS_DETECTION": "Verificação de prova de vida para detectar tentativas de fraude",
            "GEOGRAPHIC_CONSISTENCY": "Análise de consistência geográfica de atividades",
            "BEHAVIORAL_PATTERNS": "Avaliação de padrões comportamentais para detecção de anomalias",
            "WATCHLIST_CHECK": "Verificação contra listas de sanções, PEPs e outras listas de alerta",
            "IDENTITY_LONGEVITY": "Avaliação de idade e estabilidade da identidade digital"
        }
        return descriptions.get(factor, "Fator de pontuação de confiabilidade")
    
    def _get_user_history(self, ctx: Dict[str, Any], user_id: str) -> Dict[str, Any]:
        """
        Obtém o histórico de atividades do usuário.
        Esta seria uma implementação real que consultaria um banco de dados ou serviço.
        """
        # Implementação simulada para fins de exemplo
        return {
            "account_age_days": 180,
            "activities": [{"type": "login", "timestamp": "2023-01-01T12:00:00Z"}] * 15,
            "geolocations": [
                {"country": "BR", "region": "SP", "timestamp": "2023-01-01T12:00:00Z"},
                {"country": "BR", "region": "SP", "timestamp": "2023-01-15T14:30:00Z"},
                {"country": "BR", "region": "RJ", "timestamp": "2023-02-01T10:15:00Z"}
            ]
        }
    
    def _record_score_calculation(self, ctx: Dict[str, Any], result: TrustScoreResult) -> None:
        """
        Registra o cálculo de pontuação no histórico.
        Esta seria uma implementação real que salvaria em um banco de dados.
        """
        # Em uma implementação real, aqui registraríamos o resultado em um banco de dados
        logger.info(f"Registrando cálculo de pontuação: {result.user_id}, score={result.score}")
        # Implementação omitida para simplicidade


class TrustScoreQueries(ObjectType):
    """Queries relacionadas ao sistema de pontuação de confiabilidade."""
    
    get_user_trust_score = Field(
        TrustScoreType,
        user_id=String(required=True),
        description="Obtém a pontuação de confiabilidade atual de um usuário"
    )
    
    get_trust_score_history = GrapheneList(
        TrustScoreType,
        user_id=String(required=True),
        limit=Int(default_value=10),
        offset=Int(default_value=0),
        description="Obtém o histórico de pontuações de confiabilidade de um usuário"
    )
    
    @require_permission("iam:trustscore:read")
    def resolve_get_user_trust_score(self, info, user_id: str):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            
            logger.info(f"Obtendo pontuação de confiabilidade para usuário {user_id}")
            
            # Em uma implementação real, buscaríamos a pontuação mais recente do banco de dados
            # Aqui usamos uma implementação simulada para exemplo
            score_data = self._get_latest_trust_score(ctx, user_id)
            
            if not score_data:
                raise GraphQLError(f"Nenhuma pontuação encontrada para o usuário {user_id}")
                
            # Mapear resultado para o formato GraphQL
            factor_scores = []
            for factor_name, factor_data in score_data.get("factor_scores", {}).items():
                factor_scores.append({
                    "factor": factor_name,
                    "score": factor_data.get("score", 0),
                    "weight": factor_data.get("weight", 0),
                    "description": factor_data.get("description", "")
                })
            
            # Preparar dados de contexto
            context_data = score_data.get("context", {})
            result_context = {
                "total_verifications": context_data.get("total_verifications", 0),
                "document_verifications": context_data.get("document_verifications", 0),
                "biometric_verifications": context_data.get("biometric_verifications", 0),
                "additional_data": json.dumps(context_data.get("additional_data", {}))
            }
            
            return {
                "user_id": score_data.get("user_id"),
                "score": score_data.get("score"),
                "category": score_data.get("category"),
                "factor_scores": factor_scores,
                "recommendations": score_data.get("recommendations", []),
                "timestamp": score_data.get("timestamp"),
                "expires_at": score_data.get("expires_at"),
                "verification_ids": score_data.get("verification_ids", []),
                "context": result_context
            }
            
        except Exception as e:
            logger.error(f"Erro ao obter pontuação de confiabilidade: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha ao obter pontuação de confiabilidade: {str(e)}")
    
    @require_permission("iam:trustscore:read")
    def resolve_get_trust_score_history(self, info, user_id: str, limit: int, offset: int):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            
            logger.info(f"Obtendo histórico de pontuações para usuário {user_id} (limit={limit}, offset={offset})")
            
            # Em uma implementação real, buscaríamos as pontuações do banco de dados
            # Aqui usamos uma implementação simulada para exemplo
            scores_data = self._get_trust_score_history(ctx, user_id, limit, offset)
            
            results = []
            for score_data in scores_data:
                # Mapear resultado para o formato GraphQL
                factor_scores = []
                for factor_name, factor_data in score_data.get("factor_scores", {}).items():
                    factor_scores.append({
                        "factor": factor_name,
                        "score": factor_data.get("score", 0),
                        "weight": factor_data.get("weight", 0),
                        "description": factor_data.get("description", "")
                    })
                
                # Preparar dados de contexto
                context_data = score_data.get("context", {})
                result_context = {
                    "total_verifications": context_data.get("total_verifications", 0),
                    "document_verifications": context_data.get("document_verifications", 0),
                    "biometric_verifications": context_data.get("biometric_verifications", 0),
                    "additional_data": json.dumps(context_data.get("additional_data", {}))
                }
                
                results.append({
                    "user_id": score_data.get("user_id"),
                    "score": score_data.get("score"),
                    "category": score_data.get("category"),
                    "factor_scores": factor_scores,
                    "recommendations": score_data.get("recommendations", []),
                    "timestamp": score_data.get("timestamp"),
                    "expires_at": score_data.get("expires_at"),
                    "verification_ids": score_data.get("verification_ids", []),
                    "context": result_context
                })
            
            return results
            
        except Exception as e:
            logger.error(f"Erro ao obter histórico de pontuações: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha ao obter histórico de pontuações: {str(e)}")
    
    def _get_latest_trust_score(self, ctx: Dict[str, Any], user_id: str) -> Dict[str, Any]:
        """
        Obtém a pontuação de confiabilidade mais recente do usuário.
        Esta seria uma implementação real que consultaria um banco de dados.
        """
        # Implementação simulada para fins de exemplo
        import datetime
        from ..app.trust_score_engine import TrustScoreCategory
        
        return {
            "user_id": user_id,
            "score": 85.7,
            "category": TrustScoreCategory.HIGH,
            "factor_scores": {
                "DOCUMENT_VERIFICATION": {
                    "score": 90.0,
                    "weight": 25,
                    "description": "Avaliação de autenticidade e validade de documentos oficiais"
                },
                "BIOMETRIC_VERIFICATION": {
                    "score": 85.0,
                    "weight": 25,
                    "description": "Correspondência entre dados biométricos e identidade declarada"
                },
                "COMPLIANCE_CHECK": {
                    "score": 95.0,
                    "weight": 15,
                    "description": "Avaliação de conformidade regulatória e status em listas de observação"
                },
                "LIVENESS_DETECTION": {
                    "score": 80.0,
                    "weight": 15,
                    "description": "Verificação de prova de vida para detectar tentativas de fraude"
                },
                "WATCHLIST_CHECK": {
                    "score": 100.0,
                    "weight": 10,
                    "description": "Verificação contra listas de sanções, PEPs e outras listas de alerta"
                },
                "GEOGRAPHIC_CONSISTENCY": {
                    "score": 75.0,
                    "weight": 5,
                    "description": "Análise de consistência geográfica de atividades"
                },
                "ACTIVITY_HISTORY": {
                    "score": 70.0,
                    "weight": 5,
                    "description": "Análise de padrões históricos de atividade e comportamento"
                }
            },
            "recommendations": [
                "Considerar monitoramento de atividades para garantir legitimidade."
            ],
            "timestamp": datetime.datetime.utcnow().isoformat(),
            "expires_at": (datetime.datetime.utcnow() + datetime.timedelta(days=90)).isoformat(),
            "verification_ids": ["ver_12345", "ver_67890"],
            "context": {
                "total_verifications": 2,
                "document_verifications": 1,
                "biometric_verifications": 1,
                "additional_data": {}
            }
        }
    
    def _get_trust_score_history(
        self, 
        ctx: Dict[str, Any], 
        user_id: str, 
        limit: int, 
        offset: int
    ) -> List[Dict[str, Any]]:
        """
        Obtém o histórico de pontuações de confiabilidade do usuário.
        Esta seria uma implementação real que consultaria um banco de dados.
        """
        # Implementação simulada para fins de exemplo
        import datetime
        from ..app.trust_score_engine import TrustScoreCategory
        
        # Criar histórico simulado com pontuações variadas
        base_score = 85.7
        history = []
        
        for i in range(offset, offset + limit):
            days_ago = i * 15  # Cada registro é 15 dias mais antigo que o anterior
            
            # Variar a pontuação ligeiramente
            score_variation = (i % 3 - 1) * 5.0
            score = base_score + score_variation
            
            # Determinar categoria
            category = TrustScoreCategory.HIGH
            if score >= 90:
                category = TrustScoreCategory.VERY_HIGH
            elif score >= 75:
                category = TrustScoreCategory.HIGH
            elif score >= 50:
                category = TrustScoreCategory.MEDIUM
            elif score >= 30:
                category = TrustScoreCategory.LOW
            else:
                category = TrustScoreCategory.VERY_LOW
                
            # Criar registro
            timestamp = (datetime.datetime.utcnow() - datetime.timedelta(days=days_ago)).isoformat()
            expires_at = (datetime.datetime.utcnow() - datetime.timedelta(days=days_ago) + datetime.timedelta(days=90)).isoformat()
            
            history.append({
                "user_id": user_id,
                "score": score,
                "category": category,
                "factor_scores": {
                    "DOCUMENT_VERIFICATION": {
                        "score": 90.0 + score_variation,
                        "weight": 25,
                        "description": "Avaliação de autenticidade e validade de documentos oficiais"
                    },
                    "BIOMETRIC_VERIFICATION": {
                        "score": 85.0 + score_variation,
                        "weight": 25,
                        "description": "Correspondência entre dados biométricos e identidade declarada"
                    },
                    # Outros fatores omitidos para brevidade
                },
                "recommendations": [
                    "Considerar monitoramento de atividades para garantir legitimidade."
                ],
                "timestamp": timestamp,
                "expires_at": expires_at,
                "verification_ids": [f"ver_{12345 + i}", f"ver_{67890 + i}"],
                "context": {
                    "total_verifications": 2,
                    "document_verifications": 1,
                    "biometric_verifications": 1,
                    "additional_data": {}
                }
            })
            
        return history