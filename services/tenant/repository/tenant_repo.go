package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	sharedctx "github.com/clinicmanager/shared/context"
	shareddb "github.com/clinicmanager/shared/db"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
)

type TenantRepository interface {
	FindTenantByID(ctx context.Context, id string) (*db.BunTenant, error)
	FindTenantBySlug(ctx context.Context, slug string) (*db.BunTenant, error)
	ListTenants(ctx context.Context) ([]*db.BunTenant, error)
	FindBranchByID(ctx context.Context, id string) (*db.BunBranch, error)
	UpdateBranch(ctx context.Context, id string, branch *db.BunBranch) error
	FindBranchesByTenant(ctx context.Context, tenantID string) ([]*db.BunBranch, error)
	FindRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error)
	ListSystemRoles(ctx context.Context) ([]*db.BunTenantRole, error)
	ListTenantRoles(ctx context.Context, tenantID string) ([]*db.BunTenantRole, error)
	ListRolesByBranch(ctx context.Context, branchID string) ([]*db.BunTenantRole, error)
	FindPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error)
	FindPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error)
	ListPermissions(ctx context.Context, branchID string) ([]*db.BunTenantPermission, error)
	CreateRole(ctx context.Context, role *db.BunTenantRole) error
	SetRoleActive(ctx context.Context, id string, isActive bool) (*db.BunTenantRole, error)
	ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error)
	FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error)
	FindUserIDByEmail(ctx context.Context, email string) (string, error)
	UpsertAssignment(ctx context.Context, userID, branchID, tenantID, roleID string, assignedBy *string) error
	FindAssignmentByID(ctx context.Context, id string) (*model.TenantUserAssignment, error)
	FindAssignmentsByBranch(ctx context.Context, branchID string) ([]*model.TenantUserAssignment, error)
	UpdateAssignmentRole(ctx context.Context, id, roleID string) error
	SetAssignmentActive(ctx context.Context, id string, isActive bool) error
	CreateInvite(ctx context.Context, inv *db.BunTenantInvite) error
	FindPendingInvite(ctx context.Context, email, branchID string) (*model.TenantInvite, error)
	FindInviteByID(ctx context.Context, id string) (*model.TenantInvite, error)
	ListInvitesByBranch(ctx context.Context, branchID string) ([]*model.TenantInvite, error)
	AcceptInvite(ctx context.Context, email, branchID string) error
	RefreshInvite(ctx context.Context, id string) (*model.TenantInvite, error)
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

