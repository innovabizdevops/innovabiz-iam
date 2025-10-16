# 🔐 API de Registro WebAuthn/FIDO2
# INNOVABIZ IAM

```yaml
version: 1.0.0
date: 31/07/2025
status: Em desenvolvimento
classification: Confidencial - Interno
```

## 📑 Visão Geral

API para registro de credenciais WebAuthn/FIDO2 seguindo padrões W3C WebAuthn Level 3 e FIDO2 CTAP2.1. Oferece autenticação sem senha resistente a phishing com suporte para biometria e chaves de segurança.

### 🎯 Objetivos

- Registro seguro de credenciais WebAuthn
- Suporte múltiplos autenticadores por usuário
- Validação rigorosa de attestation
- Isolamento multi-tenant e multi-regional
- Compliance NIST AAL2/AAL3

## 🏛️ Conformidade

### Padrões Aplicados

```yaml
standards:
  - name: "W3C WebAuthn Level 3"
    version: "Julho 2024"
    aspects: ["PublicKeyCredential", "Attestation", "Credential management"]
  
  - name: "FIDO2 CTAP2.1"
    version: "Junho 2023"
    aspects: ["Authenticator communication", "User verification"]
    
  - name: "NIST SP 800-63B"
    version: "Rev. 4, 2024"
    aspects: ["AAL2/AAL3 compliance", "Authenticator binding"]
```

### Compliance Regulatório

| Regulação | Região | Implementação |
|-----------|--------|---------------|
| PSD2 | UE | WebAuthn como fator SCA |
| NIST SP 800-63B | EUA | AAL2/AAL3 validation |
| LGPD/GDPR | BR/EU | Armazenamento local biometria |
| PCI DSS | Global | MFA para sistemas pagamento |

## 🔌 Especificação Técnica

### Informações Gerais

```yaml
api_info:
  base_path: "/api/v1/auth/webauthn/registration"
  authentication: "Bearer JWT"
  rate_limiting: "5 tentativas/usuário/minuto"
  protocols: ["https"]
```

### Ambientes

| Ambiente | URL | Disponibilidade |
|----------|-----|----------------|
| Produção | https://iam.innovabiz.com | 99.99% |
| Homologação | https://iam-staging.innovabiz.com | 99.5% |
| Desenvolvimento | https://iam-dev.innovabiz.com | Horário comercial |

### Segurança

```yaml
security:
  authentication: "JWT Bearer Token (sessão válida)"
  authorization: "RBAC + Context"
  scopes:
    - "webauthn:register"
    - "webauthn:manage"
  protection:
    - "Rate limiting (Redis sliding window)"
    - "Origin validation (RP ID)"
    - "Attestation verification (FIDO metadata)"
```

### Headers Obrigatórios

| Header | Descrição | Exemplo |
|--------|-----------|---------|
| `Authorization` | Token JWT | `Bearer eyJhbGci...` |
| `X-Correlation-ID` | Rastreamento | `uuid-v4` |
| `X-Tenant-ID` | Identificador tenant | `acme-corp` |
| `X-Region-Code` | Código região | `BR-SP` |
| `Content-Type` | Tipo conteúdo | `application/json` |

## 🛣️ Endpoints

### 1. Gerar Opções de Registro

#### `POST /api/v1/auth/webauthn/registration/options`

Gera opções para `navigator.credentials.create()`.

**Requisição:**
```json
{
  "displayName": "João Silva",
  "authenticatorSelection": {
    "authenticatorAttachment": "platform",
    "userVerification": "required",
    "residentKey": "preferred"
  },
  "attestation": "direct"
}
```

**Resposta (200 OK):**
```json
{
  "challenge": "Y2hhbGxlbmdlLWZyb20tc2VydmVy",
  "rp": {
    "id": "innovabiz.com",
    "name": "INNOVABIZ"
  },
  "user": {
    "id": "dXNlci1pZC1oZXJl",
    "name": "joao.silva@empresa.com",
    "displayName": "João Silva"
  },
  "pubKeyCredParams": [
    {"type": "public-key", "alg": -7},
    {"type": "public-key", "alg": -257}
  ],
  "authenticatorSelection": {
    "authenticatorAttachment": "platform",
    "userVerification": "required",
    "residentKey": "preferred"
  },
  "attestation": "direct",
  "timeout": 60000
}
```

### 2. Verificar e Registrar Credencial

#### `POST /api/v1/auth/webauthn/registration/verify`

