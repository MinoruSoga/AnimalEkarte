package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpsertLineReservationSettingInput は予約設定 upsert のための入力データ
type UpsertLineReservationSettingInput struct {
	Status                  string
	HeaderText              string
	ReservationNotice       string
	CancelNotice            string
	PrivacyPolicy           string
	ClosedWeekdays          []byte
	ClosedDates             []byte
	NationalHolidayClosed   bool
	BusinessHours           []byte
	BusinessHoursByWeekday  []byte
	BreakHours              []byte
	DailyLimit              *int
	MonthlyLimit            *int
	BookingWindowMaxDays    int
	BookingWindowMinDays    int
	CalendarMonths          int
	PhoneNumber             string
	NotificationEmail       string
	RequestExample          string
	TimeSlotMode            string
	TimeSlotIntervalMinutes int
	NoStaffMode             string
	ShowNoStaffOption       bool
	AdditionalFields        []byte
	// EMR-504: LINE credentials (channel ID / LIFF ID / access token / channel secret) are not part
	// of the reservation setting; their only store is clinic_integrations (lstep-settings API).
}

// LineReservationSettingService は予約基本設定のビジネスロジックインターフェース
type LineReservationSettingService interface {
	Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	Save(ctx context.Context, clinicID uint64, input *UpsertLineReservationSettingInput) (*model.LineReservationSetting, bool, error)
}

type lineReservationSettingService struct {
	repo LineReservationSettingRepository
}

// NewLineReservationSettingService は LineReservationSettingService を初期化して返す。
func NewLineReservationSettingService(repo LineReservationSettingRepository) LineReservationSettingService {
	return &lineReservationSettingService{repo: repo}
}

func (s *lineReservationSettingService) Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	setting, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	return setting, nil
}

// validateJSONFields は []byte フィールドが有効な JSON であることを確認する。
func validateJSONFields(fields map[string][]byte) error {
	for name, data := range fields {
		if len(data) > 0 && !json.Valid(data) {
			return apperrors.WrapInvalidInput(name + " は有効な JSON である必要があります")
		}
	}
	return nil
}

var (
	emptyJSONArray           = []byte("[]")
	defaultBusinessHoursJSON = []byte(`{"start":"0900","end":"1900"}`)
)

// defaultJSONIfEmptyOrNull maps omitted or JSON null bytes to fallback.
// Empty []byte is persisted as an empty SQL string and becomes jsonb null, which violates NOT NULL.
func defaultJSONIfEmptyOrNull(data, fallback []byte) []byte {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return fallback
	}
	return data
}

// validateBreakHoursShape は break_hours が []BreakPeriod{start,end} 形式に unmarshal でき、
// 各エントリの start/end が HHMM 形式であることを確認する（go-reviewer/security-reviewer 指摘）。
//
// BE-refactor.md D10/F-2 で parseBusinessHoursForDate の break_hours unmarshal 失敗を
// fail-closed（予約拒否）に変更したため、構文的には有効だが形状が不正な JSON（例: "{}"）が
// 保存されると、保存時ではなく以降のあらゆる予約作成試行が拒否され続ける可用性劣化になる。
// 保存時にこの形状を検証し、破損データが永続化される前に 400 で拒否する。
func validateBreakHoursShape(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var breaks []BreakPeriod
	if err := json.Unmarshal(data, &breaks); err != nil {
		return apperrors.WrapInvalidInput(`break_hours は [{"start":"HHMM","end":"HHMM"}] 形式である必要があります`)
	}
	for _, b := range breaks {
		if _, err := MinutesSinceMidnight(b.Start); err != nil {
			return apperrors.WrapInvalidInput("break_hours の start は HHMM 形式である必要があります: " + b.Start)
		}
		if _, err := MinutesSinceMidnight(b.End); err != nil {
			return apperrors.WrapInvalidInput("break_hours の end は HHMM 形式である必要があります: " + b.End)
		}
	}
	return nil
}

func (s *lineReservationSettingService) Save(ctx context.Context, clinicID uint64, input *UpsertLineReservationSettingInput) (*model.LineReservationSetting, bool, error) {
	closedWeekdays, err := parseClosedWeekdaysJSON(input.ClosedWeekdays)
	if err != nil {
		return nil, false, err
	}
	canonicalClosedWeekdays, err := json.Marshal(closedWeekdays)
	if err != nil {
		return nil, false, apperrors.WrapInvalidInput("closed_weekdays は 0〜6 の整数配列である必要があります")
	}
	if err := validateJSONFields(map[string][]byte{
		"closed_dates":              input.ClosedDates,
		"business_hours":            input.BusinessHours,
		"business_hours_by_weekday": input.BusinessHoursByWeekday,
		"break_hours":               input.BreakHours,
		"additional_fields":         input.AdditionalFields,
	}); err != nil {
		return nil, false, err
	}

	closedDates := defaultJSONIfEmptyOrNull(input.ClosedDates, emptyJSONArray)
	businessHours := defaultJSONIfEmptyOrNull(input.BusinessHours, defaultBusinessHoursJSON)
	breakHours := defaultJSONIfEmptyOrNull(input.BreakHours, emptyJSONArray)
	additionalFields := defaultJSONIfEmptyOrNull(input.AdditionalFields, emptyJSONArray)
	if err := validateBreakHoursShape(breakHours); err != nil {
		return nil, false, err
	}

	// 既存レコードの有無を確認し、新規作成かどうかを判定する
	existing, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, false, apperrors.Wrap(err, "failed to get existing reservation setting")
	}
	isNew := existing == nil

	setting := &model.LineReservationSetting{
		ClinicID:                clinicID,
		Status:                  input.Status,
		HeaderText:              input.HeaderText,
		ReservationNotice:       input.ReservationNotice,
		CancelNotice:            input.CancelNotice,
		PrivacyPolicy:           input.PrivacyPolicy,
		ClosedWeekdays:          canonicalClosedWeekdays,
		ClosedDates:             closedDates,
		NationalHolidayClosed:   input.NationalHolidayClosed,
		BusinessHours:           businessHours,
		BusinessHoursByWeekday:  nullableJSONPtr(input.BusinessHoursByWeekday),
		BreakHours:              breakHours,
		DailyLimit:              input.DailyLimit,
		MonthlyLimit:            input.MonthlyLimit,
		BookingWindowMaxDays:    input.BookingWindowMaxDays,
		BookingWindowMinDays:    input.BookingWindowMinDays,
		CalendarMonths:          input.CalendarMonths,
		PhoneNumber:             input.PhoneNumber,
		NotificationEmail:       input.NotificationEmail,
		RequestExample:          input.RequestExample,
		TimeSlotMode:            input.TimeSlotMode,
		TimeSlotIntervalMinutes: input.TimeSlotIntervalMinutes,
		NoStaffMode:             input.NoStaffMode,
		ShowNoStaffOption:       input.ShowNoStaffOption,
		AdditionalFields:        additionalFields,
	}
	if err := s.repo.Save(ctx, clinicID, setting); err != nil {
		return nil, false, apperrors.Wrap(err, "failed to upsert reservation setting")
	}
	// RSV-03: return the write result; do not re-fetch after commit (CODING_RULES.md:78).
	slog.InfoContext(ctx, "reservation setting upserted",
		slog.Uint64("clinic_id", clinicID))
	return setting, isNew, nil
}
