# OAuth 2.0 & OpenID Connect - INNOVABIZ IAM

## 📋 Informações do Documento

- **Autor**: Eduardo Jeremias
- **Data**: 09/01/2025
- **Versão**: 1.0.0
- **Módulo**: IAM Core
- **Classificação**: Técnico
- **Revisão**: Trimestral
- **Audiência**: Desenvolvedores, Arquitetos, Security Team

## 🎯 Visão Geral

Este documento detalha a implementação de OAuth 2.0 e OpenID Connect (OIDC) no módulo IAM da plataforma INNOVABIZ, fornecendo autenticação e autorização seguras para aplicações internas e externas.

## 🔐 OAuth 2.0 Implementation

### 1. Authorization Server Configuration

```typescript
// oauth-server.config.ts
import { Injectable } from '@nestjs/common';
import { OAuth2Server } from 'oauth2-server';

@Injectable()
export class OAuthServerConfig {
  private server: OAuth2Server;

  constructor() {
    this.server = new OAuth2Server({
      model: this.oauthModel,
      authenticateHandler: {
        handle: this.authenticateHandler
      },
      allowBearerTokensInQueryString: false,
      allowEmptyState: false,
      authorizationCodeLifetime: 600, // 10 minutes
      accessTokenLifetime: 900, // 15 minutes
      refreshTokenLifetime: 86400, // 24 hours
      allowExtendedTokenAttributes: true,
      requireClientAuthentication: {
        password: false,
        refresh_token: true,
        authorization_code: true,
        client_credentials: true
      }
    });
  }

  /**
   * OAuth 2.0 Model Implementation
   */
  private oauthModel = {
    /**
     * Get access token
     */
    getAccessToken: async (accessToken: string): Promise<Token | false> => {
      const token = await this.tokenRepository.findOne({
        where: { accessToken },
        relations: ['client', 'user']
      });

      if (!token || token.accessTokenExpiresAt < new Date()) {
        return false;
      }

      return {
        accessToken: token.accessToken,
        accessTokenExpiresAt: token.accessTokenExpiresAt,
        scope: token.scope,
        client: { id: token.client.id },
        user: { id: token.user.id }
      };
    },

    /**
     * Get client
     */
    getClient: async (clientId: string, clientSecret: string): Promise<Client | false> => {
      const client = await this.clientRepository.findOne({
        where: { clientId }
      });

      if (!client || !await this.verifyClientSecret(clientSecret, client.clientSecret)) {
        return false;
      }

      return {
        id: client.clientId,
        redirectUris: client.redirectUris,
        grants: client.grants,
        accessTokenLifetime: client.accessTokenLifetime,
        refreshTokenLifetime: client.refreshTokenLifetime,
        scopes: client.scopes
      };
    },

    /**
     * Save authorization code
     */
    saveAuthorizationCode: async (
      code: AuthorizationCode,
      client: Client,
      user: User
    ): Promise<AuthorizationCode> => {
      const authCode = await this.authCodeRepository.save({
        authorizationCode: code.authorizationCode,
        expiresAt: code.expiresAt,
        redirectUri: code.redirectUri,
        scope: code.scope,
        clientId: client.id,
        userId: user.id,
        codeChallenge: code.codeChallenge,
        codeChallengeMethod: code.codeChallengeMethod
      });

      return {
        authorizationCode: authCode.authorizationCode,
        expiresAt: authCode.expiresAt,
        redirectUri: authCode.redirectUri,
        scope: authCode.scope,
        client: { id: client.id },
        user: { id: user.id }
      };
    },

    /**
     * Revoke authorization code
     */
    revokeAuthorizationCode: async (code: AuthorizationCode): Promise<boolean> => {
      await this.authCodeRepository.delete({
        authorizationCode: code.authorizationCode
      });
      return true;
    },

    /**
     * Save token
     */
    saveToken: async (
      token: Token,
      client: Client,
      user: User
    ): Promise<Token> => {
      const savedToken = await this.tokenRepository.save({
        accessToken: token.accessToken,
        accessTokenExpiresAt: token.accessTokenExpiresAt,
        refreshToken: token.refreshToken,
        refreshTokenExpiresAt: token.refreshTokenExpiresAt,
        scope: token.scope,
        clientId: client.id,
        userId: user.id
      });

      // Emit token issued event
      await this.eventEmitter.emit('token.issued', {
        tokenId: savedToken.id,
        clientId: client.id,
        userId: user.id,
        scope: token.scope
      });

      return token;
    },

    /**
     * Verify scope
     */
    verifyScope: async (
      token: Token,
      scope: string | string[]
    ): Promise<boolean> => {
      if (!token.scope) return false;
      
      const tokenScopes = token.scope.split(' ');
      const requiredScopes = Array.isArray(scope) ? scope : scope.split(' ');
      
      return requiredScopes.every(s => tokenScopes.includes(s));
    }
  };
}
```

