/**
 * Serviço de integração com E-Commerce
 * 
 * Implementação da integração entre o módulo IAM e a plataforma E-Commerce
 * do ecossistema INNOVABIZ, seguindo os princípios de segurança e conformidade.
 */

import axios from 'axios';
import jwt from 'jsonwebtoken';
import { v4 as uuidv4 } from 'uuid';
import { 
  EcommerceAuthResult,
  EcommerceIntegrationService,
  EcommerceUserInfo,
  RefreshTokenResult,
  ShopInfo,
  SSOLoginResult,
  TokenValidationResult,
  UserPermissions 
} from './types';
import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';

/**
 * Implementação do serviço de integração com E-Commerce
 */
export class EcommerceIntegrationServiceImpl implements EcommerceIntegrationService {
  private readonly logger: Logger;
  private readonly metrics: Metrics;
  private readonly tracer: Tracer;
  private readonly configService: any;
  private readonly iamService: any;
  private readonly cacheService: any;
  
  private apiBaseUrl: string;
  private apiKey: string;
  
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    configService: any,
    iamService: any,
    cacheService: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.configService = configService;
    this.iamService = iamService;
    this.cacheService = cacheService;
    
    // Inicializar configurações
    this.loadConfig();
  }
  
  /**
   * Carrega as configurações do serviço
   */
  private loadConfig(): void {
    try {
      const config = this.configService.getServiceConfig('ecommerce');
      this.apiBaseUrl = config.apiBaseUrl || 'https://api.ecommerce.innovabiz.com/v1';
      this.apiKey = config.apiKey;
      
      this.logger.info('EcommerceIntegrationService: Configuração carregada com sucesso');
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao carregar configuração', { error });
      throw new Error('Falha ao carregar configuração do serviço E-Commerce');
    }
  }
  
  /**
   * Autentica um usuário no sistema E-Commerce
   */
  async authenticate(email: string, password: string, tenantId: string): Promise<EcommerceAuthResult> {
    const span = this.tracer.startSpan('ecommerce.authenticate');
    
    try {
      this.logger.info('EcommerceIntegrationService: Iniciando autenticação', { 
        email,
        tenantId,
        correlationId: span.context().traceId
      });
      
      // Registro de métricas
      this.metrics.increment('ecommerce.authentication.attempt', { tenantId });
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/auth/login`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        data: {
          email,
          password
        },
        timeout: 15000
      });
      
      const result = response.data;
      
      // Registro de métricas de sucesso
      this.metrics.increment('ecommerce.authentication.success', { tenantId });
      
      this.logger.info('EcommerceIntegrationService: Autenticação bem-sucedida', { 
        userId: result.userId,
        tenantId
      });
      
      return {
        userId: result.userId,
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        roles: result.roles,
        permissions: result.permissions,
        shopId: result.shopId,
        expiresIn: result.expiresIn,
        tokenType: result.tokenType || 'Bearer'
      };
    } catch (error) {
      // Registro de métricas de falha
      this.metrics.increment('ecommerce.authentication.failure', { 
        tenantId,
        errorType: error.response?.status ? `HTTP_${error.response.status}` : 'NETWORK_ERROR' 
      });
      
      this.logger.error('EcommerceIntegrationService: Falha na autenticação', { 
        email,
        tenantId,
        error,
        errorMessage: error.response?.data?.message || error.message
      });
      
      if (error.response?.status === 401) {
        throw new Error('Credenciais inválidas para E-Commerce');
      } else if (error.response?.status === 403) {
        throw new Error('Usuário não tem permissão para acessar o E-Commerce');
      } else {
        throw new Error(`Erro na autenticação com E-Commerce: ${error.message}`);
      }
    } finally {
      span.end();
    }
  }
  
  /**
   * Valida um token de acesso ao E-Commerce
   */
  async validateToken(token: string): Promise<TokenValidationResult> {
    const span = this.tracer.startSpan('ecommerce.validateToken');
    
    try {
      this.logger.debug('EcommerceIntegrationService: Validando token');
      
      // Verificar cache primeiro
      const cacheKey = `ecommerce:token:${token.substring(0, 10)}`;
      const cachedResult = await this.cacheService.get(cacheKey);
      
      if (cachedResult) {
        this.metrics.increment('ecommerce.token_validation.cache_hit');
        return JSON.parse(cachedResult);
      }
      
      this.metrics.increment('ecommerce.token_validation.cache_miss');
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/auth/validate-token`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Correlation-ID': span.context().traceId
        },
        data: {
          token
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      // Armazenar em cache (15 min ou metade do tempo restante de expiração)
      let expirationTime = 900; // 15 min padrão
      if (result.valid && result.expiresAt) {
        const expiresAt = new Date(result.expiresAt).getTime();
        const now = Date.now();
        const timeUntilExpiration = Math.floor((expiresAt - now) / 1000);
        expirationTime = Math.min(expirationTime, Math.floor(timeUntilExpiration / 2));
      }
      
      // Só armazenar em cache se for válido
      if (result.valid) {
        await this.cacheService.set(cacheKey, JSON.stringify(result), expirationTime);
      }
      
      return {
        valid: result.valid,
        userId: result.userId,
        roles: result.roles,
        permissions: result.permissions,
        expiresAt: result.expiresAt,
        reason: result.reason
      };
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao validar token', { error });
      
      // Em caso de erro, assumimos que o token é inválido
      return {
        valid: false,
        reason: `Erro na validação: ${error.message}`
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Renova um token de acesso expirado
   */
  async refreshToken(refreshToken: string, tenantId: string): Promise<RefreshTokenResult> {
    const span = this.tracer.startSpan('ecommerce.refreshToken');
    
    try {
      this.logger.info('EcommerceIntegrationService: Renovando token', { tenantId });
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/auth/refresh-token`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        data: {
          refreshToken
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      this.logger.info('EcommerceIntegrationService: Token renovado com sucesso');
      
      return {
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        expiresIn: result.expiresIn
      };
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao renovar token', { error, tenantId });
      
      if (error.response?.status === 401) {
        throw new Error('Refresh token inválido ou expirado');
      } else {
        throw new Error(`Erro ao renovar token: ${error.message}`);
      }
    } finally {
      span.end();
    }
  }
  
  /**
   * Realiza o logout do usuário no E-Commerce
   */
  async logout(token: string, tenantId: string): Promise<boolean> {
    const span = this.tracer.startSpan('ecommerce.logout');
    
    try {
      this.logger.info('EcommerceIntegrationService: Realizando logout', { tenantId });
      
      // Invalidar cache
      let tokenIdentifier;
      try {
        const decodedToken = jwt.decode(token);
        if (decodedToken && typeof decodedToken === 'object' && decodedToken.sub) {
          tokenIdentifier = decodedToken.sub;
          const cacheKey = `ecommerce:token:${token.substring(0, 10)}`;
          await this.cacheService.delete(cacheKey);
        }
      } catch (error) {
        // Ignorar erros na decodificação do token
      }
      
      // Chamada à API do E-Commerce
      await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/auth/logout`,
        headers: {
          'Authorization': `Bearer ${token}`,
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 5000
      });
      
      this.logger.info('EcommerceIntegrationService: Logout realizado com sucesso', { 
        tenantId,
        userId: tokenIdentifier
      });
      
      return true;
    } catch (error) {
      this.logger.warn('EcommerceIntegrationService: Erro no logout, mas considerado bem-sucedido', { 
        error,
        tenantId
      });
      
      // Consideramos o logout bem-sucedido mesmo com erro na API
      // pois o importante é invalidar o token localmente
      return true;
    } finally {
      span.end();
    }
  }
  
  /**
   * Realiza login no E-Commerce via SSO com token do IAM
   */
  async ssoLogin(iamToken: string, tenantId: string, shopId?: string): Promise<SSOLoginResult> {
    const span = this.tracer.startSpan('ecommerce.ssoLogin');
    
    try {
      this.logger.info('EcommerceIntegrationService: Iniciando login SSO', { 
        tenantId,
        shopId,
        correlationId: span.context().traceId
      });
      
      // Verificar token IAM
      const iamValidation = await this.iamService.validateToken(iamToken);
      
      if (!iamValidation.valid) {
        throw new Error('Token IAM inválido');
      }
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/auth/sso-login`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        data: {
          iamToken,
          shopId
        },
        timeout: 15000
      });
      
      const result = response.data;
      
      this.logger.info('EcommerceIntegrationService: Login SSO bem-sucedido', { 
        userId: result.userId,
        tenantId,
        shopId
      });
      
      return {
        userId: result.userId,
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        roles: result.roles,
        shopId: result.shopId,
        expiresIn: result.expiresIn
      };
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Falha no login SSO', { 
        error,
        tenantId,
        shopId
      });
      
      if (error.response?.status === 401) {
        throw new Error('Usuário não encontrado ou não associado ao E-Commerce');
      } else if (error.response?.status === 403) {
        throw new Error('Usuário não tem permissão para acessar a loja especificada');
      } else {
        throw new Error(`Erro no login SSO: ${error.message}`);
      }
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém as permissões do usuário no E-Commerce
   */
  async getUserPermissions(userId: string, tenantId: string): Promise<UserPermissions> {
    const span = this.tracer.startSpan('ecommerce.getUserPermissions');
    
    try {
      this.logger.info('EcommerceIntegrationService: Buscando permissões do usuário', { 
        userId,
        tenantId
      });
      
      // Verificar cache
      const cacheKey = `ecommerce:permissions:${userId}`;
      const cachedPermissions = await this.cacheService.get(cacheKey);
      
      if (cachedPermissions) {
        return JSON.parse(cachedPermissions);
      }
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'GET',
        url: `${this.apiBaseUrl}/users/${userId}/permissions`,
        headers: {
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      // Armazenar em cache por 5 minutos
      await this.cacheService.set(cacheKey, JSON.stringify(result), 300);
      
      return {
        permissions: result.permissions,
        roles: result.roles,
        contexts: result.contexts || []
      };
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao buscar permissões', { 
        error,
        userId,
        tenantId
      });
      
      throw new Error(`Erro ao buscar permissões do usuário: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica se um usuário tem determinada permissão no E-Commerce
   */
  async checkUserAccess(userId: string, permission: string, context?: {shopId?: string, storeId?: string}): Promise<boolean> {
    const span = this.tracer.startSpan('ecommerce.checkUserAccess');
    
    try {
      // Obtém todas as permissões do usuário (utiliza cache quando disponível)
      const tenantId = await this.resolveUserTenant(userId);
      const permissions = await this.getUserPermissions(userId, tenantId);
      
      // Verifica permissões globais
      if (permissions.permissions.includes(permission)) {
        return true;
      }
      
      // Se não tem contexto específico, não há mais o que verificar
      if (!context) {
        return false;
      }
      
      // Verifica permissões específicas de contexto
      for (const ctx of permissions.contexts) {
        if (
          (context.shopId && ctx.shopId === context.shopId) ||
          (context.storeId && ctx.storeId === context.storeId)
        ) {
          if (ctx.permissions.includes(permission)) {
            return true;
          }
        }
      }
      
      return false;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao verificar acesso', { 
        error,
        userId,
        permission
      });
      
      // Em caso de erro, negamos o acesso por segurança
      return false;
    } finally {
      span.end();
    }
  }
  
  /**
   * Registra um novo usuário no E-Commerce
   */
  async registerUser(userInfo: any, tenantId: string): Promise<{userId: string}> {
    const span = this.tracer.startSpan('ecommerce.registerUser');
    
    try {
      this.logger.info('EcommerceIntegrationService: Registrando novo usuário', { 
        email: userInfo.email,
        tenantId
      });
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'POST',
        url: `${this.apiBaseUrl}/users`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        data: userInfo,
        timeout: 15000
      });
      
      const result = response.data;
      
      this.logger.info('EcommerceIntegrationService: Usuário registrado com sucesso', { 
        userId: result.userId,
        tenantId
      });
      
      return {
        userId: result.userId
      };
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao registrar usuário', { 
        error,
        email: userInfo.email,
        tenantId
      });
      
      if (error.response?.status === 409) {
        throw new Error('E-mail já registrado no sistema E-Commerce');
      } else {
        throw new Error(`Erro ao registrar usuário no E-Commerce: ${error.message}`);
      }
    } finally {
      span.end();
    }
  }
  
  /**
   * Atualiza o perfil do usuário no E-Commerce
   */
  async updateUserProfile(userId: string, profileData: any, tenantId: string): Promise<boolean> {
    const span = this.tracer.startSpan('ecommerce.updateUserProfile');
    
    try {
      this.logger.info('EcommerceIntegrationService: Atualizando perfil de usuário', { 
        userId,
        tenantId
      });
      
      // Chamada à API do E-Commerce
      await axios({
        method: 'PUT',
        url: `${this.apiBaseUrl}/users/${userId}`,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        data: profileData,
        timeout: 15000
      });
      
      // Invalidar cache
      await this.cacheService.delete(`ecommerce:user:${userId}`);
      
      this.logger.info('EcommerceIntegrationService: Perfil atualizado com sucesso', { 
        userId,
        tenantId
      });
      
      return true;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao atualizar perfil', { 
        error,
        userId,
        tenantId
      });
      
      throw new Error(`Erro ao atualizar perfil: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém informações detalhadas do usuário
   */
  async getUserInfo(userId: string, tenantId: string): Promise<EcommerceUserInfo> {
    const span = this.tracer.startSpan('ecommerce.getUserInfo');
    
    try {
      this.logger.info('EcommerceIntegrationService: Buscando informações do usuário', { 
        userId,
        tenantId
      });
      
      // Verificar cache
      const cacheKey = `ecommerce:user:${userId}`;
      const cachedUserInfo = await this.cacheService.get(cacheKey);
      
      if (cachedUserInfo) {
        return JSON.parse(cachedUserInfo);
      }
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'GET',
        url: `${this.apiBaseUrl}/users/${userId}`,
        headers: {
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      // Armazenar em cache por 10 minutos
      await this.cacheService.set(cacheKey, JSON.stringify(result), 600);
      
      return result;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao buscar informações do usuário', { 
        error,
        userId,
        tenantId
      });
      
      throw new Error(`Erro ao buscar informações do usuário: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém informações de uma loja
   */
  async getShopInfo(shopId: string, tenantId: string): Promise<ShopInfo> {
    const span = this.tracer.startSpan('ecommerce.getShopInfo');
    
    try {
      this.logger.info('EcommerceIntegrationService: Buscando informações da loja', { 
        shopId,
        tenantId
      });
      
      // Verificar cache
      const cacheKey = `ecommerce:shop:${shopId}`;
      const cachedShopInfo = await this.cacheService.get(cacheKey);
      
      if (cachedShopInfo) {
        return JSON.parse(cachedShopInfo);
      }
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'GET',
        url: `${this.apiBaseUrl}/shops/${shopId}`,
        headers: {
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      // Armazenar em cache por 15 minutos
      await this.cacheService.set(cacheKey, JSON.stringify(result), 900);
      
      return result;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao buscar informações da loja', { 
        error,
        shopId,
        tenantId
      });
      
      throw new Error(`Erro ao buscar informações da loja: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém lojas associadas ao usuário
   */
  async getStoresByUser(userId: string, tenantId: string): Promise<{id: string, name: string}[]> {
    const span = this.tracer.startSpan('ecommerce.getStoresByUser');
    
    try {
      this.logger.info('EcommerceIntegrationService: Buscando lojas do usuário', { 
        userId,
        tenantId
      });
      
      // Verificar cache
      const cacheKey = `ecommerce:user:${userId}:stores`;
      const cachedStores = await this.cacheService.get(cacheKey);
      
      if (cachedStores) {
        return JSON.parse(cachedStores);
      }
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'GET',
        url: `${this.apiBaseUrl}/users/${userId}/shops`,
        headers: {
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 10000
      });
      
      const result = response.data;
      
      // Armazenar em cache por 5 minutos
      await this.cacheService.set(cacheKey, JSON.stringify(result), 300);
      
      return result;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao buscar lojas do usuário', { 
        error,
        userId,
        tenantId
      });
      
      throw new Error(`Erro ao buscar lojas do usuário: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Valida uma sessão de checkout
   */
  async validateCheckoutSession(sessionId: string, tenantId: string): Promise<{valid: boolean, amount: number, currency: string}> {
    const span = this.tracer.startSpan('ecommerce.validateCheckoutSession');
    
    try {
      this.logger.info('EcommerceIntegrationService: Validando sessão de checkout', { 
        sessionId,
        tenantId
      });
      
      // Chamada à API do E-Commerce
      const response = await axios({
        method: 'GET',
        url: `${this.apiBaseUrl}/checkout/sessions/${sessionId}/validate`,
        headers: {
          'X-API-Key': this.apiKey,
          'X-Tenant-ID': tenantId,
          'X-Correlation-ID': span.context().traceId
        },
        timeout: 10000
      });
      
      return response.data;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao validar sessão de checkout', { 
        error,
        sessionId,
        tenantId
      });
      
      if (error.response?.status === 404) {
        return { valid: false, amount: 0, currency: 'USD' };
      }
      
      throw new Error(`Erro ao validar sessão de checkout: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Resolve o tenant ID para um usuário específico
   */
  private async resolveUserTenant(userId: string): Promise<string> {
    // Verificar cache
    const cacheKey = `ecommerce:user:${userId}:tenant`;
    const cachedTenantId = await this.cacheService.get(cacheKey);
    
    if (cachedTenantId) {
      return cachedTenantId;
    }
    
    try {
      // Obter do serviço IAM
      const userTenant = await this.iamService.getUserTenant(userId);
      
      // Armazenar em cache por 30 minutos
      await this.cacheService.set(cacheKey, userTenant.tenantId, 1800);
      
      return userTenant.tenantId;
    } catch (error) {
      this.logger.error('EcommerceIntegrationService: Erro ao resolver tenant do usuário', { 
        error,
        userId
      });
      
      throw new Error(`Não foi possível determinar o tenant do usuário: ${error.message}`);
    }
  }
}