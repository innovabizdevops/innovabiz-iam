/**
 * Testes de integração para o serviço de integração com E-Commerce
 * 
 * Este arquivo contém testes abrangentes para a integração entre o IAM
 * e o módulo E-Commerce da plataforma INNOVABIZ.
 */

import axios from 'axios';
import { EcommerceIntegrationServiceImpl } from '../../../services/e-commerce/ecommerce-integration-service';
import { EcommerceUserRole } from '../../../services/e-commerce/types';
import { mockLogger } from '../../mocks/logger.mock';
import { mockMetrics } from '../../mocks/metrics.mock';
import { mockTracer } from '../../mocks/tracer.mock';
import { mockConfigService } from '../../mocks/config-service.mock';
import { mockIamService } from '../../mocks/iam-service.mock';
import { mockCacheService } from '../../mocks/cache-service.mock';

// Mock do módulo axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('E-Commerce Integration Service Integration Tests', () => {
  let ecommerceService: EcommerceIntegrationServiceImpl;
  
  // Dados de teste
  const testTenantId = 'tenant-123';
  const testUserId = 'user-456';
  const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTQ1NiIsImlhdCI6MTUxNjIzOTAyMn0.fake-signature';
  const testRefreshToken = 'refresh-token-789';
  const testEmail = 'test@example.com';
  const testPassword = 'Password123!';
  const testShopId = 'shop-789';

  beforeEach(() => {
    // Limpar mocks
    jest.clearAllMocks();
    
    // Configurar mock do serviço de configuração
    mockConfigService.getServiceConfig.mockReturnValue({
      apiBaseUrl: 'https://api.ecommerce.innovabiz.com/v1',
      apiKey: 'test-api-key'
    });
    
    // Configurar mock do serviço IAM
    mockIamService.validateToken.mockResolvedValue({
      valid: true,
      userId: testUserId,
      permissions: ['iam:user:read', 'ecommerce:access']
    });
    
    mockIamService.getUserTenant.mockResolvedValue({
      tenantId: testTenantId
    });
    
    // Inicializar o serviço
    ecommerceService = new EcommerceIntegrationServiceImpl(
      mockLogger,
      mockMetrics,
      mockTracer,
      mockConfigService,
      mockIamService,
      mockCacheService
    );
  });

  describe('authenticate', () => {
    it('deve autenticar usuário com credenciais válidas', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          userId: testUserId,
          accessToken: testToken,
          refreshToken: testRefreshToken,
          roles: [EcommerceUserRole.CUSTOMER],
          permissions: ['ecommerce:product:view', 'ecommerce:checkout'],
          expiresIn: 3600,
          tokenType: 'Bearer'
        }
      });

      // Chamar método
      const result = await ecommerceService.authenticate(testEmail, testPassword, testTenantId);

      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.userId).toBe(testUserId);
      expect(result.accessToken).toBe(testToken);
      expect(result.refreshToken).toBe(testRefreshToken);
      expect(result.roles).toContain(EcommerceUserRole.CUSTOMER);
      expect(result.permissions).toContain('ecommerce:product:view');
      expect(result.expiresIn).toBe(3600);

      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          url: 'https://api.ecommerce.innovabiz.com/v1/auth/login',
          headers: expect.objectContaining({
            'X-API-Key': 'test-api-key',
            'X-Tenant-ID': testTenantId
          }),
          data: {
            email: testEmail,
            password: testPassword
          }
        })
      );

      // Verificar métricas
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'ecommerce.authentication.attempt', 
        { tenantId: testTenantId }
      );
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'ecommerce.authentication.success', 
        { tenantId: testTenantId }
      );
    });

    it('deve rejeitar autenticação com credenciais inválidas', async () => {
      // Configurar mock de erro axios
      mockedAxios.mockRejectedValueOnce({
        response: {
          status: 401,
          data: {
            message: 'Credenciais inválidas'
          }
        }
      });

      // Chamar método e esperar erro
      await expect(ecommerceService.authenticate(testEmail, 'senha-errada', testTenantId))
        .rejects
        .toThrow('Credenciais inválidas para E-Commerce');

      // Verificar métricas
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'ecommerce.authentication.attempt', 
        { tenantId: testTenantId }
      );
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'ecommerce.authentication.failure', 
        { tenantId: testTenantId, errorType: 'HTTP_401' }
      );
    });
  });

  describe('validateToken', () => {
    it('deve validar token com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          valid: true,
          userId: testUserId,
          roles: [EcommerceUserRole.CUSTOMER],
          permissions: ['ecommerce:product:view', 'ecommerce:checkout'],
          expiresAt: new Date(Date.now() + 3600000).toISOString()
        }
      });

      // Chamar método
      const result = await ecommerceService.validateToken(testToken);

      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.valid).toBe(true);
      expect(result.userId).toBe(testUserId);
      expect(result.roles).toContain(EcommerceUserRole.CUSTOMER);
      expect(result.permissions).toContain('ecommerce:product:view');

      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          url: 'https://api.ecommerce.innovabiz.com/v1/auth/validate-token',
          data: {
            token: testToken
          }
        })
      );
      
      // Verificar cache
      expect(mockCacheService.get).toHaveBeenCalledWith(
        expect.stringContaining('ecommerce:token:')
      );
      expect(mockCacheService.set).toHaveBeenCalled();
    });
    
    it('deve retornar resultado de cache quando disponível', async () => {
      // Configurar mock de cache
      const cachedResult = {
        valid: true,
        userId: testUserId,
        roles: [EcommerceUserRole.CUSTOMER],
        permissions: ['ecommerce:product:view']
      };
      mockCacheService.get.mockResolvedValueOnce(JSON.stringify(cachedResult));
      
      // Chamar método
      const result = await ecommerceService.validateToken(testToken);
      
      // Verificar resultado
      expect(result).toEqual(cachedResult);
      
      // Verificar que axios não foi chamado
      expect(mockedAxios).not.toHaveBeenCalled();
      
      // Verificar métricas de cache
      expect(mockMetrics.increment).toHaveBeenCalledWith('ecommerce.token_validation.cache_hit');
    });
    
    it('deve considerar token inválido em caso de erro', async () => {
      // Configurar mock de erro axios
      mockedAxios.mockRejectedValueOnce(new Error('Erro de rede'));
      
      // Chamar método
      const result = await ecommerceService.validateToken(testToken);
      
      // Verificar que resultado indica token inválido
      expect(result.valid).toBe(false);
      expect(result.reason).toContain('Erro na validação');
    });
  });
  
  describe('refreshToken', () => {
    it('deve renovar token com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          accessToken: 'novo-token-123',
          refreshToken: 'novo-refresh-token-456',
          expiresIn: 3600
        }
      });
      
      // Chamar método
      const result = await ecommerceService.refreshToken(testRefreshToken, testTenantId);
      
      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.accessToken).toBe('novo-token-123');
      expect(result.refreshToken).toBe('novo-refresh-token-456');
      expect(result.expiresIn).toBe(3600);
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          url: 'https://api.ecommerce.innovabiz.com/v1/auth/refresh-token',
          headers: expect.objectContaining({
            'X-Tenant-ID': testTenantId
          }),
          data: {
            refreshToken: testRefreshToken
          }
        })
      );
    });
    
    it('deve rejeitar refresh token inválido', async () => {
      // Configurar mock de erro axios
      mockedAxios.mockRejectedValueOnce({
        response: {
          status: 401,
          data: {
            message: 'Refresh token inválido ou expirado'
          }
        }
      });
      
      // Chamar método e esperar erro
      await expect(ecommerceService.refreshToken('token-invalido', testTenantId))
        .rejects
        .toThrow('Refresh token inválido ou expirado');
    });
  });
  
  describe('logout', () => {
    it('deve realizar logout com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({});
      
      // Chamar método
      const result = await ecommerceService.logout(testToken, testTenantId);
      
      // Verificar resultado
      expect(result).toBe(true);
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          url: 'https://api.ecommerce.innovabiz.com/v1/auth/logout',
          headers: expect.objectContaining({
            'Authorization': `Bearer ${testToken}`,
            'X-Tenant-ID': testTenantId
          })
        })
      );
      
      // Verificar que cache foi limpo
      expect(mockCacheService.delete).toHaveBeenCalledWith(
        expect.stringContaining('ecommerce:token:')
      );
    });
    
    it('deve retornar sucesso mesmo com erro na API', async () => {
      // Configurar mock de erro axios
      mockedAxios.mockRejectedValueOnce(new Error('Erro de rede'));
      
      // Chamar método
      const result = await ecommerceService.logout(testToken, testTenantId);
      
      // Verificar que resultado é verdadeiro mesmo com erro
      expect(result).toBe(true);
    });
  });
  
  describe('ssoLogin', () => {
    it('deve fazer login SSO com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          userId: testUserId,
          accessToken: testToken,
          refreshToken: testRefreshToken,
          roles: [EcommerceUserRole.CUSTOMER],
          shopId: testShopId,
          expiresIn: 3600
        }
      });
      
      // Chamar método
      const result = await ecommerceService.ssoLogin('iam-token-123', testTenantId, testShopId);
      
      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.userId).toBe(testUserId);
      expect(result.accessToken).toBe(testToken);
      expect(result.refreshToken).toBe(testRefreshToken);
      expect(result.shopId).toBe(testShopId);
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'POST',
          url: 'https://api.ecommerce.innovabiz.com/v1/auth/sso-login',
          headers: expect.objectContaining({
            'X-Tenant-ID': testTenantId
          }),
          data: {
            iamToken: 'iam-token-123',
            shopId: testShopId
          }
        })
      );
    });
    
    it('deve rejeitar token IAM inválido', async () => {
      // Configurar mock do serviço IAM para retornar token inválido
      mockIamService.validateToken.mockResolvedValueOnce({
        valid: false,
        reason: 'Token expirado'
      });
      
      // Chamar método e esperar erro
      await expect(ecommerceService.ssoLogin('iam-token-invalido', testTenantId))
        .rejects
        .toThrow('Token IAM inválido');
        
      // Verificar que axios não foi chamado
      expect(mockedAxios).not.toHaveBeenCalled();
    });
  });
  
  describe('getUserPermissions', () => {
    it('deve obter permissões do usuário com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          permissions: ['ecommerce:product:view', 'ecommerce:checkout'],
          roles: [EcommerceUserRole.CUSTOMER],
          contexts: [
            {
              shopId: testShopId,
              permissions: ['ecommerce:order:view', 'ecommerce:order:create']
            }
          ]
        }
      });
      
      // Chamar método
      const result = await ecommerceService.getUserPermissions(testUserId, testTenantId);
      
      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.permissions).toContain('ecommerce:product:view');
      expect(result.roles).toContain(EcommerceUserRole.CUSTOMER);
      expect(result.contexts).toHaveLength(1);
      expect(result.contexts[0].shopId).toBe(testShopId);
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'GET',
          url: `https://api.ecommerce.innovabiz.com/v1/users/${testUserId}/permissions`,
          headers: expect.objectContaining({
            'X-Tenant-ID': testTenantId
          })
        })
      );
      
      // Verificar cache
      expect(mockCacheService.get).toHaveBeenCalledWith(
        `ecommerce:permissions:${testUserId}`
      );
      expect(mockCacheService.set).toHaveBeenCalledWith(
        `ecommerce:permissions:${testUserId}`,
        expect.any(String),
        300
      );
    });
    
    it('deve retornar permissões do cache quando disponíveis', async () => {
      // Configurar mock de cache
      const cachedPermissions = {
        permissions: ['ecommerce:product:view'],
        roles: [EcommerceUserRole.CUSTOMER],
        contexts: []
      };
      mockCacheService.get.mockResolvedValueOnce(JSON.stringify(cachedPermissions));
      
      // Chamar método
      const result = await ecommerceService.getUserPermissions(testUserId, testTenantId);
      
      // Verificar resultado
      expect(result).toEqual(cachedPermissions);
      
      // Verificar que axios não foi chamado
      expect(mockedAxios).not.toHaveBeenCalled();
    });
  });
  
  describe('checkUserAccess', () => {
    it('deve verificar acesso com permissão global', async () => {
      // Configurar mock para getUserPermissions
      jest.spyOn(ecommerceService, 'getUserPermissions').mockResolvedValueOnce({
        permissions: ['ecommerce:product:view', 'ecommerce:checkout'],
        roles: [EcommerceUserRole.CUSTOMER],
        contexts: []
      });
      
      // Chamar método
      const result = await ecommerceService.checkUserAccess(testUserId, 'ecommerce:product:view');
      
      // Verificar resultado
      expect(result).toBe(true);
    });
    
    it('deve verificar acesso com permissão específica de contexto', async () => {
      // Configurar mock para getUserPermissions
      jest.spyOn(ecommerceService, 'getUserPermissions').mockResolvedValueOnce({
        permissions: ['ecommerce:product:view'],
        roles: [EcommerceUserRole.CUSTOMER],
        contexts: [
          {
            shopId: testShopId,
            permissions: ['ecommerce:order:create']
          }
        ]
      });
      
      // Chamar método
      const result = await ecommerceService.checkUserAccess(
        testUserId, 
        'ecommerce:order:create',
        { shopId: testShopId }
      );
      
      // Verificar resultado
      expect(result).toBe(true);
    });
    
    it('deve negar acesso quando permissão não existe', async () => {
      // Configurar mock para getUserPermissions
      jest.spyOn(ecommerceService, 'getUserPermissions').mockResolvedValueOnce({
        permissions: ['ecommerce:product:view'],
        roles: [EcommerceUserRole.CUSTOMER],
        contexts: []
      });
      
      // Chamar método
      const result = await ecommerceService.checkUserAccess(testUserId, 'ecommerce:admin:access');
      
      // Verificar resultado
      expect(result).toBe(false);
    });
  });
  
  describe('getShopInfo', () => {
    it('deve obter informações da loja com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          id: testShopId,
          name: 'Loja Teste',
          status: 'ACTIVE',
          tenantId: testTenantId,
          settings: {
            currency: 'AOA',
            languages: ['pt', 'en'],
            paymentMethods: ['CREDIT_CARD', 'MOBILE_MONEY'],
            deliveryOptions: ['STANDARD', 'EXPRESS']
          },
          contactInfo: {
            email: 'loja@exemplo.com',
            phone: '+244123456789',
            address: 'Luanda, Angola'
          }
        }
      });
      
      // Chamar método
      const result = await ecommerceService.getShopInfo(testShopId, testTenantId);
      
      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.id).toBe(testShopId);
      expect(result.name).toBe('Loja Teste');
      expect(result.status).toBe('ACTIVE');
      expect(result.settings.currency).toBe('AOA');
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'GET',
          url: `https://api.ecommerce.innovabiz.com/v1/shops/${testShopId}`,
          headers: expect.objectContaining({
            'X-Tenant-ID': testTenantId
          })
        })
      );
      
      // Verificar cache
      expect(mockCacheService.get).toHaveBeenCalledWith(
        `ecommerce:shop:${testShopId}`
      );
      expect(mockCacheService.set).toHaveBeenCalled();
    });
  });
  
  describe('validateCheckoutSession', () => {
    it('deve validar sessão de checkout com sucesso', async () => {
      // Configurar mock da resposta axios
      mockedAxios.mockResolvedValueOnce({
        data: {
          valid: true,
          amount: 5000,
          currency: 'AOA'
        }
      });
      
      // Chamar método
      const result = await ecommerceService.validateCheckoutSession('session-123', testTenantId);
      
      // Verificar resultado
      expect(result).toBeDefined();
      expect(result.valid).toBe(true);
      expect(result.amount).toBe(5000);
      expect(result.currency).toBe('AOA');
      
      // Verificar chamada axios
      expect(mockedAxios).toHaveBeenCalledWith(
        expect.objectContaining({
          method: 'GET',
          url: 'https://api.ecommerce.innovabiz.com/v1/checkout/sessions/session-123/validate',
          headers: expect.objectContaining({
            'X-Tenant-ID': testTenantId
          })
        })
      );
    });
    
    it('deve retornar sessão inválida quando não encontrada', async () => {
      // Configurar mock de erro axios (sessão não encontrada)
      mockedAxios.mockRejectedValueOnce({
        response: {
          status: 404,
          data: {
            message: 'Sessão de checkout não encontrada'
          }
        }
      });
      
      // Chamar método
      const result = await ecommerceService.validateCheckoutSession('session-inexistente', testTenantId);
      
      // Verificar resultado
      expect(result.valid).toBe(false);
    });
  });
});