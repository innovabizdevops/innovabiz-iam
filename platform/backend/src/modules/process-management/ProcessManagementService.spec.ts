import { Test, TestingModule } from '@nestjs/testing';
import { ProcessManagementService } from './ProcessManagementService';

describe('ProcessManagementService', () => {
    let service: ProcessManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [ProcessManagementService],
        }).compile();
        service = module.get<ProcessManagementService>(ProcessManagementService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    it('should return hello message', () => {
        expect(service.getHello()).toContain('Process Management');
    });

    describe('getCatalog', () => {
        it('should return process catalog', () => {
            const catalog = service.getCatalog();
            expect(catalog).toBeInstanceOf(Array);
            expect(catalog.length).toBeGreaterThan(0);
            catalog.forEach(p => {
                expect(p.id).toBeDefined();
                expect(p.domain).toBeDefined();
                expect(p.automation).toBeGreaterThanOrEqual(0);
            });
        });

        it('should filter by domain', () => {
            const finance = service.getCatalog('FINANCE');
            finance.forEach(p => expect(p.domain).toBe('FINANCE'));
        });
    });

    describe('getMiningResults', () => {
        it('should return mining results with insights', () => {
            const results = service.getMiningResults();
            expect(results.processesAnalyzed).toBeGreaterThan(0);
            expect(results.eventsProcessed).toBeGreaterThan(0);
            expect(results.insights).toBeInstanceOf(Array);
            expect(results.conformanceRate).toBeGreaterThan(0);
        });
    });

    describe('getKpis', () => {
        it('should return KPIs', () => {
            const kpis = service.getKpis();
            expect(kpis).toBeInstanceOf(Array);
            expect(kpis.length).toBeGreaterThan(0);
            kpis.forEach(k => {
                expect(k.process).toBeDefined();
                expect(k.kpi).toBeDefined();
                expect(k.current).toBeDefined();
                expect(k.target).toBeDefined();
            });
        });
    });

    describe('CRUD', () => {
        it('should create', async () => {
            const r = await service.create({ name: 'Test' });
            expect(r.status).toBe('SUCCESS');
            expect(r.id).toMatch(/^PRC-/);
        });
        it('should findAll', async () => {
            await service.create({ n: 'A' }); await service.create({ n: 'B' });
            expect((await service.findAll()).length).toBeGreaterThanOrEqual(2);
        });
        it('should findOne', async () => {
            const c = await service.create({ n: 'X' });
            expect((await service.findOne(c.id)).n).toBe('X');
        });
        it('should return not found', async () => {
            expect((await service.findOne('FAKE')).error).toBe('Not Found');
        });
        it('should update', async () => {
            const c = await service.create({ n: 'A' });
            expect((await service.update(c.id, { n: 'B' })).status).toBe('UPDATED');
        });
        it('should fail update on missing', async () => {
            expect((await service.update('FAKE', {})).error).toBe('Not Found');
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
