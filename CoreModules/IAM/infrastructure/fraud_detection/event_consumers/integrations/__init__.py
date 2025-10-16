#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Módulo de integrações para o sistema de análise comportamental.

Este pacote contém os adaptadores e integrações para conectar o sistema de
análise comportamental com outros módulos da plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from .uniconnect_notifier import (
    UniConnectNotifier,
    NotificationChannel,
    NotificationPriority,
    NotificationRecipient,
    create_uniconnect_notifier
)

__all__ = [
    'UniConnectNotifier',
    'NotificationChannel',
    'NotificationPriority',
    'NotificationRecipient',
    'create_uniconnect_notifier'
]