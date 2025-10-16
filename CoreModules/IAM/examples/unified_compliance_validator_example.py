"""
Exemplo de uso do Validador de Conformidade Unificado

Este exemplo demonstra o uso do UnifiedComplianceValidator para validar
operações em diferentes contextos regionais e cenários multi-mercado.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

import sys
import os
import json
from pprint import pprint

# Adiciona o diretório raiz ao path para permitir importações
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

# Importa o módulo de observabilidade
from infrastructure.observability.core.multi_layer_monitor import MultiLayerMonitor

# Importa o validador de conformidade unificado
from infrastructure.trustguard.compliance.unified_compliance_validator import UnifiedComplianceValidator

def main():
    """Função principal do exemplo."""
    print("Inicializando exemplo de Validador de Conformidade Unificado...")
    
    # Inicializa o monitor de observabilidade
    observability_monitor = MultiLayerMonitor(
        app_name="unified-compliance-example",
        log_level="INFO",
        enable_console=True
    )
    
    # Inicializa o validador de conformidade unificado
    validator = UnifiedComplianceValidator(observability_monitor)
    
    # Demonstra diferentes cenários de validação
    demonstrar_validacao_brasil(validator)
    demonstrar_validacao_africa_do_sul(validator)
    demonstrar_validacao_operacao_cross_regional(validator)
    demonstrar_validacao_multi_regiao(validator)

def demonstrar_validacao_brasil(validator):
    """Demonstra validação para operação no Brasil."""
    print("\n\n===== VALIDAÇÃO: OPERAÇÃO NO BRASIL (BRICS) =====")
    
    # Cria contexto de operação no Brasil
    contexto = {
        "localizacao": {
            "pais": "BR",
            "cidade": "São Paulo"
        },
        "dados_pessoais": {
            "nome": "João Silva",
            "cpf": "123.456.789-00",
            "endereco": "Av. Paulista, 1000"
        },
        "dados_financeiros": {
            "valor_transacao": 1000,
            "moeda": "BRL"
        },
        "consentimento_dados_pessoais": True,
        "tipo_operacao": "pagamento",
        "kyc_completo": True
    }
    
    # Executa validação
    resultado = validator.validar_conformidade(contexto)
    
    # Exibe resultado
    print("\nResultado da validação:")
    pprint(resultado)

def demonstrar_validacao_africa_do_sul(validator):
    """Demonstra validação para operação na África do Sul."""
    print("\n\n===== VALIDAÇÃO: OPERAÇÃO NA ÁFRICA DO SUL (BRICS) =====")
    
    # Cria contexto de operação na África do Sul
    contexto = {
        "localizacao": {
            "pais": "ZA",
            "cidade": "Joanesburgo"
        },
        "dados_pessoais": {
            "nome": "Nelson Mandela",
            "id_number": "5001010100000",  # Formato ID sul-africano
            "endereco": "Main Street, 123"
        },
        "dados_financeiros": {
            "valor_transacao": 5000,
            "moeda": "ZAR"
        },
        "consentimento_dados_pessoais": True,
        "tipo_operacao": "pagamento",
        "kyc_completo": True,
        "popia_compliance": {
            "dpo_designado": True,
            "consentimento_explicito": True,
            "finalidade_declarada": "Processamento de pagamento"
        }
    }
    
    # Executa validação
    resultado = validator.validar_conformidade(contexto)
    
    # Exibe resultado
    print("\nResultado da validação:")
    pprint(resultado)

def demonstrar_validacao_operacao_cross_regional(validator):
    """Demonstra validação para operação cross-regional (Brasil para Portugal)."""
    print("\n\n===== VALIDAÇÃO: OPERAÇÃO CROSS-REGIONAL (BRASIL → PORTUGAL) =====")
    
    # Cria contexto de operação cross-regional sem autorização específica
    contexto = {
        "tipo_operacao": "transferencia_internacional",
        "pais_origem": "BR",
        "pais_destino": "PT",
        "localizacao": {
            "pais": "BR",
            "cidade": "Rio de Janeiro"
        },
        "dados_pessoais": {
            "nome": "Maria Santos",
            "cpf": "987.654.321-00"
        },
        "dados_financeiros": {
            "valor_transacao": 3000,
            "moeda": "EUR"
        },
        "consentimento_dados_pessoais": True,
        "kyc_completo": True
        # Sem autorização cross-regional
    }
    
    # Executa validação
    resultado = validator.validar_conformidade(contexto)
    
    # Exibe resultado (deve falhar)
    print("\nResultado da validação (sem autorização cross-regional):")
    pprint(resultado)
    
    # Adiciona autorização e testa novamente
    contexto["autorizacao_transferencia_cross_regional"] = True
    
    resultado = validator.validar_conformidade(contexto)
    
    # Exibe resultado (deve passar)
    print("\nResultado da validação (com autorização cross-regional):")
    pprint(resultado)

def demonstrar_validacao_multi_regiao(validator):
    """Demonstra validação para operação com múltiplas regiões especificadas explicitamente."""
    print("\n\n===== VALIDAÇÃO: OPERAÇÃO COM MÚLTIPLAS REGIÕES ESPECIFICADAS =====")
    
    # Cria contexto com regiões explicitamente especificadas
    contexto = {
        "localizacao": {
            "pais": "AO",  # Angola
            "cidade": "Luanda"
        },
        "regioes_aplicaveis": ["SADC", "CPLP", "PALOP"],
        "dados_pessoais": {
            "nome": "António Silva",
            "bi": "123456789LA042"
        },
        "dados_financeiros": {
            "valor_transacao": 1000,
            "moeda": "AOA"
        },
        "consentimento_dados_pessoais": True,
        "tipo_operacao": "pagamento",
        "kyc_completo": True
    }
    
    # Executa validação
    resultado = validator.validar_conformidade(contexto)
    
    # Exibe resultado
    print("\nResultado da validação (Angola com múltiplas regiões aplicáveis):")
    pprint(resultado)

if __name__ == "__main__":
    main()
