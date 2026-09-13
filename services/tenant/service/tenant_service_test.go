package service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/clinicmanager/services/tenant/db"
	"github.com/clinicmanager/services/tenant/graph/model"
	"github.com/clinicmanager/services/tenant/service"
)

// fakeRepo implements repository.TenantRepository with per-test hooks.
// Unset hooks return zero values so tests only configure what they need.
type fakeRepo struct {
	findBranch func(ctx context.Context, id string) (*db.BunBranch, error)
	findRole   func(ctx context.Context, id string) (*db.BunTenantRole, error)

	createdRole *db.BunTenantRole
	createRole  func(role *db.BunTenantRole) error

	setRoleActive func(id string, active bool) (*db.BunTenantRole, error)

	listPerms     func(branchID string) ([]*db.BunTenantPermission, error)
	findPerm      func(ctx context.Context, id string) (*db.BunTenantPermission, error)
	replacedPerms []string
	replacePerms  func(roleID string, permIDs []string) error

	upserted *upsertCall
	accepts  [][2]string
	pending  *model.TenantInvite
	byToken  *model.TenantInvite
	created  *db.BunTenantInvite

	tokens    map[string]string
	setTokens [][2]string
	refreshed []string

	assignment  *model.TenantUserAssignment
	updatedRole string
	activated   *activateCall
}

type upsertCall struct {
	userID, branchID, tenantID, roleID string
	assignedBy                         *string
}

type activateCall struct {
	id       string
	isActive bool
}

func strptr(s string) *string { return &s }

func testBranch() *db.BunBranch {
	return &db.BunBranch{ID: "b1", TenantID: "t1", Name: "Branch 1"}
}

func testRole() *db.BunTenantRole {
	return &db.BunTenantRole{ID: "r1", Name: "therapist", TenantID: strptr("t1"), BranchID: strptr("b1"), IsSystemRole: true, IsActive: true}
}

func testAssignment() *model.TenantUserAssignment {
	return &model.TenantUserAssignment{
		ID:       "a1",
		Branch:   &db.BunBranch{ID: "b1", Name: "Branch 1"},
		Tenant:   &db.BunTenant{ID: "t1", Name: "Clinic"},
		Role:     &db.BunTenantRole{ID: "r1", Name: "therapist"},
		IsActive: true,
	}
}

func testInvite() *model.TenantInvite {
	return &model.TenantInvite{
		ID:        "inv1",
		Email:     "new@clinic.com",
		Branch:    &db.BunBranch{ID: "b1", Name: "Branch 1"},
		Role:      &db.BunTenantRole{ID: "r1", Name: "therapist"},
		Status:    "pending",
		InvitedAt: "2026-01-01T00:00:00Z",
		ExpiresAt: "2026-01-08T00:00:00Z",
	}
}

type mailCall struct {
	to, subject, body string
}

type fakeMailer struct {
	sends []mailCall
	err   error
}

func (m *fakeMailer) Send(to, subject, body string) error {
	m.sends = append(m.sends, mailCall{to: to, subject: subject, body: body})
	return m.err
}

type fakeAccounts struct {
	creds *service.UserCredentials
	err   error
	calls [][3]string
}

func (f *fakeAccounts) AcceptInvite(ctx context.Context, token, name, password string) (*service.UserCredentials, error) {
	f.calls = append(f.calls, [3]string{token, name, password})
	if f.err != nil {
		return nil, f.err
	}
	return f.creds, nil
}

