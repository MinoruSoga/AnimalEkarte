package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationService 予約サービスインターフェース
type ReservationService interface {
	GetAllReservations(ctx context.Context) ([]model.Reservation, error)
	GetReservationsByDate(ctx context.Context, date string) ([]model.Reservation, error)
	GetReservationByID(ctx context.Context, id string) (*model.Reservation, error)
	GetReservationsByPetID(ctx context.Context, petID string) ([]model.Reservation, error)
	GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.Reservation, error)
	CreateReservation(ctx context.Context, req *model.CreateReservationRequest) (*model.Reservation, error)
	UpdateReservation(ctx context.Context, id string, req *model.UpdateReservationRequest) (*model.Reservation, error)
	DeleteReservation(ctx context.Context, id string) error
}

// GetAllReservations 全ての予約を取得
func (s *Service) GetAllReservations(ctx context.Context) ([]model.Reservation, error) {
	return s.reservationRepo.GetAllReservations(ctx)
}

// GetReservationsByDate 指定日の予約を取得
func (s *Service) GetReservationsByDate(ctx context.Context, date string) ([]model.Reservation, error) {
	if date == "" {
		return nil, apperrors.WrapInvalidInput("date is required")
	}
	return s.reservationRepo.GetReservationsByDate(ctx, date)
}

// GetReservationByID IDで予約を取得
func (s *Service) GetReservationByID(ctx context.Context, id string) (*model.Reservation, error) {
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
func (s *Service) GetReservationsByPetID(ctx context.Context, petID string) ([]model.Reservation, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.reservationRepo.GetReservationsByPetID(ctx, uid.String())
}

// GetReservationsByOwnerID 飼い主IDで予約を取得
func (s *Service) GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.Reservation, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.reservationRepo.GetReservationsByOwnerID(ctx, uid.String())
}

// CreateReservation 予約を作成
func (s *Service) CreateReservation(ctx context.Context, req *model.CreateReservationRequest) (*model.Reservation, error) {
	// PetIDをUUIDに変換
	petID, err := uuid.Parse(req.PetID.String())
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	// OwnerIDをUUIDに変換
	ownerID, err := uuid.Parse(req.OwnerID.String())
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	reservation := &model.Reservation{
		ID:           uuid.New(),
		PetID:        petID,
		OwnerID:      ownerID,
		DoctorID:     req.DoctorID,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		VisitType:    req.VisitType,
		ServiceType:  req.ServiceType,
		IsDesignated: req.IsDesignated,
		Status:       "pending",
		Notes:        req.Notes,
	}

	if err := s.reservationRepo.CreateReservation(ctx, reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

// UpdateReservation 予約を更新
func (s *Service) UpdateReservation(ctx context.Context, id string, req *model.UpdateReservationRequest) (*model.Reservation, error) {
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
		reservation.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		reservation.EndTime = *req.EndTime
	}
	if req.VisitType != "" {
		reservation.VisitType = req.VisitType
	}
	if req.ServiceType != "" {
		reservation.ServiceType = req.ServiceType
	}
	if req.DoctorID != nil {
		reservation.DoctorID = req.DoctorID
	}
	if req.Status != "" {
		reservation.Status = req.Status
	}
	if req.Notes != "" {
		reservation.Notes = req.Notes
	}
	reservation.IsDesignated = req.IsDesignated

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
