package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/service"
)

type mockRepo struct {
	findByIDFn       func(ctx context.Context, id string) (*models.User, error)
	findUserAppRoles func(ctx context.Context, userID string) ([]*models.AppRole, error)
	hasAppRoleFn     func(ctx context.Context, userID, roleName string) (bool, error)
}

func (m *mockRepo) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepo) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) Register(ctx context.Context, email, password, validationToken string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	return nil
}
func (m *mockRepo) FindUserByValidationToken(ctx context.Context, token string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) MarkUserAsValidated(ctx context.Context, userID string) error {
	return nil
}
func (m *mockRepo) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	return nil
}
func (m *mockRepo) SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	return nil
}
func (m *mockRepo) FindUserByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) ResetPassword(ctx context.Context, userID, newPassword string) error {
	return nil
}
func (m *mockRepo) CreateSession(ctx context.Context, session *models.Session) error {
	return nil
}
func (m *mockRepo) FindSessionByID(ctx context.Context, id int64) (*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) FindSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) FindSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) RevokeSession(ctx context.Context, sessionID int64) error {
	return nil
}
func (m *mockRepo) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	return nil
}
func (m *mockRepo) UpdateSessionAccessToken(ctx context.Context, sessionID int64, accessToken string) error {
	return nil
}
func (m *mockRepo) HasAppRole(ctx context.Context, userID, roleName string) (bool, error) {
	if m.hasAppRoleFn != nil {
		return m.hasAppRoleFn(ctx, userID, roleName)
	}
	return false, nil
}
func (m *mockRepo) FindUserAppRoles(ctx context.Context, userID string) ([]*models.AppRole, error) {
	if m.findUserAppRoles != nil {
		return m.findUserAppRoles(ctx, userID)
	}
	return []*models.AppRole{}, nil
}
func (m *mockRepo) GetAppRoleByID(ctx context.Context, id int) (*models.AppRole, error) {
	return nil, nil
}

func defaultMock() *mockRepo {
	return &mockRepo{
		findByIDFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
	}
}

func userStub(id, email, name string) *models.User {
	now := time.Now().UTC()
	return &models.User{
		ID:        id,
		Email:     email,
		Name:      name,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func currentUser(userID string) service.CurrentUserFn {
	return func(_ context.Context) string { return userID }
}

func TestGetMe_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		return userStub(uid, "alice@test.com", "Alice"), nil
	}
	svc := service.NewUserService(mock, currentUser(id))

	user, err := svc.GetMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", user.Email)
	}
}

func TestGetMe_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock(), currentUser(""))
	_, err := svc.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetMe_NotFound(t *testing.T) {
	mock := defaultMock()
	svc := service.NewUserService(mock, currentUser(uuid.New().String()))
	_, err := svc.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestGetMe_RepoError(t *testing.T) {
	mock := defaultMock()
	mock.findByIDFn = func(_ context.Context, _ string) (*models.User, error) {
		return nil, errors.New("db error")
	}
	svc := service.NewUserService(mock, currentUser(uuid.New().String()))
	_, err := svc.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserByID_NotAuthenticated(t *testing.T) {
	mock := defaultMock()
	svc := service.NewUserService(mock, currentUser(""))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetUserByID_SuperAdmin(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()
	targetID := uuid.New().String()

	targetUser := userStub(targetID, "target@test.com", "Target")

	var callCount int
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		callCount++
		if uid == myID {
			return userStub(myID, "admin@test.com", "Admin"), nil
		}
		if uid == targetID {
			return targetUser, nil
		}
		return nil, nil
	}
	mock.hasAppRoleFn = func(_ context.Context, uid, role string) (bool, error) {
		if uid == myID && role == "super_admin" {
			return true, nil
		}
		return false, nil
	}

	svc := service.NewUserService(mock, currentUser(myID))
	user, err := svc.GetUserByID(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "target@test.com" {
		t.Errorf("expected target@test.com, got %s", user.Email)
	}
}

func TestGetUserByID_NotSuperAdmin(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()

	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		if uid == myID {
			return userStub(myID, "user@test.com", "User"), nil
		}
		return nil, nil
	}
	mock.hasAppRoleFn = func(_ context.Context, uid, role string) (bool, error) {
		return false, nil
	}

	svc := service.NewUserService(mock, currentUser(myID))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected forbidden error for non-super-admin")
	}
}

func TestGetUserByID_CurrentUserNotFound(t *testing.T) {
	mock := defaultMock()
	mock.findByIDFn = func(_ context.Context, _ string) (*models.User, error) { return nil, nil }
	svc := service.NewUserService(mock, currentUser(uuid.New().String()))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}
