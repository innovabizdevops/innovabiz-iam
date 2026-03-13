import { Test, TestingModule } from '@nestjs/testing';
import { NotificationServicesService } from './NotificationServicesService';

describe('NotificationServicesService', () => {
    let service: NotificationServicesService;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [NotificationServicesService],
        }).compile();
        service = module.get<NotificationServicesService>(NotificationServicesService);
    });

    it('should be defined', () => { expect(service).toBeDefined(); });
    it('should return hello', () => { expect(service.getHello()).toContain('Notification Services'); });

    describe('getChannels', () => {
        it('should return omni-channel list', () => {
            const channels = service.getChannels() as any;
            expect(channels).toBeInstanceOf(Array);
            expect(channels.length).toBeGreaterThan(5);
        });
    });

    describe('getTemplates', () => {
        it('should return all templates', () => {
            const templates = service.getTemplates() as any;
            expect(templates).toBeInstanceOf(Array);
            expect(templates.length).toBeGreaterThan(0);
        });
        it('should filter by channel', () => {
            const sms = service.getTemplates('SMS') as any;
            if (Array.isArray(sms)) sms.forEach((t: any) => expect(t.channel).toBe('SMS'));
        });
    });

    describe('send', () => {
        it('should send a notification', () => {
            const r = service.send({ channel: 'EMAIL', recipient: 'test@ibos.io' }) as any;
            expect(r.status).toBe('SENT');
            expect(r.notificationId).toContain('NOTIF-');
        });
    });

    describe('getHistory', () => {
        it('should return notification history', () => {
            const h = service.getHistory() as any;
            expect(h).toBeDefined();
        });
        it('should accept userId filter', () => {
            const h = service.getHistory('USR-001') as any;
            expect(h).toBeDefined();
        });
    });

    describe('CRUD', () => {
        it('should create', async () => { const r = await service.create({ name: 'Rule' }); expect(r.status).toBe('SUCCESS'); });
        it('should findAll', async () => { await service.create({ n: 'A' }); expect((await service.findAll()).length).toBeGreaterThanOrEqual(1); });
        it('should findOne', async () => { const c = await service.create({ n: 'X' }); expect((await service.findOne(c.id)).n).toBe('X'); });
        it('should return not found', async () => { expect((await service.findOne('FAKE')).error).toBe('Not Found'); });
        it('should update', async () => { const c = await service.create({ n: 'A' }); expect((await service.update(c.id, { n: 'B' })).status).toBe('UPDATED'); });
        it('should fail update on missing', async () => { expect((await service.update('FAKE', {})).error).toBe('Not Found'); });
        it('should remove', async () => { const c = await service.create({ n: 'D' }); expect((await service.remove(c.id)).status).toBe('DELETED'); });
    });

    it('should return capabilities', () => { expect(service.getCapabilities().length).toBeGreaterThan(5); });
});
