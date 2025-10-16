/**
 * @file tenant.service.test.ts
 * @description Testes unitários para o serviço de tenant do IAM INNOVABIZ
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { TenantService } from '../../../src/services/tenant/tenant.service';
import { TenantRepository } from '../../../src/repositories/tenant.repository';
import { CacheService } from '../../../src/services/cache/cache.service';
import { TenantNotFoundException } from '../../../src/exceptions/tenant-not-found.exception';

// Mock para o repositório de tenant
const mockTenantRepository = () => ({
  findById: jest.fn(),
  findByDomain: jest.fn(),
  create: jest.fn(),
  update: jest.fn(),
  delete: jest.fn(),
  findAll: jest.fn(),
});

// Mock para o serviço de cache
const mockCacheService = () => ({
  get: jest.fn(),
  set: jest.fn(),
  del: jest.fn(),
  reset: jest.fn(),
});

describe('TenantService', () => {
  let tenantService: TenantService;
  let tenantRepository;
  let cacheService;

  // Dados de teste
  const testTenant = {
    id: 'tenant123',
    name: 'Test Tenant',
    domain: 'test.example.com',
    status: 'active',
    createdAt: new Date(),
    updatedAt: new Date(),
    config: {
      webauthn: {
        rpId: 'example.com',
        rpName: 'INNOVABIZ IAM',
        origin: 'https://test.example.com'
      },
      passwordPolicy: {
        minLength: 8,
        requireNumbers: true,
        requireLowercase: true,
        requireUppercase: true,
        requireSpecialChars: true,
        expirationDays: 90
      },
      mfa: {
        enabled: true,
        methods: ['webauthn', 'totp']
      }
    }
  };

  beforeEach(async () => {
    // Configuração do módulo de teste
    const module: TestingModule = await Test.createTestingModule({
      imports: [
        ConfigModule.forRoot({
          isGlobal: true,
        }),
      ],
      providers: [
        TenantService,
        {
          provide: TenantRepository,
          useFactory: mockTenantRepository,
        },
        {
          provide: CacheService,
          useFactory: mockCacheService,
        },
        ConfigService,
      ],
    }).compile();

    tenantService = module.get<TenantService>(TenantService);
    tenantRepository = module.get<TenantRepository>(TenantRepository);
    cacheService = module.get<CacheService>(CacheService);
  });

  describe('getTenantById', () => {
    it('deve retornar um tenant quando ele existir no cache', async () => {
      // Configurar mock do cache para retornar o tenant
      cacheService.get.mockResolvedValue(testTenant);

      // Executar o método
      const result = await tenantService.getTenantById('tenant123');

      // Verificar o resultado
      expect(result).toEqual(testTenant);
      expect(cacheService.get).toHaveBeenCalledWith('tenant:tenant123');
      expect(tenantRepository.findById).not.toHaveBeenCalled();
    });

    it('deve buscar e retornar um tenant do repositório quando não estiver no cache', async () => {
      // Configurar mocks
      cacheService.get.mockResolvedValue(null);
      tenantRepository.findById.mockResolvedValue(testTenant);

      // Executar o método
      const result = await tenantService.getTenantById('tenant123');

      // Verificar o resultado
      expect(result).toEqual(testTenant);
      expect(cacheService.get).toHaveBeenCalledWith('tenant:tenant123');
      expect(tenantRepository.findById).toHaveBeenCalledWith('tenant123');
      expect(cacheService.set).toHaveBeenCalledWith('tenant:tenant123', testTenant, expect.any(Number));
    });

    it('deve lançar TenantNotFoundException quando o tenant não existir', async () => {
      // Configurar mocks
      cacheService.get.mockResolvedValue(null);
      tenantRepository.findById.mockResolvedValue(null);

      // Executar o método e verificar a exceção
      await expect(tenantService.getTenantById('nonexistent')).rejects.toThrow(TenantNotFoundException);
      expect(cacheService.get).toHaveBeenCalledWith('tenant:nonexistent');
      expect(tenantRepository.findById).toHaveBeenCalledWith('nonexistent');
    });
  });
});