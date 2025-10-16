import { Logger } from '../../../../infrastructure/observability/logger';
import { ContextAttributeRepository } from '../../../repositories/context-attribute-repository';
import { ContextRepository } from '../../../repositories/context-repository';
import { ContextAttributeValidator } from '../validators/context-attribute-validator';
import { AuditService } from '../../../services/audit-service';
import { 
  CreateContextAttributeInput,
  UpdateContextAttributeInput,
  DeleteContextAttributeInput,
  ContextAttributeFilterInput,
  PageInfo
} from '../types/generated';

/**
 * Resolvers para operações relacionadas aos atributos de contexto
 */
export const contextAttributeResolvers = {
  Query: {
    /**
     * Resolver para buscar um atributo de contexto pelo ID
     * @param parent - Parent resolver
     * @param args - Argumentos da query
     * @param context - Contexto do GraphQL
     */
    contextAttribute: async (parent, { id }, context) => {
      Logger.debug('Query: contextAttribute', { id });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        return await attributeRepository.getById(id);
      } catch (error) {
        Logger.error('Failed to fetch context attribute', { id, error });
        throw new Error(`Erro ao buscar atributo de contexto: ${error.message}`);
      }
    },
    
    /**
     * Resolver para buscar múltiplos atributos de contexto com filtros
     * @param parent - Parent resolver
     * @param args - Argumentos da query (filters, pagination)
     * @param context - Contexto do GraphQL
     */
    contextAttributes: async (parent, { filter, pagination }, context) => {
      Logger.debug('Query: contextAttributes', { filter, pagination });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        const { page = 1, pageSize = 20 } = pagination || {};
        const offset = (page - 1) * pageSize;
        
        const [attributes, totalCount] = await attributeRepository.findWithFilters(
          filter || {},
          pageSize,
          offset
        );
        
        const pageInfo: PageInfo = {
          currentPage: page,
          pageSize,
          totalItems: totalCount,
          totalPages: Math.ceil(totalCount / pageSize),
          hasNextPage: page < Math.ceil(totalCount / pageSize),
          hasPreviousPage: page > 1
        };
        
        return {
          items: attributes,
          pageInfo
        };
      } catch (error) {
        Logger.error('Failed to fetch context attributes', { filter, pagination, error });
        throw new Error(`Erro ao buscar atributos de contexto: ${error.message}`);
      }
    }
  },
  
  Mutation: {
    /**
     * Resolver para criar um novo atributo de contexto
     * @param parent - Parent resolver
     * @param args - Argumentos da mutation (input)
     * @param context - Contexto do GraphQL
     */
    createContextAttribute: async (parent, { input }, context) => {
      Logger.debug('Mutation: createContextAttribute', { input });
      
      try {
        // Valida os dados de entrada
        const validator = new ContextAttributeValidator();
        await validator.validateCreate(input);
        
        // Cria o novo atributo
        const attributeRepository = new ContextAttributeRepository();
        const newAttribute = await attributeRepository.create(input);
        
        // Registra a ação no log de auditoria
        await AuditService.log({
          action: 'CREATE',
          resource: 'CONTEXT_ATTRIBUTE',
          resourceId: newAttribute.attributeId,
          userId: context.user?.id,
          details: {
            contextId: input.contextId,
            attributeKey: input.attributeKey,
            sensitivityLevel: input.sensitivityLevel
          }
        });
        
        return newAttribute;
      } catch (error) {
        Logger.error('Failed to create context attribute', { input, error });
        throw new Error(`Erro ao criar atributo de contexto: ${error.message}`);
      }
    },
    
    /**
     * Resolver para atualizar um atributo de contexto existente
     * @param parent - Parent resolver
     * @param args - Argumentos da mutation (input)
     * @param context - Contexto do GraphQL
     */
    updateContextAttribute: async (parent, { input }, context) => {
      Logger.debug('Mutation: updateContextAttribute', { input });
      
      try {
        // Valida os dados de entrada
        const validator = new ContextAttributeValidator();
        await validator.validateUpdate(input);
        
        // Busca o atributo atual para comparar alterações
        const attributeRepository = new ContextAttributeRepository();
        const existingAttribute = await attributeRepository.getById(input.attributeId);
        
        if (!existingAttribute) {
          throw new Error(`Atributo não encontrado com ID: ${input.attributeId}`);
        }
        
        // Atualiza o atributo
        const updatedAttribute = await attributeRepository.update(input);
        
        // Identifica os campos alterados
        const changedFields = [];
        if (input.attributeValue !== undefined && input.attributeValue !== existingAttribute.attributeValue) {
          changedFields.push('attributeValue');
        }
        if (input.sensitivityLevel !== undefined && input.sensitivityLevel !== existingAttribute.sensitivityLevel) {
          changedFields.push('sensitivityLevel');
        }
        if (input.description !== undefined && input.description !== existingAttribute.description) {
          changedFields.push('description');
        }
        if (input.metadata !== undefined && 
            JSON.stringify(input.metadata) !== JSON.stringify(existingAttribute.metadata)) {
          changedFields.push('metadata');
        }
        
        // Registra a ação no log de auditoria
        await AuditService.log({
          action: 'UPDATE',
          resource: 'CONTEXT_ATTRIBUTE',
          resourceId: updatedAttribute.attributeId,
          userId: context.user?.id,
          details: {
            contextId: existingAttribute.contextId,
            attributeKey: existingAttribute.attributeKey,
            changedFields,
            previousValues: {
              attributeValue: existingAttribute.attributeValue,
              sensitivityLevel: existingAttribute.sensitivityLevel,
              description: existingAttribute.description
            },
            newValues: {
              attributeValue: input.attributeValue,
              sensitivityLevel: input.sensitivityLevel,
              description: input.description
            }
          }
        });
        
        return updatedAttribute;
      } catch (error) {
        Logger.error('Failed to update context attribute', { input, error });
        throw new Error(`Erro ao atualizar atributo de contexto: ${error.message}`);
      }
    },
    
    /**
     * Resolver para excluir um atributo de contexto
     * @param parent - Parent resolver
     * @param args - Argumentos da mutation (input)
     * @param context - Contexto do GraphQL
     */
    deleteContextAttribute: async (parent, { input }, context) => {
      Logger.debug('Mutation: deleteContextAttribute', { input });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        
        // Busca o atributo antes de excluir
        const existingAttribute = await attributeRepository.getById(input.attributeId);
        
        if (!existingAttribute) {
          throw new Error(`Atributo não encontrado com ID: ${input.attributeId}`);
        }
        
        // Verifica se o atributo pode ser excluído
        if (existingAttribute.isSystem) {
          throw new Error(`Não é possível excluir atributos de sistema: ${existingAttribute.attributeKey}`);
        }
        
        // Exclui o atributo
        const result = await attributeRepository.delete(input.attributeId);
        
        // Registra a ação no log de auditoria
        await AuditService.log({
          action: 'DELETE',
          resource: 'CONTEXT_ATTRIBUTE',
          resourceId: input.attributeId,
          userId: context.user?.id,
          details: {
            contextId: existingAttribute.contextId,
            attributeKey: existingAttribute.attributeKey,
            reason: input.reason || 'Não especificado'
          }
        });
        
        return {
          success: result,
          message: result 
            ? `Atributo '${existingAttribute.attributeKey}' excluído com sucesso`
            : `Falha ao excluir atributo '${existingAttribute.attributeKey}'`
        };
      } catch (error) {
        Logger.error('Failed to delete context attribute', { input, error });
        throw new Error(`Erro ao excluir atributo de contexto: ${error.message}`);
      }
    }
  },
  
  ContextAttribute: {
    /**
     * Resolver para buscar o contexto ao qual o atributo pertence
     * @param parent - Atributo de contexto
     */
    context: async (parent) => {
      Logger.debug('Field resolver: ContextAttribute.context', { 
        attributeId: parent.attributeId, contextId: parent.contextId 
      });
      
      try {
        const contextRepository = new ContextRepository();
        return await contextRepository.getById(parent.contextId);
      } catch (error) {
        Logger.error('Failed to resolve context for attribute', { 
          attributeId: parent.attributeId, contextId: parent.contextId, error 
        });
        throw new Error(`Erro ao resolver contexto para o atributo: ${error.message}`);
      }
    },
    
    /**
     * Resolver para buscar o histórico de alterações do atributo
     * @param parent - Atributo de contexto
     * @param args - Argumentos do resolver (limit, offset)
     */
    changeHistory: async (parent, { limit, offset }) => {
      Logger.debug('Field resolver: ContextAttribute.changeHistory', { 
        attributeId: parent.attributeId, limit, offset 
      });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        return await attributeRepository.getChangeHistory(
          parent.attributeId,
          limit || 10,
          offset || 0
        );
      } catch (error) {
        Logger.error('Failed to resolve change history for attribute', { 
          attributeId: parent.attributeId, limit, offset, error 
        });
        throw new Error(`Erro ao resolver histórico de alterações: ${error.message}`);
      }
    },
    
    /**
     * Resolver para verificar permissões de acesso ao atributo
     * @param parent - Atributo de contexto
     * @param args - Argumentos vazios
     * @param context - Contexto do GraphQL com dados do usuário
     */
    permissions: async (parent, args, context) => {
      Logger.debug('Field resolver: ContextAttribute.permissions', { 
        attributeId: parent.attributeId, userId: context.user?.id
      });
      
      try {
        const attributeRepository = new ContextAttributeRepository();
        const permissions = await attributeRepository.getUserPermissions(
          parent.attributeId,
          context.user?.id
        );
        
        return {
          canRead: permissions.canRead,
          canWrite: permissions.canWrite,
          canDelete: permissions.canDelete
        };
      } catch (error) {
        Logger.error('Failed to resolve permissions for attribute', { 
          attributeId: parent.attributeId, userId: context.user?.id, error 
        });
        throw new Error(`Erro ao resolver permissões para o atributo: ${error.message}`);
      }
    },
    
    /**
     * Resolver para obter os metadados do atributo
     * @param parent - Atributo de contexto
     */
    metadata: (parent) => {
      return parent.metadata || {};
    }
  }
};