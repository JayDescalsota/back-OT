package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
	branchIDKey contextKey = "branch_id"
	userIDKey   contextKey = "user_id"
)

type TenantCtx struct {
	TenantID string
	BranchID string
	UserID   string
}

func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("x-tenant-id")
		branchID := r.Header.Get("x-branch-id")
		userID := r.Header.Get("x-user-id")

		if tenantID == "" || branchID == "" || userID == "" {
			http.Error(w, "Missing required headers", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, branchIDKey, branchID)
		ctx = context.WithValue(ctx, userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) TenantCtx {
	tenantID, _ := ctx.Value(tenantIDKey).(string)
	branchID, _ := ctx.Value(branchIDKey).(string)
	userID, _ := ctx.Value(userIDKey).(string)
	return TenantCtx{TenantID: tenantID, BranchID: branchID, UserID: userID}
}

func UserIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
