/**
 * Serviço para gestão de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import { api } from './api';
import { Rule, RuleSet, RuleFilter, RuleEvaluationResult } from '../types/rules';

/**
 * Base URL para os endpoints de regras
 */
const RULES_BASE_URL = '/rules';
const RULESETS_BASE_URL = '/rulesets';

/**
 * Serviço para operações de regras dinâmicas
 */
export const rulesService = {
  // Operações de regras
  /**
   * Obtém todas as regras com filtros opcionais
   * 
   * @param filters Filtros para as regras
   * @returns Promise com lista de regras
   */
  getRules: (filters?: RuleFilter) => {
    let queryParams = '';
    
    if (filters) {
      const params = new URLSearchParams();
      
      if (filters.region) {
        params.append('region', filters.region);
      }
      
      if (filters.tags && filters.tags.length > 0) {
        params.append('tags', filters.tags.join(','));
      }
      
      queryParams = params.toString();
      if (queryParams) {
        queryParams = `?${queryParams}`;
      }
    }
    
    return api.get<Rule[]>(`${RULES_BASE_URL}${queryParams}`);
  },
  
  /**
   * Obtém uma regra pelo ID
   * 
   * @param id ID da regra
   * @returns Promise com a regra
   */
  getRule: (id: string) => {
    return api.get<Rule>(`${RULES_BASE_URL}/${id}`);
  },
  
  /**
   * Cria uma nova regra
   * 
   * @param rule Dados da regra
   * @returns Promise com a regra criada
   */
  createRule: (rule: Partial<Rule>) => {
    return api.post<Rule>(RULES_BASE_URL, rule);
  },
  
  /**
   * Atualiza uma regra existente
   * 
   * @param id ID da regra
   * @param rule Dados atualizados da regra
   * @returns Promise com a regra atualizada
   */
  updateRule: (id: string, rule: Partial<Rule>) => {
    return api.put<Rule>(`${RULES_BASE_URL}/${id}`, rule);
  },
  
  /**
   * Exclui uma regra
   * 
   * @param id ID da regra
   * @returns Promise com confirmação de exclusão
   */
  deleteRule: (id: string) => {
    return api.delete<{ message: string }>(`${RULES_BASE_URL}/${id}`);
  },
  
  /**
   * Testa uma regra com dados de evento
   * 
   * @param id ID da regra
   * @param eventData Dados do evento para teste
   * @returns Promise com o resultado do teste
   */
  testRule: (id: string, eventData: any) => {
    return api.post<RuleEvaluationResult>(`${RULES_BASE_URL}/${id}/test`, eventData);
  },
  
  // Operações de conjuntos de regras
  /**
   * Obtém todos os conjuntos de regras com filtros opcionais
   * 
   * @param filters Filtros para os conjuntos
   * @returns Promise com lista de conjuntos
   */
  getRuleSets: (filters?: RuleFilter) => {
    let queryParams = '';
    
    if (filters) {
      const params = new URLSearchParams();
      
      if (filters.region) {
        params.append('region', filters.region);
      }
      
      if (filters.tags && filters.tags.length > 0) {
        params.append('tags', filters.tags.join(','));
      }
      
      queryParams = params.toString();
      if (queryParams) {
        queryParams = `?${queryParams}`;
      }
    }
    
    return api.get<RuleSet[]>(`${RULESETS_BASE_URL}${queryParams}`);
  },
  
  /**
   * Obtém um conjunto de regras pelo ID
   * 
   * @param id ID do conjunto
   * @returns Promise com o conjunto
   */
  getRuleSet: (id: string) => {
    return api.get<RuleSet>(`${RULESETS_BASE_URL}/${id}`);
  },
  
  /**
   * Cria um novo conjunto de regras
   * 
   * @param ruleSet Dados do conjunto
   * @returns Promise com o conjunto criado
   */
  createRuleSet: (ruleSet: Partial<RuleSet>) => {
    return api.post<RuleSet>(RULESETS_BASE_URL, ruleSet);
  },
  
  /**
   * Atualiza um conjunto de regras existente
   * 
   * @param id ID do conjunto
   * @param ruleSet Dados atualizados do conjunto
   * @returns Promise com o conjunto atualizado
   */
  updateRuleSet: (id: string, ruleSet: Partial<RuleSet>) => {
    return api.put<RuleSet>(`${RULESETS_BASE_URL}/${id}`, ruleSet);
  },
  
  /**
   * Exclui um conjunto de regras
   * 
   * @param id ID do conjunto
   * @returns Promise com confirmação de exclusão
   */
  deleteRuleSet: (id: string) => {
    return api.delete<{ message: string }>(`${RULESETS_BASE_URL}/${id}`);
  },
  
  /**
   * Testa um conjunto de regras com dados de evento
   * 
   * @param id ID do conjunto
   * @param eventData Dados do evento para teste
   * @returns Promise com o resultado do teste
   */
  testRuleSet: (id: string, eventData: any) => {
    return api.post<any>(`${RULESETS_BASE_URL}/${id}/test`, eventData);
  },
  
  /**
   * Obtém estatísticas de regras para dashboard
   * 
   * @param region Região opcional
   * @param days Quantidade de dias (padrão: 7)
   * @returns Promise com estatísticas
   */
  getRuleStatistics: (region?: string, days: number = 7) => {
    let url = '/dashboard/rule_statistics';
    const params = new URLSearchParams();
    
    if (region) {
      params.append('region', region);
    }
    
    params.append('days', days.toString());
    
    return api.get<any>(`${url}?${params.toString()}`);
  },
};

export default rulesService;