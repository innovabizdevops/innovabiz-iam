"""
INNOVABIZ - Motor de Validação de Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Motor de validação de compliance para IAM baseado em
           normas, frameworks, padrões e regulamentações regionais
           e setoriais específicas.
==================================================================
"""

import json
import logging
import datetime
from typing import Dict, List, Any, Optional, Set, Tuple
from dataclasses import dataclass, field
from pathlib import Path
import importlib
import inspect

from .compliance_metadata import (
    ComplianceMetadataRegistry, 
    get_compliance_registry,
    Region, 
    Industry, 
    ComplianceFramework,
    ComplianceRequirement,
    AuthenticationFactor
)

from .models import (
    ComplianceValidationResult,
    ValidationSeverity,
    ValidationReport,
    ValidationStatus
)

from ..adaptive.risk_engine import RiskLevel

# Configuração de logging
logger = logging.getLogger(__name__)


@dataclass
class ValidationContext:
    """Contexto para validação de compliance."""
    tenant_id: str
    region: Region
    industry: Industry
    frameworks: List[ComplianceFramework] = field(default_factory=list)
    config: Dict[str, Any] = field(default_factory=dict)
    benchmarks: Dict[str, Any] = field(default_factory=dict)
    metadata: Dict[str, Any] = field(default_factory=dict)


class ValidationRule:
    """Interface para regras de validação."""
    
    def __init__(self, rule_id: str, name: str, description: str, severity: ValidationSeverity = ValidationSeverity.HIGH):
        """
        Inicializa uma regra de validação.
        
        Args:
            rule_id: ID da regra
            name: Nome da regra
            description: Descrição da regra
            severity: Severidade da validação
        """
        self.id = rule_id
        self.name = name
        self.description = description
        self.severity = severity
    
    def validate(self, context: ValidationContext) -> List[ComplianceValidationResult]:
        """
        Executa a validação.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de resultados da validação
        """
        raise NotImplementedError("Implemente na subclasse")
    
    def get_requirements(self) -> List[str]:
        """
        Obtém os IDs dos requisitos que esta regra valida.
        
        Returns:
            Lista de IDs de requisitos
        """
        raise NotImplementedError("Implemente na subclasse")
    
    def get_applicable_regions(self) -> List[Region]:
        """
        Obtém as regiões aplicáveis para esta regra.
        
        Returns:
            Lista de regiões
        """
        return [Region.GLOBAL]  # Por padrão, aplicável globalmente
    
    def get_applicable_industries(self) -> List[Industry]:
        """
        Obtém as indústrias aplicáveis para esta regra.
        
        Returns:
            Lista de indústrias
        """
        return [Industry.GENERAL]  # Por padrão, aplicável a todas as indústrias
    
    def get_applicable_frameworks(self) -> List[ComplianceFramework]:
        """
        Obtém os frameworks aplicáveis para esta regra.
        
        Returns:
            Lista de frameworks
        """
        return []  # Deve ser implementado nas subclasses


