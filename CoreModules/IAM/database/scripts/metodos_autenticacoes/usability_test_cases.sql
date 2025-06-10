-- Casos de Teste de Usabilidade

-- 1. Teste de Usabilidade da Interface de Login
SELECT test.register_test_case(
    'Teste de Usabilidade da Interface de Login',
    'USABILITY',
    'Verifica a facilidade de uso da interface de login',
    true
) as test_id;

SELECT test.run_test(
    1,
    'test_usability_login_interface',
    '{
        "test_type": "login_interface",
        "test_cases": [
            {
                "scenario": "login_with_password",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "error_messages": true,
                "help_text": true
            },
            {
                "scenario": "forgot_password",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "help_text": true,
                "email_confirmation": true
            },
            {
                "scenario": "two_factor_auth",
                "expected_time_seconds": 15,
                "success_rate": 0.97,
                "help_text": true,
                "otp_delivery": true
            }
        ],
        "validation_metrics": {
            "completion_rate": 0.99,
            "error_rate": 0.01,
            "user_satisfaction": 0.95,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 2. Teste de Usabilidade da Autenticação Biométrica
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação Biométrica',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação biométrica',
    true
) as test_id;

SELECT test.run_test(
    2,
    'test_usability_biometric_auth',
    '{
        "test_type": "biometric_auth",
        "test_cases": [
            {
                "scenario": "fingerprint",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "error_messages": true,
                "fallback_options": true
            },
            {
                "scenario": "face_recognition",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "camera_guidance": true,
                "retry_options": true
            },
            {
                "scenario": "voice_recognition",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "audio_feedback": true,
                "retry_limit": 3
            }
        ],
        "validation_metrics": {
            "user_frustration_rate": 0.01,
            "abandonment_rate": 0.02,
            "success_rate": 0.99,
            "time_to_complete_seconds": 3
        }
    }'::jsonb
);

-- 3. Teste de Usabilidade da Autenticação por Token
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Token',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por token',
    true
) as test_id;

SELECT test.run_test(
    3,
    'test_usability_token_auth',
    '{
        "test_type": "token_auth",
        "test_cases": [
            {
                "scenario": "physical_token",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "software_token",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "qr_code_scanning": true,
                "backup_codes": true
            },
            {
                "scenario": "push_notification",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "notification_clarity": true,
                "retry_options": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "token_recovery_rate": 0.99
        }
    }'::jsonb
);

-- 4. Teste de Usabilidade da Autenticação por Dispositivo
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Dispositivo',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo',
    true
) as test_id;

SELECT test.run_test(
    4,
    'test_usability_device_auth',
    '{
        "test_type": "device_auth",
        "test_cases": [
            {
                "scenario": "mobile_app",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "smart_card",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "card_reader_guidance": true,
                "error_messages": true
            },
            {
                "scenario": "hardware_token",
                "expected_time_seconds": 8,
                "success_rate": 0.97,
                "device_setup": true,
                "backup_options": true
            }
        ],
        "validation_metrics": {
            "user_frustration_rate": 0.01,
            "abandonment_rate": 0.02,
            "success_rate": 0.99,
            "time_to_complete_seconds": 6
        }
    }'::jsonb
);

-- 5. Teste de Usabilidade da Autenticação Contínua
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação Contínua',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação contínua',
    true
) as test_id;

SELECT test.run_test(
    5,
    'test_usability_continuous_auth',
    '{
        "test_type": "continuous_auth",
        "test_cases": [
            {
                "scenario": "behavior_analysis",
                "expected_time_seconds": 0,
                "success_rate": 0.99,
                "feedback_frequency": "low",
                "user_notification": true
            },
            {
                "scenario": "location_tracking",
                "expected_time_seconds": 0,
                "success_rate": 0.98,
                "privacy_controls": true,
                "user_opt_out": true
            },
            {
                "scenario": "device_monitoring",
                "expected_time_seconds": 0,
                "success_rate": 0.97,
                "user_controls": true,
                "notification_settings": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "privacy_compliance": 1.0,
            "notification_rate": 0.99
        }
    }'::jsonb
);

-- 6. Teste de Usabilidade da Recuperação de Senha
SELECT test.register_test_case(
    'Teste de Usabilidade da Recuperação de Senha',
    'USABILITY',
    'Verifica a facilidade de uso do processo de recuperação de senha',
    true
) as test_id;

