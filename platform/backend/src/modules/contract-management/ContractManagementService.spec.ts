import { Test, TestingModule } from '@nestjs/testing';
import { ContractManagementService } from './ContractManagementService';

describe('ContractManagementService', () => {
    let service: ContractManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [ContractManagementService],
        }).compile();
        service = module.get<ContractManagementService>(ContractManagementService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    it('should return hello message', () => {
        expect(service.getHello()).toContain('Contract Management');
    });

    // --- getDashboard ---
    describe('getDashboard', () => {
        it('should return dashboard with metrics', () => {
            const dashboard = service.getDashboard();
            expect(dashboard.module).toBe('ContractManagement');
            expect(dashboard.status).toBe('OPERATIONAL');
            expect(dashboard.metrics.totalContracts).toBeGreaterThan(0);
            expect(dashboard.metrics.totalContractValue).toBeGreaterThan(0);
            expect(dashboard.metrics.renewalRate).toBeGreaterThan(0);
            expect(dashboard.aiAgents).toBeInstanceOf(Array);
        });
    });

    // --- getTemplates ---
    describe('getTemplates', () => {
        it('should return all templates when no filter', () => {
            const templates = service.getTemplates();
            expect(templates).toBeInstanceOf(Array);
            expect(templates.length).toBeGreaterThan(0);
            templates.forEach(t => {
                expect(t.id).toBeDefined();
                expect(t.name).toBeDefined();
                expect(t.clauses).toBeGreaterThan(0);
            });
        });

        it('should filter templates by category', () => {
            const ndaTemplates = service.getTemplates('NDA');
            ndaTemplates.forEach(t => {
                expect(t.category).toBe('NDA');
            });
        });

        it('should return empty for unknown category', () => {
            expect(service.getTemplates('ALIEN')).toEqual([]);
        });
    });

    // --- getExpiringContracts ---
    describe('getExpiringContracts', () => {
        it('should return expiring contracts', () => {
            const contracts = service.getExpiringContracts();
            expect(contracts).toBeInstanceOf(Array);
            expect(contracts.length).toBeGreaterThan(0);
            contracts.forEach(c => {
                expect(c.id).toBeDefined();
                expect(c.type).toBeDefined();
                expect(c.value).toBeGreaterThan(0);
                expect(c.endDate).toBeInstanceOf(Date);
            });
        });

        it('should accept custom days parameter', () => {
            const contracts = service.getExpiringContracts(30);
            expect(contracts).toBeInstanceOf(Array);
        });
    });

    // --- aiReview ---
    describe('aiReview', () => {
        it('should return AI review result', () => {
            const review = service.aiReview({ contractText: 'Test contract...' });
            expect(review.status).toBe('REVIEW_COMPLETE');
            expect(review.riskScore).toBeDefined();
            expect(review.findings).toBeInstanceOf(Array);
            expect(review.findings.length).toBeGreaterThan(0);
            expect(review.missingClauses).toBeInstanceOf(Array);
            expect(review.complianceCheck).toBeDefined();
            expect(review.aiConfidence).toBeGreaterThan(0);
        });
    });

    // --- CRUD Operations ---
    describe('CRUD Operations', () => {
        it('should create a contract', async () => {
            const result = await service.create({ title: 'Test Contract', type: 'NDA' });
            expect(result.status).toBe('SUCCESS');
            expect(result.id).toMatch(/^CTR-/);
        });

        it('should find all records', async () => {
            await service.create({ title: 'Contract A' });
            await service.create({ title: 'Contract B' });
            const all = await service.findAll();
            expect(all.length).toBeGreaterThanOrEqual(2);
        });

        it('should find one by id', async () => {
            const created = await service.create({ title: 'Findable Contract' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Findable Contract');
        });

        it('should return error for non-existent id', async () => {
            const found = await service.findOne('NOPE');
            expect(found.error).toBe('Not Found');
        });

        it('should update an existing record', async () => {
            const created = await service.create({ title: 'Original' });
            await service.update(created.id, { title: 'Amended' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Amended');
        });

        it('should return error when updating non-existent', async () => {
            const result = await service.update('FAKE', {});
            expect(result.error).toBe('Not Found');
        });

        it('should remove a record', async () => {
            const created = await service.create({ title: 'ToTerminate' });
            const result = await service.remove(created.id);
            expect(result.status).toBe('DELETED');
        });
    });

    // --- getCapabilities ---
    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps).toBeInstanceOf(Array);
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('Contract Lifecycle Management (CLM)');
    });
});
