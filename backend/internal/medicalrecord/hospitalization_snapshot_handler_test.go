package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type snapshotHospitalizationService struct {
	*mockHospitalizationService
	called     bool
	snapshotFn func(context.Context, uint64, listHospitalizationFilters) ([]model.Hospitalization, int64, error)
}

func (s *snapshotHospitalizationService) listSnapshot(ctx context.Context, clinicID uint64, filters listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
	s.called = true
	if s.snapshotFn != nil {
		return s.snapshotFn(ctx, clinicID, filters)
	}
	return []model.Hospitalization{{ID: 1, ClinicID: clinicID}}, 1, nil
}

func TestListHospitalizationsSnapshotHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("snapshot opts into coherent capability", func(t *testing.T) {
		svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{
			listFn: func(context.Context, uint64, *uint64, *uint64, *string, *string, *string, int, int) ([]model.Hospitalization, int64, error) {
				t.Error("snapshot must not use legacy pagination")
				return nil, 0, nil
			},
		}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?snapshot=true", nil)
		setClinicID(c)
		newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, svc.called)
		require.Contains(t, rec.Body.String(), `"snapshot":true`)
		require.Contains(t, rec.Body.String(), `"total":1`)
	})
}

func TestListHospitalizationsSnapshotBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{
		"snapshot=yes", "snapshot=", "snapshot=true&snapshot=false",
	} {
		t.Run("invalid "+query, func(t *testing.T) {
			svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?"+query, nil)
			setClinicID(c)
			newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.False(t, svc.called)
		})
	}
	t.Run("snapshot ignores pagination and preserves filters and clinic scope", func(t *testing.T) {
		svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{}, snapshotFn: func(_ context.Context, clinicID uint64, f listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
			require.Equal(t, uint64(1), clinicID)
			require.NotNil(t, f.PetID)
			require.Equal(t, uint64(2), *f.PetID)
			require.NotNil(t, f.Status)
			require.Equal(t, "admitted", *f.Status)
			return nil, 0, nil
		}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?snapshot=true&page=bad&limit=abc&pet_id=2&status=admitted", nil)
		setClinicID(c)
		newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, svc.called)
		var body struct {
			Data     []HospitalizationResponse `json:"data"`
			Total    int64                     `json:"total"`
			Snapshot bool                      `json:"snapshot"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.NotNil(t, body.Data)
		require.Empty(t, body.Data)
		require.True(t, body.Snapshot)
	})
	for _, query := range []string{"page=1&limit=20", "snapshot=false&page=1&limit=20"} {
		t.Run("legacy "+query, func(t *testing.T) {
			legacyCalled := false
			svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, page, limit int) ([]model.Hospitalization, int64, error) {
				legacyCalled = true
				require.Equal(t, 1, page)
				require.Equal(t, 20, limit)
				return []model.Hospitalization{{ID: 1}}, 25, nil
			}}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?"+query, nil)
			setClinicID(c)
			newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
			require.Equal(t, http.StatusOK, rec.Code)
			require.True(t, legacyCalled)
			require.False(t, svc.called)
			require.NotContains(t, rec.Body.String(), `"snapshot"`)
			require.Contains(t, rec.Body.String(), `"total":25`)
		})
	}
	t.Run("denied clinic prevents snapshot read", func(t *testing.T) {
		svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?snapshot=true", nil)
		setClinicID(c)
		c.Set("is_system_admin", false)
		setResourcePermissionOnlyClinic(c, 999, string(model.ResourceHospitalization), "view")
		newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.False(t, svc.called)
	})
	t.Run("missing capability is 501 not a paginated fallback", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?snapshot=true", nil)
		setClinicID(c)
		newHandlerWithHospitalizationSvc(&mockHospitalizationService{}).ListHospitalizations(c)
		require.Equal(t, http.StatusNotImplemented, rec.Code)
	})
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"overflow", apperrors.WrapInvalidInput("絞り込んでください"), http.StatusBadRequest}, {"read failed", errors.New("private SQL fixture failure"), http.StatusInternalServerError}} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &snapshotHospitalizationService{mockHospitalizationService: &mockHospitalizationService{}, snapshotFn: func(context.Context, uint64, listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
				return nil, 0, tc.err
			}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/hospitalizations?snapshot=true", nil)
			setClinicID(c)
			newHandlerWithHospitalizationSvc(svc).ListHospitalizations(c)
			require.Equal(t, tc.status, rec.Code)
		})
	}
}
