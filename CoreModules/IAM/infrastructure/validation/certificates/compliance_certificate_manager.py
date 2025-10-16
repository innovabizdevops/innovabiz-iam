"""
INNOVABIZ - Compliance Certificate Manager
===============================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0

Gerenciador de certificados de compliance para múltiplos frameworks regulatórios,
suportando GDPR, LGPD, HIPAA e PNDSB com validação específica por região.
===============================================================
"""

import json
import uuid
import logging
import datetime
from typing import Dict, List, Optional, Any, Tuple
from enum import Enum
import hashlib
import base64
import hmac
import os

# Importações internas
from ..compliance_metadata import ComplianceFramework, Region, Industry
from ..compliance_engine import ValidationReport
from ...common.observability import ContextLogger
from ...common.security import EncryptionService, SignatureService
from ...common.storage import ConfigRepository

# Configurar logging
logger = logging.getLogger(__name__)


class CertificateStatus(str, Enum):
    """Status do certificado de compliance."""
    ACTIVE = "ACTIVE"           # Certificado válido e ativo
    EXPIRED = "EXPIRED"         # Certificado expirado
    REVOKED = "REVOKED"         # Certificado revogado (por irregularidade, problema crítico)
    SUPERSEDED = "SUPERSEDED"   # Certificado substituído por nova versão


