import { Test, TestingModule } from '@nestjs/testing';
import { DataManagementService } from './DataManagementService';

describe('DataManagementService', () => {
    let service: DataManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [DataManagementService],
        }).compile();
        service = module.get<DataManagementService>(DataManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Data Management'); });

    describe('getCatalog', () => {
        it('should return data catalog', () => {
            const c = service.getCatalog() as any;
            expect(c).toBeDefined();
            expect(c.totalAssets || c.length).toBeDefined();
        });
        it('should filter by domain', () => {
            const c = service.getCatalog('FINANCE') as any;
            if (Array.isArray(c) && c.length > 0) c.forEach((i: any) => expect(i.domain).toBe('FINANCE'));
        });
    });

    describe('getQuality', () => {
        it('should return quality metrics', () => {
            const q = service.getQuality() as any;
            expect(q).toBeDefined();
        });
    });

    describe('getLineage', () => {
        it('should return data lineage', () => {
            const l = service.getLineage() as any;
            expect(l).toBeDefined();
        });
    });

    describe('CRUD', () => {
        it('should create', async () => { const r = await service.create({ t: 'T' }); expect(r.status).toBe('SUCCESS'); });
        it('should findAll', async () => { await service.create({ n: 'A' }); expect((await service.findAll()).length).toBeGreaterThanOrEqual(1); });
        it('should findOne', async () => { const c = await service.create({ n: 'X' }); expect((await service.findOne(c.id)).n).toBe('X'); });
        it('should return not found', async () => { expect((await service.findOne('FAKE')).error).toBe('Not Found'); });
        it('should update', async () => { const c = await service.create({ n: 'A' }); expect((await service.update(c.id, { n: 'B' })).status).toBe('UPDATED'); });
        it('should fail update on missing', async () => { expect((await service.update('FAKE', {})).error).toBe('Not Found'); });
        it('should remove', async () => { const c = await service.create({ n: 'D' }); expect((await service.remove(c.id)).status).toBe('DELETED'); });
    });

    it('should return capabilities', () => { expect(service.getCapabilities().length).toBeGreaterThan(5); });
});
