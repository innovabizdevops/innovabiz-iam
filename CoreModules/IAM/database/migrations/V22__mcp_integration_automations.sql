-- V22__mcp_integration_automations.sql
-- Implementação de automatismos para integração MCP, GraphQL e outras plataformas
-- Data: 2025-06-12
-- Autor: Sistema Innovabiz

-- 1. Criar schema para integrações MCP se não existir
CREATE SCHEMA IF NOT EXISTS mcp_integration;

-- 2. Tabela para registro de servidores MCP
CREATE TABLE IF NOT EXISTS mcp_integration.servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    endpoint_url VARCHAR(255) NOT NULL,
    description TEXT,
    server_type VARCHAR(50) NOT NULL, -- 'docker', 'memory', 'sequential-thinking', 'kubernetes', etc.
    api_version VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    health_check_interval VARCHAR(50) DEFAULT '*/5 * * * *', -- cron expression
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Tabela para ferramentas MCP disponíveis
CREATE TABLE IF NOT EXISTS mcp_integration.tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    server_id UUID REFERENCES mcp_integration.servers(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parameters JSONB,
    return_schema JSONB,
    example_usage TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Tabela para integrações GraphQL
CREATE TABLE IF NOT EXISTS mcp_integration.graphql_endpoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    endpoint_url VARCHAR(255) NOT NULL,
    schema TEXT,
    introspection_enabled BOOLEAN DEFAULT TRUE,
    authentication_type VARCHAR(50), -- 'jwt', 'oauth', 'api_key', etc.
    authentication_config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Tabela para integrações Krakend API Gateway
CREATE TABLE IF NOT EXISTS mcp_integration.krakend_endpoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    endpoint_name VARCHAR(100) NOT NULL,
    method VARCHAR(10) NOT NULL, -- 'GET', 'POST', etc.
    endpoint_url VARCHAR(255) NOT NULL,
    backend_urls JSONB NOT NULL,
    rate_limit INTEGER,
    timeout_seconds INTEGER DEFAULT 30,
    cache_ttl_seconds INTEGER,
    auth_required BOOLEAN DEFAULT TRUE,
    plugins JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Tabela para registro de integrações externas
CREATE TABLE IF NOT EXISTS mcp_integration.external_integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    integration_type VARCHAR(50) NOT NULL, -- 'hugging_face', 'redis', 'kafka', 'neo4j', 'snowflake', etc.
    config_parameters JSONB,
    health_check_query TEXT,
    health_status VARCHAR(20) DEFAULT 'unknown',
    last_health_check TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Função para registrar servidor MCP
CREATE OR REPLACE FUNCTION mcp_integration.register_mcp_server(
    p_name VARCHAR(100),
    p_endpoint_url VARCHAR(255),
    p_description TEXT,
    p_server_type VARCHAR(50)
)
RETURNS UUID AS $$
DECLARE
    v_server_id UUID;
BEGIN
    INSERT INTO mcp_integration.servers (name, endpoint_url, description, server_type)
    VALUES (p_name, p_endpoint_url, p_description, p_server_type)
    RETURNING id INTO v_server_id;
    
    RETURN v_server_id;
END;
$$ LANGUAGE plpgsql;

-- 8. Função para registrar ferramenta MCP
CREATE OR REPLACE FUNCTION mcp_integration.register_mcp_tool(
    p_server_id UUID,
    p_name VARCHAR(100),
    p_description TEXT,
    p_parameters JSONB,
    p_return_schema JSONB,
    p_example_usage TEXT
)
RETURNS UUID AS $$
DECLARE
    v_tool_id UUID;
BEGIN
    INSERT INTO mcp_integration.tools (server_id, name, description, parameters, return_schema, example_usage)
    VALUES (p_server_id, p_name, p_description, p_parameters, p_return_schema, p_example_usage)
    RETURNING id INTO v_tool_id;
    
    RETURN v_tool_id;
END;
$$ LANGUAGE plpgsql;

-- 9. Função para configurar endpoint GraphQL
CREATE OR REPLACE FUNCTION mcp_integration.register_graphql_endpoint(
    p_name VARCHAR(100),
    p_endpoint_url VARCHAR(255),
    p_schema TEXT DEFAULT NULL,
    p_auth_type VARCHAR(50) DEFAULT NULL,
    p_auth_config JSONB DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_endpoint_id UUID;
BEGIN
    INSERT INTO mcp_integration.graphql_endpoints (
        name, endpoint_url, schema, authentication_type, authentication_config
    )
    VALUES (p_name, p_endpoint_url, p_schema, p_auth_type, p_auth_config)
    RETURNING id INTO v_endpoint_id;
    
    RETURN v_endpoint_id;
END;
$$ LANGUAGE plpgsql;

-- 10. Função para configurar endpoint Krakend
CREATE OR REPLACE FUNCTION mcp_integration.register_krakend_endpoint(
    p_endpoint_name VARCHAR(100),
    p_method VARCHAR(10),
    p_endpoint_url VARCHAR(255),
    p_backend_urls JSONB,
    p_auth_required BOOLEAN DEFAULT TRUE,
    p_rate_limit INTEGER DEFAULT NULL,
    p_timeout_seconds INTEGER DEFAULT 30,
    p_cache_ttl_seconds INTEGER DEFAULT NULL,
    p_plugins JSONB DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_endpoint_id UUID;
BEGIN
    INSERT INTO mcp_integration.krakend_endpoints (
        endpoint_name, method, endpoint_url, backend_urls,
        auth_required, rate_limit, timeout_seconds, cache_ttl_seconds, plugins
    )
    VALUES (
        p_endpoint_name, p_method, p_endpoint_url, p_backend_urls,
        p_auth_required, p_rate_limit, p_timeout_seconds, p_cache_ttl_seconds, p_plugins
    )
    RETURNING id INTO v_endpoint_id;
    
    RETURN v_endpoint_id;
END;
$$ LANGUAGE plpgsql;

-- 11. Função para registrar integração externa
CREATE OR REPLACE FUNCTION mcp_integration.register_external_integration(
    p_name VARCHAR(100),
    p_integration_type VARCHAR(50),
    p_config_parameters JSONB,
    p_health_check_query TEXT DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_integration_id UUID;
BEGIN
    INSERT INTO mcp_integration.external_integrations (
        name, integration_type, config_parameters, health_check_query
    )
    VALUES (p_name, p_integration_type, p_config_parameters, p_health_check_query)
    RETURNING id INTO v_integration_id;
    
    RETURN v_integration_id;
END;
$$ LANGUAGE plpgsql;

-- 12. Função para verificar saúde das integrações
CREATE OR REPLACE FUNCTION mcp_integration.check_integration_health()
RETURNS TABLE (
    integration_name VARCHAR(100),
    integration_type VARCHAR(50),
    health_status VARCHAR(20),
    last_checked TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    -- Esta função seria expandida para fazer verificações reais
    -- Por enquanto, apenas simula a atualização do status
    UPDATE mcp_integration.external_integrations
    SET health_status = 
        CASE 
            WHEN random() > 0.9 THEN 'degraded'
            WHEN random() > 0.95 THEN 'offline'
            ELSE 'healthy'
        END,
        last_health_check = NOW();
        
    RETURN QUERY
    SELECT name, integration_type, health_status, last_health_check
    FROM mcp_integration.external_integrations;
END;
$$ LANGUAGE plpgsql;

-- 13. Registrar servidores MCP padrão conforme regras da plataforma
DO $$
DECLARE
    v_docker_mcp UUID;
    v_memory_mcp UUID;
    v_sequential_thinking_mcp UUID;
    v_figures_mcp UUID;
    v_filesystem_mcp UUID;
    v_github_mcp UUID;
    v_kubernetes_mcp UUID;
BEGIN
    -- Registrar servidores MCP
    v_docker_mcp := mcp_integration.register_mcp_server(
        'MCP_DOCKER',
        'http://localhost:8080/mcp/docker',
        'Docker MCP server para automação de contêineres',
        'docker'
    );
    
    v_memory_mcp := mcp_integration.register_mcp_server(
        'memory',
        'http://localhost:8081/mcp/memory',
        'Memory MCP server para gerenciamento de memória persistente',
        'memory'
    );
    
    v_sequential_thinking_mcp := mcp_integration.register_mcp_server(
        'sequential-thinking',
        'http://localhost:8082/mcp/sequential-thinking',
        'Sequential Thinking MCP server para análise progressiva e passo-a-passo',
        'sequential-thinking'
    );
    
    v_figures_mcp := mcp_integration.register_mcp_server(
        'figma',
        'http://localhost:8083/mcp/figma',
        'Figma MCP server para integração com design UI/UX',
        'figma'
    );
    
    v_filesystem_mcp := mcp_integration.register_mcp_server(
        'filesystem',
        'http://localhost:8084/mcp/filesystem',
        'Filesystem MCP server para operações com arquivos',
        'filesystem'
    );
    
    v_github_mcp := mcp_integration.register_mcp_server(
        'github',
        'http://localhost:8085/mcp/github',
        'GitHub MCP server para integração com repositórios',
        'github'
    );
    
    v_kubernetes_mcp := mcp_integration.register_mcp_server(
        'kubernetes',
        'http://localhost:8086/mcp/kubernetes',
        'Kubernetes MCP server para orquestração de contêineres',
        'kubernetes'
    );
    
    -- Registrar algumas ferramentas exemplo para cada servidor
    PERFORM mcp_integration.register_mcp_tool(
        v_docker_mcp,
        'mcp0_docker',
        'Use the docker cli',
        '{"args": {"description": "Arguments to pass to the Docker command", "type": "array", "items": {"type": "string"}}}',
        '{"result": {"description": "Docker command output", "type": "string"}}',
        'Example: mcp0_docker(["ps", "-a"])'
    );
    
    PERFORM mcp_integration.register_mcp_tool(
        v_memory_mcp,
        'mcp10_read_graph',
        'Read the entire knowledge graph',
        '{}',
        '{"nodes": {"type": "array"}, "edges": {"type": "array"}}',
        'Example: mcp10_read_graph()'
    );
    
    -- Registrar algumas integrações GraphQL
    PERFORM mcp_integration.register_graphql_endpoint(
        'IAM_GraphQL',
        'http://localhost:4000/graphql',
        NULL, -- schema seria obtido por introspection
        'jwt',
        '{"header": "Authorization", "prefix": "Bearer "}'
    );
    
    -- Registrar alguns endpoints Krakend
    PERFORM mcp_integration.register_krakend_endpoint(
        'iam_users_endpoint',
        'GET',
        '/api/v1/users',
        '[{"url": "http://iam-service:8000/users"}, {"url": "http://backup-iam-service:8000/users"}]',
        TRUE, -- auth_required
        100,  -- rate_limit
        30,   -- timeout_seconds
        60,   -- cache_ttl_seconds
        '{"cors": {"allow_origins": ["*"], "allow_methods": ["GET", "POST"]}, "logger": {"level": "INFO"}}'
    );
    
    -- Registrar algumas integrações externas
    PERFORM mcp_integration.register_external_integration(
        'PostgreSQL_Main',
        'postgresql',
        '{"host": "db.example.com", "port": 5432, "database": "innovabiz", "user": "db_user"}',
        'SELECT 1'
    );
    
    PERFORM mcp_integration.register_external_integration(
        'Redis_Cache',
        'redis',
        '{"host": "redis.example.com", "port": 6379, "database": 0}',
        'PING'
    );
    
    PERFORM mcp_integration.register_external_integration(
        'Kafka_Events',
        'kafka',
        '{"bootstrap_servers": ["kafka1.example.com:9092", "kafka2.example.com:9092"], "client_id": "innovabiz-app"}',
        NULL
    );
    
    PERFORM mcp_integration.register_external_integration(
        'Neo4j_Graph',
        'neo4j',
        '{"uri": "bolt://neo4j.example.com:7687", "database": "innovabiz-graph"}',
        'RETURN 1'
    );
    
    PERFORM mcp_integration.register_external_integration(
        'Snowflake_Analytics',
        'snowflake',
        '{"account": "innovabiz.eu-central-1", "warehouse": "compute_wh", "database": "ANALYTICS"}',
        'SELECT 1'
    );
END
$$;
