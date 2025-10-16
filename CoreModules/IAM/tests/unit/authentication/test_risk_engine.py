"""
Testes Unitários para o Motor de Risco Adaptativo

Este módulo contém os testes unitários para o motor de risco adaptativo
do sistema IAM, verificando a correta avaliação de fatores de risco
e o cálculo dos níveis de risco.

@author: Eduardo Jeremias
@date: 08/05/2025
@version: 1.0
"""

import unittest
import unittest.mock as mock
from datetime import datetime
import json
import sys
import os
import pytest

# Ajustar caminho para importação dos módulos
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../../')))

from backend.autoagent.iam.authentication.adaptive_auth.risk_engine import (
    AdaptiveRiskEngine, RiskLevel, RiskFactor, AuthRequirement, get_risk_engine
)


class TestAdaptiveRiskEngine(unittest.TestCase):
    """Testes para o motor de risco adaptativo"""
    
    def setUp(self):
        """Configuração inicial para cada teste"""
        # Criar um mock do objeto de configuração
        config_patcher = mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.get_config')
        mock_config = config_patcher.start()
        
        # Configurar o mock para retornar configurações de teste
        mock_config.return_value = {
            "factor_weights": {
                "location": 0.2,
                "device": 0.15,
                "time": 0.1,
                "behavior": 0.25,
                "ip_reputation": 0.15,
                "previous_auth": 0.05,
                "privileges": 0.05,
                "resource_sensitivity": 0.05
            },
            "risk_thresholds": {
                "low": 0.3,
                "medium": 0.6,
                "high": 0.8,
                "critical": 0.9
            },
            "auth_requirements": {
                "low": ["passwordless"],
                "medium": ["password"],
                "high": ["password", "app_push"],
                "critical": ["password", "hardware_token"]
            },
            "regional_compliance": {
                "EU": {
                    "auth_requirements": {
                        "healthcare": {
                            "default": ["password", "sms_otp"]
                        }
                    }
                },
                "US": {
                    "auth_requirements": {
                        "healthcare": {
                            "default": ["password", "app_push"]
                        }
                    }
                }
            }
        }
        
        # Patches para os repositórios
        self.user_repo_patcher = mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.UserRepository')
        self.session_repo_patcher = mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.SessionRepository')
        self.device_repo_patcher = mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.DeviceRepository')
        
        self.mock_user_repo = self.user_repo_patcher.start()
        self.mock_session_repo = self.session_repo_patcher.start()
        self.mock_device_repo = self.device_repo_patcher.start()
        
        # Inicializar mock do módulo de ML
        self.ml_model_patcher = mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.IsolationForest')
        self.mock_ml_model = self.ml_model_patcher.start()
        
        # Inicializar o motor de risco
        self.risk_engine = AdaptiveRiskEngine()
        
        self.addCleanup(config_patcher.stop)
        self.addCleanup(self.user_repo_patcher.stop)
        self.addCleanup(self.session_repo_patcher.stop)
        self.addCleanup(self.device_repo_patcher.stop)
        self.addCleanup(self.ml_model_patcher.stop)
    
    def test_initialization(self):
        """Testa a inicialização do motor de risco"""
        self.assertIsNotNone(self.risk_engine)
        self.assertIsNotNone(self.risk_engine.factor_weights)
        self.assertIsNotNone(self.risk_engine.risk_thresholds)
        self.assertIsNotNone(self.risk_engine.auth_requirements)
    
    def test_singleton_pattern(self):
        """Testa se o padrão singleton está funcionando corretamente"""
        engine1 = get_risk_engine()
        engine2 = get_risk_engine()
        self.assertIs(engine1, engine2)
    
    def test_determine_risk_level_low(self):
        """Testa a determinação de nível de risco baixo"""
        risk_score = 0.2
        risk_level = self.risk_engine.determine_risk_level(risk_score)
        self.assertEqual(risk_level, RiskLevel.LOW)
    
    def test_determine_risk_level_medium(self):
        """Testa a determinação de nível de risco médio"""
        risk_score = 0.5
        risk_level = self.risk_engine.determine_risk_level(risk_score)
        self.assertEqual(risk_level, RiskLevel.MEDIUM)
    
    def test_determine_risk_level_high(self):
        """Testa a determinação de nível de risco alto"""
        risk_score = 0.7
        risk_level = self.risk_engine.determine_risk_level(risk_score)
        self.assertEqual(risk_level, RiskLevel.HIGH)
    
    def test_determine_risk_level_critical(self):
        """Testa a determinação de nível de risco crítico"""
        risk_score = 0.95
        risk_level = self.risk_engine.determine_risk_level(risk_score)
        self.assertEqual(risk_level, RiskLevel.CRITICAL)
    
    def test_get_auth_requirements_basic(self):
        """Testa a obtenção de requisitos de autenticação básicos"""
        requirements = self.risk_engine.get_auth_requirements(RiskLevel.MEDIUM)
        self.assertEqual(requirements, ["password"])
    
    def test_get_auth_requirements_regional(self):
        """Testa a obtenção de requisitos de autenticação com ajustes regionais"""
        requirements = self.risk_engine.get_auth_requirements(
            RiskLevel.MEDIUM, region="EU", industry="healthcare"
        )
        self.assertEqual(requirements, ["password", "sms_otp"])
    
    @mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.uuid.uuid4')
    def test_evaluate_auth_context(self, mock_uuid):
        """Testa a avaliação completa do contexto de autenticação"""
        mock_uuid.return_value = "session-id-123"
        
        # Configurar mocks para retornar valores baixos de risco
        with mock.patch.object(self.risk_engine, 'calculate_risk_score') as mock_calc:
            mock_calc.return_value = (0.2, {
                "location": 0.1,
                "device": 0.1,
                "time": 0.1
            })
            
            auth_context = {
                "user_id": "user123",
                "ip_address": "192.168.1.1",
                "user_agent": "Mozilla/5.0",
                "timestamp": datetime.now(),
                "region": "US",
                "industry": "healthcare"
            }
            
            result = self.risk_engine.evaluate_auth_context(auth_context)
            
            # Verificar resultado
            self.assertEqual(result["risk_level"], RiskLevel.LOW.value)
            self.assertEqual(result["session_id"], "session-id-123")
            self.assertEqual(result["auth_requirements"], ["password", "app_push"])
    
    def test_calculate_risk_score_missing_data(self):
        """Testa se a exceção correta é lançada quando dados obrigatórios estão faltando"""
        from backend.autoagent.iam.common.exceptions import RiskEngineException
        
        with self.assertRaises(RiskEngineException):
            self.risk_engine.calculate_risk_score({})


