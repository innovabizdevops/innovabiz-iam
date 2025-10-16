-- Casos de Teste de Compatibilidade - Parte 2

-- 6. Teste de Compatibilidade com Dispositivos de Autenticação
SELECT test.register_test_case(
    'Teste de Compatibilidade com Dispositivos de Autenticação',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes dispositivos de autenticação',
    true
) as test_id;

SELECT test.run_test(
    6,
    'test_auth_device_compatibility',
    '{
        "test_type": "auth_device_compatibility",
        "test_cases": [
            {
                "scenario": "smartcard",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest"],
                "features": ["pkcs11", "pki", "certificates"]
            },
            {
                "scenario": "hardware_token",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest"],
                "features": ["otp", "time_based", "event_based"]
            },
            {
                "scenario": "biometric_reader",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest"],
                "features": ["fingerprint", "face", "iris"]
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

-- 7. Teste de Compatibilidade com Serviços de Cloud
SELECT test.register_test_case(
    'Teste de Compatibilidade com Serviços de Cloud',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes serviços de cloud',
    true
) as test_id;

SELECT test.run_test(
    7,
    'test_cloud_service_compatibility',
    '{
        "test_type": "cloud_service_compatibility",
        "test_cases": [
            {
                "scenario": "aws",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "services": ["cognito", "iam", "sso"],
                "features": ["federation", "sso", "mfa"]
            },
            {
                "scenario": "azure",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "services": ["aad", "b2c", "b2b"],
                "features": ["federation", "sso", "mfa"]
            },
            {
                "scenario": "gcp",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "services": ["iam", "cloud_identity", "sso"],
                "features": ["federation", "sso", "mfa"]
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

-- 8. Teste de Compatibilidade com Frameworks de Autenticação
SELECT test.register_test_case(
    'Teste de Compatibilidade com Frameworks de Autenticação',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes frameworks de autenticação',
    true
) as test_id;

SELECT test.run_test(
    8,
    'test_auth_framework_compatibility',
    '{
        "test_type": "auth_framework_compatibility",
        "test_cases": [
            {
                "scenario": "keycloak",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest"],
                "features": ["sso", "mfa", "federation"]
            },
            {
                "scenario": "auth0",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest"],
                "features": ["sso", "mfa", "rules"]
            },
            {
                "scenario": "okta",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest"],
                "features": ["sso", "mfa", "workflow"]
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

-- 9. Teste de Compatibilidade com Banco de Dados
SELECT test.register_test_case(
    'Teste de Compatibilidade com Banco de Dados',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes bancos de dados',
    true
) as test_id;

SELECT test.run_test(
    9,
    'test_database_compatibility',
    '{
        "test_type": "database_compatibility",
        "test_cases": [
            {
                "scenario": "postgresql",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest", "previous"],
                "features": ["encryption", "replication", "partitioning"]
            },
            {
                "scenario": "mysql",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest", "previous"],
                "features": ["encryption", "replication", "partitioning"]
            },
            {
                "scenario": "mongodb",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest", "previous"],
                "features": ["encryption", "replication", "sharding"]
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

-- 10. Teste de Compatibilidade com Caches
SELECT test.register_test_case(
    'Teste de Compatibilidade com Caches',
    'COMPATIBILITY',
    'Verifica a compatibilidade com diferentes sistemas de cache',
    true
) as test_id;

SELECT test.run_test(
    10,
    'test_cache_compatibility',
    '{
        "test_type": "cache_compatibility",
        "test_cases": [
            {
                "scenario": "redis",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "versions": ["latest", "previous"],
                "features": ["pubsub", "persistence", "replication"]
            },
            {
                "scenario": "memcached",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "versions": ["latest", "previous"],
                "features": ["distributions", "evictions", "cas"]
            },
            {
                "scenario": "infinispan",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "versions": ["latest", "previous"],
                "features": ["clustering", "transactions", "caching"]
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
