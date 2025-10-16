import unittest
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

from ...app.trust_guard_models import (
    VerificationResponse,
    VerificationStatus,
    ComplianceStatus,
    VerificationDetails
)
from ...app.trust_score_engine import (
    TrustScoreEngine,
    TrustScoreCategory,
    TrustScoreFactor,
    TrustScoreResult
)


class TestTrustScoreEngine(unittest.TestCase):
    """Testes unitários para o motor de pontuação de confiabilidade."""

    def setUp(self):
        """Configuração inicial para cada teste."""
        self.engine = TrustScoreEngine()
        self.user_id = "user_123"
        self.timestamp = datetime.utcnow()
        
        # Verificação de documento aprovada
        self.document_verification = VerificationResponse(
            verification_id="doc_ver_123",
            user_id=self.user_id,
            verification_type="DOCUMENT",
            document_type="ID_CARD",
            status=VerificationStatus.APPROVED,
            score=90.0,
            confidence=92.0,
            timestamp=self.timestamp,
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
        
        # Verificação biométrica aprovada
        self.biometric_verification = VerificationResponse(
            verification_id="bio_ver_123",
            user_id=self.user_id,
            verification_type="BIOMETRIC",
            status=VerificationStatus.APPROVED,
            score=88.0,
            confidence=91.0,
            timestamp=self.timestamp,
            details=VerificationDetails(
                liveness_score=0.94,
                face_match_score=0.91,
                watch_list_status="CLEAR"
            )
        )
        
        # Verificação de documento rejeitada
        self.rejected_document = VerificationResponse(
            verification_id="doc_ver_456",
            user_id=self.user_id,
            verification_type="DOCUMENT",
            document_type="PASSPORT",
            status=VerificationStatus.REJECTED,
            score=20.0,
            confidence=15.0,
            timestamp=self.timestamp - timedelta(days=5),
            details=VerificationDetails(
                document_number="XYZ789012",
                expiry_date="2025-06-01",
                issuing_country="BR",
                rejection_reason="Document appears to be manipulated",
                watch_list_status="CLEAR"
            )
        )

        # Histórico de usuário simulado
        self.user_history = {
            "account_age_days": 180,
            "activities": [{"type": "login", "timestamp": "2023-01-01T12:00:00Z"}] * 15,
            "geolocations": [
                {"country": "BR", "region": "SP", "timestamp": "2023-01-01T12:00:00Z"},
                {"country": "BR", "region": "SP", "timestamp": "2023-01-15T14:30:00Z"},
                {"country": "BR", "region": "RJ", "timestamp": "2023-02-01T10:15:00Z"}
            ]
        }

    def test_calculate_score_with_approved_verifications(self):
        """Testa o cálculo de pontuação com verificações aprovadas."""
        verifications = [self.document_verification, self.biometric_verification]
        
        result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications,
            user_history=self.user_history
        )
        
        # Verifica se o resultado está no formato correto
        self.assertIsInstance(result, TrustScoreResult)
        self.assertEqual(result.user_id, self.user_id)
        self.assertGreaterEqual(result.score, 0)
        self.assertLessEqual(result.score, 100)
        self.assertIsNotNone(result.category)
        self.assertIsInstance(result.factor_scores, dict)
        self.assertIsInstance(result.recommendations, list)
        
        # Verifica se os fatores esperados foram calculados
        self.assertIn(TrustScoreFactor.DOCUMENT_VERIFICATION, result.factor_scores)
        self.assertIn(TrustScoreFactor.BIOMETRIC_VERIFICATION, result.factor_scores)
        
        # Verifica se a pontuação é alta para verificações aprovadas
        self.assertGreaterEqual(result.score, 75)
        self.assertIn(result.category, [TrustScoreCategory.HIGH, TrustScoreCategory.VERY_HIGH])

    def test_calculate_score_with_rejected_verification(self):
        """Testa o cálculo de pontuação com uma verificação rejeitada."""
        verifications = [self.rejected_document, self.biometric_verification]
        
        result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications,
            user_history=self.user_history
        )
        
        # Verifica se a pontuação é mais baixa quando há uma verificação rejeitada
        self.assertLessEqual(result.score, 70)
        self.assertIn(result.category, [TrustScoreCategory.MEDIUM, TrustScoreCategory.LOW])
        
        # Verifica se há recomendações relacionadas ao documento rejeitado
        has_document_recommendation = any(
            "documento" in rec.lower() for rec in result.recommendations
        )
        self.assertTrue(has_document_recommendation)

    def test_calculate_score_with_minimal_data(self):
        """Testa o cálculo de pontuação com dados mínimos (apenas documento)."""
        verifications = [self.document_verification]
        
        result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications
        )
        
        # Verifica se a pontuação é calculada mesmo com dados limitados
        self.assertIsNotNone(result.score)
        self.assertIsNotNone(result.category)
        
        # A pontuação deve ser média-alta, pois só tem verificação de documento aprovada
        self.assertGreaterEqual(result.score, 65)

    def test_calculate_score_with_high_risk_compliance(self):
        """Testa o cálculo de pontuação com alto risco de compliance."""
        # Modificar a verificação para ter alto risco de compliance
        high_risk_verification = self.document_verification
        high_risk_verification.compliance_status = ComplianceStatus(
            pep_status=True,
            sanctions_hit=True,
            risk_level="HIGH"
        )
        
        verifications = [high_risk_verification]
        
        result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications
        )
        
        # Verifica se a pontuação é mais baixa com alto risco de compliance
        self.assertLessEqual(result.score, 70)
        self.assertIn(
            TrustScoreFactor.COMPLIANCE_CHECK, 
            result.factor_scores
        )
        self.assertLessEqual(
            result.factor_scores[TrustScoreFactor.COMPLIANCE_CHECK], 
            50
        )
        
        # Verifica se há recomendações relacionadas a compliance
        has_compliance_recommendation = any(
            "conformidade" in rec.lower() or "compliance" in rec.lower() 
            for rec in result.recommendations
        )
        self.assertTrue(has_compliance_recommendation)

    def test_determine_category(self):
        """Testa a determinação da categoria com base na pontuação."""
        self.assertEqual(
            self.engine._determine_category(95), 
            TrustScoreCategory.VERY_HIGH
        )
        self.assertEqual(
            self.engine._determine_category(80), 
            TrustScoreCategory.HIGH
        )
        self.assertEqual(
            self.engine._determine_category(60), 
            TrustScoreCategory.MEDIUM
        )
        self.assertEqual(
            self.engine._determine_category(40), 
            TrustScoreCategory.LOW
        )
        self.assertEqual(
            self.engine._determine_category(20), 
            TrustScoreCategory.VERY_LOW
        )

    def test_generate_recommendations(self):
        """Testa a geração de recomendações com base nas pontuações."""
        factor_scores = {
            TrustScoreFactor.DOCUMENT_VERIFICATION: 55.0,
            TrustScoreFactor.BIOMETRIC_VERIFICATION: 45.0,
            TrustScoreFactor.COMPLIANCE_CHECK: 80.0
        }
        
        recommendations = self.engine._generate_recommendations(factor_scores, 60.0)
        
        # Deve haver pelo menos uma recomendação
        self.assertGreater(len(recommendations), 0)
        
        # Deve haver recomendações específicas para os fatores com pontuação baixa
        has_document_recommendation = any(
            "documento" in rec.lower() for rec in recommendations
        )
        has_biometric_recommendation = any(
            "biométric" in rec.lower() for rec in recommendations
        )
        
        self.assertTrue(has_document_recommendation)
        self.assertTrue(has_biometric_recommendation)

    def test_custom_factor_weights(self):
        """Testa o cálculo com pesos personalizados para os fatores."""
        custom_config = {
            "factor_weights": {
                TrustScoreFactor.DOCUMENT_VERIFICATION: 50,
                TrustScoreFactor.BIOMETRIC_VERIFICATION: 50
            }
        }
        
        custom_engine = TrustScoreEngine(config=custom_config)
        
        # Verificar se os pesos são normalizados corretamente
        self.assertEqual(
            custom_engine.factor_weights[TrustScoreFactor.DOCUMENT_VERIFICATION], 
            50.0
        )
        self.assertEqual(
            custom_engine.factor_weights[TrustScoreFactor.BIOMETRIC_VERIFICATION], 
            50.0
        )
        
        # Calcular pontuação com o motor personalizado
        verifications = [self.document_verification, self.biometric_verification]
        
        result = custom_engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications
        )
        
        # A pontuação deve ser diferente da pontuação calculada com pesos padrão
        standard_result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications
        )
        
        # Os resultados podem ser diferentes devido a diferentes pesos
        self.assertIsNotNone(result.score)
        self.assertIsNotNone(standard_result.score)

    def test_geographic_consistency_score(self):
        """Testa o cálculo de consistência geográfica."""
        # Localização atual dentro do padrão histórico
        current_geo = {"country": "BR", "region": "SP"}
        score_consistent = self.engine._calculate_geo_consistency(
            current_geo, self.user_history["geolocations"]
        )
        
        # Localização atual fora do padrão histórico
        current_geo_inconsistent = {"country": "US", "region": "NY"}
        score_inconsistent = self.engine._calculate_geo_consistency(
            current_geo_inconsistent, self.user_history["geolocations"]
        )
        
        # A pontuação deve ser maior quando a localização é consistente
        self.assertGreater(score_consistent, score_inconsistent)
        self.assertGreater(score_consistent, 70.0)
        self.assertLess(score_inconsistent, 60.0)

    def test_result_serialization(self):
        """Testa a serialização e desserialização do resultado."""
        verifications = [self.document_verification]
        
        result = self.engine.calculate_score(
            user_id=self.user_id,
            verifications=verifications
        )
        
        # Converter para dicionário
        result_dict = result.to_dict()
        
        # Converter de volta para objeto
        reconstructed = TrustScoreResult.from_dict(result_dict)
        
        # Verificar se os dados principais são preservados
        self.assertEqual(reconstructed.user_id, result.user_id)
        self.assertEqual(reconstructed.score, result.score)
        self.assertEqual(reconstructed.category, result.category)
        self.assertEqual(len(reconstructed.recommendations), len(result.recommendations))
        self.assertEqual(len(reconstructed.verification_ids), len(result.verification_ids))


if __name__ == "__main__":
    unittest.main()