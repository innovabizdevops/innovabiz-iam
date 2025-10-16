# PostgreSQL Database Dashboard - Documentação Técnica

## Visão Geral

O PostgreSQL Database Dashboard é um componente essencial do ecossistema de observabilidade da plataforma INNOVABIZ, projetado para monitorar e analisar o desempenho, saúde e utilização das instâncias PostgreSQL. Este dashboard oferece visibilidade abrangente sobre todos os aspectos críticos das bases de dados PostgreSQL, incluindo conexões ativas, performance de consultas, utilização de armazenamento, replicação e estatísticas de cache.

Como parte integrante da arquitetura multi-contexto da INNOVABIZ, este dashboard suporta filtragem por tenant, região e ambiente, permitindo uma visão granular dos serviços de banco de dados em diferentes contextos operacionais.

## Objetivo e Escopo

Este dashboard foi desenvolvido para atender às seguintes necessidades:

- Monitorar a saúde e desempenho das instâncias PostgreSQL em tempo real
- Identificar gargalos de performance e problemas de consultas longas
- Acompanhar o crescimento e utilização do armazenamento de dados
- Monitorar o estado da replicação e atrasos
- Analisar a eficácia do cache e taxas de acerto
- Fornecer visibilidade multi-contexto alinhada com a arquitetura da plataforma INNOVABIZ
- Suportar troubleshooting, planejamento de capacidade e otimização de desempenho

## Estrutura do Dashboard

O dashboard PostgreSQL está organizado em cinco seções principais, cada uma focada em um aspecto específico do monitoramento de banco de dados:

### 1. Database Overview

Esta seção fornece uma visão geral do estado atual das instâncias PostgreSQL:

- **Connection Usage (%)**: Medidor que mostra a porcentagem de conexões utilizadas em relação ao máximo configurado
- **Connections by State**: Gráfico temporal mostrando conexões classificadas por estado (active, idle, idle in transaction)
- **PostgreSQL Uptime**: Estatística do tempo de atividade contínua da instância PostgreSQL

### 2. Database Performance

Métricas relacionadas ao desempenho geral do banco de dados:

- **Average Query I/O Time**: Gráfico temporal do tempo médio de leitura/escrita para operações de I/O
- **Transactions (Commits vs Rollbacks)**: Comparação entre transações bem-sucedidas (commits) e revertidas (rollbacks)

### 3. Database Storage

Monitoramento da utilização de armazenamento do banco de dados:

- **Database Size by Database**: Gráfico temporal mostrando o crescimento de cada banco de dados ao longo do tempo
- **Tablespace Usage**: Medidor de barra mostrando o uso de cada tablespace

### 4. Query Performance

Análise detalhada do desempenho de consultas:

- **Active Queries by Database**: Gráfico temporal mostrando consultas ativas por banco de dados
- **Long Running Queries (>30s)**: Tabela que lista consultas em execução por mais de 30 segundos

### 5. Replication & Cache

Monitoramento de replicação e eficiência de cache:

- **Replication Lag**: Gráfico temporal mostrando o atraso de replicação em segundos
- **Cache Hit Ratio by Database**: Medidor de barra mostrando a taxa de acerto do cache por banco de dados

## Variáveis e Multi-Contexto

O dashboard implementa o padrão multi-contexto da INNOVABIZ através das seguintes variáveis:

| Variável | Descrição | Uso |
|----------|-----------|-----|
| `tenant_id` | Identificador do tenant | Filtrar métricas por tenant específico |
| `region_id` | Identificador da região | Filtrar métricas por região específica |
| `environment` | Ambiente (prod, staging, dev) | Filtrar métricas por ambiente específico |
| `instance` | Instância PostgreSQL específica | Filtrar métricas por instância específica |

Todas as variáveis suportam a seleção múltipla e a opção "All" para visualização consolidada. As variáveis são hierárquicas, de modo que a seleção de `tenant_id` afeta as opções disponíveis para `region_id`, e assim por diante.

