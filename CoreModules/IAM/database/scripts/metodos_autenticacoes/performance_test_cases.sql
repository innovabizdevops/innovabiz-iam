-- Casos de Teste de Performance

-- 1. Teste de Latência de Autenticação
SELECT test.register_test_case(
    'Teste de Latência de Autenticação',
    'PERFORMANCE',
    'Verifica latência de processamento de autenticação',
    true
) as test_id;

SELECT test.run_test(
    1,
    'perf.test_authentication_latency',
    '{
        "test_type": "authentication",
        "method": "password",
        "concurrent_users": 100,
        "iterations": 1000,
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- 2. Teste de Throughput de Autenticação
SELECT test.register_test_case(
    'Teste de Throughput de Autenticação',
    'PERFORMANCE',
    'Verifica throughput de processamento de autenticação',
    true
) as test_id;

SELECT test.run_test(
    2,
    'perf.test_authentication_throughput',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "duration": "1 hour",
        "target_throughput": "1000 req/s",
        "encryption": true
    }'::jsonb
);

-- 3. Teste de Escalabilidade de Autenticação
SELECT test.register_test_case(
    'Teste de Escalabilidade de Autenticação',
    'PERFORMANCE',
    'Verifica escalabilidade do sistema de autenticação',
    true
) as test_id;

SELECT test.run_test(
    3,
    'perf.test_authentication_scalability',
    '{
        "test_type": "authentication",
        "method": "biometric",
        "start_users": 100,
        "end_users": 10000,
        "step_size": 1000,
        "duration": "1 hour",
        "target_latency": "500ms",
        "encryption": true
    }'::jsonb
);

-- 4. Teste de Resposta sob Carga
SELECT test.register_test_case(
    'Teste de Resposta sob Carga',
    'PERFORMANCE',
    'Verifica resposta do sistema sob carga pesada',
    true
) as test_id;

SELECT test.run_test(
    4,
    'perf.test_heavy_load_response',
    '{
        "test_type": "authentication",
        "method": "multi_factor",
        "concurrent_users": 5000,
        "duration": "2 hours",
        "target_latency": "1s",
        "encryption": true
    }'::jsonb
);

-- 5. Teste de Recuperação de Performance
SELECT test.register_test_case(
    'Teste de Recuperação de Performance',
    'PERFORMANCE',
    'Verifica recuperação do sistema após sobrecarga',
    true
) as test_id;

SELECT test.run_test(
    5,
    'perf.test_performance_recovery',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 10000,
        "duration": "1 hour",
        "recovery_time": "10 minutes",
        "target_latency": "200ms",
        "encryption": true
    }'::jsonb
);

-- 6. Teste de Performance com Cache
SELECT test.register_test_case(
    'Teste de Performance com Cache',
    'PERFORMANCE',
    'Verifica performance com cache ativo',
    true
) as test_id;

SELECT test.run_test(
    6,
    'perf.test_cache_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 5000,
        "cache_size": "1GB",
        "cache_ttl": "1 hour",
        "target_latency": "50ms",
        "encryption": true
    }'::jsonb
);

-- 7. Teste de Performance com Banco de Dados
SELECT test.register_test_case(
    'Teste de Performance com Banco de Dados',
    'PERFORMANCE',
    'Verifica performance com diferentes configurações de banco de dados',
    true
) as test_id;

SELECT test.run_test(
    7,
    'perf.test_database_performance',
    '{
        "test_type": "authentication",
        "method": "password",
        "concurrent_users": 1000,
        "db_config": {
            "pool_size": 100,
            "connection_timeout": "30s",
            "query_timeout": "5s"
        },
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- 8. Teste de Performance com Criptografia
SELECT test.register_test_case(
    'Teste de Performance com Criptografia',
    'PERFORMANCE',
    'Verifica impacto da criptografia na performance',
    true
) as test_id;

SELECT test.run_test(
    8,
    'perf.test_encryption_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "encryption_algorithms": ["AES", "RSA", "ECC"],
        "key_sizes": [128, 256, 512],
        "target_latency": "200ms"
    }'::jsonb
);

-- 9. Teste de Performance com Redundância
SELECT test.register_test_case(
    'Teste de Performance com Redundância',
    'PERFORMANCE',
    'Verifica performance com diferentes níveis de redundância',
    true
) as test_id;

SELECT test.run_test(
    9,
    'perf.test_redundancy_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "redundancy": {
            "replicas": 3,
            "load_balancing": "round_robin",
            "failover": "active_standby"
        },
        "target_latency": "150ms",
        "encryption": true
    }'::jsonb
);

-- 10. Teste de Performance com Cache Distribuído
SELECT test.register_test_case(
    'Teste de Performance com Cache Distribuído',
    'PERFORMANCE',
    'Verifica performance com cache distribuído',
    true
) as test_id;

SELECT test.run_test(
    10,
    'perf.test_distributed_cache_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 5000,
        "cache_config": {
            "nodes": 3,
            "replication_factor": 2,
            "partition_strategy": "consistent_hashing"
        },
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- 11. Teste de Performance com API Gateway
SELECT test.register_test_case(
    'Teste de Performance com API Gateway',
    'PERFORMANCE',
    'Verifica performance com API Gateway',
    true
) as test_id;

SELECT test.run_test(
    11,
    'perf.test_api_gateway_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "gateway_config": {
            "rate_limit": "1000 req/s",
            "circuit_breaker": true,
            "retry_strategy": "exponential_backoff"
        },
        "target_latency": "150ms",
        "encryption": true
    }'::jsonb
);

