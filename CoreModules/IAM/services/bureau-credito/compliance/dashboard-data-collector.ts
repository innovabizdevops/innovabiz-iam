/**
 * Coletor de Dados para Dashboard de Conformidade Multi-Regulatória
 * 
 * Este módulo é responsável por coletar e agregar dados de conformidade
 * para alimentar o dashboard de conformidade multi-regulatória.
 * 
 * @module DashboardDataCollector
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

import { 
  ComplianceValidatorIntegrator,
  Region,
  ConsolidatedComplianceResult
} from './compliance-validator-integrator';

/**
 * Interface para dados agregados de conformidade
 */
export interface ComplianceAggregateData {
  tenantId: string;
  timestamp: number;
  overallStats: {
    totalOperations: number;
    compliantOperations: number;
    nonCompliantOperations: number;
    complianceRate: number;
    blockedOperations: number;
    restrictedOperations: number;
  };
  regulationStats: {
    [key: string]: {
      applicableOperations: number;
      compliantOperations: number;
      nonCompliantOperations: number;
      complianceRate: number;
    }
  };
  riskDistribution: {
    highRiskOperations: number;
    mediumRiskOperations: number;
    lowRiskOperations: number;
  };
  topIssues: {
    issueType: string;
    regulation: string;
    occurrences: number;
    severity: 'high' | 'medium' | 'low';
  }[];
  trendsData: {
    timeframe: string;
    complianceRate: number;
    operationCount: number;
  }[];
}

/**
 * Interface para dados detalhados de validação
 */
export interface ValidationDetailRecord {
  validationId: string;
  tenantId: string;
  userId: string;
  timestamp: number;
  operationId: string;
  operationType: string;
  dataSubjectCountry: string;
  dataProcessingCountry: string;
  applicableRegulations: string[];
  overallCompliant: boolean;
  processingAllowed: boolean;
  processingRestrictions?: string[];
  validationDetails: {
    regulation: string;
    validationType: string;
    valid: boolean;
    severity: 'high' | 'medium' | 'low';
    details?: string;
    errors?: string[];
    warnings?: string[];
  }[];
}

/**
 * Classe para coleta e agregação de dados para dashboard de conformidade
 */
export class DashboardDataCollector {
  // Cache local para armazenamento temporário de validações recentes
  private validationRecords: ValidationDetailRecord[] = [];
  private maxCacheSize = 1000;
  
  constructor(
    private logger: Logger,
    private metrics: Metrics,
    private tracer: Tracer,
    private complianceIntegrator: ComplianceValidatorIntegrator
  ) {}
  
