"""
INNOVABIZ - Regras de Validação de Compliance HIPAA para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Regras de validação específicas para HIPAA (Health Insurance 
           Portability and Accountability Act) dos EUA.
==================================================================
"""

from enum import Enum
from typing import Dict, List, Any, Optional
from dataclasses import dataclass

from ..compliance_engine import ValidationRule, ValidationContext
from ..compliance_metadata import Region, Industry, ComplianceFramework
from ..models import ComplianceValidationResult, ValidationSeverity, ValidationStatus


class HIPAAAuthenticationRule(ValidationRule):
    """Regra para validar autenticação conforme HIPAA."""
    
    def __init__(self):
        """Inicializa a regra de autenticação HIPAA."""
        super().__init__(
            rule_id="hipaa_authentication",
            name="HIPAA Authentication",
            description="Verifica se os controles de autenticação estão em conformidade com os requisitos de HIPAA para acesso a informações de saúde protegidas (PHI)",
            severity=ValidationSeverity.CRITICAL
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com HIPAA para autenticação.
        
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
                    message="MFA não está habilitado, o que é necessário pelo HIPAA para acesso a PHI",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(d) - Person or Entity Authentication",
                        "current_config": {"mfa_enabled": mfa_enabled},
                        "recommendation": "Implementar autenticação multi-fator para garantir que apenas pessoas autorizadas tenham acesso a PHI"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(d)",
                        "control_type": "authentication"
                    }
                )
            )
        else:
            # Verificar se os métodos MFA são adequados para HIPAA
            mfa_methods = auth_config.get("mfa_methods", [])
            strong_mfa = any(m in ["totp", "fido2", "smart_card", "biometric"] for m in mfa_methods)
            
            if not strong_mfa:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.HIGH,
                        message="MFA está habilitado, mas os métodos podem não ser suficientemente fortes para HIPAA",
                        details={
                            "requirement": "HIPAA Security Rule § 164.312(d) - Person or Entity Authentication",
                            "current_config": {"mfa_methods": mfa_methods},
                            "recommendation": "Implementar métodos MFA mais fortes como TOTP, FIDO2, ou biometria"
                        },
                        metadata={
                            "framework": "hipaa",
                            "section": "164.312(d)",
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
                        message="Configuração MFA é adequada para conformidade com HIPAA",
                        details={
                            "requirement": "HIPAA Security Rule § 164.312(d) - Person or Entity Authentication",
                            "current_config": {"mfa_methods": mfa_methods}
                        },
                        metadata={
                            "framework": "hipaa",
                            "section": "164.312(d)",
                            "control_type": "authentication"
                        }
                    )
                )
        
        # Verificar políticas de senha
        password_policy = auth_config.get("password_policy", {})
        min_length = password_policy.get("min_length", 0)
        require_complexity = password_policy.get("require_complexity", False)
        password_history = password_policy.get("password_history", 0)
        
        # HIPAA geralmente requer políticas de senha robustas
        if min_length < 8 or not require_complexity or password_history < 6:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Política de senha não atende aos requisitos comuns para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(5) - Security Awareness and Training",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity,
                            "password_history": password_history
                        },
                        "recommendation": "Implementar senha com mínimo de 8 caracteres, requisitos de complexidade e histórico de pelo menos 6 senhas"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(5)",
                        "control_type": "authentication"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.HIGH,
                    message="Política de senha atende aos requisitos comuns para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(5) - Security Awareness and Training",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity,
                            "password_history": password_history
                        }
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(5)",
                        "control_type": "authentication"
                    }
                )
            )
        
        # Verificar timeout de sessão
        session_config = auth_config.get("session", {})
        session_timeout = session_config.get("timeout_minutes", 0)
        auto_logout = session_config.get("auto_logout", False)
        
        # HIPAA geralmente requer timeout de sessão mais curto para PHI
        if not auto_logout or session_timeout <= 0 or session_timeout > 30:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Timeout de sessão não está configurado adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(2)(iii) - Automatic logoff",
                        "current_config": {
                            "session_timeout": session_timeout,
                            "auto_logout": auto_logout
                        },
                        "recommendation": "Implementar timeout de sessão automático de no máximo 30 minutos"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(2)(iii)",
                        "control_type": "session_management"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.HIGH,
                    message="Timeout de sessão está configurado adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(2)(iii) - Automatic logoff",
                        "current_config": {
                            "session_timeout": session_timeout,
                            "auto_logout": auto_logout
                        }
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(2)(iii)",
                        "control_type": "session_management"
                    }
                )
            )
        
        # Verificar bloqueio de conta
        account_lockout = auth_config.get("account_lockout", {})
        lockout_enabled = account_lockout.get("enabled", False)
        lockout_threshold = account_lockout.get("threshold", 0)
        
        if not lockout_enabled or lockout_threshold > 5:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="Configuração de bloqueio de conta não é adequada para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(5) - Security Awareness and Training",
                        "current_config": {
                            "lockout_enabled": lockout_enabled,
                            "lockout_threshold": lockout_threshold
                        },
                        "recommendation": "Implementar bloqueio de conta após no máximo 5 tentativas malsucedidas"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(5)",
                        "control_type": "account_lockout"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Configuração de bloqueio de conta é adequada para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(5) - Security Awareness and Training",
                        "current_config": {
                            "lockout_enabled": lockout_enabled,
                            "lockout_threshold": lockout_threshold
                        }
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(5)",
                        "control_type": "account_lockout"
                    }
                )
            )
        
        # Verificar autenticação para dispositivos móveis
        mobile_auth = auth_config.get("mobile", {})
        mobile_mfa_enabled = mobile_auth.get("mfa_enabled", False)
        device_pin_required = mobile_auth.get("device_pin_required", False)
        
        if not mobile_mfa_enabled or not device_pin_required:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Autenticação para dispositivos móveis não está configurada adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(d) - Person or Entity Authentication",
                        "current_config": {
                            "mobile_mfa_enabled": mobile_mfa_enabled,
                            "device_pin_required": device_pin_required
                        },
                        "recommendation": "Implementar MFA para acesso móvel e exigir PIN/biometria no dispositivo"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(d)",
                        "control_type": "mobile_authentication"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Autenticação para dispositivos móveis está configurada adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(d) - Person or Entity Authentication",
                        "current_config": {
                            "mobile_mfa_enabled": mobile_mfa_enabled,
                            "device_pin_required": device_pin_required
                        }
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(d)",
                        "control_type": "mobile_authentication"
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
        return ["hipaa_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.US_EAST, Region.US_WEST]
    
    def get_applicable_industries(self) -> List[Industry]:
        """
        Obtém as indústrias aplicáveis para esta regra.
        
        Returns:
            Lista de indústrias
        """
        return [Industry.HEALTHCARE]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.HIPAA]