func (f *fakeRepo) FindTenantByID(ctx context.Context, id string) (*db.BunTenant, error) {
	return nil, nil
}
func (f *fakeRepo) FindTenantBySlug(ctx context.Context, slug string) (*db.BunTenant, error) {
	return nil, nil
}
func (f *fakeRepo) ListTenants(ctx context.Context) ([]*db.BunTenant, error) {
	return nil, nil
}
func (f *fakeRepo) FindBranchByID(ctx context.Context, id string) (*db.BunBranch, error) {
	if f.findBranch != nil {
		return f.findBranch(ctx, id)
	}
	return testBranch(), nil
}
func (f *fakeRepo) UpdateBranch(ctx context.Context, id string, branch *db.BunBranch) error {
	return nil
}
func (f *fakeRepo) FindBranchesByTenant(ctx context.Context, tenantID string) ([]*db.BunBranch, error) {
	return nil, nil
}
func (f *fakeRepo) FindRoleByID(ctx context.Context, id string) (*db.BunTenantRole, error) {
	if f.findRole != nil {
		return f.findRole(ctx, id)
	}
	return testRole(), nil
}
func (f *fakeRepo) ListSystemRoles(ctx context.Context) ([]*db.BunTenantRole, error) {
	return nil, nil
}
func (f *fakeRepo) ListTenantRoles(ctx context.Context, tenantID string) ([]*db.BunTenantRole, error) {
	return nil, nil
}
func (f *fakeRepo) ListRolesByBranch(ctx context.Context, branchID string) ([]*db.BunTenantRole, error) {
	return []*db.BunTenantRole{testRole()}, nil
}
func (f *fakeRepo) FindPermissionByID(ctx context.Context, id string) (*db.BunTenantPermission, error) {
	if f.findPerm != nil {
		return f.findPerm(ctx, id)
	}
	return &db.BunTenantPermission{ID: id}, nil
}
func (f *fakeRepo) FindPermissionsByRole(ctx context.Context, roleID string) ([]*db.BunTenantPermission, error) {
	return nil, nil
}
func (f *fakeRepo) ListPermissions(ctx context.Context, branchID string) ([]*db.BunTenantPermission, error) {
	if f.listPerms != nil {
		return f.listPerms(branchID)
	}
	return nil, nil
}
func (f *fakeRepo) CreateRole(ctx context.Context, role *db.BunTenantRole) error {
	f.createdRole = role
	if f.createRole != nil {
		return f.createRole(role)
	}
	return nil
}
func (f *fakeRepo) SetRoleActive(ctx context.Context, id string, isActive bool) (*db.BunTenantRole, error) {
	if f.setRoleActive != nil {
		return f.setRoleActive(id, isActive)
	}
	r := testRole()
	r.IsActive = isActive
	return r, nil
}
func (f *fakeRepo) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	f.replacedPerms = permissionIDs
	if f.replacePerms != nil {
		return f.replacePerms(roleID, permissionIDs)
	}
	return nil
}
func (f *fakeRepo) FindAssignmentsByUser(ctx context.Context, userID string) ([]*model.TenantUserAssignment, error) {
	return nil, nil
}
func (f *fakeRepo) FindAssignmentsByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*model.TenantUserAssignment, error) {
	return nil, nil
}
func (f *fakeRepo) UpsertAssignment(ctx context.Context, userID, branchID, tenantID, roleID string, assignedBy *string) error {
	f.upserted = &upsertCall{userID: userID, branchID: branchID, tenantID: tenantID, roleID: roleID, assignedBy: assignedBy}
	return nil
}
func (f *fakeRepo) FindAssignmentByID(ctx context.Context, id string) (*model.TenantUserAssignment, error) {
	return f.assignment, nil
}
func (f *fakeRepo) FindAssignmentsByBranch(ctx context.Context, branchID string) ([]*model.TenantUserAssignment, error) {
	return []*model.TenantUserAssignment{testAssignment()}, nil
}
func (f *fakeRepo) UpdateAssignmentRole(ctx context.Context, id, roleID string) error {
	f.updatedRole = roleID
	return nil
}
func (f *fakeRepo) SetAssignmentActive(ctx context.Context, id string, isActive bool) error {
	f.activated = &activateCall{id: id, isActive: isActive}
	return nil
}
func (f *fakeRepo) CreateInvite(ctx context.Context, inv *db.BunTenantInvite) error {
	f.created = inv
	return nil
}
func (f *fakeRepo) FindPendingInvite(ctx context.Context, email, branchID string) (*model.TenantInvite, error) {
	return f.pending, nil
}
func (f *fakeRepo) FindInviteByID(ctx context.Context, id string) (*model.TenantInvite, error) {
	return f.pending, nil
}
func (f *fakeRepo) FindInviteByToken(ctx context.Context, token string) (*model.TenantInvite, error) {
	return f.byToken, nil
}
func (f *fakeRepo) InviteToken(ctx context.Context, id string) (string, error) {
	return f.tokens[id], nil
}
func (f *fakeRepo) SetInviteToken(ctx context.Context, id, token string) error {
	f.setTokens = append(f.setTokens, [2]string{id, token})
	if f.tokens == nil {
		f.tokens = map[string]string{}
	}
	f.tokens[id] = token
	return nil
}
func (f *fakeRepo) ListInvitesByBranch(ctx context.Context, branchID string) ([]*model.TenantInvite, error) {
	return nil, nil
}
func (f *fakeRepo) AcceptInvite(ctx context.Context, email, branchID string) error {
	f.accepts = append(f.accepts, [2]string{email, branchID})
	return nil
}
func (f *fakeRepo) RefreshInvite(ctx context.Context, id string) (*model.TenantInvite, error) {
	f.refreshed = append(f.refreshed, id)
	return &model.TenantInvite{ID: id, Status: "pending"}, nil
}
func (f *fakeRepo) FindAddressByID(ctx context.Context, id string) (*db.BunAddress, error) {
	return nil, nil
}
func (f *fakeRepo) CreateAddress(ctx context.Context, addr *db.BunAddress) error {
	return nil
}
func (f *fakeRepo) UpdateAddress(ctx context.Context, id string, addr *db.BunAddress) error {
	return nil
}

