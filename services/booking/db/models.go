package db

import (
	"time"

	"github.com/uptrace/bun"
)

// BunBranchHours defines operating hours for a clinic branch (never booked directly).
type BunBranchHours struct {
	bun.BaseModel  `bun:"table:booking_branch_hours"`
	ID             string    `bun:"id,pk" json:"id"`
	TenantID       string    `bun:"tenant_id" json:"tenant_id"`
	BranchID       string    `bun:"branch_id" json:"branch_id"`
	RecurrenceRule string    `bun:"recurrence_rule" json:"recurrence_rule"`
	OpenTime       time.Time `bun:"open_time" json:"open_time"`
	CloseTime      time.Time `bun:"close_time" json:"close_time"`
	EffectiveFrom  time.Time `bun:"effective_from" json:"effective_from"`
	EffectiveUntil time.Time `bun:"effective_until,nullzero" json:"effective_until"`
	IsActive       bool      `bun:"is_active" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunBranchHours) IsEntity() {}

// BunPractitionerBranch defines practitioner branch affiliation.
type BunPractitionerBranch struct {
	bun.BaseModel  `bun:"table:booking_practitioner_branches"`
	ID             string    `bun:"id,pk" json:"id"`
	TenantID       string    `bun:"tenant_id" json:"tenant_id"`
	PractitionerID string    `bun:"practitioner_id" json:"practitioner_id"`
	BranchID       string    `bun:"branch_id" json:"branch_id"`
	Role           *string   `bun:"role" json:"role"`
	IsPrimary      bool      `bun:"is_primary" json:"is_primary"`
	IsActive       bool      `bun:"is_active" json:"is_active"`
	JoinedAt       time.Time `bun:"joined_at" json:"joined_at"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunPractitionerBranch) IsEntity() {}

// BunPractitionerAvailability represents practitioner-managed availability shifts ("When am I available to work?").
type BunPractitionerAvailability struct {
	bun.BaseModel  `bun:"table:booking_practitioner_availabilities"`
	ID             string    `bun:"id,pk" json:"id"`
	TenantID       string    `bun:"tenant_id" json:"tenant_id"`
	PractitionerID string    `bun:"practitioner_id" json:"practitioner_id"`
	RecurrenceRule string    `bun:"recurrence_rule" json:"recurrence_rule"`
	AvailableFrom  time.Time `bun:"available_from" json:"available_from"`
	AvailableTo    time.Time `bun:"available_to" json:"available_to"`
	EffectiveFrom  time.Time `bun:"effective_from" json:"effective_from"`
	EffectiveUntil time.Time `bun:"effective_until,nullzero" json:"effective_until"`
	Status         string    `bun:"status" json:"status"` // ACTIVE / INACTIVE
	IsActive       bool      `bun:"is_active" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunPractitionerAvailability) IsEntity() {}

// BunScheduleTemplate defines recurring schedules that patients can book (PractitionerID nullable for clinic-first).
type BunScheduleTemplate struct {
	bun.BaseModel  `bun:"table:booking_schedule_templates"`
	ID             string    `bun:"id,pk" json:"id"`
	TenantID       string    `bun:"tenant_id" json:"tenant_id"`
	BranchID       string    `bun:"branch_id" json:"branch_id"`
	PractitionerID *string   `bun:"practitioner_id" json:"practitioner_id"` // Nullable for clinic-first
	RecurrenceRule string    `bun:"recurrence_rule" json:"recurrence_rule"`
	StartTime      time.Time `bun:"start_time" json:"start_time"`
	EndTime        time.Time `bun:"end_time" json:"end_time"`
	SlotDuration   int       `bun:"slot_duration" json:"slot_duration"`     // Minutes
	BufferDuration int       `bun:"buffer_duration" json:"buffer_duration"` // Minutes
	Capacity       int       `bun:"capacity" json:"capacity"`
	BookedCount    int       `bun:"booked_count" json:"booked_count"`
	EffectiveFrom  time.Time `bun:"effective_from" json:"effective_from"`
	EffectiveUntil time.Time `bun:"effective_until,nullzero" json:"effective_until"`
	Status         string    `bun:"status" json:"status"` // ACTIVE / INACTIVE / DRAFT
	IsActive       bool      `bun:"is_active" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunScheduleTemplate) IsEntity() {}

