package models

import (
	"time"

	"github.com/uptrace/bun"
)

type AppRole struct {
	bun.BaseModel `bun:"table:user_roles"`

	ID            int        `bun:"id,pk,autoincrement"`
	Name          string     `bun:"name,notnull,unique"`
	Description   string     `bun:"description"`
	CreatedAt     time.Time  `bun:"created_at,default:now()"`
	CreatedBy     *string    `bun:"created_by"`
	CreatedAction string     `bun:"created_action"`
	UpdatedAt     *time.Time `bun:"updated_at"`
	UpdatedBy     *string    `bun:"updated_by"`
	UpdatedAction string     `bun:"updated_action"`
}

type UserAppRole struct {
	bun.BaseModel `bun:"table:user_role_assignments"`

	UserID        string     `bun:"user_id,pk"`
	AppRoleID     int        `bun:"app_role_id,pk"`
	AssignedBy    *string    `bun:"assigned_by"`
	AssignedAt    time.Time  `bun:"assigned_at,default:now()"`
	UpdatedAt     *time.Time `bun:"updated_at"`
	UpdatedBy     *string    `bun:"updated_by"`
	UpdatedAction string     `bun:"updated_action"`
}
