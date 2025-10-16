/**
 * @file iam-risk-integration.spec.ts
 * @description Testes de integração entre IAM e RiskManagement
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { HttpModule, HttpService } from '@nestjs/axios';
import { JwtModule } from '@nestjs/jwt';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../../../src/app.module';
import { AuthModule } from '../../../src/modules/auth/auth.module';
import { UsersModule } from '../../../src/modules/users/users.module';
import { TenantModule } from '../../../src/modules/tenant/tenant.module';
import { RiskManagementService } from '../../../src/services/risk-management/risk-management.service';
import { RiskAssessmentRepository } from '../../../src/repositories/risk-assessment.repository';
import { AuthService } from '../../../src/services/auth/auth.service';
import { of } from 'rxjs';
import nock from 'nock';

/**
 * Testes de integração da comunicação bidirecional entre IAM e RiskManagement
 * Verifica o funcionamento correto de avaliações de risco adaptativas em eventos de autenticação,
 * aplicação de políticas de acesso baseadas em risco, e registro de eventos de segurança.
 */
describe('IAM - RiskManagement Integration', () => {
  let app: INestApplication;
  let configService: ConfigService;
  let riskManagementService: RiskManagementService;
  let authService: AuthService;
  let httpService: HttpService;
  let jwtToken: string;

  // Usuário de teste
  const testUser = {
    id: 'user-test-123',
    username: 'riskuser@test.com',
    password: 'Risk123!',
    tenantId: 'test-tenant',
    displayName: 'Risk Test User',
    email: 'riskuser@test.com'
  };

  // Evento de login simulado
  const loginEvent = {
    userId: testUser.id,
    username: testUser.username,
    tenantId: testUser.tenantId,
    ipAddress: '192.168.1.1',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    eventType: 'LOGIN_ATTEMPT',
    timestamp: new Date().toISOString(),
    deviceId: 'device-xyz-123'
  };

  beforeAll(async () => {
    // Configurar módulo de teste
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
      providers: [
        {
          provide: RiskAssessmentRepository,
          useFactory: () => ({
            saveAssessment: jest.fn(),
            findByUserId: jest.fn(),
            findBySessionId: jest.fn(),
            updateAssessment: jest.fn()
          })
        }
      ]
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();

    configService = moduleFixture.get<ConfigService>(ConfigService);
    riskManagementService = moduleFixture.get<RiskManagementService>(RiskManagementService);
    authService = moduleFixture.get<AuthService>(AuthService);
    httpService = moduleFixture.get<HttpService>(HttpService);

    // Mock para RiskManagement API
    const riskManagementApiUrl = configService.get<string>('RISK_MANAGEMENT_API_URL', 'http://risk-api.innovabiz.local');
    
    // Limpar mocks anteriores
    nock.cleanAll();

    // Mock para avaliação de risco
    nock(riskManagementApiUrl)
      .post('/api/v1/assessments')
      .reply(200, {
        assessmentId: 'risk-123456',
        riskLevel: 'low',
        score: 0.25,
        factors: [
          { name: 'location', score: 0.2, reason: 'Localização conhecida' },
          { name: 'device', score: 0.3, reason: 'Dispositivo conhecido' },
          { name: 'time', score: 0.2, reason: 'Horário comum de acesso' }
        ],
        recommendations: {
          requiresAdditionalVerification: false,
          recommendedAuthMethods: ['password']
        }
      });

    // Mock para contexto de usuário
    nock(riskManagementApiUrl)
      .get(`/api/v1/users/${testUser.id}/context`)
      .reply(200, {
        riskProfile: 'standard',
        lastLoginLocation: 'Luanda, Angola',
        lastLoginTimestamp: new Date(Date.now() - 86400000).toISOString(),
        deviceTrustScore: 0.9,
        transactionVelocity: 'normal',
        accountAge: 180, // dias
        customAttributes: {
          sector: 'retail',
          verified: true,
          kycLevel: 'basic'
        }
      });

    // Mock para registro de eventos
    nock(riskManagementApiUrl)
      .post('/api/v1/events')
      .reply(200, {
        eventId: 'evt-123456',
        status: 'recorded',
        timestamp: new Date().toISOString()
      });

    // Mock para políticas de autenticação
    nock(riskManagementApiUrl)
      .get(`/api/v1/tenants/${testUser.tenantId}/policies`)
      .reply(200, {
        policies: [
          {
            id: 'policy-1',
            name: 'Política padrão',
            riskThresholds: {
              low: { score: 0.3, requiredFactors: 1 },
              medium: { score: 0.6, requiredFactors: 2 },
              high: { score: 0.8, requiredFactors: 3 }
            },
            stepUpTriggers: ['FINANCIAL_TRANSACTION', 'PROFILE_CHANGE', 'ADMIN_ACTION']
          }
        ]
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
    
    // Encerrar aplicação
    await app.close();
  });

  describe('Avaliação de Risco na Autenticação', () => {
    it('deve realizar avaliação de risco durante tentativa de login', async () => {
      // Spy no serviço de avaliação de risco
      const assessRiskSpy = jest.spyOn(riskManagementService, 'assessRisk');
      
      // Mock para a resposta HTTP
      jest.spyOn(httpService, 'post').mockImplementationOnce(() => {
        return of({
          data: {
            assessmentId: 'risk-123456',
            riskLevel: 'low',
            score: 0.25
          },
          status: 200,
          statusText: 'OK',
          headers: {},
          config: { url: '' } as any
        });
      });

      // Realiza login com informações contextuais
      const response = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send({
          username: testUser.username,
          password: testUser.password,
          tenantId: testUser.tenantId,
          deviceInfo: {
            deviceId: 'device-xyz-123',
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
          }
        })
        .expect(200);

      // Verificar que a avaliação de risco foi chamada
      expect(assessRiskSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          userId: testUser.id,
          eventType: 'LOGIN_ATTEMPT'
        })
      );
      
      // Verificar que a resposta inclui informações de risco
      expect(response.body.riskAssessment).toBeDefined();
      expect(response.body.riskAssessment.level).toBe('low');
    });

    it('deve aplicar política de autenticação adaptativa com base na avaliação de risco', async () => {
      // Mock para resposta de avaliação de risco de alto nível
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL', 'http://risk-api.innovabiz.local'))
        .post('/api/v1/assessments')
        .reply(200, {
          assessmentId: 'risk-789012',
          riskLevel: 'high',
          score: 0.85,
          factors: [
            { name: 'location', score: 0.9, reason: 'Localização anômala' },
            { name: 'device', score: 0.8, reason: 'Dispositivo desconhecido' },
            { name: 'time', score: 0.7, reason: 'Horário incomum de acesso' }
          ],
          recommendations: {
            requiresAdditionalVerification: true,
            recommendedAuthMethods: ['webauthn', 'otp']
          }
        });

      // Tenta login com contexto suspeito
      const response = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send({
          username: testUser.username,
          password: testUser.password,
          tenantId: testUser.tenantId,
          deviceInfo: {
            deviceId: 'unknown-device',
            ipAddress: '203.0.113.5', // IP de outra região
            userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 13_3 like Mac OS X)'
          }
        })
        .expect(200);

      // Verificar que o sistema solicita autenticação adicional
      expect(response.body.requiresAdditionalVerification).toBe(true);
      expect(response.body.recommendedAuthMethods).toContain('webauthn');
      expect(response.body.riskAssessment.level).toBe('high');
    });
  });

  describe('Relatório de Eventos de Segurança', () => {
    it('deve enviar eventos de segurança para o RiskManagement', async () => {
      // Spy no serviço de relatório de eventos
      const reportEventSpy = jest.spyOn(riskManagementService, 'reportAuthEvent');
      
      // Forçar um evento de segurança (alteração de senha)
      await request(app.getHttpServer())
        .post('/api/v1/users/change-password')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          currentPassword: 'Risk123!',
          newPassword: 'NewRisk456!',
          deviceInfo: {
            deviceId: 'device-xyz-123',
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
          }
        })
        .expect(200);

      // Verificar que o evento foi reportado
      expect(reportEventSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'PASSWORD_CHANGE',
          userId: expect.any(String)
        })
      );
    });
  });

  describe('Integração com Políticas de Acesso Baseadas em Risco', () => {
    it('deve aplicar políticas de acesso adaptativas para recursos sensíveis', async () => {
      // Mock para avaliação de risco ao acessar recursos sensíveis
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL', 'http://risk-api.innovabiz.local'))
        .post('/api/v1/assessments')
        .reply(200, {
          assessmentId: 'risk-345678',
          riskLevel: 'medium',
          score: 0.55,
          factors: [
            { name: 'location', score: 0.5, reason: 'Localização parcialmente conhecida' },
            { name: 'device', score: 0.6, reason: 'Dispositivo não registrado previamente' },
            { name: 'resource_sensitivity', score: 0.7, reason: 'Recurso de alta sensibilidade' }
          ],
          recommendations: {
            requiresAdditionalVerification: true,
            recommendedAuthMethods: ['otp'],
            restrictedResources: ['FINANCIAL_DATA', 'ADMIN_FUNCTIONS'],
            allowedResources: ['BASIC_PROFILE', 'PUBLIC_DATA']
          }
        });

      // Tenta acessar um recurso sensível
      const response = await request(app.getHttpServer())
        .get('/api/v1/financial-data')
        .set('Authorization', `Bearer ${jwtToken}`)
        .expect(200);

      // Verificar que a resposta indica necessidade de autenticação adicional
      expect(response.body.accessStatus).toBe('REQUIRES_ADDITIONAL_VERIFICATION');
      expect(response.body.riskAssessment).toBeDefined();
      expect(response.body.riskAssessment.level).toBe('medium');
      expect(response.body.verificationOptions).toContain('otp');
    });
  });
  
  describe('Detecção de Comportamento Anômalo', () => {
    it('deve identificar e responder a padrões de comportamento anômalos', async () => {
      // Mock para análise comportamental
      nock(configService.get<string>('RISK_MANAGEMENT_API_URL', 'http://risk-api.innovabiz.local'))
        .post('/api/v1/behavioral-analysis')
        .reply(200, {
          analysisId: 'ba-123456',
          anomalyDetected: true,
          confidenceScore: 0.82,
          behaviorMatches: false,
          anomalyFactors: [
            { factor: 'login_time_pattern', description: 'Login em horário atípico' },
            { factor: 'access_pattern', description: 'Acesso a recursos não usuais' }
          ],
          recommendations: {
            authLevel: 'step_up',
            monitoringLevel: 'enhanced',
            actions: ['REQUEST_ADDITIONAL_VERIFICATION', 'RECORD_DETAILED_ACTIVITY']
          }
        });
      
      // Simula uma série de operações com padrão incomum
      await request(app.getHttpServer())
        .post('/api/v1/api/behavior-test')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          operations: [
            { type: 'ACCESS', resource: 'user_management', timestamp: new Date().toISOString() },
            { type: 'EXPORT', resource: 'financial_data', timestamp: new Date().toISOString() },
            { type: 'CHANGE', resource: 'security_settings', timestamp: new Date().toISOString() }
          ]
        })
        .expect(200);
      
      // Tenta uma operação sensível após comportamento suspeito
      const response = await request(app.getHttpServer())
        .post('/api/v1/financial-transactions/initiate')
        .set('Authorization', `Bearer ${jwtToken}`)
        .send({
          amount: 5000,
          recipientId: 'user-999',
          description: 'Test transfer'
        })
        .expect(200);
      
      // Verificar que o sistema aplicou restrições baseadas em comportamento
      expect(response.body.requiresAdditionalVerification).toBe(true);
      expect(response.body.behavioralAnalysis).toBeDefined();
      expect(response.body.behavioralAnalysis.anomalyDetected).toBe(true);
      expect(response.body.securityActions).toContain('REQUEST_ADDITIONAL_VERIFICATION');
    });
  });
});