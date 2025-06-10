-- Casos de Teste para Métodos Baseados em Conhecimento (KB-01)

-- 1. Teste de Senha Tradicional
SELECT test.register_test_case('Teste de Senha Tradicional', 'KB-01', 'Verifica validação de senha com requisitos básicos', true) as test_id;

SELECT test.run_test(
    1,
    'auth.verify_password',
    '{"password": "P@ssw0rd", "min_length": 8, "max_length": 32, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special": true}'::jsonb
);

-- 2. Teste de PIN Numérico
SELECT test.register_test_case('Teste de PIN Numérico', 'KB-01', 'Verifica validação de PIN com 6 dígitos', true) as test_id;

SELECT test.run_test(
    2,
    'auth.verify_pin',
    '{"pin": "123456", "length": 6, "numeric_only": true, "replay_protection": true}'::jsonb
);

-- 3. Teste de Padrão Gráfico
SELECT test.register_test_case('Teste de Padrão Gráfico', 'KB-01', 'Verifica validação de padrão gráfico com 4 pontos', true) as test_id;

SELECT test.run_test(
    3,
    'auth.verify_graphic_pattern',
    '{"pattern": [1, 2, 3, 6], "min_points": 4, "max_points": 9, "replay_protection": true}'::jsonb
);

-- 4. Teste de Perguntas de Segurança
SELECT test.register_test_case('Teste de Perguntas de Segurança', 'KB-01', 'Verifica validação de resposta à pergunta de segurança', true) as test_id;

SELECT test.run_test(
    4,
    'auth.verify_security_question',
    '{"question": "Nome do seu primeiro animal de estimação?", "answer": "Rex", "case_sensitive": false, "fuzzy_match": true}'::jsonb
);

-- 5. Teste de OTP (One-Time Password)
SELECT test.register_test_case('Teste de OTP', 'KB-01', 'Verifica validação de senha única de uso único', true) as test_id;

SELECT test.run_test(
    5,
    'auth.verify_otp',
    '{"otp": "123456", "secret": "base32secret3232", "algorithm": "SHA1", "digits": 6, "window": 1}'::jsonb
);

-- 6. Teste de Passphrase
SELECT test.register_test_case('Teste de Passphrase', 'KB-01', 'Verifica validação de frase de senha complexa', true) as test_id;

SELECT test.run_test(
    6,
    'auth.verify_passphrase',
    '{"passphrase": "This is my secure passphrase 123", "min_words": 5, "min_length": 20, "require_special": true, "entropy_check": true}'::jsonb
);

-- 7. Teste de Imagem Secreta
SELECT test.register_test_case('Teste de Imagem Secreta', 'KB-01', 'Verifica validação de imagem secreta para anti-phishing', true) as test_id;

SELECT test.run_test(
    7,
    'auth.verify_secret_image',
    '{"image_hash": "sha256_hash", "user_id": "user123", "session_id": "session456", "timestamp": "2025-05-16T00:00:00Z"}'::jsonb
);

-- 8. Teste de Senha de Uso Único
SELECT test.register_test_case('Teste de Senha de Uso Único', 'KB-01', 'Verifica validação de senha temporária para acesso emergencial', true) as test_id;

SELECT test.run_test(
    8,
    'auth.verify_single_use_password',
    '{"password": "temp123", "valid_until": "2025-05-16T01:00:00Z", "user_id": "user123", "single_use": true}'::jsonb
);

-- 9. Teste de Senhas Sem Conexão
SELECT test.register_test_case('Teste de Senhas Sem Conexão', 'KB-01', 'Verifica validação de senha em ambiente offline', true) as test_id;

SELECT test.run_test(
    9,
    'auth.verify_offline_password',
    '{"password": "offline123", "encryption_key": "key456", "local_storage": true, "encryption_type": "AES-256"}'::jsonb
);

-- 10. Teste de Gestos Customizados
SELECT test.register_test_case('Teste de Gestos Customizados', 'KB-01', 'Verifica validação de padrão de gestos personalizado', true) as test_id;

SELECT test.run_test(
    10,
    'auth.verify_custom_gesture',
    '{"gesture_data": "base64_encoded_gesture", "user_id": "user123", "threshold": 0.85, "encryption": true}'::jsonb
);

-- 11. Teste de Sequência de Ações
SELECT test.register_test_case('Teste de Sequência de Ações', 'KB-01', 'Verifica validação de sequência de ações personalizada', true) as test_id;

SELECT test.run_test(
    11,
    'auth.verify_action_sequence',
    '{"sequence": ["swipe_left", "tap", "zoom_in"], "user_id": "user123", "threshold": 0.9, "time_window": 5000}'::jsonb
);

-- 12. Teste de Localização em Imagem
SELECT test.register_test_case('Teste de Localização em Imagem', 'KB-01', 'Verifica validação de localização em imagem', true) as test_id;

SELECT test.run_test(
    12,
    'auth.verify_image_location',
    '{"image_hash": "sha256_hash", "coordinates": [100, 200], "threshold": 5, "encryption": true}'::jsonb
);

-- 13. Teste de PIN Expandido
SELECT test.register_test_case('Teste de PIN Expandido', 'KB-01', 'Verifica validação de PIN com múltiplos campos', true) as test_id;

SELECT test.run_test(
    13,
    'auth.verify_expanded_pin',
    '{"pin_parts": ["1234", "5678", "9012"], "user_id": "user123", "encryption": true, "replay_protection": true}'::jsonb
);

-- 14. Teste de Rotação de Caracteres
SELECT test.register_test_case('Teste de Rotação de Caracteres', 'KB-01', 'Verifica validação de rotação de caracteres', true) as test_id;

SELECT test.run_test(
    14,
    'auth.verify_character_rotation',
    '{"original": "password123", "rotated": "123password", "user_id": "user123", "encryption": true}'::jsonb
);

-- 15. Teste de Teclado Virtual Randomizado
SELECT test.register_test_case('Teste de Teclado Virtual Randomizado', 'KB-01', 'Verifica validação de teclado virtual com layout randomizado', true) as test_id;

SELECT test.run_test(
    15,
    'auth.verify_random_keyboard',
    '{"keyboard_layout": "random_layout_123", "password": "random123", "user_id": "user123", "encryption": true}'::jsonb
);

-- 16. Teste de Matriz de Autenticação
SELECT test.register_test_case('Teste de Matriz de Autenticação', 'KB-01', 'Verifica validação de matriz de autenticação', true) as test_id;

SELECT test.run_test(
    16,
    'auth.verify_auth_matrix',
    '{"matrix_data": "base64_matrix", "coordinates": [2, 3], "user_id": "user123", "encryption": true}'::jsonb
);

-- 17. Teste de Desafio-Resposta Baseado em Dados
SELECT test.register_test_case('Teste de Desafio-Resposta Baseado em Dados', 'KB-01', 'Verifica validação de desafio-resposta com dados', true) as test_id;

SELECT test.run_test(
    17,
    'auth.verify_data_challenge',
    '{"challenge": "balance_123", "response": "1000.00", "user_id": "user123", "encryption": true}'::jsonb
);

-- 18. Teste de Senha Dividida Multi-canal
SELECT test.register_test_case('Teste de Senha Dividida Multi-canal', 'KB-01', 'Verifica validação de senha dividida em múltiplos canais', true) as test_id;

SELECT test.run_test(
    18,
    'auth.verify_multi_channel_password',
    '{"channel_parts": {"email": "part1", "sms": "part2", "app": "part3"}, "user_id": "user123", "encryption": true}'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
