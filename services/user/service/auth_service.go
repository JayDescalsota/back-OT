package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/errors"
)

type AuthUser struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	IsActive  bool    `json:"isActive"`
	LastLogin *string `json:"lastLogin"`
}

type AuthPayload struct {
	Token string    `json:"token"`
	User  *AuthUser `json:"user"`
}

type Mailer interface {
	Send(to, subject, body string) error
}

type AuthRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Register(ctx context.Context, email, password, validationToken string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	FindByValidationToken(ctx context.Context, token string) (*models.User, error)
	MarkAsValidated(ctx context.Context, userID string) error
}

type AuthService struct {
	authRepo  AuthRepository
	jwtSecret string
	mailer    Mailer
	baseURL   string
}

func NewAuthService(authRepo AuthRepository, jwtSecret string, mailer Mailer, baseURL string) *AuthService {
	return &AuthService{
		authRepo:  authRepo,
		jwtSecret: jwtSecret,
		mailer:    mailer,
		baseURL:   baseURL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, errors.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, errors.Validation("Password must be at least 6 characters")
	}

	existing, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.Validation("Email already registered")
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	validationToken := hex.EncodeToString(tokenBytes)

	user, err := s.authRepo.Register(ctx, email, password, validationToken)
	if err != nil {
		return nil, err
	}

	validationLink := s.baseURL + "/verify?token=" + validationToken
	err = s.mailer.Send(email, "Verify your email", "Please verify your email by clicking the following link: "+validationLink)
	if err != nil {
		log.Printf("Failed to send validation email: %v", err.Error())
		return nil, errors.Internal("Failed to send validation email")
	}

	token, err := repository.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthPayload{Token: token, User: toAuthUserModel(user)}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, errors.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, errors.Validation("Password must be at least 6 characters")
	}

	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.Forbidden("Account is inactive")
	}

	if !repository.VerifyPassword(password, user.PasswordHash) {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	token, err := repository.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if err := s.authRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	return &AuthPayload{Token: token, User: toAuthUserModel(user)}, nil
}

func toAuthUserModel(u *models.User) *AuthUser {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}
	return &AuthUser{
		ID:        u.ID,
		Email:     u.Email,
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
	}
}

func (s *AuthService) ValidateEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.Validation("Token is required")
	}

	user, err := s.authRepo.FindByValidationToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.Validation("Invalid or expired validation token")
	}

	if user.IsValidated {
		return errors.Validation("Email already validated")
	}

	return s.authRepo.MarkAsValidated(ctx, user.ID)
}
