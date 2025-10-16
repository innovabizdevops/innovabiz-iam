# Orientações para Integração Externa – Plataforma InnovaBiz

## 1. Integração com ERP (SAP, Oracle, Totvs, etc)
- Estruturar tabelas de staging (ex: staging.erp_company, staging.erp_supplier) para importação de dados externos.
- Utilizar campos de referência externa (external_ref, source_system) em todas as tabelas críticas.
- Automatizar ETL com ferramentas como Apache NiFi, Talend, Pentaho ou scripts customizados (Python, SQL).
- Garantir logs de auditoria para importações e atualizações em lote.

## 2. Integração com APIs Públicas (IBGE, ESG Ratings, etc)
- Criar tabelas de staging para dados externos (ex: staging.ibge_city, staging.esg_rating).
- Utilizar scripts Python ou ferramentas de integração para consumir APIs e popular staging tables.
- Validar e normalizar dados antes de inserir nas tabelas principais.
- Documentar endpoints, cronogramas de atualização e responsáveis.

## 3. Boas Práticas de Integração
- Manter versionamento dos scripts e pipelines de integração.
- Utilizar campos de rastreabilidade (data_source, external_ref, updated_by).
- Validar integridade referencial após cada carga.
- Documentar todo fluxo de integração e manter checklist atualizado.

## 4. Segurança e Compliance
- Garantir autenticação e criptografia em integrações externas.
- Monitorar acessos e alterações via triggers/auditoria.
- Seguir normas LGPD/GDPR para dados sensíveis.

## 5. Exemplos de Queries para Integração
```sql
-- Importar dados de staging para tabela principal
INSERT INTO organization.company (name, external_ref, created_at)
SELECT name, erp_id, now()
FROM staging.erp_company
WHERE NOT EXISTS (
  SELECT 1 FROM organization.company c WHERE c.external_ref = staging.erp_company.erp_id
);

-- Atualizar dados de ESG ratings vindos de API
UPDATE esg.company_esg_report r
SET value = s.rating, updated_at = now()
FROM staging.esg_rating s
WHERE r.company_id = s.company_id AND r.year = s.year;
```
