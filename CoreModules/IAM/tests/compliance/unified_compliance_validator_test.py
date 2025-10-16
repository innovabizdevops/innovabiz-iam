"""
Testes para o validador de conformidade unificado.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

import unittest
from unittest.mock import MagicMock, patch
import sys
import os

# Adiciona o diretório raiz ao path para permitir importações
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../..')))

from infrastructure.trustguard.compliance.unified_compliance_validator import UnifiedComplianceValidator
from infrastructure.trustguard.compliance.brics_compliance_validator import BRICSComplianceValidator

class TestUnifiedComplianceValidator(unittest.TestCase):
    """Testes para o validador de conformidade unificado."""
    
    def setUp(self):
        """Configura ambiente de teste."""
        # Mock para observabilidade
        self.observability_mock = MagicMock()
        self.observability_mock.getLogger.return_value = MagicMock()
        self.observability_mock.getMetrics.return_value = MagicMock()
        
        # Patchs para validadores regionais
        self.brics_validator_patch = patch('infrastructure.trustguard.compliance.brics_compliance_validator.BRICSComplianceValidator')
        self.mock_brics_validator = self.brics_validator_patch.start()
        self.mock_brics_instance = MagicMock()
        self.mock_brics_validator.return_value = self.mock_brics_instance
        
        # Instância do validador unificado
        self.validator = UnifiedComplianceValidator(self.observability_mock)
    
    def tearDown(self):
        """Limpa ambiente após testes."""
        self.brics_validator_patch.stop()
    
    def test_validar_conformidade_pais_brics(self):
        """Testa validação para país do BRICS."""
        # Configurar mock do validador BRICS
        self.mock_brics_instance.validar_conformidade.return_value = {
            "valido": True,
            "violacoes": [],
            "avisos": [],
            "score_conformidade": 1.0
        }
        
        # Contexto com país do BRICS (Brasil)
        contexto = {
            "localizacao": {
                "pais": "BR"
            },
            "dados_pessoais": {
                "nome": "Teste",
                "cpf": "123.456.789-00"
            }
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Verificar se o validador BRICS foi chamado
        self.mock_brics_instance.validar_conformidade.assert_called_once_with(contexto)
        
        # Verificar resultado
        self.assertTrue(resultado["valido"])
        self.assertEqual(resultado["score_conformidade"], 1.0)
    
    def test_validar_conformidade_pais_europa(self):
        """Testa validação para país da Europa."""
        # Configurar mock do validador de Europa (não implementado ainda)
        # Por enquanto, o sistema deve usar validação genérica
        
        # Contexto com país europeu (Portugal)
        contexto = {
            "localizacao": {
                "pais": "PT"
            },
            "dados_pessoais": {
                "nome": "Teste",
                "nif": "123456789"
            },
            "consentimento_dados_pessoais": True
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Verificar resultado
        self.assertTrue(resultado["valido"])  # Deve passar com validação genérica se os campos requeridos existem
    
    def test_validar_conformidade_pais_nao_identificado(self):
        """Testa validação quando país não é identificado."""
        contexto = {
            "dados_pessoais": {
                "nome": "Teste"
            }
            # Sem país especificado
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Validação genérica deve ser aplicada, verificar se pede consentimento
        self.assertFalse(resultado["valido"])  # Deve falhar sem consentimento
        self.assertTrue(any("FALTA_CONSENTIMENTO_DADOS" in str(v) for v in resultado["violacoes"]))
    
    def test_validar_conformidade_cross_regional(self):
        """Testa validação para operação cross-regional."""
        # Configurar mock do validador BRICS
        self.mock_brics_instance.validar_conformidade.return_value = {
            "valido": True,
            "violacoes": [],
            "avisos": [],
            "score_conformidade": 1.0
        }
        
        # Contexto de transferência internacional entre Brasil e Portugal
        contexto = {
            "tipo_operacao": "transferencia_internacional",
            "pais_origem": "BR",
            "pais_destino": "PT",
            "dados_financeiros": {
                "valor_transacao": 5000,
                "moeda": "EUR"
            },
            "consentimento_dados_pessoais": True,
            "kyc_completo": True
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Verificar se o validador BRICS foi chamado
        self.mock_brics_instance.validar_conformidade.assert_called_once_with(contexto)
        
        # Deve falhar por falta de autorização cross-regional
        self.assertFalse(resultado["valido"])
        self.assertTrue(any("TRANSFERENCIA_CROSS_REGIONAL_SEM_AUTORIZACAO" in str(v) for v in resultado["violacoes"]))
    
    def test_validar_conformidade_cross_regional_com_autorizacao(self):
        """Testa validação para operação cross-regional com autorização."""
        # Configurar mock do validador BRICS
        self.mock_brics_instance.validar_conformidade.return_value = {
            "valido": True,
            "violacoes": [],
            "avisos": [],
            "score_conformidade": 1.0
        }
        
        # Contexto de transferência internacional entre Brasil e Portugal com autorização
        contexto = {
            "tipo_operacao": "transferencia_internacional",
            "pais_origem": "BR",
            "pais_destino": "PT",
            "dados_financeiros": {
                "valor_transacao": 5000,
                "moeda": "EUR"
            },
            "consentimento_dados_pessoais": True,
            "kyc_completo": True,
            "autorizacao_transferencia_cross_regional": True
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Verificar se o validador BRICS foi chamado
        self.mock_brics_instance.validar_conformidade.assert_called_once_with(contexto)
        
        # Deve passar com a autorização
        self.assertTrue(resultado["valido"])
    
    def test_validar_conformidade_com_violacao_regional(self):
        """Testa validação quando validador regional encontra violação."""
        # Configurar mock do validador BRICS para retornar violação
        self.mock_brics_instance.validar_conformidade.return_value = {
            "valido": False,
            "violacoes": [{
                "codigo": "BR_LGPD_FALTA_CONSENTIMENTO",
                "descricao": "LGPD requer consentimento explícito",
                "severidade": "alta",
                "impacto": "bloqueante"
            }],
            "avisos": [],
            "score_conformidade": 0.7
        }
        
        # Contexto com país do BRICS (Brasil)
        contexto = {
            "localizacao": {
                "pais": "BR"
            },
            "dados_pessoais": {
                "nome": "Teste",
                "cpf": "123.456.789-00"
            }
        }
        
        resultado = self.validator.validar_conformidade(contexto)
        
        # Verificar resultado
        self.assertFalse(resultado["valido"])
        self.assertEqual(resultado["score_conformidade"], 0.7)
        self.assertEqual(len(resultado["violacoes"]), 1)
        self.assertEqual(resultado["violacoes"][0]["codigo"], "BR_LGPD_FALTA_CONSENTIMENTO")
    
    def test_obter_pais_contexto_localizacao(self):
        """Testa obtenção de país do contexto pela localização."""
        contexto = {
            "localizacao": {
                "pais": "br"  # Minúsculo para testar normalização
            }
        }
        
        pais = self.validator._obter_pais_contexto(contexto)
        
        self.assertEqual(pais, "BR")
    
    def test_obter_pais_contexto_dados_pessoais(self):
        """Testa obtenção de país do contexto pelos dados pessoais."""
        contexto = {
            "dados_pessoais": {
                "pais": "PT"
            }
        }
        
        pais = self.validator._obter_pais_contexto(contexto)
        
        self.assertEqual(pais, "PT")
    
    def test_obter_pais_contexto_dados_financeiros(self):
        """Testa obtenção de país do contexto pelos dados financeiros."""
        contexto = {
            "dados_financeiros": {
                "pais": "US"
            }
        }
        
        pais = self.validator._obter_pais_contexto(contexto)
        
        self.assertEqual(pais, "US")
    
    def test_obter_regioes_aplicaveis_por_pais(self):
        """Testa obtenção de regiões aplicáveis por país."""
        pais = "BR"
        contexto = {}
        
        regioes = self.validator._obter_regioes_aplicaveis(pais, contexto)
        
        self.assertIn("BRICS", regioes)
        self.assertIn("MERCOSUL", regioes)
        self.assertIn("LATAM", regioes)
    
    def test_obter_regioes_aplicaveis_explicit(self):
        """Testa obtenção de regiões aplicáveis explicitamente especificadas."""
        pais = "BR"
        contexto = {
            "regioes_aplicaveis": ["BRICS", "custom_region"]
        }
        
        regioes = self.validator._obter_regioes_aplicaveis(pais, contexto)
        
        self.assertIn("BRICS", regioes)
        self.assertIn("MERCOSUL", regioes)
        self.assertIn("LATAM", regioes)
        self.assertIn("CUSTOM_REGION", regioes)  # Deve estar normalizado para maiúsculas
    
    def test_obter_regioes_aplicaveis_transferencia(self):
        """Testa obtenção de regiões aplicáveis para transferência internacional."""
        pais = "BR"
        contexto = {
            "tipo_operacao": "transferencia_internacional",
            "pais_destino": "PT"
        }
        
        regioes = self.validator._obter_regioes_aplicaveis(pais, contexto)
        
        self.assertIn("BRICS", regioes)  # Do país de origem
        self.assertIn("EU", regioes)     # Do país de destino
        self.assertIn("CPLP", regioes)   # Do país de destino

if __name__ == '__main__':
    unittest.main()
