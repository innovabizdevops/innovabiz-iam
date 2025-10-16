# Procedimentos de Backup e Recovery do IAM

## Introdução

Este documento descreve os procedimentos detalhados para backup e recuperação (recovery) do módulo IAM da plataforma INNOVABIZ. Estes procedimentos são essenciais para garantir a continuidade operacional, a proteção de dados críticos e a conformidade com requisitos regulatórios.

## Matriz de Componentes para Backup

| Componente | Criticidade | Frequência de Backup | Retenção | Método |
|------------|-------------|----------------------|----------|---------|
| Banco de Dados IAM | Crítica | Diária (completo), 15 min (incremental) | 7 dias (incremental), 1 ano (completo) | PostgreSQL nativo + WAL shipping |
| Configurações IAM | Alta | Após cada alteração | 1 ano | Git + Sistema de CI/CD |
| Chaves criptográficas | Crítica | Semanal + após alterações | 5 anos | HSM + cofre seguro |
| Logs de auditoria | Alta | Diária | 7 anos | Log shipping + armazenamento imutável |
| Políticas e roles | Alta | Diária + após alterações | 1 ano | Exportação + versionamento |
| Metadados de tenants | Alta | Diária | 1 ano | Exportação estruturada |

## Estratégia de Backup

### 1. Backup de Banco de Dados

#### 1.1 Backup Completo (Full)

**Frequência**: Diário, durante janela de manutenção (02:00 - 04:00)

**Procedimento**:
```bash
# Backup completo do PostgreSQL com roles, configurações e dados
pg_dump -h <db-host> -U <backup-user> -d iam_database -F custom -Z 9 -f /backup/iam/full/iam_$(date +%Y%m%d).dump

# Verificar integridade do backup
pg_restore -l /backup/iam/full/iam_$(date +%Y%m%d).dump > /dev/null

# Aplicar políticas de retenção
find /backup/iam/full/ -name "iam_*.dump" -mtime +365 -delete
```

**Destinos de armazenamento**:
1. Armazenamento primário de backup (on-premise)
2. Armazenamento secundário (região alternativa)
3. Armazenamento imutável de longo prazo (para backups mensais)

#### 1.2 Backup Incremental

**Frequência**: A cada 15 minutos

**Procedimento**:
```bash
# Configuração contínua de WAL shipping
cat << EOF >> postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'test ! -f /backup/iam/wal/%f && cp %p /backup/iam/wal/%f'
EOF

# Script de gerenciamento de WAL
/opt/innovabiz/iam/scripts/wal-management.sh --retention=7
```

**Configuração de replicação**:
```bash
# Setup de standby server com replicação física
pg_basebackup -h <primary-db-host> -D <standby-data-dir> -U <replication-user> -P -R
```

### 2. Backup de Configurações

**Frequência**: Contínua (via sistema de controle de versão) + semanal (snapshot)

**Procedimento**:
```bash
# Exportar configurações atuais
kubectl get configmap -n iam-namespace -o yaml > /backup/iam/config/configmaps_$(date +%Y%m%d).yaml
kubectl get secret -n iam-namespace -o yaml > /backup/iam/config/secrets_$(date +%Y%m%d).yaml

# Remover dados sensíveis para armazenamento (secrets são armazenados em cofre separado)
/opt/innovabiz/iam/scripts/sanitize-configs.sh /backup/iam/config/secrets_$(date +%Y%m%d).yaml
```

**Versionamento**:
- Todas as configurações são mantidas em repositório Git com histórico completo
- Alterações passam por processo de aprovação e CI/CD
- Tags de release são criadas para cada versão em produção

### 3. Backup de Chaves Criptográficas

**Frequência**: Semanal e após qualquer alteração ou rotação

