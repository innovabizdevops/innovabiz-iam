"""
INNOVABIZ - Teste de Integração IAM-Healthcare-HIPAA
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Script de teste para demonstrar a integração entre 
           o validador de compliance HIPAA, o módulo IAM e o módulo Healthcare
==================================================================
"""

import json
import uuid
import logging
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any

from ...iam.compliance.validator import (
    ComplianceFramework, 
    RegionCode, 
    ComplianceLevel,
    ComplianceValidatorFactory,
    MultiRegionComplianceValidator
)

# Configuração de logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("innovabiz.tests.iam_healthcare_hipaa")

# Diretório para output de relatórios
output_dir = Path(__file__).parent / "output"
output_dir.mkdir(exist_ok=True)


def get_sample_healthcare_tenant_config() -> Dict[str, Any]:
    """
    Retorna uma configuração de exemplo para um tenant que utiliza o módulo Healthcare
    """
    return {
        "tenant_id": str(uuid.uuid4()),
        "authentication": {
            "mfa_enabled": True,
            "mfa_methods": ["totp", "sms", "email", "ar_spatial_gesture"],
            "identity_verification": {
                "strong_id_check": True,
                "identity_proofing": True,
                "biometric_verification": True
            }
        },
        "sessions": {
            "inactivity_timeout_minutes": 30,
            "max_session_duration_hours": 12,
            "remember_me_enabled": False,
            "concurrent_sessions_limit": 3
        },
        "modules": {
            "healthcare": {
                "enabled": True,
                "phi_session_timeout_minutes": 15,
                "mfa_required_for_phi": True,
                "phi_access_controls": {
                    "minimum_necessary_principle": True,
                    "data_segmentation": True,
                    "contextual_access": True
                },
                "roles": {
                    "role_separation": True,
                    "physician": ["view_patient", "edit_record", "prescribe"],
                    "nurse": ["view_patient", "update_vitals"],
                    "admin": ["manage_accounts", "view_billing"],
                    "researcher": ["view_anonymized_data"]
                },
                "audit": {
                    "phi_access_logging": True,
                    "log_review_interval_hours": 24,
                    "extended_phi_audit": True
                },
                "emergency_access": True,
                "phi_data_classification": {
                    "enabled": True,
                    "auto_classification": True
                }
            }
        },
        "access_control": {
            "rbac": {
                "enabled": True,
                "default_deny": True,
                "inherited_roles": True
            },
            "abac": {
                "enabled": True,
                "context_aware_access": True
            }
        },
        "audit": {
            "enabled": True,
            "log_retention_days": 365,
            "log_review_enabled": True,
            "tamper_proof_logs": True
        },
        "adaptive_auth": {
            "enabled": True,
            "risk_based_auth": True,
            "anomaly_detection": True,
            "ar_authentication": {
                "enabled": True,
                "spatial_gestures": True,
                "gaze_patterns": True,
                "environment_auth": True
            }
        }
    }


def get_sample_non_healthcare_tenant_config() -> Dict[str, Any]:
    """
    Retorna uma configuração de exemplo para um tenant que NÃO utiliza o módulo Healthcare
    """
    config = get_sample_healthcare_tenant_config()
    if "modules" in config and "healthcare" in config["modules"]:
        config["modules"]["healthcare"]["enabled"] = False
    return config


