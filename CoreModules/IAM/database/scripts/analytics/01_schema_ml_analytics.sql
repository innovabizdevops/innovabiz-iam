-- INNOVABIZ - IAM ML/Analytics Extension
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Extensão para análise preditiva e integração com ML/BI no módulo IAM.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Criar schema para analytics
CREATE SCHEMA IF NOT EXISTS iam_analytics;

-- Alterar caminho de busca para incluir o novo schema
SET search_path TO iam_analytics, iam, public;

-- ===============================================================================
-- TABELAS PARA ARMAZENAMENTO DE DADOS ANALÍTICOS
-- ===============================================================================

-- Tabela para Feature Store - armazenamento de características para modelos ML
CREATE TABLE IF NOT EXISTS ml_feature_store (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_set_name VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    feature_data JSONB NOT NULL,
    feature_vector FLOAT[] NULL,
    embedding_model VARCHAR(255) NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_to TIMESTAMP WITH TIME ZONE DEFAULT 'infinity'::TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES iam.users(id),
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_ml_feature_store_feature_set ON ml_feature_store(feature_set_name);
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_entity ON ml_feature_store(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_validity ON ml_feature_store(valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_version ON ml_feature_store(feature_set_name, entity_id, version);

-- Tabela para armazenar definições de modelos de ML
CREATE TABLE IF NOT EXISTS ml_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    model_type VARCHAR(50) NOT NULL,
    target_variable VARCHAR(100) NOT NULL,
    features JSONB NOT NULL,
    hyperparameters JSONB,
    trained_at TIMESTAMP WITH TIME ZONE,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    accuracy FLOAT,
    performance_metrics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id),
    deployment_config JSONB,
    is_active BOOLEAN DEFAULT FALSE,
    model_uri VARCHAR(255),
    CONSTRAINT ml_models_status_valid_values CHECK (status IN ('draft', 'training', 'trained', 'deployed', 'archived', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_ml_models_status ON ml_models(status);
CREATE INDEX IF NOT EXISTS idx_ml_models_model_type ON ml_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ml_models_is_active ON ml_models(is_active);

-- Tabela para armazenar predições de modelos
CREATE TABLE IF NOT EXISTS ml_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id UUID NOT NULL REFERENCES ml_models(id),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    prediction_result JSONB NOT NULL,
    confidence_score FLOAT,
    explanation JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    input_features JSONB,
    version VARCHAR(50),
    is_feedback_provided BOOLEAN DEFAULT FALSE,
    feedback_correct BOOLEAN,
    feedback_details TEXT,
    feedback_provided_by UUID REFERENCES iam.users(id),
    feedback_timestamp TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ml_predictions_model ON ml_predictions(model_id);
CREATE INDEX IF NOT EXISTS idx_ml_predictions_entity ON ml_predictions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_predictions_timestamp ON ml_predictions(timestamp);
CREATE INDEX IF NOT EXISTS idx_ml_predictions_feedback ON ml_predictions(is_feedback_provided);

-- Tabela para pipeline de ML (ETL + treinamento)
CREATE TABLE IF NOT EXISTS ml_pipelines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    pipeline_type VARCHAR(50) NOT NULL,
    schedule_expression VARCHAR(100),
    last_run_timestamp TIMESTAMP WITH TIME ZONE,
    next_run_timestamp TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    steps JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id),
    is_active BOOLEAN DEFAULT FALSE,
    parameters JSONB,
    notification_emails TEXT[],
    CONSTRAINT ml_pipelines_status_valid_values CHECK (status IN ('inactive', 'active', 'running', 'failed', 'completed'))
);

CREATE INDEX IF NOT EXISTS idx_ml_pipelines_status ON ml_pipelines(status);
CREATE INDEX IF NOT EXISTS idx_ml_pipelines_pipeline_type ON ml_pipelines(pipeline_type);
CREATE INDEX IF NOT EXISTS idx_ml_pipelines_schedule ON ml_pipelines(next_run_timestamp);

-- Tabela para execuções de pipeline
CREATE TABLE IF NOT EXISTS ml_pipeline_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_id UUID NOT NULL REFERENCES ml_pipelines(id),
    start_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_timestamp TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    step_results JSONB,
    error_details TEXT,
    metrics JSONB,
    artifacts JSONB,
    trigger_type VARCHAR(50) NOT NULL,
    triggered_by UUID REFERENCES iam.users(id)
);

CREATE INDEX IF NOT EXISTS idx_ml_pipeline_runs_pipeline ON ml_pipeline_runs(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_ml_pipeline_runs_status ON ml_pipeline_runs(status);
CREATE INDEX IF NOT EXISTS idx_ml_pipeline_runs_timestamp ON ml_pipeline_runs(start_timestamp);

-- ===============================================================================
-- TABELAS PARA INTEGRAÇÃO BI
-- ===============================================================================

-- Tabela para definições de dashboards
CREATE TABLE IF NOT EXISTS bi_dashboards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    dashboard_type VARCHAR(50) NOT NULL,
    layout_config JSONB,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id),
    owner_id UUID REFERENCES iam.users(id),
    organization_id UUID REFERENCES iam.organizations(id),
    access_control JSONB, -- Configuração de quem pode acessar
    theme_settings JSONB,
    last_refreshed_at TIMESTAMP WITH TIME ZONE,
    refresh_schedule VARCHAR(100),
    embedded_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_bi_dashboards_type ON bi_dashboards(dashboard_type);
CREATE INDEX IF NOT EXISTS idx_bi_dashboards_organization ON bi_dashboards(organization_id);
CREATE INDEX IF NOT EXISTS idx_bi_dashboards_is_public ON bi_dashboards(is_public);

-- Tabela para visualizações dentro dos dashboards
CREATE TABLE IF NOT EXISTS bi_visualizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dashboard_id UUID REFERENCES bi_dashboards(id),
    name VARCHAR(100) NOT NULL,
    visualization_type VARCHAR(50) NOT NULL,
    query_definition JSONB NOT NULL,
    position_config JSONB, -- Posição no layout
    styling_config JSONB,
    data_mapping JSONB, -- Como mapear dados para visual
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id),
    data_source VARCHAR(100),
    last_refreshed_at TIMESTAMP WITH TIME ZONE,
    refresh_interval INTEGER, -- Em minutos
    parameters JSONB,
    drill_down_config JSONB
);