class ComplianceEngine:
    """Motor de validação de compliance para IAM."""
    
    def __init__(self, data_path: Optional[Path] = None):
        """
        Inicializa o motor de validação de compliance.
        
        Args:
            data_path: Caminho para dados de compliance
        """
        self.registry = get_compliance_registry(data_path)
        self.rules: Dict[str, ValidationRule] = {}
        self._load_validation_rules()
    
    def _load_validation_rules(self):
        """Carrega todas as regras de validação disponíveis."""
        # Em uma implementação completa, isso carregaria dinamicamente todas as classes
        # de regras disponíveis em diretórios específicos.
        # Para simplificar, carregaremos manualmente algumas regras básicas
        
        # Importar módulos de regras
        try:
            from .rules import adaptive_mfa_rules
            self._register_rules_from_module(adaptive_mfa_rules)
        except ImportError:
            logger.warning("Módulo adaptive_mfa_rules não encontrado")
        
        try:
            from .rules import ar_authentication_rules
            self._register_rules_from_module(ar_authentication_rules)
        except ImportError:
            logger.warning("Módulo ar_authentication_rules não encontrado")
        
        try:
            from .rules import regulatory_compliance_rules
            self._register_rules_from_module(regulatory_compliance_rules)
        except ImportError:
            logger.warning("Módulo regulatory_compliance_rules não encontrado")
        
        try:
            from .rules import industry_specific_rules
            self._register_rules_from_module(industry_specific_rules)
        except ImportError:
            logger.warning("Módulo industry_specific_rules não encontrado")
        
        # Importar validadores regionais
        try:
            from .validators import get_all_validators
            for rule in get_all_validators():
                self.register_rule(rule)
            logger.info("Validadores regionais carregados com sucesso")
        except ImportError as e:
            logger.warning(f"Erro ao importar validadores regionais: {str(e)}")
        
        # Importar validadores regionais específicos caso a importação geral falhe
        if "get_all_validators" not in locals():
            try:
                from .validators.eu_validator import get_eu_rules
                for rule in get_eu_rules():
                    self.register_rule(rule)
                logger.info("Validadores para UE carregados com sucesso")
            except ImportError:
                logger.warning("Módulo eu_validator não encontrado")
                
            try:
                from .validators.brazil_validator import get_brazil_rules
                for rule in get_brazil_rules():
                    self.register_rule(rule)
                logger.info("Validadores para Brasil carregados com sucesso")
            except ImportError:
                logger.warning("Módulo brazil_validator não encontrado")
                
            try:
                from .validators.us_validator import get_us_rules
                for rule in get_us_rules():
                    self.register_rule(rule)
                logger.info("Validadores para EUA carregados com sucesso")
            except ImportError:
                logger.warning("Módulo us_validator não encontrado")
                
            try:
                from .validators.africa_validator import get_africa_rules
                for rule in get_africa_rules():
                    self.register_rule(rule)
                logger.info("Validadores para África carregados com sucesso")
            except ImportError:
                logger.warning("Módulo africa_validator não encontrado")
    
    def _register_rules_from_module(self, module):
        """
        Registra todas as regras de validação de um módulo.
        
        Args:
            module: Módulo contendo regras de validação
        """
        for name, obj in inspect.getmembers(module):
            if (inspect.isclass(obj) and 
                issubclass(obj, ValidationRule) and 
                obj is not ValidationRule):
                try:
                    rule_instance = obj()
                    self.rules[rule_instance.id] = rule_instance
                    logger.info(f"Regra registrada: {rule_instance.id}")
                except Exception as e:
                    logger.error(f"Erro ao registrar regra {name}: {str(e)}")
    
    def register_rule(self, rule: ValidationRule):
        """
        Registra uma regra de validação.
        
        Args:
            rule: Regra de validação
        """
        self.rules[rule.id] = rule
    
    def validate(self, 
                tenant_id: str, 
                region: Region, 
                industry: Industry, 
                frameworks: Optional[List[ComplianceFramework]] = None,
                config: Optional[Dict[str, Any]] = None) -> ValidationReport:
        """
        Executa validação de compliance.
        
        Args:
            tenant_id: ID do tenant
            region: Região
            industry: Indústria
            frameworks: Frameworks específicos para validar (opcional)
            config: Configuração do sistema IAM
            
        Returns:
            Relatório de validação
        """
        if frameworks is None:
            # Se não forem especificados frameworks, usar todos os aplicáveis à região e indústria
            requirements = []
            requirements.extend(self.registry.get_requirements_by_region(region))
            requirements.extend(self.registry.get_requirements_by_industry(industry))
            frameworks = list({req.framework for req in requirements})
        
        if config is None:
            config = {}
        
        # Obter benchmarks atualizados
        benchmarks = self.registry.get_benchmark_data()
        
        # Criar contexto de validação
        context = ValidationContext(
            tenant_id=tenant_id,
            region=region,
            industry=industry,
            frameworks=frameworks,
            config=config,
            benchmarks=benchmarks
        )
        
        # Determinar quais regras aplicar
        applicable_rules = self._get_applicable_rules(context)
        
        # Executar validação
        results = []
        for rule in applicable_rules:
            try:
                rule_results = rule.validate(context)
                results.extend(rule_results)
            except Exception as e:
                logger.error(f"Erro ao executar regra {rule.id}: {str(e)}")
                # Adicionar resultado de erro
                results.append(
                    ComplianceValidationResult(
                        rule_id=rule.id,
                        status=ValidationStatus.ERROR,
                        severity=ValidationSeverity.HIGH,
                        message=f"Erro ao executar regra: {str(e)}",
                        details={"error": str(e)},
                        timestamp=datetime.datetime.now()
                    )
                )
        
        # Calcular pontuação de compliance
        score = self._calculate_compliance_score(results, benchmarks, region, industry)
        
        # Criar relatório
        report = ValidationReport(
            tenant_id=tenant_id,
            region=region.value,
            industry=industry.value,
            frameworks=[f.value for f in frameworks],
            results=results,
            score=score,
            timestamp=datetime.datetime.now(),
            status=self._determine_overall_status(results, score)
        )
        
        return report
    
    def _get_applicable_rules(self, context: ValidationContext) -> List[ValidationRule]:
        """
        Determina quais regras são aplicáveis ao contexto.
        
        Args:
            context: Contexto de validação
            
        Returns:
            Lista de regras aplicáveis
        """
        applicable_rules = []
        
        for rule in self.rules.values():
            # Verificar se a regra é aplicável à região
            region_applicable = context.region in rule.get_applicable_regions() or Region.GLOBAL in rule.get_applicable_regions()
            
            # Verificar se a regra é aplicável à indústria
            industry_applicable = context.industry in rule.get_applicable_industries() or Industry.GENERAL in rule.get_applicable_industries()
            
            # Verificar se a regra é aplicável aos frameworks
            framework_applicable = False
            rule_frameworks = rule.get_applicable_frameworks()
            
            # Se a regra não especificar frameworks, ela é aplicável a todos
            if not rule_frameworks:
                framework_applicable = True
            else:
                # Verificar se algum dos frameworks solicitados é suportado pela regra
                for framework in context.frameworks:
                    if framework in rule_frameworks:
                        framework_applicable = True
                        break
            
            if region_applicable and industry_applicable and framework_applicable:
                applicable_rules.append(rule)
        
        return applicable_rules
    
    def _calculate_compliance_score(self, 
                                   results: List[ComplianceValidationResult],
                                   benchmarks: Dict[str, Any],
                                   region: Region,
                                   industry: Industry) -> float:
        """
        Calcula a pontuação de compliance com base nos resultados.
        
        Args:
            results: Resultados da validação
            benchmarks: Dados de benchmark
            region: Região
            industry: Indústria
            
        Returns:
            Pontuação de compliance (0-100)
        """
        if not results:
            return 0.0
        
        # Contadores para cálculo da pontuação
        total_weight = 0
        weighted_sum = 0
        
        for result in results:
            # Definir peso com base na severidade
            if result.severity == ValidationSeverity.CRITICAL:
                weight = 4.0
            elif result.severity == ValidationSeverity.HIGH:
                weight = 2.0
            elif result.severity == ValidationSeverity.MEDIUM:
                weight = 1.0
            else:  # LOW
                weight = 0.5
            
            # Pontuação baseada no status
            if result.status == ValidationStatus.PASS:
                score = 1.0
            elif result.status == ValidationStatus.WARNING:
                score = 0.5
            elif result.status == ValidationStatus.FAIL:
                score = 0.0
            else:  # ERROR
                score = 0.0
                
            weighted_sum += weight * score
            total_weight += weight
        
        # Calcular pontuação final (0-100)
        if total_weight > 0:
            raw_score = (weighted_sum / total_weight) * 100
        else:
            raw_score = 0
            
        # Ajustar a pontuação com base nos benchmarks
        industry_benchmark = benchmarks.get("industry", {}).get(industry.value.lower(), 70)
        
        # Se a pontuação estiver acima do benchmark da indústria, aumentar um pouco
        if raw_score > industry_benchmark:
            adjusted_score = raw_score * 1.05
        else:
            adjusted_score = raw_score
            
        # Limitear entre 0 e 100
        return min(max(adjusted_score, 0), 100)
    
    def _determine_overall_status(self, 
                                results: List[ComplianceValidationResult], 
                                score: float) -> ValidationStatus:
        """
        Determina o status geral com base nos resultados e pontuação.
        
        Args:
            results: Resultados da validação
            score: Pontuação de compliance
            
        Returns:
            Status geral
        """
        # Verificar se há falhas críticas
        for result in results:
            if result.status == ValidationStatus.FAIL and result.severity == ValidationSeverity.CRITICAL:
                return ValidationStatus.FAIL
        
        # Determinar status com base na pontuação
        if score >= 90:
            return ValidationStatus.PASS
        elif score >= 70:
            return ValidationStatus.WARNING
        else:
            return ValidationStatus.FAIL
    
    def generate_certificate(self, 
                           report: ValidationReport,
                           certificate_type: str = "standard",
                           options: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Gera um certificado de compliance com base no relatório.
        
        Args:
            report: Relatório de validação
            certificate_type: Tipo de certificado
            options: Opções de configuração do certificado
            
        Returns:
            Dados do certificado
        """
        # Na implementação real, isso chamaria o CertificateGenerator implementado anteriormente
        if options is None:
            options = {}
            
        # Criar dados básicos do certificado
        certificate = {
            "id": f"cert-{report.tenant_id}-{datetime.datetime.now().strftime('%Y%m%d%H%M%S')}",
            "tenant_id": report.tenant_id,
            "region": report.region,
            "industry": report.industry,
            "frameworks": report.frameworks,
            "score": report.score,
            "status": report.status.value,
            "issued_at": datetime.datetime.now().isoformat(),
            "valid_until": (datetime.datetime.now() + datetime.timedelta(days=365)).isoformat(),
            "type": certificate_type,
            "metadata": {
                "version": "1.0",
                "issuer": "INNOVABIZ IAM Compliance Engine",
                "verification_url": f"https://innovabiz.com/verify-cert/{report.tenant_id}",
                "options": options
            }
        }
        
        # Adicionar detalhes específicos do framework
        certificate["compliance_details"] = {}
        for framework in report.frameworks:
            # Filtrar resultados para este framework
            framework_results = [r for r in report.results if r.metadata.get("framework") == framework]
            
            # Calcular pontuação específica do framework
            framework_score = self._calculate_compliance_score(
                framework_results,
                self.registry.get_benchmark_data(),
                Region(report.region),
                Industry(report.industry)
            )
            
            # Adicionar detalhes ao certificado
            certificate["compliance_details"][framework] = {
                "score": framework_score,
                "status": self._determine_overall_status(framework_results, framework_score).value,
                "requirements_count": len(framework_results),
                "pass_count": len([r for r in framework_results if r.status == ValidationStatus.PASS]),
                "warning_count": len([r for r in framework_results if r.status == ValidationStatus.WARNING]),
                "fail_count": len([r for r in framework_results if r.status == ValidationStatus.FAIL])
            }
        
        # Adicionar benchmark comparisons
        benchmarks = self.registry.get_benchmark_data()
        certificate["benchmark_comparison"] = {
            "industry_average": benchmarks.get("industry", {}).get(report.industry.lower(), 70),
            "gartner": {
                "average": benchmarks.get("gartner", {}).get("average_score", 65),
                "leader": benchmarks.get("gartner", {}).get("leader_score", 85),
                "innovator": benchmarks.get("gartner", {}).get("innovator_score", 90)
            },
            "forrester": {
                "average": benchmarks.get("forrester", {}).get("average_score", 70),
                "leader": benchmarks.get("forrester", {}).get("leader_score", 88)
            }
        }
        
        return certificate


# Singleton para acesso global ao motor
_engine_instance = None

def get_compliance_engine(data_path: Optional[Path] = None) -> ComplianceEngine:
    """
    Obtém a instância singleton do motor de validação de compliance.
    
    Args:
        data_path: Caminho opcional para dados de compliance
        
    Returns:
        Instância do motor de validação de compliance
    """
    global _engine_instance
    if _engine_instance is None:
        _engine_instance = ComplianceEngine(data_path)
    return _engine_instance
