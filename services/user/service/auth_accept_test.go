package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clinicmanager/services/user/models"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/services/user/service"
)

type stubMailer struct{}

func (stubMailer) Send(to, subject, body string) error { return nil }

// stubTenant serves tenant inviteByToken lookups: nil email => unknown token.
func stubTenant(t *testing.T, email string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if email == "" {
			_, _ = w.Write([]byte(`{"data":{"inviteByToken":null}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"inviteByToken":{"email":"` + email + `" }}}`))
	}))
}

func acceptSvc(repo *mockRepo, tenantURL string) *service.AuthService {
	svc := service.NewAuthService(repo, "test-secret", stubMailer{}, "http://front", 15*time.Minute, 7*24*time.Hour)
	svc.TenantSvcURL = tenantURL
	return svc
}

func TestAcceptInviteNewUser(t *testing.T) {
	repo := defaultMock()
	repo.registerFn = func(_ context.Context, email, password, _ string) (*models.User, error) {
		u := userStub("u-new", email)
		return u, nil
	}
	srv := stubTenant(t, "new@clinic.com")
	defer srv.Close()

	out, err := acceptSvc(repo, srv.URL).AcceptInvite(context.Background(), "tok", "New Doc", "secret123")
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if out.Token == "" || out.RefreshToken == "" || out.User == nil || out.User.ID != "u-new" {
		t.Fatalf("incomplete session payload: %+v", out)
	}
	if len(repo.validatedIDs) != 1 || repo.validatedIDs[0] != "u-new" {
		t.Fatalf("account must be validated on accept: %v", repo.validatedIDs)
	}
	if repo.upsertedProfile == nil || repo.upsertedProfile.UserID != "u-new" || repo.upsertedProfile.FirstName != "New Doc" {
		t.Fatalf("profile name not set: %+v", repo.upsertedProfile)
	}
}

func TestAcceptInviteExistingUser(t *testing.T) {
	hash, err := repository.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	existing := userStub("u-old", "doc@clinic.com")
	existing.PasswordHash = hash
	repo := defaultMock()
	repo.findByEmailFn = func(_ context.Context, _ string) (*models.User, error) { return existing, nil }
	srv := stubTenant(t, "doc@clinic.com")
	defer srv.Close()

	out, err := acceptSvc(repo, srv.URL).AcceptInvite(context.Background(), "tok", "Doc", "secret123")
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if out.User == nil || out.User.ID != "u-old" {
		t.Fatalf("expected existing account session: %+v", out)
	}
}

func TestAcceptInviteWrongPassword(t *testing.T) {
	hash, _ := repository.HashPassword("correct-horse")
	existing := userStub("u-old", "doc@clinic.com")
	existing.PasswordHash = hash
	repo := defaultMock()
	repo.findByEmailFn = func(_ context.Context, _ string) (*models.User, error) { return existing, nil }
	srv := stubTenant(t, "doc@clinic.com")
	defer srv.Close()

	if _, err := acceptSvc(repo, srv.URL).AcceptInvite(context.Background(), "tok", "Doc", "wrongpass1"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestAcceptInviteBadToken(t *testing.T) {
	repo := defaultMock()
	srv := stubTenant(t, "")
	defer srv.Close()

	if _, err := acceptSvc(repo, srv.URL).AcceptInvite(context.Background(), "bad", "Doc", "secret123"); err == nil {
		t.Fatal("expected error for unknown token")
	}
	if repo.upsertedProfile != nil || len(repo.validatedIDs) != 0 {
		t.Fatal("no writes allowed for invalid tokens")
	}
}

func TestAcceptInviteValidation(t *testing.T) {
	repo := defaultMock()
	srv := stubTenant(t, "doc@clinic.com")
	defer srv.Close()
	svc := acceptSvc(repo, srv.URL)

	if _, err := svc.AcceptInvite(context.Background(), "tok", "  ", "secret123"); err == nil {
		t.Fatal("expected error for blank name")
	}
	if _, err := svc.AcceptInvite(context.Background(), "tok", "Doc", "short1"); err == nil {
		t.Fatal("expected error for short password")
	}
}
