"""
INNOVABIZ IAM - Audit Context Integrator
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Integrador entre sistema de auditoria e contexto multi-dimensional
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA

Este módulo implementa a integração entre o sistema de auditoria e o módulo de contexto,
garantindo que os eventos de auditoria sejam enriquecidos com informações contextuais
e que as políticas de conformidade sejam aplicadas de acordo com o contexto regional.
"""

import logging
import datetime
from typing import Dict, Any, Optional, List, Set
from pydantic import BaseModel
from fastapi import Request, Depends, HTTPException, status

from .context import context_manager, get_regional_context, get_tenant_context
from .observability import logger, ObservabilityContext
from ..models.audit import (
    AuditEventCreate, 
    AuditEventSeverity,
    AuditEventCategory,
    RegionalContext,
    ComplianceFramework
)
from ..services.audit_service import AuditService

class AuditContextIntegrator:
    """
    Integrador entre o sistema de auditoria e o contexto multi-dimensional.
    
    Responsabilidades:
    1. Enriquecer eventos de auditoria com informações contextuais (tenant, região)
    2. Aplicar regras de compliance específicas por contexto regional
    3. Ajustar políticas de retenção, mascaramento e severidade conforme o contexto
    4. Aplicar tags de compliance automaticamente com base no contexto e categoria
    """
    
    def __init__(self, audit_service: AuditService):
        """
        Inicializa o integrador de auditoria e contexto.
        
        Args:
            audit_service: Serviço de auditoria a ser utilizado
        """
        self.audit_service = audit_service
        self.logger = logger.bind(module="audit_context_integrator")
    
    async def create_audit_event_with_context(
        self,
        event: AuditEventCreate,
        regional_context: RegionalContext,
        tenant_id: str,
        observability_context: Optional[ObservabilityContext] = None
    ) -> Dict[str, Any]:
        """
        Cria um evento de auditoria enriquecido com informações contextuais.
        
        Args:
            event: Evento de auditoria base a ser criado
            regional_context: Contexto regional detectado
            tenant_id: ID do tenant detectado
            observability_context: Contexto de observabilidade opcional
            
        Returns:
            Evento de auditoria criado e enriquecido
        """
        # Enriquece o evento com o contexto regional e tenant
        if not event.tenant_id:
            event.tenant_id = tenant_id
            
        if not event.regional_context:
            event.regional_context = regional_context
            
        # Aplica tags de compliance automaticamente com base no contexto
        compliance_tags = self.get_compliance_tags_for_context(event, regional_context)
        if compliance_tags:
            if not event.compliance_frameworks:
                event.compliance_frameworks = []
            event.compliance_frameworks.extend(compliance_tags)
            
            # Remove duplicatas preservando a ordem
            unique_frameworks = []
            seen = set()
            for framework in event.compliance_frameworks:
                if framework not in seen:
                    seen.add(framework)
                    unique_frameworks.append(framework)
            event.compliance_frameworks = unique_frameworks
        
        # Ajusta a severidade conforme o contexto se necessário
        event.severity = self.adjust_severity_for_context(event, regional_context)
        
        # Aplica políticas de mascaramento de dados sensíveis se necessário
        event.details = self.apply_data_masking_if_needed(event.details, regional_context)
        
        # Registra a criação do evento no log estruturado
        log_context = {"tenant_id": tenant_id, "regional_context": regional_context.value}
        if observability_context:
            log_context.update(observability_context.to_dict())
            
        self.logger.info(
            "Criando evento de auditoria com contexto", 
            event_id=event.event_id,
            category=event.category.value,
            severity=event.severity.value,
            user_id=event.user_id,
            compliance_frameworks=[f.value for f in event.compliance_frameworks] if event.compliance_frameworks else None,
            **log_context
        )
        
        # Cria o evento de auditoria através do serviço
        return await self.audit_service.create_audit_event(event)
    
    def get_compliance_tags_for_context(
        self,
        event: AuditEventCreate,
        regional_context: RegionalContext
    ) -> List[ComplianceFramework]:
        """
        Determina as tags de compliance aplicáveis com base no contexto regional
        e na categoria do evento de auditoria.
        
        Args:
            event: Evento de auditoria
            regional_context: Contexto regional
            
        Returns:
            Lista de frameworks de compliance aplicáveis
        """
        compliance_tags = []
        
        # Tags de compliance baseadas no contexto regional
        if regional_context == RegionalContext.BR:
            compliance_tags.append(ComplianceFramework.LGPD)
            compliance_tags.append(ComplianceFramework.BACEN)
            
            # Requisitos específicos para determinadas categorias
            if event.category in [
                AuditEventCategory.AUTHENTICATION,
                AuditEventCategory.AUTHORIZATION,
                AuditEventCategory.USER_MANAGEMENT
            ]:
                compliance_tags.append(ComplianceFramework.PCI_DSS)
                
        elif regional_context == RegionalContext.EU:
            compliance_tags.append(ComplianceFramework.GDPR)
            compliance_tags.append(ComplianceFramework.PSD2)
            
            # Requisitos específicos para determinadas categorias
            if event.category in [
                AuditEventCategory.PAYMENT_PROCESSING,
                AuditEventCategory.FINANCIAL_TRANSACTION
            ]:
                compliance_tags.append(ComplianceFramework.PCI_DSS)
                
        elif regional_context == RegionalContext.US:
            # Nos EUA, aplicamos PCI DSS para todas as categorias relacionadas a dados sensíveis
            if event.category in [
                AuditEventCategory.AUTHENTICATION,
                AuditEventCategory.AUTHORIZATION,
                AuditEventCategory.USER_MANAGEMENT,
                AuditEventCategory.PAYMENT_PROCESSING,
                AuditEventCategory.FINANCIAL_TRANSACTION,
                AuditEventCategory.DATA_ACCESS
            ]:
                compliance_tags.append(ComplianceFramework.PCI_DSS)
                
        elif regional_context == RegionalContext.AO:
            compliance_tags.append(ComplianceFramework.BNA)
            
            # Angola também pode requerer PCI DSS em alguns casos
            if event.category in [
                AuditEventCategory.PAYMENT_PROCESSING,
                AuditEventCategory.FINANCIAL_TRANSACTION
            ]:
                compliance_tags.append(ComplianceFramework.PCI_DSS)
        
        # PCI DSS se aplica globalmente a eventos de pagamento
        if event.category in [
            AuditEventCategory.PAYMENT_PROCESSING,
            AuditEventCategory.FINANCIAL_TRANSACTION,
            AuditEventCategory.CARD_MANAGEMENT
        ]:
            if ComplianceFramework.PCI_DSS not in compliance_tags:
                compliance_tags.append(ComplianceFramework.PCI_DSS)
        
        return compliance_tags
    
    def adjust_severity_for_context(
        self,
        event: AuditEventCreate,
        regional_context: RegionalContext
    ) -> AuditEventSeverity:
        """
        Ajusta a severidade do evento com base nas políticas específicas do contexto regional.
        
        Args:
            event: Evento de auditoria
            regional_context: Contexto regional
            
        Returns:
            Severidade ajustada
        """
        severity = event.severity
        
        # Regras específicas por contexto regional
        if regional_context == RegionalContext.BR:
            # LGPD e BACEN têm requisitos mais rigorosos para eventos de dados pessoais
            if event.category in [
                AuditEventCategory.DATA_ACCESS,
                AuditEventCategory.PERSONAL_DATA_MANAGEMENT
            ]:
                # Aumenta a severidade para eventos relacionados a dados pessoais
                if severity == AuditEventSeverity.LOW:
                    severity = AuditEventSeverity.MEDIUM
                elif severity == AuditEventSeverity.MEDIUM:
                    severity = AuditEventSeverity.HIGH
            
        elif regional_context == RegionalContext.EU:
            # GDPR tem requisitos muito rigorosos para eventos de dados pessoais
            if event.category in [
                AuditEventCategory.DATA_ACCESS,
                AuditEventCategory.PERSONAL_DATA_MANAGEMENT,
                AuditEventCategory.CONSENT_MANAGEMENT
            ]:
                # Aumenta a severidade para eventos relacionados a dados pessoais
                if severity == AuditEventSeverity.LOW:
                    severity = AuditEventSeverity.MEDIUM
                elif severity == AuditEventSeverity.MEDIUM:
                    severity = AuditEventSeverity.HIGH
                elif severity == AuditEventSeverity.HIGH:
                    severity = AuditEventSeverity.CRITICAL
            
        elif regional_context == RegionalContext.US:
            # PCI DSS nos EUA tem requisitos rigorosos para eventos financeiros
            if event.category in [
                AuditEventCategory.PAYMENT_PROCESSING,
                AuditEventCategory.FINANCIAL_TRANSACTION,
                AuditEventCategory.CARD_MANAGEMENT
            ]:
                # Aumenta a severidade para eventos relacionados a pagamentos
                if severity == AuditEventSeverity.LOW:
                    severity = AuditEventSeverity.MEDIUM
        
        # Regras globais independentes do contexto regional
        
        # Falhas de autenticação são sempre pelo menos médias
        if event.category == AuditEventCategory.AUTHENTICATION and "failed" in event.action.lower():
            if severity == AuditEventSeverity.LOW:
                severity = AuditEventSeverity.MEDIUM
        
        # Alterações de permissões são sempre pelo menos médias
        if event.category == AuditEventCategory.AUTHORIZATION and "permission" in event.action.lower():
            if severity == AuditEventSeverity.LOW:
                severity = AuditEventSeverity.MEDIUM
                
        # Modificações em configurações de segurança são sempre pelo menos altas
        if event.category == AuditEventCategory.CONFIGURATION and "security" in event.action.lower():
            if severity in [AuditEventSeverity.LOW, AuditEventSeverity.MEDIUM]:
                severity = AuditEventSeverity.HIGH
        
        return severity
    
    def apply_data_masking_if_needed(
        self,
        details: Dict[str, Any],
        regional_context: RegionalContext
    ) -> Dict[str, Any]:
        """
        Aplica mascaramento de dados sensíveis conforme as políticas do contexto regional.
        
        Args:
            details: Detalhes do evento de auditoria
            regional_context: Contexto regional
            
        Returns:
            Detalhes com dados mascarados se necessário
        """
        if not details:
            return details
        
        # Clone os detalhes para não modificar o objeto original
        masked_details = details.copy()
        
        # Obtém políticas de compliance para o contexto regional
        compliance_requirements = context_manager.get_compliance_requirements(regional_context)
        
        # Aplica mascaramento apenas se requerido pelo contexto
        if compliance_requirements.get("require_data_masking", False):
            # Campos sensíveis que sempre devem ser mascarados
            sensitive_fields = {
                "password", "senha", "secret", "token", "api_key", "private_key",
                "credit_card", "card_number", "cvv", "cvc", "card_security_code",
                "cpf", "cnpj", "ssn", "tax_id", "national_id",
                "passport", "id_number", "birth_date", "date_of_birth",
                "address", "email", "phone", "telefone", "celular"
            }
            
            # Percorre os campos recursivamente
            def mask_sensitive_data(data, path=""):
                if isinstance(data, dict):
                    result = {}
                    for key, value in data.items():
                        # Verifica se o campo é sensível
                        is_sensitive = any(
                            sensitive_field in key.lower()
                            for sensitive_field in sensitive_fields
                        )
                        
                        if is_sensitive and isinstance(value, str):
                            # Aplica mascaramento adequado conforme o tipo de dado
                            if "card" in key.lower() or "credit" in key.lower():
                                # Formato cartão: mantém primeiros 6 e últimos 4 dígitos
                                if len(value) > 10:
                                    result[key] = value[:6] + "******" + value[-4:]
                                else:
                                    result[key] = "******"
                            elif "cpf" in key.lower() or "cnpj" in key.lower() or "tax" in key.lower():
                                # Formato documento: mantém primeiros 3 e últimos 2 dígitos
                                if len(value) > 5:
                                    result[key] = value[:3] + "****" + value[-2:]
                                else:
                                    result[key] = "****"
                            elif "email" in key.lower():
                                # Formato email: mantém primeiro caractere e domínio
                                if "@" in value:
                                    username, domain = value.split("@", 1)
                                    if len(username) > 1:
                                        result[key] = username[0] + "***@" + domain
                                    else:
                                        result[key] = "***@" + domain
                                else:
                                    result[key] = "******"
                            elif any(field in key.lower() for field in ["password", "senha", "secret", "token", "key"]):
                                # Senhas e segredos: mascara completamente
                                result[key] = "**********"
                            else:
                                # Outros dados sensíveis: mascara parcialmente
                                if len(value) > 4:
                                    visible_chars = min(len(value) // 4, 3)
                                    result[key] = value[:visible_chars] + "*" * (len(value) - visible_chars)
                                else:
                                    result[key] = "****"
                        elif isinstance(value, dict) or isinstance(value, list):
                            # Processa recursivamente
                            result[key] = mask_sensitive_data(value, path + "." + key if path else key)
                        else:
                            # Mantém valores não sensíveis
                            result[key] = value
                    return result
                
                elif isinstance(data, list):
                    return [mask_sensitive_data(item, path + "[]") for item in data]
                
                else:
                    # Tipos primitivos não são mascarados
                    return data
            
            # Aplica o mascaramento recursivamente
            masked_details = mask_sensitive_data(masked_details)
        
        return masked_details


# Factory para criar o integrador
audit_context_integrator = None

def get_audit_context_integrator(audit_service: AuditService = None) -> AuditContextIntegrator:
    """
    Factory para obter a instância do integrador de auditoria e contexto.
    
    Args:
        audit_service: Serviço de auditoria a ser utilizado
        
    Returns:
        Instância do integrador
    """
    global audit_context_integrator
    
    if audit_context_integrator is None and audit_service is not None:
        audit_context_integrator = AuditContextIntegrator(audit_service)
    
    if audit_context_integrator is None:
        raise ValueError("AuditContextIntegrator não foi inicializado corretamente")
    
    return audit_context_integrator


# Função de inicialização para ser chamada na startup da aplicação
def init_audit_context_integrator(audit_service: AuditService) -> None:
    """
    Inicializa o integrador de auditoria e contexto.
    
    Args:
        audit_service: Serviço de auditoria a ser utilizado
    """
    global audit_context_integrator
    audit_context_integrator = AuditContextIntegrator(audit_service)
    logger.info("AuditContextIntegrator inicializado com sucesso")


# Dependência FastAPI para obter o integrador em rotas
async def get_audit_context_integrator_dependency(
    regional_context: RegionalContext = Depends(get_regional_context),
    tenant_id: str = Depends(get_tenant_context)
) -> AuditContextIntegrator:
    """
    Dependência FastAPI para obter o integrador de auditoria e contexto.
    
    Args:
        regional_context: Contexto regional detectado
        tenant_id: ID do tenant detectado
        
    Returns:
        Instância do integrador
    """
    integrator = get_audit_context_integrator()
    return integrator