class HIPAAAccessControlRule(ValidationRule):
    """Regra para validar controles de acesso conforme HIPAA."""
    
    def __init__(self):
        """Inicializa a regra de controle de acesso HIPAA."""
        super().__init__(
            rule_id="hipaa_access_control",
            name="HIPAA Access Control",
            description="Verifica se os controles de acesso estão em conformidade com os requisitos de HIPAA para informações de saúde protegidas (PHI)",
            severity=ValidationSeverity.CRITICAL
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com HIPAA para controles de acesso.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar controles de acesso básicos
        access_control = config.get("access_control", {})
        rbac_enabled = access_control.get("rbac_enabled", False)
        
        if not rbac_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Controle de acesso baseado em papéis (RBAC) não está habilitado, o que é necessário pelo HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(1) - Access Control",
                        "current_config": {"rbac_enabled": rbac_enabled},
                        "recommendation": "Implementar RBAC para controlar acesso a PHI"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(1)",
                        "control_type": "access_control"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=self.severity,
                    message="Controle de acesso baseado em papéis (RBAC) está habilitado conforme necessário pelo HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(1) - Access Control",
                        "current_config": {"rbac_enabled": rbac_enabled}
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(1)",
                        "control_type": "access_control"
                    }
                )
            )
        
        # Verificar identificadores únicos
        unique_identifiers = access_control.get("unique_identifiers", False)
        
        if not unique_identifiers:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Identificadores únicos não estão habilitados, o que é necessário pelo HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(2)(i) - Unique User Identification",
                        "current_config": {"unique_identifiers": unique_identifiers},
                        "recommendation": "Implementar identificadores únicos para cada usuário"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(2)(i)",
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
                    message="Identificadores únicos estão habilitados conforme necessário pelo HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(2)(i) - Unique User Identification",
                        "current_config": {"unique_identifiers": unique_identifiers}
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(2)(i)",
                        "control_type": "access_control"
                    }
                )
            )
        
        # Verificar acesso de emergência
        emergency_access = access_control.get("emergency_access", {})
        emergency_enabled = emergency_access.get("enabled", False)
        
        if not emergency_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Acesso de emergência não está habilitado, o que é necessário pelo HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.312(a)(2)(ii) - Emergency Access Procedure",
                        "current_config": {"emergency_enabled": emergency_enabled},
                        "recommendation": "Implementar procedimentos de acesso de emergência"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.312(a)(2)(ii)",
                        "control_type": "emergency_access"
                    }
                )
            )
        else:
            # Verificar auditoria de acesso de emergência
            emergency_audit = emergency_access.get("audit_enabled", False)
            
            if not emergency_audit:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Auditoria de acesso de emergência não está habilitada, o que é recomendado para HIPAA",
                        details={
                            "requirement": "HIPAA Security Rule § 164.312(a)(2)(ii) - Emergency Access Procedure",
                            "current_config": {"emergency_audit": emergency_audit},
                            "recommendation": "Implementar auditoria para acesso de emergência"
                        },
                        metadata={
                            "framework": "hipaa",
                            "section": "164.312(a)(2)(ii)",
                            "control_type": "emergency_access"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.HIGH,
                        message="Acesso de emergência está configurado adequadamente para HIPAA",
                        details={
                            "requirement": "HIPAA Security Rule § 164.312(a)(2)(ii) - Emergency Access Procedure",
                            "current_config": {
                                "emergency_enabled": emergency_enabled,
                                "emergency_audit": emergency_audit
                            }
                        },
                        metadata={
                            "framework": "hipaa",
                            "section": "164.312(a)(2)(ii)",
                            "control_type": "emergency_access"
                        }
                    )
                )
        
        # Verificar revisão de acesso
        access_review = access_control.get("access_review", {})
        review_enabled = access_review.get("enabled", False)
        review_frequency_days = access_review.get("frequency_days", 0)
        
        if not review_enabled or review_frequency_days <= 0 or review_frequency_days > 90:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.MEDIUM,
                    message="Revisão de acesso não está configurada adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(3)(ii)(B) - Workforce clearance procedure",
                        "current_config": {
                            "review_enabled": review_enabled,
                            "review_frequency_days": review_frequency_days
                        },
                        "recommendation": "Implementar revisão de acesso pelo menos trimestralmente (90 dias)"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(3)(ii)(B)",
                        "control_type": "access_review"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Revisão de acesso está configurada adequadamente para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(3)(ii)(B) - Workforce clearance procedure",
                        "current_config": {
                            "review_enabled": review_enabled,
                            "review_frequency_days": review_frequency_days
                        }
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(3)(ii)(B)",
                        "control_type": "access_review"
                    }
                )
            )
        
        # Verificar segregação de funções
        segregation_of_duties = access_control.get("segregation_of_duties", False)
        
        if not segregation_of_duties:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Segregação de funções não está habilitada, o que é recomendado para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(3) - Workforce Security",
                        "current_config": {"segregation_of_duties": segregation_of_duties},
                        "recommendation": "Implementar segregação de funções para reduzir risco de uso indevido de PHI"
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(3)",
                        "control_type": "segregation_of_duties"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Segregação de funções está habilitada conforme recomendado para HIPAA",
                    details={
                        "requirement": "HIPAA Security Rule § 164.308(a)(3) - Workforce Security",
                        "current_config": {"segregation_of_duties": segregation_of_duties}
                    },
                    metadata={
                        "framework": "hipaa",
                        "section": "164.308(a)(3)",
                        "control_type": "segregation_of_duties"
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
        return ["hipaa_access_control"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.US_EAST, Region.US_WEST]
    
    def get_applicable_industries(self) -> List[Industry]:
        """
        Obtém as indústrias aplicáveis para esta regra.
        
        Returns:
            Lista de indústrias
        """
        return [Industry.HEALTHCARE]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.HIPAA]
