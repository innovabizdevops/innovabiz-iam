import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { EmailAdapter } from '../../../src/notifications/adapters/email-adapter';
import { NotificationChannel } from '../../../src/notifications/core/notification-channel';
import { NotificationRecipient } from '../../../src/notifications/adapters/notification-adapter';

// Mock dos módulos de email
jest.mock('nodemailer', () => ({
  createTransport: jest.fn().mockReturnValue({
    sendMail: jest.fn().mockImplementation((options) => {
      return Promise.resolve({
        messageId: 'test-message-id',
        envelope: { from: options.from, to: options.to },
        accepted: [options.to],
        rejected: [],
        pending: [],
        response: 'OK'
      });
    }),
    verify: jest.fn().mockResolvedValue(true)
  })
}));

jest.mock('@sendgrid/mail', () => ({
  setApiKey: jest.fn(),
  send: jest.fn().mockResolvedValue([
    {
      statusCode: 202,
      body: {},
      headers: {}
    }
  ])
}));

jest.mock('aws-sdk', () => ({
  SES: jest.fn().mockImplementation(() => ({
    sendEmail: jest.fn().mockReturnValue({
      promise: jest.fn().mockResolvedValue({
        MessageId: 'test-message-id'
      })
    })
  }))
}));

describe('EmailAdapter', () => {
  let adapter: EmailAdapter;
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
      defaultFrom: 'noreply@innovabiz.com',
      defaultReplyTo: 'support@innovabiz.com',
      rateLimits: {
        maxPerMinute: 100,
        maxPerHour: 1000
      }
    };
    
    // @ts-ignore - Ignorar erro de tipagem do mock para o teste
    adapter = new EmailAdapter(config, mockLogger);
  });
  
  afterEach(() => {
    jest.clearAllMocks();
  });
  
  describe('initialize', () => {
    it('deve inicializar corretamente com provedor SMTP', async () => {
      const result = await adapter.initialize();
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Adaptador de email inicializado'));
    });
    
    it('deve inicializar corretamente com provedor SendGrid', async () => {
      // @ts-ignore - Reconfigurar adapter para usar SendGrid
      adapter = new EmailAdapter({
        provider: 'sendgrid',
        sendgrid: {
          apiKey: 'SG.test-api-key'
        },
        defaultFrom: 'noreply@innovabiz.com'
      }, mockLogger);
      
      const result = await adapter.initialize();
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Adaptador de email inicializado'));
    });
    
    it('deve inicializar corretamente com provedor AWS SES', async () => {
      // @ts-ignore - Reconfigurar adapter para usar AWS SES
      adapter = new EmailAdapter({
        provider: 'aws-ses',
        awsSes: {
          region: 'us-east-1',
          credentials: {
            accessKeyId: 'test-access-key',
            secretAccessKey: 'test-secret-key'
          }
        },
        defaultFrom: 'noreply@innovabiz.com'
      }, mockLogger);
      
      const result = await adapter.initialize();
      
      expect(result).toBe(true);
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Adaptador de email inicializado'));
    });
    
    it('deve falhar ao inicializar com provedor desconhecido', async () => {
      // @ts-ignore - Reconfigurar adapter para usar provedor inválido
      adapter = new EmailAdapter({
        provider: 'invalid-provider',
        defaultFrom: 'noreply@innovabiz.com'
      }, mockLogger);
      
      await expect(adapter.initialize()).rejects.toThrow();
      expect(mockLogger.error).toHaveBeenCalled();
    });
  });
  
  describe('isHealthy', () => {
    it('deve verificar saúde do adaptador SMTP corretamente', async () => {
      await adapter.initialize();
      const result = await adapter.isHealthy();
      
      expect(result).toBe(true);
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('Verificação de saúde'));
    });
    
    it('deve lidar com falha na verificação de saúde', async () => {
      // Mock da função verify para falhar
      const nodemailer = require('nodemailer');
      nodemailer.createTransport.mockReturnValueOnce({
        sendMail: jest.fn(),
        verify: jest.fn().mockRejectedValue(new Error('Conexão falhou'))
      });
      
      await adapter.initialize();
      const result = await adapter.isHealthy();
      
      expect(result).toBe(false);
      expect(mockLogger.warn).toHaveBeenCalled();
    });
  });
  
  describe('send', () => {
    it('deve enviar email com sucesso', async () => {
      const recipient: NotificationRecipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      await adapter.initialize();
      const result = await adapter.send(recipient, content);
      
      expect(result.success).toBe(true);
      expect(result.channel).toBe(NotificationChannel.EMAIL);
      expect(result.recipientId).toBe('user123');
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Email enviado com sucesso'));
    });
    
    it('deve enviar email com HTML', async () => {
      const recipient: NotificationRecipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = '<h1>Bem-vindo</h1><p>Olá João, sua conta foi criada com sucesso!</p>';
      
      await adapter.initialize();
      const result = await adapter.send(recipient, content, {
        isHtml: true
      });
      
      expect(result.success).toBe(true);
      expect(result.channel).toBe(NotificationChannel.EMAIL);
    });
    
    it('deve falhar se destinatário não tiver email', async () => {
      const recipient: NotificationRecipient = {
        id: 'user123',
        metadata: {
          name: 'João Silva'
          // Sem email
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      await adapter.initialize();
      const result = await adapter.send(recipient, content);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.error).toHaveBeenCalled();
    });
    
    it('deve falhar com erro de envio', async () => {
      // Mock da função sendMail para falhar
      const nodemailer = require('nodemailer');
      nodemailer.createTransport.mockReturnValueOnce({
        sendMail: jest.fn().mockRejectedValue(new Error('Falha no envio')),
        verify: jest.fn().mockResolvedValue(true)
      });
      
      const recipient: NotificationRecipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      await adapter.initialize();
      const result = await adapter.send(recipient, content);
      
      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(mockLogger.error).toHaveBeenCalled();
    });
    
    it('deve enviar com opções personalizadas', async () => {
      const recipient: NotificationRecipient = {
        id: 'user123',
        metadata: {
          email: 'usuario@teste.com',
          name: 'João Silva'
        }
      };
      
      const content = 'Olá João, sua conta foi criada com sucesso!';
      
      await adapter.initialize();
      const result = await adapter.send(recipient, content, {
        subject: 'Bem-vindo à InnovaBiz',
        cc: ['gerente@innovabiz.com'],
        bcc: ['registro@innovabiz.com'],
        attachments: [
          {
            filename: 'welcome.pdf',
            content: Buffer.from('test content'),
            contentType: 'application/pdf'
          }
        ]
      });
      
      expect(result.success).toBe(true);
    });
  });
  
  describe('getCapabilities', () => {
    it('deve retornar capacidades corretas do adaptador', () => {
      const capabilities = adapter.getCapabilities();
      
      expect(capabilities.channel).toBe(NotificationChannel.EMAIL);
      expect(capabilities.supportsHtml).toBe(true);
      expect(capabilities.supportsAttachments).toBe(true);
      expect(capabilities.maxContentLength).toBeGreaterThan(0);
    });
  });
});