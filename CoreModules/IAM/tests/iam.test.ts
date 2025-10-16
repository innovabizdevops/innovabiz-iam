// ========================================
// MÓDULO CORE IAM - TEST SUITE
// Jest + Testing Library
// ========================================

import { describe, it, expect, beforeAll, afterAll } from '@jest/globals';
import request from 'supertest';
import { PrismaClient } from '@prisma/client';
import { app } from '../src/app';
import { AuthenticationService } from '../src/services/authentication';
import { TokenService } from '../src/services/token';

const prisma = new PrismaClient();

describe('IAM Core Module - Test Suite', () => {
  
  beforeAll(async () => {
    // Setup test database
    await prisma.$connect();
    // Seed test data
    await seedTestData();
  });

  afterAll(async () => {
    // Cleanup
    await cleanupTestData();
    await prisma.$disconnect();
  });

  describe('Authentication Tests', () => {
    
    it('should authenticate user with valid credentials', async () => {