// Unit Tests for WebAuthnService
// Testes unitários completos seguindo padrões INNOVABIZ

import { WebAuthnService } from '../../src/services/WebAuthnService';
import { CredentialService } from '../../src/services/CredentialService';
import { RiskAssessmentService } from '../../src/services/RiskAssessmentService';
import { AuditService } from '../../src/services/AuditService';
import { AttestationService } from '../../src/services/AttestationService';
import { 
  RegistrationOptions, 
  AuthenticationOptions,
  VerifiedRegistrationResponse,
  VerifiedAuthenticationResponse
} from '../../src/types/webauthn';

// Mocks
jest.mock('../../src/services/CredentialService');
jest.mock('../../src/services/RiskAssessmentService');
jest.mock('../../src/services/AuditService');
jest.mock('../../src/services/AttestationService');

describe('WebAuthnService', () => {
  let webAuthnService: WebAuthnService;
  let mockCredentialService: jest.Mocked<CredentialService>;
  let mockRiskService: jest.Mocked<RiskAssessmentService>;
  let mockAuditService: jest.Mocked<AuditService>;
  let mockAttestationService: jest.Mocked<AttestationService>;

  const mockTenantId = 'tenant-123';
  const mockUserId = 'user-456';
  const mockCorrelationId = 'corr-789';

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Create mocked instances
    mockCredentialService = new CredentialService() as jest.Mocked<CredentialService>;
    mockRiskService = new RiskAssessmentService() as jest.Mocked<RiskAssessmentService>;
    mockAuditService = new AuditService() as jest.Mocked<AuditService>;
    mockAttestationService = new AttestationService() as jest.Mocked<AttestationService>;

    // Create service instance
    webAuthnService = new WebAuthnService(
      mockCredentialService,
      mockRiskService,
      mockAuditService,
      mockAttestationService
    );
  });

  describe('generateRegistrationOptions', () => {
    it('should generate valid registration options', async () => {
      // Arrange
      const mockExistingCredentials = [
        { credentialId: 'existing-cred-1', transports: ['usb'] },
        { credentialId: 'existing-cred-2', transports: ['nfc'] }
      ];

      mockCredentialService.getUserCredentials.mockResolvedValue(mockExistingCredentials);
      mockRiskService.assessRegistrationRisk.mockResolvedValue({
        score: 0.2,
        level: 'low',
        factors: ['new_device'],
        recommendations: []
      });

      // Act
      const result = await webAuthnService.generateRegistrationOptions(
        mockTenantId,
        mockUserId,
        'test@innovabiz.com',
        'Test User',
        mockCorrelationId
      );

      // Assert
      expect(result).toBeDefined();
      expect(result.rp).toEqual({
        name: 'INNOVABIZ',
        id: 'innovabiz.com'
      });
      expect(result.user).toEqual({
        id: expect.any(String),
        name: 'test@innovabiz.com',
        displayName: 'Test User'
      });
      expect(result.challenge).toBeDefined();
      expect(result.pubKeyCredParams).toHaveLength(2);
      expect(result.excludeCredentials).toHaveLength(2);
      expect(result.authenticatorSelection).toEqual({
        authenticatorAttachment: 'platform',
        userVerification: 'required',
        residentKey: 'preferred'
      });

      // Verify service calls
      expect(mockCredentialService.getUserCredentials).toHaveBeenCalledWith(
        mockUserId,
        mockTenantId
      );
      expect(mockRiskService.assessRegistrationRisk).toHaveBeenCalled();
      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_registration_options_generated',
          tenantId: mockTenantId,
          userId: mockUserId
        })
      );
    });

    it('should handle high risk registration', async () => {
      // Arrange
      mockCredentialService.getUserCredentials.mockResolvedValue([]);
      mockRiskService.assessRegistrationRisk.mockResolvedValue({
        score: 0.8,
        level: 'high',
        factors: ['suspicious_location', 'new_device'],
        recommendations: ['require_additional_verification']
      });

      // Act & Assert
      await expect(
        webAuthnService.generateRegistrationOptions(
          mockTenantId,
          mockUserId,
          'test@innovabiz.com',
          'Test User',
          mockCorrelationId
        )
      ).rejects.toThrow('Registration blocked due to high risk');
    });

    it('should handle service errors gracefully', async () => {
      // Arrange
      mockCredentialService.getUserCredentials.mockRejectedValue(
        new Error('Database connection failed')
      );

      // Act & Assert
      await expect(
        webAuthnService.generateRegistrationOptions(
          mockTenantId,
          mockUserId,
          'test@innovabiz.com',
          'Test User',
          mockCorrelationId
        )
      ).rejects.toThrow('Database connection failed');

      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_registration_options_error'
        })
      );
    });
  });

  describe('verifyRegistration', () => {
    it('should verify valid registration response', async () => {
      // Arrange
      const mockRegistrationResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          clientDataJSON: 'mock-client-data',
          attestationObject: 'mock-attestation'
        },
        type: 'public-key'
      };

      const mockChallenge = 'mock-challenge-123';
      const mockOrigin = 'https://app.innovabiz.com';

      mockAttestationService.verifyAttestation.mockResolvedValue({
        verified: true,
        attestationInfo: {
          fmt: 'packed',
          counter: 0,
          credentialPublicKey: new Uint8Array([1, 2, 3]),
          credentialID: new Uint8Array([4, 5, 6])
        }
      });

      mockCredentialService.saveCredential.mockResolvedValue({
        id: 'db-credential-id',
        credentialId: 'credential-id-123',
        userId: mockUserId,
        tenantId: mockTenantId,
        publicKey: new Uint8Array([1, 2, 3]),
        counter: 0,
        deviceType: 'platform',
        createdAt: new Date(),
        lastUsedAt: null
      });

      // Act
      const result = await webAuthnService.verifyRegistration(
        mockTenantId,
        mockUserId,
        mockRegistrationResponse,
        mockChallenge,
        mockOrigin,
        mockCorrelationId
      );

      // Assert
      expect(result.verified).toBe(true);
      expect(result.registrationInfo).toBeDefined();
      expect(mockAttestationService.verifyAttestation).toHaveBeenCalled();
      expect(mockCredentialService.saveCredential).toHaveBeenCalled();
      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_registration_verified'
        })
      );
    });

    it('should reject invalid attestation', async () => {
      // Arrange
      const mockRegistrationResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          clientDataJSON: 'mock-client-data',
          attestationObject: 'mock-attestation'
        },
        type: 'public-key'
      };

      mockAttestationService.verifyAttestation.mockResolvedValue({
        verified: false,
        attestationInfo: null
      });

      // Act
      const result = await webAuthnService.verifyRegistration(
        mockTenantId,
        mockUserId,
        mockRegistrationResponse,
        'mock-challenge',
        'https://app.innovabiz.com',
        mockCorrelationId
      );

      // Assert
      expect(result.verified).toBe(false);
      expect(mockCredentialService.saveCredential).not.toHaveBeenCalled();
      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_registration_failed'
        })
      );
    });
  });

  describe('generateAuthenticationOptions', () => {
    it('should generate valid authentication options', async () => {
      // Arrange
      const mockCredentials = [
        {
          credentialId: 'cred-1',
          transports: ['usb', 'nfc'],
          deviceType: 'cross-platform'
        },
        {
          credentialId: 'cred-2',
          transports: ['internal'],
          deviceType: 'platform'
        }
      ];

      mockCredentialService.getUserCredentials.mockResolvedValue(mockCredentials);
      mockRiskService.assessAuthenticationRisk.mockResolvedValue({
        score: 0.1,
        level: 'low',
        factors: [],
        recommendations: []
      });

      // Act
      const result = await webAuthnService.generateAuthenticationOptions(
        mockTenantId,
        mockUserId,
        mockCorrelationId
      );

      // Assert
      expect(result).toBeDefined();
      expect(result.challenge).toBeDefined();
      expect(result.allowCredentials).toHaveLength(2);
      expect(result.userVerification).toBe('required');
      expect(result.rpId).toBe('innovabiz.com');

      expect(mockCredentialService.getUserCredentials).toHaveBeenCalledWith(
        mockUserId,
        mockTenantId
      );
      expect(mockRiskService.assessAuthenticationRisk).toHaveBeenCalled();
    });

    it('should handle user with no credentials', async () => {
      // Arrange
      mockCredentialService.getUserCredentials.mockResolvedValue([]);

      // Act & Assert
      await expect(
        webAuthnService.generateAuthenticationOptions(
          mockTenantId,
          mockUserId,
          mockCorrelationId
        )
      ).rejects.toThrow('No credentials found for user');
    });
  });

  describe('verifyAuthentication', () => {
    it('should verify valid authentication response', async () => {
      // Arrange
      const mockAuthResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          clientDataJSON: 'mock-client-data',
          authenticatorData: 'mock-auth-data',
          signature: 'mock-signature'
        },
        type: 'public-key'
      };

      const mockCredential = {
        id: 'db-cred-id',
        credentialId: 'credential-id-123',
        publicKey: new Uint8Array([1, 2, 3]),
        counter: 5,
        userId: mockUserId,
        tenantId: mockTenantId
      };

      mockCredentialService.getCredentialById.mockResolvedValue(mockCredential);
      mockCredentialService.updateCredentialCounter.mockResolvedValue(undefined);

      // Mock crypto verification (would normally use @simplewebauthn/server)
      jest.spyOn(webAuthnService as any, 'verifyAuthenticationSignature')
        .mockResolvedValue({ verified: true, authenticationInfo: { newCounter: 6 } });

      // Act
      const result = await webAuthnService.verifyAuthentication(
        mockTenantId,
        mockUserId,
        mockAuthResponse,
        'mock-challenge',
        'https://app.innovabiz.com',
        mockCorrelationId
      );

      // Assert
      expect(result.verified).toBe(true);
      expect(mockCredentialService.getCredentialById).toHaveBeenCalledWith(
        'credential-id-123',
        mockTenantId
      );
      expect(mockCredentialService.updateCredentialCounter).toHaveBeenCalledWith(
        'credential-id-123',
        mockTenantId,
        6
      );
      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_authentication_verified'
        })
      );
    });

    it('should reject authentication with invalid signature', async () => {
      // Arrange
      const mockAuthResponse = {
        id: 'credential-id-123',
        rawId: 'credential-id-123',
        response: {
          clientDataJSON: 'mock-client-data',
          authenticatorData: 'mock-auth-data',
          signature: 'invalid-signature'
        },
        type: 'public-key'
      };

      const mockCredential = {
        id: 'db-cred-id',
        credentialId: 'credential-id-123',
        publicKey: new Uint8Array([1, 2, 3]),
        counter: 5,
        userId: mockUserId,
        tenantId: mockTenantId
      };

      mockCredentialService.getCredentialById.mockResolvedValue(mockCredential);

      // Mock failed crypto verification
      jest.spyOn(webAuthnService as any, 'verifyAuthenticationSignature')
        .mockResolvedValue({ verified: false });

      // Act
      const result = await webAuthnService.verifyAuthentication(
        mockTenantId,
        mockUserId,
        mockAuthResponse,
        'mock-challenge',
        'https://app.innovabiz.com',
        mockCorrelationId
      );

      // Assert
      expect(result.verified).toBe(false);
      expect(mockCredentialService.updateCredentialCounter).not.toHaveBeenCalled();
      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_authentication_failed'
        })
      );
    });
  });

  describe('Error Handling', () => {
    it('should handle database connection errors', async () => {
      // Arrange
      mockCredentialService.getUserCredentials.mockRejectedValue(
        new Error('Connection timeout')
      );

      // Act & Assert
      await expect(
        webAuthnService.generateAuthenticationOptions(
          mockTenantId,
          mockUserId,
          mockCorrelationId
        )
      ).rejects.toThrow('Connection timeout');

      expect(mockAuditService.logEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          eventType: 'webauthn_authentication_options_error'
        })
      );
    });

    it('should handle invalid input parameters', async () => {
      // Act & Assert
      await expect(
        webAuthnService.generateRegistrationOptions(
          '', // empty tenant ID
          mockUserId,
          'test@innovabiz.com',
          'Test User',
          mockCorrelationId
        )
      ).rejects.toThrow('Tenant ID is required');

      await expect(
        webAuthnService.generateRegistrationOptions(
          mockTenantId,
          '', // empty user ID
          'test@innovabiz.com',
          'Test User',
          mockCorrelationId
        )
      ).rejects.toThrow('User ID is required');
    });
  });

  describe('Performance Tests', () => {
    it('should complete registration options generation within SLA', async () => {
      // Arrange
      mockCredentialService.getUserCredentials.mockResolvedValue([]);
      mockRiskService.assessRegistrationRisk.mockResolvedValue({
        score: 0.1,
        level: 'low',
        factors: [],
        recommendations: []
      });

      const startTime = Date.now();

      // Act
      await webAuthnService.generateRegistrationOptions(
        mockTenantId,
        mockUserId,
        'test@innovabiz.com',
        'Test User',
        mockCorrelationId
      );

      const endTime = Date.now();
      const duration = endTime - startTime;

      // Assert - Should complete within 500ms SLA
      expect(duration).toBeLessThan(500);
    });
  });
});