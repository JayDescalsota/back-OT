package service

import (
	"context"
	"encoding/json"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	"github.com/clinicmanager/services/tenant/repository"
	"github.com/clinicmanager/shared/cache"
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

func (s *TenantService) GetAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUser(ctx, userID)
}

func (s *TenantService) GetAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error) {
	return s.tenantRepo.FindAssignmentsByUserAndTenant(ctx, userID, tenantID)
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
