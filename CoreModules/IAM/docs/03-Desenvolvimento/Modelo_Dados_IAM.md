# Modelo de Dados IAM

## Introdução

Este documento descreve o modelo de dados e estruturas de integração do módulo de Gerenciamento de Identidade e Acesso (IAM) da plataforma INNOVABIZ. O modelo de dados foi projetado para suportar os requisitos de multi-tenancy, conformidade regulatória e extensibilidade, seguindo as melhores práticas de modelagem de dados empresariais.

## Princípios de Modelagem

O modelo de dados do IAM segue os seguintes princípios:

1. **Isolamento de Tenant**: Completa separação de dados entre tenants
2. **Auditabilidade**: Rastreamento completo de todas as alterações
3. **Extensibilidade**: Suporte a atributos personalizados e extensões
4. **Normatização**: Normalização adequada para integridade de dados
5. **Performance**: Otimizações para consultas frequentes
6. **Conformidade**: Suporte a requisitos regulatórios diversos

## Modelo de Dados Lógico

### Entidades Principais

#### Tenant

Representa uma organização ou unidade organizacional isolada.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| tenant_id | UUID | Identificador único do tenant |
| name | String | Nome do tenant |
| domain | String | Domínio principal do tenant |
| status | Enum | Status do tenant (active, suspended, etc.) |
| plan_type | Enum | Tipo de plano/assinatura |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| attributes | JSONB | Atributos dinâmicos específicos do tenant |

#### User

Representa uma identidade digital individual.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| user_id | UUID | Identificador único do usuário |
| tenant_id | UUID | Referência ao tenant |
| username | String | Nome de usuário único dentro do tenant |
| email | String | Endereço de e-mail principal |
| first_name | String | Primeiro nome |
| last_name | String | Sobrenome |
| status | Enum | Status do usuário (active, suspended, etc.) |
| password_hash | String | Hash da senha (armazenado com Argon2id) |
| password_updated_at | Timestamp | Data da última atualização de senha |
| mfa_enabled | Boolean | Indica se MFA está ativado |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| last_login_at | Timestamp | Data do último login |
| attributes | JSONB | Atributos dinâmicos específicos do usuário |

#### Role

Representa um conjunto de responsabilidades e permissões.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| role_id | UUID | Identificador único do papel |
| tenant_id | UUID | Referência ao tenant |
| name | String | Nome do papel |
| description | String | Descrição do papel |
| is_system_role | Boolean | Indica se é um papel de sistema |
| parent_role_id | UUID | Referência a um papel pai (hierarquia) |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| attributes | JSONB | Atributos dinâmicos específicos do papel |

#### Permission

Representa uma capacidade específica no sistema.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| permission_id | UUID | Identificador único da permissão |
| tenant_id | UUID | Referência ao tenant |
| name | String | Nome da permissão |
| description | String | Descrição da permissão |
| resource_type | String | Tipo de recurso (ex: "PATIENT_RECORDS") |
| action | String | Ação permitida (ex: "READ", "WRITE") |
| scope | String | Escopo da permissão |
| is_system_permission | Boolean | Indica se é uma permissão de sistema |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |

#### Group

Representa um agrupamento lógico de usuários.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| group_id | UUID | Identificador único do grupo |
| tenant_id | UUID | Referência ao tenant |
| name | String | Nome do grupo |
| description | String | Descrição do grupo |
| parent_group_id | UUID | Referência a um grupo pai (hierarquia) |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| attributes | JSONB | Atributos dinâmicos específicos do grupo |

### Entidades de Relacionamento

#### UserRole

Associa usuários a papéis.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| user_id | UUID | Referência ao usuário |
| role_id | UUID | Referência ao papel |
| tenant_id | UUID | Referência ao tenant |
| assigned_by | UUID | Usuário que atribuiu o papel |
| valid_from | Timestamp | Data de início da validade |
| valid_to | Timestamp | Data de término da validade |
| created_at | Timestamp | Data de criação |

#### RolePermission

Associa papéis a permissões.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| role_id | UUID | Referência ao papel |
| permission_id | UUID | Referência à permissão |
| tenant_id | UUID | Referência ao tenant |
| created_at | Timestamp | Data de criação |

#### UserGroup

