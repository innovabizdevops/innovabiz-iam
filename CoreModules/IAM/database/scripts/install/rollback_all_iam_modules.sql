-- INNOVABIZ - IAM
-- Script Mestre de Rollback de Todos os Domínios
-- Data: 10/06/2025
-- Remove sequencialmente todos os objetos dos domínios do IAM

\echo 'Iniciando rollback completo do IAM (todos os domínios)...'

-- Exemplo de DROP em ordem inversa à criação (customize conforme dependências reais)
-- Métodos de Autenticação
DROP SCHEMA IF EXISTS metodos_autenticacoes CASCADE;
-- Metadata
DROP SCHEMA IF EXISTS metadata CASCADE;
-- ISO
DROP SCHEMA IF EXISTS iso CASCADE;
-- HIMSS
DROP SCHEMA IF EXISTS himss CASCADE;
-- Healthcare
DROP SCHEMA IF EXISTS healthcare CASCADE;
-- Fix
DROP SCHEMA IF EXISTS fix CASCADE;
-- Federation
DROP SCHEMA IF EXISTS federation CASCADE;
-- Test
DROP SCHEMA IF EXISTS test CASCADE;
-- Multi-Tenant
DROP SCHEMA IF EXISTS multi_tenant CASCADE;
-- Monitoring
DROP SCHEMA IF EXISTS monitoring CASCADE;
-- Analytics
DROP SCHEMA IF EXISTS analytics CASCADE;
-- Compliance
DROP SCHEMA IF EXISTS compliance CASCADE;
-- Core
DROP SCHEMA IF EXISTS iam CASCADE;

\echo 'Rollback completo de todos os domínios IAM finalizado!'
