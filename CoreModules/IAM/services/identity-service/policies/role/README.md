# Políticas OPA para RoleService

## Visão Geral

Este diretório contém as políticas de autorização baseadas em Open Policy Agent (OPA) para o RoleService da plataforma INNOVABIZ. Estas políticas implementam uma camada de autorização avançada para as operações de gestão de funções, seguindo os princípios de segurança e conformidade exigidos pelos padrões internacionais e frameworks adotados pela plataforma.

| Metadata | Valor |
|----------|-------|
| Versão | 1.0.0 |
| Status | Implementação |
| Classificação | Confidencial |
| Data Criação | 2025-08-05 |
| Última Atualização | 2025-08-05 |
| Autor | INNOVABIZ IAM Team |
| Aprovado por | Eduardo Jeremias |

## Estrutura e Organização

As políticas estão organizadas em módulos especializados por área funcional:

- **role_base.rego**: Definições base e regras comuns para todas as políticas
- **role_crud.rego**: Políticas para operações CRUD de funções
- **role_permissions.rego**: Políticas para gestão de permissões em funções
- **role_hierarchy.rego**: Políticas para gestão de hierarquias de funções
- **role_user_assignment.rego**: Políticas para atribuição de funções a utilizadores
- **role_test.rego**: Testes unitários para validação das políticas

## Princípios de Segurança Implementados

Estas políticas implementam os seguintes princípios de segurança:

1. **Princípio do Menor Privilégio**: Atribuição apenas das permissões mínimas necessárias
2. **Defesa em Profundidade**: Múltiplas camadas de verificação e validação
3. **Segregação de Deveres**: Separação de responsabilidades para operações sensíveis
4. **Auditabilidade Completa**: Registro detalhado de todas as decisões de autorização
5. **Isolamento Multitenancy**: Garantia de separação estrita entre tenants
6. **Prevenção de Escalonamento de Privilégios**: Bloqueio proativo contra tentativas de escalonamento
7. **Gestão de Risco Baseada em Contexto**: Consideração de fatores contextuais nas decisões

## Normas e Frameworks de Conformidade

As políticas foram desenvolvidas em conformidade com os seguintes padrões e frameworks:

- ISO/IEC 27001:2022 (Sistema de Gestão de Segurança da Informação)
- TOGAF 10.0 (The Open Group Architecture Framework)
- COBIT 2019 (Control Objectives for Information and Related Technologies)
- NIST SP 800-53 Rev. 5 (Controles de Segurança e Privacidade)
- PCI DSS v4.0 (Payment Card Industry Data Security Standard)
- GDPR (General Data Protection Regulation)
- APD Angola (Agência de Proteção de Dados de Angola)
- BNA (Banco Nacional de Angola) - Avisos e regulamentações de cibersegurança
- Basel III/IV (Regulamentações bancárias internacionais)

## Matriz de Decisões de Autorização

A tabela abaixo resume as principais decisões de autorização implementadas nas políticas:

| Operação | SUPER_ADMIN | TENANT_ADMIN | IAM_ADMIN | IAM_OPERATOR | Usuário Regular |
|----------|-------------|--------------|-----------|--------------|-----------------|
| Criar Função SYSTEM | ✅ | ❌ | ❌ | ❌ | ❌ |
| Criar Função CUSTOM | ✅ | ✅ | ✅ | ❌ | ❌ |
| Ler Função | ✅ | ✅ | ✅ | ✅ | ✅ (com permissão) |
| Atualizar Função SYSTEM | ✅ | ❌ | ❌ | ❌ | ❌ |
| Atualizar Função CUSTOM | ✅ | ✅ | ✅ | ❌ | ❌ |
| Excluir Função SYSTEM | ❌ | ❌ | ❌ | ❌ | ❌ |
| Excluir Função CUSTOM | ✅ | ✅ | ✅ | ❌ | ❌ |
| Hard Delete de Função | ✅ | ❌ | ✅ (com permissão) | ❌ | ❌ |
| Atribuir Permissão Normal | ✅ | ✅ | ✅ | ❌ | ❌ |
| Atribuir Permissão Crítica | ✅ | ❌ | ❌ | ❌ | ❌ |
| Revogar Permissão | ✅ | ✅ | ✅ | ❌ | ❌ |
| Adicionar Hierarquia | ✅ | ✅ | ✅ | ❌ | ❌ |
| Remover Hierarquia | ✅ | ✅ | ✅ | ❌ | ❌ |
| Atribuir Função Normal | ✅ | ✅ | ✅ | ✅ | ❌ |
| Atribuir Função Sensível | ✅ | ✅ | ❌ | ❌ | ❌ |
| Remover Atribuição | ✅ | ✅ | ✅ | ✅ | ❌ |
| Verificar Permissão | ✅ | ✅ | ✅ | ✅ | ✅ (com permissão) |