Associa usuários a grupos.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| user_id | UUID | Referência ao usuário |
| group_id | UUID | Referência ao grupo |
| tenant_id | UUID | Referência ao tenant |
| created_at | Timestamp | Data de criação |

### Entidades de Autenticação

#### MFAMethod

Registra métodos MFA disponíveis para um usuário.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| method_id | UUID | Identificador único do método |
| user_id | UUID | Referência ao usuário |
| tenant_id | UUID | Referência ao tenant |
| type | Enum | Tipo de método (totp, sms, email, etc.) |
| identifier | String | Identificador do método (ex: número de telefone) |
| secret | String | Segredo criptografado (quando aplicável) |
| is_primary | Boolean | Indica se é o método primário |
| is_enabled | Boolean | Indica se o método está ativo |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| last_used_at | Timestamp | Data do último uso |

#### ARVRMethod

Registra métodos de autenticação AR/VR.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| method_id | UUID | Identificador único do método |
| user_id | UUID | Referência ao usuário |
| tenant_id | UUID | Referência ao tenant |
| type | Enum | Tipo de método (gesture, gaze, spatial_password) |
| template_data | Binary | Dados do template (criptografados) |
| created_at | Timestamp | Data de criação |
| updated_at | Timestamp | Data da última atualização |
| last_used_at | Timestamp | Data do último uso |

#### Session

Armazena informações de sessão ativa.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| session_id | UUID | Identificador único da sessão |
| user_id | UUID | Referência ao usuário |
| tenant_id | UUID | Referência ao tenant |
| token_value | String | Valor do token (hashed) |
| device_info | JSONB | Informação sobre o dispositivo |
| ip_address | String | Endereço IP |
| location | JSONB | Informação de localização |
| issued_at | Timestamp | Data de emissão |
| expires_at | Timestamp | Data de expiração |
| refresh_token_id | UUID | ID do token de atualização associado |
| is_active | Boolean | Indica se a sessão está ativa |

### Entidades de Auditoria

#### AuditLog

Registra eventos de segurança e alterações.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| log_id | UUID | Identificador único do log |
| tenant_id | UUID | Referência ao tenant |
| event_type | String | Tipo de evento |
| user_id | UUID | Usuário que realizou a ação |
| resource_type | String | Tipo de recurso afetado |
| resource_id | String | ID do recurso afetado |
| action | String | Ação realizada |
| status | String | Status da ação |
| timestamp | Timestamp | Momento do evento |
| details | JSONB | Detalhes adicionais do evento |
| previous_state | JSONB | Estado anterior (quando aplicável) |
| new_state | JSONB | Novo estado (quando aplicável) |
| ip_address | String | Endereço IP |
| user_agent | String | User-Agent do cliente |

### Entidades de Compliance

#### ComplianceValidation

Registra validações de compliance realizadas.

| Atributo | Tipo | Descrição |
|----------|------|-----------|
| validation_id | UUID | Identificador único da validação |
| tenant_id | UUID | Referência ao tenant |
| regulation | String | Regulamento avaliado |
| validation_type | String | Tipo de validação |
| resource_type | String | Tipo de recurso validado |
| resource_id | String | ID do recurso validado |
| status | String | Resultado da validação |
| timestamp | Timestamp | Momento da validação |
| details | JSONB | Detalhes da validação |
| remediation_steps | JSONB | Passos de remediação (quando não conforme) |
| validated_by | UUID | Usuário ou sistema que realizou a validação |

## Modelo Físico

### Otimizações de Banco de Dados

1. **Índices**
   - Índices em chaves estrangeiras
   - Índices compostos para consultas frequentes
   - Índices parciais para subconjuntos específicos

2. **Particionamento**
   - Particionamento de tabelas de auditoria por data
   - Particionamento de dados de tenant para grandes deployments

3. **Segurança em Nível de Linha**
   - RLS (Row-Level Security) para isolamento de tenant
   - Políticas por tabela para garantir separação de dados

### Exemplo de Política RLS

