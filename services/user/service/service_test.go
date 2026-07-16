package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/services/user/service"
	"github.com/clinicmanager/services/user/graph/model"
)

type mockRepo struct {
	findByIDFn                func(ctx context.Context, id string) (*models.User, error)
	findAssignmentsByIDFn     func(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
	findAssignmentsByTenantFn func(ctx context.Context, userID, tenantID string) ([]*repository.BunUserBranchAssignment, error)
}

func (m *mockRepo) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error) {
	return m.findAssignmentsByIDFn(ctx, userID)
}
func (m *mockRepo) FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*repository.BunUserBranchAssignment, error) {
	return m.findAssignmentsByTenantFn(ctx, userID, tenantID)
}
func (m *mockRepo) FindTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	return nil, nil
}
func (m *mockRepo) FindBranchByID(ctx context.Context, id string) (*model.Branch, error) {
	return nil, nil
}
func (m *mockRepo) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	return nil, nil
}
func (m *mockRepo) FindPermissionByID(ctx context.Context, id string) (*model.Permission, error) {
	return nil, nil
}

func (m *mockRepo) Register(ctx context.Context, email, password, validationToken string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	return nil
}
func (m *mockRepo) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
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
func (m *mockRepo) FindSessionByID(ctx context.Context, id string) (*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) FindSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) FindSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	return nil, nil
}
func (m *mockRepo) RevokeSession(ctx context.Context, sessionID string) error {
	return nil
}
func (m *mockRepo) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	return nil
}

func defaultMock() *mockRepo {
	return &mockRepo{
		findByIDFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		findAssignmentsByIDFn: func(_ context.Context, _ string) ([]*repository.BunUserBranchAssignment, error) {
			return []*repository.BunUserBranchAssignment{}, nil
		},
		findAssignmentsByTenantFn: func(_ context.Context, _, _ string) ([]*repository.BunUserBranchAssignment, error) {
			return []*repository.BunUserBranchAssignment{}, nil
		},
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

func currentTenant(tenantID string) service.CurrentTenantFn {
	return func(_ context.Context) string { return tenantID }
}

func currentRole(role string) service.CurrentRoleFn {
	return func(_ context.Context) string { return role }
}

func TestGetMe_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		return userStub(uid, "alice@test.com", "Alice"), nil
	}
	svc := service.NewUserService(mock, currentUser(id), currentTenant(""), currentRole(""))

	user, err := svc.GetMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", user.Email)
	}
}

func TestGetMe_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock(), currentUser(""), currentTenant(""), currentRole(""))
	_, err := svc.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetMe_NotFound(t *testing.T) {
	mock := defaultMock()
	svc := service.NewUserService(mock, currentUser(uuid.New().String()), currentTenant(""), currentRole(""))
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
	svc := service.NewUserService(mock, currentUser(uuid.New().String()), currentTenant(""), currentRole(""))
	_, err := svc.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetByIDScopedToTenant_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		return userStub(uid, "bob@test.com", "Bob"), nil
	}
	svc := service.NewUserService(mock, currentUser(""), currentTenant("tenant-id"), currentRole(""))

	user, err := svc.GetByIDScopedToTenant(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "bob@test.com" {
		t.Errorf("expected bob@test.com, got %s", user.Email)
	}
}

func TestGetByIDScopedToTenant_NotFound(t *testing.T) {
	mock := defaultMock()
	svc := service.NewUserService(mock, currentUser(""), currentTenant("tenant-id"), currentRole(""))
	_, err := svc.GetByIDScopedToTenant(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestGetByID_SuperAdmin(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()
	targetID := uuid.New().String()

	// The current user lookup (super admin)
	superAdmin := userStub(myID, "admin@test.com", "Admin")
	superAdmin.IsSuperAdmin = true

	// Target user
	targetUser := userStub(targetID, "target@test.com", "Target")

	var callCount int
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		callCount++
		if uid == myID {
			return superAdmin, nil
		}
		if uid == targetID {
			return targetUser, nil
		}
		return nil, nil
	}

	svc := service.NewUserService(mock, currentUser(myID), currentTenant(""), currentRole("super_admin"))
	user, err := svc.GetUserByID(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "target@test.com" {
		t.Errorf("expected target@test.com, got %s", user.Email)
	}
}

func TestGetByID_NotSuperAdmin_MissingTenant(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()

	regularUser := userStub(myID, "user@test.com", "User")
	regularUser.IsSuperAdmin = false

	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		if uid == myID {
			return regularUser, nil
		}
		return nil, nil
	}

	svc := service.NewUserService(mock, currentUser(myID), currentTenant(""), currentRole(""))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error for missing tenant context")
	}
}

func TestGetByID_ClaimMismatch(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()

	// DB says super admin but JWT doesn't
	regularUser := userStub(myID, "admin@test.com", "Admin")
	regularUser.IsSuperAdmin = true

	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		if uid == myID {
			return regularUser, nil
		}
		return nil, nil
	}

	svc := service.NewUserService(mock, currentUser(myID), currentTenant("tenant-1"), currentRole(""))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error for claim mismatch")
	}
}

func TestGetByID_NotSuperAdmin_WithTenant(t *testing.T) {
	mock := defaultMock()
	myID := uuid.New().String()
	targetID := uuid.New().String()

	regularUser := userStub(myID, "user@test.com", "User")
	regularUser.IsSuperAdmin = false
	targetUser := userStub(targetID, "target@test.com", "Target")

	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		if uid == myID {
			return regularUser, nil
		}
		if uid == targetID {
			return targetUser, nil
		}
		return nil, nil
	}

	svc := service.NewUserService(mock, currentUser(myID), currentTenant("tenant-1"), currentRole(""))
	user, err := svc.GetUserByID(context.Background(), targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "target@test.com" {
		t.Errorf("expected target@test.com, got %s", user.Email)
	}
}

func TestGetByID_NotAuthenticated(t *testing.T) {
	mock := defaultMock()
	svc := service.NewUserService(mock, currentUser(""), currentTenant(""), currentRole(""))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetByID_CurrentUserNotFound(t *testing.T) {
	mock := defaultMock()
	mock.findByIDFn = func(_ context.Context, _ string) (*models.User, error) { return nil, nil }
	svc := service.NewUserService(mock, currentUser(uuid.New().String()), currentTenant(""), currentRole(""))
	_, err := svc.GetUserByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetMyAssignments_Success(t *testing.T) {
	mock := defaultMock()
	userID := uuid.New().String()
	now := time.Now()
	mock.findAssignmentsByIDFn = func(_ context.Context, uid string) ([]*repository.BunUserBranchAssignment, error) {
		return []*repository.BunUserBranchAssignment{
			{
				ID:         uuid.New().String(),
				UserID:     uid,
				BranchID:   uuid.New().String(),
				TenantID:   uuid.New().String(),
				RoleID:     uuid.New().String(),
				AssignedBy: uuid.New().String(),
				AssignedAt: now,
				IsActive:   true,
			},
		}, nil
	}
	svc := service.NewUserService(mock, currentUser(userID), currentTenant(""), currentRole(""))

	assignments, err := svc.GetMyAssignments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
}

func TestGetMyAssignments_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock(), currentUser(""), currentTenant(""), currentRole(""))
	_, err := svc.GetMyAssignments(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}
