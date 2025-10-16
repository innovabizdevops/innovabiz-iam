-- Schema do módulo IAM para INNOVABIZ Platform
-- Compatível com PostgreSQL 13+
-- Implementação multi-tenant, multi-dimensional e segura com Row-Level Security (RLS)

-- Criação do esquema IAM
CREATE SCHEMA IF NOT EXISTS iam;

-- Tabela de auditoria para registro de operações
CREATE TABLE IF NOT EXISTS iam.audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    performed_by UUID NOT NULL,
    performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Índices para a tabela de auditoria
CREATE INDEX IF NOT EXISTS idx_audit_tenant_entity ON iam.audit_log (tenant_id, entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_performed_at ON iam.audit_log (performed_at);
CREATE INDEX IF NOT EXISTS idx_audit_action ON iam.audit_log (action);
CREATE INDEX IF NOT EXISTS idx_audit_performed_by ON iam.audit_log (performed_by);

-- Tabela de usuários (simplificada, apenas para referência)
CREATE TABLE IF NOT EXISTS iam.users (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    username VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by UUID,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    metadata JSONB,
    CONSTRAINT uk_users_username_tenant UNIQUE (username, tenant_id),
    CONSTRAINT uk_users_email_tenant UNIQUE (email, tenant_id)
);

-- Índices para a tabela de usuários
CREATE INDEX IF NOT EXISTS idx_users_tenant ON iam.users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_active ON iam.users (is_active);
CREATE INDEX IF NOT EXISTS idx_users_email ON iam.users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON iam.users (deleted_at) WHERE deleted_at IS NOT NULL;

-- Tabela de funções (roles)
CREATE TABLE IF NOT EXISTS iam.roles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by UUID,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    metadata JSONB,
    CONSTRAINT uk_roles_code_tenant UNIQUE (code, tenant_id)
);

-- Índices para a tabela de funções
CREATE INDEX IF NOT EXISTS idx_roles_tenant ON iam.roles (tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_code ON iam.roles (code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_active ON iam.roles (is_active);
CREATE INDEX IF NOT EXISTS idx_roles_type ON iam.roles (type);
CREATE INDEX IF NOT EXISTS idx_roles_system ON iam.roles (is_system);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON iam.roles (deleted_at) WHERE deleted_at IS NOT NULL;

-- Tabela de permissões
CREATE TABLE IF NOT EXISTS iam.permissions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255), -- NULL para permissões gerais de tipo de recurso
    action VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by UUID,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    metadata JSONB,
    CONSTRAINT uk_permissions_code_tenant UNIQUE (code, tenant_id)
);

-- Índices para a tabela de permissões
CREATE INDEX IF NOT EXISTS idx_permissions_tenant ON iam.permissions (tenant_id);
CREATE INDEX IF NOT EXISTS idx_permissions_code ON iam.permissions (code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_permissions_active ON iam.permissions (is_active);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_type ON iam.permissions (resource_type);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON iam.permissions (resource_type, action);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_full ON iam.permissions (resource_type, action, resource_id);
CREATE INDEX IF NOT EXISTS idx_permissions_system ON iam.permissions (is_system);
CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON iam.permissions (deleted_at) WHERE deleted_at IS NOT NULL;-- Tabela de hierarquia de funções
CREATE TABLE IF NOT EXISTS iam.role_hierarchy (
    parent_id UUID NOT NULL,
    child_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    PRIMARY KEY (parent_id, child_id, tenant_id),
    FOREIGN KEY (parent_id, tenant_id) REFERENCES iam.roles (id, tenant_id),
    FOREIGN KEY (child_id, tenant_id) REFERENCES iam.roles (id, tenant_id),
    CONSTRAINT chk_no_self_reference CHECK (parent_id <> child_id)
);

-- Índice para a tabela de hierarquia de funções
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_parent ON iam.role_hierarchy (parent_id);
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_child ON iam.role_hierarchy (child_id);
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_tenant ON iam.role_hierarchy (tenant_id);

-- Tabela de associação entre funções e permissões
CREATE TABLE IF NOT EXISTS iam.role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    PRIMARY KEY (role_id, permission_id, tenant_id),
    FOREIGN KEY (role_id, tenant_id) REFERENCES iam.roles (id, tenant_id),
    FOREIGN KEY (permission_id, tenant_id) REFERENCES iam.permissions (id, tenant_id)
);

