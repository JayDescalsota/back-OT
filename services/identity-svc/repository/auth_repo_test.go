package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ot/identity-svc/repository"
)

func TestWithClaimsAndGetClaims(t *testing.T) {
	claims := &repository.UserClaims{
		UserID: uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	ctx := repository.WithClaims(context.Background(), claims)
	got := repository.GetClaims(ctx)

	if got == nil {
		t.Fatal("expected claims, got nil")
	}
	if got.UserID != claims.UserID {
		t.Errorf("expected UserID %s, got %s", claims.UserID, got.UserID)
	}
}

func TestGetClaims_NoClaimsInContext(t *testing.T) {
	got := repository.GetClaims(context.Background())
	if got != nil {
		t.Fatal("expected nil claims from empty context")
	}
}
