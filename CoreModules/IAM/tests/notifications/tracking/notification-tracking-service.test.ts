import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { NotificationTrackingService } from '../../../src/notifications/tracking/notification-tracking-service';
import { NotificationChannel } from '../../../src/notifications/core/notification-channel';
import { NotificationStatus, NotificationTrackingEventType } from '../../../src/notifications/tracking/notification-tracking-models';

// Mock do repositório de rastreamento
const mockTrackingRepository = {
  saveTrackingEvent: jest.fn().mockResolvedValue(true),
  getTrackingEvents: jest.fn().mockResolvedValue([
    {
      id: 'evt-1',
      notificationId: 'notif-123',
      recipientId: 'user-123',
      channel: NotificationChannel.EMAIL,
      eventType: NotificationTrackingEventType.SEND,
      timestamp: new Date(),
      metadata: {}
    },
    {
      id: 'evt-2',
      notificationId: 'notif-123',
      recipientId: 'user-123',
      channel: NotificationChannel.EMAIL,
      eventType: NotificationTrackingEventType.DELIVERY,
      timestamp: new Date(Date.now() + 1000),
      metadata: {}
    }
  ]),
  getTrackingSummary: jest.fn().mockResolvedValue({
    notificationId: 'notif-123',
    recipientId: 'user-123',
    status: NotificationStatus.DELIVERED,
    events: [
      {
        eventType: NotificationTrackingEventType.SEND,
        channel: NotificationChannel.EMAIL,
        timestamp: new Date()
      },
      {
        eventType: NotificationTrackingEventType.DELIVERY,
        channel: NotificationChannel.EMAIL,
        timestamp: new Date(Date.now() + 1000)
      }
    ]
  }),
  updateNotificationStatus: jest.fn().mockResolvedValue(true),
  getAggregateStats: jest.fn().mockResolvedValue({
    totalSent: 100,
    totalDelivered: 95,
    totalOpened: 75,
    totalClicked: 50,
    totalFailed: 5,
    totalBounced: 2,
    deliveryRate: 0.95,
    openRate: 0.75,
    clickRate: 0.5,
    failureRate: 0.05,
    bounceRate: 0.02
  }),
  purgeOldEvents: jest.fn().mockResolvedValue(10)
};

