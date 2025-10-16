# Documentação da API do Módulo IAM INNOVABIZ

## Visão Geral

Esta documentação descreve a API do Módulo de Gerenciamento de Identidade e Acesso (IAM) da plataforma INNOVABIZ. A API foi projetada seguindo os melhores padrões de segurança e arquitetura API RESTful, com suporte a múltiplas regiões, múltiplos inquilinos e conformidade regulatória.

## Base URL

A API está disponível em:

- **Desenvolvimento**: `https://dev-iam-api.innovabiz.io/v1`
- **Qualidade**: `https://qa-iam-api.innovabiz.io/v1`
- **Homologação**: `https://staging-iam-api.innovabiz.io/v1`
- **Produção**: `https://iam-api.innovabiz.io/v1`
- **Sandbox**: `https://sandbox-iam-api.innovabiz.io/v1`

## Cabeçalhos Comuns

Todos os endpoints exigem os seguintes cabeçalhos:

| Cabeçalho | Descrição | Obrigatório |
|-----------|-----------|------------|
| `X-Tenant-Id` | Identificador único do inquilino | Sim |
| `X-Region-Code` | Código da região (EU, BR, AO, US) | Sim |
| `X-Correlation-ID` | Identificador de correlação para rastreabilidade | Não |
| `X-Device-Id` | Identificador único do dispositivo | Não |
| `X-Client-Version` | Versão do cliente que está fazendo a solicitação | Não |
| `Authorization` | Token JWT de autenticação (Bearer token) | Para endpoints autenticados |

## Autenticação e Autorização

A API utiliza autenticação baseada em JWT (JSON Web Tokens). Os tokens podem ser obtidos através do endpoint de login e devem ser incluídos em todas as solicitações subsequentes que exigem autenticação.

### Ciclo de Vida do Token

- **Token de Acesso**: Válido por curtos períodos (15-120 minutos, dependendo da região e do contexto de segurança)
- **Token de Atualização**: Usado para obter um novo token de acesso sem necessidade de nova autenticação completa
- **Revogação**: Tokens podem ser revogados explicitamente antes de sua expiração

## Endpoints de Autenticação

### Login

```
POST /auth/login
```

Inicia o processo de autenticação de um usuário.

#### Parâmetros de Consulta

| Parâmetro | Descrição |
|-----------|-----------|
| `flow` | (Opcional) ID do fluxo de autenticação desejado |
| `method` | (Opcional) Código do método de autenticação desejado |

#### Corpo da Requisição

```json
{
  "username": "usuario@exemplo.com",
  "password": "senha123",
  "method_code": "K01",
  "tenant_id": "tenant-xyz"
}
```

#### Resposta - Autenticação Simples (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### Resposta - Desafio MFA Necessário (202 Accepted)

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "challenge_type": "otp",
  "expires_in": 300,
  "delivery_method": "email",
  "delivery_destination": "m***@e***.com",
  "next_step": "mfa_verification"
}
```

#### Códigos de Erro

| Código | Descrição |
|--------|-----------|
| 400 | Parâmetros inválidos |
| 401 | Credenciais inválidas |
| 403 | Conta bloqueada ou desativada |
| 429 | Muitas tentativas |

### Desafio MFA

```
POST /auth/mfa/challenge
```

Solicita um novo desafio de autenticação multifator.

#### Parâmetros de Consulta

| Parâmetro | Descrição |
|-----------|-----------|
| `method` | (Opcional) Código do método MFA desejado |

#### Corpo da Requisição

```json
{
  "method_code": "K05"
}
```

#### Resposta (200 OK)

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "challenge_type": "otp",
  "expires_in": 300,
  "delivery_method": "email",
  "delivery_destination": "m***@e***.com"
}
```

### Verificação MFA

```
POST /auth/mfa/verify
```

Verifica um desafio de autenticação multifator.

