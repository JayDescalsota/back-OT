package service

import (
	"context"
	"strings"

	"github.com/ot/identity-svc/graph/model"
	"github.com/ot/identity-svc/repository"
	"github.com/ot/shared/errors"
	"github.com/ot/shared/logger"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*repository.BunUser, error)
	FindByEmail(ctx context.Context, email string) (*repository.BunUser, error)
	Create(ctx context.Context, email, password, name string) (*repository.BunUser, error)
	FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

type AuthService struct {
	userRepo  UserRepository
	jwtSecret string
}

func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*model.AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	if email == "" {
		return nil, errors.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, errors.Validation("Password must be at least 6 characters")
	}
	if name == "" {
		return nil, errors.Validation("Name is required")
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Validation("Email already registered")
	}
	if existing != nil {
		return nil, errors.Validation("Email already registered")
	}

	user, err := s.userRepo.Create(ctx, email, password, name)
	if err != nil {
		return nil, err
	}

	token, err := repository.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "User registered", "id", user.ID, "email", user.Email)

	return &model.AuthPayload{
		Token: token,
		User:  toUserModel(user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, errors.Validation("Email is required")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.Forbidden("Account is deactivated")
	}

	if !repository.VerifyPassword(password, user.PasswordHash) {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		logger.Warn(ctx, "failed to update last_login", "error", err)
	}

	token, err := repository.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthPayload{
		Token: token,
		User:  toUserModel(user),
	}, nil
}

func toUserModel(u *repository.BunUser) *model.User {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
	}
}
