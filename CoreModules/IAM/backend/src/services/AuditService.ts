/**
 * ============================================================================
 * INNOVABIZ IAM - Audit Service
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Serviço de auditoria para operações WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { Logger } from 'winston';
import { Pool, PoolClient } from 'pg';
import { Kafka, Producer } from 'kafkajs';
import { randomBytes } from 'crypto';

import {
  WebAuthnAuditEvent,
  WebAuthnContext
} from '../types/webauthn';
import { webauthnMetrics } from '../metrics/webauthn';

/**
 * Serviço para registro e gerenciamento de eventos de auditoria WebAuthn
 */
export class AuditService {
  private readonly logger: Logger;
  private readonly db: Pool;
  private readonly kafkaProducer?: Producer;
  private readonly auditTopic: string;

  constructor(
    logger: Logger,
    db: Pool,
    kafka?: Kafka
  ) {
    this.logger = logger;
    this.db = db;
    this.auditTopic = process.env.WEBAUTHN_AUDIT_TOPIC || 'webauthn-audit-events';
    
    if (kafka) {
      this.kafkaProducer = kafka.producer({
        groupId: 'webauthn-audit-producer',
        transactionTimeout: 30000
      });
      this.initializeKafka();
    }
  }

  /**
   * Registra evento de auditoria WebAuthn
   */
  async logWebAuthnEvent(event: WebAuthnAuditEvent): Promise<void> {
    const timer = webauthnMetrics.auditEventProcessingDuration.startTimer({
      event_type: event.type
    });

    try {
      const eventId = randomBytes(16).toString('hex');
      const timestamp = new Date();

      // Preparar dados do evento
      const auditData = {
        id: eventId,
        ...event,
        timestamp: timestamp.toISOString(),
        version: '1.0.0'
      };

      // Registrar no banco de dados
      await this.storeAuditEvent(auditData);

      // Enviar para Kafka se configurado
      if (this.kafkaProducer) {
        await this.publishToKafka(auditData);
      }

      // Métricas
      webauthnMetrics.auditEvents.inc({
        tenant_id: event.tenantId,
        event_type: event.type,
        result: event.result
      });

      this.logger.debug('WebAuthn audit event logged', {
        eventId,
        type: event.type,
        userId: event.userId,
        tenantId: event.tenantId,
        result: event.result,
        correlationId: event.correlationId
      });

    } catch (error) {
      this.logger.error('Failed to log WebAuthn audit event', {
        type: event.type,
        userId: event.userId,
        tenantId: event.tenantId,
        correlationId: event.correlationId,
        error: error.message
      });

      // Não propagar erro para não afetar operação principal
    } finally {
      timer();
    }
  }

