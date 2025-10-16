"""
INNOVABIZ IAM - Integrador de Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Integrador para gerenciamento de auditoria com suporte a multi-contexto
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA
"""

import uuid
import structlog
from typing import Dict, List, Optional, Any, Union
from datetime import datetime
from fastapi import Depends, Request

from app.models.audit import (
    AuditEventCreate, 
    AuditEventCategory, 
    AuditEventSeverity,
    AuditHttpDetails,
    ComplianceFramework,
    RegionalCompliance
)
from app.services.audit_service import AuditService, get_audit_service

# Logger estruturado
logger = structlog.get_logger(__name__)

# Mapeamento de contextos regionais para frameworks de compliance
REGIONAL_COMPLIANCE_FRAMEWORKS = {
    "BR": [ComplianceFramework.LGPD, ComplianceFramework.BACEN, ComplianceFramework.PCI_DSS],
    "AO": [ComplianceFramework.BNA, ComplianceFramework.PCI_DSS],
    "EU": [ComplianceFramework.GDPR, ComplianceFramework.PSD2, ComplianceFramework.PCI_DSS],
    "US": [ComplianceFramework.PCI_DSS, ComplianceFramework.SOX],
    "DEFAULT": [ComplianceFramework.PCI_DSS],
}

# Períodos de retenção por região (em dias)
REGIONAL_RETENTION_PERIODS = {
    "BR": 730,   # 2 anos (LGPD)
    "AO": 1825,  # 5 anos (BNA)
    "EU": 365,   # 1 ano (GDPR)
    "US": 2555,  # 7 anos (SOX)
    "DEFAULT": 365,
}

# Campos sensíveis por tipo de recurso que devem ser mascarados
SENSITIVE_FIELDS_BY_RESOURCE = {
    "user": ["password", "document_number", "phone_number", "email"],
    "payment": ["card_number", "cvv", "account_number"],
    "address": ["street_address", "postal_code"],
    "document": ["document_number", "id_number", "passport_number"],
    "DEFAULT": []
}


