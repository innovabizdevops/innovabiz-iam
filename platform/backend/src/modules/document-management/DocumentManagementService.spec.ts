import { Test, TestingModule } from '@nestjs/testing';
import { DocumentManagementService } from './DocumentManagementService';

describe('DocumentManagementService', () => {
    let service: DocumentManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [DocumentManagementService],
        }).compile();
        service = module.get<DocumentManagementService>(DocumentManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Document Management'); });

    describe('getLibrary', () => {
        it('should return full library', () => {
            const lib = service.getLibrary() as any;
            expect(lib).toBeDefined();
            expect(lib.totalDocuments || lib.length).toBeDefined();
        });
        it('should filter by category', () => {
            const docs = service.getLibrary('COMPLIANCE') as any;
            if (Array.isArray(docs)) docs.forEach((d: any) => expect(d.category).toBe('COMPLIANCE'));
        });
    });

    describe('search', () => {
        it('should return search results', () => {
            const r = service.search('SOX') as any;
            expect(r.query).toBe('SOX');
            expect(r.totalResults).toBeGreaterThan(0);
        });
    });

    describe('getWorkflows', () => {
        it('should return workflows', () => {
            const wf = service.getWorkflows() as any;
            expect(wf).toBeDefined();
        });
    });

    describe('CRUD', () => {
        it('should create', async () => { const r = await service.create({ title: 'Doc' }); expect(r.status).toBe('SUCCESS'); });
        it('should findAll', async () => { await service.create({ n: 'A' }); expect((await service.findAll()).length).toBeGreaterThanOrEqual(1); });
        it('should findOne', async () => { const c = await service.create({ t: 'X' }); expect((await service.findOne(c.id)).t).toBe('X'); });
        it('should return not found', async () => { expect((await service.findOne('FAKE')).error).toBe('Not Found'); });
        it('should update', async () => { const c = await service.create({ n: 'A' }); expect((await service.update(c.id, { n: 'B' })).status).toBe('UPDATED'); });
        it('should fail update on missing', async () => { expect((await service.update('FAKE', {})).error).toBe('Not Found'); });
        it('should remove', async () => { const c = await service.create({ n: 'D' }); expect((await service.remove(c.id)).status).toBe('DELETED'); });
    });

    it('should return capabilities', () => { expect(service.getCapabilities().length).toBeGreaterThan(5); });
});
