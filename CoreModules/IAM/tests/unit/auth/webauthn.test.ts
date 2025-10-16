/**
 * @file webauthn.test.ts
 * @description Testes unitários para autenticação WebAuthn/FIDO2
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { WebAuthnService } from '../../../src/services/auth/webauthn.service';
import { WebAuthnRepository } from '../../../src/repositories/webauthn.repository';
import { TenantService } from '../../../src/services/tenant/tenant.service';
import { ConfigService } from '@nestjs/config';

// Mocks
jest.mock('../../../src/repositories/webauthn.repository');
jest.mock('../../../src/services/tenant/tenant.service');

describe('WebAuthnService', () => {
  let webAuthnService: WebAuthnService;
  let webAuthnRepository: jest.Mocked<WebAuthnRepository>;
  let tenantService: jest.Mocked<TenantService>;
  let configService: ConfigService;

  // Dados de teste
  const testUser = {
    id: 'user123',
    username: 'testuser@example.com',
    tenantId: 'tenant123',
    displayName: 'Test User'
  };

  const mockCredential = {
    id: 'credential123',
    userId: testUser.id,
    tenantId: testUser.tenantId,
    publicKey: 'mockPublicKey',
    counter: 0,
    credentialId: 'mockCredentialId',
    createdAt: new Date(),
    updatedAt: new Date(),
    deviceType: 'platform',
    transports: ['internal']
  };

  beforeEach(() => {
    // Configuração de mocks
    webAuthnRepository = {
      findCredentialsByUserId: jest.fn(),
      saveCredential: jest.fn(),
      updateCredentialCounter: jest.fn(),
      findCredentialById: jest.fn(),
      deleteCredential: jest.fn()
    } as unknown as jest.Mocked<WebAuthnRepository>;

    tenantService = {
      getTenantById: jest.fn(),
      getTenantConfig: jest.fn()
    } as unknown as jest.Mocked<TenantService>;

    configService = new ConfigService();
    
    // Configurações de mock para tenant
    tenantService.getTenantById.mockResolvedValue({
      id: 'tenant123',
      name: 'Test Tenant',
      domain: 'test.example.com',
      config: {
        webauthn: {
          rpId: 'example.com',
          rpName: 'INNOVABIZ IAM',
          rpIcon: 'https://example.com/logo.png',
          origin: 'https://test.example.com',
          timeout: 60000,
          attestation: 'direct',
          authenticatorAttachment: 'platform',
          userVerification: 'preferred',
          extensions: {}
        }
      }
    });

    tenantService.getTenantConfig.mockImplementation((tenantId, key) => {
      if (tenantId === 'tenant123' && key === 'webauthn') {
        return {
          rpId: 'example.com',
          rpName: 'INNOVABIZ IAM',
          rpIcon: 'https://example.com/logo.png',
          origin: 'https://test.example.com',
          timeout: 60000,
          attestation: 'direct',
          authenticatorAttachment: 'platform',
          userVerification: 'preferred',
          extensions: {}
        };
      }
      return null;
    });

    // Configura o serviço de WebAuthn
    webAuthnService = new WebAuthnService(
      webAuthnRepository,
      tenantService,
      configService
    );
  });

  describe('generateRegistrationOptions', () => {
    it('deve gerar opções de registro válidas para um novo usuário', async () => {
      // Configura o mock para simular usuário sem credenciais
      webAuthnRepository.findCredentialsByUserId.mockResolvedValue([]);

      // Executa o método
      const registrationOptions = await webAuthnService.generateRegistrationOptions(testUser);

      // Verifica as opções geradas
      expect(registrationOptions).toBeDefined();
      expect(registrationOptions.user).toBeDefined();
      expect(registrationOptions.user.id).toBeDefined();
      expect(registrationOptions.user.name).toBe(testUser.username);
      expect(registrationOptions.user.displayName).toBe(testUser.displayName);
      expect(registrationOptions.rp).toBeDefined();
      expect(registrationOptions.rp.id).toBe('example.com');
      expect(registrationOptions.rp.name).toBe('INNOVABIZ IAM');
      expect(registrationOptions.authenticatorSelection).toBeDefined();
      expect(registrationOptions.authenticatorSelection.userVerification).toBe('preferred');
      expect(registrationOptions.authenticatorSelection.authenticatorAttachment).toBe('platform');
    });

    it('deve incluir credenciais existentes como excludeCredentials', async () => {
      // Simula um usuário que já possui credenciais
      const existingCredentials = [mockCredential];
      webAuthnRepository.findCredentialsByUserId.mockResolvedValue(existingCredentials);

      // Executa o método
      const registrationOptions = await webAuthnService.generateRegistrationOptions(testUser);

      // Verifica que as credenciais existentes foram incluídas para exclusão
      expect(registrationOptions.excludeCredentials).toBeDefined();
      expect(registrationOptions.excludeCredentials.length).toBe(1);
      expect(registrationOptions.excludeCredentials[0].id).toBeDefined();
      expect(registrationOptions.excludeCredentials[0].type).toBe('public-key');
    });

    it('deve gerar opções com configurações específicas do tenant', async () => {
      // Configura um tenant com configurações personalizadas
      tenantService.getTenantConfig.mockImplementation((tenantId, key) => {
        if (tenantId === 'tenant123' && key === 'webauthn') {
          return {
            rpId: 'custom.domain.com',
            rpName: 'Tenant Personalizado',
            rpIcon: 'https://custom.domain.com/logo.png',
            origin: 'https://custom.domain.com',
            timeout: 30000,
            attestation: 'indirect',
            authenticatorAttachment: 'cross-platform',
            userVerification: 'required',
            extensions: {
              credProps: true
            }
          };
        }
        return null;
      });

      webAuthnRepository.findCredentialsByUserId.mockResolvedValue([]);

      // Executa o método
      const registrationOptions = await webAuthnService.generateRegistrationOptions(testUser);

      // Verifica configurações personalizadas
      expect(registrationOptions.rp.id).toBe('custom.domain.com');
      expect(registrationOptions.rp.name).toBe('Tenant Personalizado');
      expect(registrationOptions.authenticatorSelection.userVerification).toBe('required');
      expect(registrationOptions.authenticatorSelection.authenticatorAttachment).toBe('cross-platform');
      expect(registrationOptions.attestation).toBe('indirect');
      expect(registrationOptions.timeout).toBe(30000);
      expect(registrationOptions.extensions).toEqual({ credProps: true });
    });
  });

  describe('verifyRegistrationResponse', () => {
    const mockRegistrationResponse = {
      id: 'responseId',
      rawId: new Uint8Array([1, 2, 3, 4]),
      response: {
        clientDataJSON: new Uint8Array([10, 20, 30]),
        attestationObject: new Uint8Array([40, 50, 60])
      },
      clientExtensionResults: {},
      type: 'public-key'
    };

    it('deve verificar e salvar uma resposta de registro válida', async () => {
      // Mock para verificação bem-sucedida
      const mockVerificationResult = {
        verified: true,
        registrationInfo: {
          credentialID: new Uint8Array([1, 2, 3, 4]),
          credentialPublicKey: new Uint8Array([5, 6, 7, 8]),
          counter: 0,
          credentialDeviceType: 'platform',
          credentialBackedUp: false,
          transports: ['internal']
        }
      };

      // Substitui a função global de verificação WebAuthn
      (global as any).verifyRegistrationResponse = jest.fn().mockResolvedValue(mockVerificationResult);

      // Mock para o repositório salvar a credencial
      webAuthnRepository.saveCredential.mockResolvedValue({
        id: 'newCredentialId',
        ...mockCredential
      });

      // Executa o método
      const result = await webAuthnService.verifyRegistrationResponse(
        mockRegistrationResponse as any,
        { challenge: 'challenge123', userId: testUser.id, tenantId: testUser.tenantId }
      );

      // Verifica o resultado
      expect(result.verified).toBe(true);
      expect(result.credentialId).toBe('newCredentialId');
      expect(webAuthnRepository.saveCredential).toHaveBeenCalledTimes(1);
    });

    it('deve rejeitar uma resposta de registro inválida', async () => {
      // Mock para verificação mal-sucedida
      const mockFailedVerification = {
        verified: false,
        registrationInfo: null
      };

      // Substitui a função global de verificação WebAuthn
      (global as any).verifyRegistrationResponse = jest.fn().mockResolvedValue(mockFailedVerification);

      // Executa o método e espera que falhe
      await expect(
        webAuthnService.verifyRegistrationResponse(
          mockRegistrationResponse as any,
          { challenge: 'challenge123', userId: testUser.id, tenantId: testUser.tenantId }
        )
      ).rejects.toThrow('Falha na verificação da resposta de registro WebAuthn');

      // Verifica que nenhuma credencial foi salva
      expect(webAuthnRepository.saveCredential).not.toHaveBeenCalled();
    });
  });
});