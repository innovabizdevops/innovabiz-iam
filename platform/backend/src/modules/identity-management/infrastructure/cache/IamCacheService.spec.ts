/**
 * @module identity-management/infrastructure/cache
 * @test IamCacheService — Unit Tests
 *
 * Tests use in-memory cache-manager (no Redis needed for testing)
 * Standards: NIST 800-63-3 (session binding), OWASP (token storage)
 */

import { Test, TestingModule } from '@nestjs/testing';
import { CacheModule } from '@nestjs/cache-manager';
import { IamCacheService, CachedSession, MfaChallengeState } from './IamCacheService';

describe('IamCacheService', () => {
    let service: IamCacheService;
    const TENANT = 'tenant-test-001';

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            imports: [CacheModule.register({ ttl: 60000, max: 100 })],
            providers: [IamCacheService],
        }).compile();

        service = module.get<IamCacheService>(IamCacheService);
        await service.onModuleInit();
    });

    // ─── Session Cache ───

    describe('Session Cache', () => {
        const session: CachedSession = {
            sessionId: 'sess-001',
            userId: 'user-001',
            trustScore: 95,
            trustLevel: 'HIGH',
            mfaLevel: 'FIDO2',
            ipAddress: '10.0.0.1',
            startedAt: new Date().toISOString(),
        };

        it('should cache and retrieve a session', async () => {
            await service.cacheSession(TENANT, session);
            const result = await service.getSession(TENANT, 'sess-001');

            expect(result).toBeDefined();
            expect(result!.sessionId).toBe('sess-001');
            expect(result!.userId).toBe('user-001');
            expect(result!.trustScore).toBe(95);
        });

        it('should return null for non-existent sessions', async () => {
            const result = await service.getSession(TENANT, 'non-existent');
            expect(result).toBeNull();
        });

        it('should invalidate a session', async () => {
            await service.cacheSession(TENANT, session);
            await service.invalidateSession(TENANT, 'sess-001');
            const result = await service.getSession(TENANT, 'sess-001');
            expect(result).toBeNull();
        });

        it('should track user session IDs', async () => {
            await service.cacheSession(TENANT, session);
            const ids = await service.getUserSessionIds(TENANT, 'user-001');
            expect(ids).toContain('sess-001');
        });

        it('should invalidate all user sessions', async () => {
            const s2: CachedSession = { ...session, sessionId: 'sess-002' };
            await service.cacheSession(TENANT, session);
            await service.cacheSession(TENANT, s2);

            const count = await service.invalidateAllUserSessions(TENANT, 'user-001');
            expect(count).toBe(2);

            const r1 = await service.getSession(TENANT, 'sess-001');
            const r2 = await service.getSession(TENANT, 'sess-002');
            expect(r1).toBeNull();
            expect(r2).toBeNull();
        });
    });

    // ─── Token Blocklist ───

    describe('Token Blocklist', () => {
        it('should block a token and check if blocked', async () => {
            await service.blockToken(TENANT, 'abc123hash');
            const blocked = await service.isTokenBlocked(TENANT, 'abc123hash');
            expect(blocked).toBe(true);
        });

        it('should return false for non-blocked tokens', async () => {
            const blocked = await service.isTokenBlocked(TENANT, 'not-blocked');
            expect(blocked).toBe(false);
        });
    });

    // ─── Rate Limiting ───

    describe('Rate Limiting', () => {
        it('should allow requests within the limit', async () => {
            const result = await service.checkRateLimit(TENANT, 'user-001', 'login', 5);
            expect(result.allowed).toBe(true);
            expect(result.remaining).toBe(4);
        });

        it('should block requests exceeding the limit', async () => {
            for (let i = 0; i < 5; i++) {
                await service.checkRateLimit(TENANT, 'user-limit', 'login', 5);
            }
            const result = await service.checkRateLimit(TENANT, 'user-limit', 'login', 5);
            expect(result.allowed).toBe(false);
            expect(result.remaining).toBe(0);
        });
    });

    // ─── MFA Challenge State ───

    describe('MFA Challenge', () => {
        const challenge: MfaChallengeState = {
            userId: 'user-001',
            method: 'FIDO2',
            challengeId: 'ch-001',
            createdAt: new Date().toISOString(),
            expiresAt: new Date(Date.now() + 300_000).toISOString(),
            verified: false,
        };

        it('should store and retrieve MFA challenge', async () => {
            await service.storeMfaChallenge(TENANT, challenge);
            const result = await service.getMfaChallenge(TENANT, 'user-001');

            expect(result).toBeDefined();
            expect(result!.method).toBe('FIDO2');
            expect(result!.verified).toBe(false);
        });

        it('should clear MFA challenge', async () => {
            await service.storeMfaChallenge(TENANT, challenge);
            await service.clearMfaChallenge(TENANT, 'user-001');
            const result = await service.getMfaChallenge(TENANT, 'user-001');
            expect(result).toBeNull();
        });
    });

    // ─── User Profile Cache ───

    describe('User Profile Cache', () => {
        it('should cache and retrieve user profile', async () => {
            const profile = { email: 'test@tenant.com', role: 'admin' };
            await service.cacheUserProfile(TENANT, 'user-001', profile);
            const result = await service.getCachedUserProfile(TENANT, 'user-001');

            expect(result).toBeDefined();
            expect(result!.email).toBe('test@tenant.com');
        });

        it('should invalidate user profile cache', async () => {
            await service.cacheUserProfile(TENANT, 'user-001', { email: 'x' });
            await service.invalidateUserProfile(TENANT, 'user-001');
            const result = await service.getCachedUserProfile(TENANT, 'user-001');
            expect(result).toBeNull();
        });
    });

    // ─── Dashboard KPI Cache ───

    describe('Dashboard Cache', () => {
        it('should cache and retrieve dashboard KPIs', async () => {
            const kpis = { totalUsers: 150, activeSessions: 42 };
            await service.cacheDashboardKpis(TENANT, kpis);
            const result = await service.getCachedDashboardKpis(TENANT);

            expect(result).toBeDefined();
            expect(result!.totalUsers).toBe(150);
        });
    });

    // ─── Roles Cache ───

    describe('Roles Cache', () => {
        it('should cache and retrieve roles', async () => {
            const roles = [{ id: '1', name: 'Admin' }, { id: '2', name: 'User' }];
            await service.cacheRoles(TENANT, roles);
            const result = await service.getCachedRoles(TENANT);

            expect(result).toHaveLength(2);
        });

        it('should invalidate roles cache', async () => {
            await service.cacheRoles(TENANT, [{ id: '1' }]);
            await service.invalidateRoles(TENANT);
            const result = await service.getCachedRoles(TENANT);
            expect(result).toBeNull();
        });
    });
});
