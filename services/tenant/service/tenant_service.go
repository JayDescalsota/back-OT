package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

// Mailer sends transactional email. Failures are logged, never fatal.
type Mailer interface {
	Send(to, subject, body string) error
}

// UserCredentials is the authenticated account backing an invite acceptance.
type UserCredentials struct {
	UserID       string
	Token        string
	RefreshToken string
	Email        string
}

// UserAccountClient provisions/authenticates accounts in the user service.
// Tenant owns invites but must never touch the users table (separate database);
// all account operations go through this client (HTTP in prod, fake in tests).
// The invite token is verified inside the user service against the tenant,
// so inbox control is proven before the account is marked validated.
type UserAccountClient interface {
	AcceptInvite(ctx context.Context, token, name, password string) (*UserCredentials, error)
}

type TenantService struct {
	tenantRepo repository.TenantRepository
	Cache      *redis.Client

	// Optional integrations, wired in main.go. Nil-safe: invite flows degrade
	// to log-only email when Mailer is nil.
	Mailer       Mailer
	UserAccounts UserAccountClient
	FrontendURL  string
	UserSvcURL   string
}

func NewTenantService(tenantRepo repository.TenantRepository, cacheClient *redis.Client) *TenantService {
	return &TenantService{tenantRepo: tenantRepo, Cache: cacheClient}
}

func newInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *TenantService) inviteLink(token string) string {
	base := strings.TrimRight(s.FrontendURL, "/")
	if base == "" {
		return ""
	}
	return base + "/auth/accept-invite?token=" + token
}

// sendInviteEmail delivers the invite link. A nil Mailer or send failure only
// logs — the invite itself is already persisted and resendable from the UI.
func (s *TenantService) sendInviteEmail(ctx context.Context, email, branchName, roleName, token string) {
	if s.Mailer == nil {
		log.Printf("invite email skipped (no mailer): to=%s branch=%s", email, branchName)
		return
	}
	link := s.inviteLink(token)
	if link == "" {
		log.Printf("invite email skipped (no frontend URL): to=%s", email)
		return
	}
	subject := fmt.Sprintf("You're invited to join %s", branchName)
	body := fmt.Sprintf("You've been invited to join %s as %s.\n\nCreate your account here:\n%s\n\nThis link expires in 7 days.", branchName, roleName, link)
	if err := s.Mailer.Send(email, subject, body); err != nil {
		log.Printf("failed to send invite email: to=%s err=%v", email, err)
	} else {
		log.Printf("invite email sent: to=%s branch=%s", email, branchName)
	}
}

// userAccounts returns the configured client or a default HTTP client.
func (s *TenantService) userAccounts() UserAccountClient {
	if s.UserAccounts != nil {
		return s.UserAccounts
	}
	// Proxy: nil — service-to-service calls must stay on the docker network.
	// (The runtime injects HTTP(S)_PROXY without clinic hosts in NO_PROXY,
	// which otherwise routes internal calls to an external proxy → 502.)
	transport := &http.Transport{Proxy: nil}
	return &httpUserClient{baseURL: strings.TrimRight(s.UserSvcURL, "/"), http: &http.Client{Timeout: 10 * time.Second, Transport: transport}}
}

// httpUserClient talks to the user service's public REST + GraphQL endpoints.
type httpUserClient struct {
	baseURL string
	http    *http.Client
}

type authPayload struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func (c *httpUserClient) postJSON(ctx context.Context, path string, reqBody interface{}) (int, []byte, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, nil
}

