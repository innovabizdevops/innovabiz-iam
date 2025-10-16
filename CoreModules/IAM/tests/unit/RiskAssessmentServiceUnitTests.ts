/**
 * 🛡️ TESTES UNITÁRIOS - RISK ASSESSMENT SERVICE INNOVABIZ
 * Framework: Jest + NestJS Testing + Machine Learning
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST AI RMF, ISO/IEC 42001, PCI DSS 4.0, GDPR
 * Coverage Target: >95% | Security: OWASP ML Security
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigService } from '@nestjs/config';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Cache } from 'cache-manager';

import { RiskAssessmentService } from '../../CoreModules/Backend/IAM/RiskAssessmentService';
import { AuditService } from '../../CoreModules/Backend/IAM/AuditService';

import { RiskProfile } from '../../CoreModules/Backend/IAM/entities/RiskProfile.entity';
import { RiskEvent } from '../../CoreModules/Backend/IAM/entities/RiskEvent.entity';
import { User } from '../../CoreModules/Backend/IAM/entities/User.entity';

import {
  RiskAssessmentRequest,
  RiskAssessmentResult,
  RiskFactorType,
  RiskLevel,
  DeviceRiskFactors,
  LocationRiskFactors,
  BehavioralRiskFactors
} from '../../CoreModules/Backend/IAM/dto';

import {
  BadRequestException,
  InternalServerErrorException
} from '@nestjs/common';

// Mock do módulo de Machine Learning
jest.mock('../../CoreModules/AI/RiskMLModel', () => ({
  RiskMLModel: {
    predict: jest.fn(),
    updateModel: jest.fn(),
    getFeatureImportance: jest.fn()
  }
}));

describe('🛡️ Risk Assessment Service - Unit Tests', () => {
  let service: RiskAssessmentService;
  let auditService: AuditService;
  let configService: ConfigService;
  let cacheManager: Cache;
  let riskProfileRepository: Repository<RiskProfile>;
  let riskEventRepository: Repository<RiskEvent>;

  // Mock data
  const mockUser = {
    id: 'user-123',
    tenantId: 'tenant-123',
    email: 'test@innovabiz.com',
    username: 'testuser',
    createdAt: new Date('2024-01-01'),
    lastLoginAt: new Date()
  };

  const mockRiskProfile: RiskProfile = {
    id: 'profile-123',
    userId: 'user-123',
    tenantId: 'tenant-123',
    baselineRiskScore: 25,
    currentRiskScore: 30,
    riskLevel: 'low',
    lastAssessmentAt: new Date(),
    deviceFingerprints: ['device-123'],
    trustedLocations: ['192.168.1.0/24'],
    behaviorPatterns: {
      averageSessionDuration: 3600,
      typicalLoginHours: [9, 10, 11, 14, 15, 16],
      commonUserAgents: ['Mozilla/5.0']
    },
    riskFactors: [],
    createdAt: new Date(),
    updatedAt: new Date(),
    user: mockUser as User
  };

  const mockRiskEvent: RiskEvent = {
    id: 'event-123',
    userId: 'user-123',
    tenantId: 'tenant-123',
    eventType: 'authentication',
    riskScore: 35,
    riskLevel: 'medium',
    riskFactors: ['unusual_location', 'new_device'],
    metadata: {
      ipAddress: '203.0.113.1',
      userAgent: 'Unknown Browser',
      location: 'Unknown'
    },
    timestamp: new Date(),
    user: mockUser as User
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        RiskAssessmentService,
        {
          provide: AuditService,
          useValue: {
            logEvent: jest.fn(),
            logSecurityEvent: jest.fn(),
            logRiskEvent: jest.fn()
          }
        },
        {
          provide: ConfigService,
          useValue: {
            get: jest.fn((key: string) => {
              const config = {
                'risk.thresholds.low': 30,
                'risk.thresholds.medium': 60,
                'risk.thresholds.high': 80,
                'risk.weights.device': 0.3,
                'risk.weights.location': 0.25,
                'risk.weights.behavior': 0.25,
                'risk.weights.temporal': 0.2,
                'risk.ml.enabled': true,
                'risk.ml.modelPath': '/models/risk-model.pkl',
                'risk.cache.ttl': 300
              };
              return config[key];
            })
          }
        },
        {
          provide: CACHE_MANAGER,
          useValue: {
            get: jest.fn(),
            set: jest.fn(),
            del: jest.fn()
          }
        },
        {
          provide: getRepositoryToken(RiskProfile),
          useValue: {
            create: jest.fn(),
            save: jest.fn(),
            findOne: jest.fn(),
            update: jest.fn(),
            delete: jest.fn()
          }
        },
        {
          provide: getRepositoryToken(RiskEvent),
          useValue: {
            create: jest.fn(),
            save: jest.fn(),
            find: jest.fn(),
            count: jest.fn()
          }
        }
      ]
    }).compile();

    service = module.get<RiskAssessmentService>(RiskAssessmentService);
    auditService = module.get<AuditService>(AuditService);
    configService = module.get<ConfigService>(ConfigService);
    cacheManager = module.get<Cache>(CACHE_MANAGER);
    riskProfileRepository = module.get<Repository<RiskProfile>>(getRepositoryToken(RiskProfile));
    riskEventRepository = module.get<Repository<RiskEvent>>(getRepositoryToken(RiskEvent));
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('🔍 Registration Risk Assessment', () => {
    describe('assessRegistrationRisk', () => {
      const registrationRequest: RiskAssessmentRequest = {
        userId: 'user-123',
        tenantId: 'tenant-123',
        eventType: 'registration',
        ipAddress: '192.168.1.100',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        deviceFingerprint: 'device-fingerprint-123',
        timestamp: new Date(),
        metadata: {
          registrationMethod: 'webauthn',
          authenticatorType: 'platform'
        }
      };

      it('should assess registration risk for new user', async () => {
        // Arrange
        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(null);
        jest.spyOn(service, 'calculateDeviceRisk').mockResolvedValue({
          score: 20,
          factors: ['new_device']
        });
        jest.spyOn(service, 'calculateLocationRisk').mockResolvedValue({
          score: 15,
          factors: ['known_location']
        });
        jest.spyOn(service, 'calculateBehavioralRisk').mockResolvedValue({
          score: 10,
          factors: []
        });
        jest.spyOn(service, 'calculateTemporalRisk').mockResolvedValue({
          score: 5,
          factors: []
        });
        jest.spyOn(riskEventRepository, 'create').mockReturnValue(mockRiskEvent);
        jest.spyOn(riskEventRepository, 'save').mockResolvedValue(mockRiskEvent);
        jest.spyOn(auditService, 'logRiskEvent').mockResolvedValue(undefined);

        // Act
        const result = await service.assessRegistrationRisk(registrationRequest);

        // Assert
        expect(result).toEqual({
          riskScore: 25, // Weighted average
          riskLevel: 'low',
          factors: ['new_device', 'known_location'],
          confidence: expect.any(Number),
          recommendations: expect.any(Array)
        });
        expect(riskEventRepository.save).toHaveBeenCalled();
        expect(auditService.logRiskEvent).toHaveBeenCalledWith({
          userId: 'user-123',
          tenantId: 'tenant-123',
          eventType: 'registration',
          riskScore: 25,
          riskLevel: 'low',
          factors: ['new_device', 'known_location']
        });
      });

      it('should use cached risk profile when available', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'get').mockResolvedValue({
          riskScore: 30,
          riskLevel: 'low',
          factors: ['cached_factor']
        });

        // Act
        const result = await service.assessRegistrationRisk(registrationRequest);

        // Assert
        expect(result.riskScore).toBe(30);
        expect(result.factors).toContain('cached_factor');
        expect(riskProfileRepository.findOne).not.toHaveBeenCalled();
      });

      it('should handle high-risk registration', async () => {
        // Arrange
        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(null);
        jest.spyOn(service, 'calculateDeviceRisk').mockResolvedValue({
          score: 80,
          factors: ['suspicious_device', 'tor_network']
        });
        jest.spyOn(service, 'calculateLocationRisk').mockResolvedValue({
          score: 70,
          factors: ['high_risk_country', 'vpn_detected']
        });
        jest.spyOn(service, 'calculateBehavioralRisk').mockResolvedValue({
          score: 60,
          factors: ['rapid_registration']
        });
        jest.spyOn(service, 'calculateTemporalRisk').mockResolvedValue({
          score: 50,
          factors: ['unusual_time']
        });

        const highRiskEvent = { ...mockRiskEvent, riskScore: 85, riskLevel: 'high' };
        jest.spyOn(riskEventRepository, 'create').mockReturnValue(highRiskEvent);
        jest.spyOn(riskEventRepository, 'save').mockResolvedValue(highRiskEvent);

        // Act
        const result = await service.assessRegistrationRisk(registrationRequest);

        // Assert
        expect(result.riskLevel).toBe('high');
        expect(result.riskScore).toBeGreaterThan(80);
        expect(result.recommendations).toContain('require_additional_verification');
      });
    });
  });

  describe('🔐 Authentication Risk Assessment', () => {
    describe('assessAuthenticationRisk', () => {
      const authRequest: RiskAssessmentRequest = {
        userId: 'user-123',
        tenantId: 'tenant-123',
        eventType: 'authentication',
        ipAddress: '192.168.1.100',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        deviceFingerprint: 'device-fingerprint-123',
        timestamp: new Date(),
        metadata: {
          credentialId: 'cred-123',
          authenticatorType: 'platform'
        }
      };

      it('should assess authentication risk for existing user', async () => {
        // Arrange
        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);
        jest.spyOn(service, 'calculateDeviceRisk').mockResolvedValue({
          score: 10,
          factors: ['trusted_device']
        });
        jest.spyOn(service, 'calculateLocationRisk').mockResolvedValue({
          score: 5,
          factors: ['trusted_location']
        });
        jest.spyOn(service, 'calculateBehavioralRisk').mockResolvedValue({
          score: 15,
          factors: ['normal_pattern']
        });
        jest.spyOn(service, 'calculateTemporalRisk').mockResolvedValue({
          score: 8,
          factors: []
        });

        // Act
        const result = await service.assessAuthenticationRisk(authRequest);

        // Assert
        expect(result.riskLevel).toBe('low');
        expect(result.riskScore).toBeLessThan(30);
        expect(result.factors).toContain('trusted_device');
        expect(result.factors).toContain('trusted_location');
      });

      it('should detect anomalous authentication patterns', async () => {
        // Arrange
        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);
        jest.spyOn(service, 'calculateBehavioralRisk').mockResolvedValue({
          score: 75,
          factors: ['unusual_time', 'rapid_succession', 'velocity_anomaly']
        });
        jest.spyOn(service, 'calculateDeviceRisk').mockResolvedValue({
          score: 60,
          factors: ['new_device']
        });

        // Act
        const result = await service.assessAuthenticationRisk(authRequest);

        // Assert
        expect(result.riskLevel).toBe('high');
        expect(result.factors).toContain('unusual_time');
        expect(result.factors).toContain('velocity_anomaly');
        expect(result.recommendations).toContain('require_step_up_authentication');
      });
    });
  });

  describe('🔧 Risk Factor Calculations', () => {
    describe('calculateDeviceRisk', () => {
      it('should calculate low risk for trusted device', async () => {
        // Arrange
        const deviceFactors: DeviceRiskFactors = {
          fingerprint: 'device-123',
          isKnown: true,
          userAgent: 'Mozilla/5.0',
          platform: 'Windows',
          isMobile: false
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);

        // Act
        const result = await service.calculateDeviceRisk('user-123', 'tenant-123', deviceFactors);

        // Assert
        expect(result.score).toBeLessThan(30);
        expect(result.factors).toContain('trusted_device');
      });

      it('should calculate high risk for suspicious device', async () => {
        // Arrange
        const deviceFactors: DeviceRiskFactors = {
          fingerprint: 'suspicious-device',
          isKnown: false,
          userAgent: 'Unknown/Suspicious',
          platform: 'Unknown',
          isMobile: false,
          isTor: true,
          isVpn: true
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);

        // Act
        const result = await service.calculateDeviceRisk('user-123', 'tenant-123', deviceFactors);

        // Assert
        expect(result.score).toBeGreaterThan(70);
        expect(result.factors).toContain('tor_network');
        expect(result.factors).toContain('vpn_detected');
        expect(result.factors).toContain('unknown_device');
      });
    });

    describe('calculateLocationRisk', () => {
      it('should calculate low risk for trusted location', async () => {
        // Arrange
        const locationFactors: LocationRiskFactors = {
          ipAddress: '192.168.1.100',
          country: 'US',
          region: 'California',
          city: 'San Francisco',
          isKnown: true,
          isVpn: false,
          isTor: false
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);

        // Act
        const result = await service.calculateLocationRisk('user-123', 'tenant-123', locationFactors);

        // Assert
        expect(result.score).toBeLessThan(30);
        expect(result.factors).toContain('trusted_location');
      });

      it('should calculate high risk for high-risk location', async () => {
        // Arrange
        const locationFactors: LocationRiskFactors = {
          ipAddress: '203.0.113.1',
          country: 'XX', // High-risk country
          region: 'Unknown',
          city: 'Unknown',
          isKnown: false,
          isVpn: true,
          isTor: true,
          riskScore: 85
        };

        // Act
        const result = await service.calculateLocationRisk('user-123', 'tenant-123', locationFactors);

        // Assert
        expect(result.score).toBeGreaterThan(70);
        expect(result.factors).toContain('high_risk_country');
        expect(result.factors).toContain('vpn_detected');
        expect(result.factors).toContain('tor_network');
      });
    });

    describe('calculateBehavioralRisk', () => {
      it('should calculate low risk for normal behavior', async () => {
        // Arrange
        const behavioralFactors: BehavioralRiskFactors = {
          loginFrequency: 'normal',
          sessionDuration: 3600,
          typicalHours: [9, 10, 11, 14, 15, 16],
          currentHour: 10,
          velocityScore: 0.2,
          patternDeviation: 0.1
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);

        // Act
        const result = await service.calculateBehavioralRisk('user-123', 'tenant-123', behavioralFactors);

        // Assert
        expect(result.score).toBeLessThan(30);
        expect(result.factors).toContain('normal_pattern');
      });

      it('should calculate high risk for anomalous behavior', async () => {
        // Arrange
        const behavioralFactors: BehavioralRiskFactors = {
          loginFrequency: 'high',
          sessionDuration: 60, // Very short session
          typicalHours: [9, 10, 11, 14, 15, 16],
          currentHour: 3, // Unusual time
          velocityScore: 0.9, // High velocity
          patternDeviation: 0.8, // High deviation
          rapidSuccession: true
        };

        // Act
        const result = await service.calculateBehavioralRisk('user-123', 'tenant-123', behavioralFactors);

        // Assert
        expect(result.score).toBeGreaterThan(70);
        expect(result.factors).toContain('unusual_time');
        expect(result.factors).toContain('velocity_anomaly');
        expect(result.factors).toContain('rapid_succession');
      });
    });
  });

  describe('📊 Risk Profile Management', () => {
    describe('updateRiskProfile', () => {
      it('should update existing risk profile', async () => {
        // Arrange
        const updateData = {
          currentRiskScore: 35,
          riskLevel: 'medium' as RiskLevel,
          lastAssessmentAt: new Date()
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);
        jest.spyOn(riskProfileRepository, 'save').mockResolvedValue({
          ...mockRiskProfile,
          ...updateData
        });
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        const result = await service.updateRiskProfile('user-123', 'tenant-123', updateData);

        // Assert
        expect(result.currentRiskScore).toBe(35);
        expect(result.riskLevel).toBe('medium');
        expect(riskProfileRepository.save).toHaveBeenCalled();
        expect(cacheManager.set).toHaveBeenCalledWith(
          'risk-profile:user-123:tenant-123',
          expect.any(Object),
          300
        );
      });

      it('should create new risk profile if not exists', async () => {
        // Arrange
        const createData = {
          userId: 'user-456',
          tenantId: 'tenant-123',
          baselineRiskScore: 25,
          currentRiskScore: 25,
          riskLevel: 'low' as RiskLevel
        };

        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(null);
        jest.spyOn(riskProfileRepository, 'create').mockReturnValue(createData as any);
        jest.spyOn(riskProfileRepository, 'save').mockResolvedValue(createData as any);

        // Act
        const result = await service.updateRiskProfile('user-456', 'tenant-123', createData);

        // Assert
        expect(result.userId).toBe('user-456');
        expect(result.baselineRiskScore).toBe(25);
        expect(riskProfileRepository.create).toHaveBeenCalledWith(createData);
      });
    });

    describe('getRiskProfile', () => {
      it('should return cached risk profile', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'get').mockResolvedValue(mockRiskProfile);

        // Act
        const result = await service.getRiskProfile('user-123', 'tenant-123');

        // Assert
        expect(result).toEqual(mockRiskProfile);
        expect(riskProfileRepository.findOne).not.toHaveBeenCalled();
      });

      it('should fetch and cache risk profile from database', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'get').mockResolvedValue(null);
        jest.spyOn(riskProfileRepository, 'findOne').mockResolvedValue(mockRiskProfile);
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        const result = await service.getRiskProfile('user-123', 'tenant-123');

        // Assert
        expect(result).toEqual(mockRiskProfile);
        expect(riskProfileRepository.findOne).toHaveBeenCalledWith({
          where: { userId: 'user-123', tenantId: 'tenant-123' }
        });
        expect(cacheManager.set).toHaveBeenCalled();
      });
    });
  });

  describe('🤖 Machine Learning Integration', () => {
    describe('predictRiskWithML', () => {
      it('should predict risk using ML model', async () => {
        // Arrange
        const features = {
          deviceRisk: 0.2,
          locationRisk: 0.15,
          behavioralRisk: 0.1,
          temporalRisk: 0.05,
          historicalRisk: 0.25
        };

        const mockMLPrediction = {
          riskScore: 0.28,
          confidence: 0.85,
          featureImportance: {
            historicalRisk: 0.4,
            deviceRisk: 0.3,
            locationRisk: 0.2,
            behavioralRisk: 0.1
          }
        };

        jest.spyOn(service, 'callMLModel').mockResolvedValue(mockMLPrediction);

        // Act
        const result = await service.predictRiskWithML(features);

        // Assert
        expect(result.riskScore).toBe(28); // Converted to 0-100 scale
        expect(result.confidence).toBe(0.85);
        expect(result.featureImportance).toBeDefined();
      });

      it('should fallback to rule-based assessment if ML fails', async () => {
        // Arrange
        const features = {
          deviceRisk: 0.2,
          locationRisk: 0.15,
          behavioralRisk: 0.1,
          temporalRisk: 0.05
        };

        jest.spyOn(service, 'callMLModel').mockRejectedValue(new Error('ML model unavailable'));
        jest.spyOn(service, 'calculateRuleBasedRisk').mockReturnValue(25);

        // Act
        const result = await service.predictRiskWithML(features);

        // Assert
        expect(result.riskScore).toBe(25);
        expect(result.confidence).toBeLessThan(0.7); // Lower confidence for rule-based
      });
    });
  });

  describe('📈 Risk Analytics', () => {
    describe('getRiskTrends', () => {
      it('should return risk trends for user', async () => {
        // Arrange
        const mockRiskEvents = [
          { ...mockRiskEvent, timestamp: new Date('2024-01-01'), riskScore: 20 },
          { ...mockRiskEvent, timestamp: new Date('2024-01-02'), riskScore: 25 },
          { ...mockRiskEvent, timestamp: new Date('2024-01-03'), riskScore: 30 }
        ];

        jest.spyOn(riskEventRepository, 'find').mockResolvedValue(mockRiskEvents);

        // Act
        const result = await service.getRiskTrends('user-123', 'tenant-123', 7);

        // Assert
        expect(result.trend).toBe('increasing');
        expect(result.averageRisk).toBe(25);
        expect(result.dataPoints).toHaveLength(3);
      });
    });

    describe('getSystemRiskMetrics', () => {
      it('should return system-wide risk metrics', async () => {
        // Arrange
        jest.spyOn(riskEventRepository, 'count').mockResolvedValue(1000);
        jest.spyOn(service, 'calculateSystemRiskDistribution').mockResolvedValue({
          low: 700,
          medium: 250,
          high: 50
        });

        // Act
        const result = await service.getSystemRiskMetrics('tenant-123');

        // Assert
        expect(result.totalAssessments).toBe(1000);
        expect(result.riskDistribution.low).toBe(700);
        expect(result.riskDistribution.medium).toBe(250);
        expect(result.riskDistribution.high).toBe(50);
      });
    });
  });

  describe('⚠️ Error Handling', () => {
    describe('handleAssessmentErrors', () => {
      it('should handle database connection errors', async () => {
        // Arrange
        jest.spyOn(riskProfileRepository, 'findOne').mockRejectedValue(
          new Error('Database connection failed')
        );

        const request: RiskAssessmentRequest = {
          userId: 'user-123',
          tenantId: 'tenant-123',
          eventType: 'authentication',
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0',
          timestamp: new Date()
        };

        // Act & Assert
        await expect(service.assessAuthenticationRisk(request))
          .rejects.toThrow(InternalServerErrorException);
      });

      it('should handle invalid input data', async () => {
        // Arrange
        const invalidRequest = {
          userId: '',
          tenantId: '',
          eventType: 'invalid',
          ipAddress: 'invalid-ip'
        } as any;

        // Act & Assert
        await expect(service.assessAuthenticationRisk(invalidRequest))
          .rejects.toThrow(BadRequestException);
      });
    });
  });
});

// ========================================
// HELPER FUNCTIONS AND UTILITIES
// ========================================

/**
 * Helper function to create mock risk assessment request
 */
export function createMockRiskAssessmentRequest(overrides: Partial<RiskAssessmentRequest> = {}): RiskAssessmentRequest {
  return {
    userId: 'user-123',
    tenantId: 'tenant-123',
    eventType: 'authentication',
    ipAddress: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    deviceFingerprint: 'device-fingerprint-123',
    timestamp: new Date(),
    metadata: {},
    ...overrides
  };
}

/**
 * Helper function to create mock risk factors
 */
export function createMockRiskFactors() {
  return {
    device: {
      fingerprint: 'device-123',
      isKnown: true,
      userAgent: 'Mozilla/5.0',
      platform: 'Windows'
    },
    location: {
      ipAddress: '192.168.1.100',
      country: 'US',
      isKnown: true,
      isVpn: false
    },
    behavioral: {
      loginFrequency: 'normal',
      sessionDuration: 3600,
      velocityScore: 0.2
    },
    temporal: {
      currentHour: 10,
      isBusinessHours: true,
      dayOfWeek: 2
    }
  };
}