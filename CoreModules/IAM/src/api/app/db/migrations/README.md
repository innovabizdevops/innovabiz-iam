# INNOVABIZ IAM - Migrações do Banco de Dados para Auditoria Multi-Contexto

**Autor**: Eduardo Jeremias  
**Versão**: 1.0.0  
**Data**: 2025  

## Visão Geral

Este diretório contém os scripts de migração para o sistema de auditoria multi-contexto do módulo IAM da plataforma INNOVABIZ. Os scripts são projetados para serem executados sequencialmente para configurar todas as tabelas, funções e dados iniciais necessários para o sistema de auditoria.

## Arquitetura Multi-Contexto

O sistema de auditoria foi projetado com suporte completo para:

- **Multi-tenant**: Isolamento por tenant com chaves de particionamento otimizadas
- **Multi-regional**: Suporte para diferentes contextos regionais (BR, US, EU, AO)
- **Multi-regulatório**: Compliance com múltiplos frameworks regulatórios:
  - LGPD (Brasil)
  - GDPR (União Europeia)
  - SOX (Estados Unidos)
  - PCI DSS (Global)
  - PSD2 (União Europeia)
  - BACEN (Brasil)
  - BNA (Angola)
  - NIST (Estados Unidos)

## Estrutura das Migrações

Os scripts estão organizados na seguinte ordem:

1. **001_create_audit_types.sql** - Cria os tipos enumerados para categorias, severidade, frameworks de compliance
2. **002_create_audit_tables.sql** - Cria as tabelas principais do sistema de auditoria com índices otimizados
3. **003_create_audit_functions.sql** - Define funções e procedimentos para operações avançadas de auditoria
4. **004_insert_initial_data.sql** - Insere políticas de retenção padrão e configurações iniciais

## Principais Funcionalidades

- **Rastreabilidade Completa**: Registro detalhado de todos os eventos de auditoria
- **Políticas de Retenção**: Configuração flexível por contexto regional e framework regulatório
- **Mascaramento e Anonimização**: Suporte para GDPR/LGPD com mascaramento de dados sensíveis
- **Particionamento Virtual**: Particionamento lógico para performance com grandes volumes
- **Estatísticas e Relatórios**: Geração automática de estatísticas e relatórios de compliance

## Tabelas Principais

- **audit_events**: Armazena todos os eventos de auditoria
- **audit_retention_policies**: Define políticas de retenção e anonimização
- **audit_compliance_reports**: Armazena relatórios de compliance gerados
- **audit_statistics**: Mantém estatísticas agregadas para análise e dashboards

## Índices Otimizados

Os índices foram projetados para otimizar as consultas mais frequentes:

- Busca por tenant_id
- Busca por contexto regional
- Busca por período (created_at)
- Busca por categoria de evento
- Busca por chave de partição (partition_key)
- Busca em tags e detalhes JSONB (usando GIN)

## Execução das Migrações

Para executar as migrações, utilize o script `run_migrations.sh` incluído neste diretório:

```bash
./run_migrations.sh
```

Ou execute cada script manualmente na sequência correta:

```bash
psql -U [username] -d [database] -f 001_create_audit_types.sql
psql -U [username] -d [database] -f 002_create_audit_tables.sql
psql -U [username] -d [database] -f 003_create_audit_functions.sql
psql -U [username] -d [database] -f 004_insert_initial_data.sql
```

## Configurações Avançadas

### Configuração de Políticas de Retenção Customizadas

Para criar políticas de retenção personalizadas para novos tenants:

```sql
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES (
    'seu_tenant_id',
    'BR',  -- Ou outro contexto regional
    730,   -- Dias de retenção
    'LGPD', -- Framework de compliance
    'AUTHENTICATION', -- Categoria de eventos
    'Descrição da política',
    TRUE,  -- Anonimização automática
    ARRAY['source_ip', 'user_name'], -- Campos a anonimizar
    TRUE   -- Política ativa
);
```

### Geração de Estatísticas

Para gerar estatísticas manualmente:

```sql
SELECT generate_audit_statistics(
    'seu_tenant_id',
    'BR',  -- Contexto regional (opcional)
    NOW() - INTERVAL '30 days',  -- Data inicial
    NOW(),  -- Data final
    ARRAY['category', 'success', 'severity']  -- Campos para agrupamento
);
```

## Considerações de Performance

- Eventos de auditoria são particionados logicamente usando a coluna `partition_key`
- Para grandes volumes, considere implementar particionamento físico por data
- Índices específicos foram criados para otimizar as consultas mais comuns
- Para tenants com alto volume, considere ajustar as configurações de autovacuum

## Suporte Multilíngue

O sistema suporta internacionalização completa, com armazenamento do código de idioma em cada evento de auditoria.