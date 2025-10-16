"""
INNOVABIZ - Gerador de Relatórios de Compliance IAM
==================================================================
Autor: Eduardo Jeremias
Data: 06/05/2025
Versão: 1.0
Descrição: Sistema de geração de relatórios de compliance para o módulo IAM
           com suporte a múltiplos formatos e integrações.
==================================================================
"""

import base64
import csv
import json
import logging
import os
import uuid
from datetime import datetime
from io import StringIO
from typing import Dict, List, Optional, Union, Any, Tuple

import jinja2
import markdown
import pdfkit
import yaml

from .validator import (
    ComplianceLevel,
    ComplianceFramework,
    RegionCode,
    ComplianceValidationResult,
    MultiRegionComplianceValidator
)

# Configuração de logger
logger = logging.getLogger("innovabiz.iam.compliance.reports")


class ComplianceReportGenerator:
    """Gerador de relatórios de compliance para resultados de validação"""
    
    def __init__(self, tenant_id: uuid.UUID, output_dir: str = None):
        """
        Inicializa o gerador de relatórios
        
        Args:
            tenant_id: ID do tenant
            output_dir: Diretório opcional para saída de relatórios
        """
        self.tenant_id = tenant_id
        self.output_dir = output_dir or os.path.join(os.getcwd(), "reports")
        self.template_dir = os.path.join(os.path.dirname(__file__), "templates")
        self.jinja_env = jinja2.Environment(
            loader=jinja2.FileSystemLoader(self.template_dir),
            autoescape=jinja2.select_autoescape(['html', 'xml'])
        )
        
        # Garantir que o diretório de saída exista
        if not os.path.exists(self.output_dir):
            os.makedirs(self.output_dir)
            
        logger.info(f"Inicializado gerador de relatórios para tenant {tenant_id}")
    
    def generate_report(
        self,
        validation_results: Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]],
        format: str = "pdf",
        language: str = "pt",
        include_details: bool = True,
        include_remediation: bool = True
    ) -> str:
        """
        Gera um relatório de compliance no formato solicitado
        
        Args:
            validation_results: Resultados da validação por região e framework
            format: Formato do relatório (pdf, html, json, csv, markdown)
            language: Idioma do relatório (pt, en)
            include_details: Incluir detalhes de validação
            include_remediation: Incluir recomendações de remediação
            
        Returns:
            Caminho do arquivo de relatório gerado
        """
        # Processar resultados para formato adequado ao relatório
        processed_data = self._process_validation_results(
            validation_results,
            language,
            include_details,
            include_remediation
        )
        
        # Gerar relatório no formato solicitado
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        report_filename = f"compliance_report_{self.tenant_id}_{timestamp}"
        
        if format.lower() == "pdf":
            return self._generate_pdf_report(processed_data, report_filename, language)
        elif format.lower() == "html":
            return self._generate_html_report(processed_data, report_filename, language)
        elif format.lower() == "json":
            return self._generate_json_report(processed_data, report_filename)
        elif format.lower() == "csv":
            return self._generate_csv_report(processed_data, report_filename)
        elif format.lower() == "markdown":
            return self._generate_markdown_report(processed_data, report_filename, language)
        else:
            raise ValueError(f"Formato de relatório não suportado: {format}")
    
    def _process_validation_results(
        self,
        validation_results: Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]],
        language: str,
        include_details: bool,
        include_remediation: bool
    ) -> Dict:
        """
        Processa os resultados da validação em um formato adequado para relatórios
        
        Args:
            validation_results: Resultados da validação por região e framework
            language: Idioma do relatório
            include_details: Incluir detalhes de validação
            include_remediation: Incluir recomendações de remediação
            
        Returns:
            Dados processados para o relatório
        """
        processed_data = {
            "report_id": str(uuid.uuid4()),
            "tenant_id": str(self.tenant_id),
            "timestamp": datetime.now().isoformat(),
            "language": language,
            "overall_compliance": {
                "status": None,
                "score": 0.0,
                "summary": {},
            },
            "regions": {},
            "frameworks": {},
            "issues": [],
            "strengths": []
        }
        
        # Calcular estatísticas globais
        total_requirements = 0
        total_compliant = 0
        total_partially = 0
        total_non_compliant = 0
        total_na = 0
        
        # Processar resultados por região
        for region, frameworks in validation_results.items():
            region_data = {
                "frameworks": {},
                "total_requirements": 0,
                "compliant": 0,
                "partially_compliant": 0,
                "non_compliant": 0,
                "score": 0.0
            }
            
            # Processar resultados por framework
            for framework, results in frameworks.items():
                if not results:
                    continue
                
                compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.COMPLIANT)
                partially = sum(1 for r in results if r.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT)
                non_compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.NON_COMPLIANT)
                not_applicable = sum(1 for r in results if r.compliance_level == ComplianceLevel.NOT_APPLICABLE)
                
                # Atualizar totais
                total_requirements += len(results) - not_applicable
                total_compliant += compliant
                total_partially += partially
                total_non_compliant += non_compliant
                total_na += not_applicable
                
                # Atualizar totais da região
                region_data["total_requirements"] += len(results) - not_applicable
                region_data["compliant"] += compliant
                region_data["partially_compliant"] += partially
                region_data["non_compliant"] += non_compliant
                
                # Calcular pontuação para o framework
                framework_score = 0.0
                if (len(results) - not_applicable) > 0:
                    framework_score = (compliant + (partially * 0.5)) / (len(results) - not_applicable) * 100
                
                framework_data = {
                    "total_requirements": len(results) - not_applicable,
                    "compliant": compliant,
                    "partially_compliant": partially,
                    "non_compliant": non_compliant,
                    "not_applicable": not_applicable,
                    "score": round(framework_score, 1)
                }
                
                region_data["frameworks"][framework.value] = framework_data
                
                # Adicionar ao dicionário global de frameworks
                if framework.value not in processed_data["frameworks"]:
                    processed_data["frameworks"][framework.value] = {
                        "total_requirements": 0,
                        "compliant": 0,
                        "partially_compliant": 0,
                        "non_compliant": 0,
                        "not_applicable": 0,
                        "score": 0.0,
                        "regions": []
                    }
                
                processed_data["frameworks"][framework.value]["total_requirements"] += len(results) - not_applicable
                processed_data["frameworks"][framework.value]["compliant"] += compliant
                processed_data["frameworks"][framework.value]["partially_compliant"] += partially
                processed_data["frameworks"][framework.value]["non_compliant"] += non_compliant
                processed_data["frameworks"][framework.value]["not_applicable"] += not_applicable
                
                if region.value not in processed_data["frameworks"][framework.value]["regions"]:
                    processed_data["frameworks"][framework.value]["regions"].append(region.value)
                
                # Coletar problemas e pontos fortes
                for result in results:
                    requirement_data = {
                        "region": region.value,
                        "framework": framework.value,
                        "requirement_id": result.requirement.req_id,
                        "category": result.requirement.category,
                        "severity": result.requirement.severity,
                        "description": result.requirement.description_pt if language == "pt" else result.requirement.description,
                    }
                    
                    if include_details:
                        requirement_data["details"] = result.details_pt if language == "pt" else result.details
                    
                    if include_remediation and result.remediation:
                        requirement_data["remediation"] = result.remediation_pt if language == "pt" else result.remediation
                    
                    if result.compliance_level == ComplianceLevel.NON_COMPLIANT:
                        processed_data["issues"].append(requirement_data)
                    elif result.compliance_level == ComplianceLevel.COMPLIANT:
                        processed_data["strengths"].append(requirement_data)
            
            # Calcular pontuação para a região
            if region_data["total_requirements"] > 0:
                region_score = (region_data["compliant"] + (region_data["partially_compliant"] * 0.5)) / region_data["total_requirements"] * 100
                region_data["score"] = round(region_score, 1)
            
            processed_data["regions"][region.value] = region_data
        
        # Calcular pontuação global
        if total_requirements > 0:
            global_score = (total_compliant + (total_partially * 0.5)) / total_requirements * 100
            processed_data["overall_compliance"]["score"] = round(global_score, 1)
            
            if global_score >= 90:
                processed_data["overall_compliance"]["status"] = "high_compliance"
            elif global_score >= 75:
                processed_data["overall_compliance"]["status"] = "moderate_compliance"
            else:
                processed_data["overall_compliance"]["status"] = "low_compliance"
        
        # Calcular pontuações para frameworks
        for fw_key, fw_data in processed_data["frameworks"].items():
            if fw_data["total_requirements"] > 0:
                fw_score = (fw_data["compliant"] + (fw_data["partially_compliant"] * 0.5)) / fw_data["total_requirements"] * 100
                fw_data["score"] = round(fw_score, 1)
        
        # Gerar resumo global
        processed_data["overall_compliance"]["summary"] = {
            "total_requirements": total_requirements,
            "compliant": total_compliant,
            "partially_compliant": total_partially,
            "non_compliant": total_non_compliant,
            "not_applicable": total_na,
            "regions_count": len(validation_results),
            "frameworks_count": len(set(fw for region in validation_results.values() for fw in region.keys())),
            "critical_issues": sum(1 for issue in processed_data["issues"] if issue["severity"] == "high")
        }
        
        # Ordenar problemas por severidade
        processed_data["issues"].sort(key=lambda x: 0 if x["severity"] == "high" else (1 if x["severity"] == "medium" else 2))
        
        return processed_data
    
    def _generate_pdf_report(self, data: Dict, filename: str, language: str) -> str:
        """
        Gera relatório em formato PDF
        
        Args:
            data: Dados processados do relatório
            filename: Nome do arquivo base
            language: Idioma do relatório
            
        Returns:
            Caminho para o arquivo PDF gerado
        """
        # Primeiro geramos o HTML e depois convertemos para PDF
        html_content = self._render_html_template(data, language)
        pdf_path = os.path.join(self.output_dir, f"{filename}.pdf")
        
        try:
            pdfkit.from_string(html_content, pdf_path)
            logger.info(f"Relatório PDF gerado: {pdf_path}")
            return pdf_path
        except Exception as e:
            logger.error(f"Erro ao gerar PDF: {str(e)}")
            # Fallback para HTML se PDF falhar
            return self._generate_html_report(data, filename, language)
    
    def _generate_html_report(self, data: Dict, filename: str, language: str) -> str:
        """
        Gera relatório em formato HTML
        
        Args:
            data: Dados processados do relatório
            filename: Nome do arquivo base
            language: Idioma do relatório
            
        Returns:
            Caminho para o arquivo HTML gerado
        """
        html_content = self._render_html_template(data, language)
        html_path = os.path.join(self.output_dir, f"{filename}.html")
        
        with open(html_path, 'w', encoding='utf-8') as f:
            f.write(html_content)
        
        logger.info(f"Relatório HTML gerado: {html_path}")
        return html_path
    
    def _render_html_template(self, data: Dict, language: str) -> str:
        """
        Renderiza o template HTML para o relatório
        
        Args:
            data: Dados do relatório
            language: Idioma do relatório
            
        Returns:
            Conteúdo HTML renderizado
        """
        template_name = f"compliance_report_{'pt' if language == 'pt' else 'en'}.html"
        try:
            template = self.jinja_env.get_template(template_name)
            return template.render(report=data, now=datetime.now())
        except jinja2.exceptions.TemplateNotFound:
            # Usar template padrão se o específico do idioma não for encontrado
            logger.warning(f"Template {template_name} não encontrado, usando padrão")
            template = self.jinja_env.get_template("compliance_report_default.html")
            return template.render(report=data, now=datetime.now())
    
    def _generate_json_report(self, data: Dict, filename: str) -> str:
        """
        Gera relatório em formato JSON
        
        Args:
            data: Dados processados do relatório
            filename: Nome do arquivo base
            
        Returns:
            Caminho para o arquivo JSON gerado
        """
        json_path = os.path.join(self.output_dir, f"{filename}.json")
        
        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        
        logger.info(f"Relatório JSON gerado: {json_path}")
        return json_path
    
    def _generate_csv_report(self, data: Dict, filename: str) -> str:
        """
        Gera relatório em formato CSV
        
        Args:
            data: Dados processados do relatório
            filename: Nome do arquivo base
            
        Returns:
            Caminho para o arquivo CSV gerado
        """
        csv_path = os.path.join(self.output_dir, f"{filename}.csv")
        
        with open(csv_path, 'w', encoding='utf-8', newline='') as f:
            writer = csv.writer(f)
            
            # Cabeçalho
            writer.writerow([
                "Region", "Framework", "Requirement ID", "Category", 
                "Severity", "Description", "Status", "Details", "Remediation"
            ])
            
            # Dados de todos os requisitos
            for region_code, region_data in data["regions"].items():
                for framework, framework_data in region_data["frameworks"].items():
                    # Precisamos recuperar os resultados originais, pois o processamento simplificou
                    for issue in data["issues"]:
                        if issue["region"] == region_code and issue["framework"] == framework:
                            writer.writerow([
                                region_code,
                                framework,
                                issue["requirement_id"],
                                issue["category"],
                                issue["severity"],
                                issue["description"],
                                "NON_COMPLIANT",
                                issue.get("details", ""),
                                issue.get("remediation", "")
                            ])
                    
                    for strength in data["strengths"]:
                        if strength["region"] == region_code and strength["framework"] == framework:
                            writer.writerow([
                                region_code,
                                framework,
                                strength["requirement_id"],
                                strength["category"],
                                strength["severity"],
                                strength["description"],
                                "COMPLIANT",
                                strength.get("details", ""),
                                ""
                            ])
        
        logger.info(f"Relatório CSV gerado: {csv_path}")
        return csv_path
    
    def _generate_markdown_report(self, data: Dict, filename: str, language: str) -> str:
        """
        Gera relatório em formato Markdown
        
        Args:
            data: Dados processados do relatório
            filename: Nome do arquivo base
            language: Idioma do relatório
            
        Returns:
            Caminho para o arquivo Markdown gerado
        """
        md_path = os.path.join(self.output_dir, f"{filename}.md")
        
        # Strings localizadas
        title = "Relatório de Compliance IAM" if language == "pt" else "IAM Compliance Report"
        summary = "Resumo" if language == "pt" else "Summary"
        overall_score = "Pontuação Global" if language == "pt" else "Overall Score"
        region_details = "Detalhes por Região" if language == "pt" else "Region Details"
        framework_details = "Detalhes por Framework" if language == "pt" else "Framework Details"
        issues = "Problemas Identificados" if language == "pt" else "Identified Issues"
        strengths = "Pontos Fortes" if language == "pt" else "Strengths"
        
        with open(md_path, 'w', encoding='utf-8') as f:
            # Cabeçalho
            f.write(f"# {title}\n\n")
            f.write(f"**ID do Relatório:** {data['report_id']}\n")
            f.write(f"**Tenant ID:** {data['tenant_id']}\n")
            f.write(f"**Data:** {datetime.now().strftime('%d/%m/%Y %H:%M:%S')}\n\n")
            
            # Resumo
            f.write(f"## {summary}\n\n")
            f.write(f"**{overall_score}:** {data['overall_compliance']['score']}%\n\n")
            f.write("| Categoria | Valor |\n")
            f.write("|-----------|-------|\n")
            for key, value in data["overall_compliance"]["summary"].items():
                f.write(f"| {key} | {value} |\n")
            f.write("\n")
            
            # Detalhes por região
            f.write(f"## {region_details}\n\n")
            for region_code, region_data in data["regions"].items():
                f.write(f"### {region_code.upper()}\n\n")
                f.write(f"**Score:** {region_data['score']}%\n\n")
                f.write("| Framework | Requisitos | Conformes | Parcialmente | Não Conformes | Score |\n")
                f.write("|-----------|------------|-----------|--------------|---------------|-------|\n")
                
                for fw, fw_data in region_data["frameworks"].items():
                    f.write(f"| {fw} | {fw_data['total_requirements']} | {fw_data['compliant']} | {fw_data['partially_compliant']} | {fw_data['non_compliant']} | {fw_data['score']}% |\n")
                f.write("\n")
            
            # Detalhes por framework
            f.write(f"## {framework_details}\n\n")
            for fw_key, fw_data in data["frameworks"].items():
                f.write(f"### {fw_key}\n\n")
                f.write(f"**Score:** {fw_data['score']}%\n\n")
                f.write("| Requisitos | Conformes | Parcialmente | Não Conformes | Regiões |\n")
                f.write("|------------|-----------|--------------|---------------|--------|\n")
                f.write(f"| {fw_data['total_requirements']} | {fw_data['compliant']} | {fw_data['partially_compliant']} | {fw_data['non_compliant']} | {', '.join(fw_data['regions'])} |\n\n")
            
            # Problemas
            if data["issues"]:
                f.write(f"## {issues}\n\n")
                for issue in data["issues"]:
                    f.write(f"### {issue['requirement_id']} ({issue['framework']}, {issue['region']})\n\n")
                    f.write(f"**{issue['description']}**\n\n")
                    f.write(f"**Categoria:** {issue['category']}\n")
                    f.write(f"**Severidade:** {issue['severity']}\n")
                    
                    if "details" in issue:
                        f.write(f"**Detalhes:** {issue['details']}\n\n")
                    
                    if "remediation" in issue:
                        f.write(f"**Remediação:** {issue['remediation']}\n\n")
                    
                    f.write("---\n\n")
            
            # Pontos fortes
            if data["strengths"]:
                f.write(f"## {strengths}\n\n")
                for strength in data["strengths"]:
                    f.write(f"### {strength['requirement_id']} ({strength['framework']}, {strength['region']})\n\n")
                    f.write(f"**{strength['description']}**\n\n")
                    
                    if "details" in strength:
                        f.write(f"**Detalhes:** {strength['details']}\n\n")
                    
                    f.write("---\n\n")
        
        logger.info(f"Relatório Markdown gerado: {md_path}")
        return md_path


