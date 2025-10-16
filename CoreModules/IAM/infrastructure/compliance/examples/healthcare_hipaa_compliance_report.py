"""
INNOVABIZ - Exemplo de Geração de Relatório de Compliance HIPAA para Saúde
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Script de exemplo que demonstra como gerar relatórios de compliance
           HIPAA para o módulo Healthcare, integrado com o sistema IAM.
==================================================================
"""

import json
import uuid
import logging
import argparse
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any

from jinja2 import Environment, FileSystemLoader
import pandas as pd
from weasyprint import HTML

# Importações do sistema de validação
from ...iam.compliance.validator import (
    ComplianceFramework, 
    RegionCode, 
    ComplianceLevel,
    ComplianceValidatorFactory,
    MultiRegionComplianceValidator
)

# Configuração de logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("innovabiz.examples.healthcare_hipaa")

# Diretórios para templates e outputs
SCRIPT_DIR = Path(__file__).parent
TEMPLATE_DIR = SCRIPT_DIR.parent / "templates"
OUTPUT_DIR = SCRIPT_DIR / "output"
OUTPUT_DIR.mkdir(exist_ok=True)


class HealthcareComplianceReportGenerator:
    """Gerador de relatórios de compliance para o módulo Healthcare"""
    
    def __init__(self, tenant_id: uuid.UUID, output_dir: Path = OUTPUT_DIR):
        """Inicializa o gerador de relatórios"""
        self.tenant_id = tenant_id
        self.output_dir = output_dir
        self.jinja_env = Environment(loader=FileSystemLoader(TEMPLATE_DIR))
        self.validator = MultiRegionComplianceValidator(tenant_id)
        
        logger.info(f"Inicializado gerador de relatórios para tenant {tenant_id}")
    
    def load_config(self, config_path: Optional[str] = None) -> Dict[str, Any]:
        """
        Carrega configuração de IAM a partir de um arquivo ou usa modelo de exemplo
        Args:
            config_path: Caminho para arquivo de configuração JSON (opcional)
        Returns:
            Configuração como dicionário
        """
        if config_path and Path(config_path).exists():
            logger.info(f"Carregando configuração de {config_path}")
            with open(config_path, 'r') as f:
                return json.load(f)
        
        logger.info("Usando configuração de exemplo")
        return self._get_example_config()
    
    def _get_example_config(self) -> Dict[str, Any]:
        """Retorna uma configuração de exemplo para testes"""
        return {
            "tenant_id": str(self.tenant_id),
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
                    "emergency_access": True,
                    "phi_data_classification": {
                        "enabled": True,
                        "auto_classification": True
                    }
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
    
    def generate_report(self, config: Dict[str, Any], language: str = "pt", formats: List[str] = ["html", "pdf", "json", "csv"]) -> Dict[str, Path]:
        """
        Gera relatório de compliance para múltiplas regiões
        Args:
            config: Configuração do IAM como dicionário
            language: Idioma para o relatório ('pt' ou 'en')
            formats: Lista de formatos para gerar (html, pdf, json, csv)
        Returns:
            Dicionário com caminhos para os arquivos gerados
        """
        logger.info(f"Gerando relatório de compliance em {language} nos formatos {formats}")
        
        # Validar configuração para todas as regiões
        all_results = self.validator.validate_all_regions(config)
        
        # Extrair apenas os resultados HIPAA para análise específica de healthcare
        us_results = all_results.get(RegionCode.US, {})
        hipaa_results = us_results.get(ComplianceFramework.HIPAA, [])
        
        # Gerar relatório completo multi-framework
        report = self.validator.generate_compliance_report(all_results, language)
        
        # Adicionar timestamp ao relatório
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename_base = f"healthcare_compliance_report_{language}_{timestamp}"
        output_files = {}
        
        # Compilar informações específicas sobre HIPAA para healthcare
        hipaa_healthcare_stats = self._compile_hipaa_healthcare_stats(hipaa_results)
        ar_factor_usage = self._compile_ar_authentication_stats(config, hipaa_results)
        
        # Adicionar informações ao relatório
        report["healthcare_specific"] = {
            "hipaa_stats": hipaa_healthcare_stats,
            "ar_factors": ar_factor_usage,
            "has_healthcare_module": "healthcare" in config.get("modules", {}) and config["modules"]["healthcare"].get("enabled", False)
        }
        
        # Gerar saídas nos formatos solicitados
        if "json" in formats:
            # Exportar relatório completo como JSON
            json_path = self.output_dir / f"{filename_base}.json"
            with open(json_path, "w") as f:
                json.dump(report, f, indent=2)
            output_files["json"] = json_path
            logger.info(f"Relatório JSON gerado: {json_path}")
        
        if "csv" in formats:
            # Exportar requisitos como CSV para análise
            csv_path = self.output_dir / f"{filename_base}_requirements.csv"
            self._export_requirements_csv(report, csv_path)
            output_files["csv"] = csv_path
            logger.info(f"Relatório CSV gerado: {csv_path}")
        
        if "html" in formats or "pdf" in formats:
            # Definir template com base no idioma
            template_name = f"compliance_report_{'en' if language == 'en' else 'pt'}.html"
            template = self.jinja_env.get_template(template_name)
            
            # Renderizar HTML
            html_content = template.render(
                report=report,
                now=datetime.now(),
                healthcare_specific=report["healthcare_specific"]
            )
            
            if "html" in formats:
                html_path = self.output_dir / f"{filename_base}.html"
                with open(html_path, "w", encoding="utf-8") as f:
                    f.write(html_content)
                output_files["html"] = html_path
                logger.info(f"Relatório HTML gerado: {html_path}")
            
            if "pdf" in formats:
                pdf_path = self.output_dir / f"{filename_base}.pdf"
                HTML(string=html_content).write_pdf(pdf_path)
                output_files["pdf"] = pdf_path
                logger.info(f"Relatório PDF gerado: {pdf_path}")
        
        return output_files
    
    def _compile_hipaa_healthcare_stats(self, hipaa_results: List) -> Dict[str, Any]:
        """Compila estatísticas de compliance HIPAA específicas para healthcare"""
        categories = {}
        total_compliant = 0
        total_partially = 0
        total_non_compliant = 0
        
        for result in hipaa_results:
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
                total_compliant += 1
            elif result.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT:
                categories[category]["partially_compliant"] += 1
                total_partially += 1
            elif result.compliance_level == ComplianceLevel.NON_COMPLIANT:
                categories[category]["non_compliant"] += 1
                total_non_compliant += 1
        
        total_requirements = total_compliant + total_partially + total_non_compliant
        
        return {
            "categories": categories,
            "total_requirements": total_requirements,
            "total_compliant": total_compliant,
            "total_partially_compliant": total_partially,
            "total_non_compliant": total_non_compliant,
            "overall_score": round(((total_compliant + (total_partially * 0.5)) / total_requirements) * 100, 2) if total_requirements > 0 else 0
        }
    
    def _compile_ar_authentication_stats(self, config: Dict[str, Any], hipaa_results: List) -> Dict[str, Any]:
        """Compila estatísticas de uso de fatores AR para autenticação"""
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
            "factors": ar_factors,
            "factor_count": ar_factor_count,
            "enabled_factors": [factor for factor, enabled in ar_factors.items() if enabled],
            "ar_auth_enabled": adaptive_auth.get("enabled", False) and ar_auth.get("enabled", False),
            "enhances_phi_security": ar_factor_count >= 2 and adaptive_auth.get("enabled", False) and ar_auth.get("enabled", False)
        }
    
    def _export_requirements_csv(self, report: Dict[str, Any], csv_path: Path) -> None:
        """Exporta os requisitos de compliance como CSV para análise"""
        rows = []
        
        # Extrair dados HIPAA
        for region_code, region_data in report["regions"].items():
            for framework_key, framework_data in region_data["frameworks"].items():
                if framework_key == ComplianceFramework.HIPAA.value:
                    # Adicionar estatísticas gerais
                    rows.append({
                        "Region": region_code,
                        "Framework": framework_key,
                        "Requirement": "OVERALL",
                        "Status": framework_data.get("status", ""),
                        "Score": framework_data.get("score", 0),
                        "Category": "summary",
                        "Details": "Overall framework score"
                    })
                    
                    # Adicionar detalhes por requisito
                    for req in framework_data.get("requirements_details", []):
                        rows.append({
                            "Region": region_code,
                            "Framework": framework_key,
                            "Requirement": req.get("req_id", ""),
                            "Status": req.get("status", ""),
                            "Score": 100 if req.get("status") == "compliant" else (50 if req.get("status") == "partially_compliant" else 0),
                            "Category": req.get("category", ""),
                            "Details": req.get("details", "")
                        })
        
        # Converter para DataFrame e exportar
        if rows:
            df = pd.DataFrame(rows)
            df.to_csv(csv_path, index=False)


