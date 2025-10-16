"""
INNOVABIZ - Esquema GraphQL para Certificados de Compliance
===============================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0

Definições de tipos, queries e mutations para o sistema de certificados
de compliance, permitindo geração, validação e consulta de certificados.
===============================================================
"""

import graphene
from graphene import ObjectType, String, Boolean, Int, List, Field, InputObjectType, Enum
from datetime import datetime
import logging
from typing import Dict, Any, List as TypeList, Optional

# Importações internas
from ...common.graphql.scalars import DateTime, JSONScalar
from ..certificates.compliance_certificate_manager import (
    ComplianceCertificateManager, 
    ComplianceCertificate,
    CertificateStatus
)
from ..compliance_engine import ComplianceEngine, ValidationReport
from ..compliance_metadata import ComplianceFramework, Region, Industry
from ...common.security.signature_service import SignatureService
from ...common.storage import ConfigRepository

# Configurar logging
logger = logging.getLogger(__name__)

# Definição de enums
class CertificateStatusEnum(Enum):
    ACTIVE = "ACTIVE"
    EXPIRED = "EXPIRED"
    REVOKED = "REVOKED"
    SUPERSEDED = "SUPERSEDED"

class FrameworkEnum(Enum):
    GDPR = "GDPR"
    LGPD = "LGPD"
    HIPAA = "HIPAA"
    PNDSB = "PNDSB"

class RegionEnum(Enum):
    EU_CENTRAL = "EU_CENTRAL"
    BR = "BR"
    US_EAST = "US_EAST"
    AF_ANGOLA = "AF_ANGOLA"
    GLOBAL = "GLOBAL"

class IndustryEnum(Enum):
    FINANCIAL = "FINANCIAL"
    HEALTHCARE = "HEALTHCARE"
    RETAIL = "RETAIL"
    TELECOM = "TELECOM"
    GOVERNMENT = "GOVERNMENT"
    INSURANCE = "INSURANCE"

# Tipos de input
class CertificateFilterInput(InputObjectType):
    frameworks = List(String, description="Filtrar por frameworks específicos")
    status = List(String, description="Filtrar por status específico")
    region = String(description="Filtrar por região específica")
    
# Tipos GraphQL
class CertificateMetadata(ObjectType):
    compliance_score = Int(description="Pontuação de compliance (0-100)")
    validation_timestamp = String(description="Timestamp da validação")
    validation_scope = String(description="Escopo da validação")
    criticality = String(description="Criticidade do certificado (LOW, MEDIUM, HIGH, CRITICAL)")
    previous_certificate_id = String(description="ID do certificado anterior (se for renovação)")
    renewal_date = String(description="Data de renovação")
    
    # Campos para certificados revogados e substituídos
    revocation = JSONScalar(description="Informações de revogação (se revogado)")
    superseded_by = JSONScalar(description="Informações sobre substituição (se substituído)")

class SignatureMetadata(ObjectType):
    issuer = String(description="Emissor da assinatura")
    method = String(description="Método de assinatura utilizado")
    timestamp = String(description="Timestamp da assinatura")

class CertificateType(ObjectType):
    id = String(description="ID único do certificado")
    tenant_id = String(description="ID do tenant")
    frameworks = List(String, description="Frameworks de compliance cobertos")
    region = String(description="Região aplicável")
    industry = String(description="Indústria aplicável")
    validation_report_id = String(description="ID do relatório de validação")
    issuer = String(description="Emissor do certificado")
    issue_date = DateTime(description="Data de emissão")
    expiry_date = DateTime(description="Data de expiração")
    status = String(description="Status do certificado")
    metadata = Field(CertificateMetadata, description="Metadados do certificado")
    signature = String(description="Assinatura digital do certificado")
    signature_metadata = Field(SignatureMetadata, description="Metadados da assinatura")

class VerificationResult(ObjectType):
    valid = Boolean(description="Indica se o certificado é válido")
    message = String(description="Mensagem descritiva sobre a verificação")
    timestamp = DateTime(description="Timestamp da verificação")

