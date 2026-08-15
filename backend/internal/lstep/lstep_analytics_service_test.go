package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- LstepDeliveryTriggerLogRepository モック（analytics 用）----

type mockAnalyticsTriggerLogRepo struct {
	countByTypeAndStatusFn func(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error)
	countVisitConvByTypeFn func(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error)
}

func (m *mockAnalyticsTriggerLogRepo) Create(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
	return nil
}
func (m *mockAnalyticsTriggerLogRepo) CreateIfAbsentToday(_ context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if log != nil && log.ID == 0 {
		log.ID = 1
	}
	return true, nil
}
func (m *mockAnalyticsTriggerLogRepo) ExistsTodayByOwnerAndType(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (m *mockAnalyticsTriggerLogRepo) UpdateStatus(_ context.Context, _, _ uint64, _ string, _ *time.Time, _ *string) error {
	return nil
}
func (m *mockAnalyticsTriggerLogRepo) FindByClinicAndDate(_ context.Context, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockAnalyticsTriggerLogRepo) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockAnalyticsTriggerLogRepo) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockAnalyticsTriggerLogRepo) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}
func (m *mockAnalyticsTriggerLogRepo) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
	return nil, 0, nil
}
func (m *mockAnalyticsTriggerLogRepo) CountByTypeAndStatus(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error) {
	if m.countByTypeAndStatusFn != nil {
		return m.countByTypeAndStatusFn(ctx, clinicID, from, to)
	}
	return nil, nil
}
func (m *mockAnalyticsTriggerLogRepo) CountVisitConversionsByType(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error) {
	if m.countVisitConvByTypeFn != nil {
		return m.countVisitConvByTypeFn(ctx, clinicID, from, to, days)
	}
	return nil, nil
}
func (m *mockAnalyticsTriggerLogRepo) FindByOwnerAndDate(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockAnalyticsTriggerLogRepo) UpdateSuppressed(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

// ---- LstepFriendAttributeSnapshotRepository モック（analytics 用）----

type mockAnalyticsSnapshotRepo struct {
	findLatestFn func(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error)
}

func (m *mockAnalyticsSnapshotRepo) Create(_ context.Context, _ *model.LstepFriendAttributeSnapshot) error {
	return nil
}
func (m *mockAnalyticsSnapshotRepo) FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error) {
	if m.findLatestFn != nil {
		return m.findLatestFn(ctx, clinicID, lineUserID)
	}
	return nil, nil
}
func (m *mockAnalyticsSnapshotRepo) FindAllByClinicAndDateRange(_ context.Context, _ uint64, _, _ time.Time) ([]*model.LstepFriendAttributeSnapshot, error) {
	return nil, nil
}

// ---- OwnerRepository モック（analytics 用）----

type mockAnalyticsOwnerRepo struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

func (m *mockAnalyticsOwnerRepo) FindAll(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}
func (m *mockAnalyticsOwnerRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindByLineUserID(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindAllWithLineUserID(_ context.Context, _ uint64) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) FindAllWithLineUserIDCursor(_ context.Context, _, _ uint64, _ int) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockAnalyticsOwnerRepo) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}
func (m *mockAnalyticsOwnerRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockAnalyticsOwnerRepo) UpdateLineUserID(_ context.Context, _, _ uint64, _ *string) error {
	return nil
}
func (m *mockAnalyticsOwnerRepo) UpdateLineFollowedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}
func (m *mockAnalyticsOwnerRepo) UpdateLineBlockedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}
func (m *mockAnalyticsOwnerRepo) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockAnalyticsOwnerRepo) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockAnalyticsOwnerRepo) FindByIDs(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
	return nil, nil
}

// ---- GetMonthlyDeliveryStats tests ----

