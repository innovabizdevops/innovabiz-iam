"""
Factory para criação de conectores TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import logging
import os
from typing import Optional

from fastapi import Depends

from infrastructure.fraud_detection.neuraflow.rule_enhancer import RuleEnhancer
from infrastructure.fraud_detection.rules_engine.evaluator import RuleEvaluator

from .trust_guard_connector import TrustGuardConnector, TrustGuardConfig


def get_trust_guard_config() -> TrustGuardConfig:
    """
    Factory para criação da configuração do TrustGuard.
    
    Returns:
        TrustGuardConfig: Configuração do TrustGuard
    """
    return TrustGuardConfig(
        api_url=os.environ.get("TRUST_GUARD_API_URL", "http://trustguard-api.innovabiz.io/v1"),
        api_key=os.environ.get("TRUST_GUARD_API_KEY", ""),
        tenant_id=os.environ.get("TRUST_GUARD_TENANT_ID", "default"),
        timeout=int(os.environ.get("TRUST_GUARD_TIMEOUT", "10")),
        cache_ttl=int(os.environ.get("TRUST_GUARD_CACHE_TTL", "300")),
        cache_enabled=os.environ.get("TRUST_GUARD_CACHE_ENABLED", "true").lower() == "true",
    )


async def get_trust_guard_connector(
    config: TrustGuardConfig = Depends(get_trust_guard_config),
    rule_evaluator: Optional[RuleEvaluator] = None,
    rule_enhancer: Optional[RuleEnhancer] = None,
) -> TrustGuardConnector:
    """
    Factory para criação do conector TrustGuard.
    
    Args:
        config: Configuração do TrustGuard
        rule_evaluator: Avaliador de regras
        rule_enhancer: Enhancer de regras
        
    Returns:
        TrustGuardConnector: Conector TrustGuard
    """
    logger = logging.getLogger("trust_guard_connector")
    
    connector = TrustGuardConnector(
        config=config,
        rule_evaluator=rule_evaluator,
        rule_enhancer=rule_enhancer,
        logger=logger,
    )
    
    await connector.init_client()
    
    try:
        yield connector
    finally:
        await connector.close()