/**
 * Tipos e interfaces para o adaptador de integração com E-Commerce
 * 
 * Este arquivo define todas as interfaces e tipos necessários para a integração
 * do módulo IAM com a plataforma E-Commerce do ecossistema INNOVABIZ.
 */

/**
 * Tipos de usuário no sistema E-Commerce
 */
export enum EcommerceUserRole {
  CUSTOMER = 'CUSTOMER',
  MERCHANT = 'MERCHANT',
  ADMIN = 'ADMIN',
  SHOP_MANAGER = 'SHOP_MANAGER',
  SALES_AGENT = 'SALES_AGENT',
  SUPPORT = 'SUPPORT',
  WAREHOUSE_MANAGER = 'WAREHOUSE_MANAGER',
  LOGISTICS_OPERATOR = 'LOGISTICS_OPERATOR'
}

/**
 * Tipos de permissões no sistema E-Commerce
 */
export enum EcommercePermission {
  // Permissões de produtos
  PRODUCT_VIEW = 'ecommerce:product:view',
  PRODUCT_CREATE = 'ecommerce:product:create',
  PRODUCT_EDIT = 'ecommerce:product:edit',
  PRODUCT_DELETE = 'ecommerce:product:delete',
  
  // Permissões de pedidos
  ORDER_VIEW = 'ecommerce:order:view',
  ORDER_CREATE = 'ecommerce:order:create',
  ORDER_EDIT = 'ecommerce:order:edit',
  ORDER_CANCEL = 'ecommerce:order:cancel',
  ORDER_REFUND = 'ecommerce:order:refund',
  
  // Permissões de lojas
  SHOP_VIEW = 'ecommerce:shop:view',
  SHOP_MANAGE = 'ecommerce:shop:manage',
  SHOP_CREATE = 'ecommerce:shop:create',
  
  // Permissões de checkout
  CHECKOUT = 'ecommerce:checkout',
  
  // Permissões de clientes
  CUSTOMER_VIEW = 'ecommerce:customer:view',
  CUSTOMER_EDIT = 'ecommerce:customer:edit',
  
  // Permissões administrativas
  ADMIN_ACCESS = 'ecommerce:admin:access',
  SETTINGS_MANAGE = 'ecommerce:settings:manage',
  REPORTS_VIEW = 'ecommerce:reports:view',
  
  // Permissões de estoque
  INVENTORY_VIEW = 'ecommerce:inventory:view',
  INVENTORY_MANAGE = 'ecommerce:inventory:manage',
  
  // Permissões de pagamento
  PAYMENT_PROCESS = 'ecommerce:payment:process',
  PAYMENT_REFUND = 'ecommerce:payment:refund',
  PAYMENT_VIEW = 'ecommerce:payment:view'
}

/**
 * Interface para entrada de autenticação no E-Commerce
 */
export interface EcommerceAuthInput {
  email: string;
  password: string;
  tenantId: string;
  shopId?: string;
  deviceInfo?: {
    deviceId?: string;
    ipAddress?: string;
    userAgent?: string;
  };
}

/**
 * Interface para saída de autenticação no E-Commerce
 */
export interface EcommerceAuthResult {
  userId: string;
  accessToken: string;
  refreshToken: string;
  roles: EcommerceUserRole[];
  permissions: string[];
  shopId?: string;
  expiresIn: number;
  tokenType: string;
}

/**
 * Interface para validação de token de acesso
 */
export interface TokenValidationResult {
  valid: boolean;
  userId?: string;
  roles?: EcommerceUserRole[];
  permissions?: string[];
  expiresAt?: string;
  reason?: string;
}

/**
 * Interface para renovação de token
 */
export interface RefreshTokenResult {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

/**
 * Interface para permissões de usuário e seus contextos
 */
export interface UserPermissions {
  permissions: string[];
  roles: EcommerceUserRole[];
  contexts: UserPermissionContext[];
}

/**
 * Interface para contextos de permissão (lojas, filiais, etc.)
 */
export interface UserPermissionContext {
  shopId?: string;
  storeId?: string;
  warehouseId?: string;
  permissions: string[];
}

/**
 * Interface para resultado de login SSO
 */
export interface SSOLoginResult {
  userId: string;
  accessToken: string;
  refreshToken: string;
  roles: EcommerceUserRole[];
  shopId?: string;
  expiresIn: number;
}

/**
 * Interface para informações de loja
 */
export interface ShopInfo {
  id: string;
  name: string;
  description?: string;
  status: 'ACTIVE' | 'INACTIVE' | 'PENDING' | 'SUSPENDED';
  logo?: string;
  tenantId: string;
  settings: {
    currency: string;
    languages: string[];
    paymentMethods: string[];
    deliveryOptions: string[];
  };
  contactInfo: {
    email: string;
    phone: string;
    address: string;
  };
}

/**
 * Interface para informações detalhadas de usuário
 */
export interface EcommerceUserInfo {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phoneNumber?: string;
  roles: EcommerceUserRole[];
  status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED' | 'PENDING_VERIFICATION';
  shops?: {
    id: string;
    name: string;
    role: string;
  }[];
  preferences?: {
    language: string;
    currency: string;
    notifications: {
      email: boolean;
      sms: boolean;
      push: boolean;
    };
  };
  kycStatus?: 'VERIFIED' | 'PENDING' | 'NOT_VERIFIED';
  kycLevel?: 'BASIC' | 'MEDIUM' | 'FULL';
}

/**
 * Interface para serviço de integração com E-Commerce
 */
export interface EcommerceIntegrationService {
  // Autenticação e autorização
  authenticate(email: string, password: string, tenantId: string): Promise<EcommerceAuthResult>;
  validateToken(token: string): Promise<TokenValidationResult>;
  refreshToken(refreshToken: string, tenantId: string): Promise<RefreshTokenResult>;
  logout(token: string, tenantId: string): Promise<boolean>;
  
  // Single Sign-On
  ssoLogin(iamToken: string, tenantId: string, shopId?: string): Promise<SSOLoginResult>;
  
  // Gestão de permissões
  getUserPermissions(userId: string, tenantId: string): Promise<UserPermissions>;
  checkUserAccess(userId: string, permission: string, context?: {shopId?: string, storeId?: string}): Promise<boolean>;
  
  // Gestão de usuários
  registerUser(userInfo: any, tenantId: string): Promise<{userId: string}>;
  updateUserProfile(userId: string, profileData: any, tenantId: string): Promise<boolean>;
  getUserInfo(userId: string, tenantId: string): Promise<EcommerceUserInfo>;
  
  // Gestão de lojas
  getShopInfo(shopId: string, tenantId: string): Promise<ShopInfo>;
  getStoresByUser(userId: string, tenantId: string): Promise<{id: string, name: string}[]>;
  
  // Validação de transações e pagamentos
  validateCheckoutSession(sessionId: string, tenantId: string): Promise<{valid: boolean, amount: number, currency: string}>;
}