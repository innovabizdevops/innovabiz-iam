# Módulo de Federação de Identidade IAM - INNOVABIZ

## Visão Geral

O módulo de Federação de Identidade permite integração com múltiplos provedores de identidade externos (IdPs) através de protocolos padrão como SAML, OAuth2, OIDC e LDAP. Também suporta autenticação sem senha através de FIDO2/WebAuthn.

## Estrutura do Esquema

O módulo é organizado nos seguintes esquemas:

- `iam_federation`: Esquema principal para tabelas e funções de federação de identidade

## Tabelas Principais

### identity_providers
Armazena a configuração base de todos os provedores de identidade com suporte a múltiplos tipos de federação.

### federated_identities
Mantém o vínculo entre identidades externas (em IdPs) e usuários locais no sistema INNOVABIZ.

### federation_groups
Armazena grupos e papéis definidos em provedores de identidade externos.

### group_mappings
Mapeia grupos externos para papéis locais no sistema INNOVABIZ.

### fido2_configurations
Armazena configurações para autenticação FIDO2/WebAuthn para cada tenant.

### fido2_credentials
Armazena as credenciais de segurança FIDO2/WebAuthn registradas pelos usuários.

## Tipos de Provedores Suportados

### SAML 2.0
- Suporte a Shibboleth, ADFS, Okta, Auth0, AzureAD, etc.
- Importação automática de metadados
- Mapeamento flexível de atributos e afirmações

### OAuth2/OIDC
- Suporte a provedores populares como Google, Facebook, Microsoft, etc.
- Fluxos de autorização e autenticação completos
- Gerenciamento de tokens de acesso e refresh

### LDAP
- Integração com Active Directory, OpenLDAP, e outros diretórios LDAP
- Configuração de binding, busca de usuários e grupos
- Sincronização de atributos

### FIDO2/WebAuthn
- Autenticação sem senha usando chaves de segurança e biometria
- Conformidade com padrões W3C
- Suporte a diferentes plataformas (Windows Hello, Touch ID, YubiKey, etc.)

## Funcionalidades Principais

### JIT Provisioning
Criação automática de usuários quando autenticados via provedor externo pela primeira vez.

### Auto-Linking
Vinculação automática de identidades externas a contas locais existentes com base em atributos como email.

### Mapeamento de Grupos
Mapeamento automático de grupos do IdP para papéis locais, permitindo sincronização de permissões.

### Auditoria Completa
Registro detalhado de todas as operações de federação para fins de segurança e conformidade.

### Gestão Multi-Tenant
Isolamento completo das configurações de federação por tenant, permitindo que cada organização tenha suas próprias integrações.

## Considerações de Segurança

- Validação rigorosa de tokens e asserções
- Rotação segura de chaves e certificados
- Proteção contra ataques de replay
- Validação de origem e destino

## Conformidade Regulatória

O módulo de federação foi projetado para atender às exigências de:

- GDPR (UE)
- LGPD (Brasil)
- HIPAA (EUA, para dados de saúde)
- PCI DSS (para dados de pagamento)
- ISO/IEC 27001 (padrões globais de segurança)

## Scripts de Implantação

1. `01_schema_identity_federation.sql` - Define o esquema e tabelas base
2. `02_functions_saml_federation.sql` - Funções para federação SAML
3. `03_functions_oauth2_federation.sql` - Funções para federação OAuth2
4. `04_functions_oidc_federation.sql` - Funções para federação OpenID Connect
5. `05_functions_ldap_federation.sql` - Funções para federação LDAP
6. `06_functions_fido2_webauthn.sql` - Funções para autenticação FIDO2/WebAuthn
7. `07_functions_federation_admin.sql` - Funções administrativas da federação
