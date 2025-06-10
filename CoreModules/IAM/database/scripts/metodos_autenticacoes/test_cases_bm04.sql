-- Casos de Teste para Métodos Biométricos (BM-04)

-- 1. Teste de Impressão Digital
SELECT test.register_test_case(
    'Teste de Impressão Digital',
    'BM-04',
    'Verifica validação de impressão digital',
    true
) as test_id;

SELECT test.run_test(
    1,
    'auth.verify_fingerprint',
    '{
        "fingerprint_data": "base64_fingerprint",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 2. Teste de Reconhecimento Facial
SELECT test.register_test_case(
    'Teste de Reconhecimento Facial',
    'BM-04',
    'Verifica validação de reconhecimento facial',
    true
) as test_id;

SELECT test.run_test(
    2,
    'auth.verify_face_recognition',
    '{
        "face_data": "base64_face",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 3. Teste de Reconhecimento de Íris
SELECT test.register_test_case(
    'Teste de Reconhecimento de Íris',
    'BM-04',
    'Verifica validação de reconhecimento de íris',
    true
) as test_id;

SELECT test.run_test(
    3,
    'auth.verify_iris_recognition',
    '{
        "iris_data": "base64_iris",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 4. Teste de Reconhecimento de Voz
SELECT test.register_test_case(
    'Teste de Reconhecimento de Voz',
    'BM-04',
    'Verifica validação de reconhecimento de voz',
    true
) as test_id;

SELECT test.run_test(
    4,
    'auth.verify_voice_recognition',
    '{
        "voice_data": "base64_voice",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 5. Teste de Escaneamento de Retina
SELECT test.register_test_case(
    'Teste de Escaneamento de Retina',
    'BM-04',
    'Verifica validação de escaneamento de retina',
    true
) as test_id;

SELECT test.run_test(
    5,
    'auth.verify_retina_scan',
    '{
        "retina_data": "base64_retina",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 6. Teste de Reconhecimento Vascular
SELECT test.register_test_case(
    'Teste de Reconhecimento Vascular',
    'BM-04',
    'Verifica validação de reconhecimento vascular',
    true
) as test_id;

SELECT test.run_test(
    6,
    'auth.verify_vascular_recognition',
    '{
        "vascular_data": "base64_vascular",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 7. Teste de Geometria da Mão
SELECT test.register_test_case(
    'Teste de Geometria da Mão',
    'BM-04',
    'Verifica validação de geometria da mão',
    true
) as test_id;

SELECT test.run_test(
    7,
    'auth.verify_hand_geometry',
    '{
        "hand_data": "base64_hand",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 8. Teste de Dinâmica de Assinatura
SELECT test.register_test_case(
    'Teste de Dinâmica de Assinatura',
    'BM-04',
    'Verifica validação de dinâmica de assinatura',
    true
) as test_id;

SELECT test.run_test(
    8,
    'auth.verify_signature_dynamics',
    '{
        "signature_data": "base64_signature",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 9. Teste de Batimento Cardíaco
SELECT test.register_test_case(
    'Teste de Batimento Cardíaco',
    'BM-04',
    'Verifica validação de batimento cardíaco',
    true
) as test_id;

SELECT test.run_test(
    9,
    'auth.verify_heart_rate',
    '{
        "heart_rate_data": "base64_heart_rate",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 10. Teste de Reconhecimento de Marcha
SELECT test.register_test_case(
    'Teste de Reconhecimento de Marcha',
    'BM-04',
    'Verifica validação de reconhecimento de marcha',
    true
) as test_id;

SELECT test.run_test(
    10,
    'auth.verify_gait_recognition',
    '{
        "gait_data": "base64_gait",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 11. Teste de EEG (Eletroencefalograma)
SELECT test.register_test_case(
    'Teste de EEG',
    'BM-04',
    'Verifica validação de EEG',
    true
) as test_id;

SELECT test.run_test(
    11,
    'auth.verify_eeg',
    '{
        "eeg_data": "base64_eeg",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 12. Teste de Análise de DNA Rápida
SELECT test.register_test_case(
    'Teste de Análise de DNA',
    'BM-04',
    'Verifica validação de análise de DNA',
    true
) as test_id;

SELECT test.run_test(
    12,
    'auth.verify_dna_analysis',
    '{
        "dna_data": "base64_dna",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 13. Teste de Reconhecimento de Orelha
SELECT test.register_test_case(
    'Teste de Reconhecimento de Orelha',
    'BM-04',
    'Verifica validação de reconhecimento de orelha',
    true
) as test_id;

SELECT test.run_test(
    13,
    'auth.verify_ear_recognition',
    '{
        "ear_data": "base64_ear",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 14. Teste de Leitura Térmica Facial
SELECT test.register_test_case(
    'Teste de Leitura Térmica',
    'BM-04',
    'Verifica validação de leitura térmica facial',
    true
) as test_id;

SELECT test.run_test(
    14,
    'auth.verify_thermal_face',
    '{
        "thermal_data": "base64_thermal",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 15. Teste de Leitura de Impressão Palmar
SELECT test.register_test_case(
    'Teste de Leitura de Palma',
    'BM-04',
    'Verifica validação de leitura de impressão palmar',
    true
) as test_id;

SELECT test.run_test(
    15,
    'auth.verify_palm_print',
    '{
        "palm_data": "base64_palm",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 16. Teste de Multiespectral (Combinação de Biometrias)
SELECT test.register_test_case(
    'Teste de Multiespectral',
    'BM-04',
    'Verifica validação de combinação de biometrias',
    true
) as test_id;

SELECT test.run_test(
    16,
    'auth.verify_multispectral',
    '{
        "biometrics": {
            "fingerprint": "base64_fingerprint",
            "face": "base64_face",
            "voice": "base64_voice"
        },
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 17. Teste de Reconhecimento Facial 3D
SELECT test.register_test_case(
    'Teste de Reconhecimento Facial 3D',
    'BM-04',
    'Verifica validação de reconhecimento facial 3D',
    true
) as test_id;

SELECT test.run_test(
    17,
    'auth.verify_3d_face_recognition',
    '{
        "face_3d_data": "base64_face_3d",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 18. Teste de Reconhecimento Labial
SELECT test.register_test_case(
    'Teste de Reconhecimento Labial',
    'BM-04',
    'Verifica validação de reconhecimento labial',
    true
) as test_id;

SELECT test.run_test(
    18,
    'auth.verify_lip_recognition',
    '{
        "lip_data": "base64_lip",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 19. Teste de Odor Corporal
SELECT test.register_test_case(
    'Teste de Odor Corporal',
    'BM-04',
    'Verifica validação de odor corporal',
    true
) as test_id;

SELECT test.run_test(
    19,
    'auth.verify_body_odor',
    '{
        "odor_data": "base64_odor",
        "user_id": "user123",
        "threshold": 0.95,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- 20. Teste de Pulsação Vascular
SELECT test.register_test_case(
    'Teste de Pulsação Vascular',
    'BM-04',
    'Verifica validação de pulsação vascular',
    true
) as test_id;

SELECT test.run_test(
    20,
    'auth.verify_vascular_pulse',
    '{
        "pulse_data": "base64_pulse",
        "user_id": "user123",
        "threshold": 0.9,
        "encryption": true,
        "liveness_check": true
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
