# Sistema de Alertas Inteligentes para Open X

## Visão Geral

O Sistema de Alertas Inteligentes para Open X é uma solução projetada para monitorar, detectar e notificar sobre não-conformidades e riscos críticos em diversos domínios do ecossistema Open X, incluindo Open Banking, Open Finance, Open Insurance, Open Health e Open Government. O sistema opera de forma proativa, permitindo a detecção antecipada de problemas de conformidade e facilitando respostas rápidas para mitigar riscos econômicos, regulatórios e operacionais.

🚀 **Status**: Implementado

## Arquitetura

O sistema de alertas está implementado como um componente da plataforma INNOVABIZ e está estruturado nas seguintes camadas:

1. **Camada de Armazenamento**: Esquema PostgreSQL com tabelas para regras de alerta, alertas gerados, histórico, configurações de notificação e métricas
2. **Camada de Processamento**: Funções e procedimentos armazenados para avaliação de regras, detecção de não-conformidades e análise de tendências
3. **Camada de Notificação**: Sistema para processamento e envio de notificações por diferentes canais (email, Slack, etc.)
4. **Camada de Integração**: Conexões com dashboards, validadores de conformidade e módulos de planejamento econômico

### Diagrama de Componentes

```
┌─────────────────────────────────────┐
│   Dashboard de Conformidade Open X  │
└───────────────────┬─────────────────┘
                    │
┌───────────────────▼─────────────────┐
│    Sistema de Alertas Open X        │
├─────────────────────────────────────┤
│ ┌─────────────┐   ┌───────────────┐ │
│ │ Regras de   │   │ Processador   │ │
│ │  Alerta     │◄──┤  de Alertas   │ │
│ └─────────────┘   └───────┬───────┘ │
│                          ▲          │
│ ┌─────────────┐   ┌──────┴────────┐ │
│ │ Configuração│   │   Detector    │ │
│ │ Notificações│◄──┤ Não-Conformid.│ │
│ └──────┬──────┘   └───────┬───────┘ │
│        │                  │         │
│ ┌──────▼──────┐   ┌───────▼───────┐ │
│ │   Gestor    │   │   Analisador  │ │
│ │ Notificações│   │  de Tendência │ │
│ └─────────────┘   └───────────────┘ │
└─────────────────────────────────────┘
           │                 ▲
           │                 │
┌──────────▼─────────────────────────┐
│ Validadores de Conformidade Open X │
└──────────┬─────────────────────────┘
           │
┌──────────▼─────────────────────────┐
│    Planejamento Econômico          │
└────────────────────────────────────┘
```

## Componentes Principais

### 1. Estrutura de Dados

O sistema utiliza as seguintes estruturas principais de dados:

#### Tabelas de Configuração
- `alert_rules`: Armazena regras de alerta definidas para diferentes domínios Open X
- `notification_templates`: Templates para formatação de mensagens de notificação
- `notification_configurations`: Configurações de canais, destinatários e throttling

#### Tabelas Operacionais
- `alerts`: Registros de alertas gerados pelo sistema
- `alert_history`: Histórico de mudanças em alertas (trilha de auditoria)
- `notifications`: Notificações enviadas para diferentes canais
- `alert_metrics`: Métricas agregadas sobre alertas por período

### 2. Tipos Personalizados

O sistema utiliza tipos personalizados PostgreSQL para modelar estruturas complexas:

```sql
-- Configuração de regras de alerta
CREATE TYPE compliance_alert_system.alert_rule_config AS (
    rule_type VARCHAR(50),
    threshold_value NUMERIC,
    threshold_operator VARCHAR(2),
    look_back_minutes INTEGER,
    min_occurrences INTEGER,
    cooldown_minutes INTEGER,
    auto_resolve BOOLEAN
);

-- Destinatários de notificação
CREATE TYPE compliance_alert_system.notification_recipient AS (
    recipient_type VARCHAR(10),
    recipient_id VARCHAR(100),
    recipient_name VARCHAR(255),
    channel VARCHAR(20),
    notification_template_id VARCHAR(50)
);
```

### 3. Funções e Procedimentos Principais

#### Detecção de Alertas
- `get_critical_non_compliances`: Identifica não-conformidades críticas para um domínio
- `analyze_compliance_trend`: Analisa tendências de conformidade e deteriorações
- `evaluate_alert_rules`: Avalia todas as regras de alerta aplicáveis a um tenant

#### Processamento de Notificações
- `format_notification`: Formata uma mensagem de alerta usando um template
- `send_alert_notifications`: Envia notificações para um alerta específico

#### Geração de Alertas
- `generate_alerts`: Procedimento para gerar alertas automaticamente
- `update_alert_metrics`: Atualiza métricas agregadas de alertas

## Regras de Alerta Predefinidas

O sistema vem com regras de alerta predefinidas para diversos domínios Open X:

### Open Insurance
1. **Não-conformidades Críticas Solvência II**: Detecta requisitos com níveis críticos de risco (R3, R4)
2. **Não-conformidades Críticas SUSEP**: Identifica não-conformidades com regulações brasileiras
3. **Deterioração de Conformidade**: Alerta sobre tendências negativas de conformidade

### Open Health
1. **Não-conformidades Críticas HIPAA/GDPR**: Monitoramento de privacidade e proteção de dados
2. **Não-conformidades Críticas ANS**: Conformidade com a Agência Nacional de Saúde Suplementar
3. **Deterioração de Conformidade**: Alerta sobre tendências negativas de conformidade

