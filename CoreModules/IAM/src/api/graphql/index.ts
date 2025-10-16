/**
 * @file index.ts
 * @description Configuração do servidor GraphQL para integração do IAM com DataCore
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { ApolloServer } from 'apollo-server-express';
import { ApolloServerPluginDrainHttpServer } from 'apollo-server-core';
import express from 'express';
import http from 'http';
import { execute, subscribe } from 'graphql';
import { SubscriptionServer } from 'subscriptions-transport-ws';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { applyMiddleware } from 'graphql-middleware';
import { PubSub } from 'graphql-subscriptions';

import typeDefs from './schema';
import resolvers from './resolvers';
import { authenticate } from '../../middleware/auth';
import { createRateLimitRule } from 'graphql-rate-limit';
import { shield, rule, allow, deny } from 'graphql-shield';
import { logger } from '../../utils/logger';
import { createDataSources } from './dataSources';
import { tracingPlugin } from '../../utils/tracing';
import { loggingPlugin } from '../../utils/logging';

const pubsub = new PubSub();

// Regras de autorização para GraphQL Shield
const isAuthenticated = rule({ cache: 'contextual' })(
  async (parent, args, { user }, info) => {
    return user !== undefined && user !== null;
  }
);

const isAdmin = rule({ cache: 'contextual' })(
  async (parent, args, { user }, info) => {
    return user && user.roles && user.roles.includes('ADMIN');
  }
);

const isSystem = rule({ cache: 'contextual' })(
  async (parent, args, { user }, info) => {
    return user && user.roles && user.roles.includes('SYSTEM');
  }
);

const isRiskAnalyst = rule({ cache: 'contextual' })(
  async (parent, args, { user }, info) => {
    return user && user.roles && user.roles.includes('RISK_ANALYST');
  }
);

const isOwner = rule({ cache: 'contextual' })(
  async (parent, args, { user }, info) => {
    if (!user || !args.id) return false;
    // Aqui você teria uma lógica mais robusta para verificar propriedade
    // Esta é uma simplificação
    return args.id === user.id;
  }
);

// Configuração do Rate Limit
const rateLimitRule = createRateLimitRule({
  identifyContext: (ctx) => ctx.user ? ctx.user.id : ctx.ip,
});

// Definição das permissões
const permissions = shield({
  Query: {
    me: isAuthenticated,
    user: isAdmin,
    users: isAdmin,
    tenant: isAdmin,
    tenants: isAdmin,
    session: isAuthenticated,
    sessions: isAdmin,
    userRiskProfile: isRiskAnalyst,
    assessRisk: isSystem,
    healthCheck: allow,
  },
  Mutation: {
    login: allow,
    refreshToken: allow,
    resetPassword: allow,
    confirmPasswordReset: allow,
    logout: isAuthenticated,
    enrollMFA: isAuthenticated,
    verifyMFA: isAuthenticated,
    disableMFA: isAuthenticated,
    createUser: isAdmin,
    updateUser: isAdmin,
    deleteUser: isAdmin,
    updateMyProfile: isAuthenticated,
    changePassword: isAuthenticated,
    invalidateSession: isAuthenticated,
    invalidateAllSessions: isAuthenticated,
    reportRiskEvent: isSystem,
  },
  Subscription: {
    sessionCreated: isAuthenticated,
    sessionInvalidated: isAuthenticated,
    riskLevelChanged: isRiskAnalyst,
  },
});

// Função para inicializar o servidor Apollo/GraphQL
export async function startGraphQLServer(app: express.Application) {
  // Configurar o esquema GraphQL
  const schema = makeExecutableSchema({ typeDefs, resolvers });
  
  // Aplicar middlewares ao esquema (autenticação, rate limiting, etc.)
  const schemaWithMiddleware = applyMiddleware(
    schema,
    permissions,
    // Outras middleware podem ser adicionadas aqui
  );
  
  // Criar servidor HTTP
  const httpServer = http.createServer(app);
  
  // Criar servidor Apollo
  const server = new ApolloServer({
    schema: schemaWithMiddleware,
    context: async ({ req, connection }) => {
      // Para Websockets (subscriptions)
      if (connection) {
        return connection.context;
      }
      
      // Para HTTP (queries e mutations)
      const user = await authenticate(req);
      return {
        user,
        ip: req.ip,
        pubsub,
      };
    },
    dataSources: () => createDataSources(),
    plugins: [
      ApolloServerPluginDrainHttpServer({ httpServer }),
      tracingPlugin,
      loggingPlugin,
      {
        async serverWillStart() {
          return {
            async drainServer() {
              subscriptionServer.close();
            }
          };
        }
      }
    ],
    formatError: (error) => {
      logger.error('GraphQL Error', { 
        message: error.message,
        path: error.path,
        extensions: error.extensions,
        stack: error.extensions?.exception?.stacktrace,
      });
      
      // Ocultar detalhes internos em produção
      if (process.env.NODE_ENV === 'production') {
        return {
          message: error.message,
          path: error.path,
          extensions: {
            code: error.extensions?.code || 'INTERNAL_SERVER_ERROR'
          }
        };
      }
      
      return error;
    },
  });
  
  // Configurar servidor de subscriptions
  const subscriptionServer = SubscriptionServer.create({
    schema,
    execute,
    subscribe,
    async onConnect(connectionParams, webSocket, context) {
      // Autenticar conexões de websocket
      try {
        if (connectionParams.authorization) {
          const token = connectionParams.authorization.split(' ')[1];
          // Função para verificar token e obter usuário
          const user = await verifyToken(token);
          return { user };
        }
        return {};
      } catch (err) {
        throw new Error('Falha na autenticação para websocket');
      }
    },
  }, {
    server: httpServer,
    path: '/graphql',
  });
  
  // Iniciar o servidor Apollo
  await server.start();
  
  // Aplicar middleware do Apollo ao Express
  server.applyMiddleware({ app, path: '/graphql' });
  
  logger.info('Servidor GraphQL iniciado', { path: '/graphql' });
  
  return { server, httpServer };
}

// Função para verificar token JWT
async function verifyToken(token: string) {
  // Implementação real usaria um serviço de autenticação
  try {
    // Aqui você chamaria seu serviço de autenticação
    // const user = await authService.verifyToken(token);
    // return user;
    
    // Simulação para desenvolvimento
    return { id: 'user-123', roles: ['USER'] };
  } catch (error) {
    logger.error('Erro ao verificar token', { error });
    throw new Error('Token inválido ou expirado');
  }
}