/**
 * Monitor de Risco e Fraude
 * 
 * Componente responsável por avaliar o nível de risco e detectar possíveis
 * tentativas de fraude em processos de autenticação. Implementa análise
 * contextual, comportamental e baseada em anomalias.
 * 
 * @module risk-monitor
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { logger } = require('../utils/logging');
const { v4: uuidv4 } = require('uuid');

/**
 * Classe de monitoramento de risco
 */
class RiskMonitor {
  /**
   * Construtor do monitor de risco
   * 
   * @param {Object} framework Instância do framework de autenticação
   */
  constructor(framework) {
    this.framework = framework;
    
    // Cache de dados de avaliação
    this.assessmentCache = new Map();
    
    // Registro de padrões de comportamento por usuário
    this.behaviorProfiles = new Map();
    
    // Dicionário de dispositivos conhecidos
    this.knownDevices = new Map();
    
    // Registro de anomalias detectadas
    this.anomalyRegistry = new Map();
    
    // Motor de regras de risco
    this.ruleEngine = null;
    
    // Configurações de avaliação
    this.settings = {
      cacheDuration: 30 * 60 * 1000, // 30 minutos em ms
      ipReputationEnabled: true,
      behavioralAnalysisEnabled: true,
      locationAnalysisEnabled: true,
      deviceAnalysisEnabled: true,
      userProfileAnalysisEnabled: true,
      anomalyDetectionEnabled: true,
      threatIntelEnabled: true,
      riskThresholds: {
        low: 30,
        medium: 60,
        high: 90
      }
    };
    
    logger.info('[RiskMonitor] Monitor de risco instanciado');
  }
  