func newSvc(f *fakeRepo) (*service.TenantService, *fakeMailer, *fakeAccounts) {
	svc := service.NewTenantService(f, nil)
	m := &fakeMailer{}
	a := &fakeAccounts{creds: &service.UserCredentials{UserID: "u9", Token: "tok", RefreshToken: "rtok", Email: "new@clinic.com"}}
	svc.Mailer = m
	svc.UserAccounts = a
	svc.FrontendURL = "http://front:4200"
	return svc, m, a
}

func TestCreateTenantRole(t *testing.T) {
	f := &fakeRepo{}
	svc, _, _ := newSvc(f)

	role, err := svc.CreateTenantRole(context.Background(), "on-call", "b1", strptr("Night cover"))
	if err != nil {
		t.Fatalf("CreateTenantRole: %v", err)
	}
	if role.Name != "on-call" || role.BranchID == nil || *role.BranchID != "b1" {
		t.Fatalf("role not scoped to branch: %+v", role)
	}
	if role.TenantID == nil || *role.TenantID != "t1" {
		t.Fatalf("tenant not stamped from branch: %+v", role)
	}
	if role.IsSystemRole || !role.IsActive {
		t.Fatalf("new role must be active non-system: %+v", role)
	}
	// Regression: empty ID breaks Postgres UUID columns.
	if _, err := uuid.Parse(role.ID); err != nil {
		t.Fatalf("role ID is not a UUID: %q", role.ID)
	}
	if f.createdRole != role {
		t.Fatal("expected role to be persisted via CreateRole")
	}
}

func TestCreateTenantRoleBranchNotFound(t *testing.T) {
	f := &fakeRepo{findBranch: func(ctx context.Context, id string) (*db.BunBranch, error) {
		return nil, nil
	}}
	svc, _, _ := newSvc(f)
	if _, err := svc.CreateTenantRole(context.Background(), "x", "nope", nil); err == nil {
		t.Fatal("expected error for unknown branch")
	}
}

func TestSetRolePermissionsResolvesResourceActionRefs(t *testing.T) {
	f := &fakeRepo{
		listPerms: func(branchID string) ([]*db.BunTenantPermission, error) {
			if branchID != "b1" {
				t.Fatalf("lookup scoped to role branch, got %q", branchID)
			}
			return []*db.BunTenantPermission{
				{ID: "p-patient-read", Resource: "patient", Action: "read"},
				{ID: "p-patient-write", Resource: "patient", Action: "write"},
			}, nil
		},
	}
	svc, _, _ := newSvc(f)

	if _, err := svc.SetRolePermissions(context.Background(), "r1", []string{"patient:read", "patient:write"}); err != nil {
		t.Fatalf("SetRolePermissions: %v", err)
	}
	if len(f.replacedPerms) != 2 || f.replacedPerms[0] != "p-patient-read" || f.replacedPerms[1] != "p-patient-write" {
		t.Fatalf("refs not resolved to IDs: %v", f.replacedPerms)
	}
}

