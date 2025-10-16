"""
INNOVABIZ - Regras de Validação de Compliance GDPR para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Regras de validação específicas para GDPR (General Data
           Protection Regulation) da União Europeia.
==================================================================
"""

from enum import Enum
from typing import Dict, List, Any, Optional
from dataclasses import dataclass

from ..compliance_engine import ValidationRule, ValidationContext
from ..compliance_metadata import Region, Industry, ComplianceFramework
from ..models import ComplianceValidationResult, ValidationSeverity, ValidationStatus


class GDPRAuthenticationRule(ValidationRule):
    """Regra para validar autenticação conforme GDPR."""
    
    def __init__(self):
        """Inicializa a regra de autenticação GDPR."""
        super().__init__(
            rule_id="gdpr_authentication",
            name="GDPR Authentication",
            description="Verifica se os controles de autenticação estão em conformidade com os requisitos do GDPR para proteção de dados pessoais",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com GDPR para autenticação.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar se há mecanismos de autenticação forte
        auth_config = config.get("authentication", {})
        mfa_enabled = auth_config.get("mfa_enabled", False)
        
        if not mfa_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="MFA não está habilitado, o que é recomendado pelo GDPR para proteção de dados pessoais",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"mfa_enabled": mfa_enabled}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "authentication"
                    }
                )
            )
        else:
            # Verificar se os métodos MFA são adequados
            mfa_methods = auth_config.get("mfa_methods", [])
            strong_mfa = any(m in ["totp", "fido2", "smart_card", "biometric"] for m in mfa_methods)
            
            if not strong_mfa:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="MFA está habilitado, mas os métodos podem não ser suficientemente fortes para dados sensíveis sob GDPR",
                        details={
                            "requirement": "Artigo 32 - Segurança do processamento",
                            "current_config": {"mfa_methods": mfa_methods},
                            "recommendation": "Implementar métodos MFA mais fortes como TOTP, FIDO2, ou biometria"
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "32",
                            "control_type": "authentication"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=self.severity,
                        message="Configuração MFA é adequada para conformidade com GDPR",
                        details={
                            "requirement": "Artigo 32 - Segurança do processamento",
                            "current_config": {"mfa_methods": mfa_methods}
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "32",
                            "control_type": "authentication"
                        }
                    )
                )
        
        # Verificar políticas de senha
        password_policy = auth_config.get("password_policy", {})
        min_length = password_policy.get("min_length", 0)
        require_complexity = password_policy.get("require_complexity", False)
        
        if min_length < 12 or not require_complexity:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Política de senha pode não ser adequada para proteção de dados sob GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity
                        },
                        "recommendation": "Implementar senhas com pelo menos 12 caracteres e requisitos de complexidade"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "authentication"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=self.severity,
                    message="Política de senha é adequada para conformidade com GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity
                        }
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "authentication"
                    }
                )
            )
        
        # Verificar gestão de sessão
        session_config = auth_config.get("session", {})
        session_timeout = session_config.get("timeout_minutes", 0)
        
        if session_timeout <= 0 or session_timeout > 60:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Timeout de sessão inadequado para sistemas que processam dados pessoais sob GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"session_timeout": session_timeout},
                        "recommendation": "Implementar timeout de sessão de no máximo 60 minutos"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "session_management"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=self.severity,
                    message="Timeout de sessão adequado para conformidade com GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"session_timeout": session_timeout}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "session_management"
                    }
                )
            )
        
        return results
    
    def get_requirements(self) -> List[str]:
        """
        Obtém os IDs dos requisitos que esta regra valida.
        
        Returns:
            Lista de IDs de requisitos
        """
        return ["gdpr_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [
            Region.EU_CENTRAL, 
            Region.EU_NORTH, 
            Region.EU_SOUTH, 
            Region.EU_WEST
        ]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.GDPR]


