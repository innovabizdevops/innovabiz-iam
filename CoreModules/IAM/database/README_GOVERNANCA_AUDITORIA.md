# Guia Rápido de Governança, Compliance e Auditoria – Innovabiz Plataforma

## 1. Execução e Migração
- Execute todos os scripts SQL de migração, triggers, views e funções no banco de dados (via psql ou pgAdmin).
- Confirme que o comando `psql` está disponível no PATH do sistema.

## 2. Validação de Queries e Relatórios
- Utilize as queries de validação e relatórios integrados documentados em `V8__documentation_compliance_governance_risk.md`.
- Teste as views para BI (ex: `vw_contracts_compliance_summary`) conectando Power BI, Metabase, Superset ou Grafana.

## 3. Automação de Governança
- Agende os scripts batch/cron/pgAgent para rodar queries de compliance e gerar relatórios periódicos (exemplos fornecidos).
- Relatórios automáticos podem ser enviados por e-mail ou publicados em dashboards internos.

## 4. Changelog e Documentação
- Registre toda alteração relevante de schema, triggers e views na tabela `schema_changelog`.
- Mantenha o dicionário de dados e ERD atualizados e versionados junto ao código-fonte.

## 5. Treinamento e Transferência de Conhecimento
- Oriente os times de dados, compliance e TI sobre uso dos relatórios, queries e dashboards.
- Disponibilize este guia e a documentação de processos de governança para consulta.

## 6. Governança Contínua
- Estabeleça rotina para revisão dos relatórios e dashboards.
- Planeje revisões semestrais para expansão dos controles para novos domínios.
- Garanta que toda alteração de schema seja registrada e documentada.

## 7. Fluxo de Correção de Lacunas
- Utilize os relatórios automáticos para identificar registros sem compliance, responsável, privacidade ou acessibilidade.
- Abra tickets/tarefas corretivas para os responsáveis e acompanhe a resolução.

---

### Exemplo de Query para Auditoria de Alterações
```sql
SELECT * FROM schema_changelog ORDER BY change_timestamp DESC;
```

---

**Dúvidas, sugestões ou necessidade de expansão? Consulte a equipe de governança de dados ou o responsável técnico pela plataforma.**
