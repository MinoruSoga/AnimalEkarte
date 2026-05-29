package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateReservationAdminInput は管理者手動予約の入力データ
type CreateReservationAdminInput struct {
	StartTime         time.Time
	EndTime           time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         string
	ReservationTypeID uint64
	DoctorID          *uint64
	IsDesignated      bool
	Notes             string
	CreatedBy         *uint64
	LineCustomerID    *uint64
	IsStaffDelegated  bool
	CustomerFields    []byte
}

// ReservationAdminService は管理者向け予約管理のビジネスロジックインターフェース
type ReservationAdminService interface {
	ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error)
	ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationAdminService struct {
	repo                 repository.ReservationAdminRepository
	resRepo              repository.ReservationRepository
	tx                   repository.Transactor
	reservationStaffRepo repository.ReservationStaffRepository
}

func NewReservationAdminService(repo repository.ReservationAdminRepository, resRepo repository.ReservationRepository, tx repository.Transactor, reservationStaffRepo ...repository.ReservationStaffRepository) ReservationAdminService {
	var staffRepo repository.ReservationStaffRepository
	if len(reservationStaffRepo) > 0 {
		staffRepo = reservationStaffRepo[0]
	}
	return &reservationAdminService{repo: repo, resRepo: resRepo, tx: tx, reservationStaffRepo: staffRepo}
}

func (s *reservationAdminService) ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM format for month view")
	}
	items, err := s.repo.FindAllByMonth(ctx, clinicID, t.Year(), t.Month())
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservations by month", "error", err)
		return nil, apperrors.Wrap(err, "failed to list reservations by month")
	}
	return items, nil
}

func (s *reservationAdminService) ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	items, err := s.repo.FindAllByDay(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservations by day", "error", err)
		return nil, apperrors.Wrap(err, "failed to list reservations by day")
	}
	return items, nil
}

func (s *reservationAdminService) Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error) {
	if err := validateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate time range")
	}
	if err := validateReservationStaffCapability(ctx, s.reservationStaffRepo, clinicID, input.DoctorID, input.ReservationTypeID); err != nil {
		return nil, err
	}

	visitType := model.VisitType(input.VisitType)
	if visitType == "" {
		visitType = model.VisitTypeRevisit
	}
	customerFields := json.RawMessage(input.CustomerFields)
	if customerFields == nil {
		customerFields = json.RawMessage("{}")
	}

	var result *model.Reservation
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := checkSlotConflict(ctx, s.resRepo, clinicID, input.DoctorID, input.StartTime, input.EndTime, nil); err != nil {
			return err
		}

		ra := &model.Reservation{
			ClinicID:          clinicID,
			StartTime:         input.StartTime,
			EndTime:           input.EndTime,
			OwnerID:           input.OwnerID,
			PetID:             input.PetID,
			VisitType:         visitType,
			ReservationTypeID: input.ReservationTypeID,
			DoctorID:          input.DoctorID,
			IsDesignated:      input.IsDesignated,
			Notes:             input.Notes,
			Source:            model.ReservationSourceManual,
			CreatedBy:         input.CreatedBy,
			LineCustomerID:    input.LineCustomerID,
			IsStaffDelegated:  input.IsStaffDelegated,
			CustomerFields:    customerFields,
		}
		if err := s.resRepo.Create(ctx, ra); err != nil {
			return apperrors.Wrap(err, "create reservation appointment")
		}
		result = ra
		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "create reservation appointment (transaction)", "error", err)
		return nil, apperrors.Wrap(err, "create reservation appointment (transaction)")
	}

	slog.InfoContext(ctx, "admin reservation created",
		slog.Uint64("reservation_id", result.ID),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}

func (s *reservationAdminService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.resRepo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find reservation appointment")
	}
	if err := s.repo.SoftDelete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation appointment", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete reservation appointment")
	}
	slog.InfoContext(ctx, "admin reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
