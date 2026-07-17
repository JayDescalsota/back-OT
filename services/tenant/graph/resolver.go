package graph

import "github.com/clinicmanager/services/tenant/service"

type Resolver struct {
	TenantService *service.TenantService
}

func NewResolver(tenantService *service.TenantService) *Resolver {
	return &Resolver{TenantService: tenantService}
}
