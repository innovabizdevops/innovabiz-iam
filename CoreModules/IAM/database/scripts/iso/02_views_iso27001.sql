-- INNOVABIZ - IAM ISO 27001 Views
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Views para análise e apresentação de dados de compliance ISO 27001 com foco em saúde.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- View para visualizar detalhes completos dos controles ISO 27001
CREATE OR REPLACE VIEW vw_iso27001_controls_detail AS
SELECT
    c.id,
    c.control_id,
    c.section,
    c.name,
    c.description,
    c.healthcare_applicability,
    c.implementation_guidance,
    c.validation_rules,
    c.reference_links,
    c.category,
    c.is_active,
    c.created_at,
    c.updated_at,
    (SELECT COUNT(*) FROM iso27001_framework_mapping fm WHERE fm.iso_control_id = c.id) AS framework_mappings_count,
    ARRAY(
        SELECT DISTINCT rf.name 
        FROM iso27001_framework_mapping fm 
        JOIN regulatory_frameworks rf ON fm.framework_id = rf.id 
        WHERE fm.iso_control_id = c.id
    ) AS mapped_frameworks
FROM
    iso27001_controls c
ORDER BY
    c.control_id;

COMMENT ON VIEW vw_iso27001_controls_detail IS 'Visão detalhada dos controles ISO 27001 com mapeamentos para outros frameworks';

-- View para visualizar resumo das avaliações ISO 27001
CREATE OR REPLACE VIEW vw_iso27001_assessment_summary AS
SELECT
    a.id AS assessment_id,
    a.name AS assessment_name,
    a.description,
    a.organization_id,
    o.name AS organization_name,
    a.start_date,
    a.end_date,
    a.status,
    a.healthcare_specific,
    a.version,
    a.score,
    a.created_at,
    a.updated_at,
    u.full_name AS created_by_name,
    COUNT(cr.id) AS total_controls_assessed,
    SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END) AS compliant_controls_count,
    SUM(CASE WHEN cr.status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial_compliance_controls_count,
    SUM(CASE WHEN cr.status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant_controls_count,
    SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END) AS not_applicable_controls_count,
    ROUND(
        (SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
        NULLIF(COUNT(cr.id) - SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END), 0)) * 100, 
        2
    ) AS compliance_percentage,
    COUNT(ap.id) AS open_action_plans_count,
    rf.name AS framework_name,
    CASE 
        WHEN a.healthcare_specific THEN 'Healthcare-specific'
        ELSE 'General'
    END AS assessment_type
FROM
    iso27001_assessments a
    LEFT JOIN organizations o ON a.organization_id = o.id
    LEFT JOIN users u ON a.created_by = u.id
    LEFT JOIN iso27001_control_results cr ON a.id = cr.assessment_id
    LEFT JOIN iso27001_action_plans ap ON cr.id = ap.control_result_id AND ap.status IN ('open', 'in_progress')
    LEFT JOIN regulatory_frameworks rf ON a.framework_id = rf.id
GROUP BY
    a.id, o.name, u.full_name, rf.name
ORDER BY
    a.start_date DESC;

COMMENT ON VIEW vw_iso27001_assessment_summary IS 'Resumo das avaliações ISO 27001 com estatísticas de compliance';

-- View para visualizar resultados detalhados dos controles por avaliação
CREATE OR REPLACE VIEW vw_iso27001_control_results_detail AS
SELECT
    cr.id,
    cr.assessment_id,
    a.name AS assessment_name,
    a.organization_id,
    o.name AS organization_name,
    cr.control_id,
    c.control_id AS control_code,
    c.section,
    c.name AS control_name,
    c.description AS control_description,
    c.healthcare_applicability,
    cr.status,
    cr.score,
    cr.implementation_status,
    cr.evidence,
    cr.notes,
    cr.issues_found,
    cr.recommendations,
    cr.healthcare_specific_findings,
    u.full_name AS assessed_by,
    cr.created_at,
    cr.updated_at,
    c.category,
    a.healthcare_specific,
    (SELECT COUNT(*) FROM iso27001_action_plans ap WHERE ap.control_result_id = cr.id) AS action_plans_count,
    (SELECT COUNT(*) FROM iso27001_action_plans ap WHERE ap.control_result_id = cr.id AND ap.status IN ('open', 'in_progress')) AS open_action_plans_count