**Procedimento**:
```bash
# Exportar chaves públicas e certificados
/opt/innovabiz/iam/scripts/export-certificates.sh -o /backup/iam/keys/certs_$(date +%Y%m%d).tar.gz

# Backup de material criptográfico sensível (via HSM)
/opt/innovabiz/iam/scripts/backup-hsm.sh -d /backup/iam/keys/hsm_$(date +%Y%m%d).secure

# Verificar backup de chaves
/opt/innovabiz/iam/scripts/verify-keys-backup.sh /backup/iam/keys/hsm_$(date +%Y%m%d).secure
```

**Considerações de segurança**:
- Material criptográfico privado nunca sai do HSM em texto claro
- Backups de HSM são criptografados com chaves de custódia múltipla (M-of-N)
- Armazenamento em cofre físico seguro, além do backup digital

### 4. Backup de Logs de Auditoria

**Frequência**: Contínua (near real-time) e consolidação diária

**Procedimento**:
```bash
# Configuração de Fluentd para coleta contínua
cat << EOF > fluentd-iam-audit.conf
<match iam.audit.**>
  @type copy
  <store>
    @type elasticsearch
    host audit-es-cluster
    port 9200
    index_name iam-audit-#{Time.now.strftime('%Y.%m.%d')}
    flush_interval 10s
  </store>
  <store>
    @type s3
    aws_key_id #{ENV['AWS_ACCESS_KEY_ID']}
    aws_sec_key #{ENV['AWS_SECRET_ACCESS_KEY']}
    s3_bucket audit-logs-#{ENV['ENVIRONMENT']}
    path iam/audit-logs/%Y/%m/%d/
    buffer_path /var/log/td-agent/s3
    time_slice_format %Y%m%d%H
    time_slice_wait 10m
  </store>
</match>
EOF

# Verificação diária de integridade dos logs
/opt/innovabiz/iam/scripts/verify-audit-logs.sh --date=$(date -d "yesterday" +%Y-%m-%d)
```

**Imutabilidade e Integridade**:
- Logs armazenados com WORM (Write Once Read Many) para compliance
- Assinatura digital dos arquivos de log para verificação de integridade
- Armazenamento em múltiplas localizações geográficas

## Procedimentos de Recovery

### 1. Recovery Completo do Banco de Dados

**Cenários de uso**:
- Perda completa do banco de dados
- Corrupção severa de dados
- Migração para nova infraestrutura

**Procedimento**:

1. **Preparação do ambiente de destino**:
   ```bash
   # Verificar espaço disponível
   df -h /var/lib/postgresql/data/
   
   # Parar serviços que acessam o banco de dados
   kubectl scale deployment -n iam-namespace --replicas=0 auth-service rbac-service user-management-service
   
   # Preparar servidor PostgreSQL vazio
   initdb -D /var/lib/postgresql/data/
   ```

2. **Restauração do backup completo**:
   ```bash
   # Restaurar o backup mais recente
   LATEST_BACKUP=$(ls -t /backup/iam/full/iam_*.dump | head -1)
   pg_restore -h <db-host> -U <admin-user> -d postgres -c -C $LATEST_BACKUP
   
   # Aplicar WALs para point-in-time recovery (se necessário)
   cp /backup/iam/recovery.conf /var/lib/postgresql/data/
   systemctl restart postgresql
   ```

3. **Verificação pós-restauração**:
   ```bash
   # Verificar integridade do banco
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT count(*) FROM iam_schema.users;"
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT count(*) FROM iam_schema.tenants;"
   
   # Verificar replicação (se aplicável)
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT * FROM pg_stat_replication;"
   ```

4. **Restauração de serviços**:
   ```bash
   # Iniciar serviços com modo de reconciliação 
   kubectl set env deployment -n iam-namespace auth-service RECOVERY_MODE=true
   kubectl scale deployment -n iam-namespace --replicas=1 auth-service
   
   # Monitorar logs por erros
   kubectl logs -n iam-namespace -f deployment/auth-service
   
   # Iniciar outros serviços após validação
   kubectl scale deployment -n iam-namespace --replicas=3 rbac-service user-management-service
   ```

### 2. Recuperação Point-in-Time

