# Testes do IAM RoleHandler da Plataforma INNOVABIZ

## Visão Geral

Este diretório contém testes unitários abrangentes para o `RoleHandler` da API HTTP do IAM INNOVABIZ, garantindo a conformidade com normas internacionais de segurança e governança como ISO/IEC 27001, TOGAF, COBIT, DMBOK 2.0, e PCI DSS.

Os testes foram projetados seguindo princípios de GenAI, com foco em robustez, confiabilidade, automação inteligente e resiliência, alinhados à arquitetura de integração total da plataforma INNOVABIZ.

## Estrutura de Testes

```
tests/
├── main_test.go             # Configuração geral e ponto de entrada dos testes
├── role_handler_test.go     # Testes das operações CRUD básicas
├── role_handler_permissions_test.go  # Testes de gerenciamento de permissões
├── role_handler_hierarchy_test.go    # Testes de hierarquia de funções
├── role_handler_users_test.go        # Testes de associações usuário-função
├── role_handler_middleware_test.go   # Testes de integração com middlewares
└── README.md                # Esta documentação
```

## Cobertura de Testes

A suite de testes abrange os seguintes aspectos críticos do IAM:

1. **Operações CRUD básicas**
   - Criação de funções (roles)
   - Leitura e listagem de funções
   - Atualização de funções
   - Exclusão lógica e permanente de funções
   - Clonagem de funções
   - Sincronização de funções do sistema

2. **Gerenciamento de Permissões**
   - Atribuição de permissões a funções
   - Revogação de permissões
   - Listagem de permissões diretas e herdadas
   - Verificação de permissões

3. **Gerenciamento de Hierarquia**
   - Atribuição de funções filhas
   - Remoção de funções filhas
   - Consulta de funções pais/filhas
   - Consulta de ancestrais/descendentes
   - Prevenção de ciclos na hierarquia

4. **Gerenciamento de Usuários**
   - Atribuição de função a usuário
   - Remoção de função de usuário
   - Gestão de expiração de funções
   - Consulta de usuários por função
   - Consulta de funções por usuário

5. **Integração com Middlewares**
   - Autenticação JWT
   - Autorização baseada em políticas ABAC (Open Policy Agent)
   - CORS e segurança de cabeçalhos HTTP
   - Validação de escopo de tenant (multitenancy)

6. **Validações e Segurança**
   - Validação de entrada de dados
   - Sanitização de dados sensíveis
   - Tratamento de erros consistente
   - Isolamento por tenant
   - Prevenção de ataques comuns

## Conformidade com Normas e Frameworks

Os testes foram desenvolvidos para garantir conformidade com:

- **ISO/IEC 27001**: Segurança da informação
- **TOGAF 10.0**: Arquitetura empresarial
- **COBIT 2019**: Governança de TI
- **DMBOK 2.0**: Governança de dados
- **PCI DSS**: Segurança de dados de pagamento
- **NIST Cybersecurity Framework**: Proteção de infraestrutura
- **GDPR/LGPD**: Proteção de dados pessoais
- **Open Banking/Open Finance**: Interoperabilidade financeira

## Execução dos Testes

### Pré-requisitos

- Go 1.18 ou superior
- Pacotes de teste: `testify`, `mock`, `zerolog`, `opentelemetry`

### Comandos

**Executar todos os testes:**

```bash
go test -v ./...
```

**Executar com logs detalhados:**

```bash
TEST_LOG_LEVEL=debug go test -v ./...
```

**Executar teste específico:**

```bash
go test -v -run TestCreateRole
```

**Gerar relatório de cobertura:**

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Mocks

Os testes utilizam mocks para isolar o `RoleHandler` de suas dependências externas:

- **MockRoleService**: Simula o serviço de funções
- **MockAuthMiddleware**: Simula o middleware de autenticação JWT
- **MockAuthorizationMiddleware**: Simula o middleware de autorização OPA
- **MockCORSMiddleware**: Simula o middleware CORS

## Observabilidade e Logs

Os testes incluem observabilidade via logs estruturados e tracing:

- **Logs**: Utiliza `zerolog` para logs estruturados em JSON ou console legível
- **Tracing**: Integração com OpenTelemetry para rastreamento de requisições

## Práticas de Segurança Implementadas

- Validação rigorosa de UUIDs e entradas
- Proteção contra manipulação de tenant_id
- Tratamento adequado de datas de expiração
- Verificação de autorização para operações sensíveis
- Headers de segurança CORS apropriados
- Validação de ciclos em hierarquias

## Matriz de Conformidade

| Área | Padrões/Normas | Testes Relacionados |
|------|----------------|---------------------|
| Autenticação | ISO 27001, NIST | `TestMissingUserID`, `TestCreateRoleWithMiddleware` |
| Autorização | COBIT, ISO 27001 | `TestCreateRoleWithAuthorizationDenied`, `TestHardDeleteRoleWithMiddleware` |
| Multitenancy | TOGAF, BIAN | `TestTenantScopeValidation`, `TestMissingTenantID` |
| CORS/Segurança Web | OWASP, PCI DSS | `TestOptionsRequest` |
| Hierarquia de Funções | TOGAF, RBAC | `TestHierarchyErrorHandling`, `TestGetAncestorRoles` |
| Gestão de Permissões | ISO 27001, COBIT | `TestCheckRoleHasPermission`, `TestGetAllRolePermissions` |
| Gestão de Usuários | GDPR, ISO 27001 | `TestAssignRoleToUser`, `TestUpdateUserRoleExpiration` |

---

© 2025 INNOVABIZ - Suíte de Sistema de Governança Aumentada de Inteligência Empresarial Integrado de IA Generativa