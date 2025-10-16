/**
 * 🔐 TESTES UNITÁRIOS - WEBAUTHN SERVICE INNOVABIZ
 * Framework: Jest + @simplewebauthn/server
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Coverage Target: >95% | Security: OWASP Testing Guide
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigService } from '@nestjs/config';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';

import {
  generateRegistrationOptions,
  generateAuthenticationOptions,
  verifyRegistrationResponse,
  verifyAuthenticationResponse
} from '@simplewebauthn/server';

import { WebAuthnService } from '../../CoreModules/Backend/IAM/WebAuthnService';
import { CredentialService } from '../../CoreModules/Backend/IAM/CredentialService';
import { RiskAssessmentService } from '../../CoreModules/Backend/IAM/RiskAssessmentService';

import {
  RegistrationOptionsRequest,
  AuthenticationOptionsRequest,
  WebAuthnRegistrationResponse,
  WebAuthnAuthenticationResponse
} from '../../CoreModules/Backend/IAM/dto';

import {
  BadRequestException,
  UnauthorizedException,
  InternalServerErrorException
} from '@nestjs/common';

// Mock do @simplewebauthn/server
jest.mock('@simplewebauthn/server', () => ({
  generateRegistrationOptions: jest.fn(),
  generateAuthenticationOptions: jest.fn(),
  verifyRegistrationResponse: jest.fn(),
  verifyAuthenticationResponse: jest.fn()
}));

const mockGenerateRegistrationOptions = generateRegistrationOptions as jest.MockedFunction<typeof generateRegistrationOptions>;
const mockGenerateAuthenticationOptions = generateAuthenticationOptions as jest.MockedFunction<typeof generateAuthenticationOptions>;
const mockVerifyRegistrationResponse = verifyRegistrationResponse as jest.MockedFunction<typeof verifyRegistrationResponse>;
const mockVerifyAuthenticationResponse = verifyAuthenticationResponse as jest.MockedFunction<typeof verifyAuthenticationResponse>;

describe('🔐 WebAuthn Service - Unit Tests', () => {
  let service: WebAuthnService;
  let credentialService: CredentialService;
  let riskAssessmentService: RiskAssessmentService;
  let configService: ConfigService;
  let cacheManager: Cache;

  // Mock data
  const mockCredential = {
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
    updatedAt: new Date()
  };

  const mockWebAuthnConfig = {
    rpName: 'INNOVABIZ',
    rpID: 'localhost',
    origin: 'http://localhost:3000',
    timeout: 60000,
    attestation: 'none' as const,
    authenticatorSelection: {
      authenticatorAttachment: 'platform' as const,
      userVerification: 'required' as const,
      residentKey: 'preferred' as const
    }
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        WebAuthnService,
        {
          provide: CredentialService,
          useValue: {
            getUserCredentials: jest.fn(),
            createCredential: jest.fn(),
            updateCredential: jest.fn(),
            getCredentialById: jest.fn()
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
          provide: ConfigService,
          useValue: {
            get: jest.fn((key: string) => {
              const config = {
                'webauthn.rpName': 'INNOVABIZ',
                'webauthn.rpID': 'localhost',
                'webauthn.origin': 'http://localhost:3000',
                'webauthn.timeout': 60000,
                'webauthn.attestation': 'none',
                'webauthn.authenticatorSelection': mockWebAuthnConfig.authenticatorSelection
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
        }
      ]
    }).compile();

    service = module.get<WebAuthnService>(WebAuthnService);
    credentialService = module.get<CredentialService>(CredentialService);
    riskAssessmentService = module.get<RiskAssessmentService>(RiskAssessmentService);
    configService = module.get<ConfigService>(ConfigService);
    cacheManager = module.get<Cache>(CACHE_MANAGER);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('📝 Registration Flow', () => {
    describe('generateRegistrationOptions', () => {
      const registrationRequest: RegistrationOptionsRequest = {
        username: 'testuser',
        displayName: 'Test User',
        attestation: 'none',
        authenticatorSelection: {
          authenticatorAttachment: 'platform',
          userVerification: 'required',
          residentKey: 'preferred'
        }
      };

      it('should generate registration options successfully', async () => {
        // Arrange
        const mockOptions = {
          challenge: 'mock-challenge-12345',
          rp: { name: 'INNOVABIZ', id: 'localhost' },
          user: {
            id: Buffer.from('user-123'),
            name: 'testuser',
            displayName: 'Test User'
          },
          pubKeyCredParams: [
            { alg: -7, type: 'public-key' as const },
            { alg: -257, type: 'public-key' as const }
          ],
          timeout: 60000,
          attestation: 'none' as const,
          excludeCredentials: [],
          authenticatorSelection: mockWebAuthnConfig.authenticatorSelection
        };

        jest.spyOn(credentialService, 'getUserCredentials').mockResolvedValue([]);
        mockGenerateRegistrationOptions.mockResolvedValue(mockOptions);
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        const result = await service.generateRegistrationOptions(
          'user-123',
          'tenant-123',
          registrationRequest
        );

        // Assert
        expect(result).toEqual(mockOptions);
        expect(credentialService.getUserCredentials).toHaveBeenCalledWith('user-123', 'tenant-123');
        expect(mockGenerateRegistrationOptions).toHaveBeenCalledWith({
          rpName: 'INNOVABIZ',
          rpID: 'localhost',
          userID: Buffer.from('user-123'),
          userName: 'testuser',
          userDisplayName: 'Test User',
          timeout: 60000,
          attestationType: 'none',
          excludeCredentials: [],
          authenticatorSelection: expect.objectContaining({
            authenticatorAttachment: 'platform',
            userVerification: 'required',
            residentKey: 'preferred'
          }),
          supportedAlgorithmIDs: [-7, -257, -8, -37, -38, -39]
        });
        expect(cacheManager.set).toHaveBeenCalledWith(
          'webauthn:challenge:user-123:tenant-123',
          'mock-challenge-12345',
          300000
        );
      });

      it('should exclude existing credentials', async () => {
        // Arrange
        const existingCredentials = [mockCredential];
        const mockOptions = {
          challenge: 'mock-challenge',
          excludeCredentials: [{
            id: Buffer.from('credential-id-123', 'base64url'),
            type: 'public-key' as const,
            transports: ['internal'] as AuthenticatorTransport[]
          }]
        };

        jest.spyOn(credentialService, 'getUserCredentials').mockResolvedValue(existingCredentials);
        mockGenerateRegistrationOptions.mockResolvedValue(mockOptions);
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        await service.generateRegistrationOptions('user-123', 'tenant-123', registrationRequest);

        // Assert
        expect(mockGenerateRegistrationOptions).toHaveBeenCalledWith(
          expect.objectContaining({
            excludeCredentials: [{
              id: Buffer.from('credential-id-123', 'base64url'),
              type: 'public-key',
              transports: ['internal']
            }]
          })
        );
      });

      it('should handle credential service errors', async () => {
        // Arrange
        jest.spyOn(credentialService, 'getUserCredentials').mockRejectedValue(
          new Error('Database error')
        );

        // Act & Assert
        await expect(service.generateRegistrationOptions(
          'user-123',
          'tenant-123',
          registrationRequest
        )).rejects.toThrow(InternalServerErrorException);
      });
    });

    describe('verifyRegistrationResponse', () => {
      const registrationResponse: WebAuthnRegistrationResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          attestationObject: 'mock-attestation-object',
          clientDataJSON: 'mock-client-data-json'
        },
        type: 'public-key'
      };

      it('should verify registration response successfully', async () => {
        // Arrange
        const mockVerificationResult = {
          verified: true,
          registrationInfo: {
            credentialID: Buffer.from('credential-id-123'),
            credentialPublicKey: Buffer.from('mock-public-key'),
            counter: 0,
            credentialDeviceType: 'platform' as const,
            credentialBackedUp: false,
            origin: 'http://localhost:3000',
            rpID: 'localhost'
          }
        };

        mockVerifyRegistrationResponse.mockResolvedValue(mockVerificationResult);

        // Act
        const result = await service.verifyRegistrationResponse(
          registrationResponse,
          'mock-challenge',
          {
            rpID: 'localhost',
            expectedOrigin: 'http://localhost:3000'
          }
        );

        // Assert
        expect(result).toEqual(mockVerificationResult);
        expect(mockVerifyRegistrationResponse).toHaveBeenCalledWith({
          response: registrationResponse,
          expectedChallenge: 'mock-challenge',
          expectedOrigin: 'http://localhost:3000',
          expectedRPID: 'localhost',
          requireUserVerification: true
        });
      });

      it('should handle verification failure', async () => {
        // Arrange
        const mockVerificationResult = {
          verified: false,
          registrationInfo: undefined
        };

        mockVerifyRegistrationResponse.mockResolvedValue(mockVerificationResult);

        // Act
        const result = await service.verifyRegistrationResponse(
          registrationResponse,
          'mock-challenge',
          { rpID: 'localhost', expectedOrigin: 'http://localhost:3000' }
        );

        // Assert
        expect(result.verified).toBe(false);
      });

      it('should handle invalid response format', async () => {
        // Arrange
        const invalidResponse = { ...registrationResponse, response: null };

        // Act & Assert
        await expect(service.verifyRegistrationResponse(
          invalidResponse as any,
          'mock-challenge',
          { rpID: 'localhost', expectedOrigin: 'http://localhost:3000' }
        )).rejects.toThrow(BadRequestException);
      });
    });
  });

  describe('🔓 Authentication Flow', () => {
    describe('generateAuthenticationOptions', () => {
      const authRequest: AuthenticationOptionsRequest = {
        userVerification: 'required',
        allowCredentials: []
      };

      it('should generate authentication options successfully', async () => {
        // Arrange
        const mockOptions = {
          challenge: 'mock-auth-challenge',
          timeout: 60000,
          rpID: 'localhost',
          allowCredentials: [],
          userVerification: 'required' as const
        };

        jest.spyOn(credentialService, 'getUserCredentials').mockResolvedValue([mockCredential]);
        mockGenerateAuthenticationOptions.mockResolvedValue(mockOptions);
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        const result = await service.generateAuthenticationOptions(
          'user-123',
          'tenant-123',
          authRequest
        );

        // Assert
        expect(result).toEqual(mockOptions);
        expect(mockGenerateAuthenticationOptions).toHaveBeenCalledWith({
          timeout: 60000,
          allowCredentials: [{
            id: Buffer.from('credential-id-123', 'base64url'),
            type: 'public-key',
            transports: ['internal']
          }],
          userVerification: 'required',
          rpID: 'localhost'
        });
        expect(cacheManager.set).toHaveBeenCalledWith(
          'webauthn:auth-challenge:user-123:tenant-123',
          'mock-auth-challenge',
          300000
        );
      });

      it('should handle empty credential list', async () => {
        // Arrange
        jest.spyOn(credentialService, 'getUserCredentials').mockResolvedValue([]);
        mockGenerateAuthenticationOptions.mockResolvedValue({
          challenge: 'mock-challenge',
          allowCredentials: []
        });
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        const result = await service.generateAuthenticationOptions(
          'user-123',
          'tenant-123',
          authRequest
        );

        // Assert
        expect(result.allowCredentials).toEqual([]);
      });
    });

    describe('verifyAuthenticationResponse', () => {
      const authResponse: WebAuthnAuthenticationResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          authenticatorData: 'mock-authenticator-data',
          clientDataJSON: 'mock-client-data',
          signature: 'mock-signature'
        },
        type: 'public-key'
      };

      it('should verify authentication response successfully', async () => {
        // Arrange
        const mockVerificationResult = {
          verified: true,
          authenticationInfo: {
            newCounter: 1,
            credentialID: Buffer.from('credential-id-123')
          }
        };

        jest.spyOn(credentialService, 'getCredentialById').mockResolvedValue(mockCredential);
        mockVerifyAuthenticationResponse.mockResolvedValue(mockVerificationResult);
        jest.spyOn(credentialService, 'updateCredential').mockResolvedValue(undefined);

        // Act
        const result = await service.verifyAuthenticationResponse(
          authResponse,
          'mock-auth-challenge',
          'user-123',
          'tenant-123',
          {
            rpID: 'localhost',
            expectedOrigin: 'http://localhost:3000'
          }
        );

        // Assert
        expect(result).toEqual(mockVerificationResult);
        expect(mockVerifyAuthenticationResponse).toHaveBeenCalledWith({
          response: authResponse,
          expectedChallenge: 'mock-auth-challenge',
          expectedOrigin: 'http://localhost:3000',
          expectedRPID: 'localhost',
          authenticator: {
            credentialID: Buffer.from('credential-id-123', 'base64url'),
            credentialPublicKey: Buffer.from('mock-public-key', 'base64'),
            counter: 0,
            transports: ['internal']
          },
          requireUserVerification: true
        });
        expect(credentialService.updateCredential).toHaveBeenCalledWith(
          'credential-id-123',
          { counter: 1 }
        );
      });

      it('should throw error if credential not found', async () => {
        // Arrange
        jest.spyOn(credentialService, 'getCredentialById').mockResolvedValue(null);

        // Act & Assert
        await expect(service.verifyAuthenticationResponse(
          authResponse,
          'mock-challenge',
          'user-123',
          'tenant-123',
          { rpID: 'localhost', expectedOrigin: 'http://localhost:3000' }
        )).rejects.toThrow(UnauthorizedException);
      });

      it('should handle counter rollback attack', async () => {
        // Arrange
        const mockVerificationResult = {
          verified: true,
          authenticationInfo: {
            newCounter: 0, // Counter didn't increase - potential rollback attack
            credentialID: Buffer.from('credential-id-123')
          }
        };

        const credentialWithHigherCounter = {
          ...mockCredential,
          counter: 5 // Higher than newCounter
        };

        jest.spyOn(credentialService, 'getCredentialById').mockResolvedValue(credentialWithHigherCounter);
        mockVerifyAuthenticationResponse.mockResolvedValue(mockVerificationResult);

        // Act & Assert
        await expect(service.verifyAuthenticationResponse(
          authResponse,
          'mock-challenge',
          'user-123',
          'tenant-123',
          { rpID: 'localhost', expectedOrigin: 'http://localhost:3000' }
        )).rejects.toThrow(UnauthorizedException);
      });
    });
  });

  describe('🔒 Security Features', () => {
    describe('storeChallenge', () => {
      it('should store challenge with expiration', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'set').mockResolvedValue(undefined);

        // Act
        await service.storeChallenge('user-123', 'tenant-123', 'mock-challenge');

        // Assert
        expect(cacheManager.set).toHaveBeenCalledWith(
          'webauthn:challenge:user-123:tenant-123',
          'mock-challenge',
          300000 // 5 minutes
        );
      });
    });

    describe('getStoredChallenge', () => {
      it('should retrieve stored challenge', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'get').mockResolvedValue('stored-challenge');

        // Act
        const result = await service.getStoredChallenge('user-123', 'tenant-123');

        // Assert
        expect(result).toBe('stored-challenge');
        expect(cacheManager.get).toHaveBeenCalledWith(
          'webauthn:challenge:user-123:tenant-123'
        );
      });

      it('should return null if challenge not found', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'get').mockResolvedValue(null);

        // Act
        const result = await service.getStoredChallenge('user-123', 'tenant-123');

        // Assert
        expect(result).toBeNull();
      });
    });

    describe('clearChallenge', () => {
      it('should clear stored challenge', async () => {
        // Arrange
        jest.spyOn(cacheManager, 'del').mockResolvedValue(undefined);

        // Act
        await service.clearChallenge('user-123', 'tenant-123');

        // Assert
        expect(cacheManager.del).toHaveBeenCalledWith(
          'webauthn:challenge:user-123:tenant-123'
        );
      });
    });
  });

  describe('🛡️ Risk Assessment Integration', () => {
    describe('assessRegistrationRisk', () => {
      it('should assess registration risk', async () => {
        // Arrange
        const mockRiskAssessment = {
          riskScore: 25,
          riskLevel: 'low' as const,
          factors: ['new_device', 'known_location']
        };

        jest.spyOn(riskAssessmentService, 'assessRegistrationRisk').mockResolvedValue(mockRiskAssessment);

        // Act
        const result = await service.assessRegistrationRisk('user-123', 'tenant-123', {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0',
          deviceFingerprint: 'mock-fingerprint'
        });

        // Assert
        expect(result).toEqual(mockRiskAssessment);
        expect(riskAssessmentService.assessRegistrationRisk).toHaveBeenCalledWith(
          'user-123',
          'tenant-123',
          expect.objectContaining({
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0',
            deviceFingerprint: 'mock-fingerprint'
          })
        );
      });
    });

    describe('assessAuthenticationRisk', () => {
      it('should assess authentication risk', async () => {
        // Arrange
        const mockRiskAssessment = {
          riskScore: 15,
          riskLevel: 'low' as const,
          factors: ['known_device', 'normal_time']
        };

        jest.spyOn(riskAssessmentService, 'assessAuthenticationRisk').mockResolvedValue(mockRiskAssessment);

        // Act
        const result = await service.assessAuthenticationRisk('user-123', 'tenant-123', {
          ipAddress: '192.168.1.1',
          userAgent: 'Mozilla/5.0',
          credentialId: 'credential-id-123'
        });

        // Assert
        expect(result).toEqual(mockRiskAssessment);
        expect(riskAssessmentService.assessAuthenticationRisk).toHaveBeenCalledWith(
          'user-123',
          'tenant-123',
          expect.objectContaining({
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0',
            credentialId: 'credential-id-123'
          })
        );
      });
    });
  });

  describe('⚙️ Configuration Management', () => {
    describe('getWebAuthnConfig', () => {
      it('should return WebAuthn configuration', () => {
        // Act
        const config = service.getWebAuthnConfig();

        // Assert
        expect(config).toEqual({
          rpName: 'INNOVABIZ',
          rpID: 'localhost',
          origin: 'http://localhost:3000',
          timeout: 60000,
          attestation: 'none',
          authenticatorSelection: mockWebAuthnConfig.authenticatorSelection
        });
      });
    });

    describe('validateConfiguration', () => {
      it('should validate configuration successfully', () => {
        // Act & Assert
        expect(() => service.validateConfiguration()).not.toThrow();
      });

      it('should throw error for invalid configuration', () => {
        // Arrange
        jest.spyOn(configService, 'get').mockReturnValue(null);

        // Act & Assert
        expect(() => service.validateConfiguration()).toThrow(InternalServerErrorException);
      });
    });
  });
});

// ========================================
// HELPER FUNCTIONS AND UTILITIES
// ========================================

/**
 * Helper function to create mock WebAuthn registration options
 */
export function createMockRegistrationOptions() {
  return {
    challenge: 'mock-challenge-12345',
    rp: { name: 'INNOVABIZ', id: 'localhost' },
    user: {
      id: Buffer.from('user-123'),
      name: 'testuser',
      displayName: 'Test User'
    },
    pubKeyCredParams: [
      { alg: -7, type: 'public-key' as const },
      { alg: -257, type: 'public-key' as const }
    ],
    timeout: 60000,
    attestation: 'none' as const,
    excludeCredentials: [],
    authenticatorSelection: {
      authenticatorAttachment: 'platform' as const,
      userVerification: 'required' as const,
      residentKey: 'preferred' as const
    }
  };
}

/**
 * Helper function to create mock WebAuthn authentication options
 */
export function createMockAuthenticationOptions() {
  return {
    challenge: 'mock-auth-challenge-12345',
    timeout: 60000,
    rpID: 'localhost',
    allowCredentials: [{
      id: Buffer.from('credential-id-123', 'base64url'),
      type: 'public-key' as const,
      transports: ['internal'] as AuthenticatorTransport[]
    }],
    userVerification: 'required' as const
  };
}