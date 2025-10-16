#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Configurações para o Dashboard de Monitoramento de Anomalias Comportamentais

Este módulo contém as configurações para o dashboard de monitoramento,
incluindo definições de painéis, métricas, alertas e visualizações.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from typing import Dict, Any, List, Union, Optional

# Configurações gerais do dashboard
DASHBOARD_CONFIG = {
    "name": "Monitoramento de Anomalias Comportamentais",
    "description": "Dashboard para monitoramento em tempo real de anomalias comportamentais detectadas pelo sistema",
    "refresh_interval": 60,  # Intervalo de atualização em segundos
    "default_timespan": "24h",  # Período padrão de visualização
    "available_timespans": ["1h", "6h", "12h", "24h", "7d", "30d"],
    "default_region": "global",  # Região padrão para visualização
    "theme": "dark",  # Tema do dashboard (dark ou light)
    "enable_alerts": True,  # Habilitar alertas no dashboard
    "enable_notifications": True,  # Habilitar notificações push
    "enable_export": True,  # Habilitar exportação de dados
    "enable_drilldown": True,  # Habilitar análise detalhada
    "enable_realtime": True,  # Habilitar atualizações em tempo real
}

# Definições de painéis (panels)
DASHBOARD_PANELS = [
    # Painel de visão geral
    {
        "id": "overview",
        "title": "Visão Geral",
        "type": "summary",
        "metrics": ["total_events", "total_alerts", "anomaly_rate", "avg_risk_score"],
        "position": {"row": 0, "col": 0, "width": 12, "height": 4},
        "refresh_interval": 30,
    },
    
    # Gráfico de anomalias por tempo
    {
        "id": "anomalies_timeline",
        "title": "Anomalias por Tempo",
        "type": "timeseries",
        "metrics": ["anomaly_count", "risk_score"],
        "position": {"row": 4, "col": 0, "width": 8, "height": 6},
        "grouping": "1h",
        "chart_type": "line",
        "stacked": False,
    },
    
    # Gráfico de anomalias por severidade
    {
        "id": "anomalies_by_severity",
        "title": "Anomalias por Severidade",
        "type": "pie",
        "metrics": ["anomaly_count"],
        "position": {"row": 4, "col": 8, "width": 4, "height": 6},
        "groupBy": "severity",
    },
    
    # Gráfico de anomalias por categoria
    {
        "id": "anomalies_by_category",
        "title": "Anomalias por Categoria",
        "type": "bar",
        "metrics": ["anomaly_count"],
        "position": {"row": 10, "col": 0, "width": 6, "height": 6},
        "groupBy": "category",
        "orientation": "horizontal",
    },
    
    # Mapa de calor por região
    {
        "id": "region_heatmap",
        "title": "Mapa de Calor por Região",
        "type": "map",
        "metrics": ["anomaly_count"],
        "position": {"row": 10, "col": 6, "width": 6, "height": 6},
        "map_type": "world",
        "colorScale": "reds",
    },
    
    # Lista de alertas recentes
    {
        "id": "recent_alerts",
        "title": "Alertas Recentes",
        "type": "table",
        "metrics": ["alert_id", "severity", "category", "user_id", "region", "risk_score", "timestamp", "status"],
        "position": {"row": 16, "col": 0, "width": 12, "height": 6},
        "sort_by": "timestamp",
        "sort_direction": "desc",
        "max_items": 10,
        "enable_filtering": True,
        "enable_sorting": True,
    },
    
    # Gráfico de usuários de maior risco
    {
        "id": "high_risk_users",
        "title": "Usuários de Maior Risco",
        "type": "bar",
        "metrics": ["risk_score"],
        "position": {"row": 22, "col": 0, "width": 6, "height": 6},
        "groupBy": "user_id",
        "limit": 10,
        "orientation": "horizontal",
    },
    
    # Gráfico de dispositivos de maior risco
    {
        "id": "high_risk_devices",
        "title": "Dispositivos de Maior Risco",
        "type": "bar",
        "metrics": ["risk_score"],
        "position": {"row": 22, "col": 6, "width": 6, "height": 6},
        "groupBy": "device_id",
        "limit": 10,
        "orientation": "horizontal",
    },
    
    # Painel de métricas de aprovação
    {
        "id": "approval_metrics",
        "title": "Métricas de Aprovação",
        "type": "statistic",
        "metrics": ["pending_approvals", "auto_approved_rate", "manual_approval_time", "escalation_rate"],
        "position": {"row": 28, "col": 0, "width": 12, "height": 4},
    },
    
    # Gráfico de distribuição de scores de anomalias
    {
        "id": "anomaly_score_distribution",
        "title": "Distribuição de Scores de Anomalias",
        "type": "histogram",
        "metrics": ["anomaly_score"],
        "position": {"row": 32, "col": 0, "width": 6, "height": 6},
        "bins": 20,
    },
    
    # Gráfico de precisão do modelo ao longo do tempo
    {
        "id": "model_accuracy_timeline",
        "title": "Precisão do Modelo ao Longo do Tempo",
        "type": "timeseries",
        "metrics": ["precision", "recall", "f1_score"],
        "position": {"row": 32, "col": 6, "width": 6, "height": 6},
        "grouping": "1d",
        "chart_type": "line",
    },
]

