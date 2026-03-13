import { Test, TestingModule } from '@nestjs/testing';
import { KnowledgeManagementService } from './KnowledgeManagementService';
import { PRISMA_MOCK_PROVIDER } from '../../test-utils/prisma-mock';

describe('KnowledgeManagementService', () => {
    let service: KnowledgeManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [KnowledgeManagementService, PRISMA_MOCK_PROVIDER],
        }).compile();
        service = module.get<KnowledgeManagementService>(KnowledgeManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Knowledge Management'); });

    describe('searchKnowledgeBase', () => {
        it('should return results', () => {
            const r = service.searchKnowledgeBase();
            expect(r.totalResults).toBeGreaterThan(0);
            expect(r.results).toBeInstanceOf(Array);
            expect(r.aiSummary).toBeDefined();
        });
        it('should filter by domain', () => {
            const r = service.searchKnowledgeBase(undefined, 'COMPLIANCE');
            r.results.forEach(a => expect(a.domain).toBe('COMPLIANCE'));
        });
        it('should filter by query', () => {
            const r = service.searchKnowledgeBase('GDPR');
            r.results.forEach(a => expect(a.title.toLowerCase()).toContain('gdpr'));
        });
    });

    describe('getTaxonomy', () => {
        it('should return taxonomy structure', () => {
            const t = service.getTaxonomy();
            expect(t.domains).toBeInstanceOf(Array);
            expect(t.totalArticles).toBeGreaterThan(0);
            expect(t.languages).toBeInstanceOf(Array);
            expect(t.languages).toContain('pt');
        });
    });

    describe('findExperts', () => {
        it('should return experts', () => {
            const e = service.findExperts();
            expect(e).toBeInstanceOf(Array);
            expect(e.length).toBeGreaterThan(0);
            e.forEach(exp => { expect(exp.rating).toBeGreaterThan(0); });
        });
        it('should filter by topic', () => {
            const e = service.findExperts('GDPR');
            e.forEach(exp => expect(exp.topic.toLowerCase()).toContain('gdpr'));
        });
    });

    describe('CRUD', () => {
        it('should create', async () => {
            const r = await service.create({ title: 'KB' });
            expect(r.status).toBe('SUCCESS');
            expect(r.id).toMatch(/^KB-/);
        });
        it('should findAll', async () => {
            await service.create({ n: 'A' }); await service.create({ n: 'B' });
            expect((await service.findAll()).length).toBeGreaterThanOrEqual(2);
        });
        it('should findOne', async () => {
            const c = await service.create({ t: 'X' });
            expect(((await service.findOne(c.id)) as any).t).toBe('X');
        });
        it('should return not found', async () => {
            expect(((await service.findOne('FAKE')) as any).error).toBe('Not Found');
        });
        it('should update', async () => {
            const c = await service.create({ n: 'A' });
            expect(((await service.update(c.id, { n: 'B' })) as any).status).toBe('UPDATED');
        });
        it('should fail update on missing', async () => {
            expect(((await service.update('FAKE', {})) as any).error).toBe('Not Found');
        });
        it('should remove', async () => {
            const c = await service.create({ n: 'D' });
            expect((await service.remove(c.id)).status).toBe('DELETED');
        });
    });

    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('Enterprise Knowledge Graph');
    });
});
