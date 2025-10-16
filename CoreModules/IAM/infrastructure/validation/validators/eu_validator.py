"""
INNOVABIZ - Validador de Conformidade para União Europeia
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de validador de conformidade para a União Europeia,
           incluindo GDPR, EU AI Act, eIDAS 2.0, NIS2 e DORA.
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

class GDPRBaseRule(ValidationRule):
    """Regra base para validação GDPR"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.EU, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.GDPR]

class EUIDARule(ValidationRule):
    """Regra base para validação eIDAS 2.0"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.EU, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("eidas_2_0")]

class EUAIActRule(ValidationRule):
    """Regra base para validação EU AI Act"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.EU, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("eu_ai_act_2025")]

class NIS2Rule(ValidationRule):
    """Regra base para validação NIS2"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.EU, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("nis2")]

class DORARule(ValidationRule):
    """Regra base para validação DORA"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.EU, Region.GLOBAL]
    
    def get_applicable_industries(self) -> List[Industry]:
        return [Industry.FINANCIAL, Industry.BANKING, Industry.INSURANCE]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("dora")]

# Implementação de regras GDPR específicas
class GDPRConsentRule(GDPRBaseRule):
    """Validação de consentimento GDPR"""
    
    def __init__(self):
        super().__init__(
            "gdpr_consent",
            "GDPR Consent Requirements",
            "Verifies that authentication processes obtain explicit consent for personal data processing",
            ValidationSeverity.CRITICAL
        )
    
    def get_requirements(self) -> List[str]:
        return ["gdpr-art6-1a", "gdpr-art7"]
    
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
                    requirement_ids=["gdpr-art7"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.CRITICAL,
                    message="Authentication does not require explicit consent for data processing",
                    metadata={
                        "framework": "GDPR",
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
                    requirement_ids=["gdpr-art7"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No mechanism to withdraw consent",
                    metadata={
                        "framework": "GDPR",
                        "article": "Article 7(3)",
                        "recommendation": "Implement consent withdrawal mechanism"
                    }
                )
            )
        
        # Verificar se há informações claras sobre o processamento de dados
        clear_info = consent_config.get("clear_information", False)
        if not clear_info:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["gdpr-art13", "gdpr-art14"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No clear information about data processing provided",
                    metadata={
                        "framework": "GDPR",
                        "article": "Articles 13 and 14",
                        "recommendation": "Add clear information about data processing purposes"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if not results:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["gdpr-art6-1a", "gdpr-art7"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Authentication properly implements GDPR consent requirements",
                    metadata={
                        "framework": "GDPR",
                        "article": "Articles 6(1)(a) and 7"
                    }
                )
            )
        
        return results

class GDPRDataMinimizationRule(GDPRBaseRule):
    """Validação de minimização de dados GDPR"""
    
    def __init__(self):
        super().__init__(
            "gdpr_data_minimization",
            "GDPR Data Minimization",
            "Verifies that authentication processes collect only necessary data",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["gdpr-art5-1c"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de coleta de dados
        auth_data = context.config.get("authentication", {}).get("collected_data", [])
        
        # Lista de campos considerados excessivos para autenticação básica
        excessive_fields = [
            "address", "age", "gender", "nationality", "income", 
            "social_media", "browsing_history", "location_history"
        ]
        
        # Verificar campos excessivos
        excessive_found = [field for field in auth_data if field in excessive_fields]
        
        if excessive_found:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["gdpr-art5-1c"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message=f"Authentication collects excessive data: {', '.join(excessive_found)}",
                    metadata={
                        "framework": "GDPR",
                        "article": "Article 5(1)(c)",
                        "excessive_fields": excessive_found,
                        "recommendation": "Remove excessive data collection fields"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["gdpr-art5-1c"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Authentication properly implements data minimization",
                    metadata={
                        "framework": "GDPR",
                        "article": "Article 5(1)(c)"
                    }
                )
            )
        
        return results

class GDPRRightsImplementationRule(GDPRBaseRule):
    """Validação de implementação dos direitos do titular GDPR"""
    
    def __init__(self):
        super().__init__(
            "gdpr_rights_implementation",
            "GDPR Rights Implementation",
            "Verifies that IAM implements data subject rights",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["gdpr-art15", "gdpr-art16", "gdpr-art17", "gdpr-art20"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de direitos do titular
        rights_config = context.config.get("data_subject_rights", {})
        
        # Lista de direitos essenciais
        essential_rights = {
            "access": "gdpr-art15",
            "rectification": "gdpr-art16",
            "erasure": "gdpr-art17",
            "portability": "gdpr-art20"
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
                            "framework": "GDPR",
                            "article": f"Article {requirement.split('-')[-1]}",
                            "recommendation": f"Implement mechanism for data subject's right to {right}"
                        }
                    )
                )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=list(essential_rights.values()),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="All essential data subject rights are implemented",
                    metadata={
                        "framework": "GDPR",
                        "articles": "Articles 15, 16, 17, 20"
                    }
                )
            )
        
        return results

# Implementação de regras eIDAS 2.0 específicas
class EIDASAssuranceLevelRule(EUIDARule):
    """Validação de nível de garantia eIDAS 2.0"""
    
    def __init__(self):
        super().__init__(
            "eidas_assurance_level",
            "eIDAS 2.0 Assurance Level",
            "Verifies that authentication meets eIDAS 2.0 assurance level requirements",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["eidas-art8"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de nível de garantia
        auth_config = context.config.get("authentication", {})
        assurance_level = auth_config.get("assurance_level", "")
        
        # Níveis de garantia válidos
        valid_levels = ["low", "substantial", "high"]
        
        if assurance_level not in valid_levels:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["eidas-art8"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message=f"Invalid or missing eIDAS assurance level: {assurance_level}",
                    metadata={
                        "framework": "eIDAS 2.0",
                        "article": "Article 8",
                        "recommendation": "Implement a valid eIDAS assurance level (low, substantial, high)"
                    }
                )
            )
        else:
            # Para nível substantial e high, verificar requisitos adicionais
            if assurance_level in ["substantial", "high"]:
                factors = auth_config.get("authentication_factors", [])
                
                # Verificar MFA para nível substantial e high
                if len(factors) < 2:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["eidas-art8"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message=f"Insufficient factors for eIDAS level '{assurance_level}'",
                            metadata={
                                "framework": "eIDAS 2.0",
                                "article": "Article 8",
                                "current_factors": factors,
                                "recommendation": "Implement multi-factor authentication with at least 2 factors"
                            }
                        )
                    )
                
                # Para nível high, verificar requisitos adicionais
                if assurance_level == "high":
                    has_hardware = any(f.get("type") == "hardware" for f in factors)
                    if not has_hardware:
                        results.append(
                            ComplianceValidationResult(
                                rule_id=self.id,
                                requirement_ids=["eidas-art8"],
                                status=ValidationStatus.FAIL,
                                severity=ValidationSeverity.HIGH,
                                message="No hardware factor for eIDAS level 'high'",
                                metadata={
                                    "framework": "eIDAS 2.0",
                                    "article": "Article 8",
                                    "recommendation": "Implement hardware-based authentication factor"
                                }
                            )
                        )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["eidas-art8"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message=f"Authentication meets eIDAS 2.0 {assurance_level} assurance level",
                    metadata={
                        "framework": "eIDAS 2.0",
                        "article": "Article 8",
                        "assurance_level": assurance_level
                    }
                )
            )
        
        return results

# Implementação de regras EU AI Act específicas
class AIActRiskCategoryRule(EUAIActRule):
    """Validação de categoria de risco do EU AI Act"""
    
    def __init__(self):
        super().__init__(
            "ai_act_risk_category",
            "EU AI Act Risk Categorization",
            "Verifies that AI systems for authentication are properly categorized",
            ValidationSeverity.CRITICAL
        )
    
    def get_requirements(self) -> List[str]:
        return ["ai_act-art6", "ai_act-art7", "ai_act-art8", "ai_act-art9"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de IA
        ai_config = context.config.get("ai_systems", {})
        
        # Verificar sistemas de IA utilizados na autenticação
        auth_ai_systems = ai_config.get("authentication", [])
        
        if not auth_ai_systems:
            # Se não há sistemas de IA, esta regra não se aplica
            return [
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="No AI systems used in authentication, EU AI Act not applicable",
                    metadata={
                        "framework": "EU AI Act 2025"
                    }
                )
            ]
        
        # Verificar categorização de risco
        for system in auth_ai_systems:
            system_name = system.get("name", "Unknown AI System")
            risk_category = system.get("risk_category", "")
            
            # Categorias válidas
            valid_categories = ["unacceptable", "high_risk", "limited_risk", "minimal_risk"]
            
            if risk_category not in valid_categories:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        requirement_ids=["ai_act-art6"],
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.CRITICAL,
                        message=f"AI system '{system_name}' has invalid or missing risk category",
                        metadata={
                            "framework": "EU AI Act 2025",
                            "article": "Article 6",
                            "system": system_name,
                            "recommendation": "Properly categorize AI system according to EU AI Act risk levels"
                        }
                    )
                )
                continue
            
            # Para sistemas de alto risco, verificar requisitos adicionais
            if risk_category == "high_risk":
                # Verificar avaliação de risco
                risk_assessment = system.get("risk_assessment", False)
                if not risk_assessment:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["ai_act-art9"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message=f"No risk assessment for high-risk AI system '{system_name}'",
                            metadata={
                                "framework": "EU AI Act 2025",
                                "article": "Article 9",
                                "system": system_name,
                                "recommendation": "Conduct and document risk assessment for high-risk AI system"
                            }
                        )
                    )
                
                # Verificar supervisão humana
                human_oversight = system.get("human_oversight", False)
                if not human_oversight:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["ai_act-art14"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message=f"No human oversight for high-risk AI system '{system_name}'",
                            metadata={
                                "framework": "EU AI Act 2025",
                                "article": "Article 14",
                                "system": system_name,
                                "recommendation": "Implement human oversight mechanisms for high-risk AI system"
                            }
                        )
                    )
            
            # Para sistemas de risco inaceitável
            if risk_category == "unacceptable":
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        requirement_ids=["ai_act-art5"],
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.CRITICAL,
                        message=f"AI system '{system_name}' categorized as unacceptable risk",
                        metadata={
                            "framework": "EU AI Act 2025",
                            "article": "Article 5",
                            "system": system_name,
                            "recommendation": "Remove or replace AI system with unacceptable risk"
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
                    message="All AI systems properly categorized according to EU AI Act",
                    metadata={
                        "framework": "EU AI Act 2025"
                    }
                )
            )
        
        return results

