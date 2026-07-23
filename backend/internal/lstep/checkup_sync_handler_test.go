package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- mock CheckupSyncService ----

type mockCheckupSyncService struct {
	previewCheckupSyncFn func(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error)
	createCheckupSyncFn  func(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error)
}

func (m *mockCheckupSyncService) PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
	return m.previewCheckupSyncFn(ctx, clinicID, input, actorID)
}

func (m *mockCheckupSyncService) CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
	return m.createCheckupSyncFn(ctx, clinicID, input, actorID)
}

func newHandlerWithCheckupSyncSvc(svc CheckupSyncService) *Handler {
	return &Handler{checkupSync: svc}
}

// ---- toCheckupSyncPreviewOwnerResponse ----

func TestToCheckupSyncPreviewOwnerResponse(t *testing.T) {
	t.Run("populates all fields including optional date pointers", func(t *testing.T) {
		lastVisit := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		lastCheckup := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		minAge := 2
		maxAge := 10

		o := &CheckupSyncPreviewOwner{
			OwnerID:             1,
			OwnerName:           "田中太郎",
			PetNames:            []string{"ポチ"},
			LastVisitDate:       &lastVisit,
			HasLine:             true,
			IsOptOut:            false,
			HasLivingPet:        true,
			ExclusionReason:     nil,
			CurrentTags:         []string{"tag1"},
			MinPetAgeYears:      &minAge,
			MaxPetAgeYears:      &maxAge,
			HasChronicCondition: true,
			CPMStage:            string(CPMStageCore),
			TotalAmount:         10000,
			AnnualVisitCount:    3,
			LastCheckupDate:     &lastCheckup,
		}

		resp := toCheckupSyncPreviewOwnerResponse(o)

		assert.Equal(t, "1", resp.OwnerID)
		assert.Equal(t, "田中太郎", resp.OwnerName)
		assert.Equal(t, []string{"ポチ"}, resp.PetNames)
		require.NotNil(t, resp.LastVisitDate)
		assert.Equal(t, "2026-05-01", *resp.LastVisitDate)
		require.NotNil(t, resp.LastCheckupDate)
		assert.Equal(t, "2026-04-01", *resp.LastCheckupDate)
		assert.True(t, resp.HasLine)
		assert.False(t, resp.IsOptOut)
		assert.True(t, resp.HasLivingPet)
		assert.Nil(t, resp.ExclusionReason)
		assert.Equal(t, []string{"tag1"}, resp.CurrentTags)
		require.NotNil(t, resp.MinPetAgeYears)
		assert.Equal(t, 2, *resp.MinPetAgeYears)
		require.NotNil(t, resp.MaxPetAgeYears)
		assert.Equal(t, 10, *resp.MaxPetAgeYears)
		assert.True(t, resp.HasChronicCondition)
		assert.Equal(t, string(CPMStageCore), resp.CPMStage)
		assert.Equal(t, int64(10000), resp.TotalAmount)
		assert.Equal(t, int64(3), resp.AnnualVisitCount)
	})

	t.Run("nil date pointers and exclusion reason are preserved as nil", func(t *testing.T) {
		reason := "Lステップ配信停止中"
		o := &CheckupSyncPreviewOwner{
			OwnerID:         2,
			OwnerName:       "鈴木花子",
			ExclusionReason: &reason,
		}

		resp := toCheckupSyncPreviewOwnerResponse(o)

		assert.Equal(t, "2", resp.OwnerID)
		assert.Nil(t, resp.LastVisitDate)
		assert.Nil(t, resp.LastCheckupDate)
		require.NotNil(t, resp.ExclusionReason)
		assert.Equal(t, reason, *resp.ExclusionReason)
		assert.False(t, resp.HasChronicCondition)
	})
}

// ---- toCheckupSyncPreviewResponse ----

func TestToCheckupSyncPreviewResponse(t *testing.T) {
	t.Run("aggregates owners and counts", func(t *testing.T) {
		result := &PreviewCheckupSyncResult{
			Owners: []CheckupSyncPreviewOwner{
				{OwnerID: 1, OwnerName: "田中太郎"},
				{OwnerID: 2, OwnerName: "鈴木花子"},
			},
			TotalCount:       2,
			EligibleCount:    1,
			LineLinkedCount:  2,
			OptOutCount:      0,
			NoLivingPetCount: 0,
		}

		resp := toCheckupSyncPreviewResponse(result)

		require.Len(t, resp.Owners, 2)
		assert.Equal(t, "1", resp.Owners[0].OwnerID)
		assert.Equal(t, "2", resp.Owners[1].OwnerID)
		assert.Equal(t, 2, resp.TotalCount)
		assert.Equal(t, 1, resp.EligibleCount)
		assert.Equal(t, 2, resp.LineLinkedCount)
	})

	t.Run("empty owners produce empty (not nil) response slice", func(t *testing.T) {
		result := &PreviewCheckupSyncResult{}

		resp := toCheckupSyncPreviewResponse(result)

		assert.NotNil(t, resp.Owners)
		assert.Empty(t, resp.Owners)
		assert.Equal(t, 0, resp.TotalCount)
	})
}

