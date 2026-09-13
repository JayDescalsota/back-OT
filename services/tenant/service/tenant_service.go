package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	"github.com/clinicmanager/services/tenant/repository"
	"github.com/clinicmanager/shared/cache"
	sharedctx "github.com/clinicmanager/shared/context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TenantService struct {
	tenantRepo repository.TenantRepository
	Cache      *redis.Client
}

func NewTenantService(tenantRepo repository.TenantRepository, cacheClient *redis.Client) *TenantService {
	return &TenantService{tenantRepo: tenantRepo, Cache: cacheClient}
}

func (s *TenantService) cacheGet(ctx context.Context, key string, dest interface{}) (bool, error) {
	if s.Cache == nil {
		return false, nil
	}
	val, err := s.Cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *TenantService) cacheSet(ctx context.Context, key string, val interface{}) error {
	if s.Cache == nil {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.Cache.Set(ctx, key, data, cache.DefaultTTL).Err()
}

func (s *TenantService) cacheDel(ctx context.Context, keys ...string) {
	if s.Cache == nil {
		return
	}
	s.Cache.Del(ctx, keys...)
}

func (s *TenantService) GetTenantByID(ctx context.Context, id string) (*db.BunTenant, error) {
	ck := cache.Key("tenant", "tenant", id)
	var cached db.BunTenant
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.tenantRepo.FindTenantByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *TenantService) GetBranchByID(ctx context.Context, id string) (*db.BunBranch, error) {
	ck := cache.Key("tenant", "branch", id)
	var cached db.BunBranch
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.tenantRepo.FindBranchByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *TenantService) UpdateBranch(ctx context.Context, id string, branch *db.BunBranch) error {
	err := s.tenantRepo.UpdateBranch(ctx, id, branch)
	if err == nil {
		s.cacheDel(ctx, cache.Key("tenant", "branch", id))
	}
	return err
}

func (s *TenantService) GetRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error) {
	ck := cache.Key("tenant", "role", id)
	var cached db.BunTenantRole
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.tenantRepo.FindRoleByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *TenantService) GetRolesByBranch(ctx context.Context, branchID string) ([]*db.BunTenantRole, error) {
	return s.tenantRepo.ListRolesByBranch(ctx, branchID)
}

func (s *TenantService) GetPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error) {
	ck := cache.Key("tenant", "permission", id)
	var cached db.BunTenantPermission
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.tenantRepo.FindPermissionByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *TenantService) GetPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error) {
	return s.tenantRepo.FindPermissionsByRole(ctx, roleID)
}

func (s *TenantService) ListPermissions(ctx context.Context, branchID string) ([]*db.BunTenantPermission, error) {
	return s.tenantRepo.ListPermissions(ctx, branchID)
}

func (s *TenantService) CreateTenantRole(ctx context.Context, name, branchID string, description *string) (*db.BunTenantRole, error) {
	branch, err := s.tenantRepo.FindBranchByID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, fmt.Errorf("branch not found")
	}
	role := &db.BunTenantRole{
		ID:           uuid.NewString(),
		Name:         name,
		Description:  description,
		TenantID:     &branch.TenantID,
		BranchID:     &branchID,
		IsSystemRole: false,
		IsActive:     true,
	}
	if ct := sharedctx.FromContext(ctx); ct.UserID != "" {
		role.CreatedBy = &ct.UserID
		role.UpdatedBy = &ct.UserID
	}
	if err := s.tenantRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *TenantService) SetRoleActive(ctx context.Context, id string, isActive bool) (*db.BunTenantRole, error) {
	role, err := s.tenantRepo.SetRoleActive(ctx, id, isActive)
	if err == nil {
		s.cacheDel(ctx, cache.Key("tenant", "role", id))
	}
	return role, err
}

func (s *TenantService) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) (*db.BunTenantRole, error) {
	role, err := s.tenantRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	// The frontend sends permission refs as "resource:action" strings; also
	// accept raw permission UUIDs for API callers.
	resolved := make([]string, 0, len(permissionIDs))
	var needLookup bool
	for _, pid := range permissionIDs {
		if strings.Contains(pid, ":") {
			needLookup = true
			break
		}
	}
	byRef := map[string]string{}
	if needLookup {
		branchID := ""
		if role.BranchID != nil {
			branchID = *role.BranchID
		}
		perms, err := s.tenantRepo.ListPermissions(ctx, branchID)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			byRef[p.Resource+":"+p.Action] = p.ID
		}
	}
	for _, pid := range permissionIDs {
		if strings.Contains(pid, ":") {
			id, ok := byRef[pid]
			if !ok {
				return nil, fmt.Errorf("unknown permission ref %q for this branch", pid)
			}
			resolved = append(resolved, id)
			continue
		}
		if perm, err := s.tenantRepo.FindPermissionByID(ctx, pid); err != nil || perm == nil {
			if err == nil {
				err = fmt.Errorf("permission not found: %s", pid)
			}
			return nil, err
		}
		resolved = append(resolved, pid)
	}
	if err := s.tenantRepo.ReplaceRolePermissions(ctx, roleID, resolved); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *TenantService) GetAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUser(ctx, userID)
}

