/**
 * ============================================================================
 * INNOVABIZ IAM - Credential Management Service
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Serviço para gerenciamento de credenciais WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { Pool, PoolClient } from 'pg';
import { Logger } from 'winston';
import { Redis } from 'ioredis';
import { randomBytes } from 'crypto';

import {
  WebAuthnCredential,
  CredentialStatus,
  StoreCredentialRequest,
  CredentialFilter,
  CredentialMetadata,
  CredentialUsageUpdate
} from '../types/webauthn';
import { webauthnMetrics } from '../metrics/webauthn';

/**
 * Serviço para gerenciamento completo de credenciais WebAuthn
 */
export class CredentialService {
  private readonly logger: Logger;
  private readonly db: Pool;
  private readonly redis: Redis;

  constructor(logger: Logger, db: Pool, redis: Redis) {
    this.logger = logger;
    this.db = db;
    this.redis = redis;
  }

  /**
   * Armazena uma nova credencial WebAuthn
   */
  async storeCredential(request: StoreCredentialRequest): Promise<WebAuthnCredential> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      this.logger.info('Storing WebAuthn credential', {
        userId: request.userId,
        tenantId: request.tenantId,
        credentialId: request.credentialId
      });

      // Verificar se credencial já existe
      const existingResult = await client.query(
        'SELECT id FROM webauthn_credentials WHERE credential_id = $1',
        [request.credentialId]
      );

      if (existingResult.rows.length > 0) {
        throw new Error('Credential already exists');
      }

      // Gerar ID único
      const credentialUuid = randomBytes(16).toString('hex');

      // Inserir credencial
      const insertResult = await client.query(`
        INSERT INTO webauthn_credentials (
          id, user_id, tenant_id, credential_id, public_key, sign_count,
          aaguid, attestation_format, attestation_data, user_verified,
          backup_eligible, backup_state, transports, authenticator_type,
          device_type, friendly_name, compliance_level, risk_score,
          status, created_at, updated_at, metadata
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
          $15, $16, $17, $18, $19, NOW(), NOW(), $20
        ) RETURNING *
      `, [
        credentialUuid,
        request.userId,
        request.tenantId,
        request.credentialId,
        request.publicKey,
        request.signCount,
        request.aaguid,
        request.attestationFormat,
        JSON.stringify(request.attestationData),
        request.userVerified,
        request.backupEligible,
        request.backupState,
        JSON.stringify(request.transports),
        request.authenticatorType,
        request.deviceType,
        request.friendlyName,
        request.complianceLevel,
        request.riskScore,
        'active',
        JSON.stringify(request.metadata)
      ]);

      const credential = this.mapRowToCredential(insertResult.rows[0]);

      // Registrar evento de criação
      await client.query(`
        INSERT INTO webauthn_events (
          id, tenant_id, user_id, credential_id, event_type, event_result,
          client_data, user_verified, compliance_level, ip_address,
          user_agent, correlation_id, created_at, metadata
        ) VALUES (
          gen_random_uuid(), $1, $2, $3, 'credential_created', 'success',
          $4, $5, $6, $7, $8, $9, NOW(), $10
        )
      `, [
        request.tenantId,
        request.userId,
        credentialUuid,
        JSON.stringify({ action: 'credential_stored' }),
        request.userVerified,
        request.complianceLevel,
        request.metadata?.ipAddress || 'unknown',
        request.metadata?.userAgent || 'unknown',
        request.metadata?.correlationId || 'unknown',
        JSON.stringify({
          attestationFormat: request.attestationFormat,
          deviceType: request.deviceType,
          authenticatorType: request.authenticatorType
        })
      ]);

      await client.query('COMMIT');

      // Invalidar cache
      await this.invalidateUserCredentialsCache(request.userId, request.tenantId);

      // Métricas
      webauthnMetrics.credentialsTotal.inc({
        tenant_id: request.tenantId,
        status: 'active',
        authenticator_type: request.authenticatorType
      });

      this.logger.info('WebAuthn credential stored successfully', {
        credentialId: credential.id,
        userId: request.userId,
        tenantId: request.tenantId,
        deviceType: request.deviceType
      });

