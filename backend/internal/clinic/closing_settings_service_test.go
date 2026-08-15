package clinic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockClinicSettingsRepository struct {
	ClinicSettingsRepository
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	saveFn           func(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error)
}

func (m *mockClinicSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.ClinicSettings{ClinicID: clinicID}, nil
}

func (m *mockClinicSettingsRepository) Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, s)
	}
	return s, nil
}

type mockClosingSpecialPeriodRepository struct {
	ClosingSpecialPeriodRepository
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error)
	findByDateFn   func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
	createFn       func(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	updateFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	checkOverlapFn func(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
}

func (m *mockClosingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.ClosingSpecialPeriod{}, nil
}

func (m *mockClosingSpecialPeriodRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ClosingSpecialPeriod{ID: id, ClinicID: clinicID}, nil
}

func (m *mockClosingSpecialPeriodRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, nil
}

func (m *mockClosingSpecialPeriodRepository) Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return p, nil
}

func (m *mockClosingSpecialPeriodRepository) CreateCheckingOverlap(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if p == nil {
		return nil, errors.New("period required")
	}
	overlap, err := m.CheckOverlap(ctx, p.ClinicID, p.StartDate, p.EndDate, nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}
	return m.Create(ctx, p)
}

func (m *mockClosingSpecialPeriodRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.ClosingSpecialPeriod{ID: id, ClinicID: clinicID}, nil
}

func (m *mockClosingSpecialPeriodRepository) UpdateCheckingOverlap(ctx context.Context, clinicID, id uint64, startDate, endDate time.Time, fields map[string]any) (*model.ClosingSpecialPeriod, error) {
	excludeID := id
	overlap, err := m.CheckOverlap(ctx, clinicID, startDate, endDate, &excludeID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}
	return m.Update(ctx, clinicID, id, fields)
}

func (m *mockClosingSpecialPeriodRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockClosingSpecialPeriodRepository) CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error) {
	if m.checkOverlapFn != nil {
		return m.checkOverlapFn(ctx, clinicID, startDate, endDate, excludeID)
	}
	return false, nil
}

type mockClosingClinicHolidayRepository struct {
	ClinicHolidayRepository
	findByDateFn         func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
	findAllByYearMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
}

func (m *mockClosingClinicHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, nil
}

func (m *mockClosingClinicHolidayRepository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	if m.findAllByYearMonthFn != nil {
		return m.findAllByYearMonthFn(ctx, clinicID, yearMonth)
	}
	return []model.ClinicHoliday{}, nil
}

func TestClosingSettingsService_ListSpecialPeriods(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockClosingSpecialPeriodRepository{
			findAllFn: func(_ context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
				return []model.ClosingSpecialPeriod{{ID: 1, ClinicID: clinicID}}, nil
			},
		}
		svc := NewClosingSettingsService(nil, repo, nil, nil)
		res, err := svc.ListSpecialPeriods(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockClosingSpecialPeriodRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, repo, nil, nil)
		_, err := svc.ListSpecialPeriods(ctx, 1)
		assert.Error(t, err)
	})
}

func TestClosingSettingsService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{}
		periodRepo := &mockClosingSpecialPeriodRepository{}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		res, err := svc.Get(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, res.Settings)
		assert.Empty(t, res.SpecialPeriods)
	})

	t.Run("settings error", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, nil, nil, nil)
		_, err := svc.Get(ctx, 1)
		assert.Error(t, err)
	})

	t.Run("periods error", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		_, err := svc.Get(ctx, 1)
		assert.Error(t, err)
	})
}

func TestClosingSettingsService_UpdateStandard(t *testing.T) {
	ctx := context.Background()
	const actorID uint64 = 9

	t.Run("success", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "14:00",
					ClosingWeekdayEnd:   "18:30",
					ClosingSundayEnd:    "17:30",
				}, nil
			},
		}
		svc := NewClosingSettingsService(settingsRepo, nil, nil, integrityDeps(&mockClinicRowLocker{}, &recordingAuditTxLogger{}))
		boundary := "12:00"
		weekdayEnd := "19:00"
		sundayEnd := "17:00"
		input := UpdateClinicSettingsInput{
			ClosingAmPmBoundary: &boundary,
			ClosingWeekdayEnd:   &weekdayEnd,
			ClosingSundayEnd:    &sundayEnd,
			ClosedWeekdays:      []int64{0},
		}
		res, err := svc.UpdateStandard(ctx, 1, actorID, input)
		assert.NoError(t, err)
		assert.Equal(t, "12:00", res.ClosingAmPmBoundary)
		assert.Equal(t, "19:00", res.ClosingWeekdayEnd)
		assert.Equal(t, "17:00", res.ClosingSundayEnd)
		assert.Equal(t, pq.Int64Array{0}, res.ClosedWeekdays)
	})

	t.Run("find error", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, nil, nil, integrityDeps(&mockClinicRowLocker{}, &recordingAuditTxLogger{}))
		_, err := svc.UpdateStandard(ctx, 1, actorID, UpdateClinicSettingsInput{})
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "14:00",
					ClosingWeekdayEnd:   "18:30",
					ClosingSundayEnd:    "17:30",
				}, nil
			},
			saveFn: func(_ context.Context, _ uint64, _ *model.ClinicSettings) (*model.ClinicSettings, error) {
				return nil, errors.New("save error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, nil, nil, integrityDeps(&mockClinicRowLocker{}, &recordingAuditTxLogger{}))
		_, err := svc.UpdateStandard(ctx, 1, actorID, UpdateClinicSettingsInput{})
		assert.Error(t, err)
	})
}

