import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { NotificationAdapterFactory } from '../../../src/notifications/adapters/notification-adapter-factory';
import { NotificationChannel } from '../../../src/notifications/core/notification-channel';

// Mock das classes de adaptadores
jest.mock('../../../src/notifications/adapters/email-adapter', () => {
  return {
    EmailAdapter: jest.fn().mockImplementation(() => ({
      initialize: jest.fn().mockResolvedValue(true),
      isHealthy: jest.fn().mockResolvedValue(true),
      send: jest.fn().mockResolvedValue({
        success: true,
        notificationId: 'test-id',
        channel: 'email',
        timestamp: new Date(),
        recipientId: 'test-recipient'
      }),
      getCapabilities: jest.fn().mockReturnValue({
        channel: 'email',
        supportsHtml: true,
        supportsAttachments: true,
        maxContentLength: 10000000
      })
    }))
  };
});

jest.mock('../../../src/notifications/adapters/sms-adapter', () => {
  return {
    SmsAdapter: jest.fn().mockImplementation(() => ({
      initialize: jest.fn().mockResolvedValue(true),
      isHealthy: jest.fn().mockResolvedValue(true),
      send: jest.fn().mockResolvedValue({
        success: true,
        notificationId: 'test-id',
        channel: 'sms',
        timestamp: new Date(),
        recipientId: 'test-recipient'
      }),
      getCapabilities: jest.fn().mockReturnValue({
        channel: 'sms',
        supportsHtml: false,
        supportsAttachments: false,
        maxContentLength: 160
      })
    }))
  };
});

jest.mock('../../../src/notifications/adapters/push-adapter', () => {
  return {
    PushAdapter: jest.fn().mockImplementation(() => ({
      initialize: jest.fn().mockResolvedValue(true),
      isHealthy: jest.fn().mockResolvedValue(true),
      send: jest.fn().mockResolvedValue({
        success: true,
        notificationId: 'test-id',
        channel: 'push',
        timestamp: new Date(),
        recipientId: 'test-recipient'
      }),
      getCapabilities: jest.fn().mockReturnValue({
        channel: 'push',
        supportsHtml: false,
        supportsAttachments: false,
        maxContentLength: 4000
      })
    }))
  };
});

