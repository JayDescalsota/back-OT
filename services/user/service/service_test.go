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
)

type mockRepo struct {
	findByIDFn            func(ctx context.Context, id string) (*models.User, error)
	findAssignmentsByIDFn func(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error) {
	return m.findAssignmentsByIDFn(ctx, userID)
}

func defaultMock() *mockRepo {
	return &mockRepo{
		findByIDFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		findAssignmentsByIDFn: func(_ context.Context, _ string) ([]*repository.BunUserBranchAssignment, error) {
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

func TestGetByID_Success(t *testing.T) {
	mock := defaultMock()
	id := uuid.New().String()
	mock.findByIDFn = func(_ context.Context, uid string) (*models.User, error) {
		return userStub(uid, "bob@test.com", "Bob"), nil
	}
	svc := service.NewUserService(mock, currentUser(""))

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
	svc := service.NewUserService(mock, currentUser(""))
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
	svc := service.NewUserService(mock, currentUser(userID))

	assignments, err := svc.GetMyAssignments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
}

func TestGetMyAssignments_EmptyID(t *testing.T) {
	svc := service.NewUserService(defaultMock(), currentUser(""))
	_, err := svc.GetMyAssignments(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}