func TestClosingSettingsService_CreateSpecialPeriod(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
			Note:         "note",
		}
		res, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("invalid start date", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{StartDate: "bad"}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid end date", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{StartDate: "2026-07-01", EndDate: "bad"}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid boundary time", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "bad",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid pmEnd time", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "bad",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("pmEnd <= boundary", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "11:00",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("start after end", func(t *testing.T) {
		svc := NewClosingSettingsService(nil, nil, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-05",
			EndDate:      "2026-07-01",
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("overlap error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			checkOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return true, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("check overlap db error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			checkOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("create db error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			createFn: func(_ context.Context, _ *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		input := &CreateSpecialPeriodInput{
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-05",
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		_, err := svc.CreateSpecialPeriod(ctx, 1, input)
		assert.Error(t, err)
	})
}

func TestClosingSettingsService_UpdateSpecialPeriod(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		start := "2026-07-01"
		end := "2026-07-05"
		boundary := "13:00"
		pmEnd := "20:00"
		note := "new note"
		input := UpdateSpecialPeriodInput{
			StartDate:    &start,
			EndDate:      &end,
			AmPmBoundary: &boundary,
			PmEnd:        &pmEnd,
			Note:         &note,
		}
		res, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("find error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, UpdateSpecialPeriodInput{})
		assert.Error(t, err)
	})

	t.Run("invalid boundary time", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		bad := "bad"
		input := UpdateSpecialPeriodInput{AmPmBoundary: &bad}
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid start date format", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		bad := "bad"
		input := UpdateSpecialPeriodInput{StartDate: &bad}
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid end date format", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		bad := "bad"
		input := UpdateSpecialPeriodInput{EndDate: &bad}
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.Error(t, err)
	})

	t.Run("start after end", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		start := "2026-07-05"
		end := "2026-07-01"
		input := UpdateSpecialPeriodInput{StartDate: &start, EndDate: &end}
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.Error(t, err)
	})

	t.Run("overlap db error", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
			checkOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		note := "changed"
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, UpdateSpecialPeriodInput{Note: &note})
		assert.Error(t, err)
	})

	t.Run("overlap conflict", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
			checkOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return true, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		note := "changed"
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, UpdateSpecialPeriodInput{Note: &note})
		assert.Error(t, err)
	})

	t.Run("update db error", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		start := "2026-07-01"
		input := UpdateSpecialPeriodInput{StartDate: &start}
		_, err := svc.UpdateSpecialPeriod(ctx, 1, 1, input)
		assert.Error(t, err)
	})

	t.Run("no change", func(t *testing.T) {
		current := &model.ClosingSpecialPeriod{
			ID:           1,
			ClinicID:     1,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 5),
			AmPmBoundary: "12:00",
			PmEnd:        "19:00",
		}
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return current, nil
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		res, err := svc.UpdateSpecialPeriod(ctx, 1, 1, UpdateSpecialPeriodInput{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestClosingSettingsService_DeleteSpecialPeriod(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		err := svc.DeleteSpecialPeriod(ctx, 1, 1)
		assert.NoError(t, err)
	})

	t.Run("find error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		err := svc.DeleteSpecialPeriod(ctx, 1, 1)
		assert.Error(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		err := svc.DeleteSpecialPeriod(ctx, 1, 1)
		assert.Error(t, err)
	})
}

func TestClosingSettingsService_ResolveSchedule(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 7, 5, 0, 0, 0, 0, time.Local) // 2026-07-05 (日曜)

	t.Run("special period matches (highest priority)", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return &model.ClosingSpecialPeriod{
					AmPmBoundary: "13:00",
					PmEnd:        "20:00",
				}, nil
			},
		}
		// #215: 特別期間は am_start を持たず標準設定から継承するため settingsRepo が必要
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{ClinicID: clinicID, ClosingAmStart: "08:00"}, nil
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		res, err := svc.ResolveSchedule(ctx, 1, date)
		assert.NoError(t, err)
		assert.Equal(t, "13:00", res.AmPmBoundary)
		assert.Equal(t, "20:00", res.PmEnd)
		assert.Equal(t, "08:00", res.AmStart, "特別期間でも am_start は標準設定から継承する（#215）")
		assert.False(t, res.IsHoliday)
	})

	t.Run("special period: settings fetch error propagates (#215)", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return &model.ClosingSpecialPeriod{AmPmBoundary: "13:00", PmEnd: "20:00"}, nil
			},
		}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		_, err := svc.ResolveSchedule(ctx, 1, date)
		assert.Error(t, err)
	})

	t.Run("special period find db error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{
			findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(nil, periodRepo, nil, nil)
		_, err := svc.ResolveSchedule(ctx, 1, date)
		assert.Error(t, err)
	})

	t.Run("standard settings get db error", func(t *testing.T) {
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		_, err := svc.ResolveSchedule(ctx, 1, date)
		assert.Error(t, err)
	})

	t.Run("standard weekday scheduling", func(t *testing.T) {
		weekdayDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local) // 2026-07-01 (水曜)
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "12:00",
					ClosingWeekdayEnd:   "19:00",
					ClosingSundayEnd:    "17:00",
					ClosedWeekdays:      []int64{0}, // 日曜休み
				}, nil
			},
		}
		holidayRepo := &mockClosingClinicHolidayRepository{}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo, nil)
		res, err := svc.ResolveSchedule(ctx, 1, weekdayDate)
		assert.NoError(t, err)
		assert.Equal(t, "19:00", res.PmEnd)
		assert.Equal(t, "09:00", res.AmStart, "am_start 未設定の旧データは既定 09:00 にフォールバック（#215）")
		assert.False(t, res.IsHoliday)
	})

	t.Run("standard sunday scheduling", func(t *testing.T) {
		sundayDate := time.Date(2026, 7, 5, 0, 0, 0, 0, time.Local) // 2026-07-05 (日曜)
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "12:00",
					ClosingWeekdayEnd:   "19:00",
					ClosingSundayEnd:    "17:00",
					ClosedWeekdays:      []int64{1}, // 月曜休み
				}, nil
			},
		}
		holidayRepo := &mockClosingClinicHolidayRepository{}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo, nil)
		res, err := svc.ResolveSchedule(ctx, 1, sundayDate)
		assert.NoError(t, err)
		assert.Equal(t, "17:00", res.PmEnd)
		assert.False(t, res.IsHoliday)
	})

	t.Run("weekly closed weekday match", func(t *testing.T) {
		sundayDate := time.Date(2026, 7, 5, 0, 0, 0, 0, time.Local) // 2026-07-05 (日曜)
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "12:00",
					ClosingWeekdayEnd:   "19:00",
					ClosingSundayEnd:    "17:00",
					ClosedWeekdays:      []int64{0}, // 日曜休み
				}, nil
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, nil, nil)
		res, err := svc.ResolveSchedule(ctx, 1, sundayDate)
		assert.NoError(t, err)
		assert.True(t, res.IsHoliday)
	})

	t.Run("custom holiday match", func(t *testing.T) {
		weekdayDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local) // 2026-07-01 (水曜)
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "12:00",
					ClosingWeekdayEnd:   "19:00",
					ClosingSundayEnd:    "17:00",
					ClosedWeekdays:      []int64{0},
				}, nil
			},
		}
		holidayRepo := &mockClosingClinicHolidayRepository{
			findAllByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{
					{
						Date: weekdayDate,
					},
				}, nil
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo, nil)
		res, err := svc.ResolveSchedule(ctx, 1, weekdayDate)
		assert.NoError(t, err)
		assert.True(t, res.IsHoliday)
	})

	t.Run("custom holiday load db error", func(t *testing.T) {
		weekdayDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local) // 2026-07-01 (水曜)
		periodRepo := &mockClosingSpecialPeriodRepository{}
		settingsRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					ClinicID:            clinicID,
					ClosingAmPmBoundary: "12:00",
					ClosingWeekdayEnd:   "19:00",
					ClosingSundayEnd:    "17:00",
					ClosedWeekdays:      []int64{0},
				}, nil
			},
		}
		holidayRepo := &mockClosingClinicHolidayRepository{
			findAllByYearMonthFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo, nil)
		_, err := svc.ResolveSchedule(ctx, 1, weekdayDate)
		assert.Error(t, err)
	})
}
