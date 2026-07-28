package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestParsePagination ports internal/handler/query_helpers_test.go's coverage for the
// now-moved ParsePagination (BE9-2C query_helpers.go target:httpapi extraction).
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
		{
			name:      "per_page is an alias for limit",
			query:     "per_page=30",
			wantPage:  1,
			wantLimit: 30,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/?"+tt.query, http.NoBody)

			page, limit, err := ParsePagination(c)

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

func TestParsePaginationWithMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?limit=500", http.NoBody)

	page, limit, err := ParsePaginationWithMax(c, 1000)

	assert.NoError(t, err)
	assert.Equal(t, 1, page)
	assert.Equal(t, 500, limit)
}

func TestParseIDParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtxWithParam := func(key, value string) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)
		c.Params = gin.Params{{Key: key, Value: value}}
		return c
	}

	t.Run("valid id parses", func(t *testing.T) {
		c := newCtxWithParam("id", "42")
		id, ok := ParseIDParam(c, "id")
		assert.True(t, ok)
		assert.Equal(t, uint64(42), id)
	})

	t.Run("missing param is rejected", func(t *testing.T) {
		c := newCtxWithParam("id", "")
		id, ok := ParseIDParam(c, "id")
		assert.False(t, ok)
		assert.Equal(t, uint64(0), id)
		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("non-numeric param is rejected", func(t *testing.T) {
		c := newCtxWithParam("id", "abc")
		id, ok := ParseIDParam(c, "id")
		assert.False(t, ok)
		assert.Equal(t, uint64(0), id)
	})

	t.Run("zero is rejected", func(t *testing.T) {
		c := newCtxWithParam("id", "0")
		id, ok := ParseIDParam(c, "id")
		assert.False(t, ok)
		assert.Equal(t, uint64(0), id)
	})
}

func TestParseOptionalUint64Query(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("absent query returns nil, true", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		id, ok := ParseOptionalUint64Query(c, "type_id")
		assert.True(t, ok)
		assert.Nil(t, id)
	})

	t.Run("valid query parses", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/?type_id=7", http.NoBody)

		id, ok := ParseOptionalUint64Query(c, "type_id")
		assert.True(t, ok)
		if assert.NotNil(t, id) {
			assert.Equal(t, uint64(7), *id)
		}
	})

	t.Run("malformed query is rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/?type_id=abc", http.NoBody)

		id, ok := ParseOptionalUint64Query(c, "type_id")
		assert.False(t, ok)
		assert.Nil(t, id)
		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestParseUUIDParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid uuid parses", func(t *testing.T) {
		want := uuid.New()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: want.String()}}

		got, ok := ParseUUIDParam(c, "id")
		assert.True(t, ok)
		assert.Equal(t, want, got)
	})

	t.Run("malformed uuid is rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

		got, ok := ParseUUIDParam(c, "id")
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, got)
	})
}
