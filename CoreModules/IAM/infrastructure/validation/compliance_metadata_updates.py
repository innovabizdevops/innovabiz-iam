"""
INNOVABIZ - Atualizações de Metadata de Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação das atualizações de regulamentações e frameworks
           de compliance para o módulo IAM com suporte multi-regional.
==================================================================
"""

import json
import logging
from typing import Dict, List, Any, Optional
from enum import Enum, auto
from pathlib import Path
from dataclasses import dataclass, field

from .compliance_metadata import (
    ComplianceMetadataRegistry,
    Region,
    Industry,
    ComplianceFramework,
    ComplianceRequirement,
    AuthenticationFactor
)

# Configuração de logging
logger = logging.getLogger(__name__)

# Extensão das enumerações existentes
class ExtendedRegion(Enum):
    """Extensão das regiões suportadas com subdivisões regulatórias"""
    GLOBAL = "global"
    EU = "eu"
    EU_FINANCE = "eu_finance"
    EU_HEALTH = "eu_health"
    EU_GOVERNMENT = "eu_government"
    BRAZIL = "brazil"
    BRAZIL_FINANCE = "brazil_finance"
    BRAZIL_HEALTH = "brazil_health"
    USA = "usa"
    USA_FINANCE = "usa_finance"
    USA_HEALTH = "usa_health"
    USA_GOVERNMENT = "usa_government"
    AFRICA = "africa"
    ANGOLA = "angola"
    CONGO = "congo"

class ExtendedIndustry(Enum):
    """Extensão das indústrias suportadas com categorias mais específicas"""
    GENERAL = "general"
    FINANCIAL = "financial"
    BANKING = "banking"
    INSURANCE = "insurance"
    MICROFINANCE = "microfinance"
    HEALTHCARE = "healthcare"
    TELECOM = "telecom"
    GOVERNMENT = "government"
    RETAIL = "retail"
    EDUCATION = "education"
    ENERGY = "energy"
    MANUFACTURING = "manufacturing"
    AGRICULTURE = "agriculture"

class ExtendedComplianceFramework(Enum):
    """Frameworks de compliance atualizados para 2025"""
    # Frameworks globais
    ISO_27001_2022 = "iso_27001_2022"
    ISO_27701_2023 = "iso_27701_2023"
    ISO_24760_2023 = "iso_24760_2023"
    PCI_DSS_4_0 = "pci_dss_4_0"
    NIST_CSF_2_0 = "nist_csf_2_0"
    
    # União Europeia
    GDPR = "gdpr"
    EU_AI_ACT_2025 = "eu_ai_act_2025"
    EIDAS_2_0 = "eidas_2_0"
    NIS2 = "nis2"
    DORA = "dora"
    
    # Brasil
    LGPD = "lgpd"
    ANPD_RES_2 = "anpd_res_2"
    OPEN_FINANCE_BR = "open_finance_br"
    PNDSB = "pndsb"
    
    # EUA
    NIST_800_63_4 = "nist_800_63_4"
    CMMC_2_0 = "cmmc_2_0"
    FEDRAMP_REV_5 = "fedramp_rev_5"
    HIPAA_2023 = "hipaa_2023"
    HITRUST_11 = "hitrust_11"
    CCPA_CPRA = "ccpa_cpra"
    
    # África
    ANGOLA_DATA_PROTECTION = "angola_data_protection"
    ANGOLA_ELECTRONIC_COMM = "angola_electronic_comm"
    BNA_REGULATIONS = "bna_regulations"

