/**
 * ============================================================================
 * INNOVABIZ IAM - Main Entry Point
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Ponto de entrada principal para o servidor WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import dotenv from 'dotenv';
import { Pool } from 'pg';
import { Redis } from 'ioredis';
import { Kafka } from 'kafkajs';
import winston from 'winston';
import promClient from 'prom-client';

import { WebAuthnApp } from './app';
import { validateConfig } from './config/webauthn';

// Carregar variáveis de ambiente
dotenv.config();

/**
 * Configuração do logger
 */
function createLogger(): winston.Logger {
  const logLevel = process.env.LOG_LEVEL || 'info';
  const logFormat = process.env.LOG_FORMAT || 'json';

  const formats = [
    winston.format.timestamp(),
    winston.format.errors({ stack: true })
  ];

  if (logFormat === 'json') {
    formats.push(winston.format.json());
  } else {
    formats.push(winston.format.simple());
  }

  return winston.createLogger({
    level: logLevel,
    format: winston.format.combine(...formats),
    defaultMeta: {
      service: 'innovabiz-iam-webauthn',
      version: '1.0.0'
    },
    transports: [
      new winston.transports.Console({
        handleExceptions: true,
        handleRejections: true
      }),
      new winston.transports.File({
        filename: 'logs/error.log',
        level: 'error',
        handleExceptions: true
      }),
      new winston.transports.File({
        filename: 'logs/combined.log',
        handleExceptions: true
      })
    ]
  });
}

/**
 * Configuração do PostgreSQL
 */
function createDatabasePool(logger: winston.Logger): Pool {
  const config = {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432'),
    database: process.env.DB_NAME || 'innovabiz_iam',
    user: process.env.DB_USER || 'postgres',
    password: process.env.DB_PASSWORD || '',
    max: parseInt(process.env.DB_POOL_MAX || '20'),
    idleTimeoutMillis: parseInt(process.env.DB_IDLE_TIMEOUT || '30000'),
    connectionTimeoutMillis: parseInt(process.env.DB_CONNECTION_TIMEOUT || '2000'),
    ssl: process.env.DB_SSL === 'true' ? {
      rejectUnauthorized: process.env.DB_SSL_REJECT_UNAUTHORIZED !== 'false'
    } : false
  };

  logger.info('Connecting to PostgreSQL', {
    host: config.host,
    port: config.port,
    database: config.database,
    user: config.user,
    ssl: !!config.ssl
  });

  const pool = new Pool(config);

  // Event handlers
  pool.on('connect', (client) => {
    logger.debug('New PostgreSQL client connected');
  });

  pool.on('error', (err, client) => {
    logger.error('PostgreSQL client error', {
      error: err.message,
      stack: err.stack
    });
  });

  return pool;
}

/**
 * Configuração do Redis
 */
function createRedisClient(logger: winston.Logger): Redis {
  const config = {
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT || '6379'),
    password: process.env.REDIS_PASSWORD || undefined,
    db: parseInt(process.env.REDIS_DB || '0'),
    retryDelayOnFailover: 100,
    enableReadyCheck: false,
    maxRetriesPerRequest: 3,
    lazyConnect: true,
    keepAlive: 30000,
    connectTimeout: 10000,
    commandTimeout: 5000
  };

  logger.info('Connecting to Redis', {
    host: config.host,
    port: config.port,
    db: config.db
  });

  const redis = new Redis(config);

  // Event handlers
  redis.on('connect', () => {
    logger.info('Redis connected');
  });

  redis.on('ready', () => {
    logger.info('Redis ready');
  });

  redis.on('error', (err) => {
    logger.error('Redis error', {
      error: err.message,
      stack: err.stack
    });
  });

  redis.on('close', () => {
    logger.warn('Redis connection closed');
  });

  redis.on('reconnecting', () => {
    logger.info('Redis reconnecting');
  });

  return redis;
}

/**
 * Configuração do Kafka (opcional)
 */
