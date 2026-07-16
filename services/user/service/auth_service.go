package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/response"
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

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	mailer    Mailer
	baseURL   string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, mailer Mailer, baseURL string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		mailer:    mailer,
		baseURL:   baseURL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, response.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, response.Validation("Password must be at least 6 characters")
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, response.Validation("Email already registered")
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	validationToken := hex.EncodeToString(tokenBytes)

	user, err := s.userRepo.Register(ctx, email, password, validationToken)
	if err != nil {
		return nil, err
	}

	validationLink := s.baseURL + "/verify?token=" + validationToken
	err = s.mailer.Send(email, "Verify your email", "Please verify your email by clicking the following link: "+validationLink)
	if err != nil {
		log.Printf("[AuthService.Register] WARNING: Failed to send validation email: %v", err)
		log.Printf("[AuthService.Register] Local validation link: %s", validationLink)
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
		return nil, response.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, response.Validation("Password must be at least 6 characters")
	}

	log.Printf("[AuthService.Login] Attempting login for: %s", email)
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		log.Printf("[AuthService.Login] Error fetching user: %v", err)
		return nil, response.Unauthorized("Invalid email or password")
	}
	if user == nil {
		log.Printf("[AuthService.Login] User not found for email: %s", email)
		return nil, response.Unauthorized("Invalid email or password")
	}

	log.Printf("[AuthService.Login] User record retrieved: ID=%s, Active=%t, Hash=%q", user.ID, user.IsActive, user.PasswordHash)

	if !user.IsActive {
		log.Printf("[AuthService.Login] Account inactive for: %s", email)
		return nil, response.Forbidden("Account is inactive")
	}

	if !user.IsValidated {
		log.Printf("[AuthService.Login] Account not validated for: %s", email)
		return nil, response.Forbidden("Account is not validated")
	}

	if !repository.VerifyPassword(password, user.PasswordHash) {
		log.Printf("[AuthService.Login] Password verification failed for: %s", email)
		return nil, response.Unauthorized("Invalid email or password")
	}

	log.Printf("[AuthService.Login] Password verified successfully for: %s", email)

	token, err := repository.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
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
		return response.Validation("Token is required")
	}

	user, err := s.userRepo.FindByValidationToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return response.Validation("Invalid or expired validation token")
	}

	if user.IsValidated {
		return response.Validation("Email already validated")
	}

	return s.userRepo.MarkAsValidated(ctx, user.ID)
}

func (s *AuthService) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return response.Validation("Email is required")
	}
	if len(oldPassword) < 6 || len(newPassword) < 6 {
		return response.Validation("Password must be at least 6 characters")
	}
	if oldPassword == newPassword {
		return response.Validation("Old password and new password must be different")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil {
		return response.Validation("Invalid credentials")
	}

	if !user.IsValidated {
		return response.Forbidden("Account is not validated")
	}

	if !user.IsActive {
		return response.Forbidden("Account is inactive")
	}

	if !repository.VerifyPassword(oldPassword, user.PasswordHash) {
		return response.Unauthorized("Invalid credentials")
	}

	err = s.userRepo.UpdatePassword(ctx, user.ID, newPassword)
	if err != nil {
		return response.Internal("Failed to update password", err.Error())
	}

	return nil
}
