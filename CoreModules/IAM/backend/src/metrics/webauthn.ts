/**
 * ============================================================================
 * INNOVABIZ IAM - WebAuthn Metrics
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Métricas Prometheus para WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { 
  Counter, 
  Histogram, 
  Gauge, 
  Summary,
  register 
} from 'prom-client';

import { webauthnConfig } from '../config/webauthn';

const prefix = webauthnConfig.observability.metrics.prefix;

/**
 * Métricas de registro de credenciais
 */
export const registrationAttempts = new Counter({
  name: `${prefix}registration_attempts_total`,
  help: 'Total number of WebAuthn registration attempts',
  labelNames: ['tenant_id', 'result', 'attestation_format'],
  registers: [register]
});

export const registrationDuration = new Histogram({
  name: `${prefix}registration_duration_seconds`,
  help: 'Duration of WebAuthn registration operations',
  labelNames: ['tenant_id', 'operation_type'],
  buckets: [0.1, 0.5, 1, 2, 5, 10, 30],
  registers: [register]
});

/**
 * Métricas de autenticação
 */
export const authenticationAttempts = new Counter({
  name: `${prefix}authentication_attempts_total`,
  help: 'Total number of WebAuthn authentication attempts',
  labelNames: ['tenant_id', 'result', 'aal_level'],
  registers: [register]
});

export const authenticationDuration = new Histogram({
  name: `${prefix}authentication_duration_seconds`,
  help: 'Duration of WebAuthn authentication operations',
  labelNames: ['tenant_id', 'operation_type'],
  buckets: [0.1, 0.5, 1, 2, 5, 10, 30],
  registers: [register]
});

/**
 * Métricas de verificação
 */
export const verificationDuration = new Histogram({
  name: `${prefix}verification_duration_seconds`,
  help: 'Duration of WebAuthn verification operations',
  labelNames: ['operation_type'],
  buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [register]
});

export const verificationErrors = new Counter({
  name: `${prefix}verification_errors_total`,
  help: 'Total number of WebAuthn verification errors',
  labelNames: ['tenant_id', 'error_type', 'operation_type'],
  registers: [register]
});

/**
 * Métricas de credenciais
 */
export const credentialsTotal = new Gauge({
  name: `${prefix}credentials_total`,
  help: 'Total number of WebAuthn credentials',
  labelNames: ['tenant_id', 'status', 'authenticator_type'],
  registers: [register]
});

export const credentialsSuspended = new Counter({
  name: `${prefix}credentials_suspended_total`,
  help: 'Total number of suspended WebAuthn credentials',
  labelNames: ['tenant_id', 'reason'],
  registers: [register]
});

export const credentialsDeleted = new Counter({
  name: `${prefix}credentials_deleted_total`,
  help: 'Total number of deleted WebAuthn credentials',
  labelNames: ['tenant_id', 'reason'],
  registers: [register]
});

/**
 * Métricas de segurança
 */
export const signCountAnomalies = new Counter({
  name: `${prefix}sign_count_anomalies_total`,
  help: 'Total number of sign count anomalies detected',
  labelNames: ['tenant_id'],
  registers: [register]
});

export const attestationVerificationFailures = new Counter({
  name: `${prefix}attestation_verification_failures_total`,
  help: 'Total number of attestation verification failures',
  labelNames: ['tenant_id', 'attestation_format'],
  registers: [register]
});

export const riskScoreDistribution = new Histogram({
  name: `${prefix}risk_score_distribution`,
  help: 'Distribution of risk scores for WebAuthn operations',
  labelNames: ['tenant_id', 'operation_type'],
  buckets: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
  registers: [register]
});

export const highRiskEvents = new Counter({
  name: `${prefix}high_risk_events_total`,
  help: 'Total number of high risk WebAuthn events',
  labelNames: ['tenant_id', 'risk_level', 'event_type'],
  registers: [register]
});

/**
 * Métricas de rate limiting
 */
export const rateLimitHits = new Counter({
  name: `${prefix}rate_limit_hits_total`,
  help: 'Total number of rate limit hits',
  labelNames: ['tenant_id', 'limit_type', 'identifier'],
  registers: [register]
});

export const rateLimitResets = new Counter({
  name: `${prefix}rate_limit_resets_total`,
  help: 'Total number of rate limit resets',
  labelNames: ['tenant_id', 'limit_type'],
  registers: [register]
});

/**
 * Métricas de cache
 */
export const cacheHits = new Counter({
  name: `${prefix}cache_hits_total`,
  help: 'Total number of cache hits',
  labelNames: ['cache_type'],
  registers: [register]
});