#### Corpo da Requisição

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "code": "123456",
  "method_code": "K05"
}
```

#### Resposta (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Atualização de Token

```
POST /auth/token/refresh
```

Obtém um novo token de acesso usando um token de atualização válido.

#### Corpo da Requisição

```json
{
  "refresh_token": "eyJhbGciOiJS..."
}
```

#### Resposta (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### Revogação de Token

```
POST /auth/token/revoke
```

Revoga um token de acesso ou de atualização.

#### Corpo da Requisição

```json
{
  "token": "eyJhbGciOiJS...",
  "token_type_hint": "refresh_token"
}
```

#### Resposta (204 No Content)

### Solicitação de Redefinição de Senha

```
POST /auth/password/reset-request
```

Solicita um link de redefinição de senha.

#### Corpo da Requisição

```json
{
  "username": "usuario@exemplo.com",
  "tenant_id": "tenant-xyz"
}
```

#### Resposta (202 Accepted)

```json
{
  "message": "Se o usuário existir, um e-mail de redefinição de senha será enviado",
  "expires_in": 3600
}
```

### Redefinição de Senha

```
POST /auth/password/reset
```

Redefine a senha de um usuário usando um token de redefinição válido.

#### Corpo da Requisição

```json
{
  "token": "eyJhbGciOiJS...",
  "new_password": "NovaSenha123!"
}
```

#### Resposta (200 OK)

```json
{
  "message": "Senha redefinida com sucesso"
}
```

### Alteração de Senha

```
POST /auth/password/change
```

Altera a senha de um usuário autenticado.

#### Corpo da Requisição

```json
{
  "current_password": "SenhaAtual123",
  "new_password": "NovaSenha123!"
}
```

#### Resposta (200 OK)

```json
{
  "message": "Senha alterada com sucesso"
}
```

## Endpoints de Métodos de Autenticação

### Listar Métodos de Autenticação

```
GET /auth/methods
```

Retorna os métodos de autenticação disponíveis para o inquilino.

#### Parâmetros de Consulta

| Parâmetro | Descrição |
|-----------|-----------|
| `category` | (Opcional) Filtrar por categoria |
| `active` | (Opcional) Filtrar por status de ativação (true/false) |
| `factor` | (Opcional) Filtrar por fator de autenticação (first, second) |

#### Resposta (200 OK)

```json
{
  "methods": [
    {
      "code": "K01",
      "name": "Senha Tradicional",
      "description": "Autenticação com nome de usuário e senha",
      "category": "credencial",
      "factor": "first",
      "active": true,
      "config": {
        "password_policy": {
          "min_length": 10,
          "require_special_chars": true,
          "require_numbers": true,
          "require_uppercase": true,
          "require_lowercase": true
        }
      }
    },
    {
      "code": "K05",
      "name": "Código de Uso Único (OTP)",
      "description": "Autenticação por código temporário enviado por e-mail ou SMS",
      "category": "posse",
      "factor": "second",
      "active": true,
      "config": {
        "delivery_methods": ["email", "sms"],
        "validity_seconds": 300,
        "code_length": 6
      }
    }
  ]
}
```

### Listar Fluxos de Autenticação

```
GET /auth/flows
```

Retorna os fluxos de autenticação disponíveis para o inquilino.

#### Parâmetros de Consulta

| Parâmetro | Descrição |
|-----------|-----------|
| `security_level` | (Opcional) Filtrar por nível de segurança (low, medium, high) |
| `adaptive` | (Opcional) Filtrar por comportamento adaptativo (true/false) |

#### Resposta (200 OK)

```json
{
  "flows": [
    {
      "id": "basic",
      "name": "Autenticação Básica",
      "description": "Autenticação com nome de usuário e senha",
      "security_level": "low",
      "adaptive": false,
      "steps": [
        {
          "order": 1,
          "methods": ["K01"],
          "required": true
        }
      ]
    },
    {
      "id": "enhanced",
      "name": "Autenticação Reforçada",
      "description": "Autenticação em dois fatores com senha e OTP",
      "security_level": "high",
      "adaptive": false,
      "steps": [
        {
          "order": 1,
          "methods": ["K01"],
          "required": true
        },
        {
          "order": 2,
          "methods": ["K05"],
          "required": true
        }
      ]
    },
    {
      "id": "adaptive",
      "name": "Autenticação Adaptativa",
      "description": "Adiciona segundo fator baseado em análise de risco",
      "security_level": "medium",
      "adaptive": true,
      "steps": [
        {
          "order": 1,
          "methods": ["K01"],
          "required": true
        },
        {
          "order": 2,
          "methods": ["K05"],
          "required": "conditional",
          "condition": "risk_level > 50"
        }
      ]
    }
  ]
}
```

## Endpoints de Perfil do Usuário

### Obter Perfil do Usuário

```
GET /auth/me
```

Retorna informações sobre o usuário autenticado.

#### Resposta (200 OK)

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "usuario@exemplo.com",
  "display_name": "Nome do Usuário",
  "email": "usuario@exemplo.com",
  "email_verified": true,
  "phone": "+5511999999999",
  "phone_verified": false,
  "created_at": "2023-01-01T12:00:00Z",
  "last_login": "2023-03-01T15:30:45Z",
  "security_info": {
    "mfa_enabled": true,
    "password_last_changed": "2023-02-15T10:20:30Z",
    "risk_level": "low"
  },
  "preferences": {
    "language": "pt-BR",
    "timezone": "America/Sao_Paulo",
    "notification_channels": ["email", "sms"]
  }
}
```

### Obter Métodos de Autenticação do Usuário

```
GET /auth/me/methods
```

Retorna os métodos de autenticação configurados para o usuário.

#### Resposta (200 OK)

