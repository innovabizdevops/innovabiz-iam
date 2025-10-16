# Exemplos de Implementação de Métodos de Autenticação

## 1. Autenticação Biométrica

### 1.1 Reconhecimento Facial

```sql
-- Tabela de Configuração de Reconhecimento Facial
CREATE TABLE face_recognition_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    template BYTEA NOT NULL,  -- Template facial
    quality_score DECIMAL(5,4) NOT NULL,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação Facial
CREATE OR REPLACE FUNCTION biometric.verify_face_recognition(
    p_user_id UUID,
    p_template BYTEA,
    p_quality_score DECIMAL(5,4)
) RETURNS BOOLEAN AS $$
DECLARE
    v_existing_template BYTEA;
    v_similarity_score DECIMAL(5,4);
BEGIN
    -- Verifica qualidade mínima
    IF p_quality_score < 0.8 THEN
        RAISE EXCEPTION 'Qualidade da imagem abaixo do mínimo permitido';
    END IF;

    -- Busca template existente
    SELECT template INTO v_existing_template
    FROM face_recognition_config
    WHERE user_id = p_user_id
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Usuário não possui template facial cadastrado';
    END IF;

    -- Calcula similaridade (exemplo usando função hipotética)
    v_similarity_score := calculate_similarity(v_existing_template, p_template);

    -- Verifica threshold
    IF v_similarity_score >= 0.6 THEN
        RETURN true;
    ELSE
        RAISE EXCEPTION 'Face não reconhecida (score: %)', v_similarity_score;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 1.2 Impressão Digital

```sql
-- Tabela de Templates de Impressão Digital
CREATE TABLE fingerprint_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    finger_position VARCHAR(10) NOT NULL,  -- thumb, index, middle, ring, pinky
    template BYTEA NOT NULL,
    quality_score DECIMAL(5,4) NOT NULL,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação de Impressão Digital
CREATE OR REPLACE FUNCTION biometric.verify_fingerprint(
    p_user_id UUID,
    p_finger_position VARCHAR(10),
    p_template BYTEA,
    p_quality_score DECIMAL(5,4)
) RETURNS BOOLEAN AS $$
DECLARE
    v_existing_template BYTEA;
    v_similarity_score DECIMAL(5,4);
BEGIN
    -- Verifica qualidade mínima
    IF p_quality_score < 0.7 THEN
        RAISE EXCEPTION 'Qualidade da impressão abaixo do mínimo permitido';
    END IF;

    -- Busca template existente
    SELECT template INTO v_existing_template
    FROM fingerprint_templates
    WHERE user_id = p_user_id
    AND finger_position = p_finger_position
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Impressão digital não cadastrada para este dedo';
    END IF;

    -- Calcula similaridade
    v_similarity_score := calculate_similarity(v_existing_template, p_template);

    -- Verifica threshold
    IF v_similarity_score >= 0.7 THEN
        RETURN true;
    ELSE
        RAISE EXCEPTION 'Impressão digital não reconhecida (score: %)', v_similarity_score;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## 2. Autenticação Baseada em Senha

### 2.1 Senhas Tradicionais

```sql
-- Tabela de Senhas
CREATE TABLE passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    hash BYTEA NOT NULL,
    salt BYTEA NOT NULL,
    iterations INTEGER NOT NULL,
    memory_cost INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    active BOOLEAN DEFAULT true
);

-- Função de Verificação de Senha
CREATE OR REPLACE FUNCTION knowledge.verify_traditional_password(
    p_user_id UUID,
    p_password TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    v_hash BYTEA;
    v_salt BYTEA;
    v_iterations INTEGER;
    v_memory_cost INTEGER;
    v_hashed_password BYTEA;
BEGIN
    -- Busca informações de hash
    SELECT hash, salt, iterations, memory_cost
    INTO v_hash, v_salt, v_iterations, v_memory_cost
    FROM passwords
    WHERE user_id = p_user_id
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Usuário não encontrado ou senha não ativa';
    END IF;

    -- Calcula hash
    v_hashed_password := crypt(p_password, v_salt::text);

    -- Verifica correspondência
    IF v_hashed_password = v_hash THEN
        -- Atualiza last_used
        UPDATE passwords
        SET last_used = CURRENT_TIMESTAMP
        WHERE id = (SELECT id FROM passwords WHERE user_id = p_user_id AND active = true);
        
        RETURN true;
    ELSE
        RETURN false;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 2.2 Padrões Gráficos

```sql
-- Tabela de Padrões Gráficos
CREATE TABLE graphic_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    pattern_hash BYTEA NOT NULL,
    points_count INTEGER NOT NULL,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação de Padrão Gráfico
