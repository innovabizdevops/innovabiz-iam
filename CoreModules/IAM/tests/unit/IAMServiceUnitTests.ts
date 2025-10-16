/**
 * 🧪 TESTES UNITÁRIOS - IAM SERVICE INNOVABIZ
 * Framework: Jest + NestJS Testing + Supertest
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST SP 800-63B, W3C WebAuthn Level 3, FIDO2 CTAP2.1
 * Coverage Target: >95% | Security: OWASP Testing Guide
 */

import { Test, TestingModule } from '@nestjs/testing';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Cache } from 'cache-manager';

import { IAMService } from '../../CoreModules/Backend/IAM/IAMService';
import { WebAuthnService } from '../../CoreModules/Backend/IAM/WebAuthnService';
import { CredentialService } from '../../CoreModules/Backend/IAM/CredentialService';
import { RiskAssessmentService } from '../../CoreModules/Backend/IAM/RiskAssessmentService';
import { AuditService } from '../../CoreModules/Backend/IAM/AuditService';

import { User } from '../../CoreModules/Backend/IAM/entities/User.entity';
import { Credential } from '../../CoreModules/Backend/IAM/entities/Credential.entity';
import { Session } from '../../CoreModules/Backend/IAM/entities/Session.entity';

import {
  CreateUserDto,
  WebAuthnRegistrationDto,
  RefreshTokenDto
} from '../../CoreModules/Backend/IAM/dto';

import {
  UnauthorizedException,
  BadRequestException,
  ConflictException,
  NotFoundException
} from '@nestjs/common';

// Mock das dependências externas
jest.mock('@simplewebauthn/server');
jest.mock('crypto', () => ({
  randomUUID: jest.fn(() => 'mock-uuid-12345'),
  randomBytes: jest.fn(() => Buffer.from('mock-random-bytes')),
  createHash: jest.fn(() => ({
    update: jest.fn().mockReturnThis(),
    digest: jest.fn(() => 'mock-hash')
  }))
}));

