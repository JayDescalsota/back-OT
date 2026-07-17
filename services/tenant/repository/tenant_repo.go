package repository

import (
	"context"
	"database/sql"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	sharedctx "github.com/clinicmanager/shared/context"
	bun "github.com/uptrace/bun"
)

type TenantRepository interface {
	FindTenantByID(ctx context.Context, id string) (*db.BunTenant, error)
	FindTenantBySlug(ctx context.Context, slug string) (*db.BunTenant, error)
	ListTenants(ctx context.Context) ([]*db.BunTenant, error)
	FindBranchByID(ctx context.Context, id string) (*db.BunBranch, error)
	FindBranchesByTenant(ctx context.Context, tenantID string) ([]*db.BunBranch, error)
	FindRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error)
	ListSystemRoles(ctx context.Context) ([]*db.BunTenantRole, error)
	ListTenantRoles(ctx context.Context, tenantID string) ([]*db.BunTenantRole, error)
	FindPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error)
	FindPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error)
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error)
	FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error)
}

type TenantRepo struct {
	db *bun.DB
}

func NewTenantRepo(db *bun.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) FindTenantByID(ctx context.Context, id string) (*db.BunTenant, error) {
	tenant := new(db.BunTenant)
	err := r.db.NewSelect().Model(tenant).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return tenant, nil
}

func (r *TenantRepo) FindTenantBySlug(ctx context.Context, slug string) (*db.BunTenant, error) {
	tenant := new(db.BunTenant)
	err := r.db.NewSelect().Model(tenant).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return tenant, nil
}

func (r *TenantRepo) ListTenants(ctx context.Context) ([]*db.BunTenant, error) {
	var tenants []*db.BunTenant
	err := r.db.NewSelect().Model(&tenants).Scan(ctx)
	return tenants, err
}

func (r *TenantRepo) FindBranchByID(ctx context.Context, id string) (*db.BunBranch, error) {
	branch := new(db.BunBranch)
	err := r.db.NewSelect().Model(branch).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return branch, nil
}

func (r *TenantRepo) FindBranchesByTenant(ctx context.Context, tenantID string) ([]*db.BunBranch, error) {
	var branches []*db.BunBranch
	q := r.db.NewSelect().Model(&branches)
	scopedTenantID := sharedctx.TenantIDFromCtx(ctx)
	if scopedTenantID != "" {
		q = q.Where("tenant_id = ?", scopedTenantID)
	} else {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Scan(ctx)
	return branches, err
}

func (r *TenantRepo) FindRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error) {
	role := new(db.BunTenantRole)
	err := r.db.NewSelect().Model(role).Where("id = ? ", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func (r *TenantRepo) ListSystemRoles(ctx context.Context) ([]*db.BunTenantRole, error) {
	var roles []*db.BunTenantRole
	err := r.db.NewSelect().Model(&roles).Where("is_system_role = ?", true).Scan(ctx)
	return roles, err
}

func (r *TenantRepo) ListTenantRoles(ctx context.Context, tenantID string) ([]*db.BunTenantRole, error) {
	var roles []*db.BunTenantRole
	q := r.db.NewSelect().Model(&roles)
	scopedTenantID := sharedctx.TenantIDFromCtx(ctx)
	if scopedTenantID != "" {
		q = q.Where("(tenant_id = ? OR is_system_role = ?)", scopedTenantID, true)
	} else {
		q = q.Where("(tenant_id = ? OR is_system_role = ?)", tenantID, true)
	}
	err := q.Scan(ctx)
	return roles, err
}

func (r *TenantRepo) FindPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error) {
	perm := new(db.BunTenantPermission)
	err := r.db.NewSelect().Model(perm).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return perm, nil
}

func (r *TenantRepo) FindPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error) {
	var permissions []*db.BunTenantPermission
	err := r.db.NewSelect().
		Model(&permissions).
		Column("p.*").
		TableExpr("tenant_role_permissions AS trp").
		Join("JOIN tenant_permissions AS p ON p.id = trp.permission_id").
		Where("trp.role_id = ?", roleID).
		Scan(ctx)
	return permissions, err
}

func (r *TenantRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error) {
	var rows []*assignmentRow
	query := `
		SELECT tua.*, tr.name AS role_name, tr.description AS role_description,
		       b.name AS branch_name, t.name AS tenant_name, t.slug AS tenant_slug
		FROM tenant_user_assignments AS tua
		JOIN tenant_roles AS tr ON tr.id = tua.role_id
		JOIN branches AS b ON b.id = tua.branch_id
		JOIN tenants AS t ON t.id = tua.tenant_id
		WHERE tua.user_id = ? AND tua.is_active = ?`
	args := []interface{}{userID, true}
	tctx := sharedctx.FromContext(ctx)
	if tctx.TenantID != "" {
		query += " AND tua.tenant_id = ?"
		args = append(args, tctx.TenantID)
	}
	if tctx.BranchID != "" {
		query += " AND tua.branch_id = ?"
		args = append(args, tctx.BranchID)
	}
	err := r.db.NewRaw(query, args...).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	result := make([]*model.TenantUserAssignment, len(rows))
	for i, row := range rows {
		result[i] = toAssignmentModel(row)
	}
	return result, nil
}

func (r *TenantRepo) FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error) {
	var rows []*assignmentRow
	query := `
		SELECT tua.*, tr.name AS role_name, tr.description AS role_description,
		       b.name AS branch_name, t.name AS tenant_name, t.slug AS tenant_slug
		FROM tenant_user_assignments AS tua
		JOIN tenant_roles AS tr ON tr.id = tua.role_id
		JOIN branches AS b ON b.id = tua.branch_id
		JOIN tenants AS t ON t.id = tua.tenant_id
		WHERE tua.user_id = ? AND tua.tenant_id = ? AND tua.is_active = ?`
	args := []interface{}{userID, tenantID, true}
	tctx := sharedctx.FromContext(ctx)
	if tctx.BranchID != "" {
		query += " AND tua.branch_id = ?"
		args = append(args, tctx.BranchID)
	}
	err := r.db.NewRaw(query, args...).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	result := make([]*model.TenantUserAssignment, len(rows))
	for i, row := range rows {
		result[i] = toAssignmentModel(row)
	}
	return result, nil
}
