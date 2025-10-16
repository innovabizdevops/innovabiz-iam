# IAM Module Data Model - INNOVABIZ

## Overview

This document describes the data model for the IAM (Identity and Access Management) module of the INNOVABIZ platform, with emphasis on supporting the 70 authentication methods as defined in the implementation plan. The model was designed to meet the following requirements:

- **Multi-tenancy**: Support for multiple clients (tenants) in a single instance
- **Multi-regional**: Specific adaptations for target regions (EU/Portugal, Brazil, Angola, USA)
- **Regulatory compliance**: Adherence to GDPR, LGPD, PNDSB, and US regulations
- **Scalability**: Optimized structure for high volume of authentications
- **Flexibility**: Easy to add new authentication methods
- **Security**: Compliance with security best practices (NIST, ISO 27001)

## Database Structure

The IAM module has its own database that integrates with the main INNOVABIZ platform database. All tables are organized under the `iam` schema for isolation and organization.

### Main Entities

#### 1. Tenants (`iam.tenants`)

Stores information about organizations using the platform.

| Field | Type | Description |
|-------|------|-------------|
| tenant_id | UUID | Unique tenant identifier |
| tenant_code | VARCHAR(50) | Tenant code (unique) |
| name | VARCHAR(200) | Tenant name |
| description | TEXT | Tenant description |
| domain | VARCHAR(255) | Main tenant domain |
| region | VARCHAR(10) | Main region (EU, BR, AO, US) |
| settings | JSONB | Specific settings in JSON format |
| plan | VARCHAR(50) | Subscription plan |
| status | VARCHAR(20) | Tenant status (active, inactive, blocked, trial) |

#### 2. Users (`iam.users`)

Stores platform user information.

| Field | Type | Description |
|-------|------|-------------|
| user_id | UUID | Unique user identifier |
| tenant_id | UUID | Reference to tenant |
| username | VARCHAR(100) | Username (login) |
| email | VARCHAR(255) | User email |
| phone | VARCHAR(50) | User phone |
| password_hash | TEXT | Password hash (when applicable) |
| full_name | VARCHAR(200) | User full name |
| status | VARCHAR(20) | User status (active, inactive, blocked, pending, deleted) |
| verified_data | BOOLEAN | Indicates if data has been verified |
| verified_email | BOOLEAN | Indicates if email has been verified |
| verified_phone | BOOLEAN | Indicates if phone has been verified |
| mfa_required | BOOLEAN | Indicates if MFA is required |
| failed_attempts | INTEGER | Failed login attempts counter |
| profile_data | JSONB | Profile data in JSON format |

#### 3. Authentication Methods (`iam.authentication_methods`)

Catalog of available authentication methods.

| Field | Type | Description |
|-------|------|-------------|
| method_id | VARCHAR(10) | Unique method identifier (e.g., K01, P02) |
| method_code | VARCHAR(50) | Internal method code |
| name_pt | VARCHAR(100) | Name in Portuguese |
| name_en | VARCHAR(100) | Name in English |
| category | VARCHAR(50) | Category (knowledge, possession, biometric, context) |
| factor | VARCHAR(20) | Authentication factor (knowledge, possession, inherence) |
| complexity | VARCHAR(20) | Implementation complexity level |
| priority | INTEGER | Method priority (0-100) |
| implementation_wave | INTEGER | Implementation wave (1-7) |
| security_level | VARCHAR(20) | Security level offered |
| status | VARCHAR(20) | Method status (planned, development, active, disabled, deprecated) |
| regional_adaptations | JSONB | Region-specific adaptations |

#### 4. User Methods (`iam.user_methods`)

Association of authentication methods to each user.

| Field | Type | Description |
|-------|------|-------------|
| user_method_id | UUID | Unique association identifier |
| user_id | UUID | Reference to user |
| method_id | VARCHAR(10) | Reference to authentication method |
| enabled | BOOLEAN | Indicates if the method is enabled |
| verified | BOOLEAN | Indicates if the method has been verified |
| preferred | BOOLEAN | Indicates if it's the preferred method |
| auth_data | JSONB | Method-specific data for the user |
| device_name | VARCHAR(200) | Device name (when applicable) |

#### 5. Sessions (`iam.sessions`)

Stores information about active sessions.