class ExtendedAuthenticationFactor(Enum):
    """Fatores de autenticação atualizados para 2025"""
    # Fatores tradicionais
    PASSWORD = "password"
    TOTP = "totp"
    SMS = "sms"
    EMAIL = "email"
    PUSH = "push"
    
    # Biometria tradicional
    FINGERPRINT = "fingerprint"
    FACE_RECOGNITION = "face_recognition"
    VOICE_RECOGNITION = "voice_recognition"
    
    # Tokens e certificados
    HARDWARE_TOKEN = "hardware_token"
    DIGITAL_CERTIFICATE = "digital_certificate"
    
    # WebAuthn/FIDO2
    FIDO2_PASSKEY = "fido2_passkey"
    FIDO2_SECURITY_KEY = "fido2_security_key"
    
    # Passwordless
    MAGIC_LINK = "magic_link"
    QR_CODE_AUTH = "qr_code_auth"
    CROSS_DEVICE_AUTH = "cross_device_auth"
    
    # Biometria comportamental
    TYPING_PATTERN = "typing_pattern"
    MOUSE_MOVEMENT = "mouse_movement"
    APP_USAGE_PATTERN = "app_usage_pattern"
    TRANSACTION_BEHAVIOR = "transaction_behavior"
    
    # Fatores AR/VR
    AR_FACIAL_3D = "ar_facial_3d"
    AR_GESTURE_RECOGNITION = "ar_gesture_recognition"
    AR_EYE_TRACKING = "ar_eye_tracking"
    AR_ENVIRONMENT = "ar_environment"
    
    # Identidade descentralizada
    DID_WALLET = "did_wallet"
    VERIFIABLE_CREDENTIAL = "verifiable_credential"
    BLOCKCHAIN_AUTH = "blockchain_auth"

def update_compliance_registry(registry: ComplianceMetadataRegistry) -> ComplianceMetadataRegistry:
    """
    Atualiza o registro de metadata de compliance com os novos frameworks e requisitos.
    
    Args:
        registry: Registro de metadata de compliance existente
        
    Returns:
        Registro de metadata de compliance atualizado
    """
    # Converter enumerações estendidas para as enumerações base do sistema
    # Esta função mantém a compatibilidade com o sistema existente enquanto
    # adiciona novos valores
    
    # Atualizar frameworks de compliance
    update_eu_frameworks(registry)
    update_brazil_frameworks(registry)
    update_usa_frameworks(registry)
    update_africa_frameworks(registry)
    update_global_frameworks(registry)
    
    # Atualizar benchmarks
    update_benchmarks(registry)
    
    return registry

def update_eu_frameworks(registry: ComplianceMetadataRegistry):
    """Atualiza frameworks da União Europeia"""
    
    # EU AI Act 2025
    registry.register_framework(
        ComplianceFramework.from_string("eu_ai_act_2025"),
        {
            "name": "EU AI Act 2025",
            "description": "Regulamento da União Europeia sobre Inteligência Artificial",
            "version": "1.0",
            "effective_date": "2025-01-01",
            "url": "https://digital-strategy.ec.europa.eu/en/policies/regulatory-framework-ai",
            "regions": [Region.EU],
            "industries": [Industry.GENERAL, Industry.FINANCIAL, Industry.HEALTHCARE],
            "supersedes": None
        }
    )
    
    # eIDAS 2.0
    registry.register_framework(
        ComplianceFramework.from_string("eidas_2_0"),
        {
            "name": "eIDAS 2.0",
            "description": "Electronic IDentification Authentication and trust Services 2.0",
            "version": "2.0",
            "effective_date": "2024-06-30",
            "url": "https://digital-strategy.ec.europa.eu/en/policies/eidas-regulation",
            "regions": [Region.EU],
            "industries": [Industry.GENERAL, Industry.FINANCIAL, Industry.GOVERNMENT],
            "supersedes": "eidas"
        }
    )
    
    # NIS2 Directive
    registry.register_framework(
        ComplianceFramework.from_string("nis2"),
        {
            "name": "NIS2 Directive",
            "description": "Network and Information Security Directive 2",
            "version": "2.0",
            "effective_date": "2024-10-17",
            "url": "https://digital-strategy.ec.europa.eu/en/policies/nis2-directive",
            "regions": [Region.EU],
            "industries": [Industry.FINANCIAL, Industry.HEALTHCARE, Industry.ENERGY],
            "supersedes": "nis"
        }
    )
    
    # DORA
    registry.register_framework(
        ComplianceFramework.from_string("dora"),
        {
            "name": "DORA",
            "description": "Digital Operational Resilience Act",
            "version": "1.0",
            "effective_date": "2025-01-17",
            "url": "https://www.eba.europa.eu/regulation-and-policy/digital-operational-resilience-dora",
            "regions": [Region.EU],
            "industries": [Industry.FINANCIAL, Industry.BANKING, Industry.INSURANCE],
            "supersedes": None
        }
    )

