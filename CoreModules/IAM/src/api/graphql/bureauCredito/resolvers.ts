/**
 * ==============================================================================
 * Nome: resolvers.ts
 * Descrição: Resolvers GraphQL para integração do IAM com Bureau de Créditos
 * Autor: Equipa de Desenvolvimento INNOVABIZ
 * Data: 19/08/2025
 * ==============================================================================
 */

import { IResolvers } from '@graphql-tools/utils';
import { GraphQLScalarType } from 'graphql';
import { GraphQLError } from 'graphql/error';
import { DateTimeResolver, JSONResolver } from 'graphql-scalars';

import { BureauCreditoConnector } from '../../../../integration/bureau-credito/bureau_credito_connector';
import { DataCoreClient } from '@innovabiz/datacore-client';
import { AuthorizationService } from '../../../services/authorization';
import { AuditService } from '../../../services/audit';
import { Context } from '../../../types/context';
import { 
  BureauIdentity, 
  BureauAutorizacao, 
  BureauAccessToken 
} from '../../../types/bureau-credito';

/**
 * Resolvers para o módulo Bureau de Créditos
 * Implementa todas as queries e mutations definidas no schema GraphQL
 */
export const bureauCreditoResolvers: IResolvers = {
  // Definição de escalares customizados
  DateTime: DateTimeResolver,
  JSON: JSONResolver,

  // Resolvers para o namespace BureauCreditoQueries
  BureauCreditoQueries: {
    // Buscar um vínculo específico por ID
    bureauIdentity: async (_, { id }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:read', 
        { identityId: id }
      );

      // Registrar auditoria de acesso
      await context.services.audit.logAccess(
        context.currentUser.id,
        'BUREAU_IDENTITY_ACCESS',
        { identityId: id },
        'bureau_credito'
      );

      // Buscar identidade via DataCore
      const identity = await context.services.dataCore.getIntegrationIdentity(id);
      if (!identity || identity.integrationType !== 'bureau_credito') {
        throw new GraphQLError('Identidade de Bureau de Créditos não encontrada', {
          extensions: { code: 'NOT_FOUND' }
        });
      }

      return mapToBureauIdentity(identity);
    },

    // Buscar vínculos por usuário
    bureauIdentitiesByUser: async (_, { usuarioId }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:list', 
        { userId: usuarioId }
      );

      // Registrar auditoria de acesso
      await context.services.audit.logAccess(
        context.currentUser.id,
        'BUREAU_IDENTITIES_BY_USER_ACCESS',
        { usuarioId },
        'bureau_credito'
      );

      // Buscar identidades via DataCore
      const identities = await context.services.dataCore.getIntegrationIdentitiesByUser(
        usuarioId, 
        'bureau_credito'
      );

      return identities.map(mapToBureauIdentity);
    },

    // Buscar vínculos por tenant
    bureauIdentitiesByTenant: async (_, { tenantId }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:list_tenant', 
        { tenantId }
      );

      // Registrar auditoria de acesso
      await context.services.audit.logAccess(
        context.currentUser.id,
        'BUREAU_IDENTITIES_BY_TENANT_ACCESS',
        { tenantId },
        'bureau_credito'
      );

      // Buscar identidades via DataCore
      const identities = await context.services.dataCore.getIntegrationIdentitiesByTenant(
        tenantId, 
        'bureau_credito'
      );

      return identities.map(mapToBureauIdentity);
    },

    // Buscar autorizações para uma identidade
    bureauAutorizacoes: async (_, { identityId }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:read_autorizacoes', 
        { identityId }
      );

      // Registrar auditoria de acesso
      await context.services.audit.logAccess(
        context.currentUser.id,
        'BUREAU_AUTORIZACOES_ACCESS',
        { identityId },
        'bureau_credito'
      );

      // Buscar autorizações via DataCore
      const autorizacoes = await context.services.dataCore.getBureauAutorizacoes(identityId);
      return autorizacoes.map(mapToBureauAutorizacao);
    },

    // Verificar se um usuário tem vínculos ativos
    hasActiveBureauIdentity: async (_, { usuarioId, tenantId }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:verify', 
        { userId: usuarioId }
      );

      // Buscar identidades ativas via DataCore
      const identities = await context.services.dataCore.getActiveIntegrationIdentitiesByUser(
        usuarioId, 
        'bureau_credito',
        tenantId
      );

      return identities.length > 0;
    },

    // Verificar se um usuário tem autorizações para determinado tipo de consulta
    hasValidBureauAutorizacao: async (_, { usuarioId, tipoConsulta }, context: Context) => {
      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:verify_autorizacao', 
        { userId: usuarioId }
      );

      // Buscar identidades ativas do usuário
      const identities = await context.services.dataCore.getActiveIntegrationIdentitiesByUser(
        usuarioId, 
        'bureau_credito'
      );

      // Se não há identidades ativas, não há autorizações
      if (identities.length === 0) {
        return false;
      }

      // Para cada identidade, verificar se há autorizações válidas para o tipo de consulta
      for (const identity of identities) {
        const autorizacoes = await context.services.dataCore.getValidBureauAutorizacoes(identity.id);
        
        // Verificar se alguma autorização válida corresponde ao tipo de consulta
        const hasValidAuth = autorizacoes.some(auth => 
          auth.tipoConsulta === tipoConsulta && auth.status === 'ativa'
        );

        if (hasValidAuth) {
          return true;
        }
      }

      return false;
    }
  },

  // Resolvers para o namespace BureauCreditoMutations
  BureauCreditoMutations: {
    // Criar um novo vínculo com Bureau de Créditos
    criarVinculoBureau: async (_, { input }, context: Context) => {
      const { usuarioId, tenantId, tipoVinculo, nivelAcesso, detalhesAutorizacao } = input;

      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:create_vinculo', 
        { userId: usuarioId, tenantId }
      );

      // Registrar auditoria de operação
      await context.services.audit.logMutation(
        context.currentUser.id,
        'BUREAU_CREATE_VINCULO',
        {
          usuarioId,
          tenantId,
          tipoVinculo,
          nivelAcesso
        },
        'bureau_credito',
        'alta' // Operações em Bureau de Créditos sempre têm alta severidade
      );

      // Criar vínculo via connector
      const identity = await context.services.bureauConnector.VincularUsuario(
        context.requestContext,
        usuarioId,
        tenantId,
        tipoVinculo,
        nivelAcesso,
        detalhesAutorizacao || {}
      );

      return mapToBureauIdentity(identity);
    },

    // Criar uma nova autorização de consulta
    criarAutorizacaoBureau: async (_, { input }, context: Context) => {
      const { 
        identityId, 
        tipoConsulta, 
        finalidade, 
        justificativa, 
        duracaoDias, 
        autorizadoPor 
      } = input;

      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:create_autorizacao', 
        { identityId }
      );

      // Registrar auditoria de operação
      await context.services.audit.logMutation(
        context.currentUser.id,
        'BUREAU_CREATE_AUTORIZACAO',
        {
          identityId,
          tipoConsulta,
          finalidade,
          justificativa,
          duracaoDias,
          autorizadoPor
        },
        'bureau_credito',
        'alta'
      );

      // Criar autorização via connector
      const autorizacao = await context.services.bureauConnector.CriarAutorizacaoConsulta(
        context.requestContext,
        identityId,
        tipoConsulta,
        finalidade,
        justificativa,
        duracaoDias || 30,
        autorizadoPor || context.currentUser.id
      );

      return mapToBureauAutorizacao(autorizacao);
    },

    // Gerar um token de acesso para Bureau de Créditos
    gerarTokenBureau: async (_, { input }, context: Context) => {
      const { identityId, finalidade, escopos } = input;

      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:generate_token', 
        { identityId }
      );

      // Registrar auditoria de operação
      await context.services.audit.logMutation(
        context.currentUser.id,
        'BUREAU_GENERATE_TOKEN',
        {
          identityId,
          finalidade,
          escopos
        },
        'bureau_credito',
        'alta'
      );

      // Gerar token via connector
      const token = await context.services.bureauConnector.GerarTokenAcesso(
        context.requestContext,
        identityId,
        finalidade,
        escopos
      );

      return mapToBureauAccessToken(token);
    },

    // Revogar um vínculo com Bureau de Créditos
    revogarVinculoBureau: async (_, { input }, context: Context) => {
      const { identityId, motivo } = input;

      // Verificar autorização
      await context.services.authorization.ensurePermission(
        context.currentUser, 
        'bureau_credito:revoke_vinculo', 
        { identityId }
      );

      // Registrar auditoria de operação
      await context.services.audit.logMutation(
        context.currentUser.id,
        'BUREAU_REVOKE_VINCULO',
        {
          identityId,
          motivo
        },
        'bureau_credito',
        'alta'
      );

      // Revogar vínculo via connector
      await context.services.bureauConnector.RevogarVinculo(
        context.requestContext,
        identityId,
        motivo
      );

      return true;
    }
  },

  // Extensão das queries raiz
  Query: {
    bureauCredito: () => ({})
  },

  // Extensão das mutations raiz
  Mutation: {
    bureauCredito: () => ({})
  }
};

