# IAM Identity Federation Module - INNOVABIZ

## Overview

The Identity Federation Module enables integration with multiple external identity providers (IdPs) through standard protocols such as SAML, OAuth2, OIDC, and LDAP. It also supports passwordless authentication through FIDO2/WebAuthn.

## Schema Structure

The module is organized into the following schemas:

- `iam_federation`: Main schema for identity federation tables and functions

## Core Tables

### identity_providers
Stores the base configuration of all identity providers with support for multiple federation types.

### federated_identities
Maintains the link between external identities (in IdPs) and local users in the INNOVABIZ system.

### federation_groups
Stores groups and roles defined in external identity providers.

### group_mappings
Maps external groups to local roles in the INNOVABIZ system.

### fido2_configurations
Stores configurations for FIDO2/WebAuthn authentication for each tenant.

### fido2_credentials
Stores FIDO2/WebAuthn security credentials registered by users.

## Supported Provider Types

### SAML 2.0
- Support for Shibboleth, ADFS, Okta, Auth0, AzureAD, etc.
- Automatic metadata import
- Flexible attribute and assertion mapping

### OAuth2/OIDC
- Support for popular providers like Google, Facebook, Microsoft, etc.
- Complete authorization and authentication flows
- Access and refresh token management

### LDAP
- Integration with Active Directory, OpenLDAP, and other LDAP directories
- Configuration of binding, user and group search
- Attribute synchronization

### FIDO2/WebAuthn
- Passwordless authentication using security keys and biometrics
- Compliance with W3C standards
- Support for different platforms (Windows Hello, Touch ID, YubiKey, etc.)

## Core Features

### JIT Provisioning
Automatic user creation when authenticated via an external provider for the first time.

### Auto-Linking
Automatic linking of external identities to existing local accounts based on attributes such as email.

### Group Mapping
Automatic mapping of IdP groups to local roles, enabling permission synchronization.

### Comprehensive Auditing
Detailed logging of all federation operations for security and compliance purposes.

### Multi-Tenant Management
Complete isolation of federation configurations by tenant, allowing each organization to have its own integrations.

## Security Considerations

- Rigorous validation of tokens and assertions
- Secure key and certificate rotation
- Protection against replay attacks
- Origin and destination validation

## Regulatory Compliance

The federation module is designed to meet the requirements of:

- GDPR (EU)
- LGPD (Brazil)
- HIPAA (US, for healthcare data)
- PCI DSS (for payment data)
- ISO/IEC 27001 (global security standards)

## Deployment Scripts

1. `01_schema_identity_federation.sql` - Defines the base schema and tables
2. `02_functions_saml_federation.sql` - Functions for SAML federation
3. `03_functions_oauth2_federation.sql` - Functions for OAuth2 federation
4. `04_functions_oidc_federation.sql` - Functions for OpenID Connect federation
5. `05_functions_ldap_federation.sql` - Functions for LDAP federation
6. `06_functions_fido2_webauthn.sql` - Functions for FIDO2/WebAuthn authentication
7. `07_functions_federation_admin.sql` - Administrative functions for federation