### 2. Grant Types Implementation

```typescript
// grant-types.service.ts
@Injectable()
export class GrantTypesService {
  /**
   * Authorization Code Grant with PKCE
   */
  async authorizationCodeGrant(
    request: AuthorizationRequest
  ): Promise<AuthorizationResponse> {
    // Validate client
    const client = await this.validateClient(request.clientId);
    
    // Validate redirect URI
    if (!client.redirectUris.includes(request.redirectUri)) {
      throw new Error('Invalid redirect URI');
    }

    // Generate state for CSRF protection
    const state = request.state || crypto.randomBytes(32).toString('hex');
    
    // PKCE validation
    if (client.requirePKCE && !request.codeChallenge) {
      throw new Error('PKCE required but code_challenge missing');
    }

    // Build authorization URL
    const authUrl = new URL('/oauth/authorize', this.baseUrl);
    authUrl.searchParams.set('client_id', request.clientId);
    authUrl.searchParams.set('redirect_uri', request.redirectUri);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('scope', request.scope);
    authUrl.searchParams.set('state', state);
    
    if (request.codeChallenge) {
      authUrl.searchParams.set('code_challenge', request.codeChallenge);
      authUrl.searchParams.set('code_challenge_method', request.codeChallengeMethod || 'S256');
    }

    return {
      authorizationUrl: authUrl.toString(),
      state
    };
  }

  /**
   * Client Credentials Grant
   */
  async clientCredentialsGrant(
    clientId: string,
    clientSecret: string,
    scope?: string
  ): Promise<TokenResponse> {
    // Authenticate client
    const client = await this.authenticateClient(clientId, clientSecret);
    
    if (!client.grants.includes('client_credentials')) {
      throw new Error('Client credentials grant not allowed');
    }

    // Validate scope
    const grantedScope = await this.validateScope(scope, client.scopes);

    // Generate access token
    const accessToken = await this.generateAccessToken({
      clientId: client.id,
      scope: grantedScope,
      grantType: 'client_credentials'
    });

    return {
      access_token: accessToken.token,
      token_type: 'Bearer',
      expires_in: accessToken.expiresIn,
      scope: grantedScope
    };
  }

  /**
   * Refresh Token Grant
   */
  async refreshTokenGrant(
    refreshToken: string,
    clientId?: string,
    clientSecret?: string
  ): Promise<TokenResponse> {
    // Validate refresh token
    const storedToken = await this.validateRefreshToken(refreshToken);
    
    if (!storedToken) {
      throw new Error('Invalid refresh token');
    }

    // Validate client if provided
    if (clientId) {
      const client = await this.authenticateClient(clientId, clientSecret);
      if (client.id !== storedToken.clientId) {
        throw new Error('Token does not belong to client');
      }
    }

    // Check if refresh token is expired
    if (storedToken.refreshTokenExpiresAt < new Date()) {
      throw new Error('Refresh token expired');
    }

    // Revoke old tokens
    await this.revokeToken(storedToken.accessToken);
    await this.revokeToken(refreshToken);

    // Generate new token pair
    const newTokens = await this.generateTokenPair({
      userId: storedToken.userId,
      clientId: storedToken.clientId,
      scope: storedToken.scope
    });

    return {
      access_token: newTokens.accessToken,
      refresh_token: newTokens.refreshToken,
      token_type: 'Bearer',
      expires_in: newTokens.expiresIn,
      scope: storedToken.scope
    };
  }

  /**
   * Device Authorization Grant
   */
  async deviceAuthorizationGrant(
    clientId: string
  ): Promise<DeviceAuthorizationResponse> {
    const client = await this.validateClient(clientId);
    
    if (!client.grants.includes('device_code')) {
      throw new Error('Device authorization grant not allowed');
    }

    const deviceCode = crypto.randomBytes(40).toString('hex');
    const userCode = this.generateUserCode();
    
    await this.saveDeviceAuthorization({
      deviceCode,
      userCode,
      clientId,
      expiresAt: new Date(Date.now() + 600000), // 10 minutes
      interval: 5
    });

    return {
      device_code: deviceCode,
      user_code: userCode,
      verification_uri: `${this.baseUrl}/device`,
      verification_uri_complete: `${this.baseUrl}/device?user_code=${userCode}`,
      expires_in: 600,
      interval: 5
    };
  }
}
```

