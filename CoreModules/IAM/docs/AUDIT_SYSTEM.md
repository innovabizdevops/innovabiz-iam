# Sistema de Auditoria Multi-Contexto IAM - INNOVABIZ

## Visão Geral

O Sistema de Auditoria Multi-Contexto é um componente essencial do módulo IAM da plataforma INNOVABIZ, projetado para registrar, monitorar e gerar relatórios de eventos de segurança e conformidade em ambientes multi-tenant, multi-regional e multi-regulatório.

### Características Principais

- **Multi-contexto Completo**: Suporte a múltiplos tenants, regiões (BR, US, EU, AO), idiomas e moedas
- **Conformidade Regulatória**: Implementação automática de requisitos LGPD, GDPR, SOX, PCI DSS, BACEN, BNA
- **Políticas de Retenção**: Gerenciamento automatizado de retenção de dados com anonimização e exclusão
- **Relatórios de Compliance**: Geração automática de relatórios para frameworks regulatórios
- **Estatísticas e Dashboards**: Agregação de dados para análise e visualização
- **Mascaramento de Dados**: Proteção de informações sensíveis em logs de auditoria
- **Escalabilidade**: Particionamento lógico e processamento em lote para alto volume

## Arquitetura Técnica

### Componentes Principais

1. **Serviço de Auditoria** (`AuditService`):
   - Núcleo do sistema com lógica de negócio para gerenciamento de eventos
   - Implementação assíncrona com SQLAlchemy
   - Logging estruturado para rastreabilidade

2. **API REST** (`audit_router`):
   - Endpoints para criação, consulta e manipulação de eventos
   - Suporte à validação de dados com Pydantic
   - Endpoints para políticas de retenção e relatórios

3. **Banco de Dados**:
   - Modelo relacional com PostgreSQL
   - Procedimentos e funções armazenados para processamento eficiente
   - Particionamento lógico por tenant, região e período

4. **Jobs Agendados**:
   - Aplicação automática de políticas de retenção
   - Geração periódica de estatísticas
   - Geração automática de relatórios de compliance

### Modelo de Dados

#### Tabelas Principais

1. **audit_events**:
   - Eventos de auditoria com detalhes completos
   - Suporte a particionamento por tenant, região e período
   - Campos para classificação, contexto e rastreamento

2. **audit_retention_policies**:
   - Políticas de retenção por tenant, região e framework
   - Configuração de anonimização e exclusão
   - Controle de ativação e períodos

3. **audit_compliance_reports**:
   - Relatórios de compliance gerados
   - Metadata e resultados de análises
   - Status e controle de versão

4. **audit_statistics**:
   - Agregações estatísticas por período
   - Dados pré-calculados para dashboards
   - Suporte a múltiplos períodos (diário, semanal, mensal)

## Frameworks de Compliance Suportados

O sistema implementa automaticamente os requisitos dos seguintes frameworks regulatórios:

| Framework | Região | Descrição | Requisitos Implementados |
|-----------|--------|-----------|--------------------------|
| **LGPD** | Brasil | Lei Geral de Proteção de Dados | Rastreamento de acesso, retenção, anonimização |
| **GDPR** | União Europeia | General Data Protection Regulation | Privacy by design, pseudonimização, direito ao esquecimento |
| **SOX** | Estados Unidos | Sarbanes-Oxley | Trilha de auditoria, controles internos |
| **PCI DSS** | Global | Payment Card Industry Data Security Standard | Proteção de dados sensíveis, log de acesso |
| **BACEN** | Brasil | Banco Central do Brasil | Requisitos regulatórios financeiros |
| **BNA** | Angola | Banco Nacional de Angola | Requisitos específicos de Angola |
| **ISO 27001** | Global | Segurança da Informação | Controles de acesso, gestão de incidentes |

## Uso da API

### Endpoints Principais

#### Eventos de Auditoria

```
POST /audit/events - Cria um novo evento de auditoria
POST /audit/events/batch - Cria múltiplos eventos em lote
GET /audit/events/{event_id} - Recupera um evento específico
GET /audit/events - Lista eventos com filtros diversos
POST /audit/events/{event_id}/mask - Mascara campos sensíveis
```

#### Políticas de Retenção

```
POST /audit/retention-policies - Cria uma nova política
GET /audit/retention-policies - Lista políticas existentes
GET /audit/retention-policies/{policy_id} - Obtém política específica
PATCH /audit/retention-policies/{policy_id} - Atualiza política
POST /audit/apply-retention - Aplica políticas de retenção
```

#### Relatórios e Estatísticas

