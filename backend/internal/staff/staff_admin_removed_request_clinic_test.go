package staff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// UAT NEW-1: system admin who removes the request clinic from a target staff's
// assignments must still GET/PATCH/PUT/DELETE that staff from the same clinic
// context. Non-admin callers must keep receiving 404 (no existence leak).

func TestHandler_StaffMutationsAfterRequestClinicRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("system admin omits request clinic then GET PATCH PUT DELETE from that clinic are not 404", func(t *testing.T) {
		db, clinicA, clinicB, staff := setupStaffRemovedRequestClinicDB(t)
		handler := newStaffRemovedRequestClinicHandler(db)
		router := newStaffRemovedRequestClinicRouter(handler, clinicA.ID, true, []uint64{clinicA.ID, clinicB.ID})
		path := fmt.Sprintf("/staffs/%d", staff.ID)

		putRecorder := serveStaffJSON(
			t,
			router,
			http.MethodPut,
			path+"/clinics",
			map[string]any{"clinic_ids": []uint64{clinicB.ID}},
		)
		require.Equal(t, http.StatusOK, putRecorder.Code, putRecorder.Body.String())

		clinicsRecorder := serveStaffJSON(t, router, http.MethodGet, path+"/clinics", nil)
		assert.Equal(t, http.StatusOK, clinicsRecorder.Code, clinicsRecorder.Body.String())
		var clinics staffClinicAssignmentsResponse
		require.NoError(t, json.Unmarshal(clinicsRecorder.Body.Bytes(), &clinics))
		assert.Equal(t, []uint64{clinicB.ID}, clinics.ClinicIDs)

		getRecorder := serveStaffJSON(t, router, http.MethodGet, path, nil)
		assert.NotEqual(t, http.StatusNotFound, getRecorder.Code, getRecorder.Body.String())
		assert.Equal(t, http.StatusOK, getRecorder.Code, getRecorder.Body.String())
		assert.Contains(t, getRecorder.Body.String(), `"id":`+strconv.FormatUint(staff.ID, 10))

		patchRecorder := serveStaffJSON(
			t,
			router,
			http.MethodPatch,
			path,
			map[string]any{"name": "所属解除後も編集できるスタッフ"},
		)
		assert.NotEqual(t, http.StatusNotFound, patchRecorder.Code, patchRecorder.Body.String())
		assert.Equal(t, http.StatusOK, patchRecorder.Code, patchRecorder.Body.String())
		assert.Contains(t, patchRecorder.Body.String(), "所属解除後も編集できるスタッフ")

		restoreRecorder := serveStaffJSON(
			t,
			router,
			http.MethodPut,
			path+"/clinics",
			map[string]any{"clinic_ids": []uint64{clinicB.ID}},
		)
		assert.NotEqual(t, http.StatusNotFound, restoreRecorder.Code, restoreRecorder.Body.String())
		assert.Equal(t, http.StatusOK, restoreRecorder.Code, restoreRecorder.Body.String())

		deleteRecorder := serveStaffJSON(t, router, http.MethodDelete, path, nil)
		assert.NotEqual(t, http.StatusNotFound, deleteRecorder.Code, deleteRecorder.Body.String())
		assert.Equal(t, http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	})

	t.Run("non-admin still 404s when target is not assigned to request clinic", func(t *testing.T) {
		db, clinicA, clinicB, staff := setupStaffRemovedRequestClinicDB(t)
		require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicA.ID).
			Delete(&model.StaffClinicAssignment{}).Error)
		handler := newStaffRemovedRequestClinicHandler(db)
		router := newStaffRemovedRequestClinicRouter(handler, clinicA.ID, false, []uint64{clinicA.ID})
		path := fmt.Sprintf("/staffs/%d", staff.ID)

		getRecorder := serveStaffJSON(t, router, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusNotFound, getRecorder.Code, getRecorder.Body.String())

		patchRecorder := serveStaffJSON(
			t,
			router,
			http.MethodPatch,
			path,
			map[string]any{"name": "他院スタッフは更新できない"},
		)
		assert.Equal(t, http.StatusNotFound, patchRecorder.Code, patchRecorder.Body.String())

		putRecorder := serveStaffJSON(
			t,
			router,
			http.MethodPut,
			path+"/clinics",
			map[string]any{"clinic_ids": []uint64{clinicB.ID}},
		)
		assert.Equal(t, http.StatusNotFound, putRecorder.Code, putRecorder.Body.String())

		deleteRecorder := serveStaffJSON(t, router, http.MethodDelete, path, nil)
		assert.Equal(t, http.StatusNotFound, deleteRecorder.Code, deleteRecorder.Body.String())
	})
}

