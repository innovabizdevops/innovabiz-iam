/**
 * Resolvers GraphQL para Dashboard de Conformidade
 * 
 * Este módulo fornece os resolvers GraphQL que permitem que o front-end
 * acesse os dados do dashboard de conformidade multi-regulatória.
 * 
 * @module ComplianceDashboardResolvers
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

import { 
  DashboardDataCollector,
  ComplianceAggregateData,
  ValidationDetailRecord
} from '../compliance/dashboard-data-collector';

import { ComplianceValidatorIntegrator } from '../compliance/compliance-validator-integrator';

/**
 * Interface para parâmetros do resolver de dashboard
 */
interface DashboardQueryParams {
  tenantId: string;
  timeframeHours?: number;
  includeDetails?: boolean;
}

/**
 * Interface para parâmetros de busca de validações detalhadas
 */
interface ValidationDetailsQueryParams {
  tenantId: string;
  timeframeHours?: number;
  regulationFilter?: string;
  compliantOnly?: boolean;
  nonCompliantOnly?: boolean;
  severityFilter?: string;
  limit?: number;
  offset?: number;
}

/**
 * Classe de resolvers GraphQL para dashboard de conformidade
 */
export class ComplianceDashboardResolvers {
  private dashboardCollector: DashboardDataCollector;
  
  constructor(
    private logger: Logger,
    private metrics: Metrics,
    private tracer: Tracer
  ) {
    const complianceIntegrator = new ComplianceValidatorIntegrator(logger, metrics, tracer);
    this.dashboardCollector = new DashboardDataCollector(
      logger, 
      metrics, 
      tracer,
      complianceIntegrator
    );
  }
  
  /**
   * Define os resolvers GraphQL para o dashboard de conformidade
   */
  public getResolvers() {
    return {
      Query: {
        complianceDashboard: this.getComplianceDashboard.bind(this),
        validationDetails: this.getValidationDetails.bind(this),
        regulationCompliance: this.getRegulationCompliance.bind(this),
        complianceTrends: this.getComplianceTrends.bind(this)
      }
    };
  }
  