// BunAppointmentSlot represents actual generated bookable time slots generated from ScheduleTemplates.
type BunAppointmentSlot struct {
	bun.BaseModel      `bun:"table:booking_appointment_slots"`
	ID                 string    `bun:"id,pk" json:"id"`
	TenantID           string    `bun:"tenant_id" json:"tenant_id"`
	ScheduleTemplateID string    `bun:"schedule_template_id" json:"schedule_template_id"`
	BranchID           string    `bun:"branch_id" json:"branch_id"`
	PractitionerID     *string   `bun:"practitioner_id" json:"practitioner_id"` // Nullable
	StartAt            time.Time `bun:"start_at" json:"start_at"`
	EndAt              time.Time `bun:"end_at" json:"end_at"`
	Capacity           int       `bun:"capacity" json:"capacity"`
	BookedCount        int       `bun:"booked_count" json:"booked_count"`
	Status             string    `bun:"status" json:"status"` // AVAILABLE / BLOCKED / CANCELLED
	GeneratedAt        time.Time `bun:"generated_at" json:"generated_at"`
	IsActive           bool      `bun:"is_active" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunAppointmentSlot) IsEntity() {}

// BunAppointment represents patient bookings against appointment slots.
type BunAppointment struct {
	bun.BaseModel      `bun:"table:booking_appointments"`
	ID                 string     `bun:"id,pk" json:"id"`
	TenantID           string     `bun:"tenant_id" json:"tenant_id"`
	AppointmentSlotID  *string    `bun:"appointment_slot_id" json:"appointment_slot_id"`
	BranchID           string     `bun:"branch_id" json:"branch_id"`
	PatientID          *string    `bun:"patient_id" json:"patient_id"`
	PractitionerID     *string    `bun:"practitioner_id" json:"practitioner_id"`
	StartAt            time.Time  `bun:"start_at" json:"start_at"`
	EndAt              time.Time  `bun:"end_at" json:"end_at"`
	Status             string     `bun:"status" json:"status"` // PENDING / CONFIRMED / IN_PROGRESS / DONE / CANCELLED / RESCHEDULED
	Notes              string     `bun:"notes" json:"notes"`
	CancellationReason string     `bun:"cancellation_reason" json:"cancellation_reason"`
	CancelledAt        *time.Time `bun:"cancelled_at" json:"cancelled_at"`
	RescheduledFromID  *string    `bun:"rescheduled_from_id" json:"rescheduled_from_id"`
	RescheduledToID    *string    `bun:"rescheduled_to_id" json:"rescheduled_to_id"`
	RescheduleReason   string     `bun:"reschedule_reason" json:"reschedule_reason"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunAppointment) IsEntity() {}

// BunScheduleException represents one-time schedule overrides (vacation, lunch, holiday, emergency closure).
type BunScheduleException struct {
	bun.BaseModel      `bun:"table:booking_schedule_exceptions"`
	ID                 string    `bun:"id,pk" json:"id"`
	TenantID           string    `bun:"tenant_id" json:"tenant_id"`
	ScheduleTemplateID *string   `bun:"schedule_template_id" json:"schedule_template_id"`
	AppointmentSlotID  *string   `bun:"appointment_slot_id" json:"appointment_slot_id"`
	PractitionerID     *string   `bun:"practitioner_id" json:"practitioner_id"`
	Type               string    `bun:"type" json:"type"` // VACATION / LUNCH_BREAK / HOLIDAY / EMERGENCY_CLOSURE / OVERRIDE
	Reason             string    `bun:"reason" json:"reason"`
	StartAt            time.Time `bun:"start_at" json:"start_at"`
	EndAt              time.Time `bun:"end_at" json:"end_at"`
	Status             string    `bun:"status" json:"status"` // ACTIVE / CANCELLED
	IsActive           bool      `bun:"is_active" json:"is_active"`

	CreatedAt     time.Time `bun:"created_at" json:"-"`
	UpdatedAt     time.Time `bun:"updated_at" json:"-"`
	CreatedBy     *string   `bun:"created_by" json:"-"`
	CreatedAction string    `bun:"created_action" json:"-"`
	UpdatedBy     *string   `bun:"updated_by" json:"-"`
	UpdatedAction string    `bun:"updated_action" json:"-"`
}

func (BunScheduleException) IsEntity() {}