## 🆔 OpenID Connect Implementation

### 3. OIDC Provider Configuration

```typescript
// oidc-provider.config.ts
import { Provider, Configuration } from 'oidc-provider';

export class OIDCProviderConfig {
  private provider: Provider;

  constructor() {
    const configuration: Configuration = {
      clients: [{
        client_id: 'innovabiz-web',
        client_secret: process.env.WEB_CLIENT_SECRET,
        grant_types: ['authorization_code', 'refresh_token'],
        redirect_uris: ['https://app.innovabiz.com/callback'],
        response_types: ['code'],
        scope: 'openid profile email phone address',
        token_endpoint_auth_method: 'client_secret_post'
      }],
      
      cookies: {
        keys: [process.env.COOKIE_SECRET],
        long: { signed: true, secure: true, sameSite: 'none' },
        short: { signed: true, secure: true, sameSite: 'lax' }
      },

      claims: {
        address: ['address'],
        email: ['email', 'email_verified'],
        phone: ['phone_number', 'phone_number_verified'],
        profile: ['birthdate', 'family_name', 'gender', 'given_name', 
                  'locale', 'middle_name', 'name', 'nickname', 'picture', 
                  'preferred_username', 'profile', 'updated_at', 'website', 'zoneinfo']
      },

      features: {
        devInteractions: { enabled: false },
        deviceFlow: { enabled: true },
        introspection: { enabled: true },
        revocation: { enabled: true },
        rpInitiatedLogout: { enabled: true },
        backchannelLogout: { enabled: true },
        claimsParameter: { enabled: true },
        encryption: { enabled: true },
        jwtIntrospection: { enabled: true },
        jwtResponseModes: { enabled: true },
        pushedAuthorizationRequests: { enabled: true },
        registration: { enabled: true },
        requestObjects: {
          request: true,
          requestUri: true,
          requireSignedRequestObject: false
        },
        resourceIndicators: {
          enabled: true,
          getResourceServerInfo: this.getResourceServerInfo,
          defaultResource: this.defaultResource
        }
      },

      ttl: {
        AccessToken: 900, // 15 minutes
        AuthorizationCode: 600, // 10 minutes
        BackchannelAuthenticationRequest: 600,
        ClientCredentials: 600,
        DeviceCode: 600,
        Grant: 1209600, // 14 days
        IdToken: 3600, // 1 hour
        Interaction: 3600,
        RefreshToken: 86400, // 1 day
        Session: 86400
      },

      pkce: {
        methods: ['S256'],
        required: () => true
      },

      conformIdTokenClaims: true,

      renderError: async (ctx, out, error) => {
        ctx.type = 'application/json';
        ctx.body = {
          error: error.error,
          error_description: error.error_description,
          state: out.state
        };
      },

      findAccount: this.findAccount.bind(this),
      
      interactions: {
        url: (ctx, interaction) => {
          return `/interaction/${interaction.uid}`;
        }
      },

      jwks: {
        keys: [
          {
            kty: 'RSA',
            kid: process.env.JWK_KID,
            use: 'sig',
            alg: 'RS256',
            n: process.env.JWK_N,
            e: process.env.JWK_E,
            d: process.env.JWK_D,
            p: process.env.JWK_P,
            q: process.env.JWK_Q,
            dp: process.env.JWK_DP,
            dq: process.env.JWK_DQ,
            qi: process.env.JWK_QI
          }
        ]
      }
    };

    this.provider = new Provider('https://iam.innovabiz.com', configuration);
  }

  /**
   * Find account for OIDC
   */
  async findAccount(ctx: any, id: string): Promise<Account> {
    const user = await this.userRepository.findOne({
      where: { id },
      relations: ['profile', 'claims']
    });

    if (!user) {
      return undefined;
    }

    return {
      accountId: id,
      
      async claims(use, scope) {
        const claims: any = { sub: id };
        
        if (scope.includes('profile')) {
          claims.name = user.profile.name;
          claims.given_name = user.profile.givenName;
          claims.family_name = user.profile.familyName;
          claims.picture = user.profile.picture;
          claims.updated_at = user.profile.updatedAt;
        }
        
        if (scope.includes('email')) {
          claims.email = user.email;
          claims.email_verified = user.emailVerified;
        }
        
        if (scope.includes('phone')) {
          claims.phone_number = user.phoneNumber;
          claims.phone_number_verified = user.phoneNumberVerified;
        }
        
        if (scope.includes('address')) {
          claims.address = user.profile.address;
        }

        return claims;
      }
    };
  }
}
```

