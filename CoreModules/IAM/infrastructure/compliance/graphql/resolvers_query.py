"""
INNOVABIZ - Resolvedores de Consulta GraphQL para Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação dos resolvedores de consulta GraphQL para o 
           serviço de validação de compliance IAM, com foco em HIPAA
           para Healthcare.
==================================================================
"""

import json
import uuid
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional, Union

import graphene
from graphene import Field, List as GList, String, ID, Int, Float, Boolean, Enum, ObjectType
from graphql import GraphQLError

from ..validator import (
    ComplianceFramework, 
    RegionCode, 
    ComplianceLevel,
    ComplianceValidatorFactory,
    MultiRegionComplianceValidator
)

from .resolvers import (
    JSONObject, DateTime, 
    ComplianceFrameworkEnum, RegionCodeEnum, ComplianceLevelEnum, ReportFormatEnum,
    ComplianceFrameworkInfo, RegionInfo, ComplianceRequirementType, ValidationResult,
    ReportDownload, ReportGenerationResult, ComplianceReportSummary
)

from .resolvers_types import (
    ComplianceReport, HIPAAHealthcareReport, 
    HIPAAStats, ARFactorStats, HIPAACategoryStats, HealthcareRecommendation
)

# Configuração de logger
logger = logging.getLogger("innovabiz.iam.compliance.graphql.query")

# ============================================================================
# Resolvedores de Consulta
# ============================================================================

