/**
 * 📋 DATA TRANSFER OBJECTS (DTOs) - INNOVABIZ IAM
 * Definições de tipos para transferência de dados
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: OpenAPI 3.0.3, JSON Schema, TypeScript Strict
 */

import { IsEmail, IsString, IsOptional, IsBoolean, IsArray, IsEnum, IsUUID, IsDateString, IsNumber, Min, Max } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';

// ========================================
// USER MANAGEMENT DTOs
// ========================================

export class CreateUserDto {
  @ApiProperty({ description: 'Email do usuário', example: 'user@innovabiz.com' })
  @IsEmail()
  email: string;

  @ApiProperty({ description: 'Nome de usuário único', example: 'johndoe' })
  @IsString()
  username: string;

  @ApiProperty({ description: 'Nome de exibição', example: 'John Doe' })
  @IsString()
  displayName: string;

  @ApiProperty({ description: 'ID do tenant', example: 'tenant-123' })
  @IsUUID()
  tenantId: string;
}

export class UserResponseDto {
  @ApiProperty({ description: 'ID único do usuário' })
  id: string;

  @ApiProperty({ description: 'Email do usuário' })
  email: string;

  @ApiProperty({ description: 'Nome de usuário' })
  username: string;

  @ApiProperty({ description: 'Nome de exibição' })
  displayName: string;

  @ApiProperty({ description: 'Status ativo do usuário' })
  isActive: boolean;

  @ApiProperty({ description: 'Status de verificação' })
  isVerified: boolean;

  @ApiPropertyOptional({ description: 'Último login' })
  lastLoginAt?: Date;

  @ApiProperty({ description: 'Data de criação' })
  createdAt: Date;

  @ApiProperty({ description: 'Número de credenciais registradas' })
  credentialCount: number;
}

// ========================================
// WEBAUTHN DTOs
// ========================================

export class WebAuthnRegistrationDto {
  @ApiProperty({ description: 'ID do usuário' })
  @IsUUID()
  userId: string;

  @ApiProperty({ description: 'Nome de usuário para WebAuthn' })
  @IsString()
  username: string;

  @ApiProperty({ description: 'Nome de exibição para WebAuthn' })
  @IsString()
  displayName: string;

  @ApiPropertyOptional({ description: 'Tipo de attestation', enum: ['none', 'indirect', 'direct'] })
  @IsOptional()
  @IsEnum(['none', 'indirect', 'direct'])
  attestation?: 'none' | 'indirect' | 'direct';

  @ApiPropertyOptional({ description: 'Seleção de autenticador' })
  @IsOptional()
  authenticatorSelection?: {
    authenticatorAttachment?: 'platform' | 'cross-platform';
    userVerification?: 'required' | 'preferred' | 'discouraged';
    residentKey?: 'required' | 'preferred' | 'discouraged';
  };
}

export class WebAuthnAuthenticationDto {
  @ApiPropertyOptional({ description: 'ID do usuário' })
  @IsOptional()
  @IsUUID()
  userId?: string;

  @ApiPropertyOptional({ description: 'Email do usuário' })
  @IsOptional()
  @IsEmail()
  email?: string;

  @ApiProperty({ description: 'ID do tenant' })
  @IsUUID()
  tenantId: string;

  @ApiPropertyOptional({ description: 'Verificação do usuário', enum: ['required', 'preferred', 'discouraged'] })
  @IsOptional()
  @IsEnum(['required', 'preferred', 'discouraged'])
  userVerification?: 'required' | 'preferred' | 'discouraged';
}

export class RegistrationOptionsRequest {
  @ApiProperty({ description: 'Nome de usuário' })
  @IsString()
  username: string;

  @ApiProperty({ description: 'Nome de exibição' })
  @IsString()
  displayName: string;

  @ApiPropertyOptional({ description: 'Tipo de attestation' })
  @IsOptional()
  @IsEnum(['none', 'indirect', 'direct'])
  attestation?: 'none' | 'indirect' | 'direct';

