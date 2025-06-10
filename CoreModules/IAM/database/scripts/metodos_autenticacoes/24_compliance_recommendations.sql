-- Script de Gerenciamento de Recomendações de Conformidade - IAM Open X
-- Versão: 1.0
-- Data: 15/05/2025

-- 1. Tabelas de Recomendações

-- Tabela de Recomendações de Conformidade
CREATE TABLE IF NOT EXISTS iam_access_control.compliance_recommendations (
    id BIGSERIAL PRIMARY KEY,
    username TEXT REFERENCES iam_access_control.users(username),
    recommendation_type TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by TEXT,
    resolution_notes TEXT
);

-- Tabela de Histórico de Ações
CREATE TABLE IF NOT EXISTS iam_access_control.compliance_actions_history (
    id BIGSERIAL PRIMARY KEY,
    recommendation_id BIGINT REFERENCES iam_access_control.compliance_recommendations(id),
    action_type TEXT NOT NULL,
    action_by TEXT NOT NULL,
    action_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    action_details JSONB
);
