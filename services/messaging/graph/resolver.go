package graph

import "github.com/clinicmanager/services/messaging/service"

type Resolver struct {
	MessagingService *service.MessagingService
}

func NewResolver(svc *service.MessagingService) *Resolver {
	return &Resolver{
		MessagingService: svc,
	}
}
