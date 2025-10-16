# INNOVABIZ IAM Module API Documentation

## Overview

This documentation describes the Identity and Access Management (IAM) Module API for the INNOVABIZ platform. The API has been designed following the best security standards and RESTful API architecture, with support for multiple regions, multi-tenancy, and regulatory compliance.

## Base URL

The API is available at:

- **Development**: `https://dev-iam-api.innovabiz.io/v1`
- **Quality**: `https://qa-iam-api.innovabiz.io/v1`
- **Staging**: `https://staging-iam-api.innovabiz.io/v1`
- **Production**: `https://iam-api.innovabiz.io/v1`
- **Sandbox**: `https://sandbox-iam-api.innovabiz.io/v1`

## Common Headers

All endpoints require the following headers:

| Header | Description | Required |
|--------|-------------|----------|
| `X-Tenant-Id` | Unique tenant identifier | Yes |
| `X-Region-Code` | Region code (EU, BR, AO, US) | Yes |
| `X-Correlation-ID` | Correlation identifier for traceability | No |
| `X-Device-Id` | Unique device identifier | No |
| `X-Client-Version` | Version of the client making the request | No |
| `Authorization` | JWT authentication token (Bearer token) | For authenticated endpoints |

## Authentication and Authorization

The API uses JWT (JSON Web Tokens) based authentication. Tokens can be obtained through the login endpoint and must be included in all subsequent requests that require authentication.

### Token Lifecycle

- **Access Token**: Valid for short periods (15-120 minutes, depending on region and security context)
- **Refresh Token**: Used to obtain a new access token without requiring full re-authentication
- **Revocation**: Tokens can be explicitly revoked before their expiration

## Authentication Endpoints

### Login

```
POST /auth/login
```

Initiates the authentication process for a user.

#### Query Parameters

| Parameter | Description |
|-----------|-------------|
| `flow` | (Optional) Desired authentication flow ID |
| `method` | (Optional) Desired authentication method code |

#### Request Body

```json
{
  "username": "user@example.com",
  "password": "password123",
  "method_code": "K01",
  "tenant_id": "tenant-xyz"
}
```

#### Response - Simple Authentication (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### Response - MFA Challenge Required (202 Accepted)

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "challenge_type": "otp",
  "expires_in": 300,
  "delivery_method": "email",
  "delivery_destination": "u***@e***.com",
  "next_step": "mfa_verification"
}
```

#### Error Codes

| Code | Description |
|------|-------------|
| 400 | Invalid parameters |
| 401 | Invalid credentials |
| 403 | Account locked or disabled |
| 429 | Too many attempts |

### MFA Challenge

```
POST /auth/mfa/challenge
```

Requests a new multi-factor authentication challenge.

#### Query Parameters

| Parameter | Description |
|-----------|-------------|
| `method` | (Optional) Desired MFA method code |

#### Request Body

```json
{
  "method_code": "K05"
}
```

#### Response (200 OK)

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "challenge_type": "otp",
  "expires_in": 300,
  "delivery_method": "email",
  "delivery_destination": "u***@e***.com"
}
```

### MFA Verification

```
POST /auth/mfa/verify
```

Verifies a multi-factor authentication challenge.

#### Request Body

```json
{
  "challenge_id": "789e4567-e89b-12d3-a456-426614174999",
  "code": "123456",
  "method_code": "K05"
}
```

#### Response (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Token Refresh

```
POST /auth/token/refresh
```

Obtains a new access token using a valid refresh token.

#### Request Body

```json
{
  "refresh_token": "eyJhbGciOiJS..."
}
```

#### Response (200 OK)

