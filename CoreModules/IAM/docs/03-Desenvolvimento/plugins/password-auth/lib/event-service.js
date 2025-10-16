/**
 * Serviço de Eventos de Autenticação
 * 
 * Este módulo gerencia o registro e análise de eventos de autenticação
 * para propósitos de auditoria, segurança e detecção de anomalias.
 * 
 * @module event-service
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { logger } = require('@innovabiz/auth-framework/utils/logging');
const DatabaseService = require('@innovabiz/auth-framework/services/database');
const { v4: uuidv4 } = require('uuid');

/**
 * Serviço para gerenciamento de eventos de autenticação
 */
class EventService {
  /**
   * Cria uma nova instância do serviço de eventos
   */
  constructor() {
    this.db = new DatabaseService();
    this.memoryBuffer = new Map(); // Buffer temporário para otimização
    this.flushInterval = null;
    
    // Inicia o intervalo de despejo do buffer
    this.startBufferFlush();
    
    logger.info('[EventService] Inicializado');
  }
  
  /**
   * Inicia o intervalo de despejo do buffer de memória
   */
  startBufferFlush() {
    // Descarrega eventos armazenados a cada 30 segundos
    this.flushInterval = setInterval(() => this.flushBuffer(), 30000);
  }
  
  /**
   * Para o intervalo de despejo do buffer
   */
  stopBufferFlush() {
    if (this.flushInterval) {
      clearInterval(this.flushInterval);
      this.flushInterval = null;
    }
  }
  
  /**
   * Despeja eventos do buffer para o banco de dados
   */
  async flushBuffer() {
    if (this.memoryBuffer.size === 0) {
      return;
    }
    
    const events = Array.from(this.memoryBuffer.values());
    this.memoryBuffer.clear();
    
    try {
      // Em uma implementação real, usaria inserção em lote para eficiência
      for (const event of events) {
        await this.insertEvent(event);
      }
      
      logger.debug(`[EventService] Descarregados ${events.length} eventos do buffer`);
    } catch (error) {
      logger.error(`[EventService] Erro ao descarregar buffer: ${error.message}`);
      
      // Readiciona eventos ao buffer para tentar novamente
      events.forEach(event => {
        this.memoryBuffer.set(event.event_id, event);
      });
    }
  }
  
  /**
   * Insere um evento no banco de dados
   * 
   * @param {Object} event Evento a inserir
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async insertEvent(event) {
    try {
      const query = `
        INSERT INTO iam.authentication_events (
          event_id, tenant_id, user_id, auth_method_id, 
          event_type, status, client_ip, user_agent, 
          risk_score, risk_factors, region, geolocation, 
          session_id, correlation_id, metadata
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        )
      `;
      
      await this.db.query(query, [
        event.event_id,
        event.tenant_id,
        event.user_id,
        event.auth_method_id,
        event.event_type,
        event.status,
        event.client_ip,
        event.user_agent,
        event.risk_score || 0,
        event.risk_factors ? JSON.stringify(event.risk_factors) : null,
        event.region || null,
        event.geolocation ? JSON.stringify(event.geolocation) : null,
        event.session_id || null,
        event.correlation_id || null,
        event.metadata ? JSON.stringify(event.metadata) : null
      ]);
      
      return true;
    } catch (error) {
      logger.error(`[EventService] Erro ao inserir evento: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Registra um evento de autenticação
   * 
   * @param {Object} event Dados do evento
   * @returns {string} ID do evento criado
   */
  logEvent(event) {
    const eventId = event.event_id || uuidv4();
    
    const eventRecord = {
      ...event,
      event_id: eventId,
      timestamp: event.timestamp || new Date(),
    };
    
    // Adiciona ao buffer de memória para inserção em lote
    this.memoryBuffer.set(eventId, eventRecord);
    
    // Se o buffer ficar muito grande, força o despejo
    if (this.memoryBuffer.size >= 100) {
      this.flushBuffer().catch(err => {
        logger.error(`[EventService] Erro ao descarregar buffer cheio: ${err.message}`);
      });
    }
    
    return eventId;
  }
  
