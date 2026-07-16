package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/response"
)

type AuthUser struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	IsActive  bool    `json:"isActive"`
	LastLogin *string `json:"lastLogin"`
}

type AuthPayload struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         *AuthUser `json:"user"`
}

const (
	superAdminAccessTokenTTL  = 5 * time.Minute
	superAdminRefreshTokenTTL = 24 * time.Hour
)

type Mailer interface {
	Send(to, subject, body string) error
}

type AuthService struct {
	userRepo      repository.UserRepository
	jwtSecret     string
	mailer        Mailer
	baseURL       string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, mailer Mailer, baseURL string, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		mailer:        mailer,
		baseURL:       baseURL,
		accessTokenTTL: accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
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

	existing, err := s.userRepo.FindUserByEmail(ctx, email)
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
		logger.Warn(ctx, "failed to send validation email", "error", err, "email", email)
	}

	sessionID, err := repository.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	role := ""
	accessTTL := s.accessTokenTTL
	refreshTTL := s.refreshTokenTTL
	if user.IsSuperAdmin {
		role = "super_admin"
		accessTTL = superAdminAccessTokenTTL
		refreshTTL = superAdminRefreshTokenTTL
	}

	accessToken, err := repository.GenerateToken(user.ID, sessionID, role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	err = s.userRepo.CreateSession(ctx, &models.Session{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserAgent:    "",
		IPAddress:    "",
		ExpiresAt:    time.Now().Add(refreshTTL),
		Revoked:      false,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: refreshToken, User: toAuthUserModel(user)}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*AuthPayload, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, response.Validation("Email is required")
	}
	if len(password) < 6 {
		return nil, response.Validation("Password must be at least 6 characters")
	}

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		logger.Error(ctx, "login: error fetching user", "error", err, "email", email)
		return nil, response.Unauthorized("Invalid email or password")
	}
	if user == nil {
		logger.Warn(ctx, "login: user not found", "email", email)
		return nil, response.Unauthorized("Invalid email or password")
	}

	if !user.IsActive {
		logger.Warn(ctx, "login: account inactive", "email", email, "userID", user.ID)
		return nil, response.Forbidden("Account is inactive")
	}

	if !user.IsValidated {
		logger.Warn(ctx, "login: account not validated", "email", email, "userID", user.ID)
		return nil, response.Forbidden("Account is not validated")
	}

	if !repository.VerifyPassword(password, user.PasswordHash) {
		logger.Warn(ctx, "login: password verification failed", "email", email, "userID", user.ID)
		return nil, response.Unauthorized("Invalid email or password")
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	// Generate session ID
	sessionID, err := repository.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	role := ""
	accessTTL := s.accessTokenTTL
	refreshTTL := s.refreshTokenTTL
	if user.IsSuperAdmin {
		role = "super_admin"
		accessTTL = superAdminAccessTokenTTL
		refreshTTL = superAdminRefreshTokenTTL
	}

	accessToken, err := repository.GenerateToken(user.ID, sessionID, role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	err = s.userRepo.CreateSession(ctx, &models.Session{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(refreshTTL),
		Revoked:      false,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: refreshToken, User: toAuthUserModel(user)}, nil
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

	user, err := s.userRepo.FindUserByValidationToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return response.Validation("Invalid or expired validation token")
	}

	if user.IsValidated {
		return response.Validation("Email already validated")
	}

	return s.userRepo.MarkUserAsValidated(ctx, user.ID)
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

	user, err := s.userRepo.FindUserByEmail(ctx, email)
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


func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return response.Validation("Email is required")
	}

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		logger.Info(ctx, "password reset requested for unknown email", "email", email)
		return nil
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	resetToken := hex.EncodeToString(tokenBytes)

	err = s.userRepo.SetPasswordResetToken(ctx, user.ID, resetToken, time.Now().Add(1*time.Hour))
	if err != nil {
		return response.Internal("Failed to process password reset", err.Error())
	}

	resetLink := s.baseURL + "/reset-password?token=" + resetToken
	err = s.mailer.Send(email, "Reset your password", "Reset your password by clicking the following link: "+resetLink)
	if err != nil {
		logger.Warn(ctx, "failed to send password reset email", "error", err, "email", email)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return response.Validation("Token is required")
	}
	if len(newPassword) < 6 {
		return response.Validation("Password must be at least 6 characters")
	}

	user, err := s.userRepo.FindUserByPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return response.Validation("Invalid or expired reset token")
	}

	return s.userRepo.ResetPassword(ctx, user.ID, newPassword)
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.userRepo.RevokeSession(ctx, sessionID)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	return s.userRepo.RevokeAllSessionsForUser(ctx, userID)
}


func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthPayload, error) {
	session, err := s.userRepo.FindSessionByToken(ctx, refreshToken)
	if err != nil {
		return nil, response.Unauthorized("Invalid refresh token")
	}
	if session == nil {
		return nil, response.Unauthorized("Invalid refresh token")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, response.Unauthorized("Session has expired")
	}

	// Get user to check super admin status before generating token
	user, err := s.userRepo.FindUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, response.Unauthorized("User not found")
	}

	newSessionID, err := repository.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	newRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(newRefreshTokenBytes); err != nil {
		return nil, err
	}
	newRefreshToken := hex.EncodeToString(newRefreshTokenBytes)

	role := ""
	accessTTL := s.accessTokenTTL
	refreshTTL := s.refreshTokenTTL
	if user.IsSuperAdmin {
		role = "super_admin"
		accessTTL = superAdminAccessTokenTTL
		refreshTTL = superAdminRefreshTokenTTL
	}

	accessToken, err := repository.GenerateToken(session.UserID, newSessionID, role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.RevokeSession(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	newSession := &models.Session{
		ID:           newSessionID,
		UserID:       session.UserID,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserAgent:    session.UserAgent,
		IPAddress:    session.IPAddress,
		ExpiresAt:    time.Now().Add(refreshTTL),
		Revoked:      false,
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.CreateSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: newRefreshToken, User: toAuthUserModel(user)}, nil
}
