/**
 * @file authentication-security.test.ts
 * @description Testes de segurança para o módulo de autenticação do IAM INNOVABIZ
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { Test, TestingModule } from '@nestjs/testing';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { HttpModule } from '@nestjs/axios';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../../src/app.module';
import { AuthModule } from '../../src/modules/auth/auth.module';
import { UsersModule } from '../../src/modules/users/users.module';
import { TenantModule } from '../../src/modules/tenant/tenant.module';
import { JwtModule } from '@nestjs/jwt';

/**
 * Testes de segurança para verificar vulnerabilidades comuns
 * em fluxos de autenticação e autorização
 */
describe('Testes de Segurança de Autenticação', () => {
  let app: INestApplication;
  let configService: ConfigService;

  beforeAll(async () => {
    // Configuração do módulo de teste
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [
        ConfigModule.forRoot({
          isGlobal: true,
          envFilePath: ['./tests/.env.test'],
        }),
        HttpModule,
        JwtModule.registerAsync({
          imports: [ConfigModule],
          useFactory: async (configService: ConfigService) => ({
            secret: configService.get<string>('JWT_SECRET'),
            signOptions: {
              expiresIn: configService.get<string>('JWT_EXPIRATION', '1h'),
            },
          }),
          inject: [ConfigService],
        }),
        AppModule,
        AuthModule,
        UsersModule,
        TenantModule,
      ],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();
    configService = moduleFixture.get<ConfigService>(ConfigService);
  });

  afterAll(async () => {
    await app.close();
  });

  describe('Proteção contra ataques de força bruta', () => {
    it('deve bloquear temporariamente após múltiplas tentativas de login inválidas', async () => {
      const invalidCredentials = {
        username: 'usuario@test.com',
        password: 'SenhaIncorreta123!',
        tenantId: 'test-tenant'
      };

      // Realiza 5 tentativas de login com credenciais inválidas
      for (let i = 0; i < 5; i++) {
        await request(app.getHttpServer())
          .post('/api/v1/auth/login')
          .send(invalidCredentials)
          .expect(401);
      }

      // Na 6ª tentativa, deve ser bloqueado temporariamente
      const response = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send(invalidCredentials)
        .expect(429);

      expect(response.body.message).toContain('muitas tentativas');
      expect(response.headers['retry-after']).toBeDefined();
    });
  });

  describe('Segurança de tokens JWT', () => {
    it('deve invalidar tokens com assinatura inválida', async () => {
      // Token JWT com assinatura manipulada
      const invalidToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';

      // Tenta acessar um endpoint protegido
      await request(app.getHttpServer())
        .get('/api/v1/users/me')
        .set('Authorization', `Bearer ${invalidToken}`)
        .expect(401);
    });

    it('deve rejeitar tokens expirados', async () => {
      // Token JWT expirado (expirado em 1 de janeiro de 2020)
      const expiredToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsInRlbmFudElkIjoidGVzdC10ZW5hbnQiLCJpYXQiOjE1Nzc4MzY4MDAsImV4cCI6MTU3NzgzNzEwMH0.3JCBFgpJ-f6PgTsdD2hKvGqUdQPXB-fPQWEpZ_RWxEE';

      // Tenta acessar um endpoint protegido
      await request(app.getHttpServer())
        .get('/api/v1/users/me')
        .set('Authorization', `Bearer ${expiredToken}`)
        .expect(401);
    });
  });

  describe('Proteção contra ataques de cross-site request forgery (CSRF)', () => {
    it('deve exigir token CSRF para operações sensíveis', async () => {
      // Primeiro faz login para obter cookie de sessão
      const loginResponse = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send({
          username: 'admin@test.com',
          password: 'Admin123!',
          tenantId: 'test-tenant'
        });

      const cookies = loginResponse.headers['set-cookie'];

      // Tenta uma operação sensível sem o token CSRF
      await request(app.getHttpServer())
        .post('/api/v1/users/change-password')
        .set('Cookie', cookies)
        .send({
          currentPassword: 'Admin123!',
          newPassword: 'NovoAdmin123!'
        })
        .expect(403); // Deve falhar com 403 Forbidden por falta do token CSRF
    });
  });

  describe('Proteção contra ataques de injeção', () => {
    it('deve validar e sanitizar inputs para prevenir injeções NoSQL', async () => {
      // Tenta login com injeção NoSQL básica
      const maliciousPayload = {
        username: { $ne: null }, // Tenta NoSQL injection para fazer match em qualquer username
        password: { $ne: null }, // Tenta NoSQL injection para fazer match em qualquer senha
        tenantId: 'test-tenant'
      };

      const response = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send(maliciousPayload)
        .expect(400); // Deve rejeitar com bad request por validação

      expect(response.body.message).toContain('validação');
    });

    it('deve rejeitar tentativas de SQL injection em parâmetros', async () => {
      // Tenta SQL injection via parâmetros de query
      const response = await request(app.getHttpServer())
        .get('/api/v1/users/search')
        .query({ query: "'; DROP TABLE users; --" })
        .expect(400); // Deve rejeitar com bad request por validação

      expect(response.body.message).toContain('validação');
    });
  });

  describe('Proteção de headers HTTP', () => {
    it('deve incluir headers de segurança nas respostas', async () => {
      const response = await request(app.getHttpServer())
        .get('/api/v1/health')
        .expect(200);

      // Verificar headers de segurança importantes
      expect(response.headers['x-content-type-options']).toBe('nosniff');
      expect(response.headers['x-xss-protection']).toBe('1; mode=block');
      expect(response.headers['x-frame-options']).toBe('DENY');
      expect(response.headers['content-security-policy']).toBeDefined();
      expect(response.headers['strict-transport-security']).toBeDefined();
    });
  });

  describe('Validação de privilégios', () => {
    let adminToken: string;
    let regularUserToken: string;

    beforeAll(async () => {
      // Obter token de admin
      const adminResponse = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send({
          username: 'admin@test.com',
          password: 'Admin123!',
          tenantId: 'test-tenant'
        });
      adminToken = adminResponse.body.access_token;

      // Obter token de usuário comum
      const userResponse = await request(app.getHttpServer())
        .post('/api/v1/auth/login')
        .send({
          username: 'user@test.com',
          password: 'User123!',
          tenantId: 'test-tenant'
        });
      regularUserToken = userResponse.body.access_token;
    });

    it('deve impedir que usuários comuns acessem endpoints administrativos', async () => {
      // Tenta acessar endpoint administrativo com token de usuário comum
      await request(app.getHttpServer())
        .get('/api/v1/admin/users')
        .set('Authorization', `Bearer ${regularUserToken}`)
        .expect(403);
    });

    it('deve permitir que admins acessem endpoints administrativos', async () => {
      // Acessa endpoint administrativo com token de admin
      await request(app.getHttpServer())
        .get('/api/v1/admin/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .expect(200);
    });

    it('deve implementar corretamente segregação por tenant', async () => {
      // Tenta acessar dados de outro tenant
      await request(app.getHttpServer())
        .get('/api/v1/users/tenant/other-tenant')
        .set('Authorization', `Bearer ${adminToken}`) // Mesmo sendo admin
        .expect(403); // Não deve permitir acesso cross-tenant
    });
  });
});