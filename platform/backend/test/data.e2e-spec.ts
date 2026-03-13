import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('Data & Integration E2E', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const moduleFixture: TestingModule = await Test.createTestingModule({
            imports: [AppModule],
        }).compile();
        app = moduleFixture.createNestApplication();
        await app.init();
    });

    afterAll(async () => { await app.close(); });

    describe('/data-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/data-management')
                .expect(200)
                .expect((res) => { expect(res.text).toContain('Data Management'); });
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/data-management/capabilities')
                .expect(200)
                .expect((res) => { expect(res.body).toBeInstanceOf(Array); });
        });

        it('POST / should create data entry', () => {
            return request(app.getHttpServer())
                .post('/data-management')
                .send({ name: 'E2E Data Asset', domain: 'FINANCE' })
                .expect(201)
                .expect((res) => { expect(res.body.status).toBe('SUCCESS'); });
        });
    });

    describe('/device-management', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/device-management')
                .expect(200);
        });
    });

    describe('/integration-services', () => {
        it('GET / should return hello', () => {
            return request(app.getHttpServer())
                .get('/integration-services')
                .expect(200);
        });
    });
});
