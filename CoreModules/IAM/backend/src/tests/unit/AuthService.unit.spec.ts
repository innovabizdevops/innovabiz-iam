/**
 * 🧪 TESTES UNITÁRIOS - AUTH SERVICE
 * Testes unitários para o serviço de autenticação IAM
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST Cybersecurity Framework, ISO 27001, OWASP Testing Guide
 * Cobertura: Login, Logout, Refresh Token, Validação, Rate Limiting
 */

import { Test, TestingModule } from '@nestjs/testing';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import * as bcrypt from 'bcrypt';

// Service sob teste
import { AuthService } from '../../services/AuthService';

// Entidades
import { User } from '../../entities/User';
import { Session } from '../../entities/Session';
import { AuditLog } from '../../entities/AuditLog';

// DTOs
import { LoginDto } from '../../dto/LoginDto';
import { RefreshTokenDto } from '../../dto/RefreshTokenDto';

// Exceptions
import { UnauthorizedException, TooManyRequestsException } from '@nestjs/common';

describe('AuthService - Testes Unitários', () => {
  let authService: AuthService;
  let userRepository: jest.Mocked<Repository<User>>;
  let sessionRepository: jest.Mocked<Repository<Session>>;
  let auditLogRepository: jest.Mocked<Repository<AuditLog>>;
  let jwtService: jest.Mocked<JwtService>;
  let cacheManager: any;
  let configService: ConfigService;

  // Dados de teste
  const mockUser: Partial<User> = {
    id: '1',
    email: 'test@innovabiz.com',
    username: 'testuser',
    password: '$2b$10$hashedPassword',
    firstName: 'Test',
    lastName: 'User',
    isActive: true,
    isEmailVerified: true,
    tenantId: 'tenant-1',
    roles: [],
    lastLoginAt: null,
    loginAttempts: 0,
    lockedUntil: null
  };

  beforeEach(async () => {
    const mockUserRepository = {
      findOne: jest.fn(),
      save: jest.fn(),
      update: jest.fn(),
      count: jest.fn(),
      find: jest.fn()
    };

    const mockSessionRepository = {
      findOne: jest.fn(),
      save: jest.fn(),
      update: jest.fn(),
      count: jest.fn(),
      find: jest.fn()
    };

    const mockAuditLogRepository = {
      save: jest.fn()
    };

    const mockJwtService = {
      sign: jest.fn(),
      verify: jest.fn()
    };

    const mockCacheManager = {
      get: jest.fn(),
      set: jest.fn(),
      del: jest.fn()
    };

    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthService,
        {
          provide: getRepositoryToken(User),
          useValue: mockUserRepository
        },
        {
          provide: getRepositoryToken(Session),
          useValue: mockSessionRepository
        },
        {
          provide: getRepositoryToken(AuditLog),
          useValue: mockAuditLogRepository
        },
        {
          provide: JwtService,
          useValue: mockJwtService
        },
        {
          provide: CACHE_MANAGER,
          useValue: mockCacheManager
        },
        {
          provide: ConfigService,
          useValue: {
            get: jest.fn((key: string) => {
              const config = {
                'auth.jwt.accessTokenExpiration': '15m',
                'auth.jwt.refreshTokenExpiration': '7d',
                'auth.rateLimit.maxAttempts': 5,
                'auth.rateLimit.windowMs': 900000,
                'auth.session.maxConcurrent': 3
              };
              return config[key];
            })
          }
        }
      ]
    }).compile();

    authService = module.get<AuthService>(AuthService);
    userRepository = module.get(getRepositoryToken(User));
    sessionRepository = module.get(getRepositoryToken(Session));
    auditLogRepository = module.get(getRepositoryToken(AuditLog));
    jwtService = module.get<JwtService>(JwtService) as jest.Mocked<JwtService>;
    cacheManager = module.get(CACHE_MANAGER);
    configService = module.get<ConfigService>(ConfigService);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  /**
   * Testes de Login
   */
  describe('login', () => {
    const loginDto: LoginDto = {
      email: 'test@innovabiz.com',
      password: 'plainPassword',
      tenantId: 'tenant-1'
    };

    it('deve autenticar usuário com credenciais válidas', async () => {
      // Arrange
      userRepository.findOne.mockResolvedValue(mockUser as User);
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(true as never);
      jwtService.sign.mockReturnValueOnce('access-token').mockReturnValueOnce('refresh-token');
      sessionRepository.save.mockResolvedValue({} as Session);
      auditLogRepository.save.mockResolvedValue({} as AuditLog);

      // Act
      const result = await authService.login(loginDto);

      // Assert
      expect(result).toEqual({
        accessToken: 'access-token',
        refreshToken: 'refresh-token',
        user: expect.objectContaining({
          id: mockUser.id,
          email: mockUser.email
        })
      });
      expect(userRepository.findOne).toHaveBeenCalledWith({
        where: { email: loginDto.email, tenantId: loginDto.tenantId },
        relations: ['roles', 'roles.permissions']
      });
      expect(bcrypt.compare).toHaveBeenCalledWith(loginDto.password, mockUser.password);
    });

    it('deve rejeitar credenciais inválidas', async () => {
      // Arrange
      userRepository.findOne.mockResolvedValue(mockUser as User);
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(false as never);

      // Act & Assert
      await expect(authService.login(loginDto)).rejects.toThrow(UnauthorizedException);
      expect(userRepository.update).toHaveBeenCalledWith(
        mockUser.id,
        expect.objectContaining({ loginAttempts: 1 })
      );
    });

    it('deve rejeitar usuário inexistente', async () => {
      // Arrange
      userRepository.findOne.mockResolvedValue(null);

      // Act & Assert
      await expect(authService.login(loginDto)).rejects.toThrow(UnauthorizedException);
    });

    it('deve rejeitar usuário inativo', async () => {
      // Arrange
      const inactiveUser = { ...mockUser, isActive: false };
      userRepository.findOne.mockResolvedValue(inactiveUser as User);

      // Act & Assert
      await expect(authService.login(loginDto)).rejects.toThrow(UnauthorizedException);
    });

    it('deve bloquear usuário após múltiplas tentativas falhadas', async () => {
      // Arrange
      const userWithAttempts = { ...mockUser, loginAttempts: 4 };
      userRepository.findOne.mockResolvedValue(userWithAttempts as User);
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(false as never);

      // Act
      await expect(authService.login(loginDto)).rejects.toThrow(UnauthorizedException);

      // Assert
      expect(userRepository.update).toHaveBeenCalledWith(
        mockUser.id,
        expect.objectContaining({
          loginAttempts: 5,
          lockedUntil: expect.any(Date)
        })
      );
    });

    it('deve aplicar rate limiting por IP', async () => {
      // Arrange
      const rateLimitKey = `login_attempts:192.168.1.1`;
      cacheManager.get.mockResolvedValue(5); // Máximo de tentativas atingido

      // Act & Assert
      await expect(authService.login({
        ...loginDto,
        ipAddress: '192.168.1.1'
      })).rejects.toThrow(TooManyRequestsException);
    });
  });

  /**
   * Testes de Logout
   */
  describe('logout', () => {
    const sessionId = 'session-123';

    it('deve fazer logout com sucesso', async () => {
      // Arrange
      const mockSession = {
        id: sessionId,
        userId: mockUser.id,
        isActive: true
      };
      sessionRepository.findOne.mockResolvedValue(mockSession as Session);

      // Act
      await authService.logout(sessionId);

      // Assert
      expect(sessionRepository.update).toHaveBeenCalledWith(
        sessionId,
        { isActive: false, loggedOutAt: expect.any(Date) }
      );
      expect(auditLogRepository.save).toHaveBeenCalledWith(
        expect.objectContaining({
          action: 'LOGOUT',
          userId: mockUser.id
        })
      );
    });

    it('deve rejeitar logout de sessão inexistente', async () => {
      // Arrange
      sessionRepository.findOne.mockResolvedValue(null);

      // Act & Assert
      await expect(authService.logout(sessionId)).rejects.toThrow(UnauthorizedException);
    });
  });

  /**
   * Testes de Refresh Token
   */
  describe('refreshToken', () => {
    const refreshTokenDto: RefreshTokenDto = {
      refreshToken: 'valid-refresh-token'
    };

    it('deve renovar token com refresh token válido', async () => {
      // Arrange
      const tokenPayload = { sub: mockUser.id, sessionId: 'session-123' };
      jwtService.verify.mockReturnValue(tokenPayload);
      sessionRepository.findOne.mockResolvedValue({
        id: 'session-123',
        userId: mockUser.id,
        isActive: true
      } as Session);
      userRepository.findOne.mockResolvedValue(mockUser as User);
      jwtService.sign.mockReturnValueOnce('new-access-token').mockReturnValueOnce('new-refresh-token');

      // Act
      const result = await authService.refreshToken(refreshTokenDto);

      // Assert
      expect(result).toEqual({
        accessToken: 'new-access-token',
        refreshToken: 'new-refresh-token'
      });
      expect(jwtService.verify).toHaveBeenCalledWith(refreshTokenDto.refreshToken);
    });

    it('deve rejeitar refresh token inválido', async () => {
      // Arrange
      jwtService.verify.mockImplementation(() => {
        throw new Error('Invalid token');
      });

      // Act & Assert
      await expect(authService.refreshToken(refreshTokenDto)).rejects.toThrow(UnauthorizedException);
    });
  });

  /**
   * Testes de Validação de Token
   */
  describe('validateToken', () => {
    const accessToken = 'valid-access-token';

    it('deve validar token válido', async () => {
      // Arrange
      const tokenPayload = { 
        sub: mockUser.id, 
        sessionId: 'session-123',
        tenantId: 'tenant-1'
      };
      jwtService.verify.mockReturnValue(tokenPayload);
      sessionRepository.findOne.mockResolvedValue({
        id: 'session-123',
        userId: mockUser.id,
        isActive: true
      } as Session);
      userRepository.findOne.mockResolvedValue(mockUser as User);

      // Act
      const result = await authService.validateToken(accessToken);

      // Assert
      expect(result).toEqual(
        expect.objectContaining({
          id: mockUser.id,
          email: mockUser.email
        })
      );
    });

    it('deve rejeitar token inválido', async () => {
      // Arrange
      jwtService.verify.mockImplementation(() => {
        throw new Error('Invalid token');
      });

      // Act & Assert
      await expect(authService.validateToken(accessToken)).rejects.toThrow(UnauthorizedException);
    });
  });

  /**
   * Testes de Segurança
   */
  describe('Segurança', () => {
    it('deve hash senha corretamente', async () => {
      // Arrange
      const plainPassword = 'TestPassword123!';
      jest.spyOn(bcrypt, 'hash').mockResolvedValue('hashed-password' as never);

      // Act
      const hashedPassword = await authService.hashPassword(plainPassword);

      // Assert
      expect(bcrypt.hash).toHaveBeenCalledWith(plainPassword, 12);
      expect(hashedPassword).toBe('hashed-password');
    });

    it('deve verificar força da senha', () => {
      // Senhas válidas
      expect(authService.validatePasswordStrength('StrongPass123!')).toBe(true);
      expect(authService.validatePasswordStrength('AnotherStr0ng@')).toBe(true);

      // Senhas inválidas
      expect(authService.validatePasswordStrength('weak')).toBe(false);
      expect(authService.validatePasswordStrength('NoNumbers!')).toBe(false);
      expect(authService.validatePasswordStrength('nonumbersorspecial')).toBe(false);
      expect(authService.validatePasswordStrength('12345678')).toBe(false);
    });

    it('deve gerar tokens seguros', () => {
      // Arrange
      jwtService.sign.mockReturnValue('secure-token');

      // Act
      const token = authService.generateAccessToken({
        sub: mockUser.id,
        email: mockUser.email,
        tenantId: mockUser.tenantId
      });

      // Assert
      expect(jwtService.sign).toHaveBeenCalledWith(
        expect.objectContaining({
          sub: mockUser.id,
          email: mockUser.email,
          tenantId: mockUser.tenantId
        }),
        { expiresIn: '15m' }
      );
      expect(token).toBe('secure-token');
    });

    it('deve limpar dados sensíveis do usuário', () => {
      // Act
      const sanitizedUser = authService.sanitizeUser(mockUser as User);

      // Assert
      expect(sanitizedUser.password).toBeUndefined();
      expect(sanitizedUser.loginAttempts).toBeUndefined();
      expect(sanitizedUser.lockedUntil).toBeUndefined();
      expect(sanitizedUser.email).toBe(mockUser.email);
      expect(sanitizedUser.id).toBe(mockUser.id);
    });
  });

  /**
   * Testes de Rate Limiting
   */
  describe('Rate Limiting', () => {
    it('deve incrementar contador de tentativas', async () => {
      // Arrange
      const ipAddress = '192.168.1.1';
      const rateLimitKey = `login_attempts:${ipAddress}`;
      cacheManager.get.mockResolvedValue(2);

      // Act
      await authService.incrementLoginAttempts(ipAddress);

      // Assert
      expect(cacheManager.set).toHaveBeenCalledWith(
        rateLimitKey,
        3,
        900000 // 15 minutos
      );
    });

    it('deve verificar se IP está bloqueado', async () => {
      // Arrange
      const ipAddress = '192.168.1.1';
      cacheManager.get.mockResolvedValue(6); // Acima do limite

      // Act
      const isBlocked = await authService.isIpBlocked(ipAddress);

      // Assert
      expect(isBlocked).toBe(true);
    });

    it('deve resetar contador após período de bloqueio', async () => {
      // Arrange
      const ipAddress = '192.168.1.1';
      const rateLimitKey = `login_attempts:${ipAddress}`;

      // Act
      await authService.resetLoginAttempts(ipAddress);

      // Assert
      expect(cacheManager.del).toHaveBeenCalledWith(rateLimitKey);
    });
  });
});