def update_brazil_frameworks(registry: ComplianceMetadataRegistry):
    """Atualiza frameworks do Brasil"""
    
    # ANPD Resolução CD/ANPD nº 2
    registry.register_framework(
        ComplianceFramework.from_string("anpd_res_2"),
        {
            "name": "ANPD Resolução CD/ANPD nº 2",
            "description": "Regulamento de aplicação da LGPD para Pequenas e Médias Empresas",
            "version": "1.0",
            "effective_date": "2023-09-05",
            "url": "https://www.gov.br/anpd/pt-br",
            "regions": [Region.BRAZIL],
            "industries": [Industry.GENERAL, Industry.FINANCIAL, Industry.HEALTHCARE],
            "supersedes": None
        }
    )
    
    # Open Finance BR
    registry.register_framework(
        ComplianceFramework.from_string("open_finance_br"),
        {
            "name": "Open Finance Brasil",
            "description": "Sistema Financeiro Aberto do Brasil",
            "version": "3.0",
            "effective_date": "2024-05-30",
            "url": "https://openfinancebrasil.org.br/",
            "regions": [Region.BRAZIL],
            "industries": [Industry.FINANCIAL, Industry.BANKING],
            "supersedes": "open_banking_br"
        }
    )
    
    # PNDSB
    registry.register_framework(
        ComplianceFramework.from_string("pndsb"),
        {
            "name": "PNDSB",
            "description": "Política Nacional de Segurança de Barragens",
            "version": "2024",
            "effective_date": "2024-01-01",
            "url": "https://www.gov.br/ana/pt-br",
            "regions": [Region.BRAZIL],
            "industries": [Industry.ENERGY, Industry.GENERAL],
            "supersedes": None
        }
    )

def update_usa_frameworks(registry: ComplianceMetadataRegistry):
    """Atualiza frameworks dos EUA"""
    
    # NIST SP 800-63-4
    registry.register_framework(
        ComplianceFramework.from_string("nist_800_63_4"),
        {
            "name": "NIST SP 800-63-4",
            "description": "Digital Identity Guidelines",
            "version": "4",
            "effective_date": "2024-07-01",
            "url": "https://pages.nist.gov/800-63-4/",
            "regions": [Region.USA],
            "industries": [Industry.GENERAL, Industry.FINANCIAL, Industry.GOVERNMENT],
            "supersedes": "nist_800_63_3"
        }
    )
    
    # CMMC 2.0
    registry.register_framework(
        ComplianceFramework.from_string("cmmc_2_0"),
        {
            "name": "CMMC 2.0",
            "description": "Cybersecurity Maturity Model Certification",
            "version": "2.0",
            "effective_date": "2023-10-01",
            "url": "https://www.acq.osd.mil/cmmc/",
            "regions": [Region.USA],
            "industries": [Industry.GOVERNMENT, Industry.MANUFACTURING],
            "supersedes": "cmmc_1_0"
        }
    )
    
    # FedRAMP Rev 5
    registry.register_framework(
        ComplianceFramework.from_string("fedramp_rev_5"),
        {
            "name": "FedRAMP Rev 5",
            "description": "Federal Risk and Authorization Management Program",
            "version": "5.0",
            "effective_date": "2024-04-01",
            "url": "https://www.fedramp.gov/",
            "regions": [Region.USA],
            "industries": [Industry.GOVERNMENT],
            "supersedes": "fedramp_rev_4"
        }
    )
    
    # HIPAA 2023
    registry.register_framework(
        ComplianceFramework.from_string("hipaa_2023"),
        {
            "name": "HIPAA 2023",
            "description": "Health Insurance Portability and Accountability Act",
            "version": "2023",
            "effective_date": "2023-07-01",
            "url": "https://www.hhs.gov/hipaa/index.html",
            "regions": [Region.USA],
            "industries": [Industry.HEALTHCARE],
            "supersedes": "hipaa"
        }
    )
    
    # HITRUST 11.0
    registry.register_framework(
        ComplianceFramework.from_string("hitrust_11"),
        {
            "name": "HITRUST 11.0",
            "description": "Health Information Trust Alliance",
            "version": "11.0",
            "effective_date": "2024-01-01",
            "url": "https://hitrustalliance.net/",
            "regions": [Region.USA],
            "industries": [Industry.HEALTHCARE],
            "supersedes": "hitrust_10"
        }
    )

