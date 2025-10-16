"""
INNOVABIZ IAM - Modelo de Banco de Dados para Auditoria
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Modelos SQLAlchemy para persistência de eventos de auditoria com suporte a multi-contexto
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA, SOX
"""

import uuid
from datetime import datetime
from typing import List, Dict, Any, Optional

from sqlalchemy import Column, String, Boolean, DateTime, Integer, Text, JSON, ForeignKey, Index
from sqlalchemy.dialects.postgresql import JSONB, UUID, ARRAY
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship

from app.db.base import Base
from app.models.audit import AuditEventCategory, AuditEventSeverity, ComplianceFramework


class AuditEventEntity(Base):
    """
    Entidade de banco de dados para eventos de auditoria.
    
    Suporta multi-contexto (tenant, regional, idioma, moeda) e requisitos de compliance.
    Implementa estrutura otimizada para consultas e armazenamento eficiente de eventos.
    """
    __tablename__ = "audit_events"
    
    # Identificação principal
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    
    # Contexto multi-tenant e regional
    tenant_id = Column(String(50), nullable=False, index=True)
    regional_context = Column(String(10), nullable=True, index=True)
    country_code = Column(String(2), nullable=True, index=True)
    language = Column(String(5), nullable=True)
    
    # Informações do evento
    category = Column(String(50), nullable=False, index=True)
    action = Column(String(100), nullable=False, index=True)
    description = Column(Text, nullable=False)
    resource_type = Column(String(100), nullable=True, index=True)
    resource_id = Column(String(100), nullable=True, index=True)
    resource_name = Column(String(255), nullable=True)
    
    # Severidade e resultado
    severity = Column(String(20), nullable=False, index=True)
    success = Column(Boolean, nullable=False, default=True, index=True)
    error_message = Column(Text, nullable=True)
    
    # Detalhes e tags
    details = Column(JSONB, nullable=True)
    tags = Column(ARRAY(String), nullable=False, default=[])
    
    # Informações do usuário
    user_id = Column(String(50), nullable=True, index=True)
    user_name = Column(String(255), nullable=True)
    
    # Correlação e origem
    correlation_id = Column(String(100), nullable=False, index=True)
    source_ip = Column(String(45), nullable=True)  # Suporte para IPv6
    source_system = Column(String(100), nullable=True)
    
    # Detalhes HTTP (opcional)
    http_details = Column(JSONB, nullable=True)
    
    # Informações de compliance
    compliance = Column(JSONB, nullable=True)
    compliance_tags = Column(ARRAY(String), nullable=False, default=[])
    
    # Timestamps
    timestamp = Column(DateTime(timezone=True), nullable=False, index=True, default=func.now())
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    
    # Metadados do evento para rastreabilidade interna
    version = Column(Integer, nullable=False, default=1)
    is_anonymized = Column(Boolean, nullable=False, default=False)
    anonymized_at = Column(DateTime(timezone=True), nullable=True)
    
    # Particionamento virtual
    partition_key = Column(String(50), nullable=False, index=True)
    
    def __init__(self, **kwargs):
        """
        Inicializa um novo evento de auditoria.
        
        Define automaticamente a partition_key com base no tenant_id e mês/ano
        para otimizar consultas e particionamento.
        """
        super().__init__(**kwargs)
        
        # Gerar partition_key se não fornecido (usado para particionamento virtual)
        if not self.partition_key and self.tenant_id and self.timestamp:
            # Formato: tenant_id:YYYY-MM
            month_year = self.timestamp.strftime("%Y-%m")
            self.partition_key = f"{self.tenant_id}:{month_year}"
        elif not self.partition_key and self.tenant_id:
            # Se timestamp não está disponível, usar data atual
            month_year = datetime.utcnow().strftime("%Y-%m")
            self.partition_key = f"{self.tenant_id}:{month_year}"


# Definir índices para otimização de consultas comuns
Index('idx_audit_tenant_timestamp', AuditEventEntity.tenant_id, AuditEventEntity.timestamp.desc())
Index('idx_audit_tenant_category_timestamp', 
      AuditEventEntity.tenant_id, AuditEventEntity.category, AuditEventEntity.timestamp.desc())
Index('idx_audit_tenant_resource', 
      AuditEventEntity.tenant_id, AuditEventEntity.resource_type, AuditEventEntity.resource_id)
Index('idx_audit_tenant_user_timestamp', 
      AuditEventEntity.tenant_id, AuditEventEntity.user_id, AuditEventEntity.timestamp.desc())
