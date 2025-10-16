"""
INNOVABIZ - Validador de Conformidade para Estados Unidos
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de validador de conformidade para os EUA,
           incluindo NIST 800-63-4, CMMC 2.0 e HIPAA.
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

class NISTBaseRule(ValidationRule):
    """Regra base para validação NIST 800-63"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.USA, Region.GLOBAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("nist_800_63_4")]

class CMMCBaseRule(ValidationRule):
    """Regra base para validação CMMC 2.0"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.USA, Region.GLOBAL]
    
    def get_applicable_industries(self) -> List[Industry]:
        return [Industry.DEFENSE, Industry.GOVERNMENT]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("cmmc_2_0")]

class HIPAABaseRule(ValidationRule):
    """Regra base para validação HIPAA"""
    
    def __init__(
        self, 
        rule_id: str, 
        name: str, 
        description: str, 
        severity: ValidationSeverity = ValidationSeverity.HIGH
    ):
        super().__init__(rule_id, name, description, severity)
    
    def get_applicable_regions(self) -> List[Region]:
        return [Region.USA, Region.GLOBAL]
    
    def get_applicable_industries(self) -> List[Industry]:
        return [Industry.HEALTHCARE, Industry.INSURANCE]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        return [ComplianceFramework.from_string("hipaa")]

# Implementação de regras NIST 800-63-4 específicas
class NISTIALRule(NISTBaseRule):
    """Validação de níveis de garantia de identidade NIST"""
    
    def __init__(self):
        super().__init__(
            "nist_ial",
            "NIST Identity Assurance Level",
            "Verifies compliance with NIST 800-63-4 Identity Assurance Levels",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["nist-800-63a"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de IAL
        auth_config = context.config.get("authentication", {})
        identity_level = auth_config.get("identity_assurance_level", "")
        
        # Níveis válidos
        valid_levels = ["ial1", "ial2", "ial3"]
        
        if identity_level not in valid_levels:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63a"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message=f"Invalid or missing NIST Identity Assurance Level: {identity_level}",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63A",
                        "recommendation": "Implement a valid NIST IAL (ial1, ial2, ial3)"
                    }
                )
            )
        else:
            # Verificar requisitos específicos para cada nível
            if identity_level in ["ial2", "ial3"]:
                id_proofing = auth_config.get("identity_proofing", False)
                if not id_proofing:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["nist-800-63a"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message=f"No identity proofing process for {identity_level}",
                            metadata={
                                "framework": "NIST 800-63-4",
                                "section": "800-63A",
                                "recommendation": "Implement identity proofing process"
                            }
                        )
                    )
            
            # Requisitos específicos para IAL3
            if identity_level == "ial3":
                in_person = auth_config.get("in_person_proofing", False)
                if not in_person:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["nist-800-63a"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message="No in-person proofing for IAL3",
                            metadata={
                                "framework": "NIST 800-63-4",
                                "section": "800-63A",
                                "recommendation": "Implement in-person identity proofing"
                            }
                        )
                    )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63a"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message=f"Authentication meets NIST 800-63-4 {identity_level} requirements",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63A",
                        "identity_level": identity_level
                    }
                )
            )
        
        return results

class NISTAALRule(NISTBaseRule):
    """Validação de níveis de garantia de autenticação NIST"""
    
    def __init__(self):
        super().__init__(
            "nist_aal",
            "NIST Authentication Assurance Level",
            "Verifies compliance with NIST 800-63-4 Authentication Assurance Levels",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["nist-800-63b"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de AAL
        auth_config = context.config.get("authentication", {})
        auth_level = auth_config.get("authentication_assurance_level", "")
        
        # Níveis válidos
        valid_levels = ["aal1", "aal2", "aal3"]
        
        if auth_level not in valid_levels:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63b"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message=f"Invalid or missing NIST Authentication Assurance Level: {auth_level}",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63B",
                        "recommendation": "Implement a valid NIST AAL (aal1, aal2, aal3)"
                    }
                )
            )
        else:
            # Verificar requisitos específicos para cada nível
            if auth_level in ["aal2", "aal3"]:
                factors = auth_config.get("authentication_factors", [])
                
                # Verificar MFA para AAL2 e AAL3
                if len(factors) < 2:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["nist-800-63b"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.HIGH,
                            message=f"Insufficient factors for {auth_level}",
                            metadata={
                                "framework": "NIST 800-63-4",
                                "section": "800-63B",
                                "current_factors": factors,
                                "recommendation": "Implement multi-factor authentication with at least 2 factors"
                            }
                        )
                    )
                
                # Requisitos específicos para AAL3
                if auth_level == "aal3":
                    has_hardware = any(f.get("type") == "hardware" for f in factors)
                    if not has_hardware:
                        results.append(
                            ComplianceValidationResult(
                                rule_id=self.id,
                                requirement_ids=["nist-800-63b"],
                                status=ValidationStatus.FAIL,
                                severity=ValidationSeverity.HIGH,
                                message="No hardware factor for AAL3",
                                metadata={
                                    "framework": "NIST 800-63-4",
                                    "section": "800-63B",
                                    "recommendation": "Implement hardware-based authentication factor"
                                }
                            )
                        )
        
        # Verificar políticas de senha
        if auth_level in valid_levels:
            password_policy = auth_config.get("password_policy", {})
            min_length = password_policy.get("min_length", 0)
            
            if min_length < 8:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        requirement_ids=["nist-800-63b"],
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.MEDIUM,
                        message=f"Password minimum length ({min_length}) below NIST recommendation",
                        metadata={
                            "framework": "NIST 800-63-4",
                            "section": "800-63B",
                            "current_length": min_length,
                            "recommendation": "Set password minimum length to at least 8 characters"
                        }
                    )
                )
            
            # Verificar verificação de senhas comprometidas
            check_breached = password_policy.get("check_breached_passwords", False)
            if not check_breached:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        requirement_ids=["nist-800-63b"],
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.MEDIUM,
                        message="No check for breached/compromised passwords",
                        metadata={
                            "framework": "NIST 800-63-4",
                            "section": "800-63B",
                            "recommendation": "Implement check for compromised passwords"
                        }
                    )
                )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63b"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message=f"Authentication meets NIST 800-63-4 {auth_level} requirements",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63B",
                        "authentication_level": auth_level
                    }
                )
            )
        
        return results

class NISTFALRule(NISTBaseRule):
    """Validação de níveis de garantia federada NIST"""
    
    def __init__(self):
        super().__init__(
            "nist_fal",
            "NIST Federation Assurance Level",
            "Verifies compliance with NIST 800-63-4 Federation Assurance Levels",
            ValidationSeverity.MEDIUM
        )
    
    def get_requirements(self) -> List[str]:
        return ["nist-800-63c"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de FAL
        federation_config = context.config.get("federation", {})
        
        # Verificar se federação está em uso
        federation_enabled = federation_config.get("enabled", False)
        
        if not federation_enabled:
            # Se federação não está em uso, esta regra não se aplica
            return [
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=self.get_requirements(),
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message="Federation not in use, NIST FAL not applicable",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63C"
                    }
                )
            ]
        
        # Verificar configuração de FAL
        fed_level = federation_config.get("federation_assurance_level", "")
        
        # Níveis válidos
        valid_levels = ["fal1", "fal2", "fal3"]
        
        if fed_level not in valid_levels:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63c"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message=f"Invalid or missing NIST Federation Assurance Level: {fed_level}",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63C",
                        "recommendation": "Implement a valid NIST FAL (fal1, fal2, fal3)"
                    }
                )
            )
        else:
            # Verificar requisitos específicos para cada nível
            if fed_level in ["fal2", "fal3"]:
                signed_assertions = federation_config.get("signed_assertions", False)
                if not signed_assertions:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["nist-800-63c"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.MEDIUM,
                            message=f"No signed assertions for {fed_level}",
                            metadata={
                                "framework": "NIST 800-63-4",
                                "section": "800-63C",
                                "recommendation": "Implement signed assertions for federation"
                            }
                        )
                    )
            
            # Requisitos específicos para FAL3
            if fed_level == "fal3":
                encrypted_assertions = federation_config.get("encrypted_assertions", False)
                if not encrypted_assertions:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            requirement_ids=["nist-800-63c"],
                            status=ValidationStatus.FAIL,
                            severity=ValidationSeverity.MEDIUM,
                            message="No encrypted assertions for FAL3",
                            metadata={
                                "framework": "NIST 800-63-4",
                                "section": "800-63C",
                                "recommendation": "Implement encrypted assertions for federation"
                            }
                        )
                    )
        
        # Se não houver falhas, adicionar resultado de aprovação
        if len(results) == 0:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["nist-800-63c"],
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.NONE,
                    message=f"Federation meets NIST 800-63-4 {fed_level} requirements",
                    metadata={
                        "framework": "NIST 800-63-4",
                        "section": "800-63C",
                        "federation_level": fed_level
                    }
                )
            )
        
        return results

class CMMCAccessControlRule(CMMCBaseRule):
    """Validação de controle de acesso CMMC 2.0"""
    
    def __init__(self):
        super().__init__(
            "cmmc_access_control",
            "CMMC 2.0 Access Control",
            "Verifies compliance with CMMC 2.0 Access Control practices",
            ValidationSeverity.HIGH
        )
    
    def get_requirements(self) -> List[str]:
        return ["cmmc-ac.1.001", "cmmc-ac.1.002", "cmmc-ac.2.007", "cmmc-ac.2.008"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de controle de acesso
        access_config = context.config.get("access_control", {})
        
        # Verificar controle de acesso baseado em papéis
        rbac_implemented = access_config.get("rbac_implemented", False)
        if not rbac_implemented:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["cmmc-ac.1.001"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No role-based access control implemented",
                    metadata={
                        "framework": "CMMC 2.0",
                        "practice": "AC.1.001",
                        "recommendation": "Implement role-based access control"
                    }
                )
            )
        
        # Verificar segregação de funções
        segregation = access_config.get("segregation_of_duties", False)
        if not segregation:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["cmmc-ac.2.007"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No segregation of duties implemented",
                    metadata={
                        "framework": "CMMC 2.0",
                        "practice": "AC.2.007",
                        "recommendation": "Implement segregation of duties"
                    }
                )
            )
        
        # Verificar controle de acesso privilegiado
        privileged_access = access_config.get("privileged_access_management", False)
        if not privileged_access:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["cmmc-ac.2.008"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No privileged access management",
                    metadata={
                        "framework": "CMMC 2.0",
                        "practice": "AC.2.008",
                        "recommendation": "Implement privileged access management"
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
                    message="Access control complies with CMMC 2.0 requirements",
                    metadata={
                        "framework": "CMMC 2.0",
                        "practices": "AC.1.001, AC.1.002, AC.2.007, AC.2.008"
                    }
                )
            )
        
        return results

class HIPAAAccessControlRule(HIPAABaseRule):
    """Validação de controle de acesso HIPAA"""
    
    def __init__(self):
        super().__init__(
            "hipaa_access_control",
            "HIPAA Access Control",
            "Verifies compliance with HIPAA access control requirements",
            ValidationSeverity.CRITICAL
        )
    
    def get_requirements(self) -> List[str]:
        return ["hipaa-164.312(a)(1)", "hipaa-164.312(a)(2)(i)", "hipaa-164.312(a)(2)(iv)"]
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        results = []
        
        # Verificar configuração de controle de acesso
        access_config = context.config.get("access_control", {})
        
        # Verificar controle de acesso único
        unique_user_ids = access_config.get("unique_user_ids", False)
        if not unique_user_ids:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["hipaa-164.312(a)(2)(i)"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.CRITICAL,
                    message="No unique user identification implemented",
                    metadata={
                        "framework": "HIPAA",
                        "section": "164.312(a)(2)(i)",
                        "recommendation": "Implement unique user identification"
                    }
                )
            )
        
        # Verificar procedimentos de emergência
        emergency_access = access_config.get("emergency_access_procedure", False)
        if not emergency_access:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["hipaa-164.312(a)(2)(ii)"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="No emergency access procedure",
                    metadata={
                        "framework": "HIPAA",
                        "section": "164.312(a)(2)(ii)",
                        "recommendation": "Implement emergency access procedure"
                    }
                )
            )
        
        # Verificar terminação de sessão
        auto_logout = access_config.get("auto_logout", False)
        if not auto_logout:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["hipaa-164.312(a)(2)(iii)"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="No automatic session termination",
                    metadata={
                        "framework": "HIPAA",
                        "section": "164.312(a)(2)(iii)",
                        "recommendation": "Implement automatic session termination"
                    }
                )
            )
        
        # Verificar criptografia
        encryption = access_config.get("encryption", False)
        if not encryption:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    requirement_ids=["hipaa-164.312(a)(2)(iv)"],
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.CRITICAL,
                    message="No encryption mechanism for PHI",
                    metadata={
                        "framework": "HIPAA",
                        "section": "164.312(a)(2)(iv)",
                        "recommendation": "Implement encryption for protected health information"
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
                    message="Access control complies with HIPAA requirements",
                    metadata={
                        "framework": "HIPAA",
                        "section": "164.312(a)"
                    }
                )
            )
        
        return results

# Lista de todas as regras implementadas
def get_us_rules() -> List[ValidationRule]:
    """Retorna todas as regras de validação para os EUA"""
    return [
        # NIST 800-63-4
        NISTIALRule(),
        NISTAALRule(),
        NISTFALRule(),
        
        # CMMC 2.0
        CMMCAccessControlRule(),
        
        # HIPAA
        HIPAAAccessControlRule()
    ]
