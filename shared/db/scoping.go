package db

import (
	"context"
	"fmt"
	"reflect"

	sharedctx "github.com/clinicmanager/shared/context"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// resolveTable gets the schema.Table metadata for the given model.
func resolveTable(db bun.IDB, model interface{}) *schema.Table {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return db.Dialect().Tables().Get(t)
}

// TenantQuery returns a select query pre-filtered by the tenant_id in the context.
// Accepts bun.IDB so it can be used with both *bun.DB and *bun.Tx.
// If the table does not support tenant scoping (missing tenant_id column), it returns a query in error state.
func TenantQuery(ctx context.Context, db bun.IDB, model interface{}) *bun.SelectQuery {
	query := db.NewSelect().Model(model)

	table := resolveTable(db, model)
	hasTenantID := false
	for _, f := range table.Fields {
		if f.SQLName == "tenant_id" {
			hasTenantID = true
			break
		}
	}

	if !hasTenantID {
		return query.Err(fmt.Errorf("table %q cannot be tenant scoped: missing tenant_id column", table.Name))
	}

	tenantID := sharedctx.TenantIDFromCtx(ctx)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query
}

// BranchQuery returns a select query pre-filtered by both tenant_id and branch_id in the context.
// Accepts bun.IDB so it can be used with both *bun.DB and *bun.Tx.
// If the table does not support branch scoping (missing tenant_id or branch_id column), it returns a query in error state.
func BranchQuery(ctx context.Context, db bun.IDB, model interface{}) *bun.SelectQuery {
	query := db.NewSelect().Model(model)

	table := resolveTable(db, model)
	hasTenantID := false
	hasBranchID := false
	for _, f := range table.Fields {
		if f.SQLName == "tenant_id" {
			hasTenantID = true
		}
		if f.SQLName == "branch_id" {
			hasBranchID = true
		}
	}

	if !hasTenantID || !hasBranchID {
		var missing []string
		if !hasTenantID {
			missing = append(missing, "tenant_id")
		}
		if !hasBranchID {
			missing = append(missing, "branch_id")
		}
		return query.Err(fmt.Errorf("table %q cannot be branch scoped: missing %v column(s)", table.Name, missing))
	}

	tctx := sharedctx.FromContext(ctx)
	if tctx.TenantID != "" {
		query = query.Where("tenant_id = ?", tctx.TenantID)
	}
	if tctx.BranchID != "" {
		query = query.Where("branch_id = ?", tctx.BranchID)
	}
	return query
}
