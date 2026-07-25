package models

import (
	"time"

	"github.com/uptrace/bun"
)

type UserProfile struct {
	bun.BaseModel `bun:"table:user_profiles"`

	UserID                   string     `bun:"user_id,pk"`
	FirstName                string     `bun:"first_name,notnull"`
	LastName                 *string    `bun:"last_name"`
	MiddleName               *string    `bun:"middle_name"`
	Suffix                   *string    `bun:"suffix"`
	Title                    *string    `bun:"title"`
	Phone                    *string    `bun:"phone"`
	Mobile                   *string    `bun:"mobile"`
	DateOfBirth              *time.Time `bun:"date_of_birth"`
	Gender                   *string    `bun:"gender"`
	AddressID                *string    `bun:"address_id"`
	Timezone                 string     `bun:"timezone,default:'Asia/Manila'"`
	PreferredLanguage        string     `bun:"preferred_language,default:'en'"`
	EmergencyContactName     *string    `bun:"emergency_contact_name"`
	EmergencyContactPhone    *string    `bun:"emergency_contact_phone"`
	EmergencyContactRelation *string    `bun:"emergency_contact_relation"`
	Notes                    *string    `bun:"notes"`
	CreatedAt                time.Time  `bun:"created_at"`
	UpdatedAt                time.Time  `bun:"updated_at"`
	CreatedBy                *string    `bun:"created_by"`
	UpdatedBy                *string    `bun:"updated_by"`
	CreatedAction            string     `bun:"created_action"`
	UpdatedAction            string     `bun:"updated_action"`
}
