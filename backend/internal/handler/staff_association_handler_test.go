package handler

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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

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
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	return c
}

// notMemberErr is what the real staffService.VerifyClinicMembership returns for
// a staff not assigned to the requesting clinic.
func notMemberErr() error { return apperrors.WrapNotFound("staff", "10") }

// ---- GET PermissionGroups ----

func TestGetStaffPermissionGroups_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns group_ids", func(t *testing.T) {
		svc := &mockStaffService{
			getPermissionGroupIDsFn: func(_ context.Context, staffID uint64) ([]uint64, error) {
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
			getPermissionGroupIDsFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				t.Fatal("downstream GetPermissionGroupIDs must not run when membership fails")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffPermissionGroups(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
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
}

// ---- GET/SET ClinicAssignments ----

func TestGetStaffClinicAssignments_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns clinic_ids from assignment service", func(t *testing.T) {
		// mockStaffClinicAssignmentService.FindAllByStaffID returns clinics {1,3}.
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(&mockStaffService{}).GetStaffClinicAssignments(staffAssocCtx(w, http.MethodGet, nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[1,3]}`, w.Body.String())
	})

	t.Run("404 when not clinic member", func(t *testing.T) {
		svc := &mockStaffService{
			verifyClinicMembershipFn: func(_ context.Context, _, _ uint64) error { return notMemberErr() },
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffClinicAssignments(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSetStaffClinicAssignments_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 echoes clinic_ids", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, staffID uint64, ids []uint64) error {
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, []uint64{1, 3}, ids)
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_ids":[1,3]}`, w.Body.String())
	})

	t.Run("200 normalizes null to empty array", func(t *testing.T) {
		svc := &mockStaffService{
			setClinicAssignmentsFn: func(_ context.Context, _ uint64, ids []uint64) error {
				assert.Equal(t, []uint64{}, ids)
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
			setClinicAssignmentsFn: func(_ context.Context, _ uint64, _ []uint64) error {
				t.Fatal("downstream SetClinicAssignments must not run when membership fails")
				return nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).SetStaffClinicAssignments(staffAssocCtx(w, http.MethodPut, []byte(`{"clinic_ids":[1,3]}`)))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ---- GET/SET ExcludedReservationTypes ----

func TestGetStaffExcludedReservationTypes_Characterization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns reservation_type_ids", func(t *testing.T) {
		svc := &mockStaffService{
			getExcludedReservationTypesFn: func(_ context.Context, staffID uint64) ([]uint64, error) {
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
			getExcludedReservationTypesFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				t.Fatal("downstream must not run when membership fails")
				return nil, nil
			},
		}
		w := httptest.NewRecorder()
		newHandlerWithStaffSvc(svc).GetStaffExcludedReservationTypes(staffAssocCtx(w, http.MethodGet, nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
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
}

// ---- preamble: missing clinic_id short-circuits before downstream ----

func TestStaffAssociation_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockStaffService{
		getPermissionGroupIDsFn: func(_ context.Context, _ uint64) ([]uint64, error) {
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
