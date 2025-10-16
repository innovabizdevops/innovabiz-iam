/**
 * @file update_context_trust_score.go
 * @description Comando e handler para atualização da pontuação de confiança de um contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// UpdateContextTrustScoreCommand representa o comando para atualizar a pontuação de confiança de um contexto
type UpdateContextTrustScoreCommand struct {
	ContextID     uuid.UUID // ID do contexto a ser atualizado
	TrustScore    float64   // Nova pontuação de confiança
	EvaluationSource string  // Fonte da avaliação (ex: TrustGuard, FraudDetection, KYC)
	Reason        string    // Motivo da atualização para fins de auditoria
	Factors       map[string]interface{} // Fatores que influenciaram a pontuação
	RequestedBy   string    // Utilizador ou sistema que solicitou a atualização
}

// UpdateContextTrustScoreHandler gerencia a atualização da pontuação de confiança de um contexto
type UpdateContextTrustScoreHandler struct {
	contextService *services.ContextService
	auditLogger    services.AuditLogger
}

// NewUpdateContextTrustScoreHandler cria uma nova instância do handler
func NewUpdateContextTrustScoreHandler(
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *UpdateContextTrustScoreHandler {
	return &UpdateContextTrustScoreHandler{
		contextService: contextService,
		auditLogger:    auditLogger,
	}
}

// Handle processa o comando de atualização da pontuação de confiança
func (h *UpdateContextTrustScoreHandler) Handle(ctx context.Context, cmd UpdateContextTrustScoreCommand) error {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_TRUST_SCORE_INITIATED",
		ResourceID:  cmd.ContextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      cmd.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"context_id":        cmd.ContextID,
			"trust_score":       cmd.TrustScore,
			"evaluation_source": cmd.EvaluationSource,
			"reason":            cmd.Reason,
		},
	})
	
	// Validar a pontuação de confiança (deve estar entre 0 e 1)
	if cmd.TrustScore < 0 || cmd.TrustScore > 1 {
		err := fmt.Errorf("pontuação de confiança inválida: %.2f (deve estar entre 0 e 1)", cmd.TrustScore)
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_TRUST_SCORE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"trust_score":       cmd.TrustScore,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return err
	}
	
	// Buscar o contexto atual para verificar o estado e registrar mudanças
	currentContext, err := h.contextService.GetContextByID(ctx, cmd.ContextID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_TRUST_SCORE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"trust_score":       cmd.TrustScore,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return fmt.Errorf("erro ao buscar contexto: %w", err)
	}
	
	// Verificar se o contexto está ativo
	if currentContext.Status != models.ContextStatusActive {
		err := fmt.Errorf("não é possível atualizar a pontuação de confiança de um contexto não ativo")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_TRUST_SCORE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"trust_score":       cmd.TrustScore,
				"current_status":    currentContext.Status,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return err
	}
	
	// Atualizar a pontuação de confiança no serviço de domínio
	if err := h.contextService.UpdateTrustScore(ctx, cmd.ContextID, cmd.TrustScore); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_TRUST_SCORE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"trust_score":       cmd.TrustScore,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return fmt.Errorf("erro ao atualizar pontuação de confiança: %w", err)
	}
	
	// Atualizar metadados do contexto com informações sobre a avaliação
	// Buscar o contexto novamente para obter a versão atualizada
	updatedContext, err := h.contextService.GetContextByID(ctx, cmd.ContextID)
	if err != nil {
		// Apenas registrar o erro, mas não falhar a operação já que a pontuação foi atualizada
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_TRUST_SCORE_METADATA_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"error":             err.Error(),
			},
		})
	} else {
		// Adicionar informações de avaliação aos metadados
		if updatedContext.Metadata == nil {
			updatedContext.Metadata = make(map[string]interface{})
		}
		
		trustScoreHistory := []map[string]interface{}{}
		
		// Verificar se já existe histórico de pontuações
		if history, ok := updatedContext.Metadata["trust_score_history"].([]map[string]interface{}); ok {
			trustScoreHistory = history
		}
		
		// Adicionar nova entrada ao histórico
		historyEntry := map[string]interface{}{
			"timestamp":         time.Now().UTC(),
			"trust_score":       cmd.TrustScore,
			"evaluation_source": cmd.EvaluationSource,
			"reason":            cmd.Reason,
			"requested_by":      cmd.RequestedBy,
		}
		
		// Adicionar fatores de avaliação se fornecidos
		if cmd.Factors != nil && len(cmd.Factors) > 0 {
			historyEntry["factors"] = cmd.Factors
		}
		
		trustScoreHistory = append(trustScoreHistory, historyEntry)
		
		// Limitar o tamanho do histórico (manter apenas as 20 entradas mais recentes)
		if len(trustScoreHistory) > 20 {
			trustScoreHistory = trustScoreHistory[len(trustScoreHistory)-20:]
		}
		
		updatedContext.Metadata["trust_score_history"] = trustScoreHistory
		updatedContext.Metadata["last_evaluation_source"] = cmd.EvaluationSource
		updatedContext.Metadata["last_evaluation_timestamp"] = time.Now().UTC()
		
		// Se a pontuação é baixa e representa uma degradação significativa, marcar como suspeito
		if cmd.TrustScore < 0.4 && (currentContext.TrustScore - cmd.TrustScore) > 0.2 {
			updatedContext.Metadata["risk_flags"] = append(
				updatedContext.Metadata["risk_flags"].([]string), 
				"significant_trust_degradation",
			)
		}
		
		// Atualizar o contexto com os novos metadados
		if err := h.contextService.UpdateContext(ctx, updatedContext); err != nil {
			// Apenas registrar o erro, mas não falhar a operação já que a pontuação foi atualizada
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "UPDATE_TRUST_SCORE_METADATA_FAILED",
				ResourceID:  cmd.ContextID.String(),
				ResourceType: "IDENTITY_CONTEXT",
				UserID:      cmd.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id": operationID,
					"context_id":   cmd.ContextID,
					"error":        err.Error(),
				},
			})
		}
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_TRUST_SCORE_SUCCEEDED",
		ResourceID:  cmd.ContextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      cmd.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":        operationID,
			"context_id":          cmd.ContextID,
			"previous_trust_score": currentContext.TrustScore,
			"new_trust_score":     cmd.TrustScore,
			"evaluation_source":   cmd.EvaluationSource,
			"reason":              cmd.Reason,
			"duration_ms":         time.Since(startTime).Milliseconds(),
		},
	})
	
	return nil
}