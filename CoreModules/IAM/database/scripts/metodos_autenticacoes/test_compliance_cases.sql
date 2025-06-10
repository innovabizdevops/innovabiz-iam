-- Casos de Teste de Conformidade com Regulamentações

-- 1. Teste de Conformidade com GDPR
SELECT test.register_test_case(
    'Teste de Conformidade com GDPR',
    'COMPLIANCE',
    'Verifica conformidade com General Data Protection Regulation (GDPR)',
    true
) as test_id;

SELECT test.run_test(
    1,
    'compliance.verify_gdpr',
    '{
        "data_processing": {
            "consent": true,
            "data_minimization": true,
            "storage_limitation": true,
            "data_protection": true
        },
        "encryption": {
            "at_rest": true,
            "in_transit": true,
            "key_management": true
        },
        "audit": {
            "logs": true,
            "access_control": true,
            "data_breach": true
        },
        "user_rights": {
            "access": true,
            "rectification": true,
            "erasure": true,
            "portability": true
        }
    }'::jsonb
);

-- 2. Teste de Conformidade com CCPA
SELECT test.register_test_case(
    'Teste de Conformidade com CCPA',
    'COMPLIANCE',
    'Verifica conformidade com California Consumer Privacy Act (CCPA)',
    true
) as test_id;

SELECT test.run_test(
    2,
    'compliance.verify_ccpa',
    '{
        "consumer_rights": {
            "access": true,
            "deletion": true,
            "opt_out": true,
            "non_discrimination": true
        },
        "data_collection": {
            "notification": true,
            "transparency": true,
            "minimization": true
        },
        "security": {
            "encryption": true,
            "access_control": true,
            "audit": true
        }
    }'::jsonb
);

-- 3. Teste de Conformidade com LGPD
SELECT test.register_test_case(
    'Teste de Conformidade com LGPD',
    'COMPLIANCE',
    'Verifica conformidade com Lei Geral de Proteção de Dados (LGPD)',
    true
) as test_id;

SELECT test.run_test(
    3,
    'compliance.verify_lgpd',
    '{
        "data_protection": {
            "consent": true,
            "transparency": true,
            "security": true,
            "accountability": true
        },
        "user_rights": {
            "access": true,
            "correction": true,
            "deletion": true,
            "portability": true
        },
        "security": {
            "encryption": true,
            "access_control": true,
            "audit": true
        }
    }'::jsonb
);

-- 4. Teste de Conformidade com PCI DSS
SELECT test.register_test_case(
    'Teste de Conformidade com PCI DSS',
    'COMPLIANCE',
    'Verifica conformidade com Payment Card Industry Data Security Standard (PCI DSS)',
    true
) as test_id;

SELECT test.run_test(
    4,
    'compliance.verify_pci_dss',
    '{
        "security": {
            "firewalls": true,
            "encryption": true,
            "access_control": true,
            "monitoring": true
        },
        "data_protection": {
            "storage": true,
            "transmission": true,
            "destruction": true
        },
        "operations": {
            "maintenance": true,
            "testing": true,
            "documentation": true
        }
    }'::jsonb
);

-- 5. Teste de Conformidade com HIPAA
SELECT test.register_test_case(
    'Teste de Conformidade com HIPAA',
    'COMPLIANCE',
    'Verifica conformidade com Health Insurance Portability and Accountability Act (HIPAA)',
    true
) as test_id;

SELECT test.run_test(
    5,
    'compliance.verify_hipaa',
    '{
        "security": {
            "administrative": true,
            "physical": true,
            "technical": true
        },
        "data_protection": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "privacy": {
            "rules": true,
            "procedures": true,
            "training": true
        }
    }'::jsonb
);

-- 6. Teste de Conformidade com ISO 27001
SELECT test.register_test_case(
    'Teste de Conformidade com ISO 27001',
    'COMPLIANCE',
    'Verifica conformidade com International Organization for Standardization (ISO 27001)',
    true
) as test_id;

SELECT test.run_test(
    6,
    'compliance.verify_iso_27001',
    '{
        "information_security": {
            "policy": true,
            "objectives": true,
            "risk_assessment": true,
            "controls": true
        },
        "documentation": {
            "procedures": true,
            "records": true,
            "reviews": true
        },
        "operations": {
            "monitoring": true,
            "measurement": true,
            "audit": true
        }
    }'::jsonb
);

-- 7. Teste de Conformidade com NIST SP 800-63B
SELECT test.register_test_case(
    'Teste de Conformidade com NIST SP 800-63B',
    'COMPLIANCE',
    'Verifica conformidade com National Institute of Standards and Technology (NIST SP 800-63B)',
    true
) as test_id;

SELECT test.run_test(
    7,
    'compliance.verify_nist_800_63b',
    '{
        "authentication": {
            "knowledge": true,
            "possession": true,
            "inherence": true
        },
        "security": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "assurance": {
            "levels": true,
            "verification": true,
            "validation": true
        }
    }'::jsonb
);

