# Documentação do Esquema IAM - INNOVABIZ

## Visão Geral

Este documento descreve a estrutura do banco de dados do módulo de Identidade e Gerenciamento de Acesso (IAM) da plataforma INNOVABIZ. O esquema IAM é responsável por gerenciar usuários, funções, permissões, autenticação, autorização e auditoria no sistema.

## Tabelas

A seguir estão as tabelas que compõem o esquema IAM:

| Nome | Descrição |
|------|-----------|
| ar_continuous_auth_sessions | Sem descrição |
| attribute_policies | Sem descrição |
| audit_logs | Registra todas as ações significativas realizadas no sistema para fins de auditoria e conformidade. |
| authorization_decisions_cache | Sem descrição |
| compliance_validators | Armazena os validadores de conformidade disponíveis para verificação automática de requisitos regulatórios. |
| detailed_audit_logs | Sem descrição |
| detailed_permissions | Sem descrição |
| detailed_roles | Sem descrição |
| federated_groups | Sem descrição |
| federated_identities | Sem descrição |
| federated_user_groups | Sem descrição |
| federation_sessions | Sem descrição |
| healthcare_compliance_validations | Sem descrição |
| healthcare_regulatory_requirements | Sem descrição |
| himss_emram_action_plans | Planos de ação para endereçar não conformidades com HIMSS EMRAM |
| himss_emram_assessments | Avaliações HIMSS EMRAM realizadas para organizações de saúde |
| himss_emram_benchmarks | Dados de benchmark HIMSS EMRAM por região, país e tipo de estabelecimento |
| himss_emram_certifications | Certificações HIMSS EMRAM alcançadas por organizações de saúde |
| himss_emram_criteria | Critérios para cada estágio do modelo HIMSS EMRAM |
| himss_emram_criteria_results | Resultados da avaliação de cada critério HIMSS EMRAM |
| himss_emram_stages | Estágios definidos pelo modelo HIMSS EMRAM para adoção de registros médicos eletrônicos |
| iam_metrics_history | Armazena m�tricas hist�ricas do m�dulo IAM para an�lise de tend�ncias e monitoramento |
| identity_provider_attribute_mappings | Sem descrição |
| identity_provider_role_mappings | Sem descrição |
| identity_providers | External identity providers for federation |
| iso27001_action_plans | Planos de ação para endereçar não conformidades com ISO 27001 |
| iso27001_assessments | Avaliações de conformidade com ISO 27001 realizadas pela organização |
| iso27001_control_results | Resultados da avaliação de cada controle ISO 27001 |
| iso27001_controls | Controles definidos pelo padrão ISO/IEC 27001 para segurança da informação |
| iso27001_documents | Documentos relacionados à implementação e manutenção do SGSI conforme ISO 27001 |
| iso27001_framework_mapping | Mapeamento entre controles ISO 27001 e outros frameworks regulatórios |
| mfa_organization_settings | Sem descrição |
| mfa_sessions | Sem descrição |
| organizations | Armazena informações sobre as organizações que utilizam a plataforma INNOVABIZ, incluindo configurações específicas e metadados. |
| password_policies | Armazena as políticas de senha das organizações |
| permissions | Lista todas as permissões disponíveis no sistema, que podem ser associadas a funções. |
| policy_set_policies | Sem descrição |
| policy_sets | Sem descrição |
| regulatory_frameworks | Lista os frameworks regulatórios suportados pelo sistema (ex: GDPR, LGPD, HIPAA). |
| role_permissions | Tabela de associação entre Roles e Permissions, definindo quais permissões cada role concede dentro de uma organização. |
| role_policy_sets | Sem descrição |
| roles | Define os papéis que podem ser atribuídos aos usuários, agrupando conjuntos de permissões. |
| schema_migrations | Registra todas as migra��es de esquema aplicadas ao banco de dados |
| security_policies | Armazena as políticas de segurança configuráveis para cada organização. |
| sessions | Registra as sessões ativas dos usuários, incluindo tokens e informações de autenticação. |
| trusted_devices | Sem descrição |
| user_ar_gaze_auth | Sem descrição |
| user_ar_gesture_auth | Sem descrição |
| user_ar_spatial_password | Sem descrição |
| user_mfa_backup_codes | Sem descrição |
| user_mfa_methods | Sem descrição |
| user_roles | Relaciona usuários a funções, permitindo herdar permissões. |
| users | Contém as informações dos usuários do sistema, incluindo credenciais, status e preferências. |

## Detalhes das Tabelas

