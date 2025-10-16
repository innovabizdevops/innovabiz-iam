# Checklist de Governança, Compliance e Risco – Innovabiz

## Estrutura e Modelagem
- [x] Chaves primárias UUID em todas as tabelas
- [x] Campos de auditoria (created_at, updated_at, etc.)
- [x] Tabelas de relacionamento N:N para integrações normativas/frameworks
- [x] Controle de status e versionamento
- [x] Relacionamentos explícitos com FK

## Compliance e Auditoria
- [x] Tabela de consentimento e logs de auditoria ativos
- [x] Triggers de auditoria implementadas nas tabelas críticas
- [x] Campos para privacidade, consentimento, revogação
- [x] Logs de acesso e alteração rastreáveis

## Checklist de Governança e Compliance para Domínios Corporativos

- [ ] Todos os registros possuem campos de auditoria (`created_at`, `updated_at`, `created_by`, `updated_by`)
- [ ] Todos os registros possuem campo `responsável` preenchido
- [ ] Todos os registros possuem status administrativo válido
- [ ] Todos os registros possuem código internacional (ISO, UN/LOCODE, etc, quando aplicável)
- [ ] Todos os registros possuem fonte de dados documentada
- [ ] Todos os registros possuem vigência (`valid_from`/`valid_to`)
- [ ] Todos os registros possuem referência externa (`external_ref`) quando aplicável
- [ ] Todos os registros possuem nível hierárquico (`hierarchy_level`) e `parent_id` quando aplicável
- [ ] Todos os registros possuem `compliance_status` preenchido
- [ ] Todos os registros possuem privacidade (`privacy_level`) e acessibilidade (`accessibility`) documentadas
- [ ] Todos os registros possuem idioma principal (`main_language`) documentado
- [ ] Todos os registros de localização possuem coordenadas (`latitude`, `longitude`) e bounding box (quando aplicável)
- [ ] Todos os domínios possuem triggers de auditoria e logs de alteração
- [ ] Todos os domínios possuem documentação atualizada no dicionário de dados e ERD

## Governança e Expansão
- [x] Documentação (dicionário de dados, ERD, checklist)
- [x] Estrutura modular para expansão de domínios
- [x] Integração com módulos IAM, produtos, serviços
- [x] Revisão periódica de compliance e segurança

> Atualize este checklist a cada nova feature, migração ou integração!
