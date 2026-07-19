package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestParsePagination はここに残す：query_helpers.go は target:httpapi ではあるが
// BE9-2B（manualarticle pilot）が使わないため未移動（docs/architecture/be9-2a-boundary-map.md
// 参照。他の target:httpapi 5ファイルのみ internal/httpapi へ移動済み）。response_test.go
// 分割時に置き場を失わないよう、query_helpers.go と同じ internal/handler package に残す。
func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
		wantErr   bool
	}{
		{
			name:      "defaults return page 1 and limit 20",
			query:     "",
			wantPage:  1,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "custom valid values are accepted",
			query:     "page=2&limit=50",
			wantPage:  2,
			wantLimit: 50,
			wantErr:   false,
		},
		{
			name:      "limit of 100 is accepted",
			query:     "page=1&limit=100",
			wantPage:  1,
			wantLimit: 100,
			wantErr:   false,
		},
		{
			name:    "page zero is invalid",
			query:   "page=0",
			wantErr: true,
		},
		{
			name:    "negative page is invalid",
			query:   "page=-1",
			wantErr: true,
		},
		{
			name:    "limit over 100 is invalid",
			query:   "limit=101",
			wantErr: true,
		},
		{
			name:    "limit zero is invalid",
			query:   "limit=0",
			wantErr: true,
		},
		{
			name:    "non-numeric page is invalid",
			query:   "page=abc",
			wantErr: true,
		},
		{
			name:    "non-numeric limit is invalid",
			query:   "limit=xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/?"+tt.query, http.NoBody)

			page, limit, err := parsePagination(c)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, page)
				assert.Equal(t, 0, limit)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPage, page)
				assert.Equal(t, tt.wantLimit, limit)
			}
		})
	}
}