def update_africa_frameworks(registry: ComplianceMetadataRegistry):
    """Atualiza frameworks da África"""
    
    # Lei de Proteção de Dados de Angola
    registry.register_framework(
        ComplianceFramework.from_string("angola_data_protection"),
        {
            "name": "Lei de Proteção de Dados de Angola",
            "description": "Lei n.º 22/11",
            "version": "1.0",
            "effective_date": "2011-06-17",
            "url": "https://www.governo.gov.ao/",
            "regions": [Region.AFRICA],
            "industries": [Industry.GENERAL],
            "supersedes": None
        }
    )
    
    # Lei das Comunicações Eletrônicas de Angola
    registry.register_framework(
        ComplianceFramework.from_string("angola_electronic_comm"),
        {
            "name": "Lei das Comunicações Eletrônicas",
            "description": "Lei das Comunicações Eletrônicas e dos Serviços da Sociedade de Informação",
            "version": "1.0",
            "effective_date": "2017-01-01",
            "url": "https://www.inacom.gov.ao/",
            "regions": [Region.AFRICA],
            "industries": [Industry.TELECOM, Industry.GENERAL],
            "supersedes": None
        }
    )
    
    # Regulamentos do BNA
    registry.register_framework(
        ComplianceFramework.from_string("bna_regulations"),
        {
            "name": "Regulamentos do BNA",
            "description": "Regulamentos do Banco Nacional de Angola para serviços financeiros",
            "version": "2023",
            "effective_date": "2023-01-01",
            "url": "https://www.bna.ao/",
            "regions": [Region.AFRICA],
            "industries": [Industry.FINANCIAL, Industry.BANKING],
            "supersedes": None
        }
    )

def update_global_frameworks(registry: ComplianceMetadataRegistry):
    """Atualiza frameworks globais"""
    
    # ISO 27001:2022
    registry.register_framework(
        ComplianceFramework.from_string("iso_27001_2022"),
        {
            "name": "ISO/IEC 27001:2022",
            "description": "Information Security Management",
            "version": "2022",
            "effective_date": "2022-10-25",
            "url": "https://www.iso.org/standard/27001",
            "regions": [Region.GLOBAL],
            "industries": [Industry.GENERAL],
            "supersedes": "iso_27001_2013"
        }
    )
    
    # ISO 27701:2023
    registry.register_framework(
        ComplianceFramework.from_string("iso_27701_2023"),
        {
            "name": "ISO/IEC 27701:2023",
            "description": "Privacy Information Management",
            "version": "2023",
            "effective_date": "2023-06-01",
            "url": "https://www.iso.org/standard/27701",
            "regions": [Region.GLOBAL],
            "industries": [Industry.GENERAL],
            "supersedes": "iso_27701_2019"
        }
    )
    
    # ISO 24760:2023
    registry.register_framework(
        ComplianceFramework.from_string("iso_24760_2023"),
        {
            "name": "ISO/IEC 24760-3:2023",
            "description": "IT Security and Privacy — A framework for identity management",
            "version": "2023",
            "effective_date": "2023-05-01",
            "url": "https://www.iso.org/standard/24760",
            "regions": [Region.GLOBAL],
            "industries": [Industry.GENERAL],
            "supersedes": "iso_24760_2019"
        }
    )
    
    # PCI DSS 4.0
    registry.register_framework(
        ComplianceFramework.from_string("pci_dss_4_0"),
        {
            "name": "PCI DSS 4.0",
            "description": "Payment Card Industry Data Security Standard",
            "version": "4.0",
            "effective_date": "2022-03-31",
            "url": "https://www.pcisecuritystandards.org/",
            "regions": [Region.GLOBAL],
            "industries": [Industry.FINANCIAL, Industry.RETAIL],
            "supersedes": "pci_dss_3_2_1"
        }
    )
    
    # NIST CSF 2.0
    registry.register_framework(
        ComplianceFramework.from_string("nist_csf_2_0"),
        {
            "name": "NIST CSF 2.0",
            "description": "Cybersecurity Framework",
            "version": "2.0",
            "effective_date": "2023-02-22",
            "url": "https://www.nist.gov/cyberframework",
            "regions": [Region.GLOBAL],
            "industries": [Industry.GENERAL],
            "supersedes": "nist_csf_1_1"
        }
    )

