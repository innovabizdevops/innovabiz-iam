"""
INNOVABIZ - Integração do Validador IAM com Healthcare
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação da integração entre o motor de validação IAM
           e o módulo Healthcare, garantindo compliance específico para
           regulamentações de saúde.
==================================================================
"""

import os
import sys
import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple, Union

# Configuração de logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("healthcare_integration")

# Adicionar diretório pai ao path para importação correta
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from validation.compliance_engine import ComplianceEngine, ValidationContext, ValidationResult
from validation.compliance_metadata import Region, Industry, ComplianceFramework
from validation.models import ValidationStatus, ValidationSeverity, ComplianceReport
from validation.validators.base_validator import BaseValidator


class HealthcareValidator(BaseValidator):
    """Validador específico para integração com o módulo Healthcare."""
    
    def __init__(self) -> None:
        """Inicializa o validador com configurações padrão."""
        super().__init__(
            id="healthcare_validator",
            name="Healthcare Validator",
            description="Validador de compliance integrado ao módulo Healthcare",
            frameworks=[
                ComplianceFramework.HIPAA,
                ComplianceFramework.GDPR,
                ComplianceFramework.LGPD
            ],
            regions=[
                Region.USA,
                Region.EU,
                Region.BRAZIL,
                Region.AFRICA
            ],
            industries=[Industry.HEALTHCARE]
        )
        self.healthcare_api_url = os.environ.get("HEALTHCARE_API_URL", "http://localhost:8085/api/healthcare")
        
    def validate(self, context: ValidationContext) -> List[ValidationResult]:
        """
        Executa validação específica para Healthcare.
        
        Args:
            context: Contexto de validação contendo configurações e ambiente
            
        Returns:
            Lista de resultados de validação
        """
        logger.info(f"Iniciando validação Healthcare para região {context.region}")
        results = []
        
        # Obter configurações específicas de Healthcare
        healthcare_config = self._get_healthcare_config(context.tenant_id, context.region)
        
        if not healthcare_config:
            logger.warning(f"Configuração Healthcare não encontrada para tenant {context.tenant_id}")
            return results
        
        # Executar validações específicas por região
        if context.region == Region.USA:
            results.extend(self._validate_hipaa(context, healthcare_config))
        elif context.region == Region.EU:
            results.extend(self._validate_gdpr_healthcare(context, healthcare_config))
        elif context.region == Region.BRAZIL:
            results.extend(self._validate_lgpd_healthcare(context, healthcare_config))
        elif context.region == Region.AFRICA:
            results.extend(self._validate_pndsb(context, healthcare_config))
        
        # Validações comuns a todas as regiões
        results.extend(self._validate_common_healthcare(context, healthcare_config))
        
        return results
    
    def _get_healthcare_config(self, tenant_id: str, region: Region) -> Dict[str, Any]:
        """
        Obtém configuração do módulo Healthcare.
        
        Args:
            tenant_id: ID do tenant
            region: Região para configuração
            
        Returns:
            Configuração do Healthcare ou dicionário vazio
        """
        try:
            import requests
            
            # Formatar região para o padrão do Healthcare
            region_code = region.value.lower()
            
            # Obter configuração via API do módulo Healthcare
            response = requests.get(
                f"{self.healthcare_api_url}/compliance/config",
                params={
                    "tenant_id": tenant_id,
                    "region": region_code
                },
                headers={
                    "Content-Type": "application/json",
                    "X-API-KEY": os.environ.get("HEALTHCARE_API_KEY", "")
                },
                timeout=10
            )
            
            if response.status_code == 200:
                return response.json()
            else:
                logger.error(f"Erro ao obter configuração Healthcare: {response.status_code}")
                return {}
                
        except Exception as e:
            logger.error(f"Erro ao obter configuração Healthcare: {str(e)}")
            return {}
            
    def _validate_hipaa(self, context: ValidationContext, config: Dict[str, Any]) -> List[ValidationResult]:
        """
        Valida conformidade com HIPAA (EUA).
        
        Args:
            context: Contexto de validação
            config: Configuração do Healthcare
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Validação de Privacy Rule
        privacy_result = ValidationResult(
            rule_id="hipaa_privacy_rule",
            status=ValidationStatus.PASSED if config.get("privacy_rule_implemented", False) else ValidationStatus.FAILED,
            details="Verificação da implementação da Privacy Rule do HIPAA",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "HIPAA",
                "section": "Privacy Rule",
                "requirement": "§164.502 - Uses and disclosures of protected health information",
                "recommendation": "Implementar controles de acesso granulares para PHI"
            }
        )
        results.append(privacy_result)
        
        # Validação de Security Rule
        security_controls = config.get("security_rule", {})
        security_passed = all([
            security_controls.get("admin_safeguards", False),
            security_controls.get("physical_safeguards", False),
            security_controls.get("technical_safeguards", False)
        ])
        
        security_result = ValidationResult(
            rule_id="hipaa_security_rule",
            status=ValidationStatus.PASSED if security_passed else ValidationStatus.WARNING,
            details="Verificação da implementação da Security Rule do HIPAA",
            severity=ValidationSeverity.CRITICAL,
            metadata={
                "framework": "HIPAA",
                "section": "Security Rule",
                "requirement": "§164.306 - Security standards",
                "recommendation": "Implementar todas as salvaguardas administrativas, físicas e técnicas"
            }
        )
        results.append(security_result)
        
        # Validação de Breach Notification
        breach_result = ValidationResult(
            rule_id="hipaa_breach_notification",
            status=(ValidationStatus.PASSED if config.get("breach_notification_implemented", False) 
                   else ValidationStatus.FAILED),
            details="Verificação da implementação da regra de notificação de violação do HIPAA",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "HIPAA",
                "section": "Breach Notification Rule",
                "requirement": "§164.400-414 - Notification in the case of breach",
                "recommendation": "Implementar processo de notificação e documentação de violações"
            }
        )
        results.append(breach_result)
        
        return results
        
    def _validate_gdpr_healthcare(self, context: ValidationContext, config: Dict[str, Any]) -> List[ValidationResult]:
        """
        Valida conformidade com GDPR para Healthcare (UE).
        
        Args:
            context: Contexto de validação
            config: Configuração do Healthcare
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Validação de consentimento específico para dados de saúde
        consent_result = ValidationResult(
            rule_id="gdpr_health_consent",
            status=(ValidationStatus.PASSED if config.get("explicit_health_consent", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de consentimento explícito para processamento de dados de saúde",
            severity=ValidationSeverity.CRITICAL,
            metadata={
                "framework": "GDPR",
                "section": "Artigo 9",
                "requirement": "Processamento de categorias especiais de dados pessoais",
                "recommendation": "Implementar processo de consentimento explícito para dados de saúde"
            }
        )
        results.append(consent_result)
        
        # Validação de DPO dedicado para Healthcare
        dpo_result = ValidationResult(
            rule_id="gdpr_healthcare_dpo",
            status=(ValidationStatus.PASSED if config.get("healthcare_dpo_appointed", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de designação de DPO especializado em dados de saúde",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "GDPR",
                "section": "Artigo 37",
                "requirement": "Designação do encarregado da proteção de dados",
                "recommendation": "Designar DPO com experiência em processamento de dados de saúde"
            }
        )
        results.append(dpo_result)
        
        # Validação de DPIA para Healthcare
        dpia_result = ValidationResult(
            rule_id="gdpr_healthcare_dpia",
            status=(ValidationStatus.PASSED if config.get("health_dpia_conducted", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de DPIA para processamento de dados de saúde",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "GDPR",
                "section": "Artigo 35",
                "requirement": "Avaliação de impacto sobre a proteção de dados",
                "recommendation": "Conduzir DPIA específico para dados de saúde"
            }
        )
        results.append(dpia_result)
        
        return results
        
    def _validate_lgpd_healthcare(self, context: ValidationContext, config: Dict[str, Any]) -> List[ValidationResult]:
        """
        Valida conformidade com LGPD para Healthcare (Brasil).
        
        Args:
            context: Contexto de validação
            config: Configuração do Healthcare
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Validação de tratamento de dados sensíveis de saúde
        sensitive_data_result = ValidationResult(
            rule_id="lgpd_health_sensitive_data",
            status=(ValidationStatus.PASSED if config.get("lgpd_health_data_handling", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de conformidade no tratamento de dados sensíveis de saúde",
            severity=ValidationSeverity.CRITICAL,
            metadata={
                "framework": "LGPD",
                "section": "Artigo 11",
                "requirement": "Tratamento de dados pessoais sensíveis",
                "recommendation": "Implementar controles específicos para dados sensíveis de saúde"
            }
        )
        results.append(sensitive_data_result)
        
        # Validação de Relatório de Impacto
        report_result = ValidationResult(
            rule_id="lgpd_health_impact_report",
            status=(ValidationStatus.PASSED if config.get("health_impact_report", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de Relatório de Impacto à Proteção de Dados Pessoais",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "LGPD",
                "section": "Artigo 5, XVII",
                "requirement": "Relatório de impacto à proteção de dados pessoais",
                "recommendation": "Elaborar relatório de impacto específico para o setor de saúde"
            }
        )
        results.append(report_result)
        
        # Validação de integração com Sistema Único de Saúde
        sus_result = ValidationResult(
            rule_id="lgpd_sus_integration",
            status=(ValidationStatus.PASSED if config.get("sus_integration_compliant", False) 
                  else ValidationStatus.NOT_APPLICABLE),
            details="Verificação de conformidade na integração com o Sistema Único de Saúde",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "LGPD",
                "section": "Artigo 11, § 4º",
                "requirement": "Compartilhamento de dados de saúde para finalidades específicas",
                "recommendation": "Implementar controles para conformidade com regulações do SUS"
            }
        )
        results.append(sus_result)
        
        return results
        
    def _validate_pndsb(self, context: ValidationContext, config: Dict[str, Any]) -> List[ValidationResult]:
        """
        Valida conformidade com Política Nacional de Dados de Saúde (Angola).
        
        Args:
            context: Contexto de validação
            config: Configuração do Healthcare
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Validação de integração com sistemas de saúde locais
        local_integration_result = ValidationResult(
            rule_id="pndsb_local_integration",
            status=(ValidationStatus.PASSED if config.get("local_health_systems_integration", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de integração com sistemas de saúde locais",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "PNDSB Angola",
                "section": "Integração de Sistemas",
                "requirement": "Interoperabilidade com sistemas de saúde locais",
                "recommendation": "Implementar interfaces de interoperabilidade adaptadas ao contexto local"
            }
        )
        results.append(local_integration_result)
        
        # Validação de suporte a métodos alternativos
        alternative_methods_result = ValidationResult(
            rule_id="pndsb_alternative_methods",
            status=(ValidationStatus.PASSED if config.get("alternative_health_data_collection", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de suporte a métodos alternativos de coleta de dados de saúde",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "PNDSB Angola",
                "section": "Coleta de Dados",
                "requirement": "Métodos adaptados ao contexto local",
                "recommendation": "Implementar métodos offline e USSD para coleta de dados de saúde"
            }
        )
        results.append(alternative_methods_result)
        
        # Validação de privacidade contextual
        privacy_result = ValidationResult(
            rule_id="pndsb_contextual_privacy",
            status=(ValidationStatus.PASSED if config.get("contextual_privacy_implemented", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de implementação de privacidade contextual para dados de saúde",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "PNDSB Angola",
                "section": "Privacidade",
                "requirement": "Proteção de dados sensíveis adaptada ao contexto",
                "recommendation": "Implementar controles de privacidade adaptados às necessidades locais"
            }
        )
        results.append(privacy_result)
        
        return results
        
    def _validate_common_healthcare(self, context: ValidationContext, config: Dict[str, Any]) -> List[ValidationResult]:
        """
        Executa validações comuns a todas as regiões para Healthcare.
        
        Args:
            context: Contexto de validação
            config: Configuração do Healthcare
            
        Returns:
            Lista de resultados de validação
        """
        results = []
        
        # Validação de interoperabilidade HL7 FHIR
        fhir_result = ValidationResult(
            rule_id="healthcare_fhir_interoperability",
            status=(ValidationStatus.PASSED if config.get("fhir_implemented", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de implementação de interoperabilidade HL7 FHIR",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "Healthcare Standards",
                "section": "Interoperabilidade",
                "requirement": "HL7 FHIR R4",
                "recommendation": "Implementar endpoints FHIR para interoperabilidade de dados de saúde"
            }
        )
        results.append(fhir_result)
        
        # Validação de segurança específica para Healthcare
        security_result = ValidationResult(
            rule_id="healthcare_specific_security",
            status=(ValidationStatus.PASSED if config.get("healthcare_specific_security", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de implementação de controles de segurança específicos para Healthcare",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "Healthcare Security",
                "section": "Segurança",
                "requirement": "Controles de segurança específicos para dados de saúde",
                "recommendation": "Implementar controles de segurança específicos para Healthcare"
            }
        )
        results.append(security_result)
        
        # Validação de auditoria para Healthcare
        audit_result = ValidationResult(
            rule_id="healthcare_audit_trail",
            status=(ValidationStatus.PASSED if config.get("healthcare_audit_trail", False) 
                  else ValidationStatus.FAILED),
            details="Verificação de implementação de trilha de auditoria para Healthcare",
            severity=ValidationSeverity.HIGH,
            metadata={
                "framework": "Healthcare Audit",
                "section": "Auditoria",
                "requirement": "Trilha de auditoria para acesso a dados de saúde",
                "recommendation": "Implementar trilha de auditoria completa para dados de saúde"
            }
        )
        results.append(audit_result)
        
        # Validação de acessibilidade para Healthcare
        accessibility_result = ValidationResult(
            rule_id="healthcare_accessibility",
            status=(ValidationStatus.PASSED if config.get("healthcare_accessibility", False) 
                  else ValidationStatus.WARNING),
            details="Verificação de conformidade de acessibilidade WCAG para Healthcare",
            severity=ValidationSeverity.MEDIUM,
            metadata={
                "framework": "Healthcare Accessibility",
                "section": "Acessibilidade",
                "requirement": "WCAG 2.1 AAA",
                "recommendation": "Implementar conformidade com WCAG 2.1 AAA para interfaces de Healthcare"
            }
        )
        results.append(accessibility_result)
        
        return results


def register_healthcare_validator(engine: ComplianceEngine) -> None:
    """
    Registra o validador Healthcare no motor de compliance.
    
    Args:
        engine: Motor de compliance para registrar o validador
    """
    healthcare_validator = HealthcareValidator()
    engine.register_validator(healthcare_validator)
    logger.info("Validador Healthcare registrado com sucesso")


def generate_healthcare_compliance_report(
    tenant_id: str,
    validation_id: str,
    format: str = "pdf",
    language: str = "pt"
) -> Dict[str, Any]:
    """
    Gera relatório de compliance específico para Healthcare.
    
    Args:
        tenant_id: ID do tenant
        validation_id: ID da validação
        format: Formato do relatório (pdf, html, json)
        language: Idioma do relatório (pt, en)
        
    Returns:
        Informações sobre o relatório gerado
    """
    try:
        import requests
        
        # Obter relatório via API do módulo Healthcare
        healthcare_api_url = os.environ.get("HEALTHCARE_API_URL", "http://localhost:8085/api/healthcare")
        
        response = requests.post(
            f"{healthcare_api_url}/compliance/report",
            json={
                "tenant_id": tenant_id,
                "validation_id": validation_id,
                "format": format,
                "language": language
            },
            headers={
                "Content-Type": "application/json",
                "X-API-KEY": os.environ.get("HEALTHCARE_API_KEY", "")
            },
            timeout=30
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            logger.error(f"Erro ao gerar relatório Healthcare: {response.status_code}")
            return {
                "success": False,
                "error": f"Erro ao gerar relatório: {response.status_code}",
                "reportId": None,
                "reportUrl": None
            }
                
    except Exception as e:
        logger.error(f"Erro ao gerar relatório Healthcare: {str(e)}")
        return {
            "success": False,
            "error": str(e),
            "reportId": None,
            "reportUrl": None
        }


def get_healthcare_compliance_status(
    tenant_id: str,
    hipaa_validation_id: str
) -> Dict[str, Any]:
    """
    Obtém status de compliance HIPAA específico para Healthcare.
    
    Args:
        tenant_id: ID do tenant
        hipaa_validation_id: ID da validação HIPAA
        
    Returns:
        Status de compliance Healthcare
    """
    try:
        import requests
        
        # Obter status via API do módulo Healthcare
        healthcare_api_url = os.environ.get("HEALTHCARE_API_URL", "http://localhost:8085/api/healthcare")
        
        response = requests.get(
            f"{healthcare_api_url}/compliance/status",
            params={
                "tenant_id": tenant_id,
                "hipaa_validation_id": hipaa_validation_id
            },
            headers={
                "Content-Type": "application/json",
                "X-API-KEY": os.environ.get("HEALTHCARE_API_KEY", "")
            },
            timeout=10
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            logger.error(f"Erro ao obter status Healthcare: {response.status_code}")
            return {
                "success": False,
                "error": f"Erro ao obter status: {response.status_code}",
                "status": "UNKNOWN"
            }
                
    except Exception as e:
        logger.error(f"Erro ao obter status Healthcare: {str(e)}")
        return {
            "success": False,
            "error": str(e),
            "status": "ERROR"
        }


if __name__ == "__main__":
    # Teste básico
    engine = ComplianceEngine()
    register_healthcare_validator(engine)
    
    print("Validador Healthcare registrado com sucesso!")
