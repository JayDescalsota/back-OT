package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

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
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Register(ctx context.Context, email, password, validationToken string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	FindByValidationToken(ctx context.Context, token string) (*models.User, error)
	MarkAsValidated(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, newPassword string) error
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*BunUserBranchAssignment, error)
}

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
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

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
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

func (r *UserRepo) FindByValidationToken(ctx context.Context, token string) (*models.User, error) {
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

func (r *UserRepo) MarkAsValidated(ctx context.Context, userID string) error {
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
		Where("user_id = ? AND is_active = ?", userID, true).
		Scan(ctx)
	return assignments, err
}