  /**
   * Registra um evento de início de autenticação
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logAuthStart(data) {
    return this.logEvent({
      ...data,
      event_type: 'auth_start',
      status: 'pending'
    });
  }
  
  /**
   * Registra um evento de autenticação bem-sucedida
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logAuthSuccess(data) {
    return this.logEvent({
      ...data,
      event_type: 'auth_complete',
      status: 'success'
    });
  }
  
  /**
   * Registra um evento de falha de autenticação
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logAuthFailure(data) {
    return this.logEvent({
      ...data,
      event_type: 'auth_complete',
      status: 'failure'
    });
  }
  
  /**
   * Registra um evento de redefinição de senha
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logPasswordReset(data) {
    return this.logEvent({
      ...data,
      event_type: 'password_reset',
      status: data.status || 'success'
    });
  }
  
  /**
   * Registra um evento de alteração de senha
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logPasswordChange(data) {
    return this.logEvent({
      ...data,
      event_type: 'password_change',
      status: data.status || 'success'
    });
  }
  
  /**
   * Registra um evento de bloqueio de conta
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logAccountLockout(data) {
    return this.logEvent({
      ...data,
      event_type: 'account_lockout',
      status: 'locked'
    });
  }
  
  /**
   * Registra um evento de desbloqueio de conta
   * 
   * @param {Object} data Dados do evento
   * @returns {string} ID do evento criado
   */
  logAccountUnlock(data) {
    return this.logEvent({
      ...data,
      event_type: 'account_unlock',
      status: 'unlocked'
    });
  }
  
  /**
   * Busca eventos recentes para um usuário
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {number} limit Limite de resultados
   * @returns {Promise<Array>} Lista de eventos
   */
  async getRecentEvents(tenantId, userId, limit = 10) {
    try {
      const query = `
        SELECT * FROM iam.authentication_events
        WHERE tenant_id = $1
        AND user_id = $2
        ORDER BY timestamp DESC
        LIMIT $3
      `;
      
      const result = await this.db.query(query, [tenantId, userId, limit]);
      return result.rows;
    } catch (error) {
      logger.error(`[EventService] Erro ao buscar eventos recentes: ${error.message}`);
      return [];
    }
  }
  
  /**
   * Busca eventos com base em critérios de filtro
   * 
   * @param {Object} filters Critérios de filtro
   * @param {number} limit Limite de resultados
   * @param {number} offset Deslocamento para paginação
   * @returns {Promise<Object>} Resultados e contagem total
   */
  async searchEvents(filters = {}, limit = 50, offset = 0) {
    try {
      let whereClause = 'WHERE 1=1';
      const params = [];
      let paramIndex = 1;
      
      // Constrói a cláusula WHERE dinamicamente baseada nos filtros
      if (filters.tenantId) {
        whereClause += ` AND tenant_id = $${paramIndex++}`;
        params.push(filters.tenantId);
      }
      
      if (filters.userId) {
        whereClause += ` AND user_id = $${paramIndex++}`;
        params.push(filters.userId);
      }
      
      if (filters.eventTypes && filters.eventTypes.length > 0) {
        whereClause += ` AND event_type IN (${filters.eventTypes.map(() => `$${paramIndex++}`).join(',')})`;
        params.push(...filters.eventTypes);
      }
      
      if (filters.status) {
        whereClause += ` AND status = $${paramIndex++}`;
        params.push(filters.status);
      }
      
      if (filters.clientIp) {
        whereClause += ` AND client_ip = $${paramIndex++}`;
        params.push(filters.clientIp);
      }
      
      if (filters.startDate) {
        whereClause += ` AND timestamp >= $${paramIndex++}`;
        params.push(new Date(filters.startDate));
      }
      
      if (filters.endDate) {
        whereClause += ` AND timestamp <= $${paramIndex++}`;
        params.push(new Date(filters.endDate));
      }
      
      if (filters.region) {
        whereClause += ` AND region = $${paramIndex++}`;
        params.push(filters.region);
      }
      
      // Consulta de contagem total
      const countQuery = `
        SELECT COUNT(*) as total
        FROM iam.authentication_events
        ${whereClause}
      `;
      
      // Consulta principal
      const mainQuery = `
        SELECT *
        FROM iam.authentication_events
        ${whereClause}
        ORDER BY timestamp DESC
        LIMIT $${paramIndex++} OFFSET $${paramIndex++}
      `;
      
      // Adiciona parâmetros de paginação
      params.push(limit, offset);
      
      // Executa as consultas
      const [countResult, dataResult] = await Promise.all([
        this.db.query(countQuery, params.slice(0, -2)),
        this.db.query(mainQuery, params)
      ]);
      
      return {
        total: parseInt(countResult.rows[0].total),
        events: dataResult.rows,
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
        totalPages: Math.ceil(parseInt(countResult.rows[0].total) / limit)
      };
      
    } catch (error) {
      logger.error(`[EventService] Erro na busca de eventos: ${error.message}`);
      return {
        total: 0,
        events: [],
        page: 1,
        pageSize: limit,
        totalPages: 0
      };
    }
  }
  
