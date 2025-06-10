-- INNOVABIZ - IAM
-- Script Mestre de Instalação de Todos os Domínios
-- Data: 10/06/2025
-- Garante a instalação sequencial e completa de todos os domínios e módulos do IAM

\echo 'Iniciando instalação completa do IAM (todos os domínios)...'

-- Extensões e Schema Base
\i 00_install_extensions.sql
\i 00_install_iam_module.sql

-- Núcleo Core
\i ../core/01_schema_iam_core.sql
\i ../core/02_views_iam_core.sql
\i ../core/03_functions_iam_core.sql
\i ../core/04_triggers_iam_core.sql
\i ../core/05_mfa_authentication_part1.sql
\i ../core/05_mfa_authentication_part2.sql
\i ../core/05_mfa_authentication_ar_integration.sql
\i ../core/06_authorization_engine_part1.sql
\i ../core/06_authorization_engine_part2.sql

-- Compliance
\i ../compliance/01_schema_compliance.sql
\i ../compliance/02_views_compliance.sql
\i ../compliance/03_functions_compliance.sql
\i ../compliance/04_triggers_compliance.sql

-- Analytics
\i ../analytics/01_schema_analytics.sql
\i ../analytics/02_views_analytics.sql
\i ../analytics/03_functions_analytics.sql
\i ../analytics/04_triggers_analytics.sql

-- Monitoring
\i ../monitoring/01_schema_monitoring.sql
\i ../monitoring/02_views_monitoring.sql
\i ../monitoring/03_functions_monitoring.sql
\i ../monitoring/04_triggers_monitoring.sql

-- Multi-Tenant
\i ../multi_tenant/01_schema_multi_tenant.sql
\i ../multi_tenant/02_views_multi_tenant.sql
\i ../multi_tenant/03_functions_multi_tenant.sql
\i ../multi_tenant/04_triggers_multi_tenant.sql

-- Testes
\i ../test/01_schema_test.sql
\i ../test/02_views_test.sql
\i ../test/03_functions_test.sql
\i ../test/04_triggers_test.sql

-- Federation
\i ../federation/01_schema_federation.sql
\i ../federation/02_views_federation.sql
\i ../federation/03_functions_federation.sql
\i ../federation/04_triggers_federation.sql

-- Fix
\i ../fix/01_schema_fix.sql
\i ../fix/02_views_fix.sql
\i ../fix/03_functions_fix.sql
\i ../fix/04_triggers_fix.sql

-- Healthcare
\i ../healthcare/01_schema_healthcare_compliance.sql
\i ../healthcare/02_views_healthcare_compliance.sql
\i ../healthcare/03_functions_healthcare_compliance.sql
\i ../healthcare/04_triggers_healthcare_compliance.sql

-- HIMSS
\i ../himss/01_schema_himss_emram.sql
\i ../himss/04_triggers_himss_emram.sql

-- ISO
\i ../iso/01_schema_iso27001.sql
\i ../iso/02_views_iso27001.sql
\i ../iso/03_functions_iso27001_part1.sql
\i ../iso/03_functions_iso27001_part2.sql
\i ../iso/04_triggers_iso27001.sql

-- Metadata
\i ../metadata/01_add_metadata_to_iam_tables.sql
\i ../metadata/02_generate_iam_documentation.sql
\i ../metadata/03_generate_iam_compliance_report.sql
\i ../metadata/04_generate_iam_metrics.sql
\i ../metadata/05_iam_schema_migration_manager.sql
\i ../metadata/06_generate_iam_schema_documentation.sql
\i ../metadata/07_iam_performance_analysis.sql

-- Métodos de Autenticação (exemplo de inclusão, adaptar conforme necessidade)
\i ../metodos_autenticacoes/01_iam_core_schema.sql
\i ../metodos_autenticacoes/02_auth_methods_data.sql
\i ../metodos_autenticacoes/03_auth_policies_config.sql

-- Adicione aqui outros scripts de domínios ou integrações futuras

\echo 'Instalação completa de todos os domínios IAM finalizada!'
