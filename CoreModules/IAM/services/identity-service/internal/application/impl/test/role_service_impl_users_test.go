/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Testes unitários para as operações de gerenciamento de usuários do RoleService.
 * Implementa casos de teste abrangentes para garantir a integridade das operações.
 * Segue princípios TDD, BDD e padrões Clean Architecture/Hexagonal.
 */

package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/innovabiz/iam/internal/application"
	"github.com/innovabiz/iam/internal/domain/event"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/domain/repository"
)

// createMockUser cria uma instância de modelo User para testes
func createMockUser(id, tenantID uuid.UUID, username string) *model.User {
	now := time.Now()
	return &model.User{
		ID_:          id,
		TenantID_:    tenantID,
		Username_:    username,
		Email_:       username + "@example.com",
		FirstName_:   "Test",
		LastName_:    "User",
		IsActive_:    true,
		CreatedAt_:   now,
		UpdatedAt_:   now,
		CreatedBy_:   uuid.New(),
		Metadata_:    map[string]interface{}{"key": "value"},
	}
}

func TestAssignRoleToUsers(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar usuários para o teste
	userID1 := uuid.New()
	userID2 := uuid.New()
	userIDs := []uuid.UUID{userID1, userID2}
	
	// Configurar request de atribuição
	activatesAt := time.Now().Add(1 * time.Hour)
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 dias
	request := &application.AssignRoleToUsersRequest{
		UserIDs:     userIDs,
		ActivatesAt: &activatesAt,
		ExpiresAt:   &expiresAt,
		Metadata: map[string]interface{}{
			"reason": "test assignment",
		},
	}
	
	// Configurar o mock para atribuir usuários
	mockRoleRepo.On("AssignUsers", ctx, tenantID, roleID, userIDs, mock.AnythingOfType("*repository.RoleAssignmentDetails")).
		Return(nil)
	
	// Configurar o mock para publicar evento
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleAssignedToUsers, mock.Anything).
		Return(nil)
	
	// Act
	result, err := service.AssignUsers(ctx, tenantID, roleID, request)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(userIDs), result.UsersAssigned)
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestRevokeRoleFromUsers(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar usuários para o teste
	userID1 := uuid.New()
	userID2 := uuid.New()
	userIDs := []uuid.UUID{userID1, userID2}
	
	// Configurar o mock para revogar usuários
	mockRoleRepo.On("RevokeUsers", ctx, tenantID, roleID, userIDs).Return(nil)
	
	// Configurar o mock para publicar evento
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleRevokedFromUsers, mock.Anything).
		Return(nil)
	
	// Act
	result, err := service.RevokeUsers(ctx, tenantID, roleID, userIDs)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(userIDs), result.UsersRevoked)
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestGetRoleUsers(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar usuários de teste
	mockUsers := []*model.User{
		createMockUser(uuid.New(), tenantID, "user1"),
		createMockUser(uuid.New(), tenantID, "user2"),
		createMockUser(uuid.New(), tenantID, "user3"),
	}
	
	pagination := application.Pagination{
		Page:     1,
		PageSize: 10,
	}
	
	expectedTotal := int64(3)
	
	// Configurar o mock para retornar usuários
	mockRoleRepo.On("GetAssignedUsers", ctx, tenantID, roleID, mock.AnythingOfType("repository.Pagination")).
		Return(mockUsers, expectedTotal, nil)
	
	// Act
	response, err := service.GetUsers(ctx, tenantID, roleID, pagination)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, len(mockUsers), len(response.Users))
	assert.Equal(t, expectedTotal, response.Total)
	assert.Equal(t, pagination.Page, response.Page)
	assert.Equal(t, pagination.PageSize, response.PageSize)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignRoleToUsersWithInvalidRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Configurar o mock para retornar erro ao buscar função
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(nil, repository.ErrRoleNotFound)
	
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	request := &application.AssignRoleToUsersRequest{
		UserIDs: userIDs,
	}
	
	// Act
	result, err := service.AssignUsers(ctx, tenantID, roleID, request)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrRoleNotFound, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignRoleToUsersWithInactiveRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role inativa para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRole.IsActive_ = false
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	request := &application.AssignRoleToUsersRequest{
		UserIDs: userIDs,
	}
	
	// Act
	result, err := service.AssignUsers(ctx, tenantID, roleID, request)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "inactive")
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignRoleToUsersWithInvalidActivationDate(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar datas inválidas (expiração antes da ativação)
	activatesAt := time.Now().Add(2 * time.Hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	request := &application.AssignRoleToUsersRequest{
		UserIDs:     userIDs,
		ActivatesAt: &activatesAt,
		ExpiresAt:   &expiresAt,
	}
	
	// Act
	result, err := service.AssignUsers(ctx, tenantID, roleID, request)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expiration date must be after activation date")
	mockRoleRepo.AssertExpectations(t)
}

func TestSyncSystemRoles(t *testing.T) {
	// Arrange
	service, mockRoleRepo, mockPermRepo, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	
	// Mocks para uma função existente
	existingRoleID := uuid.New()
	existingRole := createMockRole(existingRoleID, tenantID, "system.admin")
	mockRoleRepo.On("FindByCode", mock.Anything, tenantID, "system.admin").Return(existingRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)
	
	// Mocks para uma função nova
	mockRoleRepo.On("FindByCode", mock.Anything, tenantID, "system.user").Return(nil, repository.ErrRoleNotFound)
	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)
	
	// Mocks para uma função nova sem permissões
	mockRoleRepo.On("FindByCode", mock.Anything, tenantID, "system.auditor").Return(nil, repository.ErrRoleNotFound)
	
	// Mocks para verificação de ciclos
	mockRoleRepo.On("HasCyclicDependency", mock.Anything, tenantID, mock.Anything, mock.Anything).Return(false, nil)
	
	// Mocks para permissões
	permIDs := []string{
		"system.admin.*", "users.manage.*", "roles.admin.manage", "permissions.admin.manage",
		"system.access", "users.profile.view", "users.profile.edit",
	}
	for _, code := range permIDs {
		permID := uuid.New()
		mockPermRepo.On("FindByCode", mock.Anything, tenantID, code).
			Return(createMockPermission(permID, tenantID, code), nil)
	}
	
	// Mock para atribuir permissões
	mockRoleRepo.On("AssignPermissions", mock.Anything, tenantID, mock.Anything, mock.Anything).Return(nil)
	
	// Mocks para eventos
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleCreated, mock.Anything).Return(nil)
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleUpdated, mock.Anything).Return(nil)
	mockEventBus.On("Publish", mock.Anything, event.TopicPermissionsAssignedToRole, mock.Anything).Return(nil)
	
	// Act
	result, err := service.SyncSystemRoles(ctx, tenantID, userID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.RolesCreated, 1)
	assert.GreaterOrEqual(t, result.RolesUpdated, 0)
	assert.GreaterOrEqual(t, result.PermissionsAssigned, 1)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}