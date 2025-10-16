-- Casos de Teste de Compatibilidade - Parte 1

-- 1. Teste de Compatibilidade com Navegadores
SELECT test.register_test_case(
    'Teste de Compatibilidade com Navegadores',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes navegadores',
    true
) as test_id;

SELECT test.run_test(
    1,
    'test_browser_compatibility',
    '{
        "test_type": "browser_compatibility",
        "test_cases": [
            {
                "scenario": "chrome",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest", "latest-1", "latest-2"],
                "features": ["webauthn", "webcrypto", "biometrics"]
            },
            {
                "scenario": "firefox",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest", "latest-1", "latest-2"],
                "features": ["webauthn", "webcrypto", "biometrics"]
            },
            {
                "scenario": "safari",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest", "latest-1", "latest-2"],
                "features": ["webauthn", "webcrypto", "biometrics"]
            },
            {
                "scenario": "edge",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest", "latest-1", "latest-2"],
                "features": ["webauthn", "webcrypto", "biometrics"]
            }
        ],
        "validation_metrics": {
            "compatibility_score": 0.99,
            "error_rate": 0.01,
            "feature_support": 0.99,
            "performance_score": 0.95
        }
    }'::jsonb
);

-- 2. Teste de Compatibilidade com Dispositivos Móveis
SELECT test.register_test_case(
    'Teste de Compatibilidade com Dispositivos Móveis',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes dispositivos móveis',
    true
) as test_id;

SELECT test.run_test(
    2,
    'test_mobile_device_compatibility',
    '{
        "test_type": "mobile_device_compatibility",
        "test_cases": [
            {
                "scenario": "android",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["13", "12", "11"],
                "features": ["biometrics", "fingerprint", "face_id"]
            },
            {
                "scenario": "ios",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["16", "15", "14"],
                "features": ["biometrics", "touch_id", "face_id"]
            },
            {
                "scenario": "tablet",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest"],
                "features": ["biometrics", "fingerprint", "face_id"]
            }
        ],
        "validation_metrics": {
            "compatibility_score": 0.99,
            "error_rate": 0.01,
            "feature_support": 0.99,
            "performance_score": 0.95
        }
    }'::jsonb
);

-- 3. Teste de Compatibilidade com Sistemas Operacionais
SELECT test.register_test_case(
    'Teste de Compatibilidade com Sistemas Operacionais',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes sistemas operacionais',
    true
) as test_id;

SELECT test.run_test(
    3,
    'test_os_compatibility',
    '{
        "test_type": "os_compatibility",
        "test_cases": [
            {
                "scenario": "windows",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["11", "10", "8.1"],
                "features": ["biometrics", "fingerprint", "smartcard"]
            },
            {
                "scenario": "macos",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest", "previous"],
                "features": ["biometrics", "touch_id", "smartcard"]
            },
            {
                "scenario": "linux",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["ubuntu", "debian", "fedora"],
                "features": ["biometrics", "fingerprint", "smartcard"]
            }
        ],
        "validation_metrics": {
            "compatibility_score": 0.99,
            "error_rate": 0.01,
            "feature_support": 0.99,
            "performance_score": 0.95
        }
    }'::jsonb
);

-- 4. Teste de Compatibilidade com APIs
SELECT test.register_test_case(
    'Teste de Compatibilidade com APIs',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes APIs',
    true
) as test_id;

SELECT test.run_test(
    4,
    'test_api_compatibility',
    '{
        "test_type": "api_compatibility",
        "test_cases": [
            {
                "scenario": "oauth2",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["2.0", "1.0"],
                "features": ["token_exchange", "refresh_tokens", "scopes"]
            },
            {
                "scenario": "openid_connect",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["1.0"],
                "features": ["id_token", "userinfo_endpoint", "discovery"]
            },
            {
                "scenario": "saml",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["2.0", "1.1"],
                "features": ["assertions", "metadata", "bindings"]
            }
        ],
        "validation_metrics": {
            "compatibility_score": 0.99,
            "error_rate": 0.01,
            "feature_support": 0.99,
            "performance_score": 0.95
        }
    }'::jsonb
);

-- 5. Teste de Compatibilidade com Protocolos de Autenticação
SELECT test.register_test_case(
    'Teste de Compatibilidade com Protocolos de Autenticação',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes protocolos de autenticação',
    true
) as test_id;

SELECT test.run_test(
    5,
    'test_auth_protocol_compatibility',
    '{
        "test_type": "auth_protocol_compatibility",
        "test_cases": [
            {
                "scenario": "ldap",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["3.0", "2.0"],
                "features": ["bind", "search", "modify"]
            },
            {
                "scenario": "radius",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest"],
                "features": ["authentication", "authorization", "accounting"]
            },
            {
                "scenario": "kerberos",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest"],
                "features": ["ticket_granting", "service_tickets", "renewal"]
            }
        ],
        "validation_metrics": {
            "compatibility_score": 0.99,
            "error_rate": 0.01,
            "feature_support": 0.99,
            "performance_score": 0.95
        }
    }'::jsonb
);
