#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Gerenciador de Dashboard de Monitoramento de Anomalias Comportamentais

Este módulo implementa o gerenciador do dashboard de monitoramento, responsável
por buscar e agregar os dados, aplicar filtros e fornecer visualizações.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import json
import logging
import datetime
from typing import Dict, Any, List, Optional, Union, Tuple

from .config import get_dashboard_config_for_region
from ..authorization.approval_matrix import ApprovalMatrix
from ..ml_models.anomaly_detection import AnomalyDetector

# Configuração do logger
logger = logging.getLogger("iam.trustguard.dashboard.manager")


class DashboardManager:
    """
    Gerenciador do dashboard de monitoramento de anomalias comportamentais.
    
    Responsável por buscar dados, aplicar filtros e gerenciar visualizações.
    """
    
    def __init__(self, data_source=None, region: Optional[str] = None):
        """
        Inicializa o gerenciador do dashboard.
        
        Args:
            data_source: Fonte de dados (opcional, será simulada se não fornecida)
            region: Região para configurações específicas
        """
        self.data_source = data_source or self._create_mock_data_source()
        self.region = region
        
        # Carregar configurações do dashboard
        self.config = get_dashboard_config_for_region(region)
        
        # Inicializar cache de dados
        self.data_cache = {}
        self.cache_timestamp = {}
        self.cache_expiry = 60  # Segundos
        
        logger.info(
            f"Dashboard Manager inicializado para região: "
            f"{region or 'global'}"
        )
    
    def _create_mock_data_source(self):
        """
        Cria uma fonte de dados simulada para testes.
        
        Returns:
            Fonte de dados simulada
        """
        # Em produção, seria conectado a bancos de dados reais
        return {
            "events": {},
            "alerts": {},
            "approval_requests": {},
            "user_profiles": {},
        }
    
    def get_dashboard_configuration(self) -> Dict[str, Any]:
        """
        Obtém a configuração completa do dashboard.
        
        Returns:
            Configuração do dashboard
        """
        return self.config
    
    def get_panel_data(self, panel_id: str, filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Obtém dados para um painel específico com filtros opcionais.
        
        Args:
            panel_id: ID do painel
            filters: Filtros a serem aplicados
            
        Returns:
            Dados do painel
        """
        try:
            # Encontrar definição do painel
            panel_def = None
            for panel in self.config["panels"]:
                if panel["id"] == panel_id:
                    panel_def = panel
                    break
            
            if not panel_def:
                logger.warning(f"Painel não encontrado: {panel_id}")
                return {"error": f"Painel não encontrado: {panel_id}"}
            
            # Verificar cache
            cache_key = f"{panel_id}_{json.dumps(filters or {})}"
            if cache_key in self.data_cache:
                cache_time = self.cache_timestamp.get(cache_key, 0)
                current_time = datetime.datetime.now().timestamp()
                
                # Usar cache se não expirou
                if current_time - cache_time < self.cache_expiry:
                    return self.data_cache[cache_key]
            
            # Buscar dados de acordo com o tipo de painel
            data = self._fetch_panel_data(panel_def, filters)
            
            # Armazenar no cache
            self.data_cache[cache_key] = data
            self.cache_timestamp[cache_key] = datetime.datetime.now().timestamp()
            
            return data
            
        except Exception as e:
            logger.error(f"Erro ao obter dados do painel {panel_id}: {str(e)}")
            return {"error": f"Erro ao obter dados: {str(e)}"}
    
    def _fetch_panel_data(self, panel_def: Dict[str, Any], filters: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Busca dados para um painel com base em sua definição e filtros.
        
        Args:
            panel_def: Definição do painel
            filters: Filtros a serem aplicados
            
        Returns:
            Dados do painel
        """
        panel_type = panel_def.get("type", "")
        panel_id = panel_def.get("id", "")
        
        # Aplicar filtros padrão se não fornecidos
        if not filters:
            filters = {}
        
        # Filtro de período
        timespan = filters.get("timespan", panel_def.get("timespan", "24h"))
        end_date = datetime.datetime.now()
        
        # Converter período para timedelta
        if timespan == "1h":
            start_date = end_date - datetime.timedelta(hours=1)
        elif timespan == "6h":
            start_date = end_date - datetime.timedelta(hours=6)
        elif timespan == "12h":
            start_date = end_date - datetime.timedelta(hours=12)
        elif timespan == "24h":
            start_date = end_date - datetime.timedelta(hours=24)
        elif timespan == "7d":
            start_date = end_date - datetime.timedelta(days=7)
        elif timespan == "30d":
            start_date = end_date - datetime.timedelta(days=30)
        else:
            start_date = end_date - datetime.timedelta(hours=24)  # Padrão
        
        # Filtro de período personalizado
        if "start_date" in filters and "end_date" in filters:
            try:
                start_date = datetime.datetime.fromisoformat(filters["start_date"])
                end_date = datetime.datetime.fromisoformat(filters["end_date"])
            except (ValueError, TypeError):
                logger.warning("Formato de data inválido nos filtros")
        
        # Buscar dados específicos com base no tipo de painel
        if panel_type == "summary":
            return self._fetch_summary_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "timeseries":
            return self._fetch_timeseries_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "pie":
            return self._fetch_pie_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "bar":
            return self._fetch_bar_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "map":
            return self._fetch_map_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "table":
            return self._fetch_table_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "statistic":
            return self._fetch_statistic_data(panel_def, start_date, end_date, filters)
        
        elif panel_type == "histogram":
            return self._fetch_histogram_data(panel_def, start_date, end_date, filters)
        
        else:
            logger.warning(f"Tipo de painel não suportado: {panel_type}")
            return {"error": f"Tipo de painel não suportado: {panel_type}"}
    
    def _fetch_summary_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                           end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de resumo.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        # Em produção, buscar dos bancos de dados reais
        metrics = panel_def.get("metrics", [])
        data = {}
        
        # Aplicar filtros regionais se necessário
        region = filters.get("region", self.region)
        
        for metric in metrics:
            # Simular dados para cada métrica
            if metric == "total_events":
                data[metric] = 12458  # Simulado
            elif metric == "total_alerts":
                data[metric] = 347  # Simulado
            elif metric == "anomaly_rate":
                data[metric] = 0.028  # Simulado (2.8%)
            elif metric == "avg_risk_score":
                data[metric] = 0.42  # Simulado
        
        return {
            "title": panel_def.get("title", ""),
            "type": "summary",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            }
        }
        
    def _fetch_timeseries_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                              end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de série temporal.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        grouping = panel_def.get("grouping", "1h")
        
        # Converter grouping para timedelta
        if grouping == "1m":
            delta = datetime.timedelta(minutes=1)
        elif grouping == "5m":
            delta = datetime.timedelta(minutes=5)
        elif grouping == "15m":
            delta = datetime.timedelta(minutes=15)
        elif grouping == "30m":
            delta = datetime.timedelta(minutes=30)
        elif grouping == "1h":
            delta = datetime.timedelta(hours=1)
        elif grouping == "6h":
            delta = datetime.timedelta(hours=6)
        elif grouping == "12h":
            delta = datetime.timedelta(hours=12)
        elif grouping == "1d":
            delta = datetime.timedelta(days=1)
        else:
            delta = datetime.timedelta(hours=1)  # Padrão
            
        # Gerar pontos no tempo
        current = start_date
        timestamps = []
        while current <= end_date:
            timestamps.append(current)
            current += delta
            
        # Gerar dados simulados para cada métrica
        data = {metric: [] for metric in metrics}
        
        import random
        for timestamp in timestamps:
            for metric in metrics:
                if metric == "anomaly_count":
                    # Simular contagem de anomalias com tendência crescente
                    base = 5 + (timestamp - start_date).total_seconds() / 86400  # Aumenta ao longo do dia
                    value = max(0, base + random.randint(-2, 5))
                elif metric == "risk_score":
                    # Simular score de risco
                    value = 0.3 + random.random() * 0.4  # Entre 0.3 e 0.7
                elif metric == "precision":
                    # Simular precisão do modelo
                    value = 0.85 + random.random() * 0.1  # Entre 0.85 e 0.95
                elif metric == "recall":
                    # Simular recall do modelo
                    value = 0.75 + random.random() * 0.15  # Entre 0.75 e 0.9
                elif metric == "f1_score":
                    # Simular F1-Score do modelo
                    value = 0.8 + random.random() * 0.12  # Entre 0.8 e 0.92
                else:
                    value = random.random()  # Valor aleatório para outras métricas
                    
                data[metric].append({
                    "timestamp": timestamp.isoformat(),
                    "value": value
                })
        
        return {
            "title": panel_def.get("title", ""),
            "type": "timeseries",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
            "grouping": grouping,
        }
    
    def _fetch_pie_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                       end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de gráfico de pizza.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        group_by = panel_def.get("groupBy", "")
        
        data = []
        
        if group_by == "severity":
            # Dados de severidade
            data = [
                {"label": "Baixa", "value": 142, "color": self.config["styles"]["severity_colors"]["low"]},
                {"label": "Média", "value": 98, "color": self.config["styles"]["severity_colors"]["medium"]},
                {"label": "Alta", "value": 65, "color": self.config["styles"]["severity_colors"]["high"]},
                {"label": "Crítica", "value": 42, "color": self.config["styles"]["severity_colors"]["critical"]},
            ]
        elif group_by == "category":
            # Dados de categoria
            data = [
                {"label": "Autenticação", "value": 78, "color": self.config["styles"]["category_colors"]["authentication"]},
                {"label": "Transação", "value": 103, "color": self.config["styles"]["category_colors"]["transaction"]},
                {"label": "Sessão", "value": 64, "color": self.config["styles"]["category_colors"]["session"]},
                {"label": "Dispositivo", "value": 45, "color": self.config["styles"]["category_colors"]["device"]},
                {"label": "Localização", "value": 36, "color": self.config["styles"]["category_colors"]["location"]},
                {"label": "Perfil", "value": 21, "color": self.config["styles"]["category_colors"]["profile"]},
            ]
        elif group_by == "region":
            # Dados de região
            data = [
                {"label": "Brasil", "value": 120},
                {"label": "Moçambique", "value": 84},
                {"label": "Angola", "value": 72},
                {"label": "Portugal", "value": 45},
                {"label": "Outros", "value": 26},
            ]
        
        return {
            "title": panel_def.get("title", ""),
            "type": "pie",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
        }
    
    def _fetch_bar_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                       end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de gráfico de barras.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        group_by = panel_def.get("groupBy", "")
        limit = panel_def.get("limit", 10)
        orientation = panel_def.get("orientation", "vertical")
        
        data = []
        
        if group_by == "category":
            # Dados de categoria
            data = [
                {"label": "Autenticação", "value": 78, "color": self.config["styles"]["category_colors"]["authentication"]},
                {"label": "Transação", "value": 103, "color": self.config["styles"]["category_colors"]["transaction"]},
                {"label": "Sessão", "value": 64, "color": self.config["styles"]["category_colors"]["session"]},
                {"label": "Dispositivo", "value": 45, "color": self.config["styles"]["category_colors"]["device"]},
                {"label": "Localização", "value": 36, "color": self.config["styles"]["category_colors"]["location"]},
                {"label": "Perfil", "value": 21, "color": self.config["styles"]["category_colors"]["profile"]},
            ]
        elif group_by == "user_id":
            # Dados de usuários (top N)
            import random
            data = []
            for i in range(limit):
                user_id = f"user_{1000 + i}"
                risk_score = 0.4 + (limit - i) / limit * 0.5  # Decrescente
                data.append({
                    "label": user_id,
                    "value": risk_score,
                })
        elif group_by == "device_id":
            # Dados de dispositivos (top N)
            import random
            data = []
            for i in range(limit):
                device_id = f"device_{2000 + i}"
                risk_score = 0.3 + (limit - i) / limit * 0.6  # Decrescente
                data.append({
                    "label": device_id,
                    "value": risk_score,
                })
        
        return {
            "title": panel_def.get("title", ""),
            "type": "bar",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
            "orientation": orientation,
        }
    
    def _fetch_map_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                       end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de mapa.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        map_type = panel_def.get("map_type", "world")
        
        # Simulação de dados para mapa
        data = [
            {"code": "BR", "name": "Brasil", "value": 120, "lat": -14.235, "lng": -51.925},
            {"code": "MZ", "name": "Moçambique", "value": 84, "lat": -18.665, "lng": 35.529},
            {"code": "AO", "name": "Angola", "value": 72, "lat": -11.202, "lng": 17.873},
            {"code": "PT", "name": "Portugal", "value": 45, "lat": 39.399, "lng": -8.224},
            {"code": "CV", "name": "Cabo Verde", "value": 18, "lat": 16.002, "lng": -24.013},
            {"code": "GW", "name": "Guiné-Bissau", "value": 12, "lat": 11.803, "lng": -15.180},
            {"code": "TL", "name": "Timor-Leste", "value": 8, "lat": -8.874, "lng": 125.728},
            {"code": "ST", "name": "São Tomé e Príncipe", "value": 5, "lat": 0.186, "lng": 6.613},
        ]
        
        return {
            "title": panel_def.get("title", ""),
            "type": "map",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
            "map_type": map_type,
        }
    
    def _fetch_table_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                         end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de tabela.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        columns = panel_def.get("metrics", [])
        sort_by = panel_def.get("sort_by", "timestamp")
        sort_direction = panel_def.get("sort_direction", "desc")
        max_items = panel_def.get("max_items", 10)
        
        # Simulação de alertas para a tabela
        import random
        from datetime import timedelta
        
        data = []
        severities = ["low", "medium", "high", "critical"]
        categories = ["authentication", "transaction", "session", "device", "location", "profile"]
        regions = ["BR", "MZ", "AO", "PT", "CV"]
        statuses = ["pending", "approved", "rejected", "investigating"]
        
        # Gerar registros simulados
        for i in range(max_items):
            # Timestamp aleatório dentro do período
            seconds = random.randint(0, int((end_date - start_date).total_seconds()))
            timestamp = start_date + timedelta(seconds=seconds)
            
            # Atributos aleatórios
            severity = random.choice(severities)
            category = random.choice(categories)
            region = random.choice(regions)
            status = random.choice(statuses)
            
            # Montar registro
            record = {
                "alert_id": f"ALT-{100000 + i}",
                "severity": severity,
                "category": category,
                "user_id": f"user_{1000 + random.randint(0, 50)}",
                "region": region,
                "risk_score": round(0.3 + random.random() * 0.6, 2),
                "timestamp": timestamp.isoformat(),
                "status": status,
            }
            
            data.append(record)
        
        # Ordenar dados
        data.sort(key=lambda x: x.get(sort_by, ""), reverse=(sort_direction == "desc"))
        
        return {
            "title": panel_def.get("title", ""),
            "type": "table",
            "columns": columns,
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
            "sort": {
                "column": sort_by,
                "direction": sort_direction,
            },
        }
        
    def _fetch_statistic_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                             end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de estatísticas.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        data = {}
        
        # Simular valores para cada métrica
        for metric in metrics:
            if metric == "pending_approvals":
                data[metric] = 8  # Simulado
            elif metric == "auto_approved_rate":
                data[metric] = 0.65  # Simulado (65%)
            elif metric == "manual_approval_time":
                data[metric] = 12.3  # Simulado (12.3 minutos)
            elif metric == "escalation_rate":
                data[metric] = 0.18  # Simulado (18%)
        
        return {
            "title": panel_def.get("title", ""),
            "type": "statistic",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
        }
        
    def _fetch_histogram_data(self, panel_def: Dict[str, Any], start_date: datetime.datetime, 
                             end_date: datetime.datetime, filters: Dict[str, Any]) -> Dict[str, Any]:
        """
        Busca dados para um painel de histograma.
        
        Args:
            panel_def: Definição do painel
            start_date: Data inicial
            end_date: Data final
            filters: Filtros adicionais
            
        Returns:
            Dados do painel
        """
        metrics = panel_def.get("metrics", [])
        bins = panel_def.get("bins", 10)
        
        data = []
        
        # Simular distribuição para cada métrica
        for metric in metrics:
            if metric == "anomaly_score":
                # Criar histograma de scores
                import numpy as np
                
                # Distribuição normal com média 0.45 e desvio 0.15
                mean = 0.45
                stddev = 0.15
                values = np.random.normal(mean, stddev, 500)
                
                # Limitar aos valores entre 0 e 1
                values = np.clip(values, 0, 1)
                
                # Calcular histograma
                hist, bin_edges = np.histogram(values, bins=bins, range=(0, 1))
                
                # Formatar dados para visualização
                hist_data = []
                for i in range(len(hist)):
                    bin_start = bin_edges[i]
                    bin_end = bin_edges[i + 1]
                    bin_center = (bin_start + bin_end) / 2
                    hist_data.append({
                        "bin_start": float(bin_start),
                        "bin_end": float(bin_end),
                        "bin_center": float(bin_center),
                        "count": int(hist[i]),
                    })
                
                data = hist_data
        
        return {
            "title": panel_def.get("title", ""),
            "type": "histogram",
            "data": data,
            "period": {
                "start": start_date.isoformat(),
                "end": end_date.isoformat(),
            },
            "bins": bins,
        }