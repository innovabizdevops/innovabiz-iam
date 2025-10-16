/**
 * @file bureau-credit-events.ts
 * @description Define os tipos de eventos específicos do Bureau de Créditos
 * 
 * Este módulo implementa as estruturas de dados para eventos relacionados
 * ao Bureau de Créditos, incluindo alterações de score, alertas de fraude,
 * atualizações de perfil de crédito e outras operações relacionadas.
 */

import { v4 as uuidv4 } from 'uuid';
import { BaseEvent, EventBuilder, EventCategory, EventPriority } from '../core/base-event';
import { RegulatoryFramework } from '../core/compliance-types';

/**
 * Tipos de eventos do Bureau de Créditos
 */
export enum BureauCreditEventType {
  // Eventos de alteração de score
  SCORE_CHANGED = 'SCORE_CHANGED',
  SCORE_THRESHOLD_REACHED = 'SCORE_THRESHOLD_REACHED',
  
  // Eventos de fraude e segurança
  FRAUD_ALERT = 'FRAUD_ALERT',
  IDENTITY_VERIFICATION_FAILED = 'IDENTITY_VERIFICATION_FAILED',
  SUSPICIOUS_ACTIVITY = 'SUSPICIOUS_ACTIVITY',
  IDENTITY_THEFT_WARNING = 'IDENTITY_THEFT_WARNING',
  
  // Eventos de atualizações de conta
  CREDIT_REPORT_UPDATED = 'CREDIT_REPORT_UPDATED',
  CREDIT_INQUIRY = 'CREDIT_INQUIRY',
  NEW_ACCOUNT_CREATED = 'NEW_ACCOUNT_CREATED',
  ACCOUNT_STATUS_CHANGED = 'ACCOUNT_STATUS_CHANGED',
  
  // Eventos de pagamentos
  PAYMENT_LATE = 'PAYMENT_LATE',
  PAYMENT_MISSED = 'PAYMENT_MISSED',
  PAYMENT_MADE = 'PAYMENT_MADE',
  
  // Eventos de limites
  CREDIT_LIMIT_CHANGED = 'CREDIT_LIMIT_CHANGED',
  CREDIT_LIMIT_REACHED = 'CREDIT_LIMIT_REACHED',
  
  // Eventos de perfil de crédito
  CREDIT_MIX_CHANGED = 'CREDIT_MIX_CHANGED',
  CREDIT_HISTORY_LENGTH_MILESTONE = 'CREDIT_HISTORY_LENGTH_MILESTONE',
  
  // Eventos de dívidas e cobranças
  DEBT_COLLECTION_INITIATED = 'DEBT_COLLECTION_INITIATED',
  DEBT_SETTLED = 'DEBT_SETTLED',
  DEBT_TRANSFERRED = 'DEBT_TRANSFERRED',
  DEBT_WRITTEN_OFF = 'DEBT_WRITTEN_OFF',
  
  // Eventos regulatórios
  REGULATORY_STATUS_CHANGED = 'REGULATORY_STATUS_CHANGED',
  COMPLIANCE_ALERT = 'COMPLIANCE_ALERT',
  WATCHLIST_STATUS_CHANGED = 'WATCHLIST_STATUS_CHANGED',
  
  // Eventos de solicitação e aprovação
  CREDIT_APPLICATION_SUBMITTED = 'CREDIT_APPLICATION_SUBMITTED',
  CREDIT_APPLICATION_APPROVED = 'CREDIT_APPLICATION_APPROVED',
  CREDIT_APPLICATION_DENIED = 'CREDIT_APPLICATION_DENIED',
  
  // Eventos de disputa
  DISPUTE_FILED = 'DISPUTE_FILED',
  DISPUTE_RESOLVED = 'DISPUTE_RESOLVED',
  
  // Eventos de integração
  DATA_SOURCE_UPDATED = 'DATA_SOURCE_UPDATED',
  INTEGRATION_STATUS_CHANGED = 'INTEGRATION_STATUS_CHANGED'
}

/**
 * Severidade de eventos de crédito
 */
export enum CreditEventSeverity {
  INFORMATIONAL = 'INFORMATIONAL',
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL'
}

/**
 * Status de alerta de crédito
 */
export enum CreditAlertStatus {
  NEW = 'NEW',
  ACKNOWLEDGED = 'ACKNOWLEDGED',
  IN_PROGRESS = 'IN_PROGRESS',
  RESOLVED = 'RESOLVED',
  DISMISSED = 'DISMISSED',
  ESCALATED = 'ESCALATED',
  REQUIRES_ACTION = 'REQUIRES_ACTION'
}

