package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaginatedResponse(t *testing.T) {
	data := []int{1, 2, 3}
	got := NewPaginatedResponse(data, 42, 2, 10)

	assert.Equal(t, data, got.Data)
	assert.Equal(t, int64(42), got.Total)
	assert.Equal(t, 2, got.Page)
	assert.Equal(t, 10, got.Limit)
}
