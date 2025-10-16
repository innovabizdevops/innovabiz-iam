/**
 * Módulo de Logging do Framework de Autenticação
 * 
 * Implementa as funcionalidades de logging para o framework de autenticação,
 * com suporte a múltiplos níveis, formatação, e integração com sistemas de
 * observabilidade.
 * 
 * @module logging
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

// Dependências
const path = require('path');
const fs = require('fs');
const { format } = require('util');
const { createHash } = require('crypto');

// Constantes
const LOG_LEVELS = {
  trace: 0,
  debug: 1,
  info: 2,
  warn: 3,
  error: 4,
  fatal: 5,
  silent: 6
};

// Cores para o console (quando habilitado)
const COLORS = {
  reset: '\x1b[0m',
  dim: '\x1b[2m',
  black: '\x1b[30m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  brightRed: '\x1b[91m',
  brightGreen: '\x1b[92m',
  brightYellow: '\x1b[93m',
  brightBlue: '\x1b[94m',
  brightMagenta: '\x1b[95m',
  brightCyan: '\x1b[96m',
  brightWhite: '\x1b[97m'
};

// Mapeamento de cores por nível
const LEVEL_COLORS = {
  trace: COLORS.dim,
  debug: COLORS.cyan,
  info: COLORS.green,
  warn: COLORS.yellow,
  error: COLORS.red,
  fatal: COLORS.brightRed
};

/**
 * Classe de Logger
 */
class Logger {
  /**
   * Cria uma nova instância de Logger
   * 
   * @param {Object} options Opções de configuração
   */
  constructor(options = {}) {
    this.options = {
      level: options.level || 'info',
      enabled: options.enabled !== undefined ? options.enabled : true,
      colorize: options.colorize !== undefined ? options.colorize : true,
      timestamp: options.timestamp !== undefined ? options.timestamp : true,
      outputFile: options.outputFile || null,
      outputDir: options.outputDir || null,
      maxFiles: options.maxFiles || 10,
      maxFileSize: options.maxFileSize || 10 * 1024 * 1024, // 10MB
      serviceName: options.serviceName || 'auth-framework',
      maskSensitiveData: options.maskSensitiveData !== undefined ? options.maskSensitiveData : true,
      sensitiveFields: options.sensitiveFields || [
        'password', 'senha', 'token', 'secret', 'key', 'chave', 'pin', 
        'credential', 'credencial', 'accessToken', 'refreshToken', 'apiKey',
        'sessionKey', 'credit_card', 'cartao', 'cvv', 'credit_number'
      ],
      observabilityEnabled: options.observabilityEnabled !== undefined ? options.observabilityEnabled : false,
      observabilityEndpoint: options.observabilityEndpoint || null,
      ...options
    };
    
    this.currentLevel = LOG_LEVELS[this.options.level] || LOG_LEVELS.info;
    this.outputStream = null;
    this.outputSize = 0;
    this.fileCount = 0;
    
    // Inicializa arquivo de saída se configurado
    if (this.options.outputFile) {
      this.initOutputFile();
    }
    
    // Registra eventos de processo para limpeza adequada
    process.on('exit', () => this.cleanup());
    process.on('SIGINT', () => this.cleanup());
    process.on('SIGTERM', () => this.cleanup());
    process.on('uncaughtException', (err) => {
      this.error('Uncaught Exception:', err);
      this.cleanup();
      process.exit(1);
    });
  }
  
  /**
   * Inicializa o arquivo de saída
   */
  initOutputFile() {
    if (!this.options.outputFile) return;
    
    try {
      // Cria diretório se não existir
      if (this.options.outputDir) {
        if (!fs.existsSync(this.options.outputDir)) {
          fs.mkdirSync(this.options.outputDir, { recursive: true });
        }
        this.options.outputFile = path.join(this.options.outputDir, this.options.outputFile);
      }
      
      // Abre stream de escrita (ou cria arquivo)
      this.outputStream = fs.createWriteStream(this.options.outputFile, { flags: 'a' });
      
      // Obtém tamanho atual do arquivo
      try {
        const stats = fs.statSync(this.options.outputFile);
        this.outputSize = stats.size;
      } catch (err) {
        this.outputSize = 0;
      }
      
      // Conta arquivos de log existentes
      if (this.options.outputDir) {
        const files = fs.readdirSync(this.options.outputDir);
        this.fileCount = files.filter(f => f.startsWith(path.basename(this.options.outputFile))).length;
      }
      
    } catch (err) {
      console.error(`Error initializing log file: ${err.message}`);
      this.outputStream = null;
    }
  }
  
