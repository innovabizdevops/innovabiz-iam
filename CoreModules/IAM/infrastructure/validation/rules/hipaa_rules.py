"""
INNOVABIZ - Regras de Validação HIPAA para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação das regras de validação específicas para 
           conformidade HIPAA no módulo IAM, com foco em proteção
           de informações de saúde (PHI).
==================================================================
"""

import re
import uuid
import datetime
from typing import List, Dict, Any, Optional, Tuple

from ..iam_validator import (
    ValidationResult,
    ValidationStatus,
    ValidationSeverity,
    ValidationType
)

# Versão do módulo
VERSION = "1.0.0"

# Definições de constantes
HIPAA_SECTIONS = {
    "164.312(a)(1)": "Controle de Acesso",
    "164.312(b)": "Trilhas de Auditoria",
    "164.312(c)(1)": "Integridade",
    "164.312(d)": "Autenticação de Pessoa ou Entidade",
    "164.312(e)(1)": "Segurança de Transmissão",
    "164.308(a)(5)(ii)(C)": "Proteção contra Software Malicioso",
    "164.308(a)(7)(ii)": "Plano de Contingência",
    "164.308(a)(1)(ii)(D)": "Análise de Risco",
    "164.308(a)(3)(ii)(A)": "Autorização/Supervisão",
    "164.308(a)(4)": "Gerenciamento de Acesso"
}

# Constantes para autenticação AR 
AR_AUTH_FACTORS = [
    "spatial_gesture",
    "gaze_pattern",
    "environment",
    "biometric"
]

