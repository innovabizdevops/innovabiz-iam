# Exemplos para Teste de Políticas OPA do RoleService

## Visão Geral

Este diretório contém exemplos de arquivos JSON para simular e testar as políticas de autorização OPA (Open Policy Agent) do módulo RoleService da plataforma INNOVABIZ. Estes exemplos podem ser utilizados para testar diferentes cenários de autorização e validar o comportamento esperado das políticas em diversos contextos.

| Metadata | Valor |
|----------|-------|
| Versão | 1.0.0 |
| Status | Implementação |
| Classificação | Interno |
| Data Criação | 2025-08-05 |
| Autor | INNOVABIZ IAM Team |
| Aprovado por | Eduardo Jeremias |

## Objetivos

- Simular decisões de autorização em diversos cenários
- Validar o comportamento correto das políticas OPA
- Facilitar testes manuais durante o desenvolvimento
- Servir como exemplos para a implementação de novos casos de teste
- Garantir conformidade com padrões de segurança e regulamentações

## Arquivos Disponíveis

1. **super_admin_create_role.json**
   - Cenário: Super Admin criando uma função de sistema
   - Resultado esperado: Autorizado (allow: true)

2. **tenant_admin_create_role.json**
   - Cenário: Tenant Admin criando uma função customizada no seu tenant
   - Resultado esperado: Autorizado (allow: true)

3. **assign_permission.json**
   - Cenário: Tenant Admin atribuindo uma permissão a uma função
   - Resultado esperado: Autorizado (allow: true)

4. **add_role_hierarchy.json**
   - Cenário: IAM Admin criando uma relação hierárquica entre funções
   - Resultado esperado: Autorizado (allow: true)

5. **assign_role_to_user.json**
   - Cenário: Tenant Admin atribuindo uma função a um usuário
   - Resultado esperado: Autorizado (allow: true)

6. **check_user_roles.json**
   - Cenário: IAM Operator consultando as funções de um usuário
   - Resultado esperado: Autorizado (allow: true)

## Como Utilizar

### Simulação via CLI

Para testar uma política usando o CLI do OPA com estes exemplos:

```bash
# Navegar até o diretório de políticas
cd C:\Users\EDUARDO JEREMIAS\Dropbox\InnovaBiz\CoreModules\IAM\services\identity-service\policies\role

# Simular uma decisão de criação de função
opa eval --data . --input examples/super_admin_create_role.json "data.innovabiz.iam.role.crud.create_decision" --format=pretty

# Simular uma decisão de atribuição de permissão
opa eval --data . --input examples/assign_permission.json "data.innovabiz.iam.role.permissions.permission_assignment_decision" --format=pretty

# Simular uma decisão de hierarquia de funções
opa eval --data . --input examples/add_role_hierarchy.json "data.innovabiz.iam.role.hierarchy.hierarchy_addition_decision" --format=pretty

# Simular uma decisão de atribuição de função a usuário
opa eval --data . --input examples/assign_role_to_user.json "data.innovabiz.iam.role.user_assignment.role_assignment_decision" --format=pretty

# Simular uma decisão de verificação de funções de usuário
opa eval --data . --input examples/check_user_roles.json "data.innovabiz.iam.role.user_assignment.role_check_decision" --format=pretty
```

### Integração com o Makefile

O Makefile na pasta pai fornece um comando para simular decisões:

```bash
# Simular uma decisão
make simulate-decision POLICY=role/role_crud.rego INPUT=role/examples/super_admin_create_role.json
```

## Dados de Teste

Os exemplos utilizam dados fictícios mas representativos:

- **IDs de Tenant**: 
  - `10000000-0000-0000-0000-000000000001` (Tenant do Super Admin)
  - `20000000-0000-0000-0000-000000000001` (Tenant padrão)
  - `30000000-0000-0000-0000-000000000001` (Outro tenant)

- **Usuários**:
  - `00000000-0000-0000-0000-000000000001` (Super Admin)
  - `00000000-0000-0000-0000-000000000002` (Tenant Admin)
  - `00000000-0000-0000-0000-000000000003` (IAM Admin)
  - `00000000-0000-0000-0000-000000000004` (IAM Operator)
  - `00000000-0000-0000-0000-000000000005` (Usuário Regular)

- **Funções**:
  - `10000000-0000-0000-0000-000000000001` (SUPER_ADMIN)
  - `20000000-0000-0000-0000-000000000001` (TENANT_ADMIN)
  - `20000000-0000-0000-0000-000000000002` (CUSTOM_ROLE)
  - `20000000-0000-0000-0000-000000000003` (INACTIVE_ROLE)

## Conformidade e Segurança

Todos os exemplos foram criados seguindo os padrões de:

- ISO/IEC 27001:2022 (Sistema de Gestão de Segurança da Informação)
- TOGAF 10.0 (The Open Group Architecture Framework)
- COBIT 2019 (Control Objectives for Information and Related Technologies)
- NIST SP 800-53 Rev. 5 (Controles de Segurança e Privacidade)
- PCI DSS v4.0 (Payment Card Industry Data Security Standard)

## Próximos Passos

1. Criar exemplos adicionais para casos de negação esperada
2. Implementar exemplos para cenários de escalação de privilégios
3. Adicionar exemplos com múltiplos tenants
4. Integrar com testes automatizados de integração
5. Expandir para cobrir fluxos de aprovação

---

© 2025 INNOVABIZ - Todos os direitos reservados