FROM
    iso27001_control_results cr
    JOIN iso27001_controls c ON cr.control_id = c.id
    JOIN iso27001_assessments a ON cr.assessment_id = a.id
    JOIN organizations o ON a.organization_id = o.id
    LEFT JOIN users u ON cr.created_by = u.id
ORDER BY
    a.name, c.control_id;

COMMENT ON VIEW vw_iso27001_control_results_detail IS 'Detalhes dos resultados de avaliação de controles ISO 27001';

-- View para visualizar planos de ação ISO 27001
CREATE OR REPLACE VIEW vw_iso27001_action_plans AS
SELECT
    ap.id,
    ap.organization_id,
    o.name AS organization_name,
    ap.assessment_id,
    a.name AS assessment_name,
    ap.control_result_id,
    cr.status AS control_status,
    c.control_id AS control_code,
    c.name AS control_name,
    ap.title,
    ap.description,
    ap.priority,
    ap.status,
    ap.due_date,
    ap.assigned_to,
    u1.full_name AS assigned_to_name,
    ap.created_by,
    u2.full_name AS created_by_name,
    ap.created_at,
    ap.updated_at,
    ap.completed_at,
    ap.completed_by,
    u3.full_name AS completed_by_name,
    ap.completion_notes,
    ap.estimated_effort,
    ap.healthcare_related,
    CASE 
        WHEN ap.due_date < CURRENT_DATE AND ap.status IN ('open', 'in_progress') THEN TRUE
        ELSE FALSE
    END AS is_overdue,
    CASE 
        WHEN ap.due_date IS NOT NULL THEN CURRENT_DATE - ap.due_date
        ELSE NULL
    END AS days_overdue,
    a.healthcare_specific
FROM
    iso27001_action_plans ap
    JOIN organizations o ON ap.organization_id = o.id
    LEFT JOIN iso27001_assessments a ON ap.assessment_id = a.id
    LEFT JOIN iso27001_control_results cr ON ap.control_result_id = cr.id
    LEFT JOIN iso27001_controls c ON cr.control_id = c.id
    LEFT JOIN users u1 ON ap.assigned_to = u1.id
    LEFT JOIN users u2 ON ap.created_by = u2.id
    LEFT JOIN users u3 ON ap.completed_by = u3.id
ORDER BY
    ap.priority DESC, ap.due_date ASC;

COMMENT ON VIEW vw_iso27001_action_plans IS 'Visão detalhada dos planos de ação para compliance com ISO 27001';

-- View para visualizar mapeamentos entre ISO 27001 e outros frameworks
CREATE OR REPLACE VIEW vw_iso27001_framework_mappings AS
SELECT
    fm.id,
    fm.iso_control_id,
    ic.control_id AS iso_control_code,
    ic.name AS iso_control_name,
    ic.section AS iso_section,
    fm.framework_id,
    rf.code AS framework_code,
    rf.name AS framework_name,
    fm.framework_control_id,
    fm.framework_control_name,
    fm.mapping_type,
    fm.mapping_strength,
    fm.notes,
    ic.healthcare_applicability,
    fm.created_at,
    fm.updated_at,
    u.full_name AS created_by_name
FROM
    iso27001_framework_mapping fm
    JOIN iso27001_controls ic ON fm.iso_control_id = ic.id
    JOIN regulatory_frameworks rf ON fm.framework_id = rf.id
    LEFT JOIN users u ON fm.created_by = u.id
ORDER BY
    ic.control_id, rf.name, fm.framework_control_id;

COMMENT ON VIEW vw_iso27001_framework_mappings IS 'Mapeamentos entre controles ISO 27001 e outros frameworks regulatórios';

