package staff

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func attachAllowAllClinicPermission(c *gin.Context) {
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

type scopedHandlerService struct {
	Service
	verifyFn func(context.Context, uint64, uint64) error
	getFn    func(context.Context, uint64, uint64) (*model.Staff, error)
}

func (s *scopedHandlerService) VerifyClinicMembership(
	ctx context.Context,
	staffID, clinicID uint64,
) error {
	return s.verifyFn(ctx, staffID, clinicID)
}

func (s *scopedHandlerService) GetByIDInClinic(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	return s.getFn(ctx, clinicID, staffID)
}

type scopedHandlerAssignmentService struct {
	StaffClinicAssignmentService
	findAllFn func(context.Context, uint64) ([]model.StaffClinicAssignment, error)
}

func (s *scopedHandlerAssignmentService) FindAllByStaffID(
	ctx context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return s.findAllFn(ctx, staffID)
}

func newScopedStaffHandlerRouter(handler *Handler, route string, endpoint gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(route, func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Next()
	}, endpoint)
	return router
}

func TestHandler_GetStaffUsesClinicScopedRepresentation(t *testing.T) {
	service := &scopedHandlerService{
		verifyFn: func(_ context.Context, staffID, clinicID uint64) error {
			assert.Equal(t, uint64(7), staffID)
			assert.Equal(t, uint64(20), clinicID)
			return nil
		},
		getFn: func(_ context.Context, clinicID, staffID uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(20), clinicID)
			return &model.Staff{
				ID:           staffID,
				Name:         "多施設所属スタッフ",
				OccupationID: nil,
				Occupation:   nil,
			}, nil
		},
	}
	handler := NewHandler(service, nil, nil, nil, nil, nil)
	router := newScopedStaffHandlerRouter(handler, "/staffs/:id", handler.GetStaff)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/staffs/7", http.NoBody))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.NotContains(t, body, "occupation_id")
	assert.NotContains(t, body, "occupation")
}

func TestHandler_GetStaffClinicAssignmentsIntersectsSystemAdminActiveClinics(t *testing.T) {
	staffService := &scopedHandlerService{
		verifyFn: func(_ context.Context, _, _ uint64) error { return nil },
		getFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, nil
		},
	}
	assignmentService := &scopedHandlerAssignmentService{
		findAllFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			assert.Equal(t, uint64(7), staffID)
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: 30},
				{StaffID: staffID, ClinicID: 20},
				{StaffID: staffID, ClinicID: 20},
				{
					StaffID:  staffID,
					ClinicID: 50,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
					},
				},
				{StaffID: 99, ClinicID: 60},
				{StaffID: staffID, ClinicID: 0},
			}, nil
		},
	}
	handler := NewHandler(staffService, assignmentService, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/staffs/:id/clinics", func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Set("is_system_admin", true)
		c.Set("clinic_ids", []uint64{20})
		c.Next()
	}, handler.GetStaffClinicAssignments)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/staffs/7/clinics", http.NoBody),
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body staffClinicAssignmentsResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	// BUG-010: system-admin claims are active clinics only. GET must still
	// return assignments whose clinic is missing from claims (inactive),
	// otherwise a later PUT of the GET list deletes those rows.
	assert.Equal(t, []uint64{20, 30}, body.ClinicIDs)
}

func TestHandler_GetStaffClinicAssignmentsIntersectsNonAdminAuthorizedClinics(t *testing.T) {
	staffService := &scopedHandlerService{
		verifyFn: func(_ context.Context, staffID, clinicID uint64) error {
			assert.Equal(t, uint64(7), staffID)
			assert.Equal(t, uint64(20), clinicID)
			return nil
		},
		getFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, nil
		},
	}
	assignments := []model.StaffClinicAssignment{
		{StaffID: 7, ClinicID: 40},
		{StaffID: 7, ClinicID: 30},
		{StaffID: 7, ClinicID: 20},
	}
	assignmentService := &scopedHandlerAssignmentService{
		findAllFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			assert.Equal(t, uint64(7), staffID)
			return assignments, nil
		},
	}
	handler := NewHandler(staffService, assignmentService, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/staffs/:id/clinics", func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Set("is_system_admin", false)
		c.Set("clinic_ids", []uint64{20, 40})
		attachAllowAllClinicPermission(c)
		c.Next()
	}, handler.GetStaffClinicAssignments)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/staffs/7/clinics", http.NoBody),
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body staffClinicAssignmentsResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, []uint64{20, 40}, body.ClinicIDs)
	assert.NotContains(t, body.ClinicIDs, uint64(30), "non-admin GET must hide clinic IDs outside authorizedClinicIDs")
	assert.Equal(t, []model.StaffClinicAssignment{
		{StaffID: 7, ClinicID: 40},
		{StaffID: 7, ClinicID: 30},
		{StaffID: 7, ClinicID: 20},
	}, assignments)
}

func TestHandler_GetStaffClinicAssignmentsRejectsMissingAssignmentService(t *testing.T) {
	staffService := &scopedHandlerService{
		verifyFn: func(_ context.Context, _, _ uint64) error { return nil },
		getFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, nil
		},
	}
	handler := NewHandler(staffService, nil, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/staffs/:id/clinics", func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Set("is_system_admin", true)
		c.Set("clinic_ids", []uint64{20})
		c.Next()
	}, handler.GetStaffClinicAssignments)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/staffs/7/clinics", http.NoBody),
	)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandler_GetStaffClinicAssignmentsRejectsMissingSystemAdminClinicAuthority(t *testing.T) {
	staffService := &scopedHandlerService{
		verifyFn: func(_ context.Context, _, _ uint64) error { return nil },
		getFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, nil
		},
	}
	assignmentService := &scopedHandlerAssignmentService{
		findAllFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
			t.Fatal("assignment lookup must not run without trusted clinic authority")
			return nil, nil
		},
	}
	handler := NewHandler(staffService, assignmentService, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/staffs/:id/clinics", func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Set("is_system_admin", true)
		c.Next()
	}, handler.GetStaffClinicAssignments)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/staffs/7/clinics", http.NoBody),
	)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandler_UpdateStaffRejectsPasswordOnlyCrossClinicTarget(t *testing.T) {
	accountUpdated := false
	accountID := uint64(9)
	staffRepo := &coreMockStaffRepository{
		lockInClinicFn: func(_ context.Context, clinicID, staffID uint64) (*model.Staff, error) {
			return &model.Staff{
				ID:        staffID,
				ClinicID:  clinicID,
				AccountID: &accountID,
			}, nil
		},
	}
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(_ context.Context, _ uint64, _ string, _ time.Time) error {
			accountUpdated = true
			return nil
		},
	}
	assignmentRepo := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(
			_ context.Context,
			staffID uint64,
		) ([]model.StaffClinicAssignment, error) {
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: 20},
				{StaffID: staffID, ClinicID: 30},
			}, nil
		},
	}
	service := NewServiceWithCredentialAudit(
		staffRepo,
		accountRepo,
		assignmentRepo,
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		nil,
		nil,
		&coreFakeTransactor{},
		noopStaffCredentialAuditTxLogger{},
	)
	handler := NewHandler(
		service,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/staffs/:id", func(c *gin.Context) {
		c.Set("clinic_id", "20")
		c.Set("clinic_ids", []uint64{20})
		c.Set("is_system_admin", false)
		c.Set("user_id", "11")
		c.Next()
	}, handler.UpdateStaff)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/staffs/7",
		bytes.NewBufferString(`{"password":"newpassword1"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.False(t, accountUpdated)
}
