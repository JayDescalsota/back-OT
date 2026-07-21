package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/clinicmanager/services/user/graph/model"
	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/cache"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/response"
	"github.com/redis/go-redis/v9"
)

type CurrentUserFn func(ctx context.Context) string

type UserService struct {
	userRepo    repository.UserRepository
	currentUser CurrentUserFn
	Cache       *redis.Client
}

func NewUserService(userRepo repository.UserRepository, currentUser CurrentUserFn, cacheClient *redis.Client) *UserService {
	return &UserService{userRepo: userRepo, currentUser: currentUser, Cache: cacheClient}
}

func (s *UserService) cacheGet(ctx context.Context, key string, dest interface{}) (bool, error) {
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

func (s *UserService) cacheSet(ctx context.Context, key string, val interface{}) error {
	if s.Cache == nil {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.Cache.Set(ctx, key, data, cache.DefaultTTL).Err()
}

func (s *UserService) cacheDel(ctx context.Context, keys ...string) {
	if s.Cache == nil {
		return
	}
	s.Cache.Del(ctx, keys...)
}

func (s *UserService) GetMe(ctx context.Context) (*model.User, error) {
	userID := s.currentUser(ctx)
	if userID == "" {
		return nil, response.Unauthorized("not authenticated")
	}

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, response.NotFound("User")
	}

	appRoles, err := s.userRepo.FindUserAppRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toUserModel(user, appRoles), nil
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

	isSuperAdmin, err := s.userRepo.HasAppRole(ctx, userID, "super_admin")
	if err != nil {
		return nil, err
	}

	if isSuperAdmin {
		logger.Info(ctx, "super admin data access",
			"adminID", userID,
			"targetID", targetID,
		)
		user, err := s.userRepo.FindUserByID(ctx, targetID)
		if err != nil || user == nil {
			return nil, response.NotFound("User")
		}
		appRoles, err := s.userRepo.FindUserAppRoles(ctx, targetID)
		if err != nil {
			return nil, err
		}
		return toUserModel(user, appRoles), nil
	}

	return nil, response.Forbidden("insufficient permissions")
}

func (s *UserService) GetAppRoleByID(ctx context.Context, id int) (*models.AppRole, error) {
	ck := cache.Key("user", "app_role", strconv.Itoa(id))
	var cached models.AppRole
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.userRepo.GetAppRoleByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *UserService) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, response.NotFound("User")
	}
	appRoles, err := s.userRepo.FindUserAppRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserModel(user, appRoles), nil
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02T15:04:05Z")
	return &s
}

func stringPtr(s string) *string {
	return &s
}

func toUserModel(u *models.User, appRoles []*models.AppRole) *model.User {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}

	roleModels := make([]*model.AppRole, len(appRoles))
	for i, r := range appRoles {
		roleModels[i] = &model.AppRole{
			ID:          strconv.Itoa(r.ID),
			Name:        r.Name,
			Description: &r.Description,
		}
	}

	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
		AppRoles:  roleModels,
	}
}
