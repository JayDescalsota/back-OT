package db

import (
	"time"

	"github.com/uptrace/bun"
)

type BunMessageThread struct {
	bun.BaseModel `bun:"table:message_thread"`

	ID            string    `bun:"id,pk" json:"id"`
	TenantID      string    `bun:"tenant_id" json:"tenant_id"`
	BranchID      string    `bun:"branch_id" json:"branch_id"`
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
	bun.BaseModel `bun:"table:message"`

	ID       string `bun:"id,pk" json:"id"`
	TenantID string `bun:"tenant_id" json:"tenant_id"`
	ThreadID string `bun:"thread_id" json:"thread_id"`
	SenderID string `bun:"sender_id" json:"sender_id"`
	Body     string `bun:"body,notnull" json:"-"`  // encrypted (base64)
	Nonce    string `bun:"nonce,notnull" json:"-"` // AES-GCM nonce (base64)
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
	bun.BaseModel `bun:"table:message_participant"`

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