// AcceptInvite provisions (or authenticates) the account behind an invite link
// and returns session credentials. Inbox control is proven by the token,
// which the user service re-verifies against the tenant before proceeding.
func (c *httpUserClient) AcceptInvite(ctx context.Context, token, name, password string) (*UserCredentials, error) {
	status, body, err := c.postJSON(ctx, "/invite-accept", map[string]string{"token": token, "name": name, "password": password})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("account setup failed (HTTP %d): %s", status, strings.TrimSpace(string(body)))
	}
	var p authPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.User.ID == "" || p.Token == "" {
		return nil, fmt.Errorf("account setup returned an incomplete session")
	}
	return &UserCredentials{UserID: p.User.ID, Token: p.Token, RefreshToken: p.RefreshToken, Email: p.User.Email}, nil
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
	// The tenant database cannot resolve emails to user IDs (users live in the
	// user service's database), so every invite becomes a pending invite with
	// an emailed link — including for addresses that already have accounts.
	// Idempotent per email+branch.
	pending, err := s.tenantRepo.FindPendingInvite(ctx, email, branchID)
	if err != nil {
		return false, err
	}
	var token string
	if pending != nil {
		token, err = s.tenantRepo.InviteToken(ctx, pending.ID)
		if err != nil {
			return false, err
		}
		if token == "" {
			token, err = newInviteToken()
			if err != nil {
				return false, err
			}
			if err := s.tenantRepo.SetInviteToken(ctx, pending.ID, token); err != nil {
				return false, err
			}
		}
		if _, err := s.tenantRepo.RefreshInvite(ctx, pending.ID); err != nil {
			return false, err
		}
	} else {
		token, err = newInviteToken()
		if err != nil {
			return false, err
		}
		inv := &db.BunTenantInvite{
			ID:        uuid.NewString(),
			Email:     email,
			BranchID:  branchID,
			TenantID:  branch.TenantID,
			RoleID:    roleID,
			Status:    "pending",
			Token:     &token,
			InvitedBy: assignedBy,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if err := s.tenantRepo.CreateInvite(ctx, inv); err != nil {
			return false, err
		}
	}
	s.sendInviteEmail(ctx, email, branch.Name, role.Name, token)
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
	refreshed, err := s.tenantRepo.RefreshInvite(ctx, id)
	if err != nil {
		return nil, err
	}
	token, err := s.tenantRepo.InviteToken(ctx, id)
	if err != nil {
		return nil, err
	}
	if token == "" {
		token, err = newInviteToken()
		if err != nil {
			return nil, err
		}
		if err := s.tenantRepo.SetInviteToken(ctx, id, token); err != nil {
			return nil, err
		}
	}
	s.sendInviteEmail(ctx, inv.Email, inv.Branch.Name, inv.Role.Name, token)
	return refreshed, nil
}

func (s *TenantService) InviteByToken(ctx context.Context, token string) (*model.TenantInvite, error) {
	return s.tenantRepo.FindInviteByToken(ctx, token)
}

// AcceptInvite converts a pending invite link into an account + assignment.
// The email is fixed from the invite (never client-supplied). Account
// provisioning goes through the user service; the assignment is written here.
func (s *TenantService) AcceptInvite(ctx context.Context, token, name, password string) (*model.AcceptInvitePayload, error) {
	inv, err := s.tenantRepo.FindInviteByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("invalid or expired invite link")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	creds, err := s.userAccounts().AcceptInvite(ctx, token, name, password)
	if err != nil {
		return nil, err
	}
	var assignedBy *string
	if ct := sharedctx.FromContext(ctx); ct.UserID != "" {
		assignedBy = &ct.UserID
	}
	branch, err := s.tenantRepo.FindBranchByID(ctx, inv.Branch.ID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, fmt.Errorf("branch not found")
	}
	if err := s.tenantRepo.UpsertAssignment(ctx, creds.UserID, branch.ID, branch.TenantID, inv.Role.ID, assignedBy); err != nil {
		return nil, err
	}
	if err := s.tenantRepo.AcceptInvite(ctx, inv.Email, branch.ID); err != nil {
		return nil, err
	}
	return &model.AcceptInvitePayload{
		Token:        creds.Token,
		RefreshToken: creds.RefreshToken,
		UserID:       creds.UserID,
		Email:        inv.Email,
	}, nil
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
