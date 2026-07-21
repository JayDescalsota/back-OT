package graph

import (
	"context"

	"github.com/clinicmanager/services/booking/db"
	"github.com/clinicmanager/services/booking/graph/generated"
)

func (r *entityResolver) FindAppointmentByID(ctx context.Context, id string) (*db.BunAppointment, error) {
	return r.BookingService.GetAppointmentByID(ctx, id)
}

func (r *entityResolver) FindAppointmentSlotByID(ctx context.Context, id string) (*db.BunAppointmentSlot, error) {
	return r.BookingService.GetAppointmentSlotByID(ctx, id)
}

func (r *entityResolver) FindBranchHoursByID(ctx context.Context, id string) (*db.BunBranchHours, error) {
	return r.BookingService.GetBranchHoursByID(ctx, id)
}

func (r *entityResolver) FindPractitionerAvailabilityByID(ctx context.Context, id string) (*db.BunPractitionerAvailability, error) {
	return r.BookingService.GetPractitionerAvailabilityByID(ctx, id)
}

func (r *entityResolver) FindPractitionerBranchByID(ctx context.Context, id string) (*db.BunPractitionerBranch, error) {
	return r.BookingService.GetPractitionerBranchByID(ctx, id)
}

func (r *entityResolver) FindScheduleExceptionByID(ctx context.Context, id string) (*db.BunScheduleException, error) {
	return r.BookingService.GetScheduleExceptionByID(ctx, id)
}

func (r *entityResolver) FindScheduleTemplateByID(ctx context.Context, id string) (*db.BunScheduleTemplate, error) {
	return r.BookingService.GetScheduleTemplateByID(ctx, id)
}

func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
