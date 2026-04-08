package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationAdminFilter は管理者向け予約一覧のフィルタ条件
type ReservationAdminFilter struct {
	View string // "month" or "day"
	Date string // "YYYY-MM" or "YYYY-MM-DD"
}

// ReservationAdminRepository は管理者向け予約管理のデータアクセスインターフェース
type ReservationAdminRepository interface {
	FindByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.ReservationAppointment, error)
	FindByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.ReservationAppointment, error)
	Create(ctx context.Context, r *model.ReservationAppointment) error
	SoftDelete(ctx context.Context, clinicID, id uint64) error
}

type reservationAdminRepository struct{ db *gorm.DB }

func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return &reservationAdminRepository{db: db}
}

func (r *reservationAdminRepository) FindByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.ReservationAppointment, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	items := make([]model.ReservationAppointment, 0)
	err := r.db.WithContext(ctx).
		Preload("ServiceType").
		Preload("Doctor").
		Preload("LineCustomer").
		Where("clinic_id = ? AND start_time >= ? AND start_time < ?", clinicID, start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "find reservations by month")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.ReservationAppointment, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	items := make([]model.ReservationAppointment, 0)
	err := r.db.WithContext(ctx).
		Preload("ServiceType").
		Preload("Doctor").
		Preload("LineCustomer").
		Preload("Owner").
		Preload("Pet").
		Where("clinic_id = ? AND start_time >= ? AND start_time < ?", clinicID, start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "find reservations by day")
	}
	return items, nil
}

func (r *reservationAdminRepository) Create(ctx context.Context, ra *model.ReservationAppointment) error {
	if err := r.db.WithContext(ctx).Create(ra).Error; err != nil {
		return apperrors.Wrap(err, "create reservation appointment")
	}
	return nil
}

func (r *reservationAdminRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Delete(&model.ReservationAppointment{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete reservation appointment")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_appointment", fmt.Sprintf("%d", id))
	}
	return nil
}
