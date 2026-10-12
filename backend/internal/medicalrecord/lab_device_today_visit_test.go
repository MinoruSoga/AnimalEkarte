package medicalrecord

// lab_device_today_visit_test.go — wait board の死亡フラグが OR 契約に従うことの検証（EMR-328）。
// status='alive' + deceased_at≠NULL の半死亡行も死亡として表示される必要がある。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestLabDeviceTodayVisitFinder_PetIsDeceasedFollowsOrContract(t *testing.T) {
	deceasedAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	petAlive := uint64(11)
	petHalfDead := uint64(12)
	petStatusOnly := uint64(13)

	records := []model.MedicalRecord{
		{
			ID:    1,
			PetID: &petAlive,
			Pet:   &model.Pet{ID: petAlive, Name: "生存", Status: model.PetStatusAlive},
		},
		{
			ID:    2,
			PetID: &petHalfDead,
			Pet: &model.Pet{
				ID:         petHalfDead,
				Name:       "半死亡",
				Status:     model.PetStatusAlive,
				DeceasedAt: &deceasedAt,
			},
		},
		{
			ID:    3,
			PetID: &petStatusOnly,
			Pet:   &model.Pet{ID: petStatusOnly, Name: "旧死亡", Status: model.PetStatusDeceased},
		},
	}
	repo := &mockMedicalRecordRepository{
		findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
			return records, int64(len(records)), nil
		},
	}

	finder := NewLabDeviceTodayVisitFinder(repo)
	visits, err := finder.ListTodayDraft(context.Background(), 1, "2026-10-07")
	require.NoError(t, err)
	require.Len(t, visits, 3)

	assert.False(t, visits[0].PetIsDeceased, "status=alive + deceased_at NULL must be shown alive")
	assert.True(t, visits[1].PetIsDeceased, "status=alive + deceased_at≠NULL must be shown deceased (OR contract)")
	assert.True(t, visits[2].PetIsDeceased, "status=deceased + deceased_at NULL must remain deceased")
}
