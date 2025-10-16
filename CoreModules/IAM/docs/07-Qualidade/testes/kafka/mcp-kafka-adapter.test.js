/**
 * @fileoverview Testes para o adaptador MCP (Model Context Protocol) para Kafka
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/tests/kafka/mcp-kafka-adapter.test
 */

const { expect } = require('chai');
const sinon = require('sinon');
const { v4: uuidv4 } = require('uuid');
const { MCPKafkaAdapter } = require('../../../03-Desenvolvimento/auth-framework/kafka/mcp/mcp-kafka-adapter');
const { MCPClient, MCPMessage, MCPContext } = require('@innovabiz/mcp-client');

// Mock para o cliente MCP
class MockMCPClient {
  constructor() {
    this.isConnectedValue = false;
    this.eventHandlers = {};
    this.publishCalls = [];
    this.subscribeCalls = [];
  }
  
  connect() {
    this.isConnectedValue = true;
    if (this.eventHandlers.connected) {
      this.eventHandlers.connected();
    }
    return Promise.resolve();
  }
  
  disconnect() {
    this.isConnectedValue = false;
    if (this.eventHandlers.disconnected) {
      this.eventHandlers.disconnected();
    }
    return Promise.resolve();
  }
  
  isConnected() {
    return this.isConnectedValue;
  }
  
  on(event, handler) {
    this.eventHandlers[event] = handler;
    return this;
  }
  
  publish(mcpMessage) {
    this.publishCalls.push(mcpMessage);
    return Promise.resolve({ messageId: uuidv4() });
  }
  
  subscribe(channel, handler, options) {
    const subscriptionId = uuidv4();
    this.subscribeCalls.push({ channel, options, subscriptionId });
    this.subscriptionHandlers = this.subscriptionHandlers || {};
    this.subscriptionHandlers[subscriptionId] = handler;
    return Promise.resolve(subscriptionId);
  }
  
  // Método auxiliar para simular recebimento de mensagem
  simulateMessage(channel, message) {
    const subscribeCall = this.subscribeCalls.find(call => call.channel === channel);
    if (subscribeCall && this.subscriptionHandlers[subscribeCall.subscriptionId]) {
      return this.subscriptionHandlers[subscribeCall.subscriptionId](message);
    }
    return Promise.resolve();
  }
}

