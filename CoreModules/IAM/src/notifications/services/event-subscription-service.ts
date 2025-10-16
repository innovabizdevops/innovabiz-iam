/**
 * @file event-subscription-service.ts
 * @description Serviço de assinatura de eventos para notificações
 * 
 * Este serviço implementa o padrão Observer para gerenciar assinaturas de eventos,
 * permitindo que componentes do sistema se inscrevam para receber notificações
 * específicas quando determinados eventos ocorrerem, com suporte a filtros e
 * condições avançadas.
 */

import { BaseEvent, EventCategory, EventPriority } from '../core/base-event';
import { NotificationChannel } from '../core/notification-channel';
import { Logger } from '../../../infrastructure/observability/logger';
import { 
  BureauCreditEvent, 
  BureauCreditEventType, 
  CreditEventSeverity 
} from '../schemas/bureau-credit-events';

/**
 * Interface para filtro de eventos
 */
export interface EventFilter {
  // Filtros genéricos para eventos base
  eventId?: string;
  category?: EventCategory | EventCategory[];
  priority?: EventPriority | EventPriority[];
  source?: string | string[];
  code?: string | string[];
  
  // Filtros específicos para Bureau de Créditos
  bureauCreditType?: BureauCreditEventType | BureauCreditEventType[];
  severity?: CreditEventSeverity | CreditEventSeverity[];
  identityIds?: string[];
  
  // Filtros por metadados personalizados
  metadata?: Record<string, any>;
  
  // Função de filtro customizada
  customFilter?: (event: BaseEvent) => boolean;
}

/**
 * Interface para manipulador de eventos
 */
export interface EventHandler {
  /**
   * Identificador único do manipulador
   */
  handlerId: string;
  
  /**
   * Nome descritivo do manipulador
   */
  name: string;
  
  /**
   * Função de callback a ser executada quando o evento ocorrer
   * @param event Evento que disparou o callback
   */
  callback: (event: BaseEvent) => Promise<void>;
  
  /**
   * Filtros de eventos que este manipulador deve processar
   */
  filter?: EventFilter;
  
  /**
   * Canais de notificação preferidos para este manipulador
   */
  preferredChannels?: NotificationChannel[];
  
  /**
   * Indica se este manipulador está ativo
   */
  isActive: boolean;
  
  /**
   * Informações do assinante
   */
  subscriber: {
    id: string;
    type: 'USER' | 'SYSTEM' | 'SERVICE' | 'EXTERNAL';
    name?: string;
  };
  
  /**
   * Metadados adicionais do manipulador
   */
  metadata?: Record<string, any>;
}

/**
 * Serviço de assinatura de eventos
 */
export class EventSubscriptionService {
  private static instance: EventSubscriptionService;
  private eventHandlers: Map<string, EventHandler[]>;
  private logger: Logger;
  
  private constructor() {
    this.eventHandlers = new Map();
    this.logger = new Logger('EventSubscriptionService');
  }
  
  /**
   * Obtém a instância única do serviço (padrão Singleton)
   */
  public static getInstance(): EventSubscriptionService {
    if (!EventSubscriptionService.instance) {
      EventSubscriptionService.instance = new EventSubscriptionService();
    }
    return EventSubscriptionService.instance;
  }
  
  /**
   * Adiciona um manipulador de eventos ao serviço
   * @param eventType Tipo de evento a ser assinado
   * @param handler Manipulador de eventos
   * @returns ID do manipulador registrado
   */
  public subscribe(eventType: string, handler: EventHandler): string {
    this.logger.debug('Registrando novo manipulador de eventos', {
      eventType,
      handlerId: handler.handlerId,
      handlerName: handler.name
    });
    
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    
    const handlers = this.eventHandlers.get(eventType);
    handlers!.push(handler);
    
    return handler.handlerId;
  }
  
