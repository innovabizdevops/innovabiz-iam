/**
 * Regional Adapter for FIDO2/WebAuthn Provider
 * 
 * This module handles region-specific adaptations for the FIDO2/WebAuthn authentication provider,
 * ensuring compliance with regional requirements and optimizing for regional characteristics.
 * 
 * @module regional-adapter
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { logger } = require('@innovabiz/auth-framework/utils/logging');

/**
 * Regional constants
 */
const REGIONS = {
  EU: 'EU',   // European Union (Portugal)
  BR: 'BR',   // Brazil
  AO: 'AO',   // Angola
  US: 'US'    // United States
};

/**
 * Regional compliance frameworks
 */
const COMPLIANCE_FRAMEWORKS = {
  EU: ['GDPR', 'eIDAS'],
  BR: ['LGPD'],
  AO: ['PNDSB'],
  US: ['NIST-800-63-3', 'HIPAA', 'PCI-DSS']
};

/**
 * Class for handling regional adaptations for FIDO2/WebAuthn authentication
 */
class RegionalAdapter {
  /**
   * Creates a new RegionalAdapter instance
   * 
   * @param {Object} regionalSettings Region-specific settings
   */
  constructor(regionalSettings = {}) {
    this.settings = regionalSettings;
    
    // Default settings for regions if not provided
    this.ensureDefaultSettings();
    
    logger.info('[RegionalAdapter] Initialized with settings for regions: ' + 
                Object.keys(this.settings).join(', '));
  }
  
  /**
   * Ensures default settings exist for all supported regions
   */
  ensureDefaultSettings() {
    for (const region of Object.values(REGIONS)) {
      if (!this.settings[region]) {
        this.settings[region] = this.getDefaultSettingsForRegion(region);
      } else {
        // Merge with defaults to ensure all properties exist
        this.settings[region] = {
          ...this.getDefaultSettingsForRegion(region),
          ...this.settings[region]
        };
      }
    }
  }
  
  /**
   * Gets default settings for a specific region
   * 
   * @param {string} region Region code
   * @returns {Object} Default settings for the region
   */
  getDefaultSettingsForRegion(region) {
    switch (region) {
      case REGIONS.EU:
        return {
          attestationRequired: false,
          minAssuranceLevel: 2,
          userVerification: 'preferred',
          preferPlatformAuthenticator: false,
          compatibilityMode: false,
          supportedCardReaders: ['Cartão de Cidadão'],
          enforceStrongAttestationForFinance: true,
          enforceBiometricForHealthcare: false,
          privacyNoticeRequired: true,
          eIDAS_compliance: {
            required: false,
            level: 'substantial'
          }
        };
        
      case REGIONS.BR:
        return {
          attestationRequired: false,
          minAssuranceLevel: 2,
          userVerification: 'preferred',
          preferPlatformAuthenticator: true,
          compatibilityMode: false,
          supportIcpBrasil: true,
          enforceBiometricForFinance: true,
          enforceBiometricForHealthcare: false,
          privacyNoticeRequired: true,
          lgpdComplianceMode: 'strict'
        };
        
      case REGIONS.AO:
        return {
          attestationRequired: false,
          minAssuranceLevel: 1,
          userVerification: 'preferred',
          preferPlatformAuthenticator: true,
          compatibilityMode: true,
          optimizeForLowBandwidth: true,
          offlineAuthEnabled: true,
          enforceBiometricForFinance: false,
          enforceBiometricForHealthcare: false,
          privacyNoticeRequired: false,
          pndsbComplianceMode: 'standard'
        };
        
      case REGIONS.US:
        return {
          attestationRequired: false,
          minAssuranceLevel: 3,
          userVerification: 'required',
          preferPlatformAuthenticator: false,
          compatibilityMode: false,
          healthcareMode: false,
          enforceBiometricForFinance: true,
          enforceHardwareTokenForHealthcare: true,
          privacyNoticeRequired: true,
          nistCompliance: {
            required: true,
            level: 'AAL3'
          }
        };
        
      default:
        return {
          attestationRequired: false,
          minAssuranceLevel: 2,
          userVerification: 'preferred',
          preferPlatformAuthenticator: false,
          compatibilityMode: false
        };
    }
  }
  