@pytest.mark.asyncio
async def test_async_risk_evaluation():
    """Teste assíncrono para avaliação de risco"""
    # Configurar mocks
    with mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.get_config') as mock_config:
        mock_config.return_value = {
            "factor_weights": {"location": 0.5, "device": 0.5},
            "risk_thresholds": {"low": 0.3, "medium": 0.6, "high": 0.8},
            "auth_requirements": {"low": ["passwordless"]}
        }
        
        with mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.UserRepository'):
            with mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.SessionRepository'):
                with mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.DeviceRepository'):
                    with mock.patch('backend.autoagent.iam.authentication.adaptive_auth.risk_engine.IsolationForest'):
                        # Inicializar o motor de risco
                        risk_engine = AdaptiveRiskEngine()
                        
                        # Mockar os métodos de cálculo de risco individuais
                        with mock.patch.object(risk_engine, '_calculate_location_risk', return_value=0.1):
                            with mock.patch.object(risk_engine, '_calculate_device_risk', return_value=0.1):
                                # Calcular score de risco
                                auth_context = {
                                    "user_id": "user123",
                                    "ip_address": "192.168.1.1"
                                }
                                
                                score, factors = risk_engine.calculate_risk_score(auth_context)
                                
                                # Verificar resultado
                                assert score == 0.1
                                assert "location" in factors
                                assert "device" in factors


if __name__ == '__main__':
    unittest.main()
