import { Test, TestingModule } from '@nestjs/testing';
import { QualityManagementService } from './QualityManagementService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('QualityManagementService', () => {
    let service: QualityManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [QualityManagementService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<QualityManagementService>(QualityManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Quality Management'); });

    describe('getDashboard', () => {
        it('should return dashboard with ISO certifications', () => {
            const d = service.getDashboard();
            expect(d.module).toBe('QualityManagement');
            expect(d.metrics.certifications).toContain('ISO 9001:2015');
            expect(d.metrics.sigmaLevel).toBeGreaterThan(0);
            expect(d.metrics.inspectionPassRate).toBeGreaterThan(80);
        });
    });

    describe('getNCRs', () => {
        it('should return NCRs', () => {
            const ncrs = service.getNCRs();
            expect(ncrs).toBeInstanceOf(Array);
            expect(ncrs.length).toBeGreaterThan(0);
            ncrs.forEach(n => { expect(n.severity).toBeDefined(); expect(n.status).toBeDefined(); });
        });
        it('should filter by status', () => {
            service.getNCRs('OPEN').forEach(n => expect(n.status).toBe('OPEN'));
        });
    });

    describe('getCAPAs', () => {
        it('should return CAPAs', () => {
            const capas = service.getCAPAs();
            expect(capas).toBeInstanceOf(Array);
            expect(capas.length).toBeGreaterThan(0);
            capas.forEach(c => { expect(c.id).toBeDefined(); expect(c.type).toBeDefined(); });
        });
    });

    describe('CRUD', () => {
        it('should create', async () => {
            const r = await service.create({ name: 'QMS' });
            expect(r.status).toBe('SUCCESS');
            expect(r.id).toMatch(/^QMS-/);
        });
        it('should findAll', async () => {
            await service.create({ n: 'A' }); await service.create({ n: 'B' });
            expect((await service.findAll()).length).toBeGreaterThanOrEqual(2);
        });
        it('should findOne', async () => {
            const c = await service.create({ n: 'X' });
            expect(((await service.findOne(c.id)) as any).n).toBe('X');
        });
        it('should return not found', async () => {
            expect(((await service.findOne('FAKE')) as any).error).toBe('Not Found');
        });
        it('should update', async () => {
            const c = await service.create({ n: 'A' });
            expect(((await service.update(c.id, { n: 'B' })) as any).status).toBe('UPDATED');
        });
        it('should fail update on missing', async () => {
            expect(((await service.update('FAKE', {})) as any).error).toBe('Not Found');
        });
        it('should remove', async () => {
            const c = await service.create({ n: 'D' });
            expect((await service.remove(c.id)).status).toBe('DELETED');
        });
    });

    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps).toBeInstanceOf(Array);
        expect(caps.length).toBeGreaterThan(5);
    });
});
