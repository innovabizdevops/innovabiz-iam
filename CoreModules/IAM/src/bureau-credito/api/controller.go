/**
 * @file controller.go
 * @description Controlador REST para API do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/google/uuid"
	
	"innovabiz/iam/src/bureau-credito/orchestration"
	"innovabiz/iam/src/bureau-credito/orchestration/models"
)

// BureauController gerencia as operações da API REST do Bureau de Crédito
type BureauController struct {
	orchestrator     *orchestration.BureauOrchestrator
	validate         *validator.Validate
}

// NewBureauController cria um novo controlador para a API
func NewBureauController(orchestrator *orchestration.BureauOrchestrator) *BureauController {
	return &BureauController{
		orchestrator:     orchestrator,
		validate:         validator.New(),
	}
}

// RegisterRoutes registra as rotas REST do controlador no roteador
func (c *BureauController) RegisterRoutes(router *mux.Router) {
	// Rotas para avaliações
	router.HandleFunc("/assessments", c.CreateAssessment).Methods("POST")
	router.HandleFunc("/assessments/batch", c.CreateBatchAssessment).Methods("POST")
	router.HandleFunc("/assessments/{id}", c.GetAssessment).Methods("GET")
	router.HandleFunc("/assessments/{id}/status", c.GetAssessmentStatus).Methods("GET")
	
	// Rotas para health check e informações
	router.HandleFunc("/health", c.HealthCheck).Methods("GET")
	router.HandleFunc("/info", c.GetInfo).Methods("GET")
	
	log.Info().Msg("Rotas da API do Bureau de Crédito registradas com sucesso")
}

// CreateAssessment processa uma solicitação de avaliação
func (c *BureauController) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Decodificar solicitação
	var requestDTO AssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Erro ao decodificar solicitação", err)
		return
	}
	
	// Validar solicitação
	if err := c.validate.Struct(requestDTO); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Dados de solicitação inválidos", err)
		return
	}
	
	// Criar ID de correlação se não fornecido
	if requestDTO.CorrelationID == "" {
		requestDTO.CorrelationID = uuid.New().String()
	}
	
	// Log da solicitação recebida
	log.Info().
		Str("userId", requestDTO.UserID).
		Str("tenantId", requestDTO.TenantID).
		Str("correlationId", requestDTO.CorrelationID).
		Strs("assessmentTypes", requestDTO.AssessmentTypes).
		Msg("Solicitação de avaliação recebida")
	
	// Converter para modelo interno
	internalRequest := requestDTO.ToInternalModel()
	
	// Processar avaliação
	startTime := time.Now()
	response, err := c.orchestrator.RequestAssessment(ctx, internalRequest)
	processingTime := time.Since(startTime).Milliseconds()
	
	// Verificar se ocorreu erro ao processar
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", requestDTO.UserID).
			Str("correlationId", requestDTO.CorrelationID).
			Int64("processingTimeMs", processingTime).
			Msg("Erro ao processar solicitação de avaliação")
			
		// Responder com erro
		c.respondWithError(w, http.StatusInternalServerError, "Erro ao processar solicitação de avaliação", err)
		return
	}
	
	// Converter resposta para o modelo de API
	apiResponse := FromInternalModel(response)
	apiResponse.ProcessingTimeMs = processingTime
	
	// Registrar sucesso
	log.Info().
		Str("userId", requestDTO.UserID).
		Str("correlationId", requestDTO.CorrelationID).
		Str("requestId", response.RequestID).
		Str("responseId", response.ResponseID).
		Int64("processingTimeMs", processingTime).
		Int("trustScore", response.TrustScore).
		Str("riskLevel", response.RiskLevel).
		Str("decision", response.Decision).
		Msg("Avaliação processada com sucesso")
	
	// Retornar resposta
	c.respondWithJSON(w, http.StatusOK, apiResponse)
}

// CreateBatchAssessment processa uma solicitação em lote de avaliações
func (c *BureauController) CreateBatchAssessment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Decodificar solicitação
	var batchRequest BatchAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&batchRequest); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Erro ao decodificar solicitação em lote", err)
		return
	}
	
	// Validar solicitação
	if err := c.validate.Struct(batchRequest); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Dados de solicitação em lote inválidos", err)
		return
	}
	
	// Log da solicitação em lote
	log.Info().
		Int("totalRequests", len(batchRequest.Requests)).
		Msg("Solicitação em lote de avaliações recebida")
	
	// Processar cada solicitação
	responses := make([]AssessmentResponse, 0, len(batchRequest.Requests))
	success := 0
	failed := 0
	
	for _, req := range batchRequest.Requests {
		// Criar ID de correlação se não fornecido
		if req.CorrelationID == "" {
			req.CorrelationID = uuid.New().String()
		}
		
		// Converter para modelo interno
		internalReq := req.ToInternalModel()
		
		// Processar avaliação
		internalResp, err := c.orchestrator.RequestAssessment(ctx, internalReq)
		
		// Verificar se ocorreu erro ao processar
		if err != nil {
			failed++
			log.Error().
				Err(err).
				Str("userId", req.UserID).
				Str("correlationId", req.CorrelationID).
				Msg("Erro ao processar solicitação em lote")
				
			// Criar resposta de erro
			errorResp := &AssessmentResponse{
				ResponseID:    uuid.New().String(),
				UserID:        req.UserID,
				TenantID:      req.TenantID,
				CorrelationID: req.CorrelationID,
				Status:        "FAILED",
				ErrorDetails: &ErrorDetails{
					ErrorCode:    "PROCESSING_ERROR",
					ErrorMessage: err.Error(),
					PartialResults: false,
				},
			}
			responses = append(responses, *errorResp)
			continue
		}
		
		success++
		// Converter resposta para o modelo de API
		apiResp := FromInternalModel(internalResp)
		responses = append(responses, *apiResp)
	}
	
	// Criar resposta do lote
	batchResponse := &BatchAssessmentResponse{
		Responses: responses,
		Success:   success,
		Failed:    failed,
		Total:     len(batchRequest.Requests),
	}
	
	// Registrar resultado do lote
	log.Info().
		Int("totalRequests", len(batchRequest.Requests)).
		Int("success", success).
		Int("failed", failed).
		Msg("Processamento em lote concluído")
	
	// Retornar resposta
	c.respondWithJSON(w, http.StatusOK, batchResponse)
}

// GetAssessment recupera uma avaliação específica pelo ID
func (c *BureauController) GetAssessment(w http.ResponseWriter, r *http.Request) {
	// Obter parâmetros da URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.respondWithError(w, http.StatusBadRequest, "ID de avaliação não fornecido", nil)
		return
	}
	
	// Obter avaliação
	response, err := c.orchestrator.GetAssessment(r.Context(), id)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Erro ao recuperar avaliação", err)
		return
	}
	
	// Verificar se avaliação existe
	if response == nil {
		c.respondWithError(w, http.StatusNotFound, "Avaliação não encontrada", nil)
		return
	}
	
	// Converter para modelo de API
	apiResponse := FromInternalModel(response)
	
	// Retornar resposta
	c.respondWithJSON(w, http.StatusOK, apiResponse)
}

// GetAssessmentStatus recupera o status de uma avaliação específica
func (c *BureauController) GetAssessmentStatus(w http.ResponseWriter, r *http.Request) {
	// Obter parâmetros da URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.respondWithError(w, http.StatusBadRequest, "ID de avaliação não fornecido", nil)
		return
	}
	
	// Obter status da avaliação
	status, err := c.orchestrator.GetAssessmentStatus(r.Context(), id)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Erro ao recuperar status da avaliação", err)
		return
	}
	
	// Verificar se avaliação existe
	if status == "" {
		c.respondWithError(w, http.StatusNotFound, "Avaliação não encontrada", nil)
		return
	}
	
	// Criar resposta de status
	statusResponse := &AssessmentStatusResponse{
		RequestID: id,
		Status:    string(status),
	}
	
	// Retornar resposta
	c.respondWithJSON(w, http.StatusOK, statusResponse)
}

// HealthCheck verifica se o serviço está operacional
func (c *BureauController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Verificar se o orquestrador está operacional
	status := c.orchestrator.IsHealthy()
	
	// Criar resposta de saúde
	healthResponse := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Status:    "UP",
		Timestamp: time.Now(),
	}
	
	if !status {
		healthResponse.Status = "DEGRADED"
	}
	
	// Retornar resposta
	if status {
		c.respondWithJSON(w, http.StatusOK, healthResponse)
	} else {
		c.respondWithJSON(w, http.StatusServiceUnavailable, healthResponse)
	}
}

// GetInfo retorna informações sobre o serviço
func (c *BureauController) GetInfo(w http.ResponseWriter, r *http.Request) {
	// Criar resposta de informações
	infoResponse := struct {
		ServiceName    string   `json:"serviceName"`
		Version        string   `json:"version"`
		BuildTimestamp string   `json:"buildTimestamp"`
		Features       []string `json:"features"`
	}{
		ServiceName:    "InnovaBiz Bureau de Crédito API",
		Version:        "1.0.0",
		BuildTimestamp: "2025-01-15T10:00:00Z",
		Features: []string{
			"Avaliação de identidade",
			"Avaliação de crédito",
			"Detecção de fraudes",
			"Verificação de conformidade",
			"Avaliação de risco",
			"Avaliação abrangente",
		},
	}
	
	// Retornar resposta
	c.respondWithJSON(w, http.StatusOK, infoResponse)
}

// respondWithJSON envia uma resposta JSON
func (c *BureauController) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	// Definir cabeçalho e status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	// Codificar resposta
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Error().Err(err).Msg("Erro ao codificar resposta JSON")
		}
	}
}

// respondWithError envia uma resposta de erro
func (c *BureauController) respondWithError(w http.ResponseWriter, status int, message string, err error) {
	// Criar resposta de erro
	errorResponse := struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}{
		Status:  status,
		Message: message,
	}
	
	// Adicionar detalhes do erro, se fornecido
	if err != nil {
		errorResponse.Error = err.Error()
	}
	
	// Log do erro
	log.Error().
		Int("status", status).
		Str("message", message).
		AnErr("error", err).
		Msg("Erro na API do Bureau de Crédito")
	
	// Enviar resposta
	c.respondWithJSON(w, status, errorResponse)
}