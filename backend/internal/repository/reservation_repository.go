package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ReservationRepository interface {
	FindAll(ctx context.Context, page, limit int, date *time.Time, status *string) ([]model.ReservationAppointment, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ReservationAppointment, error)
	Create(ctx context.Context, reservation *model.ReservationAppointment) error
	Update(ctx context.Context, reservation *model.ReservationAppointment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindAll(ctx context.Context, page, limit int, date *time.Time, status *string) ([]model.ReservationAppointment, int64, error) {
	var reservations []model.ReservationAppointment
	var total int64

	q := r.db.WithContext(ctx).Model(&model.ReservationAppointment{})
	if date != nil {
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_time >= ? AND start_time < ?", start, end)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count reservations")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("start_time ASC").Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find reservations")
	}
	return reservations, total, nil
}

func (r *reservationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ReservationAppointment, error) {
	var reservation model.ReservationAppointment
	if err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("ServiceType").
		Preload("Doctor").
		First(&reservation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("reservation", id.String())
		}
		return nil, apperrors.Wrap(err, "find reservation by id")
	}
	return &reservation, nil
}

func (r *reservationRepository) Create(ctx context.Context, reservation *model.ReservationAppointment) error {
	if err := r.db.WithContext(ctx).Create(reservation).Error; err != nil {
		return apperrors.Wrap(err, "create reservation")
	}
	return nil
}

func (r *reservationRepository) Update(ctx context.Context, reservation *model.ReservationAppointment) error {
	if err := r.db.WithContext(ctx).Save(reservation).Error; err != nil {
		return apperrors.Wrap(err, "update reservation")
	}
	return nil
}

func (r *reservationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.ReservationAppointment{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete reservation")
	}
	return nil
}
