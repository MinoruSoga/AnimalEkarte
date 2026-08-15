package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ReservationScheduleBreakInput は休憩時間の入力データ
type ReservationScheduleBreakInput struct {
	Start string
	End   string
}

// CreateReservationScheduleInput はスケジュール upsert の入力データ
type CreateReservationScheduleInput struct {
	ShiftType string
	WorkStart *string
	WorkEnd   *string
	Breaks    []ReservationScheduleBreakInput
}

// ScheduleEntry はスケジュールレスポンス用の集約データ
type ScheduleEntry struct {
	Entry  model.ShiftEntry
	Breaks []model.ShiftEntryBreak
}

// ReservationScheduleService はスタッフスケジュールのビジネスロジックインターフェース
type ReservationScheduleService interface {
	ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error)
	Save(ctx context.Context, clinicID, staffID uint64, date time.Time, input *CreateReservationScheduleInput) (*ScheduleEntry, bool, error)
	Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

type reservationScheduleService struct {
	repo ReservationScheduleRepository
}

func NewReservationScheduleService(repo ReservationScheduleRepository) ReservationScheduleService {
	return &reservationScheduleService{repo: repo}
}

func (s *reservationScheduleService) ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error) {
	entries, err := s.repo.FindAllByMonth(ctx, clinicID, staffID, month)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list schedules", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list schedules")
	}
	if len(entries) == 0 {
		return []ScheduleEntry{}, nil
	}
	entryIDs := make([]uint64, len(entries))
	for i := range entries {
		entryIDs[i] = entries[i].ID
	}
	breaksMap, err := s.repo.FindAllBreaksByEntryIDs(ctx, clinicID, entryIDs)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list schedule breaks", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list schedule breaks")
	}
	result := make([]ScheduleEntry, 0, len(entries))
	for i := range entries {
		result = append(result, ScheduleEntry{
			Entry:  entries[i],
			Breaks: breaksMap[entries[i].ID],
		})
	}
	return result, nil
}

func (s *reservationScheduleService) Save(ctx context.Context, clinicID, staffID uint64, date time.Time, input *CreateReservationScheduleInput) (*ScheduleEntry, bool, error) {
	if input == nil {
		return nil, false, apperrors.WrapInvalidInput("schedule input is required")
	}
	shiftType := model.ShiftType(input.ShiftType)
	startTime := sharedkernel.NormalizeTimeString(input.WorkStart)
	endTime := sharedkernel.NormalizeTimeString(input.WorkEnd)
	if err := sharedkernel.ValidateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, false, err
	}

	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: shiftType,
		StartTime: startTime,
		EndTime:   endTime,
	}

	breaks := make([]model.ShiftEntryBreak, 0, len(input.Breaks))
	for _, b := range input.Breaks {
		breaks = append(breaks, model.ShiftEntryBreak{
			BreakStart: b.Start,
			BreakEnd:   b.End,
		})
	}

	savedEntry, savedBreaks, created, err := s.repo.Save(ctx, clinicID, entry, breaks)
	if err != nil {
		slog.ErrorContext(ctx, "failed to upsert schedule", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to upsert schedule")
	}
	if savedEntry == nil {
		return nil, false, apperrors.WrapInternalServerError(
			"schedule repository returned an empty save result",
		)
	}
	slog.InfoContext(ctx, "schedule upserted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID),
		slog.String("date", date.Format(time.DateOnly)))
	return &ScheduleEntry{Entry: *savedEntry, Breaks: savedBreaks}, created, nil
}

func (s *reservationScheduleService) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	// 存在確認（NotFound は FromGORM 経由で伝播）
	if _, err := s.repo.FindAllByDate(ctx, clinicID, staffID, date); err != nil {
		slog.ErrorContext(ctx, "failed to find schedule before delete", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to find schedule before delete")
	}

	if err := s.repo.Delete(ctx, clinicID, staffID, date); err != nil {
		slog.ErrorContext(ctx, "failed to delete schedule", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete schedule")
	}
	slog.InfoContext(ctx, "schedule deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID),
		slog.String("date", date.Format(time.DateOnly)))
	return nil
}
