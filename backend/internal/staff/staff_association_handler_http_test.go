package staff_test

// Characterization (golden) tests for the 8 staff association endpoints.
//
// These lock the CURRENT behavior of the association handlers BEFORE any
// behavior-preserving refactor (A3 preamble helper / B1 typed responses /
// A4 audit helper):
//   - exact 200 response body (JSON shape + key names)
//   - nil → [] normalization on SET endpoints
//   - clinic-membership short-circuit: a VerifyClinicMembership failure must
//     propagate via RespondError (404) and MUST NOT call the downstream
//     service method
//   - missing clinic_id short-circuits with 401 before any downstream work
//
// staff_handler_test.go covers only List/Create/Update/Delete; the association
// endpoints had no characterization net. Keep these tests after the refactor.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

// mockStaffClinicAssignmentServiceErr is a StaffClinicAssignmentService mock that
// always fails FindAllByStaffID, used to exercise the 500 passthrough on
// GetStaffClinicAssignments.
type mockStaffClinicAssignmentServiceErr struct{}

func (m *mockStaffClinicAssignmentServiceErr) FindAllByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, fmt.Errorf("db failure")
}
func (m *mockStaffClinicAssignmentServiceErr) FindByStaffAndClinic(
	_ context.Context,
	_, _ uint64,
) (*model.StaffClinicAssignment, error) {
	return nil, fmt.Errorf("db failure")
}
func (m *mockStaffClinicAssignmentServiceErr) Create(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}

// newHandlerWithStaffSvcAndAssignmentErr builds a handler whose
// StaffClinicAssignment dependency always errors, for GetStaffClinicAssignments
// 500 coverage.
func newHandlerWithStaffSvcAndAssignmentErr(staffSvc staffdomain.StaffService) *staffdomain.Handler {
	return staffdomain.NewHandler(
		staffSvc,
		&mockStaffClinicAssignmentServiceErr{},
		nil,
		nil,
		nil,
		nil,
	)
}

// staffAssocCtx builds a gin test context with clinic_id=1 and :id=10.
func staffAssocCtx(w *httptest.ResponseRecorder, method string, body []byte) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	if body != nil {
		c.Request = httptest.NewRequest(method, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, "/", http.NoBody)
	}
	setClinicID(c)
	c.Set("user_id", "7")
	c.Set("is_system_admin", false)
	c.Set("clinic_ids", []uint64{1, 3})
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	return c
}

// notMemberErr is what the real staffService.VerifyClinicMembership returns for
// a staff not assigned to the requesting clinic.
func notMemberErr() error { return apperrors.WrapNotFound("staff", "10") }

func TestSetStaffPermissionGroups_RoutedCompositeAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range []struct {
		name        string
		permissions map[string]bool
		wantStatus  int
		wantCalls   int
	}{
		{
			name: "403 with master staff edit only and no mutation",
			permissions: map[string]bool{
				string(model.ResourceMasterStaff) + ":edit": true,
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "200 with both required edit permissions",
			permissions: map[string]bool{
				string(model.ResourceMasterStaff) + ":edit":      true,
				string(model.ResourceMasterPermission) + ":edit": true,
			},
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mutationCalls := 0
			svc := &mockStaffService{
				setPermissionGroupIDsFn: func(_ context.Context, staffID uint64, ids []uint64) error {
					mutationCalls++
					assert.Equal(t, uint64(10), staffID)
					assert.Equal(t, []uint64{2, 5}, ids)
					return nil
				},
			}
			requirePermission := func(resource, action string) gin.HandlerFunc {
				return func(c *gin.Context) {
					if !tt.permissions[resource+":"+action] {
						staffdomain.RespondError(c, apperrors.WrapForbidden("forbidden"))
						c.Abort()
						return
					}
					c.Next()
				}
			}
			handler := staffdomain.NewHandler(svc, nil, nil, nil, nil, requirePermission)
			router := gin.New()
			protected := router.Group("/api/v1")
			protected.Use(func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Set("user_id", "7")
				c.Next()
			})
			handler.RegisterRoutes(protected)

			request := httptest.NewRequest(
				http.MethodPut,
				"/api/v1/masters/staffs/10/permission-groups",
				bytes.NewBufferString(`{"group_ids":[2,5]}`),
			)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			assert.Equal(t, tt.wantStatus, response.Code)
			assert.Equal(t, tt.wantCalls, mutationCalls)
		})
	}
}

// ---- GET PermissionGroups ----