func TestSetRolePermissionsUnknownRef(t *testing.T) {
	f := &fakeRepo{
		listPerms: func(branchID string) ([]*db.BunTenantPermission, error) {
			return []*db.BunTenantPermission{{ID: "p1", Resource: "patient", Action: "read"}}, nil
		},
	}
	svc, _, _ := newSvc(f)
	if _, err := svc.SetRolePermissions(context.Background(), "r1", []string{"billing:write"}); err == nil {
		t.Fatal("expected error for unknown permission ref")
	}
	if f.replacedPerms != nil {
		t.Fatal("must not persist on resolution failure")
	}
}

func TestSetRolePermissionsAcceptsUUIDs(t *testing.T) {
	f := &fakeRepo{findPerm: func(ctx context.Context, id string) (*db.BunTenantPermission, error) {
		if id != "p-uuid-1" {
			return nil, nil
		}
		return &db.BunTenantPermission{ID: id}, nil
	}}
	svc, _, _ := newSvc(f)
	if _, err := svc.SetRolePermissions(context.Background(), "r1", []string{"p-uuid-1"}); err != nil {
		t.Fatalf("SetRolePermissions: %v", err)
	}
	if len(f.replacedPerms) != 1 || f.replacedPerms[0] != "p-uuid-1" {
		t.Fatalf("UUID passthrough broken: %v", f.replacedPerms)
	}
}

func TestSetRolePermissionsRoleNotFound(t *testing.T) {
	f := &fakeRepo{findRole: func(ctx context.Context, id string) (*db.BunTenantRole, error) {
		return nil, nil
	}}
	svc, _, _ := newSvc(f)
	if _, err := svc.SetRolePermissions(context.Background(), "nope", []string{"patient:read"}); err == nil {
		t.Fatal("expected error for unknown role")
	}
}

func TestInviteCreatesPendingInviteAndSendsLink(t *testing.T) {
	f := &fakeRepo{}
	svc, m, _ := newSvc(f)

	ok, err := svc.InviteUser(context.Background(), "  New@clinic.com ", "b1", "r1")
	if err != nil || !ok {
		t.Fatalf("InviteUser: ok=%v err=%v", ok, err)
	}
	if f.created == nil {
		t.Fatal("expected pending invite to be created")
	}
	if f.created.Email != "new@clinic.com" {
		t.Fatalf("email not normalized: %q", f.created.Email)
	}
	if f.created.Status != "pending" || f.created.BranchID != "b1" || f.created.TenantID != "t1" || f.created.RoleID != "r1" {
		t.Fatalf("invite fields wrong: %+v", f.created)
	}
	if f.created.Token == nil || len(*f.created.Token) < 32 {
		t.Fatal("invite must carry a secure token for the email link")
	}
	if _, err := uuid.Parse(f.created.ID); err != nil {
		t.Fatalf("invite ID is not a UUID: %q", f.created.ID)
	}
	if f.upserted != nil {
		t.Fatal("no assignment should exist before acceptance")
	}
	if len(m.sends) != 1 {
		t.Fatalf("expected one invite email, got %d", len(m.sends))
	}
	if m.sends[0].to != "new@clinic.com" {
		t.Fatalf("email to wrong address: %q", m.sends[0].to)
	}
	if !strings.Contains(m.sends[0].body, *f.created.Token) || !strings.Contains(m.sends[0].body, "/auth/accept-invite?token=") {
		t.Fatalf("email body missing invite link: %q", m.sends[0].body)
	}
}

