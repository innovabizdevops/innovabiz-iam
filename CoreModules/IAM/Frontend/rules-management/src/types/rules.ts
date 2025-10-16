/**
 * Tipos e interfaces para o sistema de regras dinâmicas
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

/**
 * Operadores para condições de regras
 */
export enum RuleOperator {
  // Operadores de igualdade
  EQUALS = "equals",
  NOT_EQUALS = "not_equals",
  
  // Operadores de comparação
  GREATER_THAN = "gt",
  GREATER_THAN_OR_EQUALS = "gte",
  LESS_THAN = "lt",
  LESS_THAN_OR_EQUALS = "lte",
  
  // Operadores de texto
  CONTAINS = "contains",
  NOT_CONTAINS = "not_contains",
  STARTS_WITH = "starts_with",
  ENDS_WITH = "ends_with",
  REGEX = "regex",
  
  // Operadores de coleção
  IN = "in",
  NOT_IN = "not_in",
  
  // Operadores de existência
  EXISTS = "exists",
  NOT_EXISTS = "not_exists",
  
  // Operadores geográficos
  GEO_DISTANCE = "geo_distance",
  
  // Operadores de tempo
  TIME_RANGE = "time_range",
  TIME_AFTER = "time_after",
  TIME_BEFORE = "time_before",
}

/**
 * Operadores lógicos para grupos de condições
 */
export enum RuleLogicalOperator {
  AND = "and",
  OR = "or",
  NOT = "not",
}

/**
 * Tipos de valor para condições de regras
 */
export enum RuleValueType {
  STRING = "string",
  NUMBER = "number",
  BOOLEAN = "boolean",
  ARRAY = "array",
  OBJECT = "object",
  NULL = "null",
  UNDEFINED = "undefined",
  REGEX = "regex",
  REFERENCE = "reference",
  FUNCTION = "function",
}

/**
 * Severidade de regras
 */
export enum RuleSeverity {
  CRITICAL = "critical",
  HIGH = "high",
  MEDIUM = "medium",
  LOW = "low",
  INFO = "info",
}

/**
 * Categorias de regras
 */
export enum RuleCategory {
  AUTHENTICATION = "authentication",
  AUTHORIZATION = "authorization",
  ACCOUNT = "account",
  TRANSACTION = "transaction",
  PAYMENT = "payment",
  SESSION = "session",
  DEVICE = "device",
  LOCATION = "location",
  NETWORK = "network",
  BEHAVIORAL = "behavioral",
  CUSTOM = "custom",
}

/**
 * Ações associadas a regras
 */
export enum RuleAction {
  LOG = "log",
  ALERT = "alert",
  NOTIFY = "notify",
  BLOCK = "block",
  CHALLENGE = "challenge",
  ESCALATE = "escalate",
  RESTRICT = "restrict",
  CUSTOM = "custom",
}

/**
 * Condição de regra
 */
export interface RuleCondition {
  id?: string;
  field: string;
  operator: RuleOperator;
  value: any;
  valueType?: RuleValueType;
  description?: string;
}

/**
 * Grupo de condições de regra
 */
export interface RuleGroup {
  id?: string;
  operator: RuleLogicalOperator;
  conditions?: RuleCondition[];
  groups?: RuleGroup[];
  description?: string;
}

/**
 * Regra para detecção de anomalias
 */
export interface Rule {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  severity: RuleSeverity;
  category: RuleCategory;
  region?: string;
  tags?: string[];
  version?: string;
  createdAt?: string;
  updatedAt?: string;
  createdBy?: string;
  updatedBy?: string;
  condition?: RuleCondition;
  group?: RuleGroup;
  actions: RuleAction[];
  actionConfig?: Record<string, any>;
  score?: number;
}

/**
 * Conjunto de regras
 */
export interface RuleSet {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  region?: string;
  tags?: string[];
  rules: string[];  // IDs das regras
  version?: string;
  createdAt?: string;
  updatedAt?: string;
  createdBy?: string;
  updatedBy?: string;
}

/**
 * Resultado da avaliação de regra
 */
export interface RuleEvaluationResult {
  rule: Rule;
  matched: boolean;
  score: number;
  actions: RuleAction[];
  matchedFields?: Record<string, any>;
  timestamp?: string;
  evaluationTime?: number;
}

/**
 * Filtro para listagem de regras
 */
export interface RuleFilter {
  region?: string;
  tags?: string[];
  category?: RuleCategory;
  severity?: RuleSeverity;
  enabled?: boolean;
  searchTerm?: string;
}