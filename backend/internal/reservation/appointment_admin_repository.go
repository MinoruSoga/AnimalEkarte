package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationAdminRepository は管理者向け予約管理のデータアクセスインターフェース
type ReservationAdminRepository interface {
	FindAllByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error)
	FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	// FindTimeRangesByDateRange は [from, to) 半開区間の予約から id/doctor_id/start_time/end_time/status のみを
	// Preload なしで一括取得する(G7-1: 日付ループN+1回避のプリフェッチ用軽量クエリ)。
	FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error)
	Create(ctx context.Context, r *model.Reservation) error
	SoftDelete(ctx context.Context, clinicID, id uint64) error
	// LIFF用
	FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	CancelByID(ctx context.Context, clinicID, customerID, id uint64) error
	// 通知用（キャンセル前に関連エンティティを含めて取得）
	FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
}

type reservationAdminRepository struct{ db *gorm.DB }

// Safety caps for admin list queries (G2F-07). Calendar month remains the
// primary bound; these hard ceilings prevent pathological Preload growth.
// Customer history returns the most recent rows only (full paging is a
// follow-up outside this repository unit).
const (
	maxAdminMonthRows      = persistence.MaxMasterListRows
	maxCustomerHistoryRows = 100
)

func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return &reservationAdminRepository{db: db}
}

func (r *reservationAdminRepository) FindAllByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("LineCustomer", "clinic_id = ?", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ?", start, end).
		Order("start_time ASC").
		Limit(maxAdminMonthRows).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	dateJST := date.In(time.Local)
	start := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)

	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("CreatedByStaff", "deleted_at IS NULL").
		Preload("LineCustomer", "clinic_id = ?", clinicID).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ?", start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

// FindTimeRangesByDateRange は日付範囲計算専用の軽量クエリ。slot 計算は Status/DoctorID/StartTime/EndTime
// しか使わないため、FindAllByDay の6 Preload を伴わない Select 限定版として提供する(G7-1)。
func (r *reservationAdminRepository) FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error) {
	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Select("id", "doctor_id", "start_time", "end_time", "status").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ?", from, to).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) Create(ctx context.Context, ra *model.Reservation) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(ra).Error; err != nil {
		return apperrors.FromGORM(err, "appointment", "")
	}
	return nil
}

func (r *reservationAdminRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.Reservation{}, "appointment", clinicID, id)
}

func (r *reservationAdminRepository) FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_customer_id = ? AND deleted_at IS NULL", customerID).
		Order("start_time DESC").
		Limit(maxCustomerHistoryRows).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	var appt model.Reservation
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&appt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

func (r *reservationAdminRepository) CancelByID(ctx context.Context, clinicID, customerID, id uint64) error {
	// BUG-029: LIFF cancel must keep the row visible as status=cancelled
	// (my-reservations + staff calendar). Do not set deleted_at.
	result := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND line_customer_id = ? AND status != ? AND deleted_at IS NULL",
			id, customerID, model.ReservationStatusCancelled).
		Updates(map[string]any{
			"status": model.ReservationStatusCancelled,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "appointment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("appointment", fmt.Sprintf("%d", id))
	}
	return nil
}