func TestHandler_SystemAdminGetAndPutClinicsSkipRequestClinicMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const (
		staffID = uint64(7)
		clinicA = uint64(10)
		clinicB = uint64(20)
	)

	tests := []struct {
		name          string
		isSystemAdmin bool
		adminSet      bool
		wantGET       int
		wantPUT       int
		wantGetByID   bool
		wantSet       bool
	}{
		{
			name:          "system admin GET and PUT clinics succeed without request-clinic membership",
			isSystemAdmin: true,
			adminSet:      true,
			wantGET:       http.StatusOK,
			wantPUT:       http.StatusOK,
			wantGetByID:   true,
			wantSet:       true,
		},
		{
			name:          "non-admin still 404s when membership is missing",
			isSystemAdmin: false,
			adminSet:      true,
			wantGET:       http.StatusNotFound,
			wantPUT:       http.StatusNotFound,
		},
		{
			name:     "missing system-admin peek stays fail-closed and 404s",
			adminSet: false,
			wantGET:  http.StatusNotFound,
			wantPUT:  http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotByID := false
			setCalled := false
			service := &removedRequestClinicHandlerService{
				verifyFn: func(_ context.Context, gotStaffID, gotClinicID uint64) error {
					assert.Equal(t, staffID, gotStaffID)
					assert.Equal(t, clinicA, gotClinicID)
					return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
				},
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					gotByID = true
					assert.Equal(t, staffID, id)
					return &model.Staff{ID: id, Name: "他院所属スタッフ", ClinicID: clinicB}, nil
				},
				getInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					t.Fatalf("clinic-scoped GET must not run without request-clinic membership: clinic=%d staff=%d", clinicID, id)
					return nil, nil
				},
				setClinicsFn: func(_ context.Context, input *SetClinicAssignmentsInput) error {
					setCalled = true
					assert.Equal(t, staffID, input.StaffID)
					assert.Equal(t, []uint64{clinicB}, input.ClinicIDs)
					assert.True(t, input.IsSystemAdmin)
					return nil
				},
			}
			handler := NewHandler(service, nil, nil, nil, nil, nil)
			router := gin.New()
			withAuth := func(c *gin.Context) {
				c.Set("clinic_id", strconv.FormatUint(clinicA, 10))
				c.Set("clinic_ids", []uint64{clinicA, clinicB})
				if test.adminSet {
					c.Set("is_system_admin", test.isSystemAdmin)
				}
				attachAllowAllClinicPermission(c)
				c.Next()
			}
			router.GET("/staffs/:id", withAuth, handler.GetStaff)
			router.PUT("/staffs/:id/clinics", withAuth, handler.SetStaffClinicAssignments)

			getRecorder := serveStaffJSON(t, router, http.MethodGet, "/staffs/7", nil)
			assert.Equal(t, test.wantGET, getRecorder.Code, getRecorder.Body.String())
			assert.Equal(t, test.wantGetByID, gotByID)

			putRecorder := serveStaffJSON(
				t,
				router,
				http.MethodPut,
				"/staffs/7/clinics",
				map[string]any{"clinic_ids": []uint64{clinicB}},
			)
			assert.Equal(t, test.wantPUT, putRecorder.Code, putRecorder.Body.String())
			assert.Equal(t, test.wantSet, setCalled)
		})
	}
}

