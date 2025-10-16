#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Pacote GraphQL para análise comportamental

Este pacote fornece uma API GraphQL para consulta de eventos, alertas e análises comportamentais
do sistema de detecção de fraudes do INNOVABIZ IAM/TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from .schema import schema
from .resolvers import bind_resolvers_to_schema
from .controller import get_graphql_router

__all__ = ['schema', 'bind_resolvers_to_schema', 'get_graphql_router']