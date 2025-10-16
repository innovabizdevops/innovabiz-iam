/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Testes unitários para o serviço de funções (RoleService).
 * Implementa testes abrangentes para todas as operações do serviço.
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
	"github.com/innovabiz/iam/internal/application/impl"
	"github.com/innovabiz/iam/internal/domain/event"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/domain/repository"
)

// MockRoleRepository é um mock do repositório de funções para testes
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Role, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Role, error) {
	args := m.Called(ctx, tenantID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) SoftDelete(ctx context.Context, tenantID, id, deletedBy uuid.UUID) error {
	args := m.Called(ctx, tenantID, id, deletedBy)
	return args.Error(0)
}

func (m *MockRoleRepository) HardDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockRoleRepository) FindAll(ctx context.Context, tenantID uuid.UUID, filter repository.RoleFilter, pagination repository.Pagination) ([]*model.Role, int64, error) {
	args := m.Called(ctx, tenantID, filter, pagination)
	return args.Get(0).([]*model.Role), args.Int64(1), args.Error(2)
}

func (m *MockRoleRepository) AssignPermissions(ctx context.Context, tenantID, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) RevokePermissions(ctx context.Context, tenantID, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) GetPermissions(ctx context.Context, tenantID, roleID uuid.UUID, pagination repository.Pagination) ([]*model.Permission, int64, error) {
	args := m.Called(ctx, tenantID, roleID, pagination)
	return args.Get(0).([]*model.Permission), args.Int64(1), args.Error(2)
}

func (m *MockRoleRepository) AssignUsers(ctx context.Context, tenantID, roleID uuid.UUID, userIDs []uuid.UUID, details *repository.RoleAssignmentDetails) error {
	args := m.Called(ctx, tenantID, roleID, userIDs, details)
	return args.Error(0)
}

func (m *MockRoleRepository) RevokeUsers(ctx context.Context, tenantID, roleID uuid.UUID, userIDs []uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, userIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAssignedUsers(ctx context.Context, tenantID, roleID uuid.UUID, pagination repository.Pagination) ([]*model.User, int64, error) {
	args := m.Called(ctx, tenantID, roleID, pagination)
	return args.Get(0).([]*model.User), args.Int64(1), args.Error(2)
}

func (m *MockRoleRepository) HasCyclicDependency(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, parentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tenantID, roleID, parentID)
	return args.Bool(0), args.Error(1)
}

// MockPermissionRepository é um mock do repositório de permissões para testes
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Permission, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Permission, error) {
	args := m.Called(ctx, tenantID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

// MockEventBus é um mock do barramento de eventos para testes
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventType string, evt event.Event) error {
	args := m.Called(ctx, eventType, evt)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler func(ctx context.Context, event event.Event) error) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(eventType string, handler func(ctx context.Context, event event.Event) error) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

// setupRoleService configura uma instância de teste do serviço de funções
func setupRoleService() (*impl.RoleServiceImpl, *MockRoleRepository, *MockPermissionRepository, *MockEventBus) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockEventBus := new(MockEventBus)
	
	service := impl.NewRoleService(mockRoleRepo, mockPermRepo, mockEventBus)
	
	return service, mockRoleRepo, mockPermRepo, mockEventBus
}

// createMockRole cria uma instância de modelo Role para testes
func createMockRole(id, tenantID uuid.UUID, code string) *model.Role {
	now := time.Now()
	return &model.Role{
		// Nota: Em um cenário real, usaríamos construtores adequados do modelo
		// Aqui simplificamos para teste, como se tivéssemos acesso aos campos diretamente
		ID_:          id,
		TenantID_:    tenantID,
		Code_:        code,
		Name_:        "Test Role",
		Description_: "Description for test role",
		Type_:        model.RoleTypeCustom,
		Priority_:    100,
		IsActive_:    true,
		CreatedAt_:   now,
		UpdatedAt_:   now,
		CreatedBy_:   uuid.New(),
		Metadata_:    map[string]interface{}{"key": "value"},
	}
}

func TestCreateRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	
	request := &application.CreateRoleRequest{
		Code:        "test.role",
		Name:        "Test Role",
		Description: "Test role description",
		Type:        model.RoleTypeCustom,
		Priority:    100,
		IsActive:    true,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}
	
	// Configurar mock para verificar ciclo (não há ciclo neste caso)
	mockRoleRepo.On("HasCyclicDependency", mock.Anything, tenantID, mock.Anything, mock.Anything).
		Return(false, nil)
	
	// Configurar mock para criar role
	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).
		Return(nil)
	
	// Configurar mock para publicar evento
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleCreated, mock.Anything).
		Return(nil)
	
	// Act
	response, err := service.Create(ctx, tenantID, userID, request)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, request.Code, response.Code)
	assert.Equal(t, request.Name, response.Name)
	assert.Equal(t, request.Description, response.Description)
	assert.Equal(t, request.Type, response.Type)
	assert.Equal(t, request.Priority, response.Priority)
	assert.Equal(t, request.IsActive, response.IsActive)
	assert.Equal(t, request.Metadata, response.Metadata)
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestGetRoleByID(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleID := uuid.New()
	
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	
	// Act
	response, err := service.GetByID(ctx, tenantID, roleID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, mockRole.ID(), response.ID)
	assert.Equal(t, mockRole.Code(), response.Code)
	assert.Equal(t, mockRole.Name(), response.Name)
	mockRoleRepo.AssertExpectations(t)
}

func TestGetRoleByCode(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	roleCode := "test.role"
	roleID := uuid.New()
	
	mockRole := createMockRole(roleID, tenantID, roleCode)
	mockRoleRepo.On("FindByCode", ctx, tenantID, roleCode).Return(mockRole, nil)
	
	// Act
	response, err := service.GetByCode(ctx, tenantID, roleCode)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, mockRole.ID(), response.ID)
	assert.Equal(t, mockRole.Code(), response.Code)
	assert.Equal(t, mockRole.Name(), response.Name)
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	
	mockRole := createMockRole(roleID, tenantID, "test.role")
	request := &application.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
		Priority:    200,
		IsActive:    false,
		ParentID:    nil,
		Metadata: map[string]interface{}{
			"updated": true,
		},
	}
	
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	mockRoleRepo.On("HasCyclicDependency", mock.Anything, tenantID, mockRole.ID(), mock.Anything).
		Return(false, nil)
	mockRoleRepo.On("Update", ctx, mock.AnythingOfType("*model.Role")).Return(nil)
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleUpdated, mock.Anything).Return(nil)
	
	// Act
	response, err := service.Update(ctx, tenantID, userID, roleID, request)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, request.Name, response.Name)
	assert.Equal(t, request.Description, response.Description)
	assert.Equal(t, request.Priority, response.Priority)
	assert.Equal(t, request.IsActive, response.IsActive)
	assert.Equal(t, mockRole.Code(), response.Code) // Código não deve mudar
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestSoftDeleteRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	mockRoleRepo.On("SoftDelete", ctx, tenantID, roleID, userID).Return(nil)
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleSoftDeleted, mock.Anything).Return(nil)
	
	// Act
	err := service.Delete(ctx, tenantID, userID, roleID, false)
	
	// Assert
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHardDeleteRole(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, mockEventBus := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	
	mockRole := createMockRole(roleID, tenantID, "test.role")
	mockRoleRepo.On("FindByID", ctx, tenantID, roleID).Return(mockRole, nil)
	mockRoleRepo.On("HardDelete", ctx, tenantID, roleID).Return(nil)
	mockEventBus.On("Publish", mock.Anything, event.TopicRoleHardDeleted, mock.Anything).Return(nil)
	
	// Act
	err := service.Delete(ctx, tenantID, userID, roleID, true)
	
	// Assert
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestListRoles(t *testing.T) {
	// Arrange
	service, mockRoleRepo, _, _ := setupRoleService()
	ctx := context.Background()
	tenantID := uuid.New()
	
	mockRoles := []*model.Role{
		createMockRole(uuid.New(), tenantID, "role1"),
		createMockRole(uuid.New(), tenantID, "role2"),
		createMockRole(uuid.New(), tenantID, "role3"),
	}
	
	filter := application.RoleFilter{
		NameOrCodeContains: "role",
		Types:              []string{model.RoleTypeCustom},
		IsActive:           nil, // nil = retorna ambos
	}
	
	pagination := application.Pagination{
		Page:     1,
		PageSize: 10,
	}
	
	expectedTotal := int64(3)
	
	mockRoleRepo.On("FindAll", ctx, tenantID, mock.AnythingOfType("repository.RoleFilter"), mock.AnythingOfType("repository.Pagination")).
		Return(mockRoles, expectedTotal, nil)
	
	// Act
	response, err := service.List(ctx, tenantID, filter, pagination)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, len(mockRoles), len(response.Roles))
	assert.Equal(t, expectedTotal, response.Total)
	assert.Equal(t, pagination.Page, response.Page)
	assert.Equal(t, pagination.PageSize, response.PageSize)
	assert.Equal(t, 1, response.TotalPages)
	mockRoleRepo.AssertExpectations(t)
}