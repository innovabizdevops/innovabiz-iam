"""
INNOVABIZ IAM - Context Module
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Módulo de gerenciamento de contexto regional e tenant para o sistema IAM
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA

Este módulo implementa o gerenciamento de contexto para suportar:
1. Multi-tenant: Isolamento de dados por tenant/organização
2. Multi-regional: Suporte para diferentes regiões com suas regulamentações específicas
3. Multi-compliance: Adaptação automática às regulamentações aplicáveis por contexto
"""

import os
import logging
from typing import Optional, Dict, Any, Callable
from functools import lru_cache
from fastapi import Request, Depends, Header, Query, HTTPException, status
from pydantic import BaseModel, Field

from ..models.audit import RegionalContext

# Logger configurado
logger = logging.getLogger("context")

# Mapeamento de códigos de país para contextos regionais
COUNTRY_CODE_TO_REGIONAL_CONTEXT = {
    # Brasil
    "BR": RegionalContext.BR,
    "BRA": RegionalContext.BR,
    "076": RegionalContext.BR,
    
    # Estados Unidos
    "US": RegionalContext.US,
    "USA": RegionalContext.US,
    "840": RegionalContext.US,
    
    # Europa (União Europeia)
    "EU": RegionalContext.EU,
    "EUR": RegionalContext.EU,
    "PT": RegionalContext.EU,  # Portugal
    "PRT": RegionalContext.EU,
    "620": RegionalContext.EU,
    "DE": RegionalContext.EU,  # Alemanha
    "DEU": RegionalContext.EU,
    "276": RegionalContext.EU,
    "FR": RegionalContext.EU,  # França
    "FRA": RegionalContext.EU,
    "250": RegionalContext.EU,
    "ES": RegionalContext.EU,  # Espanha
    "ESP": RegionalContext.EU,
    "724": RegionalContext.EU,
    "IT": RegionalContext.EU,  # Itália
    "ITA": RegionalContext.EU,
    "380": RegionalContext.EU,
    
    # Angola
    "AO": RegionalContext.AO,
    "AGO": RegionalContext.AO,
    "024": RegionalContext.AO
}

# Headers padrão para detecção de contexto
DEFAULT_REGIONAL_CONTEXT_HEADER = "X-Regional-Context"
DEFAULT_TENANT_HEADER = "X-Tenant-ID"
DEFAULT_COUNTRY_HEADER = "X-Country-Code"

# Parâmetros de query padrão para override do contexto
DEFAULT_REGIONAL_CONTEXT_PARAM = "regional_context"
DEFAULT_TENANT_PARAM = "tenant_id"
DEFAULT_COUNTRY_PARAM = "country_code"

class ContextConfig:
    """Configuração para detecção de contexto."""
    
    def __init__(self):
        """Inicializa a configuração de contexto com valores padrão ou do ambiente."""
        # Headers para detecção de contexto
        self.regional_context_header = os.environ.get(
            "REGIONAL_CONTEXT_HEADER", DEFAULT_REGIONAL_CONTEXT_HEADER
        )
        self.tenant_header = os.environ.get(
            "TENANT_HEADER", DEFAULT_TENANT_HEADER
        )
        self.country_header = os.environ.get(
            "COUNTRY_HEADER", DEFAULT_COUNTRY_HEADER
        )
        
        # Parâmetros de query para override
        self.regional_context_param = os.environ.get(
            "REGIONAL_CONTEXT_PARAM", DEFAULT_REGIONAL_CONTEXT_PARAM
        )
        self.tenant_param = os.environ.get(
            "TENANT_PARAM", DEFAULT_TENANT_PARAM
        )
        self.country_param = os.environ.get(
            "COUNTRY_PARAM", DEFAULT_COUNTRY_PARAM
        )
        
        # Valores padrão quando contexto não é detectado
        self.default_regional_context = RegionalContext(
            os.environ.get("DEFAULT_REGIONAL_CONTEXT", RegionalContext.BR.value)
        )
        self.default_tenant_id = os.environ.get("DEFAULT_TENANT_ID", "default")

# Instância singleton da configuração
context_config = ContextConfig()