-- View para visualizar documentos ISO 27001 com metadados
CREATE OR REPLACE VIEW vw_iso27001_documents_detail AS
SELECT
    d.id,
    d.organization_id,
    o.name AS organization_name,
    d.title,
    d.document_type,
    d.description,
    d.version,
    d.status,
    d.content_url,
    d.storage_path,
    d.file_type,
    d.file_size,
    d.created_at,
    d.updated_at,
    u1.full_name AS created_by_name,
    u2.full_name AS updated_by_name,
    d.approved_by,
    u3.full_name AS approved_by_name,
    d.approved_at,
    d.related_controls,
    d.last_review_date,
    d.next_review_date,
    d.healthcare_specific,
    CASE 
        WHEN d.next_review_date < CURRENT_DATE AND d.status = 'published' THEN TRUE
        ELSE FALSE
    END AS review_overdue,
    ARRAY(
        SELECT ic.control_id 
        FROM iso27001_controls ic
        WHERE ic.id = ANY(ARRAY(SELECT jsonb_array_elements_text(d.related_controls)::UUID))
    ) AS related_control_codes,
    CASE
        WHEN d.next_review_date IS NOT NULL THEN 
            CASE
                WHEN d.next_review_date - CURRENT_DATE <= 30 AND d.next_review_date >= CURRENT_DATE THEN 'Due soon'
                WHEN d.next_review_date < CURRENT_DATE THEN 'Overdue'
                ELSE 'On schedule'
            END
        ELSE 'No review scheduled'
    END AS review_status
FROM
    iso27001_documents d
    JOIN organizations o ON d.organization_id = o.id
    LEFT JOIN users u1 ON d.created_by = u1.id
    LEFT JOIN users u2 ON d.updated_by = u2.id
    LEFT JOIN users u3 ON d.approved_by = u3.id
ORDER BY
    d.organization_id, d.document_type, d.title;

COMMENT ON VIEW vw_iso27001_documents_detail IS 'Visão detalhada dos documentos ISO 27001 com metadados e status de revisão';

