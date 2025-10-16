"""
INNOVABIZ - Schema GraphQL para Validação IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Definição do schema GraphQL para exposição das 
           funcionalidades de validação, certificação e conformidade
           do módulo IAM, seguindo as boas práticas internacionais.
==================================================================
"""

import graphene
import datetime
from enum import Enum
from typing import List, Dict, Any, Optional

from ..iam_validator import (
    IAMValidator, 
    ValidationReport, 
    ValidationResult,
    ValidationStatus,
    ValidationSeverity,
    ValidationType
)
from ..certificate_generator import CertificateGenerator

# Enums GraphQL
class GraphQLValidationStatus(graphene.Enum):
    PASSED = "PASSED"
    FAILED = "FAILED"
    WARNING = "WARNING"
    NOT_APPLICABLE = "NOT_APPLICABLE"


class GraphQLValidationSeverity(graphene.Enum):
    CRITICAL = "CRITICAL"
    HIGH = "HIGH"
    MEDIUM = "MEDIUM"
    LOW = "LOW"
    INFO = "INFO"


class GraphQLValidationType(graphene.Enum):
    SECURITY = "SECURITY"
    COMPLIANCE = "COMPLIANCE"
    PERFORMANCE = "PERFORMANCE"
    INTEGRITY = "INTEGRITY"
    AUTHENTICATION = "AUTHENTICATION"
    AUTHORIZATION = "AUTHORIZATION"
    AUDIT = "AUDIT"
    INTEGRATION = "INTEGRATION"


# Tipos GraphQL
class ValidationResultType(graphene.ObjectType):
    id = graphene.String(description="ID único do resultado de validação")
    name = graphene.String(description="Nome da validação")
    description = graphene.String(description="Descrição da validação")
    type = graphene.Field(GraphQLValidationType, description="Tipo de validação")
    severity = graphene.Field(GraphQLValidationSeverity, description="Severidade da validação")
    status = graphene.Field(GraphQLValidationStatus, description="Status da validação")
    timestamp = graphene.String(description="Timestamp da validação")
    details = graphene.String(description="Detalhes adicionais da validação")
    affected_components = graphene.List(graphene.String, description="Componentes afetados")
    reference = graphene.String(description="Referência normativa ou documentação")
    remediation = graphene.String(description="Instruções de remediação, se aplicável")
    
    @staticmethod
    def from_model(result: ValidationResult) -> 'ValidationResultType':
        """Converte um modelo ValidationResult para um tipo GraphQL"""
        return ValidationResultType(
            id=result.id,
            name=result.name,
            description=result.description,
            type=result.type.value if result.type else None,
            severity=result.severity.value if result.severity else None,
            status=result.status.value if result.status else None,
            timestamp=result.timestamp,
            details=result.details,
            affected_components=result.affected_components,
            reference=result.reference,
            remediation=result.remediation
        )


class ValidationReportType(graphene.ObjectType):
    id = graphene.String(description="ID único do relatório")
    tenant_id = graphene.String(description="ID do tenant")
    framework = graphene.String(description="Framework de validação")
    timestamp = graphene.String(description="Timestamp da validação")
    overall_status = graphene.Field(GraphQLValidationStatus, description="Status geral da validação")
    passed_count = graphene.Int(description="Número de validações aprovadas")
    failed_count = graphene.Int(description="Número de validações falhas")
    warning_count = graphene.Int(description="Número de validações com avisos")
    results = graphene.List(ValidationResultType, description="Resultados de validação")
    
    @staticmethod
    def from_model(report: ValidationReport) -> 'ValidationReportType':
        """Converte um modelo ValidationReport para um tipo GraphQL"""
        return ValidationReportType(
            id=report.id,
            tenant_id=report.tenant_id,
            framework=report.framework,
            timestamp=report.timestamp,
            overall_status=report.overall_status.value if report.overall_status else None,
            passed_count=report.passed_count,
            failed_count=report.failed_count,
            warning_count=report.warning_count,
            results=[ValidationResultType.from_model(result) for result in report.results]
        )


class CertificateType(graphene.ObjectType):
    certificate_id = graphene.String(description="ID único do certificado")
    tenant_id = graphene.String(description="ID do tenant")
    issued_date = graphene.String(description="Data de emissão")
    valid_until = graphene.String(description="Data de validade")
    frameworks = graphene.List(graphene.String, description="Frameworks validados")
    total_validations = graphene.Int(description="Total de validações")
    passed_validations = graphene.Int(description="Total de validações aprovadas")
    compliance_score = graphene.Float(description="Pontuação de conformidade")
    overall_status = graphene.Field(GraphQLValidationStatus, description="Status geral")
    verification_hash = graphene.String(description="Hash de verificação")
    file_path = graphene.String(description="Caminho do arquivo do certificado")


class ValidationSummaryType(graphene.ObjectType):
    tenant_id = graphene.String(description="ID do tenant")
    timestamp = graphene.String(description="Timestamp da validação")
    frameworks = graphene.List(graphene.String, description="Frameworks validados")
    validation_reports = graphene.List(ValidationReportType, description="Relatórios de validação")
    certification = graphene.Field(CertificateType, description="Dados do certificado")
    export_paths = graphene.JSONString(description="Caminhos dos relatórios exportados")


