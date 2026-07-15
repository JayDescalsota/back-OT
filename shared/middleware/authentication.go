package middleware

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
				ctx = SetUserID(ctx, userID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
