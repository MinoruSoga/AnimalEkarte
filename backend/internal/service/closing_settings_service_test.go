package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- モックリポジトリ定義 ----

type mockClinicSettingsRepository struct {
	findByClinicIDFn         func(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	upsertFn                 func(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error)
	updateLstepFireHourJSTFn func(ctx context.Context, clinicID uint64, hour int) error
}

func (m *mockClinicSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.ClinicSettings{}, nil
}

func (m *mockClinicSettingsRepository) Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, clinicID, s)
	}
	return s, nil
}
func (m *mockClinicSettingsRepository) UpdateLstepFireHourJST(ctx context.Context, clinicID uint64, hour int) error {
	if m.updateLstepFireHourJSTFn != nil {
		return m.updateLstepFireHourJSTFn(ctx, clinicID, hour)
	}
	return nil
}

type mockClosingSpecialPeriodRepository struct {
	findAllFn    func(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	findByIDFn   func(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error)
	findByDateFn func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
	createFn     func(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	updateFn     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error)
	deleteFn     func(ctx context.Context, clinicID, id uint64) error
	hasOverlapFn func(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
}

func (m *mockClosingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockClosingSpecialPeriodRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
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

func (m *mockClosingSpecialPeriodRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockClosingSpecialPeriodRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockClosingSpecialPeriodRepository) CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error) {
	if m.hasOverlapFn != nil {
		return m.hasOverlapFn(ctx, clinicID, startDate, endDate, excludeID)
	}
	return false, nil
}

// ---- ヘルパー ----

// mockClosingHolidayRepository は ClinicHolidayRepository のテスト用モック（closing_settings 専用）
type mockClosingHolidayRepository struct {
	findByDateFn      func(context.Context, uint64, time.Time) (*model.ClinicHoliday, error)
	findByYearMonthFn func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	upsertFn          func(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	deleteFn          func(ctx context.Context, clinicID uint64, date time.Time) error
}

func (m *mockClosingHolidayRepository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	if m.findByYearMonthFn != nil {
		return m.findByYearMonthFn(ctx, clinicID, yearMonth)
	}
	return nil, nil
}

func (m *mockClosingHolidayRepository) Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, clinicID, holiday)
	}
	return holiday, nil
}

func (m *mockClosingHolidayRepository) Delete(ctx context.Context, clinicID uint64, date time.Time) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, date)
	}
	return nil
}

func defaultSettings() *model.ClinicSettings {
	return &model.ClinicSettings{
		ClinicID:            1,
		ClosingAmPmBoundary: "14:00",
		ClosingWeekdayEnd:   "18:30",
		ClosingSundayEnd:    "17:30",
		ClosedWeekdays:      nil,
	}
}

func newClosingService(
	settingsRepo *mockClinicSettingsRepository,
	periodRepo *mockClosingSpecialPeriodRepository,
	holidayRepo *mockClosingHolidayRepository,
) ClosingSettingsService {
	return NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo)
}

// ---- テスト ----

func TestClosingSettingsService_ResolveSchedule(t *testing.T) {
	// 2026-04-20 (月曜日)
	monday := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	// 2026-04-19 (日曜日)
	sunday := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)

	specialPeriod := &model.ClosingSpecialPeriod{
		ID:           10,
		ClinicID:     1,
		StartDate:    monday,
		EndDate:      monday,
		AmPmBoundary: "13:00",
		PmEnd:        "17:00",
	}

	tests := []struct {
		name        string
		date        time.Time
		settingsFn  func(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
		periodFn    func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
		holidayFn   func(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
		wantErr     bool
		wantErrIs   error
		checkResult func(t *testing.T, got *DaySchedule)
	}{
		{
			name: "特別期間内の日付: 特別期間設定を返す",
			date: monday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return specialPeriod, nil
			},
			checkResult: func(t *testing.T, got *DaySchedule) {
				assert.Equal(t, "13:00", got.AmPmBoundary)
				assert.Equal(t, "17:00", got.PmEnd)
				assert.False(t, got.IsHoliday)
			},
		},
		{
			name: "個別休診日: IsHoliday=true を返す",
			date: monday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, nil
			},
			settingsFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return defaultSettings(), nil
			},
			holidayFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{
					{ClinicID: 1, Date: monday},
				}, nil
			},
			checkResult: func(t *testing.T, got *DaySchedule) {
				assert.True(t, got.IsHoliday)
				assert.Equal(t, "14:00", got.AmPmBoundary)
			},
		},
		{
			name: "通常の平日: デフォルト設定を返す",
			date: monday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, nil
			},
			settingsFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return defaultSettings(), nil
			},
			holidayFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *DaySchedule) {
				assert.Equal(t, "14:00", got.AmPmBoundary)
				assert.Equal(t, "18:30", got.PmEnd)
				assert.False(t, got.IsHoliday)
			},
		},
		{
			name: "日曜日: sunday 設定を返す",
			date: sunday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, nil
			},
			settingsFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return defaultSettings(), nil
			},
			holidayFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *DaySchedule) {
				assert.Equal(t, "14:00", got.AmPmBoundary)
				assert.Equal(t, "17:30", got.PmEnd)
				assert.False(t, got.IsHoliday)
			},
		},
		{
			name: "週次休診曜日（日曜）: IsHoliday=true を返す",
			date: sunday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, nil
			},
			settingsFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				s := defaultSettings()
				s.ClosedWeekdays = []int64{0} // 0 = 日曜
				return s, nil
			},
			holidayFn: func(_ context.Context, _ uint64, _ string) ([]model.ClinicHoliday, error) {
				return []model.ClinicHoliday{}, nil
			},
			checkResult: func(t *testing.T, got *DaySchedule) {
				assert.True(t, got.IsHoliday)
			},
		},
		{
			name: "periodRepo.FindByDate エラー: エラーを返す",
			date: monday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "settingsRepo.Get エラー: エラーを返す",
			date: monday,
			periodFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClosingSpecialPeriod, error) {
				return nil, nil
			},
			settingsFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			settingsRepo := &mockClinicSettingsRepository{findByClinicIDFn: tt.settingsFn}
			periodRepo := &mockClosingSpecialPeriodRepository{findByDateFn: tt.periodFn}
			holidayRepo := &mockClosingHolidayRepository{findByYearMonthFn: tt.holidayFn}
			svc := newClosingService(settingsRepo, periodRepo, holidayRepo)

			// Act
			got, err := svc.ResolveSchedule(context.Background(), 1, tt.date)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

