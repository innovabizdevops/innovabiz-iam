"""
INNOVABIZ - Sistema de Validação de Compliance para IAM
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Sistema de validação de compliance para o módulo IAM
           com suporte a múltiplas regiões e regulamentações.
==================================================================
"""

import logging
import uuid
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Set, Tuple, Any

# Configuração de logger
logger = logging.getLogger("innovabiz.iam.compliance")


class ComplianceLevel(str, Enum):
    """Níveis de conformidade para requisitos regulatórios"""
    COMPLIANT = "compliant"
    PARTIALLY_COMPLIANT = "partially_compliant"
    NON_COMPLIANT = "non_compliant"
    NOT_APPLICABLE = "not_applicable"


class ComplianceFramework(str, Enum):
    """Frameworks de compliance suportados"""
    GDPR = "gdpr"  # Europa/Portugal
    LGPD = "lgpd"  # Brasil
    CCPA = "ccpa"  # EUA/Califórnia
    HIPAA = "hipaa"  # EUA/Saúde
    SOX = "sox"  # EUA/Financeiro
    PCI_DSS = "pci_dss"  # Global/Pagamentos
    ISO_27001 = "iso_27001"  # Global/Segurança
    PIPEDA = "pipeda"  # Canadá
    PDPA = "pdpa"  # Singapura
    POPIA = "popia"  # África do Sul
    APP = "app"  # Austrália
    APPI = "appi"  # Japão
    NDPR = "ndpr"  # Nigéria
    LAI = "lai"  # Angola
    NIST_800_53 = "nist_800_53"  # EUA/Federal


class RegionCode(str, Enum):
    """Códigos de regiões suportadas"""
    EU = "eu"  # União Europeia
    PT = "pt"  # Portugal
    BR = "br"  # Brasil
    US = "us"  # Estados Unidos
    AO = "ao"  # Angola
    CD = "cd"  # Congo
    GLOBAL = "global"  # Global


class ComplianceRequirement:
    """Requisito de compliance específico"""
    
    def __init__(
        self,
        req_id: str,
        framework: ComplianceFramework,
        description: str,
        description_pt: str,
        category: str,
        severity: str = "high",
        applies_to: Optional[List[str]] = None,
        exemptions: Optional[List[str]] = None,
        technical_controls: Optional[List[str]] = None
    ):
        self.req_id = req_id
        self.framework = framework
        self.description = description
        self.description_pt = description_pt
        self.category = category
        self.severity = severity
        self.applies_to = applies_to or []
        self.exemptions = exemptions or []
        self.technical_controls = technical_controls or []


class ComplianceValidationResult:
    """Resultado da validação de compliance para um requisito"""
    
    def __init__(
        self,
        requirement: ComplianceRequirement,
        compliance_level: ComplianceLevel,
        details: str,
        details_pt: str,
        evidence: Optional[Dict[str, Any]] = None,
        remediation: Optional[str] = None,
        remediation_pt: Optional[str] = None,
        timestamp: datetime = None
    ):
        self.requirement = requirement
        self.compliance_level = compliance_level
        self.details = details
        self.details_pt = details_pt
        self.evidence = evidence or {}
        self.remediation = remediation
        self.remediation_pt = remediation_pt
        self.timestamp = timestamp or datetime.utcnow()


class ComplianceValidator:
    """Validador de compliance base"""
    
    def __init__(self, framework: ComplianceFramework, tenant_id: uuid.UUID):
        self.framework = framework
        self.tenant_id = tenant_id
        self.requirements = self._load_requirements()
        logger.info(f"Inicializado validador para framework {framework.value}")
    
    def _load_requirements(self) -> List[ComplianceRequirement]:
        """Carrega requisitos específicos do framework. Implementado por subclasses."""
        raise NotImplementedError("Subclasses devem implementar este método")
    
    def validate(self, iam_config: Dict, region: RegionCode) -> List[ComplianceValidationResult]:
        """
        Valida a configuração do IAM contra os requisitos do framework.
        Args:
            iam_config: Configuração completa do IAM
            region: Código da região para aplicar requisitos específicos
        Returns:
            Lista de resultados de validação
        """
        raise NotImplementedError("Subclasses devem implementar este método")


