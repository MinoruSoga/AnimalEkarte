package lstep

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagCodeMappingService struct {
	listMappingsFn      func(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	putMappingsForTagFn func(ctx context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error)
}

func (m *mockLstepTagCodeMappingService) ListMappings(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	if m.listMappingsFn != nil {
		return m.listMappingsFn(ctx, clinicID)
	}
	return []*model.LstepTagCodeMapping{}, nil
}

func (m *mockLstepTagCodeMappingService) PutMappingsForTag(ctx context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error) {
	if m.putMappingsForTagFn != nil {
		return m.putMappingsForTagFn(ctx, clinicID, tagName, entries)
	}
	return []*model.LstepTagCodeMapping{}, nil
}

type denyTagPermission struct{}

func newLstepTagCodeMappingsRouter(tagSvc LstepTagCodeMappingService, deny *denyTagPermission, setupCtx gin.HandlerFunc) *gin.Engine {
	permission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {
			if deny != nil {
				c.AbortWithStatus(http.StatusForbidden)
			}
		}
	}
	h := &Handler{tagCodeMapping: tagSvc, requirePermission: permission}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep-tag-code-mappings", setupCtx, h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListTagCodeMappings)
	r.PUT("/clinics/:clinic_id/lstep-tag-code-mappings/:tag_name", setupCtx, h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.ReplaceTagCodeMappingsForTag)
	return r
}

func TestListTagCodeMappings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("403 no permission", func(t *testing.T) {
		r := newLstepTagCodeMappingsRouter(
			&mockLstepTagCodeMappingService{},
			&denyTagPermission{},
			func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
		)

		req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep-tag-code-mappings", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("200 returns mappings", func(t *testing.T) {
		r := newLstepTagCodeMappingsRouter(
			&mockLstepTagCodeMappingService{
				listMappingsFn: func(_ context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
					return []*model.LstepTagCodeMapping{
						{
							ID:       1,
							ClinicID: clinicID,
							TagName:  HlthHealthcheckDoneTag,
							CodeType: model.CodeTypeCheckupType,
							Codes:    pq.StringArray{"CHECKUP_A"},
						},
					}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
		)

		req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep-tag-code-mappings", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var body []map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		if assert.Len(t, body, 1) {
			assert.Equal(t, HlthHealthcheckDoneTag, body[0]["tag_name"])
		}
	})
}

func TestPutTagCodeMappingsForTag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("400 invalid tag", func(t *testing.T) {
		r := newLstepTagCodeMappingsRouter(
			&mockLstepTagCodeMappingService{
				putMappingsForTagFn: func(_ context.Context, _ uint64, _ string, _ []PutMappingEntry) ([]*model.LstepTagCodeMapping, error) {
					return nil, apperrors.WrapInvalidInput("tag_name is invalid")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
		)

		body := `{"entries":[{"code_type":"checkup_type","codes":["A"]}]}`
		req := httptest.NewRequest(http.MethodPut, "/clinics/1/lstep-tag-code-mappings/NOT_ALLOWED", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("200 returns created mappings and location", func(t *testing.T) {
		var gotTag string
		var gotEntries []PutMappingEntry
		r := newLstepTagCodeMappingsRouter(
			&mockLstepTagCodeMappingService{
				putMappingsForTagFn: func(_ context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error) {
					gotTag = tagName
					gotEntries = entries
					return []*model.LstepTagCodeMapping{
						{
							ID:       1,
							ClinicID: clinicID,
							TagName:  tagName,
							CodeType: entries[0].CodeType,
							Codes:    entries[0].Codes,
						},
					}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
		)

		body := `{"entries":[{"code_type":"checkup_type","codes":["A","B"],"species_scope":"dog","age_min":8}]}`
		req := httptest.NewRequest(http.MethodPut, "/clinics/1/lstep-tag-code-mappings/"+url.PathEscape(HlthHealthcheckDoneTag), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/api/v1/lstep-tag-code-mappings/"+HlthHealthcheckDoneTag, w.Header().Get("Location"))
		assert.Equal(t, HlthHealthcheckDoneTag, gotTag)
		if assert.Len(t, gotEntries, 1) {
			assert.Equal(t, "checkup_type", gotEntries[0].CodeType)
			assert.Equal(t, []string{"A", "B"}, gotEntries[0].Codes)
			assert.Equal(t, "dog", gotEntries[0].SpeciesScope)
			if assert.NotNil(t, gotEntries[0].AgeMin) {
				assert.Equal(t, 8, *gotEntries[0].AgeMin)
			}
		}
	})
}
