package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

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

// ReservationScheduleService はスタッフスケジュールのビジネスロジックインターフェース
type ReservationScheduleService interface {
	ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error)
	Upsert(ctx context.Context, clinicID, staffID uint64, date time.Time, input *UpsertScheduleInput) (*ScheduleEntry, bool, error)
	Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error
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
	breaksMap, err := s.repo.FindBreaksByEntryIDs(ctx, entryIDs)
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

func (s *reservationScheduleService) Upsert(ctx context.Context, clinicID, staffID uint64, date time.Time, input *UpsertScheduleInput) (*ScheduleEntry, bool, error) {
	shiftType := model.ShiftType(input.ShiftType)
	startTime := normalizeTimeString(input.WorkStart)
	endTime := normalizeTimeString(input.WorkEnd)
	if err := validateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, false, err
	}

	// 既存レコードの有無を確認し、新規作成かどうかを判定する
	existing, err := s.repo.FindByDate(ctx, clinicID, staffID, date)
	isNew := err != nil || existing == nil
	if err != nil && !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to find schedule before upsert", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to find schedule before upsert")
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

	if err := s.repo.Upsert(ctx, clinicID, entry, breaks); err != nil {
		slog.ErrorContext(ctx, "failed to upsert schedule", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to upsert schedule")
	}
	// Upsert後に DB から最新の breaks を取得（ID が振られた状態で返す）
	savedBreaks, err := s.repo.FindBreaksByEntryID(ctx, entry.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load schedule breaks after upsert", "error", err, "clinic_id", clinicID)
		return nil, false, apperrors.Wrap(err, "failed to load schedule breaks after upsert")
	}
	slog.InfoContext(ctx, "schedule upserted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID),
		slog.String("date", date.Format("2006-01-02")))
	return &ScheduleEntry{Entry: *entry, Breaks: savedBreaks}, isNew, nil
}

func (s *reservationScheduleService) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	// 存在確認（NotFound は FromGORM 経由で伝播）
	if _, err := s.repo.FindByDate(ctx, clinicID, staffID, date); err != nil {
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
		slog.String("date", date.Format("2006-01-02")))
	return nil
}
