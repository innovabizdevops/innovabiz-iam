import { TrustGuardService } from '../../../../app/trust_guard_service';
import { ComplianceService } from '../../../../infrastructure/compliance/compliance_service';
import { ContextRepository } from '../../../repositories/context-repository';
import { ContextAttributeRepository } from '../../../repositories/context-attribute-repository';
import { ContextIntegrationRepository } from '../../../repositories/context-integration-repository';
import { ContextHistoryRepository } from '../../../repositories/context-history-repository';
import { 
  IdentityContext, 
  ContextFilterInput, 
  PaginationInput, 
  CreateIdentityContextInput,
  UpdateIdentityContextInput,
  OperationResult
} from '../types/generated';
import { Logger } from '../../../../infrastructure/common/logging';
import { AuthorizationService } from '../../../../services/authorization-service';
import { IdentityContextValidator } from '../validators/identity-context-validator';
import { AuditService } from '../../../../services/audit-service';

/**
 * Resolvers para as operações de contexto de identidade
 */
export const identityContextResolvers = {
  Query: {
    /**
     * Obtém um contexto de identidade pelo ID
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (contextID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    identityContext: async (_, { contextID }, context) => {
      Logger.debug('Query: identityContext', { contextID });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:read', 
        contextID
      );
      
      try {
        const repository = new ContextRepository();
        const identityContext = await repository.getById(contextID);
        
        if (!identityContext) {
          Logger.warn('Context not found', { contextID });
          return null;
        }
        
        // Registra auditoria da consulta
        await AuditService.log({
          action: 'READ',
          resource: 'IDENTITY_CONTEXT',
          resourceId: contextID,
          userId: context.user?.id,
          details: { contextType: identityContext.contextType }
        });
        
        return identityContext;
      } catch (error) {
        Logger.error('Failed to get identity context', { contextID, error });
        throw new Error(`Erro ao obter contexto de identidade: ${error.message}`);
      }
    },
    
    /**
     * Lista contextos de identidade com filtros e paginação
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (filter, pagination)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    identityContexts: async (_, { filter, pagination }, context) => {
      Logger.debug('Query: identityContexts', { filter, pagination });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:contexts:list'
      );
      
      try {
        const repository = new ContextRepository();
        const defaultPagination: PaginationInput = {
          page: 1,
          pageSize: 20
        };
        
        const appliedPagination = pagination || defaultPagination;
        const result = await repository.findAll(filter, appliedPagination);
        
        // Registra auditoria da consulta
        await AuditService.log({
          action: 'LIST',
          resource: 'IDENTITY_CONTEXTS',
          userId: context.user?.id,
          details: { filter, pagination: appliedPagination }
        });
        
        return {
          items: result.items,
          totalCount: result.totalCount,
          hasMore: result.hasMore
        };
      } catch (error) {
        Logger.error('Failed to list identity contexts', { filter, pagination, error });
        throw new Error(`Erro ao listar contextos de identidade: ${error.message}`);
      }
    },
    
    /**
     * Obtém o histórico de alterações de um contexto
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (contextID, pagination)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    contextHistory: async (_, { contextID, pagination }, context) => {
      Logger.debug('Query: contextHistory', { contextID, pagination });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:history:read', 
        contextID
      );
      
      try {
        const historyRepository = new ContextHistoryRepository();
        const defaultPagination: PaginationInput = {
          page: 1,
          pageSize: 20
        };
        
        const appliedPagination = pagination || defaultPagination;
        const result = await historyRepository.getContextHistory(contextID, appliedPagination);
        
        return {
          items: result.items,
          totalCount: result.totalCount,
          hasMore: result.hasMore
        };
      } catch (error) {
        Logger.error('Failed to get context history', { contextID, pagination, error });
        throw new Error(`Erro ao obter histórico de contexto: ${error.message}`);
      }
    },
    
    /**
     * Obtém informações de conformidade para um contexto
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (contextID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    contextCompliance: async (_, { contextID }, context) => {
      Logger.debug('Query: contextCompliance', { contextID });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:compliance:read', 
        contextID
      );
      
      try {
        const repository = new ContextRepository();
        const identityContext = await repository.getById(contextID);
        
        if (!identityContext) {
          Logger.warn('Context not found', { contextID });
          throw new Error(`Contexto de identidade não encontrado: ${contextID}`);
        }
        
        const complianceService = new ComplianceService();
        const compliance = await complianceService.assessContextCompliance(contextID);
        
        return compliance;
      } catch (error) {
        Logger.error('Failed to get context compliance', { contextID, error });
        throw new Error(`Erro ao obter informações de conformidade: ${error.message}`);
      }
    }
  },
  
  Mutation: {
    /**
     * Cria um novo contexto de identidade
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (input)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    createIdentityContext: async (_, { input }, context) => {
      Logger.debug('Mutation: createIdentityContext', { input });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:create'
      );
      
      try {
        // Validação dos dados de entrada
        const validator = new IdentityContextValidator();
        await validator.validateCreate(input);
        
        const repository = new ContextRepository();
        
        // Verifica se já existe um contexto com a mesma identidade e tipo
        const existingContext = await repository.findByIdentityAndType(
          input.identityId,
          input.contextType
        );
        
        if (existingContext) {
          throw new Error(`Já existe um contexto do tipo ${input.contextType} para esta identidade`);
        }
        
        // Prepara dados para inserção
        const contextData = {
          ...input,
          createdBy: context.user?.id,
          updatedBy: context.user?.id,
          createdAt: new Date(),
          updatedAt: new Date()
        };
        
        // Cria o contexto
        const newContext = await repository.create(contextData);
        
        // Registra auditoria da criação
        await AuditService.log({
          action: 'CREATE',
          resource: 'IDENTITY_CONTEXT',
          resourceId: newContext.contextID,
          userId: context.user?.id,
          details: {
            identityId: input.identityId,
            contextType: input.contextType
          }
        });
        
        return newContext;
      } catch (error) {
        Logger.error('Failed to create identity context', { input, error });
        throw new Error(`Erro ao criar contexto de identidade: ${error.message}`);
      }
    },
    
    /**
     * Atualiza um contexto de identidade existente
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (input)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    updateIdentityContext: async (_, { input }, context) => {
      Logger.debug('Mutation: updateIdentityContext', { input });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:update', 
        input.contextId
      );
      
      try {
        // Validação dos dados de entrada
        const validator = new IdentityContextValidator();
        await validator.validateUpdate(input);
        
        const repository = new ContextRepository();
        const existingContext = await repository.getById(input.contextId);
        
        if (!existingContext) {
          throw new Error(`Contexto de identidade não encontrado: ${input.contextId}`);
        }
        
        // Prepara dados para atualização
        const contextData = {
          ...input,
          updatedBy: context.user?.id,
          updatedAt: new Date()
        };
        
        // Registra os campos alterados para o histórico
        const changedFields = Object.keys(input).filter(key => 
          key !== 'contextId' && input[key] !== undefined && 
          input[key] !== existingContext[key]
        );
        
        // Atualiza o contexto
        const updatedContext = await repository.update(input.contextId, contextData);
        
        // Registra histórico de alterações
        if (changedFields.length > 0) {
          const historyRepository = new ContextHistoryRepository();
          await historyRepository.createHistoryEntry({
            contextID: input.contextId,
            changeType: 'UPDATE',
            changedFields,
            previousValues: changedFields.reduce((acc, field) => {
              acc[field] = existingContext[field];
              return acc;
            }, {}),
            newValues: changedFields.reduce((acc, field) => {
              acc[field] = input[field];
              return acc;
            }, {}),
            changedBy: context.user?.id,
            timestamp: new Date()
          });
        }
        
        // Registra auditoria da atualização
        await AuditService.log({
          action: 'UPDATE',
          resource: 'IDENTITY_CONTEXT',
          resourceId: input.contextId,
          userId: context.user?.id,
          details: {
            changedFields,
            contextType: existingContext.contextType
          }
        });
        
        return updatedContext;
      } catch (error) {
        Logger.error('Failed to update identity context', { input, error });
        throw new Error(`Erro ao atualizar contexto de identidade: ${error.message}`);
      }
    },
    
    /**
     * Remove um contexto de identidade
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (contextID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    removeIdentityContext: async (_, { contextID }, context) => {
      Logger.debug('Mutation: removeIdentityContext', { contextID });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:context:delete', 
        contextID
      );
      
      try {
        const repository = new ContextRepository();
        const existingContext = await repository.getById(contextID);
        
        if (!existingContext) {
          throw new Error(`Contexto de identidade não encontrado: ${contextID}`);
        }
        
        // Verifica se o contexto tem integrações ativas
        const integrationRepository = new ContextIntegrationRepository();
        const integrations = await integrationRepository.findByContextId(contextID);
        
        if (integrations.length > 0) {
          throw new Error(`Não é possível remover o contexto pois existem ${integrations.length} integrações associadas a ele`);
        }
        
        // Verifica se o contexto tem atributos
        const attributeRepository = new ContextAttributeRepository();
        const attributes = await attributeRepository.findByContextId(contextID);
        
        // Remove atributos se existirem
        if (attributes.length > 0) {
          for (const attribute of attributes) {
            await attributeRepository.delete(attribute.attributeID);
          }
        }
        
        // Remove o contexto
        await repository.delete(contextID);
        
        // Registra histórico de alteração
        const historyRepository = new ContextHistoryRepository();
        await historyRepository.createHistoryEntry({
          contextID,
          changeType: 'DELETE',
          changedFields: ['*'],
          previousValues: {
            contextType: existingContext.contextType,
            contextStatus: existingContext.contextStatus
          },
          newValues: {},
          changedBy: context.user?.id,
          timestamp: new Date(),
          reason: 'Contexto removido manualmente'
        });
        
        // Registra auditoria da remoção
        await AuditService.log({
          action: 'DELETE',
          resource: 'IDENTITY_CONTEXT',
          resourceId: contextID,
          userId: context.user?.id,
          details: {
            identityId: existingContext.identityID,
            contextType: existingContext.contextType,
            attributesRemoved: attributes.length
          }
        });
        
        return {
          success: true,
          message: `Contexto de identidade removido com sucesso`,
          code: 'CONTEXT_DELETED'
        };
      } catch (error) {
        Logger.error('Failed to remove identity context', { contextID, error });
        return {
          success: false,
          message: `Erro ao remover contexto de identidade: ${error.message}`,
          code: 'CONTEXT_DELETE_FAILED'
        };
      }
    },
    
    /**
     * Verifica uma identidade usando o TrustGuard
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (identityID, contextType, verificationLevel)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    verifyIdentityWithTrustGuard: async (_, { identityID, contextType, verificationLevel }, context) => {
      Logger.debug('Mutation: verifyIdentityWithTrustGuard', { identityID, contextType, verificationLevel });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:verify'
      );
      
      try {
        // Busca o contexto associado
        const repository = new ContextRepository();
        const existingContext = await repository.findByIdentityAndType(
          identityID,
          contextType
        );
        
        if (!existingContext) {
          throw new Error(`Contexto de identidade do tipo ${contextType} não encontrado para identidade ${identityID}`);
        }
        
        // Inicia o processo de verificação
        const trustGuardService = new TrustGuardService();
        const verificationResult = await trustGuardService.verifyIdentity({
          identityID,
          contextID: existingContext.contextID,
          verificationLevel,
          requestedBy: context.user?.id
        });
        
        // Atualiza o contexto com os resultados da verificação
        if (verificationResult.verificationStatus === 'VERIFIED') {
          await repository.update(existingContext.contextID, {
            verificationLevel: verificationLevel,
            trustScore: verificationResult.trustScore,
            lastVerifiedAt: new Date(),
            updatedBy: context.user?.id,
            updatedAt: new Date()
          });
          
          // Registra histórico de alteração
          const historyRepository = new ContextHistoryRepository();
          await historyRepository.createHistoryEntry({
            contextID: existingContext.contextID,
            changeType: 'VERIFICATION',
            changedFields: ['verificationLevel', 'trustScore', 'lastVerifiedAt'],
            previousValues: {
              verificationLevel: existingContext.verificationLevel,
              trustScore: existingContext.trustScore
            },
            newValues: {
              verificationLevel: verificationLevel,
              trustScore: verificationResult.trustScore
            },
            changedBy: context.user?.id,
            timestamp: new Date(),
            reason: `Verificação de identidade com TrustGuard (${verificationLevel})`
          });
        }
        
        // Registra auditoria da verificação
        await AuditService.log({
          action: 'VERIFY',
          resource: 'IDENTITY',
          resourceId: identityID,
          userId: context.user?.id,
          details: {
            contextID: existingContext.contextID,
            contextType,
            verificationLevel,
            result: verificationResult.verificationStatus,
            trustScore: verificationResult.trustScore
          }
        });
        
        return verificationResult;
      } catch (error) {
        Logger.error('Failed to verify identity with TrustGuard', { 
          identityID, contextType, verificationLevel, error 
        });
        throw new Error(`Erro ao verificar identidade com TrustGuard: ${error.message}`);
      }
    }
  },
  
  IdentityContext: {
    /**
     * Resolver para obter os atributos associados ao contexto
     * @param parent - Contexto de identidade
     * @param args - Argumentos (filter)
     */
    attributes: async (parent, { filter }) => {
      Logger.debug('Field resolver: IdentityContext.attributes', { 
        contextID: parent.contextID, filter 
      });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        const combinedFilter = {
          ...filter,
          contextIds: [parent.contextID]
        };
        
        const attributes = await attributeRepository.findAll(combinedFilter);
        return attributes.items;
      } catch (error) {
        Logger.error('Failed to resolve context attributes', { 
          contextID: parent.contextID, filter, error 
        });
        throw new Error(`Erro ao resolver atributos do contexto: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter as integrações nas quais este contexto é a origem
     * @param parent - Contexto de identidade
     */
    sourceIntegrations: async (parent) => {
      Logger.debug('Field resolver: IdentityContext.sourceIntegrations', { 
        contextID: parent.contextID 
      });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        const integrations = await integrationRepository.findBySourceContextId(parent.contextID);
        return integrations;
      } catch (error) {
        Logger.error('Failed to resolve source integrations', { 
          contextID: parent.contextID, error 
        });
        throw new Error(`Erro ao resolver integrações de origem: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter as integrações nas quais este contexto é o destino
     * @param parent - Contexto de identidade
     */
    targetIntegrations: async (parent) => {
      Logger.debug('Field resolver: IdentityContext.targetIntegrations', { 
        contextID: parent.contextID 
      });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        const integrations = await integrationRepository.findByTargetContextId(parent.contextID);
        return integrations;
      } catch (error) {
        Logger.error('Failed to resolve target integrations', { 
          contextID: parent.contextID, error 
        });
        throw new Error(`Erro ao resolver integrações de destino: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter informações de conformidade do contexto
     * @param parent - Contexto de identidade
     */
    compliance: async (parent) => {
      Logger.debug('Field resolver: IdentityContext.compliance', { 
        contextID: parent.contextID 
      });
      
      try {
        const complianceService = new ComplianceService();
        const compliance = await complianceService.assessContextCompliance(parent.contextID);
        return compliance;
      } catch (error) {
        Logger.error('Failed to resolve context compliance', { 
          contextID: parent.contextID, error 
        });
        throw new Error(`Erro ao resolver conformidade do contexto: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter o histórico de alterações do contexto
     * @param parent - Contexto de identidade
     * @param args - Argumentos (pagination)
     */
    history: async (parent, { pagination }) => {
      Logger.debug('Field resolver: IdentityContext.history', { 
        contextID: parent.contextID, pagination 
      });
      
      try {
        const historyRepository = new ContextHistoryRepository();
        const defaultPagination = {
          page: 1,
          pageSize: 20
        };
        
        const appliedPagination = pagination || defaultPagination;
        const result = await historyRepository.getContextHistory(parent.contextID, appliedPagination);
        
        return {
          items: result.items,
          totalCount: result.totalCount,
          hasMore: result.hasMore
        };
      } catch (error) {
        Logger.error('Failed to resolve context history', { 
          contextID: parent.contextID, pagination, error 
        });
        throw new Error(`Erro ao resolver histórico do contexto: ${error.message}`);
      }
    }
  }
};