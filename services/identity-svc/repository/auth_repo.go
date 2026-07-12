package repository

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "jwt_claims"

type UserClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func WithClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(claimsKey).(*UserClaims)
	return claims
}

func GenerateToken(userID, secret string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
