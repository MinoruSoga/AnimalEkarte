package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ClosingSettingsService は締め時間設定のビジネスロジックインターフェース
type ClosingSettingsService interface {
	Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error)
	UpdateStandard(ctx context.Context, clinicID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error)
	CreateSpecialPeriod(ctx context.Context, clinicID uint64, input CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error
	ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
}

// ClosingSettingsResponse は設定・特別期間をまとめたレスポンス
type ClosingSettingsResponse struct {
	Settings       *model.ClinicSettings        `json:"settings"`
	SpecialPeriods []model.ClosingSpecialPeriod `json:"special_periods"`
}

// DaySchedule は指定日の締め時間スケジュール
type DaySchedule struct {
	AmPmBoundary string `json:"am_pm_boundary"`
	PmEnd        string `json:"pm_end"`
	IsHoliday    bool   `json:"is_holiday"`
}

// UpdateClinicSettingsInput は標準設定更新の入力
type UpdateClinicSettingsInput struct {
	ClosingAmPmBoundary *string
	ClosingWeekdayEnd   *string
	ClosingSundayEnd    *string
	ClosedWeekdays      []int64
}

// CreateSpecialPeriodInput は特別期間作成の入力
type CreateSpecialPeriodInput struct {
	StartDate    time.Time
	EndDate      time.Time
	AmPmBoundary string
	PmEnd        string
	Note         string
}

// UpdateSpecialPeriodInput は特別期間更新の入力
type UpdateSpecialPeriodInput struct {
	StartDate    *time.Time
	EndDate      *time.Time
	AmPmBoundary *string
	PmEnd        *string
	Note         *string
}

type closingSettingsService struct {
	settingsRepo repository.ClinicSettingsRepository
	periodRepo   repository.ClosingSpecialPeriodRepository
	holidayRepo  repository.ClinicHolidayRepository
}

// NewClosingSettingsService は ClosingSettingsService を初期化して返す
func NewClosingSettingsService(
	settingsRepo repository.ClinicSettingsRepository,
	periodRepo repository.ClosingSpecialPeriodRepository,
	holidayRepo repository.ClinicHolidayRepository,
) ClosingSettingsService {
	return &closingSettingsService{
		settingsRepo: settingsRepo,
		periodRepo:   periodRepo,
		holidayRepo:  holidayRepo,
	}
}

func (s *closingSettingsService) Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
	settings, err := s.settingsRepo.Get(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinic settings")
	}
	periods, err := s.periodRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get special periods")
	}
	return &ClosingSettingsResponse{Settings: settings, SpecialPeriods: periods}, nil
}

func (s *closingSettingsService) UpdateStandard(ctx context.Context, clinicID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
	slog.InfoContext(ctx, "updating clinic settings", slog.Uint64("clinic_id", clinicID))
	current, err := s.settingsRepo.Get(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get current settings")
	}
	if input.ClosingAmPmBoundary != nil {
		current.ClosingAmPmBoundary = *input.ClosingAmPmBoundary
	}
	if input.ClosingWeekdayEnd != nil {
		current.ClosingWeekdayEnd = *input.ClosingWeekdayEnd
	}
	if input.ClosingSundayEnd != nil {
		current.ClosingSundayEnd = *input.ClosingSundayEnd
	}
	if input.ClosedWeekdays != nil {
		current.ClosedWeekdays = input.ClosedWeekdays
	}
	result, err := s.settingsRepo.Upsert(ctx, current)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update clinic settings")
	}
	return result, nil
}

