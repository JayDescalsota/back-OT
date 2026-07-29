package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/clinicmanager/services/patient/db"
	"github.com/clinicmanager/services/patient/graph/model"
	shareddb "github.com/clinicmanager/shared/db"
	bun "github.com/uptrace/bun"
)

type PatientRepository struct {
	db       *shareddb.ScopedDB
	tenantdb *shareddb.ScopedDB
	alldb    *shareddb.ScopedDB
}

func NewPatientRepository(dbs *shareddb.DBSet) *PatientRepository {
	return &PatientRepository{
		db:       dbs.DB,
		tenantdb: dbs.TenantDB,
		alldb:    dbs.AllDB,
	}
}

func (r *PatientRepository) FindPatientByID(ctx context.Context, id string) (*db.BunPatients, error) {
	var patient db.BunPatients
	err := r.db.NewSelect(ctx, &patient).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &patient, nil
}

func (r *PatientRepository) GetPatientsByIDs(ctx context.Context, ids []string) ([]*db.BunPatients, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var patients []*db.BunPatients
	err := r.db.NewSelect(ctx, &patients).Where("id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) FindPatientsByBranch(ctx context.Context, branchID string) ([]*db.BunPatients, error) {
	var patients []*db.BunPatients
	err := r.tenantdb.NewSelect(ctx, &patients).Where("branch_id = ?", branchID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) FindGuardianByID(ctx context.Context, id string) (*db.BunGuardians, error) {
	var guardian db.BunGuardians
	err := r.db.NewSelect(ctx, &guardian).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &guardian, nil
}

func (r *PatientRepository) FindGuardianPatients(ctx context.Context, guardianID string) ([]*db.BunPatients, error) {
	var patients []*db.BunPatients
	err := r.db.NewSelect(ctx, &patients).
		Join("JOIN patient_guardian_links pg ON pg.patient_id = bun_patients.id").
		Where("pg.guardian_id = ?", guardianID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) SearchPatients(ctx context.Context, queryStr string, limit int) ([]*db.BunPatients, error) {
	var patients []*db.BunPatients
	searchQuery := "%" + strings.ToLower(queryStr) + "%"
	err := r.db.NewSelect(ctx, &patients).
		Where("first_name ILIKE ? OR last_name ILIKE ?", searchQuery, searchQuery).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) CreatePatient(ctx context.Context, patient *db.BunPatients) error {
	_, err := r.db.NewInsert(patient).Exec(ctx)
	return err
}

func (r *PatientRepository) UpdatePatient(ctx context.Context, patient *db.BunPatients) error {
	_, err := r.db.NewUpdate(ctx, patient).Where("id = ?", patient.ID).Exec(ctx)
	return err
}

func (r *PatientRepository) DeletePatient(ctx context.Context, id string) error {
	var patient db.BunPatients
	_, err := r.db.NewDelete(ctx, &patient).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *PatientRepository) ListPatients(ctx context.Context, filter *model.PatientsFilter) ([]*db.BunPatients, error) {
	var patients []*db.BunPatients
	query := r.db.NewSelect(ctx, &patients)

	limit := 10
	offset := 0

	if filter != nil {
		for _, f := range filter.Filter {
			if f != nil && f.Search != nil && *f.Search != "" {
				searchQuery := "%" + *f.Search + "%"
				query = query.Where("first_name ILIKE ? OR last_name ILIKE ?", searchQuery, searchQuery)
			}
		}

		for _, p := range filter.Pagination {
			if p != nil {
				if p.Limit != nil && *p.Limit > 0 {
					limit = *p.Limit
				}
				if p.Offset != nil && *p.Offset >= 0 {
					offset = *p.Offset
				}
			}
		}
	}

	if err := query.Limit(limit).Offset(offset).Scan(ctx); err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) SearchPatientsForMessaging(ctx context.Context, queryStr string, limit int) ([]*db.BunPatients, error) {
	var patients []*db.BunPatients
	searchQuery := "%" + strings.ToLower(queryStr) + "%"
	err := r.tenantdb.NewSelect(ctx, &patients).
		Where("first_name ILIKE ? OR last_name ILIKE ?", searchQuery, searchQuery).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (r *PatientRepository) GetPatient(ctx context.Context, id string) (*db.BunPatients, error) {
	return r.FindPatientByID(ctx, id)
}

func (r *PatientRepository) GetPatientTags(ctx context.Context, patientID string) ([]*db.BunPatientTags, error) {
	var tags []*db.BunPatientTags
	err := r.db.NewSelect(ctx, &tags).Where("patient_id = ?", patientID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *PatientRepository) GetPatientGuardians(ctx context.Context, patientID string) ([]*db.BunGuardians, error) {
	var guardians []*db.BunGuardians
	err := r.db.NewSelect(ctx, &guardians).
		Join("JOIN patient_guardian_links pg ON pg.guardian_id = bun_guardians.id").
		Where("pg.patient_id = ? AND pg.is_active = true", patientID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return guardians, nil
}

func (r *PatientRepository) CreateGuardian(ctx context.Context, guardian *db.BunGuardians) error {
	_, err := r.db.NewInsert(guardian).Exec(ctx)
	return err
}

func (r *PatientRepository) UpdateGuardian(ctx context.Context, guardian *db.BunGuardians) error {
	_, err := r.db.NewUpdate(ctx, guardian).Where("id = ?", guardian.ID).Exec(ctx)
	return err
}

func (r *PatientRepository) AddPatientGuardian(ctx context.Context, pg *db.BunPatientGuardians) error {
	_, err := r.db.NewInsert(pg).Exec(ctx)
	return err
}

func (r *PatientRepository) UpdatePatientGuardian(ctx context.Context, patientID, guardianID, relationship string) error {
	_, err := r.db.NewUpdate(ctx, (*db.BunPatientGuardians)(nil)).
		Where("patient_id = ?", patientID).
		Where("guardian_id = ?", guardianID).
		Set("relationship = ?", relationship).
		Exec(ctx)
	return err
}

func (r *PatientRepository) RemovePatientGuardian(ctx context.Context, patientID, guardianID string) error {
	_, err := r.db.NewUpdate(ctx, (*db.BunPatientGuardians)(nil)).
		Where("patient_id = ?", patientID).
		Where("guardian_id = ?", guardianID).
		Set("is_active = ?", false).
		Exec(ctx)
	return err
}

func (r *PatientRepository) ReactivatePatientGuardian(ctx context.Context, patientID, guardianID string) error {
	_, err := r.db.NewUpdate(ctx, (*db.BunPatientGuardians)(nil)).
		Where("patient_id = ?", patientID).
		Where("guardian_id = ?", guardianID).
		Set("is_active = ?", true).
		Exec(ctx)
	return err
}