def update_benchmarks(registry: ComplianceMetadataRegistry):
    """Atualiza benchmarks baseados em Gartner, Forrester e outros"""
    
    benchmarks = {
        "gartner": {
            "year": 2025,
            "average_score": 72,
            "leader_score": 90,
            "innovator_score": 95,
            "leaders": ["Microsoft Entra", "Okta", "ForgeRock", "Ping Identity", "IBM Security Verify"],
            "challengers": ["Oracle", "Thales", "OneLogin", "Auth0", "CyberArk"],
            "visionaries": ["Auth0", "Transmit Security", "Beyond Identity", "STRIVACITY", "Cloudentity"],
            "key_capabilities": [
                "Passwordless Authentication",
                "Continuous Risk Assessment",
                "AR/VR Authentication",
                "Decentralized Identity",
                "AI-Driven Authentication Orchestration"
            ]
        },
        "forrester": {
            "year": 2025,
            "average_score": 75,
            "leader_score": 92,
            "leaders": ["Okta", "Microsoft", "IBM", "ForgeRock", "Ping Identity"],
            "strong_performers": ["OneLogin", "Auth0", "Thales", "SecureAuth", "Duo Security"],
            "key_capabilities": [
                "Adaptive Authentication",
                "FIDO2/WebAuthn Support",
                "Behavioral Biometrics",
                "Identity Governance",
                "Zero Trust Implementation"
            ]
        },
        "industry": {
            "financial": 85,
            "healthcare": 80,
            "government": 78,
            "retail": 68,
            "manufacturing": 65,
            "education": 60,
            "general": 70
        },
        "big_four": {
            "deloitte": {
                "leading_practices": [
                    "Continuous Authentication",
                    "Risk-Based Authorization",
                    "Identity Governance and Lifecycle Management",
                    "Zero Trust Architecture",
                    "Privileged Access Management"
                ]
            },
            "kpmg": {
                "leading_practices": [
                    "Decentralized Identity Integration",
                    "AI-Driven Identity Analytics",
                    "Cloud Identity Security",
                    "Identity-as-a-Code",
                    "Cross-Border Identity Compliance"
                ]
            },
            "pwc": {
                "leading_practices": [
                    "Digital Identity Wallets",
                    "Biometric Authentication Patterns",
                    "Privacy-Preserving Authentication",
                    "Supply Chain Identity Verification",
                    "IoT Device Identity Management"
                ]
            },
            "ey": {
                "leading_practices": [
                    "Web3 Identity Integration",
                    "Identity Attestation Services",
                    "Sectorial Identity Compliance",
                    "Next-Gen MFA Implementation",
                    "Digital Identity Inclusion"
                ]
            }
        }
    }
    
    registry.update_benchmark_data(benchmarks)
    
    return registry

# Função para aplicar todas as atualizações ao sistema
def apply_updates():
    """Aplica todas as atualizações ao sistema de compliance"""
    from .compliance_metadata import get_compliance_registry
    
    registry = get_compliance_registry()
    update_compliance_registry(registry)
    
    logger.info("Atualizações de compliance aplicadas com sucesso")
    return True