# Queries
class CertificateQueries(ObjectType):
    certificate = Field(
        CertificateType, 
        id=String(required=True),
        description="Obtém um certificado pelo ID"
    )
    
    certificates = List(
        CertificateType,
        tenant_id=String(required=True),
        filters=CertificateFilterInput(),
        description="Lista certificados com filtragem opcional"
    )
    
    expiring_certificates = List(
        CertificateType,
        tenant_id=String(required=False),
        days_threshold=Int(default_value=30),
        description="Lista certificados próximos da expiração"
    )
    
    verify_certificate = Field(
        VerificationResult,
        id=String(required=True),
        description="Verifica a validade de um certificado"
    )
    
    def resolve_certificate(self, info, id):
        """Resolve a query para obter um certificado pelo ID."""
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            
            # Buscar certificado
            certificate = certificate_manager.get_certificate(id)
            if not certificate:
                return None
                
            return convert_certificate_to_graphql(certificate)
        except Exception as e:
            logger.error(f"Erro ao resolver certificate: {e}")
            raise
    
    def resolve_certificates(self, info, tenant_id, filters=None):
        """Resolve a query para listar certificados."""
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            
            # Aplicar filtros se fornecidos
            framework = None
            region = None
            active_only = True
            
            if filters:
                if filters.frameworks and len(filters.frameworks) > 0:
                    framework = filters.frameworks[0]  # Por hora, suporta apenas um framework
                
                if filters.region:
                    region = filters.region
                
                if filters.status and "ACTIVE" not in filters.status:
                    active_only = False
            
            # Buscar certificados
            certificates = certificate_manager.get_certificates_for_tenant(
                tenant_id=tenant_id,
                framework=framework,
                region=region,
                active_only=active_only
            )
            
            # Converter para formato GraphQL
            return [convert_certificate_to_graphql(cert) for cert in certificates]
        except Exception as e:
            logger.error(f"Erro ao resolver certificates: {e}")
            raise
    
    def resolve_expiring_certificates(self, info, days_threshold, tenant_id=None):
        """Resolve a query para listar certificados próximos da expiração."""
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            
            # Buscar certificados expirando
            certificates = certificate_manager.check_expiring_certificates(
                tenant_id=tenant_id,
                days_threshold=days_threshold
            )
            
            # Converter para formato GraphQL
            return [convert_certificate_to_graphql(cert) for cert in certificates]
        except Exception as e:
            logger.error(f"Erro ao resolver expiring_certificates: {e}")
            raise
    
    def resolve_verify_certificate(self, info, id):
        """Resolve a query para verificar um certificado."""
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            
            # Buscar certificado
            certificate = certificate_manager.get_certificate(id)
            if not certificate:
                return VerificationResult(
                    valid=False,
                    message="Certificado não encontrado",
                    timestamp=datetime.utcnow().isoformat()
                )
            
            # Verificar validade
            if not certificate.is_valid():
                status_messages = {
                    "EXPIRED": "Certificado expirado",
                    "REVOKED": "Certificado revogado",
                    "SUPERSEDED": "Certificado substituído por uma versão mais recente"
                }
                
                message = status_messages.get(certificate.status, "Certificado inválido")
                
                return VerificationResult(
                    valid=False,
                    message=message,
                    timestamp=datetime.utcnow().isoformat()
                )
            
            # Verificar assinatura se o serviço de assinatura estiver disponível
            signature_service = context.get('signature_service')
            if signature_service and certificate.signature:
                content = certificate_manager._get_certificate_content_for_signature(certificate)
                if not signature_service.verify(content, certificate.signature):
                    return VerificationResult(
                        valid=False,
                        message="Assinatura digital inválida. Certificado pode ter sido adulterado.",
                        timestamp=datetime.utcnow().isoformat()
                    )
            
            # Certificado válido
            return VerificationResult(
                valid=True,
                message=f"Certificado válido até {certificate.expiry_date.strftime('%d/%m/%Y')}",
                timestamp=datetime.utcnow().isoformat()
            )
        except Exception as e:
            logger.error(f"Erro ao resolver verify_certificate: {e}")
            raise

# Mutations
class GenerateCertificateMutation(graphene.Mutation):
    class Arguments:
        tenant_id = String(required=True)
        validation_report_id = String(required=True)
        valid_days = Int(default_value=365)
    
    certificate = Field(CertificateType)
    
    def mutate(self, info, tenant_id, validation_report_id, valid_days):
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            compliance_engine = context.get('compliance_engine')
            
            # Buscar relatório de validação
            validation_report = compliance_engine.get_validation_report(validation_report_id)
            if not validation_report:
                raise Exception(f"Relatório de validação não encontrado: {validation_report_id}")
            
            # Verificar se relatório indica conformidade
            if validation_report.overall_status != "PASS":
                raise Exception(f"Não é possível gerar certificado para relatório não conforme: {validation_report.overall_status}")
            
            # Gerar certificado
            certificate = certificate_manager.generate_certificate(
                tenant_id=tenant_id,
                validation_report=validation_report,
                valid_days=valid_days
            )
            
            return GenerateCertificateMutation(certificate=convert_certificate_to_graphql(certificate))
        except Exception as e:
            logger.error(f"Erro na mutation GenerateCertificate: {e}")
            raise