Index('idx_audit_tenant_correlation', 
      AuditEventEntity.tenant_id, AuditEventEntity.correlation_id)
Index('idx_audit_partition_timestamp', 
      AuditEventEntity.partition_key, AuditEventEntity.timestamp.desc())
Index('idx_audit_compliance_success', 
      AuditEventEntity.tenant_id, AuditEventEntity.success, 
      postgresql_where=(AuditEventEntity.compliance.isnot(None)))


class AuditRetentionPolicy(Base):
    """
    Políticas de retenção de dados de auditoria por tenant e framework de compliance.
    """
    __tablename__ = "audit_retention_policies"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    tenant_id = Column(String(50), nullable=False, index=True)
    regional_context = Column(String(10), nullable=True, index=True)
    framework = Column(String(50), nullable=False, index=True)
    
    # Período de retenção em dias
    retention_period_days = Column(Integer, nullable=False)
    
    # Regras específicas de anonimização
    anonymize_after_days = Column(Integer, nullable=True)
    fields_to_anonymize = Column(ARRAY(String), nullable=True)
    
    # Metadados
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at = Column(DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now())
    created_by = Column(String(50), nullable=True)
    updated_by = Column(String(50), nullable=True)
    
    # Restrição única para evitar políticas duplicadas
    __table_args__ = (
        Index('idx_unique_retention_policy', 
              tenant_id, regional_context, framework, 
              unique=True, postgresql_where=(regional_context.isnot(None))),
        Index('idx_unique_retention_policy_global', 
              tenant_id, framework, 
              unique=True, postgresql_where=(regional_context.is_(None))),
    )


class AuditComplianceReport(Base):
    """
    Relatórios de compliance gerados a partir dos dados de auditoria.
    """
    __tablename__ = "audit_compliance_reports"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    tenant_id = Column(String(50), nullable=False, index=True)
    regional_context = Column(String(10), nullable=True, index=True)
    framework = Column(String(50), nullable=False, index=True)
    
    # Período do relatório
    start_date = Column(DateTime(timezone=True), nullable=False)
    end_date = Column(DateTime(timezone=True), nullable=False)
    
    # Resumo do relatório
    total_events = Column(Integer, nullable=False, default=0)
    critical_events = Column(Integer, nullable=False, default=0)
    failure_events = Column(Integer, nullable=False, default=0)
    status = Column(String(50), nullable=False, index=True)
    
    # Detalhes completos do relatório
    categories_count = Column(JSONB, nullable=True)
    details = Column(JSONB, nullable=True)
    
    # Metadados
    generated_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    data_up_to = Column(DateTime(timezone=True), nullable=False)
    generated_by = Column(String(50), nullable=True)
    
    # Índices para consulta eficiente
    __table_args__ = (
        Index('idx_compliance_report_tenant_framework', 
              tenant_id, framework, generated_at.desc()),
        Index('idx_compliance_report_tenant_status', 
              tenant_id, status, generated_at.desc()),
    )


class AuditStatistics(Base):
    """
    Estatísticas agregadas de eventos de auditoria para otimizar consultas de dashboard.
    
    Esta tabela é atualizada periodicamente por um job agendado.
    """
    __tablename__ = "audit_statistics"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    tenant_id = Column(String(50), nullable=False, index=True)
    regional_context = Column(String(10), nullable=True, index=True)
    
    # Período da estatística
    period_type = Column(String(20), nullable=False, index=True)  # daily, weekly, monthly
    period_start = Column(DateTime(timezone=True), nullable=False, index=True)
    period_end = Column(DateTime(timezone=True), nullable=False)
    
    # Contagens por categoria
    categories = Column(JSONB, nullable=False)
    
    # Contagens por severidade
    severities = Column(JSONB, nullable=False)
    
    # Contagens por resultado
    success_count = Column(Integer, nullable=False, default=0)
    failure_count = Column(Integer, nullable=False, default=0)
    
    # Contagens por compliance
    compliance_frameworks = Column(JSONB, nullable=True)
    
    # Top recursos e usuários
    top_resources = Column(JSONB, nullable=True)
    top_users = Column(JSONB, nullable=True)
    
    # Metadados
    total_count = Column(Integer, nullable=False, default=0)
    generated_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    
    # Restrição única para evitar estatísticas duplicadas
    __table_args__ = (
        Index('idx_unique_audit_statistics', 
              tenant_id, regional_context, period_type, period_start, 
              unique=True),
    )