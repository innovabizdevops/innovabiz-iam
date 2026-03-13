/**
 * @module identity-management/infrastructure/cache
 * @description IAM Cache Service — Redis-backed session/token/rate-limit cache
 *
 * Architecture: Infrastructure Adapter (cache layer)
 * Technology: ioredis via @nestjs/cache-manager + cache-manager-ioredis-yet
 * Standards: NIST 800-63-3 (session binding), OWASP (token storage), ISO 27001
 * Cognitive: RTM (Real-Time Mesh < 2ms), EDGE3 (Edge Cognitive Mesh)
 *
 * Key Namespaces:
 *   iam:session:{tenantId}:{sessionId}   — Active session data (TTL: 24h)
 *   iam:token:{tenantId}:{tokenHash}     — Token blocklist/allowlist (TTL: configurable)
 *   iam:rate:{tenantId}:{userId}:{action} — Rate limiting counters (TTL: 1min)
 *   iam:mfa:{tenantId}:{userId}          — MFA challenge state (TTL: 5min)
 *   iam:user:{tenantId}:{userId}         — User profile cache (TTL: 15min)
 */

import { Injectable, Logger, Inject, OnModuleInit } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';

// ─── TTL Constants (seconds) ───
const TTL = {
    SESSION: 86400,        // 24 hours
    TOKEN_BLOCKLIST: 3600, // 1 hour
    RATE_LIMIT: 60,        // 1 minute window
    MFA_CHALLENGE: 300,    // 5 minutes
    USER_PROFILE: 900,     // 15 minutes
    ROLE_CACHE: 1800,      // 30 minutes
    DASHBOARD: 30,         // 30 seconds (real-time feel)
} as const;

// ─── Key Builders ───
const key = {
    session: (tenantId: string, sessionId: string) => `iam:session:${tenantId}:${sessionId}`,
    userSessions: (tenantId: string, userId: string) => `iam:sessions:${tenantId}:${userId}`,
    token: (tenantId: string, tokenHash: string) => `iam:token:${tenantId}:${tokenHash}`,
    rate: (tenantId: string, userId: string, action: string) => `iam:rate:${tenantId}:${userId}:${action}`,
    mfa: (tenantId: string, userId: string) => `iam:mfa:${tenantId}:${userId}`,
    user: (tenantId: string, userId: string) => `iam:user:${tenantId}:${userId}`,
    roles: (tenantId: string) => `iam:roles:${tenantId}`,
    dashboard: (tenantId: string) => `iam:dashboard:${tenantId}`,
};

export interface CachedSession {
    sessionId: string;
    userId: string;
    trustScore: number;
    trustLevel: string;
    mfaLevel: string;
    deviceFingerprint?: string;
    ipAddress?: string;
    startedAt: string;
}

export interface MfaChallengeState {
    userId: string;
    method: string;       // FIDO2, TOTP, PUSH, SMS
    challengeId: string;
    createdAt: string;
    expiresAt: string;
    verified: boolean;
}

export interface RateLimitResult {
    allowed: boolean;
    remaining: number;
    resetAt: number;       // Unix timestamp
}

@Injectable()
export class IamCacheService implements OnModuleInit {
    private readonly logger = new Logger(IamCacheService.name);

    constructor(@Inject(CACHE_MANAGER) private readonly cache: Cache) { }

    async onModuleInit() {
        this.logger.log('🔴 IAM Cache Service initialized (Redis-backed)');
    }

    // ═══════════════════════════════════════════════════════════
    // Session Cache
    // ═══════════════════════════════════════════════════════════

    async cacheSession(tenantId: string, session: CachedSession): Promise<void> {
        await this.cache.set(
            key.session(tenantId, session.sessionId),
            JSON.stringify(session),
            TTL.SESSION * 1000,
        );
        // Also track user's active sessions
        const userKey = key.userSessions(tenantId, session.userId);
        const existing = await this.cache.get<string>(userKey);
        const sessionIds: string[] = existing ? JSON.parse(existing) : [];
        if (!sessionIds.includes(session.sessionId)) {
            sessionIds.push(session.sessionId);
            await this.cache.set(userKey, JSON.stringify(sessionIds), TTL.SESSION * 1000);
        }
        this.logger.debug(`[Cache] Session ${session.sessionId} cached — tenant: ${tenantId}`);
    }

    async getSession(tenantId: string, sessionId: string): Promise<CachedSession | null> {
        const raw = await this.cache.get<string>(key.session(tenantId, sessionId));
        return raw ? JSON.parse(raw) : null;
    }

    async invalidateSession(tenantId: string, sessionId: string): Promise<void> {
        await this.cache.del(key.session(tenantId, sessionId));
        this.logger.log(`[Cache] Session ${sessionId} invalidated — tenant: ${tenantId}`);
    }

    async getUserSessionIds(tenantId: string, userId: string): Promise<string[]> {
        const raw = await this.cache.get<string>(key.userSessions(tenantId, userId));
        return raw ? JSON.parse(raw) : [];
    }

