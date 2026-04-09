package service

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationAdminService は管理者向け予約管理のビジネスロジックインターフェース
type ReservationAdminService interface {
	ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ReservationAppointment, error)
	ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.ReservationAppointment, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.ReservationAppointment, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// CreateReservationAdminInput は管理者手動予約の入力データ
type CreateReservationAdminInput struct {
	StartTime        time.Time
	EndTime          time.Time
	OwnerID          *uint64
	PetID            *uint64
	VisitType        string
	ServiceTypeID    uint64
	DoctorID         *uint64
	IsDesignated     bool
	Notes            string
	LineCustomerID   *uint64
	IsStaffDelegated bool
	CustomerFields   []byte
}

type reservationAdminService struct {
	repo repository.ReservationAdminRepository
	db   *gorm.DB
}

func NewReservationAdminService(repo repository.ReservationAdminRepository, db *gorm.DB) ReservationAdminService {
	return &reservationAdminService{repo: repo, db: db}
}

func (s *reservationAdminService) ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ReservationAppointment, error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM format for month view")
	}
	items, err := s.repo.FindByMonth(ctx, clinicID, t.Year(), t.Month())
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservations by month")
	}
	return items, nil
}

func (s *reservationAdminService) ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.ReservationAppointment, error) {
	items, err := s.repo.FindByDay(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservations by day")
	}
	return items, nil
}

func (s *reservationAdminService) Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.ReservationAppointment, error) {
	if !input.EndTime.After(input.StartTime) {
		return nil, apperrors.WrapInvalidInput("end_time must be after start_time")
	}

	visitType := model.VisitType(input.VisitType)
	if visitType == "" {
		visitType = model.VisitTypeRevisit
	}
	customerFields := input.CustomerFields
	if customerFields == nil {
		customerFields = []byte("{}")
	}

	var result *model.ReservationAppointment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE で該当時間枠の既存予約を行ロック
		var existing []model.ReservationAppointment
		lockQuery := tx.Raw(`
			SELECT * FROM reservation_appointments
			WHERE clinic_id = ?
			  AND deleted_at IS NULL
			  AND status NOT IN ('cancelled')
			  AND start_time < ?
			  AND end_time > ?
			  AND (? = 0 OR doctor_id = ?)
			FOR UPDATE`,
			clinicID,
			input.EndTime,
			input.StartTime,
			ptrToUint64(input.DoctorID), ptrToUint64(input.DoctorID),
		)
		if lockQuery.Scan(&existing); lockQuery.Error != nil {
			return apperrors.Wrap(lockQuery.Error, "lock reservations for conflict check")
		}
		if len(existing) > 0 {
			return apperrors.WrapConflict("この時間枠は既に予約が入っています")
		}

		ra := &model.ReservationAppointment{
			ClinicID:         clinicID,
			StartTime:        input.StartTime,
			EndTime:          input.EndTime,
			OwnerID:          input.OwnerID,
			PetID:            input.PetID,
			VisitType:        visitType,
			ServiceTypeID:    input.ServiceTypeID,
			DoctorID:         input.DoctorID,
			IsDesignated:     input.IsDesignated,
			Notes:            input.Notes,
			Source:           model.ReservationSourceManual,
			LineCustomerID:   input.LineCustomerID,
			IsStaffDelegated: input.IsStaffDelegated,
			CustomerFields:   customerFields,
		}
		if err := tx.Create(ra).Error; err != nil {
			return apperrors.Wrap(err, "create reservation appointment")
		}
		result = ra
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "admin reservation created",
		slog.Uint64("reservation_id", result.ID),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}

func (s *reservationAdminService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.SoftDelete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation appointment")
	}
	slog.InfoContext(ctx, "admin reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
