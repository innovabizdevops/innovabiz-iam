package handlers

import (
	"encoding/json"
	"net/http"
	
	"github.com/rs/zerolog/log"
	
	"innovabiz/iam/identity-service/internal/application"
)

// ErrorResponse representa uma resposta de erro padronizada da API
type ErrorResponse struct {
	Status     int    `json:"-"` // Status HTTP não é serializado
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	InstanceID string `json:"instance_id,omitempty"` // ID de rastreamento para logs
}

// respondWithJSON serializa uma resposta para JSON e a envia
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar resposta para JSON")
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(response)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao escrever resposta HTTP")
	}
}

// respondWithError envia uma resposta de erro padronizada
func respondWithError(w http.ResponseWriter, status int, message string) {
	errResponse := ErrorResponse{
		Status:  status,
		Code:    getErrorCodeFromStatus(status),
		Message: message,
	}
	
	respondWithJSON(w, status, errResponse)
}

// handleRoleServiceError mapeia erros do serviço para respostas HTTP apropriadas
func handleRoleServiceError(w http.ResponseWriter, err error) {
	// Mapear erros específicos do domínio para status HTTP
	switch err {
	case application.ErrRoleNotFound:
		respondWithError(w, http.StatusNotFound, "Função não encontrada")
	case application.ErrParentRoleNotFound:
		respondWithError(w, http.StatusNotFound, "Função pai não encontrada")
	case application.ErrChildRoleNotFound:
		respondWithError(w, http.StatusNotFound, "Função filha não encontrada")
	case application.ErrPermissionNotFound:
		respondWithError(w, http.StatusNotFound, "Permissão não encontrada")
	case application.ErrRoleCodeAlreadyExists:
		respondWithError(w, http.StatusConflict, "Já existe uma função com este código")
	case application.ErrPermissionAlreadyAssigned:
		respondWithError(w, http.StatusConflict, "Esta permissão já está atribuída à função")
	case application.ErrPermissionNotAssigned:
		respondWithError(w, http.StatusBadRequest, "Esta permissão não está atribuída à função")
	case application.ErrUserAlreadyAssigned:
		respondWithError(w, http.StatusConflict, "Este usuário já está atribuído à função")
	case application.ErrUserNotAssigned:
		respondWithError(w, http.StatusBadRequest, "Este usuário não está atribuído à função")
	case application.ErrChildRoleAlreadyAssigned:
		respondWithError(w, http.StatusConflict, "Esta função filha já está atribuída à função pai")
	case application.ErrChildRoleNotAssigned:
		respondWithError(w, http.StatusBadRequest, "Esta função filha não está atribuída à função pai")
	case application.ErrCyclicRoleHierarchy:
		respondWithError(w, http.StatusBadRequest, "Esta atribuição criaria um ciclo na hierarquia de funções")
	case application.ErrRolesTypeMismatch:
		respondWithError(w, http.StatusBadRequest, "Funções de tipos diferentes não podem ser relacionadas hierarquicamente")
	case application.ErrCannotDeleteSystemRole:
		respondWithError(w, http.StatusForbidden, "Não é permitido excluir funções do sistema sem usar a opção Force")
	case application.ErrRoleHasChildren:
		respondWithError(w, http.StatusBadRequest, "Esta função tem funções filhas e não pode ser excluída sem usar a opção Force")
	case application.ErrRoleHasUsers:
		respondWithError(w, http.StatusBadRequest, "Esta função tem usuários atribuídos e não pode ser excluída sem usar a opção Force")
	default:
		// Erro genérico ou inesperado
		log.Error().Err(err).Msg("Erro não mapeado no serviço de funções")
		respondWithError(w, http.StatusInternalServerError, "Erro interno do servidor")
	}
}

// getErrorCodeFromStatus retorna um código de erro baseado no status HTTP
func getErrorCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}