func TestStaffService_UpdateAndDeleteAfterRequestClinicRemoved(t *testing.T) {
	const (
		staffID = uint64(7)
		clinicA = uint64(10)
		clinicB = uint64(20)
		clinicC = uint64(30)
	)
	name := "所属解除後の更新"

	tests := []struct {
		name          string
		isSystemAdmin bool
		assignments   []uint64
		wantUpdateErr bool
		wantNotFound  bool
		wantConflict  bool
		wantUpdated   bool
		wantDeleted   bool
	}{
		{
			name:          "system admin can update when only another clinic remains",
			isSystemAdmin: true,
			assignments:   []uint64{clinicB},
			wantUpdated:   true,
		},
		{
			name:          "non-admin update 404s when request clinic assignment is gone",
			assignments:   []uint64{clinicB},
			wantUpdateErr: true,
			wantNotFound:  true,
		},
		{
			name:          "system admin delete is 409 not 404 when other clinics remain",
			isSystemAdmin: true,
			assignments:   []uint64{clinicB, clinicC},
			wantConflict:  true,
		},
		{
			name:          "system admin delete of last remaining other clinic is not 404",
			isSystemAdmin: true,
			assignments:   []uint64{clinicB},
			wantDeleted:   true,
		},
		{
			name:         "non-admin delete 404s when request clinic assignment is gone",
			assignments:  []uint64{clinicB},
			wantNotFound: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updated := false
			deleted := false
			repo := &coreMockStaffRepository{
				lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					for _, assigned := range test.assignments {
						if assigned == clinicID {
							return &model.Staff{ID: id, ClinicID: assigned}, nil
						}
					}
					return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
				},
				lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, ClinicID: test.assignments[0]}, nil
				},
				findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, Name: name, ClinicID: test.assignments[0]}, nil
				},
				findByIDInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					for _, assigned := range test.assignments {
						if assigned == clinicID {
							return &model.Staff{ID: id, Name: name, ClinicID: assigned}, nil
						}
					}
					return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
				},
				updateFn: func(_ context.Context, _, id uint64, _ map[string]any) error {
					assert.Equal(t, staffID, id)
					updated = true
					return nil
				},
				deleteFn: func(_ context.Context, _, id uint64) error {
					assert.Equal(t, staffID, id)
					deleted = true
					return nil
				},
			}
			assignments := &coreMockStaffClinicAssignmentRepository{
				lockActiveFn: func(_ context.Context, id uint64) ([]model.StaffClinicAssignment, error) {
					rows := make([]model.StaffClinicAssignment, 0, len(test.assignments))
					for i, clinicID := range test.assignments {
						rows = append(rows, model.StaffClinicAssignment{
							StaffID:  id,
							ClinicID: clinicID,
							IsMain:   i == 0,
						})
					}
					return rows, nil
				},
			}
			service := newCoreStaffService(
				repo,
				&coreMockAccountRepository{},
				assignments,
				&coreMockReservationQueryRepository{},
				&coreMockShiftEntryRepository{},
				&coreFakeTransactor{},
			)

			if test.wantUpdated || test.wantUpdateErr {
				result, err := service.Update(
					context.Background(),
					clinicA,
					staffID,
					&UpdateStaffInput{
						Name:                &name,
						AuthorizedClinicIDs: []uint64{clinicA, clinicB, clinicC},
						IsSystemAdmin:       test.isSystemAdmin,
					},
				)
				if test.wantUpdateErr || test.wantNotFound {
					require.Error(t, err)
					assert.Nil(t, result)
					assert.False(t, updated)
					if test.wantNotFound {
						assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
					}
					return
				}
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, updated)
				return
			}

			err := service.Delete(context.Background(), clinicA, staffID, test.isSystemAdmin)
			if test.wantNotFound {
				require.Error(t, err)
				assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
				assert.False(t, deleted)
				return
			}
			if test.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
				assert.False(t, deleted)
				return
			}
			require.NoError(t, err)
			assert.True(t, deleted)
		})
	}
}

func TestAuthorizeGlobalStaffUpdateAllowsSystemAdminWithoutCurrentClinic(t *testing.T) {
	err := authorizeGlobalStaffUpdate(
		7,
		10,
		[]model.StaffClinicAssignment{{StaffID: 7, ClinicID: 20}},
		nil,
		true,
	)

	require.NoError(t, err)
}

