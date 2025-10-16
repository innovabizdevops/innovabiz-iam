# Funcionalidades de Segurança - IAM Open X

## 1. Proteção contra DDoS

### 1.1 Componentes Principais
- Monitoramento em tempo real de requisições
- Sistema de score de risco
- Detecção de padrões de ataque
- Sistema de bloqueio automático
- Logs detalhados
- Análise de padrões distribuídos
- Monitoramento de user agents

### 1.2 Métricas de Detecção

#### 1.2.1 Requisições Suspeitas
- Limite de requisições por minuto: 800 (+45 pontos)
- Limite de IPs únicos: 40 (+25 pontos)
- Limite de requisições por endpoint: 400 (+35 pontos)
- Intervalo mínimo entre requisições: 0.1 segundos (+20 pontos)
- Número de user agents diferentes: 5 (+30 pontos)
- Ataque distribuído (10+ IPs em 5s com 100+ requisições): (+40 pontos)

### 1.3 Sistema de Score

| Fator | Pontos | Descrição |
|-------|--------|-----------|
| Requisições frequentes | 45 | Mais de 800 requisições/minuto |
| Muitos IPs | 25 | Mais de 40 IPs únicos/minuto |
| Concentração endpoint | 35 | Mais de 400 requisições/endpoint/minuto |
| Velocidade rápida | 20 | Intervalo < 0.1s entre requisições |
| Muitos user agents | 30 | Mais de 5 user agents diferentes |
| Ataque distribuído | 40 | 10+ IPs e 100+ requisições em 5s |

#### 1.2.2 Score de Risco
- Requisições muito frequentes: +45 pontos
- Muitos IPs diferentes: +25 pontos
- Ataque concentrado: +35 pontos
- Padrão conhecido: +50 pontos
- Requisições muito rápidas: +20 pontos
- Limite de bloqueio: 70 pontos

### 1.3 Recomendações Automáticas

#### Score ≥ 90 (CRÍTICO)
- Bloquear IP temporariamente por 24 horas
- Notificar equipe de segurança imediatamente
- Habilitar rate limiting estrito
- Monitorar endpoints afetados
- Iniciar investigação completa
- Revisar logs de acesso
- Verificar atividades relacionadas

#### Score 70-89 (ALTO)
- Bloquear IP temporariamente por 4 horas
- Notificar equipe de segurança
- Habilitar rate limiting
- Monitorar endpoints
- Analisar padrão de acesso

#### Score 50-69 (MODERADO)
- Avisar equipe de segurança
- Habilitar rate limiting leve
- Monitorar comportamento
- Analisar histórico
- Verificar padrões

#### Score < 50 (BAIXO)
- Monitoramento básico
- Registro de atividade
- Análise de padrões

## 2. Proteção contra SQL Injection

### 2.1 Componentes Principais
- Análise de entradas em tempo real
- Sistema de score de risco
- Detecção de padrões suspeitos
- Recomendações de segurança
- Logs de tentativas
- Monitoramento de padrões de ataque
- Análise de funções SQL maliciosas

### 2.2 Métricas de Detecção

#### 2.2.1 Padrões Suspeitos
- Caracteres especiais: +20 pontos
- Palavras-chave SQL: +30 pontos
- Comentários SQL: +20 pontos
- Caracteres de escape: +15 pontos
- Padrões de injeção: +40 pontos
- Padrões comuns: +35 pontos
- Funções SQL maliciosas: +45 pontos

### 2.3 Sistema de Score

| Fator | Pontos | Descrição |
|-------|--------|-----------|
| Caracteres especiais | 20 | ; ' " \ |
| Palavras-chave SQL | 30 | SELECT, INSERT, UPDATE, etc. |
| Comentários SQL | 20 | -- /* */ |
| Caracteres escape | 15 | \x \u |
| Padrões de injeção | 40 | Padrões conhecidos |
| Padrões comuns | 35 | OR 1=1 etc. |
| Funções SQL | 45 | SLEEP, BENCHMARK, etc. |

### 2.2 Métricas de Detecção

#### 2.2.1 Caracteres Suspeitos
- Pontuação por caracteres especiais: +20 pontos
- Pontuação por palavras-chave SQL: +30 pontos
- Pontuação por comentários: +20 pontos
- Pontuação por caracteres escape: +15 pontos
- Pontuação por padrão conhecido: +40 pontos
- Limite de bloqueio: 50 pontos

### 2.4 Recomendações de Segurança

#### Score ≥ 90 (CRÍTICO)
- Validar entrada estritamente
- Usar prepared statements obrigatórios
- Notificar equipe de segurança imediatamente
- Bloquear IP por 24 horas
- Iniciar investigação completa
- Revisar logs de acesso
- Verificar atividades relacionadas

#### Score 70-89 (ALTO)
- Validar entrada estritamente
- Usar prepared statements
- Notificar equipe de segurança
- Bloquear IP temporariamente
- Monitorar endpoint
- Analisar padrão de acesso

#### Score 50-69 (MODERADO)
- Validar entrada
- Usar prepared statements
- Monitorar comportamento
- Analisar histórico
- Verificar padrões

#### Score < 50 (BAIXO)
- Monitoramento básico
- Registro de atividade
- Análise de padrões