  /**
   * Calcula métricas de eventos de autenticação
   * 
   * @param {string} tenantId ID do tenant
   * @param {Object} filters Filtros adicionais
   * @returns {Promise<Object>} Métricas calculadas
   */
  async calculateMetrics(tenantId, filters = {}) {
    try {
      // Constrói condição de data
      let dateCondition = '';
      const params = [tenantId];
      let paramIndex = 2;
      
      if (filters.startDate && filters.endDate) {
        dateCondition = `AND timestamp BETWEEN $${paramIndex++} AND $${paramIndex++}`;
        params.push(new Date(filters.startDate), new Date(filters.endDate));
      } else if (filters.startDate) {
        dateCondition = `AND timestamp >= $${paramIndex++}`;
        params.push(new Date(filters.startDate));
      } else if (filters.endDate) {
        dateCondition = `AND timestamp <= $${paramIndex++}`;
        params.push(new Date(filters.endDate));
      } else {
        // Padrão: últimos 30 dias
        dateCondition = `AND timestamp >= NOW() - INTERVAL '30 days'`;
      }
      
      // Consulta principal para métricas
      const query = `
        SELECT
          COUNT(*) as total_events,
          SUM(CASE WHEN event_type = 'auth_complete' AND status = 'success' THEN 1 ELSE 0 END) as successful_auths,
          SUM(CASE WHEN event_type = 'auth_complete' AND status = 'failure' THEN 1 ELSE 0 END) as failed_auths,
          SUM(CASE WHEN event_type = 'account_lockout' THEN 1 ELSE 0 END) as account_lockouts,
          SUM(CASE WHEN event_type = 'password_reset' THEN 1 ELSE 0 END) as password_resets,
          SUM(CASE WHEN event_type = 'password_change' THEN 1 ELSE 0 END) as password_changes,
          COUNT(DISTINCT user_id) as unique_users,
          AVG(risk_score) as avg_risk_score
        FROM iam.authentication_events
        WHERE tenant_id = $1
        ${dateCondition}
      `;
      
      // Consulta de distribuição por região
      const regionQuery = `
        SELECT region, COUNT(*) as count
        FROM iam.authentication_events
        WHERE tenant_id = $1
        ${dateCondition}
        AND region IS NOT NULL
        GROUP BY region
        ORDER BY count DESC
      `;
      
      // Consulta de eventos por hora do dia
      const hourlyQuery = `
        SELECT EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as count
        FROM iam.authentication_events
        WHERE tenant_id = $1
        ${dateCondition}
        GROUP BY hour
        ORDER BY hour
      `;
      
      // Executa as consultas em paralelo
      const [metricsResult, regionResult, hourlyResult] = await Promise.all([
        this.db.query(query, params),
        this.db.query(regionQuery, params),
        this.db.query(hourlyQuery, params)
      ]);
      
      // Formata o resultado
      const metrics = metricsResult.rows[0];
      
      // Calcula taxas
      const successRate = metrics.total_events > 0 
        ? (metrics.successful_auths / (metrics.successful_auths + metrics.failed_auths)) * 100
        : 0;
      
      return {
        summary: {
          totalEvents: parseInt(metrics.total_events),
          successfulAuths: parseInt(metrics.successful_auths),
          failedAuths: parseInt(metrics.failed_auths),
          successRate: parseFloat(successRate.toFixed(2)),
          accountLockouts: parseInt(metrics.account_lockouts),
          passwordResets: parseInt(metrics.password_resets),
          passwordChanges: parseInt(metrics.password_changes),
          uniqueUsers: parseInt(metrics.unique_users),
          averageRiskScore: parseFloat(parseFloat(metrics.avg_risk_score || 0).toFixed(2))
        },
        regional: regionResult.rows.map(row => ({
          region: row.region,
          count: parseInt(row.count)
        })),
        hourly: hourlyResult.rows.map(row => ({
          hour: parseInt(row.hour),
          count: parseInt(row.count)
        }))
      };
      
    } catch (error) {
      logger.error(`[EventService] Erro ao calcular métricas: ${error.message}`);
      return {
        summary: {
          totalEvents: 0,
          successfulAuths: 0,
          failedAuths: 0,
          successRate: 0,
          accountLockouts: 0,
          passwordResets: 0,
          passwordChanges: 0,
          uniqueUsers: 0,
          averageRiskScore: 0
        },
        regional: [],
        hourly: []
      };
    }
  }
  
  /**
   * Verifica a saúde do serviço
   * 
   * @returns {Promise<Object>} Status de saúde
   */
  async checkHealth() {
    try {
      // Verifica conexão com banco de dados
      await this.db.query('SELECT 1');
      
      return {
        status: 'healthy',
        details: {
          pendingEvents: this.memoryBuffer.size,
          bufferActive: this.flushInterval !== null
        },
        timestamp: new Date()
      };
    } catch (error) {
      return {
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Limpa recursos ao destruir a instância
   */
  destroy() {
    this.stopBufferFlush();
    
    // Despeja eventos restantes
    if (this.memoryBuffer.size > 0) {
      this.flushBuffer().catch(err => {
        logger.error(`[EventService] Erro ao descarregar buffer final: ${err.message}`);
      });
    }
  }
}

module.exports = EventService;
