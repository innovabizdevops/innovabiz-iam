# INNOVABIZ IAM Audit Service - Documentação Fluentd

**Versão:** 3.0.0  
**Data:** 31/07/2025  
**Classificação:** Oficial  
**Status:** ✅ Implementado  
**Autor:** Equipe INNOVABIZ DevOps  
**Aprovado por:** Eduardo Jeremias  

## 1. Visão Geral

O Fluentd é o componente central de coleta e processamento de logs na arquitetura de observabilidade do IAM Audit Service da INNOVABIZ. Como coletor unificado de logs, o Fluentd oferece um pipeline flexível e robusto para coletar, filtrar, transformar e rotear logs de diversas fontes para múltiplos destinos, implementando a estratégia "dual-write" para garantir redundância e complementaridade entre diferentes sistemas de armazenamento.

### 1.1 Funcionalidades Principais

- **Coleta Universal**: Ingestão de logs de diversas fontes (arquivos, syslog, containers)
- **Processamento em Pipeline**: Filtragem, parsing, enriquecimento e transformação
- **Roteamento Inteligente**: Distribuição de logs para múltiplos destinos baseado em regras
- **Buffering Confiável**: Buffer persistente para resistência a falhas
- **Extensibilidade**: +500 plugins disponíveis para integração
- **Alta Performance**: Escrito em Ruby/C com otimizações para alta vazão
- **Multi-tenant**: Isolamento completo de processamento por tenant
- **Compliance**: Capacidades de mascaramento e anonimização para conformidade

### 1.2 Posicionamento na Arquitetura

O Fluentd atua como o backbone de coleta de logs, recebendo dados de:

- Aplicações IAM (via log files e FLUENT_FORWARD)
- Kubernetes logs (via node logging agent)
- Syslog de componentes de infraestrutura
- HTTP/JSON de aplicações customizadas
- Agentes de terceiros via conversor

E enviando dados para:

- Elasticsearch (armazenamento primário, pesquisa avançada)
- Loki (armazenamento econômico, consultas rápidas)
- S3/Azure Blob (armazenamento de longo prazo)
- Kafka (para processamento em tempo real)

## 2. Implementação Técnica

### 2.1 Manifesto Kubernetes

O Fluentd é implementado como DaemonSet no Kubernetes, garantindo que uma instância seja executada em cada node do cluster:

```yaml
# Trecho exemplificativo do manifesto
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  namespace: iam-system
  labels:
    app.kubernetes.io/name: fluentd
    app.kubernetes.io/part-of: innovabiz-observability
    innovabiz.com/module: iam-audit
    innovabiz.com/tier: observability
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: fluentd
  template:
    metadata:
      labels:
        app.kubernetes.io/name: fluentd
        app.kubernetes.io/part-of: innovabiz-observability
        innovabiz.com/module: iam-audit
    spec:
      serviceAccountName: fluentd
      tolerations:
      - key: node-role.kubernetes.io/control-plane
        effect: NoSchedule
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd
        image: innovabiz/fluentd-iam:3.0.0
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: elasticsearch-master.iam-system.svc
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
        - name: FLUENT_LOKI_URL
          value: http://loki-distributor.iam-system.svc:3100
        - name: TENANT_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['innovabiz.com/tenant-id']
        - name: REGION_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['innovabiz.com/region-id']
        - name: ENVIRONMENT
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['innovabiz.com/environment']
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
          requests:
            cpu: 250m
            memory: 512Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: config
          mountPath: /fluentd/etc
        - name: buffer
          mountPath: /fluentd/buffer
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: config
        configMap:
          name: fluentd-config
      - name: buffer
        persistentVolumeClaim:
          claimName: fluentd-buffer
```

### 2.2 Recursos e Escala

| Componente | Implementação | Recursos (req/limits) | Escalabilidade |
|------------|--------------|------------------------|----------------|
| **Fluentd DaemonSet** | 1 por node | 250m/500m CPU, 512Mi/1Gi Mem | Escala com cluster |
| **Fluentd Aggregators** | StatefulSet, 2+ réplicas | 500m/1000m CPU, 1Gi/2Gi Mem | Horizontal |
| **Buffer PVC** | 10Gi por instância | SSD StorageClass | Expansível |

### 2.3 Configuração Principal