func TestInviteRefreshesExistingPending(t *testing.T) {
	f := &fakeRepo{
		pending: testInvite(),
		tokens:  map[string]string{"inv1": "tok-abc"},
	}
	svc, m, _ := newSvc(f)

	if _, err := svc.InviteUser(context.Background(), "new@clinic.com", "b1", "r1"); err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	if len(f.refreshed) != 1 || f.refreshed[0] != "inv1" {
		t.Fatalf("expected pending invite refresh: %v", f.refreshed)
	}
	if f.created != nil {
		t.Fatal("must not duplicate pending invites")
	}
	if len(m.sends) != 1 || !strings.Contains(m.sends[0].body, "tok-abc") {
		t.Fatal("resent email must reuse the existing link token")
	}
}

func TestInviteBackfillsMissingToken(t *testing.T) {
	f := &fakeRepo{pending: testInvite()} // tokens map empty: legacy row
	svc, m, _ := newSvc(f)

	if _, err := svc.InviteUser(context.Background(), "new@clinic.com", "b1", "r1"); err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	if len(f.setTokens) != 1 || f.setTokens[0][0] != "inv1" || f.setTokens[0][1] == "" {
		t.Fatalf("expected token backfill: %v", f.setTokens)
	}
	if len(m.sends) != 1 || !strings.Contains(m.sends[0].body, f.setTokens[0][1]) {
		t.Fatal("email must carry the backfilled token")
	}
}

func TestInviteValidation(t *testing.T) {
	svc, _, _ := newSvc(&fakeRepo{})
	if _, err := svc.InviteUser(context.Background(), "   ", "b1", "r1"); err == nil {
		t.Fatal("expected error for blank email")
	}
}

func TestInviteRejectsForeignRole(t *testing.T) {
	other := testRole()
	other.BranchID = strptr("other-branch")
	f := &fakeRepo{findRole: func(ctx context.Context, id string) (*db.BunTenantRole, error) {
		return other, nil
	}}
	svc, _, _ := newSvc(f)
	if _, err := svc.InviteUser(context.Background(), "a@b.com", "b1", "r1"); err == nil {
		t.Fatal("expected error for role from another branch")
	}
}

func TestInviteEmailFailureDoesNotFailInvite(t *testing.T) {
	f := &fakeRepo{}
	svc, m, _ := newSvc(f)
	m.err = fmt.Errorf("smtp down")
	if _, err := svc.InviteUser(context.Background(), "a@b.com", "b1", "r1"); err != nil {
		t.Fatalf("mail failure must not fail the invite: %v", err)
	}
	if f.created == nil {
		t.Fatal("invite must persist even when email fails")
	}
}

func TestAcceptInvite(t *testing.T) {
	f := &fakeRepo{byToken: testInvite()}
	svc, _, a := newSvc(f)

	out, err := svc.AcceptInvite(context.Background(), "tok-abc", "New Doc", "secret123")
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if len(a.calls) != 1 || a.calls[0] != [3]string{"tok-abc", "New Doc", "secret123"} {
		t.Fatalf("expected single user-service accept call: %v", a.calls)
	}
	if f.upserted == nil || f.upserted.userID != "u9" || f.upserted.branchID != "b1" || f.upserted.tenantID != "t1" || f.upserted.roleID != "r1" {
		t.Fatalf("assignment not created: %+v", f.upserted)
	}
	if len(f.accepts) != 1 || f.accepts[0] != [2]string{"new@clinic.com", "b1"} {
		t.Fatalf("invite not accepted: %v", f.accepts)
	}
	if out.Token != "tok" || out.RefreshToken != "rtok" || out.UserID != "u9" || out.Email != "new@clinic.com" {
		t.Fatalf("payload wrong: %+v", out)
	}
}

func TestAcceptInviteValidation(t *testing.T) {
	f := &fakeRepo{byToken: testInvite()}
	svc, _, _ := newSvc(f)

	if _, err := svc.AcceptInvite(context.Background(), "tok", "  ", "secret123"); err == nil {
		t.Fatal("expected error for blank name")
	}
	if _, err := svc.AcceptInvite(context.Background(), "tok", "Doc", "123"); err == nil {
		t.Fatal("expected error for short password")
	}
	f.byToken = nil
	if _, err := svc.AcceptInvite(context.Background(), "bad", "Doc", "secret123"); err == nil {
		t.Fatal("expected error for unknown token")
	}
}

