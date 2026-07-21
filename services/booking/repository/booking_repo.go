package repository

import (
	"context"
	"database/sql"

	"github.com/clinicmanager/services/booking/db"
	shareddb "github.com/clinicmanager/shared/db"
)

type BookingRepository struct {
	db       *shareddb.ScopedDB
	tenantdb *shareddb.ScopedDB
	alldb    *shareddb.ScopedDB
}

func NewBookingRepository(dbs *shareddb.DBSet) *BookingRepository {
	return &BookingRepository{
		db:       dbs.DB,
		tenantdb: dbs.TenantDB,
		alldb:    dbs.AllDB,
	}
}

// BranchHours

func (r *BookingRepository) FindBranchHoursByID(ctx context.Context, id string) (*db.BunBranchHours, error) {
	var m db.BunBranchHours
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindBranchHoursByBranch(ctx context.Context, branchID string) ([]*db.BunBranchHours, error) {
	var list []*db.BunBranchHours
	err := r.tenantdb.NewSelect(ctx, &list).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) ListBranchHours(ctx context.Context) ([]*db.BunBranchHours, error) {
	var list []*db.BunBranchHours
	err := r.db.NewSelect(ctx, &list).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreateBranchHours(ctx context.Context, m *db.BunBranchHours) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdateBranchHours(ctx context.Context, m *db.BunBranchHours) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// PractitionerBranch

func (r *BookingRepository) FindPractitionerBranchByID(ctx context.Context, id string) (*db.BunPractitionerBranch, error) {
	var m db.BunPractitionerBranch
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindPractitionerBranchesByBranch(ctx context.Context, branchID string) ([]*db.BunPractitionerBranch, error) {
	var list []*db.BunPractitionerBranch
	err := r.tenantdb.NewSelect(ctx, &list).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) ListPractitionerBranches(ctx context.Context) ([]*db.BunPractitionerBranch, error) {
	var list []*db.BunPractitionerBranch
	err := r.db.NewSelect(ctx, &list).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreatePractitionerBranch(ctx context.Context, m *db.BunPractitionerBranch) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdatePractitionerBranch(ctx context.Context, m *db.BunPractitionerBranch) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// PractitionerAvailability

func (r *BookingRepository) FindPractitionerAvailabilityByID(ctx context.Context, id string) (*db.BunPractitionerAvailability, error) {
	var m db.BunPractitionerAvailability
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindPractitionerAvailabilitiesByBranch(ctx context.Context, branchID string) ([]*db.BunPractitionerAvailability, error) {
	var list []*db.BunPractitionerAvailability
	err := r.tenantdb.NewSelect(ctx, &list).Where("practitioner_id IN (SELECT practitioner_id FROM practitioner_branch WHERE branch_id = ?)", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) FindPractitionerAvailabilitiesByPractitioner(ctx context.Context, practitionerID string) ([]*db.BunPractitionerAvailability, error) {
	var list []*db.BunPractitionerAvailability
	err := r.db.NewSelect(ctx, &list).Where("practitioner_id = ?", practitionerID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) ListPractitionerAvailabilities(ctx context.Context) ([]*db.BunPractitionerAvailability, error) {
	var list []*db.BunPractitionerAvailability
	err := r.db.NewSelect(ctx, &list).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreatePractitionerAvailability(ctx context.Context, m *db.BunPractitionerAvailability) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdatePractitionerAvailability(ctx context.Context, m *db.BunPractitionerAvailability) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// ScheduleTemplate

func (r *BookingRepository) FindScheduleTemplateByID(ctx context.Context, id string) (*db.BunScheduleTemplate, error) {
	var m db.BunScheduleTemplate
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindScheduleTemplatesByBranch(ctx context.Context, branchID string) ([]*db.BunScheduleTemplate, error) {
	var list []*db.BunScheduleTemplate
	err := r.tenantdb.NewSelect(ctx, &list).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) FindScheduleTemplatesByPractitioner(ctx context.Context, practitionerID string) ([]*db.BunScheduleTemplate, error) {
	var list []*db.BunScheduleTemplate
	err := r.db.NewSelect(ctx, &list).Where("practitioner_id = ?", practitionerID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) ListScheduleTemplates(ctx context.Context) ([]*db.BunScheduleTemplate, error) {
	var list []*db.BunScheduleTemplate
	err := r.db.NewSelect(ctx, &list).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreateScheduleTemplate(ctx context.Context, m *db.BunScheduleTemplate) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdateScheduleTemplate(ctx context.Context, m *db.BunScheduleTemplate) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// AppointmentSlot

func (r *BookingRepository) FindAppointmentSlotByID(ctx context.Context, id string) (*db.BunAppointmentSlot, error) {
	var m db.BunAppointmentSlot
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindAppointmentSlotsByFilter(ctx context.Context, branchID, practitionerID, slotDate, status *string) ([]*db.BunAppointmentSlot, error) {
	var list []*db.BunAppointmentSlot
	query := r.db.NewSelect(ctx, &list)
	if branchID != nil && *branchID != "" {
		query = query.Where("branch_id = ?", *branchID)
	}
	if practitionerID != nil && *practitionerID != "" {
		query = query.Where("practitioner_id = ?", *practitionerID)
	}
	if slotDate != nil && *slotDate != "" {
		query = query.Where("start_at::date = ?::date", *slotDate)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreateAppointmentSlot(ctx context.Context, m *db.BunAppointmentSlot) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdateAppointmentSlot(ctx context.Context, m *db.BunAppointmentSlot) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// Appointment

func (r *BookingRepository) FindAppointmentByID(ctx context.Context, id string) (*db.BunAppointment, error) {
	var m db.BunAppointment
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindAppointmentsByFilter(ctx context.Context, branchID, patientID, practitionerID, status, startAt, endAt *string) ([]*db.BunAppointment, error) {
	var list []*db.BunAppointment
	query := r.db.NewSelect(ctx, &list)
	if branchID != nil && *branchID != "" {
		query = query.Where("branch_id = ?", *branchID)
	}
	if patientID != nil && *patientID != "" {
		query = query.Where("patient_id = ?", *patientID)
	}
	if practitionerID != nil && *practitionerID != "" {
		query = query.Where("practitioner_id = ?", *practitionerID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if startAt != nil && *startAt != "" {
		query = query.Where("start_at >= ?", *startAt)
	}
	if endAt != nil && *endAt != "" {
		query = query.Where("end_at <= ?", *endAt)
	}
	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreateAppointment(ctx context.Context, m *db.BunAppointment) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdateAppointment(ctx context.Context, m *db.BunAppointment) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}

// ScheduleException

func (r *BookingRepository) FindScheduleExceptionByID(ctx context.Context, id string) (*db.BunScheduleException, error) {
	var m db.BunScheduleException
	err := r.tenantdb.NewSelect(ctx, &m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *BookingRepository) FindScheduleExceptionsByBranch(ctx context.Context, branchID string) ([]*db.BunScheduleException, error) {
	var list []*db.BunScheduleException
	err := r.tenantdb.NewSelect(ctx, &list).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) FindScheduleExceptionsByPractitioner(ctx context.Context, practitionerID string) ([]*db.BunScheduleException, error) {
	var list []*db.BunScheduleException
	err := r.db.NewSelect(ctx, &list).Where("practitioner_id = ?", practitionerID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) ListScheduleExceptions(ctx context.Context) ([]*db.BunScheduleException, error) {
	var list []*db.BunScheduleException
	err := r.db.NewSelect(ctx, &list).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) CreateScheduleException(ctx context.Context, m *db.BunScheduleException) error {
	_, err := r.db.NewInsert(m).Exec(ctx)
	return err
}

func (r *BookingRepository) UpdateScheduleException(ctx context.Context, m *db.BunScheduleException) error {
	_, err := r.tenantdb.NewUpdate(ctx, m).Where("id = ?", m.ID).Exec(ctx)
	return err
}