  /**
   * Determines the region from the authentication context
   * 
   * @param {Object} context Authentication context
   * @returns {string} Determined region code
   */
  determineRegion(context) {
    // Use explicit region if provided
    if (context.region && REGIONS[context.region]) {
      return context.region;
    }
    
    // Try to determine from tenant configuration
    if (context.tenantId && context.tenantRegion) {
      return context.tenantRegion;
    }
    
    // Try to determine from geolocation
    if (context.geoLocation) {
      return this.determineRegionFromGeolocation(context.geoLocation);
    }
    
    // Use locale as a fallback
    if (context.locale) {
      return this.determineRegionFromLocale(context.locale);
    }
    
    // Default to EU as a safe fallback (most stringent requirements)
    logger.warn('[RegionalAdapter] Could not determine region, defaulting to EU');
    return REGIONS.EU;
  }
  
  /**
   * Determines region from geolocation
   * 
   * @param {Object} geoLocation Geolocation data
   * @returns {string} Determined region code
   */
  determineRegionFromGeolocation(geoLocation) {
    // Extract country code from geolocation
    const country = geoLocation.countryCode || geoLocation.country || '';
    
    // Map country to region
    switch (country.toUpperCase()) {
      case 'PT':
      case 'ES':
      case 'FR':
      case 'DE':
      case 'IT':
      case 'BE':
      case 'NL':
      case 'LU':
      case 'DK':
      case 'IE':
      case 'GR':
      case 'AT':
      case 'FI':
      case 'SE':
        return REGIONS.EU;
        
      case 'BR':
        return REGIONS.BR;
        
      case 'AO':
      case 'CD':  // Democratic Republic of Congo
      case 'CG':  // Republic of Congo
        return REGIONS.AO;
        
      case 'US':
        return REGIONS.US;
        
      default:
        // Default based on continent
        if (geoLocation.continent === 'Europe') {
          return REGIONS.EU;
        } else if (geoLocation.continent === 'South America') {
          return REGIONS.BR;
        } else if (geoLocation.continent === 'Africa') {
          return REGIONS.AO;
        } else if (geoLocation.continent === 'North America') {
          return REGIONS.US;
        }
        
        // Default to EU as a safe fallback
        return REGIONS.EU;
    }
  }
  
  /**
   * Determines region from locale
   * 
   * @param {string} locale User locale
   * @returns {string} Determined region code
   */
  determineRegionFromLocale(locale) {
    const lang = locale.split('-')[0].toLowerCase();
    const country = locale.split('-')[1]?.toUpperCase();
    
    if (country) {
      // Try to match by country code in locale
      if (country === 'PT' || country === 'EU') {
        return REGIONS.EU;
      } else if (country === 'BR') {
        return REGIONS.BR;
      } else if (country === 'AO') {
        return REGIONS.AO;
      } else if (country === 'US') {
        return REGIONS.US;
      }
    }
    
    // Try to match by language
    if (lang === 'pt') {
      // Additional heuristic could be applied here
      // For now, default Portuguese to EU
      return REGIONS.EU;
    } else if (lang === 'en') {
      return REGIONS.US;
    }
    
    // Default to EU as a safe fallback
    return REGIONS.EU;
  }
  
  /**
   * Adapts authentication context based on regional requirements
   * 
   * @param {Object} context Authentication context
   * @returns {Promise<Object>} Adapted context
   */
  async adaptAuthenticationContext(context) {
    // Determine region
    const region = this.determineRegion(context);
    logger.debug(`[RegionalAdapter] Adapting authentication for region: ${region}`);
    
    // Get regional settings
    const regionalSettings = this.settings[region] || this.getDefaultSettingsForRegion(region);
    
    // Create adapted context
    const adaptedContext = { ...context };
    
    // Apply regional adaptations
    this.applyGeneralRegionalAdaptations(adaptedContext, regionalSettings, region);
    
    // Apply industry-specific adaptations if applicable
    if (context.industry) {
      this.applyIndustrySpecificAdaptations(adaptedContext, regionalSettings, region, context.industry);
    }
    
    // Apply user-specific adaptations if applicable
    if (context.userId) {
      await this.applyUserSpecificAdaptations(adaptedContext, regionalSettings, region, context.userId);
    }
    
    // Add compliance requirements
    adaptedContext.complianceFrameworks = COMPLIANCE_FRAMEWORKS[region] || [];
    
    // Add adapted context flags
    adaptedContext.adaptedForRegion = region;
    
    return adaptedContext;
  }
  
