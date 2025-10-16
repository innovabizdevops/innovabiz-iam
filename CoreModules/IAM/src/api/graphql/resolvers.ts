/**
 * @file resolvers.ts
 * @description Resolvers GraphQL para integração do IAM com DataCore
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { IResolvers } from 'apollo-server-express';
import { v4 as uuidv4 } from 'uuid';
import { GraphQLScalarType } from 'graphql';
import { Kind } from 'graphql/language';
import { UserInputError, AuthenticationError, ForbiddenError } from 'apollo-server-express';

// Importações de serviços
import UserService from '../../services/user/user.service';
import TenantService from '../../services/tenant/tenant.service';
import SessionService from '../../services/session/session.service';
import AuthService from '../../services/auth/auth.service';
import MFAService from '../../services/mfa/mfa.service';
import RiskService from '../../services/risk/risk.service';
import { logger } from '../../utils/logger';
import { PubSub } from 'graphql-subscriptions';

// Criar instância de PubSub
const pubsub = new PubSub();

// Definir tópicos de evento para subscriptions
const EVENT_SESSION_CREATED = 'SESSION_CREATED';
const EVENT_SESSION_INVALIDATED = 'SESSION_INVALIDATED';
const EVENT_RISK_LEVEL_CHANGED = 'RISK_LEVEL_CHANGED';

// Escalares personalizados
const dateScalar = new GraphQLScalarType({
  name: 'Date',
  description: 'Data representada como string ISO',
  serialize(value) {
    return value.toISOString().split('T')[0];
  },
  parseValue(value) {
    return new Date(value);
  },
  parseLiteral(ast) {
    if (ast.kind === Kind.STRING) {
      return new Date(ast.value);
    }
    return null;
  },
});

const dateTimeScalar = new GraphQLScalarType({
  name: 'DateTime',
  description: 'Data e hora representadas como string ISO',
  serialize(value) {
    return value instanceof Date ? value.toISOString() : value;
  },
  parseValue(value) {
    return new Date(value);
  },
  parseLiteral(ast) {
    if (ast.kind === Kind.STRING) {
      return new Date(ast.value);
    }
    return null;
  },
});

const jsonScalar = new GraphQLScalarType({
  name: 'JSON',
  description: 'Valor JSON representado como objeto JavaScript',
  serialize(value) {
    return value;
  },
  parseValue(value) {
    return value;
  },
  parseLiteral(ast) {
    switch (ast.kind) {
      case Kind.STRING:
        try {
          return JSON.parse(ast.value);
        } catch (e) {
          return ast.value;
        }
      case Kind.OBJECT:
        // @ts-ignore
        const value = Object.create(null);
        // @ts-ignore
        ast.fields.forEach((field) => {
          // @ts-ignore
          value[field.name.value] = parseLiteral(field.value);
        });
        return value;
      default:
        return null;
    }
  },
});

const uuidScalar = new GraphQLScalarType({
  name: 'UUID',
  description: 'UUID representado como string',
  serialize(value) {
    return value.toString();
  },
  parseValue(value) {
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)) {
      throw new UserInputError('UUID inválido');
    }
    return value;
  },
  parseLiteral(ast) {
    if (ast.kind === Kind.STRING) {
      const value = ast.value;
      if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)) {
        throw new UserInputError('UUID inválido');
      }
      return value;
    }
    return null;
  },
});

// Helper para paginação
const handlePagination = (items, pagination, defaultPageSize = 10) => {
  const { first = defaultPageSize, after, last, before } = pagination || {};
  
  // Implementação básica de paginação baseada em cursor
  let startIndex = 0;
  let endIndex = items.length;
  
  if (after) {
    const afterIndex = items.findIndex(item => Buffer.from(item.id).toString('base64') === after);
    if (afterIndex >= 0) {
      startIndex = afterIndex + 1;
    }
  }
  
  if (before) {
    const beforeIndex = items.findIndex(item => Buffer.from(item.id).toString('base64') === before);
    if (beforeIndex >= 0) {
      endIndex = beforeIndex;
    }
  }
  
  // Ajustar com base em first/last
  if (first !== undefined && first > 0) {
    endIndex = Math.min(startIndex + first, endIndex);
  }
  
  if (last !== undefined && last > 0) {
    startIndex = Math.max(endIndex - last, startIndex);
  }
  
  const pageItems = items.slice(startIndex, endIndex);
  
  // Criar edges e pageInfo
  const edges = pageItems.map(item => ({
    cursor: Buffer.from(item.id).toString('base64'),
    node: item,
  }));
  
  const pageInfo = {
    hasNextPage: endIndex < items.length,
    hasPreviousPage: startIndex > 0,
    startCursor: edges.length > 0 ? edges[0].cursor : null,
    endCursor: edges.length > 0 ? edges[edges.length - 1].cursor : null,
    totalCount: items.length,
  };
  
  return {
    edges,
    pageInfo,
  };
};

// Helper para verificação de autenticação
const checkAuth = (context) => {
  if (!context.user) {
    throw new AuthenticationError('Usuário não autenticado');
  }
  return context.user;
};

// Helper para verificação de autorização
const checkRole = (user, requiredRole) => {
  if (!user.roles.includes(requiredRole)) {
    throw new ForbiddenError(`Acesso negado. Função ${requiredRole} necessária.`);
  }
  return true;
};

// Definir resolvers
export const resolvers: IResolvers = {
  // Resolvers para tipos escalares personalizados
  Date: dateScalar,
  DateTime: dateTimeScalar,
  JSON: jsonScalar,
  UUID: uuidScalar,
  
  // Resolvers para tipos de entidade
  User: {
    displayName: (user) => {
      return user.displayName || `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.username;
    },
    roles: async (user, _, { dataSources }) => {
      return user.roles || [];
    },
    permissions: async (user, _, { dataSources }) => {
      try {
        return await dataSources.userAPI.getUserPermissions(user.id, user.tenantId);
      } catch (error) {
        logger.error('Erro ao buscar permissões do usuário', { userId: user.id, error });
        return [];
      }
    },
    authMethods: async (user, _, { dataSources }) => {
      try {
        return await dataSources.authAPI.getUserAuthMethods(user.id);
      } catch (error) {
        logger.error('Erro ao buscar métodos de autenticação', { userId: user.id, error });
        return [];
      }
    },
    sessions: async (user, _, { dataSources }) => {
      try {
        return await dataSources.sessionAPI.getUserActiveSessions(user.id);
      } catch (error) {
        logger.error('Erro ao buscar sessões ativas', { userId: user.id, error });
        return [];
      }
    },
    mfaEnabled: async (user, _, { dataSources }) => {
      try {
        const mfaMethods = await dataSources.mfaAPI.getUserMFAMethods(user.id);
        return mfaMethods && mfaMethods.length > 0;
      } catch (error) {
        logger.error('Erro ao verificar status MFA', { userId: user.id, error });
        return false;
      }
    },
    mfaMethods: async (user, _, { dataSources }) => {
      try {
        return await dataSources.mfaAPI.getUserMFAMethods(user.id);
      } catch (error) {
        logger.error('Erro ao buscar métodos MFA', { userId: user.id, error });
        return [];
      }
    },
    riskProfile: async (user, _, { dataSources }) => {
      try {
        return await dataSources.riskAPI.getUserRiskProfile(user.id, user.tenantId);
      } catch (error) {
        logger.error('Erro ao buscar perfil de risco', { userId: user.id, error });
        return null;
      }
    }
  },
  
  // Queries
  Query: {
    me: async (_, __, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      try {
        return await dataSources.userAPI.getUserById(authenticatedUser.id);
      } catch (error) {
        logger.error('Erro ao buscar usuário atual', { userId: authenticatedUser.id, error });
        throw new Error('Não foi possível buscar dados do usuário');
      }
    },
    
    user: async (_, { id }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        return await dataSources.userAPI.getUserById(id);
      } catch (error) {
        logger.error('Erro ao buscar usuário por ID', { userId: id, error });
        throw new Error('Usuário não encontrado');
      }
    },
    
    users: async (_, { filter, pagination }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        const users = await dataSources.userAPI.getUsers(filter);
        return handlePagination(users, pagination);
      } catch (error) {
        logger.error('Erro ao buscar usuários', { filter, error });
        throw new Error('Não foi possível buscar usuários');
      }
    },
    
    tenant: async (_, { id }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        return await dataSources.tenantAPI.getTenantById(id);
      } catch (error) {
        logger.error('Erro ao buscar tenant', { tenantId: id, error });
        throw new Error('Tenant não encontrado');
      }
    },
    
    tenants: async (_, { filter, pagination }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        const tenants = await dataSources.tenantAPI.getTenants(filter);
        return handlePagination(tenants, pagination);
      } catch (error) {
        logger.error('Erro ao buscar tenants', { filter, error });
        throw new Error('Não foi possível buscar tenants');
      }
    },
    
    session: async (_, { id }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        const session = await dataSources.sessionAPI.getSessionById(id);
        
        // Verificar se o usuário tem permissão para visualizar essa sessão
        if (
          session.userId !== authenticatedUser.id && 
          !authenticatedUser.roles.includes('ADMIN') && 
          !authenticatedUser.roles.includes('SYSTEM')
        ) {
          throw new ForbiddenError('Acesso negado');
        }
        
        return session;
      } catch (error) {
        logger.error('Erro ao buscar sessão', { sessionId: id, error });
        throw new Error('Sessão não encontrada');
      }
    },
    
    sessions: async (_, { filter, pagination }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        const sessions = await dataSources.sessionAPI.getSessions(filter);
        return handlePagination(sessions, pagination);
      } catch (error) {
        logger.error('Erro ao buscar sessões', { filter, error });
        throw new Error('Não foi possível buscar sessões');
      }
    },
    
    userRiskProfile: async (_, { userId }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'RISK_ANALYST');
      
      try {
        return await dataSources.riskAPI.getUserRiskProfile(userId, user.tenantId);
      } catch (error) {
        logger.error('Erro ao buscar perfil de risco', { userId, error });
        throw new Error('Perfil de risco não encontrado');
      }
    },
    
    assessRisk: async (_, { assessment }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'SYSTEM');
      
      try {
        return await dataSources.riskAPI.assessRisk(assessment);
      } catch (error) {
        logger.error('Erro na avaliação de risco', { assessment, error });
        throw new Error('Falha na avaliação de risco');
      }
    },
    
    healthCheck: async () => {
      return true;
    }
  },
  
  // Mutations
  Mutation: {
    login: async (_, { username, password, tenantId }, { dataSources }) => {
      try {
        // Registrar tentativa de login para análise
        logger.info('Tentativa de login', { username, tenantId });
        
        const result = await dataSources.authAPI.login(username, password, tenantId);
        
        // Registrar login bem-sucedido para análise
        logger.info('Login bem-sucedido', { username, tenantId, userId: result.user.id });
        
        return result;
      } catch (error) {
        logger.error('Falha no login', { username, tenantId, error });
        throw new AuthenticationError(error.message || 'Credenciais inválidas');
      }
    },
    
    refreshToken: async (_, { refreshToken }, { dataSources }) => {
      try {
        return await dataSources.authAPI.refreshToken(refreshToken);
      } catch (error) {
        logger.error('Falha ao atualizar token', { error });
        throw new AuthenticationError('Token de atualização inválido ou expirado');
      }
    },
    
    logout: async (_, __, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        await dataSources.sessionAPI.invalidateSession(authenticatedUser.sessionId);
        return { success: true, message: 'Logout bem-sucedido' };
      } catch (error) {
        logger.error('Falha no logout', { userId: authenticatedUser.id, error });
        return { success: false, message: 'Falha ao encerrar sessão', code: 'LOGOUT_FAILED' };
      }
    },
    
    enrollMFA: async (_, { input }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        return await dataSources.mfaAPI.enrollMFAMethod(
          authenticatedUser.id,
          input.type,
          {
            phoneNumber: input.phoneNumber,
            email: input.email
          }
        );
      } catch (error) {
        logger.error('Falha no registro MFA', { userId: authenticatedUser.id, mfaType: input.type, error });
        throw new Error(error.message || 'Falha ao registrar método MFA');
      }
    },
    
    verifyMFA: async (_, { type, code }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        return await dataSources.mfaAPI.verifyMFA(
          authenticatedUser.id,
          type,
          code
        );
      } catch (error) {
        logger.error('Falha na verificação MFA', { userId: authenticatedUser.id, mfaType: type, error });
        throw new Error(error.message || 'Código de verificação inválido');
      }
    },
    
    disableMFA: async (_, { type }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        await dataSources.mfaAPI.disableMFAMethod(authenticatedUser.id, type);
        return { success: true, message: `Método MFA ${type} desativado com sucesso` };
      } catch (error) {
        logger.error('Falha ao desativar MFA', { userId: authenticatedUser.id, mfaType: type, error });
        return { success: false, message: 'Falha ao desativar método MFA', code: 'MFA_DISABLE_FAILED' };
      }
    },
    
    createUser: async (_, { input }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        const newUser = await dataSources.userAPI.createUser(input);
        return { 
          success: true, 
          message: 'Usuário criado com sucesso', 
          code: 'USER_CREATED',
          user: newUser 
        };
      } catch (error) {
        logger.error('Falha ao criar usuário', { input, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao criar usuário', 
          code: 'USER_CREATE_FAILED' 
        };
      }
    },
    
    updateUser: async (_, { id, input }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        const updatedUser = await dataSources.userAPI.updateUser(id, input);
        return { 
          success: true, 
          message: 'Usuário atualizado com sucesso', 
          code: 'USER_UPDATED',
          user: updatedUser 
        };
      } catch (error) {
        logger.error('Falha ao atualizar usuário', { id, input, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao atualizar usuário', 
          code: 'USER_UPDATE_FAILED' 
        };
      }
    },
    
    deleteUser: async (_, { id }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'ADMIN');
      
      try {
        await dataSources.userAPI.deleteUser(id);
        return { success: true, message: 'Usuário excluído com sucesso', code: 'USER_DELETED' };
      } catch (error) {
        logger.error('Falha ao excluir usuário', { id, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao excluir usuário', 
          code: 'USER_DELETE_FAILED' 
        };
      }
    },
    
    updateMyProfile: async (_, { input }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        const updatedUser = await dataSources.userAPI.updateUser(authenticatedUser.id, input);
        return { 
          success: true, 
          message: 'Perfil atualizado com sucesso', 
          code: 'PROFILE_UPDATED',
          user: updatedUser 
        };
      } catch (error) {
        logger.error('Falha ao atualizar perfil', { userId: authenticatedUser.id, input, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao atualizar perfil', 
          code: 'PROFILE_UPDATE_FAILED' 
        };
      }
    },
    
    changePassword: async (_, { input }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        await dataSources.authAPI.changePassword(
          authenticatedUser.id,
          input.currentPassword,
          input.newPassword
        );
        return { success: true, message: 'Senha alterada com sucesso', code: 'PASSWORD_CHANGED' };
      } catch (error) {
        logger.error('Falha na alteração de senha', { userId: authenticatedUser.id, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao alterar senha', 
          code: 'PASSWORD_CHANGE_FAILED' 
        };
      }
    },
    
    resetPassword: async (_, { email, tenantId }, { dataSources }) => {
      try {
        await dataSources.authAPI.requestPasswordReset(email, tenantId);
        return { 
          success: true, 
          message: 'Se o e-mail for válido, um link de redefinição de senha será enviado', 
          code: 'PASSWORD_RESET_REQUESTED' 
        };
      } catch (error) {
        logger.error('Falha na solicitação de redefinição de senha', { email, tenantId, error });
        return { 
          success: true, // Retornar sempre sucesso para evitar enumeração de e-mails
          message: 'Se o e-mail for válido, um link de redefinição de senha será enviado', 
          code: 'PASSWORD_RESET_REQUESTED' 
        };
      }
    },
    
    confirmPasswordReset: async (_, { token, newPassword }, { dataSources }) => {
      try {
        await dataSources.authAPI.confirmPasswordReset(token, newPassword);
        return { 
          success: true, 
          message: 'Senha redefinida com sucesso', 
          code: 'PASSWORD_RESET_SUCCESS' 
        };
      } catch (error) {
        logger.error('Falha na redefinição de senha', { error });
        return { 
          success: false, 
          message: error.message || 'Token inválido ou expirado', 
          code: 'PASSWORD_RESET_FAILED' 
        };
      }
    },
    
    invalidateSession: async (_, { id }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        const session = await dataSources.sessionAPI.getSessionById(id);
        
        // Verificar permissão para invalidar a sessão
        if (
          session.userId !== authenticatedUser.id && 
          !authenticatedUser.roles.includes('ADMIN') && 
          !authenticatedUser.roles.includes('SYSTEM')
        ) {
          throw new ForbiddenError('Acesso negado');
        }
        
        await dataSources.sessionAPI.invalidateSession(id);
        
        // Publicar evento de sessão invalidada
        pubsub.publish(EVENT_SESSION_INVALIDATED, { 
          sessionInvalidated: session 
        });
        
        return { 
          success: true, 
          message: 'Sessão invalidada com sucesso', 
          code: 'SESSION_INVALIDATED' 
        };
      } catch (error) {
        logger.error('Falha ao invalidar sessão', { sessionId: id, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao invalidar sessão', 
          code: 'SESSION_INVALIDATION_FAILED' 
        };
      }
    },
    
    invalidateAllSessions: async (_, __, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      
      try {
        await dataSources.sessionAPI.invalidateAllUserSessions(
          authenticatedUser.id,
          authenticatedUser.sessionId // Exceção para a sessão atual
        );
        
        return { 
          success: true, 
          message: 'Todas as outras sessões foram invalidadas com sucesso', 
          code: 'SESSIONS_INVALIDATED' 
        };
      } catch (error) {
        logger.error('Falha ao invalidar todas as sessões', { userId: authenticatedUser.id, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao invalidar sessões', 
          code: 'SESSIONS_INVALIDATION_FAILED' 
        };
      }
    },
    
    reportRiskEvent: async (_, { userId, eventType, metadata }, { user, dataSources }) => {
      const authenticatedUser = checkAuth({ user });
      checkRole(authenticatedUser, 'SYSTEM');
      
      try {
        const result = await dataSources.riskAPI.reportEvent(userId, eventType, metadata);
        
        // Se o evento alterou o nível de risco, publicar evento
        if (result.riskLevelChanged) {
          const riskProfile = await dataSources.riskAPI.getUserRiskProfile(userId, metadata.tenantId);
          pubsub.publish(EVENT_RISK_LEVEL_CHANGED, { 
            riskLevelChanged: riskProfile 
          });
        }
        
        return { 
          success: true, 
          message: 'Evento de risco reportado com sucesso', 
          code: 'RISK_EVENT_REPORTED' 
        };
      } catch (error) {
        logger.error('Falha ao reportar evento de risco', { userId, eventType, error });
        return { 
          success: false, 
          message: error.message || 'Falha ao reportar evento de risco', 
          code: 'RISK_EVENT_REPORT_FAILED' 
        };
      }
    }
  },
  
  // Subscriptions
  Subscription: {
    sessionCreated: {
      subscribe: (_, { userId }) => pubsub.asyncIterator([EVENT_SESSION_CREATED])
    },
    
    sessionInvalidated: {
      subscribe: (_, { userId }) => pubsub.asyncIterator([EVENT_SESSION_INVALIDATED])
    },
    
    riskLevelChanged: {
      subscribe: (_, { userId }) => pubsub.asyncIterator([EVENT_RISK_LEVEL_CHANGED])
    }
  }
};

export default resolvers;