func (s *TenantService) GetAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUserAndTenant(ctx, userID, tenantID)
}

func (s *TenantService) ListAssignmentsByBranch(ctx context.Context, branchID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByBranch(ctx, branchID)
}

func (s *TenantService) ListInvitesByBranch(ctx context.Context, branchID string) ([]*model.TenantInvite, error) {
	return s.tenantRepo.ListInvitesByBranch(ctx, branchID)
}

// checkRoleForBranch ensures a role may be granted in the given branch.
func checkRoleForBranch(role *db.BunTenantRole, branch *db.BunBranch) error {
	if role.BranchID != nil && *role.BranchID != branch.ID {
		return fmt.Errorf("role does not belong to this branch")
	}
	if role.BranchID == nil && role.TenantID != nil && *role.TenantID != branch.TenantID {
		return fmt.Errorf("role does not belong to this tenant")
	}
	return nil
}

func (s *TenantService) InviteUser(ctx context.Context, email, branchID, roleID string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, fmt.Errorf("email is required")
	}
	branch, err := s.tenantRepo.FindBranchByID(ctx, branchID)
	if err != nil {
		return false, err
	}
	if branch == nil {
		return false, fmt.Errorf("branch not found")
	}
	role, err := s.tenantRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return false, err
	}
	if role == nil {
		return false, fmt.Errorf("role not found")
	}
	if err := checkRoleForBranch(role, branch); err != nil {
		return false, err
	}
	var assignedBy *string
	if ct := sharedctx.FromContext(ctx); ct.UserID != "" {
		assignedBy = &ct.UserID
	}
	userID, err := s.tenantRepo.FindUserIDByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if userID != "" {
		// Existing account: assign immediately and close any pending invite.
		if err := s.tenantRepo.UpsertAssignment(ctx, userID, branchID, branch.TenantID, roleID, assignedBy); err != nil {
			return false, err
		}
		if err := s.tenantRepo.AcceptInvite(ctx, email, branchID); err != nil {
			return false, err
		}
		return true, nil
	}
	// No account yet: keep a pending invite (idempotent per email+branch).
	if pending, err := s.tenantRepo.FindPendingInvite(ctx, email, branchID); err != nil {
		return false, err
	} else if pending != nil {
		if _, err := s.tenantRepo.RefreshInvite(ctx, pending.ID); err != nil {
			return false, err
		}
		return true, nil
	}
	inv := &db.BunTenantInvite{
		ID:        uuid.NewString(),
		Email:     email,
		BranchID:  branchID,
		TenantID:  branch.TenantID,
		RoleID:    roleID,
		Status:    "pending",
		InvitedBy: assignedBy,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.tenantRepo.CreateInvite(ctx, inv); err != nil {
		return false, err
	}
	return true, nil
}

func (s *TenantService) ResendInvite(ctx context.Context, id string) (*model.TenantInvite, error) {
	inv, err := s.tenantRepo.FindInviteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("invite not found")
	}
	if inv.Status == "accepted" {
		return nil, fmt.Errorf("invite already accepted")
	}
	return s.tenantRepo.RefreshInvite(ctx, id)
}

func (s *TenantService) UpdateAssignment(ctx context.Context, id, roleID string) (*model.TenantUserAssignment, error) {
	current, err := s.tenantRepo.FindAssignmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("assignment not found")
	}
	role, err := s.tenantRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	if role.BranchID != nil && *role.BranchID != current.Branch.ID {
		return nil, fmt.Errorf("role does not belong to this branch")
	}
	if err := s.tenantRepo.UpdateAssignmentRole(ctx, id, roleID); err != nil {
		return nil, err
	}
	return s.tenantRepo.FindAssignmentByID(ctx, id)
}

func (s *TenantService) SetAssignmentActive(ctx context.Context, id string, isActive bool) (*model.TenantUserAssignment, error) {
	if err := s.tenantRepo.SetAssignmentActive(ctx, id, isActive); err != nil {
		return nil, err
	}
	updated, err := s.tenantRepo.FindAssignmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, fmt.Errorf("assignment not found")
	}
	return updated, nil
}

func (s *TenantService) GetAddressByID(ctx context.Context, id string) (*db.BunAddress, error) {
	return s.tenantRepo.FindAddressByID(ctx, id)
}

func (s *TenantService) CreateAddress(ctx context.Context, addr *db.BunAddress) error {
	return s.tenantRepo.CreateAddress(ctx, addr)
}

func (s *TenantService) UpdateAddress(ctx context.Context, id string, addr *db.BunAddress) error {
	return s.tenantRepo.UpdateAddress(ctx, id, addr)
}