| Field | Type | Description |
|-------|------|-------------|
| session_id | UUID | Unique session identifier |
| user_id | UUID | Reference to user |
| refresh_token | TEXT | Refresh token |
| client_id | VARCHAR(100) | Client/application identifier |
| ip_address | VARCHAR(45) | Source IP address |
| device_id | VARCHAR(255) | Device identifier |
| created_at | TIMESTAMP | Session creation date |
| expires_at | TIMESTAMP | Session expiration date |
| active | BOOLEAN | Indicates if the session is active |
| auth_factors | JSONB | List of factors used in authentication |
| auth_level | VARCHAR(20) | Authentication level (single_factor, two_factor, multi_factor) |

#### 6. Applications (`iam.applications`)

Applications registered by tenant.

| Field | Type | Description |
|-------|------|-------------|
| application_id | UUID | Unique application identifier |
| tenant_id | UUID | Reference to tenant |
| name | VARCHAR(200) | Application name |
| app_type | VARCHAR(50) | Application type (web, mobile, desktop, api) |
| client_id | VARCHAR(100) | OAuth client ID |
| client_secret | TEXT | OAuth client secret |
| redirect_uris | TEXT[] | Allowed redirect URIs |
| status | VARCHAR(20) | Application status |

#### 7. Authentication Flows (`iam.authentication_flows`)

Definition of configurable authentication flows.

| Field | Type | Description |
|-------|------|-------------|
| flow_id | UUID | Unique flow identifier |
| tenant_id | UUID | Reference to tenant |
| name | VARCHAR(100) | Flow name |
| steps | JSONB | Flow steps in JSON format |
| adaptive | BOOLEAN | Indicates if the flow is risk-based adaptive |
| security_level | VARCHAR(20) | Flow security level |
| status | VARCHAR(20) | Flow status |

#### 8. Risk Profiles (`iam.risk_profiles`)

User risk profiles for adaptive authentication.

| Field | Type | Description |
|-------|------|-------------|
| profile_id | UUID | Unique profile identifier |
| user_id | UUID | Reference to user |
| risk_score | INTEGER | Risk score (0-100) |
| risk_level | VARCHAR(20) | Risk level (low, medium, high) |
| common_locations | JSONB | Known user locations |
| common_devices | JSONB | Known user devices |
| time_patterns | JSONB | Temporal usage patterns |
| behavior_patterns | JSONB | Behavioral patterns |
| detected_anomalies | JSONB | Record of detected anomalies |

### Relationships

The data model follows a relational structure with the following connections:

1. A **tenant** can have many **users**
2. A **user** can have multiple associated **authentication methods**
3. A **user** can have many active **sessions**
4. A **tenant** can define multiple **authentication flows**
5. A **user** has one **risk profile**
6. A **tenant** can have multiple **applications**
7. An **authentication method** can be associated with multiple **tenants**

## Regional Adaptations

### European Union (Portugal)

- Maximum password retention period: 90 days
- Minimum password complexity: 12 characters
- Password history: 10 most recent passwords
- Privacy by default enabled
- GDPR-specific terms and policies

### Brazil

- Maximum password retention period: 60 days
- Minimum password complexity: 10 characters
- Password history: 5 most recent passwords
- Integration with ICP-Brasil certificates
- LGPD-specific terms and policies

### Angola

- Maximum password retention period: 90 days
- Minimum password complexity: 8 characters
- Password history: 3 most recent passwords
- Support for alternative methods in areas with limited connectivity
- PNDSB-specific terms and policies

### United States

- Maximum password retention period: 120 days
- Minimum password complexity: 8 characters
- Password history: 5 most recent passwords
- Specific settings for regulated sectors (HIPAA, SOX, GLBA)
- Compliance with NIST 800-63

## Secure Storage

- **Passwords**: Stored using Argon2id (recommended algorithm for 2025+)
- **Sensitive Data**: Individually encrypted
- **Tokens**: Using secure signature mechanisms (with key rotation)
- **Biometric Data**: Stored as secure templates, never in raw format
- **Personal Data**: Respecting data minimization principles

## Security Considerations

- Complete protection against information leakage through SQL Injection
- Prevention of timing attacks in critical operations such as credential verification
- Real-time monitoring of suspicious authentication attempts
- Complete auditing of all authentication operations
- Optimized indexes for reducing response time without compromising security

## Extensibility

The model was designed to facilitate the addition of new authentication methods through:

1. Registration of the new method in the `iam.authentication_methods` table
2. Enabling the method for specific tenants
3. Configuration of regional adaptations
4. Integration with existing authentication flows

## Integration with the INNOVABIZ Platform

The IAM module integrates with other INNOVABIZ platform modules through:

1. Secure internal APIs
2. Events published on the event bus
3. Centralized authentication and authorization via KrakenD API Gateway
4. Support for the Model Context Protocol (MCP) for inter-module communication

---

© 2025 INNOVABIZ - All rights reserved
