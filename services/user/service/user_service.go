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

func (s *UserService) GetCurrentUserID(ctx context.Context) string {
	return s.currentUser(ctx)
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

	result := toUserModel(user, appRoles)
	profile, err := s.userRepo.FindProfileByUserID(ctx, userID)
	if err == nil && profile != nil {
		result.Profile = toProfileModel(profile)
	}
	return result, nil
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
		result := toUserModel(user, appRoles)
		profile, _ := s.userRepo.FindProfileByUserID(ctx, targetID)
		if profile != nil {
			result.Profile = toProfileModel(profile)
		}
		return result, nil
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
	result := toUserModel(user, appRoles)
	profile, err := s.userRepo.FindProfileByUserID(ctx, id)
	if err == nil && profile != nil {
		result.Profile = toProfileModel(profile)
	}
	return result, nil
}

func (s *UserService) ListUsers(ctx context.Context, ids []string) ([]*model.User, error) {
	var users []*models.User
	var err error
	if len(ids) > 0 {
		users, err = s.userRepo.FindUsersByIDs(ctx, ids)
	} else {
		users, err = s.userRepo.FindAllUsers(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*model.User, 0, len(users))
	for _, u := range users {
		roles, err := s.userRepo.FindUserAppRoles(ctx, u.ID)
		if err != nil {
			continue
		}
		m := toUserModel(u, roles)
		profile, _ := s.userRepo.FindProfileByUserID(ctx, u.ID)
		if profile != nil {
			m.Profile = toProfileModel(profile)
		}
		result = append(result, m)
	}
	return result, nil
}

func (s *UserService) CreateStaffUser(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.userRepo.CreateUser(ctx, email, password)
	if err != nil {
		return nil, err
	}
	// Assign 'user' app role (role ID 4 from seed data)
	if err := s.userRepo.AssignAppRole(ctx, user.ID, 4); err != nil {
		return nil, err
	}
	appRoles, err := s.userRepo.FindUserAppRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	result := toUserModel(user, appRoles)
	profile, _ := s.userRepo.FindProfileByUserID(ctx, user.ID)
	if profile != nil {
		result.Profile = toProfileModel(profile)
	}
	return result, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, isActive bool) (*model.User, error) {
	if err := s.userRepo.UpdateUser(ctx, id, isActive); err != nil {
		return nil, err
	}
	return s.FindUserByID(ctx, id)
}

func (s *UserService) UpsertProfile(ctx context.Context, userID string, input model.UserProfileInput) (*model.UserProfile, error) {
	profile := &models.UserProfile{
		UserID:    userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if input.MiddleName != nil {
		profile.MiddleName = input.MiddleName
	}
	if input.Suffix != nil {
		profile.Suffix = input.Suffix
	}
	if input.Title != nil {
		profile.Title = input.Title
	}
	if input.Phone != nil {
		profile.Phone = input.Phone
	}
	if input.Mobile != nil {
		profile.Mobile = input.Mobile
	}
	if input.DateOfBirth != nil {
		t, _ := time.Parse("2006-01-02", *input.DateOfBirth)
		profile.DateOfBirth = &t
	}
	if input.Gender != nil {
		profile.Gender = input.Gender
	}
	if input.AddressID != nil {
		profile.AddressID = input.AddressID
	}
	if input.Timezone != nil {
		profile.Timezone = *input.Timezone
	}
	if input.PreferredLanguage != nil {
		profile.PreferredLanguage = *input.PreferredLanguage
	}
	if err := s.userRepo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}
	return toProfileModel(profile), nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	p, err := s.userRepo.FindProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return toProfileModel(p), nil
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
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
		AppRoles:  roleModels,
	}
}

func toProfileModel(p *models.UserProfile) *model.UserProfile {
	dateOfBirth := ""
	if p.DateOfBirth != nil {
		dateOfBirth = p.DateOfBirth.Format("2006-01-02")
	}
	var address *model.Address
	if p.AddressID != nil {
		address = &model.Address{ID: *p.AddressID}
	}
	return &model.UserProfile{
		UserID:                   p.UserID,
		FirstName:                p.FirstName,
		LastName:                 p.LastName,
		MiddleName:               p.MiddleName,
		Suffix:                   p.Suffix,
		Title:                    p.Title,
		Phone:                    p.Phone,
		Mobile:                   p.Mobile,
		DateOfBirth:              &dateOfBirth,
		Gender:                   p.Gender,
		Address:                  address,
		Timezone:                 &p.Timezone,
		PreferredLanguage:        &p.PreferredLanguage,
		EmergencyContactName:     p.EmergencyContactName,
		EmergencyContactPhone:    p.EmergencyContactPhone,
		EmergencyContactRelation: p.EmergencyContactRelation,
		Notes:                    p.Notes,
	}
}