class AuditIntegrator:
    """
    Integrador para gerenciamento de auditoria com suporte a multi-contexto.
    
    Responsabilidades:
    - Gerenciamento de contexto regional e tenant
    - Aplicação de políticas de compliance baseadas no contexto
    - Enriquecimento de eventos com informações de compliance
    - Mascaramento de dados sensíveis
    - Geração de eventos de auditoria estruturados
    """
    
    def __init__(
        self,
        audit_service: AuditService,
        tenant_id: Optional[str] = None,
        regional_context: Optional[str] = None,
        country_code: Optional[str] = None,
        language: Optional[str] = None,
        correlation_id: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None
    ):
        """
        Inicializa o integrador de auditoria.
        
        Args:
            audit_service: Serviço de auditoria para registro de eventos
            tenant_id: ID do tenant atual (isolamento multi-tenant)
            regional_context: Contexto regional (BR, US, EU, AO)
            country_code: Código ISO do país
            language: Código de idioma (ex: pt-BR, en-US)
            correlation_id: ID de correlação para eventos relacionados
            user_id: ID do usuário atual
            user_name: Nome do usuário atual
        """
        self.audit_service = audit_service
        self.tenant_id = tenant_id
        self.regional_context = regional_context or "DEFAULT"
        self.country_code = country_code
        self.language = language
        self.correlation_id = correlation_id or str(uuid.uuid4())
        self.user_id = user_id
        self.user_name = user_name
        
        # Inicializar informações de compliance
        self.compliance_info = self._initialize_compliance_info()
        
        logger.debug(
            "AuditIntegrator inicializado",
            tenant_id=self.tenant_id,
            regional_context=self.regional_context,
            correlation_id=self.correlation_id,
            user_id=self.user_id
        )
    
    def _initialize_compliance_info(self) -> RegionalCompliance:
        """
        Inicializa as informações de compliance baseadas no contexto regional.
        """
        frameworks = REGIONAL_COMPLIANCE_FRAMEWORKS.get(
            self.regional_context, 
            REGIONAL_COMPLIANCE_FRAMEWORKS["DEFAULT"]
        )
        
        data_retention = REGIONAL_RETENTION_PERIODS.get(
            self.regional_context, 
            REGIONAL_RETENTION_PERIODS["DEFAULT"]
        )
        
        required_fields = ["tenant_id", "regional_context"]
        if "GDPR" in frameworks or "LGPD" in frameworks:
            required_fields.extend(["user_id", "action", "resource_type"])
        
        return RegionalCompliance(
            frameworks=frameworks,
            data_residency=self.regional_context,
            data_retention=data_retention,
            required_fields=required_fields,
            sensitive_fields=[]
        )
    
    async def create_event(
        self,
        category: AuditEventCategory,
        action: str,
        description: str,
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        resource_name: Optional[str] = None,
        severity: AuditEventSeverity = AuditEventSeverity.INFO,
        success: bool = True,
        error_message: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
        tags: Optional[List[str]] = None,
        correlation_id: Optional[str] = None,
        http_details: Optional[AuditHttpDetails] = None,
        source_ip: Optional[str] = None,
        source_system: Optional[str] = None,
    ) -> str:
        """
        Cria um evento de auditoria com o contexto atual.
        
        Args:
            category: Categoria do evento
            action: Ação realizada
            description: Descrição detalhada
            resource_type: Tipo do recurso afetado
            resource_id: ID do recurso afetado
            resource_name: Nome do recurso afetado
            severity: Nível de severidade
            success: Indica se a ação foi bem-sucedida
            error_message: Mensagem de erro se a ação falhou
            details: Detalhes adicionais específicos do evento
            tags: Tags para classificação e busca
            correlation_id: ID de correlação (substitui o atual se fornecido)
            http_details: Detalhes HTTP (para eventos de API)
            source_ip: IP de origem da ação
            source_system: Sistema de origem
            
        Returns:
            str: ID do evento criado
        """
        # Mascarar dados sensíveis nos detalhes, se aplicável
        masked_details = None
        if details:
            masked_details = self._mask_sensitive_data(details, resource_type)
        
        # Determinar compliance tags baseadas na ação e recurso
        compliance_tags = self._determine_compliance_tags(
            action, resource_type, category
        )
        
        # Criar evento base
        audit_event = AuditEventCreate(
            category=category,
            action=action,
            description=description,
            resource_type=resource_type,
            resource_id=resource_id,
            resource_name=resource_name,
            severity=severity,
            success=success,
            error_message=error_message,
            details=masked_details,
            tags=tags or [],
            tenant_id=self.tenant_id,
            regional_context=self.regional_context,
            country_code=self.country_code,
            language=self.language,
            user_id=self.user_id,
            user_name=self.user_name,
            correlation_id=correlation_id or self.correlation_id,
            http_details=http_details,
            source_ip=source_ip,
            source_system=source_system,
            compliance=self.compliance_info,
            compliance_tags=compliance_tags
        )
        
        # Validar o evento de acordo com as regras de compliance
        self._validate_compliance(audit_event)
        
        # Enviar evento para o serviço de auditoria
        event_id = await self.audit_service.create_event(audit_event)
        
        logger.debug(
            "Evento de auditoria criado",
            event_id=event_id,
            category=category,
            action=action,
            tenant_id=self.tenant_id,
            regional_context=self.regional_context,
            correlation_id=correlation_id or self.correlation_id
        )
        
        return event_id
    
    def _mask_sensitive_data(
        self, 
        data: Dict[str, Any], 
        resource_type: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Mascara dados sensíveis com base no tipo de recurso.
        
        Args:
            data: Dados a serem mascarados
            resource_type: Tipo do recurso para determinar campos sensíveis
            
        Returns:
            Dict[str, Any]: Dados com campos sensíveis mascarados
        """
        if not resource_type:
            return data
            
        # Obter campos sensíveis para o tipo de recurso
        sensitive_fields = SENSITIVE_FIELDS_BY_RESOURCE.get(
            resource_type.lower(),
            SENSITIVE_FIELDS_BY_RESOURCE["DEFAULT"]
        )
        
        if not sensitive_fields:
            return data
            
        # Criar uma cópia para não modificar o original
        masked_data = data.copy()
        
        # Mascarar campos sensíveis
        for field in sensitive_fields:
            if field in masked_data:
                value = masked_data[field]
                if isinstance(value, str):
                    if len(value) > 6:
                        # Preservar os primeiros 2 e últimos 2 caracteres
                        masked_data[field] = value[:2] + "****" + value[-2:]
                    else:
                        masked_data[field] = "****"
        
        return masked_data
    
    def _determine_compliance_tags(
        self,
        action: str,
        resource_type: Optional[str] = None,
        category: AuditEventCategory = AuditEventCategory.OTHER
    ) -> List[str]:
        """
        Determina tags de compliance com base na ação e tipo de recurso.
        
        Args:
            action: Ação realizada
            resource_type: Tipo do recurso
            category: Categoria do evento
            
        Returns:
            List[str]: Lista de tags de compliance
        """
        tags = []
        
        # Tags baseadas no tipo de recurso
        if resource_type:
            resource_lower = resource_type.lower()
            
            if "user" in resource_lower or "profile" in resource_lower or "customer" in resource_lower:
                tags.append("PII")  # Personally Identifiable Information
                
            if "payment" in resource_lower or "card" in resource_lower or "credit" in resource_lower:
                tags.append("PCI")  # Payment Card Industry
                
            if "health" in resource_lower or "medical" in resource_lower:
                tags.append("PHI")  # Protected Health Information
                
            if "document" in resource_lower or "kyc" in resource_lower:
                tags.append("KYC")  # Know Your Customer
        
        # Tags baseadas na categoria
        if category == AuditEventCategory.PAYMENT or category == AuditEventCategory.CARD_DATA:
            tags.append("PCI")
            
        if category == AuditEventCategory.AUTHENTICATION or category == AuditEventCategory.USER_MANAGEMENT:
            tags.append("IAM")  # Identity and Access Management
            
        if category == AuditEventCategory.PRIVACY or category == AuditEventCategory.CONSENT:
            tags.append("CONSENT")
            
        # Tags baseadas na ação
        action_lower = action.lower()
        
        if "delete" in action_lower or "remove" in action_lower:
            tags.append("DATA_DELETION")
            
        if "export" in action_lower or "download" in action_lower:
            tags.append("DATA_EXPORT")
            
        if "consent" in action_lower:
            tags.append("CONSENT")
        
        # Adicionar tags específicas de frameworks se aplicável
        if self.regional_context == "BR":
            tags.append("LGPD")
        elif self.regional_context == "EU":
            tags.append("GDPR")
        elif self.regional_context == "AO":
            tags.append("BNA")
        elif self.regional_context == "US":
            if "payment" in action_lower or "card" in action_lower:
                tags.append("SOX")
        
        return tags
    
    def _validate_compliance(self, event: AuditEventCreate) -> None:
        """
        Valida o evento de acordo com as regras de compliance.
        Lança exceção se o evento não estiver em conformidade.
        
        Args:
            event: Evento de auditoria a ser validado
            
        Raises:
            ValueError: Se o evento não estiver em conformidade
        """
        if not event.compliance:
            return
            
        # Validar campos obrigatórios
        for field in event.compliance.required_fields:
            value = getattr(event, field, None)
            if value is None or (isinstance(value, str) and value.strip() == ""):
                raise ValueError(
                    f"Campo obrigatório '{field}' não fornecido para "
                    f"evento de auditoria (frameworks: {event.compliance.frameworks})"
                )

        # Validações específicas por framework
        frameworks = event.compliance.frameworks
        
        # GDPR ou LGPD requerem usuário identificado ou justificativa
        if (ComplianceFramework.GDPR in frameworks or 
            ComplianceFramework.LGPD in frameworks):
            if not event.user_id and not event.tags:
                if "ANONYMOUS" not in event.tags:
                    event.tags.append("ANONYMOUS")
                    logger.warning(
                        "Evento de auditoria sem user_id identificado",
                        action=event.action,
                        category=event.category
                    )

        # PCI DSS requer mascaramento de dados sensíveis
        if ComplianceFramework.PCI_DSS in frameworks:
            if event.category == AuditEventCategory.CARD_DATA:
                if event.details and any(
                    key in str(event.details).lower() 
                    for key in ["card", "cvv", "pan", "account"]
                ):
                    logger.warning(
                        "Possíveis dados sensíveis em evento PCI",
                        action=event.action,
                        category=event.category
                    )


class CurrentAuditContext:
    """
    Classe para extrair o contexto atual de auditoria da requisição.
    Utilizada como dependência FastAPI.
    """
    
    def __init__(
        self,
        request: Request = None,
        tenant_id: Optional[str] = None,
        regional_context: Optional[str] = None,
        language: Optional[str] = None
    ):
        """
        Inicializa o contexto de auditoria.
        
        Args:
            request: A requisição HTTP atual (opcional)
            tenant_id: ID do tenant (substitui o da requisição se fornecido)
            regional_context: Contexto regional (substitui o da requisição se fornecido)
            language: Código de idioma (substitui o da requisição se fornecido)
        """
        # Inicializar a partir do request se disponível
        self._request = request
        self._tenant_id = tenant_id
        self._regional_context = regional_context
        self._language = language
        self._correlation_id = None
        self._user_id = None
        self._user_name = None
        
        # Extrair do request se disponível e valores não fornecidos
        if request:
            if not self._tenant_id and hasattr(request.state, "tenant_id"):
                self._tenant_id = request.state.tenant_id
                
            if not self._regional_context and hasattr(request.state, "regional_context"):
                self._regional_context = request.state.regional_context
                
            if not self._language and hasattr(request.state, "language"):
                self._language = request.state.language
                
            if hasattr(request.state, "correlation_id"):
                self._correlation_id = request.state.correlation_id
                
            if hasattr(request.state, "user_id"):
                self._user_id = request.state.user_id
                
            if hasattr(request.state, "user_name"):
                self._user_name = request.state.user_name
    
    @property
    def tenant_id(self) -> Optional[str]:
        return self._tenant_id
        
    @property
    def regional_context(self) -> Optional[str]:
        return self._regional_context
        
    @property
    def language(self) -> Optional[str]:
        return self._language
        
    @property
    def correlation_id(self) -> Optional[str]:
        return self._correlation_id
        
    @property
    def user_id(self) -> Optional[str]:
        return self._user_id
        
    @property
    def user_name(self) -> Optional[str]:
        return self._user_name
        
    def with_tenant(self, tenant_id: str) -> 'CurrentAuditContext':
        """Retorna uma nova instância com o tenant_id especificado."""
        return CurrentAuditContext(
            request=self._request,
            tenant_id=tenant_id,
            regional_context=self._regional_context,
            language=self._language
        )
        
    def with_regional_context(self, regional_context: str) -> 'CurrentAuditContext':
        """Retorna uma nova instância com o regional_context especificado."""
        return CurrentAuditContext(
            request=self._request,
            tenant_id=self._tenant_id,
            regional_context=regional_context,
            language=self._language
        )
        
    def with_language(self, language: str) -> 'CurrentAuditContext':
        """Retorna uma nova instância com o language especificado."""
        return CurrentAuditContext(
            request=self._request,
            tenant_id=self._tenant_id,
            regional_context=self._regional_context,
            language=language
        )


async def get_audit_context(
    request: Request,
    tenant_id: Optional[str] = None,
    regional_context: Optional[str] = None,
    language: Optional[str] = None
) -> CurrentAuditContext:
    """
    Dependência FastAPI para obter o contexto atual de auditoria.
    
    Args:
        request: A requisição HTTP atual
        tenant_id: ID do tenant (opcional, substitui o da requisição)
        regional_context: Contexto regional (opcional, substitui o da requisição)
        language: Código de idioma (opcional, substitui o da requisição)
        
    Returns:
        CurrentAuditContext: O contexto atual de auditoria
    """
    return CurrentAuditContext(
        request=request,
        tenant_id=tenant_id,
        regional_context=regional_context,
        language=language
    )


async def get_audit_integrator(
    audit_service: AuditService = Depends(get_audit_service),
    audit_context: CurrentAuditContext = Depends(get_audit_context)
) -> AuditIntegrator:
    """
    Dependência FastAPI para obter um integrador de auditoria.
    
    Args:
        audit_service: O serviço de auditoria
        audit_context: O contexto atual de auditoria
        
    Returns:
        AuditIntegrator: Um integrador de auditoria configurado
    """
    return AuditIntegrator(
        audit_service=audit_service,
        tenant_id=audit_context.tenant_id,
        regional_context=audit_context.regional_context,
        language=audit_context.language,
        correlation_id=audit_context.correlation_id,
        user_id=audit_context.user_id,
        user_name=audit_context.user_name
    )