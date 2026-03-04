package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationRepository 予約リポジトリインターフェース
type ReservationRepository interface {
	GetAllReservations(ctx context.Context) ([]model.Reservation, error)
	GetReservationsByDate(ctx context.Context, date string) ([]model.Reservation, error)
	GetReservationByID(ctx context.Context, id string) (*model.Reservation, error)
	GetReservationsByPetID(ctx context.Context, petID string) ([]model.Reservation, error)
	GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.Reservation, error)
	CreateReservation(ctx context.Context, reservation *model.Reservation) error
	UpdateReservation(ctx context.Context, reservation *model.Reservation) error
	DeleteReservation(ctx context.Context, id string) error
}

// reservationRepository 予約リポジトリ実装
type reservationRepository struct {
	db *gorm.DB
}

// NewReservationRepository 新しい予約リポジトリを作成
func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

// GetAllReservations 全ての予約を取得
func (r *reservationRepository) GetAllReservations(ctx context.Context) ([]model.Reservation, error) {
	var reservations []model.Reservation
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Order("start_time ASC, created_at ASC").
		Find(&reservations)

	if result.Error != nil {
		return nil, result.Error
	}

	return reservations, nil
}

// GetReservationsByDate 指定日の予約を取得
func (r *reservationRepository) GetReservationsByDate(ctx context.Context, date string) ([]model.Reservation, error) {
	var reservations []model.Reservation
	// date は YYYY-MM-DD 形式。start_time が当日の 00:00:00 〜 23:59:59 の範囲を対象とする
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Where("DATE(start_time) = ?", date).
		Order("start_time ASC, created_at ASC").
		Find(&reservations)

	if result.Error != nil {
		return nil, result.Error
	}

	return reservations, nil
}

// GetReservationByID IDで予約を取得
func (r *reservationRepository) GetReservationByID(ctx context.Context, id string) (*model.Reservation, error) {
	var reservation model.Reservation
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		First(&reservation, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("reservation with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get reservation")
	}

	return &reservation, nil
}

// GetReservationsByPetID ペットIDで予約を取得
func (r *reservationRepository) GetReservationsByPetID(ctx context.Context, petID string) ([]model.Reservation, error) {
	var reservations []model.Reservation
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Where("pet_id = ?", petID).
		Order("start_time DESC, created_at DESC").
		Find(&reservations)

	if result.Error != nil {
		return nil, result.Error
	}

	return reservations, nil
}

// GetReservationsByOwnerID 飼い主IDで予約を取得
func (r *reservationRepository) GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.Reservation, error) {
	var reservations []model.Reservation
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Where("owner_id = ?", ownerID).
		Order("start_time DESC, created_at DESC").
		Find(&reservations)

	if result.Error != nil {
		return nil, result.Error
	}

	return reservations, nil
}

// CreateReservation 予約を作成
func (r *reservationRepository) CreateReservation(ctx context.Context, reservation *model.Reservation) error {
	// Generate UUID if not set
	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(reservation)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateReservation 予約を更新
func (r *reservationRepository) UpdateReservation(ctx context.Context, reservation *model.Reservation) error {
	result := r.db.WithContext(ctx).Save(reservation)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteReservation 予約を削除
func (r *reservationRepository) DeleteReservation(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Reservation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
