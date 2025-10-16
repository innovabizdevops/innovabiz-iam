/**
 * @file schema.ts
 * @description Esquema GraphQL para integração do IAM com DataCore
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { gql } from 'apollo-server-express';

export const typeDefs = gql`
  """
  Tipos de escalares personalizados para o esquema
  """
  scalar Date
  scalar DateTime
  scalar JSON
  scalar UUID

  """
  Directivas personalizadas para controle de acesso
  """
  directive @auth(
    requires: Role = USER,
    tenant: Boolean = true
  ) on FIELD_DEFINITION

  directive @rateLimit(
    max: Int = 50,
    window: String = "10s",
    message: String = "Muitas requisições. Tente novamente mais tarde."
  ) on FIELD_DEFINITION

  directive @cacheControl(
    maxAge: Int,
    scope: CacheControlScope
  ) on FIELD_DEFINITION | OBJECT | INTERFACE | UNION

  """
  Enumerações para tipos específicos
  """
  enum Role {
    ADMIN
    USER
    SYSTEM
    AUDITOR
    RISK_ANALYST
    CUSTOMER_SUPPORT
  }

  enum CacheControlScope {
    PUBLIC
    PRIVATE
  }

  enum UserStatus {
    ACTIVE
    INACTIVE
    PENDING
    BLOCKED
    DELETED
  }

  enum AuthMethod {
    PASSWORD
    WEBAUTHN
    TOTP
    SMS
    EMAIL
    BIOMETRIC
    SOCIAL
  }

  enum PermissionType {
    CREATE
    READ
    UPDATE
    DELETE
    EXECUTE
    ADMIN
  }

  enum TenantType {
    ENTERPRISE
    SMB
    INDIVIDUAL
    GOVERNMENT
    FINANCIAL
    MARKETPLACE
  }

  enum RiskLevel {
    LOW
    MEDIUM
    HIGH
    CRITICAL
  }

  """
  Interface para entidades rastreáveis
  """
  interface Auditable {
    createdAt: DateTime!
    updatedAt: DateTime
    createdBy: String
    updatedBy: String
  }

  """
  Tipo para informações de paginação
  """
  type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
    totalCount: Int!
  }

  """
  Tipos principais do esquema GraphQL
  """
  type User implements Auditable @cacheControl(maxAge: 300, scope: PRIVATE) {
    id: UUID!
    tenantId: String!
    username: String!
    email: String!
    emailVerified: Boolean!
    firstName: String
    lastName: String
    displayName: String
    phoneNumber: String
    phoneVerified: Boolean
    status: UserStatus!
    lastLogin: DateTime
    lastFailedLogin: DateTime
    failedLoginAttempts: Int
    locale: String
    timezone: String
    picture: String
    metadata: JSON
    roles: [Role!]!
    permissions: [Permission!]!
    authMethods: [UserAuthMethod!]!
    sessions: [Session!]!
    mfaEnabled: Boolean!
    mfaMethods: [MFAMethod!]!
    createdAt: DateTime!
    updatedAt: DateTime
    createdBy: String
    updatedBy: String
    riskProfile: RiskProfile
  }

  type UserAuthMethod {
    method: AuthMethod!
    isEnabled: Boolean!
    lastUsed: DateTime
    metadata: JSON
  }

  type MFAMethod {
    type: AuthMethod!
    isDefault: Boolean!
    lastVerified: DateTime
    metadata: JSON
  }

  type Permission {
    resource: String!
    type: PermissionType!
    conditions: JSON
  }

  type Session {
    id: UUID!
    userId: UUID!
    tenantId: String!
    userAgent: String
    ipAddress: String
    deviceId: String
    isActive: Boolean!
    expiresAt: DateTime!
    createdAt: DateTime!
    lastActivity: DateTime
    metadata: JSON
  }

  type Tenant {
    id: String!
    name: String!
    displayName: String
    type: TenantType!
    isActive: Boolean!
    subscriptionStatus: String
    subscriptionExpiry: Date
    domains: [String!]
    settings: TenantSettings!
    createdAt: DateTime!
    updatedAt: DateTime
    metadata: JSON
  }

  type TenantSettings {
    passwordPolicy: PasswordPolicy
    mfaPolicy: MFAPolicy
    sessionPolicy: SessionPolicy
    allowedIPs: [String!]
    blockedIPs: [String!]
    allowedCountries: [String!]
    blockedCountries: [String!]
    customSettings: JSON
  }

  type PasswordPolicy {
    minLength: Int!
    requireLowercase: Boolean!
    requireUppercase: Boolean!
    requireNumbers: Boolean!
    requireSpecialChars: Boolean!
    passwordExpiryDays: Int
    preventPasswordReuse: Int
    allowPasswordReset: Boolean!
  }

  type MFAPolicy {
    enabled: Boolean!
    requiredMethods: Int!
    allowedMethods: [AuthMethod!]!
    rememberDeviceDays: Int
    enforceForRoles: [Role!]
    enforceForIPs: [String!]
    enforceForCountries: [String!]
    adaptiveEnabled: Boolean!
  }

  type SessionPolicy {
    idleTimeoutMinutes: Int!
    absoluteTimeoutMinutes: Int!
    refreshTokenValidityDays: Int!
    singleSession: Boolean!
    mobileSessionTimeoutMinutes: Int
    webSessionTimeoutMinutes: Int
  }

  type RiskProfile {
    userId: UUID!
    tenantId: String!
    riskLevel: RiskLevel!
    riskScore: Float!
    flags: [String!]
    lastAssessment: DateTime!
    factors: JSON
    metadata: JSON
  }

  """
  Tipos para Conexões (Paginação)
  """
  type UserConnection {
    edges: [UserEdge!]!
    pageInfo: PageInfo!
  }

  type UserEdge {
    node: User!
    cursor: String!
  }

  type TenantConnection {
    edges: [TenantEdge!]!
    pageInfo: PageInfo!
  }

  type TenantEdge {
    node: Tenant!
    cursor: String!
  }

  type SessionConnection {
    edges: [SessionEdge!]!
    pageInfo: PageInfo!
  }

  type SessionEdge {
    node: Session!
    cursor: String!
  }

  """
  Tipos para entradas de mutação
  """
  input CreateUserInput {
    tenantId: String!
    username: String!
    email: String!
    password: String
    firstName: String
    lastName: String
    phoneNumber: String
    roles: [Role!]!
    status: UserStatus
    locale: String
    timezone: String
    metadata: JSON
  }

  input UpdateUserInput {
    email: String
    firstName: String
    lastName: String
    phoneNumber: String
    status: UserStatus
    roles: [Role!]
    locale: String
    timezone: String
    metadata: JSON
  }

  input UserFilterInput {
    tenantId: String
    status: UserStatus
    email: String
    role: Role
    search: String
    createdAfter: DateTime
    createdBefore: DateTime
    lastLoginAfter: DateTime
    lastLoginBefore: DateTime
    mfaEnabled: Boolean
  }

  input TenantFilterInput {
    type: TenantType
    isActive: Boolean
    search: String
    createdAfter: DateTime
    createdBefore: DateTime
  }

  input SessionFilterInput {
    userId: UUID
    tenantId: String
    isActive: Boolean
    createdAfter: DateTime
    createdBefore: DateTime
    deviceId: String
  }

  input PaginationInput {
    first: Int
    after: String
    last: Int
    before: String
  }

  input PasswordUpdateInput {
    currentPassword: String!
    newPassword: String!
  }

  input MFAEnrollInput {
    type: AuthMethod!
    phoneNumber: String
    email: String
  }

  input RiskAssessmentInput {
    userId: UUID!
    tenantId: String!
    ipAddress: String!
    deviceId: String
    userAgent: String
    geoLocation: JSON
    timestamp: DateTime
    requestType: String!
    metadata: JSON
  }

  """
  Tipos para respostas de operações
  """
  type AuthPayload {
    user: User!
    token: String!
    refreshToken: String!
    expiresAt: DateTime!
    mfaRequired: Boolean!
    mfaOptions: [AuthMethod!]
  }

  type MFAEnrollPayload {
    success: Boolean!
    secret: String
    qrCodeUrl: String
    recoveryCode: String
    message: String
    nextStep: String
  }

  type MFAVerifyPayload {
    success: Boolean!
    token: String
    refreshToken: String
    expiresAt: DateTime
    message: String
  }

  type OperationResult {
    success: Boolean!
    message: String
    code: String
  }

  type UserMutationResult {
    success: Boolean!
    message: String
    code: String
    user: User
  }

  type TenantMutationResult {
    success: Boolean!
    message: String
    code: String
    tenant: Tenant
  }

  type RiskAssessmentResult {
    riskLevel: RiskLevel!
    riskScore: Float!
    action: String!
    requiresMFA: Boolean!
    allowLogin: Boolean!
    message: String
    additionalFactors: [String!]
  }

  """
  Queries - Operações de leitura
  """
  type Query {
    # Consultas de Usuário
    me: User! @auth
    user(id: UUID!): User! @auth(requires: ADMIN)
    users(filter: UserFilterInput, pagination: PaginationInput): UserConnection! @auth(requires: ADMIN)
    
    # Consultas de Tenant
    tenant(id: String!): Tenant! @auth(requires: ADMIN)
    tenants(filter: TenantFilterInput, pagination: PaginationInput): TenantConnection! @auth(requires: ADMIN)
    
    # Consultas de Sessão
    session(id: UUID!): Session! @auth
    sessions(filter: SessionFilterInput, pagination: PaginationInput): SessionConnection! @auth(requires: ADMIN)
    
    # Consultas de Risco e Segurança
    userRiskProfile(userId: UUID!): RiskProfile! @auth(requires: RISK_ANALYST)
    assessRisk(assessment: RiskAssessmentInput!): RiskAssessmentResult! @auth(requires: SYSTEM)
    
    # Verificações de saúde
    healthCheck: Boolean! @cacheControl(maxAge: 0)
  }

  """
  Mutations - Operações de escrita
  """
  type Mutation {
    # Autenticação
    login(username: String!, password: String!, tenantId: String!): AuthPayload!
    refreshToken(refreshToken: String!): AuthPayload!
    logout: OperationResult! @auth
    
    # Operações MFA
    enrollMFA(input: MFAEnrollInput!): MFAEnrollPayload! @auth
    verifyMFA(type: AuthMethod!, code: String!): MFAVerifyPayload! @auth
    disableMFA(type: AuthMethod!): OperationResult! @auth
    
    # Gestão de Usuários
    createUser(input: CreateUserInput!): UserMutationResult! @auth(requires: ADMIN)
    updateUser(id: UUID!, input: UpdateUserInput!): UserMutationResult! @auth(requires: ADMIN)
    deleteUser(id: UUID!): OperationResult! @auth(requires: ADMIN)
    updateMyProfile(input: UpdateUserInput!): UserMutationResult! @auth
    changePassword(input: PasswordUpdateInput!): OperationResult! @auth
    resetPassword(email: String!, tenantId: String!): OperationResult!
    confirmPasswordReset(token: String!, newPassword: String!): OperationResult!
    
    # Gestão de Sessões
    invalidateSession(id: UUID!): OperationResult! @auth
    invalidateAllSessions: OperationResult! @auth
    
    # Integração com RiskManagement
    reportRiskEvent(userId: UUID!, eventType: String!, metadata: JSON!): OperationResult! @auth(requires: SYSTEM)
  }

  """
  Subscriptions - Operações em tempo real
  """
  type Subscription {
    # Notificações de sessão
    sessionCreated(userId: UUID!): Session! @auth
    sessionInvalidated(userId: UUID!): Session! @auth
    
    # Notificações de risco
    riskLevelChanged(userId: UUID!): RiskProfile! @auth(requires: RISK_ANALYST)
  }
`;

export default typeDefs;