func TestGetStaffPermissionGroups_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns group_ids", func(t *testing.T) {
		svc := &mockStaffService{
			getPermissionGroupIDsFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				return []uint64{2, 5}, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffPermissionGroups(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"group_ids":[2,5]}`, w.Body.String())
	})

	t.Run("404 when not clinic member; downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
			getPermissionGroupIDsFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
				t.Fatal("downstream GetPermissionGroupIDs must not run when membership fails")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffPermissionGroups(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			getPermissionGroupIDsFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
				return nil, fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffPermissionGroups(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- SET PermissionGroups ----

func TestSetStaffPermissionGroups_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 echoes group_ids", func(t *testing.T) {
		svc := &mockStaffService{
			setPermissionGroupIDsFn: func(_ context.Context, staffID uint64, ids []uint64) error {
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, []uint64{2, 5}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffPermissionGroups(staffAssocCtx(w, http.MethodPut, []byte(`{"group_ids":[2,5]}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"group_ids":[2,5]}`, w.Body.String())
	})

	t.Run("200 normalizes null to empty array", func(t *testing.T) {
		svc := &mockStaffService{
			setPermissionGroupIDsFn: func(_ context.Context, _ uint64, ids []uint64) error {
				assert.Equal(t, []uint64{}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffPermissionGroups(staffAssocCtx(w, http.MethodPut, []byte(`{"group_ids":null}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"group_ids":[]}`, w.Body.String())
	})

	t.Run("404 when not clinic member; downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
			setPermissionGroupIDsFn: func(_ context.Context, _ uint64, _ []uint64) error {
				t.Fatal("downstream SetPermissionGroupIDs must not run when membership fails")
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffPermissionGroups(staffAssocCtx(w, http.MethodPut, []byte(`{"group_ids":[2,5]}`)))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"group_ids":[2,5]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffPermissionGroups(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 for invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffPermissionGroups(staffAssocCtx(w, http.MethodPut, []byte(`not-json`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			setPermissionGroupIDsFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffPermissionGroups(staffAssocCtx(w, http.MethodPut, []byte(`{"group_ids":[2,5]}`)))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- GET/SET ClinicAssignments ----

func TestGetStaffClinicAssignments_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns clinic_ids from assignment service", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).GetStaffClinicAssignments(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[1]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffClinicAssignments(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).GetStaffClinicAssignments(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("500 on assignment service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvcAndAssignmentErr(&mockStaffService{}).GetStaffClinicAssignments(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSetStaffClinicAssignments_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 echoes clinic_ids", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, input *staffdomain.SetClinicAssignmentsInput) error {
				assert.Equal(t, uint64(10), input.StaffID)
				assert.Equal(t, []uint64{1, 3}, input.ClinicIDs)
				assert.Equal(t, []uint64{1, 3}, input.AuthorizedClinicIDs)
				assert.False(t, input.IsSystemAdmin)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[1,3]}`, w.Body.String())
	})

	t.Run("403 when adding a clinic that lacks master-staff:edit", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				t.Fatal("service must not assign a clinic that lacks master-staff:edit")
				return nil
			},
		}
		w := httptest.NewRecorder()
		c := staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`))
		httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
			return clinicID == 1 &&
				resource == string(model.ResourceMasterStaff) &&
				action == "edit"
		})

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("200 deduplicates clinic_ids while preserving order", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, input *staffdomain.SetClinicAssignmentsInput) error {
				assert.Equal(t, []uint64{2, 4}, input.ClinicIDs)
				return nil
			},
		}
		w := httptest.NewRecorder()
		c := staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[2,2,4,2]}`))
		c.Set("clinic_ids", []uint64{2, 4})

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(c)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[2,4]}`, w.Body.String())
	})

	t.Run("system admin scope carries trusted active clinic authority", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, input *staffdomain.SetClinicAssignmentsInput) error {
				assert.Equal(t, uint64(10), input.StaffID)
				assert.Equal(t, []uint64{2}, input.ClinicIDs)
				assert.Equal(t, []uint64{1, 2}, input.AuthorizedClinicIDs)
				assert.True(t, input.IsSystemAdmin)
				return nil
			},
		}
		w := httptest.NewRecorder()
		c := staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[2]}`))
		c.Set("is_system_admin", true)
		c.Set("clinic_ids", []uint64{1, 2})

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(c)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[2]}`, w.Body.String())
	})

	t.Run("401 when non-admin clinic assignments are missing", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				t.Fatal("service must not run without authenticated clinic ids")
				return nil
			},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"clinic_ids":[1]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		setClinicID(c)
		c.Set("is_system_admin", false)
		c.Params = gin.Params{{Key: "id", Value: "10"}}

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("401 when system admin context is missing", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				t.Fatal("service must not run without authenticated admin status")
				return nil
			},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"clinic_ids":[1]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		setClinicID(c)
		c.Set("clinic_ids", []uint64{1})
		c.Params = gin.Params{{Key: "id", Value: "10"}}

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 when more than 50 clinic ids are requested", func(t *testing.T) {
		clinicIDs := make([]uint64, 51)
		for i := range clinicIDs {
			clinicIDs[i] = uint64(i + 1)
		}
		body, err := json.Marshal(map[string][]uint64{"clinic_ids": clinicIDs})
		require.NoError(t, err)
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				t.Fatal("service must not run when request binding rejects too many clinic ids")
				return nil
			},
		}
		w := httptest.NewRecorder()

		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("200 normalizes null to empty array", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, input *staffdomain.SetClinicAssignmentsInput) error {
				assert.Equal(t, []uint64{}, input.ClinicIDs)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":null}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[]}`, w.Body.String())
	})

	t.Run("404 when not clinic member; downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				t.Fatal("downstream SetClinicAssignments must not run when membership fails")
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`)))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"clinic_ids":[1,3]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffClinicAssignments(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 for invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`not-json`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ *staffdomain.SetClinicAssignmentsInput) error {
				return fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`)))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- GET/SET ExcludedReservationTypes ----

