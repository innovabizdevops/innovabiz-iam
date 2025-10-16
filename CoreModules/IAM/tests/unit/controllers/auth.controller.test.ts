/**
 * @file auth.controller.test.ts
 * @description Testes unitários para o controlador de autenticação do IAM INNOVABIZ
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';
import { AuthController } from '../../../src/controllers/auth/auth.controller';
import { AuthService } from '../../../src/services/auth/auth.service';
import { WebAuthnService } from '../../../src/services/auth/webauthn.service';
import { TenantService } from '../../../src/services/tenant/tenant.service';
import { UsersService } from '../../../src/services/users/users.service';
import { RiskManagementService } from '../../../src/services/risk-management/risk-management.service';

describe('AuthController', () => {
  let authController: AuthController;
  let authService: AuthService;
  let webAuthnService: WebAuthnService;
  let usersService: UsersService;
  let riskManagementService: RiskManagementService;

  // Mock data
  const testUser = {
    id: 'user123',
    username: 'testuser@example.com',
    password: '$2b$10$GQH.mfmR4Z9Vx0c8KhwZpOUQsJ5YZVK2O6JWcbKZqgPjyb1ydlFe6', // hashed password
    tenantId: 'tenant123',
    displayName: 'Test User',
    email: 'testuser@example.com',
    isActive: true,
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  const mockAuthService = {
    validateUser: jest.fn(),
    login: jest.fn(),
    refresh: jest.fn(),
    logout: jest.fn(),
    validateTwoFactor: jest.fn(),
  };

  const mockWebAuthnService = {
    generateRegistrationOptions: jest.fn(),
    verifyRegistrationResponse: jest.fn(),
    generateAuthenticationOptions: jest.fn(),
    verifyAuthenticationResponse: jest.fn(),
  };

  const mockUsersService = {
    findOne: jest.fn(),
    findById: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
  };

  const mockRiskManagementService = {
    assessRisk: jest.fn(),
    reportAuthEvent: jest.fn(),
  };

  const mockJwtService = {
    sign: jest.fn(),
    verify: jest.fn(),
  };

  const mockTenantService = {
    getTenantById: jest.fn(),
    getTenantConfig: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      imports: [
        ConfigModule.forRoot({
          isGlobal: true,
        }),
      ],
      controllers: [AuthController],
      providers: [
        {
          provide: AuthService,
          useValue: mockAuthService,
        },
        {
          provide: WebAuthnService,
          useValue: mockWebAuthnService,
        },
        {
          provide: UsersService,
          useValue: mockUsersService,
        },
        {
          provide: RiskManagementService,
          useValue: mockRiskManagementService,
        },
        {
          provide: JwtService,
          useValue: mockJwtService,
        },
        {
          provide: TenantService,
          useValue: mockTenantService,
        },
      ],
    }).compile();

    authController = module.get<AuthController>(AuthController);
    authService = module.get<AuthService>(AuthService);
    webAuthnService = module.get<WebAuthnService>(WebAuthnService);
    usersService = module.get<UsersService>(UsersService);
    riskManagementService = module.get<RiskManagementService>(RiskManagementService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });
}