  /**
   * Adapts enrollment context based on regional requirements
   * 
   * @param {Object} context Enrollment context
   * @returns {Promise<Object>} Adapted context
   */
  async adaptEnrollmentContext(context) {
    // Determine region
    const region = this.determineRegion(context);
    logger.debug(`[RegionalAdapter] Adapting enrollment for region: ${region}`);
    
    // Get regional settings
    const regionalSettings = this.settings[region] || this.getDefaultSettingsForRegion(region);
    
    // Create adapted context
    const adaptedContext = { ...context };
    
    // Apply regional adaptations for enrollment
    this.applyEnrollmentRegionalAdaptations(adaptedContext, regionalSettings, region);
    
    // Apply industry-specific adaptations if applicable
    if (context.industry) {
      this.applyIndustrySpecificEnrollmentAdaptations(
        adaptedContext, 
        regionalSettings, 
        region, 
        context.industry
      );
    }
    
    // Add compliance requirements
    adaptedContext.complianceFrameworks = COMPLIANCE_FRAMEWORKS[region] || [];
    adaptedContext.privacyNoticeRequired = regionalSettings.privacyNoticeRequired;
    
    // Add adapted context flags
    adaptedContext.adaptedForRegion = region;
    
    return adaptedContext;
  }
  
  /**
   * Applies general regional adaptations to context
   * 
   * @param {Object} context Context to adapt
   * @param {Object} settings Regional settings
   * @param {string} region Region code
   */
  applyGeneralRegionalAdaptations(context, settings, region) {
    // Set user verification requirement
    context.userVerification = settings.userVerification;
    
    // Set authenticator attachment preference
    if (settings.preferPlatformAuthenticator) {
      context.authenticatorAttachment = 'platform';
    }
    
    // Set timeout for low-bandwidth regions
    if (settings.optimizeForLowBandwidth) {
      context.timeout = 120000; // Extended timeout for low-bandwidth regions
    }
    
    // Set compatibility mode if needed
    if (settings.compatibilityMode) {
      context.compatibilityMode = true;
    }
    
    // Region-specific adaptations
    switch (region) {
      case REGIONS.EU:
        // Apply eIDAS adaptations if required
        if (settings.eIDAS_compliance && settings.eIDAS_compliance.required) {
          context.requireStrongAttestation = true;
          context.assuranceLevel = settings.eIDAS_compliance.level === 'high' ? 3 : 2;
        }
        break;
        
      case REGIONS.BR:
        // Apply ICP-Brasil adaptations if supported
        if (settings.supportIcpBrasil) {
          context.supportIcpBrasil = true;
        }
        break;
        
      case REGIONS.AO:
        // Apply offline authentication adaptations if enabled
        if (settings.offlineAuthEnabled) {
          context.allowOfflineAuth = true;
        }
        break;
        
      case REGIONS.US:
        // Apply NIST compliance adaptations if required
        if (settings.nistCompliance && settings.nistCompliance.required) {
          context.userVerification = 'required';
          context.requireStrongAttestation = true;
        }
        
        // Apply healthcare mode if enabled
        if (settings.healthcareMode) {
          context.requireHardwareToken = true;
        }
        break;
    }
  }
  