A configuração do Fluentd é gerenciada via ConfigMap e segue este padrão:

```xml
# Seção global e input
<source>
  @type forward
  port 24224
  bind 0.0.0.0
</source>

<source>
  @type tail
  path /var/log/containers/iam-*.log
  pos_file /fluentd/buffer/iam-container.pos
  tag kubernetes.*
  read_from_head true
  <parse>
    @type json
    time_format %Y-%m-%dT%H:%M:%S.%NZ
  </parse>
</source>

# Processamento e filtragem
<filter kubernetes.**>
  @type kubernetes_metadata
  annotation_match ["^innovabiz\.com/"]
  watch true
  bearer_token_file /var/run/secrets/kubernetes.io/serviceaccount/token
  ca_file /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
</filter>

<filter kubernetes.**>
  @type record_transformer
  enable_ruby true
  <record>
    tenant_id ${record.dig("kubernetes", "annotations", "innovabiz.com/tenant-id") || ENV["TENANT_ID"] || "default"}
    region_id ${record.dig("kubernetes", "annotations", "innovabiz.com/region-id") || ENV["REGION_ID"] || "unknown"}
    environment ${record.dig("kubernetes", "annotations", "innovabiz.com/environment") || ENV["ENVIRONMENT"] || "production"}
    module ${record.dig("kubernetes", "labels", "app.kubernetes.io/name") || "unknown"}
    component ${record.dig("kubernetes", "labels", "app.kubernetes.io/component") || "unknown"}
    level ${record["stream"] == "stderr" ? "error" : (record["log"] =~ /ERROR|WARN|INFO|DEBUG/i)&.to_s || "info"}
    host ${record.dig("kubernetes", "host") || hostname}
  </record>
</filter>

<filter kubernetes.**>
  @type grep
  <regexp>
    key log
    pattern /\S+/
  </regexp>
</filter>

# Processamento avançado para logs específicos
<filter kubernetes.var.log.containers.iam-auth-**.log>
  @type parser
  key_name log
  reserve_data true
  <parse>
    @type multi_format
    <pattern>
      format json
      time_key timestamp
      time_format %Y-%m-%dT%H:%M:%S.%NZ
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z) (?<level>[A-Z]+) (?<message>.*)$/
      time_key time
      time_format %Y-%m-%dT%H:%M:%S.%NZ
    </pattern>
  </parse>
</filter>

# Mascaramento de dados sensíveis
<filter **>
  @type mask
  <rule>
    keys password,token,api_key,secret,credential
    mask_with [REDACTED]
  </rule>
  <rule>
    keys email
    mask_with ${record["email"].to_s.gsub(/(?<=.).(?=.*@)/, '*')}
    use_regex true
    replace_method regex
  </rule>
  <rule>
    keys document,cpf,cnpj
    mask_with ${record["document"].to_s.gsub(/(?<=^.{3}).(?=.{4}$)/, '*')}
    use_regex true
    replace_method regex
  </rule>
</filter>

# Roteamento dual-write
<match **>
  @type copy
  <store>
    @type elasticsearch
    host "#{ENV['FLUENT_ELASTICSEARCH_HOST']}"
    port "#{ENV['FLUENT_ELASTICSEARCH_PORT']}"
    scheme https
    ssl_verify false
    user "#{ENV['FLUENT_ELASTICSEARCH_USER']}"
    password "#{ENV['FLUENT_ELASTICSEARCH_PASSWORD']}"
    index_name ${tenant_id}.${module}.${component}.%Y%m%d
    include_tag_key true
    tag_key @log_name
    <buffer tag, tenant_id, module, component, time>
      @type file
      path /fluentd/buffer/elasticsearch
      timekey 1h
      timekey_wait 5m
      timekey_use_utc true
      chunk_limit_size 16M
      flush_thread_count 4
      retry_forever true
      retry_max_interval 30s
    </buffer>
  </store>
  <store>
    @type loki
    url "#{ENV['FLUENT_LOKI_URL']}"
    tenant "#{ENV['TENANT_ID']}"
    extra_labels {"region_id":"#{ENV['REGION_ID']}","environment":"#{ENV['ENVIRONMENT']}"}
    label_keys tenant_id,region_id,environment,module,component,level,host
    line_format json
    <buffer>
      @type file
      path /fluentd/buffer/loki
      flush_mode interval
      flush_interval 5s
      flush_thread_count 4
      retry_forever true
      retry_max_interval 30
      chunk_limit_size 2M
    </buffer>
  </store>
  <store>
    @type s3
    s3_bucket "#{ENV['ARCHIVE_BUCKET']}"
    s3_region "#{ENV['ARCHIVE_REGION']}"
    path logs/${tenant_id}/${region_id}/${module}/%Y/%m/%d/
    <format>
      @type json
    </format>
    <buffer tag, tenant_id, region_id, module, time>
      @type file
      path /fluentd/buffer/s3
      timekey 1h
      timekey_wait 10m
      timekey_use_utc true
      chunk_limit_size 32M
      total_limit_size 256M
      retry_forever true
      retry_max_interval 30
    </buffer>
  </store>
</match>
```## 3. Configuração Multi-dimensional

