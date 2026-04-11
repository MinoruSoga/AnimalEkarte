package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationAdminRepository は管理者向け予約管理のデータアクセスインターフェース
type ReservationAdminRepository interface {
	FindByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Appointment, error)
	FindByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Appointment, error)
	Create(ctx context.Context, r *model.Appointment) error
	SoftDelete(ctx context.Context, clinicID, id uint64) error
	// LIFF用
	FindByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error)
	CancelByID(ctx context.Context, clinicID, customerID, id uint64) error
	// 通知用（キャンセル前に関連エンティティを含めて取得）
	FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
}

type reservationAdminRepository struct{ db *gorm.DB }

func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return &reservationAdminRepository{db: db}
}

func (r *reservationAdminRepository) FindByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Appointment, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	items := make([]model.Appointment, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Preload("Doctor").
		Preload("LineCustomer").
		Where("clinic_id = ? AND start_time >= ? AND start_time < ?", clinicID, start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Appointment, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	items := make([]model.Appointment, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Preload("Doctor").
		Preload("LineCustomer").
		Preload("Owner").
		Preload("Pet").
		Where("clinic_id = ? AND start_time >= ? AND start_time < ?", clinicID, start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) Create(ctx context.Context, ra *model.Appointment) error {
	if err := r.db.WithContext(ctx).Create(ra).Error; err != nil {
		return apperrors.FromGORM(err, "appointment", "")
	}
	return nil
}

func (r *reservationAdminRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Delete(&model.Appointment{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "appointment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("appointment", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationAdminRepository) FindByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error) {
	items := make([]model.Appointment, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Preload("Doctor").
		Where("clinic_id = ? AND line_customer_id = ? AND deleted_at IS NULL", clinicID, customerID).
		Order("start_time DESC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	var appt model.Appointment
	err := r.db.WithContext(ctx).
		Preload("ReservationType").
		Preload("Doctor").
		Preload("Owner").
		Preload("Pet").
		First(&appt, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

func (r *reservationAdminRepository) CancelByID(ctx context.Context, clinicID, customerID, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("id = ? AND clinic_id = ? AND line_customer_id = ? AND status != ?",
			id, clinicID, customerID, model.ReservationStatusCancelled).
		Update("status", model.ReservationStatusCancelled)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "appointment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("appointment", fmt.Sprintf("%d", id))
	}
	return nil
}
