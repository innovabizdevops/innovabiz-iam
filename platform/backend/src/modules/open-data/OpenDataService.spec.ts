import { Test, TestingModule } from '@nestjs/testing';
import { OpenDataService } from './OpenDataService';

describe('OpenDataService', () => {
    let service: OpenDataService;
    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({ providers: [OpenDataService] }).compile();
        service = module.get<OpenDataService>(OpenDataService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Open Data'); });

    describe('domain methods', () => {
        it('should return datasets', () => { const r = service.getDatasets() as any; expect(r).toBeDefined(); expect(r.totalDatasets).toBeGreaterThan(0); });
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