  /**
   * Busca eventos de auditoria por filtros
   */
  async getAuditEvents(filters: {
    tenantId?: string;
    userId?: string;
    credentialId?: string;
    eventType?: string;
    result?: string;
    startDate?: Date;
    endDate?: Date;
    limit?: number;
    offset?: number;
  }): Promise<any[]> {
    try {
      let query = 'SELECT * FROM webauthn_events WHERE 1=1';
      const params: any[] = [];
      let paramIndex = 1;

      // Aplicar filtros
      if (filters.tenantId) {
        query += ` AND tenant_id = $${paramIndex++}`;
        params.push(filters.tenantId);
      }

      if (filters.userId) {
        query += ` AND user_id = $${paramIndex++}`;
        params.push(filters.userId);
      }

      if (filters.credentialId) {
        query += ` AND credential_id = $${paramIndex++}`;
        params.push(filters.credentialId);
      }

      if (filters.eventType) {
        query += ` AND event_type = $${paramIndex++}`;
        params.push(filters.eventType);
      }

      if (filters.result) {
        query += ` AND event_result = $${paramIndex++}`;
        params.push(filters.result);
      }

      if (filters.startDate) {
        query += ` AND created_at >= $${paramIndex++}`;
        params.push(filters.startDate);
      }

      if (filters.endDate) {
        query += ` AND created_at <= $${paramIndex++}`;
        params.push(filters.endDate);
      }

      // Ordenação e paginação
      query += ' ORDER BY created_at DESC';

      if (filters.limit) {
        query += ` LIMIT $${paramIndex++}`;
        params.push(filters.limit);
      }

      if (filters.offset) {
        query += ` OFFSET $${paramIndex++}`;
        params.push(filters.offset);
      }

      const result = await this.db.query(query, params);
      return result.rows;

    } catch (error) {
      this.logger.error('Failed to get audit events', {
        filters,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Gera relatório de auditoria
   */
  async generateAuditReport(filters: {
    tenantId: string;
    startDate: Date;
    endDate: Date;
    includeDetails?: boolean;
  }): Promise<any> {
    try {
      const { tenantId, startDate, endDate, includeDetails = false } = filters;

      // Estatísticas gerais
      const statsQuery = `
        SELECT 
          event_type,
          event_result,
          COUNT(*) as count,
          MIN(created_at) as first_occurrence,
          MAX(created_at) as last_occurrence
        FROM webauthn_events 
        WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
        GROUP BY event_type, event_result
        ORDER BY count DESC
      `;

      const statsResult = await this.db.query(statsQuery, [tenantId, startDate, endDate]);

      // Usuários únicos
      const usersQuery = `
        SELECT COUNT(DISTINCT user_id) as unique_users
        FROM webauthn_events 
        WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
      `;

      const usersResult = await this.db.query(usersQuery, [tenantId, startDate, endDate]);

      // Credenciais únicas
      const credentialsQuery = `
        SELECT COUNT(DISTINCT credential_id) as unique_credentials
        FROM webauthn_events 
        WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
        AND credential_id IS NOT NULL
      `;

      const credentialsResult = await this.db.query(credentialsQuery, [tenantId, startDate, endDate]);

      // Eventos de alto risco
      const highRiskQuery = `
        SELECT COUNT(*) as high_risk_events
        FROM webauthn_events 
        WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
        AND (risk_score > 0.7 OR event_result = 'warning')
      `;

      const highRiskResult = await this.db.query(highRiskQuery, [tenantId, startDate, endDate]);

      // Distribuição por IP
      const ipDistributionQuery = `
        SELECT 
          ip_address,
          COUNT(*) as count,
          COUNT(DISTINCT user_id) as unique_users
        FROM webauthn_events 
        WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
        GROUP BY ip_address
        ORDER BY count DESC
        LIMIT 20
      `;

      const ipDistributionResult = await this.db.query(ipDistributionQuery, [tenantId, startDate, endDate]);

      const report = {
        period: {
          start: startDate,
          end: endDate
        },
        tenant: tenantId,
        summary: {
          uniqueUsers: parseInt(usersResult.rows[0].unique_users),
          uniqueCredentials: parseInt(credentialsResult.rows[0].unique_credentials),
          highRiskEvents: parseInt(highRiskResult.rows[0].high_risk_events),
          totalEvents: statsResult.rows.reduce((sum, row) => sum + parseInt(row.count), 0)
        },
        eventStatistics: statsResult.rows,
        ipDistribution: ipDistributionResult.rows,
        generatedAt: new Date().toISOString(),
        version: '1.0.0'
      };

      // Incluir detalhes se solicitado
      if (includeDetails) {
        const detailsQuery = `
          SELECT *
          FROM webauthn_events 
          WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
          ORDER BY created_at DESC
          LIMIT 1000
        `;

        const detailsResult = await this.db.query(detailsQuery, [tenantId, startDate, endDate]);
        report.details = detailsResult.rows;
      }

      this.logger.info('Audit report generated', {
        tenantId,
        startDate,
        endDate,
        totalEvents: report.summary.totalEvents,
        includeDetails
      });

      return report;

    } catch (error) {
      this.logger.error('Failed to generate audit report', {
        filters,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Detecta anomalias nos eventos de auditoria
   */
  async detectAnomalies(tenantId: string, lookbackHours: number = 24): Promise<any[]> {
    try {
      const anomalies: any[] = [];

      // Detectar picos de tentativas falhadas
      const failureSpikesQuery = `
        SELECT 
          DATE_TRUNC('hour', created_at) as hour,
          COUNT(*) as failure_count
        FROM webauthn_events 
        WHERE tenant_id = $1 
        AND created_at > NOW() - INTERVAL '${lookbackHours} hours'
        AND event_result = 'failure'
        GROUP BY DATE_TRUNC('hour', created_at)
        HAVING COUNT(*) > 50
        ORDER BY failure_count DESC
      `;

      const failureSpikesResult = await this.db.query(failureSpikesQuery, [tenantId]);
      
      failureSpikesResult.rows.forEach(row => {
        anomalies.push({
          type: 'failure_spike',
          description: `High failure rate detected: ${row.failure_count} failures in hour ${row.hour}`,
          severity: 'high',
          timestamp: row.hour,
          metadata: { failureCount: row.failure_count }
        });
      });

      // Detectar tentativas de múltiplos IPs para mesmo usuário
      const multipleIPsQuery = `
        SELECT 
          user_id,
          COUNT(DISTINCT ip_address) as ip_count,
          ARRAY_AGG(DISTINCT ip_address) as ip_addresses
        FROM webauthn_events 
        WHERE tenant_id = $1 
        AND created_at > NOW() - INTERVAL '${lookbackHours} hours'
        GROUP BY user_id
        HAVING COUNT(DISTINCT ip_address) > 5
        ORDER BY ip_count DESC
      `;

      const multipleIPsResult = await this.db.query(multipleIPsQuery, [tenantId]);
      
      multipleIPsResult.rows.forEach(row => {
        anomalies.push({
          type: 'multiple_ips',
          description: `User ${row.user_id} accessed from ${row.ip_count} different IPs`,
          severity: 'medium',
          timestamp: new Date(),
          metadata: { 
            userId: row.user_id, 
            ipCount: row.ip_count,
            ipAddresses: row.ip_addresses
          }
        });
      });

      // Detectar credenciais com muitas anomalias de sign count
      const signCountAnomaliesQuery = `
        SELECT 
          credential_id,
          COUNT(*) as anomaly_count
        FROM webauthn_events 
        WHERE tenant_id = $1 
        AND created_at > NOW() - INTERVAL '${lookbackHours} hours'
        AND event_type = 'sign_count_anomaly'
        GROUP BY credential_id
        HAVING COUNT(*) > 3
        ORDER BY anomaly_count DESC
      `;

      const signCountAnomaliesResult = await this.db.query(signCountAnomaliesQuery, [tenantId]);
      
      signCountAnomaliesResult.rows.forEach(row => {
        anomalies.push({
          type: 'sign_count_anomalies',
          description: `Credential ${row.credential_id} has ${row.anomaly_count} sign count anomalies`,
          severity: 'critical',
          timestamp: new Date(),
          metadata: { 
            credentialId: row.credential_id, 
            anomalyCount: row.anomaly_count
          }
        });
      });

      // Detectar padrões de acesso suspeitos (velocidade alta)
      const velocityAnomaliesQuery = `
        SELECT 
          user_id,
          ip_address,
          COUNT(*) as event_count,
          MIN(created_at) as first_event,
          MAX(created_at) as last_event,
          EXTRACT(EPOCH FROM (MAX(created_at) - MIN(created_at))) as duration_seconds
        FROM webauthn_events 
        WHERE tenant_id = $1 
        AND created_at > NOW() - INTERVAL '1 hour'
        GROUP BY user_id, ip_address
        HAVING COUNT(*) > 20 AND EXTRACT(EPOCH FROM (MAX(created_at) - MIN(created_at))) < 300
        ORDER BY event_count DESC
      `;

      const velocityAnomaliesResult = await this.db.query(velocityAnomaliesQuery, [tenantId]);
      
      velocityAnomaliesResult.rows.forEach(row => {
        anomalies.push({
          type: 'velocity_anomaly',
          description: `High velocity detected: ${row.event_count} events in ${row.duration_seconds} seconds`,
          severity: 'high',
          timestamp: row.last_event,
          metadata: { 
            userId: row.user_id,
            ipAddress: row.ip_address,
            eventCount: row.event_count,
            durationSeconds: row.duration_seconds
          }
        });
      });

      this.logger.info('Anomaly detection completed', {
        tenantId,
        lookbackHours,
        anomaliesFound: anomalies.length
      });

      return anomalies;

    } catch (error) {
      this.logger.error('Failed to detect anomalies', {
        tenantId,
        lookbackHours,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Arquiva eventos antigos
   */
  async archiveOldEvents(retentionDays: number = 365): Promise<number> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      // Mover eventos antigos para tabela de arquivo
      const archiveQuery = `
        INSERT INTO webauthn_events_archive 
        SELECT * FROM webauthn_events 
        WHERE created_at < NOW() - INTERVAL '${retentionDays} days'
      `;

      const archiveResult = await client.query(archiveQuery);
      const archivedCount = archiveResult.rowCount || 0;

      // Remover eventos arquivados da tabela principal
      const deleteQuery = `
        DELETE FROM webauthn_events 
        WHERE created_at < NOW() - INTERVAL '${retentionDays} days'
      `;

      await client.query(deleteQuery);

      await client.query('COMMIT');

      this.logger.info('Old audit events archived', {
        retentionDays,
        archivedCount
      });

      return archivedCount;

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to archive old events', {
        retentionDays,
        error: error.message
      });
      
      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Métodos privados
   */

  private async initializeKafka(): Promise<void> {
    try {
      if (this.kafkaProducer) {
        await this.kafkaProducer.connect();
        this.logger.info('Kafka producer connected for audit events');
      }
    } catch (error) {
      this.logger.error('Failed to initialize Kafka producer', {
        error: error.message
      });
    }
  }

  private async storeAuditEvent(event: any): Promise<void> {
    const query = `
      INSERT INTO webauthn_events (
        id, tenant_id, user_id, credential_id, event_type, event_result,
        error_code, error_message, client_data, authenticator_data, signature,
        user_verified, sign_count, risk_score, compliance_level,
        ip_address, user_agent, correlation_id, created_at, metadata
      ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
        $16, $17, $18, $19, $20
      )
    `;

    const values = [
      event.id,
      event.tenantId,
      event.userId,
      event.credentialId,
      event.type,
      event.result,
      event.errorCode,
      event.errorMessage,
      event.clientData,
      event.authenticatorData,
      event.signature,
      event.userVerified,
      event.signCount,
      event.riskScore,
      event.complianceLevel,
      event.ipAddress,
      event.userAgent,
      event.correlationId,
      event.timestamp,
      JSON.stringify(event.metadata || {})
    ];

    await this.db.query(query, values);
  }

  private async publishToKafka(event: any): Promise<void> {
    if (!this.kafkaProducer) return;

    try {
      await this.kafkaProducer.send({
        topic: this.auditTopic,
        messages: [{
          key: event.correlationId,
          value: JSON.stringify(event),
          headers: {
            eventType: event.type,
            tenantId: event.tenantId,
            userId: event.userId,
            timestamp: event.timestamp
          }
        }]
      });

    } catch (error) {
      this.logger.error('Failed to publish audit event to Kafka', {
        eventId: event.id,
        error: error.message
      });
      // Não propagar erro para não afetar operação principal
    }
  }

  /**
   * Cleanup ao finalizar
   */
  async disconnect(): Promise<void> {
    try {
      if (this.kafkaProducer) {
        await this.kafkaProducer.disconnect();
        this.logger.info('Kafka producer disconnected');
      }
    } catch (error) {
      this.logger.error('Error disconnecting Kafka producer', {
        error: error.message
      });
    }
  }
}