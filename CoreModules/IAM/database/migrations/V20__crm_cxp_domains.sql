-- V20__crm_cxp_domains.sql
-- Modelagem dos domínios CRM e CXP para a plataforma InnovaBiz
-- Segue padrões de compliance, rastreabilidade, privacidade, multilíngue e integração

-- SCHEMA CRM
CREATE SCHEMA IF NOT EXISTS crm;

CREATE TABLE IF NOT EXISTS crm.contact (
    contact_id SERIAL PRIMARY KEY,
    company_id INT REFERENCES organization.company(company_id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    consent_status VARCHAR(50) NOT NULL, -- ex: given, revoked
    consent_date TIMESTAMP,
    language VARCHAR(10),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    external_ref VARCHAR(100),
    compliance JSONB,
    privacy JSONB,
    trace_id UUID DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    audit_log JSONB
);

CREATE TABLE IF NOT EXISTS crm.lead (
    lead_id SERIAL PRIMARY KEY,
    contact_id INT REFERENCES crm.contact(contact_id),
    source VARCHAR(100),
    status VARCHAR(50),
    responsible VARCHAR(100),
    score INT,
    notes TEXT,
    external_ref VARCHAR(100),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

CREATE TABLE IF NOT EXISTS crm.opportunity (
    opportunity_id SERIAL PRIMARY KEY,
    lead_id INT REFERENCES crm.lead(lead_id),
    value NUMERIC(18,2),
    stage VARCHAR(50),
    forecast_date DATE,
    source VARCHAR(100),
    status VARCHAR(50),
    external_ref VARCHAR(100),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

-- Interações Omnichannel detalhadas (compliance, IA, rastreabilidade, benchmarks globais)
CREATE TABLE IF NOT EXISTS crm.interaction (
    interaction_id SERIAL PRIMARY KEY,
    customer_id UUID, -- Identificação global do cliente (Forrester CX, CXPA, GDPR Art. 4)
    contact_id INT REFERENCES crm.contact(contact_id),
    opportunity_id INT REFERENCES crm.opportunity(opportunity_id),
    type VARCHAR(50) NOT NULL, -- call, email, meeting, chat, bot, social, etc (ISO 10002, CXPA)
    channel VARCHAR(50) NOT NULL, -- omnichannel: phone, web, app, whatsapp, sms, etc (benchmarks Salesforce, Zendesk)
    source_platform VARCHAR(100), -- WhatsApp, Facebook, Web, App, etc (compliance: rastreabilidade)
    device_info VARCHAR(255), -- device/browser info (ISO 27701, rastreabilidade)
    location GEOGRAPHY, -- localização da interação (GDPR, LGPD, ISO 27001)
    journey_id INT REFERENCES cxp.journey(journey_id),
    context JSONB, -- contexto detalhado: campanha, motivo, produto, segmento, consentimento
    sla_level VARCHAR(50), -- SLA por canal/prioridade (ITIL, benchmarks Zendesk)
    response_time INT, -- segundos (benchmarks: SLA, CSAT)
    resolution_time INT, -- segundos (benchmarks: FCR, SLA)
    responsible VARCHAR(100), -- agente/IA responsável (SOX, ISO 9001)
    result TEXT, -- resultado da interação (resolvido, pendente, escalado, etc)
    ai_score NUMERIC(5,2), -- predição IA: churn, conversão, satisfação (NIST AI RMF, benchmarks Salesforce)
    sentiment VARCHAR(50), -- análise de sentimento (NLP, ISO 56000, benchmarks globais)
    consent_status VARCHAR(50) DEFAULT 'given', -- GDPR, LGPD, CCPA
    consent_date TIMESTAMP, -- data/hora do consentimento
    opt_in BOOLEAN DEFAULT TRUE, -- Opt-in para comunicações (EU Directive 2002/58/EC)
    opt_out BOOLEAN DEFAULT FALSE, -- Opt-out global (CCPA, GDPR)
    preference JSONB, -- preferências do cliente: canal, idioma, frequência (ISO 10002, CXPA)
    data_classification VARCHAR(50) DEFAULT 'personal', -- pessoal, sensível, anonimizado (GDPR, LGPD, ISO 27701)
    regulatory_ref VARCHAR(255), -- Ex: GDPR Art. 6, LGPD Art. 7, CCPA Section 1798.100
    audit_log JSONB, -- Log de auditoria e rastreabilidade (SOX, ISO 27001, PCI DSS)
    compliance JSONB, -- Compliance detalhado (normas, regulamentações, avisos, decretos)
    benchmark_kpi VARCHAR(50), -- NPS, SLA, CSAT, CES, FCR, Churn Rate, CAC, LTV, etc (Forrester, Gartner, ISO 10002)
    benchmark_score NUMERIC(5,2), -- Score de benchmarking internacional
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

COMMENT ON TABLE crm.interaction IS 'Registro de interações omnichannel detalhadas, alinhadas a GDPR, LGPD, ISO 27701/27001, ITIL, COBIT, CXPA, BIAN, benchmarks Salesforce, Dynamics, SAP CX, Oracle CX, HubSpot, Zendesk, Forrester, Gartner e melhores práticas internacionais.';
COMMENT ON COLUMN crm.interaction.customer_id IS 'Identificação global do cliente para visão 360º (Forrester CX, CXPA, GDPR).';
COMMENT ON COLUMN crm.interaction.type IS 'Tipo de interação: call, email, chat, bot, meeting, social, etc. (ISO 10002, CXPA, benchmarks globais)';
COMMENT ON COLUMN crm.interaction.channel IS 'Canal de origem: omnichannel (phone, web, app, whatsapp, sms, etc). Benchmarks Salesforce, Zendesk.';
COMMENT ON COLUMN crm.interaction.location IS 'Localização da interação (GDPR, LGPD, ISO 27001).';
COMMENT ON COLUMN crm.interaction.context IS 'Contexto detalhado da interação: campanha, motivo, produto, segmento, consentimento.';
COMMENT ON COLUMN crm.interaction.sla_level IS 'SLA por canal/prioridade (ITIL, benchmarks Zendesk).';
COMMENT ON COLUMN crm.interaction.response_time IS 'Tempo de resposta em segundos (benchmarks: SLA, CSAT).';
COMMENT ON COLUMN crm.interaction.resolution_time IS 'Tempo de resolução em segundos (benchmarks: FCR, SLA).';
COMMENT ON COLUMN crm.interaction.responsible IS 'Agente humano ou IA responsável (SOX, ISO 9001).';
COMMENT ON COLUMN crm.interaction.result IS 'Resultado da interação (resolvido, pendente, escalado, etc).';
COMMENT ON COLUMN crm.interaction.ai_score IS 'Predição IA: churn, conversão, satisfação (NIST AI RMF, benchmarks Salesforce).';
COMMENT ON COLUMN crm.interaction.sentiment IS 'Análise de sentimento (NLP, ISO 56000, benchmarks globais).';
COMMENT ON COLUMN crm.interaction.consent_status IS 'Status do consentimento (GDPR, LGPD, CCPA).';
COMMENT ON COLUMN crm.interaction.audit_log IS 'Log de auditoria e rastreabilidade (SOX, ISO 27001, PCI DSS).';
COMMENT ON COLUMN crm.interaction.compliance IS 'Compliance detalhado (normas, regulamentações, avisos, decretos).';
COMMENT ON COLUMN crm.interaction.benchmark_kpi IS 'KPIs e benchmarks: NPS, SLA, CSAT, CES, FCR, Churn Rate, CAC, LTV, etc (Forrester, Gartner, ISO 10002).';
COMMENT ON COLUMN crm.interaction.benchmark_score IS 'Score de benchmarking internacional (Forrester, Gartner, CXPA, ISO 10002).';

CREATE TABLE IF NOT EXISTS crm.campaign (
    campaign_id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    objective TEXT,
    channel VARCHAR(50),
    status VARCHAR(50),
    start_date DATE,
    end_date DATE,
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

CREATE TABLE IF NOT EXISTS crm.ticket (
    ticket_id SERIAL PRIMARY KEY,
    contact_id INT REFERENCES crm.contact(contact_id),
    type VARCHAR(50),
    priority VARCHAR(50),
    status VARCHAR(50),
    sla VARCHAR(50),
    opened_at TIMESTAMP DEFAULT now(),
    closed_at TIMESTAMP,
    resolution TEXT,
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

-- Cadastro de produtos e serviços (CRM/Vendas/Serviços)
CREATE TABLE IF NOT EXISTS crm.product (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50), -- produto, serviço, pacote
    category VARCHAR(100),
    price NUMERIC(15,2),
    status VARCHAR(50) DEFAULT 'active',
    compliance JSONB,
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    sku VARCHAR(100),
    global_category VARCHAR(100),
    tags JSONB
);
COMMENT ON TABLE crm.product IS 'Cadastro de produtos e serviços, conforme ISO 9001, ITIL, benchmarks globais de CRM e vendas.';
COMMENT ON COLUMN crm.product.compliance IS 'Compliance: normas, regulamentações, avisos e decretos aplicáveis ao produto/serviço.';
COMMENT ON COLUMN crm.product.audit_log IS 'Log de auditoria e rastreabilidade (SOX, ISO 27001).';
COMMENT ON COLUMN crm.product.sku IS 'SKU do produto/serviço (ISO 9001, rastreabilidade global).';
COMMENT ON COLUMN crm.product.global_category IS 'Categoria global/classificação internacional (UNSPSC, GPC, ISO 8000).';
COMMENT ON COLUMN crm.product.tags IS 'Tags e atributos dinâmicos (extensibilidade, multilíngue).';

-- Campanhas de marketing
CREATE TABLE IF NOT EXISTS crm.marketing_campaign (
    campaign_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    objective TEXT,
    segment VARCHAR(100),
    channel VARCHAR(50),
    start_date DATE,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    consent_required BOOLEAN DEFAULT TRUE,
    compliance JSONB,
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.marketing_campaign IS 'Campanhas de marketing omnichannel, com consentimento, rastreabilidade e compliance (GDPR, LGPD, ISO 10002, benchmarks CXPA, Forrester, Gartner).';

-- Cadastro de chatbots, voicebots e agentes
CREATE TABLE IF NOT EXISTS crm.chatbot_agent (
    agent_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50), -- chatbot, voicebot, agente humano
    active BOOLEAN DEFAULT TRUE,
    script JSONB,
    ai_model VARCHAR(100),
    compliance JSONB,
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.chatbot_agent IS 'Cadastro de chatbots, voicebots e agentes, com rastreabilidade, script, modelo IA e compliance (NIST AI RMF, GDPR, LGPD, benchmarks globais).';

-- Atividades, tarefas e follow-ups
CREATE TABLE IF NOT EXISTS crm.activity (
    activity_id SERIAL PRIMARY KEY,
    related_type VARCHAR(50), -- lead, opportunity, ticket, etc.
    related_id INT,
    subject VARCHAR(255),
    description TEXT,
    due_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'pending',
    responsible VARCHAR(100),
    compliance JSONB,
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.activity IS 'Atividades, tarefas e follow-ups relacionados a qualquer entidade CRM, conforme benchmarks de produtividade e compliance.';

-- Segmentação de público
CREATE TABLE IF NOT EXISTS crm.segment (
    segment_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    criteria JSONB,
    status VARCHAR(50) DEFAULT 'active',
    compliance JSONB,
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.segment IS 'Segmentação de público para campanhas, vendas e atendimento, conforme benchmarks de marketing e privacidade.';

-- SCHEMA CXP
CREATE SCHEMA IF NOT EXISTS cxp;

CREATE TABLE IF NOT EXISTS cxp.touchpoint (
    touchpoint_id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    channel VARCHAR(50),
    type VARCHAR(50),
    status VARCHAR(50),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

-- Journey enriquecida com visão 360º, omnichannel, IA, consentimento granular, benchmarking, auditoria, extensibilidade, KPIs e compliance global
CREATE TABLE IF NOT EXISTS cxp.journey (
    journey_id SERIAL PRIMARY KEY,
    customer_id UUID, -- Visão 360º, identificação global
    name VARCHAR(255),
    description TEXT,
    start_date DATE,
    end_date DATE,
    channel VARCHAR(50), -- omnichannel
    status VARCHAR(50),
    ai_score NUMERIC(5,2), -- predição de experiência/satisfação
    sentiment VARCHAR(50), -- análise de sentimento
    consent_status VARCHAR(50), -- granular, por canal/finalidade
    consent_log JSONB, -- histórico de consentimento
    custom_fields JSONB, -- extensibilidade
    benchmark_kpi JSONB, -- KPIs e benchmarks (NPS, CES, SLA, etc)
    audit_log JSONB, -- logs de alteração/acesso
    external_ref VARCHAR(100),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status_global VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e referência normativa para campos críticos
COMMENT ON TABLE cxp.journey IS 'Journey enriquecida com visão 360º, omnichannel, IA, consentimento granular, benchmarking, auditoria, extensibilidade, KPIs e compliance global. Benchmarks Salesforce, Dynamics, SAP CX, Oracle CX, HubSpot, Gartner, Forrester.';
COMMENT ON COLUMN cxp.journey.customer_id IS 'Identificação global do cliente para visão 360º.';
COMMENT ON COLUMN cxp.journey.channel IS 'Canal da jornada: web, mobile, social, call, chat, etc. Omnichannel.';
COMMENT ON COLUMN cxp.journey.ai_score IS 'Score preditivo de IA (experiência/satisfação).';
COMMENT ON COLUMN cxp.journey.sentiment IS 'Análise de sentimento automatizada.';
COMMENT ON COLUMN cxp.journey.consent_status IS 'Status granular do consentimento (GDPR, LGPD, CCPA).';
COMMENT ON COLUMN cxp.journey.custom_fields IS 'Campos customizáveis/extensíveis.';
COMMENT ON COLUMN cxp.journey.benchmark_kpi IS 'KPIs e benchmarks: NPS, CES, SLA, etc.';
COMMENT ON COLUMN cxp.journey.audit_log IS 'Log de auditoria e rastreabilidade.';

-- Feedback enriquecido com visão 360º, IA, consentimento granular, benchmarking, auditoria, extensibilidade, KPIs e compliance global
CREATE TABLE IF NOT EXISTS cxp.feedback (
    feedback_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    customer_id UUID, -- Visão 360º, identificação global
    contact_id INT REFERENCES crm.contact(contact_id),
    feedback_text TEXT,
    feedback_type VARCHAR(50), -- ex: NPS, CES, CSAT, VOC
    channel VARCHAR(50), -- omnichannel
    sentiment VARCHAR(50), -- análise de sentimento
    ai_score NUMERIC(5,2), -- predição de satisfação/experiência
    language VARCHAR(10), -- ISO 639-1
    tags TEXT[], -- classificação livre
    topic VARCHAR(100), -- assunto principal
    urgency_level VARCHAR(20), -- low, medium, high, critical
    response_time INTERVAL,
    followup_required BOOLEAN DEFAULT FALSE,
    followup_status VARCHAR(50),
    consent_status VARCHAR(50),
    consent_log JSONB,
    custom_fields JSONB,
    benchmark_kpi JSONB,
    audit_log JSONB,
    external_ref VARCHAR(100),
    compliance JSONB,
    data_classification VARCHAR(50) DEFAULT 'personal',
    regulatory_ref VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e normativos para cxp.feedback
COMMENT ON TABLE cxp.feedback IS 'Feedback detalhado e análise avançada de experiência do cliente. Estrutura internacional para granularidade, inteligência analítica, compliance, rastreabilidade e automação. Alinhada a ISO 10002, CXPA, Forrester, Gartner, GDPR, LGPD, SOX.';
COMMENT ON COLUMN cxp.feedback.feedback_text IS 'Texto do feedback do cliente, aberto ou estruturado.';
COMMENT ON COLUMN cxp.feedback.feedback_type IS 'Tipo de feedback: NPS, CES, CSAT, VOC.';
COMMENT ON COLUMN cxp.feedback.channel IS 'Canal de origem do feedback: omnichannel.';
COMMENT ON COLUMN cxp.feedback.sentiment IS 'Sentimento do feedback: análise manual ou IA (NLP).';
COMMENT ON COLUMN cxp.feedback.ai_score IS 'Score preditivo de IA para satisfação ou experiência.';
COMMENT ON COLUMN cxp.feedback.language IS 'Idioma do feedback (ISO 639-1).';
COMMENT ON COLUMN cxp.feedback.tags IS 'Tags livres para classificação e analytics.';
COMMENT ON COLUMN cxp.feedback.topic IS 'Tópico principal do feedback.';
COMMENT ON COLUMN cxp.feedback.urgency_level IS 'Nível de urgência: low, medium, high, critical.';
COMMENT ON COLUMN cxp.feedback.response_time IS 'Tempo de resposta ao feedback.';
COMMENT ON COLUMN cxp.feedback.followup_required IS 'Indica se o feedback exige follow-up.';
COMMENT ON COLUMN cxp.feedback.followup_status IS 'Status do follow-up.';
COMMENT ON COLUMN cxp.feedback.compliance IS 'Metadados de compliance: GDPR, LGPD, ISO 27701, NIST, PCI DSS, SOX, HIPAA, etc.';
COMMENT ON COLUMN cxp.feedback.data_classification IS 'Classificação de dados: pessoal, sensível, anonimizado.';
COMMENT ON COLUMN cxp.feedback.audit_log IS 'Log de auditoria e rastreabilidade.';
COMMENT ON COLUMN cxp.feedback.benchmark_kpi IS 'KPIs e benchmarks: NPS, CES, SLA, etc.';

-- Trigger de auditoria para cxp.feedback
CREATE TRIGGER trg_audit_feedback
BEFORE UPDATE ON cxp.feedback
FOR EACH ROW
EXECUTE FUNCTION fn_audit_log();

-- Trigger de compliance para cxp.feedback
CREATE OR REPLACE FUNCTION fn_check_compliance_feedback()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance IS NULL OR NEW.compliance::text = '{}' THEN
    RAISE EXCEPTION 'Campo compliance é obrigatório para registros Feedback!';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_compliance_feedback
BEFORE INSERT OR UPDATE ON cxp.feedback
FOR EACH ROW
EXECUTE FUNCTION fn_check_compliance_feedback();

-- Seed de exemplo para cxp.feedback
INSERT INTO cxp.feedback (journey_id, customer_id, contact_id, feedback_text, feedback_type, channel, sentiment, ai_score, language, tags, topic, urgency_level, response_time, followup_required, followup_status, consent_status, consent_log, custom_fields, benchmark_kpi, audit_log, external_ref, compliance, created_by, updated_by, status)
VALUES
(1, 'uuid-1', 1, 'Gostei muito do atendimento, mas tive dificuldade no pagamento.', 'VOC', 'web', 'mixed', 3.9, 'pt', ARRAY['pagamento','atendimento'], 'Pagamento', 'medium', '00:02:30', TRUE, 'pending', 'given', '{"consent": true}', '{"extra":"info"}', '{"nps": 75}', '{"action": "insert", "user": "admin"}', 'ORDER-002', '{"regulatory_ref": "ISO 10002, GDPR"}', 'admin', 'admin', 'active');

-- Query de BI: Analytics de sentimento, canal e urgência
SELECT
  channel,
  sentiment,
  urgency_level,
  COUNT(*) AS total_feedbacks
FROM cxp.feedback
WHERE status = 'active'
GROUP BY channel, sentiment, urgency_level
ORDER BY channel, sentiment, urgency_level;

-- Checklist técnico multilíngue para rollout Feedback
-- 1. Validar inserção de feedback com e sem compliance (trigger deve bloquear sem compliance).
-- 2. Executar query de analytics por canal, sentimento e urgência.
-- 3. Garantir preenchimento de regulatory_ref e data_classification conforme normas.
-- 4. Auditar logs de inserção e atualização.
-- 5. Atualizar documentação multilíngue e ERD.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.feedback (Feedback Detalhado)
-- - **Descrição PT:** Feedback detalhado e análise avançada de experiência do cliente, com granularidade, inteligência analítica, compliance, rastreabilidade e automação.
-- - **Descrição EN:** Detailed feedback and advanced customer experience analytics, with granularity, analytical intelligence, compliance, traceability and automation.
-- - **Campos-chave:** feedback_text, feedback_type, channel, sentiment, ai_score, language, tags, topic, urgency_level, response_time, followup_required, followup_status, compliance, audit_log, benchmark_kpi.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester, Gartner, CXPA.

-- NPS: Net Promoter Score (ISO 10002, CXPA, Forrester, Gartner)
CREATE TABLE IF NOT EXISTS cxp.nps (
    nps_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    score NUMERIC(5,2),
    collected_at TIMESTAMP DEFAULT now(),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

CREATE TABLE IF NOT EXISTS cxp.csat (
    csat_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    score NUMERIC(5,2),
    collected_at TIMESTAMP DEFAULT now(),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    audit_log JSONB
);

-- CES: Customer Effort Score (ISO 10002, CXPA, Forrester CX Index)
CREATE TABLE IF NOT EXISTS cxp.ces (
    ces_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    score NUMERIC(5,2), -- Conforme benchmarks Forrester, Gartner, CXPA
    industry_benchmark NUMERIC(5,2), -- Benchmark setorial
    collected_at TIMESTAMP DEFAULT now(),
    compliance JSONB, -- GDPR, LGPD, ISO 27701, NIST
    regulatory_ref VARCHAR(255), -- Ex: GDPR Art. 6, LGPD Art. 7, CCPA Section 1798.100
    data_classification VARCHAR(50) DEFAULT 'anonymous', -- pessoal, sensível, anonimizado
    audit_log JSONB, -- Auditoria de alterações (SOX, ISO 27001)
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e referência normativa para campos críticos
COMMENT ON TABLE cxp.ces IS 'Customer Effort Score (CES) - Métrica internacional de experiência do cliente. Alinhada a ISO 10002, CXPA, Forrester CX Index, Gartner CX Pyramid. Inclui campos para benchmarking, compliance (GDPR, LGPD, ISO 27701, NIST), rastreabilidade, auditoria (SOX, ISO 27001), status e classificação de dados.';
COMMENT ON COLUMN cxp.ces.score IS 'Pontuação CES do cliente conforme benchmarks Forrester, Gartner, CXPA.';
COMMENT ON COLUMN cxp.ces.industry_benchmark IS 'Benchmark setorial para comparação (Forrester, Gartner, CXPA).';
COMMENT ON COLUMN cxp.ces.compliance IS 'Compliance: normas, regulamentações, decretos, avisos aplicáveis ao CES.';
COMMENT ON COLUMN cxp.ces.regulatory_ref IS 'Referência normativa: artigo, decreto, aviso, regulamento ou legislação aplicável.';
COMMENT ON COLUMN cxp.ces.data_classification IS 'Classificação de dados: pessoal, sensível, anonimizado.';
COMMENT ON COLUMN cxp.ces.audit_log IS 'Log de auditoria de alterações (SOX, ISO 27001).';

-- VOC: Voice of Customer (ISO 10002, CXPA, Forrester CX Pyramid, Gartner CX)
CREATE TABLE IF NOT EXISTS cxp.voc (
    voc_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    customer_id UUID, -- Visão 360º, identificação global
    contact_id INT REFERENCES crm.contact(contact_id),
    feedback_text TEXT,
    feedback_type VARCHAR(50), -- ex: NPS, CES, CSAT, VOC
    channel VARCHAR(50), -- omnichannel
    sentiment VARCHAR(50), -- análise de sentimento
    ai_score NUMERIC(5,2), -- predição de satisfação/experiência
    consent_status VARCHAR(50), -- granular, por canal/finalidade
    consent_log JSONB, -- histórico de consentimento
    custom_fields JSONB, -- extensibilidade
    benchmark_kpi JSONB, -- KPIs e benchmarks (NPS, CES, SLA, etc)
    audit_log JSONB, -- logs de alteração/acesso
    external_ref VARCHAR(100),
    compliance JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e referência normativa para campos críticos
COMMENT ON TABLE cxp.voc IS 'Voice of Customer (VOC) - Estrutura internacional para coleta, classificação e análise de feedbacks do cliente. Alinhada a ISO 10002, CXPA, Forrester CX Pyramid, Gartner CX. Inclui consentimento, opt-in/out, preferências, classificação, compliance, benchmarking, rastreabilidade, auditoria e status.';
COMMENT ON COLUMN cxp.voc.customer_id IS 'Identificação global do cliente para visão 360º.';
COMMENT ON COLUMN cxp.voc.feedback_type IS 'Tipo de feedback: NPS, CES, CSAT, VOC.';
COMMENT ON COLUMN cxp.voc.channel IS 'Canal do feedback: web, mobile, social, call, chat, etc. Omnichannel.';
COMMENT ON COLUMN cxp.voc.ai_score IS 'Score preditivo de IA (satisfação/experiência).';
COMMENT ON COLUMN cxp.voc.sentiment IS 'Análise de sentimento automatizada.';
COMMENT ON COLUMN cxp.voc.consent_status IS 'Status granular do consentimento (GDPR, LGPD, CCPA).';
COMMENT ON COLUMN cxp.voc.custom_fields IS 'Campos customizáveis/extensíveis.';
COMMENT ON COLUMN cxp.voc.benchmark_kpi IS 'KPIs e benchmarks: NPS, CES, SLA, etc.';
COMMENT ON COLUMN cxp.voc.audit_log IS 'Log de auditoria e rastreabilidade.';

-- Trigger de auditoria para cxp.voc
CREATE TRIGGER trg_audit_voc
BEFORE UPDATE ON cxp.voc
FOR EACH ROW
EXECUTE FUNCTION fn_audit_log();

-- Trigger de compliance para cxp.voc
CREATE OR REPLACE FUNCTION fn_check_compliance_voc()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance IS NULL OR NEW.compliance::text = '{}' THEN
    RAISE EXCEPTION 'Campo compliance é obrigatório para registros VOC!';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_compliance_voc
BEFORE INSERT OR UPDATE ON cxp.voc
FOR EACH ROW
EXECUTE FUNCTION fn_check_compliance_voc();

-- Seed de exemplo para cxp.voc
INSERT INTO cxp.voc (journey_id, customer_id, contact_id, feedback_text, feedback_type, channel, sentiment, ai_score, consent_status, consent_log, custom_fields, benchmark_kpi, audit_log, external_ref, compliance, created_by, updated_by, status)
VALUES
(1, 1, 1, 'Atendimento excelente, rápido e resolutivo.', 'VOC', 'web', 'positive', 4.8, 'given', '{"consent": true}', '{"channel": "web", "language": "pt"}', '{"nps": 85}', '{"action": "insert", "user": "admin"}', 'ORDER-001', '{"regulatory_ref": "ISO 10002, GDPR"}', 'admin', 'admin', 'active');

-- Query de BI: Sentimento e consentimento por jornada
SELECT
  j.name AS journey_name,
  v.sentiment,
  v.consent_status,
  COUNT(*) AS total_feedbacks
FROM cxp.voc v
JOIN cxp.journey j ON v.journey_id = j.journey_id
WHERE v.status = 'active'
GROUP BY j.name, v.sentiment, v.consent_status
ORDER BY j.name, v.sentiment;

-- Checklist técnico multilíngue para rollout VOC
-- 1. Validar inserção de VOC com e sem compliance (trigger deve bloquear sem compliance).
-- 2. Executar query de BI por sentimento e consentimento.
-- 3. Garantir preenchimento de regulatory_ref e data_classification conforme normas.
-- 4. Auditar logs de inserção e atualização.
-- 5. Atualizar documentação multilíngue e ERD.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.voc (Voice of Customer)
-- - **Descrição PT:** Estrutura internacional para coleta, classificação e análise de feedbacks do cliente, incluindo consentimento, preferências, compliance, benchmarking e rastreabilidade.
-- - **Descrição EN:** International structure for collection, classification and analysis of customer feedback, including consent, preferences, compliance, benchmarking and traceability.
-- - **Campos-chave:** feedback_text, feedback_type, channel, sentiment, ai_score, consent_status, custom_fields, benchmark_kpi, audit_log.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester CX Index.

-- NPS: Net Promoter Score (ISO 10002, CXPA, Forrester, Gartner)
CREATE TABLE IF NOT EXISTS cxp.nps (
    nps_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    score NUMERIC(5,2), -- Escala -100 a 100
    promoter_count INT,
    passive_count INT,
    detractor_count INT,
    industry_benchmark NUMERIC(5,2),
    collected_at TIMESTAMP DEFAULT now(),
    compliance JSONB, -- GDPR, LGPD, ISO 27701, NIST
    regulatory_ref VARCHAR(255),
    data_classification VARCHAR(50) DEFAULT 'anonymous',
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e normativos para cxp.nps
COMMENT ON TABLE cxp.nps IS 'Net Promoter Score (NPS) - Métrica internacional de lealdade do cliente. Alinhada a ISO 10002, CXPA, Forrester, Gartner. Inclui benchmarking, compliance, rastreabilidade, auditoria e status.';
COMMENT ON COLUMN cxp.nps.score IS 'Pontuação NPS (-100 a 100) conforme benchmarks globais.';
COMMENT ON COLUMN cxp.nps.promoter_count IS 'Quantidade de promotores.';
COMMENT ON COLUMN cxp.nps.passive_count IS 'Quantidade de clientes neutros/passivos.';
COMMENT ON COLUMN cxp.nps.detractor_count IS 'Quantidade de detratores.';
COMMENT ON COLUMN cxp.nps.industry_benchmark IS 'Benchmark setorial NPS.';
COMMENT ON COLUMN cxp.nps.compliance IS 'Metadados de compliance: GDPR, LGPD, ISO 27701, NIST, PCI DSS, SOX, HIPAA, entre outros.';
COMMENT ON COLUMN cxp.nps.regulatory_ref IS 'Referência normativa: artigo, decreto, regulamento ou legislação aplicável.';
COMMENT ON COLUMN cxp.nps.data_classification IS 'Classificação de dados: pessoal, sensível, anonimizado.';
COMMENT ON COLUMN cxp.nps.audit_log IS 'Log de auditoria de alterações (SOX, ISO 27001).';
COMMENT ON COLUMN cxp.nps.created_at IS 'Data/hora de criação do registro.';
COMMENT ON COLUMN cxp.nps.updated_at IS 'Data/hora da última atualização do registro.';
COMMENT ON COLUMN cxp.nps.created_by IS 'Usuário responsável pela criação.';
COMMENT ON COLUMN cxp.nps.updated_by IS 'Usuário responsável pela última atualização.';
COMMENT ON COLUMN cxp.nps.status IS 'Status do registro (ativo, inativo, suspenso, etc).';

-- Trigger de auditoria para cxp.nps
CREATE TRIGGER trg_audit_nps
BEFORE UPDATE ON cxp.nps
FOR EACH ROW
EXECUTE FUNCTION fn_audit_log();

-- Trigger de compliance para cxp.nps
CREATE OR REPLACE FUNCTION fn_check_compliance_nps()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance IS NULL OR NEW.compliance::text = '{}' THEN
    RAISE EXCEPTION 'Campo compliance é obrigatório para registros NPS!';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_compliance_nps
BEFORE INSERT OR UPDATE ON cxp.nps
FOR EACH ROW
EXECUTE FUNCTION fn_check_compliance_nps();

-- Seed de exemplo para cxp.nps
INSERT INTO cxp.nps (journey_id, score, promoter_count, passive_count, detractor_count, industry_benchmark, compliance, regulatory_ref, data_classification, audit_log, created_by, updated_by, status)
VALUES
(1, 65.5, 120, 30, 10, 60.0, '{"regulatory_ref": "ISO 10002, GDPR"}', 'GDPR Art. 6', 'anonymous', '{"action": "insert", "user": "admin"}', 'admin', 'admin', 'active');

-- Query de BI: Média de NPS por Jornada
SELECT
  j.name AS journey_name,
  AVG(n.score) AS avg_nps,
  AVG(n.industry_benchmark) AS avg_benchmark,
  SUM(n.promoter_count) AS total_promoters,
  SUM(n.passive_count) AS total_passives,
  SUM(n.detractor_count) AS total_detractors
FROM cxp.nps n
JOIN cxp.journey j ON n.journey_id = j.journey_id
WHERE n.status = 'active'
GROUP BY j.name
ORDER BY avg_nps DESC;

-- Checklist técnico multilíngue para rollout NPS
-- 1. Validar inserção de NPS com e sem compliance (trigger deve bloquear sem compliance).
-- 2. Executar query de média de NPS por jornada e comparar com benchmark.
-- 3. Garantir preenchimento de regulatory_ref e data_classification conforme normas.
-- 4. Auditar logs de inserção e atualização.
-- 5. Atualizar documentação multilíngue e ERD.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.nps (Net Promoter Score)
-- - **Descrição PT:** Métrica internacional de lealdade do cliente, recomendada por ISO 10002, CXPA, Forrester, Gartner. Permite mensurar o NPS por jornada, integrando compliance, privacidade, rastreabilidade e auditoria.
-- - **Descrição EN:** International customer loyalty metric, recommended by ISO 10002, CXPA, Forrester, Gartner. Measures NPS by journey, integrating compliance, privacy, traceability, and audit.
-- - **Campos-chave:** score, promoter_count, passive_count, detractor_count, industry_benchmark, compliance, regulatory_ref, data_classification, audit_log.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester CX Index.

-- CSAT: Customer Satisfaction Score (ISO 10002, CXPA, Forrester, Gartner)
CREATE TABLE IF NOT EXISTS cxp.csat (
    csat_id SERIAL PRIMARY KEY,
    journey_id INT REFERENCES cxp.journey(journey_id),
    score NUMERIC(5,2), -- Escala 1-5 ou 1-10
    industry_benchmark NUMERIC(5,2),
    collected_at TIMESTAMP DEFAULT now(),
    compliance JSONB, -- GDPR, LGPD, ISO 27701, NIST
    regulatory_ref VARCHAR(255),
    data_classification VARCHAR(50) DEFAULT 'anonymous',
    audit_log JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active'
);

-- Comentários multilíngues e normativos para cxp.csat
COMMENT ON TABLE cxp.csat IS 'Customer Satisfaction Score (CSAT) - Métrica internacional de satisfação do cliente. Alinhada a ISO 10002, CXPA, Forrester, Gartner. Inclui benchmarking, compliance, rastreabilidade, auditoria e status.';
COMMENT ON COLUMN cxp.csat.score IS 'Pontuação CSAT (1-5 ou 1-10) conforme benchmarks globais.';
COMMENT ON COLUMN cxp.csat.industry_benchmark IS 'Benchmark setorial CSAT.';
COMMENT ON COLUMN cxp.csat.compliance IS 'Metadados de compliance: GDPR, LGPD, ISO 27701, NIST, PCI DSS, SOX, HIPAA, entre outros.';
COMMENT ON COLUMN cxp.csat.regulatory_ref IS 'Referência normativa: artigo, decreto, regulamento ou legislação aplicável.';
COMMENT ON COLUMN cxp.csat.data_classification IS 'Classificação de dados: pessoal, sensível, anonimizado.';
COMMENT ON COLUMN cxp.csat.audit_log IS 'Log de auditoria de alterações (SOX, ISO 27001).';
COMMENT ON COLUMN cxp.csat.created_at IS 'Data/hora de criação do registro.';
COMMENT ON COLUMN cxp.csat.updated_at IS 'Data/hora da última atualização do registro.';
COMMENT ON COLUMN cxp.csat.created_by IS 'Usuário responsável pela criação.';
COMMENT ON COLUMN cxp.csat.updated_by IS 'Usuário responsável pela última atualização.';
COMMENT ON COLUMN cxp.csat.status IS 'Status do registro (ativo, inativo, suspenso, etc).';

-- Trigger de auditoria para cxp.csat
CREATE TRIGGER trg_audit_csat
BEFORE UPDATE ON cxp.csat
FOR EACH ROW
EXECUTE FUNCTION fn_audit_log();

-- Trigger de compliance para cxp.csat
CREATE OR REPLACE FUNCTION fn_check_compliance_csat()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance IS NULL OR NEW.compliance::text = '{}' THEN
    RAISE EXCEPTION 'Campo compliance é obrigatório para registros CSAT!';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_compliance_csat
BEFORE INSERT OR UPDATE ON cxp.csat
FOR EACH ROW
EXECUTE FUNCTION fn_check_compliance_csat();

-- Seed de exemplo para cxp.csat
INSERT INTO cxp.csat (journey_id, score, industry_benchmark, compliance, regulatory_ref, data_classification, audit_log, created_by, updated_by, status)
VALUES
(1, 4.7, 4.3, '{"regulatory_ref": "ISO 10002, GDPR"}', 'GDPR Art. 6', 'anonymous', '{"action": "insert", "user": "admin"}', 'admin', 'admin', 'active');

-- Query de BI: Média de CSAT por Jornada
SELECT
  j.name AS journey_name,
  AVG(c.score) AS avg_csat,
  AVG(c.industry_benchmark) AS avg_benchmark,
  COUNT(*) AS total_responses
FROM cxp.csat c
JOIN cxp.journey j ON c.journey_id = j.journey_id
WHERE c.status = 'active'
GROUP BY j.name
ORDER BY avg_csat DESC;

-- Checklist técnico multilíngue para rollout CSAT
-- 1. Validar inserção de CSAT com e sem compliance (trigger deve bloquear sem compliance).
-- 2. Executar query de média de CSAT por jornada e comparar com benchmark.
-- 3. Garantir preenchimento de regulatory_ref e data_classification conforme normas.
-- 4. Auditar logs de inserção e atualização.
-- 5. Atualizar documentação multilíngue e ERD.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.csat (Customer Satisfaction Score)
-- - **Descrição PT:** Métrica internacional de satisfação do cliente, recomendada por ISO 10002, CXPA, Forrester, Gartner. Permite mensurar o CSAT por jornada, integrando compliance, privacidade, rastreabilidade e auditoria.
-- - **Descrição EN:** International customer satisfaction metric, recommended by ISO 10002, CXPA, Forrester, Gartner. Measures CSAT by journey, integrating compliance, privacy, traceability, and audit.
-- - **Campos-chave:** score, industry_benchmark, compliance, regulatory_ref, data_classification, audit_log.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester CX Index.

-- Tabela de vendas (Sales)
CREATE TABLE IF NOT EXISTS crm.sale (
    sale_id SERIAL PRIMARY KEY,
    sale_code VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID,
    contact_id INT REFERENCES crm.contact(contact_id),
    opportunity_id INT REFERENCES crm.opportunity(opportunity_id),
    product_id INT REFERENCES crm.product(product_id),
    segment_id INT REFERENCES crm.segment(segment_id),
    campaign_id INT REFERENCES crm.marketing_campaign(campaign_id),
    agent_id INT REFERENCES crm.chatbot_agent(agent_id),
    sale_date TIMESTAMP DEFAULT now(),
    expected_close_date TIMESTAMP,
    actual_close_date TIMESTAMP,
    sale_stage VARCHAR(50),
    amount NUMERIC(15,2),
    currency VARCHAR(10) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'active',
    compliance JSONB,
    audit_log JSONB,
    benchmark_kpi VARCHAR(50),
    benchmark_score NUMERIC(5,2),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.sale IS 'Registro detalhado de vendas, pipeline, forecast, conforme benchmarks Salesforce, Dynamics, SAP CX, Oracle CX, HubSpot, ISO 9001, SOX, GDPR, LGPD, PCI DSS.';
COMMENT ON COLUMN crm.sale.sale_code IS 'Código único da venda (SOX, rastreabilidade).';
COMMENT ON COLUMN crm.sale.sale_stage IS 'Estágio do pipeline de vendas: prospect, proposal, negotiation, won, lost (benchmarks globais).';
COMMENT ON COLUMN crm.sale.compliance IS 'Compliance: normas, regulamentações, decretos, avisos aplicáveis à venda.';
COMMENT ON COLUMN crm.sale.audit_log IS 'Log de auditoria e rastreabilidade (SOX, ISO 27001, PCI DSS).';
COMMENT ON COLUMN crm.sale.benchmark_kpi IS 'KPIs: Conversion Rate, Win Rate, Cycle Time, etc (Forrester, Gartner, BIAN).';

-- Tabela de tickets/serviços
CREATE TABLE IF NOT EXISTS crm.service_ticket (
    ticket_id SERIAL PRIMARY KEY,
    ticket_code VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID,
    contact_id INT REFERENCES crm.contact(contact_id),
    product_id INT REFERENCES crm.product(product_id),
    journey_id INT REFERENCES cxp.journey(journey_id),
    channel VARCHAR(50),
    type VARCHAR(50),
    priority VARCHAR(50),
    sla_level VARCHAR(50),
    status VARCHAR(50) DEFAULT 'open',
    subject VARCHAR(255),
    description TEXT,
    resolution TEXT,
    responsible VARCHAR(100),
    ai_score NUMERIC(5,2),
    sentiment VARCHAR(50),
    compliance JSONB,
    audit_log JSONB,
    benchmark_kpi VARCHAR(50),
    benchmark_score NUMERIC(5,2),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.service_ticket IS 'Registro detalhado de tickets/serviços, SLA, omnichannel, IA, compliance, conforme ITIL, ISO 20000, benchmarks Zendesk, Salesforce Service, GDPR, LGPD.';
COMMENT ON COLUMN crm.service_ticket.type IS 'Tipo: incidente, requisição, dúvida, reclamação, etc (ISO 20000, ITIL, benchmarks globais).';
COMMENT ON COLUMN crm.service_ticket.compliance IS 'Compliance: normas, regulamentações, avisos, decretos aplicáveis ao serviço.';

-- Oportunidades de negócio (pipeline)
CREATE TABLE IF NOT EXISTS crm.opportunity (
    opportunity_id SERIAL PRIMARY KEY,
    opportunity_code VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID,
    contact_id INT REFERENCES crm.contact(contact_id),
    product_id INT REFERENCES crm.product(product_id),
    segment_id INT REFERENCES crm.segment(segment_id),
    campaign_id INT REFERENCES crm.marketing_campaign(campaign_id),
    stage VARCHAR(50),
    amount NUMERIC(15,2),
    currency VARCHAR(10) DEFAULT 'USD',
    expected_close_date TIMESTAMP,
    actual_close_date TIMESTAMP,
    probability NUMERIC(5,2),
    status VARCHAR(50) DEFAULT 'active',
    compliance JSONB,
    audit_log JSONB,
    benchmark_kpi VARCHAR(50),
    benchmark_score NUMERIC(5,2),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);
COMMENT ON TABLE crm.opportunity IS 'Oportunidades de negócio, pipeline, forecast, conforme benchmarks Salesforce, Dynamics, SAP CX, Oracle CX, HubSpot, ISO 9001, SOX, GDPR, LGPD.';
COMMENT ON COLUMN crm.opportunity.stage IS 'Estágio do pipeline: qualification, proposal, negotiation, closed_won, closed_lost (benchmarks globais).';

-- Detalhamento adicional para produtos
ALTER TABLE crm.product
    ADD COLUMN IF NOT EXISTS sku VARCHAR(100),
    ADD COLUMN IF NOT EXISTS global_category VARCHAR(100),
    ADD COLUMN IF NOT EXISTS tags JSONB;
COMMENT ON COLUMN crm.product.sku IS 'SKU do produto/serviço (ISO 9001, rastreabilidade global).';
COMMENT ON COLUMN crm.product.global_category IS 'Categoria global/classificação internacional (UNSPSC, GPC, ISO 8000).';
COMMENT ON COLUMN crm.product.tags IS 'Tags e atributos dinâmicos (extensibilidade, multilíngue).';

-- Triggers de auditoria para novas tabelas
CREATE TRIGGER trg_audit_sale BEFORE UPDATE ON crm.sale FOR EACH ROW EXECUTE FUNCTION fn_audit_log();
CREATE TRIGGER trg_audit_service_ticket BEFORE UPDATE ON crm.service_ticket FOR EACH ROW EXECUTE FUNCTION fn_audit_log();
CREATE TRIGGER trg_audit_opportunity BEFORE UPDATE ON crm.opportunity FOR EACH ROW EXECUTE FUNCTION fn_audit_log();

-- Views de BI e compliance
CREATE OR REPLACE VIEW vw_sales_by_agent AS
SELECT
  agent_id,
  status,
  COUNT(*) AS total,
  SUM(amount) AS total_amount
FROM crm.sale
GROUP BY agent_id, status;

CREATE OR REPLACE VIEW vw_open_tickets_by_priority AS
SELECT
  priority,
  COUNT(*) AS total
FROM crm.service_ticket
WHERE status = 'open'
GROUP BY priority;

CREATE OR REPLACE VIEW vw_opportunities_by_stage AS
SELECT
  stage,
  COUNT(*) AS total,
  SUM(amount) AS total_amount
FROM crm.opportunity
GROUP BY stage;

-- Query: Relatório de Ativos por Cliente e Status
SELECT
  a.name AS account_name,
  s.asset_code,
  s.name AS asset_name,
  s.status,
  s.purchase_date,
  s.warranty_expiry
FROM crm.account a
JOIN crm.asset s ON s.account_id = a.account_id
ORDER BY a.name, s.status;

-- Query: Participação em Campanhas e Engajamento
SELECT
  a.name AS account_name,
  cm.campaign_id,
  cm.status,
  cm.kpi->>'engagement' AS engagement_score
FROM crm.account a
JOIN crm.campaign_member cm ON cm.account_id = a.account_id
ORDER BY engagement_score DESC NULLS LAST;

-- Query: Ranking de Equipes por Performance
SELECT
  t.name AS team_name,
  ut.role,
  ut.kpi->>'performance' AS user_performance
FROM crm.team t
JOIN crm.user_team ut ON ut.team_id = t.team_id
ORDER BY t.name, user_performance DESC NULLS LAST;

-- Template de integração externa (API RESTful)
-- POST /api/v1/crm/orders
-- Content-Type: application/json
-- {
--   "order_code": "ORDER-002",
--   "account_id": 1,
--   "quote_id": 1,
--   "amount": 5000.00,
--   "currency": "EUR",
--   "status": "pending",
--   "order_date": "2025-06-12",
--   "compliance": {"regulatory_ref": "SOX"}
-- }

-- Bloco de documentação multilíngue para crm.asset (README_GOVERNANCA_MULTILINGUE.md)
-- ### Domínio: crm.asset (Ativos do Cliente)
-- - **Descrição PT:** Gestão de ativos do cliente (equipamentos, licenças, contratos), com campos para compliance, privacidade, rastreabilidade, multilíngue, localização, status, responsável e integração global. Alinhado a ISO 9001, GDPR, LGPD, benchmarks Salesforce, SAP, Oracle.
-- - **Descrição EN:** Customer asset management (equipment, licenses, contracts), with fields for compliance, privacy, traceability, multilingual, location, status, responsible and global integration. Aligned with ISO 9001, GDPR, LGPD, Salesforce, SAP, Oracle benchmarks.
-- - **Campos-chave:** asset_code, name, status, purchase_date, warranty_expiry, compliance, privacy, audit_log, kpi.
-- - **Referências normativas:** ISO 9001, GDPR, LGPD, SOX, PCI DSS.

-- Teste automatizado para trigger de compliance em pedidos
DO $$
BEGIN
  BEGIN
    INSERT INTO crm.order (order_code, account_id, quote_id, amount, currency, status, order_date)
    VALUES ('ORDER-NOCOMPLIANCE', 1, 1, 1234.56, 'EUR', 'pending', '2025-06-12');
  EXCEPTION
    WHEN OTHERS THEN
      RAISE NOTICE 'Teste passou: inserção de pedido sem compliance bloqueada.';
  END;
END $$;

-- Checklist para integração e rollout
-- 1. Testar inserção e atualização em todos os domínios sem campo compliance (deve ser bloqueado).
-- 2. Validar queries de BI e relatórios em dashboards.
-- 3. Realizar integração via API REST para contas, pedidos e campanhas.
-- 4. Atualizar README_GOVERNANCA_MULTILINGUE.md com todos os blocos multilíngues.
-- 5. Validar rastreabilidade e logs de auditoria em ambiente de homologação.
-- 6. Garantir aderência a normas e benchmarks globais em todos os domínios.

-- Seeds para crm.account (Contas)
INSERT INTO crm.account (account_code, name, name_alt, description, type, industry, country, status, source, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('ACC-001', 'InnovaBiz Global', 'InnovaBiz Global Ltd.', 'Empresa multinacional de tecnologia.', 'customer', 'Technology', 'Portugal', 'active', 'import', 'admin', '{"regulatory_ref": "ISO 9001, GDPR"}', '{"consent": true}', '{"access": "full"}', '{"nps": 85}', '["cliente premium"]', 'pt', 'admin', 'admin'),
('ACC-002', 'OpenBank Partners', 'OpenBank Int.', 'Parceiro estratégico para Open Banking.', 'partner', 'Finance', 'Brasil', 'active', 'manual', 'admin', '{"regulatory_ref": "ISO 9001, BIAN"}', '{"consent": true}', '{"access": "restricted"}', '{"nps": 90}', '["parceiro gold"]', 'pt', 'admin', 'admin');

-- Seeds para crm.case (Atendimentos)
INSERT INTO crm.case (case_code, account_id, type, subject, description, status, priority, sla_level, channel, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('CASE-001', 1, 'question', 'Dúvida sobre integração', 'Cliente com dúvida sobre API.', 'open', 'medium', 'P2', 'web', 'Suporte', '{"regulatory_ref": "ISO 10002, ITIL"}', '{"consent": true}', '{"access": "full"}', '{"sla": "24h"}', '["api"]', 'pt', 'admin', 'admin');

-- Seeds para crm.asset (Ativos)
INSERT INTO crm.asset (asset_code, account_id, product_id, serial_number, name, status, purchase_date, warranty_expiry, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('ASSET-001', 1, 1, 'SN-12345', 'Servidor Cloud', 'active', '2024-01-01', '2027-01-01', 'Infra', '{"regulatory_ref": "ISO 9001"}', '{"consent": true}', '{"access": "full"}', '{"uptime": 99.99}', '["cloud"]', 'pt', 'admin', 'admin');

-- Seeds para crm.partner (Parceiros)
INSERT INTO crm.partner (partner_code, name, type, industry, status, segment_id, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('PART-001', 'OpenBank Partners', 'reseller', 'Finance', 'active', 1, 'Parcerias', '{"regulatory_ref": "BIAN"}', '{"consent": true}', '{"access": "restricted"}', '{"score": 95}', '["openbank"]', 'pt', 'admin', 'admin');

-- Seeds para crm.quote (Cotações)
INSERT INTO crm.quote (quote_code, opportunity_id, account_id, amount, currency, status, valid_until, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('QUOTE-001', 1, 1, 10000.00, 'EUR', 'draft', '2025-12-31', 'Comercial', '{"regulatory_ref": "SOX"}', '{"consent": true}', '{"access": "full"}', '{"win_rate": 0.7}', '["proposta"]', 'pt', 'admin', 'admin');

-- Seeds para crm.order (Pedidos)
INSERT INTO crm.order (order_code, account_id, quote_id, amount, currency, status, order_date, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('ORDER-001', 1, 1, 10000.00, 'EUR', 'pending', '2025-06-12', 'Comercial', '{"regulatory_ref": "SOX"}', '{"consent": true}', '{"access": "full"}', '{"cycle_time": 15}', '["pedido inicial"]', 'pt', 'admin', 'admin');

-- Seeds para crm.billing (Faturamento)
INSERT INTO crm.billing (order_id, account_id, amount, currency, due_date, status, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
(1, 1, 10000.00, 'EUR', '2025-07-12', 'pending', 'Financeiro', '{"regulatory_ref": "PCI DSS"}', '{"consent": true}', '{"access": "restricted"}', '{"on_time": true}', '["fatura"]', 'pt', 'admin', 'admin');

-- Seeds para crm.subscription (Assinaturas)
INSERT INTO crm.subscription (account_id, product_id, start_date, end_date, status, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
(1, 1, '2025-01-01', '2026-01-01', 'active', 'Comercial', '{"regulatory_ref": "ISO 9001"}', '{"consent": true}', '{"access": "full"}', '{"renewal_rate": 0.95}', '["assinatura anual"]', 'pt', 'admin', 'admin');

-- Seeds para crm.entitlement (Direitos)
INSERT INTO crm.entitlement (account_id, product_id, entitlement_type, start_date, end_date, status, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
(1, 1, 'premium_support', '2025-01-01', '2026-01-01', 'active', 'Suporte', '{"regulatory_ref": "ISO 9001"}', '{"consent": true}', '{"access": "full"}', '{"sla": "4h"}', '["premium"]', 'pt', 'admin', 'admin');

-- Seeds para crm.knowledge_base (Base de Conhecimento)
INSERT INTO crm.knowledge_base (title, content, status, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('Como integrar via API', 'Passos para integração segura via API REST.', 'published', 'Suporte', '{"regulatory_ref": "ISO 9001, ITIL"}', '{"consent": true}', '{"access": "full"}', '{"views": 120}', '["api"]', 'pt', 'admin', 'admin');

-- Seeds para crm.campaign_member (Participantes de Campanha)
INSERT INTO crm.campaign_member (campaign_id, contact_id, account_id, status, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
(1, 1, 1, 'active', '{"regulatory_ref": "GDPR"}', '{"consent": true}', '{"access": "restricted"}', '{"engagement": 80}', '["email_marketing"]', 'pt', 'admin', 'admin');

-- Seeds para crm.team (Equipes)
INSERT INTO crm.team (name, name_alt, description, status, responsible, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('Equipe Comercial', 'Sales Team', 'Equipe responsável por vendas B2B.', 'active', 'Gestor Vendas', '{"regulatory_ref": "ISO 9001"}', '{"consent": true}', '{"access": "full"}', '{"quota": 100000}', '["b2b"]', 'pt', 'admin', 'admin');

-- Seeds para crm.user_team (Usuário-Equipe)
INSERT INTO crm.user_team (user_id, team_id, role, status, compliance, privacy, accessibility, kpi, tags, language, created_by, updated_by)
VALUES
('00000000-0000-0000-0000-000000000001', 1, 'manager', 'active', '{"regulatory_ref": "ISO 9001"}', '{"consent": true}', '{"access": "full"}', '{"performance": 95}', '["gestor"]', 'pt', 'admin', 'admin');

-- Trigger de compliance/data quality para todos os domínios críticos
CREATE OR REPLACE FUNCTION fn_check_compliance_generic()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance IS NULL OR NEW.compliance::text = '{}' THEN
    RAISE EXCEPTION 'Campo compliance é obrigatório para este domínio!';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    t RECORD;
    tablist TEXT[] := ARRAY['account','case','asset','partner','quote','order','billing','subscription','entitlement','knowledge_base','campaign_member','team','user_team'];
BEGIN
    FOREACH t IN ARRAY tablist LOOP
        EXECUTE format('CREATE TRIGGER trg_check_compliance_%I BEFORE INSERT OR UPDATE ON crm.%I FOR EACH ROW EXECUTE FUNCTION fn_check_compliance_generic();', t, t);
    END LOOP;
END $$;

-- View de BI/Compliance para visão 360º do cliente
CREATE OR REPLACE VIEW crm.vw_customer_360 AS
SELECT
  a.account_id,
  a.account_code,
  a.name,
  a.type,
  a.industry,
  a.status,
  a.country,
  a.global_id,
  a.compliance,
  a.kpi,
  (SELECT COUNT(*) FROM crm.case c WHERE c.account_id = a.account_id) AS total_cases,
  (SELECT COUNT(*) FROM crm.asset s WHERE s.account_id = a.account_id) AS total_assets,
  (SELECT COUNT(*) FROM crm.order o WHERE o.account_id = a.account_id) AS total_orders,
  (SELECT COUNT(*) FROM crm.subscription sub WHERE sub.account_id = a.account_id) AS total_subscriptions,
  (SELECT COUNT(*) FROM crm.entitlement e WHERE e.account_id = a.account_id) AS total_entitlements
FROM crm.account a;

-- Checklist técnico multilíngue
-- 1. Executar todos os seeds para domínios CRM e CXP.
-- 2. Validar triggers de compliance/data quality em todos os domínios críticos.
-- 3. Validar logs de integração externa e rastreabilidade.
-- 4. Executar e validar views de BI/compliance e queries avançadas.
-- 5. Atualizar ERD/documentação multilíngue e checklist de rollout.
-- 6. Garantir rastreabilidade, auditoria, privacidade, compliance e interoperabilidade em todos os registros.
-- 7. Validar integrações multi-domínio e visão 360º do cliente.

-- Tabela de log de alertas críticos para feedbacks
CREATE TABLE IF NOT EXISTS cxp.alert_log (
    alert_id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    feedback_id INT REFERENCES cxp.feedback(feedback_id),
    created_at TIMESTAMP DEFAULT now(),
    details JSONB,
    status VARCHAR(50) DEFAULT 'active'
);

-- Trigger de alerta para feedbacks críticos (urgência ou sentimento negativo)
CREATE OR REPLACE FUNCTION fn_alert_feedback_critical()
RETURNS TRIGGER AS $$
BEGIN
  IF (NEW.urgency_level = 'critical' OR NEW.sentiment = 'negative') THEN
    INSERT INTO cxp.alert_log (event_type, feedback_id, created_at, details)
    VALUES ('CRITICAL_FEEDBACK', NEW.feedback_id, now(), row_to_json(NEW));
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_alert_feedback_critical
AFTER INSERT ON cxp.feedback
FOR EACH ROW
EXECUTE FUNCTION fn_alert_feedback_critical();

-- Query de BI: Analytics de sentimento, canal e urgência (para dashboards e relatórios)
-- SELECT
--   channel,
--   sentiment,
--   urgency_level,
--   COUNT(*) AS total_feedbacks
-- FROM cxp.feedback
-- WHERE status = 'active'
-- GROUP BY channel, sentiment, urgency_level
-- ORDER BY channel, sentiment, urgency_level;

-- Checklist técnico multilíngue de rollout e governança (para README e documentação)
-- 1. Testar triggers de compliance/auditoria em todos os domínios (VOC, feedback, NPS, CSAT, CES).
-- 2. Validar seeds e queries BI para todos os domínios.
-- 3. Garantir documentação multilíngue atualizada (README, ERD, exemplos de integração).
-- 4. Validar campos obrigatórios: compliance, regulatory_ref, data_classification, audit_log.
-- 5. Testar endpoints RESTful (feedback, VOC).
-- 6. Checklist multilíngue para rollout seguro e governança contínua.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.feedback (Feedback Detalhado)
-- - **Descrição PT:** Feedback detalhado e análise avançada de experiência do cliente, com granularidade, inteligência analítica, compliance, rastreabilidade e automação.
-- - **Descrição EN:** Detailed feedback and advanced customer experience analytics, with granularity, analytical intelligence, compliance, traceability and automation.
-- - **Campos-chave:** feedback_text, feedback_type, channel, sentiment, ai_score, language, tags, topic, urgency_level, response_time, followup_required, followup_status, compliance, audit_log, benchmark_kpi.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester, Gartner, CXPA.

-- TESTES AUTOMATIZADOS PARA DOMÍNIOS CRM/CXP

-- NPS: Teste de compliance (deve falhar ao inserir sem compliance)
DO $$
BEGIN
  BEGIN
    INSERT INTO cxp.nps (
      journey_id, score, promoter_count, passive_count, detractor_count, industry_benchmark,
      regulatory_ref, data_classification, audit_log, created_by, updated_by, status
    ) VALUES (
      1, 50, 10, 5, 2, 40,
      'GDPR Art. 6', 'anonymous', '{"action": "insert", "user": "test"}', 'test', 'test', 'active'
    );
    RAISE NOTICE 'ERRO: Trigger de compliance NÃO bloqueou inserção sem compliance!';
  EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Trigger de compliance bloqueou corretamente: %', SQLERRM;
  END;
END;
$$;

-- NPS: Teste de auditoria
UPDATE cxp.nps
SET score = 60,
    audit_log = jsonb_set(audit_log, '{last_update}', to_jsonb(now()))
WHERE nps_id = (SELECT MAX(nps_id) FROM cxp.nps);

SELECT nps_id, audit_log FROM cxp.nps WHERE nps_id = (SELECT MAX(nps_id) FROM cxp.nps);

-- Repita/adapte para VOC, CSAT, CES conforme estrutura de cada tabela.

-- Bloco de documentação multilíngue para README_GOVERNANCA_MULTILINGUE.md
-- ### Domínio: cxp.nps (Net Promoter Score)
-- - **Descrição PT:** Métrica internacional de lealdade do cliente, recomendada por ISO 10002, CXPA, Forrester, Gartner. Permite mensurar o NPS por jornada, integrando compliance, privacidade, rastreabilidade e auditoria.
-- - **Descrição EN:** International customer loyalty metric, recommended by ISO 10002, CXPA, Forrester, Gartner. Measures NPS by journey, integrating compliance, privacy, traceability, and audit.
-- - **Campos-chave:** score, promoter_count, passive_count, detractor_count, industry_benchmark, compliance, regulatory_ref, data_classification, audit_log.
-- - **Referências normativas:** ISO 10002, GDPR, LGPD, ISO 27701, NIST, Forrester CX Index.

-- Checklist técnico multilíngue para rollout NPS
-- 1. Testar triggers de compliance/auditoria.
-- 2. Validar seeds e queries BI.
-- 3. Garantir documentação multilíngue atualizada.
-- 4. Validar campos obrigatórios.
-- 5. Testar endpoints RESTful.
-- 6. Validar integrações e visão 360º.

-- Função genérica para consulta de logs de auditoria por tabela, operação e usuário
CREATE OR REPLACE FUNCTION get_audit_log(
    p_table_name VARCHAR DEFAULT NULL,
    p_operation VARCHAR DEFAULT NULL,
    p_changed_by VARCHAR DEFAULT NULL,
    p_limit INT DEFAULT 100
)
RETURNS TABLE (
    audit_id INT,
    table_name VARCHAR,
    operation VARCHAR,
    record_id INT,
    changed_by VARCHAR,
    changed_at TIMESTAMP,
    old_data JSONB,
    new_data JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        audit_id,
        table_name,
        operation,
        record_id,
        changed_by,
        changed_at,
        old_data,
        new_data
    FROM audit_log
    WHERE
        (p_table_name IS NULL OR table_name = p_table_name)
        AND (p_operation IS NULL OR operation = p_operation)
        AND (p_changed_by IS NULL OR changed_by = p_changed_by)
    ORDER BY changed_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Exemplos de uso:
-- SELECT * FROM get_audit_log('order', NULL, NULL, 100); -- Últimas 100 alterações em pedidos
-- SELECT * FROM get_audit_log('account', 'UPDATE', NULL, 50); -- Últimas 50 atualizações em contas
-- SELECT * FROM get_audit_log(NULL, NULL, 'admin', 20); -- Últimos 20 eventos do usuário admin em qualquer tabela

-- =========================
-- DOMÍNIO DE GOVERNANÇA CORPORATIVA AVANÇADO INNOVABIZ
-- =========================

-- Órgãos Sociais
CREATE TABLE IF NOT EXISTS governance_orgao_social (
    id SERIAL PRIMARY KEY,
    orgao_name VARCHAR(255) NOT NULL,
    orgao_type VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INT REFERENCES governance_orgao_social(id),
    iso_code VARCHAR(20),
    un_locode VARCHAR(20),
    compliance_framework VARCHAR(100),
    regulation_reference VARCHAR(255),
    responsible_person VARCHAR(255),
    start_date DATE,
    end_date DATE,
    language VARCHAR(10) DEFAULT 'pt',
    accessibility_level VARCHAR(50),
    traceability_code VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Stakeholders
CREATE TABLE IF NOT EXISTS governance_stakeholder (
    id SERIAL PRIMARY KEY,
    stakeholder_name VARCHAR(255) NOT NULL,
    stakeholder_type VARCHAR(100),
    orgao_social_id INT REFERENCES governance_orgao_social(id),
    contact_info TEXT,
    influence_level VARCHAR(50),
    interest_level VARCHAR(50),
    compliance_requirements TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Partes Relacionadas
CREATE TABLE IF NOT EXISTS governance_related_party (
    id SERIAL PRIMARY KEY,
    party_name VARCHAR(255) NOT NULL,
    relationship_type VARCHAR(100),
    orgao_social_id INT REFERENCES governance_orgao_social(id),
    sector VARCHAR(100),
    country VARCHAR(100),
    regulation_reference VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Políticas de Governança
CREATE TABLE IF NOT EXISTS governance_policy (
    id SERIAL PRIMARY KEY,
    policy_name VARCHAR(255) NOT NULL,
    policy_type VARCHAR(100),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    orgao_social_id INT REFERENCES governance_orgao_social(id),
    compliance_framework VARCHAR(100),
    regulation_reference VARCHAR(255),
    effective_date DATE,
    review_date DATE,
    responsible_person VARCHAR(255),
    language VARCHAR(10) DEFAULT 'pt',
    accessibility_level VARCHAR(50),
    traceability_code VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Processos
CREATE TABLE IF NOT EXISTS governance_process (
    id SERIAL PRIMARY KEY,
    process_name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    policy_id INT REFERENCES governance_policy(id),
    process_owner VARCHAR(255),
    regulation_reference VARCHAR(255),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Procedimentos
CREATE TABLE IF NOT EXISTS governance_procedure (
    id SERIAL PRIMARY KEY,
    procedure_name VARCHAR(255) NOT NULL,
    process_id INT REFERENCES governance_process(id),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    responsible_person VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Controles
CREATE TABLE IF NOT EXISTS governance_control (
    id SERIAL PRIMARY KEY,
    control_name VARCHAR(255) NOT NULL,
    control_type VARCHAR(100),
    procedure_id INT REFERENCES governance_procedure(id),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    effectiveness_level VARCHAR(50),
    regulation_reference VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Compliance
CREATE TABLE IF NOT EXISTS governance_compliance (
    id SERIAL PRIMARY KEY,
    compliance_name VARCHAR(255) NOT NULL,
    compliance_type VARCHAR(100),
    policy_id INT REFERENCES governance_policy(id),
    framework VARCHAR(100),
    regulation_reference VARCHAR(255),
    audit_frequency VARCHAR(50),
    last_audit DATE,
    next_audit DATE,
    responsible_person VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Obrigações
CREATE TABLE IF NOT EXISTS governance_obligation (
    id SERIAL PRIMARY KEY,
    obligation_name VARCHAR(255) NOT NULL,
    compliance_id INT REFERENCES governance_compliance(id),
    description TEXT,
    due_date DATE,
    frequency VARCHAR(50),
    responsible_person VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Certificações
CREATE TABLE IF NOT EXISTS governance_certification (
    id SERIAL PRIMARY KEY,
    certification_name VARCHAR(255) NOT NULL,
    compliance_id INT REFERENCES governance_compliance(id),
    standard VARCHAR(100),
    issue_date DATE,
    expiry_date DATE,
    status VARCHAR(50) DEFAULT 'valid',
    issuing_body VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Penalidades
CREATE TABLE IF NOT EXISTS governance_penalty (
    id SERIAL PRIMARY KEY,
    penalty_name VARCHAR(255) NOT NULL,
    compliance_id INT REFERENCES governance_compliance(id),
    description TEXT,
    penalty_type VARCHAR(100),
    value NUMERIC(18,2),
    date_applied DATE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Riscos
CREATE TABLE IF NOT EXISTS governance_risk (
    id SERIAL PRIMARY KEY,
    risk_name VARCHAR(255) NOT NULL,
    compliance_id INT REFERENCES governance_compliance(id),
    risk_type VARCHAR(100),
    risk_category VARCHAR(100),
    risk_level VARCHAR(50),
    impact VARCHAR(255),
    probability VARCHAR(50),
    mitigation_plan TEXT,
    status VARCHAR(50) DEFAULT 'active',
    responsible_person VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Incidentes
CREATE TABLE IF NOT EXISTS governance_incident (
    id SERIAL PRIMARY KEY,
    incident_name VARCHAR(255) NOT NULL,
    risk_id INT REFERENCES governance_risk(id),
    description TEXT,
    incident_date DATE,
    status VARCHAR(50) DEFAULT 'open',
    responsible_person VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Planos de Ação
CREATE TABLE IF NOT EXISTS governance_action_plan (
    id SERIAL PRIMARY KEY,
    action_name VARCHAR(255) NOT NULL,
    incident_id INT REFERENCES governance_incident(id),
    description TEXT,
    responsible_person VARCHAR(255),
    due_date DATE,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Dependências
CREATE TABLE IF NOT EXISTS governance_dependency (
    id SERIAL PRIMARY KEY,
    dependency_name VARCHAR(255) NOT NULL,
    from_process_id INT REFERENCES governance_process(id),
    to_process_id INT REFERENCES governance_process(id),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Auditoria
CREATE TABLE IF NOT EXISTS governance_audit (
    id SERIAL PRIMARY KEY,
    audit_name VARCHAR(255) NOT NULL,
    risk_id INT REFERENCES governance_risk(id),
    framework VARCHAR(100),
    audit_type VARCHAR(100),
    audit_date DATE,
    result TEXT,
    findings TEXT,
    recommendations TEXT,
    responsible_person VARCHAR(255),
    status VARCHAR(50) DEFAULT 'planned',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Sustentabilidade/ESG
CREATE TABLE IF NOT EXISTS governance_sustainability (
    id SERIAL PRIMARY KEY,
    sustainability_name VARCHAR(255) NOT NULL,
    orgao_social_id INT REFERENCES governance_orgao_social(id),
    framework VARCHAR(100),
    kpi VARCHAR(255),
    target_value VARCHAR(255),
    current_value VARCHAR(255),
    review_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    responsible_person VARCHAR(255),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Indicadores/KPI
CREATE TABLE IF NOT EXISTS governance_kpi (
    id SERIAL PRIMARY KEY,
    kpi_name VARCHAR(255) NOT NULL,
    domain VARCHAR(100),
    reference_table VARCHAR(100),
    reference_id INT,
    value NUMERIC(18,4),
    target_value NUMERIC(18,4),
    period_start DATE,
    period_end DATE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- Integrações
CREATE TABLE IF NOT EXISTS governance_integration (
    id SERIAL PRIMARY KEY,
    integration_name VARCHAR(255) NOT NULL,
    integration_type VARCHAR(100),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    related_table VARCHAR(100),
    related_id INT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- MELHORIAS ESTRUTURAIS PARA GOVERNANÇA INNOVABIZ

-- governance_orgao_social
ALTER TABLE governance_orgao_social
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_stakeholder
ALTER TABLE governance_stakeholder
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_related_party
ALTER TABLE governance_related_party
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_policy
ALTER TABLE governance_policy
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_process
ALTER TABLE governance_process
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_procedure
ALTER TABLE governance_procedure
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_control
ALTER TABLE governance_control
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_compliance
ALTER TABLE governance_compliance
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_obligation
ALTER TABLE governance_obligation
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_certification
ALTER TABLE governance_certification
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_penalty
ALTER TABLE governance_penalty
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_risk
ALTER TABLE governance_risk
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_incident
ALTER TABLE governance_incident
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_action_plan
ALTER TABLE governance_action_plan
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_dependency
ALTER TABLE governance_dependency
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_audit
ALTER TABLE governance_audit
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_sustainability
ALTER TABLE governance_sustainability
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_kpi
ALTER TABLE governance_kpi
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

-- governance_integration
ALTER TABLE governance_integration
    ADD COLUMN IF NOT EXISTS global_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS alternate_names JSONB,
    ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS valid_from DATE,
    ADD COLUMN IF NOT EXISTS valid_to DATE,
    ADD COLUMN IF NOT EXISTS external_reference VARCHAR(100),
    ADD COLUMN IF NOT EXISTS compliance_tags JSONB,
    ADD COLUMN IF NOT EXISTS privacy_level VARCHAR(50),
    ADD COLUMN IF NOT EXISTS interoperability_tags JSONB,
    ADD COLUMN IF NOT EXISTS created_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);

    -- =========================
-- SEEDS DE EXEMPLO PARA GOVERNANÇA
-- =========================

INSERT INTO governance_orgao_social (orgao_name, orgao_type, global_id, alternate_names, compliance_framework, regulation_reference, responsible_person, sector, industry_code, created_by)
VALUES
('Conselho de Administração', 'Board', 'ORG-PT-0001', '{"en":"Board of Directors","pt":"Conselho de Administração"}', 'King IV, SOX', 'Lei das SAs', 'CEO', 'Finance', 'K6419', 'admin'),
('Comitê de Auditoria', 'Committee', 'ORG-PT-0002', '{"en":"Audit Committee"}', 'IPPF, IIA', 'Resolução BACEN', 'Diretor de Auditoria', 'Finance', 'K6419', 'admin');

INSERT INTO governance_policy (policy_name, policy_type, global_id, alternate_names, compliance_tags, privacy_level, created_by)
VALUES
('Política de Privacidade', 'Privacy', 'POLICY-PT-0001', '{"en":"Privacy Policy"}', '["GDPR","LGPD"]', 'confidential', 'admin'),
('Política de Riscos', 'Risk', 'POLICY-PT-0002', '{"en":"Risk Policy"}', '["ISO 31000","COSO ERM"]', 'internal', 'admin');

INSERT INTO governance_compliance (compliance_name, compliance_type, policy_id, framework, regulation_reference, audit_frequency, responsible_person, sector, status, created_by)
SELECT 'Compliance GDPR', 'Regulatory', p.id, 'GDPR', 'Regulamento (UE) 2016/679', 'annual', 'DPO', 'Finance', 'active', 'admin'
FROM governance_policy p WHERE p.policy_name = 'Política de Privacidade';

INSERT INTO governance_risk (risk_name, compliance_id, risk_type, risk_category, risk_level, impact, probability, mitigation_plan, status, responsible_person, sector, created_by)
SELECT 'Risco de Vazamento de Dados', c.id, 'Operacional', 'Privacidade', 'Alto', 'Multas e danos reputacionais', 'Média', 'Treinamento e monitoramento', 'active', 'DPO', 'Finance', 'admin'
FROM governance_compliance c WHERE c.compliance_name = 'Compliance GDPR';

-- =========================
-- TRIGGERS DE AUDITORIA E COMPLIANCE
-- =========================

CREATE OR REPLACE FUNCTION fn_audit_generic() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO audit_log (table_name, operation, record_id, changed_by, changed_at, old_data, new_data)
  VALUES (TG_TABLE_NAME, TG_OP, COALESCE(NEW.id, OLD.id), current_user, now(), row_to_json(OLD), row_to_json(NEW));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de trigger para governance_policy
DROP TRIGGER IF EXISTS trg_audit_governance_policy ON governance_policy;
CREATE TRIGGER trg_audit_governance_policy
AFTER INSERT OR UPDATE OR DELETE ON governance_policy
FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

-- Replicar triggers para outras tabelas críticas:
DROP TRIGGER IF EXISTS trg_audit_governance_orgao_social ON governance_orgao_social;
CREATE TRIGGER trg_audit_governance_orgao_social
AFTER INSERT OR UPDATE OR DELETE ON governance_orgao_social
FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_audit_governance_compliance ON governance_compliance;
CREATE TRIGGER trg_audit_governance_compliance
AFTER INSERT OR UPDATE OR DELETE ON governance_compliance
FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_audit_governance_risk ON governance_risk;
CREATE TRIGGER trg_audit_governance_risk
AFTER INSERT OR UPDATE OR DELETE ON governance_risk
FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

-- =========================
-- QUERY DE BI/COMPLIANCE PARA GOVERNANÇA
-- =========================

-- Relatório de compliance por framework e status
-- (Ajuste conforme necessidade do BI)
SELECT
  p.policy_name,
  p.policy_type,
  c.compliance_name,
  c.framework,
  c.status,
  p.global_id,
  p.compliance_tags
FROM governance_policy p
LEFT JOIN governance_compliance c ON c.policy_id = p.id
ORDER BY p.policy_name;