-- View para dashboard de compliance ISO 27001
CREATE OR REPLACE VIEW vw_iso27001_compliance_dashboard AS
WITH assessment_stats AS (
    SELECT
        a.organization_id,
        o.name AS organization_name,
        a.healthcare_specific,
        COUNT(DISTINCT a.id) AS total_assessments,
        COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'completed') AS completed_assessments,
        COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'in_progress') AS in_progress_assessments,
        COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'planned') AS planned_assessments,
        AVG(a.score) FILTER (WHERE a.status = 'completed') AS avg_assessment_score,
        MAX(a.end_date) FILTER (WHERE a.status = 'completed') AS last_assessment_date,
        SUM(CASE WHEN a.end_date < CURRENT_DATE - INTERVAL '1 year' AND a.status = 'completed' THEN 1 ELSE 0 END) AS assessments_older_than_1year
    FROM
        iso27001_assessments a
        JOIN organizations o ON a.organization_id = o.id
    GROUP BY
        a.organization_id, o.name, a.healthcare_specific
),
control_stats AS (
    SELECT
        a.organization_id,
        a.healthcare_specific,
        COUNT(cr.id) AS total_controls_assessed,
        SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END) AS compliant_controls,
        SUM(CASE WHEN cr.status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial_compliance_controls,
        SUM(CASE WHEN cr.status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant_controls,
        ROUND(
            (SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
            NULLIF(COUNT(cr.id) - SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END), 0)) * 100, 
            2
        ) AS overall_compliance_percentage
    FROM
        iso27001_assessments a
        JOIN iso27001_control_results cr ON a.id = cr.assessment_id
    WHERE
        a.status = 'completed'
    GROUP BY
        a.organization_id, a.healthcare_specific
),
action_plan_stats AS (
    SELECT
        ap.organization_id,
        a.healthcare_specific,
        COUNT(ap.id) AS total_action_plans,
        SUM(CASE WHEN ap.status IN ('open', 'in_progress') THEN 1 ELSE 0 END) AS open_action_plans,
        SUM(CASE WHEN ap.status = 'completed' THEN 1 ELSE 0 END) AS completed_action_plans,
        SUM(CASE WHEN ap.due_date < CURRENT_DATE AND ap.status IN ('open', 'in_progress') THEN 1 ELSE 0 END) AS overdue_action_plans,
        SUM(CASE WHEN ap.priority = 'critical' AND ap.status IN ('open', 'in_progress') THEN 1 ELSE 0 END) AS critical_open_actions,
        SUM(CASE WHEN ap.priority = 'high' AND ap.status IN ('open', 'in_progress') THEN 1 ELSE 0 END) AS high_open_actions
    FROM
        iso27001_action_plans ap
        LEFT JOIN iso27001_assessments a ON ap.assessment_id = a.id
    GROUP BY
        ap.organization_id, a.healthcare_specific
),
document_stats AS (
    SELECT
        d.organization_id,
        d.healthcare_specific,
        COUNT(d.id) AS total_documents,
        SUM(CASE WHEN d.status = 'published' THEN 1 ELSE 0 END) AS published_documents,
        SUM(CASE WHEN d.status = 'draft' THEN 1 ELSE 0 END) AS draft_documents,
        SUM(CASE WHEN d.next_review_date < CURRENT_DATE AND d.status = 'published' THEN 1 ELSE 0 END) AS documents_requiring_review
    FROM
        iso27001_documents d
    GROUP BY
        d.organization_id, d.healthcare_specific
)
SELECT
    as_stats.organization_id,
    as_stats.organization_name,
    as_stats.healthcare_specific,
    as_stats.total_assessments,
    as_stats.completed_assessments,
    as_stats.in_progress_assessments,
    as_stats.planned_assessments,
    as_stats.avg_assessment_score,
    as_stats.last_assessment_date,
    as_stats.assessments_older_than_1year,
    CASE 
        WHEN as_stats.last_assessment_date < CURRENT_DATE - INTERVAL '1 year' THEN 'Required'
        WHEN as_stats.last_assessment_date < CURRENT_DATE - INTERVAL '9 months' THEN 'Recommended'
        ELSE 'Up to date'
    END AS reassessment_status,
    
    c_stats.total_controls_assessed,
    c_stats.compliant_controls,
    c_stats.partial_compliance_controls,
    c_stats.non_compliant_controls,
    c_stats.overall_compliance_percentage,
    
    ap_stats.total_action_plans,
    ap_stats.open_action_plans,
    ap_stats.completed_action_plans,
    ap_stats.overdue_action_plans,
    ap_stats.critical_open_actions,
    ap_stats.high_open_actions,
    
    doc_stats.total_documents,
    doc_stats.published_documents,
    doc_stats.draft_documents,
    doc_stats.documents_requiring_review,
    
    CASE
        WHEN c_stats.overall_compliance_percentage >= 90 THEN 'High'
        WHEN c_stats.overall_compliance_percentage >= 75 THEN 'Medium'
        ELSE 'Low'
    END AS compliance_level,
    
    CASE
        WHEN ap_stats.critical_open_actions > 0 OR ap_stats.overdue_action_plans > 5 THEN 'High'
        WHEN ap_stats.high_open_actions > 3 OR ap_stats.overdue_action_plans > 0 THEN 'Medium'
        ELSE 'Low'
    END AS risk_level,
    
    CASE
        WHEN doc_stats.documents_requiring_review > 0 OR as_stats.assessments_older_than_1year > 0 THEN 'Action Required'
        WHEN as_stats.last_assessment_date < CURRENT_DATE - INTERVAL '9 months' THEN 'Review Recommended'
        ELSE 'Up to Date'
    END AS maintenance_status
FROM
    assessment_stats as_stats
    LEFT JOIN control_stats c_stats ON as_stats.organization_id = c_stats.organization_id AND as_stats.healthcare_specific = c_stats.healthcare_specific
    LEFT JOIN action_plan_stats ap_stats ON as_stats.organization_id = ap_stats.organization_id AND as_stats.healthcare_specific = ap_stats.healthcare_specific
    LEFT JOIN document_stats doc_stats ON as_stats.organization_id = doc_stats.organization_id AND as_stats.healthcare_specific = doc_stats.healthcare_specific
ORDER BY
    as_stats.organization_name, as_stats.healthcare_specific;

