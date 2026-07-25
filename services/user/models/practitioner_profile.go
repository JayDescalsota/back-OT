package models

import (
	"time"

	"github.com/uptrace/bun"
)

type PractitionerProfile struct {
	bun.BaseModel `bun:"table:practitioner_profiles"`

	UserID              string    `bun:"user_id,pk"`
	LicenseNumber       *string   `bun:"license_number"`
	LicenseState        *string   `bun:"license_state"`
	NpiNumber           *string   `bun:"npi_number,unique"`
	Specialty           *string   `bun:"specialty"`
	SubSpecialty        *string   `bun:"sub_specialty"`
	Qualifications      []string  `bun:"qualifications,array"`
	Credentials         []string  `bun:"credentials,array"`
	Education           *string   `bun:"education"`
	YearsOfExperience   *int      `bun:"years_of_experience"`
	Bio                 *string   `bun:"bio"`
	IsAcceptingPatients bool      `bun:"is_accepting_patients,default:true"`
	CreatedAt           time.Time `bun:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at"`
	CreatedBy           *string   `bun:"created_by"`
	UpdatedBy           *string   `bun:"updated_by"`
	CreatedAction       string    `bun:"created_action"`
	UpdatedAction       string    `bun:"updated_action"`
}