# Definições de métricas
DASHBOARD_METRICS = {
    "total_events": {
        "name": "Total de Eventos",
        "description": "Número total de eventos comportamentais processados",
        "format": "number",
        "aggregation": "sum",
    },
    "total_alerts": {
        "name": "Total de Alertas",
        "description": "Número total de alertas gerados",
        "format": "number",
        "aggregation": "sum",
    },
    "anomaly_count": {
        "name": "Contagem de Anomalias",
        "description": "Número de anomalias detectadas",
        "format": "number",
        "aggregation": "sum",
    },
    "anomaly_rate": {
        "name": "Taxa de Anomalias",
        "description": "Porcentagem de eventos classificados como anômalos",
        "format": "percentage",
        "aggregation": "avg",
    },
    "risk_score": {
        "name": "Score de Risco",
        "description": "Pontuação de risco (0-1)",
        "format": "float",
        "aggregation": "avg",
    },
    "anomaly_score": {
        "name": "Score de Anomalia",
        "description": "Pontuação de anomalia calculada pelo modelo (0-1)",
        "format": "float",
        "aggregation": "avg",
    },
    "avg_risk_score": {
        "name": "Score de Risco Médio",
        "description": "Média de scores de risco para o período",
        "format": "float",
        "decimals": 2,
        "aggregation": "avg",
    },
    "pending_approvals": {
        "name": "Aprovações Pendentes",
        "description": "Número de aprovações pendentes",
        "format": "number",
        "aggregation": "sum",
    },
    "auto_approved_rate": {
        "name": "Taxa de Auto-Aprovação",
        "description": "Porcentagem de alertas aprovados automaticamente",
        "format": "percentage",
        "aggregation": "avg",
    },
    "manual_approval_time": {
        "name": "Tempo de Aprovação Manual",
        "description": "Tempo médio para aprovações manuais (minutos)",
        "format": "time",
        "unit": "min",
        "aggregation": "avg",
    },
    "escalation_rate": {
        "name": "Taxa de Escalação",
        "description": "Porcentagem de alertas que precisaram ser escalados",
        "format": "percentage",
        "aggregation": "avg",
    },
    "precision": {
        "name": "Precisão",
        "description": "Precisão do modelo (0-1)",
        "format": "float",
        "decimals": 3,
        "aggregation": "avg",
    },
    "recall": {
        "name": "Recall",
        "description": "Recall do modelo (0-1)",
        "format": "float",
        "decimals": 3,
        "aggregation": "avg",
    },
    "f1_score": {
        "name": "F1-Score",
        "description": "F1-Score do modelo (0-1)",
        "format": "float",
        "decimals": 3,
        "aggregation": "avg",
    },
}

