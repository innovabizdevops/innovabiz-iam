-- Testes de Verificação de Autenticação Baseada em IA e Aprendizado de Máquina

-- 1. Teste de Detecção de Anomalias
SELECT ai.verify_anomaly_detection(
    'user123',
    '{"login_frequency": 0.8, "activity_pattern": 0.9, "transaction_pattern": 0.7}'::jsonb,
    0.6
) AS anomaly_detection_test;

-- 2. Teste de Classificação de Risco
SELECT ai.verify_risk_classification(
    'user123',
    '{"location_change": 0.2, "device_change": 0.1, "behavior_change": 0.3}'::jsonb,
    0.7
) AS risk_classification_test;

-- 3. Teste de Análise de Comportamento
SELECT ai.verify_behavior_analysis(
    'user123',
    '{"login_pattern": 0.9, "activity_pattern": 0.8, "transaction_pattern": 0.7}'::jsonb,
    0.65
) AS behavior_analysis_test;

-- 4. Teste de Detecção de Intenção Fraudulenta
SELECT ai.verify_fraud_intent(
    'user123',
    '{"transaction_amount": 0.5, "frequency": 0.3, "pattern": 0.4}'::jsonb,
    0.8
) AS fraud_intent_test;

-- 5. Teste de Análise em Tempo Real
SELECT ai.verify_realtime_behavior(
    'user123',
    '{"activity_count": 50, "error_count": 2, "time_spent": 300}'::jsonb,
    interval '5 minutes'
) AS realtime_behavior_test;

-- 6. Teste de Detecção de Deepfake
SELECT ai.verify_deepfake_detection(
    'media_data_123',
    0.95
) AS deepfake_detection_test;

-- 7. Teste de Análise de Discurso
SELECT ai.verify_speech_analysis(
    'speech_data_123',
    'pt-BR',
    0.85
) AS speech_analysis_test;

-- 8. Teste de Análise de Texto
SELECT ai.verify_text_analysis(
    'text_data_123',
    'pt-BR',
    0.8
) AS text_analysis_test;

-- 9. Teste de Análise de Imagem
SELECT ai.verify_image_analysis(
    'image_data_123',
    'photo',
    0.9
) AS image_analysis_test;

-- 10. Teste de Análise de Vídeo
SELECT ai.verify_video_analysis(
    'video_data_123',
    interval '10 minutes',
    0.85
) AS video_analysis_test;

-- 11. Teste de Análise Multicanal
SELECT ai.verify_multichannel_behavior(
    'user123',
    ARRAY[
        '{"activity_level": 0.8, "pattern_consistency": 0.9, "time_spent": 300}'::jsonb,
        '{"activity_level": 0.7, "pattern_consistency": 0.8, "time_spent": 250}'::jsonb
    ],
    0.75
) AS multichannel_behavior_test;

-- 12. Teste de Análise Temporal
SELECT ai.verify_temporal_behavior(
    'user123',
    '{"activity_trend": 0.8, "pattern_consistency": 0.9, "anomaly_score": 0.2}'::jsonb,
    interval '1 hour'
) AS temporal_behavior_test;

-- 13. Teste de Análise Contextual
SELECT ai.verify_contextual_behavior(
    'user123',
    '{"location_consistency": 0.9, "device_consistency": 0.8, "activity_consistency": 0.7}'::jsonb,
    0.7
) AS contextual_behavior_test;

-- 14. Teste de Análise em Rede
SELECT ai.verify_network_behavior(
    'user123',
    '{"connection_pattern": 0.8, "traffic_pattern": 0.7, "anomaly_score": 0.3}'::jsonb,
    0.75
) AS network_behavior_test;

-- 15. Teste de Análise de Grupo
SELECT ai.verify_group_behavior(
    'user123',
    '{"activity_level": 0.8, "pattern_consistency": 0.9, "anomaly_score": 0.2}'::jsonb,
    0.7
) AS group_behavior_test;

-- 16. Teste de Análise Temporal Avançada
SELECT ai.verify_advanced_temporal(
    'user123',
    '{"trend": 0.8, "seasonality": 0.7, "anomaly_score": 0.2}'::jsonb,
    ARRAY[interval '1 hour', interval '1 day'],
    0.75
) AS advanced_temporal_test;

-- 17. Teste de Análise Multidimensional
SELECT ai.verify_multidimensional_behavior(
    'user123',
    ARRAY[
        '{"activity_level": 0.8, "pattern_consistency": 0.9, "anomaly_score": 0.2}'::jsonb,
        '{"activity_level": 0.7, "pattern_consistency": 0.8, "anomaly_score": 0.3}'::jsonb,
        '{"activity_level": 0.9, "pattern_consistency": 0.7, "anomaly_score": 0.1}'::jsonb
    ],
    0.7
) AS multidimensional_behavior_test;

-- 18. Teste de Análise Adaptativa
SELECT ai.verify_adaptive_behavior(
    'user123',
    '{"adaptation_score": 0.8, "learning_progress": 0.7, "confidence_score": 0.9}'::jsonb,
    0.75
) AS adaptive_behavior_test;

-- 19. Teste de Análise Contextual Avançada
SELECT ai.verify_advanced_contextual(
    'user123',
    '{"location": 0.9, "device": 0.8, "activity": 0.7, "consistency_score": 0.8, "anomaly_score": 0.2}'::jsonb,
    ARRAY['location', 'device', 'activity'],
    0.7
) AS advanced_contextual_test;

-- 20. Teste de Análise Híbrida
SELECT ai.verify_hybrid_behavior(
    'user123',
    '{"method1": 0.8, "method2": 0.7, "method3": 0.9, "consistency_score": 0.85, "anomaly_score": 0.15}'::jsonb,
    ARRAY['method1', 'method2', 'method3'],
    0.75
) AS hybrid_behavior_test;
