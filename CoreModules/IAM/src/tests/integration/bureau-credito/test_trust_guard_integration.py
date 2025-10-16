"""
Testes de integração para o sistema de pontuação de confiabilidade e TrustGuard.
"""

import os
import json
import unittest
import pytest
from datetime import datetime, timedelta

from ....app.trust_guard_models import (
    VerificationResponse,
    VerificationStatus,
    ComplianceStatus,
    VerificationDetails
)
from ....app.trust_score_engine import (
    TrustScoreEngine,
    TrustScoreCategory,
    TrustScoreFactor
)
from ....app.trust_guard_service import TrustGuardService


@pytest.mark.integration
class TestTrustGuardIntegration(unittest.TestCase):
    """
    Testes de integração para o sistema TrustGuard.
    
    Estes testes requerem um ambiente de API TrustGuard configurado.
    São executados apenas quando variáveis de ambiente apropriadas estão configuradas.
    """

    @classmethod
    def setUpClass(cls):
        """Configuração inicial para toda a suíte de testes."""
        cls.api_url = os.getenv("TRUST_GUARD_API_URL")
        cls.api_key = os.getenv("TRUST_GUARD_API_KEY")
        
        # Pular todos os testes se as credenciais não estiverem configuradas
        if not cls.api_url or not cls.api_key:
            pytest.skip("Credenciais TrustGuard não configuradas. Pulando testes de integração.")
        
        cls.service = TrustGuardService(
            api_url=cls.api_url,
            api_key=cls.api_key
        )
        cls.engine = TrustScoreEngine()
        
        # Criar IDs de teste
        cls.test_user_id = f"test_user_{datetime.utcnow().strftime('%Y%m%d%H%M%S')}"

    def setUp(self):
        """Configuração para cada teste."""
        # Verificar se o serviço está disponível
        try:
            status = self.service.check_status({})
            if not status.get("status") == "available":
                self.skipTest("Serviço TrustGuard não está disponível")
        except Exception as e:
            self.skipTest(f"Erro ao verificar status do TrustGuard: {str(e)}")

    def test_document_verification_flow(self):
        """Testa o fluxo completo de verificação de documento."""
        # Dados simulados para teste
        test_data = {
            "user_id": self.test_user_id,
            "document_type": "ID_CARD",
            "document_country": "BR",
            "document_number": "12345678901",
            "document_image_front": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
            "document_image_back": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
            "name": "John Test",
            "birthdate": "1990-01-01"
        }
        
        # Iniciar verificação de documento
        try:
            verification_id = self.service.initiate_document_verification(
                {}, test_data["user_id"], test_data
            )
            self.assertIsNotNone(verification_id)
            self.assertTrue(len(verification_id) > 5)
            
            # Verificar status inicial (deve estar pendente)
            verification = self.service.get_verification_status({}, verification_id)
            self.assertEqual(verification.user_id, test_data["user_id"])
            self.assertEqual(verification.verification_type, "DOCUMENT")
            
            # Em um ambiente real, aguardaríamos o processamento
            # Aqui, apenas verificamos se o status foi obtido corretamente
            self.assertIn(
                verification.status, 
                [VerificationStatus.PENDING, VerificationStatus.REVIEW_NEEDED]
            )
            
        except Exception as e:
            self.fail(f"Falha no fluxo de verificação de documento: {str(e)}")

    def test_biometric_verification_flow(self):
        """Testa o fluxo completo de verificação biométrica."""
        # Dados simulados para teste
        test_data = {
            "user_id": self.test_user_id,
            "selfie_image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
            "liveness_check": True
        }
        
        # Iniciar verificação biométrica
        try:
            verification_id = self.service.initiate_biometric_verification(
                {}, test_data["user_id"], test_data
            )
            self.assertIsNotNone(verification_id)
            self.assertTrue(len(verification_id) > 5)
            
            # Verificar status inicial (deve estar pendente)
            verification = self.service.get_verification_status({}, verification_id)
            self.assertEqual(verification.user_id, test_data["user_id"])
            self.assertEqual(verification.verification_type, "BIOMETRIC")
            
            # Em um ambiente real, aguardaríamos o processamento
            # Aqui, apenas verificamos se o status foi obtido corretamente
            self.assertIn(
                verification.status, 
                [VerificationStatus.PENDING, VerificationStatus.REVIEW_NEEDED]
            )
            
        except Exception as e:
            self.fail(f"Falha no fluxo de verificação biométrica: {str(e)}")

    def test_trust_score_calculation_with_real_verifications(self):
        """Testa o cálculo de pontuação com verificações reais."""
        # Obter histórico de verificações
        try:
            # Usamos um ID simulado para obter algumas verificações de exemplo
            # Em um ambiente de produção, usaríamos um ID real
            verifications = self.service.get_user_verification_history(
                {}, "demo_user_001", limit=5
            )
            
            if not verifications:
                self.skipTest("Sem verificações disponíveis para teste")
                
            # Calcular pontuação de confiabilidade
            result = self.engine.calculate_score(
                user_id="demo_user_001",
                verifications=verifications
            )
            
            # Verificar se o resultado é válido
            self.assertIsNotNone(result)
            self.assertIsNotNone(result.score)
            self.assertIsNotNone(result.category)
            self.assertGreaterEqual(result.score, 0)
            self.assertLessEqual(result.score, 100)
            self.assertTrue(hasattr(result, 'factor_scores'))
            self.assertTrue(len(result.factor_scores) > 0)
            
        except Exception as e:
            self.fail(f"Falha ao calcular pontuação com verificações reais: {str(e)}")

    def test_verify_and_calculate(self):
        """Testa o fluxo completo: verificação seguida de cálculo de pontuação."""
        # Dados simulados para teste
        test_data = {
            "user_id": self.test_user_id,
            "document_type": "ID_CARD",
            "document_country": "BR",
            "document_number": "12345678901",
            "document_image_front": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
            "document_image_back": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
            "name": "John Test",
            "birthdate": "1990-01-01"
        }
        
        try:
            # Iniciar verificação de documento
            verification_id = self.service.initiate_document_verification(
                {}, test_data["user_id"], test_data
            )
            self.assertIsNotNone(verification_id)
            
            # Obter verificação
            verification = self.service.get_verification_status({}, verification_id)
            
            # Simular histórico de usuário
            user_history = {
                "account_age_days": 30,
                "activities": [{"type": "login", "timestamp": datetime.utcnow().isoformat()}],
                "geolocations": [
                    {"country": "BR", "region": "SP", "timestamp": datetime.utcnow().isoformat()}
                ]
            }
            
            # Calcular pontuação com a verificação
            result = self.engine.calculate_score(
                user_id=test_data["user_id"],
                verifications=[verification],
                user_history=user_history
            )
            
            # Verificar se o resultado é válido
            self.assertIsNotNone(result)
            self.assertEqual(result.user_id, test_data["user_id"])
            
            # Verificar se a verificação foi considerada
            self.assertIn(verification_id, result.verification_ids)
            
        except Exception as e:
            self.fail(f"Falha no fluxo completo de verificação e cálculo: {str(e)}")


if __name__ == "__main__":
    unittest.main()