/**
 * Interface para Eventos do Bureau de Créditos
 */
export interface BureauCreditEvent extends BaseEvent {
  /**
   * Tipo específico do evento de crédito
   */
  eventType: BureauCreditEventType;
  
  /**
   * Severidade do evento
   */
  severity: CreditEventSeverity;
  
  /**
   * Status do alerta, se aplicável
   */
  alertStatus?: CreditAlertStatus;
  
  /**
   * Data de expiração do evento, se aplicável
   */
  expiresAt?: Date;
  
  /**
   * Valor anterior relacionado ao evento (ex: score anterior)
   */
  previousValue?: number | string;
  
  /**
   * Novo valor relacionado ao evento (ex: novo score)
   */
  newValue?: number | string;
  
  /**
   * Lista de IDs de identidade relacionados ao evento
   */
  relatedIdentityIds: string[];
  
  /**
   * Dados do relatório de crédito
   */
  creditReportData?: {
    reportId?: string;
    reportDate?: Date;
    reportType?: string;
    reportSource?: string;
    score?: number;
    maxScore?: number;
    scoreBand?: string;
    indicators?: {
      name: string;
      value: string | number;
      status: string;
    }[];
  };
  
  /**
   * Informações sobre fraude, se aplicável
   */
  fraudData?: {
    fraudType?: string;
    riskLevel?: string;
    confidenceScore?: number;
    detectionMethod?: string;
    recommendedActions?: string[];
  };
  
  /**
   * Informações sobre disputas, se aplicável
   */
  disputeData?: {
    disputeId?: string;
    disputeReason?: string;
    disputeDate?: Date;
    disputeStatus?: string;
    estimatedResolutionDate?: Date;
  };
  
  /**
   * Informações regulatórias
   */
  regulatoryInfo?: {
    frameworks: RegulatoryFramework[];
    complianceStatus?: string;
    requiredActions?: string[];
  };
}

/**
 * Builder para eventos do Bureau de Créditos
 */
export class BureauCreditEventBuilder extends EventBuilder<BureauCreditEvent> {
  constructor() {
    super();
    // Valores padrão para eventos do bureau
    this.event.category = EventCategory.BUSINESS_PROCESS;
    this.event.source = 'bureau-creditos';
    this.event.code = 'BUREAU_CREDIT_EVENT';
    this.event.relatedIdentityIds = [];
  }
  
