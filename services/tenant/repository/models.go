package repository

import (
	"time"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
)

// assignmentRow is an internal JOIN result — not exposed as a gqlgen type.
type assignmentRow struct {
	ID              string     `bun:"id"`
	UserID          string     `bun:"user_id"`
	BranchID        string     `bun:"branch_id"`
	TenantID        string     `bun:"tenant_id"`
	RoleID          string     `bun:"role_id"`
	AssignedBy      *string    `bun:"assigned_by"`
	AssignedAt      time.Time  `bun:"assigned_at"`
	UpdatedAt       *time.Time `bun:"updated_at"`
	UpdatedBy       *string    `bun:"updated_by"`
	UpdatedAction   string     `bun:"updated_action"`
	IsActive        bool       `bun:"is_active"`
	IsUserPrimary   bool       `bun:"is_user_primary"`
	RoleName        string     `bun:"role_name"`
	RoleDescription *string    `bun:"role_description"`
	BranchName      string     `bun:"branch_name"`
	TenantName      string     `bun:"tenant_name"`
	TenantSlug      string     `bun:"tenant_slug"`
}

// toAssignmentModel maps the JOIN row into the generated TenantUserAssignment model.
// TenantUserAssignment stays generated (not mapped in gqlgen.yml) because it is a
// composite of data from multiple tables, not a direct table struct.
func toAssignmentModel(row *assignmentRow) *model.TenantUserAssignment {
	assignedAt := row.AssignedAt.Format(time.RFC3339)
	var updatedAt *string
	if row.UpdatedAt != nil {
		s := row.UpdatedAt.Format(time.RFC3339)
		updatedAt = &s
	}
	updatedAction := row.UpdatedAction
	return &model.TenantUserAssignment{
		ID:            row.ID,
		UserID:        row.UserID,
		Branch:        &db.BunBranch{ID: row.BranchID, Name: row.BranchName},
		Tenant:        &db.BunTenant{ID: row.TenantID, Name: row.TenantName, Slug: row.TenantSlug},
		Role:          &db.BunTenantRole{ID: row.RoleID, Name: row.RoleName, Description: row.RoleDescription},
		AssignedBy:    row.AssignedBy,
		AssignedAt:    assignedAt,
		UpdatedAt:     updatedAt,
		UpdatedBy:     row.UpdatedBy,
		UpdatedAction: &updatedAction,
		IsActive:      row.IsActive,
	}
}
