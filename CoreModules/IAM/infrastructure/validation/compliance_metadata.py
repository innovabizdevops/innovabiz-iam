"""
INNOVABIZ - Sistema de Metadados de Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Sistema de metadados para certificação e validação IAM
           baseado em normas, frameworks, padrões, legislações e
           regulamentações específicas por região, indústria e setor.
==================================================================
"""

from enum import Enum, auto
from typing import Dict, List, Any, Optional, Set, Tuple
from dataclasses import dataclass, field
import datetime
import json
from pathlib import Path

# Regiões suportadas pelo INNOVABIZ
class Region(Enum):
    US_EAST = "us_east"               # EUA (Leste)
    US_WEST = "us_west"               # EUA (Oeste)
    EU_CENTRAL = "eu_central"         # União Europeia (Central/Portugal)
    EU_NORTH = "eu_north"             # União Europeia (Norte)
    EU_SOUTH = "eu_south"             # União Europeia (Sul)
    EU_WEST = "eu_west"               # União Europeia (Oeste)
    BR = "br"                         # Brasil
    AF_ANGOLA = "af_angola"           # Angola
    AF_CONGO = "af_congo"             # Congo
    GLOBAL = "global"                 # Global (aplicável a todas as regiões)


# Setores/Indústrias suportados pelo INNOVABIZ
class Industry(Enum):
    HEALTHCARE = "healthcare"         # Saúde
    FINANCIAL = "financial"           # Financeiro
    INSURANCE = "insurance"           # Seguros
    RETAIL = "retail"                 # Varejo
    MANUFACTURING = "manufacturing"   # Manufatura
    TELECOM = "telecom"               # Telecomunicações
    GOVERNMENT = "government"         # Governo
    EDUCATION = "education"           # Educação
    ENERGY = "energy"                 # Energia
    GENERAL = "general"               # Geral (aplicável a todos os setores)


# Frameworks de Compliance
class ComplianceFramework(Enum):
    # Frameworks gerais de segurança
    ISO27001 = "iso27001"             # ISO/IEC 27001:2013
    NIST_CSF = "nist_csf"             # NIST Cybersecurity Framework
    NIST_800_53 = "nist_800_53"       # NIST SP 800-53
    NIST_800_63 = "nist_800_63"       # NIST SP 800-63 (Digital Identity)
    CIS_CONTROLS = "cis_controls"     # CIS Controls
    COBIT = "cobit"                   # COBIT 2019
    SOC2 = "soc2"                     # SOC 2
    
    # Frameworks financeiros
    PCI_DSS = "pci_dss"               # Payment Card Industry Data Security Standard
    PSD2 = "psd2"                     # EU Payment Services Directive 2
    SOX = "sox"                       # Sarbanes-Oxley Act
    BASEL = "basel"                   # Basel II/III
    
    # Frameworks de saúde
    HIPAA = "hipaa"                   # Health Insurance Portability and Accountability Act
    HITRUST = "hitrust"               # HITRUST CSF
    NHS_DS = "nhs_ds"                 # NHS Data Security
    
    # Frameworks de privacidade
    GDPR = "gdpr"                     # General Data Protection Regulation (UE)
    LGPD = "lgpd"                     # Lei Geral de Proteção de Dados (Brasil)
    CCPA = "ccpa"                     # California Consumer Privacy Act
    PIPEDA = "pipeda"                 # Personal Information Protection and Electronic Documents Act (Canadá)
    
    # Frameworks de autenticação
    FIDO2 = "fido2"                   # FIDO2 WebAuthn/CTAP
    ZERO_TRUST = "zero_trust"         # Zero Trust Architecture
    OPENID = "openid"                 # OpenID Connect
    SAML = "saml"                     # SAML 2.0
    
    # Frameworks de IAM específicos
    GARTNER_IAM = "gartner_iam"       # Gartner IAM Benchmark
    FORRESTER_IAM = "forrester_iam"   # Forrester IAM Wave
    OAUTH2 = "oauth2"                 # OAuth 2.0
    IDAM_GOOD_PRACTICE = "idam_gp"    # Identity and Access Management Good Practice
    
    # Frameworks AR/VR
    IEEE_2888 = "ieee_2888"           # IEEE 2888 (AR/VR)
    XR_SECURITY = "xr_security"       # XR Security Framework
    
    # Frameworks específicos de Angola
    PNDSB = "pndsb"                   # Política Nacional para Desenvolvimento de Serviços Bancários (Angola)