class ComplianceCertificate:
    """
    Representação de um certificado de compliance que atesta a conformidade
    com um ou mais frameworks regulatórios.
    """
    
    def __init__(
        self,
        tenant_id: str,
        frameworks: List[ComplianceFramework],
        region: Region,
        industry: Industry,
        validation_report_id: str,
        issuer: str = "INNOVABIZ Compliance Authority",
        id: str = None,
        issue_date: datetime.datetime = None,
        expiry_date: datetime.datetime = None,
        status: CertificateStatus = CertificateStatus.ACTIVE,
        metadata: Dict[str, Any] = None,
        signature: str = None
    ):
        self.id = id or str(uuid.uuid4())
        self.tenant_id = tenant_id
        self.frameworks = frameworks
        self.region = region
        self.industry = industry
        self.validation_report_id = validation_report_id
        self.issuer = issuer
        self.issue_date = issue_date or datetime.datetime.utcnow()
        self.expiry_date = expiry_date or (self.issue_date + datetime.timedelta(days=365))
        self.status = status
        self.metadata = metadata or {}
        self.signature = signature
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o certificado para dicionário."""
        return {
            "id": self.id,
            "tenant_id": self.tenant_id,
            "frameworks": [fw for fw in self.frameworks],
            "region": self.region,
            "industry": self.industry,
            "validation_report_id": self.validation_report_id,
            "issuer": self.issuer,
            "issue_date": self.issue_date.isoformat(),
            "expiry_date": self.expiry_date.isoformat(),
            "status": self.status,
            "metadata": self.metadata,
            "signature": self.signature
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ComplianceCertificate':
        """Cria um certificado a partir de um dicionário."""
        return cls(
            id=data.get("id"),
            tenant_id=data.get("tenant_id"),
            frameworks=data.get("frameworks", []),
            region=data.get("region"),
            industry=data.get("industry"),
            validation_report_id=data.get("validation_report_id"),
            issuer=data.get("issuer", "INNOVABIZ Compliance Authority"),
            issue_date=datetime.datetime.fromisoformat(data.get("issue_date")),
            expiry_date=datetime.datetime.fromisoformat(data.get("expiry_date")),
            status=data.get("status", CertificateStatus.ACTIVE),
            metadata=data.get("metadata", {}),
            signature=data.get("signature")
        )
    
    def is_valid(self) -> bool:
        """Verifica se o certificado está válido (não expirado, não revogado)."""
        now = datetime.datetime.utcnow()
        return (
            self.status == CertificateStatus.ACTIVE and
            now < self.expiry_date
        )
    
    def days_until_expiry(self) -> int:
        """Retorna o número de dias até a expiração do certificado."""
        now = datetime.datetime.utcnow()
        if now > self.expiry_date:
            return 0
        delta = self.expiry_date - now
        return delta.days
    
    def get_framework_names(self) -> List[str]:
        """Retorna os nomes dos frameworks de compliance."""
        framework_names = {
            "GDPR": "General Data Protection Regulation",
            "LGPD": "Lei Geral de Proteção de Dados",
            "HIPAA": "Health Insurance Portability and Accountability Act",
            "PNDSB": "Política Nacional de Desenvolvimento do Setor Bancário"
        }
        
        return [framework_names.get(fw, fw) for fw in self.frameworks]
    
    def get_region_name(self) -> str:
        """Retorna o nome da região."""
        region_names = {
            "EU_CENTRAL": "União Europeia/Portugal",
            "BR": "Brasil",
            "US_EAST": "Estados Unidos",
            "AF_ANGOLA": "Angola",
            "GLOBAL": "Global"
        }
        
        return region_names.get(self.region, self.region)
    
    def get_industry_name(self) -> str:
        """Retorna o nome da indústria."""
        industry_names = {
            "FINANCIAL": "Serviços Financeiros",
            "HEALTHCARE": "Saúde",
            "RETAIL": "Varejo",
            "TELECOM": "Telecomunicações",
            "GOVERNMENT": "Governo",
            "INSURANCE": "Seguros"
        }
        
        return industry_names.get(self.industry, self.industry)


class ComplianceCertificateManager:
    """
    Gerenciador de certificados de compliance, responsável por gerar,
    validar, revogar e gerenciar certificados de conformidade regulatória.
    """
    
    def __init__(
        self, 
        config_repository: ConfigRepository,
        encryption_service: EncryptionService = None,
        signature_service: SignatureService = None
    ):
        """
        Inicializa o gerenciador de certificados.
        
        Args:
            config_repository: Repositório para armazenamento de certificados
            encryption_service: Serviço de criptografia para dados sensíveis
            signature_service: Serviço de assinatura para validação de integridade
        """
        self.config_repository = config_repository
        self.encryption_service = encryption_service
        self.signature_service = signature_service
        self.logger = ContextLogger("certificate_manager")
        
        # Configurar o caminho no repositório
        self.certificate_config_path = "iam/compliance/certificates"
        
        # Inicializar cache de certificados
        self._certificates_cache = {}
    
    def generate_certificate(
        self,
        tenant_id: str,
        validation_report: ValidationReport,
        valid_days: int = 365,
        metadata: Dict[str, Any] = None
    ) -> ComplianceCertificate:
        """
        Gera um novo certificado de compliance com base em um relatório de validação.
        
        Args:
            tenant_id: ID do tenant
            validation_report: Relatório de validação que atesta conformidade
            valid_days: Número de dias de validade do certificado
            metadata: Metadados adicionais para o certificado
            
        Returns:
            Certificado de compliance gerado
        """
        self.logger.info(f"Gerando certificado para tenant {tenant_id}", 
                        extra={"tenant_id": tenant_id})
        
        # Verificar se relatório indica conformidade
        if validation_report.overall_status != "PASS":
            self.logger.warning(
                f"Tentativa de gerar certificado para relatório não conforme: {validation_report.id}",
                extra={"report_id": validation_report.id, "status": validation_report.overall_status}
            )
            raise ValueError(f"Não é possível gerar certificado para relatório não conforme: {validation_report.overall_status}")
        
        # Preparar metadados para o certificado
        meta = metadata or {}
        meta.update({
            "compliance_score": validation_report.score,
            "validation_timestamp": validation_report.timestamp,
            "validation_scope": validation_report.scope,
            "criticality": self._determine_criticality(validation_report)
        })
        
        # Criar o certificado
        now = datetime.datetime.utcnow()
        certificate = ComplianceCertificate(
            tenant_id=tenant_id,
            frameworks=validation_report.frameworks,
            region=validation_report.region,
            industry=validation_report.industry,
            validation_report_id=validation_report.id,
            issue_date=now,
            expiry_date=now + datetime.timedelta(days=valid_days),
            metadata=meta
        )
        
        # Assinar o certificado para garantir integridade
        if self.signature_service:
            certificate = self._sign_certificate(certificate)
        
        # Persistir o certificado
        self._save_certificate(certificate)
        
        # Atualizar cache
        self._certificates_cache[certificate.id] = certificate
        
        self.logger.info(f"Certificado gerado com sucesso: {certificate.id}", 
                        extra={"certificate_id": certificate.id})
        
        return certificate
    
    def get_certificate(self, certificate_id: str) -> Optional[ComplianceCertificate]:
        """
        Obtém um certificado pelo ID.
        
        Args:
            certificate_id: ID do certificado
            
        Returns:
            Certificado encontrado ou None
        """
        # Verificar cache
        if certificate_id in self._certificates_cache:
            return self._certificates_cache[certificate_id]
        
        # Buscar no repositório
        config_key = f"{self.certificate_config_path}/{certificate_id}"
        certificate_data = self.config_repository.get(config_key)
        
        if not certificate_data:
            self.logger.warning(f"Certificado não encontrado: {certificate_id}")
            return None
        
        # Criar objeto de certificado
        certificate = ComplianceCertificate.from_dict(certificate_data)
        
        # Verificar assinatura se existir
        if certificate.signature and self.signature_service:
            if not self._verify_certificate_signature(certificate):
                self.logger.warning(f"Assinatura inválida para certificado: {certificate_id}",
                                 extra={"certificate_id": certificate_id})
                return None
        
        # Atualizar cache
        self._certificates_cache[certificate_id] = certificate
        
        return certificate
    
    def get_certificates_for_tenant(
        self, 
        tenant_id: str,
        framework: Optional[ComplianceFramework] = None,
        region: Optional[Region] = None,
        active_only: bool = True
    ) -> List[ComplianceCertificate]:
        """
        Obtém todos os certificados para um tenant, com filtragem opcional.
        
        Args:
            tenant_id: ID do tenant
            framework: Filtrar por framework específico
            region: Filtrar por região específica
            active_only: Retornar apenas certificados ativos
            
        Returns:
            Lista de certificados encontrados
        """
        # Buscar todos os certificados do tenant
        config_prefix = f"{self.certificate_config_path}/tenant/{tenant_id}/"
        certificates_data = self.config_repository.get_all(config_prefix)
        
        certificates = []
        for cert_data in certificates_data:
            try:
                certificate = ComplianceCertificate.from_dict(cert_data)
                
                # Aplicar filtros
                if framework and framework not in certificate.frameworks:
                    continue
                
                if region and certificate.region != region:
                    continue
                
                if active_only and not certificate.is_valid():
                    continue
                
                # Verificar assinatura se existir
                if certificate.signature and self.signature_service:
                    if not self._verify_certificate_signature(certificate):
                        self.logger.warning(f"Assinatura inválida para certificado: {certificate.id}",
                                         extra={"certificate_id": certificate.id})
                        continue
                
                certificates.append(certificate)
                
                # Atualizar cache
                self._certificates_cache[certificate.id] = certificate
                
            except Exception as e:
                self.logger.error(f"Erro ao processar certificado: {e}")
                continue
        
        return certificates
    
    def revoke_certificate(
        self, 
        certificate_id: str,
        reason: str
    ) -> bool:
        """
        Revoga um certificado de compliance.
        
        Args:
            certificate_id: ID do certificado a ser revogado
            reason: Motivo da revogação
            
        Returns:
            True se revogado com sucesso, False caso contrário
        """
        certificate = self.get_certificate(certificate_id)
        if not certificate:
            self.logger.warning(f"Certificado não encontrado para revogação: {certificate_id}")
            return False
        
        self.logger.info(f"Revogando certificado: {certificate_id}", 
                       extra={"certificate_id": certificate_id, "reason": reason})
        
        # Atualizar status
        certificate.status = CertificateStatus.REVOKED
        certificate.metadata["revocation"] = {
            "timestamp": datetime.datetime.utcnow().isoformat(),
            "reason": reason
        }
        
        # Assinar novamente após alterações
        if self.signature_service:
            certificate = self._sign_certificate(certificate)
        
        # Persistir alterações
        self._save_certificate(certificate)
        
        # Atualizar cache
        self._certificates_cache[certificate.id] = certificate
        
        self.logger.info(f"Certificado revogado com sucesso: {certificate_id}")
        
        return True
    
    def renew_certificate(
        self,
        certificate_id: str,
        validation_report: ValidationReport,
        valid_days: int = 365
    ) -> Optional[ComplianceCertificate]:
        """
        Renova um certificado existente com base em um novo relatório de validação.
        
        Args:
            certificate_id: ID do certificado a ser renovado
            validation_report: Novo relatório de validação
            valid_days: Número de dias de validade do novo certificado
            
        Returns:
            Novo certificado gerado, ou None se falhar
        """
        old_certificate = self.get_certificate(certificate_id)
        if not old_certificate:
            self.logger.warning(f"Certificado não encontrado para renovação: {certificate_id}")
            return None
        
        # Verificar se relatório indica conformidade
        if validation_report.overall_status != "PASS":
            self.logger.warning(
                f"Não é possível renovar com relatório não conforme: {validation_report.id}",
                extra={"report_id": validation_report.id, "status": validation_report.overall_status}
            )
            return None
        
        self.logger.info(f"Renovando certificado: {certificate_id}", 
                       extra={"certificate_id": certificate_id})
        
        # Atualizar status do certificado antigo
        old_certificate.status = CertificateStatus.SUPERSEDED
        old_certificate.metadata["superseded_by"] = {
            "timestamp": datetime.datetime.utcnow().isoformat(),
            "new_certificate_id": None  # Será preenchido depois
        }
        
        # Persistir alterações no certificado antigo
        if self.signature_service:
            old_certificate = self._sign_certificate(old_certificate)
        
        # Criar novo certificado
        now = datetime.datetime.utcnow()
        new_certificate = ComplianceCertificate(
            tenant_id=old_certificate.tenant_id,
            frameworks=validation_report.frameworks,
            region=validation_report.region,
            industry=validation_report.industry,
            validation_report_id=validation_report.id,
            issue_date=now,
            expiry_date=now + datetime.timedelta(days=valid_days),
            metadata={
                "compliance_score": validation_report.score,
                "validation_timestamp": validation_report.timestamp,
                "validation_scope": validation_report.scope,
                "criticality": self._determine_criticality(validation_report),
                "previous_certificate_id": old_certificate.id,
                "renewal_date": now.isoformat()
            }
        )
        
        # Assinar o novo certificado
        if self.signature_service:
            new_certificate = self._sign_certificate(new_certificate)
        
        # Atualizar metadados do certificado antigo com ID do novo
        old_certificate.metadata["superseded_by"]["new_certificate_id"] = new_certificate.id
        
        # Persistir ambos certificados
        self._save_certificate(old_certificate)
        self._save_certificate(new_certificate)
        
        # Atualizar cache
        self._certificates_cache[old_certificate.id] = old_certificate
        self._certificates_cache[new_certificate.id] = new_certificate
        
        self.logger.info(f"Certificado renovado com sucesso: {new_certificate.id}", 
                       extra={"old_id": certificate_id, "new_id": new_certificate.id})
        
        return new_certificate
    
    def check_expiring_certificates(
        self,
        tenant_id: Optional[str] = None,
        days_threshold: int = 30
    ) -> List[ComplianceCertificate]:
        """
        Verifica certificados próximos da expiração.
        
        Args:
            tenant_id: ID do tenant (opcional, para verificar todos)
            days_threshold: Limiar de dias para considerar como próximo da expiração
            
        Returns:
            Lista de certificados próximos da expiração
        """
        expiring_certificates = []
        
        # Determinar certificados a verificar
        if tenant_id:
            certificates = self.get_certificates_for_tenant(tenant_id, active_only=True)
        else:
            # Buscar todos os certificados ativos
            config_prefix = f"{self.certificate_config_path}/"
            certificates_data = self.config_repository.get_all(config_prefix)
            
            certificates = []
            for cert_data in certificates_data:
                try:
                    certificate = ComplianceCertificate.from_dict(cert_data)
                    if certificate.is_valid():
                        certificates.append(certificate)
                except Exception as e:
                    self.logger.error(f"Erro ao processar certificado: {e}")
                    continue
        
        # Verificar expiração
        now = datetime.datetime.utcnow()
        for certificate in certificates:
            days_remaining = certificate.days_until_expiry()
            if days_remaining <= days_threshold:
                expiring_certificates.append(certificate)
        
        return expiring_certificates
    
    def export_certificate(
        self, 
        certificate_id: str,
        format: str = "json"
    ) -> Dict[str, Any]:
        """
        Exporta um certificado em um formato específico.
        
        Args:
            certificate_id: ID do certificado a ser exportado
            format: Formato de exportação ("json", "pdf")
            
        Returns:
            Dicionário com dados da exportação
        """
        certificate = self.get_certificate(certificate_id)
        if not certificate:
            raise ValueError(f"Certificado não encontrado: {certificate_id}")
        
        if format == "json":
            return certificate.to_dict()
        elif format == "pdf":
            # Implementação de exportação para PDF seria feita aqui
            raise NotImplementedError("Exportação para PDF não implementada")
        else:
            raise ValueError(f"Formato de exportação não suportado: {format}")
    
    def _sign_certificate(self, certificate: ComplianceCertificate) -> ComplianceCertificate:
        """
        Assina um certificado para garantir sua integridade.
        
        Args:
            certificate: Certificado a ser assinado
            
        Returns:
            Certificado com assinatura
        """
        if not self.signature_service:
            return certificate
        
        # Preparar conteúdo para assinatura
        content = self._get_certificate_content_for_signature(certificate)
        
        # Gerar assinatura
        signature = self.signature_service.sign(content)
        
        # Atualizar certificado
        certificate.signature = signature
        
        return certificate
    
    def _verify_certificate_signature(self, certificate: ComplianceCertificate) -> bool:
        """
        Verifica a assinatura de um certificado.
        
        Args:
            certificate: Certificado a verificar
            
        Returns:
            True se assinatura válida, False caso contrário
        """
        if not self.signature_service or not certificate.signature:
            return False
        
        # Preparar conteúdo para verificação
        content = self._get_certificate_content_for_signature(certificate)
        
        # Verificar assinatura
        return self.signature_service.verify(content, certificate.signature)
    
    def _get_certificate_content_for_signature(self, certificate: ComplianceCertificate) -> str:
        """
        Obtém o conteúdo do certificado para assinatura/verificação.
        
        Args:
            certificate: Certificado
            
        Returns:
            String representando o conteúdo canônico do certificado
        """
        # Criar uma cópia do certificado sem a assinatura
        cert_dict = certificate.to_dict()
        cert_dict.pop("signature", None)
        
        # Serializar em JSON ordenado
        return json.dumps(cert_dict, sort_keys=True)
    
    def _save_certificate(self, certificate: ComplianceCertificate) -> bool:
        """
        Salva um certificado no repositório.
        
        Args:
            certificate: Certificado a ser salvo
            
        Returns:
            True se salvo com sucesso, False caso contrário
        """
        try:
            # Salvar no repositório principal
            config_key = f"{self.certificate_config_path}/{certificate.id}"
            self.config_repository.upsert(config_key, certificate.to_dict())
            
            # Salvar na estrutura indexada por tenant
            tenant_key = f"{self.certificate_config_path}/tenant/{certificate.tenant_id}/{certificate.id}"
            self.config_repository.upsert(tenant_key, certificate.to_dict())
            
            return True
        except Exception as e:
            self.logger.error(f"Erro ao salvar certificado: {e}")
            return False
    
    def _determine_criticality(self, validation_report: ValidationReport) -> str:
        """
        Determina a criticidade do certificado com base no relatório de validação.
        
        Args:
            validation_report: Relatório de validação
            
        Returns:
            Criticidade ("LOW", "MEDIUM", "HIGH", "CRITICAL")
        """
        # Analisar indústria e região para determinar criticidade
        industry = validation_report.industry
        
        # Indústrias altamente reguladas são mais críticas
        if industry in ["HEALTHCARE", "FINANCIAL"]:
            return "HIGH"
        elif industry in ["GOVERNMENT", "TELECOM"]:
            return "MEDIUM"
        
        # Verificar resultados do relatório
        high_severity_count = 0
        for result in validation_report.results:
            if result.get("severity") == "HIGH":
                high_severity_count += 1
        
        if high_severity_count > 5:
            return "CRITICAL"
        elif high_severity_count > 2:
            return "HIGH"
        
        return "MEDIUM"