## Requisitos e Dependências

### Métricas Prometheus Requeridas

Este dashboard utiliza métricas coletadas pelo PostgreSQL Exporter para Prometheus. As principais métricas utilizadas são:

- `pg_up`: Status de atividade da instância PostgreSQL
- `pg_stat_activity_count`: Contador de atividade por estado
- `pg_settings_max_connections`: Número máximo de conexões permitidas
- `pg_stat_database_*`: Métricas diversas de atividade do banco de dados
- `pg_postmaster_uptime_seconds`: Tempo de atividade do processo PostgreSQL
- `pg_stat_database_blk_read_time_seconds_total`: Tempo total de leitura de blocos
- `pg_stat_database_blk_write_time_seconds_total`: Tempo total de escrita de blocos
- `pg_stat_database_xact_commit_total`: Contador de transações confirmadas
- `pg_stat_database_xact_rollback_total`: Contador de transações revertidas
- `pg_database_size_bytes`: Tamanho de cada banco de dados em bytes
- `pg_tablespace_size_bytes`: Tamanho de cada tablespace em bytes
- `pg_stat_activity_seconds`: Duração de atividade de cada sessão
- `pg_replication_lag_seconds`: Atraso de replicação em segundos
- `pg_stat_database_blks_hit`: Contador de acertos no cache
- `pg_stat_database_blks_read`: Contador de leituras em disco (misses de cache)

### Requisitos de Configuração

Para a operação adequada deste dashboard, é necessário:

1. **Prometheus**: Versão 2.30+ configurado para raspagem de métricas PostgreSQL
2. **PostgreSQL Exporter**: Versão 0.10.0+ configurado corretamente
3. **Grafana**: Versão 9.0+ para suportar todos os recursos visuais
4. **Labels de Contexto**: Métricas com labels `tenant_id`, `region_id` e `environment` para suportar filtragem multi-contexto

## Implementação e Configuração

### Instalação do PostgreSQL Exporter

O PostgreSQL Exporter deve ser configurado para cada instância PostgreSQL que se deseja monitorar:

```bash
# Exemplo de configuração do PostgreSQL Exporter
docker run --name postgres_exporter \
  -e DATA_SOURCE_NAME="postgresql://username:password@hostname:5432/database?sslmode=disable" \
  -e PG_EXPORTER_EXTEND_QUERY_PATH="/path/to/custom-queries.yaml" \
  -p 9187:9187 \
  quay.io/prometheuscommunity/postgres-exporter
```

### Configuração de Labels Multi-Contexto

Para suportar a filtragem multi-contexto da INNOVABIZ, configure o Prometheus para adicionar os labels necessários:

```yaml
# Trecho do prometheus.yml
scrape_configs:
  - job_name: 'postgresql'
    static_configs:
      - targets: ['postgres_exporter:9187']
        labels:
          tenant_id: 'tenant1'
          region_id: 'br-east'
          environment: 'production'
```

### Importação do Dashboard

1. Navegue até o Grafana e selecione "Import" no menu lateral
2. Faça upload do arquivo JSON do dashboard ou cole seu conteúdo
3. Configure a fonte de dados Prometheus
4. Clique em "Import" para finalizar

## Casos de Uso Operacionais

### SRE e DevOps

- **Monitoramento Proativo**: Visualize tendências de crescimento de dados e gargalos de conexão
- **Investigação de Incidentes**: Identifique consultas problemáticas e problemas de replicação
- **Planejamento de Capacidade**: Analise o crescimento de bancos de dados e use históricos para previsões
- **Otimização de Performance**: Identifique oportunidades de melhorias em cache e índices

### DBAs e Desenvolvedores

- **Diagnóstico de Problemas**: Identifique consultas lentas e problemas de bloqueio
- **Otimização de Aplicações**: Correlacione problemas de performance com changes em aplicações
- **Planejamento de Manutenção**: Programe manutenções com base em períodos de baixa utilização
- **Validação de Changes**: Confirme que alterações de esquema não impactam negativamente a performance

