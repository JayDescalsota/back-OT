package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type BunUser struct {
	bun.BaseModel `bun:"table:users"`

	ID           string     `bun:"id,pk"`
	Email        string     `bun:"email,notnull,unique"`
	Name         string     `bun:"name,notnull"`
	PasswordHash string     `bun:"password_hash,notnull"`
	IsActive     bool       `bun:"is_active,default:true"`
	LastLogin    *time.Time `bun:"last_login"`
	CreatedAt    time.Time  `bun:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at"`
}

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

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*BunUser, error) {
	user := new(BunUser)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, nil
	}
	return user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*BunUser, error) {
	user := new(BunUser)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, nil
	}
	return user, nil
}

func (r *UserRepo) Create(ctx context.Context, email, password, name string) (*BunUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &BunUser{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*BunUserBranchAssignment, error) {
	var assignments []*BunUserBranchAssignment
	err := r.db.NewSelect().Model(&assignments).
		Where("user_id = ? AND is_active = ?", userID, true).
		Scan(ctx)
	return assignments, err
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.db.NewUpdate().Model(&BunUser{}).
		Set("last_login = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(plain, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