  @ApiPropertyOptional({ description: 'Seleção de autenticador' })
  @IsOptional()
  authenticatorSelection?: {
    authenticatorAttachment?: 'platform' | 'cross-platform';
    userVerification?: 'required' | 'preferred' | 'discouraged';
    residentKey?: 'required' | 'preferred' | 'discouraged';
  };
}

export class AuthenticationOptionsRequest {
  @ApiPropertyOptional({ description: 'Verificação do usuário' })
  @IsOptional()
  @IsEnum(['required', 'preferred', 'discouraged'])
  userVerification?: 'required' | 'preferred' | 'discouraged';

  @ApiPropertyOptional({ description: 'Credenciais permitidas' })
  @IsOptional()
  @IsArray()
  allowCredentials?: Array<{
    id: string;
    type: 'public-key';
    transports?: AuthenticatorTransport[];
  }>;
}

export class WebAuthnRegistrationResponse {
  @ApiProperty({ description: 'ID da credencial' })
  @IsString()
  id: string;

  @ApiProperty({ description: 'Raw ID da credencial' })
  @IsString()
  rawId: string;

  @ApiProperty({ description: 'Resposta do autenticador' })
  response: {
    attestationObject: string;
    clientDataJSON: string;
  };

  @ApiProperty({ description: 'Tipo da credencial' })
  @IsString()
  type: 'public-key';
}

export class WebAuthnAuthenticationResponse {
  @ApiProperty({ description: 'ID da credencial' })
  @IsString()
  id: string;

  @ApiProperty({ description: 'Raw ID da credencial' })
  @IsString()
  rawId: string;

  @ApiProperty({ description: 'Resposta do autenticador' })
  response: {
    authenticatorData: string;
    clientDataJSON: string;
    signature: string;
  };

  @ApiProperty({ description: 'Tipo da credencial' })
  @IsString()
  type: 'public-key';
}

// ========================================
// TOKEN MANAGEMENT DTOs
// ========================================

export class TokenResponseDto {
  @ApiProperty({ description: 'Token de acesso JWT' })
  accessToken: string;

  @ApiProperty({ description: 'Token de renovação' })
  refreshToken: string;

  @ApiProperty({ description: 'Tempo de expiração em segundos' })
  expiresIn: number;
}

export class RefreshTokenDto {
  @ApiProperty({ description: 'Token de renovação' })
  @IsString()
  refreshToken: string;
}

export class AuthenticationResponseDto {
  @ApiProperty({ description: 'Dados do usuário autenticado' })
  user: UserResponseDto;

  @ApiProperty({ description: 'Tokens de autenticação' })
  tokens: TokenResponseDto;

  @ApiProperty({ description: 'Status de autenticação' })
  authenticated: boolean;
}

// ========================================
// RISK ASSESSMENT DTOs
// ========================================

export class RiskAssessmentRequest {
  @ApiProperty({ description: 'ID do usuário' })
  @IsUUID()
  userId: string;

  @ApiProperty({ description: 'ID do tenant' })
  @IsUUID()
  tenantId: string;

  @ApiProperty({ description: 'Tipo do evento', enum: ['registration', 'authentication', 'password_reset'] })
  @IsEnum(['registration', 'authentication', 'password_reset'])
  eventType: 'registration' | 'authentication' | 'password_reset';

  @ApiProperty({ description: 'Endereço IP' })
  @IsString()
  ipAddress: string;

  @ApiProperty({ description: 'User Agent' })
  @IsString()
  userAgent: string;

  @ApiPropertyOptional({ description: 'Fingerprint do dispositivo' })
  @IsOptional()
  @IsString()
  deviceFingerprint?: string;

  @ApiProperty({ description: 'Timestamp do evento' })
  @IsDateString()
  timestamp: Date;

  @ApiPropertyOptional({ description: 'Metadados adicionais' })
  @IsOptional()
  metadata?: Record<string, any>;
}

export class RiskAssessmentResult {
  @ApiProperty({ description: 'Pontuação de risco (0-100)' })
  @IsNumber()
  @Min(0)
  @Max(100)
  riskScore: number;

