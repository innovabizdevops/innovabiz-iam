# IAM API Documentation

## Overview

This document describes the Identity and Access Management (IAM) API for the INNOVABIZ platform. The API provides comprehensive identity, authentication, authorization, and compliance services for internal and external consumption.

## API Design Principles

- **API-First Design**: All functionality is exposed through consistent APIs
- **RESTful and GraphQL**: Support for both paradigms based on use case
- **Security by Default**: Security is built into the core API design
- **Versioning**: Support for API evolution with backward compatibility
- **Developer Experience**: Clear, predictable, and well-documented interfaces
- **Multi-Tenancy**: Tenant context enforced across all operations

## Base URLs

- **REST API**: `https://{tenant-id}.api.innovabiz.com/iam/v1`
- **GraphQL API**: `https://{tenant-id}.api.innovabiz.com/iam/graphql`
- **OAuth/OIDC**: `https://{tenant-id}.auth.innovabiz.com`

## Authentication & Authorization

### Authentication to the API

All API calls must be authenticated using one of the following methods:

1. **OAuth 2.1 Bearer Token**:
   ```
   Authorization: Bearer {access_token}
   ```

2. **API Key** (for service-to-service):
   ```
   X-API-Key: {api_key}
   ```

3. **Mutual TLS** (for high-security integrations)

### Tenant Context

All requests must include the tenant context:

1. **URL-based**: Using the tenant subdomain or ID in URL
2. **Header-based**:
   ```
   X-Tenant-ID: {tenant-id}
   ```

## Rate Limiting

Rate limits are applied per tenant and API key:

- **Standard Tier**: 100 requests/minute
- **Enterprise Tier**: 1000 requests/minute
- **Custom Tier**: Configurable based on requirements

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1614556800
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "Invalid credentials provided",
    "details": "The provided JWT token has expired",
    "request_id": "f7a8c39e-8dfc-42f9-9738-f82c6a99a354"
  }
}
```

Common error codes:
- `AUTHENTICATION_FAILED`: Authentication issues
- `AUTHORIZATION_FAILED`: Permission issues
- `VALIDATION_ERROR`: Invalid input
- `RESOURCE_NOT_FOUND`: Requested resource doesn't exist
- `RATE_LIMIT_EXCEEDED`: API rate limit reached
- `TENANT_CONTEXT_MISSING`: No tenant context provided
- `INTERNAL_ERROR`: Server-side error

## Core API Services

### Identity Management API

Endpoints for managing users, groups, roles, and permissions.

#### User Management

| Method | Endpoint | Description |
|--------|---------|-------------|
| `GET` | `/users` | List users with pagination and filtering |
| `POST` | `/users` | Create a new user |
| `GET` | `/users/{id}` | Get a specific user |
| `PUT` | `/users/{id}` | Update a user |
| `DELETE` | `/users/{id}` | Delete a user |
| `GET` | `/users/{id}/groups` | List groups for a user |
| `GET` | `/users/{id}/roles` | List roles for a user |
| `GET` | `/users/{id}/permissions` | List effective permissions |

Example request:
```
POST /users
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "email": "john.doe@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "attributes": {
    "department": "Engineering",
    "location": "Lisbon"
  },
  "roles": ["developer"]
}
```

Example response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john.doe@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "status": "active",
  "createdAt": "2025-05-09T12:00:00Z",
  "updatedAt": "2025-05-09T12:00:00Z",
  "attributes": {
    "department": "Engineering",
    "location": "Lisbon"
  },
  "roles": ["developer"]
}
```

#### Role & Permission Management

| Method | Endpoint | Description |
|--------|---------|-------------|
| `GET` | `/roles` | List roles |
| `POST` | `/roles` | Create a role |
| `GET` | `/roles/{id}` | Get a specific role |
| `PUT` | `/roles/{id}` | Update a role |
| `DELETE` | `/roles/{id}` | Delete a role |
| `GET` | `/permissions` | List permissions |
| `POST` | `/roles/{id}/permissions` | Add permissions to a role |

### Authentication API

Endpoints for authentication and session management.

#### Authentication Operations

| Method | Endpoint | Description |
|--------|---------|-------------|
| `POST` | `/auth/login` | Authenticate a user |
| `POST` | `/auth/logout` | End a session |
| `POST` | `/auth/token` | Get a new access token |
| `GET` | `/auth/status` | Check authentication status |
| `POST` | `/auth/mfa/initiate` | Start MFA process |
| `POST` | `/auth/mfa/verify` | Verify MFA code |

Example authentication request:
```
POST /auth/login
Content-Type: application/json
X-Tenant-ID: acme

{
  "username": "john.doe@example.com",
  "password": "securePassword123!",
  "factors": ["password"]
}
```

Example response:
```json
{
  "token": {
    "access_token": "eyJhbGciOiJSUzI1...",
    "refresh_token": "eyJhbGciOiJSUzI1...",
    "token_type": "Bearer",
    "expires_in": 3600
  },
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe"
  },
  "mfa_required": true,
  "mfa_options": ["totp", "sms"]
}
```