```sql
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

## Extensibilidade

### Atributos Customizados

O modelo suporta atributos personalizados via colunas JSONB:

1. **Schema Validation**: Validação de esquema para garantir consistência
2. **Indexação**: Índices GIN para busca eficiente em atributos
3. **Path Queries**: Consultas de caminho para acesso a atributos aninhados

Exemplo de consulta com atributos personalizados:
```sql
SELECT * FROM users
WHERE attributes->>'department' = 'Engineering'
AND (attributes->>'location')::jsonb ? 'Lisboa';
```

### Extensões de Modelo

Para extensões específicas por setor, utilizamos tabelas de extensão com chaves estrangeiras:

```sql
CREATE TABLE healthcare_user_extensions (
    user_id UUID REFERENCES users(user_id),
    medical_license_number VARCHAR(50),
    specialty VARCHAR(100),
    hospital_affiliations JSONB,
    PRIMARY KEY (user_id)
);
```

## Integração com Outros Módulos

### Integrações Internas

| Módulo | Entidades Compartilhadas | Tipo de Integração |
|--------|--------------------------|-------------------|
| ERP | User, Role, Permission | API REST/GraphQL + Event-based |
| CRM | User, Group | API REST/GraphQL + Event-based |
| Pagamentos | User, Permission | API REST/GraphQL + Event-based |
| Marketplaces | User, Session | API REST/GraphQL + Event-based |

### Integrações Externas

1. **Diretórios Corporativos**
   - LDAP/Active Directory usando mapeamentos configuráveis
   - SCIM 2.0 para provisionamento automático

2. **Provedores de Identidade**
   - OpenID Connect para federação de identidade
   - SAML 2.0 para autenticação empresarial

## Versionamento e Migração

### Estratégia de Versionamento

- Controle de versão explícito em todos os objetos de banco de dados
- Scripts de migração com upgrade e rollback
- Transações atômicas para alterações relacionadas

### Sistema de Migração

Utiliza Flyway para gerenciamento de esquema:

```sql
-- V1_0_0__initial_schema.sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    ...
);
```

## Considerações de Performance

### Otimizações de Consulta

1. **Caching Estratégico**
   - Cache de permissões efetivas
   - Cache de dados de sessão
   - Cache de decisões de autorização

2. **Consultas Eficientes**
   - Materialização de visões para consultas complexas
   - Funções específicas para cálculos frequentes

### Exemplo de Função para Permissões Efetivas

```sql
CREATE OR REPLACE FUNCTION get_effective_permissions(
    p_user_id UUID,
    p_tenant_id UUID
) RETURNS TABLE (
    permission_id UUID,
    name VARCHAR,
    resource_type VARCHAR,
    action VARCHAR
) AS $$
BEGIN
    RETURN QUERY
        SELECT DISTINCT p.permission_id, p.name, p.resource_type, p.action
        FROM permissions p
        JOIN role_permissions rp ON p.permission_id = rp.permission_id
        JOIN user_roles ur ON rp.role_id = ur.role_id
        WHERE ur.user_id = p_user_id
        AND p.tenant_id = p_tenant_id
        AND ur.tenant_id = p_tenant_id
        AND (ur.valid_to IS NULL OR ur.valid_to > NOW());
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## Gestão de Dados e Retenção

### Políticas de Retenção

1. **Logs de Auditoria**
   - Retenção baseada em requisitos regulatórios
   - Arquivamento automático para storage de longo prazo

2. **Dados de Sessão**
   - Limpeza automática após expiração
   - Retenção estendida para investigação de incidentes

### Níveis de Isolamento

1. **Isolamento Padrão**
   - Nível READ COMMITTED para operações regulares
  
2. **Isolamento Elevado**
   - SERIALIZABLE para operações críticas de segurança

## Considerações de Backup e Recuperação

### Estratégia de Backup

1. **Backups Incrementais**
   - Diários para dados transacionais
   - Com validação de integridade

2. **Backups Completos**
   - Semanais para recuperação completa
   - Armazenados com criptografia

3. **Point-in-Time Recovery**
   - Archive logs para recuperação em qualquer ponto

## Conclusão

O modelo de dados do IAM da plataforma INNOVABIZ foi projetado para fornecer uma base robusta, segura e extensível para operações de identidade e acesso. Ele atende aos requisitos de multi-tenancy, escalabilidade e conformidade regulatória, enquanto mantém flexibilidade para adaptação a diferentes setores e casos de uso.
