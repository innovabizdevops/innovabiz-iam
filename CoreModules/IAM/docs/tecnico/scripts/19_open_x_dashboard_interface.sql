-- =============================================
-- Interface de Visualização do Dashboard Open X
-- INNOVABIZ - IAM - Open X Dashboard UI
-- Versão: 1.0.0
-- Data: 15/05/2025
-- Autor: INNOVABIZ DevOps Team
-- =============================================

-- Este script implementa a interface de visualização para o Dashboard Open X,
-- permitindo acesso interativo às métricas e análises de conformidade
-- do ecossistema Open X.

-- =============================================
-- Schemas necessários
-- =============================================

-- Verifica e cria os schemas necessários
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_dashboard_ui') THEN
        CREATE SCHEMA compliance_dashboard_ui;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_validators') THEN
        RAISE EXCEPTION 'O schema compliance_validators não existe. Execute primeiro os scripts de validadores.';
    END IF;
END $$;

-- =============================================
-- Tipos para componentes de UI
-- =============================================

-- Tipo para configuração de gráficos
CREATE TYPE compliance_dashboard_ui.chart_config AS (
    chart_id VARCHAR(50),
    chart_type VARCHAR(20),
    title TEXT,
    subtitle TEXT,
    x_axis_label TEXT,
    y_axis_label TEXT,
    color_palette TEXT[],
    show_legend BOOLEAN,
    is_interactive BOOLEAN,
    height_px INTEGER,
    width_px INTEGER,
    data_source TEXT,
    refresh_interval_seconds INTEGER
);

-- Tipo para configuração de filtros
CREATE TYPE compliance_dashboard_ui.filter_config AS (
    filter_id VARCHAR(50),
    filter_type VARCHAR(20),
    label TEXT,
    placeholder TEXT,
    default_value TEXT,
    available_values TEXT[],
    is_multi_select BOOLEAN,
    parent_filter_id VARCHAR(50),
    is_mandatory BOOLEAN
);

-- Tipo para configuração de cards
CREATE TYPE compliance_dashboard_ui.card_config AS (
    card_id VARCHAR(50),
    title TEXT,
    icon TEXT,
    theme VARCHAR(20),
    data_source TEXT,
    value_format TEXT,
    sub_value_format TEXT,
    show_trend BOOLEAN,
    threshold_warning NUMERIC,
    threshold_critical NUMERIC
);

-- Tipo para layout do dashboard
CREATE TYPE compliance_dashboard_ui.layout_config AS (
    component_id VARCHAR(50),
    component_type VARCHAR(20),
    grid_x INTEGER,
    grid_y INTEGER,
    width INTEGER,
    height INTEGER,
    is_resizable BOOLEAN,
    is_draggable BOOLEAN,
    min_width INTEGER,
    min_height INTEGER
);

-- =============================================
-- Tabelas para gerenciamento dos componentes de UI
-- =============================================

-- Tabela para armazenar dashboards
CREATE TABLE compliance_dashboard_ui.dashboards (
    dashboard_id VARCHAR(50) PRIMARY KEY,
    dashboard_name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    icon TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    theme VARCHAR(20) DEFAULT 'light',
    refresh_interval_seconds INTEGER DEFAULT 300,
    tenant_id UUID
);

-- Tabela para armazenar componentes dos dashboards
CREATE TABLE compliance_dashboard_ui.dashboard_components (
    component_id VARCHAR(50) PRIMARY KEY,
    dashboard_id VARCHAR(50) REFERENCES compliance_dashboard_ui.dashboards(dashboard_id) ON DELETE CASCADE,
    component_type VARCHAR(20) NOT NULL CHECK (component_type IN ('chart', 'card', 'filter', 'table', 'alert', 'text')),
    component_config JSONB NOT NULL,
    layout compliance_dashboard_ui.layout_config,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    visible_to VARCHAR(50)[] DEFAULT '{economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role}'
);