# Definições de filtros disponíveis
DASHBOARD_FILTERS = [
    {
        "id": "region",
        "name": "Região",
        "type": "select",
        "options": "dynamic",  # Carregar opções dinamicamente
        "default": "global",
        "multi_select": True,
    },
    {
        "id": "severity",
        "name": "Severidade",
        "type": "select",
        "options": [
            {"value": "low", "label": "Baixa"},
            {"value": "medium", "label": "Média"},
            {"value": "high", "label": "Alta"},
            {"value": "critical", "label": "Crítica"},
        ],
        "default": None,  # Todas selecionadas por padrão
        "multi_select": True,
    },
    {
        "id": "category",
        "name": "Categoria",
        "type": "select",
        "options": [
            {"value": "authentication", "label": "Autenticação"},
            {"value": "transaction", "label": "Transação"},
            {"value": "session", "label": "Sessão"},
            {"value": "device", "label": "Dispositivo"},
            {"value": "location", "label": "Localização"},
            {"value": "profile", "label": "Perfil"},
            {"value": "combined", "label": "Combinado"},
        ],
        "default": None,  # Todas selecionadas por padrão
        "multi_select": True,
    },
    {
        "id": "timespan",
        "name": "Período",
        "type": "select",
        "options": [
            {"value": "1h", "label": "Última hora"},
            {"value": "6h", "label": "Últimas 6 horas"},
            {"value": "12h", "label": "Últimas 12 horas"},
            {"value": "24h", "label": "Últimas 24 horas"},
            {"value": "7d", "label": "Últimos 7 dias"},
            {"value": "30d", "label": "Últimos 30 dias"},
        ],
        "default": "24h",
        "multi_select": False,
    },
    {
        "id": "custom_timerange",
        "name": "Período Personalizado",
        "type": "daterange",
        "default": None,
    },
    {
        "id": "user_id",
        "name": "ID do Usuário",
        "type": "text",
        "default": None,
    },
]

# Definições de alertas
DASHBOARD_ALERTS = [
    {
        "id": "high_anomaly_rate",
        "name": "Taxa de Anomalias Alta",
        "description": "Alerta quando a taxa de anomalias ultrapassa um limite",
        "metric": "anomaly_rate",
        "condition": "gt",  # greater than
        "threshold": 0.15,  # 15%
        "timespan": "1h",
        "severity": "high",
    },
    {
        "id": "critical_anomalies_increase",
        "name": "Aumento de Anomalias Críticas",
        "description": "Alerta quando há aumento significativo de anomalias críticas",
        "metric": "anomaly_count",
        "filter": {"severity": "critical"},
        "condition": "percent_increase",
        "threshold": 50,  # 50% de aumento
        "timespan": "1h",
        "comparison_timespan": "1h",  # Comparar com a hora anterior
        "severity": "critical",
    },
    {
        "id": "high_risk_transactions",
        "name": "Transações de Alto Risco",
        "description": "Alerta quando há transações com score de risco muito alto",
        "metric": "risk_score",
        "filter": {"category": "transaction"},
        "condition": "gt",
        "threshold": 0.85,  # Score acima de 0.85
        "timespan": "10m",  # Últimos 10 minutos
        "severity": "critical",
    },
    {
        "id": "pending_approvals_high",
        "name": "Muitas Aprovações Pendentes",
        "description": "Alerta quando há muitas aprovações pendentes",
        "metric": "pending_approvals",
        "condition": "gt",
        "threshold": 10,  # Mais de 10 aprovações pendentes
        "timespan": "30m",
        "severity": "medium",
    },
    {
        "id": "model_accuracy_drop",
        "name": "Queda na Precisão do Modelo",
        "description": "Alerta quando há queda significativa na precisão do modelo",
        "metric": "precision",
        "condition": "percent_decrease",
        "threshold": 10,  # 10% de queda
        "timespan": "1d",
        "comparison_timespan": "1d",  # Comparar com o dia anterior
        "severity": "high",
    },
]

