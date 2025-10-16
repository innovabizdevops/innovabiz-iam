"""
Executor de Testes de Conformidade Regulatória para Mercados-Alvo
"""
import unittest
import sys
import os
import json
from datetime import datetime

# Importar os testes específicos de cada país/região
from angola_compliance_test import AngolaComplianceTest
from brasil_lgpd_compliance_test import BrasilLGPDComplianceTest
from south_africa_popia_test import SouthAfricaPOPIAComplianceTest

def run_all_tests(output_format='text', output_file=None):
    """
    Executa todos os testes de conformidade e gera relatório
    
    Args:
        output_format: Formato de saída ('text', 'json', 'html')
        output_file: Arquivo para salvar o relatório (opcional)
    """
    # Criar suite com todos os testes
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # Adicionar testes de Angola
    suite.addTests(loader.loadTestsFromTestCase(AngolaComplianceTest))
    
    # Adicionar testes do Brasil
    suite.addTests(loader.loadTestsFromTestCase(BrasilLGPDComplianceTest))
    
    # Adicionar testes da África do Sul
    suite.addTests(loader.loadTestsFromTestCase(SouthAfricaPOPIAComplianceTest))
    
    # Definir formato de saída
    if output_format == 'text':
        runner = unittest.TextTestRunner(verbosity=2)
        result = runner.run(suite)
        
        # Se arquivo de saída for especificado
        if output_file:
            with open(output_file, 'w') as f:
                f.write(f"Relatório de Conformidade Regulatória\n")
                f.write(f"Data: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
                f.write(f"Testes Executados: {result.testsRun}\n")
                f.write(f"Testes com Sucesso: {result.testsRun - len(result.failures) - len(result.errors)}\n")
                f.write(f"Falhas: {len(result.failures)}\n")
                f.write(f"Erros: {len(result.errors)}\n\n")
                
                if result.failures:
                    f.write("=== FALHAS ===\n\n")
                    for test, error in result.failures:
                        f.write(f"- {test}\n")
                        f.write(f"{error}\n\n")
                        
                if result.errors:
                    f.write("=== ERROS ===\n\n")
                    for test, error in result.errors:
                        f.write(f"- {test}\n")
                        f.write(f"{error}\n\n")
        
        return result
        
    elif output_format == 'json':
        # Usar JSONTestRunner
        result = JSONTestRunner().run(suite)
        
        if output_file:
            with open(output_file, 'w') as f:
                json.dump(result, f, indent=2)
                
        return result
        
    elif output_format == 'html':
        # Usar HTMLTestRunner se disponível
        try:
            from HTMLTestRunner import HTMLTestRunner
            
            if output_file:
                with open(output_file, 'w') as f:
                    runner = HTMLTestRunner(
                        stream=f,
                        title='Relatório de Conformidade Regulatória',
                        description='Testes de conformidade para Angola, Brasil e SADC'
                    )
                    result = runner.run(suite)
            else:
                print("Arquivo de saída necessário para relatório HTML")
                return None
                
            return result
            
        except ImportError:
            print("HTMLTestRunner não está disponível. Usando formato texto.")
            return run_all_tests('text', output_file)
    
    else:
        print(f"Formato de saída não suportado: {output_format}")
        return None


class JSONTestRunner(unittest.TextTestResult):
    """Custom test runner que gera resultados em formato JSON"""
    
    def __init__(self, *args, **kwargs):
        super().__init__(stream=sys.stdout, descriptions=True, verbosity=2)
        self.results = {
            "timestamp": datetime.now().isoformat(),
            "total_tests": 0,
            "passed": 0,
            "failures": 0,
            "errors": 0,
            "test_cases": []
        }
        
    def startTest(self, test):
        super().startTest(test)
        self.results["total_tests"] += 1
        
    def addSuccess(self, test):
        super().addSuccess(test)
        self.results["passed"] += 1
        self.results["test_cases"].append({
            "name": str(test),
            "status": "pass",
            "message": ""
        })
        
    def addFailure(self, test, err):
        super().addFailure(test, err)
        self.results["failures"] += 1
        self.results["test_cases"].append({
            "name": str(test),
            "status": "fail",
            "message": str(err[1])
        })
        
    def addError(self, test, err):
        super().addError(test, err)
        self.results["errors"] += 1
        self.results["test_cases"].append({
            "name": str(test),
            "status": "error",
            "message": str(err[1])
        })
        
    def run(self, test):
        test(self)
        return self.results


def generate_compliance_matrix():
    """
    Gera uma matriz de conformidade para visualizar a cobertura regulatória
    """
    # Definir requisitos regulatórios por país/região
    regulations = {
        "Angola": [
            "Minimização de dados", 
            "Limitação de finalidade",
            "Transparência",
            "Exatidão dos dados",
            "Medidas de segurança",
            "Portabilidade de dados",
            "Transferência internacional",
            "Registro de auditoria"
        ],
        "Brasil (LGPD)": [
            "Base legal",
            "Direitos do titular",
            "Gestão de consentimento",
            "Oficial de proteção de dados",
            "Privacidade por design",
            "Medidas de segurança",
            "Notificação de violação",
            "Transferência internacional",
            "Registros de processamento"
        ],
        "África do Sul (POPIA)": [
            "Legalidade do processamento",
            "Limitação de processamento",
            "Especificação de finalidade",
            "Limitação de processamento adicional",
            "Qualidade da informação",
            "Transparência",
            "Medidas de segurança",
            "Participação do titular dos dados",
            "Oficial de informação",
            "Avaliação de impacto"
        ]
    }
    
    # Gerar matriz em formato markdown
    matrix = "# Matriz de Conformidade Regulatória\n\n"
    matrix += "| Requisito | Angola | Brasil (LGPD) | África do Sul (POPIA) |\n"
    matrix += "|-----------|:------:|:------------:|:--------------------:|\n"
    
    # Consolidar todos os requisitos únicos
    all_requirements = set()
    for country, reqs in regulations.items():
        all_requirements.update(reqs)
        
    # Criar linhas da matriz
    for req in sorted(all_requirements):
        matrix += f"| {req} "
        for country in ["Angola", "Brasil (LGPD)", "África do Sul (POPIA)"]:
            if req in regulations[country]:
                matrix += "| ✓ "
            else:
                matrix += "|   "
        matrix += "|\n"
        
    # Retornar matriz formatada
    return matrix


if __name__ == "__main__":
    # Processar argumentos de linha de comando
    import argparse
    
    parser = argparse.ArgumentParser(description='Executor de Testes de Conformidade Regulatória')
    parser.add_argument('--format', choices=['text', 'json', 'html'], default='text',
                        help='Formato de saída do relatório')
    parser.add_argument('--output', type=str, help='Arquivo de saída para o relatório')
    parser.add_argument('--matrix', action='store_true', 
                        help='Gerar matriz de conformidade em Markdown')
    
    args = parser.parse_args()
    
    # Executar todos os testes
    result = run_all_tests(args.format, args.output)
    
    # Gerar matriz de conformidade se solicitado
    if args.matrix:
        matrix = generate_compliance_matrix()
        matrix_file = 'compliance_matrix.md'
        with open(matrix_file, 'w') as f:
            f.write(matrix)
        print(f"Matriz de conformidade gerada em {matrix_file}")
        
    # Resumo dos resultados
    if result and args.format == 'text':
        print(f"\nResumo:")
        print(f"Testes Executados: {result.testsRun}")
        print(f"Testes com Sucesso: {result.testsRun - len(result.failures) - len(result.errors)}")
        print(f"Falhas: {len(result.failures)}")
        print(f"Erros: {len(result.errors)}")