func TestStaffService_UpdateOccupationLockedAgainstWriteClinicAfterRequestClinicRemoved(t *testing.T) {
	const (
		staffID = uint64(7)
		clinicA = uint64(10)
		clinicB = uint64(20)
		occA    = uint64(31)
		occB    = uint64(32)
	)

	t.Run("request-clinic occupation does not persist onto remaining clinic", func(t *testing.T) {
		lockedClinic := uint64(0)
		updated := false
		repo := &coreMockStaffRepository{
			lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
				t.Fatalf("in-clinic lock must not run for system admin: clinic=%d staff=%d", clinicID, id)
				return nil, nil
			},
			lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicB}, nil
			},
			findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicB}, nil
			},
			updateFn: func(_ context.Context, _, id uint64, _ map[string]any) error {
				updated = true
				assert.Equal(t, staffID, id)
				return nil
			},
		}
		occupations := &mockOccupationRepository{
			lockForShareFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
				lockedClinic = clinicID
				if clinicID != clinicB || id != occB {
					return nil, apperrors.WrapNotFound("occupation", fmt.Sprintf("%d", id))
				}
				return &model.Occupation{ID: id, ClinicID: clinicID}, nil
			},
		}
		service := newStaffServiceWithOccupationRepo(repo, occupations, []uint64{clinicB})
		occupationID := occA
		result, err := service.Update(
			context.Background(),
			clinicA,
			staffID,
			&UpdateStaffInput{
				OccupationID:        &occupationID,
				AuthorizedClinicIDs: []uint64{clinicA, clinicB},
				IsSystemAdmin:       true,
			},
		)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
		assert.Nil(t, result)
		assert.Equal(t, clinicB, lockedClinic)
		assert.False(t, updated)
	})

	t.Run("remaining-clinic occupation is locked and persisted", func(t *testing.T) {
		lockedClinic := uint64(0)
		writeClinic := uint64(0)
		repo := &coreMockStaffRepository{
			lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
				t.Fatalf("in-clinic lock must not run for system admin: clinic=%d staff=%d", clinicID, id)
				return nil, nil
			},
			lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicB}, nil
			},
			findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicB}, nil
			},
			updateFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) error {
				writeClinic = clinicID
				assert.Equal(t, staffID, id)
				return nil
			},
		}
		occupations := &mockOccupationRepository{
			lockForShareFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
				lockedClinic = clinicID
				if clinicID != clinicB || id != occB {
					return nil, apperrors.WrapNotFound("occupation", fmt.Sprintf("%d", id))
				}
				return &model.Occupation{ID: id, ClinicID: clinicID}, nil
			},
		}
		service := newStaffServiceWithOccupationRepo(repo, occupations, []uint64{clinicB})
		occupationID := occB
		result, err := service.Update(
			context.Background(),
			clinicA,
			staffID,
			&UpdateStaffInput{
				OccupationID:        &occupationID,
				AuthorizedClinicIDs: []uint64{clinicA, clinicB},
				IsSystemAdmin:       true,
			},
		)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, clinicB, lockedClinic)
		assert.Equal(t, clinicB, writeClinic)
	})
}

func newStaffServiceWithOccupationRepo(
	repo *coreMockStaffRepository,
	occupations *mockOccupationRepository,
	assignedClinics []uint64,
) StaffService {
	assignments := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(_ context.Context, id uint64) ([]model.StaffClinicAssignment, error) {
			rows := make([]model.StaffClinicAssignment, 0, len(assignedClinics))
			for i, clinicID := range assignedClinics {
				rows = append(rows, model.StaffClinicAssignment{
					StaffID:  id,
					ClinicID: clinicID,
					IsMain:   i == 0,
				})
			}
			return rows, nil
		},
	}
	return NewStaffService(
		repo,
		&coreMockAccountRepository{},
		assignments,
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		occupations,
		nil,
		&coreFakeTransactor{},
	)
}

type removedRequestClinicHandlerService struct {
	StaffService
	verifyFn      func(context.Context, uint64, uint64) error
	getByIDFn     func(context.Context, uint64) (*model.Staff, error)
	getInClinicFn func(context.Context, uint64, uint64) (*model.Staff, error)
	setClinicsFn  func(context.Context, *SetClinicAssignmentsInput) error
}

