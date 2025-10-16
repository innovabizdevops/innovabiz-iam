"""
Testes Unitários para os Validadores de Conformidade em Saúde

Este módulo contém os testes unitários para os validadores de conformidade
regulatória para o setor de saúde (HIPAA, GDPR, LGPD) do sistema IAM.

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

from backend.autoagent.iam.compliance.validators.base_validator import (
    ComplianceStatus, ComplianceSeverity, ComplianceControl, ComplianceIssue, ControlAssessment
)
from backend.autoagent.iam.compliance.validators.healthcare_validators import HealthcareComplianceValidator
from backend.autoagent.iam.compliance.validators.healthcare_validators_consent import assess_consent_management
from backend.autoagent.iam.compliance.validators.healthcare_validators_breach import assess_breach_notification


class TestHealthcareComplianceValidator(unittest.TestCase):
    """Testes para o validador de conformidade base para saúde"""
    
    def setUp(self):
        """Configuração inicial para cada teste"""
        # Mockear o carregamento de controles
        with mock.patch.object(HealthcareComplianceValidator, '_load_controls') as mock_load:
            mock_load.return_value = {}
            self.validator = HealthcareComplianceValidator()
    
    def test_initialization(self):
        """Testa a inicialização do validador"""
        self.assertIsNotNone(self.validator)
        self.assertIsNotNone(self.validator.healthcare_categories)
        # Verificar categorias específicas
        self.assertIn("consent_management", self.validator.healthcare_categories)
        self.assertIn("breach_notification", self.validator.healthcare_categories)
    
    def test_assess_control_routing(self):
        """Testa o roteamento de avaliação de controle baseado na categoria"""
        # Criar um controle mock
        control = mock.MagicMock(spec=ComplianceControl)
        control.category = "authentication"
        
        # Criar um mock para o método específico de categoria
        with mock.patch.object(self.validator, '_assess_authentication') as mock_method:
            mock_method.return_value = ControlAssessment(
                control=control,
                status=ComplianceStatus.COMPLIANT,
                details="Test assessment",
                issues=[]
            )
            
            # Chamar o método de avaliação
            config = {"test": "config"}
            context = {"test": "context"}
            
            # Executar o método de forma síncrona para o teste
            import asyncio
            assessment = asyncio.run(self.validator._assess_control(control, config, context))
            
            # Verificar se o método específico foi chamado
            mock_method.assert_called_once_with(control, config, context)
            
            # Verificar resultado
            self.assertEqual(assessment.status, ComplianceStatus.COMPLIANT)
            self.assertEqual(assessment.details, "Test assessment")


@pytest.mark.asyncio
async def test_assess_authentication():
    """Teste para o método de avaliação de autenticação"""
    # Criar uma instância de validador (com mocks)
    with mock.patch.object(HealthcareComplianceValidator, '_load_controls') as mock_load:
        mock_load.return_value = {}
        validator = HealthcareComplianceValidator()
        
        # Criar controle para teste
        control = ComplianceControl(
            id="test-auth-001",
            name="Test Authentication Control",
            description="Test control for authentication",
            regulation="Test",
            section="Test Section",
            category="authentication",
            implementation_requirements=["Test requirement"],
            verification_method="Test"
        )
        
        # Configuração sem autenticação adequada
        config_non_compliant = {
            "auth_methods": {
                "mfa": {
                    "enabled": False
                },
                "password": {
                    "policy": {
                        "min_length": 8
                    }
                }
            }
        }
        
        # Avaliar com configuração não conforme
        assessment = await validator._assess_authentication(control, config_non_compliant, {})
        
        # Verificar resultado
        assert assessment.status == ComplianceStatus.NON_COMPLIANT
        assert len(assessment.issues) >= 1
        
        # Configuração conforme
        config_compliant = {
            "auth_methods": {
                "mfa": {
                    "enabled": True
                },
                "password": {
                    "policy": {
                        "min_length": 14
                    }
                },
                "lockout_policy": {
                    "enabled": True
                },
                "adaptive_auth": {
                    "enabled": True
                }
            }
        }
        
        # Avaliar com configuração conforme
        assessment = await validator._assess_authentication(control, config_compliant, {})
        
        # Verificar resultado
        assert assessment.status == ComplianceStatus.COMPLIANT
        assert len(assessment.issues) == 0


@pytest.mark.asyncio
async def test_consent_management():
    """Teste para o método de avaliação de gestão de consentimento"""
    # Criar objeto validador mock para receber o método
    validator_mock = mock.MagicMock()
    
    # Criar controle para teste
    control = ComplianceControl(
        id="test-consent-001",
        name="Test Consent Control",
        description="Test control for consent management",
        regulation="Test",
        section="Test Section",
        category="consent_management",
        implementation_requirements=["Test requirement"],
        verification_method="Test"
    )
    
    # Configuração sem gestão de consentimento
    config_non_compliant = {
        "consent_management": {
            "enabled": False
        }
    }
    
    # Avaliar com configuração não conforme
    assessment = await assess_consent_management(validator_mock, control, config_non_compliant, {})
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.NON_COMPLIANT
    assert len(assessment.issues) == 1
    
    # Configuração parcialmente conforme
    config_partial = {
        "consent_management": {
            "enabled": True,
            "granular_consent": False,
            "revocation": {
                "enabled": True
            },
            "audit_trail": {
                "enabled": False
            },
            "expiration": {
                "enabled": False
            },
            "evidence": {
                "enabled": True
            }
        }
    }
    
    # Avaliar com configuração parcialmente conforme
    assessment = await assess_consent_management(validator_mock, control, config_partial, {})
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.PARTIALLY_COMPLIANT
    assert len(assessment.issues) >= 1
    
    # Configuração totalmente conforme
    config_compliant = {
        "consent_management": {
            "enabled": True,
            "granular_consent": True,
            "revocation": {
                "enabled": True
            },
            "audit_trail": {
                "enabled": True
            },
            "expiration": {
                "enabled": True
            },
            "evidence": {
                "enabled": True
            }
        }
    }
    
    # Avaliar com configuração conforme
    assessment = await assess_consent_management(validator_mock, control, config_compliant, {})
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.COMPLIANT
    assert len(assessment.issues) == 0


@pytest.mark.asyncio
async def test_breach_notification():
    """Teste para o método de avaliação de notificação de violação"""
    # Criar objeto validador mock para receber o método
    validator_mock = mock.MagicMock()
    
    # Criar controle para teste
    control = ComplianceControl(
        id="test-breach-001",
        name="Test Breach Notification Control",
        description="Test control for breach notification",
        regulation="Test",
        section="Test Section",
        category="breach_notification",
        implementation_requirements=["Test requirement"],
        verification_method="Test"
    )
    
    # Configuração não conforme
    config_non_compliant = {
        "breach_management": {
            "enabled": False
        }
    }
    
    # Contexto regional - UE (72 horas)
    context_eu = {
        "region": "EU"
    }
    
    # Avaliar com configuração não conforme
    assessment = await assess_breach_notification(validator_mock, control, config_non_compliant, context_eu)
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.NON_COMPLIANT
    assert len(assessment.issues) == 1
    
    # Configuração parcialmente conforme
    config_partial = {
        "breach_management": {
            "enabled": True,
            "detection": {
                "enabled": True
            },
            "notification": {
                "timeframe_hours": 96,  # Acima do limite da UE (72 horas)
                "authorities": False,
                "affected_individuals": True
            },
            "classification": {
                "enabled": False
            },
            "tracking": {
                "enabled": True
            }
        }
    }
    
    # Avaliar com configuração parcialmente conforme
    assessment = await assess_breach_notification(validator_mock, control, config_partial, context_eu)
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.PARTIALLY_COMPLIANT
    assert len(assessment.issues) >= 1
    
    # Configuração conforme para Brasil
    config_compliant_br = {
        "breach_management": {
            "enabled": True,
            "detection": {
                "enabled": True
            },
            "notification": {
                "timeframe_hours": 48,  # Dentro do limite para Brasil
                "authorities": True,
                "affected_individuals": True
            },
            "classification": {
                "enabled": True
            },
            "tracking": {
                "enabled": True
            }
        }
    }
    
    # Contexto regional - Brasil (48 horas)
    context_br = {
        "region": "BR"
    }
    
    # Avaliar com configuração conforme para Brasil
    assessment = await assess_breach_notification(validator_mock, control, config_compliant_br, context_br)
    
    # Verificar resultado
    assert assessment.status == ComplianceStatus.COMPLIANT
    assert len(assessment.issues) == 0


if __name__ == '__main__':
    unittest.main()
