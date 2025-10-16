/**
 * Mock do serviço de E-Commerce para testes de integração
 */

export const mockEcommerceService = {
  authenticate: jest.fn(),
  validateToken: jest.fn(),
  refreshToken: jest.fn(),
  getUserPermissions: jest.fn(),
  ssoLogin: jest.fn(),
  logoutUser: jest.fn(),
  registerUser: jest.fn(),
  updateUserProfile: jest.fn(),
  getShopInfo: jest.fn(),
  getStoresByUser: jest.fn(),
  checkUserAccess: jest.fn()
};

// Mock de autenticação bem-sucedida
mockEcommerceService.authenticate.mockResolvedValue({
  userId: 'user-123',
  accessToken: 'valid-access-token',
  refreshToken: 'valid-refresh-token',
  roles: ['CUSTOMER'],
  permissions: ['ecommerce:read', 'ecommerce:write', 'ecommerce:checkout'],
  shopId: 'shop-123',
  expiresIn: 3600,
  tokenType: 'Bearer'
});

// Mock de validação de token
mockEcommerceService.validateToken.mockResolvedValue({
  valid: true,
  userId: 'user-123',
  roles: ['CUSTOMER'],
  permissions: ['ecommerce:read', 'ecommerce:write'],
  expiresAt: new Date(Date.now() + 3600000).toISOString()
});

// Mock de renovação de token
mockEcommerceService.refreshToken.mockResolvedValue({
  accessToken: 'new-access-token',
  refreshToken: 'new-refresh-token',
  expiresIn: 3600
});

// Mock de permissões de usuário
mockEcommerceService.getUserPermissions.mockResolvedValue({
  permissions: ['ecommerce:read', 'ecommerce:write', 'ecommerce:checkout'],
  roles: ['CUSTOMER'],
  contexts: [
    {
      shopId: 'shop-123',
      storeId: 'store-456',
      warehouseId: null,
      permissions: ['ecommerce:order:create', 'ecommerce:product:view']
    }
  ]
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

// Mock de informações da loja
mockEcommerceService.getShopInfo.mockResolvedValue({
  id: 'shop-123',
  name: 'Loja Teste',
  description: 'Loja de teste para E-Commerce',
  status: 'ACTIVE',
  logo: 'https://example.com/logo.png',
  tenantId: 'tenant-123',
  settings: {
    currency: 'AOA',
    languages: ['pt-AO', 'en'],
    paymentMethods: ['MOBILE_MONEY', 'CREDIT_CARD', 'BANK_TRANSFER'],
    deliveryOptions: ['STORE_PICKUP', 'HOME_DELIVERY']
  },
  contactInfo: {
    email: 'contato@lojateste.com',
    phone: '+244123456789',
    address: 'Rua Teste, 123, Luanda'
  }
});