CREATE OR REPLACE FUNCTION knowledge.verify_graphic_pattern(
    p_user_id UUID,
    p_pattern_points VARCHAR[]
) RETURNS BOOLEAN AS $$
DECLARE
    v_pattern_hash BYTEA;
    v_points_count INTEGER;
    v_new_hash BYTEA;
BEGIN
    -- Verifica número mínimo de pontos
    IF array_length(p_pattern_points, 1) < 4 THEN
        RAISE EXCEPTION 'Número mínimo de pontos não atingido';
    END IF;

    -- Busca padrão existente
    SELECT pattern_hash, points_count
    INTO v_pattern_hash, v_points_count
    FROM graphic_patterns
    WHERE user_id = p_user_id
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Usuário não possui padrão gráfico cadastrado';
    END IF;

    -- Gera hash do novo padrão
    v_new_hash := generate_pattern_hash(p_pattern_points);

    -- Verifica correspondência
    IF v_new_hash = v_pattern_hash THEN
        RETURN true;
    ELSE
        RETURN false;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## 3. Autenticação Baseada em Posse

### 3.1 Aplicativos Autenticadores

```sql
-- Tabela de Configuração de Aplicativos
CREATE TABLE authenticator_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    app_id VARCHAR(64) NOT NULL,  -- ID único do app
    secret_key BYTEA NOT NULL,    -- Chave secreta
    issuer VARCHAR(255) NOT NULL, -- Nome do emissor
    algorithm VARCHAR(10) NOT NULL, -- Algoritmo (HOTP/TOTP)
    digits INTEGER NOT NULL,      -- Número de dígitos
    period INTEGER NOT NULL,      -- Período (para TOTP)
    counter BIGINT,               -- Contador (para HOTP)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_verified TIMESTAMP WITH TIME ZONE
);

-- Função de Verificação de OTP
CREATE OR REPLACE FUNCTION possession.verify_app(
    p_user_id UUID,
    p_otp VARCHAR(6)
) RETURNS BOOLEAN AS $$
DECLARE
    v_secret_key BYTEA;
    v_algorithm VARCHAR(10);
    v_digits INTEGER;
    v_period INTEGER;
    v_counter BIGINT;
    v_expected_otp VARCHAR(6);
BEGIN
    -- Busca configuração do app
    SELECT secret_key, algorithm, digits, period, counter
    INTO v_secret_key, v_algorithm, v_digits, v_period, v_counter
    FROM authenticator_apps
    WHERE user_id = p_user_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Usuário não possui aplicativo autenticador configurado';
    END IF;

    -- Gera OTP esperado
    v_expected_otp := generate_otp(
        v_secret_key,
        v_algorithm,
        v_digits,
        v_period,
        v_counter
    );

    -- Verifica correspondência
    IF p_otp = v_expected_otp THEN
        -- Atualiza contador para HOTP
        IF v_algorithm = 'HOTP' THEN
            UPDATE authenticator_apps
            SET counter = counter + 1,
                last_verified = CURRENT_TIMESTAMP
            WHERE user_id = p_user_id;
        ELSE
            UPDATE authenticator_apps
            SET last_verified = CURRENT_TIMESTAMP
            WHERE user_id = p_user_id;
        END IF;
        
        RETURN true;
    ELSE
        RETURN false;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 3.2 SMS/Email OTP

```sql
-- Tabela de OTPs
CREATE TABLE one_time_passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    otp VARCHAR(6) NOT NULL,
    channel VARCHAR(10) NOT NULL,  -- sms/email
    destination VARCHAR(255) NOT NULL,  -- número/email
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP WITH TIME ZONE,
    attempts INTEGER DEFAULT 0
);