### Open Government
1. **Não-conformidades Críticas eIDAS**: Monitoramento de identidade digital e assinaturas
2. **Não-conformidades Críticas Gov.br**: Conformidade com padrões Gov.br brasileiros
3. **Deterioração de Conformidade**: Alerta sobre tendências negativas de conformidade

## Templates de Notificação

O sistema implementa templates predefinidos para diferentes canais:

### Email
- **Template de Alerta Crítico**: Formatação detalhada para alertas críticos
- **Template de Alerta Alto**: Formatação para alertas de alto risco
  
### Slack
- **Template de Notificação**: Formatação em Markdown com links para dashboard

## Configuração e Personalização

### Criação de Novas Regras de Alerta

Para criar uma nova regra de alerta, use a seguinte sintaxe SQL:

```sql
INSERT INTO compliance_alert_system.alert_rules (
    rule_id,
    rule_name,
    description,
    enabled,
    open_x_domain,
    framework,
    irr_thresholds,
    alert_severity,
    rule_config,
    created_by,
    notification_groups,
    priority,
    tags
) VALUES (
    'MY_CUSTOM_RULE',
    'Minha Regra Personalizada',
    'Descrição da regra personalizada',
    TRUE,
    'OPEN_BANKING',  -- Domínio Open X aplicável
    'BACEN',         -- Framework específico ou NULL para todos
    ARRAY['R3'],     -- Níveis de IRR a monitorar
    'ALTO',          -- Severidade do alerta
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'admin',
    ARRAY['compliance_team', 'banking_team'],
    2,
    ARRAY['open_banking', 'bacen', 'custom']
);
```

### Configuração de Notificações

Para configurar novas notificações, utilize:

```sql
INSERT INTO compliance_alert_system.notification_configurations (
    config_id,
    notification_group,
    alert_severity,
    recipients,
    template_id,
    throttling_enabled,
    throttling_window_minutes,
    throttling_max_notifications
) VALUES (
    'MY_CONFIG',
    'my_team',
    'CRITICO',
    ARRAY[
        ROW('USER', 'user@example.com', 'Nome do Usuário', 'EMAIL', 'CRITICAL_ALERT_EMAIL')::compliance_alert_system.notification_recipient,
        ROW('GROUP', 'my_slack_channel', 'Meu Canal', 'SLACK', 'SLACK_ALERT')::compliance_alert_system.notification_recipient
    ],
    NULL,
    TRUE,
    60,
    5
);
```

## Permissões e Segurança

O sistema implementa um modelo de permissões baseado em roles:

- **compliance_manager_role**: Acesso total ao sistema de alertas
- **risk_analyst_role**: Acesso para visualizar e atualizar alertas
- **dashboard_viewer_role**: Acesso somente leitura para visualização

## Implementação e Integração

### Integração com Dashboards

O sistema de alertas está integrado aos dashboards de conformidade Open X através de views que expõem:

- Alertas ativos por domínio, framework e severidade
- Histórico de alertas e análises de tendências
- Impacto econômico estimado de não-conformidades

### Monitoramento Automatizado

O sistema executa monitoramento automático através de:

- Jobs agendados via pgAgent para avaliar regras de alerta
- Procedimentos armazenados para geração de alertas periódicos
- Triggers para manutenção de histórico de alterações

## Fluxo de Processamento de Alertas

1. Avaliação periódica das regras de alerta configuradas
2. Detecção de não-conformidades ou tendências negativas
3. Geração de alertas com detalhes relevantes (domínio, framework, requisitos, etc.)
4. Processamento e envio de notificações conforme configuração
5. Registro de métricas agregadas para análise e relatórios

## Exemplos de Uso

### Exemplo de Geração Manual de Alertas

```sql
-- Gerar alertas para um tenant específico
CALL compliance_alert_system.generate_alerts('550e8400-e29b-41d4-a716-446655440000'::UUID);
```

### Exemplo de Consulta de Alertas Ativos

```sql
-- Consultar alertas críticos ativos
SELECT alert_id, alert_title, open_x_domain, framework, economic_impact
FROM compliance_alert_system.alerts
WHERE alert_severity = 'CRITICO'
AND status = 'ATIVO'
ORDER BY created_at DESC;
```

## Manutenção e Monitoramento

### Atualização de Métricas

```sql
-- Atualizar métricas para os últimos 30 dias
CALL compliance_alert_system.update_alert_metrics(
    '550e8400-e29b-41d4-a716-446655440000'::UUID,
    30
);
```

### Visualização de Métricas

```sql
-- Consultar métricas de alertas por domínio
SELECT 
    open_x_domain, 
    total_alerts,
    critical_alerts,
    high_alerts,
    avg_resolution_time_minutes
FROM 
    compliance_alert_system.alert_metrics
WHERE 
    tenant_id = '550e8400-e29b-41d4-a716-446655440000'::UUID
    AND period_end > (CURRENT_DATE - INTERVAL '7 days');
```

## Considerações Finais

O Sistema de Alertas Inteligentes para Open X é um componente crítico da plataforma INNOVABIZ para garantir conformidade regulatória contínua e mitigação proativa de riscos. A natureza parametrizada do sistema permite sua adaptação para diferentes frameworks regulatórios, domínios Open X e requisitos específicos de cada tenant.

## Próximos Passos

- Implementar mecanismos de machine learning para detecção de anomalias de conformidade
- Expandir canais de notificação para incluir APIs de mensageria adicionais
- Desenvolver análise preditiva de risco com base em histórico de alertas
