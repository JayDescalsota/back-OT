package repository

import (
	"context"
	"database/sql"
	"time"

	sharedctx "github.com/clinicmanager/shared/context"
	shareddb "github.com/clinicmanager/shared/db"
	"github.com/google/uuid"

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
	FindAllUsers(ctx context.Context) ([]*models.User, error)
	FindAllAppRoles(ctx context.Context) ([]*models.AppRole, error)
	CreateUser(ctx context.Context, email, password string) (*models.User, error)
	AssignAppRole(ctx context.Context, userID string, appRoleID int) error
	UpdateUser(ctx context.Context, id string, isActive bool) error
	FindProfileByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	UpsertProfile(ctx context.Context, profile *models.UserProfile) error
	FindPractitionerProfileByUserID(ctx context.Context, userID string) (*models.PractitionerProfile, error)
	UpsertPractitionerProfile(ctx context.Context, profile *models.PractitionerProfile) error
}

type UserRepo struct {
	// db: default scoped DB (ScopeBranch).
	// Users and sessions currently have no tenant_id/branch_id columns so scoping is skipped silently.
	// Using r.db as default ensures that if tenant_id or branch_id columns are ever added in the future,
	// scoping will automatically be applied with zero code changes needed.
	db    *shareddb.ScopedDB
	alldb *shareddb.ScopedDB
}

func NewUserRepo(dbs *shareddb.DBSet) *UserRepo {
	return &UserRepo{
		db:    dbs.DB,
		alldb: dbs.AllDB,
	}
}

func (r *UserRepo) CreateSession(ctx context.Context, session *models.Session) error {
	_, err := r.db.NewInsert(session).Returning("*").Exec(ctx)
	return err
}

func (r *UserRepo) FindSessionByID(ctx context.Context, id int64) (*models.Session, error) {
	var session models.Session
	err := r.alldb.NewSelect(ctx, &session).Where("id = ?", id).Scan(ctx)
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
	err := r.alldb.NewSelect(ctx, &session).
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
	err := r.alldb.NewSelect(ctx, &sessions).
		Where("user_id = ?", userID).
		Where("revoked = false").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *UserRepo) RevokeSession(ctx context.Context, sessionID int64) error {
	_, err := r.db.NewUpdate(ctx, (*models.Session)(nil)).
		Where("id = ?", sessionID).
		Set("revoked = true").
		Exec(ctx)
	return err
}

func (r *UserRepo) UpdateSessionAccessToken(ctx context.Context, sessionID int64, accessToken string) error {
	_, err := r.db.NewUpdate(ctx, (*models.Session)(nil)).
		Where("id = ?", sessionID).
		Set("access_token = ?", accessToken).
		Exec(ctx)
	return err
}

