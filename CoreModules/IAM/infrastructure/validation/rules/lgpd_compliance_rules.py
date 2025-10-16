"""
INNOVABIZ - Regras de Validação de Compliance LGPD para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Regras de validação específicas para LGPD (Lei Geral de 
           Proteção de Dados) do Brasil.
==================================================================
"""

from enum import Enum
from typing import Dict, List, Any, Optional
from dataclasses import dataclass

from ..compliance_engine import ValidationRule, ValidationContext
from ..compliance_metadata import Region, Industry, ComplianceFramework
from ..models import ComplianceValidationResult, ValidationSeverity, ValidationStatus


class LGPDAuthenticationRule(ValidationRule):
    """Regra para validar autenticação conforme LGPD."""
    
    def __init__(self):
        """Inicializa a regra de autenticação LGPD."""
        super().__init__(
            rule_id="lgpd_authentication",
            name="LGPD Autenticação",
            description="Verifica se os controles de autenticação estão em conformidade com os requisitos da LGPD para proteção de dados pessoais",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com LGPD para autenticação.
        
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
                    message="MFA não está habilitado, o que é recomendado pela LGPD para proteção de dados pessoais",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"mfa_enabled": mfa_enabled}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                        message="MFA está habilitado, mas os métodos podem não ser suficientemente fortes para dados sensíveis sob LGPD",
                        details={
                            "requirement": "Art. 46 - Segurança e sigilo dos dados",
                            "current_config": {"mfa_methods": mfa_methods},
                            "recommendation": "Implementar métodos MFA mais fortes como TOTP, FIDO2, ou biometria"
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "46",
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
                        message="Configuração MFA é adequada para conformidade com LGPD",
                        details={
                            "requirement": "Art. 46 - Segurança e sigilo dos dados",
                            "current_config": {"mfa_methods": mfa_methods}
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "46",
                            "control_type": "authentication"
                        }
                    )
                )
        
        # Verificar políticas de senha
        password_policy = auth_config.get("password_policy", {})
        min_length = password_policy.get("min_length", 0)
        require_complexity = password_policy.get("require_complexity", False)
        
        if min_length < 10 or not require_complexity:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Política de senha pode não ser adequada para proteção de dados sob LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity
                        },
                        "recommendation": "Implementar senhas com pelo menos 10 caracteres e requisitos de complexidade"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Política de senha é adequada para conformidade com LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {
                            "min_length": min_length,
                            "require_complexity": require_complexity
                        }
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Timeout de sessão inadequado para sistemas que processam dados pessoais sob LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"session_timeout": session_timeout},
                        "recommendation": "Implementar timeout de sessão de no máximo 60 minutos"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Timeout de sessão adequado para conformidade com LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"session_timeout": session_timeout}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
                        "control_type": "session_management"
                    }
                )
            )
        
        # Verificar suporte a autenticação específica para contexto brasileiro
        br_specific_auth = auth_config.get("regional_settings", {}).get("br", {}).get("enabled", False)
        
        if not br_specific_auth:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.LOW,
                    message="Configurações específicas para Brasil não estão habilitadas, o que pode ser útil para conformidade contextual com LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"br_specific_auth": br_specific_auth},
                        "recommendation": "Implementar configurações específicas para o contexto brasileiro"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
                        "control_type": "regional_authentication"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.LOW,
                    message="Configurações específicas para Brasil estão habilitadas, ajudando na conformidade contextual com LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"br_specific_auth": br_specific_auth}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
                        "control_type": "regional_authentication"
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
        return ["lgpd_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.BR]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.LGPD]


class LGPDDataProtectionRule(ValidationRule):
    """Regra para validar proteção de dados conforme LGPD."""
    
    def __init__(self):
        """Inicializa a regra de proteção de dados LGPD."""
        super().__init__(
            rule_id="lgpd_data_protection",
            name="LGPD Proteção de Dados",
            description="Verifica se os controles de proteção de dados estão em conformidade com os requisitos da LGPD",
            severity=ValidationSeverity.CRITICAL
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com LGPD para proteção de dados.
        
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
                    message="Criptografia de dados em repouso não está habilitada, o que é necessário pela LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"at_rest_encryption": at_rest_encryption},
                        "recommendation": "Implementar criptografia de dados em repouso"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Criptografia de dados em repouso está habilitada conforme requisitos da LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"at_rest_encryption": at_rest_encryption}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Criptografia de dados em trânsito não está habilitada, o que é necessário pela LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"in_transit_encryption": in_transit_encryption},
                        "recommendation": "Implementar criptografia de dados em trânsito (TLS)"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
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
                    message="Criptografia de dados em trânsito está habilitada conforme requisitos da LGPD",
                    details={
                        "requirement": "Art. 46 - Segurança e sigilo dos dados",
                        "current_config": {"in_transit_encryption": in_transit_encryption}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "46",
                        "control_type": "encryption"
                    }
                )
            )
        
        # Verificar classificação de dados
        data_classification = config.get("data_protection", {}).get("classification", {})
        sensitive_data_classified = data_classification.get("enabled", False)
        
        if not sensitive_data_classified:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.HIGH,
                    message="Classificação de dados sensíveis não está habilitada, o que é importante para conformidade com LGPD",
                    details={
                        "requirement": "Art. 11 - Tratamento de dados pessoais sensíveis",
                        "current_config": {"sensitive_data_classified": sensitive_data_classified},
                        "recommendation": "Implementar classificação de dados para identificar dados sensíveis"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "11",
                        "control_type": "data_classification"
                    }
                )
            )
        else:
            # Verificar se inclui classificação específica para dados sensíveis da LGPD
            lgpd_categories = data_classification.get("categories", {}).get("lgpd_sensitive", False)
            
            if not lgpd_categories:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Classificação de dados não inclui categorias específicas da LGPD para dados sensíveis",
                        details={
                            "requirement": "Art. 11 - Tratamento de dados pessoais sensíveis",
                            "current_config": {"lgpd_categories": lgpd_categories},
                            "recommendation": "Adicionar categorias de classificação específicas da LGPD para dados sensíveis"
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "11",
                            "control_type": "data_classification"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.HIGH,
                        message="Classificação de dados inclui categorias específicas da LGPD para dados sensíveis",
                        details={
                            "requirement": "Art. 11 - Tratamento de dados pessoais sensíveis",
                            "current_config": {"lgpd_categories": lgpd_categories}
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "11",
                            "control_type": "data_classification"
                        }
                    )
                )
        
        # Verificar controles específicos para dados de crianças e adolescentes
        children_data_protection = config.get("data_protection", {}).get("special_categories", {}).get("children", False)
        
        if not children_data_protection:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.HIGH,
                    message="Proteção específica para dados de crianças e adolescentes não está configurada, o que é exigido pela LGPD",
                    details={
                        "requirement": "Art. 14 - Tratamento de dados pessoais de crianças e adolescentes",
                        "current_config": {"children_data_protection": children_data_protection},
                        "recommendation": "Implementar controles específicos para dados de crianças e adolescentes"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "14",
                        "control_type": "special_categories"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.HIGH,
                    message="Proteção específica para dados de crianças e adolescentes está configurada conforme exigido pela LGPD",
                    details={
                        "requirement": "Art. 14 - Tratamento de dados pessoais de crianças e adolescentes",
                        "current_config": {"children_data_protection": children_data_protection}
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "14",
                        "control_type": "special_categories"
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
        return ["lgpd_data_protection"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.BR]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.LGPD]


class LGPDSubjectRightsRule(ValidationRule):
    """Regra para validar direitos do titular conforme LGPD."""
    
    def __init__(self):
        """Inicializa a regra de direitos do titular LGPD."""
        super().__init__(
            rule_id="lgpd_subject_rights",
            name="LGPD Direitos do Titular",
            description="Verifica se os controles para suportar os direitos do titular estão em conformidade com os requisitos da LGPD",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com LGPD para direitos do titular.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar se há funcionalidade de direitos do titular
        subject_rights = config.get("subject_rights", {})
        enabled = subject_rights.get("enabled", False)
        
        if not enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Funcionalidade de direitos do titular não está habilitada, o que é necessário pela LGPD",
                    details={
                        "requirement": "Art. 18 - Direitos do titular",
                        "current_config": {"enabled": enabled},
                        "recommendation": "Implementar funcionalidade de direitos do titular"
                    },
                    metadata={
                        "framework": "lgpd",
                        "article": "18",
                        "control_type": "subject_rights"
                    }
                )
            )
        else:
            # Verificar direitos específicos
            supported_rights = subject_rights.get("supported_rights", [])
            required_rights = [
                "access", "rectification", "deletion", "portability", 
                "information", "revocation_consent", "complaint_anpd"
            ]
            
            missing_rights = [r for r in required_rights if r not in supported_rights]
            
            if missing_rights:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.FAIL,
                        severity=self.severity,
                        message=f"Alguns direitos do titular necessários pela LGPD não são suportados: {', '.join(missing_rights)}",
                        details={
                            "requirement": "Art. 18 - Direitos do titular",
                            "current_config": {"supported_rights": supported_rights},
                            "missing_rights": missing_rights,
                            "recommendation": f"Implementar os direitos do titular faltantes: {', '.join(missing_rights)}"
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "18",
                            "control_type": "subject_rights"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=self.severity,
                        message="Todos os direitos do titular necessários pela LGPD são suportados",
                        details={
                            "requirement": "Art. 18 - Direitos do titular",
                            "current_config": {"supported_rights": supported_rights}
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "18",
                            "control_type": "subject_rights"
                        }
                    )
                )
            
            # Verificar processo automatizado para exercer direitos
            automated_process = subject_rights.get("automated_process", False)
            
            if not automated_process:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Processo automatizado para exercício de direitos do titular não está configurado, o que é recomendado para conformidade com LGPD",
                        details={
                            "requirement": "Art. 18 - Direitos do titular",
                            "current_config": {"automated_process": automated_process},
                            "recommendation": "Implementar processo automatizado para exercício de direitos do titular"
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "18",
                            "control_type": "subject_rights"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="Processo automatizado para exercício de direitos do titular está configurado conforme recomendado pela LGPD",
                        details={
                            "requirement": "Art. 18 - Direitos do titular",
                            "current_config": {"automated_process": automated_process}
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "18",
                            "control_type": "subject_rights"
                        }
                    )
                )
            
            # Verificar tempo de resposta configurado
            response_time_days = subject_rights.get("response_time_days", 0)
            
            if response_time_days <= 0 or response_time_days > 15:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.HIGH,
                        message="Tempo de resposta para solicitações de direitos do titular não está configurado adequadamente conforme LGPD",
                        details={
                            "requirement": "Art. 19 - Confirmação de existência ou acesso a dados",
                            "current_config": {"response_time_days": response_time_days},
                            "recommendation": "Configurar tempo de resposta de no máximo 15 dias conforme LGPD"
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "19",
                            "control_type": "subject_rights"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.HIGH,
                        message="Tempo de resposta para solicitações de direitos do titular está configurado adequadamente conforme LGPD",
                        details={
                            "requirement": "Art. 19 - Confirmação de existência ou acesso a dados",
                            "current_config": {"response_time_days": response_time_days}
                        },
                        metadata={
                            "framework": "lgpd",
                            "article": "19",
                            "control_type": "subject_rights"
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
        return ["lgpd_subject_rights"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.BR]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.LGPD]
