-- Casos de Teste de Acessibilidade

-- 1. Teste de Acessibilidade para Usuários com Deficiência Visual
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Deficiência Visual',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com deficiência visual',
    true
) as test_id;

SELECT test.run_test(
    1,
    'test_accessibility_visual_impairment',
    '{
        "test_type": "visual_impairment",
        "test_cases": [
            {
                "scenario": "screen_reader_support",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "aria_labels": true,
                "keyboard_navigation": true
            },
            {
                "scenario": "high_contrast_mode",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "color_contrast_ratio": 4.5,
                "text_size_adjustment": true
            },
            {
                "scenario": "text_to_speech",
                "expected_time_seconds": 15,
                "success_rate": 0.97,
                "speech_rate": "normal",
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 2. Teste de Acessibilidade para Usuários com Deficiência Auditiva
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Deficiência Auditiva',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com deficiência auditiva',
    true
) as test_id;

SELECT test.run_test(
    2,
    'test_accessibility_hearing_impairment',
    '{
        "test_type": "hearing_impairment",
        "test_cases": [
            {
                "scenario": "closed_captions",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "caption_accuracy": 0.99,
                "timing_sync": true
            },
            {
                "scenario": "visual_alerts",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "alert_visibility": true,
                "color_coding": true
            },
            {
                "scenario": "text_transcription",
                "expected_time_seconds": 10,
                "success_rate": 0.97,
                "transcription_quality": 0.99,
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 6
        }
    }'::jsonb
);

-- 3. Teste de Acessibilidade para Usuários com Deficiência Física
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Deficiência Física',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com deficiência física',
    true
) as test_id;

SELECT test.run_test(
    3,
    'test_accessibility_physical_impairment',
    '{
        "test_type": "physical_impairment",
        "test_cases": [
            {
                "scenario": "keyboard_navigation",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "tab_order": true,
                "keyboard_shortcuts": true
            },
            {
                "scenario": "voice_commands",
                "expected_time_seconds": 15,
                "success_rate": 0.98,
                "command_accuracy": 0.99,
                "language_support": true
            },
            {
                "scenario": "adaptive_input",
                "expected_time_seconds": 20,
                "success_rate": 0.97,
                "input_methods": true,
                "customization_options": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 12
        }
    }'::jsonb
);

-- 4. Teste de Acessibilidade para Usuários com Deficiência Cognitiva
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Deficiência Cognitiva',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com deficiência cognitiva',
    true
) as test_id;

