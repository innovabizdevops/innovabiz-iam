"""
INNOVABIZ IAM - Middleware Package
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Middlewares para o sistema IAM
"""

from .audit_middleware import AuditMiddleware

__all__ = ["AuditMiddleware"]