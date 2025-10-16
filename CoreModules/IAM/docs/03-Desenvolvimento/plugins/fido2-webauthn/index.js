/**
 * FIDO2/WebAuthn Authentication Provider
 * 
 * This plugin implements FIDO2/WebAuthn authentication according to the W3C WebAuthn Level 2 specification.
 * It provides strong phishing-resistant authentication using platform and roaming authenticators.
 * 
 * @module fido2-webauthn
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { v4: uuidv4 } = require('uuid');
const base64url = require('base64url');
const { AuthenticationProvider, AuthMethodCategory, AssuranceLevel } = require('@innovabiz/auth-framework');
const CryptoService = require('@innovabiz/auth-framework/services/crypto');
const SessionManager = require('@innovabiz/auth-framework/services/session');
const AuditService = require('@innovabiz/auth-framework/services/audit');
const WebAuthnService = require('./lib/webauthn-service');
const CredentialRepository = require('./lib/credential-repository');
const AttestationService = require('./lib/attestation-service');
const { logger } = require('@innovabiz/auth-framework/utils/logging');
const RegionalAdapter = require('./lib/regional-adapter');

/**
 * FIDO2/WebAuthn Authentication Provider implementation
 */
class Fido2WebAuthnProvider extends AuthenticationProvider {
  /**
   * Unique identifier for this provider
   */
  get id() {
    return 'fido2-webauthn';
  }
  
  /**
   * Provider metadata
   */
  get metadata() {
    return {
      name: 'FIDO2/WebAuthn Authentication',
      description: 'Provides strong phishing-resistant authentication using FIDO2/WebAuthn standard',
      version: '1.0.0',
      category: AuthMethodCategory.MULTIFACTOR,
      assuranceLevel: AssuranceLevel.HIGH,
      capabilities: {
        supportsPasswordless: true,
        supportsFederatedLogin: false,
        supportsCrossPlatform: true,
        supportsOfflineMode: false,
        supportsSilentAuthentication: false,
        requiresUserInteraction: true,
        isPhishingResistant: true,
        supportsBiometrics: true,
        supportsHardwareTokens: true,
        supportsMobileDevices: true,
        supportsUserVerification: true
      }
    };
  }
  
  /**
   * Initializes the provider with configuration
   * 
   * @param {Object} config Provider configuration
   * @returns {Promise<void>}
   */
  async initialize(config) {
    this.config = this.validateConfig(config);
    
    // Initialize services
    this.cryptoService = new CryptoService();
    this.sessionManager = new SessionManager();
    this.auditService = new AuditService();
    this.credentialRepository = new CredentialRepository();
    
    // Initialize WebAuthn service
    this.webAuthnService = new WebAuthnService({
      rpId: this.config.rpId,
      rpName: this.config.rpName,
      origin: this.config.origin,
      timeout: this.config.timeout,
      userVerification: this.config.userVerification,
      attestation: this.config.attestation,
      authenticatorAttachment: this.config.authenticatorAttachment,
      challengeTtl: this.config.challengeTtl,
      supportedAlgorithms: this.config.supportedAlgorithms,
      requireResidentKey: this.config.requireResidentKey
    });
    
    // Initialize attestation service if configured
    if (this.config.metadataServiceConfig?.enabled) {
      this.attestationService = new AttestationService(this.config.metadataServiceConfig);
      await this.attestationService.initialize();
    }
    
    // Initialize regional adapter
    this.regionalAdapter = new RegionalAdapter(this.config.regionalSettings);
    
    logger.info(`[Fido2WebAuthnProvider] Initialized with rpId=${this.config.rpId}`);
  }
  
  /**
   * Validates provider configuration
   * 
   * @param {Object} config Configuration to validate
   * @returns {Object} Validated configuration
   * @throws {Error} If configuration is invalid
   */
  validateConfig(config) {
    // Required parameters
    if (!config.rpId) {
      throw new Error('rpId is required');
    }
    
    if (!config.rpName) {
      throw new Error('rpName is required');
    }
    
    if (!config.origin) {
      throw new Error('origin is required');
    }
    
    // Default values for optional parameters
    return {
      ...config,
      timeout: config.timeout || 60000,
      userVerification: config.userVerification || 'preferred',
      attestation: config.attestation || 'none',
      authenticatorAttachment: config.authenticatorAttachment || null,
      challengeTtl: config.challengeTtl || 300,
      maxAllowedCredentials: config.maxAllowedCredentials || 10,
      supportedAlgorithms: config.supportedAlgorithms || [-7, -257], // ES256, RS256
      requireResidentKey: config.requireResidentKey || false,
      regionalSettings: config.regionalSettings || {}
    };
  }
  
