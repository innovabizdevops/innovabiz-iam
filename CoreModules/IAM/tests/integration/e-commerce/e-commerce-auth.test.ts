/**
 * Testes de integração para autenticação e autorização do E-Commerce
 * 
 * Verifica a integração entre IAM e E-Commerce para:
 * - Autenticação de usuários
 * - Gestão de permissões de acesso
 * - Single Sign-On entre plataformas
 * - Validação de tokens de acesso
 * - Autorização contextual multi-tenant
 */

import { ApolloServer } from 'apollo-server-express';
import { createTestClient } from 'apollo-server-testing';
import { gql } from 'apollo-server-core';
import jwt from 'jsonwebtoken';
import { typeDefs } from '../../../src/api/graphql/schema';
import { resolvers } from '../../../src/api/graphql/resolvers';
import { mockEcommerceService } from '../mocks/e-commerce-service.mock';
import { mockIamService } from '../mocks/iam-service.mock';

// Mock dos módulos de serviço
jest.mock('../../../src/integration/ecommerce/service', () => mockEcommerceService);
jest.mock('../../../src/services/iam-service', () => ({
  IamService: mockIamService
}));

// Chave secreta para testes de JWT
const JWT_SECRET = 'test-jwt-secret-key';

// Auxiliar para criar tokens JWT para testes
const generateTestToken = (payload: any) => {
  return jwt.sign(payload, JWT_SECRET, { expiresIn: '1h' });
};

// Configuração do servidor de teste
const setupTestServer = (context = {}) => {
  const server = new ApolloServer({
    typeDefs,
    resolvers,
    context: () => context
  });

  return createTestClient(server);
};

