/**
 * @file risk-management.integration.test.ts
 * @description Testes de integração para o serviço de integração com RiskManagement
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule } from '@nestjs/config';
import { HttpModule } from '@nestjs/axios';
import { KafkaModule } from '../../../src/modules/kafka/kafka.module';
import { RiskManagementService } from '../../../src/services/risk-management/risk-management.service';
import { RiskManagementClient } from '../../../src/clients/risk-management/risk-management.client';
import { KafkaProducer } from '../../../src/modules/kafka/kafka-producer.service';
import { HttpService } from '@nestjs/axios';
import { of } from 'rxjs';
import nock from 'nock';

// Import mocks
import { 
  mockStandardRiskAssessment,
  mockHighRiskAssessment,
  mockAuthEventReport
} from '../../setup/mocks/risk-management/risk-management-mock';

describe('RiskManagement Service Integration', () => {
  let riskManagementService: RiskManagementService;
  let riskManagementClient: RiskManagementClient;
  let httpService: HttpService;
  let kafkaProducer: KafkaProducer;
  
  // Mock para o KafkaProducer
  const mockKafkaProducer = {
    connect: jest.fn().mockResolvedValue(undefined),
    disconnect: jest.fn().mockResolvedValue(undefined),
    produce: jest.fn().mockImplementation((topic, key, value) => {
      return Promise.resolve({
        topicName: topic,
        partition: 0,
        errorCode: 0,
        offset: '0',
        timestamp: new Date().getTime(),
      });
    }),
  };
  
  beforeAll(async () => {
    const module: TestingModule = await Test.createTestingModule({
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
        KafkaModule,
      ],
      providers: [
        RiskManagementService,
        RiskManagementClient,
        {
          provide: KafkaProducer,
          useValue: mockKafkaProducer,
        },
      ],
    }).compile();

    riskManagementService = module.get<RiskManagementService>(RiskManagementService);
    riskManagementClient = module.get<RiskManagementClient>(RiskManagementClient);
    httpService = module.get<HttpService>(HttpService);
    kafkaProducer = module.get<KafkaProducer>(KafkaProducer);
  });

  beforeEach(() => {
    nock.cleanAll();
    jest.clearAllMocks();
  });

  afterAll(async () => {
    nock.restore();
  });

  describe('assessRisk', () => {
    it('deve avaliar risco para uma tentativa de login padrão', async () => {
      // Arrange
      const assessmentRequest = {
        userId: 'a8f7e52c-3f0d-4c12-9a85-4d789ad7b23c',
        tenantId: 'test-tenant',
        sessionId: '5f9e8a3d-1b9c-4e8f-8b1c-3d9e8a5f1b9c',
        ipAddress: '41.223.112.245',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        eventType: 'LOGIN_ATTEMPT',
        timestamp: new Date().toISOString(),
        deviceId: 'device-id-123',
        contextData: {
          location: 'Luanda, Angola',
          authenticationType: 'password',
          previousLogin: '2025-01-01T12:00:00Z'
        }
      };

      // Mock HTTP response
      jest.spyOn(httpService, 'post').mockReturnValueOnce(
        of({
          data: mockStandardRiskAssessment,
          status: 200,
          statusText: 'OK',
          headers: {},
          config: { url: '' } as any
        })
      );

      // Act
      const result = await riskManagementService.assessRisk(assessmentRequest);

      // Assert
      expect(result).toBeDefined();
      expect(result.riskLevel).toBe('low');
      expect(result.score).toBe(0.25);
      expect(result.actions).toEqual([]);
      expect(result.recommendations).toBeDefined();
      expect(result.recommendations.authLevel).toBe('standard');
      expect(httpService.post).toHaveBeenCalledTimes(1);
    });
  });
});