class RenewCertificateMutation(graphene.Mutation):
    class Arguments:
        certificate_id = String(required=True)
        validation_report_id = String(required=True)
        valid_days = Int(default_value=365)
    
    certificate = Field(CertificateType)
    
    def mutate(self, info, certificate_id, validation_report_id, valid_days):
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            compliance_engine = context.get('compliance_engine')
            
            # Buscar relatório de validação
            validation_report = compliance_engine.get_validation_report(validation_report_id)
            if not validation_report:
                raise Exception(f"Relatório de validação não encontrado: {validation_report_id}")
            
            # Verificar se relatório indica conformidade
            if validation_report.overall_status != "PASS":
                raise Exception(f"Não é possível renovar com relatório não conforme: {validation_report.overall_status}")
            
            # Renovar certificado
            certificate = certificate_manager.renew_certificate(
                certificate_id=certificate_id,
                validation_report=validation_report,
                valid_days=valid_days
            )
            
            if not certificate:
                raise Exception(f"Falha ao renovar certificado: {certificate_id}")
            
            return RenewCertificateMutation(certificate=convert_certificate_to_graphql(certificate))
        except Exception as e:
            logger.error(f"Erro na mutation RenewCertificate: {e}")
            raise

class RevokeCertificateMutation(graphene.Mutation):
    class Arguments:
        certificate_id = String(required=True)
        reason = String(required=True)
    
    success = Boolean()
    message = String()
    
    def mutate(self, info, certificate_id, reason):
        try:
            # Obter dependências do contexto
            context = info.context
            certificate_manager = get_certificate_manager(context)
            
            # Revogar certificado
            success = certificate_manager.revoke_certificate(
                certificate_id=certificate_id,
                reason=reason
            )
            
            if not success:
                return RevokeCertificateMutation(
                    success=False,
                    message=f"Falha ao revogar certificado: {certificate_id}"
                )
            
            return RevokeCertificateMutation(
                success=True,
                message=f"Certificado revogado com sucesso: {certificate_id}"
            )
        except Exception as e:
            logger.error(f"Erro na mutation RevokeCertificate: {e}")
            raise

class CertificateMutations(ObjectType):
    generate_certificate = GenerateCertificateMutation.Field(
        description="Gera um novo certificado de compliance"
    )
    renew_certificate = RenewCertificateMutation.Field(
        description="Renova um certificado existente"
    )
    revoke_certificate = RevokeCertificateMutation.Field(
        description="Revoga um certificado existente"
    )

# Funções auxiliares
def get_certificate_manager(context) -> ComplianceCertificateManager:
    """
    Obtém ou cria uma instância do gerenciador de certificados.
    
    Args:
        context: Contexto GraphQL
        
    Returns:
        Instância do gerenciador de certificados
    """
    if 'certificate_manager' in context:
        return context.get('certificate_manager')
    
    # Criar dependências necessárias
    config_repository = context.get('config_repository', ConfigRepository())
    signature_service = context.get('signature_service')
    
    # Criar gerenciador
    certificate_manager = ComplianceCertificateManager(
        config_repository=config_repository,
        signature_service=signature_service
    )
    
    # Armazenar no contexto para reutilização
    context['certificate_manager'] = certificate_manager
    
    return certificate_manager

def convert_certificate_to_graphql(certificate: ComplianceCertificate) -> Dict[str, Any]:
    """
    Converte um objeto de certificado para o formato compatível com GraphQL.
    
    Args:
        certificate: Certificado a converter
        
    Returns:
        Dicionário com dados do certificado no formato GraphQL
    """
    cert_dict = certificate.to_dict()
    
    # Ajustar formato dos metadados
    metadata = cert_dict.get('metadata', {})
    signature_metadata = None
    
    if 'signature_metadata' in cert_dict:
        signature_metadata = cert_dict['signature_metadata']
    
    return {
        'id': cert_dict.get('id'),
        'tenant_id': cert_dict.get('tenant_id'),
        'frameworks': cert_dict.get('frameworks', []),
        'region': cert_dict.get('region'),
        'industry': cert_dict.get('industry'),
        'validation_report_id': cert_dict.get('validation_report_id'),
        'issuer': cert_dict.get('issuer'),
        'issue_date': cert_dict.get('issue_date'),
        'expiry_date': cert_dict.get('expiry_date'),
        'status': cert_dict.get('status'),
        'metadata': metadata,
        'signature': cert_dict.get('signature'),
        'signature_metadata': signature_metadata
    }