class ContextManager:
    """
    Gerenciador de contexto para a aplicação.
    
    Responsável por:
    - Detectar contexto regional a partir de headers ou parâmetros
    - Detectar tenant a partir de headers ou parâmetros
    - Aplicar políticas específicas por contexto
    """
    
    def __init__(self, config: ContextConfig = context_config):
        """
        Inicializa o gerenciador de contexto.
        
        Args:
            config: Configuração para detecção de contexto
        """
        self.config = config
    
    async def get_regional_context_from_request(
        self, request: Request
    ) -> RegionalContext:
        """
        Detecta o contexto regional a partir de uma requisição.
        
        A detecção segue esta ordem:
        1. Parâmetro de query (override explícito)
        2. Header específico de contexto regional
        3. Header de país (mapeado para contexto regional)
        4. Valor padrão configurado
        
        Args:
            request: Objeto de requisição FastAPI
            
        Returns:
            Contexto regional detectado
        """
        # Prioridade 1: Parâmetro de query
        regional_context_str = request.query_params.get(self.config.regional_context_param)
        if regional_context_str:
            try:
                return RegionalContext(regional_context_str)
            except ValueError:
                logger.warning(f"Contexto regional inválido: {regional_context_str}")
        
        # Prioridade 2: Header específico
        regional_context_header = request.headers.get(self.config.regional_context_header)
        if regional_context_header:
            try:
                return RegionalContext(regional_context_header)
            except ValueError:
                logger.warning(f"Contexto regional inválido no header: {regional_context_header}")
        
        # Prioridade 3: Header de país
        country_code = request.headers.get(self.config.country_header)
        if country_code:
            if country_code.upper() in COUNTRY_CODE_TO_REGIONAL_CONTEXT:
                return COUNTRY_CODE_TO_REGIONAL_CONTEXT[country_code.upper()]
            else:
                logger.warning(f"Código de país não mapeado: {country_code}")
        
        # Prioridade 4: Parâmetro de país
        country_param = request.query_params.get(self.config.country_param)
        if country_param:
            if country_param.upper() in COUNTRY_CODE_TO_REGIONAL_CONTEXT:
                return COUNTRY_CODE_TO_REGIONAL_CONTEXT[country_param.upper()]
            else:
                logger.warning(f"Código de país não mapeado: {country_param}")
        
        # Fallback: Usar valor padrão
        return self.config.default_regional_context
    
    async def get_tenant_id_from_request(self, request: Request) -> str:
        """
        Detecta o ID do tenant a partir de uma requisição.
        
        A detecção segue esta ordem:
        1. Parâmetro de query (override explícito)
        2. Header específico de tenant
        3. Valor padrão configurado
        
        Args:
            request: Objeto de requisição FastAPI
            
        Returns:
            ID do tenant detectado
        """
        # Prioridade 1: Parâmetro de query
        tenant_id = request.query_params.get(self.config.tenant_param)
        if tenant_id:
            return tenant_id
        
        # Prioridade 2: Header específico
        tenant_header = request.headers.get(self.config.tenant_header)
        if tenant_header:
            return tenant_header
        
        # Fallback: Usar valor padrão
        return self.config.default_tenant_id
    
    def get_compliance_requirements(self, regional_context: RegionalContext) -> Dict[str, Any]:
        """
        Obtém os requisitos de compliance específicos para um contexto regional.
        
        Args:
            regional_context: Contexto regional
            
        Returns:
            Dicionário com requisitos de compliance específicos
        """
        requirements = {
            "data_retention_days": 365,  # Valor padrão: 1 ano
            "require_consent": False,
            "require_data_masking": False,
            "require_explicit_deletion": False,
            "allow_cross_border_transfer": True,
            "require_breach_notification": False,
            "max_password_age_days": 90,
            "audit_trail_required": True
        }
        
        # Aplicar requisitos específicos por contexto regional
        if regional_context == RegionalContext.BR:
            # Requisitos LGPD (Brasil)
            requirements.update({
                "require_consent": True,
                "require_data_masking": True,
                "data_retention_days": 730,  # 2 anos (requisito BACEN)
                "require_breach_notification": True,
                "max_password_age_days": 60,  # Requisito mais restritivo do BACEN
                "require_explicit_deletion": True
            })
        
        elif regional_context == RegionalContext.EU:
            # Requisitos GDPR (Europa)
            requirements.update({
                "require_consent": True,
                "require_data_masking": True,
                "data_retention_days": 395,  # 13 meses (requisito PSD2)
                "require_breach_notification": True,
                "max_password_age_days": 90,
                "require_explicit_deletion": True,
                "allow_cross_border_transfer": False  # Restrições de transferência internacional
            })
        
        elif regional_context == RegionalContext.US:
            # Requisitos PCI DSS (EUA)
            requirements.update({
                "data_retention_days": 365,  # 1 ano (PCI DSS)
                "require_data_masking": True,  # Mascaramento de PCI
                "max_password_age_days": 90,
                "audit_trail_retention_days": 365  # PCI DSS requer 1 ano
            })
        
        elif regional_context == RegionalContext.AO:
            # Requisitos BNA (Angola)
            requirements.update({
                "data_retention_days": 1095,  # 3 anos
                "require_consent": True,
                "max_password_age_days": 60,
                "audit_trail_required": True
            })
        
        return requirements

# Instância singleton do gerenciador de contexto
context_manager = ContextManager()

# Dependências para FastAPI

async def get_regional_context(
    request: Request,
    x_regional_context: Optional[str] = Header(None, alias=context_config.regional_context_header),
    x_country_code: Optional[str] = Header(None, alias=context_config.country_header),
    regional_context: Optional[str] = Query(None, alias=context_config.regional_context_param),
    country_code: Optional[str] = Query(None, alias=context_config.country_param)
) -> RegionalContext:
    """
    Dependência FastAPI para obter o contexto regional.
    
    Args:
        request: Objeto de requisição
        x_regional_context: Header de contexto regional
        x_country_code: Header de código de país
        regional_context: Parâmetro de query para contexto regional
        country_code: Parâmetro de query para código de país
        
    Returns:
        Contexto regional detectado
    """
    return await context_manager.get_regional_context_from_request(request)

async def get_tenant_context(
    request: Request,
    x_tenant_id: Optional[str] = Header(None, alias=context_config.tenant_header),
    tenant_id: Optional[str] = Query(None, alias=context_config.tenant_param)
) -> str:
    """
    Dependência FastAPI para obter o ID do tenant.
    
    Args:
        request: Objeto de requisição
        x_tenant_id: Header de ID do tenant
        tenant_id: Parâmetro de query para ID do tenant
        
    Returns:
        ID do tenant detectado
    """
    return await context_manager.get_tenant_id_from_request(request)

@lru_cache(maxsize=32)
def get_compliance_requirements_for_context(
    regional_context: RegionalContext
) -> Dict[str, Any]:
    """
    Obtém os requisitos de compliance para um contexto regional (com cache).
    
    Args:
        regional_context: Contexto regional
        
    Returns:
        Dicionário com requisitos de compliance
    """
    return context_manager.get_compliance_requirements(regional_context)