func TestAcceptInviteAccountError(t *testing.T) {
	f := &fakeRepo{byToken: testInvite()}
	svc, _, a := newSvc(f)
	a.err = fmt.Errorf("email already registered with a different password")
	if _, err := svc.AcceptInvite(context.Background(), "tok", "Doc", "secret123"); err == nil {
		t.Fatal("expected account error to propagate")
	}
	if f.upserted != nil {
		t.Fatal("must not assign without an account")
	}
}

func TestResendInvite(t *testing.T) {
	f := &fakeRepo{pending: testInvite(), tokens: map[string]string{"inv1": "tok-abc"}}
	svc, m, _ := newSvc(f)

	inv, err := svc.ResendInvite(context.Background(), "inv1")
	if err != nil || inv.ID != "inv1" {
		t.Fatalf("ResendInvite: %+v %v", inv, err)
	}
	if len(m.sends) != 1 || !strings.Contains(m.sends[0].body, "tok-abc") {
		t.Fatal("resend must re-email the invite link")
	}

	f.pending = &model.TenantInvite{ID: "inv1", Status: "accepted"}
	if _, err := svc.ResendInvite(context.Background(), "inv1"); err == nil {
		t.Fatal("expected error resending accepted invite")
	}

	f.pending = nil
	if _, err := svc.ResendInvite(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for unknown invite")
	}
}

func TestUpdateAssignment(t *testing.T) {
	f := &fakeRepo{assignment: testAssignment()}
	svc, _, _ := newSvc(f)
	updated, err := svc.UpdateAssignment(context.Background(), "a1", "r1")
	if err != nil || updated == nil {
		t.Fatalf("UpdateAssignment: %+v %v", updated, err)
	}
	if f.updatedRole != "r1" {
		t.Fatalf("role not updated: %q", f.updatedRole)
	}
}

func TestUpdateAssignmentGuards(t *testing.T) {
	f := &fakeRepo{assignment: nil}
	svc, _, _ := newSvc(f)
	if _, err := svc.UpdateAssignment(context.Background(), "missing", "r1"); err == nil {
		t.Fatal("expected error for unknown assignment")
	}

	other := testRole()
	other.BranchID = strptr("other-branch")
	f = &fakeRepo{
		assignment: testAssignment(),
		findRole: func(ctx context.Context, id string) (*db.BunTenantRole, error) {
			return other, nil
		},
	}
	svc, _, _ = newSvc(f)
	if _, err := svc.UpdateAssignment(context.Background(), "a1", "r1"); err == nil {
		t.Fatal("expected error for foreign-branch role")
	}
}

func TestSetAssignmentActive(t *testing.T) {
	f := &fakeRepo{assignment: testAssignment()}
	svc, _, _ := newSvc(f)
	updated, err := svc.SetAssignmentActive(context.Background(), "a1", false)
	if err != nil || updated == nil {
		t.Fatalf("SetAssignmentActive: %+v %v", updated, err)
	}
	if f.activated == nil || f.activated.id != "a1" || f.activated.isActive {
		t.Fatalf("activation not persisted: %+v", f.activated)
	}
}

func TestListPassthroughs(t *testing.T) {
	svc, _, _ := newSvc(&fakeRepo{assignment: testAssignment()})
	if _, err := svc.GetRolesByBranch(context.Background(), "b1"); err != nil {
		t.Fatalf("GetRolesByBranch: %v", err)
	}
	if _, err := svc.ListPermissions(context.Background(), "b1"); err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if _, err := svc.ListInvitesByBranch(context.Background(), "b1"); err != nil {
		t.Fatalf("ListInvitesByBranch: %v", err)
	}
	if _, err := svc.ListAssignmentsByBranch(context.Background(), "b1"); err != nil {
		t.Fatalf("ListAssignmentsByBranch: %v", err)
	}
	if _, err := svc.InviteByToken(context.Background(), "tok"); err != nil {
		t.Fatalf("InviteByToken: %v", err)
	}
}