  /**
   * Define o nível de log
   * 
   * @param {string} level Nível de log
   */
  setLevel(level) {
    if (LOG_LEVELS[level] !== undefined) {
      this.currentLevel = LOG_LEVELS[level];
      this.options.level = level;
    }
  }
  
  /**
   * Limpa recursos
   */
  cleanup() {
    if (this.outputStream) {
      this.outputStream.end();
      this.outputStream = null;
    }
  }
  
  /**
   * Rotaciona o arquivo de log se necessário
   */
  rotateLogFileIfNeeded() {
    if (!this.outputStream || !this.options.outputFile) return;
    
    // Verifica se o arquivo atingiu o tamanho máximo
    if (this.outputSize >= this.options.maxFileSize) {
      // Fecha stream atual
      this.outputStream.end();
      
      // Define novo nome de arquivo com timestamp
      const extname = path.extname(this.options.outputFile);
      const basename = path.basename(this.options.outputFile, extname);
      const timestamp = new Date().toISOString().replace(/:/g, '-').replace(/\..+/, '');
      const rotatedFile = path.join(
        path.dirname(this.options.outputFile),
        `${basename}.${timestamp}${extname}`
      );
      
      // Renomeia arquivo atual
      try {
        fs.renameSync(this.options.outputFile, rotatedFile);
      } catch (err) {
        console.error(`Error rotating log file: ${err.message}`);
      }
      
      // Incrementa contador de arquivos
      this.fileCount++;
      
      // Remove arquivos antigos se exceder o máximo
      if (this.fileCount > this.options.maxFiles && this.options.outputDir) {
        try {
          const files = fs.readdirSync(this.options.outputDir)
            .filter(f => f.startsWith(basename) && f !== path.basename(this.options.outputFile))
            .map(f => ({
              name: f,
              time: fs.statSync(path.join(this.options.outputDir, f)).mtime.getTime()
            }))
            .sort((a, b) => a.time - b.time);
          
          // Remove os arquivos mais antigos
          const filesToRemove = files.slice(0, files.length - this.options.maxFiles + 1);
          for (const file of filesToRemove) {
            fs.unlinkSync(path.join(this.options.outputDir, file.name));
            this.fileCount--;
          }
        } catch (err) {
          console.error(`Error cleaning old log files: ${err.message}`);
        }
      }
      
      // Cria novo arquivo de log
      this.outputStream = fs.createWriteStream(this.options.outputFile, { flags: 'a' });
      this.outputSize = 0;
    }
  }
  
  /**
   * Formata uma mensagem de log
   * 
   * @param {string} level Nível do log
   * @param {Array} args Argumentos do log
   * @returns {string} Mensagem formatada
   */
  formatMessage(level, args) {
    // Formata argumentos
    let message = format(...args);
    
    // Mascara dados sensíveis se habilitado
    if (this.options.maskSensitiveData) {
      message = this.maskSensitiveData(message);
    }
    
    // Adiciona timestamp
    let formattedMessage = '';
    if (this.options.timestamp) {
      const timestamp = new Date().toISOString();
      formattedMessage += `[${timestamp}] `;
    }
    
    // Adiciona nível
    formattedMessage += `[${level.toUpperCase()}] `;
    
    // Adiciona nome do serviço
    if (this.options.serviceName) {
      formattedMessage += `[${this.options.serviceName}] `;
    }
    
    // Adiciona mensagem
    formattedMessage += message;
    
    return formattedMessage;
  }
  
