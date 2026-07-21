package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/clinicmanager/services/booking/db"
	"github.com/clinicmanager/services/booking/graph/generated"
	"github.com/clinicmanager/services/booking/graph/model"
)

// BranchHours field resolvers

func (r *branchHoursResolver) DayOfWeek(ctx context.Context, obj *db.BunBranchHours) (int, error) {
	days := []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}
	for i, d := range days {
		for j := 0; j < len(obj.RecurrenceRule)-1; j++ {
			if obj.RecurrenceRule[j:j+2] == d {
				return i, nil
			}
		}
	}
	return 0, nil
}

func (r *branchHoursResolver) OpenTime(ctx context.Context, obj *db.BunBranchHours) (*string, error) {
	if obj.OpenTime.IsZero() {
		return nil, nil
	}
	s := obj.OpenTime.Format("15:04")
	return &s, nil
}

func (r *branchHoursResolver) CloseTime(ctx context.Context, obj *db.BunBranchHours) (*string, error) {
	if obj.CloseTime.IsZero() {
		return nil, nil
	}
	s := obj.CloseTime.Format("15:04")
	return &s, nil
}

func (r *branchHoursResolver) CreatedAt(ctx context.Context, obj *db.BunBranchHours) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *branchHoursResolver) UpdatedAt(ctx context.Context, obj *db.BunBranchHours) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// PractitionerBranch field resolvers

func (r *practitionerBranchResolver) CreatedAt(ctx context.Context, obj *db.BunPractitionerBranch) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *practitionerBranchResolver) UpdatedAt(ctx context.Context, obj *db.BunPractitionerBranch) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// PractitionerAvailability field resolvers

func (r *practitionerAvailabilityResolver) BranchID(ctx context.Context, obj *db.BunPractitionerAvailability) (string, error) {
	return "", nil
}

func (r *practitionerAvailabilityResolver) StartTime(ctx context.Context, obj *db.BunPractitionerAvailability) (string, error) {
	return obj.AvailableFrom.Format(time.RFC3339), nil
}

func (r *practitionerAvailabilityResolver) EndTime(ctx context.Context, obj *db.BunPractitionerAvailability) (string, error) {
	return obj.AvailableTo.Format(time.RFC3339), nil
}

func (r *practitionerAvailabilityResolver) CreatedAt(ctx context.Context, obj *db.BunPractitionerAvailability) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *practitionerAvailabilityResolver) UpdatedAt(ctx context.Context, obj *db.BunPractitionerAvailability) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// ScheduleTemplate field resolvers

func (r *scheduleTemplateResolver) TemplateName(ctx context.Context, obj *db.BunScheduleTemplate) (string, error) {
	return fmt.Sprintf("Schedule %s", obj.ID[:8]), nil
}

func (r *scheduleTemplateResolver) DayOfWeek(ctx context.Context, obj *db.BunScheduleTemplate) (int, error) {
	days := []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}
	for i, d := range days {
		for j := 0; j < len(obj.RecurrenceRule)-1; j++ {
			if obj.RecurrenceRule[j:j+2] == d {
				return i, nil
			}
		}
	}
	return 0, nil
}

func (r *scheduleTemplateResolver) StartTime(ctx context.Context, obj *db.BunScheduleTemplate) (string, error) {
	return obj.StartTime.Format("15:04"), nil
}

func (r *scheduleTemplateResolver) EndTime(ctx context.Context, obj *db.BunScheduleTemplate) (string, error) {
	return obj.EndTime.Format("15:04"), nil
}

func (r *scheduleTemplateResolver) SlotDurationMinutes(ctx context.Context, obj *db.BunScheduleTemplate) (int, error) {
	return obj.SlotDuration, nil
}

func (r *scheduleTemplateResolver) BreakStart(ctx context.Context, obj *db.BunScheduleTemplate) (*string, error) {
	return nil, nil
}

func (r *scheduleTemplateResolver) BreakEnd(ctx context.Context, obj *db.BunScheduleTemplate) (*string, error) {
	return nil, nil
}

func (r *scheduleTemplateResolver) CreatedAt(ctx context.Context, obj *db.BunScheduleTemplate) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *scheduleTemplateResolver) UpdatedAt(ctx context.Context, obj *db.BunScheduleTemplate) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// AppointmentSlot field resolvers

func (r *appointmentSlotResolver) SlotDate(ctx context.Context, obj *db.BunAppointmentSlot) (string, error) {
	return obj.StartAt.Format("2006-01-02"), nil
}

func (r *appointmentSlotResolver) SlotStart(ctx context.Context, obj *db.BunAppointmentSlot) (string, error) {
	return obj.StartAt.Format(time.RFC3339), nil
}

func (r *appointmentSlotResolver) SlotEnd(ctx context.Context, obj *db.BunAppointmentSlot) (string, error) {
	return obj.EndAt.Format(time.RFC3339), nil
}

func (r *appointmentSlotResolver) AppointmentID(ctx context.Context, obj *db.BunAppointmentSlot) (*string, error) {
	return nil, nil
}