```json
{
  "methods": [
    {
      "id": "456e4567-e89b-12d3-a456-426614174000",
      "code": "K01",
      "name": "Senha Tradicional",
      "status": "active",
      "last_used": "2023-03-01T15:30:45Z",
      "created_at": "2023-01-01T12:00:00Z"
    },
    {
      "id": "789e4567-e89b-12d3-a456-426614174000",
      "code": "K05",
      "name": "Código de Uso Único (OTP)",
      "status": "active",
      "delivery_method": "email",
      "delivery_destination": "m***@e***.com",
      "last_used": "2023-03-01T15:30:45Z",
      "created_at": "2023-01-10T14:20:30Z"
    }
  ]
}
```

### Atualizar Método de Autenticação do Usuário

```
PUT /auth/me/methods/{method_id}
```

Atualiza as configurações de um método de autenticação específico do usuário.

#### Corpo da Requisição

```json
{
  "status": "active",
  "delivery_method": "sms",
  "delivery_destination": "+5511999999999"
}
```

#### Resposta (200 OK)

```json
{
  "id": "789e4567-e89b-12d3-a456-426614174000",
  "code": "K05",
  "name": "Código de Uso Único (OTP)",
  "status": "active",
  "delivery_method": "sms",
  "delivery_destination": "+*****99999",
  "updated_at": "2023-03-05T10:15:20Z"
}
```

### Remover Método de Autenticação do Usuário

```
DELETE /auth/me/methods/{method_id}
```

Remove um método de autenticação específico do usuário.

#### Resposta (204 No Content)

### Listar Sessões do Usuário

```
GET /auth/me/sessions
```

Retorna as sessões ativas do usuário.

#### Resposta (200 OK)

```json
{
  "sessions": [
    {
      "id": "abc4567-e89b-12d3-a456-426614174000",
      "device": "Chrome no Windows",
      "ip_address": "192.168.1.1",
      "location": "São Paulo, Brasil",
      "created_at": "2023-03-01T15:30:45Z",
      "last_activity": "2023-03-05T10:15:20Z",
      "is_current": true
    },
    {
      "id": "def4567-e89b-12d3-a456-426614174000",
      "device": "App Móvel no Android",
      "ip_address": "192.168.2.2",
      "location": "Rio de Janeiro, Brasil",
      "created_at": "2023-03-02T09:45:30Z",
      "last_activity": "2023-03-04T18:20:10Z",
      "is_current": false
    }
  ]
}
```

### Encerrar Sessão do Usuário

```
DELETE /auth/me/sessions/{session_id}
```

Encerra uma sessão específica do usuário.

#### Resposta (204 No Content)

## Respostas de Erro

Todas as respostas de erro seguem o mesmo formato:

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "As credenciais fornecidas são inválidas",
    "details": {
      "field": "password",
      "reason": "A senha fornecida não corresponde ao usuário"
    },
    "trace_id": "abc-xyz-123",
    "documentation_url": "https://docs.innovabiz.io/errors/INVALID_CREDENTIALS"
  }
}
```

## Considerações Regionais

A API do Módulo IAM INNOVABIZ está adaptada para atender aos requisitos específicos de cada região de implementação:

### União Europeia/Portugal (EU)
- Conformidade com GDPR
- Políticas de senha mais rigorosas
- Avisos de privacidade obrigatórios
- Sessões mais curtas (30 minutos)
- Verificação de geolocalização por IP

### Brasil (BR)
- Conformidade com LGPD
- Suporte a ICP-Brasil
- Configurações adaptadas aos requisitos locais
- Sessões de duração média (45 minutos)

### Angola (AO)
- Conformidade com PNDSB (Política Nacional de Dados de Angola)
- Suporte a autenticação offline
- Requisitos de senha mais flexíveis
- Sessões mais longas (60 minutos)

### Estados Unidos (US)
- Conformidade com NIST, SOC2
- Adaptações setoriais específicas (Saúde, Finanças)
- Sessões mais longas (120 minutos)
- Políticas configuráveis por tenant

## Limites de Taxa

A API implementa limites de taxa para proteger contra abusos:

| Endpoint | Limite por IP/Minuto | Limite por Tenant/Minuto |
|----------|----------------------|--------------------------|
| Login | 10-30 (depende da região) | 100-300 (depende da região) |
| Verificação MFA | 5-15 (depende da região) | 50-150 (depende da região) |
| Redefinição de Senha | 3-5 por hora | 30-50 por hora |
| Outros endpoints | 50 | 500 |

## Melhores Práticas

1. **Segurança**:
   - Armazene tokens com segurança e nunca os exponha em código-fonte ou logs
   - Implemente o logout adequado para invalidar tokens
   - Use HTTPS para todas as chamadas de API

2. **Desempenho**:
   - Implemente cache de tokens
   - Atualize tokens antes da expiração para evitar interrupções
   - Minimize o número de solicitações de autenticação

3. **Integração**:
   - Use o SDK oficial INNOVABIZ quando disponível
   - Implemente tratamento de erro apropriado
   - Siga a estratégia de retry com backoff exponencial para falhas temporárias

## Suporte e Contato

Para suporte relacionado à API IAM, entre em contato:
- Email: iam-support@innovabiz.io
- Portal do Desenvolvedor: https://developers.innovabiz.io
