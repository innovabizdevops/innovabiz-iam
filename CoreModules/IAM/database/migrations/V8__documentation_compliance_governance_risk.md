# Dicionário de Dados – Innovabiz (Compliance, Governança e Risco)

## Schema: compliance

### norma
- **id**: UUID, PK – Identificador único
- **codigo**: Código da norma
- **nome**: Nome da norma
- **descricao**: Descrição detalhada
- **tipo_norma**: Tipo (Norma, Decreto, Aviso, etc)
- **orgao_emissor**: Órgão emissor
- **data_publicacao**: Data de publicação
- **status**: Vigente, revogada, etc
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### framework
- **id**: UUID, PK
- **nome**: Nome do framework
- **descricao**: Descrição
- **categoria**: Governança, Risco, Dados, etc
- **status**: Ativo, inativo
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### norma_framework
- **id**: UUID, PK
- **norma_id**: FK para norma
- **framework_id**: FK para framework

### consentimento
- **id**: UUID, PK
- **usuario_id**: FK para usuário
- **tipo_consentimento**: Tipo de consentimento
- **dado_referente**: Dado ao qual se refere
- **consentido**: Booleano
- **data_consentimento**: Data
- **revogado_em**: Data de revogação

### audit_log
- **id**: UUID, PK
- **entidade**: Nome da entidade
- **entidade_id**: UUID da entidade
- **acao**: Ação realizada
- **usuario_id**: FK para usuário
- **data_acao**: Data da ação
- **valor_anterior**: JSONB
- **valor_novo**: JSONB
- **origem**: Origem da ação

---

## Schema: governance

### tipo_orgao_social
- **id**: UUID, PK
- **nome**: Nome do tipo
- **descricao**: Descrição

### orgao_social
- **id**: UUID, PK
- **nome**: Nome do órgão
- **tipo_orgao_social_id**: FK para tipo_orgao_social

### papel
- **id**: UUID, PK
- **nome**: Nome do papel
- **descricao**: Descrição

### responsabilidade
- **id**: UUID, PK
- **papel_id**: FK para papel
- **entidade**: Nome da entidade
- **entidade_id**: UUID da entidade

---

## Schema: risk

### tipo_risco
- **id**: UUID, PK
- **nome**: Nome do tipo de risco
- **categoria**: Categoria do risco

### risco
- **id**: UUID, PK
- **nome**: Nome do risco
- **tipo_risco_id**: FK para tipo_risco
- **impacto**: Impacto do risco
- **probabilidade**: Probabilidade
- **status**: Status do risco

### plano_mitigacao
- **id**: UUID, PK
- **risco_id**: FK para risco
- **descricao**: Descrição do plano
- **responsavel_id**: UUID do responsável
- **prazo**: Prazo para mitigação
- **status**: Status do plano

---

## Schema: contracts

### contract
- **id**: UUID, PK
- **numero**: Número do contrato
- **descricao**: Descrição do contrato
- **parte_a_id**: UUID da parte A (FK para organizations ou clientes)
- **parte_b_id**: UUID da parte B (FK para organizations ou fornecedores)
- **data_inicio**: Data de início
- **data_fim**: Data de término
- **status**: Status do contrato (ativo, encerrado, etc)
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### attachment
- **id**: UUID, PK
- **contract_id**: FK para contract
- **filename**: Nome do arquivo
- **file_url**: URL do arquivo
- **uploaded_at**: Data/hora do upload

---

## Schema: processes

### process
- **id**: UUID, PK
- **nome**: Nome do processo
- **descricao**: Descrição do processo
- **status**: Status do processo (ativo, encerrado, etc)
- **responsavel_id**: UUID do responsável (FK para organizations ou users)
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### step
- **id**: UUID, PK
- **process_id**: FK para process
- **nome**: Nome da etapa
- **ordem**: Ordem da etapa
- **descricao**: Descrição da etapa
- **responsavel_id**: UUID do responsável (FK para organizations ou users)
- **status**: Status da etapa (pendente, concluída, etc)
- **inicio_previsto**: Data prevista de início
- **fim_previsto**: Data prevista de fim
- **inicio_real**: Data real de início
- **fim_real**: Data real de fim
- **created_at**: Data de criação
- **updated_at**: Data de atualização

---

## Schema: kpis

