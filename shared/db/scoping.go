package db

import (
	"context"
	"reflect"
	"strings"

	sharedctx "github.com/clinicmanager/shared/context"
	"github.com/uptrace/bun"
)

// Scope controls which context values are applied as WHERE filters automatically.
type Scope int

const (
	// ScopeBranch filters by both tenant_id AND branch_id (default — most restrictive).
	ScopeBranch Scope = iota
	// ScopeTenant filters by tenant_id only (cross-branch access).
	ScopeTenant
	// ScopeAll applies no automatic filters (unscoped/admin access).
	ScopeAll
)

// ScopedDB wraps a *bun.DB and automatically injects tenant_id and branch_id
// WHERE clauses based on context and scope configuration.
type ScopedDB struct {
	raw   *bun.DB
	scope Scope
}

// DBSet holds the 3 standard database scoping levels for a service.
type DBSet struct {
	// DB: tenant_id + branch_id scoped (default — use for almost all queries).
	DB *ScopedDB
	// TenantDB: tenant_id scoped only (use for cross-branch queries within a tenant).
	TenantDB *ScopedDB
	// AllDB: unscoped (use ONLY for global admin or identity queries).
	AllDB *ScopedDB
}

// NewDBSet creates a DBSet with all 3 scoping levels initialized.
func NewDBSet(raw *bun.DB) *DBSet {
	return &DBSet{
		DB:       &ScopedDB{raw: raw, scope: ScopeBranch},
		TenantDB: &ScopedDB{raw: raw, scope: ScopeTenant},
		AllDB:    &ScopedDB{raw: raw, scope: ScopeAll},
	}
}

// hasColumn checks whether a bun model's table has a given SQL column name.
// Returns false (not an error) if the table is not registered or the column is absent.
func hasColumn(raw *bun.DB, model interface{}, colName string) bool {
	if model == nil {
		return false
	}
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	table := raw.Table(t)
	if table == nil {
		return false
	}
	for _, f := range table.Fields {
		sqlName := strings.Trim(string(f.SQLName), `"`)
		if sqlName == colName {
			return true
		}
	}
	return false
}



// applySelectScope adds the appropriate WHERE clauses to a SelectQuery.
// Silently skips any column that doesn't exist on the model's table.
func (s *ScopedDB) applySelectScope(ctx context.Context, q *bun.SelectQuery, model interface{}) *bun.SelectQuery {
	if s.scope == ScopeAll {
		return q
	}
	tctx := sharedctx.FromContext(ctx)
	if tctx.TenantID != "" && hasColumn(s.raw, model, "tenant_id") {
		q = q.Where("tenant_id = ?", tctx.TenantID)
	}
	if s.scope == ScopeBranch && tctx.BranchID != "" && hasColumn(s.raw, model, "branch_id") {
		q = q.Where("branch_id = ?", tctx.BranchID)
	}
	return q
}

// NewSelect returns a *bun.SelectQuery pre-filtered by scope using values from ctx.
// Pass your model pointer so column introspection can determine which filters apply.
func (s *ScopedDB) NewSelect(ctx context.Context, model interface{}) *bun.SelectQuery {
	q := s.raw.NewSelect().Model(model)
	return s.applySelectScope(ctx, q, model)
}

// NewUpdate returns a *bun.UpdateQuery pre-filtered by scope using values from ctx.
// Silently skips columns absent from the model's table.
func (s *ScopedDB) NewUpdate(ctx context.Context, model interface{}) *bun.UpdateQuery {
	q := s.raw.NewUpdate().Model(model)
	if s.scope == ScopeAll {
		return q
	}
	tctx := sharedctx.FromContext(ctx)
	if tctx.TenantID != "" && hasColumn(s.raw, model, "tenant_id") {
		q = q.Where("tenant_id = ?", tctx.TenantID)
	}
	if s.scope == ScopeBranch && tctx.BranchID != "" && hasColumn(s.raw, model, "branch_id") {
		q = q.Where("branch_id = ?", tctx.BranchID)
	}
	return q
}

// NewDelete returns a *bun.DeleteQuery pre-filtered by scope using values from ctx.
// Silently skips columns absent from the model's table.
func (s *ScopedDB) NewDelete(ctx context.Context, model interface{}) *bun.DeleteQuery {
	q := s.raw.NewDelete().Model(model)
	if s.scope == ScopeAll {
		return q
	}
	tctx := sharedctx.FromContext(ctx)
	if tctx.TenantID != "" && hasColumn(s.raw, model, "tenant_id") {
		q = q.Where("tenant_id = ?", tctx.TenantID)
	}
	if s.scope == ScopeBranch && tctx.BranchID != "" && hasColumn(s.raw, model, "branch_id") {
		q = q.Where("branch_id = ?", tctx.BranchID)
	}
	return q
}

// NewInsert returns a plain *bun.InsertQuery (inserts are not WHERE-scoped;
// the model's tenant_id/branch_id fields must be set on the struct before inserting).
func (s *ScopedDB) NewInsert(model interface{}) *bun.InsertQuery {
	return s.raw.NewInsert().Model(model)
}

// Raw returns the underlying *bun.DB for operations that require it directly
// (e.g. transactions). Use sparingly — prefer the scoped methods above.
func (s *ScopedDB) Raw() *bun.DB {
	return s.raw
}
