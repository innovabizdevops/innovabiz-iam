/**
 * @file compliance-types.ts
 * @description Define tipos e enumerações relacionados a compliance e regulação
 * 
 * Este módulo implementa estruturas para conformidade regulatória, atendendo
 * aos requisitos de múltiplas jurisdições e frameworks regulatórios, especialmente
 * para mercados como Angola, CPLP, SADC, PALOP, BRICS, Europa e EUA.
 */

/**
 * Enumeração de frameworks regulatórios aplicáveis
 */
export enum RegulatoryFramework {
  // Frameworks globais
  GDPR = 'GDPR',                        // Regulamento Geral de Proteção de Dados (UE)
  PCI_DSS = 'PCI_DSS',                  // Payment Card Industry Data Security Standard
  ISO_27001 = 'ISO_27001',              // Segurança da Informação
  ISO_27701 = 'ISO_27701',              // Gestão de Privacidade
  ISO_20022 = 'ISO_20022',              // Mensagens financeiras
  BASEL_III = 'BASEL_III',              // Regulação bancária global
  AML_CFT = 'AML_CFT',                  // Anti-Money Laundering/Combating Financing of Terrorism
  FATF = 'FATF',                        // Financial Action Task Force
  
  // África e PALOP
  POPIA_ANGOLA = 'POPIA_ANGOLA',        // Proteção de Dados - Angola
  BNA_REGULATIONS = 'BNA_REGULATIONS',  // Banco Nacional de Angola
  CMC_REGULATIONS = 'CMC_REGULATIONS',  // Comissão do Mercado de Capitais (Angola)
  SADC_FI_PROTOCOL = 'SADC_FI_PROTOCOL',// Protocolo Financeiro da SADC
  BCEAO = 'BCEAO',                      // Banco Central dos Estados da África Ocidental
  ARSOP = 'ARSOP',                      // Autoridade Reguladora de Seguros (PALOP)
  
  // Brasil e América Latina
  LGPD = 'LGPD',                        // Lei Geral de Proteção de Dados (Brasil)
  BACEN = 'BACEN',                      // Banco Central do Brasil
  CVM = 'CVM',                          // Comissão de Valores Mobiliários (Brasil)
  PIX = 'PIX',                          // Sistema de Pagamentos Instantâneos (Brasil)
  SUSEP = 'SUSEP',                      // Superintendência de Seguros Privados (Brasil)
  
  // Europa
  MIFID_II = 'MIFID_II',                // Markets in Financial Instruments Directive
  PSD2 = 'PSD2',                        // Payment Services Directive 2
  SEPA = 'SEPA',                        // Single Euro Payments Area
  EBA_GUIDELINES = 'EBA_GUIDELINES',    // European Banking Authority
  
  // Estados Unidos
  SOX = 'SOX',                          // Sarbanes-Oxley Act
  GLBA = 'GLBA',                        // Gramm-Leach-Bliley Act
  CCPA = 'CCPA',                        // California Consumer Privacy Act
  FFIEC = 'FFIEC',                      // Federal Financial Institutions Examination Council
  FINRA = 'FINRA',                      // Financial Industry Regulatory Authority
  SEC = 'SEC',                          // Securities and Exchange Commission
  
  // BRICS e outras economias emergentes
  POPI = 'POPI',                        // Protection of Personal Information Act (África do Sul)
  PIPL = 'PIPL',                        // Personal Information Protection Law (China)
  RBI_GUIDELINES = 'RBI_GUIDELINES',    // Reserve Bank of India
  
  // Open Banking / Open Finance
  OPEN_BANKING_BR = 'OPEN_BANKING_BR',  // Open Banking Brasil
  OPEN_BANKING_EU = 'OPEN_BANKING_EU',  // Open Banking Europa
  OPEN_BANKING_AFRICA = 'OPEN_BANKING_AFRICA', // Open Banking África
  
  // Outros
  CUSTOM = 'CUSTOM'                     // Framework personalizado
}

/**
 * Nível de conformidade regulatória
 */
export enum ComplianceLevel {
  FULL = 'FULL',               // Conformidade total
  PARTIAL = 'PARTIAL',         // Conformidade parcial
  NON_COMPLIANT = 'NON_COMPLIANT', // Não conforme
  EXEMPTED = 'EXEMPTED',       // Isento
  NOT_APPLICABLE = 'NOT_APPLICABLE', // Não aplicável
  UNDER_REVIEW = 'UNDER_REVIEW'    // Em análise
}

/**
 * Status de processamento de dados pessoais
 */
export enum DataProcessingStatus {
  CONSENT_OBTAINED = 'CONSENT_OBTAINED',
  LEGITIMATE_INTEREST = 'LEGITIMATE_INTEREST',
  CONTRACT_FULFILLMENT = 'CONTRACT_FULFILLMENT',
  LEGAL_OBLIGATION = 'LEGAL_OBLIGATION',
  VITAL_INTEREST = 'VITAL_INTEREST',
  PUBLIC_INTEREST = 'PUBLIC_INTEREST',
  NO_CONSENT = 'NO_CONSENT',
  CONSENT_WITHDRAWN = 'CONSENT_WITHDRAWN',
  CONSENT_EXPIRED = 'CONSENT_EXPIRED'
}

/**
 * Tipos de dados sensíveis conforme regulações
 */