class GDPRDataProtectionRule(ValidationRule):
    """Regra para validar proteção de dados conforme GDPR."""
    
    def __init__(self):
        """Inicializa a regra de proteção de dados GDPR."""
        super().__init__(
            rule_id="gdpr_data_protection",
            name="GDPR Data Protection",
            description="Verifica se os controles de proteção de dados estão em conformidade com os requisitos do GDPR",
            severity=ValidationSeverity.CRITICAL
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com GDPR para proteção de dados.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar criptografia de dados em repouso
        encryption_config = config.get("data_protection", {}).get("encryption", {})
        at_rest_encryption = encryption_config.get("at_rest", False)
        
        if not at_rest_encryption:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Criptografia de dados em repouso não está habilitada, o que é necessário pelo GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"at_rest_encryption": at_rest_encryption},
                        "recommendation": "Implementar criptografia de dados em repouso"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "encryption"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=self.severity,
                    message="Criptografia de dados em repouso está habilitada conforme requisitos do GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"at_rest_encryption": at_rest_encryption}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "encryption"
                    }
                )
            )
        
        # Verificar criptografia de dados em trânsito
        in_transit_encryption = encryption_config.get("in_transit", False)
        
        if not in_transit_encryption:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Criptografia de dados em trânsito não está habilitada, o que é necessário pelo GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"in_transit_encryption": in_transit_encryption},
                        "recommendation": "Implementar criptografia de dados em trânsito (TLS)"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "encryption"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=self.severity,
                    message="Criptografia de dados em trânsito está habilitada conforme requisitos do GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"in_transit_encryption": in_transit_encryption}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "encryption"
                    }
                )
            )
        
        # Verificar controle de acesso a dados pessoais
        access_control = config.get("access_control", {})
        rbac_enabled = access_control.get("rbac_enabled", False)
        
        if not rbac_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Controle de acesso baseado em papéis (RBAC) não está habilitado, o que é recomendado pelo GDPR",
                    details={
                        "requirement": "Artigo 25 - Proteção de dados por design",
                        "current_config": {"rbac_enabled": rbac_enabled},
                        "recommendation": "Implementar RBAC para controlar acesso a dados pessoais"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "25",
                        "control_type": "access_control"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.HIGH,
                    message="Controle de acesso baseado em papéis (RBAC) está habilitado conforme recomendado pelo GDPR",
                    details={
                        "requirement": "Artigo 25 - Proteção de dados por design",
                        "current_config": {"rbac_enabled": rbac_enabled}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "25",
                        "control_type": "access_control"
                    }
                )
            )
        
        # Verificar anonimização e pseudonimização
        data_processing = config.get("data_processing", {})
        pseudonymization = data_processing.get("pseudonymization", False)
        
        if not pseudonymization:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Pseudonimização de dados não está configurada, o que é explicitamente mencionado pelo GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"pseudonymization": pseudonymization},
                        "recommendation": "Implementar técnicas de pseudonimização para dados pessoais"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "pseudonymization"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Pseudonimização de dados está configurada conforme recomendado pelo GDPR",
                    details={
                        "requirement": "Artigo 32 - Segurança do processamento",
                        "current_config": {"pseudonymization": pseudonymization}
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "32",
                        "control_type": "pseudonymization"
                    }
                )
            )
        
        return results
    
    def get_requirements(self) -> List[str]:
        """
        Obtém os IDs dos requisitos que esta regra valida.
        
        Returns:
            Lista de IDs de requisitos
        """
        return ["gdpr_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [
            Region.EU_CENTRAL, 
            Region.EU_NORTH, 
            Region.EU_SOUTH, 
            Region.EU_WEST
        ]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.GDPR]


class GDPRConsentManagementRule(ValidationRule):
    """Regra para validar gestão de consentimento conforme GDPR."""
    
    def __init__(self):
        """Inicializa a regra de gestão de consentimento GDPR."""
        super().__init__(
            rule_id="gdpr_consent_management",
            name="GDPR Consent Management",
            description="Verifica se os controles de gestão de consentimento estão em conformidade com os requisitos do GDPR",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com GDPR para gestão de consentimento.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar se há funcionalidade de gestão de consentimento
        consent_mgmt = config.get("consent_management", {})
        consent_enabled = consent_mgmt.get("enabled", False)
        
        if not consent_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Gestão de consentimento não está habilitada, o que é necessário pelo GDPR",
                    details={
                        "requirement": "Artigo 7 - Condições para consentimento",
                        "current_config": {"consent_enabled": consent_enabled},
                        "recommendation": "Implementar gestão de consentimento com capacidade de revogação"
                    },
                    metadata={
                        "framework": "gdpr",
                        "article": "7",
                        "control_type": "consent"
                    }
                )
            )
        else:
            # Verificar recursos específicos de gestão de consentimento
            revocation = consent_mgmt.get("revocation_supported", False)
            granular_options = consent_mgmt.get("granular_options", False)
            audit_trail = consent_mgmt.get("audit_trail", False)
            
            if not revocation:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.FAIL,
                        severity=self.severity,
                        message="Revogação de consentimento não é suportada, o que é explicitamente exigido pelo GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"revocation_supported": revocation},
                            "recommendation": "Implementar funcionalidade de revogação de consentimento"
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=self.severity,
                        message="Revogação de consentimento é suportada conforme exigido pelo GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"revocation_supported": revocation}
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
            
            if not granular_options:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Opções granulares de consentimento não estão disponíveis, o que é recomendado para conformidade com GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"granular_options": granular_options},
                            "recommendation": "Implementar opções granulares de consentimento para diferentes tipos de processamento"
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="Opções granulares de consentimento estão disponíveis conforme recomendado para GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"granular_options": granular_options}
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
            
            if not audit_trail:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Registro de auditoria de consentimento não está habilitado, o que é importante para demonstrar conformidade com GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"audit_trail": audit_trail},
                            "recommendation": "Implementar registro de auditoria para consentimento"
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="Registro de auditoria de consentimento está habilitado conforme recomendado para GDPR",
                        details={
                            "requirement": "Artigo 7 - Condições para consentimento",
                            "current_config": {"audit_trail": audit_trail}
                        },
                        metadata={
                            "framework": "gdpr",
                            "article": "7",
                            "control_type": "consent"
                        }
                    )
                )
        
        return results
    
    def get_requirements(self) -> List[str]:
        """
        Obtém os IDs dos requisitos que esta regra valida.
        
        Returns:
            Lista de IDs de requisitos
        """
        return ["gdpr_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [
            Region.EU_CENTRAL, 
            Region.EU_NORTH, 
            Region.EU_SOUTH, 
            Region.EU_WEST
        ]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.GDPR]
