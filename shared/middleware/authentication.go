package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	c "github.com/clinicmanager/shared/context"
	jwt "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtKey string, isRevoked func(ctx context.Context, sessionID string) (bool, error), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"errors":[{"message":"unauthorized"}]}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, `{"errors":[{"message":"unauthorized"}]}`, http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"errors":[{"message":"unauthorized"}]}`, http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		if userID, ok := claims["userId"].(string); ok {
			ctx = c.SetUserID(ctx, userID)
		}
		if role, ok := claims["role"].(string); ok {
			ctx = c.SetUserRole(ctx, role)
		}

		if isRevoked != nil {
			if sessionID, ok := claims["sessionId"].(string); ok && sessionID != "" {
				revoked, err := isRevoked(ctx, sessionID)
				if err != nil || revoked {
					http.Error(w, `{"errors":[{"message":"unauthorized"}]}`, http.StatusUnauthorized)
					return
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
