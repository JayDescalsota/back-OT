package service

import (
	"context"
	"fmt"

	"github.com/clinicmanager/services/user/graph/model"
	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/errors"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
}

type CurrentUserFn func(ctx context.Context) string

type UserService struct {
	userRepo    UserRepository
	currentUser CurrentUserFn
}

func NewUserService(userRepo UserRepository, currentUser CurrentUserFn) *UserService {
	return &UserService{userRepo: userRepo, currentUser: currentUser}
}

func (s *UserService) fetchUserWithAssignments(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.NotFound("User")
	}

	assignments, err := s.userRepo.FindAssignmentsByUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}

	result := toUserModel(user)
	for _, a := range assignments {
		result.Assignments = append(result.Assignments, &model.UserBranchAssignment{
			ID:         a.ID,
			UserID:     a.UserID,
			BranchID:   a.BranchID,
			TenantID:   a.TenantID,
			Role:       &model.Role{ID: a.RoleID},
			AssignedBy: a.AssignedBy,
			AssignedAt: a.AssignedAt.Format("2006-01-02T15:04:05Z"),
			IsActive:   a.IsActive,
		})
	}

	return result, nil
}

func (s *UserService) GetMe(ctx context.Context) (*model.User, error) {
	userID := s.currentUser(ctx)
	if userID == "" {
		return nil, errors.Unauthorized("not authenticated")
	}

	return s.fetchUserWithAssignments(ctx, userID)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.fetchUserWithAssignments(ctx, id)
}

func (s *UserService) GetMyAssignments(ctx context.Context) ([]*model.UserBranchAssignment, error) {
	userID := s.currentUser(ctx)
	if userID == "" {
		return nil, errors.Unauthorized("not authenticated")
	}

	assignments, err := s.userRepo.FindAssignmentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}

	result := make([]*model.UserBranchAssignment, len(assignments))
	for i, a := range assignments {
		result[i] = &model.UserBranchAssignment{
			ID:         a.ID,
			UserID:     a.UserID,
			BranchID:   a.BranchID,
			TenantID:   a.TenantID,
			Role:       &model.Role{ID: a.RoleID},
			AssignedBy: a.AssignedBy,
			AssignedAt: a.AssignedAt.Format("2006-01-02T15:04:05Z"),
			IsActive:   a.IsActive,
		}
	}

	return result, nil
}

func toUserModel(u *models.User) *model.User {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}
	return &model.User{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		IsActive:    u.IsActive,
		LastLogin:   &lastLogin,
		Assignments: []*model.UserBranchAssignment{},
	}
}