func (r *appointmentSlotResolver) CreatedAt(ctx context.Context, obj *db.BunAppointmentSlot) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *appointmentSlotResolver) UpdatedAt(ctx context.Context, obj *db.BunAppointmentSlot) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// Appointment field resolvers

func (r *appointmentResolver) ScheduledStart(ctx context.Context, obj *db.BunAppointment) (string, error) {
	return obj.StartAt.Format(time.RFC3339), nil
}

func (r *appointmentResolver) ScheduledEnd(ctx context.Context, obj *db.BunAppointment) (string, error) {
	return obj.EndAt.Format(time.RFC3339), nil
}

func (r *appointmentResolver) Note(ctx context.Context, obj *db.BunAppointment) (*string, error) {
	if obj.Notes == "" {
		return nil, nil
	}
	return &obj.Notes, nil
}

func (r *appointmentResolver) SlotID(ctx context.Context, obj *db.BunAppointment) (*string, error) {
	if obj.AppointmentSlotID == "" {
		return nil, nil
	}
	return &obj.AppointmentSlotID, nil
}

func (r *appointmentResolver) CreatedAt(ctx context.Context, obj *db.BunAppointment) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *appointmentResolver) UpdatedAt(ctx context.Context, obj *db.BunAppointment) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// ScheduleException field resolvers

func (r *scheduleExceptionResolver) BranchID(ctx context.Context, obj *db.BunScheduleException) (string, error) {
	return "", nil
}

func (r *scheduleExceptionResolver) ExceptionDate(ctx context.Context, obj *db.BunScheduleException) (string, error) {
	return obj.StartAt.Format("2006-01-02"), nil
}