#### AR/VR Authentication

| Method | Endpoint | Description |
|--------|---------|-------------|
| `POST` | `/auth/ar-vr/register` | Register spatial authentication |
| `POST` | `/auth/ar-vr/authenticate` | Authenticate with spatial patterns |
| `GET` | `/auth/ar-vr/methods` | List available methods |

### Authorization API

Endpoints for authorization decisions and policy management.

| Method | Endpoint | Description |
|--------|---------|-------------|
| `POST` | `/authz/check` | Check permission for a resource |
| `POST` | `/authz/bulk-check` | Check multiple permissions |
| `GET` | `/authz/policies` | List authorization policies |
| `POST` | `/authz/policies` | Create authorization policy |

Example authorization check:
```
POST /authz/check
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "principal": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "USER"
  },
  "action": "READ",
  "resource": {
    "type": "PATIENT_RECORD",
    "id": "123456"
  },
  "context": {
    "location": "hospital-ward-1",
    "emergency": false
  }
}
```

Example response:
```json
{
  "decision": "ALLOW",
  "policies_evaluated": [
    "healthcare-data-access-policy",
    "gdpr-healthcare-policy"
  ],
  "reason": "User has doctor role with direct patient relationship"
}
```

### Compliance API

Endpoints for compliance validation and reporting.

| Method | Endpoint | Description |
|--------|---------|-------------|
| `POST` | `/compliance/validate` | Validate against compliance rules |
| `GET` | `/compliance/reports` | Get compliance reports |
| `GET` | `/compliance/regulations` | List supported regulations |

Example compliance validation:
```
POST /compliance/validate
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "regulation": "HIPAA",
  "context": {
    "data_type": "PHI",
    "operation": "TRANSFER",
    "recipient": {
      "type": "HEALTHCARE_PROVIDER",
      "id": "hospital-123"
    }
  }
}
```

Example response:
```json
{
  "compliant": true,
  "rules_evaluated": [
    "hipaa-minimum-necessary",
    "hipaa-authorized-disclosure"
  ],
  "validations": [
    {
      "rule": "hipaa-minimum-necessary",
      "status": "PASS",
      "details": "Transfer limited to required data elements"
    },
    {
      "rule": "hipaa-authorized-disclosure",
      "status": "PASS",
      "details": "Recipient is an authorized healthcare provider"
    }
  ]
}
```

## OAuth 2.1 / OpenID Connect

The IAM module implements OAuth 2.1 and OpenID Connect 1.0 standards for federated authentication.

### OAuth Endpoints

| Endpoint | Description |
|----------|-------------|
| `/oauth/authorize` | Authorization endpoint |
| `/oauth/token` | Token endpoint |
| `/oauth/revoke` | Token revocation |
| `/oauth/introspect` | Token introspection |
| `/oauth/userinfo` | User information |

### OpenID Connect Discovery

OpenID Configuration is available at:
```
GET /.well-known/openid-configuration
```

## GraphQL API

In addition to the REST API, a comprehensive GraphQL API is available for complex data operations.

Example query:
```graphql
query {
  user(id: "550e8400-e29b-41d4-a716-446655440000") {
    id
    email
    firstName
    lastName
    roles {
      name
      permissions {
        name
        description
      }
    }
    groups {
      name
      members {
        totalCount
      }
    }
    authenticationMethods {
      type
      isEnabled
      lastUsed
    }
  }
}
```

## Webhooks

The IAM system can send event notifications via webhooks:

1. **Registration**: Register webhook endpoints via API
2. **Event Types**: Security events, user lifecycle events, compliance alerts
3. **Security**: Webhooks are signed with HMAC for verification

Example webhook payload:
```json
{
  "event_type": "user.login_succeeded",
  "tenant_id": "acme",
  "timestamp": "2025-05-09T12:34:56Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "authentication_method": "password",
    "mfa_used": true
  },
  "signature": "sha256=5d3997..."
}
```

## SDKs and Client Libraries

Official SDKs are available for common platforms:

- **JavaScript/TypeScript**: For web applications
- **Python**: For backend services
- **Java**: For enterprise integrations
- **Swift/Kotlin**: For mobile applications
- **C#/.NET**: For Windows environments

## API Versioning

The API uses semantic versioning:

- **Path Versioning**: `/v1/`, `/v2/` for major versions
- **Header Versioning**: `X-API-Version: 1.2` for minor versions
- **Deprecation Notices**: `X-API-Deprecated: true` with sunset dates

## Security Considerations

- All API endpoints use TLS 1.3
- JSON payloads are validated against schemas
- Sensitive operations require elevated permissions
- API keys have defined scopes for least privilege
- Rate limiting prevents abuse
- All API calls are logged for audit

## Conclusion

The IAM API provides comprehensive identity and access management capabilities for the INNOVABIZ platform. For detailed schema definitions, examples, and interactive testing, refer to the Swagger/OpenAPI documentation available at: `https://{tenant-id}.docs.innovabiz.com/iam`.
