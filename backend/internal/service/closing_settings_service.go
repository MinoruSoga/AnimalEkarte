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

const (
	colSpecialPeriodStartDate    = "start_date"
	colSpecialPeriodEndDate      = "end_date"
	colSpecialPeriodAmPmBoundary = "am_pm_boundary"
	colSpecialPeriodPmEnd        = "pm_end"
	colSpecialPeriodNote         = "note"
)

func buildSpecialPeriodUpdate(input UpdateSpecialPeriodInput, parsedStart, parsedEnd *time.Time) map[string]any {
	fields := make(map[string]any)
	if parsedStart != nil {
		fields[colSpecialPeriodStartDate] = *parsedStart
	}
	if parsedEnd != nil {
		fields[colSpecialPeriodEndDate] = *parsedEnd
	}
	if input.AmPmBoundary != nil {
		fields[colSpecialPeriodAmPmBoundary] = *input.AmPmBoundary
	}
	if input.PmEnd != nil {
		fields[colSpecialPeriodPmEnd] = *input.PmEnd
	}
	if input.Note != nil {
		fields[colSpecialPeriodNote] = *input.Note
	}
	return fields
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
	StartDate    string // YYYY-MM-DD
	EndDate      string // YYYY-MM-DD
	AmPmBoundary string
	PmEnd        string
	Note         string
}

// UpdateSpecialPeriodInput は特別期間更新の入力
type UpdateSpecialPeriodInput struct {
	StartDate    *string // YYYY-MM-DD
	EndDate      *string // YYYY-MM-DD
	AmPmBoundary *string
	PmEnd        *string
	Note         *string
}

// ClosingSettingsService は締め時間設定のビジネスロジックインターフェース
type ClosingSettingsService interface {
	Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error)
	ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	UpdateStandard(ctx context.Context, clinicID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error)
	CreateSpecialPeriod(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error
	ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
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

func (s *closingSettingsService) ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	periods, err := s.periodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list special periods", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list special periods")
	}
	return periods, nil
}

func (s *closingSettingsService) Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
	settings, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get clinic settings")
	}
	periods, err := s.periodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get special periods", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get special periods")
	}
	return &ClosingSettingsResponse{Settings: settings, SpecialPeriods: periods}, nil
}

func (s *closingSettingsService) UpdateStandard(ctx context.Context, clinicID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
	slog.InfoContext(ctx, "updating clinic settings", slog.Uint64("clinic_id", clinicID))
	current, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get current settings", "error", err, "clinic_id", clinicID)
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
	result, err := s.settingsRepo.Save(ctx, clinicID, current)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update clinic settings", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update clinic settings")
	}
	return result, nil
}

func (s *closingSettingsService) CreateSpecialPeriod(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	startDate, err := time.ParseInLocation("2006-01-02", input.StartDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
	}
	endDate, err := time.ParseInLocation("2006-01-02", input.EndDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
	}
	if err := validateSpecialPeriodTimes(input.AmPmBoundary, input.PmEnd); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate special period times")
	}
	if startDate.After(endDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
	}
	overlap, err := s.periodRepo.CheckOverlap(ctx, clinicID, startDate, endDate, nil)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check period overlap", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to check period overlap")
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}
	slog.InfoContext(ctx, "creating closing special period",
		slog.Uint64("clinic_id", clinicID),
		slog.String("start_date", input.StartDate),
		slog.String("end_date", input.EndDate))
	p := &model.ClosingSpecialPeriod{
		ClinicID:     clinicID,
		StartDate:    startDate,
		EndDate:      endDate,
		AmPmBoundary: input.AmPmBoundary,
		PmEnd:        input.PmEnd,
		Note:         input.Note,
	}
	created, err := s.periodRepo.Create(ctx, p)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create closing special period", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create special period")
	}
	return created, nil
}

func (s *closingSettingsService) UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	current, err := s.periodRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get special period", "error", err, "id", id, "clinic_id", clinicID)
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
		return nil, apperrors.Wrap(err, "failed to validate special period times")
	}

	// 期間バリデーション（変更がある場合のみ）
	startDate := current.StartDate
	endDate := current.EndDate
	var parsedStart, parsedEnd *time.Time
	if input.StartDate != nil {
		t, err := time.ParseInLocation("2006-01-02", *input.StartDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
		}
		parsedStart = &t
		startDate = t
	}
	if input.EndDate != nil {
		t, err := time.ParseInLocation("2006-01-02", *input.EndDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
		}
		parsedEnd = &t
		endDate = t
	}
	if startDate.After(endDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
	}

	// 重複チェック（自分自身を除外）
	overlap, err := s.periodRepo.CheckOverlap(ctx, clinicID, startDate, endDate, &id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check period overlap", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to check period overlap")
	}
	if overlap {
		return nil, apperrors.WrapConflict("期間が他の特別期間と重複しています")
	}

	fields := buildSpecialPeriodUpdate(input, parsedStart, parsedEnd)
	if len(fields) == 0 {
		return current, nil
	}
	result, err := s.periodRepo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update closing special period",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("id", id),
			slog.Any("error", err))
		slog.ErrorContext(ctx, "failed to update special period", "error", err)
		return nil, apperrors.Wrap(err, "failed to update special period")
	}
	slog.InfoContext(ctx, "closing special period updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return result, nil
}

func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.periodRepo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get special period")
	}
	slog.InfoContext(ctx, "deleting closing special period",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	if err := s.periodRepo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete special period", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete special period")
	}
	return nil
}

func (s *closingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error) {
	// 特別期間をチェック（優先）
	special, err := s.periodRepo.FindByDate(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find special period", "error", err, "clinic_id", clinicID)
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
	settings, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings", "error", err, "clinic_id", clinicID)
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
		holidays, err := s.holidayRepo.FindAllByYearMonth(ctx, clinicID, yearMonth)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find clinic holidays", "error", err, "clinic_id", clinicID)
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
	// parseHHMM を使って "HH:MM" / "HH:MM:SS" 混在を正規化してから分単位で比較する。
	// 文字列比較だと "13:30" < "13:30:00" になり同一時刻を誤って通過させる。
	bh, bm, err := parseHHMM(boundary)
	if err != nil {
		return apperrors.WrapInvalidInput("境界時刻の形式が正しくありません")
	}
	ph, pm, err := parseHHMM(pmEnd)
	if err != nil {
		return apperrors.WrapInvalidInput("PM締め終了時刻の形式が正しくありません")
	}
	if bh*60+bm >= ph*60+pm {
		return apperrors.WrapInvalidInput(fmt.Sprintf("PM締め終了時刻(%s)は境界時刻(%s)より後に設定してください", pmEnd, boundary))
	}
	return nil
}
