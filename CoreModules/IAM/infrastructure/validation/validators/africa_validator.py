"""
INNOVABIZ - Validador de Conformidade para África (Angola)
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de validador de conformidade para África (Angola),
           incluindo Lei de Proteção de Dados de Angola e regulamentos aplicáveis.
==================================================================
"""

import logging
from typing import Dict, List, Any, Optional, Set
from dataclasses import dataclass

from ..compliance_metadata import (
    ComplianceMetadataRegistry,
    Region,
    Industry,
    ComplianceFramework
)

from ..models import (
    ComplianceValidationResult,
    ValidationSeverity,
    ValidationStatus
)

from ..compliance_engine import ValidationContext, ValidationRule

# Configuração de logging
logger = logging.getLogger(__name__)

class AngolaDataProtectionBaseRule(ValidationRule):
    """Regra base para validação da Lei de Proteção de Dados de Angola"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.AFRICA, Region.ANGOLA, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("angola_data_protection")]

class AlternativeCredentialsRule(ValidationRule):
    """Regra base para validação de credenciais alternativas para mercados emergentes"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.MEDIUM
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.AFRICA, Region.ANGOLA, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("alternative_credentials")]

# Implementação de regras específicas para Lei de Proteção de Dados de Angola
class AngolaConsentRule(AngolaDataProtectionBaseRule):
    """Validação de consentimento conforme Lei de Proteção de Dados de Angola"""
    
    def __init__(self):
        super().__init__(
            "angola_consent",
            "Angola Data Protection Law Consent Requirements",
            "Verifies that authentication processes obtain proper consent for personal data processing",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["adp-art6", "adp-art7", "adp-art8"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de consent
        consent_config = context.config.get("consent", {})
        
        # Verificar se o consentimento é explícito
        explicit_consent = consent_config.get("explicit_consent", False)
        if not explicit_consent:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art7"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Authentication does not require explicit consent for data processing",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 7",
                        "recommendation": "Implement explicit consent mechanism before authentication"
                    }
                )
            )
        
        # Verificar se é possível retirar o consentimento
        withdrawable = consent_config.get("withdrawable", False)
        if not withdrawable:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art7"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No mechanism to withdraw consent",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 7",
                        "recommendation": "Implement consent withdrawal mechanism"
                    }
                )
            )
        
        # Verificar consentimento multilíngue (português e línguas locais)
        multilingual = consent_config.get("multilingual_support", False)
        if not multilingual:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art8"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No support for multilingual consent",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 8",
                        "recommendation": "Implement multilingual consent forms (Portuguese and local languages)"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if not results:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Authentication properly implements Angola Data Protection Law consent requirements",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "articles": "Articles 6, 7, 8"
                    }
                )
            )
        
        return results

class AngolaDataTransferRule(AngolaDataProtectionBaseRule):
    """Validação de transferência internacional de dados conforme Lei de Proteção de Dados de Angola"""
    
    def __init__(self):
        super().__init__(
            "angola_data_transfer",
            "Angola Data Protection Law International Transfers",
            "Verifies compliance with requirements for international data transfers",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["adp-art30", "adp-art31", "adp-art32"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de transferência de dados
        transfer_config = context.config.get("data_transfer", {})
        
        # Verificar se há transferência internacional
        international_transfer = transfer_config.get("international_transfer", False)
        
        if not international_transfer:
            # Se não há transferência internacional, esta regra não se aplica diretamente
            return [
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="No international data transfers configured, requirements not applicable",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "articles": "Articles 30, 31, 32"
                    }
                )
            ]
        
        # Verificar aprovação da autoridade de proteção de dados
        dpa_approval = transfer_config.get("data_protection_authority_approval", False)
        if not dpa_approval:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art31"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No Data Protection Authority approval for international transfers",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 31",
                        "recommendation": "Obtain approval from Angola's Data Protection Authority"
                    }
                )
            )
        
        # Verificar verificação de nível adequado de proteção
        adequacy_check = transfer_config.get("adequacy_level_check", False)
        if not adequacy_check:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art30"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No adequacy level check for destination countries",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 30",
                        "recommendation": "Implement adequacy level verification for destination countries"
                    }
                )
            )
        
        # Verificar se há contrato de transferência de dados
        data_transfer_agreement = transfer_config.get("data_transfer_agreement", False)
        if not data_transfer_agreement:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art32"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No data transfer agreement implemented",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 32",
                        "recommendation": "Implement data transfer agreements for international transfers"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="International data transfers comply with Angola Data Protection Law",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "articles": "Articles 30, 31, 32"
                    }
                )
            )
        
        return results

class AngolaDataSubjectRightsRule(AngolaDataProtectionBaseRule):
    """Validação de direitos dos titulares conforme Lei de Proteção de Dados de Angola"""
    
    def __init__(self):
        super().__init__(
            "angola_data_subject_rights",
            "Angola Data Protection Law Data Subject Rights",
            "Verifies compliance with data subject rights requirements",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["adp-art14", "adp-art15", "adp-art16", "adp-art17"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de direitos do titular
        rights_config = context.config.get("data_subject_rights", {})
        
        # Lista de direitos essenciais
        essential_rights = {
            "access": "adp-art14",
            "rectification": "adp-art15",
            "erasure": "adp-art16",
            "objection": "adp-art17"
        }
        
        # Verificar cada direito
        for right, requirement in essential_rights.items():
            implemented = rights_config.get(right, False)
            if not implemented:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        requirement_ids=[requirement],
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.HIGH,
                        message=f"Right to {right} not implemented",
                        metadata={
                            "framework": "Angola Data Protection Law",
                            "article": f"Article {requirement.split('-')[-1]}",
                            "recommendation": f"Implement mechanism for data subject's right to {right}"
                        }
                    )
                )
        
        # Verificar interface multilíngue para exercício de direitos
        multilingual_interface = rights_config.get("multilingual_interface", False)
        if not multilingual_interface:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["adp-art14"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No multilingual interface for exercising rights",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "article": "Article 14",
                        "recommendation": "Implement multilingual interface for rights exercise (Portuguese and local languages)"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="All essential data subject rights are implemented",
                    metadata={
                        "framework": "Angola Data Protection Law",
                        "articles": "Articles 14, 15, 16, 17"
                    }
                )
            )
        
        return results

