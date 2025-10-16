// Modelos de função e permissão para o serviço de identidade - INNOVABIZ Platform
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III, PSD2, AML/KYC
package models

import (
	"time"
)

// Status possíveis para uma função
const (
	RoleStatusActive   = "ACTIVE"
	RoleStatusInactive = "INACTIVE"
	RoleStatusDeleted  = "DELETED"
)

// Tipos de função
const (
	RoleTypeSystem = "SYSTEM"  // Funções internas do sistema
	RoleTypeCustom = "CUSTOM"  // Funções personalizadas por tenant
)

// Categorias de função
const (
	RoleCategoryAdministrative = "ADMINISTRATIVE"  // Funções administrativas
	RoleCategoryOperational    = "OPERATIONAL"     // Funções operacionais
	RoleCategoryBusiness       = "BUSINESS"        // Funções de negócio
	RoleCategoryAudit          = "AUDIT"           // Funções de auditoria
	RoleCategoryCompliance     = "COMPLIANCE"      // Funções de conformidade
	RoleCategoryTechnical      = "TECHNICAL"       // Funções técnicas
)

// Role representa uma função no sistema de IAM
type Role struct {
	ID            string    `json:"id" bson:"_id"`
	TenantID      string    `json:"tenant_id" bson:"tenant_id"`
	Name          string    `json:"name" bson:"name"`
	DisplayName   string    `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Description   string    `json:"description" bson:"description"`
	Type          string    `json:"type" bson:"type"`
	Category      string    `json:"category" bson:"category"`
	Status        string    `json:"status" bson:"status"`
	Permissions   []string  `json:"permissions,omitempty" bson:"permissions,omitempty"`
	RequiresMFA   bool      `json:"requires_mfa" bson:"requires_mfa"`
	MaxSessionDuration int      `json:"max_session_duration" bson:"max_session_duration"` // Em minutos
	AllowedTimeWindows []TimeWindow `json:"allowed_time_windows,omitempty" bson:"allowed_time_windows,omitempty"`
	AllowedLocations   []Location   `json:"allowed_locations,omitempty" bson:"allowed_locations,omitempty"`
	SensitivityLevel   string    `json:"sensitivity_level" bson:"sensitivity_level"` // LOW, MEDIUM, HIGH, CRITICAL
	ApprovalRequired   bool      `json:"approval_required" bson:"approval_required"`
	ApproverRoles      []string  `json:"approver_roles,omitempty" bson:"approver_roles,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
	CreatedBy          string    `json:"created_by" bson:"created_by"`
	UpdatedBy          string    `json:"updated_by" bson:"updated_by"`
	Deleted            bool      `json:"deleted" bson:"deleted"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy          string    `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	
	// Campos para conformidade com normas
	ComplianceRefs     []ComplianceRef `json:"compliance_refs,omitempty" bson:"compliance_refs,omitempty"`
	AuditEvents        []string  `json:"audit_events,omitempty" bson:"audit_events,omitempty"`
	VersionHistory     []RoleVersion `json:"version_history,omitempty" bson:"version_history,omitempty"`
	CurrentVersion     int       `json:"current_version" bson:"current_version"`
}

// RoleVersion representa uma versão histórica de uma função
type RoleVersion struct {
	Version     int       `json:"version" bson:"version"`
	ChangedAt   time.Time `json:"changed_at" bson:"changed_at"`
	ChangedBy   string    `json:"changed_by" bson:"changed_by"`
	Permissions []string  `json:"permissions" bson:"permissions"`
	Description string    `json:"description" bson:"description"`
	Status      string    `json:"status" bson:"status"`
	ChangeType  string    `json:"change_type" bson:"change_type"` // CREATE, UPDATE, DELETE
	ChangeNote  string    `json:"change_note" bson:"change_note"`
}

// RoleHierarchy representa uma relação hierárquica entre funções
type RoleHierarchy struct {
	ID          string    `json:"id" bson:"_id"`
	ParentRoleID string    `json:"parent_role_id" bson:"parent_role_id"`
	ChildRoleID  string    `json:"child_role_id" bson:"child_role_id"`
	TenantID     string    `json:"tenant_id" bson:"tenant_id"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	CreatedBy    string    `json:"created_by" bson:"created_by"`
}

// Permission representa uma permissão no sistema de IAM
type Permission struct {
	ID            string    `json:"id" bson:"_id"`
	Name          string    `json:"name" bson:"name"` // Formato: recurso:ação (ex: user:read)
	DisplayName   string    `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Description   string    `json:"description" bson:"description"`
	ResourceType  string    `json:"resource_type" bson:"resource_type"`
	ActionType    string    `json:"action_type" bson:"action_type"` // CREATE, READ, UPDATE, DELETE, EXECUTE
	ScopeType     string    `json:"scope_type" bson:"scope_type"` // GLOBAL, TENANT, USER
	Category      string    `json:"category" bson:"category"`
	SensitivityLevel string  `json:"sensitivity_level" bson:"sensitivity_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RequiresMFA   bool      `json:"requires_mfa" bson:"requires_mfa"`
	DualControl   bool      `json:"dual_control" bson:"dual_control"` // Requer aprovação
	RelatedPermissions []string `json:"related_permissions,omitempty" bson:"related_permissions,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
	CreatedBy     string    `json:"created_by" bson:"created_by"`
	UpdatedBy     string    `json:"updated_by" bson:"updated_by"`
	
	// Campos para conformidade com normas
	ComplianceRefs []ComplianceRef `json:"compliance_refs,omitempty" bson:"compliance_refs,omitempty"`
	AuditEvents    []string  `json:"audit_events,omitempty" bson:"audit_events,omitempty"`
}

