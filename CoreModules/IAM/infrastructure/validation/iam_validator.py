"""
INNOVABIZ - Validador de Módulo IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Sistema de validação e certificação do módulo IAM,
           incluindo verificação de conformidade com múltiplos frameworks,
           testes de integração com módulos associados e certificação
           para fins de auditoria.
==================================================================
"""

import os
import json
import uuid
import logging
import hashlib
import datetime
import importlib
import dataclasses
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple, Set, Union, Callable
from enum import Enum, auto

# Configuração do logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("innovabiz.iam.validation")

# Enums para validação
class ValidationType(str, Enum):
    SECURITY = "security"
    PERFORMANCE = "performance"
    COMPLIANCE = "compliance"
    INTEGRATION = "integration"
    FUNCTIONAL = "functional"


class ValidationSeverity(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"


class ValidationStatus(str, Enum):
    PASSED = "passed"
    FAILED = "failed"
    WARNING = "warning"
    NOT_APPLICABLE = "not_applicable"
    PENDING = "pending"


@dataclasses.dataclass
class ValidationResult:
    """Resultado de uma validação específica"""
    id: str
    name: str
    description: str
    type: ValidationType
    severity: ValidationSeverity
    status: ValidationStatus
    timestamp: str
    details: Optional[str] = None
    affected_components: List[str] = dataclasses.field(default_factory=list)
    reference: Optional[str] = None
    metadata: Dict[str, Any] = dataclasses.field(default_factory=dict)
    remediation: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o resultado para dicionário"""
        return dataclasses.asdict(self)


@dataclasses.dataclass
class ValidationReport:
    """Relatório de validação completo"""
    id: str
    tenant_id: str
    timestamp: str
    framework: str
    version: str
    results: List[ValidationResult]
    summary: Dict[str, Any]
    metadata: Dict[str, Any] = dataclasses.field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o relatório para dicionário"""
        return {
            "id": self.id,
            "tenant_id": self.tenant_id,
            "timestamp": self.timestamp,
            "framework": self.framework,
            "version": self.version,
            "results": [r.to_dict() for r in self.results],
            "summary": self.summary,
            "metadata": self.metadata
        }
    
    def to_json(self, indent: int = 2) -> str:
        """Converte o relatório para JSON"""
        return json.dumps(self.to_dict(), indent=indent)
    
    def save(self, output_dir: Path, filename: Optional[str] = None) -> Path:
        """Salva o relatório como arquivo JSON"""
        if filename is None:
            timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"iam_validation_{self.framework}_{timestamp}.json"
        
        output_path = output_dir / filename
        with open(output_path, "w") as f:
            f.write(self.to_json())
        
        logger.info(f"Relatório de validação salvo em: {output_path}")
        return output_path
    
    @property
    def passed_count(self) -> int:
        """Retorna o número de validações que passaram"""
        return sum(1 for r in self.results if r.status == ValidationStatus.PASSED)
    
    @property
    def failed_count(self) -> int:
        """Retorna o número de validações que falharam"""
        return sum(1 for r in self.results if r.status == ValidationStatus.FAILED)
    
    @property
    def warning_count(self) -> int:
        """Retorna o número de validações com aviso"""
        return sum(1 for r in self.results if r.status == ValidationStatus.WARNING)
    
    @property
    def overall_status(self) -> ValidationStatus:
        """Retorna o status geral do relatório"""
        if self.failed_count > 0:
            return ValidationStatus.FAILED
        elif self.warning_count > 0:
            return ValidationStatus.WARNING
        else:
            return ValidationStatus.PASSED
    
    @property
    def critical_failures(self) -> List[ValidationResult]:
        """Retorna as falhas críticas"""
        return [r for r in self.results 
                if r.status == ValidationStatus.FAILED 
                and r.severity == ValidationSeverity.CRITICAL]


class ComplianceFramework(Enum):
    """Frameworks de conformidade suportados."""
    HIPAA = auto()
    GDPR = auto()
    PCI_DSS = auto()
    ISO27001 = auto()
    NIST = auto()
    SOC2 = auto()
    AR_AUTH = auto()  # Adicionando framework específico para Autenticação AR


@dataclass
class ComplianceRequirement:
    """Classe para representar um requisito de conformidade."""
    id: str
    name: str
    description: str
    category: str
    validation_func: Callable[[Dict[str, Any]], Dict[str, Any]]
    is_mandatory: bool = True
    applies_to_regions: List[str] = field(default_factory=lambda: [])
    metadata: Dict[str, Any] = field(default_factory=dict)


class IAMValidator:
    """Validador base para conformidade do IAM."""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
    
    def validate(self, iam_config: Dict, region: str) -> List[ValidationResult]:
        """Método não implementado na classe base."""
        raise NotImplementedError("Implemente na subclasse")
        
    def _filter_region_requirements(self, requirements: List[ComplianceRequirement], 
                                   region: str) -> List[ComplianceRequirement]:
        """Filtra requisitos aplicáveis à região específica."""
        if not region:
            return requirements
            
        # Se o requisito não tiver regiões especificadas, ele se aplica a todas as regiões
        return [
            req for req in requirements 
            if not req.applies_to_regions or region in req.applies_to_regions
        ]


class MultiRegionComplianceValidator:
    """Validador que suporta múltiplas regiões e frameworks."""
    
    def __init__(self, validators: Dict[ComplianceFramework, IAMValidator] = None):
        self.validators = validators or {}
        self.logger = logging.getLogger(__name__)
        
        # Carregar validadores adicionais
        self._load_ar_auth_validator()
    
    def _load_ar_auth_validator(self):
        """Carrega o validador de autenticação AR."""
        try:
            from .rules.ar_authentication_rules import ARAuthValidator
            self.validators[ComplianceFramework.AR_AUTH] = ARAuthValidator()
        except ImportError as e:
            self.logger.warning(f"Não foi possível carregar o validador AR_AUTH: {e}")
    
    def validate_all_regions(self, iam_config: Dict) -> Dict[str, Dict[ComplianceFramework, List[ValidationResult]]]:
        """Valida a configuração IAM em todas as regiões com todos os validadores registrados."""
        results = {}
        
        # Para cada região suportada
        for region in ["us-east-1", "us-west-2", "eu-west-1"]:  # Exemplo de regiões
            results[region] = {}
            region_config = self._get_region_config(iam_config, region)
            
            # Para cada framework de conformidade
            for framework, validator in self.validators.items():
                try:
                    # Executar validação para este framework nesta região
                    validation_results = validator.validate(region_config, region)
                    results[region][framework] = validation_results
                except Exception as e:
                    self.logger.error(f"Erro na validação de {framework.name} na região {region}: {e}")
                    results[region][framework] = []
        
        return results
        
    def _get_region_config(self, iam_config: Dict, region: str) -> Dict:
        """Extrai configuração específica da região ou retorna configuração global."""
        if not iam_config:
            return {}
            
        # Verificar se existe configuração específica para a região
        regions_config = iam_config.get("regions", {})
        region_specific = regions_config.get(region, {})
        
        # Mesclar configuração específica com global, com precedência para a específica
        global_config = {k: v for k, v in iam_config.items() if k != "regions"}
        
        # Configuração final prioriza configurações específicas da região
        return {**global_config, **region_specific}


class IAMValidator:
    """
    Validador do módulo IAM para conformidade, segurança e integridade.
    Realiza verificações abrangentes e gera relatórios de validação.
    """
    
    def __init__(self, base_dir: Optional[Union[str, Path]] = None):
        """
        Inicializa o validador IAM
        
        Args:
            base_dir: Diretório base do projeto INNOVABIZ (opcional)
        """
        self.base_dir = Path(base_dir) if base_dir else Path(__file__).parent.parent.parent.parent
        self.iam_dir = self.base_dir / "infrastructure" / "iam"
        self.validation_rules_dir = self.iam_dir / "validation" / "rules"
        self.reports_dir = self.iam_dir / "validation" / "reports"
        
        # Garantir que diretórios existam
        self.reports_dir.mkdir(exist_ok=True, parents=True)
        
        # Cache de módulos para validação
        self._validator_modules = {}
        
        logger.info(f"IAMValidator inicializado com diretório base: {self.base_dir}")
    
    def _load_validation_module(self, framework: str) -> Any:
        """
        Carrega dinamicamente o módulo de validação para um framework
        
        Args:
            framework: Nome do framework de validação
        
        Returns:
            Módulo de validação
        """
        if framework in self._validator_modules:
            return self._validator_modules[framework]
        
        try:
            module_path = f"infrastructure.iam.validation.rules.{framework}_rules"
            module = importlib.import_module(module_path)
            self._validator_modules[framework] = module
            return module
        except ImportError as e:
            logger.error(f"Erro ao carregar módulo de validação para '{framework}': {e}")
            raise ValueError(f"Framework de validação não suportado: {framework}")
    
    def validate(self, 
                tenant_id: str, 
                framework: str,
                iam_config: Optional[Dict[str, Any]] = None) -> ValidationReport:
        """
        Executa validação do IAM para um framework específico
        
        Args:
            tenant_id: ID do tenant
            framework: Framework de validação (ex: 'hipaa', 'gdpr', 'security')
            iam_config: Configuração IAM (opcional, será carregada se não fornecida)
        
        Returns:
            Relatório de validação
        """
        # Carregar módulo do validador
        validator_module = self._load_validation_module(framework)
        
        # Carregar configuração IAM se não fornecida
        if iam_config is None:
            iam_config = self._load_iam_config(tenant_id)
        
        # Executar validações
        logger.info(f"Iniciando validação IAM para framework '{framework}', tenant: {tenant_id}")
        validator = validator_module.Validator(tenant_id, iam_config)
        results = validator.run_validations()
        
        # Criar relatório
        timestamp = datetime.datetime.now().isoformat()
        report_id = str(uuid.uuid4())
        
        # Criar resumo
        summary = {
            "total": len(results),
            "passed": sum(1 for r in results if r.status == ValidationStatus.PASSED),
            "failed": sum(1 for r in results if r.status == ValidationStatus.FAILED),
            "warning": sum(1 for r in results if r.status == ValidationStatus.WARNING),
            "not_applicable": sum(1 for r in results if r.status == ValidationStatus.NOT_APPLICABLE),
            "by_severity": {
                "critical": sum(1 for r in results 
                               if r.severity == ValidationSeverity.CRITICAL 
                               and r.status != ValidationStatus.PASSED),
                "high": sum(1 for r in results 
                           if r.severity == ValidationSeverity.HIGH 
                           and r.status != ValidationStatus.PASSED),
                "medium": sum(1 for r in results 
                             if r.severity == ValidationSeverity.MEDIUM 
                             and r.status != ValidationStatus.PASSED),
                "low": sum(1 for r in results 
                          if r.severity == ValidationSeverity.LOW 
                          and r.status != ValidationStatus.PASSED)
            },
            "by_type": {
                "security": sum(1 for r in results if r.type == ValidationType.SECURITY),
                "compliance": sum(1 for r in results if r.type == ValidationType.COMPLIANCE),
                "integration": sum(1 for r in results if r.type == ValidationType.INTEGRATION),
                "performance": sum(1 for r in results if r.type == ValidationType.PERFORMANCE),
                "functional": sum(1 for r in results if r.type == ValidationType.FUNCTIONAL)
            }
        }
        
        # Metadados adicionais
        metadata = {
            "validator_version": "1.0.0",
            "validation_timestamp": timestamp,
            "framework_version": getattr(validator_module, "VERSION", "1.0.0"),
            "tenant_id": tenant_id,
            "execution_environment": {
                "python_version": os.sys.version,
                "platform": os.sys.platform
            }
        }
        
        # Criar relatório
        report = ValidationReport(
            id=report_id,
            tenant_id=tenant_id,
            timestamp=timestamp,
            framework=framework,
            version=metadata["framework_version"],
            results=results,
            summary=summary,
            metadata=metadata
        )
        
        # Salvar relatório
        report_path = report.save(self.reports_dir)
        logger.info(f"Validação concluída. Relatório salvo em: {report_path}")
        
        return report
    
    def validate_all_frameworks(self, 
                              tenant_id: str,
                              frameworks: Optional[List[str]] = None) -> Dict[str, ValidationReport]:
        """
        Executa validação do IAM para múltiplos frameworks
        
        Args:
            tenant_id: ID do tenant
            frameworks: Lista de frameworks para validar (opcional)
        
        Returns:
            Dicionário com relatórios de validação por framework
        """
        # Determinar frameworks a serem validados
        if frameworks is None:
            frameworks = self._discover_available_frameworks()
        
        # Carregar configuração IAM uma vez para todas as validações
        iam_config = self._load_iam_config(tenant_id)
        
        # Executar validações para cada framework
        reports = {}
        for framework in frameworks:
            try:
                reports[framework] = self.validate(tenant_id, framework, iam_config)
            except Exception as e:
                logger.error(f"Erro ao validar framework '{framework}': {e}")
                # Continuar para o próximo framework em caso de erro
        
        return reports
    
    def _discover_available_frameworks(self) -> List[str]:
        """
        Descobre automaticamente frameworks de validação disponíveis
        
        Returns:
            Lista de frameworks disponíveis
        """
        frameworks = []
        
        # Verificar diretório de regras de validação
        if self.validation_rules_dir.exists():
            for file_path in self.validation_rules_dir.glob("*_rules.py"):
                framework = file_path.stem.replace("_rules", "")
                frameworks.append(framework)
        
        return frameworks
    
    def _load_iam_config(self, tenant_id: str) -> Dict[str, Any]:
        """
        Carrega a configuração IAM para um tenant
        
        Args:
            tenant_id: ID do tenant
        
        Returns:
            Configuração IAM
        """
        # Na implementação real, isso carregaria a configuração do banco de dados
        # Para fins de demonstração, usamos um exemplo
        
        # Tentar carregar de um arquivo de exemplo
        example_path = self.iam_dir / "examples" / f"iam_config_{tenant_id}.json"
        if example_path.exists():
            with open(example_path, "r") as f:
                return json.load(f)
        
        # Caso contrário, usar configuração padrão
        default_path = self.iam_dir / "examples" / "default_iam_config.json"
        if default_path.exists():
            with open(default_path, "r") as f:
                return json.load(f)
        
        # Se nem isso existir, criar uma configuração básica
        return {
            "tenant_id": tenant_id,
            "authentication": {
                "mfa_enabled": True,
                "mfa_methods": ["totp", "sms", "email"]
            },
            "authorization": {
                "rbac_enabled": True,
                "abac_enabled": False
            },
            "session_management": {
                "timeout_minutes": 30
            },
            "audit": {
                "enabled": True,
                "log_retention_days": 90
            }
        }
    
    def generate_certification(self, 
                             tenant_id: str,
                             reports: Dict[str, ValidationReport],
                             output_dir: Optional[Path] = None) -> Path:
        """
        Gera certificado de validação para o módulo IAM
        
        Args:
            tenant_id: ID do tenant
            reports: Relatórios de validação
            output_dir: Diretório para salvar o certificado
        
        Returns:
            Caminho para o certificado gerado
        """
        if output_dir is None:
            output_dir = self.reports_dir / "certifications"
            output_dir.mkdir(exist_ok=True, parents=True)
        
        # Verificar se todos os requisitos críticos foram atendidos
        all_passed = True
        critical_issues = []
        
        for framework, report in reports.items():
            if report.overall_status == ValidationStatus.FAILED:
                all_passed = False
                critical_issues.extend([
                    {
                        "framework": framework,
                        "id": issue.id,
                        "name": issue.name,
                        "severity": issue.severity,
                        "details": issue.details
                    }
                    for issue in report.critical_failures
                ])
        
        # Criar certificado
        timestamp = datetime.datetime.now().isoformat()
        expiration = (datetime.datetime.now() + datetime.timedelta(days=365)).isoformat()
        
        # Calcular hash do certificado para verificação
        hash_data = f"{tenant_id}:{timestamp}:{expiration}"
        for framework in sorted(reports.keys()):
            hash_data += f":{framework}:{reports[framework].id}"
        
        certificate_hash = hashlib.sha256(hash_data.encode()).hexdigest()
        
        certificate = {
            "id": str(uuid.uuid4()),
            "tenant_id": tenant_id,
            "timestamp": timestamp,
            "expiration": expiration,
            "status": "approved" if all_passed else "rejected",
            "frameworks": list(reports.keys()),
            "reports": {framework: report.id for framework, report in reports.items()},
            "verification_hash": certificate_hash,
            "critical_issues": critical_issues,
            "metadata": {
                "issuer": "INNOVABIZ IAM Validator",
                "version": "1.0.0",
                "validity_period_days": 365
            }
        }
        
        # Salvar certificado
        timestamp_str = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        certificate_path = output_dir / f"iam_certificate_{tenant_id}_{timestamp_str}.json"
        
        with open(certificate_path, "w") as f:
            json.dump(certificate, f, indent=2)
        
        logger.info(f"Certificado de validação IAM gerado: {certificate_path}")
        return certificate_path
    
    def verify_certification(self, certificate_path: Union[str, Path]) -> bool:
        """
        Verifica a autenticidade de um certificado
        
        Args:
            certificate_path: Caminho para o certificado
        
        Returns:
            True se o certificado for válido
        """
        try:
            # Carregar certificado
            with open(certificate_path, "r") as f:
                certificate = json.load(f)
            
            # Recalcular hash para verificação
            hash_data = f"{certificate['tenant_id']}:{certificate['timestamp']}:{certificate['expiration']}"
            for framework in sorted(certificate['frameworks']):
                hash_data += f":{framework}:{certificate['reports'][framework]}"
            
            calculated_hash = hashlib.sha256(hash_data.encode()).hexdigest()
            
            # Verificar hash
            if calculated_hash != certificate['verification_hash']:
                logger.warning(f"Hash de verificação inválido para certificado: {certificate_path}")
                return False
            
            # Verificar se não está expirado
            expiration = datetime.datetime.fromisoformat(certificate['expiration'])
            if expiration < datetime.datetime.now():
                logger.warning(f"Certificado expirado: {certificate_path}")
                return False
            
            return True
        
        except Exception as e:
            logger.error(f"Erro ao verificar certificado: {e}")
            return False


# Função principal para execução como script
def main():
    """Função principal para execução como script"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Validador de Módulo IAM")
    parser.add_argument("--tenant", required=True, help="ID do tenant")
    parser.add_argument("--frameworks", nargs="*", help="Frameworks para validar")
    parser.add_argument("--output", help="Diretório para salvar relatórios")
    parser.add_argument("--certify", action="store_true", help="Gerar certificado de validação")
    
    args = parser.parse_args()
    
    # Inicializar validador
    validator = IAMValidator()
    
    # Executar validações
    if args.frameworks:
        reports = validator.validate_all_frameworks(args.tenant, args.frameworks)
    else:
        reports = validator.validate_all_frameworks(args.tenant)
    
    # Gerar certificado se solicitado
    if args.certify:
        output_dir = Path(args.output) if args.output else None
        certificate_path = validator.generate_certification(args.tenant, reports, output_dir)
        print(f"Certificado gerado: {certificate_path}")


if __name__ == "__main__":
    main()