func (s *removedRequestClinicHandlerService) VerifyClinicMembership(
	ctx context.Context,
	staffID, clinicID uint64,
) error {
	return s.verifyFn(ctx, staffID, clinicID)
}

func (s *removedRequestClinicHandlerService) GetByID(
	ctx context.Context,
	id uint64,
) (*model.Staff, error) {
	return s.getByIDFn(ctx, id)
}

func (s *removedRequestClinicHandlerService) GetByIDInClinic(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	return s.getInClinicFn(ctx, clinicID, staffID)
}

func (s *removedRequestClinicHandlerService) SetClinicAssignments(
	ctx context.Context,
	input *SetClinicAssignmentsInput,
) error {
	return s.setClinicsFn(ctx, input)
}

func setupStaffRemovedRequestClinicDB(
	t *testing.T,
) (db *gorm.DB, clinicA, clinicB *model.Clinic, staff *model.Staff) {
	t.Helper()
	db = testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Occupation{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ShiftEntry{},
		&model.Reservation{},
		&model.Hospitalization{},
		&model.Examination{},
		&model.MedicalRecordAddendum{},
		&model.VitalRecord{},
		&model.DailyRecord{},
		&model.CashRegisterClose{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_clinic_assignments, staffs, occupations, accounts, shift_entries CASCADE",
	).Error)

	company := &model.Company{Name: "UAT NEW-1 company"}
	require.NoError(t, db.Create(company).Error)
	clinicA = &model.Clinic{CompanyID: company.ID, Name: "UAT NEW-1 医院A", IsActive: true}
	clinicB = &model.Clinic{CompanyID: company.ID, Name: "UAT NEW-1 医院B", IsActive: true}
	require.NoError(t, db.Create(clinicA).Error)
	require.NoError(t, db.Create(clinicB).Error)

	staff = &model.Staff{
		ClinicID:  clinicA.ID,
		Name:      "UAT NEW-1 対象スタッフ",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicA.ID,
		IsMain:   true,
	}).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicB.ID,
	}).Error)
	return db, clinicA, clinicB, staff
}

func newStaffRemovedRequestClinicHandler(db *gorm.DB) *Handler {
	assignmentRepo := NewStaffClinicAssignmentRepository(db)
	service := NewStaffService(
		NewStaffRepository(db),
		nil,
		assignmentRepo,
		&stubReservationForStaff{},
		NewShiftEntryRepository(db),
		nil,
		nil,
		NewOccupationRepository(db),
		clinic.NewClinicRepository(db),
		persistence.NewTransactor(db),
	)
	return NewHandler(
		service,
		NewStaffClinicAssignmentService(assignmentRepo),
		nil,
		nil,
		nil,
		nil,
	)
}

func newStaffRemovedRequestClinicRouter(
	handler *Handler,
	clinicID uint64,
	isSystemAdmin bool,
	authorized []uint64,
) *gin.Engine {
	router := gin.New()
	withAuth := func(c *gin.Context) {
		c.Set("clinic_id", strconv.FormatUint(clinicID, 10))
		c.Set("is_system_admin", isSystemAdmin)
		c.Set("clinic_ids", authorized)
		attachAllowAllClinicPermission(c)
		c.Next()
	}
	router.GET("/staffs/:id", withAuth, handler.GetStaff)
	router.PATCH("/staffs/:id", withAuth, handler.UpdateStaff)
	router.GET("/staffs/:id/clinics", withAuth, handler.GetStaffClinicAssignments)
	router.PUT("/staffs/:id/clinics", withAuth, handler.SetStaffClinicAssignments)
	router.DELETE("/staffs/:id", withAuth, handler.DeleteStaff)
	return router
}

func serveStaffJSON(
	t *testing.T,
	router *gin.Engine,
	method, path string,
	body map[string]any,
) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(method, path, http.NoBody)
	} else {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		request = httptest.NewRequest(method, path, bytes.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
	}
	router.ServeHTTP(recorder, request)
	return recorder
}
