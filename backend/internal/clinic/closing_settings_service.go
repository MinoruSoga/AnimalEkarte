package clinic

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

func specialPeriodUpdateFields(input UpdateSpecialPeriodInput) map[string]any {
	var parsedStart, parsedEnd *time.Time
	if input.StartDate != nil {
		if t, err := time.ParseInLocation(time.DateOnly, *input.StartDate, time.Local); err == nil {
			parsedStart = &t
		}
	}
	if input.EndDate != nil {
		if t, err := time.ParseInLocation(time.DateOnly, *input.EndDate, time.Local); err == nil {
			parsedEnd = &t
		}
	}
	return buildSpecialPeriodUpdate(input, parsedStart, parsedEnd)
}

// ClosingSettingsResponse は設定・特別期間をまとめたレスポンス
type ClosingSettingsResponse struct {
	Settings       *model.ClinicSettings        `json:"settings"`
	SpecialPeriods []model.ClosingSpecialPeriod `json:"special_periods"`
}

// defaultClosingAmStart は AM 開始時刻の既定値（#215）。migration 011 の DB default と一致させる。
const defaultClosingAmStart = "09:00"

// DaySchedule は指定日の締め時間スケジュール（実装は sharedkernel へ昇格・B⑤）。
type DaySchedule = sharedkernel.DaySchedule

// amStartOrDefault は設定の AM 開始時刻を返す。未設定（migration 011 以前のデータ・zero-value）は既定 09:00。
func amStartOrDefault(settings *model.ClinicSettings) string {
	if settings != nil && settings.ClosingAmStart != "" {
		return settings.ClosingAmStart
	}
	return defaultClosingAmStart
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
	UpdateStandard(ctx context.Context, clinicID, actorID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error)
	CreateSpecialPeriod(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error
	ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
}

// ClinicRowLocker serializes clinic-scoped writes via FOR UPDATE on the clinics row.
// Implemented by Repository.LockByIDForUpdate; fail-closed without ambient tx.
type ClinicRowLocker interface {
	LockByIDForUpdate(ctx context.Context, id uint64) (*model.Clinic, error)
}

// AuditEntry is clinic's consumer-side audit payload (mirrors audit.Entry fields used here).
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
}

// AuditTxLogger is the fail-closed ambient-transaction audit port for closing settings.
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}

// ClosingSettingsServiceDeps holds optional integrity dependencies for UpdateStandard.
// Nil deps (or nil fields) are allowed for read-only paths; UpdateStandard fail-closes if missing.
type ClosingSettingsServiceDeps struct {
	Transactor   Transactor
	ClinicLocker ClinicRowLocker
	AuditTx      AuditTxLogger
}

const (
	auditActionClosingSettingsUpdateStandard = "closing_settings.update_standard"
	auditResourceClinicSettings              = "clinic_settings"
)

type closingSettingsService struct {
	settingsRepo ClinicSettingsRepository
	periodRepo   ClosingSpecialPeriodRepository
	holidayRepo  ClinicHolidayRepository
	transactor   Transactor
	clinicLocker ClinicRowLocker
	auditTx      AuditTxLogger
}

// NewClosingSettingsService は ClosingSettingsService を初期化して返す。
// deps は UpdateStandard の transaction / 行ロック / audit 用。nil 可（読取系のみのテスト向け）。
func NewClosingSettingsService(
	settingsRepo ClinicSettingsRepository,
	periodRepo ClosingSpecialPeriodRepository,
	holidayRepo ClinicHolidayRepository,
	deps *ClosingSettingsServiceDeps,
) ClosingSettingsService {
	svc := &closingSettingsService{
		settingsRepo: settingsRepo,
		periodRepo:   periodRepo,
		holidayRepo:  holidayRepo,
	}
	if deps != nil {
		svc.transactor = deps.Transactor
		svc.clinicLocker = deps.ClinicLocker
		svc.auditTx = deps.AuditTx
	}
	return svc
}

func (s *closingSettingsService) ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	periods, err := s.periodRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list special periods")
	}
	return periods, nil
}

func (s *closingSettingsService) Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
	settings, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinic settings")
	}
	periods, err := s.periodRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get special periods")
	}
	return &ClosingSettingsResponse{Settings: settings, SpecialPeriods: periods}, nil
}

