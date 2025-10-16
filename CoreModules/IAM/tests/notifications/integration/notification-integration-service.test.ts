import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { NotificationIntegrationService } from '../../../src/notifications/integration/notification-integration-service';
import { NotificationChannel } from '../../../src/notifications/core/notification-channel';
import { BaseEvent } from '../../../src/notifications/core/base-event';

// Mock do serviço de notificação
const mockNotificationService = {
  initialize: jest.fn().mockResolvedValue(true),
  send: jest.fn().mockResolvedValue({
    success: true,
    notificationId: 'test-notif-id',
    channel: NotificationChannel.EMAIL,
    timestamp: new Date(),
    recipientId: 'test-recipient'
  }),
  sendWithTemplate: jest.fn().mockResolvedValue({
    success: true,
    notificationId: 'test-notif-id',
    channel: NotificationChannel.EMAIL,
    timestamp: new Date(),
    recipientId: 'test-recipient'
  })
};

// Mock do serviço de templates
const mockTemplateService = {
  render: jest.fn().mockImplementation((templateId, options) => {
    return Promise.resolve({
      content: `Conteúdo renderizado para ${templateId}`,
      metadata: {
        subject: 'Assunto do Template',
        templateId,
        locale: options.locale || 'pt-BR'
      }
    });
  }),
  getTemplate: jest.fn().mockImplementation((templateId) => {
    return Promise.resolve({
      id: templateId,
      name: `Template ${templateId}`,
      category: 'test',
      version: 1,
      active: true,
      content: {
        'pt-BR': {
          subject: 'Assunto de Teste',
          body: 'Conteúdo de teste {{variable}}'
        }
      }
    });
  })
};