// TimeWindow representa uma janela de tempo em que a função pode ser usada
type TimeWindow struct {
	DayOfWeek  int  `json:"day_of_week" bson:"day_of_week"` // 0-6, 0 é domingo
	StartHour  int  `json:"start_hour" bson:"start_hour"`   // 0-23
	StartMinute int  `json:"start_minute" bson:"start_minute"` // 0-59
	EndHour    int  `json:"end_hour" bson:"end_hour"`       // 0-23
	EndMinute  int  `json:"end_minute" bson:"end_minute"`   // 0-59
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

// Location representa uma localização geográfica permitida
type Location struct {
	CountryCode string `json:"country_code" bson:"country_code"`
	Region      string `json:"region,omitempty" bson:"region,omitempty"`
	City        string `json:"city,omitempty" bson:"city,omitempty"`
	IPRange     string `json:"ip_range,omitempty" bson:"ip_range,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

// ComplianceRef representa uma referência a uma norma ou framework de conformidade
type ComplianceRef struct {
	Standard    string `json:"standard" bson:"standard"`         // ISO27001, PCI-DSS, GDPR, etc.
	Version     string `json:"version,omitempty" bson:"version,omitempty"` // 2013, 3.2.1, etc.
	ControlID   string `json:"control_id" bson:"control_id"`     // A.9.2.3, 8.1.1, etc.
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

// RoleDefinition representa uma definição de função para criação ou atualização
type RoleDefinition struct {
	Name          string    `json:"name" validate:"required,min=3,max=50"`
	DisplayName   string    `json:"display_name,omitempty"`
	Description   string    `json:"description" validate:"required,min=5,max=500"`
	Type          string    `json:"type" validate:"required,oneof=SYSTEM CUSTOM"`
	Category      string    `json:"category" validate:"required"`
	Permissions   []string  `json:"permissions,omitempty"`
	RequiresMFA   bool      `json:"requires_mfa"`
	MaxSessionDuration int   `json:"max_session_duration,omitempty"`
	SensitivityLevel string  `json:"sensitivity_level" validate:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	ApprovalRequired bool    `json:"approval_required"`
	ApproverRoles   []string `json:"approver_roles,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionDefinition representa uma definição de permissão para criação ou atualização
type PermissionDefinition struct {
	Name          string    `json:"name" validate:"required,min=3,max=100"`
	DisplayName   string    `json:"display_name,omitempty"`
	Description   string    `json:"description" validate:"required,min=5,max=500"`
	ResourceType  string    `json:"resource_type" validate:"required"`
	ActionType    string    `json:"action_type" validate:"required,oneof=CREATE READ UPDATE DELETE EXECUTE"`
	ScopeType     string    `json:"scope_type" validate:"required,oneof=GLOBAL TENANT USER"`
	Category      string    `json:"category" validate:"required"`
	SensitivityLevel string  `json:"sensitivity_level" validate:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	RequiresMFA   bool      `json:"requires_mfa"`
	DualControl   bool      `json:"dual_control"`
	RelatedPermissions []string `json:"related_permissions,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ComplianceRefs []ComplianceRef `json:"compliance_refs,omitempty"`
}

// RoleAssignmentRequest representa uma solicitação de atribuição de função
type RoleAssignmentRequest struct {
	UserID        string    `json:"user_id" validate:"required"`
	RoleID        string    `json:"role_id" validate:"required"`
	ExpiresAt     time.Time `json:"expires_at" validate:"required,gt=now"`
	Justification string    `json:"justification" validate:"required,min=10,max=1000"`
	RequestedBy   string    `json:"requested_by" validate:"required"`
}

// RoleAssignmentResponse representa a resposta a uma solicitação de atribuição
type RoleAssignmentResponse struct {
	AssignmentID  string    `json:"assignment_id"`
	UserID        string    `json:"user_id"`
	RoleID        string    `json:"role_id"`
	RoleName      string    `json:"role_name"`
	Status        string    `json:"status"`
	AssignedAt    time.Time `json:"assigned_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	RequiresApproval bool     `json:"requires_approval"`
	ApprovalStatus   string   `json:"approval_status,omitempty"`
}