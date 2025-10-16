/**
 * Serviço de Restrições de Senha
 * 
 * Este módulo gerencia verificações de senhas comprometidas, listas
 * de bloqueio e integração com APIs de verificação externas.
 * 
 * @module constraint-service
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const crypto = require('crypto');
const fetch = require('node-fetch');
const { logger } = require('@innovabiz/auth-framework/utils/logging');
const DatabaseService = require('@innovabiz/auth-framework/services/database');
const RegionalAdapter = require('./regional-adapter');

/**
 * Serviço para gerenciamento de restrições de senha
 */
class ConstraintService {
  /**
   * Cria uma nova instância do serviço de restrições
   */
  constructor() {
    this.db = new DatabaseService();
    this.regionalAdapter = new RegionalAdapter();
    this.commonPasswordCache = new Set();
    this.haveibeenpwnedEnabled = true;
    this.blockedPatternsCache = [];
    
    // Carrega senhas comuns e padrões bloqueados
    this.loadCommonPasswords();
    this.loadBlockedPatterns();
    
    logger.info('[ConstraintService] Inicializado');
  }
  
  /**
   * Carrega senhas comuns para cache
   */
  async loadCommonPasswords() {
    try {
      // Em produção, carregaria de um arquivo ou banco de dados
      // Aqui estamos usando uma versão simplificada
      const query = `
        SELECT pattern 
        FROM iam.password_constraints 
        WHERE constraint_type = 'common_password' 
        AND status = 'active'
        LIMIT 1000
      `;
      
      const result = await this.db.query(query);
      
      result.rows.forEach(row => {
        this.commonPasswordCache.add(row.pattern.toLowerCase());
      });
      
      logger.info(`[ConstraintService] Carregadas ${this.commonPasswordCache.size} senhas comuns`);
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao carregar senhas comuns: ${error.message}`);
      
      // Fallback para algumas senhas comuns básicas
      const basicCommonPasswords = [
        'password', '123456', 'qwerty', 'admin', 'welcome',
        'senha', '123456789', 'abc123', 'password1', '1234567890',
        'letmein', 'passw0rd', 'trustno1', 'senha123', 'admin123'
      ];
      
      basicCommonPasswords.forEach(pwd => this.commonPasswordCache.add(pwd));
      logger.info(`[ConstraintService] Carregadas ${this.commonPasswordCache.size} senhas comuns (fallback)`);
    }
  }
  
  /**
   * Carrega padrões bloqueados para cache
   */
  async loadBlockedPatterns() {
    try {
      // Busca padrões bloqueados do banco de dados
      const query = `
        SELECT pattern, match_type 
        FROM iam.password_constraints 
        WHERE constraint_type = 'pattern' 
        AND status = 'active'
      `;
      
      const result = await this.db.query(query);
      
      this.blockedPatternsCache = result.rows.map(row => ({
        pattern: row.pattern,
        matchType: row.match_type,
        regex: this.createRegexFromPattern(row.pattern, row.match_type)
      }));
      
      logger.info(`[ConstraintService] Carregados ${this.blockedPatternsCache.length} padrões bloqueados`);
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao carregar padrões bloqueados: ${error.message}`);
      
      // Fallback para alguns padrões básicos
      this.blockedPatternsCache = [
        { 
          pattern: '(\\d)\\1{5,}', 
          matchType: 'regex',
          regex: /(\\d)\\1{5,}/
        },
        { 
          pattern: 'qwerty', 
          matchType: 'contains',
          regex: /qwerty/i
        },
        { 
          pattern: '12345', 
          matchType: 'contains',
          regex: /12345/
        }
      ];
      
      logger.info(`[ConstraintService] Carregados ${this.blockedPatternsCache.length} padrões bloqueados (fallback)`);
    }
  }
  
  /**
   * Cria uma expressão regular a partir de um padrão
   * 
   * @param {string} pattern Padrão a converter
   * @param {string} matchType Tipo de correspondência
   * @returns {RegExp} Expressão regular
   */
  createRegexFromPattern(pattern, matchType) {
    try {
      switch (matchType) {
        case 'regex':
          return new RegExp(pattern);
          
        case 'exact':
          return new RegExp(`^${this.escapeRegex(pattern)}$`, 'i');
          
        case 'contains':
          return new RegExp(this.escapeRegex(pattern), 'i');
          
        case 'starts_with':
          return new RegExp(`^${this.escapeRegex(pattern)}`, 'i');
          
        case 'ends_with':
          return new RegExp(`${this.escapeRegex(pattern)}$`, 'i');
          
        default:
          return new RegExp(this.escapeRegex(pattern), 'i');
      }
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao criar regex: ${error.message}`);
      return /invalid-pattern/;
    }
  }
  
  /**
   * Escapa caracteres especiais para expressão regular
   * 
   * @param {string} string String a escapar
   * @returns {string} String escapada
   */
  escapeRegex(string) {
    return string.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
  }
  
  /**
   * Verifica se uma senha atende a todas as restrições
   * 
   * @param {string} password Senha a verificar
   * @param {Object} userData Dados do usuário
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Resultado da verificação
   */
  async checkPasswordConstraints(password, userData = {}, context = {}) {
    const region = context.region || 'EU';
    const industry = context.industry || 'general';
    
    // Obtém configuração regional
    const regionConfig = this.regionalAdapter.getRegionConfig(region);
    
    // Prepara resultados
    const result = {
      valid: true,
      constraints: []
    };
    
    // Verifica se é uma senha comum
    if (regionConfig.passwordConstraints.checkCommonPasswords) {
      const isCommon = this.isCommonPassword(password);
      if (isCommon) {
        result.valid = false;
        result.constraints.push({
          type: 'common_password',
          message: 'A senha é muito comum e fácil de adivinhar'
        });
      }
    }
    
    // Verifica padrões bloqueados
    if (regionConfig.passwordConstraints.checkBlockedPatterns) {
      const blockedPattern = this.matchesBlockedPattern(password);
      if (blockedPattern) {
        result.valid = false;
        result.constraints.push({
          type: 'blocked_pattern',
          message: 'A senha contém um padrão não permitido',
          details: blockedPattern
        });
      }
    }
    
    // Verifica se contém informações pessoais
    if (regionConfig.passwordConstraints.checkPersonalInfo && userData) {
      const containsPersonalInfo = this.containsPersonalInfo(password, userData);
      if (containsPersonalInfo) {
        result.valid = false;
        result.constraints.push({
          type: 'contains_personal_info',
          message: 'A senha contém informações pessoais',
          details: containsPersonalInfo
        });
      }
    }
    
    // Verifica com serviço externo (como Have I Been Pwned)
    if (regionConfig.passwordConstraints.checkBreachDatabases) {
      try {
        const isCompromised = await this.isPasswordCompromised(password);
        if (isCompromised) {
          result.valid = false;
          result.constraints.push({
            type: 'compromised_password',
            message: 'Esta senha foi exposta em vazamentos de dados'
          });
        }
      } catch (error) {
        logger.error(`[ConstraintService] Erro na API de verificação: ${error.message}`);
        // Falha aberta - se a API falhar, continua com outras verificações
      }
    }
    
    // Considera restrições setoriais adicionais
    if (industry === 'finance' || industry === 'healthcare' || industry === 'government') {
      // Setores regulados têm requisitos adicionais
      const sectorConstraints = this.checkSectorSpecificConstraints(password, industry);
      
      if (!sectorConstraints.valid) {
        result.valid = false;
        result.constraints = [
          ...result.constraints,
          ...sectorConstraints.constraints
        ];
      }
    }
    
    return result;
  }
  
  /**
   * Verifica se uma senha é comum
   * 
   * @param {string} password Senha a verificar
   * @returns {boolean} Verdadeiro se for comum
   */
  isCommonPassword(password) {
    // Verifica cache em memória
    return this.commonPasswordCache.has(password.toLowerCase());
  }
  
  /**
   * Verifica se a senha corresponde a algum padrão bloqueado
   * 
   * @param {string} password Senha a verificar
   * @returns {Object|null} Padrão encontrado ou null
   */
  matchesBlockedPattern(password) {
    for (const pattern of this.blockedPatternsCache) {
      if (pattern.regex.test(password)) {
        return {
          pattern: pattern.pattern,
          matchType: pattern.matchType
        };
      }
    }
    
    return null;
  }
  
  /**
   * Verifica se a senha contém informações pessoais do usuário
   * 
   * @param {string} password Senha a verificar
   * @param {Object} userData Dados do usuário
   * @returns {Object|null} Informação encontrada ou null
   */
  containsPersonalInfo(password, userData) {
    const lowerPassword = password.toLowerCase();
    
    // Lista de dados pessoais a verificar
    const personalInfoFields = [
      { field: 'first_name', label: 'nome' },
      { field: 'last_name', label: 'sobrenome' },
      { field: 'username', label: 'nome de usuário' },
      { field: 'email', label: 'email', transform: value => value.split('@')[0] },
      { field: 'birth_date', label: 'data de nascimento', transform: value => {
        if (!value) return null;
        const date = new Date(value);
        if (isNaN(date.getTime())) return null;
        
        return [
          date.getFullYear().toString(),
          `${date.getDate()}${date.getMonth() + 1}${date.getFullYear() % 100}`,
          `${date.getDate()}${date.getMonth() + 1}${date.getFullYear()}`
        ];
      }},
      { field: 'phone', label: 'telefone', transform: value => value ? value.replace(/\D/g, '') : null }
    ];
    
    for (const info of personalInfoFields) {
      let value = userData[info.field];
      
      if (!value) continue;
      
      // Aplica transformação se configurada
      if (info.transform) {
        value = info.transform(value);
      }
      
      // Permite múltiplos valores (como para datas)
      const values = Array.isArray(value) ? value : [value];
      
      for (const val of values) {
        if (!val || val.length < 3) continue;
        
        if (lowerPassword.includes(val.toLowerCase())) {
          return {
            field: info.field,
            label: info.label
          };
        }
      }
    }
    
    return null;
  }
  
  /**
   * Verifica se uma senha foi comprometida em vazamentos
   * 
   * @param {string} password Senha a verificar
   * @returns {Promise<boolean>} Verdadeiro se comprometida
   */
  async isPasswordCompromised(password) {
    try {
      if (!this.haveibeenpwnedEnabled) {
        return false;
      }
      
      // Computa o hash SHA-1 da senha
      const sha1 = crypto.createHash('sha1').update(password).digest('hex').toUpperCase();
      const prefix = sha1.substring(0, 5);
      const suffix = sha1.substring(5);
      
      // Consulta a API k-Anonymity de Have I Been Pwned
      const response = await fetch(`https://api.pwnedpasswords.com/range/${prefix}`, {
        headers: {
          'User-Agent': 'INNOVABIZ-Security-Check/1.0'
        }
      });
      
      if (!response.ok) {
        throw new Error(`Erro na API: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.text();
      
      // Procura pelo sufixo do hash
      const lines = data.split('\n');
      for (const line of lines) {
        const parts = line.split(':');
        if (parts[0] === suffix) {
          const count = parseInt(parts[1].trim(), 10);
          return count > 0;
        }
      }
      
      return false;
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao verificar vazamentos: ${error.message}`);
      // Em caso de erro, assume que não está comprometida
      return false;
    }
  }
  
  /**
   * Verifica restrições específicas para setores regulados
   * 
   * @param {string} password Senha a verificar
   * @param {string} industry Setor
   * @returns {Object} Resultado da verificação
   */
  checkSectorSpecificConstraints(password, industry) {
    const result = {
      valid: true,
      constraints: []
    };
    
    switch (industry) {
      case 'finance':
        // Não pode conter apenas números consecutivos
        if (/^(?:(?:0(?=1)|1(?=2)|2(?=3)|3(?=4)|4(?=5)|5(?=6)|6(?=7)|7(?=8)|8(?=9)){5,}\d|\d{6,})$/.test(password)) {
          result.valid = false;
          result.constraints.push({
            type: 'finance_sequential',
            message: 'Senhas para o setor financeiro não podem conter apenas números sequenciais'
          });
        }
        
        // Não pode ter caracteres repetidos em sequência mais de duas vezes
        if (/(.)\1{2,}/.test(password)) {
          result.valid = false;
          result.constraints.push({
            type: 'finance_repetition',
            message: 'Senhas para o setor financeiro não podem ter caracteres repetidos em sequência'
          });
        }
        break;
        
      case 'healthcare':
        // Verifica presença de termos médicos comuns
        const medicalTerms = ['doctor', 'patient', 'nurse', 'medic', 'hospital', 'clinic', 
                             'doutor', 'paciente', 'enfermeira', 'medico', 'hospital', 'clinica'];
        
        for (const term of medicalTerms) {
          if (password.toLowerCase().includes(term)) {
            result.valid = false;
            result.constraints.push({
              type: 'healthcare_terms',
              message: 'Senhas para o setor de saúde não podem conter termos médicos comuns'
            });
            break;
          }
        }
        break;
        
      case 'government':
        // Restrições para senhas governamentais
        if (!/^(?=.*[A-Z])(?=.*[a-z])(?=.*\d)(?=.*[^A-Za-z\d]).{14,}$/.test(password)) {
          result.valid = false;
          result.constraints.push({
            type: 'government_complexity',
            message: 'Senhas para entidades governamentais precisam ter pelo menos 14 caracteres, incluindo maiúsculas, minúsculas, números e caracteres especiais'
          });
        }
        break;
    }
    
    return result;
  }
  
  /**
   * Adiciona uma senha à lista de senhas comuns
   * 
   * @param {string} password Senha a adicionar
   * @param {string} source Fonte da senha
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async addCommonPassword(password, source = 'manual') {
    try {
      // Adiciona ao banco de dados
      const query = `
        INSERT INTO iam.password_constraints (
          constraint_type, pattern, match_type, 
          description, source, status
        ) VALUES (
          'common_password', $1, 'exact',
          'Senha comum adicionada manualmente', $2, 'active'
        )
        ON CONFLICT (constraint_type, pattern) 
        DO NOTHING
        RETURNING constraint_id
      `;
      
      const result = await this.db.query(query, [
        password.toLowerCase(),
        source
      ]);
      
      // Adiciona ao cache
      this.commonPasswordCache.add(password.toLowerCase());
      
      return result.rows.length > 0;
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao adicionar senha comum: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Adiciona um padrão bloqueado
   * 
   * @param {string} pattern Padrão a adicionar
   * @param {string} matchType Tipo de correspondência
   * @param {string} description Descrição do padrão
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async addBlockedPattern(pattern, matchType = 'contains', description = '') {
    try {
      // Testa se o padrão é válido como regex
      if (matchType === 'regex') {
        try {
          new RegExp(pattern);
        } catch (e) {
          throw new Error(`Padrão de regex inválido: ${e.message}`);
        }
      }
      
      // Adiciona ao banco de dados
      const query = `
        INSERT INTO iam.password_constraints (
          constraint_type, pattern, match_type, 
          description, status
        ) VALUES (
          'pattern', $1, $2, $3, 'active'
        )
        ON CONFLICT (constraint_type, pattern) 
        DO UPDATE SET 
          match_type = $2,
          description = $3,
          status = 'active',
          updated_at = NOW()
        RETURNING constraint_id
      `;
      
      const result = await this.db.query(query, [
        pattern,
        matchType,
        description || `Padrão bloqueado (${matchType})`
      ]);
      
      // Atualiza o cache
      this.loadBlockedPatterns();
      
      return result.rows.length > 0;
    } catch (error) {
      logger.error(`[ConstraintService] Erro ao adicionar padrão bloqueado: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Importa uma lista de senhas comuns
   * 
   * @param {Array<string>} passwords Lista de senhas
   * @param {string} source Fonte das senhas
   * @returns {Promise<Object>} Resultado da importação
   */
  async importCommonPasswords(passwords, source = 'import') {
    try {
      let imported = 0;
      let skipped = 0;
      
      // Usa transação para melhor desempenho
      await this.db.beginTransaction();
      
      for (const password of passwords) {
        // Pula senhas muito curtas
        if (password.length < 4) {
          skipped++;
          continue;
        }
        
        // Insere no banco de dados
        const query = `
          INSERT INTO iam.password_constraints (
            constraint_type, pattern, match_type, 
            description, source, status
          ) VALUES (
            'common_password', $1, 'exact',
            'Senha comum importada', $2, 'active'
          )
          ON CONFLICT (constraint_type, pattern) 
          DO NOTHING
        `;
        
        try {
          await this.db.query(query, [
            password.toLowerCase(),
            source
          ]);
          
          // Adiciona ao cache
          this.commonPasswordCache.add(password.toLowerCase());
          imported++;
        } catch (err) {
          skipped++;
          continue;
        }
      }
      
      // Confirma a transação
      await this.db.commitTransaction();
      
      return {
        success: true,
        imported,
        skipped,
        total: passwords.length
      };
    } catch (error) {
      // Reverte a transação em caso de erro
      await this.db.rollbackTransaction();
      
      logger.error(`[ConstraintService] Erro ao importar senhas comuns: ${error.message}`);
      return {
        success: false,
        error: error.message,
        imported: 0,
        skipped: 0,
        total: passwords.length
      };
    }
  }
  
  /**
   * Configura a integração com Have I Been Pwned
   * 
   * @param {boolean} enabled Status da integração
   * @returns {Object} Configuração atual
   */
  configureHaveIBeenPwned(enabled) {
    this.haveibeenpwnedEnabled = enabled;
    
    logger.info(`[ConstraintService] Integração Have I Been Pwned ${enabled ? 'ativada' : 'desativada'}`);
    
    return {
      haveibeenpwnedEnabled: this.haveibeenpwnedEnabled
    };
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
          commonPasswordsCount: this.commonPasswordCache.size,
          blockedPatternsCount: this.blockedPatternsCache.length,
          haveibeenpwnedEnabled: this.haveibeenpwnedEnabled
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
}

module.exports = ConstraintService;
