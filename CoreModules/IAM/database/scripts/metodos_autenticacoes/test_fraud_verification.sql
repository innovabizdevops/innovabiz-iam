-- Testes de Verificação de Autenticação Anti-Fraude e Comportamental

-- 1. Teste de Análise de Comportamento do Usuário
SELECT fraud.verify_user_behavior(
    'user123',
    '{"score": 0.3, "activity": "normal"}'::jsonb,
    0.5,
    interval '1 hour'
) AS user_behavior_test;

-- 2. Teste de Detecção de Bot/Automação
SELECT fraud.verify_bot_detection(
    'session123',
    '{"click_rate": 50, "scroll_rate": 25, "form_fill_time": 1.5, "score": 0.2}'::jsonb,
    0.5
) AS bot_detection_test;

-- 3. Teste de Análise de Padrão de Digitação
SELECT fraud.verify_typing_pattern(
    'user123',
    '{"key_press_time": 0.3, "key_release_time": 0.4, "key_hold_time": 0.5, "score": 0.2}'::jsonb,
    0.5
) AS typing_pattern_test;

-- 4. Teste de Posicionamento do Mouse
SELECT fraud.verify_mouse_position(
    'user123',
    '{"speed": 500, "acceleration": 2000, "jerk": 5000, "score": 0.2}'::jsonb,
    0.5
) AS mouse_position_test;

-- 5. Teste de Reconhecimento de Estilo de Escrita
SELECT fraud.verify_writing_style(
    'user123',
    '{"word_length": 5, "sentence_length": 10, "typing_speed": 0.5, "score": 0.2}'::jsonb,
    0.5
) AS writing_style_test;

-- 6. Teste de Gestos em Tela Touchscreen
SELECT fraud.verify_touch_gestures(
    'user123',
    '{"velocity": 500, "acceleration": 2000, "pressure": 0.5, "score": 0.2}'::jsonb,
    0.5
) AS touch_gestures_test;

-- 7. Teste de Padrão de Uso de Aplicativo
SELECT fraud.verify_app_usage(
    'user123',
    '{"session_duration": 120, "screen_transitions": 10, "error_rate": 0.1, "score": 0.2}'::jsonb,
    0.5
) AS app_usage_test;

-- 8. Teste de Padrão de Interação com Interface
SELECT fraud.verify_interface_interaction(
    'user123',
    '{"click_rate": 50, "scroll_rate": 25, "form_fill_time": 1.5, "score": 0.2}'::jsonb,
    0.5
) AS interface_interaction_test;

-- 9. Teste de Análise de Comunicação (Linguística)
SELECT fraud.verify_linguistic_analysis(
    'user123',
    '{"sentiment_score": 0.7, "formality_score": 0.8, "complexity_score": 0.8, "score": 0.2}'::jsonb,
    0.5
) AS linguistic_analysis_test;

-- 10. Teste de Análise de Navegação Web
SELECT fraud.verify_web_navigation(
    'user123',
    '{"page_load_time": 1.5, "click_through_rate": 50, "bounce_rate": 0.3, "score": 0.2}'::jsonb,
    0.5
) AS web_navigation_test;

-- 11. Teste de Detecção de Jailbreak/Root
SELECT fraud.verify_jailbreak_detection(
    'device123',
    '{"jailbreak_detected": false, "root_detected": false, "tampering_detected": false, "score": 0.2}'::jsonb,
    0.5
) AS jailbreak_detection_test;

-- 12. Teste de Análise de Toque (Pressão/Velocidade)
SELECT fraud.verify_touch_analysis(
    'user123',
    '{"pressure": 0.5, "velocity": 500, "acceleration": 2000, "score": 0.2}'::jsonb,
    0.5
) AS touch_analysis_test;

-- 13. Teste de Análise de Postura em Dispositivos
SELECT fraud.verify_device_posture(
    'user123',
    '{"tilt_angle": 15, "rotation_angle": 10, "shake_detected": false, "score": 0.2}'::jsonb,
    0.5
) AS device_posture_test;

-- 14. Teste de Detecção de Emuladores
SELECT fraud.verify_emulator_detection(
    'device123',
    '{"emulator_detected": false, "virtual_device": false, "hypervisor_detected": false, "score": 0.2}'::jsonb,
    0.5
) AS emulator_detection_test;

-- 15. Teste de Detecção de Ataques de Replay
SELECT fraud.verify_replay_attack(
    'session123',
    '{"event_sequence": "1,2,3,4,5", "original_sequence": "1,2,3,4,5", "timestamp_diff": 0.5, "signature_match": false, "score": 0.2}'::jsonb,
    0.5
) AS replay_attack_test;

-- 16. Teste de Análise Temporal de Transações
SELECT fraud.verify_transaction_timing(
    'user123',
    '{"time_between_transactions": 1.5, "transaction_frequency": 50, "time_of_day_anomaly": false, "score": 0.2}'::jsonb,
    0.5
) AS transaction_timing_test;

-- 17. Teste de Machine Learning Comportamental
SELECT fraud.verify_behavioral_ml(
    'user123',
    '{"ml_score": 0.2, "anomaly_score": 0.3, "confidence_score": 0.8, "prediction_score": 0.8, "score": 0.2}'::jsonb,
    0.5
) AS behavioral_ml_test;

-- 18. Teste de Análise de Cohort/Peer Group
SELECT fraud.verify_cohort_analysis(
    'user123',
    '{"deviation_score": 0.3, "similarity_score": 0.8, "outlier_score": 0.2, "score": 0.2}'::jsonb,
    0.5
) AS cohort_analysis_test;

-- 19. Teste de Detecção de Phishing Comportamental
SELECT fraud.verify_phishing_behavior(
    'user123',
    '{"suspicious_links": 0, "suspicious_domains": 0, "suspicious_attachments": 0, "score": 0.2}'::jsonb,
    0.5
) AS phishing_behavior_test;

-- 20. Teste de Análise de Intenção Fraudulenta
SELECT fraud.verify_fraud_intent(
    'user123',
    '{"risk_score": 0.3, "anomaly_score": 0.2, "pattern_match": false, "score": 0.2}'::jsonb,
    0.5
) AS fraud_intent_test;
