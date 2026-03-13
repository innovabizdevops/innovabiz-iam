import { Test, TestingModule } from '@nestjs/testing';
import { PartnerManagementService } from './PartnerManagementService';

describe('PartnerManagementService', () => {
    let service: PartnerManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [PartnerManagementService],
        }).compile();
        service = module.get<PartnerManagementService>(PartnerManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Partner Management'); });

    describe('getEcosystem', () => {
        it('should return partner ecosystem', () => {
            const e = service.getEcosystem();
            expect(e).toBeDefined();
        });
    });

    describe('getTiers', () => {
        it('should return partner tiers', () => {
            const t = service.getTiers();
            expect(t).toBeInstanceOf(Array);
            expect(t.length).toBeGreaterThan(0);
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
