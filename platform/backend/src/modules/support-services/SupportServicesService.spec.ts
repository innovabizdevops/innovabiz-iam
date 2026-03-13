import { Test, TestingModule } from '@nestjs/testing';
import { SupportServicesService } from './SupportServicesService';

describe('SupportServicesService', () => {
    let service: SupportServicesService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [SupportServicesService],
        }).compile();
        service = module.get<SupportServicesService>(SupportServicesService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Support Services'); });

    describe('getDashboard', () => {
        it('should return support dashboard', () => {
            const d = service.getDashboard() as any;
            expect(d).toBeDefined();
        });
    });

    describe('getTickets', () => {
        it('should return tickets', () => {
            const t = service.getTickets() as any;
            expect(t).toBeDefined();
        });
        it('should filter by status', () => {
            const t = service.getTickets('OPEN') as any;
            if (Array.isArray(t) && t.length > 0) t.forEach((i: any) => expect(i.status).toBe('OPEN'));
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
