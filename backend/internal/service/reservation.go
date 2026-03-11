package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationService 予約サービスインターフェース
type ReservationService interface {
	GetAllReservations(ctx context.Context) ([]model.ReservationAppointment, error)
	GetReservationsByDate(ctx context.Context, date string) ([]model.ReservationAppointment, error)
	GetReservationByID(ctx context.Context, id string) (*model.ReservationAppointment, error)
	GetReservationsByPetID(ctx context.Context, petID string) ([]model.ReservationAppointment, error)
	GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.ReservationAppointment, error)
	CreateReservation(ctx context.Context, req *model.CreateReservationRequest) (*model.ReservationAppointment, error)
	UpdateReservation(ctx context.Context, id string, req *model.UpdateReservationRequest) (*model.ReservationAppointment, error)
	DeleteReservation(ctx context.Context, id string) error
}

// GetAllReservations 全ての予約を取得
func (s *Service) GetAllReservations(ctx context.Context) ([]model.ReservationAppointment, error) {
	return s.reservationRepo.GetAllReservations(ctx)
}

// GetReservationsByDate 指定日の予約を取得
func (s *Service) GetReservationsByDate(ctx context.Context, date string) ([]model.ReservationAppointment, error) {
	if date == "" {
		return nil, apperrors.WrapInvalidInput("date is required")
	}
	return s.reservationRepo.GetReservationsByDate(ctx, date)
}

// GetReservationByID IDで予約を取得
func (s *Service) GetReservationByID(ctx context.Context, id string) (*model.ReservationAppointment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid reservation ID format")
	}

	reservation, err := s.reservationRepo.GetReservationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if reservation == nil {
		return nil, apperrors.WrapNotFound("reservation with id %s not found", id)
	}

	return reservation, nil
}

// GetReservationsByPetID ペットIDで予約を取得
func (s *Service) GetReservationsByPetID(ctx context.Context, petID string) ([]model.ReservationAppointment, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.reservationRepo.GetReservationsByPetID(ctx, uid.String())
}

// GetReservationsByOwnerID 飼い主IDで予約を取得
func (s *Service) GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.ReservationAppointment, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.reservationRepo.GetReservationsByOwnerID(ctx, uid.String())
}

// CreateReservation 予約を作成
func (s *Service) CreateReservation(ctx context.Context, req *model.CreateReservationRequest) (*model.ReservationAppointment, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		// Try alternate format
		startTime, err = time.Parse("2006-01-02T15:04:05", req.StartTime)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid start_time format")
		}
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		endTime, err = time.Parse("2006-01-02T15:04:05", req.EndTime)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid end_time format")
		}
	}

	reservation := &model.ReservationAppointment{
		ID:           uuid.New(),
		StartTime:    startTime,
		EndTime:      endTime,
		OwnerName:    req.OwnerName,
		PetName:      req.PetName,
		VisitType:    req.VisitType,
		Type:         req.Type,
		Doctor:       req.Doctor,
		IsDesignated: req.IsDesignated,
		Status:       req.Status,
		Notes:        req.Notes,
	}

	if reservation.Status == "" {
		reservation.Status = "pending"
	}

	if req.PetID != nil {
		petUID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		reservation.PetID = &petUID
	}

	if err := s.reservationRepo.CreateReservation(ctx, reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

// UpdateReservation 予約を更新
func (s *Service) UpdateReservation(ctx context.Context, id string, req *model.UpdateReservationRequest) (*model.ReservationAppointment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid reservation ID format")
	}

	reservation, err := s.reservationRepo.GetReservationByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if reservation == nil {
		return nil, apperrors.WrapNotFound("reservation with id %s not found", id)
	}

	// Update fields
	if req.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", *req.StartTime)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid start_time format")
			}
		}
		reservation.StartTime = t
	}
	if req.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", *req.EndTime)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid end_time format")
			}
		}
		reservation.EndTime = t
	}
	if req.OwnerName != nil {
		reservation.OwnerName = *req.OwnerName
	}
	if req.PetName != nil {
		reservation.PetName = *req.PetName
	}
	if req.PetID != nil {
		petUID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		reservation.PetID = &petUID
	}
	if req.VisitType != nil {
		reservation.VisitType = *req.VisitType
	}
	if req.Type != nil {
		reservation.Type = *req.Type
	}
	if req.Doctor != nil {
		reservation.Doctor = *req.Doctor
	}
	if req.IsDesignated != nil {
		reservation.IsDesignated = *req.IsDesignated
	}
	if req.Status != nil {
		reservation.Status = *req.Status
	}
	if req.Notes != nil {
		reservation.Notes = *req.Notes
	}

	if err := s.reservationRepo.UpdateReservation(ctx, reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

// DeleteReservation 予約を削除
func (s *Service) DeleteReservation(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid reservation ID format")
	}

	return s.reservationRepo.DeleteReservation(ctx, uid.String())
}