class Query(ObjectType):
    """Consultas GraphQL para o serviço de compliance IAM"""
    
    # Definição das consultas
    compliance_report = Field(
        ComplianceReport,
        tenant_id=ID(required=True),
        region=String(),
        framework=String(),
        language=String(default_value="pt")
    )
    
    supported_frameworks = GList(ComplianceFrameworkInfo, required=True)
    
    supported_regions = GList(RegionInfo, required=True)
    
    compliance_requirements = GList(
        ComplianceRequirementType,
        required=True,
        framework=String(required=True),
        language=String(default_value="pt")
    )
    
    hipaa_healthcare_compliance = Field(
        HIPAAHealthcareReport,
        tenant_id=ID(required=True),
        language=String(default_value="pt")
    )
    
    compliance_report_history = GList(
        ComplianceReportSummary,
        required=True,
        tenant_id=ID(required=True),
        limit=Int(default_value=10),
        offset=Int(default_value=0)
    )
    
    # Implementação dos resolvedores
    
    def resolve_compliance_report(self, info, tenant_id, region=None, framework=None, language="pt"):
        """Resolver para obter relatório de compliance"""
        try:
            # Validar entrada
            tenant_uuid = uuid.UUID(tenant_id)
            
            # Converter string para enum se fornecido
            region_code = None
            if region:
                try:
                    region_code = RegionCode(region.lower())
                except ValueError:
                    raise GraphQLError(f"Região inválida: {region}")
            
            framework_enum = None
            if framework:
                try:
                    framework_enum = ComplianceFramework(framework.lower())
                except ValueError:
                    raise GraphQLError(f"Framework inválido: {framework}")
            
            # Acessar serviço de banco de dados para buscar relatório
            # Na implementação real, isso seria buscado de um banco de dados
            # Aqui vamos apenas criar um exemplo para demonstração
            
            # Criar validador
            validator = MultiRegionComplianceValidator(tenant_uuid)
            
            # Carregar configuração do tenant
            config = self._get_tenant_config(tenant_uuid)
            
            # Executar validação
            if region_code and framework_enum:
                # Validação específica
                framework_validator = ComplianceValidatorFactory.create_validator(
                    framework_enum, tenant_uuid
                )
                results = {region_code: {framework_enum: framework_validator.validate(config, region_code)}}
            elif region_code:
                # Validação para região específica
                region_frameworks = validator._initialize_region_frameworks().get(region_code, [])
                results = {region_code: {}}
                for fw in region_frameworks:
                    fw_validator = ComplianceValidatorFactory.create_validator(fw, tenant_uuid)
                    results[region_code][fw] = fw_validator.validate(config, region_code)
            else:
                # Validação completa
                results = validator.validate_all_regions(config)
            
            # Gerar relatório
            report_data = validator.generate_compliance_report(results, language)
            
            # Converter para formato GraphQL
            report_id = str(uuid.uuid4())
            return ComplianceReport(
                id=report_id,
                tenant=tenant_id,
                timestamp=datetime.now(),
                language=language,
                **self._convert_report_data_to_graphql(report_data)
            )
            
        except Exception as e:
            logger.error(f"Erro ao gerar relatório de compliance: {str(e)}")
            raise GraphQLError(f"Erro ao gerar relatório de compliance: {str(e)}")
    
    def resolve_supported_frameworks(self, info):
        """Resolver para listar frameworks suportados"""
        frameworks = []
        
        # Adicionar todos os frameworks suportados
        for framework in ComplianceFramework:
            # Transformar framework em informações detalhadas
            # Na implementação real, isso viria de um banco de dados
            framework_info = self._get_framework_info(framework)
            frameworks.append(ComplianceFrameworkInfo(**framework_info))
        
        return frameworks
    
    def resolve_supported_regions(self, info):
        """Resolver para listar regiões suportadas"""
        regions = []
        
        # Criar validador temporário para acessar mapeamento de regiões
        validator = MultiRegionComplianceValidator(uuid.uuid4())
        region_frameworks = validator._initialize_region_frameworks()
        
        # Para cada região, buscar informações e frameworks aplicáveis
        for region_code, frameworks in region_frameworks.items():
            # Buscar informações da região
            region_info = self._get_region_info(region_code)
            
            # Adicionar frameworks aplicáveis
            framework_infos = []
            for framework in frameworks:
                framework_info = self._get_framework_info(framework)
                framework_infos.append(ComplianceFrameworkInfo(**framework_info))
            
            # Criar objeto de região
            region_info["applicable_frameworks"] = framework_infos
            regions.append(RegionInfo(**region_info))
        
        return regions
    
    def resolve_compliance_requirements(self, info, framework, language="pt"):
        """Resolver para listar requisitos de um framework"""
        try:
            # Converter string para enum
            try:
                framework_enum = ComplianceFramework(framework.lower())
            except ValueError:
                raise GraphQLError(f"Framework inválido: {framework}")
            
            # Criar validador para acessar requisitos
            validator = ComplianceValidatorFactory.create_validator(
                framework_enum, uuid.uuid4()
            )
            
            # Converter requisitos para o formato GraphQL
            requirements = []
            for req in validator.requirements:
                # Usar descrição no idioma solicitado
                description = req.description if language == "en" else req.description_pt
                
                # Criar objeto de requisito
                requirements.append(ComplianceRequirementType(
                    id=req.req_id,
                    code=req.req_id,
                    title=req.req_id,  # Na implementação real, teria um título apropriado
                    description=description,
                    category=req.category,
                    severity=req.severity,
                    framework=ComplianceFrameworkInfo(**self._get_framework_info(framework_enum)),
                    control_type="technical",  # Simplificado para o exemplo
                    validation_logic=None  # Na implementação real, poderia mostrar a lógica
                ))
            
            return requirements
            
        except Exception as e:
            logger.error(f"Erro ao listar requisitos: {str(e)}")
            raise GraphQLError(f"Erro ao listar requisitos: {str(e)}")
    
    def resolve_hipaa_healthcare_compliance(self, info, tenant_id, language="pt"):
        """Resolver para obter relatório específico HIPAA para healthcare"""
        try:
            # Validar entrada
            tenant_uuid = uuid.UUID(tenant_id)
            
            # Carregar configuração do tenant
            config = self._get_tenant_config(tenant_uuid)
            
            # Verificar se módulo healthcare está ativo
            has_healthcare = "healthcare" in config.get("modules", {}) and config["modules"]["healthcare"].get("enabled", False)
            
            # Se não tiver healthcare, retornar relatório básico
            if not has_healthcare:
                return HIPAAHealthcareReport(
                    tenant=tenant_id,
                    timestamp=datetime.now(),
                    has_healthcare_module=False
                )
            
            # Criar validador HIPAA
            hipaa_validator = ComplianceValidatorFactory.create_validator(
                ComplianceFramework.HIPAA, tenant_uuid
            )
            
            # Validar apenas para região US
            results = hipaa_validator.validate(config, RegionCode.US)
            
            # Compilar estatísticas
            hipaa_stats = self._compile_hipaa_stats(results)
            ar_stats = self._compile_ar_stats(config)
            category_stats = self._compile_category_stats(results)
            recommendations = self._generate_healthcare_recommendations(results, language)
            
            # Criar relatório
            return HIPAAHealthcareReport(
                tenant=tenant_id,
                timestamp=datetime.now(),
                has_healthcare_module=True,
                hipaa_stats=HIPAAStats(**hipaa_stats),
                ar_factors=ARFactorStats(**ar_stats),
                categories=[HIPAACategoryStats(**category) for category in category_stats],
                healthcare_recommendations=[HealthcareRecommendation(**rec) for rec in recommendations]
            )
            
        except Exception as e:
            logger.error(f"Erro ao gerar relatório HIPAA-Healthcare: {str(e)}")
            raise GraphQLError(f"Erro ao gerar relatório HIPAA-Healthcare: {str(e)}")
    
    def resolve_compliance_report_history(self, info, tenant_id, limit=10, offset=0):
        """Resolver para obter histórico de relatórios"""
        try:
            # Validar entrada
            tenant_uuid = uuid.UUID(tenant_id)
            
            # Na implementação real, isso buscaria relatórios de um banco de dados
            # Aqui vamos apenas criar alguns exemplos para demonstração
            reports = []
            for i in range(offset, offset + limit):
                if i > 20:  # Limitar a 20 relatórios de exemplo
                    break
                    
                # Criar relatório de exemplo
                report_date = datetime.now() - timedelta(days=i)
                score = 90 - (i * 2)  # Pontuação decrescente para exemplos
                status = "high_compliance"
                if score < 75:
                    status = "moderate_compliance"
                if score < 60:
                    status = "low_compliance"
                
                reports.append(ComplianceReportSummary(
                    id=str(uuid.uuid4()),
                    tenant=tenant_id,
                    timestamp=report_date,
                    region="US" if i % 3 == 0 else ("BR" if i % 3 == 1 else "EU"),
                    framework="HIPAA" if i % 3 == 0 else ("LGPD" if i % 3 == 1 else "GDPR"),
                    overall_compliance_score=score,
                    status=status
                ))
            
            return reports
            
        except Exception as e:
            logger.error(f"Erro ao buscar histórico de relatórios: {str(e)}")
            raise GraphQLError(f"Erro ao buscar histórico de relatórios: {str(e)}")
    
    # Métodos auxiliares
    
    def _get_tenant_config(self, tenant_id):
        """Obtém configuração do tenant"""
        # Na implementação real, isso buscaria a configuração do tenant de um banco de dados
        # Aqui vamos apenas criar um exemplo para demonstração
        return {
            "tenant_id": str(tenant_id),
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
                "inactivity_timeout_minutes": 15,
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
                    "emergency_access": True
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
    
    def _get_framework_info(self, framework):
        """Obtém informações sobre framework"""
        # Na implementação real, isso buscaria informações do framework de um banco de dados
        # Aqui vamos apenas criar alguns exemplos para demonstração
        framework_infos = {
            ComplianceFramework.GDPR: {
                "code": "GDPR",
                "name": "General Data Protection Regulation",
                "name_pt": "Regulamento Geral de Proteção de Dados",
                "description": "EU data protection and privacy regulation",
                "description_pt": "Regulamentação de proteção de dados e privacidade da UE",
                "region_scope": "EU",
                "industry_scope": "all",
                "version": "2016/679",
                "effective_date": datetime(2018, 5, 25),
                "status": "active"
            },
            ComplianceFramework.LGPD: {
                "code": "LGPD",
                "name": "General Data Protection Law",
                "name_pt": "Lei Geral de Proteção de Dados",
                "description": "Brazilian data protection law",
                "description_pt": "Lei brasileira de proteção de dados",
                "region_scope": "BR",
                "industry_scope": "all",
                "version": "13.709/2018",
                "effective_date": datetime(2020, 9, 18),
                "status": "active"
            },
            ComplianceFramework.HIPAA: {
                "code": "HIPAA",
                "name": "Health Insurance Portability and Accountability Act",
                "name_pt": "Lei de Portabilidade e Responsabilidade de Seguros de Saúde",
                "description": "US healthcare data privacy and security regulation",
                "description_pt": "Regulamentação de privacidade e segurança de dados de saúde dos EUA",
                "region_scope": "US",
                "industry_scope": "healthcare",
                "version": "1996 with 2013 HITECH amendments",
                "effective_date": datetime(1996, 8, 21),
                "status": "active"
            },
            ComplianceFramework.PCI_DSS: {
                "code": "PCI_DSS",
                "name": "Payment Card Industry Data Security Standard",
                "name_pt": "Padrão de Segurança de Dados da Indústria de Cartões de Pagamento",
                "description": "Information security standard for payment card processing",
                "description_pt": "Padrão de segurança da informação para processamento de cartões de pagamento",
                "region_scope": "global",
                "industry_scope": "payment",
                "version": "4.0",
                "effective_date": datetime(2022, 3, 31),
                "status": "active"
            }
        }
        
        # Retornar informações para o framework solicitado ou padrão se não encontrado
        return framework_infos.get(framework, {
            "code": framework.value.upper(),
            "name": framework.value.upper(),
            "name_pt": framework.value.upper(),
            "description": f"Compliance framework {framework.value}",
            "description_pt": f"Framework de compliance {framework.value}",
            "region_scope": "global",
            "industry_scope": "all",
            "status": "active"
        })
    
    def _get_region_info(self, region_code):
        """Obtém informações sobre região"""
        # Na implementação real, isso buscaria informações da região de um banco de dados
        # Aqui vamos apenas criar alguns exemplos para demonstração
        region_infos = {
            RegionCode.EU: {
                "code": "EU",
                "name": "European Union",
                "name_pt": "União Europeia"
            },
            RegionCode.PT: {
                "code": "PT",
                "name": "Portugal",
                "name_pt": "Portugal"
            },
            RegionCode.BR: {
                "code": "BR",
                "name": "Brazil",
                "name_pt": "Brasil"
            },
            RegionCode.US: {
                "code": "US",
                "name": "United States",
                "name_pt": "Estados Unidos"
            },
            RegionCode.AO: {
                "code": "AO",
                "name": "Angola",
                "name_pt": "Angola"
            },
            RegionCode.CD: {
                "code": "CD",
                "name": "Democratic Republic of the Congo",
                "name_pt": "República Democrática do Congo"
            },
            RegionCode.GLOBAL: {
                "code": "GLOBAL",
                "name": "Global",
                "name_pt": "Global"
            }
        }
        
        # Retornar informações para a região solicitada ou padrão se não encontrada
        return region_infos.get(region_code, {
            "code": region_code.value.upper(),
            "name": region_code.value.upper(),
            "name_pt": region_code.value.upper()
        })
    
    def _convert_report_data_to_graphql(self, report_data):
        """Converte dados do relatório para formato GraphQL"""
        # Na implementação real, isso faria uma conversão mais complexa
        # Aqui vamos simplificar para o exemplo
        return {
            "overall_compliance": {
                "score": report_data["overall_compliance"]["score"],
                "status": report_data["overall_compliance"]["status"],
                "summary": report_data["overall_compliance"]["summary"]
            },
            "regions": [
                {
                    "code": region_code,
                    "name": self._get_region_info(RegionCode(region_code))["name"],
                    "overall_score": region_data["overall_score"],
                    "status": region_data["status"],
                    "frameworks": []  # Simplificado para o exemplo
                }
                for region_code, region_data in report_data["regions"].items()
            ],
            "frameworks": [],  # Simplificado para o exemplo
            "issues": [],  # Simplificado para o exemplo
            "recommendations": []  # Simplificado para o exemplo
        }
    
    def _compile_hipaa_stats(self, results):
        """Compila estatísticas HIPAA"""
        # Contar resultados por nível de compliance
        compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.COMPLIANT)
        partially = sum(1 for r in results if r.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT)
        non_compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.NON_COMPLIANT)
        total = compliant + partially + non_compliant
        
        # Calcular pontuação
        score = 0
        if total > 0:
            score = round(((compliant + (partially * 0.5)) / total) * 100, 2)
        
        return {
            "total_requirements": total,
            "compliant": compliant,
            "partially_compliant": partially,
            "non_compliant": non_compliant,
            "overall_score": score
        }
    
    def _compile_ar_stats(self, config):
        """Compila estatísticas de fatores AR"""
        # Extrair informações sobre autenticação AR
        adaptive_auth = config.get("adaptive_auth", {})
        ar_auth = adaptive_auth.get("ar_authentication", {})
        
        ar_factors = {
            "ar_spatial_gesture": ar_auth.get("spatial_gestures", False),
            "ar_gaze_pattern": ar_auth.get("gaze_patterns", False),
            "ar_environment": ar_auth.get("environment_auth", False),
            "ar_biometric": ar_auth.get("biometric", False)
        }
        
        ar_factor_count = sum(1 for factor, enabled in ar_factors.items() if enabled)
        
        return {
            "ar_auth_enabled": adaptive_auth.get("enabled", False) and ar_auth.get("enabled", False),
            "factor_count": ar_factor_count,
            "enabled_factors": [factor for factor, enabled in ar_factors.items() if enabled],
            "enhances_phi_security": ar_factor_count >= 2 and adaptive_auth.get("enabled", False) and ar_auth.get("enabled", False)
        }
    
    def _compile_category_stats(self, results):
        """Compila estatísticas por categoria"""
        # Agrupar resultados por categoria
        categories = {}
        for result in results:
            category = result.requirement.category
            
            if category not in categories:
                categories[category] = {
                    "compliant": 0,
                    "partially_compliant": 0,
                    "non_compliant": 0,
                    "total": 0
                }
            
            categories[category]["total"] += 1
            
            if result.compliance_level == ComplianceLevel.COMPLIANT:
                categories[category]["compliant"] += 1
            elif result.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT:
                categories[category]["partially_compliant"] += 1
            elif result.compliance_level == ComplianceLevel.NON_COMPLIANT:
                categories[category]["non_compliant"] += 1
        
        # Converter para lista de estatísticas por categoria
        category_stats = []
        for category, stats in categories.items():
            score = 0
            if stats["total"] > 0:
                score = round(((stats["compliant"] + (stats["partially_compliant"] * 0.5)) / stats["total"]) * 100, 2)
            
            category_stats.append({
                "category": category,
                "total": stats["total"],
                "compliant": stats["compliant"],
                "partially_compliant": stats["partially_compliant"],
                "non_compliant": stats["non_compliant"],
                "score": score
            })
        
        return category_stats
    
    def _generate_healthcare_recommendations(self, results, language):
        """Gera recomendações para o contexto healthcare"""
        # Extrair recomendações dos resultados
        recommendations = []
        
        # Adicionar recomendações para requisitos não-conformes
        for result in results:
            if result.compliance_level == ComplianceLevel.NON_COMPLIANT and result.remediation:
                remediation = result.remediation_pt if language == "pt" else result.remediation
                recommendations.append({
                    "category": result.requirement.category,
                    "title": f"Remediar {result.requirement.req_id}",
                    "description": remediation,
                    "priority": "high" if result.requirement.severity == "high" else "medium",
                    "best_practice_reference": "HIPAA Security Rule §164.312"
                })
        
        # Se não houver não-conforme, adicionar recomendações para parcialmente conformes
        if not recommendations:
            for result in results:
                if result.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT and result.remediation:
                    remediation = result.remediation_pt if language == "pt" else result.remediation
                    recommendations.append({
                        "category": result.requirement.category,
                        "title": f"Melhorar {result.requirement.req_id}",
                        "description": remediation,
                        "priority": "medium",
                        "best_practice_reference": "HIPAA Security Rule §164.312"
                    })
        
        # Adicionar recomendações gerais
        general_recommendations = {
            "pt": [
                {
                    "category": "best_practice",
                    "title": "Implementar Autenticação Multi-fator para Todos os Acessos a PHI",
                    "description": "Recomenda-se fortemente que o MFA seja obrigatório para todos os acessos a informações de saúde protegidas (PHI), utilizando pelo menos dois fatores independentes.",
                    "priority": "high",
                    "best_practice_reference": "NIST SP 800-63B"
                },
                {
                    "category": "integration",
                    "title": "Integrar com Mecanismos de Gestão de Consentimento",
                    "description": "Integrar o módulo Healthcare com mecanismos robustos de gestão de consentimento para garantir que os pacientes possam gerenciar detalhadamente quem tem acesso a seus dados de saúde.",
                    "priority": "medium",
                    "best_practice_reference": "OCR HIPAA Guidance"
                }
            ],
            "en": [
                {
                    "category": "best_practice",
                    "title": "Implement Multi-factor Authentication for All PHI Access",
                    "description": "It is strongly recommended that MFA be mandatory for all access to protected health information (PHI), using at least two independent factors.",
                    "priority": "high",
                    "best_practice_reference": "NIST SP 800-63B"
                },
                {
                    "category": "integration",
                    "title": "Integrate with Consent Management Mechanisms",
                    "description": "Integrate the Healthcare module with robust consent management mechanisms to ensure patients can granularly manage who has access to their health data.",
                    "priority": "medium",
                    "best_practice_reference": "OCR HIPAA Guidance"
                }
            ]
        }
        
        # Adicionar recomendações gerais com base no idioma
        recommendations.extend(general_recommendations.get(language, general_recommendations["en"]))
        
        return recommendations
