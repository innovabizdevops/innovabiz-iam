-- Testes de Verificação de Autenticação Baseada em Geolocalização e IoT

-- 1. Teste de Localização
SELECT geo.verify_location(
    -23.5505,
    -46.6333,
    '[{"type": "Polygon", "coordinates": [[[-46.65, -23.55], [-46.62, -23.55], [-46.62, -23.57], [-46.65, -23.57], [-46.65, -23.55]]]}]'::jsonb,
    0.001
) AS location_test;

-- 2. Teste de Movimento
SELECT geo.verify_movement(
    '{"latitude": -23.5505, "longitude": -46.6333, "timestamp": "2025-05-15T22:20:00"}'::jsonb,
    '{"latitude": -23.5506, "longitude": -46.6334, "timestamp": "2025-05-15T22:19:59"}'::jsonb,
    100.0,
    interval '1 second'
) AS movement_test;

-- 3. Teste de Zona de Confiabilidade
SELECT geo.verify_trust_zone(
    '{"latitude": -23.5505, "longitude": -46.6333}',
    'safe_zone',
    '{"type": "Polygon", "coordinates": [[[-46.65, -23.55], [-46.62, -23.55], [-46.62, -23.57], [-46.65, -23.57], [-46.65, -23.55]]]}'::jsonb
) AS trust_zone_test;

-- 4. Teste de Dispositivo IoT
SELECT iot.verify_device(
    'device_123',
    '{"status": "active", "type": "sensor", "location": {"latitude": -23.5505, "longitude": -46.6333}}'::jsonb,
    CURRENT_TIMESTAMP
) AS device_test;

-- 5. Teste de Sensores IoT
SELECT iot.verify_sensors(
    '{"temperature": 25.0, "humidity": 50.0, "pressure": 1013.0}'::jsonb,
    '{"temperature": 50.0, "humidity": 100.0, "pressure": 1100.0}'::jsonb,
    interval '1 minute'
) AS sensors_test;

-- 6. Teste de Comunicação IoT
SELECT iot.verify_communication(
    '{"type": "status", "data": {"status": "active"}}'::jsonb,
    'signature_123',
    CURRENT_TIMESTAMP
) AS communication_test;

-- 7. Teste de Estado do Dispositivo
SELECT iot.verify_device_state(
    'device_123',
    '{"state": "active", "timestamp": "2025-05-15T22:20:00"}'::jsonb,
    ARRAY['active', 'idle', 'maintenance']
) AS device_state_test;

-- 8. Teste de Comportamento IoT
SELECT iot.verify_behavior(
    'device_123',
    '{"activity": 0.8, "energy": 0.7, "temperature": 0.6}'::jsonb,
    '{"activity": 0.9, "energy": 0.8, "temperature": 0.7}'::jsonb
) AS behavior_test;

-- 9. Teste de Localização IoT
SELECT iot.verify_location(
    'device_123',
    '{"latitude": -23.5505, "longitude": -46.6333}',
    '[{"type": "Polygon", "coordinates": [[[-46.65, -23.55], [-46.62, -23.55], [-46.62, -23.57], [-46.65, -23.57], [-46.65, -23.55]]]}]'::jsonb
) AS iot_location_test;

-- 10. Teste de Comunicação em Rede IoT
SELECT iot.verify_network_communication(
    'device_123',
    '{"signal_strength": -50, "latency": 100, "packet_loss": 0.01}'::jsonb,
    '{"signal_strength": -100, "latency": 200, "packet_loss": 0.1}'::jsonb
) AS network_communication_test;

-- 11. Teste de Comportamento Geográfico
SELECT geo.verify_geographic_behavior(
    'user123',
    ARRAY[
        '{"latitude": -23.5505, "longitude": -46.6333, "timestamp": "2025-05-15T22:20:00"}'::jsonb,
        '{"latitude": -23.5506, "longitude": -46.6334, "timestamp": "2025-05-15T22:20:01"}'::jsonb
    ],
    '{"max_distance": 1000.0}'::jsonb
) AS geographic_behavior_test;