### 4. OIDC Endpoints

```typescript
// oidc-endpoints.controller.ts
@Controller('oidc')
export class OIDCEndpointsController {
  /**
   * Discovery Endpoint
   */
  @Get('.well-known/openid-configuration')
  async discovery(): Promise<DiscoveryDocument> {
    return {
      issuer: 'https://iam.innovabiz.com',
      authorization_endpoint: 'https://iam.innovabiz.com/oauth/authorize',
      token_endpoint: 'https://iam.innovabiz.com/oauth/token',
      userinfo_endpoint: 'https://iam.innovabiz.com/oidc/userinfo',
      jwks_uri: 'https://iam.innovabiz.com/.well-known/jwks.json',
      registration_endpoint: 'https://iam.innovabiz.com/oidc/register',
      introspection_endpoint: 'https://iam.innovabiz.com/oauth/introspect',
      revocation_endpoint: 'https://iam.innovabiz.com/oauth/revoke',
      end_session_endpoint: 'https://iam.innovabiz.com/oidc/logout',
      
      scopes_supported: [
        'openid', 'profile', 'email', 'address', 'phone', 'offline_access'
      ],
      
      response_types_supported: [
        'code', 'token', 'id_token', 'code token', 'code id_token',
        'token id_token', 'code token id_token'
      ],
      
      grant_types_supported: [
        'authorization_code', 'implicit', 'refresh_token',
        'client_credentials', 'urn:ietf:params:oauth:grant-type:device_code'
      ],
      
      acr_values_supported: ['urn:mace:incommon:iap:silver', 'urn:mace:incommon:iap:bronze'],
      
      subject_types_supported: ['public', 'pairwise'],
      
      id_token_signing_alg_values_supported: ['RS256', 'ES256'],
      id_token_encryption_alg_values_supported: ['RSA-OAEP', 'A256KW'],
      id_token_encryption_enc_values_supported: ['A128CBC-HS256', 'A256GCM'],
      
      userinfo_signing_alg_values_supported: ['RS256', 'ES256'],
      userinfo_encryption_alg_values_supported: ['RSA-OAEP', 'A256KW'],
      userinfo_encryption_enc_values_supported: ['A128CBC-HS256', 'A256GCM'],
      
      request_object_signing_alg_values_supported: ['RS256', 'ES256'],
      request_object_encryption_alg_values_supported: ['RSA-OAEP', 'A256KW'],
      request_object_encryption_enc_values_supported: ['A128CBC-HS256', 'A256GCM'],
      
      token_endpoint_auth_methods_supported: [
        'client_secret_basic', 'client_secret_post',
        'client_secret_jwt', 'private_key_jwt', 'none'
      ],
      
      token_endpoint_auth_signing_alg_values_supported: ['RS256', 'ES256'],
      
      display_values_supported: ['page', 'popup', 'touch', 'wap'],
      
      claim_types_supported: ['normal', 'aggregated', 'distributed'],
      
      claims_supported: [
        'sub', 'name', 'given_name', 'family_name', 'middle_name',
        'nickname', 'preferred_username', 'profile', 'picture',
        'website', 'email', 'email_verified', 'gender', 'birthdate',
        'zoneinfo', 'locale', 'phone_number', 'phone_number_verified',
        'address', 'updated_at'
      ],
      
      service_documentation: 'https://docs.innovabiz.com/iam',
      claims_locales_supported: ['en-US', 'pt-BR'],
      ui_locales_supported: ['en-US', 'pt-BR'],
      
      claims_parameter_supported: true,
      request_parameter_supported: true,
      request_uri_parameter_supported: true,
      require_request_uri_registration: false,
      
      op_policy_uri: 'https://innovabiz.com/privacy',
      op_tos_uri: 'https://innovabiz.com/terms',
      
      code_challenge_methods_supported: ['S256'],
      
      backchannel_logout_supported: true,
      backchannel_logout_session_supported: true,
      
      frontchannel_logout_supported: true,
      frontchannel_logout_session_supported: true,
      
      tls_client_certificate_bound_access_tokens: true,
      
      pushed_authorization_request_endpoint: 'https://iam.innovabiz.com/oauth/par'
    };
  }

  /**
   * UserInfo Endpoint
   */
  @Get('userinfo')
  @UseGuards(BearerTokenGuard)
  async userInfo(@Request() req): Promise<UserInfoResponse> {
    const token = req.token;
    const user = await this.userService.findById(token.sub);
    
    const userInfo: UserInfoResponse = {
      sub: user.id,
      name: user.name,
      given_name: user.givenName,
      family_name: user.familyName,
      picture: user.picture,
      email: user.email,
      email_verified: user.emailVerified,
      locale: user.locale,
      updated_at: user.updatedAt.toISOString()
    };

    // Add additional claims based on scope
    if (token.scope.includes('phone')) {
      userInfo.phone_number = user.phoneNumber;
      userInfo.phone_number_verified = user.phoneNumberVerified;
    }

    if (token.scope.includes('address')) {
      userInfo.address = user.address;
    }

    return userInfo;
  }
}
```