# Definições de cores e estilos
DASHBOARD_STYLES = {
    "severity_colors": {
        "low": "#5cb85c",       # Verde
        "medium": "#f0ad4e",    # Laranja
        "high": "#d9534f",      # Vermelho
        "critical": "#9c27b0",  # Roxo
    },
    "category_colors": {
        "authentication": "#2196f3",  # Azul
        "transaction": "#ff9800",     # Laranja
        "session": "#4caf50",         # Verde
        "device": "#9e9e9e",          # Cinza
        "location": "#e91e63",        # Rosa
        "profile": "#673ab7",         # Roxo
        "combined": "#3f51b5",        # Indigo
    },
    "chart_colors": [
        "#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", 
        "#9467bd", "#8c564b", "#e377c2", "#7f7f7f", 
        "#bcbd22", "#17becf"
    ],
    "themes": {
        "dark": {
            "background": "#121212",
            "panel_background": "#1e1e1e",
            "text_color": "#ffffff",
            "border_color": "#333333",
            "grid_color": "#333333",
        },
        "light": {
            "background": "#f5f5f5",
            "panel_background": "#ffffff",
            "text_color": "#333333",
            "border_color": "#dddddd",
            "grid_color": "#eeeeee",
        }
    }
}

# Configurações específicas para cada região
REGION_SPECIFIC_CONFIGS = {
    "BR": {
        "dashboard_title": "Monitoramento de Anomalias Comportamentais - Brasil",
        "timespan": "12h",  # Período padrão menor para Brasil
        "alert_thresholds": {
            "high_anomaly_rate": 0.12,  # Limite mais baixo para Brasil
            "high_risk_transactions": 0.80,
        }
    },
    "MZ": {
        "dashboard_title": "Monitoramento de Anomalias Comportamentais - Moçambique",
        "alert_thresholds": {
            "high_anomaly_rate": 0.10,  # Limite mais baixo para Moçambique
            "high_risk_transactions": 0.75,
        }
    },
    "AO": {
        "dashboard_title": "Monitoramento de Anomalias Comportamentais - Angola",
        "alert_thresholds": {
            "high_anomaly_rate": 0.10,  # Limite mais baixo para Angola
            "high_risk_transactions": 0.75,
        }
    },
}

# Obter configuração para uma região específica
def get_dashboard_config_for_region(region: Optional[str] = None) -> Dict[str, Any]:
    """
    Obtém a configuração do dashboard para uma região específica.
    
    Args:
        region: Código da região
        
    Returns:
        Configuração do dashboard
    """
    config = DASHBOARD_CONFIG.copy()
    
    if region and region in REGION_SPECIFIC_CONFIGS:
        region_config = REGION_SPECIFIC_CONFIGS[region]
        
        # Sobrescrever configurações gerais com as específicas da região
        for key, value in region_config.items():
            if key != "alert_thresholds":
                config[key] = value
        
        # Ajustar thresholds de alertas
        if "alert_thresholds" in region_config:
            for alert_id, threshold in region_config["alert_thresholds"].items():
                for i, alert in enumerate(DASHBOARD_ALERTS):
                    if alert["id"] == alert_id:
                        config_alerts = DASHBOARD_ALERTS.copy()
                        config_alerts[i] = alert.copy()
                        config_alerts[i]["threshold"] = threshold
                        break
    
    return {
        "general": config,
        "panels": DASHBOARD_PANELS,
        "metrics": DASHBOARD_METRICS,
        "filters": DASHBOARD_FILTERS,
        "alerts": DASHBOARD_ALERTS,
        "styles": DASHBOARD_STYLES,
    }