```json
{
  "access_token": "eyJhbGciOiJS...",
  "refresh_token": "eyJhbGciOiJS...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### Token Revocation

```
POST /auth/token/revoke
```

Revokes an access or refresh token.

#### Request Body

```json
{
  "token": "eyJhbGciOiJS...",
  "token_type_hint": "refresh_token"
}
```

#### Response (204 No Content)

### Password Reset Request

```
POST /auth/password/reset-request
```

Requests a password reset link.

#### Request Body

```json
{
  "username": "user@example.com",
  "tenant_id": "tenant-xyz"
}
```

#### Response (202 Accepted)

```json
{
  "message": "If the user exists, a password reset email will be sent",
  "expires_in": 3600
}
```

### Password Reset

```
POST /auth/password/reset
```

Resets a user's password using a valid reset token.

#### Request Body

```json
{
  "token": "eyJhbGciOiJS...",
  "new_password": "NewPassword123!"
}
```

#### Response (200 OK)

```json
{
  "message": "Password reset successfully"
}
```

### Password Change

```
POST /auth/password/change
```

Changes the password for an authenticated user.

#### Request Body

```json
{
  "current_password": "CurrentPassword123",
  "new_password": "NewPassword123!"
}
```

#### Response (200 OK)

```json
{
  "message": "Password changed successfully"
}
```

## Authentication Methods Endpoints

### List Authentication Methods

```
GET /auth/methods
```

Returns the authentication methods available for the tenant.

#### Query Parameters

| Parameter | Description |
|-----------|-------------|
| `category` | (Optional) Filter by category |
| `active` | (Optional) Filter by activation status (true/false) |
| `factor` | (Optional) Filter by authentication factor (first, second) |

#### Response (200 OK)

```json
{
  "methods": [
    {
      "code": "K01",
      "name": "Traditional Password",
      "description": "Authentication with username and password",
      "category": "credential",
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
      "name": "One-Time Password (OTP)",
      "description": "Authentication by temporary code sent by email or SMS",
      "category": "possession",
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

### List Authentication Flows

```
GET /auth/flows
```

Returns the authentication flows available for the tenant.

#### Query Parameters

| Parameter | Description |
|-----------|-------------|
| `security_level` | (Optional) Filter by security level (low, medium, high) |
| `adaptive` | (Optional) Filter by adaptive behavior (true/false) |

#### Response (200 OK)

```json
{
  "flows": [
    {
      "id": "basic",
      "name": "Basic Authentication",
      "description": "Authentication with username and password",
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
      "name": "Enhanced Authentication",
      "description": "Two-factor authentication with password and OTP",
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
      "name": "Adaptive Authentication",
      "description": "Adds second factor based on risk analysis",
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

## User Profile Endpoints

### Get User Profile

```
GET /auth/me
```

Returns information about the authenticated user.

#### Response (200 OK)

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user@example.com",
  "display_name": "User Name",
  "email": "user@example.com",
  "email_verified": true,
  "phone": "+15551234567",
  "phone_verified": false,
  "created_at": "2023-01-01T12:00:00Z",
  "last_login": "2023-03-01T15:30:45Z",
  "security_info": {
    "mfa_enabled": true,
    "password_last_changed": "2023-02-15T10:20:30Z",
    "risk_level": "low"
  },
  "preferences": {
    "language": "en-US",
    "timezone": "America/New_York",
    "notification_channels": ["email", "sms"]
  }
}
```

### Get User Authentication Methods

```
GET /auth/me/methods
```

Returns the authentication methods configured for the user.

#### Response (200 OK)

```json
{
  "methods": [
    {
      "id": "456e4567-e89b-12d3-a456-426614174000",
      "code": "K01",
      "name": "Traditional Password",
      "status": "active",
      "last_used": "2023-03-01T15:30:45Z",
      "created_at": "2023-01-01T12:00:00Z"
    },
    {
      "id": "789e4567-e89b-12d3-a456-426614174000",
      "code": "K05",
      "name": "One-Time Password (OTP)",
      "status": "active",
      "delivery_method": "email",
      "delivery_destination": "u***@e***.com",
      "last_used": "2023-03-01T15:30:45Z",
      "created_at": "2023-01-10T14:20:30Z"
    }
  ]
}
```

### Update User Authentication Method

```
PUT /auth/me/methods/{method_id}
```

Updates the settings for a specific user authentication method.

#### Request Body

```json
{
  "status": "active",
  "delivery_method": "sms",
  "delivery_destination": "+15551234567"
}
```

#### Response (200 OK)

```json
{
  "id": "789e4567-e89b-12d3-a456-426614174000",
  "code": "K05",
  "name": "One-Time Password (OTP)",
  "status": "active",
  "delivery_method": "sms",
  "delivery_destination": "+*****34567",
  "updated_at": "2023-03-05T10:15:20Z"
}
```

### Remove User Authentication Method

```
DELETE /auth/me/methods/{method_id}
```

Removes a specific authentication method from the user.

#### Response (204 No Content)

### List User Sessions

```
GET /auth/me/sessions
```

Returns the active sessions for the user.

#### Response (200 OK)

```json
{
  "sessions": [
    {
      "id": "abc4567-e89b-12d3-a456-426614174000",
      "device": "Chrome on Windows",
      "ip_address": "192.168.1.1",
      "location": "New York, USA",
      "created_at": "2023-03-01T15:30:45Z",
      "last_activity": "2023-03-05T10:15:20Z",
      "is_current": true
    },
    {
      "id": "def4567-e89b-12d3-a456-426614174000",
      "device": "Mobile App on iOS",
      "ip_address": "192.168.2.2",
      "location": "Boston, USA",
      "created_at": "2023-03-02T09:45:30Z",
      "last_activity": "2023-03-04T18:20:10Z",
      "is_current": false
    }
  ]
}
```

### Terminate User Session

```
DELETE /auth/me/sessions/{session_id}
```

Terminates a specific user session.

#### Response (204 No Content)

## Error Responses

All error responses follow the same format:

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "The provided credentials are invalid",
    "details": {
      "field": "password",
      "reason": "The provided password does not match the user"
    },
    "trace_id": "abc-xyz-123",
    "documentation_url": "https://docs.innovabiz.io/errors/INVALID_CREDENTIALS"
  }
}
```

## Regional Considerations

The INNOVABIZ IAM Module API is adapted to meet the specific requirements of each implementation region:

### European Union/Portugal (EU)
- GDPR compliance
- Stricter password policies
- Mandatory privacy notices
- Shorter sessions (30 minutes)
- IP geolocation verification

### Brazil (BR)
- LGPD compliance
- ICP-Brasil support
- Configurations adapted to local requirements
- Medium duration sessions (45 minutes)

### Angola (AO)
- PNDSB (Angola National Data Policy) compliance
- Offline authentication support
- More flexible password requirements
- Longer sessions (60 minutes)

### United States (US)
- NIST, SOC2 compliance
- Sector-specific adaptations (Healthcare, Finance)
- Longer sessions (120 minutes)
- Tenant-configurable policies

## Rate Limits

The API implements rate limits to protect against abuse:

| Endpoint | IP/Minute Limit | Tenant/Minute Limit |
|----------|-----------------|---------------------|
| Login | 10-30 (region dependent) | 100-300 (region dependent) |
| MFA Verification | 5-15 (region dependent) | 50-150 (region dependent) |
| Password Reset | 3-5 per hour | 30-50 per hour |
| Other endpoints | 50 | 500 |

## Best Practices

1. **Security**:
   - Store tokens securely and never expose them in source code or logs
   - Implement proper logout to invalidate tokens
   - Use HTTPS for all API calls

2. **Performance**:
   - Implement token caching
   - Refresh tokens before expiration to avoid disruptions
   - Minimize the number of authentication requests

3. **Integration**:
   - Use the official INNOVABIZ SDK when available
   - Implement appropriate error handling
   - Follow exponential backoff retry strategy for temporary failures

## Support and Contact

For IAM API related support, please contact:
- Email: iam-support@innovabiz.io
- Developer Portal: https://developers.innovabiz.io
