import { Test, TestingModule } from '@nestjs/testing';
import { DeviceManagementService } from './DeviceManagementService';

describe('DeviceManagementService', () => {
    let service: DeviceManagementService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [DeviceManagementService],
        }).compile();
        service = module.get<DeviceManagementService>(DeviceManagementService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Device Management'); });

    describe('getFleet', () => {
        it('should return device fleet', () => {
            const f = service.getFleet() as any;
            expect(f).toBeDefined();
            expect(f.totalDevices || f.length).toBeDefined();
        });
        it('should filter by type', () => {
            const f = service.getFleet('IOT') as any;
            if (Array.isArray(f) && f.length > 0) f.forEach((i: any) => expect(i.type).toBe('IOT'));
        });
    });

    describe('getTelemetry', () => {
        it('should return telemetry data', () => {
            const t = service.getTelemetry() as any;
            expect(t).toBeDefined();
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