SELECT test.run_test(
    6,
    'test_usability_password_recovery',
    '{
        "test_type": "password_recovery",
        "test_cases": [
            {
                "scenario": "email_recovery",
                "expected_time_seconds": 30,
                "success_rate": 0.99,
                "help_text": true,
                "progress_indicator": true
            },
            {
                "scenario": "security_questions",
                "expected_time_seconds": 45,
                "success_rate": 0.98,
                "question_clarity": true,
                "retry_options": true
            },
            {
                "scenario": "backup_code",
                "expected_time_seconds": 20,
                "success_rate": 0.97,
                "code_format": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 30,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 7. Teste de Usabilidade da Autenticação Multifatorial
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação Multifatorial',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação multifatorial',
    true
) as test_id;

SELECT test.run_test(
    7,
    'test_usability_mfa',
    '{
        "test_type": "mfa",
        "test_cases": [
            {
                "scenario": "password_and_token",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "password_and_biometric",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "token_and_biometric",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 8. Teste de Usabilidade da Autenticação por Reconhecimento de Padrões
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Padrões',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por reconhecimento de padrões',
    true
) as test_id;

SELECT test.run_test(
    8,
    'test_usability_pattern_recognition',
    '{
        "test_type": "pattern_recognition",
        "test_cases": [
            {
                "scenario": "typing_pattern",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "mouse_movement",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "touch_pattern",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 9. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por reconhecimento de dispositivo',
    true
) as test_id;

SELECT test.run_test(
    9,
    'test_usability_device_recognition',
    '{
        "test_type": "device_recognition",
        "test_cases": [
            {
                "scenario": "device_fingerprinting",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "device_trust_score",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "device_verification",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 3,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 10. Teste de Usabilidade da Autenticação por Reconhecimento de Comportamento
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Comportamento',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por reconhecimento de comportamento',
    true
) as test_id;

SELECT test.run_test(
    10,
    'test_usability_behavior_recognition',
    '{
        "test_type": "behavior_recognition",
        "test_cases": [
            {
                "scenario": "activity_patterns",
                "expected_time_seconds": 0,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "usage_patterns",
                "expected_time_seconds": 0,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "interaction_patterns",
                "expected_time_seconds": 0,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 0,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 11. Teste de Usabilidade da Autenticação por Reconhecimento de Ambiente
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Ambiente',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por reconhecimento de ambiente',
    true
) as test_id;

SELECT test.run_test(
    11,
    'test_usability_environment_recognition',
    '{
        "test_type": "environment_recognition",
        "test_cases": [
            {
                "scenario": "location_based",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "time_based",
                "expected_time_seconds": 1,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "context_based",
                "expected_time_seconds": 3,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 2,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 12. Teste de Usabilidade da Autenticação por Reconhecimento de Biometria Avançada
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Biometria Avançada',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por biometria avançada',
    true
) as test_id;

SELECT test.run_test(
    12,
    'test_usability_advanced_biometrics',
    '{
        "test_type": "advanced_biometrics",
        "test_cases": [
            {
                "scenario": "iris_recognition",
                "expected_time_seconds": 3,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "vein_pattern",
                "expected_time_seconds": 4,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "pulse_recognition",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 13. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Móvel
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Móvel',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo móvel',
    true
) as test_id;

SELECT test.run_test(
    13,
    'test_usability_mobile_device_recognition',
    '{
        "test_type": "mobile_device_recognition",
        "test_cases": [
            {
                "scenario": "mobile_fingerprint",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "mobile_face_id",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "mobile_voice_recognition",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 3,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 14. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo de Desktop
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo de Desktop',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo de desktop',
    true
) as test_id;

SELECT test.run_test(
    14,
    'test_usability_desktop_device_recognition',
    '{
        "test_type": "desktop_device_recognition",
        "test_cases": [
            {
                "scenario": "desktop_fingerprint",
                "expected_time_seconds": 3,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "desktop_face_id",
                "expected_time_seconds": 4,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "desktop_voice_recognition",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 15. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo IoT
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo IoT',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo IoT',
    true
) as test_id;

SELECT test.run_test(
    15,
    'test_usability_iot_device_recognition',
    '{
        "test_type": "iot_device_recognition",
        "test_cases": [
            {
                "scenario": "iot_fingerprint",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "iot_face_id",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "iot_voice_recognition",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 3,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 16. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo VR/AR
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo VR/AR',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo VR/AR',
    true
) as test_id;

SELECT test.run_test(
    16,
    'test_usability_vr_ar_device_recognition',
    '{
        "test_type": "vr_ar_device_recognition",
        "test_cases": [
            {
                "scenario": "vr_fingerprint",
                "expected_time_seconds": 4,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "vr_face_id",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "vr_voice_recognition",
                "expected_time_seconds": 6,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 5,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 17. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Wearable
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Wearable',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo wearable',
    true
) as test_id;

SELECT test.run_test(
    17,
    'test_usability_wearable_device_recognition',
    '{
        "test_type": "wearable_device_recognition",
        "test_cases": [
            {
                "scenario": "wearable_fingerprint",
                "expected_time_seconds": 2,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "wearable_face_id",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "wearable_voice_recognition",
                "expected_time_seconds": 4,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 3,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 18. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Híbrido
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Híbrido',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo híbrido',
    true
) as test_id;

SELECT test.run_test(
    18,
    'test_usability_hybrid_device_recognition',
    '{
        "test_type": "hybrid_device_recognition",
        "test_cases": [
            {
                "scenario": "hybrid_fingerprint",
                "expected_time_seconds": 3,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "hybrid_face_id",
                "expected_time_seconds": 4,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "hybrid_voice_recognition",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 19. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Virtual
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Virtual',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por dispositivo virtual',
    true
) as test_id;

SELECT test.run_test(
    19,
    'test_usability_virtual_device_recognition',
    '{
        "test_type": "virtual_device_recognition",
        "test_cases": [
            {
                "scenario": "virtual_fingerprint",
                "expected_time_seconds": 3,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "virtual_face_id",
                "expected_time_seconds": 4,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "virtual_voice_recognition",
                "expected_time_seconds": 5,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 4,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- 20. Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Múltiplo
SELECT test.register_test_case(
    'Teste de Usabilidade da Autenticação por Reconhecimento de Dispositivo Múltiplo',
    'USABILITY',
    'Verifica a facilidade de uso da autenticação por múltiplos dispositivos',
    true
) as test_id;

SELECT test.run_test(
    20,
    'test_usability_multi_device_recognition',
    '{
        "test_type": "multi_device_recognition",
        "test_cases": [
            {
                "scenario": "mobile_and_desktop",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "help_text": true,
                "error_handling": true
            },
            {
                "scenario": "mobile_and_iot",
                "expected_time_seconds": 6,
                "success_rate": 0.98,
                "guidance_text": true,
                "retry_options": true
            },
            {
                "scenario": "desktop_and_iot",
                "expected_time_seconds": 7,
                "success_rate": 0.97,
                "device_guidance": true,
                "error_messages": true
            }
        ],
        "validation_metrics": {
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 6,
            "success_rate": 0.99
        }
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
