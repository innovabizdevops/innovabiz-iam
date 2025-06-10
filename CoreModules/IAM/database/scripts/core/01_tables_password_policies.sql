-- INNOVABIZ - Tabela de Políticas de Senha
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Tabela para armazenar políticas de senha das organizações

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Tabela de políticas de senha
CREATE TABLE IF NOT EXISTS password_policies (
    id UUID DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    min_length INTEGER NOT NULL DEFAULT 8,
    require_uppercase BOOLEAN NOT NULL DEFAULT TRUE,
    require_lowercase BOOLEAN NOT NULL DEFAULT TRUE,
    require_number BOOLEAN NOT NULL DEFAULT TRUE,
    require_special_char BOOLEAN NOT NULL DEFAULT TRUE,
    max_age_days INTEGER,
    history_size INTEGER DEFAULT 5,
    max_attempts INTEGER DEFAULT 5,
    lockout_duration_minutes INTEGER DEFAULT 30,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to VARCHAR(50) NOT NULL DEFAULT 'all_users', -- all_users, specific_roles, specific_users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT fk_password_policies_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT pk_password_policies PRIMARY KEY (id)
);

-- Comentários
COMMENT ON TABLE password_policies IS 'Armazena as políticas de senha das organizações';
COMMENT ON COLUMN password_policies.id IS 'Identificador único da política de senha';
COMMENT ON COLUMN password_policies.organization_id IS 'Organização à qual a política se aplica';
COMMENT ON COLUMN password_policies.name IS 'Nome da política de senha';
COMMENT ON COLUMN password_policies.description IS 'Descrição detalhada da política';
COMMENT ON COLUMN password_policies.min_length IS 'Tamanho mínimo da senha';
COMMENT ON COLUMN password_policies.require_uppercase IS 'Se a senha deve conter letras maiúsculas';
COMMENT ON COLUMN password_policies.require_lowercase IS 'Se a senha deve conter letras minúsculas';
COMMENT ON COLUMN password_policies.require_number IS 'Se a senha deve conter números';
COMMENT ON COLUMN password_policies.require_special_char IS 'Se a senha deve conter caracteres especiais';
COMMENT ON COLUMN password_policies.max_age_days IS 'Número máximo de dias que uma senha pode ser usada antes de expirar';
COMMENT ON COLUMN password_policies.history_size IS 'Número de senhas anteriores que não podem ser reutilizadas';
COMMENT ON COLUMN password_policies.max_attempts IS 'Número máximo de tentativas de login antes do bloqueio';
COMMENT ON COLUMN password_policies.lockout_duration_minutes IS 'Duração do bloqueio da conta após exceder o número máximo de tentativas';
COMMENT ON COLUMN password_policies.is_active IS 'Indica se a política está ativa';
COMMENT ON COLUMN password_policies.applies_to IS 'A quem a política se aplica (todos, funções específicas, usuários específicos)';
COMMENT ON COLUMN password_policies.created_at IS 'Data de criação do registro';
COMMENT ON COLUMN password_policies.updated_at IS 'Data da última atualização do registro';
COMMENT ON COLUMN password_policies.created_by IS 'ID do usuário que criou o registro';
COMMENT ON COLUMN password_policies.updated_by IS 'ID do último usuário que atualizou o registro';
COMMENT ON COLUMN password_policies.metadata IS 'Metadados adicionais da política';

-- Índices
CREATE INDEX IF NOT EXISTS idx_password_policies_organization_id ON iam.password_policies(organization_id);
CREATE INDEX IF NOT EXISTS idx_password_policies_is_active ON iam.password_policies(is_active);

-- Trigger para atualizar o campo updated_at
CREATE OR REPLACE FUNCTION update_password_policies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_password_policies_updated_at
BEFORE UPDATE ON iam.password_policies
FOR EACH ROW EXECUTE FUNCTION iam.update_password_policies_updated_at();

-- Inserir política de senha padrão para a organização InnovaBiz
INSERT INTO iam.password_policies (
    id,
    organization_id,
    name,
    description,
    min_length,
    require_uppercase,
    require_lowercase,
    require_number,
    require_special_char,
    max_age_days,
    history_size,
    max_attempts,
    lockout_duration_minutes,
    is_active,
    applies_to,
    created_by,
    updated_by
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111', -- ID da organização InnovaBiz
    'Política de Senha Padrão',
    'Política de senha padrão para todos os usuários',
    12, -- Mínimo de 12 caracteres
    TRUE, -- Requer letras maiúsculas
    TRUE, -- Requer letras minúsculas
    TRUE, -- Requer números
    TRUE, -- Requer caracteres especiais
    90, -- Expira a cada 90 dias
    5, -- Armazena as últimas 5 senhas
    5, -- 5 tentativas de login
    30, -- Bloqueia por 30 minutos
    TRUE, -- Ativa
    'all_users', -- Aplica-se a todos os usuários
    '22222222-2222-2222-2222-222222222222', -- ID do usuário admin
    '22222222-2222-2222-2222-222222222222'  -- ID do usuário admin
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    min_length = EXCLUDED.min_length,
    require_uppercase = EXCLUDED.require_uppercase,
    require_lowercase = EXCLUDED.require_lowercase,
    require_number = EXCLUDED.require_number,
    require_special_char = EXCLUDED.require_special_char,
    max_age_days = EXCLUDED.max_age_days,
    history_size = EXCLUDED.history_size,
    max_attempts = EXCLUDED.max_attempts,
    lockout_duration_minutes = EXCLUDED.lockout_duration_minutes,
    is_active = EXCLUDED.is_active,
    applies_to = EXCLUDED.applies_to,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

-- Comentário para o trigger
COMMENT ON TRIGGER trg_password_policies_updated_at ON iam.password_policies IS 'Atualiza automaticamente o campo updated_at quando um registro é atualizado';

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Tabela password_policies criada com sucesso';
END
$$;
