"""
INNOVABIZ - API GraphQL para Validação IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Pacote GraphQL para exposição das funcionalidades de
           validação, certificação e conformidade do módulo IAM,
           integrando com o sistema de autenticação e autorização.
==================================================================
"""

from .schema import schema
from .server import start_server

__all__ = ["schema", "start_server"]
