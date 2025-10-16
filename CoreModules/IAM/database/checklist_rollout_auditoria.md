# Checklist de Rollout e Auditoria – Plataforma InnovaBiz

## 1. Pré-Implantação
- [ ] Validar ambiente de homologação e produção (PostgreSQL, permissões, porta 5433)
- [ ] Backup completo do banco antes da execução
- [ ] Equipe de dados/governança ciente do cronograma

## 2. Execução dos Scripts
- [ ] Executar scripts na ordem: organization, business, products/services, compliance, risk, processes, reference, ESG, supply_chain, hr, governance automation, automation_bi_views_esg_supplychain_hr
- [ ] Verificar ausência de erros em cada etapa
- [ ] Validar criação de triggers, views e funções

## 3. Integração BI
- [ ] Conectar Power BI, Metabase ou Superset ao banco
- [ ] Importar views do schema bi (compliance, risco, ESG, supply chain, RH)
- [ ] Criar dashboards e validar visualizações

## 4. Checklist de Governança
- [ ] Todos os domínios com compliance, privacidade, rastreabilidade e auditoria
- [ ] Documentação multilíngue e checklist acessíveis
- [ ] Pipeline CI/CD configurado (opcional, recomendado)

## 5. Integração Externa
- [ ] Estruturar tabelas/campos para integração ERP, APIs públicas
- [ ] Validar integração e logs de auditoria

## 6. Pós-Implantação
- [ ] Realizar treinamento das equipes (usar slides)
- [ ] Revisão periódica dos dashboards e checklist
- [ ] Atualizar scripts e documentação conforme evolução

## 7. Suporte e Expansão
- [ ] Canal de suporte e governança disponível
- [ ] Processo para sugestão de melhorias e novos domínios