**Cenários de uso**:
- Erro operacional (ex: exclusão acidental de dados)
- Corrupção parcial de dados
- Recuperação após incidente de segurança

**Procedimento**:

1. **Identificar ponto de recuperação**:
   ```bash
   # Determinar timestamp para recovery
   RECOVERY_TIME="2025-05-08 14:30:00"
   
   # Identificar backup completo anterior ao ponto de recuperação
   BACKUP_FILE=$(ls -t /backup/iam/full/iam_*.dump | grep -A1 "`date -d "$RECOVERY_TIME" +%Y%m%d`" | tail -1)
   ```

2. **Preparar ambiente de recovery**:
   ```bash
   # Criar servidor dedicado para o recovery
   kubectl apply -f recovery-database-instance.yaml
   
   # Configurar recovery.conf para PITR
   cat << EOF > recovery.conf
   restore_command = 'cp /backup/iam/wal/%f %p'
   recovery_target_time = '$RECOVERY_TIME'
   recovery_target_action = 'promote'
   EOF
   ```

3. **Executar restore com PITR**:
   ```bash
   # Restaurar backup base
   pg_restore -h <recovery-db-host> -U <admin-user> -d postgres -c -C $BACKUP_FILE
   
   # Copiar recovery.conf e reiniciar para aplicar WALs
   kubectl cp recovery.conf iam-recovery-pod:/var/lib/postgresql/data/
   kubectl exec iam-recovery-pod -- systemctl restart postgresql
   
   # Monitorar progresso
   kubectl exec iam-recovery-pod -- tail -f /var/log/postgresql/postgresql.log
   ```

4. **Validar dados recuperados**:
   ```bash
   # Conectar à instância de recovery
   psql -h <recovery-db-host> -U <admin-user> -d iam_database
   
   # Executar consultas de validação
   SELECT count(*) FROM iam_schema.users;
   SELECT max(created_at) FROM iam_schema.users;
   SELECT count(*) FROM iam_schema.tenants;
   ```

5. **Extrair dados recuperados ou promover instância**:
   ```bash
   # Extrair dados específicos para importação seletiva
   pg_dump -h <recovery-db-host> -U <admin-user> -d iam_database -t iam_schema.users --data-only -f recovered_users.sql
   
   # Ou promover a instância de recovery para produção (substituição completa)
   kubectl apply -f promote-recovery-to-production.yaml
   ```

### 3. Recuperação de Configurações e Chaves

**Cenários de uso**:
- Configurações inválidas afetando o serviço
- Comprometimento de chaves criptográficas
- Migração de ambiente

**Procedimento para configurações**:

1. **Identificar versão para restauração**:
   ```bash
   # Listar snapshots de configuração disponíveis
   ls -la /backup/iam/config/
   
   # Ou identificar commit específico do Git
   git log --pretty=format:"%h - %an, %ar : %s" --graph -n 20
   ```

2. **Restaurar configurações**:
   ```bash
   # A partir de snapshot
   kubectl apply -f /backup/iam/config/configmaps_20250501.yaml
   
   # Ou a partir do Git (método preferido)
   git checkout tags/v2.5.0 -- config/
   kubectl apply -k ./config/overlays/production/
   ```

3. **Verificar aplicação das configurações**:
   ```bash
   # Verificar configmaps aplicados
   kubectl describe configmap -n iam-namespace auth-service-config
   
   # Reiniciar serviços para aplicar novas configurações
   kubectl rollout restart deployment -n iam-namespace auth-service rbac-service
   ```

**Procedimento para chaves criptográficas**:

1. **Restringir acesso durante recuperação**:
   ```bash
   # Colocar serviços em modo de manutenção
   kubectl patch deployment -n iam-namespace -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"MAINTENANCE_MODE","value":"true"}]}]}}}}'
   ```

