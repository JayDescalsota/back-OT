# Backend Coding Standards (Go + gqlgen + Bun)

Every microservice follows a strict **Resolver → Service → Repository** three-layer separation. Each layer has one responsibility and depends only on the layer below it.

```
┌──────────────────────────┐
│      graph/resolver.go   │  ← Thinnest possible — auth check, call service, return
├──────────────────────────┤
│      service/             │  ← Business logic, validation, orchestration
├──────────────────────────┤
│      repository/          │  ← DB queries only (Bun), struct-injected *bun.DB
└──────────────────────────┘
```

**Rules:**
- Resolver NEVER calls repository directly.
- Service NEVER sees `*bun.DB` — only repository interfaces.
- Repository NEVER contains business logic, only SQL/queries.

---

## 1. Directory Layout

```
back/
├── shared/                          # Shared Go module
│   ├── auth/auth.go                 # JWT, bcrypt, token utilities
│   ├── errors/errors.go             # AppError, NotFound, Validation, Unauthorized
│   ├── logger/logger.go             # slog wrapper
│   ├── db/pool.go                   # Bun DB connection helper
│   └── middleware/
│       ├── authentication.go        # Unified JWT auth middleware
│       └── context.go               # Tenant/branch context types
│
├── services/
│   ├── auth-svc/                    # Authentication service
│   │   ├── main.go                  # Entry point — wires dependencies
│   │   ├── graph/                   # gqlgen managed
│   │   │   ├── schema.graphqls      # Type definitions (you write)
│   │   │   ├── generated/           # AUTO-GENERATED
│   │   │   │   ├── generated.go
│   │   │   │   └── federation.go
│   │   │   ├── model/models_gen.go  # AUTO-GENERATED
│   │   │   ├── resolver.go          # Resolver struct (you write — thin)
│   │   │   └── schema.resolvers.go  # Generated stubs → you fill in
│   │   ├── service/
│   │   │   ├── auth_service.go      # Register, login, token logic
│   │   │   └── service_test.go
│   │   ├── repository/
│   │   │   ├── user_repo.go         # User CRUD for auth
│   │   │   ├── auth_repo.go         # JWT claims/token helpers
│   │   │   └── user_repo_test.go
│   │   ├── middleware/
│   │   │   └── context.go           # GraphQL auth middleware
│   │   ├── migrations/
│   │   ├── gqlgen.yml
│   │   └── Makefile
│   │
│   ├── user-svc/                    # User management service
│   │   ├── main.go
│   │   ├── graph/
│   │   │   ├── schema.graphqls
│   │   │   ├── generated/
│   │   │   ├── model/models_gen.go
│   │   │   ├── resolver.go
│   │   │   └── schema.resolvers.go
│   │   ├── service/
│   │   │   ├── user_service.go      # User CRUD, assignments, roles
│   │   │   └── service_test.go
│   │   ├── repository/
│   │   │   ├── user_repo.go
│   │   │   └── user_repo_test.go
│   │   ├── middleware/
│   │   │   └── context.go           # JWT auth → userID context
│   │   ├── migrations/
│   │   ├── gqlgen.yml
│   │   └── Makefile
│   │
│   ├── patient-svc/
│   ├── scheduling-svc/
│   ├── clinical-svc/
│   ├── billing-svc/
│   ├── referral-svc/
│   ├── hr-svc/
│   ├── messaging-svc/
│   ├── notification-svc/
│   └── analytics-svc/
│
└── gateway/
```

---

## 2. Three-Layer Rules

### Resolver (`graph/schema.resolvers.go`)

- 2-4 lines per method.
- Extracts claims/auth from context.
- Calls ONE service method.
- NEVER does validation, NEVER calls repository.

```go
// ✅ Correct
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
    return r.AuthService.Login(ctx, input.Email, input.Password)
}
```

### Service (`service/*.go`)

- All business logic, validation, and orchestration.
- Receives repository interface via constructor injection.
- NEVER imports `*bun.DB` or any database driver.
- Returns domain models or error.

```go
// ✅ Correct
type AuthService struct {
    userRepo  UserRepository   // ← interface, not concrete type
    jwtSecret string
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthPayload, error) {
    // validation
    // logic
    // call repository
}
```

### Repository (`repository/*.go`)

- Struct with injected `*bun.DB` — never passed as parameter.
- One file per aggregate root (not per table).
- Methods: `FindByID`, `FindByEmail`, `Create`, `Update`, `Delete`.
- Bun models stay inside repository — never exported.
- Always check for `sql.ErrNoRows` to distinguish "not found" from DB errors.

```go
// ✅ Correct
func (r *UserRepo) FindByID(ctx context.Context, id string) (*BunUser, error) {
    user := new(BunUser)
    err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return user, nil
}
```

---

## 3. Injection Chain

```
main.go
  │
  ├── db := sharedDB.NewDB()
  │
  ├── userRepo  := repository.NewUserRepo(db)
  │
  ├── authSvc   := service.NewAuthService(userRepo, jwtSecret)
  ├── userSvc   := service.NewUserService(userRepo, currentUserFn)
  │
  └── graph.Resolver{ AuthService: authSvc, UserService: userSvc }
         │
         └── Injected into gqlgen Config → Resolvers
```