    async invalidateAllUserSessions(tenantId: string, userId: string): Promise<number> {
        const sessionIds = await this.getUserSessionIds(tenantId, userId);
        for (const sid of sessionIds) {
            await this.cache.del(key.session(tenantId, sid));
        }
        await this.cache.del(key.userSessions(tenantId, userId));
        this.logger.log(`[Cache] All sessions invalidated for user ${userId} — count: ${sessionIds.length}`);
        return sessionIds.length;
    }

    // ═══════════════════════════════════════════════════════════
    // Token Blocklist
    // ═══════════════════════════════════════════════════════════

    async blockToken(tenantId: string, tokenHash: string, ttlSeconds?: number): Promise<void> {
        await this.cache.set(
            key.token(tenantId, tokenHash),
            'blocked',
            (ttlSeconds ?? TTL.TOKEN_BLOCKLIST) * 1000,
        );
        this.logger.log(`[Cache] Token blocked — tenant: ${tenantId}`);
    }

    async isTokenBlocked(tenantId: string, tokenHash: string): Promise<boolean> {
        const val = await this.cache.get<string>(key.token(tenantId, tokenHash));
        return val === 'blocked';
    }

    // ═══════════════════════════════════════════════════════════
    // Rate Limiting
    // ═══════════════════════════════════════════════════════════

    async checkRateLimit(
        tenantId: string,
        userId: string,
        action: string,
        maxPerMinute: number = 60,
    ): Promise<RateLimitResult> {
        const k = key.rate(tenantId, userId, action);
        const raw = await this.cache.get<string>(k);
        const current = raw ? parseInt(raw, 10) : 0;

        if (current >= maxPerMinute) {
            return { allowed: false, remaining: 0, resetAt: Date.now() + TTL.RATE_LIMIT * 1000 };
        }

        await this.cache.set(k, String(current + 1), TTL.RATE_LIMIT * 1000);
        return {
            allowed: true,
            remaining: maxPerMinute - current - 1,
            resetAt: Date.now() + TTL.RATE_LIMIT * 1000,
        };
    }

    // ═══════════════════════════════════════════════════════════
    // MFA Challenge State
    // ═══════════════════════════════════════════════════════════

    async storeMfaChallenge(tenantId: string, state: MfaChallengeState): Promise<void> {
        await this.cache.set(
            key.mfa(tenantId, state.userId),
            JSON.stringify(state),
            TTL.MFA_CHALLENGE * 1000,
        );
    }

    async getMfaChallenge(tenantId: string, userId: string): Promise<MfaChallengeState | null> {
        const raw = await this.cache.get<string>(key.mfa(tenantId, userId));
        return raw ? JSON.parse(raw) : null;
    }

    async clearMfaChallenge(tenantId: string, userId: string): Promise<void> {
        await this.cache.del(key.mfa(tenantId, userId));
    }

    // ═══════════════════════════════════════════════════════════
    // User Profile Cache
    // ═══════════════════════════════════════════════════════════

    async cacheUserProfile(tenantId: string, userId: string, profile: Record<string, unknown>): Promise<void> {
        await this.cache.set(key.user(tenantId, userId), JSON.stringify(profile), TTL.USER_PROFILE * 1000);
    }

    async getCachedUserProfile(tenantId: string, userId: string): Promise<Record<string, unknown> | null> {
        const raw = await this.cache.get<string>(key.user(tenantId, userId));
        return raw ? JSON.parse(raw) : null;
    }

    async invalidateUserProfile(tenantId: string, userId: string): Promise<void> {
        await this.cache.del(key.user(tenantId, userId));
    }

    // ═══════════════════════════════════════════════════════════
    // Dashboard KPI Cache (fast aggregation)
    // ═══════════════════════════════════════════════════════════

    async cacheDashboardKpis(tenantId: string, kpis: Record<string, unknown>): Promise<void> {
        await this.cache.set(key.dashboard(tenantId), JSON.stringify(kpis), TTL.DASHBOARD * 1000);
    }

    async getCachedDashboardKpis(tenantId: string): Promise<Record<string, unknown> | null> {
        const raw = await this.cache.get<string>(key.dashboard(tenantId));
        return raw ? JSON.parse(raw) : null;
    }

    // ═══════════════════════════════════════════════════════════
    // Roles Cache (reduces DB queries)
    // ═══════════════════════════════════════════════════════════

    async cacheRoles(tenantId: string, roles: unknown[]): Promise<void> {
        await this.cache.set(key.roles(tenantId), JSON.stringify(roles), TTL.ROLE_CACHE * 1000);
    }

    async getCachedRoles(tenantId: string): Promise<unknown[] | null> {
        const raw = await this.cache.get<string>(key.roles(tenantId));
        return raw ? JSON.parse(raw) : null;
    }

    async invalidateRoles(tenantId: string): Promise<void> {
        await this.cache.del(key.roles(tenantId));
    }
}
