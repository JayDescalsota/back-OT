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
│   ├── errors/errors.go             # AppError, NotFound, Validation, Unauthorized
│   ├── logger/logger.go             # slog wrapper
│   ├── db/pool.go                   # Bun DB connection helper
│   └── middleware/context.go        # Tenant/branch context types
│
├── services/
│   ├── identity-svc/                #
│   │   ├── main.go                  # Entry point — wires dependencies
│   │   ├── graph/                   # gqlgen managed
│   │   │   ├── schema.graphqls      # Type definitions (you write)
│   │   │   ├── generated.go         # AUTO-GENERATED
│   │   │   ├── models_gen.go        # AUTO-GENERATED
│   │   │   ├── resolver.go          # Resolver struct (you write — thin)
│   │   │   └── schema.resolvers.go  # Generated stubs → you fill in (2-4 lines each)
│   │   ├── service/                 # ★ Business logic layer
│   │   │   ├── auth_service.go      # Login, register, token logic
│   │   │   └── user_service.go      # User CRUD, assignment queries
│   │   ├── repository/              # ★ Data access layer (struct methods)
│   │   │   ├── user_repo.go         # Bun queries as struct methods
│   │   │   └── user_repo_test.go    # sqlmock tests
│   │   ├── middleware/              # HTTP middleware
│   │   │   └── context.go
│   │   ├── go.mod / go.sum
│   │   ├── gqlgen.yml
│   │   └── Makefile
│   │
│   ├── tenant-svc/ ...              # Same three-layer structure
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

```go
// ❌ Wrong — validation and token generation in resolver
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
    if input.Email == "" { return nil, errors.Validation("...") }
    user, _ := repository.FindUserByEmail(ctx, r.DB, input.Email)
    token, _ := repository.GenerateToken(user.ID, r.JWTSecret)
    return &model.AuthPayload{Token: token, User: user}, nil
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

```go
// ✅ Correct
type UserRepo struct {
    db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*BunUser, error) {
    // r.db used directly
}
```

```go
// ❌ Wrong — db passed as parameter
func FindUserByID(ctx context.Context, db *bun.DB, id string) (*BunUser, error) { ... }
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
  ├── userSvc   := service.NewUserService(userRepo)
  │
  └── graph.Resolver{ AuthService: authSvc, UserService: userSvc }
         │
         └── Injected into gqlgen Config → Resolvers
```

Dependencies flow down: `main.go` wires everything, injects `*bun.DB` only into repositories, repositories into services, services into resolver.

---

## 4. Resolver Root (`graph/resolver.go`)

```go
package graph

import "github.com/ot/identity-svc/service"

type Resolver struct {
    AuthService *service.AuthService
    UserService *service.UserService
}
```

No `*bun.DB` here. The resolver only knows about services.

---

## 5. Repository Pattern

### 5a. Repository struct

```go
// repository/user_repo.go
package repository

import (
    "context"
    "time"
    "github.com/uptrace/bun"
)

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

func (r *UserRepo) FindByID(ctx context.Context, id string) (*BunUser, error) {
    user := new(BunUser)
    err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
    if err != nil {
        return nil, nil
    }
    return user, nil
}
```

### 5b. Repository test (sqlmock)

```go
// repository/user_repo_test.go
package repository_test

func newMockRepo(t *testing.T) (*repository.UserRepo, sqlmock.Sqlmock) {
    t.Helper()
    db, mock, err := sqlmock.New()
    // ...
    repo := repository.NewUserRepo(bunDB)
    return repo, mock
}