class NIS2IncidentReportingRule(NIS2Rule):
    """Validação de relatórios de incidentes NIS2"""
    
    def __init__(self):
        super().__init__(
            "nis2_incident_reporting",
            "NIS2 Incident Reporting",
            "Verifies that IAM has mechanisms for incident reporting according to NIS2",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["nis2-art20"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de relatórios de incidentes
        incident_config = context.config.get("incident_management", {})
        
        # Verificar se existe processo de relatório de incidentes
        reporting_process = incident_config.get("reporting_process", False)
        if not reporting_process:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nis2-art20"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No incident reporting process defined",
                    metadata={
                        "framework": "NIS2",
                        "article": "Article 20",
                        "recommendation": "Implement formal incident reporting process according to NIS2"
                    }
                )
            )
        
        # Verificar tempo máximo para relatório de incidentes
        reporting_time = incident_config.get("max_reporting_time_hours", 72)
        if reporting_time > 24:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nis2-art20"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message=f"Incident reporting time ({reporting_time} hours) exceeds NIS2 recommendation",
                    metadata={
                        "framework": "NIS2",
                        "article": "Article 20",
                        "current_time": reporting_time,
                        "recommendation": "Reduce incident reporting time to 24 hours or less"
                    }
                )
            )
        
        # Verificar classificação de severidade de incidentes
        has_severity_classification = incident_config.get("severity_classification", False)
        if not has_severity_classification:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nis2-art20"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No incident severity classification system",
                    metadata={
                        "framework": "NIS2",
                        "article": "Article 20",
                        "recommendation": "Implement incident severity classification according to NIS2"
                    }
                )
            )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nis2-art20"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Incident reporting process complies with NIS2 requirements",
                    metadata={
                        "framework": "NIS2",
                        "article": "Article 20"
                    }
                )
            )
        
        return results