  /**
   * Registra um novo resultado de validação para análise posterior
   * 
   * @param validationResult Resultado da validação de conformidade
   * @param context Contexto da validação
   */
  public recordValidationResult(
    validationResult: ConsolidatedComplianceResult,
    context: {
      tenantId: string;
      userId: string;
      operationId: string;
      operationType: string;
      dataSubjectCountry: Region | string;
      dataProcessingCountry: string;
    }
  ): void {
    const validationRecord: ValidationDetailRecord = {
      validationId: `val-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      tenantId: context.tenantId,
      userId: context.userId,
      timestamp: Date.now(),
      operationId: context.operationId,
      operationType: context.operationType,
      dataSubjectCountry: context.dataSubjectCountry.toString(),
      dataProcessingCountry: context.dataProcessingCountry,
      applicableRegulations: validationResult.applicableRegulations,
      overallCompliant: validationResult.overallCompliant,
      processingAllowed: validationResult.processingAllowed,
      processingRestrictions: validationResult.processingRestrictions,
      validationDetails: validationResult.validationDetails?.map(detail => ({
        regulation: detail.regulation,
        validationType: detail.validationType,
        valid: detail.valid,
        severity: detail.severity || 'medium',
        details: detail.details,
        errors: detail.errors,
        warnings: detail.warnings
      })) || []
    };
    
    // Adiciona ao cache local
    this.validationRecords.push(validationRecord);
    
    // Limita tamanho do cache
    if (this.validationRecords.length > this.maxCacheSize) {
      this.validationRecords = this.validationRecords.slice(-this.maxCacheSize);
    }
    
    // Registra métricas para monitoramento em tempo real
    this.recordMetrics(validationRecord);
    
    this.logger.debug({
      message: 'Registro de validação adicionado para dashboard',
      validationId: validationRecord.validationId,
      tenantId: context.tenantId,
      operationType: context.operationType
    });
  }
  
  /**
   * Gera dados agregados de conformidade para o dashboard
   * 
   * @param tenantId ID do tenant para filtrar os dados
   * @param timeframeHours Janela de tempo em horas para análise
   * @returns Dados agregados de conformidade
   */
  public async generateComplianceAggregateData(
    tenantId: string,
    timeframeHours: number = 24
  ): Promise<ComplianceAggregateData> {
    const span = this.tracer.startSpan('dashboard_data_collector.generate_aggregate_data', {
      attributes: {
        'tenant.id': tenantId,
        'timeframe.hours': timeframeHours
      }
    });
    
    try {
      this.logger.info({
        message: 'Gerando dados agregados para dashboard de conformidade',
        tenantId,
        timeframeHours
      });
      
      const startTime = Date.now() - (timeframeHours * 60 * 60 * 1000);
      
      // Filtra registros pelo tenant e janela de tempo
      const relevantRecords = this.validationRecords.filter(record => 
        record.tenantId === tenantId && record.timestamp >= startTime
      );
      
      // Estatísticas gerais
      const totalOperations = relevantRecords.length;
      const compliantOperations = relevantRecords.filter(r => r.overallCompliant).length;
      const nonCompliantOperations = totalOperations - compliantOperations;
      const blockedOperations = relevantRecords.filter(r => !r.processingAllowed).length;
      const restrictedOperations = relevantRecords.filter(r => 
        r.processingAllowed && r.processingRestrictions && r.processingRestrictions.length > 0
      ).length;
      
      // Estatísticas por regulamentação
      const regulationStats: {[key: string]: any} = {};
      
      // Coleta todas as regulamentações únicas
      const allRegulations = Array.from(new Set(
        relevantRecords.flatMap(r => r.applicableRegulations)
      ));
      
      // Calcula estatísticas para cada regulamentação
      allRegulations.forEach(regulation => {
        const applicableRecords = relevantRecords.filter(r => 
          r.applicableRegulations.includes(regulation)
        );
        
        const applicableCount = applicableRecords.length;
        const compliantCount = applicableRecords.filter(r => {
          const regDetails = r.validationDetails.filter(d => d.regulation === regulation);
          return regDetails.every(d => d.valid);
        }).length;
        
        regulationStats[regulation] = {
          applicableOperations: applicableCount,
          compliantOperations: compliantCount,
          nonCompliantOperations: applicableCount - compliantCount,
          complianceRate: applicableCount > 0 ? (compliantCount / applicableCount) * 100 : 100
        };
      });
      
      // Distribuição de risco
      const highRiskOperations = relevantRecords.filter(r => 
        r.validationDetails.some(d => !d.valid && d.severity === 'high')
      ).length;
      
      const mediumRiskOperations = relevantRecords.filter(r => 
        r.validationDetails.some(d => !d.valid && d.severity === 'medium') &&
        !r.validationDetails.some(d => !d.valid && d.severity === 'high')
      ).length;
      
      const lowRiskOperations = relevantRecords.filter(r => 
        r.validationDetails.some(d => !d.valid && d.severity === 'low') &&
        !r.validationDetails.some(d => !d.valid && (d.severity === 'high' || d.severity === 'medium'))
      ).length;
      
      // Identificar principais problemas
      const issueCounter: Map<string, { 
        count: number; 
        regulation: string; 
        severity: 'high' | 'medium' | 'low';
      }> = new Map();
      
      relevantRecords.forEach(record => {
        record.validationDetails
          .filter(detail => !detail.valid)
          .forEach(detail => {
            const issueKey = `${detail.regulation}:${detail.validationType}`;
            const current = issueCounter.get(issueKey) || { 
              count: 0, 
              regulation: detail.regulation, 
              severity: detail.severity 
            };
            current.count++;
            issueCounter.set(issueKey, current);
          });
      });
      
      // Ordenar problemas por ocorrências e pegar os 10 principais
      const topIssues = Array.from(issueCounter.entries())
        .map(([issueType, data]) => ({
          issueType: issueType.split(':')[1], // Remove prefixo da regulamentação
          regulation: data.regulation,
          occurrences: data.count,
          severity: data.severity
        }))
        .sort((a, b) => b.occurrences - a.occurrences)
        .slice(0, 10);
      
      // Dados de tendência (dividindo o período em 6 segmentos)
      const segmentCount = 6;
      const segmentDuration = (timeframeHours * 60 * 60 * 1000) / segmentCount;
      const trendsData = [];
      
      for (let i = 0; i < segmentCount; i++) {
        const segmentStart = startTime + (i * segmentDuration);
        const segmentEnd = segmentStart + segmentDuration;
        
        const segmentRecords = relevantRecords.filter(r => 
          r.timestamp >= segmentStart && r.timestamp < segmentEnd
        );
        
        const segmentTotal = segmentRecords.length;
        const segmentCompliant = segmentRecords.filter(r => r.overallCompliant).length;
        
        trendsData.push({
          timeframe: new Date(segmentStart).toISOString(),
          complianceRate: segmentTotal > 0 ? (segmentCompliant / segmentTotal) * 100 : 100,
          operationCount: segmentTotal
        });
      }
      
      // Compila o resultado final
      const aggregateData: ComplianceAggregateData = {
        tenantId,
        timestamp: Date.now(),
        overallStats: {
          totalOperations,
          compliantOperations,
          nonCompliantOperations,
          complianceRate: totalOperations > 0 ? (compliantOperations / totalOperations) * 100 : 100,
          blockedOperations,
          restrictedOperations
        },
        regulationStats,
        riskDistribution: {
          highRiskOperations,
          mediumRiskOperations,
          lowRiskOperations
        },
        topIssues,
        trendsData
      };
      
      this.logger.info({
        message: 'Dados agregados para dashboard gerados com sucesso',
        tenantId,
        totalOperations,
        complianceRate: aggregateData.overallStats.complianceRate.toFixed(2)
      });
      
      return aggregateData;
    } catch (error) {
      this.logger.error({
        message: 'Erro ao gerar dados agregados para dashboard',
        error: error.message,
        stack: error.stack,
        tenantId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Registra métricas para monitoramento em tempo real
   * 
   * @param record Registro de validação
   */
  private recordMetrics(record: ValidationDetailRecord): void {
    // Métricas de operações por regulamentação
    record.applicableRegulations.forEach(regulation => {
      this.metrics.increment('compliance_dashboard.operations_by_regulation', {
        tenant_id: record.tenantId,
        regulation,
        compliant: record.overallCompliant.toString(),
        operation_type: record.operationType
      });
    });
    
    // Métricas de tipos de problemas
    record.validationDetails
      .filter(detail => !detail.valid)
      .forEach(detail => {
        this.metrics.increment('compliance_dashboard.validation_issues', {
          tenant_id: record.tenantId,
          regulation: detail.regulation,
          validation_type: detail.validationType,
          severity: detail.severity
        });
      });
    
    // Métricas gerais de conformidade
    this.metrics.increment('compliance_dashboard.validation_results', {
      tenant_id: record.tenantId,
      compliant: record.overallCompliant.toString(),
      processing_allowed: record.processingAllowed.toString(),
      has_restrictions: (record.processingRestrictions && record.processingRestrictions.length > 0).toString()
    });
  }
}