  /**
   * Applies enrollment-specific regional adaptations to context
   * 
   * @param {Object} context Context to adapt
   * @param {Object} settings Regional settings
   * @param {string} region Region code
   */
  applyEnrollmentRegionalAdaptations(context, settings, region) {
    // Set attestation requirement
    context.attestation = settings.attestationRequired ? 'direct' : 'none';
    
    // Set user verification requirement for enrollment
    context.userVerification = settings.userVerification;
    
    // Set authenticator attachment preference
    if (settings.preferPlatformAuthenticator) {
      context.authenticatorAttachment = 'platform';
    }
    
    // Set resident key requirement
    // For EU and US under certain conditions, require resident keys
    if ((region === REGIONS.EU && settings.eIDAS_compliance?.level === 'high') ||
        (region === REGIONS.US && settings.nistCompliance?.level === 'AAL3')) {
      context.requireResidentKey = true;
    }
    
    // Region-specific enrollment adaptations
    switch (region) {
      case REGIONS.EU:
        // Apply eIDAS adaptations for enrollment
        if (settings.eIDAS_compliance && settings.eIDAS_compliance.required) {
          context.attestation = 'direct';
          context.requireQualifiedAuthenticator = true;
        }
        break;
        
      case REGIONS.BR:
        // Apply LGPD-specific adaptations for enrollment
        if (settings.lgpdComplianceMode === 'strict') {
          context.requireExplicitConsent = true;
          context.consentLanguage = 'pt-BR';
        }
        break;
        
      case REGIONS.AO:
        // Optimize for low-bandwidth enrollment
        if (settings.optimizeForLowBandwidth) {
          context.timeout = 180000; // 3 minutes for enrollment in low-bandwidth areas
          context.progressiveEnrollment = true;
        }
        break;
        
      case REGIONS.US:
        // Apply NIST-specific adaptations for enrollment
        if (settings.nistCompliance && settings.nistCompliance.required) {
          context.attestation = 'direct';
          context.userVerification = 'required';
        }
        break;
    }
  }
  
  /**
   * Applies industry-specific adaptations to context
   * 
   * @param {Object} context Context to adapt
   * @param {Object} settings Regional settings
   * @param {string} region Region code
   * @param {string} industry Industry code
   */
  applyIndustrySpecificAdaptations(context, settings, region, industry) {
    switch (industry.toLowerCase()) {
      case 'finance':
      case 'banking':
      case 'insurance':
        // Financial industry typically requires stronger authentication
        context.userVerification = 'required';
        
        // For EU financial services, enforce strong attestation
        if (region === REGIONS.EU && settings.enforceStrongAttestationForFinance) {
          context.requireStrongAttestation = true;
        }
        
        // For Brazil, enforce biometric for finance
        if (region === REGIONS.BR && settings.enforceBiometricForFinance) {
          context.authenticatorAttachment = 'platform'; // Often for biometrics
          context.userVerification = 'required';
        }
        
        // For US, apply PCI-DSS specific requirements
        if (region === REGIONS.US) {
          context.requireStrongAttestation = true;
          context.userVerification = 'required';
        }
        break;
        
      case 'healthcare':
      case 'health':
      case 'medical':
        // Healthcare typically requires strong verification
        
        // For EU healthcare under GDPR
        if (region === REGIONS.EU) {
          context.requireStrongAttestation = true;
          
          if (settings.enforceBiometricForHealthcare) {
            context.authenticatorAttachment = 'platform';
          }
        }
        
        // For Brazil healthcare under LGPD
        if (region === REGIONS.BR && settings.enforceBiometricForHealthcare) {
          context.authenticatorAttachment = 'platform';
          context.userVerification = 'required';
        }
        
        // For US healthcare under HIPAA
        if (region === REGIONS.US && settings.enforceHardwareTokenForHealthcare) {
          context.authenticatorAttachment = 'cross-platform'; // Hardware tokens
          context.userVerification = 'required';
        }
        break;
        
      case 'government':
      case 'public':
        // Government typically requires very strong authentication
        context.userVerification = 'required';
        
        // For EU government services, often require eID integration
        if (region === REGIONS.EU) {
          context.requireStrongAttestation = true;
          context.preferQualifiedAuthenticators = true;
        }
        
        // For Brazil gov services, support ICP-Brasil
        if (region === REGIONS.BR && settings.supportIcpBrasil) {
          context.supportIcpBrasil = true;
        }
        
        // For US government
        if (region === REGIONS.US) {
          context.requireStrongAttestation = true;
          context.preferFips140Authenticators = true;
        }
        break;
        
      default:
        // For other industries, use regional defaults
        break;
    }
  }
  
