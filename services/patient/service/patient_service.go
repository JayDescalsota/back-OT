package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	sharedCtx "github.com/clinicmanager/shared/context"
	"github.com/clinicmanager/services/patient/db"
	"github.com/clinicmanager/services/patient/graph/model"
	"github.com/clinicmanager/services/patient/repository"
)

type PatientService struct {
	PatientRepository *repository.PatientRepository
}

func NewPatientService(patientRepository *repository.PatientRepository) *PatientService {
	return &PatientService{PatientRepository: patientRepository}
}

func (s *PatientService) GetPatientByID(ctx context.Context, id string) (*db.BunPatients, error) {
	return s.PatientRepository.FindPatientByID(ctx, id)
}

func (s *PatientService) CreatePatient(ctx context.Context, input model.PatientInput) (*db.BunPatients, error) {
	dob, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		return nil, err
	}

	tctx := sharedCtx.FromContext(ctx)

	patient := &db.BunPatients{
		ID:          uuid.NewString(),
		TenantID:    tctx.TenantID,
		BranchID:    tctx.BranchID,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		DateOfBirth: dob,
		Gender:      input.Gender,
		Notes:       input.Notes,
		Height:      input.Height,
		Weight:      input.Weight,
		IsActive:    true,
	}

	if err := s.PatientRepository.CreatePatient(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}

func (s *PatientService) UpdatePatient(ctx context.Context, id string, input model.PatientUpdateInput) (*db.BunPatients, error) {
	existing, err := s.PatientRepository.GetPatient(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if input.FirstName != nil {
		existing.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		existing.LastName = *input.LastName
	}
	if input.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *input.DateOfBirth)
		if err != nil {
			return nil, err
		}
		existing.DateOfBirth = dob
	}
	if input.Gender != nil {
		existing.Gender = *input.Gender
	}
	if input.Notes != nil {
		existing.Notes = input.Notes
	}
	if input.Height != nil {
		existing.Height = input.Height
	}
	if input.Weight != nil {
		existing.Weight = input.Weight
	}

	if err := s.PatientRepository.UpdatePatient(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *PatientService) DeletePatient(ctx context.Context, id string) (*db.BunPatients, error) {
	return s.InactivatePatient(ctx, id)
}

func (s *PatientService) InactivatePatient(ctx context.Context, id string) (*db.BunPatients, error) {
	patient, err := s.PatientRepository.GetPatient(ctx, id)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, nil
	}

	patient.IsActive = false
	if err := s.PatientRepository.UpdatePatient(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}

func (s *PatientService) ReactivatePatient(ctx context.Context, id string) (*db.BunPatients, error) {
	patient, err := s.PatientRepository.GetPatient(ctx, id)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, nil
	}

	patient.IsActive = true
	if err := s.PatientRepository.UpdatePatient(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}

func (s *PatientService) ListPatients(ctx context.Context, filter *model.PatientsFilter) ([]*db.BunPatients, error) {
	return s.PatientRepository.ListPatients(ctx, filter)
}

func (s *PatientService) GetPatientAddress(ctx context.Context, patientID string) (*db.BunPatientAddress, error) {
	return s.PatientRepository.GetPatientAddress(ctx, patientID)
}

func (s *PatientService) GetPatientTags(ctx context.Context, patientID string) ([]*db.BunPatientTags, error) {
	return s.PatientRepository.GetPatientTags(ctx, patientID)
}

func (s *PatientService) GetPatientGuardians(ctx context.Context, patientID string) ([]*db.BunGuardians, error) {
	return s.PatientRepository.GetPatientGuardians(ctx, patientID)
}

func (s *PatientService) GetGuardianPatients(ctx context.Context, guardianID string) ([]*db.BunPatients, error) {
	return s.PatientRepository.FindGuardianPatients(ctx, guardianID)
}

func (s *PatientService) CreateGuardian(ctx context.Context, input model.GuardianInput) (*db.BunGuardians, error) {
	phone := ""
	if input.Phone != nil {
		phone = *input.Phone
	}
	guardian := &db.BunGuardians{
		ID:        uuid.NewString(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Gender:    input.Gender,
		Email:     input.Email,
		Phone:     phone,
		Notes:     input.Notes,
		IsActive:  true,
	}

	if err := s.PatientRepository.CreateGuardian(ctx, guardian); err != nil {
		return nil, err
	}

	return guardian, nil
}

func (s *PatientService) UpdateGuardian(ctx context.Context, id string, input model.GuardianUpdateInput) (*db.BunGuardians, error) {
	existing, err := s.PatientRepository.FindGuardianByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if input.FirstName != nil {
		existing.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		existing.LastName = *input.LastName
	}
	if input.Gender != nil {
		existing.Gender = *input.Gender
	}
	if input.Email != nil {
		existing.Email = input.Email
	}
	if input.Phone != nil {
		existing.Phone = *input.Phone
	}
	if input.Notes != nil {
		existing.Notes = input.Notes
	}

	if err := s.PatientRepository.UpdateGuardian(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *PatientService) DeleteGuardian(ctx context.Context, id string) (*db.BunGuardians, error) {
	guardian, err := s.PatientRepository.FindGuardianByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if guardian == nil {
		return nil, nil
	}

	guardian.IsActive = false
	if err := s.PatientRepository.UpdateGuardian(ctx, guardian); err != nil {
		return nil, err
	}

	return guardian, nil
}

func (s *PatientService) AddPatientGuardian(ctx context.Context, patientID, guardianID, relationship string) (*model.PatientGuardianRelationship, error) {
	pg := &db.BunPatientGuardians{
		PatientID:    patientID,
		GuardianID:   guardianID,
		Relationship: relationship,
		IsActive:     true,
	}

	if err := s.PatientRepository.AddPatientGuardian(ctx, pg); err != nil {
		return nil, err
	}

	return &model.PatientGuardianRelationship{
		PatientID:    patientID,
		GuardianID:   guardianID,
		Relationship: relationship,
	}, nil
}

func (s *PatientService) UpdatePatientGuardian(ctx context.Context, patientID, guardianID, relationship string) (*model.PatientGuardianRelationship, error) {
	if err := s.PatientRepository.UpdatePatientGuardian(ctx, patientID, guardianID, relationship); err != nil {
		return nil, err
	}

	return &model.PatientGuardianRelationship{
		PatientID:    patientID,
		GuardianID:   guardianID,
		Relationship: relationship,
	}, nil
}

func (s *PatientService) RemovePatientGuardian(ctx context.Context, patientID, guardianID string) (bool, error) {
	if err := s.PatientRepository.RemovePatientGuardian(ctx, patientID, guardianID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *PatientService) ReactivatePatientGuardian(ctx context.Context, patientID, guardianID string) (bool, error) {
	if err := s.PatientRepository.ReactivatePatientGuardian(ctx, patientID, guardianID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *PatientService) GetGuardianByID(ctx context.Context, id string) (*db.BunGuardians, error) {
	return s.PatientRepository.FindGuardianByID(ctx, id)
}
