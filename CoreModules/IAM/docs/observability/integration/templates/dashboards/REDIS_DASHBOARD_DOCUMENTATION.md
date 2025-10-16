# Redis Database Dashboard - Documentação Técnica

## Visão Geral

O Redis Database Dashboard é um componente crítico do ecossistema de observabilidade da plataforma INNOVABIZ, projetado para monitorar e analisar o desempenho, saúde e utilização das instâncias Redis. Este dashboard fornece visibilidade abrangente sobre todos os aspectos fundamentais dos serviços de cache e armazenamento em memória Redis, incluindo utilização de memória, performance de comandos, métricas de cache, e estado de conexão dos clientes.

Como parte integrante da arquitetura multi-contexto da INNOVABIZ, este dashboard suporta filtragem por tenant, região e ambiente, permitindo uma visão granular dos serviços de cache em diferentes contextos operacionais, alinhado com os requisitos de observabilidade e governança da plataforma.

## Objetivo e Escopo

Este dashboard foi desenvolvido para atender às seguintes necessidades:

- Monitorar a saúde e desempenho das instâncias Redis em tempo real
- Visualizar métricas críticas de utilização de memória e cache
- Identificar gargalos de performance e latência de comandos
- Analisar padrões de tráfego de rede e comandos mais utilizados
- Detectar problemas com evicção de chaves e rejeição de conexões
- Fornecer visibilidade multi-contexto alinhada com a arquitetura da plataforma INNOVABIZ
- Suportar troubleshooting, planejamento de capacidade e otimização de desempenho

## Estrutura do Dashboard

O dashboard Redis está organizado em quatro seções principais, cada uma focada em um aspecto específico do monitoramento de serviços Redis:

### 1. Redis Overview

Esta seção fornece uma visão geral do estado atual das instâncias Redis:

- **Uptime**: Estatística do tempo de atividade contínua da instância Redis
- **Connected Clients**: Número de clientes atualmente conectados ao Redis
- **Commands Processed**: Gráfico temporal da taxa de comandos processados por segundo
- **Memory Usage**: Medidor de barra que mostra o uso de memória em bytes
- **Memory Usage %**: Medidor que mostra a porcentagem de utilização de memória em relação ao máximo configurado
- **Total Keys**: Gráfico temporal mostrando o número total de chaves armazenadas

### 2. Redis Performance

Métricas relacionadas ao desempenho geral do sistema de cache:

- **Cache Hit Rate**: Taxa de acerto do cache (hits vs misses) ao longo do tempo
- **Command Latency**: Latência média de execução dos comandos Redis

### 3. Redis Operations

Análise detalhada do tráfego de rede e uso de comandos:

- **Network Traffic**: Tráfego de entrada e saída da rede em bytes por segundo
- **Commands by Type**: Distribuição de comandos por tipo (GET, SET, DEL, etc.)

### 4. Redis Health

Monitoramento de indicadores de saúde do sistema:

- **Evicted Keys Rate**: Taxa de chaves removidas por pressão de memória
- **Expired Keys Rate**: Taxa de chaves expiradas automaticamente
- **Rejected Connections**: Conexões rejeitadas devido ao limite máximo de clientes

## Variáveis e Multi-Contexto

O dashboard implementa o padrão multi-contexto da INNOVABIZ através das seguintes variáveis:

| Variável | Descrição | Uso |
|----------|-----------|-----|
| `tenant_id` | Identificador do tenant | Filtrar métricas por tenant específico |
| `region_id` | Identificador da região | Filtrar métricas por região específica |
| `environment` | Ambiente (prod, staging, dev) | Filtrar métricas por ambiente específico |
| `instance` | Instância Redis específica | Filtrar métricas por instância específica |

Todas as variáveis suportam a seleção múltipla e a opção "All" para visualização consolidada. As variáveis são hierárquicas, de modo que a seleção de `tenant_id` afeta as opções disponíveis para `region_id`, e assim por diante, garantindo consistência nos filtros aplicados.

## Requisitos e Dependências

### Métricas Prometheus Requeridas

Este dashboard utiliza métricas coletadas pelo Redis Exporter para Prometheus. As principais métricas utilizadas são:

- `redis_up`: Status de atividade da instância Redis
- `redis_uptime_in_seconds`: Tempo de atividade do servidor Redis
- `redis_connected_clients`: Número de clientes conectados
- `redis_commands_processed_total`: Total de comandos processados
- `redis_memory_used_bytes`: Memória utilizada em bytes
- `redis_memory_max_bytes`: Limite máximo de memória configurado
- `redis_db_keys`: Número de chaves por banco de dados
- `redis_keyspace_hits_total`: Contador de acertos no cache
- `redis_keyspace_misses_total`: Contador de erros no cache
- `redis_latency_spike_seconds`: Latência de comandos
- `redis_net_input_bytes_total`: Total de bytes recebidos
- `redis_net_output_bytes_total`: Total de bytes enviados
- `redis_commands_total`: Contador de comandos por tipo
- `redis_evicted_keys_total`: Total de chaves removidas por pressão de memória
- `redis_expired_keys_total`: Total de chaves expiradas automaticamente
- `redis_rejected_connections_total`: Total de conexões rejeitadas

### Requisitos de Configuração

Para a operação adequada deste dashboard, é necessário:

1. **Prometheus**: Versão 2.30+ configurado para raspagem de métricas Redis
2. **Redis Exporter**: Versão 1.20.0+ configurado corretamente
3. **Grafana**: Versão 9.0+ para suportar todos os recursos visuais
4. **Labels de Contexto**: Métricas com labels `tenant_id`, `region_id` e `environment` para suportar filtragem multi-contexto

