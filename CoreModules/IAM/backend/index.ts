// ========================================
// MÓDULO CORE IAM - BACKEND SERVICE
// Node.js 22+ com TypeScript 5.3+
// ========================================

import express from 'express';
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { buildSchema } from 'type-graphql';
import { Container } from 'typedi';
import { PrismaClient } from '@prisma/client';
import Redis from 'ioredis';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import rateLimit from 'express-rate-limit';

// Security imports
import { OpenPolicyAgent } from './security/opa';
import { VaultClient } from './security/vault';
import { ZeroTrustMiddleware } from './middleware/zero-trust';

// Service imports
import { AuthenticationService } from './services/authentication';
import { AuthorizationService } from './services/authorization';
import { IdentityService } from './services/identity';
import { TokenService } from './services/token';
import { SessionService } from './services/session';

// GraphQL resolvers
import { AuthResolver } from './resolvers/auth.resolver';