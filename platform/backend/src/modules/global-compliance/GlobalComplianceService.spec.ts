import { Test, TestingModule } from '@nestjs/testing';
import { GlobalComplianceService } from './GlobalComplianceService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('GlobalComplianceService', () => {
    let service: GlobalComplianceService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [GlobalComplianceService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<GlobalComplianceService>(GlobalComplianceService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    it('should return hello message', () => {
        expect(service.getHello()).toContain('Global Compliance');
    });

    // --- getJurisdictions ---
    describe('getJurisdictions', () => {
        it('should return jurisdictions array with required structure', () => {
            const jurisdictions = service.getJurisdictions();
            expect(jurisdictions).toBeInstanceOf(Array);
            expect(jurisdictions.length).toBeGreaterThan(5);
            jurisdictions.forEach(j => {
                expect(j.code).toBeDefined();
                expect(j.name).toBeDefined();
                expect(j.region).toBeDefined();
                expect(j.regulations).toBeInstanceOf(Array);
                expect(j.regulations.length).toBeGreaterThan(0);
            });
        });

        it('should include pilot implementation regions', () => {
            const jurisdictions = service.getJurisdictions();
            const codes = jurisdictions.map(j => j.code);
            expect(codes).toContain('EU');
            expect(codes).toContain('BR');
            expect(codes).toContain('AO');
            expect(codes).toContain('PT');
            expect(codes).toContain('US');
        });
    });

    // --- getRegulations ---
    describe('getRegulations', () => {
        it('should return all regulations when no filter', () => {
            const regulations = service.getRegulations();
            expect(regulations).toBeInstanceOf(Array);
            expect(regulations.length).toBeGreaterThan(0);
            regulations.forEach(r => {
                expect(r.code).toBeDefined();
                expect(r.status).toBeDefined();
                expect(r.complianceRate).toBeGreaterThanOrEqual(0);
            });
        });

        it('should filter regulations by region', () => {
            const euRegs = service.getRegulations('EU');
            euRegs.forEach(r => {
                expect(r.jurisdiction.toLowerCase()).toContain('eu');
            });
        });

        it('should return empty array for unknown region', () => {
            const regs = service.getRegulations('MARS');
            expect(regs).toEqual([]);
        });
    });

    // --- getRegulatoryChanges ---
    describe('getRegulatoryChanges', () => {
        it('should return regulatory changes', () => {
            const changes = service.getRegulatoryChanges();
            expect(changes).toBeInstanceOf(Array);
            expect(changes.length).toBeGreaterThan(0);
            changes.forEach(c => {
                expect(c.id).toBeDefined();
                expect(c.changeType).toBeDefined();
                expect(c.impactLevel).toBeDefined();
                expect(c.aiAnalysis).toBeDefined();
                expect(c.effectiveDate).toBeInstanceOf(Date);
            });
        });
    });

    // --- CRUD Operations ---
    describe('CRUD Operations', () => {
        it('should create a compliance entry', async () => {
            const result = await service.create({ regulation: 'GDPR', status: 'COMPLIANT' });
            expect(result.status).toBe('SUCCESS');
            expect(result.id).toMatch(/^COMPL-/);
        });

        it('should find all records', async () => {
            await service.create({ name: 'A' });
            await service.create({ name: 'B' });
            const all = await service.findAll();
            expect(all.length).toBeGreaterThanOrEqual(2);
        });

        it('should find one by id', async () => {
            const created = await service.create({ regulation: 'LGPD' });
            const found = await service.findOne(created.id);
            expect(found.regulation).toBe('LGPD');
        });

        it('should return error for non-existent id', async () => {
            const found = await service.findOne('NON_EXISTENT');
            expect(found.error).toBe('Not Found');
        });

        it('should update an existing record', async () => {
            const created = await service.create({ status: 'DRAFT' });
            const updated = await service.update(created.id, { status: 'ACTIVE' });
            expect(updated.status).toBe('UPDATED');
            const found = await service.findOne(created.id);
            expect(found.status).toBe('ACTIVE');
        });

        it('should return error when updating non-existent', async () => {
            const result = await service.update('FAKE', { data: true });
            expect(result.error).toBe('Not Found');
        });

        it('should remove a record', async () => {
            const created = await service.create({ name: 'ToDelete' });
            const result = await service.remove(created.id);
            expect(result.status).toBe('DELETED');
            const found = await service.findOne(created.id);
            expect(found.error).toBe('Not Found');
        });
    });

    // --- getCapabilities ---
    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps).toBeInstanceOf(Array);
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('Multi-Jurisdiction Regulatory Engine');
    });
});
