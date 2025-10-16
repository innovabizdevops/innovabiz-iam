// Integration Tests for WebAuthn API
// Testes de integração end-to-end seguindo padrões INNOVABIZ

import request from 'supertest';
import { app } from '../../src/app';
import { Pool } from 'pg';
import Redis from 'ioredis';

describe('WebAuthn Integration Tests', () => {
  let dbPool: Pool;
  let redisClient: Redis;
  let testTenantId: string;
  let testUserId: string;
  let authToken: string;

  beforeAll(async () => {
    // Setup test database connection
    dbPool = new Pool({
      connectionString: process.env.TEST_DATABASE_URL,
      max: 5
    });

    // Setup test Redis connection
    redisClient = new Redis(process.env.TEST_REDIS_URL);

    // Create test tenant and user
    testTenantId = 'test-tenant-integration';
    testUserId = 'test-user-integration';
    
    // Setup test data
    await setupTestData();
    
    // Get auth token for API calls
    authToken = await getTestAuthToken();
  });

  afterAll(async () => {
    // Cleanup test data
    await cleanupTestData();
    
    // Close connections
    await dbPool.end();
    await redisClient.quit();
  });

  describe('POST /api/v1/webauthn/registration/options', () => {
    it('should generate registration options successfully', async () => {
      const response = await request(app)
        .post('/api/v1/webauthn/registration/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          username: 'test@innovabiz.com',
          displayName: 'Test User Integration'
        })
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: {
          rp: {
            name: 'INNOVABIZ',
            id: 'innovabiz.com'
          },
          user: {
            name: 'test@innovabiz.com',
            displayName: 'Test User Integration'
          },
          challenge: expect.any(String),
          pubKeyCredParams: expect.arrayContaining([
            { alg: -7, type: 'public-key' },
            { alg: -257, type: 'public-key' }
          ]),
          authenticatorSelection: {
            authenticatorAttachment: 'platform',
            userVerification: 'required',
            residentKey: 'preferred'
          },
          attestation: 'direct'
        },
        correlationId: expect.any(String)
      });

      // Verify challenge is stored in Redis
      const storedChallenge = await redisClient.get(
        `webauthn:challenge:${testTenantId}:${testUserId}`
      );
      expect(storedChallenge).toBeTruthy();
    });

    it('should return 400 for missing required fields', async () => {
      const response = await request(app)
        .post('/api/v1/webauthn/registration/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId
          // Missing username and displayName
        })
        .expect(400);

      expect(response.body).toMatchObject({
        success: false,
        error: {
          code: 'VALIDATION_ERROR',
          message: expect.stringContaining('username')
        }
      });
    });

    it('should return 401 for invalid authentication', async () => {
      await request(app)
        .post('/api/v1/webauthn/registration/options')
        .set('Authorization', 'Bearer invalid-token')
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          username: 'test@innovabiz.com',
          displayName: 'Test User'
        })
        .expect(401);
    });

    it('should return 403 for missing tenant ID', async () => {
      await request(app)
        .post('/api/v1/webauthn/registration/options')
        .set('Authorization', `Bearer ${authToken}`)
        // Missing X-Tenant-ID header
        .send({
          userId: testUserId,
          username: 'test@innovabiz.com',
          displayName: 'Test User'
        })
        .expect(403);
    });
  });

  describe('POST /api/v1/webauthn/registration/verify', () => {
    let registrationChallenge: string;

    beforeEach(async () => {
      // Generate registration options first
      const optionsResponse = await request(app)
        .post('/api/v1/webauthn/registration/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          username: 'test@innovabiz.com',
          displayName: 'Test User Integration'
        });

      registrationChallenge = optionsResponse.body.data.challenge;
    });

    it('should verify valid registration response', async () => {
      // Mock valid registration response
      const mockRegistrationResponse = {
        id: 'mock-credential-id-123',
        rawId: 'mock-credential-id-123',
        response: {
          clientDataJSON: createMockClientDataJSON(registrationChallenge, 'webauthn.create'),
          attestationObject: createMockAttestationObject()
        },
        type: 'public-key'
      };

      const response = await request(app)
        .post('/api/v1/webauthn/registration/verify')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          credential: mockRegistrationResponse,
          challenge: registrationChallenge,
          origin: 'https://app.innovabiz.com'
        })
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: {
          verified: true,
          registrationInfo: expect.objectContaining({
            credentialID: expect.any(String),
            credentialPublicKey: expect.any(String),
            counter: expect.any(Number)
          })
        }
      });

      // Verify credential is stored in database
      const storedCredential = await dbPool.query(
        'SELECT * FROM webauthn_credentials WHERE credential_id = $1 AND tenant_id = $2',
        ['mock-credential-id-123', testTenantId]
      );
      expect(storedCredential.rows).toHaveLength(1);
    });

    it('should reject registration with invalid challenge', async () => {
      const mockRegistrationResponse = {
        id: 'mock-credential-id-456',
        rawId: 'mock-credential-id-456',
        response: {
          clientDataJSON: createMockClientDataJSON('invalid-challenge', 'webauthn.create'),
          attestationObject: createMockAttestationObject()
        },
        type: 'public-key'
      };

      const response = await request(app)
        .post('/api/v1/webauthn/registration/verify')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          credential: mockRegistrationResponse,
          challenge: registrationChallenge,
          origin: 'https://app.innovabiz.com'
        })
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: {
          verified: false
        }
      });
    });
  });

  describe('POST /api/v1/webauthn/authentication/options', () => {
    beforeEach(async () => {
      // Ensure user has at least one credential
      await createTestCredential();
    });

    it('should generate authentication options successfully', async () => {
      const response = await request(app)
        .post('/api/v1/webauthn/authentication/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId
        })
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: {
          challenge: expect.any(String),
          allowCredentials: expect.arrayContaining([
            expect.objectContaining({
              id: expect.any(String),
              type: 'public-key',
              transports: expect.any(Array)
            })
          ]),
          userVerification: 'required',
          rpId: 'innovabiz.com'
        }
      });
    });

    it('should return 404 for user with no credentials', async () => {
      // Create user without credentials
      const userWithoutCreds = 'user-no-creds';
      
      const response = await request(app)
        .post('/api/v1/webauthn/authentication/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: userWithoutCreds
        })
        .expect(404);

      expect(response.body).toMatchObject({
        success: false,
        error: {
          code: 'NO_CREDENTIALS_FOUND'
        }
      });
    });
  });

  describe('POST /api/v1/webauthn/authentication/verify', () => {
    let authenticationChallenge: string;
    let testCredentialId: string;

    beforeEach(async () => {
      // Create test credential and get authentication options
      testCredentialId = await createTestCredential();
      
      const optionsResponse = await request(app)
        .post('/api/v1/webauthn/authentication/options')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId
        });

      authenticationChallenge = optionsResponse.body.data.challenge;
    });

    it('should verify valid authentication response', async () => {
      const mockAuthResponse = {
        id: testCredentialId,
        rawId: testCredentialId,
        response: {
          clientDataJSON: createMockClientDataJSON(authenticationChallenge, 'webauthn.get'),
          authenticatorData: createMockAuthenticatorData(),
          signature: createMockSignature()
        },
        type: 'public-key'
      };

      const response = await request(app)
        .post('/api/v1/webauthn/authentication/verify')
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .send({
          userId: testUserId,
          credential: mockAuthResponse,
          challenge: authenticationChallenge,
          origin: 'https://app.innovabiz.com'
        })
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: {
          verified: true,
          authenticationInfo: expect.objectContaining({
            newCounter: expect.any(Number)
          })
        }
      });

      // Verify credential counter was updated
      const updatedCredential = await dbPool.query(
        'SELECT sign_count FROM webauthn_credentials WHERE credential_id = $1 AND tenant_id = $2',
        [testCredentialId, testTenantId]
      );
      expect(updatedCredential.rows[0].sign_count).toBeGreaterThan(0);
    });
  });

  describe('GET /api/v1/webauthn/credentials/:userId', () => {
    beforeEach(async () => {
      await createTestCredential();
    });

    it('should list user credentials successfully', async () => {
      const response = await request(app)
        .get(`/api/v1/webauthn/credentials/${testUserId}`)
        .set('Authorization', `Bearer ${authToken}`)
        .set('X-Tenant-ID', testTenantId)
        .expect(200);

      expect(response.body).toMatchObject({
        success: true,
        data: expect.arrayContaining([
          expect.objectContaining({
            credentialId: expect.any(String),
            friendlyName: expect.any(String),
            deviceType: expect.any(String),
            createdAt: expect.any(String),
            lastUsedAt: expect.any(String)
          })
        ])
      });
    });
  });

  describe('Rate Limiting', () => {
    it('should enforce rate limits on registration options', async () => {
      const requests = Array(15).fill(null).map(() =>
        request(app)
          .post('/api/v1/webauthn/registration/options')
          .set('Authorization', `Bearer ${authToken}`)
          .set('X-Tenant-ID', testTenantId)
          .send({
            userId: testUserId,
            username: 'test@innovabiz.com',
            displayName: 'Test User'
          })
      );

      const responses = await Promise.all(requests);
      
      // Should have some 429 responses due to rate limiting
      const rateLimitedResponses = responses.filter(r => r.status === 429);
      expect(rateLimitedResponses.length).toBeGreaterThan(0);
    });
  });

  describe('Metrics and Monitoring', () => {
    it('should expose Prometheus metrics', async () => {
      const response = await request(app)
        .get('/metrics')
        .expect(200);

      expect(response.text).toContain('webauthn_registration_total');
      expect(response.text).toContain('webauthn_authentication_total');
      expect(response.text).toContain('webauthn_registration_duration_seconds');
    });
  });

  // Helper functions
  async function setupTestData(): Promise<void> {
    // Create test tenant
    await dbPool.query(
      'INSERT INTO tenants (id, name, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING',
      [testTenantId, 'Test Tenant Integration']
    );

    // Create test user
    await dbPool.query(
      'INSERT INTO users (id, tenant_id, email, created_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING',
      [testUserId, testTenantId, 'test@innovabiz.com']
    );
  }

  async function cleanupTestData(): Promise<void> {
    // Clean up in reverse order of dependencies
    await dbPool.query('DELETE FROM webauthn_credentials WHERE tenant_id = $1', [testTenantId]);
    await dbPool.query('DELETE FROM audit_events WHERE tenant_id = $1', [testTenantId]);
    await dbPool.query('DELETE FROM users WHERE tenant_id = $1', [testTenantId]);
    await dbPool.query('DELETE FROM tenants WHERE id = $1', [testTenantId]);
    
    // Clean up Redis
    const keys = await redisClient.keys(`webauthn:*:${testTenantId}:*`);
    if (keys.length > 0) {
      await redisClient.del(...keys);
    }
  }

  async function getTestAuthToken(): Promise<string> {
    // Mock JWT token for testing
    return 'mock-jwt-token-for-integration-tests';
  }

  async function createTestCredential(): Promise<string> {
    const credentialId = `test-credential-${Date.now()}`;
    
    await dbPool.query(`
      INSERT INTO webauthn_credentials (
        credential_id, user_id, tenant_id, public_key, 
        sign_count, device_type, friendly_name, created_at
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
    `, [
      credentialId,
      testUserId,
      testTenantId,
      Buffer.from('mock-public-key'),
      0,
      'platform',
      'Test Credential'
    ]);

    return credentialId;
  }

  function createMockClientDataJSON(challenge: string, type: string): string {
    const clientData = {
      type,
      challenge,
      origin: 'https://app.innovabiz.com'
    };
    return Buffer.from(JSON.stringify(clientData)).toString('base64');
  }

  function createMockAttestationObject(): string {
    // Simplified mock attestation object
    return 'mock-attestation-object-base64';
  }

  function createMockAuthenticatorData(): string {
    // Simplified mock authenticator data
    return 'mock-authenticator-data-base64';
  }

  function createMockSignature(): string {
    // Simplified mock signature
    return 'mock-signature-base64';
  }
});