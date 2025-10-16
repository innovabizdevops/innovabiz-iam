# Innovabiz Data Governance Suite – Documentação Multilíngue

## Visão Geral
Este pacote cobre a modelagem, automação e integração de todos os domínios críticos da plataforma InnovaBiz, alinhando-se com normas, frameworks e melhores práticas internacionais para governança de dados, compliance, risco, processos, produtos, serviços e referência geográfica. Todos os campos estão em inglês, com descrições multilíngues.

---

## Estrutura dos Módulos e Scripts
- **Organizational Domains:** Estruturas, empresas, acionistas, clientes, departamentos, comitês.
- **Business Domains:** Modelos, planos, estratégias, mercados, segmentos.
- **Products & Services:** Categorias, tipos, produtos, serviços, distribuição, canais.
- **Compliance:** Normas, frameworks, regulamentos, legislações e suas relações.
- **Risk:** Categorias, tipos, níveis, planos, impactos, consequências.
- **Processes:** Tipos, processos, procedimentos, atividades.
- **Reference:** Continentes, países, estados, cidades, distritos, bairros.
- **ESG:** Indicadores ambientais, sociais e de governança, relatórios por empresa e ano, alinhados a GRI, SASB, ISO 14001, ISO 26000.
- **Supply Chain:** Fornecedores, contratos logísticos, rastreabilidade de lotes, auditorias de compliance, conforme ISO 28000, GS1, SCOR.
- **HR (Recursos Humanos):** Colaboradores, cargos, competências, matrizes, indicadores de RH e relatórios (People Analytics, ISO 30414).
- **Governance Automation:** Triggers, funções, views para auditoria, compliance e BI.

---

## Padrões e Normas Adotadas
- **Data Governance:** DAMA-DMBOK, ISO/IEC 11179, ISO 8000, DMBOK, TOGAF
- **Compliance:** ISO 37301, ISO 27001, GDPR, ITIL, COBIT
- **Risco:** ISO 31000, COSO, Basel II/III
- **Processos:** BPMN, ITIL, ISO 9001, ISO 56000
- **Referência Geográfica:** ISO 3166, UN/LOCODE
- **Produtos/Serviços:** ISO 9001, BIAN

---

## Exemplos de Queries de Governança e BI
```sql
-- Empresas sem compliance ou responsável
SELECT * FROM bi.vw_companies_missing_compliance;

-- Riscos críticos
SELECT * FROM bi.vw_risk_critical;

-- Processos sem responsável
SELECT * FROM processes.process WHERE responsible IS NULL OR responsible = '';

-- Produtos/Serviços sem compliance
SELECT p.* FROM products.product p LEFT JOIN compliance.product_service_compliance c ON p.id = c.product_id WHERE c.id IS NULL;

-- Empresas sem relatório ESG no ano
SELECT * FROM bi.vw_esg_companies_missing_reports;

-- Fornecedores sem auditoria recente
SELECT * FROM bi.vw_suppliers_missing_audit;

-- Colaboradores sem indicadores de RH reportados
SELECT * FROM bi.vw_employees_missing_indicators;
```

---

## Checklist de Governança
- [x] Todos os domínios modelados com metadados, compliance, privacidade e rastreabilidade (incluindo ESG, Supply Chain e RH)
- [x] Triggers de auditoria, compliance e atualização de timestamps implementadas em todos os módulos
- [x] Views integradas para BI e auditoria (com dashboards para ESG, Supply Chain, RH)
- [x] Documentação multilíngue e comentários em todos os campos/tabelas
- [x] Scripts versionados e prontos para execução e integração
- [x] Estrutura pronta para expansão modular e integração externa (ERP, APIs)

---

## Como Executar
1. Execute os scripts de migração na ordem dos módulos (organization, business, products/services, compliance, risk, processes, reference, ESG, supply_chain, hr, governance automation, automation_bi_views_esg_supplychain_hr).
2. Valide a criação dos triggers, views e funções em todos os domínios.
3. Integre as views BI às ferramentas de análise (Power BI, Metabase, Superset), incluindo dashboards para ESG, Supply Chain e RH.
4. Consulte as queries e checklist para governança contínua e auditoria.
5. Prepare integração externa conforme necessidade (ERP, APIs públicas, etc).

---

## Observações
- Todos os campos estão em inglês, descrições multilíngues via COMMENT ON COLUMN.
- Adapte e expanda os módulos conforme necessidades futuras.
- Consulte sempre as normas e frameworks de referência para garantir aderência regulatória.

---

**Dúvidas ou necessidade de expansão? Consulte a equipe de governança de dados ou o responsável técnico.**