  /**
   * Cancela a assinatura de um manipulador de eventos
   * @param eventType Tipo de evento
   * @param handlerId ID do manipulador
   * @returns Verdadeiro se o manipulador foi encontrado e removido
   */
  public unsubscribe(eventType: string, handlerId: string): boolean {
    this.logger.debug('Cancelando assinatura de eventos', {
      eventType,
      handlerId
    });
    
    if (!this.eventHandlers.has(eventType)) {
      return false;
    }
    
    const handlers = this.eventHandlers.get(eventType);
    const initialLength = handlers!.length;
    
    const filteredHandlers = handlers!.filter(h => h.handlerId !== handlerId);
    this.eventHandlers.set(eventType, filteredHandlers);
    
    return filteredHandlers.length < initialLength;
  }
  
  /**
   * Pausa a assinatura de um manipulador de eventos (mantém registrado mas inativo)
   * @param eventType Tipo de evento
   * @param handlerId ID do manipulador
   * @returns Verdadeiro se o manipulador foi encontrado e pausado
   */
  public pauseSubscription(eventType: string, handlerId: string): boolean {
    this.logger.debug('Pausando assinatura de eventos', {
      eventType,
      handlerId
    });
    
    if (!this.eventHandlers.has(eventType)) {
      return false;
    }
    
    const handlers = this.eventHandlers.get(eventType);
    const handler = handlers!.find(h => h.handlerId === handlerId);
    
    if (handler) {
      handler.isActive = false;
      return true;
    }
    
    return false;
  }
  
  /**
   * Retoma a assinatura de um manipulador de eventos pausado
   * @param eventType Tipo de evento
   * @param handlerId ID do manipulador
   * @returns Verdadeiro se o manipulador foi encontrado e retomado
   */
  public resumeSubscription(eventType: string, handlerId: string): boolean {
    this.logger.debug('Retomando assinatura de eventos', {
      eventType,
      handlerId
    });
    
    if (!this.eventHandlers.has(eventType)) {
      return false;
    }
    
    const handlers = this.eventHandlers.get(eventType);
    const handler = handlers!.find(h => h.handlerId === handlerId);
    
    if (handler) {
      handler.isActive = true;
      return true;
    }
    
    return false;
  }
  
  /**
   * Atualiza os filtros de um manipulador de eventos
   * @param eventType Tipo de evento
   * @param handlerId ID do manipulador
   * @param newFilter Novos filtros
   * @returns Verdadeiro se o manipulador foi encontrado e atualizado
   */
  public updateFilter(eventType: string, handlerId: string, newFilter: EventFilter): boolean {
    this.logger.debug('Atualizando filtros de assinatura de eventos', {
      eventType,
      handlerId
    });
    
    if (!this.eventHandlers.has(eventType)) {
      return false;
    }
    
    const handlers = this.eventHandlers.get(eventType);
    const handler = handlers!.find(h => h.handlerId === handlerId);
    
    if (handler) {
      handler.filter = newFilter;
      return true;
    }
    
    return false;
  }
  
  /**
   * Obtém todos os manipuladores ativos para um tipo de evento
   * @param eventType Tipo de evento
   * @returns Lista de manipuladores ativos
   */
  public getActiveHandlers(eventType: string): EventHandler[] {
    if (!this.eventHandlers.has(eventType)) {
      return [];
    }
    
    return this.eventHandlers.get(eventType)!.filter(h => h.isActive);
  }
  
  /**
   * Filtra manipuladores para um evento específico
   * @param event Evento a ser processado
   * @returns Lista de manipuladores que devem processar o evento
   */
  public getMatchingHandlers(event: BaseEvent): EventHandler[] {
    // Determina o tipo de evento específico
    let eventType = '';
    
    // Caso seja um evento de Bureau de Créditos
    if ('eventType' in event) {
      const bureauEvent = event as BureauCreditEvent;
      eventType = bureauEvent.eventType;
    } else {
      // Para outros tipos de eventos, usa o código como tipo
      eventType = event.code;
    }
    
    // Obtém todos os manipuladores para este tipo de evento
    const handlers = this.getActiveHandlers(eventType);
    
    // Filtra manipuladores baseado em seus filtros
    return handlers.filter(handler => this.matchesFilter(event, handler.filter));
  }
  
