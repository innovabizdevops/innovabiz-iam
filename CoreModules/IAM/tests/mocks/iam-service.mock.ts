/**
 * Mock do serviço IAM para testes de integração
 * Simula autenticação, validação de tokens e verificação de permissões
 */

export const mockIamService = jest.fn().mockImplementation(() => {
  return {
    // Validação de tokens JWT
    validateToken: jest.fn().mockResolvedValue({
      valid: true,
      payload: {
        userId: 'user-123',
        roles: ['USER'],
        permissions: ['iam:read'],
        tenantId: 'tenant-123'
      }
    }),
    
    // Autenticação de usuários
    authenticateUser: jest.fn().mockResolvedValue({
      userId: 'user-123',
      accessToken: 'valid-access-token',
      refreshToken: 'valid-refresh-token',
      expiresIn: 3600
    }),
    
    // Renovação de tokens
    refreshToken: jest.fn().mockResolvedValue({
      accessToken: 'new-access-token',
      refreshToken: 'new-refresh-token',
      expiresIn: 3600
    }),
    
    // Obtenção de permissões de usuário
    getUserPermissions: jest.fn().mockResolvedValue([
      'iam:read',
      'iam:write',
      'mobile_money:read',
      'mobile_money:write',
      'ecommerce:read'
    ]),
    
    // Verificação de permissão específica
    hasPermission: jest.fn().mockImplementation((userId, permission) => {
      const permissionMap = {
        'user-123': ['iam:read', 'iam:write', 'mobile_money:read', 'ecommerce:read'],
        'user-456': ['iam:read', 'mobile_money:read'],
        'admin-789': ['iam:admin', 'mobile_money:admin', 'ecommerce:admin']
      };
      
      if (!permissionMap[userId]) {
        return Promise.resolve(false);
      }
      
      return Promise.resolve(permissionMap[userId].includes(permission));
    }),
    
    // Verificação de papel/função específica
    hasRole: jest.fn().mockImplementation((userId, role) => {
      const roleMap = {
        'user-123': ['USER', 'CUSTOMER'],
        'user-456': ['USER'],
        'admin-789': ['ADMIN', 'SUPER_USER']
      };
      
      if (!roleMap[userId]) {
        return Promise.resolve(false);
      }
      
      return Promise.resolve(roleMap[userId].includes(role));
    }),
    
    // Obtenção de informações de usuário
    getUserInfo: jest.fn().mockImplementation((userId) => {
      const userMap = {
        'user-123': {
          id: 'user-123',
          email: 'user@example.com',
          firstName: 'Test',
          lastName: 'User',
          roles: ['USER', 'CUSTOMER'],
          status: 'ACTIVE'
        },
        'user-456': {
          id: 'user-456',
          email: 'another@example.com',
          firstName: 'Another',
          lastName: 'User',
          roles: ['USER'],
          status: 'ACTIVE'
        },
        'admin-789': {
          id: 'admin-789',
          email: 'admin@example.com',
          firstName: 'Admin',
          lastName: 'User',
          roles: ['ADMIN', 'SUPER_USER'],
          status: 'ACTIVE'
        }
      };
      
      return Promise.resolve(userMap[userId] || null);
    }),
    
    // SSO com outros serviços
    generateSSOToken: jest.fn().mockResolvedValue({
      token: 'sso-token-123',
      expiresIn: 300
    }),
    
    // Verificação de token SSO
    validateSSOToken: jest.fn().mockResolvedValue({
      valid: true,
      userId: 'user-123',
      targetService: 'ecommerce',
      expiresAt: new Date(Date.now() + 300000).toISOString()
    }),
    
    // Verificação de autenticação biométrica
    verifyBiometric: jest.fn().mockResolvedValue({
      valid: true,
      userId: 'user-123',
      biometricType: 'FINGERPRINT'
    }),
    
    // Validação de credentials step-up
    validateStepUpCredentials: jest.fn().mockResolvedValue({
      valid: true,
      elevationLevel: 'HIGH'
    }),
    
    // Registro de atividade de autenticação
    logAuthActivity: jest.fn().mockResolvedValue(true),
    
    // Verificação de contexto de autenticação
    validateAuthContext: jest.fn().mockImplementation((context) => {
      return Promise.resolve({
        valid: true,
        riskLevel: 'LOW',
        requiresStepUp: false
      });
    })
  };
});