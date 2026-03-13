import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('Open Ecosystem E2E', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const moduleFixture: TestingModule = await Test.createTestingModule({
            imports: [AppModule],
        }).compile();
        app = moduleFixture.createNestApplication();
        await app.init();
    });

    afterAll(async () => { await app.close(); });

    describe('/open-banking', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-banking')
                .expect(200)
                .expect((res) => { expect(res.text).toContain('Open Banking'); });
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/open-banking/capabilities')
                .expect(200)
                .expect((res) => { expect(res.body).toBeInstanceOf(Array); });
        });

        it('POST / should create entry', () => {
            return request(app.getHttpServer())
                .post('/open-banking')
                .send({ name: 'E2E Bank API', type: 'PSD2' })
                .expect(201)
                .expect((res) => { expect(res.body.status).toBe('SUCCESS'); });
        });
    });

    describe('/open-finance', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-finance')
                .expect(200);
        });
    });

    describe('/open-insurance', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-insurance')
                .expect(200);
        });
    });

    describe('/open-data', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-data')
                .expect(200);
        });
    });

    describe('/open-health', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-health')
                .expect(200);
        });
    });

    describe('/open-education', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-education')
                .expect(200);
        });
    });

    describe('/open-innovation', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/open-innovation')
                .expect(200);
        });
    });
});