class AlternativeAuthenticationRule(AlternativeCredentialsRule):
    """Validação de autenticação alternativa para mercados emergentes"""
    
    def __init__(self):
        super().__init__(
            "alternative_authentication",
            "Alternative Authentication Methods",
            "Verifies support for alternative authentication methods suitable for emerging markets",
            ValidationSeverity.MEDIUM
        )
    
    def get_requirements(self) -> List[str]:
        return ["alt-auth-mobile", "alt-auth-biometric", "alt-auth-offline"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de autenticação alternativa
        auth_config = context.config.get("alternative_authentication", {})
        
        # Verificar autenticação por USSD (comum em África)
        ussd_support = auth_config.get("ussd_support", False)
        if not ussd_support:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["alt-auth-mobile"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No USSD authentication support",
                    metadata={
                        "framework": "Alternative Credentials",
                        "requirement": "Mobile Authentication",
                        "recommendation": "Implement USSD-based authentication for feature phones"
                    }
                )
            )
        
        # Verificar autenticação offline
        offline_support = auth_config.get("offline_authentication", False)
        if not offline_support:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["alt-auth-offline"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No offline authentication support",
                    metadata={
                        "framework": "Alternative Credentials",
                        "requirement": "Offline Authentication",
                        "recommendation": "Implement offline authentication for areas with limited connectivity"
                    }
                )
            )
        
        # Verificar biometria adaptada para diversidade
        adapted_biometrics = auth_config.get("adapted_biometrics", False)
        if not adapted_biometrics:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["alt-auth-biometric"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No adapted biometrics for diverse populations",
                    metadata={
                        "framework": "Alternative Credentials",
                        "requirement": "Biometric Authentication",
                        "recommendation": "Implement biometric solutions adapted for diverse population characteristics"
                    }
                )
            )
        
        # Verificar suporte para autenticação via mobile money
        mobile_money_auth = auth_config.get("mobile_money_integration", False)
        if not mobile_money_auth:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["alt-auth-mobile"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.LOW,
                    message="No mobile money authentication integration",
                    metadata={
                        "framework": "Alternative Credentials",
                        "requirement": "Mobile Authentication",
                        "recommendation": "Integrate with local mobile money platforms for authentication"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Alternative authentication methods are properly implemented",
                    metadata={
                        "framework": "Alternative Credentials",
                        "requirements": "Mobile, Biometric, Offline"
                    }
                )
            )
        
        return results

class InclusiveDesignRule(AlternativeCredentialsRule):
    """Validação de design inclusivo para mercados emergentes"""
    
    def __init__(self):
        super().__init__(
            "inclusive_design",
            "Inclusive Authentication Design",
            "Verifies that authentication is designed inclusively for all users",
            ValidationSeverity.MEDIUM
        )
    
    def get_requirements(self) -> List[str]:
        return ["inclusive-literacy", "inclusive-accessibility", "inclusive-cultural"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de design inclusivo
        inclusive_config = context.config.get("inclusive_design", {})
        
        # Verificar suporte para baixo letramento
        low_literacy = inclusive_config.get("low_literacy_support", False)
        if not low_literacy:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["inclusive-literacy"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No support for users with low literacy",
                    metadata={
                        "framework": "Inclusive Design",
                        "requirement": "Literacy Support",
                        "recommendation": "Implement voice instructions and icon-based interfaces"
                    }
                )
            )
        
        # Verificar suporte para acessibilidade
        accessibility = inclusive_config.get("accessibility_support", False)
        if not accessibility:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["inclusive-accessibility"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No accessibility features implemented",
                    metadata={
                        "framework": "Inclusive Design",
                        "requirement": "Accessibility",
                        "recommendation": "Implement WCAG 2.1 accessibility features"
                    }
                )
            )
        
        # Verificar sensibilidade cultural
        cultural_sensitivity = inclusive_config.get("cultural_sensitivity", False)
        if not cultural_sensitivity:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["inclusive-cultural"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.LOW,
                    message="No cultural sensitivity in design",
                    metadata={
                        "framework": "Inclusive Design",
                        "requirement": "Cultural Sensitivity",
                        "recommendation": "Implement culturally appropriate design elements and terminology"
                    }
                )
            )
        
        # Verificar suporte a múltiplos idiomas (incluindo línguas locais)
        local_languages = inclusive_config.get("local_language_support", False)
        if not local_languages:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["inclusive-cultural"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No support for local languages",
                    metadata={
                        "framework": "Inclusive Design",
                        "requirement": "Cultural Sensitivity",
                        "recommendation": "Implement support for local Angolan languages"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Inclusive design principles are properly implemented",
                    metadata={
                        "framework": "Inclusive Design",
                        "requirements": "Literacy, Accessibility, Cultural"
                    }
                )
            )
        
        return results

# Lista de todas as regras implementadas
def get_africa_rules() -> List[ValidationRule]:
    """Retorna todas as regras de validação para África (Angola)"""
    return [
        # Lei de Proteção de Dados de Angola
        AngolaConsentRule(),
        AngolaDataTransferRule(),
        AngolaDataSubjectRightsRule(),
        
        # Regras para mercados emergentes
        AlternativeAuthenticationRule(),
        InclusiveDesignRule()
    ]
