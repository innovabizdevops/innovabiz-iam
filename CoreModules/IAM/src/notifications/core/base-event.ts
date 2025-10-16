/**
 * @file base-event.ts
 * @description Interface base para eventos do sistema
 * 
 * Define a estrutura básica para eventos que podem gerar notificações
 * em todo o ecossistema InnovaBiz.
 */

import { NotificationRecipient } from '../adapters/notification-adapter';

/**
 * Interface base para eventos do sistema
 */
export interface BaseEvent {
  /**
   * Identificador único do evento
   */
  id: string;
  
  /**
   * Timestamp de ocorrência do evento
   */
  timestamp: Date;
  
  /**
   * Módulo que originou o evento
   */
  module: string;
  
  /**
   * Categoria do evento dentro do módulo
   */
  category: string;
  
  /**
   * Tipo específico do evento (opcional)
   */
  type?: string;
  
  /**
   * Dados específicos do evento
   */
  data?: Record<string, any>;
  
  /**
   * Contexto do evento (ex: ambiente, tenant)
   */
  context?: {
    /**
     * ID do tenant
     */
    tenantId?: string;
    
    /**
     * Ambiente (dev, staging, prod)
     */
    environment?: string;
    
    /**
     * Origem da requisição
     */
    source?: string;
    
    /**
     * Localidade/idioma relacionado ao evento
     */
    locale?: string;
    
    /**
     * Metadados adicionais de contexto
     */
    metadata?: Record<string, any>;
  };
  
  /**
   * Criticidade do evento
   */
  severity?: 'low' | 'medium' | 'high' | 'critical';
  
  /**
   * Destinatários explícitos do evento
   */
  recipients?: NotificationRecipient[];
  
  /**
   * Metadados adicionais do evento
   */
  metadata?: Record<string, any>;
}