/**
 * Funções de mapeamento de tipos do DataCore para tipos GraphQL
 */

// Mapear objeto de identidade do DataCore para o tipo BureauIdentity do GraphQL
function mapToBureauIdentity(identity: any): BureauIdentity {
  let detalhes = {};
  try {
    if (identity.details) {
      detalhes = JSON.parse(identity.details);
    }
  } catch (e) {
    // Ignorar erro de parse
  }

  return {
    id: identity.id,
    usuarioId: identity.usuarioId || identity.usuarioID,
    tenantId: identity.tenantId || identity.tenantID,
    externalId: identity.externalId || identity.externalID,
    externalTenantId: identity.externalTenantId || identity.externalTenantID,
    tipoVinculo: identity.profileType,
    nivelAcesso: identity.accessLevel,
    status: identity.status,
    dataCriacao: identity.createdAt || new Date(),
    dataAtualizacao: identity.updatedAt,
    motivoStatus: identity.statusReason,
    detalhes
  };
}

// Mapear objeto de autorização do DataCore para o tipo BureauAutorizacao do GraphQL
function mapToBureauAutorizacao(autorizacao: any): BureauAutorizacao {
  return {
    id: autorizacao.id,
    identityId: autorizacao.identityId || autorizacao.identityID,
    tipoConsulta: autorizacao.tipoConsulta,
    finalidade: autorizacao.finalidade,
    justificativa: autorizacao.justificativa,
    dataAutorizacao: autorizacao.dataAutorizacao,
    dataValidade: autorizacao.dataValidade,
    status: autorizacao.status,
    autorizadoPor: autorizacao.autorizadoPor
  };
}

// Mapear objeto de token do DataCore para o tipo BureauAccessToken do GraphQL
function mapToBureauAccessToken(token: any): BureauAccessToken {
  return {
    token: token.token,
    type: token.type,
    expiresAt: token.expiresAt,
    refreshToken: token.refreshToken,
    identityId: token.identityId || token.identityID,
    escopos: token.scope,
    finalidade: token.finalidade
  };
}