def main():
    """Função principal para executar o gerador de relatórios"""
    parser = argparse.ArgumentParser(description="Gerador de Relatórios de Compliance HIPAA para Healthcare")
    parser.add_argument("--config", help="Caminho para arquivo de configuração (opcional)")
    parser.add_argument("--language", choices=["pt", "en"], default="pt", help="Idioma para o relatório")
    parser.add_argument("--formats", nargs="+", choices=["html", "pdf", "json", "csv"], default=["html", "json"], 
                        help="Formatos para os relatórios")
    parser.add_argument("--output-dir", help="Diretório para saída (opcional)")
    
    args = parser.parse_args()
    
    # Configurar diretório de saída
    output_dir = Path(args.output_dir) if args.output_dir else OUTPUT_DIR
    output_dir.mkdir(exist_ok=True)
    
    # Gerar UUID para tenant de exemplo
    tenant_id = uuid.uuid4()
    
    # Criar gerador de relatórios
    generator = HealthcareComplianceReportGenerator(tenant_id, output_dir)
    
    # Carregar configuração
    config = generator.load_config(args.config)
    
    # Gerar relatório
    output_files = generator.generate_report(config, args.language, args.formats)
    
    logger.info(f"Geração de relatórios concluída com sucesso. Arquivos gerados:")
    for format_name, file_path in output_files.items():
        logger.info(f"- {format_name.upper()}: {file_path}")


if __name__ == "__main__":
    main()