  /**
   * Inicializa o monitor de risco
   * 
   * @returns {Promise<void>}
   */
  async initialize() {
    try {
      // Inicializa o motor de regras
      this.initializeRuleEngine();
      
      // Carrega dados iniciais
      await this.loadInitialData();
      
      logger.info('[RiskMonitor] Monitor de risco inicializado com sucesso');
    } catch (error) {
      logger.error(`[RiskMonitor] Erro ao inicializar monitor de risco: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Inicializa o motor de regras
   */
  initializeRuleEngine() {
    // Implementação básica do motor de regras
    this.ruleEngine = {
      /**
       * Avalia um contexto contra regras predefinidas
       * 
       * @param {Object} context Contexto a avaliar
       * @returns {Object} Pontuação e detalhes da avaliação
       */
      evaluate: (context) => {
        const assessment = {
          score: 0,
          factors: [],
          details: {}
        };
        
        // Implementação simplificada de regras para demonstração
        // Em produção: motor de regras mais sofisticado com pesos dinâmicos
        
        // Análise de IP
        if (context.ip) {
          if (this.isKnownBadIP(context.ip)) {
            assessment.score += 75;
            assessment.factors.push('known_bad_ip');
            assessment.details.ip = { risk: 'high', reason: 'known_bad_ip' };
          } else if (this.isAnonymousIP(context.ip)) {
            assessment.score += 40;
            assessment.factors.push('anonymous_ip');
            assessment.details.ip = { risk: 'medium', reason: 'anonymous_ip' };
          }
        }
        
        // Análise de localização
        if (context.geolocation) {
          if (context.userId && this.isUnusualLocation(context.userId, context.geolocation)) {
            assessment.score += 50;
            assessment.factors.push('unusual_location');
            assessment.details.location = { risk: 'medium', reason: 'unusual_location' };
          }
          
          if (this.isHighRiskLocation(context.geolocation)) {
            assessment.score += 30;
            assessment.factors.push('high_risk_location');
            assessment.details.location = {
              ...(assessment.details.location || {}),
              risk: 'medium',
              reason: 'high_risk_location'
            };
          }
        }
        
        // Análise de dispositivo
        if (context.deviceId) {
          if (!this.isKnownDevice(context.userId, context.deviceId)) {
            assessment.score += 40;
            assessment.factors.push('unknown_device');
            assessment.details.device = { risk: 'medium', reason: 'unknown_device' };
          }
          
          if (this.hasDeviceAnomalies(context.deviceId, context.deviceAttributes)) {
            assessment.score += 30;
            assessment.factors.push('device_anomalies');
            assessment.details.device = {
              ...(assessment.details.device || {}),
              risk: 'medium',
              reason: 'device_anomalies'
            };
          }
        }
        
        // Análise temporal
        if (context.timestamp) {
          if (context.userId && this.isUnusualTime(context.userId, context.timestamp)) {
            assessment.score += 25;
            assessment.factors.push('unusual_time');
            assessment.details.time = { risk: 'low', reason: 'unusual_time' };
          }
        }
        
        // Análise de padrão comportamental
        if (context.userId && this.settings.behavioralAnalysisEnabled) {
          const behaviorScore = this.analyzeBehavior(context.userId, context);
          
          if (behaviorScore > 40) {
            assessment.score += behaviorScore;
            assessment.factors.push('behavioral_anomaly');
            assessment.details.behavior = { risk: 'high', reason: 'behavioral_anomaly', score: behaviorScore };
          }
        }
        
        // Análise de velocidade de movimento
        if (context.previousGeolocation && context.geolocation && context.previousTimestamp && context.timestamp) {
          const impossibleTravel = this.detectImpossibleTravel(
            context.previousGeolocation,
            context.geolocation,
            context.previousTimestamp,
            context.timestamp
          );
          
          if (impossibleTravel) {
            assessment.score += 80;
            assessment.factors.push('impossible_travel');
            assessment.details.travel = { risk: 'high', reason: 'impossible_travel' };
          }
        }
        
        // Determina nível de risco com base na pontuação
        let level;
        if (assessment.score >= this.settings.riskThresholds.high) {
          level = 'high';
        } else if (assessment.score >= this.settings.riskThresholds.medium) {
          level = 'medium';
        } else {
          level = 'low';
        }
        
        return {
          id: uuidv4(),
          timestamp: new Date(),
          score: assessment.score,
          level,
          factors: assessment.factors,
          details: assessment.details,
          context: this.sanitizeContext(context)
        };
      }
    };
    
    logger.debug('[RiskMonitor] Motor de regras inicializado');
  }
  
  /**
   * Carrega dados iniciais
   * 
   * @returns {Promise<void>}
   */
  async loadInitialData() {
    // Em produção: carregar de banco de dados, serviços de inteligência, etc.
    // Para esta demonstração, usamos dados simulados
    
    logger.debug('[RiskMonitor] Dados iniciais carregados');
  }
  
  /**
   * Avalia o risco de uma requisição de autenticação
   * 
   * @param {Object} request Requisição de autenticação
   * @param {Object} context Contexto da autenticação
   * @returns {Promise<Object>} Resultado da avaliação de risco
   */
  async assessRisk(request, context = {}) {
    try {
      // Constrói contexto de avaliação
      const assessmentContext = this.buildAssessmentContext(request, context);
      
      // Consulta cache se disponível
      const cacheKey = this.generateCacheKey(assessmentContext);
      if (this.assessmentCache.has(cacheKey)) {
        const cachedAssessment = this.assessmentCache.get(cacheKey);
        
        // Verifica validade do cache
        if (Date.now() - cachedAssessment.timestamp < this.settings.cacheDuration) {
          logger.debug(`[RiskMonitor] Usando avaliação em cache: ${cachedAssessment.id}`);
          return cachedAssessment;
        }
      }
      
      // Avalia risco usando o motor de regras
      const assessment = this.ruleEngine.evaluate(assessmentContext);
      
      // Armazena em cache
      this.assessmentCache.set(cacheKey, assessment);
      
      // Registra a avaliação para análise futura
      this.recordAssessment(assessment);
      
      // Log de avaliação de alto risco
      if (assessment.level === 'high') {
        logger.warn(`[RiskMonitor] Avaliação de alto risco: ${assessment.id}, score: ${assessment.score}, fatores: ${assessment.factors.join(', ')}`);
      } else {
        logger.debug(`[RiskMonitor] Avaliação de risco: ${assessment.id}, nível: ${assessment.level}, score: ${assessment.score}`);
      }
      
      return assessment;
    } catch (error) {
      logger.error(`[RiskMonitor] Erro ao avaliar risco: ${error.message}`);
      
      // Retorna avaliação padrão em caso de erro
      return {
        id: uuidv4(),
        timestamp: new Date(),
        score: 50, // Médio por padrão em caso de erro
        level: 'medium',
        factors: ['evaluation_error'],
        details: { error: error.message },
        context: this.sanitizeContext({ ...request, ...context })
      };
    }
  }
  
  /**
   * Constrói o contexto para avaliação de risco
   * 
   * @param {Object} request Requisição de autenticação
   * @param {Object} context Contexto da autenticação
   * @returns {Object} Contexto de avaliação
   */
  buildAssessmentContext(request, context) {
    // Combina e enriquece o contexto com dados para avaliação
    return {
      // Dados da requisição
      userId: request.userId,
      tenantId: request.tenantId,
      appId: request.appId,
      ip: request.ip || context.ip,
      
      // Dados de dispositivo
      deviceId: request.deviceId || context.deviceId,
      deviceAttributes: {
        ...(request.deviceAttributes || {}),
        ...(context.deviceAttributes || {})
      },
      userAgent: request.userAgent || context.userAgent,
      
      // Dados de localização
      geolocation: request.geolocation || context.geolocation,
      network: request.network || context.network,
      
      // Dados temporais
      timestamp: new Date(),
      
      // Dados de contexto
      methodCode: request.methodCode || context.methodCode,
      methodType: request.methodType || context.methodType,
      resourceId: request.resourceId,
      resourceType: request.resourceType,
      
      // Dados históricos
      previousGeolocation: context.previousGeolocation,
      previousTimestamp: context.previousTimestamp,
      failedAttempts: context.failedAttempts || 0,
      lastSuccessfulLogin: context.lastSuccessfulLogin,
      
      // Dados de sessão
      sessionId: context.sessionId,
      authenticationType: context.authenticationType,
      
      // Metadados
      correlationId: context.correlationId,
      requestedAt: context.timestamp || new Date()
    };
  }
  
  /**
   * Gera uma chave para cache de avaliação
   * 
   * @param {Object} context Contexto de avaliação
   * @returns {string} Chave de cache
   */
  generateCacheKey(context) {
    // Extrai componentes principais para a chave
    const components = [
      context.userId,
      context.ip,
      context.deviceId,
      context.methodCode,
      context.resourceId
    ].filter(Boolean);
    
    return components.join(':');
  }
  
  /**
   * Registra uma avaliação para análise futura
   * 
   * @param {Object} assessment Avaliação de risco
   */
  recordAssessment(assessment) {
    // Em produção: persistir em banco de dados, enviar para análise, etc.
    // Para esta demonstração, apenas um log
    
    // Atualiza perfil comportamental se usuário presente
    if (assessment.context.userId) {
      this.updateBehaviorProfile(assessment.context.userId, assessment);
    }
    
    // Registra dispositivo se for nova avaliação bem-sucedida
    if (assessment.level === 'low' && assessment.context.deviceId && assessment.context.userId) {
      this.recordKnownDevice(assessment.context.userId, assessment.context.deviceId, assessment.context);
    }
  }
  
  /**
   * Atualiza o perfil comportamental de um usuário
   * 
   * @param {string} userId ID do usuário
   * @param {Object} assessment Avaliação de risco
   */
  updateBehaviorProfile(userId, assessment) {
    // Em produção: atualizações incrementais sofisticadas, análise estatística
    // Para esta demonstração, implementação simplificada
    
    if (!this.behaviorProfiles.has(userId)) {
      this.behaviorProfiles.set(userId, {
        locations: [],
        devices: new Set(),
        timePatterns: [],
        lastUpdated: new Date()
      });
    }
    
    const profile = this.behaviorProfiles.get(userId);
    
    // Atualiza localizações conhecidas
    if (assessment.context.geolocation) {
      // Em produção: agrupamento geográfico, frequência, etc.
      profile.locations.push({
        latitude: assessment.context.geolocation.latitude,
        longitude: assessment.context.geolocation.longitude,
        timestamp: new Date(),
        assessmentId: assessment.id
      });
      
      // Manter apenas as 10 mais recentes para demo
      if (profile.locations.length > 10) {
        profile.locations.shift();
      }
    }
    
    // Atualiza dispositivos
    if (assessment.context.deviceId) {
      profile.devices.add(assessment.context.deviceId);
    }
    
    // Atualiza padrões temporais
    if (assessment.context.timestamp) {
      const timestamp = new Date(assessment.context.timestamp);
      const timeEntry = {
        hour: timestamp.getHours(),
        minute: timestamp.getMinutes(),
        dayOfWeek: timestamp.getDay(),
        timestamp: new Date(),
        assessmentId: assessment.id
      };
      
      profile.timePatterns.push(timeEntry);
      
      // Manter apenas os 30 mais recentes para demo
      if (profile.timePatterns.length > 30) {
        profile.timePatterns.shift();
      }
    }
    
    profile.lastUpdated = new Date();
  }
  
  /**
   * Registra um dispositivo conhecido
   * 
   * @param {string} userId ID do usuário
   * @param {string} deviceId ID do dispositivo
   * @param {Object} context Contexto adicional
   */
  recordKnownDevice(userId, deviceId, context) {
    const key = `${userId}:${deviceId}`;
    
    if (!this.knownDevices.has(key)) {
      this.knownDevices.set(key, {
        userId,
        deviceId,
        firstSeen: new Date(),
        attributes: {},
        locations: []
      });
    }
    
    const device = this.knownDevices.get(key);
    
    // Atualiza atributos
    if (context.deviceAttributes) {
      device.attributes = {
        ...device.attributes,
        ...context.deviceAttributes
      };
    }
    
    // Atualiza localizações
    if (context.geolocation) {
      device.locations.push({
        ...context.geolocation,
        timestamp: new Date()
      });
      
      // Manter apenas as 5 mais recentes
      if (device.locations.length > 5) {
        device.locations.shift();
      }
    }
    
    device.lastSeen = new Date();
  }
  
  /**
   * Verifica se um IP é conhecido como malicioso
   * 
   * @param {string} ip Endereço IP
   * @returns {boolean} Verdadeiro se IP é malicioso
   */
  isKnownBadIP(ip) {
    // Em produção: consulta a banco de dados, serviços de reputação, etc.
    // Para esta demonstração, implementação fictícia
    
    // IPs maliciosos de exemplo
    const badIps = [
      '192.0.2.1',
      '198.51.100.1',
      '203.0.113.1'
    ];
    
    return badIps.includes(ip);
  }
  
  /**
   * Verifica se um IP é de serviço anônimo (VPN, proxy, Tor)
   * 
   * @param {string} ip Endereço IP
   * @returns {boolean} Verdadeiro se IP é anônimo
   */
  isAnonymousIP(ip) {
    // Em produção: consulta a serviços especializados
    // Para esta demonstração, implementação fictícia
    
    // Simulação para demonstração
    return ip && ip.startsWith('10.') || ip.includes('proxy');
  }
  
  /**
   * Verifica se uma localização é incomum para um usuário
   * 
   * @param {string} userId ID do usuário
   * @param {Object} geolocation Dados de geolocalização
   * @returns {boolean} Verdadeiro se localização é incomum
   */
  isUnusualLocation(userId, geolocation) {
    if (!geolocation || !userId || !this.behaviorProfiles.has(userId)) {
      return false;
    }
    
    const profile = this.behaviorProfiles.get(userId);
    
    // Se não há localizações conhecidas, qualquer nova é considerada incomum
    if (!profile.locations || profile.locations.length === 0) {
      return true;
    }
    
    // Em produção: cálculos mais sofisticados, clusters, frequência, etc.
    // Para esta demonstração, verificação simples de distância
    
    // Verifica se há alguma localização próxima
    for (const knownLocation of profile.locations) {
      const distance = this.calculateDistance(
        geolocation.latitude, geolocation.longitude,
        knownLocation.latitude, knownLocation.longitude
      );
      
      // Se estiver a menos de 100km de uma localização conhecida, não é incomum
      if (distance < 100) {
        return false;
      }
    }
    
    // Se chegou aqui, não encontrou nenhuma localização próxima
    return true;
  }
  
  /**
   * Calcula a distância entre duas coordenadas (fórmula haversine)
   * 
   * @param {number} lat1 Latitude do ponto 1
   * @param {number} lon1 Longitude do ponto 1
   * @param {number} lat2 Latitude do ponto 2
   * @param {number} lon2 Longitude do ponto 2
   * @returns {number} Distância em quilômetros
   */
  calculateDistance(lat1, lon1, lat2, lon2) {
    const R = 6371; // Raio da Terra em km
    const dLat = this.toRadians(lat2 - lat1);
    const dLon = this.toRadians(lon2 - lon1);
    
    const a = 
      Math.sin(dLat/2) * Math.sin(dLat/2) +
      Math.cos(this.toRadians(lat1)) * Math.cos(this.toRadians(lat2)) * 
      Math.sin(dLon/2) * Math.sin(dLon/2);
    
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    const distance = R * c;
    
    return distance;
  }
  
  /**
   * Converte graus para radianos
   * 
   * @param {number} degrees Ângulo em graus
   * @returns {number} Ângulo em radianos
   */
  toRadians(degrees) {
    return degrees * Math.PI / 180;
  }
  
  /**
   * Verifica se uma localização é considerada de alto risco
   * 
   * @param {Object} geolocation Dados de geolocalização
   * @returns {boolean} Verdadeiro se localização é de alto risco
   */
  isHighRiskLocation(geolocation) {
    // Em produção: consulta a banco de dados de locais de risco, serviços, etc.
    // Para esta demonstração, implementação fictícia
    
    // Verifica país (código ISO)
    if (geolocation && geolocation.country) {
      // Lista fictícia de países/regiões com alto índice de fraude
      const highRiskCountries = ['XY', 'ZZ', 'YY'];
      
      return highRiskCountries.includes(geolocation.country);
    }
    
    return false;
  }
  
  /**
   * Verifica se um dispositivo é conhecido para um usuário
   * 
   * @param {string} userId ID do usuário
   * @param {string} deviceId ID do dispositivo
   * @returns {boolean} Verdadeiro se dispositivo é conhecido
   */
  isKnownDevice(userId, deviceId) {
    if (!userId || !deviceId) {
      return false;
    }
    
    return this.knownDevices.has(`${userId}:${deviceId}`);
  }
  
  /**
   * Verifica anomalias em atributos de dispositivo
   * 
   * @param {string} deviceId ID do dispositivo
   * @param {Object} attributes Atributos atuais do dispositivo
   * @returns {boolean} Verdadeiro se há anomalias
   */
  hasDeviceAnomalies(deviceId, attributes) {
    // Em produção: análises mais complexas, fingerprinting, etc.
    // Para esta demonstração, lógica simplificada
    
    // Verificação básica para demonstração
    // Retorna falso se não há dados suficientes
    if (!deviceId || !attributes) {
      return false;
    }
    
    // Verificações simuladas para demonstração
    if (attributes.emulated === true || 
        attributes.rooted === true || 
        attributes.developer_mode === true) {
      return true;
    }
    
    return false;
  }
  
  /**
   * Verifica se o horário é incomum para um usuário
   * 
   * @param {string} userId ID do usuário
   * @param {Date|string|number} timestamp Timestamp do evento
   * @returns {boolean} Verdadeiro se horário é incomum
   */
  isUnusualTime(userId, timestamp) {
    if (!userId || !timestamp || !this.behaviorProfiles.has(userId)) {
      return false;
    }
    
    const profile = this.behaviorProfiles.get(userId);
    
    // Se não há padrões temporais conhecidos, qualquer novo é considerado incomum
    if (!profile.timePatterns || profile.timePatterns.length === 0) {
      return true;
    }
    
    // Normaliza o timestamp
    const eventTime = new Date(timestamp);
    const hour = eventTime.getHours();
    const dayOfWeek = eventTime.getDay();
    
    // Em produção: análise estatística detalhada, distribuição de probabilidade
    // Para esta demonstração, verificação simples de padrões
    
    // Verifica se o usuário costuma usar o sistema neste horário/dia da semana
    const similarTimePatterns = profile.timePatterns.filter(tp => {
      const hourDiff = Math.abs(tp.hour - hour);
      return tp.dayOfWeek === dayOfWeek && hourDiff <= 2; // +/- 2 horas
    });
    
    // Se não encontrou nenhum padrão similar, é incomum
    return similarTimePatterns.length === 0;
  }
  
  /**
   * Analisa o comportamento do usuário para detectar anomalias
   * 
   * @param {string} userId ID do usuário
   * @param {Object} context Contexto atual
   * @returns {number} Pontuação de anomalia (0-100)
   */
  analyzeBehavior(userId, context) {
    if (!userId || !this.behaviorProfiles.has(userId)) {
      return 50; // Pontuação média para comportamento desconhecido
    }
    
    // Em produção: modelo de machine learning, regras complexas, etc.
    // Para esta demonstração, implementação simplificada
    
    let anomalyScore = 0;
    
    // Verifica localização incomum
    if (context.geolocation && this.isUnusualLocation(userId, context.geolocation)) {
      anomalyScore += 30;
    }
    
    // Verifica horário incomum
    if (context.timestamp && this.isUnusualTime(userId, context.timestamp)) {
      anomalyScore += 20;
    }
    
    // Verifica dispositivo desconhecido
    if (context.deviceId && !this.isKnownDevice(userId, context.deviceId)) {
      anomalyScore += 25;
    }
    
    // Verifica padrões de interação (simulado)
    if (context.interactionPattern) {
      // Lógica para comparar com padrões normais
      // Simulação básica
      anomalyScore += 15;
    }
    
    return anomalyScore;
  }
  
  /**
   * Detecta viagens impossíveis (velocidade de movimento impossível)
   * 
   * @param {Object} location1 Localização anterior
   * @param {Object} location2 Localização atual
   * @param {Date|string|number} timestamp1 Timestamp anterior
   * @param {Date|string|number} timestamp2 Timestamp atual
   * @returns {boolean} Verdadeiro se detectou viagem impossível
   */
  detectImpossibleTravel(location1, location2, timestamp1, timestamp2) {
    if (!location1 || !location2 || !timestamp1 || !timestamp2) {
      return false;
    }
    
    // Calcula distância entre localizações
    const distance = this.calculateDistance(
      location1.latitude, location1.longitude,
      location2.latitude, location2.longitude
    );
    
    // Calcula tempo decorrido em horas
    const time1 = new Date(timestamp1).getTime();
    const time2 = new Date(timestamp2).getTime();
    const timeElapsed = (time2 - time1) / (1000 * 60 * 60); // horas
    
    // Se o tempo é zero ou negativo, algo está errado
    if (timeElapsed <= 0) {
      return false;
    }
    
    // Calcula velocidade em km/h
    const speed = distance / timeElapsed;
    
    // Verifica se a velocidade é humanamente impossível
    // Velocidade máxima de avião comercial ~900 km/h
    return speed > 1100; // Limiar um pouco acima de avião comercial
  }
  
  /**
   * Sanitiza um contexto para armazenamento seguro
   * 
   * @param {Object} context Contexto original
   * @returns {Object} Contexto sanitizado
   */
  sanitizeContext(context) {
    // Remove dados sensíveis ou PII
    const { password, pin, secret, credentials, ...sanitized } = context;
    
    // Em produção: mais sanitizações, mascaramento de PII, etc.
    
    return sanitized;
  }
  
  /**
   * Finaliza o monitor de risco
   * 
   * @returns {Promise<boolean>} Sucesso da finalização
   */
  async shutdown() {
    logger.info('[RiskMonitor] Finalizando monitor de risco');
    
    // Em produção: persistir estados, liberar recursos, etc.
    
    return true;
  }
}

module.exports = RiskMonitor;