2. **Recuperar material criptográfico**:
   ```bash
   # Restaurar de HSM usando custódia M-of-N
   /opt/innovabiz/iam/scripts/restore-hsm.sh -i /backup/iam/keys/hsm_20250501.secure
   
   # Restaurar certificados e chaves públicas
   mkdir -p /tmp/certs-restore
   tar -xzf /backup/iam/keys/certs_20250501.tar.gz -C /tmp/certs-restore
   kubectl create secret generic -n iam-namespace iam-certificates --from-file=/tmp/certs-restore --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Verificar integridade das chaves**:
   ```bash
   # Testar assinatura/verificação
   /opt/innovabiz/iam/scripts/test-crypto-keys.sh
   
   # Testar funcionalidade JWT
   /opt/innovabiz/iam/scripts/test-jwt-signing.sh
   ```

4. **Retornar ao modo operacional**:
   ```bash
   # Remover modo de manutenção
   kubectl patch deployment -n iam-namespace -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"MAINTENANCE_MODE","value":"false"}]}]}}}}'
   
   # Verificar funcionalidade
   curl -v https://iam-api.innovabiz.com/health/crypto
   ```

## Testes e Validação de Recuperação

### 1. Testes Regulares

É obrigatória a realização de testes regulares dos procedimentos de backup e recovery:

| Tipo de Teste | Frequência | Escopo |
|---------------|------------|--------|
| Teste de recuperação de BD completo | Bimestral | Restauração completa em ambiente isolado |
| Teste de PITR | Trimestral | Recuperação para ponto específico |
| Teste de recuperação de chaves | Semestral | HSM + certificados |
| Teste de disaster recovery | Anual | Cenário completo com mudança de região |

### 2. Procedimento de Teste Padronizado

1. **Preparação**:
   - Criar ambiente isolado para testes
   - Preparar dados de validação pré-conhecidos
   - Documentar estado atual para comparação

2. **Execução**:
   - Seguir procedimentos exatos de produção
   - Medir tempo de execução de cada etapa
   - Documentar problemas encontrados

3. **Validação**:
   - Verificar integridade dos dados restaurados
   - Testar funcionalidade de aplicação
   - Validar funcionamento de integrações

4. **Documentação**:
   - Registrar RTO e RPO alcançados
   - Documentar melhorias necessárias
   - Atualizar runbooks com lições aprendidas

### 3. Checklist de Validação Pós-Recovery

- [ ] Contagem de registros correspondente ao momento do backup
- [ ] Consistência referencial entre tabelas
- [ ] Usuários podem autenticar com sucesso
- [ ] Tokens são validados corretamente
- [ ] Autorizações são aplicadas conforme esperado
- [ ] Tenants estão isolados adequadamente
- [ ] Configurações estão aplicadas corretamente
- [ ] Logs de auditoria estão sendo gerados
- [ ] Métricas de monitoramento estão funcionais
- [ ] Integrações com sistemas externos estão operacionais

## Governança e Compliance

### 1. Responsabilidades

| Função | Responsabilidades |
|--------|-------------------|
| DBA | Execução e monitoramento de backups de banco de dados |
| DevOps | Backup de configurações e infraestrutura |
| Segurança | Gestão de chaves e material criptográfico |
| IAM Admin | Validação funcional após recuperação |
| Compliance | Verificação de conformidade do processo |

### 2. Documentação e Auditoria

- Todos os procedimentos de backup e recovery devem ser documentados em logs de auditoria
- Testes de recuperação devem gerar relatórios formais para auditoria
- Desvios de política de backup devem passar por processo de exceção formal
- Registros de backup/recovery devem ser mantidos por prazo definido pela política de retenção

### 3. Requisitos Regulatórios

| Região | Regulamento | Requisito Específico |
|--------|-------------|----------------------|
| UE/Portugal | GDPR | Garantia de recuperação em 72h, demonstração de integridade dos dados |
| Brasil | LGPD | Registros de auditoria por 6 meses, validação de integridade |
| Angola | PNDSB | Backup local + remoto, recuperação em 48h para dados críticos |
| EUA | HIPAA | Testes de recuperação documentados, encriptação em trânsito e repouso |

## Relatórios e Monitoramento

### 1. Relatórios Diários

- Status de sucesso/falha de todos os backups programados
- Tamanho dos backups e tendências de crescimento
- Tempo de execução de cada operação de backup
- Alertas sobre falhas ou atrasos nos processos de backup

### 2. Monitoramento Contínuo

- Verificação de integridade de backups recentes
- Validação de acessibilidade dos meios de armazenamento
- Monitoramento de replicação e arquivamento de WAL
- Alertas sobre falhas na cadeia de arquivamento contínuo

### 3. Dashboard de Backup/Recovery

- Visualização de estado atual da estratégia de backup
- Indicadores de RPO/RTO atual vs. objetivo
- Histórico de operações de restore e testes
- Status de compliance com política de backup

## Referências

- [Requisitos de Infraestrutura IAM](../04-Infraestrutura/Requisitos_Infraestrutura_IAM.md)
- [Arquitetura Técnica IAM](../02-Arquitetura/Arquitetura_Tecnica_IAM.md)
- [Guia Operacional IAM](../08-Operacoes/Guia_Operacional_IAM.md)
- [Framework de Compliance IAM](../10-Governanca/Framework_Compliance_IAM.md)

## Apêndices

### A. Template de Plano de Recovery

```markdown
# Plano de Recovery IAM - Incidente #[ID]

