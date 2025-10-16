import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { NotificationService } from '../../../src/notifications/services/notification-service';
import { NotificationChannel } from '../../../src/notifications/core/notification-channel';

// Mock da fábrica de adaptadores
const mockAdapterFactory = {
  initialize: jest.fn().mockResolvedValue(true),
  getAdapter: jest.fn(),
  getHealthStatus: jest.fn().mockResolvedValue({
    [NotificationChannel.EMAIL]: true,
    [NotificationChannel.SMS]: true,
    [NotificationChannel.PUSH]: true
  }),
  getSupportedChannels: jest.fn().mockReturnValue([
    NotificationChannel.EMAIL,
    NotificationChannel.SMS,
    NotificationChannel.PUSH
  ]),
  getCapabilities: jest.fn().mockImplementation((channel) => {
    if (channel === NotificationChannel.EMAIL) {
      return Promise.resolve({
        channel: NotificationChannel.EMAIL,
        supportsHtml: true,
        supportsAttachments: true,
        maxContentLength: 10000000
      });
    } else if (channel === NotificationChannel.SMS) {
      return Promise.resolve({
        channel: NotificationChannel.SMS,
        supportsHtml: false,
        supportsAttachments: false,
        maxContentLength: 160
      });
    } else if (channel === NotificationChannel.PUSH) {
      return Promise.resolve({
        channel: NotificationChannel.PUSH,
        supportsHtml: false,
        supportsAttachments: false,
        maxContentLength: 4000
      });
    }
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

// Mock do serviço de rastreamento
const mockTrackingService = {
  trackSend: jest.fn().mockResolvedValue(undefined),
  trackDelivery: jest.fn().mockResolvedValue(undefined),
  getTrackingSummary: jest.fn().mockResolvedValue({
    notificationId: 'test-notif-id',
    recipientId: 'test-recipient',
    status: 'SENT',
    events: [
      {
        eventType: 'SEND',
        channel: NotificationChannel.EMAIL,
        timestamp: new Date()
      }
    ]
  }),
  generateTrackingPixel: jest.fn().mockReturnValue('https://track.example.com/p/pixel.gif'),
  generateTrackingUrl: jest.fn().mockImplementation((url) => `https://track.example.com/t/redirect?url=${encodeURIComponent(url)}`)
};

// Mock dos adaptadores
const mockEmailAdapter = {
  initialize: jest.fn().mockResolvedValue(true),
  isHealthy: jest.fn().mockResolvedValue(true),
  send: jest.fn().mockResolvedValue({
    success: true,
    notificationId: 'test-email-id',
    channel: NotificationChannel.EMAIL,
    timestamp: new Date(),
    recipientId: 'test-recipient'
  }),
  getCapabilities: jest.fn().mockReturnValue({
    channel: NotificationChannel.EMAIL,
    supportsHtml: true,
    supportsAttachments: true,
    maxContentLength: 10000000
  })
};

const mockSmsAdapter = {
  initialize: jest.fn().mockResolvedValue(true),
  isHealthy: jest.fn().mockResolvedValue(true),
  send: jest.fn().mockResolvedValue({
    success: true,
    notificationId: 'test-sms-id',
    channel: NotificationChannel.SMS,
    timestamp: new Date(),
    recipientId: 'test-recipient'
  }),
  getCapabilities: jest.fn().mockReturnValue({
    channel: NotificationChannel.SMS,
    supportsHtml: false,
    supportsAttachments: false,
    maxContentLength: 160
  })
};

// Configuração do mock getAdapter
mockAdapterFactory.getAdapter.mockImplementation((channel: NotificationChannel) => {
  if (channel === NotificationChannel.EMAIL) {
    return Promise.resolve(mockEmailAdapter);
  } else if (channel === NotificationChannel.SMS) {
    return Promise.resolve(mockSmsAdapter);
  } else {
    return Promise.reject(new Error(`Adaptador não disponível: ${channel}`));
  }
});

describe('NotificationService', () => {
  let service: NotificationService;
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
      defaultChannel: NotificationChannel.EMAIL,
      fallbackChannel: NotificationChannel.SMS,
      retry: {
        enabled: true,
        maxAttempts: 3,
        initialDelayMs: 1000,
        backoffMultiplier: 2
      },
      tracking: {
        enabled: true
      }
    };
    
    // @ts-ignore - Ignorar erro de tipagem do mock para o teste
    service = new NotificationService(
      mockAdapterFactory, 
      mockTemplateService, 
      mockTrackingService,
      config, 
      mockLogger
    );
  });
  
  afterEach(() => {
    jest.clearAllMocks();
  });
  
  describe('initialize', () => {
    it('deve inicializar corretamente', async () => {
      const result = await service.initialize();
      
      expect(result).toBe(true);
      expect(mockAdapterFactory.initialize).toHaveBeenCalled();
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Serviço de notificação inicializado'));
    });
    
    it('deve lidar com erro na inicialização dos adaptadores', async () => {
      mockAdapterFactory.initialize.mockRejectedValueOnce(new Error('Erro na inicialização'));
      
      await expect(service.initialize()).rejects.toThrow();
      expect(mockLogger.error).toHaveBeenCalled();
    });
  });
  
  describe('send', () => {
    beforeEach(async () => {
      await service.initialize();
    });
    
    it('deve enviar notificação com sucesso via canal padrão', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      const result = await service.send(recipient, content);
      
      expect(result.success).toBe(true);
      expect(result.channel).toBe(NotificationChannel.EMAIL);
      expect(mockEmailAdapter.send).toHaveBeenCalled();
      expect(mockTrackingService.trackSend).toHaveBeenCalled();
    });
    
    it('deve respeitar canal preferido do usuário', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          phone: '+5511999999999',
          name: 'João Silva',
          preferredChannels: [NotificationChannel.SMS]
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      const result = await service.send(recipient, content);
      
      expect(result.success).toBe(true);
      expect(result.channel).toBe(NotificationChannel.SMS);
      expect(mockSmsAdapter.send).toHaveBeenCalled();
      expect(mockEmailAdapter.send).not.toHaveBeenCalled();
    });
    
    it('deve usar fallback quando canal preferido falhar', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          phone: '+5511999999999',
          name: 'João Silva',
          preferredChannels: [NotificationChannel.SMS]
        }
      };
      
      // Simular falha no SMS
      mockSmsAdapter.send.mockRejectedValueOnce(new Error('Falha no envio de SMS'));
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      const result = await service.send(recipient, content);
      
      expect(result.success).toBe(true);
      expect(result.channel).toBe(NotificationChannel.EMAIL);
      expect(mockSmsAdapter.send).toHaveBeenCalled();
      expect(mockEmailAdapter.send).toHaveBeenCalled();
      expect(mockLogger.warn).toHaveBeenCalledWith(expect.stringContaining('Usando canal de fallback'));
    });
    
    it('deve enviar usando template', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const result = await service.sendWithTemplate(recipient, 'welcome-email', {
        variables: {
          name: 'João Silva',
          activationLink: 'https://exemplo.com/activate'
        }
      });
      
      expect(result.success).toBe(true);
      expect(mockTemplateService.render).toHaveBeenCalledWith('welcome-email', expect.anything());
      expect(mockEmailAdapter.send).toHaveBeenCalled();
    });
    
    it('deve falhar se não houver canal disponível para o destinatário', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          name: 'João Silva'
          // Sem email nem telefone
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      const result = await service.send(recipient, content);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Não foi possível determinar um canal'));
    });
    
    it('deve enviar para múltiplos canais simultaneamente quando configurado', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          phone: '+5511999999999',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      const result = await service.send(recipient, content, {
        simultaneousChannels: true,
        preferredChannels: [NotificationChannel.EMAIL, NotificationChannel.SMS]
      });
      
      expect(result.success).toBe(true);
      expect(result.multiChannelResults).toBeDefined();
      expect(result.multiChannelResults?.length).toBe(2);
      expect(mockEmailAdapter.send).toHaveBeenCalled();
      expect(mockSmsAdapter.send).toHaveBeenCalled();
    });
  });
  
  describe('sendBatch', () => {
    beforeEach(async () => {
      await service.initialize();
    });
    
    it('deve enviar notificações em lote com sucesso', async () => {
      const recipients = [
        {
          id: 'user1',
          metadata: {
            email: 'user1@teste.com',
            name: 'Usuário 1'
          }
        },
        {
          id: 'user2',
          metadata: {
            email: 'user2@teste.com',
            name: 'Usuário 2'
          }
        }
      ];
      
      const content = 'Mensagem em lote para todos os usuários';
      
      const result = await service.sendBatch(recipients, content);
      
      expect(result.success).toBe(true);
      expect(result.successCount).toBe(2);
      expect(result.failureCount).toBe(0);
      expect(mockEmailAdapter.send).toHaveBeenCalledTimes(2);
    });
    
    it('deve reportar falhas parciais em envio em lote', async () => {
      const recipients = [
        {
          id: 'user1',
          metadata: {
            email: 'user1@teste.com',
            name: 'Usuário 1'
          }
        },
        {
          id: 'user2',
          metadata: {
            email: 'user2@teste.com',
            name: 'Usuário 2'
          }
        },
        {
          id: 'user3',
          metadata: {
            name: 'Usuário 3'
            // Sem email
          }
        }
      ];
      
      const content = 'Mensagem em lote para todos os usuários';
      
      const result = await service.sendBatch(recipients, content);
      
      expect(result.success).toBe(true); // Sucesso parcial ainda é considerado sucesso
      expect(result.successCount).toBe(2);
      expect(result.failureCount).toBe(1);
      expect(mockEmailAdapter.send).toHaveBeenCalledTimes(2);
      expect(mockLogger.warn).toHaveBeenCalledWith(expect.stringContaining('falhas no envio em lote'));
    });
  });
  
  describe('scheduleNotification', () => {
    beforeEach(async () => {
      await service.initialize();
    });
    
    it('deve agendar notificação com sucesso', async () => {
      const recipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, lembrete sobre sua reunião!';
      const scheduledTime = new Date(Date.now() + 3600000); // 1 hora no futuro
      
      // Mock da função interna de agendamento
      // @ts-ignore - Acessando propriedade privada para teste
      service.scheduleQueue = {
        add: jest.fn().mockResolvedValue({ id: 'job-123' })
      };
      
      const result = await service.scheduleNotification(recipient, content, scheduledTime);
      
      expect(result.success).toBe(true);
      expect(result.scheduledId).toBe('job-123');
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Notificação agendada'));
    });
  });
  
  describe('cancelNotification', () => {
    beforeEach(async () => {
      await service.initialize();
    });
    
    it('deve cancelar notificação agendada com sucesso', async () => {
      // Mock da função interna de agendamento
      // @ts-ignore - Acessando propriedade privada para teste
      service.scheduleQueue = {
        getJob: jest.fn().mockResolvedValue({ remove: jest.fn().mockResolvedValue(true) })
      };
      
      const result = await service.cancelNotification('job-123');
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Notificação cancelada'));
    });
    
    it('deve lidar com notificação não encontrada', async () => {
      // Mock da função interna de agendamento
      // @ts-ignore - Acessando propriedade privada para teste
      service.scheduleQueue = {
        getJob: jest.fn().mockResolvedValue(null)
      };
      
      const result = await service.cancelNotification('job-not-found');
      
      expect(result).toBe(false);
      expect(mockLogger.warn).toHaveBeenCalledWith(expect.stringContaining('não encontrada'));
    });
  });
  
  describe('getNotificationStatus', () => {
    beforeEach(async () => {
      await service.initialize();
    });
    
    it('deve retornar status da notificação', async () => {
      const notificationId = 'test-notif-id';
      
      const status = await service.getNotificationStatus(notificationId);
      
      expect(status).toBeDefined();
      expect(status.notificationId).toBe(notificationId);
      expect(status.status).toBe('SENT');
      expect(mockTrackingService.getTrackingSummary).toHaveBeenCalledWith(notificationId);
    });
  });
});