-- Função de Geração de OTP
CREATE OR REPLACE FUNCTION possession.generate_otp(
    p_user_id UUID,
    p_channel VARCHAR(10),
    p_destination VARCHAR(255)
) RETURNS VARCHAR AS $$
DECLARE
    v_otp VARCHAR(6);
BEGIN
    -- Gera OTP aleatório
    v_otp := lpad(to_hex(floor(random() * 1000000)::bigint)::text, 6, '0');

    -- Insere no banco
    INSERT INTO one_time_passwords (
        user_id,
        otp,
        channel,
        destination,
        expires_at
    ) VALUES (
        p_user_id,
        v_otp,
        p_channel,
        p_destination,
        CURRENT_TIMESTAMP + INTERVAL '30 seconds'
    );

    -- Retorna OTP para envio
    RETURN v_otp;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função de Verificação de OTP
CREATE OR REPLACE FUNCTION possession.verify_sms(
    p_user_id UUID,
    p_otp VARCHAR(6)
) RETURNS BOOLEAN AS $$
DECLARE
    v_destination VARCHAR(255);
    v_expires_at TIMESTAMP WITH TIME ZONE;
    v_attempts INTEGER;
BEGIN
    -- Busca OTP
    SELECT destination, expires_at, attempts
    INTO v_destination, v_expires_at, v_attempts
    FROM one_time_passwords
    WHERE user_id = p_user_id
    AND otp = p_otp
    AND channel = 'sms'
    AND verified_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'OTP inválido ou já utilizado';
    END IF;

    -- Verifica expiração
    IF CURRENT_TIMESTAMP > v_expires_at THEN
        RAISE EXCEPTION 'OTP expirado';
    END IF;

    -- Verifica tentativas
    IF v_attempts >= 5 THEN
        RAISE EXCEPTION 'Limite de tentativas excedido';
    END IF;

    -- Atualiza como verificado
    UPDATE one_time_passwords
    SET verified_at = CURRENT_TIMESTAMP,
        attempts = attempts + 1
    WHERE user_id = p_user_id
    AND otp = p_otp;

    RETURN true;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## 4. Autenticação Baseada em Localização

### 4.1 Geolocalização

```sql
-- Tabela de Regras de Localização
CREATE TABLE location_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    location_type VARCHAR(20) NOT NULL,  -- country/region/city
    allowed_locations TEXT[] NOT NULL,   -- array de códigos
    radius_meters INTEGER,               -- para localização por coordenadas
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação de Localização
CREATE OR REPLACE FUNCTION location.verify_geolocation(
    p_user_id UUID,
    p_latitude DECIMAL(10,8),
    p_longitude DECIMAL(11,8),
    p_accuracy DECIMAL(10,2)
) RETURNS BOOLEAN AS $$
DECLARE
    v_allowed_locations TEXT[];
    v_location_type VARCHAR(20);
    v_radius_meters INTEGER;
    v_current_location TEXT;
BEGIN
    -- Verifica precisão mínima
    IF p_accuracy > 50 THEN
        RAISE EXCEPTION 'Precisão da localização abaixo do mínimo permitido';
    END IF;

    -- Busca regras de localização
    SELECT allowed_locations, location_type, radius_meters
    INTO v_allowed_locations, v_location_type, v_radius_meters
    FROM location_rules
    WHERE user_id = p_user_id;

    IF NOT FOUND THEN
        RETURN true;  -- Sem regras, permite qualquer localização
    END IF;

    -- Verifica tipo de localização
    IF v_location_type = 'coordinates' THEN
        -- Verifica se está dentro do raio permitido
        IF calculate_distance(
            p_latitude,
            p_longitude,
            v_allowed_locations[1]::decimal,
            v_allowed_locations[2]::decimal
        ) <= v_radius_meters THEN
            RETURN true;
        ELSE
            RAISE EXCEPTION 'Localização fora da área permitida';
        END IF;
    ELSE
        -- Obtém localização atual
        v_current_location := get_location_from_coordinates(
            p_latitude,
            p_longitude
        );

        -- Verifica se está na lista permitida
        IF v_current_location = ANY(v_allowed_locations) THEN
            RETURN true;
        ELSE
            RAISE EXCEPTION 'Localização não autorizada';
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 4.2 Beacon

```sql
-- Tabela de Configuração de Beacons
CREATE TABLE beacons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    beacon_id VARCHAR(32) NOT NULL,  -- ID único do beacon
    location_id UUID REFERENCES locations(id),
    major INTEGER,  -- Identificador maior
    minor INTEGER,  -- Identificador menor
    uuid VARCHAR(36),  -- UUID do beacon
    allowed_distance_meters INTEGER NOT NULL,  -- Distância máxima permitida
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação de Beacon
CREATE OR REPLACE FUNCTION location.verify_beacon(
    p_user_id UUID,
    p_beacon_uuid VARCHAR(36),
    p_major INTEGER,
    p_minor INTEGER,
    p_distance_meters DECIMAL(10,2)
) RETURNS BOOLEAN AS $$
DECLARE
    v_allowed_distance INTEGER;
