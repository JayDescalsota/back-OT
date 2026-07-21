package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clinicmanager/services/booking/graph"
	"github.com/clinicmanager/services/booking/graph/generated"
	graphmodel "github.com/clinicmanager/services/booking/graph/model"
	"github.com/clinicmanager/services/booking/repository"
	"github.com/clinicmanager/services/booking/service"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
)

func getTestDBURL() string {
	url := os.Getenv("BOOKINGDB_URL")
	if url == "" {
		url = "postgres://postgres:1@localhost:5432/booking?sslmode=disable"
	}
	return url
}

func setupIntegrationTest(t *testing.T) (context.Context, *service.BookingService) {
	t.Helper()

	dbURL := getTestDBURL()
	db, err := sharedDB.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to real PostgreSQL database at %s: %v", dbURL, err)
	}
	t.Cleanup(func() { db.Close() })

	dbSet := sharedDB.NewDBSet(db)

	repo := repository.NewBookingRepository(dbSet)
	svc := service.NewBookingService(repo, nil)

	ctx := context.Background()
	ctx = sharedCtx.SetTenantID(ctx, uuid.NewString())
	ctx = sharedCtx.SetBranchID(ctx, uuid.NewString())
	ctx = sharedCtx.SetUserID(ctx, uuid.NewString())

	return ctx, svc
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	dbURL := getTestDBURL()
	db, err := sharedDB.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to real PostgreSQL database at %s: %v", dbURL, err)
	}
	t.Cleanup(func() { db.Close() })

	dbSet := sharedDB.NewDBSet(db)

	repo := repository.NewBookingRepository(dbSet)
	svc := service.NewBookingService(repo, nil)

	resolver := graph.NewResolver(svc)
	executableSchema := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	gqlServer := handler.NewDefaultServer(executableSchema)
	contextedHandler := sharedCtx.Tenant(gqlServer)

	return httptest.NewServer(contextedHandler)
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func TestBookingService_Integration_CreateBranchHours(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	branchHours, err := svc.CreateBranchHours(ctx, graphmodel.BranchHoursInput{
		BranchID:  testBranchID,
		DayOfWeek: 1,
		OpenTime:  &[]string{"09:00"}[0],
		CloseTime: &[]string{"17:00"}[0],
	})
	require.NoError(t, err)
	require.NotNil(t, branchHours)
	assert.NotEmpty(t, branchHours.ID)
	assert.Equal(t, testBranchID, branchHours.BranchID)
	assert.Equal(t, "09:00", branchHours.OpenTime.Format("15:04"))
	assert.Equal(t, "17:00", branchHours.CloseTime.Format("15:04"))
	assert.True(t, branchHours.IsActive)

	// Verify branch hours exists in DB
	fetched, err := svc.GetBranchHoursByID(ctx, branchHours.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, branchHours.ID, fetched.ID)
}

func TestBookingService_Integration_UpdateBranchHours(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	// Create branch hours first
	branchHours, err := svc.CreateBranchHours(ctx, graphmodel.BranchHoursInput{
		BranchID:  testBranchID,
		DayOfWeek: 1,
		OpenTime:  &[]string{"09:00"}[0],
		CloseTime: &[]string{"17:00"}[0],
	})
	require.NoError(t, err)
	require.NotNil(t, branchHours)

	// Update branch hours
	newDayOfWeek := 2
	newOpenTime := "10:00"
	newCloseTime := "18:00"
	updateInput := graphmodel.BranchHoursUpdateInput{
		DayOfWeek: &newDayOfWeek,
		OpenTime:  &newOpenTime,
		CloseTime: &newCloseTime,
	}

	updated, err := svc.UpdateBranchHours(ctx, branchHours.ID, updateInput)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "10:00", updated.OpenTime.Format("15:04"))
	assert.Equal(t, "18:00", updated.CloseTime.Format("15:04"))
}

func TestBookingService_Integration_DeleteBranchHours(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	// Create branch hours first
	branchHours, err := svc.CreateBranchHours(ctx, graphmodel.BranchHoursInput{
		BranchID:  testBranchID,
		DayOfWeek: 1,
		OpenTime:  &[]string{"09:00"}[0],
		CloseTime: &[]string{"17:00"}[0],
	})
	require.NoError(t, err)
	require.NotNil(t, branchHours)
	assert.True(t, branchHours.IsActive)

	// Delete (inactivate) branch hours
	deleted, err := svc.DeleteBranchHours(ctx, branchHours.ID)
	require.NoError(t, err)
	require.True(t, deleted)

	// Verify branch hours is inactive in DB
	fetched, err := svc.GetBranchHoursByID(ctx, branchHours.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.IsActive)
}