  /**
   * Verifica se um evento atende aos critérios de um filtro
   * @param event Evento a verificar
   * @param filter Filtro a aplicar
   * @returns Verdadeiro se o evento passar pelos filtros
   */
  private matchesFilter(event: BaseEvent, filter?: EventFilter): boolean {
    // Se não houver filtro, o evento sempre passa
    if (!filter) {
      return true;
    }
    
    // Verifica filtros genéricos
    if (filter.eventId && event.eventId !== filter.eventId) {
      return false;
    }
    
    if (filter.category) {
      if (Array.isArray(filter.category)) {
        if (!filter.category.includes(event.category)) {
          return false;
        }
      } else if (event.category !== filter.category) {
        return false;
      }
    }
    
    if (filter.priority) {
      if (Array.isArray(filter.priority)) {
        if (!filter.priority.includes(event.priority)) {
          return false;
        }
      } else if (event.priority !== filter.priority) {
        return false;
      }
    }
    
    if (filter.source) {
      if (Array.isArray(filter.source)) {
        if (!filter.source.includes(event.source)) {
          return false;
        }
      } else if (event.source !== filter.source) {
        return false;
      }
    }
    
    if (filter.code) {
      if (Array.isArray(filter.code)) {
        if (!filter.code.includes(event.code)) {
          return false;
        }
      } else if (event.code !== filter.code) {
        return false;
      }
    }
    
    // Verifica filtros específicos para Bureau de Créditos
    if ('eventType' in event && filter.bureauCreditType) {
      const bureauEvent = event as BureauCreditEvent;
      
      if (Array.isArray(filter.bureauCreditType)) {
        if (!filter.bureauCreditType.includes(bureauEvent.eventType)) {
          return false;
        }
      } else if (bureauEvent.eventType !== filter.bureauCreditType) {
        return false;
      }
    }
    
    if ('severity' in event && filter.severity) {
      const bureauEvent = event as BureauCreditEvent;
      
      if (Array.isArray(filter.severity)) {
        if (!filter.severity.includes(bureauEvent.severity)) {
          return false;
        }
      } else if (bureauEvent.severity !== filter.severity) {
        return false;
      }
    }
    
    // Verifica IDs de identidade relacionados
    if (filter.identityIds && filter.identityIds.length > 0) {
      if ('relatedIdentityIds' in event) {
        const bureauEvent = event as BureauCreditEvent;
        if (!bureauEvent.relatedIdentityIds || 
            !bureauEvent.relatedIdentityIds.some(id => 
              filter.identityIds!.includes(id))) {
          return false;
        }
      } else {
        // Se o evento não tiver IDs de identidade relacionados, não passa pelo filtro
        return false;
      }
    }
    
    // Verifica filtros de metadados
    if (filter.metadata && Object.keys(filter.metadata).length > 0) {
      if (!event.metadata) {
        return false;
      }
      
      // Verifica se todos os metadados do filtro estão presentes e iguais no evento
      for (const [key, value] of Object.entries(filter.metadata)) {
        if (event.metadata[key] !== value) {
          return false;
        }
      }
    }
    
    // Aplica filtro personalizado, se existir
    if (filter.customFilter && !filter.customFilter(event)) {
      return false;
    }
    
    // Se passou por todos os filtros, retorna verdadeiro
    return true;
  }
  
  /**
   * Processa um evento, notificando todos os manipuladores compatíveis
   * @param event Evento a ser processado
   */
  public async processEvent(event: BaseEvent): Promise<void> {
    const matchingHandlers = this.getMatchingHandlers(event);
    
    this.logger.debug('Processando evento', {
      eventId: event.eventId,
      category: event.category,
      code: event.code,
      matchingHandlersCount: matchingHandlers.length
    });
    
    const notificationPromises = matchingHandlers.map(async (handler) => {
      try {
        await handler.callback(event);
      } catch (error) {
        this.logger.error('Erro ao processar evento no manipulador', {
          eventId: event.eventId,
          handlerId: handler.handlerId,
          handlerName: handler.name,
          error
        });
      }
    });
    
    await Promise.allSettled(notificationPromises);
  }
}