// ================================================================================================
// INNOVABIZ IAM - CORE SERVICE IMPLEMENTATION
// ================================================================================================
// Módulo: IAM - Core Service
// Autor: Eduardo Jeremias - INNOVABIZ DevOps Team
// Data: 2025-08-03
// Versão: 2.0.0
// Descrição: Implementação completa do serviço IAM integrado com TransactionManagement
// ================================================================================================

import { Pool } from 'pg';
import { Redis } from 'ioredis';
import jwt from 'jsonwebtoken';
import bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import winston from 'winston';

// Interfaces para tipos de dados
export interface User {
  userId: string;
  tenantId: string;
  email: string;
  username?: string;
  firstName: string;
  lastName: string;
  roles: string[];
  permissions: string[];
  isActive: boolean;
  lastLoginAt?: Date;
  createdAt: Date;
  updatedAt: Date;
  metadata?: Record<string, any>;
}

export interface Tenant {
  tenantId: string;
  name: string;
  domain: string;
  isActive: boolean;
  subscriptionPlan: string;
  limits: {
    maxUsers: number;
    maxTransactions: number;
    maxAmount: number;
    allowedCurrencies: string[];
  };
  settings: Record<string, any>;
  createdAt: Date;
  updatedAt: Date;
}

export interface AuthToken {
  token: string;
  userId: string;
  tenantId: string;
  scopes: string[];
  expiresAt: Date;
  tokenType: 'access' | 'refresh';
}

export interface Permission {
  permissionId: string;
  name: string;
  resource: string;
  action: string;
  conditions?: Record<string, any>;
}

export interface Role {
  roleId: string;
  tenantId: string;
  name: string;
  description: string;
  permissions: Permission[];
  isSystem: boolean;
}

export interface AuditEvent {
  eventId: string;
  tenantId: string;
  userId?: string;
  eventType: string;
  resource: string;
  action: string;
  details: Record<string, any>;
  ipAddress?: string;
  userAgent?: string;
  timestamp: Date;
}

export interface LoginAttempt {
  userId: string;
  tenantId: string;
  success: boolean;
  ipAddress: string;
  userAgent: string;
  timestamp: Date;
  failureReason?: string;
}

// Comandos e queries
export interface CreateUserCommand {
  tenantId: string;
  email: string;
  username?: string;
  firstName: string;
  lastName: string;
  password: string;
  roles: string[];
  metadata?: Record<string, any>;
}

export interface AuthenticateCommand {
  tenantId: string;
  email: string;
  password: string;
  ipAddress?: string;
  userAgent?: string;
}

export interface ValidateTokenCommand {
  token: string;
  requiredScopes?: string[];
}

export interface CheckPermissionCommand {
  userId: string;
  tenantId: string;
  resource: string;
  action: string;
  context?: Record<string, any>;
}

// Resultados
export interface AuthenticationResult {
  success: boolean;
  user?: User;
  accessToken?: string;
  refreshToken?: string;
  expiresIn?: number;
  error?: string;
}

export interface ValidationResult {
  valid: boolean;
  user?: User;
  tenant?: Tenant;
  scopes?: string[];
  error?: string;
}

export interface PermissionResult {
  allowed: boolean;
  reason?: string;
  conditions?: Record<string, any>;
}

/**
 * Serviço principal do IAM
 * Implementa autenticação, autorização e auditoria
 */
export class IAMService {
  private readonly pool: Pool;
  private readonly redis: Redis;
  private readonly logger: winston.Logger;
  private readonly jwtSecret: string;
  private readonly jwtExpiresIn: string;
  private readonly refreshExpiresIn: string;

  constructor(
    pool: Pool,
    redis: Redis,
    logger: winston.Logger,
    config: {
      jwtSecret: string;
      jwtExpiresIn: string;
      refreshExpiresIn: string;
    }
  ) {
    this.pool = pool;
    this.redis = redis;
    this.logger = logger;
    this.jwtSecret = config.jwtSecret;
    this.jwtExpiresIn = config.jwtExpiresIn;
    this.refreshExpiresIn = config.refreshExpiresIn;
  }