func TestClosingSettingsService_CreateSpecialPeriod(t *testing.T) {
	startTime := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	startStr := "2026-05-01"
	endStr := "2026-05-05"

	created := &model.ClosingSpecialPeriod{
		ID:           1,
		ClinicID:     1,
		StartDate:    startTime,
		EndDate:      endTime,
		AmPmBoundary: "13:00",
		PmEnd:        "17:00",
		Note:         "GW",
	}

	validInput := CreateSpecialPeriodInput{
		StartDate:    startStr,
		EndDate:      endStr,
		AmPmBoundary: "13:00",
		PmEnd:        "17:00",
		Note:         "GW",
	}

	tests := []struct {
		name         string
		input        CreateSpecialPeriodInput
		hasOverlapFn func(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
		createFn     func(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
		wantErr      bool
		wantErrIs    error
		wantResult   *model.ClosingSpecialPeriod
	}{
		{
			name:  "正常: 作成済みレコードを返す",
			input: validInput,
			hasOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return false, nil
			},
			createFn: func(_ context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
				return created, nil
			},
			wantResult: created,
		},
		{
			name:  "エラー: 期間重複あり → ErrConflict",
			input: validInput,
			hasOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return true, nil
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrConflict,
		},
		{
			name: "エラー: 開始日が終了日より後 → ErrInvalidInput",
			input: CreateSpecialPeriodInput{
				StartDate:    endStr,
				EndDate:      startStr, // 逆転
				AmPmBoundary: "13:00",
				PmEnd:        "17:00",
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name: "エラー: 時刻バリデーション不正（boundary >= pmEnd） → ErrInvalidInput",
			input: CreateSpecialPeriodInput{
				StartDate:    startStr,
				EndDate:      endStr,
				AmPmBoundary: "18:00",
				PmEnd:        "14:00",
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrInvalidInput,
		},
		{
			name:  "エラー: CheckOverlap がエラーを返す",
			input: validInput,
			hasOverlapFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			periodRepo := &mockClosingSpecialPeriodRepository{
				hasOverlapFn: tt.hasOverlapFn,
				createFn:     tt.createFn,
			}
			svc := newClosingService(
				&mockClinicSettingsRepository{},
				periodRepo,
				&mockClosingHolidayRepository{},
			)

			// Act
			got, err := svc.CreateSpecialPeriod(context.Background(), 1, &tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestClosingSettingsService_ListSpecialPeriods(t *testing.T) {
	periods := []model.ClosingSpecialPeriod{
		{ID: 1, ClinicID: 1, AmPmBoundary: "13:00", PmEnd: "17:00"},
		{ID: 2, ClinicID: 1, AmPmBoundary: "13:00", PmEnd: "17:00"},
	}

	tests := []struct {
		name       string
		findAllFn  func(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
		wantErr    bool
		wantResult []model.ClosingSpecialPeriod
	}{
		{
			name: "正常: 全件を返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
				return periods, nil
			},
			wantResult: periods,
		},
		{
			name: "エラー: DB エラーを返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.ClosingSpecialPeriod, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			svc := newClosingService(
				&mockClinicSettingsRepository{},
				&mockClosingSpecialPeriodRepository{findAllFn: tt.findAllFn},
				&mockClosingHolidayRepository{},
			)

			// Act
			got, err := svc.ListSpecialPeriods(context.Background(), 1)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func (m *mockClosingHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, nil
}
