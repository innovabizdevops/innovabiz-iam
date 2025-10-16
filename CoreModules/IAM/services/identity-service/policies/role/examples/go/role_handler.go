// Package handler fornece implementações de controladores HTTP para o RoleService
//
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/application/dto"
	"github.com/innovabiz/iam/internal/application/service"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/infrastructure/authz"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
)

// RoleHandler gerencia requisições HTTP relacionadas a funções (roles)
type RoleHandler struct {
	roleService       service.RoleService
	permissionService service.PermissionService
	userService       service.UserService
}

// NewRoleHandler cria uma nova instância de RoleHandler
func NewRoleHandler(
	roleService service.RoleService,
	permissionService service.PermissionService,
	userService service.UserService,
) *RoleHandler {
	return &RoleHandler{
		roleService:       roleService,
		permissionService: permissionService,
		userService:       userService,
	}
}

// RegisterRoutes registra as rotas do controlador no router Gin
func (h *RoleHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Configurações de autorização
	authzConfig := authz.DefaultConfig()
	
	// Grupo de rotas para gestão de funções
	roles := router.Group("/roles")
	
	// Rotas para CRUD básico de funções
	roles.POST("", authz.Middleware(authzConfig, "crud.create_decision"), h.CreateRole)
	roles.GET("", authz.Middleware(authzConfig, "crud.list_decision"), h.ListRoles)
	roles.GET("/:id", authz.Middleware(authzConfig, "crud.read_decision"), h.GetRole)
	roles.PUT("/:id", authz.Middleware(authzConfig, "crud.update_decision"), h.UpdateRole)
	roles.DELETE("/:id", authz.Middleware(authzConfig, "crud.delete_decision"), h.DeleteRole)
	roles.DELETE("/:id/permanent", authz.Middleware(authzConfig, "crud.permanent_delete_decision"), h.PermanentDeleteRole)
	
	// Rotas para gestão de permissões
	roles.POST("/:id/permissions", authz.Middleware(authzConfig, "permissions.permission_assignment_decision"), h.AssignPermission)
	roles.DELETE("/:id/permissions/:permission_id", authz.Middleware(authzConfig, "permissions.permission_revocation_decision"), h.RevokePermission)
	roles.GET("/:id/permissions", authz.Middleware(authzConfig, "permissions.permission_check_decision"), h.ListPermissions)
	
	// Rotas para gestão de hierarquia
	roles.POST("/hierarchy", authz.Middleware(authzConfig, "hierarchy.hierarchy_addition_decision"), h.AddHierarchy)
	roles.DELETE("/hierarchy/:parent_id/:child_id", authz.Middleware(authzConfig, "hierarchy.hierarchy_removal_decision"), h.RemoveHierarchy)
	roles.GET("/:id/hierarchy", authz.Middleware(authzConfig, "hierarchy.hierarchy_query_decision"), h.GetHierarchy)
	
	// Rotas para atribuição de função a usuário
	roles.POST("/assignments", authz.Middleware(authzConfig, "user_assignment.role_assignment_decision"), h.AssignRoleToUser)
	roles.DELETE("/assignments/:assignment_id", authz.Middleware(authzConfig, "user_assignment.role_removal_decision"), h.RemoveRoleFromUser)
	roles.PUT("/assignments/:assignment_id/expiration", authz.Middleware(authzConfig, "user_assignment.expiration_update_decision"), h.UpdateAssignmentExpiration)
	roles.GET("/users/:user_id", authz.Middleware(authzConfig, "user_assignment.role_check_decision"), h.GetUserRoles)
}

// ----------------------------------------
// Handlers para CRUD básico de funções
// ----------------------------------------

// CreateRole cria uma nova função
// @Summary Criar nova função
// @Description Cria uma nova função no sistema
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "ID do Tenant"
// @Param role body dto.CreateRoleDTO true "Dados da função"
// @Success 201 {object} dto.RoleDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	tenantID := c.GetHeader("X-Tenant-ID")
	
	var createDTO dto.CreateRoleDTO
	if err := c.ShouldBindJSON(&createDTO); err != nil {
		logger.Error("falha ao vincular JSON", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_INPUT",
			"Dados de entrada inválidos",
			err.Error(),
		))
		return
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", createDTO)
	
	// Extrair claims do usuário do middleware de autenticação
	userClaims, _ := c.Get("userClaims")
	
	// Adicionar metadados de auditoria
	createDTO.Metadata = enrichMetadata(createDTO.Metadata, c, userClaims)
	
	// Criar função
	role, err := h.roleService.Create(c.Request.Context(), tenantID, createDTO)
	if err != nil {
		handleRoleServiceError(c, err, "criar função")
		return
	}
	
	// Retornar função criada
	c.JSON(http.StatusCreated, dto.NewRoleDTO(role))
}

