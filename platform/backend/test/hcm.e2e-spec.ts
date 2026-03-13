import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('HCM & Processes E2E', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const moduleFixture: TestingModule = await Test.createTestingModule({
            imports: [AppModule],
        }).compile();
        app = moduleFixture.createNestApplication();
        await app.init();
    });

    afterAll(async () => { await app.close(); });

    describe('/human-capital-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/human-capital-management')
                .expect(200)
                .expect((res) => { expect(res.text).toContain('Human Capital Management'); });
        });

        it('GET /dashboard should return HCM dashboard', () => {
            return request(app.getHttpServer())
                .get('/human-capital-management/dashboard')
                .expect(200)
                .expect((res) => {
                    expect(res.body.module).toBe('HumanCapitalManagement');
                    expect(res.body.metrics.totalEmployees).toBeGreaterThan(0);
                });
        });

        it('GET /workforce-planning should return planning data', () => {
            return request(app.getHttpServer())
                .get('/human-capital-management/workforce-planning')
                .expect(200)
                .expect((res) => {
                    expect(res.body.currentHeadcount).toBeGreaterThan(0);
                    expect(res.body.hiringNeeds).toBeInstanceOf(Array);
                });
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/human-capital-management/capabilities')
                .expect(200)
                .expect((res) => { expect(res.body).toBeInstanceOf(Array); });
        });

        it('POST / should create record', () => {
            return request(app.getHttpServer())
                .post('/human-capital-management')
                .send({ name: 'E2E Employee', department: 'Engineering' })
                .expect(201)
                .expect((res) => { expect(res.body.status).toBe('SUCCESS'); });
        });
    });

    describe('/process-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/process-management')
                .expect(200);
        });
    });

    describe('/quality-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/quality-management')
                .expect(200);
        });
    });

    describe('/notification-services', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/notification-services')
                .expect(200);
        });
    });
});