BEGIN
    -- Busca configuração do beacon
    SELECT allowed_distance_meters
    INTO v_allowed_distance
    FROM beacons
    WHERE uuid = p_beacon_uuid
    AND major = p_major
    AND minor = p_minor
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Beacon não encontrado ou não ativo';
    END IF;

    -- Verifica distância
    IF p_distance_meters <= v_allowed_distance THEN
        RETURN true;
    ELSE
        RAISE EXCEPTION 'Usuário está fora do alcance permitido do beacon';
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## 5. Federação de Identidades

### 5.1 Identidades Federadas

```sql
-- Tabela de Identidades Federadas
CREATE TABLE federated_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    provider_type VARCHAR(20) NOT NULL,  -- saml/oauth/oidc
    provider_id VARCHAR(64) NOT NULL,    -- ID do provedor
    provider_user_id VARCHAR(255) NOT NULL,  -- ID do usuário no provedor
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    last_verified TIMESTAMP WITH TIME ZONE,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Verificação de Identidade Federada
CREATE OR REPLACE FUNCTION federated.verify_identity(
    p_user_id UUID,
    p_provider_type VARCHAR(20),
    p_provider_user_id VARCHAR(255)
) RETURNS BOOLEAN AS $$
DECLARE
    v_provider_id VARCHAR(64);
    v_token_expires_at TIMESTAMP WITH TIME ZONE;
    v_access_token TEXT;
BEGIN
    -- Busca identidade federada
    SELECT provider_id, token_expires_at, access_token
    INTO v_provider_id, v_token_expires_at, v_access_token
    FROM federated_identities
    WHERE user_id = p_user_id
    AND provider_type = p_provider_type
    AND provider_user_id = p_provider_user_id
    AND active = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Identidade federada não encontrada ou não ativa';
    END IF;

    -- Verifica se token está expirado
    IF CURRENT_TIMESTAMP > v_token_expires_at THEN
        -- Atualiza token usando refresh_token
        PERFORM refresh_federated_token(p_user_id, v_provider_id);
    END IF;

    -- Atualiza timestamp de verificação
    UPDATE federated_identities
    SET last_verified = CURRENT_TIMESTAMP
    WHERE user_id = p_user_id
    AND provider_type = p_provider_type
    AND provider_user_id = p_provider_user_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 5.2 Auditoria de Federação

```sql
-- Tabela de Logs de Federação
CREATE TABLE federation_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    provider_type VARCHAR(20) NOT NULL,
    provider_id VARCHAR(64) NOT NULL,
    action_type VARCHAR(20) NOT NULL,  -- login/logout/token_refresh
    status VARCHAR(20) NOT NULL,      -- success/failure
    error_message TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Função de Logging de Federação
CREATE OR REPLACE FUNCTION federated.log_federation_event(
    p_user_id UUID,
    p_provider_type VARCHAR(20),
    p_provider_id VARCHAR(64),
    p_action_type VARCHAR(20),
    p_status VARCHAR(20),
    p_error_message TEXT,
    p_ip_address VARCHAR(45),
    p_user_agent TEXT
) RETURNS VOID AS $$
BEGIN
    INSERT INTO federation_audit_logs (
        user_id,
        provider_type,
        provider_id,
        action_type,
        status,
        error_message,
        ip_address,
        user_agent
    ) VALUES (
        p_user_id,
        p_provider_type,
        p_provider_id,
        p_action_type,
        p_status,
        p_error_message,
        p_ip_address,
        p_user_agent
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```