  /**
   * Applies industry-specific adaptations to enrollment context
   * 
   * @param {Object} context Context to adapt
   * @param {Object} settings Regional settings
   * @param {string} region Region code
   * @param {string} industry Industry code
   */
  applyIndustrySpecificEnrollmentAdaptations(context, settings, region, industry) {
    switch (industry.toLowerCase()) {
      case 'finance':
      case 'banking':
      case 'insurance':
        // Financial industry enrollment requirements
        context.attestation = 'direct'; // Typically want attestation for financial
        context.userVerification = 'required';
        
        if (region === REGIONS.EU || region === REGIONS.US) {
          context.requireResidentKey = true; // Better UX for financial sector
        }
        
        break;
        
      case 'healthcare':
      case 'health':
      case 'medical':
        // Healthcare enrollment requirements
        context.attestation = 'direct';
        
        // For US healthcare, often need hardware tokens
        if (region === REGIONS.US && settings.enforceHardwareTokenForHealthcare) {
          context.authenticatorAttachment = 'cross-platform';
        }
        
        // Enhanced consent for healthcare in EU and BR
        if (region === REGIONS.EU || region === REGIONS.BR) {
          context.requireExplicitConsent = true;
          context.consentContext = 'healthcare';
        }
        
        break;
        
      case 'government':
      case 'public':
        // Government enrollment requirements
        context.attestation = 'direct';
        context.userVerification = 'required';
        
        // For EU government, often integrate with national eID
        if (region === REGIONS.EU) {
          context.supportedAuthenticators = [...(context.supportedAuthenticators || []), ...settings.supportedCardReaders];
        }
        
        break;
        
      default:
        // For other industries, use regional defaults
        break;
    }
  }
  
  /**
   * Applies user-specific adaptations to context
   * 
   * @param {Object} context Context to adapt
   * @param {Object} settings Regional settings
   * @param {string} region Region code
   * @param {string} userId User ID
   * @returns {Promise<void>}
   */
  async applyUserSpecificAdaptations(context, settings, region, userId) {
    // This would typically involve looking up user preferences or history
    // For now, we'll use a simplified approach
    
    // In a real implementation, this would query the user profile service
    // and apply user-specific adaptations
    
    // For demonstration, just log the intent
    logger.debug(`[RegionalAdapter] Would apply user-specific adaptations for user ${userId} in region ${region}`);
  }
  
