/**
 * 🔗 INTERFACES - INNOVABIZ IAM
 * Definições de interfaces TypeScript
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: TypeScript Strict Mode, Domain-Driven Design
 */

// ========================================
// CORE INTERFACES
// ========================================

export interface UserContext {
  ipAddress: string;
  userAgent: string;
  deviceFingerprint?: string;
  timestamp?: Date;
  sessionId?: string;
  tenantId?: string;
}

export interface AuthenticationResult {
  success: boolean;
  user?: any;
  tokens?: any;
  riskAssessment?: any;
  requiresStepUp?: boolean;
  metadata?: Record<string, any>;
}

export interface RegistrationResult {
  verified: boolean;
  credential?: any;
  riskAssessment?: any;
  metadata?: Record<string, any>;
}

// ========================================
// METRICS INTERFACES
// ========================================

export interface UserMetrics {
  totalLogins: number;
  successfulLogins: number;
  failedLogins: number;
  lastLogin: Date;
  averageSessionDuration: number;
  deviceCount: number;
  riskScore: number;
}

export interface SystemMetrics {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  activeSessions: number;
  authenticationRate: number;
  averageResponseTime: number;
  errorRate: number;
}

// ========================================
// CONFIGURATION INTERFACES
// ========================================

export interface WebAuthnConfig {
  rpName: string;
  rpID: string;
  origin: string;
  timeout: number;
  attestation: 'none' | 'indirect' | 'direct';
  authenticatorSelection: {
    authenticatorAttachment?: 'platform' | 'cross-platform';
    userVerification?: 'required' | 'preferred' | 'discouraged';
    residentKey?: 'required' | 'preferred' | 'discouraged';
  };
}

export interface JWTConfig {
  secret: string;
  expiresIn: string;
  refreshExpiresIn: string;
  issuer: string;
  audience: string;
}

export interface RiskConfig {
  thresholds: {
    low: number;
    medium: number;
    high: number;
  };
  weights: {
    device: number;
    location: number;
    behavior: number;
    temporal: number;
  };
  ml: {
    enabled: boolean;
    modelPath: string;
  };
}

// ========================================
// SERVICE INTERFACES
// ========================================

export interface IAMServiceInterface {
  createUser(createUserDto: any): Promise<any>;
  getUserById(userId: string, tenantId: string): Promise<any>;
  beginRegistration(registrationDto: any, context: UserContext): Promise<any>;
  completeRegistration(userId: string, response: any, context: UserContext): Promise<RegistrationResult>;
  beginAuthentication(authDto: any, context: UserContext): Promise<any>;
  completeAuthentication(response: any, context: UserContext): Promise<AuthenticationResult>;
  generateTokens(user: any, context: UserContext): Promise<any>;
  refreshTokens(refreshDto: any, context: UserContext): Promise<any>;
  validateSession(sessionToken: string): Promise<any>;
  revokeSession(sessionId: string): Promise<void>;
}

export interface WebAuthnServiceInterface {
  generateRegistrationOptions(userId: string, tenantId: string, options: any): Promise<any>;
  generateAuthenticationOptions(userId: string, tenantId: string, options: any): Promise<any>;
  verifyRegistrationResponse(response: any, challenge: string, options: any): Promise<any>;
  verifyAuthenticationResponse(response: any, challenge: string, userId: string, tenantId: string, options: any): Promise<any>;
}

export interface RiskAssessmentServiceInterface {
  assessRegistrationRisk(request: any): Promise<any>;
  assessAuthenticationRisk(request: any): Promise<any>;
  updateRiskProfile(userId: string, tenantId: string, data: any): Promise<any>;
  getRiskProfile(userId: string, tenantId: string): Promise<any>;
}

export interface CredentialServiceInterface {
  createCredential(data: any): Promise<any>;
  getUserCredentials(userId: string, tenantId: string): Promise<any[]>;
  getCredentialById(credentialId: string, tenantId: string): Promise<any>;
  updateCredential(credentialId: string, data: any): Promise<void>;
  deleteCredential(credentialId: string): Promise<void>;
}

export interface AuditServiceInterface {
  logEvent(event: any): Promise<void>;
  logSecurityEvent(event: any): Promise<void>;
  logComplianceEvent(event: any): Promise<void>;
  logRiskEvent(event: any): Promise<void>;
}

// ========================================
// REPOSITORY INTERFACES
// ========================================

export interface UserRepositoryInterface {
  create(userData: any): any;
  save(user: any): Promise<any>;
  findOne(options: any): Promise<any>;
  findOneBy(criteria: any): Promise<any>;
  update(id: string, data: any): Promise<void>;
  delete(id: string): Promise<void>;
}

export interface SessionRepositoryInterface {
  create(sessionData: any): any;
  save(session: any): Promise<any>;
  findOne(options: any): Promise<any>;
  update(id: string, data: any): Promise<void>;
  delete(id: string): Promise<void>;
}

export interface CredentialRepositoryInterface {
  create(credentialData: any): any;
  save(credential: any): Promise<any>;
  findOne(options: any): Promise<any>;
  find(options: any): Promise<any[]>;
  update(id: string, data: any): Promise<void>;
  delete(id: string): Promise<void>;
}