// ---- GetCheckupSyncPreview ----

func TestGetCheckupSyncPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupSyncService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns preview successfully",
			query:    "checkup_type=annual",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			svc: &mockCheckupSyncService{
				previewCheckupSyncFn: func(_ context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "annual", input.CheckupType)
					require.NotNil(t, actorID)
					assert.Equal(t, uint64(10), *actorID)
					return &PreviewCheckupSyncResult{
						Owners:     []CheckupSyncPreviewOwner{{OwnerID: 1, OwnerName: "田中太郎"}},
						TotalCount: 1,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_id":"1"`,
		},
		{
			name:     "returns 401 without invoking service when staff context is missing",
			query:    "checkup_type=annual",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupSyncService{
				previewCheckupSyncFn: func(_ context.Context, _ uint64, _ *PreviewCheckupSyncInput, _ *uint64) (*PreviewCheckupSyncResult, error) {
					t.Fatal("PreviewCheckupSync must not be called after authentication failure")
					return nil, nil
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "checkup_type=annual",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupSyncService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid query parameters",
			query:      "has_chronic_condition=maybe",
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			svc:        &mockCheckupSyncService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "checkup_type=annual",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			svc: &mockCheckupSyncService{
				previewCheckupSyncFn: func(_ context.Context, _ uint64, _ *PreviewCheckupSyncInput, _ *uint64) (*PreviewCheckupSyncResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSyncSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetCheckupSyncPreview(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetCheckupSyncPreview_UsesAuthenticatedClinicInsteadOfURLAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCheckupSyncService{
		previewCheckupSyncFn: func(_ context.Context, clinicID uint64, _ *PreviewCheckupSyncInput, _ *uint64) (*PreviewCheckupSyncResult, error) {
			assert.Equal(t, uint64(1), clinicID)
			return &PreviewCheckupSyncResult{}, nil
		},
	}
	h := newHandlerWithCheckupSyncSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/clinics/999/lstep/checkup-sync/preview", http.NoBody)
	c.Params = gin.Params{{Key: "clinic_id", Value: "999"}}
	setClinicID(c)
	c.Set("user_id", "10")

	h.GetCheckupSyncPreview(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- CreateCheckupSync ----

func TestCreateCheckupSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{
		"checkup_type": "annual",
		"owner_ids":    []string{"1", "2"},
		"tag_name":     "annual_checkup",
	}

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockCheckupSyncService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates checkup sync successfully",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:     validBody,
			svc: &mockCheckupSyncService{
				createCheckupSyncFn: func(_ context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "annual", input.CheckupType)
					assert.Equal(t, []uint64{1, 2}, input.OwnerIDs)
					require.NotNil(t, actorID)
					assert.Equal(t, uint64(10), *actorID)
					return &CreateCheckupSyncResult{
						SuccessCount:   2,
						FailedOwnerIDs: []uint64{},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"success_count":2`,
		},
		{
			name:     "includes failed owner ids as strings",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:     validBody,
			svc: &mockCheckupSyncService{
				createCheckupSyncFn: func(_ context.Context, _ uint64, _ CreateCheckupSyncInput, _ *uint64) (*CreateCheckupSyncResult, error) {
					return &CreateCheckupSyncResult{
						FailedCount:    1,
						FailedOwnerIDs: []uint64{99},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"failed_owner_ids":["99"]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			body:       validBody,
			svc:        &mockCheckupSyncService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 401 without invoking service when staff context is missing",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			body:     validBody,
			svc: &mockCheckupSyncService{
				createCheckupSyncFn: func(_ context.Context, _ uint64, _ CreateCheckupSyncInput, _ *uint64) (*CreateCheckupSyncResult, error) {
					t.Fatal("CreateCheckupSync must not be called after authentication failure")
					return nil, nil
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required field missing",
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:       map[string]any{"tag_name": "annual_checkup"},
			svc:        &mockCheckupSyncService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when tag_name is invalid",
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:       map[string]any{"checkup_type": "annual", "owner_ids": []string{"1"}, "tag_name": "invalid tag!"},
			svc:        &mockCheckupSyncService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "10") },
			body:     validBody,
			svc: &mockCheckupSyncService{
				createCheckupSyncFn: func(_ context.Context, _ uint64, _ CreateCheckupSyncInput, _ *uint64) (*CreateCheckupSyncResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSyncSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateCheckupSync(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestCreateCheckupSync_UsesAuthenticatedClinicInsteadOfURLAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCheckupSyncService{
		createCheckupSyncFn: func(_ context.Context, clinicID uint64, _ CreateCheckupSyncInput, _ *uint64) (*CreateCheckupSyncResult, error) {
			assert.Equal(t, uint64(1), clinicID)
			return &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}, nil
		},
	}
	h := newHandlerWithCheckupSyncSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"checkup_type":"annual","owner_ids":["1"],"tag_name":"annual_checkup"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/clinics/999/lstep/checkup-sync", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "clinic_id", Value: "999"}}
	setClinicID(c)
	c.Set("user_id", "10")

	h.CreateCheckupSync(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
