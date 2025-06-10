-- Testes de Verificação de Autenticação Baseada em Conhecimento

-- 1. Teste de Senha Tradicional
SELECT knowledge.verify_traditional_password(
    'P@ssw0rd2025',
    12,
    'HIGH',
    current_timestamp
) AS traditional_password_test;

-- 2. Teste de PIN Numérico
SELECT knowledge.verify_numeric_pin(
    '123456',
    6,
    3,
    current_timestamp
) AS numeric_pin_test;

-- 3. Teste de Padrão Gráfico
SELECT knowledge.verify_graphic_pattern(
    '1,2,3,4,5,6',
    4,
    30
) AS graphic_pattern_test;

-- 4. Teste de Perguntas de Segurança
SELECT knowledge.verify_security_questions(
    'Qual é sua cor favorita?',
    'Azul',
    3,
    current_timestamp
) AS security_questions_test;

-- 5. Teste de Senha Única (OTP)
SELECT knowledge.verify_otp(
    '123456',
    current_timestamp + interval '30 seconds',
    'TOTP'
) AS otp_test;

-- 6. Teste de Verificação de Conhecimento
SELECT knowledge.verify_knowledge_verification(
    'Qual é sua cor favorita?',
    'Azul',
    30
) AS knowledge_verification_test;

-- 7. Teste de Passphrase
SELECT knowledge.verify_passphrase(
    'This is a strong passphrase with numbers 123',
    4,
    'HIGH'
) AS passphrase_test;

-- 8. Teste de Senha com Requisitos Complexos
SELECT knowledge.verify_complex_password(
    'P@ssw0rd2025',
    12,
    4,
    current_timestamp
) AS complex_password_test;

-- 9. Teste de Imagem Secreta
SELECT knowledge.verify_secret_image(
    'img123',
    4,
    4
) AS secret_image_test;

-- 10. Teste de Senha de Uso Único
SELECT knowledge.verify_single_use_password(
    '123456',
    current_timestamp + interval '5 minutes',
    'EMAIL'
) AS single_use_password_test;

-- 11. Teste de Senhas Sem Conexão
SELECT knowledge.verify_offline_password(
    'P@ssw0rd2025',
    12,
    'HIGH'
) AS offline_password_test;

-- 12. Teste de Gestos Customizados
SELECT knowledge.verify_custom_gesture(
    '1,2,3,4,5,6,7,8',
    4,
    30
) AS custom_gesture_test;

-- 13. Teste de Sequência de Ações
SELECT knowledge.verify_action_sequence(
    '1,2,3,4,5,6,7,8',
    4,
    30
) AS action_sequence_test;

-- 14. Teste de Localização em Imagem
SELECT knowledge.verify_image_location(
    'img123',
    4,
    4
) AS image_location_test;

-- 15. Teste de PIN Expandido
SELECT knowledge.verify_expanded_pin(
    '1234567890',
    8,
    3,
    current_timestamp
) AS expanded_pin_test;

-- 16. Teste de Rotação de Caracteres
SELECT knowledge.verify_character_rotation(
    'password',
    '123',
    3
) AS character_rotation_test;

-- 17. Teste de Teclado Virtual Randomizado
SELECT knowledge.verify_random_keyboard(
    'P@ssw0rd2025',
    12,
    'HIGH'
) AS random_keyboard_test;

-- 18. Teste de Matriz de Autenticação
SELECT knowledge.verify_authentication_matrix(
    'matrix123',
    2,
    2,
    30
) AS authentication_matrix_test;

-- 19. Teste de Desafio-Resposta Baseado em Dados
SELECT knowledge.verify_data_challenge(
    'Qual é sua cor favorita?',
    'Azul',
    30
) AS data_challenge_test;

-- 20. Teste de Senha Dividida Multi-canal
SELECT knowledge.verify_multi_channel_password(
    'EMAIL',
    'SMS',
    '123456',
    '654321',
    30
) AS multi_channel_password_test;