func TestGetStaffExcludedReservationTypes_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns reservation_type_ids", func(t *testing.T) {
		svc := &mockStaffService{
			getExcludedReservationTypesFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				return []uint64{7, 9}, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[7,9]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
			getExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
				t.Fatal("downstream must not run when membership fails")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).GetStaffExcludedReservationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			getExcludedReservationTypesFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
				return nil, fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSetStaffExcludedReservationTypes_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 echoes reservation_type_ids", func(t *testing.T) {
		svc := &mockStaffService{
			setExcludedReservationTypesFn: func(_ context.Context, staffID uint64, ids []uint64) error {
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, []uint64{7, 9}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[7,9]}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[7,9]}`, w.Body.String())
	})

	t.Run("200 normalizes null to empty array", func(t *testing.T) {
		svc := &mockStaffService{
			setExcludedReservationTypesFn: func(_ context.Context, _ uint64, ids []uint64) error {
				assert.Equal(t, []uint64{}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":null}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[7,9]}`)))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"reservation_type_ids":[7,9]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffExcludedReservationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 for invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`not-json`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			setExcludedReservationTypesFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[7,9]}`)))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- GET/SET CapableReservationTypes ----

func TestGetStaffCapableReservationTypes_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns reservation_type_ids", func(t *testing.T) {
		svc := &mockStaffService{
			getCapableReservationTypesFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				return []uint64{4}, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[4]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).GetStaffCapableReservationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			getCapableReservationTypesFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
				return nil, fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSetStaffCapableReservationTypes_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 echoes reservation_type_ids", func(t *testing.T) {
		svc := &mockStaffService{
			setCapableReservationTypesFn: func(_ context.Context, clinicID, staffID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, []uint64{4}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[4]}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[4]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[4]}`)))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("200 normalizes null to empty array", func(t *testing.T) {
		svc := &mockStaffService{
			setCapableReservationTypesFn: func(_ context.Context, _, _ uint64, ids []uint64) error {
				assert.Equal(t, []uint64{}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":null}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"reservation_type_ids":[]}`, w.Body.String())
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(`{"reservation_type_ids":[4]}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffCapableReservationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 for invalid JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).SetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`not-json`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			setCapableReservationTypesFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				return fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffCapableReservationTypes(staffAssocCtx(w, http.MethodPut, []byte(`{"reservation_type_ids":[4]}`)))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- preamble: missing clinic_id short-circuits before downstream ----

func TestStaffAssociation_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockStaffService{
		getPermissionGroupIDsFn: func(_ context.Context, _, _ uint64) ([]uint64, error) {
			t.Fatal("downstream must not run when clinic_id is missing")
			return nil, nil
		},
		verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error {
			t.Fatal("membership check must not run when clinic_id is missing")
			return nil
		},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	// NOTE: clinic_id intentionally NOT set.
	newHandlerWithStaffSvc(svc).GetStaffPermissionGroups(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---- GetStaff (9th endpoint refactored by A3 — completes the preamble net) ----

func TestGetStaff_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns staff", func(t *testing.T) {
		svc := &mockStaffService{
			getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				assert.Equal(t, uint64(10), id)
				return &model.Staff{ID: 10, Name: "田中太郎", StaffType: model.StaffTypeDoctor}, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaff(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":10`)
		assert.Contains(t, w.Body.String(), `"name":"田中太郎"`)
	})

	t.Run("404 when not clinic member; membership checked with correct tenant args; downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			// Locks the tenant arguments: membership must be checked for the URL
			// staff id (10) against the authenticated clinic (1). Catches an
			// accidental clinicID/staffID transposition in the preamble helper.
			verifyClinicMembershipFn: func(_ context.Context, staffID, clinicID uint64) error {
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, uint64(1), clinicID)
				return notMemberErr()
			},
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				t.Fatal("downstream GetByID must not run when membership fails")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaff(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("401 when clinic_id missing; membership and downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error {
				t.Fatal("membership must not run when clinic_id is missing")
				return nil
			},
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				t.Fatal("downstream must not run when clinic_id is missing")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		// clinic_id intentionally NOT set.
		newHandlerWithStaffSvc(svc).GetStaff(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 when id is not numeric; membership and downstream not called", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error {
				t.Fatal("membership must not run when id is invalid")
				return nil
			},
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				t.Fatal("downstream must not run when id is invalid")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		setClinicID(c)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		newHandlerWithStaffSvc(svc).GetStaff(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on service error", func(t *testing.T) {
		svc := &mockStaffService{
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, fmt.Errorf("db failure")
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaff(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