class Validator:
    """
    Validador de conformidade HIPAA para o módulo IAM.
    Implementa as regras específicas para proteção de informações de saúde (PHI).
    """
    
    def __init__(self, tenant_id: str, iam_config: Dict[str, Any]):
        """
        Inicializa o validador HIPAA
        
        Args:
            tenant_id: ID do tenant
            iam_config: Configuração do IAM a ser validada
        """
        self.tenant_id = tenant_id
        self.iam_config = iam_config
        self.timestamp = datetime.datetime.now().isoformat()
    
    def run_validations(self) -> List[ValidationResult]:
        """
        Executa todas as validações HIPAA
        
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Verificar se o módulo Healthcare está habilitado
        has_healthcare = self._has_healthcare_module()
        
        # Validações específicas do HIPAA - Administrative Safeguards
        results.append(self._validate_risk_analysis())
        results.append(self._validate_access_authorization())
        results.append(self._validate_security_awareness())
        results.append(self._validate_contingency_plan())
        
        # Validações específicas do HIPAA - Technical Safeguards
        results.append(self._validate_access_control())
        results.append(self._validate_audit_controls())
        results.append(self._validate_integrity_controls())
        results.append(self._validate_person_authentication())
        results.append(self._validate_transmission_security())
        
        # Validações específicas para módulo Healthcare
        if has_healthcare:
            results.append(self._validate_phi_access_controls())
            results.append(self._validate_minimal_necessary_principle())
            results.append(self._validate_emergency_access())
            results.append(self._validate_phi_encryption())
            results.append(self._validate_phi_authentication())
            
            # Validações AR se aplicável
            if self._has_ar_authentication():
                results.append(self._validate_ar_authentication_factors())
                results.append(self._validate_ar_phi_authorization())
                results.append(self._validate_ar_context_awareness())
        
        return results
    
    def _has_healthcare_module(self) -> bool:
        """Verifica se o módulo Healthcare está habilitado"""
        modules = self.iam_config.get("modules", {})
        return modules.get("healthcare", {}).get("enabled", False)
    
    def _has_ar_authentication(self) -> bool:
        """Verifica se a autenticação AR está habilitada"""
        auth = self.iam_config.get("adaptive_auth", {})
        return auth.get("ar_authentication", {}).get("enabled", False)
    
    def _validate_risk_analysis(self) -> ValidationResult:
        """
        Valida se há análise de risco implementada [§164.308(a)(1)(ii)(A)]
        
        Returns:
            Resultado da validação
        """
        risk_management = self.iam_config.get("risk_management", {})
        risk_analysis_enabled = risk_management.get("risk_analysis_enabled", False)
        last_analysis = risk_management.get("last_risk_analysis")
        
        if not risk_analysis_enabled:
            return ValidationResult(
                id=f"HIPAA-RISK-001-{str(uuid.uuid4())[:8]}",
                name="Análise de Risco HIPAA",
                description="A HIPAA exige que as organizações realizem análises de risco de segurança periodicamente",
                type=ValidationType.COMPLIANCE,
                severity=ValidationSeverity.HIGH,
                status=ValidationStatus.FAILED,
                timestamp=self.timestamp,
                details="A análise de risco não está habilitada na configuração do IAM",
                affected_components=["risk_management"],
                reference="HIPAA §164.308(a)(1)(ii)(A)",
                remediation="Habilitar e configurar a análise de risco periódica no módulo IAM"
            )
        
        # Verificar se a última análise foi há mais de 1 ano
        if last_analysis:
            try:
                last_date = datetime.datetime.fromisoformat(last_analysis)
                one_year_ago = datetime.datetime.now() - datetime.timedelta(days=365)
                
                if last_date < one_year_ago:
                    return ValidationResult(
                        id=f"HIPAA-RISK-002-{str(uuid.uuid4())[:8]}",
                        name="Periodicidade da Análise de Risco",
                        description="A análise de risco deve ser realizada pelo menos anualmente",
                        type=ValidationType.COMPLIANCE,
                        severity=ValidationSeverity.MEDIUM,
                        status=ValidationStatus.WARNING,
                        timestamp=self.timestamp,
                        details=f"A última análise de risco foi realizada em {last_analysis}, há mais de um ano",
                        affected_components=["risk_management"],
                        reference="HIPAA §164.308(a)(1)(ii)(A)",
                        remediation="Realizar nova análise de risco e atualizar a configuração"
                    )
            except (ValueError, TypeError):
                return ValidationResult(
                    id=f"HIPAA-RISK-003-{str(uuid.uuid4())[:8]}",
                    name="Data de Análise de Risco Inválida",
                    description="A data da última análise de risco está em formato inválido",
                    type=ValidationType.COMPLIANCE,
                    severity=ValidationSeverity.LOW,
                    status=ValidationStatus.WARNING,
                    timestamp=self.timestamp,
                    details=f"A data da última análise de risco '{last_analysis}' está em formato inválido",
                    affected_components=["risk_management"],
                    reference="HIPAA §164.308(a)(1)(ii)(A)",
                    remediation="Corrigir o formato da data para ISO 8601 (YYYY-MM-DDThh:mm:ss.sssZ)"
                )
        
        # Validação bem-sucedida
        return ValidationResult(
            id=f"HIPAA-RISK-004-{str(uuid.uuid4())[:8]}",
            name="Análise de Risco HIPAA",
            description="A HIPAA exige que as organizações realizem análises de risco de segurança periodicamente",
            type=ValidationType.COMPLIANCE,
            severity=ValidationSeverity.HIGH,
            status=ValidationStatus.PASSED,
            timestamp=self.timestamp,
            details="A análise de risco está habilitada e atualizada",
            affected_components=["risk_management"],
            reference="HIPAA §164.308(a)(1)(ii)(A)"
        )

    def _validate_access_authorization(self) -> ValidationResult:
        """
        Valida se há processo de autorização de acesso [§164.308(a)(4)]
        
        Returns:
            Resultado da validação
        """
        access_control = self.iam_config.get("access_control", {})
        rbac_enabled = access_control.get("rbac", {}).get("enabled", False)
        abac_enabled = access_control.get("abac", {}).get("enabled", False)
        
        if not rbac_enabled and not abac_enabled:
            return ValidationResult(
                id=f"HIPAA-AUTH-001-{str(uuid.uuid4())[:8]}",
                name="Controle de Acesso HIPAA",
                description="A HIPAA exige implementação de políticas e procedimentos para autorização de acesso",
                type=ValidationType.COMPLIANCE,
                severity=ValidationSeverity.CRITICAL,
                status=ValidationStatus.FAILED,
                timestamp=self.timestamp,
                details="Nenhum sistema de controle de acesso (RBAC ou ABAC) está habilitado",
                affected_components=["access_control"],
                reference="HIPAA §164.308(a)(4)",
                remediation="Habilitar RBAC ou ABAC para controle de acesso"
            )
        
        # Verificar configurações específicas de RBAC
        if rbac_enabled:
            default_deny = access_control.get("rbac", {}).get("default_deny", False)
            if not default_deny:
                return ValidationResult(
                    id=f"HIPAA-AUTH-002-{str(uuid.uuid4())[:8]}",
                    name="Política de Negação por Padrão",
                    description="Deve-se implementar uma política de negação por padrão para acesso seguro",
                    type=ValidationType.COMPLIANCE,
                    severity=ValidationSeverity.HIGH,
                    status=ValidationStatus.WARNING,
                    timestamp=self.timestamp,
                    details="RBAC está habilitado, mas a política de negação por padrão não está ativada",
                    affected_components=["access_control", "rbac"],
                    reference="HIPAA §164.308(a)(4)",
                    remediation="Habilitar a política de negação por padrão para RBAC"
                )
        
        # Validação bem-sucedida
        return ValidationResult(
            id=f"HIPAA-AUTH-003-{str(uuid.uuid4())[:8]}",
            name="Controle de Acesso HIPAA",
            description="A HIPAA exige implementação de políticas e procedimentos para autorização de acesso",
            type=ValidationType.COMPLIANCE,
            severity=ValidationSeverity.CRITICAL,
            status=ValidationStatus.PASSED,
            timestamp=self.timestamp,
            details=f"Sistema de controle de acesso habilitado: {'RBAC' if rbac_enabled else ''}{' e ABAC' if abac_enabled else ''}",
            affected_components=["access_control"],
            reference="HIPAA §164.308(a)(4)"
        )