func (r *scheduleExceptionResolver) CreatedAt(ctx context.Context, obj *db.BunScheduleException) (string, error) {
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (r *scheduleExceptionResolver) UpdatedAt(ctx context.Context, obj *db.BunScheduleException) (*string, error) {
	if obj.UpdatedAt.IsZero() {
		return nil, nil
	}
	s := obj.UpdatedAt.Format(time.RFC3339)
	return &s, nil
}

// Query resolvers

func (r *queryResolver) BranchHour(ctx context.Context, id string) (*db.BunBranchHours, error) {
	return r.BookingService.GetBranchHoursByID(ctx, id)
}

func (r *queryResolver) BranchHours(ctx context.Context, branchID *string) ([]*db.BunBranchHours, error) {
	if branchID != nil && *branchID != "" {
		return r.BookingService.GetBranchHoursByBranch(ctx, *branchID)
	}
	return r.BookingService.ListBranchHours(ctx)
}

func (r *queryResolver) PractitionerBranch(ctx context.Context, id string) (*db.BunPractitionerBranch, error) {
	return r.BookingService.GetPractitionerBranchByID(ctx, id)
}

func (r *queryResolver) PractitionerBranches(ctx context.Context, branchID *string) ([]*db.BunPractitionerBranch, error) {
	if branchID != nil && *branchID != "" {
		return r.BookingService.GetPractitionerBranchesByBranch(ctx, *branchID)
	}
	return r.BookingService.ListPractitionerBranches(ctx)
}

func (r *queryResolver) PractitionerAvailability(ctx context.Context, id string) (*db.BunPractitionerAvailability, error) {
	return r.BookingService.GetPractitionerAvailabilityByID(ctx, id)
}

func (r *queryResolver) PractitionerAvailabilities(ctx context.Context, branchID *string, practitionerID *string) ([]*db.BunPractitionerAvailability, error) {
	return r.BookingService.GetPractitionerAvailabilities(ctx, branchID, practitionerID)
}

func (r *queryResolver) ScheduleTemplate(ctx context.Context, id string) (*db.BunScheduleTemplate, error) {
	return r.BookingService.GetScheduleTemplateByID(ctx, id)
}

func (r *queryResolver) ScheduleTemplates(ctx context.Context, branchID *string, practitionerID *string) ([]*db.BunScheduleTemplate, error) {
	return r.BookingService.GetScheduleTemplates(ctx, branchID, practitionerID)
}

func (r *queryResolver) AppointmentSlot(ctx context.Context, id string) (*db.BunAppointmentSlot, error) {
	return r.BookingService.GetAppointmentSlotByID(ctx, id)
}

func (r *queryResolver) AppointmentSlots(ctx context.Context, filter model.AppointmentSlotsFilter) ([]*db.BunAppointmentSlot, error) {
	return r.BookingService.GetAppointmentSlots(ctx, filter)
}

func (r *queryResolver) Appointment(ctx context.Context, id string) (*db.BunAppointment, error) {
	return r.BookingService.GetAppointmentByID(ctx, id)
}

func (r *queryResolver) Appointments(ctx context.Context, filter model.AppointmentsFilter) ([]*db.BunAppointment, error) {
	return r.BookingService.GetAppointments(ctx, filter)
}

func (r *queryResolver) ScheduleException(ctx context.Context, id string) (*db.BunScheduleException, error) {
	return r.BookingService.GetScheduleExceptionByID(ctx, id)
}

func (r *queryResolver) ScheduleExceptions(ctx context.Context, branchID *string, practitionerID *string) ([]*db.BunScheduleException, error) {
	return r.BookingService.GetScheduleExceptions(ctx, branchID, practitionerID)
}

// Mutation resolvers

func (r *mutationResolver) CreateBranchHours(ctx context.Context, input model.BranchHoursInput) (*db.BunBranchHours, error) {
	return r.BookingService.CreateBranchHours(ctx, input)
}

func (r *mutationResolver) UpdateBranchHours(ctx context.Context, id string, input model.BranchHoursUpdateInput) (*db.BunBranchHours, error) {
	return r.BookingService.UpdateBranchHours(ctx, id, input)
}

func (r *mutationResolver) DeleteBranchHours(ctx context.Context, id string) (bool, error) {
	return r.BookingService.DeleteBranchHours(ctx, id)
}

func (r *mutationResolver) CreatePractitionerBranch(ctx context.Context, input model.PractitionerBranchInput) (*db.BunPractitionerBranch, error) {
	return r.BookingService.CreatePractitionerBranch(ctx, input)
}

func (r *mutationResolver) DeletePractitionerBranch(ctx context.Context, id string) (bool, error) {
	return r.BookingService.DeletePractitionerBranch(ctx, id)
}

func (r *mutationResolver) CreatePractitionerAvailability(ctx context.Context, input model.PractitionerAvailabilityInput) (*db.BunPractitionerAvailability, error) {
	return r.BookingService.CreatePractitionerAvailability(ctx, input)
}

func (r *mutationResolver) DeletePractitionerAvailability(ctx context.Context, id string) (bool, error) {
	return r.BookingService.DeletePractitionerAvailability(ctx, id)
}

func (r *mutationResolver) CreateScheduleTemplate(ctx context.Context, input model.ScheduleTemplateInput) (*db.BunScheduleTemplate, error) {
	return r.BookingService.CreateScheduleTemplate(ctx, input)
}

func (r *mutationResolver) UpdateScheduleTemplate(ctx context.Context, id string, input model.ScheduleTemplateUpdateInput) (*db.BunScheduleTemplate, error) {
	return r.BookingService.UpdateScheduleTemplate(ctx, id, input)
}

func (r *mutationResolver) DeleteScheduleTemplate(ctx context.Context, id string) (bool, error) {
	return r.BookingService.DeleteScheduleTemplate(ctx, id)
}

func (r *mutationResolver) GenerateSlots(ctx context.Context, branchID string, practitionerID string, date string) ([]*db.BunAppointmentSlot, error) {
	return r.BookingService.GenerateSlots(ctx, branchID, practitionerID, date)
}

func (r *mutationResolver) UpdateSlotStatus(ctx context.Context, id string, status string) (*db.BunAppointmentSlot, error) {
	return r.BookingService.UpdateSlotStatus(ctx, id, status)
}

func (r *mutationResolver) CreateAppointment(ctx context.Context, input model.AppointmentInput) (*db.BunAppointment, error) {
	return r.BookingService.CreateAppointment(ctx, input)
}

func (r *mutationResolver) UpdateAppointment(ctx context.Context, id string, input model.AppointmentUpdateInput) (*db.BunAppointment, error) {
	return r.BookingService.UpdateAppointment(ctx, id, input)
}

func (r *mutationResolver) CancelAppointment(ctx context.Context, id string) (*db.BunAppointment, error) {
	return r.BookingService.CancelAppointment(ctx, id)
}

func (r *mutationResolver) CreateScheduleException(ctx context.Context, input model.ScheduleExceptionInput) (*db.BunScheduleException, error) {
	return r.BookingService.CreateScheduleException(ctx, input)
}

func (r *mutationResolver) DeleteScheduleException(ctx context.Context, id string) (bool, error) {
	return r.BookingService.DeleteScheduleException(ctx, id)
}

// Interface implementations

func (r *Resolver) Appointment() generated.AppointmentResolver { return &appointmentResolver{r} }

func (r *Resolver) AppointmentSlot() generated.AppointmentSlotResolver {
	return &appointmentSlotResolver{r}
}

func (r *Resolver) BranchHours() generated.BranchHoursResolver { return &branchHoursResolver{r} }

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) PractitionerAvailability() generated.PractitionerAvailabilityResolver {
	return &practitionerAvailabilityResolver{r}
}

func (r *Resolver) PractitionerBranch() generated.PractitionerBranchResolver {
	return &practitionerBranchResolver{r}
}

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

func (r *Resolver) ScheduleException() generated.ScheduleExceptionResolver {
	return &scheduleExceptionResolver{r}
}

func (r *Resolver) ScheduleTemplate() generated.ScheduleTemplateResolver {
	return &scheduleTemplateResolver{r}
}

type (
	appointmentResolver              struct{ *Resolver }
	appointmentSlotResolver          struct{ *Resolver }
	branchHoursResolver              struct{ *Resolver }
	mutationResolver                 struct{ *Resolver }
	practitionerAvailabilityResolver struct{ *Resolver }
	practitionerBranchResolver       struct{ *Resolver }
	queryResolver                    struct{ *Resolver }
	scheduleExceptionResolver        struct{ *Resolver }
	scheduleTemplateResolver         struct{ *Resolver }
)