  /**
   * Criar novo usuário
   */
  async createUser(command: CreateUserCommand): Promise<User> {
    const client = await this.pool.connect();
    
    try {
      await client.query('BEGIN');

      // Verificar se tenant existe e está ativo
      const tenantResult = await client.query(
        'SELECT * FROM iam.tenants WHERE tenant_id = $1 AND is_active = true',
        [command.tenantId]
      );

      if (tenantResult.rows.length === 0) {
        throw new Error('Tenant não encontrado ou inativo');
      }

      // Verificar se email já existe no tenant
      const existingUser = await client.query(
        'SELECT user_id FROM iam.users WHERE tenant_id = $1 AND email = $2',
        [command.tenantId, command.email]
      );

      if (existingUser.rows.length > 0) {
        throw new Error('Email já está em uso neste tenant');
      }

      // Hash da senha
      const passwordHash = await bcrypt.hash(command.password, 12);

      // Criar usuário
      const userId = uuidv4();
      const userResult = await client.query(`
        INSERT INTO iam.users (
          user_id, tenant_id, email, username, first_name, last_name, 
          password_hash, is_active, created_at, updated_at, metadata
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW(), NOW(), $8)
        RETURNING *
      `, [
        userId, command.tenantId, command.email, command.username,
        command.firstName, command.lastName, passwordHash,
        JSON.stringify(command.metadata || {})
      ]);

      // Associar roles
      for (const roleName of command.roles) {
        await client.query(`
          INSERT INTO iam.user_roles (user_id, tenant_id, role_name)
          SELECT $1, $2, $3
          WHERE EXISTS (
            SELECT 1 FROM iam.roles 
            WHERE tenant_id = $2 AND name = $3 AND is_active = true
          )
        `, [userId, command.tenantId, roleName]);
      }

      await client.query('COMMIT');

      // Buscar usuário completo com roles e permissions
      const user = await this.getUserById(userId, command.tenantId);
      
      // Log de auditoria
      await this.logAuditEvent({
        eventId: uuidv4(),
        tenantId: command.tenantId,
        userId,
        eventType: 'USER_CREATED',
        resource: 'user',
        action: 'create',
        details: {
          email: command.email,
          roles: command.roles
        },
        timestamp: new Date()
      });

      this.logger.info('Usuário criado com sucesso', {
        userId,
        tenantId: command.tenantId,
        email: command.email
      });

      return user!;

    } catch (error) {
      await client.query('ROLLBACK');
      this.logger.error('Erro ao criar usuário', {
        error: error.message,
        tenantId: command.tenantId,
        email: command.email
      });
      throw error;
    } finally {
      client.release();
    }
  }

  /**
   * Autenticar usuário
   */
  async authenticate(command: AuthenticateCommand): Promise<AuthenticationResult> {
    try {
      // Buscar usuário
      const userResult = await this.pool.query(`
        SELECT u.*, t.name as tenant_name, t.is_active as tenant_active
        FROM iam.users u
        JOIN iam.tenants t ON u.tenant_id = t.tenant_id
        WHERE u.tenant_id = $1 AND u.email = $2 AND u.is_active = true
      `, [command.tenantId, command.email]);

      if (userResult.rows.length === 0) {
        await this.logLoginAttempt({
          userId: '',
          tenantId: command.tenantId,
          success: false,
          ipAddress: command.ipAddress || '',
          userAgent: command.userAgent || '',
          timestamp: new Date(),
          failureReason: 'USER_NOT_FOUND'
        });

        return {
          success: false,
          error: 'Credenciais inválidas'
        };
      }

      const userRow = userResult.rows[0];

      // Verificar se tenant está ativo
      if (!userRow.tenant_active) {
        return {
          success: false,
          error: 'Tenant inativo'
        };
      }

      // Verificar senha
      const passwordValid = await bcrypt.compare(command.password, userRow.password_hash);
      
      if (!passwordValid) {
        await this.logLoginAttempt({
          userId: userRow.user_id,
          tenantId: command.tenantId,
          success: false,
          ipAddress: command.ipAddress || '',
          userAgent: command.userAgent || '',
          timestamp: new Date(),
          failureReason: 'INVALID_PASSWORD'
        });

        return {
          success: false,
          error: 'Credenciais inválidas'
        };
      }

      // Buscar usuário completo
      const user = await this.getUserById(userRow.user_id, command.tenantId);
      
      if (!user) {
        return {
          success: false,
          error: 'Erro interno'
        };
      }

      // Gerar tokens
      const accessToken = this.generateAccessToken(user);
      const refreshToken = this.generateRefreshToken(user);

      // Salvar refresh token no Redis
      await this.redis.setex(
        `refresh_token:${user.userId}:${command.tenantId}`,
        parseInt(this.refreshExpiresIn) * 24 * 60 * 60, // 7 dias em segundos
        refreshToken
      );

      // Atualizar último login
      await this.pool.query(
        'UPDATE iam.users SET last_login_at = NOW() WHERE user_id = $1',
        [user.userId]
      );

      // Log de auditoria
      await this.logAuditEvent({
        eventId: uuidv4(),
        tenantId: command.tenantId,
        userId: user.userId,
        eventType: 'USER_LOGIN',
        resource: 'authentication',
        action: 'login',
        details: {
          email: command.email,
          ipAddress: command.ipAddress,
          userAgent: command.userAgent
        },
        timestamp: new Date()
      });

      // Log login bem-sucedido
      await this.logLoginAttempt({
        userId: user.userId,
        tenantId: command.tenantId,
        success: true,
        ipAddress: command.ipAddress || '',
        userAgent: command.userAgent || '',
        timestamp: new Date()
      });

      this.logger.info('Autenticação bem-sucedida', {
        userId: user.userId,
        tenantId: command.tenantId,
        email: command.email
      });

      return {
        success: true,
        user,
        accessToken,
        refreshToken,
        expiresIn: parseInt(this.jwtExpiresIn) * 60 * 60 // em segundos
      };

    } catch (error) {
      this.logger.error('Erro na autenticação', {
        error: error.message,
        tenantId: command.tenantId,
        email: command.email
      });

      return {
        success: false,
        error: 'Erro interno do servidor'
      };
    }
  }