## 3. Sistema de Score de Risco

### 3.1 Escala de Pontuação

| Pontuação | Nível | Ações |
|-----------|-------|-------|
| 0-49 | BAIXO | Monitoramento básico |
| 50-69 | MODERADO | Autenticação adicional |
| 70-89 | ALTO | Autenticação 2FA e bloqueio temporário |
| ≥ 90 | CRÍTICO | Autenticação 2FA obrigatória e bloqueio IP por 24h |

### 3.2 Fatores de Pontuação

| Fator | Pontos | Descrição |
|-------|--------|-----------|
| IP Desconhecido | 20 | IP não registrado no sistema |
| Mudança de País | 30 | Acesso de país diferente do padrão |
| Horário Incomum | 10 | Acesso entre 0h e 6h |
| Tentativas Recentes | 25 | 3 ou mais tentativas nos últimos 15 minutos |
| Velocidade Suspeita | 35 | 5 ou mais acessos em menos de 1 minuto |
| Comportamento Anômalo | 20 | Padrão de acesso fora do normal |

#### 3.2.1 Recomendações por Nível de Risco

##### Score ≥ 90 (CRÍTICO)
- Ativar autenticação 2FA obrigatória
- Bloquear IP temporariamente por 24 horas
- Notificar administrador de segurança
- Iniciar investigação completa
- Monitorar em tempo real

##### Score 70-89 (ALTO)
- Requerer autenticação de dois fatores
- Bloquear IP temporariamente por 4 horas
- Notificar administrador
- Monitorar comportamento

##### Score 50-69 (MODERADO)
- Requerer autenticação adicional
- Bloquear IP temporariamente por 1 hora
- Monitorar acesso

##### Score < 50 (BAIXO)
- Monitoramento básico
- Registro de atividade
- Análise de padrões

#### 3.2.2 Limpeza de Logs
- Período de retenção padrão: 30 dias
- Período de retenção em emergência: 15 dias
- Logs mantidos:
  - Login_attempts
  - Access_logs
  - Security_events
  - SQL_injection_attempts
  - DDoS_attempts
  - Audit_logs

#### 3.2.3 Modo de Emergência
- Ativado quando global_attempts > 100
- Redução automática do período de retenção
- Limite de tentativas mais restritivo
- Bloqueios mais longos
- Monitoramento intensificado

## 4. Logs e Auditoria

### 4.1 Tipos de Eventos

| Tipo de Evento | Descrição |
|----------------|-----------|
| ACESSO_SUSPEITO | Acesso com score de risco ≥ 50 |
| ACESSO_NORMAL | Acesso com score de risco < 50 |
| FORÇA_BRUTA | Tentativas múltiplas de login |
| PATTERN_SUSPECT | Padrão de acesso suspeito |
| BEHAVIOR_ANOMALY | Comportamento fora do padrão |
| SQL_INJECTION | Tentativa de injeção SQL |
| DDoS_ATTACK | Ataque de negação de serviço |

### 4.2 Informações Registradas

```json
{
    "timestamp": "2025-01-01T00:00:00Z",
    "ip_address": "192.168.1.1",
    "country_code": "BR",
    "score_risco": 75,
    "event_type": "ACESSO_SUSPEITO",
    "detalhes": {
        "motivos": [
            "IP desconhecido",
            "Mudança de país",
            "Horário incomum"
        ],
        "recomendacoes": [
            "Requerer autenticação de dois fatores",
            "Bloquear IP temporariamente por 4 horas",
            "Notificar administrador"
        ]
    },
    "acoes_tomadas": [
        "2FA ativado",
        "IP bloqueado por 4h"
    ]
}
```

### 4.3 Políticas de Retenção
- Modo Normal (30 dias):
  - Logs de acesso
  - Tentativas de login
  - Eventos de segurança
  - Tentativas de injeção
  - Tentativas de DDoS
  - Logs de auditoria

- Modo Emergência (15 dias):
  - Redução automática de retenção
  - Limpeza mais frequente
  - Priorização de logs críticos

### 4.4 Limpeza Automática
- Executada periodicamente
- Baseada em risco atual
- Preserva logs críticos
- Remove logs antigos
- Mantém histórico de emergências

### 4.2 Informações Registradas
- Timestamp
- IP de origem
- Tipo de evento
- Score de risco
- Razões detectadas
- Recomendações

## 5. Integração com Sistema de Eventos

### 5.1 Tipos de Integração
- Notificação de segurança
- Monitoramento em tempo real
- Alertas automáticos
- Relatórios de segurança

### 5.2 Métricas de Performance
- Tempo de detecção
- Taxa de falsos positivos
- Taxa de bloqueios
- Tempo de resposta

## 6. Recomendações de Uso

### 6.1 Configuração Inicial
- Definir limites de score
- Configurar padrões de bloqueio
- Estabelecer regras de notificação

### 6.2 Monitoramento Contínuo
- Verificar logs diariamente
- Ajustar limites conforme necessário
- Atualizar padrões conhecidos
- Analisar falsos positivos

### 6.3 Melhorias Contínuas
- Atualizar padrões de detecção
- Ajustar scores de risco
- Implementar novas funcionalidades
- Melhorar precisão de detecção
