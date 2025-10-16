/**
 * Resolvers GraphQL para integração com Mobile Money
 * Implementa padrões avançados de segurança, observabilidade e validação
 * Compatível com requisitos regulatórios dos mercados PALOP, SADC e CPLP
 */

import { GraphQLResolveInfo, GraphQLScalarType } from 'graphql';
import { IResolvers } from '@graphql-tools/utils';
import { AuthenticationError, ForbiddenError, UserInputError, ApolloError } from 'apollo-server-express';
import { v4 as uuidv4 } from 'uuid';
import * as mobileMoney from '../../../../integration/mobile-money/service';
import { validateTransactionRequest } from './validators';
import { enrichTransactionWithRiskData } from './risk-service';
import { TenantContext, AuthContext } from '../../types';
import { checkPermissions } from '../../directives/auth';
import { logger } from '../../../../observability/logging';
import { metrics } from '../../../../observability/metrics';
import { tracer } from '../../../../observability/tracing';
import { ComplianceError } from '../../../../errors/compliance-error';
import { TransactionLimitsService } from './transaction-limits-service';
import { resolveRegionalCompliance } from './regional-compliance';
import { DateTimeResolver, JSONResolver } from '../../scalars';

/**
 * Mapeamento de tipos entre GraphQL e serviço de Mobile Money
 */
const mapProviderType = (provider: string): string => {
  const mapping: { [key: string]: string } = {
    'MPESA': 'mpesa',
    'UNITEL': 'unitel',
    'ECOCASH': 'ecocash',
    'AIRTEL': 'airtel',
    'ORANGE': 'orange',
    'MTN': 'mtn',
  };
  return mapping[provider] || provider.toLowerCase();
};

const mapTransactionType = (type: string): string => {
  const mapping: { [key: string]: string } = {
    'DEPOSIT': 'DEPOSIT',
    'WITHDRAWAL': 'WITHDRAWAL',
    'PAYMENT': 'PAYMENT',
    'TRANSFER': 'TRANSFER',
  };
  return mapping[type] || type;
};

/**
 * Transformação de modelos entre GraphQL e domínio interno
 */
const transformTransactionToGraphQL = (transaction: any): any => {
  // Converter propriedades para o formato esperado pelo schema GraphQL
  return {
    id: transaction.transactionId || transaction.id,
    provider: transaction.provider?.toUpperCase() || 'UNKNOWN',
    type: transaction.type,
    amount: transaction.amount,
    currency: transaction.currency,
    phoneNumber: transaction.phoneNumber,
    referenceId: transaction.referenceId || transaction.referenceID,
    userId: transaction.userId || transaction.userID,
    description: transaction.description,
    status: transaction.status,
    statusCode: transaction.statusCode,
    message: transaction.message,
    providerRef: transaction.providerRef,
    otpRequired: Boolean(transaction.otpRequired),
    otpReference: transaction.otpReference,
    redirectUrl: transaction.redirectUrl,
    fee: transaction.fee || 0,
    processedAt: transaction.processedAt || new Date(),
    completedAt: transaction.completedAt,
    riskScore: transaction.riskScore,
    riskAssessment: transaction.riskAssessment || {},
    tenantId: transaction.tenantId || transaction.tenantID,
    metadata: transaction.metadata || {},
    complianceData: transaction.complianceData ? {
      consentId: transaction.complianceData.consentId || transaction.complianceData.consentID,
      authorizationId: transaction.complianceData.authorizationId || transaction.complianceData.authorizationID,
      purposeCode: transaction.complianceData.purposeCode,
      consentDate: transaction.complianceData.consentDate,
      kycLevel: transaction.complianceData.kycLevel,
      regulatoryId: transaction.complianceData.regulatoryId || transaction.complianceData.regulatoryID,
      requiredDocuments: transaction.complianceData.requiredDocuments || [],
      verificationStatus: transaction.complianceData.verificationStatus || 'PENDING',
      verifiedAt: transaction.complianceData.verifiedAt,
      complianceOfficer: transaction.complianceData.complianceOfficer || '',
      riskCategory: transaction.complianceData.riskCategory || 'STANDARD',
    } : null,
    createdAt: transaction.createdAt || transaction.processedAt || new Date(),
    updatedAt: transaction.updatedAt || new Date(),
  };
};

/**
 * Implementação dos resolvers GraphQL
 */
