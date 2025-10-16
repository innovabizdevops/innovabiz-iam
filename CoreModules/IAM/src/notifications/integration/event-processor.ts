/**
 * @file event-processor.ts
 * @description Definições para processadores de eventos do sistema de notificações
 * 
 * Define a interface para processadores de eventos que transformam eventos
 * de sistema em notificações para os usuários.
 */

import { BaseEvent } from '../core/base-event';

/**
 * Tipo para função de processamento de eventos
 */
export type EventProcessor = (event: BaseEvent) => Promise<any>;

/**
 * Interface para eventos específicos do módulo IAM
 */
export interface IamEvent extends BaseEvent {
  module: 'iam';
  category: 'authentication' | 'authorization' | 'user-management';
  type?: 'login' | 'login_failed' | 'logout' | 'password_reset' | 'account_locked' | 
         'account_unlocked' | 'mfa_enabled' | 'mfa_disabled' | 'registration' | 
         'email_verification' | 'phone_verification' | 'permission_granted' | 
         'permission_revoked' | 'role_assigned' | 'role_removed';
  data: {
    userId: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo Payment Gateway
 */
export interface PaymentGatewayEvent extends BaseEvent {
  module: 'payment-gateway';
  category: 'payment' | 'refund' | 'dispute' | 'payout';
  type?: 'payment_success' | 'payment_failed' | 'payment_pending' | 'payment_expired' |
         'refund_initiated' | 'refund_completed' | 'dispute_opened' | 'dispute_resolved' |
         'payout_initiated' | 'payout_completed' | 'payout_failed';
  data: {
    transactionId: string;
    amount: number;
    currency: string;
    buyerId?: string;
    buyerEmail?: string;
    buyerPhone?: string;
    buyerName?: string;
    sellerId?: string;
    sellerEmail?: string;
    sellerPhone?: string;
    sellerName?: string;
    status: string;
    paymentMethod?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo Mobile Money
 */
export interface MobileMoneyEvent extends BaseEvent {
  module: 'mobile-money';
  category: 'transaction' | 'wallet' | 'agent';
  type?: 'deposit' | 'withdrawal' | 'transfer' | 'bill_payment' | 'merchant_payment' |
         'wallet_created' | 'wallet_locked' | 'wallet_unlocked' | 'balance_inquiry' |
         'agent_commission' | 'agent_float_update';
  data: {
    transactionId?: string;
    walletId?: string;
    agentId?: string;
    userId: string;
    userPhone?: string;
    userEmail?: string;
    amount?: number;
    currency?: string;
    status?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo E-Commerce
 */
export interface ECommerceEvent extends BaseEvent {
  module: 'e-commerce';
  category: 'order' | 'product' | 'cart' | 'shipping' | 'inventory';
  type?: 'order_created' | 'order_confirmed' | 'order_shipped' | 'order_delivered' |
         'order_cancelled' | 'product_added' | 'product_updated' | 'product_removed' |
         'cart_updated' | 'cart_abandoned' | 'shipping_update' | 'inventory_low';
  data: {
    orderId?: string;
    productId?: string;
    cartId?: string;
    customerId: string;
    customerEmail?: string;
    customerPhone?: string;
    customerName?: string;
    amount?: number;
    currency?: string;
    status?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo Bureau de Crédito
 */
export interface BureauCreditoEvent extends BaseEvent {
  module: 'bureau-credito';
  category: 'credit-check' | 'score' | 'report';
  type?: 'credit_check_completed' | 'credit_check_failed' | 'score_updated' |
         'report_available' | 'report_requested' | 'credit_alert';
  data: {
    requestId?: string;
    reportId?: string;
    userId: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    score?: number;
    status?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo de Segurança
 */
export interface SecurityEvent extends BaseEvent {
  module: 'security';
  category: 'alert' | 'compliance' | 'audit';
  type?: 'suspicious_activity' | 'threat_detected' | 'compliance_violation' |
         'audit_completed' | 'security_update' | 'policy_violation';
  data: {
    alertId?: string;
    userId?: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    severity?: 'low' | 'medium' | 'high' | 'critical';
    status?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo de Compliance
 */
export interface ComplianceEvent extends BaseEvent {
  module: 'compliance';
  category: 'kyc' | 'aml' | 'regulation';
  type?: 'kyc_approved' | 'kyc_rejected' | 'kyc_pending' | 'aml_alert' |
         'regulatory_update' | 'compliance_report';
  data: {
    caseId?: string;
    userId?: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    status?: string;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo GenAI
 */
export interface GenAIEvent extends BaseEvent {
  module: 'genai';
  category: 'recommendation' | 'prediction' | 'insight';
  type?: 'recommendation_ready' | 'prediction_completed' | 'insight_generated' |
         'model_trained' | 'anomaly_detected';
  data: {
    analysisId?: string;
    userId?: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    modelType?: string;
    confidence?: number;
    [key: string]: any;
  };
}

/**
 * Interface para eventos específicos do módulo Open Ecosystem
 */
export interface OpenEcosystemEvent extends BaseEvent {
  module: 'open-ecosystem';
  category: 'banking' | 'finance' | 'insurance' | 'marketplace';
  type?: 'account_linked' | 'consent_granted' | 'consent_revoked' |
         'data_shared' | 'api_access' | 'service_connected';
  data: {
    serviceId?: string;
    userId: string;
    userEmail?: string;
    userPhone?: string;
    userName?: string;
    providerId?: string;
    providerName?: string;
    status?: string;
    [key: string]: any;
  };
}

/**
 * Mapa de tipos de eventos por módulo e categoria
 */
export const EVENT_TYPES = {
  iam: {
    authentication: [
      'login', 'login_failed', 'logout', 'password_reset', 'account_locked', 
      'account_unlocked', 'mfa_enabled', 'mfa_disabled', 'registration', 
      'email_verification', 'phone_verification'
    ],
    authorization: [
      'permission_granted', 'permission_revoked', 'role_assigned', 'role_removed'
    ],
    'user-management': [
      'user_created', 'user_updated', 'user_deleted', 'profile_updated'
    ]
  },
  'payment-gateway': {
    payment: [
      'payment_success', 'payment_failed', 'payment_pending', 'payment_expired'
    ],
    refund: [
      'refund_initiated', 'refund_completed', 'refund_failed'
    ],
    dispute: [
      'dispute_opened', 'dispute_updated', 'dispute_resolved'
    ],
    payout: [
      'payout_initiated', 'payout_completed', 'payout_failed'
    ]
  },
  'mobile-money': {
    transaction: [
      'deposit', 'withdrawal', 'transfer', 'bill_payment', 'merchant_payment'
    ],
    wallet: [
      'wallet_created', 'wallet_locked', 'wallet_unlocked', 'balance_inquiry'
    ],
    agent: [
      'agent_commission', 'agent_float_update', 'agent_transaction'
    ]
  },
  'e-commerce': {
    order: [
      'order_created', 'order_confirmed', 'order_shipped', 'order_delivered',
      'order_cancelled'
    ],
    product: [
      'product_added', 'product_updated', 'product_removed'
    ],
    cart: [
      'cart_updated', 'cart_abandoned', 'cart_checkout'
    ],
    shipping: [
      'shipping_update', 'delivery_scheduled', 'delivery_delayed'
    ],
    inventory: [
      'inventory_low', 'inventory_updated', 'stock_replenished'
    ]
  },
  'bureau-credito': {
    'credit-check': [
      'credit_check_completed', 'credit_check_failed'
    ],
    score: [
      'score_updated', 'score_improved', 'score_declined'
    ],
    report: [
      'report_available', 'report_requested', 'credit_alert'
    ]
  }
};

/**
 * Verifica se um evento é válido
 * @param event Evento a ser validado
 * @returns True se o evento for válido, false caso contrário
 */
export function validateEvent(event: BaseEvent): boolean {
  // Verificar propriedades obrigatórias
  if (!event.id || !event.timestamp || !event.module || !event.category) {
    return false;
  }
  
  // Verificar se o módulo é suportado
  if (!(event.module in EVENT_TYPES)) {
    return false;
  }
  
  // Verificar se a categoria é suportada para o módulo
  if (!(event.category in EVENT_TYPES[event.module as keyof typeof EVENT_TYPES])) {
    return false;
  }
  
  // Se o tipo for especificado, verificar se é válido para o módulo e categoria
  if (event.type) {
    const validTypes = EVENT_TYPES[event.module as keyof typeof EVENT_TYPES][
      event.category as keyof (typeof EVENT_TYPES)[keyof typeof EVENT_TYPES]
    ];
    
    if (!validTypes.includes(event.type)) {
      return false;
    }
  }
  
  return true;
}