// ListRoles lista funções com base em filtros
// @Summary Listar funções
// @Description Lista funções com base em filtros
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "ID do Tenant"
// @Param name query string false "Nome da função"
// @Param type query string false "Tipo da função (SYSTEM, CUSTOM)"
// @Param status query string false "Status da função (ACTIVE, INACTIVE)"
// @Param page query int false "Número da página"
// @Param size query int false "Tamanho da página"
// @Success 200 {object} dto.PageResponseDTO
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	tenantID := c.GetHeader("X-Tenant-ID")
	
	// Extrair parâmetros de consulta
	var filter dto.RoleFilterDTO
	if err := c.ShouldBindQuery(&filter); err != nil {
		logger.Error("falha ao vincular parâmetros de consulta", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_PARAMETERS",
			"Parâmetros de consulta inválidos",
			err.Error(),
		))
		return
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", filter)
	
	// Buscar funções
	page, err := h.roleService.List(c.Request.Context(), tenantID, filter)
	if err != nil {
		handleRoleServiceError(c, err, "listar funções")
		return
	}
	
	// Retornar página de resultados
	c.JSON(http.StatusOK, page)
}

// ----------------------------------------
// Handlers para gestão de permissões
// ----------------------------------------

// AssignPermission atribui uma permissão a uma função
// @Summary Atribuir permissão a função
// @Description Atribui uma permissão existente a uma função
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "ID do Tenant"
// @Param id path string true "ID da função"
// @Param permission body dto.PermissionAssignmentDTO true "Dados da atribuição"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermission(c *gin.Context) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	tenantID := c.GetHeader("X-Tenant-ID")
	roleID := c.Param("id")
	
	// Validar ID da função
	if _, err := uuid.Parse(roleID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_ID",
			"ID da função inválido",
			err.Error(),
		))
		return
	}
	
	var assignDTO dto.PermissionAssignmentDTO
	if err := c.ShouldBindJSON(&assignDTO); err != nil {
		logger.Error("falha ao vincular JSON", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_INPUT",
			"Dados de entrada inválidos",
			err.Error(),
		))
		return
	}
	
	// Construir dados do recurso para o middleware de autorização
	resourceData := map[string]interface{}{
		"role_id":       roleID,
		"permission_id": assignDTO.PermissionID,
		"metadata":      assignDTO.Metadata,
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", resourceData)
	
	// Extrair claims do usuário do middleware de autenticação
	userClaims, _ := c.Get("userClaims")
	
	// Adicionar metadados de auditoria
	assignDTO.Metadata = enrichMetadata(assignDTO.Metadata, c, userClaims)
	
	// Atribuir permissão
	err := h.roleService.AssignPermission(
		c.Request.Context(),
		tenantID,
		roleID,
		assignDTO.PermissionID,
		assignDTO.Metadata,
	)
	
	if err != nil {
		handleRoleServiceError(c, err, "atribuir permissão")
		return
	}
	
	// Retornar sucesso
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Permissão atribuída com sucesso"))
}

// ----------------------------------------
// Handlers para gestão de hierarquia
// ----------------------------------------

// AddHierarchy adiciona uma relação hierárquica entre funções
// @Summary Adicionar hierarquia entre funções
// @Description Adiciona uma relação hierárquica entre uma função pai e uma função filha
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "ID do Tenant"
// @Param hierarchy body dto.HierarchyDTO true "Dados da hierarquia"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/hierarchy [post]
func (h *RoleHandler) AddHierarchy(c *gin.Context) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	tenantID := c.GetHeader("X-Tenant-ID")
	
	var hierarchyDTO dto.HierarchyDTO
	if err := c.ShouldBindJSON(&hierarchyDTO); err != nil {
		logger.Error("falha ao vincular JSON", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_INPUT",
			"Dados de entrada inválidos",
			err.Error(),
		))
		return
	}
	
	// Validar IDs
	if _, err := uuid.Parse(hierarchyDTO.ParentID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_PARENT_ID",
			"ID da função pai inválido",
			err.Error(),
		))
		return
	}
	
	if _, err := uuid.Parse(hierarchyDTO.ChildID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_CHILD_ID",
			"ID da função filha inválido",
			err.Error(),
		))
		return
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", hierarchyDTO)
	
	// Extrair claims do usuário do middleware de autenticação
	userClaims, _ := c.Get("userClaims")
	
	// Adicionar metadados de auditoria
	hierarchyDTO.Metadata = enrichMetadata(hierarchyDTO.Metadata, c, userClaims)
	
	// Adicionar hierarquia
	err := h.roleService.AddHierarchy(
		c.Request.Context(),
		tenantID,
		hierarchyDTO.ParentID,
		hierarchyDTO.ChildID,
		hierarchyDTO.Metadata,
	)
	
	if err != nil {
		handleRoleServiceError(c, err, "adicionar hierarquia")
		return
	}
	
	// Retornar sucesso
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Hierarquia adicionada com sucesso"))
}