  /**
   * Resolver para obter dados agregados do dashboard
   * 
   * @param _ Parent object (não utilizado)
   * @param args Argumentos da query GraphQL
   * @param context Contexto da requisição GraphQL
   * @returns Dados agregados do dashboard
   */
  private async getComplianceDashboard(_: any, args: DashboardQueryParams, context: any) {
    const { tenantId, timeframeHours = 24 } = args;
    const requestId = context.requestId || `req-${Date.now()}`;
    
    const span = this.tracer.startSpan('compliance_dashboard.get_dashboard_data', {
      attributes: {
        'tenant.id': tenantId,
        'timeframe.hours': timeframeHours,
        'request.id': requestId
      }
    });
    
    try {
      this.logger.info({
        message: 'Recebida requisição para dados do dashboard de conformidade',
        requestId,
        tenantId,
        timeframeHours
      });
      
      // Autenticação e autorização (normalmente implementadas no middleware GraphQL)
      this.validateAccess(context, tenantId, 'dashboard:compliance:read');
      
      // Obter dados agregados para o tenant
      const dashboardData = await this.dashboardCollector.generateComplianceAggregateData(
        tenantId,
        timeframeHours
      );
      
      this.metrics.increment('compliance_dashboard.graphql_queries', {
        tenant_id: tenantId,
        operation: 'complianceDashboard',
        result: 'success'
      });
      
      this.logger.debug({
        message: 'Dados do dashboard de conformidade retornados com sucesso',
        requestId,
        tenantId,
        totalOperations: dashboardData.overallStats.totalOperations
      });
      
      return dashboardData;
    } catch (error) {
      this.metrics.increment('compliance_dashboard.graphql_errors', {
        tenant_id: tenantId,
        operation: 'complianceDashboard',
        error_type: error.name || 'unknown'
      });
      
      this.logger.error({
        message: 'Erro ao obter dados do dashboard de conformidade',
        error: error.message,
        stack: error.stack,
        requestId,
        tenantId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Resolver para obter detalhes de validações específicas
   * 
   * @param _ Parent object (não utilizado)
   * @param args Argumentos da query GraphQL
   * @param context Contexto da requisição GraphQL
   * @returns Lista detalhada de validações
   */
  private async getValidationDetails(_: any, args: ValidationDetailsQueryParams, context: any) {
    const { 
      tenantId, 
      timeframeHours = 24,
      regulationFilter,
      compliantOnly = false,
      nonCompliantOnly = false,
      severityFilter,
      limit = 50,
      offset = 0
    } = args;
    
    const requestId = context.requestId || `req-${Date.now()}`;
    
    const span = this.tracer.startSpan('compliance_dashboard.get_validation_details', {
      attributes: {
        'tenant.id': tenantId,
        'request.id': requestId
      }
    });
    
    try {
      this.logger.info({
        message: 'Recebida requisição para detalhes de validação',
        requestId,
        tenantId,
        timeframeHours,
        regulationFilter,
        compliantOnly,
        nonCompliantOnly
      });
      
      // Autenticação e autorização
      this.validateAccess(context, tenantId, 'dashboard:compliance:read:details');
      
      // Esta é uma implementação simplificada
      // Na prática, você buscaria estes dados de um banco de dados
      // ou outro sistema de armazenamento persistente
      const validationDetails: ValidationDetailRecord[] = [];
      
      // Como este é um exemplo, retornamos apenas um mock
      // Em uma implementação real, você buscaria os dados do banco
      return {
        validations: validationDetails,
        totalCount: validationDetails.length,
        hasMoreResults: false
      };
    } catch (error) {
      this.metrics.increment('compliance_dashboard.graphql_errors', {
        tenant_id: tenantId,
        operation: 'validationDetails',
        error_type: error.name || 'unknown'
      });
      
      this.logger.error({
        message: 'Erro ao obter detalhes de validação',
        error: error.message,
        stack: error.stack,
        requestId,
        tenantId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Resolver para obter conformidade específica por regulamentação
   * 
   * @param _ Parent object (não utilizado)
   * @param args Argumentos da query GraphQL
   * @param context Contexto da requisição GraphQL
   * @returns Dados de conformidade para a regulamentação
   */
  private async getRegulationCompliance(_: any, args: { tenantId: string, regulation: string, timeframeHours?: number }, context: any) {
    const { tenantId, regulation, timeframeHours = 24 } = args;
    const requestId = context.requestId || `req-${Date.now()}`;
    
    const span = this.tracer.startSpan('compliance_dashboard.get_regulation_compliance', {
      attributes: {
        'tenant.id': tenantId,
        'regulation': regulation,
        'request.id': requestId
      }
    });
    
    try {
      this.logger.info({
        message: 'Recebida requisição para conformidade por regulamentação',
        requestId,
        tenantId,
        regulation,
        timeframeHours
      });
      
      // Autenticação e autorização
      this.validateAccess(context, tenantId, 'dashboard:compliance:read');
      
      // Obter dados agregados para o tenant
      const dashboardData = await this.dashboardCollector.generateComplianceAggregateData(
        tenantId,
        timeframeHours
      );
      
      // Filtrar apenas dados da regulamentação solicitada
      const regulationData = dashboardData.regulationStats[regulation] || {
        applicableOperations: 0,
        compliantOperations: 0,
        nonCompliantOperations: 0,
        complianceRate: 0
      };
      
      // Filtrar problemas relacionados a esta regulamentação
      const regulationIssues = dashboardData.topIssues
        .filter(issue => issue.regulation === regulation)
        .map(issue => ({
          issueType: issue.issueType,
          occurrences: issue.occurrences,
          severity: issue.severity
        }));
      
      return {
        regulation,
        stats: regulationData,
        topIssues: regulationIssues,
        // Outros dados relevantes para esta regulamentação
      };
    } catch (error) {
      this.metrics.increment('compliance_dashboard.graphql_errors', {
        tenant_id: tenantId,
        operation: 'regulationCompliance',
        error_type: error.name || 'unknown'
      });
      
      this.logger.error({
        message: 'Erro ao obter dados de conformidade por regulamentação',
        error: error.message,
        stack: error.stack,
        requestId,
        tenantId,
        regulation
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Resolver para obter tendências de conformidade ao longo do tempo
   * 
   * @param _ Parent object (não utilizado)
   * @param args Argumentos da query GraphQL
   * @param context Contexto da requisição GraphQL
   * @returns Dados de tendência de conformidade
   */
  private async getComplianceTrends(_: any, args: { tenantId: string, timeframeHours?: number }, context: any) {
    const { tenantId, timeframeHours = 168 } = args; // 168 = 7 dias
    const requestId = context.requestId || `req-${Date.now()}`;
    
    const span = this.tracer.startSpan('compliance_dashboard.get_compliance_trends', {
      attributes: {
        'tenant.id': tenantId,
        'timeframe.hours': timeframeHours,
        'request.id': requestId
      }
    });
    
    try {
      this.logger.info({
        message: 'Recebida requisição para tendências de conformidade',
        requestId,
        tenantId,
        timeframeHours
      });
      
      // Autenticação e autorização
      this.validateAccess(context, tenantId, 'dashboard:compliance:read');
      
      // Obter dados agregados para o tenant
      const dashboardData = await this.dashboardCollector.generateComplianceAggregateData(
        tenantId,
        timeframeHours
      );
      
      return dashboardData.trendsData;
    } catch (error) {
      this.metrics.increment('compliance_dashboard.graphql_errors', {
        tenant_id: tenantId,
        operation: 'complianceTrends',
        error_type: error.name || 'unknown'
      });
      
      this.logger.error({
        message: 'Erro ao obter tendências de conformidade',
        error: error.message,
        stack: error.stack,
        requestId,
        tenantId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Valida se o usuário tem acesso ao recurso solicitado
   * 
   * @param context Contexto da requisição GraphQL
   * @param tenantId ID do tenant sendo acessado
   * @param permission Permissão necessária
   * @throws Error se o acesso for negado
   */
  private validateAccess(context: any, tenantId: string, permission: string): void {
    // Implementação simplificada - em um sistema real, você verificaria
    // as permissões do usuário no contexto da requisição GraphQL
    
    if (!context.user) {
      throw new Error('Usuário não autenticado');
    }
    
    const userTenantId = context.user.tenantId;
    const userPermissions = context.user.permissions || [];
    
    // Verifica se o usuário pertence ao tenant
    if (userTenantId !== tenantId && !context.user.isSystemAdmin) {
      throw new Error('Acesso negado: usuário não pertence ao tenant solicitado');
    }
    
    // Verifica se o usuário tem a permissão necessária
    const hasPermission = userPermissions.some(p => 
      p === permission || p === '*' || p === 'dashboard:*' || p === 'dashboard:compliance:*'
    );
    
    if (!hasPermission && !context.user.isSystemAdmin) {
      throw new Error(`Acesso negado: permissão '${permission}' necessária`);
    }
  }
}