-- Sistema de Testes para Métodos de Autenticação

-- Configuração inicial do ambiente de testes
CREATE OR REPLACE FUNCTION test.setup_test_environment()
RETURNS VOID AS $$
BEGIN
    -- Limpar dados de teste existentes
    DELETE FROM test_results;
    DELETE FROM test_cases;
    
    -- Criar tabelas de suporte
    CREATE TABLE IF NOT EXISTS test_cases (
        test_id SERIAL PRIMARY KEY,
        test_name TEXT NOT NULL,
        category TEXT NOT NULL,
        description TEXT,
        expected_result BOOLEAN,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS test_results (
        result_id SERIAL PRIMARY KEY,
        test_id INTEGER REFERENCES test_cases(test_id),
        function_name TEXT NOT NULL,
        input_data JSONB,
        actual_result BOOLEAN,
        status TEXT,
        error_message TEXT,
        executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
END;
$$ LANGUAGE plpgsql;

-- Função para registrar um novo caso de teste
CREATE OR REPLACE FUNCTION test.register_test_case(
    p_test_name TEXT,
    p_category TEXT,
    p_description TEXT,
    p_expected_result BOOLEAN
)
RETURNS INTEGER AS $$
DECLARE
    v_test_id INTEGER;
BEGIN
    INSERT INTO test_cases (test_name, category, description, expected_result)
    VALUES (p_test_name, p_category, p_description, p_expected_result)
    RETURNING test_id INTO v_test_id;
    
    RETURN v_test_id;
END;
$$ LANGUAGE plpgsql;

-- Função para executar um teste
CREATE OR REPLACE FUNCTION test.run_test(
    p_test_id INTEGER,
    p_function_name TEXT,
    p_input_data JSONB
)
RETURNS BOOLEAN AS $$
DECLARE
    v_actual_result BOOLEAN;
    v_status TEXT;
    v_error_message TEXT;
    v_expected_result BOOLEAN;
BEGIN
    -- Obter o resultado esperado
    SELECT expected_result INTO v_expected_result
    FROM test_cases 
    WHERE test_id = p_test_id;
    
    -- Executar a função de teste
    BEGIN
        EXECUTE format('SELECT %I(%L)', p_function_name, p_input_data::TEXT)
        INTO v_actual_result;
        
        v_status := CASE 
            WHEN v_actual_result = v_expected_result THEN 'PASS'
            ELSE 'FAIL'
        END;
        
        v_error_message := NULL;
    
    EXCEPTION WHEN OTHERS THEN
        v_actual_result := NULL;
        v_status := 'ERROR';
        v_error_message := SQLERRM;
    END;
    
    -- Registrar o resultado
    INSERT INTO test_results (
        test_id,
        function_name,
        input_data,
        actual_result,
        status,
        error_message
    ) VALUES (
        p_test_id,
        p_function_name,
        p_input_data,
        v_actual_result,
        v_status,
        v_error_message
    );
    
    RETURN v_status = 'PASS';
END;
$$ LANGUAGE plpgsql;

-- Função para executar todos os testes de uma categoria
CREATE OR REPLACE FUNCTION test.run_category_tests(
    p_category TEXT
)
RETURNS TABLE (
    test_name TEXT,
    function_name TEXT,
    status TEXT,
    error_message TEXT
) AS $$
DECLARE
    v_test_id INTEGER;
    v_test_name TEXT;
    v_function_name TEXT;
    v_input_data JSONB;
BEGIN
    FOR v_test_id IN 
        SELECT test_id 
        FROM test_cases 
        WHERE category = p_category
    LOOP
        SELECT test_name INTO v_test_name
        FROM test_cases 
        WHERE test_id = v_test_id;
        
        -- Obter os dados de teste específicos para cada função
        SELECT function_name, input_data INTO v_function_name, v_input_data
        FROM test_cases 
        WHERE test_id = v_test_id;
        
        -- Executar o teste
        PERFORM test.run_test(v_test_id, v_function_name, v_input_data);
        
        -- Retornar os resultados
        RETURN QUERY
        SELECT 
            tc.test_name,
            tr.function_name,
            tr.status,
            tr.error_message
        FROM test_cases tc
        JOIN test_results tr ON tc.test_id = tr.test_id
        WHERE tc.test_id = v_test_id
        ORDER BY tr.executed_at DESC
        LIMIT 1;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar relatório de testes
CREATE OR REPLACE FUNCTION test.generate_test_report()
RETURNS TABLE (
    category TEXT,
    total_tests INTEGER,
    passed INTEGER,
    failed INTEGER,
    errors INTEGER,
    pass_rate FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        tc.category,
        COUNT(DISTINCT tc.test_id) as total_tests,
        COUNT(CASE WHEN tr.status = 'PASS' THEN 1 END) as passed,
        COUNT(CASE WHEN tr.status = 'FAIL' THEN 1 END) as failed,
        COUNT(CASE WHEN tr.status = 'ERROR' THEN 1 END) as errors,
        ROUND(COUNT(CASE WHEN tr.status = 'PASS' THEN 1 END)::FLOAT / COUNT(DISTINCT tc.test_id) * 100, 2) as pass_rate
    FROM test_cases tc
    LEFT JOIN test_results tr ON tc.test_id = tr.test_id
    GROUP BY tc.category
    ORDER BY tc.category;
END;
$$ LANGUAGE plpgsql;

-- Função para limpar o ambiente de testes
CREATE OR REPLACE FUNCTION test.cleanup_test_environment()
RETURNS VOID AS $$
BEGIN
    DELETE FROM test_results;
    DELETE FROM test_cases;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de uso:
-- SELECT test.setup_test_environment();
-- SELECT test.register_test_case('Teste de Login Básico', 'KB-01', 'Verifica login com senha simples', true);
-- SELECT test.run_test(1, 'auth.verify_password', '{"password": "123456"}'::jsonb);
-- SELECT test.generate_test_report();
-- SELECT test.cleanup_test_environment();
