package pet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPetService_GetFirstVisitDate は #158 初診日導出のサービス委譲を検証する。
// repository が返す MIN(date) をそのまま返し、エラーはラップ、無ければ nil を返す。
func TestPetService_GetFirstVisitDate(t *testing.T) {
	ctx := context.Background()
	firstVisit := time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)

	t.Run("repository の初診日をそのまま返す", func(t *testing.T) {
		mrRepo := &mockMedicalRecordRepository{
			findFirstVisitDateByPetIDFn: func(_ context.Context, clinicID, petID uint64) (*time.Time, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), petID)
				return &firstVisit, nil
			},
		}
		svc := newPetSvc(&mockPetRepository{}, defaultOwnerRepo(), defaultInsuranceRepo(1), mrRepo)

		got, err := svc.GetFirstVisitDate(ctx, 1, 7)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, firstVisit, *got)
	})

	t.Run("カルテが無ければ nil を返す（捏造しない）", func(t *testing.T) {
		svc := newPetSvc(&mockPetRepository{}, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo())

		got, err := svc.GetFirstVisitDate(ctx, 1, 7)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("repository エラーはラップして返す", func(t *testing.T) {
		mrRepo := &mockMedicalRecordRepository{
			findFirstVisitDateByPetIDFn: func(_ context.Context, _, _ uint64) (*time.Time, error) {
				return nil, errors.New("db down")
			},
		}
		svc := newPetSvc(&mockPetRepository{}, defaultOwnerRepo(), defaultInsuranceRepo(1), mrRepo)

		_, err := svc.GetFirstVisitDate(ctx, 1, 7)
		require.Error(t, err)
	})
}
