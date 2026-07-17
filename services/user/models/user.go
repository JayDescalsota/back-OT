package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID                     string     `bun:"id,pk"`
	Email                  string     `bun:"email,notnull,unique"`
	Name                   string     `bun:"name,notnull,default:'User'"`
	PasswordHash           string     `bun:"password_hash,notnull"`
	IsActive               bool       `bun:"is_active,default:true"`
	IsValidated            bool       `bun:"is_validated,default:false"`
	ValidatedAt            *time.Time `bun:"validated_at"`
	ValidationToken        string     `bun:"validation_token"`
	PasswordResetAt        *time.Time `bun:"password_reset_at"`
	PasswordResetToken     string     `bun:"password_reset_token"`
	PasswordResetExpiresAt *time.Time `bun:"password_reset_expires_at"`
	LastLogin              *time.Time `bun:"last_login"`
	CreatedAt              time.Time  `bun:"created_at"`
	UpdatedAt              time.Time  `bun:"updated_at"`
	CreatedBy              *string    `bun:"created_by"`
	UpdatedBy              *string    `bun:"updated_by"`
	CreatedAction          string     `bun:"created_action"`
	UpdatedAction          string     `bun:"updated_action"`
}