  /**
   * Validar token JWT
   */
  async validateToken(command: ValidateTokenCommand): Promise<ValidationResult> {
    try {
      // Verificar se token está na blacklist
      const blacklisted = await this.redis.get(`blacklist:${command.token}`);
      if (blacklisted) {
        return {
          valid: false,
          error: 'Token revogado'
        };
      }

      // Decodificar e verificar token
      const decoded = jwt.verify(command.token, this.jwtSecret) as any;

      // Buscar usuário
      const user = await this.getUserById(decoded.userId, decoded.tenantId);
      if (!user || !user.isActive) {
        return {
          valid: false,
          error: 'Usuário inativo'
        };
      }

      // Buscar tenant
      const tenant = await this.getTenantById(decoded.tenantId);
      if (!tenant || !tenant.isActive) {
        return {
          valid: false,
          error: 'Tenant inativo'
        };
      }

      // Verificar scopes se necessário
      if (command.requiredScopes && command.requiredScopes.length > 0) {
        const hasRequiredScopes = command.requiredScopes.every(scope => 
          decoded.scopes && decoded.scopes.includes(scope)
        );

        if (!hasRequiredScopes) {
          return {
            valid: false,
            error: 'Scopes insuficientes'
          };
        }
      }

      return {
        valid: true,
        user,
        tenant,
        scopes: decoded.scopes || []
      };

    } catch (error) {
      if (error.name === 'TokenExpiredError') {
        return {
          valid: false,
          error: 'Token expirado'
        };
      }

      if (error.name === 'JsonWebTokenError') {
        return {
          valid: false,
          error: 'Token inválido'
        };
      }

      this.logger.error('Erro na validação do token', {
        error: error.message
      });

      return {
        valid: false,
        error: 'Erro interno'
      };
    }
  }

  /**
   * Verificar permissão
   */
  async checkPermission(command: CheckPermissionCommand): Promise<PermissionResult> {
    try {
      // Buscar permissões do usuário
      const permissionsResult = await this.pool.query(`
        SELECT DISTINCT p.name, p.resource, p.action, p.conditions
        FROM iam.permissions p
        JOIN iam.role_permissions rp ON p.permission_id = rp.permission_id
        JOIN iam.user_roles ur ON rp.role_name = ur.role_name AND rp.tenant_id = ur.tenant_id
        WHERE ur.user_id = $1 AND ur.tenant_id = $2
        AND p.resource = $3 AND p.action = $4
      `, [command.userId, command.tenantId, command.resource, command.action]);

      if (permissionsResult.rows.length === 0) {
        // Verificar permissões diretas do usuário
        const directPermissions = await this.pool.query(`
          SELECT p.name, p.resource, p.action, p.conditions
          FROM iam.permissions p
          JOIN iam.user_permissions up ON p.permission_id = up.permission_id
          WHERE up.user_id = $1 AND up.tenant_id = $2
          AND p.resource = $3 AND p.action = $4
        `, [command.userId, command.tenantId, command.resource, command.action]);

        if (directPermissions.rows.length === 0) {
          return {
            allowed: false,
            reason: 'Permissão não encontrada'
          };
        }
      }

      // Log de auditoria para verificação de permissão
      await this.logAuditEvent({
        eventId: uuidv4(),
        tenantId: command.tenantId,
        userId: command.userId,
        eventType: 'PERMISSION_CHECK',
        resource: command.resource,
        action: command.action,
        details: {
          context: command.context,
          result: 'ALLOWED'
        },
        timestamp: new Date()
      });

      return {
        allowed: true
      };

    } catch (error) {
      this.logger.error('Erro na verificação de permissão', {
        error: error.message,
        userId: command.userId,
        tenantId: command.tenantId,
        resource: command.resource,
        action: command.action
      });

      return {
        allowed: false,
        reason: 'Erro interno'
      };
    }
  }