function createKafkaClient(logger: winston.Logger): Kafka | undefined {
  const brokers = process.env.KAFKA_BROKERS;
  
  if (!brokers) {
    logger.info('Kafka not configured, audit events will only be stored in database');
    return undefined;
  }

  const config = {
    clientId: process.env.KAFKA_CLIENT_ID || 'innovabiz-iam-webauthn',
    brokers: brokers.split(','),
    ssl: process.env.KAFKA_SSL === 'true',
    sasl: process.env.KAFKA_SASL_MECHANISM ? {
      mechanism: process.env.KAFKA_SASL_MECHANISM as any,
      username: process.env.KAFKA_SASL_USERNAME || '',
      password: process.env.KAFKA_SASL_PASSWORD || ''
    } : undefined,
    retry: {
      initialRetryTime: 100,
      retries: 8
    },
    connectionTimeout: 3000,
    requestTimeout: 30000
  };

  logger.info('Connecting to Kafka', {
    clientId: config.clientId,
    brokers: config.brokers,
    ssl: config.ssl,
    sasl: !!config.sasl
  });

  return new Kafka(config);
}

/**
 * Configuração das métricas Prometheus
 */
function setupPrometheusMetrics(): void {
  // Coletar métricas padrão do Node.js
  promClient.collectDefaultMetrics({
    prefix: 'innovabiz_webauthn_nodejs_',
    timeout: 5000,
    gcDurationBuckets: [0.001, 0.01, 0.1, 1, 2, 5]
  });

  // Registrar métricas customizadas
  promClient.register.setDefaultLabels({
    app: 'innovabiz-iam-webauthn',
    version: '1.0.0',
    environment: process.env.NODE_ENV || 'development'
  });
}

/**
 * Verificação de saúde das dependências
 */
async function healthCheck(
  logger: winston.Logger,
  db: Pool,
  redis: Redis,
  kafka?: Kafka
): Promise<void> {
  logger.info('Performing health check');

  try {
    // Verificar PostgreSQL
    const dbResult = await db.query('SELECT NOW() as current_time');
    logger.info('PostgreSQL health check passed', {
      currentTime: dbResult.rows[0].current_time
    });

    // Verificar Redis
    const redisPong = await redis.ping();
    logger.info('Redis health check passed', {
      response: redisPong
    });

    // Verificar Kafka (se configurado)
    if (kafka) {
      const admin = kafka.admin();
      await admin.connect();
      const metadata = await admin.fetchTopicMetadata();
      await admin.disconnect();
      logger.info('Kafka health check passed', {
        topicsCount: metadata.topics.length
      });
    }

    logger.info('All health checks passed');

  } catch (error) {
    logger.error('Health check failed', {
      error: error.message,
      stack: error.stack
    });
    throw error;
  }
}

/**
 * Função principal
 */
async function main(): Promise<void> {
  const logger = createLogger();
  
  try {
    logger.info('Starting INNOVABIZ IAM WebAuthn Server', {
      version: '1.0.0',
      nodeVersion: process.version,
      platform: process.platform,
      environment: process.env.NODE_ENV || 'development'
    });

    // Validar configuração WebAuthn
    validateConfig();
    logger.info('WebAuthn configuration validated');

    // Configurar métricas
    setupPrometheusMetrics();
    logger.info('Prometheus metrics configured');

    // Criar conexões
    const db = createDatabasePool(logger);
    const redis = createRedisClient(logger);
    const kafka = createKafkaClient(logger);

    // Conectar ao Redis
    await redis.connect();

    // Verificar saúde das dependências
    await healthCheck(logger, db, redis, kafka);

    // Criar e iniciar aplicação
    const app = new WebAuthnApp(logger, db, redis, kafka);
    const port = parseInt(process.env.PORT || '3000');
    
    await app.start(port);

    logger.info('INNOVABIZ IAM WebAuthn Server started successfully', {
      port,
      environment: process.env.NODE_ENV || 'development'
    });

  } catch (error) {
    logger.error('Failed to start server', {
      error: error.message,
      stack: error.stack
    });
    process.exit(1);
  }
}

/**
 * Tratamento de sinais do sistema
 */
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully');
  process.exit(0);
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down gracefully');
  process.exit(0);
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// Iniciar aplicação
if (require.main === module) {
  main().catch((error) => {
    console.error('Fatal error during startup:', error);
    process.exit(1);
  });
}

export { main, WebAuthnApp };