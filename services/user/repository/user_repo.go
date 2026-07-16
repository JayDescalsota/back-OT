package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/clinicmanager/services/user/graph/model"
	"github.com/clinicmanager/services/user/models"
)

type BunUserBranchAssignment struct {
	bun.BaseModel `bun:"table:user_branch_assignments"`

	ID         string    `bun:"id,pk"`
	UserID     string    `bun:"user_id,notnull"`
	BranchID   string    `bun:"branch_id,notnull"`
	TenantID   string    `bun:"tenant_id,notnull"`
	RoleID     string    `bun:"role_id,notnull"`
	AssignedBy string    `bun:"assigned_by,notnull"`
	AssignedAt time.Time `bun:"assigned_at"`
	IsActive   bool      `bun:"is_active,default:true"`

	// Joined fields
	RoleName        string `bun:"role_name"`
	RoleDescription string `bun:"role_description"`
	BranchName      string `bun:"branch_name"`
	TenantName      string `bun:"tenant_name"`
}

type UserRepository interface {
	FindUserByID(ctx context.Context, id string) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	Register(ctx context.Context, email, password, validationToken string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	FindUserByValidationToken(ctx context.Context, token string) (*models.User, error)
	MarkUserAsValidated(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, newPassword string) error
	SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	FindUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
	ResetPassword(ctx context.Context, userID, newPassword string) error
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*BunUserBranchAssignment, error)
	FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*BunUserBranchAssignment, error)
	CreateSession(ctx context.Context, session *models.Session) error
	FindSessionByID(ctx context.Context, id string) (*models.Session, error)
	FindSessionByToken(ctx context.Context, token string) (*models.Session, error)
	FindSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllSessionsForUser(ctx context.Context, userID string) error
	FindTenantByID(ctx context.Context, id string) (*model.Tenant, error)
	FindBranchByID(ctx context.Context, id string) (*model.Branch, error)
	FindRoleByID(ctx context.Context, id string) (*model.Role, error)
	FindPermissionByID(ctx context.Context, id string) (*model.Permission, error)
}

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateSession(ctx context.Context, session *models.Session) error {
	_, err := r.db.NewInsert().Model(session).Exec(ctx)
	return err
}

func (r *UserRepo) FindSessionByID(ctx context.Context, id string) (*models.Session, error) {
	var session models.Session
	err := r.db.NewSelect().Model(&session).
		Where("id = ?", id).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserRepo) FindSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	err := r.db.NewSelect().Model(&session).
		Where("refresh_token = ?", token).
		Where("revoked = false").
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserRepo) FindSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.NewSelect().Model(&sessions).
		Where("user_id = ?", userID).
		Where("revoked = false").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *UserRepo) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.db.NewUpdate().Model((*models.Session)(nil)).
		Where("id = ?", sessionID).
		Set("revoked = true").
		Exec(ctx)
	return err
}

func (r *UserRepo) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().Model((*models.Session)(nil)).
		Where("user_id = ?", userID).
		Set("revoked = true").
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) Register(ctx context.Context, email, password, validationToken string) (*models.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	uuid := uuid.New()
	now := time.Now().UTC()
	user := &models.User{
		ID:              uuid.String(),
		Email:           email,
		Name:            "User",
		PasswordHash:    hash,
		IsActive:        true,
		ValidationToken: validationToken,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err = r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.db.NewUpdate().Model(&models.User{}).
		Set("last_login = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = r.db.NewUpdate().Model(&models.User{}).
		Set("password_hash = ?", hash).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByValidationToken(ctx context.Context, token string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("validation_token = ?", token).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.db.NewUpdate().Model(&models.User{}).
		Set("password_reset_token = ?", token).
		Set("password_reset_expires_at = ?", expiresAt).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).
		Where("password_reset_token = ?", token).
		Where("password_reset_expires_at > NOW()").
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) ResetPassword(ctx context.Context, userID, newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = r.db.NewUpdate().Model(&models.User{}).
		Set("password_hash = ?", hash).
		Set("password_reset_token = ?", "").
		Set("password_reset_expires_at = ?", nil).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) MarkUserAsValidated(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.db.NewUpdate().Model(&models.User{}).
		Set("is_validated = ?", true).
		Set("validated_at = ?", now).
		Set("validation_token = ?", "").
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*BunUserBranchAssignment, error) {
	var assignments []*BunUserBranchAssignment
	err := r.db.NewSelect().Model(&assignments).
		ModelTableExpr("user_branch_assignments AS uba").
		Column("uba.*").
		ColumnExpr("r.name AS role_name").
		ColumnExpr("r.description AS role_description").
		ColumnExpr("b.name AS branch_name").
		ColumnExpr("t.name AS tenant_name").
		Join("JOIN roles AS r ON r.id = uba.role_id").
		Join("JOIN branches AS b ON b.id = uba.branch_id").
		Join("JOIN tenants AS t ON t.id = uba.tenant_id").
		Where("uba.user_id = ? AND uba.is_active = ?", userID, true).
		Scan(ctx)
	return assignments, err
}

func (r *UserRepo) FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*BunUserBranchAssignment, error) {
	var assignments []*BunUserBranchAssignment
	err := r.db.NewSelect().Model(&assignments).
		ModelTableExpr("user_branch_assignments AS uba").
		Column("uba.*").
		ColumnExpr("r.name AS role_name").
		ColumnExpr("r.description AS role_description").
		ColumnExpr("b.name AS branch_name").
		ColumnExpr("t.name AS tenant_name").
		Join("JOIN roles AS r ON r.id = uba.role_id").
		Join("JOIN branches AS b ON b.id = uba.branch_id").
		Join("JOIN tenants AS t ON t.id = uba.tenant_id").
		Where("uba.user_id = ? AND uba.tenant_id = ? AND uba.is_active = ?", userID, tenantID, true).
		Scan(ctx)
	return assignments, err
}

func (r *UserRepo) FindTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	tenant := new(model.Tenant)
	err := r.db.NewSelect().Model(tenant).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tenant, err
}

func (r *UserRepo) FindBranchByID(ctx context.Context, id string) (*model.Branch, error) {
	branch := new(model.Branch)
	err := r.db.NewSelect().Model(branch).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return branch, err
}

func (r *UserRepo) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	role := new(model.Role)
	err := r.db.NewSelect().Model(role).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Fetch role permissions
	var permissions []*model.Permission
	err = r.db.NewSelect().
		Model(&permissions).
		Column("p.*").
		TableExpr("role_permissions AS rp").
		Join("JOIN permissions AS p ON p.id = rp.permission_id").
		Where("rp.role_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions
	return role, nil
}

func (r *UserRepo) FindPermissionByID(ctx context.Context, id string) (*model.Permission, error) {
	permission := new(model.Permission)
	err := r.db.NewSelect().Model(permission).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return permission, err
}