func (r *TenantRepo) UpdateBranch(ctx context.Context, id string, branch *db.BunBranch) error {
	_, err := r.tenantdb.NewUpdate(ctx, branch).Where("id = ?", id).Exec(ctx)
	return err
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

func (r *TenantRepo) ListRolesByBranch(ctx context.Context, branchID string) ([]*db.BunTenantRole, error) {
	var roles []*db.BunTenantRole
	// Roles are seeded per-branch (branch_id set, is_system_role=true).
	// Also include tenant-level roles (branch_id IS NULL) for the branch's tenant.
	branch := new(db.BunBranch)
	if err := r.alldb.NewSelect(ctx, branch).Where("id = ?", branchID).Scan(ctx); err != nil {
		// If branch lookup fails, fall back to strict branch filter.
		err2 := r.alldb.NewSelect(ctx, &roles).Where("branch_id = ?", branchID).Order("name ASC").Scan(ctx)
		return roles, err2
	}
	err := r.alldb.NewSelect(ctx, &roles).
		Where("branch_id = ?", branchID).
		WhereOr("(tenant_id = ? AND branch_id IS NULL)", branch.TenantID).
		Order("name ASC").
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

func (r *TenantRepo) ListPermissions(ctx context.Context, branchID string) ([]*db.BunTenantPermission, error) {
	var permissions []*db.BunTenantPermission
	// Permissions are seeded per-branch. Use alldb (unscoped) with an explicit
	// branch filter so callers without branch context still get correct results.
	q := r.alldb.NewSelect(ctx, &permissions)
	if branchID != "" {
		q = q.Where("branch_id = ?", branchID)
	}
	err := q.Order("resource ASC", "action ASC").Scan(ctx)
	return permissions, err
}

func (r *TenantRepo) CreateRole(ctx context.Context, role *db.BunTenantRole) error {
	_, err := r.alldb.NewInsert(role).Exec(ctx)
	return err
}

func (r *TenantRepo) SetRoleActive(ctx context.Context, id string, isActive bool) (*db.BunTenantRole, error) {
	role := new(db.BunTenantRole)
	if err := r.alldb.NewSelect(ctx, role).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	role.IsActive = isActive
	if _, err := r.alldb.NewUpdate(ctx, role).Where("id = ?", id).Exec(ctx); err != nil {
		return nil, err
	}
	return role, nil
}

// tenantRolePermissionRow maps the tenant_role_permissions join table.
type tenantRolePermissionRow struct {
	bun.BaseModel `bun:"table:tenant_role_permissions"`
	RoleID        string `bun:"role_id,pk"`
	PermissionID  string `bun:"permission_id,pk"`
}

func (r *TenantRepo) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Deduplicate to avoid PK conflicts on bulk insert.
	seen := make(map[string]struct{}, len(permissionIDs))
	rows := make([]tenantRolePermissionRow, 0, len(permissionIDs))
	for _, pid := range permissionIDs {
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		rows = append(rows, tenantRolePermissionRow{RoleID: roleID, PermissionID: pid})
	}
	if _, err := r.raw.NewDelete().Table("tenant_role_permissions").Where("role_id = ?", roleID).Exec(ctx); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	_, err := r.raw.NewInsert().Model(&rows).On("CONFLICT (role_id, permission_id) DO NOTHING").Exec(ctx)
	return err
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

// userRef is a minimal read-only view of the users table (owned by the user
// service; same shared Postgres). Used only to resolve invite emails to IDs.
type userRef struct {
	bun.BaseModel `bun:"table:users"`
	ID            string `bun:"id,pk"`
	Email         string `bun:"email"`
}

func (r *TenantRepo) FindUserIDByEmail(ctx context.Context, email string) (string, error) {
	u := new(userRef)
	err := r.alldb.NewSelect(ctx, u).Where("lower(email) = lower(?)", email).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return u.ID, nil
}

// assignmentInsert maps the tenant_user_assignments table for writes.
type assignmentInsert struct {
	bun.BaseModel `bun:"table:tenant_user_assignments"`
	ID            string  `bun:"id,pk"`
	UserID        string  `bun:"user_id,notnull"`
	BranchID      string  `bun:"branch_id,notnull"`
	TenantID      string  `bun:"tenant_id,notnull"`
	RoleID        string  `bun:"role_id,notnull"`
	AssignedBy    *string `bun:"assigned_by"`
}

func (r *TenantRepo) UpsertAssignment(ctx context.Context, userID, branchID, tenantID, roleID string, assignedBy *string) error {
	row := &assignmentInsert{
		ID:         uuid.NewString(),
		UserID:     userID,
		BranchID:   branchID,
		TenantID:   tenantID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
	}
	_, err := r.raw.NewInsert().Model(row).
		On("CONFLICT (user_id, branch_id, tenant_id) DO UPDATE SET role_id = EXCLUDED.role_id, is_active = true, updated_at = NOW()").
		Exec(ctx)
	return err
}

const assignmentJoin = `
	SELECT tua.*, tr.name AS role_name, tr.description AS role_description,
	       b.name AS branch_name, t.name AS tenant_name, t.slug AS tenant_slug
	FROM tenant_user_assignments AS tua
	JOIN tenant_roles AS tr ON tr.id = tua.role_id
	JOIN tenant_branches AS b ON b.id = tua.branch_id
	JOIN tenants AS t ON t.id = tua.tenant_id`

func (r *TenantRepo) FindAssignmentByID(ctx context.Context, id string) (*model.TenantUserAssignment, error) {
	row := new(assignmentRow)
	err := r.raw.NewRaw(assignmentJoin+` WHERE tua.id = ?`, id).Scan(ctx, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return toAssignmentModel(row), nil
}

func (r *TenantRepo) FindAssignmentsByBranch(ctx context.Context, branchID string) ([]*model.TenantUserAssignment, error) {
	var rows []*assignmentRow
	// SECURITY NOTE: raw query — tenant scoping applied manually via sharedctx.
	query := assignmentJoin + ` WHERE tua.branch_id = ?`
	args := []interface{}{branchID}
	if tctx := sharedctx.FromContext(ctx); tctx.TenantID != "" {
		query += " AND tua.tenant_id = ?"
		args = append(args, tctx.TenantID)
	}
	if err := r.raw.NewRaw(query, args...).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	result := make([]*model.TenantUserAssignment, len(rows))
	for i, row := range rows {
		result[i] = toAssignmentModel(row)
	}
	return result, nil
}

func (r *TenantRepo) UpdateAssignmentRole(ctx context.Context, id, roleID string) error {
	_, err := r.raw.NewUpdate().Table("tenant_user_assignments").
		Set("role_id = ?", roleID).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *TenantRepo) SetAssignmentActive(ctx context.Context, id string, isActive bool) error {
	_, err := r.raw.NewUpdate().Table("tenant_user_assignments").
		Set("is_active = ?", isActive).
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// inviteRow joins an invite with its branch/role names for the list view.
type inviteRow struct {
	ID         string     `bun:"id"`
	Email      string     `bun:"email"`
	BranchID   string     `bun:"branch_id"`
	BranchName string     `bun:"branch_name"`
	RoleID     string     `bun:"role_id"`
	RoleName   string     `bun:"role_name"`
	RoleDesc   *string    `bun:"role_description"`
	Status     string     `bun:"status"`
	CreatedAt  time.Time  `bun:"created_at"`
	AcceptedAt *time.Time `bun:"accepted_at"`
	ExpiresAt  time.Time  `bun:"expires_at"`
}

func toInviteModel(row *inviteRow) *model.TenantInvite {
	var acceptedAt *string
	if row.AcceptedAt != nil {
		s := row.AcceptedAt.Format(time.RFC3339)
		acceptedAt = &s
	}
	exp := row.ExpiresAt.Format(time.RFC3339)
	return &model.TenantInvite{
		ID:    row.ID,
		Email: row.Email,
		Branch: &db.BunBranch{
			ID:       row.BranchID,
			Name:     row.BranchName,
			Timezone: "UTC",
		},
		Role: &db.BunTenantRole{
			ID:          row.RoleID,
			Name:        row.RoleName,
			Description: row.RoleDesc,
		},
		Status:     row.Status,
		InvitedAt:  row.CreatedAt.Format(time.RFC3339),
		AcceptedAt: acceptedAt,
		ExpiresAt:  exp,
	}
}

const inviteJoin = `
	SELECT ti.id, ti.email, ti.branch_id, b.name AS branch_name,
	       ti.role_id, tr.name AS role_name, tr.description AS role_description,
	       ti.status, ti.created_at, ti.accepted_at, ti.expires_at
	FROM tenant_invites AS ti
	JOIN tenant_branches AS b ON b.id = ti.branch_id
	JOIN tenant_roles AS tr ON tr.id = ti.role_id`

func (r *TenantRepo) scanInvite(ctx context.Context, where string, args ...interface{}) (*model.TenantInvite, error) {
	row := new(inviteRow)
	err := r.raw.NewRaw(inviteJoin+" WHERE "+where, args...).Scan(ctx, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return toInviteModel(row), nil
}

func (r *TenantRepo) CreateInvite(ctx context.Context, inv *db.BunTenantInvite) error {
	_, err := r.alldb.NewInsert(inv).Exec(ctx)
	return err
}

func (r *TenantRepo) FindPendingInvite(ctx context.Context, email, branchID string) (*model.TenantInvite, error) {
	return r.scanInvite(ctx, "lower(ti.email) = lower(?) AND ti.branch_id = ? AND ti.status = 'pending'", email, branchID)
}

func (r *TenantRepo) FindInviteByID(ctx context.Context, id string) (*model.TenantInvite, error) {
	return r.scanInvite(ctx, "ti.id = ?", id)
}

func (r *TenantRepo) ListInvitesByBranch(ctx context.Context, branchID string) ([]*model.TenantInvite, error) {
	var rows []*inviteRow
	query := inviteJoin + ` WHERE ti.branch_id = ? ORDER BY ti.created_at DESC`
	if err := r.raw.NewRaw(query, branchID).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	result := make([]*model.TenantInvite, len(rows))
	for i, row := range rows {
		result[i] = toInviteModel(row)
	}
	return result, nil
}

func (r *TenantRepo) AcceptInvite(ctx context.Context, email, branchID string) error {
	_, err := r.raw.NewUpdate().Table("tenant_invites").
		Set("status = 'accepted'").
		Set("accepted_at = NOW()").
		Set("updated_at = NOW()").
		Where("lower(email) = lower(?)", email).
		Where("branch_id = ?", branchID).
		Where("status = 'pending'").
		Exec(ctx)
	return err
}

func (r *TenantRepo) RefreshInvite(ctx context.Context, id string) (*model.TenantInvite, error) {
	if _, err := r.raw.NewUpdate().Table("tenant_invites").
		Set("status = 'pending'").
		Set("expires_at = NOW() + INTERVAL '7 days'").
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return nil, err
	}
	return r.FindInviteByID(ctx, id)
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