def test_hipaa_validator_with_healthcare_tenant():
    """
    Testa o validador HIPAA com um tenant que utiliza o módulo Healthcare
    """
    # Configuração do tenant
    tenant_id = uuid.uuid4()
    iam_config = get_sample_healthcare_tenant_config()
    
    # Criação do validador
    hipaa_validator = ComplianceValidatorFactory.create_validator(
        ComplianceFramework.HIPAA, 
        tenant_id
    )
    
    # Execução da validação para região US
    results = hipaa_validator.validate(iam_config, RegionCode.US)
    
    # Análise dos resultados
    compliant_count = 0
    partially_compliant_count = 0
    non_compliant_count = 0
    not_applicable_count = 0
    
    for result in results:
        if result.compliance_level == ComplianceLevel.COMPLIANT:
            compliant_count += 1
        elif result.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT:
            partially_compliant_count += 1
        elif result.compliance_level == ComplianceLevel.NON_COMPLIANT:
            non_compliant_count += 1
        elif result.compliance_level == ComplianceLevel.NOT_APPLICABLE:
            not_applicable_count += 1
    
    logger.info(f"=== HIPAA Healthcare Tenant (US) ===")
    logger.info(f"Compliant: {compliant_count}")
    logger.info(f"Partially Compliant: {partially_compliant_count}")
    logger.info(f"Non-Compliant: {non_compliant_count}")
    logger.info(f"Not Applicable: {not_applicable_count}")
    
    # Salvar resultado em JSON
    results_dict = []
    for result in results:
        results_dict.append({
            "requirement_id": result.requirement.req_id,
            "compliance_level": result.compliance_level.value,
            "details": result.details,
            "remediation": result.remediation if result.remediation else "N/A"
        })
    
    output_file = output_dir / "hipaa_healthcare_us_results.json"
    with open(output_file, "w") as f:
        json.dump(results_dict, f, indent=2)
    
    logger.info(f"Resultados detalhados salvos em: {output_file}")
    
    # Validação para uma região não-US (deve marcar tudo como não aplicável)
    eu_results = hipaa_validator.validate(iam_config, RegionCode.EU)
    eu_not_applicable = sum(1 for r in eu_results if r.compliance_level == ComplianceLevel.NOT_APPLICABLE)
    logger.info(f"=== HIPAA Healthcare Tenant (EU) ===")
    logger.info(f"Not Applicable: {eu_not_applicable} (esperado: todos)")


def test_hipaa_validator_without_healthcare():
    """
    Testa o validador HIPAA com um tenant que NÃO utiliza o módulo Healthcare
    """
    # Configuração do tenant
    tenant_id = uuid.uuid4()
    iam_config = get_sample_non_healthcare_tenant_config()
    
    # Criação do validador
    hipaa_validator = ComplianceValidatorFactory.create_validator(
        ComplianceFramework.HIPAA, 
        tenant_id
    )
    
    # Execução da validação
    results = hipaa_validator.validate(iam_config, RegionCode.US)
    
    # Todos os resultados devem ser marcados como não aplicáveis
    not_applicable_count = sum(1 for r in results if r.compliance_level == ComplianceLevel.NOT_APPLICABLE)
    
    logger.info(f"=== HIPAA Non-Healthcare Tenant (US) ===")
    logger.info(f"Not Applicable: {not_applicable_count} (esperado: todos)")


def test_multi_region_validation():
    """
    Testa a validação multi-regional que inclui HIPAA para EUA
    """
    # Configuração do tenant
    tenant_id = uuid.uuid4()
    iam_config = get_sample_healthcare_tenant_config()
    
    # Criação do validador multi-regional
    validator = MultiRegionComplianceValidator(tenant_id)
    
    # Execução da validação
    all_results = validator.validate_all_regions(iam_config)
    
    # Gerar relatório em português
    report_pt = validator.generate_compliance_report(all_results, "pt")
    output_file_pt = output_dir / "multi_region_report_pt.json"
    with open(output_file_pt, "w") as f:
        json.dump(report_pt, f, indent=2)
    
    # Gerar relatório em inglês
    report_en = validator.generate_compliance_report(all_results, "en")
    output_file_en = output_dir / "multi_region_report_en.json"
    with open(output_file_en, "w") as f:
        json.dump(report_en, f, indent=2)
    
    # Verificar se HIPAA está incluído nos resultados dos EUA
    us_frameworks = list(all_results.get(RegionCode.US, {}).keys())
    has_hipaa = ComplianceFramework.HIPAA in us_frameworks
    
    logger.info(f"=== Multi-Region Validation ===")
    logger.info(f"HIPAA incluído nos resultados dos EUA: {has_hipaa}")
    logger.info(f"Relatório PT: {output_file_pt}")
    logger.info(f"Relatório EN: {output_file_en}")


def main():
    """Função principal para executar os testes"""
    logger.info("Iniciando testes de integração IAM-Healthcare-HIPAA...")
    
    test_hipaa_validator_with_healthcare_tenant()
    test_hipaa_validator_without_healthcare()
    test_multi_region_validation()
    
    logger.info("Testes concluídos com sucesso.")


if __name__ == "__main__":
    main()
