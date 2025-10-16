/**
 * ==============================================================================
 * Nome: bureau-credito.ts
 * Descrição: Definição de tipos para integração do IAM com Bureau de Créditos
 * Autor: Equipa de Desenvolvimento INNOVABIZ
 * Data: 19/08/2025
 * ==============================================================================
 */

/**
 * Tipo de vínculo com o Bureau de Créditos
 */
export enum BureauVinculoTipo {
  // Acesso apenas para consultas
  CONSULTA = 'CONSULTA',
  // Acesso para registro e gestão de informações
  GESTAO = 'GESTAO',
  // Acesso administrativo completo
  ADMIN = 'ADMIN',
  // Acesso específico para auditoria
  AUDITORIA = 'AUDITORIA',
}

/**
 * Nível de acesso ao Bureau de Créditos
 */
export enum BureauNivelAcesso {
  // Acesso básico (dados limitados)
  BASICO = 'BASICO',
  // Acesso intermediário (maioria dos dados)
  INTERMEDIARIO = 'INTERMEDIARIO',
  // Acesso completo (todos os dados)
  COMPLETO = 'COMPLETO',
}

/**
 * Status de vínculo com o Bureau de Créditos
 */
export enum BureauVinculoStatus {
  ATIVO = 'ATIVO',
  INATIVO = 'INATIVO',
  SUSPENSO = 'SUSPENSO',
  PENDENTE_APROVACAO = 'PENDENTE_APROVACAO',
}

/**
 * Tipo de consulta ao Bureau de Créditos
 */
export enum BureauTipoConsulta {
  // Consulta simples (score e situação geral)
  SIMPLES = 'SIMPLES',
  // Consulta detalhada (histórico completo)
  DETALHADA = 'DETALHADA',
  // Consulta para análise de crédito
  ANALISE_CREDITO = 'ANALISE_CREDITO',
  // Consulta para prevenção à fraude
  PREVENCAO_FRAUDE = 'PREVENCAO_FRAUDE',
  // Consulta para validação de identidade
  VALIDACAO_IDENTIDADE = 'VALIDACAO_IDENTIDADE',
  // Consulta para conformidade regulatória
  CONFORMIDADE = 'CONFORMIDADE',
  // Consulta para monitoramento contínuo
  MONITORAMENTO = 'MONITORAMENTO',
}

/**
 * Status de uma autorização de consulta
 */
export enum BureauAutorizacaoStatus {
  ATIVA = 'ATIVA',
  EXPIRADA = 'EXPIRADA',
  REVOGADA = 'REVOGADA',
  PENDENTE = 'PENDENTE',
}

/**
 * Interface para vínculo de identidade do usuário com o Bureau de Créditos
 */
export interface BureauIdentity {
  // ID único da identidade de integração
  id: string;
  // ID do usuário no IAM
  usuarioId: string;
  // ID do tenant no IAM
  tenantId: string;
  // ID externo no Bureau de Créditos
  externalId: string;
  // ID do tenant no Bureau de Créditos
  externalTenantId: string;
  // Tipo de vínculo
  tipoVinculo: BureauVinculoTipo;
  // Nível de acesso
  nivelAcesso: BureauNivelAcesso;
  // Status do vínculo
  status: BureauVinculoStatus;
  // Data de criação
  dataCriacao: Date;
  // Data da última atualização
  dataAtualizacao?: Date;
  // Motivo do status atual (se aplicável)
  motivoStatus?: string;
  // Detalhes adicionais (JSON)
  detalhes?: Record<string, any>;
}

/**
 * Interface para autorização de consulta ao Bureau de Créditos
 */
export interface BureauAutorizacao {
  // ID único da autorização
  id: string;
  // ID da identidade vinculada
  identityId: string;
  // Tipo de consulta autorizada
  tipoConsulta: BureauTipoConsulta;
  // Finalidade da consulta
  finalidade: string;
  // Justificativa para a autorização
  justificativa: string;
  // Data de criação da autorização
  dataAutorizacao: Date;
  // Data de validade da autorização
  dataValidade: Date;
  // Status atual da autorização
  status: BureauAutorizacaoStatus;
  // Usuário que autorizou a consulta
  autorizadoPor: string;
}

/**
 * Interface para token de acesso ao Bureau de Créditos
 */
export interface BureauAccessToken {
  // Token de acesso
  token: string;
  // Tipo de token (geralmente "bearer")
  type: string;
  // Data de expiração do token
  expiresAt: Date;
  // Token de atualização (se disponível)
  refreshToken?: string;
  // ID da identidade vinculada
  identityId: string;
  // Escopos do token
  escopos: string[];
  // Finalidade do token
  finalidade: string;
}

/**
 * Entrada para criação de vínculo com Bureau de Créditos
 */
export interface CriarVinculoBureauInput {
  // ID do usuário no IAM
  usuarioId: string;
  // ID do tenant
  tenantId: string;
  // Tipo de vínculo a ser criado
  tipoVinculo: BureauVinculoTipo;
  // Nível de acesso solicitado
  nivelAcesso: BureauNivelAcesso;
  // Detalhes de autorização (JSON)
  detalhesAutorizacao?: Record<string, any>;
}

/**
 * Entrada para criação de autorização de consulta
 */
export interface CriarAutorizacaoBureauInput {
  // ID da identidade de integração
  identityId: string;
  // Tipo de consulta a ser autorizada
  tipoConsulta: BureauTipoConsulta;
  // Finalidade da consulta
  finalidade: string;
  // Justificativa para a autorização
  justificativa: string;
  // Duração em dias da autorização
  duracaoDias?: number;
  // Usuário que autoriza a consulta
  autorizadoPor: string;
}

/**
 * Entrada para geração de token de acesso
 */
export interface GerarTokenBureauInput {
  // ID da identidade de integração
  identityId: string;
  // Finalidade do token
  finalidade: string;
  // Escopos solicitados
  escopos: string[];
}

/**
 * Entrada para revogação de vínculo
 */
export interface RevogarVinculoBureauInput {
  // ID da identidade de integração
  identityId: string;
  // Motivo da revogação
  motivo: string;
}