class ComplianceAuditTrail:
    """Registro de trilha de auditoria para validações de compliance"""
    
    def __init__(self, tenant_id: uuid.UUID, storage_path: str = None):
        self.tenant_id = tenant_id
        self.storage_path = storage_path or os.path.join(os.getcwd(), "audit_trails")
        
        # Garantir que o diretório de armazenamento exista
        if not os.path.exists(self.storage_path):
            os.makedirs(self.storage_path)
            
        logger.info(f"Inicializado sistema de trilha de auditoria para tenant {tenant_id}")
    
    def record_validation(
        self,
        validation_results: Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]],
        user_id: Optional[str] = None,
        metadata: Optional[Dict] = None
    ) -> str:
        """
        Registra uma validação de compliance na trilha de auditoria
        
        Args:
            validation_results: Resultados da validação
            user_id: ID opcional do usuário que executou a validação
            metadata: Metadados adicionais
            
        Returns:
            ID do registro de auditoria
        """
        audit_id = str(uuid.uuid4())
        timestamp = datetime.now().isoformat()
        
        # Preparar registro de auditoria
        audit_record = {
            "audit_id": audit_id,
            "tenant_id": str(self.tenant_id),
            "timestamp": timestamp,
            "user_id": user_id,
            "metadata": metadata or {},
            "summary": self._generate_summary(validation_results),
            # Não armazenamos resultados completos aqui para economizar espaço
            # Em vez disso, armazenamos apenas o resumo
        }
        
        # Salvar registro
        audit_file = os.path.join(self.storage_path, f"audit_{audit_id}.json")
        with open(audit_file, 'w', encoding='utf-8') as f:
            json.dump(audit_record, f, indent=2, ensure_ascii=False)
        
        logger.info(f"Registro de validação criado: {audit_id}")
        return audit_id
    
    def _generate_summary(
        self,
        validation_results: Dict[RegionCode, Dict[ComplianceFramework, List[ComplianceValidationResult]]]
    ) -> Dict:
        """
        Gera um resumo dos resultados de validação
        
        Args:
            validation_results: Resultados da validação
            
        Returns:
            Resumo dos resultados
        """
        summary = {
            "regions": {},
            "frameworks": {},
            "overall": {
                "total_requirements": 0,
                "compliant": 0,
                "partially_compliant": 0,
                "non_compliant": 0,
                "not_applicable": 0
            }
        }
        
        # Processar resultados por região e framework
        for region, frameworks in validation_results.items():
            region_key = region.value
            summary["regions"][region_key] = {
                "frameworks": [],
                "total_requirements": 0,
                "compliant": 0,
                "partially_compliant": 0,
                "non_compliant": 0,
                "not_applicable": 0
            }
            
            for framework, results in frameworks.items():
                framework_key = framework.value
                
                # Adicionar framework à região
                if framework_key not in summary["regions"][region_key]["frameworks"]:
                    summary["regions"][region_key]["frameworks"].append(framework_key)
                
                # Adicionar região ao framework
                if framework_key not in summary["frameworks"]:
                    summary["frameworks"][framework_key] = {
                        "regions": [],
                        "total_requirements": 0,
                        "compliant": 0,
                        "partially_compliant": 0,
                        "non_compliant": 0,
                        "not_applicable": 0
                    }
                
                if region_key not in summary["frameworks"][framework_key]["regions"]:
                    summary["frameworks"][framework_key]["regions"].append(region_key)
                
                # Calcular estatísticas
                if not results:
                    continue
                    
                compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.COMPLIANT)
                partially = sum(1 for r in results if r.compliance_level == ComplianceLevel.PARTIALLY_COMPLIANT)
                non_compliant = sum(1 for r in results if r.compliance_level == ComplianceLevel.NON_COMPLIANT)
                not_applicable = sum(1 for r in results if r.compliance_level == ComplianceLevel.NOT_APPLICABLE)
                total = len(results)
                
                # Atualizar estatísticas da região
                summary["regions"][region_key]["total_requirements"] += total
                summary["regions"][region_key]["compliant"] += compliant
                summary["regions"][region_key]["partially_compliant"] += partially
                summary["regions"][region_key]["non_compliant"] += non_compliant
                summary["regions"][region_key]["not_applicable"] += not_applicable
                
                # Atualizar estatísticas do framework
                summary["frameworks"][framework_key]["total_requirements"] += total
                summary["frameworks"][framework_key]["compliant"] += compliant
                summary["frameworks"][framework_key]["partially_compliant"] += partially
                summary["frameworks"][framework_key]["non_compliant"] += non_compliant
                summary["frameworks"][framework_key]["not_applicable"] += not_applicable
                
                # Atualizar estatísticas globais
                summary["overall"]["total_requirements"] += total
                summary["overall"]["compliant"] += compliant
                summary["overall"]["partially_compliant"] += partially
                summary["overall"]["non_compliant"] += non_compliant
                summary["overall"]["not_applicable"] += not_applicable
        
        return summary