### indicator
- **id**: UUID, PK
- **nome**: Nome do indicador
- **descricao**: Descrição do indicador
- **unidade**: Unidade de medida (pontos, %, BRL, etc)
- **tipo**: Tipo do indicador (desempenho, financeiro, operacional, etc)
- **periodicidade**: Periodicidade de medição (mensal, trimestral, etc)
- **responsavel_id**: UUID do responsável (FK para organizations ou users)
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### target
- **id**: UUID, PK
- **indicator_id**: FK para indicator
- **valor**: Valor/meta definida
- **data_inicio**: Data de início da meta
- **data_fim**: Data de término da meta
- **status**: Status da meta (ativo, encerrado, etc)
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### result
- **id**: UUID, PK
- **indicator_id**: FK para indicator
- **valor**: Valor mensurado
- **data**: Data do resultado
- **observacao**: Observação sobre o resultado
- **created_at**: Data de criação
- **updated_at**: Data de atualização

---

## Schema: geographies

### country
- **id**: UUID, PK
- **nome**: Nome do país
- **codigo_iso**: Código ISO do país (ex: BR, PT, US)
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### province
- **id**: UUID, PK
- **nome**: Nome da província
- **country_id**: FK para country
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### state
- **id**: UUID, PK
- **nome**: Nome do estado
- **codigo_uf**: Código da UF (opcional)
- **country_id**: FK para country
- **province_id**: FK para province
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### municipality
- **id**: UUID, PK
- **nome**: Nome do município
- **state_id**: FK para state
- **province_id**: FK para province
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### district
- **id**: UUID, PK
- **nome**: Nome do distrito
- **municipality_id**: FK para municipality
- **state_id**: FK para state
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### county (concelho)
- **id**: UUID, PK
- **nome**: Nome do conselho/concelho
- **district_id**: FK para district
- **municipality_id**: FK para municipality
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### parish (freguesia)
- **id**: UUID, PK
- **nome**: Nome da freguesia
- **county_id**: FK para county
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### commune (comuna)
- **id**: UUID, PK
- **nome**: Nome da comuna
- **municipality_id**: FK para municipality
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### neighborhood (bairro)
- **id**: UUID, PK
- **nome**: Nome do bairro
- **district_id**: FK para district
- **municipality_id**: FK para municipality
- **commune_id**: FK para commune
- **created_at**: Data de criação
- **updated_at**: Data de atualização

### city
- **id**: UUID, PK
- **nome**: Nome da cidade
- **state_id**: FK para state
- **municipality_id**: FK para municipality
- **district_id**: FK para district
- **created_at**: Data de criação
- **updated_at**: Data de atualização

---

> Este dicionário de dados deve ser mantido atualizado a cada evolução do modelo. Recomenda-se complementar com ERD visual e exemplos de queries para auditoria e compliance.

## Queries de Validação e Compliance

## Exemplos de Relatórios Integrados e Consultas Práticas

### 1. Relatório de contratos por status de compliance e localização
```sql
SELECT c.id AS contract_id, c.nome AS contract_name, c.compliance_status, g.country, g.province, g.city
FROM contracts c
LEFT JOIN geographies.city g ON c.geography_city_id = g.id
ORDER BY c.compliance_status, g.country, g.city;
```

### 2. Relatório de processos com responsáveis e compliance cruzado
```sql
SELECT p.id AS process_id, p.nome AS process_name, p.responsible, p.compliance_status, c.nome AS contract_name
FROM processes p
LEFT JOIN contracts c ON p.contract_id = c.id
ORDER BY p.responsible, p.compliance_status;
```

### 3. Indicadores (KPIs) sem referência externa ou código internacional
```sql
SELECT id, nome, official_code, external_ref FROM kpis.indicator WHERE (external_ref IS NULL OR external_ref = '') OR (official_code IS NULL OR official_code = '');
```

### 4. Auditoria de alterações recentes em contratos e processos
```sql
SELECT 'contract' AS domain, id, updated_at, updated_by FROM contracts WHERE updated_at > NOW() - INTERVAL '30 days'
UNION ALL
SELECT 'process', id, updated_at, updated_by FROM processes WHERE updated_at > NOW() - INTERVAL '30 days'
ORDER BY updated_at DESC;
```

### 5. Relatório de lacunas de privacidade e acessibilidade
```sql
SELECT 'contract' AS domain, id, nome, privacy_level, accessibility FROM contracts WHERE privacy_level IS NULL OR accessibility IS NULL
UNION ALL
SELECT 'process', id, nome, privacy_level, accessibility FROM processes WHERE privacy_level IS NULL OR accessibility IS NULL;
```

### 6. Relatório de interoperabilidade (registros com referência externa)
```sql
SELECT 'contract' AS domain, id, nome, external_ref FROM contracts WHERE external_ref IS NOT NULL AND external_ref <> ''
UNION ALL
SELECT 'process', id, nome, external_ref FROM processes WHERE external_ref IS NOT NULL AND external_ref <> ''
UNION ALL
SELECT 'kpi', id, nome, external_ref FROM kpis.indicator WHERE external_ref IS NOT NULL AND external_ref <> '';
```
