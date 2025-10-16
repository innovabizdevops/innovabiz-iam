"""
Módulo de inicialização do controlador TrustScore.

Este módulo exporta o router e o controlador TrustScore
para integração com a aplicação principal.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from .trustscore_controller import router, controller, TrustScoreGraphQLController

__all__ = ["router", "controller", "TrustScoreGraphQLController"]