describe('NotificationIntegrationService', () => {
  let service: NotificationIntegrationService;
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
      defaultOptions: {
        retryStrategy: {
          maxAttempts: 3,
          initialDelayMs: 1000,
          maxDelayMs: 30000
        },
        tracking: {
          enabled: true
        }
      },
      modules: {
        iam: {
          templateMappings: {
            'authentication': {
              'login-success': 'iam-login-success',
              'login-failure': 'iam-login-failure',
              'password-reset': 'iam-password-reset'
            },
            'registration': {
              'user-created': 'iam-welcome',
              'email-verification': 'iam-email-verification'
            }
          }
        },
        'payment-gateway': {
          templateMappings: {
            'payment': {
              'payment-success': 'pg-payment-success',
              'payment-failure': 'pg-payment-failure',
              'payment-refund': 'pg-payment-refund'
            }
          }
        }
      }
    };
    
    // @ts-ignore - Ignorar erro de tipagem do mock para o teste
    service = new NotificationIntegrationService(
      mockNotificationService,
      mockTemplateService,
      config,
      mockLogger
    );
  });
  
  afterEach(() => {
    jest.clearAllMocks();
  });
  
  describe('initialize', () => {
    it('deve inicializar serviço de integração corretamente', async () => {
      const result = await service.initialize();
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Serviço de integração de notificações inicializado'));
    });
    
    it('deve registrar processadores padrão para módulos', async () => {
      await service.initialize();
      
      // @ts-ignore - Verificar registro interno de processadores
      const processors = service.eventProcessors;
      expect(processors).toBeDefined();
      expect(processors.iam).toBeDefined();
      expect(processors['payment-gateway']).toBeDefined();
    });
  });
  
  describe('registerEventProcessor', () => {
    it('deve registrar processador de evento personalizado', async () => {
      const customProcessor = jest.fn().mockResolvedValue({
        success: true
      });
      
      service.registerEventProcessor('custom-module', 'custom-category', customProcessor);
      
      // @ts-ignore - Verificar registro interno de processadores
      const processors = service.eventProcessors;
      expect(processors['custom-module']).toBeDefined();
      expect(processors['custom-module']['custom-category']).toBe(customProcessor);
      
      expect(mockLogger.info).toHaveBeenCalledWith(
        expect.stringContaining('Processador de eventos registrado')
      );
    });
    
    it('deve substituir processador existente', async () => {
      const customProcessor1 = jest.fn();
      const customProcessor2 = jest.fn();
      
      service.registerEventProcessor('module', 'category', customProcessor1);
      service.registerEventProcessor('module', 'category', customProcessor2);
      
      // @ts-ignore - Verificar registro interno de processadores
      const processors = service.eventProcessors;
      expect(processors['module']['category']).toBe(customProcessor2);
      
      expect(mockLogger.warn).toHaveBeenCalledWith(
        expect.stringContaining('Processador de eventos substituído')
      );
    });
  });
  
  describe('processEvent', () => {
    it('deve processar evento IAM corretamente', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {
          userId: 'user123',
          userName: 'João Silva',
          userEmail: 'joao@example.com',
          ipAddress: '192.168.1.1',
          device: 'Mobile - Android'
        },
        context: {
          tenantId: 'tenant1',
          environment: 'production',
          locale: 'pt-BR'
        },
        recipients: [{
          id: 'user123',
          metadata: {
            email: 'joao@example.com',
            name: 'João Silva'
          }
        }]
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(true);
      expect(mockTemplateService.render).toHaveBeenCalledWith('iam-login-success', expect.anything());
      expect(mockNotificationService.sendWithTemplate).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'user123' }),
        'iam-login-success',
        expect.anything()
      );
    });
    
    it('deve processar evento Payment Gateway corretamente', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-789012',
        timestamp: new Date(),
        module: 'payment-gateway',
        category: 'payment',
        type: 'payment-success',
        data: {
          transactionId: 'tx-456',
          amount: 150.75,
          currency: 'BRL',
          paymentMethod: 'credit-card',
          userId: 'user456',
          userEmail: 'maria@example.com',
          userName: 'Maria Santos'
        },
        context: {
          tenantId: 'tenant1',
          environment: 'production',
          locale: 'pt-BR'
        }
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(true);
      expect(mockTemplateService.render).toHaveBeenCalledWith('pg-payment-success', expect.anything());
      expect(mockNotificationService.sendWithTemplate).toHaveBeenCalled();
    });
    
    it('deve extrair destinatário do evento se não fornecido', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {
          userId: 'user123',
          userName: 'João Silva',
          userEmail: 'joao@example.com'
        },
        context: {
          tenantId: 'tenant1'
        }
        // Sem recipients explícitos
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(true);
      expect(mockNotificationService.sendWithTemplate).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'user123',
          metadata: expect.objectContaining({
            email: 'joao@example.com',
            name: 'João Silva'
          })
        }),
        expect.anything(),
        expect.anything()
      );
    });
    
    it('deve falhar se não houver mapeamento para o evento', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'unknown-module',
        category: 'unknown-category',
        type: 'unknown-event',
        data: {},
        context: {}
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.warn).toHaveBeenCalledWith(expect.stringContaining('Nenhum processador encontrado'));
      expect(mockNotificationService.sendWithTemplate).not.toHaveBeenCalled();
    });
    
    it('deve falhar se não conseguir extrair destinatário', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {
          // Sem informações de usuário
        },
        context: {}
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Não foi possível extrair destinatário'));
      expect(mockNotificationService.sendWithTemplate).not.toHaveBeenCalled();
    });
    
    it('deve processar evento com processador personalizado', async () => {
      const customProcessor = jest.fn().mockResolvedValue({
        success: true,
        notificationId: 'custom-notif-id'
      });
      
      service.registerEventProcessor('custom-module', 'custom-category', customProcessor);
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'custom-module',
        category: 'custom-category',
        type: 'custom-event',
        data: {
          customField: 'customValue'
        },
        context: {}
      };
      
      const result = await service.processEvent(event);
      
      expect(result.success).toBe(true);
      expect(customProcessor).toHaveBeenCalledWith(event);
      expect(mockNotificationService.sendWithTemplate).not.toHaveBeenCalled(); // Processador custom assumiu controle
    });
    
    it('deve processar evento de forma assíncrona quando configurado', async () => {
      await service.initialize();
      
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {
          userId: 'user123',
          userName: 'João Silva',
          userEmail: 'joao@example.com'
        },
        context: {}
      };
      
      // @ts-ignore - Mock de setTimeout para testar execução assíncrona
      jest.spyOn(global, 'setTimeout').mockImplementation((callback) => {
        callback();
        return {} as any;
      });
      
      const result = await service.processEvent(event, { async: true });
      
      expect(result.success).toBe(true);
      expect(result.async).toBe(true);
      expect(global.setTimeout).toHaveBeenCalled();
      // Verificamos que eventualmente o sendWithTemplate é chamado
      expect(mockNotificationService.sendWithTemplate).toHaveBeenCalled();
    });
  });
  
  describe('extractRecipient', () => {
    it('deve extrair destinatário corretamente de dados do evento', () => {
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {
          userId: 'user123',
          userName: 'João Silva',
          userEmail: 'joao@example.com',
          phone: '+5511999999999'
        },
        context: {}
      };
      
      // @ts-ignore - Testando método privado
      const recipient = service.extractRecipient(event);
      
      expect(recipient).toBeDefined();
      expect(recipient.id).toBe('user123');
      expect(recipient.metadata).toBeDefined();
      expect(recipient.metadata?.name).toBe('João Silva');
      expect(recipient.metadata?.email).toBe('joao@example.com');
      expect(recipient.metadata?.phone).toBe('+5511999999999');
    });
    
    it('deve extrair destinatário corretamente de recipients do evento', () => {
      const event: BaseEvent = {
        id: 'evt-123456',
        timestamp: new Date(),
        module: 'iam',
        category: 'authentication',
        type: 'login-success',
        data: {},
        recipients: [{
          id: 'user123',
          metadata: {
            email: 'joao@example.com',
            name: 'João Silva',
            phone: '+5511999999999'
          }
        }],
        context: {}
      };
      
      // @ts-ignore - Testando método privado
      const recipient = service.extractRecipient(event);
      
      expect(recipient).toBeDefined();
      expect(recipient.id).toBe('user123');
      expect(recipient.metadata).toBeDefined();
      expect(recipient.metadata?.name).toBe('João Silva');
      expect(recipient.metadata?.email).toBe('joao@example.com');
      expect(recipient.metadata?.phone).toBe('+5511999999999');
    });
  });
});