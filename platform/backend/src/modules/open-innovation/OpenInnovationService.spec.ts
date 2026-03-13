import { Test, TestingModule } from '@nestjs/testing';
import { OpenInnovationService } from './OpenInnovationService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('OpenInnovationService', () => {
    let service: OpenInnovationService;
    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({ providers: [OpenInnovationService, PRISMA_MOCK_PROVIDER] }).compile();
        service = module.get<OpenInnovationService>(OpenInnovationService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Open Innovation'); });

    describe('domain methods', () => {
        it('should return challenges', () => { const r = service.getChallenges() as any; expect(r).toBeDefined(); expect(r.totalChallenges).toBeGreaterThan(0); });
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