### 3.1 Estratégia Multi-tenant

A separação completa por tenant é implementada em múltiplos níveis:

- **Enriquecimento de Logs**: Todos os logs são etiquetados com `tenant_id` obrigatoriamente
- **Roteamento Específico**: Opção de configuração diferente por tenant
- **Índices Isolados**: Índices específicos por tenant no Elasticsearch
- **Labels**: Labels específicos por tenant no Loki
- **Armazenamento**: Diretórios separados por tenant no S3

Exemplo de configuração específica por tenant:

```xml
<match tenant1.**>
  @type copy
  <store>
    @type elasticsearch
    host elasticsearch-premium.iam-system.svc
    port 9200
    # ... configuração específica para tenant premium
    <buffer>
      # ... buffer maior para tenant premium
      flush_interval 1m
      chunk_limit_size 32M
    </buffer>
  </store>
  # ... outras saídas
</match>
```

### 3.2 Contexto Regional

O contexto regional é implementado através de:

- **Label Regional**: Todos os logs incluem `region_id` obrigatoriamente
- **Configuração Regional**: Plugins específicos por região
- **Compliance Regional**: Regras específicas conforme regulamentação local:
  - Brasil: LGPD - anonimização de dados pessoais
  - EU: GDPR - mascaramento extensivo e direito ao esquecimento
  - EUA: Regras específicas por estado (CCPA para CA)
  - Angola: Requisitos do BNA para transações financeiras

### 3.3 Contexto Ambiental

- **Label de Ambiente**: Todos os logs incluem `environment` (production, staging, etc.)
- **Configuração por Ambiente**: Níveis de logging e filtros específicos:
  - Production: Logs criticamente importantes, filtragem agressiva
  - Staging: Logs de performance e funcionalidade 
  - Development: Logs verbosos para debugging

### 3.4 Geração de Métricas por Dimensão

O Fluentd gera métricas sobre os logs processados, segmentadas por todas as dimensões:

```xml
<filter **>
  @type prometheus
  <metric>
    name fluentd_input_status_num_records_total
    type counter
    desc The total number of incoming records
    <labels>
      tenant ${tenant_id}
      region ${region_id}
      environment ${environment}
      module ${module}
      component ${component}
    </labels>
  </metric>
</filter>
```

## 4. Estratégias de Processamento

### 4.1 Parsing e Estruturação

O Fluentd utiliza diversos parsers para extrair dados estruturados de diferentes formatos de log:

| Tipo de Log | Parser | Campos Extraídos |
|-------------|--------|------------------|
| **JSON** | json | Todos os campos do JSON |
| **Logs de Aplicação** | regexp | timestamp, level, message, etc. |
| **Nginx Access** | nginx | method, path, status, user_agent, etc. |
| **Apache** | apache2 | client, method, path, code, size, etc. |
| **Syslog** | syslog | facility, severity, hostname, etc. |
| **Multiline** | multiline | message completo com quebras |
| **Grok** | grok | Campos customizados via padrões |

Exemplo de parser para logs de autenticação:

```xml
<filter iam-auth.**>
  @type parser
  key_name message
  reserve_data true
  <parse>
    @type grok
    <grok>
      pattern %{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} \[%{DATA:service}\] \[%{DATA:trace_id}\] \[%{DATA:user_id}\] %{GREEDYDATA:message}
    </grok>
  </parse>
</filter>
```

### 4.2 Enriquecimento de Logs

Logs são enriquecidos com informações contextuais e derivadas:

- **Metadados Kubernetes**: Namespace, pod, container, node, labels
- **Metadados de Sistema**: Hostname, IP, datacenter, região
- **Informações de Serviço**: Versão, ambiente, tier, criticidade
- **Derivação de Campos**: Severidade calculada, categorias, duração
- **Correlação**: IDs de correlação, trace IDs, session IDs

Exemplo de enriquecimento:

```xml
<filter **>
  @type record_transformer
  enable_ruby true
  <record>
    service_version ${record.dig("kubernetes", "labels", "version") || "unknown"}
    datacenter ${ENV["DATACENTER"] || "primary"}
    event_source "iam-audit"
    log_timestamp ${Time.now.utc.iso8601}
    correlation_id ${record["correlation_id"] || record["trace_id"] || record.dig("kubernetes", "pod_id") || "unknown"}
    duration ${record["end_time"] && record["start_time"] ? (Time.parse(record["end_time"]) - Time.parse(record["start_time"])).to_f : nil}
  </record>
</filter>
```

### 4.3 Filtragem e Sampling

Para gerenciar o volume de logs, são implementadas estratégias de filtragem:

- **Filtragem por Severidade**: Logs de DEBUG filtrados em produção
- **Filtragem de Ruído**: Remoção de logs de heartbeat e health check
- **Sampling Inteligente**: Redução de volume para logs repetitivos
- **Rate Limiting**: Limites por fonte para evitar inundações

Exemplo de sampling:

```xml
<filter high-volume.**>
  @type sampling
  interval 10  # Manter 1 a cada 10 logs
  <rule>
    key message
    pattern /repeated log pattern/
    interval 100  # Sampling mais agressivo para padrões específicos
  </rule>
</filter>
```

### 4.4 Mascaramento e Compliance

Dados sensíveis são tratados conforme exigências regulatórias:

- **PCI DSS**: Mascaramento completo de PAN, CVV, senhas
- **GDPR/LGPD**: Pseudonimização de identificadores pessoais
- **Dados Biométricos**: Remoção completa de dados biométricos
- **Tokenização**: Substituição de dados sensíveis por tokens

Exemplo de implementação:

```xml
<filter **>
  @type mask_sensitive
  mask_patterns [
    {"regex": "\\b(?:\\d[ -]*?){13,16}\\b", "replace": "[MASKED-PAN]"}, # PAN
    {"regex": "\\b\\d{3,4}\\b", "group": "cvv", "replace": "[CVV]"}, # CVV
    {"regex": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}\\b", "replace": "[EMAIL]"}, # Email
    {"regex": "\\b\\d{3}\\.\\d{3}\\.\\d{3}\\-\\d{2}\\b", "replace": "[CPF]"} # CPF
  ]
</filter>
```

## 5. Gestão de Buffer e Confiabilidade

### 5.1 Estratégia de Buffer

O Fluentd utiliza buffer persistente para garantir resistência a falhas:

- **Buffer File**: Armazenamento em disco para persistência
- **Chunking**: Divisão em chunks para processamento eficiente
- **Compressão**: Redução de tamanho para economia de espaço
- **Retry Exponencial**: Backoff exponencial para retentativas
- **Overflow Protection**: Proteção contra esgotamento de recursos

Configuração típica de buffer:

```xml
<buffer tag, time>
  @type file
  path /fluentd/buffer/${tag}
  timekey 1h
  timekey_wait 10m
  timekey_use_utc true
  chunk_limit_size 16M
  total_limit_size 4G
  queue_limit_length 32
  flush_interval 60s
  flush_thread_count 4
  overflow_action block
  retry_forever true
  retry_max_interval 30
  retry_randomize true
  compress gzip
</buffer>
```

### 5.2 Alta Disponibilidade

Para garantir zero perda de logs, são implementados:

- **Arquitetura Resiliente**: DaemonSet com pelo menos um agente por node
- **Heartbeat e Health Check**: Monitoramento constante de saúde
- **Circuit Breaking**: Proteção contra destinos indisponíveis
- **Load Balancing**: Distribuição entre agregadores
- **Replicação**: Múltiplas cópias de logs críticos

### 5.3 Monitoramento de Performance

O monitoramento do Fluentd inclui:

- **Métricas de Buffer**: Utilização, flush rate, age
- **Métricas de Processamento**: Taxa de ingestão, erros, latência
- **Métricas de Sistema**: CPU, memória, IO, rede
- **Alertas Preditivos**: Alertas antes de problemas críticos

Dashboard específico no Grafana monitora todos estes aspectos.

## 6. Integração com Outras Ferramentas

### 6.1 Integração com Elasticsearch

A integração com Elasticsearch é configurada para otimizar indexação e busca:

- **Índices por Tenant/Módulo**: Formato `<tenant_id>.<module>.<component>.YYYYMMDD`
- **Mapeamento Dinâmico**: Detecção automática de tipos com regras
- **Indexação em Massa**: Bulk indexing para performance
- **Retry com Backoff**: Tentativas exponenciais para resiliência
- **Autenticação**: TLS mútuo com certificados cliente
- **Pipeline de Ingestão**: Pré-processamento no Elasticsearch

### 6.2 Integração com Loki

A integração com Loki é otimizada para eficiência:

- **Label Selection**: Envio apenas de labels relevantes
- **Formato JSON**: Preservação de estrutura para consultas
- **Compressão**: Redução de tamanho para transferência eficiente
- **Batching**: Envio em lotes para melhor throughput
- **Tenant ID**: Isolamento explícito por tenant

### 6.3 Integração com S3/Blob Storage

Para arquivamento de longo prazo:

- **Organização Hierárquica**: `tenant_id/region_id/module/YYYY/MM/DD/HH/`
- **Compressão GZIP**: Redução de espaço de armazenamento
- **Object Lifecycle**: Políticas de transição para classes econômicas
- **Encryption**: Criptografia de objetos com chaves gerenciadas
- **IAM**: Controle de acesso granular por tenant
- **Metadata**: Tags e metadados para fácil identificação

### 6.4 Integração com Observability Portal

- **API GraphQL**: Exposição de status e métricas
- **Webhooks**: Notificações de eventos críticos
- **Configuração Dinâmica**: Ajustes de configuração via API
- **Status Dashboard**: Visualização de saúde e performance

## 7. Monitoramento e Alertas

### 7.1 Métricas de Performance

O Fluentd expõe métricas via Prometheus:

| Métrica | Descrição | Threshold de Alerta |
|---------|-----------|---------------------|
| `fluentd_input_status_num_records_total` | Registros processados por input | Queda >50% |
| `fluentd_output_status_num_errors_total` | Erros por output | >0 por 5min |
| `fluentd_output_status_buffer_queue_length` | Tamanho da fila de buffer | >80% capacidade |
| `fluentd_output_status_buffer_total_bytes` | Bytes em buffer | >80% capacidade |
| `fluentd_output_status_retry_count` | Contagem de retentativas | >10 por 5min |
| `fluentd_output_status_emit_records` | Registros emitidos | N/A (baseline) |
| `fluentd_output_status_emit_count` | Contagem de emissões | N/A (baseline) |

### 7.2 Alertas Configurados

```yaml
# Alertas para monitorar o Fluentd
- name: FluentdAlerts
  rules:
  - alert: FluentdBufferNearFull
    expr: fluentd_output_status_buffer_queue_length / fluentd_output_status_buffer_queue_limit * 100 > 80
    for: 5m
    labels:
      severity: warning
      component: fluentd
    annotations:
      summary: "Buffer Fluentd próximo da capacidade"
      description: "O buffer de {{ $labels.type }} está acima de 80% da capacidade"
      runbook: "https://docs.innovabiz.com/observability/runbooks/fluentd-buffer-full"

  - alert: FluentdHighErrorRate
    expr: sum(rate(fluentd_output_status_num_errors_total[5m])) by (type) > 0
    for: 5m
    labels:
      severity: critical
      component: fluentd
    annotations:
      summary: "Alta taxa de erros no Fluentd"
      description: "O output {{ $labels.type }} está apresentando erros"
      runbook: "https://docs.innovabiz.com/observability/runbooks/fluentd-error-rate"
```

### 7.3 Dashboards de Monitoramento

- **Fluentd Overview**: Visão geral de saúde e performance
- **Buffer Status**: Estado e utilização dos buffers
- **Error Tracking**: Rastreamento detalhado de erros
- **Throughput**: Vazão por tipo de log e destino
- **Performance Metrics**: CPU, memória e latência

