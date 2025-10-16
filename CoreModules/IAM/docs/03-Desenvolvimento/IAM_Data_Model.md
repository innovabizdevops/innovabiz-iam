# IAM Data Model

## Introduction

This document describes the data model and integration structures of the Identity and Access Management (IAM) module of the INNOVABIZ platform. The data model has been designed to support multi-tenancy requirements, regulatory compliance, and extensibility, following enterprise data modeling best practices.

## Modeling Principles

The IAM data model follows these principles:

1. **Tenant Isolation**: Complete separation of data between tenants
2. **Auditability**: Comprehensive tracking of all changes
3. **Extensibility**: Support for custom attributes and extensions
4. **Normalization**: Appropriate normalization for data integrity
5. **Performance**: Optimizations for frequent queries
6. **Compliance**: Support for diverse regulatory requirements

## Logical Data Model

### Primary Entities

#### Tenant

Represents an isolated organization or organizational unit.

| Attribute | Type | Description |
|----------|------|-----------|
| tenant_id | UUID | Unique tenant identifier |
| name | String | Tenant name |
| domain | String | Primary tenant domain |
| status | Enum | Tenant status (active, suspended, etc.) |
| plan_type | Enum | Plan/subscription type |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| attributes | JSONB | Tenant-specific dynamic attributes |

#### User

Represents an individual digital identity.

| Attribute | Type | Description |
|----------|------|-----------|
| user_id | UUID | Unique user identifier |
| tenant_id | UUID | Reference to tenant |
| username | String | Username unique within tenant |
| email | String | Primary email address |
| first_name | String | First name |
| last_name | String | Last name |
| status | Enum | User status (active, suspended, etc.) |
| password_hash | String | Password hash (stored with Argon2id) |
| password_updated_at | Timestamp | Date of last password update |
| mfa_enabled | Boolean | Indicates if MFA is enabled |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| last_login_at | Timestamp | Date of last login |
| attributes | JSONB | User-specific dynamic attributes |

#### Role

Represents a set of responsibilities and permissions.

| Attribute | Type | Description |
|----------|------|-----------|
| role_id | UUID | Unique role identifier |
| tenant_id | UUID | Reference to tenant |
| name | String | Role name |
| description | String | Role description |
| is_system_role | Boolean | Indicates if it's a system role |
| parent_role_id | UUID | Reference to a parent role (hierarchy) |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| attributes | JSONB | Role-specific dynamic attributes |

#### Permission

Represents a specific capability in the system.

| Attribute | Type | Description |
|----------|------|-----------|
| permission_id | UUID | Unique permission identifier |
| tenant_id | UUID | Reference to tenant |
| name | String | Permission name |
| description | String | Permission description |
| resource_type | String | Resource type (e.g., "PATIENT_RECORDS") |
| action | String | Allowed action (e.g., "READ", "WRITE") |
| scope | String | Permission scope |
| is_system_permission | Boolean | Indicates if it's a system permission |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |

#### Group

Represents a logical grouping of users.

| Attribute | Type | Description |
|----------|------|-----------|
| group_id | UUID | Unique group identifier |
| tenant_id | UUID | Reference to tenant |
| name | String | Group name |
| description | String | Group description |
| parent_group_id | UUID | Reference to a parent group (hierarchy) |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| attributes | JSONB | Group-specific dynamic attributes |

### Relationship Entities

#### UserRole

Associates users with roles.

| Attribute | Type | Description |
|----------|------|-----------|
| user_id | UUID | Reference to user |
| role_id | UUID | Reference to role |
| tenant_id | UUID | Reference to tenant |
| assigned_by | UUID | User who assigned the role |
| valid_from | Timestamp | Start date of validity |
| valid_to | Timestamp | End date of validity |
| created_at | Timestamp | Creation date |

#### RolePermission

Associates roles with permissions.

