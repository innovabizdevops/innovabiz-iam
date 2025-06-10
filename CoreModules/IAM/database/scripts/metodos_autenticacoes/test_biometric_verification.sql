-- Testes de Verificação de Autenticação Biométrica

-- 1. Teste de Impressão Digital
SELECT biometric.verify_fingerprint(
    'finger123',
    0.85,
    'device123'
) AS fingerprint_test;

-- 2. Teste de Reconhecimento Facial
SELECT biometric.verify_face_recognition(
    'face123',
    0.9,
    TRUE
) AS face_recognition_test;

-- 3. Teste de Reconhecimento de Íris
SELECT biometric.verify_iris_recognition(
    'iris123',
    0.95,
    0.9
) AS iris_recognition_test;

-- 4. Teste de Reconhecimento de Voz
SELECT biometric.verify_voice_recognition(
    'voice123',
    0.85,
    0.1
) AS voice_recognition_test;

-- 5. Teste de Escaneamento de Retina
SELECT biometric.verify_retina_scan(
    'retina123',
    0.95,
    0.85
) AS retina_scan_test;

-- 6. Teste de Reconhecimento Vascular
SELECT biometric.verify_vascular_recognition(
    'vascular123',
    0.9,
    0.8
) AS vascular_recognition_test;

-- 7. Teste de Geometria da Mão
SELECT biometric.verify_hand_geometry(
    'hand123',
    0.85,
    0.85
) AS hand_geometry_test;

-- 8. Teste de Dinâmica de Assinatura
SELECT biometric.verify_signature_dynamics(
    'signature123',
    0.85,
    'pressure123'
) AS signature_dynamics_test;

-- 9. Teste de Batimento Cardíaco
SELECT biometric.verify_heart_rate(
    'heart123',
    0.8,
    0.15
) AS heart_rate_test;

-- 10. Teste de Reconhecimento de Marcha
SELECT biometric.verify_gait_recognition(
    'gait123',
    0.8,
    0.75
) AS gait_recognition_test;

-- 11. Teste de EEG
SELECT biometric.verify_eeg(
    'eeg123',
    0.85,
    0.85
) AS eeg_test;

-- 12. Teste de Análise de DNA Rápida
SELECT biometric.verify_rapid_dna(
    'dna123',
    0.98,
    0.9
) AS rapid_dna_test;

-- 13. Teste de Reconhecimento de Orelha
SELECT biometric.verify_ear_recognition(
    'ear123',
    0.85,
    0.8
) AS ear_recognition_test;

-- 14. Teste de Leitura Térmica Facial
SELECT biometric.verify_thermal_face(
    'thermal123',
    0.85,
    36.5
) AS thermal_face_test;

-- 15. Teste de Leitura de Impressão Palmar
SELECT biometric.verify_palm_print(
    'palm123',
    0.85,
    0.85
) AS palm_print_test;

-- 16. Teste de Multiespectral
SELECT biometric.verify_multispectral(
    'multi123',
    0.85,
    3
) AS multispectral_test;

-- 17. Teste de Reconhecimento Facial 3D
SELECT biometric.verify_3d_face_recognition(
    '3dface123',
    0.9,
    0.85
) AS face_3d_test;

-- 18. Teste de Reconhecimento Labial
SELECT biometric.verify_lip_recognition(
    'lip123',
    0.85,
    0.75
) AS lip_recognition_test;

-- 19. Teste de Odor Corporal
SELECT biometric.verify_body_odor(
    'odor123',
    0.85,
    0.8
) AS body_odor_test;

-- 20. Teste de Pulsação Vascular
SELECT biometric.verify_vascular_pulse(
    'pulse123',
    0.85,
    0.85
) AS vascular_pulse_test;