class ComplianceValidatorFactory:
    """Fábrica para criar validadores de compliance específicos"""
    
    @staticmethod
    def create_validator(
        framework: ComplianceFramework, 
        tenant_id: uuid.UUID
    ) -> ComplianceValidator:
        """Cria um validador específico para o framework solicitado"""
        if framework == ComplianceFramework.GDPR:
            from .validators.gdpr import GDPRValidator
            return GDPRValidator(framework, tenant_id)
        elif framework == ComplianceFramework.LGPD:
            from .validators.lgpd import LGPDValidator
            return LGPDValidator(framework, tenant_id)
        elif framework == ComplianceFramework.SOX:
            from .validators.sox import SOXValidator
            return SOXValidator(framework, tenant_id)
        elif framework == ComplianceFramework.PCI_DSS:
            from .validators.pci_dss import PCIDSSValidator
            return PCIDSSValidator(framework, tenant_id)
        elif framework == ComplianceFramework.ISO_27001:
            from .validators.iso27001 import ISO27001Validator
            return ISO27001Validator(framework, tenant_id)
        elif framework == ComplianceFramework.HIPAA:
            from .validators.hipaa import HIPAAValidator
            return HIPAAValidator(framework, tenant_id)
        elif framework == ComplianceFramework.NIST_800_53:
            from .validators.nist import NIST80053Validator
            return NIST80053Validator(framework, tenant_id)
        elif framework == ComplianceFramework.LAI:
            from .validators.lai import LAIValidator
            return LAIValidator(framework, tenant_id)
        else:
            raise ValueError(f"Framework não suportado: {framework.value}")