class DORAOperationalResilienceRule(DORARule):
    """Validação de resiliência operacional DORA"""
    
    def __init__(self):
        super().__init__(
            "dora_operational_resilience",
            "DORA Operational Resilience",
            "Verifies that IAM implements operational resilience according to DORA",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["dora-art11", "dora-art12", "dora-art13"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de resiliência operacional
        resilience_config = context.config.get("operational_resilience", {})
        
        # Verificar plano de continuidade de negócios
        has_bcp = resilience_config.get("business_continuity_plan", False)
        if not has_bcp:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["dora-art11"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No business continuity plan for IAM",
                    metadata={
                        "framework": "DORA",
                        "article": "Article 11",
                        "recommendation": "Implement business continuity plan for IAM"
                    }
                )
            )
        
        # Verificar testes de resiliência
        testing_frequency = resilience_config.get("resilience_testing_frequency_months", 0)
        if testing_frequency == 0 or testing_frequency > 12:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["dora-art12"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message=f"Insufficient resilience testing frequency: {testing_frequency} months",
                    metadata={
                        "framework": "DORA",
                        "article": "Article 12",
                        "current_frequency": testing_frequency,
                        "recommendation": "Conduct resilience testing at least annually"
                    }
                )
            )
        
        # Verificar plano de recuperação de desastres
        has_drp = resilience_config.get("disaster_recovery_plan", False)
        if not has_drp:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["dora-art13"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No disaster recovery plan for IAM",
                    metadata={
                        "framework": "DORA",
                        "article": "Article 13",
                        "recommendation": "Implement disaster recovery plan for IAM"
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
                    message="Operational resilience complies with DORA requirements",
                    metadata={
                        "framework": "DORA",
                        "articles": "Articles 11, 12, 13"
                    }
                )
            )
        
        return results

# Lista de todas as regras implementadas
def get_eu_rules() -> List[ValidationRule]:
    """Retorna todas as regras de validação para a UE"""
    return [
        # GDPR
        GDPRConsentRule(),
        GDPRDataMinimizationRule(),
        GDPRRightsImplementationRule(),
        
        # eIDAS 2.0
        EIDASAssuranceLevelRule(),
        
        # EU AI Act
        AIActRiskCategoryRule(),
        
        # NIS2
        NIS2IncidentReportingRule(),
        
        # DORA
        DORAOperationalResilienceRule()
    ]
