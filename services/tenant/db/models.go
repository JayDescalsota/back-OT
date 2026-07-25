package db

import (
	"time"

	"github.com/uptrace/bun"
)

// BunTenant maps to the tenants table.
// Fields with json:"-" are DB-only and hidden from the GraphQL schema.
type BunTenant struct {
	bun.BaseModel `bun:"table:tenants"`

	ID     string  `bun:"id,pk"                           json:"id"`
	Name   string  `bun:"name,notnull"                    json:"name"`
	Slug   string  `bun:"slug,notnull,unique"             json:"slug"`
	Domain *string `bun:"domain"                          json:"domain,omitempty"`
	Status string  `bun:"status,notnull,default:'active'" json:"status"`

	CreatedAt     time.Time `bun:"created_at,default:current_timestamp" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at,default:current_timestamp" json:"-"`
	CreatedBy     *string   `bun:"created_by"                           json:"-"`
	CreatedAction string    `bun:"created_action"                       json:"-"`
	UpdatedBy     *string   `bun:"updated_by"                           json:"-"`
	UpdatedAction string    `bun:"updated_action"                       json:"-"`
}

func (BunTenant) IsEntity() {}

// BunBranch maps to the branches table.
type BunBranch struct {
	bun.BaseModel `bun:"table:branches"`

	ID        string  `bun:"id,pk"                           json:"id"`
	TenantID  string  `bun:"tenant_id,notnull"               json:"tenantId"`
	Name      string  `bun:"name,notnull"                    json:"name"`
	AddressID *string `bun:"address_id"                     json:"addressId,omitempty"`
	Timezone  string  `bun:"timezone,notnull,default:'UTC'"  json:"timezone"`
	Phone     *string `bun:"phone"                           json:"phone,omitempty"`
	IsActive  bool    `bun:"is_active,notnull,default:true"  json:"isActive"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunBranch) IsEntity() {}

type BunAddress struct {
	bun.BaseModel `bun:"table:addresses"`

	ID        string `bun:"id,pk"            json:"id"`
	Address   string `bun:"address,notnull"  json:"address"`
	Baranggay string `bun:"baranggay,notnull" json:"baranggay"`
	City      string `bun:"city,notnull"     json:"city"`
	State     string `bun:"state,notnull"    json:"state"`
	ZipCode   string `bun:"zip_code,notnull" json:"zipCode"`
	Country   string `bun:"country,notnull"  json:"country"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunAddress) IsEntity() {}

// BunTenantRole maps to the tenant_roles table.
// Permissions is a relation — gqlgen will generate a field resolver for it.
type BunTenantRole struct {
	bun.BaseModel `bun:"table:tenant_roles"`

	ID           string  `bun:"id,pk"                                json:"id"`
	Name         string  `bun:"name,notnull"                         json:"name"`
	Description  *string `bun:"description"                          json:"description,omitempty"`
	TenantID     *string `bun:"tenant_id"                            json:"tenantId,omitempty"`
	BranchID     *string `bun:"branch_id"                            json:"branchId,omitempty"`
	IsSystemRole bool    `bun:"is_system_role,notnull,default:false" json:"isSystemRole"`

	CreatedAt     time.Time  `bun:"created_at" json:"-"`
	CreatedBy     *string    `bun:"created_by" json:"-"`
	CreatedAction string     `bun:"created_action" json:"-"`
	UpdatedAt     *time.Time `bun:"updated_at" json:"-"`
	UpdatedBy     *string    `bun:"updated_by" json:"-"`
	UpdatedAction string     `bun:"updated_action" json:"-"`
}

func (BunTenantRole) IsEntity() {}

// BunTenantPermission maps to the tenant_permissions table.
type BunTenantPermission struct {
	bun.BaseModel `bun:"table:tenant_permissions"`

	ID          string  `bun:"id,pk"                        json:"id"`
	Resource    string  `bun:"resource,notnull"             json:"resource"`
	Action      string  `bun:"action,notnull"               json:"action"`
	Scope       string  `bun:"scope,notnull,default:'branch'"  json:"scope"`
	Description *string `bun:"description"                  json:"description,omitempty"`
	TenantID    *string `bun:"tenant_id"                    json:"tenantId,omitempty"`
	BranchID    *string `bun:"branch_id"                    json:"branchId,omitempty"`

	CreatedAt     time.Time  `bun:"created_at" json:"-"`
	CreatedBy     *string    `bun:"created_by" json:"-"`
	CreatedAction string     `bun:"created_action" json:"-"`
	UpdatedAt     *time.Time `bun:"updated_at" json:"-"`
	UpdatedBy     *string    `bun:"updated_by" json:"-"`
	UpdatedAction string     `bun:"updated_action" json:"-"`
}

func (BunTenantPermission) IsEntity() {}