```
POST /audit/compliance-reports - Gera um relatório de compliance
GET /audit/compliance-reports - Lista relatórios disponíveis
GET /audit/compliance-reports/{report_id} - Obtém relatório específico
POST /audit/statistics - Gera estatísticas de auditoria
GET /audit/statistics - Recupera estatísticas existentes
```

### Exemplos de Uso

#### Criação de Evento de Auditoria

```json
POST /audit/events
{
  "category": "USER_MANAGEMENT",
  "action": "CREATE_USER",
  "description": "Criação de novo usuário",
  "resource_type": "USER",
  "resource_id": "user-123",
  "resource_name": "john.doe",
  "severity": "INFO",
  "success": true,
  "details": {"role": "admin", "department": "IT"},
  "tags": ["user", "creation"],
  "tenant_id": "tenant1",
  "regional_context": "BR",
  "country_code": "BR",
  "language": "pt-BR",
  "user_id": "admin-456",
  "user_name": "Admin User",
  "correlation_id": "corr-789",
  "source_ip": "192.168.1.1"
}
```

#### Criação de Política de Retenção

```json
POST /audit/retention-policies
{
  "tenant_id": "tenant1",
  "regional_context": "BR",
  "compliance_framework": "LGPD",
  "retention_period_days": 730,
  "policy_type": "ANONYMIZATION",
  "fields_to_anonymize": ["user_id", "user_name", "source_ip"],
  "description": "Política LGPD para anonimização de dados pessoais",
  "active": true
}
```

## Jobs Agendados

O sistema implementa três jobs agendados principais:

1. **Aplicação de Políticas de Retenção**:
   - Execução diária às 01:00
   - Processa eventos expirados conforme políticas
   - Realiza anonimização ou exclusão de dados

2. **Geração de Estatísticas**:
   - Execução diária às 02:00
   - Gera agregações para períodos diário, semanal e mensal
   - Otimizado para suportar dashboards e relatórios

3. **Geração de Relatórios de Compliance**:
   - Execução semanal aos domingos às 03:00
   - Cria relatórios para frameworks relevantes por região
   - Documenta automaticamente eventos relevantes

## Considerações de Performance

Para garantir alta performance mesmo com volumes elevados de eventos, o sistema implementa:

1. **Particionamento Lógico**:
   - Eventos particionados por tenant, região e mês
   - Indexação otimizada para consultas frequentes

2. **Processamento em Lote**:
   - Suporte a criação de eventos em lote
   - Aplicação de políticas em lotes configuráveis

3. **Funções no Banco de Dados**:
   - Procedimentos armazenados para operações intensivas
   - Minimização de tráfego de rede para operações em massa

4. **Consultas Otimizadas**:
   - Índices específicos para padrões de consulta comuns
   - Estatísticas pré-calculadas para dashboards

## Integração com Outros Módulos

O Sistema de Auditoria se integra com os seguintes módulos da plataforma INNOVABIZ:

1. **IAM Core**:
   - Eventos de autenticação e autorização
   - Mudanças de permissões e papéis

2. **Serviço de Notificação**:
   - Alertas de eventos críticos
   - Notificações de compliance

3. **Dashboards**:
   - Visualização de estatísticas
   - Relatórios de atividade

4. **Compliance Hub**:
   - Exportação para ferramentas de compliance
   - Relatórios regulatórios

## Guia de Implementação

### Pré-requisitos

- PostgreSQL 13+
- Python 3.9+
- FastAPI
- SQLAlchemy 2.0+

### Instalação

1. Execute os scripts de migração do banco de dados:
   ```
   cd src/api/app/db/migrations
   ./run_migrations.ps1
   ```

2. Configure as variáveis de ambiente:
   ```
   DATABASE_URL=postgresql+asyncpg://user:password@localhost/innovabiz_iam
   AUDIT_RETENTION_ENABLED=true
   AUDIT_BATCH_SIZE=100
   ```

3. Inicie o serviço:
   ```
   uvicorn app.main:app --reload
   ```

### Monitoramento

O sistema registra logs estruturados com campos padronizados para facilitar a observabilidade:

- **tenant_id**: Identificação do tenant
- **regional_context**: Contexto regional
- **category**: Categoria do evento
- **severity**: Nível de severidade
- **correlation_id**: ID de correlação para rastreamento

## Próximos Passos

1. Implementação de exportação para SIEM
2. Visualizações avançadas com Grafana
3. Machine Learning para detecção de anomalias
4. Suporte a regulamentações adicionais para novas regiões
5. Integração com ferramentas externas de compliance

## Contato

Para questões técnicas ou suporte:

- **Email**: innovabizdevops@gmail.com
- **Responsável**: Eduardo Jeremias
- **Versão**: 1.0 - 2025