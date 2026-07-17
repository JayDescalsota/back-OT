package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/clinicmanager/services/user/repository"
)

func newMockDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	bunDB := bun.NewDB(db, pgdialect.New())
	t.Cleanup(func() { db.Close() })
	return bunDB, mock
}

func newUserRepo(t *testing.T) (*repository.UserRepo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newMockDB(t)
	return repository.NewUserRepo(db), mock
}

func TestUserRepo_FindUserByID_Found(t *testing.T) {
	repo, mock := newUserRepo(t)
	id, email, name := uuid.New().String(), "alice@example.com", "Alice"
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM "users" .+ WHERE .+`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "name", "password_hash",
			"is_active", "last_login", "created_at", "updated_at",
		}).AddRow(id, email, name, "hash", true, nil, now, now))

	user, err := repo.FindUserByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != email {
		t.Errorf("expected email %s, got %s", email, user.Email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepo_FindUserByID_NotFound(t *testing.T) {
	repo, mock := newUserRepo(t)
	id := uuid.New().String()

	mock.ExpectQuery(`SELECT .+ FROM "users" .+ WHERE .+`).
		WillReturnError(sqlmock.ErrCancelled)

	user, err := repo.FindUserByID(context.Background(), id)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
