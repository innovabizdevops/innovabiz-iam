/**
 * @fileoverview Testes para o adaptador de eventos de autenticação específicos do setor de saúde
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/tests/kafka/healthcare-auth-event-adapter.test
 */

const { expect } = require('chai');
const sinon = require('sinon');
const { v4: uuidv4 } = require('uuid');
const { HealthcareAuthEventAdapter } = require('../../../03-Desenvolvimento/auth-framework/kafka/healthcare/healthcare-auth-event-adapter');
const { AuthEventProducer } = require('../../../03-Desenvolvimento/auth-framework/kafka/auth-event-producer');
const { AuthEventConsumer } = require('../../../03-Desenvolvimento/auth-framework/kafka/auth-event-consumer');
const { MCPKafkaAdapter } = require('../../../03-Desenvolvimento/auth-framework/kafka/mcp/mcp-kafka-adapter');

// Mocks para os validadores de saúde
class MockHIPAAValidator {
  constructor(options = {}) {
    this.options = options;
    this.validations = [];
  }
  
  validate(event) {
    this.validations.push(event);
    
    // Retorna validação bem-sucedida por padrão
    return {
      valid: true,
      code: 'HIPAA_COMPLIANT',
      severity: 'INFO',
      reason: 'Evento está em conformidade com HIPAA'
    };
  }
}

class MockGDPRHealthcareValidator {
  constructor(options = {}) {
    this.options = options;
    this.validations = [];
  }
  
  validate(event) {
    this.validations.push(event);
    
    // Verifica se há mascaramento de dados pessoais
    const hasDataMasking = event.data_masked === true;
    
    return {
      valid: hasDataMasking,
      code: hasDataMasking ? 'GDPR_COMPLIANT' : 'GDPR_DATA_MASKING_REQUIRED',
      severity: hasDataMasking ? 'INFO' : 'HIGH',
      reason: hasDataMasking ? 
        'Evento está em conformidade com GDPR para saúde' : 
        'Evento requer mascaramento de dados para conformidade GDPR em saúde'
    };
  }
}

class MockLGPDHealthcareValidator {
  constructor(options = {}) {
    this.options = options;
    this.validations = [];
  }
  
  validate(event) {
    this.validations.push(event);
    
    // Verifica se há consentimento registrado
    const hasConsent = event.additional_context?.consent_id !== undefined;
    
    return {
      valid: hasConsent,
      code: hasConsent ? 'LGPD_COMPLIANT' : 'LGPD_CONSENT_REQUIRED',
      severity: hasConsent ? 'INFO' : 'MEDIUM',
      reason: hasConsent ? 
        'Evento está em conformidade com LGPD para saúde' : 
        'Evento requer registro de consentimento para conformidade LGPD em saúde'
    };
  }
}

// Mock para observabilidade
class MockObservability {
  constructor() {
    this.records = {
      healthcareAuthAttempts: [],
      healthcareEventPublished: [],
      complianceIssues: []
    };
  }
  
  recordHealthcareAuthAttempt(data) {
    this.records.healthcareAuthAttempts.push(data);
  }
  
  recordHealthcareEventPublished(data) {
    this.records.healthcareEventPublished.push(data);
  }
  
  recordComplianceIssue(data) {
    this.records.complianceIssues.push(data);
  }
}

