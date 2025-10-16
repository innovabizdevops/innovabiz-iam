"""
INNOVABIZ - Testes de Validadores Regionais de IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Testes unitários para os validadores regionais de IAM
==================================================================
"""

import unittest
import os
import sys
import json
from pathlib import Path

# Adicionar diretório pai ao path para importação correta
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from validation.compliance_engine import ComplianceEngine, ValidationContext
from validation.compliance_metadata import Region, Industry, ComplianceFramework
from validation.models import ValidationStatus, ValidationSeverity
from validation.validators import get_validators_for_region, get_all_validators


class TestRegionalValidators(unittest.TestCase):
    """Classe de teste para os validadores regionais."""
    
    def setUp(self):
        """Configuração dos testes."""
        self.engine = ComplianceEngine()
        
        # Configuração de teste para simular um sistema IAM configurado
        self.test_config = {
            "consent": {
                "explicit_consent": True,
                "withdrawable": True,
                "clear_information": True,
                "multilingual_support": True,
                "sensitive_data_consent": True,
                "child_consent": True
            },
            "authentication": {
                "identity_assurance_level": "ial2",
                "authentication_assurance_level": "aal2",
                "assurance_level": "substantial",
                "collected_data": ["email", "phone", "user_id"],
                "authentication_factors": [
                    {"type": "knowledge", "name": "password"},
                    {"type": "possession", "name": "mobile_app"}
                ]
            },
            "data_subject_rights": {
                "access": True,
                "rectification": True,
                "erasure": True,
                "portability": True,
                "objection": True,
                "confirmation": True,
                "anonymization": True,
                "correction": True,
                "max_response_time_days": 10
            },
            "security_measures": {
                "technical_measures": [
                    "encryption", "access_control", "logging", "backup", "breach_detection"
                ],
                "breach_notification": True,
                "impact_assessment": True
            },
            "federation": {
                "enabled": True,
                "federation_assurance_level": "fal2",
                "signed_assertions": True,
                "encrypted_assertions": False
            },
            "access_control": {
                "rbac_implemented": True,
                "segregation_of_duties": True,
                "privileged_access_management": True,
                "unique_user_ids": True,
                "emergency_access_procedure": True,
                "auto_logout": True,
                "encryption": True
            },
            "alternative_authentication": {
                "ussd_support": True,
                "offline_authentication": True,
                "adapted_biometrics": True,
                "mobile_money_integration": True
            },
            "inclusive_design": {
                "low_literacy_support": True,
                "accessibility_support": True,
                "cultural_sensitivity": True,
                "local_language_support": True
            },
            "incident_management": {
                "reporting_process": True,
                "max_reporting_time_hours": 24,
                "severity_classification": True
            },
            "operational_resilience": {
                "business_continuity_plan": True,
                "resilience_testing_frequency_months": 6,
                "disaster_recovery_plan": True
            },
            "data_protection_officer": {
                "designated": True,
                "public_contact": True,
                "communication_channel": True
            },
            "ai_systems": {
                "authentication": [
                    {
                        "name": "Behavioral Analysis AI",
                        "risk_category": "high_risk",
                        "risk_assessment": True,
                        "human_oversight": True
                    }
                ]
            },
            "data_transfer": {
                "international_transfer": True,
                "data_protection_authority_approval": True,
                "adequacy_level_check": True,
                "data_transfer_agreement": True
            }
        }
    
    def test_validators_loaded(self):
        """Testa se os validadores foram carregados corretamente."""
        all_validators = get_all_validators()
        self.assertGreater(len(all_validators), 0, "Nenhum validador regional carregado")
        
        # Verificar se há validadores para cada região
        eu_validators = get_validators_for_region(Region.EU)
        brazil_validators = get_validators_for_region(Region.BRAZIL)
        us_validators = get_validators_for_region(Region.USA)
        africa_validators = get_validators_for_region(Region.AFRICA)
        
        self.assertGreater(len(eu_validators), 0, "Nenhum validador para UE carregado")
        self.assertGreater(len(brazil_validators), 0, "Nenhum validador para Brasil carregado")
        self.assertGreater(len(us_validators), 0, "Nenhum validador para EUA carregado")
        self.assertGreater(len(africa_validators), 0, "Nenhum validador para África carregado")
    
    def test_eu_validation(self):
        """Testa a validação de conformidade para UE."""
        report = self.engine.validate(
            tenant_id="test-tenant",
            region=Region.EU,
            industry=Industry.FINANCIAL,
            config=self.test_config
        )
        
        self.assertIsNotNone(report, "Relatório de validação não foi gerado")
        self.assertEqual(report.region, Region.EU.value, "Região do relatório incorreta")
        
        # Verificar se há resultados para frameworks específicos da UE
        framework_results = {result.metadata.get("framework", "") 
                            for result in report.results 
                            if hasattr(result, "metadata") and result.metadata}
        
        expected_frameworks = {"GDPR", "eIDAS 2.0", "EU AI Act 2025", "NIS2", "DORA"}
        found_frameworks = set()
        
        for fw in framework_results:
            for expected in expected_frameworks:
                if expected in fw:
                    found_frameworks.add(expected)
        
        self.assertGreaterEqual(len(found_frameworks), 3, 
                              f"Frameworks insuficientes detectados: {found_frameworks}")
    
    def test_brazil_validation(self):
        """Testa a validação de conformidade para Brasil."""
        report = self.engine.validate(
            tenant_id="test-tenant",
            region=Region.BRAZIL,
            industry=Industry.FINANCIAL,
            config=self.test_config
        )
        
        self.assertIsNotNone(report, "Relatório de validação não foi gerado")
        self.assertEqual(report.region, Region.BRAZIL.value, "Região do relatório incorreta")
        
        # Verificar se há resultados para frameworks específicos do Brasil
        framework_results = {result.metadata.get("framework", "") 
                            for result in report.results 
                            if hasattr(result, "metadata") and result.metadata}
        
        expected_frameworks = {"LGPD", "Open Finance Brazil", "ANPD"}
        found_frameworks = set()
        
        for fw in framework_results:
            for expected in expected_frameworks:
                if expected in fw:
                    found_frameworks.add(expected)
        
        self.assertGreaterEqual(len(found_frameworks), 2, 
                              f"Frameworks insuficientes detectados: {found_frameworks}")
    
    def test_usa_validation(self):
        """Testa a validação de conformidade para EUA."""
        report = self.engine.validate(
            tenant_id="test-tenant",
            region=Region.USA,
            industry=Industry.HEALTHCARE,
            config=self.test_config
        )
        
        self.assertIsNotNone(report, "Relatório de validação não foi gerado")
        self.assertEqual(report.region, Region.USA.value, "Região do relatório incorreta")
        
        # Verificar se há resultados para frameworks específicos dos EUA
        framework_results = {result.metadata.get("framework", "") 
                            for result in report.results 
                            if hasattr(result, "metadata") and result.metadata}
        
        expected_frameworks = {"NIST 800-63-4", "HIPAA"}
        found_frameworks = set()
        
        for fw in framework_results:
            for expected in expected_frameworks:
                if expected in fw:
                    found_frameworks.add(expected)
        
        self.assertGreaterEqual(len(found_frameworks), 1, 
                              f"Frameworks insuficientes detectados: {found_frameworks}")
    
    def test_africa_validation(self):
        """Testa a validação de conformidade para África."""
        report = self.engine.validate(
            tenant_id="test-tenant",
            region=Region.AFRICA,
            industry=Industry.FINANCIAL,
            config=self.test_config
        )
        
        self.assertIsNotNone(report, "Relatório de validação não foi gerado")
        self.assertEqual(report.region, Region.AFRICA.value, "Região do relatório incorreta")
        
        # Verificar se há resultados para frameworks específicos da África
        framework_results = {result.metadata.get("framework", "") 
                            for result in report.results 
                            if hasattr(result, "metadata") and result.metadata}
        
        expected_frameworks = {"Angola Data Protection Law", "Alternative Credentials", "Inclusive Design"}
        found_frameworks = set()
        
        for fw in framework_results:
            for expected in expected_frameworks:
                if expected in fw:
                    found_frameworks.add(expected)
        
        self.assertGreaterEqual(len(found_frameworks), 1, 
                              f"Frameworks insuficientes detectados: {found_frameworks}")


if __name__ == "__main__":
    unittest.main()