-- 8. Teste de Conformidade com SOC 2
SELECT test.register_test_case(
    'Teste de Conformidade com SOC 2',
    'COMPLIANCE',
    'Verifica conformidade com Service Organization Control (SOC 2)',
    true
) as test_id;

SELECT test.run_test(
    8,
    'compliance.verify_soc_2',
    '{
        "trust_services": {
            "security": true,
            "availability": true,
            "processing_integrity": true,
            "confidentiality": true,
            "privacy": true
        },
        "controls": {
            "design": true,
            "operation": true,
            "monitoring": true
        }
    }'::jsonb
);

-- 9. Teste de Conformidade com FERPA
SELECT test.register_test_case(
    'Teste de Conformidade com FERPA',
    'COMPLIANCE',
    'Verifica conformidade com Family Educational Rights and Privacy Act (FERPA)',
    true
) as test_id;

SELECT test.run_test(
    9,
    'compliance.verify_ferpa',
    '{
        "student_rights": {
            "access": true,
            "amendment": true,
            "disclosure": true
        },
        "data_protection": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "records": {
            "maintenance": true,
            "retention": true,
            "destruction": true
        }
    }'::jsonb
);

-- 10. Teste de Conformidade com COPPA
SELECT test.register_test_case(
    'Teste de Conformidade com COPPA',
    'COMPLIANCE',
    'Verifica conformidade com Children''s Online Privacy Protection Act (COPPA)',
    true
) as test_id;

SELECT test.run_test(
    10,
    'compliance.verify_coppa',
    '{
        "child_protection": {
            "verifiable_parental_consent": true,
            "data_collection": true,
            "disclosure": true
        },
        "security": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "privacy": {
            "notice": true,
            "parental_rights": true,
            "child_rights": true
        }
    }'::jsonb
);

-- 11. Teste de Conformidade com Solvência II
SELECT test.register_test_case(
    'Teste de Conformidade com Solvência II',
    'COMPLIANCE',
    'Verifica conformidade com Solvência II',
    true
) as test_id;

SELECT test.run_test(
    11,
    'compliance.verify_solvency_ii',
    '{
        "risk_management": {
            "framework": true,
            "controls": true,
            "monitoring": true
        },
        "data_protection": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "reporting": {
            "accuracy": true,
            "timeliness": true,
            "completeness": true
        }
    }'::jsonb
);

-- 12. Teste de Conformidade com Basel III
SELECT test.register_test_case(
    'Teste de Conformidade com Basel III',
    'COMPLIANCE',
    'Verifica conformidade com Basel III',
    true
) as test_id;

SELECT test.run_test(
    12,
    'compliance.verify_basel_iii',
    '{
        "risk_management": {
            "framework": true,
            "controls": true,
            "monitoring": true
        },
        "data_protection": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "reporting": {
            "accuracy": true,
            "timeliness": true,
            "completeness": true
        }
    }'::jsonb
);

-- 13. Teste de Conformidade com Open Banking
SELECT test.register_test_case(
    'Teste de Conformidade com Open Banking',
    'COMPLIANCE',
    'Verifica conformidade com Open Banking',
    true
) as test_id;

SELECT test.run_test(
    13,
    'compliance.verify_open_banking',
    '{
        "api_security": {
            "authentication": true,
            "authorization": true,
            "encryption": true
        },
        "data_protection": {
            "consent": true,
            "privacy": true,
            "audit": true
        },
        "operations": {
            "monitoring": true,
            "logging": true,
            "incident_response": true
        }
    }'::jsonb
);

-- 14. Teste de Conformidade com PSD2
SELECT test.register_test_case(
    'Teste de Conformidade com PSD2',
    'COMPLIANCE',
    'Verifica conformidade com Payment Services Directive 2 (PSD2)',
    true
) as test_id;

SELECT test.run_test(
    14,
    'compliance.verify_psd2',
    '{
        "strong_customer_auth": {
            "multi_factor": true,
            "risk_based": true,
            "exemptions": true
        },
        "api_security": {
            "authentication": true,
            "authorization": true,
            "encryption": true
        },
        "data_protection": {
            "consent": true,
            "privacy": true,
            "audit": true
        }
    }'::jsonb
);

-- 15. Teste de Conformidade com eIDAS
SELECT test.register_test_case(
    'Teste de Conformidade com eIDAS',
    'COMPLIANCE',
    'Verifica conformidade com Electronic Identification and Trust Services (eIDAS)',
    true
) as test_id;

SELECT test.run_test(
    15,
    'compliance.verify_eidas',
    '{
        "digital_id": {
            "authentication": true,
            "verification": true,
            "trust_services": true
        },
        "security": {
            "encryption": true,
            "access_control": true,
            "audit": true
        },
        "cross_border": {
            "recognition": true,
            "interoperability": true,
            "compliance": true
        }
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
