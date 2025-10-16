# ADR-040: Microserviço de Gestão de Identidade Multi-Contexto

## Status

Proposto

## Contexto

A plataforma INNOVABIZ necessita de um sistema robusto de gestão de identidade que suporte múltiplos contextos de uso (financeiro, saúde, governamental, e-commerce, etc.), permitindo que uma identidade básica seja enriquecida com atributos específicos de cada contexto, mantendo conformidade com regulamentações internacionais (GDPR, LGPD, POPIA) e regionais (SADC, CPLP, PALOP's).

O atual sistema IAM possui limitações ao lidar com identidades em múltiplos contextos, gerando fragmentação de dados, inconsistências e dificuldades na gestão de consentimento por contexto.

## Decisão

Implementaremos um microserviço dedicado à gestão de identidade multi-contexto que:

1. **Adotará arquitetura baseada em domínios**:
   - Core Domain: Identidade base (atributos universais)
   - Subdomínios: Contextos específicos (financeiro, saúde, governo, etc.)
   - Domínios de Suporte: Consentimento, verificação, auditoria

2. **Utilizará modelo de dados hierárquico e extensível**:
   - Modelo base de identidade (universal)
   - Extensões de identidade por contexto
   - Mapeamentos de atributos entre contextos

3. **Implementará padrão de federação de identidade avançado**:
   - Suporte a múltiplos protocolos (OAuth 2.1, OIDC, SAML 2.0)
   - Gerenciamento de claims por contexto
   - Mapeamento dinâmico de atributos

4. **Integrará profundamente com TrustGuard**:
   - Pontuação de confiabilidade por contexto
   - Políticas de verificação contextuais
   - Atestados de identidade personalizados

5. **Adotará arquitetura técnica**:
   - Backend em Go para core de identidade (performance crítica)
   - APIs GraphQL para consultas complexas contextuais
   - Event-driven para propagação de mudanças entre contextos
   - Cache distribuído para performance

## Componentes Principais

### 1. Core de Identidade

```
- identity-core/
  - domain/
    - models/
      - base_identity.go
      - identity_context.go
      - context_attributes.go
    - repositories/
      - identity_repository.go
    - services/
      - identity_service.go
  - application/
    - commands/
      - create_identity.go
      - add_identity_context.go
    - queries/
      - get_identity.go
      - list_identity_contexts.go
  - infrastructure/
    - persistence/
      - postgres/
        - identity_repository_impl.go
    - api/
      - graphql/
        - resolvers/
          - identity_resolver.go
      - grpc/
        - identity_service.go
  - ports/
    - http/
      - rest/
        - identity_controller.go
      - graphql/
        - schema.graphql
```

### 2. Gestão de Contextos

```
- context-management/
  - domain/
    - models/
      - context_definition.go
      - context_mapping.go
      - context_policy.go
    - services/
      - context_service.go
  - infrastructure/
    - repositories/
      - context_repository.go
    - mappings/
      - attribute_mapper.go
  - application/
    - use_cases/
      - create_context.go
      - map_contexts.go
      - apply_policy.go
```

### 3. Serviço de Consentimento Contextual

```
- contextual-consent/
  - domain/
    - models/
      - consent.go
      - consent_scope.go
      - consent_context.go
    - services/
      - consent_service.go
  - application/
    - commands/
      - grant_consent.go
      - revoke_consent.go
    - queries/
      - validate_consent.go
  - infrastructure/
    - repositories/
      - consent_repository.go
    - api/
      - consent_controller.go
```

### 4. Sistema de Auditoria Multi-Contexto

```
- multi-context-audit/
  - domain/
    - models/
      - audit_event.go
      - context_access.go
    - services/
      - audit_service.go
  - infrastructure/
    - persistence/
      - audit_repository.go
    - messaging/
      - audit_publisher.go
  - application/
    - use_cases/
      - log_context_access.go
      - generate_audit_report.go
```

### 5. API Gateway Contextual

```
- contextual-gateway/
  - api/
    - graphql/
      - schema/
        - identity_schema.graphql
        - context_schema.graphql
      - resolvers/
        - identity_resolver.go
        - context_resolver.go
    - rest/
      - controllers/
        - identity_controller.go
        - context_controller.go
  - middleware/
    - context_authorization.go
    - context_validation.go
  - adapters/
    - service_clients/
      - identity_client.go
      - context_client.go
```

## Integração com Outros Módulos

1. **IAM Core**: Autenticação primária e autorização básica
2. **TrustGuard**: Verificação de identidade e pontuação de confiabilidade
3. **Mobile Money**: Verificação em contexto financeiro
4. **E-Commerce**: Verificação em contexto comercial
5. **Bureau de Créditos**: Dados de identidade financeira
6. **UniConnect**: Notificações multi-contexto

## Modelo de Dados

### Entidade Base: Identity

```sql
CREATE TABLE identities (
    identity_id UUID PRIMARY KEY,
    primary_key_type VARCHAR(50) NOT NULL,  -- cpf, passport, etc.
    primary_key_value VARCHAR(100) NOT NULL,
    master_person_id UUID,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    metadata JSONB
);

CREATE UNIQUE INDEX idx_identity_primary_key ON identities(primary_key_type, primary_key_value);
```

### Contextos

```sql
CREATE TABLE identity_contexts (
    context_id UUID PRIMARY KEY,
    identity_id UUID NOT NULL REFERENCES identities(identity_id),
    context_type VARCHAR(50) NOT NULL,  -- financial, health, government, etc.
    context_status VARCHAR(20) NOT NULL,
    trust_score DECIMAL,
    verification_level VARCHAR(20),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(identity_id, context_type)
);
```

### Atributos Contextuais

```sql
CREATE TABLE context_attributes (
    attribute_id UUID PRIMARY KEY,
    context_id UUID NOT NULL REFERENCES identity_contexts(context_id),
    attribute_key VARCHAR(100) NOT NULL,
    attribute_value TEXT,
    sensitivity_level VARCHAR(20) NOT NULL,
    verification_status VARCHAR(20) NOT NULL,
    verification_source VARCHAR(100),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(context_id, attribute_key)
);
```

### Consentimentos Contextuais

```sql
CREATE TABLE context_consents (
    consent_id UUID PRIMARY KEY,
    identity_id UUID NOT NULL REFERENCES identities(identity_id),
    context_id UUID NOT NULL REFERENCES identity_contexts(context_id),
    purpose VARCHAR(100) NOT NULL,
    third_party VARCHAR(100),
    granted_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    consent_proof TEXT,
    UNIQUE(context_id, purpose, third_party)
);
```

## Considerações de Segurança

1. **Isolamento de Dados Contextuais**:
   - Criptografia em nível de contexto
   - Chaves diferentes por contexto
   - Políticas de acesso granulares

2. **Consentimento Contextual**:
   - Gerenciamento de consentimento específico por contexto
   - Rastreamento de uso de dados por contexto
   - Interface de gerenciamento para usuário final

3. **Auditoria Multi-Nível**:
   - Registro de acesso por contexto
   - Rastreamento de alterações em atributos contextuais
   - Relatórios de auditoria específicos por contexto

## Alternativas Consideradas

1. **Modelo de identidade monolítico**:
   - Rejeitado por falta de flexibilidade e escalabilidade
   - Dificuldade em atender requisitos regulatórios específicos de cada contexto

2. **Identidades totalmente separadas por contexto**:
   - Rejeitado por criar silos de dados
   - Duplicação de informações e inconsistências

3. **Sistema baseado apenas em federação**:
   - Rejeitado por depender de provedores externos
   - Limitações em requisitos específicos de negócio

## Conformidade Regulatória

O microserviço foi projetado considerando os seguintes requisitos de conformidade:

1. **GDPR/LGPD/POPIA**:
   - Minimização de dados por contexto
   - Direito ao esquecimento contextual
   - Portabilidade de dados específica de contexto

2. **Regulamentações Setoriais**:
   - Saúde: HIPAA, regulamentos locais
   - Finanças: PCI DSS, Basel II/III, regulamentações locais
   - Governo: eID regulamentações regionais

3. **Requisitos Regionais**:
   - SADC: Requisitos de identidade transfronteiriça
   - CPLP: Regulamentações específicas de países lusófonos
   - PALOP's: Requisitos específicos de Angola e outros países

## Consequências

### Positivas

1. Flexibilidade para suportar múltiplos contextos de identidade
2. Melhor conformidade regulatória por segmentação contextual
3. Redução de duplicação de dados entre contextos
4. Escalabilidade para adicionar novos contextos sem alterar arquitetura base
5. Modelo de segurança granular por contexto
6. Suporte a múltiplos mercados e regiões regulatórias

### Negativas

1. Aumento da complexidade arquitetural
2. Maior overhead de desenvolvimento inicial
3. Necessidade de orquestração complexa entre contextos
4. Possíveis desafios de performance em consultas multi-contexto

## Implementação

A implementação será faseada:

1. **Fase 1**: Modelo de dados base e APIs core
2. **Fase 2**: Gestão de contexto e mapeamentos
3. **Fase 3**: Consentimento contextual
4. **Fase 4**: Integração com TrustGuard e outros módulos
5. **Fase 5**: Auditoria multi-contexto

## Monitoramento e Métricas

1. **Performance**:
   - Tempo médio de resolução de identidade por contexto
   - Latência em consultas multi-contexto

2. **Segurança**:
   - Tentativas de acesso cross-contexto não autorizadas
   - Anomalias em padrões de acesso contextual

3. **Negócio**:
   - Taxa de utilização por contexto
   - Completude de perfil por contexto

## Aprovação

- [ ] Arquiteto de Segurança
- [ ] Líder Técnico IAM
- [ ] Oficial de Proteção de Dados
- [ ] Gerente de Produto IAM

## Referências

1. ISO/IEC 24760-1:2019 - Framework de gestão de identidade
2. NIST SP 800-63-3 - Diretrizes de identidade digital
3. Padrões OpenID Connect para setores específicos
4. Enterprise IAM Architecture (Gartner)