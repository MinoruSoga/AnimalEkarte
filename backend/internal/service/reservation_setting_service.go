package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationSettingService は予約基本設定のビジネスロジックインターフェース
type ReservationSettingService interface {
	Get(ctx context.Context, clinicID uint64) (*model.ReservationSetting, error)
	Upsert(ctx context.Context, clinicID uint64, input *UpsertReservationSettingInput) (*model.ReservationSetting, error)
}

// UpsertReservationSettingInput は予約設定 upsert のための入力データ
type UpsertReservationSettingInput struct {
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

type reservationSettingService struct {
	repo repository.ReservationSettingRepository
}

func NewReservationSettingService(repo repository.ReservationSettingRepository) ReservationSettingService {
	return &reservationSettingService{repo: repo}
}

func (s *reservationSettingService) Get(ctx context.Context, clinicID uint64) (*model.ReservationSetting, error) {
	setting, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	return setting, nil
}

func (s *reservationSettingService) Upsert(ctx context.Context, clinicID uint64, input *UpsertReservationSettingInput) (*model.ReservationSetting, error) {
	setting := &model.ReservationSetting{
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
		LineChannelSecret:       input.LineChannelSecret,
		LiffID:                  input.LiffID,
		LineAccessToken:         input.LineAccessToken,
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, apperrors.Wrap(err, "failed to upsert reservation setting")
	}
	result, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation setting after upsert")
	}
	return result, nil
}
