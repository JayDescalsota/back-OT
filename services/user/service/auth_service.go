package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/response"
)

type AuthUser struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	IsActive  bool     `json:"isActive"`
	LastLogin *string  `json:"lastLogin"`
	AppRoles  []string `json:"appRoles,omitempty"`
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
	userRepo        repository.UserRepository
	jwtSecret       string
	mailer          Mailer
	baseURL         string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	// TenantSvcURL points at the tenant subgraph (server-to-server). Used to
	// verify invite tokens, which only the tenant service can resolve.
	// Proxy: nil HTTP is used so calls stay on the container network.
	TenantSvcURL string
	TenantHTTP   *http.Client
}

type tenantInvite struct {
	Email string `json:"email"`
}

// fetchInvite resolves an invite link token via the tenant service. A nil
// result means invalid, expired, or already-accepted — never distinguished.
func (s *AuthService) fetchInvite(ctx context.Context, token string) (*tenantInvite, error) {
	if token == "" || s.TenantSvcURL == "" {
		return nil, nil
	}
	client := s.TenantHTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: nil}}
	}
	const query = `query InviteByToken($token: String!) { inviteByToken(token: $token) { email } }`
	data, err := json.Marshal(map[string]interface{}{
		"query":         query,
		"variables":     map[string]interface{}{"token": token},
		"operationName": "InviteByToken",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.TenantSvcURL, "/")+"/graphql", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.Internal("Invite lookup failed", resp.Status)
	}
	var gql struct {
		Data struct {
			Invite *tenantInvite `json:"inviteByToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &gql); err != nil {
		return nil, err
	}
	return gql.Data.Invite, nil
}

// issueTokens creates a session + JWT pair for an already-persisted user.
func (s *AuthService) issueTokens(ctx context.Context, user *models.User) (*AuthPayload, error) {
	appRoleNames, err := s.getAppRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	role := s.resolveJwtRole(appRoleNames)
	accessTTL, refreshTTL := s.resolveTokenTTLs(role)

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    "",
		IPAddress:    "",
		ExpiresAt:    time.Now().Add(refreshTTL),
	}
	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	accessToken, err := repository.GenerateToken(user.ID, strconv.FormatInt(session.ID, 10), role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateSessionAccessToken(ctx, session.ID, accessToken); err != nil {
		return nil, err
	}
	return &AuthPayload{Token: accessToken, RefreshToken: refreshToken, User: toAuthUserModel(user, appRoleNames)}, nil
}

// AcceptInvite converts an invite link into an active account. The invite
// token (proof of inbox control) is verified against the tenant service, so
// the account is marked validated immediately — no separate verify step.
// Existing accounts must present the correct password to claim the invite.
func (s *AuthService) AcceptInvite(ctx context.Context, token, name, password string) (*AuthPayload, error) {
	inv, err := s.fetchInvite(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.Email == "" {
		return nil, response.Validation("Invalid or expired invite link")
	}
	email := strings.TrimSpace(strings.ToLower(inv.Email))
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, response.Validation("Name is required")
	}
	if len(password) < 8 {
		return nil, response.Validation("Password must be at least 8 characters")
	}

	existing, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	var user *models.User
	if existing != nil {
		if !repository.VerifyPassword(password, existing.PasswordHash) {
			return nil, response.Unauthorized("Invalid email or password")
		}
		user = existing
	} else {
		user, err = s.userRepo.Register(ctx, email, password, "")
		if err != nil {
			return nil, err
		}
	}
	// Token possession proves inbox control — no verify-email round trip.
	if !user.IsValidated {
		if err := s.userRepo.MarkUserAsValidated(ctx, user.ID); err != nil {
			return nil, err
		}
		user.IsValidated = true
	}
	now := time.Now().UTC()
	if err := s.userRepo.UpsertProfile(ctx, &models.UserProfile{
		UserID:    user.ID,
		FirstName: name,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		logger.Warn(ctx, "accept invite: profile update failed", "error", err, "userID", user.ID)
	}
	return s.issueTokens(ctx, user)
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, mailer Mailer, baseURL string, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		mailer:          mailer,
		baseURL:         baseURL,
		accessTokenTTL:  accessTokenTTL,
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

	appRoleNames, err := s.getAppRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	role := s.resolveJwtRole(appRoleNames)
	accessTTL, refreshTTL := s.resolveTokenTTLs(role)

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    "",
		IPAddress:    "",
		ExpiresAt:    time.Now().Add(refreshTTL),
	}
	err = s.userRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	accessToken, err := repository.GenerateToken(user.ID, strconv.FormatInt(session.ID, 10), role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateSessionAccessToken(ctx, session.ID, accessToken); err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: refreshToken, User: toAuthUserModel(user, appRoleNames)}, nil
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
		logger.Error(ctx, "login: error fetching user", "email", email, "error", err)
		return nil, response.Internal("Internal server error", err.Error())
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

	appRoleNames, err := s.getAppRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	role := s.resolveJwtRole(appRoleNames)
	accessTTL, refreshTTL := s.resolveTokenTTLs(role)

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(refreshTTL),
	}
	err = s.userRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	accessToken, err := repository.GenerateToken(user.ID, strconv.FormatInt(session.ID, 10), role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateSessionAccessToken(ctx, session.ID, accessToken); err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: refreshToken, User: toAuthUserModel(user, appRoleNames)}, nil
}

func (s *AuthService) getAppRoleNames(ctx context.Context, userID string) ([]string, error) {
	roles, err := s.userRepo.FindUserAppRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

func (s *AuthService) resolveJwtRole(appRoleNames []string) string {
	for _, name := range appRoleNames {
		if name == "super_admin" {
			return "super_admin"
		}
	}
	return ""
}

func (s *AuthService) resolveTokenTTLs(role string) (access, refresh time.Duration) {
	if role == "super_admin" {
		return superAdminAccessTokenTTL, superAdminRefreshTokenTTL
	}
	return s.accessTokenTTL, s.refreshTokenTTL
}

func toAuthUserModel(u *models.User, appRoleNames []string) *AuthUser {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}
	roles := appRoleNames
	if roles == nil {
		roles = []string{}
	}
	return &AuthUser{
		ID:        u.ID,
		Email:     u.Email,
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
		AppRoles:  roles,
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
	id, err := strconv.ParseInt(sessionID, 10, 64)
	if err != nil {
		return response.Validation("invalid session ID")
	}
	return s.userRepo.RevokeSession(ctx, id)
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

	user, err := s.userRepo.FindUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, response.Unauthorized("User not found")
	}

	appRoleNames, err := s.getAppRoleNames(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	role := s.resolveJwtRole(appRoleNames)
	accessTTL, refreshTTL := s.resolveTokenTTLs(role)

	newRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(newRefreshTokenBytes); err != nil {
		return nil, err
	}
	newRefreshToken := hex.EncodeToString(newRefreshTokenBytes)

	err = s.userRepo.RevokeSession(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	newSession := &models.Session{
		UserID:       session.UserID,
		RefreshToken: newRefreshToken,
		UserAgent:    session.UserAgent,
		IPAddress:    session.IPAddress,
		ExpiresAt:    time.Now().Add(refreshTTL),
	}
	err = s.userRepo.CreateSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	accessToken, err := repository.GenerateToken(session.UserID, strconv.FormatInt(newSession.ID, 10), role, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateSessionAccessToken(ctx, newSession.ID, accessToken); err != nil {
		return nil, err
	}

	return &AuthPayload{Token: accessToken, RefreshToken: newRefreshToken, User: toAuthUserModel(user, appRoleNames)}, nil
}