describe('🔐 IAM Service - Unit Tests', () => {
  let service: IAMService;
  let webAuthnService: WebAuthnService;
  let credentialService: CredentialService;
  let riskAssessmentService: RiskAssessmentService;
  let auditService: AuditService;
  let jwtService: JwtService;
  let cacheManager: Cache;
  let userRepository: Repository<User>;
  let sessionRepository: Repository<Session>;

  // Mock data
  const mockUser: User = {
    id: 'user-123',
    tenantId: 'tenant-123',
    email: 'test@innovabiz.com',
    username: 'testuser',
    displayName: 'Test User',
    isActive: true,
    isVerified: true,
    lastLoginAt: new Date(),
    createdAt: new Date(),
    updatedAt: new Date(),
    credentials: [],
    sessions: [],
    auditLogs: []
  };

  const mockCredential: Credential = {
    id: 'cred-123',
    userId: 'user-123',
    tenantId: 'tenant-123',
    credentialId: 'credential-id-123',
    publicKey: 'mock-public-key',
    counter: 0,
    deviceType: 'platform',
    transports: ['internal'],
    isActive: true,
    createdAt: new Date(),
    updatedAt: new Date(),
    user: mockUser
  };

  const mockSession: Session = {
    id: 'session-123',
    userId: 'user-123',
    tenantId: 'tenant-123',
    sessionToken: 'mock-session-token',
    refreshToken: 'mock-refresh-token',
    expiresAt: new Date(Date.now() + 3600000),
    isActive: true,
    ipAddress: '192.168.1.1',
    userAgent: 'Mozilla/5.0',
    createdAt: new Date(),
    updatedAt: new Date(),
    user: mockUser
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        IAMService,
        {
          provide: WebAuthnService,
          useValue: {
            generateRegistrationOptions: jest.fn(),
            generateAuthenticationOptions: jest.fn(),
            verifyRegistrationResponse: jest.fn(),
            verifyAuthenticationResponse: jest.fn()
          }
        },
        {
          provide: CredentialService,
          useValue: {
            createCredential: jest.fn(),
            getUserCredentials: jest.fn(),
            updateCredential: jest.fn(),
            deleteCredential: jest.fn()
          }
        },
        {
          provide: RiskAssessmentService,
          useValue: {
            assessRegistrationRisk: jest.fn(),
            assessAuthenticationRisk: jest.fn(),
            updateRiskProfile: jest.fn()
          }
        },
        {
          provide: AuditService,
          useValue: {
            logEvent: jest.fn(),
            logSecurityEvent: jest.fn(),
            logComplianceEvent: jest.fn()
          }
        },
        {
          provide: JwtService,
          useValue: {
            sign: jest.fn(),
            verify: jest.fn(),
            decode: jest.fn()
          }
        },
        {
          provide: ConfigService,
          useValue: {
            get: jest.fn((key: string) => {
              const config = {
                'jwt.secret': 'test-secret',
                'jwt.expiresIn': '1h',
                'jwt.refreshExpiresIn': '7d',
                'webauthn.rpName': 'INNOVABIZ',
                'webauthn.rpID': 'localhost',
                'webauthn.origin': 'http://localhost:3000'
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
            del: jest.fn(),
            reset: jest.fn()
          }
        },
        {
          provide: getRepositoryToken(User),
          useValue: {
            create: jest.fn(),
            save: jest.fn(),
            findOne: jest.fn(),
            findOneBy: jest.fn(),
            find: jest.fn(),
            update: jest.fn(),
            delete: jest.fn()
          }
        },
        {
          provide: getRepositoryToken(Session),
          useValue: {
            create: jest.fn(),
            save: jest.fn(),
            findOne: jest.fn(),
            find: jest.fn(),
            update: jest.fn(),
            delete: jest.fn()
          }
        }
      ]
    }).compile();

    service = module.get<IAMService>(IAMService);
    webAuthnService = module.get<WebAuthnService>(WebAuthnService);
    credentialService = module.get<CredentialService>(CredentialService);
    riskAssessmentService = module.get<RiskAssessmentService>(RiskAssessmentService);
    auditService = module.get<AuditService>(AuditService);
    jwtService = module.get<JwtService>(JwtService);
    cacheManager = module.get<Cache>(CACHE_MANAGER);
    userRepository = module.get<Repository<User>>(getRepositoryToken(User));
    sessionRepository = module.get<Repository<Session>>(getRepositoryToken(Session));
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('👤 User Management', () => {
    describe('createUser', () => {
      const createUserDto: CreateUserDto = {
        email: 'test@innovabiz.com',
        username: 'testuser',
        displayName: 'Test User',
        tenantId: 'tenant-123'
      };

      it('should create a new user successfully', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOneBy').mockResolvedValue(null);
        jest.spyOn(userRepository, 'create').mockReturnValue(mockUser);
        jest.spyOn(userRepository, 'save').mockResolvedValue(mockUser);
        jest.spyOn(auditService, 'logEvent').mockResolvedValue(undefined);

        // Act
        const result = await service.createUser(createUserDto);

        // Assert
        expect(result).toEqual(mockUser);
        expect(userRepository.findOneBy).toHaveBeenCalledWith({
          email: createUserDto.email,
          tenantId: createUserDto.tenantId
        });
        expect(userRepository.create).toHaveBeenCalledWith(createUserDto);
        expect(userRepository.save).toHaveBeenCalledWith(mockUser);
        expect(auditService.logEvent).toHaveBeenCalledWith({
          userId: mockUser.id,
          tenantId: mockUser.tenantId,
          action: 'USER_CREATED',
          resource: 'User',
          resourceId: mockUser.id,
          metadata: { email: mockUser.email }
        });
      });

      it('should throw ConflictException if user already exists', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOneBy').mockResolvedValue(mockUser);

        // Act & Assert
        await expect(service.createUser(createUserDto))
          .rejects.toThrow(ConflictException);
        
        expect(userRepository.findOneBy).toHaveBeenCalledWith({
          email: createUserDto.email,
          tenantId: createUserDto.tenantId
        });
        expect(userRepository.create).not.toHaveBeenCalled();
      });

      it('should validate email format', async () => {
        // Arrange
        const invalidEmailDto = { ...createUserDto, email: 'invalid-email' };

        // Act & Assert
        await expect(service.createUser(invalidEmailDto))
          .rejects.toThrow(BadRequestException);
      });
    });

    describe('getUserById', () => {
      it('should return user by id', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOne').mockResolvedValue(mockUser);

        // Act
        const result = await service.getUserById('user-123', 'tenant-123');

        // Assert
        expect(result).toEqual(mockUser);
        expect(userRepository.findOne).toHaveBeenCalledWith({
          where: { id: 'user-123', tenantId: 'tenant-123' },
          relations: ['credentials', 'sessions']
        });
      });

      it('should throw NotFoundException if user not found', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOne').mockResolvedValue(null);

        // Act & Assert
        await expect(service.getUserById('non-existent', 'tenant-123'))
          .rejects.toThrow(NotFoundException);
      });
    });
  });

  describe('🔐 WebAuthn Authentication', () => {
    describe('beginRegistration', () => {
      const registrationDto: WebAuthnRegistrationDto = {
        userId: 'user-123',
        username: 'testuser',
        displayName: 'Test User'
      };

      it('should generate registration options successfully', async () => {
        // Arrange
        const mockOptions = {
          challenge: 'mock-challenge',
          rp: { name: 'INNOVABIZ', id: 'localhost' },
          user: { id: 'user-123', name: 'testuser', displayName: 'Test User' },
          pubKeyCredParams: [{ alg: -7, type: 'public-key' }],
          timeout: 60000,
          excludeCredentials: []
        };

        jest.spyOn(userRepository, 'findOne').mockResolvedValue(mockUser);
        jest.spyOn(credentialService, 'getUserCredentials').mockResolvedValue([]);
        jest.spyOn(webAuthnService, 'generateRegistrationOptions').mockResolvedValue(mockOptions);
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);
        jest.spyOn(riskAssessmentService, 'assessRegistrationRisk').mockResolvedValue({
          riskScore: 25,
          riskLevel: 'low',
          factors: []
        });

        // Act
        const result = await service.beginRegistration(registrationDto, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        });

        // Assert
        expect(result).toEqual(mockOptions);
        expect(webAuthnService.generateRegistrationOptions).toHaveBeenCalledWith(
          'user-123',
          'tenant-123',
          expect.objectContaining({
            username: 'testuser',
            displayName: 'Test User'
          })
        );
        expect(cacheManager.set).toHaveBeenCalledWith(
          'webauthn:challenge:user-123',
          'mock-challenge',
          300000 // 5 minutes
        );
      });

      it('should throw NotFoundException if user not found', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOne').mockResolvedValue(null);

        // Act & Assert
        await expect(service.beginRegistration(registrationDto, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        })).rejects.toThrow(NotFoundException);
      });
    });

    describe('completeRegistration', () => {
      const registrationResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          attestationObject: 'mock-attestation',
          clientDataJSON: 'mock-client-data'
        },
        type: 'public-key' as const
      };

      it('should complete registration successfully', async () => {
        // Arrange
        const verificationResult = {
          verified: true,
          registrationInfo: {
            credentialID: Buffer.from('credential-id-123'),
            credentialPublicKey: Buffer.from('public-key'),
            counter: 0,
            credentialDeviceType: 'platform' as const,
            credentialBackedUp: false,
            origin: 'http://localhost:3000'
          }
        };

        jest.spyOn(userRepository, 'findOne').mockResolvedValue(mockUser);
        jest.spyOn(cacheManager, 'get').mockResolvedValue('mock-challenge');
        jest.spyOn(webAuthnService, 'verifyRegistrationResponse').mockResolvedValue(verificationResult);
        jest.spyOn(credentialService, 'createCredential').mockResolvedValue(mockCredential);
        jest.spyOn(cacheManager, 'del').mockResolvedValue(undefined);
        jest.spyOn(auditService, 'logSecurityEvent').mockResolvedValue(undefined);

        // Act
        const result = await service.completeRegistration('user-123', registrationResponse, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        });

        // Assert
        expect(result).toEqual({
          verified: true,
          credential: mockCredential
        });
        expect(webAuthnService.verifyRegistrationResponse).toHaveBeenCalledWith(
          registrationResponse,
          'mock-challenge',
          expect.any(Object)
        );
        expect(credentialService.createCredential).toHaveBeenCalled();
        expect(auditService.logSecurityEvent).toHaveBeenCalledWith({
          userId: 'user-123',
          tenantId: 'tenant-123',
          action: 'WEBAUTHN_REGISTRATION_SUCCESS',
          severity: 'info',
          metadata: expect.any(Object)
        });
      });

      it('should throw BadRequestException if challenge not found', async () => {
        // Arrange
        jest.spyOn(userRepository, 'findOne').mockResolvedValue(mockUser);
        jest.spyOn(cacheManager, 'get').mockResolvedValue(null);

        // Act & Assert
        await expect(service.completeRegistration('user-123', registrationResponse, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        })).rejects.toThrow(BadRequestException);
      });
    });
  });

  describe('🎫 Token Management', () => {
    describe('generateTokens', () => {
      it('should generate access and refresh tokens', async () => {
        // Arrange
        jest.spyOn(jwtService, 'sign')
          .mockReturnValueOnce('mock-access-token')
          .mockReturnValueOnce('mock-refresh-token');
        jest.spyOn(sessionRepository, 'create').mockReturnValue(mockSession);
        jest.spyOn(sessionRepository, 'save').mockResolvedValue(mockSession);

        // Act
        const result = await service.generateTokens(mockUser, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        });

        // Assert
        expect(result).toEqual({
          accessToken: 'mock-access-token',
          refreshToken: 'mock-refresh-token',
          expiresIn: 3600
        });
        expect(jwtService.sign).toHaveBeenCalledTimes(2);
        expect(sessionRepository.save).toHaveBeenCalled();
      });
    });

    describe('refreshTokens', () => {
      const refreshDto: RefreshTokenDto = {
        refreshToken: 'mock-refresh-token'
      };

      it('should refresh tokens successfully', async () => {
        // Arrange
        const decodedToken = { userId: 'user-123', sessionId: 'session-123' };
        jest.spyOn(jwtService, 'verify').mockReturnValue(decodedToken);
        jest.spyOn(sessionRepository, 'findOne').mockResolvedValue(mockSession);
        jest.spyOn(userRepository, 'findOne').mockResolvedValue(mockUser);
        jest.spyOn(service, 'generateTokens').mockResolvedValue({
          accessToken: 'new-access-token',
          refreshToken: 'new-refresh-token',
          expiresIn: 3600
        });

        // Act
        const result = await service.refreshTokens(refreshDto, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        });

        // Assert
        expect(result).toEqual({
          accessToken: 'new-access-token',
          refreshToken: 'new-refresh-token',
          expiresIn: 3600
        });
        expect(jwtService.verify).toHaveBeenCalledWith('mock-refresh-token');
      });

      it('should throw UnauthorizedException for invalid refresh token', async () => {
        // Arrange
        jest.spyOn(jwtService, 'verify').mockImplementation(() => {
          throw new Error('Invalid token');
        });

        // Act & Assert
        await expect(service.refreshTokens(refreshDto, {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0'
        })).rejects.toThrow(UnauthorizedException);
      });
    });
  });

  describe('🔒 Security Features', () => {
    describe('validateSession', () => {
      it('should validate active session successfully', async () => {
        // Arrange
        jest.spyOn(sessionRepository, 'findOne').mockResolvedValue(mockSession);

        // Act
        const result = await service.validateSession('mock-session-token');

        // Assert
        expect(result).toEqual(mockSession);
        expect(sessionRepository.findOne).toHaveBeenCalledWith({
          where: { 
            sessionToken: 'mock-session-token',
            isActive: true,
            expiresAt: expect.any(Object)
          },
          relations: ['user']
        });
      });

      it('should throw UnauthorizedException for expired session', async () => {
        // Arrange
        const expiredSession = { 
          ...mockSession, 
          expiresAt: new Date(Date.now() - 3600000) 
        };
        jest.spyOn(sessionRepository, 'findOne').mockResolvedValue(expiredSession);

        // Act & Assert
        await expect(service.validateSession('expired-token'))
          .rejects.toThrow(UnauthorizedException);
      });
    });

    describe('revokeSession', () => {
      it('should revoke session successfully', async () => {
        // Arrange
        jest.spyOn(sessionRepository, 'findOne').mockResolvedValue(mockSession);
        jest.spyOn(sessionRepository, 'save').mockResolvedValue({
          ...mockSession,
          isActive: false
        });
        jest.spyOn(auditService, 'logSecurityEvent').mockResolvedValue(undefined);

        // Act
        await service.revokeSession('session-123');

        // Assert
        expect(sessionRepository.save).toHaveBeenCalledWith({
          ...mockSession,
          isActive: false
        });
        expect(auditService.logSecurityEvent).toHaveBeenCalledWith({
          userId: mockSession.userId,
          tenantId: mockSession.tenantId,
          action: 'SESSION_REVOKED',
          severity: 'info',
          metadata: { sessionId: 'session-123' }
        });
      });
    });
  });

  describe('📊 Metrics and Monitoring', () => {
    describe('getUserMetrics', () => {
      it('should return user authentication metrics', async () => {
        // Arrange
        const mockMetrics = {
          totalLogins: 150,
          successfulLogins: 148,
          failedLogins: 2,
          lastLogin: new Date(),
          averageSessionDuration: 3600,
          deviceCount: 3,
          riskScore: 25
        };

        jest.spyOn(service, 'calculateUserMetrics').mockResolvedValue(mockMetrics);

        // Act
        const result = await service.getUserMetrics('user-123', 'tenant-123');

        // Assert
        expect(result).toEqual(mockMetrics);
        expect(service.calculateUserMetrics).toHaveBeenCalledWith('user-123', 'tenant-123');
      });
    });

    describe('getSystemMetrics', () => {
      it('should return system-wide metrics', async () => {
        // Arrange
        const mockSystemMetrics = {
          totalUsers: 10000,
          activeUsers: 8500,
          totalSessions: 15000,
          activeSessions: 2500,
          authenticationRate: 0.985,
          averageResponseTime: 150,
          errorRate: 0.002
        };

        jest.spyOn(service, 'calculateSystemMetrics').mockResolvedValue(mockSystemMetrics);

        // Act
        const result = await service.getSystemMetrics('tenant-123');

        // Assert
        expect(result).toEqual(mockSystemMetrics);
        expect(service.calculateSystemMetrics).toHaveBeenCalledWith('tenant-123');
      });
    });
  });
});