COMMENT ON VIEW vw_iso27001_compliance_dashboard IS 'Dashboard de compliance ISO 27001 com métricas de avaliação, controles, planos de ação e documentos';

-- View para áreas de melhoria em controles ISO 27001 para saúde
CREATE OR REPLACE VIEW vw_iso27001_healthcare_improvement_areas AS
WITH control_compliance AS (
    SELECT
        cr.control_id,
        c.control_id AS control_code,
        c.section,
        c.name,
        c.category,
        c.healthcare_applicability,
        COUNT(cr.id) AS total_assessments,
        SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END) AS compliant_count,
        SUM(CASE WHEN cr.status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial_compliance_count,
        SUM(CASE WHEN cr.status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant_count,
        ROUND(
            (SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
            NULLIF(COUNT(cr.id), 0)) * 100, 
            2
        ) AS compliance_percentage
    FROM
        iso27001_control_results cr
        JOIN iso27001_controls c ON cr.control_id = c.id
        JOIN iso27001_assessments a ON cr.assessment_id = a.id
    WHERE
        a.healthcare_specific = TRUE
    GROUP BY
        cr.control_id, c.control_id, c.section, c.name, c.category, c.healthcare_applicability
)
SELECT
    control_id,
    control_code,
    section,
    name,
    category,
    healthcare_applicability,
    total_assessments,
    compliant_count,
    partial_compliance_count,
    non_compliant_count,
    compliance_percentage,
    CASE
        WHEN compliance_percentage < 50 THEN 'Critical'
        WHEN compliance_percentage < 75 THEN 'High'
        WHEN compliance_percentage < 90 THEN 'Medium'
        ELSE 'Low'
    END AS improvement_priority,
    ARRAY(
        SELECT DISTINCT framework_code 
        FROM vw_iso27001_framework_mappings fm
        WHERE fm.iso_control_id = control_id
    ) AS related_frameworks
FROM
    control_compliance
WHERE
    healthcare_applicability IS NOT NULL
    AND total_assessments > 0
ORDER BY
    compliance_percentage ASC,
    total_assessments DESC;

COMMENT ON VIEW vw_iso27001_healthcare_improvement_areas IS 'Identifica áreas de melhoria em controles ISO 27001 relacionados à saúde';

-- View para resumo de compliance por categoria de controle ISO 27001
CREATE OR REPLACE VIEW vw_iso27001_category_compliance AS
WITH category_stats AS (
    SELECT
        c.category,
        a.healthcare_specific,
        COUNT(cr.id) AS total_controls,
        SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END) AS compliant_count,
        SUM(CASE WHEN cr.status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial_compliance_count,
        SUM(CASE WHEN cr.status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant_count,
        SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END) AS not_applicable_count,
        COUNT(DISTINCT a.organization_id) AS organizations_count,
        ROUND(
            (SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
            NULLIF(COUNT(cr.id) - SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END), 0)) * 100, 
            2
        ) AS compliance_percentage
    FROM
        iso27001_control_results cr
        JOIN iso27001_controls c ON cr.control_id = c.id
        JOIN iso27001_assessments a ON cr.assessment_id = a.id
    WHERE
        a.status = 'completed'
    GROUP BY
        c.category, a.healthcare_specific
)
SELECT
    category,
    healthcare_specific,
    total_controls,
    compliant_count,
    partial_compliance_count,
    non_compliant_count,
    not_applicable_count,
    organizations_count,
    compliance_percentage,
    CASE
        WHEN compliance_percentage >= 90 THEN 'Strong'
        WHEN compliance_percentage >= 75 THEN 'Moderate'
        WHEN compliance_percentage >= 50 THEN 'Weak'
        ELSE 'Critical'
    END AS compliance_strength,
    CASE
        WHEN compliance_percentage < 50 THEN 'High'
        WHEN compliance_percentage < 75 THEN 'Medium'
        ELSE 'Low'
    END AS risk_level
FROM
    category_stats
ORDER BY
    healthcare_specific DESC,
    compliance_percentage ASC;

COMMENT ON VIEW vw_iso27001_category_compliance IS 'Resumo de compliance por categoria de controle ISO 27001, separado por assessments gerais e de saúde';