-- Tabela para armazenar configurações personalizadas por usuário
CREATE TABLE compliance_dashboard_ui.user_dashboard_preferences (
    user_id VARCHAR(50),
    dashboard_id VARCHAR(50) REFERENCES compliance_dashboard_ui.dashboards(dashboard_id) ON DELETE CASCADE,
    preferences JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, dashboard_id)
);

-- Tabela para histórico de visualizações
CREATE TABLE compliance_dashboard_ui.dashboard_view_history (
    view_id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    dashboard_id VARCHAR(50) REFERENCES compliance_dashboard_ui.dashboards(dashboard_id) ON DELETE CASCADE,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    session_duration_seconds INTEGER,
    filters_applied JSONB,
    components_interacted VARCHAR(50)[]
);

-- =============================================
-- Definições dos dashboards Open X padrão
-- =============================================

-- Dashboard principal Open X
INSERT INTO compliance_dashboard_ui.dashboards (
    dashboard_id, dashboard_name, description, created_by, icon, 
    is_public, is_default, theme, refresh_interval_seconds
) VALUES (
    'open_x_main_dashboard',
    'Dashboard Principal Open X',
    'Visão consolidada de conformidade para todo o ecossistema Open X',
    'system',
    'dashboard',
    TRUE,
    TRUE,
    'light',
    300
);

-- Dashboard por domínio Open X
INSERT INTO compliance_dashboard_ui.dashboards (
    dashboard_id, dashboard_name, description, created_by, icon, 
    is_public, is_default, theme, refresh_interval_seconds
) VALUES (
    'open_x_domain_dashboard',
    'Dashboard por Domínio Open X',
    'Análise detalhada de conformidade por domínio específico (Insurance, Health, Government)',
    'system',
    'pie_chart',
    TRUE,
    FALSE,
    'light',
    300
);

-- Dashboard de impacto econômico Open X
INSERT INTO compliance_dashboard_ui.dashboards (
    dashboard_id, dashboard_name, description, created_by, icon, 
    is_public, is_default, theme, refresh_interval_seconds
) VALUES (
    'open_x_economic_dashboard',
    'Dashboard de Impacto Econômico Open X',
    'Análise de impacto econômico e ROI para conformidade Open X',
    'system',
    'monetization_on',
    TRUE,
    FALSE,
    'light',
    300
);

-- Dashboard de análise por jurisdição
INSERT INTO compliance_dashboard_ui.dashboards (
    dashboard_id, dashboard_name, description, created_by, icon, 
    is_public, is_default, theme, refresh_interval_seconds
) VALUES (
    'open_x_jurisdiction_dashboard',
    'Dashboard por Jurisdição Open X',
    'Análise de conformidade por jurisdição (Portugal/UE, Brasil, Angola, EUA)',
    'system',
    'public',
    TRUE,
    FALSE,
    'light',
    300
);

-- =============================================
-- Componentes do Dashboard Principal Open X
-- =============================================

