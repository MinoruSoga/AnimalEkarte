package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ReservationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) ([]model.ReservationAppointment, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error)
	Create(ctx context.Context, reservation *model.ReservationAppointment) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationAppointment, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) ([]model.ReservationAppointment, int64, error) {
	reservations := make([]model.ReservationAppointment, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.ReservationAppointment{}).Where("clinic_id = ?", clinicID)
	if date != nil {
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_time >= ? AND start_time < ?", start, end)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count reservations")
	}
	if err := q.Preload("Pet").Preload("Pet.Owner").Preload("Pet.AnimalSpecies").Preload("ServiceType").Preload("Doctor").
		Offset((page - 1) * limit).Limit(limit).Order("start_time ASC").Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find reservations")
	}
	return reservations, total, nil
}

func (r *reservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error) {
	var reservation model.ReservationAppointment
	if err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Pet.Owner").
		Preload("Pet.AnimalSpecies").
		Preload("ServiceType").
		Preload("Doctor").
		First(&reservation, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find reservation by id")
	}
	return &reservation, nil
}

func (r *reservationRepository) Create(ctx context.Context, reservation *model.ReservationAppointment) error {
	if err := r.db.WithContext(ctx).Create(reservation).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("reservation", reservation.StartTime.String())
		}
		return apperrors.Wrap(err, "create reservation")
	}
	return nil
}

func (r *reservationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationAppointment, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationAppointment{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update reservation")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ReservationAppointment{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete reservation")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return nil
}
