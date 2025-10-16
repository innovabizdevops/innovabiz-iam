"""
Testes de integração para os resolvers GraphQL do sistema de pontuação de confiabilidade.
"""

import os
import json
import unittest
import asyncio
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

# Assumindo que a aplicação GraphQL está usando graphene
import graphene
from graphql.execution.executors.asyncio import AsyncioExecutor

# Importar esquema e tipos GraphQL
from ....api.graphql import schema
from ....app.trust_guard_models import (
    VerificationResponse,
    VerificationStatus,
    ComplianceStatus,
    VerificationDetails
)
from ....app.trust_score_engine import (
    TrustScoreEngine,
    TrustScoreCategory,
    TrustScoreFactor,
    TrustScoreResult
)


class TestTrustScoreGraphQL(unittest.TestCase):
    """Testes para os resolvers GraphQL do sistema de pontuação de confiabilidade."""

    @classmethod
    def setUpClass(cls):
        """Configuração inicial para toda a suíte de testes."""
        # Configurar o schema GraphQL para testes
        cls.schema = schema

        # Mock para o contexto da requisição
        cls.context = MagicMock()
        cls.context.user = {"id": "test_user_001", "roles": ["user"]}
        
        # Mock para o serviço TrustGuard
        cls.mock_trust_service = MagicMock()
        
        # Dados de teste
        cls.test_user_id = "test_user_001"
        cls.test_verification_id = "ver_test_123456"
        cls.timestamp = datetime.utcnow()
        
        # Exemplo de verificação de documento
        cls.document_verification = VerificationResponse(
            verification_id=cls.test_verification_id,
            user_id=cls.test_user_id,
            verification_type="DOCUMENT",
            document_type="ID_CARD",
            status=VerificationStatus.APPROVED,
            score=90.0,
            confidence=92.0,
            timestamp=cls.timestamp,
            compliance_status=ComplianceStatus(
                pep_status=False,
                sanctions_hit=False,
                risk_level="LOW"
            ),
            details=VerificationDetails(
                document_number="ABC123456",
                expiry_date="2028-01-01",
                issuing_country="BR",
                document_fields={"name": "John Doe", "nationality": "Brazilian"},
                liveness_score=0.95,
                face_match_score=0.92,
                watch_list_status="CLEAR"
            )
        )
        
        # Exemplo de resultado de pontuação de confiabilidade
        cls.trust_score_result = TrustScoreResult(
            user_id=cls.test_user_id,
            score=85.0,
            category=TrustScoreCategory.HIGH,
            factor_scores={
                TrustScoreFactor.DOCUMENT_VERIFICATION: 90.0,
                TrustScoreFactor.BIOMETRIC_VERIFICATION: 85.0,
                TrustScoreFactor.COMPLIANCE_CHECK: 80.0
            },
            recommendations=[
                "Adicione mais métodos de verificação para aumentar sua pontuação"
            ],
            verification_ids=[cls.test_verification_id],
            timestamp=cls.timestamp
        )

    def execute_query(self, query, variables=None):
        """Executa uma consulta GraphQL e retorna o resultado."""
        # Usar executor assíncrono para simular o comportamento do servidor
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        
        # Executar a consulta
        result = self.schema.execute(
            query,
            context=self.context,
            variables=variables,
            return_promise=True,
            executor=AsyncioExecutor(loop=loop)
        )
        
        # Processar o resultado
        if isinstance(result, Exception):
            raise result
            
        if hasattr(result, "errors") and result.errors:
            # Imprimir erros para depuração
            for error in result.errors:
                print(f"GraphQL Error: {error}")
            
        # Retornar o resultado
        return result.data

    @patch('app.trust_guard_service.TrustGuardService')
    @patch('app.trust_score_engine.TrustScoreEngine')
    def test_get_trust_score(self, mock_engine_class, mock_service_class):
        """Testa a consulta GraphQL para obter pontuação de confiabilidade."""
        # Configurar mocks
        mock_service = mock_service_class.return_value
        mock_service.get_user_verification_history.return_value = [self.document_verification]
        
        mock_engine = mock_engine_class.return_value
        mock_engine.calculate_score.return_value = self.trust_score_result
        
        # Definir consulta GraphQL
        query = """
        query GetTrustScore($userId: ID!) {
            trustScore(userId: $userId) {
                userId
                score
                category
                factorScores {
                    factor
                    score
                }
                recommendations
                verificationIds
                timestamp
            }
        }
        """
        
        # Executar consulta
        result = self.execute_query(query, variables={"userId": self.test_user_id})
        
        # Verificar resultado
        self.assertIsNotNone(result)
        self.assertIn("trustScore", result)
        trust_score = result["trustScore"]
        
        # Verificar campos obrigatórios
        self.assertEqual(trust_score["userId"], self.test_user_id)
        self.assertEqual(trust_score["score"], 85.0)
        self.assertEqual(trust_score["category"], "HIGH")
        
        # Verificar factor scores
        self.assertTrue(len(trust_score["factorScores"]) >= 3)
        
        # Verificar se o mock foi chamado corretamente
        mock_service.get_user_verification_history.assert_called_once_with(
            self.context, self.test_user_id
        )
        
        # Verificar se o motor de pontuação foi chamado com os parâmetros corretos
        mock_engine.calculate_score.assert_called_once()
        args, kwargs = mock_engine.calculate_score.call_args
        self.assertEqual(kwargs["user_id"], self.test_user_id)
        self.assertEqual(kwargs["verifications"], [self.document_verification])

    @patch('app.trust_guard_service.TrustGuardService')
    def test_get_verification_status(self, mock_service_class):
        """Testa a consulta GraphQL para obter status de verificação."""
        # Configurar mock
        mock_service = mock_service_class.return_value
        mock_service.get_verification_status.return_value = self.document_verification
        
        # Definir consulta GraphQL
        query = """
        query GetVerificationStatus($verificationId: ID!) {
            verificationStatus(verificationId: $verificationId) {
                verificationId
                userId
                verificationType
                documentType
                status
                score
                confidence
                timestamp
                complianceStatus {
                    pepStatus
                    sanctionsHit
                    riskLevel
                }
                details {
                    documentNumber
                    expiryDate
                    issuingCountry
                    documentFields
                    livenessScore
                    faceMatchScore
                    watchListStatus
                }
            }
        }
        """
        
        # Executar consulta
        result = self.execute_query(
            query, 
            variables={"verificationId": self.test_verification_id}
        )
        
        # Verificar resultado
        self.assertIsNotNone(result)
        self.assertIn("verificationStatus", result)
        verification = result["verificationStatus"]
        
        # Verificar campos obrigatórios
        self.assertEqual(verification["verificationId"], self.test_verification_id)
        self.assertEqual(verification["userId"], self.test_user_id)
        self.assertEqual(verification["verificationType"], "DOCUMENT")
        self.assertEqual(verification["status"], "APPROVED")
        
        # Verificar detalhes
        self.assertIsNotNone(verification["details"])
        self.assertEqual(verification["details"]["documentNumber"], "ABC123456")
        
        # Verificar compliance
        self.assertIsNotNone(verification["complianceStatus"])
        self.assertEqual(verification["complianceStatus"]["riskLevel"], "LOW")
        
        # Verificar se o mock foi chamado corretamente
        mock_service.get_verification_status.assert_called_once_with(
            self.context, self.test_verification_id
        )

    @patch('app.trust_guard_service.TrustGuardService')
    def test_initiate_document_verification(self, mock_service_class):
        """Testa a mutação GraphQL para iniciar verificação de documento."""
        # Configurar mock
        mock_service = mock_service_class.return_value
        mock_service.initiate_document_verification.return_value = self.test_verification_id
        
        # Definir mutação GraphQL
        mutation = """
        mutation InitiateDocumentVerification($input: DocumentVerificationInput!) {
            initiateDocumentVerification(input: $input) {
                verificationId
                status
            }
        }
        """
        
        # Dados de entrada
        variables = {
            "input": {
                "userId": self.test_user_id,
                "documentType": "ID_CARD",
                "documentCountry": "BR",
                "documentNumber": "123456789",
                "documentImageFront": "base64_image_front",
                "documentImageBack": "base64_image_back",
                "name": "John Doe",
                "birthdate": "1990-01-01"
            }
        }
        
        # Executar mutação
        result = self.execute_query(mutation, variables=variables)
        
        # Verificar resultado
        self.assertIsNotNone(result)
        self.assertIn("initiateDocumentVerification", result)
        response = result["initiateDocumentVerification"]
        
        # Verificar campos obrigatórios
        self.assertEqual(response["verificationId"], self.test_verification_id)
        self.assertEqual(response["status"], "SUCCESS")
        
        # Verificar se o mock foi chamado corretamente
        mock_service.initiate_document_verification.assert_called_once()
        args, kwargs = mock_service.initiate_document_verification.call_args
        self.assertEqual(args[1], self.test_user_id)
        self.assertEqual(args[2]["document_type"], "ID_CARD")


if __name__ == "__main__":
    unittest.main()