SELECT test.run_test(
    4,
    'test_accessibility_cognitive_impairment',
    '{
        "test_type": "cognitive_impairment",
        "test_cases": [
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "clear_navigation": true,
                "visual_clarity": true
            },
            {
                "scenario": "step_by_step_guidance",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "error_prevention",
                "expected_time_seconds": 15,
                "success_rate": 0.97,
                "error_messages": true,
                "recovery_options": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 5. Teste de Acessibilidade para Usuários com Deficiência Múltipla
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Deficiência Múltipla',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com múltiplas deficiências',
    true
) as test_id;

SELECT test.run_test(
    5,
    'test_accessibility_multiple_impairments',
    '{
        "test_type": "multiple_impairments",
        "test_cases": [
            {
                "scenario": "combined_visual_and_hearing",
                "expected_time_seconds": 20,
                "success_rate": 0.99,
                "multimodal_support": true,
                "adaptive_features": true
            },
            {
                "scenario": "combined_physical_and_cognitive",
                "expected_time_seconds": 25,
                "success_rate": 0.98,
                "customizable_interface": true,
                "assistive_technology": true
            },
            {
                "scenario": "combined_visual_and_physical",
                "expected_time_seconds": 30,
                "success_rate": 0.97,
                "alternative_input": true,
                "output_methods": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 22
        }
    }'::jsonb
);

-- 6. Teste de Acessibilidade para Idosos
SELECT test.register_test_case(
    'Teste de Acessibilidade para Idosos',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários idosos',
    true
) as test_id;

SELECT test.run_test(
    6,
    'test_accessibility_elderly',
    '{
        "test_type": "elderly",
        "test_cases": [
            {
                "scenario": "large_text_support",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "text_size": "large",
                "high_contrast": true
            },
            {
                "scenario": "simple_navigation",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "clear_labels": true,
                "intuitive_flow": true
            },
            {
                "scenario": "memory_assistance",
                "expected_time_seconds": 15,
                "success_rate": 0.97,
                "step_guidance": true,
                "progress_tracking": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 7. Teste de Acessibilidade para Usuários com Dificuldades de Leitura
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Leitura',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de leitura',
    true
) as test_id;

SELECT test.run_test(
    7,
    'test_accessibility_reading_difficulties',
    '{
        "test_type": "reading_difficulties",
        "test_cases": [
            {
                "scenario": "text_to_speech",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "speech_rate": "normal",
                "language_support": true
            },
            {
                "scenario": "simplified_text",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "easy_reading": true,
                "visual_support": true
            },
            {
                "scenario": "visual_guidance",
                "expected_time_seconds": 8,
                "success_rate": 0.97,
                "icons_support": true,
                "multimedia_alternatives": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 8. Teste de Acessibilidade para Usuários com Dificuldades de Escrita
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Escrita',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de escrita',
    true
) as test_id;

SELECT test.run_test(
    8,
    'test_accessibility_writing_difficulties',
    '{
        "test_type": "writing_difficulties",
        "test_cases": [
            {
                "scenario": "speech_to_text",
                "expected_time_seconds": 15,
                "success_rate": 0.99,
                "speech_recognition": true,
                "language_support": true
            },
            {
                "scenario": "voice_commands",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "command_accuracy": 0.99,
                "shortcut_support": true
            },
            {
                "scenario": "adaptive_input",
                "expected_time_seconds": 20,
                "success_rate": 0.97,
                "input_methods": true,
                "customization_options": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 12
        }
    }'::jsonb
);

-- 9. Teste de Acessibilidade para Usuários com Dificuldades de Memória
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Memória',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de memória',
    true
) as test_id;

SELECT test.run_test(
    9,
    'test_accessibility_memory_difficulties',
    '{
        "test_type": "memory_difficulties",
        "test_cases": [
            {
                "scenario": "step_by_step_guidance",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_reminders",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "contextual_help": true,
                "reminder_system": true
            },
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "clear_navigation": true,
                "visual_clarity": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 10. Teste de Acessibilidade para Usuários com Dificuldades de Atenção
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Atenção',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de atenção',
    true
) as test_id;

SELECT test.run_test(
    10,
    'test_accessibility_attention_difficulties',
    '{
        "test_type": "attention_difficulties",
        "test_cases": [
            {
                "scenario": "minimal_distractions",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "clean_interface": true,
                "focus_management": true
            },
            {
                "scenario": "step_by_step_process",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_cues",
                "expected_time_seconds": 8,
                "success_rate": 0.97,
                "contextual_help": true,
                "reminder_system": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 11. Teste de Acessibilidade para Usuários com Dificuldades de Comunicação
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Comunicação',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de comunicação',
    true
) as test_id;

SELECT test.run_test(
    11,
    'test_accessibility_communication_difficulties',
    '{
        "test_type": "communication_difficulties",
        "test_cases": [
            {
                "scenario": "alternative_input",
                "expected_time_seconds": 15,
                "success_rate": 0.99,
                "input_methods": true,
                "customization_options": true
            },
            {
                "scenario": "visual_communication",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "icons_support": true,
                "multimedia_alternatives": true
            },
            {
                "scenario": "text_to_speech",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "speech_rate": "normal",
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 12
        }
    }'::jsonb
);

-- 12. Teste de Acessibilidade para Usuários com Dificuldades de Mobilidade
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Mobilidade',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de mobilidade',
    true
) as test_id;

SELECT test.run_test(
    12,
    'test_accessibility_mobility_difficulties',
    '{
        "test_type": "mobility_difficulties",
        "test_cases": [
            {
                "scenario": "keyboard_navigation",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "tab_order": true,
                "keyboard_shortcuts": true
            },
            {
                "scenario": "voice_commands",
                "expected_time_seconds": 15,
                "success_rate": 0.98,
                "command_accuracy": 0.99,
                "language_support": true
            },
            {
                "scenario": "adaptive_input",
                "expected_time_seconds": 20,
                "success_rate": 0.97,
                "input_methods": true,
                "customization_options": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 12
        }
    }'::jsonb
);

-- 13. Teste de Acessibilidade para Usuários com Dificuldades de Processamento
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento',
    true
) as test_id;

SELECT test.run_test(
    13,
    'test_accessibility_processing_difficulties',
    '{
        "test_type": "processing_difficulties",
        "test_cases": [
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "clear_navigation": true,
                "visual_clarity": true
            },
            {
                "scenario": "step_by_step_guidance",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_support",
                "expected_time_seconds": 8,
                "success_rate": 0.97,
                "icons_support": true,
                "multimedia_alternatives": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 14. Teste de Acessibilidade para Usuários com Dificuldades de Memória de Curto Prazo
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Memória de Curto Prazo',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de memória de curto prazo',
    true
) as test_id;

SELECT test.run_test(
    14,
    'test_accessibility_short_term_memory',
    '{
        "test_type": "short_term_memory",
        "test_cases": [
            {
                "scenario": "step_by_step_process",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_reminders",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "contextual_help": true,
                "reminder_system": true
            },
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "clear_navigation": true,
                "visual_clarity": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 15. Teste de Acessibilidade para Usuários com Dificuldades de Memória de Longo Prazo
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Memória de Longo Prazo',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de memória de longo prazo',
    true
) as test_id;

SELECT test.run_test(
    15,
    'test_accessibility_long_term_memory',
    '{
        "test_type": "long_term_memory",
        "test_cases": [
            {
                "scenario": "step_by_step_guidance",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_reminders",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "contextual_help": true,
                "reminder_system": true
            },
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "clear_navigation": true,
                "visual_clarity": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 16. Teste de Acessibilidade para Usuários com Dificuldades de Processamento Auditivo
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento Auditivo',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento auditivo',
    true
) as test_id;

SELECT test.run_test(
    16,
    'test_accessibility_auditory_processing',
    '{
        "test_type": "auditory_processing",
        "test_cases": [
            {
                "scenario": "closed_captions",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "caption_accuracy": 0.99,
                "timing_sync": true
            },
            {
                "scenario": "visual_alerts",
                "expected_time_seconds": 3,
                "success_rate": 0.98,
                "alert_visibility": true,
                "color_coding": true
            },
            {
                "scenario": "text_transcription",
                "expected_time_seconds": 10,
                "success_rate": 0.97,
                "transcription_quality": 0.99,
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 6
        }
    }'::jsonb
);

-- 17. Teste de Acessibilidade para Usuários com Dificuldades de Processamento Visual
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento Visual',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento visual',
    true
) as test_id;

SELECT test.run_test(
    17,
    'test_accessibility_visual_processing',
    '{
        "test_type": "visual_processing",
        "test_cases": [
            {
                "scenario": "screen_reader_support",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "aria_labels": true,
                "keyboard_navigation": true
            },
            {
                "scenario": "high_contrast_mode",
                "expected_time_seconds": 5,
                "success_rate": 0.98,
                "color_contrast_ratio": 4.5,
                "text_size_adjustment": true
            },
            {
                "scenario": "text_to_speech",
                "expected_time_seconds": 15,
                "success_rate": 0.97,
                "speech_rate": "normal",
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 18. Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Linguagem
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Linguagem',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento de linguagem',
    true
) as test_id;

SELECT test.run_test(
    18,
    'test_accessibility_language_processing',
    '{
        "test_type": "language_processing",
        "test_cases": [
            {
                "scenario": "simplified_text",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "easy_reading": true,
                "visual_support": true
            },
            {
                "scenario": "visual_guidance",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "icons_support": true,
                "multimedia_alternatives": true
            },
            {
                "scenario": "text_to_speech",
                "expected_time_seconds": 10,
                "success_rate": 0.97,
                "speech_rate": "normal",
                "language_support": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- 19. Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Informação
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Informação',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento de informação',
    true
) as test_id;

SELECT test.run_test(
    19,
    'test_accessibility_information_processing',
    '{
        "test_type": "information_processing",
        "test_cases": [
            {
                "scenario": "step_by_step_guidance",
                "expected_time_seconds": 10,
                "success_rate": 0.99,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_reminders",
                "expected_time_seconds": 8,
                "success_rate": 0.98,
                "contextual_help": true,
                "reminder_system": true
            },
            {
                "scenario": "simplified_interface",
                "expected_time_seconds": 12,
                "success_rate": 0.97,
                "clear_navigation": true,
                "visual_clarity": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 10
        }
    }'::jsonb
);

-- 20. Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Emoções
SELECT test.register_test_case(
    'Teste de Acessibilidade para Usuários com Dificuldades de Processamento de Emoções',
    'ACCESSIBILITY',
    'Verifica a acessibilidade para usuários com dificuldades de processamento de emoções',
    true
) as test_id;

SELECT test.run_test(
    20,
    'test_accessibility_emotional_processing',
    '{
        "test_type": "emotional_processing",
        "test_cases": [
            {
                "scenario": "minimal_distractions",
                "expected_time_seconds": 5,
                "success_rate": 0.99,
                "clean_interface": true,
                "focus_management": true
            },
            {
                "scenario": "step_by_step_process",
                "expected_time_seconds": 10,
                "success_rate": 0.98,
                "progress_indicator": true,
                "help_text": true
            },
            {
                "scenario": "visual_cues",
                "expected_time_seconds": 8,
                "success_rate": 0.97,
                "contextual_help": true,
                "reminder_system": true
            }
        ],
        "validation_metrics": {
            "accessibility_score": 0.99,
            "user_satisfaction": 0.95,
            "error_rate": 0.01,
            "time_to_complete_seconds": 8
        }
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
