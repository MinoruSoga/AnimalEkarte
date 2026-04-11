package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ReservationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
	Create(ctx context.Context, reservation *model.Appointment) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	ExistsByReservationTypeID(ctx context.Context, reservationCategoryID uint64) (bool, error)
	ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error)
	CountMedicalRecordsByReservationID(ctx context.Context, reservationID uint64) (int64, error)
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error) {
	reservations := make([]model.Appointment, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Appointment{}).Where("clinic_id = ?", clinicID)
	if date != nil {
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_time >= ? AND start_time < ?", start, end)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if source != nil {
		q = q.Where("source = ?", *source)
	}
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	if err := q.Preload("Pet").Preload("Pet.Owner").Preload("Pet.AnimalSpecies").Preload("ReservationType").Preload("Doctor").
		Offset((page - 1) * limit).Limit(limit).Order("start_time ASC").Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, total, nil
}

func (r *reservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	var reservation model.Appointment
	err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Pet.Owner").
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType").
		Preload("Doctor").
		First(&reservation, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return &reservation, nil
}

func (r *reservationRepository) Create(ctx context.Context, reservation *model.Appointment) error {
	if err := r.db.WithContext(ctx).Create(reservation).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("reservation", reservation.StartTime.String())
		}
		return apperrors.FromGORM(err, "reservation", "")
	}
	return nil
}

func (r *reservationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Appointment{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationRepository) ExistsByReservationTypeID(ctx context.Context, reservationCategoryID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Appointment{}).
		Where("reservation_type_id = ?", reservationCategoryID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Appointment{}).
		Where("doctor_id = ?", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

// CountMedicalRecordsByReservationID は予約を参照しているカルテの件数を返す（BUG-201）
func (r *reservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, reservationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Where("reservation_appointment_id = ? AND deleted_at IS NULL", reservationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}