func TestUserRepo_FindByID_Found(t *testing.T) {
    repo, mock := newMockRepo(t)
    mock.ExpectQuery(`SELECT .+ FROM "users" .+ WHERE .+`).
        WillReturnRows(...)
    user, err := repo.FindByID(context.Background(), id)
    // assert
}
```

---

## 6. Repository Interface

Services depend on an interface, not the concrete `*bun.DB`-holding struct, so they can be tested with a mock:

```go
// service/auth_service.go
package service

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*repository.BunUser, error)
    FindByEmail(ctx context.Context, email string) (*repository.BunUser, error)
    Create(ctx context.Context, email, password, name string) (*repository.BunUser, error)
    FindAssignmentsByUser(ctx context.Context, userID string) ([]*repository.BunUserBranchAssignment, error)
    UpdateLastLogin(ctx context.Context, userID string) error
}
```

The concrete `*repository.UserRepo` satisfies this interface implicitly.

## 7. Service Pattern

```go
// service/auth_service.go
package service

import (
    "context"
    "strings"
    "github.com/ot/identity-svc/graph/model"
    "github.com/ot/identity-svc/repository"
    "github.com/ot/shared/errors"
)

type AuthService struct {
    userRepo  UserRepository   // ← interface, not concrete *repository.UserRepo
    jwtSecret string
}

func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
    return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthPayload, error) {
    email = strings.TrimSpace(strings.ToLower(email))
    if email == "" {
        return nil, errors.Validation("Email is required")
    }

    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return nil, errors.Unauthorized("Invalid email or password")
    }
    if !user.IsActive {
        return nil, errors.Forbidden("Account is deactivated")
    }
    if !verifyPassword(password, user.PasswordHash) {
        return nil, errors.Unauthorized("Invalid email or password")
    }

    token, err := generateToken(user.ID, s.jwtSecret)
    if err != nil {
        return nil, err
    }

    return &model.AuthPayload{Token: token, User: toUserModel(user)}, nil
}
```

---

## 8. Remaining Files (unchanged patterns)

- `graph/schema.graphqls` — same as before (gqlgen schema)
- `middleware/context.go` — HTTP middleware for tenant/branch/JWT extraction
- `main.go` — wires dependencies, starts HTTP server
- `gqlgen.yml` — gqlgen configuration
- `Makefile` — build/test/generate/cover targets

---

## 9. Testing

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

test-service:
    go test ./service/... -count=1

cover:
    go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
    go tool cover -func=coverage.out
```

---

## 10. Starting a New Service

1. Copy an existing service (e.g. `identity-svc`) — folder structure is identical.
2. Update `go.mod` — module name to `github.com/ot/<your-svc>`.
3. Write `graph/schema.graphqls` — types, queries, mutations.
4. Write `repository/<entity>_repo.go` — Bun struct methods.
5. Write `repository/<entity>_repo_test.go` — sqlmock tests (90%+ coverage).
6. Define `UserRepository` interface in `service/<entity>_service.go` (or alongside).
7. Write `service/<entity>_service.go` — business logic calling repository via interface.
8. Write `service/<entity>_service_test.go` — mock repo, test all validation paths.
9. Run `make generate` — gqlgen creates stubs.
10. Fill in `graph/schema.resolvers.go` — thin wrappers calling services.
11. Update `main.go` — wire the dependency chain.
12. Run `make test` — verify all tests pass.

---

## 11. Summary Checklist

- [ ] Three layers: Resolver → Service → Repository
- [ ] Resolver is ≤4 lines per method — calls service only
- [ ] Service has all validation and business logic
- [ ] Repository uses struct methods (`repo.FindByID(ctx, id)`) — not free functions
- [ ] `*bun.DB` injected into repository struct, never passed as parameter
- [ ] Resolver never imports `*bun.DB` or `repository/` directly
- [ ] Service never imports `*bun.DB` — depends on repository interface, not concrete type
- [ ] Service tested with mock repository (all validation paths)
- [ ] Repository converts Bun models → gqlgen models at boundary
- [ ] `main.go` wires the full chain: DB → repo → service → resolver
- [ ] Every repository function tested with sqlmock
- [ ] Repository coverage 90%+
- [ ] CI enforces coverage threshold