export const resolvers: IResolvers = {
  // Definição dos scalars personalizados
  DateTime: DateTimeResolver,
  JSON: JSONResolver,

  // Resolvers de Query
  Query: {
    /**
     * Consulta detalhes de uma transação Mobile Money específica
     */
    mobileMoneyTransaction: async (_, { id, tenantId }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.mobileMoneyTransaction');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para acessar dados de transação');
        }

        checkPermissions(context.user, ['mobile_money:read', `mobile_money:transaction:${id}:read`]);

        // Usar tenantId do contexto se não especificado
        const effectiveTenantId = tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        logger.info('Consultando transação Mobile Money', { 
          transactionId: id,
          tenantId: effectiveTenantId,
          userId: context.user.id
        });

        metrics.increment('graphql.query.mobile_money_transaction', {
          tenantId: effectiveTenantId
        });

        // Chamar serviço para obter detalhes da transação
        const transaction = await mobileMoney.getTransactionById(id, effectiveTenantId);
        
        if (!transaction) {
          throw new ApolloError('Transação não encontrada', 'TRANSACTION_NOT_FOUND');
        }

        // Verificar se o usuário tem permissão para acessar esta transação específica
        if (transaction.userId !== context.user.id && 
            !context.user.permissions.includes('mobile_money:admin') &&
            !context.user.permissions.includes('mobile_money:read:all')) {
          throw new ForbiddenError('Você não tem permissão para acessar esta transação');
        }

        return transformTransactionToGraphQL(transaction);
      } catch (error) {
        logger.error('Erro ao consultar transação Mobile Money', { 
          error: error.message,
          transactionId: id,
          stack: error.stack
        });

        metrics.increment('graphql.error.mobile_money_transaction', {
          error: error.name
        });

        throw error;
      } finally {
        span.end();
      }
    },

    /**
     * Consulta histórico de transações Mobile Money com filtros e paginação
     */
    mobileMoneyTransactionHistory: async (_, { filters, pagination }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.mobileMoneyTransactionHistory');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para acessar histórico de transações');
        }

        checkPermissions(context.user, ['mobile_money:read']);

        // Usar tenantId do contexto se não especificado no filtro
        const effectiveTenantId = filters.tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        // Verificar permissões específicas para o caso de consulta de todas as transações
        if (!filters.userId) {
          checkPermissions(context.user, ['mobile_money:read:all']);
        } else if (filters.userId !== context.user.id) {
          // Se está tentando acessar transações de outro usuário
          checkPermissions(context.user, ['mobile_money:read:all']);
        }

        // Configurar paginação padrão se não especificada
        const paginationConfig = {
          first: pagination?.first || 10,
          after: pagination?.after || null,
          last: pagination?.last || null,
          before: pagination?.before || null
        };

        logger.info('Consultando histórico de transações Mobile Money', { 
          filters,
          pagination: paginationConfig,
          tenantId: effectiveTenantId,
          userId: context.user.id
        });

        metrics.increment('graphql.query.mobile_money_transaction_history', {
          tenantId: effectiveTenantId
        });

        // Chamar serviço para obter histórico de transações
        const result = await mobileMoney.getTransactionHistory({
          ...filters,
          tenantId: effectiveTenantId,
          pagination: paginationConfig
        });
        
        // Transformar resultado para o formato GraphQL
        return {
          transactions: result.transactions.map(transformTransactionToGraphQL),
          totalCount: result.totalCount,
          pageInfo: {
            hasNextPage: result.pageInfo.hasNextPage,
            hasPreviousPage: result.pageInfo.hasPreviousPage,
            startCursor: result.pageInfo.startCursor,
            endCursor: result.pageInfo.endCursor
          }
        };
      } catch (error) {
        logger.error('Erro ao consultar histórico de transações Mobile Money', { 
          error: error.message,
          filters,
          stack: error.stack
        });

        metrics.increment('graphql.error.mobile_money_transaction_history', {
          error: error.name
        });

        throw error;
      } finally {
        span.end();
      }
    },

    /**
     * Verifica elegibilidade de um usuário para serviços Mobile Money
     */
    checkMobileMoneyEligibility: async (_, { input }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.checkMobileMoneyEligibility');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para verificar elegibilidade');
        }

        checkPermissions(context.user, ['mobile_money:eligibility:read']);

        // Verificar se está consultando elegibilidade própria ou de outro usuário
        if (input.userId !== context.user.id) {
          checkPermissions(context.user, ['mobile_money:eligibility:read:all']);
        }

        // Usar tenantId do contexto se não especificado
        const effectiveTenantId = input.tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        logger.info('Verificando elegibilidade para Mobile Money', { 
          userId: input.userId,
          provider: input.provider,
          regionCode: input.regionCode,
          tenantId: effectiveTenantId
        });

        metrics.increment('graphql.query.mobile_money_eligibility', {
          tenantId: effectiveTenantId,
          provider: input.provider,
          regionCode: input.regionCode
        });

        // Chamar serviço para verificar elegibilidade
        const eligibility = await mobileMoney.checkEligibility({
          userId: input.userId,
          provider: mapProviderType(input.provider),
          regionCode: input.regionCode,
          tenantId: effectiveTenantId,
          transactionType: input.transactionType ? mapTransactionType(input.transactionType) : undefined
        });
        
        return {
          eligible: eligibility.eligible,
          services: eligibility.services || [],
          limits: eligibility.limits || {},
          restrictions: eligibility.restrictions || [],
          kycRequired: eligibility.kycRequired || false,
          requiresUpgrade: eligibility.requiresUpgrade || false,
          message: eligibility.message || '',
          regionRequirements: eligibility.regionRequirements || {}
        };
      } catch (error) {
        logger.error('Erro ao verificar elegibilidade para Mobile Money', { 
          error: error.message,
          input,
          stack: error.stack
        });

        metrics.increment('graphql.error.mobile_money_eligibility', {
          error: error.name
        });

        throw error;
      } finally {
        span.end();
      }
    },

    /**
     * Consulta limites de transação por provedor e região
     */
    mobileMoneyTransactionLimits: async (_, { provider, regionCode, kycLevel, tenantId }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.mobileMoneyTransactionLimits');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para consultar limites');
        }

        checkPermissions(context.user, ['mobile_money:limits:read']);

        // Usar tenantId do contexto se não especificado
        const effectiveTenantId = tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        logger.info('Consultando limites de transação Mobile Money', { 
          provider,
          regionCode,
          kycLevel,
          tenantId: effectiveTenantId
        });

        metrics.increment('graphql.query.mobile_money_limits', {
          tenantId: effectiveTenantId,
          provider,
          regionCode
        });

        // Instanciar serviço de limites de transação
        const limitsService = new TransactionLimitsService(effectiveTenantId);
        
        // Consultar limites de transação
        const limits = await limitsService.getLimits({
          provider: mapProviderType(provider),
          regionCode,
          kycLevel
        });
        
        return limits;
      } catch (error) {
        logger.error('Erro ao consultar limites de transação Mobile Money', { 
          error: error.message,
          provider,
          regionCode,
          stack: error.stack
        });

        metrics.increment('graphql.error.mobile_money_limits', {
          error: error.name
        });

        throw error;
      } finally {
        span.end();
      }
    }
  },

  // Resolvers de Mutation
  Mutation: {
    /**
     * Inicia uma transação Mobile Money
     */
    initiateMobileMoneyTransaction: async (_, { input }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.initiateMobileMoneyTransaction');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para iniciar transação');
        }

        checkPermissions(context.user, ['mobile_money:write']);

        // Verificar se a transação é para o próprio usuário ou para outro
        if (input.userId !== context.user.id) {
          checkPermissions(context.user, ['mobile_money:write:all']);
        }

        // Usar tenantId do contexto se não especificado
        const effectiveTenantId = input.tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        // Validar dados da requisição
        await validateTransactionRequest(input, effectiveTenantId);

        logger.info('Iniciando transação Mobile Money', { 
          provider: input.provider,
          type: input.type,
          amount: input.amount,
          currency: input.currency,
          userId: input.userId,
          tenantId: effectiveTenantId
        });

        metrics.increment('graphql.mutation.initiate_mobile_money_transaction', {
          tenantId: effectiveTenantId,
          provider: input.provider,
          type: input.type,
          currency: input.currency
        });

        // Verificar conformidade regulatória para a região
        const complianceResult = await resolveRegionalCompliance(input, context);
        
        if (!complianceResult.compliant) {
          throw new ComplianceError(
            complianceResult.message || 'Requisitos regulatórios não atendidos',
            complianceResult.details
          );
        }

        // Preparar requisição para o serviço de Mobile Money
        const transactionRequest = {
          provider: mapProviderType(input.provider),
          type: mapTransactionType(input.type),
          amount: input.amount,
          currency: input.currency,
          phoneNumber: input.phoneNumber,
          referenceId: input.referenceId || uuidv4(),
          userId: input.userId,
          description: input.description || '',
          callbackUrl: input.callbackUrl,
          tenantId: effectiveTenantId,
          metadata: input.metadata || {},
          requireOTP: input.requireOtp || false,
          regionCode: input.regionCode,
          complianceData: input.complianceData ? {
            consentId: input.complianceData.consentId,
            authorizationId: input.complianceData.authorizationId,
            purposeCode: input.complianceData.purposeCode || 'DEFAULT',
            consentDate: input.complianceData.consentDate || new Date(),
            kycLevel: input.complianceData.kycLevel,
            regulatoryId: input.complianceData.regulatoryId
          } : undefined
        };

        // Enriquecer com dados de avaliação de risco
        const enrichedRequest = await enrichTransactionWithRiskData(transactionRequest, context);

        // Chamar serviço para iniciar transação
        const transaction = await mobileMoney.initiateTransaction(enrichedRequest);
        
        // Transformar resultado para formato GraphQL
        const transformedTransaction = transformTransactionToGraphQL(transaction);
        
        return {
          transaction: transformedTransaction,
          success: true,
          message: 'Transação iniciada com sucesso',
          errors: []
        };
      } catch (error) {
        logger.error('Erro ao iniciar transação Mobile Money', { 
          error: error.message,
          input,
          stack: error.stack
        });

        metrics.increment('graphql.error.initiate_mobile_money_transaction', {
          error: error.name
        });

        // Retornar resposta com detalhes do erro
        return {
          transaction: null,
          success: false,
          message: error.message,
          errors: [error.message]
        };
      } finally {
        span.end();
      }
    },

    /**
     * Verifica OTP para uma transação Mobile Money
     */
    verifyMobileMoneyOTP: async (_, { input }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.verifyMobileMoneyOTP');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para verificar OTP');
        }

        checkPermissions(context.user, ['mobile_money:write']);

        // Obter detalhes da transação para verificar propriedade
        const effectiveTenantId = input.tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        const transaction = await mobileMoney.getTransactionById(input.transactionId, effectiveTenantId);
        
        if (!transaction) {
          throw new ApolloError('Transação não encontrada', 'TRANSACTION_NOT_FOUND');
        }

        // Verificar se o usuário tem permissão para verificar OTP para esta transação
        if (transaction.userId !== context.user.id && 
            !context.user.permissions.includes('mobile_money:admin') &&
            !context.user.permissions.includes('mobile_money:write:all')) {
          throw new ForbiddenError('Você não tem permissão para verificar OTP para esta transação');
        }

        logger.info('Verificando OTP para transação Mobile Money', { 
          transactionId: input.transactionId,
          tenantId: effectiveTenantId,
          userId: context.user.id
        });

        metrics.increment('graphql.mutation.verify_mobile_money_otp', {
          tenantId: effectiveTenantId
        });

        // Chamar serviço para verificar OTP
        const verificationResult = await mobileMoney.verifyOTP(
          input.transactionId,
          input.otp,
          effectiveTenantId
        );
        
        // Transformar resultado para formato GraphQL
        const transformedTransaction = transformTransactionToGraphQL(verificationResult);
        
        return {
          transaction: transformedTransaction,
          success: true,
          message: 'OTP verificado com sucesso',
          errors: []
        };
      } catch (error) {
        logger.error('Erro ao verificar OTP para transação Mobile Money', { 
          error: error.message,
          transactionId: input.transactionId,
          stack: error.stack
        });

        metrics.increment('graphql.error.verify_mobile_money_otp', {
          error: error.name
        });

        // Retornar resposta com detalhes do erro
        return {
          transaction: null,
          success: false,
          message: error.message,
          errors: [error.message]
        };
      } finally {
        span.end();
      }
    },

    /**
     * Cancela uma transação Mobile Money pendente
     */
    cancelMobileMoneyTransaction: async (_, { transactionId, reason, tenantId }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.cancelMobileMoneyTransaction');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para cancelar transação');
        }

        checkPermissions(context.user, ['mobile_money:write']);

        // Obter detalhes da transação para verificar propriedade
        const effectiveTenantId = tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        const transaction = await mobileMoney.getTransactionById(transactionId, effectiveTenantId);
        
        if (!transaction) {
          throw new ApolloError('Transação não encontrada', 'TRANSACTION_NOT_FOUND');
        }

        // Verificar se o usuário tem permissão para cancelar esta transação
        if (transaction.userId !== context.user.id && 
            !context.user.permissions.includes('mobile_money:admin') &&
            !context.user.permissions.includes('mobile_money:write:all')) {
          throw new ForbiddenError('Você não tem permissão para cancelar esta transação');
        }

        logger.info('Cancelando transação Mobile Money', { 
          transactionId,
          reason,
          tenantId: effectiveTenantId,
          userId: context.user.id
        });

        metrics.increment('graphql.mutation.cancel_mobile_money_transaction', {
          tenantId: effectiveTenantId
        });

        // Chamar serviço para cancelar transação
        const cancelResult = await mobileMoney.cancelTransaction(
          transactionId,
          reason,
          effectiveTenantId
        );
        
        // Transformar resultado para formato GraphQL
        const transformedTransaction = transformTransactionToGraphQL(cancelResult);
        
        return {
          transaction: transformedTransaction,
          success: true,
          message: 'Transação cancelada com sucesso',
          errors: []
        };
      } catch (error) {
        logger.error('Erro ao cancelar transação Mobile Money', { 
          error: error.message,
          transactionId,
          stack: error.stack
        });

        metrics.increment('graphql.error.cancel_mobile_money_transaction', {
          error: error.name
        });

        // Retornar resposta com detalhes do erro
        return {
          transaction: null,
          success: false,
          message: error.message,
          errors: [error.message]
        };
      } finally {
        span.end();
      }
    },

    /**
     * Registra consentimento para operações Mobile Money (compliance)
     */
    registerMobileMoneyConsent: async (_, { userId, purpose, scope, expiresAt, consentText, tenantId }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
      const span = tracer.startSpan('resolvers.registerMobileMoneyConsent');
      
      try {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para registrar consentimento');
        }

        checkPermissions(context.user, ['mobile_money:consent:write']);

        // Verificar se está registrando consentimento próprio ou de outro usuário
        if (userId !== context.user.id) {
          checkPermissions(context.user, ['mobile_money:consent:write:all']);
        }

        // Usar tenantId do contexto se não especificado
        const effectiveTenantId = tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        logger.info('Registrando consentimento para Mobile Money', { 
          userId,
          purpose,
          scope,
          tenantId: effectiveTenantId
        });

        metrics.increment('graphql.mutation.register_mobile_money_consent', {
          tenantId: effectiveTenantId
        });

        // Chamar serviço para registrar consentimento
        const consentId = await mobileMoney.registerConsent({
          userId,
          purpose,
          scope,
          expiresAt: expiresAt || new Date(Date.now() + 90 * 24 * 60 * 60 * 1000), // 90 dias por padrão
          consentText,
          tenantId: effectiveTenantId,
          consentVersion: '1.0',
          registeredBy: context.user.id
        });
        
        return consentId;
      } catch (error) {
        logger.error('Erro ao registrar consentimento para Mobile Money', { 
          error: error.message,
          userId,
          purpose,
          stack: error.stack
        });

        metrics.increment('graphql.error.register_mobile_money_consent', {
          error: error.name
        });

        throw error;
      } finally {
        span.end();
      }
    }
  },

  // Resolvers de Subscription
  Subscription: {
    /**
     * Subscrição para atualizações de status de transações Mobile Money
     */
    mobileMoneyTransactionStatusChanged: {
      subscribe: async (_, { transactionId, tenantId }, context: TenantContext & AuthContext, info: GraphQLResolveInfo) => {
        // Verificar autenticação e autorização
        if (!context.user) {
          throw new AuthenticationError('Autenticação necessária para subscrever atualizações');
        }

        checkPermissions(context.user, ['mobile_money:read']);

        // Obter detalhes da transação para verificar propriedade
        const effectiveTenantId = tenantId || context.tenantId;
        
        if (!effectiveTenantId) {
          throw new UserInputError('TenantId é necessário');
        }

        const transaction = await mobileMoney.getTransactionById(transactionId, effectiveTenantId);
        
        if (!transaction) {
          throw new ApolloError('Transação não encontrada', 'TRANSACTION_NOT_FOUND');
        }

        // Verificar se o usuário tem permissão para receber atualizações desta transação
        if (transaction.userId !== context.user.id && 
            !context.user.permissions.includes('mobile_money:admin') &&
            !context.user.permissions.includes('mobile_money:read:all')) {
          throw new ForbiddenError('Você não tem permissão para receber atualizações desta transação');
        }

        logger.info('Subscrevendo a atualizações de status de transação Mobile Money', { 
          transactionId,
          tenantId: effectiveTenantId,
          userId: context.user.id
        });

        metrics.increment('graphql.subscription.mobile_money_transaction_status', {
          tenantId: effectiveTenantId
        });

        // Configurar e retornar AsyncIterator para esta subscrição
        const pubsub = context.pubsub;
        const channel = `MOBILE_MONEY_TRANSACTION_${transactionId}_${effectiveTenantId}`;
        
        return pubsub.asyncIterator([channel]);
      }
    }
  }
};

export default resolvers;