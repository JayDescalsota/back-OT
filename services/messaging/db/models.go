package db

import (
	"time"

	"github.com/uptrace/bun"
)

type BunMessageThread struct {
	bun.BaseModel `bun:"table:messaging_threads"`

	ID            string    `bun:"id,pk" json:"id"`
	TenantID      string    `bun:"tenant_id" json:"tenant_id"`
	BranchID      string    `bun:"branch_id" json:"branch_id"`
	Title         string    `bun:"subject" json:"title"`
	Type          string    `bun:"type" json:"type"`
	IsActive      bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunMessageThread) IsEntity() {}

type BunMessage struct {
	bun.BaseModel `bun:"table:messaging_messages"`

	ID       string `bun:"id,pk" json:"id"`
	TenantID string `bun:"tenant_id" json:"tenant_id"`
	ThreadID string `bun:"thread_id" json:"thread_id"`
	SenderID string `bun:"sender_id" json:"sender_id"`
	Body     string `bun:"body,notnull" json:"body"`
	Nonce    string `bun:"nonce,notnull" json:"nonce"`
	IsActive bool   `bun:"is_active,default:true" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunMessage) IsEntity() {}

type BunMessageParticipant struct {
	bun.BaseModel `bun:"table:messaging_participants"`

	ID            string    `bun:"id,pk" json:"id"`
	TenantID      string    `bun:"tenant_id" json:"tenant_id"`
	ThreadID      string    `bun:"thread_id" json:"thread_id"`
	ParticipantID string    `bun:"participant_id" json:"participant_id"`
	Role          string    `bun:"role" json:"role"`
	LastReadAt    time.Time `bun:"last_read_at" json:"last_read_at"`
	IsActive      bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunMessageParticipant) IsEntity() {}

type BunUserPublicKey struct {
	bun.BaseModel `bun:"table:messaging_user_public_keys"`

	UserID    string    `bun:"user_id,pk" json:"user_id"`
	PublicKey string    `bun:"public_key,notnull" json:"public_key"`
	CreatedAt time.Time `bun:"created_at" json:"-"`
}

func (BunUserPublicKey) IsEntity() {}

type BunThreadKey struct {
	bun.BaseModel `bun:"table:messaging_thread_keys"`

	ThreadID     string    `bun:"thread_id,pk" json:"thread_id"`
	UserID       string    `bun:"user_id,pk" json:"user_id"`
	EncryptedKey string    `bun:"encrypted_key,notnull" json:"encrypted_key"`
	CreatedAt    time.Time `bun:"created_at" json:"-"`
}

func (BunThreadKey) IsEntity() {}