## Implementação e Configuração

### Instalação do Redis Exporter

O Redis Exporter deve ser configurado para cada instância Redis que se deseja monitorar:

```bash
# Exemplo de configuração do Redis Exporter
docker run --name redis_exporter \
  -e REDIS_ADDR=redis://redis-server:6379 \
  -e REDIS_PASSWORD=your-password \
  -e REDIS_EXPORTER_CHECK_KEYS="*" \
  -p 9121:9121 \
  oliver006/redis_exporter:latest
```

### Configuração de Labels Multi-Contexto

Para suportar a filtragem multi-contexto da INNOVABIZ, configure o Prometheus para adicionar os labels necessários:

```yaml
# Trecho do prometheus.yml
scrape_configs:
  - job_name: 'redis'
    static_configs:
      - targets: ['redis_exporter:9121']
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

- **Monitoramento Proativo**: Detecte anomalias de uso de memória e conexões antes que afetem o serviço
- **Investigação de Incidentes**: Identifique gargalos de performance e relações com picos de tráfego
- **Planejamento de Capacidade**: Avalie tendências de uso de memória e crescimento de chaves para escalar adequadamente
- **Otimização de Performance**: Identifique comandos ineficientes e oportunidades de melhoria no cache

### Desenvolvedores

- **Diagnóstico de Problemas**: Visualize latência de comandos e taxas de acerto de cache
- **Otimização de Aplicações**: Ajuste políticas de cache e expiração com base em padrões de uso observados
- **Validação de Changes**: Confirme que alterações na aplicação não impactam negativamente o uso do Redis
- **Detecção de Anti-padrões**: Identifique uso inadequado do Redis (como comandos custosos ou sem expiração)

### Gestão de Incidentes

- **Detecção de Anomalias**: Identifique rapidamente problemas de memória ou conexão
- **Triagem de Problemas**: Determine se o Redis é causa raiz ou componente afetado em incidentes
- **Coordenação de Resposta**: Compartilhe visualizações consistentes entre equipes durante incidentes
- **Análise Post-Mortem**: Use dados históricos para análise após resolução de incidentes

## Governança e Manutenção

### Propriedade e Responsabilidade

Este dashboard é mantido pela equipe de Observabilidade e Plataforma da INNOVABIZ, com suporte das equipes de Backend e Infra-Cloud. Qualquer alteração significativa deve ser aprovada pelo processo padrão de change management.

### Ciclo de Vida e Versionamento

- O dashboard segue o versionamento semântico (X.Y.Z)
- Alterações são documentadas no controle de versão
- Atualizações são publicadas via processo GitOps

### Manutenção e Atualizações

As seguintes atividades de manutenção são recomendadas:

- Revisão trimestral de thresholds e alertas
- Validação após atualizações do Redis ou do exporter
- Adição de métricas conforme necessidades evoluem
- Teste em ambientes não-produtivos antes de atualizar dashboards de produção

## Compliance e Segurança

### Considerações de Compliance

Este dashboard foi projetado para suportar os seguintes frameworks e regulamentações:

- **PCI DSS 4.0**: Suporta os requisitos 10.4.1 (monitoramento de acesso a dados) e 6.4.3 (monitoramento de cache de dados sensíveis)
- **ISO 27001**: Alinhado com controles A.12.1.3 (gestão de capacidade) e A.12.4 (logging e monitoramento)
- **LGPD/GDPR**: Não exibe dados pessoais sensíveis, suportando privacy by design
- **NIST CSF**: Suporta as funções Identify, Protect e Detect do framework

### Controle de Acesso

O acesso ao dashboard deve ser controlado via IAM da INNOVABIZ, seguindo o princípio de menor privilégio:

- **View-only**: Para desenvolvedores e equipes de suporte
- **Editor**: Para SREs e equipes de plataforma
- **Admin**: Para proprietários de serviços e administradores de plataforma

### Isolamento de Tenant

O dashboard implementa isolamento de tenant via:

- Variáveis de filtro com validação de permissões
- Autenticação integrada com o IAM central
- Logging de acesso para auditoria

## Integração com Ecossistema INNOVABIZ

### Alerting

As principais métricas deste dashboard podem ser usadas para configurar alertas no Prometheus AlertManager:

- Utilização de memória acima de 85% por mais de 5 minutos
- Taxa de acerto de cache abaixo de 70% por mais de 10 minutos
- Latência de comandos acima de 100ms por mais de 2 minutos
- Taxa de evicção de chaves acima de 100 por segundo por mais de 1 minuto
- Conexões rejeitadas acima de 0 por mais de 1 minuto

### Integração com Outros Sistemas

Este dashboard se integra com outros componentes do ecossistema INNOVABIZ:

- **Incident Management**: Correlação via tenant_id, region_id e timestamps
- **Service Catalog**: Mapeamento de instâncias Redis para serviços
- **CI/CD**: Correlação com deployments para análise de impacto
- **RunBooks**: Links para procedimentos específicos de troubleshooting

## Suporte e Contato

Para questões relacionadas a este dashboard:

- **Problemas Técnicos**: Abra um ticket na categoria "Observabilidade > Cache"
- **Sugestões de Melhoria**: Submeta via portal de feedback ou abra um PR no repositório
- **Documentação**: Consulte a wiki da plataforma para tutoriais adicionais

## Histórico de Versões

| Versão | Data | Descrição | Autor |
|--------|------|-----------|-------|
| 1.0.0  | 25/07/2025 | Versão inicial do dashboard | Eduardo Jeremias |

---

**© 2025 INNOVABIZ - Documento Interno - Confidencial**