  /**
   * Define o tipo específico do evento
   * @param eventType Tipo de evento do bureau de créditos
   */
  withEventType(eventType: BureauCreditEventType): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).eventType = eventType;
    // Atualiza o código com base no tipo de evento
    this.event.code = `BUREAU_CREDIT_${eventType}`;
    return this;
  }
  
  /**
   * Define a severidade do evento
   * @param severity Severidade do evento
   */
  withSeverity(severity: CreditEventSeverity): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).severity = severity;
    
    // Atualiza prioridade baseado na severidade
    switch (severity) {
      case CreditEventSeverity.CRITICAL:
        this.withPriority(EventPriority.CRITICAL);
        break;
      case CreditEventSeverity.HIGH:
        this.withPriority(EventPriority.HIGH);
        break;
      case CreditEventSeverity.MEDIUM:
        this.withPriority(EventPriority.MEDIUM);
        break;
      case CreditEventSeverity.LOW:
        this.withPriority(EventPriority.LOW);
        break;
      case CreditEventSeverity.INFORMATIONAL:
        this.withPriority(EventPriority.LOW);
        break;
    }
    
    return this;
  }
  
  /**
   * Define o status do alerta
   * @param status Status do alerta
   */
  withAlertStatus(status: CreditAlertStatus): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).alertStatus = status;
    return this;
  }
  
  /**
   * Define a data de expiração
   * @param date Data de expiração
   */
  withExpirationDate(date: Date): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).expiresAt = date;
    return this;
  }
  
  /**
   * Define valores anteriores e novos
   * @param previous Valor anterior
   * @param newValue Novo valor
   */
  withValueChange(previous: number | string, newValue: number | string): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).previousValue = previous;
    (this.event as BureauCreditEvent).newValue = newValue;
    return this;
  }
  
  /**
   * Adiciona uma identidade relacionada
   * @param identityId ID da identidade
   */
  withRelatedIdentity(identityId: string): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).relatedIdentityIds.push(identityId);
    return this;
  }
  
  /**
   * Define múltiplas identidades relacionadas
   * @param identityIds Lista de IDs de identidades
   */
  withRelatedIdentities(identityIds: string[]): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).relatedIdentityIds = identityIds;
    return this;
  }
  
  /**
   * Define dados de relatório de crédito
   * @param reportData Dados do relatório
   */
  withCreditReportData(reportData: BureauCreditEvent['creditReportData']): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).creditReportData = reportData;
    return this;
  }
  
  /**
   * Define dados de fraude
   * @param fraudData Dados de fraude
   */
  withFraudData(fraudData: BureauCreditEvent['fraudData']): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).fraudData = fraudData;
    return this;
  }
  
  /**
   * Define dados de disputa
   * @param disputeData Dados de disputa
   */
  withDisputeData(disputeData: BureauCreditEvent['disputeData']): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).disputeData = disputeData;
    return this;
  }
  
  /**
   * Define informações regulatórias
   * @param regulatoryInfo Informações regulatórias
   */
  withRegulatoryInfo(regulatoryInfo: BureauCreditEvent['regulatoryInfo']): BureauCreditEventBuilder {
    (this.event as BureauCreditEvent).regulatoryInfo = regulatoryInfo;
    return this;
  }
  
  /**
   * Constrói e retorna o evento
   * @throws Error se informações obrigatórias estiverem faltando
   */
  build(): BureauCreditEvent {
    // Verificações específicas para eventos do bureau
    if (!(this.event as BureauCreditEvent).eventType) {
      throw new Error('Tipo de evento do Bureau de Créditos é obrigatório');
    }
    
    if (!(this.event as BureauCreditEvent).severity) {
      throw new Error('Severidade do evento é obrigatória');
    }
    
    if (!(this.event as BureauCreditEvent).relatedIdentityIds || 
        (this.event as BureauCreditEvent).relatedIdentityIds.length === 0) {
      throw new Error('Pelo menos uma identidade relacionada é obrigatória');
    }
    
    return super.build() as BureauCreditEvent;
  }
  
  /**
   * Cria um evento de alerta de fraude
   * @param identityId ID da identidade afetada
   * @param fraudType Tipo de fraude
   * @param riskLevel Nível de risco
   * @returns Evento de alerta de fraude
   */
  static createFraudAlertEvent(
    identityId: string, 
    fraudType: string, 
    riskLevel: string
  ): BureauCreditEvent {
    return new BureauCreditEventBuilder()
      .withEventType(BureauCreditEventType.FRAUD_ALERT)
      .withSeverity(CreditEventSeverity.HIGH)
      .withRelatedIdentity(identityId)
      .withAlertStatus(CreditAlertStatus.NEW)
      .withFraudData({
        fraudType,
        riskLevel,
        detectionMethod: 'bureau-analysis',
        confidenceScore: 0.85,
        recommendedActions: ['verify-identity', 'block-transactions', 'contact-customer']
      })
      .withAcknowledgmentRequired(true)
      .build();
  }
  
  /**
   * Cria um evento de alteração de score de crédito
   * @param identityId ID da identidade
   * @param previousScore Score anterior
   * @param newScore Novo score
   * @returns Evento de alteração de score
   */
  static createScoreChangedEvent(
    identityId: string, 
    previousScore: number, 
    newScore: number
  ): BureauCreditEvent {
    const scoreDiff = newScore - previousScore;
    let severity = CreditEventSeverity.INFORMATIONAL;
    
    // Determina severidade com base na mudança de score
    if (Math.abs(scoreDiff) >= 100) {
      severity = CreditEventSeverity.HIGH;
    } else if (Math.abs(scoreDiff) >= 50) {
      severity = CreditEventSeverity.MEDIUM;
    } else if (Math.abs(scoreDiff) >= 20) {
      severity = CreditEventSeverity.LOW;
    }
    
    return new BureauCreditEventBuilder()
      .withEventType(BureauCreditEventType.SCORE_CHANGED)
      .withSeverity(severity)
      .withRelatedIdentity(identityId)
      .withValueChange(previousScore, newScore)
      .withCreditReportData({
        reportId: uuidv4(),
        reportDate: new Date(),
        reportType: 'score-update',
        score: newScore,
        maxScore: 1000,
        scoreBand: newScore >= 800 ? 'EXCELENTE' : 
                  newScore >= 700 ? 'MUITO_BOM' :
                  newScore >= 600 ? 'BOM' :
                  newScore >= 500 ? 'REGULAR' : 'BAIXO'
      })
      .build();
  }
}