// ========================================
// HELPER FUNCTIONS AND UTILITIES
// ========================================

/**
 * Helper function to create mock WebAuthn registration response
 */
export function createMockRegistrationResponse(credentialId: string = 'test-credential') {
  return {
    id: credentialId,
    rawId: credentialId,
    response: {
      attestationObject: Buffer.from('mock-attestation-object').toString('base64'),
      clientDataJSON: Buffer.from(JSON.stringify({
        type: 'webauthn.create',
        challenge: 'mock-challenge',
        origin: 'http://localhost:3000'
      })).toString('base64')
    },
    type: 'public-key' as const
  };
}

/**
 * Helper function to create mock WebAuthn authentication response
 */
export function createMockAuthenticationResponse(credentialId: string = 'test-credential') {
  return {
    id: credentialId,
    rawId: credentialId,
    response: {
      authenticatorData: Buffer.from('mock-authenticator-data').toString('base64'),
      clientDataJSON: Buffer.from(JSON.stringify({
        type: 'webauthn.get',
        challenge: 'mock-auth-challenge',
        origin: 'http://localhost:3000'
      })).toString('base64'),
      signature: Buffer.from('mock-signature').toString('base64')
    },
    type: 'public-key' as const
  };
}

/**
 * Helper function to create mock user context
 */
export function createMockUserContext() {
  return {
    ipAddress: '192.168.1.1',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    timestamp: new Date(),
    sessionId: 'mock-session-id'
  };
}