  /**
   * Validates attestation for a specific region
   * 
   * @param {Object} attestationInfo Attestation information
   * @param {Object} context Authentication context
   * @returns {Promise<Object>} Validation result
   */
  async validateAttestationForRegion(attestationInfo, context) {
    // Determine region
    const region = this.determineRegion(context);
    logger.debug(`[RegionalAdapter] Validating attestation for region: ${region}`);
    
    // Get regional settings
    const regionalSettings = this.settings[region] || this.getDefaultSettingsForRegion(region);
    
    // Basic validation result
    const result = {
      valid: true,
      assuranceLevel: 1,
      error: null
    };
    
    // Check if attestation meets minimum regional requirements
    if (attestationInfo.attestationLevel === 'none' && regionalSettings.attestationRequired) {
      result.valid = false;
      result.error = 'Attestation required for this region';
      return result;
    }
    
    // Region-specific validations
    switch (region) {
      case REGIONS.EU:
        // eIDAS specific validations if applicable
        if (regionalSettings.eIDAS_compliance && regionalSettings.eIDAS_compliance.required) {
          if (regionalSettings.eIDAS_compliance.level === 'high') {
            // For high level eIDAS, verify hardware attestation
            if (!attestationInfo.isHardwareBacked) {
              result.valid = false;
              result.error = 'Hardware-backed authenticator required for eIDAS high level';
            }
          }
        }
        
        // Check for qualified certificates if required
        if (context.requireQualifiedAuthenticator && !attestationInfo.isQualified) {
          result.valid = false;
          result.error = 'Qualified authenticator required';
        }
        break;
        
      case REGIONS.US:
        // NIST specific validations if applicable
        if (regionalSettings.nistCompliance && regionalSettings.nistCompliance.required) {
          if (regionalSettings.nistCompliance.level === 'AAL3') {
            // For NIST AAL3, verify hardware attestation
            if (!attestationInfo.isHardwareBacked) {
              result.valid = false;
              result.error = 'Hardware-backed authenticator required for NIST AAL3';
            }
          }
        }
        
        // FIPS 140 validation if required
        if (context.preferFips140Authenticators && context.industry === 'government') {
          if (!attestationInfo.isFips140Certified) {
            // For government, this might be a hard requirement
            result.valid = false;
            result.error = 'FIPS 140 certified authenticator required for government systems';
          }
        }
        break;
        
      case REGIONS.BR:
        // ICP-Brasil validations if applicable
        if (context.supportIcpBrasil && context.requireIcpBrasilAttestation) {
          if (!attestationInfo.isIcpBrasilCertified) {
            result.valid = false;
            result.error = 'ICP-Brasil certified authenticator required';
          }
        }
        break;
    }
    
    // Industry-specific validations
    if (context.industry) {
      switch (context.industry.toLowerCase()) {
        case 'finance':
          // For finance, enforce strong attestation requirements
          if (context.requireStrongAttestation && !attestationInfo.isHardwareBacked) {
            result.valid = false;
            result.error = 'Hardware-backed authenticator required for financial services';
          }
          break;
          
        case 'healthcare':
          // For healthcare, verify appropriate attestation level
          if (region === REGIONS.US && !attestationInfo.isHardwareBacked && context.enforceHardwareTokenForHealthcare) {
            result.valid = false;
            result.error = 'Hardware-backed authenticator required for healthcare under HIPAA';
          }
          break;
      }
    }
    
    // Calculate assurance level based on attestation and context
    result.assuranceLevel = this.calculateAssuranceLevel(attestationInfo, context, region);
    
    // Check if assurance level meets minimum requirement
    if (result.valid && result.assuranceLevel < regionalSettings.minAssuranceLevel) {
      result.valid = false;
      result.error = `Authenticator does not meet minimum assurance level for ${region}`;
    }
    
    return result;
  }
  
  /**
   * Calculates assurance level based on attestation and context
   * 
   * @param {Object} attestationInfo Attestation information
   * @param {Object} context Authentication context
   * @param {string} region Region code
   * @returns {number} Assurance level (1-3)
   */
  calculateAssuranceLevel(attestationInfo, context, region) {
    let level = 1;
    
    // Base level on attestation properties
    if (attestationInfo.isHardwareBacked) {
      level = 3;
    } else if (attestationInfo.isSoftwareProtected) {
      level = 2;
    }
    
    // Adjust for authenticator properties
    if (attestationInfo.userVerification === 'required') {
      level = Math.max(level, 2);
    }
    
    // Adjust for certification status
    if (attestationInfo.certificationLevel) {
      switch (attestationInfo.certificationLevel) {
        case 'FIDO_CERTIFIED_L1':
          level = Math.max(level, 1);
          break;
        case 'FIDO_CERTIFIED_L2':
          level = Math.max(level, 2);
          break;
        case 'FIDO_CERTIFIED_L3':
          level = Math.max(level, 3);
          break;
      }
    }
    
    // Region-specific adjustments
    switch (region) {
      case REGIONS.EU:
        // For EU, qualified status increases assurance
        if (attestationInfo.isQualified) {
          level = 3;
        }
        break;
        
      case REGIONS.BR:
        // For Brazil, ICP-Brasil certification increases assurance
        if (attestationInfo.isIcpBrasilCertified) {
          level = 3;
        }
        break;
        
      case REGIONS.US:
        // For US, FIPS certification increases assurance
        if (attestationInfo.isFips140Certified) {
          level = 3;
        }
        break;
    }
    
    return level;
  }
}

module.exports = RegionalAdapter;