  /**
   * Starts the authentication process
   * 
   * @param {Object} context Authentication context
   * @returns {Promise<Object>} Authentication challenge
   */
  async startAuthentication(context) {
    logger.debug(`[Fido2WebAuthnProvider] Starting authentication for user ${context.userId}`);
    
    try {
      // Create a session ID for this authentication attempt
      const sessionId = uuidv4();
      
      // Apply regional adaptations based on context
      const adaptedContext = await this.regionalAdapter.adaptAuthenticationContext(context);
      
      // Get user's registered credentials
      let userCredentials = [];
      if (context.userId) {
        userCredentials = await this.credentialRepository.getCredentialsByUserId(
          context.userId,
          context.tenantId
        );
      }
      
      // Generate authentication options
      const options = await this.webAuthnService.generateAuthenticationOptions({
        userCredentials,
        userVerification: adaptedContext.userVerification || this.config.userVerification,
        timeout: this.config.timeout,
        allowCredentials: userCredentials.map(cred => ({
          id: cred.credentialId,
          type: 'public-key',
          transports: cred.transports || []
        }))
      });
      
      // Store challenge for later verification
      await this.sessionManager.setAuthenticationChallenge(
        sessionId,
        options.challenge,
        context.userId,
        {
          rpId: this.config.rpId,
          origin: context.origin || this.config.origin,
          userVerification: options.userVerification,
          allowCredentials: options.allowCredentials?.map(c => c.id) || []
        },
        this.config.challengeTtl
      );
      
      // Audit authentication start
      await this.auditService.logAuthEvent({
        eventType: 'authentication:started',
        providerId: this.id,
        userId: context.userId,
        tenantId: context.tenantId,
        sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: true
      });
      
      // Return authentication challenge
      return {
        sessionId,
        challengeType: 'webauthn.get',
        challenge: options,
        expiresAt: new Date(Date.now() + this.config.timeout),
        uiOptions: {
          title: 'FIDO2 Authentication',
          message: 'Please complete authentication with your security key or biometric sensor',
          uiExtension: 'fido2-auth-ui'
        }
      };
    } catch (error) {
      logger.error(`[Fido2WebAuthnProvider] Error starting authentication: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Verifies the authentication response
   * 
   * @param {Object} response Authentication response
   * @param {Object} context Authentication context
   * @returns {Promise<Object>} Authentication result
   */
  async verifyResponse(response, context) {
    logger.debug(`[Fido2WebAuthnProvider] Verifying authentication response for session ${response.sessionId}`);
    
    try {
      // Get stored challenge data
      const challengeData = await this.sessionManager.getAuthenticationChallenge(response.sessionId);
      
      if (!challengeData) {
        throw new Error('Authentication challenge not found or expired');
      }
      
      // Get user's credential
      const credential = await this.getCredentialForAssertion(
        response.response,
        challengeData.userId,
        context.tenantId
      );
      
      if (!credential) {
        throw new Error('Credential not found');
      }
      
      // Prepare assertion for verification
      const assertion = {
        id: response.response.id,
        rawId: base64url.toBuffer(response.response.rawId),
        response: {
          authenticatorData: base64url.toBuffer(response.response.response.authenticatorData),
          clientDataJSON: base64url.toBuffer(response.response.response.clientDataJSON),
          signature: base64url.toBuffer(response.response.response.signature),
          userHandle: response.response.response.userHandle ? 
            base64url.toBuffer(response.response.response.userHandle) : null
        },
        type: response.response.type
      };
      
      // Verify assertion
      const verificationResult = await this.webAuthnService.verifyAssertion({
        credential,
        assertion,
        expectedChallenge: challengeData.challenge,
        expectedOrigin: challengeData.origin || this.config.origin,
        expectedRpId: this.config.rpId,
        userVerification: challengeData.userVerification
      });
      
      if (!verificationResult.success) {
        throw new Error(verificationResult.error || 'Authentication verification failed');
      }
      
      // Update credential counter
      await this.credentialRepository.updateCredentialCounter(
        credential.id, 
        verificationResult.counter,
        context.tenantId
      );
      
      // Audit successful authentication
      await this.auditService.logAuthEvent({
        eventType: 'authentication:completed',
        providerId: this.id,
        userId: challengeData.userId,
        tenantId: context.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: true,
        credentialId: credential.id,
        authenticatorData: {
          userVerified: verificationResult.userVerified,
          authenticatorAttachment: credential.authenticatorAttachment
        }
      });
      
      // Return authentication result
      return {
        success: true,
        userId: challengeData.userId,
        authTime: new Date(),
        expiresAt: new Date(Date.now() + 3600 * 1000), // 1 hour
        amr: ['fido2.webauthn'],
        acr: verificationResult.userVerified ? 'urn:innovabiz:ac:classes:fido2:userverified' : 'urn:innovabiz:ac:classes:fido2',
        sessionId: response.sessionId,
        identityAttributes: {
          credentialId: credential.id,
          credentialName: credential.name,
          authenticatorAttachment: credential.authenticatorAttachment
        }
      };
    } catch (error) {
      logger.error(`[Fido2WebAuthnProvider] Error verifying authentication: ${error.message}`);
      
      // Audit failed authentication
      await this.auditService.logAuthEvent({
        eventType: 'authentication:failed',
        providerId: this.id,
        userId: context.userId,
        tenantId: context.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: false,
        error: error.message
      });
      
      return {
        success: false,
        error: {
          code: 'auth_verification_failed',
          message: 'Authentication verification failed',
          details: error.message
        }
      };
    }
  }
  
  /**
   * Gets the credential for assertion verification
   * 
   * @param {Object} response Authentication response
   * @param {string} userId User ID
   * @param {string} tenantId Tenant ID
   * @returns {Promise<Object>} Credential
   */
  async getCredentialForAssertion(response, userId, tenantId) {
    // If credential ID is provided
    if (response.id) {
      return await this.credentialRepository.getCredentialById(
        response.id,
        tenantId
      );
    }
    
    // If user handle is provided (for resident keys)
    if (response.response.userHandle) {
      const userHandle = base64url.decode(response.response.userHandle);
      if (userHandle !== userId) {
        throw new Error('User handle mismatch');
      }
      
      // Get credential by ID from authenticator data
      // This requires parsing authenticatorData to extract credential ID
      // Implementation depends on the exact format of authenticator data
      
      // For now, we'll use a simple approach of getting all user credentials
      // and matching against the provided rawId
      const userCredentials = await this.credentialRepository.getCredentialsByUserId(
        userId,
        tenantId
      );
      
      return userCredentials.find(cred => cred.credentialId === response.rawId);
    }
    
    throw new Error('Unable to identify credential');
  }
  
  /**
   * Cancels an ongoing authentication
   * 
   * @param {string} sessionId ID of authentication session to cancel
   * @returns {Promise<void>}
   */
  async cancelAuthentication(sessionId) {
    logger.debug(`[Fido2WebAuthnProvider] Cancelling authentication for session ${sessionId}`);
    
    // Remove challenge from session
    await this.sessionManager.removeAuthenticationChallenge(sessionId);
    
    // Audit cancellation
    await this.auditService.logAuthEvent({
      eventType: 'authentication:cancelled',
      providerId: this.id,
      sessionId,
      success: true
    });
  }
  
  /**
   * Checks if this provider supports enrollment
   * 
   * @returns {boolean} Whether enrollment is supported
   */
  supportsEnrollment() {
    return true;
  }
  
  /**
   * Starts the enrollment process
   * 
   * @param {string} userId User to enroll
   * @param {Object} context Enrollment context
   * @returns {Promise<Object>} Enrollment challenge
   */
  async startEnrollment(userId, context) {
    logger.debug(`[Fido2WebAuthnProvider] Starting enrollment for user ${userId}`);
    
    try {
      // Create a session ID for this enrollment attempt
      const sessionId = uuidv4();
      
      // Apply regional adaptations based on context
      const adaptedContext = await this.regionalAdapter.adaptEnrollmentContext(context);
      
      // Get user info
      const user = await context.getUserInfo(userId);
      
      if (!user) {
        throw new Error('User not found');
      }
      
      // Check how many credentials the user already has
      const existingCredentials = await this.credentialRepository.getCredentialsByUserId(
        userId,
        context.tenantId
      );
      
      if (existingCredentials.length >= this.config.maxAllowedCredentials) {
        throw new Error(`User already has maximum allowed credentials (${this.config.maxAllowedCredentials})`);
      }
      
      // Generate registration options
      const options = await this.webAuthnService.generateRegistrationOptions({
        user: {
          id: userId,
          name: user.username || user.email,
          displayName: user.displayName || user.username || user.email
        },
        attestation: adaptedContext.attestation || this.config.attestation,
        authenticatorSelection: {
          authenticatorAttachment: adaptedContext.authenticatorAttachment || this.config.authenticatorAttachment,
          requireResidentKey: adaptedContext.requireResidentKey || this.config.requireResidentKey,
          userVerification: adaptedContext.userVerification || this.config.userVerification
        },
        timeout: this.config.timeout,
        excludeCredentials: existingCredentials.map(cred => ({
          id: cred.credentialId,
          type: 'public-key',
          transports: cred.transports || []
        }))
      });
      
      // Store challenge for later verification
      await this.sessionManager.setEnrollmentChallenge(
        sessionId,
        options.challenge,
        userId,
        {
          rpId: this.config.rpId,
          origin: context.origin || this.config.origin,
          attestation: options.attestation,
          authenticatorSelection: options.authenticatorSelection
        },
        this.config.challengeTtl
      );
      
      // Audit enrollment start
      await this.auditService.logAuthEvent({
        eventType: 'enrollment:started',
        providerId: this.id,
        userId,
        tenantId: context.tenantId,
        sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: true
      });
      
      // Return enrollment challenge
      return {
        sessionId,
        challengeType: 'webauthn.create',
        challenge: options,
        expiresAt: new Date(Date.now() + this.config.timeout),
        uiOptions: {
          title: 'FIDO2 Registration',
          message: 'Please register your security key or biometric sensor',
          uiExtension: 'fido2-enrollment-ui',
          uiData: {
            authenticatorType: options.authenticatorSelection.authenticatorAttachment || 'any',
            requireUserVerification: options.authenticatorSelection.userVerification === 'required'
          }
        }
      };
    } catch (error) {
      logger.error(`[Fido2WebAuthnProvider] Error starting enrollment: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Completes the enrollment process
   * 
   * @param {Object} response Enrollment response
   * @param {Object} context Enrollment context
   * @returns {Promise<Object>} Enrollment result
   */
  async completeEnrollment(response, context) {
    logger.debug(`[Fido2WebAuthnProvider] Completing enrollment for session ${response.sessionId}`);
    
    try {
      // Get stored challenge data
      const challengeData = await this.sessionManager.getEnrollmentChallenge(response.sessionId);
      
      if (!challengeData) {
        throw new Error('Enrollment challenge not found or expired');
      }
      
      // Prepare credential for verification
      const credential = {
        id: response.response.id,
        rawId: base64url.toBuffer(response.response.rawId),
        response: {
          attestationObject: base64url.toBuffer(response.response.response.attestationObject),
          clientDataJSON: base64url.toBuffer(response.response.response.clientDataJSON)
        },
        type: response.response.type,
        transports: response.response.response.transports || []
      };
      
      // Verify attestation
      const verificationResult = await this.webAuthnService.verifyAttestation({
        credential,
        expectedChallenge: challengeData.challenge,
        expectedOrigin: challengeData.origin || this.config.origin,
        expectedRpId: this.config.rpId
      });
      
      if (!verificationResult.success) {
        throw new Error(verificationResult.error || 'Enrollment verification failed');
      }
      
      // Verify attestation certificate if attestation service is configured
      let attestationInfo = {};
      if (this.attestationService && verificationResult.attestationInfo) {
        attestationInfo = await this.attestationService.verifyAttestation(
          verificationResult.attestationInfo
        );
        
        // Apply regional requirements for attestation
        const regionalResult = await this.regionalAdapter.validateAttestationForRegion(
          attestationInfo,
          context
        );
        
        if (!regionalResult.valid) {
          throw new Error(regionalResult.error || 'Attestation does not meet regional requirements');
        }
      }
      
      // Store credential
      const credentialData = {
        id: uuidv4(),
        userId: challengeData.userId,
        tenantId: context.tenantId,
        credentialId: response.response.id,
        publicKey: verificationResult.publicKey,
        counter: verificationResult.counter,
        algorithm: verificationResult.algorithm,
        transports: response.response.response.transports || [],
        authenticatorAttachment: responseHasAttachment(response) || 'unknown',
        name: response.friendlyName || `Security Key (${new Date().toLocaleDateString()})`,
        aaguid: verificationResult.aaguid,
        attestationFormat: verificationResult.format,
        attestationLevel: attestationInfo.level || 'none',
        created: new Date(),
        lastUsed: new Date()
      };
      
      await this.credentialRepository.saveCredential(credentialData);
      
      // Audit successful enrollment
      await this.auditService.logAuthEvent({
        eventType: 'enrollment:completed',
        providerId: this.id,
        userId: challengeData.userId,
        tenantId: context.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: true,
        credentialId: credentialData.id,
        authenticatorData: {
          aaguid: credentialData.aaguid,
          attestationFormat: credentialData.attestationFormat,
          attestationLevel: credentialData.attestationLevel,
          authenticatorAttachment: credentialData.authenticatorAttachment
        }
      });
      
      // Return enrollment result
      return {
        success: true,
        userId: challengeData.userId,
        credentialId: credentialData.id,
        credentialName: credentialData.name,
        authenticatorInfo: {
          aaguid: credentialData.aaguid,
          attestationFormat: credentialData.attestationFormat,
          attestationLevel: credentialData.attestationLevel,
          authenticatorAttachment: credentialData.authenticatorAttachment
        }
      };
    } catch (error) {
      logger.error(`[Fido2WebAuthnProvider] Error completing enrollment: ${error.message}`);
      
      // Audit failed enrollment
      await this.auditService.logAuthEvent({
        eventType: 'enrollment:failed',
        providerId: this.id,
        userId: context.userId,
        tenantId: context.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: false,
        error: error.message
      });
      
      return {
        success: false,
        error: {
          code: 'enrollment_verification_failed',
          message: 'Enrollment verification failed',
          details: error.message
        }
      };
    }
  }
  
  /**
   * Checks provider health
   * 
   * @returns {Promise<Object>} Health status
   */
  async checkHealth() {
    try {
      // Check repository connection
      const repoHealth = await this.credentialRepository.checkHealth();
      
      // Check session manager
      const sessionHealth = await this.sessionManager.checkHealth();
      
      // Check attestation service if configured
      let attestationHealth = { status: 'not_configured' };
      if (this.attestationService) {
        attestationHealth = await this.attestationService.checkHealth();
      }
      
      const isHealthy = repoHealth.status === 'healthy' && 
                       sessionHealth.status === 'healthy' &&
                       (attestationHealth.status === 'healthy' || attestationHealth.status === 'not_configured');
      
      return {
        status: isHealthy ? 'healthy' : 'unhealthy',
        components: {
          repository: repoHealth,
          sessionManager: sessionHealth,
          attestationService: attestationHealth
        },
        timestamp: new Date()
      };
    } catch (error) {
      logger.error(`[Fido2WebAuthnProvider] Health check error: ${error.message}`);
      
      return {
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Gets provider metrics
   * 
   * @returns {Object} Usage and performance metrics
   */
  getMetrics() {
    return {
      activeEnrollments: this.sessionManager.getActiveEnrollmentCount(this.id),
      activeAuthentications: this.sessionManager.getActiveAuthenticationCount(this.id),
      averageAuthenticationTime: this.auditService.getAverageAuthenticationTime(this.id),
      successRate: this.auditService.getAuthenticationSuccessRate(this.id),
      enrollmentCount: this.auditService.getEnrollmentCount(this.id),
      timestamp: new Date()
    };
  }
  
  /**
   * Shuts down the provider
   * 
   * @returns {Promise<void>}
   */
  async shutdown() {
    logger.info('[Fido2WebAuthnProvider] Shutting down');
    
    // Clean up resources
    if (this.attestationService) {
      await this.attestationService.shutdown();
    }
    
    await this.credentialRepository.close();
  }
}

/**
 * Helper function to determine if response has authenticator attachment information
 * 
 * @param {Object} response The response to check
 * @returns {string|null} The attachment type or null
 */
function responseHasAttachment(response) {
  if (response.response && response.response.authenticatorAttachment) {
    return response.response.authenticatorAttachment;
  }
  
  return null;
}

/**
 * Register provider with Authentication Framework
 * 
 * @param {Object} registry The provider registry
 */
function register(registry) {
  registry.registerProvider(new Fido2WebAuthnProvider());
}

module.exports = {
  Fido2WebAuthnProvider,
  register
};