func TestBookingService_Integration_CreatePractitionerBranch(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testPractitionerID := uuid.NewString()
	testBranchID := uuid.NewString()
	practitionerBranch, err := svc.CreatePractitionerBranch(ctx, graphmodel.PractitionerBranchInput{
		PractitionerID: testPractitionerID,
		BranchID:       testBranchID,
	})
	require.NoError(t, err)
	require.NotNil(t, practitionerBranch)
	assert.NotEmpty(t, practitionerBranch.ID)
	assert.Equal(t, testPractitionerID, practitionerBranch.PractitionerID)
	assert.Equal(t, testBranchID, practitionerBranch.BranchID)
	assert.True(t, practitionerBranch.IsActive)

	// Verify practitioner branch exists in DB
	fetched, err := svc.GetPractitionerBranchByID(ctx, practitionerBranch.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, practitionerBranch.ID, fetched.ID)
}

func TestBookingService_Integration_DeletePractitionerBranch(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testPractitionerID := uuid.NewString()
	testBranchID := uuid.NewString()
	// Create practitioner branch first
	practitionerBranch, err := svc.CreatePractitionerBranch(ctx, graphmodel.PractitionerBranchInput{
		PractitionerID: testPractitionerID,
		BranchID:       testBranchID,
	})
	require.NoError(t, err)
	require.NotNil(t, practitionerBranch)
	assert.True(t, practitionerBranch.IsActive)

	// Delete (inactivate) practitioner branch
	deleted, err := svc.DeletePractitionerBranch(ctx, practitionerBranch.ID)
	require.NoError(t, err)
	require.True(t, deleted)

	// Verify practitioner branch is inactive in DB
	fetched, err := svc.GetPractitionerBranchByID(ctx, practitionerBranch.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.IsActive)
}

func TestBookingService_Integration_CreatePractitionerAvailability(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testPractitionerID := uuid.NewString()
	now := time.Now().UTC()
	startTime := now.Format(time.RFC3339)
	endTime := now.Add(8 * time.Hour).Format(time.RFC3339)

	practitionerAvailability, err := svc.CreatePractitionerAvailability(ctx, graphmodel.PractitionerAvailabilityInput{
		PractitionerID: testPractitionerID,
		StartTime:      startTime,
		EndTime:        endTime,
	})
	require.NoError(t, err)
	require.NotNil(t, practitionerAvailability)
	assert.NotEmpty(t, practitionerAvailability.ID)
	assert.Equal(t, testPractitionerID, practitionerAvailability.PractitionerID)
	assert.Equal(t, "ACTIVE", practitionerAvailability.Status)
	assert.True(t, practitionerAvailability.IsActive)

	// Verify practitioner availability exists in DB
	fetched, err := svc.GetPractitionerAvailabilityByID(ctx, practitionerAvailability.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, practitionerAvailability.ID, fetched.ID)
}

