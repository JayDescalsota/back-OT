package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/clinicmanager/services/user/models"
)

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
	CreateSession(ctx context.Context, session *models.Session) error
	FindSessionByID(ctx context.Context, id int64) (*models.Session, error)
	FindSessionByToken(ctx context.Context, token string) (*models.Session, error)
	FindSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error)
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeAllSessionsForUser(ctx context.Context, userID string) error
	UpdateSessionAccessToken(ctx context.Context, sessionID int64, accessToken string) error
	HasAppRole(ctx context.Context, userID, roleName string) (bool, error)
	FindUserAppRoles(ctx context.Context, userID string) ([]*models.AppRole, error)
	GetAppRoleByID(ctx context.Context, id int) (*models.AppRole, error)
}

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateSession(ctx context.Context, session *models.Session) error {
	_, err := r.db.NewInsert().Model(session).Returning("*").Exec(ctx)
	return err
}

func (r *UserRepo) FindSessionByID(ctx context.Context, id int64) (*models.Session, error) {
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

func (r *UserRepo) RevokeSession(ctx context.Context, sessionID int64) error {
	_, err := r.db.NewUpdate().Model((*models.Session)(nil)).
		Where("id = ?", sessionID).
		Set("revoked = true").
		Exec(ctx)
	return err
}

func (r *UserRepo) UpdateSessionAccessToken(ctx context.Context, sessionID int64, accessToken string) error {
	_, err := r.db.NewUpdate().Model((*models.Session)(nil)).
		Where("id = ?", sessionID).
		Set("access_token = ?", accessToken).
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

func (r *UserRepo) HasAppRole(ctx context.Context, userID, roleName string) (bool, error) {
	var count int
	err := r.db.NewSelect().
		Model((*models.UserAppRole)(nil)).
		ColumnExpr("COUNT(*)").
		Join("JOIN app_roles AS ar ON ar.id = user_app_role.app_role_id").
		Where("user_app_role.user_id = ?", userID).
		Where("ar.name = ?", roleName).
		Scan(ctx, &count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) FindUserAppRoles(ctx context.Context, userID string) ([]*models.AppRole, error) {
	var roles []*models.AppRole
	err := r.db.NewSelect().
		Model(&roles).
		Join("JOIN user_app_roles AS uar ON uar.app_role_id = app_role.id").
		Where("uar.user_id = ?", userID).
		Scan(ctx)
	return roles, err
}

func (r *UserRepo) GetAppRoleByID(ctx context.Context, id int) (*models.AppRole, error) {
	role := new(models.AppRole)
	err := r.db.NewSelect().Model(role).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return role, nil
}
