# Testes de Integração do MCP-IAM - Elevação de Privilégios

## Visão Geral

Este diretório contém testes de integração end-to-end (E2E) para o sistema de elevação de privilégios do MCP-IAM, especificamente focados na integração com os hooks MCP:

- **MCP Docker**: Testes para operações sensíveis em containers e Kubernetes
- **MCP Desktop Commander**: Testes para operações sensíveis no filesystem local
- **MCP GitHub**: Testes para operações em repositórios e branches protegidos
- **MCP Figma**: Testes para operações em designs protegidos

Os testes validam aspectos fundamentais:

1. **Multi-dimensionalidade**:
   - Multi-tenant: Isolamento entre tenants
   - Multi-market: Aplicação de regulações específicas por mercado
   - Multi-contexto: Adaptação baseada no contexto de uso

2. **Conformidade e Segurança**:
   - Autenticação multi-fator (MFA) para operações críticas
   - Auditoria detalhada de todas as operações
   - Rastreabilidade completa de quem fez o quê, quando e por quê
   - Isolamento estrito entre tenants e mercados

3. **Integração com Hooks MCP**:
   - Mapeamento de comandos para escopos
   - Autorização baseada em políticas
   - Verificação de elevação
   - Revogação de privilégios

## Pré-requisitos

Para executar estes testes, você precisará:

- Go 1.19+
- Docker
- Redis
- PostgreSQL
- Acesso às dependências do projeto

## Configuração

Os testes usam Docker Testcontainers para criar ambientes isolados. As configurações podem ser personalizadas através de variáveis de ambiente:

```bash
# Observabilidade
export TEST_LOG_FILE=./mcp_iam_test.log  # Opcional: Arquivo para logging
export JAEGER_AGENT_HOST=localhost       # Opcional: Host do agente Jaeger
export JAEGER_AGENT_PORT=6831            # Opcional: Porta do agente Jaeger

# Banco de dados (opcional, testcontainers cria automaticamente)
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=testuser
export TEST_DB_PASSWORD=testpass
export TEST_DB_NAME=innovabiz_iam_test

# Redis (opcional, testcontainers cria automaticamente)
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6379
```

## Execução dos Testes

### Executar todos os testes de integração:

```bash
go test -v ./tests/integration/elevation/...
```

### Executar apenas testes específicos:

```bash
go test -v ./tests/integration/elevation/... -run TestMCPHooksIntegration/FluxoCompletoDockerElevation
go test -v ./tests/integration/elevation/... -run TestMCPHooksIntegration/IsolamentoMultiTenantDesktopCommander
```

### Pular testes de integração (útil para CI/CD rápido):

```bash
go test -v ./tests/integration/elevation/... -short
```

## Estrutura dos Testes

Cada teste segue um fluxo específico:

1. **Configuração do ambiente**:
   - Inicialização de containers (PostgreSQL, Redis)
   - Configuração de observabilidade (logging, tracing)
   - Configuração dos hooks MCP

2. **Execução do fluxo**:
   - Solicitação de elevação
   - Verificação de MFA (quando aplicável)
   - Autorização de operações
   - Auditoria de ações
   - Verificações de isolamento multi-tenant
   - Revogação de privilégios

3. **Verificações**:
   - Validação de autorização correta
   - Verificação de logs de auditoria
   - Confirmação de isolamento entre tenants
   - Validação de conformidade regulatória específica por mercado

## Conformidade com Normas e Regulações

Os testes verificam conformidade com diversas normas e regulações, incluindo:

| Mercado | Regulações |
|---------|------------|
| Angola | Angola Financial Services Authority, Angola Data Protection Law, SADC Financial Regulations |
| Brasil | Banco Central do Brasil, LGPD, CVM |
| Moçambique | Banco de Moçambique, Mozambique Financial Services Regulation, SADC Financial Regulations |
| Global | PCI DSS, ISO 27001, GDPR, SOX |

## Observabilidade

Os testes incluem instrumentação para observabilidade através de:

- **Logging estruturado**: Utilizando Zap Logger
- **Tracing distribuído**: Utilizando OpenTracing com Jaeger
- **Métricas**: Simulação de coleta de métricas para tempos de resposta e taxas de erro

Os logs são essenciais para debugar falhas nos testes e entender o fluxo exato de execução.

## Extensão dos Testes

Para adicionar novos testes:

1. Adicione novos casos de teste no formato `testXXX` em arquivos separados
2. Configure hooks MCP adicionais conforme necessário
3. Adicione verificações para regulações específicas do seu mercado
4. Garanta a cobertura de todos os fluxos críticos

## Integração com CI/CD

Para integração em pipelines CI/CD, use:

```bash
# Em CI/CD
go test -v ./tests/integration/elevation/... -short -timeout 10m
```

## Considerações de Segurança

- Os testes **não** devem ser executados em ambientes de produção
- As credenciais usadas são apenas para fins de teste
- Revise qualquer código que manipula elevação de privilégios para vulnerabilidades

## Troubleshooting

Problemas comuns e soluções:

1. **Falhas de conexão com containers**:
   - Verifique se o Docker está em execução
   - Aumente os tempos de espera para inicialização

2. **Falhas intermitentes**:
   - Use a flag `-count=3` para executar múltiplas vezes
   - Implemente lógica de retry nos testes mais instáveis

3. **Problemas de memória**:
   - Aumente a memória disponível para os containers
   - Execute testes em paralelo com `-parallel 4`

## Contribuições

Ao contribuir com novos testes:

1. Siga a estrutura existente
2. Documente casos de teste detalhadamente
3. Garanta isolamento entre testes
4. Cubra cenários negativos (falhas, revogações)
5. Valide conformidade regulatória específica

---

## Apêndice: Modelo de Ameaças

Os testes consideram as seguintes ameaças ao sistema de elevação de privilégios:

1. Acesso não autorizado entre tenants
2. Bypass de requisitos MFA
3. Elevação sem justificação adequada
4. Uso de privilégios após revogação
5. Manipulação de tokens de elevação
6. Tentativas de acesso a operações não autorizadas
7. Falha em registrar ações privilegiadas

Cada teste valida proteções contra uma ou mais destas ameaças.