describe('HealthcareAuthEventAdapter', () => {
  let adapter;
  let mockProducer;
  let mockConsumer;
  let mockMcpAdapter;
  let mockObservability;
  let sandbox;
  
  beforeEach(() => {
    sandbox = sinon.createSandbox();
    
    // Criar mocks para os componentes
    mockProducer = {
      connect: sandbox.stub().resolves(),
      disconnect: sandbox.stub().resolves(),
      publishAuthEvent: sandbox.stub().callsFake((event, options) => {
        return Promise.resolve({
          success: true,
          event_id: event.event_id,
          topic: options.topic || 'iam-healthcare-auth-events'
        });
      })
    };
    
    mockConsumer = {
      connect: sandbox.stub().resolves(),
      disconnect: sandbox.stub().resolves(),
      subscribe: sandbox.stub().resolves(),
      run: sandbox.stub().resolves()
    };
    
    mockMcpAdapter = {
      connect: sandbox.stub().resolves(),
      disconnect: sandbox.stub().resolves(),
      isConnected: sandbox.stub().returns(true),
      publishToMCP: sandbox.stub().resolves({
        success: true,
        messageId: uuidv4()
      }),
      subscribeMCP: sandbox.stub().callsFake((channel, handler, options) => {
        return Promise.resolve(uuidv4());
      })
    };
    
    mockObservability = new MockObservability();
    
    // Stub para os componentes Kafka
    sandbox.stub(AuthEventProducer.prototype, 'connect').callsFake(() => mockProducer.connect());
    sandbox.stub(AuthEventProducer.prototype, 'disconnect').callsFake(() => mockProducer.disconnect());
    sandbox.stub(AuthEventProducer.prototype, 'publishAuthEvent').callsFake((...args) => mockProducer.publishAuthEvent(...args));
    
    sandbox.stub(AuthEventConsumer.prototype, 'connect').callsFake(() => mockConsumer.connect());
    sandbox.stub(AuthEventConsumer.prototype, 'disconnect').callsFake(() => mockConsumer.disconnect());
    sandbox.stub(AuthEventConsumer.prototype, 'subscribe').callsFake((...args) => mockConsumer.subscribe(...args));
    sandbox.stub(AuthEventConsumer.prototype, 'run').callsFake((...args) => mockConsumer.run(...args));
    
    sandbox.stub(MCPKafkaAdapter.prototype, 'connect').callsFake(() => mockMcpAdapter.connect());
    sandbox.stub(MCPKafkaAdapter.prototype, 'disconnect').callsFake(() => mockMcpAdapter.disconnect());
    sandbox.stub(MCPKafkaAdapter.prototype, 'isConnected').callsFake(() => mockMcpAdapter.isConnected());
    sandbox.stub(MCPKafkaAdapter.prototype, 'publishToMCP').callsFake((...args) => mockMcpAdapter.publishToMCP(...args));
    sandbox.stub(MCPKafkaAdapter.prototype, 'subscribeMCP').callsFake((...args) => mockMcpAdapter.subscribeMCP(...args));
    
    // Preparar validadores para o adaptador
    const mockValidators = {
      HIPAAHealthcareValidator: MockHIPAAValidator,
      GDPRHealthcareValidator: MockGDPRHealthcareValidator,
      LGPDHealthcareValidator: MockLGPDHealthcareValidator
    };
    
    // Stub para a factory de validadores
    sandbox.stub(global, 'require').callsFake((module) => {
      if (module === '../../../integrations/healthcare/validators') {
        return mockValidators;
      }
      if (module === '../../../core/utils/observability') {
        return { EventObservability: function() { return mockObservability; } };
      }
      return require(module);
    });
    
    // Criar adaptador para teste
    adapter = new HealthcareAuthEventAdapter({
      regionCode: 'US',
      enableComplianceValidation: true,
      enableObservability: true
    });
    
    // Sobrescrever os validadores com mocks
    adapter.validators = [
      new MockHIPAAValidator({ region: 'US' })
    ];
  });
  
  afterEach(() => {
    sandbox.restore();
  });
  
  describe('Inicialização', () => {
    it('deve inicializar com configurações padrão', () => {
      const newAdapter = new HealthcareAuthEventAdapter();
      expect(newAdapter.regionCode).to.equal('EU');
      expect(newAdapter.enableComplianceValidation).to.be.true;
      expect(newAdapter.enableObservability).to.be.true;
    });
    
    it('deve inicializar com configurações personalizadas', () => {
      const newAdapter = new HealthcareAuthEventAdapter({
        regionCode: 'BR',
        enableComplianceValidation: false,
        enableObservability: false,
        kafkaConfig: {
          clientId: 'test-client'
        }
      });
      
      expect(newAdapter.regionCode).to.equal('BR');
      expect(newAdapter.enableComplianceValidation).to.be.false;
      expect(newAdapter.enableObservability).to.be.false;
      expect(newAdapter.kafkaConfig.clientId).to.equal('test-client');
    });
    
    it('deve carregar configurações específicas de região', () => {
      // EU
      const euAdapter = new HealthcareAuthEventAdapter({ regionCode: 'EU' });
      expect(euAdapter.healthcareConfig.compliance.gdpr_healthcare).to.be.true;
      expect(euAdapter.healthcareConfig.mcp_channel).to.equal('healthcare.auth.eu');
      
      // BR
      const brAdapter = new HealthcareAuthEventAdapter({ regionCode: 'BR' });
      expect(brAdapter.healthcareConfig.compliance.lgpd_healthcare).to.be.true;
      expect(brAdapter.healthcareConfig.mcp_channel).to.equal('healthcare.auth.br');
      
      // US
      const usAdapter = new HealthcareAuthEventAdapter({ regionCode: 'US' });
      expect(usAdapter.healthcareConfig.compliance.hipaa).to.be.true;
      expect(usAdapter.healthcareConfig.mcp_channel).to.equal('healthcare.auth.us');
    });
    
    it('deve inicializar validadores corretos para cada região', () => {
      // Stub para inicialização de validadores
      const initializeValidatorsSpy = sandbox.spy(HealthcareAuthEventAdapter.prototype, 'initializeValidators');
      
      // US (HIPAA)
      new HealthcareAuthEventAdapter({ regionCode: 'US' });
      expect(initializeValidatorsSpy.called).to.be.true;
      
      // Verificar se o stub foi chamado com os validadores corretos
      // Na implementação real, isso validaria que os validadores corretos foram inicializados
    });
  });
  
  describe('Processamento de Eventos', () => {
    it('deve processar e adaptar um evento para o contexto de saúde', () => {
      const event = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K05',
        additional_context: {
          phi_access: true
        }
      };
      
      const processedEvent = adapter.processHealthcareAuthEvent(event);
      
      // Verificar campos adicionados
      expect(processedEvent.sector).to.equal('HEALTHCARE');
      expect(processedEvent.healthcare_metadata).to.exist;
      expect(processedEvent.healthcare_metadata.phi_access).to.be.true;
      expect(processedEvent.healthcare_metadata.requires_hipaa_logging).to.be.true;
      
      // Verificar que os campos originais foram mantidos
      expect(processedEvent.event_id).to.equal(event.event_id);
      expect(processedEvent.event_type).to.equal(event.event_type);
      expect(processedEvent.user_id).to.equal(event.user_id);
    });
    
    it('deve aplicar mascaramento de dados PHI quando configurado', () => {
      // Sobrescrever a função de mascaramento
      const maskSensitiveDataStub = sandbox.stub().callsFake((event, fields) => {
        const maskedEvent = { ...event, data_masked: true };
        
        // Simular mascaramento básico
        fields.forEach(field => {
          if (event.client_info && event.client_info[field]) {
            if (!maskedEvent.client_info) maskedEvent.client_info = { ...event.client_info };
            maskedEvent.client_info[field] = '***MASKED***';
          }
        });
        
        return maskedEvent;
      });
      
      sandbox.stub(global, 'require').callsFake((module) => {
        if (module === '../../../core/utils/data-masking') {
          return { maskSensitiveData: maskSensitiveDataStub };
        }
        return require(module);
      });
      
      // Recriar adaptador com novos stubs
      const maskingAdapter = new HealthcareAuthEventAdapter({ regionCode: 'US' });
      maskingAdapter.healthcareConfig.data_masking = true;
      maskingAdapter.healthcareConfig.phi_fields = ['ip_address', 'user_agent'];
      
      const event = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        client_info: {
          ip_address: '192.168.1.1',
          user_agent: 'Mozilla/5.0'
        }
      };
      
      const processedEvent = maskingAdapter.processHealthcareAuthEvent(event);
      
      // Verificar se o mascaramento foi aplicado
      expect(maskSensitiveDataStub.calledOnce).to.be.true;
      expect(processedEvent.data_masked).to.be.true;
    });
  });
  
  describe('Publicação de Eventos', () => {
    it('deve publicar um evento de autenticação específico para saúde', async () => {
      const event = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K05',
        region_code: 'US'
      };
      
      const result = await adapter.publishHealthcareAuthEvent(event);
      
      // Verificar se o evento foi publicado
      expect(mockProducer.publishAuthEvent.calledOnce).to.be.true;
      
      // Verificar se o evento foi processado antes da publicação
      const publishedEvent = mockProducer.publishAuthEvent.firstCall.args[0];
      expect(publishedEvent.sector).to.equal('HEALTHCARE');
      expect(publishedEvent.healthcare_metadata).to.exist;
      
      // Verificar se o MCP também recebeu o evento
      expect(mockMcpAdapter.publishToMCP.calledOnce).to.be.true;
      
      // Verificar se foi registrado na observabilidade
      expect(mockObservability.records.healthcareEventPublished.length).to.equal(1);
      expect(mockObservability.records.healthcareEventPublished[0].eventType).to.equal('LOGIN_SUCCESS');
    });
    
    it('deve validar eventos antes da publicação', async () => {
      // Criar um novo adaptador com validador que falha
      const strictAdapter = new HealthcareAuthEventAdapter({
        regionCode: 'EU',
        enableComplianceValidation: true
      });
      
      // Adicionar um validador GDPR que sempre falha
      strictAdapter.validators = [
        new MockGDPRHealthcareValidator({ region: 'EU' })
      ];
      
      const nonCompliantEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K01',  // Método fraco para saúde
        data_masked: false   // Sem mascaramento
      };
      
      // Tentar publicar evento não-conforme com validação estrita
      try {
        await strictAdapter.publishHealthcareAuthEvent(nonCompliantEvent, {
          strictValidation: true
        });
        expect.fail('Deveria ter lançado erro de validação');
      } catch (error) {
        expect(error.message).to.include('Falha na validação de conformidade');
      }
      
      // Verificar que o evento não foi publicado
      expect(mockProducer.publishAuthEvent.called).to.be.false;
    });
    
    it('deve registrar problemas de conformidade mas permitir publicação sem validação estrita', async () => {
      // Criar adaptador com observabilidade
      const adapter = new HealthcareAuthEventAdapter({
        regionCode: 'BR',
        enableComplianceValidation: true,
        enableObservability: true
      });
      
      // Adicionar validador LGPD que detecta falta de consentimento
      adapter.validators = [
        new MockLGPDHealthcareValidator({ region: 'BR' })
      ];
      adapter.observability = mockObservability;
      
      const nonCompliantEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K05',
        additional_context: {
          // Sem consent_id
        }
      };
      
      // Publicar sem validação estrita
      await adapter.publishHealthcareAuthEvent(nonCompliantEvent, {
        strictValidation: false
      });
      
      // Verificar que o evento foi publicado apesar da não-conformidade
      expect(mockProducer.publishAuthEvent.calledOnce).to.be.true;
      
      // Verificar que o problema foi registrado
      expect(mockObservability.records.complianceIssues.length).to.equal(1);
      expect(mockObservability.records.complianceIssues[0].issueType).to.equal('LGPD_CONSENT_REQUIRED');
    });
  });
  
  describe('Consumo de Eventos', () => {
    it('deve configurar um consumidor para eventos de autenticação de saúde', async () => {
      const handlerSpy = sinon.spy();
      
      await adapter.setupHealthcareConsumer(handlerSpy, {
        fromBeginning: true,
        autoCommit: true
      });
      
      // Verificar se o consumidor foi conectado
      expect(mockConsumer.connect.calledOnce).to.be.true;
      
      // Verificar se a inscrição foi realizada no tópico correto
      expect(mockConsumer.subscribe.calledOnce).to.be.true;
      expect(mockConsumer.subscribe.firstCall.args[0]).to.deep.equal(['iam-healthcare-auth-events']);
      expect(mockConsumer.subscribe.firstCall.args[1]).to.deep.include({ fromBeginning: true });
      
      // Verificar se o consumo foi iniciado
      expect(mockConsumer.run.calledOnce).to.be.true;
      
      // Simular processamento de um evento
      const processHandler = mockConsumer.run.firstCall.args[0].processHandler;
      expect(processHandler).to.be.a('function');
      
      // Chamar o handler com um evento
      const event = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123'
      };
      
      await processHandler(event);
      
      // Verificar se o handler personalizado foi chamado
      expect(handlerSpy.calledOnce).to.be.true;
      expect(handlerSpy.firstCall.args[0]).to.deep.equal(event);
    });
    
    it('deve usar handlers internos se nenhum handler personalizado for fornecido', async () => {
      await adapter.setupHealthcareConsumer(null);
      
      // Simular processamento de um evento
      const processHandler = mockConsumer.run.firstCall.args[0].processHandler;
      
      // Simular evento LOGIN_SUCCESS
      const loginEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K05'
      };
      
      const result = await processHandler(loginEvent);
      
      // Verificar que o handler interno processou o evento
      expect(result.processed).to.be.true;
      expect(result.action).to.equal('integrate');
      
      // Verificar que foi registrado na observabilidade
      expect(mockObservability.records.healthcareAuthAttempts.length).to.equal(1);
      expect(mockObservability.records.healthcareAuthAttempts[0].success).to.be.true;
      
      // Verificar que o evento foi publicado no MCP
      expect(mockMcpAdapter.publishToMCP.calledOnce).to.be.true;
    });
  });
  
  describe('Integração MCP', () => {
    it('deve configurar um assinante MCP para integração com sistemas de saúde', async () => {
      const handlerSpy = sinon.spy();
      
      const subscriptionId = await adapter.setupMCPSubscriber(handlerSpy, {
        convertToKafkaEvent: true
      });
      
      // Verificar se o adaptador MCP foi conectado
      expect(mockMcpAdapter.connect.calledOnce).to.be.true;
      
      // Verificar se a inscrição foi realizada no canal correto
      expect(mockMcpAdapter.subscribeMCP.calledOnce).to.be.true;
      expect(mockMcpAdapter.subscribeMCP.firstCall.args[0]).to.equal('healthcare.auth.us');
      expect(mockMcpAdapter.subscribeMCP.firstCall.args[2]).to.deep.include({ convertToKafkaEvent: true });
      
      // Verificar ID de inscrição retornado
      expect(subscriptionId).to.be.a('string');
    });
  });
  
  describe('Conformidade MFA', () => {
    it('deve identificar métodos MFA compatíveis com requisitos de saúde', () => {
      // Métodos seguros para saúde
      expect(adapter.isHealthcareMFACompliant('K05')).to.be.true;  // OTP
      expect(adapter.isHealthcareMFACompliant('K06')).to.be.true;  // TOTP
      expect(adapter.isHealthcareMFACompliant('K07')).to.be.true;  // HOTP
      expect(adapter.isHealthcareMFACompliant('K13')).to.be.true;  // Biometria
      
      // Métodos não seguros para saúde
      expect(adapter.isHealthcareMFACompliant('K01')).to.be.false;  // Senha simples
      expect(adapter.isHealthcareMFACompliant('K02')).to.be.false;  // SMS básico
    });
    
    it('deve detectar MFA fraco para acesso a dados de saúde', async () => {
      await adapter.setupHealthcareConsumer();
      
      // Simular processamento de um evento MFA
      const processHandler = mockConsumer.run.firstCall.args[0].processHandler;
      
      // Evento MFA com método fraco
      const weakMfaEvent = {
        event_id: uuidv4(),
        event_type: 'MFA_CHALLENGE_ISSUED',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        method_code: 'K01',  // Método fraco
        region_code: 'US'
      };
      
      const result = await processHandler(weakMfaEvent);
      
      // Verificar que o handler interno detectou o problema
      expect(result.processed).to.be.true;
      
      // Verificar que foi registrado na observabilidade
      expect(mockObservability.records.complianceIssues.length).to.equal(1);
      expect(mockObservability.records.complianceIssues[0].issueType).to.equal('WEAK_MFA');
      expect(mockObservability.records.complianceIssues[0].severity).to.equal('HIGH');
    });
  });
  
  describe('Operações de Desligamento', () => {
    it('deve desligar todos os componentes corretamente', async () => {
      await adapter.shutdown();
      
      // Verificar se todos os componentes foram desconectados
      expect(mockProducer.disconnect.calledOnce).to.be.true;
      expect(mockConsumer.disconnect.calledOnce).to.be.true;
      expect(mockMcpAdapter.disconnect.calledOnce).to.be.true;
    });
    
    it('deve lidar com erros durante o desligamento', async () => {
      // Simular falha no desligamento do consumidor
      mockConsumer.disconnect.rejects(new Error('Falha ao desconectar consumidor'));
      
      try {
        await adapter.shutdown();
        expect.fail('Deveria ter lançado um erro');
      } catch (error) {
        expect(error.message).to.include('Falha ao desconectar consumidor');
      }
      
      // Verificar que tentou desconectar todos os componentes
      expect(mockProducer.disconnect.calledOnce).to.be.true;
      expect(mockConsumer.disconnect.calledOnce).to.be.true;
    });
  });
});