describe('NotificationAdapterFactory', () => {
  let factory: NotificationAdapterFactory;
  let mockLogger: any;
  
  beforeEach(() => {
    // Mock do logger
    mockLogger = {
      debug: jest.fn(),
      info: jest.fn(),
      warn: jest.fn(),
      error: jest.fn()
    };
    
    // Configuração padrão para testes
    const config = {
      email: {
        provider: 'smtp',
        smtp: {
          host: 'smtp.teste.com',
          port: 587,
          secure: false,
          auth: {
            user: 'test@teste.com',
            pass: 'password123'
          }
        },
        defaultFrom: 'noreply@innovabiz.com'
      },
      sms: {
        provider: 'twilio',
        twilio: {
          accountSid: 'AC123',
          authToken: 'token123',
          from: '+12345678900'
        }
      },
      push: {
        provider: 'firebase',
        firebase: {
          serviceAccountKey: '/path/to/key.json',
          databaseURL: 'https://innovabiz-app.firebaseio.com'
        }
      },
      autoRecovery: {
        enabled: true,
        maxAttempts: 3,
        interval: 60000
      }
    };
    
    // @ts-ignore - Ignorar erro de tipagem do mock para o teste
    factory = new NotificationAdapterFactory(config, mockLogger);
  });
  
  afterEach(() => {
    jest.clearAllMocks();
  });
  
  describe('initialize', () => {
    it('deve inicializar todos os adaptadores configurados', async () => {
      const result = await factory.initialize();
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Fábrica de adaptadores inicializada'));
    });
    
    it('deve continuar mesmo se um adaptador falhar na inicialização', async () => {
      // Mock para simular falha no adaptador SMS
      require('../../../src/notifications/adapters/sms-adapter').SmsAdapter.mockImplementationOnce(() => ({
        initialize: jest.fn().mockRejectedValue(new Error('Falha na inicialização')),
        isHealthy: jest.fn().mockResolvedValue(false)
      }));
      
      const result = await factory.initialize();
      
      expect(result).toBe(true); // A fábrica ainda inicializa
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao inicializar adaptador'));
    });
  });
  
  describe('getAdapter', () => {
    it('deve retornar adaptador solicitado após inicialização', async () => {
      await factory.initialize();
      
      const emailAdapter = await factory.getAdapter(NotificationChannel.EMAIL);
      expect(emailAdapter).toBeDefined();
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('Adaptador email obtido'));
      
      const smsAdapter = await factory.getAdapter(NotificationChannel.SMS);
      expect(smsAdapter).toBeDefined();
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('Adaptador sms obtido'));
      
      const pushAdapter = await factory.getAdapter(NotificationChannel.PUSH);
      expect(pushAdapter).toBeDefined();
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('Adaptador push obtido'));
    });
    
    it('deve falhar se adaptador não estiver disponível', async () => {
      await factory.initialize();
      
      await expect(factory.getAdapter(NotificationChannel.WEBHOOK)).rejects.toThrow();
      expect(mockLogger.error).toHaveBeenCalled();
    });
    
    it('deve tentar inicializar adaptador sob demanda', async () => {
      // Sem inicialização prévia
      const emailAdapter = await factory.getAdapter(NotificationChannel.EMAIL);
      
      expect(emailAdapter).toBeDefined();
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Inicializando adaptador sob demanda'));
    });
  });
  
  describe('getHealthStatus', () => {
    it('deve retornar status de saúde de todos os adaptadores', async () => {
      await factory.initialize();
      
      const healthStatus = await factory.getHealthStatus();
      
      expect(healthStatus).toBeDefined();
      expect(healthStatus[NotificationChannel.EMAIL]).toBe(true);
      expect(healthStatus[NotificationChannel.SMS]).toBe(true);
      expect(healthStatus[NotificationChannel.PUSH]).toBe(true);
    });
    
    it('deve refletir quando um adaptador não está saudável', async () => {
      // Mock para simular adaptador não saudável
      require('../../../src/notifications/adapters/email-adapter').EmailAdapter.mockImplementationOnce(() => ({
        initialize: jest.fn().mockResolvedValue(true),
        isHealthy: jest.fn().mockResolvedValue(false),
        getCapabilities: jest.fn().mockReturnValue({
          channel: 'email'
        })
      }));
      
      await factory.initialize();
      
      const healthStatus = await factory.getHealthStatus();
      
      expect(healthStatus[NotificationChannel.EMAIL]).toBe(false);
      expect(mockLogger.warn).toHaveBeenCalledWith(expect.stringContaining('Adaptador email não está saudável'));
    });
  });
  
  describe('restartAdapter', () => {
    it('deve reiniciar um adaptador específico', async () => {
      await factory.initialize();
      
      const result = await factory.restartAdapter(NotificationChannel.EMAIL);
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Adaptador email reiniciado com sucesso'));
    });
    
    it('deve lidar com falha na reinicialização', async () => {
      // Mock para simular falha na inicialização
      require('../../../src/notifications/adapters/email-adapter').EmailAdapter.mockImplementationOnce(() => ({
        initialize: jest.fn().mockRejectedValue(new Error('Falha na reinicialização'))
      }));
      
      await factory.initialize();
      
      const result = await factory.restartAdapter(NotificationChannel.EMAIL);
      
      expect(result).toBe(false);
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao reiniciar adaptador'));
    });
  });
  
  describe('getCapabilities', () => {
    it('deve retornar capacidades de um adaptador específico', async () => {
      await factory.initialize();
      
      const capabilities = await factory.getCapabilities(NotificationChannel.EMAIL);
      
      expect(capabilities).toBeDefined();
      expect(capabilities.channel).toBe('email');
      expect(capabilities.supportsHtml).toBe(true);
    });
    
    it('deve falhar se adaptador não estiver disponível', async () => {
      await factory.initialize();
      
      await expect(factory.getCapabilities(NotificationChannel.WEBHOOK)).rejects.toThrow();
    });
  });
  
  describe('getSupportedChannels', () => {
    it('deve retornar todos os canais suportados', async () => {
      await factory.initialize();
      
      const channels = factory.getSupportedChannels();
      
      expect(channels).toContain(NotificationChannel.EMAIL);
      expect(channels).toContain(NotificationChannel.SMS);
      expect(channels).toContain(NotificationChannel.PUSH);
      expect(channels).not.toContain(NotificationChannel.WEBHOOK);
    });
  });
});