-- Filtro de Tenant
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_tenant_main',
    'open_x_main_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'tenant_filter',
        'filter_type', 'dropdown',
        'label', 'Tenant',
        'placeholder', 'Selecione o tenant',
        'is_multi_select', false,
        'is_mandatory', true,
        'data_source', 'SELECT tenant_id, tenant_name FROM iam.tenants ORDER BY tenant_name'
    ),
    ROW('filter_tenant_main', 'filter', 0, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Filtro de Período
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_period_main',
    'open_x_main_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'period_filter',
        'filter_type', 'daterange',
        'label', 'Período',
        'default_value', 'last_30_days',
        'available_values', array['today', 'yesterday', 'last_7_days', 'last_30_days', 'last_90_days', 'year_to_date', 'custom'],
        'is_mandatory', true
    ),
    ROW('filter_period_main', 'filter', 3, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Conformidade Geral
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_overall_compliance',
    'open_x_main_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'overall_compliance_card',
        'title', 'Conformidade Geral',
        'icon', 'check_circle',
        'theme', 'primary',
        'data_source', 'SELECT AVG(compliance_percentage) FROM compliance_validators.vw_open_x_compliance_summary WHERE tenant_id = :tenant_id',
        'value_format', '{value}%',
        'show_trend', true,
        'threshold_warning', 75,
        'threshold_critical', 50
    ),
    ROW('card_overall_compliance', 'card', 0, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Domínios em Conformidade
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_compliant_domains',
    'open_x_main_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'compliant_domains_card',
        'title', 'Domínios em Conformidade',
        'icon', 'domain',
        'theme', 'success',
        'data_source', 'SELECT COUNT(*) FROM compliance_validators.vw_open_x_domain_comparison WHERE tenant_id = :tenant_id AND compliance_percentage >= 90',
        'value_format', '{value} de 3',
        'show_trend', false
    ),
    ROW('card_compliant_domains', 'card', 2, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Impacto Econômico Total
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_total_economic_impact',
    'open_x_main_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'economic_impact_card',
        'title', 'Impacto Econômico Total',
        'icon', 'monetization_on',
        'theme', 'warning',
        'data_source', 'SELECT SUM(total_economic_impact) FROM compliance_validators.vw_open_x_economic_impact WHERE tenant_id = :tenant_id',
        'value_format', 'R$ {value}',
        'show_trend', true
    ),
    ROW('card_total_economic_impact', 'card', 4, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Conformidade por Domínio
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_compliance_by_domain',
    'open_x_main_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'compliance_by_domain_chart',
        'chart_type', 'bar',
        'title', 'Conformidade por Domínio',
        'subtitle', 'Percentual de conformidade para cada domínio Open X',
        'x_axis_label', 'Domínio',
        'y_axis_label', 'Conformidade (%)',
        'color_palette', array['#4CAF50', '#2196F3', '#FFC107'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT open_x_domain as label, compliance_percentage as value FROM compliance_validators.vw_open_x_domain_comparison WHERE tenant_id = :tenant_id'
    ),
    ROW('chart_compliance_by_domain', 'chart', 0, 2, 3, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Impacto Econômico por Domínio
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_economic_impact_by_domain',
    'open_x_main_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'economic_impact_by_domain_chart',
        'chart_type', 'pie',
        'title', 'Impacto Econômico por Domínio',
        'subtitle', 'Distribuição do impacto econômico total por domínio',
        'color_palette', array['#F44336', '#9C27B0', '#3F51B5'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT open_x_domain as label, SUM(total_economic_impact) as value FROM compliance_validators.vw_open_x_economic_impact WHERE tenant_id = :tenant_id GROUP BY open_x_domain'
    ),
    ROW('chart_economic_impact_by_domain', 'chart', 3, 2, 3, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Tabela de Não-Conformidades Críticas
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'table_critical_non_compliance',
    'open_x_main_dashboard',
    'table',
    jsonb_build_object(
        'table_id', 'critical_non_compliance_table',
        'title', 'Não-Conformidades Críticas',
        'columns', array[
            '{"field": "open_x_domain", "headerName": "Domínio", "width": 120}',
            '{"field": "framework", "headerName": "Framework", "width": 120}',
            '{"field": "requirement_id", "headerName": "ID Requisito", "width": 100}',
            '{"field": "requirement_name", "headerName": "Nome do Requisito", "width": 200}',
            '{"field": "irr_threshold", "headerName": "Nível de Risco", "width": 120, "cellRenderer": "riskLevelRenderer"}'
        ],
        'data_source', 'SELECT open_x_domain, framework, requirement_id, requirement_name, irr_threshold FROM compliance_validators.vw_open_x_non_compliance_details WHERE tenant_id = :tenant_id AND irr_threshold IN (''R3'', ''R4'') ORDER BY CASE irr_threshold WHEN ''R4'' THEN 1 WHEN ''R3'' THEN 2 ELSE 3 END',
        'pagination', true,
        'page_size', 5,
        'sortable', true,
        'filterable', true
    ),
    ROW('table_critical_non_compliance', 'table', 0, 4, 6, 2, true, false, 6, 2)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Tendência de Conformidade
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_compliance_trend',
    'open_x_main_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'compliance_trend_chart',
        'chart_type', 'line',
        'title', 'Tendência de Conformidade',
        'subtitle', 'Evolução da conformidade ao longo do tempo',
        'x_axis_label', 'Data',
        'y_axis_label', 'Conformidade (%)',
        'color_palette', array['#4CAF50', '#F44336', '#2196F3'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT date_trunc(''day'', validation_date) as label, AVG(CASE WHEN is_compliant THEN 100 ELSE 0 END) as value, open_x_domain as series FROM compliance_validators.open_x_validation_history WHERE tenant_id = :tenant_id AND validation_date BETWEEN :start_date AND :end_date GROUP BY date_trunc(''day'', validation_date), open_x_domain ORDER BY date_trunc(''day'', validation_date)'
    ),
    ROW('chart_compliance_trend', 'chart', 0, 6, 6, 2, true, false, 6, 2)::compliance_dashboard_ui.layout_config
);

-- =============================================
-- Componentes do Dashboard por Domínio Open X
-- =============================================

-- Filtro de Tenant
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_tenant_domain',
    'open_x_domain_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'tenant_filter',
        'filter_type', 'dropdown',
        'label', 'Tenant',
        'placeholder', 'Selecione o tenant',
        'is_multi_select', false,
        'is_mandatory', true,
        'data_source', 'SELECT tenant_id, tenant_name FROM iam.tenants ORDER BY tenant_name'
    ),
    ROW('filter_tenant_domain', 'filter', 0, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Filtro de Domínio
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_domain_select',
    'open_x_domain_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'domain_filter',
        'filter_type', 'dropdown',
        'label', 'Domínio Open X',
        'placeholder', 'Selecione o domínio',
        'default_value', 'OPEN_INSURANCE',
        'available_values', array['OPEN_INSURANCE', 'OPEN_HEALTH', 'OPEN_GOVERNMENT'],
        'is_multi_select', false,
        'is_mandatory', true
    ),
    ROW('filter_domain_select', 'filter', 3, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Conformidade do Domínio
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_domain_compliance',
    'open_x_domain_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'domain_compliance_card',
        'title', 'Conformidade',
        'icon', 'check_circle',
        'theme', 'primary',
        'data_source', 'SELECT compliance_percentage FROM compliance_validators.vw_open_x_domain_comparison WHERE tenant_id = :tenant_id AND open_x_domain = :domain',
        'value_format', '{value}%',
        'show_trend', true,
        'threshold_warning', 75,
        'threshold_critical', 50
    ),
    ROW('card_domain_compliance', 'card', 0, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Conformidade por Framework
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_compliance_by_framework',
    'open_x_domain_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'compliance_by_framework_chart',
        'chart_type', 'bar',
        'title', 'Conformidade por Framework',
        'subtitle', 'Percentual de conformidade para cada framework',
        'x_axis_label', 'Framework',
        'y_axis_label', 'Conformidade (%)',
        'color_palette', array['#4CAF50', '#2196F3'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT framework as label, compliance_percentage as value FROM compliance_validators.vw_open_x_compliance_summary WHERE tenant_id = :tenant_id AND open_x_domain = :domain'
    ),
    ROW('chart_compliance_by_framework', 'chart', 2, 1, 4, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Tabela detalhada de requisitos
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'table_domain_requirements',
    'open_x_domain_dashboard',
    'table',
    jsonb_build_object(
        'table_id', 'domain_requirements_table',
        'title', 'Requisitos por Framework',
        'columns', array[
            '{"field": "framework", "headerName": "Framework", "width": 120}',
            '{"field": "requirement_id", "headerName": "ID Requisito", "width": 100}',
            '{"field": "requirement_name", "headerName": "Nome do Requisito", "width": 200}',
            '{"field": "is_compliant", "headerName": "Conforme", "width": 100, "cellRenderer": "complianceRenderer"}',
            '{"field": "irr_threshold", "headerName": "Nível de Risco", "width": 120, "cellRenderer": "riskLevelRenderer"}'
        ],
        'data_source', 'SELECT v.framework_id as framework, v.requirement_id, r.requirement_name, v.is_compliant, r.irr_threshold FROM compliance_validators.get_open_x_domain_validations(:tenant_id, :domain) v JOIN compliance_validators.get_open_x_domain_requirements(:domain) r ON v.requirement_id = r.requirement_id ORDER BY framework, requirement_id',
        'pagination', true,
        'page_size', 10,
        'sortable', true,
        'filterable', true
    ),
    ROW('table_domain_requirements', 'table', 0, 3, 6, 3, true, false, 6, 3)::compliance_dashboard_ui.layout_config
);

-- =============================================
-- Componentes do Dashboard Econômico Open X
-- =============================================

-- Filtro de Tenant
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_tenant_economic',
    'open_x_economic_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'tenant_filter',
        'filter_type', 'dropdown',
        'label', 'Tenant',
        'placeholder', 'Selecione o tenant',
        'is_multi_select', false,
        'is_mandatory', true,
        'data_source', 'SELECT tenant_id, tenant_name FROM iam.tenants ORDER BY tenant_name'
    ),
    ROW('filter_tenant_economic', 'filter', 0, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Filtro de Domínio (multi-select)
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_domain_economic',
    'open_x_economic_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'domain_filter',
        'filter_type', 'dropdown',
        'label', 'Domínios Open X',
        'placeholder', 'Selecione os domínios',
        'available_values', array['OPEN_INSURANCE', 'OPEN_HEALTH', 'OPEN_GOVERNMENT'],
        'is_multi_select', true
    ),
    ROW('filter_domain_economic', 'filter', 3, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Impacto Econômico Total
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_total_impact_economic',
    'open_x_economic_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'total_impact_card',
        'title', 'Impacto Econômico Total',
        'icon', 'monetization_on',
        'theme', 'warning',
        'data_source', 'SELECT SUM(total_economic_impact) FROM compliance_validators.vw_open_x_economic_impact WHERE tenant_id = :tenant_id AND (:domains IS NULL OR open_x_domain = ANY(:domains))',
        'value_format', 'R$ {value}',
        'show_trend', true
    ),
    ROW('card_total_impact_economic', 'card', 0, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Impacto Médio por Não-Conformidade
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_avg_impact_economic',
    'open_x_economic_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'avg_impact_card',
        'title', 'Impacto Médio por Não-Conformidade',
        'icon', 'trending_up',
        'theme', 'info',
        'data_source', 'SELECT AVG(avg_impact_per_non_compliance) FROM compliance_validators.vw_open_x_economic_impact WHERE tenant_id = :tenant_id AND (:domains IS NULL OR open_x_domain = ANY(:domains))',
        'value_format', 'R$ {value}',
        'show_trend', false
    ),
    ROW('card_avg_impact_economic', 'card', 2, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Impacto Econômico por Domínio
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_impact_by_domain_economic',
    'open_x_economic_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'impact_by_domain_chart',
        'chart_type', 'pie',
        'title', 'Distribuição do Impacto Econômico',
        'subtitle', 'Impacto econômico por domínio Open X',
        'color_palette', array['#F44336', '#9C27B0', '#3F51B5'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT open_x_domain as label, SUM(total_economic_impact) as value FROM compliance_validators.vw_open_x_economic_impact WHERE tenant_id = :tenant_id AND (:domains IS NULL OR open_x_domain = ANY(:domains)) GROUP BY open_x_domain'
    ),
    ROW('chart_impact_by_domain_economic', 'chart', 0, 2, 3, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Simulação de ROI
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_roi_simulation',
    'open_x_economic_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'roi_simulation_chart',
        'chart_type', 'bar',
        'title', 'Simulação de ROI',
        'subtitle', 'Estimativa de retorno sobre investimento para diferentes níveis de melhoria',
        'x_axis_label', 'Melhoria (%)',
        'y_axis_label', 'ROI Estimado (%)',
        'color_palette', array['#4CAF50'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT improvement as label, roi as value FROM compliance_validators.get_open_x_roi_simulations(:tenant_id, :domains, 24)'
    ),
    ROW('chart_roi_simulation', 'chart', 3, 2, 3, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Tabela de Não-Conformidades com Maior Impacto
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'table_high_impact_non_compliance',
    'open_x_economic_dashboard',
    'table',
    jsonb_build_object(
        'table_id', 'high_impact_table',
        'title', 'Não-Conformidades com Maior Impacto Econômico',
        'columns', array[
            '{"field": "open_x_domain", "headerName": "Domínio", "width": 120}',
            '{"field": "framework", "headerName": "Framework", "width": 120}',
            '{"field": "requirement_id", "headerName": "ID Requisito", "width": 100}',
            '{"field": "requirement_name", "headerName": "Nome do Requisito", "width": 200}',
            '{"field": "economic_impact", "headerName": "Impacto (R$)", "width": 120, "type": "number", "valueFormatter": "currencyFormatter"}'
        ],
        'data_source', 'SELECT nc.open_x_domain, nc.framework, nc.requirement_id, nc.requirement_name, ei.impact_amount as economic_impact FROM compliance_validators.vw_open_x_non_compliance_details nc JOIN economic_planning.economic_impacts ei ON nc.tenant_id = ei.tenant_id AND nc.requirement_id = ei.validator_id WHERE nc.tenant_id = :tenant_id AND (:domains IS NULL OR nc.open_x_domain = ANY(:domains)) ORDER BY ei.impact_amount DESC LIMIT 10',
        'pagination', true,
        'page_size', 10,
        'sortable', true,
        'filterable', true
    ),
    ROW('table_high_impact_non_compliance', 'table', 0, 4, 6, 3, true, false, 6, 3)::compliance_dashboard_ui.layout_config
);

-- =============================================
-- Componentes do Dashboard por Jurisdição Open X
-- =============================================

-- Filtro de Tenant
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_tenant_jurisdiction',
    'open_x_jurisdiction_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'tenant_filter',
        'filter_type', 'dropdown',
        'label', 'Tenant',
        'placeholder', 'Selecione o tenant',
        'is_multi_select', false,
        'is_mandatory', true,
        'data_source', 'SELECT tenant_id, tenant_name FROM iam.tenants ORDER BY tenant_name'
    ),
    ROW('filter_tenant_jurisdiction', 'filter', 0, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Filtro de Jurisdição
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'filter_jurisdiction_select',
    'open_x_jurisdiction_dashboard',
    'filter',
    jsonb_build_object(
        'filter_id', 'jurisdiction_filter',
        'filter_type', 'dropdown',
        'label', 'Jurisdição',
        'placeholder', 'Selecione a jurisdição',
        'available_values', array['PORTUGAL_UE', 'BRASIL', 'ANGOLA', 'EUA'],
        'is_multi_select', false,
        'is_mandatory', true
    ),
    ROW('filter_jurisdiction_select', 'filter', 3, 0, 3, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Card de Conformidade da Jurisdição
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'card_jurisdiction_compliance',
    'open_x_jurisdiction_dashboard',
    'card',
    jsonb_build_object(
        'card_id', 'jurisdiction_compliance_card',
        'title', 'Conformidade Geral',
        'icon', 'public',
        'theme', 'primary',
        'data_source', 'SELECT AVG(compliance_percentage) FROM compliance_validators.vw_open_x_jurisdiction_compliance WHERE tenant_id = :tenant_id AND jurisdiction = :jurisdiction',
        'value_format', '{value}%',
        'show_trend', true,
        'threshold_warning', 75,
        'threshold_critical', 50
    ),
    ROW('card_jurisdiction_compliance', 'card', 0, 1, 2, 1, false, false, 2, 1)::compliance_dashboard_ui.layout_config
);

-- Gráfico de Conformidade por Domínio na Jurisdição
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'chart_jurisdiction_domains',
    'open_x_jurisdiction_dashboard',
    'chart',
    jsonb_build_object(
        'chart_id', 'jurisdiction_domains_chart',
        'chart_type', 'bar',
        'title', 'Conformidade por Domínio',
        'subtitle', 'Percentual de conformidade para cada domínio Open X na jurisdição selecionada',
        'x_axis_label', 'Domínio',
        'y_axis_label', 'Conformidade (%)',
        'color_palette', array['#4CAF50', '#2196F3', '#FFC107'],
        'show_legend', true,
        'is_interactive', true,
        'data_source', 'SELECT open_x_domain as label, AVG(compliance_percentage) as value FROM compliance_validators.vw_open_x_jurisdiction_compliance WHERE tenant_id = :tenant_id AND jurisdiction = :jurisdiction GROUP BY open_x_domain'
    ),
    ROW('chart_jurisdiction_domains', 'chart', 2, 1, 4, 2, true, false, 3, 2)::compliance_dashboard_ui.layout_config
);

-- Tabela de Frameworks por Jurisdição
INSERT INTO compliance_dashboard_ui.dashboard_components (
    component_id, dashboard_id, component_type, component_config, layout
) VALUES (
    'table_jurisdiction_frameworks',
    'open_x_jurisdiction_dashboard',
    'table',
    jsonb_build_object(
        'table_id', 'jurisdiction_frameworks_table',
        'title', 'Frameworks Regulatórios por Jurisdição',
        'columns', array[
            '{"field": "open_x_domain", "headerName": "Domínio", "width": 120}',
            '{"field": "framework", "headerName": "Framework", "width": 120}',
            '{"field": "total_requirements", "headerName": "Total Requisitos", "width": 120}',
            '{"field": "compliant_requirements", "headerName": "Req. Conformes", "width": 120}',
            '{"field": "compliance_percentage", "headerName": "Conformidade (%)", "width": 120, "type": "number", "valueFormatter": "percentFormatter"}',
            '{"field": "risk_level", "headerName": "Nível de Risco", "width": 120, "cellRenderer": "riskLevelRenderer"}'
        ],
        'data_source', 'SELECT open_x_domain, framework, total_requirements, compliant_requirements, compliance_percentage, risk_level FROM compliance_validators.vw_open_x_jurisdiction_compliance WHERE tenant_id = :tenant_id AND jurisdiction = :jurisdiction ORDER BY compliance_percentage DESC',
        'pagination', true,
        'page_size', 10,
        'sortable', true,
        'filterable', true
    ),
    ROW('table_jurisdiction_frameworks', 'table', 0, 3, 6, 3, true, false, 6, 3)::compliance_dashboard_ui.layout_config
);

-- =============================================
-- Funções de suporte para o dashboard
-- =============================================

-- Função para obter validações por domínio
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_domain_validations(
    p_tenant_id UUID,
    p_domain VARCHAR
)
RETURNS TABLE (
    framework_id VARCHAR,
    requirement_id VARCHAR,
    is_compliant BOOLEAN
)
AS $$
BEGIN
    IF p_domain = 'OPEN_INSURANCE' THEN
        RETURN QUERY
        SELECT v.framework_id, v.requirement_id, v.is_compliant
        FROM compliance_validators.open_insurance_validations v
        WHERE v.tenant_id = p_tenant_id;
    ELSIF p_domain = 'OPEN_HEALTH' THEN
        RETURN QUERY
        SELECT v.framework_id, v.requirement_id, v.is_compliant
        FROM compliance_validators.open_health_validations v
        WHERE v.tenant_id = p_tenant_id;
    ELSIF p_domain = 'OPEN_GOVERNMENT' THEN
        RETURN QUERY
        SELECT v.framework_id, v.requirement_id, v.is_compliant
        FROM compliance_validators.open_government_validations v
        WHERE v.tenant_id = p_tenant_id;
    ELSE
        RAISE EXCEPTION 'Domínio % não reconhecido', p_domain;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para obter requisitos por domínio
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_domain_requirements(
    p_domain VARCHAR
)
RETURNS TABLE (
    requirement_id VARCHAR,
    requirement_name VARCHAR,
    irr_threshold VARCHAR
)
AS $$
BEGIN
    IF p_domain = 'OPEN_INSURANCE' THEN
        RETURN QUERY
        SELECT r.requirement_id, r.requirement_name, r.irr_threshold
        FROM compliance_validators.open_insurance_requirements r;
    ELSIF p_domain = 'OPEN_HEALTH' THEN
        RETURN QUERY
        SELECT r.requirement_id, r.requirement_name, r.irr_threshold
        FROM compliance_validators.open_health_requirements r;
    ELSIF p_domain = 'OPEN_GOVERNMENT' THEN
        RETURN QUERY
        SELECT r.requirement_id, r.requirement_name, r.irr_threshold
        FROM compliance_validators.open_government_requirements r;
    ELSE
        RAISE EXCEPTION 'Domínio % não reconhecido', p_domain;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para simular ROI com diferentes níveis de melhoria
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_roi_simulations(
    p_tenant_id UUID,
    p_domains VARCHAR[] DEFAULT NULL,
    p_months_horizon INTEGER DEFAULT 12
)
RETURNS TABLE (
    improvement INTEGER,
    roi NUMERIC
)
AS $$
BEGIN
    RETURN QUERY
    WITH improvements AS (
        SELECT generate_series(10, 50, 10) AS improvement_pct
    ),
    simulations AS (
        SELECT
            imp.improvement_pct,
            s.domain,
            s.framework,
            s.current_compliance_percentage,
            s.simulated_compliance_percentage,
            s.current_economic_impact,
            s.simulated_economic_impact,
            s.potential_savings,
            s.estimated_roi
        FROM improvements imp
        CROSS JOIN LATERAL (
            SELECT * FROM compliance_validators.simulate_open_x_improvements(
                p_tenant_id,
                NULL,
                imp.improvement_pct,
                p_months_horizon
            ) sim
            WHERE p_domains IS NULL OR sim.domain = ANY(p_domains)
        ) s
    )
    SELECT
        improvement_pct AS improvement,
        AVG(estimated_roi) AS roi
    FROM simulations
    GROUP BY improvement_pct
    ORDER BY improvement_pct;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Configuração de Permissões
-- =============================================

-- Conceder permissões para os roles necessários
DO $$
BEGIN
    -- Verificar se os roles existem antes de conceder permissões
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'economic_analyst_role') THEN
        -- Permissões para analistas econômicos
        GRANT USAGE ON SCHEMA compliance_dashboard_ui TO economic_analyst_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_dashboard_ui TO economic_analyst_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'compliance_manager_role') THEN
        -- Permissões para gestores de conformidade
        GRANT USAGE ON SCHEMA compliance_dashboard_ui TO compliance_manager_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_dashboard_ui TO compliance_manager_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'risk_analyst_role') THEN
        -- Permissões para analistas de risco
        GRANT USAGE ON SCHEMA compliance_dashboard_ui TO risk_analyst_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_dashboard_ui TO risk_analyst_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dashboard_viewer_role') THEN
        -- Permissões para visualizadores do dashboard
        GRANT USAGE ON SCHEMA compliance_dashboard_ui TO dashboard_viewer_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_dashboard_ui TO dashboard_viewer_role;
    END IF;
END $$;

-- =============================================
-- Notificar Conclusão
-- =============================================

DO $$
BEGIN
    RAISE NOTICE 'Interface de Visualização do Dashboard Open X configurada com sucesso!';
    RAISE NOTICE 'Dashboards configurados:';
    RAISE NOTICE '- Dashboard Principal Open X';
    RAISE NOTICE '- Dashboard por Domínio Open X';
    RAISE NOTICE '- Dashboard de Impacto Econômico Open X';
    RAISE NOTICE '- Dashboard por Jurisdição Open X';
END $$;