@dataclass
class ComplianceRequirement:
    """Requisito de conformidade específico."""
    id: str
    framework: ComplianceFramework
    name: str
    description: str
    industries: List[Industry] = field(default_factory=list)
    regions: List[Region] = field(default_factory=list)
    benchmark_score: Optional[float] = None  # Pontuação de referência baseada em benchmarks
    validation_rules: List[str] = field(default_factory=list)  # IDs das regras de validação
    implementation_date: Optional[datetime.date] = None  # Data de implementação do requisito
    sunset_date: Optional[datetime.date] = None  # Data de obsolescência (se aplicável)
    documentation_urls: List[str] = field(default_factory=list)  # URLs para documentação
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class AuthenticationFactor:
    """Definição de um fator de autenticação."""
    id: str
    name: str
    type: str  # knowledge, possession, inherence, behavioral, contextual, spatial
    strength: float  # 0.0 a 1.0, onde 1.0 é o mais forte
    assurance_level: int  # 1-3, mapeando para AAL1, AAL2, AAL3 do NIST
    frameworks: List[ComplianceFramework] = field(default_factory=list)
    industries: List[Industry] = field(default_factory=list)
    regions: List[Region] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)


class ComplianceMetadataRegistry:
    """Registro de metadados de compliance para IAM."""
    
    def __init__(self, data_path: Optional[Path] = None):
        """
        Inicializa o registro de metadados de compliance.
        
        Args:
            data_path: Caminho para o diretório de dados de compliance
        """
        self.data_path = data_path
        self.requirements: Dict[str, ComplianceRequirement] = {}
        self.factors: Dict[str, AuthenticationFactor] = {}
        self._load_data()
    
    def _load_data(self):
        """Carrega dados de compliance do disco ou inicializa com padrões."""
        # Implementação real carregaria de arquivos JSON ou banco de dados
        # Para fins de demonstração, inicializamos com alguns padrões
        self._initialize_default_requirements()
        self._initialize_default_factors()
    
    def _initialize_default_requirements(self):
        """Inicializa requisitos de compliance padrão."""
        # Exemplos de requisitos NIST 800-63 para autenticação digital
        self.requirements["nist_800_63_aal2"] = ComplianceRequirement(
            id="nist_800_63_aal2",
            framework=ComplianceFramework.NIST_800_63,
            name="NIST 800-63 AAL2",
            description="Nível de garantia de autenticação 2, exigindo MFA com dois fatores distintos",
            industries=[Industry.GENERAL],
            regions=[Region.US_EAST, Region.US_WEST, Region.GLOBAL],
            benchmark_score=0.8,
            validation_rules=["validate_mfa_diversity", "validate_channel_separation"]
        )
        
        # Exemplo de requisito GDPR para autenticação
        self.requirements["gdpr_auth"] = ComplianceRequirement(
            id="gdpr_auth",
            framework=ComplianceFramework.GDPR,
            name="GDPR Autenticação Segura",
            description="Controles técnicos apropriados para proteger dados pessoais, incluindo autenticação forte",
            industries=[Industry.GENERAL],
            regions=[Region.EU_CENTRAL, Region.EU_NORTH, Region.EU_SOUTH, Region.EU_WEST],
            benchmark_score=0.85,
            validation_rules=["validate_secure_auth", "validate_data_protection"]
        )
        
        # Exemplo de requisito PSD2 para autenticação forte
        self.requirements["psd2_sca"] = ComplianceRequirement(
            id="psd2_sca",
            framework=ComplianceFramework.PSD2,
            name="PSD2 SCA",
            description="Autenticação forte do cliente (SCA) para transações de pagamento",
            industries=[Industry.FINANCIAL],
            regions=[Region.EU_CENTRAL, Region.EU_NORTH, Region.EU_SOUTH, Region.EU_WEST],
            benchmark_score=0.9,
            validation_rules=["validate_dynamic_linking", "validate_mfa_diversity"]
        )
        
        # Exemplo de requisito HIPAA para autenticação em saúde
        self.requirements["hipaa_auth"] = ComplianceRequirement(
            id="hipaa_auth",
            framework=ComplianceFramework.HIPAA,
            name="HIPAA Autenticação",
            description="Implementação de procedimentos para verificar que uma pessoa ou entidade buscando acesso a PHI é quem alega ser",
            industries=[Industry.HEALTHCARE],
            regions=[Region.US_EAST, Region.US_WEST],
            benchmark_score=0.85,
            validation_rules=["validate_hipaa_auth", "validate_audit_logging"]
        )
        
        # Exemplo de requisito PNDSB (Angola) para autenticação
        self.requirements["pndsb_auth"] = ComplianceRequirement(
            id="pndsb_auth",
            framework=ComplianceFramework.PNDSB,
            name="PNDSB Autenticação Segura",
            description="Requisitos de autenticação segura para serviços bancários em Angola",
            industries=[Industry.FINANCIAL],
            regions=[Region.AF_ANGOLA],
            benchmark_score=0.75,
            validation_rules=["validate_mfa", "validate_transaction_auth"]
        )
        
        # Exemplo de requisito LGPD (Brasil) para autenticação
        self.requirements["lgpd_auth"] = ComplianceRequirement(
            id="lgpd_auth",
            framework=ComplianceFramework.LGPD,
            name="LGPD Autenticação",
            description="Medidas de segurança, técnicas e administrativas para proteger dados pessoais, incluindo autenticação segura",
            industries=[Industry.GENERAL],
            regions=[Region.BR],
            benchmark_score=0.8,
            validation_rules=["validate_secure_auth", "validate_data_protection"]
        )
        
        # Exemplo de requisito AR/VR para autenticação em saúde
        self.requirements["ar_health_auth"] = ComplianceRequirement(
            id="ar_health_auth",
            framework=ComplianceFramework.XR_SECURITY,
            name="AR Healthcare Autenticação",
            description="Requisitos de autenticação para aplicações de saúde em AR",
            industries=[Industry.HEALTHCARE],
            regions=[Region.GLOBAL],
            benchmark_score=0.85,
            validation_rules=["validate_ar_auth", "validate_spatial_factors"]
        )
        
        # Exemplo de benchmark Gartner para IAM
        self.requirements["gartner_iam_benchmark"] = ComplianceRequirement(
            id="gartner_iam_benchmark",
            framework=ComplianceFramework.GARTNER_IAM,
            name="Gartner IAM Benchmark 2025",
            description="Benchmark Gartner para IAM de classe mundial, incluindo autenticação adaptativa e zero trust",
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL],
            benchmark_score=0.9,
            validation_rules=["validate_adaptive_auth", "validate_zero_trust"]
        )
        
        # Exemplo de benchmark Forrester para IAM
        self.requirements["forrester_iam_wave"] = ComplianceRequirement(
            id="forrester_iam_wave",
            framework=ComplianceFramework.FORRESTER_IAM,
            name="Forrester IAM Wave 2025",
            description="Critérios Forrester Wave para IAM líderes, incluindo autenticação contínua e baseada em risco",
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL],
            benchmark_score=0.85,
            validation_rules=["validate_continuous_auth", "validate_risk_based_auth"]
        )
    
    def _initialize_default_factors(self):
        """Inicializa fatores de autenticação padrão."""
        # Fatores de conhecimento
        self.factors["password"] = AuthenticationFactor(
            id="password",
            name="Senha",
            type="knowledge",
            strength=0.3,
            assurance_level=1,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.ISO27001,
                ComplianceFramework.GDPR
            ],
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        self.factors["pin"] = AuthenticationFactor(
            id="pin",
            name="PIN",
            type="knowledge",
            strength=0.2,
            assurance_level=1,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.PCI_DSS
            ],
            industries=[Industry.FINANCIAL, Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        # Fatores de posse
        self.factors["totp"] = AuthenticationFactor(
            id="totp",
            name="TOTP",
            type="possession",
            strength=0.7,
            assurance_level=2,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.PSD2,
                ComplianceFramework.GDPR
            ],
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        self.factors["fido_key"] = AuthenticationFactor(
            id="fido_key",
            name="FIDO Security Key",
            type="possession",
            strength=0.9,
            assurance_level=3,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.FIDO2,
                ComplianceFramework.ZERO_TRUST
            ],
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        # Fatores biométricos
        self.factors["fingerprint"] = AuthenticationFactor(
            id="fingerprint",
            name="Impressão Digital",
            type="inherence",
            strength=0.8,
            assurance_level=2,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.GDPR,
                ComplianceFramework.HIPAA
            ],
            industries=[Industry.GENERAL, Industry.HEALTHCARE, Industry.FINANCIAL],
            regions=[Region.GLOBAL]
        )
        
        self.factors["facial"] = AuthenticationFactor(
            id="facial",
            name="Reconhecimento Facial",
            type="inherence",
            strength=0.85,
            assurance_level=2,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.GDPR,
                ComplianceFramework.PNDSB
            ],
            industries=[Industry.GENERAL, Industry.FINANCIAL],
            regions=[Region.GLOBAL, Region.AF_ANGOLA]
        )
        
        # Fatores comportamentais
        self.factors["typing_pattern"] = AuthenticationFactor(
            id="typing_pattern",
            name="Padrão de Digitação",
            type="behavioral",
            strength=0.6,
            assurance_level=2,
            frameworks=[
                ComplianceFramework.NIST_800_63,
                ComplianceFramework.FORRESTER_IAM
            ],
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        # Fatores contextuais
        self.factors["geo_location"] = AuthenticationFactor(
            id="geo_location",
            name="Localização Geográfica",
            type="contextual",
            strength=0.5,
            assurance_level=1,
            frameworks=[
                ComplianceFramework.GARTNER_IAM,
                ComplianceFramework.FORRESTER_IAM
            ],
            industries=[Industry.GENERAL],
            regions=[Region.GLOBAL]
        )
        
        # Fatores espaciais (AR)
        self.factors["spatial_gesture"] = AuthenticationFactor(
            id="spatial_gesture",
            name="Gesto Espacial 3D",
            type="spatial",
            strength=0.8,
            assurance_level=3,
            frameworks=[
                ComplianceFramework.IEEE_2888,
                ComplianceFramework.XR_SECURITY
            ],
            industries=[Industry.HEALTHCARE, Industry.FINANCIAL],
            regions=[Region.GLOBAL]
        )
        
        self.factors["gaze_pattern"] = AuthenticationFactor(
            id="gaze_pattern",
            name="Padrão de Olhar",
            type="spatial",
            strength=0.85,
            assurance_level=3,
            frameworks=[
                ComplianceFramework.IEEE_2888,
                ComplianceFramework.XR_SECURITY
            ],
            industries=[Industry.HEALTHCARE, Industry.FINANCIAL],
            regions=[Region.GLOBAL]
        )
    
    def get_requirement(self, requirement_id: str) -> Optional[ComplianceRequirement]:
        """
        Obtém um requisito de compliance pelo ID.
        
        Args:
            requirement_id: ID do requisito
            
        Returns:
            Requisito de compliance ou None se não encontrado
        """
        return self.requirements.get(requirement_id)
    
    def get_factor(self, factor_id: str) -> Optional[AuthenticationFactor]:
        """
        Obtém um fator de autenticação pelo ID.
        
        Args:
            factor_id: ID do fator
            
        Returns:
            Fator de autenticação ou None se não encontrado
        """
        return self.factors.get(factor_id)
    
    def get_requirements_by_region(self, region: Region) -> List[ComplianceRequirement]:
        """
        Obtém requisitos de compliance aplicáveis a uma região.
        
        Args:
            region: Região desejada
            
        Returns:
            Lista de requisitos de compliance
        """
        return [
            req for req in self.requirements.values()
            if region in req.regions or Region.GLOBAL in req.regions
        ]
    
    def get_requirements_by_industry(self, industry: Industry) -> List[ComplianceRequirement]:
        """
        Obtém requisitos de compliance aplicáveis a uma indústria.
        
        Args:
            industry: Indústria desejada
            
        Returns:
            Lista de requisitos de compliance
        """
        return [
            req for req in self.requirements.values()
            if industry in req.industries or Industry.GENERAL in req.industries
        ]
    
    def get_requirements_by_framework(self, framework: ComplianceFramework) -> List[ComplianceRequirement]:
        """
        Obtém requisitos de compliance de um framework.
        
        Args:
            framework: Framework desejado
            
        Returns:
            Lista de requisitos de compliance
        """
        return [
            req for req in self.requirements.values()
            if req.framework == framework
        ]
    
    def get_factors_by_requirements(self, requirements: List[ComplianceRequirement]) -> Dict[str, AuthenticationFactor]:
        """
        Obtém fatores de autenticação que atendem a um conjunto de requisitos.
        
        Args:
            requirements: Lista de requisitos
            
        Returns:
            Dicionário de fatores de autenticação
        """
        frameworks = set()
        for req in requirements:
            frameworks.add(req.framework)
        
        return {
            factor_id: factor for factor_id, factor in self.factors.items()
            if any(framework in factor.frameworks for framework in frameworks)
        }
    
    def get_benchmark_data(self) -> Dict[str, Any]:
        """
        Obtém dados de benchmark para comparação.
        
        Returns:
            Dados de benchmark
        """
        # Na implementação real, isso poderia vir de APIs externas ou banco de dados
        return {
            "gartner": {
                "average_score": 0.65,
                "leader_score": 0.85,
                "innovator_score": 0.9
            },
            "forrester": {
                "average_score": 0.7,
                "leader_score": 0.88,
                "strong_performer_score": 0.82
            },
            "industry": {
                "financial": 0.75,
                "healthcare": 0.72,
                "retail": 0.68,
                "government": 0.8
            }
        }


# Singleton para acesso global ao registro
_registry_instance = None

def get_compliance_registry(data_path: Optional[Path] = None) -> ComplianceMetadataRegistry:
    """
    Obtém a instância singleton do registro de metadados de compliance.
    
    Args:
        data_path: Caminho opcional para dados de compliance
        
    Returns:
        Instância do registro de metadados de compliance
    """
    global _registry_instance
    if _registry_instance is None:
        _registry_instance = ComplianceMetadataRegistry(data_path)
    return _registry_instance
