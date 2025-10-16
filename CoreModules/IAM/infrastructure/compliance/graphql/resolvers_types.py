"""
INNOVABIZ - Tipos GraphQL para Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Definição dos tipos GraphQL para o serviço de validação
           de compliance IAM, com foco em HIPAA para Healthcare.
==================================================================
"""

import graphene
from graphene import Field, List, String, ID, Int, Float, Boolean

from .resolvers import (
    JSONObject, DateTime, 
    ComplianceFrameworkEnum, RegionCodeEnum, ComplianceLevelEnum, ReportFormatEnum,
    ComplianceFrameworkInfo, RegionInfo, ComplianceRequirementType, ValidationResult,
    ReportDownload, ReportGenerationResult, ComplianceReportSummary
)

# ============================================================================
# Tipos para Relatório de Compliance
# ============================================================================

class ComplianceSummary(graphene.ObjectType):
    """Tipo para resumo de compliance"""
    regions_count = Int(required=True)
    frameworks_count = Int(required=True)
    requirements_count = Int(required=True)
    issues_count = Int(required=True)
    critical_issues = Int(required=True)

class OverallCompliance(graphene.ObjectType):
    """Tipo para compliance geral"""
    score = Float(required=True)
    status = String(required=True)
    summary = Field(ComplianceSummary, required=True)

class RequirementsSummary(graphene.ObjectType):
    """Tipo para resumo de requisitos"""
    total = Int(required=True)
    compliant = Int(required=True)
    partially_compliant = Int(required=True)
    non_compliant = Int(required=True)
    not_applicable = Int(required=True)

class RequirementCompliance(graphene.ObjectType):
    """Tipo para compliance de requisito específico"""
    req_id = String(required=True)
    title = String(required=True)
    description = String(required=True)
    category = String(required=True)
    severity = String(required=True)
    status = String(required=True)
    details = String()
    evidence = JSONObject()
    remediation = String()

class ComplianceIssue(graphene.ObjectType):
    """Tipo para problema de compliance"""
    id = ID(required=True)
    requirement_id = String(required=True)
    framework = String(required=True)
    region = String(required=True)
    description = String(required=True)
    status = String(required=True)
    details = String()
    remediation = String()
    severity = String(required=True)

class ComplianceRecommendation(graphene.ObjectType):
    """Tipo para recomendação de compliance"""
    region = String(required=True)
    framework = String(required=True)
    requirement_id = String(required=True)
    priority = String(required=True)
    action = String(required=True)

class FrameworkCompliance(graphene.ObjectType):
    """Tipo para compliance por framework"""
    code = String(required=True)
    name = String(required=True)
    score = Float(required=True)
    status = String(required=True)
    requirements = Field(RequirementsSummary, required=True)
    requirements_details = List(RequirementCompliance, required=True)
    issues = List(ComplianceIssue, required=True)

class RegionCompliance(graphene.ObjectType):
    """Tipo para compliance por região"""
    code = String(required=True)
    name = String(required=True)
    overall_score = Float(required=True)
    status = String(required=True)
    frameworks = List(FrameworkCompliance, required=True)

class ComplianceReport(graphene.ObjectType):
    """Tipo para relatório completo de compliance"""
    id = ID(required=True)
    tenant = ID(required=True)
    timestamp = DateTime(required=True)
    language = String(required=True)
    overall_compliance = Field(OverallCompliance, required=True)
    regions = List(RegionCompliance, required=True)
    frameworks = List(FrameworkCompliance, required=True)
    issues = List(ComplianceIssue, required=True)
    recommendations = List(ComplianceRecommendation, required=True)

class RemediationPlan(graphene.ObjectType):
    """Tipo para plano de remediação"""
    id = ID(required=True)
    tenant = ID(required=True)
    created_at = DateTime(required=True)
    due_date = DateTime()
    assignee = String()
    issues = List(ComplianceIssue, required=True)
    status = String(required=True)

# ============================================================================
# Tipos Específicos para HIPAA Healthcare
# ============================================================================

class HIPAAStats(graphene.ObjectType):
    """Tipo para estatísticas HIPAA"""
    total_requirements = Int(required=True)
    compliant = Int(required=True)
    partially_compliant = Int(required=True)
    non_compliant = Int(required=True)
    overall_score = Float(required=True)

class ARFactorStats(graphene.ObjectType):
    """Tipo para estatísticas de fatores AR"""
    ar_auth_enabled = Boolean(required=True)
    factor_count = Int(required=True)
    enabled_factors = List(String, required=True)
    enhances_phi_security = Boolean(required=True)

class HIPAACategoryStats(graphene.ObjectType):
    """Tipo para estatísticas por categoria HIPAA"""
    category = String(required=True)
    total = Int(required=True)
    compliant = Int(required=True)
    partially_compliant = Int(required=True)
    non_compliant = Int(required=True)
    score = Float(required=True)

class HealthcareRecommendation(graphene.ObjectType):
    """Tipo para recomendação específica para healthcare"""
    category = String(required=True)
    title = String(required=True)
    description = String(required=True)
    priority = String(required=True)
    best_practice_reference = String()

class HIPAAHealthcareReport(graphene.ObjectType):
    """Tipo para relatório específico HIPAA-Healthcare"""
    tenant = ID(required=True)
    timestamp = DateTime(required=True)
    has_healthcare_module = Boolean(required=True)
    hipaa_stats = Field(HIPAAStats)
    ar_factors = Field(ARFactorStats)
    categories = List(HIPAACategoryStats)
    healthcare_recommendations = List(HealthcareRecommendation)
