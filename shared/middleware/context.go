package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
	branchIDKey contextKey = "branch_id"
)

func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("x-tenant-id")
		branchID := r.Header.Get("x-branch-id")
		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, branchIDKey, branchID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TenantCtx struct {
	TenantID string
	BranchID string
}

func FromContext(ctx context.Context) TenantCtx {
	return TenantCtx{
		TenantID: ctx.Value(tenantIDKey).(string),
		BranchID: ctx.Value(branchIDKey).(string),
	}
}