      return credential;

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to store WebAuthn credential', {
        userId: request.userId,
        tenantId: request.tenantId,
        error: error.message
      });

      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Busca credenciais de um usuário
   */
  async getUserCredentials(
    userId: string,
    tenantId: string,
    filter: CredentialFilter = {}
  ): Promise<WebAuthnCredential[]> {
    const cacheKey = `webauthn:credentials:${userId}:${tenantId}:${JSON.stringify(filter)}`;
    
    try {
      // Verificar cache
      const cached = await this.redis.get(cacheKey);
      if (cached) {
        return JSON.parse(cached);
      }

      let query = `
        SELECT * FROM webauthn_credentials 
        WHERE user_id = $1 AND tenant_id = $2
      `;
      const params: any[] = [userId, tenantId];

      // Aplicar filtros
      if (filter.status) {
        query += ` AND status = $${params.length + 1}`;
        params.push(filter.status);
      }

      if (filter.authenticatorType) {
        query += ` AND authenticator_type = $${params.length + 1}`;
        params.push(filter.authenticatorType);
      }

      if (filter.complianceLevel) {
        query += ` AND compliance_level = $${params.length + 1}`;
        params.push(filter.complianceLevel);
      }

      query += ` ORDER BY created_at DESC`;

      if (filter.limit) {
        query += ` LIMIT $${params.length + 1}`;
        params.push(filter.limit);
      }

      const result = await this.db.query(query, params);
      const credentials = result.rows.map(row => this.mapRowToCredential(row));

      // Cache por 5 minutos
      await this.redis.setex(cacheKey, 300, JSON.stringify(credentials));

      return credentials;

    } catch (error) {
      this.logger.error('Failed to get user credentials', {
        userId,
        tenantId,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Busca credencial por ID
   */
  async getCredentialById(credentialId: string): Promise<WebAuthnCredential | null> {
    const cacheKey = `webauthn:credential:${credentialId}`;
    
    try {
      // Verificar cache
      const cached = await this.redis.get(cacheKey);
      if (cached) {
        return JSON.parse(cached);
      }

      const result = await this.db.query(
        'SELECT * FROM webauthn_credentials WHERE credential_id = $1',
        [credentialId]
      );

      if (result.rows.length === 0) {
        return null;
      }

      const credential = this.mapRowToCredential(result.rows[0]);

      // Cache por 10 minutos
      await this.redis.setex(cacheKey, 600, JSON.stringify(credential));

      return credential;

    } catch (error) {
      this.logger.error('Failed to get credential by ID', {
        credentialId,
        error: error.message
      });
      throw error;
    }
  }

  /**
   * Atualiza uso da credencial
   */
  async updateCredentialUsage(
    credentialId: string,
    newSignCount: number,
    ipAddress: string,
    userAgent: string,
    riskScore?: number
  ): Promise<void> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      // Atualizar credencial
      await client.query(`
        UPDATE webauthn_credentials 
        SET 
          sign_count = $2,
          last_used_at = NOW(),
          last_used_ip = $3,
          last_used_user_agent = $4,
          risk_score = COALESCE($5, risk_score),
          updated_at = NOW()
        WHERE id = $1
      `, [credentialId, newSignCount, ipAddress, userAgent, riskScore]);

      // Registrar uso
      await client.query(`
        INSERT INTO webauthn_credential_usage (
          credential_id, used_at, ip_address, user_agent, sign_count, risk_score
        ) VALUES ($1, NOW(), $2, $3, $4, $5)
      `, [credentialId, ipAddress, userAgent, newSignCount, riskScore]);

      await client.query('COMMIT');

      // Invalidar cache
      await this.invalidateCredentialCache(credentialId);

      this.logger.debug('Credential usage updated', {
        credentialId,
        newSignCount,
        riskScore
      });

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to update credential usage', {
        credentialId,
        error: error.message
      });

      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Suspende uma credencial
   */
  async suspendCredential(
    credentialId: string,
    reason: string,
    details?: string
  ): Promise<void> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      // Atualizar status
      const result = await client.query(`
        UPDATE webauthn_credentials 
        SET 
          status = 'suspended',
          suspension_reason = $2,
          suspension_details = $3,
          suspended_at = NOW(),
          updated_at = NOW()
        WHERE id = $1
        RETURNING user_id, tenant_id
      `, [credentialId, reason, details]);

      if (result.rows.length === 0) {
        throw new Error('Credential not found');
      }

      const { user_id, tenant_id } = result.rows[0];

      // Registrar evento
      await client.query(`
        INSERT INTO webauthn_events (
          id, tenant_id, user_id, credential_id, event_type, event_result,
          error_code, error_message, created_at, metadata
        ) VALUES (
          gen_random_uuid(), $1, $2, $3, 'credential_suspended', 'warning',
          $4, $5, NOW(), $6
        )
      `, [
        tenant_id,
        user_id,
        credentialId,
        reason,
        details || 'Credential suspended',
        JSON.stringify({
          suspensionReason: reason,
          suspensionDetails: details,
          suspendedAt: new Date().toISOString()
        })
      ]);

      await client.query('COMMIT');

      // Invalidar caches
      await this.invalidateCredentialCache(credentialId);
      await this.invalidateUserCredentialsCache(user_id, tenant_id);

      // Métricas
      webauthnMetrics.credentialsSuspended.inc({
        tenant_id,
        reason
      });

      this.logger.warn('Credential suspended', {
        credentialId,
        userId: user_id,
        tenantId: tenant_id,
        reason,
        details
      });

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to suspend credential', {
        credentialId,
        error: error.message
      });

      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Reativa uma credencial suspensa
   */
  async reactivateCredential(
    credentialId: string,
    reason: string
  ): Promise<void> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      const result = await client.query(`
        UPDATE webauthn_credentials 
        SET 
          status = 'active',
          suspension_reason = NULL,
          suspension_details = NULL,
          suspended_at = NULL,
          reactivated_at = NOW(),
          reactivation_reason = $2,
          updated_at = NOW()
        WHERE id = $1 AND status = 'suspended'
        RETURNING user_id, tenant_id
      `, [credentialId, reason]);

      if (result.rows.length === 0) {
        throw new Error('Credential not found or not suspended');
      }

      const { user_id, tenant_id } = result.rows[0];

      // Registrar evento
      await client.query(`
        INSERT INTO webauthn_events (
          id, tenant_id, user_id, credential_id, event_type, event_result,
          created_at, metadata
        ) VALUES (
          gen_random_uuid(), $1, $2, $3, 'credential_reactivated', 'success',
          NOW(), $4
        )
      `, [
        tenant_id,
        user_id,
        credentialId,
        JSON.stringify({
          reactivationReason: reason,
          reactivatedAt: new Date().toISOString()
        })
      ]);

      await client.query('COMMIT');

      // Invalidar caches
      await this.invalidateCredentialCache(credentialId);
      await this.invalidateUserCredentialsCache(user_id, tenant_id);

      this.logger.info('Credential reactivated', {
        credentialId,
        userId: user_id,
        tenantId: tenant_id,
        reason
      });

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to reactivate credential', {
        credentialId,
        error: error.message
      });

      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Remove uma credencial (soft delete)
   */
  async deleteCredential(
    credentialId: string,
    reason: string
  ): Promise<void> {
    const client = await this.db.connect();
    
    try {
      await client.query('BEGIN');

      const result = await client.query(`
        UPDATE webauthn_credentials 
        SET 
          status = 'deleted',
          deleted_at = NOW(),
          deletion_reason = $2,
          updated_at = NOW()
        WHERE id = $1 AND status != 'deleted'
        RETURNING user_id, tenant_id
      `, [credentialId, reason]);

      if (result.rows.length === 0) {
        throw new Error('Credential not found or already deleted');
      }

      const { user_id, tenant_id } = result.rows[0];

      // Registrar evento
      await client.query(`
        INSERT INTO webauthn_events (
          id, tenant_id, user_id, credential_id, event_type, event_result,
          created_at, metadata
        ) VALUES (
          gen_random_uuid(), $1, $2, $3, 'credential_deleted', 'success',
          NOW(), $4
        )
      `, [
        tenant_id,
        user_id,
        credentialId,
        JSON.stringify({
          deletionReason: reason,
          deletedAt: new Date().toISOString()
        })
      ]);

      await client.query('COMMIT');

      // Invalidar caches
      await this.invalidateCredentialCache(credentialId);
      await this.invalidateUserCredentialsCache(user_id, tenant_id);

      // Métricas
      webauthnMetrics.credentialsDeleted.inc({
        tenant_id,
        reason
      });

      this.logger.info('Credential deleted', {
        credentialId,
        userId: user_id,
        tenantId: tenant_id,
        reason
      });

    } catch (error) {
      await client.query('ROLLBACK');
      
      this.logger.error('Failed to delete credential', {
        credentialId,
        error: error.message
      });

      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Atualiza nome amigável da credencial
   */
  async updateCredentialName(
    credentialId: string,
    userId: string,
    tenantId: string,
    friendlyName: string
  ): Promise<void> {
    try {
      const result = await this.db.query(`
        UPDATE webauthn_credentials 
        SET 
          friendly_name = $4,
          updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND tenant_id = $3
      `, [credentialId, userId, tenantId, friendlyName]);

      if (result.rowCount === 0) {
        throw new Error('Credential not found or access denied');
      }

      // Invalidar caches
      await this.invalidateCredentialCache(credentialId);
      await this.invalidateUserCredentialsCache(userId, tenantId);

      this.logger.info('Credential name updated', {
        credentialId,
        userId,
        tenantId,
        friendlyName
      });

    } catch (error) {
      this.logger.error('Failed to update credential name', {
        credentialId,
        userId,
        tenantId,
        error: error.message
      });

      throw error;
    }
  }

  /**
   * Conta credenciais ativas de um usuário
   */
  async getUserCredentialCount(userId: string, tenantId: string): Promise<number> {
    try {
      const result = await this.db.query(`
        SELECT COUNT(*) as count 
        FROM webauthn_credentials 
        WHERE user_id = $1 AND tenant_id = $2 AND status = 'active'
      `, [userId, tenantId]);

      return parseInt(result.rows[0].count);

    } catch (error) {
      this.logger.error('Failed to get user credential count', {
        userId,
        tenantId,
        error: error.message
      });

      throw error;
    }
  }

  /**
   * Busca estatísticas de credenciais
   */
  async getCredentialStats(tenantId: string): Promise<any> {
    try {
      const result = await this.db.query(`
        SELECT 
          status,
          authenticator_type,
          compliance_level,
          COUNT(*) as count
        FROM webauthn_credentials 
        WHERE tenant_id = $1
        GROUP BY status, authenticator_type, compliance_level
        ORDER BY count DESC
      `, [tenantId]);

      return result.rows;

    } catch (error) {
      this.logger.error('Failed to get credential stats', {
        tenantId,
        error: error.message
      });

      throw error;
    }
  }

  /**
   * Métodos privados
   */

  private mapRowToCredential(row: any): WebAuthnCredential {
    return {
      id: row.id,
      userId: row.user_id,
      tenantId: row.tenant_id,
      credentialId: row.credential_id,
      publicKey: row.public_key,
      signCount: row.sign_count,
      aaguid: row.aaguid,
      attestationFormat: row.attestation_format,
      attestationData: typeof row.attestation_data === 'string' 
        ? JSON.parse(row.attestation_data) 
        : row.attestation_data,
      userVerified: row.user_verified,
      backupEligible: row.backup_eligible,
      backupState: row.backup_state,
      transports: typeof row.transports === 'string' 
        ? JSON.parse(row.transports) 
        : row.transports,
      authenticatorType: row.authenticator_type,
      deviceType: row.device_type,
      friendlyName: row.friendly_name,
      complianceLevel: row.compliance_level,
      riskScore: row.risk_score,
      status: row.status as CredentialStatus,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
      lastUsedAt: row.last_used_at,
      lastUsedIp: row.last_used_ip,
      lastUsedUserAgent: row.last_used_user_agent,
      suspendedAt: row.suspended_at,
      suspensionReason: row.suspension_reason,
      suspensionDetails: row.suspension_details,
      deletedAt: row.deleted_at,
      deletionReason: row.deletion_reason,
      reactivatedAt: row.reactivated_at,
      reactivationReason: row.reactivation_reason,
      metadata: typeof row.metadata === 'string' 
        ? JSON.parse(row.metadata) 
        : row.metadata
    };
  }

  private async invalidateCredentialCache(credentialId: string): Promise<void> {
    const keys = await this.redis.keys(`webauthn:credential:*${credentialId}*`);
    if (keys.length > 0) {
      await this.redis.del(...keys);
    }
  }

  private async invalidateUserCredentialsCache(userId: string, tenantId: string): Promise<void> {
    const keys = await this.redis.keys(`webauthn:credentials:${userId}:${tenantId}:*`);
    if (keys.length > 0) {
      await this.redis.del(...keys);
    }
  }
}