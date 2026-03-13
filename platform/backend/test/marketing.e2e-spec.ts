import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('Marketing & Partners E2E', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const moduleFixture: TestingModule = await Test.createTestingModule({
            imports: [AppModule],
        }).compile();
        app = moduleFixture.createNestApplication();
        await app.init();
    });

    afterAll(async () => { await app.close(); });

    describe('/marketing-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/marketing-management')
                .expect(200)
                .expect((res) => { expect(res.text).toContain('Marketing Management'); });
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/marketing-management/capabilities')
                .expect(200)
                .expect((res) => { expect(res.body).toBeInstanceOf(Array); });
        });

        it('POST / should create campaign', () => {
            return request(app.getHttpServer())
                .post('/marketing-management')
                .send({ name: 'E2E Campaign', channel: 'EMAIL' })
                .expect(201)
                .expect((res) => { expect(res.body.status).toBe('SUCCESS'); });
        });
    });

    describe('/partner-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/partner-management')
                .expect(200);
        });
    });

    describe('/innovation-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/innovation-management')
                .expect(200);
        });
    });

    describe('/support-services', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/support-services')
                .expect(200);
        });
    });

    describe('/vendor-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/vendor-management')
                .expect(200);
        });
    });
});
