import { Test, TestingModule } from '@nestjs/testing';
import { ComplianceRiskManagementService } from './ComplianceRiskManagementService';

describe('ComplianceRiskManagementService', () => {
    let service: ComplianceRiskManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [ComplianceRiskManagementService],
        }).compile();
        service = module.get<ComplianceRiskManagementService>(ComplianceRiskManagementService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    it('should return hello message', () => {
        expect(service.getHello()).toContain('Compliance Risk Management');
    });

    // --- getRiskRegister ---
    describe('getRiskRegister', () => {
        it('should return all risks when no filter', () => {
            const risks = service.getRiskRegister();
            expect(risks).toBeInstanceOf(Array);
            expect(risks.length).toBeGreaterThan(3);
            risks.forEach(r => {
                expect(r.id).toBeDefined();
                expect(r.category).toBeDefined();
                expect(r.likelihood).toBeGreaterThanOrEqual(1);
                expect(r.impact).toBeGreaterThanOrEqual(1);
                expect(r.inherentScore).toBeGreaterThanOrEqual(r.residualScore);
            });
        });

        it('should filter risks by category', () => {
            const cyber = service.getRiskRegister('CYBER');
            cyber.forEach(r => {
                expect(r.category).toBe('CYBER');
            });
        });

        it('should return empty for unknown category', () => {
            expect(service.getRiskRegister('ALIEN')).toEqual([]);
        });
    });

    // --- getImpactAssessment ---
    describe('getImpactAssessment', () => {
        it('should return impact assessment with financial data', () => {
            const assessment = service.getImpactAssessment('RISK-002');
            expect(assessment.riskId).toBe('RISK-002');
            expect(assessment.assessment).toBeDefined();
            expect(assessment.assessment.financialImpact).toBeDefined();
            expect(assessment.assessment.financialImpact.worstCase).toBeGreaterThan(0);
            expect(assessment.aiInsights).toBeInstanceOf(Array);
            expect(assessment.mitigationOptions).toBeInstanceOf(Array);
            assessment.mitigationOptions.forEach(o => {
                expect(o.cost).toBeGreaterThan(0);
                expect(o.riskReduction).toBeGreaterThan(0);
                expect(o.roi).toBeGreaterThan(0);
            });
        });

        it('should default to RISK-002 when no riskId', () => {
            const assessment = service.getImpactAssessment();
            expect(assessment.riskId).toBe('RISK-002');
        });
    });

    // --- getMitigationPlans ---
    describe('getMitigationPlans', () => {
        it('should return mitigation plans', () => {
            const plans = service.getMitigationPlans();
            expect(plans).toBeInstanceOf(Array);
            expect(plans.length).toBeGreaterThan(0);
            plans.forEach(p => {
                expect(p.id).toBeDefined();
                expect(p.riskId).toBeDefined();
                expect(p.progress).toBeGreaterThanOrEqual(0);
                expect(p.progress).toBeLessThanOrEqual(100);
                expect(p.budget).toBeGreaterThan(0);
            });
        });
    });

    // --- getKRI ---
    describe('getKRI', () => {
        it('should return key risk indicators', () => {
            const kris = service.getKRI();
            expect(kris).toBeInstanceOf(Array);
            expect(kris.length).toBeGreaterThan(3);
            kris.forEach(kri => {
                expect(kri.id).toBeDefined();
                expect(['GREEN', 'AMBER', 'RED']).toContain(kri.status);
                expect(['IMPROVING', 'STABLE', 'DETERIORATING']).toContain(kri.trend);
            });
        });
    });

    // --- CRUD Operations ---
    describe('CRUD Operations', () => {
        it('should create a risk entry', async () => {
            const result = await service.create({ title: 'Test Risk', category: 'CYBER' });
            expect(result.status).toBe('SUCCESS');
            expect(result.id).toMatch(/^RISK-/);
        });

        it('should find all records', async () => {
            await service.create({ title: 'Risk A' });
            await service.create({ title: 'Risk B' });
            const all = await service.findAll();
            expect(all.length).toBeGreaterThanOrEqual(2);
        });

        it('should find one by id', async () => {
            const created = await service.create({ title: 'Findable Risk' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Findable Risk');
        });

        it('should return error for non-existent id', async () => {
            const found = await service.findOne('GHOST');
            expect(found.error).toBe('Not Found');
        });

        it('should update an existing record', async () => {
            const created = await service.create({ title: 'Original Risk' });
            await service.update(created.id, { title: 'Mitigated Risk' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Mitigated Risk');
        });

        it('should return error when updating non-existent', async () => {
            const result = await service.update('GHOST', {});
            expect(result.error).toBe('Not Found');
        });

        it('should remove a record', async () => {
            const created = await service.create({ title: 'ToClose' });
            const result = await service.remove(created.id);
            expect(result.status).toBe('DELETED');
        });
    });

    // --- getCapabilities ---
    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps).toBeInstanceOf(Array);
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('Enterprise Risk Register');
    });
});
