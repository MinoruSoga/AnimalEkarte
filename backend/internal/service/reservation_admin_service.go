package service

import (
	"context"
	"time"

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
}

func NewReservationAdminService(repo repository.ReservationAdminRepository) ReservationAdminService {
	return &reservationAdminService{repo: repo}
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
	visitType := model.VisitType(input.VisitType)
	if visitType == "" {
		visitType = model.VisitTypeRevisit
	}
	customerFields := input.CustomerFields
	if customerFields == nil {
		customerFields = []byte("{}")
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
	if err := s.repo.Create(ctx, ra); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation appointment")
	}
	return ra, nil
}

func (s *reservationAdminService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.SoftDelete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation appointment")
	}
	return nil
}
