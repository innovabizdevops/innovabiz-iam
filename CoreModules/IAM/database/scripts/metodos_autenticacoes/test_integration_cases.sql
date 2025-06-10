-- Casos de Teste de Integração entre Categorias

-- 1. Teste de Integração Conhecimento + Posse
SELECT test.register_test_case(
    'Teste de Integração Conhecimento + Posse',
    'INTEGRATION',
    'Verifica integração entre autenticação por conhecimento e posse',
    true
) as test_id;

SELECT test.run_test(
    1,
    'auth.verify_knowledge_and_possession',
    '{
        "knowledge": {
            "password": "password123",
            "otp": "123456"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 2. Teste de Integração Conhecimento + Biometria
SELECT test.register_test_case(
    'Teste de Integração Conhecimento + Biometria',
    'INTEGRATION',
    'Verifica integração entre autenticação por conhecimento e biometria',
    true
) as test_id;

SELECT test.run_test(
    2,
    'auth.verify_knowledge_and_biometrics',
    '{
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 3. Teste de Integração Posse + Biometria
SELECT test.register_test_case(
    'Teste de Integração Posse + Biometria',
    'INTEGRATION',
    'Verifica integração entre autenticação por posse e biometria',
    true
) as test_id;

SELECT test.run_test(
    3,
    'auth.verify_possession_and_biometrics',
    '{
        "possession": {
            "token_id": "token123",
            "device_id": "device123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "voice": "base64_voice"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 4. Teste de Integração Anti-Fraude + Conhecimento
SELECT test.register_test_case(
    'Teste de Integração Anti-Fraude + Conhecimento',
    'INTEGRATION',
    'Verifica integração entre anti-fraude e autenticação por conhecimento',
    true
) as test_id;

SELECT test.run_test(
    4,
    'auth.verify_fraud_and_knowledge',
    '{
        "fraud": {
            "behavior": "base64_behavior",
            "location": "base64_location"
        },
        "knowledge": {
            "password": "password123",
            "security_question": "base64_question"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 5. Teste de Integração Anti-Fraude + Posse
SELECT test.register_test_case(
    'Teste de Integração Anti-Fraude + Posse',
    'INTEGRATION',
    'Verifica integração entre anti-fraude e autenticação por posse',
    true
) as test_id;

SELECT test.run_test(
    5,
    'auth.verify_fraud_and_possession',
    '{
        "fraud": {
            "behavior": "base64_behavior",
            "device_check": "base64_device"
        },
        "possession": {
            "token_id": "token123",
            "device_id": "device123"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 6. Teste de Integração Anti-Fraude + Biometria
SELECT test.register_test_case(
    'Teste de Integração Anti-Fraude + Biometria',
    'INTEGRATION',
    'Verifica integração entre anti-fraude e biometria',
    true
) as test_id;

SELECT test.run_test(
    6,
    'auth.verify_fraud_and_biometrics',
    '{
        "fraud": {
            "behavior": "base64_behavior",
            "device_check": "base64_device"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 7. Teste de Integração Dispositivos/Tokens + Conhecimento
SELECT test.register_test_case(
    'Teste de Integração Dispositivos/Tokens + Conhecimento',
    'INTEGRATION',
    'Verifica integração entre dispositivos/tokens e autenticação por conhecimento',
    true
) as test_id;

SELECT test.run_test(
    7,
    'auth.verify_tokens_and_knowledge',
    '{
        "tokens": {
            "token_id": "token123",
            "certificate": "base64_certificate"
        },
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 8. Teste de Integração Dispositivos/Tokens + Posse
SELECT test.register_test_case(
    'Teste de Integração Dispositivos/Tokens + Posse',
    'INTEGRATION',
    'Verifica integração entre dispositivos/tokens e autenticação por posse',
    true
) as test_id;

SELECT test.run_test(
    8,
    'auth.verify_tokens_and_possession',
    '{
        "tokens": {
            "token_id": "token123",
            "certificate": "base64_certificate"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 9. Teste de Integração Dispositivos/Tokens + Biometria
SELECT test.register_test_case(
    'Teste de Integração Dispositivos/Tokens + Biometria',
    'INTEGRATION',
    'Verifica integração entre dispositivos/tokens e biometria',
    true
) as test_id;

SELECT test.run_test(
    9,
    'auth.verify_tokens_and_biometrics',
    '{
        "tokens": {
            "token_id": "token123",
            "certificate": "base64_certificate"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 10. Teste de Integração Multi-Factor (KB + PB + BM)
SELECT test.register_test_case(
    'Teste de Integração Multi-Factor',
    'INTEGRATION',
    'Verifica integração de multi-factor authentication',
    true
) as test_id;

SELECT test.run_test(
    10,
    'auth.verify_multi_factor',
    '{
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 11. Teste de Integração com Anti-Fraude Completo
SELECT test.register_test_case(
    'Teste de Integração com Anti-Fraude',
    'INTEGRATION',
    'Verifica integração completa com anti-fraude',
    true
) as test_id;

SELECT test.run_test(
    11,
    'auth.verify_full_fraud_protection',
    '{
        "fraud": {
            "behavior": "base64_behavior",
            "device_check": "base64_device",
            "location": "base64_location"
        },
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 12. Teste de Integração com Dispositivos/Tokens Completo
SELECT test.register_test_case(
    'Teste de Integração com Dispositivos/Tokens',
    'INTEGRATION',
    'Verifica integração completa com dispositivos/tokens',
    true
) as test_id;

SELECT test.run_test(
    12,
    'auth.verify_full_token_integration',
    '{
        "tokens": {
            "token_id": "token123",
            "certificate": "base64_certificate",
            "token_type": "HOTP"
        },
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 13. Teste de Integração com Cenário de Recuperação
SELECT test.register_test_case(
    'Teste de Integração com Recuperação',
    'INTEGRATION',
    'Verifica integração em cenário de recuperação',
    true
) as test_id;

SELECT test.run_test(
    13,
    'auth.verify_recovery_integration',
    '{
        "recovery": {
            "backup_code": "backup123",
            "security_questions": ["question1", "question2"]
        },
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 14. Teste de Integração com Cenário de MFA Adaptativo
SELECT test.register_test_case(
    'Teste de Integração com MFA Adaptativo',
    'INTEGRATION',
    'Verifica integração com MFA adaptativo',
    true
) as test_id;

SELECT test.run_test(
    14,
    'auth.verify_adaptive_mfa',
    '{
        "risky_behavior": true,
        "location": "new_location",
        "device": "new_device",
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- 15. Teste de Integração com Cenário de Autenticação Contínua
SELECT test.register_test_case(
    'Teste de Integração com Autenticação Contínua',
    'INTEGRATION',
    'Verifica integração com autenticação contínua',
    true
) as test_id;

SELECT test.run_test(
    15,
    'auth.verify_continuous_auth',
    '{
        "continuous": {
            "behavior": "base64_behavior",
            "location": "base64_location",
            "device": "base64_device"
        },
        "knowledge": {
            "password": "password123",
            "pin": "1234"
        },
        "possession": {
            "device_id": "device123",
            "token_id": "token123"
        },
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face"
        },
        "user_id": "user123",
        "encryption": true
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
