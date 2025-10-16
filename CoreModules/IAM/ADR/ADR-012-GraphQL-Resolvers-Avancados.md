# ADR-012: Implementação de Resolvers GraphQL Avançados para IAM Multi-Tenant

## Status

Aprovado

## Contexto

A implementação do módulo IAM da plataforma INNOVABIZ requer uma camada de API GraphQL robusta e flexível para gerenciar identidades, grupos, papéis, permissões e eventos de segurança em um ambiente multi-tenant com alta observabilidade, auditoria e conformidade com múltiplas regulamentações internacionais.

É necessário definir como os resolvers GraphQL serão estruturados para atingir os seguintes objetivos:
1. Garantir isolamento multi-tenant rigoroso
2. Implementar controle de acesso granular 
3. Facilitar auditoria e observabilidade
4. Permitir escalabilidade horizontal
5. Garantir compliance com regulamentações internacionais

## Decisão

Implementar resolvers GraphQL especializados seguindo um padrão modular que separa queries, mutations e subscriptions por domínio (usuários, grupos, papéis, permissões, tenants e eventos de segurança), com recursos avançados integrados:

### Estrutura de Resolvers

1. **Organização por Domínio**:
   - `user_resolver.go` - Operações relacionadas a usuários
   - `group_resolver.go` - Queries para grupos
   - `group_resolver_mutations.go` - Mutations para grupos
   - `role_resolver.go` - Operações para papéis (roles)
   - `permission_resolver.go` - Queries para permissões
   - `permission_resolver_mutations.go` - Mutations para permissões
   - `tenant_resolver.go` - Operações para tenants
   - `security_event_resolver.go` - Operações para eventos de segurança
   - `statistics_resolver.go` - Operações para estatísticas e dashboards

2. **Controle de Acesso Multi-tenant**:
   - Verificação de contexto de tenant em todas as operações
   - Validação rigorosa de cross-tenant access para operações entre tenants
   - Filtros automáticos baseados em tenant para consultas de listagem
   - Permissões específicas para operações de sistema vs. operações regulares

### Recursos Transversais Integrados

1. **Observabilidade**:
   - Integração com OpenTelemetry para distributed tracing
   - Spans com atributos detalhados para cada operação
   - Métricas de performance e utilização

2. **Auditoria**:
   - Logging estruturado com contexto de tenant e usuário
   - Publicação de eventos de segurança para operações críticas
   - Rastreabilidade completa das ações

3. **Validação e Segurança**:
   - Diretivas GraphQL para validação, autenticação e redação de dados sensíveis
   - Verificação de acesso baseada em papéis e permissões específicas
   - Sanitização de inputs e defesa contra injeção

4. **Controle de Recursos**:
   - Paginação configurável para consultas de listagem
   - Limitação de tamanho de página para prevenção de DoS
   - Timeouts configuráveis

## Consequências

### Positivas

1. **Segurança e Isolamento**: A validação de tenant em cada operação garante que não há vazamento de dados entre tenants.

2. **Observabilidade**: A integração com OpenTelemetry permite rastreamento distribuído completo e facilita a identificação de gargalos e problemas.

3. **Auditoria Detalhada**: O logging estruturado e eventos de segurança proporcionam uma trilha de auditoria completa para compliance.

4. **Controle Granular**: A verificação de permissões específicas para cada operação garante aplicação precisa dos princípios de menor privilégio.

5. **Manutenibilidade**: A estrutura modular facilita a manutenção e evolução do código.

6. **Desempenho Otimizado**: Os mecanismos de paginação e limitação de recursos previnem sobrecarga do sistema.

7. **Flexibilidade Multi-regional**: A estrutura suporta requisitos específicos por região.

### Negativas

1. **Complexidade**: A implementação introduz complexidade adicional devido aos múltiplos recursos transversais.

2. **Overhead de Processamento**: A validação constante de tenants e permissões adiciona algum overhead.

3. **Curva de Aprendizado**: Desenvolvedores novos precisarão entender padrões específicos para contribuir efetivamente.

## Alternativas Consideradas

1. **REST API Tradicional**:
   - Rejeitada devido à falta de flexibilidade para buscar apenas os dados necessários
   - Demandaria múltiplas chamadas para operações complexas

2. **Resolvers Monolíticos**:
   - Rejeitada por dificultar a manutenção e escalabilidade do código

3. **Separação em Microserviços por Domínio**:
   - Adiada para uma fase futura de escala, mas arquitetura atual permite esta evolução

## Conformidade

Esta decisão está alinhada com:

1. Padrões ISO/IEC 27001:2022 para segurança da informação
2. GDPR, LGPD e outras regulamentações de privacidade
3. Princípios NIST para controle de acesso
4. Padrões de observabilidade OpenTelemetry
5. Práticas de desenvolvimento GraphQL da GraphQL Foundation

## Validadores de Conformidade

1. Isolamento de tenant deve ser validado por testes automatizados
2. Auditoria completa deve ser verificada para todas as operações críticas
3. Performance deve ser monitorada para garantir SLAs 
4. Segurança deve ser verificada por análise estática e testes de penetração

## Implementação

A implementação dos resolvers segue este padrão consistente:

1. Verificação de contexto de autenticação
2. Validação de permissões específicas
3. Verificação de tenant matching ou cross-tenant permission
4. Logging estruturado de auditoria
5. Execução da lógica de negócios via serviço
6. Publicação de eventos de segurança para auditoria
7. Controle de resposta e tratamento de erros

Exemplo de implementação (CreateUser):

```go
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // Observabilidade - iniciando span
    ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createUser")
    defer span.End()

    // Verificação de autenticação
    authInfo := auth.GetAuthInfoFromContext(ctx)
    if authInfo == nil {
        return nil, errors.ErrUnauthorized
    }

    // Verificação de permissão específica
    if !authInfo.HasPermission("IAM:ManageUsers") {
        r.logger.Warn(ctx, "Permission denied for creating user", 
            "requester_id", authInfo.UserID)
        return nil, errors.NewForbiddenError(...)
    }

    // Auditoria - logging estruturado
    r.logger.Info(ctx, "GraphQL mutation: createUser", 
        "input", sanitizedInput,
        "requester_id", authInfo.UserID)

    // Validação de tenant
    if input.TenantID != authInfo.TenantID && !authInfo.HasPermission("IAM:CrossTenantAccess") {
        return nil, errors.NewForbiddenError(...)
    }

    // Lógica de negócios via serviço
    user, err := r.userService.Create(ctx, &input)
    
    // Publicação de evento de segurança para auditoria
    r.securityService.LogEvent(...)

    return user, nil
}
```

## Referências

1. GraphQL Best Practices - https://graphql.org/learn/best-practices/
2. OpenTelemetry Tracing - https://opentelemetry.io/docs/concepts/signals/traces/
3. OWASP API Security Top 10 - https://owasp.org/www-project-api-security/
4. Multi-tenant Data Isolation Patterns - https://docs.microsoft.com/azure/architecture/guide/multitenant/service/data-isolation

## Autores

- Time de Desenvolvimento IAM - INNOVABIZ
- Comitê de Arquitetura - INNOVABIZ

Data: 06/08/2025