func TestBookingService_Integration_DeletePractitionerAvailability(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testPractitionerID := uuid.NewString()
	// Create practitioner availability first
	now := time.Now().UTC()
	practitionerAvailability, err := svc.CreatePractitionerAvailability(ctx, graphmodel.PractitionerAvailabilityInput{
		PractitionerID: testPractitionerID,
		StartTime:      now.Format(time.RFC3339),
		EndTime:        now.Add(8 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, practitionerAvailability)
	assert.True(t, practitionerAvailability.IsActive)

	// Delete (inactivate) practitioner availability
	deleted, err := svc.DeletePractitionerAvailability(ctx, practitionerAvailability.ID)
	require.NoError(t, err)
	require.True(t, deleted)

	// Verify practitioner availability is inactive in DB
	fetched, err := svc.GetPractitionerAvailabilityByID(ctx, practitionerAvailability.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.IsActive)
}

func TestBookingService_Integration_CreateAppointment(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPatientID := uuid.NewString()
	testPractitionerID := uuid.NewString()
	now := time.Now().UTC()
	scheduledStart := now.Add(24 * time.Hour).Format(time.RFC3339)
	scheduledEnd := now.Add(25 * time.Hour).Format(time.RFC3339)

	appointment, err := svc.CreateAppointment(ctx, graphmodel.AppointmentInput{
		BranchID:       testBranchID,
		SlotID:         &[]string{uuid.NewString()}[0],
		PatientID:      testPatientID,
		PractitionerID: testPractitionerID,
		ScheduledStart: scheduledStart,
		ScheduledEnd:   scheduledEnd,
	})
	require.NoError(t, err)
	require.NotNil(t, appointment)
	assert.NotEmpty(t, appointment.ID)
	assert.Equal(t, testBranchID, appointment.BranchID)
	require.NotNil(t, appointment.PatientID)
	assert.Equal(t, testPatientID, *appointment.PatientID)
	require.NotNil(t, appointment.PractitionerID)
	assert.Equal(t, testPractitionerID, *appointment.PractitionerID)
	assert.Equal(t, "PENDING", appointment.Status)

	// Verify appointment exists in DB
	fetched, err := svc.GetAppointmentByID(ctx, appointment.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, appointment.ID, fetched.ID)
}

func TestBookingService_Integration_UpdateAppointment(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPatientID := uuid.NewString()
	testPractitionerID := uuid.NewString()
	// Create appointment first
	now := time.Now().UTC()
	appointment, err := svc.CreateAppointment(ctx, graphmodel.AppointmentInput{
		BranchID:       testBranchID,
		SlotID:         &[]string{uuid.NewString()}[0],
		PatientID:      testPatientID,
		PractitionerID: testPractitionerID,
		ScheduledStart: now.Add(24 * time.Hour).Format(time.RFC3339),
		ScheduledEnd:   now.Add(25 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, appointment)

	// Update appointment
	updateInput := graphmodel.AppointmentUpdateInput{
		Status: func() *string { s := "CONFIRMED"; return &s }(),
		Note:   func() *string { n := "Test appointment note"; return &n }(),
	}

	updated, err := svc.UpdateAppointment(ctx, appointment.ID, updateInput)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "CONFIRMED", updated.Status)
	assert.Equal(t, "Test appointment note", updated.Notes)
}

func TestBookingService_Integration_CancelAppointment(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	// Create appointment first
	now := time.Now().UTC()
	appointment, err := svc.CreateAppointment(ctx, graphmodel.AppointmentInput{
		BranchID:       uuid.NewString(),
		SlotID:         &[]string{uuid.NewString()}[0],
		PatientID:      uuid.NewString(),
		PractitionerID: uuid.NewString(),
		ScheduledStart: now.Add(24 * time.Hour).Format(time.RFC3339),
		ScheduledEnd:   now.Add(25 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, appointment)
	assert.Equal(t, "PENDING", appointment.Status)

	// Cancel appointment
	cancelled, err := svc.CancelAppointment(ctx, appointment.ID)
	require.NoError(t, err)
	require.NotNil(t, cancelled)
	assert.Equal(t, "CANCELLED", cancelled.Status)

	// Verify appointment is cancelled in DB
	fetched, err := svc.GetAppointmentByID(ctx, appointment.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "CANCELLED", fetched.Status)
}

func TestBookingService_Integration_CreateScheduleTemplate(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	scheduleTemplate, err := svc.CreateScheduleTemplate(ctx, graphmodel.ScheduleTemplateInput{
		BranchID:            testBranchID,
		PractitionerID:      testPractitionerID,
		DayOfWeek:           1,
		StartTime:           "09:00",
		EndTime:             "17:00",
		SlotDurationMinutes: 30,
	})
	require.NoError(t, err)
	require.NotNil(t, scheduleTemplate)
	assert.NotEmpty(t, scheduleTemplate.ID)
	assert.Equal(t, testBranchID, scheduleTemplate.BranchID)
	require.NotNil(t, scheduleTemplate.PractitionerID)
	assert.Equal(t, testPractitionerID, *scheduleTemplate.PractitionerID)
	assert.Equal(t, "09:00", scheduleTemplate.StartTime.Format("15:04"))
	assert.Equal(t, "17:00", scheduleTemplate.EndTime.Format("15:04"))
	assert.Equal(t, 30, scheduleTemplate.SlotDuration)
	assert.Equal(t, "ACTIVE", scheduleTemplate.Status)
	assert.True(t, scheduleTemplate.IsActive)

	// Verify schedule template exists in DB
	fetched, err := svc.GetScheduleTemplateByID(ctx, scheduleTemplate.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, scheduleTemplate.ID, fetched.ID)
}

func TestBookingService_Integration_UpdateScheduleTemplate(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	// Create schedule template first
	scheduleTemplate, err := svc.CreateScheduleTemplate(ctx, graphmodel.ScheduleTemplateInput{
		BranchID:            testBranchID,
		PractitionerID:      testPractitionerID,
		DayOfWeek:           1,
		StartTime:           "09:00",
		EndTime:             "17:00",
		SlotDurationMinutes: 30,
	})
	require.NoError(t, err)
	require.NotNil(t, scheduleTemplate)

	// Update schedule template
	newDayOfWeek := 2
	newStartTime := "10:00"
	newEndTime := "18:00"
	newSlotDuration := 45
	updateInput := graphmodel.ScheduleTemplateUpdateInput{
		DayOfWeek:           &newDayOfWeek,
		StartTime:           &newStartTime,
		EndTime:             &newEndTime,
		SlotDurationMinutes: &newSlotDuration,
	}

	updated, err := svc.UpdateScheduleTemplate(ctx, scheduleTemplate.ID, updateInput)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "10:00", updated.StartTime.Format("15:04"))
	assert.Equal(t, "18:00", updated.EndTime.Format("15:04"))
	assert.Equal(t, 45, updated.SlotDuration)
}

func TestBookingService_Integration_DeleteScheduleTemplate(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	// Create schedule template first
	scheduleTemplate, err := svc.CreateScheduleTemplate(ctx, graphmodel.ScheduleTemplateInput{
		BranchID:            testBranchID,
		PractitionerID:      testPractitionerID,
		DayOfWeek:           1,
		StartTime:           "09:00",
		EndTime:             "17:00",
		SlotDurationMinutes: 30,
	})
	require.NoError(t, err)
	require.NotNil(t, scheduleTemplate)
	assert.True(t, scheduleTemplate.IsActive)

	// Delete (inactivate) schedule template
	deleted, err := svc.DeleteScheduleTemplate(ctx, scheduleTemplate.ID)
	require.NoError(t, err)
	require.True(t, deleted)

	// Verify schedule template is inactive in DB
	fetched, err := svc.GetScheduleTemplateByID(ctx, scheduleTemplate.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.IsActive)
}

func TestBookingService_Integration_CreateScheduleException(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	now := time.Now().UTC()
	exceptionDate := now.Add(48 * time.Hour).Format(time.RFC3339)

	scheduleException, err := svc.CreateScheduleException(ctx, graphmodel.ScheduleExceptionInput{
		BranchID:       testBranchID,
		PractitionerID: testPractitionerID,
		ExceptionDate:  exceptionDate,
	})
	require.NoError(t, err)
	require.NotNil(t, scheduleException)
	assert.NotEmpty(t, scheduleException.ID)
	require.NotNil(t, scheduleException.PractitionerID)
	assert.Equal(t, testPractitionerID, *scheduleException.PractitionerID)
	assert.Equal(t, "ACTIVE", scheduleException.Status)
	assert.True(t, scheduleException.IsActive)

	// Verify schedule exception exists in DB
	fetched, err := svc.GetScheduleExceptionByID(ctx, scheduleException.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, scheduleException.ID, fetched.ID)
}

func TestBookingService_Integration_DeleteScheduleException(t *testing.T) {
	ctx, svc := setupIntegrationTest(t)

	testBranchID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	// Create schedule exception first
	now := time.Now().UTC()
	scheduleException, err := svc.CreateScheduleException(ctx, graphmodel.ScheduleExceptionInput{
		BranchID:       testBranchID,
		PractitionerID: testPractitionerID,
		ExceptionDate:  now.Add(48 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, scheduleException)
	assert.True(t, scheduleException.IsActive)

	// Delete (inactivate) schedule exception
	deleted, err := svc.DeleteScheduleException(ctx, scheduleException.ID)
	require.NoError(t, err)
	require.True(t, deleted)

	// Verify schedule exception is inactive in DB
	fetched, err := svc.GetScheduleExceptionByID(ctx, scheduleException.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.IsActive)
}

func TestBookingService_Integration_AppointmentSlotsQuery(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Create appointment slots via service layer first
	ctx, svc := setupIntegrationTest(t)

	now := time.Now().UTC()
	scheduledStart := now.Add(24 * time.Hour).Format(time.RFC3339)
	scheduledEnd := now.Add(25 * time.Hour).Format(time.RFC3339)

	testBranchID := uuid.NewString()
	testPatientID := uuid.NewString()
	testPractitionerID := uuid.NewString()

	_, err := svc.CreateAppointment(ctx, graphmodel.AppointmentInput{
		BranchID:       testBranchID,
		SlotID:         &[]string{uuid.NewString()}[0],
		PatientID:      testPatientID,
		PractitionerID: testPractitionerID,
		ScheduledStart: scheduledStart,
		ScheduledEnd:   scheduledEnd,
	})
	require.NoError(t, err)

	// Query appointments via GraphQL
	query := map[string]string{
		"query": `
			query {
				appointments(filter: {branch_id: "` + testBranchID + `"}) {
					id
					branch_id
					patient_id
					practitioner_id
					scheduled_start
					scheduled_end
					status
				}
			}
		`,
	}

	bodyBytes, err := json.Marshal(query)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", uuid.NewString())
	req.Header.Set("x-branch-id", uuid.NewString())

	client := server.Client()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gqlResp GraphQLResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&gqlResp))
	assert.Empty(t, gqlResp.Errors)
}

func TestBookingService_Integration_BranchHoursQuery(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	testBranchID := uuid.NewString()

	// Create branch hours via service layer first
	ctx, svc := setupIntegrationTest(t)

	_, err := svc.CreateBranchHours(ctx, graphmodel.BranchHoursInput{
		BranchID:  testBranchID,
		DayOfWeek: 1,
		OpenTime:  &[]string{"09:00"}[0],
		CloseTime: &[]string{"17:00"}[0],
	})
	require.NoError(t, err)

	// Query branch hours via GraphQL
	query := map[string]string{
		"query": `
			query {
				branchHours {
					id
					branch_id
					day_of_week
					open_time
					close_time
					is_active
				}
			}
		`,
	}

	bodyBytes, err := json.Marshal(query)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", uuid.NewString())
	req.Header.Set("x-branch-id", testBranchID)

	client := server.Client()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gqlResp GraphQLResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&gqlResp))
	assert.Empty(t, gqlResp.Errors)
}

func TestBookingService_Integration_PractitionerBranchesQuery(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	testPractitionerID := uuid.NewString()
	testBranchID := uuid.NewString()

	// Create practitioner branch via service layer first
	ctx, svc := setupIntegrationTest(t)

	_, err := svc.CreatePractitionerBranch(ctx, graphmodel.PractitionerBranchInput{
		PractitionerID: testPractitionerID,
		BranchID:       testBranchID,
	})
	require.NoError(t, err)

	// Query practitioner branches via GraphQL
	query := map[string]string{
		"query": `
			query {
				practitionerBranches {
					id
					practitioner_id
					branch_id
					is_active
				}
			}
		`,
	}

	bodyBytes, err := json.Marshal(query)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", uuid.NewString())
	req.Header.Set("x-branch-id", testBranchID)

	client := server.Client()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gqlResp GraphQLResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&gqlResp))
	assert.Empty(t, gqlResp.Errors)
}

func TestBookingService_Integration_PractitionerAvailabilitiesQuery(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	testPractitionerID := uuid.NewString()

	// Create practitioner availability via service layer first
	ctx, svc := setupIntegrationTest(t)

	now := time.Now().UTC()
	_, err := svc.CreatePractitionerAvailability(ctx, graphmodel.PractitionerAvailabilityInput{
		PractitionerID: testPractitionerID,
		StartTime:      now.Format(time.RFC3339),
		EndTime:        now.Add(8 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)

	// Query practitioner availabilities via GraphQL
	query := map[string]string{
		"query": `
			query {
				practitionerAvailabilities {
					id
					practitioner_id
					start_time
					end_time
					is_active
				}
			}
		`,
	}

	bodyBytes, err := json.Marshal(query)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", uuid.NewString())
	req.Header.Set("x-branch-id", testPractitionerID)

	client := server.Client()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gqlResp GraphQLResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&gqlResp))
	assert.Empty(t, gqlResp.Errors)
}