describe('NotificationTrackingService', () => {
  let service: NotificationTrackingService;
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
      enabled: true,
      trackingDomain: 'track.innovabiz.com',
      pixelTrackingEnabled: true,
      linkTrackingEnabled: true,
      enrichTrackingData: true,
      dataRetentionDays: 90,
      cacheEnabled: true,
      cacheTTLSeconds: 300
    };
    
    // @ts-ignore - Ignorar erro de tipagem do mock para o teste
    service = new NotificationTrackingService(mockTrackingRepository, config, mockLogger);
  });
  
  afterEach(() => {
    jest.clearAllMocks();
  });
  
  describe('trackSend', () => {
    it('deve registrar evento de envio com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const provider = 'sendgrid';
      
      await service.trackSend(notificationId, recipientId, channel, provider);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.SEND,
          metadata: expect.objectContaining({ provider })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.SENT
      );
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Evento de envio registrado'));
    });
    
    it('deve lidar com erro no registro de envio', async () => {
      mockTrackingRepository.saveTrackingEvent.mockRejectedValueOnce(new Error('Erro de banco de dados'));
      
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      
      await expect(service.trackSend(notificationId, recipientId, channel)).resolves.not.toThrow();
      
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao registrar evento de envio'));
    });
    
    it('não deve fazer nada se o rastreamento estiver desabilitado', async () => {
      // @ts-ignore - Reconfigurar serviço com rastreamento desabilitado
      service = new NotificationTrackingService(mockTrackingRepository, { enabled: false }, mockLogger);
      
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      
      await service.trackSend(notificationId, recipientId, channel);
      
      expect(mockTrackingRepository.saveTrackingEvent).not.toHaveBeenCalled();
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('Rastreamento desabilitado'));
    });
  });
  
  describe('trackDelivery', () => {
    it('deve registrar evento de entrega com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const deliveryData = {
        providerReference: 'ref-789',
        timestamp: new Date()
      };
      
      await service.trackDelivery(notificationId, recipientId, channel, deliveryData);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.DELIVERY,
          metadata: expect.objectContaining({
            providerReference: 'ref-789'
          })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.DELIVERED
      );
    });
  });
  
  describe('trackOpen', () => {
    it('deve registrar evento de abertura com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const openData = {
        userAgent: 'Mozilla/5.0...',
        ipAddress: '192.168.1.1',
        timestamp: new Date()
      };
      
      await service.trackOpen(notificationId, recipientId, channel, openData);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.OPEN,
          metadata: expect.objectContaining({
            userAgent: 'Mozilla/5.0...',
            ipAddress: '192.168.1.1'
          })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.OPENED
      );
    });
  });
  
  describe('trackClick', () => {
    it('deve registrar evento de clique com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const clickData = {
        url: 'https://example.com/offer',
        linkId: 'link-1',
        userAgent: 'Mozilla/5.0...',
        ipAddress: '192.168.1.1',
        timestamp: new Date()
      };
      
      await service.trackClick(notificationId, recipientId, channel, clickData);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.CLICK,
          metadata: expect.objectContaining({
            url: 'https://example.com/offer',
            linkId: 'link-1'
          })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.CLICKED
      );
    });
  });
  
  describe('trackFailure', () => {
    it('deve registrar evento de falha com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const failureData = {
        errorCode: 'INVALID_RECIPIENT',
        errorMessage: 'Endereço de email inválido',
        attempts: 3
      };
      
      await service.trackFailure(notificationId, recipientId, channel, failureData);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.FAILURE,
          metadata: expect.objectContaining({
            errorCode: 'INVALID_RECIPIENT',
            attempts: 3
          })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.FAILED
      );
    });
  });
  
  describe('trackBounce', () => {
    it('deve registrar evento de bounce com sucesso', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const bounceData = {
        bounceType: 'hard',
        bounceCategory: 'suppress_bounce',
        reason: 'Recipient address rejected: Domain not found'
      };
      
      await service.trackBounce(notificationId, recipientId, channel, bounceData);
      
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          notificationId,
          recipientId,
          channel,
          eventType: NotificationTrackingEventType.BOUNCE,
          metadata: expect.objectContaining({
            bounceType: 'hard',
            bounceCategory: 'suppress_bounce'
          })
        })
      );
      expect(mockTrackingRepository.updateNotificationStatus).toHaveBeenCalledWith(
        notificationId,
        recipientId,
        NotificationStatus.BOUNCED
      );
    });
  });
  
  describe('processProviderWebhook', () => {
    it('deve processar webhook do SendGrid com sucesso', async () => {
      const provider = 'sendgrid';
      const payload = [
        {
          event: 'delivered',
          sg_message_id: 'sg-123',
          sg_event_id: 'sg-evt-456',
          email: 'recipient@example.com',
          timestamp: 1597845301
        },
        {
          event: 'open',
          sg_message_id: 'sg-123',
          sg_event_id: 'sg-evt-457',
          email: 'recipient@example.com',
          timestamp: 1597845401,
          useragent: 'Mozilla/5.0...',
          ip: '192.168.1.1'
        }
      ];
      
      // Mock para extrair informações do ID da mensagem
      // @ts-ignore - Acessando método privado para teste
      jest.spyOn(service, 'extractIdsFromProviderReference').mockReturnValue({
        notificationId: 'notif-123',
        recipientId: 'user-123'
      });
      
      const result = await service.processProviderWebhook(provider, payload);
      
      expect(result.success).toBe(true);
      expect(result.eventsProcessed).toBe(2);
      expect(mockTrackingRepository.saveTrackingEvent).toHaveBeenCalledTimes(2);
    });
    
    it('deve lidar com erro no processamento de webhook', async () => {
      const provider = 'unknown';
      const payload = { test: 'data' };
      
      const result = await service.processProviderWebhook(provider, payload);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.error).toHaveBeenCalled();
    });
  });
  
  describe('getTrackingEvents', () => {
    it('deve obter eventos de rastreamento de uma notificação', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      
      const events = await service.getTrackingEvents(notificationId, recipientId);
      
      expect(events).toHaveLength(2);
      expect(events[0].eventType).toBe(NotificationTrackingEventType.SEND);
      expect(events[1].eventType).toBe(NotificationTrackingEventType.DELIVERY);
      expect(mockTrackingRepository.getTrackingEvents).toHaveBeenCalledWith(notificationId, recipientId);
    });
  });
  
  describe('getTrackingSummary', () => {
    it('deve obter resumo de rastreamento de uma notificação', async () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      
      const summary = await service.getTrackingSummary(notificationId, recipientId);
      
      expect(summary.status).toBe(NotificationStatus.DELIVERED);
      expect(summary.events).toHaveLength(2);
      expect(mockTrackingRepository.getTrackingSummary).toHaveBeenCalledWith(notificationId, recipientId);
    });
  });
  
  describe('getAggregateStats', () => {
    it('deve obter estatísticas agregadas com filtros', async () => {
      const filter = {
        startDate: new Date('2023-01-01'),
        endDate: new Date('2023-01-31'),
        channel: NotificationChannel.EMAIL,
        module: 'iam'
      };
      
      const stats = await service.getAggregateStats(filter);
      
      expect(stats.totalSent).toBe(100);
      expect(stats.deliveryRate).toBe(0.95);
      expect(mockTrackingRepository.getAggregateStats).toHaveBeenCalledWith(filter);
    });
  });
  
  describe('generateTrackingPixel', () => {
    it('deve gerar URL de pixel de rastreamento', () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      
      const pixelUrl = service.generateTrackingPixel(notificationId, recipientId, channel);
      
      expect(pixelUrl).toContain('track.innovabiz.com');
      expect(pixelUrl).toContain(notificationId);
      expect(pixelUrl).toContain(recipientId);
      expect(pixelUrl).toContain(channel);
    });
    
    it('deve retornar undefined se rastreamento de pixel estiver desabilitado', () => {
      // @ts-ignore - Reconfigurar serviço com rastreamento de pixel desabilitado
      service = new NotificationTrackingService(mockTrackingRepository, {
        enabled: true,
        pixelTrackingEnabled: false
      }, mockLogger);
      
      const pixelUrl = service.generateTrackingPixel('notif-123', 'user-123', NotificationChannel.EMAIL);
      
      expect(pixelUrl).toBeUndefined();
    });
  });
  
  describe('generateTrackingUrl', () => {
    it('deve gerar URL de rastreamento para links', () => {
      const notificationId = 'notif-123';
      const recipientId = 'user-123';
      const channel = NotificationChannel.EMAIL;
      const originalUrl = 'https://example.com/offer';
      const linkId = 'link-1';
      
      const trackingUrl = service.generateTrackingUrl(
        originalUrl,
        notificationId,
        recipientId,
        channel,
        linkId
      );
      
      expect(trackingUrl).toContain('track.innovabiz.com');
      expect(trackingUrl).toContain(notificationId);
      expect(trackingUrl).toContain(recipientId);
      expect(trackingUrl).toContain(channel);
      expect(trackingUrl).toContain(linkId);
      expect(trackingUrl).toContain(encodeURIComponent(originalUrl));
    });
    
    it('deve retornar URL original se rastreamento de link estiver desabilitado', () => {
      // @ts-ignore - Reconfigurar serviço com rastreamento de link desabilitado
      service = new NotificationTrackingService(mockTrackingRepository, {
        enabled: true,
        linkTrackingEnabled: false
      }, mockLogger);
      
      const originalUrl = 'https://example.com/offer';
      const trackingUrl = service.generateTrackingUrl(
        originalUrl,
        'notif-123',
        'user-123',
        NotificationChannel.EMAIL
      );
      
      expect(trackingUrl).toBe(originalUrl);
    });
  });
  
  describe('purgeOldEvents', () => {
    it('deve purgar eventos antigos com base na configuração de retenção', async () => {
      const result = await service.purgeOldEvents();
      
      expect(result).toBe(10); // Número de eventos purgados
      expect(mockTrackingRepository.purgeOldEvents).toHaveBeenCalledWith(90); // configuração de 90 dias
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Eventos antigos purgados'));
    });
  });
});