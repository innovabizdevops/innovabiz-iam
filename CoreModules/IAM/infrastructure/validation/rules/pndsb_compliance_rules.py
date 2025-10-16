"""
INNOVABIZ - Regras de Validação de Compliance PNDSB para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Regras de validação específicas para PNDSB (Política Nacional 
           para Desenvolvimento de Serviços Bancários) de Angola.
==================================================================
"""

from enum import Enum
from typing import Dict, List, Any, Optional
from dataclasses import dataclass

from ..compliance_engine import ValidationRule, ValidationContext
from ..compliance_metadata import Region, Industry, ComplianceFramework
from ..models import ComplianceValidationResult, ValidationSeverity, ValidationStatus


class PNDSBAuthenticationRule(ValidationRule):
    """Regra para validar autenticação conforme PNDSB."""
    
    def __init__(self):
        """Inicializa a regra de autenticação PNDSB."""
        super().__init__(
            rule_id="pndsb_authentication",
            name="PNDSB Autenticação",
            description="Verifica se os controles de autenticação estão em conformidade com os requisitos do PNDSB para serviços bancários em Angola",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com PNDSB para autenticação.
        
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
                    message="MFA não está habilitado, o que é exigido pelo PNDSB para transações financeiras em Angola",
                    details={
                        "requirement": "PNDSB Seção 4.2 - Segurança em transações eletrônicas",
                        "current_config": {"mfa_enabled": mfa_enabled}
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "4.2",
                        "control_type": "authentication"
                    }
                )
            )
        else:
            # Verificar se os métodos MFA são adequados para o contexto angolano
            mfa_methods = auth_config.get("mfa_methods", [])
            mobile_mfa = any(m in ["sms", "mobile_app", "ussd"] for m in mfa_methods)
            
            if not mobile_mfa:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="MFA está habilitado, mas não inclui métodos baseados em dispositivos móveis, que são preferenciais no contexto angolano",
                        details={
                            "requirement": "PNDSB Seção 4.2 - Segurança em transações eletrônicas",
                            "current_config": {"mfa_methods": mfa_methods},
                            "recommendation": "Implementar métodos MFA baseados em dispositivos móveis (SMS, USSD ou aplicativo)"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "4.2",
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
                        message="Configuração MFA é adequada para conformidade com PNDSB, incluindo métodos baseados em dispositivos móveis",
                        details={
                            "requirement": "PNDSB Seção 4.2 - Segurança em transações eletrônicas",
                            "current_config": {"mfa_methods": mfa_methods}
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "4.2",
                            "control_type": "authentication"
                        }
                    )
                )
        
        # Verificar autenticação para transações financeiras
        transaction_auth = auth_config.get("transaction_verification", {})
        transaction_auth_enabled = transaction_auth.get("enabled", False)
        
        if not transaction_auth_enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=ValidationSeverity.CRITICAL,
                    message="Verificação específica para transações financeiras não está habilitada, o que é exigido pelo PNDSB",
                    details={
                        "requirement": "PNDSB Seção 4.3 - Verificação de transações",
                        "current_config": {"transaction_auth_enabled": transaction_auth_enabled},
                        "recommendation": "Implementar verificação específica para transações financeiras"
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "4.3",
                        "control_type": "transaction_authentication"
                    }
                )
            )
        else:
            # Verificar métodos de verificação de transação
            transaction_methods = transaction_auth.get("methods", [])
            strong_transaction_auth = any(m in ["otp", "mobile_confirmation", "biometric"] for m in transaction_methods)
            
            if not strong_transaction_auth:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.HIGH,
                        message="Métodos de verificação de transação não são suficientemente fortes conforme exigido pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 4.3 - Verificação de transações",
                            "current_config": {"transaction_methods": transaction_methods},
                            "recommendation": "Implementar métodos fortes de verificação de transação (OTP, confirmação móvel, biometria)"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "4.3",
                            "control_type": "transaction_authentication"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.CRITICAL,
                        message="Métodos de verificação de transação são adequados para conformidade com PNDSB",
                        details={
                            "requirement": "PNDSB Seção 4.3 - Verificação de transações",
                            "current_config": {"transaction_methods": transaction_methods}
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "4.3",
                            "control_type": "transaction_authentication"
                        }
                    )
                )
        
        # Verificar suporte a autenticação offline/baixa conectividade
        offline_auth = auth_config.get("regional_settings", {}).get("angola", {}).get("offline_auth", False)
        
        if not offline_auth:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.MEDIUM,
                    message="Autenticação em cenários de baixa conectividade não está habilitada, o que é recomendado no contexto angolano",
                    details={
                        "requirement": "PNDSB Seção 2.5 - Inclusão financeira",
                        "current_config": {"offline_auth": offline_auth},
                        "recommendation": "Implementar mecanismos de autenticação que funcionem em condições de baixa conectividade"
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "2.5",
                        "control_type": "offline_authentication"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.MEDIUM,
                    message="Autenticação em cenários de baixa conectividade está habilitada, o que é recomendado no contexto angolano",
                    details={
                        "requirement": "PNDSB Seção 2.5 - Inclusão financeira",
                        "current_config": {"offline_auth": offline_auth}
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "2.5",
                        "control_type": "offline_authentication"
                    }
                )
            )
        
        # Verificar suporte a idiomas locais
        local_languages = auth_config.get("regional_settings", {}).get("angola", {}).get("languages", [])
        
        if not local_languages or "portuguese" not in local_languages:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.WARNING,
                    severity=ValidationSeverity.LOW,
                    message="Suporte a idiomas locais não está adequadamente configurado para Angola",
                    details={
                        "requirement": "PNDSB Seção 2.4 - Acessibilidade",
                        "current_config": {"local_languages": local_languages},
                        "recommendation": "Implementar suporte ao idioma português e, idealmente, línguas locais angolanas"
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "2.4",
                        "control_type": "localization"
                    }
                )
            )
        else:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.PASS,
                    severity=ValidationSeverity.LOW,
                    message="Suporte a idiomas locais está adequadamente configurado para Angola",
                    details={
                        "requirement": "PNDSB Seção 2.4 - Acessibilidade",
                        "current_config": {"local_languages": local_languages}
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "2.4",
                        "control_type": "localization"
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
        return ["pndsb_auth"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.AF_ANGOLA]
    
    def get_applicable_industries(self) -> List[Industry]:
        """
        Obtém as indústrias aplicáveis para esta regra.
        
        Returns:
            Lista de indústrias
        """
        return [Industry.FINANCIAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.PNDSB]


class PNDSBInclusaoFinanceiraRule(ValidationRule):
    """Regra para validar controles de inclusão financeira conforme PNDSB."""
    
    def __init__(self):
        """Inicializa a regra de inclusão financeira PNDSB."""
        super().__init__(
            rule_id="pndsb_inclusao_financeira",
            name="PNDSB Inclusão Financeira",
            description="Verifica se os controles de inclusão financeira estão em conformidade com os requisitos do PNDSB para Angola",
            severity=ValidationSeverity.HIGH
        )
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Valida a conformidade com PNDSB para inclusão financeira.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        config = context.config
        
        # Verificar configurações de inclusão financeira
        inclusao_config = config.get("regional_settings", {}).get("angola", {}).get("inclusao_financeira", {})
        enabled = inclusao_config.get("enabled", False)
        
        if not enabled:
            results.append(
                ComplianceValidationResult(
                    rule_id=self.id,
                    status=ValidationStatus.FAIL,
                    severity=self.severity,
                    message="Configurações de inclusão financeira não estão habilitadas, o que é um requisito central do PNDSB",
                    details={
                        "requirement": "PNDSB Seção 2 - Inclusão Financeira",
                        "current_config": {"enabled": enabled},
                        "recommendation": "Habilitar configurações específicas para inclusão financeira em Angola"
                    },
                    metadata={
                        "framework": "pndsb",
                        "section": "2",
                        "control_type": "financial_inclusion"
                    }
                )
            )
        else:
            # Verificar suporte a documentação simplificada para áreas rurais
            simplified_kyc = inclusao_config.get("simplified_kyc", False)
            
            if not simplified_kyc:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="KYC simplificado para áreas rurais não está habilitado, o que é recomendado pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.3 - Acesso simplificado a serviços financeiros",
                            "current_config": {"simplified_kyc": simplified_kyc},
                            "recommendation": "Implementar processo de KYC simplificado para inclusão financeira em áreas rurais"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.3",
                            "control_type": "kyc"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="KYC simplificado para áreas rurais está habilitado, conforme recomendado pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.3 - Acesso simplificado a serviços financeiros",
                            "current_config": {"simplified_kyc": simplified_kyc}
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.3",
                            "control_type": "kyc"
                        }
                    )
                )
            
            # Verificar suporte a agentes bancários
            agent_banking = inclusao_config.get("agent_banking", False)
            
            if not agent_banking:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Suporte a agentes bancários não está habilitado, o que é uma estratégia de inclusão financeira promovida pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.2 - Canais alternativos",
                            "current_config": {"agent_banking": agent_banking},
                            "recommendation": "Implementar suporte a agentes bancários para expandir o acesso a serviços financeiros"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.2",
                            "control_type": "agent_banking"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="Suporte a agentes bancários está habilitado, conforme promovido pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.2 - Canais alternativos",
                            "current_config": {"agent_banking": agent_banking}
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.2",
                            "control_type": "agent_banking"
                        }
                    )
                )
            
            # Verificar suporte a transações com baixo saldo
            low_balance_transactions = inclusao_config.get("low_balance_transactions", False)
            
            if not low_balance_transactions:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.WARNING,
                        severity=ValidationSeverity.MEDIUM,
                        message="Suporte a transações com baixo saldo não está habilitado, o que é importante para inclusão financeira conforme PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.4 - Produtos financeiros inclusivos",
                            "current_config": {"low_balance_transactions": low_balance_transactions},
                            "recommendation": "Implementar suporte a transações com baixo saldo para inclusão financeira"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.4",
                            "control_type": "financial_inclusion"
                        }
                    )
                )
            else:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.PASS,
                        severity=ValidationSeverity.MEDIUM,
                        message="Suporte a transações com baixo saldo está habilitado, conforme recomendado para inclusão financeira pelo PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.4 - Produtos financeiros inclusivos",
                            "current_config": {"low_balance_transactions": low_balance_transactions}
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.4",
                            "control_type": "financial_inclusion"
                        }
                    )
                )
            
            # Verificar integração com sistemas de pagamento móvel
            mobile_money_integration = inclusao_config.get("mobile_money_integration", False)
            
            if not mobile_money_integration:
                results.append(
                    ComplianceValidationResult(
                        rule_id=self.id,
                        status=ValidationStatus.FAIL,
                        severity=ValidationSeverity.HIGH,
                        message="Integração com sistemas de pagamento móvel não está habilitada, o que é crucial para inclusão financeira em Angola conforme PNDSB",
                        details={
                            "requirement": "PNDSB Seção 2.6 - Pagamentos móveis",
                            "current_config": {"mobile_money_integration": mobile_money_integration},
                            "recommendation": "Implementar integração com sistemas de pagamento móvel em Angola"
                        },
                        metadata={
                            "framework": "pndsb",
                            "section": "2.6",
                            "control_type": "mobile_payments"
                        }
                    )
                )
            else:
                # Verificar operadoras de pagamento móvel suportadas
                mobile_operators = inclusao_config.get("mobile_operators", [])
                major_operators = ["unitel", "movicel"]
                
                if not any(op in mobile_operators for op in major_operators):
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            status=ValidationStatus.WARNING,
                            severity=ValidationSeverity.MEDIUM,
                            message="Integração com pagamento móvel não inclui as principais operadoras de Angola",
                            details={
                                "requirement": "PNDSB Seção 2.6 - Pagamentos móveis",
                                "current_config": {"mobile_operators": mobile_operators},
                                "recommendation": f"Adicionar suporte para as principais operadoras: {', '.join(major_operators)}"
                            },
                            metadata={
                                "framework": "pndsb",
                                "section": "2.6",
                                "control_type": "mobile_payments"
                            }
                        )
                    )
                else:
                    results.append(
                        ComplianceValidationResult(
                            rule_id=self.id,
                            status=ValidationStatus.PASS,
                            severity=ValidationSeverity.HIGH,
                            message="Integração com sistemas de pagamento móvel está adequadamente configurada para Angola",
                            details={
                                "requirement": "PNDSB Seção 2.6 - Pagamentos móveis",
                                "current_config": {
                                    "mobile_money_integration": mobile_money_integration,
                                    "mobile_operators": mobile_operators
                                }
                            },
                            metadata={
                                "framework": "pndsb",
                                "section": "2.6",
                                "control_type": "mobile_payments"
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
        return ["pndsb_inclusao_financeira"]
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.AF_ANGOLA]
    
    def get_applicable_industries(self) -> List[Industry]:
        """
        Obtém as indústrias aplicáveis para esta regra.
        
        Returns:
            Lista de indústrias
        """
        return [Industry.FINANCIAL]
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return [ComplianceFramework.PNDSB]
