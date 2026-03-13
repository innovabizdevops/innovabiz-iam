import { Test, TestingModule } from '@nestjs/testing';
import { IntegrationServicesService } from './IntegrationServicesService';

describe('IntegrationServicesService', () => {
    let service: IntegrationServicesService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [IntegrationServicesService],
        }).compile();
        service = module.get<IntegrationServicesService>(IntegrationServicesService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Integration Services'); });

    describe('getConnectors', () => {
        it('should return connectors', () => {
            const c = service.getConnectors() as any;
            expect(c).toBeDefined();
            expect(c.totalConnectors || c.length).toBeDefined();
        });
        it('should filter by category', () => {
            const c = service.getConnectors('CRM') as any;
            if (Array.isArray(c) && c.length > 0) c.forEach((i: any) => expect(i.category).toBe('CRM'));
        });
    });

    describe('getFlows', () => {
        it('should return integration flows', () => {
            const f = service.getFlows() as any;
            expect(f).toBeDefined();
        });
    });

    describe('getHealth', () => {
        it('should return health status', () => {
            const h = service.getHealth() as any;
            expect(h).toBeDefined();
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
