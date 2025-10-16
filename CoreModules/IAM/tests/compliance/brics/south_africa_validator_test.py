"""
Testes para o validador de conformidade da África do Sul.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

import unittest
from unittest.mock import MagicMock
import sys
import os

# Adiciona o diretório raiz ao path para permitir importações
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../../..')))

from infrastructure.trustguard.compliance.brics.south_africa_validator import SouthAfricaComplianceValidator

class TestSouthAfricaComplianceValidator(unittest.TestCase):
    """Testes para o validador de conformidade da África do Sul."""
    
    def setUp(self):
        """Configura ambiente de teste."""
        # Mock para logger
        self.logger_mock = MagicMock()
        # Mock para métricas
        self.metrics_mock = MagicMock()
        # Instância do validador
        self.validador = SouthAfricaComplianceValidator(
            logger=self.logger_mock,
            metrics=self.metrics_mock
        )
        
        # Resultado base para testes
        self.resultado_base = {
            "valido": True,
            "violacoes": [],
            "avisos": [],
            "score_conformidade": 1.0
        }
        
    def test_validar_dados_pessoais_ok(self):
        """Testa validação de dados pessoais válidos."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_pessoais": {
                "consentimento_explicito_za": True
            },
            "mecanismo_direito_acesso_za": True,
            "mecanismo_direito_correcao_za": True,
            "mecanismo_direito_exclusao_za": True
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_falta_consentimento(self):
        """Testa validação quando falta consentimento explícito."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_pessoais": {
                "consentimento_explicito_za": False
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_POPIA_FALTA_CONSENTIMENTO")
        
    def test_validar_fica_transacao_reportada(self):
        """Testa validação de transação acima do limite reportável corretamente reportada."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "valor_transacao": 30000,
                "moeda": "ZAR",
                "reportado_fic_za": True
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_fica_transacao_nao_reportada(self):
        """Testa validação de transação acima do limite reportável não reportada."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "valor_transacao": 30000,
                "moeda": "ZAR",
                "reportado_fic_za": False
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_FICA_TRANSACAO_NAO_REPORTADA")
        
    def test_validar_limites_cambio_dentro_limite(self):
        """Testa validação de limites de câmbio dentro do permitido."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "tipo_transacao": "cambio",
                "valor_transacao": 500000,
                "valor_acumulado_anual": 400000
            },
            "tipo_cliente": "individual"
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_limites_cambio_excedido(self):
        """Testa validação de limites de câmbio excedidos."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "tipo_transacao": "cambio",
                "valor_transacao": 800000,
                "valor_acumulado_anual": 400000
            },
            "tipo_cliente": "individual"
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_SARB_LIMITE_CAMBIO_EXCEDIDO")
        
    def test_validar_limites_cambio_com_autorizacao(self):
        """Testa validação de limites de câmbio excedidos com autorização."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "tipo_transacao": "cambio",
                "valor_transacao": 800000,
                "valor_acumulado_anual": 400000,
                "autorizacao_especial_sarb": True
            },
            "tipo_cliente": "individual"
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_fsca_sem_licenca(self):
        """Testa validação de provedor de serviço financeiro sem licença."""
        resultado = self.resultado_base.copy()
        contexto = {
            "tipo_entidade": "provedor_servico_financeiro",
            "licenca_fsca": False
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_FSCA_FALTA_LICENCA")
        
    def test_validar_fsca_com_licenca(self):
        """Testa validação de provedor de serviço financeiro com licença."""
        resultado = self.resultado_base.copy()
        contexto = {
            "tipo_entidade": "provedor_servico_financeiro",
            "licenca_fsca": True
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_sistema_pagamento_invalido(self):
        """Testa validação de sistema de pagamento não aprovado."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "sistema_pagamento": "sistema_nao_aprovado"
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_SARB_SISTEMA_PAGAMENTO_NAO_APROVADO")
        
    def test_validar_sistema_pagamento_valido(self):
        """Testa validação de sistema de pagamento aprovado."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "sistema_pagamento": "rtgs"
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)
        
    def test_validar_tratamento_justo_clientes(self):
        """Testa validação de tratamento justo de clientes."""
        resultado = self.resultado_base.copy()
        contexto = {
            "tcf_implementado_za": False
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])  # É apenas um aviso, não uma violação bloqueante
        self.assertEqual(len(resultado["avisos"]), 1)
        self.assertEqual(resultado["avisos"][0]["codigo"], "ZA_FSCA_FALTA_TCF")
        
    def test_validar_transferencia_internacional_sem_declaracao(self):
        """Testa validação de transferência internacional sem declaração obrigatória."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "tipo_transacao": "internacional",
                "declaracao_sarb": False
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertFalse(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "ZA_SARB_FALTA_DECLARACAO")
        
    def test_validar_transferencia_internacional_com_declaracao(self):
        """Testa validação de transferência internacional com declaração obrigatória."""
        resultado = self.resultado_base.copy()
        contexto = {
            "dados_financeiros": {
                "tipo_transacao": "internacional",
                "declaracao_sarb": True
            }
        }
        
        self.validador.validar(resultado, contexto)
        
        self.assertTrue(resultado["valido"])
        self.assertEqual(len(resultado["violacoes"]), 0)

if __name__ == '__main__':
    unittest.main()
