package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationScheduleService はスタッフスケジュールのビジネスロジックインターフェース
type ReservationScheduleService interface {
	ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error)
	Upsert(ctx context.Context, clinicID, staffID uint64, date time.Time, input *UpsertScheduleInput) (*ScheduleEntry, error)
	Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

// BreakInput は休憩時間の入力データ
type BreakInput struct {
	Start string
	End   string
}

// UpsertScheduleInput はスケジュール upsert の入力データ
type UpsertScheduleInput struct {
	ShiftType string
	WorkStart *string
	WorkEnd   *string
	Breaks    []BreakInput
}

// ScheduleEntry はスケジュールレスポンス用の集約データ
type ScheduleEntry struct {
	Entry  model.ShiftEntry
	Breaks []model.ShiftEntryBreak
}

type reservationScheduleService struct {
	repo repository.ReservationScheduleRepository
}

func NewReservationScheduleService(repo repository.ReservationScheduleRepository) ReservationScheduleService {
	return &reservationScheduleService{repo: repo}
}

func (s *reservationScheduleService) ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error) {
	entries, err := s.repo.FindByMonth(ctx, clinicID, staffID, month)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list schedules")
	}
	if len(entries) == 0 {
		return []ScheduleEntry{}, nil
	}
	entryIDs := make([]uint64, len(entries))
	for i := range entries {
		entryIDs[i] = entries[i].ID
	}
	breaksMap, err := s.repo.FindBreaksByEntryIDs(ctx, entryIDs)
	if err != nil {
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

func (s *reservationScheduleService) Upsert(ctx context.Context, clinicID, staffID uint64, date time.Time, input *UpsertScheduleInput) (*ScheduleEntry, error) {
	shiftType := model.ShiftType(input.ShiftType)
	startTime := normalizeTimeString(input.WorkStart)
	endTime := normalizeTimeString(input.WorkEnd)
	if err := validateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, err
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

	if err := s.repo.Upsert(ctx, entry, breaks); err != nil {
		return nil, apperrors.Wrap(err, "failed to upsert schedule")
	}
	// Upsert後に DB から最新の breaks を取得（ID が振られた状態で返す）
	savedBreaks, err := s.repo.FindBreaksByEntryID(ctx, entry.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to load schedule breaks after upsert")
	}
	slog.InfoContext(ctx, "schedule upserted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID),
		slog.String("date", date.Format("2006-01-02")))
	return &ScheduleEntry{Entry: *entry, Breaks: savedBreaks}, nil
}

func (s *reservationScheduleService) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	if err := s.repo.DeleteByDate(ctx, clinicID, staffID, date); err != nil {
		return apperrors.Wrap(err, "failed to delete schedule")
	}
	slog.InfoContext(ctx, "schedule deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID),
		slog.String("date", date.Format("2006-01-02")))
	return nil
}
