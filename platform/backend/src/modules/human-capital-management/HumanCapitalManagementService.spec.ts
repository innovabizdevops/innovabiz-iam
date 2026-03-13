import { Test, TestingModule } from '@nestjs/testing';
import { HumanCapitalManagementService } from './HumanCapitalManagementService';

describe('HumanCapitalManagementService', () => {
    let service: HumanCapitalManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [HumanCapitalManagementService],
        }).compile();
        service = module.get<HumanCapitalManagementService>(HumanCapitalManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Human Capital Management'); });

    describe('getDashboard', () => {
        it('should return HCM dashboard', () => {
            const d = service.getDashboard();
            expect(d.module).toBe('HumanCapitalManagement');
            expect(d.metrics.totalEmployees).toBeGreaterThan(0);
            expect(d.metrics.turnoverRate).toBeGreaterThan(0);
            expect(d.metrics.engagementScore).toBeGreaterThan(0);
            expect(d.metrics.countries).toBeGreaterThan(0);
        });
    });

    describe('getWorkforcePlanning', () => {
        it('should return workforce plan', () => {
            const wp = service.getWorkforcePlanning();
            expect(wp.currentHeadcount).toBeGreaterThan(0);
            expect(wp.projectedHeadcount).toBeGreaterThan(wp.currentHeadcount);
            expect(wp.hiringNeeds).toBeInstanceOf(Array);
            expect(wp.attritionRisk).toBeInstanceOf(Array);
            expect(wp.aiRecommendations).toBeInstanceOf(Array);
        });
    });

    describe('getSkillsMatrix', () => {
        it('should return skills matrix', () => {
            const sm = service.getSkillsMatrix();
            expect(sm.totalSkillsTracked).toBeGreaterThan(0);
            expect(sm.skills).toBeInstanceOf(Array);
            sm.skills.forEach(s => {
                expect(s.proficient).toBeGreaterThanOrEqual(0);
                expect(s.criticality).toBeDefined();
            });
        });
        it('should filter by department', () => {
            const sm = service.getSkillsMatrix('Engineering');
            expect(sm).toBeInstanceOf(Array);
            (sm as any[]).forEach(s => expect(s.department).toBe('Engineering'));
        });
    });

    describe('getSuccessionPlanning', () => {
        it('should return succession pipeline', () => {
            const sp = service.getSuccessionPlanning();
            expect(sp).toBeInstanceOf(Array);
            expect(sp.length).toBeGreaterThan(0);
            sp.forEach(p => {
                expect(p.position).toBeDefined();
                expect(p.riskIfVacant).toBeDefined();
                expect(p.readyNow).toBeGreaterThanOrEqual(0);
            });
        });
    });

    describe('CRUD', () => {
        it('should create', async () => {
            const r = await service.create({ name: 'Employee' });
            expect(r.status).toBe('SUCCESS');
            expect(r.id).toMatch(/^HCM-/);
        });
        it('should findAll', async () => {
            await service.create({ n: 'A' }); await service.create({ n: 'B' });
            expect((await service.findAll()).length).toBeGreaterThanOrEqual(2);
        });
        it('should findOne', async () => {
            const c = await service.create({ n: 'X' });
            expect((await service.findOne(c.id)).n).toBe('X');
        });
        it('should return not found', async () => {
            expect((await service.findOne('FAKE')).error).toBe('Not Found');
        });
        it('should update', async () => {
            const c = await service.create({ n: 'A' });
            expect((await service.update(c.id, { n: 'B' })).status).toBe('UPDATED');
        });
        it('should fail update on missing', async () => {
            expect((await service.update('FAKE', {})).error).toBe('Not Found');
        });
        it('should remove', async () => {
            const c = await service.create({ n: 'D' });
            expect((await service.remove(c.id)).status).toBe('DELETED');
        });
    });

    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('AI Workforce Planning & Forecasting');
    });
});
