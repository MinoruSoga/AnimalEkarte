package clinic

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicHolidayService は個別休診日管理のビジネスロジックインターフェース
type ClinicHolidayService interface {
	List(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	Set(ctx context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error)
	Remove(ctx context.Context, clinicID uint64, date time.Time) error
}

type clinicHolidayService struct {
	repo ClinicHolidayRepository
}

// NewClinicHolidayService はClinicHolidayServiceを初期化して返す
func NewClinicHolidayService(repo ClinicHolidayRepository) ClinicHolidayService {
	return &clinicHolidayService{repo: repo}
}

func (s *clinicHolidayService) List(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	holidays, err := s.repo.FindAllByYearMonth(ctx, clinicID, yearMonth)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list clinic holidays", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list clinic holidays")
	}
	return holidays, nil
}

func (s *clinicHolidayService) Set(ctx context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error) {
	holiday := &model.ClinicHoliday{
		ClinicID: clinicID,
		Date:     date,
		Reason:   reason,
	}
	result, err := s.repo.Save(ctx, clinicID, holiday)
	if err != nil {
		slog.ErrorContext(ctx, "failed to set clinic holiday", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to set clinic holiday")
	}
	slog.InfoContext(ctx, "clinic holiday set",
		slog.Uint64("clinic_id", clinicID),
		slog.String("date", date.Format(time.DateOnly)))
	return result, nil
}

func (s *clinicHolidayService) Remove(ctx context.Context, clinicID uint64, date time.Time) error {
	if _, err := s.repo.FindByDate(ctx, clinicID, date); err != nil {
		if apperrors.IsNotFound(err) {
			return nil // 冪等: 存在しない場合も成功
		}
		slog.ErrorContext(ctx, "failed to find clinic holiday", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to find clinic holiday")
	}
	if err := s.repo.Delete(ctx, clinicID, date); err != nil {
		if apperrors.IsNotFound(err) {
			return nil // 冪等: Delete タイミングで削除済みの場合も成功
		}
		slog.ErrorContext(ctx, "failed to remove clinic holiday", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to remove clinic holiday")
	}
	slog.InfoContext(ctx, "clinic holiday removed",
		slog.Uint64("clinic_id", clinicID),
		slog.String("date", date.Format(time.DateOnly)))
	return nil
}