export const cacheMisses = new Counter({
  name: `${prefix}cache_misses_total`,
  help: 'Total number of cache misses',
  labelNames: ['cache_type'],
  registers: [register]
});

export const cacheOperationDuration = new Histogram({
  name: `${prefix}cache_operation_duration_seconds`,
  help: 'Duration of cache operations',
  labelNames: ['cache_type', 'operation'],
  buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1],
  registers: [register]
});

/**
 * Métricas de banco de dados
 */
export const databaseOperationDuration = new Histogram({
  name: `${prefix}database_operation_duration_seconds`,
  help: 'Duration of database operations',
  labelNames: ['operation_type', 'table'],
  buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5],
  registers: [register]
});

export const databaseConnectionsActive = new Gauge({
  name: `${prefix}database_connections_active`,
  help: 'Number of active database connections',
  registers: [register]
});

export const databaseConnectionsIdle = new Gauge({
  name: `${prefix}database_connections_idle`,
  help: 'Number of idle database connections',
  registers: [register]
});

/**
 * Métricas de API
 */
export const httpRequestsTotal = new Counter({
  name: `${prefix}http_requests_total`,
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'endpoint', 'status_code', 'tenant_id'],
  registers: [register]
});

export const httpRequestDuration = new Histogram({
  name: `${prefix}http_request_duration_seconds`,
  help: 'Duration of HTTP requests',
  labelNames: ['method', 'endpoint', 'status_code'],
  buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [register]
});

export const httpRequestSize = new Summary({
  name: `${prefix}http_request_size_bytes`,
  help: 'Size of HTTP requests',
  labelNames: ['method', 'endpoint'],
  registers: [register]
});

export const httpResponseSize = new Summary({
  name: `${prefix}http_response_size_bytes`,
  help: 'Size of HTTP responses',
  labelNames: ['method', 'endpoint', 'status_code'],
  registers: [register]
});

/**
 * Métricas de conformidade
 */
export const complianceChecks = new Counter({
  name: `${prefix}compliance_checks_total`,
  help: 'Total number of compliance checks performed',
  labelNames: ['tenant_id', 'check_type', 'result'],
  registers: [register]
});

export const complianceViolations = new Counter({
  name: `${prefix}compliance_violations_total`,
  help: 'Total number of compliance violations detected',
  labelNames: ['tenant_id', 'violation_type', 'severity'],
  registers: [register]
});

export const aalDistribution = new Gauge({
  name: `${prefix}aal_distribution`,
  help: 'Distribution of Authentication Assurance Levels',
  labelNames: ['tenant_id', 'aal_level'],
  registers: [register]
});

/**
 * Métricas de dispositivos
 */
export const deviceTypeDistribution = new Gauge({
  name: `${prefix}device_type_distribution`,
  help: 'Distribution of device types',
  labelNames: ['tenant_id', 'device_type', 'authenticator_type'],
  registers: [register]
});

export const authenticatorUsage = new Counter({
  name: `${prefix}authenticator_usage_total`,
  help: 'Total usage of different authenticator types',
  labelNames: ['tenant_id', 'aaguid', 'device_type'],
  registers: [register]
});

/**
 * Métricas de eventos de auditoria
 */
export const auditEvents = new Counter({
  name: `${prefix}audit_events_total`,
  help: 'Total number of audit events',
  labelNames: ['tenant_id', 'event_type', 'result'],
  registers: [register]
});

export const auditEventProcessingDuration = new Histogram({
  name: `${prefix}audit_event_processing_duration_seconds`,
  help: 'Duration of audit event processing',
  labelNames: ['event_type'],
  buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1],
  registers: [register]
});

/**
 * Métricas de sistema
 */
export const systemHealth = new Gauge({
  name: `${prefix}system_health`,
  help: 'System health status (1 = healthy, 0 = unhealthy)',
  labelNames: ['component'],
  registers: [register]
});

export const memoryUsage = new Gauge({
  name: `${prefix}memory_usage_bytes`,
  help: 'Memory usage in bytes',
  labelNames: ['type'],
  registers: [register]
});

export const cpuUsage = new Gauge({
  name: `${prefix}cpu_usage_percent`,
  help: 'CPU usage percentage',
  registers: [register]
});

/**
 * Métricas de configuração regional
 */
export const regionalConfigUsage = new Counter({
  name: `${prefix}regional_config_usage_total`,
  help: 'Usage of regional configurations',
  labelNames: ['region_code', 'config_type'],
  registers: [register]
});

/**
 * Métricas de challenges
 */