# Queries
class Query(graphene.ObjectType):
    validation_history = graphene.List(
        ValidationSummaryType,
        tenant_id=graphene.String(required=True),
        limit=graphene.Int(default_value=10),
        description="Recupera o histórico de validações para um tenant"
    )
    
    validation_report = graphene.Field(
        ValidationReportType,
        report_id=graphene.String(required=True),
        description="Recupera um relatório de validação específico"
    )
    
    certificate = graphene.Field(
        CertificateType,
        certificate_id=graphene.String(required=True),
        description="Recupera um certificado de conformidade específico"
    )
    
    supported_frameworks = graphene.List(
        graphene.String,
        description="Lista os frameworks de validação suportados"
    )
    
    def resolve_validation_history(self, info, tenant_id, limit):
        """Resolve o histórico de validações para um tenant"""
        # Na implementação real, buscaria do banco de dados
        # Para esta demonstração, retornamos uma lista vazia
        return []
    
    def resolve_validation_report(self, info, report_id):
        """Resolve um relatório de validação específico"""
        # Na implementação real, buscaria do banco de dados
        # Para esta demonstração, retornamos None
        return None
    
    def resolve_certificate(self, info, certificate_id):
        """Resolve um certificado específico"""
        # Na implementação real, buscaria do banco de dados
        # Para esta demonstração, retornamos None
        return None
    
    def resolve_supported_frameworks(self, info):
        """Resolve os frameworks suportados"""
        # Mapeamento de frameworks suportados
        return [
            "hipaa",
            "gdpr",
            "lgpd",
            "pci_dss",
            "security",
            "ar_auth"
        ]


# Input Types
class ValidationResultInput(graphene.InputObjectType):
    framework = graphene.String(required=True, description="Framework para validação")
    include_tests = graphene.Boolean(default_value=False, description="Incluir testes detalhados")


class ExportReportInput(graphene.InputObjectType):
    report_id = graphene.String(required=True, description="ID do relatório a exportar")
    format = graphene.String(required=True, description="Formato de exportação (json, html, pdf)")
    language = graphene.String(required=True, description="Idioma (pt, en)")


class CertificateGenerationInput(graphene.InputObjectType):
    tenant_id = graphene.String(required=True, description="ID do tenant")
    frameworks = graphene.List(graphene.String, required=True, description="Frameworks para validar")
    language = graphene.String(default_value="pt", description="Idioma do certificado")


# Mutations
class RunValidation(graphene.Mutation):
    class Arguments:
        tenant_id = graphene.String(required=True)
        frameworks = graphene.List(graphene.String, required=True)
    
    success = graphene.Boolean()
    reports = graphene.List(ValidationReportType)
    
    def mutate(self, info, tenant_id, frameworks):
        """Executa validação para os frameworks especificados"""
        # Inicializar o validador
        validator = IAMValidator()
        
        # Executar validação
        try:
            reports = validator.validate_all_frameworks(tenant_id, frameworks)
            
            # Converter relatórios
            report_types = [
                ValidationReportType.from_model(report)
                for report in reports.values()
            ]
            
            return RunValidation(success=True, reports=report_types)
        except Exception as e:
            return RunValidation(success=False, reports=[])


class GenerateCertificate(graphene.Mutation):
    class Arguments:
        input = CertificateGenerationInput(required=True)
    
    success = graphene.Boolean()
    certificate = graphene.Field(CertificateType)
    
    def mutate(self, info, input):
        """Gera um certificado de conformidade"""
        # Inicializar o validador e gerador de certificados
        validator = IAMValidator()
        certificate_generator = CertificateGenerator()
        
        try:
            # Executar validação
            reports = validator.validate_all_frameworks(input.tenant_id, input.frameworks)
            
            # Gerar certificado
            certificate_data = certificate_generator.generate_certificate(
                tenant_id=input.tenant_id,
                validation_reports=reports,
                language=input.language
            )
            
            # Converter para tipo GraphQL
            certificate = CertificateType(
                certificate_id=certificate_data["certificate_id"],
                tenant_id=certificate_data["tenant_id"],
                issued_date=certificate_data["issued_date"],
                valid_until=certificate_data["valid_until"],
                frameworks=certificate_data["frameworks"],
                total_validations=certificate_data["validation_summary"]["total_validations"],
                passed_validations=certificate_data["validation_summary"]["passed_validations"],
                compliance_score=certificate_data["validation_summary"]["compliance_score"],
                overall_status=certificate_data["validation_summary"]["overall_status"],
                verification_hash=certificate_data["verification_hash"],
                file_path=certificate_data["file_path"]
            )
            
            return GenerateCertificate(success=True, certificate=certificate)
        except Exception as e:
            return GenerateCertificate(success=False, certificate=None)


class ExportReport(graphene.Mutation):
    class Arguments:
        input = ExportReportInput(required=True)
    
    success = graphene.Boolean()
    file_path = graphene.String()
    
    def mutate(self, info, input):
        """Exporta um relatório em formato específico"""
        # Na implementação real, buscaria o relatório e exportaria
        # Para esta demonstração, retornamos False
        return ExportReport(success=False, file_path=None)


class Mutation(graphene.ObjectType):
    run_validation = RunValidation.Field(description="Executa validação para os frameworks especificados")
    generate_certificate = GenerateCertificate.Field(description="Gera um certificado de conformidade")
    export_report = ExportReport.Field(description="Exporta um relatório em formato específico")


# Schema
schema = graphene.Schema(query=Query, mutation=Mutation)
