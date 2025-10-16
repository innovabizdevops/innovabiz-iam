#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Schema GraphQL para consulta de análises comportamentais

Este módulo define o schema GraphQL para consultas relacionadas a
análises comportamentais, eventos, alertas e aprovações.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import graphene
from graphene import ObjectType, String, Float, Int, Boolean, List, Field
from graphene import Enum as GrapheneEnum
from datetime import datetime, timedelta

from ..authorization.approval_matrix import AlertSeverity, AlertCategory, ApprovalAction


# Enums do GraphQL
class GraphQLAlertSeverity(GrapheneEnum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class GraphQLAlertCategory(GrapheneEnum):
    AUTHENTICATION = "authentication"
    TRANSACTION = "transaction"
    SESSION = "session"
    DEVICE = "device"
    LOCATION = "location"
    PROFILE = "profile"
    COMBINED = "combined"


class GraphQLApprovalStatus(GrapheneEnum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    AUTO_APPROVED = "auto_approved"
    ESCALATED = "escalated"
    INVESTIGATING = "investigating"
    CHALLENGING = "challenging"
    BLOCKED = "blocked"
    RESTRICTED = "restricted"
    ERROR = "error"


class GraphQLApprovalAction(GrapheneEnum):
    APPROVE = "approve"
    REJECT = "reject"
    ESCALATE = "escalate"
    INVESTIGATE = "investigate"
    CHALLENGE = "challenge"
    BLOCK = "block"
    RESTRICT = "restrict"
    MONITOR = "monitor"


# Tipos do GraphQL
class BehavioralFeature(ObjectType):
    """Característica comportamental extraída de um evento."""
    name = String(description="Nome da característica")
    value = Float(description="Valor numérico da característica")
    category = String(description="Categoria da característica")
    normalized_value = Float(description="Valor normalizado (0-1)")
    description = String(description="Descrição da característica")


class BehavioralIndicator(ObjectType):
    """Indicador comportamental calculado."""
    name = String(description="Nome do indicador")
    value = Float(description="Valor do indicador")
    threshold = Float(description="Valor limite para alerta")
    description = String(description="Descrição do indicador")
    severity = GraphQLAlertSeverity(description="Severidade associada ao indicador")
    exceeded = Boolean(description="Indica se o limite foi excedido")


class AnomalyScore(ObjectType):
    """Score de anomalia calculado para um evento."""
    score = Float(description="Score de anomalia (0-1)")
    threshold = Float(description="Limite para determinar anomalia")
    is_anomaly = Boolean(description="Indica se é anomalia")
    confidence = Float(description="Confiança da detecção (0-1)")
    model_type = String(description="Tipo de modelo usado")
    contributing_factors = List(BehavioralFeature, description="Fatores que contribuíram para o score")
    timestamp = String(description="Timestamp da detecção")


class BehavioralEvent(ObjectType):
    """Evento comportamental capturado."""
    event_id = String(description="ID do evento")
    user_id = String(description="ID do usuário")
    session_id = String(description="ID da sessão")
    event_type = String(description="Tipo do evento")
    timestamp = String(description="Timestamp do evento")
    region = String(description="Região do evento")
    device_info = graphene.JSONString(description="Informações sobre o dispositivo")
    location_info = graphene.JSONString(description="Informações sobre localização")
    network_info = graphene.JSONString(description="Informações sobre rede")
    transaction_info = graphene.JSONString(description="Informações sobre transação")
    risk_score = Float(description="Score de risco calculado")
    features = List(BehavioralFeature, description="Características extraídas")
    anomaly_score = Field(AnomalyScore, description="Score de anomalia")
    alerts_generated = List(String, description="IDs de alertas gerados")


class BehavioralAlert(ObjectType):
    """Alerta comportamental."""
    alert_id = String(description="ID do alerta")
    event_id = String(description="ID do evento que gerou o alerta")
    user_id = String(description="ID do usuário")
    severity = GraphQLAlertSeverity(description="Severidade do alerta")
    category = GraphQLAlertCategory(description="Categoria do alerta")
    timestamp = String(description="Timestamp de criação do alerta")
    region = String(description="Região do alerta")
    risk_score = Float(description="Score de risco")
    transaction_amount = Float(description="Valor da transação, se aplicável")
    details = graphene.JSONString(description="Detalhes do alerta")
    status = String(description="Status atual do alerta")
    approval_request_id = String(description="ID da solicitação de aprovação, se houver")
    indicators = List(BehavioralIndicator, description="Indicadores que geraram o alerta")


class ApprovalEntry(ObjectType):
    """Entrada de aprovação por um usuário."""
    user_id = String(description="ID do usuário")
    user_name = String(description="Nome do usuário")
    approval_level = String(description="Nível de aprovação do usuário")
    action = GraphQLApprovalAction(description="Ação realizada")
    timestamp = String(description="Timestamp da aprovação")
    comments = String(description="Comentários do aprovador")


class ApprovalRequest(ObjectType):
    """Solicitação de aprovação para um alerta."""
    request_id = String(description="ID da solicitação")
    alert_id = String(description="ID do alerta")
    created_at = String(description="Timestamp de criação")
    updated_at = String(description="Timestamp da última atualização")
    status = GraphQLApprovalStatus(description="Status da aprovação")
    required_approval_level = String(description="Nível de aprovação requerido")
    current_approval_level = String(description="Nível de aprovação atual")
    escalation_count = Int(description="Número de escalações")
    can_auto_approve = Boolean(description="Indica se pode ser auto-aprovado")
    approvals = List(ApprovalEntry, description="Aprovações recebidas")
    comments = List(graphene.JSONString, description="Comentários na solicitação")
    alert = Field(BehavioralAlert, description="Alerta associado")


class UserRiskProfile(ObjectType):
    """Perfil de risco de um usuário."""
    user_id = String(description="ID do usuário")
    base_risk_score = Float(description="Score de risco base")
    current_risk_score = Float(description="Score de risco atual")
    last_updated = String(description="Última atualização do perfil")
    risk_factors = List(BehavioralIndicator, description="Fatores de risco ativos")
    historical_scores = List(graphene.JSONString, description="Histórico de scores")
    alerts_count = Int(description="Número de alertas gerados")
    recent_alerts = List(BehavioralAlert, description="Alertas recentes")
    region = String(description="Região principal do usuário")


class BehavioralAnalysisStatistics(ObjectType):
    """Estatísticas de análise comportamental."""
    period_start = String(description="Início do período")
    period_end = String(description="Fim do período")
    total_events = Int(description="Total de eventos analisados")
    total_alerts = Int(description="Total de alertas gerados")
    alerts_by_severity = graphene.JSONString(description="Alertas por severidade")
    alerts_by_category = graphene.JSONString(description="Alertas por categoria")
    alerts_by_region = graphene.JSONString(description="Alertas por região")
    approval_statistics = graphene.JSONString(description="Estatísticas de aprovação")
    avg_processing_time = Float(description="Tempo médio de processamento (ms)")
    avg_approval_time = Float(description="Tempo médio de aprovação (min)")


class Query(ObjectType):
    """Consultas GraphQL para análise comportamental."""
    
    # Consultas para eventos comportamentais
    event = Field(
        BehavioralEvent, 
        id=String(required=True, description="ID do evento"),
        description="Obter um evento comportamental por ID"
    )
    
    events_by_user = Field(
        List(BehavioralEvent),
        user_id=String(required=True, description="ID do usuário"),
        start_date=String(description="Data inicial (ISO format)"),
        end_date=String(description="Data final (ISO format)"),
        limit=Int(default_value=10, description="Limite de resultados"),
        offset=Int(default_value=0, description="Offset para paginação"),
        description="Obter eventos comportamentais de um usuário"
    )
    
    # Consultas para alertas comportamentais
    alert = Field(
        BehavioralAlert,
        id=String(required=True, description="ID do alerta"),
        description="Obter um alerta comportamental por ID"
    )
    
    alerts = Field(
        List(BehavioralAlert),
        user_id=String(description="Filtrar por ID do usuário"),
        severity=GraphQLAlertSeverity(description="Filtrar por severidade"),
        category=GraphQLAlertCategory(description="Filtrar por categoria"),
        region=String(description="Filtrar por região"),
        start_date=String(description="Data inicial (ISO format)"),
        end_date=String(description="Data final (ISO format)"),
        status=String(description="Filtrar por status"),
        limit=Int(default_value=10, description="Limite de resultados"),
        offset=Int(default_value=0, description="Offset para paginação"),
        description="Obter alertas comportamentais com filtros"
    )
    
    # Consultas para solicitações de aprovação
    approval_request = Field(
        ApprovalRequest,
        id=String(required=True, description="ID da solicitação"),
        description="Obter uma solicitação de aprovação por ID"
    )
    
    approval_requests = Field(
        List(ApprovalRequest),
        status=GraphQLApprovalStatus(description="Filtrar por status"),
        approval_level=String(description="Filtrar por nível de aprovação"),
        region=String(description="Filtrar por região"),
        start_date=String(description="Data inicial (ISO format)"),
        end_date=String(description="Data final (ISO format)"),
        limit=Int(default_value=10, description="Limite de resultados"),
        offset=Int(default_value=0, description="Offset para paginação"),
        description="Obter solicitações de aprovação com filtros"
    )
    
    # Consulta para perfil de risco do usuário
    user_risk_profile = Field(
        UserRiskProfile,
        user_id=String(required=True, description="ID do usuário"),
        description="Obter o perfil de risco de um usuário"
    )
    
    # Consulta para estatísticas
    behavior_analysis_statistics = Field(
        BehavioralAnalysisStatistics,
        region=String(description="Filtrar por região"),
        start_date=String(description="Data inicial (ISO format)"),
        end_date=String(description="Data final (ISO format)"),
        description="Obter estatísticas de análise comportamental"
    )


schema = graphene.Schema(query=Query)