## Integração com o Sistema

### Fluxo de Autorização

1. **Solicitação HTTP**: Cliente faz requisição para o endpoint do RoleService
2. **Middleware de Autenticação**: Valida o token JWT e extrai claims
3. **Middleware de Autorização**: 
   - Prepara o input para a política OPA
   - Envia o input ao servidor OPA
   - Recebe a decisão de autorização
   - Permite ou bloqueia a requisição com base na decisão
4. **Handler**: Processa a requisição se autorizada
5. **Logging e Auditoria**: Registra detalhes da decisão de autorização

### Estrutura do Input para OPA

```json
{
  "user": {
    "id": "uuid",
    "tenant_id": "uuid",
    "roles": ["ROLE1", "ROLE2"],
    "permissions": ["perm1", "perm2"]
  },
  "resource": {
    "type": "roles",
    "id": "uuid",
    "tenant_id": "uuid",
    "data": {
      "name": "string",
      "type": "SYSTEM|CUSTOM",
      "status": "ACTIVE|INACTIVE|DELETED"
    }
  },
  "method": "GET|POST|PUT|DELETE",
  "context": {
    "requests_per_minute": 10,
    "geo": {
      "country": "AO",
      "region": "Luanda"
    }
  }
}
```

### Estrutura da Resposta do OPA

```json
{
  "allow": true|false,
  "reason": "string explicando o motivo da decisão"
}
```

## Melhores Práticas de Desenvolvimento

1. **Imutabilidade**: As políticas são tratadas como código imutável e versionado
2. **Testabilidade**: Todos os cenários de decisão possuem testes unitários
3. **Legibilidade**: Código bem estruturado e comentado para facilitar manutenção
4. **Desempenho**: Otimização para decisões rápidas, com suporte a cache
5. **Rastreabilidade**: Cada decisão inclui a razão para facilitar diagnósticos
6. **Centralização**: Política centralizada para evitar lógica de autorização dispersa
7. **Validação Contínua**: Integração com CI/CD para validação automática

## Executando Testes e Validação

O Makefile fornecido no diretório inclui comandos para validar e testar as políticas:

```bash
# Validar sintaxe de todas as políticas
make validate-policies

# Executar todos os testes unitários
make test-policies

# Gerar relatório de cobertura de testes
make policy-coverage

# Executar teste específico
make test-policy POLICY=role_crud_test.rego
```

## Governança e Ciclo de Vida

1. **Revisão**: As políticas são revisadas trimestralmente ou após mudanças significativas
2. **Aprovação**: Modificações requerem aprovação do IAM Lead e Security Officer
3. **Publicação**: Políticas aprovadas são publicadas no servidor OPA via CI/CD
4. **Monitoramento**: Métricas de decisões são coletadas para análise contínua
5. **Auditoria**: Registros de auditoria são preservados conforme política de retenção

## Próximos Passos

- Implementação de políticas para detecção de anomalias baseadas em ML
- Expansão do suporte a decisões baseadas em atributos dinâmicos (ABAC avançado)
- Integração com sistemas externos de gerenciamento de risco
- Desenvolvimento de console de administração para políticas
- Implementação de análise preditiva para risco de autorização

---

© 2025 INNOVABIZ - Todos os direitos reservados