## 🔒 Security Considerations

### 5. OAuth/OIDC Security

```typescript
// oauth-security.service.ts
@Injectable()
export class OAuthSecurityService {
  /**
   * Implement security best practices
   */
  async validateSecurityRequirements(
    request: OAuthRequest
  ): Promise<ValidationResult> {
    const checks = [];

    // 1. Require HTTPS
    if (!request.secure) {
      throw new Error('HTTPS required');
    }

    // 2. Validate redirect URI
    checks.push(await this.validateRedirectUri(request.redirectUri));

    // 3. Enforce PKCE for public clients
    if (request.clientType === 'public' && !request.codeChallenge) {
      throw new Error('PKCE required for public clients');
    }

    // 4. State parameter for CSRF protection
    if (!request.state) {
      throw new Error('State parameter required');
    }

    // 5. Nonce for replay protection (OIDC)
    if (request.responseType.includes('id_token') && !request.nonce) {
      throw new Error('Nonce required for implicit flow');
    }

    // 6. Validate scope
    checks.push(await this.validateScope(request.scope));

    // 7. Rate limiting
    await this.enforceRateLimit(request.clientId);

    // 8. Token binding
    if (this.config.requireTokenBinding) {
      checks.push(await this.validateTokenBinding(request));
    }

    return {
      valid: checks.every(c => c.valid),
      checks
    };
  }

  /**
   * Implement DPoP (Demonstrating Proof-of-Possession)
   */
  async validateDPoP(request: Request): Promise<boolean> {
    const dpopHeader = request.headers['dpop'];
    
    if (!dpopHeader) {
      return false;
    }

    try {
      const dpopToken = jwt.decode(dpopHeader, { complete: true });
      
      // Validate DPoP JWT
      const validations = [
        this.validateDPoPHeader(dpopToken.header),
        this.validateDPoPPayload(dpopToken.payload, request),
        this.validateDPoPSignature(dpopToken)
      ];

      const results = await Promise.all(validations);
      return results.every(r => r === true);
    } catch (error) {
      return false;
    }
  }
}
```

## 🚀 Roadmap OAuth/OIDC

### 6. Evolução 2025

#### Q1 2025
- ✅ OAuth 2.0 Core
- ✅ OIDC Core
- 🔄 PKCE Support
- 🔄 DPoP Implementation

#### Q2 2025
- 📅 OAuth 2.1 Migration
- 📅 FAPI 2.0 Compliance
- 📅 PAR (Pushed Authorization)
- 📅 RAR (Rich Authorization)

#### Q3 2025
- 📅 Verifiable Credentials
- 📅 GNAP Support
- 📅 Distributed OAuth
- 📅 Zero-Knowledge Auth

#### Q4 2025
- 📅 Quantum-Safe OAuth
- 📅 Decentralized Identity
- 📅 Self-Issued OP
- 📅 Universal Login

---

**Mantido por**: IAM Team  
**Última Atualização**: 09/01/2025  
**Próxima Revisão**: 09/04/2025  
**Contato**: oauth@innovabiz.com