| Attribute | Type | Description |
|----------|------|-----------|
| role_id | UUID | Reference to role |
| permission_id | UUID | Reference to permission |
| tenant_id | UUID | Reference to tenant |
| created_at | Timestamp | Creation date |

#### UserGroup

Associates users with groups.

| Attribute | Type | Description |
|----------|------|-----------|
| user_id | UUID | Reference to user |
| group_id | UUID | Reference to group |
| tenant_id | UUID | Reference to tenant |
| created_at | Timestamp | Creation date |

### Authentication Entities

#### MFAMethod

Records MFA methods available to a user.

| Attribute | Type | Description |
|----------|------|-----------|
| method_id | UUID | Unique method identifier |
| user_id | UUID | Reference to user |
| tenant_id | UUID | Reference to tenant |
| type | Enum | Method type (totp, sms, email, etc.) |
| identifier | String | Method identifier (e.g., phone number) |
| secret | String | Encrypted secret (when applicable) |
| is_primary | Boolean | Indicates if it's the primary method |
| is_enabled | Boolean | Indicates if the method is active |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| last_used_at | Timestamp | Date of last use |

#### ARVRMethod

Records AR/VR authentication methods.

| Attribute | Type | Description |
|----------|------|-----------|
| method_id | UUID | Unique method identifier |
| user_id | UUID | Reference to user |
| tenant_id | UUID | Reference to tenant |
| type | Enum | Method type (gesture, gaze, spatial_password) |
| template_data | Binary | Template data (encrypted) |
| created_at | Timestamp | Creation date |
| updated_at | Timestamp | Last update date |
| last_used_at | Timestamp | Date of last use |

#### Session

Stores active session information.

| Attribute | Type | Description |
|----------|------|-----------|
| session_id | UUID | Unique session identifier |
| user_id | UUID | Reference to user |
| tenant_id | UUID | Reference to tenant |
| token_value | String | Token value (hashed) |
| device_info | JSONB | Device information |
| ip_address | String | IP address |
| location | JSONB | Location information |
| issued_at | Timestamp | Issuance date |
| expires_at | Timestamp | Expiration date |
| refresh_token_id | UUID | ID of associated refresh token |
| is_active | Boolean | Indicates if the session is active |

### Audit Entities

#### AuditLog

Records security events and changes.

| Attribute | Type | Description |
|----------|------|-----------|
| log_id | UUID | Unique log identifier |
| tenant_id | UUID | Reference to tenant |
| event_type | String | Event type |
| user_id | UUID | User who performed the action |
| resource_type | String | Type of affected resource |
| resource_id | String | ID of affected resource |
| action | String | Action performed |
| status | String | Action status |
| timestamp | Timestamp | Event time |
| details | JSONB | Additional event details |
| previous_state | JSONB | Previous state (when applicable) |
| new_state | JSONB | New state (when applicable) |
| ip_address | String | IP address |
| user_agent | String | Client User-Agent |

### Compliance Entities

#### ComplianceValidation

Records compliance validations performed.

| Attribute | Type | Description |
|----------|------|-----------|
| validation_id | UUID | Unique validation identifier |
| tenant_id | UUID | Reference to tenant |
| regulation | String | Evaluated regulation |
| validation_type | String | Validation type |
| resource_type | String | Type of validated resource |
| resource_id | String | ID of validated resource |
| status | String | Validation result |
| timestamp | Timestamp | Validation time |
| details | JSONB | Validation details |
| remediation_steps | JSONB | Remediation steps (when non-compliant) |
| validated_by | UUID | User or system that performed the validation |

## Physical Model

### Database Optimizations

1. **Indexes**
   - Indexes on foreign keys
   - Composite indexes for frequent queries
   - Partial indexes for specific subsets

2. **Partitioning**
   - Partitioning of audit tables by date
   - Partitioning of tenant data for large deployments

3. **Row-Level Security**
   - RLS for tenant isolation
   - Policies per table to ensure data separation

### Example RLS Policy