CREATE INDEX IF NOT EXISTS idx_bi_visualizations_dashboard ON bi_visualizations(dashboard_id);
CREATE INDEX IF NOT EXISTS idx_bi_visualizations_type ON bi_visualizations(visualization_type);
CREATE INDEX IF NOT EXISTS idx_bi_visualizations_is_active ON bi_visualizations(is_active);

-- Tabela para configuração de alertas baseados em data insights
CREATE TABLE IF NOT EXISTS bi_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    visualization_id UUID REFERENCES bi_visualizations(id),
    alert_condition JSONB NOT NULL, -- Condição para disparar alerta
    alert_threshold FLOAT NOT NULL,
    alert_operator VARCHAR(20) NOT NULL, -- gt, lt, eq, etc.
    is_active BOOLEAN DEFAULT TRUE,
    recipients JSONB, -- Lista de destinatários
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id),
    cooldown_period INTEGER, -- Período mínimo entre alertas em minutos
    severity VARCHAR(20) DEFAULT 'medium',
    message_template TEXT,
    notification_channels TEXT[]
);

CREATE INDEX IF NOT EXISTS idx_bi_alerts_visualization ON bi_alerts(visualization_id);
CREATE INDEX IF NOT EXISTS idx_bi_alerts_is_active ON bi_alerts(is_active);
CREATE INDEX IF NOT EXISTS idx_bi_alerts_severity ON bi_alerts(severity);

