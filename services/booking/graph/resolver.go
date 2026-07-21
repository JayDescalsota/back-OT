package graph

import (
	"github.com/clinicmanager/services/booking/service"
)

type Resolver struct {
	BookingService *service.BookingService
}

func NewResolver(bookingService *service.BookingService) *Resolver {
	return &Resolver{
		BookingService: bookingService,
	}
}
