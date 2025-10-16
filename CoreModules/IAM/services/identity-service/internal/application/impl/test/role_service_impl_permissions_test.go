/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Testes unitários para as operações de gerenciamento de permissões do RoleService.
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

// createMockPermission cria uma instância de modelo Permission para testes
func createMockPermission(id, tenantID uuid.UUID, code string) *model.Permission {
	now := time.Now()
	return &model.Permission{
		ID_:          id,
		TenantID_:    tenantID,
		Code_:        code,
		Name_:        "Test Permission",
		Description_: "Description for test permission",
		IsActive_:    true,
		CreatedAt_:   now,
		UpdatedAt_:   now,
		CreatedBy_:   uuid.New(),
		Metadata_:    map[string]interface{}{"key": "value"},
	}
}

func TestAssignPermissionsToRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, mockPermRepo, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar permissões para o teste
	permID1 := uuid.New()
	permID2 := uuid.New()
	permissionIDs := []uuid.UUID{permID1, permID2}
	
	// Configurar o mock para encontrar as permissões
	mockPermRepo.On("FindByID", ctx, tenantID, permID1).Return(
		createMockPermission(permID1, tenantID, "test.permission.1"), nil)
	mockPermRepo.On("FindByID", ctx, tenantID, permID2).Return(
		createMockPermission(permID2, tenantID, "test.permission.2"), nil)
	
	// Configurar o mock para atribuir permissões
	mockRoleRepo.On("AssignPermissions", ctx, tenantID, roleID, permissionIDs).Return(nil)
	
	// Configurar o mock para publicar evento
	mockEventBus.On("Publish", mock.Anything, event.TopicPermissionsAssignedToRole, mock.Anything).Return(nil)
	
	// Act
	result, err := service.AssignPermissions(ctx, tenantID, roleID, permissionIDs)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(permissionIDs), result.PermissionsAssigned)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestRevokePermissionsFromRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, mockPermRepo, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar permissões para o teste
	permID1 := uuid.New()
	permID2 := uuid.New()
	permissionIDs := []uuid.UUID{permID1, permID2}
	
	// Configurar o mock para encontrar as permissões
	mockPermRepo.On("FindByID", ctx, tenantID, permID1).Return(
		createMockPermission(permID1, tenantID, "test.permission.1"), nil)
	mockPermRepo.On("FindByID", ctx, tenantID, permID2).Return(
		createMockPermission(permID2, tenantID, "test.permission.2"), nil)
	
	// Configurar o mock para revogar permissões
	mockRoleRepo.On("RevokePermissions", ctx, tenantID, roleID, permissionIDs).Return(nil)
	
	// Configurar o mock para publicar evento
	mockEventBus.On("Publish", mock.Anything, event.TopicPermissionsRevokedFromRole, mock.Anything).Return(nil)
	
	// Act
	result, err := service.RevokePermissions(ctx, tenantID, roleID, permissionIDs)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(permissionIDs), result.PermissionsRevoked)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestGetRolePermissions(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar permissões de teste
	mockPermissions := []*model.Permission{
		createMockPermission(uuid.New(), tenantID, "test.permission.1"),
		createMockPermission(uuid.New(), tenantID, "test.permission.2"),
		createMockPermission(uuid.New(), tenantID, "test.permission.3"),
	}
	
	pagination := application.Pagination{
		Page:     1,
		PageSize: 10,
	}
	
	expectedTotal := int64(3)
	
	// Configurar o mock para retornar permissões
	mockRoleRepo.On("GetPermissions", ctx, tenantID, roleID, mock.AnythingOfType("repository.Pagination")).
		Return(mockPermissions, expectedTotal, nil)
	
	// Act
	response, err := service.GetPermissions(ctx, tenantID, roleID, pagination)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, len(mockPermissions), len(response.Permissions))
	assert.Equal(t, expectedTotal, response.Total)
	assert.Equal(t, pagination.Page, response.Page)
	assert.Equal(t, pagination.PageSize, response.PageSize)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignPermissionsToRoleWithInvalidRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Configurar o mock para retornar erro ao buscar função
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(nil, repository.ErrRoleNotFound)
	
	permissionIDs := []uuid.UUID{uuid.New(), uuid.New()}
	
	// Act
	result, err := service.AssignPermissions(ctx, tenantID, roleID, permissionIDs)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrRoleNotFound, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignPermissionsToRoleWithInvalidPermission(t *testing.T) {
	// Arrange
	service, mockRoleRepo, mockPermRepo, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar permissões para o teste
	permID1 := uuid.New()
	permID2 := uuid.New()
	permissionIDs := []uuid.UUID{permID1, permID2}
	
	// Configurar o mock para encontrar a primeira permissão
	mockPermRepo.On("FindByID", ctx, tenantID, permID1).Return(
		createMockPermission(permID1, tenantID, "test.permission.1"), nil)
	
	// Configurar o mock para retornar erro na segunda permissão
	mockPermRepo.On("FindByID", ctx, tenantID, permID2).Return(nil, repository.ErrPermissionNotFound)
	
	// Act
	result, err := service.AssignPermissions(ctx, tenantID, roleID, permissionIDs)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestAssignPermissionsToRoleWithInactivePermission(t *testing.T) {
	// Arrange
	service, mockRoleRepo, mockPermRepo, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	// Criar role para o teste
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Criar permissão inativa
	permID := uuid.New()
	inactivePermission := createMockPermission(permID, tenantID, "test.permission.1")
	inactivePermission.IsActive_ = false
	
	permissionIDs := []uuid.UUID{permID}
	
	// Configurar o mock para encontrar a permissão inativa
	mockPermRepo.On("FindByID", ctx, tenantID, permID).Return(inactivePermission, nil)
	
	// Act
	result, err := service.AssignPermissions(ctx, tenantID, roleID, permissionIDs)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "inactive")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}