# ============================================================================
# INNOVABIZ IAM - WebAuthn Backend Dockerfile
# ============================================================================
# Version: 1.0.0
# Date: 31/07/2025
# Author: Equipe de Segurança INNOVABIZ
# Classification: Confidencial - Interno
# 
# Description: Dockerfile para containerização do backend WebAuthn/FIDO2
# Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
# Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
# ============================================================================

# Estágio 1: Build
FROM node:18.17.1-alpine AS builder

# Metadados
LABEL maintainer="innovabizdevops@gmail.com"
LABEL version="1.0.0"
LABEL description="INNOVABIZ IAM WebAuthn/FIDO2 Backend Service"
LABEL org.opencontainers.image.source="https://github.com/innovabiz/iam-webauthn-backend"
LABEL org.opencontainers.image.licenses="PROPRIETARY"

# Instalar dependências do sistema
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    git \
    curl \
    ca-certificates

# Criar usuário não-root
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY package*.json ./
COPY tsconfig.json ./

# Instalar dependências
RUN npm ci --only=production && \
    npm cache clean --force

# Copiar código fonte
COPY src ./src

# Build da aplicação
RUN npm run build

# Remover dependências de desenvolvimento
RUN npm prune --production

# Estágio 2: Runtime
FROM node:18.17.1-alpine AS runtime

# Instalar dependências de runtime
RUN apk add --no-cache \
    dumb-init \
    curl \
    ca-certificates && \
    rm -rf /var/cache/apk/*

# Criar usuário não-root
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Definir diretório de trabalho
WORKDIR /app

# Copiar aplicação buildada
COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodejs:nodejs /app/package*.json ./

# Criar diretórios necessários
RUN mkdir -p logs && \
    chown -R nodejs:nodejs logs

# Configurar usuário
USER nodejs

# Expor porta
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Configurar variáveis de ambiente
ENV NODE_ENV=production
ENV PORT=3000
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json

# Comando de inicialização
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "dist/index.js"]

# ============================================================================
# Multi-stage build para desenvolvimento
# ============================================================================

FROM node:18.17.1-alpine AS development

# Instalar dependências do sistema
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    git \
    curl \
    ca-certificates

# Criar usuário não-root
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY package*.json ./
COPY tsconfig.json ./

# Instalar todas as dependências (incluindo dev)
RUN npm ci

# Copiar código fonte
COPY src ./src
COPY tests ./tests

# Criar diretórios necessários
RUN mkdir -p logs && \
    chown -R nodejs:nodejs /app

# Configurar usuário
USER nodejs

# Expor porta
EXPOSE 3000

# Configurar variáveis de ambiente para desenvolvimento
ENV NODE_ENV=development
ENV PORT=3000
ENV LOG_LEVEL=debug
ENV LOG_FORMAT=simple

# Comando para desenvolvimento
CMD ["npm", "run", "dev"]

# ============================================================================
# Estágio para testes
# ============================================================================

FROM development AS test

# Executar como root para instalar dependências de teste
USER root

# Instalar dependências adicionais para testes
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont

# Configurar Chromium para testes
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium-browser

# Voltar para usuário nodejs
USER nodejs

# Executar testes
CMD ["npm", "test"]

# ============================================================================
# Estágio para análise de segurança
# ============================================================================

FROM node:18.17.1-alpine AS security

# Instalar ferramentas de segurança
RUN npm install -g \
    snyk \
    npm-audit-resolver \
    retire

WORKDIR /app

# Copiar arquivos necessários
COPY package*.json ./
COPY --from=builder /app/node_modules ./node_modules

# Executar análise de segurança
CMD ["sh", "-c", "npm audit && snyk test && retire --path node_modules"]

# ============================================================================
# Configurações de build args e labels
# ============================================================================

# Build args
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=1.0.0

# Labels adicionais
LABEL org.opencontainers.image.created=$BUILD_DATE
LABEL org.opencontainers.image.revision=$VCS_REF
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.title="INNOVABIZ IAM WebAuthn Backend"
LABEL org.opencontainers.image.description="Backend service for WebAuthn/FIDO2 authentication"
LABEL org.opencontainers.image.vendor="INNOVABIZ"
LABEL org.opencontainers.image.authors="innovabizdevops@gmail.com"

# Configurações de segurança
LABEL security.scan="enabled"
LABEL security.compliance="PCI-DSS,GDPR,LGPD,PSD2,ISO-27001"
LABEL security.standards="W3C-WebAuthn-Level-3,FIDO2-CTAP2.1,NIST-SP-800-63B"