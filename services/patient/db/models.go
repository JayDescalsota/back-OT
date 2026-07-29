package db

import (
	"time"

	"github.com/uptrace/bun"
)

type BunPatients struct {
	bun.BaseModel `bun:"table:patients"`

	ID          string    `bun:"id,pk" json:"id"`
	TenantID    string    `bun:"tenant_id" json:"tenant_id"`
	BranchID    string    `bun:"branch_id" json:"branch_id"`
	FirstName   string    `bun:"first_name,notnull" json:"first_name"`
	LastName    string    `bun:"last_name,notnull" json:"last_name"`
	DateOfBirth time.Time `bun:"date_of_birth" json:"date_of_birth"`
	Gender      string    `bun:"gender,notnull" json:"gender"`
	Notes       *string   `bun:"notes" json:"notes,omitempty"`
	Height      *string   `bun:"height" json:"height,omitempty"`
	Weight      *string   `bun:"weight" json:"weight,omitempty"`
	IsActive    bool      `bun:"is_active,default:true" json:"is_active"`
	AddressID   *string   `bun:"address_id" json:"addressId,omitempty"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunPatients) IsEntity() {}

type BunGuardians struct {
	bun.BaseModel `bun:"table:patient_guardian_profiles"`

	ID        string  `bun:"id,pk" json:"id"`
	FirstName string  `bun:"first_name,notnull" json:"first_name"`
	LastName  string  `bun:"last_name,notnull" json:"last_name"`
	Gender    string  `bun:"gender,notnull" json:"gender"`
	Email     *string `bun:"email" json:"email,omitempty"`
	Phone     string  `bun:"phone" json:"phone,omitempty"`
	Notes     *string `bun:"notes" json:"notes,omitempty"`
	IsActive  bool    `bun:"is_active,default:true" json:"is_active"`
	AddressID *string `bun:"address_id" json:"addressId,omitempty"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunGuardians) IsEntity() {}

type BunPatientGuardians struct {
	bun.BaseModel `bun:"table:patient_guardian_links"`

	PatientID    string `bun:"patient_id,pk" json:"patientId"`
	GuardianID   string `bun:"guardian_id,pk" json:"guardianId"`
	Relationship string `bun:"relationship,notnull" json:"relationship"`
	IsActive     bool   `bun:"is_active,default:true" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

type BunPatientTags struct {
	bun.BaseModel `bun:"table:patient_tags"`

	PatientID string `bun:"patient_id,pk" json:"patientId"`
	TagID     string `bun:"tag_id,pk" json:"tagId"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}