  @ApiProperty({ description: 'Nível de risco', enum: ['low', 'medium', 'high', 'critical'] })
  @IsEnum(['low', 'medium', 'high', 'critical'])
  riskLevel: 'low' | 'medium' | 'high' | 'critical';

  @ApiProperty({ description: 'Fatores de risco identificados' })
  @IsArray()
  factors: string[];

  @ApiProperty({ description: 'Confiança da avaliação (0-1)' })
  @IsNumber()
  @Min(0)
  @Max(1)
  confidence: number;

  @ApiProperty({ description: 'Recomendações de segurança' })
  @IsArray()
  recommendations: string[];
}

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical';
export type RiskFactorType = 'device' | 'location' | 'behavioral' | 'temporal' | 'credential';

export class DeviceRiskFactors {
  @ApiProperty({ description: 'Fingerprint do dispositivo' })
  fingerprint: string;

  @ApiProperty({ description: 'Dispositivo conhecido' })
  isKnown: boolean;

  @ApiProperty({ description: 'User Agent' })
  userAgent: string;

  @ApiProperty({ description: 'Plataforma do dispositivo' })
  platform: string;

  @ApiPropertyOptional({ description: 'Dispositivo móvel' })
  isMobile?: boolean;

  @ApiPropertyOptional({ description: 'Conexão Tor detectada' })
  isTor?: boolean;

  @ApiPropertyOptional({ description: 'VPN detectada' })
  isVpn?: boolean;
}

export class LocationRiskFactors {
  @ApiProperty({ description: 'Endereço IP' })
  ipAddress: string;

  @ApiProperty({ description: 'País' })
  country: string;

  @ApiPropertyOptional({ description: 'Região/Estado' })
  region?: string;

  @ApiPropertyOptional({ description: 'Cidade' })
  city?: string;

  @ApiProperty({ description: 'Localização conhecida' })
  isKnown: boolean;

  @ApiPropertyOptional({ description: 'VPN detectada' })
  isVpn?: boolean;

  @ApiPropertyOptional({ description: 'Tor detectado' })
  isTor?: boolean;

  @ApiPropertyOptional({ description: 'Pontuação de risco da localização' })
  riskScore?: number;
}

export class BehavioralRiskFactors {
  @ApiProperty({ description: 'Frequência de login', enum: ['low', 'normal', 'high'] })
  loginFrequency: 'low' | 'normal' | 'high';

  @ApiProperty({ description: 'Duração da sessão em segundos' })
  sessionDuration: number;

  @ApiProperty({ description: 'Horas típicas de acesso' })
  typicalHours: number[];

  @ApiProperty({ description: 'Hora atual' })
  currentHour: number;

  @ApiProperty({ description: 'Pontuação de velocidade (0-1)' })
  velocityScore: number;

  @ApiProperty({ description: 'Desvio do padrão (0-1)' })
  patternDeviation: number;

  @ApiPropertyOptional({ description: 'Sucessão rápida de tentativas' })
  rapidSuccession?: boolean;
}

// ========================================
// INTERFACES
// ========================================

export interface UserContext {
  ipAddress: string;
  userAgent: string;
  deviceFingerprint?: string;
  timestamp?: Date;
  sessionId?: string;
}

export interface AuthenticationResult {
  success: boolean;
  user?: UserResponseDto;
  tokens?: TokenResponseDto;
  riskAssessment?: RiskAssessmentResult;
  requiresStepUp?: boolean;
}

export interface RegistrationResult {
  verified: boolean;
  credential?: any;
  riskAssessment?: RiskAssessmentResult;
}

export interface UserMetrics {
  totalLogins: number;
  successfulLogins: number;
  failedLogins: number;
  lastLogin: Date;
  averageSessionDuration: number;
  deviceCount: number;
  riskScore: number;
}

export interface SystemMetrics {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  activeSessions: number;
  authenticationRate: number;
  averageResponseTime: number;
  errorRate: number;
}