Dependencies flow down: `main.go` wires everything, injects `*bun.DB` only into repositories, repositories into services, services into resolver.

---

## 4. Service-to-Service Communication

Services share the same database in this monorepo architecture. Ownership:
- **auth-svc** owns `users` table (credentials, auth operations)
- **user-svc** reads `users` table, owns `roles`, `permissions`, `user_branch_assignments`

JWT tokens contain only `userId`. Tenant/branch context is resolved via `myAssignments` query and passed as HTTP headers (`x-tenant-id`, `x-branch-id`).

---

## 5. Repository Pattern

### 5a. Repository struct

```go
// repository/user_repo.go
package repository

type BunUser struct {
    bun.BaseModel `bun:"table:users"`
    ID            string     `bun:"id,pk"`
    Email         string     `bun:"email,notnull,unique"`
    PasswordHash  string     `bun:"password_hash,notnull"`
    IsActive      bool       `bun:"is_active,default:true"`
    CreatedAt     time.Time  `bun:"created_at"`
    UpdatedAt     time.Time  `bun:"updated_at"`
}

type UserRepo struct {
    db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
    return &UserRepo{db: db}
}
```

### 5b. Error handling — check sql.ErrNoRows

```go
func (r *UserRepo) FindByID(ctx context.Context, id string) (*BunUser, error) {
    user := new(BunUser)
    err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return user, nil
}
```

### 5c. Repository test (sqlmock)

```go
// repository/user_repo_test.go
package repository_test

func newMockRepo(t *testing.T) (*repository.UserRepo, sqlmock.Sqlmock) {
    t.Helper()
    db, mock, err := sqlmock.New()
    repo := repository.NewUserRepo(bunDB)
    return repo, mock
}
```

---

## 6. Repository Interface

Services depend on an interface, not the concrete `*bun.DB`-holding struct:

```go
// service/auth_service.go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*repository.BunUser, error)
    FindByEmail(ctx context.Context, email string) (*repository.BunUser, error)
    Create(ctx context.Context, email, password, name string) (*repository.BunUser, error)
    UpdateLastLogin(ctx context.Context, userID string) error
}
```

---

## 7. Current User Resolution

User-svc resolves the current user via a `CurrentUserFn` function injected at startup:

```go
// service/user_service.go
type CurrentUserFn func(ctx context.Context) string

type UserService struct {
    userRepo    UserRepository
    currentUser CurrentUserFn
}
```

The middleware extracts `userId` from the JWT and puts it in context. `CurrentUserFn` reads it back. This avoids importing middleware packages into the service layer.

---

## 8. Testing

**Test each layer independently:**

| Layer | How | What to cover |
|---|---|---|
| **Repository** | sqlmock (mock `*sql.DB`) | Every query path: found, not found, error, edge cases |
| **Service** | Mock repository (hand-written interface or structs) | Validation, error mapping, happy path |
| **Resolver** | NOT tested (thin wrappers, 2-4 lines) | Covered by integration tests later |

**Coverage target:** 90%+ for `repository/` package only (enforced in CI).

```makefile
test:
    go test ./... -count=1

cover:
    go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
    go tool cover -func=coverage.out
```

---

## 9. Starting a New Service

1. Copy an existing service (e.g. `auth-svc`) — folder structure is identical.
2. Update `go.mod` if needed.
3. Write `graph/schema.graphqls` — types, queries, mutations.
4. Write `repository/<entity>_repo.go` — Bun struct methods.
5. Write `repository/<entity>_repo_test.go` — sqlmock tests (90%+ coverage).
6. Define repository interface in `service/<entity>_service.go`.
7. Write `service/<entity>_service.go` — business logic calling repository via interface.
8. Write `service/<entity>_service_test.go` — mock repo, test all validation paths.
9. Run `make generate` — gqlgen creates stubs.
10. Fill in `graph/schema.resolvers.go` — thin wrappers calling services.
11. Update `main.go` — wire the dependency chain.
12. Run `make test` — verify all tests pass.

---

## 10. Summary Checklist

- [ ] Three layers: Resolver → Service → Repository
- [ ] Resolver is ≤4 lines per method — calls service only
- [ ] Service has all validation and business logic
- [ ] Repository uses struct methods — not free functions
- [ ] `*bun.DB` injected into repository struct, never passed as parameter
- [ ] Repository checks `sql.ErrNoRows` for not-found cases
- [ ] Resolver never imports `*bun.DB` or `repository/` directly
- [ ] Service never imports `*bun.DB` — depends on repository interface
- [ ] Service tested with mock repository (all validation paths)
- [ ] Repository tested with sqlmock (every query path)
- [ ] Repository coverage 90%+
- [ ] `main.go` wires the full chain: DB → repo → service → resolver
- [ ] JWT auth via `shared/middleware/authentication.go` or service-specific middleware
- [ ] User context resolved via `CurrentUserFn`, not direct middleware import