func TestLstepAnalyticsService_GetMonthlyDeliveryStats(t *testing.T) {
	t.Run("yearMonth パース + JST 月境界で repo 呼び出し", func(t *testing.T) {
		expectedFrom := time.Date(2026, 5, 1, 0, 0, 0, 0, config.JST)
		expectedUntil := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST)

		triggerRepo := &mockAnalyticsTriggerLogRepo{
			countByTypeAndStatusFn: func(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error) {
				assert.True(t, from.Equal(expectedFrom), "from mismatch: got %v, want %v", from, expectedFrom)
				assert.True(t, to.Equal(expectedUntil), "to mismatch: got %v, want %v", to, expectedUntil)
				return []DeliveryStatsRow{
					{TriggerType: "TriggerBirthdayMessage", Status: "fired", Count: 10},
					{TriggerType: "TriggerBirthdayMessage", Status: "excluded", Count: 2},
					{TriggerType: "TriggerVaccineDeadline60", Status: "fired", Count: 5},
				}, nil
			},
		}
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, triggerRepo, &mockAnalyticsSnapshotRepo{})

		stats, err := svc.GetMonthlyDeliveryStats(context.Background(), 100, "2026-05")

		assert.NoError(t, err)
		assert.Equal(t, "2026-05", stats.YearMonth)
		assert.Len(t, stats.Rows, 3)
	})

	t.Run("yearMonth 不正フォーマット → InvalidInput", func(t *testing.T) {
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetMonthlyDeliveryStats(context.Background(), 100, "invalid")

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("空文字 yearMonth → InvalidInput", func(t *testing.T) {
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetMonthlyDeliveryStats(context.Background(), 100, "")

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("repo error は wrap されて返る", func(t *testing.T) {
		triggerRepo := &mockAnalyticsTriggerLogRepo{
			countByTypeAndStatusFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, triggerRepo, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetMonthlyDeliveryStats(context.Background(), 100, "2026-05")

		assert.Error(t, err)
	})
}

// ---- GetLatestFriendAttributes tests ----

func TestLstepAnalyticsService_GetLatestFriendAttributes(t *testing.T) {
	t.Run("owner.LineUserID 取得成功 → snapshot 返却", func(t *testing.T) {
		lu := "U1234"
		ownerRepo := &mockAnalyticsOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &lu}, nil
			},
		}
		snapshotRepo := &mockAnalyticsSnapshotRepo{
			findLatestFn: func(_ context.Context, _ uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error) {
				assert.Equal(t, "U1234", lineUserID)
				return &model.LstepFriendAttributeSnapshot{ID: 1, LineUserID: lineUserID}, nil
			},
		}
		svc := NewLstepAnalyticsService(ownerRepo, &mockAnalyticsTriggerLogRepo{}, snapshotRepo)

		got, err := svc.GetLatestFriendAttributes(context.Background(), 100, 200)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, uint64(1), got.ID)
	})

	t.Run("owner.LineUserID nil → NotFound", func(t *testing.T) {
		ownerRepo := &mockAnalyticsOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: nil}, nil
			},
		}
		svc := NewLstepAnalyticsService(ownerRepo, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetLatestFriendAttributes(context.Background(), 100, 200)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("owner.LineUserID 空文字 → NotFound", func(t *testing.T) {
		empty := ""
		ownerRepo := &mockAnalyticsOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &empty}, nil
			},
		}
		svc := NewLstepAnalyticsService(ownerRepo, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetLatestFriendAttributes(context.Background(), 100, 200)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("owner repo error は wrap されて返る", func(t *testing.T) {
		ownerRepo := &mockAnalyticsOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepAnalyticsService(ownerRepo, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetLatestFriendAttributes(context.Background(), 100, 200)

		assert.Error(t, err)
	})

	t.Run("snapshotRepo error は wrap されて返る", func(t *testing.T) {
		lu := "U9999"
		ownerRepo := &mockAnalyticsOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &lu}, nil
			},
		}
		snapshotRepo := &mockAnalyticsSnapshotRepo{
			findLatestFn: func(_ context.Context, _ uint64, _ string) (*model.LstepFriendAttributeSnapshot, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepAnalyticsService(ownerRepo, &mockAnalyticsTriggerLogRepo{}, snapshotRepo)

		got, err := svc.GetLatestFriendAttributes(context.Background(), 100, 200)

		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

// ---- calculateRate direct tests ----

func TestCalculateRate(t *testing.T) {
	tests := []struct {
		name        string
		numerator   int64
		denominator int64
		want        float64
	}{
		{name: "denominator zero returns zero", numerator: 5, denominator: 0, want: 0},
		{name: "denominator zero with zero numerator returns zero", numerator: 0, denominator: 0, want: 0},
		{name: "normal ratio", numerator: 1, denominator: 4, want: 0.25},
		{name: "full ratio", numerator: 10, denominator: 10, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRate(tt.numerator, tt.denominator)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestLstepAnalyticsService_GetVisitConversionSummary(t *testing.T) {
	t.Run("yearMonth + days を使って repo 集計を呼び、率を計算する", func(t *testing.T) {
		expectedFrom := time.Date(2026, 5, 1, 0, 0, 0, 0, config.JST)
		expectedUntil := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST)
		triggerRepo := &mockAnalyticsTriggerLogRepo{
			countVisitConvByTypeFn: func(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error) {
				assert.Equal(t, uint64(100), clinicID)
				assert.Equal(t, 30, days)
				assert.True(t, from.Equal(expectedFrom))
				assert.True(t, to.Equal(expectedUntil))
				return []VisitConversionRow{
					{TriggerType: "birthday_message", DeliveredCount: 10, VisitedCount: 4},
					{TriggerType: "next_visit_reminder", DeliveredCount: 5, VisitedCount: 1},
				}, nil
			},
		}
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, triggerRepo, &mockAnalyticsSnapshotRepo{})

		stats, err := svc.GetVisitConversionSummary(context.Background(), 100, "2026-05", 30)

		assert.NoError(t, err)
		assert.Equal(t, "2026-05", stats.YearMonth)
		assert.Equal(t, 30, stats.Days)
		assert.Equal(t, int64(15), stats.DeliveredCount)
		assert.Equal(t, int64(5), stats.VisitedCount)
		assert.InDelta(t, 5.0/15.0, stats.VisitRate, 0.0001)
		assert.Len(t, stats.Rows, 2)
		assert.InDelta(t, 0.4, stats.Rows[0].VisitRate, 0.0001)
	})

	t.Run("days が 0 以下なら InvalidInput", func(t *testing.T) {
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetVisitConversionSummary(context.Background(), 100, "2026-05", 0)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("days が上限を超えたら InvalidInput", func(t *testing.T) {
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, &mockAnalyticsTriggerLogRepo{}, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetVisitConversionSummary(context.Background(), 100, "2026-05", 366)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("repo error は wrap されて返る", func(t *testing.T) {
		triggerRepo := &mockAnalyticsTriggerLogRepo{
			countVisitConvByTypeFn: func(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepAnalyticsService(&mockAnalyticsOwnerRepo{}, triggerRepo, &mockAnalyticsSnapshotRepo{})

		_, err := svc.GetVisitConversionSummary(context.Background(), 100, "2026-05", 30)

		assert.Error(t, err)
	})
}
