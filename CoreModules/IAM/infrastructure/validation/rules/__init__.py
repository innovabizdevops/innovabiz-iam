"""
INNOVABIZ - Regras de Validação IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Pacote de regras de validação para o módulo IAM,
           incluindo regras específicas para múltiplos frameworks
           regulatórios e boas práticas de segurança.
==================================================================
"""

from .hipaa_rules import Validator as HIPAAValidator

__all__ = [
    "HIPAAValidator"
]
