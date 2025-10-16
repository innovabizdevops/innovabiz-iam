# Testes de Hooks MCP-IAM - INNOVABIZ

## Visão Geral

Este diretório contém testes automatizados para os hooks de elevação de privilégios do módulo MCP-IAM (Identity and Access Management) da plataforma INNOVABIZ. Os testes verificam a conformidade com requisitos de segurança, observabilidade e regulações específicas para múltiplos mercados e tenants.

## Estrutura de Testes

Os testes estão organizados pelos tipos de hooks suportados:

- **Docker Hook**: Validação de elevação para comandos Docker
- **GitHub Hook**: Validação de elevação para operações GitHub
- **Desktop Commander Hook**: Validação de elevação para comandos do sistema operacional
- **Figma Hook**: Validação de elevação para operações no Figma
- **Testes de Integração**: Validação de fluxos completos em diferentes contextos de mercado

## Casos de Teste Cobertos

Para cada hook, os testes incluem:

1. **Validação de Escopo**: Verifica se o escopo solicitado é válido para o mercado e tenant
2. **Requisitos MFA**: Testa a determinação correta de níveis MFA necessários
3. **Obtenção de Aprovadores**: Verifica a recuperação de aprovadores conforme regras de governança
4. **Validação de Uso de Token**: Testa a validação de operações contra tokens concedidos
5. **Geração de Metadados de Auditoria**: Verifica a geração de metadados conforme requisitos regulatórios

## Cobertura Multi-Mercado e Multi-Tenant

Os testes cobrem especificidades de:

- **Angola**: Conformidade com regulações BNA, dupla aprovação, regras específicas do mercado
- **Brasil**: Conformidade com LGPD, estruturas de metadados específicas, justificativas de acesso
- **UE**: Conformidade com GDPR, base legal, minimização de dados, proteção de repositórios
- **China**: Restrições específicas de mercado, armazenamento local, requisitos de aprovação
- **Moçambique**: Regras específicas para mercados emergentes
- **BRICS**: Regras de conformidade específicas para cooperação entre países do bloco

## Mocks Utilizados

Os testes utilizam os seguintes mocks:

- `MockMetadataProvider`: Fornece metadados de conformidade e regras específicas por mercado/tenant
- `MockElevationStore`: Simula o repositório de tokens e requests de elevação
- `MockUserService`: Simula o serviço de gerenciamento de usuários e verificação MFA
- `MockAuditService`: Simula o serviço de auditoria para logging de operações

## Padrões de Observabilidade

Os testes validam a integração com padrões de observabilidade:

- **Tracing**: Integração com OpenTelemetry para rastreamento de operações
- **Logging**: Validação de logs estruturados via Zap Logger
- **Auditoria**: Verificação de metadados de auditoria específicos para cada mercado

## Fluxos de Integração Testados

Os testes de integração validam fluxos completos:

1. **Fluxo Angola**: Requisição → MFA → Dupla Aprovação → Uso de Token com metadados BNA
2. **Fluxo UE**: Requisição → MFA → Dupla Aprovação → Uso de Token com conformidade GDPR
3. **Fluxo Rejeitado Brasil**: Tentativa de elevação para escopo restrito → Rejeição apropriada
4. **Fluxo Token Expirado Moçambique**: Tentativa de uso de token expirado → Rejeição apropriada

## Execução dos Testes

Para executar os testes:

```bash
cd CoreModules/IAM
go test -v ./tests/hooks/...
```

Para verificar a cobertura de testes:

```bash
cd CoreModules/IAM
go test -cover ./tests/hooks/...
```

## Conformidade e Governança

Os testes estão alinhados com:

- **TOGAF 10.0**: Framework de arquitetura empresarial
- **COBIT 2019**: Framework de governança de TI
- **ISO/IEC 27001**: Gestão de segurança da informação
- **ISO/IEC 38500**: Governança corporativa de TI
- **GDPR/LGPD**: Regulamentos de proteção de dados

## Integrações

Os testes validam a integração com:
- **Krakend (API Gateway)**
- **MCP (Model Context Protocol)**
- **IAM (Identity and Access Management)**
- **OPA (Open Policy Agent)**

## Próximos Passos

1. Desenvolver testes de performance para validar comportamento sob carga
2. Implementar testes para novos hooks conforme adicionados à plataforma
3. Integrar com CI/CD para execução automatizada e alertas
4. Desenvolver dashboards de monitoramento de cobertura de testes

---

© 2025 INNOVABIZ - Eduardo Jeremias - Sistema de Governança Aumentada de Inteligência Empresarial Integrado de IA Generativa