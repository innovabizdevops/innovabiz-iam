# Padrões de Segurança para APIs do Bureau de Créditos

**Autor:** Eduardo Jeremias - InnovaBiz  
**Data:** 19/08/2025  
**Versão:** 1.0  
**Classificação:** Restrito

## 1. Introdução

Este documento estabelece os padrões de segurança obrigatórios para todas as APIs que interagem com os serviços de Bureau de Créditos, em conformidade com regulamentações internacionais e melhores práticas de segurança. Estes padrões se aplicam a todos os endpoints GraphQL e REST expostos pelo módulo IAM para interação com sistemas de Bureau de Créditos.

## 2. Princípios Fundamentais

### 2.1. Zero Trust Architecture

Todas as APIs seguem o princípio de Zero Trust, onde nenhuma solicitação é confiável por padrão, independentemente da sua origem. Cada solicitação deve:

- Ser autenticada completamente
- Ser autorizada explicitamente
- Ter seus privilégios verificados
- Ser registrada para auditoria

### 2.2. Defense in Depth

Múltiplas camadas de segurança são implementadas para proteger dados sensíveis:

1. Camada de rede (firewalls, WAF)
2. Camada de aplicação (validação de entrada)
3. Camada de API Gateway (Krakend)
4. Camada de serviço (lógica de negócios)
5. Camada de dados (criptografia e isolamento)

### 2.3. Privacy by Design

Privacidade é incorporada na arquitetura desde o início:

- Minimização de dados (apenas dados necessários são coletados)
- Propósito explícito para cada consulta
- Consentimento verificável e rastreável
- Limitação de uso aos fins declarados## 3. Autenticação e Autorização

### 3.1. Tokens JWT

Todas as APIs utilizam tokens JWT com as seguintes características:

- Assinados com algoritmo RS256 (chaves assimétricas)
- Tempo de validade máximo de 60 minutos
- Inclusão de claims padronizados (iss, sub, exp, iat, jti)
- Claims personalizados para controle granular:
  - `tenant_id`: Identificador do tenant
  - `scope`: Escopos específicos de permissão
  - `purpose`: Finalidade declarada da consulta
  - `auth_id`: ID da autorização vinculada

#### Exemplo de Payload JWT:

```json
{
  "iss": "innovabiz-iam",
  "sub": "user-12345",
  "exp": 1629383825,
  "iat": 1629380225,
  "jti": "d9b1c3fa-85eb-42fb-8967-3a479510698b",
  "tenant_id": "tenant-angola-001",
  "scope": "bureau:score:read bureau:history:read",
  "purpose": "credit_evaluation",
  "auth_id": "auth-7890"
}
```

### 3.2. Rotação de Tokens

Implementamos rotação automática de tokens com as seguintes características:

- Refresh tokens com validade de 24 horas
- Possibilidade de revogação imediata
- Detecção de reutilização de tokens expirados
- Rotação completa a cada uso (tanto access token quanto refresh token)

### 3.3. Controle de Acesso Baseado em Atributos (ABAC)

O controle de acesso incorpora múltiplos atributos:

- Identidade do usuário
- Função/papel do usuário
- Tipo de vinculação com o Bureau
- Nível de acesso concedido
- Contexto da solicitação (hora, localização, dispositivo)
- Finalidade declarada## 4. Proteção de Dados

### 4.1. Criptografia em Trânsito

- TLS 1.3 obrigatório para todas as comunicações
- Cipher suites seguras conforme recomendações NIST
- Certificate pinning para comunicações críticas
- Forward secrecy para garantir segurança futura das comunicações

### 4.2. Criptografia em Repouso

- Dados de autorização: AES-256-GCM
- Registros de consultas: AES-256-CBC com rotação de chaves
- Tokens e credenciais: PBKDF2 com salt único
- Chaves armazenadas em HSM ou gerenciador de segredos

### 4.3. Mascaramento de Dados Sensíveis

- Números de documentos (CPF, BI, NIF): parcialmente mascarados
- Dados financeiros: totalmente mascarados em logs
- Informações de contato: parcialmente mascaradas
- Histórico de crédito: dados agregados quando possível

## 5. Proteção da API

### 5.1. Validação de Entradas

- Validação de esquema GraphQL rigorosa
- Sanitização de todas as entradas para prevenir injeções
- Validação de tipo, formato e tamanho para todos os parâmetros
- Limitação de profundidade e complexidade em consultas GraphQL

### 5.2. Rate Limiting

- Limites por tenant (configurável)
- Limites por usuário (ajustável por nível de acesso)
- Limites por IP de origem
- Controle de rajadas (burst control)
- Penalidade progressiva para excesso de tentativas

#### Configuração de Rate Limiting no Krakend:

```json
"extra_config": {
  "github_com/devopsfaith/krakend-ratelimit/juju/router": {
    "max_rate": 100,
    "client_max_rate": 10,
    "strategy": "ip",
    "capacity": 100
  }
}
```

### 5.3. Proteções Adicionais

- Proteção contra ataques de brute force
- Detecção de padrões anômalos de consulta
- Monitoramento de comportamento suspeito
- Circuit breakers para proteção contra falhas em cascata## 6. Auditoria e Logs

### 6.1. Eventos Auditáveis

Todos os seguintes eventos são registrados para auditoria:

- Criação/modificação/revogação de vínculos com Bureau
- Criação de autorizações de consulta
- Geração de tokens de acesso
- Todas as consultas realizadas ao Bureau
- Revogação de autorizações e tokens
- Alterações de configuração de segurança

### 6.2. Formato de Logs