### ar_continuous_auth_sessions

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| user_id | uuid | NÃO |  |  |
| device_id | character varying(255 | NÃO |  |  |
| session_id | uuid | NÃO |  |  |
| revoked_reason | character varying(100 | SIM |  |  |
| confidence_score | double precision(53) | NÃO |  |  |
| revoked | boolean | SIM | false |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| last_verification | timestamp with time zone | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| expires_at | timestamp with time zone | NÃO |  |  |
### attribute_policies

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| priority | integer(32,0) | NÃO | 100 |  |
| condition_expression | jsonb | NÃO |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| name | character varying(255 | NÃO |  |  |
| description | text | SIM |  |  |
| condition_attributes | jsonb | NÃO |  |  |
| resource_type | character varying(100 | NÃO |  |  |
| resource_pattern | character varying(255 | SIM |  |  |
| action_pattern | character varying(255 | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| is_active | boolean | SIM | true |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| effect | USER-DEFINED | NÃO |  |  |
### audit_logs

Registra todas as ações significativas realizadas no sistema para fins de auditoria e conformidade.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | SIM |  |  |
| status | character varying(50 | NÃO |  |  |
| request_id | character varying(255 | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| resource_type | character varying(100 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| ip_address | character varying(50 | SIM |  |  |
| user_id | uuid | SIM |  |  |
| action | character varying(100 | NÃO |  |  |
| session_id | uuid | SIM |  |  |
| timestamp | timestamp with time zone | SIM | now() |  |
| details | jsonb | SIM | '{}'::jsonb |  |
### authorization_decisions_cache

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| resource_type | character varying(100 | NÃO |  |  |
| resource_id | character varying(255 | SIM |  |  |
| action | character varying(100 | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| expires_at | timestamp with time zone | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| decision_context | jsonb | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| decision | USER-DEFINED | NÃO |  |  |
| user_id | uuid | NÃO |  |  |
### compliance_validators

Armazena os validadores de conformidade disponíveis para verificação automática de requisitos regulatórios.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| updated_at | timestamp with time zone | SIM | now() |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| version | character varying(50 | NÃO |  |  |
| validator_class | character varying(255 | NÃO |  |  |
| description | text | SIM |  |  |
| name | character varying(255 | NÃO |  |  |
| code | character varying(100 | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| created_at | timestamp with time zone | SIM | now() |  |
| is_active | boolean | SIM | true |  |
| configuration | jsonb | SIM | '{}'::jsonb |  |
| framework_id | uuid | NÃO |  |  |
### detailed_audit_logs

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| source_ip | character varying(45 | SIM |  |  |
| details | jsonb | NÃO | '{}'::jsonb |  |
| response_time | integer(32,0) | SIM |  |  |
| user_agent | text | SIM |  |  |
| status | character varying(50 | NÃO |  |  |
| compliance_tags | ARRAY | SIM | ARRAY[]::character varying[] |  |
| regulatory_references | ARRAY | SIM | ARRAY[]::character varying[] |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | SIM |  |  |
| response_payload | jsonb | SIM |  |  |
| user_id | uuid | SIM |  |  |
| event_time | timestamp with time zone | SIM | now() |  |
| event_category | USER-DEFINED | NÃO |  |  |
| severity_level | USER-DEFINED | NÃO | 'info'::iam.audit_severity_level |  |
| request_id | uuid | SIM |  |  |
| session_id | uuid | SIM |  |  |
| request_payload | jsonb | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| geo_location | jsonb | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| action | character varying(100 | NÃO |  |  |
### detailed_permissions

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
| code | character varying(255 | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| permission_scope | USER-DEFINED | NÃO |  |  |
| actions | ARRAY | NÃO |  |  |
| resource_type | character varying(100 | NÃO |  |  |
| description | text | SIM |  |  |
### detailed_roles

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| name | character varying(255 | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| code | character varying(100 | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| parent_role_id | uuid | SIM |  |  |
| is_active | boolean | SIM | true |  |
| is_system_role | boolean | SIM | false |  |
| description | text | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### federated_groups

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| updated_at | timestamp with time zone | SIM | now() |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| provider_id | uuid | NÃO |  |  |
| external_group_id | character varying(255 | NÃO |  |  |
| external_group_name | character varying(255 | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| auto_role_assignment | boolean | SIM | false |  |
| internal_role_id | uuid | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
### federated_identities

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| last_login | timestamp with time zone | SIM |  |  |
| external_id | character varying(255 | NÃO |  |  |
| external_username | character varying(255 | SIM |  |  |
| external_email | character varying(255 | SIM |  |  |
| external_data | jsonb | SIM | '{}'::jsonb |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| provider_id | uuid | NÃO |  |  |
| user_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| organization_id | uuid | NÃO |  |  |
### federated_user_groups

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| federated_identity_id | uuid | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| federated_group_id | uuid | NÃO |  |  |
### federation_sessions

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| provider_id | uuid | NÃO |  |  |
| expires_at | timestamp with time zone | NÃO |  |  |
| revoked | boolean | SIM | false |  |
| user_id | uuid | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| session_token | character varying(255 | NÃO |  |  |
| external_session_id | character varying(255 | SIM |  |  |
| ip_address | character varying(45 | SIM |  |  |
| user_agent | text | SIM |  |  |
| revoked_reason | character varying(100 | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| organization_id | uuid | NÃO |  |  |
### gdpr_audit_view

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_name | character varying(255 | SIM |  |  |
| user_name | character varying(255 | SIM |  |  |
| user_email | character varying(255 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| action | character varying(100 | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| compliance_tags | ARRAY | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| source_ip | character varying(45 | SIM |  |  |
| id | uuid | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| user_id | uuid | SIM |  |  |
| severity_level | USER-DEFINED | SIM |  |  |
| event_category | USER-DEFINED | SIM |  |  |
| regulatory_references | ARRAY | SIM |  |  |
| event_time | timestamp with time zone | SIM |  |  |
| details | jsonb | SIM |  |  |
### healthcare_compliance_validations

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| validator_name | character varying(100 | NÃO |  |  |
| score | integer(32,0) | SIM |  |  |
| regulation | USER-DEFINED | NÃO |  |  |
| validation_timestamp | timestamp with time zone | SIM | now() |  |
| organization_id | uuid | NÃO |  |  |
| validated_by | uuid | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| remediation_plan | text | SIM |  |  |
| status | character varying(50 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| details | jsonb | NÃO |  |  |
### healthcare_regulatory_requirements

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| data_category | USER-DEFINED | NÃO |  |  |
| remediation_steps | text | SIM |  |  |
| requirement_level | character varying(50 | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| requirement_description | text | SIM |  |  |
| requirement_name | character varying(255 | NÃO |  |  |
| requirement_code | character varying(50 | NÃO |  |  |
| region_code | character varying(50 | SIM |  |  |
| country_code | character varying(3 | SIM |  |  |
| is_active | boolean | SIM | true |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| validation_criteria | jsonb | NÃO |  |  |
| regulation | USER-DEFINED | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
### himss_emram_action_plans

Planos de ação para endereçar não conformidades com HIMSS EMRAM

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| title | character varying(255 | NÃO |  |  |
| description | text | SIM |  |  |
| priority | character varying(50 | NÃO |  |  |
| status | character varying(50 | NÃO | 'open'::character varying |  |
| completion_notes | text | SIM |  |  |
| estimated_effort | character varying(100 | SIM |  |  |
| target_stage | integer(32,0) | SIM |  |  |
| completed_by | uuid | SIM |  |  |
| completed_at | timestamp with time zone | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| created_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| assigned_to | uuid | SIM |  |  |
| due_date | date | SIM |  |  |
| criteria_result_id | uuid | SIM |  |  |
| assessment_id | uuid | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### himss_emram_assessments

Avaliações HIMSS EMRAM realizadas para organizações de saúde

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| description | text | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| healthcare_facility_name | character varying(255 | NÃO |  |  |
| facility_type | character varying(100 | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| previous_assessment_id | uuid | SIM |  |  |
| primary_contact_id | uuid | SIM |  |  |
| current_stage | integer(32,0) | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| status | character varying(50 | NÃO | 'in_progress'::character varying |  |
| created_by | uuid | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| scope | jsonb | NÃO |  |  |
| end_date | timestamp with time zone | SIM |  |  |
| target_stage | integer(32,0) | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| start_date | timestamp with time zone | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### himss_emram_benchmarks

Dados de benchmark HIMSS EMRAM por região, país e tipo de estabelecimento

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| stage7_percentage | double precision(53) | SIM |  |  |
| avg_stage | double precision(53) | SIM |  |  |
| median_stage | integer(32,0) | SIM |  |  |
| sample_size | integer(32,0) | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_by | uuid | SIM |  |  |
| notes | text | SIM |  |  |
| source | character varying(255 | SIM |  |  |
| facility_type | character varying(100 | NÃO |  |  |
| region | character varying(100 | SIM |  |  |
| country | character varying(100 | NÃO |  |  |
| stage6_percentage | double precision(53) | SIM |  |  |
| stage5_percentage | double precision(53) | SIM |  |  |
| stage4_percentage | double precision(53) | SIM |  |  |
| stage3_percentage | double precision(53) | SIM |  |  |
| stage2_percentage | double precision(53) | SIM |  |  |
| stage1_percentage | double precision(53) | SIM |  |  |
| stage0_percentage | double precision(53) | SIM |  |  |
| year | integer(32,0) | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### himss_emram_certifications

Certificações HIMSS EMRAM alcançadas por organizações de saúde

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| certification_date | timestamp with time zone | NÃO |  |  |
| healthcare_facility_name | character varying(255 | NÃO |  |  |
| certificate_number | character varying(100 | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| certifying_body | character varying(255 | NÃO |  |  |
| certifying_assessor | character varying(255 | SIM |  |  |
| status | character varying(50 | NÃO | 'active'::character varying |  |
| created_by | uuid | SIM |  |  |
| notes | text | SIM |  |  |
| certificate_url | character varying(255 | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| assessment_id | uuid | NÃO |  |  |
| expiration_date | timestamp with time zone | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| stage_achieved | integer(32,0) | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
### himss_emram_criteria

Critérios para cada estágio do modelo HIMSS EMRAM

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| validation_rules | jsonb | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| is_mandatory | boolean | NÃO | true |  |
| criteria_code | character varying(50 | NÃO |  |  |
| stage_id | uuid | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
| description | text | NÃO |  |  |
| category | character varying(100 | SIM |  |  |
| is_active | boolean | NÃO | true |  |
| implementation_guidance | text | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### himss_emram_criteria_results

Resultados da avaliação de cada critério HIMSS EMRAM

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| recommendations | jsonb | SIM | '[]'::jsonb |  |
| issues_found | jsonb | SIM | '[]'::jsonb |  |
| updated_by | uuid | SIM |  |  |
| created_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| status | character varying(50 | NÃO |  |  |
| compliance_percentage | double precision(53) | SIM |  |  |
| criteria_id | uuid | NÃO |  |  |
| implementation_status | character varying(50 | SIM |  |  |
| notes | text | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| evidence | text | SIM |  |  |
| assessment_id | uuid | NÃO |  |  |
| validation_data | jsonb | SIM |  |  |
### himss_emram_stages

Estágios definidos pelo modelo HIMSS EMRAM para adoção de registros médicos eletrônicos

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| stage_number | integer(32,0) | NÃO |  |  |
| cumulative | boolean | NÃO | true |  |
| description | text | NÃO |  |  |
| name | character varying(100 | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
### hipaa_audit_view

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | SIM |  |  |
| id | uuid | SIM |  |  |
| event_time | timestamp with time zone | SIM |  |  |
| event_category | USER-DEFINED | SIM |  |  |
| severity_level | USER-DEFINED | SIM |  |  |
| user_id | uuid | SIM |  |  |
| details | jsonb | SIM |  |  |
| regulatory_references | ARRAY | SIM |  |  |
| compliance_tags | ARRAY | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| source_ip | character varying(45 | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| action | character varying(100 | SIM |  |  |
| user_email | character varying(255 | SIM |  |  |
| user_name | character varying(255 | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
### iam_metrics_history

Armazena m�tricas hist�ricas do m�dulo IAM para an�lise de tend�ncias e monitoramento

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| metric_value | numeric | NÃO |  | Valor num�rico da m�trica |
| metric_date | timestamp with time zone | NÃO | now() | Data e hora em que a m�trica foi coletada |
| id | uuid | NÃO | gen_random_uuid() |  |
| metric_name | text | NÃO |  | Nome da m�trica (ex: total_users, active_sessions, etc.) |
| created_at | timestamp with time zone | SIM | now() |  |
| metric_details | jsonb | SIM |  | Detalhes adicionais em formato JSON |
### identity_provider_attribute_mappings

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| created_at | timestamp with time zone | SIM | now() |  |
| external_attribute | character varying(255 | NÃO |  |  |
| internal_attribute | character varying(255 | NÃO |  |  |
| is_required | boolean | SIM | false |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| organization_id | uuid | NÃO |  |  |
| transformation_expression | text | SIM |  |  |
| provider_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### identity_provider_role_mappings

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| external_role | character varying(255 | NÃO |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| mapping_condition | jsonb | SIM |  |  |
| internal_role_id | uuid | NÃO |  |  |
| provider_id | uuid | NÃO |  |  |
### identity_providers

External identity providers for federation

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| client_id | character varying(255 | SIM |  |  |
| certificate | text | SIM |  |  |
| private_key | text | SIM |  |  |
| authorization_endpoint | character varying(255 | SIM |  |  |
| token_endpoint | character varying(255 | SIM |  |  |
| userinfo_endpoint | character varying(255 | SIM |  |  |
| jwks_uri | character varying(255 | SIM |  |  |
| end_session_endpoint | character varying(255 | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| config_metadata | jsonb | SIM | '{}'::jsonb |  |
| last_verified_at | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| mapping_strategy | USER-DEFINED | NÃO | 'just_in_time_provisioning'::iam.identity_mapping_strategy |  |
| status | USER-DEFINED | NÃO | 'inactive'::iam.identity_provider_status |  |
| protocol | USER-DEFINED | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
| description | text | SIM |  |  |
| issuer_url | character varying(255 | NÃO |  |  |
| metadata_url | character varying(255 | SIM |  |  |
| client_secret | text | SIM |  |  |
### iso27001_action_plans

Planos de ação para endereçar não conformidades com ISO 27001

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| status | character varying(50 | NÃO | 'open'::character varying |  |
| completed_by | uuid | SIM |  |  |
| completed_at | timestamp with time zone | SIM |  |  |
| created_by | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| assessment_id | uuid | SIM |  |  |
| due_date | date | SIM |  |  |
| description | text | SIM |  |  |
| assigned_to | uuid | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| completion_notes | text | SIM |  |  |
| estimated_effort | character varying(100 | SIM |  |  |
| title | character varying(255 | NÃO |  |  |
| control_result_id | uuid | SIM |  |  |
| priority | character varying(50 | NÃO |  |  |
| healthcare_related | boolean | SIM | false |  |
### iso27001_assessments

Avaliações de conformidade com ISO 27001 realizadas pela organização

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| framework_id | uuid | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| start_date | timestamp with time zone | NÃO |  |  |
| end_date | timestamp with time zone | SIM |  |  |
| healthcare_specific | boolean | SIM | false |  |
| score | double precision(53) | SIM |  |  |
| version | character varying(50 | SIM | '2013'::character varying |  |
| status | character varying(50 | NÃO | 'in_progress'::character varying |  |
| description | text | SIM |  |  |
| name | character varying(255 | NÃO |  |  |
| scope | jsonb | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| created_by | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
### iso27001_control_results

Resultados da avaliação de cada controle ISO 27001

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| recommendations | jsonb | SIM | '[]'::jsonb |  |
| status | character varying(50 | NÃO |  |  |
| issues_found | jsonb | SIM | '[]'::jsonb |  |
| created_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| evidence | text | SIM |  |  |
| score | double precision(53) | SIM |  |  |
| notes | text | SIM |  |  |
| implementation_status | character varying(50 | SIM |  |  |
| healthcare_specific_findings | text | SIM |  |  |
| control_id | uuid | NÃO |  |  |
| assessment_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### iso27001_controls

Controles definidos pelo padrão ISO/IEC 27001 para segurança da informação

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| section | character varying(100 | NÃO |  |  |
| description | text | NÃO |  |  |
| healthcare_applicability | text | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| implementation_guidance | text | SIM |  |  |
| category | character varying(100 | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| is_active | boolean | SIM | true |  |
| reference_links | jsonb | SIM |  |  |
| validation_rules | jsonb | SIM |  |  |
| control_id | character varying(50 | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
### iso27001_documents

Documentos relacionados à implementação e manutenção do SGSI conforme ISO 27001

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| version | character varying(50 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| file_size | bigint(64,0) | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_by | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| approved_by | uuid | SIM |  |  |
| approved_at | timestamp with time zone | SIM |  |  |
| related_controls | jsonb | SIM | '[]'::jsonb |  |
| last_review_date | timestamp with time zone | SIM |  |  |
| next_review_date | timestamp with time zone | SIM |  |  |
| healthcare_specific | boolean | SIM | false |  |
| title | character varying(255 | NÃO |  |  |
| document_type | character varying(100 | NÃO |  |  |
| description | text | SIM |  |  |
| status | character varying(50 | NÃO |  |  |
| content_url | character varying(255 | SIM |  |  |
| storage_path | character varying(255 | SIM |  |  |
| file_type | character varying(50 | SIM |  |  |
### iso27001_framework_mapping

Mapeamento entre controles ISO 27001 e outros frameworks regulatórios

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| created_by | uuid | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| iso_control_id | uuid | NÃO |  |  |
| framework_id | uuid | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| framework_control_id | character varying(100 | NÃO |  |  |
| framework_control_name | character varying(255 | SIM |  |  |
| mapping_type | character varying(50 | NÃO |  |  |
| mapping_strength | character varying(50 | SIM |  |  |
| notes | text | SIM |  |  |
### lgpd_audit_view

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_name | character varying(255 | SIM |  |  |
| id | uuid | SIM |  |  |
| compliance_tags | ARRAY | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| details | jsonb | SIM |  |  |
| source_ip | character varying(45 | SIM |  |  |
| severity_level | USER-DEFINED | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| action | character varying(100 | SIM |  |  |
| user_email | character varying(255 | SIM |  |  |
| user_name | character varying(255 | SIM |  |  |
| event_category | USER-DEFINED | SIM |  |  |
| event_time | timestamp with time zone | SIM |  |  |
| user_id | uuid | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| regulatory_references | ARRAY | SIM |  |  |
### mfa_organization_settings

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| remember_device_days | integer(32,0) | SIM | 30 |  |
| min_required_methods | integer(32,0) | SIM | 1 |  |
| allowed_methods | ARRAY | NÃO | ARRAY['totp'::iam.mfa_method_type, 'email'::iam.mfa_method_type] |  |
| required_for_all | boolean | SIM | false |  |
| custom_settings | jsonb | SIM | '{}'::jsonb |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
### mfa_sessions

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| session_token | character varying(255 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| user_id | uuid | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| verified | boolean | SIM | false |  |
| verified_method | USER-DEFINED | SIM |  |  |
| expires_at | timestamp with time zone | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| user_agent | text | SIM |  |  |
| ip_address | character varying(45 | SIM |  |  |
| challenge_token | character varying(255 | SIM |  |  |
### organizations

Armazena informações sobre as organizações que utilizam a plataforma INNOVABIZ, incluindo configurações específicas e metadados.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| compliance_settings | jsonb | SIM | '{}'::jsonb | Configurações de conformidade específicas da organização. |
| metadata | jsonb | SIM | '{}'::jsonb | Metadados adicionais da organização. |
| settings | jsonb | SIM | '{}'::jsonb | Configurações específicas da organização em formato JSON. |
| updated_at | timestamp with time zone | SIM | now() | Data e hora da última atualização do registro. |
| created_at | timestamp with time zone | SIM | now() | Data e hora de criação do registro. |
| id | uuid | NÃO | iam.uuid_generate_v4() | Identificador único da organização (UUIDv4). |
| region_code | character varying(50 | SIM |  | Código da região/estado da organização. |
| country_code | character varying(3 | SIM |  | Código do país da organização (ISO 3166-1 alpha-2). |
| sector | character varying(100 | SIM |  | Segmento específico dentro do setor. |
| industry | character varying(100 | SIM |  | Setor de atuação da organização. |
| code | character varying(50 | NÃO |  | Código único da organização (usado para referência). |
| name | character varying(255 | NÃO |  | Nome completo da organização. |
| is_active | boolean | SIM | true | Indica se a organização está ativa no sistema. |
### password_policies

Armazena as políticas de senha das organizações

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| max_age_days | integer(32,0) | SIM |  | Número máximo de dias que uma senha pode ser usada antes de expirar |
| require_special_char | boolean | NÃO | true | Se a senha deve conter caracteres especiais |
| require_number | boolean | NÃO | true | Se a senha deve conter números |
| applies_to | character varying(50 | NÃO | 'all_users'::character varying | A quem a política se aplica (todos, funções específicas, usuários específicos) |
| description | text | SIM |  | Descrição detalhada da política |
| name | character varying(255 | NÃO |  | Nome da política de senha |
| history_size | integer(32,0) | SIM | 5 | Número de senhas anteriores que não podem ser reutilizadas |
| min_length | integer(32,0) | NÃO | 8 | Tamanho mínimo da senha |
| metadata | jsonb | SIM | '{}'::jsonb | Metadados adicionais da política |
| updated_by | uuid | SIM |  | ID do último usuário que atualizou o registro |
| created_by | uuid | SIM |  | ID do usuário que criou o registro |
| id | uuid | NÃO | iam.uuid_generate_v4() | Identificador único da política de senha |
| organization_id | uuid | NÃO |  | Organização à qual a política se aplica |
| require_uppercase | boolean | NÃO | true | Se a senha deve conter letras maiúsculas |
| require_lowercase | boolean | NÃO | true | Se a senha deve conter letras minúsculas |
| is_active | boolean | NÃO | true | Indica se a política está ativa |
| lockout_duration_minutes | integer(32,0) | SIM | 30 | Duração do bloqueio da conta após exceder o número máximo de tentativas |
| created_at | timestamp with time zone | SIM | now() | Data de criação do registro |
| updated_at | timestamp with time zone | SIM | now() | Data da última atualização do registro |
| max_attempts | integer(32,0) | SIM | 5 | Número máximo de tentativas de login antes do bloqueio |
### permissions

Lista todas as permissões disponíveis no sistema, que podem ser associadas a funções.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| resource | character varying(100 | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| code | character varying(255 | NÃO |  |  |
| name | character varying(255 | NÃO |  |  |
| description | text | SIM |  |  |
| action | character varying(100 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### policy_set_policies

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| priority | integer(32,0) | NÃO | 100 |  |
| organization_id | uuid | NÃO |  |  |
| attribute_policy_id | uuid | NÃO |  |  |
| policy_set_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| created_at | timestamp with time zone | SIM | now() |  |
### policy_sets

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| name | character varying(255 | NÃO |  |  |
| priority | integer(32,0) | NÃO | 100 |  |
| evaluation_strategy | USER-DEFINED | NÃO | 'deny_overrides'::iam.policy_evaluation_strategy |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| organization_id | uuid | NÃO |  |  |
| description | text | SIM |  |  |
| is_active | boolean | SIM | true |  |
### regulatory_frameworks

Lista os frameworks regulatórios suportados pelo sistema (ex: GDPR, LGPD, HIPAA).

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| code | character varying(100 | NÃO |  |  |
| sector | character varying(100 | SIM |  |  |
| version | character varying(50 | SIM |  |  |
| description | text | SIM |  |  |
| effective_date | date | SIM |  |  |
| region | character varying(100 | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| is_active | boolean | SIM | true |  |
| name | character varying(255 | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
### role_permissions

Tabela de associação entre Roles e Permissions, definindo quais permissões cada role concede dentro de uma organização.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| permission_id | uuid | NÃO |  | Identificador da permissão atribuída à role. |
| organization_id | uuid | NÃO |  | Identificador da organização à qual esta atribuição de permissão de role se aplica. |
| role_id | uuid | NÃO |  | Identificador da role. |
| assigned_at | timestamp with time zone | SIM | now() | Timestamp de quando a permissão foi efetivamente atribuída à role. |
| assigned_by | uuid | SIM |  | Identificador do usuário que realizou a atribuição da permissão à role. |
| created_at | timestamp with time zone | SIM | now() | Timestamp da criação do registo de atribuição. |
| updated_at | timestamp with time zone | SIM | now() | Timestamp da última atualização do registo de atribuição. |
| metadata | jsonb | SIM | '{}'::jsonb |  |
### role_policy_sets

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| policy_set_id | uuid | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| role_id | uuid | NÃO |  |  |
### roles

Define os papéis que podem ser atribuídos aos usuários, agrupando conjuntos de permissões.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| created_at | timestamp with time zone | SIM | now() |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_by | uuid | SIM |  |  |
| is_system_role | boolean | SIM | false |  |
| organization_id | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| name | character varying(100 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| description | text | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
### schema_migrations

Registra todas as migra��es de esquema aplicadas ao banco de dados

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| metadata | jsonb | SIM |  | Metadados adicionais sobre a migra��o |
| checksum | character varying(64 | SIM |  | Hash SHA-256 do script de migra��o para verifica��o de integridade |
| installed_by | character varying(100 | NÃO |  | Usu�rio que aplicou a migra��o |
| script_name | character varying(255 | NÃO |  | Nome do script de migra��o |
| description | text | NÃO |  | Descri��o da migra��o |
| id | integer(32,0) | NÃO | nextval('iam.schema_migrations_id_seq'::regclass) |  |
| installed_on | timestamp with time zone | NÃO | now() | Data e hora em que a migra��o foi aplicada |
| execution_time_ms | bigint(64,0) | SIM |  | Tempo de execu��o da migra��o em milissegundos |
| success | boolean | NÃO |  | Indica se a migra��o foi aplicada com sucesso |
| version | character varying(50 | NÃO |  | Identificador �nico da vers�o da migra��o (formato: YYYY.MM.DD.HHMM) |
### security_policies

Armazena as políticas de segurança configuráveis para cada organização.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| description | text | SIM |  |  |
| name | character varying(255 | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| settings | jsonb | NÃO |  |  |
| organization_id | uuid | SIM |  |  |
| created_by | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| is_active | boolean | SIM | true |  |
| policy_type | character varying(100 | NÃO |  |  |
### sessions

Registra as sessões ativas dos usuários, incluindo tokens e informações de autenticação.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| expires_at | timestamp with time zone | NÃO |  |  |
| last_activity | timestamp with time zone | SIM | now() |  |
| user_id | uuid | NÃO |  |  |
| user_agent | text | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| is_active | boolean | SIM | true |  |
| ip_address | character varying(50 | SIM |  |  |
| token | character varying(255 | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### trusted_devices

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | NÃO |  |  |
| user_id | uuid | NÃO |  |  |
| device_name | character varying(255 | SIM |  |  |
| device_identifier | character varying(255 | NÃO |  |  |
| ip_address | character varying(45 | SIM |  |  |
| user_agent | text | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| revoked_at | timestamp with time zone | SIM |  |  |
| revoked | boolean | SIM | false |  |
| expires_at | timestamp with time zone | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| last_used | timestamp with time zone | SIM | now() |  |
### user_ar_gaze_auth

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| metadata | jsonb | SIM | '{}'::jsonb |  |
| last_used | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| created_at | timestamp with time zone | SIM | now() |  |
| pattern_data | bytea | NÃO |  |  |
| gaze_type | USER-DEFINED | NÃO |  |  |
| mfa_method_id | uuid | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
| user_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| pattern_name | character varying(100 | NÃO |  |  |
| pattern_hash | character varying(255 | NÃO |  |  |
| status | USER-DEFINED | SIM | 'enabled'::iam.mfa_status |  |
### user_ar_gesture_auth

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| created_at | timestamp with time zone | SIM | now() |  |
| complexity_score | integer(32,0) | NÃO |  |  |
| gesture_hash | character varying(255 | NÃO |  |  |
| gesture_data | bytea | NÃO |  |  |
| last_used | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| gesture_type | USER-DEFINED | NÃO |  |  |
| mfa_method_id | uuid | SIM |  |  |
| organization_id | uuid | NÃO |  |  |
| user_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| status | USER-DEFINED | SIM | 'enabled'::iam.mfa_status |  |
| gesture_name | character varying(100 | NÃO |  |  |
### user_ar_spatial_password

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| password_data | bytea | NÃO |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| user_id | uuid | NÃO |  |  |
| organization_id | uuid | NÃO |  |  |
| password_name | character varying(100 | NÃO |  |  |
| password_hash | character varying(255 | NÃO |  |  |
| mfa_method_id | uuid | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| complexity_score | integer(32,0) | NÃO |  |  |
| status | USER-DEFINED | SIM | 'enabled'::iam.mfa_status |  |
| last_used | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| dimension_count | integer(32,0) | NÃO |  |  |
### user_mfa_backup_codes

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | NÃO |  |  |
| code_hash | character varying(255 | NÃO |  |  |
| expires_at | timestamp with time zone | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| used_at | timestamp with time zone | SIM |  |  |
| used | boolean | SIM | false |  |
| user_id | uuid | NÃO |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
### user_mfa_methods

Sem descrição

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| metadata | jsonb | SIM | '{}'::jsonb |  |
| user_id | uuid | NÃO |  |  |
| email | character varying(255 | SIM |  |  |
| created_at | timestamp with time zone | SIM | now() |  |
| organization_id | uuid | NÃO |  |  |
| method_type | USER-DEFINED | NÃO |  |  |
| status | USER-DEFINED | NÃO | 'pending_activation'::iam.mfa_status |  |
| updated_at | timestamp with time zone | SIM | now() |  |
| phone_number | character varying(50 | SIM |  |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| secret | text | SIM |  |  |
| name | character varying(100 | SIM |  |  |
| last_used | timestamp with time zone | SIM |  |  |
### user_roles

Relaciona usuários a funções, permitindo herdar permissões.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| role_id | uuid | NÃO |  |  |
| is_active | boolean | SIM | true |  |
| id | uuid | NÃO | iam.uuid_generate_v4() |  |
| assigned_by | uuid | SIM |  |  |
| user_id | uuid | NÃO |  |  |
| metadata | jsonb | SIM | '{}'::jsonb |  |
| expires_at | timestamp with time zone | SIM |  |  |
| granted_at | timestamp with time zone | SIM | now() |  |
### users

Contém as informações dos usuários do sistema, incluindo credenciais, status e preferências.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | NÃO | iam.uuid_generate_v4() | Identificador único do usuário (UUIDv4). |
| status | character varying(50 | NÃO | 'active'::character varying | Status atual do usuário (active, inactive, suspended, locked). |
| password_hash | character varying(255 | NÃO |  | Hash da senha do usuário (usando criptografia forte). |
| full_name | character varying(255 | NÃO |  | Nome completo do usuário. |
| email | character varying(255 | NÃO |  | Endereço de e-mail do usuário (deve ser único). |
| username | character varying(255 | NÃO |  | Nome de usuário único para login. |
| organization_id | uuid | SIM |  | Referência à organização à qual o usuário pertence. |
| created_at | timestamp with time zone | SIM | now() | Data e hora de criação do registro. |
| updated_at | timestamp with time zone | SIM | now() | Data e hora da última atualização do registro. |
| last_login | timestamp with time zone | SIM |  | Data e hora do último login bem-sucedido. |
| created_by | uuid | SIM |  |  |
| updated_by | uuid | SIM |  |  |
| metadata | jsonb | SIM | '{}'::jsonb | Metadados adicionais do usuário. |
| preferences | jsonb | SIM | '{}'::jsonb | Preferências do usuário em formato JSON. |
### vw_audit_log_summary

Resumo diário de eventos de auditoria dos últimos 90 dias

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| action | character varying(100 | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| event_count | bigint(64,0) | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| log_date | timestamp with time zone | SIM |  |  |
### vw_compliance_validators_details

Detalhes dos validadores de compliance configurados no sistema

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| validator_name | character varying(255 | SIM |  |  |
| framework_id | uuid | SIM |  |  |
| is_active | boolean | SIM |  |  |
| validator_id | uuid | SIM |  |  |
| framework_code | character varying(100 | SIM |  |  |
| framework_name | character varying(255 | SIM |  |  |
| validator_code | character varying(100 | SIM |  |  |
| region | character varying(100 | SIM |  |  |
| sector | character varying(100 | SIM |  |  |
| validator_version | character varying(50 | SIM |  |  |
| validator_class | character varying(255 | SIM |  |  |
| description | text | SIM |  |  |
### vw_iso27001_action_plans

Visão detalhada dos planos de ação para compliance com ISO 27001

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| completed_at | timestamp with time zone | SIM |  |  |
| healthcare_specific | boolean | SIM |  |  |
| days_overdue | integer(32,0) | SIM |  |  |
| is_overdue | boolean | SIM |  |  |
| healthcare_related | boolean | SIM |  |  |
| completed_by | uuid | SIM |  |  |
| updated_at | timestamp with time zone | SIM |  |  |
| created_at | timestamp with time zone | SIM |  |  |
| created_by | uuid | SIM |  |  |
| assigned_to | uuid | SIM |  |  |
| due_date | date | SIM |  |  |
| control_result_id | uuid | SIM |  |  |
| assessment_id | uuid | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| id | uuid | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
| assessment_name | character varying(255 | SIM |  |  |
| control_status | character varying(50 | SIM |  |  |
| control_code | character varying(50 | SIM |  |  |
| control_name | character varying(255 | SIM |  |  |
| title | character varying(255 | SIM |  |  |
| description | text | SIM |  |  |
| priority | character varying(50 | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| assigned_to_name | character varying(255 | SIM |  |  |
| created_by_name | character varying(255 | SIM |  |  |
| completed_by_name | character varying(255 | SIM |  |  |
| completion_notes | text | SIM |  |  |
| estimated_effort | character varying(100 | SIM |  |  |
### vw_iso27001_control_results_detail

Detalhes dos resultados de avaliação de controles ISO 27001

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_name | character varying(255 | SIM |  |  |
| assessment_name | character varying(255 | SIM |  |  |
| healthcare_specific | boolean | SIM |  |  |
| updated_at | timestamp with time zone | SIM |  |  |
| created_at | timestamp with time zone | SIM |  |  |
| score | double precision(53) | SIM |  |  |
| issues_found | jsonb | SIM |  |  |
| category | character varying(100 | SIM |  |  |
| action_plans_count | bigint(64,0) | SIM |  |  |
| open_action_plans_count | bigint(64,0) | SIM |  |  |
| control_id | uuid | SIM |  |  |
| assessed_by | character varying(255 | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| healthcare_specific_findings | text | SIM |  |  |
| assessment_id | uuid | SIM |  |  |
| id | uuid | SIM |  |  |
| notes | text | SIM |  |  |
| evidence | text | SIM |  |  |
| implementation_status | character varying(50 | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| recommendations | jsonb | SIM |  |  |
| healthcare_applicability | text | SIM |  |  |
| control_description | text | SIM |  |  |
| control_name | character varying(255 | SIM |  |  |
| section | character varying(100 | SIM |  |  |
| control_code | character varying(50 | SIM |  |  |
### vw_iso27001_controls_detail

Visão detalhada dos controles ISO 27001 com mapeamentos para outros frameworks

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| id | uuid | SIM |  |  |
| validation_rules | jsonb | SIM |  |  |
| control_id | character varying(50 | SIM |  |  |
| section | character varying(100 | SIM |  |  |
| name | character varying(255 | SIM |  |  |
| description | text | SIM |  |  |
| healthcare_applicability | text | SIM |  |  |
| implementation_guidance | text | SIM |  |  |
| category | character varying(100 | SIM |  |  |
| reference_links | jsonb | SIM |  |  |
| is_active | boolean | SIM |  |  |
| created_at | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM |  |  |
| framework_mappings_count | bigint(64,0) | SIM |  |  |
| mapped_frameworks | ARRAY | SIM |  |  |
### vw_iso27001_framework_mappings

Mapeamentos entre controles ISO 27001 e outros frameworks regulatórios

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| mapping_type | character varying(50 | SIM |  |  |
| mapping_strength | character varying(50 | SIM |  |  |
| notes | text | SIM |  |  |
| healthcare_applicability | text | SIM |  |  |
| id | uuid | SIM |  |  |
| iso_control_id | uuid | SIM |  |  |
| framework_id | uuid | SIM |  |  |
| created_at | timestamp with time zone | SIM |  |  |
| updated_at | timestamp with time zone | SIM |  |  |
| created_by_name | character varying(255 | SIM |  |  |
| iso_control_code | character varying(50 | SIM |  |  |
| iso_control_name | character varying(255 | SIM |  |  |
| iso_section | character varying(100 | SIM |  |  |
| framework_code | character varying(100 | SIM |  |  |
| framework_name | character varying(255 | SIM |  |  |
| framework_control_id | character varying(100 | SIM |  |  |
| framework_control_name | character varying(255 | SIM |  |  |
### vw_regulatory_frameworks_summary

Resumo dos frameworks regulatórios e contagem de validadores associados

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| sector | character varying(100 | SIM |  |  |
| name | character varying(255 | SIM |  |  |
| code | character varying(100 | SIM |  |  |
| id | uuid | SIM |  |  |
| validators_count | bigint(64,0) | SIM |  |  |
| is_active | boolean | SIM |  |  |
| effective_date | date | SIM |  |  |
| version | character varying(50 | SIM |  |  |
| region | character varying(100 | SIM |  |  |
### vw_role_usage_stats

Estatísticas de uso de roles por organização

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_id | uuid | SIM |  |  |
| assigned_users_count | bigint(64,0) | SIM |  |  |
| is_system_role | boolean | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
| role_name | character varying(100 | SIM |  |  |
| last_assignment | timestamp with time zone | SIM |  |  |
| first_assignment | timestamp with time zone | SIM |  |  |
| permissions_count | bigint(64,0) | SIM |  |  |
| role_id | uuid | SIM |  |  |
### vw_security_audit

Visão para auditoria de segurança, mostrando ações realizadas no sistema.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_name | character varying(255 | SIM |  |  |
| details | jsonb | SIM |  |  |
| timestamp | timestamp with time zone | SIM |  |  |
| id | uuid | SIM |  |  |
| username | character varying(255 | SIM |  |  |
| email | character varying(255 | SIM |  |  |
| action | character varying(100 | SIM |  |  |
| resource_type | character varying(100 | SIM |  |  |
| resource_id | character varying(255 | SIM |  |  |
| status | character varying(50 | SIM |  |  |
| ip_address | character varying(50 | SIM |  |  |
### vw_security_policies_by_organization

Visão de políticas de segurança configuradas por organização

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| created_at | timestamp with time zone | SIM |  |  |
| is_active | boolean | SIM |  |  |
| updated_at | timestamp with time zone | SIM |  |  |
| settings | jsonb | SIM |  |  |
| country_code | character varying(3 | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| policy_id | uuid | SIM |  |  |
| industry | character varying(100 | SIM |  |  |
| updated_by_user | character varying(255 | SIM |  |  |
| created_by_user | character varying(255 | SIM |  |  |
| policy_type | character varying(100 | SIM |  |  |
| policy_name | character varying(255 | SIM |  |  |
| region_code | character varying(50 | SIM |  |  |
| sector | character varying(100 | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
### vw_session_activity

Visão que mostra a atividade de sessões dos últimos 30 dias

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| user_agent | text | SIM |  |  |
| session_duration_hours | numeric | SIM |  |  |
| expires_at | timestamp with time zone | SIM |  |  |
| last_activity | timestamp with time zone | SIM |  |  |
| session_start | timestamp with time zone | SIM |  |  |
| session_id | uuid | SIM |  |  |
| user_id | uuid | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| session_status | text | SIM |  |  |
| ip_address | character varying(50 | SIM |  |  |
| email | character varying(255 | SIM |  |  |
| username | character varying(255 | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
### vw_user_permissions

Visão que mostra todas as permissões de cada usuário ativo no sistema

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| email | character varying(255 | SIM |  |  |
| organization_name | character varying(255 | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| user_id | uuid | SIM |  |  |
| username | character varying(255 | SIM |  |  |
| role_id | uuid | SIM |  |  |
| permission_id | uuid | SIM |  |  |
| role_name | character varying(100 | SIM |  |  |
| permission_code | character varying(255 | SIM |  |  |
| permission_name | character varying(255 | SIM |  |  |
| resource | character varying(100 | SIM |  |  |
| action | character varying(100 | SIM |  |  |
| permission_assigned_at | timestamp with time zone | SIM |  |  |
### vw_user_roles

Visão que lista todos os usuários com suas respectivas funções atribuídas.

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| user_id | uuid | SIM |  |  |
| role_assignment_active | boolean | SIM |  |  |
| expires_at | timestamp with time zone | SIM |  |  |
| granted_at | timestamp with time zone | SIM |  |  |
| role_id | uuid | SIM |  |  |
| user_status | character varying(50 | SIM |  |  |
| full_name | character varying(255 | SIM |  |  |
| role_description | text | SIM |  |  |
| email | character varying(255 | SIM |  |  |
| role_name | character varying(100 | SIM |  |  |
| username | character varying(255 | SIM |  |  |
### vw_user_status_by_organization

Visão que mostra métricas de status de usuários por organização

| Coluna | Tipo | Nulo | Padrão | Descrição |\n|--------|------|------|--------|-----------|\n| organization_name | character varying(255 | SIM |  |  |
| active_users | bigint(64,0) | SIM |  |  |
| total_users | bigint(64,0) | SIM |  |  |
| organization_id | uuid | SIM |  |  |
| industry | character varying(100 | SIM |  |  |
| sector | character varying(100 | SIM |  |  |
| active_percentage | numeric | SIM |  |  |
| locked_users | bigint(64,0) | SIM |  |  |
| country_code | character varying(3 | SIM |  |  |
| suspended_users | bigint(64,0) | SIM |  |  |
| region_code | character varying(50 | SIM |  |  |
| inactive_users | bigint(64,0) | SIM |  |  |
