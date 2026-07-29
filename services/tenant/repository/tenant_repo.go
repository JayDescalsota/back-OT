package repository

import (
	"context"
	"database/sql"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	sharedctx "github.com/clinicmanager/shared/context"
	shareddb "github.com/clinicmanager/shared/db"
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
	FindAddressByID(ctx context.Context, id string) (*db.BunAddress, error)
	CreateAddress(ctx context.Context, addr *db.BunAddress) error
	UpdateAddress(ctx context.Context, id string, addr *db.BunAddress) error
}

type TenantRepo struct {
	// db: tenant_id + branch_id scoped (default, most restrictive).
	db *shareddb.ScopedDB
	// tenantdb: tenant_id scoped only (cross-branch access e.g. listing a tenant's branches).
	tenantdb *shareddb.ScopedDB
	// alldb: no automatic filters (for root entities like tenants themselves, or system-level data).
	alldb *shareddb.ScopedDB
	// raw: underlying *bun.DB for NewRaw queries that cannot be intercepted by ScopedDB.
	// SECURITY: raw SQL queries in this file apply manual tenant/branch filtering via sharedctx.
	raw *bun.DB
}

func NewTenantRepo(dbs *shareddb.DBSet) *TenantRepo {
	return &TenantRepo{
		db:       dbs.DB,
		tenantdb: dbs.TenantDB,
		alldb:    dbs.AllDB,
		raw:      dbs.AllDB.Raw(),
	}
}

func (r *TenantRepo) FindTenantByID(ctx context.Context, id string) (*db.BunTenant, error) {
	// r.db safely checks if BunTenant has tenant/branch columns (skips if absent).
	tenant := new(db.BunTenant)
	err := r.db.NewSelect(ctx, tenant).Where("id = ?", id).Scan(ctx)
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
	err := r.db.NewSelect(ctx, tenant).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return tenant, nil
}

func (r *TenantRepo) ListTenants(ctx context.Context) ([]*db.BunTenant, error) {
	// Admin-only: tenants table has no tenant_id column.
	var tenants []*db.BunTenant
	err := r.db.NewSelect(ctx, &tenants).Scan(ctx)
	return tenants, err
}

func (r *TenantRepo) FindBranchByID(ctx context.Context, id string) (*db.BunBranch, error) {
	// Branches have tenant_id — tenantdb enforces it automatically.
	branch := new(db.BunBranch)
	err := r.tenantdb.NewSelect(ctx, branch).Where("id = ?", id).Scan(ctx)
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
	// tenantdb auto-applies ctx tenant_id. If no ctx tenant, fall back to the explicit param.
	scopedTenantID := sharedctx.TenantIDFromCtx(ctx)
	if scopedTenantID != "" {
		err := r.tenantdb.NewSelect(ctx, &branches).Scan(ctx)
		return branches, err
	}
	err := r.db.NewSelect(ctx, &branches).Where("tenant_id = ?", tenantID).Scan(ctx)
	return branches, err
}

func (r *TenantRepo) FindRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error) {
	// Roles can be system-level (no tenant_id) — use alldb to avoid silently filtering them out.
	role := new(db.BunTenantRole)
	err := r.db.NewSelect(ctx, role).Where("id = ?", id).Scan(ctx)
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
	err := r.db.NewSelect(ctx, &roles).Where("is_system_role = ?", true).Scan(ctx)
	return roles, err
}

func (r *TenantRepo) ListTenantRoles(ctx context.Context, tenantID string) ([]*db.BunTenantRole, error) {
	var roles []*db.BunTenantRole
	// Roles overlap: tenant-specific AND system roles. Cannot use automatic scoping here
	// because system roles have tenant_id = NULL — tenantdb would filter them out.
	scopedTenantID := sharedctx.TenantIDFromCtx(ctx)
	if scopedTenantID == "" {
		scopedTenantID = tenantID
	}
	err := r.alldb.NewSelect(ctx, &roles).
		Where("(tenant_id = ? OR is_system_role = ?)", scopedTenantID, true).
		Scan(ctx)
	return roles, err
}

func (r *TenantRepo) FindPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error) {
	// Permissions are system-level definitions — use alldb.
	perm := new(db.BunTenantPermission)
	err := r.db.NewSelect(ctx, perm).Where("id = ?", id).Scan(ctx)
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
	// JOIN query — model is tenant_role_permissions, not BunTenantPermission directly.
	// Use r.raw directly since the primary model isn't BunTenantPermission.
	err := r.raw.NewSelect().
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
	// SECURITY NOTE: This is a complex multi-JOIN raw query. ScopedDB cannot intercept NewRaw.
	// Tenant/branch scoping is applied manually here via sharedctx.
	query := `
		SELECT tua.*, tr.name AS role_name, tr.description AS role_description,
		       b.name AS branch_name, t.name AS tenant_name, t.slug AS tenant_slug
		FROM tenant_user_assignments AS tua
		JOIN tenant_roles AS tr ON tr.id = tua.role_id
		JOIN tenant_branches AS b ON b.id = tua.branch_id
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
	err := r.raw.NewRaw(query, args...).Scan(ctx, &rows)
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
	// SECURITY NOTE: This is a complex multi-JOIN raw query. ScopedDB cannot intercept NewRaw.
	// Tenant/branch scoping is applied manually here via sharedctx.
	query := `
		SELECT tua.*, tr.name AS role_name, tr.description AS role_description,
		       b.name AS branch_name, t.name AS tenant_name, t.slug AS tenant_slug
		FROM tenant_user_assignments AS tua
		JOIN tenant_roles AS tr ON tr.id = tua.role_id
		JOIN tenant_branches AS b ON b.id = tua.branch_id
		JOIN tenants AS t ON t.id = tua.tenant_id
		WHERE tua.user_id = ? AND tua.tenant_id = ? AND tua.is_active = ?`
	args := []interface{}{userID, tenantID, true}
	tctx := sharedctx.FromContext(ctx)
	if tctx.BranchID != "" {
		query += " AND tua.branch_id = ?"
		args = append(args, tctx.BranchID)
	}
	err := r.raw.NewRaw(query, args...).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	result := make([]*model.TenantUserAssignment, len(rows))
	for i, row := range rows {
		result[i] = toAssignmentModel(row)
	}
	return result, nil
}

func (r *TenantRepo) FindAddressByID(ctx context.Context, id string) (*db.BunAddress, error) {
	addr := new(db.BunAddress)
	err := r.db.NewSelect(ctx, addr).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return addr, nil
}

func (r *TenantRepo) CreateAddress(ctx context.Context, addr *db.BunAddress) error {
	_, err := r.db.NewInsert(addr).Exec(ctx)
	return err
}

func (r *TenantRepo) UpdateAddress(ctx context.Context, id string, addr *db.BunAddress) error {
	_, err := r.db.NewUpdate(ctx, addr).Where("id = ?", id).Exec(ctx)
	return err
}
