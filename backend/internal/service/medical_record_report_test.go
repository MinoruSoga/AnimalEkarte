package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// TestMedicalRecordService_GetOwnerMedicationHistory は #158 のサービス層を検証する。
// 期待:
//   - 飼い主の所属医院を先に検証し、他院/不存在 ID は NotFound（repository 未照会）。
//   - 所属確認後、repository に clinicID/ownerID/page/limit を委譲し結果を DTO へ変換して返す。
//   - repository エラーは apperrors.Wrap で包んで伝播する（握りつぶさない）。
func TestMedicalRecordService_GetOwnerMedicationHistory(t *testing.T) {
	okOwner := &mrMockOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{}, nil
		},
	}

	t.Run("delegates args and returns rows + total", func(t *testing.T) {
		var gotClinic, gotOwner uint64
		var gotPage, gotLimit int
		repo := &mockMedicalRecordRepository{
			findOwnerMedicationHistoryFn: func(_ context.Context, clinicID, ownerID uint64, page, limit int) ([]repository.OwnerMedicationHistoryRow, int64, error) {
				gotClinic, gotOwner, gotPage, gotLimit = clinicID, ownerID, page, limit
				return []repository.OwnerMedicationHistoryRow{
					{TreatmentID: 11, PetName: "ポチ", MedicineName: "アモキシシリン"},
				}, 7, nil
			},
		}
		svc := NewMedicalRecordService(repo, okOwner, nil, nil, nil, nil, nil, nil, nil)

		rows, total, err := svc.GetOwnerMedicationHistory(context.Background(), 3, 42, 2, 10)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), gotClinic)
		assert.Equal(t, uint64(42), gotOwner)
		assert.Equal(t, 2, gotPage)
		assert.Equal(t, 10, gotLimit)
		assert.Equal(t, int64(7), total)
		require.Len(t, rows, 1)
		assert.Equal(t, "アモキシシリン", rows[0].MedicineName)
	})

	t.Run("returns not-found and skips query when owner is outside the clinic", func(t *testing.T) {
		called := false
		repo := &mockMedicalRecordRepository{
			findOwnerMedicationHistoryFn: func(_ context.Context, _, _ uint64, _, _ int) ([]repository.OwnerMedicationHistoryRow, int64, error) {
				called = true
				return nil, 0, nil
			},
		}
		notFoundOwner := &mrMockOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("owner not found")
			},
		}
		svc := NewMedicalRecordService(repo, notFoundOwner, nil, nil, nil, nil, nil, nil, nil)

		rows, total, err := svc.GetOwnerMedicationHistory(context.Background(), 1, 999, 1, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Equal(t, int64(0), total)
		assert.False(t, called, "repository must not be queried when owner is not in the clinic")
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findOwnerMedicationHistoryFn: func(_ context.Context, _, _ uint64, _, _ int) ([]repository.OwnerMedicationHistoryRow, int64, error) {
				return nil, 0, errors.New("db down")
			},
		}
		svc := NewMedicalRecordService(repo, okOwner, nil, nil, nil, nil, nil, nil, nil)

		rows, total, err := svc.GetOwnerMedicationHistory(context.Background(), 1, 1, 1, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Equal(t, int64(0), total)
	})
}
