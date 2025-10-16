import { ContextIntegrationRepository } from '../../../repositories/context-integration-repository';
import { ContextRepository } from '../../../repositories/context-repository';
import { AttributeMappingRepository } from '../../../repositories/attribute-mapping-repository';
import { ContextHistoryRepository } from '../../../repositories/context-history-repository';
import { SyncHistoryRepository } from '../../../repositories/sync-history-repository';
import { 
  ContextIntegration,
  IntegrationFilterInput,
  PaginationInput,
  CreateContextIntegrationInput,
  UpdateContextIntegrationInput,
  CreateAttributeMappingInput,
  OperationResult,
  SyncResult,
  AttributeMappingInput
} from '../types/generated';
import { Logger } from '../../../../infrastructure/common/logging';
import { AuthorizationService } from '../../../../services/authorization-service';
import { ContextIntegrationValidator } from '../validators/context-integration-validator';
import { AttributeMappingValidator } from '../validators/attribute-mapping-validator';
import { AuditService } from '../../../../services/audit-service';
import { SyncService } from '../../../services/sync-service';
import { v4 as uuidv4 } from 'uuid';

/**
 * Resolvers para as operações de integrações entre contextos
 */
export const contextIntegrationResolvers = {
  Query: {
    /**
     * Obtém uma integração entre contextos pelo ID
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (integrationID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    contextIntegration: async (_, { integrationID }, context) => {
      Logger.debug('Query: contextIntegration', { integrationID });
      
      try {
        const repository = new ContextIntegrationRepository();
        const integration = await repository.getById(integrationID);
        
        if (!integration) {
          Logger.warn('Integration not found', { integrationID });
          return null;
        }
        
        // Verifica autorização - precisa ter permissão em ambos os contextos
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:read', 
          integration.sourceContextID
        );
        
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:read', 
          integration.targetContextID
        );
        
        // Registra auditoria da consulta
        await AuditService.log({
          action: 'READ',
          resource: 'CONTEXT_INTEGRATION',
          resourceId: integrationID,
          userId: context.user?.id,
          details: { 
            sourceContextID: integration.sourceContextID,
            targetContextID: integration.targetContextID,
            integrationType: integration.integrationType
          }
        });
        
        return integration;
      } catch (error) {
        Logger.error('Failed to get context integration', { integrationID, error });
        throw new Error(`Erro ao obter integração entre contextos: ${error.message}`);
      }
    },
    
    /**
     * Lista integrações entre contextos com filtros e paginação
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da query (filter, pagination)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    contextIntegrations: async (_, { filter, pagination }, context) => {
      Logger.debug('Query: contextIntegrations', { filter, pagination });
      
      // Verifica autorização
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:integrations:list'
      );
      
      try {
        const repository = new ContextIntegrationRepository();
        const defaultPagination: PaginationInput = {
          page: 1,
          pageSize: 20
        };
        
        const appliedPagination = pagination || defaultPagination;
        const result = await repository.findAll(filter, appliedPagination);
        
        // Se filtro por contextos específicos, verifica permissão para cada contexto
        if (filter?.contextIds && filter.contextIds.length > 0) {
          for (const contextId of filter.contextIds) {
            await AuthorizationService.checkPermission(
              context.user,
              'identity:context:read',
              contextId
            );
          }
        }
        
        // Registra auditoria da consulta
        await AuditService.log({
          action: 'LIST',
          resource: 'CONTEXT_INTEGRATIONS',
          userId: context.user?.id,
          details: { filter, pagination: appliedPagination }
        });
        
        return {
          items: result.items,
          totalCount: result.totalCount,
          hasMore: result.hasMore
        };
      } catch (error) {
        Logger.error('Failed to list context integrations', { filter, pagination, error });
        throw new Error(`Erro ao listar integrações entre contextos: ${error.message}`);
      }
    }
  },
  
  Mutation: {
    /**
     * Cria uma nova integração entre contextos
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (input)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    createContextIntegration: async (_, { input }, context) => {
      Logger.debug('Mutation: createContextIntegration', { input });
      
      // Verifica autorização para ambos os contextos
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:integration:create', 
        input.sourceContextId
      );
      
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:integration:create', 
        input.targetContextId
      );
      
      try {
        // Validação dos dados de entrada
        const validator = new ContextIntegrationValidator();
        await validator.validateCreate(input);
        
        // Verifica se os contextos existem
        const contextRepository = new ContextRepository();
        const sourceContext = await contextRepository.getById(input.sourceContextId);
        const targetContext = await contextRepository.getById(input.targetContextId);
        
        if (!sourceContext) {
          throw new Error(`Contexto de origem não encontrado: ${input.sourceContextId}`);
        }
        
        if (!targetContext) {
          throw new Error(`Contexto de destino não encontrado: ${input.targetContextId}`);
        }
        
        // Verifica se já existe uma integração entre os mesmos contextos
        const integrationRepository = new ContextIntegrationRepository();
        const existingIntegration = await integrationRepository.findBySourceAndTarget(
          input.sourceContextId,
          input.targetContextId
        );
        
        if (existingIntegration) {
          throw new Error(`Já existe uma integração entre os contextos especificados`);
        }
        
        // Direção padrão de sincronização se não especificada
        const syncDirection = input.syncDirection || 'BIDIRECTIONAL';
        
        // Prepara dados para inserção
        const integrationData = {
          ...input,
          syncDirection,
          createdBy: context.user?.id,
          createdAt: new Date(),
          updatedAt: new Date()
        };
        
        // Cria a integração
        const newIntegration = await integrationRepository.create(integrationData);
        
        // Processa os mapeamentos de atributos
        if (input.attributeMappings && input.attributeMappings.length > 0) {
          const mappingRepository = new AttributeMappingRepository();
          const mappingValidator = new AttributeMappingValidator();
          
          for (const mapping of input.attributeMappings) {
            const mappingInput: CreateAttributeMappingInput = {
              sourceContextId: input.sourceContextId,
              targetContextId: input.targetContextId,
              sourceAttributeKey: mapping.sourceAttribute,
              targetAttributeKey: mapping.targetAttribute,
              mappingType: 'DIRECT', // Tipo padrão
              transformationRule: mapping.transformation,
              isActive: true
            };
            
            await mappingValidator.validateCreate(mappingInput);
            await mappingRepository.create({
              ...mappingInput,
              mappingID: uuidv4(),
              createdBy: context.user?.id,
              createdAt: new Date(),
              updatedAt: new Date()
            });
          }
        }
        
        // Registra auditoria da criação
        await AuditService.log({
          action: 'CREATE',
          resource: 'CONTEXT_INTEGRATION',
          resourceId: newIntegration.integrationID,
          userId: context.user?.id,
          details: {
            sourceContextId: input.sourceContextId,
            targetContextId: input.targetContextId,
            integrationType: input.integrationType,
            syncMode: input.syncMode,
            attributeMappingsCount: input.attributeMappings?.length || 0
          }
        });
        
        return newIntegration;
      } catch (error) {
        Logger.error('Failed to create context integration', { input, error });
        throw new Error(`Erro ao criar integração entre contextos: ${error.message}`);
      }
    },
    
    /**
     * Atualiza uma integração entre contextos existente
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (input)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    updateContextIntegration: async (_, { input }, context) => {
      Logger.debug('Mutation: updateContextIntegration', { input });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        const existingIntegration = await integrationRepository.getById(input.integrationId);
        
        if (!existingIntegration) {
          throw new Error(`Integração entre contextos não encontrada: ${input.integrationId}`);
        }
        
        // Verifica autorização para ambos os contextos
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:update', 
          existingIntegration.sourceContextID
        );
        
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:update', 
          existingIntegration.targetContextID
        );
        
        // Validação dos dados de entrada
        const validator = new ContextIntegrationValidator();
        await validator.validateUpdate(input);
        
        // Registra os campos alterados para o histórico
        const changedFields = Object.keys(input).filter(key => 
          key !== 'integrationId' && 
          key !== 'attributeMappings' &&
          input[key] !== undefined && 
          input[key] !== existingIntegration[key]
        );
        
        // Prepara dados para atualização
        const integrationData = {
          ...input,
          updatedAt: new Date()
        };
        
        // Atualiza a integração
        const updatedIntegration = await integrationRepository.update(
          input.integrationId, 
          integrationData
        );
        
        // Atualiza mapeamentos de atributos, se fornecidos
        if (input.attributeMappings && input.attributeMappings.length > 0) {
          // Remove mapeamentos antigos
          const mappingRepository = new AttributeMappingRepository();
          const existingMappings = await mappingRepository.findByIntegration(input.integrationId);
          
          for (const mapping of existingMappings) {
            await mappingRepository.delete(mapping.mappingID);
          }
          
          // Cria novos mapeamentos
          const mappingValidator = new AttributeMappingValidator();
          
          for (const mapping of input.attributeMappings) {
            const mappingInput: CreateAttributeMappingInput = {
              sourceContextId: existingIntegration.sourceContextID,
              targetContextId: existingIntegration.targetContextID,
              sourceAttributeKey: mapping.sourceAttribute,
              targetAttributeKey: mapping.targetAttribute,
              mappingType: 'DIRECT', // Tipo padrão
              transformationRule: mapping.transformation,
              isActive: true
            };
            
            await mappingValidator.validateCreate(mappingInput);
            await mappingRepository.create({
              ...mappingInput,
              mappingID: uuidv4(),
              createdBy: context.user?.id,
              createdAt: new Date(),
              updatedAt: new Date()
            });
          }
          
          changedFields.push('attributeMappings');
        }
        
        // Registra histórico de alteração
        if (changedFields.length > 0) {
          const historyEntries = [];
          
          // Registra alterações do contexto de origem
          historyEntries.push({
            contextID: existingIntegration.sourceContextID,
            changeType: 'INTEGRATION_UPDATED',
            changedFields,
            previousValues: changedFields.reduce((acc, field) => {
              acc[field] = existingIntegration[field];
              return acc;
            }, {}),
            newValues: changedFields.reduce((acc, field) => {
              acc[field] = input[field] !== undefined ? input[field] : existingIntegration[field];
              return acc;
            }, {}),
            changedBy: context.user?.id,
            timestamp: new Date(),
            reason: `Integração com contexto ${existingIntegration.targetContextID} atualizada`
          });
          
          // Registra alterações do contexto de destino
          historyEntries.push({
            contextID: existingIntegration.targetContextID,
            changeType: 'INTEGRATION_UPDATED',
            changedFields,
            previousValues: changedFields.reduce((acc, field) => {
              acc[field] = existingIntegration[field];
              return acc;
            }, {}),
            newValues: changedFields.reduce((acc, field) => {
              acc[field] = input[field] !== undefined ? input[field] : existingIntegration[field];
              return acc;
            }, {}),
            changedBy: context.user?.id,
            timestamp: new Date(),
            reason: `Integração com contexto ${existingIntegration.sourceContextID} atualizada`
          });
          
          // Salva os históricos
          const historyRepository = new ContextHistoryRepository();
          for (const entry of historyEntries) {
            await historyRepository.createHistoryEntry(entry);
          }
        }
        
        // Registra auditoria da atualização
        await AuditService.log({
          action: 'UPDATE',
          resource: 'CONTEXT_INTEGRATION',
          resourceId: input.integrationId,
          userId: context.user?.id,
          details: {
            changedFields,
            sourceContextId: existingIntegration.sourceContextID,
            targetContextId: existingIntegration.targetContextID
          }
        });
        
        return updatedIntegration;
      } catch (error) {
        Logger.error('Failed to update context integration', { input, error });
        throw new Error(`Erro ao atualizar integração entre contextos: ${error.message}`);
      }
    },
    
    /**
     * Remove uma integração entre contextos
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (integrationID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    removeContextIntegration: async (_, { integrationID }, context) => {
      Logger.debug('Mutation: removeContextIntegration', { integrationID });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        const existingIntegration = await integrationRepository.getById(integrationID);
        
        if (!existingIntegration) {
          return {
            success: false,
            message: `Integração entre contextos não encontrada: ${integrationID}`,
            code: 'INTEGRATION_NOT_FOUND'
          };
        }
        
        // Verifica autorização para ambos os contextos
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:delete', 
          existingIntegration.sourceContextID
        );
        
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:integration:delete', 
          existingIntegration.targetContextID
        );
        
        // Remove mapeamentos associados
        const mappingRepository = new AttributeMappingRepository();
        const existingMappings = await mappingRepository.findByIntegration(integrationID);
        
        for (const mapping of existingMappings) {
          await mappingRepository.delete(mapping.mappingID);
        }
        
        // Remove histórico de sincronizações
        const syncHistoryRepository = new SyncHistoryRepository();
        await syncHistoryRepository.deleteByIntegration(integrationID);
        
        // Remove a integração
        await integrationRepository.delete(integrationID);
        
        // Registra histórico de alteração
        const historyEntries = [];
        
        // Registra alteração no contexto de origem
        historyEntries.push({
          contextID: existingIntegration.sourceContextID,
          changeType: 'INTEGRATION_REMOVED',
          changedFields: ['integrations'],
          previousValues: {
            integrationID,
            integrationType: existingIntegration.integrationType,
            targetContextID: existingIntegration.targetContextID
          },
          newValues: {},
          changedBy: context.user?.id,
          timestamp: new Date(),
          reason: `Integração com contexto ${existingIntegration.targetContextID} removida`
        });
        
        // Registra alteração no contexto de destino
        historyEntries.push({
          contextID: existingIntegration.targetContextID,
          changeType: 'INTEGRATION_REMOVED',
          changedFields: ['integrations'],
          previousValues: {
            integrationID,
            integrationType: existingIntegration.integrationType,
            sourceContextID: existingIntegration.sourceContextID
          },
          newValues: {},
          changedBy: context.user?.id,
          timestamp: new Date(),
          reason: `Integração com contexto ${existingIntegration.sourceContextID} removida`
        });
        
        // Salva os históricos
        const historyRepository = new ContextHistoryRepository();
        for (const entry of historyEntries) {
          await historyRepository.createHistoryEntry(entry);
        }
        
        // Registra auditoria da remoção
        await AuditService.log({
          action: 'DELETE',
          resource: 'CONTEXT_INTEGRATION',
          resourceId: integrationID,
          userId: context.user?.id,
          details: {
            sourceContextId: existingIntegration.sourceContextID,
            targetContextId: existingIntegration.targetContextID,
            mappingsRemoved: existingMappings.length
          }
        });
        
        return {
          success: true,
          message: `Integração entre contextos removida com sucesso`,
          code: 'INTEGRATION_DELETED'
        };
      } catch (error) {
        Logger.error('Failed to remove context integration', { integrationID, error });
        return {
          success: false,
          message: `Erro ao remover integração entre contextos: ${error.message}`,
          code: 'INTEGRATION_DELETE_FAILED'
        };
      }
    },
    
    /**
     * Cria um mapeamento entre atributos de diferentes contextos
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (input)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    createAttributeMapping: async (_, { input }, context) => {
      Logger.debug('Mutation: createAttributeMapping', { input });
      
      // Verifica autorização para ambos os contextos
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:mapping:create', 
        input.sourceContextId
      );
      
      await AuthorizationService.checkPermission(
        context.user, 
        'identity:mapping:create', 
        input.targetContextId
      );
      
      try {
        // Validação dos dados de entrada
        const validator = new AttributeMappingValidator();
        await validator.validateCreate(input);
        
        // Verifica se os contextos existem
        const contextRepository = new ContextRepository();
        const sourceContext = await contextRepository.getById(input.sourceContextId);
        const targetContext = await contextRepository.getById(input.targetContextId);
        
        if (!sourceContext) {
          throw new Error(`Contexto de origem não encontrado: ${input.sourceContextId}`);
        }
        
        if (!targetContext) {
          throw new Error(`Contexto de destino não encontrado: ${input.targetContextId}`);
        }
        
        // Verifica se já existe um mapeamento para os mesmos atributos
        const mappingRepository = new AttributeMappingRepository();
        const existingMapping = await mappingRepository.findBySourceAndTargetAttribute(
          input.sourceContextId,
          input.targetContextId,
          input.sourceAttributeKey,
          input.targetAttributeKey
        );
        
        if (existingMapping) {
          throw new Error(`Já existe um mapeamento entre os atributos especificados`);
        }
        
        // Prepara dados para inserção
        const mappingData = {
          ...input,
          mappingID: uuidv4(),
          createdBy: context.user?.id,
          createdAt: new Date(),
          updatedAt: new Date()
        };
        
        // Cria o mapeamento
        const newMapping = await mappingRepository.create(mappingData);
        
        // Registra auditoria da criação
        await AuditService.log({
          action: 'CREATE',
          resource: 'ATTRIBUTE_MAPPING',
          resourceId: newMapping.mappingID,
          userId: context.user?.id,
          details: {
            sourceContextId: input.sourceContextId,
            targetContextId: input.targetContextId,
            sourceAttributeKey: input.sourceAttributeKey,
            targetAttributeKey: input.targetAttributeKey
          }
        });
        
        return newMapping;
      } catch (error) {
        Logger.error('Failed to create attribute mapping', { input, error });
        throw new Error(`Erro ao criar mapeamento de atributos: ${error.message}`);
      }
    },
    
    /**
     * Remove um mapeamento entre atributos
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (mappingID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    removeAttributeMapping: async (_, { mappingID }, context) => {
      Logger.debug('Mutation: removeAttributeMapping', { mappingID });
      
      try {
        const mappingRepository = new AttributeMappingRepository();
        const existingMapping = await mappingRepository.getById(mappingID);
        
        if (!existingMapping) {
          return {
            success: false,
            message: `Mapeamento de atributos não encontrado: ${mappingID}`,
            code: 'MAPPING_NOT_FOUND'
          };
        }
        
        // Verifica autorização para ambos os contextos
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:mapping:delete', 
          existingMapping.sourceContextID
        );
        
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:mapping:delete', 
          existingMapping.targetContextID
        );
        
        // Remove o mapeamento
        await mappingRepository.delete(mappingID);
        
        // Registra auditoria da remoção
        await AuditService.log({
          action: 'DELETE',
          resource: 'ATTRIBUTE_MAPPING',
          resourceId: mappingID,
          userId: context.user?.id,
          details: {
            sourceContextId: existingMapping.sourceContextID,
            targetContextId: existingMapping.targetContextID,
            sourceAttributeKey: existingMapping.sourceAttributeKey,
            targetAttributeKey: existingMapping.targetAttributeKey
          }
        });
        
        return {
          success: true,
          message: `Mapeamento de atributos removido com sucesso`,
          code: 'MAPPING_DELETED'
        };
      } catch (error) {
        Logger.error('Failed to remove attribute mapping', { mappingID, error });
        return {
          success: false,
          message: `Erro ao remover mapeamento de atributos: ${error.message}`,
          code: 'MAPPING_DELETE_FAILED'
        };
      }
    },
    
    /**
     * Sincroniza manualmente contextos de identidade
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (integrationID)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    synchronizeContexts: async (_, { integrationID }, context) => {
      Logger.debug('Mutation: synchronizeContexts', { integrationID });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        const integration = await integrationRepository.getById(integrationID);
        
        if (!integration) {
          throw new Error(`Integração entre contextos não encontrada: ${integrationID}`);
        }
        
        // Verifica autorização para ambos os contextos
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:context:sync', 
          integration.sourceContextID
        );
        
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:context:sync', 
          integration.targetContextID
        );
        
        // Executa a sincronização
        const syncService = new SyncService();
        const syncResult = await syncService.synchronize(
          integrationID, 
          context.user?.id
        );
        
        // Registra auditoria da sincronização
        await AuditService.log({
          action: 'SYNC',
          resource: 'CONTEXT_INTEGRATION',
          resourceId: integrationID,
          userId: context.user?.id,
          details: {
            syncId: syncResult.syncID,
            sourceContextId: integration.sourceContextID,
            targetContextId: integration.targetContextID,
            success: syncResult.success,
            syncedAttributes: syncResult.syncedAttributes.length,
            failedAttributes: syncResult.failedAttributes.length,
            conflictedAttributes: Object.keys(syncResult.conflictedAttributes).length
          }
        });
        
        return syncResult;
      } catch (error) {
        Logger.error('Failed to synchronize contexts', { integrationID, error });
        throw new Error(`Erro ao sincronizar contextos: ${error.message}`);
      }
    },
    
    /**
     * Aprova uma sincronização que requer aprovação
     * @param _ - Parent object (não utilizado)
     * @param args - Argumentos da mutação (syncID, approvedAttributes)
     * @param context - Contexto GraphQL contendo dados de autenticação e autorização
     */
    approveSync: async (_, { syncID, approvedAttributes }, context) => {
      Logger.debug('Mutation: approveSync', { syncID, approvedAttributes });
      
      try {
        const syncHistoryRepository = new SyncHistoryRepository();
        const syncRecord = await syncHistoryRepository.getById(syncID);
        
        if (!syncRecord) {
          throw new Error(`Registro de sincronização não encontrado: ${syncID}`);
        }
        
        const integrationRepository = new ContextIntegrationRepository();
        const integration = await integrationRepository.getById(syncRecord.integrationID);
        
        if (!integration) {
          throw new Error(`Integração associada à sincronização não encontrada`);
        }
        
        // Verifica autorização para o contexto de destino (onde as alterações serão aplicadas)
        await AuthorizationService.checkPermission(
          context.user, 
          'identity:sync:approve', 
          integration.targetContextID
        );
        
        // Executa a aprovação da sincronização
        const syncService = new SyncService();
        const syncResult = await syncService.approvePendingSync(
          syncID,
          approvedAttributes,
          context.user?.id
        );
        
        // Registra auditoria da aprovação
        await AuditService.log({
          action: 'APPROVE_SYNC',
          resource: 'SYNC_OPERATION',
          resourceId: syncID,
          userId: context.user?.id,
          details: {
            integrationId: integration.integrationID,
            sourceContextId: integration.sourceContextID,
            targetContextId: integration.targetContextID,
            approvedAttributes,
            success: syncResult.success
          }
        });
        
        return syncResult;
      } catch (error) {
        Logger.error('Failed to approve sync', { syncID, approvedAttributes, error });
        throw new Error(`Erro ao aprovar sincronização: ${error.message}`);
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
  
  ContextIntegration: {
    /**
     * Resolver para obter o contexto de origem da integração
     * @param parent - Integração entre contextos
     */
    sourceContext: async (parent) => {
      Logger.debug('Field resolver: ContextIntegration.sourceContext', { 
        integrationID: parent.integrationID, sourceContextID: parent.sourceContextID 
      });
      
      try {
        const contextRepository = new ContextRepository();
        return await contextRepository.getById(parent.sourceContextID);
      } catch (error) {
        Logger.error('Failed to resolve source context', { 
          integrationID: parent.integrationID, sourceContextID: parent.sourceContextID, error 
        });
        throw new Error(`Erro ao resolver contexto de origem: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter o contexto de destino da integração
     * @param parent - Integração entre contextos
     */
    targetContext: async (parent) => {
      Logger.debug('Field resolver: ContextIntegration.targetContext', { 
        integrationID: parent.integrationID, targetContextID: parent.targetContextID 
      });
      
      try {
        const contextRepository = new ContextRepository();
        return await contextRepository.getById(parent.targetContextID);
      } catch (error) {
        Logger.error('Failed to resolve target context', { 
          integrationID: parent.integrationID, targetContextID: parent.targetContextID, error 
        });
        throw new Error(`Erro ao resolver contexto de destino: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter os mapeamentos de atributos desta integração
     * @param parent - Integração entre contextos
     */
    attributeMappings: async (parent) => {
      Logger.debug('Field resolver: ContextIntegration.attributeMappings', { 
        integrationID: parent.integrationID
      });
      
      try {
        const mappingRepository = new AttributeMappingRepository();
        return await mappingRepository.findByIntegration(parent.integrationID);
      } catch (error) {
        Logger.error('Failed to resolve attribute mappings', { 
          integrationID: parent.integrationID, error 
        });
        throw new Error(`Erro ao resolver mapeamentos de atributos: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter o histórico de sincronizações desta integração
     * @param parent - Integração entre contextos
     * @param args - Argumentos do resolver (limit, offset)
     */
    syncHistory: async (parent, { limit, offset }) => {
      Logger.debug('Field resolver: ContextIntegration.syncHistory', { 
        integrationID: parent.integrationID, limit, offset
      });
      
      try {
        const syncHistoryRepository = new SyncHistoryRepository();
        return await syncHistoryRepository.findByIntegration(
          parent.integrationID,
          limit || 10,
          offset || 0
        );
      } catch (error) {
        Logger.error('Failed to resolve sync history', { 
          integrationID: parent.integrationID, limit, offset, error 
        });
        throw new Error(`Erro ao resolver histórico de sincronizações: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter a última sincronização desta integração
     * @param parent - Integração entre contextos
     */
    lastSync: async (parent) => {
      Logger.debug('Field resolver: ContextIntegration.lastSync', { 
        integrationID: parent.integrationID
      });
      
      try {
        const syncHistoryRepository = new SyncHistoryRepository();
        const results = await syncHistoryRepository.findByIntegration(parent.integrationID, 1, 0);
        return results.length > 0 ? results[0] : null;
      } catch (error) {
        Logger.error('Failed to resolve last sync', { 
          integrationID: parent.integrationID, error 
        });
        throw new Error(`Erro ao resolver última sincronização: ${error.message}`);
      }
    }
  },
  
  AttributeMapping: {
    /**
     * Resolver para obter o atributo de origem do mapeamento
     * @param parent - Mapeamento de atributos
     */
    sourceAttribute: async (parent) => {
      Logger.debug('Field resolver: AttributeMapping.sourceAttribute', { 
        mappingID: parent.mappingID, sourceContextID: parent.sourceContextID, 
        sourceAttributeKey: parent.sourceAttributeKey
      });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        return await attributeRepository.findByContextAndKey(
          parent.sourceContextID,
          parent.sourceAttributeKey
        );
      } catch (error) {
        Logger.error('Failed to resolve source attribute', { 
          mappingID: parent.mappingID, sourceContextID: parent.sourceContextID, 
          sourceAttributeKey: parent.sourceAttributeKey, error 
        });
        throw new Error(`Erro ao resolver atributo de origem: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter o atributo de destino do mapeamento
     * @param parent - Mapeamento de atributos
     */
    targetAttribute: async (parent) => {
      Logger.debug('Field resolver: AttributeMapping.targetAttribute', { 
        mappingID: parent.mappingID, targetContextID: parent.targetContextID, 
        targetAttributeKey: parent.targetAttributeKey
      });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        return await attributeRepository.findByContextAndKey(
          parent.targetContextID,
          parent.targetAttributeKey
        );
      } catch (error) {
        Logger.error('Failed to resolve target attribute', { 
          mappingID: parent.mappingID, targetContextID: parent.targetContextID, 
          targetAttributeKey: parent.targetAttributeKey, error 
        });
        throw new Error(`Erro ao resolver atributo de destino: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter a integração à qual este mapeamento pertence
     * @param parent - Mapeamento de atributos
     */
    integration: async (parent) => {
      Logger.debug('Field resolver: AttributeMapping.integration', { 
        mappingID: parent.mappingID, sourceContextID: parent.sourceContextID, 
        targetContextID: parent.targetContextID
      });
      
      try {
        const integrationRepository = new ContextIntegrationRepository();
        return await integrationRepository.findBySourceAndTarget(
          parent.sourceContextID,
          parent.targetContextID
        );
      } catch (error) {
        Logger.error('Failed to resolve integration', { 
          mappingID: parent.mappingID, sourceContextID: parent.sourceContextID, 
          targetContextID: parent.targetContextID, error 
        });
        throw new Error(`Erro ao resolver integração: ${error.message}`);
      }
    }
  },