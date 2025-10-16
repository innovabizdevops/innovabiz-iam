# 🏗️ Arquitetura do Módulo Core IAM

## 📊 **VISÃO GERAL DA ARQUITETURA**

### **Princípios Arquiteturais**
- **Zero Trust Architecture**: Nunca confiar, sempre verificar
- **Defense in Depth**: Múltiplas camadas de segurança
- **Least Privilege**: Acesso mínimo necessário
- **Separation of Duties**: Segregação de responsabilidades
- **Privacy by Design**: Privacidade desde a concepção

## 🎯 **ARQUITETURA DE ALTO NÍVEL**

```mermaid
graph TB
    subgraph "Frontend Layer"
        UI[UI Components]
        SDK[IAM SDKs]
        Widget[Auth Widgets]
    end
    
    subgraph "API Gateway Layer"
        KRA[KrakenD Gateway]
        GQL[GraphQL Federation]
        MCP[Model Context Protocol]
    end
    
    subgraph "Service Layer"
        AUTH[Authentication Service]
        AUTHZ[Authorization Service]
        IDENT[Identity Service]
        TOKEN[Token Service]
        SESS[Session Service]
    end
    
    subgraph "Security Layer"
        OPA[Open Policy Agent]
        VAULT[HashiCorp Vault]
        HSM[Hardware Security Module]
    end
    
    subgraph "Data Layer"
        PG[PostgreSQL Primary]
        REDIS[Redis Cache]
        NEO[Neo4J Graph]
    end
```

## 🔧 **COMPONENTES PRINCIPAIS**

### **1. Authentication Engine**
- **Métodos Suportados**: 400+ métodos
- **Biometria**: Face, Fingerprint, Iris, Voice, Behavior
- **Passwordless**: Magic Links, WebAuthn, FIDO2
- **Traditional**: Username/Password, PIN, OTP
- **Social**: OAuth2, SAML, OpenID Connect
- **Blockchain**: DID, Verifiable Credentials