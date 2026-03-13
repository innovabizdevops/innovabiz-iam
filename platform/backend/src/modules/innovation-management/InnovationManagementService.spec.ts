import { Test, TestingModule } from '@nestjs/testing';
import { InnovationManagementService } from './InnovationManagementService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('InnovationManagementService', () => {
    let service: InnovationManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [InnovationManagementService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<InnovationManagementService>(InnovationManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Innovation Management'); });

    describe('getPipeline', () => {
        it('should return innovation pipeline', () => {
            const p = service.getPipeline() as any;
            expect(p).toBeDefined();
            expect(p.totalIdeas || p.length).toBeDefined();
        });
        it('should filter by stage', () => {
            const p = service.getPipeline('IDEATION') as any;
            if (Array.isArray(p) && p.length > 0) p.forEach((i: any) => expect(i.stage).toBe('IDEATION'));
        });
    });

    describe('getHackathons', () => {
        it('should return hackathons', () => {
            const h = service.getHackathons() as any;
            expect(h).toBeDefined();
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