-- 12. Teste de Zona de Segurança
SELECT geo.verify_security_zone(
    '{"latitude": -23.5505, "longitude": -46.6333}',
    'security_zone',
    '{"type": "Polygon", "coordinates": [[[-46.65, -23.55], [-46.62, -23.55], [-46.62, -23.57], [-46.65, -23.57], [-46.65, -23.55]]]}'::jsonb,
    '{"max_distance": 1000.0}'::jsonb
) AS security_zone_test;

-- 13. Teste de Comportamento Temporal Geográfico
SELECT geo.verify_temporal_behavior(
    'user123',
    ARRAY[
        '{"latitude": -23.5505, "longitude": -46.6333, "timestamp": "2025-05-15T22:20:00"}'::jsonb,
        '{"latitude": -23.5506, "longitude": -46.6334, "timestamp": "2025-05-15T22:20:01"}'::jsonb
    ],
    ARRAY[interval '1 minute'],
    '{"max_time": 60.0}'::jsonb
) AS temporal_behavior_test;

-- 14. Teste de Comportamento de Dispositivo IoT
SELECT iot.verify_device_behavior(
    'device_123',
    ARRAY[
        '{"activity": 0.8, "energy": 0.7, "temperature": 0.6}'::jsonb,
        '{"activity": 0.85, "energy": 0.75, "temperature": 0.65}'::jsonb
    ],
    '{"activity": 0.9, "energy": 0.8, "temperature": 0.7}'::jsonb
) AS device_behavior_test;

-- 15. Teste de Comportamento em Rede IoT
SELECT iot.verify_network_behavior(
    'device_123',
    ARRAY[
        '{"signal_strength": -50, "latency": 100, "packet_loss": 0.01}'::jsonb,
        '{"signal_strength": -45, "latency": 90, "packet_loss": 0.005}'::jsonb
    ],
    '{"signal_strength": -100, "latency": 200, "packet_loss": 0.1}'::jsonb
) AS network_behavior_test;

-- 16. Teste de Comportamento Multicanal IoT
SELECT iot.verify_multichannel_behavior(
    'device_123',
    ARRAY[
        '{"channel1": 0.8, "channel2": 0.7, "channel3": 0.6}'::jsonb,
        '{"channel1": 0.85, "channel2": 0.75, "channel3": 0.65}'::jsonb
    ],
    '{"channel1": 0.9, "channel2": 0.8, "channel3": 0.7}'::jsonb
) AS multichannel_behavior_test;

-- 17. Teste de Comportamento Contextual IoT
SELECT iot.verify_contextual_behavior(
    'device_123',
    '{"context1": 0.8, "context2": 0.7, "context3": 0.6}'::jsonb,
    '{"context1": 0.9, "context2": 0.8, "context3": 0.7}'::jsonb
) AS contextual_behavior_test;

-- 18. Teste de Comportamento Adaptativo IoT
SELECT iot.verify_adaptive_behavior(
    'device_123',
    '{"activity": 0.8, "energy": 0.7, "temperature": 0.6}'::jsonb,
    0.85
) AS adaptive_behavior_test;

-- 19. Teste de Comportamento Híbrido IoT
SELECT iot.verify_hybrid_behavior(
    'device_123',
    ARRAY[
        '{"activity": 0.8, "energy": 0.7, "temperature": 0.6}'::jsonb,
        '{"activity": 0.85, "energy": 0.75, "temperature": 0.65}'::jsonb
    ],
    ARRAY['activity', 'energy', 'temperature'],
    '{"activity": 0.9, "energy": 0.8, "temperature": 0.7}'::jsonb
) AS hybrid_behavior_test;

-- 20. Teste de Comportamento em Rede Híbrida IoT
SELECT iot.verify_hybrid_network(
    'device_123',
    ARRAY[
        '{"signal_strength": -50, "latency": 100, "packet_loss": 0.01}'::jsonb,
        '{"signal_strength": -45, "latency": 90, "packet_loss": 0.005}'::jsonb
    ],
    ARRAY['signal_strength', 'latency', 'packet_loss'],
    '{"signal_strength": -100, "latency": 200, "packet_loss": 0.1}'::jsonb
) AS hybrid_network_test;