func (r *UserRepo) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate(ctx, (*models.Session)(nil)).
		Where("user_id = ?", userID).
		Set("revoked = true").
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect(ctx, user).Where("id = ?", id).Scan(ctx)
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
	err := r.db.NewSelect(ctx, user).Where("email = ?", email).Scan(ctx)
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
	uid := uuid.New()
	now := time.Now().UTC()
	user := &models.User{
		ID:              uid.String(),
		Email:           email,
		PasswordHash:    hash,
		IsActive:        true,
		ValidationToken: validationToken,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err = r.db.NewInsert(user).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.db.NewUpdate(ctx, (*models.User)(nil)).
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
	_, err = r.db.NewUpdate(ctx, (*models.User)(nil)).
		Set("password_hash = ?", hash).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByValidationToken(ctx context.Context, token string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect(ctx, user).Where("validation_token = ?", token).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.db.NewUpdate(ctx, (*models.User)(nil)).
		Set("password_reset_token = ?", token).
		Set("password_reset_expires_at = ?", expiresAt).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect(ctx, user).
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
	_, err = r.db.NewUpdate(ctx, (*models.User)(nil)).
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
	_, err := r.db.NewUpdate(ctx, (*models.User)(nil)).
		Set("is_validated = ?", true).
		Set("validated_at = ?", now).
		Set("validation_token = ?", "").
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

func (r *UserRepo) HasAppRole(ctx context.Context, userID, roleName string) (bool, error) {
	var count int
	err := r.alldb.NewSelect(ctx, (*models.UserAppRole)(nil)).
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
	err := r.alldb.NewSelect(ctx, &roles).
		Join("JOIN user_app_roles AS uar ON uar.app_role_id = app_role.id").
		Where("uar.user_id = ?", userID).
		Scan(ctx)
	return roles, err
}

func (r *UserRepo) GetAppRoleByID(ctx context.Context, id int) (*models.AppRole, error) {
	role := new(models.AppRole)
	err := r.alldb.NewSelect(ctx, role).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserRepo) FindAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	tctx := sharedctx.FromContext(ctx)
	query := r.alldb.NewSelect(ctx, &users).
		Distinct().
		Join("JOIN tenant_user_assignments AS tua ON tua.user_id = users.id").
		Where("tua.is_active = ?", true)
	if tctx.TenantID != "" {
		query = query.Where("tua.tenant_id = ?", tctx.TenantID)
	}
	if tctx.BranchID != "" {
		query = query.Where("tua.branch_id = ?", tctx.BranchID)
	}
	err := query.Order("users.created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) FindAllAppRoles(ctx context.Context) ([]*models.AppRole, error) {
	var roles []*models.AppRole
	err := r.alldb.NewSelect(ctx, &roles).Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	uid := uuid.New()
	now := time.Now().UTC()
	user := &models.User{
		ID:           uid.String(),
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err = r.db.NewInsert(user).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) AssignAppRole(ctx context.Context, userID string, appRoleID int) error {
	_, err := r.db.NewInsert(&models.UserAppRole{
		UserID:    userID,
		AppRoleID: appRoleID,
	}).Exec(ctx)
	return err
}

func (r *UserRepo) UpdateUser(ctx context.Context, id string, isActive bool) error {
	now := time.Now().UTC()
	_, err := r.db.NewUpdate(ctx, (*models.User)(nil)).
		Set("is_active = ?", isActive).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *UserRepo) FindProfileByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	p := new(models.UserProfile)
	err := r.db.NewSelect(ctx, p).Where("user_id = ?", userID).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *UserRepo) UpsertProfile(ctx context.Context, profile *models.UserProfile) error {
	_, err := r.db.NewInsert(profile).
		On("CONFLICT (user_id) DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, middle_name = EXCLUDED.middle_name, suffix = EXCLUDED.suffix, title = EXCLUDED.title, phone = EXCLUDED.phone, mobile = EXCLUDED.mobile, date_of_birth = EXCLUDED.date_of_birth, gender = EXCLUDED.gender, street_address = EXCLUDED.street_address, barangay = EXCLUDED.barangay, city = EXCLUDED.city, province = EXCLUDED.province, zip = EXCLUDED.zip, country = EXCLUDED.country, timezone = EXCLUDED.timezone, preferred_language = EXCLUDED.preferred_language, emergency_contact_name = EXCLUDED.emergency_contact_name, emergency_contact_phone = EXCLUDED.emergency_contact_phone, emergency_contact_relation = EXCLUDED.emergency_contact_relation, notes = EXCLUDED.notes, updated_at = EXCLUDED.updated_at, updated_by = EXCLUDED.updated_by, updated_action = EXCLUDED.updated_action").
		Exec(ctx)
	return err
}

func (r *UserRepo) FindPractitionerProfileByUserID(ctx context.Context, userID string) (*models.PractitionerProfile, error) {
	p := new(models.PractitionerProfile)
	err := r.db.NewSelect(ctx, p).Where("user_id = ?", userID).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *UserRepo) UpsertPractitionerProfile(ctx context.Context, profile *models.PractitionerProfile) error {
	_, err := r.db.NewInsert(profile).
		On("CONFLICT (user_id) DO UPDATE SET license_number = EXCLUDED.license_number, license_state = EXCLUDED.license_state, npi_number = EXCLUDED.npi_number, specialty = EXCLUDED.specialty, sub_specialty = EXCLUDED.sub_specialty, qualifications = EXCLUDED.qualifications, credentials = EXCLUDED.credentials, education = EXCLUDED.education, years_of_experience = EXCLUDED.years_of_experience, bio = EXCLUDED.bio, is_accepting_patients = EXCLUDED.is_accepting_patients, updated_at = EXCLUDED.updated_at, updated_by = EXCLUDED.updated_by, updated_action = EXCLUDED.updated_action").
		Exec(ctx)
	return err
}