Processa resposta de `navigator.credentials.create()`.

**Requisição:**
```json
{
  "id": "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyAhIiMkJSYnKCkqKywtLi8w",
  "rawId": "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyAhIiMkJSYnKCkqKywtLi8w",
  "type": "public-key",
  "response": {
    "clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIi...",
    "attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVik..."
  }
}
```

**Resposta (201 Created):**
```json
{
  "credentialId": "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyAhIiMkJSYnKCkqKywtLi8w",
  "displayName": "João Silva",
  "createdAt": "2025-07-31T19:13:42Z",
  "attestationFormat": "packed",
  "authenticatorData": {
    "aaguid": "12345678-1234-5678-9012-123456789012",
    "credentialBackedUp": true,
    "credentialDeviceType": "multiDevice",
    "userVerified": true
  },
  "transports": ["internal", "hybrid"]
}
```

## 🚨 Tratamento de Erros

### Formato Padrão

```json
{
  "error": {
    "code": "WEBAUTHN_REGISTRATION_FAILED",
    "message": "Falha no registro da credencial",
    "target": "attestationObject",
    "details": [
      {
        "code": "INVALID_ATTESTATION",
        "message": "Attestation inválida",
        "target": "attestationObject.attStmt"
      }
    ]
  },
  "correlation_id": "uuid-v4",
  "timestamp": "2025-07-31T19:13:42Z"
}
```

### Códigos de Erro

| Status | Código | Descrição | Ação |
|--------|--------|-----------|------|
| 400 | `INVALID_REGISTRATION_REQUEST` | Dados inválidos | Verificar formato |
| 400 | `INVALID_ATTESTATION` | Attestation inválida | Tentar outro autenticador |
| 409 | `CREDENTIAL_ALREADY_EXISTS` | Credencial já existe | Usar existente |
| 422 | `ATTESTATION_VERIFICATION_FAILED` | Falha verificação | Verificar integridade |
| 429 | `REGISTRATION_RATE_LIMIT_EXCEEDED` | Limite excedido | Aguardar |

## 🧪 Testes

### Cenários de Teste

| Cenário | Descrição | Resultado Esperado |
|---------|-----------|-------------------|
| Registro Sucesso | Credencial válida com attestation | 201 Created |
| Attestation Inválida | Attestation corrompida | 422 Validation Failed |
| Credencial Duplicada | Mesmo credentialId | 409 Conflict |
| Rate Limit | Excesso tentativas | 429 Too Many Requests |

### Ambiente Sandbox

```yaml
sandbox:
  url: "https://iam-sandbox.innovabiz.com"
  credentials:
    username: "test@innovabiz.com"
    password: "TestPass123!"
  limitations:
    - "Apenas autenticadores de teste"
    - "Dados removidos a cada 24h"
```

## 📊 Observabilidade

### Métricas

| Métrica | Descrição | Relevância |
|---------|-----------|-----------|
| `webauthn_registration_attempts_total` | Total tentativas | Volume operacional |
| `webauthn_registration_success_rate` | Taxa sucesso | Qualidade serviço |
| `webauthn_attestation_verification_duration` | Tempo verificação | Performance |

### Logs

| Evento | Nível | Dados |
|--------|-------|-------|
| `WEBAUTHN_REGISTRATION_STARTED` | INFO | userId, tenantId, correlationId |
| `WEBAUTHN_REGISTRATION_SUCCESS` | INFO | credentialId, attestationFormat |
| `WEBAUTHN_REGISTRATION_FAILED` | WARN | errorCode, reason |

### Alertas

| Alerta | Condição | Ação |
|--------|----------|------|
| `HighRegistrationFailureRate` | Taxa falha > 10% | Investigar problemas |
| `AttestationVerificationSlow` | Latência > 2s | Otimizar verificação |

## 🔄 Integrações

### Dependências

| Serviço | Propósito | Criticidade |
|---------|-----------|-------------|
| IAM Core | Validação usuário | Alta |
| Vault | Armazenamento chaves | Alta |
| Redis | Rate limiting | Média |
| Kafka | Eventos auditoria | Baixa |

### Eventos Produzidos

| Evento | Tópico | Consumidores |
|--------|--------|--------------|
| `CredentialRegistered` | `iam.webauthn.events` | Audit, Analytics |
| `RegistrationFailed` | `iam.webauthn.events` | Security, Monitoring |

---

*Preparado pela Equipe de Segurança INNOVABIZ | Última Atualização: 31/07/2025*