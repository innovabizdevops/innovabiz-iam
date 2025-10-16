/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Definição de eventos relacionados a funções (roles) no sistema IAM.
 * Implementa os contratos e estruturas para eventos publicados pelo domínio de funções.
 * Segue princípios de Event-Driven Architecture e Domain-Driven Design (DDD).
 */

package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Tópicos para eventos relacionados a funções
	TopicRoleCreated             = "iam.role.created"
	TopicRoleUpdated             = "iam.role.updated"
	TopicRoleSoftDeleted         = "iam.role.soft_deleted"
	TopicRoleHardDeleted         = "iam.role.hard_deleted"
	
	// Tópicos para eventos relacionados a permissões em funções
	TopicPermissionsAssignedToRole  = "iam.role.permissions.assigned"
	TopicPermissionsRevokedFromRole = "iam.role.permissions.revoked"
	
	// Tópicos para eventos relacionados a usuários em funções
	TopicRoleAssignedToUsers     = "iam.role.users.assigned"
	TopicRoleRevokedFromUsers    = "iam.role.users.revoked"
)

// RoleEvent interface base para eventos relacionados a funções
type RoleEvent interface {
	Event
	GetRoleID() uuid.UUID
	GetTenantID() uuid.UUID
}

// RoleCreatedEvent evento emitido quando uma nova função é criada
type RoleCreatedEvent struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	RoleID      uuid.UUID  `json:"role_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	IsActive    bool       `json:"is_active"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	EventTime   time.Time  `json:"event_time"`
}

func (e *RoleCreatedEvent) GetType() string {
	return TopicRoleCreated
}

func (e *RoleCreatedEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleCreatedEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleCreatedEvent) GetTime() time.Time {
	return e.EventTime
}

// RoleUpdatedEvent evento emitido quando uma função é atualizada
type RoleUpdatedEvent struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	RoleID      uuid.UUID  `json:"role_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	IsActive    bool       `json:"is_active"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
	EventTime   time.Time  `json:"event_time"`
}

func (e *RoleUpdatedEvent) GetType() string {
	return TopicRoleUpdated
}

func (e *RoleUpdatedEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleUpdatedEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleUpdatedEvent) GetTime() time.Time {
	return e.EventTime
}

// RoleSoftDeletedEvent evento emitido quando uma função é excluída logicamente
type RoleSoftDeletedEvent struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	RoleID      uuid.UUID  `json:"role_id"`
	Code        string     `json:"code"`
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at"`
	EventTime   time.Time  `json:"event_time"`
}

func (e *RoleSoftDeletedEvent) GetType() string {
	return TopicRoleSoftDeleted
}

func (e *RoleSoftDeletedEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleSoftDeletedEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleSoftDeletedEvent) GetTime() time.Time {
	return e.EventTime
}

// RoleHardDeletedEvent evento emitido quando uma função é excluída permanentemente
type RoleHardDeletedEvent struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	RoleID      uuid.UUID  `json:"role_id"`
	Code        string     `json:"code"`
	EventTime   time.Time  `json:"event_time"`
}

func (e *RoleHardDeletedEvent) GetType() string {
	return TopicRoleHardDeleted
}

func (e *RoleHardDeletedEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleHardDeletedEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleHardDeletedEvent) GetTime() time.Time {
	return e.EventTime
}

// PermissionsAssignedToRoleEvent evento emitido quando permissões são atribuídas a uma função
type PermissionsAssignedToRoleEvent struct {
	TenantID      uuid.UUID    `json:"tenant_id"`
	RoleID        uuid.UUID    `json:"role_id"`
	RoleCode      string       `json:"role_code"`
	PermissionIDs []uuid.UUID  `json:"permission_ids"`
	EventTime     time.Time    `json:"event_time"`
}

func (e *PermissionsAssignedToRoleEvent) GetType() string {
	return TopicPermissionsAssignedToRole
}

func (e *PermissionsAssignedToRoleEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *PermissionsAssignedToRoleEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *PermissionsAssignedToRoleEvent) GetTime() time.Time {
	return e.EventTime
}

// PermissionsRevokedFromRoleEvent evento emitido quando permissões são revogadas de uma função
type PermissionsRevokedFromRoleEvent struct {
	TenantID      uuid.UUID    `json:"tenant_id"`
	RoleID        uuid.UUID    `json:"role_id"`
	RoleCode      string       `json:"role_code"`
	PermissionIDs []uuid.UUID  `json:"permission_ids"`
	EventTime     time.Time    `json:"event_time"`
}

func (e *PermissionsRevokedFromRoleEvent) GetType() string {
	return TopicPermissionsRevokedFromRole
}

func (e *PermissionsRevokedFromRoleEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *PermissionsRevokedFromRoleEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *PermissionsRevokedFromRoleEvent) GetTime() time.Time {
	return e.EventTime
}

// RoleAssignedToUsersEvent evento emitido quando uma função é atribuída a usuários
type RoleAssignedToUsersEvent struct {
	TenantID    uuid.UUID    `json:"tenant_id"`
	RoleID      uuid.UUID    `json:"role_id"`
	RoleCode    string       `json:"role_code"`
	UserIDs     []uuid.UUID  `json:"user_ids"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"`
	ActivatesAt *time.Time   `json:"activates_at,omitempty"`
	AssignedAt  time.Time    `json:"assigned_at"`
	EventTime   time.Time    `json:"event_time"`
}

func (e *RoleAssignedToUsersEvent) GetType() string {
	return TopicRoleAssignedToUsers
}

func (e *RoleAssignedToUsersEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleAssignedToUsersEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleAssignedToUsersEvent) GetTime() time.Time {
	return e.EventTime
}

// RoleRevokedFromUsersEvent evento emitido quando uma função é revogada de usuários
type RoleRevokedFromUsersEvent struct {
	TenantID    uuid.UUID    `json:"tenant_id"`
	RoleID      uuid.UUID    `json:"role_id"`
	RoleCode    string       `json:"role_code"`
	UserIDs     []uuid.UUID  `json:"user_ids"`
	EventTime   time.Time    `json:"event_time"`
}

func (e *RoleRevokedFromUsersEvent) GetType() string {
	return TopicRoleRevokedFromUsers
}

func (e *RoleRevokedFromUsersEvent) GetRoleID() uuid.UUID {
	return e.RoleID
}

func (e *RoleRevokedFromUsersEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *RoleRevokedFromUsersEvent) GetTime() time.Time {
	return e.EventTime
}