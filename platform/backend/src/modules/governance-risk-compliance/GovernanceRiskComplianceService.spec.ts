import { Test, TestingModule } from '@nestjs/testing';
import { GovernanceRiskComplianceService } from './GovernanceRiskComplianceService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('GovernanceRiskComplianceService', () => {
    let service: GovernanceRiskComplianceService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [GovernanceRiskComplianceService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<GovernanceRiskComplianceService>(GovernanceRiskComplianceService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    // --- getHello ---
    it('should return hello message', () => {
        const result = service.getHello();
        expect(result).toContain('Governance Risk');
        expect(typeof result).toBe('string');
    });

    // --- getDashboard ---
    describe('getDashboard', () => {
        it('should return dashboard with required fields', () => {
            const dashboard = service.getDashboard();
            expect(dashboard.module).toBe('GovernanceRiskCompliance');
            expect(dashboard.status).toBe('OPERATIONAL');
            expect(dashboard.metrics).toBeDefined();
            expect(dashboard.metrics.totalControls).toBeGreaterThan(0);
            expect(dashboard.metrics.overallComplianceRate).toBeGreaterThan(0);
            expect(dashboard.aiAgents).toBeInstanceOf(Array);
            expect(dashboard.aiAgents.length).toBeGreaterThan(0);
            expect(dashboard.lastUpdated).toBeInstanceOf(Date);
        });

        it('should have frameworks array', () => {
            const dashboard = service.getDashboard();
            expect(dashboard.frameworks).toContain('COBIT 2019');
            expect(dashboard.frameworks).toContain('ISO 31000:2018');
        });
    });

    // --- getControls ---
    describe('getControls', () => {
        it('should return all controls when no filter', () => {
            const controls = service.getControls();
            expect(controls).toBeInstanceOf(Array);
            expect(controls.length).toBeGreaterThan(0);
            controls.forEach(c => {
                expect(c.id).toBeDefined();
                expect(c.framework).toBeDefined();
                expect(c.status).toBeDefined();
            });
        });

        it('should filter controls by framework', () => {
            const soxControls = service.getControls('SOX');
            soxControls.forEach(c => {
                expect(c.framework).toBe('SOX');
            });
        });

        it('should return empty array for unknown framework', () => {
            const controls = service.getControls('UNKNOWN_FRAMEWORK');
            expect(controls).toEqual([]);
        });
    });

    // --- testControl ---
    describe('testControl', () => {
        it('should return test result for a control', () => {
            const result = service.testControl('CTL-001');
            expect(result.controlId).toBe('CTL-001');
            expect(result.testResult).toBe('PASS');
            expect(result.testDate).toBeInstanceOf(Date);
            expect(result.nextTestDate).toBeInstanceOf(Date);
            expect(result.evidence).toBeInstanceOf(Array);
        });
    });

    // --- getRiskMatrix ---
    describe('getRiskMatrix', () => {
        it('should return risk matrix entries', () => {
            const matrix = service.getRiskMatrix();
            expect(matrix).toBeInstanceOf(Array);
            expect(matrix.length).toBeGreaterThan(0);
            matrix.forEach(entry => {
                expect(entry.riskId).toBeDefined();
                expect(entry.likelihood).toBeGreaterThanOrEqual(1);
                expect(entry.likelihood).toBeLessThanOrEqual(5);
                expect(entry.impact).toBeGreaterThanOrEqual(1);
                expect(entry.impact).toBeLessThanOrEqual(5);
                expect(entry.inherentScore).toBeGreaterThanOrEqual(entry.residualScore);
                expect(entry.mitigationControls).toBeInstanceOf(Array);
            });
        });
    });

    // --- CRUD Operations ---
    describe('CRUD Operations', () => {
        it('should create a record', async () => {
            const result = await service.create({ name: 'Test GRC Entry', framework: 'COBIT' });
            expect(result.status).toBe('SUCCESS');
            expect(result.id).toBeDefined();
            expect(result.id).toMatch(/^GRC-/);
        });

        it('should find all records', async () => {
            await service.create({ name: 'Entry A' });
            await service.create({ name: 'Entry B' });
            const all = await service.findAll();
            expect(all.length).toBeGreaterThanOrEqual(2);
        });

        it('should find one record by id', async () => {
            const created = await service.create({ name: 'Findable Entry' });
            const found = await service.findOne(created.id);
            expect(found.name).toBe('Findable Entry');
            expect(found.id).toBe(created.id);
        });

        it('should return error for non-existent id', async () => {
            const found = await service.findOne('NON_EXISTENT_ID');
            expect(found.error).toBe('Not Found');
        });

        it('should update an existing record', async () => {
            const created = await service.create({ name: 'Original' });
            const updated = await service.update(created.id, { name: 'Updated' });
            expect(updated.status).toBe('UPDATED');
            const found = await service.findOne(created.id);
            expect(found.name).toBe('Updated');
            expect(found.updatedAt).toBeInstanceOf(Date);
        });

        it('should return error when updating non-existent record', async () => {
            const result = await service.update('NON_EXISTENT', { name: 'Test' });
            expect(result.error).toBe('Not Found');
        });

        it('should remove a record', async () => {
            const created = await service.create({ name: 'To Delete' });
            const result = await service.remove(created.id);
            expect(result.status).toBe('DELETED');
            const found = await service.findOne(created.id);
            expect(found.error).toBe('Not Found');
        });
    });

    // --- getCapabilities ---
    describe('getCapabilities', () => {
        it('should return capabilities array', () => {
            const capabilities = service.getCapabilities();
            expect(capabilities).toBeInstanceOf(Array);
            expect(capabilities.length).toBeGreaterThan(5);
            expect(capabilities).toContain('GRC Unified Dashboard');
        });
    });
});