  /**
   * Mascara dados sensíveis em uma string
   * 
   * @param {string} message Mensagem a mascarar
   * @returns {string} Mensagem com dados sensíveis mascarados
   */
  maskSensitiveData(message) {
    if (typeof message !== 'string') return message;
    
    let maskedMessage = message;
    
    // Expressão regular para detectar campos sensíveis em JSON e logs
    // Exemplo: "password": "secret123" ou password=secret123
    for (const field of this.options.sensitiveFields) {
      const patterns = [
        // JSON format: "field": "value" ou 'field': 'value'
        new RegExp(`["']${field}["']\\s*:\\s*["']([^"']+)["']`, 'gi'),
        // Form/query format: field=value
        new RegExp(`${field}=([^&\\s]+)`, 'gi'),
        // Log format: field: value
        new RegExp(`${field}:\\s+([^,\\s]+)`, 'gi')
      ];
      
      for (const pattern of patterns) {
        maskedMessage = maskedMessage.replace(pattern, (match, value) => {
          // Substitui o valor por uma versão resumida e mascarada
          if (value.length <= 4) {
            return match.replace(value, '****');
          } else {
            const firstTwo = value.substring(0, 2);
            const lastTwo = value.substring(value.length - 2);
            const hash = createHash('sha256').update(value).digest('hex').substring(0, 6);
            return match.replace(value, `${firstTwo}****${lastTwo}[${hash}]`);
          }
        });
      }
    }
    
    return maskedMessage;
  }
  
  /**
   * Envia um log para os destinos configurados
   * 
   * @param {string} level Nível do log
   * @param {Array} args Argumentos do log
   */
  log(level, ...args) {
    if (!this.options.enabled || LOG_LEVELS[level] < this.currentLevel) {
      return;
    }
    
    // Formata a mensagem
    const formattedMessage = this.formatMessage(level, args);
    
    // Escreve no console com cores se habilitado
    if (this.options.colorize && LEVEL_COLORS[level]) {
      console.log(`${LEVEL_COLORS[level]}${formattedMessage}${COLORS.reset}`);
    } else {
      console.log(formattedMessage);
    }
    
    // Escreve no arquivo se configurado
    if (this.outputStream) {
      this.rotateLogFileIfNeeded();
      this.outputStream.write(`${formattedMessage}\n`);
      this.outputSize += formattedMessage.length + 1; // +1 para o \n
    }
    
    // Envia para sistema de observabilidade se habilitado
    if (this.options.observabilityEnabled && this.options.observabilityEndpoint) {
      this.sendToObservability(level, formattedMessage);
    }
  }
  
  /**
   * Envia log para sistema de observabilidade
   * 
   * @param {string} level Nível do log
   * @param {string} message Mensagem formatada
   */
  sendToObservability(level, message) {
    // Implementação básica - em produção, usaria bibliotecas específicas
    try {
      // Aqui seria implementada a integração com sistemas como OpenTelemetry, 
      // Elastic APM, Datadog, New Relic, etc.
      // 
      // Este é apenas um esboço para demonstrar a integração
      
      if (!this.options.observabilityEndpoint) return;
      
      const logData = {
        timestamp: new Date().toISOString(),
        level,
        service: this.options.serviceName,
        message,
        metadata: {
          pid: process.pid,
          hostname: require('os').hostname(),
          application: 'InnovaBiz IAM',
          module: 'auth-framework'
        }
      };
      
      // Enviaria de forma assíncrona
      // Em produção: usar processos em background ou buffers para não bloquear
      setTimeout(() => {
        // Simulação de envio
        if (process.env.NODE_ENV !== 'production') {
          console.debug(`[Observability] Would send to ${this.options.observabilityEndpoint}: `, 
            JSON.stringify(logData).substring(0, 100) + '...');
        }
      }, 0);
    } catch (err) {
      // Falha silenciosamente para não impactar o aplicativo principal
      console.error(`Error sending to observability: ${err.message}`);
    }
  }
  
  /**
   * Log de nível trace
   */
  trace(...args) {
    this.log('trace', ...args);
  }
  
  /**
   * Log de nível debug
   */
  debug(...args) {
    this.log('debug', ...args);
  }
  
  /**
   * Log de nível info
   */
  info(...args) {
    this.log('info', ...args);
  }
  
  /**
   * Log de nível warn
   */
  warn(...args) {
    this.log('warn', ...args);
  }
  
  /**
   * Log de nível error
   */
  error(...args) {
    this.log('error', ...args);
  }
  
  /**
   * Log de nível fatal
   */
  fatal(...args) {
    this.log('fatal', ...args);
  }
}

// Exporta instância singleton
const logger = new Logger();

module.exports = { Logger, logger, LOG_LEVELS };