-- 12. Teste de Performance com Balanceamento de Carga
SELECT test.register_test_case(
    'Teste de Performance com Balanceamento de Carga',
    'PERFORMANCE',
    'Verifica performance com diferentes estratégias de balanceamento',
    true
) as test_id;

SELECT test.run_test(
    12,
    'perf.test_load_balancing_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 10000,
        "balancing_strategies": ["round_robin", "least_connections", "ip_hash"],
        "target_latency": "200ms",
        "encryption": true
    }'::jsonb
);

-- 13. Teste de Performance com Cache em Memória
SELECT test.register_test_case(
    'Teste de Performance com Cache em Memória',
    'PERFORMANCE',
    'Verifica performance com cache em memória',
    true
) as test_id;

SELECT test.run_test(
    13,
    'perf.test_memory_cache_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 5000,
        "cache_config": {
            "size": "1GB",
            "eviction_policy": "LRU",
            "max_age": "1 hour"
        },
        "target_latency": "50ms",
        "encryption": true
    }'::jsonb
);

-- 14. Teste de Performance com Banco de Dados Distribuído
SELECT test.register_test_case(
    'Teste de Performance com Banco de Dados Distribuído',
    'PERFORMANCE',
    'Verifica performance com banco de dados distribuído',
    true
) as test_id;

SELECT test.run_test(
    14,
    'perf.test_distributed_db_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 10000,
        "db_config": {
            "shards": 4,
            "replicas": 2,
            "partition_strategy": "range"
        },
        "target_latency": "200ms",
        "encryption": true
    }'::jsonb
);

-- 15. Teste de Performance com Criptografia Assimétrica
SELECT test.register_test_case(
    'Teste de Performance com Criptografia Assimétrica',
    'PERFORMANCE',
    'Verifica performance com criptografia assimétrica',
    true
) as test_id;

SELECT test.run_test(
    15,
    'perf.test_asymmetric_encryption_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "encryption_config": {
            "algorithm": "RSA",
            "key_size": 2048,
            "padding": "PKCS1"
        },
        "target_latency": "300ms",
        "encryption": true
    }'::jsonb
);

-- 16. Teste de Performance com Banco de Dados em Memória
SELECT test.register_test_case(
    'Teste de Performance com Banco de Dados em Memória',
    'PERFORMANCE',
    'Verifica performance com banco de dados em memória',
    true
) as test_id;

SELECT test.run_test(
    16,
    'perf.test_in_memory_db_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 5000,
        "db_config": {
            "size": "1GB",
            "persistence": "disk",
            "backup_interval": "1 hour"
        },
        "target_latency": "50ms",
        "encryption": true
    }'::jsonb
);

-- 17. Teste de Performance com Cache em Disco
SELECT test.register_test_case(
    'Teste de Performance com Cache em Disco',
    'PERFORMANCE',
    'Verifica performance com cache em disco',
    true
) as test_id;

SELECT test.run_test(
    17,
    'perf.test_disk_cache_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "cache_config": {
            "size": "10GB",
            "compression": true,
            "encryption": true
        },
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- 18. Teste de Performance com Banco de Dados em Nuvem
SELECT test.register_test_case(
    'Teste de Performance com Banco de Dados em Nuvem',
    'PERFORMANCE',
    'Verifica performance com banco de dados em nuvem',
    true
) as test_id;

SELECT test.run_test(
    18,
    'perf.test_cloud_db_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 10000,
        "db_config": {
            "provider": "AWS",
            "region": "us-east-1",
            "instance_type": "db.r5.large"
        },
        "target_latency": "200ms",
        "encryption": true
    }'::jsonb
);

-- 19. Teste de Performance com Criptografia Simétrica
SELECT test.register_test_case(
    'Teste de Performance com Criptografia Simétrica',
    'PERFORMANCE',
    'Verifica performance com criptografia simétrica',
    true
) as test_id;

SELECT test.run_test(
    19,
    'perf.test_symmetric_encryption_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 1000,
        "encryption_config": {
            "algorithm": "AES",
            "key_size": 256,
            "mode": "GCM"
        },
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- 20. Teste de Performance com Banco de Dados em Memória Distribuído
SELECT test.register_test_case(
    'Teste de Performance com Banco de Dados em Memória Distribuído',
    'PERFORMANCE',
    'Verifica performance com banco de dados em memória distribuído',
    true
) as test_id;

SELECT test.run_test(
    20,
    'perf.test_distributed_in_memory_db_performance',
    '{
        "test_type": "authentication",
        "method": "token",
        "concurrent_users": 10000,
        "db_config": {
            "nodes": 4,
            "replication_factor": 2,
            "partition_strategy": "consistent_hashing"
        },
        "target_latency": "100ms",
        "encryption": true
    }'::jsonb
);

-- Exemplo de execução completa
-- SELECT test.generate_test_report();