## Detalhes do Incidente
- **Data/Hora de Detecção**: [YYYY-MM-DD HH:MM]
- **Tipo de Incidente**: [Corrupção de Dados / Falha de Sistema / Erro Operacional]
- **Impacto**: [Crítico / Alto / Médio / Baixo]
- **Componentes Afetados**: [DB / Configurações / Chaves / Integrações]

## Objetivo de Recovery
- **Ponto de Recuperação Alvo (RPO)**: [Data/Hora]
- **Tempo de Recuperação Alvo (RTO)**: [Duração]
- **Estado Final Desejado**: [Descrição]

## Equipe de Recovery
- **Coordenador**: [Nome] - [Contato]
- **DBA**: [Nome] - [Contato]
- **DevOps**: [Nome] - [Contato]
- **Segurança**: [Nome] - [Contato]
- **Validador Funcional**: [Nome] - [Contato]

## Plano de Execução
1. **Preparação (T+0h)**
   - [ ] Estabelecer sala de crise
   - [ ] Notificar stakeholders
   - [ ] Identificar backups necessários
   
2. **Restauração (T+1h)**
   - [ ] Executar procedimento X
   - [ ] Aplicar configurações Y
   
3. **Validação (T+3h)**
   - [ ] Executar checklist de validação
   - [ ] Confirmar integridade de dados
   
4. **Ativação (T+4h)**
   - [ ] Redirecionar tráfego
   - [ ] Monitorar estabilidade
   - [ ] Comunicar conclusão

## Plano de Contingência
- **Se procedimento X falhar**: [Ação alternativa]
- **Se validação falhar**: [Procedimento de rollback]
- **Se o tempo exceder RTO**: [Abordagem simplificada]

## Aprovações
- [ ] Coordenador de Recovery
- [ ] CISO ou representante
- [ ] Gerente de Operações
```

### B. Matriz de Recovery por Cenário

| Cenário | Procedimento Principal | Alternativa | Tempo Estimado |
|---------|------------------------|-------------|----------------|
| Corrupção de dados de usuário | Recovery PITR | Restauração seletiva | 2-4 horas |
| Falha completa de banco de dados | Restauração completa + WAL | Promoção de standby | 1-2 horas |
| Comprometimento de chaves | Restauração de HSM + rotação | Geração de novas chaves | 2-3 horas |
| Disaster recovery regional | Ativação de região DR | Reconstrução em nova região | 4-8 horas |
| Recuperação após ataque | Restauração para ponto seguro + patches | Reconstrução limpa | 8-12 horas |