export const challengesGenerated = new Counter({
  name: `${prefix}challenges_generated_total`,
  help: 'Total number of challenges generated',
  labelNames: ['tenant_id', 'challenge_type'],
  registers: [register]
});

export const challengesExpired = new Counter({
  name: `${prefix}challenges_expired_total`,
  help: 'Total number of challenges that expired',
  labelNames: ['tenant_id', 'challenge_type'],
  registers: [register]
});

export const challengeLifetime = new Histogram({
  name: `${prefix}challenge_lifetime_seconds`,
  help: 'Lifetime of challenges from generation to use',
  labelNames: ['tenant_id', 'challenge_type'],
  buckets: [1, 5, 10, 30, 60, 120, 300, 600],
  registers: [register]
});

/**
 * Função para coletar métricas customizadas
 */
export function collectCustomMetrics(): void {
  // Coletar métricas de memória
  const memUsage = process.memoryUsage();
  memoryUsage.set({ type: 'rss' }, memUsage.rss);
  memoryUsage.set({ type: 'heapTotal' }, memUsage.heapTotal);
  memoryUsage.set({ type: 'heapUsed' }, memUsage.heapUsed);
  memoryUsage.set({ type: 'external' }, memUsage.external);

  // Coletar métricas de CPU (simplificado)
  const cpuUsageValue = process.cpuUsage();
  const cpuPercent = (cpuUsageValue.user + cpuUsageValue.system) / 1000000; // Convert to seconds
  cpuUsage.set(cpuPercent);
}

/**
 * Função para resetar métricas (útil para testes)
 */
export function resetMetrics(): void {
  register.clear();
}

/**
 * Função para obter todas as métricas
 */
export function getMetrics(): string {
  return register.metrics();
}

/**
 * Função para registrar métricas de saúde do sistema
 */
export function updateSystemHealth(component: string, healthy: boolean): void {
  systemHealth.set({ component }, healthy ? 1 : 0);
}

/**
 * Função para registrar uso de configuração regional
 */
export function recordRegionalConfigUsage(regionCode: string, configType: string): void {
  regionalConfigUsage.inc({ region_code: regionCode, config_type: configType });
}

/**
 * Função para registrar distribuição de AAL
 */
export function updateAALDistribution(tenantId: string, aalLevel: string, count: number): void {
  aalDistribution.set({ tenant_id: tenantId, aal_level: aalLevel }, count);
}

/**
 * Função para registrar distribuição de tipos de dispositivo
 */
export function updateDeviceTypeDistribution(
  tenantId: string, 
  deviceType: string, 
  authenticatorType: string, 
  count: number
): void {
  deviceTypeDistribution.set(
    { tenant_id: tenantId, device_type: deviceType, authenticator_type: authenticatorType }, 
    count
  );
}

// Inicializar coleta de métricas customizadas
if (webauthnConfig.observability.metrics.enabled) {
  setInterval(collectCustomMetrics, 30000); // A cada 30 segundos
}

// Exportar todas as métricas como um objeto para facilitar o uso
export const webauthnMetrics = {
  // Registro
  registrationAttempts,
  registrationDuration,
  
  // Autenticação
  authenticationAttempts,
  authenticationDuration,
  
  // Verificação
  verificationDuration,
  verificationErrors,
  
  // Credenciais
  credentialsTotal,
  credentialsSuspended,
  credentialsDeleted,
  
  // Segurança
  signCountAnomalies,
  attestationVerificationFailures,
  riskScoreDistribution,
  highRiskEvents,
  
  // Rate limiting
  rateLimitHits,
  rateLimitResets,
  
  // Cache
  cacheHits,
  cacheMisses,
  cacheOperationDuration,
  
  // Banco de dados
  databaseOperationDuration,
  databaseConnectionsActive,
  databaseConnectionsIdle,
  
  // API
  httpRequestsTotal,
  httpRequestDuration,
  httpRequestSize,
  httpResponseSize,
  
  // Conformidade
  complianceChecks,
  complianceViolations,
  aalDistribution,
  
  // Dispositivos
  deviceTypeDistribution,
  authenticatorUsage,
  
  // Auditoria
  auditEvents,
  auditEventProcessingDuration,
  
  // Sistema
  systemHealth,
  memoryUsage,
  cpuUsage,
  
  // Regional
  regionalConfigUsage,
  
  // Challenges
  challengesGenerated,
  challengesExpired,
  challengeLifetime,
  
  // Funções utilitárias
  collectCustomMetrics,
  resetMetrics,
  getMetrics,
  updateSystemHealth,
  recordRegionalConfigUsage,
  updateAALDistribution,
  updateDeviceTypeDistribution
};