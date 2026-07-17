package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepTagSummaryService ----

type mockLstepTagSummaryService struct {
	getTagSummaryFn     func(ctx context.Context, clinicID uint64) (service.TagSummaryResponse, error)
	listOwnersByTagFn   func(ctx context.Context, clinicID uint64, input service.ListOwnersByTagInput) (service.TagOwnerListResponse, error)
	exportOwnersByTagFn func(ctx context.Context, clinicID uint64, tagName, nameQuery string, w io.Writer) error
}

func (m *mockLstepTagSummaryService) GetTagSummary(ctx context.Context, clinicID uint64) (service.TagSummaryResponse, error) {
	if m.getTagSummaryFn != nil {
		return m.getTagSummaryFn(ctx, clinicID)
	}
	return service.TagSummaryResponse{}, nil
}
func (m *mockLstepTagSummaryService) ListOwnersByTag(ctx context.Context, clinicID uint64, input service.ListOwnersByTagInput) (service.TagOwnerListResponse, error) {
	if m.listOwnersByTagFn != nil {
		return m.listOwnersByTagFn(ctx, clinicID, input)
	}
	return service.TagOwnerListResponse{}, nil
}
func (m *mockLstepTagSummaryService) ExportOwnersByTagCSV(ctx context.Context, clinicID uint64, tagName, nameQuery string, w io.Writer) error {
	if m.exportOwnersByTagFn != nil {
		return m.exportOwnersByTagFn(ctx, clinicID, tagName, nameQuery, w)
	}
	_, _ = w.Write([]byte("owner_id,owner_name\n1,田中 太郎\n"))
	return nil
}

// ---- helpers ----

func ptrString(s string) *string { return &s }

func newHandlerWithLstepTagSummarySvc(svc service.LstepTagSummaryService) *Handler {
	return &Handler{svc: &service.Services{LstepTagSummary: svc}}
}

func newGetLstepTagSummaryRouter(svc service.LstepTagSummaryService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepTagSummarySvc(svc)
	if withClinicID {
		r.GET("/lstep/tag-summary", func(c *gin.Context) { setClinicID(c) }, h.GetLstepTagSummary)
	} else {
		r.GET("/lstep/tag-summary", h.GetLstepTagSummary)
	}
	return r
}

func newSearchLstepOwnersByTagRouter(svc service.LstepTagSummaryService, withClinicID bool) *gin.Engine {
	r := gin.New()
	h := newHandlerWithLstepTagSummarySvc(svc)
	if withClinicID {
		r.GET("/lstep/owners", func(c *gin.Context) { setClinicID(c) }, h.SearchLstepOwnersByTag)
	} else {
		r.GET("/lstep/owners", h.SearchLstepOwnersByTag)
	}
	return r
}

// ---- tests ----

func TestGetLstepTagSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		svc        *mockLstepTagSummaryService
		wantStatus int
	}{
		{
			name:       "200 success",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 no clinic",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 service error",
			svc: &mockLstepTagSummaryService{
				getTagSummaryFn: func(_ context.Context, _ uint64) (service.TagSummaryResponse, error) {
					return service.TagSummaryResponse{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newGetLstepTagSummaryRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodGet, "/lstep/tag-summary", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestSearchLstepOwnersByTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		query      string
		svc        *mockLstepTagSummaryService
		wantStatus int
	}{
		{
			name:       "200 success with tag",
			query:      "?tag=my_tag",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 missing tag param",
			query:      "",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 bad page value",
			query:      "?tag=my_tag&page=0",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 per_page too large",
			query:      "?tag=my_tag&per_page=200",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "401 no clinic",
			query:      "?tag=my_tag",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "200 CSV format",
			query:      "?tag=my_tag&format=csv",
			svc:        &mockLstepTagSummaryService{},
			wantStatus: http.StatusOK,
		},
		{
			name:  "500 service error",
			query: "?tag=my_tag",
			svc: &mockLstepTagSummaryService{
				listOwnersByTagFn: func(_ context.Context, _ uint64, _ service.ListOwnersByTagInput) (service.TagOwnerListResponse, error) {
					return service.TagOwnerListResponse{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newSearchLstepOwnersByTagRouter(tt.svc, tt.wantStatus != http.StatusUnauthorized)
			req := httptest.NewRequest(http.MethodGet, "/lstep/owners"+tt.query, http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// #179 ③: 日本語タグ名を含む CSV のファイル名が RFC 5987 でエンコードされ、
// Content-Disposition で文字化けしないことを検証する。
func TestSearchLstepOwnersByTag_CSVFilenameRFC5987(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const tag = "優良顧客"
	router := newSearchLstepOwnersByTagRouter(&mockLstepTagSummaryService{}, true)
	req := httptest.NewRequest(
		http.MethodGet,
		"/lstep/owners?tag="+url.QueryEscape(tag)+"&format=csv",
		http.NoBody,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cd := w.Header().Get("Content-Disposition")
	// RFC 5987 の filename*（UTF-8）が併記される
	assert.Contains(t, cd, "filename*=UTF-8''")
	// 日本語タグ名はパーセントエンコードされて含まれる（生の UTF-8 を直書きしない）
	assert.Contains(t, cd, url.PathEscape(tag))
	// 非対応クライアント向け ASCII フォールバックも併記する
	assert.Contains(t, cd, `filename="lstep-`)
}

func TestSearchLstepOwnersByTag_ReasonField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name          string
		mockReason    *string
		wantReasonKey bool
		wantReasonVal string
	}{
		{
			name:          "reason set → JSON に reason 含む",
			mockReason:    ptrString("最終健診: 2025-12-01"),
			wantReasonKey: true,
			wantReasonVal: "最終健診: 2025-12-01",
		},
		{
			name:          "reason nil → JSON から omitempty で省略",
			mockReason:    nil,
			wantReasonKey: false,
		},
		{
			name:          "reason empty string pointer → JSON に空文字含む",
			mockReason:    ptrString(""),
			wantReasonKey: true,
			wantReasonVal: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockLstepTagSummaryService{
				listOwnersByTagFn: func(_ context.Context, _ uint64, _ service.ListOwnersByTagInput) (service.TagOwnerListResponse, error) {
					return service.TagOwnerListResponse{
						Owners: []service.TagOwnerItem{
							{
								OwnerID:   1,
								OwnerName: "田中 太郎",
								Reason:    tt.mockReason,
							},
						},
						Total:   1,
						Page:    1,
						PerPage: 20,
					}, nil
				},
			}
			router := newSearchLstepOwnersByTagRouter(svc, true)
			req := httptest.NewRequest(http.MethodGet, "/lstep/owners?tag=my_tag", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			var body map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &body)
			assert.NoError(t, err)
			owners, ok := body["owners"].([]any)
			assert.True(t, ok)
			assert.Len(t, owners, 1)
			owner := owners[0].(map[string]any)

			if tt.wantReasonKey {
				got, exists := owner["reason"]
				assert.True(t, exists, "reason key should be present")
				assert.Equal(t, tt.wantReasonVal, got)
			} else {
				_, exists := owner["reason"]
				assert.False(t, exists, "reason key should be omitted (omitempty)")
			}
		})
	}
}
