import { Test, TestingModule } from '@nestjs/testing';
import { AuditManagementService } from './AuditManagementService';

describe('AuditManagementService', () => {
    let service: AuditManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [AuditManagementService],
        }).compile();
        service = module.get<AuditManagementService>(AuditManagementService);
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });

    it('should return hello message', () => {
        expect(service.getHello()).toContain('Audit Management');
    });

    // --- getAuditPlan ---
    describe('getAuditPlan', () => {
        it('should return audit plan for current year', () => {
            const plan = service.getAuditPlan();
            expect(plan.year).toBe('2026');
            expect(plan.status).toBe('APPROVED');
            expect(plan.totalAudits).toBeGreaterThan(0);
            expect(plan.schedule).toBeInstanceOf(Array);
            expect(plan.schedule.length).toBe(4); // Q1-Q4
        });

        it('should accept custom year', () => {
            const plan = service.getAuditPlan('2025');
            expect(plan.year).toBe('2025');
        });
    });

    // --- getAuditTrail ---
    describe('getAuditTrail', () => {
        it('should return audit trail entries', () => {
            const trail = service.getAuditTrail();
            expect(trail).toBeInstanceOf(Array);
            expect(trail.length).toBeGreaterThan(0);
            trail.forEach(entry => {
                expect(entry.id).toBeDefined();
                expect(entry.action).toBeDefined();
                expect(entry.immutableHash).toContain('sha256:');
            });
        });

        it('should filter trail by entity', () => {
            const paymentTrail = service.getAuditTrail('Payment');
            paymentTrail.forEach(t => {
                expect(t.entity.toLowerCase()).toBe('payment');
            });
        });

        it('should respect limit parameter', () => {
            const trail = service.getAuditTrail(undefined, 1);
            expect(trail.length).toBeLessThanOrEqual(1);
        });
    });

    // --- getFindings ---
    describe('getFindings', () => {
        it('should return all findings when no filter', () => {
            const findings = service.getFindings();
            expect(findings).toBeInstanceOf(Array);
            expect(findings.length).toBeGreaterThan(0);
            findings.forEach(f => {
                expect(f.severity).toBeDefined();
                expect(f.status).toBeDefined();
                expect(f.owner).toBeDefined();
            });
        });

        it('should filter findings by status', () => {
            const open = service.getFindings('OPEN');
            open.forEach(f => {
                expect(f.status).toBe('OPEN');
            });
        });
    });

    // --- submitEvidence ---
    describe('submitEvidence', () => {
        it('should accept evidence submission', () => {
            const result = service.submitEvidence({ type: 'screenshot', data: 'base64...' });
            expect(result.status).toBe('ACCEPTED');
            expect(result.evidenceId).toMatch(/^EVD-/);
            expect(result.hash).toContain('sha256:');
            expect(result.timestamp).toBeInstanceOf(Date);
        });
    });

    // --- CRUD Operations ---
    describe('CRUD Operations', () => {
        it('should create an audit entry', async () => {
            const result = await service.create({ title: 'SOX Audit Q1' });
            expect(result.status).toBe('SUCCESS');
            expect(result.id).toMatch(/^AUD-/);
        });

        it('should find all records', async () => {
            await service.create({ title: 'A' });
            await service.create({ title: 'B' });
            const all = await service.findAll();
            expect(all.length).toBeGreaterThanOrEqual(2);
        });

        it('should find one by id', async () => {
            const created = await service.create({ title: 'Findable Audit' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Findable Audit');
        });

        it('should return error for non-existent id', async () => {
            const found = await service.findOne('NON_EXISTENT');
            expect(found.error).toBe('Not Found');
        });

        it('should update an existing record', async () => {
            const created = await service.create({ title: 'Original' });
            await service.update(created.id, { title: 'Updated' });
            const found = await service.findOne(created.id);
            expect(found.title).toBe('Updated');
        });

        it('should return error when updating non-existent', async () => {
            const result = await service.update('FAKE', {});
            expect(result.error).toBe('Not Found');
        });

        it('should remove a record', async () => {
            const created = await service.create({ title: 'ToDelete' });
            const result = await service.remove(created.id);
            expect(result.status).toBe('DELETED');
        });
    });

    // --- getCapabilities ---
    it('should return capabilities', () => {
        const caps = service.getCapabilities();
        expect(caps).toBeInstanceOf(Array);
        expect(caps.length).toBeGreaterThan(5);
        expect(caps).toContain('Immutable Audit Trail (Blockchain-backed)');
    });
});