-- Tabela para métricas pré-calculadas (cubo OLAP simplificado)
CREATE TABLE IF NOT EXISTS bi_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    time_dimension DATE NOT NULL,
    dimensions JSONB, -- Dimensões para slice & dice
    measures JSONB NOT NULL, -- Medidas numéricas
    is_aggregated BOOLEAN DEFAULT FALSE,
    aggregation_level VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source_query TEXT,
    refresh_frequency VARCHAR(50),
    last_refreshed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_bi_metrics_metric_name ON bi_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_bi_metrics_entity ON bi_metrics(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_bi_metrics_time_dimension ON bi_metrics(time_dimension);
CREATE INDEX IF NOT EXISTS idx_bi_metrics_dimensions ON bi_metrics USING GIN(dimensions);

-- ===============================================================================
-- RECURSOS DE DATA SCIENCE PARA ANÁLISE PREDITIVA
-- ===============================================================================

-- Extensão para cálculos estatísticos e análise avançada
CREATE EXTENSION IF NOT EXISTS tablefunc;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Extensão para suporte a vetores (para machine learning)
CREATE EXTENSION IF NOT EXISTS vector;

-- Tabela para armazenar embeddings (representações vetoriais)
CREATE TABLE IF NOT EXISTS ml_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    embedding_type VARCHAR(100) NOT NULL,
    embedding_model VARCHAR(255) NOT NULL,
    embedding_vector vector(1536), -- Dimensão típica para embeddings modernos
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ml_embeddings_entity ON ml_embeddings(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_embeddings_type_model ON ml_embeddings(embedding_type, embedding_model);
CREATE INDEX IF NOT EXISTS idx_ml_embeddings_vector ON ml_embeddings USING ivfflat (embedding_vector vector_cosine_ops);

-- Tabela para armazenar clusters e segmentos
CREATE TABLE IF NOT EXISTS ml_clusters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cluster_model_name VARCHAR(100) NOT NULL,
    cluster_number INTEGER NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    confidence_score FLOAT,
    distance_to_centroid FLOAT,
    cluster_description TEXT,
    features_importance JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    model_version VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_ml_clusters_model ON ml_clusters(cluster_model_name, model_version);
CREATE INDEX IF NOT EXISTS idx_ml_clusters_cluster ON ml_clusters(cluster_number);
CREATE INDEX IF NOT EXISTS idx_ml_clusters_entity ON ml_clusters(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_clusters_is_active ON ml_clusters(is_active);

-- Tabela para armazenar anomalias detectadas
CREATE TABLE IF NOT EXISTS ml_anomalies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    anomaly_type VARCHAR(100) NOT NULL,
    detection_model VARCHAR(100) NOT NULL,
    anomaly_score FLOAT NOT NULL,
    anomaly_threshold FLOAT NOT NULL,
    features_contribution JSONB,
    detection_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_result VARCHAR(50),
    verified_by UUID REFERENCES iam.users(id),
    verification_notes TEXT,
    verification_timestamp TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'open',
    resolution_details JSONB,
    CONSTRAINT ml_anomalies_status_valid_values CHECK (status IN ('open', 'investigating', 'resolved', 'false_positive', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_ml_anomalies_entity ON ml_anomalies(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_anomalies_type ON ml_anomalies(anomaly_type);
CREATE INDEX IF NOT EXISTS idx_ml_anomalies_detection_timestamp ON ml_anomalies(detection_timestamp);
CREATE INDEX IF NOT EXISTS idx_ml_anomalies_status ON ml_anomalies(status);
CREATE INDEX IF NOT EXISTS idx_ml_anomalies_score ON ml_anomalies(anomaly_score);

-- Tabela para recomendações geradas por ML
CREATE TABLE IF NOT EXISTS ml_recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_user_id UUID REFERENCES iam.users(id),
    target_entity_type VARCHAR(100),
    target_entity_id UUID,
    recommendation_type VARCHAR(100) NOT NULL,
    recommendation_model VARCHAR(100) NOT NULL,
    recommendation_content JSONB NOT NULL,
    relevance_score FLOAT NOT NULL,
    explanation TEXT,
    is_viewed BOOLEAN DEFAULT FALSE,
    is_applied BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    feedback VARCHAR(50),
    feedback_details TEXT,
    feedback_timestamp TIMESTAMP WITH TIME ZONE,
    context JSONB
);

CREATE INDEX IF NOT EXISTS idx_ml_recommendations_target_user ON ml_recommendations(target_user_id);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_target_entity ON ml_recommendations(target_entity_type, target_entity_id);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_type ON ml_recommendations(recommendation_type);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_is_viewed ON ml_recommendations(is_viewed);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_is_applied ON ml_recommendations(is_applied);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_created_at ON ml_recommendations(created_at);
CREATE INDEX IF NOT EXISTS idx_ml_recommendations_expires_at ON ml_recommendations(expires_at);

-- ===============================================================================
-- FUNÇÕES E PROCEDIMENTOS PARA ML/BI
-- ===============================================================================

-- Função para registrar métricas de BI
CREATE OR REPLACE FUNCTION register_bi_metric(
    p_metric_name VARCHAR(100),
    p_entity_type VARCHAR(100),
    p_entity_id UUID,
    p_time_dimension DATE,
    p_dimensions JSONB,
    p_measures JSONB
) RETURNS UUID AS $$
DECLARE
    v_metric_id UUID;
BEGIN
    -- Verificar se já existe uma métrica com os mesmos parâmetros
    SELECT id INTO v_metric_id
    FROM bi_metrics
    WHERE metric_name = p_metric_name
      AND entity_type = p_entity_type
      AND entity_id = p_entity_id
      AND time_dimension = p_time_dimension
      AND dimensions = p_dimensions;
      
    -- Se existir, atualizar as medidas
    IF v_metric_id IS NOT NULL THEN
        UPDATE bi_metrics
        SET measures = p_measures,
            last_refreshed_at = NOW()
        WHERE id = v_metric_id;
    ELSE
        -- Senão, inserir nova métrica
        INSERT INTO bi_metrics (
            metric_name,
            entity_type,
            entity_id,
            time_dimension,
            dimensions,
            measures,
            last_refreshed_at
        ) VALUES (
            p_metric_name,
            p_entity_type,
            p_entity_id,
            p_time_dimension,
            p_dimensions,
            p_measures,
            NOW()
        ) RETURNING id INTO v_metric_id;
    END IF;
    
    RETURN v_metric_id;
END;
$$ LANGUAGE plpgsql;

-- Função para salvar embeddings
CREATE OR REPLACE FUNCTION save_entity_embedding(
    p_entity_type VARCHAR(100),
    p_entity_id UUID,
    p_embedding_type VARCHAR(100),
    p_embedding_model VARCHAR(255),
    p_embedding_vector FLOAT[],
    p_metadata JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_embedding_id UUID;
    v_vector vector(1536);
BEGIN
    -- Converter array para tipo vector
    v_vector = p_embedding_vector::vector;
    
    -- Verificar se já existe um embedding para esta entidade/modelo
    SELECT id INTO v_embedding_id
    FROM ml_embeddings
    WHERE entity_type = p_entity_type
      AND entity_id = p_entity_id
      AND embedding_type = p_embedding_type
      AND embedding_model = p_embedding_model;
      
    -- Se existir, atualizar
    IF v_embedding_id IS NOT NULL THEN
        UPDATE ml_embeddings
        SET embedding_vector = v_vector,
            metadata = COALESCE(p_metadata, metadata),
            updated_at = NOW()
        WHERE id = v_embedding_id;
    ELSE
        -- Senão, inserir novo
        INSERT INTO ml_embeddings (
            entity_type,
            entity_id,
            embedding_type,
            embedding_model,
            embedding_vector,
            metadata
        ) VALUES (
            p_entity_type,
            p_entity_id,
            p_embedding_type,
            p_embedding_model,
            v_vector,
            p_metadata
        ) RETURNING id INTO v_embedding_id;
    END IF;
    
    RETURN v_embedding_id;
END;
$$ LANGUAGE plpgsql;

-- Função para encontrar entidades similares por embedding
CREATE OR REPLACE FUNCTION find_similar_entities(
    p_entity_type VARCHAR(100),
    p_entity_id UUID,
    p_embedding_type VARCHAR(100),
    p_embedding_model VARCHAR(255),
    p_limit INTEGER DEFAULT 10,
    p_threshold FLOAT DEFAULT 0.7
) RETURNS TABLE(
    entity_id UUID,
    similarity FLOAT,
    metadata JSONB
) AS $$
DECLARE
    v_query_vector vector(1536);
BEGIN
    -- Obter o vetor da entidade de consulta
    SELECT embedding_vector INTO v_query_vector
    FROM ml_embeddings
    WHERE entity_type = p_entity_type
      AND entity_id = p_entity_id
      AND embedding_type = p_embedding_type
      AND embedding_model = p_embedding_model;
    
    IF v_query_vector IS NULL THEN
        RAISE EXCEPTION 'Embedding não encontrado para a entidade especificada';
    END IF;
    
    -- Encontrar entidades similares
    RETURN QUERY
    SELECT 
        e.entity_id,
        (1 - (e.embedding_vector <=> v_query_vector)) AS similarity,
        e.metadata
    FROM 
        ml_embeddings e
    WHERE 
        e.entity_type = p_entity_type
        AND e.entity_id != p_entity_id
        AND e.embedding_type = p_embedding_type
        AND e.embedding_model = p_embedding_model
        AND (1 - (e.embedding_vector <=> v_query_vector)) >= p_threshold
    ORDER BY 
        similarity DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Função para registrar anomalias
CREATE OR REPLACE FUNCTION register_anomaly(
    p_entity_type VARCHAR(100),
    p_entity_id UUID,
    p_anomaly_type VARCHAR(100),
    p_detection_model VARCHAR(100),
    p_anomaly_score FLOAT,
    p_anomaly_threshold FLOAT,
    p_features_contribution JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_anomaly_id UUID;
BEGIN
    -- Inserir nova anomalia
    INSERT INTO ml_anomalies (
        entity_type,
        entity_id,
        anomaly_type,
        detection_model,
        anomaly_score,
        anomaly_threshold,
        features_contribution,
        detection_timestamp,
        status
    ) VALUES (
        p_entity_type,
        p_entity_id,
        p_anomaly_type,
        p_detection_model,
        p_anomaly_score,
        p_anomaly_threshold,
        p_features_contribution,
        NOW(),
        'open'
    ) RETURNING id INTO v_anomaly_id;
    
    -- Criar notificação para a anomalia detectada
    INSERT INTO iam.notifications (
        notification_type,
        title,
        message,
        target_users,
        priority,
        entity_type,
        entity_id,
        status
    )
    SELECT
        'anomaly_detected',
        'Anomalia detectada: ' || p_anomaly_type,
        'Uma anomalia foi detectada no ' || p_entity_type || ' com ID ' || p_entity_id || 
        '. Score de anomalia: ' || p_anomaly_score || ' (limiar: ' || p_anomaly_threshold || ')',
        ARRAY_AGG(u.id),
        CASE 
            WHEN p_anomaly_score > p_anomaly_threshold * 1.5 THEN 'high'
            WHEN p_anomaly_score > p_anomaly_threshold * 1.2 THEN 'medium'
            ELSE 'low'
        END,
        'ml_anomalies',
        v_anomaly_id,
        'pending'
    FROM 
        iam.users u
        JOIN iam.roles r ON u.role_id = r.id
    WHERE 
        r.name IN ('Admin', 'Security Officer', 'Data Scientist')
    GROUP BY 
        1, 2, 3, 5, 6, 7, 8;
    
    RETURN v_anomaly_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar uma recomendação 
CREATE OR REPLACE FUNCTION create_recommendation(
    p_target_user_id UUID,
    p_target_entity_type VARCHAR(100),
    p_target_entity_id UUID,
    p_recommendation_type VARCHAR(100),
    p_recommendation_model VARCHAR(100),
    p_recommendation_content JSONB,
    p_relevance_score FLOAT,
    p_explanation TEXT DEFAULT NULL,
    p_expires_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_context JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_recommendation_id UUID;
BEGIN
    -- Inserir nova recomendação
    INSERT INTO ml_recommendations (
        target_user_id,
        target_entity_type,
        target_entity_id,
        recommendation_type,
        recommendation_model,
        recommendation_content,
        relevance_score,
        explanation,
        expires_at,
        context
    ) VALUES (
        p_target_user_id,
        p_target_entity_type,
        p_target_entity_id,
        p_recommendation_type,
        p_recommendation_model,
        p_recommendation_content,
        p_relevance_score,
        p_explanation,
        COALESCE(p_expires_at, NOW() + INTERVAL '30 days'),
        p_context
    ) RETURNING id INTO v_recommendation_id;
    
    RETURN v_recommendation_id;
END;
$$ LANGUAGE plpgsql;

-- Comentários nas tabelas
COMMENT ON TABLE ml_feature_store IS 'Armazena características (features) extraídas para modelos de machine learning';
COMMENT ON TABLE ml_models IS 'Registros de modelos de machine learning, incluindo metadados e configurações';
COMMENT ON TABLE ml_predictions IS 'Armazena predições geradas pelos modelos de ML para diferentes entidades';
COMMENT ON TABLE ml_pipelines IS 'Definições de pipelines de ML para processamento de dados e treinamento automático';
COMMENT ON TABLE ml_pipeline_runs IS 'Registros de execuções de pipelines de ML, incluindo resultados e métricas';
COMMENT ON TABLE bi_dashboards IS 'Definições de dashboards de BI para visualização de insights';
COMMENT ON TABLE bi_visualizations IS 'Componentes visuais dentro dos dashboards de BI';
COMMENT ON TABLE bi_alerts IS 'Configurações de alertas baseados em dados para notificações automáticas';
COMMENT ON TABLE bi_metrics IS 'Métricas pré-calculadas para análise dimensional (OLAP simplificado)';
COMMENT ON TABLE ml_embeddings IS 'Representações vetoriais de entidades para análise de similaridade e modelos deep learning';
COMMENT ON TABLE ml_clusters IS 'Resultados de algoritmos de clustering para segmentação';
COMMENT ON TABLE ml_anomalies IS 'Anomalias detectadas por modelos de detecção de outliers';
COMMENT ON TABLE ml_recommendations IS 'Recomendações personalizadas geradas por modelos de ML';