export enum SensitiveDataType {
  PERSONAL_ID = 'PERSONAL_ID',             // Números de identificação pessoal
  FINANCIAL = 'FINANCIAL',                 // Dados financeiros
  HEALTH = 'HEALTH',                       // Dados de saúde
  BIOMETRIC = 'BIOMETRIC',                 // Dados biométricos
  LOCATION = 'LOCATION',                   // Dados de localização
  RACIAL_ETHNIC = 'RACIAL_ETHNIC',         // Origem racial ou étnica
  POLITICAL_OPINIONS = 'POLITICAL_OPINIONS', // Opiniões políticas
  RELIGIOUS_BELIEFS = 'RELIGIOUS_BELIEFS', // Crenças religiosas
  TRADE_UNION = 'TRADE_UNION',             // Filiação sindical
  SEXUAL_ORIENTATION = 'SEXUAL_ORIENTATION', // Orientação sexual
  GENETIC_DATA = 'GENETIC_DATA',           // Dados genéticos
  CRIMINAL_RECORDS = 'CRIMINAL_RECORDS',   // Registos criminais
  CHILDREN_DATA = 'CHILDREN_DATA'          // Dados de menores
}

/**
 * Interface para requisitos de retenção de dados
 */
export interface DataRetentionRequirements {
  regulatoryFramework: RegulatoryFramework;
  region?: string;
  country?: string;
  dataCategory: string;
  minimumRetentionPeriod?: {
    years?: number;
    months?: number;
    days?: number;
  };
  maximumRetentionPeriod?: {
    years?: number;
    months?: number;
    days?: number;
  };
  requiresAnonymization?: boolean;
  requiresEncryption?: boolean;
  additionalRequirements?: string[];
}

/**
 * Interface para consentimento de processamento de dados
 */
export interface DataProcessingConsent {
  consentId: string;
  userId: string;
  consentType: string;
  dataCategories: string[];
  purposes: string[];
  grantedAt: Date;
  expiresAt?: Date;
  withdrawableAt?: Date;
  status: DataProcessingStatus;
  regulatoryFrameworks: RegulatoryFramework[];
  channels?: {
    channelType: string;
    consentGiven: boolean;
    consentTimestamp?: Date;
  }[];
  auditTrail?: {
    action: 'GRANTED' | 'MODIFIED' | 'WITHDRAWN' | 'EXPIRED';
    timestamp: Date;
    ipAddress?: string;
    userAgent?: string;
  }[];
}

/**
 * Interface para informações de compliance de uma região
 */
export interface RegionComplianceInfo {
  region: string;
  countries: string[];
  primaryRegulator?: string;
  applicableFrameworks: RegulatoryFramework[];
  dataLocalizationRequired: boolean;
  specialRequirements?: {
    framework: RegulatoryFramework;
    requirement: string;
    implementationStatus: ComplianceLevel;
    details?: string;
  }[];
}

/**
 * Interface para regras de conformidade
 */
export interface ComplianceRule {
  ruleId: string;
  name: string;
  description: string;
  regulatoryFramework: RegulatoryFramework;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  applicableRegions: string[];
  applicableCountries: string[];
  implementationStatus: ComplianceLevel;
  validationCriteria: {
    criteriaType: string;
    criteriaValue: any;
  }[];
  enforcementAction?: 'BLOCK' | 'WARN' | 'LOG' | 'NOTIFY';
  remediationSteps?: string[];
}

/**
 * Utilitário para verificação de compliance de eventos
 */
export class ComplianceVerifier {
  /**
   * Verifica se um evento está em conformidade com uma regra específica
   * @param event Evento a ser verificado
   * @param rule Regra de conformidade
   * @returns Status de conformidade e detalhes
   */
  static verifyEventCompliance(event: any, rule: ComplianceRule): {
    isCompliant: boolean;
    details: string;
    remediationRequired: boolean;
  } {
    // Implementação básica para demonstração
    let isCompliant = true;
    let details = '';
    
    for (const criteria of rule.validationCriteria) {
      // Verificação simplificada para demonstração
      if (event[criteria.criteriaType] !== criteria.criteriaValue) {
        isCompliant = false;
        details = `Evento não atende ao critério: ${criteria.criteriaType}`;
        break;
      }
    }
    
    return {
      isCompliant,
      details,
      remediationRequired: !isCompliant && rule.enforcementAction !== 'LOG'
    };
  }
  
  /**
   * Verifica se um evento necessita de consentimento específico
   * @param eventType Tipo do evento
   * @param dataCategories Categorias de dados envolvidas
   * @returns Informações de consentimento necessário
   */
  static verifyConsentRequirements(
    eventType: string, 
    dataCategories: string[]
  ): {
    consentRequired: boolean;
    regulatoryFrameworks: RegulatoryFramework[];
  } {
    // Implementação simplificada para demonstração
    const sensitiveCategories = [
      'PERSONAL_ID', 'FINANCIAL', 'HEALTH', 'BIOMETRIC', 'LOCATION'
    ];
    
    const hasAnySensitiveData = dataCategories.some(
      category => sensitiveCategories.includes(category)
    );
    
    return {
      consentRequired: hasAnySensitiveData,
      regulatoryFrameworks: hasAnySensitiveData 
        ? [RegulatoryFramework.GDPR, RegulatoryFramework.POPIA_ANGOLA, RegulatoryFramework.LGPD] 
        : []
    };
  }
}