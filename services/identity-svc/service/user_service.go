package service

import (
	"context"
	"fmt"

	"github.com/ot/identity-svc/graph/model"
	"github.com/ot/shared/errors"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetMe(ctx context.Context, userID string) (*model.User, error) {
	if userID == "" {
		return nil, errors.Unauthorized("not authenticated")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.NotFound("User")
	}

	return toUserModel(user), nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.NotFound("User")
	}

	return toUserModel(user), nil
}

func (s *UserService) GetMyAssignments(ctx context.Context, userID string) ([]*model.UserBranchAssignment, error) {
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
