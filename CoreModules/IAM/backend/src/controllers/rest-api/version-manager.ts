/**
 * Sistema de Versionamento de API
 * 
 * Este módulo implementa o gerenciamento de versões para APIs REST,
 * permitindo manter compatibilidade com clientes legados enquanto
 * a API evolui.
 */

import { Request, Response, NextFunction, Router } from 'express';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';

/**
 * Interface para definir um controlador de versão
 */
export interface VersionController {
  version: string;
  router: Router;
  deprecated?: boolean;
  sunset?: Date;
}

/**
 * Interface para configurações do gerenciador de versões
 */
export interface VersionManagerConfig {
  basePath: string;
  defaultVersion: string;
  allowVersionOverride: boolean;
  versionHeaderName: string;
  versionQueryParam: string;
  logger: Logger;
  metrics: Metrics;
}

/**
 * Gerenciador de versões da API
 */
export class ApiVersionManager {
  private versions: Map<string, VersionController>;
  private config: VersionManagerConfig;
  
  /**
   * Construtor do gerenciador de versões
   * 
   * @param config Configurações do gerenciador
   */
  constructor(config: VersionManagerConfig) {
    this.versions = new Map();
    this.config = {
      basePath: '/api',
      defaultVersion: 'v1',
      allowVersionOverride: true,
      versionHeaderName: 'X-API-Version',
      versionQueryParam: 'api-version',
      ...config
    };
  }
  
  /**
   * Registra uma versão da API
   * 
   * @param version Controlador de versão
   */
  public registerVersion(version: VersionController): void {
    if (this.versions.has(version.version)) {
      throw new Error(`Versão '${version.version}' já está registrada`);
    }
    
    this.versions.set(version.version, version);
    this.config.logger.info({
      message: `Versão de API '${version.version}' registrada`,
      deprecated: version.deprecated || false,
      sunset: version.sunset ? version.sunset.toISOString() : null
    });
  }
  
  /**
   * Obtém todas as versões registradas
   */
  public getVersions(): VersionController[] {
    return Array.from(this.versions.values());
  }
  
  /**
   * Marcar uma versão como obsoleta
   * 
   * @param version Versão da API
   * @param sunsetDate Data em que a versão será desativada
   */
  public deprecateVersion(version: string, sunsetDate?: Date): void {
    const versionController = this.versions.get(version);
    
    if (!versionController) {
      throw new Error(`Versão '${version}' não encontrada`);
    }
    
    versionController.deprecated = true;
    
    if (sunsetDate) {
      versionController.sunset = sunsetDate;
    }
    
    this.config.logger.info({
      message: `Versão de API '${version}' marcada como obsoleta`,
      sunset: versionController.sunset ? versionController.sunset.toISOString() : null
    });
  }
  
  /**
   * Configura o middleware de versionamento para Express
   */
  public getVersioningMiddleware(): (req: Request, res: Response, next: NextFunction) => void {
    const versionManager = this;
    const { logger, metrics } = this.config;
    
    return function(req: Request, res: Response, next: NextFunction) {
      const requestedVersion = versionManager.determineRequestedVersion(req);
      const versionController = versionManager.versions.get(requestedVersion);
      
      // Versão não encontrada
      if (!versionController) {
        logger.warn({
          message: `Versão da API '${requestedVersion}' não encontrada`,
          requestPath: req.path,
          availableVersions: Array.from(versionManager.versions.keys())
        });
        
        metrics.increment('api.version.not_found', {
          requested_version: requestedVersion
        });
        
        return res.status(404).json({
          error: 'API_VERSION_NOT_FOUND',
          message: `Versão da API '${requestedVersion}' não encontrada`,
          availableVersions: Array.from(versionManager.versions.keys())
        });
      }
      
      // Adicionar informações de versão ao request
      req.apiVersion = requestedVersion;
      
      // Adicionar headers para versões obsoletas
      if (versionController.deprecated) {
        res.setHeader('Deprecation', 'true');
        
        if (versionController.sunset) {
          res.setHeader('Sunset', versionController.sunset.toISOString());
        }
      }
      
      // Registrar métrica de uso da versão
      metrics.increment('api.version.usage', {
        version: requestedVersion,
        deprecated: versionController.deprecated ? 'true' : 'false'
      });
      
      // Prosseguir com o próximo middleware
      next();
    };
  }
  
  /**
   * Determina qual versão da API o cliente está solicitando
   * 
   * Prioridade:
   * 1. URL path (/api/v2/...)
   * 2. Header personalizado (X-API-Version: v2)
   * 3. Query parameter (api-version=v2)
   * 4. Versão padrão
   * 
   * @param req Request do Express
   */
  private determineRequestedVersion(req: Request): string {
    // Verificar versão no path
    const pathMatch = new RegExp(`^${this.config.basePath}\\/([^\\/]+)`).exec(req.path);
    if (pathMatch && this.versions.has(pathMatch[1])) {
      return pathMatch[1];
    }
    
    if (this.config.allowVersionOverride) {
      // Verificar versão no header
      const headerVersion = req.header(this.config.versionHeaderName);
      if (headerVersion && this.versions.has(headerVersion)) {
        return headerVersion;
      }
      
      // Verificar versão no query parameter
      const queryVersion = req.query[this.config.versionQueryParam] as string;
      if (queryVersion && this.versions.has(queryVersion)) {
        return queryVersion;
      }
    }
    
    // Retornar versão padrão
    return this.config.defaultVersion;
  }
  
  /**
   * Cria e configura um router para o Express
   */
  public createRouter(): Router {
    const router = Router();
    
    // Adicionar middleware para verificação de versão
    router.use(this.getVersioningMiddleware());
    
    // Rota para informações de versão
    router.get('/api/versions', (req, res) => {
      res.json({
        versions: this.getVersions().map(v => ({
          version: v.version,
          deprecated: v.deprecated || false,
          sunset: v.sunset ? v.sunset.toISOString() : null
        })),
        defaultVersion: this.config.defaultVersion,
        currentVersion: req.apiVersion
      });
    });
    
    // Configurar todas as versões registradas
    for (const [version, versionController] of this.versions.entries()) {
      const basePath = `${this.config.basePath}/${version}`;
      router.use(basePath, versionController.router);
      
      // Registrar métrica para versão disponível
      this.config.metrics.gauge('api.version.available', 1, {
        version,
        deprecated: versionController.deprecated ? 'true' : 'false'
      });
    }
    
    return router;
  }
}