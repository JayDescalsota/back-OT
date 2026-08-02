package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clinicmanager/services/patient/db"
	"github.com/clinicmanager/services/patient/graph/model"
	"github.com/clinicmanager/services/patient/repository"
	"github.com/clinicmanager/shared/cache"
	sharedCtx "github.com/clinicmanager/shared/context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PatientService struct {
	PatientRepository *repository.PatientRepository
	Cache             *redis.Client
}

func NewPatientService(patientRepository *repository.PatientRepository, cacheClient *redis.Client) *PatientService {
	return &PatientService{PatientRepository: patientRepository, Cache: cacheClient}
}

func (s *PatientService) cacheGet(ctx context.Context, key string, dest interface{}) (bool, error) {
	if s.Cache == nil {
		return false, nil
	}
	val, err := s.Cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *PatientService) cacheSet(ctx context.Context, key string, val interface{}) error {
	if s.Cache == nil {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.Cache.Set(ctx, key, data, cache.DefaultTTL).Err()
}

func (s *PatientService) cacheDel(ctx context.Context, keys ...string) {
	if s.Cache == nil {
		return
	}
	s.Cache.Del(ctx, keys...)
}

func (s *PatientService) GetPatientByID(ctx context.Context, id string) (*db.BunPatients, error) {
	ck := cache.Key("patient", "patient", id)
	var cached db.BunPatients
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.PatientRepository.FindPatientByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *PatientService) GetPatientsByIDs(ctx context.Context, ids []string) ([]*db.BunPatients, error) {
	return s.PatientRepository.GetPatientsByIDs(ctx, ids)
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
		AddressID:   input.AddressID,
	}

	if err := s.PatientRepository.CreatePatient(ctx, patient); err != nil {
		return nil, err
	}

	s.cacheDel(ctx, cache.Key("patient", "patient", patient.ID))
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

	s.cacheDel(ctx, cache.Key("patient", "patient", id))
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

	s.cacheDel(ctx, cache.Key("patient", "patient", id))
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

	s.cacheDel(ctx, cache.Key("patient", "patient", id))
	return patient, nil
}

func (s *PatientService) ListPatients(ctx context.Context, filter *model.PatientsFilter) ([]*db.BunPatients, error) {
	return s.PatientRepository.ListPatients(ctx, filter)
}

func (s *PatientService) SearchPatientsForMessaging(ctx context.Context, query string, limit int) ([]*db.BunPatients, error) {
	return s.PatientRepository.SearchPatientsForMessaging(ctx, query, limit)
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

	s.cacheDel(ctx, cache.Key("patient", "guardian", guardian.ID))
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

	s.cacheDel(ctx, cache.Key("patient", "guardian", id))
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

	s.cacheDel(ctx, cache.Key("patient", "guardian", id))
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

	s.cacheDel(ctx,
		cache.Key("patient", "guardian", guardianID),
		cache.Key("patient", "patient", patientID),
	)
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
	ck := cache.Key("patient", "guardian", id)
	var cached db.BunGuardians
	if ok, _ := s.cacheGet(ctx, ck, &cached); ok {
		return &cached, nil
	}
	m, err := s.PatientRepository.FindGuardianByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	s.cacheSet(ctx, ck, m)
	return m, nil
}

func (s *PatientService) GetPatientGoals(ctx context.Context, patientID string) ([]*db.BunGoal, error) {
	return s.PatientRepository.GetPatientGoals(ctx, patientID)
}

func (s *PatientService) CreateGoal(ctx context.Context, input model.GoalInput) (*db.BunGoal, error) {
	progress := 0
	if input.Progress != nil {
		progress = *input.Progress
	}
	status := "Not Started"
	if input.Status != nil && *input.Status != "" {
		status = *input.Status
	}
	goal := &db.BunGoal{
		ID:        uuid.NewString(),
		PatientID: input.PatientID,
		Goal:      input.Goal,
		Target:    input.Target,
		Progress:  progress,
		Status:    status,
		IsActive:  true,
	}

	if err := s.PatientRepository.CreateGoal(ctx, goal); err != nil {
		return nil, err
	}

	s.cacheDel(ctx, cache.Key("patient", "patient", input.PatientID))
	return goal, nil
}

func (s *PatientService) UpdateGoal(ctx context.Context, id string, input model.GoalUpdateInput) (*db.BunGoal, error) {
	existing, err := s.PatientRepository.FindGoalByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if input.Goal != nil {
		existing.Goal = *input.Goal
	}
	if input.Target != nil {
		existing.Target = input.Target
	}
	if input.Progress != nil {
		existing.Progress = *input.Progress
	}
	if input.Status != nil && *input.Status != "" {
		existing.Status = *input.Status
	}

	if err := s.PatientRepository.UpdateGoal(ctx, existing); err != nil {
		return nil, err
	}

	s.cacheDel(ctx, cache.Key("patient", "patient", existing.PatientID))
	return existing, nil
}

func (s *PatientService) DeleteGoal(ctx context.Context, id string) (*db.BunGoal, error) {
	existing, err := s.PatientRepository.FindGoalByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if err := s.PatientRepository.DeleteGoal(ctx, id); err != nil {
		return nil, err
	}

	s.cacheDel(ctx, cache.Key("patient", "patient", existing.PatientID))
	return existing, nil
}