// ========================================
// EVENT INTERFACES
// ========================================

export interface DomainEvent {
  id: string;
  aggregateId: string;
  eventType: string;
  eventData: any;
  timestamp: Date;
  version: number;
}

export interface UserCreatedEvent extends DomainEvent {
  eventType: 'UserCreated';
  eventData: {
    userId: string;
    tenantId: string;
    email: string;
    username: string;
  };
}

export interface UserAuthenticatedEvent extends DomainEvent {
  eventType: 'UserAuthenticated';
  eventData: {
    userId: string;
    tenantId: string;
    credentialId: string;
    ipAddress: string;
    userAgent: string;
    riskScore: number;
  };
}

export interface CredentialRegisteredEvent extends DomainEvent {
  eventType: 'CredentialRegistered';
  eventData: {
    userId: string;
    tenantId: string;
    credentialId: string;
    deviceType: string;
    transports: string[];
  };
}

// ========================================
// ERROR INTERFACES
// ========================================

export interface ErrorDetails {
  code: string;
  message: string;
  details?: any;
  timestamp: Date;
  path?: string;
  method?: string;
}

export interface ValidationError {
  field: string;
  value: any;
  constraints: Record<string, string>;
}

// ========================================
// CACHE INTERFACES
// ========================================

export interface CacheInterface {
  get<T>(key: string): Promise<T | null>;
  set<T>(key: string, value: T, ttl?: number): Promise<void>;
  del(key: string): Promise<void>;
  reset(): Promise<void>;
}

// ========================================
// MONITORING INTERFACES
// ========================================

export interface MetricsCollector {
  incrementCounter(name: string, labels?: Record<string, string>): void;
  recordHistogram(name: string, value: number, labels?: Record<string, string>): void;
  setGauge(name: string, value: number, labels?: Record<string, string>): void;
}

export interface HealthCheck {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  details?: any;
  timestamp: Date;
}

export interface SystemHealth {
  status: 'healthy' | 'unhealthy' | 'degraded';
  checks: HealthCheck[];
  uptime: number;
  version: string;
}

// ========================================
// COMPLIANCE INTERFACES
// ========================================

export interface ComplianceEvent {
  eventId: string;
  userId: string;
  tenantId: string;
  eventType: string;
  regulation: string; // GDPR, PCI-DSS, etc.
  action: string;
  resource: string;
  timestamp: Date;
  metadata: Record<string, any>;
}

export interface DataRetentionPolicy {
  dataType: string;
  retentionPeriod: number; // in days
  deletionMethod: 'soft' | 'hard';
  archiveBeforeDeletion: boolean;
}

export interface ConsentRecord {
  userId: string;
  tenantId: string;
  consentType: string;
  granted: boolean;
  timestamp: Date;
  ipAddress: string;
  userAgent: string;
  version: string;
}

// ========================================
// MULTI-TENANT INTERFACES
// ========================================

export interface TenantContext {
  tenantId: string;
  tenantName: string;
  region: string;
  dataResidency: string;
  complianceRequirements: string[];
  features: string[];
}

export interface TenantConfiguration {
  tenantId: string;
  webauthnConfig: Partial<WebAuthnConfig>;
  riskConfig: Partial<RiskConfig>;
  jwtConfig: Partial<JWTConfig>;
  customSettings: Record<string, any>;
}

// ========================================
// INTEGRATION INTERFACES
// ========================================

export interface ExternalAuthProvider {
  providerId: string;
  name: string;
  type: 'oauth2' | 'saml' | 'oidc';
  configuration: Record<string, any>;
  isActive: boolean;
}

export interface APIGatewayConfig {
  endpoint: string;
  apiKey: string;
  rateLimits: {
    requests: number;
    window: number; // in seconds
  };
  timeout: number;
}

// ========================================
// MACHINE LEARNING INTERFACES
// ========================================

export interface MLModelInterface {
  predict(features: Record<string, number>): Promise<{
    prediction: number;
    confidence: number;
    featureImportance?: Record<string, number>;
  }>;
  
  train(data: any[]): Promise<void>;
  
  evaluate(testData: any[]): Promise<{
    accuracy: number;
    precision: number;
    recall: number;
    f1Score: number;
  }>;
}

export interface RiskFeatures {
  deviceRisk: number;
  locationRisk: number;
  behavioralRisk: number;
  temporalRisk: number;
  historicalRisk: number;
  velocityRisk: number;
}

// ========================================
// WORKFLOW INTERFACES
// ========================================

export interface WorkflowStep {
  stepId: string;
  name: string;
  type: 'validation' | 'transformation' | 'decision' | 'action';
  configuration: Record<string, any>;
  nextSteps: string[];
}

export interface WorkflowDefinition {
  workflowId: string;
  name: string;
  version: string;
  steps: WorkflowStep[];
  triggers: string[];
  isActive: boolean;
}

export interface WorkflowExecution {
  executionId: string;
  workflowId: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  currentStep: string;
  context: Record<string, any>;
  startTime: Date;
  endTime?: Date;
  error?: string;
}