### Gestão de Incidentes

- **Detecção de Anomalias**: Identifique rapidamente problemas de replicação ou conectividade
- **Triagem de Problemas**: Determine se o banco de dados é causa raiz de incidentes
- **Coordenação de Resposta**: Compartilhe visualizações consistentes durante incidentes
- **Análise Post-Mortem**: Use dados históricos para análise após resolução de incidentes

## Governança e Manutenção

### Propriedade e Responsabilidade

Este dashboard é mantido pela equipe de Observabilidade e DBA da INNOVABIZ, sob supervisão do time de Plataforma. Qualquer alteração significativa deve ser aprovada pelo processo padrão de change management.

### Ciclo de Vida e Versionamento

- O dashboard segue o versionamento semântico (X.Y.Z)
- Alterações são documentadas no controle de versão
- Atualizações são publicadas via processo GitOps

### Manutenção e Atualizações

As seguintes atividades de manutenção são recomendadas:

- Revisão trimestral de thresholds e alertas
- Validação após atualizações do PostgreSQL ou do exporter
- Adição de métricas conforme necessidades evoluem
- Teste em ambientes não-produtivos antes de atualizar dashboards de produção

## Compliance e Segurança

### Considerações de Compliance

Este dashboard foi projetado para suportar os seguintes frameworks e regulamentações:

- **PCI DSS 4.0**: Suporta os requisitos 10.4.1 (monitoramento de acesso a dados sensíveis)
- **ISO 27001**: Alinhado com controles A.12.1.3 (gestão de capacidade) e A.12.4 (logging e monitoramento)
- **LGPD/GDPR**: Não exibe dados pessoais sensíveis, suportando privacy by design
- **NIST CSF**: Suporta as funções Identify, Protect e Detect

### Controle de Acesso

O acesso ao dashboard deve ser controlado via IAM da INNOVABIZ, seguindo o princípio de menor privilégio:

- **View-only**: Para desenvolvedores e equipes de suporte de primeiro nível
- **Editor**: Para DBAs e equipes de operações
- **Admin**: Para proprietários de serviços e SREs seniores

### Isolamento de Tenant

O dashboard implementa isolamento de tenant via:

- Variáveis de filtro com validação de permissões
- Autenticação integrada com o IAM central
- Logging de acesso para auditoria

## Integração com Ecossistema INNOVABIZ

### Alerting

As principais métricas deste dashboard podem ser usadas para configurar alertas no Prometheus AlertManager:

- Conexões acima de 85% do limite por mais de 15 minutos
- Replication lag acima de 300 segundos por mais de 5 minutos
- Taxa de acerto de cache abaixo de 80% por mais de 10 minutos
- Consultas executando por mais de 10 minutos

### Integração com Outros Sistemas

Este dashboard se integra com outros componentes do ecossistema INNOVABIZ:

- **Incident Management**: Correlação via tenant_id, region_id e timestamps
- **Service Catalog**: Mapeamento de bases de dados para serviços
- **CI/CD**: Correlação com deployments para análise de impacto
- **RunBooks**: Links para procedimentos específicos de troubleshooting

## Suporte e Contato

Para questões relacionadas a este dashboard:

- **Problemas Técnicos**: Abra um ticket na categoria "Observabilidade > Dashboards"
- **Sugestões de Melhoria**: Submeta via portal de feedback ou abra um PR no repositório
- **Documentação**: Consulte a wiki da plataforma para tutoriais adicionais

## Histórico de Versões

| Versão | Data | Descrição | Autor |
|--------|------|-----------|-------|
| 1.0.0  | 25/07/2025 | Versão inicial do dashboard | Eduardo Jeremias |

---

**© 2025 INNOVABIZ - Documento Interno - Confidencial**