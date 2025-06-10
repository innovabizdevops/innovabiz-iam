-- Casos de Teste para Métodos de Dispositivos/Tokens (DT-05)

-- 1. Teste de Cartão Inteligente
SELECT test.register_test_case(
    'Teste de Cartão Inteligente',
    'DT-05',
    'Verifica validação de cartão inteligente',
    true
) as test_id;

SELECT test.run_test(
    1,
    'auth.verify_smart_card',
    '{
        "card_id": "card123",
        "serial_number": "123456",
        "card_type": "EMV",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 2. Teste de Token Físico
SELECT test.register_test_case(
    'Teste de Token Físico',
    'DT-05',
    'Verifica validação de token físico',
    true
) as test_id;

SELECT test.run_test(
    2,
    'auth.verify_physical_token',
    '{
        "token_id": "token123",
        "serial_number": "123456",
        "token_type": "HOTP",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 3. Teste de Token Virtual
SELECT test.register_test_case(
    'Teste de Token Virtual',
    'DT-05',
    'Verifica validação de token virtual',
    true
) as test_id;

SELECT test.run_test(
    3,
    'auth.verify_virtual_token',
    '{
        "token_id": "token123",
        "device_id": "device123",
        "token_type": "TOTP",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 4. Teste de Certificado Digital
SELECT test.register_test_case(
    'Teste de Certificado Digital',
    'DT-05',
    'Verifica validação de certificado digital',
    true
) as test_id;

SELECT test.run_test(
    4,
    'auth.verify_digital_certificate',
    '{
        "certificate": "base64_certificate",
        "issuer": "issuer",
        "subject": "subject",
        "expiry_date": "2025-12-31",
        "encryption_key": "key456"
    }'::jsonb
);

-- 5. Teste de Cartão OTP
SELECT test.register_test_case(
    'Teste de Cartão OTP',
    'DT-05',
    'Verifica validação de cartão OTP',
    true
) as test_id;

SELECT test.run_test(
    5,
    'auth.verify_otp_card',
    '{
        "card_id": "card123",
        "otp": "123456",
        "expiry_time": "2025-05-16T00:05:00Z",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 6. Teste de Token de Segurança
SELECT test.register_test_case(
    'Teste de Token de Segurança',
    'DT-05',
    'Verifica validação de token de segurança',
    true
) as test_id;

SELECT test.run_test(
    6,
    'auth.verify_security_token',
    '{
        "token_id": "token123",
        "token_type": "HSM",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 7. Teste de Token de Hardware
SELECT test.register_test_case(
    'Teste de Token de Hardware',
    'DT-05',
    'Verifica validação de token de hardware',
    true
) as test_id;

SELECT test.run_test(
    7,
    'auth.verify_hardware_token',
    '{
        "token_id": "token123",
        "serial_number": "123456",
        "token_type": "USB",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 8. Teste de Token de Software
SELECT test.register_test_case(
    'Teste de Token de Software',
    'DT-05',
    'Verifica validação de token de software',
    true
) as test_id;

SELECT test.run_test(
    8,
    'auth.verify_software_token',
    '{
        "token_id": "token123",
        "token_type": "APP",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 9. Teste de Token de USB
SELECT test.register_test_case(
    'Teste de Token de USB',
    'DT-05',
    'Verifica validação de token USB',
    true
) as test_id;

SELECT test.run_test(
    9,
    'auth.verify_usb_token',
    '{
        "token_id": "token123",
        "serial_number": "123456",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 10. Teste de Token de Smartwatch
SELECT test.register_test_case(
    'Teste de Token de Smartwatch',
    'DT-05',
    'Verifica validação de token smartwatch',
    true
) as test_id;

SELECT test.run_test(
    10,
    'auth.verify_watch_token',
    '{
        "token_id": "token123",
        "device_id": "watch123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 11. Teste de Token de Smartglasses
SELECT test.register_test_case(
    'Teste de Token de Smartglasses',
    'DT-05',
    'Verifica validação de token smartglasses',
    true
) as test_id;

SELECT test.run_test(
    11,
    'auth.verify_glasses_token',
    '{
        "token_id": "token123",
        "device_id": "glasses123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 12. Teste de Token de Smart Ring
SELECT test.register_test_case(
    'Teste de Token de Smart Ring',
    'DT-05',
    'Verifica validação de token smart ring',
    true
) as test_id;

SELECT test.run_test(
    12,
    'auth.verify_ring_token',
    '{
        "token_id": "token123",
        "device_id": "ring123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 13. Teste de Token de Smart Tag
SELECT test.register_test_case(
    'Teste de Token de Smart Tag',
    'DT-05',
    'Verifica validação de token smart tag',
    true
) as test_id;

SELECT test.run_test(
    13,
    'auth.verify_tag_token',
    '{
        "token_id": "token123",
        "device_id": "tag123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 14. Teste de Token de Smart Pen
SELECT test.register_test_case(
    'Teste de Token de Smart Pen',
    'DT-05',
    'Verifica validação de token smart pen',
    true
) as test_id;

SELECT test.run_test(
    14,
    'auth.verify_pen_token',
    '{
        "token_id": "token123",
        "device_id": "pen123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 15. Teste de Token de Smart Key
SELECT test.register_test_case(
    'Teste de Token de Smart Key',
    'DT-05',
    'Verifica validação de token smart key',
    true
) as test_id;

SELECT test.run_test(
    15,
    'auth.verify_key_token',
    '{
        "token_id": "token123",
        "device_id": "key123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 16. Teste de Token de Smart Badge
SELECT test.register_test_case(
    'Teste de Token de Smart Badge',
    'DT-05',
    'Verifica validação de token smart badge',
    true
) as test_id;

SELECT test.run_test(
    16,
    'auth.verify_badge_token',
    '{
        "token_id": "token123",
        "device_id": "badge123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 17. Teste de Token de Smart Card
SELECT test.register_test_case(
    'Teste de Token de Smart Card',
    'DT-05',
    'Verifica validação de token smart card',
    true
) as test_id;

SELECT test.run_test(
    17,
    'auth.verify_card_token',
    '{
        "token_id": "token123",
        "device_id": "card123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 18. Teste de Token de Smart Device
SELECT test.register_test_case(
    'Teste de Token de Smart Device',
    'DT-05',
    'Verifica validação de token smart device',
    true
) as test_id;

SELECT test.run_test(
    18,
    'auth.verify_device_token',
    '{
        "token_id": "token123",
        "device_id": "device123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 19. Teste de Token de Smart Accessory
SELECT test.register_test_case(
    'Teste de Token de Smart Accessory',
    'DT-05',
    'Verifica validação de token smart accessory',
    true
) as test_id;

SELECT test.run_test(
    19,
    'auth.verify_accessory_token',
    '{
        "token_id": "token123",
        "device_id": "accessory123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- 20. Teste de Token de Smart Wearable
SELECT test.register_test_case(
    'Teste de Token de Smart Wearable',
    'DT-05',
    'Verifica validação de token smart wearable',
    true
) as test_id;

SELECT test.run_test(
    20,
    'auth.verify_wearable_token',
    '{
        "token_id": "token123",
        "device_id": "wearable123",
        "encryption_key": "key456",
        "certificate": "base64_certificate"
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
