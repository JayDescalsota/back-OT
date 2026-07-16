package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/clinicmanager/shared/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(r.Context(), "panic recovered", "error", rec, "stack", string(debug.Stack()))
				http.Error(w, `{"errors":[{"message":"internal server error"}]}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
