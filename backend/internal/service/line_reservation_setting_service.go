package service

import (
	"context"
	"encoding/json"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	LineChannelSecret       string
	LiffID                  string
	LineAccessToken         string
}

// LineReservationSettingService は予約基本設定のビジネスロジックインターフェース
type LineReservationSettingService interface {
	Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	Upsert(ctx context.Context, clinicID uint64, input *UpsertLineReservationSettingInput) (*model.LineReservationSetting, error)
}

type lineReservationSettingService struct {
	repo repository.LineReservationSettingRepository
}

func NewLineReservationSettingService(repo repository.LineReservationSettingRepository) LineReservationSettingService {
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

func (s *lineReservationSettingService) Upsert(ctx context.Context, clinicID uint64, input *UpsertLineReservationSettingInput) (*model.LineReservationSetting, error) {
	if err := validateJSONFields(map[string][]byte{
		"closed_weekdays":           input.ClosedWeekdays,
		"closed_dates":              input.ClosedDates,
		"business_hours":            input.BusinessHours,
		"business_hours_by_weekday": input.BusinessHoursByWeekday,
		"break_hours":               input.BreakHours,
		"additional_fields":         input.AdditionalFields,
	}); err != nil {
		return nil, err
	}

	// LineChannelSecret / LineAccessToken はレスポンスに含まれないため、
	// フロントエンドは既存値を読み取れない。空文字が送られてきた場合は既存値を保持する。
	channelSecret := input.LineChannelSecret
	accessToken := input.LineAccessToken
	if channelSecret == "" || accessToken == "" {
		existing, err := s.repo.FindByClinicID(ctx, clinicID)
		if err != nil && !apperrors.IsNotFound(err) {
			return nil, apperrors.Wrap(err, "failed to get existing reservation setting")
		}
		if existing != nil {
			if channelSecret == "" {
				channelSecret = existing.LineChannelSecret
			}
			if accessToken == "" {
				accessToken = existing.LineAccessToken
			}
		}
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
		LineChannelSecret:       channelSecret,
		LiffID:                  input.LiffID,
		LineAccessToken:         accessToken,
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, apperrors.Wrap(err, "failed to upsert reservation setting")
	}
	result, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation setting after upsert")
	}
	slog.InfoContext(ctx, "reservation setting upserted",
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}