func (s *closingSettingsService) UpdateStandard(ctx context.Context, clinicID, actorID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error) {
	slog.InfoContext(ctx, "updating clinic settings", slog.Uint64("clinic_id", clinicID), slog.Uint64("actor_id", actorID))
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("closing settings transactor is required")
	}
	if s.clinicLocker == nil {
		return nil, apperrors.WrapInternalServerError("closing settings clinic lock is required")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("closing settings audit dependency is required")
	}
	if actorID == 0 {
		return nil, apperrors.WrapInvalidInput("actor_id is required")
	}
	if err := validateClosedWeekdays(input.ClosedWeekdays); err != nil {
		return nil, err
	}

	var result *model.ClinicSettings
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Lock parent clinic row BEFORE read so concurrent partial PATCHes serialize
		// even when clinic_settings row is still missing (first upsert).
		if _, err := s.clinicLocker.LockByIDForUpdate(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to lock clinic for settings update")
		}

		current, err := s.settingsRepo.FindByClinicID(txCtx, clinicID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get current settings")
		}
		beforeMeta := closingSettingsFieldPresence(current)

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

		if err := validateStandardClosingTimes(current.ClosingAmPmBoundary, current.ClosingWeekdayEnd, current.ClosingSundayEnd); err != nil {
			return err
		}
		if err := validateClosedWeekdays(current.ClosedWeekdays); err != nil {
			return err
		}

		saved, err := s.settingsRepo.Save(txCtx, clinicID, current)
		if err != nil {
			return apperrors.Wrap(err, "failed to update clinic settings")
		}
		result = saved

		afterMeta := closingSettingsFieldPresence(saved)
		changed := closingSettingsChangedFields(beforeMeta, afterMeta, input)
		actor := actorID
		clinic := clinicID
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinic,
			ActorID:    &actor,
			ActorType:  sharedkernel.AuditActorTypeFor(&actor),
			Action:     auditActionClosingSettingsUpdateStandard,
			Resource:   auditResourceClinicSettings,
			ResourceID: &clinic,
			OldValue: map[string]any{
				"fields": beforeMeta,
			},
			NewValue: map[string]any{
				"fields":         afterMeta,
				"changed_fields": changed,
			},
		}); err != nil {
			return apperrors.Wrap(err, "failed to write closing settings audit")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *closingSettingsService) CreateSpecialPeriod(ctx context.Context, clinicID uint64, input *CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	startDate, err := time.ParseInLocation(time.DateOnly, input.StartDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
	}
	endDate, err := time.ParseInLocation(time.DateOnly, input.EndDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
	}
	if err := validateSpecialPeriodTimes(input.AmPmBoundary, input.PmEnd); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate special period times")
	}
	if startDate.After(endDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
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
	// POC-05 / X-05: CheckOverlap + Create を同一 tx（clinic advisory lock）で直列化する。
	created, err := s.periodRepo.CreateCheckingOverlap(ctx, p)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create special period")
	}
	return created, nil
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
		return nil, apperrors.Wrap(err, "failed to validate special period times")
	}

	// 期間バリデーション（変更がある場合のみ）
	startDate := current.StartDate
	endDate := current.EndDate
	var parsedStart, parsedEnd *time.Time
	if input.StartDate != nil {
		t, err := time.ParseInLocation(time.DateOnly, *input.StartDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
		}
		parsedStart = &t
		startDate = t
	}
	if input.EndDate != nil {
		t, err := time.ParseInLocation(time.DateOnly, *input.EndDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
		}
		parsedEnd = &t
		endDate = t
	}
	if startDate.After(endDate) {
		return nil, apperrors.WrapInvalidInput("開始日は終了日以前に設定してください")
	}

	fields := buildSpecialPeriodUpdate(input, parsedStart, parsedEnd)
	if len(fields) == 0 {
		return current, nil
	}
	// POC-05 / X-05: CheckOverlap + Update を同一 tx（clinic advisory lock）で直列化する。
	result, err := s.periodRepo.UpdateCheckingOverlap(ctx, clinicID, id, startDate, endDate, input)
	if err != nil {
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
		return apperrors.Wrap(err, "failed to delete special period")
	}
	return nil
}