// ----------------------------------------
// Handlers para atribuição de função a usuário
// ----------------------------------------

// AssignRoleToUser atribui uma função a um usuário
// @Summary Atribuir função a usuário
// @Description Atribui uma função existente a um usuário
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "ID do Tenant"
// @Param assignment body dto.RoleAssignmentDTO true "Dados da atribuição"
// @Success 200 {object} dto.RoleAssignmentResponseDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/assignments [post]
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	tenantID := c.GetHeader("X-Tenant-ID")
	
	var assignDTO dto.RoleAssignmentDTO
	if err := c.ShouldBindJSON(&assignDTO); err != nil {
		logger.Error("falha ao vincular JSON", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_INPUT",
			"Dados de entrada inválidos",
			err.Error(),
		))
		return
	}
	
	// Validar IDs
	if _, err := uuid.Parse(assignDTO.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_ROLE_ID",
			"ID da função inválido",
			err.Error(),
		))
		return
	}
	
	if _, err := uuid.Parse(assignDTO.UserID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"INVALID_USER_ID",
			"ID do usuário inválido",
			err.Error(),
		))
		return
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", assignDTO)
	
	// Extrair claims do usuário do middleware de autenticação
	userClaims, _ := c.Get("userClaims")
	
	// Adicionar metadados de auditoria
	assignDTO.Metadata = enrichMetadata(assignDTO.Metadata, c, userClaims)
	
	// Atribuir função ao usuário
	assignment, err := h.roleService.AssignRoleToUser(
		c.Request.Context(),
		tenantID,
		assignDTO.RoleID,
		assignDTO.UserID,
		assignDTO.ExpiresAt,
		assignDTO.Metadata,
	)
	
	if err != nil {
		handleRoleServiceError(c, err, "atribuir função a usuário")
		return
	}
	
	// Retornar atribuição criada
	c.JSON(http.StatusCreated, dto.NewRoleAssignmentResponseDTO(assignment))
}

// ----------------------------------------
// Funções auxiliares
// ----------------------------------------

// enrichMetadata adiciona metadados de auditoria aos metadados existentes
func enrichMetadata(metadata map[string]interface{}, c *gin.Context, userClaims interface{}) map[string]interface{} {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	
	// Obter ID do usuário autenticado
	var userID string
	if claims, ok := userClaims.(map[string]interface{}); ok {
		if id, exists := claims["id"]; exists {
			userID, _ = id.(string)
		}
	}
	
	// Adicionar metadados de auditoria
	metadata["created_by"] = userID
	metadata["created_at"] = time.Now().UTC().Format(time.RFC3339)
	metadata["client_ip"] = c.ClientIP()
	metadata["request_id"] = c.GetString("X-Request-ID")
	
	return metadata
}

// handleRoleServiceError trata erros do serviço de funções
func handleRoleServiceError(c *gin.Context, err error, operation string) {
	logger := observability.GetLoggerFromContext(c.Request.Context())
	logger.Error("erro ao "+operation, "error", err)
	
	var statusCode int
	var errorDTO dto.ErrorResponse
	
	switch e := err.(type) {
	case model.ValidationError:
		statusCode = http.StatusBadRequest
		errorDTO = dto.NewErrorResponse("VALIDATION_ERROR", e.Message, e.Details)
		
	case model.NotFoundError:
		statusCode = http.StatusNotFound
		errorDTO = dto.NewErrorResponse("NOT_FOUND", e.Message, e.Details)
		
	case model.ConflictError:
		statusCode = http.StatusConflict
		errorDTO = dto.NewErrorResponse("CONFLICT", e.Message, e.Details)
		
	case model.UnauthorizedError:
		statusCode = http.StatusForbidden
		errorDTO = dto.NewErrorResponse("UNAUTHORIZED", e.Message, e.Details)
		
	default:
		statusCode = http.StatusInternalServerError
		errorDTO = dto.NewErrorResponse(
			"INTERNAL_ERROR",
			"Erro interno ao "+operation,
			"Entre em contato com o administrador do sistema",
		)
	}
	
	c.JSON(statusCode, errorDTO)
}