```sql
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

## Extensibility

### Custom Attributes

The model supports custom attributes via JSONB columns:

1. **Schema Validation**: Schema validation to ensure consistency
2. **Indexing**: GIN indexes for efficient attribute searches
3. **Path Queries**: Path queries for accessing nested attributes

Example query with custom attributes:
```sql
SELECT * FROM users
WHERE attributes->>'department' = 'Engineering'
AND (attributes->>'location')::jsonb ? 'Lisbon';
```

### Model Extensions

For sector-specific extensions, we use extension tables with foreign keys:

```sql
CREATE TABLE healthcare_user_extensions (
    user_id UUID REFERENCES users(user_id),
    medical_license_number VARCHAR(50),
    specialty VARCHAR(100),
    hospital_affiliations JSONB,
    PRIMARY KEY (user_id)
);
```

## Integration with Other Modules

### Internal Integrations

| Module | Shared Entities | Integration Type |
|--------|--------------------------|-------------------|
| ERP | User, Role, Permission | REST/GraphQL API + Event-based |
| CRM | User, Group | REST/GraphQL API + Event-based |
| Payments | User, Permission | REST/GraphQL API + Event-based |
| Marketplaces | User, Session | REST/GraphQL API + Event-based |

### External Integrations

1. **Corporate Directories**
   - LDAP/Active Directory using configurable mappings
   - SCIM 2.0 for automatic provisioning

2. **Identity Providers**
   - OpenID Connect for identity federation
   - SAML 2.0 for enterprise authentication

## Versioning and Migration

### Versioning Strategy

- Explicit version control for all database objects
- Migration scripts with upgrade and rollback
- Atomic transactions for related changes

### Migration System

Uses Flyway for schema management:

```sql
-- V1_0_0__initial_schema.sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    ...
);
```

## Performance Considerations

### Query Optimizations

1. **Strategic Caching**
   - Cache of effective permissions
   - Cache of session data
   - Cache of authorization decisions

2. **Efficient Queries**
   - Materialization of views for complex queries
   - Specific functions for frequent calculations

### Example Function for Effective Permissions

```sql
CREATE OR REPLACE FUNCTION get_effective_permissions(
    p_user_id UUID,
    p_tenant_id UUID
) RETURNS TABLE (
    permission_id UUID,
    name VARCHAR,
    resource_type VARCHAR,
    action VARCHAR
) AS $$
BEGIN
    RETURN QUERY
        SELECT DISTINCT p.permission_id, p.name, p.resource_type, p.action
        FROM permissions p
        JOIN role_permissions rp ON p.permission_id = rp.permission_id
        JOIN user_roles ur ON rp.role_id = ur.role_id
        WHERE ur.user_id = p_user_id
        AND p.tenant_id = p_tenant_id
        AND ur.tenant_id = p_tenant_id
        AND (ur.valid_to IS NULL OR ur.valid_to > NOW());
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## Data Management and Retention

### Retention Policies

1. **Audit Logs**
   - Retention based on regulatory requirements
   - Automatic archiving for long-term storage

2. **Session Data**
   - Automatic cleanup after expiration
   - Extended retention for incident investigation

### Isolation Levels

1. **Standard Isolation**
   - READ COMMITTED level for regular operations
  
2. **Elevated Isolation**
   - SERIALIZABLE for critical security operations

## Backup and Recovery Considerations

### Backup Strategy

1. **Incremental Backups**
   - Daily for transactional data
   - With integrity validation

2. **Full Backups**
   - Weekly for complete recovery
   - Stored with encryption

3. **Point-in-Time Recovery**
   - Archive logs for recovery at any point

## Conclusion

The IAM data model of the INNOVABIZ platform was designed to provide a robust, secure, and extensible foundation for identity and access operations. It meets requirements for multi-tenancy, scalability, and regulatory compliance, while maintaining flexibility for adaptation to different sectors and use cases.
