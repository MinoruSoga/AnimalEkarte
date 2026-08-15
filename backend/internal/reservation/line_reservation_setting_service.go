package reservation

import (
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
	LineChannelID           string
	// LineChannelSecret removed in R-05 Phase B — canonical SoT is clinic_integrations.
	LiffID          string
	LineAccessToken string
}

// LineReservationSettingService は予約基本設定のビジネスロジックインターフェース
type LineReservationSettingService interface {
	Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	Save(ctx context.Context, clinicID uint64, input *UpsertLineReservationSettingInput) (*model.LineReservationSetting, bool, error)
}

type lineReservationSettingService struct {
	repo LineReservationSettingRepository
	// encrypt/decrypt は LINE credential の暗号化 helper（lstep 系実装・topo で import 不可）の
	// closure 注入（R④ notification と同型・composition rootが実装を包んで渡す）。
	encryptCredential func(value string) (string, error)
	decryptCredential func(ctx context.Context, value string) string
	// nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
}

// NewLineReservationSettingService は LineReservationSettingService を初期化して返す。
// closure が nil/空文字素通しを判定する場合（旧 cipher=nil 契約）は暗号化なしで動作する（lstep 連携と同一の cipher をcomposition rootから受け取る）。
func NewLineReservationSettingService(
	repo LineReservationSettingRepository,
	encryptCredential func(value string) (string, error),
	decryptCredential func(ctx context.Context, value string) string,
) LineReservationSettingService {
	return &lineReservationSettingService{repo: repo, encryptCredential: encryptCredential, decryptCredential: decryptCredential}
}

func (s *lineReservationSettingService) Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	setting, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil
		}
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err, "clinic_id", clinicID)
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
	if err := validateJSONFields(map[string][]byte{
		"closed_weekdays":           input.ClosedWeekdays,
		"closed_dates":              input.ClosedDates,
		"business_hours":            input.BusinessHours,
		"business_hours_by_weekday": input.BusinessHoursByWeekday,
		"break_hours":               input.BreakHours,
		"additional_fields":         input.AdditionalFields,
	}); err != nil {
		return nil, false, err
	}
	if err := validateBreakHoursShape(input.BreakHours); err != nil {
		return nil, false, err
	}

	// 既存レコードの有無を確認し、新規作成かどうかを判定する
	existing, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil && !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to get existing reservation setting", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to get existing reservation setting")
	}
	isNew := existing == nil

	// LineAccessToken はレスポンスに含まれないため、フロントエンドは既存値を読み取れない。
	// 空文字が送られてきた場合は既存値（DB 上は暗号文）を復号して平文として保持する。
	// lstep.DecryptLineCredential はレガシー平文行もそのまま返す。
	//
	// R-05 Phase B: LineChannelSecret は本経路では一切読まない・書かない。
	// 正規 write owner は clinic_integrations（L-step 設定 API）。
	// 既存 DB 列の値は repository の OnConflict 除外で温存し、presence SELECT 用に残す。
	accessToken := input.LineAccessToken
	if existing != nil && accessToken == "" {
		accessToken = s.decryptCredential(ctx, existing.LineAccessToken)
	}

	// H-4: access token は保存時に常に暗号化する。既存のレガシー平文行は、この経路
	// （次回保存）で自然に暗号化される（機会的再暗号化）。一括 migration は行わない。
	encryptedToken, err := s.encryptCredential(accessToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to encrypt line access token", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to encrypt line access token")
	}

	setting := &model.LineReservationSetting{
		ClinicID:                clinicID,
		Status:                  input.Status,
		HeaderText:              input.HeaderText,
		ReservationNotice:       input.ReservationNotice,
		CancelNotice:            input.CancelNotice,
		PrivacyPolicy:           input.PrivacyPolicy,
		ClosedWeekdays:          input.ClosedWeekdays,
		ClosedDates:             input.ClosedDates,
		NationalHolidayClosed:   input.NationalHolidayClosed,
		BusinessHours:           input.BusinessHours,
		BusinessHoursByWeekday:  input.BusinessHoursByWeekday,
		BreakHours:              input.BreakHours,
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
		AdditionalFields:        input.AdditionalFields,
		LineChannelID:           input.LineChannelID,
		// LineChannelSecret: intentionally unset (zero). Create inserts empty;
		// update OnConflict excludes the column so existing values stay intact.
		LiffID:          input.LiffID,
		LineAccessToken: encryptedToken,
	}
	if err := s.repo.Save(ctx, clinicID, setting); err != nil {
		slog.ErrorContext(ctx, "failed to upsert reservation setting", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to upsert reservation setting")
	}
	// RSV-03: return the write result; do not re-fetch after commit (CODING_RULES.md:78).
	slog.InfoContext(ctx, "reservation setting upserted",
		slog.Uint64("clinic_id", clinicID))
	return setting, isNew, nil
}
