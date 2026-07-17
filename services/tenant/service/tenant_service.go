package service

import (
	"context"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	"github.com/clinicmanager/services/tenant/repository"
)

type TenantService struct {
	tenantRepo repository.TenantRepository
}

func NewTenantService(tenantRepo repository.TenantRepository) *TenantService {
	return &TenantService{tenantRepo: tenantRepo}
}

func (s *TenantService) GetTenantByID(ctx context.Context, id string) (*db.BunTenant, error) {
	return s.tenantRepo.FindTenantByID(ctx, id)
}

func (s *TenantService) GetBranchByID(ctx context.Context, id string) (*db.BunBranch, error) {
	return s.tenantRepo.FindBranchByID(ctx, id)
}

func (s *TenantService) GetRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error) {
	return s.tenantRepo.FindRoleByID(ctx, id)
}

func (s *TenantService) GetPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error) {
	return s.tenantRepo.FindPermissionByID(ctx, id)
}

func (s *TenantService) GetPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error) {
	return s.tenantRepo.FindPermissionsByRole(ctx, roleID)
}

func (s *TenantService) GetAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUser(ctx, userID)
}

func (s *TenantService) GetAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUserAndTenant(ctx, userID, tenantID)
}
