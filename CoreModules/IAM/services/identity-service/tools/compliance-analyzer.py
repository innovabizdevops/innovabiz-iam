#!/usr/bin/env python3
"""
Analisador de Resultados de Compliance para INNOVABIZ IAM
--------------------------------------------------------

Este script analisa os resultados dos testes de compliance ao longo do tempo,
identificando tendências, correlações e áreas de risco.

Uso:
    python3 compliance-analyzer.py --reports-dir=./reports --region=AO --output=./analysis

Autor: INNOVABIZ DevSecOps
Data: 05/08/2025
"""

import argparse
import json
import os
import glob
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
from datetime import datetime
import re
from pathlib import Path


class ComplianceAnalyzer:
    def __init__(self, reports_dir, region=None, output_dir="./compliance-analysis"):
        """
        Inicializa o analisador de compliance.
        
        Args:
            reports_dir: Diretório contendo os relatórios de compliance JSON
            region: Região específica a analisar (opcional)
            output_dir: Diretório para saída dos resultados da análise
        """
        self.reports_dir = reports_dir
        self.region = region
        self.output_dir = output_dir
        self.reports_data = []
        self.df_scores = None
        self.df_frameworks = None
        self.df_tests = None
        
        # Garantir que o diretório de saída exista
        os.makedirs(output_dir, exist_ok=True)

    def load_reports(self):
        """Carrega todos os relatórios de compliance no diretório especificado"""
        print(f"Carregando relatórios de: {self.reports_dir}")
        
        # Padrão para arquivos de relatório
        pattern = "compliance_report_*.json"
        if self.region:
            pattern = f"compliance_report_{self.region}_*.json"
        
        # Procurar em todos os subdiretórios
        report_files = []
        for root, dirs, files in os.walk(self.reports_dir):
            for file in files:
                if re.match(f"compliance_report_.*\.json", file):
                    if self.region and not file.startswith(f"compliance_report_{self.region}_"):
                        continue
                    report_files.append(os.path.join(root, file))
        
        print(f"Encontrados {len(report_files)} relatórios.")
        
        # Processar cada relatório
        for report_file in report_files:
            try:
                with open(report_file, 'r') as f:
                    report_data = json.load(f)
                    
                # Extrair data do nome do arquivo ou do próprio relatório
                timestamp = None
                file_name = os.path.basename(report_file)
                date_match = re.search(r'(\d{8}_\d{6})', file_name)
                if date_match:
                    date_str = date_match.group(1)
                    timestamp = datetime.strptime(date_str, "%Y%m%d_%H%M%S")
                elif "executedAt" in report_data:
                    timestamp = datetime.fromisoformat(report_data["executedAt"].replace("Z", "+00:00"))
                else:
                    # Usar a data de modificação do arquivo como fallback
                    timestamp = datetime.fromtimestamp(os.path.getmtime(report_file))
                
                report_data["timestamp"] = timestamp
                report_data["file"] = report_file
                self.reports_data.append(report_data)
                
            except Exception as e:
                print(f"Erro ao processar {report_file}: {str(e)}")
                
        # Ordenar por data
        self.reports_data.sort(key=lambda x: x["timestamp"])
        
        print(f"Carregados {len(self.reports_data)} relatórios válidos.")
        return len(self.reports_data) > 0
        
    def prepare_dataframes(self):
        """Prepara DataFrames pandas para análise"""
        if not self.reports_data:
            print("Nenhum dado de relatório para analisar.")
            return False
            
        # DataFrame principal de pontuações gerais
        scores_data = []
        for report in self.reports_data:
            scores_data.append({
                "timestamp": report["timestamp"],
                "region": report["region"],
                "regionName": report.get("regionName", report["region"]),
                "complianceScore": report["complianceScore"],
                "totalTests": report["totalTests"],
                "passedTests": report["passedTests"],
                "failedTests": report["failedTests"],
                "duration": report.get("duration", 0)
            })
        
        self.df_scores = pd.DataFrame(scores_data)
        
        # DataFrame de frameworks
        frameworks_data = []
        for report in self.reports_data:
            for fw_id, fw_data in report.get("frameworkScores", {}).items():
                frameworks_data.append({
                    "timestamp": report["timestamp"],
                    "region": report["region"],
                    "frameworkId": fw_id,
                    "frameworkName": fw_data.get("name", fw_id),
                    "complianceScore": fw_data.get("complianceScore", 0),
                    "totalTests": fw_data.get("totalTests", 0),
                    "passedTests": fw_data.get("passedTests", 0),
                    "failedTests": fw_data.get("failedTests", 0)
                })
        
        self.df_frameworks = pd.DataFrame(frameworks_data)
        
        # DataFrame de testes individuais
        tests_data = []
        for report in self.reports_data:
            for test_result in report.get("testResults", []):
                test_case = test_result.get("testCase", {})
                
                # Alguns campos básicos comuns
                test_entry = {
                    "timestamp": report["timestamp"],
                    "region": report["region"],
                    "testId": test_case.get("id", "unknown"),
                    "testName": test_case.get("name", "Unknown"),
                    "policyPath": test_result.get("policyPath", ""),
                    "passed": test_result.get("passed", False),
                    "criticality": test_result.get("criticality", "baixa"),
                    "executionTimeMs": test_result.get("executionTimeMs", 0),
                    "requirements": ",".join(test_result.get("requirements", [])),
                    "frameworks": ",".join(test_result.get("frameworks", []))
                }
                
                # Adicionar violações se existirem
                if not test_result.get("passed", True) and test_result.get("violations"):
                    test_entry["violations"] = ",".join(test_result.get("violations", []))
                
                tests_data.append(test_entry)
        
        self.df_tests = pd.DataFrame(tests_data)
        
        print("DataFrames preparados para análise.")
        return True
        
    def analyze(self):
        """Executa a análise completa dos dados de compliance"""
        print("Iniciando análise de dados de compliance...")
        
        if not self.prepare_dataframes():
            return False
            
        # Análises a realizar:
        self.analyze_trend()
        self.analyze_frameworks()
        self.analyze_critical_tests()
        self.analyze_correlations()
        self.analyze_volatility()
        
        print(f"Análise concluída. Resultados disponíveis em: {self.output_dir}")
        return True
        
    def analyze_trend(self):
        """Analisa a tendência geral da pontuação de compliance ao longo do tempo"""
        if self.df_scores.empty:
            return
            
        print("Analisando tendências de compliance...")
        
        plt.figure(figsize=(12, 6))
        
        # Pontuação geral ao longo do tempo
        sns.lineplot(data=self.df_scores, x="timestamp", y="complianceScore", marker="o", linewidth=2)
        
        # Adicionar linha de tendência
        if len(self.df_scores) > 1:
            z = np.polyfit(range(len(self.df_scores)), self.df_scores["complianceScore"], 1)
            p = np.poly1d(z)
            plt.plot(self.df_scores["timestamp"], p(range(len(self.df_scores))), "r--", linewidth=1)
            
            # Calcular taxa de variação
            slope = z[0]
            trend_direction = "positiva" if slope > 0 else "negativa"
            print(f"Tendência {trend_direction} detectada: {slope:.4f} pontos por período")
        
        plt.title(f"Tendência de Conformidade ao Longo do Tempo - {self.region if self.region else 'Todas Regiões'}")
        plt.xlabel("Data")
        plt.ylabel("Pontuação de Conformidade (%)")
        plt.grid(True, linestyle="--", alpha=0.7)
        plt.xticks(rotation=45)
        plt.tight_layout()
        
        # Adicionar linhas de limiar
        plt.axhline(y=90, color="green", linestyle="-.", alpha=0.5, label="Conformidade Alta (90%)")
        plt.axhline(y=70, color="orange", linestyle="-.", alpha=0.5, label="Conformidade Média (70%)")
        plt.axhline(y=50, color="red", linestyle="-.", alpha=0.5, label="Conformidade Crítica (50%)")
        
        plt.legend()
        
        # Salvar gráfico
        trend_path = os.path.join(self.output_dir, f"compliance_trend_{self.region if self.region else 'all'}.png")
        plt.savefig(trend_path, dpi=300)
        print(f"Gráfico de tendência salvo em: {trend_path}")
        
    def analyze_frameworks(self):
        """Analisa o desempenho por framework regulatório"""
        if self.df_frameworks.empty:
            return
            
        print("Analisando desempenho por framework regulatório...")
        
        # Último relatório por framework
        latest_frameworks = self.df_frameworks[self.df_frameworks["timestamp"] == self.df_frameworks["timestamp"].max()]
        
        plt.figure(figsize=(10, 8))
        
        # Ordenar do menor para o maior para melhor visualização
        latest_frameworks = latest_frameworks.sort_values("complianceScore")
        
        # Criar gráfico de barras
        bars = sns.barplot(data=latest_frameworks, x="complianceScore", y="frameworkName", palette="RdYlGn")
        
        # Adicionar valores às barras
        for i, v in enumerate(latest_frameworks["complianceScore"]):
            bars.text(v + 1, i, f"{v:.2f}%", va="center")
        
        plt.title(f"Conformidade por Framework - {self.region if self.region else 'Todas Regiões'}")
        plt.xlabel("Pontuação de Conformidade (%)")
        plt.ylabel("Framework Regulatório")
        plt.grid(True, linestyle="--", alpha=0.3, axis="x")
        plt.tight_layout()
        
        # Adicionar linhas de limiar
        plt.axvline(x=90, color="green", linestyle="-.", alpha=0.5, label="Conformidade Alta (90%)")
        plt.axvline(x=70, color="orange", linestyle="-.", alpha=0.5, label="Conformidade Média (70%)")
        plt.axvline(x=50, color="red", linestyle="-.", alpha=0.5, label="Conformidade Crítica (50%)")
        
        plt.legend(loc="lower right")
        
        # Salvar gráfico
        frameworks_path = os.path.join(self.output_dir, f"frameworks_compliance_{self.region if self.region else 'all'}.png")
        plt.savefig(frameworks_path, dpi=300)
        print(f"Gráfico de frameworks salvo em: {frameworks_path}")
        
        # Analisar tendências por framework
        plt.figure(figsize=(12, 8))
        
        # Filtrar pelos frameworks mais importantes
        top_frameworks = latest_frameworks.nlargest(5, "totalTests")["frameworkId"].tolist()
        
        # Filtrar dados apenas para os top frameworks
        df_top_frameworks = self.df_frameworks[self.df_frameworks["frameworkId"].isin(top_frameworks)]
        
        # Plot por framework
        sns.lineplot(data=df_top_frameworks, x="timestamp", y="complianceScore", hue="frameworkName", marker="o")
        
        plt.title(f"Tendência de Conformidade por Framework - {self.region if self.region else 'Todas Regiões'}")
        plt.xlabel("Data")
        plt.ylabel("Pontuação de Conformidade (%)")
        plt.grid(True, linestyle="--", alpha=0.7)
        plt.xticks(rotation=45)
        plt.legend(title="Framework")
        plt.tight_layout()
        
        # Salvar gráfico
        fw_trends_path = os.path.join(self.output_dir, f"framework_trends_{self.region if self.region else 'all'}.png")
        plt.savefig(fw_trends_path, dpi=300)
        
    def analyze_critical_tests(self):
        """Analisa os testes críticos que falham com mais frequência"""
        if self.df_tests.empty:
            return
            
        print("Analisando testes críticos com falhas frequentes...")
        
        # Filtrar apenas testes críticos que falharam
        critical_failures = self.df_tests[(self.df_tests["criticality"] == "alta") & (~self.df_tests["passed"])]
        
        if critical_failures.empty:
            print("Não foram encontradas falhas em testes críticos.")
            return
            
        # Agrupar por ID de teste e contar falhas
        failure_counts = critical_failures.groupby("testId").size().reset_index(name="failures")
        failure_counts = failure_counts.sort_values("failures", ascending=False)
        
        # Adicionar nome do teste
        test_names = critical_failures.groupby("testId")["testName"].first().reset_index()
        failure_counts = failure_counts.merge(test_names, on="testId")
        
        # Top 10 testes críticos que mais falham
        top_failures = failure_counts.head(10)
        
        plt.figure(figsize=(12, 6))
        
        bars = sns.barplot(data=top_failures, x="failures", y="testId")
        
        # Adicionar rótulos com nomes dos testes
        for i, (_, row) in enumerate(top_failures.iterrows()):
            bars.text(row["failures"] + 0.1, i, row["testName"], va="center")
        
        plt.title(f"Top Testes Críticos com Falhas - {self.region if self.region else 'Todas Regiões'}")
        plt.xlabel("Número de Falhas")
        plt.ylabel("ID do Teste")
        plt.grid(True, linestyle="--", alpha=0.3, axis="x")
        plt.tight_layout()
        
        # Salvar gráfico
        critical_path = os.path.join(self.output_dir, f"critical_failures_{self.region if self.region else 'all'}.png")
        plt.savefig(critical_path, dpi=300)
        print(f"Gráfico de falhas críticas salvo em: {critical_path}")
        
        # Salvar dados detalhados
        details_path = os.path.join(self.output_dir, f"critical_failures_details_{self.region if self.region else 'all'}.csv")
        critical_failures.to_csv(details_path, index=False)
        
    def analyze_correlations(self):
        """Analisa correlações entre diferentes métricas"""
        if self.df_scores.empty:
            return
            
        print("Analisando correlações entre métricas...")
        
        # Preparar dados para correlação
        corr_data = self.df_scores[["complianceScore", "totalTests", "failedTests", "duration"]]
        
        # Calcular matriz de correlação
        corr_matrix = corr_data.corr()
        
        plt.figure(figsize=(8, 6))
        
        # Criar heatmap
        sns.heatmap(corr_matrix, annot=True, cmap="coolwarm", vmin=-1, vmax=1, center=0)
        
        plt.title(f"Correlações entre Métricas de Compliance - {self.region if self.region else 'Todas Regiões'}")
        plt.tight_layout()
        
        # Salvar gráfico
        corr_path = os.path.join(self.output_dir, f"correlations_{self.region if self.region else 'all'}.png")
        plt.savefig(corr_path, dpi=300)
        print(f"Matriz de correlação salva em: {corr_path}")
        
    def analyze_volatility(self):
        """Analisa a volatilidade nas pontuações de compliance"""
        if len(self.df_scores) <= 1:
            return
            
        print("Analisando volatilidade da conformidade...")
        
        # Calcular diferenças entre pontuações consecutivas
        self.df_scores["score_diff"] = self.df_scores["complianceScore"].diff()
        
        plt.figure(figsize=(12, 6))
        
        # Gráfico de volatilidade
        sns.barplot(data=self.df_scores[1:], x=self.df_scores[1:]["timestamp"].dt.strftime("%Y-%m-%d"), y="score_diff", palette="RdBu_r")
        
        plt.title(f"Volatilidade na Pontuação de Compliance - {self.region if self.region else 'Todas Regiões'}")
        plt.xlabel("Data")
        plt.ylabel("Variação de Pontuação (%)")
        plt.grid(True, linestyle="--", alpha=0.3, axis="y")
        plt.axhline(y=0, color="black", linestyle="-", alpha=0.3)
        plt.xticks(rotation=45)
        plt.tight_layout()
        
        # Salvar gráfico
        volatility_path = os.path.join(self.output_dir, f"volatility_{self.region if self.region else 'all'}.png")
        plt.savefig(volatility_path, dpi=300)
        print(f"Gráfico de volatilidade salvo em: {volatility_path}")
        
        # Estatísticas de volatilidade
        volatility_stats = {
            "max_increase": self.df_scores["score_diff"].max(),
            "max_decrease": self.df_scores["score_diff"].min(),
            "avg_change": self.df_scores["score_diff"].abs().mean(),
            "std_dev": self.df_scores["score_diff"].std()
        }
        
        print(f"Estatísticas de volatilidade:")
        print(f"  - Maior aumento: {volatility_stats['max_increase']:.2f}%")
        print(f"  - Maior queda: {volatility_stats['max_decrease']:.2f}%")
        print(f"  - Variação média: {volatility_stats['avg_change']:.2f}%")
        print(f"  - Desvio padrão: {volatility_stats['std_dev']:.2f}%")
        
        # Salvar estatísticas
        with open(os.path.join(self.output_dir, f"volatility_stats_{self.region if self.region else 'all'}.json"), "w") as f:
            json.dump(volatility_stats, f, indent=2)
    
    def generate_summary_report(self):
        """Gera um relatório de resumo com os principais insights"""
        if self.df_scores.empty:
            return
            
        print("Gerando relatório de resumo...")
        
        # Dados do último relatório
        latest = self.df_scores[self.df_scores["timestamp"] == self.df_scores["timestamp"].max()].iloc[0]
        
        # Tendência
        if len(self.df_scores) > 1:
            z = np.polyfit(range(len(self.df_scores)), self.df_scores["complianceScore"], 1)
            trend_slope = z[0]
            trend_direction = "positiva" if trend_slope > 0 else "negativa"
        else:
            trend_slope = 0
            trend_direction = "estável"
            
        # Frameworks de menor pontuação
        latest_frameworks = self.df_frameworks[self.df_frameworks["timestamp"] == self.df_frameworks["timestamp"].max()]
        lowest_frameworks = latest_frameworks.nsmallest(3, "complianceScore")
        
        # Gerar relatório em markdown
        report_file = os.path.join(self.output_dir, f"compliance_summary_{self.region if self.region else 'all'}.md")
        
        with open(report_file, "w") as f:
            f.write(f"# Resumo da Análise de Compliance\n\n")
            f.write(f"**Região:** {self.region if self.region else 'Todas'}\n")
            f.write(f"**Data da análise:** {datetime.now().strftime('%d/%m/%Y %H:%M:%S')}\n")
            f.write(f"**Período analisado:** {self.df_scores['timestamp'].min().strftime('%d/%m/%Y')} a {self.df_scores['timestamp'].max().strftime('%d/%m/%Y')}\n\n")
            
            f.write(f"## Pontuação Atual\n\n")
            f.write(f"- **Pontuação geral:** {latest['complianceScore']:.2f}%\n")
            f.write(f"- **Testes passados:** {latest['passedTests']} de {latest['totalTests']} ({latest['passedTests']/latest['totalTests']*100:.2f}%)\n")
            f.write(f"- **Testes falhos:** {latest['failedTests']}\n\n")
            
            f.write(f"## Tendências\n\n")
            f.write(f"- **Direção da tendência:** {trend_direction}\n")
            f.write(f"- **Taxa de variação:** {trend_slope:.4f} pontos por período\n")
            
            if len(self.df_scores) > 1:
                first = self.df_scores.iloc[0]
                last = self.df_scores.iloc[-1]
                total_change = last["complianceScore"] - first["complianceScore"]
                f.write(f"- **Variação total no período:** {total_change:.2f}%\n")
                
            # Adicionar frameworks problemáticos
            if not lowest_frameworks.empty:
                f.write(f"\n## Frameworks com Menor Pontuação\n\n")
                for _, fw in lowest_frameworks.iterrows():
                    f.write(f"- **{fw['frameworkName']}:** {fw['complianceScore']:.2f}% ({fw['passedTests']} de {fw['totalTests']} testes)\n")
            
            # Adicionar testes críticos com falha
            critical_failures = self.df_tests[(self.df_tests["criticality"] == "alta") & 
                                              (~self.df_tests["passed"]) &
                                              (self.df_tests["timestamp"] == self.df_tests["timestamp"].max())]
            
            if not critical_failures.empty:
                f.write(f"\n## Testes Críticos com Falha\n\n")
                for i, (_, test) in enumerate(critical_failures.iterrows()):
                    if i >= 5:  # Limitar a 5 testes
                        remaining = len(critical_failures) - 5
                        f.write(f"\n... e mais {remaining} falhas críticas.\n")
                        break
                    f.write(f"- **{test['testId']}:** {test['testName']}\n")
            
            # Recomendações gerais
            f.write(f"\n## Recomendações\n\n")
            
            if latest["complianceScore"] < 70:
                f.write("⚠️ **ATENÇÃO: Pontuação abaixo do limite mínimo aceitável (70%)!**\n\n")
                f.write("Recomendações de alta prioridade:\n")
                f.write("1. Revisar urgentemente os testes críticos com falha\n")
                f.write("2. Aplicar remediação automática sempre que possível\n")
                f.write("3. Agendar revisão manual das políticas OPA\n")
            elif latest["complianceScore"] < 90:
                f.write("⚠️ **Pontuação abaixo do objetivo (90%).**\n\n")
                f.write("Recomendações:\n")
                f.write("1. Revisar os testes críticos com falha\n")
                f.write("2. Considerar remediação automática para problemas comuns\n")
            else:
                f.write("✅ **Pontuação satisfatória.**\n\n")
                f.write("Recomendações para manutenção:\n")
                f.write("1. Monitorar tendências e volatilidade\n")
                f.write("2. Verificar novos requisitos regulatórios\n")
            
            # Adicionar referência às imagens
            f.write(f"\n## Visualizações\n\n")
            f.write(f"- [Tendência Geral](./compliance_trend_{self.region if self.region else 'all'}.png)\n")
            f.write(f"- [Compliance por Framework](./frameworks_compliance_{self.region if self.region else 'all'}.png)\n")
            f.write(f"- [Tendência por Framework](./framework_trends_{self.region if self.region else 'all'}.png)\n")
            
            if os.path.exists(os.path.join(self.output_dir, f"critical_failures_{self.region if self.region else 'all'}.png")):
                f.write(f"- [Falhas Críticas](./critical_failures_{self.region if self.region else 'all'}.png)\n")
            
            f.write(f"- [Volatilidade](./volatility_{self.region if self.region else 'all'}.png)\n")
            f.write(f"- [Correlações](./correlations_{self.region if self.region else 'all'}.png)\n")
            
        print(f"Relatório de resumo gerado em: {report_file}")


def main():
    parser = argparse.ArgumentParser(description="Analisador de Resultados de Compliance")
    parser.add_argument("--reports-dir", required=True, help="Diretório contendo os relatórios de compliance")
    parser.add_argument("--region", help="Região específica a analisar (opcional)")
    parser.add_argument("--output", default="./compliance-analysis", help="Diretório para saída dos resultados da análise")
    args = parser.parse_args()
    
    # Inicializar e executar o analisador
    analyzer = ComplianceAnalyzer(
        reports_dir=args.reports_dir,
        region=args.region,
        output_dir=args.output
    )
    
    if analyzer.load_reports():
        analyzer.analyze()
        analyzer.generate_summary_report()
    else:
        print("Não foi possível carregar relatórios para análise.")
        return 1
    
    return 0


if __name__ == "__main__":
    exit(main())