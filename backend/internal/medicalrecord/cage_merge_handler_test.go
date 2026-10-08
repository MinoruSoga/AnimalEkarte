package medicalrecord

// cage_merge_handler_test.go — MergeCage / UnmergeCage の HTTP 境界テスト。
// バインド/バリデーションとサービスエラー→ステータス変換をスタブサービスで検証する。

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
	"github.com/animal-ekarte/backend/internal/model"
)

// stubCageMergerService は CageService（埋込・未使用メソッドは nil）と
// CageMerger の両方を満たすテストダブル。NewCageHandler の型アサーション経路を
// 本物と同じ形で通す。
type stubCageMergerService struct {
	CageService
	mergeFn   func(ctx context.Context, clinicID, representativeID uint64, memberCageIDs []uint64) (*model.Cage, error)
	unmergeFn func(ctx context.Context, clinicID, representativeID uint64) error
}

func (s *stubCageMergerService) Merge(ctx context.Context, clinicID, representativeID uint64, memberCageIDs []uint64) (*model.Cage, error) {
	return s.mergeFn(ctx, clinicID, representativeID, memberCageIDs)
}

func (s *stubCageMergerService) Unmerge(ctx context.Context, clinicID, representativeID uint64) error {
	return s.unmergeFn(ctx, clinicID, representativeID)
}

func newMergeCageContext(method, paramID string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(bodyBytes)
	} else {
		reader = bytes.NewReader(nil)
	}
	c.Request = httptest.NewRequest(method, "/", reader)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: paramID}}
	return w, c
}

func TestMergeCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{"member_cage_ids": []int{2, 3}}

	t.Run("merges cages successfully and returns representative", func(t *testing.T) {
		svc := &stubCageMergerService{
			mergeFn: func(_ context.Context, clinicID, representativeID uint64, memberCageIDs []uint64) (*model.Cage, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(1), representativeID)
				assert.Equal(t, []uint64{2, 3}, memberCageIDs)
				return &model.Cage{ID: representativeID, ClinicID: clinicID, Name: "代表ケージ"}, nil
			},
		}
		h := newHandlerWithCageSvc(svc)
		w, c := newMergeCageContext(http.MethodPost, "1", validBody)
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"代表ケージ"`)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "1", validBody)

		h.MergeCage(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for non-numeric id", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "abc", validBody)
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when member_cage_ids is empty", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "1", map[string]any{"member_cage_ids": []int{}})
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when member_cage_ids is missing", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "1", map[string]any{})
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("maps service errors to status codes", func(t *testing.T) {
		tests := []struct {
			name       string
			svcErr     error
			wantStatus int
		}{
			{name: "NotFound → 404", svcErr: apperrors.WrapNotFound("cage", "999"), wantStatus: http.StatusNotFound},
			{name: "InvalidInput → 400", svcErr: apperrors.WrapInvalidInput("member_cage_ids に重複があります"), wantStatus: http.StatusBadRequest},
			{name: "Conflict → 409", svcErr: apperrors.WrapConflict("すでに連結済みのケージは連結できません"), wantStatus: http.StatusConflict},
			{name: "generic error → 500", svcErr: fmt.Errorf("db failure"), wantStatus: http.StatusInternalServerError},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &stubCageMergerService{
					mergeFn: func(_ context.Context, _, _ uint64, _ []uint64) (*model.Cage, error) {
						return nil, tt.svcErr
					},
				}
				h := newHandlerWithCageSvc(svc)
				w, c := newMergeCageContext(http.MethodPost, "1", validBody)
				setClinicID(c)

				h.MergeCage(c)

				assert.Equal(t, tt.wantStatus, w.Code)
			})
		}
	})

	t.Run("returns 500 when wired service lacks CageMerger", func(t *testing.T) {
		h := NewCageHandler(nil)
		w, c := newMergeCageContext(http.MethodPost, "1", validBody)
		setClinicID(c)

		h.MergeCage(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// newUnmergeCageRouter mirrors newDeleteCageRouter: a 204 status-only response is
// only flushed through the gin engine (c.Status alone does not write the recorder
// on direct handler invocation), so the happy path goes through a real route.
func newUnmergeCageRouter(svc CageService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCageSvc(svc)
	r.POST("/cages/:id/unmerge", func(c *gin.Context) {
		setClinicID(c)
	}, h.UnmergeCage)
	return r
}

func TestUnmergeCage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("unmerges successfully and returns 204", func(t *testing.T) {
		svc := &stubCageMergerService{
			unmergeFn: func(_ context.Context, clinicID, representativeID uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), representativeID)
				return nil
			},
		}
		router := newUnmergeCageRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cages/7/unmerge", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "7", nil)

		h.UnmergeCage(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for non-numeric id", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageMergerService{})
		w, c := newMergeCageContext(http.MethodPost, "xyz", nil)
		setClinicID(c)

		h.UnmergeCage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("maps service errors to status codes", func(t *testing.T) {
		tests := []struct {
			name       string
			svcErr     error
			wantStatus int
		}{
			{name: "NotFound → 404", svcErr: apperrors.WrapNotFound("cage", "999"), wantStatus: http.StatusNotFound},
			{name: "Conflict → 409", svcErr: apperrors.WrapConflict("連結グループを持たないケージは連結解除できません"), wantStatus: http.StatusConflict},
			{name: "generic error → 500", svcErr: fmt.Errorf("db failure"), wantStatus: http.StatusInternalServerError},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &stubCageMergerService{
					unmergeFn: func(_ context.Context, _, _ uint64) error {
						return tt.svcErr
					},
				}
				h := newHandlerWithCageSvc(svc)
				w, c := newMergeCageContext(http.MethodPost, "7", nil)
				setClinicID(c)

				h.UnmergeCage(c)

				assert.Equal(t, tt.wantStatus, w.Code)
			})
		}
	})

	t.Run("returns 500 when wired service lacks CageMerger", func(t *testing.T) {
		h := NewCageHandler(nil)
		w, c := newMergeCageContext(http.MethodPost, "7", nil)
		setClinicID(c)

		h.UnmergeCage(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestMergeCageRequest_Binding は mergeCageRequest の binding 規則を直接検証する。
func TestMergeCageRequest_Binding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// maxCageMergeMembers 超の配列は binding max=32 で 400 に落ちる。
	overLimitIDs := make([]uint64, maxCageMergeMembers+1)
	for i := range overLimitIDs {
		overLimitIDs[i] = uint64(i + 1)
	}
	overLimitBody, err := json.Marshal(map[string]any{"member_cage_ids": overLimitIDs})
	require.NoError(t, err)

	tests := []struct {
		name      string
		body      string
		wantError bool
	}{
		{name: "valid single member", body: `{"member_cage_ids":[2]}`, wantError: false},
		{name: "valid multiple members", body: `{"member_cage_ids":[2,3,10]}`, wantError: false},
		{name: "empty array rejected", body: `{"member_cage_ids":[]}`, wantError: true},
		{name: "missing field rejected", body: `{}`, wantError: true},
		{name: "null rejected", body: `{"member_cage_ids":null}`, wantError: true},
		{name: "string id rejected", body: `{"member_cage_ids":["2"]}`, wantError: true},
		{name: "over max members rejected", body: string(overLimitBody), wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req mergeCageRequest
			err := c.ShouldBindJSON(&req)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// stubCageOccupancyService は CageService（埋込・未使用メソッドは nil）と
// CageOccupancyReader の両方を満たすテストダブル。
type stubCageOccupancyService struct {
	CageService
	occupiedCageIDsFn func(ctx context.Context, clinicID uint64) ([]uint64, error)
}

func (s *stubCageOccupancyService) OccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error) {
	return s.occupiedCageIDsFn(ctx, clinicID)
}

// CageService のみで CageOccupancyReader を実装しないテストダブル（fail-closed 検証用）。
type stubCageServiceNoOccupancy struct {
	CageService
}

func TestOccupiedCageIDsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns occupied cage id set", func(t *testing.T) {
		svc := &stubCageOccupancyService{
			occupiedCageIDsFn: func(_ context.Context, clinicID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				return []uint64{3, 7}, nil
			},
		}
		h := newHandlerWithCageSvc(svc)
		w, c := newMergeCageContext(http.MethodGet, "", nil)
		setClinicID(c)

		h.OccupiedCageIDs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"occupied_cage_ids":[3,7]}`, w.Body.String())
	})

	t.Run("empty result returns empty array not null", func(t *testing.T) {
		svc := &stubCageOccupancyService{
			occupiedCageIDsFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				return nil, nil
			},
		}
		h := newHandlerWithCageSvc(svc)
		w, c := newMergeCageContext(http.MethodGet, "", nil)
		setClinicID(c)

		h.OccupiedCageIDs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"occupied_cage_ids":[]}`, w.Body.String())
	})

	t.Run("service error propagates status", func(t *testing.T) {
		svc := &stubCageOccupancyService{
			occupiedCageIDsFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				return nil, apperrors.WrapInternalServerError("db down")
			},
		}
		h := newHandlerWithCageSvc(svc)
		w, c := newMergeCageContext(http.MethodGet, "", nil)
		setClinicID(c)

		h.OccupiedCageIDs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("missing occupancy face is fail-closed 500", func(t *testing.T) {
		h := newHandlerWithCageSvc(&stubCageServiceNoOccupancy{})
		w, c := newMergeCageContext(http.MethodGet, "", nil)
		setClinicID(c)

		h.OccupiedCageIDs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
