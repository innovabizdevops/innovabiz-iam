"""
INNOVABIZ - Gerador de Certificados de Conformidade IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Sistema para geração de certificados de conformidade
           para o módulo IAM, usando criptografia para garantir
           a autenticidade e verificabilidade dos certificados.
==================================================================
"""

import os
import json
import uuid
import base64
import hashlib
import jinja2
import datetime
import dataclasses
from pathlib import Path
from typing import Dict, List, Any, Optional, Set, Tuple, Union

import cryptography
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import padding, rsa
from cryptography.hazmat.primitives.serialization import (
    Encoding, PrivateFormat, PublicFormat, NoEncryption, load_pem_private_key, load_pem_public_key
)

from .iam_validator import ValidationReport, ValidationStatus
from ..extensions.healthcare.integration import HealthcareComplianceService

# Configuração para template Jinja2
TEMPLATE_DIR = Path(__file__).parent / "templates"
TEMPLATE_ENV = jinja2.Environment(
    loader=jinja2.FileSystemLoader(TEMPLATE_DIR),
    autoescape=jinja2.select_autoescape(['html', 'xml'])
)

class CertificateGenerator:
    """
    Gerador de certificados de conformidade criptograficamente verificáveis.
    """
    
    def __init__(self, keys_dir: Optional[Path] = None):
        """
        Inicializa o gerador de certificados.
        
        Args:
            keys_dir: Diretório para armazenar chaves criptográficas
        """
        self.keys_dir = keys_dir if keys_dir else Path(__file__).parent / "keys"
        self.keys_dir.mkdir(exist_ok=True, parents=True)
        
        self._ensure_key_pair()
    
    def _ensure_key_pair(self):
        """Garante que exista um par de chaves para assinatura dos certificados"""
        private_key_path = self.keys_dir / "iam_certificate_private.pem"
        public_key_path = self.keys_dir / "iam_certificate_public.pem"
        
        # Verificar se as chaves já existem
        if not private_key_path.exists() or not public_key_path.exists():
            # Gerar novo par de chaves
            private_key = rsa.generate_private_key(
                public_exponent=65537,
                key_size=2048
            )
            public_key = private_key.public_key()
            
            # Salvar chave privada
            private_key_pem = private_key.private_bytes(
                encoding=Encoding.PEM,
                format=PrivateFormat.PKCS8,
                encryption_algorithm=NoEncryption()
            )
            with open(private_key_path, "wb") as f:
                f.write(private_key_pem)
            
            # Salvar chave pública
            public_key_pem = public_key.public_bytes(
                encoding=Encoding.PEM,
                format=PublicFormat.SubjectPublicKeyInfo
            )
            with open(public_key_path, "wb") as f:
                f.write(public_key_pem)
    
    def _load_private_key(self):
        """Carrega a chave privada para assinatura"""
        private_key_path = self.keys_dir / "iam_certificate_private.pem"
        with open(private_key_path, "rb") as f:
            private_key_data = f.read()
        
        return load_pem_private_key(
            private_key_data,
            password=None
        )
    
    def _load_public_key(self):
        """Carrega a chave pública para verificação"""
        public_key_path = self.keys_dir / "iam_certificate_public.pem"
        with open(public_key_path, "rb") as f:
            public_key_data = f.read()
        
        return load_pem_public_key(public_key_data)
    
    def _sign_data(self, data: bytes) -> bytes:
        """
        Assina os dados com a chave privada
        
        Args:
            data: Dados a serem assinados
            
        Returns:
            Assinatura digital
        """
        private_key = self._load_private_key()
        
        signature = private_key.sign(
            data,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        
        return signature
    
    def _verify_signature(self, data: bytes, signature: bytes) -> bool:
        """
        Verifica a assinatura digital
        
        Args:
            data: Dados originais
            signature: Assinatura a ser verificada
            
        Returns:
            True se a assinatura for válida, False caso contrário
        """
        public_key = self._load_public_key()
        
        try:
            public_key.verify(
                signature,
                data,
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            return True
        except Exception:
            return False
    
    def generate_certificate(self, 
                           tenant_id: str, 
                           validation_reports: Dict[str, ValidationReport],
                           language: str = "pt",
                           certificate_options: Optional[Dict[str, Any]] = None,
                           metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Gera um certificado de conformidade com base nos relatórios de validação
        
        Args:
            tenant_id: ID do tenant
            validation_reports: Relatórios de validação por framework
            language: Idioma do certificado (pt/en)
            certificate_options: Opções de configuração do certificado
            metadata: Metadados adicionais do certificado
            
        Returns:
            Dados do certificado, incluindo caminho do arquivo gerado
        """
        # Opções do certificado (com valores padrão)
        options = {
            "includeDetails": True,
            "includeAttachments": True,
            "includeLogo": True,
            "includeDashboard": True,
            "validityPeriod": 365  # dias
        }
        
        if certificate_options:
            options.update(certificate_options)
            
        # Metadados opcionais do certificado
        cert_metadata = {
            "organizationName": "",
            "department": "",
            "contact": "",
            "comments": ""
        }
        
        if metadata:
            cert_metadata.update(metadata)
        
        # Preparar os dados do certificado
        certificate_id = f"CERT-{uuid.uuid4()}"
        timestamp = datetime.datetime.now()
        expiration_date = timestamp + datetime.timedelta(days=options["validityPeriod"])
        
        # Calcular estatísticas de validação
        total_validations = 0
        total_passed = 0
        framework_names = []
        
        for framework, report in validation_reports.items():
            framework_names.append(SUPPORTED_FRAMEWORKS.get(framework, framework))
            total_validations += len(report.results)
            total_passed += report.passed_count
        
        # Calcular pontuação de conformidade
        compliance_score = (total_passed / total_validations * 100) if total_validations > 0 else 0
        
        # Determinar status geral
        overall_status = ValidationStatus.PASSED.value
        for report in validation_reports.values():
            if report.overall_status == ValidationStatus.FAILED:
                overall_status = ValidationStatus.FAILED.value
                break
        
        # Criar dados do certificado
        certificate_data = {
            "certificate_id": certificate_id,
            "tenant_id": tenant_id,
            "timestamp": timestamp.isoformat(),
            "expiration_date": expiration_date.isoformat(),
            "frameworks": framework_names,
            "validation_summary": {
                "total_validations": total_validations,
                "passed_validations": total_passed,
                "compliance_score": round(compliance_score, 2),
                "overall_status": overall_status
            },
            "options": options,
            "metadata": cert_metadata,
            "verification_hash": "",  # Será preenchido após serialização
            "digital_signature": ""  # Será preenchida após assinatura
        }
        
        # Hash dos dados para verificação
        data_json = json.dumps(certificate_data, sort_keys=True)
        data_hash = hashlib.sha256(data_json.encode()).hexdigest()
        
        # Assinar dados
        signature = self._sign_data(data_json.encode())
        signature_b64 = base64.b64encode(signature).decode('utf-8')
        
        # Adicionar hash e assinatura ao certificado
        certificate_data["verification_hash"] = data_hash
        certificate_data["digital_signature"] = signature_b64
        
        # Gerar HTML do certificado
        html_certificate = self._generate_html_certificate(
            certificate_data, language
        )
        
        # Salvar certificado
        reports_dir = Path(__file__).parent.parent / "reports"
        cert_dir = reports_dir / "certifications"
        cert_dir.mkdir(exist_ok=True, parents=True)
        
        certificate_path = cert_dir / f"{certificate_id}_{language}.html"
        with open(certificate_path, "w", encoding="utf-8") as f:
            f.write(html_certificate)
        
        # Adicionar caminho ao resultado
        certificate_data["file_path"] = str(certificate_path)
        
        return certificate_data
    
    def _generate_html_certificate(self, certificate_data: Dict[str, Any], language: str) -> str:
        """
        Gera o HTML do certificado usando o template
        
        Args:
            certificate_data: Dados do certificado
            language: Idioma (pt/en)
            
        Returns:
            HTML do certificado
        """
        # Carregar traduções
        translations_dir = Path(__file__).parent / "translations"
        translations_path = translations_dir / f"{language}.json"
        
        with open(translations_path, "r", encoding="utf-8") as f:
            translations = json.load(f)
        
        # Carregar template
        template = TEMPLATE_ENV.get_template("compliance_certificate_template.html")
        
        # Preparar dados para o template
        issued_date = datetime.datetime.fromisoformat(certificate_data["issued_date"])
        valid_until = datetime.datetime.fromisoformat(certificate_data["valid_until"])
        
        # INNOVABIZ logo placeholder
        # Em uma implementação real, você carregaria o logo da empresa
        logo_base64 = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxMDAiIGhlaWdodD0iMTAwIiB2aWV3Qm94PSIwIDAgMTAwIDEwMCI+PHJlY3Qgd2lkdGg9IjEwMCIgaGVpZ2h0PSIxMDAiIGZpbGw9IiMwMDU2YjMiLz48dGV4dCB4PSI1MCIgeT0iNTAiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNiIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGFsaWdubWVudC1iYXNlbGluZT0ibWlkZGxlIj5JTk5PVkFCSVo8L3RleHQ+PC9zdmc+"
        
        data_for_template = {
            # Cabeçalho
            "logo_base64": logo_base64,
            "certificate_title": translations["validation"]["certificate"]["title"],
            "certificate_subtitle": translations["validation"]["certificate"]["subtitle"],
            
            # Dados do certificado
            "tenant_label": translations["validation"]["certificate"]["tenant"],
            "tenant_id": certificate_data["tenant_id"],
            "certificate_id_label": "ID",
            "certificate_id": certificate_data["certificate_id"],
            "issued_date_label": translations["validation"]["certificate"]["issued_date"],
            "issued_date": issued_date.strftime("%d/%m/%Y %H:%M:%S"),
            "valid_until_label": translations["validation"]["certificate"]["valid_until"],
            "valid_until": valid_until.strftime("%d/%m/%Y %H:%M:%S"),
            
            # Frameworks
            "frameworks_label": translations["validation"]["certificate"]["frameworks"],
            "frameworks": certificate_data["frameworks"],
            
            # Resumo da validação
            "validation_summary_label": translations["validation"]["report"]["summary"],
            "validation_summary": translations["validation"]["report"]["title"],
            "total_validations_label": translations["validation"]["report"]["total_validations"],
            "total_validations": certificate_data["validation_summary"]["total_validations"],
            "passed_validations_label": translations["validation"]["report"]["passed"],
            "passed_validations": certificate_data["validation_summary"]["passed_validations"],
            "compliance_score_label": "Compliance Score",
            "compliance_score": certificate_data["validation_summary"]["compliance_score"],
            "overall_status_label": translations["validation"]["report"]["overall_status"],
            "overall_status": certificate_data["validation_summary"]["overall_status"],
            
            # Verificação
            "verification_label": translations["validation"]["certificate"]["verification"],
            "verification_instructions": translations["validation"]["certificate"]["verification_instructions"],
            "verification_hash": certificate_data["verification_hash"],
            
            # Assinatura
            "signature_label": translations["validation"]["certificate"]["signature"],
            "digital_signature": certificate_data["digital_signature"],
            
            # Rodapé
            "footer_text": translations["validation"]["report"]["footer"] + "1.0.0",
            "current_year": datetime.datetime.now().year,
            "all_rights_reserved": "All rights reserved."
        }
        
        # Renderizar template
        return template.render(**data_for_template)
    
    def generate_healthcare_compliance_certificate(self, 
                                               hipaa_validation_id: str,
                                               tenant_id: str,
                                               regions: List[str],
                                               language: str = "pt",
                                               certificate_options: Optional[Dict[str, Any]] = None,
                                               metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Gera um certificado específico para conformidade HIPAA com integração ao módulo Healthcare
        
        Args:
            hipaa_validation_id: ID da validação HIPAA
            tenant_id: ID do tenant
            regions: Lista de regiões validadas
            language: Idioma do certificado (pt/en)
            certificate_options: Opções de configuração do certificado
            metadata: Metadados adicionais do certificado
            
        Returns:
            Dados do certificado gerado incluindo caminhos para os arquivos
        """
        # Integrar com o serviço de conformidade Healthcare
        healthcare_service = HealthcareComplianceService()
        
        # Obter os dados do relatório de conformidade HIPAA
        healthcare_compliance_data = healthcare_service.get_compliance_data(hipaa_validation_id)
        
        # Criar um ValidationReport a partir dos dados de Healthcare
        healthcare_validation_report = ValidationReport(
            framework="hipaa",
            status=ValidationStatus.COMPLIANT if healthcare_compliance_data["overallScore"] >= 80 else ValidationStatus.NON_COMPLIANT,
            timestamp=datetime.datetime.now(),
            score=healthcare_compliance_data["overallScore"],
            details={
                "findings": healthcare_compliance_data["findings"],
                "categoryScores": healthcare_compliance_data["categoryScores"],
                "policies": healthcare_compliance_data["policies"],
                "healthcareSpecific": True,
                "healthcareServices": healthcare_compliance_data.get("services", []),
                "hipaaCategories": [
                    "Privacy Rule",
                    "Security Rule",
                    "Breach Notification Rule",
                    "Patient Rights"
                ]
            }
        )
        
        # Preparar metadados específicos para Healthcare
        if not metadata:
            metadata = {}
        
        metadata.update({
            "healthcareIntegration": True,
            "healthcareModule": "INNOVABIZ Healthcare",
            "hipaaVersion": "2023 Omnibus",
            "regions": regions
        })
        
        # Gerar o certificado usando o método padrão
        validation_reports = {"hipaa": healthcare_validation_report}
        return self.generate_certificate(
            tenant_id=tenant_id,
            validation_reports=validation_reports,
            language=language,
            certificate_options=certificate_options,
            metadata=metadata
        )
    
    def verify_certificate(self, certificate_path: Union[str, Path]) -> Dict[str, Any]:
        """
        Verifica a autenticidade de um certificado
        
        Args:
            certificate_path: Caminho para o arquivo do certificado
            
        Returns:
            Resultado da verificação
        """
        if isinstance(certificate_path, str):
            certificate_path = Path(certificate_path)
        
        # Verificar se o arquivo existe
        if not certificate_path.exists():
            return {
                "verified": False,
                "error": "Certificate file not found"
            }
        
        # Extrair dados do certificado
        with open(certificate_path, "r", encoding="utf-8") as f:
            certificate_content = f.read()
        
        # Extrair hash e assinatura
        import re
        hash_match = re.search(r'<div class="verification-code">([a-f0-9]+)</div>', certificate_content)
        signature_match = re.search(r'<div class="signature-value">([A-Za-z0-9+/=]+)</div>', certificate_content)
        
        if not hash_match or not signature_match:
            return {
                "verified": False,
                "error": "Certificate format is invalid"
            }
        
        verification_hash = hash_match.group(1)
        signature_b64 = signature_match.group(1)
        
        # Extrair dados do certificado
        # Nota: Isso é simplesmente uma simulação de extração.
        # Em uma implementação real, você analisaria o HTML corretamente.
        certificate_id_match = re.search(r'<span class="value">(CERT-[a-f0-9-]+)</span>', certificate_content)
        tenant_id_match = re.search(r'<span class="label">Tenant</span>\s*<span class="value">([^<]+)</span>', certificate_content)
        
        if not certificate_id_match or not tenant_id_match:
            return {
                "verified": False,
                "error": "Cannot extract certificate data"
            }
        
        certificate_id = certificate_id_match.group(1)
        tenant_id = tenant_id_match.group(1)
        
        # Resultado da verificação
        verification_result = {
            "certificate_id": certificate_id,
            "tenant_id": tenant_id,
            "verification_hash": verification_hash,
            "verified": False
        }
        
        # Em uma implementação real, você reconstruiria os dados originais
        # para verificar a assinatura. Isto é uma simulação.
        
        return verification_result


# Frameworks suportados para tradução em certificados
SUPPORTED_FRAMEWORKS = {
    "hipaa": "Health Insurance Portability and Accountability Act (HIPAA)",
    "gdpr": "General Data Protection Regulation (GDPR)",
    "lgpd": "Lei Geral de Proteção de Dados (LGPD)",
    "pci_dss": "Payment Card Industry Data Security Standard (PCI DSS)",
    "security": "INNOVABIZ Security Baseline",
    "ar_auth": "AR Authentication Standards",
    "iso27001": "ISO/IEC 27001 Information Security Management",
    "nist": "NIST Cybersecurity Framework"
}