Logs estruturados em JSON com os seguintes campos obrigatórios:

```json
{
  "timestamp": "2025-08-19T15:23:45.678Z",
  "event_id": "evt-12345-abcde",
  "event_type": "bureau_query",
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "auth_id": "auth-789",
  "token_id": "tok-012",
  "purpose": "credit_evaluation",
  "source_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "query_details": {
    "type": "score_check",
    "parameters": "..."
  },
  "status": "success",
  "response_time_ms": 345
}
```### 6.3. Retenção e Imutabilidade

- Logs armazenados por 5 anos (configurável por região)
- Assinatura digital para garantir imutabilidade
- Separação entre logs operacionais e de auditoria
- Consulta/exportação para fins de compliance

## 7. Gerenciamento de Vulnerabilidades

### 7.1. SAST e DAST

- Análise estática de código em cada commit (SonarQube)
- Análise dinâmica em ambiente de staging (OWASP ZAP)
- Verificação de dependências (Snyk, OWASP Dependency Check)
- Escaneamento de containers (Trivy)

### 7.2. Gestão de Patches

- Atualização automática de dependências não críticas
- Janela de manutenção mensal para atualizações críticas
- Processo de hotfix para vulnerabilidades críticas
- Monitoramento contínuo de CVEs relevantes

## 8. Resposta a Incidentes

### 8.1. Detecção

- Monitoramento de padrões anômalos de acesso
- Alertas para volumes incomuns de consultas
- Detecção de tentativas de acesso não autorizado
- Identificação de comportamentos suspeitos

### 8.2. Resposta

Em caso de detecção de incidente:

1. Revogação automática de tokens suspeitos
2. Bloqueio temporário de acessos do IP de origem
3. Notificação para equipe de segurança
4. Registro detalhado do incidente
5. Análise forense dos logs relacionados## 9. Configurações de Segurança

### 9.1. Headers HTTP Seguros

Todos os endpoints implementam os seguintes headers:

- `Content-Security-Policy`: Restrito a origens confiáveis
- `Strict-Transport-Security`: max-age=31536000; includeSubDomains
- `X-Content-Type-Options`: nosniff
- `X-Frame-Options`: DENY
- `X-XSS-Protection`: 1; mode=block
- `Cache-Control`: no-store, no-cache

### 9.2. CORS

Configuração restritiva de CORS:

```json
"extra_config": {
  "security/cors": {
    "allow_origins": ["https://admin.innovabiz.com", "https://app.innovabiz.com"],
    "allow_methods": ["POST", "OPTIONS"],
    "allow_headers": ["Authorization", "Content-Type"],
    "expose_headers": ["X-Request-ID"],
    "max_age": 3600
  }
}
```

## 10. Conformidade Regulatória

### 10.1. GDPR (Europa)

- Bases legais para processamento claramente documentadas
- Direitos do titular implementados (acesso, retificação, exclusão)
- Avaliação de impacto (DPIA) para todas as operações de alto risco
- Cláusulas contratuais para transferência de dados

### 10.2. LGPD (Brasil)

- Consentimento específico e granular
- Registro de tratamento de dados
- Suporte a relatórios de impacto (RIPD)
- Implementação de direitos do titular

### 10.3. POPIA (África do Sul)

- Propósito específico para coleta de dados
- Limitação de retenção conforme requisitos
- Notificação de violações
- Registro de atividades de processamento### 10.4. Regulações Financeiras

- Adequação às diretrizes do BNA (Angola)
- Conformidade com instruções do Banco de Moçambique
- Alinhamento com requisitos do SARB (África do Sul)
- Suporte a mecanismos anti-fraude e AML

## 11. Testes de Segurança

### 11.1. Pipeline de CI/CD

Testes automatizados incorporados ao pipeline:

- Verificação de segredos vazados (git-secrets)
- Análise estática de código (SonarQube)
- Verificação de dependências (OWASP Dependency Check)
- Testes de penetração automatizados (OWASP ZAP)

### 11.2. Testes Periódicos

- Penetration testing completo trimestral
- Revisão de configuração mensal
- Simulação de resposta a incidentes semestral
- Bug bounty program para identificação de vulnerabilidades

## 12. Implementação em GraphQL

### 12.1. Proteções Específicas para GraphQL

- Limitação de complexidade de query
- Limitação de profundidade de query
- Limitação de quantidade de nós retornados
- Validação de aliases e fragmentos

Exemplo de configuração:

```typescript
// Limitação de complexidade
const validationRules = [
  depthLimit(5),
  createComplexityLimitRule(1000, {
    onCost: cost => {
      logger.debug(`Query cost: ${cost}`);
    },
    createError: cost => {
      return new Error(`Query is too complex: ${cost}. Maximum allowed complexity is 1000.`);
    }
  })
];
```### 12.2. Controle de Campos Sensíveis

Utilização de diretivas para controle de acesso a campos sensíveis:

```graphql
type BureauCreditReport {
  score: Int! @requireScope(scope: "bureau:score:read")
  paymentHistory: [Payment!]! @requireScope(scope: "bureau:history:read")
  debtLevel: DebtLevel! @requireScope(scope: "bureau:debt:read")
  recommendations: [Recommendation!]! @requireScope(scope: "bureau:recommendations:read")
}
```

## 13. Referências

1. [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
2. [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
3. [GraphQL Security Checklists](https://www.apollographql.com/blog/graphql/security/9-ways-to-secure-your-graphql-api/)
4. [BNA Instruções sobre Proteção de Dados](https://www.bna.ao)
5. [GDPR Official Documentation](https://gdpr.eu/tag/gdpr/)