import { Test, TestingModule } from '@nestjs/testing';
import { OpenHealthService } from './OpenHealthService';

describe('OpenHealthService', () => {
    let service: OpenHealthService;
    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({ providers: [OpenHealthService] }).compile();
        service = module.get<OpenHealthService>(OpenHealthService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Open Health'); });

    describe('domain methods', () => {
        it('should return FHIR resources', () => { const r = service.getFhirResources() as any; expect(r).toBeDefined(); expect(r.standard).toBe('FHIR R4'); });
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
