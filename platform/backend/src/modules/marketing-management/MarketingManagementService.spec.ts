import { Test, TestingModule } from '@nestjs/testing';
import { MarketingManagementService } from './MarketingManagementService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('MarketingManagementService', () => {
    let service: MarketingManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [MarketingManagementService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<MarketingManagementService>(MarketingManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Marketing Management'); });

    describe('getCampaigns', () => {
        it('should return campaigns', () => {
            const c = service.getCampaigns() as any;
            expect(c).toBeDefined();
            expect(c.totalCampaigns || c.length).toBeDefined();
        });
        it('should filter by channel', () => {
            const c = service.getCampaigns('EMAIL') as any;
            if (Array.isArray(c) && c.length > 0) c.forEach((i: any) => expect(i.channel).toBe('EMAIL'));
        });
    });

    describe('getAnalytics', () => {
        it('should return analytics', () => {
            const a = service.getAnalytics() as any;
            expect(a).toBeDefined();
        });
    });

    describe('getLeads', () => {
        it('should return leads', () => {
            const l = service.getLeads() as any;
            expect(l).toBeDefined();
        });
    });

    describe('CRUD', () => {
        it('should create', async () => { const r = await service.create({ t: 'T' }); expect(r.status).toBe('SUCCESS'); });
        it('should findAll', async () => { await service.create({ n: 'A' }); expect((await service.findAll()).length).toBeGreaterThanOrEqual(1); });
        it('should findOne', async () => { const c = await service.create({ n: 'X' }); expect(((await service.findOne(c.id)) as any).n).toBe('X'); });
        it('should return not found', async () => { expect(((await service.findOne('FAKE')) as any).error).toBe('Not Found'); });
        it('should update', async () => { const c = await service.create({ n: 'A' }); expect(((await service.update(c.id, { n: 'B' })) as any).status).toBe('UPDATED'); });
        it('should fail update on missing', async () => { expect(((await service.update('FAKE', {})) as any).error).toBe('Not Found'); });
        it('should remove', async () => { const c = await service.create({ n: 'D' }); expect((await service.remove(c.id)).status).toBe('DELETED'); });
    });

    it('should return capabilities', () => { expect(service.getCapabilities().length).toBeGreaterThan(5); });
});