class MultiRegionComplianceValidator:
    """Validador de compliance para múltiplas regiões e frameworks"""
    
    def __init__(self, tenant_id: uuid.UUID):
        self.tenant_id = tenant_id
        self.region_frameworks = self._initialize_region_frameworks()
        logger.info("Validador de compliance multi-região inicializado")
    
    def _initialize_region_frameworks(self) -> Dict[RegionCode, List[ComplianceFramework]]:
        """Inicializa os frameworks aplicáveis por região"""
        return {
            RegionCode.EU: [
                ComplianceFramework.GDPR, 
                ComplianceFramework.PCI_DSS, 
                ComplianceFramework.ISO_27001
            ],
            RegionCode.PT: [
                ComplianceFramework.GDPR, 
                ComplianceFramework.PCI_DSS, 
                ComplianceFramework.ISO_27001
            ],
            RegionCode.BR: [
                ComplianceFramework.LGPD, 
                ComplianceFramework.PCI_DSS, 
                ComplianceFramework.ISO_27001
            ],
            RegionCode.US: [
                ComplianceFramework.SOX, 
                ComplianceFramework.HIPAA, 
                ComplianceFramework.PCI_DSS,
                ComplianceFramework.NIST_800_53
            ],
            RegionCode.AO: [
                ComplianceFramework.LAI, 
                ComplianceFramework.PCI_DSS, 
                ComplianceFramework.ISO_27001
            ],
            RegionCode.CD: [
                ComplianceFramework.LAI, 
                ComplianceFramework.PCI_DSS, 
                ComplianceFramework.ISO_27001
            ]
        }
    
    def validate_all_regions(self, iam_config: Dict) -> Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]]:
        """
        Executa validação de compliance para todas as regiões configuradas.
        
        Args:
            iam_config: Configuração completa do IAM
            
        Returns:
            Dicionário contendo resultados de validação por região e framework
        """
        all_results = {}
        
        for region, frameworks in self.region_frameworks.items():
            region_results = {}
            
            for framework in frameworks:
                try:
                    validator = ComplianceValidatorFactory.create_validator(framework, self.tenant_id)
                    results = validator.validate(iam_config, region)
                    region_results[framework] = results
                    
                    # Log de resumo
                    compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.COMPLIANT)
                    partial = sum(1 for r in results if r.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT)
                    non_compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.NON_COMPLIANT)
                    
                    logger.info(
                        f"Região {region.value}, Framework {framework.value}: "
                        f"{compliant} compliant, {partial} partially compliant, {non_compliant} non-compliant"
                    )
                    
                except Exception as e:
                    logger.error(f"Erro ao validar framework {framework.value} para região {region.value}: {str(e)}")
            
            all_results[region] = region_results
        
        return all_results
    
    def generate_compliance_report(
        self, 
        validation_results: Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]],
        language: str = "pt"
    ) -> Dict:
        """
        Gera relatório de compliance com base nos resultados de validação.
        
        Args:
            validation_results: Resultados de validação por região e framework
            language: Idioma do relatório ('pt' ou 'en')
            
        Returns:
            Relatório estruturado de compliance
        """
        report = {
            "report_id": str(uuid.uuid4()),
            "tenant_id": str(self.tenant_id),
            "timestamp": datetime.utcnow().isoformat(),
            "overall_compliance": {
                "status": None,
                "score": 0.0,
                "summary": {},
            },
            "regions": {},
            "recommendations": []
        }
        
        total_reqs = 0
        total_compliant = 0
        
        # Processa resultados por região
        for region, framework_results in validation_results.items():
            region_data = {
                "frameworks": {},
                "overall_score": 0.0,
                "status": None
            }
            
            region_reqs = 0
            region_compliant = 0
            
            # Processa frameworks por região
            for framework, results in framework_results.items():
                if not results:
                    continue
                    
                framework_data = {
                    "requirements": {
                        "total": len(results),
                        "compliant": sum(1 for r in results if r.compliance_level == ComplianceLevel.COMPLIANT),
                        "partially_compliant": sum(1 for r in results if r.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT),
                        "non_compliant": sum(1 for r in results if r.compliance_level == ComplianceLevel.NON_COMPLIANT),
                        "not_applicable": sum(1 for r in results if r.compliance_level == ComplianceLevel.NOT_APPLICABLE)
                    },
                    "score": 0.0,
                    "issues": []
                }
                
                # Calcula score para o framework
                total_applicable = framework_data["requirements"]["total"] - framework_data["requirements"]["not_applicable"]
                if total_applicable > 0:
                    score = (
                        framework_data["requirements"]["compliant"] + 
                        (framework_data["requirements"]["partially_compliant"] * 0.5)
                    ) / total_applicable
                    framework_data["score"] = round(score * 100, 2)
                
                # Adiciona problemas identificados
                for result in results:
                    if result.compliance_level != ComplianceLevel.COMPLIANT and result.compliance_level != ComplianceLevel.NOT_APPLICABLE:
                        issue = {
                            "requirement_id": result.requirement.req_id,
                            "description": result.requirement.description_pt if language == "pt" else result.requirement.description,
                            "status": result.compliance_level.value,
                            "details": result.details_pt if language == "pt" else result.details,
                            "remediation": result.remediation_pt if language == "pt" else result.remediation
                        }
                        framework_data["issues"].append(issue)
                        
                        # Adiciona recomendação global se for não-conforme
                        if result.compliance_level == ComplianceLevel.NON_COMPLIANT and result.remediation:
                            report["recommendations"].append({
                                "region": region.value,
                                "framework": framework.value,
                                "requirement_id": result.requirement.req_id,
                                "priority": "high" if result.requirement.severity == "high" else "medium",
                                "action": result.remediation_pt if language == "pt" else result.remediation
                            })
                
                region_data["frameworks"][framework.value] = framework_data
                
                # Atualiza contadores regionais
                fw_applicable = total_applicable
                fw_compliant = framework_data["requirements"]["compliant"]
                fw_partial = framework_data["requirements"]["partially_compliant"]
                
                region_reqs += fw_applicable
                region_compliant += fw_compliant + (fw_partial * 0.5)
            
            # Calcula score regional
            if region_reqs > 0:
                region_data["overall_score"] = round((region_compliant / region_reqs) * 100, 2)
                
                if region_data["overall_score"] >= 90:
                    region_data["status"] = "high_compliance"
                elif region_data["overall_score"] >= 75:
                    region_data["status"] = "moderate_compliance"
                else:
                    region_data["status"] = "low_compliance"
            
            report["regions"][region.value] = region_data
            
            # Atualiza contadores globais
            total_reqs += region_reqs
            total_compliant += region_compliant
        
        # Calcula score global
        if total_reqs > 0:
            report["overall_compliance"]["score"] = round((total_compliant / total_reqs) * 100, 2)
            
            if report["overall_compliance"]["score"] >= 90:
                report["overall_compliance"]["status"] = "high_compliance"
            elif report["overall_compliance"]["score"] >= 75:
                report["overall_compliance"]["status"] = "moderate_compliance"
            else:
                report["overall_compliance"]["status"] = "low_compliance"
        
        # Gera resumo global
        report["overall_compliance"]["summary"] = {
            "regions_count": len(validation_results),
            "frameworks_count": len(set(fw for region in validation_results.values() for fw in region.keys())),
            "requirements_count": total_reqs,
            "issues_count": len(report["recommendations"]),
            "critical_issues": sum(1 for rec in report["recommendations"] if rec["priority"] == "high")
        }
        
        return report
