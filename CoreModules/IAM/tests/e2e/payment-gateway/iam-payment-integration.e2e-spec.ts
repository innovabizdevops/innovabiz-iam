/**
 * @file iam-payment-integration.e2e-spec.ts
 * @description Testes de integração end-to-end entre IAM e Payment Gateway
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { HttpModule } from '@nestjs/axios';
import { JwtModule } from '@nestjs/jwt';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { TestContainersSetup } from '../../setup/testcontainers/testcontainers.setup';
import { AppModule } from '../../../src/app.module';
import { PaymentGatewayService } from '../../../src/services/payment-gateway/payment-gateway.service';
import { AuthModule } from '../../../src/modules/auth/auth.module';
import { UsersModule } from '../../../src/modules/users/users.module';
import { TenantModule } from '../../../src/modules/tenant/tenant.module';
import nock from 'nock';

describe('IAM - Payment Gateway Integration (E2E)', () => {
  let app: INestApplication;
  let configService: ConfigService;
  let paymentGatewayService: PaymentGatewayService;
  let jwtToken: string;

  const testUser = {
    username: 'paymentuser@test.com',
    password: 'Payment123!',
    name: 'Payment Test User',
    tenantId: 'test-tenant'
  };

  const testPaymentMethod = {
    id: 'pm_test_12345',
    type: 'card',
    name: 'Visa **** 4242',
    expMonth: 12,
    expYear: 2030,
    last4: '4242',
    isDefault: true
  };

  beforeAll(async () => {
    // Configurar TestContainers
    await TestContainersSetup.setupContainers();

    // Configurar o módulo de teste
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [
        ConfigModule.forRoot({
          isGlobal: true,
          envFilePath: ['./tests/.env.test'],
        }),
        HttpModule.registerAsync({
          useFactory: () => ({
            timeout: 5000,
            maxRedirects: 5,
          }),
        }),
        JwtModule.registerAsync({
          imports: [ConfigModule],
          useFactory: async (configService: ConfigService) => ({
            secret: configService.get<string>('JWT_SECRET'),
            signOptions: {
              expiresIn: configService.get<string>('JWT_EXPIRATION', '1h'),
            },
          }),
          inject: [ConfigService],
        }),
        AppModule,
        AuthModule,
        UsersModule,
        TenantModule,
      ],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();

    configService = moduleFixture.get<ConfigService>(ConfigService);
    paymentGatewayService = moduleFixture.get<PaymentGatewayService>(PaymentGatewayService);

    // Configurar mocks para Payment Gateway API
    const paymentGatewayApiUrl = configService.get<string>('PAYMENT_GATEWAY_API_URL');
    
    // Limpar qualquer mock anterior
    nock.cleanAll();

    // Mock para autenticação com Payment Gateway
    nock(paymentGatewayApiUrl)
      .post('/api/v1/auth/token')
      .reply(200, {
        access_token: 'payment-gateway-mock-token',
        expires_in: 3600,
        token_type: 'Bearer'
      });

    // Mock para verificação de pagamentos do usuário
    nock(paymentGatewayApiUrl)
      .get(`/api/v1/users/${testUser.username}/payment-methods`)
      .reply(200, {
        payment_methods: [testPaymentMethod]
      });

    // Realizar autenticação para obter token JWT
    const authResponse = await request(app.getHttpServer())
      .post('/api/v1/auth/login')
      .send({
        username: testUser.username,
        password: testUser.password,
        tenantId: testUser.tenantId
      });

    jwtToken = authResponse.body.access_token;
  });

  afterAll(async () => {
    // Limpar mocks
    nock.cleanAll();
    
    // Encerrar aplicação e containers
    await app.close();
    await TestContainersSetup.teardownContainers();
  });

  describe('Integração IAM-PaymentGateway', () => {
    it('deve autenticar usuário e verificar métodos de pagamento', async () => {
      // Executar requisição para endpoint que integra com Payment Gateway
      const response = await request(app.getHttpServer())
        .get('/api/v1/user/payment-methods')
        .set('Authorization', `Bearer ${jwtToken}`)
        .expect(200);

      // Verificar resposta
      expect(response.body).toBeDefined();
      expect(response.body.paymentMethods).toHaveLength(1);
      expect(response.body.paymentMethods[0].id).toBe(testPaymentMethod.id);
      expect(response.body.paymentMethods[0].last4).toBe(testPaymentMethod.last4);
    });

    it('deve rejeitar acesso a dados de pagamento para usuário não autenticado', async () => {
      // Tentar acessar endpoint sem autenticação
      await request(app.getHttpServer())
        .get('/api/v1/user/payment-methods')
        .expect(401);
    });

    it('deve aplicar step-up authentication para operações financeiras de alto valor', async () => {
      // Mock para tentativa de transação de alto valor
      nock(configService.get<string>('PAYMENT_GATEWAY_API_URL'))
        .post('/api/v1/transactions/validate')
        .reply(200, {
          requiresStepUpAuth: true,
          riskLevel: 'high',
          allowedAuthMethods: ['webauthn', 'otp']
        });

      // Tentar realizar uma transação de alto valor
      const response = await request(app.getHttpServer())
        .post('/api/v1/transactions/initiate')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          amount: 10000,
          currency: 'USD',
          description: 'Transação de alto valor para teste'
        })
        .expect(200);

      // Verificar que a resposta indica necessidade de autenticação adicional
      expect(response.body.requiresStepUpAuth).toBe(true);
      expect(response.body.allowedAuthMethods).toContain('webauthn');
    });
    
    it('deve integrar com avaliação de risco do RiskManagement durante transações financeiras', async () => {
      // Mock para APIs do RiskManagement e PaymentGateway
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL'))
        .post('/api/v1/assessments')
        .reply(200, {
          assessmentId: 'risk-123456',
          riskLevel: 'medium',
          score: 0.65,
          factors: [
            { name: 'transaction_amount', score: 0.7, reason: 'Valor elevado para o padrão do usuário' },
            { name: 'user_history', score: 0.3, reason: 'Usuário com histórico de transações regulares' },
            { name: 'device_trust', score: 0.6, reason: 'Dispositivo conhecido mas localização incomum' }
          ],
          recommendations: {
            requiresAdditionalVerification: true,
            recommendedAuthMethods: ['webauthn']
          }
        });
      
      nock(configService.get<string>('PAYMENT_GATEWAY_API_URL'))
        .post('/api/v1/transactions/process')
        .reply(200, {
          transactionId: 'txn_123456',
          status: 'pending_authentication',
          requiresVerification: true,
          verificationDetails: {
            type: 'step_up_auth',
            methods: ['webauthn'],
            challengeId: 'challenge_xyz',
            expiresAt: new Date(Date.now() + 300000).toISOString()
          }
        });

      // Executar transação com integração de risco
      const response = await request(app.getHttpServer())
        .post('/api/v1/transactions/secure-payment')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          amount: 5000,
          currency: 'USD',
          paymentMethodId: testPaymentMethod.id,
          description: 'Pagamento com avaliação de risco integrada'
        })
        .expect(200);

      // Verificar integração com avaliação de risco
      expect(response.body.status).toBe('pending_authentication');
      expect(response.body.riskAssessment).toBeDefined();
      expect(response.body.riskAssessment.level).toBe('medium');
      expect(response.body.verificationDetails).toBeDefined();
      expect(response.body.verificationDetails.methods).toContain('webauthn');
    });
    
    it('deve implementar autorização adaptativa baseada no contexto de usuário e risco', async () => {
      // Mock para contexto do usuário no RiskManagement
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL'))
        .get(`/api/v1/users/${testUser.username}/context`)
        .reply(200, {
          riskProfile: 'standard',
          lastLoginLocation: 'Luanda, Angola',
          lastLoginTimestamp: new Date(Date.now() - 86400000).toISOString(),
          deviceTrustScore: 0.85,
          transactionVelocity: 'normal',
          accountAge: 365, // dias
          customAttributes: {
            sector: 'banking',
            verified: true,
            kycLevel: 'full'
          }
        });
      
      // Mock para análise comportamental
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL'))
        .post('/api/v1/behavioral-analysis')
        .reply(200, {
          anomalyDetected: false,
          confidenceScore: 0.92,
          behaviorMatches: true,
          recommendations: {
            authLevel: 'standard',
            monitoringLevel: 'normal'
          }
        });
      
      // Teste de autorização adaptativa para um recurso sensível
      const response = await request(app.getHttpServer())
        .get('/api/v1/user/account-statements')
        .set('Authorization', `Bearer ${jwtToken}`)
        .query({ sensitive: true })
        .expect(200);
      
      // Verificar que a autorização adaptativa foi aplicada
      expect(response.body.accessGranted).toBe(true);
      expect(response.body.adaptiveContext).toBeDefined();
      expect(response.body.adaptiveContext.appliedPolicy).toBe('context_aware_authentication');
      expect(response.body.adaptiveContext.riskBasedDecision).toBeDefined();
      expect(response.body.adaptiveContext.requiredAuthLevel).toBe('standard');
      expect(response.body.sensitiveDataFiltered).toBe(false);
    });

    it('deve implementar autenticação multi-fator com PaymentGateway para transações internacionais', async () => {
      // Mock para avaliação de risco de transação internacional
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL'))
        .post('/api/v1/transaction-assessment')
        .reply(200, {
          riskLevel: 'high',
          reason: 'international_transaction',
          requiresMFA: true
        });
      
      // Mock para iniciar transação internacional
      nock(configService.get<string>('PAYMENT_GATEWAY_API_URL'))
        .post('/api/v1/international/transfer/initiate')
        .reply(200, {
          transferId: 'intl_transfer_123',
          status: 'awaiting_verification',
          verificationRequired: true,
          verificationMethods: ['otp', 'webauthn'],
          sessionId: 'verify_session_123',
          expiresAt: new Date(Date.now() + 600000).toISOString()
        });
      
      // Tentar realizar transferência internacional
      const response = await request(app.getHttpServer())
        .post('/api/v1/transfers/international')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          sourceAccount: '123456789',
          destinationAccount: 'BE71096123456769',
          destinationBankCode: 'BBRUBEBB',
          amount: 2500,
          currency: 'EUR',
          description: 'Transferência internacional de teste',
          recipientName: 'João Silva',
          recipientAddress: {
            street: 'Rue de la Loi',
            city: 'Brussels',
            country: 'Belgium',
            postalCode: '1000'
          }
        })
        .expect(200);
      
      // Verificar que a transação requer verificação adicional
      expect(response.body.status).toBe('awaiting_verification');
      expect(response.body.verificationRequired).toBe(true);
      expect(response.body.verificationMethods).toContain('webauthn');
      expect(response.body.verificationMethods).toContain('otp');
      expect(response.body.riskAssessment.level).toBe('high');
    });
  });
});