func (s *closingSettingsService) ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error) {
	// 特別期間をチェック（優先）
	special, err := s.periodRepo.FindByDate(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find special period")
	}
	if special != nil {
		// #215: 特別期間は am_start を持たないため標準設定の値を継承する。
		settings, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get clinic settings")
		}
		return &DaySchedule{
			AmPmBoundary: special.AmPmBoundary,
			PmEnd:        special.PmEnd,
			AmStart:      amStartOrDefault(settings),
			IsHoliday:    false,
		}, nil
	}

	// 標準設定を取得
	settings, err := s.settingsRepo.FindByClinicID(ctx, clinicID)
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
		holidays, err := s.holidayRepo.FindAllByYearMonth(ctx, clinicID, yearMonth)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to find clinic holidays")
		}
		dateStr := date.Format(time.DateOnly)
		for _, h := range holidays {
			if h.Date.Format(time.DateOnly) == dateStr {
				isHoliday = true
				break
			}
		}
	}

	return &DaySchedule{
		AmPmBoundary: settings.ClosingAmPmBoundary,
		PmEnd:        pmEnd,
		AmStart:      amStartOrDefault(settings),
		IsHoliday:    isHoliday,
	}, nil
}

func validateSpecialPeriodTimes(boundary, pmEnd string) error {
	// sharedkernel.ParseHHMM で "HH:MM" / "HH:MM:SS" 混在を正規化してから分単位で比較する。
	// 文字列比較だと "13:30" < "13:30:00" になり同一時刻を誤って通過させる。
	bh, bm, err := sharedkernel.ParseHHMM(boundary)
	if err != nil {
		return apperrors.WrapInvalidInput("境界時刻の形式が正しくありません")
	}
	ph, pm, err := sharedkernel.ParseHHMM(pmEnd)
	if err != nil {
		return apperrors.WrapInvalidInput("PM締め終了時刻の形式が正しくありません")
	}
	if bh*60+bm >= ph*60+pm {
		return apperrors.WrapInvalidInput(fmt.Sprintf("PM締め終了時刻(%s)は境界時刻(%s)より後に設定してください", pmEnd, boundary))
	}
	return nil
}

// validateStandardClosingTimes reuses sharedkernel.ParseHHMM (same as validateSpecialPeriodTimes)
// and requires boundary strictly before both weekday and sunday ends.
func validateStandardClosingTimes(boundary, weekdayEnd, sundayEnd string) error {
	bh, bm, err := sharedkernel.ParseHHMM(boundary)
	if err != nil {
		return apperrors.WrapInvalidInput("境界時刻の形式が正しくありません")
	}
	wh, wm, err := sharedkernel.ParseHHMM(weekdayEnd)
	if err != nil {
		return apperrors.WrapInvalidInput("平日締め終了時刻の形式が正しくありません")
	}
	sh, sm, err := sharedkernel.ParseHHMM(sundayEnd)
	if err != nil {
		return apperrors.WrapInvalidInput("日曜締め終了時刻の形式が正しくありません")
	}
	bMin := bh*60 + bm
	if bMin >= wh*60+wm {
		return apperrors.WrapInvalidInput(fmt.Sprintf("平日締め終了時刻(%s)は境界時刻(%s)より後に設定してください", weekdayEnd, boundary))
	}
	if bMin >= sh*60+sm {
		return apperrors.WrapInvalidInput(fmt.Sprintf("日曜締め終了時刻(%s)は境界時刻(%s)より後に設定してください", sundayEnd, boundary))
	}
	return nil
}

func validateClosedWeekdays(days []int64) error {
	if days == nil {
		return nil
	}
	seen := make(map[int64]struct{}, len(days))
	for _, d := range days {
		if d < 0 || d > 6 {
			return apperrors.WrapInvalidInput("closed_weekdays は 0〜6 の範囲で指定してください")
		}
		if _, ok := seen[d]; ok {
			return apperrors.WrapInvalidInput("closed_weekdays に重複があります")
		}
		seen[d] = struct{}{}
	}
	return nil
}

// closingSettingsFieldPresence returns non-secret field markers (no actual clock values).
func closingSettingsFieldPresence(s *model.ClinicSettings) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	count := 0
	if s.ClosedWeekdays != nil {
		count = len(s.ClosedWeekdays)
	}
	return map[string]any{
		"closing_am_pm_boundary": "present",
		"closing_weekday_end":    "present",
		"closing_sunday_end":     "present",
		"closed_weekdays_count":  count,
	}
}

func closingSettingsChangedFields(before, after map[string]any, input UpdateClinicSettingsInput) []string {
	changed := make([]string, 0, 4)
	if input.ClosingAmPmBoundary != nil {
		changed = append(changed, "closing_am_pm_boundary")
	}
	if input.ClosingWeekdayEnd != nil {
		changed = append(changed, "closing_weekday_end")
	}
	if input.ClosingSundayEnd != nil {
		changed = append(changed, "closing_sunday_end")
	}
	if input.ClosedWeekdays != nil {
		changed = append(changed, "closed_weekdays")
	}
	_ = before
	_ = after
	return changed
}