describe('E-Commerce Authentication Integration Tests', () => {
  // Contexto de autenticação padrão para os testes
  const defaultContext = {
    user: {
      id: 'user-123',
      roles: ['CUSTOMER'],
      permissions: ['ecommerce:read', 'ecommerce:write']
    },
    tenantId: 'tenant-123',
    session: {
      sessionId: 'session-123'
    }
  };

  beforeEach(() => {
    // Resetar todos os mocks antes de cada teste
    jest.clearAllMocks();
    
    // Configurar comportamento padrão dos mocks
    mockIamService.prototype.validateToken.mockResolvedValue({
      valid: true,
      payload: {
        userId: 'user-123',
        roles: ['CUSTOMER'],
        permissions: ['ecommerce:read', 'ecommerce:write'],
        tenantId: 'tenant-123'
      }
    });
    
    mockIamService.prototype.getUserPermissions.mockResolvedValue([
      'ecommerce:read', 
      'ecommerce:write', 
      'ecommerce:checkout',
      'ecommerce:order:view'
    ]);
  });

  describe('Mutation: authenticateEcommerce', () => {
    // Definição da query GraphQL
    const AUTHENTICATE_ECOMMERCE = gql`
      mutation AuthenticateEcommerce($input: EcommerceAuthInput!) {
        authenticateEcommerce(input: $input) {
          accessToken
          refreshToken
          user {
            id
            roles
            permissions
            shopId
          }
          expiresIn
          tokenType
        }
      }
    `;

    it('deve autenticar usuário com credenciais válidas', async () => {
      // Mock da resposta de autenticação
      mockEcommerceService.authenticate.mockResolvedValue({
        userId: 'user-123',
        accessToken: 'valid-access-token',
        refreshToken: 'valid-refresh-token',
        roles: ['CUSTOMER'],
        permissions: ['ecommerce:read', 'ecommerce:write'],
        shopId: 'shop-123',
        expiresIn: 3600,
        tokenType: 'Bearer'
      });

      // Configurar o cliente de teste
      const { mutate } = setupTestServer();

      // Executar a mutação
      const authInput = {
        email: 'user@example.com',
        password: 'valid-password',
        tenantId: 'tenant-123'
      };

      const response = await mutate({
        mutation: AUTHENTICATE_ECOMMERCE,
        variables: { input: authInput }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.authenticateEcommerce).toBeDefined();
      expect(response.data.authenticateEcommerce.accessToken).toBe('valid-access-token');
      expect(response.data.authenticateEcommerce.user.id).toBe('user-123');
      expect(response.data.authenticateEcommerce.user.roles).toContain('CUSTOMER');

      // Verificar se o serviço foi chamado corretamente
      expect(mockEcommerceService.authenticate).toHaveBeenCalledWith(
        authInput.email,
        authInput.password,
        authInput.tenantId
      );
    });

    it('deve rejeitar autenticação com credenciais inválidas', async () => {
      // Mock de falha de autenticação
      mockEcommerceService.authenticate.mockRejectedValue(
        new Error('Credenciais inválidas')
      );

      // Configurar o cliente de teste
      const { mutate } = setupTestServer();

      // Executar a mutação
      const authInput = {
        email: 'user@example.com',
        password: 'invalid-password',
        tenantId: 'tenant-123'
      };

      const response = await mutate({
        mutation: AUTHENTICATE_ECOMMERCE,
        variables: { input: authInput }
      });

      // Verificar erro de autenticação
      expect(response.errors).toBeDefined();
      expect(response.errors[0].message).toContain('Credenciais inválidas');
    });
  });

  describe('Query: validateEcommerceToken', () => {
    // Definição da query GraphQL
    const VALIDATE_TOKEN = gql`
      query ValidateEcommerceToken($token: String!) {
        validateEcommerceToken(token: $token) {
          valid
          user {
            id
            roles
            permissions
          }
          expiresAt
        }
      }
    `;

    it('deve validar um token de acesso válido', async () => {
      // Criar token de teste
      const testToken = generateTestToken({
        userId: 'user-123',
        roles: ['CUSTOMER'],
        permissions: ['ecommerce:read']
      });

      // Mock de validação de token bem-sucedida
      mockEcommerceService.validateToken.mockResolvedValue({
        valid: true,
        userId: 'user-123',
        roles: ['CUSTOMER'],
        permissions: ['ecommerce:read', 'ecommerce:write'],
        expiresAt: new Date(Date.now() + 3600000).toISOString()
      });

      // Configurar o cliente de teste
      const { query } = setupTestServer(defaultContext);

      // Executar a query
      const response = await query({
        query: VALIDATE_TOKEN,
        variables: { token: testToken }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.validateEcommerceToken.valid).toBeTruthy();
      expect(response.data.validateEcommerceToken.user).toBeDefined();
      expect(response.data.validateEcommerceToken.user.id).toBe('user-123');

      // Verificar se o serviço foi chamado corretamente
      expect(mockEcommerceService.validateToken).toHaveBeenCalledWith(testToken);
    });

    it('deve rejeitar um token inválido ou expirado', async () => {
      // Mock de validação de token falha
      mockEcommerceService.validateToken.mockResolvedValue({
        valid: false,
        reason: 'Token expirado'
      });

      // Configurar o cliente de teste
      const { query } = setupTestServer(defaultContext);

      // Executar a query
      const response = await query({
        query: VALIDATE_TOKEN,
        variables: { token: 'invalid-or-expired-token' }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.validateEcommerceToken.valid).toBeFalsy();
    });
  });

  describe('Query: ecommerceUserPermissions', () => {
    // Definição da query GraphQL
    const GET_USER_PERMISSIONS = gql`
      query GetEcommerceUserPermissions($userId: ID!, $tenantId: ID!) {
        ecommerceUserPermissions(userId: $userId, tenantId: $tenantId) {
          permissions
          roles
          contexts {
            shopId
            storeId
            warehouseId
            permissions
          }
        }
      }
    `;

    it('deve retornar permissões e contextos para um usuário', async () => {
      // Mock de permissões de usuário
      mockEcommerceService.getUserPermissions.mockResolvedValue({
        permissions: ['ecommerce:read', 'ecommerce:write', 'ecommerce:checkout'],
        roles: ['CUSTOMER', 'PREMIUM_CUSTOMER'],
        contexts: [
          {
            shopId: 'shop-123',
            storeId: 'store-456',
            warehouseId: null,
            permissions: ['ecommerce:order:create', 'ecommerce:product:view']
          },
          {
            shopId: 'shop-789',
            storeId: null,
            warehouseId: null,
            permissions: ['ecommerce:product:view']
          }
        ]
      });

      // Configurar o cliente de teste
      const { query } = setupTestServer(defaultContext);

      // Executar a query
      const response = await query({
        query: GET_USER_PERMISSIONS,
        variables: { 
          userId: 'user-123',
          tenantId: 'tenant-123'
        }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.ecommerceUserPermissions).toBeDefined();
      expect(response.data.ecommerceUserPermissions.permissions).toContain('ecommerce:checkout');
      expect(response.data.ecommerceUserPermissions.roles).toContain('PREMIUM_CUSTOMER');
      expect(response.data.ecommerceUserPermissions.contexts).toHaveLength(2);
      expect(response.data.ecommerceUserPermissions.contexts[0].shopId).toBe('shop-123');
      
      // Verificar se o serviço foi chamado corretamente
      expect(mockEcommerceService.getUserPermissions).toHaveBeenCalledWith(
        'user-123',
        'tenant-123'
      );
    });
  });

  describe('Mutation: refreshEcommerceToken', () => {
    // Definição da query GraphQL
    const REFRESH_TOKEN = gql`
      mutation RefreshEcommerceToken($refreshToken: String!, $tenantId: ID!) {
        refreshEcommerceToken(refreshToken: $refreshToken, tenantId: $tenantId) {
          accessToken
          refreshToken
          expiresIn
        }
      }
    `;

    it('deve renovar token com refresh token válido', async () => {
      // Mock de renovação de token
      mockEcommerceService.refreshToken.mockResolvedValue({
        accessToken: 'new-access-token',
        refreshToken: 'new-refresh-token',
        expiresIn: 3600
      });

      // Configurar o cliente de teste
      const { mutate } = setupTestServer();

      // Executar a mutação
      const response = await mutate({
        mutation: REFRESH_TOKEN,
        variables: { 
          refreshToken: 'valid-refresh-token',
          tenantId: 'tenant-123'
        }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.refreshEcommerceToken).toBeDefined();
      expect(response.data.refreshEcommerceToken.accessToken).toBe('new-access-token');
      expect(response.data.refreshEcommerceToken.refreshToken).toBe('new-refresh-token');

      // Verificar se o serviço foi chamado corretamente
      expect(mockEcommerceService.refreshToken).toHaveBeenCalledWith(
        'valid-refresh-token',
        'tenant-123'
      );
    });

    it('deve rejeitar refresh token inválido', async () => {
      // Mock de falha na renovação de token
      mockEcommerceService.refreshToken.mockRejectedValue(
        new Error('Refresh token inválido ou expirado')
      );

      // Configurar o cliente de teste
      const { mutate } = setupTestServer();

      // Executar a mutação
      const response = await mutate({
        mutation: REFRESH_TOKEN,
        variables: { 
          refreshToken: 'invalid-refresh-token',
          tenantId: 'tenant-123'
        }
      });

      // Verificar erro
      expect(response.errors).toBeDefined();
      expect(response.errors[0].message).toContain('Refresh token inválido');
    });
  });

  describe('Mutation: ssoEcommerceLogin', () => {
    // Definição da query GraphQL
    const SSO_LOGIN = gql`
      mutation SSOEcommerceLogin($token: String!, $tenantId: ID!, $shopId: ID) {
        ssoEcommerceLogin(token: $token, tenantId: $tenantId, shopId: $shopId) {
          accessToken
          refreshToken
          user {
            id
            roles
            shopId
          }
        }
      }
    `;

    it('deve realizar login SSO com token IAM válido', async () => {
      // Criar token IAM de teste
      const iamToken = generateTestToken({
        userId: 'user-123',
        tenantId: 'tenant-123'
      });

      // Mock de login SSO
      mockEcommerceService.ssoLogin.mockResolvedValue({
        userId: 'user-123',
        accessToken: 'ecommerce-access-token',
        refreshToken: 'ecommerce-refresh-token',
        roles: ['CUSTOMER'],
        shopId: 'shop-456',
        expiresIn: 3600
      });

      // Configurar o cliente de teste
      const { mutate } = setupTestServer();

      // Executar a mutação
      const response = await mutate({
        mutation: SSO_LOGIN,
        variables: { 
          token: iamToken,
          tenantId: 'tenant-123',
          shopId: 'shop-456'
        }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.ssoEcommerceLogin).toBeDefined();
      expect(response.data.ssoEcommerceLogin.accessToken).toBe('ecommerce-access-token');
      expect(response.data.ssoEcommerceLogin.user.id).toBe('user-123');
      expect(response.data.ssoEcommerceLogin.user.shopId).toBe('shop-456');

      // Verificar se o serviço foi chamado corretamente
      expect(mockEcommerceService.ssoLogin).toHaveBeenCalledWith(
        iamToken,
        'tenant-123',
        'shop-456'
      );
    });
  });
});