  /**
   * Buscar usuário por ID
   */
  async getUserById(userId: string, tenantId: string): Promise<User | null> {
    try {
      const userResult = await this.pool.query(`
        SELECT u.*, 
               COALESCE(
                 json_agg(
                   DISTINCT ur.role_name
                 ) FILTER (WHERE ur.role_name IS NOT NULL), 
                 '[]'
               ) as roles,
               COALESCE(
                 json_agg(
                   DISTINCT p.name
                 ) FILTER (WHERE p.name IS NOT NULL), 
                 '[]'
               ) as permissions
        FROM iam.users u
        LEFT JOIN iam.user_roles ur ON u.user_id = ur.user_id AND u.tenant_id = ur.tenant_id
        LEFT JOIN iam.role_permissions rp ON ur.role_name = rp.role_name AND ur.tenant_id = rp.tenant_id
        LEFT JOIN iam.permissions p ON rp.permission_id = p.permission_id
        WHERE u.user_id = $1 AND u.tenant_id = $2
        GROUP BY u.user_id, u.tenant_id, u.email, u.username, u.first_name, u.last_name, 
                 u.is_active, u.last_login_at, u.created_at, u.updated_at, u.metadata
      `, [userId, tenantId]);

      if (userResult.rows.length === 0) {
        return null;
      }

      const row = userResult.rows[0];
      
      return {
        userId: row.user_id,
        tenantId: row.tenant_id,
        email: row.email,
        username: row.username,
        firstName: row.first_name,
        lastName: row.last_name,
        roles: row.roles || [],
        permissions: row.permissions || [],
        isActive: row.is_active,
        lastLoginAt: row.last_login_at,
        createdAt: row.created_at,
        updatedAt: row.updated_at,
        metadata: row.metadata || {}
      };

    } catch (error) {
      this.logger.error('Erro ao buscar usuário', {
        error: error.message,
        userId,
        tenantId
      });
      return null;
    }
  }

  /**
   * Buscar tenant por ID
   */
  async getTenantById(tenantId: string): Promise<Tenant | null> {
    try {
      const result = await this.pool.query(
        'SELECT * FROM iam.tenants WHERE tenant_id = $1',
        [tenantId]
      );

      if (result.rows.length === 0) {
        return null;
      }

      const row = result.rows[0];
      
      return {
        tenantId: row.tenant_id,
        name: row.name,
        domain: row.domain,
        isActive: row.is_active,
        subscriptionPlan: row.subscription_plan,
        limits: row.limits || {},
        settings: row.settings || {},
        createdAt: row.created_at,
        updatedAt: row.updated_at
      };

    } catch (error) {
      this.logger.error('Erro ao buscar tenant', {
        error: error.message,
        tenantId
      });
      return null;
    }
  }

  /**
   * Gerar access token JWT
   */
  private generateAccessToken(user: User): string {
    const payload = {
      userId: user.userId,
      tenantId: user.tenantId,
      email: user.email,
      roles: user.roles,
      scopes: user.permissions,
      type: 'access'
    };

    return jwt.sign(payload, this.jwtSecret, {
      expiresIn: this.jwtExpiresIn,
      issuer: 'innovabiz-iam',
      audience: 'innovabiz-platform'
    });
  }

  /**
   * Gerar refresh token JWT
   */
  private generateRefreshToken(user: User): string {
    const payload = {
      userId: user.userId,
      tenantId: user.tenantId,
      type: 'refresh'
    };

    return jwt.sign(payload, this.jwtSecret, {
      expiresIn: this.refreshExpiresIn,
      issuer: 'innovabiz-iam',
      audience: 'innovabiz-platform'
    });
  }

  /**
   * Log de evento de auditoria
   */
  private async logAuditEvent(event: AuditEvent): Promise<void> {
    try {
      await this.pool.query(`
        INSERT INTO iam.audit_events (
          event_id, tenant_id, user_id, event_type, resource, action,
          details, ip_address, user_agent, timestamp
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
      `, [
        event.eventId,
        event.tenantId,
        event.userId,
        event.eventType,
        event.resource,
        event.action,
        JSON.stringify(event.details),
        event.ipAddress,
        event.userAgent,
        event.timestamp
      ]);
    } catch (error) {
      this.logger.error('Erro ao registrar evento de auditoria', {
        error: error.message,
        eventId: event.eventId
      });
    }
  }

  /**
   * Log de tentativa de login
   */
  private async logLoginAttempt(attempt: LoginAttempt): Promise<void> {
    try {
      await this.pool.query(`
        INSERT INTO iam.login_attempts (
          user_id, tenant_id, success, ip_address, user_agent, 
          timestamp, failure_reason
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
      `, [
        attempt.userId || null,
        attempt.tenantId,
        attempt.success,
        attempt.ipAddress,
        attempt.userAgent,
        attempt.timestamp,
        attempt.failureReason
      ]);
    } catch (error) {
      this.logger.error('Erro ao registrar tentativa de login', {
        error: error.message,
        tenantId: attempt.tenantId
      });
    }
  }
}