func (s *closingSettingsService) CreateSpecialPeriod(ctx context.Context, clinicID uint64, input CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	if err := validateSpecialPeriodTimes(input.AmPmBoundary, input.PmEnd); err != nil {
		return nil, err
	}
	if input.StartDate.After(input.EndDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
	}
	overlap, err := s.periodRepo.HasOverlap(ctx, clinicID, input.StartDate, input.EndDate, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check period overlap")
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}
	slog.InfoContext(ctx, "creating closing special period",
		slog.Uint64("clinic_id", clinicID),
		slog.String("start_date", input.StartDate.Format("2006-01-02")),
		slog.String("end_date", input.EndDate.Format("2006-01-02")))
	p := &model.ClosingSpecialPeriod{
		ClinicID:     clinicID,
		StartDate:    input.StartDate,
		EndDate:      input.EndDate,
		AmPmBoundary: input.AmPmBoundary,
		PmEnd:        input.PmEnd,
		Note:         input.Note,
	}
	return s.periodRepo.Create(ctx, p)
}

func (s *closingSettingsService) UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	current, err := s.periodRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get special period")
	}

	// 時刻バリデーション（変更がある場合のみ）
	boundary := current.AmPmBoundary
	pmEnd := current.PmEnd
	if input.AmPmBoundary != nil {
		boundary = *input.AmPmBoundary
	}
	if input.PmEnd != nil {
		pmEnd = *input.PmEnd
	}
	if err := validateSpecialPeriodTimes(boundary, pmEnd); err != nil {
		return nil, err
	}

	// 期間バリデーション（変更がある場合のみ）
	startDate := current.StartDate
	endDate := current.EndDate
	if input.StartDate != nil {
		startDate = *input.StartDate
	}
	if input.EndDate != nil {
		endDate = *input.EndDate
	}
	if startDate.After(endDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
	}

	// 重複チェック（自分自身を除外）
	overlap, err := s.periodRepo.HasOverlap(ctx, clinicID, startDate, endDate, &id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check period overlap")
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}

	fields := map[string]any{}
	if input.StartDate != nil {
		fields["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		fields["end_date"] = *input.EndDate
	}
	if input.AmPmBoundary != nil {
		fields["am_pm_boundary"] = *input.AmPmBoundary
	}
	if input.PmEnd != nil {
		fields["pm_end"] = *input.PmEnd
	}
	if input.Note != nil {
		fields["note"] = *input.Note
	}
	if len(fields) == 0 {
		return current, nil
	}
	return s.periodRepo.UpdateFields(ctx, clinicID, id, fields)
}

func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
	slog.InfoContext(ctx, "deleting closing special period",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return s.periodRepo.Delete(ctx, clinicID, id)
}

func (s *closingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error) {
	// 特別期間をチェック（優先）
	special, err := s.periodRepo.FindByDate(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find special period")
	}
	if special != nil {
		return &DaySchedule{
			AmPmBoundary: special.AmPmBoundary,
			PmEnd:        special.PmEnd,
			IsHoliday:    false,
		}, nil
	}

	// 標準設定を取得
	settings, err := s.settingsRepo.Get(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinic settings")
	}
	pmEnd := settings.ClosingWeekdayEnd
	if date.Weekday() == time.Sunday {
		pmEnd = settings.ClosingSundayEnd
	}

	// 週次休診曜日チェック
	isHoliday := false
	wd := int64(date.Weekday())
	for _, cw := range settings.ClosedWeekdays {
		if cw == wd {
			isHoliday = true
			break
		}
	}

	// 個別休診日チェック
	if !isHoliday {
		yearMonth := date.Format("2006-01")
		holidays, err := s.holidayRepo.FindByYearMonth(ctx, clinicID, yearMonth)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to find clinic holidays")
		}
		dateStr := date.Format("2006-01-02")
		for _, h := range holidays {
			if h.Date.Format("2006-01-02") == dateStr {
				isHoliday = true
				break
			}
		}
	}

	return &DaySchedule{
		AmPmBoundary: settings.ClosingAmPmBoundary,
		PmEnd:        pmEnd,
		IsHoliday:    isHoliday,
	}, nil
}

func validateSpecialPeriodTimes(boundary, pmEnd string) error {
	if boundary >= pmEnd {
		return apperrors.WrapInvalidInput(fmt.Sprintf("PM締め終了時刻(%s)は境界時刻(%s)より後に設定してください", pmEnd, boundary))
	}
	return nil
}
