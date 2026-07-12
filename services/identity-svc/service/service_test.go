package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ot/identity-svc/repository"
	"github.com/ot/identity-svc/service"
)

type mockRepo struct {
	findByIDFn           func(ctx context.Context, id string) (*repository.BunUser, error)
	findByEmailFn        func(ctx context.Context, email string) (*repository.BunUser, error)
	createFn             func(ctx context.Context, email, password, name string) (*repository.BunUser, error)
	findAssignmentsByIDFn func(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
	updateLastLoginFn    func(ctx context.Context, userID string) error
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*repository.BunUser, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepo) FindByEmail(ctx context.Context, email string) (*repository.BunUser, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockRepo) Create(ctx context.Context, email, password, name string) (*repository.BunUser, error) {
	return m.createFn(ctx, email, password, name)
}
func (m *mockRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error) {
	return m.findAssignmentsByIDFn(ctx, userID)
}
func (m *mockRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	return m.updateLastLoginFn(ctx, userID)
}

func defaultMock() *mockRepo {
	return &mockRepo{
		findByIDFn:    func(_ context.Context, _ string) (*repository.BunUser, error) { return nil, nil },
		findByEmailFn: func(_ context.Context, _ string) (*repository.BunUser, error) { return nil, nil },
		createFn: func(_ context.Context, email, password, name string) (*repository.BunUser, error) {
			now := time.Now().UTC()
			hash, _ := repository.HashPassword(password)
			return &repository.BunUser{
				ID:           uuid.New().String(),
				Email:        email,
				Name:         name,
				PasswordHash: hash,
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
		findAssignmentsByIDFn: func(_ context.Context, _ string) ([]*repository.BunUserBranchAssignment, error) {
			return []*repository.BunUserBranchAssignment{}, nil
		},
		updateLastLoginFn: func(_ context.Context, _ string) error { return nil },
	}
}

func userStub(id, email, name string) *repository.BunUser {
	now := time.Now().UTC()
	return &repository.BunUser{
		ID:           id,
		Email:        email,
		Name:         name,
		PasswordHash: "",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// --- AuthService ---

func TestRegister_Success(t *testing.T) {
	mock := defaultMock()
	svc := service.NewAuthService(mock, "secret")

	payload, err := svc.Register(context.Background(), "alice@test.com", "pass123", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected payload")
	}
	if payload.User.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", payload.User.Email)
	}
	if payload.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	svc := service.NewAuthService(defaultMock(), "secret")
	_, err := svc.Register(context.Background(), "", "pass123", "Alice")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := service.NewAuthService(defaultMock(), "secret")
	_, err := svc.Register(context.Background(), "a@b.com", "12", "Alice")
	if err == nil {
		t.Fatal("expected validation error for short password")
	}
}

func TestRegister_EmptyName(t *testing.T) {
	svc := service.NewAuthService(defaultMock(), "secret")
	_, err := svc.Register(context.Background(), "a@b.com", "pass123", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mock := defaultMock()
	mock.findByEmailFn = func(_ context.Context, email string) (*repository.BunUser, error) {
		return userStub(uuid.New().String(), email, "Existing"), nil
	}
	svc := service.NewAuthService(mock, "secret")
	_, err := svc.Register(context.Background(), "dup@test.com", "pass123", "Dup")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestLogin_Success(t *testing.T) {
	mock := defaultMock()
	hash, _ := repository.HashPassword("mypass")
	mock.findByEmailFn = func(_ context.Context, email string) (*repository.BunUser, error) {
		u := userStub(uuid.New().String(), email, "Bob")
		u.PasswordHash = hash
		return u, nil
	}
	svc := service.NewAuthService(mock, "secret")

	payload, err := svc.Login(context.Background(), "bob@test.com", "mypass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected payload")
	}
	if payload.User.Email != "bob@test.com" {
		t.Errorf("expected bob@test.com, got %s", payload.User.Email)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	mock := defaultMock()
	hash, _ := repository.HashPassword("correctpass")
	mock.findByEmailFn = func(_ context.Context, email string) (*repository.BunUser, error) {
		u := userStub(uuid.New().String(), email, "Bob")
		u.PasswordHash = hash
		return u, nil
	}
	svc := service.NewAuthService(mock, "secret")
	_, err := svc.Login(context.Background(), "bob@test.com", "wrongpass")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestLogin_InactiveAccount(t *testing.T) {
	mock := defaultMock()
	hash, _ := repository.HashPassword("mypass")
	mock.findByEmailFn = func(_ context.Context, email string) (*repository.BunUser, error) {
		u := userStub(uuid.New().String(), email, "Bob")
		u.PasswordHash = hash
		u.IsActive = false
		return u, nil
	}
	svc := service.NewAuthService(mock, "secret")
	_, err := svc.Login(context.Background(), "bob@test.com", "mypass")
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := defaultMock()
	mock.findByEmailFn = func(_ context.Context, _ string) (*repository.BunUser, error) {
		return nil, nil
	}
	svc := service.NewAuthService(mock, "secret")
	_, err := svc.Login(context.Background(), "missing@test.com", "pass123")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestLogin_EmptyEmail(t *testing.T) {
	svc := service.NewAuthService(defaultMock(), "secret")
	_, err := svc.Login(context.Background(), "", "pass123")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// --- UserService ---

func TestGetMe_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*repository.BunUser, error) {
		return userStub(uid, "alice@test.com", "Alice"), nil
	}
	svc := service.NewUserService(mock)

	user, err := svc.GetMe(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", user.Email)
	}
}

func TestGetMe_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock())
	_, err := svc.GetMe(context.Background(), "")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestGetMe_NotFound(t *testing.T) {
	mock := defaultMock()
	mock.findByIDFn = func(_ context.Context, _ string) (*repository.BunUser, error) {
		return nil, errors.New("not found")
	}
	svc := service.NewUserService(mock)
	_, err := svc.GetMe(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestGetByID_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*repository.BunUser, error) {
		return userStub(uid, "bob@test.com", "Bob"), nil
	}
	svc := service.NewUserService(mock)

	user, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "bob@test.com" {
		t.Errorf("expected bob@test.com, got %s", user.Email)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	mock := defaultMock()
	mock.findByIDFn = func(_ context.Context, _ string) (*repository.BunUser, error) {
		return nil, nil
	}
	svc := service.NewUserService(mock)
	_, err := svc.GetByID(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected not found error")
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
	svc := service.NewUserService(mock)

	assignments, err := svc.GetMyAssignments(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
}

func TestGetMyAssignments_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock())
	_, err := svc.GetMyAssignments(context.Background(), "")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}