-- Índices para a tabela de associação entre funções e permissões
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON iam.role_permissions (role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON iam.role_permissions (permission_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_tenant ON iam.role_permissions (tenant_id);

-- Tabela de associação entre usuários e funções
CREATE TABLE IF NOT EXISTS iam.user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    activates_at TIMESTAMP WITH TIME ZONE, -- Data em que a função se torna ativa para o usuário
    expires_at TIMESTAMP WITH TIME ZONE,   -- Data em que a função expira para o usuário
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    PRIMARY KEY (user_id, role_id, tenant_id),
    FOREIGN KEY (user_id, tenant_id) REFERENCES iam.users (id, tenant_id),
    FOREIGN KEY (role_id, tenant_id) REFERENCES iam.roles (id, tenant_id)
);

-- Índices para a tabela de associação entre usuários e funções
CREATE INDEX IF NOT EXISTS idx_user_roles_user ON iam.user_roles (user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON iam.user_roles (role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant ON iam.user_roles (tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_activates_at ON iam.user_roles (activates_at) 
    WHERE activates_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_roles_expires_at ON iam.user_roles (expires_at) 
    WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_roles_active_window ON iam.user_roles (activates_at, expires_at);

-- Funções para verificação de ciclos na hierarquia de funções
CREATE OR REPLACE FUNCTION iam.check_role_cycle() RETURNS TRIGGER AS $$
BEGIN
    -- Verifica se a inserção criaria um ciclo
    IF EXISTS (
        WITH RECURSIVE cycle_check AS (
            -- Caso base: comece com o child_id
            SELECT NEW.child_id AS id, 1 AS depth
            UNION ALL
            -- Parte recursiva: obtenha todos os ancestrais
            SELECT rh.parent_id, cc.depth + 1
            FROM iam.role_hierarchy rh
            JOIN cycle_check cc ON cc.id = rh.child_id
            WHERE rh.tenant_id = NEW.tenant_id AND cc.depth < 100 -- limite de profundidade para evitar loops infinitos
        )
        SELECT 1 FROM cycle_check WHERE id = NEW.parent_id
    ) THEN
        RAISE EXCEPTION 'Ciclo detectado na hierarquia de funções';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;-- Trigger para evitar ciclos na hierarquia de funções
CREATE TRIGGER trig_check_role_cycle
BEFORE INSERT ON iam.role_hierarchy
FOR EACH ROW
EXECUTE FUNCTION iam.check_role_cycle();

-- Trigger de auditoria para funções
CREATE OR REPLACE FUNCTION iam.audit_role_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            new_values, metadata, ip_address, user_agent
        ) VALUES (
            NEW.tenant_id, 'ROLE', NEW.id, 'CREATE', 
            COALESCE(NEW.created_by, '00000000-0000-0000-0000-000000000000'::UUID),
            row_to_json(NEW), 
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    ELSIF TG_OP = 'UPDATE' THEN
        -- Se o campo deleted_at foi preenchido, é uma operação de exclusão lógica
        IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
            INSERT INTO iam.audit_log (
                tenant_id, entity_type, entity_id, action, performed_by, 
                old_values, new_values, metadata, ip_address, user_agent
            ) VALUES (
                NEW.tenant_id, 'ROLE', NEW.id, 'DELETE', 
                COALESCE(NEW.deleted_by, '00000000-0000-0000-0000-000000000000'::UUID),
                row_to_json(OLD), row_to_json(NEW), 
                jsonb_build_object('application', current_setting('app.name', true)),
                current_setting('request.ip', true),
                current_setting('request.user_agent', true)
            );
        ELSE
            INSERT INTO iam.audit_log (
                tenant_id, entity_type, entity_id, action, performed_by, 
                old_values, new_values, metadata, ip_address, user_agent
            ) VALUES (
                NEW.tenant_id, 'ROLE', NEW.id, 'UPDATE', 
                COALESCE(NEW.updated_by, '00000000-0000-0000-0000-000000000000'::UUID),
                row_to_json(OLD), row_to_json(NEW), 
                jsonb_build_object('application', current_setting('app.name', true)),
                current_setting('request.ip', true),
                current_setting('request.user_agent', true)
            );
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            old_values, metadata, ip_address, user_agent
        ) VALUES (
            OLD.tenant_id, 'ROLE', OLD.id, 'HARD_DELETE', 
            COALESCE(current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            row_to_json(OLD),
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Aplicar o trigger de auditoria para funções
CREATE TRIGGER trig_audit_roles
AFTER INSERT OR UPDATE OR DELETE ON iam.roles
FOR EACH ROW
EXECUTE FUNCTION iam.audit_role_changes();

-- Trigger de auditoria para permissões
CREATE OR REPLACE FUNCTION iam.audit_permission_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            new_values, metadata, ip_address, user_agent
        ) VALUES (
            NEW.tenant_id, 'PERMISSION', NEW.id, 'CREATE', 
            COALESCE(NEW.created_by, '00000000-0000-0000-0000-000000000000'::UUID),
            row_to_json(NEW), 
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    ELSIF TG_OP = 'UPDATE' THEN
        -- Se o campo deleted_at foi preenchido, é uma operação de exclusão lógica
        IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
            INSERT INTO iam.audit_log (
                tenant_id, entity_type, entity_id, action, performed_by, 
                old_values, new_values, metadata, ip_address, user_agent
            ) VALUES (
                NEW.tenant_id, 'PERMISSION', NEW.id, 'DELETE', 
                COALESCE(NEW.deleted_by, '00000000-0000-0000-0000-000000000000'::UUID),
                row_to_json(OLD), row_to_json(NEW), 
                jsonb_build_object('application', current_setting('app.name', true)),
                current_setting('request.ip', true),
                current_setting('request.user_agent', true)
            );
        ELSE
            INSERT INTO iam.audit_log (
                tenant_id, entity_type, entity_id, action, performed_by, 
                old_values, new_values, metadata, ip_address, user_agent
            ) VALUES (
                NEW.tenant_id, 'PERMISSION', NEW.id, 'UPDATE', 
                COALESCE(NEW.updated_by, '00000000-0000-0000-0000-000000000000'::UUID),
                row_to_json(OLD), row_to_json(NEW), 
                jsonb_build_object('application', current_setting('app.name', true)),
                current_setting('request.ip', true),
                current_setting('request.user_agent', true)
            );
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            old_values, metadata, ip_address, user_agent
        ) VALUES (
            OLD.tenant_id, 'PERMISSION', OLD.id, 'HARD_DELETE', 
            COALESCE(current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            row_to_json(OLD),
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;-- Aplicar o trigger de auditoria para permissões
CREATE TRIGGER trig_audit_permissions
AFTER INSERT OR UPDATE OR DELETE ON iam.permissions
FOR EACH ROW
EXECUTE FUNCTION iam.audit_permission_changes();

-- Trigger de auditoria para associações entre funções e permissões
CREATE OR REPLACE FUNCTION iam.audit_role_permission_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            new_values, metadata, ip_address, user_agent
        ) VALUES (
            NEW.tenant_id, 'ROLE_PERMISSION', NEW.role_id, 'ASSIGN_PERMISSION', 
            COALESCE(NEW.created_by, current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            jsonb_build_object('role_id', NEW.role_id, 'permission_id', NEW.permission_id), 
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            old_values, metadata, ip_address, user_agent
        ) VALUES (
            OLD.tenant_id, 'ROLE_PERMISSION', OLD.role_id, 'REVOKE_PERMISSION', 
            COALESCE(current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            jsonb_build_object('role_id', OLD.role_id, 'permission_id', OLD.permission_id),
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Aplicar o trigger de auditoria para associações entre funções e permissões
CREATE TRIGGER trig_audit_role_permissions
AFTER INSERT OR DELETE ON iam.role_permissions
FOR EACH ROW
EXECUTE FUNCTION iam.audit_role_permission_changes();

-- Trigger de auditoria para associações entre usuários e funções
CREATE OR REPLACE FUNCTION iam.audit_user_role_changes() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            new_values, metadata, ip_address, user_agent
        ) VALUES (
            NEW.tenant_id, 'USER_ROLE', NEW.user_id, 'ASSIGN_ROLE', 
            COALESCE(NEW.created_by, current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            jsonb_build_object('user_id', NEW.user_id, 'role_id', NEW.role_id, 'activates_at', NEW.activates_at, 'expires_at', NEW.expires_at), 
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO iam.audit_log (
            tenant_id, entity_type, entity_id, action, performed_by, 
            old_values, metadata, ip_address, user_agent
        ) VALUES (
            OLD.tenant_id, 'USER_ROLE', OLD.user_id, 'REVOKE_ROLE', 
            COALESCE(current_setting('app.user_id', true)::UUID, '00000000-0000-0000-0000-000000000000'::UUID),
            jsonb_build_object('user_id', OLD.user_id, 'role_id', OLD.role_id),
            jsonb_build_object('application', current_setting('app.name', true)),
            current_setting('request.ip', true),
            current_setting('request.user_agent', true)
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Aplicar o trigger de auditoria para associações entre usuários e funções
CREATE TRIGGER trig_audit_user_roles
AFTER INSERT OR DELETE ON iam.user_roles
FOR EACH ROW
EXECUTE FUNCTION iam.audit_user_role_changes();

-- Habilitar o Row-Level Security para todas as tabelas
ALTER TABLE iam.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.role_hierarchy ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.role_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.audit_log ENABLE ROW LEVEL SECURITY;

-- Criar função para verificar tenant do usuário atual
CREATE OR REPLACE FUNCTION iam.current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.tenant_id', TRUE)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Criar função para verificar se o usuário atual é um administrador de sistema
CREATE OR REPLACE FUNCTION iam.is_system_admin() RETURNS BOOLEAN AS $$
BEGIN
    RETURN current_setting('app.is_system_admin', TRUE)::BOOLEAN;
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Políticas de segurança por tenant para cada tabela
CREATE POLICY users_tenant_isolation ON iam.users
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY roles_tenant_isolation ON iam.roles
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY permissions_tenant_isolation ON iam.permissions
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY role_hierarchy_tenant_isolation ON iam.role_hierarchy
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY role_permissions_tenant_isolation ON iam.role_permissions
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY user_roles_tenant_isolation ON iam.user_roles
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

CREATE POLICY audit_log_tenant_isolation ON iam.audit_log
    USING (tenant_id = iam.current_tenant_id() OR iam.is_system_admin());

-- Comentários de documentação das tabelas
COMMENT ON SCHEMA iam IS 'Esquema para o módulo de Identity and Access Management (IAM) da plataforma INNOVABIZ';

COMMENT ON TABLE iam.users IS 'Tabela que armazena informações dos usuários no sistema';
COMMENT ON TABLE iam.roles IS 'Tabela que armazena funções (roles) que podem ser atribuídas a usuários';
COMMENT ON TABLE iam.permissions IS 'Tabela que armazena permissões que podem ser atribuídas a funções';
COMMENT ON TABLE iam.role_hierarchy IS 'Tabela que armazena relações hierárquicas entre funções (funções pai-filho)';
COMMENT ON TABLE iam.role_permissions IS 'Tabela de associação entre funções e permissões';
COMMENT ON TABLE iam.user_roles IS 'Tabela de associação entre usuários e funções';
COMMENT ON TABLE iam.audit_log IS 'Tabela que armazena registros de auditoria para operações no sistema IAM';

-- Função para atualizar o timestamp de atualização automaticamente
CREATE OR REPLACE FUNCTION iam.update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualização automática de timestamps
CREATE TRIGGER update_roles_modtime
    BEFORE UPDATE ON iam.roles
    FOR EACH ROW
    EXECUTE FUNCTION iam.update_modified_column();

CREATE TRIGGER update_permissions_modtime
    BEFORE UPDATE ON iam.permissions
    FOR EACH ROW
    EXECUTE FUNCTION iam.update_modified_column();

CREATE TRIGGER update_users_modtime
    BEFORE UPDATE ON iam.users
    FOR EACH ROW
    EXECUTE FUNCTION iam.update_modified_column();