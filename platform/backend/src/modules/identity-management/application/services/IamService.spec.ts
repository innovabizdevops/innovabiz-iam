/**
 * @module identity-management/application/services
 * @test IamService — Unit Tests
 *
 * TDD: Test-first approach per InnovaBiz principles
 * Standards: NIST CSF 2.0, ISO 27001
 */

import { IamService } from './IamService';

describe('IamService', () => {
    let service: IamService;

    beforeEach(() => {
        service = new IamService();
    });

    describe('getDashboard', () => {
        it('should return dashboard KPIs for a given tenant', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result).toBeDefined();
            expect(result.tenantId).toBe('tenant-001');
        });

        it('should include security posture with score and grade', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result.securityPosture).toBeDefined();
            expect(result.securityPosture.score).toBeGreaterThanOrEqual(0);
            expect(result.securityPosture.maxScore).toBe(100);
            expect(result.securityPosture.grade).toBeDefined();
            expect(typeof result.securityPosture.grade).toBe('string');
        });

        it('should include exactly 6 security categories', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result.securityPosture.categories).toHaveLength(6);
            result.securityPosture.categories.forEach((cat) => {
                expect(cat.name).toBeDefined();
                expect(cat.score).toBeGreaterThanOrEqual(0);
                expect(cat.max).toBe(100);
                expect(['pass', 'warn', 'fail']).toContain(cat.status);
            });
        });

        it('should include identity fabric with multi-region data', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result.identityFabric).toBeDefined();
            expect(result.identityFabric.regions).toBeDefined();
            expect(result.identityFabric.regions.length).toBeGreaterThan(0);
            expect(result.identityFabric.totalIdentities).toBeGreaterThan(0);
            expect(result.identityFabric.activeNow).toBeGreaterThan(0);
        });

        it('should include threat intelligence data', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result.threats).toBeDefined();
            expect(typeof result.threats.active).toBe('number');
            expect(typeof result.threats.mitigated24h).toBe('number');
            expect(typeof result.threats.investigating).toBe('number');
        });

        it('should include cognitive analytics', async () => {
            const result = await service.getDashboard('tenant-001');

            expect(result.cognitive).toBeDefined();
            expect(result.cognitive.genaiEnrichedProfiles).toBeGreaterThan(0);
            expect(result.cognitive.neuraBaselines).toBeGreaterThan(0);
            expect(result.cognitive.xaiDecisions24h).toBeGreaterThan(0);
        });
    });

    describe('getHealth', () => {
        it('should return active status', () => {
            const result = service.getHealth();

            expect(result.status).toBe('active');
            expect(result.module).toBe('identity-management');
            expect(result.cognitive).toContain('GenAI');
        });
    });
});
