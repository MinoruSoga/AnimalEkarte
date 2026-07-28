package trimming

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToTrimmingCourseTypeResponse(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)

	m := &model.TrimmingCourseType{
		ID:        7,
		ClinicID:  1,
		Name:      "シャンプーコース",
		SortOrder: 2,
		IsActive:  true,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	got := toTrimmingCourseTypeResponse(m)

	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
	if got.ClinicID != 1 {
		t.Errorf("ClinicID = %d, want 1", got.ClinicID)
	}
	if got.Name != "シャンプーコース" {
		t.Errorf("Name = %q, want シャンプーコース", got.Name)
	}
	if got.SortOrder != 2 {
		t.Errorf("SortOrder = %d, want 2", got.SortOrder)
	}
	if !got.IsActive {
		t.Errorf("IsActive = false, want true")
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updatedAt)
	}
}

func TestToTrimmingCourseTypeResponse_ZeroValue(t *testing.T) {
	m := &model.TrimmingCourseType{}

	got := toTrimmingCourseTypeResponse(m)

	if got.ID != 0 {
		t.Errorf("ID = %d, want 0", got.ID)
	}
	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}
	if got.IsActive {
		t.Errorf("IsActive = true, want false")
	}
	if !got.CreatedAt.IsZero() {
		t.Errorf("CreatedAt = %v, want zero", got.CreatedAt)
	}
}

func TestToTrimmingCourseTypeResponse_InactiveWithHighSortOrder(t *testing.T) {
	m := &model.TrimmingCourseType{
		ID:        99,
		Name:      "廃止コース",
		SortOrder: 999,
		IsActive:  false,
	}

	got := toTrimmingCourseTypeResponse(m)

	if got.IsActive {
		t.Errorf("IsActive = true, want false")
	}
	if got.SortOrder != 999 {
		t.Errorf("SortOrder = %d, want 999", got.SortOrder)
	}
}
