package service

import (
	"context"
	"fmt"

	"github.com/clinicmanager/services/user/graph/model"
	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/response"
)

type CurrentUserFn func(ctx context.Context) string

type CurrentTenantFn func(ctx context.Context) string

type CurrentRoleFn func(ctx context.Context) string

type UserService struct {
	userRepo      repository.UserRepository
	currentUser   CurrentUserFn
	currentTenant CurrentTenantFn
	currentRole   CurrentRoleFn
}

func NewUserService(userRepo repository.UserRepository, currentUser CurrentUserFn, currentTenant CurrentTenantFn, currentRole CurrentRoleFn) *UserService {
	return &UserService{userRepo: userRepo, currentUser: currentUser, currentTenant: currentTenant, currentRole: currentRole}
}

func (s *UserService) fetchUserWithAssignments(ctx context.Context, id string, tenantID string) (*model.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, response.NotFound("User")
	}

	var assignments []*repository.BunUserBranchAssignment
	if tenantID != "" {
		assignments, err = s.userRepo.FindAssignmentsByUserAndTenant(ctx, id, tenantID)
	} else {
		assignments, err = s.userRepo.FindAssignmentsByUser(ctx, id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}

	result := toUserModel(user)
	for _, a := range assignments {
		result.Assignments = append(result.Assignments, &model.UserBranchAssignment{
			ID: a.ID,
			Branch: &model.Branch{
				ID:   a.BranchID,
				Name: a.BranchName,
			},
			Tenant: &model.Tenant{
				ID:   a.TenantID,
				Name: a.TenantName,
			},
			Role: &model.Role{
				ID:          a.RoleID,
				Name:        a.RoleName,
				Description: &a.RoleDescription,
			},
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
		return nil, response.Unauthorized("not authenticated")
	}

	return s.fetchUserWithAssignments(ctx, userID, "")
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	userID := s.currentUser(ctx)
	if userID == "" {
		return nil, response.Unauthorized("not authenticated")
	}

	targetID := id
	if targetID == "" {
		return nil, response.Validation("User ID is required")
	}

	currentUser, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, response.Unauthorized("not authenticated")
	}

	roleClaim := s.currentRole(ctx)
	isSuperAdmin := currentUser.IsSuperAdmin && roleClaim == "super_admin"

	if currentUser.IsSuperAdmin != (roleClaim == "super_admin") {
		logger.Error(ctx, "super admin claim mismatch: possible token tampering",
			"userID", userID,
			"dbIsSuperAdmin", currentUser.IsSuperAdmin,
			"jwtRole", roleClaim,
			"targetID", targetID,
		)
		return nil, response.Unauthorized("invalid token claims")
	}

	if isSuperAdmin {
		logger.Info(ctx, "super admin data access",
			"adminID", userID,
			"targetID", targetID,
		)
		return s.fetchUserWithAssignments(ctx, targetID, "")
	}

	tenantID := s.currentTenant(ctx)
	if tenantID == "" {
		return nil, response.Unauthorized("tenant context required")
	}

	return s.fetchUserWithAssignments(ctx, targetID, tenantID)
}

func (s *UserService) GetByIDScopedToTenant(ctx context.Context, id string) (*model.User, error) {
	tenantID := s.currentTenant(ctx)
	if tenantID == "" {
		return nil, response.Unauthorized("tenant context required")
	}

	return s.fetchUserWithAssignments(ctx, id, tenantID)
}

func (s *UserService) GetMyAssignments(ctx context.Context) ([]*model.UserBranchAssignment, error) {
	userID := s.currentUser(ctx)
	if userID == "" {
		return nil, response.Unauthorized("not authenticated")
	}

	assignments, err := s.userRepo.FindAssignmentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}

	result := make([]*model.UserBranchAssignment, len(assignments))
	for i, a := range assignments {
		result[i] = &model.UserBranchAssignment{
			ID: a.ID,
			Branch: &model.Branch{
				ID:   a.BranchID,
				Name: a.BranchName,
			},
			Tenant: &model.Tenant{
				ID:   a.TenantID,
				Name: a.TenantName,
			},
			Role: &model.Role{
				ID:          a.RoleID,
				Name:        a.RoleName,
				Description: &a.RoleDescription,
			},
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

func (s *UserService) GetTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	return s.userRepo.FindTenantByID(ctx, id)
}

func (s *UserService) GetBranchByID(ctx context.Context, id string) (*model.Branch, error) {
	return s.userRepo.FindBranchByID(ctx, id)
}

func (s *UserService) GetRoleByID(ctx context.Context, id string) (*model.Role, error) {
	return s.userRepo.FindRoleByID(ctx, id)
}

func (s *UserService) GetPermissionByID(ctx context.Context, id string) (*model.Permission, error) {
	return s.userRepo.FindPermissionByID(ctx, id)
}