## 8. Backup e Recuperação

### 8.1 Estratégia de Backup

- **Volumes de Buffer**: Backup diário dos volumes persistentes
- **Configuração**: Versionada e armazenada no GitOps
- **Metrics**: Snapshot de métricas históricas
- **Posição de Leitura**: Backup de arquivos .pos para continuidade

### 8.2 Procedimento de Recuperação

1. **Recuperação de Configuração**:
   - Aplicar ConfigMaps do repositório GitOps
   - Verificar versões e compatibilidade

2. **Recuperação de Buffers**:
   - Restaurar volumes PVC do backup
   - Verificar permissões e propriedade

3. **Restauração de Posição**:
   - Restaurar arquivos .pos para continuar do ponto correto
   - Ajustar timestamps se necessário

4. **Verificação de Integridade**:
   - Validar conectividade com fontes e destinos
   - Confirmar processamento de novos logs

### 8.3 Resposta a Incidentes

Procedimentos documentados para cenários comuns:

- **Perda de Buffer**: Restauração a partir de backups
- **Destino Indisponível**: Ativação de destino secundário
- **Sobrecarga de Volume**: Implementação de sampling emergencial
- **Corrupção de Dados**: Isolamento e reconstrução de índices## 9. Conformidade e Segurança

### 9.1 Requisitos de Conformidade

| Regulação | Requisito | Implementação |
|-----------|-----------|---------------|
| **PCI DSS 4.0** | 3.3 Mascaramento PAN | Plugin mask para dados de cartão |
| **PCI DSS 4.0** | 10.2 Trilhas de auditoria | Persistência de todos os logs de autenticação |
| **GDPR/LGPD** | Art. 17 Direito ao esquecimento | Pipeline para anonimização + API de exclusão |
| **GDPR/LGPD** | Art. 25 Privacy by Design | Mascaramento automático de dados pessoais |
| **ISO 27001** | A.12.4 Registros de eventos | Captura abrangente e protegida |
| **ISO 27001** | A.18.1.3 Proteção de registros | Integridade e não-repúdio dos logs |
| **NIST 800-53** | SI-4 Monitoramento do sistema | Correlação de eventos de segurança |
| **SOC 2** | CC 7.2 Monitoramento de anomalias | Detecção de padrões incomuns |

### 9.2 Controles de Segurança

- **Autenticação**: TLS mútuo para todas as comunicações
- **Autorização**: RBAC para acesso à configuração
- **Criptografia**: Em trânsito (TLS) e em repouso (volumes)
- **Segredos**: Integração com Kubernetes Secrets
- **Validação de Entrada**: Prevenção contra log injection
- **Auditoria**: Logs de alterações e acesso à configuração

### 9.3 Proteção de Dados Sensíveis

Mecanismos implementados para proteção de dados:

- **Detecção Automatizada**: Identificação de padrões sensíveis via regex
- **Tokenização**: Substituição reversível para analytics
- **Pseudonimização**: Substituição consistente para correlação
- **Hashing**: One-way para identificadores sensíveis
- **Restrição de Acesso**: RBAC para logs sensíveis

## 10. Operação e Manutenção

### 10.1 Procedimentos Operacionais

- **Health Check**: Verificação automática a cada 5 minutos
- **Rotação de Logs**: Automática baseada em tamanho e idade
- **Limpeza de Buffer**: Agendada durante períodos de baixo uso
- **Atualização**: Procedimento de rolling update sem perda de logs
- **Monitoramento Proativo**: Alertas preditivos antes de problemas

### 10.2 Troubleshooting

| Problema | Possíveis Causas | Resolução |
|----------|-----------------|-----------|
| Buffer cheio | Volume alto, destino lento | Aumentar flush threads, verificar destinos |
| Erros de parsing | Formato de log alterado | Ajustar padrões de parse, adicionar formatos |
| Alta latência | CPU/IO limitado, rede congestionada | Verificar recursos, aumentar buffers |
| Perda de logs | Buffer overflow, crash | Verificar logs de erro, restaurar de backup |
| Duplicação | Posição de leitura corrompida | Reajustar arquivo .pos, implementar deduplicação |

### 10.3 Runbooks

Runbooks detalhados foram criados para cenários comuns:

1. **Inicialização e Verificação**:
   - Validação de configuração
   - Verificação de conectividade
   - Teste de ingestão e entrega

2. **Diagnóstico de Performance**:
   - Análise de gargalos
   - Verificação de uso de recursos
   - Otimização de configuração

3. **Recuperação de Falhas**:
   - Procedimentos de restauração
   - Verificação de integridade
   - Reprocessamento de logs perdidos

4. **Escalabilidade**:
   - Adição de novos nodes
   - Balanceamento de carga
   - Ajuste de recursos

## 11. Considerações de Evolução

### 11.1 Roadmap

1. **Curto Prazo** (3-6 meses):
   - Implementação de ML para detecção de anomalias
   - Melhoria na compressão e economia de armazenamento
   - Expansão de capacidades de parsing para novos formatos

2. **Médio Prazo** (6-12 meses):
   - Implementação de stream processing para analytics em tempo real
   - Auto-scaling baseado em volume de logs
   - Enriquecimento automático com dados de contexto

3. **Longo Prazo** (12+ meses):
   - Implementação de IA para classificação automática de logs
   - Análise preditiva para detecção de incidentes
   - Correlação automática entre logs, métricas e traces

### 11.2 Oportunidades de Melhorias

- **Otimização de Performance**: Migração para plugins de alta performance (C/C++)
- **Redução de Overhead**: Sampling inteligente baseado em conteúdo
- **Integração Aprimorada**: Conectores nativos para novas tecnologias
- **Automação**: Configuração dinâmica baseada em padrões de uso
- **ML/AI**: Análise avançada para insights operacionais

## 12. Referências

1. [Fluentd Documentation](https://docs.fluentd.org/)
2. [Kubernetes Logging Architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/)
3. [Fluentd Plugins](https://www.fluentd.org/plugins)
4. [Unified Logging Layer](https://www.fluentd.org/blog/unified-logging-layer)
5. [PCI DSS 4.0 Requirements](https://www.pcisecuritystandards.org/)
6. [GDPR Compliance](https://gdpr.eu/compliance/)
7. [ISO 27001 Information Security](https://www.iso.org/isoiec-27001-information-security.html)
8. [NIST SP 800-53](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
9. [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)
10. [Logging Best Practices](https://www.scalyr.com/blog/log-formatting-best-practices-readable/)

## 13. Anexos

### 13.1 Glossário

| Termo | Definição |
|-------|-----------|
| **Buffer** | Armazenamento temporário para garantir entrega confiável de logs |
| **Chunk** | Unidade de processamento de logs no buffer |
| **Dual-write** | Estratégia de envio para múltiplos destinos simultaneamente |
| **Parser** | Plugin que extrai campos estruturados de logs raw |
| **Formatter** | Plugin que formata logs estruturados para saída |
| **Tag** | Identificador usado para roteamento de logs |
| **Label** | Metadados para categorização de logs |
| **Filter** | Plugin que modifica ou filtra logs em trânsito |

### 13.2 Arquitetura de Plugins

| Categoria | Plugins Utilizados | Função |
|-----------|-------------------|--------|
| **Input** | tail, forward, http, syslog | Coleta de logs |
| **Parser** | json, regexp, multiline, grok | Extração de estrutura |
| **Filter** | record_transformer, grep, mask | Modificação e filtragem |
| **Formatter** | json, csv, msgpack | Formatação para output |
| **Buffer** | file, memory | Armazenamento temporário |
| **Output** | elasticsearch, loki, s3, kafka | Destinos de entrega |

### 13.3 Exemplo de Configuração Completa

Um repositório GitOps mantém a configuração completa em:
- `infrastructure/observability/fluentd/configs/`

### 13.4 Lista de Verificação de Implementação

- [x] Implantação como DaemonSet em todos os nodes
- [x] Configuração de coleta para todos os tipos de log
- [x] Implementação de parsing para formatos conhecidos
- [x] Configuração de dual-write para Elasticsearch e Loki
- [x] Configuração de buffer persistente com retry
- [x] Implementação de mascaramento para dados sensíveis
- [x] Configuração de monitoramento e alertas
- [x] Documentação de runbooks operacionais
- [x] Teste de carga e performance
- [x] Validação de conformidade regulatória

---

*Documento aprovado pelo Comitê de Arquitetura INNOVABIZ em 28/07/2025*