describe('MCPKafkaAdapter', () => {
  let adapter;
  let mockMCPClient;
  let sandbox;
  
  beforeEach(() => {
    sandbox = sinon.createSandbox();
    mockMCPClient = new MockMCPClient();
    
    // Mock para o módulo MCPClient
    sandbox.stub(MCPClient.prototype, 'connect').callsFake(() => mockMCPClient.connect());
    sandbox.stub(MCPClient.prototype, 'disconnect').callsFake(() => mockMCPClient.disconnect());
    sandbox.stub(MCPClient.prototype, 'isConnected').callsFake(() => mockMCPClient.isConnected());
    sandbox.stub(MCPClient.prototype, 'on').callsFake((event, handler) => {
      mockMCPClient.on(event, handler);
      return MCPClient.prototype;
    });
    sandbox.stub(MCPClient.prototype, 'publish').callsFake((message) => mockMCPClient.publish(message));
    sandbox.stub(MCPClient.prototype, 'subscribe').callsFake((channel, handler, options) => {
      return mockMCPClient.subscribe(channel, handler, options);
    });
    
    // Criar adaptador para teste
    adapter = new MCPKafkaAdapter({
      regionCode: 'BR',
      contextEnrichment: true
    });
  });
  
  afterEach(() => {
    sandbox.restore();
  });
  
  describe('Inicialização', () => {
    it('deve inicializar com configurações padrão', () => {
      const newAdapter = new MCPKafkaAdapter();
      expect(newAdapter.regionCode).to.equal('EU');
      expect(newAdapter.contextEnrichment).to.be.true;
    });
    
    it('deve inicializar com configurações personalizadas', () => {
      const newAdapter = new MCPKafkaAdapter({
        regionCode: 'US',
        contextEnrichment: false,
        mcpConfig: {
          maxReconnectAttempts: 5
        }
      });
      
      expect(newAdapter.regionCode).to.equal('US');
      expect(newAdapter.contextEnrichment).to.be.false;
    });
    
    it('deve aplicar configurações específicas de setor quando fornecidas', () => {
      const newAdapter = new MCPKafkaAdapter({
        regionCode: 'US',
        sectorConfig: {
          sector: 'HEALTHCARE'
        }
      });
      
      // Verificar se as configurações específicas de saúde foram aplicadas
      expect(newAdapter.regionalConfig.compliance.hipaa).to.be.true;
    });
  });
  
  describe('Operações de Conexão', () => {
    it('deve conectar com sucesso ao broker MCP', async () => {
      await adapter.connect();
      expect(mockMCPClient.isConnected()).to.be.true;
    });
    
    it('deve desconectar com sucesso do broker MCP', async () => {
      await adapter.connect();
      await adapter.disconnect();
      expect(mockMCPClient.isConnected()).to.be.false;
    });
    
    it('deve lidar com erros de conexão', async () => {
      // Simular falha de conexão
      MCPClient.prototype.connect.restore();
      sandbox.stub(MCPClient.prototype, 'connect').rejects(new Error('Falha de conexão'));
      
      try {
        await adapter.connect();
        expect.fail('Deveria ter lançado um erro');
      } catch (error) {
        expect(error.message).to.equal('Falha de conexão');
      }
    });
  });
  
  describe('Transformação de Eventos', () => {
    it('deve transformar um evento Kafka em uma mensagem MCP', () => {
      const kafkaEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'user123',
        tenant_id: 'tenant456',
        timestamp: Date.now(),
        method_code: 'K01'
      };
      
      const headers = {
        'correlation-id': 'corr123',
        partition: 0,
        offset: 100
      };
      
      const mcpMessage = adapter.transformToMCPMessage('iam-auth-events', kafkaEvent, headers);
      
      // Verificar campos básicos da mensagem MCP
      expect(mcpMessage.channel).to.equal('auth.events');
      expect(mcpMessage.type).to.equal('LOGIN_SUCCESS');
      expect(mcpMessage.payload).to.deep.include(kafkaEvent);
      
      // Verificar contexto
      expect(mcpMessage.context.requestId).to.equal('corr123');
      expect(mcpMessage.context.tenantId).to.equal('tenant456');
      expect(mcpMessage.context.userId).to.equal('user123');
      expect(mcpMessage.context.regionCode).to.equal('BR');
      
      // Verificar metadados
      expect(mcpMessage.metadata.originalTopic).to.equal('iam-auth-events');
      expect(mcpMessage.metadata.kafkaOffset).to.equal(100);
      expect(mcpMessage.metadata.kafkaPartition).to.equal(0);
      expect(mcpMessage.metadata.regionCode).to.equal('BR');
    });
    
    it('deve enriquecer o contexto MCP com base no tipo de evento', () => {
      const loginEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'user123',
        tenant_id: 'tenant456',
        timestamp: Date.now(),
        method_code: 'K01',
        client_info: {
          ip_address: '192.168.1.1',
          user_agent: 'Mozilla/5.0',
          device_id: 'device123'
        }
      };
      
      const mcpMessage = adapter.transformToMCPMessage('iam-auth-events', loginEvent);
      
      // Verificar enriquecimento de contexto para LOGIN_SUCCESS
      expect(mcpMessage.context.properties.authMethod).to.equal('K01');
      expect(mcpMessage.context.properties.clientIp).to.equal('192.168.1.1');
      expect(mcpMessage.context.properties.userAgent).to.equal('Mozilla/5.0');
      expect(mcpMessage.context.properties.deviceId).to.equal('device123');
    });
    
    it('deve aplicar validação específica de setor para evento de saúde', () => {
      // Criar adaptador com configuração de saúde
      const healthcareAdapter = new MCPKafkaAdapter({
        regionCode: 'US',
        sectorConfig: {
          sector: 'HEALTHCARE'
        }
      });
      
      // Simular validador de saúde
      const validateSpy = sandbox.spy();
      healthcareAdapter.healthcareValidators = [{
        validate: validateSpy
      }];
      
      const healthcareEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'doctor123',
        tenant_id: 'hospital456',
        timestamp: Date.now(),
        method_code: 'K05'
      };
      
      healthcareAdapter.transformToMCPMessage('iam-healthcare-auth-events', healthcareEvent);
      
      // Verificar se o validador foi chamado
      expect(validateSpy.calledOnce).to.be.true;
      expect(validateSpy.firstCall.args[0]).to.deep.include(healthcareEvent);
    });
  });
  
  describe('Operações de Publicação', () => {
    beforeEach(async () => {
      await adapter.connect();
    });
    
    it('deve publicar um evento Kafka como mensagem MCP', async () => {
      const kafkaEvent = {
        event_id: uuidv4(),
        event_type: 'LOGIN_SUCCESS',
        user_id: 'user123',
        tenant_id: 'tenant456',
        timestamp: Date.now()
      };
      
      const result = await adapter.publishToMCP('iam-auth-events', kafkaEvent);
      
      // Verificar resultado
      expect(result.success).to.be.true;
      expect(result.event_id).to.equal(kafkaEvent.event_id);
      
      // Verificar se a mensagem foi publicada corretamente
      expect(mockMCPClient.publishCalls.length).to.equal(1);
      const publishedMessage = mockMCPClient.publishCalls[0];
      expect(publishedMessage.channel).to.equal('auth.events');
      expect(publishedMessage.payload).to.deep.include(kafkaEvent);
    });
    
    it('deve rejeitar publicação quando não conectado', async () => {
      await adapter.disconnect();
      
      try {
        await adapter.publishToMCP('iam-auth-events', { event_id: uuidv4() });
        expect.fail('Deveria ter lançado um erro');
      } catch (error) {
        expect(error.message).to.include('não está conectado');
      }
    });
    
    it('deve lidar com erros de publicação', async () => {
      // Simular falha de publicação
      MCPClient.prototype.publish.restore();
      sandbox.stub(MCPClient.prototype, 'publish').rejects(new Error('Falha de publicação'));
      
      try {
        await adapter.publishToMCP('iam-auth-events', { event_id: uuidv4() });
        expect.fail('Deveria ter lançado um erro');
      } catch (error) {
        expect(error.message).to.equal('Falha de publicação');
      }
    });
  });
  
  describe('Operações de Inscrição', () => {
    beforeEach(async () => {
      await adapter.connect();
    });
    
    it('deve inscrever-se em um canal MCP', async () => {
      const handler = sinon.spy();
      const subscriptionId = await adapter.subscribeMCP('auth.events', handler);
      
      // Verificar se a inscrição foi realizada
      expect(subscriptionId).to.be.a('string');
      expect(mockMCPClient.subscribeCalls.length).to.equal(1);
      expect(mockMCPClient.subscribeCalls[0].channel).to.equal('auth.events');
    });
    
    it('deve processar mensagens MCP recebidas', async () => {
      const handlerSpy = sinon.spy();
      await adapter.subscribeMCP('auth.events', handlerSpy);
      
      // Simular recebimento de mensagem
      const mcpMessage = new MCPMessage({
        channel: 'auth.events',
        type: 'LOGIN_SUCCESS',
        payload: {
          event_id: uuidv4(),
          event_type: 'LOGIN_SUCCESS',
          user_id: 'user123'
        },
        context: new MCPContext({
          requestId: uuidv4(),
          tenantId: 'tenant456'
        })
      });
      
      await mockMCPClient.simulateMessage('auth.events', mcpMessage);
      
      // Verificar se o handler foi chamado com o evento convertido
      expect(handlerSpy.calledOnce).to.be.true;
      const kafkaEvent = handlerSpy.firstCall.args[0];
      expect(kafkaEvent).to.deep.include(mcpMessage.payload);
    });
    
    it('deve rejeitar inscrição quando não conectado', async () => {
      await adapter.disconnect();
      
      try {
        await adapter.subscribeMCP('auth.events', () => {});
        expect.fail('Deveria ter lançado um erro');
      } catch (error) {
        expect(error.message).to.include('não está conectado');
      }
    });
  });
  
  describe('Transformação MCP para Kafka', () => {
    it('deve transformar uma mensagem MCP em um evento Kafka', () => {
      const mcpMessage = new MCPMessage({
        channel: 'auth.events',
        type: 'LOGIN_SUCCESS',
        payload: {
          event_type: 'LOGIN_SUCCESS',
          user_id: 'user123'
        },
        context: new MCPContext({
          requestId: 'corr123',
          tenantId: 'tenant456',
          regionCode: 'BR',
          properties: {
            securityFlag_SUSPICIOUS_LOCATION: true
          }
        })
      });
      
      const kafkaEvent = adapter.transformToKafkaEvent(mcpMessage);
      
      // Verificar campos básicos
      expect(kafkaEvent.event_type).to.equal('LOGIN_SUCCESS');
      expect(kafkaEvent.user_id).to.equal('user123');
      
      // Verificar informações do contexto
      expect(kafkaEvent.correlation_id).to.equal('corr123');
      expect(kafkaEvent.tenant_id).to.equal('tenant456');
      expect(kafkaEvent.region_code).to.equal('BR');
      
      // Verificar flags de segurança
      expect(kafkaEvent.security_flags).to.include('SUSPICIOUS_LOCATION');
    });
  });
});
